# Logging Rules

**Read before:** Adding logging to your code.

> **Logger:** `github.com/ubgo/logger` (aliased `ulog`), wired through `core/pkg/applog`. This replaced the old `lace/gozap`/zap wrapper — there is no `gozap` and no `zap` dependency anymore.

## Checklist

- [ ] Use `applog` for the logger + `ulog` for fields
- [ ] Prefer `applog.FromCtx(ctx)` for trace correlation
- [ ] Use an appropriate log level
- [ ] Add structured fields, not string interpolation

## Import

```go
import (
    "core/pkg/applog"

    ulog "github.com/ubgo/logger"
)
```

## Do

- Use `applog.FromCtx(ctx)` when you have context (auto-adds `trace_id` + `span_id`)
- Use `applog.GetLogger()` when no context is available
- Add structured fields with `ulog.String()`, `ulog.Err()`, `ulog.Any()`, etc.
- Log at appropriate levels (Info, Error, Warn, Debug)
- Log at boundaries (handlers, job start/end), not every function

## Don't

- Use `fmt.Printf` or `log.Println` for application logs
- Log sensitive data (passwords, tokens, PII)
- Log AND return errors (double logging)
- Use string formatting in log messages

## Basic Usage

```go
// With context (preferred — includes trace_id, span_id)
applog.FromCtx(ctx).Info("operation completed",
    ulog.String("user_id", userID),
    ulog.Any("count", count),
)

// Without context
applog.GetLogger().Info("server started",
    ulog.String("port", port),
)
```

## Log Levels

| Level | Use For |
|-------|---------|
| `Debug` | Development debugging, verbose info |
| `Info` | Normal operations, milestones |
| `Warn` | Recoverable issues, deprecations |
| `Error` | Failures that need attention |

```go
applog.FromCtx(ctx).Debug("cache lookup", ulog.String("key", key))
applog.FromCtx(ctx).Info("user created", ulog.String("id", user.ID))
applog.FromCtx(ctx).Warn("rate limit approaching", ulog.Any("current", count))
applog.FromCtx(ctx).Error("database query failed", ulog.Err(err))
```

## Structured Fields

```go
// Good - structured fields
applog.FromCtx(ctx).Info("order processed",
    ulog.String("order_id", order.ID),
    ulog.Any("amount", order.Amount),
    ulog.Dur("processing_time", elapsed),
    ulog.Bool("express", order.IsExpress),
)

// Bad - string interpolation
applog.FromCtx(ctx).Info(fmt.Sprintf("order %s processed for $%.2f", order.ID, order.Amount))
```

## Error Logging

```go
// At boundary (handler) - log the error
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(ctx, id)
    if err != nil {
        applog.FromCtx(ctx).Error("get user failed",
            ulog.Err(err),
            ulog.String("user_id", id),
        )
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }
}

// In service - just return, don't log
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.Find(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("find user: %w", err)  // Don't log here
    }
    return user, nil
}
```

## Field Constructors

`github.com/ubgo/logger` (aliased `ulog`) exposes a small, typed set of field constructors:

| Field | Example | Notes |
|------------|---------|-------|
| `ulog.String` | `ulog.String("id", "123")` | |
| `ulog.Bool` | `ulog.Bool("success", true)` | |
| `ulog.Dur` | `ulog.Dur("elapsed", time.Since(start))` | duration (not `Duration`) |
| `ulog.Time` | `ulog.Time("at", t)` | |
| `ulog.Err` | `ulog.Err(err)` | the error field (not `Error`) |
| `ulog.NamedErr` | `ulog.NamedErr("cause", err)` | named error field |
| `ulog.Any` | `ulog.Any("count", 42)` | **use for ints, floats, slices, structs** — there are no `Int`/`Float`/`Strings` constructors |

> There is no sugared logger (`.Sugar().Infof(...)`). Use the typed methods with `ulog.*` fields, or `ulog.Any` for arbitrary values.

## Trace Correlation

When using `applog.FromCtx(ctx)`, `trace_id` and `span_id` are automatically added from the active span:

```json
{
  "level": "info",
  "msg": "user created",
  "trace_id": "abc123",
  "span_id": "def456",
  "user_id": "usr_789"
}
```

This enables log correlation in observability tools. The logger's OTEL sink is initialized in `applog.Init` (called during bootstrap after the OpenTelemetry providers are registered), so these logs also ship through the OTEL pipeline.
