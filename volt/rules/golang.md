# Go backend rules (Volt API)

**Read before ANY change to Go code.** This is the standing Go taste for this repo, grounded in the actual v2 stack. It pairs with [code-standards.md](code-standards.md) (the universal engineering bar — type safety, no hardcoded closed-set values, context-dense comments, modular/no-slop) and the per-action guides in this folder (endpoints, schemas, migrations, …). When a topic has a dedicated guide, this doc points to it rather than repeating it.

## 0 · The 3 commandments

1. **No bare strings for any closed-set value.** Status, kind, sort key, role, scope, source — each gets a typed constant in a `constants.go` beside the package that owns it, plus a `…Values` slice for iteration. A typo becomes a compile error; a rename is one place. (See §2 and [code-standards.md §1.2](code-standards.md).)
2. **No hardcoded enum lists in the data layer.** A `field.Enum(...)` baked into an ent schema locks user-facing taxonomy into the schema — growing it needs a migration and it can't be tenant-scoped. Model those as **reference tables** (§4). The current schemas hold no `field.Enum`; keep it that way.
3. **No business validation at the DB level.** No triggers, no CHECK constraints beyond what ent emits structurally, no hardening SQL. The DB is dumb storage; validation + integrity live in the app layer (§5). Sharding compatibility is the load-bearing reason.

If you're about to violate one of these, stop and ask.

## 1 · The v2 stack

Use the same libraries everywhere; extend the wrapper, don't bypass it. This repo is a **hybrid**: `lace/*` wraps the HTTP/DB/CLI/gqlgen layer, while **`github.com/ubgo/*`** owns observability, errors, cache, and infra. There is **no `lace/gozap`, no `gotel`, no `goerr`** — those were replaced in v2.

| Concern | v2 choice | Notes |
|---|---|---|
| HTTP server | `gin` via `lace/ginserver` | otelgin instrumentation wired |
| GraphQL | `gqlgen` (+ `entgo.io/contrib` entgql) | schema-first; resolvers thin, one-line-delegated |
| ORM | `entgo.io/ent` | `BaseMixin` for id/timestamps; see §12 |
| CLI | `cobra` via `lace/cli` | |
| Config | `github.com/ubgo/cfgkit` | typed structs in `modules/config/config.go`; secrets tagged `secret:"true"` |
| Validation | `valgo` (`github.com/cohesivestack/valgo`) | business rules; ent handles structural |
| **Logging** | **`github.com/ubgo/logger`** via `core/pkg/applog` (aliased `ulog`) | `applog.FromCtx(ctx)` propagates trace IDs. NOT `lace/gozap`/`zap`. |
| **Tracing** | **`github.com/ubgo/otelkit`** | one-constructor OTel bootstrap. NOT `lace/gotel`. |
| **Errors** | stdlib `errors` + `fmt.Errorf("…: %w")`, structured via **`github.com/ubgo/errx`** | NOT `goerr`. |
| Postgres driver | `pgx/v5` via `lace/db` | |
| Cache | `github.com/ubgo/cache` + `github.com/ubgo/cache-redis` | interface + Redis impl |
| Messaging | NATS JetStream + `github.com/ThreeDotsLabs/watermill` | register in `WithAfterNatsInitiated` (see [nats-handler.md](nats-handler.md)) |
| Scheduler | `github.com/go-co-op/gocron/v2` | see [new-cron-job.md](new-cron-job.md) |
| Health | `github.com/ubgo/health` (+ gin, redis contribs) | `/api/i/bootinfo` |
| Shutdown | `github.com/ubgo/shutdown` | graceful, phased |
| Build metadata | `github.com/ubgo/buildinfo` | version/commit/modules |
| Encryption | `github.com/ubgo/crypt` | AES; see [secrets.md](secrets.md) |
| IDs | `BaseMixin{Prefix: "…"}` (nanoid via `go-nanoid`) | prefixed ids (`prd_…`); never roll your own |

The full ecosystem — wired-in vs available-to-add (`ubgo/entx`, `ubgo/lock`, `ubgo/publicid`, the `cache-*` family, …) — is catalogued in [../docs/ubgo-packages.md](../docs/ubgo-packages.md).

**Do not use:** `encoding/json` for validation (use `valgo`) · `database/sql` directly (use `ent`) · stdlib `log` / `fmt.Println` for app output (use `applog`) · `net/http` server directly (use `gin` via `lace/ginserver`) · `math/rand` for tokens/ids/nonces (use `crypto/rand`).

## 2 · Typed constants over string literals

Per-package `constants.go` at the package root, grouped by entity, with a `…Values` slice as the canonical iteration order (drives seeds, UI lists, histograms — read and write share one source).

```go
// pkg/foo/constants.go
const (
    StatusActive   = "active"
    StatusDraft    = "draft"
    StatusArchived = "archived"
)
var StatusValues = []string{StatusActive, StatusDraft, StatusArchived}
```

