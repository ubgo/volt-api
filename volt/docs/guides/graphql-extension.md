# GraphQL extension pattern (Pattern D from ADR-V2-012)

Module ships a base `User` type. Host adds fields without touching the module.

## Module schema (ships in `modules/mod-auth/graphql/user.graphqls` once GraphQL
lands in the module — currently the module exposes only Go API):

```graphql
type User {
    id: ID!
    email: String!
    firstName: String
    lastName: String
    isActive: Boolean!
    createdAt: Time!
}

extend type Query {
    user(id: ID!): User
    users(limit: Int, offset: Int): [User!]!
}

extend type Mutation {
    createUser(input: CreateUserInput!): User!
    deleteUser(id: ID!): Boolean!
}
```

## Host extension (`apidash/internal/graph/schemas/user_ext.graphqls`):

```graphql
extend type User {
    dateOfBirth: Time
    phone: String
    timezone: String
}
```

gqlgen merges both schemas at codegen → unified `User` type from the GraphQL
client's perspective.

## Host field resolvers (`apidash/internal/graph/resolver/user_ext.resolvers.go`):

```go
package resolver

import (
    "context"
    "time"

    authapi "github.com/khanakia/boilerplate-golang/modules/mod-auth/api"
)

func (r *userResolver) DateOfBirth(ctx context.Context, obj *authapi.User) (*time.Time, error) {
    if v, ok := obj.Metadata["date_of_birth"].(string); ok {
        if t, err := time.Parse(time.RFC3339, v); err == nil {
            return &t, nil
        }
    }
    return nil, nil
}

func (r *userResolver) Phone(ctx context.Context, obj *authapi.User) (*string, error) {
    if v, ok := obj.Metadata["phone"].(string); ok {
        return &v, nil
    }
    return nil, nil
}

func (r *userResolver) Timezone(ctx context.Context, obj *authapi.User) (*string, error) {
    if v, ok := obj.Metadata["timezone"].(string); ok {
        return &v, nil
    }
    return nil, nil
}
```

## When sidecar table lands

Once `dbent/schema/user_profile.go` exists, swap the metadata reads for
typed sidecar lookups behind a DataLoader:

```go
func (r *userResolver) DateOfBirth(ctx context.Context, obj *authapi.User) (*time.Time, error) {
    p, err := r.loaders.UserProfile.Load(ctx, obj.ID)   // batched per request
    if err != nil { return nil, err }
    return &p.DateOfBirth, nil
}
```

DataLoader batches all `UserProfile.Load` calls in one GraphQL query into
a single `SELECT ... WHERE user_id IN (...)` — no N+1.

## Why this pattern wins

- Module's GraphQL contract stays small + stable.
- Host extends without forking module.
- Multiple consumers add different fields without conflict.
- gqlgen's schema-merge is the only mechanism doing real work — no custom
  middleware, no schema stitching service.
