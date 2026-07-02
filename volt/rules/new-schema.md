# New Ent Schema Rules

**Read before:** Adding a new database entity or modifying existing schema.

## Checklist

- [ ] Create/edit schema in `dbent/schema/`
- [ ] Run `task dbent:g` to generate code
- [ ] Run `task dbent:migrate` to update database
- [ ] Verify GraphQL types updated (if using ent-gqlgen)

## File Locations

| What | Where |
|------|-------|
| Schema definitions | `dbent/schema/*.go` |
| Generated code | `dbent/gen/ent/` (don't edit) |
| Mixins | `dbent/schema/mixin/` |
| Migration runner | `dbent/pkg/dbent_migrate/` |

## Do

- Use mixins for common fields (`BaseMixin` adds id, created_at, updated_at)
- Add indexes for frequently queried fields
- Use `Sensitive()` for passwords and secrets
- Set `Unique()` on fields like email
- Add edges (relations) with clear naming

## Don't

- Edit files in `dbent/gen/ent/`
- Forget to run migrations after schema changes
- Skip `Sensitive()` on password fields

## New Entity Example

```go
// dbent/schema/product.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "dbent/schema/mixin"
)

type Product struct {
    ent.Schema
}

func (Product) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.BaseMixin{Prefix: "prd"},  // ID: prd_xxxxx
    }
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        field.String("sku").Unique(),
        field.Float("price"),
        field.Bool("is_active").Default(true),
        // status_code references a ProductStatus reference-table row — NOT a
        // field.Enum. Validate the code app-side before write. See golang.md §4.
        field.String("status_code").Default("draft"),
    }
}

func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).Ref("products").Unique(),
        edge.To("categories", Category.Type),
    }
}

func (Product) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("sku"),
        index.Fields("status_code", "is_active"),
    }
}
```

## The 3 commandments for schemas (see [golang.md](golang.md))

- **Reference tables, not `field.Enum`** (§4). A user-facing closed enumeration (status, kind, type) lives in a reference table (`code`/`label`/`active`, tenant-scopable via nullable `team_id`), validated app-side. `field.Enum` locks taxonomy into the schema — growing it needs a migration and it can't be tenant-scoped. The only allowed `field.Enum` is the one entpoly emits for `owner_type`.
- **Polymorphic owner relations use entpoly** (§3) — `github.com/ubgo/entx` `MorphMixin` + `MorphTo` + `MixinAllowed` + a module-namespaced index name. Never a hand-rolled `owner_kind` column.
- **No bare-noun columns** (§12). Every column gets a role-suffix (`_status`, `_at`, `_id`, `_kind`, `_url`, `_count`) or domain-prefix (`payout_status` not `payment`; `customer_phone` not `phone`). Bare nouns collide with current/future entities and read ambiguously.
- **DB is dumb storage** (§5) — no triggers, CHECK functions, or hardening SQL; structural constraints (`.NotEmpty()`, `.Unique()`, `.MaxLen()`) only. `WithForeignKeys(false)` is the project default.

## Adding Field to Existing Entity

```go
// Add to Fields() slice
field.String("new_field").Optional().Nillable(),
field.Time("processed_at").Optional().Nillable(),
```

## Common Field Patterns

| Pattern | Code |
|---------|------|
| Required string | `field.String("name").NotEmpty()` |
| Optional string | `field.String("bio").Optional().Nillable()` |
| Unique field | `field.String("email").Unique()` |
| Default value | `field.Bool("active").Default(true)` |
| Enum / closed set | reference table + `field.String("x_code")` — NOT `field.Enum` (see [golang.md §4](golang.md)) |
| Sensitive | `field.String("password").Sensitive()` |
| JSON field | `field.JSON("metadata", map[string]any{})` |

## After Schema Changes

```bash
# Always run both commands
task dbent:g     # Generate Go code
task dbent:migrate  # Update database
```
