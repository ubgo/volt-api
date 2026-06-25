# Secrets Build Package

## Overview

The `secrets_build` package automates management of secrets and key files for multi-environment GitHub Actions deployments. It reads a single `.env.{env}` file per environment, extracts GitHub-specific secrets, base64-encodes key files, and generates both the secrets file and the deploy workflow YAML.

## How It Works

Each environment (prod, staging, dev, etc.) is driven by a single `.env.{env}` file at the project root. This file contains:

1. **App config** — regular env vars consumed by the container at runtime
2. **GitHub secrets** — vars prefixed with `GITHUB_SECRET_` that become GitHub Environment secrets
3. **Key file list** — `GITHUB_SECRET_KEY_FILES` pointing to files in `_keys/` that get base64-encoded

### Flow

```
.env.prod                          ─┐
_keys/vps_deploy.key                │  go run ./scripts gh:secrets:update --env=prod
_keys/nats_prod.creds               │
                                    ▼
                    ┌──────────────────────────────────┐
                    │  secrets_build.go                 │
                    │  1. Parse GITHUB_SECRET_* vars    │
                    │  2. Base64-encode key files       │
                    │  3. Base64-encode filtered .env   │
                    │  4. Write .secrets.{env}          │
                    │  5. Render deploy.{env}.yml       │
                    └──────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────────┐
                    ▼               ▼                   ▼
          .github/.secrets.prod   deploy.prod.yml    (gh:secrets:push)
          (local reference)       (committed)        → GitHub Environment
```

### Env File Convention

In `.env.{env}`, variables are split into two categories:

**App vars** (deployed to the container via `ENV_B64`):
```
PORT=2310
DB_HOST=localhost
NATS_CREDS_FILE=/app/_keys/nats_prod.creds
```

**GitHub-only vars** (stripped before deployment, used only for GitHub Actions):
```
GITHUB_SECRET_VPS_HOST=188.34.136.211
GITHUB_SECRET_VPS_USER=root
GITHUB_SECRET_VPS_PORT=22
GITHUB_SECRET_GHCR_PAT=ghp_xxx
GITHUB_SECRET_CONTAINER_NAME=myapp-api
GITHUB_SECRET_DEPLOY_PATH=/deploy/myapp
GITHUB_SECRET_VPS_DEPLOY_KEY_FILE=vps_deploy.key
GITHUB_SECRET_KEY_FILES=["vps_deploy.key","nats_prod.creds"]
```

The `GITHUB_SECRET_` prefix is stripped to form the GitHub secret name (e.g., `GITHUB_SECRET_VPS_HOST` becomes secret `VPS_HOST`).

`GITHUB_SECRET_VPS_DEPLOY_KEY_FILE` specifies which key file (from `_keys/`) is the SSH deploy key used to connect to the VPS. Each environment can point to a different key file for deploying to different servers. If not set, defaults to `VPS_DEPLOY_KEY_B64`.

### Key File Handling

Files listed in `GITHUB_SECRET_KEY_FILES` are read from `_keys/`, base64-encoded, and stored as secrets with generated names:

| File in `_keys/`             | GitHub Secret Name                                          |
|------------------------------|-------------------------------------------------------------|
| `vps_deploy.key`             | `VPS_DEPLOY_KEY_B64`                                        |
| `nats_prod.creds`            | `NATS_PROD_KEY_B64`                                         |
| `github-app-2025.prod.pem`   | `GITHUB_APP_2025_PROD_KEY_B64`                              |

Naming rule: strip extension, uppercase, replace `-` and `.` with `_`, append `_KEY_B64`.

### Dynamic Workflow Generation

The deploy workflow is generated from `.github/workflows/deploy.yml.gotmpl`. The template uses `[[ ]]` delimiters (to avoid conflicts with GitHub Actions `${{ }}`).

Key files are **dynamically injected** into the generated workflow:

1. **`env:` block** — each key file secret is passed to the SSH action
2. **`envs:` list** — each key file secret name is forwarded to the remote host
3. **Deploy script** — each key file is decoded from base64 and written to `${DEPLOY_PATH}/_keys/` with `chmod 600`

When an environment has no key files, these sections are omitted cleanly.

Template data available:

