# Code standards

**Read before:** writing or editing any code. This is the standing engineering bar for every change in Volt API — applied by default, not re-stated per task. The four pillars are non-negotiable; §3 has the concrete Go rules for this repo.

## 1 · The four pillars — apply to EVERY change

### 1.1 Full type safety — no escape hatches

- Zero escape hatches. BANNED: naked `interface{}` / `any`, empty-interface maps used to dodge a real type, type assertions without the `, ok` form when the value can legitimately be another type. There is no "just this once" — if you reach for one, stop and model the type.
- Unknown shapes cross into the type system only at a **system boundary** (HTTP body, JSON, env, external API) and MUST be validated there (valgo / explicit decode + guard) into a concrete typed result. The untyped value never leaks past that line.
- Prefer the type system to runtime checks: discriminated types, branded/opaque ID types where ids are confusable, and **exhaustive `switch` with a `default:` panic arm** so adding an enum case is caught.
- The compiler IS the test for type safety. Not done until `go build ./...` + `go vet ./...` pass clean.

### 1.2 Zero hardcoded strings (and magic numbers) for closed-set values

- Any closed-set value — status, kind, role, scope, event subject, cache-key prefix, GraphQL field name, config key — gets a named constant, defined once, referenced everywhere. Never a bare literal at the use site.
- Keep a canonical iteration list beside the constants (a `…Values` slice / `var All… = []T{…}`) so seeds, resolvers, and aggregations read from one source — read and write paths never diverge.
- The test: grep your diff for string literals compared with `==` or used as `case`/map keys. Each is a named constant or a documented exception (free-form user text, or an upstream wire field name kept verbatim — translate INTO your constant, never rename the fixture).
- Magic numbers get the same treatment — a named constant, never `timeout = 30_000` inline.

### 1.3 Context-dense comments — so a fresh session loses nothing

Any new AI session or human opening this code must understand *what it is, why it exists, and what not to break* without re-deriving it.

- **Every exported func / type / const gets a doc comment** stating what it does, *why it exists* (the non-obvious reason), and any invariant or gotcha a caller must respect. `// GetUser gets the user` adds nothing — comment intent and constraints.
- **Document multi-shape fields at the declaration.** If a column / payload / `any` field can hold more than one shape depending on the writer, enumerate every shape in the comment right there — no sidecar docs.
- **Document non-obvious framework rules where used** (e.g. "Ent field immutable but settable on Create" goes in the schema comment; "NATS handler registered in `WithAfterNatsInitiated`, not at import time" beside the registration).
- **Explain the why** for any workaround, ordering dependency, or "looks wrong but is intentional" block — else a future editor "fixes" it and reintroduces the bug.
- Don't comment the obvious. Target: a reader who's never seen the file can answer "why does this exist and what breaks if I change it?" from the comments alone.

```go
// ResolveTenant maps an inbound request to its owning workspace.
//
// Why it exists: tenancy is derived from the API-key prefix, NOT a header —
// headers are spoofable and are never trusted for authorization.
//
// Invariant: returns ErrNoTenant (never nil,nil) for anon-scoped keys —
// callers MUST branch on that rather than assume a workspace is present.
```

### 1.4 Modular, importable, no AI slop

- Organize so a package can be lifted into another project with minimal surgery: clear boundaries, no hidden coupling to app-global singletons, dependencies injected (constructor args) not reached-for via a global `*volt.Kernel` inside methods.
- One responsibility per package/file, named for what it owns. No generic `util`/`helpers`/`common` grab-bags, no 800-line files.
- Intentional public surface: export what callers need, keep the rest unexported.
- "AI slop" smells to avoid: dead scaffolding, copy-pasted near-duplicate blocks instead of a shared helper, generic names (`data`, `handleStuff`, `Manager`), commented-out code left in (unless tagged `@tag('dnd')`), and stubs that silently return empty — a data-returning stub must return an error or log a warning, never look like a real run on empty input.
- Reuse before rebuild: if a `lace/`, `core/`, or `webapp/pkg/` helper already exists, use it — don't hand-roll a local duplicate.

## 2 · Checklists

### 2.1 Pre-flight (before writing code)

