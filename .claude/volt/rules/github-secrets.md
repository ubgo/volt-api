# GitHub Secrets Rules

**Read before:** Updating GitHub repository secrets from local files.

> **Deep docs:** `.claude/volt/docs/backend/secrets-cli.md`, `.claude/volt/docs/backend/github-actions-deployment.md`

## Quick Reference

### Update .env.prod → GitHub

```bash
# 1. Edit .env.prod with new values
vim .env.prod

# 2. Sync to GitHub secrets
task scripts:gh:secrets:update
```

This:
1. Base64 encodes `.env.prod`
2. Updates `ENV_PROD_B64` in `.github/.secrets`
3. Pushes to GitHub repository secrets

### Push All Secrets

```bash
# Push everything in .github/.secrets to GitHub
task scripts:gh:secrets:push
```

## File Locations

| File | Purpose |
|------|---------|
| `.env.prod` | Production environment variables |
| `.github/.secrets` | All GitHub secrets (local copy) |
| `.github/.secrets.example` | Template for secrets |

## Workflow

```
Edit .env.prod
    ↓
task secrets:update    # Encodes and pushes ENV_PROD_B64
    ↓
GitHub Actions uses ENV_PROD_B64 during deploy
    ↓
VPS: Decodes to .env.prod file
```

## Adding New Secret

1. Add to `.github/.secrets`:
   ```
   NEW_SECRET=value
   ```

2. Push to GitHub:
   ```bash
   task secrets:push
   ```

3. Update workflow if needed (`.github/workflows/deploy.yml`)

## Adding New Key File

1. Add file to `_keys/` directory

2. Base64 encode:
   ```bash
   cat _keys/newfile.pem | base64 | tr -d '\n'
   ```

3. Add to `.github/.secrets`:
   ```
   KEYS_NEWFILE_B64=<base64-content>
   ```

4. Push: `task secrets:push`

5. Update workflow to decode:
   ```yaml
   echo "${KEYS_NEWFILE_B64}" | base64 -d > ${DEPLOY_PATH}/_keys/newfile.pem
   ```

## Required Secrets

| Secret | Description |
|--------|-------------|
| `VPS_HOST` | VPS IP address |
| `VPS_USER` | SSH username (root) |
| `VPS_SSH_KEY` | SSH private key |
| `GHCR_PAT` | GitHub PAT with `read:packages` |
| `ENV_PROD_B64` | Base64-encoded .env.prod |

## Skipped Secrets

These are NOT pushed (only for local `act` testing):
- `GITHUB_TOKEN`
- `GITHUB_ACTOR`

## Common Tasks

```bash
# View current secrets on GitHub
gh secret list

# Set single secret manually
gh secret set SECRET_NAME --body "value"

# View secret (won't show value, just confirms exists)
gh secret list | grep SECRET_NAME
```

## Task Commands

```bash
# Update .env.prod → GitHub (encodes and pushes ENV_PROD_B64)
task scripts:gh:secrets:update

# Push all secrets from .github/.secrets
task scripts:gh:secrets:push

# List secrets on GitHub
task gh:secrets:list

# Set a single secret
task gh:secrets:set KEY=MY_KEY VALUE=myvalue

# Read local secrets file
task scripts:gh:secrets:read

# Clean orphan secrets from GitHub
task scripts:gh:secrets:clean
```

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `gh: command not found` | Install GitHub CLI: `brew install gh` |
| `gh: not authenticated` | Run `gh auth login` |
| Secret not updating | Check `.github/.secrets` format |
| Deployment uses old secret | Re-run deploy workflow |

## Security Notes

- Never commit `.env.prod` or `.github/.secrets` to git
- Use `read:packages` scope only for `GHCR_PAT`
- Rotate secrets periodically
- Use dedicated SSH key for deployments
