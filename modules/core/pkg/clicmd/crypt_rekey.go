// crypt_rekey.go — re-encrypt at-rest secrets from one ENCRYPTION_KEY to
// another, in place, against a single database.
//
// Why it exists: environments deliberately hold DIFFERENT ENCRYPTION_KEYs, so
// a database copied from one environment to another carries ciphertext the
// target cannot read. Copying staging rows into prod would leave every stored
// credential as undecryptable bytes — and nothing would fail loudly: the rows
// exist, the UI lists them, and only an actual use discovers the credential is
// garbage. This command closes that gap as the final step of an environment
// migration. (Ported from the sync fork, where its first production use was
// staging → prod on 2026-08-20.)
//
// Invariant: it NEVER writes unless every targeted row decrypts with the old
// key first (see the two-phase design in runRekey). A wrong --old-key
// therefore aborts having changed nothing, rather than writing rows encrypted
// from garbage plaintext.
package clicmd

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	ubgocrypt "github.com/ubgo/crypt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// RekeyTarget describes one column that holds crypt-encrypted text.
//
// Why a registration table rather than hardcoded SQL: the set of encrypted
// columns is app-specific and WILL grow, and a future maintainer must be able
// to see the complete list in one place instead of grepping for Sensitive()
// and guessing which are reversible.
//
// Only REVERSIBLY-encrypted columns belong here (values written with
// ubgo/crypt EncryptCBC under core.encryptionKey). Hash columns (password,
// *_hash, *_lookup) are one-way and key-independent — they survive a
// cross-environment copy untouched and must NOT be registered.
type RekeyTarget struct {
	// Table/Column identify the value to rewrite.
	Table  string
	Column string
	// IDColumn is the primary key used to address a single row on update.
	IDColumn string
	// Filter is an optional extra WHERE predicate (no leading AND). Empty
	// means "every non-empty value in the column".
	Filter string
	// Why documents the column's purpose so an operator reading --dry-run
	// output knows what breaks if this row fails.
	Why string
}

// rekeyTargets — the complete set of reversibly-encrypted columns, filled by
// the application via RegisterRekeyTargets (the Volt boilerplate itself ships
// none: its only Sensitive() column, users.password, is a one-way hash and is
// deliberately NOT a target — re-encrypting a hash would lock every user out).
var rekeyTargets []RekeyTarget

// RegisterRekeyTargets appends app-defined encrypted columns to the rekey set.
// Call it from your app's wiring (e.g. an init in the cmd package) so
// `cli crypt-rekey` knows every column that must be rewritten when a database
// moves between environments. Registering a hash column is a data-destroying
// mistake — see RekeyTarget.
func RegisterRekeyTargets(targets ...RekeyTarget) {
	rekeyTargets = append(rekeyTargets, targets...)
}

func cryptRekeyCommand() *cobra.Command {
	var (
		dsn    string
		oldKey string
		newKey string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "crypt-rekey",
		Short: "Re-encrypt at-rest secrets from one ENCRYPTION_KEY to another",
		Long: "Decrypts every registered reversibly-encrypted column with --old-key and rewrites it " +
			"with --new-key, in place, against --dsn.\n\n" +
			"This is the final step of an environment migration: rows copied from another environment " +
			"carry that environment's ciphertext, which is unreadable by the target's key. Without " +
			"this, credentials look present but fail on first use.\n\n" +
			"SAFETY: every row is decrypted and verified BEFORE anything is written. A " +
			"wrong --old-key aborts with zero writes. Run --dry-run first; it performs " +
			"the full decrypt pass and reports counts without touching the database.\n\n" +
			"Hash columns (passwords, *_hash) are intentionally untouched — they are one-way " +
			"and key-independent, so logins keep working.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dsn == "" || oldKey == "" || newKey == "" {
				return fmt.Errorf("crypt-rekey: --dsn, --old-key and --new-key are all required")
			}
			if oldKey == newKey {
				return fmt.Errorf("crypt-rekey: --old-key and --new-key are identical; nothing to do")
			}
			// Skeleton honesty: an empty target set silently "succeeding"
			// would read as "migration done" while every secret stays
			// unreadable. Fail loudly instead.
			if len(rekeyTargets) == 0 {
				return fmt.Errorf("crypt-rekey: no rekey targets registered — if this app stores " +
					"reversibly-encrypted columns, register them via clicmd.RegisterRekeyTargets " +
					"before running; if it stores none, this command has nothing to do")
			}
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return fmt.Errorf("crypt-rekey: open db: %w", err)
			}
			defer func() { _ = db.Close() }()
			if err := db.PingContext(cmd.Context()); err != nil {
				return fmt.Errorf("crypt-rekey: connect: %w", err)
			}
			return runRekey(cmd.Context(), db, rekeyTargets, oldKey, newKey, dryRun)
		},
	}
	cmd.Flags().StringVar(&dsn, "dsn", "", "target database DSN (the environment being migrated INTO)")
	cmd.Flags().StringVar(&oldKey, "old-key", "", "ENCRYPTION_KEY the data is currently encrypted with (the SOURCE env's key)")
	cmd.Flags().StringVar(&newKey, "new-key", "", "ENCRYPTION_KEY to re-encrypt with (the TARGET env's key)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "decrypt and verify everything, report counts, write nothing")
	return cmd
}

