# Tracing Rules

**Read before:** Adding OpenTelemetry tracing.

## Checklist

- [ ] Create span with meaningful name
- [ ] Add relevant attributes
- [ ] Set status on completion
- [ ] Propagate context to child operations

## Import

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"

    "lace/gotel"
)
```

## Do

- Name spans as `<component>.<operation>` (e.g., `order.create`)
- Add business-relevant attributes
- Set status (Ok/Error) on completion
- Pass context to all downstream calls
- End spans (use `defer span.End()`)

## Don't

- Create spans for trivial operations
- Add sensitive data as attributes (passwords, tokens)
- Forget to end spans (memory leak)
- Ignore context propagation

## Basic Span

```go
func processOrder(ctx context.Context, orderID string) error {
    ctx, span := otel.Tracer("order").Start(ctx, "order.process")
    defer span.End()

    span.SetAttributes(
        attribute.String("order.id", orderID),
    )

    // Do work with ctx
    err := doWork(ctx)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        return err
    }

    span.SetStatus(codes.Ok, "success")
    return nil
}
```

## Using gotel Helper

```go
func myOperation(ctx context.Context) error {
    telem := gotel.MustGetTelemetry()
    ctx, span := telem.TraceStart(ctx, "myOperation")
    defer span.End()

    // ... operation
    return nil
}
```

## Span Naming Convention

| Pattern | Example |
|---------|---------|
| `<service>.<operation>` | `order.create`, `user.authenticate` |
| `<layer>.<entity>.<action>` | `db.user.query`, `http.order.get` |
| `cron.<job>` | `cron.cleanup`, `cron.sync` |
| `nats.<subject>` | `nats.orders.created` |

## Adding Attributes

```go
span.SetAttributes(
    // Business context
    attribute.String("user.id", userID),
    attribute.String("order.id", orderID),
    attribute.Float64("order.amount", amount),

    // Technical context
    attribute.Int("db.rows_affected", count),
    attribute.String("cache.status", "hit"),  // or "miss"
    attribute.Bool("feature.enabled", true),
)
```

## Events

```go
// Mark significant points
span.AddEvent("validation_passed")
span.AddEvent("payment_initiated", trace.WithAttributes(
    attribute.String("payment.method", "credit_card"),
))
span.AddEvent("job_completed")
```

## Error Recording

```go
if err != nil {
    span.SetStatus(codes.Error, err.Error())
    span.RecordError(err, trace.WithAttributes(
        attribute.String("error.type", "validation"),
    ))
    return err
}
```

## Child Spans

```go
func parentOperation(ctx context.Context) error {
    ctx, span := otel.Tracer("app").Start(ctx, "parent")
    defer span.End()

    // Child operation automatically linked
    if err := childOperation(ctx); err != nil {
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "success")
    return nil
}

func childOperation(ctx context.Context) error {
    ctx, span := otel.Tracer("app").Start(ctx, "child")
    defer span.End()

    // Work here
    return nil
}
```

## HTTP Handler Tracing

Already handled by middleware in `saas/pkg/httppkg/middlewares`:

```go
// Middleware adds span automatically
// Access span in handler:
span := trace.SpanFromContext(c.Request.Context())
span.SetAttributes(attribute.String("custom.attr", "value"))
```

## Database Query Tracing

Ent operations automatically traced when using context:

```go
// Context propagates trace
user, err := client.User.Get(ctx, id)  // Traced automatically
```

## Tracer Names

Use consistent tracer names per component:

```go
const tracerName = "order"  // or "cron", "nats", "api"

otel.Tracer(tracerName).Start(ctx, "operation")
```

## Flush Before Exit

For CLI commands, allow time for traces to export:

```go
// At end of CLI command
time.Sleep(2 * time.Second)  // or use proper shutdown

// Or explicit flush
gotel.MustGetTelemetry().Shutdown(ctx)
```
