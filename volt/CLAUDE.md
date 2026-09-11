# Volt API — agent & developer guide

Volt API is a commercial Go backend boilerplate: a `go.work` monorepo with schema-first GraphQL (gqlgen), an Ent ORM with typed migrations, NATS JetStream events, type-safe Go configuration, OpenTelemetry + structured logging, a Cobra CLI, and cron workers. This is the operating guide for AI agents and developers working in the codebase.

## Repository layout

- `modules/` — the workspace modules, linked by `go.work` (import paths keep the bare module name; only filesystem paths carry the `modules/` prefix).
  - `modules/saas/` — core app: business logic (`pkg/`), and the API, cron, and CLI entrypoints (`cmd/`).
  - `modules/apidash/` — the dashboard GraphQL API (`/api/sa`), gqlgen resolvers.
  - `modules/dbent/` — Ent ORM schema definitions + generated client.
  - `modules/config/` — the configuration: schema, defaults and validation, all in `config.go`. Nothing is generated.
  - `modules/lace/`, `modules/core/`, `modules/scripts/` — shared/external, **read-only** (don't edit).
  - `modules/examples/` — example code; `modules/saas/cmd/test/` — a scratch sandbox (safe to delete).
- `.docker/` — Docker Compose (Postgres, NATS) + the production Dockerfile.
- `volt/` — everything Volt-specific (this guide, `README.md`, `CONTRIBUTING.md`, `rules/`, `templates/`, brand kit, docs). Deletable when you make the repo your own.
- `.claude/rules/` — **your** space for project rules (Volt never writes there); `.claude/settings*.json` — your Claude config.

## Code standards

Read two rules before writing or editing code: **[`volt/rules/code-standards.md`](rules/code-standards.md)** — the universal engineering bar (full type safety, no hardcoded closed-set values, context-dense why-and-invariant comments, modular/no-slop) — and **[`volt/rules/golang.md`](rules/golang.md)** — the v2 Go taste (the 3 commandments, the real stack, reference tables over `field.Enum`, DB-as-dumb-storage, column naming, multi-tenant safety, the anti-pattern catalog, and the pre-PR checklist). The per-action guides (endpoints, schemas, migrations, NATS, secrets, …) live beside them in [`volt/rules/`](rules/README.md).

## Conventions (do / don't)

- **Never edit generated code.** Anything under `gen/`, `generated/`, `ent/`, or a `*_gen.go` / file with a `Code generated` header is output — edit the source and regenerate (see Commands). This includes Ent (`modules/dbent/gen/ent/`), and gqlgen (`modules/apidash/internal/graph/generated/`). Configuration is **not** in this list — `modules/config/config.go` is hand-written and is the schema itself.
- **Config is typed Go, never raw `os.Getenv`.** Add a field to `modules/config/config.go` with an `env:` tag and read it from `kernel.RootConfig` — a bound field appears in `config explain`, in the generated `.env.example`, and in validation, which `os.Getenv` does not. Secrets get `secret:"true"`, which masks them structurally. `CONFIG_ENV` (`dev` / `stag` / `prod`) selects which validation rules apply, not which file is read.
- **A process that has a kernel already has the config.** Read `kernel.RootConfig`; do not call `config.GetOrDie()` a second time — that builds a second, independent snapshot.
- **Logging goes through `core/pkg/applog`** (wraps `github.com/ubgo/logger`, aliased `ulog`). Use `applog.FromCtx(ctx)` for trace-correlated logs; fields are `ulog.String`, `ulog.Err`, `ulog.Any`, etc. Never `fmt.Println`/`log.Println` in production code. There is no `gozap`/`zap`.
- **Context first, error last:** `func Do(ctx context.Context, ...) (*Result, error)`. Wrap errors with `fmt.Errorf("op: %w", err)`; never `panic` in libraries.
- **GraphQL is schema-first:** edit `modules/apidash/internal/graph/*.graphqls`, regenerate, then implement resolvers. Business logic lives in `modules/saas/pkg/`, not in resolvers. Never add helper files in `resolver/` (gqlgen overwrites them). Use base64 strings for file uploads, never `graphql.Upload`.
- **Database access** goes through `app.New()` → the Ent client; never `ent.Open()`/`sql.Open()` directly.

## Commands (Taskfile — run `task --list`)

- `task dev` — API with air hot reload (port from config/`$PORT`). `task dev:portless` — the same at **https://volt-api.localhost** via portless (the dev URL comes from `modules/saas/portless.json`, `name: "volt-api"`).
- `task dbent:g` (regen Ent) · `task dbent:gm` (regen + migrate) · `task dbent:migrate` (migrate) · `task dbent:new -- User` (new schema).
- `task apidash:gql` / `task gql` — regenerate gqlgen resolvers · `task config:check` — validate the configuration · `task config:explain` / `task config:document`.
- `task lint` (gofmt + vet) · `task saas:build` (build api/cli/cron) · `task workvendor` (sync workspace) · `task docker:up` / `docker:down`.
- Run tests with `go test ./...` from a module.

## Git

- Conventional commits: `type(scope): description`. See [`volt/rules/git-commit.md`](rules/git-commit.md).
- Never include AI attribution — no "Generated with Claude Code", no "Co-Authored-By: Claude".
- Don't commit `.env`, `*.creds`, `*.pem`, `*.key`, or API keys.

## Docs

- Setup + dev workflow: `volt/CONTRIBUTING.md`.
- Documentation: `volt/docs/` — start at [`volt/docs/README.md`](docs/README.md) (routes by intent). Overview: `getting-started.md`, `architecture.md`, `project-structure.md`, `stack.md`, [`ubgo-packages.md`](docs/ubgo-packages.md) (the ubgo ecosystem). How-tos in `docs/guides/`; ops in `docs/deployment/`; postmortems in `docs/incidents/`.
- Context-specific rules (read before an action — new endpoint, schema, migration, NATS handler, logging, secrets, deployment): [`volt/rules/`](rules/README.md).
- Deeper backend references: `volt/docs/backend/`, `volt/docs/cli/`. Code templates: `volt/templates/`.
- Brand identity (for reuse across the family, e.g. Volt Web): `volt/brand/BRAND_GUIDELINES.md`.
