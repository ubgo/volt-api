# Volt API documentation

Everything you need to build on **Volt API** — the batteries-included Go backend starter. These are the framework docs; they live under `volt/` so they stay separable from your own app's docs and update as one unit with the Volt layer.

New here? Start with **[Getting started](getting-started.md)**, then skim **[Architecture](architecture.md)**.

## Find your way by intent

### 🚀 I want to get running
- [Getting started](getting-started.md) — clone → configure → run → first GraphQL call.
- [Full setup & dev workflow](../CONTRIBUTING.md) — every prerequisite, the portless HTTPS loop, all commands.
- [CLI reference](reference/cli.md) — the `cli` binary's commands.

### 🧭 I want to understand how it fits together
- [Architecture](architecture.md) — modules, binaries, request flow, bootstrap.
- [Project structure](project-structure.md) — the directory map.
- [Stack](stack.md) — the technology choices.
- [ubgo packages](ubgo-packages.md) — the library ecosystem it's built on (wired-in + available to add).

### 🛠️ I want to build a feature
- [GraphQL API](guides/graphql-api.md) — add a query/mutation (schema-first, gqlgen).
- [Ent CRUD](guides/ent-crud.md) — generate CRUD from an Ent schema.
- [GraphQL extension](guides/graphql-extension.md) — extend a module's type from the host.
- [Cron jobs](guides/cron-jobs.md) — scheduled background work.
- [Observability](guides/observability.md) — logging, tracing, and the telemetry bootstrap.
- [Backend patterns](guides/backend-patterns.md) — the condensed Go cheat-sheet.
- [Go conventions](guides/go-conventions.md) · [pgvector notes](guides/pgvector.md) · [SQLite mode](guides/sqlite-mode.md) · [GraphQL testing](guides/graphql-testing.md) · [Workflow](guides/workflow.md).

> Building a feature? The **[`../rules/`](../rules/README.md)** folder has terse, action-triggered checklists (read `new-endpoint.md` before adding an endpoint, etc.) — and [`../rules/code-standards.md`](../rules/code-standards.md) is the standing engineering bar. These guides are the narrative; the rules are the checklist.

### 🚢 I want to deploy & operate
- [Deployment overview](deployment/README.md) — the full flow.
- [Docker](deployment/docker.md) — production image + persistent storage.
- [GitHub Actions](deployment/github-actions.md) · [VPS access](deployment/vps-access.md) · [NetBird](deployment/netbird.md) · [Secrets](deployment/secrets.md).

### 📉 I want to learn from what broke
- [Incidents](incidents/) — postmortems.

## Map

```
volt/docs/
├── getting-started.md · architecture.md · project-structure.md · stack.md · ubgo-packages.md
├── guides/       ← task how-tos (narrative; ../rules/ is the agent checklist)
├── deployment/   ← Docker, GitHub Actions, VPS, NetBird, secrets
├── reference/    ← CLI reference
└── incidents/    ← postmortems
```

The `framework/` folder is external reference — do not edit or grep it.
