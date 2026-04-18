# Backend Patterns

<!-- Canonical Go patterns specific to this codebase. -->

## Entry Point Boot

```go
func main() {
    plugin := app.New(
        app.WithAfterNatsInitiated(registerHandlers),
    )
    defer plugin.Shutdown()
    // start server / scheduler / cli
}
```

## Ent Entity (with BaseMixin)

```go
type Product struct{ ent.Schema }

func (Product) Mixin() []ent.Mixin {
    return []ent.Mixin{mixin.BaseMixin{Prefix: "prd"}}  // id: prd_xxxxx
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        field.String("sku").Unique(),
        field.Enum("status").Values("draft", "published").Default("draft"),
    }
}

func (Product) Edges() []ent.Edge {
    return []ent.Edge{edge.From("owner", User.Type).Ref("products").Unique()}
}
```

## Resolver (thin delegation)

```go
func (r *queryResolver) Product(ctx context.Context, id string) (*ent.Product, error) {
    return r.Plugin.EntDB.Client().Product.Get(ctx, id)
}
```

## CLI Command

```go
func AddCommands(plugin *app.Plugin) {
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "...",
        Run: func(cmd *cobra.Command, args []string) { /* ... */ },
    }
    plugin.Cli.RootCmd.AddCommand(cmd)
}
```

## Options Pattern

```go
type Option func(*config)

func WithTimeout(d time.Duration) Option {
    return func(c *config) { c.timeout = d }
}

func New(opts ...Option) *Client {
    cfg := &config{timeout: 30 * time.Second}
    for _, opt := range opts { opt(cfg) }
    return &Client{cfg: cfg}
}
```
