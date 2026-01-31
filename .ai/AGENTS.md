# Detailed Agent Rules

> **Quick Start:** See root `AGENTS.md` for directory map, task commands, and basic rules.

This file contains detailed rules for specific scenarios.

---

## Code Generation Rules

**CRITICAL:** Never modify generated files. Always edit the source.

| Generated Path | Source | Regenerate |
|---------------|--------|------------|
| `dbent/gen/ent/*` | `dbent/schema/*.go` | `task dbent:entg` |
| `apidash/internal/graph/generated/*` | `apidash/internal/graph/*.graphqls` | `task apidash:gql` |
| `config/gen/*` | `config/*.pkl` | `task config` |

---

## GraphQL Resolver Rules

- **Only** resolver implementations in `resolver/` directory
- **Never** create helper files in `resolver/` (gqlgen overwrites)
- Put helpers in: `internal/<package>/`, `lace/`, or `saas/pkg/`

---

## File Upload Pattern

**Always** use base64, **never** use `graphql.Upload`.

```graphql
input FileUploadInput {
  fileName: String!
  content: String!  # Base64 encoded
}
```

```go
// Validate by magic bytes, not filename
fileContent, _ := base64.StdEncoding.DecodeString(input.Content)
mime := http.DetectContentType(fileContent)
```

---

## Coding Standards

| Rule | Example |
|------|---------|
| Context first | `func Do(ctx context.Context, ...)` |
| Error last | `func Do(...) (*Result, error)` |
| Wrap errors | `fmt.Errorf("action: %w", err)` |
| Imports grouped | stdlib → third-party → local |
| Packages | `snake_case` (no dashes) |
| Exported | `PascalCase` |
| Unexported | `camelCase` |

---

## Common Pitfalls

| Pitfall | Fix |
|---------|-----|
| Editing `gen/` files | Edit source, run generator |
| Helpers in `resolver/` | Move to `internal/` or `pkg/` |
| `graphql.Upload` | Use base64 string |
| `go get` in workspace | Use `task workvendor` |
| Forgetting migrations | Always run `task dbent:migrate` after schema changes |
| Reverting `go.work` | Don't, unless intentional |

---

## Testing Guidelines

- Table-driven tests in `_test.go` beside source
- Mock external services (DB, Redis, HTTP)
- Cover: domain services, resolvers, migrations, error cases
- Run: `task test` or `go test ./...`

---

## Commit Guidelines

- Present-tense messages: `feat(scope): add user lookup`
- Generated code: `task gitgen` → commit as `gen`
- **Never** commit `.env` values
- **Never** add AI attribution or co-author lines
