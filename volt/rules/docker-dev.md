# Docker Development Rules

**Read before:** Setting up or managing local Docker environment.

## Quick Start

```bash
# Start development services (PostgreSQL, Redis, NATS)
task docker:up

# Check status
task docker:ps

# View logs
task docker:logs

# Stop all
task docker:down
```

## Development Commands

| Command | Purpose |
|---------|---------|
| `task docker:up` | Start containers (rebuild if needed) |
| `task docker:up:force` | Force rebuild and restart |
| `task docker:down` | Stop and remove containers |
| `task docker:stop` | Stop without removing |
| `task docker:logs` | View all container logs |
| `task docker:ps` | Show container status |
| `task docker:rebuild` | Rebuild without cache |

## Production Testing (Local)

| Command | Purpose |
|---------|---------|
| `task docker:prod:build` | Build production image |
| `task docker:prod:build:nocache` | Build without cache |
| `task docker:prod:run` | Run production container locally |
| `task docker:prod:stop` | Stop production container |
| `task docker:prod:logs` | View production logs |
| `task docker:prod:logs:api` | View API logs only |
| `task docker:prod:logs:cron` | View cron logs only |
| `task docker:prod:shell` | Shell into container |
| `task docker:prod:cli -- <cmd>` | Run CLI command |
| `task docker:prod:status` | Supervisor process status |

## GitHub Actions Testing (Local)

| Command | Purpose |
|---------|---------|
| `task docker:act:build-push` | Test build-and-push job |
| `task docker:act:deploy` | Test deploy job (skips build) |
| `task docker:act:full` | Test full workflow |
| `task docker:act:dry` | Dry run (show what would happen) |
| `task docker:act:list` | List available jobs |

## Development Workflow

```bash
# 1. Start infrastructure
task docker:up

# 2. Run migrations
task dbent:migrate

# 3. Start API (separate terminal)
go run ./saas/cmd/api

# 4. Start cron if needed (separate terminal)
go run ./saas/cmd/cron
```

## Test Production Build Locally

```bash
# Build and run
task docker:prod:build
task docker:prod:run

# Check it's working
curl http://localhost:2315/api/version

# View logs
task docker:prod:logs

# Stop when done
task docker:prod:stop
```

## Troubleshooting

### Containers Won't Start

```bash
# Check what's using ports
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
lsof -i :4222  # NATS

# Force restart
task docker:up:force
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
task docker:ps

# View PostgreSQL logs
docker compose logs postgres

# Connect manually
docker compose exec postgres psql -U postgres
```

### Fresh Start (Remove Volumes)

```bash
cd .docker
docker compose down -v
task docker:up
```

### Port Conflicts

If host services conflict:

```bash
# Stop host PostgreSQL/Redis
brew services stop postgresql
brew services stop redis

# Or change ports in docker-compose.yaml
```

## File Locations

| File | Purpose |
|------|---------|
| `.docker/docker-compose.yaml` | Development services |
| `.docker/Dockerfile.prod` | Production build |
| `.docker/Taskfile.yml` | Docker task commands |
| `.env.prod` | Production environment (for local prod testing) |
| `_keys/` | Private keys (mounted to container) |

## Custom Variables

```bash
# Custom image name
task docker:prod:build IMAGE_NAME=myapp IMAGE_TAG=v1.0.0

# Custom volume paths
task docker:prod:run KEYS_DIR=/path/to/keys LOGS_DIR=/path/to/logs
```
