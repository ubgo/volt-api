# Incident: Cron Graceful Shutdown Not Working with `go run`

**Date**: 2025-01-18
**Status**: Resolved
**Severity**: Medium

## Summary

When using `supervisorctl stop cron` or `docker stop`, the cron service terminates immediately instead of waiting for running jobs to complete. The graceful shutdown handler in the Go code never receives the SIGTERM signal.

## Root Cause

Two issues were identified:

### Issue 1: `go run` doesn't forward signals

The supervisor config uses `go run ./cmd/cron/main.go` which spawns the actual Go binary as a **child process**. When supervisor sends SIGTERM:

1. SIGTERM is sent to the `go run` process (the parent)
2. `go run` exits immediately
3. The child process (actual cron binary) gets orphaned and killed
4. The signal handler in our code **never receives SIGTERM**

```
supervisor --SIGTERM--> go run (PID 100) --dies--> child process (PID 101) gets killed
                                                    ^ our signal handler is here, never triggered
```

### Issue 2: gocron default stop timeout is 10 seconds

Even with a pre-built binary, gocron's `StopJobs()` has a **default 10-second timeout**. Jobs running longer than 10 seconds will be killed with error: `gocron: timed out waiting for jobs to finish`

## Solution

### Fix 1: Use pre-built binary

Build the binary first, then run the compiled binary directly instead of using `go run`.

### Fix 2: Increase gocron stop timeout

In `saas/cmd/cron/main.go`, add `WithStopTimeout` when creating the scheduler:

```go
scheduler, err := gocron.NewScheduler(
    gocron.WithLogger(gocronlogger.New(applog.GetLogger())),
    gocron.WithMonitor(gocronmonitor.New(applog.GetLogger())),
    gocron.WithMonitorStatus(gocronmonitor.New(applog.GetLogger())),
    gocron.WithStopTimeout(24*time.Hour), // Wait up to 24h for jobs to complete on shutdown
)
```

### Fix 3: Docker compose stop_grace_period

In `.docker/docker-compose.yaml`, add `stop_grace_period` so `docker stop` waits:

```yaml
services:
  chatapi-dev:
    # ... other config
    stop_grace_period: 24h  # Wait for cron jobs to complete on docker stop
```

Without this, Docker's default is 10 seconds, then SIGKILL.

### Fix 4: Remove extra shutdown delay

In `saas/cmd/cron/main.go`, add `WithGracefulShutdownTimeout(0)` to skip the 30-second post-handler sleep:

```go
shutdown.New(
    shutdown.WithMsgShuttingDown("Shutting down scheduler..."),
    shutdown.WithMsgShutdownCompleted("Scheduler stopped."),
    shutdown.WithMsgShutdownFailed("Error shutting down scheduler"),
    shutdown.WithLogger(applog.GetLogger()),
    shutdown.WithGracefulShutdownTimeout(0), // No extra sleep needed after jobs complete
    shutdown.WithShutdownHandler(...),
).Listen()
```

### Supervisor Config Change

Change from:
```ini
[program:cron]
command=go run ./cmd/cron/main.go
```

To:
```ini
[program:cron]
command=/app/saas/bin/cron
```

### Required supervisor settings

Ensure these settings are in the `[program:cron]` section:

```ini
stopsignal=SIGTERM
stopwaitsecs=86400
```

- `stopsignal=SIGTERM` - Send SIGTERM to allow graceful shutdown
- `stopwaitsecs=86400` - Wait up to 24 hours for process to exit (use large value, NOT 0)

**WARNING**: `stopwaitsecs=0` means "kill immediately" - the opposite of waiting!

### Required supervisorctl settings

Ensure these sections exist for `supervisorctl` to work:

```ini
[unix_http_server]
file=/run/supervisord.sock

[supervisorctl]
serverurl=unix:///run/supervisord.sock

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface
```

## Testing Procedure

### Step 1: Exec into the container

```bash
docker exec -it <container_id> /bin/sh
```

### Step 2: Build the cron binary

```bash
cd /app/saas
go build -o ./bin/cron ./cmd/cron/main.go
```

### Step 3: Temporarily modify supervisor config

Edit `/etc/supervisor/conf.d/supervisord.conf` (or wherever your config is mounted):

```bash
vi /etc/supervisor/conf.d/supervisord.conf
```

Change:
```ini
command=go run ./cmd/cron/main.go
```
To:
```ini
command=/app/saas/bin/cron
```

### Step 4: Reload supervisor and restart cron

```bash
supervisorctl reread
supervisorctl update
supervisorctl restart cron
```

### Step 5: Verify cron is running

```bash
supervisorctl status
```

Expected output:
```
cron                             RUNNING   pid 1234, uptime 0:00:05
```

### Step 6: Wait for a job to start

Watch the logs until you see a job running (e.g., "Test shutdown job started 5"):

```bash
# In the container, watch stdout or check logs
```

### Step 7: Test graceful shutdown with supervisorctl

```bash
supervisorctl stop cron
```

**Expected behavior**:
- Supervisor logs: `INFO waiting for cron to stop`
- Cron continues printing job progress (e.g., "Test shutdown job started 6", "7", "8"...)
- After job completes (20 iterations), cron exits
- Supervisor logs: `INFO stopped: cron (exit status 0)`

**Bad behavior (without fix)**:
- Cron terminates immediately mid-job
- Supervisor logs: `WARN stopped: cron (terminated by SIGTERM)`

