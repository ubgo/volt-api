---
title: Mission V2 — Build Log
status: completed
owner: khanakia
created: 2026-04-25
defaults:
  project: mission-v2
  area: work/architecture
  owner: khanakia
---

# Mission V2 — Build Log

Phase-by-phase record of what was built and where. All code compiles; the
modapi smoke test verifies the end-to-end flow.

<Callout type="tip" title="Run the smoke test">

```bash
task pkl:link            # stabilize PKL paths (no-op if no module PKL deps yet)
task modules:tidy        # tidy each module
task modapi:run          # run the V2 smoke test against your local Postgres
```

Or:

```bash
cd saas && go run ./cmd/modapi -smoke
```

Requires `DATABASE_URL` env (defaults to localhost Postgres).

</Callout>

## What got built

| phase | scope | result |
|---|---|---|
| 1 | modkit (interfaces only) | done |
| 2 | mod-auth (User entity + api/ + ent gen + config) | done |
| 3 | mod-workspace (Workspace entity + soft FK to User + cron registrar) | done |
| 4 | saas/pkg/modhost (Services impl + Loader + DispatchCron/CLI/HTTP/NATS) | done |
| 5 | saas/cmd/modapi (V2 reference entry + smoke test) | done |
| 6 | task pkl:link + modules:entg + modules:tidy + modapi:run | done |
| 7 | GraphQL extend type pattern documented (.ai/examples/graphql-extend) | done |
| 8 | workspace.daily_count cron via RegisterCron registrar | done |

## Final layout

```
modules/
├── modkit/                          ← contract repo (interfaces only)
│   ├── go.mod
│   ├── services.go                  ← Services{DB, WithTx, Log, Tracer, Publish, Now}
│   ├── module.go                    ← Module{Name, Init, Shutdown, DependsOn}
│   ├── registrars.go                ← Cron/NATS/CLI/HTTP/GQLResolver registrars
│   ├── tx.go                        ← TxFromContext / WithTxContext
│   └── page.go                      ← PageRequest / PageInfo
├── mod-auth/                        ← peer module
│   ├── go.mod
│   ├── api/                         ← contract: types, reader, writer, events, errors
│   ├── schema/                      ← ent: user.go, mixin_base.go
│   ├── gen/ent/                     ← committed entc output
│   ├── cmd/entg/main.go             ← codegen runner
│   ├── config/{Config.pkl, embed.go}
│   ├── internal/util/password.go    ← unimportable from outside
│   ├── deps.go
│   ├── module.go                    ← impl satisfies api.Reader+Writer+modkit.Module
│   └── Taskfile.yml
└── mod-workspace/                   ← peer module, depends on mod-auth/api
    ├── go.mod
    ├── api/                         ← contract: types, reader, writer, events, errors
    ├── schema/workspace.go          ← owner_id soft FK
    ├── gen/ent/
    ├── cmd/entg/main.go
    ├── config/{Config.pkl, embed.go}
    ├── deps.go                      ← depends on authapi.Reader
    ├── module.go
    ├── cron.go                      ← implements modkit.CronRegistrar
    └── Taskfile.yml

saas/
├── pkg/
│   ├── modhost/
│   │   ├── services.go              ← modkit.Services impl over *sql.DB
│   │   ├── loader.go                ← topological Init / Shutdown
│   │   └── registrars.go            ← DispatchCron/CLI/HTTP/NATS/GQL
│   ├── users/
│   │   └── service.go               ← EnrichedUser composition (ADR-V2-012)
│   └── lifecycle/
│       └── user_deletion.go         ← cascade orchestrator (ADR-V2-010)
└── cmd/
    └── modapi/main.go               ← reference entry point + -smoke flag
```

## Smoke test scenario (saas/cmd/modapi -smoke)

Walks the canonical end-to-end flow:

```
1. Open *sql.DB
2. Build modhost.Services
3. Construct auth + workspace modules
4. Loader.InitAll (topo order — auth before workspace)
5. Migrate per-module ent schemas
6. CREATE user with metadata (phone + timezone)
7. CREATE workspace with owner_id pointing at user
   → workspace module calls auth.Reader.Get to validate owner exists (cross-module read)
8. ListByOwner(user.ID) → returns the workspace
9. users.Service.Get(user.ID) → EnrichedUser with phone + timezone read from metadata
10. lifecycle.UserDeleter.Delete(user.ID)
    → services.WithTx wraps:
      a. workspace.DeleteByOwner(ctx, userID)
      b. auth.Delete(ctx, userID)  + emits "auth.user.deleted" via Publish
    → both writes share the same *sql.Tx via clientForCtx
11. Verify: workspace.Get returns ErrNotFound; auth.Get returns ErrNotFound
```

