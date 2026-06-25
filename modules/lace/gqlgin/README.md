# gqlgin

A Gin-integrated GraphQL server setup package that provides secure endpoint registration with configurable playground access and introspection control.

## Purpose

`gqlgin` simplifies setting up GraphQL endpoints with Gin by handling:

- **Query endpoint** (`/query`) - Main GraphQL endpoint for API requests
- **Playground endpoint** (`/gql`) - GraphQL Playground UI for development/debugging
- **Introspection endpoint** (`/query_playground`) - Separate endpoint for schema introspection

The package separates the playground/introspection from the main query endpoint to:
1. Optimize performance on production queries (no playground middleware overhead)
2. Protect schema introspection behind a secret key
3. Allow different authentication strategies for API vs playground access

## Installation

The package is part of the `lace` module workspace.

```go
import "lace/gqlgin"
```

## Quick Start

```go
package main

import (
    "lace/gqlgin"
    "lace/gqlgenfn"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    rg := r.Group("/api")

    // Create your gqlgen server
    gserver := handler.NewDefaultServer(generated.NewExecutableSchema(config))

    // Setup GraphQL endpoints
    gqlgin.New(
        gqlgin.WithRouterGroup(rg),
        gqlgin.WithGqlServer(gserver),
        gqlgin.WithPlaygroundKey("your-secret-key"),
        gqlgin.WithRouteGroupPrefix(rg.BasePath()),
        gqlgin.WithMiddleware(
            gqlgenfn.GinContextToContextMiddleware(),
        ),
    )

    r.Run(":8080")
}
```

This registers:
- `POST /api/query` - Main GraphQL endpoint
- `POST /api/query_playground` - Introspection endpoint (requires `?pkey=your-secret-key`)
- `GET /api/gql` - Playground UI (requires `?pkey=your-secret-key`)

## Configuration Options

### Required Options

| Option | Description |
|--------|-------------|
| `WithRouterGroup(rg)` | Gin router group to register routes on |
| `WithGqlServer(server)` | gqlgen handler.Server instance |

### Optional Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithPlaygroundKey(key)` | `"pkey"` | Secret key required to access playground |
| `WithRouteGroupPrefix(prefix)` | `""` | Prefix for playground handler URLs |
| `WithMiddleware(handlers...)` | `nil` | Middleware chain for all endpoints |
| `WithQueryRoute(route)` | `"/query"` | Path for main query endpoint |
| `WithQueryPlaygroundRoute(route)` | `"/query_playground"` | Path for introspection endpoint |
| `WithPlaygroundRoute(route)` | `"/gql"` | Path for playground UI |
| `WithDisablePlayground(bool)` | `false` | Disable playground routes entirely |
| `WithEnableDefaultIntrospection(bool)` | `false` | Enable introspection on main query route |

## Usage Examples

### Basic Setup

```go
gqlgin.New(
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithPlaygroundKey(config.Graphql.Key),
    gqlgin.WithRouteGroupPrefix(rg.BasePath()),
)
```

### Custom Route Paths

```go
gqlgin.New(
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithPlaygroundKey("secret"),
    gqlgin.WithQueryRoute("/graphql"),           // POST /api/graphql
    gqlgin.WithPlaygroundRoute("/playground"),   // GET /api/playground
    gqlgin.WithQueryPlaygroundRoute("/introspect"), // POST /api/introspect
)
```

### With Authentication Middleware

```go
gqlgin.New(
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithPlaygroundKey(config.Graphql.Key),
    gqlgin.WithRouteGroupPrefix(rg.BasePath()),
    gqlgin.WithMiddleware(
        gqlgenfn.GinContextToContextMiddleware(),
        authMiddleware(),      // Your auth middleware
        userContextMiddleware(), // Add user to context
    ),
)
```

### Production API (No Playground)

For external/partner APIs where you don't want to expose the playground:

```go
gqlgin.New(
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithDisablePlayground(true), // No /gql or /query_playground routes
    gqlgin.WithMiddleware(
        gqlgenfn.GinContextToContextMiddleware(),
        apiKeyMiddleware(),
    ),
)
```

