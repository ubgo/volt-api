# Project Taste

<!-- Project-specific overrides to global taste. Keep short — only deltas from global. -->

## Code Philosophy
- Monorepo with `go.work` — each service is its own module, changes isolated per module. — @khanakia
- Separate binaries (API, cron, CLI) — not one monolith. — @khanakia

## Patterns
- Schema-first GraphQL: write `.graphqls`, run `task apidash:gql`. Never code-first. — @khanakia
- Ent ORM with `BaseMixin` for id/timestamps on every entity. — @khanakia
- PKL config with `read("env:VAR")` — never raw `os.Getenv()` in app code. — @khanakia
- NATS handlers registered via `app.WithAfterNatsInitiated()` callback, never at import time. — @khanakia

## Structure
- `saas/pkg/<feature>/` for business logic. `lace/` for generic utilities only. — @khanakia
- Resolvers in `apidash/internal/graph/resolver/` stay thin — delegate to services in `saas/pkg/`. — @khanakia
