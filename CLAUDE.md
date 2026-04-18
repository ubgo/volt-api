# Project: boilerplate-golang

<!-- GENERATED FILE — do not edit directly. Run /codeskill sync to regenerate. -->
<!-- Source: .ai/RULES.md, .ai/HOTFIXES.md, .ai/taste.md, .ai/global/ -->

## codeskill — ALWAYS ACTIVE

This project uses the `codeskill` skill as its default operating system. You MUST use it automatically — the user should never have to type `/codeskill` manually.

**Auto-trigger rules:**
- **Session start**: Run `/codeskill session start` at the beginning of every conversation
- **User corrects you**: Immediately run `/codeskill feedback` or `/codeskill learn` — don't wait to be asked
- **You make a decision**: Run `/codeskill decide` for non-obvious architectural choices
- **You make a mistake**: Run `/codeskill incident` — self-report, don't hide it
- **Before exploring code**: Check snapshots first (Exploration Protocol below)
- **Before starting a task**: Run `/codeskill task next` or `/codeskill task start <id>`
- **Task finished**: Run `/codeskill task done <id>`
- **Session ending**: Run `/codeskill session end` before the conversation ends
- **Context running low**: Run `/codeskill save-state` to write CONTINUATION.md
- **User says "remember", "learn", "don't forget"**: Run `/codeskill learn` immediately
- **User says "plan", "feature"**: Run `/codeskill plan`
- **User says "add task"**: Run `/codeskill task add`

**The goal: codeskill runs in the background like an OS. The user just works.**

**MANDATORY — non-negotiable:**
- You MUST run `/codeskill session start` as your FIRST action in every conversation. No exceptions.
- You MUST run `/codeskill session end` as your LAST action before the conversation ends. No exceptions.
- These log prompts, decisions, and corrections to `.ai/logs/`. Without them, the session is invisible.

## Where We Left Off
- **Last session**: _none — fresh init on 2026-04-18 by @khanakia_
- **Working on**: Scaffolding `.ai/` directory via `/codeskill init`
- **Next step**: Verify files, commit the scaffold, then start using codeskill workflows

<!-- Auto-updated by /codeskill session end -->

## Critical Rules

### Global — Generated Code (NEVER EDIT)

Before editing ANY file: check for `Code generated` header.
If present → find the source file, edit that, then run the regenerate command.

| Generated | Source | Regenerate |
|-----------|--------|------------|
| `*/gen/ent/*` | `dbent/schema/*.go` | `task dbent:entg` |
| `*/graph/generated/*` | `*/graph/schemas/*.graphql` | `task <module>:gql` |
| `*/graph/model/models_gen.go` | `*.graphql` schemas | `task <module>:gql` |
| `config/gen/*` | `config/pkl/**/*.pkl` | `task config` |

### Global — Function Signatures
- Context always first param: `func Do(ctx context.Context, ...)`
- Error always last return: `func Do(...) (*Result, error)`
- Utility functions that don't need context can skip it

### Global — Error Handling
- Use `goerr.New()` with options for user-facing errors
- Wrap errors with context: `fmt.Errorf("create user: %w", err)`
- Use `errors.As()` for type assertion, `errors.Join()` to consolidate multiple errors
- Never panic in libraries — return errors. Only `main()` or init can panic.
- Error messages tell the caller what to do, not just what failed

### Global — Imports (3 Groups)
```go
import (
    "context"          // 1. stdlib (alphabetical)
    "fmt"

    "lace/gozap"       // 2. local monorepo packages
    "saas/pkg/app"

    "github.com/gin-gonic/gin"  // 3. third-party (alphabetical)
)
```

### Global — Packages
- All lowercase, no underscores: `apidash`, `gozap`, `goerr`, `natso`
- Never generic names (`util`, `helpers`, `common`) — be specific

### Global — Constructors (Options Pattern)
```go
func New(opts ...Option) *Thing {
    cfg := &config{/* defaults */}
    for _, opt := range opts { opt(cfg) }
    return &Thing{cfg: cfg}
}
```
Use for any struct with 3+ configuration fields.

### Global — Database Access
- ALWAYS use `app.New()` → `app.GetPlugins().EntDB.Client` for DB access
- NEVER call `ent.Open()` or `sql.Open()` directly — skips config, encryption, pooling
- Applies to: CLI commands, cron jobs, seed scripts, tests — EVERYTHING

### Global — GraphQL
- Schema in `*/internal/graph/schemas/*.graphql`
- Resolvers ONLY in `resolver/` — no helpers there (gqlgen overwrites)
- Helpers → `internal/<pkg>/`, `lace/`, or `saas/pkg/`
- File uploads: base64 string, NEVER `graphql.Upload`
- Use `@internal` directive for admin-only endpoints

