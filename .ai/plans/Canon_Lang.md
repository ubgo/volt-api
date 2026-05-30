---
title: Canon — proposed programming language for modular monolith SaaS
status: thought-experiment
owner: khanakia
created: 2026-04-26
defaults:
  project: mission-v2
  area: work/architecture
  owner: khanakia
---

# Canon — Programming Language for Pluggable SaaS

A thought experiment: if we built our own language to solve the
pluggability + extensibility + type-safety problem cleanly, what would
it look like?

<Callout type="warning" title="Status">

This is a **design sketch**, not a buildable spec. The point is to
clarify what the ideal solution looks like, why no existing language
nails it, and what tradeoffs a custom language would still have to
accept. Not a recommendation to actually build it.

</Callout>

## Why a new language

Existing approaches (Go modkit + AST, Convex Components, Rails Engines,
Encore.dev) each sacrifice something fundamental:

| Stack | Sacrifices |
|---|---|
| Go + V2 modular | cross-module SQL JOINs, ent edges, DB-level FKs, schema customization without AST tool |
| Convex | language (TS only), self-host (cloud-mostly), DB (NoSQL), vendor lock |
| Rails Engines | static type safety, performance, compile-time correctness |
| Encore.dev | opinionation, cloud lock-in (mostly), Go ecosystem mismatch |
| Prisma + TS | runtime safety only, schema customization mechanism limited |

Canon's design goal: keep type safety + self-hosted + Postgres + cross-module
joins + per-consumer customization, all simultaneously. The math says this
needs language-level features that no existing mainstream language provides
together.

## Core design principles

| Principle | Effect |
|---|---|
| Schema is first-class type system | No codegen step; types ARE schemas |
| Row-polymorphic extensible records | Add fields to types defined elsewhere, compile-checked |
| Capability traits with config | "User has Phone with unique=true" |
| First-class cross-module references | Real FK + JOINs across module boundaries |
| Effects tracked in types | DB, time, network all visible in signatures |
| Postgres as runtime | Self-hostable, real SQL semantics |
| Reactive queries native | UI auto-subscribes — no manual cache wiring |
| Migrations derived from schema diff | Single source: schema; migrations follow |
| No glue code generation | Compiler emits SQL, GraphQL, client SDKs as part of build |

## Language tour

### Entities and extension

```canon
// mod-auth/user.canon
module mod-auth

entity User {
    id: ID
    email: Email unique
    password: Sensitive<String>
    is_active: Bool = true
    created_at: Time = now()
}
```

A consumer adds fields without templates, scaffolders, or AST tools:

```canon
// my-app/extensions.canon
extend mod-auth.User {
    phone: Phone unique optional
    timezone: Timezone = "UTC"
    department: String
}
```

The compiler merges all `extend` clauses into the final `User` type.
Unique constraints are preserved. Migrations are auto-derived. Type-safe
everywhere; renaming a field surfaces a compile error in every reference.

### Cross-module relationships

```canon
// mod-workspace/workspace.canon
module mod-workspace
import mod-auth.User

entity Workspace {
    id: ID
    name: String
    slug: String unique
    owner: ref User                          // typed cross-module reference → real FK
    members: [ref User] via WorkspaceMember  // typed many-to-many through join
}

entity WorkspaceMember {
    workspace: ref Workspace
    user: ref User
    role: Enum["owner", "admin", "member", "viewer"]
}
```

`ref User` becomes a foreign key at the database level. The compiler
errors if `User` does not exist or the import path is wrong. Queries
JOIN natively:

```canon
let stats = query {
    User
    .filter(.is_active)
    .with(.workspaces)                   // typed JOIN; column resolved by compiler
    .select(.email, count(.workspaces) as ws_count)
}
```

The compiler emits efficient SQL. There is no `WithWorkspaces()` codegen
step. Renaming a column ripples through every query at compile time —
no runtime drift.

### Capabilities — typed extension primitives

```canon
// mod-auth-phone/capability.canon
module mod-auth-phone

capability Phone {
    fields {
        phone: String
        verified_at: Time?
    }

    config {
        unique: Bool = true
        verification_method: Enum["sms", "email"] = "sms"
    }

    fn verify(code: String) -> Result<()> ![DB, Time, SMS]
    fn send_verification() -> Result<()> ![DB, Time, SMS]
}
```

A consumer enables the capability with config:

```canon
// my-app/main.canon
extend mod-auth.User has Phone {
    unique = true
    verification_method = "sms"
}
```

At compile time, the User entity gains the `phone` and `verified_at`
fields with proper constraints; `user.verify("123456")` becomes
callable. No glue code; no separate sidecar table; no `requirePhoneOnSignup`
config in mod-auth — capability declarations are the contract.

### Effects in signatures

```canon
fn create_workspace(owner: User, name: String) -> Workspace
    ![DB, Audit, Email, Time]
{
    let ws = Workspace.create({ name, owner })
    Audit.log("workspace.created", ws.id)
    Email.send(owner.email, "Welcome to {ws.name}")
    return ws
}
```

The `![...]` annotation tracks side effects in the type signature:

