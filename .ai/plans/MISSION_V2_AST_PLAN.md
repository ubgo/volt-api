---
title: Mission V2 — AST Scaffolding Plan (locked)
status: locked-deferred
owner: khanakia
created: 2026-04-26
defaults:
  project: mission-v2
  area: work/architecture
  owner: khanakia
---

# Mission V2 — AST Scaffolding Plan

The chosen extension mechanism for "consumer needs typed fields on User
that the module didn't ship." Locked decision; deferred build.

<Callout type="info" title="Status: locked architecturally — deferred operationally">

The architecture is decided. The tool is **NOT** built yet, and that's
intentional — current V2 modules work as standalone code. AST tooling is
purely additive, can be built any time. **Build trigger:** 2+ consumers
need diverging schemas, OR copy-pasting modules between projects becomes
routine.

Until trigger fires: keep building features on existing V2 modules.

</Callout>

## Why AST won

After exhaustive brainstorm (see `Golang_Pluggability.md` for all 11
approaches), AST scaffolding is the only mechanism that satisfies:

| requirement | satisfied |
|---|---|
| typed fields on User table (real columns, real DB constraints) | satisfied |
| no text templates / .tmpl files (debugging hell) | satisfied |
| source ships as REAL working Go (compiles standalone, tests pass) | satisfied |
| output is REAL Go in consumer's repo (debuggable, testable) | satisfied |
| each consumer can have different field shapes | satisfied |
| upstream bug fixes flow when consumer re-scaffolds | partial (re-scaffold is hard problem) |
| zero infrastructure (no codegen at runtime) | satisfied |
| fits Go ecosystem norms (wire/mockgen/protoc-gen-go all do this) | satisfied |

**Approaches considered and why they fell short for THIS specific need:**

| approach | why-not-primary |
|---|---|
| Approach 1 — independent modules with own ent (V2 today) | no per-consumer schema variation |
| Approach 2 — capability modules | proliferates separate tables; can't merge into single User row |
| Approach 3 — centralized schema (short-name dbent) | no cross-project distribution |
| Approach 4 — text-template scaffold | debug/test pain on .tmpl files |
| Approach 5 — hybrid scaffold + core | still uses templates at consumer boundary |
| Behavior interface + DefaultBehavior | customizes methods only — not struct fields |
| Generics over entity type | consumer redeclares User entirely — no extension story |
| Middleware chains | wraps operations — no schema impact |
| ent mixins (used internally today) | requires consumer to write whole User schema themselves |
| **AST scaffolding** | **solves typed-field-on-User-table without templates** |

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ github.com/khanakia/mod-auth-template                           │
│                                                                 │
│ REAL working Go module:                                         │
│   ├── go.mod                                                    │
│   ├── api/types.go        (User struct — minimal core fields)   │
│   ├── schema/user.go      (ent schema — Fields() returns slice) │
│   ├── module.go           (Behavior + Module impl)              │
│   ├── api/reader.go       (Reader iface)                        │
│   ├── api/writer.go       (Writer iface)                        │
│   └── ... full module that compiles + tests pass standalone     │
└─────────────────────────────────────────────────────────────────┘
                            │
                            │ consumer runs: mod-cli scaffold ...
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ github.com/khanakia/mod-cli                                     │
│                                                                 │
│ Go program using github.com/dave/dst (Decorated Syntax Tree):   │
│   ├── fetches consumable module (git clone --depth=1)           │
│   ├── parses each .go file with dst.Decorator                   │
│   ├── walks AST, finds named structs/methods                    │
│   ├── injects field declarations / replaces method bodies       │
│   ├── writes mutated Go to consumer's target dir                │
│   └── runs post-hooks (go run ./cmd/entg, go mod tidy)          │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ consumer's repo at modules/mod-auth/                            │
│                                                                 │
│ REAL Go with consumer's customizations baked in:                │
│   ├── api/types.go        (User now has phone + department)     │
│   ├── schema/user.go      (Fields() now declares phone unique)  │
│   ├── module.go           (Behavior maybe uses argon2)          │
│   ├── gen/ent/            (entc output — real fields)           │
│   └── consumer commits, owns, evolves freely                    │
└─────────────────────────────────────────────────────────────────┘
```

## Concrete examples

### Example 1 — Inject fields into ent schema

**Source — consumable module's `schema/user.go` (real Go, compiles):**

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique(),
        field.String("password").Sensitive(),
    }
}
```

