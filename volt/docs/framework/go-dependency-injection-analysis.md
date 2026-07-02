# Go Dependency Injection: Complete Analysis and Best Practices

## Executive Summary

This document provides a comprehensive analysis of dependency injection (DI) approaches in Go, comparing various methods and recommending optimal approaches for building maintainable, testable, and debuggable applications.

**Primary Recommendation:** Interfaces + Constructor Injection + Clean Package Structure

This approach maximizes loose coupling, debuggability, testability, flexibility, extensibility, maintainability, modularity, and reusability with compile-time safety.

**Valid Alternatives:** Runtime DI containers (Astra, Dig, Fx) are legitimate choices for teams familiar with DI patterns from other languages, or for large applications where manual wiring becomes tedious. These containers ARE debuggable and testable—the main tradeoff is runtime vs compile-time error detection.

---

## Table of Contents

1. [Goals and Priorities](#goals-and-priorities)
2. [DI Approaches Analyzed](#di-approaches-analyzed)
3. [Detailed Comparison](#detailed-comparison)
4. [Why Not Context-Based DI](#why-not-context-based-di)
5. [Runtime DI Containers: Dig, Fx, and Astra](#runtime-di-containers-dig-fx-and-astra)
6. [The Recommended Approach](#the-recommended-approach)
7. [Implementation Guide](#implementation-guide)
8. [Testing Strategy](#testing-strategy)
9. [Advanced Patterns](#advanced-patterns)
10. [Comparison Tables](#comparison-tables)

---

## Goals and Priorities

The following goals were prioritized (in order of importance):

| Priority | Goal | Description |
|----------|------|-------------|
| 1 | **Loose Coupling** | Components should depend on abstractions, not concrete implementations |
| 2 | **Debuggability** | Easy to trace dependencies and understand data flow |
| 3 | **Testability** | Simple to write unit tests with minimal setup |
| 4 | **Dependencies Visible in Signature** | Function signatures should reveal all required dependencies |
| 5 | **Flexibility** | Easy to swap implementations without changing consumers |
| 6 | **Extensibility** | Simple to add new functionality without modifying existing code |
| 7 | **Maintainability** | Code is easy to understand and modify over time |
| 8 | **Modularity** | Components are self-contained and independent |
| 9 | **Reusability** | Components can be reused across different contexts |

> **Why "Dependencies Visible in Signature" is Critical:**
>
> ```go
> // ✅ GOOD: Self-documenting - you know exactly what this needs
> func NewOrderService(
>     repo OrderRepository,
>     userSvc UserService,
>     logger Logger,
>     notifier Notifier,
> ) *OrderService
>
> // ❌ HIDDEN: What does this need? Must read the implementation
> func OrderServiceProvider(c *Container) (*OrderService, error)
> ```
>
> When dependencies are visible in signatures:
> - **Code reviews** catch dependency changes immediately
> - **Refactoring** is safer - compiler enforces correctness
> - **IDE support** works fully ("Find Usages", "Go to Definition")
> - **Documentation** is automatic - the signature IS the contract
> - **Onboarding** is faster - new developers understand the code instantly

**Non-Goal:** Minimizing writing effort was explicitly not a priority.

---

## DI Approaches Analyzed

### 1. Context-Based DI (`context.Context`)

Using Go's context to pass dependencies through the call stack.

```go
// Anti-pattern: Using context for DI
func GetUser(ctx context.Context, id string) (*User, error) {
    repo := ctx.Value("userRepo").(UserRepository)  // Type assertion required
    return repo.GetByID(ctx, id)
}
```

### 2. Constructor Injection (Manual)

Explicitly passing dependencies through constructors.

```go
type UserService struct {
    repo   UserRepository
    logger Logger
}

func NewUserService(repo UserRepository, logger Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}
```

### 3. Google Wire (Compile-Time DI)

Code generation-based dependency injection.

```go
// +build wireinject

func InitializeApp() *App {
    wire.Build(NewConfig, NewDatabase, NewUserService, NewApp)
    return &App{}
}
```

### 4. Uber Dig (Runtime DI Container)

Reflection-based dependency resolution.

```go
container := dig.New()
container.Provide(NewConfig)
container.Provide(NewDatabase)
container.Provide(NewUserService)
container.Invoke(func(svc *UserService) { /* use svc */ })
```

### 5. Uber Fx (Application Framework)

Full application framework built on top of Dig.

```go
fx.New(
    fx.Provide(NewConfig),
    fx.Provide(NewDatabase),
    fx.Provide(NewUserService),
    fx.Invoke(StartServer),
).Run()
```

### 6. Service Locator Pattern

Global registry for service resolution.

```go
var services = make(map[string]interface{})
func Register(name string, svc interface{}) { services[name] = svc }
func Get[T any](name string) T { return services[name].(T) }
```

### 7. Closure Injection

Passing dependencies via closures.

```go
func MakeUserHandler(repo UserRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // repo is captured in closure
    }
}
```

### 8. Config Struct / God Object

Bundling all dependencies into a single struct.

```go
type Deps struct {
    DB     *sql.DB
    Cache  Cache
    Logger Logger
}
```

### 9. Astra (Lightweight Generic Container)

A lightweight DI container using Go generics, inspired by Laravel's container.

```go
container := astra.NewContainer(10)

// Register with provider functions
astra.Register[*Config](container, ConfigProvider)
astra.Register[*UserService](container, UserServiceProvider)

// Deferred registration (lazy loading)
astra.RegisterDeferred[*redis.Client](container, RedisProvider)

// Bootstrap resolves all non-deferred services
container.Bootstrap()

// Resolve on demand
svc := astra.Make[*UserService](container)
```

Provider functions are explicit and readable:

```go
func UserServiceProvider(c *astra.Container) (*UserService, error) {
    return &UserService{
        repo:   astra.Make[UserRepository](c),
        logger: astra.Make[Logger](c),
    }, nil
}
```

---

## Detailed Comparison

### Context-Based DI

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ❌ Poor - Dependencies are hidden |
| Debuggability | ❌ Poor - Can't see what's in context |
| Testability | ❌ Poor - Must construct context with correct values |
| Type Safety | ❌ Poor - Runtime type assertions required |
| Appropriate Use | Request-scoped values (trace IDs, deadlines, cancellation) |

### Constructor Injection (Manual)

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ✅ Excellent - Depends on interfaces |
| Debuggability | ✅ Excellent - All dependencies visible in signature |
| Testability | ✅ Excellent - Just pass mocks to constructor |
| Type Safety | ✅ Excellent - Compile-time checked |
| Writing Effort | ⚠️ Medium-High for large applications |

### Google Wire

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ✅ Excellent - Uses interfaces |
| Debuggability | ✅ Excellent - Generated code is readable |
| Testability | ✅ Excellent - Same as manual injection |
| Type Safety | ✅ Excellent - Compile-time errors |
| Writing Effort | ✅ Low - Auto-generates wiring code |

### Uber Dig

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ✅ Good - Uses interfaces |
| Debuggability | ⚠️ Medium - Resolution is implicit |
| Testability | ⚠️ Medium - Requires container setup |
| Type Safety | ⚠️ Medium - Runtime errors for missing deps |
| Writing Effort | ✅ Low - Automatic resolution |

### Uber Fx

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ✅ Good - Uses interfaces |
| Debuggability | ⚠️ Medium - More abstraction layers |
| Testability | ⚠️ Medium - Requires fxtest helpers |
| Lifecycle Mgmt | ✅ Excellent - Built-in start/stop hooks |
| Writing Effort | ✅ Low - Declarative configuration |

### Service Locator

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ❌ Poor - Global state |
| Debuggability | ❌ Poor - Hidden dependencies |
| Testability | ❌ Poor - Global state contamination |
| Type Safety | ❌ Poor - Runtime lookups |
| **Recommendation** | ❌ **Avoid this pattern** |

### Astra (Lightweight Generic Container)

| Aspect | Assessment |
|--------|------------|
| Loose Coupling | ✅ Excellent - Uses interfaces |
| Debuggability | ✅ Good - Provider functions are explicit and readable |
| Testability | ✅ Good - Register mocks in test container |
| Type Safety | ✅ Good - Uses Go generics |
| Runtime Safety | ⚠️ Medium - Missing registrations panic at runtime |
| Writing Effort | ✅ Low - Minimal boilerplate |

---

## Why Not Context-Based DI

Context in Go was designed for:
- Request cancellation
- Deadlines and timeouts
- Request-scoped values (trace IDs, auth tokens)

**Problems with using context for DI:**

1. **Loss of type safety**
   ```go
   // Must use type assertions - runtime panics possible
   repo := ctx.Value("repo").(UserRepository)
   ```

2. **Hidden dependencies**
   ```go
   // Function signature doesn't reveal what it needs
   func ProcessOrder(ctx context.Context, orderID string) error
   // What does this actually require? Unknown without reading code.
   ```

3. **Testing complexity**
   ```go
   // Must construct context correctly
   ctx := context.WithValue(ctx, "repo", mockRepo)
   ctx = context.WithValue(ctx, "logger", mockLogger)
   ctx = context.WithValue(ctx, "cache", mockCache)
   // Easy to forget one, leading to nil pointer panics
   ```

4. **No compile-time verification**
   - Missing dependencies discovered at runtime
   - Typos in key strings cause silent failures

---

## Runtime DI Containers: Dig, Fx, and Astra

Runtime DI containers (Dig, Fx, Astra) are legitimate approaches with their own strengths. This section provides a balanced analysis.

### Overview of Runtime Containers

| Container | Author | Approach | Key Feature |
|-----------|--------|----------|-------------|
| **Dig** | Uber | Reflection-based | Mature, battle-tested |
| **Fx** | Uber | Built on Dig | Full lifecycle management |
| **Astra** | Mohamed Said | Go generics | Lightweight, Laravel-inspired |

### Astra Deep Dive

Astra is a lightweight container created by Mohamed Said (Laravel core team). It uses Go generics instead of pure reflection, providing better type safety than Dig.

**Core API (4 functions):**

```go
// Create container
container := astra.NewContainer(10)

// Register services
astra.Register[*Config](container, ConfigProvider)
astra.RegisterDeferred[*Cache](container, CacheProvider)  // Lazy

// Bootstrap (resolves non-deferred services)
container.Bootstrap()

// Resolve on demand
config := astra.Make[*Config](container)
```

**Provider Functions:**

```go
func ConfigProvider(c *astra.Container) (*Config, error) {
    return &Config{
        RedisAddress: os.Getenv("REDIS_ADDRESS"),
    }, nil
}

func UserServiceProvider(c *astra.Container) (*UserService, error) {
    return &UserService{
        repo:   astra.Make[UserRepository](c),
        logger: astra.Make[Logger](c),
    }, nil
}
```

**Interface Binding:**

```go
func CacheProvider(c *astra.Container) (Cache, error) {
    if os.Getenv("CACHE_DRIVER") == "redis" {
        return astra.Make[*RedisCache](c), nil
    }
    return astra.Make[*MemoryCache](c), nil
}

// Usage - returns appropriate implementation
cache := astra.Make[Cache](container)
```

### Debuggability of Runtime Containers

**Common misconception:** Runtime containers are hard to debug.

**Reality:** Provider functions are explicit and debuggable:

```go
func UserServiceProvider(c *astra.Container) (*UserService, error) {
    // You CAN:
    // - Set breakpoints here
    // - Step through resolution
    // - See exactly what dependencies are requested
    // - Inspect values at runtime
    
    repo := astra.Make[UserRepository](c)   // Clear dependency
    logger := astra.Make[Logger](c)         // Clear dependency
    
    return &UserService{repo: repo, logger: logger}, nil
}
```

**What you can do:**
- Set breakpoints in provider functions
- Step through dependency resolution
- Inspect resolved instances
- Trace the dependency graph by reading provider code

**Limitation compared to manual DI:**
- Can't "go to definition" from `astra.Make[T]` to see the concrete type
- Dependencies not visible in function signatures (must read provider body)

### Testability of Runtime Containers

**Testing with Astra:**

```go
func TestUserService(t *testing.T) {
    // Create test container
    container := astra.NewContainer(10)
    
    // Register mocks
    astra.Register[UserRepository](container, func(c *astra.Container) (UserRepository, error) {
        return &MockUserRepository{
            GetByIDFunc: func(ctx context.Context, id string) (*User, error) {
                return &User{ID: id, Name: "Test"}, nil
            },
        }, nil
    })
    
    astra.Register[Logger](container, func(c *astra.Container) (Logger, error) {
        return &NullLogger{}, nil
    })
    
    astra.Register[*UserService](container, UserServiceProvider)
    container.Bootstrap()
    
    // Test
    svc := astra.Make[*UserService](container)
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", user.Name)
}
```

**Testing with Manual DI:**

```go
func TestUserService(t *testing.T) {
    // Direct instantiation
    svc := NewUserService(&MockUserRepository{}, &NullLogger{})
    
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", user.Name)
}
```

**Comparison:**
- Astra: ~15 lines setup
- Manual: ~3 lines setup

Both are testable. Manual DI is more concise for unit tests.

### The Real Tradeoffs

| Aspect | Runtime Containers (Dig/Astra) | Manual Constructor Injection |
|--------|-------------------------------|------------------------------|
| **Dependencies visible in signature** | ❌ No (must read provider body) | ✅ Yes (self-documenting) |
| **Missing dependency errors** | Runtime panic | Compile-time error |
| **IDE "Find Usages"** | Limited | Full support |
| **Wiring boilerplate** | Low | Medium-High |
| **Debugging** | Good (providers are readable) | Excellent (direct calls) |
| **Testing** | Good (requires container setup) | Excellent (direct instantiation) |
| **Learning curve** | Low-Medium | None |

#### Why "Dependencies Visible in Signature" Matters

This is arguably the **most important tradeoff** to consider:

**Manual DI - Self-Documenting Code:**
```go
// Just by reading this signature, you know EXACTLY what UserService needs
func NewUserService(
    repo UserRepository,
    logger Logger,
    cache Cache,
    metrics MetricsCollector,
) *UserService

// Adding a dependency? The signature changes - visible in code review
// Removing a dependency? The signature changes - compiler catches unused imports
```

**Container-Based DI - Hidden Dependencies:**
```go
// What does this need? No idea without reading the implementation
func UserServiceProvider(c *astra.Container) (*UserService, error) {
    return &UserService{
        repo:    astra.Make[UserRepository](c),    // Hidden here
        logger:  astra.Make[Logger](c),            // Hidden here
        cache:   astra.Make[Cache](c),             // Hidden here
        metrics: astra.Make[MetricsCollector](c),  // Hidden here
    }, nil
}

// Adding a dependency? Just add another Make[] call - easy to miss in review
// Removing a dependency? No compile error if you forget to clean up registration
```

**Practical Impact:**

| Scenario | Manual DI | Container DI |
|----------|-----------|--------------|
| Code review: "What changed?" | Signature diff shows new dependency | Must read full provider body |
| Onboarding: "What does this service need?" | Read function signature | Read provider implementation |
| Refactoring: "What uses Logger?" | IDE "Find Usages" works perfectly | IDE won't find `astra.Make[Logger]` calls |
| Documentation | Signature IS the documentation | Must maintain separate docs |

**This is why Manual DI is still recommended as the primary approach** - the self-documenting nature of explicit constructor parameters is a significant advantage for maintainability.

### When Runtime Containers Excel

1. **Large applications** with many services where manual wiring becomes tedious
2. **Dynamic resolution** based on configuration (interface binding)
3. **Deferred/lazy loading** of expensive resources
4. **Teams familiar** with DI containers from other languages (Spring, Laravel)
5. **Fx specifically:** When you need lifecycle management (start/stop hooks)

### When Manual DI Excels

1. **Compile-time safety** is critical
2. **Small to medium** applications
3. **Maximum IDE support** needed
4. **Team prefers** explicit over implicit
5. **Dependencies must be visible** in function signatures

### Astra vs Dig vs Fx

| Feature | Astra | Dig | Fx |
|---------|:-----:|:---:|:--:|
| Uses Go generics | ✅ | ❌ | ❌ |
| Reflection-based | Minimal | Heavy | Heavy |
| Type safety | ✅ Better | ⚠️ Good | ⚠️ Good |
| Lifecycle management | ❌ | ❌ | ✅ |
| Maturity | New | Mature | Mature |
| Community/Support | Small | Large | Large |
| API simplicity | ✅ 4 functions | ⚠️ More complex | ⚠️ More complex |
| Laravel-style | ✅ | ❌ | ❌ |

---

## The Recommended Approach

### Core Principles

1. **Define interfaces where they're consumed, not where they're implemented**
2. **Use small, focused interfaces (Interface Segregation Principle)**
3. **Wire everything explicitly in main (Composition Root)**
4. **Use decorators/wrappers for cross-cutting concerns**

### Project Structure

```
myapp/
├── cmd/
│   └── server/
│       └── main.go              # Composition root - ALL wiring here
│
├── internal/
│   ├── user/
│   │   ├── service.go           # Defines its OWN interfaces
│   │   ├── service_test.go      # Tests with simple mocks
│   │   └── model.go
│   │
│   ├── order/
│   │   ├── service.go           # Defines its OWN interfaces
│   │   ├── service_test.go
│   │   └── model.go
│   │
│   ├── postgres/                # Concrete implementations
│   │   ├── user_repo.go         # Implements user.UserRepository
│   │   └── order_repo.go        # Implements order.OrderRepository
│   │
│   ├── redis/
│   │   └── cache.go
│   │
│   └── api/
│       ├── handler.go
│       └── router.go
│
└── pkg/                         # Shared utilities
    └── log/
        └── logger.go
```

**Key Insight:** The `user` package has NO knowledge of `postgres`. It only knows about its own `UserRepository` interface.

---

## Implementation Guide

### Step 1: Define Interfaces in Consumer Package

```go
// internal/user/service.go
package user

import "context"

// Define exactly what this package needs - nothing more
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

type Logger interface {
    Info(msg string, fields ...any)
    Error(msg string, fields ...any)
}

type UserService struct {
    repo   UserRepository
    logger Logger
}

func New(repo UserRepository, logger Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    s.logger.Info("fetching user", "id", id)
    
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to fetch user", "id", id, "error", err)
        return nil, err
    }
    
    return user, nil
}
```

### Step 2: Implement Interfaces in Separate Packages

```go
// internal/postgres/user_repo.go
package postgres

import (
    "context"
    "database/sql"
    
    "myapp/internal/user"
)

// Verify interface compliance at compile time
var _ user.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    var u user.User
    err := r.db.QueryRowContext(ctx, 
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email)
    return &u, err
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
        u.ID, u.Name, u.Email,
    )
    return err
}
```

### Step 3: Wire Everything in Main

```go
// cmd/server/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"
    
    "myapp/internal/api"
    "myapp/internal/order"
    "myapp/internal/postgres"
    "myapp/internal/user"
    "myapp/pkg/config"
    zaplog "myapp/pkg/log"
    
    _ "github.com/lib/pq"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize infrastructure
    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    logger := zaplog.New(cfg.LogLevel)
    
    // Build dependency graph (bottom-up)
    // Repositories
    userRepo := postgres.NewUserRepository(db)
    orderRepo := postgres.NewOrderRepository(db)
    
    // Services
    userService := user.New(userRepo, logger)
    orderService := order.New(orderRepo, userService, logger)
    
    // HTTP Handlers
    userHandler := api.NewUserHandler(userService)
    orderHandler := api.NewOrderHandler(orderService)
    
    // Router
    router := api.NewRouter(userHandler, orderHandler)
    
    // Start server
    logger.Info("starting server", "port", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
        logger.Error("server failed", "error", err)
    }
}
```

---

## Testing Strategy

### Unit Testing with Simple Mocks

```go
// internal/user/service_test.go
package user

import (
    "context"
    "errors"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

// Function-based mock for maximum flexibility
type mockUserRepository struct {
    getByIDFunc func(ctx context.Context, id string) (*User, error)
    saveFunc    func(ctx context.Context, user *User) error
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    if m.getByIDFunc != nil {
        return m.getByIDFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}

func (m *mockUserRepository) Save(ctx context.Context, user *User) error {
    if m.saveFunc != nil {
        return m.saveFunc(ctx, user)
    }
    return errors.New("not implemented")
}

// Null logger for tests
type nullLogger struct{}

func (nullLogger) Info(msg string, fields ...any)  {}
func (nullLogger) Error(msg string, fields ...any) {}

// Tests
func TestGetUser_Success(t *testing.T) {
    // Arrange
    expectedUser := &User{ID: "123", Name: "Alice", Email: "alice@example.com"}
    
    repo := &mockUserRepository{
        getByIDFunc: func(ctx context.Context, id string) (*User, error) {
            assert.Equal(t, "123", id)
            return expectedUser, nil
        },
    }
    
    svc := New(repo, &nullLogger{})
    
    // Act
    user, err := svc.GetUser(context.Background(), "123")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
}

func TestGetUser_NotFound(t *testing.T) {
    // Arrange
    repo := &mockUserRepository{
        getByIDFunc: func(ctx context.Context, id string) (*User, error) {
            return nil, errors.New("user not found")
        },
    }
    
    svc := New(repo, &nullLogger{})
    
    // Act
    user, err := svc.GetUser(context.Background(), "999")
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
}

func TestGetUser_RepositoryCalled(t *testing.T) {
    // Arrange
    called := false
    repo := &mockUserRepository{
        getByIDFunc: func(ctx context.Context, id string) (*User, error) {
            called = true
            return &User{ID: id}, nil
        },
    }
    
    svc := New(repo, &nullLogger{})
    
    // Act
    _, _ = svc.GetUser(context.Background(), "123")
    
    // Assert
    assert.True(t, called, "repository should have been called")
}
```

### Integration Testing

```go
// internal/postgres/user_repo_test.go
package postgres

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "myapp/internal/user"
    "myapp/pkg/testutil"
)

func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test database
    db := testutil.SetupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    ctx := context.Background()
    
    // Test Save
    testUser := &user.User{
        ID:    "test-123",
        Name:  "Test User",
        Email: "test@example.com",
    }
    
    err := repo.Save(ctx, testUser)
    require.NoError(t, err)
    
    // Test GetByID
    retrieved, err := repo.GetByID(ctx, "test-123")
    require.NoError(t, err)
    assert.Equal(t, testUser.Name, retrieved.Name)
    assert.Equal(t, testUser.Email, retrieved.Email)
}
```

---

## Advanced Patterns

### Decorator Pattern for Cross-Cutting Concerns

```go
// Logging decorator
type loggingUserRepository struct {
    repo   user.UserRepository
    logger user.Logger
}

func NewLoggingUserRepository(repo user.UserRepository, logger user.Logger) user.UserRepository {
    return &loggingUserRepository{repo: repo, logger: logger}
}

func (r *loggingUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    r.logger.Info("repository.GetByID called", "id", id)
    
    u, err := r.repo.GetByID(ctx, id)
    
    if err != nil {
        r.logger.Error("repository.GetByID failed", "id", id, "error", err)
    }
    return u, err
}

func (r *loggingUserRepository) Save(ctx context.Context, u *user.User) error {
    r.logger.Info("repository.Save called", "userID", u.ID)
    return r.repo.Save(ctx, u)
}
```

### Caching Decorator

```go
type cachedUserRepository struct {
    repo  user.UserRepository
    cache Cache
    ttl   time.Duration
}

func NewCachedUserRepository(repo user.UserRepository, cache Cache, ttl time.Duration) user.UserRepository {
    return &cachedUserRepository{repo: repo, cache: cache, ttl: ttl}
}

func (r *cachedUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    // Check cache first
    cacheKey := "user:" + id
    if cached, ok := r.cache.Get(cacheKey); ok {
        return cached.(*user.User), nil
    }
    
    // Fetch from repository
    u, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    r.cache.Set(cacheKey, u, r.ttl)
    return u, nil
}

func (r *cachedUserRepository) Save(ctx context.Context, u *user.User) error {
    // Invalidate cache on save
    r.cache.Delete("user:" + u.ID)
    return r.repo.Save(ctx, u)
}
```

### Composing Decorators in Main

```go
// cmd/server/main.go
func main() {
    // ... initialization ...
    
    // Build repository with decorators (order matters!)
    baseRepo := postgres.NewUserRepository(db)
    loggingRepo := middleware.NewLoggingUserRepository(baseRepo, logger)
    cachedRepo := middleware.NewCachedUserRepository(loggingRepo, cache, 5*time.Minute)
    
    // Request flow: cachedRepo -> loggingRepo -> baseRepo
    // 1. Check cache
    // 2. Log the call (if cache miss)
    // 3. Query database
    
    userService := user.New(cachedRepo, logger)
    // ...
}
```

### Functional Options for Optional Dependencies

```go
type UserServiceOption func(*UserService)

func WithCache(cache Cache) UserServiceOption {
    return func(s *UserService) {
        s.cache = cache
    }
}

func WithMetrics(metrics Metrics) UserServiceOption {
    return func(s *UserService) {
        s.metrics = metrics
    }
}

func New(repo UserRepository, logger Logger, opts ...UserServiceOption) *UserService {
    s := &UserService{
        repo:   repo,
        logger: logger,
    }
    
    for _, opt := range opts {
        opt(s)
    }
    
    return s
}

// Usage
svc := user.New(repo, logger, 
    user.WithCache(redisCache),
    user.WithMetrics(prometheus),
)
```

---

## Comparison Tables

### Overall Comparison Matrix

| Approach | Loose Coupling | Debuggability | Testability | Type Safety | Runtime Safety | Writing Effort |
|----------|:--------------:|:-------------:|:-----------:|:-----------:|:--------------:|:--------------:|
| Context-Based DI | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Constructor Injection | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Google Wire | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Uber Dig | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ✅ |
| Uber Fx | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ✅ |
| Astra | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ |
| Service Locator | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Closure Injection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Config Struct | ⚠️ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ |

Legend: ✅ Excellent/Good | ⚠️ Acceptable | ❌ Poor

### Quality Attributes Comparison

| Approach | Flexibility | Extensibility | Maintainability | Modularity | Reusability |
|----------|:-----------:|:-------------:|:---------------:|:----------:|:-----------:|
| Context-Based DI | ❌ | ❌ | ❌ | ❌ | ❌ |
| Constructor Injection | ✅ | ✅ | ✅ | ✅ | ✅ |
| Google Wire | ✅ | ✅ | ✅ | ✅ | ✅ |
| Uber Dig | ✅ | ✅ | ✅ | ✅ | ✅ |
| Uber Fx | ✅ | ✅ | ✅ | ✅ | ✅ |
| Astra | ✅ | ✅ | ✅ | ✅ | ✅ |
| Service Locator | ⚠️ | ❌ | ❌ | ❌ | ❌ |
| Closure Injection | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Config Struct | ⚠️ | ⚠️ | ⚠️ | ❌ | ⚠️ |

### Detailed Feature Comparison

| Feature | Manual DI | Wire | Dig | Fx | Astra |
|---------|:---------:|:----:|:---:|:--:|:-----:|
| **Dependencies visible in signature** | ✅ | ✅ | ❌ | ❌ | ❌ |
| Compile-time error detection | ✅ | ✅ | ❌ | ❌ | ❌ |
| IDE "Go to Definition" | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ |
| IDE "Find Usages" | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Debuggable (breakpoints) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Testable | ✅ | ✅ | ✅ | ✅ | ✅ |
| Minimal test setup | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Uses Go generics | N/A | N/A | ❌ | ❌ | ✅ |
| Lifecycle management | ❌ | ❌ | ❌ | ✅ | ❌ |
| Lazy/deferred loading | Manual | ❌ | ✅ | ✅ | ✅ |
| Interface binding | Manual | ✅ | ✅ | ✅ | ✅ |
| Low boilerplate | ❌ | ✅ | ✅ | ✅ | ✅ |
| No external dependencies | ✅ | ❌ | ❌ | ❌ | ❌ |
| Learning curve | None | Low | Medium | Medium | Low |

> **⚠️ Critical Tradeoff: Dependencies Visible in Signature**
>
> This is the most significant difference between manual DI and container-based approaches.
>
> **Manual DI / Wire:**
> ```go
> // Dependencies are EXPLICIT in the function signature
> func NewUserService(repo UserRepository, logger Logger, cache Cache) *UserService
> // You know exactly what this needs just by looking at the signature
> ```
>
> **Dig / Fx / Astra:**
> ```go
> // Dependencies are HIDDEN inside the provider body
> func UserServiceProvider(c *astra.Container) (*UserService, error)
> // What does this need? You must read the function body to find out
> ```
>
> **Why this matters:**
> - Code review: Reviewers can't see dependency changes from signatures alone
> - Refactoring: Harder to track what depends on what
> - Documentation: Signatures don't self-document requirements
> - IDE support: "Find all usages of Logger" won't find container-based consumers

### When to Use Each Approach

| Approach | Best For | Avoid When |
|----------|----------|------------|
| Constructor Injection | Most applications, maximum control | Never - always appropriate |
| Google Wire | Large apps, compile-time safety needed | Small apps (overkill) |
| Uber Dig | Medium-large apps, familiar with DI containers | Compile-time safety critical |
| Uber Fx | Apps needing lifecycle management | Simple apps without start/stop needs |
| Astra | Laravel developers, simple container needs | Need lifecycle management |
| Closure Injection | HTTP handlers, callbacks | Many dependencies needed |
| Context | Request-scoped data only | Dependency injection |
| Service Locator | Legacy code migration | New development |

### Final Recommendation by Project Size and Team

| Scenario | Recommended Approach |
|----------|---------------------|
| Small project (< 10 services) | Manual Constructor Injection |
| Medium project (10-50 services) | Manual DI or Astra |
| Large project (50+ services) | Manual DI + Wire, or Dig/Fx |
| Team from Laravel/Spring background | Astra or Dig |
| Microservices | Manual DI per service |
| Need lifecycle hooks | Fx |
| Maximum compile-time safety | Manual DI or Wire |

---

## Conclusion

### Primary Recommendation: Interfaces + Constructor Injection

For most Go projects, **Interfaces + Constructor Injection + Clean Package Structure** remains the recommended approach because it provides:

1. **Maximum clarity** - All dependencies visible in function signatures
2. **Compile-time safety** - Missing dependencies caught before runtime
3. **Trivial testing** - No framework setup, just pass mocks
4. **Easy debugging** - Step through code, set breakpoints, trace flow
5. **True loose coupling** - Packages depend only on their own interfaces
6. **Flexible composition** - Decorators and wrappers for cross-cutting concerns
7. **No magic** - What you see is what you get

### Valid Alternatives

However, runtime DI containers (Dig, Fx, Astra) are **legitimate choices** when:

- Your team is familiar with DI containers from other languages
- You have a large application where manual wiring is tedious
- You need dynamic interface binding based on configuration
- You need lifecycle management (Fx specifically)
- You prefer the Laravel-style conventions (Astra)

**Key insight:** Runtime containers like Astra and Dig ARE debuggable and testable. The provider functions are explicit and can be stepped through with a debugger. The main tradeoff is runtime vs compile-time error detection, not debuggability.

### Decision Framework

| If you prioritize... | Choose... |
|---------------------|-----------|
| Compile-time safety | Manual DI or Wire |
| Minimal boilerplate | Astra, Dig, or Fx |
| Maximum IDE support | Manual DI or Wire |
| Lifecycle management | Fx |
| Simplicity + type safety | Astra |
| Team familiarity (Laravel/Spring) | Astra or Dig |
| No external dependencies | Manual DI |

### Final Thought

There is no single "best" approach. The right choice depends on your team's background, project size, and specific requirements. All approaches discussed (except context-based DI and service locator) can produce maintainable, testable, and debuggable code when used correctly.

---

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Google Wire](https://github.com/google/wire)
- [Uber Dig](https://github.com/uber-go/dig)
- [Uber Fx](https://github.com/uber-go/fx)
- [Astra - Building a DI Container for Go](https://themsaid.com/building-a-dependency-injection-container-for-go)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