- `[DB]` reads or writes the database
- `[Audit]` emits audit events
- `[Email]` sends email
- `[Time]` reads the wall clock

Tests inject fakes:

```canon
test "create_workspace sends welcome email" {
    let fake_email = mock(Email)
    let ws = create_workspace(alice, "Alice's space")
        with [Email = fake_email]

    assert fake_email.sent.first.subject matches "Welcome*"
}
```

Effects are mockable, traceable, and refuse to compile when the caller
does not declare them. This replaces dependency injection frameworks
(wire, fx) and explicit interface plumbing (modkit.Services).

### Module distribution

```toml
# my-app/canon.toml
[modules]
mod-auth        = { source = "github.com/khanakia/mod-auth", version = "1.0.0" }
mod-auth-phone  = { source = "github.com/khanakia/mod-auth-phone", version = "0.5.0" }
mod-workspace   = { path = "./modules/workspace" }   # local

[extensions]
"mod-auth.User".phone                       = { capability = "Phone", unique = true }
"mod-auth.User".phone.verification_method   = "sms"
"mod-auth.User".department                  = "String"
"mod-auth.User".timezone                    = { type = "Timezone", default = "UTC" }
```

The compiler resolves all modules at build time. The final binary
contains the merged schema, generated migrations, a typed query API, and
a GraphQL endpoint emitted from the same source.

### Reactive queries

```canon
view UserDashboard(user_id: ID) {
    let user       = User.get(user_id)                                 reactive
    let workspaces = user.workspaces.filter(.is_active)                reactive
    let recent     = workspaces.sortBy(.updated_at).take(5)            reactive

    return <Page>
        <h1>{user.email}</h1>
        <WorkspaceList items={recent} />
    </Page>
}
```

Reactive queries auto-subscribe. When the data changes, the UI
re-renders. Like Convex, but compile-checked and self-hostable.

### Schema migrations

```bash
$ canon migrate plan
> add column users.phone (string, unique, optional)
> add column users.timezone (string, default "UTC")
> add column users.department (string)
> add table workspaces
> add foreign key workspaces.owner_id → users.id
> add table workspace_members

$ canon migrate apply
✓ applied 6 changes (3.2s)
```

Migrations derive from the diff between current and previous compiled
schema. No `task entg`; no Atlas; no migration-file babysitting.

## How Canon solves V2's specific problems

| V2 problem | Canon answer |
|---|---|
| Cross-module ent edges impossible | First-class `ref` syntax at language level |
| Schema customization needs an AST tool | Native `extend` keyword merges at compile time |
| Cross-module SQL JOINs lost | Compiler emits JOINs from `ref` declarations |
| Sidecar tables required for typed extension | `extend` adds real columns to the same table |
| Cascade orchestration manual + fragile | Compiler derives from `ref` with `on_delete: cascade` |
| Codegen drift risk (ent, sqlc) | No codegen — compiler IS the system |
| Type drift between modules | Single compile pass validates all module references |
| jsonb metadata as escape hatch | Untyped fields are outlawed; `extend` is type-safe |
| modkit interface boilerplate | Effects + capabilities replace explicit interfaces |
| Cross-module testing harness | Effects mocking is built into the language |
| Real-time UI is manual (NATS/SSE) | Reactive queries are free |
| Schema versioning across consumers | Module versions + extensions in canon.toml |
| Per-consumer "phone unique or not" | `extend` config carries the variation |

## Comparison to existing languages

Each took a different bet. Canon would steal one feature from each.

| Language | What Canon would borrow |
|---|---|
| Elm | Row polymorphism — extensible records |
| PureScript | Row polymorphism + effects |
| OCaml objects | Structural typing |
| Effect-TS | Algebraic effects pattern |
| Convex | Reactive queries + schema-as-code |
| Hasura's Action DL | Schema-first cross-module relationships |
| Prisma SDL | Declarative schema → DB + types + migrations |
| GraphQL SDL | Declarative schema, distributed |
| Datomic / XTDB | First-class relationships |
| Unison | Content-addressed code; types as ABI |
| Rust | Ownership / borrow checker for safe concurrency |
| Roc | Effects + ergonomic syntax |

Canon mostly resembles Prisma's schema language extended into a full
programming language with runtime semantics. Closest historical relative
might be Datomic + a typed lambda calculus on top.

## Implementation effort (rough estimate)

| Layer | Effort |
|---|---|
| Parser + AST | 2-3 months |
| Type checker (with row polymorphism) | 4-6 months |
| Effect system | 3-4 months |
| Postgres backend (schema diff, migrations, query compiler) | 6 months |
| Reactive runtime | 6 months |
| Module resolver + distribution | 2 months |
| Standard library | 6 months |
| Tooling (LSP, debugger, formatter) | 6 months |
| Ecosystem (drivers, integrations, libs) | infinite |

Realistic single-team year-long project to reach "usable for one app."
Multi-year for "production-grade with ecosystem."

## What Canon would still cost

