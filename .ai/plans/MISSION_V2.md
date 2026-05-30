---
title: Mission V2 — Pluggable, Loosely Coupled Boilerplate
status: proposed
owner: khanakia
created: 2026-04-24
defaults:
  project: mission-v2
  area: work/architecture
  owner: khanakia
---

# Mission V2 — Pluggable Boilerplate

Reduce coupling so features can be authored, tested, and removed independently. Accept the few couplings the stack forces (codegen) and maximize leverage everywhere else.

<Callout type="info" title="TL;DR">

"Pluggable" means **build-time composition**, not runtime module loading. Ent + gqlgen + PKL are codegen-driven — schemas must be known at compile time. We decouple everything around them.

</Callout>

## Why

Today's pain:

- `app.Plugin` is a god-struct. Every feature reaches into `r.Plugin.EntDB`, `r.Plugin.Natso`, `r.Plugin.Redis`, etc. → shotgun coupling.
- Features spread across `saas/pkg/*`, `apidash/resolver/*`, `dbent/schema/*`, `saas/cmd/cron/main.go`, NATS callback. Removing one = 6 edits.
- `Root.pkl` is hand-merged. Adding a feature config = edit central file.
- Resolvers untestable without booting whole plugin graph.
- No way to disable a feature at boot (feature flags live in code).

## Non-goals

- **Not runtime hot-loading** — Go compiles monolithically. Accept it.
- **Not micro-services** — stays one deploy target (API + cron + CLI binaries).
- **Not DI framework (wire/fx)** — options pattern is enough.
- **Not rewriting `lace/` or `core/`** — marked read-only. V2 lives in app layer.

## Coupling audit

<Callout type="warning" title="Accept these couplings">

Ent, gqlgen, and PKL drive codegen. You cannot register schemas at runtime. V2 keeps them — we only make the *inputs* to codegen composable.

</Callout>

