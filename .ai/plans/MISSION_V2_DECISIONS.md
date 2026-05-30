---
title: Mission V2 — Architecture Decisions
status: in-review
owner: khanakia
created: 2026-04-25
defaults:
  project: mission-v2
  area: work/architecture
  owner: khanakia
---

# Mission V2 — Architecture Decisions

Captures the decisions made for V2's pluggable, distributable module architecture. Each section lists the options considered, pros/cons, and the locked decision. **Final Decisions** at bottom is the executive summary.

<Callout type="info" title="Reading order">

If you're new: skim this doc top-to-bottom. The decisions build on each other — earlier choices constrain later ones.

If you're returning: jump to **Final Decisions** at the bottom for the locked summary, then back-fill rationale only when needed.

</Callout>

## Context — why V2

Today's pain in the boilerplate:

- `app.Plugin` is a god-struct. Every feature reaches into `r.Plugin.EntDB`, `r.Plugin.Natso`, etc. → shotgun coupling, untestable.
- Features spread across 5+ dirs (`saas/pkg/`, `apidash/`, `dbent/schema/`, cron main, NATS callback). Removing one = many edits.
- `Root.pkl` is hand-merged. Adding a feature config = edit central file.
- No way to ship a feature to another project without copy-paste.

**V2 goal:** features = distributable Go modules, importable across projects, swappable, individually versioned. Find the sweet spot between portability and Go's compile-time realities.

---

## Decision 1 — What stays coupled (immovable)

<Callout type="warning" title="Accept these — fighting them creates worse problems">

These constraints are forced by Go + the toolchain. V2 works AROUND them, not against them.

</Callout>

| constraint | reason | fixable |
|---|---|---|
| ent codegen is whole-package | entc.LoadGraph reads ONE schema dir | no |
| gqlgen codegen needs all schemas | one ResolverRoot per gqlgen module | no |
| PKL Root.pkl shape codegen'd | pkl-gen-go reads typed PKL | partial |
| Single ent.Client per app for shared tx | ent design | no |
| Module bumps require host rebuild | Go is statically compiled | no |

**Decision:** accept. Build the rest of V2 to make these constraints invisible from the outside.

---

## Decision 2 — Module distribution model

### Options

| Option | What | Reuse across projects | Versioning |
|---|---|---|---|
| A. Monorepo dirs | `modules/<name>/` inside boilerplate | no | locked together |
| B. Separate Go modules | `github.com/khanakia/mod-<name>` | yes | independent |
| C. Internal package replication | copy-paste via tooling | hacky | none |

### Tradeoffs

| option | setup-cost | maintenance | reuse | level |
|---|---|---|---|---|
| A. Monorepo dirs | low | low | no | low |
| B. Separate Go modules | med | med | yes | med |
| C. Replication | high | high | partial | high |

<ADR status="accepted" id="ADR-V2-002" date="2026-04-25" title="Modules are separate Go modules with own repos">

### Context

User stated requirement: "I want to be able to distribute and use in multiple projects." Monorepo dirs (A) defeat this. Replication (C) is operationally awful.

### Decision

Each feature module is its own Go module, own GitHub repo, own SemVer. Path convention: `github.com/khanakia/mod-<domain>` (e.g. `mod-iam`, `mod-billing`, `mod-workspace`).

### Consequences

- **Win:** publish, version, share across N projects. Other consumers `go get` what they need.
- **Win:** module dev cycle independent of host.
- **Cost:** more repos to maintain. Each has own CI, releases, README.
- **Cost:** local dev across modules requires `go.work` + `replace` directives.
- **Forces:** every architectural decision below must respect the rule "module never imports host".

</ADR>

---

## Decision 3 — Contract repo (modkit)

The chicken-and-egg: module needs services from host (DB, logger), but cannot import host (path varies). Solution: **interface-only package** that both sides import.

### Options

| Name | What | Pros | Cons |
|---|---|---|---|
| `modkit` | short, generic | brand-neutral | generic |
| `modcore` | descriptive | clear | vague |
| `appkit` | app-flavored | onboarding-friendly | implies "app" not "framework" |
| `<brand>kit` | branded | identity | renames are painful |

<Callout type="danger" title="This name is forever">

Once published, every module ever shipped references it. Renaming = breaking change for every consumer. Decide once.

</Callout>

<ADR status="accepted" id="ADR-V2-003" date="2026-04-25" title="Adopt `modkit` as the contract repo">

### Context

Module → host → module circular import is impossible in Go. Standard solution: introduce a tiny interface-only package both sides import. This is how `database/sql/driver`, `io.Reader`, Hashicorp plugins all work.

### Decision

Create `github.com/khanakia/modkit` as the canonical contract repo. Contains:

- `Services` interface — host contract (DB, logger, tracer, publisher, clock, WithTx)
- `Module` interface — `Name()`, `Init(Services) error`, `Shutdown(ctx) error`, `DependsOn() []string`
- Registrar interfaces — `CronRegistrar`, `NATSRegistrar`, `CLIRegistrar`, `HTTPRegistrar`, `GQLResolverRegistrar`
- Tx context helpers — `TxFromContext`, `WithTxContext`
- Generic value types — `PageRequest`, `Metadata = map[string]any`, etc.

Strict rules:

- **Interfaces only**, no impl. Tiny value types OK.
- **Near-zero deps.** Allowed: `zap`, `otel/trace`, `database/sql/driver`. **Not allowed:** `ent`, `gin`, NATS, gqlgen.
- **Versioned with extreme care.** Adding a method = breaking change. Use `Services` + `ServicesV2` if you must extend.
- Aim for <300 LOC total.

### Consequences

- Module imports `modkit` only. Host imports `modkit` + concrete modules.
- Host can be **any module path** (`github.com/anyone/their-app`). No special name required.
- Adding a method to `Services` is a breaking change. Discipline required.

</ADR>

---

## Decision 4 — Module repo layout (api/ contracts + impl)

Once we have separate Go modules, how do they reference each other? **Per-module `api/` sub-package** holds the contract.

### Options

| Option | Where contracts live | Cross-module dep weight |
|---|---|---|
| A. Module ships impl only; host writes Store adapters | host glue per module | high |
| B. Module ships own `api/` contracts sub-package | tiny contracts package | low |
| C. Central contracts repo (`mod-contracts`) for all domains | one mega-repo | medium, but bottleneck |

<ADR status="accepted" id="ADR-V2-004" date="2026-04-25" title="Per-module api/ sub-package for cross-module contracts">

### Context

Modules need to reference each other (mod-billing wants user data from mod-iam). Direct impl import drags ent + heavy deps + couples implementations. We want each cross-module reference to cost a single thin import.

### Decision

Every module has an `api/` sub-package containing only:

- DTOs (plain structs): `User`, `Workspace`, `Subscription`
- Reader interfaces: `Reader { Get, GetMany, ListByOwner, ... }`
- Writer interfaces: `Writer { Create, Update, Delete, *By<Foreign>, ... }`
- Domain events: `WorkspaceCreated`, `InvoicePaid`

