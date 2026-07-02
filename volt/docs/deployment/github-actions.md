# GitHub Actions Deployment Guide

## Overview

Automated CI/CD pipeline that builds and deploys the Docker container to a Hetzner VPS when commits are pushed to the `main` branch.

## Architecture

```
Push to main → GitHub Actions → Build Image → Push to GHCR → SSH to VPS → Pull & Run
```

## File Locations

| File | Purpose |
|------|---------|
| `.github/workflows/deploy.yml` | Main CI/CD workflow |
| `.docker/Dockerfile.prod` | Production Docker build |
| `.env.prod` | Production environment variables (not in git) |
| `_keys/` | Private keys directory (not in git) |

***

## Step-by-Step Setup Tutorial

### Step 1: Create GitHub Personal Access Token (PAT)

The VPS needs a token to pull images from GitHub Container Registry (GHCR).

1. Go to: **GitHub.com → Settings → Developer settings → Personal access tokens → Tokens (classic)**
2. Click **"Generate new token (classic)"**
3. Set:
   * **Note:** `GHCR Pull Access for VPS`
   * **Expiration:** No expiration (or set reminder to rotate)
   * **Scopes:** Check only `read:packages`
4. Click **"Generate token"**
5. **Copy the token immediately** (you won't see it again)

### Step 2: Generate SSH Key for VPS Access

If you don't have a dedicated deploy key:

```bash
# Generate new SSH key pair
ssh-keygen -t ed25519 -C "github-actions-deploy-wildai" -f ~/.ssh/github_actions_deploy_wildai

# View the private key (for GitHub secret)
cat ~/.ssh/github_actions_deploy_wildai

# View the public key (for VPS authorized_keys)
cat ~/.ssh/github_actions_deploy_wildai.pub
```

Add the **public key** to your VPS:

```bash
# On VPS
echo "ssh-ed25519 AAAA... github-actions-deploy-wildai" >> ~/.ssh/authorized_keys
```

### Step 3: Prepare Base64-Encoded Secrets

All sensitive files are stored as base64-encoded GitHub secrets.

```bash
# Encode .env.prod file
cat .env.prod | base64 | tr -d '\n' | pbcopy
# Paste into GitHub secret: ENV_PROD_B64

# Encode GitHub App private key
cat _keys/githubapp_thewildai_2025_12_28_private_key.pem | base64 | tr -d '\n' | pbcopy
# Paste into GitHub secret: GITHUBAPP_THEWILDAI_2025_12_28_PRIVATE_KEY

# If you have additional key files (e.g., NATS credentials)
cat _keys/nats.creds | base64 | tr -d '\n' | pbcopy
# Paste into GitHub secret: KEYS_NATS_CREDS_B64
```

### Step 4: Configure GitHub Secrets

Go to: **Repository → Settings → Secrets and variables → Actions → New repository secret**

Add these secrets:

| Secret Name | Value | Description |
|-------------|-------|-------------|
| `VPS_HOST` | `65.108.xxx.xxx` | Hetzner VPS IP address |
| `VPS_USER` | `root` | SSH username |
| `VPS_SSH_KEY` | `-----BEGIN OPENSSH...` | Full private SSH key content |
| `GHCR_PAT` | `ghp_xxxx...` | GitHub PAT from Step 1 |
| `ENV_PROD_B64` | `Q09ORklHX0...` | Base64-encoded .env.prod |
| `GITHUBAPP_THEWILDAI_2025_12_28_PRIVATE_KEY` | `LS0tLS1CRU...` | Base64-encoded GitHub App key |

### Step 5: VPS Prerequisites

The workflow automatically creates all directories and files. You only need:

1. **Docker installed** on the VPS
2. **SSH access** with the key you configured in Step 2

The workflow will:

* Create `/khanakia/wildai_githubapp/` directory
* Create `_keys/` and `logs/` subdirectories
* Write `.env.prod` and key files
* Login to GHCR, pull image, and start container

### Step 6: Test the Deployment

**Option A: Push to main**

```bash
git add .
git commit -m "chore: trigger deployment"
git push origin main
```

**Option B: Manual trigger**

1. Go to: **Repository → Actions → Build and Deploy**
2. Click **"Run workflow"**
3. Select branch: `main`
4. Click **"Run workflow"**

### Step 7: Monitor Deployment

1. Go to: **Repository → Actions**
2. Click on the running workflow
3. Watch both jobs: `build-and-push` and `deploy`

***

## VPS Directory Structure

After deployment, the VPS will have:

```
/khanakia/wildai_githubapp/
├── .env.prod                        # Environment variables
├── _keys/                           # Private keys (read-only mount)
│   └── githubapp_thewildai_2025_12_28_private_key.pem
└── logs/                            # Application logs
    ├── api.out.log
    ├── cron.out.log
    └── supervisord.log
```

***

## GitHub Secrets Reference

### Required Secrets

| Secret | Example | Notes |
|--------|---------|-------|
| `VPS_HOST` | `65.108.123.45` | IPv4 address |
| `VPS_USER` | `root` or `deploy` | User with Docker permissions |
| `VPS_SSH_KEY` | PEM content | Include `-----BEGIN` and `-----END` lines |
| `GHCR_PAT` | `ghp_xxx...` | Needs `read:packages` scope |
| `ENV_PROD_B64` | Base64 string | Full .env.prod content |
| `GITHUBAPP_THEWILDAI_2025_12_28_PRIVATE_KEY` | Base64 string | GitHub App private key |

SSH port is hardcoded to 22 in the workflow. Edit `.github/workflows/deploy.yml` if different.

### GITHUB\_TOKEN vs GHCR\_PAT - Important Difference

The workflow uses **two different tokens** for different purposes:

| Token | Used In | Purpose | Where It Runs | Scope Needed |
|-------|---------|---------|---------------|--------------|
| `GITHUB_TOKEN` | build-and-push job | Push images to GHCR | GitHub Runner | `write:packages` (auto) |
| `GHCR_PAT` | deploy job | Pull images from GHCR | VPS (external) | `read:packages` |

**Why two tokens?**

1. **`GITHUB_TOKEN`** is a special token auto-generated by GitHub Actions with proper permissions based on the workflow's `permissions:` block. It **only works inside GitHub's runners** and cannot be used externally.

2. **`GHCR_PAT`** (Personal Access Token) is needed because the VPS is external to GitHub and cannot use `GITHUB_TOKEN`. The VPS needs a real PAT to authenticate with GHCR to pull images.

**For local testing with `act`:**

Since `act` doesn't auto-generate `GITHUB_TOKEN`, you must add it to `.github/.secrets`:

```bash
# .github/.secrets needs both:
GITHUB_TOKEN=ghp_your_token_with_write_packages
GHCR_PAT=ghp_your_token_with_read_packages

# Can be the same token if it has both read and write scopes
```

**In production (GitHub Actions):**

* `GITHUB_TOKEN` is auto-generated - no secret needed
* `GHCR_PAT` must be created manually and added to repository secrets

### Optional Secrets (for additional keys)

| Secret | Purpose |
|--------|---------|
| `KEYS_NATS_CREDS_B64` | NATS credentials file |
| `KEYS_OTHER_B64` | Any other private key/config |

See "Adding New Key Files" section below for how to add more keys to the workflow.

***

## Workflow Details

### Triggers

```yaml
on:
  push:
    branches: [main]      # Auto-deploy on push to main
  workflow_dispatch:       # Manual trigger button
```

### Jobs

**1. build-and-push**

* Checkout code
* Login to GHCR
* Build Docker image with version info
* Push to `ghcr.io/<org>/wildai_githubapp:latest`

**2. deploy**

* SSH to VPS
* Write `.env.prod` and `_keys/` from secrets
* Pull latest image
* Stop old container, start new one
* Run database migrations (`/app/dbmigrate`)

### Build Arguments

The workflow passes version info to the Docker build:

| Arg | Source | Example |
|-----|--------|---------|
| `VERSION` | Branch name | `main` |
| `COMMIT` | Short SHA | `481a523` |
| `BRANCH` | Current branch | `main` |
| `BUILD_TIME` | UTC timestamp | `2026-01-09T10:30:00Z` |

Check deployed version:

```bash
curl http://your-vps:2315/api/version
```

***

## Container Management on VPS

### View Logs

```bash
# All container output (supervisor)
docker logs -f wildai-githubapp

# API logs
docker exec wildai-githubapp tail -f /app/tmp/logs/api.out.log

# Cron logs
docker exec wildai-githubapp tail -f /app/tmp/logs/cron.out.log
```

### Supervisor Status

```bash
docker exec wildai-githubapp supervisorctl status
```

### Run CLI Commands

```bash
docker exec wildai-githubapp /app/cli ping
docker exec wildai-githubapp /app/cli version
```

### Shell Access

```bash
docker exec -it wildai-githubapp /bin/sh
```

### Manual Restart

```bash
docker restart wildai-githubapp
```

### Full Redeploy (manual)

```bash
cd /khanakia/wildai_githubapp

# Pull latest
docker pull ghcr.io/khanakia/wildai_githubapp:latest

# Restart
docker stop wildai-githubapp
docker rm wildai-githubapp

docker run -d \
  --name wildai-githubapp \
  --restart unless-stopped \
  --network host \
  --env-file .env.prod \
  -v /khanakia/wildai_githubapp/_keys:/app/_keys:ro \
  -v /khanakia/wildai_githubapp/logs:/app/tmp/logs \
  ghcr.io/khanakia/wildai_githubapp:latest
```

> **Why `--network host`?**
>
> PKL (Apple's configuration language) needs network access to fetch package metadata from `pkg.pkl-lang.org` at startup. Docker's default bridge network mode can cause DNS resolution failures on some VPS configurations. With `--network host`, the container shares the host's network stack directly, bypassing these issues.
>
> This also means **no port mapping (`-p`) is needed** - the app binds directly to the host's port.

> **Why `chown 1000:1000` on logs directory?**
>
> The container runs as non-root user `appuser` (UID 1000) for security. When the workflow creates directories via SSH (as root), they're owned by root. The container's `appuser` can't write logs to a root-owned directory. The `chown` ensures the logs directory is writable by the container.

***

## Troubleshooting

### Build Fails

**"Permission denied" on GHCR push:**

* Check repository Settings → Actions → General → Workflow permissions
* Should be "Read and write permissions"

**"Go version mismatch":**

* Dockerfile uses `GOTOOLCHAIN=auto` which auto-downloads required Go version
* Check `go.work` for required version

### Deploy Fails

**"SSH connection refused":**

* Verify `VPS_HOST`, `VPS_USER`, `VPS_PORT` secrets
* Check VPS firewall allows SSH from GitHub IPs
* Test SSH manually: `ssh -i key user@host`

**"Docker login failed":**

* Verify `GHCR_PAT` has `read:packages` scope
* Check PAT hasn't expired
* Test manually on VPS: `echo "PAT" | docker login ghcr.io -u username --password-stdin`

**"Container exits immediately":**

* Check logs: `docker logs wildai-githubapp`
* Verify `.env.prod` has all required variables
* Check database connectivity from VPS

**"PKL network/DNS error" or "Error connecting to host 'pkg.pkl-lang.org'":**

* PKL needs network access to fetch package metadata at startup
* Ensure container runs with `--network host` (not bridge mode)
* Bridge network mode can cause DNS resolution failures on some VPS setups

### Common Fixes

```bash
# On VPS - check if container is running
docker ps -a | grep wildai

# View recent logs
docker logs --tail 100 wildai-githubapp

# Check env file was written correctly
cat /khanakia/wildai_githubapp/.env.prod | head -5

# Check key file exists
ls -la /khanakia/wildai_githubapp/_keys/

# Test database connectivity
docker exec wildai-githubapp /app/cli ping
```

***

## Security Notes

1. **SSH Key:** Use a dedicated key for GitHub Actions, not your personal key
2. **PAT Scope:** Only grant `read:packages`, nothing more
3. **File Permissions:** Keys and env files are `chmod 600`
4. **Volume Mounts:** `_keys` is mounted read-only (`:ro`)
5. **Non-root:** Container runs as `appuser` (UID 1000), not root

***

## Adding New Key Files

1. Add file to `_keys/` locally
2. Generate base64: `cat _keys/newfile | base64 | tr -d '\n'`
3. Add GitHub secret: `KEYS_NEWFILE_B64`
4. Edit workflow - add to `env:` section under deploy job:
   ```yaml
   env:
     # ... existing vars ...
     NEW_KEY_B64: ${{ secrets.KEYS_NEWFILE_B64 }}
   ```
5. Add to `envs:` list:
   ```yaml
   envs: DEPLOY_PATH,IMAGE,CONTAINER,...,NEW_KEY_B64
   ```
6. Add decode command in script:
   ```bash
   echo "${NEW_KEY_B64}" | base64 -d > ${DEPLOY_PATH}/_keys/newfile
   chmod 600 ${DEPLOY_PATH}/_keys/newfile
   ```
7. Update `.env.prod` if needed to reference the new file path

***

## Rollback

To rollback to a previous version:

```bash
# On VPS
docker pull ghcr.io/khanakia/wildai_githubapp:<commit-sha>
docker stop wildai-githubapp
docker rm wildai-githubapp

# Run with specific tag
docker run -d \
  --name wildai-githubapp \
  ... \
  ghcr.io/khanakia/wildai_githubapp:<commit-sha>
```

Each push creates a tagged image with the commit SHA (e.g., `ghcr.io/.../wildai_githubapp:481a523`).

***

## Local Testing with Act

[Act](https://github.com/nektos/act) is a tool that runs GitHub Actions locally using Docker. This allows you to test workflows before pushing to GitHub.

### Install Act

```bash
# macOS
brew install act

# Linux (with Homebrew)
brew install act

# Linux (direct download)
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Windows (with Chocolatey)
choco install act-cli

# Windows (with Scoop)
scoop install act

# Windows (with winget)
winget install nektos.act
```

### Prerequisites

* **Docker** must be installed and running
* First run downloads `catthehacker/ubuntu:act-latest` image (~1GB)
* On Apple Silicon (M1/M2/M3), you may need `--container-architecture linux/amd64` flag

### Setup Secrets File

```bash
# Copy example and fill in real values
cp .github/.secrets.example .github/.secrets

# Edit with your real secrets
vim .github/.secrets
```

### Test Commands (using Taskfile)

```bash
cd .docker

# List all available tasks
task --list

# Test build only (no push) - safest option
task act:build

# Test build + push to GHCR
task act:build-push

# Test deploy job only (skip build)
task act:deploy

# Test full workflow (build + deploy)
task act:full

# Dry run - show what would happen
task act:dry

# List all jobs in workflows
task act:list

# Custom actor (if your username differs)
task act:build-push ACTOR=yourusername
```

### Test Commands (using act directly)

```bash
# List available jobs
act -l

# Dry-run (show what would happen)
act -n --secret-file .github/.secrets

# Test build job only (no push)
act -j build -W .github/workflows/test-build.yml --bind

# Test build + push
act -j build-and-push --secret-file .github/.secrets --bind -a thewildai

# Test full workflow
act push --secret-file .github/.secrets --bind -a thewildai
```

### Test Individual Steps

If you want to test without act:

```bash
# 1. Test Docker build locally
cd .docker
task prod:build

# 2. Test SSH connection
ssh -i ~/.ssh/github_actions_deploy_wildai root@YOUR_VPS_IP "echo connected"

# 3. Test GHCR login (on VPS)
ssh root@YOUR_VPS_IP
echo "YOUR_GHCR_PAT" | docker login ghcr.io -u khanakia --password-stdin

# 4. Test base64 encoding/decoding
echo "test content" | base64 | base64 -d
```

### Common Act Issues

**"Cannot connect to Docker daemon":**

* Ensure Docker Desktop is running

**"Image not found":**

* Act uses Docker images to simulate GitHub runners
* First run downloads `catthehacker/ubuntu:act-latest` (~1GB)

**"Secret not found":**

* Check `.github/.secrets` file exists and has correct format
* Use `--secret-file` flag

***

## Related Files

* [Docker Production Setup](./docker.md) - Dockerfile.prod details
* [env.prod.sample](../../../env.prod.sample) - Environment variable template
* [.docker/Taskfile.yml](../../../.docker/Taskfile.yml) - Local Docker commands
* [.github/.secrets.example](../../../.github/secrets.example) - Secrets template for local testing