| Field          | Example              | Description                                    |
|----------------|----------------------|------------------------------------------------|
| `.Name`        | `prod`               | Environment name                               |
| `.NameUpper`   | `PROD`               | Uppercased environment name                    |
| `.Branch`      | `main`               | Git branch (prod→main, else same)              |
| `.KeyFiles`    | `[]KeyFileInfo`      | List of `{SecretName, Filename}`               |
| `.SSHKeySecret`| `VPS_DEPLOY_KEY_B64` | Secret name for the SSH deploy key             |

## Commands

```bash
# Build secrets + generate workflow for one environment
go run ./scripts gh:secrets:update --env=prod

# Build secrets + generate workflows for all environments
go run ./scripts gh:secrets:update

# Push secrets to GitHub Environment
go run ./scripts gh:secrets:push --env=prod

# Full setup: create GitHub Environment + build + push + generate workflow
go run ./scripts gh:env:setup --env=prod

# Clean orphan secrets from GitHub Environment
go run ./scripts gh:secrets:clean --env=prod

# List available environments (detected from .env.{env} files)
go run ./scripts gh:secrets:list

# Read generated secret keys (values hidden)
go run ./scripts gh:secrets:read --env=prod

# Preview normalized environment name for a branch
go run ./scripts gh:env:normalize feature/my-branch
```

## File Structure

```
_keys/
  ├── vps_deploy.key                    # SSH key for deployment
  ├── nats_prod.creds                   # NATS credentials
  └── ...                               # Any key files referenced in .env.{env}

.env.prod                               # App config + GITHUB_SECRET_* vars
.env.staging                            # Same structure, different values

.github/
  ├── .secrets.prod                     # Auto-generated (base64-encoded secrets)
  ├── .secrets.staging                  # Per-environment
  └── workflows/
      ├── deploy.yml.gotmpl            # Template (committed, hand-edited)
      ├── deploy.prod.yml              # Auto-generated (committed)
      └── deploy.staging.yml           # Auto-generated (committed)
```

## Workflow

### Adding a New Key

1. Place the key file in `_keys/` (any name):
   ```bash
   cp ~/new_service.pem _keys/new_service.pem
   ```

2. Add it to `GITHUB_SECRET_KEY_FILES` in `.env.{env}`:
   ```
   GITHUB_SECRET_KEY_FILES=["vps_deploy.key","nats_prod.creds","new_service.pem"]
   ```

3. Reference the path in your app config (same `.env.{env}` file):
   ```
   NEW_SERVICE_KEY_FILE=/app/_keys/new_service.pem
   ```

4. Regenerate and push:
   ```bash
   go run ./scripts gh:secrets:update --env=prod
   go run ./scripts gh:secrets:push --env=prod
   ```

The generated `deploy.prod.yml` will automatically include the new key file in `env:`, `envs:`, and the decode script.

### Rotating a Key

1. Replace the file in `_keys/` (keep the same filename)
2. Run `go run ./scripts gh:secrets:update --env=prod`
3. Run `go run ./scripts gh:secrets:push --env=prod`

### Removing a Key

1. Remove the filename from `GITHUB_SECRET_KEY_FILES` in `.env.{env}`
2. Run `go run ./scripts gh:secrets:update --env=prod`
3. Run `go run ./scripts gh:secrets:clean --env=prod` to remove the orphan secret from GitHub

### Adding a New Environment

```bash
# 1. Create the env file (copy from existing and modify)
cp .env.prod .env.staging

# 2. Full setup: creates GitHub Environment, builds secrets, pushes, generates workflow
go run ./scripts gh:env:setup --env=staging

# 3. Commit the generated deploy.staging.yml
```

Branch mapping: `prod` → `main`, everything else uses the environment name as the branch name.

## Technical Details

### Template Delimiters

The `.gotmpl` uses `[[` and `]]` as delimiters to coexist with GitHub Actions' `${{ }}` syntax. Whitespace trimming (`[[-` and `-]]`) is used to keep the generated YAML clean.

### ENV_B64

The entire `.env.{env}` file (excluding `GITHUB_SECRET_*` lines) is base64-encoded as the `ENV_B64` secret. During deployment, it's decoded and written as the container's env file.

### Secret Name Restrictions

GitHub does not allow secrets with the `GITHUB_` prefix. Any `GITHUB_SECRET_*` vars that would produce a `GITHUB_*` secret name are automatically skipped during push.