Other modules import only `<module>/api`. The impl can churn freely.

```
mod-workspace/
├── go.mod
├── api/                    ← contracts, tiny, importable by other modules
│   ├── reader.go
│   ├── writer.go
│   ├── types.go
│   └── events.go
├── module.go               ← impl, imports ent + heavy deps
├── deps.go
├── service.go
├── schema/
│   └── workspace.go
├── gen/ent/                ← own ent client (committed)
└── config/
    ├── Config.pkl
    ├── defaults.pkl
    ├── embed.go
    └── load.go
```

### Consequences

- mod-billing imports `mod-workspace/api` — pulls only contracts, no ent/Postgres deps.
- Module impl can refactor without breaking consumers (as long as `api/` is stable).
- `api/` stability is the new "API contract" — versioned with semver discipline.
- Adding a method to `Reader`/`Writer` = breaking change.

</ADR>

---

## Decision 5 — No foundational tier; all modules are peers

Earlier draft proposed a "mod-iam owns User, others reference it" tier. User pushed back: "many cases — workspaces, billing, etc. — none more foundational than another."

<ADR status="accepted" id="ADR-V2-005" date="2026-04-25" title="All modules are peers — no foundational tier">

### Context

Initial draft of MISSION_V2.md had a "foundational tier" with mod-iam owning User and other modules referencing it natively. This overfits to identity. Real apps reference workspace, billing, organizations, projects, audit-logs. None is more "foundational"; they're peers that get referenced.

### Decision

Drop the tier concept. Every module is a peer. Every module ships its own `api/`. Other modules reference what they need via that `api/` import. Soft FK by string ID across modules.

### Consequences

- Pattern generalizes infinitely — add `mod-projects`, `mod-audit`, `mod-notifications` with same shape.
- No artificial hierarchy; consumers wire what they need.
- modkit stays infra-only. Domain entities never appear in modkit.
- Cost: lose the temptation of cross-module ent edges (all references are soft FK).

</ADR>

---

## Decision 6 — Database / ORM strategy

If each module is its own Go module, can it share host's `*ent.Client`? **No.** Host's gen path is `github.com/anyone/host/dbent/gen/ent` — module can't import it.

### Options

| Option | Module imports ent? | Cross-module ent edges | Per-module ent client | Module portability |
|---|---|---|---|---|
| A. Each module owns its ent client; soft FK across | yes | no | yes | high |
| B. Module declares Store iface; host writes adapter using aggregated ent | no | yes (host-level) | n/a | highest |
| C. Aggregate at host, modules use host's client | impossible — host path varies | n/a | n/a | impossible |

| aspect | A_own_client | B_store_iface |
|---|---|---|
| typed-queries-inside-module | 5 | 2 |
| boilerplate-per-module | 5 | 2 |
| ORM-agnostic | 2 | 5 |
| cross-module-SQL-JOIN | 1 | 4 |
| host-glue-cost | 5 | 1 |
| module-dev-DX | 5 | 2 |

<ADR status="accepted" id="ADR-V2-006" date="2026-04-25" title="Pattern A — each module owns its ent client; soft FK across modules">

### Context

The lossy part of B is real: hexagonal stores + DTO mapping per module = lots of glue per module addition. The lossy part of A is real too: cross-module SQL JOINs are gone, replaced with batch fetch + service-layer assembly.

### Decision

Each module ships its own ent schemas (`<module>/schema/`) and its own committed ent client (`<module>/gen/ent/`). Cross-module references are **soft FK by string ID** at the DB level, with corresponding `api.Reader.GetMany` batch lookups for service-layer joins.

Host injects the SQL driver via `modkit.Services.DB() driver.Driver`. Each module wraps it:

```go
func (m *Module) Init(s modkit.Services) error {
    m.client = wsent.NewClient(wsent.Driver(s.DB()))
    return nil
}
```

### Consequences

- **Win:** typed `m.client.Workspace.Query()` inside module. Great DX, no DTO mapping inside module.
- **Win:** module is self-contained — schemas, client, queries all in one repo.
- **Cost:** no native SQL JOINs across modules. Service-layer batch fetch instead.
- **Cost:** N+1 risk. Mitigated by `GetMany` batch APIs and DataLoader at GraphQL boundary.
- **Cost:** each module runs own `task entg`. Host doesn't need a single aggregated client for module business logic.
- **Optional escape hatch:** host MAY have its own aggregated ent client for cross-cutting admin/reporting queries. Independent codegen pass. Module logic never touches it. YAGNI until proven need.

</ADR>

---

## Decision 7 — PKL config strategy

Each module needs config. Should host be able to import module's PKL schema for typed composition?

### Options

| Option | What | Path stability | Setup cost |
|---|---|---|---|
| 1. `go mod vendor` + `//go:embed` | module embeds PKL; vendor copies it; host imports `vendor/.../Config.pkl` | only with vendor | low |
| 2. Symlink-stabilized path via `task pkl:link` | `go list -m` finds module dir; symlinks to `config/pkl/modules/<name>/` | always | low |
| 3. PKL package registry | publish PKL packages to `pkg.pkl-lang.org` or self-host | always | high |
| 4. Host configures via Go only; no PKL import | duplicate field shape in host's PKL | always | low |

<Callout type="info" title="Path 4 was rejected">

User: "no host won't duplicate, we need importability." Rules out approaches that re-declare module config shape in host's PKL.

</Callout>

<ADR status="accepted" id="ADR-V2-007" date="2026-04-25" title="Path 2 — symlink-stabilized PKL paths via task pkl:link">

### Context

Module ships PKL schema as source of truth. Host wants typed import without redeclaring the shape. Path differs by dev mode (vendor vs go.work replace vs module cache). We need one stable import path regardless.

### Decision

Each module embeds its PKL via `//go:embed *.pkl` (forces files to survive `go mod vendor`). Host runs a `task pkl:link` step that uses `go list -m -f '{{.Dir}}'` to find each module's filesystem location and creates symlinks (or copies on Windows) to `config/pkl/modules/<name>/`.

Root.pkl imports use the stable path:

```pkl
import "../modules/iam/Config.pkl" as Iam
import "../modules/billing/Config.pkl" as Billing

modules {
  iam = new Iam.Setting { jwtSecret = read("env:JWT_SECRET") }
  billing = new Billing.Setting { stripeKey = read("env:STRIPE_KEY") }
}
```

`task config:gen` runs `task pkl:link` first, then `pkl-gen-go`.

### Consequences

- Same import path under: vendor mode, module cache mode, `go.work` replace, CI.
- Host's typed config struct includes module config shapes natively (no duplication).
- Cost: one extra task step before codegen. Symlinks awkward on Windows; copy fallback.
- Future: if going public registry / multi-language, upgrade to Path 3.

