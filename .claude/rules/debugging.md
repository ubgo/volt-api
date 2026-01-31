# Debugging Rules

**Read before:** Investigating issues in the application.

## Quick Diagnosis

| Symptom | Check First |
|---------|-------------|
| App won't start | `task docker:up` (Docker services), config errors |
| DB connection fails | PostgreSQL running, connection string |
| API returns 500 | Logs, trace in observability tool |
| NATS not working | NATS container, `/api/i/bootinfo` |
| Config not loading | `CONFIG_ENV`, PKL syntax |

## Boot Info Endpoint

Check service status after startup:

```bash
curl -H "X-Internal-Key: <your-internal-key>" http://localhost:8080/api/i/bootinfo
```

Returns status of: database, redis, nats, otel

## Log Locations

| Environment | Location |
|-------------|----------|
| Development | stdout (terminal) |
| Production | Configured OTEL endpoint, log files |
| Docker | `docker compose logs -f` |

## Common Issues & Fixes

### Database Connection

```bash
# Check PostgreSQL is running
docker compose ps

# Check connection manually
docker compose exec postgres psql -U postgres -d mydb

# Verify config
CONFIG_ENV=default go run ./saas/cmd/cli ping
```

### NATS Connection

```bash
# Check NATS is running
docker compose ps

# NATS starts async - check bootinfo
curl http://localhost:8080/api/i/bootinfo

# Check NATS logs
docker compose logs nats
```

### Redis Connection

```bash
# Check Redis
docker compose exec redis redis-cli ping

# Check via bootinfo
curl http://localhost:8080/api/i/bootinfo
```

### GraphQL Errors

```bash
# Enable GraphQL playground
# Set graphql.playgroundKey in config, then:
http://localhost:8080/api/sa?key=<playground-key>

# Check resolver panic in logs
docker compose logs api
```

## Adding Debug Logging

```go
// Temporary debug logging
gozap.GetLogger().Debug("debug point",
    zap.Any("value", someValue),
    zap.String("state", "before-operation"),
)

// With context for trace correlation
gozap.FromCtx(ctx).Debug("debug point",
    zap.Any("request", input),
)
```

## Tracing Investigation

1. Find trace_id in logs
2. Search in observability tool (HyperDX, Jaeger, etc.)
3. Follow span hierarchy to find failure point

```go
// Add custom span for debugging
ctx, span := otel.Tracer("debug").Start(ctx, "investigate.issue")
span.SetAttributes(attribute.String("debug.info", "value"))
defer span.End()
```

## Database Query Debugging

The custom Ent driver logs query origins (file:line):

```go
// Enable in development - already configured
// Check logs for query source location
```

Manual query inspection:

```go
// Print query without executing
query := client.User.Query().Where(user.EmailEQ("test@example.com"))
fmt.Println(query.QueryContext(ctx))
```

## Panic Recovery

Check for panics in:
- Resolver code (gqlgen recovers but logs)
- Gin handlers (middleware recovers)
- Background jobs (may crash scheduler)

```bash
# Look for panic in logs
grep -i "panic" logs.txt
```

## Environment Debugging

```bash
# Check which config is loaded
echo $CONFIG_ENV

# Validate PKL config
pkl eval config/pkl/env/default/app.pkl

# Check Go workspace
go work sync
cat go.work
```

## Docker Debugging

```bash
# Restart all services
task docker:up

# View logs
docker compose logs -f

# Shell into container
docker compose exec api sh

# Check container status
docker compose ps

# Rebuild after code changes
docker compose up --build
```

## Profiling (if needed)

```go
import _ "net/http/pprof"

// Add to main.go for profiling endpoints
// Access at :6060/debug/pprof/
```
