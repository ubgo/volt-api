# Getting started

The fastest path from clone to a running API and your first GraphQL response. For the full setup — every prerequisite, the portless HTTPS workflow, and deployment — see [CONTRIBUTING.md](../CONTRIBUTING.md).

## Prerequisites

Minimum to run: **Go 1.25+**, **PostgreSQL 15+**, **NATS (JetStream)**, **[Task](https://taskfile.dev)**, and the **[PKL CLI](https://pkl-lang.org)**. Recommended for a nicer loop: **air** (hot reload) and **portless** (stable HTTPS URL). Docker can supply Postgres + NATS if you'd rather not install them. Full list with install links: [CONTRIBUTING.md#prerequisites](../CONTRIBUTING.md#prerequisites).

## Five steps to a running API

```bash
# 1. Sync the go.work workspace
task workvendor

# 2. Start Postgres + NATS locally (or point at your own)
task docker:up

# 3. Configure — copy the sample env; PKL reads secrets from it via read("env:VAR")
cp env.sample .env
#    Set PORT (default 2310) and CONFIG_ENV (default). DB/NATS defaults match `task docker:up`.

# 4. Create the database + run migrations
PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 -c "CREATE DATABASE appdb;"
task config:gen        # generate the typed PKL config bindings
task dbent:migrate     # apply the Ent schema to the database

# 5. Run the API (pick one)
task dev               # air hot reload → http://localhost:2310
task dev:portless      # ★ stable HTTPS → https://api.app.localhost  (needs portless)
```

## Your first GraphQL call

The dashboard API is served at **`/api/sa`** and guarded by the internal key (`INTERNAL_KEY` in your `.env`, sent as the `X-Internal-Key` header). Confirm the service is healthy first:

```bash
curl -H "X-Internal-Key: $INTERNAL_KEY" http://localhost:2310/api/i/bootinfo
# → status of database, redis, nats, otel
```

Then open the GraphQL playground (key-gated) or POST a query:

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
