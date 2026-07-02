# Project structure

A map of the `boilerplate-golang` (Volt API) repository, so you know where each kind of logic lives. It's a `go.work` monorepo: all modules sit under `modules/`, but their **import paths keep the bare module name** (e.g. `import "saas/pkg/app"`) — only filesystem paths carry the `modules/` prefix.

## Root

| Path | Description |
| :--- | :--- |
| **`modules/`** | The workspace modules (below). |
| **`volt/`** | Everything Volt-specific — this documentation, the brand kit, rules, templates. Deletable when you make the repo your own. |
| **`.docker/`** | Docker Compose for local infra (Postgres, NATS) + the production Dockerfile. |
| **`.claude/`** | Your project's Claude config (`rules/` is your space; Volt's rules live in `volt/rules/`). |
| **`.github/`** | CI/CD workflows + secrets templates. |

## Modules (`modules/`)

| Module | Role | Writable |
| :--- | :--- | :--- |
| **`saas/`** | App core — business services (`pkg/`) and the API, cron, and CLI entrypoints (`cmd/`). The most important module. | yes |
| **`apidash/`** | The dashboard GraphQL API (`/api/sa`) — gqlgen schema + resolvers. | yes |
| **`dbent/`** | Ent ORM schema definitions + the generated client. | yes |
| **`config/`** | PKL configuration schemas + generated Go bindings. | yes |
| **`lace/`** | Shared utility library. | **read-only** |
| **`core/`** | External core module (bootstrap, plugin, logging wrapper). | **read-only** |
| **`scripts/`** | Deploy/ops scripts. | **read-only** |
| **`examples/`** | Reference/example code. | yes |

## Where things live

### `saas/` — the core
- **`cmd/{api,cron,cli}/`** — the three binaries (see [architecture.md](architecture.md)). `cmd/test/` is a scratch sandbox, safe to delete.
- **`pkg/<feature>/`** — business logic services. Resolvers delegate here.

### `apidash/` — the GraphQL gateway
- **`internal/graph/*.graphqls`** — **SOURCE**. The GraphQL schema. Edit, then `task apidash:gql`.
- **`internal/graph/generated/`** — **GENERATED**. Do not edit.
- **`internal/graph/resolver/`** — resolver implementations. No helper files here (gqlgen overwrites).

### `dbent/` — the ORM
- **`schema/*.go`** — **SOURCE**. Edit to change tables/fields, then `task dbent:g`.
- **`gen/ent/`** — **GENERATED**. Do not edit.

### `config/` — configuration
- **`pkl/`** — **SOURCE**. PKL files (schema + per-environment values, selected by `CONFIG_ENV`).
- **`gen/`** — **GENERATED**. Go bindings. Do not edit.

### `lace/` · `core/` (read-only shared)
- **Logging**: `core/pkg/applog` — wraps [`github.com/ubgo/logger`](ubgo-packages.md) (console / OTEL / file sinks).
- **Bootstrap**: `core/pkg/plugin` — composes the app `*Plugin` (DB, Redis, NATS, config, telemetry).
- **`lace/natso/`**: NATS JetStream wrapper. Other shared helpers live under `lace/`.

> Rule of thumb: **never edit anything under `gen/`, `generated/`, or `ent/`** (or a file with a `Code generated` header) — edit the source and regenerate. See [`../rules/code-standards.md`](../rules/code-standards.md).
