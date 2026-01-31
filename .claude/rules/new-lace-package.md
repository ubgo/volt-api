# New Lace Package Rules

**Read before:** Adding a shared utility package to lace/.

## Checklist

- [ ] Check if similar package already exists
- [ ] Create package in `lace/<packagename>/`
- [ ] Keep it generic (no project-specific logic)
- [ ] Add tests
- [ ] Run `task workvendor` after adding

## When to Add to Lace

| Add to Lace | Add to saas/pkg/ |
|-------------|------------------|
| Generic utilities | Project-specific logic |
| Reusable across projects | Uses project config/models |
| Infrastructure wrappers | Business domain services |
| No project dependencies | Depends on dbent, config |

## Existing Lace Packages

Check before creating:

| Package | Purpose |
|---------|---------|
| `gozap` | Structured logging |
| `gotel` | OpenTelemetry tracing |
| `goerr` | Error handling |
| `cache` | Cache interface |
| `cacheredis` | Redis cache implementation |
| `natso` | NATS wrapper |
| `db` | Database connection |
| `cli` | Cobra CLI wrapper |
| `ginserver` | Gin HTTP server |
| `crypt` | AES encryption |
| `shutdown` | Graceful shutdown |
| `httpreq` | HTTP client utilities |
| `filelock` | File locking |
| `filesystem` | File operations |
| `publicid` | ID generation |
| `util` | Misc utilities |

## Package Structure

```
lace/<packagename>/
├── <packagename>.go    # Main implementation
├── <packagename>_test.go  # Tests
├── options.go          # Options pattern (if needed)
└── doc.go              # Package documentation (optional)
```

## Package Template

```go
// lace/myutil/myutil.go
package myutil

import (
    "context"
    "fmt"
)

// Client provides myutil functionality.
type Client struct {
    config Config
}

// Config holds configuration for Client.
type Config struct {
    Option1 string
    Option2 int
}

// New creates a new Client with the given options.
func New(opts ...Option) *Client {
    cfg := Config{
        Option1: "default",
        Option2: 10,
    }

    for _, opt := range opts {
        opt(&cfg)
    }

    return &Client{config: cfg}
}

// DoSomething performs the main operation.
func (c *Client) DoSomething(ctx context.Context, input string) (string, error) {
    if input == "" {
        return "", fmt.Errorf("myutil: input cannot be empty")
    }

    // Implementation
    return input + c.config.Option1, nil
}
```

## Options Pattern

```go
// lace/myutil/options.go
package myutil

// Option configures the Client.
type Option func(*Config)

// WithOption1 sets Option1.
func WithOption1(val string) Option {
    return func(c *Config) {
        c.Option1 = val
    }
}

// WithOption2 sets Option2.
func WithOption2(val int) Option {
    return func(c *Config) {
        c.Option2 = val
    }
}
```

## Error Conventions

```go
// Prefix errors with package name
return fmt.Errorf("myutil: %w", err)
return fmt.Errorf("myutil: operation failed: %w", err)

// Define sentinel errors if needed
var (
    ErrNotFound = errors.New("myutil: not found")
    ErrInvalid  = errors.New("myutil: invalid input")
)
```

## Testing

```go
// lace/myutil/myutil_test.go
package myutil

import (
    "context"
    "testing"
)

func TestDoSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "hellodefault", false},
        {"empty input", "", "", true},
    }

    client := New()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := client.DoSomething(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("DoSomething() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("DoSomething() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## After Creating

```bash
# Sync workspace
task workvendor

# Run tests
go test ./lace/myutil/...

# Use in project
import "lace/myutil"
```

## Interface Pattern (for testability)

```go
// Define interface for mocking
type Store interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
}

// Implementation
type RedisStore struct { ... }

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) { ... }
func (r *RedisStore) Set(ctx context.Context, key, value string) error { ... }
```
