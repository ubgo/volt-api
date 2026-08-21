package clicmd

import (
	"bytes"
	"testing"

	ubgocrypt "github.com/ubgo/crypt"
)

// TestRekeyTargetsExcludeHashes pins the ONE classification mistake that would
// destroy data: registering a hash column as a rekey target.
//
// Why it matters: hashes (passwords, *_hash, *_lookup) are one-way. If a
// future maintainer registers one, runRekey would try to "decrypt" it, fail,
// and abort — or worse, if it happened to decrypt, would overwrite the hash
// with ciphertext and lock every user out. The exclusion is a correctness
// requirement, not a style preference, so it gets a test.
func TestRekeyTargetsExcludeHashes(t *testing.T) {
	t.Parallel()
	banned := map[string]string{
		"password":      "one-way hash — re-encrypting it would lock out every user",
		"key_hash":      "one-way hash",
		"token_hash":    "one-way hash",
		"secret_lookup": "deterministic lookup hash — must stay stable",
	}
	for _, tgt := range rekeyTargets {
		if why, bad := banned[tgt.Column]; bad {
			t.Errorf("rekeyTargets includes hash column %s.%s: %s", tgt.Table, tgt.Column, why)
		}
	}
}

// TestRegisteredTargetsAreWellFormed guards the SQL builder's assumptions:
// runRekey interpolates these identifiers directly, so an empty one produces
// malformed SQL that would only surface against a live database. (An EMPTY
// registration set is legal for the bare boilerplate — the command itself
// refuses to run on it — but every registered entry must be complete.)
func TestRegisteredTargetsAreWellFormed(t *testing.T) {
	RegisterRekeyTargets(RekeyTarget{
		Table: "example_secrets", Column: "secret_enc", IDColumn: "id",
		Why: "test registration",
	})
	t.Cleanup(func() { rekeyTargets = nil })

	seen := map[string]bool{}
	for _, tgt := range rekeyTargets {
		if tgt.Table == "" || tgt.Column == "" || tgt.IDColumn == "" {
			t.Errorf("incomplete target %+v: Table, Column and IDColumn are all required", tgt)
		}
		if tgt.Why == "" {
			t.Errorf("%s.%s has no Why — dry-run output must explain what breaks if the row fails",
				tgt.Table, tgt.Column)
		}
		k := tgt.Table + "." + tgt.Column
		if seen[k] {
			t.Errorf("%s listed twice — it would be re-encrypted twice, making it undecryptable", k)
		}
		seen[k] = true
	}
}

// TestRekeyRoundTrip proves the actual crypto contract the command depends on:
// text encrypted with the OLD key decrypts with that key and re-encrypts under
// the NEW key such that ONLY the new key can read it afterwards.
//
// The final assertion is the important one — it is what makes the migration
// meaningful. If old-key decryption still worked after the rewrite, the
// environments would not really be key-isolated.
func TestRekeyRoundTrip(t *testing.T) {
	t.Parallel()
	oldKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes, AES-256
	newKey := []byte("fedcba9876543210fedcba9876543210")

	for _, plain := range []string{"shpat_fake_token_for_tests", "", "unicode ✓ value"} {
		if plain == "" {
			continue // empty values are excluded by the SQL WHERE clause
		}
		cipher, err := ubgocrypt.EncryptCBC(oldKey, []byte(plain))
		if err != nil {
			t.Fatalf("encrypt with old key: %v", err)
		}
		got, err := ubgocrypt.DecryptCBC(oldKey, cipher)
		if err != nil || !bytes.Equal(got, []byte(plain)) {
			t.Fatalf("old-key round trip failed: %v (got %q)", err, got)
		}
		next, err := ubgocrypt.EncryptCBC(newKey, got)
		if err != nil {
			t.Fatalf("re-encrypt with new key: %v", err)
		}
		if _, err := ubgocrypt.DecryptCBC(oldKey, next); err == nil {
			t.Fatalf("re-encrypted value still decrypts with the OLD key — environments are not key-isolated")
		}
		back, err := ubgocrypt.DecryptCBC(newKey, next)
		if err != nil || !bytes.Equal(back, []byte(plain)) {
			t.Fatalf("new-key decrypt failed: %v (got %q)", err, back)
		}
	}
}
