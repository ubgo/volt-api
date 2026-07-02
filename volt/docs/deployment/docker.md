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

> **Note:** `--network host` is required for PKL to access package metadata. See [GitHub Actions Deployment](./github-actions.md) for the full workflow.

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

* [GitHub Actions Deployment](./github-actions.md) - CI/CD pipeline setup
* [env.prod.sample](../../../env.prod.sample) - Environment variable template
* [.docker/Taskfile.yml](../../../.docker/Taskfile.yml) - Local Docker commands


## Persistent storage


How this project handles persistent data (uploads, logs, keys) across Docker container recreations using external storage mounts.

## Problem

Docker containers are ephemeral. Any data written inside the container (e.g., `/app/uploads/`) is lost when the container is recreated during deployments. For user-uploaded files, this is unacceptable.

## Solution

Mount host directories (including external storage like SMB/CIFS shares) into the container as Docker volumes. The deploy workflow reads mount paths from the `.env.{env}` file and conditionally attaches them to `docker run`.

## Architecture

```
VPS Host
├── /khanakia/wildai_chatapi/          ← DEPLOY_PATH (GitHub secret)
│   ├── .env.prod                       ← Decoded from ENV_B64 at deploy time
│   ├── _keys/                          ← Private keys (mounted read-only)
│   └── logs/                           ← Application logs
│
├── /mnt/hetzner_u417192/backup/        ← External SMB/CIFS mount
│   └── wildaichatgo/
│       └── uploads/                    ← Persistent uploads directory
│           ├── chatbots/
│           └── itemgroups/
│
└── Docker Container (wildai-chatapi)
    ├── /app/_keys      → DEPLOY_PATH/_keys (read-only)
    ├── /app/tmp/logs   → DEPLOY_PATH/logs
    └── /app/uploads    → UPLOADS_PATH from .env (external storage)
```

## Container User & Permissions

The production Dockerfile creates a non-root user for security:

```dockerfile
# .docker/Dockerfile.prod
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser
# ...
USER appuser
```

| Property | Value |
|----------|-------|
| User | `appuser` |
| UID | `1000` |
| Group | `appgroup` |
| GID | `1000` |

All mounted directories must be owned by UID 1000 / GID 1000 for the container to read/write them.

### Verifying Permissions

```bash
# From host — check numeric ownership
ls -ln /path/to/mounted/dir

# From inside container — check user identity
docker exec <container> id
# uid=1000(appuser) gid=1000(appgroup)

# Write test from inside container
docker exec <container> touch /app/uploads/.write-test
docker exec <container> rm /app/uploads/.write-test
```

## How UID/GID Permissions Actually Work

UIDs are just numbers stored on files. Usernames are just labels mapped by each system's `/etc/passwd`.

### Each system has its own name lookup table

```
Host /etc/passwd:              Container /etc/passwd:
  root    → UID 0                root    → UID 0
  ana     → UID 1000             appuser → UID 1000
```

The disk only stores the **number**, never the name. When you run `ls -la`, the OS looks up the number in its own table to display a name:

```
WHO WRITES              STORED ON DISK     HOST SEES            CONTAINER SEES
────────────            ──────────────     ──────────           ──────────────
root (UID 0)            owner=0            root:root            root:root
appuser (UID 1000)      owner=1000         ana:cpaneldemo       appuser:appgroup
SMB mount (forced)      forced to 1000     ana:cpaneldemo       appuser:appgroup
```

All three columns on the right are the **same file** — just different name labels for the same UID.

### Why root on the host doesn't affect container permissions

You SSH as root (UID 0), but root doesn't write application files. The Docker process chain looks like this:

```
You (root, UID 0)
  └── Docker daemon (runs as root)
        └── Container process (runs as UID 1000 — set by USER appuser in Dockerfile)
              └── Writes files → stored as owner=1000, group=1000
```

The permission check is purely numeric: "Is the file owned by UID 1000? Is the process running as UID 1000? Access granted." The human-readable names (`ana`, `appuser`, `root`) are irrelevant.

