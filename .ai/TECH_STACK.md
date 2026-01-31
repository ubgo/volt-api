# Tech Stack

## Language & Tooling

* Go 1.25 workspace (`go.work`) with multi-module layout; `go work vendor` for server parity.
* Taskfile (`Taskfile.yml`) as the single entrypoint for builds, codegen, and migrations.

## Core Libraries

* HTTP: Gin for routing/middleware.
* GraphQL: gqlgen for schema-first APIs; `gql` (public) and `apidash` (service) modules.
* ORM/DB: Ent for schema + migrations; pgx driver for Postgres. Optional Atlas if ClickHouse is introduced.
* Config: PKL configs compiled via `pkl-gen-go` in `config`.
* Logging/Monitoring: `lace` wrappers for Loki + Sentry; structured logging via zap-style API.
* Messaging: NATS for async workflows and inter-service events.
* Validation/CLI: valgo or validator where useful; Cobra for CLI commands.

## Why These Choices

* gqlgen + Ent: typed schemas end-to-end, minimal runtime reflection, strong codegen support.
* PKL: declarative, typed configs with generation safety.
* NATS: lightweight, fits event-heavy workflows without heavyweight brokers.
* Taskfile: reproducible commands across modules, easier than make for cross-platform dev.