**Consumer's `.mod-cli.yaml`:**

```yaml
extensions:
  schema/user.go:
    User.Fields:
      - { kind: string, name: phone, opts: [unique, optional] }
      - { kind: string, name: department }
      - { kind: time, name: deleted_at, opts: [optional, nillable] }
```

**Tool runs (pseudocode):**

```go
import "github.com/dave/dst"
import "github.com/dave/dst/decorator"

func ExtendEntFields(path, structName string, newFields []FieldSpec) error {
    f, err := decorator.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
    if err != nil { return err }

    for _, decl := range f.Decls {
        fn, ok := decl.(*dst.FuncDecl)
        if !ok || fn.Name.Name != "Fields"  { continue }
        if !receiverIs(fn, structName)       { continue }

        // find the slice literal in `return []ent.Field{...}`
        for _, stmt := range fn.Body.List {
            ret, ok := stmt.(*dst.ReturnStmt)
            if !ok { continue }
            slice := ret.Results[0].(*dst.CompositeLit)
            for _, nf := range newFields {
                slice.Elts = append(slice.Elts, buildFieldCall(nf))
            }
        }
    }

    return decorator.Print(f, outPath)
}

func buildFieldCall(spec FieldSpec) dst.Expr {
    // assembles: field.String("phone").Unique().Optional()
    base := callExpr(selector("field", spec.Kind), strLit(spec.Name))
    for _, opt := range spec.Opts {
        base = methodCall(base, opt)
    }
    return base
}
```

**Output written to consumer's `modules/mod-auth/schema/user.go`:**

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique(),
        field.String("password").Sensitive(),
        field.String("phone").Unique().Optional(),     // ← injected
        field.String("department"),                    // ← injected
        field.Time("deleted_at").Optional().Nillable(), // ← injected
    }
}
```

Real Go in. Real Go out. Compiler validates both ends.

### Example 2 — Inject fields into api/types.go

**Source:**

```go
package api

type User struct {
    ID    string
    Email string
}
```

**Config:**

```yaml
extensions:
  api/types.go:
    User:
      - { name: Phone, type: string, json: phone }
      - { name: Department, type: string, json: department }
      - { name: DeletedAt, type: "*time.Time", json: deleted_at, omitempty: true }
```

**Output:**

```go
package api

type User struct {
    ID         string
    Email      string
    Phone      string     `json:"phone"`              // ← injected
    Department string     `json:"department"`          // ← injected
    DeletedAt  *time.Time `json:"deleted_at,omitempty"` // ← injected
}
```

### Example 3 — Replace method body (behavior swap)

**Source — consumable's `module.go`:**

```go
func (m *Module) HashPassword(plain string) (string, error) {
    h, err := bcrypt.GenerateFromPassword([]byte(plain), m.cfg.BcryptCost)
    return string(h), err
}
```

**Config:**

```yaml
extensions:
  module.go:
    Module.HashPassword:
      replace: |
        return argon2id.CreateHash(plain, &argon2id.Params{
            Memory: 64 * 1024,
            Iterations: 3,
            Parallelism: 4,
            SaltLength: 16,
            KeyLength:  32,
        })
    add_imports:
      - github.com/alexedwards/argon2id
```

**Output:**

```go
import "github.com/alexedwards/argon2id"   // ← injected import