</ADR>

---

## Decision 8 — Cross-module JOINs

Soft FK across modules → no SQL JOIN. How do we handle queries that need data from multiple modules?

### Options

| Option | What | Performance | Complexity |
|---|---|---|---|
| A. Service-layer batch fetch | call each module's `GetMany`, zip in code | O(modules) round trips | low |
| B. DataLoader at GraphQL resolver | batch automatically per request | excellent | medium |
| C. Host's aggregated ent client | real SQL JOIN at host level | best | high |
| D. Materialized views / read models | denormalize via events | best | very high |

<ADR status="accepted" id="ADR-V2-008" date="2026-04-25" title="Service-layer batch + DataLoader at GraphQL boundary; aggregated client only as escape hatch">

### Context

Most JOINs in app code are 2-3 entities, batch-able. GraphQL hits the N+1 trap most often, and DataLoader solves that idiomatically. SQL-level JOINs are only critical for analytics/admin reports.

### Decision

- **Default pattern:** service calls multiple modules' `GetMany` batch APIs and assembles results in Go.
- **GraphQL pattern:** DataLoader at the resolver layer. Each module exposes a loader per entity. Field resolvers for cross-module references use loaders.
- **Reserved escape hatch:** host MAY create an aggregated ent client (separate codegen, all schemas) for admin/reporting queries that genuinely need SQL JOIN. Module business logic never uses it. Build this only when a real query forces the issue.
- **Outlawed:** materialized views via events (D) — too much complexity for modular monolith stage.

### Consequences

- Two queries instead of one for cross-module reads — measure if it actually hurts.
- Every Reader iface MUST include `GetMany(ctx, ids) (map[string]T, error)`. Lock as convention.
- DataLoader is mandatory for GraphQL; document in module-author guide.

</ADR>

---

## Decision 9 — Cross-module transactions

Same Postgres → atomic tx achievable. Module ent clients can share `*sql.Tx`.

<ADR status="accepted" id="ADR-V2-009" date="2026-04-25" title="modkit owns WithTx + TxFromContext; module clients adapt to ctx-tx">

### Context

Module ent clients are independent objects but share the underlying SQL driver. ent supports binding a client to a specific `*sql.Tx`. We need a single source of truth for tx propagation.

### Decision

`modkit.Services` exposes:

```go
type Services interface {
    DB() driver.Driver
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
    // ...
}
```

modkit also provides `TxFromContext(ctx) (*sql.Tx, bool)` and `WithTxContext(ctx, tx)` helpers.

Each module implements a `clientForCtx(ctx)` private helper that returns either the baseline client OR a tx-bound client when `TxFromContext(ctx)` returns a tx. All `Reader`/`Writer` impl methods MUST use `clientForCtx(ctx)`.

### Consequences

- Cross-module atomic operations work via host orchestration:
  ```go
  services.WithTx(ctx, func(ctx context.Context) error {
      if err := workspaceMod.DeleteByOwner(ctx, userID); err != nil { return err }
      if err := billingMod.CancelByUser(ctx, userID); err != nil { return err }
      return iamMod.DeleteUser(ctx, userID)
  })
  ```
- Lock acquisition spans all participating modules — keep tx scope narrow.
- Module impl boilerplate: every method needs `m.clientForCtx(ctx)` not `m.client`. Easy to forget; lint with a check.

</ADR>

---

## Decision 10 — Cascade-delete strategy

Delete user → billing rows + workspaces need cleanup.

### Options

| Strategy | Consistency | Coupling | Use when |
|---|---|---|---|
| 1. Sync orchestrated in one tx | strong | host knows topology | data-integrity cascades (billing, sessions, workspaces tied to user) |
| 2. Event-driven (outbox + NATS) | eventual | none | reactive side effects (search index, audit log mirror, email) |
| 3. Soft delete + nightly sweeper | hybrid | medium | UI shows deletion immediately, hard-delete is forgiving |

<ADR status="accepted" id="ADR-V2-010" date="2026-04-25" title="Strategy 1 for hot path cascades; Strategy 2 for reactive side effects">

### Context

Real apps need both. Sync-tx for "data must be consistent at commit boundary" (billing must cancel before user is gone). Async events for "this should happen eventually" (search index, audit log). Mixing both is fine; rule must be explicit.

### Decision

- **Cascade orchestration lives in host**, not in a module. Host has a `lifecycle/` package with functions like `DeleteUser(ctx, userID)` that call each module's `*By<Foreign>` Writer methods inside `services.WithTx`.
- Each module's `api.Writer` MUST expose cascade hooks: `DeleteByOwner(ctx, ownerID)`, `CancelByUser(ctx, userID)`, etc. Convention: `<Action>By<ForeignType>(ctx, foreignID)`.
- For non-critical reactions (analytics, search, audit mirror), modules emit NATS events from inside their `api.Writer` impl (using outbox pattern). Listening modules handle async.
- **Rule:** if losing the side effect breaks data integrity → Strategy 1. If reactive/observability → Strategy 2.

### Consequences

- Adding a new module that holds user-scoped data → update host's `DeleteUser` orchestrator. Acceptable single point of change.
- Cascade methods on `api.Writer` are forever API. Plan once per module.
- Async handlers need DLQ + retry semantics — decide at module level.

</ADR>

---

## Decision 11 — Schema extension (host adds fields)

Consumer wants to add `dateOfBirth` + `phone` to User. mod-iam's schema is sealed (separate Go module).

### Options

| Pattern | What | Typed? | DB-level FK | Best for |
|---|---|---|---|---|
| 1. Sidecar table at host | host owns `user_profile` with `user_id` FK | yes | soft FK | 3+ typed fields |
| 2. JSON metadata column on module entity | module ships `metadata jsonb` | no (app-level) | no | 1-2 loose fields |
| 3. Fork module | maintain own version of mod-iam | yes | yes | structural mismatch |

<ADR status="accepted" id="ADR-V2-011" date="2026-04-25" title="Sidecar table for typed extensions; metadata jsonb for loose; fork is escape hatch">

### Context

You cannot extend a module's ent schema from outside — ent codegen seals it. Three real ways to add fields, each with a clear use case.

### Decision

- **Default pattern:** every "extensible" entity in modules ships a `metadata jsonb` column + `Metadata map[string]any` field on its `api/` DTO. Free insurance for loose data.
- **Typed extension pattern:** host creates a sidecar table (e.g. `user_profile`) in host's own dbent, with `user_id` as soft FK. Host owns its own ent client over its sidecar tables.
- **Fork:** acceptable last resort when the module's abstraction is structurally wrong for your domain. Acknowledge the cost; don't pretend modules fit every case.
- **modkit.Metadata** = type alias for `map[string]any` + helpers `EncodeMetadata` / `DecodeMetadata`.

### Consequences