### SMB mounts force ownership

The fstab `uid=1000,gid=1000` option tells Linux: "present every file on this mount as owned by UID 1000, regardless of what's on the remote server." This means:

- `chown` has no effect (the mount overrides it)
- All files are always accessible by UID 1000 (the container user)
- `ls -la` shows the host's name for UID 1000 (e.g., `ana`), not `appuser`

Use `ls -ln` (numeric) to see the truth:

```bash
ls -la /mnt/storage/uploads/   # Shows: ana:cpaneldemo (host label)
ls -ln /mnt/storage/uploads/   # Shows: 1000:1000      (actual UID/GID)
```

## Volume Mounts

### Current Mounts

| Container Path | Host Source | Purpose | Mode |
|----------------|------------|---------|------|
| `/app/_keys` | `${DEPLOY_PATH}/_keys` | Private keys (SSH, NATS creds) | read-only |
| `/app/tmp/logs` | `${DEPLOY_PATH}/logs` | Application log files | read-write |
| `/app/uploads` | `${UPLOADS_PATH}` from `.env` | User-uploaded files | read-write |

### How Mounts Are Configured

Mounts are **not** hardcoded in the deploy workflow. Instead:

1. **Fixed mounts** (`_keys`, `logs`) use `DEPLOY_PATH` (a GitHub secret)
2. **Configurable mounts** (`uploads`) are read from the `.env.{env}` file at deploy time

This avoids GitHub Actions secret masking issues (see Lessons Learned below).

## Adding a New Persistent Volume

### Step 1: Add to `.env.{env}`

```bash
# .env.prod (or .env.stag)
# NOT prefixed with GITHUB_SECRET_ — it stays in the env file, not as a separate secret
MY_DATA_PATH=/mnt/some/external/path/data
```

### Step 2: Update deploy template

In `.github/workflows/deploy.yml.gotmpl`, add three things:

**a) Read the path from env file (after env file is written):**
```bash
echo "==> Reading config from .env.${ENVIRONMENT}..."
PORT=$(grep -E '^PORT=' ${DEPLOY_PATH}/.env.${ENVIRONMENT} | cut -d '=' -f2 | tr -d ' ' || echo "2315")
UPLOADS_PATH=$(grep -E '^UPLOADS_PATH=' ${DEPLOY_PATH}/.env.${ENVIRONMENT} | cut -d '=' -f2 | tr -d ' ' || echo "")
MY_DATA_PATH=$(grep -E '^MY_DATA_PATH=' ${DEPLOY_PATH}/.env.${ENVIRONMENT} | cut -d '=' -f2 | tr -d ' ' || echo "")
```

**b) Create directory and fix permissions:**
```bash
if [ -n "${MY_DATA_PATH}" ]; then
  mkdir -p "${MY_DATA_PATH}"
  chown -R 1000:1000 "${MY_DATA_PATH}"
fi
```

**c) Add volume to docker run command:**
```bash
if [ -n "${MY_DATA_PATH}" ]; then
  DOCKER_CMD="${DOCKER_CMD} -v ${MY_DATA_PATH}:/app/data"
fi
```

### Step 3: Regenerate workflows

```bash
cd scripts && go run . gh:secrets:update
# This regenerates deploy.prod.yml and deploy.stag.yml from the template

cd scripts && go run . gh:secrets:push --env=prod
# This pushes updated ENV_B64 (which now includes MY_DATA_PATH) to GitHub
```

### Step 4: Update PKL config if the app needs to know the path

```pkl
# config/pkl/env/prod/app.pkl
core {
  uploadsdir = rootdir + "/uploads"    # Maps to /app/uploads inside container
  datadir = rootdir + "/data"          # Maps to /app/data inside container
}
```

The app always references `/app/data` — it doesn't know or care about the host path.

## External Storage (SMB/CIFS) Setup

### Hetzner Storage Box Mount

The VPS mounts a Hetzner Storage Box via CIFS/SMB:

