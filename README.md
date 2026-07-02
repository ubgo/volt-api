<div align="center">

<img src="brand/png/icon-512.png" alt="Volt API" width="112" height="112" />

# Volt API

**The batteries-included Go backend starter.**

Volt API is a production-ready Go backend boilerplate for building modular services — schema-first GraphQL, Ent ORM, NATS JetStream events, type-safe PKL config, OpenTelemetry, a Cobra CLI, and cron workers, all wired end-to-end under a `go.work` monorepo.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![GraphQL](https://img.shields.io/badge/GraphQL-gqlgen-E10098?logo=graphql&logoColor=white)](https://gqlgen.com)
[![Ent](https://img.shields.io/badge/ORM-Ent-179BD7)](https://entgo.io)
[![NATS](https://img.shields.io/badge/Events-NATS%20JetStream-27AAE1?logo=natsdotio&logoColor=white)](https://nats.io)
[![PostgreSQL](https://img.shields.io/badge/DB-PostgreSQL-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/license-Proprietary-7C3AED)](LICENSE)

</div>

---

## Why Volt API?

Most Go "starters" hand you a router and a `main.go` and leave the hard parts — a GraphQL layer, an ORM with migrations, config, events, tracing, a CLI, background jobs — for you to wire yourself. Every project re-solves the same plumbing before writing a single line of business logic.

**Volt API wires all of it end-to-end so you start on features in minutes, not plumbing.** GraphQL resolvers map straight onto Ent types, config is type-checked at compile time, NATS handlers slot into the app lifecycle, and tracing/logging are on by default with trace correlation. It's the same batteries-included philosophy as [Volt Web](https://github.com/khanakia/boilerplate-tanstack), for the backend.

| | Bare `net/http` starter | Framework scaffold | **Volt API** |
|---|:---:|:---:|:---:|
| Schema-first GraphQL (gqlgen) | ❌ | ~ | ✅ |
| ORM + typed migrations (Ent) | ❌ | ~ | ✅ |
| Type-safe config (PKL → Go) | ❌ | ❌ | ✅ |
| Event-driven (NATS JetStream) | ❌ | ❌ | ✅ |
| Tracing + structured logs (OTel + Zap) | ❌ | ~ | ✅ |
| Separate API / cron / CLI binaries | ❌ | ~ | ✅ |
| `go.work` monorepo + shared utils | ❌ | ❌ | ✅ |

## Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development](#development)
- [API Endpoints](#api-endpoints)
- [Docker](#docker)
- [Architecture](#architecture)
- [Dependencies](#dependencies)
- [FAQ](#faq)
- [License](#license)

## Features

* **Schema-first GraphQL** — dashboard API via gqlgen with Ent types auto-mapped; edit the `.graphqls`, regenerate resolvers.
* **Ent ORM** — typed schema, a shared `BaseMixin`, auto-migrations, and pgvector support.
* **Event-Driven** — NATS JetStream for async messaging and real-time features, registered in the app bootstrap lifecycle.
* **Type-Safe Config** — PKL configuration language with generated Go bindings; no raw `os.Getenv` in app code.
* **Observability** — OpenTelemetry tracing, structured Zap logging with trace correlation, and error tracking.
* **Code Generation** — Ent ORM and gqlgen for type-safe database and API code.
* **Separate Binaries** — independent API, cron, and CLI processes under one `go.work` monorepo.

## Prerequisites

* Go 1.25+
* PostgreSQL 15+
* NATS Server with JetStream
* Redis (optional, for caching)
* [Task](https://taskfile.dev/) (task runner)
* [PKL CLI](https://pkl-lang.org/) (for configuration)

## Project Structure

```
.
├── modules/
│   ├── saas/       # Core business logic, CLI, API + cron bootstrap
│   ├── apidash/    # Dashboard GraphQL API (/api/sa)
│   ├── dbent/      # Ent ORM schema definitions
│   ├── config/     # PKL configuration schemas and environments
│   ├── lace/       # Shared utility library (read-only)
│   ├── core/       # External core module (read-only)
│   ├── scripts/    # Deploy/ops scripts (read-only)
│   └── examples/   # Example code
├── .docker/        # Docker and development infrastructure
├── brand/          # Brand kit (identity, assets, tokens)
└── go.work         # Go workspace linking the modules
```

## Getting Started

### 1. Clone and Install Dependencies

```bash
git clone git@github.com:khanakia/boilerplate-golang.git
cd boilerplate-golang
task workvendor
```

### 2. Configure Environment

Copy the sample environment file and configure your settings:

```bash
cp env.sample .env
```

Key environment variables:

| Variable | Description |
|----------|-------------|
| `CONFIG_ENV` | Configuration environment (`default`, `prod`, `sample`) |
| `PORT` | API port (also used by Docker Compose) |
| `DB_HOST`, `DB_PORT`, `DB_NAME` | PostgreSQL connection |
| `DB_USERNAME`, `DB_PASSWORD` | Database credentials |
| `NATS_URL` | NATS server URL |
| `REDIS_URL` | Redis connection URL |
| `JWT_SECRET` | Secret for JWT token signing |
| `INTERNAL_KEY` | Key for internal API authentication |

### 3. Start Infrastructure

```bash
task dev
```

This starts the local stack (PostgreSQL, NATS) and runs the API with live reload.

### 4. Run Database Migrations

```bash
cd modules/saas && go run ./cmd/cli migrate
```

### 5. Start the API Server

```bash
cd modules/saas && go run ./cmd/api
```

The server listens on the port from your config (`PORT`, e.g. `2310`).

## Development

### Task Commands

```bash
task -l                 # List all available tasks
task workvendor         # Sync workspace dependencies
task config:gen         # Regenerate PKL config bindings
task dbent:g            # Regenerate Ent ORM code
task dbent:gm           # Regenerate Ent code and migrate the DB
task gql                # Regenerate all GraphQL resolvers
task lint               # gofmt + go vet on all modules
task test               # Run tests
task dev                # Start local infra + API with live reload
```

### Code Generation

After modifying schemas, regenerate code:

```bash
# After Ent schema changes (modules/dbent/schema/)
task dbent:g

# After GraphQL schema changes (*.graphqls)
task gql

# After PKL config changes (modules/config/pkl/)
task config:gen
```

### CLI Commands

```bash
cd modules/saas
go run ./cmd/cli ping                    # Health check
go run ./cmd/cli migrate                 # Run migrations
```

## API Endpoints

| Endpoint | Description | Auth |
|----------|-------------|------|
| `/api/sa` | Dashboard GraphQL API | Internal Key |

## Docker

### Development

```bash
docker compose -f .docker/docker-compose.yaml up
```

### Production Build

```bash
docker build -f .docker/Dockerfile.prod -t app .
```

See `task -l | grep docker` for the full production, local-prod, and CI (`act`) task set.

## Architecture

```
GraphQL Resolvers (modules/apidash)
         │
         ▼
    App Plugin (modules/saas/pkg/app)
         │
         ▼
  Domain Services (modules/saas/pkg/*)
         │
         ▼
    Ent ORM Client
         │
         ▼
     PostgreSQL
```

**Key Principles:**

* Business logic lives in `modules/saas/pkg/`, not in resolvers.
* Never edit generated code in `gen/`, `generated/`, or `ent/` directories.
* Use `app.New()` for database access.
* Run code generation via Taskfile commands.

## Dependencies

Core technologies:

* **HTTP**: Gin
* **GraphQL**: gqlgen
* **ORM**: Ent
* **Database**: PostgreSQL with pgvector
* **Messaging**: NATS JetStream
* **Caching**: Redis
* **Config**: PKL
* **Observability**: OpenTelemetry, Zap

## FAQ

**Is Volt API related to Volt Web?**
Yes — they're the same brand family. [Volt Web](https://github.com/khanakia/boilerplate-tanstack) is the batteries-included TanStack Start React frontend; Volt API is the Go backend. Same identity, same "batteries-included starter" philosophy.

**Why GraphQL instead of REST?**
The dashboard API is schema-first via gqlgen, with Ent entity types auto-mapped — you get a typed, introspectable API without hand-writing marshaling. You can still add REST handlers via Gin where it fits.

**Why PKL for config instead of env vars?**
PKL is type-checked and compiles to Go bindings, so config errors are caught at build time rather than at runtime. Secrets still come from the environment via PKL's `read("env:VAR")`.

**Why a `go.work` monorepo with separate binaries?**
The API, cron, and CLI are independent processes that share the same domain code and a common `lace/` utility library. The workspace keeps them in one repo without coupling them into a single binary.

**Can I remove the modules I don't need?**
Yes. `modules/saas`, `modules/apidash`, `modules/dbent`, and `modules/config` are the writable core; `lace/`, `core/`, and `scripts/` are shared/external. Trim the GraphQL, NATS, or cron pieces you don't use.

**What database is supported?**
PostgreSQL (with pgvector for embeddings). Ent can target other databases, but the migrations and schema here assume Postgres.

## License

**Proprietary — all rights reserved.** See [LICENSE](LICENSE). This is commercial, private software; it is not open source. Contact the author for licensing.

---

<div align="center">
<sub>

**Volt API** — Go backend boilerplate · GraphQL · gqlgen · Ent ORM · NATS JetStream · PKL config · OpenTelemetry · PostgreSQL · pgvector · Gin · Zap · Cobra CLI · cron · `go.work` monorepo · type-safe · code generation · schema-first · event-driven · production-ready starter kit

</sub>
</div>
