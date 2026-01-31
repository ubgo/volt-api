# Development Workflow Guide

This guide outlines the end-to-end process of adding a new feature to the `ai_golang_boilerplate` project. The stack relies heavily on code generation, so following the correct order of operations is crucial.

## The "Backend-First" Flow

The typical workflow for adding a new feature (e.g., a "Product" entity) follows this path:

1. **Database (Ent):** Define the data structure.
2. **Code Generation (Ent):** Generate Go structs and DB client code.
3. **Migration:** Update the local database schema.
4. **API Contract (GraphQL):** Define how the client interacts with the data.
5. **Code Generation (GQL):** Generate API interfaces and resolver stubs.
6. **Implementation:** Write the actual business logic connecting GQL to Ent.

***

## Step 1: Define Database Schema (Ent)

Navigate to the `dbent` module. Ent schemas define your database tables.

1. Create a new schema file in `dbent/schema/`.
   * *Example:* `dbent/schema/product.go`
2. Define your Fields and Edges.

```go
// dbent/schema/product.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type Product struct {
    ent.Schema
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        field.Float("price").Min(0),
    }
}
```

## Step 2: Generate Ent Code

Once the schema is defined, generate the Go code (client, entities, builders).

```bash
# From root
task dbent:entg
```

> **Tip:** If you see errors about missing imports, run `go mod tidy` in the `dbent` directory.

## Step 3: Update Local Database

Apply the schema changes to your local Postgres container.

```bash
# From root
task dev           # Ensure DB is running
task migrate       # Runs the migration tool
```

***

## Step 4: Define GraphQL Schema

Navigate to the `apidash` module.

1. Create or edit a `.graphqls` file in `apidash/internal/graph/schemas/`.
   * *Example:* `apidash/internal/graph/schemas/product.graphqls`

```graphql
type Product {
  id: ID!
  name: String!
  price: Float!
}

extend type Query {
  product(id: ID!): Product
}

extend type Mutation {
  createProduct(name: String!, price: Float!): Product!
}
```

## Step 5: Generate GraphQL Code

This step reads your GraphQL files and generates the Go interfaces and resolver stubs.

```bash
# From root
task apidash:gql
```

This will:

1. Update `apidash/internal/graph/generated/generated.go` (System code - DO NOT EDIT).
2. Update `apidash/internal/graph/model/models_gen.go` (Go structs for GQL types).
3. **Create/Update** `apidash/internal/graph/resolver/product.resolvers.go` (Your implementation stubs).

***

## Step 6: Implement Logic (Resolvers)

Open the newly generated/updated resolver file (e.g., `apidash/internal/graph/resolver/product.resolvers.go`). You will see methods with `panic("not implemented")`.

**Replace the panic with actual logic using the Ent client.**

```go
// apidash/internal/graph/resolver/product.resolvers.go

func (r *mutationResolver) CreateProduct(ctx context.Context, name string, price float64) (*model.Product, error) {
    // Access the Ent client via the Plugin (Dependency Injection)
    p, err := r.Plugin.EntDB.Client().Product.Create().
        SetName(name).
        SetPrice(price).
        Save(ctx)

    if err != nil {
        return nil, err
    }

    // Convert Ent entity to GQL model (if structures differ, map them manually)
    return &model.Product{
        ID:    p.ID,
        Name:  p.Name,
        Price: p.Price,
    }, nil
}
```

***

## Step 7: Commit Generated Artifacts

This project follows the pattern of **committing generated code** to ensure build consistency and easier CI deployments. Since hundreds of files can be generated, we use a specialized command to stage and commit them with a standard message.

```bash
# From root
task gitgen
```

This command automatically:

1. Stages Ent generated files (`dbent/gen/...`).
2. Stages GraphQL generated files (`apidash/internal/graph/generated/...`).
3. Stages Config bindings (`config/gen/...`).
4. Commits them with the message `"gen"`.

> **Note:** Run this *after* you have verified your changes and before you commit your manual code changes (or as part of the same PR).

## Checklist Before Commit

1. [ ] Run `task workvendor` to sync dependencies.
2. [ ] Run `task lint` to ensure code style.
3. [ ] Run `task gitgen` to stage and commit generated artifacts.
4. [ ] Run `task test` to verify nothing is broken.

***

## Adding a New CLI Command

All CLI tools are located in the `saas` module to leverage the core business logic.

1. **Open `saas/cmd/cli/main.go`.**
2. **Define a constructor function** for your command:
   ```go
   func newMyCommand(plugin app.Plugin) *cobra.Command {
       return &cobra.Command{
           Use:   "my-command",
           Short: "Description of my command",
           RunE: func(cmd *cobra.Command, args []string) error {
               // Your logic here
               return nil
           },
       }
   }
   ```
3. **Register the command** in the `main` function:
   ```go
   root.AddCommand(newMyCommand(plugin))
   ```
4. **Run/Test:**
   ```bash
   cd saas
   go run ./cmd/cli my-command
   ```

***

## Recipe: Building a CLI Scraper

If you are building a web scraper, follow this pattern to keep your code organized.

### 1. Define the Data Model

Create an Ent schema for the data you are scraping.

* `dbent/schema/scraped_item.go`

### 2. Implement the Scraper Logic

Don't put logic in the CLI command. Create a dedicated package.

* **Path:** `saas/pkg/scraper/scraper.go`
* **Structure:**
  ```go
  package scraper

  import (
      "context"
      "dbent/gen/ent"
  )

  type Scraper struct {
      Client *ent.Client
  }

  func (s *Scraper) ScrapeTarget(ctx context.Context, url string) error {
      // 1. Fetch HTML (use http.Get or a library like colly)
      // 2. Parse Data
      // 3. Save to DB
      return s.Client.ScrapedItem.Create().SetURL(url).Exec(ctx)
  }
  ```

### 3. Create the CLI Command

Register a new command in `saas/cmd/cli`.

* **Path:** `saas/cmd/cli/cmd/scrape.go`
* **Code:**
  ```go
  var scrapeCmd = &cobra.Command{
      Use:   "scrape [url]",
      Short: "Scrapes the given URL",
      RunE: func(cmd *cobra.Command, args []string) error {
          // Initialize dependencies (DB, etc.) - provided by boilerplate bootstrapper
          client := app.GetEntClient()
          srv := scraper.Scraper{Client: client}
          return srv.ScrapeTarget(cmd.Context(), args[0])
      },
  }
  ```