| area | coupling | can-fix | priority | notes |
|---|---|---|---|---|
| app.Plugin god struct | every feature reaches into it | can-fix | p0 | Replace with narrow HostAPI iface |
| Feature sprawl | logic across 5 dirs per feature | can-fix | p0 | Feature-first layout in modules/<name>/ |
| Init order | hardcoded in app.New() | can-fix | p1 | Module iface + DependsOn graph |
| Cron registration | one monolithic main.go | can-fix | p1 | CronRegistrar self-registration |
| NATS handlers | WithAfterNatsInitiated callback | can-fix | p1 | NATSRegistrar self-registration |
| CLI commands | switch in cli/main.go | can-fix | p1 | CLIRegistrar self-registration |
| HTTP routes | wired in api main | can-fix | p1 | HTTPRegistrar self-registration |
| PKL Root | hand-edited per feature | can-fix | p2 | Module ships config.pkl snippet merged at codegen |
| Ent schemas | one dir dbent/schema/ | partial | p2 | Glob modules/*/schema/ at codegen |
| gqlgen schemas | one dir per gql module | partial | p2 | Glob modules/*/graphql/*.graphqls |
| Ent codegen itself | compile-time only | accept | p3 | Cannot be runtime-pluggable |
| gqlgen codegen itself | compile-time only | accept | p3 | Cannot be runtime-pluggable |
| Single DB pool | one pool shared | accept | p3 | Multi-tenant is different problem |

## Architecture decision

<ADR status="proposed" id="ADR-V2-001" date="2026-04-24" title="Module + HostAPI pattern for V2">

### Context

Feature code is scattered across 5+ directories and every piece reaches into `*app.Plugin` directly. Adding or removing a feature requires touching all of: `dbent/schema/`, `apidash/internal/graph/*.graphqls`, `apidash/internal/graph/resolver/`, `saas/pkg/<feat>/`, `saas/cmd/cron/main.go`, NATS init callback, optionally `saas/cmd/cli/main.go`, optionally PKL. Resolvers are untestable without booting the whole plugin graph.

### Decision

Introduce two contracts:

1. **`HostAPI`** — narrow interface exposing `{DB, Log, Trace, Publish, Cache, Now, Config}`. Replaces direct `*app.Plugin` reach-ins.
2. **`Module`** — per-feature type with `Name() string`, `Init(HostAPI) error`, `Shutdown(ctx) error`, `DependsOn() []string`.

Modules self-register via registrar interfaces they optionally implement: `CronRegistrar`, `NATSRegistrar`, `CLIRegistrar`, `HTTPRegistrar`, `GQLResolverRegistrar`. Host inspects each module, collects artifacts, wires them.

Layout: each feature lives in `modules/<name>/` with its own `module.go`, `schema/`, `graphql/`, `service/`, `cron/`, `nats/`, `cli/`, `http/`, `config.pkl`. Codegen (ent, gqlgen, PKL) globs across `modules/*/` instead of central dirs.

### Consequences

- **Win:** Features become removable (one dir delete + re-run codegen). Resolvers testable with fake `HostAPI`. Cron/NATS/CLI self-register. Config composes.
- **Win:** Module enable/disable gate at host boot enables cheap feature flags.
- **Cost:** More files + `module.go` boilerplate per feature. Worth it past ~5 features; not worth it at 2.
- **Cost:** Codegen taskfiles need glob updates. Risk of cyclic Ent edges across modules — enforce at review time.
- **Cost:** `HostAPI` is a living interface; every new cross-cutting need (metrics, feature flags, tenant ctx) adds a method. Must stay disciplined or it becomes `Plugin` again.
- **Accept:** Ent + gqlgen still codegen'd from union of module inputs. No runtime module discovery.

</ADR>

## Phased rollout

<TaskTimeline md lane="status"></TaskTimeline>

### Phase 1 — prove the pattern (lowest risk)

Kickoff **2026-04-28**, target **2026-05-23**. Deliverables:

- [ ] Design HostAPI iface — minimal surface, no wildcards !p0 ~2026-04-29 &4h #arch ^task-v2-hostapi-design
- [ ] Implement HostAPI adapter over existing app.Plugin !p0 ~2026-05-02 &1d after:task-v2-hostapi-design ^task-v2-hostapi-impl
- [ ] Design Module iface + 5 registrar ifaces !p0 ~2026-05-04 &4h #arch ^task-v2-module-iface
- [ ] Build modules/ loader in saas/pkg/host/ !p0 ~2026-05-08 &2d after:task-v2-module-iface ^task-v2-loader
- [ ] Pick pilot feature (candidate: users) and convert to modules/users/ !p0 ~2026-05-15 &3d after:task-v2-loader ^task-v2-pilot-users
- [ ] Update taskfile globs for ent + gqlgen across modules/*/ !p1 ~2026-05-16 &4h after:task-v2-pilot-users ^task-v2-codegen-globs
- [ ] Write fake HostAPI for resolver tests !p1 ~2026-05-20 &4h ^task-v2-fake-host
- [ ] Document module authoring pattern in .ai/skills/backend/ !p1 ~2026-05-22 &2h after:task-v2-pilot-users ^task-v2-docs

**Exit criteria:** pilot feature runs in prod with same behavior. Adding feature #2 feels strictly lighter than today. If not → revert, rethink.

### Phase 2 — migrate (medium risk)

Kickoff **2026-06-01**, target **2026-07-15**. One feature at a time.

- [ ] Inventory existing features → migration order !p1 ~2026-06-02 &4h ^task-v2-inventory
- [ ] Migrate feature #2 (auth) !p1 ~2026-06-08 &2d after:task-v2-inventory ^task-v2-migrate-auth
- [ ] Migrate feature #3 !p2 ~2026-06-15 &2d after:task-v2-migrate-auth ^task-v2-migrate-3
- [ ] Migrate feature #4 !p2 ~2026-06-22 &2d ^task-v2-migrate-4
- [ ] Migrate remaining features !p2 ~2026-07-10 &5d ^task-v2-migrate-rest
- [ ] PKL config composition across modules/*/config.pkl !p1 ~2026-07-12 &1d ^task-v2-pkl-compose
- [ ] Drop dead code from app.Plugin !p2 ~2026-07-14 &4h after:task-v2-migrate-rest ^task-v2-cleanup

### Phase 3 — polish (optional)

Kickoff **2026-07-20**, no hard deadline.

- [ ] Module DependsOn graph + topological init !p3 ^task-v2-depgraph
- [ ] Parallel shutdown with timeout !p3 ^task-v2-shutdown
- [ ] Feature flags → module enable/disable at boot !p3 ^task-v2-flags
- [ ] Per-module metrics + tracing labels !p3 ^task-v2-per-module-metrics

## Board view

<Kanban md group-by="status" order="todo,wip,done" title="Mission V2 — all tasks"></Kanban>

## By priority

<TaskList group-by="priority" filter="is:open" sort="priority:asc,due:asc" title="Open tasks"></TaskList>

## Proposed layout

```
modules/
  users/
    module.go              # implements Module + registrars
    schema/
      user.go              # ent schema (globbed by task dbent:entg)
    graphql/
      user.graphqls        # gqlgen (globbed by task apidash:gql)
      user.resolvers.go    # generated landing point
    service/
      service.go           # business logic, takes HostAPI
      service_test.go      # fake HostAPI
    cron/
      cleanup.go           # registered via CronRegistrar
    nats/
      handlers.go          # registered via NATSRegistrar
    cli/
      commands.go          # registered via CLIRegistrar
    http/
      routes.go            # registered via HTTPRegistrar (if any)
    config.pkl             # merged into Root at codegen