func (m *Module) HashPassword(plain string) (string, error) {
    return argon2id.CreateHash(plain, &argon2id.Params{
        Memory: 64 * 1024,
        Iterations: 3,
        Parallelism: 4,
        SaltLength: 16,
        KeyLength:  32,
    })
}
```

## Tool: mod-cli

```
github.com/khanakia/mod-cli/
├── go.mod
├── cmd/
│   └── scaffold/main.go            ← CLI entry: parse flags + dispatch
├── pkg/
│   ├── ast/
│   │   ├── struct_extender.go      ← inject fields into struct decl
│   │   ├── ent_schema_extender.go  ← inject fields into Fields() return
│   │   ├── method_replacer.go      ← swap function body + imports
│   │   ├── import_manager.go       ← add/remove imports cleanly
│   │   └── helpers.go              ← dst utilities (selector, callExpr, ...)
│   ├── config/
│   │   └── parser.go               ← YAML → ExtensionSpec
│   ├── fetch/
│   │   └── git.go                  ← shallow git clone of source
│   └── apply/
│       └── orchestrator.go         ← fetch → mutate → write
├── testdata/                       ← golden files for AST mutation tests
└── README.md
```

**Dependencies:**

| dep | role |
|---|---|
| github.com/dave/dst | Decorated Syntax Tree (formatting + comment preserved) |
| github.com/dave/dst/decorator | parse + print |
| gopkg.in/yaml.v3 | config parsing |
| github.com/spf13/cobra | CLI plumbing (already in repo) |
| github.com/go-git/go-git/v5 | fetch source repo |

## Config schema (.mod-cli.yaml)

```yaml
# .mod-cli.yaml — declarative extensions

source:
  type: git
  url: github.com/khanakia/mod-auth-template
  ref: v1.0.0                          # tag, branch, or commit SHA

target: modules/mod-auth                # where to write output

extensions:
  # ent schema field injection
  schema/user.go:
    User.Fields:
      - { kind: string,  name: phone,        opts: [unique, optional] }
      - { kind: string,  name: department }
      - { kind: bool,    name: email_verified, default: "false" }

  # struct field injection
  api/types.go:
    User:
      - { name: Phone,         type: string,     json: phone }
      - { name: Department,    type: string,     json: department }
      - { name: EmailVerified, type: bool,       json: email_verified }

  # method body replacement + imports
  module.go:
    Module.HashPassword:
      replace: |
        return argon2id.CreateHash(plain, argon2id.DefaultParams)
      add_imports:
        - github.com/alexedwards/argon2id

post_hooks:
  - go run ./cmd/entg
  - go mod tidy
  - go build ./...

# version pinning written into output for traceability
version_marker:
  file: .scaffold-meta.json
  contents:
    template: github.com/khanakia/mod-auth-template
    version: v1.0.0
    scaffolded_at: <auto>
