# Migrating from PostgreSQL to SQLite (with entgo)

This guide documents every change required to switch this boilerplate from PostgreSQL to SQLite while keeping entgo as the ORM.

---

## Why SQLite?

- Zero infrastructure — no Docker container, no server process
- Single file database — easy to version, copy, reset
- Perfect for local development, CLI tools, and lightweight apps
- entgo supports SQLite natively via `dialect.SQLite`

## Prerequisites

- Go 1.21+ with CGO enabled (required by `go-sqlite3`)
- `gcc` installed (macOS: comes with Xcode CLI tools, Linux: `apt install gcc`)

---

## Step-by-Step Migration

### 1. Replace the database driver

**File: `lace/db/db.go`**

```diff
 import (
     "database/sql"
     "fmt"
     "os"

-    _ "github.com/jackc/pgx/v5/stdlib"
+    _ "github.com/mattn/go-sqlite3"
 )

 func New(opts ...Option) *sql.DB {
     cfg := config{
-        DataSourceName: "host=localhost user=postgres dbname=lacedb sslmode=disable password=postgres port=5432",
-        DriverName:     "pgx",
+        DataSourceName: "file:app.db?cache=shared&_fk=1",
+        DriverName:     "sqlite3",
     }
     // ... rest unchanged
 }
```

**What this does:** Registers the `sqlite3` driver instead of `pgx`. The default DSN is a SQLite file with shared cache and foreign key enforcement enabled.

### 2. Change the ent dialect in the client constructor

**File: `dbent/main.go`**

```diff
-import (
-    "config/gen/databaseconfig"
-    ...
-)
+import (
+    // removed "config/gen/databaseconfig"
+    ...
+)

 func New(db *sql.DB) EntDB {
-    drv := entsql.OpenDB(dialect.Postgres, db)
+    drv := entsql.OpenDB(dialect.SQLite, db)
     client := ent.NewClient(ent.Driver(&CustomDriver{drv}), ent.Debug())
     // ... rest unchanged
 }
```

### 3. Simplify `InitDB` — no more connection string builder

**File: `dbent/main.go`**

PostgreSQL needed host, port, username, password, dbname, sslmode. SQLite just needs a file path.

```diff
-func InitDB(databaseconfig databaseconfig.Setting) *sql.DB {
-    return db.New(db.WithDataSourceName(dsName(databaseconfig)), db.WithDriverName("pgx"))
-}
-
-func dsName(databaseconfig databaseconfig.Setting) string {
-    username := databaseconfig.Username
-    password := databaseconfig.Password
-    databaseName := databaseconfig.DbName
-    databaseHost := databaseconfig.Host
-    databasePort := databaseconfig.Port
-    sslmode := databaseconfig.Sslmode
-
-    return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s port=%d",
-        databaseHost, username, databaseName, sslmode, password, databasePort)
-}
+func InitDB(dbPath string) *sql.DB {
+    dsn := fmt.Sprintf("file:%s?cache=shared&_fk=1", dbPath)
+    return db.New(db.WithDataSourceName(dsn), db.WithDriverName("sqlite3"))
+}
```

**DSN parameters explained:**
- `file:<path>` — SQLite file path
- `cache=shared` — shared cache mode for better concurrent read performance
- `_fk=1` — enable foreign key constraint enforcement (off by default in SQLite)

### 4. Change the migration dialect

**File: `dbent/pkg/dbent_migrate/dbent_migrate.go`**

```diff
 func Migrate(ctx context.Context, db *sql.DB) error {
-    drv := entsql.OpenDB(dialect.Postgres, db)
+    drv := entsql.OpenDB(dialect.SQLite, db)
     client := ent.NewClient(ent.Driver(drv), ent.Debug())
     // ... rest unchanged
 }
```

### 5. Update all callers of `InitDB`

The signature changed from `InitDB(databaseconfig.Setting)` to `InitDB(string)`. Use `cfg.Database.DbName + ".db"` to derive the file path from existing PKL config.

**File: `core/pkg/plugin/app.go`**

```diff
 func initDB(databaseconfig databaseconfig.Setting, plugins *Plugin) error {
-    db := dbent.InitDB(databaseconfig)
+    db := dbent.InitDB(databaseconfig.DbName + ".db")
     // ... rest unchanged
 }
```

**File: `dbent/cmd/migrate/main.go`**

```diff
 func main() {
     cfg := config.Get()
-    db := dbent.InitDB(cfg.Database)
+    db := dbent.InitDB(cfg.Database.DbName + ".db")
     dbent_migrate.Migrate(context.Background(), db)
 }
```

