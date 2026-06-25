# gocronui

A wrapper around [gocron-ui](https://github.com/go-co-op/gocron-ui) that adds internal-key authentication. Provides a web dashboard for monitoring and managing gocron scheduler jobs.

## Features

- Real-time job status via WebSocket
- Manual job triggering from the UI
- Next/last run times, schedule details
- Protected by internal key — no public access

## Pkl Configuration

The dashboard is controlled by the `app.cronUI` section in your pkl config.

### Schema (`config/pkl/schema/App.pkl`)

```pkl
class CronUI {
  enable: Boolean
  port: String
}
```

### Default config (`config/pkl/env/default/app.pkl`)

```pkl
app {
  cronUI {
    enable = true
    port = "8089"
  }
}
```

### Production config (`config/pkl/env/prod/app.pkl`)

```pkl
app {
  cronUI {
    enable = read("env:CRON_UI_ENABLE").toBoolean()
    port = read("env:CRON_UI_PORT")
  }
}
```

Production environment variables:

| Variable          | Example | Description                    |
|-------------------|---------|--------------------------------|
| `CRON_UI_ENABLE`  | `true`  | Enable/disable the dashboard   |
| `CRON_UI_PORT`    | `8089`  | TCP port the dashboard runs on |

Set `CRON_UI_ENABLE=false` in production to disable the dashboard entirely.

## Integration

### 1. Import the package

```go
import (
    "lace/gocron/gocronui"
)
```

### 2. Start the dashboard after `scheduler.Start()`

```go
// Start the scheduler
scheduler.Start()

// Start gocron-ui dashboard server
cronUICfg := p.RootConfig.App.CronUI
var httpSrv *http.Server
if cronUICfg.Enable {
    httpSrv = gocronui.NewServer(scheduler, gocronui.Config{
        Port:        cronUICfg.Port,
        Title:       p.RootConfig.Core.AppName + " Cron",
        InternalKey: p.RootConfig.Core.InternalKey,
    })
    gocronui.ListenAndLog(httpSrv, cronUICfg.Port, func(err error) {
        gozap.GetLogger().Error("gocron-ui server failed", zap.Error(err))
    })
}
```

### 3. Add graceful shutdown

Inside your shutdown handler, shut down the HTTP server **before** stopping the scheduler:

```go
shutdown.WithShutdownHandler(
    func(ctx context.Context) error {
        // Shutdown the gocron-ui HTTP server first
        if httpSrv != nil {
            if err := httpSrv.Shutdown(ctx); err != nil {
                gozap.GetLogger().Error("Error shutting down gocron-ui server", zap.Error(err))
            }
        }
        // Then stop scheduler jobs and shut down
        if err := scheduler.StopJobs(); err != nil {
            return err
        }
        return scheduler.Shutdown()
    },
),
```

## Accessing the Dashboard

Open the dashboard in your browser with the internal key as a query parameter:

```
http://localhost:8089/?key=<your-internal-key>
```

The internal key comes from `core.internalKey` in your pkl config. For the default local config this is `1234`, so:

```
http://localhost:8089/?key=1234
```

For programmatic access (e.g. from a script or monitoring tool), use the header instead:

```bash
curl -H "X-Internal-Key: 1234" http://localhost:8089/api/jobs
```

## Authentication Details

### Why cookies?

The gocron-ui dashboard is a single-page application. When you open it in a browser, the initial HTML page triggers many follow-up requests:

- Static assets: `app.js`, `app.css`, etc.
- REST API calls: `GET /api/jobs`, `GET /api/config`
- WebSocket upgrade: `GET /ws`

These requests are made by the browser automatically and cannot include a custom `?key=` query parameter or `X-Internal-Key` header.

To handle this, the middleware sets an `HttpOnly` session cookie on the first successful key-authenticated request. All subsequent requests from the same browser session include this cookie automatically, so static assets, API calls, and the WebSocket connection all pass through without requiring the key again.

### Cookie security

- The cookie value is an **HMAC-SHA256 signature** derived from the internal key — it cannot be guessed or forged without knowing the key.
- The cookie is **HttpOnly** — JavaScript on the page cannot read it.
- The cookie uses **SameSite=Lax** — it is not sent on cross-origin requests.
- The cookie has **no expiry** — it lasts for the browser session. Closing the browser clears it.

### Auth check order

1. **Cookie** — checked first for performance (most requests after initial auth)
2. **`?key=` query parameter** — primary method for browser access
3. **`X-Internal-Key` header** — for programmatic/API access

If none match, the request gets a `401 Unauthorized` response.