```bash
# /etc/fstab
//u417192.your-storagebox.de/backup  /mnt/hetzner_u417192/backup  cifs  credentials=/etc/hetzner_u417192.credentials,uid=1000,gid=1000,iocharset=utf8,vers=3.0,_netdev  0  0
```

Key fstab options:

| Option | Purpose |
|--------|---------|
| `uid=1000` | All files appear owned by UID 1000 (matches container `appuser`) |
| `gid=1000` | All files appear owned by GID 1000 (matches container `appgroup`) |
| `_netdev` | Wait for network before mounting |
| `vers=3.0` | SMB protocol version |
| `credentials=...` | Credential file (not inline password) |

### SMB Permission Gotchas

**`chown` has no effect on SMB mounts.** The mount options (`uid=`, `gid=`) determine ownership for all files, regardless of `chown` calls. The deploy script's `chown -R 1000:1000` is a no-op on SMB but harmless — it succeeds silently.

**`ls -la` shows host username, not UID.** On the host, you'll see `ana:cpaneldemo` (whatever user has UID 1000 on the host). Inside the container, you'll see `appuser:appgroup`. They're the same UID/GID — just different name mappings.

```bash
# Host shows:    ana:cpaneldemo  (host's UID 1000 mapping)
# Container shows: appuser:appgroup (container's UID 1000 mapping)
# Numeric shows:   1000:1000       (the actual UID/GID — same in both)
```

### Setting Up a New SMB Mount

```bash
# 1. Create credentials file
cat > /etc/storage_creds <<EOF
username=your_username
password=your_password
EOF
chmod 600 /etc/storage_creds

# 2. Create mount point
mkdir -p /mnt/storage

# 3. Add to /etc/fstab (uid/gid MUST match container user)
echo '//storage.example.com/share  /mnt/storage  cifs  credentials=/etc/storage_creds,uid=1000,gid=1000,iocharset=utf8,vers=3.0,_netdev  0  0' >> /etc/fstab

# 4. Mount
mount /mnt/storage

# 5. Verify
ls -ln /mnt/storage  # Should show 1000:1000
```

## Deploy Workflow Details

The deploy template (`.github/workflows/deploy.yml.gotmpl`) handles volumes in this order:

```
1. Create directories (DEPLOY_PATH/_keys, DEPLOY_PATH/logs)
2. Fix permissions (chown 1000:1000)
3. Write key files (decoded from base64 secrets)
4. Write .env file (decoded from ENV_B64 secret)
5. Read config from .env (PORT, UPLOADS_PATH)     ← env file must exist first
6. Create uploads directory + fix permissions
7. Pull Docker image
8. Stop old container
9. Build docker run command with conditional volumes
10. Start new container via eval
11. Run database migrations
```

### Why `eval` for `docker run`

The `docker run` command is built as a string (`DOCKER_CMD`) because volume mounts are conditional. Direct variable expansion in a multiline shell command caused issues with `appleboy/ssh-action`:

```bash
# This pattern works:
DOCKER_CMD="docker run -d --name ${CONTAINER} ..."
if [ -n "${UPLOADS_PATH}" ]; then
  DOCKER_CMD="${DOCKER_CMD} -v ${UPLOADS_PATH}:/app/uploads"
fi
DOCKER_CMD="${DOCKER_CMD} ${IMAGE}"
eval ${DOCKER_CMD}
```

### CRITICAL: Never Use GitHub Secrets for Host Paths

GitHub Actions automatically masks **all** secret values everywhere in shell output and variable expansion. This isn't just cosmetic log masking — it corrupts values when they pass through `appleboy/ssh-action` to the remote shell.

**What happened when we used `GITHUB_SECRET_UPLOADS_PATH`:**

