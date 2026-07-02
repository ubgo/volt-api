# Architecture

Go 1.25 workspace monorepo. Schema-first GraphQL, Ent ORM, PKL config, NATS JetStream, OpenTelemetry.

## Workspace Modules

| Module | Role | Key Paths |
|--------|------|-----------|
| `modules/saas/` | App core — CLI, API, cron entrypoints + business services | `modules/saas/cmd/{api,cron,cli}/`, `modules/saas/pkg/` |
| `modules/apidash/` | GraphQL schema + resolvers (`/api/sa` endpoint) | `modules/apidash/internal/graph/` |
| `modules/dbent/` | Ent schema defs + migration runner | `modules/dbent/schema/`, `modules/dbent/gen/ent/` (generated) |
| `modules/config/` | PKL schema + generated Go bindings | `modules/config/pkl/schema/`, `modules/config/gen/` (generated) |
| `modules/lace/` | Shared HTTP/DB/CLI/gqlgen wrappers (**read-only**) | `modules/lace/ginserver/`, `modules/lace/db/`, `modules/lace/cli/`, etc. (observability/errors/cache are `github.com/ubgo/*` — see [ubgo-packages.md](ubgo-packages.md)) |
| `modules/core/` | Core module — bootstrap, plugin, `applog` (**read-only**) | `modules/core/pkg/{plugin,applog,clicmd}/` |
| `modules/scripts/` | Deploy/ops (**read-only**) | `modules/scripts/` |
| `modules/examples/` | Reference code | `modules/examples/` |

## Binaries

| Binary | Source | Purpose |
|--------|--------|---------|
| API | `modules/saas/cmd/api/main.go` | HTTP server — Gin + GraphQL + REST |
| Cron | `modules/saas/cmd/cron/main.go` | Scheduled jobs via gocron |
| CLI | `modules/saas/cmd/cli/main.go` | Cobra-based admin commands |

## Request Flow (GraphQL)

```
HTTP → Gin middleware (auth, tracing, logging)
     → gqlgen handler
     → modules/apidash/internal/graph/resolver/*.resolvers.go
     → modules/saas/pkg/<feature>/service.go
     → modules/dbent/gen/ent (via plugin.EntDB.Client)
     → PostgreSQL
```

## Bootstrap — `app.New()`

`modules/saas/pkg/app/` composes a `*Plugin` struct holding: EntDB, Redis, Natso, Cli, RootConfig, telemetry.
All entrypoints (api, cron, cli) start with `plugin := app.New(opts...)` then wire routes/jobs/commands.

## Generated Code

Never edit. Always edit source and regenerate. See `volt/rules/new-schema.md` and the "Never edit generated code" rule in `volt/CLAUDE.md`.

## Infrastructure (Local Dev)

`task docker:up` starts PostgreSQL, Redis, NATS via `.docker/docker-compose.yaml`.

## Deployment

GitHub Actions → builds multi-stage Docker image → pushes to GHCR → SSH deploy to VPS.
Secrets sync: `.env.prod` → `task scripts:gh:secrets:update` → `ENV_PROD_B64` GitHub secret.
