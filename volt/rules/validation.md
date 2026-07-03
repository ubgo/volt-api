# Validation Rules

**Read before:** Adding input validation.

> **Primary library:** `github.com/cohesivestack/valgo` - Use for all business validation.

## Checklist

- [ ] Validate at API boundary (handlers/resolvers)
- [ ] Use Valgo for business validation
- [ ] Return user-friendly error messages
- [ ] Don't trust client-provided data

## Validation Locations

| Layer | What to Validate |
|-------|-----------------|
| GraphQL schema | Type constraints (required, types) |
| Resolver | Business rules, authorization |
| Service | Domain invariants |
| Ent schema | Database constraints (unique, not null) |

## Do

- Validate early (at API boundary)
- Return specific error messages
- Validate file content by magic bytes, not extension
- Sanitize HTML/user content

## Don't

- Trust client-provided MIME types
- Skip validation for "internal" APIs
- Expose internal errors to users

## Gin Handler Validation

```go
type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=8"`
    FirstName string `json:"firstName" binding:"required,max=100"`
    Age       int    `json:"age" binding:"omitempty,gte=0,lte=150"`
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // Validation passed, proceed
}
```

## Valgo Validation (Primary)

```go
import "github.com/cohesivestack/valgo"

func validateOrder(input CreateOrderInput) error {
    val := valgo.Is(
        valgo.String(input.CustomerID, "customer_id").Not().Empty(),
        valgo.Number(input.Amount, "amount").GreaterThan(0),
        valgo.String(input.Currency, "currency").InSlice([]string{"USD", "EUR", "GBP"}),
        valgo.Slice(input.Items, "items").Not().Empty(),
    )

    if !val.Valid() {
        return fmt.Errorf("validation failed: %v", val.Errors())
    }
    return nil
}
```

## Valgo Common Validators

```go
// String validators
valgo.String(s, "field").Not().Empty()
valgo.String(s, "field").MinLength(3).MaxLength(100)
valgo.String(s, "field").MatchingTo(regexp.MustCompile(`^[a-z]+$`))
valgo.String(s, "field").InSlice([]string{"a", "b", "c"})

// Number validators
valgo.Number(n, "field").GreaterThan(0)
valgo.Number(n, "field").Between(1, 100)
valgo.Number(n, "field").LessThanOrEqualTo(1000)

// Slice validators
valgo.Slice(items, "field").Not().Empty()
valgo.Slice(items, "field").MaxLength(10)

// Boolean
valgo.Bool(b, "field").True()

// Pointer (optional fields)
valgo.StringP(ptr, "field").Not().Empty()  // validates if not nil
```

## Valgo Error Handling

```go
val := valgo.Is(
    valgo.String(input.Email, "email").Not().Empty(),
    valgo.Number(input.Age, "age").GreaterThan(0),
)

if !val.Valid() {
    // Get all errors as map
    errors := val.Errors()
    // {"email": ["Email is required"], "age": ["Age must be greater than 0"]}

    // Or iterate
    for field, msgs := range errors {
        fmt.Printf("%s: %v\n", field, msgs)
    }
}
```

## GraphQL Resolver Validation

```go
func (r *mutationResolver) CreateOrder(ctx context.Context, input model.CreateOrderInput) (*ent.Order, error) {
    // Validate business rules
    if input.Amount <= 0 {
        return nil, fmt.Errorf("amount must be positive")
    }

    if len(input.Items) == 0 {
        return nil, fmt.Errorf("order must have at least one item")
    }

    // Check authorization
    user := auth.UserFromContext(ctx)
    if user == nil {
        return nil, fmt.Errorf("unauthorized")
    }

    // Proceed with creation
    return r.Circuit.EntDB.Client().Order.Create()...
}
```

## File Upload Validation

```go
import (
    "encoding/base64"
    "net/http"
)

func validateFileUpload(input FileUploadInput) error {
    // Decode base64
    content, err := base64.StdEncoding.DecodeString(input.Content)
    if err != nil {
        return fmt.Errorf("invalid base64 content")
    }

    // Check size
    maxSize := 10 * 1024 * 1024  // 10MB
    if len(content) > maxSize {
        return fmt.Errorf("file too large (max 10MB)")
    }

    // Validate by magic bytes, NOT filename
    mime := http.DetectContentType(content)
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
        "application/pdf": true,
    }

    if !allowedTypes[mime] {
        return fmt.Errorf("file type not allowed: %s", mime)
    }

    return nil
}
```

## Common Validation Patterns

```go
// Email
import "net/mail"
func isValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

// UUID
import "github.com/google/uuid"
func isValidUUID(s string) bool {
    _, err := uuid.Parse(s)
    return err == nil
}

// URL
import "net/url"
func isValidURL(s string) bool {
    u, err := url.Parse(s)
    return err == nil && u.Scheme != "" && u.Host != ""
}
```

## Error Response Format

```go
// Consistent error response
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (h *Handler) handleValidationErrors(c *gin.Context, errors []ValidationError) {
    c.JSON(400, gin.H{
        "error": "validation_failed",
        "details": errors,
    })
}
```

## Ent Schema Validation

```go
// In dbent/schema/user.go
func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").
            NotEmpty().
            Unique().
            Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)),
        field.String("name").
            NotEmpty().
            MaxLen(100),
        field.Int("age").
            Positive().
            Max(150),
    }
}
```
