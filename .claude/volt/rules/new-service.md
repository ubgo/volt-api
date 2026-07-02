# New Service Rules

**Read before:** Adding a new business logic service.

## Checklist

- [ ] Create package in `saas/pkg/<servicename>/`
- [ ] Define service struct with dependencies
- [ ] Add constructor with dependency injection
- [ ] Register in `app.Plugin` if needed globally

## File Location

```
saas/pkg/<servicename>/
├── service.go      # Main service logic
├── types.go        # Input/output types (if many)
└── service_test.go # Tests
```

## Do

- Accept dependencies via constructor (not globals)
- Pass `context.Context` as first parameter
- Return `error` as last return value
- Wrap errors with operation context
- Keep services focused (single responsibility)

## Don't

- Call `app.GetPlugins()` inside service methods (inject instead)
- Mix HTTP/GraphQL concerns into services
- Create services for single-use utilities

## Service Pattern

```go
package myservice

import (
    "context"
    "fmt"

    "dbent/gen/ent"
)

type Service struct {
    client *ent.Client
    // Add other dependencies
}

func New(client *ent.Client) *Service {
    return &Service{client: client}
}

func (s *Service) DoSomething(ctx context.Context, input DoInput) (*Result, error) {
    // Business logic here
    result, err := s.client.User.Get(ctx, input.UserID)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return &Result{User: result}, nil
}

type DoInput struct {
    UserID string
}

type Result struct {
    User *ent.User
}
```

## Using in Resolvers

```go
// In resolver
func (r *queryResolver) MyQuery(ctx context.Context) (*Result, error) {
    svc := myservice.New(r.Plugin.EntDB.Client())
    return svc.DoSomething(ctx, myservice.DoInput{...})
}
```

## Using Globally (if needed)

```go
// In saas/pkg/app/app.go - add to Plugin struct
type Plugin struct {
    // ...existing fields
    MyService *myservice.Service
}

// In New() function
plugins.MyService = myservice.New(plugins.EntDB.Client())
```

## Template

See `.claude/volt/templates/service.go.tmpl`