### Global — Logging
- Always use `gozap.FromCtx(ctx)` — extracts trace_id + span_id from OpenTelemetry
- Never use `fmt.Println` or `log.Println` in production code
- Structured fields: `zap.String("key", val)`, `zap.Error(err)`
- Don't log AND return errors — one or the other
- Log at boundaries (handlers, job start/end), not every function

### Global — Security
- Constant-time comparison for API keys: `crypto/subtle.ConstantTimeCompare()`
- Never commit `.env`, `*.creds`, `*.pem`, `*.key` files
- Extract user from context in resolvers, don't trust client headers
- Check permissions at resolver level (not middleware only)

### Global — Testing
- Table-driven tests with `t.Run()` subtests in `_test.go` beside source
- Use `t.Helper()` in test helper functions
- Use `t.Skipf()` when external deps unavailable (Redis, DB)
- Separate test DB index (e.g., Redis DB 1 for tests)

### Global — Config
- PKL is the config system — never use raw env vars
- Access via `config.Get()` singleton
- Never commit API keys or secrets in PKL files

### Global — Commits
- Present-tense conventional: `feat(scope): add user lookup`
- Never commit .env values, API keys, or credentials
- Never add AI attribution or co-author lines

### Global — AI Behavior
- Don't add features beyond what was asked
- Don't refactor adjacent code that isn't broken
- Don't add speculative error handling for impossible scenarios
- Don't update the plan or integrate ideas unless explicitly asked
- Match existing code style, even if you'd do it differently

### Project — Generated Code Regeneration

| Generated | Source | Regenerate |
|-----------|--------|------------|
| `dbent/gen/ent/*` | `dbent/schema/*.go` | `task dbent:entg` |
| `apidash/internal/graph/generated/*` | `apidash/internal/graph/*.graphqls` | `task apidash:gql` |
| `apidash/internal/graph/model/models_gen.go` | `apidash/internal/graph/*.graphqls` | `task apidash:gql` |
| `config/gen/*` | `config/pkl/**/*.pkl` | `task config:gen` |

### Project — Critical
- NEVER modify files in `core/`, `lace/`, `scripts/` (external modules, read-only)
- NEVER create helper files in `resolver/` — gqlgen overwrites them
- NEVER use `graphql.Upload` for file uploads — use base64 string
- ALWAYS run migrations after schema changes (`task dbent:migrate`)
- ALWAYS pass `context.Context` as first param to I/O functions

### Project — Workspace Modules

| Module | Purpose | Writable |
|--------|---------|----------|
| `saas/` | Core app: CLI, API bootstrap, business logic | yes |
| `apidash/` | GraphQL API (`/api/sa`) | yes |
| `dbent/` | Ent ORM schemas & migrations | yes |
| `config/` | PKL configuration | yes |
| `lace/` | Shared utility library (30+ packages) | **no** |
| `core/` | External core module | **no** |
| `scripts/` | Deploy/ops scripts | **no** |
| `examples/` | Example code | yes |

### Project — Tags in Comments

`@tag('dnd', 'skip')` — `dnd` means never remove the commented-out code block.

### Project — Ignore

Never grep, explore, or reference:
- `docs/framework`
- `*/gen/*`, `*/generated/*`, `*/vendor/*`

### Project — Commits
- Conventional format: `feat(scope): description`
- Never add "Co-Authored-By" lines or any mention of Claude/AI
- Never commit `.env`, `*.creds`, `*.pem`, `*.key`, or API keys

## Active Hotfixes

_No active hotfixes._

## Taste

### Global

**Code Philosophy**
- Explicit over implicit. If 2 more lines make it clearer, take them. — @khanakia
- Errors are values, not exceptions. Handle at every layer. — @khanakia
- Composition over inheritance. Small interfaces, concrete implementations. — @khanakia
- No magic. If readers can't trace the flow, it's wrong. — @khanakia
- Options pattern over 10-field config structs. — @khanakia

**Naming**
- Functions verb-first (`CreateUser`, `ValidateInput`) — @khanakia
- Variables: full words (`userCount` not `cnt`) — @khanakia
- Packages: single-word lowercase, never generic — @khanakia
- Constants describe the value (`MaxRetries` not `RETRY_CONST`) — @khanakia

**Structure**
- Flat > nested. Max 3 levels, then refactor. — @khanakia
- Early returns > deep else chains. — @khanakia
- Table-driven > switch (for tests and mappings). — @khanakia
- Group by feature, not by type. — @khanakia
- `internal/` for unexported service logic, `pkg/` for shared code. — @khanakia