### API with Public Introspection

For APIs where consumers need schema introspection (protected by API key, not playground key):

```go
gqlgin.New(
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithDisablePlayground(true),           // Hide playground
    gqlgin.WithEnableDefaultIntrospection(true),  // Allow introspection on /query
    gqlgin.WithMiddleware(
        gqlgenfn.GinContextToContextMiddleware(),
        apiKeyMiddleware(), // Introspection protected by API key
    ),
)
```

## Accessing the Playground

The playground is protected by a secret key passed as a query parameter:

```
http://localhost:8080/api/gql?pkey=your-secret-key
```

The playground UI will automatically use the `/query_playground` endpoint for requests, which also requires the key.

## Architecture

### Route Structure

```
/api
├── /query              # Main GraphQL endpoint (POST)
│   └── Middleware: GinContextToContext → [UserMiddleware] → ServerHandler
│
├── /query_playground   # Introspection endpoint (POST) - if playground enabled
│   └── Middleware: GinContextToContext → PlaygroundAccess → Introspection → [UserMiddleware] → ServerHandler
│
└── /gql                # Playground UI (GET) - if playground enabled
    └── Middleware: PlaygroundAccess → PlaygroundHandler
```

### Introspection Control

By default, introspection is **disabled** on the `/query` endpoint to prevent schema leakage in production. Introspection is only enabled:

1. On `/query_playground` when the correct `pkey` is provided
2. On `/query` if `WithEnableDefaultIntrospection(true)` is set

### Performance Optimization

The `/query` endpoint doesn't include playground middleware, ensuring:
- No key validation overhead on every request
- No introspection middleware processing
- Minimal latency for production API calls

## Default Constants

```go
const (
    DefaultQueryRoute           = "/query"
    DefaultQueryPlaygroundRoute = "/query_playground"
    DefaultPlaygroundRoute      = "/gql"
    DefaultPlaygroundKey        = "pkey"
)
```

## File Structure

```
lace/gqlgin/
├── gqlgin.go    # Core implementation (Gql struct, New function, handlers)
├── options.go   # Configuration (Config, Option type, With* functions)
└── README.md    # This documentation
```

## Related Packages

- `lace/gqlgenfn` - GraphQL helper functions and middleware
  - `GinContextToContextMiddleware()` - Adds Gin context to GraphQL context
  - `PlaygroundAccessMiddleware(key)` - Validates playground access key
  - `IsPlaygroundGqlCtx(ctx)` - Checks if request is from playground

## Common Patterns

### Multiple GraphQL APIs

You can register multiple GraphQL APIs with different configurations:

```go
// Public API
publicRg := r.Group("/api/public")
gqlgin.New(
    gqlgin.WithRouterGroup(publicRg),
    gqlgin.WithGqlServer(publicServer),
    gqlgin.WithDisablePlayground(true),
    gqlgin.WithMiddleware(apiKeyMiddleware()),
)

// Admin API with playground
adminRg := r.Group("/api/admin")
gqlgin.New(
    gqlgin.WithRouterGroup(adminRg),
    gqlgin.WithGqlServer(adminServer),
    gqlgin.WithPlaygroundKey(config.AdminPlaygroundKey),
    gqlgin.WithMiddleware(adminAuthMiddleware()),
)
```

### Conditional Playground

Enable playground only in development:

```go
opts := []gqlgin.Option{
    gqlgin.WithRouterGroup(rg),
    gqlgin.WithGqlServer(gserver),
    gqlgin.WithMiddleware(gqlgenfn.GinContextToContextMiddleware()),
}

if config.Environment != "production" {
    opts = append(opts,
        gqlgin.WithPlaygroundKey(config.Graphql.Key),
        gqlgin.WithRouteGroupPrefix(rg.BasePath()),
    )
} else {
    opts = append(opts, gqlgin.WithDisablePlayground(true))
}

gqlgin.New(opts...)
```
