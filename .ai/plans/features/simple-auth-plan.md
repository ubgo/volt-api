# Plan: Simple Authentication

**Status**: in-progress
**Created**: 2026-04-18
**Priority**: high
**User**: @khanakia

## Goal
Add JWT-based email/password auth for GraphQL API: register, login, authenticated context extraction, protected resolvers. Minimal surface, leverage existing `User` entity (`dbent/schema/user.go` has email + Sensitive password + is_active).

## Non-Goals
- OAuth / social login (future)
- Magic links / passwordless (future)
- Session revocation / token blocklist (future — add Redis if needed)
- Password reset flow (separate plan)
- Email verification (separate plan)
- Role/permission system beyond `is_active` check (future)
- Refresh tokens (start stateless, single access token)

## Requirements

- **User schema confirmed** (Priority: high)
  - `email` unique, `password` Sensitive — already in `dbent/schema/user.go`
  - No schema change required for v1

- **`saas/pkg/auth/` package** (Priority: high)
  - `HashPassword(pw string) (string, error)` — bcrypt cost 12
  - `CheckPassword(pw, hash string) bool` — bcrypt compare
  - `GenerateToken(userID string, ttl time.Duration, secret string) (string, error)` — HS256
  - `ParseToken(token, secret string) (*Claims, error)`
  - `UserFromContext(ctx) *ent.User` + `WithUser(ctx, *ent.User) ctx`
  - Context key = unexported type

- **PKL config** (Priority: high)
  - `auth.jwtSecret = read("env:AUTH_JWT_SECRET")`
  - `auth.tokenTTL = 24.h` (configurable)
  - Add `Auth.pkl` schema + wire into `Root.pkl`
  - Update `config/pkl/env/default/app.pkl`

- **GraphQL schema + resolvers** (Priority: high)
  - Schema: `apidash/internal/graph/schemas/auth.graphql`
    ```
    type AuthPayload { token: String!, user: User! }
    input RegisterInput { email: String!, password: String!, firstName: String, lastName: String }
    input LoginInput { email: String!, password: String! }
    extend type Mutation {
      register(input: RegisterInput!): AuthPayload!
      login(input: LoginInput!): AuthPayload!
      me: User  # current user or null
    }
    ```
  - `me` → Query, not Mutation (fix in impl)
  - Resolvers in `apidash/internal/graph/resolver/auth.resolvers.go`
  - Delegate to `saas/pkg/auth/` service — thin resolver
  - Use `valgo` for input validation (email format, password min 8)

- **Gin middleware** (Priority: high)
  - `auth.Middleware(secret string) gin.HandlerFunc`
  - Parse `Authorization: Bearer <jwt>` → load user from DB → inject into ctx
  - Missing/invalid token → continue without user (resolvers decide if required)
  - Wire into API bootstrap (`saas/cmd/api/main.go` or `saas/pkg/httppkg/`)

- **Protected resolver guard** (Priority: medium)
  - Helper `auth.RequireUser(ctx) (*ent.User, error)` → returns unauth error
  - Example: gate existing `user` query / mutations

- **Tests** (Priority: medium)
  - Table-driven: hash/check, token sign/parse (happy + expired + bad secret)
  - Resolver tests with SQLite via `enttest` — register → login → me

## Tasks
- [ ] Add `saas/pkg/auth/` package (hash, jwt, context helpers)
- [ ] Add PKL `Auth` schema + env wiring → `task config:gen`
- [ ] Add `auth.graphql` schema → `task apidash:gql`
- [ ] Implement register / login / me resolvers
- [ ] Add Gin auth middleware, wire in API bootstrap
- [ ] Add `RequireUser` guard helper
- [ ] Unit tests for auth pkg
- [ ] Resolver integration tests
- [ ] Update `.claude/rules/authentication.md` with concrete file refs

## Decisions
- **JWT over sessions**: stateless, no Redis dep for v1. Tradeoff: no server-side revoke. Acceptable for boilerplate.
- **HS256 over RS256**: single service, symmetric simpler. RS256 when/if we split API.
- **bcrypt cost 12**: OWASP baseline. Tune if login latency becomes issue.
- **Separate access token only (no refresh)**: 24h TTL. Refresh tokens if session UX demands.
- **Context injection via middleware**: avoids per-resolver token parsing. Resolvers just call `UserFromContext`.

## Handoff Context
<!-- Self-contained within this repo. No downstream yet. Frontend will consume via GraphQL. -->

## Related
- Rules: `.claude/rules/authentication.md`, `.claude/rules/secrets.md`
- Schema: `dbent/schema/user.go`
- Config pattern: `.claude/rules/configuration.md`
