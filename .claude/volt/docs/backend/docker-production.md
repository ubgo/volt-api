# Docker Production Setup

## Overview

Production Docker setup for wildai\_githubapp backend Go API.

## File Locations

| File | Purpose |
|------|---------|
| `.docker/Dockerfile.prod` | Production multi-stage Dockerfile |
| `.docker/Taskfile.yml` | Task commands for build/run/stop |
| `config/pkl/prod/app.pkl` | Production PKL config (reads from env vars) |
| `config/pkl/schema/AppConfig.pkl` | PKL schema (base config definition) |
| `env.prod.sample` | Sample environment variables file |
| `.env.prod` | Actual production env file (not in git) |
| `config/main.go` | Config loader (supports CONFIG\_ENV) |

***

## Dockerfile.prod Structure

### Stage 1: Builder

```dockerfile
FROM golang:${GO_VERSION}-alpine AS builder
```

* Uses Go Alpine image
* `GOTOOLCHAIN=auto` - automatically downloads Go version required by go.work (1.24.11)
* Installs pkl CLI for any build-time config operations
* Copies go.work, go.mod/go.sum files
* Copies source code: config, apidash, lace, saas, wildaiadminapi, wildaipartnerapi, config
* Runs `go work vendor` to download dependencies (not copied from local)
* Builds binary: `go build -mod=vendor -ldflags="-w -s" -o /build/bin/api ./saas/cmd/api`

### Stage 2: Runtime

```dockerfile
FROM alpine:3.20
```

* Minimal Alpine image
* Installs: ca-certificates, tzdata, curl, gcompat, libstdc++
* **gcompat** - Required for pkl binary (glibc compatibility layer for Alpine)
* Installs pkl CLI for runtime config evaluation
* Creates non-root user `appuser:appgroup` (UID/GID 1000)
* Creates directories: `/app/tmp/logs`, `/app/_keys`
* Copies only prod config: `config/pkl/schema/` and `config/pkl/prod/`
* Runs as non-root user

### Key Environment Variables

```dockerfile
ENV CONFIG_ENV=prod    # Which pkl config to load (default, prod)
ENV PORT=2315          # Default port (overridden by .env.prod)
```

### Health Check

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${PORT}/ || exit 1
```

***

## PKL Configuration System

### How it works

1. `config/main.go` reads `CONFIG_ENV` environment variable
2. Loads config from `config/pkl/${CONFIG_ENV}/app.pkl`
3. Production config (`prod/app.pkl`) reads all values from environment variables using `read("env:VAR_NAME")`

### Production PKL Config Pattern

```pkl
amends "../schema/AppConfig.pkl"

appName = "wildai_githubapp_prod"
appKey = read("env:APP_KEY")
port = read("env:PORT")
environment = "prod"

database {
  host = read("env:DB_HOST")
  port = read("env:DB_PORT").toInt()
  // ... etc
}
```

### Important PKL Notes

* `read("env:VAR")` - reads environment variable as string
* `.toInt()` - converts string to integer
* `.toBoolean()` - converts string to boolean
* `read?("env:VAR")` - optional read (returns null if not set)
* **No `.toUInt16()` method** - use `.toInt()` instead

***

## Taskfile Commands

Located in `.docker/Taskfile.yml`:

| Command | Description |
|---------|-------------|
| `task prod:build` | Build production image with version info |
| `task prod:build:nocache` | Build without Docker cache |
| `task prod:run` | Run container (auto-removes old, reads PORT from .env.prod) |
| `task prod:stop` | Stop and remove container |
| `task prod:logs` | View container logs (supervisor) |
| `task prod:logs:api` | View API logs from supervisor |
| `task prod:logs:cron` | View Cron logs from supervisor |
| `task prod:status` | Show supervisor process status |
| `task prod:cli -- <cmd>` | Run CLI command in container |
| `task prod:shell` | Shell access to container |

### Custom Variables

```bash
# Custom image name/tag
task prod:build IMAGE_NAME=myregistry/app IMAGE_TAG=v1.0.0

# Custom volume paths
task prod:run KEYS_DIR=/path/to/keys LOGS_DIR=/path/to/logs
```

### How prod:run Works

1. Reads PORT from `.env.prod` using: `grep -E '^PORT=' ../.env.prod | cut -d '=' -f2`
2. Creates host directories for volume mounts
3. Runs container with:
   * Port mapping: `-p ${PORT}:${PORT}` (for local testing)
   * Env file: `--env-file ../.env.prod`
   * Volume mounts for keys (read-only) and logs

> **Note for VPS deployment:** Use `--network host` instead of port mapping. See "Network and Permissions" section below for details.

***

## Volume Mounts

| Host Path | Container Path | Mode | Purpose |
|-----------|----------------|------|---------|
| `_keys/` | `/app/_keys` | read-only | GitHub App private key, etc. |
| `/tmp/wildai-githubapp/logs/` | `/app/tmp/logs` | read-write | Application logs |

***

## Required Environment Variables

See `env.prod.sample` for full list. Key categories:

### Application

* `CONFIG_ENV=prod`
* `PORT` - API port
* `APP_KEY`, `HOSTNAME`, `INTERNAL_KEY`
* `JWT_SECRET`, `ENCRYPTION_KEY`

### Database

* `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`

### Redis

* `REDIS_URL`, `REDIS_DB`

### NATS (optional)

* `NATS_STATUS` (true/false), `NATS_APP_KEY`, `NATS_URL`, `NATS_CREDS_FILE`

### GitHub App

* `GITHUB_APP_ID`, `GITHUB_APP_NAME`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
* `GITHUB_PRIVATE_KEY_PATH` (path inside container: `/app/_keys/...`)
* `GITHUB_WEBHOOK_SECRET`, `GITHUB_CALLBACK_URL`

### OpenTelemetry

* `OTEL_ENABLE`, `OTEL_ENDPOINT`, `OTEL_API_KEY`
* Various `OTEL_ENABLE_*` flags

### Logger

* `LOGGER_ENABLE_STDOUT` - set to `true` to see logs in `docker logs`
* `LOGGER_ENABLE_OTEL`, `LOGGER_ENABLE_FILE`

***

## Build & Deploy

### Local Testing

```bash
cd .docker
task prod:build
task prod:run
task prod:logs
```

### Rebuild and Run

```bash
task prod:stop && task prod:build && task prod:run
```

### GitHub Actions Deployment

1. Store secrets in GitHub Secrets:
   * `GITHUB_APP_PRIVATE_KEY` - the PEM file content
   * All other env vars

2. Workflow creates key file at deploy time:

```yaml
- name: Create keys
  run: |
    mkdir -p _keys
    echo "${{ secrets.GITHUB_APP_PRIVATE_KEY }}" > _keys/github-app.pem