### 6. Remove unused PostgreSQL imports

**File: `dbent/cmd/entg/main.go`**

```diff
-    _ "github.com/lib/pq"
```

This was a side-effect import to register the `pq` driver for code generation. Not needed for SQLite.

### 7. Update Go module dependencies

```bash
# Add sqlite3 driver to both modules that need it
cd lace && go get github.com/mattn/go-sqlite3 && go mod tidy
cd dbent && go get github.com/mattn/go-sqlite3 && go mod tidy
cd core && go mod tidy
cd saas && go mod tidy

# Sync the workspace vendor directory
cd <project-root> && go work vendor
```

### 8. Add `.db` files to `.gitignore`

```gitignore
# SQLite database files
*.db
```

### 9. Regenerate ent code and run migration

```bash
# Generate ent code (unchanged command)
cd dbent && go run ./cmd/entg .

# Run migration — creates SQLite file and tables
cd dbent && go run ./cmd/migrate .
```

---

## Files Changed Summary

| File | Change |
|------|--------|
| `lace/db/db.go` | `pgx` driver → `go-sqlite3`, default DSN to SQLite |
| `dbent/main.go` | `dialect.Postgres` → `dialect.SQLite`, `InitDB` takes `string` path |
| `dbent/pkg/dbent_migrate/dbent_migrate.go` | `dialect.Postgres` → `dialect.SQLite` |
| `dbent/cmd/migrate/main.go` | Pass `cfg.Database.DbName + ".db"` to `InitDB` |
| `dbent/cmd/entg/main.go` | Remove `_ "github.com/lib/pq"` import |
| `core/pkg/plugin/app.go` | Pass `databaseconfig.DbName + ".db"` to `InitDB` |
| `lace/go.mod` | Add `github.com/mattn/go-sqlite3`, remove `pgx` |
| `dbent/go.mod` | Add `github.com/mattn/go-sqlite3`, remove `lib/pq` |
| `.gitignore` | Add `*.db` |

---

## SQLite vs PostgreSQL — What to Watch

### Things that work the same

- All standard entgo schema definitions (fields, edges, indexes)
- `Schema.Create` auto-migration
- Basic CRUD queries via ent client
- Transaction support (`WithTx`)

### Things that differ

| Feature | PostgreSQL | SQLite |
|---------|-----------|--------|
| JSON fields | Native `jsonb` operators | Stored as text, no native operators |
| Array fields | Native `text[]` | Not supported — use JSON |
| LISTEN/NOTIFY | Supported | Not available |
| Concurrent writes | Full MVCC | WAL mode recommended for concurrency |
| Full-text search | `tsvector` / `tsquery` | FTS5 extension (different API) |
| UPSERT syntax | `ON CONFLICT DO UPDATE` | Same (supported since SQLite 3.24) |
| Schema ALTER | Most ALTER TABLE ops | Limited ALTER TABLE (no DROP COLUMN before 3.35) |
| Enums | Native ENUM type | Stored as text with CHECK constraint |

### Recommendations for SQLite

1. **Enable WAL mode** for better concurrent read/write:
   ```go
   db.Exec("PRAGMA journal_mode=WAL")
   ```

2. **Busy timeout** to handle lock contention:
   ```go
   db.Exec("PRAGMA busy_timeout=5000")
   ```

3. **Avoid postgres-specific raw SQL** — any `entsql.Selector` modifications that use postgres syntax (e.g., `::jsonb`, `ANY()`, `ARRAY[]`) will break.

4. **File location** — the `.db` file is created relative to the working directory when running the command. In production, use an absolute path.

---

## Switching Back to PostgreSQL

To reverse this migration:

1. Restore `pgx` driver in `lace/db/db.go`
2. Change `dialect.SQLite` back to `dialect.Postgres` in `dbent/main.go` and `dbent_migrate.go`
3. Restore the `dsName()` function and `InitDB(databaseconfig.Setting)` signature
4. Update callers to pass the full `databaseconfig.Setting` again
5. Run `go mod tidy` and `go work vendor`

---

## Testing the Migration

```bash
# 1. Build all modules
cd dbent && go build ./...
cd core && go build ./...
cd saas && go build ./...

# 2. Run migration
cd dbent && go run ./cmd/migrate .

# 3. Verify tables
sqlite3 dbent/<dbname>.db ".tables"
sqlite3 dbent/<dbname>.db ".schema users"

# 4. Start the API
go run ./saas/cmd/api
```
