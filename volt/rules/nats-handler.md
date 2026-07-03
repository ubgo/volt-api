# NATS Handler Rules

**Read before:** Adding NATS message handlers.

## Checklist

- [ ] Define handler function
- [ ] Register after NATS initialized (use `WithAfterNatsInitiated`)
- [ ] Add tracing for observability
- [ ] Handle errors gracefully (don't panic)

## File Locations

| Type | Location |
|------|----------|
| Handler registration | `webapp/cmd/api/main.go` (in `WithAfterNatsInitiated`) |
| Handler functions | `webapp/pkg/<feature>/nats_handlers.go` |
| NATS utilities | `lace/natso/`, `lace/natsutil/` |

## Do

- Register handlers in `WithAfterNatsInitiated` callback
- Add OpenTelemetry tracing to handlers
- Use structured logging
- Acknowledge messages appropriately

## Don't

- Register handlers before NATS is connected
- Panic in handlers (will crash subscriber)
- Block handlers for too long without processing

## Handler Registration Pattern

```go
// webapp/cmd/api/main.go
func main() {
    circuit := webapp.New(
        webapp.WithAfterNatsInitiated(func(p *volt.Kernel) {
            registerNatsHandlers(p)
        }),
    )
    // ...
}

func registerNatsHandlers(plugin *volt.Kernel) {
    nc := plugin.Natso.GetConn()
    js, _ := nc.JetStream()

    // Subscribe to subject
    _, err := js.Subscribe("orders.created", func(msg *nats.Msg) {
        handleOrderCreated(plugin, msg)
    }, nats.Durable("order-processor"))

    if err != nil {
        applog.GetLogger().Error("failed to subscribe", ulog.Err(err))
    }
}
```

## Handler with Tracing

```go
func handleOrderCreated(plugin *volt.Kernel, msg *nats.Msg) {
    ctx, span := otel.Tracer("nats").Start(
        context.Background(),
        "nats.orders.created",
        trace.WithAttributes(
            attribute.String("subject", msg.Subject),
        ),
    )
    defer span.End()

    span.AddEvent("message_received")

    // Parse message
    var order OrderEvent
    if err := json.Unmarshal(msg.Data, &order); err != nil {
        span.SetStatus(codes.Error, "invalid message")
        span.RecordError(err)
        applog.FromCtx(ctx).Error("failed to parse message", ulog.Err(err))
        msg.Nak()  // Negative acknowledgment
        return
    }

    // Process
    if err := processOrder(ctx, plugin, &order); err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        applog.FromCtx(ctx).Error("failed to process order", ulog.Err(err))
        msg.Nak()
        return
    }

    span.SetStatus(codes.Ok, "success")
    span.AddEvent("message_processed")
    msg.Ack()  // Acknowledge success
}
```

## Publishing Messages

```go
func publishOrderEvent(ctx context.Context, plugin *volt.Kernel, order *Order) error {
    nc := plugin.Natso.GetConn()
    js, err := nc.JetStream()
    if err != nil {
        return fmt.Errorf("get jetstream: %w", err)
    }

    data, err := json.Marshal(OrderEvent{
        ID:        order.ID,
        Action:    "created",
        Timestamp: time.Now(),
    })
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }

    _, err = js.Publish("orders.created", data)
    if err != nil {
        return fmt.Errorf("publish event: %w", err)
    }

    return nil
}
```

## JetStream Patterns

```go
// Durable subscription (survives restarts)
js.Subscribe("subject", handler, nats.Durable("consumer-name"))

// Queue group (load balancing)
js.QueueSubscribe("subject", "queue-group", handler, nats.Durable("consumer"))

// Pull-based consumer
sub, _ := js.PullSubscribe("subject", "consumer")
msgs, _ := sub.Fetch(10, nats.MaxWait(5*time.Second))
```

## Message Acknowledgment

| Method | When to Use |
|--------|-------------|
| `msg.Ack()` | Successfully processed |
| `msg.Nak()` | Failed, retry immediately |
| `msg.Term()` | Failed, don't retry (poison message) |
| `msg.InProgress()` | Still processing, extend timeout |

## Event Types

```go
// webapp/pkg/events/types.go
type OrderEvent struct {
    ID        string    `json:"id"`
    Action    string    `json:"action"`  // created, updated, deleted
    Timestamp time.Time `json:"timestamp"`
    Data      any       `json:"data,omitempty"`
}
```

## Check NATS Status

```go
// Check if NATS is connected
if plugin.Natso != nil && plugin.Natso.GetConn().IsConnected() {
    // NATS is ready
}
```

Or via bootinfo endpoint: `GET /api/i/bootinfo`
