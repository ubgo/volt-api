# Tech Stack

## Core

| Purpose | Library |
|---------|---------|
| Language | Go 1.25 |
| Workspace | `go.work` monorepo |
| HTTP Server | `github.com/gin-gonic/gin` |
| GraphQL | `github.com/99designs/gqlgen` |
| ORM | `entgo.io/ent` |
| CLI | `github.com/spf13/cobra` |
| Config | PKL (Apple) via `pkl-go` |

## Data

| Purpose | Library |
|---------|---------|
| PostgreSQL driver | `github.com/jackc/pgx/v5` |
| Redis | `github.com/redis/go-redis/v9` |
| pgvector | `github.com/pgvector/pgvector-go` |

## Messaging

| Purpose | Library |
|---------|---------|
| NATS / JetStream | `github.com/nats-io/nats.go` |
| Event streaming | `github.com/ThreeDotsLabs/watermill` |

## Observability

| Purpose | Library |
|---------|---------|
| Tracing + metrics | `go.opentelemetry.io/otel` |
| Structured logging | `github.com/ubgo/logger` (via `core/pkg/applog`) |
| Error tracking | `github.com/getsentry/sentry-go` |

## Validation & Utilities

| Purpose | Library |
|---------|---------|
| Input validation | `github.com/cohesivestack/valgo` |
| Cron scheduling | `github.com/go-co-op/gocron/v2` |
| UUIDs | `github.com/google/uuid` |
| Short IDs | `github.com/matoous/go-nanoid/v2` |
| Type casting | `github.com/spf13/cast` |
| LRU cache | `github.com/hashicorp/golang-lru/v2` |

## External SDKs

| Purpose | Library |
|---------|---------|
| AWS S3 | `github.com/aws/aws-sdk-go-v2` |
| Stripe | `github.com/stripe/stripe-go/v82` |
| OpenAI | `github.com/openai/openai-go` |
| LangChain | `github.com/tmc/langchaingo` |

## Infrastructure

| Purpose | Tool |
|---------|------|
| Local services | Docker Compose (`task docker:up`) |
| Task runner | `task` (Taskfile.yml) |
| Container registry | GHCR (GitHub Container Registry) |
| Deployment | GitHub Actions → SSH to VPS |
| Process supervisor (prod container) | supervisord |
