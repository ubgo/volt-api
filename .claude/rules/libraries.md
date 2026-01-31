# Preferred Libraries

**Read before:** Adding new dependencies or suggesting alternatives.

## Rule

**Always use these libraries.** Do not suggest alternatives unless there's a specific limitation.

## Core Framework

| Purpose | Library | Notes |
|---------|---------|-------|
| HTTP Server | `github.com/gin-gonic/gin` | REST + middleware |
| GraphQL | `github.com/99designs/gqlgen` | Code-first, type-safe |
| ORM | `entgo.io/ent` | Schema-first, codegen |
| CLI | `github.com/spf13/cobra` | Commands + flags |
| Config | PKL (Apple) | Type-safe, codegen |

## Database & Cache

| Purpose | Library | Notes |
|---------|---------|-------|
| PostgreSQL Driver | `github.com/jackc/pgx/v5` | Primary driver |
| Redis | `github.com/redis/go-redis/v9` | Cache + sessions |
| Vector Search | `github.com/pgvector/pgvector-go` | pgvector support |

## Validation

| Purpose | Library | Notes |
|---------|---------|-------|
| **Input Validation** | `github.com/cohesivestack/valgo` | **Primary choice** |

```go
// Use valgo for business validation
import "github.com/cohesivestack/valgo"

val := valgo.Is(
    valgo.String(input.Email, "email").Not().Empty().MatchingTo(emailRegex),
    valgo.Number(input.Amount, "amount").GreaterThan(0),
)
if !val.Valid() {
    return val.Errors()
}
```

## Messaging & Events

| Purpose | Library | Notes |
|---------|---------|-------|
| Message Broker | `github.com/nats-io/nats.go` | NATS JetStream |
| Event Streaming | `github.com/ThreeDotsLabs/watermill` | Pub/sub abstraction |

## Observability

| Purpose | Library | Notes |
|---------|---------|-------|
| Tracing/Metrics | `go.opentelemetry.io/otel` | OpenTelemetry |
| Logging | `go.uber.org/zap` | Structured logs |
| Error Tracking | `github.com/getsentry/sentry-go` | Production errors |

## Scheduling

| Purpose | Library | Notes |
|---------|---------|-------|
| Cron Jobs | `github.com/go-co-op/gocron/v2` | Scheduler |

## External Services

| Purpose | Library | Notes |
|---------|---------|-------|
| AWS S3 | `github.com/aws/aws-sdk-go-v2` | File storage |
| Stripe | `github.com/stripe/stripe-go/v82` | Payments |
| OpenAI | `github.com/openai/openai-go` | AI/LLM |
| LangChain | `github.com/tmc/langchaingo` | AI orchestration |

## Utilities

| Purpose | Library | Notes |
|---------|---------|-------|
| UUID | `github.com/google/uuid` | Standard UUIDs |
| Nano ID | `github.com/matoous/go-nanoid/v2` | Short IDs |
| HTML to Markdown | `github.com/JohannesKaufmann/html-to-markdown/v2` | Content conversion |
| HTML Parsing | `github.com/PuerkitoBio/goquery` | jQuery-like |
| Type Casting | `github.com/spf13/cast` | Safe conversions |
| LRU Cache | `github.com/hashicorp/golang-lru/v2` | In-memory cache |
| JSON Schema | `github.com/invopop/jsonschema` | Schema generation |
| Tokenizer | `github.com/pandodao/tokenizer-go` | Token counting |

## Do Not Use

| Instead Of | Use |
|------------|-----|
| `encoding/json` for validation | `valgo` |
| `database/sql` directly | `ent` ORM |
| `log` package | `gozap` (wrapper around zap) |
| `net/http` directly | `gin` |
| Custom cron | `gocron/v2` |
| `jwt-go` (deprecated) | Consider updating |

## Adding New Dependencies

1. Check if functionality exists in current libraries
2. Check if `lace/` has a wrapper
3. If new library needed, discuss before adding
4. Prefer well-maintained, widely-used packages

## Wrappers in lace/

These wrap external libraries with project conventions:

| Package | Wraps | Purpose |
|---------|-------|---------|
| `lace/gozap` | zap | Logging with OTEL |
| `lace/gotel` | otel | Telemetry setup |
| `lace/natso` | nats.go | NATS connection |
| `lace/cacheredis` | go-redis | Cache interface |
| `lace/ginserver` | gin | Server setup |
| `lace/cli` | cobra | CLI setup |
| `lace/crypt` | stdlib | AES encryption |
