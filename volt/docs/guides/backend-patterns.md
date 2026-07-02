# Backend Rules (Go)

## Package Layout

- Business logic → `modules/saas/pkg/<feature>/`
- Generic utilities → `modules/lace/<package>/` (read-only — requires separate PR)
- Ent schemas → `modules/dbent/schema/`
- GraphQL schemas → `modules/apidash/internal/graph/*.graphqls`
- Resolvers → `modules/apidash/internal/graph/resolver/` (**no helpers**, gqlgen overwrites)
- CLI commands → `modules/saas/cmd/cli/`
- Cron jobs → `modules/saas/cmd/cron/` + `modules/saas/pkg/cronfn/` for reusable functions

## Services

```go
type Service struct {
    client *ent.Client
}

func New(client *ent.Client) *Service {
    return &Service{client: client}
}

func (s *Service) Do(ctx context.Context, input Input) (*Result, error) {
    // ...
}
```

Accept deps via constructor — no globals, no `app.GetPlugins()` inside service methods.

## Database

- ALWAYS `plugin.EntDB.Client()` — never `ent.Open()` directly
- Migrations: edit `modules/dbent/schema/`, run `task dbent:g && task dbent:migrate`
- Foreign keys disabled by default (`WithForeignKeys(false)`) to avoid ordering issues
- Sensitive fields: `field.String("password").Sensitive()`

## GraphQL

- Schema-first: edit `.graphqls`, run `task apidash:gql`
- Ent types auto-map — use them directly as GraphQL types
- File uploads: `String!` (base64), decode with `base64.StdEncoding.DecodeString()`
- Admin-only: `@internal` directive (requires `X-Internal-Key` header)
- Extract user via `auth.UserFromContext(ctx)` in resolver — don't trust client headers

## Cron Jobs

```go
scheduler.NewJob(
    gocron.DurationJob(5*time.Minute),
    gocron.NewTask(func() {
        ctx, span := otel.Tracer("cron").Start(context.Background(), "cron.myJob")
        defer span.End()
        // work
    }),
    gocron.WithName("myJob"),
    gocron.WithSingletonMode(gocron.LimitModeReschedule),
)
```

Always: singleton mode, OTEL span, recover from panic (don't crash scheduler).

## NATS

Register in `WithAfterNatsInitiated` callback — NATS connects async after `app.New()`.

```go
plugin := app.New(
    app.WithAfterNatsInitiated(func(p *app.Plugin) {
        js, _ := p.Natso.GetConn().JetStream()
        js.Subscribe("orders.created", handler, nats.Durable("order-processor"))
    }),
)
```

Ack on success, `Nak()` for retry, `Term()` for poison messages.

## Errors

```go
return nil, fmt.Errorf("getUser %s: %w", id, err)  // wrap with context
```

Log at boundaries (handler/job start), not in every function. Don't log AND return.

## Logging

```go
applog.FromCtx(ctx).Info("user created",
    ulog.String("user_id", user.ID),
    ulog.Dur("elapsed", time.Since(start)),
)
```

`FromCtx(ctx)` adds `trace_id`/`span_id` automatically. Never use `fmt.Printf` for logs.

## Tracing

```go
ctx, span := otel.Tracer("component").Start(ctx, "component.operation")
defer span.End()
span.SetAttributes(attribute.String("entity.id", id))
// on error:
span.SetStatus(codes.Error, err.Error())
span.RecordError(err)
```

Span naming: `<component>.<operation>` (e.g., `order.create`, `cron.cleanup`, `nats.orders.created`).

## Validation

Use `valgo` for business validation:

```go
val := valgo.Is(
    valgo.String(input.Email, "email").Not().Empty(),
    valgo.Number(input.Amount, "amount").GreaterThan(0),
)
if !val.Valid() { return val.Errors() }
```

File uploads: validate by magic bytes (`http.DetectContentType`), never extension.

## Testing

- `_test.go` beside source
- Table-driven with `t.Run()` subtests
- Ent: `enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")`
- HTTP: `httptest.NewServer()`
- Skip if external deps missing: `t.Skipf("redis not available")`
