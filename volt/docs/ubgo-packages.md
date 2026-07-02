# The ubgo package ecosystem

Volt API is built on **[ubgo](https://github.com/orgs/ubgo/repositories)** — a family of focused, zero-/low-dependency Go libraries (many with a zero-dep core and isolated `contrib/` adapters). Some are wired into the boilerplate out of the box; the rest pair cleanly with it when you need them. All share the same design philosophy: small interfaces, swappable backends, OpenTelemetry-aware, 100%-tested cores.

## Wired in (used by the boilerplate)

These ship configured — you're already using them.

| Package | What it does | Where it's used here |
|---|---|---|
| [`logger`](https://github.com/ubgo/logger) | Adapter-based, slog-native, zero-alloc structured logging — sampling, redaction, OTel, FingersCrossed debug-on-error, rotation. | Everywhere, via `core/pkg/applog` (aliased `ulog`). `applog.FromCtx(ctx)` adds trace correlation. See [guides/observability.md](guides/observability.md). |
| [`otelkit`](https://github.com/ubgo/otelkit) | One-constructor OpenTelemetry bootstrap (traces, metrics, logs) with vendor presets (HyperDX, Grafana, Honeycomb, Datadog, New Relic). | Telemetry bootstrap in `core/pkg/plugin`; `applog`'s OTel sink reads the provider it registers. |
| [`errx`](https://github.com/ubgo/errx) | One structured error: stable identity, class, redactable fields, return-trace, wire codec. Replaces the old `goerr`. | Error type + explicit `ReportError` sink in the app bootstrap. |
| [`cache`](https://github.com/ubgo/cache) + [`cache-redis`](https://github.com/ubgo/cache-redis) | Bytes-level cache interface with generics, single-flight `Remember` (SWR/stale-if-error/negative/jitter), distributed locker, pluggable codecs; Redis adapter with SCAN prefix ops + Pub/Sub invalidation. | Caching layer, Redis-backed. |
| [`health`](https://github.com/ubgo/health) (+ `health-gin`, `health-redis`) | k8s-style health/readiness registry with framework adapters and concrete checkers (Postgres, Redis, NATS, HTTP, DNS). | `/api/i/bootinfo` and readiness probes; Gin + Redis contribs. |
| [`shutdown`](https://github.com/ubgo/shutdown) | Phased, parallel-within-phase, observable graceful shutdown — watchdog hard-exit, force-exit on second signal. | Graceful shutdown in the api/cron entrypoints. |
| [`buildinfo`](https://github.com/ubgo/buildinfo) | Build metadata (version, commit, time, modules) auto-populated from `runtime/debug` with `-ldflags` overrides; OTel/slog adapters. | Version reporting; OTel resource attributes. |
| [`crypt`](https://github.com/ubgo/crypt) | AES-CBC + PKCS#7 with hex output — for storing short secrets at rest. | Encrypting sensitive values (see [`../rules/secrets.md`](../rules/secrets.md)). |
| [`goutil`](https://github.com/ubgo/goutil) | Small set of Go utility helpers. | Misc helpers across modules. |

## Available to add (pair well with the boilerplate)

Not wired in, but designed to drop in. Add with `go get`, then follow the package README.

| Package | Use it for |
|---|---|
| [`lock`](https://github.com/ubgo/lock) | Distributed locks behind one interface — filelock, flock, Redis, Postgres, etcd — with fencing tokens, semaphore mode, TraceID propagation. Swap backends in one line. |
| [`gqlkit`](https://github.com/ubgo/gqlkit) | Generate fully-typed GraphQL **client** SDKs (Go + TypeScript) with a builder pattern — no query-first codegen loop. (Volt Web uses this to call this API.) |
| [`entx`](https://github.com/ubgo/entx) | Ent ORM extensions: polymorphic relationships (`MorphTo`/`MorphMany`), cascade-delete generation, per-API GraphQL schema splitting for gqlgen. |
| [`entkit`](https://github.com/ubgo/entkit) | Ent + gqlgen column types and field helpers (multi-module umbrella). |
| [`publicid`](https://github.com/ubgo/publicid) | URL-safe, non-sequential public IDs on a fixed `[0-9a-z]` nanoid alphabet. |
| [`jsontype`](https://github.com/ubgo/jsontype) | A JSON type that round-trips cleanly between Go, JSON, and SQL drivers (Postgres `jsonb`, MySQL `json`, SQLite `TEXT`). |
| [`jsonslice`](https://github.com/ubgo/jsonslice) | JSON-backed slice that round-trips through SQL columns and gqlgen scalars — zero deps. |
| [`httpreq`](https://github.com/ubgo/httpreq) | JSON-over-HTTP convenience on `net/http` — stdlib-only, options pattern. |
| [`threadsafecache`](https://github.com/ubgo/threadsafecache) | Generic thread-safe LRU + TTL cache, one dependency (lighter than the `cache` family). |
| [`gots`](https://github.com/ubgo/gots) / [`gouuid`](https://github.com/ubgo/gouuid) | Go→TypeScript codegen (types/enums/constants) · `google/uuid` helpers. |

### The `cache` family

`cache` is the interface; pick backends/composers as needed — [`cache-mem`](https://github.com/ubgo/cache-mem) (sharded W-TinyLFU), [`cache-pg`](https://github.com/ubgo/cache-pg) (Postgres/SQLite), [`cache-memcached`](https://github.com/ubgo/cache-memcached), [`cache-tiered`](https://github.com/ubgo/cache-tiered) (L1/L2/L3), [`cache-cluster`](https://github.com/ubgo/cache-cluster) (peer-aware), [`cache-obj`](https://github.com/ubgo/cache-obj) (live-object, zero-serialization), plus [`cache-cli`](https://github.com/ubgo/cache-cli) to inspect and [`cache-examples`](https://github.com/ubgo/cache-examples) to learn from. All implement the same conformance suite, so backends swap without touching call sites.

## Conventions when adding an ubgo package

- Prefer the ubgo package over hand-rolling or a heavier third-party dep — they follow this repo's conventions (context-first, OTel-aware, options pattern).
- Wire long-lived clients into the app `*Plugin` (see [architecture.md](architecture.md)) and inject them; don't reach for globals.
- Most have a zero-dep core plus `contrib/` adapters — import only the adapter you need.
- Retired: **`goerr`** is superseded by **`errx`**; **`gozap`** by **`logger`**. Don't reintroduce the old ones.
