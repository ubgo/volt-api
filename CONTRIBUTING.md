# Development guide

How to set up, run, and work in the Volt API codebase. This is the durable technical reference — the [README](README.md) is the product landing page and will typically be replaced by your own app's README once you start building, so **keep this file** for setup and workflow notes. The original Volt API README is preserved in [VOLT.md](VOLT.md).

> Volt API is a commercial product, not an open-source project — this guide is for licensed users developing on top of Volt API, not a public-contribution workflow. For support or licensing questions see [SUPPORT.md](SUPPORT.md).

## Contents

- [Prerequisites](#prerequisites)
- [Getting started](#getting-started)
- [Local development (portless)](#local-development-portless)
- [Configuration (PKL)](#configuration-pkl)
- [Code generation](#code-generation)
- [Database & migrations](#database--migrations)
- [Scripts & tasks](#scripts--tasks)
- [Project structure](#project-structure)
- [API endpoints](#api-endpoints)
- [Starting your real app](#starting-your-real-app)
- [Deployment](#deployment)
- [Further docs](#further-docs)

## Prerequisites

Install these before you start. The first four are required to run the app; **portless** and **air** make day-to-day dev much nicer; **Redis** is optional.

| Tool | Why you need it | Install |
|---|---|---|
| **Go 1.25+** | The language/toolchain. The repo is a `go.work` workspace. | [go.dev/dl](https://go.dev/dl/) — or `brew install go` |
| **PostgreSQL 15+** | Primary database (with pgvector for embeddings). | [postgresql.org/download](https://www.postgresql.org/download/), Postgres.app, or `task docker:up` (runs it in Docker) |
| **NATS Server (JetStream)** | Event bus for async/real-time features. | [nats.io/download](https://nats.io/download/), or via the Docker stack (`task docker:up`) |
| **Task** | Runs every shortcut command in `Taskfile.yml` (`task dev`, `task dbent:g`, …). | [taskfile.dev/installation](https://taskfile.dev/installation/) — e.g. `brew install go-task` |
| **PKL CLI** | Compiles the type-safe configuration into Go bindings (`task config:gen`). | [pkl-lang.org](https://pkl-lang.org/main/current/pkl-cli/index.html) — e.g. `brew install pkl` |
| **portless** _(recommended)_ | Serves the API at a **stable HTTPS URL** (`https://api.app.localhost`) instead of a raw port. | `npm install -g portless` — see [Local development](#local-development-portless) |
| **air** _(recommended)_ | Live reload for the Go API server; `task dev` uses it. | `go install github.com/air-verse/air@latest` |
| **Redis** _(optional)_ | Caching / sessions. | [redis.io/download](https://redis.io/download/), or via the Docker stack |
| **Docker** _(optional)_ | Runs Postgres + NATS locally with one command (`task docker:up`). | [docker.com](https://www.docker.com/) |

## Getting started

```bash
# 1. Sync the go.work workspace (vendored/replace modules)
task workvendor

# 2. (recommended) Activate the git secret-scanner pre-commit hook
task hooks:install

# 3. Configure env — copy the example and fill in values
cp env.sample .env
#    PORT ....................... API port (default 2310)
#    CONFIG_ENV ................. which PKL env to load (default | prod | sample)
#    Deploy/CI secrets .......... VPS_*, GHCR_PAT, etc. (only needed for deployment)
#    App secrets (DB, keys) ..... read by PKL via read("env:VAR") — see Configuration

# 4. Start local infrastructure (Postgres + NATS) — or use your own
task docker:up

# 5. Create the database if it doesn't exist
PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 -c "CREATE DATABASE appdb;"

# 6. Generate config bindings + run migrations
task config:gen
task dbent:migrate

# 7. Run the dev server — pick one:
task dev                 # air hot reload, port from PKL/$PORT
task dev:portless        # ★ https://api.app.localhost  (needs portless — see below)
```

The API listens on the port from your PKL config (`PORT`, default `2310`), or at `https://api.app.localhost` when run via portless.

## Local development (portless)

Volt API supports two ways to run the dev server:

| Command | URL | Notes |
|---|---|---|
| `task dev` | `http://localhost:2310` | air hot reload. Port comes from your PKL config / `$PORT`. |
| `task dev:portless` | `https://api.app.localhost` | **Recommended.** Stable HTTPS URL via [portless](https://github.com/vercel-labs/portless), air hot reload underneath. |

**Why portless?** It gives your dev server a fixed, memorable HTTPS URL (`https://api.app.localhost`) instead of a raw port — no port conflicts, no cookie/CORS clashes with the frontend, and real HTTPS locally (needed for secure cookies, cross-origin calls from Volt Web, webhooks, etc.). It runs a small proxy on port 443 and injects the actual port into the app for you.

### One-time setup

```bash
npm install -g portless     # if not already installed
portless trust              # install portless' local CA into your system trust store (needs sudo)
```

`portless trust` is what makes `https://api.app.localhost` load with **no browser warning** — it adds a locally-generated certificate authority to your trust store. You only do this once per machine.

### Every day

```bash
portless proxy start        # start the HTTPS proxy daemon on :443 (once per boot; auto-elevates with sudo)
task dev:portless           # start the API at https://api.app.localhost with air hot reload
```

Then open **[https://api.app.localhost](https://api.app.localhost)**. Press **Ctrl-C** to stop. Logs are teed to `tmp/air.log`.

### Changing the dev URL

The subdomain comes from **`modules/saas/portless.json`**:

```json
{ "name": "api.app" }
```

The `name` field is the host: `name: "api.app"` → `https://api.app.localhost`. To change it (e.g. to `https://volt-api.localhost`), edit `name` to `"volt-api"` and restart the dev server. Print all active routes with `portless list`.

> If Volt Web calls this API, point its `VITE_GRAPHQL_URL` (or equivalent) at `https://api.app.localhost` so cookies and CORS line up over HTTPS.

### portless helper commands

| Command | What it does |
|---|---|
| `portless trust` | Install the local CA (one-time, sudo). |
| `portless proxy start` | Start the HTTPS proxy daemon (port 443). |
| `portless proxy stop` | Stop the proxy daemon. |
| `portless list` | Show all active portless routes. |
| `portless run air` | What `task dev:portless` runs under the hood (from `modules/saas`). |

## Configuration (PKL)

Configuration is **type-safe PKL**, not raw env vars. Schemas and per-environment values live in `modules/config/pkl/`, and `task config:gen` compiles them to Go bindings in `modules/config/gen/` (never edit the generated code).

```
modules/config/pkl/
├── schema/          # PKL type definitions (Root.pkl + sections)
└── env/
    ├── default/     # local development values
    ├── prod/        # production values
    └── sample/      # example/reference values
```

- **Select an environment** with `CONFIG_ENV` (`default`, `prod`, or `sample`). Defaults to `default`.
- **Secrets** are never hardcoded — PKL reads them from the environment via `read("env:VAR")`, so put real values in `.env`.
- **After editing any `.pkl`**, run `task config:gen` to regenerate the bindings, and `task config:validate` to check them.

See [.claude/volt/rules/configuration.md](.claude/volt/rules/configuration.md) for adding new config fields/sections.

## Code generation

This is a codegen-heavy repo. **Never edit generated code** (`gen/`, `generated/`, `ent/`); edit the source and regenerate.

| Change | Edit | Regenerate |
|---|---|---|
| Database schema | `modules/dbent/schema/*.go` | `task dbent:g` (or `task dbent:gm` to also migrate) |
| New Ent entity | — | `task dbent:new -- User` then edit + `task dbent:g` |
| GraphQL schema | `modules/apidash/internal/graph/*.graphqls` | `task apidash:gql` (or `task gql` for all) |
| PKL config | `modules/config/pkl/**/*.pkl` | `task config:gen` |

## Database & migrations

```bash
task dbent:g          # regenerate Ent code after editing modules/dbent/schema/*.go
task dbent:migrate    # apply migrations to the database
task dbent:gm         # do both: regenerate + migrate
```

You can also run migrations through the CLI:

```bash
cd modules/saas && go run ./cmd/cli migrate
```

Migrations run with `WithForeignKeys(false)` to avoid ordering issues. For risky changes (required fields, renames, unique constraints), see [.claude/volt/rules/migrations.md](.claude/volt/rules/migrations.md) — the short version is: add nullable/with-default first, backfill, then tighten.

## Scripts & tasks

Reusable commands live in the [Taskfile](https://taskfile.dev) — run `task -l` to see them all.

```bash
task workvendor       # sync the go.work workspace
task dev              # API with air hot reload (logs → tmp/air.log)
task dev:portless     # API at https://api.app.localhost
task lint             # gofmt + go vet across all modules
task saas:build       # build the api, cli, and cron binaries → tmp/bin/
task config:gen       # regenerate PKL config bindings
task config:validate  # validate the config
task dbent:g          # regenerate Ent code
task dbent:migrate    # run migrations
task gql              # regenerate all gqlgen resolvers
task docker:up        # start Postgres + NATS in Docker
task docker:down      # stop the Docker stack
task hooks:install    # activate the pre-commit secret scanner
```

Run tests with `go test ./...` from a module, or across the workspace as needed.

CLI commands run from `modules/saas`:

```bash
cd modules/saas
go run ./cmd/cli ping        # health check
go run ./cmd/cli migrate     # run migrations
go run ./cmd/api             # run the API directly (no hot reload)
go run ./cmd/cron            # run the cron worker
```

## Project structure

```
.
├── modules/
│   ├── saas/       # Core: business logic (pkg/), CLI, API + cron bootstrap (cmd/)
│   ├── apidash/    # Dashboard GraphQL API (/api/sa) — gqlgen resolvers
│   ├── dbent/      # Ent ORM schema definitions + generated client
│   ├── config/     # PKL configuration schemas + environments
│   ├── lace/       # Shared utility library (read-only)
│   ├── core/       # External core module (read-only)
│   ├── scripts/    # Deploy/ops scripts (read-only)
│   └── examples/   # Example code
├── .docker/        # Docker Compose (Postgres, NATS) + prod Dockerfile
├── brand/          # Brand kit (identity, assets, tokens) — regenerates from brand.json
├── docs/volt/           # Deeper guides (dev, deploy, structure, workflow)
├── .github/        # Deploy workflows + secrets templates
└── go.work         # Go workspace linking the modules
```

**Import paths keep the bare module name** (`saas/pkg/app`, `dbent/gen/ent`); only filesystem paths carry the `modules/` prefix.

## API endpoints

| Endpoint | Description | Auth |
|----------|-------------|------|
| `/api/sa` | Dashboard GraphQL API (apidash) | Internal Key header |
| `/api/i/bootinfo` | Boot/health status (DB, Redis, NATS, OTel) | Internal Key header |

The GraphQL playground is key-gated (a `pkey` query param) so it isn't exposed by default. Internal (`/api/i/*`) routes require the `X-Internal-Key` header (value from config / `INTERNAL_KEY`). See [.claude/volt/rules/authentication.md](.claude/volt/rules/authentication.md).

## Starting your real app

The boilerplate is structured so you can carve out your app quickly:

1. Define your domain in `modules/dbent/schema/` (Ent entities) and run `task dbent:g && task dbent:migrate`.
2. Add your GraphQL schema in `modules/apidash/internal/graph/*.graphqls`, run `task apidash:gql`, and implement resolvers (business logic lives in `modules/saas/pkg/`, not in resolvers).
3. Set your configuration in `modules/config/pkl/env/` and regenerate with `task config:gen`.
4. Remove what you don't need — `modules/examples` and the `modules/saas/cmd/test` scratch sandbox are safe to delete; trim NATS/cron if unused.
5. **Rebrand** — the identity lives in [`brand/`](brand/) and regenerates from `brand/brand.json` (`task -t brand/Taskfile.yml all`). Upload the GitHub social preview from `brand/png/og-social-1280.png`.
6. **Replace this project's README** with your own app's README (and update `LICENSE` as needed). Keep this `CONTRIBUTING.md` and `VOLT.md` — `VOLT.md` is the preserved copy of Volt API's original README (the "built on Volt API" reference), and `CONTRIBUTING.md` is your team's setup reference.

What remains — the GraphQL/ORM/events wiring, config, observability, CLI, and cron — is your foundation.

## Deployment

Production ships as a Docker image built from `.docker/Dockerfile.prod` and deployed to a VPS via GitHub Actions on push to the deploy branch.

```bash
task docker:prod:build     # build the production image locally
task docker:prod:run       # run it locally to test
task docker:prod:logs      # tail its logs
```

GitHub secrets are managed from `.env.prod` / `.github/.secrets` and synced with the `scripts` module:

```bash
task scripts:gh:secrets:update   # encode .env.prod → ENV_PROD_B64 and push
task scripts:gh:secrets:push     # push all secrets to the repo
```

See [docs/volt/DEPLOY_GUIDE.md](docs/volt/DEPLOY_GUIDE.md), [.claude/volt/rules/deployment.md](.claude/volt/rules/deployment.md), and [.claude/volt/rules/github-secrets.md](.claude/volt/rules/github-secrets.md) for the full flow.

## Further docs

Deeper guides live in [`docs/volt/`](docs/volt/) and [`.claude/volt/rules/`](.claude/volt/rules/):

- [docs/volt/DEV_GUIDE.md](docs/volt/DEV_GUIDE.md) — quick dev notes (hooks, codegen, endpoints).
- [docs/volt/STRUCTURE.md](docs/volt/STRUCTURE.md) — detailed module map.
- [docs/volt/API_GQL_GUIDE.md](docs/volt/API_GQL_GUIDE.md) — GraphQL API guide.
- [docs/volt/DEPLOY_GUIDE.md](docs/volt/DEPLOY_GUIDE.md) — deployment walkthrough.
- [docs/volt/WORKFLOW_GUIDE.md](docs/volt/WORKFLOW_GUIDE.md) — day-to-day workflow.
- [.claude/volt/rules/](.claude/volt/rules/) — focused rules for endpoints, schemas, services, migrations, NATS, logging, tracing, auth, secrets, and more.
