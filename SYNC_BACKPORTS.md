# Sync → Volt backport list

Core fixes and hardening that landed in `sync_go` (built on Volt API) but are missing from `volt-api`. Compiled 2026-08-21 by diffing the shared modules (`core`, `lace`, `config`, hooks, CI) and mining sync's commit history.

**Implementation status (2026-08-21): items 1–10 are APPLIED to the working tree** (all module builds, tests, gofmt, and vet green). Deltas vs the original plan: item 3 uses `{{.ROOT_DIR}}` instead of sync's `{{.TASKFILE_DIR}}` (the latter resolves to the module dir inside included Taskfiles); item 7 ships as registration scaffolding (`clicmd.RegisterRekeyTargets`) because the bare boilerplate has zero reversibly-encrypted columns — the command fails loudly when run with none registered; item 9 keeps `otelkit.WithEnvOverrides(false)` (PKL stays the single config authority — `OTEL_SERVICE_NAME` flows through PKL `read?()`) and adds `volt/docs/guides/otel-service-naming.md`. Item 11 (boot-preflight) skipped: volt has no external storage/credential config to preflight. Item 12 was a decision note, no action.

Note on noise filtering: sync's `lace` still carries the old stack (`gozap`, `gotel`, `goerr`, `crypt`, `cache`, `natso`, `bootinfo`, …) that volt intentionally replaced with `ubgo/*` packages, and sync has many app modules volt doesn't (`wfrun`, `prism`, `filemgr`, `kvstore`, `pim`, `addons`, …). Those divergences are excluded; only genuine fixes volt lacks are listed.

## P1 — real bugs / incident-driven fixes

### 1. Churn-safe DB pool defaults (the "max connection" / EADDRNOTAVAIL fix)

