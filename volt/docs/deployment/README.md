# Deployment Guide (Fly.io)

This project is configured for **One-Click Deployment** on [Fly.io](https://fly.io/). It uses a multi-stage Docker build optimized for production.

## 🚀 Prerequisites

1. **Install Fly CLI:**
   ```bash
   brew install flyctl
   ```
2. **Login:**
   ```bash
   fly auth login
   ```

## 📦 Initial Setup (First Time Only)

Since we already have a `fly.toml`, you just need to initialize the app on Fly's servers.

1. **Launch the App:**
   Run this command in the project root. It will detect `fly.toml`.
   ```bash
   fly launch --copy-config
   ```
   * **Choose an App Name:** (e.g., `my-scraper-app`)
   * **Select Region:** (e.g., Amsterdam `ams`, Frankfurt `fra`)
   * **Database:** Select **Yes** to set up a Postgres database automatically.
   * **Redis:** No (unless you need it).
   * **Deploy now?** No (we need to set secrets first).

2. **Verify Database Config:**
   When you attach a Postgres DB, Fly automatically sets a `DATABASE_URL` secret.
   However, our application might expect specific config keys (like `DB_HOST`, `DB_USER`).

   You have two options:

   * **Option A (Easy):** Update your PKL config (`config/RootConfig.pkl`) to read `DATABASE_URL` directly.
   * **Option B (Manual):** Parse the connection string and set individual secrets:
     ```bash
     fly secrets set DB_HOST=... DB_USER=... DB_PASS=...
     ```

## 🚀 Deploying

Once setup is done, deploying is a single command:

```bash
fly deploy
```

This will:

1. Build the Docker image using `.docker/Dockerfile.prod`.
2. Push the image to Fly's registry.
3. Update the running machines.

## ⚙️ Configuration in Production

We use environment variables to override defaults in production. The `config` logic should be set up to read these.

**Recommended Secrets to Set:**

```bash
fly secrets set \
    APP_ENV=production \
    GRAPHQL_KEY=$(openssl rand -hex 16) \
    GRAPHQL_QUERY_TOKEN=$(openssl rand -hex 16) \
    SENTRY_DSN=https://...
```

## 🔄 CI/CD (GitHub Actions)

To automate deployment on every push to `main`:

1. Get a Fly API Token: `fly tokens create deploy -x 999999h`
2. Add it to GitHub Secrets as `FLY_API_TOKEN`.
3. The workflow in `.github/workflows/deploy.yml` (if created) will handle the rest.

---

# VPS Deployment (Docker + GitHub Actions)

This project also supports deployment to a VPS using Docker containers with GitHub Actions CI/CD.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Container                         │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                    Supervisord                          │ │
│  │  ┌─────────────────┐    ┌─────────────────────────┐    │ │
│  │  │   API Server    │    │    Cron Scheduler       │    │ │
│  │  │   (Gin HTTP)    │    │    (gocron/v2)          │    │ │
│  │  │                 │    │                         │    │ │
│  │  │  - GraphQL      │    │  - processFiles         │    │ │
│  │  │  - REST APIs    │    │  - processLinks         │    │ │
│  │  │  - Health check │    │  - processTexts         │    │ │
│  │  │                 │    │  - processEmbeddings    │    │ │
│  │  │                 │    │  - syncStripe           │    │ │
│  │  │                 │    │  - workspaceLimit       │    │ │
│  │  │                 │    │  - etc.                 │    │ │
│  │  └─────────────────┘    └─────────────────────────┘    │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Deployment Flow

```
Push to main branch
        │
        ▼
┌───────────────────────┐
│  GitHub Actions       │
│  build-and-push job   │
│  - Build Docker image │
│  - Push to GHCR       │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│  GitHub Actions       │
│  deploy job           │
│  - SSH to VPS         │
│  - Pull new image     │
│  - Graceful stop      │◄─── docker stop -t 180
│  - Start new container│
│  - Run migrations     │
└───────────────────────┘
```

## Graceful Shutdown System

The deployment process ensures zero data loss by implementing graceful shutdown at multiple layers.

### The Problem

When deploying a new version:
1. **Cron jobs** might be mid-execution (processing files, syncing Stripe, etc.)
2. **API requests** might be in-flight (GraphQL queries, file uploads, etc.)

Simply killing the container would:
- Corrupt data if a cron job is mid-write
- Return errors to users with pending requests
- Leave external systems (Stripe, etc.) in inconsistent states

### The Solution: Multi-Layer Graceful Shutdown

```
docker stop -t 180  ──────►  SIGTERM sent to container
                                    │
                                    ▼
                            ┌───────────────────┐
                            │   Supervisord     │
                            │   receives SIGTERM│
                            └─────────┬─────────┘
                                      │
                    ┌─────────────────┴─────────────────┐
                    │                                   │
                    ▼                                   ▼
        ┌───────────────────┐               ┌───────────────────┐
        │   API Process     │               │   Cron Process    │
        │                   │               │                   │
        │ 1. Stop accepting │               │ 1. scheduler.     │
        │    new connections│               │    StopJobs()     │
        │                   │               │    (waits for     │
        │ 2. Wait for       │               │    running jobs)  │
        │    in-flight      │               │                   │
        │    requests       │               │ 2. scheduler.     │
        │    (45s timeout)  │               │    Shutdown()     │
        │                   │               │    (2min timeout) │
        │ 3. Close DB, Redis│               │                   │
        │    NATS, Otel     │               │ 3. Close DB, Redis│
        │                   │               │    NATS, Otel     │
        └───────────────────┘               └───────────────────┘
                    │                                   │
                    └─────────────────┬─────────────────┘
                                      │
                                      ▼
                            ┌───────────────────┐
                            │  Clean shutdown   │
                            │  Container exits  │
                            └───────────────────┘
```

### Configuration Files

#### 1. Docker Stop Timeout (`.github/workflows/deploy.yml`)

```yaml
# Sends SIGTERM and waits up to 180 seconds for graceful shutdown
docker stop -t 180 ${CONTAINER}
```

#### 2. Supervisor Configuration (`.docker/supervisord.prod.conf`)

```ini
[program:api]
command=/app/api
stopsignal=TERM          # Send SIGTERM for graceful shutdown
stopwaitsecs=45          # Wait 45s for API to drain connections
stopasgroup=true         # Kill all child processes together
killasgroup=true

[program:cron]
command=/app/cron
stopsignal=TERM          # Send SIGTERM for graceful shutdown
stopwaitsecs=150         # Wait 2.5min for cron jobs to complete
stopasgroup=true         # Kill all child processes together
killasgroup=true
```

#### 3. Cron Shutdown Handler (`webapp/cmd/cron/main.go`)

```go
shutdown.New(
    shutdown.WithTimeout(2*time.Minute), // Allow running jobs to complete
    shutdown.WithShutdownHandler(
        func(ctx context.Context) error {
            // StopJobs() waits for all running jobs to finish
            if err := scheduler.StopJobs(); err != nil {
                return err
            }
            // Shutdown() releases scheduler resources
            return scheduler.Shutdown()
        },
        app.ShutdownHandler, // Close DB, Redis, NATS, Otel
    ),
).Listen()
```

#### 4. API Shutdown Handler (`webapp/cmd/api/main.go`)

```go
shutdown.New(
    shutdown.WithShutdownHandler(
        app.ShutdownHandler,           // Close DB, Redis, NATS, Otel
        func(ctx context.Context) error {
            return server.Shutdown(ctx) // Graceful HTTP shutdown
        },
    ),
).Listen()
```

### Timeout Hierarchy

| Layer | Timeout | Purpose |
|-------|---------|---------|
| Docker stop | 180s | Maximum time before SIGKILL |
| Supervisor (cron) | 150s | Wait for cron jobs |
| Supervisor (api) | 45s | Wait for HTTP drain |
| Cron shutdown handler | 120s | Context timeout for job completion |
| API shutdown handler | 5s | Context timeout for HTTP drain |

### Signal Flow

1. **SIGTERM** → Docker sends to container's PID 1 (supervisord)
2. **Supervisord** → Forwards SIGTERM to managed processes (api, cron)
3. **Go signal handler** → Catches SIGTERM via `signal.Notify()`
4. **Shutdown handlers** → Execute cleanup in order
5. **Process exits** → Returns control to supervisord
6. **Container stops** → Docker removes container

### Cron Job Safety Features

Each cron job uses `gocron.WithSingletonMode(gocron.LimitModeReschedule)`:
- Prevents concurrent execution of the same job
- If a job is running when scheduled again, it reschedules instead of stacking
- During shutdown, `StopJobs()` waits for any running job to complete

### What Happens During Deployment

```
Time    Event
─────   ──────────────────────────────────────────────────────
T+0     Push to main branch triggers GitHub Action
T+2m    Docker image built and pushed to GHCR
T+3m    SSH to VPS, pull new image
T+3m    docker stop -t 180 sent (SIGTERM)
T+3m    Cron: If processEmbeddings is running, it continues...
T+3m    API: Stops accepting new connections
T+3m    API: Waits for in-flight requests to complete
T+3m30s API: Shutdown complete, closes connections
T+4m    Cron: processEmbeddings completes naturally
T+4m    Cron: scheduler.Shutdown() called
T+4m    Cron: DB, Redis, NATS connections closed
T+4m    Container exits cleanly
T+4m    New container starts
T+4m5s  Health check passes
T+4m10s Migrations run
T+4m15s Deployment complete!
```

### Monitoring Shutdown

Check logs during deployment:

```bash
# On VPS
docker logs -f ${CONTAINER_NAME}

# Look for these messages:
# "Shutting down scheduler..."
# "Scheduler jobs stopped gracefully"
# "Scheduler stopped."
# "Shutting down API..."
# "API stopped."
```

### Troubleshooting

#### Cron job killed mid-execution
- **Symptom**: Data corruption, incomplete processing
- **Check**: `stopwaitsecs` in supervisor config (should be > max job duration)
- **Check**: `docker stop -t` timeout (should be > supervisor stopwaitsecs)

#### API returning 502 during deployment
- **Symptom**: Users see errors during deploy
- **Solution**: Implement blue-green deployment or use a load balancer with health checks

#### Container takes too long to stop
- **Symptom**: Deployment is slow
- **Check**: Which cron job is running (`docker exec ${CONTAINER} supervisorctl status`)
- **Consider**: Reducing job batch sizes or adding checkpoints

### Future Improvements: Zero-Downtime Deployment

For true zero-downtime, consider:

1. **Blue-Green Deployment**
   - Run two containers (blue/green)
   - Switch traffic after new container is healthy
   - Gracefully stop old container

2. **Rolling Deployment with Load Balancer**
   - Use nginx/traefik as reverse proxy
   - Health check removes unhealthy containers from rotation
   - New container added to rotation when healthy

3. **Kubernetes**
   - Native rolling update support
   - Pod disruption budgets
   - Readiness/liveness probes
