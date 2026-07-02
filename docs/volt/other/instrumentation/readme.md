## Instrumentation

We have 3 packages (all migrated to `github.com/ubgo/*` in v2):

- **otelkit** (`github.com/ubgo/otelkit`) — OpenTelemetry instrumentation (traces, metrics, logs pipeline). Replaces the old `gotel`.
- **logger** (`github.com/ubgo/logger`, wired via `core/pkg/applog`) — structured logging to console / OTEL / file. Replaces the old `lace/gozap`.
- **errx** (`github.com/ubgo/errx`) — hiding sensitive errors, returning custom user-facing errors, and reporting the real error to OTEL. Replaces the old `goerr`.

## init telemetry

```go
tel, err := otelkit.Init(context.Background(),
    otelkit.WithConfig(otelkit.Config{
        ServiceName:    appcfg.Otel.ServiceName,
        ServiceVersion: appcfg.Otel.ServiceVersion,
        Traces:         signal(appcfg.Otel.EnableTraces),
        Metrics:        signal(appcfg.Otel.EnableMetrics),
        Logs:           signal(appcfg.Otel.EnableLogs),
    }),
    otelkit.WithEnvOverrides(false), // PKL config is authoritative
    otelkit.WithTLS(tlsMode),
)
if err != nil {
    panic(err)
}

// register the tracer + logger provider as the OpenTelemetry globals
tel.SetGlobal()
```

`signal(enabled)` maps the PKL otel config onto an `otelkit.SignalConfig` (HTTP vs stdout transport, endpoint, headers). otelkit is single-transport-per-signal, so when both HTTP and stdout are enabled, HTTP wins.

## init logging (applog)

`applog.Init` must run **after** `tel.SetGlobal()` — its OTEL sink reads the global `LoggerProvider` that otelkit just registered, so logs ship through the same pipeline.

```go
applog.Init(applog.Config{
    EnableStdout: appcfg.Logger.EnableStdout,
    EnableOtel:   appcfg.Logger.EnableOtel,
    EnableFile:   appcfg.Logger.EnableFile && appcfg.Logger.LocalConfig.Enabled,
    ServiceName:  appcfg.Otel.ServiceName,
    LogFile:      appcfg.Logger.LocalConfig.LogFile,
})
```

Then log through the global or the context-aware accessor:

```go
applog.GetLogger().Info("server started", ulog.String("port", port))

// FromCtx auto-adds trace_id + span_id from the active span, so logs
// correlate with traces — no manual field wiring needed.
applog.FromCtx(ctx).Error("request failed", ulog.Err(err))
```

## integrating errx error reporting

Unlike `goerr`'s old global `AddCallback` (which auto-fired on `Error()`), **errx is explicit**: callers report via `ReportError(ctx, err)`, which fans out to a registered sink. Wire a sink that logs the structured error through applog to OTEL:

```go
// register once during bootstrap
errReporter.Add(errxLogSink{})

// the sink logs the reported error as JSON so keys/values stay expandable in the backend
func (errxLogSink) Report(ctx context.Context, r *errx.Report) {
    data, _ := json.Marshal(r.Safe())
    applog.FromCtx(ctx).Error(r.Message, ulog.String("result", string(data)))
}
```

## Few other things

### add otel tracing middleware to graphql gqlgen
use `"github.com/ravilushqa/otelgqlgen"` pkg
```go
gserver := handler.New(schm)
gserver.Use(otelgqlgen.Middleware())
```

### for gin gonic
use `"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"`
or check `OtelMiddleware` for how it's implemented.
