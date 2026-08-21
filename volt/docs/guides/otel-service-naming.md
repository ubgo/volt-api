# OTel service naming — one service name per environment

Decides what `service.name` each deployment of a Volt app emits to the observability backend (HyperDX, Grafana, etc.). Ported from an operator-reviewed decision on a Volt-derived production app (ADR-OTEL-001, 2026-08-18), where it deliberately REVERSED the earlier textbook convention.

## The problem

Backends like self-hosted HyperDX have no tenant separation: staging and prod telemetry land in one pile. The OTel textbook answer is one shared `service.name` split by a `deployment.environment` attribute — dashboards stay singular, environments are a filter. Operator review rejected that on operational grounds:

- Each environment is a fully separate deployment; filtering `ServiceName:myapp-api` should mean PROD and nothing else.
- Per-service data lifecycle (drop/retain one env's telemetry wholesale) is easier when each env IS its own service.
- Alert rules scoped to the prod service name structurally cannot fire from staging/dev noise — no per-rule environment exclusions to maintain.

The textbook counterarguments (dashboard duplication per env, no one-click cross-env comparison, service-map fan-out) were weighed and accepted as costs.

## The convention

Every environment emits its own `service.name`:

| Environment | service.name | Where it's set |
|---|---|---|
| prod | `myapp-api` | env `OTEL_SERVICE_NAME` (read by `pkl/env/prod/app.pkl`) |
| staging | `myapp-api-stag` | same env var, in the staging environment file |
| local dev | PKL default (`core.appKey`) | `pkl/env/default/app.pkl` fallback — already distinct, no env var needed |

`deployment.environment.name` is still emitted as a secondary attribute (wired from `core.environment` into `otelkit.Config.Environment` in `core/pkg/volt/app.go` — harmless, occasionally useful), but the SERVICE NAME is the primary environment discriminator.

## Mechanics in Volt

- `otel.serviceName` in `pkl/env/default/app.pkl` and `sample/app.pkl` is `read?("env:OTEL_SERVICE_NAME") ?? <fallback>`; `prod/app.pkl` reads the env var unconditionally. Set the var per deployed environment — no code change needed to rename a service.
- Note: `core/pkg/volt/app.go` passes `otelkit.WithEnvOverrides(false)` — PKL is the single authority for OTEL config, so `OTEL_SERVICE_NAME` only takes effect through the PKL `read?()` above, not via the SDK's own env handling. Don't set other `OTEL_*` SDK vars and expect them to apply.

## Consequences

- The backend's ServiceName facet lists each env separately; `ServiceName:myapp-api` = prod only.
- Alerts are scoped by service name — prod rules structurally cannot match staging/dev.
- Dashboards/saved searches are per-env by construction (accepted cost).