Needs a constant if the value: appears in 2+ switch cases / map keys, is compared with `==` to a literal, is a closed enumeration the API consumer must match, or appears in both seed data and runtime code. Free-form user text (titles, search queries) stays a plain string. Upstream wire-format fixtures (Shopify/Stripe field names) stay verbatim in `testdata/` — translate INTO your constant, never rename the fixture.

## 3 · Polymorphic relations — entpoly

For any polymorphic owner relation (an `owner_type` + `owner_id` discriminator pair), use **`github.com/ubgo/entx` entpoly** (`MorphMixin` + `MorphTo` + `MixinAllowed` + a module-namespaced `MixinIndexName`). Hand-rolled `owner_kind` enum columns are banned — they drift from real ent types and are convention-only. entpoly is in the ubgo ecosystem ([../docs/ubgo-packages.md](../docs/ubgo-packages.md)); add it when you introduce a polymorphic table. Read paths use the emitted `*MorphKey` constants and typed `OwnerTypeEQ` predicates, never raw strings.

## 4 · Reference tables, not `field.Enum`

Any user-facing closed enumeration the schema models (status, kind, type, …) lives in a **reference table** (a `code`/`label`/`active` row, tenant-scopable via a nullable `team_id`), not a hardcoded `field.Enum`. Validate the code against the table in the app layer before insert/update; seed the global rows from the typed constants in §2. Adding a value becomes a `task seed` / one INSERT instead of a migration, and values can be tenant-scoped. The only acceptable `field.Enum` is the one entpoly emits internally for `owner_type`.

## 5 · DB as dumb storage

All validation, integrity, and hardening live in the **app layer**. Do NOT add triggers, CHECK constraints beyond what ent emits structurally, `NULLS NOT DISTINCT` hardening, polymorphic-FK enforcement, or any hand-rolled `*-hardening.sql`. Acceptable (rides along free): PK + unique indexes ent emits, the entpoly `MixinAllowed` validator, and ent migrate with `WithForeignKeys(false)` (the project default). Validation instead lives in: the service entry (`valgo`, §13), the resolver (authz + request shape), and the ent schema (structural only — `.NotEmpty()`, `.MaxLen()`, `.Unique()`).

## 6 · Errors

Wrap with operation context using `%w`: `return nil, fmt.Errorf("getUser %s: %w", id, err)`. Sentinel errors (`var ErrNotFound = errors.New("svc: not found")`) for expected conditions; `errors.As` for structured access; **`github.com/ubgo/errx`** when you need stable identity / redactable fields / a wire codec. Error is always the last return value. Never silently ignore an error (`_ =` only for documented optional-cleanup, with a comment). Log once, at the boundary — never log AND return (§10, [error-handling.md](error-handling.md)).

## 7 · Context

Any call that can block accepts `context.Context` as the **first** argument and respects cancellation + deadline. Never accept `context.TODO()` from a caller — push the requirement up. Background goroutines derive their own context from the parent; don't capture a request's `ctx` and outlive it.

## 8 · Naming

Packages: singular, short, specific (`httputil` not `httputils`, `dbhook` not `hook`, never `utils`). Structs read the relationship in order — "group has many users" → `GroupUser`. Exported names get a doc comment starting with the symbol name. Column/field names avoid bare nouns — see §12.

## 9 · Service / package design

Accept interfaces, return structs. Inject dependencies via constructor — no package-level globals, no a global `*volt.Kernel` inside service methods. One responsibility per package; split past ~600 lines. Never `panic` for operational errors (only init-time fatals); **`log.Fatal` is banned** — it kills graceful shutdown, return the error to `main`.

## 10 · Logging — `applog` (ubgo/logger)

Always `core/pkg/applog` (wraps `github.com/ubgo/logger`, aliased `ulog`). Never stdlib `log`/`fmt.Println`/`fmt.Printf` for app output. There is no `gozap`.

```go
import (
    "core/pkg/applog"
    ulog "github.com/ubgo/logger"
)

applog.FromCtx(ctx).Info("operation done", ulog.String("user_id", userID), ulog.Any("count", n))
applog.GetLogger().Warn("rate limit approaching", ulog.Any("current", n))
```

Log at the boundary (handler, resolver, job/cron entry); services return errors, callers log. Structured fields only — `ulog.Err(err)`, `ulog.String`, `ulog.Dur`, `ulog.Bool`, and `ulog.Any` for ints/floats/slices (there is no `Int`/`Strings`). `FromCtx(ctx)` auto-adds `trace_id`/`span_id`. Never log secrets — mark sensitive ent fields `.Sensitive()`. Full field reference: [logging.md](logging.md).

## 11 · Tracing — OTel via `otelkit`