## Key architecture choices in code

| file | decision | adr |
|---|---|---|
| modules/modkit/services.go | Services interface keeps modkit dep-light | ADR-V2-003 |
| modules/mod-auth/api/*.go | Per-module api/ contracts (no ent leak) | ADR-V2-004 |
| modules/mod-auth/module.go clientForCtx | Tx-aware client wrapping | ADR-V2-009 |
| modules/mod-workspace/api/types.go OwnerID | Soft FK across modules | ADR-V2-006 |
| modules/mod-workspace/module.go DeleteByOwner | Cascade hook convention | ADR-V2-010 |
| modules/mod-auth/schema/user.go Metadata field | Module-level extension point | ADR-V2-011 |
| saas/pkg/users/service.go EnrichedUser | Host owns composition layer | ADR-V2-012 |
| saas/pkg/lifecycle/user_deletion.go | Host owns cascade orchestrator | ADR-V2-010 |
| saas/cmd/modapi/main.go | Explicit module list (not init-based discovery) | per Open Q |

## What still needs doing (before V2 is fully production-ready)

- [ ] Switch password hash from sha256 to bcrypt (uses cfg.BcryptCost which exists) !p1 ^task-v2-bcrypt
- [ ] PKL → typed Go config wiring (currently using config.Default()) !p1 ^task-v2-pkl-typed
- [ ] Add real publisher (NATS or watermill) instead of NoopPublisher !p1 ^task-v2-publisher
- [ ] Add HostAPI fields the reference host already has: Cache, Redis !p2 ^task-v2-host-extras
- [ ] Sidecar `user_profile` ent schema in dbent + swap users.Service from metadata to typed lookup !p2 ^task-v2-sidecar-profile
- [ ] DataLoader integration in apidash for batch user lookups !p2 ^task-v2-dataloader
- [ ] Wire `extend type User` into apidash gqlgen schema once sidecar lands !p2 ^task-v2-graphql-extend
- [ ] Migrate one existing legacy feature from saas/pkg/* to a new mod-* module !p1 ^task-v2-migrate-legacy
- [ ] Module-author guide in modkit README !p2 ^task-v2-author-guide
- [ ] Linter rule: every Reader/Writer impl method must call clientForCtx !p3 ^task-v2-lint-clientforctx

<TaskList md filter="is:open" group-by="priority" title="V2 follow-up backlog"></TaskList>

## How to add a new module (after this scaffolding)

1. `mkdir -p modules/mod-<name>/{api,schema,cmd/entg,config}`
2. Copy `mod-workspace/go.mod` template, change paths
3. Add to `go.work`
4. Write `api/{types,reader,writer,errors,events}.go`
5. Write `schema/<entity>.go` + reuse `mixin_base.go` from another module
6. `cd modules/mod-<name> && go mod tidy && go run ./cmd/entg`
7. Write `module.go` with `clientForCtx`, Reader/Writer impl, `publish` helper
8. Write `deps.go` declaring what it needs from other modules' api/
9. Add to `saas/cmd/modapi/main.go` wiring section
10. If module participates in cascades, add `*By<Foreign>` writer methods
11. If module schedules jobs, implement `RegisterCron`

## Known limitations

<Callout type="warning" title="Same module path inside this repo">

Modules are currently at `github.com/khanakia/boilerplate-golang/modules/mod-*`
because they live inside this repo's go.work. To split to standalone repos
later: change module path in go.mod, move dir to its own repo, drop the
`replace` directives, update consumer imports. Mechanical refactor.

</Callout>

<Callout type="warning" title="No PKL → Go config wiring yet">

Modules use `config.Default()` Go-side. The `Config.pkl` files exist but
host doesn't yet load them via pkl-go. `task pkl:link` is wired but the
PKL evaluator integration is a follow-up. Defaults are sufficient for the
smoke test today.

</Callout>

<Callout type="warning" title="apidash GraphQL not modified">

To avoid breaking the existing apidash gqlgen build, the `extend type User`
pattern is documented in `.ai/examples/graphql-extend/README.md` only.
Wiring happens once the sidecar table + DataLoader are in place.

</Callout>
