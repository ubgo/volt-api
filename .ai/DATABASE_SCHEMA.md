# Database Schema

## Default Backing Store

* Primary: Postgres (pgx). Ent schemas live in `dbent/schema`, generated code in `dbent/gen/ent`.
* Migrations: run via Ent migration commands (`dbent/cmd/dbmigrate`) orchestrated by Taskfile. Use `go work vendor` to keep migration binaries reproducible on servers.

## Modeling Conventions

* One Ent schema per core aggregate/table; name structs singular (`User`, `Space`, `Task`).
* Common fields: `id`, `created_at`, `updated_at`, soft-delete flags only when needed.
* Use enums for constrained domains; prefer relations over JSON blobs except for opaque external payloads.

## ClickHouse / Analytics (Optional)

* If analytics tables are needed, mirror `reportgo` pattern: define HCL schemas under `atlas/clickhouse/` and generate Go structs into `saas/gen/chousetypes`.
* Keep `space_id`/`tenant` columns for multi-tenant isolation and `ref_id` for external correlation.

## Data Access Patterns

* Service layer depends on Ent clients; keep raw SQL limited to performance-critical paths with tests.
* Transactions wrap use-cases (unit-of-work); avoid long-lived transactions in HTTP handlers.
* Apply idempotency keys for event/queue handlers to prevent duplicate inserts.