- Sync commit: `fd84ccc` — `fix(db): churn-safe pool defaults — idle == open, bounded, recycled` (2026-08-18, incident-driven).
- The bug: Go's `database/sql` defaults (`MaxIdleConns=2`, unlimited opens) turn burst load into open/close churn. Nine concurrent jobs flooded TIME_WAIT sockets to pgbouncer and exhausted the ephemeral-port tuple space → `connect: cannot assign requested address`, while ESTABLISHED counts stayed in double digits.
- The fix in sync: `lace/db` applies defaults `MaxOpen=MaxIdle=30`, `ConnMaxLifetime=30m`, `ConnMaxIdleTime=5m`, all overridable via new options (`WithMaxOpenConns`, `WithMaxIdleConns`, `WithConnMaxLifetime`, `WithConnMaxIdleTime`), with rationale comments on the constants and `db_test.go` pinning defaults-applied, override, and idle==open.
- Volt status: `modules/lace/db/config.go` + `db.go` have NO pool sizing at all — stock Go defaults, no tests. Same latent incident waiting.
- Files to port: `modules/lace/db/config.go`, `modules/lace/db/db.go`, `modules/lace/db/db_test.go`.
- Corollary: any secondary `sql.Open` pool (sync patched its cron's kvstore-prune pool in the same commit) needs the same sizing — audit volt's cron/CLI for extra pools when porting.

### 2. Fail-closed boot guard on dev-default encryption key

- Sync commit: `49ce911` — `feat(core): fail closed on dev-default encryption key in non-dev`.
- The fix in sync: `crypt.IsWeakOrDefaultKey` (`modules/lace/crypt/weakkey.go` + test) enumerates the empty case plus the well-known committed placeholder keys (`abcdefghijklmnopqrstuvwxyz012345`, `…123456`); boot (`core/pkg/plugin/app.go`) panics when `env != dev && IsWeakOrDefaultKey(Core.EncryptionKey)` instead of silently encrypting production data with a guessable key.
- Volt status: missing entirely. Worse — volt's ubgo migration dropped the boot-time `crypt.LoadKey` from `core/pkg/volt/app.go` altogether; the only `EncryptionKey` consumer is the `crypt` CLI command (`modules/core/pkg/clicmd/clicmd.go:60`). No weak-key check anywhere. Port the guard adapted to the ubgo/crypt wiring.

### 3. Repo-root env override for PKL `rootdir`

- Sync commit: `481b5d8` — `fix(config): SYNC_ROOT env makes rootdir resolve to repo root`.
- The bug: `rootdir = "."` in `pkl/env/default/app.pkl` resolves against whatever cwd the binary inherited, so tasks that chdir (e.g. `dev:portless` running from the app dir) scatter `tmp/`, `uploads/`, and `logs/` into subdirectories.
- The fix in sync: PKL reads `rootdir = read?("env:SYNC_ROOT") ?? "."`, and the top-level Taskfile exports `SYNC_ROOT: "{{.TASKFILE_DIR}}"` so every task resolves paths from the repo root.
- Volt status: `modules/config/pkl/env/default/app.pkl` still has bare `rootdir = "."` and the Taskfile exports no root env. Port as `VOLT_ROOT` (Taskfile `env:` block + the PKL `read?()` fallback + a rationale comment on `core.rootdir`).

### 4. Trace-correlated logs lose the binary TraceId (gozap fix `57a90c0`, re-broken by the ubgo migration)

- Sync commit: `57a90c0` — `OTEL - gozap: stamp ctx field so otelzap populates TraceId column`. The bug it fixed: log records reached HyperDX with `trace_id` only as a string attribute; the canonical binary TraceId column stayed empty, so the UI's "Trace" tab on a log row showed "No trace id found for event". The cure was passing the real `context.Context` through to the OTLP emitter.
- Volt status: the mechanism exists but the wiring defeats it. `ubgo/logger`'s otel sink correctly emits with the record's ctx (`Sink.Emit` → `s.lg.Emit(r.Ctx, rec)`), **but** `applog.FromCtx` (`modules/core/pkg/applog/applog.go:73-81`) returns `global.With(string trace_id/span_id)` and every caller then uses the non-Context level methods (`.Error(msg)`, `.Info(msg)`), which hardcode `context.Background()` into the record. Net effect: OTLP log records leave with an empty binary TraceId — the same symptom sync fixed, resurrected through the `FromCtx` ergonomic.
- Fix shape: make `FromCtx` return something that binds ctx into emissions (ubgo `Logger` has `ErrorContext(ctx, …)` variants — either a ctx-carrying wrapper, or migrate call sites), keeping the string fields for console/file sinks.

## P2 — hardening / infrastructure volt lacks

### 5. Pre-commit hook hardening (post-2026-05-30 revisions)

- Sync commits: `6e4ebb2`, `df3ae22`, plus a later revision (sync's hook is dated 2026-07-11; volt's copy is the 2026-05-30 version).
- What volt is missing from sync's newer hook: **Layer 0 — refuse compiled binaries by magic bytes** (Mach-O/ELF/PE/fat-binary prefixes, catching Go's unpredictably-named build outputs before they become a 30 MB commit), and additional allowlist entries (`asec_demo…`, `X-Shopify-Access-Token`, Google/GitHub OAuth token-URL lines). Sync-specific allowlist entries should be adapted, not copied blindly.
- `.gitleaks.toml` also differs between the repos — re-sync the shared parts (the per-match allowlist philosophy from `df3ae22`: never whole-line `grep -v` skips, and no loose stopwords like `example`/`placeholder`).
- Files: `.githooks/pre-commit`, `.gitleaks.toml`.

### 6. CI: test + lint workflows

- Sync commits: `21962eb` (`ci: add test + lint workflows with an honest coverage badge`), `0da5055` (documents the `[skip ci]` commit-token trap).
- Volt status: `.github/workflows/` has only deploy templates (`deploy.yml.gotmpl`, `test-deploy.yml`) — no `test.yml`, no `lint.yml`. Sync's are 110-line workflows covering the go.work monorepo with a coverage badge.

### 7. `crypt-rekey` CLI command

- Sync commit: `2b3b78d` — re-encrypts at-rest secrets when rotating the environment encryption key, with tests (`apps/sync/cmd/sync/cli_crypt_rekey.go`).
- Volt status: absent. Generic boilerplate value: without it, rotating `core.encryptionKey` orphans every encrypted row. Pairs naturally with item 2.

### 8. DB-connection-exhaustion postmortem doc