- [ ] **Shared resources** — grep every reader/writer of any field/column/table/env var/GraphQL type I'll touch; list them `file:line`; confirm compatibility.
- [ ] **Constants** — every closed-set value has a named constant + canonical iteration list. No bare literals.
- [ ] **Types** — every shape expressed; no untyped escape hatch planned; opaque IDs where confusable.
- [ ] **Reuse** — does a `lace/` / `webapp/pkg/` helper already exist? Extend it.
- [ ] **Empty / fallback path** — the zero/nil/missing case for every input has a defined graceful degradation (not a crash, not a silent zero).
- [ ] **Module boundary** — where does this live so it's importable? Dependencies injected, not reached-for.

### 2.2 Post-flight (before saying "done")

- [ ] **Build passes** — `go build ./...` clean (type safety verified, not assumed).
- [ ] **Vet + lint** — `task lint` (gofmt + go vet) clean.
- [ ] **Codegen current** — if a schema changed, the generator ran (`task dbent:g` / `task apidash:gql` / `task config:gen`) and generated code compiles.
- [ ] **Comments** — every new exported symbol has a why-and-invariant doc comment; multi-shape fields documented at the declaration; every workaround explains itself.
- [ ] **Hardcoded-value grep** — diff has no bare closed-set literals and no magic numbers.
- [ ] **Run twice / inverse / lossy-diff** — idempotent on re-run; the opposite direction tested; every source field reached the destination or is documented as dropped.
- [ ] **Blast radius** — `git diff`; for each removed/renamed symbol, grep for callers.

## 3 · Go stack rules for this repo

The universal floor above, plus the concrete rules for Volt API. When these and a pillar disagree on a specific, follow the pillar's intent.

- **Never edit generated code.** Anything under `gen/`, `generated/`, `ent/`, or a `*_gen.go` file with a `Code generated` header is output. Edit the source and regenerate: Ent (`modules/dbent/schema/*.go` → `task dbent:g`), gqlgen (`*.graphqls` → `task apidash:gql`), PKL bindings (`modules/config/pkl/**` → `task config:gen`). No sqlc/mockgen/wire in this repo.
- **Config is type-safe PKL, never raw `os.Getenv`** in app code. Access through the generated bindings; secrets come from the environment via PKL's `read("env:VAR")`. Environments selected by `CONFIG_ENV` (`default` / `prod` / `sample`).
- **Signatures:** context first, error last — `func Do(ctx context.Context, …) (*Result, error)`. Wrap errors with `fmt.Errorf("op: %w", err)`. Never `panic` in libraries; only `main`/`init` may.
- **Logging** goes through `core/pkg/applog` (wraps `github.com/ubgo/logger`, aliased `ulog`). Use `applog.FromCtx(ctx)` for trace-correlated logs; fields are `ulog.String` / `ulog.Err` / `ulog.Any` (there is no `zap`). Log at boundaries, not every function; don't log AND return the same error.
- **Tracing** via OpenTelemetry (`github.com/ubgo/otelkit` bootstrap); span names `<component>.<operation>`. **Errors** reported through `github.com/ubgo/errx`.
- **Database** access via `webapp.New()` → the Ent client; never `ent.Open()` / `sql.Open()` directly. Migrations run with `WithForeignKeys(false)`. Sensitive fields use `.Sensitive()`.
- **GraphQL** is schema-first: edit `.graphqls`, regenerate, implement resolvers. Business logic lives in `modules/webapp/pkg/`, not resolvers. Never add helper files in `resolver/` (gqlgen overwrites). File uploads are base64 `String!`, never `graphql.Upload`.
- **Packages** modular under their boundaries; the `/demo`-equivalent scratch (`modules/examples`, `modules/webapp/cmd/test`) is quarantined and deletable.
- **Post-flight build IS the test:** `go build ./...`, `go vet ./...` (`task lint`), and the repo's tests pass clean.

See the per-action guides in this folder (endpoints, schemas, migrations, NATS, secrets, …) and `../docs/BACKEND_PATTERNS.md` for the condensed cheat-sheet. Commits follow [git-commit.md](git-commit.md).
