# Authentication Rules

**Read before:** Working with authentication/authorization.

## Checklist

- [ ] Use existing auth middleware
- [ ] Extract user from context in resolvers
- [ ] Check permissions at resolver level
- [ ] Use `@internal` directive for admin endpoints

## Key Files

| Purpose | Location |
|---------|----------|
| JWT handling | `saas/pkg/auth/` (if exists) |
| Internal key guard | `saas/pkg/apphandlers/` |
| GraphQL directive | `@internal` in schema |

## Do

- Extract user from context, don't trust client headers
- Check permissions at the start of resolvers
- Use `@internal` directive for admin-only queries/mutations
- Hash passwords before storing

## Don't

- Store plain-text passwords
- Trust user-provided IDs without verification
- Skip authorization checks
- Expose internal endpoints without protection

## GraphQL @internal Directive

For admin/internal endpoints:

```graphql
# In schema
directive @internal on FIELD_DEFINITION

type Query {
    # Public endpoint
    user(id: ID!): User

    # Internal only - requires X-Internal-Key header
    allUsers: [User!]! @internal
    systemStats: Stats! @internal
}
```

## Extracting User in Resolver

```go
func (r *queryResolver) MyProtectedQuery(ctx context.Context) (*Result, error) {
    // Get user from context (set by auth middleware)
    user := auth.UserFromContext(ctx)
    if user == nil {
        return nil, fmt.Errorf("unauthorized")
    }

    // Check specific permission
    if !user.HasPermission("read:data") {
        return nil, fmt.Errorf("forbidden")
    }

    // Proceed with authorized user
    return r.service.GetData(ctx, user.ID)
}
```

## Internal API Key Guard

For HTTP endpoints:

```go
// Already implemented in apphandlers
// Protects /api/i/* routes

// Requires header: X-Internal-Key: <value from config>
```

## Password Handling

```go
import "golang.org/x/crypto/bcrypt"

// Hash password before storing
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Verify password
func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

## JWT Pattern (if used)

```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

// Generate token
func generateToken(user *ent.User, secret string) (string, error) {
    claims := Claims{
        UserID: user.ID,
        Email:  user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// Validate token
func validateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}
```

## Middleware Pattern

```go
func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }

        // Remove "Bearer " prefix
        token = strings.TrimPrefix(token, "Bearer ")

        claims, err := validateToken(token, secret)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        // Add to context
        ctx := context.WithValue(c.Request.Context(), userContextKey, claims)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

## Context Keys

```go
type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) *Claims {
    if user, ok := ctx.Value(userContextKey).(*Claims); ok {
        return user
    }
    return nil
}
```

## Role-Based Access

```go
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
    user := auth.UserFromContext(ctx)
    if user == nil {
        return false, fmt.Errorf("unauthorized")
    }

    // Check role
    if user.Role != "admin" {
        return false, fmt.Errorf("forbidden: admin only")
    }

    // Proceed
    return r.service.DeleteUser(ctx, id)
}
```