- Sync commit: `d1b2681` — `docs/plans/db-pool-connection-exhaustion.md` (the incident behind item 1: client-side exhaustion, EADDRNOTAVAIL, why idle==open is the cure).
- Volt status: `volt/docs/incidents/` has only the cron-graceful-shutdown postmortem. Port the doc alongside item 1 so the constants carry their story.

### 9. Per-environment OTEL service names (HyperDX env separation)

- Sync commits/docs: `1d02ed6` + `657ff8f` — ADR-OTEL-001 (`docs/plans/otel-service-naming.md`), `gotel.EnvServiceName` (`modules/lace/gotel/gotel.go:205`).
- The problem: HyperDX has no multi-tenant separation, so staging and prod telemetry land in one pile. Sync's decision (operator-reviewed, reversing the textbook shared-name convention): every environment emits its OWN `service.name` — prod `sync-api`, staging `sync-api-stag`, local `sync` — set via `OTEL_SERVICE_NAME`, resolved with SDK precedence (env > `OTEL_RESOURCE_ATTRIBUTES` service.name > config fallback), with `deployment.environment` kept as a secondary attribute. Prod alert rules then structurally cannot match stag noise, and one env's telemetry can be dropped/retained wholesale.
- Volt status: missing in three ways. (1) Default PKL emits `serviceName = core.appKey` for every environment and there is no staging PKL env — two deployments would mix under one HyperDX service name. (2) `core/pkg/volt/app.go` passes `otelkit.WithEnvOverrides(false)`, so the `OTEL_SERVICE_NAME` env var is ignored at the SDK level; only the prod PKL's `read("env:OTEL_SERVICE_NAME")` rescues it. (3) otelkit supports `deployment.environment.name` via `Config.Environment` / `WithEnvironment`, but volt never sets it even though `Core.Environment` is in config.
- Fix shape: adopt the per-env naming convention (port the ADR), wire `Environment: Core.Environment` into `otelkit.Config`, and either enable env overrides or standardize on PKL-read `OTEL_SERVICE_NAME` across all non-local envs.

## P3 — nice-to-have / judgment calls

### 10. `lace/util` helper additions

- Sync's `modules/lace/util/util.go` grew generic helpers volt lacks: `ISO`/`ISOPtr` (RFC3339 rendering for string-typed GraphQL time fields), `StrPtr` (empty→nil), `Deref[T]` (pointer-with-default, replaces per-type deref helpers). Additive, no behavior change to existing funcs.

### 11. Boot-preflight pattern (fail fast on broken credentials)

- Sync commits: `aef6df2` (boot preflight catches a broken CDN purge credential), `46e8c65` (boot-time storage preflight). The implementations are sync-module-specific, but the pattern — verify external credentials/buckets at boot and refuse to start, rather than failing on first use — is worth a volt equivalent for its S3/storage config. Optional.

### 12. `/api/i/bootinfo` internal endpoint

- Volt's ubgo migration removed the internal-key-guarded bootinfo endpoint from `core/pkg/coreapi/coreapi.go` (sync still has it). Likely intentional — volt's `health.Registry` + `/healthz` family covers the readiness half — but noted here so the drop stays a decision, not an accident.

## Explicitly excluded (checked, not applicable to volt)

- **OAuth token-at-rest encryption / TokenCodec** (`6762262`): fixes sync's `authmgr` module, which volt doesn't ship; volt has no OAuth code. Becomes relevant only if authmgr is ever ported.
- **kvstore self-healing Set upsert** (`ff9106f`), **wfrun/jobq failed-run reporting** (`ea20b07`), **filemgr mirror fixes** (`3965cdf`, `29250bf`), **hatchet pin** (`8b19c3b`): all in sync-only modules.
- **gotel GetTracerProvider** (`87e9650`): superseded by volt's `otelkit` stack. (The companion gozap TraceId fix `57a90c0` is NOT excluded — see item 4: volt's `applog.FromCtx` re-introduces the symptom.)
- **gqlgin / gocron logger diffs**: file-layout reordering (sync) and ubgo rewrites (volt) — no behavioral fix on either side.
- **Deploy/devbox/prod fixes** (`f9154a8`, `13f62b8`, `6f39592`, …): specific to sync's Hetzner/NetBird/staging topology.
