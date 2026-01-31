# Shutdown Module

A graceful shutdown handler for Go applications that manages OS signal handling and executes cleanup handlers in order.

## Installation

```go
import "lace/shutdown"
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"

    "lace/shutdown"
)

func main() {
    server := &http.Server{Addr: ":8080"}

    // Create shutdown manager
    s := shutdown.New()

    // Add shutdown handler for HTTP server
    s.AddShutdownHandler(func(ctx context.Context) error {
        log.Println("Shutting down HTTP server...")
        return server.Shutdown(ctx)
    })

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Block and wait for shutdown signal (SIGINT, SIGTERM)
    s.Listen()
}
```

## Configuration Options

The module uses the functional options pattern for configuration:

### WithTimeout

Sets the context timeout for shutdown handlers (default: 5 seconds).

```go
s := shutdown.New(
    shutdown.WithTimeout(10 * time.Second),
)
```

### WithGracefulShutdownTimeout

Sets additional wait time after handlers complete to allow pending operations to finish (default: 30 seconds).

```go
s := shutdown.New(
    shutdown.WithGracefulShutdownTimeout(60 * time.Second),
)
```

### WithLogger

Sets a custom zap logger (default: production logger).

```go
logger, _ := zap.NewDevelopment()
s := shutdown.New(
    shutdown.WithLogger(logger),
)
```

### WithShutdownHandler

Adds one or more shutdown handlers at initialization.

```go
s := shutdown.New(
    shutdown.WithShutdownHandler(
        func(ctx context.Context) error {
            // cleanup database
            return db.Close()
        },
        func(ctx context.Context) error {
            // cleanup cache
            return cache.Close()
        },
    ),
)
```

### Custom Messages

Customize log messages for shutdown events:

```go
s := shutdown.New(
    shutdown.WithMsgShuttingDown("Server is shutting down..."),
    shutdown.WithMsgShutdownCompleted("Server shutdown complete"),
    shutdown.WithMsgShutdownFailed("Server shutdown encountered an error"),
)
```

## Adding Handlers

### At Initialization

```go
s := shutdown.New(
    shutdown.WithShutdownHandler(handler1, handler2),
)
```

### After Initialization

```go
s := shutdown.New()
s.AddShutdownHandler(func(ctx context.Context) error {
    // cleanup logic
    return nil
})
```

## Handler Execution

* Handlers are executed **sequentially** in the order they were added
* Execution **stops on first error** - subsequent handlers will not run
* Handlers receive a context with the configured timeout
* Handlers should respect context cancellation for proper timeout handling

### Context-Aware Handler Example

```go
s.AddShutdownHandler(func(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // Timeout exceeded
    case <-longRunningOperation():
        return nil
    }
})
```

## Complete Example

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "time"

    "go.uber.org/zap"
    "lace/shutdown"
)

func main() {
    logger, _ := zap.NewProduction()

    // Initialize resources
    db, _ := sql.Open("postgres", "...")
    server := &http.Server{Addr: ":8080"}

    // Create shutdown manager with custom configuration
    s := shutdown.New(
        shutdown.WithLogger(logger),
        shutdown.WithTimeout(15 * time.Second),
        shutdown.WithGracefulShutdownTimeout(30 * time.Second),
        shutdown.WithMsgShuttingDown("Initiating graceful shutdown..."),
    )

    // Add handlers in order of priority
    // HTTP server should stop accepting new requests first
    s.AddShutdownHandler(func(ctx context.Context) error {
        logger.Info("Stopping HTTP server")
        return server.Shutdown(ctx)
    })

    // Then close database connections
    s.AddShutdownHandler(func(ctx context.Context) error {
        logger.Info("Closing database connection")
        return db.Close()
    })

    // Start server
    go func() {
        logger.Info("Starting HTTP server", zap.String("addr", ":8080"))
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            logger.Fatal("HTTP server error", zap.Error(err))
        }
    }()

    // Block until shutdown signal received
    s.Listen()
}
```

## Signal Handling

The `Listen()` method blocks and waits for:

* `SIGINT` (Ctrl+C)
* `SIGTERM` (kill command)

When a signal is received:

1. Log "Shutting down..." message
2. Create context with configured timeout
3. Execute all handlers sequentially
4. Wait for graceful shutdown timeout
5. Log completion message

## Default Configuration

| Option | Default Value |
|--------|---------------|
| Timeout | 5 seconds |
| GracefulShutdownTimeout | 30 seconds |
| MsgShuttingDown | "Shutting down..." |
| MsgShutdownCompleted | "Shutdown completed" |
| MsgShutdownFailed | "Shutdown failed" |
| Logger | zap.NewProduction() |

## API Reference

### Types

```go
type ShutdownHandler func(ctx context.Context) error

type Option func(*config)

type Shutdown struct {
    // contains filtered or unexported fields
}
```

### Functions

```go
func New(opts ...Option) *Shutdown
func (s *Shutdown) AddShutdownHandler(handler func(ctx context.Context) error)
func (s *Shutdown) Listen()
```

### Options

```go
func WithTimeout(timeout time.Duration) Option
func WithGracefulShutdownTimeout(gracefulShutdownTimeout time.Duration) Option
func WithLogger(logger *zap.Logger) Option
func WithShutdownHandler(handlers ...ShutdownHandler) Option
func WithMsgShuttingDown(msgShuttingDown string) Option
func WithMsgShutdownCompleted(msgShutdownCompleted string) Option
func WithMsgShutdownFailed(msgShutdownFailed string) Option
```

## Snippets

### gin-gonic test

```go
package main

import (
	"context"
	"lace/shutdown"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	srv := &http.Server{
		Addr:    ":9001",
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	s := shutdown.New(
		shutdown.WithShutdownHandler(func(ctx context.Context) error {
			log.Println("Shutting down server...")
			return srv.Shutdown(ctx)
		}),
	)

	s.Listen()
}
```