- mod-iam.User ships `Metadata` field always. All consumers can stuff loose data.
- Typed sidecar requires host-level dbent setup (small ent client just for sidecars).
- Migrations: module migrations and host's sidecar migrations are independent; run module first, then host.
- No DB-level FK constraint between sidecar and module table — soft FK in app code only.

</ADR>

---

## Decision 12 — Returning extra fields via API

mod-iam.api.Reader.GetUser is sealed. Sidecar fields live at host. How do consumers get enriched data?

### Options

| Pattern | Where enrichment happens | Module API change |
|---|---|---|
| A. Host owns `users.Service.Get` returning `EnrichedUser` | host service layer | none |
| B. Stuff sidecar into Metadata | mod-iam impl writes through, untyped | none |
| C. Decorate mod-iam.Reader with host-level wrapper | host wraps interface | none |
| D. GraphQL `extend type User` + field resolvers | resolver layer | none |
| E. Other module needs sidecar data → caller passes as parameter | call site | none |
| F. Other module needs sidecar data → inject provider iface (e.g. `ContactProvider`) | module deps grow | none |

<ADR status="accepted" id="ADR-V2-012" date="2026-04-25" title="Layered composition — module API stays sealed; host stitches enriched DTOs">

### Context

Module API ≠ host API. Module gives primitives. Host stitches the final shape. Consumers split into three categories with different patterns.

### Decision

Three rules, by consumer type:

**Internal Go code** (host services, handlers, jobs):
- Host owns a single composition service per extensible entity (e.g. `users.Service.Get` returns `EnrichedUser` embedding `iamapi.User` + sidecar fields).
- Single composition point — don't sprinkle profile-fetching across handlers.
- Pattern A.

**GraphQL clients:**
- mod-iam ships base `User` GraphQL type. Host adds `extend type User { dateOfBirth: Time; phone: String; ... }` in own `.graphqls`.
- Host writes field resolvers for the new fields, using DataLoader against sidecar table.
- gqlgen merges schemas at codegen → unified `User` type from client perspective.
- Pattern D.

**Other modules** (e.g. mod-billing needs phone for invoice):
- Default: caller passes enriched data as parameter. mod-billing's API takes `Contact{Email, Phone}` not just `userID`. Host enriches before calling. Pattern E.
- For structural deps (used in many places inside the module): module declares `ContactProvider` interface; host implements via its `users.Service`. Pattern F.
- **Forbidden:** module imports host's sidecar — that's the wrong direction.

### Consequences

- Module API stays small, stable, and useful as a primitive.
- Host owns its application-specific types (`EnrichedUser`).
- GraphQL extension is the cleanest UI-facing pattern; works with gqlgen schema merging.
- DataLoader is mandatory for GraphQL field resolvers hitting sidecar.

</ADR>

---

## End-to-end walkthrough — code edition

A complete trace: "delete user → cascade workspaces + billing, with sidecar `user_profile` extension, returning enriched data via GraphQL". Every layer shown with real code.

<Callout type="tip" title="Read this if the architecture doesn't click from text alone">

The earlier sections describe what each piece *does*. This section shows what each piece *looks like*. Same architecture, just with files you can copy-paste from.

</Callout>

### Repos involved

```
github.com/khanakia/modkit              ← contract repo (interfaces only)
github.com/khanakia/mod-iam             ← User/Tenant/Auth
github.com/khanakia/mod-billing         ← Subscriptions/Invoices
github.com/khanakia/mod-workspace       ← Workspaces
github.com/anyone/their-app             ← host (any path; this boilerplate is one example)
```

### Layer 1 — modkit (the contract, ~200 LOC)

```go
// github.com/khanakia/modkit/services.go
package modkit

import (
    "context"
    "database/sql/driver"
    "time"

    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
)

type Services interface {
    DB() driver.Driver
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
    Log(ctx context.Context) *zap.Logger
    Tracer(name string) trace.Tracer
    Publish(ctx context.Context, subject string, payload []byte) error
    Now() time.Time
}
```

```go
// github.com/khanakia/modkit/module.go
package modkit

import "context"

type Module interface {
    Name() string
    Init(Services) error
    Shutdown(ctx context.Context) error
    DependsOn() []string
}

type Metadata = map[string]any
```

```go
// github.com/khanakia/modkit/tx.go
package modkit

import (
    "context"
    "database/sql"
)

type ctxKey int
const txKey ctxKey = 1

func WithTxContext(ctx context.Context, tx *sql.Tx) context.Context {
    return context.WithValue(ctx, txKey, tx)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
    tx, ok := ctx.Value(txKey).(*sql.Tx)
    return tx, ok
}
```

```go
// github.com/khanakia/modkit/registrars.go
package modkit

import (
    "github.com/gin-gonic/gin"
    "github.com/spf13/cobra"
)

type CronRegistrar     interface { RegisterCron(scheduler any) error }
type NATSRegistrar     interface { RegisterNATS(js any) error }
type CLIRegistrar      interface { RegisterCLI(root *cobra.Command) }
type HTTPRegistrar     interface { RegisterHTTP(r gin.IRouter) }
type GQLResolverRegistrar interface { RegisterResolvers(reg any) }
```

### Layer 2 — mod-iam (owns User; one of N peer modules)

Repo layout:

```
mod-iam/
├── go.mod                      ← module github.com/khanakia/mod-iam
├── api/                        ← contracts — other modules import this only
│   ├── reader.go
│   ├── writer.go
│   └── types.go
├── schema/
│   └── user.go                 ← ent schema
├── gen/ent/                    ← committed entc output
├── config/
│   ├── Config.pkl              ← schema source
│   ├── defaults.pkl
│   └── embed.go                ← //go:embed *.pkl
├── deps.go
└── module.go                   ← impl
```

```go
// mod-iam/api/types.go — DTOs
package api

import "time"

type User struct {
    ID, Email, Name string
    Metadata        map[string]any   // host stuffs loose data here
    CreatedAt       time.Time
}
```

```go
// mod-iam/api/reader.go — interface other modules import
package api

import "context"

type Reader interface {
    Get(ctx context.Context, id string) (User, error)
    GetMany(ctx context.Context, ids []string) (map[string]User, error)   // N+1 friendly
    GetByEmail(ctx context.Context, email string) (User, error)
}
```

```go
// mod-iam/api/writer.go
package api

import "context"

type Writer interface {
    Create(ctx context.Context, in CreateInput) (User, error)
    Update(ctx context.Context, id string, in UpdateInput) (User, error)
    Delete(ctx context.Context, id string) error
}

type CreateInput struct { Email, Name string; Metadata map[string]any }
type UpdateInput struct { Name *string; Metadata map[string]any }
```

```go
// mod-iam/schema/user.go — ent schema, owned by module
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("id"),
        field.String("email").Unique(),
        field.String("name"),
        field.JSON("metadata", map[string]any{}).Optional(),
        field.Time("created_at").Default(time.Now),
    }
}
```

