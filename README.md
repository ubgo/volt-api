<p align="center">
  <img src="volt/brand/png/icon-512.png" alt="Volt logo — a lightning bolt on a violet tile" width="72" height="72" />
</p>

<h1 align="center">Volt API ⚡</h1>

<p align="center"><strong>The batteries-included Go backend starter.</strong></p>

<p align="center">A production-ready Go backend boilerplate for modular services — schema-first <a href="https://gqlgen.com">GraphQL</a> (gqlgen), <a href="https://entgo.io">Ent</a> ORM, NATS JetStream events, type-safe Go config via <a href="https://github.com/ubgo/cfgkit">cfgkit</a>, OpenTelemetry, a Cobra CLI, and cron workers, all wired end-to-end under a <code>go.work</code> monorepo.</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-7C3AED" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/GraphQL-gqlgen-7C3AED" alt="GraphQL via gqlgen" />
  <img src="https://img.shields.io/badge/ORM-Ent-7C3AED" alt="Ent ORM" />
  <img src="https://img.shields.io/badge/Events-NATS%20JetStream-7C3AED" alt="NATS JetStream" />
  <img src="https://img.shields.io/badge/Config-cfgkit-7C3AED" alt="Typed Go config via cfgkit" />
  <img src="https://img.shields.io/badge/coverage-2.7%25-red" alt="Hand-written test coverage (auto-updated by CI)" />
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/license-Apache--2.0-7C3AED" alt="License: Apache-2.0" /></a>
</p>

**Volt API** is a batteries-included starter for quickly bootstrapping a **modular Go backend service**. It ships with schema-first **GraphQL** (gqlgen with Ent types auto-mapped), an **Ent** ORM and typed migrations, **NATS JetStream** events, type-safe **Go configuration** (defaults compiled in, env-overridable), **OpenTelemetry** tracing and structured logging, a **Cobra** CLI, and **cron** workers — so you skip the plumbing and start on business logic in minutes.