**Prefer**
- Options pattern for constructors — @khanakia
- Context propagation everywhere — @khanakia
- Structured logging with zap fields, never string interpolation — @khanakia
- Schema-first GraphQL — @khanakia
- Ent mixins for shared fields (`BaseMixin`) — @khanakia

**Hate**
- Speculative generalization — @khanakia
- Comment-heavy code — @khanakia
- Deep abstraction layers with no value — @khanakia
- AI-generated code that looks AI-generated — @khanakia
- Global singletons with panic on nil — @khanakia
- `fmt.Println` in production code — @khanakia

### Project

- Monorepo with `go.work` — each service its own module. — @khanakia
- Separate binaries (API, cron, CLI) — not one monolith. — @khanakia
- Schema-first GraphQL: edit `.graphqls`, run `task apidash:gql`. — @khanakia
- Ent ORM with `BaseMixin` on every entity. — @khanakia
- PKL `read("env:VAR")` — never raw `os.Getenv()` in app code. — @khanakia
- NATS handlers in `app.WithAfterNatsInitiated()` callback, never at import time. — @khanakia
- `saas/pkg/<feature>/` for business logic. `lace/` for generic utilities only. — @khanakia
- Resolvers stay thin — delegate to `saas/pkg/` services. — @khanakia

## Active Behaviour

**default** — Balanced mode. Switch with `/codeskill behaviour <name>`.
Available: default, careful, review, debug, scaffold.

## Memory Protocol
- ALWAYS save user corrections → `.ai/memory/feedback/`
- ALWAYS log non-obvious decisions → `.ai/memory/decisions/`
- ALWAYS check feedback memory before starting related work
- When user says "remember this" or corrects you → save immediately
- Include @username attribution on all entries
- Tag every entry with `scope: project | global_candidate | global`

## Exploration Protocol
BEFORE spawning any Explore() agent:
1. Check `.ai/snapshots/INDEX.md` for existing snapshot
2. If exists → check staleness via `git diff --name-only <hash> HEAD -- <globs>`
3. Fresh → READ snapshot. Do NOT re-explore.
4. Stale → Re-explore, UPDATE snapshot.
5. Missing → Explore, CREATE snapshot.
NEVER re-explore a fresh snapshot unless user explicitly asks.

## Before Starting Work
1. Check `.ai/HOTFIXES.md` for active gotchas
2. Check `.ai/snapshots/` for relevant cached explorations
3. Identify task type → load relevant `.ai/workflows/*.md`
4. Check `.ai/prompts/` for a pre-built prompt matching this task
5. Check `.ai/memory/feedback/` for related corrections
6. Check `.ai/skills/backend/RULES.md` for domain-specific rules

## Auto-Behaviour Hints
If task touches these paths → suggest switching to `careful`:
- `**/auth*`, `**/session*`, `**/token*`
- `**/payment*`, `**/billing*`, `**/stripe*`
- database migrations (`dbent/schema/*`, anything touching `task dbent:migrate`)
- secrets, PKL config edits

## Workflow Index

_No workflows defined. Add to `.ai/workflows/` as patterns emerge._

Context-specific rules still live in `.claude/rules/` (legacy). See table in `AGENTS.md` for the full list — load relevant file before the matching action:
- `new-endpoint.md` · `new-schema.md` · `new-service.md` · `new-cli-command.md` · `new-cron-job.md`
- `migrations.md` · `configuration.md` · `nats-handler.md`
- `logging.md` · `tracing.md` · `error-handling.md` · `validation.md` · `testing.md`
- `authentication.md` · `secrets.md` · `deployment.md` · `github-secrets.md` · `docker-dev.md`
- `libraries.md` · `debugging.md` · `git-commit.md`

## After Session
ALWAYS run before ending:
1. Save any corrections received: `/codeskill feedback`
2. End session: `/codeskill session end`
If mistakes happened:
3. Create incident: `/codeskill incident`
If pattern is recurring (2+ times):
4. Promote to hotfix: `/codeskill hotfix add`

## Logging — Who Did What
All logs include user attribution (email + GitHub username from git config).
Stored in: `.ai/logs/sessions/`, `.ai/logs/activity/`, `.ai/logs/incidents/`

## Reference
- `.ai/ARCHITECTURE.md` — system overview
- `.ai/STACK.md` — tech stack
- `.ai/PATTERNS.md` — code conventions index
- `.ai/skills/backend/` — Go-specific rules, patterns, snippets
- `.ai/snippets/` — reusable code patterns
- `.ai/prompts/` — reusable prompt templates
- `.ai/guides/` — deep reference docs
- `.claude/rules/` — legacy context-specific rules (still authoritative)