```
Step 1: GitHub secret set
   UPLOADS_PATH = /mnt/hetzner_u417192/backup/wildaichatgo/uploads

Step 2: Deploy script built the volume flag
   UPLOADS_VOLUME="-v /mnt/hetzner_u417192/backup/wildaichatgo/uploads:/app/uploads"

Step 3: appleboy/ssh-action transmitted env vars to VPS
   → GitHub's masking engine intercepted the secret value
   → The -v flag got separated from its argument
   → Docker received: " /mnt/hetzner_u417192/backup/wildaichatgo/uploads"
     (note the leading space and missing -v flag)

Step 4: Docker tried to parse it
   → Interpreted " /mnt/..." as a volume NAME (not a host path)
   → Volume names only allow [a-zA-Z0-9][a-zA-Z0-9_.-]
   → FAILED
```

**The actual error:**
```
docker: Error response from daemon: create  /mnt/hetzner_u417192/backup/wildaichatgo/uploads:
" /mnt/hetzner_u417192/backup/wildaichatgo/uploads" includes invalid characters for a local
volume name, only "[a-zA-Z0-9][a-zA-Z0-9_.-]" are allowed. If you intended to pass a host
directory, use absolute path
```

**The rule:** Any value that will be used in shell commands on the VPS (paths, ports, flags) should be a **regular env var** in `.env.{env}` — NOT a `GITHUB_SECRET_*` variable. The deploy script reads it from the decoded `.env` file on the VPS, completely bypassing GitHub's masking.

```
BROKEN:   GITHUB_SECRET_UPLOADS_PATH=/mnt/...  →  GitHub secret  →  masked/corrupted
WORKS:    UPLOADS_PATH=/mnt/...                 →  inside ENV_B64 →  decoded on VPS  →  grep reads it cleanly
```

**When to use `GITHUB_SECRET_*` prefix vs regular env var:**

| Use `GITHUB_SECRET_*` prefix | Use regular env var |
|-------------------------------|---------------------|
| Values needed by GitHub Actions steps themselves | Values only needed on the VPS |
| SSH keys, registry tokens, host/port for SSH | File paths, feature flags, app config |
| Values that must exist as individual GH secrets | Values that just need to reach the container's `.env` |

## Lessons Learned

| Issue | Root Cause | Fix |
|-------|-----------|-----|
| Uploads lost on deploy | Directory lived inside container, no volume mount | Mount external storage to `/app/uploads` |
| `docker run -v ***` error | GitHub Actions masked the secret path value | Read path from `.env` file instead of GitHub secret |
| `" /mnt/..."` invalid chars | Shell variable expansion split `-v` from path argument | Build command as string + `eval` |
| `chown` no effect on SMB | SMB mount uses fixed uid/gid from fstab | Ensure fstab `uid=1000,gid=1000` matches container user |
| `ls` shows wrong user | Host maps UID 1000 to different username | Normal — use `ls -ln` for numeric UIDs to verify |

## Migrating Uploads from Non-Dockerized Setup

When migrating from a bare-metal deployment to Docker with external storage:

```bash
# 1. Copy files preserving structure
cp -a /old/path/uploads/* /mnt/external/storage/uploads/

# 2. Fix ownership (works on local dirs, no-op on SMB with fixed uid)
chown -R 1000:1000 /mnt/external/storage/uploads/

# 3. Verify from inside the running container
docker exec <container> ls -la /app/uploads/
docker exec <container> find /app/uploads -type f | wc -l

# 4. Write test
docker exec <container> touch /app/uploads/.test && docker exec <container> rm /app/uploads/.test
```

## Quick Reference

```bash
# Check container volume mounts
docker inspect <container> --format '{{range .Mounts}}{{.Source}} -> {{.Destination}}{{println}}{{end}}'

# Check container user UID/GID
docker exec <container> id

# Check file ownership (numeric)
ls -ln /path/to/dir

# Check SMB mount options
mount | grep cifs

# Regenerate deploy workflows after changing template
cd scripts && go run . gh:secrets:update

# Push updated secrets (includes env file changes)
cd scripts && go run . gh:secrets:push --env=prod

# Deploy without rebuilding image
gh workflow run "Deploy prod" --field skip_build=true
```
