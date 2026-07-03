# Golang Code Guidelines

Project-specific Go conventions and best practices. Follow these alongside [Effective Go](https://go.dev/doc/effective_go) and standard style (e.g. `gofmt`).

***

## Project-Specific Rules

### Ent ORM

* **Use field constants, not string literals.** Never hardcode field names in queries; use the generated field constants so renames stay safe and refactors are reliable.

  ```go
  // Bad
  Order(ent.Desc("created_at"))

  // Good
  Order(ent.Desc(user.FieldCreatedAt))
  ```

* **Disable foreign-key constraint creation in the database** where the project requires it (e.g. for migration or deployment constraints). Configure this in the Ent schema or migration flow, not ad hoc in SQL.

### File locks

* **Keep filelock keys lowercase.** Lock names are used as filesystem identifiers; lowercase avoids case-sensitivity issues across platforms.

  ```go
  // Good
  filelock.New("mailerlitesubscribechargegroup")
  ```

***

## Naming Conventions

### Package names: singular, not plural

Package names should be singular. Plural names are a common carryover from other languages and conflict with Go style.

```go
// Bad
package httputils

// Good
package httputil
```

### Prefer underscore-style package paths

Use short, project-specific path segments with underscores rather than generic or easily conflicting names.

```go
// Prefer: specific, underscore-style
import "api/db/hooks"
import "webapp/hook"

// Avoid: generic names that collide easily
import "api/db/hook"  // too generic
```

Choose names like `dbhook` or `saashook` over generic `hooks` where it clarifies ownership and avoids clashes.

### Struct names: relationship order

Name structs so the relationship reads clearly: “A X has many Y” → `X_Y` or “Y belongs to X” → `X_Y`.

```go
// A group has many users (group–user relationship)
type GroupUser struct {}

// A user has many images (user–image relationship)
type UserImage struct {}
```

***

## Functions and Errors

### Return errors as the last value

Functions that can fail must return `error` as the last return value. No Boolean “ok” or error in the middle.

```go
// Good
func demo() (bool, error) {
    // ...
}

// Bad
func demo() (error, bool) {}
func demo() (bool, error, int) {}
```

### Handle errors; avoid silent ignore

Check every `error` return. Do not use `_` to discard errors unless you have a documented reason (e.g. optional cleanup).

```go
// Good
if err != nil {
    return fmt.Errorf("doing X: %w", err)
}

// Bad
result, _ := doSomething()
```

### Prefer sentinel or wrapped errors for semantics

Use `errors.Is` / `errors.As` by defining sentinel errors or custom types and wrapping with `%w` when adding context.

```go
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) { ... }
return fmt.Errorf("load user: %w", err)
```

***

## General Best Practices

### Keep packages focused

Each package should have a clear, narrow responsibility. If a file or package grows into many unrelated concerns, split it.

### Accept interfaces, return structs

Take dependencies as interfaces (defined in the caller or a small, shared package) and return concrete types so the API stays stable and testable.

```go
// Good: accept interface, return concrete
func Process(r io.Reader) (*Result, error) {}

// Caller can pass *os.File, *bytes.Reader, etc.
```

### Avoid global state

Prefer explicit dependencies (constructors, options, context) over package-level globals. This improves testability and makes data flow obvious.

### Use context for cancellation and timeouts

For any call that can block (HTTP, DB, RPC), accept a `context.Context` as the first argument and respect its cancellation and deadline.

```go
func (s *Service) Fetch(ctx context.Context, id string) (*Item, error) {
    // Use ctx in HTTP calls, DB queries, etc.
}
```

### Prefer table-driven tests

Structure tests as a slice of cases and loop over them. This keeps tests readable and easy to extend.

```go
func TestParse(t *testing.T) {
    for _, tc := range []struct {
        name, in string
        want     int
    }{
        {"empty", "", 0},
        {"single", "a", 1},
    } {
        t.Run(tc.name, func(t *testing.T) {
            got := Parse(tc.in)
            if got != tc.want {
                t.Errorf("Parse(%q) = %d; want %d", tc.in, got, tc.want)
            }
        })
    }
}
```

### Document exported names

Exported functions, types, and constants should have a doc comment starting with the symbol name. Unexported code can use brief comments where they add clarity.

```go
// ParseConfig loads and validates the config from path.
func ParseConfig(path string) (*Config, error) {}
```

***

## Summary

| Area              | Rule |
|-------------------|------|
| Ent               | Use `ent.FieldX` constants, not `"field_x"` strings; disable FK creation where required. |
| File locks        | Lowercase lock names. |
| Packages          | Singular names; prefer underscore-style, project-specific paths. |
| Structs           | Name by relationship (e.g. `GroupUser`, `UserImage`). |
| Errors            | Last return value; always handle; prefer `%w` and `errors.Is`/`As`. |
| Design            | Accept interfaces, return structs; avoid globals; use `context`. |
| Tests             | Prefer table-driven tests; document exported APIs. |
