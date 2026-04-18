# Project Rules

## Generated Code — NEVER EDIT

Always check for `Code generated` header before editing. Edit source, then regenerate.

| Generated | Source | Regenerate |
|-----------|--------|------------|
| `dbent/gen/ent/*` | `dbent/schema/*.go` | `task dbent:entg` |
| `apidash/internal/graph/generated/*` | `apidash/internal/graph/*.graphqls` | `task apidash:gql` |
| `apidash/internal/graph/model/models_gen.go` | `apidash/internal/graph/*.graphqls` | `task apidash:gql` |
| `config/gen/*` | `config/pkl/**/*.pkl` | `task config:gen` |

## Critical

- NEVER modify files in `core/`, `lace/`, `scripts/` (external modules, read-only)
- NEVER create helper files in `resolver/` directories — gqlgen overwrites them
- NEVER use `graphql.Upload` for file uploads — use base64 string
- ALWAYS run migrations after schema changes (`task dbent:migrate`)
- ALWAYS pass `context.Context` as first param to I/O functions

## Workspace Modules

| Module | Purpose | Writable |
|--------|---------|----------|
| `saas/` | Core app: CLI, API bootstrap, business logic | yes |
| `apidash/` | GraphQL API (`/api/sa`) | yes |
| `dbent/` | Ent ORM schemas & migrations | yes |
| `config/` | PKL configuration | yes |
| `lace/` | Shared utility library (30+ packages) | **no** |
| `core/` | External core module | **no** |
| `scripts/` | Deploy/ops scripts | **no** |
| `examples/` | Example code | yes |

## Code Style

- Context first param, error last return: `func Do(ctx context.Context, id string) (*Result, error)`
- Wrap errors with context: `fmt.Errorf("doThing: %w", err)`
- Import order: stdlib → local monorepo → third-party (blank line between)
- Package names: lowercase, single word, specific (`goerr` not `errors`, `gozap` not `logger`)
- Constructors: `New()` or `NewX()` using options pattern for 3+ config fields
- Logging: `gozap.FromCtx(ctx)` — never `fmt.Println`/`log.Println` in production code
- Database access: ALWAYS `app.New()` → `app.GetPlugins().EntDB.Client` — never direct `ent.Open()` / `sql.Open()`

## Tags in Comments

`@tag('dnd', 'skip')` — `dnd` means never remove the commented-out code block.

## Ignore

Never grep, explore, or reference:
- `docs/framework` — framework internals, off-limits
- `*/gen/*`, `*/generated/*`, `*/vendor/*` — generated/vendored code

## Commits

- Conventional format: `feat(scope): description`, `fix(scope): ...`, `refactor(scope): ...`
- Never add "Co-Authored-By" lines or any mention of Claude/AI
- Never commit `.env`, `*.creds`, `*.pem`, `*.key`, or API keys