### Step 8: Test graceful shutdown with docker stop

From outside the container:

```bash
# Give plenty of time for graceful shutdown
docker stop -t 300 <container_id>

# Or with container name
docker stop -t 300 wildai-chatapi-dev
```

The `-t 300` flag tells Docker to wait up to 300 seconds before force-killing.

**Expected behavior**: Container waits for cron job to complete before stopping.

## Verification Logs

### Successful graceful shutdown

```
looping for 20 seconds 18
looping for 20 seconds 19
looping for 20 seconds 20
{"level":"info","msg":"job completed","job_name":"testShutdown","status":"success"}
INFO waiting for cron to stop
{"level":"info","msg":"Shutting down scheduler..."}
{"level":"info","msg":"Scheduler jobs stopped gracefully"}
{"level":"info","msg":"Scheduler stopped."}
INFO stopped: cron (exit status 0)
```

### Failed: with `go run` (signal not forwarded)

```
looping for 20 seconds 6
INFO waiting for cron to stop
WARN stopped: cron (terminated by SIGTERM)
INFO reaped unknown pid 6314 (terminated by SIGPIPE)
```

### Failed: gocron timeout (default 10s)

```
looping for 20 seconds 11
INFO waiting for cron to stop
{"level":"error","msg":"Error stopping scheduler jobs","error":"gocron: timed out waiting for jobs to finish"}
INFO stopped: cron (exit status 0)
```

## Trade-offs

| Approach | Hot Reload | Graceful Shutdown |
|----------|------------|-------------------|
| `go run ./cmd/cron/main.go` | Yes | No |
| Pre-built binary | No | Yes |
| `air` with signal forwarding | Yes | Yes (needs config) |

## Recommendations

1. **Development**: Keep `go run` for hot reload, accept that graceful shutdown won't work
2. **Production**: Always use pre-built binary (`supervisord.prod.conf`)
3. **If graceful shutdown needed in dev**: Use `air` or similar tools that properly forward signals

## Related Files

- `.docker/docker-compose.yaml` - Dev docker compose (add `stop_grace_period: 24h`)
- `.docker/supervisord.conf` - Dev supervisor config (use pre-built binary, set `stopwaitsecs=86400`)
- `.docker/supervisord.prod.conf` - Prod supervisor config (set `stopsignal=SIGTERM`, `stopwaitsecs=86400`)
- `.github/workflows/deploy.yml.gotmpl` - Deploy workflow (use `docker run --stop-timeout 86400`)
- `saas/cmd/cron/main.go` - Cron service entry point (add `WithStopTimeout`, `WithGracefulShutdownTimeout`)
- `lace/shutdown/shutdown.go` - Graceful shutdown handler

## How It Works Under The Hood

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GRACEFUL SHUTDOWN FLOW                               │
└─────────────────────────────────────────────────────────────────────────────┘

1. User runs: supervisorctl stop cron (or docker stop)
   - docker stop: stop_grace_period=24h tells Docker to wait before SIGKILL
                    │
                    ▼
2. Supervisor sends SIGTERM to cron binary (PID)
   - stopwaitsecs=86400 tells supervisor to wait up to 24h before SIGKILL
                    │
                    ▼
3. shutdown.Listen() receives signal via signal.Notify()
        ┌───────────────────────────────────────┐
        │  quit := make(chan os.Signal, 1)      │
        │  signal.Notify(quit, SIGTERM, SIGINT) │
        │  <-quit  ◄── blocks until signal      │
        └───────────────────────────────────────┘
                    │
                    ▼
4. Shutdown handler calls scheduler.StopJobs()
        ┌───────────────────────────────────────┐
        │  - Stops accepting NEW jobs           │
        │  - Waits for RUNNING jobs to complete │
        │  - Respects WithStopTimeout (24h)     │
        └───────────────────────────────────────┘
                    │
                    ▼
5. Running job continues until completion
        ┌───────────────────────────────────────┐
        │  looping for 20 seconds 18            │
        │  looping for 20 seconds 19            │
        │  looping for 20 seconds 20  ◄── done! │
        └───────────────────────────────────────┘
                    │
                    ▼
6. StopJobs() returns, scheduler.Shutdown() called
                    │
                    ▼
7. Process exits with code 0 (GracefulShutdownTimeout=0, no extra wait)
                    │
                    ▼
8. Supervisor sees clean exit, logs: "stopped: cron (exit status 0)"
```

### Why Each Fix Was Needed

| Fix | Problem | Solution |
|-----|---------|----------|
| Pre-built binary | `go run` spawns child process, SIGTERM goes to parent only | Binary receives SIGTERM directly |
| `stopwaitsecs=86400` | Supervisor default 10s, then SIGKILL | Wait 24h for clean exit |
| `stop_grace_period: 24h` | Docker default 10s, then SIGKILL | Wait 24h for container to stop |
| `WithStopTimeout(24h)` | gocron default 10s timeout on StopJobs() | Wait 24h for jobs to complete |
| `WithGracefulShutdownTimeout(0)` | shutdown package sleeps 30s after handlers | No extra wait needed for cron |

## References

- [go run does not forward signals](https://github.com/golang/go/issues/40467)
- [Supervisor stopwaitsecs documentation](http://supervisord.org/configuration.html#program-x-section-settings)