```go
// mod-iam/module.go — impl satisfies api.Reader + api.Writer + modkit.Module
package iam

import (
    "context"
    "database/sql"
    "encoding/json"

    "entgo.io/ent/dialect"
    entsql "entgo.io/ent/dialect/sql"

    "github.com/khanakia/modkit"
    "github.com/khanakia/mod-iam/api"
    "github.com/khanakia/mod-iam/config"
    iament "github.com/khanakia/mod-iam/gen/ent"
    "github.com/khanakia/mod-iam/gen/ent/user"
)

type Module struct {
    cfg    config.Config
    client *iament.Client
    s      modkit.Services
}

func New(cfg config.Config) *Module { return &Module{cfg: cfg} }

func (m *Module) Name() string         { return "iam" }
func (m *Module) DependsOn() []string  { return nil }
func (m *Module) Shutdown(ctx context.Context) error { return m.client.Close() }

func (m *Module) Init(s modkit.Services) error {
    m.s = s
    drv := entsql.OpenDB(dialect.Postgres, sql.OpenDB(s.DB()))
    m.client = iament.NewClient(iament.Driver(drv))
    return nil
}

// clientForCtx — THE tx-aware helper. Every method uses this, not m.client.
func (m *Module) clientForCtx(ctx context.Context) *iament.Client {
    if tx, ok := modkit.TxFromContext(ctx); ok {
        // wrap tx as ent driver — simplified
        return iament.NewClient(iament.Driver(entsql.NewDriver(dialect.Postgres, entsql.Conn{ExecQuerier: tx})))
    }
    return m.client
}

// ---- api.Reader impl ----

func (m *Module) Get(ctx context.Context, id string) (api.User, error) {
    u, err := m.clientForCtx(ctx).User.Get(ctx, id)
    if err != nil { return api.User{}, err }
    return toAPI(u), nil
}

func (m *Module) GetMany(ctx context.Context, ids []string) (map[string]api.User, error) {
    us, err := m.clientForCtx(ctx).User.Query().Where(user.IDIn(ids...)).All(ctx)
    if err != nil { return nil, err }
    out := make(map[string]api.User, len(us))
    for _, u := range us { out[u.ID] = toAPI(u) }
    return out, nil
}

// ---- api.Writer impl ----

func (m *Module) Delete(ctx context.Context, id string) error {
    if err := m.clientForCtx(ctx).User.DeleteOneID(id).Exec(ctx); err != nil { return err }
    payload, _ := json.Marshal(map[string]string{"user_id": id})
    return m.s.Publish(ctx, "iam.user.deleted", payload)
}

func toAPI(u *iament.User) api.User {
    return api.User{
        ID: u.ID, Email: u.Email, Name: u.Name,
        Metadata: u.Metadata, CreatedAt: u.CreatedAt,
    }
}
```

### Layer 2 — mod-billing (depends on mod-iam contracts only)

```go
// mod-billing/deps.go — what billing needs from outside
package billing

import (
    "github.com/khanakia/modkit"
    iamapi "github.com/khanakia/mod-iam/api"   // ← contracts only, no ent/heavy deps
)

type Deps struct {
    Core modkit.Services
    IAM  iamapi.Reader   // soft dep — any impl satisfying iamapi.Reader works
}
```

```go
// mod-billing/api/writer.go — note CancelByUser cascade hook
package api

import "context"

type Writer interface {
    Create(ctx context.Context, in CreateInput) (Subscription, error)
    Cancel(ctx context.Context, id string) error

    // Cascade hook — host calls this when user is deleted.
    // Convention: <Action>By<ForeignType>(ctx, foreignID)
    CancelByUser(ctx context.Context, userID string) error
}
```

```go
// mod-billing/module.go (excerpts)

// Cascade hook — uses ctx-tx if set
func (m *Module) CancelByUser(ctx context.Context, userID string) error {
    _, err := m.clientForCtx(ctx).Subscription.
        Update().
        Where(
            subscription.UserIDEQ(userID),
            subscription.StatusEQ("active"),
        ).
        SetStatus("cancelled").
        SetCancelledAt(m.s.Now()).
        Save(ctx)
    return err
}

// Cross-module call — needs user data for invoice
func (m *Module) CreateInvoice(ctx context.Context, userID string, amount int64) (api.Invoice, error) {
    user, err := m.deps.IAM.Get(ctx, userID)         // ← cross-module call via api.Reader
    if err != nil { return api.Invoice{}, err }

    inv, err := m.clientForCtx(ctx).Invoice.Create().
        SetUserID(userID).
        SetUserEmailSnapshot(user.Email).            // denormalized for audit
        SetAmount(amount).
        Save(ctx)
    if err != nil { return api.Invoice{}, err }
    return toInvoiceAPI(inv), nil
}
```

### Layer 3 — host (wires modules + sidecar + lifecycle)

Host repo layout (in this boilerplate, but works for any path):

```
github.com/anyone/their-app/
├── go.mod                              ← imports modkit + mod-* modules
├── go.work                             ← (dev only) replace directives
├── dbent/                              ← HOST's own ent client for sidecar tables
│   ├── schema/
│   │   └── user_profile.go             ← sidecar — adds typed fields to User
│   └── gen/ent/
├── config/pkl/
│   ├── modules/                        ← symlinked by `task pkl:link`
│   │   ├── iam → vendor/github.com/khanakia/mod-iam/config/
│   │   ├── billing → vendor/.../mod-billing/config/
│   │   └── workspace → vendor/.../mod-workspace/config/
│   └── schema/Root.pkl                 ← imports module Config.pkl files
├── apidash/internal/graph/
│   └── schemas/user_ext.graphqls       ← extend type User with sidecar fields
├── internal/
│   ├── users/
│   │   └── service.go                  ← composition point: EnrichedUser
│   └── lifecycle/
│       └── user_deletion.go            ← cascade orchestrator
├── saas/pkg/host/
│   ├── services.go                     ← modkit.Services impl
│   └── tx.go
└── saas/cmd/api/main.go                ← wires modules
```

#### Host implements modkit.Services

```go
// host/saas/pkg/host/services.go
package host

import (
    "context"
    "database/sql"
    "database/sql/driver"
    "time"

    "github.com/khanakia/modkit"
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

type Services struct {
    db   *sql.DB
    log  *zap.Logger
    nc   *nats.Conn
}

func (s *Services) DB() driver.Driver { return s.db.Driver() }
func (s *Services) Log(ctx context.Context) *zap.Logger { return s.log /* + ctx fields */ }
func (s *Services) Now() time.Time { return time.Now() }

func (s *Services) Publish(ctx context.Context, subj string, payload []byte) error {
    return s.nc.Publish(subj, payload)
}

func (s *Services) WithTx(ctx context.Context, fn func(context.Context) error) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil { return err }

    if err := fn(modkit.WithTxContext(ctx, tx)); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

#### Host wires modules in main

```go
// host/saas/cmd/api/main.go
package main