Telemetry is bootstrapped with **`github.com/ubgo/otelkit`** (not `lace/gotel`); `applog`'s OTel sink ships logs through the same pipeline once `otelkit` registers the provider. Add a span at each meaningful boundary, named `<component>.<operation>`, set status + `RecordError` on failure. DB/HTTP/NATS are auto-traced when `ctx` flows through. Details + the bootstrap sequence: [tracing.md](tracing.md), [../docs/guides/observability.md](../docs/guides/observability.md).

## 12 · ORM rules (ent)

- **Generated field constants, never string field names:** `Order(ent.Desc(user.FieldCreatedAt))`, not `"created_at"`. Applies to predicates, ordering, group-by.
- **IDs via `BaseMixin{Prefix: "…"}`** (nanoid-backed prefixed ids like `prd_…`). Don't roll your own id scheme.
- **Field comments at the declaration**; when a value takes multiple shapes by writer, enumerate them right there.
- **Foreign keys:** `migrate.WithForeignKeys(false)` (app-layer validation). Configure once at the migration layer.
- **Sensitive fields:** `field.String("password").Sensitive()` — stripped from logs.
- **No bare-noun columns.** A column name must be explicit: add a role-suffix (`_status`, `_at`, `_id`, `_kind`, `_url`, `_count`) or domain-prefix (`payout_status` not `payment`; `customer_phone` not `phone`) — bare nouns collide with current/future entities, read ambiguously at call sites, and flood search. Smell test: if `git grep '\bX\b'` matches a table AND a column AND a Go var, `X` is too generic.

Schema workflow: [new-schema.md](new-schema.md) · migrations: [migrations.md](migrations.md).

## 13 · Validation — `valgo`

Business rules via `valgo` at the **service entry**, once, at the boundary:

```go
val := valgo.Is(
    valgo.String(in.CustomerID, "customer_id").Not().Empty(),
    valgo.Number(in.Amount, "amount").GreaterThan(0),
)
if !val.Valid() { return fmt.Errorf("validation failed: %v", val.Errors()) }
```

Layers: gqlgen enforces type/nullability; resolver does authz + request-shape; service entry does the `valgo` block; ent schema does structural constraints. Details: [validation.md](validation.md).

## 14 · Multi-tenant safety

When the API serves multiple tenants (workspaces/teams) from one DB, **every** resolver/query scopes to the caller's tenant, and the tenant id comes from the **authenticated context — never** a client-supplied input field, header, or URL param. A single middleware pins the workspace id from the session/JWT into a typed context key; downstream reads it via a `WorkspaceFromContext(ctx)` accessor and filters every query by it. Owner/role-restricted operations are enforced **server-side** at the resolver (the client gate is UX only); centralize membership/role checks in an auth service, not inline per resolver. Audit before merging: `rg "WorkspaceID\(.*input\.|WorkspaceID\(.*header" -t go` must return zero hits. See [authentication.md](authentication.md).

## 15 · Go gotchas (not covered by a dedicated guide)

- **Concurrency:** every goroutine has a clear exit; pair `go` with `ctx` + `<-ctx.Done()`. Sender closes channels. Use `errgroup.WithContext` (not bare `WaitGroup`) for fan-out where one failure cancels siblings. `go test -race ./...` in CI is non-negotiable. Watch the pre-1.22 loop-variable capture in plain `for i` loops.
- **Defer:** capture the error from a deferred `Close` on writers (flush failures matter) via a named return; don't `defer` inside a long loop (wrap the iteration in a closure); `defer` evaluates args immediately, runs the call later.
- **Time:** always zone-aware (`time.Now().UTC()`); never `==` on `time.Time` (use `.Equal()`); strip monotonic reading (`.UTC()`/`.Round(0)`) before storing; never `time.Sleep` to wait for a condition in prod.
- **Receivers:** one type uses pointer OR value receivers, not both. Never return a typed nil through an interface return type (the `err != nil` check passes on a "nil" error) — return literal `nil`.
- **HTTP client:** always a timeout (client or per-request `context.WithTimeout`); reuse one client per dependency; drain + close the response body; status codes are not errors — check `resp.StatusCode`.
- **`context.Value`:** request-scoped data only (trace id, identity, tenant) — never config or optional params. Typed unexported key + exported accessor; callers never `ctx.Value(...)` directly.
- **Constructors:** functional options for > 3 optional fields (not a giant Config struct or a long positional list). `iota` for related numeric constants; explicit string values for API-facing enums (§2).
- **JSON / random:** explicit `json:"…"` tags on every serialized field; `omitempty` is a footgun on `int`/`bool` (use `*int`/`*bool` when absent ≠ zero); `crypto/rand` for tokens/ids/nonces, `math/rand/v2` only for non-security jitter.
- **`init()`:** avoid except driver/codec registration; lift setup into an explicit constructor.

## 16 · Codegen & gqlgen resolver rule