// rekeyRow is one pending rewrite held in memory between the two phases.
type rekeyRow struct {
	id     string
	cipher string
}

// runRekey performs the re-encryption in two phases, per target.
//
// Phase 1 reads and decrypts EVERY row with the old key, accumulating the new
// ciphertext in memory. Phase 2 writes. They are separate on purpose: a wrong
// --old-key is discovered during phase 1, before a single UPDATE has run, so
// the database is left exactly as it was. Interleaving them would leave a
// half-rekeyed table with no way to tell which rows used which key.
//
// The in-memory hold is bounded by the number of encrypted secrets (hundreds
// at most — these are credentials, not user data), so buffering is not a
// concern.
func runRekey(ctx context.Context, db *sql.DB, targets []RekeyTarget, oldKey, newKey string, dryRun bool) error {
	total := 0
	for _, t := range targets {
		where := fmt.Sprintf("%s IS NOT NULL AND %s <> ''", t.Column, t.Column)
		if t.Filter != "" {
			where += " AND " + t.Filter
		}
		q := fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s", t.IDColumn, t.Column, t.Table, where)

		rows, err := db.QueryContext(ctx, q)
		if err != nil {
			// A missing table is fatal, not skippable: it means the schema does
			// not match what this command was written against, and silently
			// skipping would leave secrets unreadable with no warning.
			return fmt.Errorf("crypt-rekey: read %s.%s: %w", t.Table, t.Column, err)
		}

		var pending []rekeyRow
		for rows.Next() {
			var id, cipher string
			if err := rows.Scan(&id, &cipher); err != nil {
				_ = rows.Close()
				return fmt.Errorf("crypt-rekey: scan %s: %w", t.Table, err)
			}
			plain, err := ubgocrypt.DecryptCBC([]byte(oldKey), cipher)
			if err != nil {
				_ = rows.Close()
				return fmt.Errorf("crypt-rekey: %s.%s id=%s failed to decrypt with --old-key "+
					"(wrong key, or this row was already re-encrypted): %w", t.Table, t.Column, id, err)
			}
			next, err := ubgocrypt.EncryptCBC([]byte(newKey), plain)
			if err != nil {
				_ = rows.Close()
				return fmt.Errorf("crypt-rekey: %s.%s id=%s re-encrypt: %w", t.Table, t.Column, id, err)
			}
			pending = append(pending, rekeyRow{id: id, cipher: next})
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("crypt-rekey: iterate %s: %w", t.Table, err)
		}
		_ = rows.Close()

		fmt.Printf("%-22s %-12s %4d row(s) decrypted OK   (%s)\n", t.Table, t.Column, len(pending), t.Why)
		total += len(pending)

		if dryRun || len(pending) == 0 {
			continue
		}

		// Phase 2 — one transaction per target so a failure mid-table rolls
		// back rather than leaving mixed-key rows.
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("crypt-rekey: begin %s: %w", t.Table, err)
		}
		upd := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", t.Table, t.Column, t.IDColumn)
		for _, r := range pending {
			if _, err := tx.ExecContext(ctx, upd, r.cipher, r.id); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("crypt-rekey: update %s id=%s: %w", t.Table, r.id, err)
			}
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("crypt-rekey: commit %s: %w", t.Table, err)
		}
	}

	if dryRun {
		fmt.Printf("\nDRY RUN — %d row(s) verified decryptable with --old-key. Nothing written.\n", total)
		fmt.Println("Re-run without --dry-run to apply.")
		return nil
	}
	fmt.Printf("\nre-encrypted %d row(s) with the new key\n", total)
	return nil
}