import (
    "log"

    "github.com/khanakia/modkit"
    "github.com/khanakia/mod-iam"
    "github.com/khanakia/mod-billing"
    "github.com/khanakia/mod-workspace"

    "github.com/anyone/their-app/saas/pkg/host"
)

func main() {
    services := host.NewServices(/* db, nats, log */)

    // Wire in dependency order — host knows the topology
    iamMod := iam.New(loadIAMCfg())

    workspaceMod := workspace.New(loadWorkspaceCfg(), workspace.Deps{
        Core: services,
        IAM:  iamMod,                   // workspaceMod sees iamMod as iamapi.Reader
    })

    billingMod := billing.New(loadBillingCfg(), billing.Deps{
        Core: services,
        IAM:  iamMod,
    })

    mods := []modkit.Module{iamMod, workspaceMod, billingMod}
    for _, m := range mods {
        if err := m.Init(services); err != nil {
            log.Fatalf("init %s: %v", m.Name(), err)
        }
    }

    // ... start API, register resolvers, mount HTTP routes, etc.
}
```

#### Sidecar table — host owns it, mod-iam never sees it

```go
// host/dbent/schema/user_profile.go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type UserProfile struct{ ent.Schema }

func (UserProfile) Fields() []ent.Field {
    return []ent.Field{
        field.String("user_id").Unique(),     // soft FK to mod-iam User — no DB constraint
        field.Time("date_of_birth").Optional(),
        field.String("phone").Optional(),
        field.String("timezone").Default("UTC"),
    }
}
```

#### Composition service — EnrichedUser lives at host

```go
// host/internal/users/service.go
package users

import (
    "context"
    "time"

    iamapi "github.com/khanakia/mod-iam/api"

    "github.com/anyone/their-app/dbent/gen/ent"
    "github.com/anyone/their-app/dbent/gen/ent/userprofile"
)

type EnrichedUser struct {
    iamapi.User                       // embed mod-iam primitive
    DateOfBirth *time.Time
    Phone       string
    Timezone    string
}

type Service struct {
    iam     iamapi.Reader
    profile *ent.UserProfileClient
}

func (s *Service) Get(ctx context.Context, id string) (*EnrichedUser, error) {
    u, err := s.iam.Get(ctx, id)
    if err != nil { return nil, err }

    p, _ := s.profile.Query().Where(userprofile.UserIDEQ(id)).Only(ctx)
    eu := &EnrichedUser{User: u}
    if p != nil {
        eu.DateOfBirth = &p.DateOfBirth
        eu.Phone = p.Phone
        eu.Timezone = p.Timezone
    }
    return eu, nil
}

// Batch for N+1 mitigation
func (s *Service) GetMany(ctx context.Context, ids []string) (map[string]*EnrichedUser, error) {
    bases, err := s.iam.GetMany(ctx, ids)
    if err != nil { return nil, err }

    profiles, _ := s.profile.Query().Where(userprofile.UserIDIn(ids...)).All(ctx)
    pMap := make(map[string]*ent.UserProfile, len(profiles))
    for _, p := range profiles { pMap[p.UserID] = p }

    out := make(map[string]*EnrichedUser, len(ids))
    for id, u := range bases {
        eu := &EnrichedUser{User: u}
        if p, ok := pMap[id]; ok {
            eu.DateOfBirth = &p.DateOfBirth
            eu.Phone = p.Phone
            eu.Timezone = p.Timezone
        }
        out[id] = eu
    }
    return out, nil
}
```

#### Lifecycle orchestrator — the cascade in one tx

```go
// host/internal/lifecycle/user_deletion.go
package lifecycle

import (
    "context"

    "github.com/khanakia/modkit"
    iamapi "github.com/khanakia/mod-iam/api"
    billingapi "github.com/khanakia/mod-billing/api"
    workspaceapi "github.com/khanakia/mod-workspace/api"

    "github.com/anyone/their-app/dbent/gen/ent"
    "github.com/anyone/their-app/dbent/gen/ent/userprofile"
)

type UserDeleter struct {
    s         modkit.Services
    iam       iamapi.Writer
    billing   billingapi.Writer
    workspace workspaceapi.Writer
    profile   *ent.UserProfileClient
}

// Delete — atomic across all 4 writers via shared tx
func (d *UserDeleter) Delete(ctx context.Context, userID string) error {
    return d.s.WithTx(ctx, func(ctx context.Context) error {
        // Order matters — dependents first, owner last
        if err := d.workspace.DeleteByOwner(ctx, userID); err != nil { return err }
        if err := d.billing.CancelByUser(ctx, userID); err != nil { return err }

        // Sidecar cleanup uses host's own ent client + same tx
        if _, err := d.profile.Delete().Where(userprofile.UserIDEQ(userID)).Exec(ctx); err != nil {
            return err
        }

        return d.iam.Delete(ctx, userID)
    })
}
```

#### GraphQL extension — host adds fields to module's User type

```graphql
# host/apidash/internal/graph/schemas/user_ext.graphqls
extend type User {
    dateOfBirth: Time
    phone: String
    timezone: String
}
```

```go
// host/apidash/internal/graph/resolver/user.resolvers.go
// gqlgen generates these stubs because of the extend type above

func (r *userResolver) DateOfBirth(ctx context.Context, obj *iamapi.User) (*time.Time, error) {
    p, err := r.loaders.UserProfile.Load(ctx, obj.ID)   // DataLoader batches across query
    if err != nil { return nil, err }
    return &p.DateOfBirth, nil
}

func (r *userResolver) Phone(ctx context.Context, obj *iamapi.User) (*string, error) {
    p, err := r.loaders.UserProfile.Load(ctx, obj.ID)   // hits cached load
    if err != nil { return nil, err }
    return &p.Phone, nil
}
```

### The full request flow

#### Mutation: `deleteUser(id: "u_42")`

```
1. Client → POST /graphql {mutation deleteUser(id: "u_42")}
   ↓
2. Gin → mutation resolver `deleteUser(ctx, id)`
   ↓
3. Resolver → host.lifecycle.UserDeleter.Delete(ctx, "u_42")
   ↓
4. Delete → services.WithTx(ctx, fn)
   ↓ opens *sql.Tx, stores in ctx
5. Inside the tx (atomic):
   a. workspace.DeleteByOwner(ctx, "u_42")
      → mod-workspace.clientForCtx(ctx) returns tx-bound client
      → DELETE FROM workspaces WHERE owner_id = 'u_42'
   b. billing.CancelByUser(ctx, "u_42")
      → mod-billing.clientForCtx(ctx) — same tx
      → UPDATE subscriptions SET status='cancelled' WHERE user_id='u_42'
   c. profile.Delete()
      → host's own ent client — same tx
      → DELETE FROM user_profiles WHERE user_id='u_42'
   d. iam.Delete(ctx, "u_42")
      → mod-iam.clientForCtx(ctx) — same tx
      → DELETE FROM users WHERE id='u_42'
      → publishes "iam.user.deleted" via NATS
   ↓
