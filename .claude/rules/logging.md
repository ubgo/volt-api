# Logging Rules

**Read before:** Adding logging to your code.

## Checklist

- [ ] Use `gozap` package for structured logging
- [ ] Include context for trace correlation
- [ ] Use appropriate log level
- [ ] Add structured fields, not string interpolation

## Import

```go
import (
    "lace/gozap"
    "go.uber.org/zap"
)
```

## Do

- Use `gozap.FromCtx(ctx)` when you have context (preserves trace IDs)
- Use `gozap.GetLogger()` when no context available
- Add structured fields with `zap.String()`, `zap.Int()`, etc.
- Log at appropriate levels (Info, Error, Warn, Debug)
- Log at boundaries (handlers, job start/end), not every function

## Don't

- Use `fmt.Printf` or `log.Println` for application logs
- Log sensitive data (passwords, tokens, PII)
- Log AND return errors (double logging)
- Use string formatting in log messages

## Basic Usage

```go
// With context (preferred - includes trace_id, span_id)
gozap.FromCtx(ctx).Info("operation completed",
    zap.String("user_id", userID),
    zap.Int("count", count),
)

// Without context
gozap.GetLogger().Info("server started",
    zap.String("port", port),
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
gozap.FromCtx(ctx).Debug("cache lookup", zap.String("key", key))
gozap.FromCtx(ctx).Info("user created", zap.String("id", user.ID))
gozap.FromCtx(ctx).Warn("rate limit approaching", zap.Int("current", count))
gozap.FromCtx(ctx).Error("database query failed", zap.Error(err))
```

## Structured Fields

```go
// Good - structured fields
gozap.FromCtx(ctx).Info("order processed",
    zap.String("order_id", order.ID),
    zap.Float64("amount", order.Amount),
    zap.Duration("processing_time", elapsed),
    zap.Bool("express", order.IsExpress),
)

// Bad - string interpolation
gozap.FromCtx(ctx).Info(fmt.Sprintf("order %s processed for $%.2f", order.ID, order.Amount))
```

## Error Logging

```go
// At boundary (handler) - log the error
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(ctx, id)
    if err != nil {
        gozap.FromCtx(ctx).Error("get user failed",
            zap.Error(err),
            zap.String("user_id", id),
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

## Sugar Logger (for quick prototyping)

```go
// Less type-safe but more convenient
gozap.GetLogger().Sugar().Infof("user %s logged in", userID)
gozap.GetLogger().Sugar().Errorw("operation failed",
    "error", err,
    "user_id", userID,
)
```

## Common Field Types

| Field Type | Example |
|------------|---------|
| `zap.String` | `zap.String("id", "123")` |
| `zap.Int` | `zap.Int("count", 42)` |
| `zap.Error` | `zap.Error(err)` |
| `zap.Duration` | `zap.Duration("elapsed", time.Since(start))` |
| `zap.Bool` | `zap.Bool("success", true)` |
| `zap.Any` | `zap.Any("data", complexStruct)` |
| `zap.Strings` | `zap.Strings("tags", []string{"a", "b"})` |

## Trace Correlation

When using `gozap.FromCtx(ctx)`, trace_id and span_id are automatically added:

```json
{
  "level": "info",
  "msg": "user created",
  "trace_id": "abc123",
  "span_id": "def456",
  "user_id": "usr_789"
}
```

This enables log correlation in observability tools.