saas/pkg/host/             # replaces saas/pkg/app as host
  host.go                  # HostAPI impl, module loader
  module.go                # Module + registrar ifaces
  testing/
    fakehost.go            # drop-in HostAPI for tests
```

## HostAPI sketch

```go
// saas/pkg/host/api.go
type HostAPI interface {
    DB() *ent.Client
    Log(ctx context.Context) *zap.Logger   // wraps gozap.FromCtx
    Tracer(name string) trace.Tracer
    Publish(ctx context.Context, subject string, payload []byte) error
    Cache() cache.Cache
    Now() time.Time                         // injectable clock for tests
    Config() *config.Root                   // read-only
}
```

Disciplined keepers — no direct `*ent.Client` type leakage beyond `DB()`, no NATS conn leak beyond `Publish`. If a module needs more → it's an escape hatch; audit before allowing.

## Module sketch

```go
// saas/pkg/host/module.go
type Module interface {
    Name() string
    Init(HostAPI) error
    Shutdown(ctx context.Context) error
    DependsOn() []string
}

type CronRegistrar interface {
    RegisterCron(scheduler gocron.Scheduler) error
}

type NATSRegistrar interface {
    RegisterNATS(js nats.JetStreamContext) error
}

type CLIRegistrar interface {
    RegisterCLI(root *cobra.Command)
}

type HTTPRegistrar interface {
    RegisterHTTP(r gin.IRouter)
}

type GQLResolverRegistrar interface {
    RegisterResolvers(reg *gqlreg.Registry)
}
```

## Open questions

<Details summary="Ent cross-module edges — how to handle?">

If `modules/orders/` has `edge.From("customer", modules/users.User.Type)`, the import creates a dependency arrow. Options:

1. Allow — module dep graph mirrors ent edges. Document as a real constraint.
2. Ban cross-module edges — forces ID-only references + join in service layer. Loses ent's eager-load magic.
3. Shared `schema/` for cross-cutting entities (User, Tenant) outside modules/.

Lean toward **option 3** for a small set of core entities, option 1 for everything else.

</Details>

<Details summary="gqlgen resolver registry — does it exist?">

gqlgen generates one `ResolverRoot` per module. Current setup has one module (apidash). To keep modules independent, either:

1. Keep one apidash module and have each `modules/<feat>/` contribute resolver files that gqlgen discovers via glob.
2. Run gqlgen per feature module → multiple GraphQL endpoints. Probably overkill.

Lean toward **option 1** — one endpoint, globbed schemas + resolver files.

</Details>

<Details summary="Migration path without a long freeze?">

Phase 2 migrates one feature at a time. During migration, both layouts coexist — old features still in `saas/pkg/*`, new ones in `modules/*/`. HostAPI adapts over legacy `app.Plugin`. No big-bang cutover.

</Details>

<Details summary="Who holds CLI root cmd + gocron scheduler?">

Host owns both. Modules register onto them via registrars. Only `saas/cmd/{api,cron,cli}/main.go` stay — each constructs Host with the module list.

</Details>

## Risks

<Callout type="danger" title="HostAPI drift">

Every new cross-cutting concern (feature flags, tenant ctx, metrics, audit log) tempts a new `HostAPI` method. At 20+ methods it's the Plugin again. **Rule:** new method requires ADR. Prefer passing via ctx or via module config over adding a method.

</Callout>

<Callout type="warning" title="Codegen breakage">

Ent + gqlgen globbing across modules is the highest-risk change in Phase 1. Test `task dbent:entg` + `task apidash:gql` on the pilot before committing. If edges or schema types collide, fall back to central schemas for shared entities and module-scoped schemas elsewhere.

</Callout>

<Callout type="warning" title="Over-engineering at 2 features">

If repo stays at 2–3 features, V2 is net-negative. Commit to V2 only if the roadmap has 5+ features landing in 12 months.

</Callout>

## Success metrics

<Stats cols="3">
  <Stat title="Files touched to add a feature" value="~25" delta="→ ~8" intent="success" description="measured on pilot" />
  <Stat title="Files touched to remove a feature" value="~15" delta="→ 1 dir" intent="success" />
  <Stat title="Resolver test boot time" value="full plugin" delta="→ fake HostAPI" intent="success" />
</Stats>

Numbers are proposals — validate on the pilot before committing to the phase 2 plan.

## References

- Current architecture: `.ai/ARCHITECTURE.md`
- Stack: `.ai/STACK.md`
- Patterns: `.ai/PATTERNS.md`
- Backend rules: `.ai/skills/backend/RULES.md`
- Prior plan: `.ai/plans/PLAN.md`