> Volt API is the backend half of a full-stack family — pair it with **[Volt](https://github.com/khanakia/boilerplate-tanstack)** (the TanStack Start web starter) for an end-to-end stack.

> **Setting up or running Volt API?** All technical docs — prerequisites, install, the portless HTTPS dev workflow (`https://volt-api.localhost`), config, code generation, scripts, project structure, and deployment — live in **[CONTRIBUTING.md](volt/CONTRIBUTING.md)**.

## Contents

- [Why Volt API?](#why-volt-api)
- [What's inside](#whats-inside)
- [Features](#features)
- [Getting started](#getting-started)
- [FAQ](#faq)
- [License](#license)

## Why Volt API?

Most Go "starters" hand you a router and a `main.go` and leave the hard parts — a GraphQL layer, an ORM with migrations, config, events, tracing, a CLI, background jobs — for you to wire yourself. Every project re-solves the same plumbing before writing a single line of business logic.

**Volt API is that plumbing, already done.** Every layer is wired end-to-end and agrees with the others, so day one is feature work.

- **Schema-first GraphQL** — edit the `.graphqls`, run one task, and gqlgen regenerates resolvers with your Ent entity types auto-mapped. No hand-written marshaling.
- **Ent ORM** — a typed schema with a shared `BaseMixin`, auto-migrations, and pgvector support.
- **Type-safe config** — one Go file is the schema, the defaults and the validation; zero input is already valid, and secrets still come from the environment.
- **Event-driven** — NATS JetStream handlers slot into the app bootstrap lifecycle, not import-time globals.
- **Observability on by default** — OpenTelemetry tracing and structured logging with trace correlation, out of the box.
- **Separate binaries** — independent API, cron, and CLI processes sharing one domain and a `lace/` utility library, under a `go.work` monorepo.

**Who it's for:** solo developers shipping a SaaS or side project who want a full backend ready on day one, and teams that want one opinionated, consistent Go foundation across services.

## What's inside

| Area | Choice |
|---|---|
| Language | Go 1.25, `go.work` monorepo (API · cron · CLI binaries) |
| HTTP | Gin |
| GraphQL | gqlgen — schema-first, Ent types auto-mapped |
| ORM | Ent + typed migrations, `BaseMixin`, pgvector |
| Database | PostgreSQL (with pgvector) |
| Events | NATS JetStream, registered in the app lifecycle |
| Config | Typed Go structs via [cfgkit](https://github.com/ubgo/cfgkit) — defaults → `.env` → environment |
| Observability | OpenTelemetry tracing + structured logging via ubgo/logger (trace-correlated) |
| CLI | Cobra |
| Scheduling | gocron cron workers |
| Caching | Redis (optional) |
| Dev workflow | air hot reload + portless HTTPS (`https://volt-api.localhost`) |

## Features

- **Schema-first GraphQL** — dashboard API via gqlgen with Ent types auto-mapped; edit the `.graphqls`, regenerate resolvers.
- **Ent ORM** — typed schema, a shared `BaseMixin`, auto-migrations, and pgvector support for embeddings.
- **Type-safe config** — typed Go structs with no raw `os.Getenv` in app code; `CONFIG_ENV` selects which validation rules apply.
- **Event-driven** — NATS JetStream for async messaging and real-time features, wired into the bootstrap lifecycle.
- **Observability** — OpenTelemetry tracing, structured logging with trace correlation, and error tracking.
- **Code generation** — Ent ORM and gqlgen for type-safe database and API code.
- **Separate binaries** — independent API, cron, and CLI processes under one `go.work` monorepo.

## Getting started

```bash
task workvendor          # sync the go.work workspace
cp env.sample .env       # set PORT + deploy/config values
task dev                 # air hot reload (port from config/$PORT)
# or, stable HTTPS:
task dev:portless        # ★ https://volt-api.localhost  (needs portless)
```

That's the short path. For the full setup — **prerequisites** (Go, PostgreSQL, NATS, Task, portless), the **portless HTTPS dev workflow** (`task dev:portless` → `https://volt-api.localhost`), **configuration**, **code generation**, all **scripts & tasks**, the **project structure**, how to **start your real app**, and **deployment** — see **[CONTRIBUTING.md](volt/CONTRIBUTING.md)**. Deeper per-topic guides live in [`volt/docs/`](volt/docs/).

## FAQ

**Is Volt API open source?**
Yes. Volt API is open source under the [Apache License 2.0](LICENSE): you can use, modify and redistribute it, including in commercial applications, as long as you keep the license and copyright notices. See [License](#license).

**Is Volt API related to Volt (Web)?**
Yes — same brand family. [Volt](https://github.com/khanakia/boilerplate-tanstack) is the batteries-included TanStack Start React frontend; Volt API is the Go backend. Same identity, same "batteries-included starter" philosophy.

**Why GraphQL instead of REST?**
The dashboard API is schema-first via gqlgen, with Ent entity types auto-mapped — you get a typed, introspectable API without hand-writing marshaling. You can still add REST handlers via Gin where it fits.

**Why typed Go config instead of raw env vars?**
Because a bound field is more than a lookup: it carries a compiled-in default, a documentation line, a validation rule, and a `secret` flag. So the service runs with no `.env` at all, `task config:check` fails a deploy carrying a placeholder secret before the rollout, `.env.example` is generated from the struct and cannot drift, and `task config:explain` tells you which layer set each value. Secrets still come from the environment — they are just declared rather than fetched.

**Why a `go.work` monorepo with separate binaries?**
The API, cron, and CLI are independent processes that share the same domain code and a common `lace/` utility library. The workspace keeps them in one repo without coupling them into a single binary.

**What database is supported?**
PostgreSQL (with pgvector for embeddings). Ent can target other databases, but the migrations and schema here assume Postgres.

**Can I remove the parts I don't need?**
Yes. `modules/saas`, `modules/apidash`, `modules/dbent`, and `modules/config` are the writable core; `lace/`, `core/`, and `scripts/` are shared/external. Trim the GraphQL, NATS, or cron pieces you don't use.

## Making it your own

Everything Volt-specific lives under **[`volt/`](volt/)** — this pitch's preserved copy, the [dev guide](volt/CONTRIBUTING.md), the [brand kit](volt/brand/), and the [docs](volt/docs/). When you build your app on top of Volt API: replace this `README.md` and the root `LICENSE` with your own, then keep `volt/` as the boilerplate reference or delete it. The agent guide chains through the root `CLAUDE.md` → `volt/CLAUDE.md` import, so removing `volt/` means dropping that one import line too.

## License

Volt API is licensed under the [Apache License 2.0](LICENSE). You can use, modify and redistribute it, including as the foundation of your own commercial applications, provided you keep the license and copyright notices. For support or licensing questions see [SUPPORT.md](volt/SUPPORT.md).

<sub>Volt API — an open-source (Apache-2.0) Go backend boilerplate / service starter kit: schema-first GraphQL (gqlgen), Ent ORM, NATS JetStream, type-safe Go config, OpenTelemetry, structured logging, Cobra CLI, cron, PostgreSQL + pgvector, in a <code>go.work</code> monorepo.</sub>
