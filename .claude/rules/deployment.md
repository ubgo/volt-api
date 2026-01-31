# Deployment Rules

**Read before:** Deploying to production VPS.

> **Deep docs:** `.claude/docs/backend/github-actions-deployment.md`, `.claude/docs/backend/docker-production.md`

## Quick Reference

### Automatic Deployment (Push to Main)

```bash
git push origin main
# → GitHub Actions builds image → Pushes to GHCR → Deploys to VPS
```

### Manual Deployment Trigger

1. Go to: **Repository → Actions → Build and Deploy**
2. Click **"Run workflow"**
3. Select branch: `main`

## Task Commands

| Task | Purpose |
|------|---------|
| `task docker:prod:build` | Build production Docker image locally |
| `task docker:prod:run` | Run container locally for testing |
| `task docker:prod:logs` | View container logs |
| `task scripts:gh:secrets:update` | Sync .env.prod to GitHub secrets |
| `task scripts:gh:secrets:push` | Push all secrets to GitHub |
| `task gh:secrets:list` | List secrets on GitHub |

## Deployment Flow

```
Push to main
    ↓
GitHub Actions: build-and-push job
    ↓
Docker image → ghcr.io/<org>/<repo>:latest
    ↓
GitHub Actions: deploy job (SSH to VPS)
    ↓
VPS: Pull image, write secrets, start container
    ↓
Run migrations: /app/dbmigrate
```

## Pre-Deployment Checklist

- [ ] `.env.prod` updated with all required vars
- [ ] GitHub secrets synced: `task scripts:gh:secrets:update`
- [ ] Database migrations tested locally
- [ ] No secrets in code/git

## VPS Directory Structure

```
/khanakia/wildai_githubapp/
├── .env.prod           # Environment vars (from GitHub secret)
├── _keys/              # Private keys (read-only mount)
└── logs/               # Application logs
```

## Checking Deployed Version

```bash
# Via API
curl http://your-vps:2315/api/version

# Via SSH
ssh root@vps docker exec wildai-githubapp /app/cli version
```

## View Logs on VPS

```bash
# All logs
docker logs -f wildai-githubapp

# API only
docker exec wildai-githubapp tail -f /app/tmp/logs/api.out.log

# Cron only
docker exec wildai-githubapp tail -f /app/tmp/logs/cron.out.log
```

## Rollback

```bash
# SSH to VPS
ssh root@vps

# Pull specific version (use commit SHA)
docker pull ghcr.io/<org>/<repo>:<commit-sha>

# Restart with old version
docker stop wildai-githubapp && docker rm wildai-githubapp
docker run -d --name wildai-githubapp ... ghcr.io/<org>/<repo>:<commit-sha>
```

## Troubleshooting

| Issue | Check |
|-------|-------|
| Build fails | GitHub Actions logs, Dockerfile syntax |
| Deploy fails | SSH key, VPS_HOST, GHCR_PAT token |
| Container won't start | `docker logs`, missing env vars |
| App errors | `/app/tmp/logs/`, database connectivity |

## Related Docs

- `.claude/docs/backend/github-actions-deployment.md` - Full setup guide
- `.claude/docs/backend/docker-production.md` - Dockerfile details
- `.claude/docs/backend/secrets-cli.md` - Secrets management
