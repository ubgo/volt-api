# AI Agent Guidelines — Volt API

Volt API — the batteries-included Go backend starter. A `go.work` monorepo with modules under `modules/`, heavy **code generation**. Never modify `gen/`, `generated/`, or `ent/` directories.

## Stack

Go 1.25 | Gin | Ent ORM | gqlgen | PostgreSQL | Redis | NATS | PKL Config | OpenTelemetry

## Workspace Modules

All modules live under `modules/`. Import paths keep the bare module name (e.g. `saas/pkg/app`); filesystem paths are prefixed with `modules/`.

| Module | Purpose |
|--------|---------|
| `modules/saas/` | Core app: CLI, API + cron bootstrap, business logic |
| `modules/apidash/` | GraphQL API (`/api/sa`) |
| `modules/dbent/` | Ent ORM schemas & migrations |
| `modules/lace/` | Shared utility library (30+ packages) |
| `modules/config/` | PKL configuration |

## Where to Make Changes

| Task | Edit | Run | Avoid |
|------|------|-----|-------|
| Database schema | `modules/dbent/schema/*.go` | `task dbent:g && task dbent:migrate` | `modules/dbent/gen/ent/*` |
| GraphQL API | `modules/apidash/internal/graph/*.graphqls` | `task apidash:gql` | `modules/apidash/internal/graph/generated/*` |
| Configuration | `modules/config/*.pkl` | `task config:gen` | `modules/config/gen/*` |
| Shared utility | `modules/lace/` | - | - |
| Business logic | `modules/saas/pkg/` | - | - |
| CLI commands | `modules/saas/cmd/cli/` | - | - |

## Task Commands

| Command | Purpose |
|---------|---------|
| `task docker:up` | Start Docker infrastructure |
| `task docker:down` | Stop Docker containers |
| `task dbent:g` | Generate Ent ORM code |
| `task dbent:gm` | Generate Ent code and migrate the DB |
| `task dbent:migrate` | Run database migrations |
| `task apidash:gql` | Generate GraphQL resolvers |
| `task config:gen` | Generate PKL config bindings |
| `task workvendor` | Sync workspace dependencies |
| `task lint` | Format and vet code |
| `task gitgen` | Stage generated files (commit as `gen`) |
| `task scripts:gh:secrets:update` | Sync .env.prod to GitHub |
| `task scripts:gh:secrets:push` | Push all secrets to GitHub |

## Context-Specific Rules

Read before performing these actions:

| Action | Read First |
|--------|------------|
| Git commit | `.claude/volt/rules/git-commit.md` |
| New GraphQL endpoint | `.claude/volt/rules/new-endpoint.md` |
| New Ent schema/field | `.claude/volt/rules/new-schema.md` |
| New service | `.claude/volt/rules/new-service.md` |
| New CLI command | `.claude/volt/rules/new-cli-command.md` |
| New cron job | `.claude/volt/rules/new-cron-job.md` |
| New lace package | `.claude/volt/rules/new-lace-package.md` |
| Configuration | `.claude/volt/rules/configuration.md` |
| NATS handler | `.claude/volt/rules/nats-handler.md` |
| Migrations | `.claude/volt/rules/migrations.md` |
| Writing tests | `.claude/volt/rules/testing.md` |
| Error handling | `.claude/volt/rules/error-handling.md` |
| Logging | `.claude/volt/rules/logging.md` |
| Tracing | `.claude/volt/rules/tracing.md` |
| Validation | `.claude/volt/rules/validation.md` |
| Adding dependencies | `.claude/volt/rules/libraries.md` |
| Authentication | `.claude/volt/rules/authentication.md` |
| Secrets | `.claude/volt/rules/secrets.md` |
| Debugging | `.claude/volt/rules/debugging.md` |
| Docker/local dev | `.claude/volt/rules/docker-dev.md` |
| Deployment | `.claude/volt/rules/deployment.md` |
| GitHub secrets | `.claude/volt/rules/github-secrets.md` |

## Code Conventions

```go
// Context first, error last
func DoThing(ctx context.Context, id string) (*Result, error)

// Wrap errors with context
return fmt.Errorf("doThing: %w", err)

// Import order: stdlib, third-party, local
import (
    "context"

    "github.com/gin-gonic/gin"

    "saas/pkg/app"
)
```

## Tags

We use some tag system in comments like `@tag('dnd', 'skip')` so each tag has special meaning

* `dnd` - means never remove the commented out code block

## Critical Rules

* **Never** modify generated files directly
* **Never** create helpers in `resolver/` directories (gqlgen will overwrite)
* **Never** use `graphql.Upload` for file uploads (use base64 instead)
* **Always** run migrations after schema changes
* **Always** pass `context.Context` as first param to I/O functions
* **Never** modify any file or create any file in these go modules directories `core`, `lace`, `scripts`

## Detailed Documentation

| Topic | Location |
|-------|----------|
| Architecture | `.ai/ARCHITECTURE.md` |
| Backend packages | `.claude/volt/docs/backend/` |
| CLI reference | `.claude/volt/docs/cli/` |
| Templates | `.claude/volt/templates/` |

## Ignore Files and Folders

Never include these files and folders in your agentic flow or decisions or grep anything at all.

* docs/volt/framework
