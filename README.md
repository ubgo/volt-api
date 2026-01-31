# Go Boilerplate

A modular Go backend boilerplate featuring GraphQL APIs, event-driven architecture, and code generation.

## Features

* **Multi-API Architecture** - GraphQL API for dashboard endpoints
* **Event-Driven** - NATS JetStream for async messaging and real-time features
* **Type-Safe Config** - PKL configuration language with Go bindings
* **Observability** - OpenTelemetry tracing, structured logging, and error tracking
* **Code Generation** - Ent ORM and gqlgen for type-safe database and API code

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
├── saas/           # Core business logic, CLI, API bootstrap
├── apidash/        # Dashboard GraphQL API (/api/sa)
├── dbent/          # Ent ORM schema definitions
├── config/         # PKL configuration schemas and environments
├── lace/           # Shared utility library
├── scripts/        # Helper scripts
└── .docker/        # Docker and development infrastructure
```

## Getting Started

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd golang_boilerplate
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

This starts PostgreSQL and NATS via Docker Compose.

### 4. Run Database Migrations

```bash
cd saas && go run ./cmd/dbmigrate
```

### 5. Start the API Server

```bash
cd saas && go run ./cmd/api/main.go
```

The server starts on port `9001` by default.

## Development

### Task Commands

```bash
task -l                 # List all available tasks
task workvendor         # Sync workspace dependencies
task config             # Regenerate PKL config bindings
task dbent:g          # Regenerate Ent ORM code
task gql                # Regenerate all GraphQL resolvers
task lint               # Format and vet code
task test               # Run tests
task dev                # Start Docker infrastructure
```

### Code Generation

After modifying schemas, regenerate code:

```bash
# After Ent schema changes (dbent/schema/)
task saas:entg

# After GraphQL schema changes (*.graphqls)
task gql

# After PKL config changes
task config
```

### CLI Commands

```bash
cd saas
go run ./cmd/cli ping                    # Health check
go run ./cmd/cli migrate                 # Run migrations
```

## API Endpoints

| Endpoint | Description | Auth |
|----------|-------------|------|
| `/api/sa` | Dashboard API | Internal Key |

## Docker

### Development

```bash
docker-compose -f .docker/docker-compose.yaml up
```

### Production Build

```bash
docker build -f .docker/Dockerfile.prod -t app .
```

## Architecture

```
GraphQL Resolvers (apidash)
         │
         ▼
    App Plugin (saas/pkg/app)
         │
         ▼
  Domain Services (saas/pkg/*)
         │
         ▼
    Ent ORM Client
         │
         ▼
     PostgreSQL
```

**Key Principles:**

* Business logic lives in `saas/pkg/`, not in resolvers
* Never edit generated code in `gen/`, `generated/`, or `ent/` directories
* Use `app.New()` for database access
* Run code generation via Taskfile commands

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

## License

MIT
