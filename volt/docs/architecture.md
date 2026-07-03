# Architecture

Go 1.25 workspace monorepo. Schema-first GraphQL, Ent ORM, PKL config, NATS JetStream, OpenTelemetry.

## Workspace Modules

| Module | Role | Key Paths |
|--------|------|-----------|
| `modules/webapp/` | App core — CLI, API, cron entrypoints + business services | `modules/webapp/cmd/{api,cron,cli}/`, `modules/webapp/pkg/` |
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
| API | `modules/webapp/cmd/api/main.go` | HTTP server — Gin + GraphQL + REST |
| Cron | `modules/webapp/cmd/cron/main.go` | Scheduled jobs via gocron |
| CLI | `modules/webapp/cmd/cli/main.go` | Cobra-based admin commands |

## Request Flow (GraphQL)

```
HTTP → Gin middleware (auth, tracing, logging)
     → gqlgen handler
     → modules/apidash/internal/graph/resolver/*.resolvers.go
     → modules/webapp/pkg/<feature>/service.go
     → modules/dbent/gen/ent (via plugin.EntDB.Client)
     → PostgreSQL
```

## Bootstrap — `volt.New()` + `webapp.New()`

`volt.New()` (package `core/pkg/volt`) boots the **`*volt.Kernel`** — the shared engine holding EntDB, RedisClient, Cache, Cli, RootConfig, Health. `webapp.New()` (package `webapp`) wraps it in a **`*webapp.Circuit`**, which embeds the Kernel (so `circuit.EntDB`, `circuit.Cache`, … are promoted) and is the seam you extend with your own services. One Kernel can back many Circuits (webapp, adminapp, …).
All entrypoints (api, cron, cli) start with `circuit := webapp.New(opts...)` then wire routes/jobs/commands.

## Generated Code

Never edit. Always edit source and regenerate. See `volt/rules/new-schema.md` and the "Never edit generated code" rule in `volt/CLAUDE.md`.

## Infrastructure (Local Dev)

`task docker:up` starts PostgreSQL, Redis, NATS via `.docker/docker-compose.yaml`.

## Deployment

GitHub Actions → builds multi-stage Docker image → pushes to GHCR → SSH deploy to VPS.
Secrets sync: `.env.prod` → `task scripts:gh:secrets:update` → `ENV_PROD_B64` GitHub secret.