6. WithTx → COMMIT (or ROLLBACK if any step erred)
   ↓
7. Async NATS subscribers (eventual consistency):
   - mod-audit-log: writes audit row
   - mod-search: removes user from search index
   These may fail without rolling back the deletion (acceptable side-effects)
   ↓
8. Resolver → returns success → client
```

#### Query: `query { user(id: "u_42") { id email dateOfBirth phone } }`

```
1. resolver.Query.User(ctx, "u_42")
   ↓
2. host.users.Service.Get(ctx, "u_42"):
   a. iam.Get(ctx, id)         → SELECT FROM users → returns iamapi.User
   b. profile.Query()...        → SELECT FROM user_profiles → enriches
   ↓
3. Resolver returns *iamapi.User to gqlgen
   ↓
4. gqlgen sees extra fields (dateOfBirth, phone) → invokes field resolvers:
   a. userResolver.DateOfBirth(ctx, obj)
      → loaders.UserProfile.Load(ctx, obj.ID) — DataLoader batches THIS request
        with any other UserProfile.Load in this query
      → ONE SQL: SELECT FROM user_profiles WHERE user_id IN (...)
      → returns DateOfBirth from cached load
   b. userResolver.Phone(ctx, obj) — reuses cached load result, NO new query
   ↓
5. gqlgen → assembles JSON → client
```

#### Cross-module call inside a module: `billing.CreateInvoice`

```
1. Caller → billingMod.CreateInvoice(ctx, userID, amount)
   ↓
2. Inside mod-billing impl:
   a. m.deps.IAM.Get(ctx, userID)  ← injected at construction in main.go
      → resolves to iamMod.Get (iamMod satisfies iamapi.Reader)
      → returns iamapi.User
   b. m.clientForCtx(ctx).Invoice.Create()...Save(ctx)
      → uses tx if present, else baseline client
   ↓
3. Returns api.Invoice
```

The dependency arrow at compile time: `mod-billing → mod-iam/api` (contracts only). At runtime: `billingMod.deps.IAM = iamMod` (host wires the concrete impl).

### Why each layer exists

| layer | owns | never-touches | why |
|---|---|---|---|
| modkit | Services + Module + tx helpers | domain types | bridge between modules and any host |
| module impl | ent schemas + own client + business logic | host paths or other modules' impl | self-contained portable unit |
| module api | DTOs + Reader/Writer interfaces | heavy deps (ent etc) | tiny stable cross-module contract |
| host | sidecars + composition + lifecycle + GraphQL extensions | module internals | application-specific glue + extension |

---

## Final Decisions

The locked-in summary. Every decision below is committed unless explicitly revisited.

| id | decision | status |
|---|---|---|
| ADR-V2-001 | Module + HostAPI pattern (overarching) | locked |
| ADR-V2-002 | Modules are separate Go modules with own repos | locked |
| ADR-V2-003 | modkit is the contract repo (interface-only ~300 LOC) | locked |
| ADR-V2-004 | Per-module api/ sub-package for cross-module contracts | locked |
| ADR-V2-005 | All modules are peers — no foundational tier | locked |
| ADR-V2-006 | Pattern A — each module owns its ent client; soft FK across | locked |
| ADR-V2-007 | PKL via task pkl:link symlink-stabilized paths | locked |
| ADR-V2-008 | Service-layer batch + DataLoader; aggregated client only as escape hatch | locked |
| ADR-V2-009 | modkit.Services.WithTx for cross-module transactions | locked |
| ADR-V2-010 | Sync-tx for hot cascades; events for reactive side effects | locked |
| ADR-V2-011 | Sidecar table + metadata jsonb for schema extension | locked |
| ADR-V2-012 | Layered composition — host stitches enriched DTOs | locked |

### What gets built

<Stats cols="3">
  <Stat title="Repos" value="N+1" description="modkit + mod-* per domain" />
  <Stat title="modkit LOC budget" value="<300" description="ifaces only" />
  <Stat title="Per-module dirs" value="6-8" description="api/, schema/, gen/, config/, ..." />
</Stats>

### Concrete repo set (initial)

- `github.com/khanakia/modkit` — contract repo
- `github.com/khanakia/mod-iam` — User, Tenant, Session, Auth
- `github.com/khanakia/mod-workspace` — Workspaces
- `github.com/khanakia/mod-billing` — Subscriptions, Invoices
- `github.com/khanakia/mod-rbac` — Roles, Permissions
- `github.com/khanakia/boilerplate-golang` — reference host (this repo, refactored)

Other modules added as needed. Pattern is fixed; domain choice is per-project.

### Required conventions per module

- [x] Lives at `github.com/khanakia/mod-<domain>` ^conv-repo-path
- [x] Has `api/` sub-package with DTOs + Reader + Writer interfaces ^conv-api-pkg
- [x] Reader includes `GetMany(ctx, ids) (map[string]T, error)` ^conv-getmany
- [x] Writer includes cascade hooks `*By<Foreign>(ctx, foreignID)` ^conv-cascade
- [x] Ships own ent schemas in `schema/` ^conv-schema
- [x] Ships own committed ent client in `gen/ent/` ^conv-entgen
- [x] Ships PKL config in `config/Config.pkl` + defaults + `//go:embed` ^conv-pkl
- [x] All Reader/Writer impl methods use `clientForCtx(ctx)` not raw `client` ^conv-clientforctx
- [x] Extensible entities ship `metadata jsonb` column + `Metadata` field on DTO ^conv-metadata

### Required conventions per host

- [x] Implements `modkit.Services` ^conv-host-services
- [x] Calls `task pkl:link` before `task config:gen` ^conv-host-pkl
- [x] Owns `lifecycle/` package for cascade orchestration ^conv-host-lifecycle
- [x] Owns single composition service per extensible entity (e.g. `users.Service`) ^conv-host-compose
- [x] GraphQL extensions in host's own `.graphqls` files; field resolvers use DataLoader ^conv-host-graphql
- [x] Sidecar tables in host's own `dbent/schema/`; soft FK only ^conv-host-sidecar

---

## Open questions

<Details summary="modkit version-stability strategy when adding services?">

When `Services` needs a new method (e.g. `FeatureFlags()`, `TenantContext()`), every consumer breaks. Options:

1. Bump major version — every module re-releases against `modkit/v2`.
2. Add `ServicesV2 interface { Services; NewMethod() }` and let modules opt-in via type assertion.
3. Bake all foreseeable services in Day 1.

Lean toward option 2 — additive, no force-rebuild. Worth deciding before first publish.

</Details>

<Details summary="Migration coordination across modules?">

