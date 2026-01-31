## cacheredis

Redis-backed cache for the Jeoga cache plugin. The Redis client is injected as a dependency (`github.com/redis/go-redis/v9`).

### Run tests

Requires Redis (e.g. `task dev` or local Redis on `localhost:6379`). Tests use **DB 1** and skip if Redis is not available. A deferred `FlushDB` clears DB 1 when each test exits, so keys are not visible in redis-cli after the run. To inspect keys while debugging, comment out the `defer client.FlushDB(ctx)` in the test and then:

```bash
redis-cli -n 1 KEYS '*'
redis-cli -n 1 GET cacheredis_test_put_get
```

```bash
# From repo root
go test ./lace/cacheredis/... -v

# From saas/
cd saas && go test ./cacheredis/... -v
```
