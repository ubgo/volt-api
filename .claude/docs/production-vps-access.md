# Production VPS Access

## Connection Details

| Field | Value |
|-------|-------|
| **Host** | `188.34.136.211` |
| **User** | `root` |
| **SSH Key Source** | `.github/.secrets` → `VPS_SSH_KEY` |

## Quick Access

```bash
# 1. Extract SSH key from .github/.secrets (VPS_SSH_KEY field)
# 2. Write to temp file:
cat > /tmp/vps_ssh_key << 'EOF'
<paste VPS_SSH_KEY value here>
EOF
chmod 600 /tmp/vps_ssh_key

# 3. SSH to VPS
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /tmp/vps_ssh_key root@188.34.136.211
```

## Common Commands

```bash
# Docker logs (supervisor)
ssh -i /tmp/vps_ssh_key root@188.34.136.211 "docker logs --tail 100 wildai-githubapp"

# API application logs
ssh -i /tmp/vps_ssh_key root@188.34.136.211 "docker exec wildai-githubapp tail -100 /app/tmp/logs/api.out.log"

# Cron logs
ssh -i /tmp/vps_ssh_key root@188.34.136.211 "docker exec wildai-githubapp tail -100 /app/tmp/logs/cron.out.log"

# Supervisor status
ssh -i /tmp/vps_ssh_key root@188.34.136.211 "docker exec wildai-githubapp supervisorctl status"

# Shell into container
ssh -i /tmp/vps_ssh_key root@188.34.136.211 "docker exec -it wildai-githubapp /bin/sh"
```

## Container Info

- **Container name**: `wildai-githubapp`
- **Processes**: `api` + `cron` managed by supervisord
- **Log paths inside container**:
  - `/app/tmp/logs/api.out.log`
  - `/app/tmp/logs/cron.out.log`