Each module owns its migrations. Host runs them in dependency order. Problem: simultaneous module bumps (e.g. mod-iam@v1.5.0 and mod-billing@v2.0.0) each shipping migrations may have ordering constraints. Need a migration manifest per module that declares "requires mod-iam >= v1.4.0" or similar.

Defer: handle when first cross-module migration constraint shows up.

</Details>

<Details summary="Module discovery — explicit list or auto-register via init()?">

Host can either (a) explicitly list modules in `main()` or (b) modules self-register via `init()` on import.

Lean toward explicit (a) — debuggable, no init-order surprises, no hidden side effects. Cost: one line per module in host main. Worth it.

</Details>

<Details summary="Host-level aggregated ent client — when?">

Decision 8 says build only when a real query forces the issue. Defer until the first cross-cutting admin report needs SQL JOIN. Track as a "watch for" item; don't preempt.

</Details>

---

## Risks

<Callout type="danger" title="modkit drift">

Every new cross-cutting concern tempts a new method on `Services`. At 20+ methods it's the Plugin god-struct again. **Rule:** new method requires ADR + version-bump strategy decision (open question above).

</Callout>

<Callout type="warning" title="Per-module api/ stability">

Once published, adding methods to `Reader`/`Writer` breaks consumers. Use semver. Add new ifaces (`ReaderV2`) instead of mutating. Plan domain methods carefully on first publish.

</Callout>

<Callout type="warning" title="Tx-aware client wrapping is easy to forget">

Every Reader/Writer impl method must call `clientForCtx(ctx)` not `m.client`. Easy to slip. Add a linter rule or codegen template.

</Callout>

<Callout type="warning" title="Cascade scope creep">

Long-running tx across many modules = lock contention. Keep cascades to ≤3 modules; fan-out beyond that goes async via events.

</Callout>

---

## Tasks ahead

- [ ] Lock open questions (modkit versioning strategy, module discovery) !p0 ~2026-04-30 &4h ^task-v2-lock-opens
- [ ] Publish skeleton modkit repo with ifaces + helpers !p0 ~2026-05-04 &1d after:task-v2-lock-opens ^task-v2-modkit-skeleton
- [ ] Pilot module: convert one feature (candidate: iam) to mod-iam !p0 ~2026-05-15 &3d after:task-v2-modkit-skeleton ^task-v2-pilot-iam
- [ ] Implement task pkl:link in host !p1 ~2026-05-16 &4h after:task-v2-pilot-iam ^task-v2-pkl-link
- [ ] Linter rule for clientForCtx usage !p2 ~2026-05-22 &4h ^task-v2-lint-clientforctx
- [ ] Document module-author guide in modkit README !p1 ~2026-05-23 &4h after:task-v2-pilot-iam ^task-v2-author-guide

<Kanban md group-by="status" title="Mission V2 — decision-execution board"></Kanban>

---

## Glossary

<Details summary="Terms used in this doc">

- **modkit** — interface-only contract repo; both module and host import it.
- **Module** — a feature published as its own Go module (`mod-iam`, `mod-billing`, …).
- **Host** — the consuming app. Any Go module path. Implements `modkit.Services`.
- **api/** — sub-package inside each module containing DTOs + Reader/Writer interfaces.
- **Sidecar** — host-owned table/entity that extends a module entity with extra typed fields.
- **Soft FK** — string ID column referencing another module's entity. No DB-level FK constraint.
- **DataLoader** — per-request batch+cache mechanism for GraphQL field resolvers (mitigates N+1).
- **Cascade hook** — `*By<Foreign>` method on `api.Writer` (e.g. `CancelByUser`) for cross-module deletion orchestration.
- **clientForCtx** — module-internal helper that returns a tx-bound or baseline ent client based on ctx.
- **Aggregated ent client** — optional host-level ent client over all module schemas. Used only for cross-cutting admin queries. Module business code never touches it.

</Details>

---

## References

- `.ai/plans/MISSION_V2.md` — mission overview + phased rollout
- `.ai/ARCHITECTURE.md` — current architecture (pre-V2)
- `.ai/STACK.md` — tech stack
- `.ai/PATTERNS.md` — code conventions index

---

## Decision 13 — AST scaffolding as customization mechanism

After exhaustive brainstorm (11 pluggability approaches, 3 real-Go-code
patterns, ent mixins), the only mechanism that solves typed-field
extension without text templates is **AST-based code transformation**.

<ADR status="accepted" id="ADR-V2-013" date="2026-04-26" title="AST scaffolding via mod-cli + dave/dst is the V2 customization mechanism">

### Context

Consumers need to add typed fields to entities the module ships with
(e.g. add `phone` to mod-auth's User with a unique constraint). Earlier
options fell short:

- Behavior interfaces customize methods, not struct fields.
- Generics let consumer redeclare types, not extend module's types.
- Sidecar tables put fields in a separate table — not the same row.
- Text-template scaffolding (.tmpl files) is hard to debug/test.
- ent mixins help but require consumer to write whole User schema themselves.

The user's actual model: a consumable module that exists as REAL working
Go, ships, can be debugged. Consumer declares extensions; a Go program
reads the consumable's source, mutates AST (adds fields, swaps methods),
writes real Go to `modules/mod-auth/`. Output is also real Go — debuggable,
testable, owned by consumer.

### Decision

Adopt AST scaffolding as the V2 customization mechanism:

- **Tool:** `github.com/khanakia/mod-cli` (Go program using
  `github.com/dave/dst` for AST parsing/printing with comment + format
  preservation).
- **Config:** `.mod-cli.yaml` — declarative extensions per file
  (struct field inject, ent.Fields() append, method body replace,
  import management).
- **Source:** consumable modules ship as ordinary Go modules
  (e.g. `mod-auth-template`) — they compile, test, and run standalone.
- **Output:** real Go in consumer's `modules/mod-auth/`. Consumer commits,
  owns, evolves freely.
- **Build status:** **deferred.** Existing V2 modules require zero
  changes to be AST-ready. Build mod-cli when consumer demand triggers
  (2+ consumers needing diverging schemas, OR routine module copy-paste).

### Consequences

- **Win:** typed fields on entity tables with real DB constraints, real
  Go end-to-end, no template engine to debug, fits Go ecosystem norms
  (wire/mockgen/protoc-gen-go all use AST).
- **Win:** consumer takes ownership of scaffolded code; can edit further
  with full IDE support.
- **Win:** zero runtime overhead (it's all compile-time).
- **Cost:** re-scaffold (applying upstream template changes to already-
  customized output) is hard problem — defer to Phase 3 of mod-cli build.
- **Cost:** mod-cli is ~770 LOC of Go to build (1-2 days focused work).
- **Cost:** module authors must keep AST-friendly conventions (predictable
  function shapes, no exotic dynamic constructions). Light hygiene.
- **Accept:** existing V2 modules stay as-is. AST is purely additive
  tooling whenever you need it.

</ADR>

See `MISSION_V2_AST_PLAN.md` for the detailed plan, examples, and roadmap.