| Cost | Why |
|---|---|
| Learning curve | new language, new mental model |
| Zero ecosystem at start | no Postgres extensions wrapped, no AWS/Stripe SDKs |
| Self-hosting is on you | compiler + runtime + Postgres orchestration to bundle |
| Migration from existing systems | porting ent schemas is mechanical but real |
| FFI to Go / JS for tools | extra interop layer per integration |
| Hiring | nobody knows it; you teach everyone |
| Documentation burden | every feature needs guides + examples |
| Compiler bugs | one team's bug surface is your problem to fix |

## Open design questions

<Details summary="How does row polymorphism + DB schema interact at runtime?">

Row polymorphism normally requires that the type of a value carries its
row signature. For DB-persisted entities, the row is fixed at module load
(after all `extend` clauses merge). So the polymorphism is purely
compile-time; runtime values have a fixed shape.

This works as long as no module dynamically loads at runtime. If hot-load
becomes a goal, schema migration plus type-system reconciliation becomes
much harder.

</Details>

<Details summary="What happens when two modules extend the same field with conflicting types?">

Compile error. Canon's module resolver detects:

- `extend User { phone: String }` from module A
- `extend User { phone: Int }` from module B

→ "Conflicting extension of `mod-auth.User.phone` from modules A and B."
Build fails. User must resolve by picking one or aliasing.

</Details>

<Details summary="How are migrations safe across distributed deployments?">

Each compiled binary contains a schema hash. Pre-deploy check ensures
the new binary's expected schema matches the database's current
migration version. If not, deploy aborts with "run `canon migrate apply`
first."

For zero-downtime deploys: schema changes that are backward-compatible
(add nullable column, add table) are safe; breaking changes (drop
column, change type) require explicit two-phase rollout (compatibility
mode → new schema → old binaries removed).

</Details>

<Details summary="How does Canon handle multi-tenancy?">

Tenant-scoping is a built-in modifier on entities:

```canon
entity User {
    @scope(tenant)
    ...
}
```

Compiler ensures every query filters by the current tenant. Effects
include `[TenantCtx]`. Cross-tenant access requires explicit privilege.

</Details>

<Details summary="GraphQL / REST endpoint generation?">

Compiler emits GraphQL schemas and (optionally) REST routes from entity
definitions plus annotations:

```canon
@expose(graphql, query, mutation)
@auth(role = "admin", for = ["delete"])
entity User { ... }
```

GraphQL resolvers are auto-generated from `query` expressions. The
GraphQL endpoint is a build artifact, not separate code.

</Details>

<Details summary="What about extension by composition vs inheritance?">

Canon avoids inheritance entirely. All extension is via `extend` (row
augmentation) or `has Capability` (trait-style composition). No subclass
dispatch; no virtual methods. Functions are values; behavior is wired by
passing them.

</Details>

## What's not in Canon (deliberate omissions)

| Omitted | Why |
|---|---|
| Inheritance | Composition (capabilities + extend) covers all cases |
| Untyped fields / any-type | Defeats the purpose; ban at language level |
| Manual codegen step | Compiler IS the codegen |
| Multiple ORMs / schema langs | Schema is built in; only one truth |
| Imperative migrations | Diff-derived only; no manual migration scripting |
| Runtime hot-loading of new modules | Out of scope; requires schema-version negotiation |
| Cross-language interop without explicit FFI | Stays focused on backend Postgres apps |
| NoSQL / graph DB backends | v1 is Postgres-only; broaden later if it works |

## Realistic verdict

| Question | Answer |
|---|---|
| Could a language solve V2's pluggability cleanly? | Yes, with row polymorphism + capabilities + first-class refs |
| Has anyone built it for backends? | Closest: Convex, Hasura, Prisma — each missing pieces |
| Worth building Canon? | Only as a multi-year research project with ecosystem ambition |
| Closest existing thing for your needs? | Convex (TS, cloud-locked) or stay Go + V2 tradeoffs |
| If you started Canon as a research prototype? | 6 months for type checker + Postgres compiler; don't quit Go |

The reason no language nails this is not lack of imagination. **Language
design is hard; ecosystem matters more than features.** Convex got
closest because it bet on TypeScript (massive ecosystem leverage) plus
cloud (controlled runtime). Self-hosted, multi-language, full type-safe
modular monolith with extension is genuinely an open problem.

## What this thought experiment teaches

| Insight | Implication |
|---|---|
| The Go limitations are language-shaped, not engineering-shaped | V2's tradeoffs are real, not workaround-able |
| Row polymorphism + capabilities would solve it elegantly | Languages exploring this (Roc, Effect-TS, Unison) are interesting to watch |
| Reactive + typed + self-hosted is open territory | Whoever builds it well captures a niche |
| AST scaffolding (V2's path) is a pragmatic stand-in | Real fix requires language-level features Go doesn't have |
| Convex demonstrates user demand for the shape | Market for "Canon-like" exists, just not in Go |

## Cross-references

- `MISSION_V2_DECISIONS.md` — locked V2 architecture under Go's constraints
- `MISSION_V2_AST_PLAN.md` — pragmatic Go answer to extension problem
- `Golang_Pluggability.md` (private-repo) — full survey of approaches
- `MISSION_V2_LOG.md` — discussion threads that produced this doc