- name: Run container
  run: |
    docker run -d \
      --network host \
      -v $(pwd)/_keys:/app/_keys:ro \
      --env-file .env.prod \
      wildai-githubapp:latest
```

> **Note:** `--network host` is required for PKL to access package metadata. See [GitHub Actions Deployment](./github-actions-deployment.md) for the full workflow.

***

## Troubleshooting

### No logs showing

* Check `LOGGER_ENABLE_STDOUT=true` in .env.prod

### pkl binary "not found" error

* This means glibc compatibility issue
* Ensure `gcompat` and `libstdc++` are installed in runtime image

### Container exits immediately

* Check `docker logs wildai-githubapp` for panic messages
* Common issues: missing env vars, database not reachable

### PKL network/DNS error

* Error: `Error connecting to host 'pkg.pkl-lang.org'`
* PKL needs network access at startup to fetch package metadata
* **Solution:** Use `--network host` instead of bridge network mode
* Bridge networking can cause DNS resolution failures on some VPS configurations

### Go version mismatch

* `GOTOOLCHAIN=auto` in Dockerfile allows Go to download required version
* Project requires Go 1.24.11 (specified in go.work)

***

## Network and Permissions (VPS Deployment)

### Why `--network host`?

PKL (Apple's configuration language) needs network access to fetch package metadata from `pkg.pkl-lang.org` at startup. Docker's default bridge network mode can cause DNS resolution failures on some VPS configurations.

**Bridge mode (default):**

```bash
docker run -p 2315:2315 ...  # Container has isolated network, needs port mapping
```

**Host mode (required for this app):**

```bash
docker run --network host ...  # Container shares host's network stack
```

With `--network host`:

* No port mapping (`-p`) needed - app binds directly to host's port
* DNS resolution works reliably
* Container uses the PORT from `.env.prod` directly

### Why `chown 1000:1000` on logs directory?

The container runs as non-root user `appuser` (UID/GID 1000) for security. When directories are created on the host (e.g., by SSH as root during deployment), they're owned by root.

```bash
# Problem: root-owned directory
mkdir -p /khanakia/wildai_githubapp/logs
# Container's appuser (UID 1000) can't write here!

# Solution: fix ownership
chown -R 1000:1000 /khanakia/wildai_githubapp/logs
# Now appuser can write logs
```

The `_keys/` directory doesn't need this fix because it's mounted read-only (`:ro`).

***

## Architecture Notes

* **Multi-stage build** keeps image small (~250MB)
* **Supervisor** runs as non-root user (appuser), managing API and Cron processes
* **Four binaries**: `/app/api`, `/app/cron`, `/app/cli`, `/app/dbmigrate`
* **Non-root user** for security (appuser:appgroup, UID/GID 1000)
* **gcompat layer** needed because pkl binary is glibc-linked but Alpine uses musl
* **PKL** is Apple's configuration language - configs are evaluated at runtime
* **Vendor mode** - dependencies downloaded via `go work vendor` during build (ensures fresh deps)

## Multi-Process Architecture

The container runs supervisor which manages:

| Process | Binary | Purpose |
|---------|--------|---------|
| api | `/app/api` | HTTP API server (GraphQL) |
| cron | `/app/cron` | Background job scheduler |
| cli | `/app/cli` | Manual commands (not auto-started) |
| dbmigrate | `/app/dbmigrate` | Database migrations (run during deploy) |

Supervisor config: `/etc/supervisor/conf.d/supervisord.conf`

Logs location inside container (owned by appuser):

* `/app/tmp/logs/api.out.log`
* `/app/tmp/logs/cron.out.log`
* `/app/tmp/logs/supervisord.log`

## Version Info

Two ways to check the deployed version:

### HTTP Endpoint

```bash
curl http://localhost:2315/api/version
```

```json
{
  "version": "1.0.0",
  "commit": "481a523",
  "branch": "main",
  "build_time": "2026-01-09T07:11:53Z"
}
```

### CLI Command

```bash
task prod:cli -- version
```

```
Version:    1.0.0
Commit:     481a523
Branch:     main
Build Time: 2026-01-09T07:11:53Z
```

### Setting Version at Build Time

```bash
task prod:build VERSION=1.0.0
```

Commit, branch, and build\_time are auto-detected from git.

***

## Related Files

* [GitHub Actions Deployment](./github-actions-deployment.md) - CI/CD pipeline setup
* [env.prod.sample](../../../env.prod.sample) - Environment variable template
* [.docker/Taskfile.yml](../../../.docker/Taskfile.yml) - Local Docker commands
