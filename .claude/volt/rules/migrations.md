# Migration Rules

**Read before:** Making database schema changes.

## Checklist

- [ ] Edit schema in `dbent/schema/`
- [ ] Run `task dbent:g` to regenerate code
- [ ] Run `task dbent:migrate` to apply migration
- [ ] Test migration locally before deploying
- [ ] Consider data migration if needed

## Commands

```bash
# Generate Ent code from schema
task dbent:g

# Apply migrations to database
task dbent:migrate

# Or via CLI
go run ./saas/cmd/cli migrate
```

## Safe Changes (Auto-migrate handles)

| Change | Notes |
|--------|-------|
| Add new entity | Safe |
| Add nullable field | Safe |
| Add field with default | Safe |
| Add index | Safe |
| Remove index | Safe |

## Risky Changes (Need care)

| Change | Risk | Mitigation |
|--------|------|------------|
| Add required field | Fails if data exists | Add with default or nullable first |
| Remove field | Data loss | Backup first, remove in stages |
| Rename field | Data loss | Add new, migrate data, remove old |
| Change field type | May fail | Depends on conversion compatibility |
| Add unique constraint | Fails if duplicates | Clean data first |

## Adding Required Field to Existing Table

**Wrong:** Adding `field.String("code").NotEmpty()` directly fails if rows exist.

**Right:** Two-step migration:

```go
// Step 1: Add as nullable with default
field.String("code").Optional().Default("")

// Run: task dbent:g && task dbent:migrate

// Step 2: After backfilling data, make required
field.String("code").NotEmpty()

// Run: task dbent:g && task dbent:migrate
```

## Data Migration Pattern

For complex data migrations, create a CLI command:

```go
// saas/cmd/cli/main.go
var migrateDataCommand = &cobra.Command{
    Use:   "migrate-data-v2",
    Short: "Migrate data for v2 schema",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.Background()
        client := plugin.EntDB.Client()

        // Migrate in batches
        users, _ := client.User.Query().
            Where(user.NewFieldIsNil()).
            Limit(1000).
            All(ctx)

        for _, u := range users {
            client.User.UpdateOne(u).
                SetNewField(computeValue(u)).
                Save(ctx)
        }

        fmt.Println("Migration completed")
    },
}
```

## Renaming a Field

1. Add new field:
```go
field.String("full_name").Optional()  // New name
field.String("name").Optional()       // Old name (keep)
```

2. Run migration: `task dbent:g && task dbent:migrate`

3. Create data migration to copy values:
```go
// Copy name -> full_name for all records
```

4. Update code to use new field

5. Remove old field in next release

## Adding Unique Constraint

1. Check for duplicates first:
```sql
SELECT email, COUNT(*) FROM users GROUP BY email HAVING COUNT(*) > 1;
```

2. Clean duplicates manually or via script

3. Add constraint:
```go
field.String("email").Unique()
```

4. Run migration

## Rollback Strategy

Ent auto-migrate doesn't support rollback. Strategies:

1. **Database backup** before migration
2. **Additive changes** - add new, don't remove old
3. **Manual rollback SQL** for critical changes

```bash
# Backup before migration
pg_dump -U postgres mydb > backup_$(date +%Y%m%d).sql

# Restore if needed
psql -U postgres mydb < backup_20240115.sql
```

## Testing Migrations

```bash
# Test on fresh database
docker compose down -v
docker compose up -d postgres
task dbent:migrate

# Test on copy of production
pg_dump prod_db > test_backup.sql
psql test_db < test_backup.sql
task dbent:migrate  # with test_db config
```

## Migration in Production

1. Take database backup
2. Deploy new code (without migration)
3. Run migration command
4. Verify application works
5. Keep backup for 24-48 hours

```bash
# Via CLI in production
./api migrate
```

## Foreign Keys

By default, migrations run with `WithForeignKeys(false)`:

```go
// dbent/pkg/dbent_migrate/dbent_migrate.go
migrate.Create(ctx, migrate.WithForeignKeys(false))
```

This prevents ordering issues during migration.
