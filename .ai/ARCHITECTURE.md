# Architecture

Go 1.25 workspace monorepo. Schema-first GraphQL, Ent ORM, PKL config, NATS JetStream, OpenTelemetry.

## Workspace Modules

| Module | Role | Key Paths |
|--------|------|-----------|
| `saas/` | App core — CLI, API, cron entrypoints + business services | `saas/cmd/{api,cron,cli}/`, `saas/pkg/` |
| `apidash/` | GraphQL schema + resolvers (`/api/sa` endpoint) | `apidash/internal/graph/` |
| `dbent/` | Ent schema defs + migration runner | `dbent/schema/`, `dbent/gen/ent/` (generated) |
| `config/` | PKL schema + generated Go bindings | `config/pkl/schema/`, `config/gen/` (generated) |
| `lace/` | Shared library — 30+ utility packages (**read-only**) | `lace/gozap/`, `lace/gotel/`, `lace/natso/`, etc. |
| `core/` | External core module (**read-only**) | `core/` |
| `scripts/` | Deploy/ops (**read-only**) | `scripts/` |
| `examples/` | Reference code | `examples/` |

## Binaries

| Binary | Source | Purpose |
|--------|--------|---------|
| API | `saas/cmd/api/main.go` | HTTP server — Gin + GraphQL + REST |
| Cron | `saas/cmd/cron/main.go` | Scheduled jobs via gocron |
| CLI | `saas/cmd/cli/main.go` | Cobra-based admin commands |

## Request Flow (GraphQL)

```
HTTP → Gin middleware (auth, tracing, logging)
     → gqlgen handler
     → apidash/internal/graph/resolver/*.resolvers.go
     → saas/pkg/<feature>/service.go
     → dbent/gen/ent (via plugin.EntDB.Client)
     → PostgreSQL
```

## Bootstrap — `app.New()`

`saas/pkg/app/` composes a `*Plugin` struct holding: EntDB, Redis, Natso, Cli, RootConfig, telemetry.
All entrypoints (api, cron, cli) start with `plugin := app.New(opts...)` then wire routes/jobs/commands.

## Generated Code

Never edit. Always edit source and regenerate. See `.ai/RULES.md` for table.

## Infrastructure (Local Dev)

`task docker:up` starts PostgreSQL, Redis, NATS via `.docker/docker-compose.yaml`.

## Deployment

GitHub Actions → builds multi-stage Docker image → pushes to GHCR → SSH deploy to VPS.
Secrets sync: `.env.prod` → `task scripts:gh:secrets:update` → `ENV_PROD_B64` GitHub secret.
