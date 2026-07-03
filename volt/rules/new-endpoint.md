# New GraphQL Endpoint Rules

**Read before:** Adding a new GraphQL query or mutation.

## Checklist

- [ ] Define schema in `apidash/internal/graph/*.graphqls`
- [ ] Run `task apidash:gql`
- [ ] Implement resolver in `apidash/internal/graph/resolver/`
- [ ] Test with GraphQL playground

## File Locations

| What | Where |
|------|-------|
| Schema definitions | `apidash/internal/graph/*.graphqls` |
| Generated code | `apidash/internal/graph/generated/` (don't edit) |
| Resolvers | `apidash/internal/graph/resolver/` |
| Custom models | `apidash/internal/graph/model/` |

## Do

- Define clear input/output types in schema
- Use Ent types for database entities (auto-mapped)
- Access services via `r.Circuit` in resolvers
- Use `@internal` directive for admin-only endpoints

## Don't

- Create helper files in `resolver/` (gqlgen will overwrite)
- Use `graphql.Upload` for files (use base64 string instead)
- Edit files in `generated/` directory

## Schema Example

```graphql
# In apidash/internal/graph/user.graphqls

type Query {
  user(id: ID!): User
  users(first: Int, after: Cursor): UserConnection!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
}

input CreateUserInput {
  email: String!
  firstName: String
  lastName: String
}
```

## Resolver Example

```go
// In resolver/user.resolvers.go (after running task apidash:gql)

func (r *queryResolver) User(ctx context.Context, id string) (*ent.User, error) {
    return r.Circuit.EntDB.Client.User.Get(ctx, id)
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*ent.User, error) {
    return r.Circuit.EntDB.Client.User.
        Create().
        SetEmail(input.Email).
        SetNillableFirstName(input.FirstName).
        Save(ctx)
}
```

## File Upload Pattern

```graphql
input FileUploadInput {
  fileName: String!
  content: String!  # Base64 encoded
}
```

```go
import "encoding/base64"
import "net/http"

fileContent, err := base64.StdEncoding.DecodeString(input.Content)
if err != nil {
    return nil, fmt.Errorf("decode file: %w", err)
}

// Validate by magic bytes, not filename
mime := http.DetectContentType(fileContent)
```
