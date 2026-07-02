# Testing Rules

**Read before:** Writing tests for this codebase.

## Checklist

- [ ] Test file named `*_test.go` beside source file
- [ ] Use table-driven tests for multiple cases
- [ ] Mock external dependencies
- [ ] Run with `task test`

## Test Structure

```
package/
├── service.go
└── service_test.go  # Tests go beside source
```

## Do

- Use table-driven tests for multiple cases
- Mock external services (DB, Redis, HTTP)
- Test error cases, not just happy path
- Keep tests hermetic (no network unless Docker running)
- Use `t.Parallel()` for independent tests

## Don't

- Test generated code
- Rely on external services without mocks
- Skip error case testing
- Leave test data in database

## Table-Driven Test Example

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty string", "", true},
        {"spaces", "user @example.com", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

## Testing with Ent (Database)

```go
import "dbent/gen/ent/enttest"

func TestUserCreate(t *testing.T) {
    // In-memory SQLite for tests
    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
    defer client.Close()

    ctx := context.Background()
    user, err := client.User.Create().
        SetEmail("test@example.com").
        Save(ctx)

    if err != nil {
        t.Fatalf("failed creating user: %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("unexpected email: %s", user.Email)
    }
}
```

## Mocking HTTP

```go
import "net/http/httptest"

func TestFetchData(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()

    result, err := FetchData(server.URL)
    // assertions...
}
```

## Running Tests

```bash
task test              # Run all tests
go test ./...          # Alternative
go test -v ./pkg/...   # Verbose, specific path
go test -run TestName  # Run specific test
```