Codegen output is never hand-edited — change the source, re-run the task (`task dbent:g`, `task apidash:gql`), commit the regenerated files. gqlgen **owns** `*.resolvers.go`: those files hold **only** resolver methods (plus the `Resolver` struct in `resolver.go`). Never put a free function, helper type, shared map, or const in a `*.resolvers.go` file — gqlgen relocates or drops it on the next generate. Helpers go in a sibling package (`internal/<domain>fn/`); resolver bodies stay one-line-delegated to a service. Codegen-driven helpers ship a runnable example that regenerates + builds + asserts (see [new-endpoint.md](new-endpoint.md), [testing.md](testing.md)).

## 17 · Secrets (summary)

Real credentials live in `.env` (gitignored) and nowhere else — never in a `.go`, `.md`, comment, example, or commit message, not even "to rotate later." Docs/examples use placeholders (`<TOKEN>`, `shpat_xxx`). Sensitive ent fields use `.Sensitive()`; API-key compares use `subtle.ConstantTimeCompare`; encrypt at rest with `ubgo/crypt`. The pre-commit secret scanner — **gitleaks**, via `task hooks:install` (also `task secrets:scan` / `secrets:scan:staged`) — is defense-in-depth; self-scan the staged diff before every commit regardless. Full policy: [secrets.md](secrets.md), [git-commit.md](git-commit.md).

## 18 · Anti-pattern catalog

| Anti-pattern | Instead |
|---|---|
| `case "active":` / `return "draft"` (bare strings) | typed constants (§2) |
| `Order(ent.Desc("created_at"))` | generated `user.FieldCreatedAt` (§12) |
| `field.Enum("status").Values(...)` in schema | reference table (§4) |
| hand-rolled `owner_kind` column | entpoly `MorphMixin` + `MorphTo` (§3) |
| triggers / CHECK / hardening SQL | app-layer validation (§5) |
| `gozap` / `zap` / `fmt.Println` for logs | `applog` + `ulog` fields (§10) |
| `lace/gotel` / `goerr` | `otelkit` (§11) / `errx` (§6) |
| `panic(err)` / `log.Fatal` in service code | return error to `main` (§9) |
| package-level globals | inject via constructor (§9) |
| bare-noun column (`payment`, `note`, `kind`, `phone`) | role-suffix / domain-prefix (§12) |
| tenant id from input / header / URL | `WorkspaceFromContext(ctx)` (§14) |
| helper/type/const in a `*.resolvers.go` | sibling `internal/<domain>fn/` package (§16) |
| bare `go func()` with no exit | pair with `ctx` + `<-ctx.Done()` (§15) |
| `defer f.Close()` (error dropped) on a writer | captured-err named return (§15) |
| `t1 == t2` on `time.Time` | `.Equal()` (§15) |
| typed nil returned through an interface | explicit `nil` return (§15) |
| `http.Client{}` with no timeout | timeout + reused per dep (§15) |
| `ctx.Value("user", u)` string key | typed unexported key + accessor (§15) |
| `math/rand` for tokens/ids | `crypto/rand` (§15) |
| hardcoded secret in code / committed `.env` | a config field with `env:` + `secret:"true"` + gitignore (§17) |

## 19 · Pre-PR checklist

Before a Go change is "done":

- [ ] §0 — no new bare-string closed-set value; no new `field.Enum` (reference table instead); no DB-level business validation.
- [ ] §1 — no direct dep on a library `lace/*` or `ubgo/*` already wraps; no reintroduced `gozap`/`gotel`/`goerr`.
- [ ] §6/§7 — errors wrap context with `%w`; every blocking call takes + uses `ctx`.
- [ ] §10/§11 — `applog` logging at boundaries; `otelkit`/OTel spans where they matter.
- [ ] §12 — ent queries use generated field constants; sensitive fields `.Sensitive()`; every new column has a role-suffix/domain-prefix (no bare nouns).
- [ ] §13 — business validation via `valgo` at the service entry.
- [ ] §14 — tenant-scoped resolvers source the workspace id from context, not input (`rg "WorkspaceID\(.*input\." -t go` → zero hits).
- [ ] §15 — goroutines have a clear exit; deferred writer-Close errors captured; times zone-aware; http clients have timeouts.
- [ ] §16 — no non-resolver code in `*.resolvers.go`; codegen re-run and committed.
- [ ] §17 — no hardcoded secrets; staged diff self-scanned; `git config --local core.hooksPath` = `.githooks`.
- [ ] §18/migrations — if schema changed, backup + staged migration; row counts preserved ([migrations.md](migrations.md)).
- [ ] `task lint` (gofmt + vet) clean · `go build ./...` clean · tests pass · `go mod tidy` a no-op.
- [ ] [code-standards.md](code-standards.md) pillars applied (type safety, comments, modularity).