```

## Phased build roadmap

| phase | when | scope |
|---|---|---|
| 0. Don't build | now (no consumer needs it) | keep shipping V2 features as-is |
| 1. Minimal viable mod-cli | when 1st consumer asks for custom field | scaffold + struct field injection (~300 LOC) |
| 2. Behavior swap support | when consumer needs argon2 / different JWT | method replacer + import manager (~150 LOC) |
| 3. Re-scaffold | when consumer wants to update from upstream template | 3-way merge tool (HARD) |
| 4. Multi-module scaffolding | when consumers want diverging mod-billing too | generalize from mod-auth-only |
| 5. Registry of consumable modules | when ecosystem grows | catalog + version pinning + signing |

**Trigger conditions to start Phase 1:**

| trigger | build-now? |
|---|---|
| 2+ consumers need different User fields | yes |
| copy-pasting modules between projects becomes routine | yes |
| selling boilerplate as product (each customer customizes) | yes |
| single team using monorepo only | no — never need it |
| no real consumer demand | no — defer |

## Build effort estimate

| piece | scope | effort |
|---|---|---|
| fetch consumable module (git shallow clone) | ~50 LOC | low |
| struct field append (struct decl) | ~80 LOC | low |
| ent schema Fields() append | ~120 LOC | med |
| method body replace | ~100 LOC | med |
| import manager (add without dup) | ~60 LOC | low |
| config parsing (YAML → spec) | ~80 LOC | low |
| CLI plumbing (cobra) | ~80 LOC | low |
| golden-file tests for AST mutations | ~200 LOC | low |
| **phase 1 total** | **~770 LOC Go** | **med (1-2 days)** |
| re-scaffold (3-way merge) | hard problem | high (1-2 weeks) |
| multi-module orchestration | medium | med (3-5 days) |

Phase 1 deliverable in 1-2 days of focused work.

## What stays decoupled today

The current V2 modules require **zero changes** to be AST-friendly. They
already are real Go with predictable structure:

| property | current-state |
|---|---|
| modules compile standalone | yes |
| schemas in identifiable files (schema/*.go) | yes |
| api/types.go is flat struct | yes |
| behavior on Module struct in module.go | yes |
| Fields() returns []ent.Field literal | yes |
| ent codegen is local per-module | yes |
| no exotic dynamic constructions in source | yes |

When you build mod-cli later, it reads what's there. **No migration step.**

## Light hygiene to maintain (low cost, future-friendly)

| practice | reason |
|---|---|
| keep `field.String("name").Unique()` style (not exotic) | AST tool finds these reliably |
| keep behavior methods on Module struct (not nested types) | easy to find via AST receiver match |
| keep api/types.go a flat struct (no embedding) | easy to inject fields |
| keep imports clean (no anonymous _ imports for side-effects) | import manager simpler |
| keep one schema per file (user.go = User schema only) | tool addresses by filename |

These are good practices anyway. None require thinking about AST when building features.

## Open questions

<Details summary="Re-scaffold: how does upstream template change flow into already-scaffolded consumers?">

Hard problem. Three options:

1. **Manual diff** — consumer compares old scaffold output with new, applies upstream changes by hand. Most template tools punt this.
2. **Three-way merge** — tool retains pristine scaffold output, applies upstream delta, prompts consumer for conflicts. Like git rerere.
3. **Regenerate + conflict markers** — re-scaffold writes new files, leaves consumer to merge using their normal git workflow.

Defer. Address when first re-scaffold need arises.

</Details>

<Details summary="Versioning the consumable module itself?">

Use Git tags on the template repo. Consumer's `.mod-cli.yaml` pins
`source.ref: v1.0.0`. Tool clones at that ref. Output includes
`.scaffold-meta.json` with version + timestamp for audit.

</Details>

<Details summary="What if consumer wants to remove a field the consumable shipped with?">

Phase 1 only adds. Phase 2 could support `extensions.User.RemoveFields:
[some_field]`. Or consumer just edits the output by hand after scaffold —
it's their code now.

</Details>

<Details summary="Multi-module template (mod-billing also needs custom fields)?">

Phase 4. Same mechanism, scaffold each module separately with its own
config block. Single `.mod-cli.yaml` lists multiple sources + targets.

</Details>

<Details summary="What if a consumer wants a custom mixin (not field-by-field)?">

Either: (a) write the mixin in their own repo and add it via
`schema/user.go:User.Mixin: append`, OR (b) ship the mixin in
mod-auth-template's `mixins/` and consumer enables via config flag.

</Details>

## Decisions to lock

| decision | status |
|---|---|
| AST is the customization mechanism for typed schema extension | locked |
| github.com/dave/dst is the parsing library | locked |
| mod-cli is the tool name | locked (subject to bikeshed) |
| .mod-cli.yaml is the config format | locked (YAML over JSON for comments) |
| build is deferred until consumer demand triggers | locked-deferred |
| existing V2 modules need zero changes today | locked |
| re-scaffold strategy | parked (defer to Phase 3) |
| multi-language template support | parked (Go-only for now) |

## Cross-references

- `Golang_Pluggability.md` (in private-repo) — all 11 approaches; AST is Approach 4 evolved
- `MISSION_V2_DECISIONS.md` — ADR-V2-013 locks AST as final mechanism
- `MISSION_V2_BUILD_LOG.md` — what's actually built (V2 base)
- `MISSION_V2_TRANSCRIPT.md` — conversation that produced this decision
- `MISSION_V2_LOG.md` — topic-organized summary

## TL;DR

**Final extension mechanism for V2:** AST scaffolding via `mod-cli` tool
using `github.com/dave/dst`. Module ships as real working Go; consumer
declares extensions in `.mod-cli.yaml`; tool mutates AST and writes real
Go to `modules/mod-auth/`.

**Build status:** deferred until consumer demand triggers. Existing V2
modules require zero changes to be AST-ready when the time comes.

**Default plan:** keep building features on V2's current shape. Pick up
mod-cli the day a consumer's schema diverges from yours. Phase 1 is
~770 LOC, 1-2 days of focused work.

