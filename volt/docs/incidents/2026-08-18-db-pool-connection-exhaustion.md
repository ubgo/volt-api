# Postmortem — client-side DB connection exhaustion (EADDRNOTAVAIL)

> Incident from a Volt-derived production app (the sync fork), 2026-08-18, sev2, ~40m + wedged jobs. Ported here because the root cause lived in shared Volt code (`lace/db`) and the fix now ships in the boilerplate (`modules/lace/db/config.go` + `db.go` + `db_test.go`).

## Summary

Firing 9 concurrent data-processing jobs made the worker's dials to pgbouncer fail with `connect: cannot assign requested address` (EADDRNOTAVAIL). Two jobs errored at start; six more wedged mid-stream (goroutines stuck on dead connections, job rows frozen in `running`). pgbouncer, Postgres, and conntrack were all healthy — the failure was entirely inside the app container's kernel: no free local ports.

## Root cause

Go's `database/sql` ships `MaxOpenConns = unlimited` and `MaxIdleConns = 2`, and no code anywhere called the `Set*` tuning methods. The two knobs are independent, and this combination is the worst possible one:

| Knob | Question it answers | Default | Effect |
|---|---|---|---|
| `MaxOpenConns` | may I open another connection? | unlimited | no cap → **no queueing, ever** — every concurrent query dials a NEW connection |
| `MaxIdleConns` | how many may I KEEP after use? | 2 | everything beyond 2 is **closed immediately** after the query finishes |

Under the burst: query arrives → dial new conn (no cap) → query finishes → pool already holds 2 idle → close the conn → the local port enters TIME_WAIT for 60s → next query dials again… Thousands of open/close cycles per minute, each one poisoning a port toward the single (container-IP → pgbouncer:6432) tuple, whose usable range is ~28k ports. The range exhausted inside the TIME_WAIT window and `connect()` started failing in the kernel — **before any packet reached pgbouncer**.

Why latency never rose first: graceful degradation (queueing) is a property of a CAPPED pool. With unlimited opens nothing ever waits in line — connections are free right up until the instant they are impossible.

## Why pgbouncer did not help

pgbouncer pools the OTHER leg of the wire:

```mermaid
flowchart LR
    A[Go process<br/>leg A: client conns<br/>OUR pool's job] --> B[pgbouncer<br/>max_client_conn 1000]
    B --> C[Postgres<br/>leg B: 20 server conns<br/>PGBOUNCER's job]
```

Leg B worked flawlessly throughout (zero pgbouncer log entries, cap never approached). Leg A — how our process opens/closes TCP sockets TOWARD pgbouncer — is entirely the client's responsibility. If anything, pgbouncer's cheap accepts make client-side churn feel free, which is exactly why nobody notices until the kernel's port math runs out.

## Diagnosis trail (what ruled things out)

- pgbouncer logs: clean; `max_client_conn=1000`, `default_pool_size=20`, transaction mode.
- Host conntrack: 236 of 262,144 — nowhere near full.
- Container steady-state sockets: ~30 ESTABLISHED — the damage was DEAD sockets (TIME_WAIT), invisible after 60s, which is why post-hoc inspection looked innocent.
- Wedged jobs: `pg_stat_activity` showed 0 active queries and 5 idle conns while 6 jobs claimed `running` — their goroutines died on dead conns without finalizing the job rows (no stale-run reconciler existed; rows were failed manually).

## Fix

`lace/db.New` (the single constructor every process's shared pool flows through) now applies churn-safe sizing — `modules/lace/db/config.go` + `db.go`:

- `MaxOpenConns = 30` — a HARD valve in the stdlib: the 31st concurrent query WAITS in a queue for a freed connection; nothing can dial past the cap.
- `MaxIdleConns = 30` — **idle == open is the load-bearing invariant**: any gap converts finished queries into close+reopen churn. Held connections are reused, never closed, so TIME_WAIT never accumulates.
- `ConnMaxLifetime = 30m`, `ConnMaxIdleTime = 5m` — gentle recycling for pgbouncer/Postgres restarts.
- Option funcs (`WithMaxOpenConns` …) allow deliberate per-env retuning.
- `modules/lace/db/db_test.go` pins defaults-applied + the idle==open invariant.

Sizing rationale: 30 per process × 5 processes = 150 client conns, far under pgbouncer's 1000; pgbouncer multiplexes them onto 20 real Postgres conns, so >20-wide client parallelism buys nothing anyway — the cap costs no throughput.

## How over-demand behaves now

100 concurrent queries → 30 run, 70 wait in the stdlib's queue (typically milliseconds each, bounded by each caller's context timeout). Load shows up as query latency, never as socket growth: a 10× day lengthens the queue, not the port table. Verified on the origin app by re-running the identical 9-job burst post-fix: all 9 succeeded, the container peaked at 46 sockets, TIME_WAIT ≈ 0 throughout.

## The rules that keep this fixed

1. Get your DB from the kernel's shared pool (`app.New()` → Ent client) — a rogue `sql.Open` elsewhere escapes the cap; if one is ever justified, it must set the same knobs (small secondary pools can size down, e.g. 4/4, but idle must equal open).
2. Keep `MaxIdleConns == MaxOpenConns` — the unit test enforces it; lowering idle reintroduces churn.
3. Retuning capacity is a PAIRED change: pgbouncer `pool_size` and our `MaxOpenConns` move together.
