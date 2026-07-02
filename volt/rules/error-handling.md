# Error Handling Rules

**Read before:** Implementing error handling in this codebase.

## Core Principles

1. **Wrap errors with context** - Always add what operation failed
2. **Return errors, don't panic** - Panics only for unrecoverable init failures
3. **Log at boundaries** - Log when handling, not when returning

## Do

- Wrap errors with `fmt.Errorf("operation: %w", err)`
- Use `%w` verb to preserve error chain
- Check errors immediately after call
- Return early on error (guard clause pattern)

## Don't

- Use `panic` for operational errors
- Ignore errors with `_`
- Log AND return the same error (double logging)
- Create custom error types unless needed for `errors.Is/As`

## Error Wrapping

```go
// Good - adds context
func GetUser(ctx context.Context, id string) (*User, error) {
    user, err := db.User.Get(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("getUser %s: %w", id, err)
    }
    return user, nil
}

// Bad - no context
func GetUser(ctx context.Context, id string) (*User, error) {
    return db.User.Get(ctx, id)  // Caller has no idea what failed
}
```

## Guard Clause Pattern

```go
func ProcessOrder(ctx context.Context, orderID string) error {
    order, err := getOrder(ctx, orderID)
    if err != nil {
        return fmt.Errorf("get order: %w", err)
    }

    if !order.IsValid() {
        return fmt.Errorf("order %s: invalid state", orderID)
    }

    if err := validateStock(ctx, order); err != nil {
        return fmt.Errorf("validate stock: %w", err)
    }

    if err := chargePayment(ctx, order); err != nil {
        return fmt.Errorf("charge payment: %w", err)
    }

    return nil
}
```

## Error Checking

```go
// Check specific error types
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}

// Check error interface
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    return nil, fmt.Errorf("validation: %s", validationErr.Field)
}
```

## Logging Errors

```go
// Log at the handler/boundary level, not in every function
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(c.Request.Context(), id)
    if err != nil {
        // Log here - at the boundary
        h.logger.Error("get user failed", ulog.Err(err), ulog.String("id", id))
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }
    c.JSON(200, user)
}

// Don't log in internal functions - just return
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.Find(ctx, id)
    if err != nil {
        // Don't log here - just wrap and return
        return nil, fmt.Errorf("find user %s: %w", id, err)
    }
    return user, nil
}
```

## Sentinel Errors

```go
// Define package-level sentinel errors for expected conditions
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

// Usage
if errors.Is(err, ErrNotFound) {
    c.JSON(404, gin.H{"error": "resource not found"})
    return
}
```
