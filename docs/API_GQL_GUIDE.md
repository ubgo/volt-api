# GraphQL API Development Guide

This project uses [99designs/gqlgen](https://github.com/99designs/gqlgen) to generate type-safe GraphQL servers. We have one main API gateway:

* `apidash`: The Dashboard/Admin API.

## The Golden Rule: NO Helpers in Resolver Files

**Crucial Warning:** Do not define helper functions, private methods, or structs inside `*.resolvers.go` files.

### Why?

The `gqlgen` tool **owns** these files. Every time you run `task gql`:

1. It scans the `.graphqls` schema.
2. It matches schema fields to Go methods in the resolver.
3. **It removes or comments out** any Go code that does not match the schema.

### What happens if I write a helper function?

If you add `func (r *mutationResolver) myHelper() {...}` manually:

* **Result:** The next time you run codegen, `gqlgen` will see that `myHelper` is not in your GraphQL schema. It will move your code to the bottom of the file and wrap it in `/* deleted ... */` comments, effectively "deleting" it from the build.

***

## Configuration (`gqlgen.yml`)

The `gqlgen.yml` file controls how code is generated.

### 1. Auto-Binding (`autobind`)

We automatically map Ent database models to GraphQL types.

```yaml
autobind:
  - dbent/gen/ent
  - dbent/gen/ent/user
```

**Benefit:** If you have a `User` type in GraphQL and a `User` struct in Ent with matching fields (e.g., `Name`, `Email`), `gqlgen` will use the Ent struct directly. You don't need to write a converter function!

### 2. Custom Scalars & Models

We use specialized Go types for certain GraphQL scalars:

```yaml
models:
  JSON:
    model: lace/jsontype.JSON  # Handles JSON blobs from DB
  Date:
    model: lace/gqlmodel.Date  # Handles Date formatting
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int # Allows using Int IDs as GQL IDs
```

***

## Directives

We use schema directives to handle authorization declaratively.

### Available Directives

| Directive | Usage | Description |
| :--- | :--- | :--- |
| `@auth` | `field: String! @auth` | Ensures the user is logged in (Authentication). |
| `@role` | `field: String! @role(required: "admin")` | Ensures the user has a specific role (Authorization). |

### How to Use

Add the directive to your schema definition (`.graphqls`):

```graphql
extend type Mutation {
  # Only logged-in users can update their profile
  updateProfile(name: String!): User! @auth

  # Only admins can delete users
  deleteUser(id: ID!): Boolean! @role(required: "admin")
}
```

**Under the hood:**
The logic for these directives resides in `apidash/internal/gqldirective/auth.go`. The resolver function (`updateProfile`) will **only** be called if the directive check passes.

***

## Workflow: Adding a New API

### 1. Define the Schema

Navigate to `apidash/internal/graph/schemas/`. Create or edit a `.graphqls` file.

```graphql
# apidash/internal/graph/schemas/todo.graphqls

type Todo {
  id: ID!
  text: String!
  done: Boolean!
}

extend type Query {
  todos: [Todo!]! @auth
}

extend type Mutation {
  createTodo(text: String!): Todo! @auth
}
```

### 2. Generate Code

Run the generation task to create the interfaces and resolver stubs.

```bash
task apidash:gql
```

This will:

1. Update `apidash/internal/graph/generated/generated.go` (System code - DO NOT EDIT).
2. Update `apidash/internal/graph/model/models_gen.go` (Go structs for GQL types).
3. **Create/Update** `apidash/internal/graph/resolver/todo.resolvers.go` (Your implementation stubs).

***

### 3. Implement Logic (Resolvers)

Open the newly generated/updated resolver file (e.g., `apidash/internal/graph/resolver/todo.resolvers.go`). You will see methods with `panic("not implemented")`.

**Replace the panic with actual logic using the Ent client.**

**Implementation Strategy:**
Resolvers should be **thin wrappers**. They should only:

1. **Validate** input using [valgo](https://github.com/cohesivestack/valgo).
2. **Authenticate** using `middleware.GetUserFromGqlCtx(ctx)`.
3. **Call** the logic layer (e.g., `saas/pkg/todofn` or `apidash/internal/apifn`).
4. **Handle Errors** using `lace/cerr`.

```go
package resolver

import (
    "context"
    "lace/cerr"
    "saas/pkg/middleware"
    "saas/pkg/predefinederrors"
    "github.com/cohesivestack/valgo"
)

func (r *mutationResolver) CreateTodo(ctx context.Context, text string) (*model.Todo, error) {
    // 1. Auth Check (Double check, though @auth directive handles it usually)
    user, err := middleware.GetUserFromGqlCtx(ctx)
    if err != nil {
        return nil, cerr.NewFromCtx(ctx).SetFriendlyFromErrs(predefinederrors.ErrUnauthenticated, err)
    }

    // 2. Validation
    val := valgo.Is(valgo.String(text, "text").Not().Blank())
    if err := val.Error(); err != nil {
         return nil, cerr.NewFromCtx(ctx).SetFriendlyFromErrs(err, err)
    }

    // 3. Call Logic (located in saas/pkg/...)
    item, err := r.Plugin.EntDB.Client().Todo.Create().
        SetText(text).
        SetUserID(user.ID).
        Save(ctx)

    if err != nil {
        return nil, cerr.NewFromCtx(ctx).SetFriendly("failed to create todo", err.Error())
    }

    // 4. Map to GQL Model
    return &model.Todo{ID: item.ID, Text: item.Text}, nil
}
```

## Architecture & Best Practices

### 1. Logic Placement (No Helpers!)

Do NOT write private helper functions inside `resolver` packages. Instead, create specialized packages:

* **Business Logic:** `saas/pkg/yourdomainfn` (e.g., `saas/pkg/userfn`).
* **API Logic:** `apidash/internal/apifn` (if specific to GQL).

### 2. Error Handling (`lace/cerr`)

We use `lace/cerr` to provide user-friendly errors while logging technical details.

* `SetFriendly(msg, log)`: Shows `msg` to user, logs `log`.
* `SetFriendlyFromErrs(publicErr, logErr)`: Wraps existing errors.

### 3. Validation (`valgo`)

Use the fluent `valgo` API for robust input validation.
