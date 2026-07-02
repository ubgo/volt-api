# Getting started

Everything you need to get Volt API running locally: what to install, how to configure your environment, how to set up the database, and the two ways to run the dev server. Follow it top to bottom and you'll be answering GraphQL requests in a few minutes. For the deeper reference (config internals, deployment) see [CONTRIBUTING.md](../CONTRIBUTING.md).

## Prerequisites

Install these before you start. **Go**, **Task**, the **PKL CLI**, and **pkl-gen-go** are required to build and generate. **PostgreSQL** and **NATS** are required to run, but the built-in Docker stack (`task docker:up`) can supply both if you don't want to install them. **air** and **portless** just make the dev loop nicer.

| Tool | Required? | Why you need it | Install |
|---|---|---|---|
| **Go 1.25+** | Required | The language/toolchain; the repo is a `go.work` workspace. | [go.dev/dl](https://go.dev/dl/) — or `brew install go` |
| **Task** | Required | Runs every command in these docs (`task dev`, `task dbent:g`, `task config:gen`, …). | [taskfile.dev/installation](https://taskfile.dev/installation/) — or `brew install go-task` |
| **PKL CLI** | Required | Evaluates the type-safe PKL configuration (used by codegen + at runtime). | `brew install pkl` — or [pkl-lang.org install](https://pkl-lang.org/main/current/pkl-cli/index.html) |
| **pkl-gen-go** | Required | Generates the Go config bindings from the PKL schema — what `task config:gen` runs. | `go install github.com/apple/pkl-go/cmd/pkl-gen-go@latest` — [github.com/apple/pkl-go](https://github.com/apple/pkl-go) |
| **PostgreSQL 15+** | Required (or via Docker) | Primary database (with pgvector for embeddings). | [postgresql.org/download](https://www.postgresql.org/download/), Postgres.app, or `task docker:up` |
| **NATS (JetStream)** | Required (or via Docker) | Event bus for async / real-time features. | [nats.io/download](https://nats.io/download/), or `task docker:up` |
| **air** | Recommended | Live reload for the API server; `task dev` uses it. | `go install github.com/air-verse/air@latest` |
| **portless** | Recommended | Serves the API at a stable HTTPS URL (`https://volt-api.localhost`) instead of a raw port. | `npm install -g portless` — [github.com/vercel-labs/portless](https://github.com/vercel-labs/portless) |
| **Docker** | Optional | Runs Postgres + NATS locally with one command (`task docker:up`). | [docker.com](https://www.docker.com/) |
| **Redis** | Optional | Caching / sessions. | [redis.io/download](https://redis.io/download/), or via the Docker stack |

> Quick check: `go version` (≥ 1.25), `task --version`, `pkl --version`, and `pkl-gen-go --help` should all succeed before you continue.

## 1. Clone and sync the workspace

```bash
git clone git@github.com:khanakia/volt-api.git
cd volt-api
task workvendor          # sync the go.work workspace (vendored/replace modules)
task hooks:install       # (recommended) activate the gitleaks pre-commit secret scanner (installs gitleaks if missing)
```

## 2. Configure your environment

Copy the sample env. Real secrets live only in `.env` (gitignored); PKL reads them via `read("env:VAR")`.

```bash
cp env.sample .env
```

At minimum, set:

| Variable | What to set it to |
|---|---|
| `PORT` | API port (default `2310`; also used by Docker Compose). |
| `CONFIG_ENV` | Which PKL environment to load — `default`, `prod`, or `sample` (default `default`). |
| `INTERNAL_KEY` | Any random string — protects the internal endpoints under `/api/i/*` (sent as `X-Internal-Key`). |

DB/NATS defaults line up with `task docker:up`. Deploy-only secrets (`VPS_*`, `GHCR_PAT`, …) are only needed for deployment. Full config model: [Configuration in CONTRIBUTING](../CONTRIBUTING.md#configuration-pkl).

## 3. Start infrastructure

Either start Postgres + NATS with the bundled Docker stack, or point your `.env` at your own instances.

```bash
task docker:up           # Postgres + NATS in Docker
```

Create the database if it doesn't exist yet:

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 -c "CREATE DATABASE appdb;"
```

## 4. Generate config bindings and run migrations

```bash
task config:gen          # pkl-gen-go → typed Go config bindings (needs pkl + pkl-gen-go)
task dbent:migrate       # apply the Ent schema to the database
```

## 5. Run the dev server

Pick one of two paths.

### Recommended: portless at `https://volt-api.localhost`

A stable HTTPS URL with a trusted cert — no random ports, no browser warnings, real HTTPS locally (which secure cookies and cross-origin calls from a frontend need).

One-time setup per machine:

```bash
npm install -g portless    # if you haven't already
portless trust             # install portless' local CA into your system trust store (needs sudo)
```

Then, day to day:

```bash
portless proxy start       # start the HTTPS proxy daemon on :443 (once per boot; auto-elevates with sudo)
task dev:portless          # run the API at https://volt-api.localhost with air hot reload
```

Open **https://volt-api.localhost** and press **Ctrl-C** to stop. The dev URL comes from `modules/saas/portless.json` (`name: "volt-api"`); change `name` to change the subdomain.

Handy portless commands:

| Command | What it does |
|---|---|
| `portless trust` | Install the local CA (one-time, sudo). |
| `portless proxy start` / `stop` | Start / stop the HTTPS proxy daemon (port 443). |
| `portless list` | Show all active portless routes. |

### Simple: `task dev` at `http://localhost:2310`

No extra HTTPS setup, fixed port from your config:

```bash
task dev                   # air hot reload → http://localhost:2310
```

## Your first GraphQL call

The dashboard API is served at **`/api/sa`** and guarded by the internal key (`INTERNAL_KEY`, sent as the `X-Internal-Key` header). Confirm the service is healthy first:

```bash
curl -H "X-Internal-Key: $INTERNAL_KEY" http://localhost:2310/api/i/bootinfo
# → status of database, redis, nats, otel
```

Then POST a query (or open the key-gated GraphQL playground):

```bash
curl -X POST http://localhost:2310/api/sa \
  -H "X-Internal-Key: $INTERNAL_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
```

## Where to go next

- **Build a feature** → [guides/graphql-api.md](guides/graphql-api.md) (add a query/mutation), [guides/ent-crud.md](guides/ent-crud.md) (generate CRUD), and the per-action checklists in [`../rules/`](../rules/README.md).
- **Understand the layout** → [architecture.md](architecture.md), [project-structure.md](project-structure.md).
- **See what's wired in** → [stack.md](stack.md) and the [ubgo package ecosystem](ubgo-packages.md).
- **Ship it** → [deployment/](deployment/README.md).
