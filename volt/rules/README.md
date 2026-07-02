# Volt rules

Volt's baseline engineering rules, read **on-demand** before the matching action. They live here (inside `volt/`) so they update as one unit with the rest of the Volt layer — a Volt update touches `volt/`, never your own rules.

`volt/CLAUDE.md` (in the root `CLAUDE.md` import chain) points agents at these.

## Baseline (read for every change)

| File | Read before |
|------|-------------|
| [code-standards.md](code-standards.md) | Writing or editing **any** code — the standing engineering bar (type safety, no hardcoded closed-set values, context-dense comments, modular/no-slop). |
| [golang.md](golang.md) | Any change to **Go** code — the v2 Go taste: the 3 commandments, the real stack, entpoly, reference tables, DB-as-dumb-storage, column naming, multi-tenant safety, Go gotchas, anti-pattern catalog, pre-PR checklist. |
| [git-commit.md](git-commit.md) | Running `git commit`. |

## Per-action guides (read before the matching task)

### Development workflows

| File | Read before |
|------|-------------|
| [new-endpoint.md](new-endpoint.md) | Adding a GraphQL query/mutation |
| [new-schema.md](new-schema.md) | Adding an Ent entity or field |
| [new-service.md](new-service.md) | Adding a business-logic service |
| [new-cli-command.md](new-cli-command.md) | Adding a CLI command |
| [new-cron-job.md](new-cron-job.md) | Adding a scheduled job |
| [new-lace-package.md](new-lace-package.md) | Adding a shared utility to `lace/` |

### Infrastructure

| File | Read before |
|------|-------------|
| [configuration.md](configuration.md) | Adding/modifying PKL config |
| [nats-handler.md](nats-handler.md) | Adding NATS message handlers |
| [migrations.md](migrations.md) | Complex database schema changes |
| [docker-dev.md](docker-dev.md) | Local development setup |

### Code quality

| File | Read before |
|------|-------------|
| [error-handling.md](error-handling.md) | Implementing error handling |
| [logging.md](logging.md) | Adding logging |
| [tracing.md](tracing.md) | Adding OpenTelemetry tracing |
| [validation.md](validation.md) | Input validation |
| [testing.md](testing.md) | Writing tests |
| [libraries.md](libraries.md) | Adding dependencies |

### Security & operations

| File | Read before |
|------|-------------|
| [authentication.md](authentication.md) | Auth-related code |
| [secrets.md](secrets.md) | Handling secrets/API keys |
| [debugging.md](debugging.md) | Investigating issues |
| [deployment.md](deployment.md) | Deploying to the VPS |
| [github-secrets.md](github-secrets.md) | Updating GitHub secrets |

## Deep documentation

Comprehensive guides live in [`../docs/`](../docs/README.md) — including [`../docs/guides/backend-patterns.md`](../docs/guides/backend-patterns.md) (condensed Go cheat-sheet), [`../docs/deployment/`](../docs/deployment/README.md) (Docker, GitHub Actions, secrets), and [`../docs/architecture.md`](../docs/architecture.md) / [`../docs/stack.md`](../docs/stack.md) / [`../docs/ubgo-packages.md`](../docs/ubgo-packages.md).

## Yours vs Volt's

- **Volt's rules** live here in `volt/rules/` — edit them, or pull Volt updates by replacing the `volt/` folder.
- **Your own project rules** go in [`../../.claude/rules/`](../../.claude/rules/) — Volt never writes there, so your rules and Volt's never clash. Reference them from the root `CLAUDE.md` (add lines after the `@volt/CLAUDE.md` import).
