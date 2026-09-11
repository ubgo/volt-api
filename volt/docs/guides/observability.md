# Cross-Service OpenTelemetry Tracing

This document explains how distributed tracing works between `githubapp` and `wildaichat` microservices using OpenTelemetry.

## Overview

Distributed tracing allows you to follow a request as it travels across multiple services. Each service creates "spans" that are linked together by a common `trace_id`, forming a complete picture of the request flow.

```
┌─────────────────┐         HTTP Request         ┌─────────────────┐
│                 │   (with traceparent header)  │                 │
│   githubapp     │ ──────────────────────────►  │   wildaichat    │
│                 │                              │                 │
│  Parent Span    │                              │  Child Span     │
│  trace_id: abc  │                              │  trace_id: abc  │
│  span_id: 111   │                              │  span_id: 222   │
│                 │                              │  parent_id: 111 │
└─────────────────┘                              └─────────────────┘
```

## W3C Trace Context Standard

OpenTelemetry uses the [W3C Trace Context](https://www.w3.org/TR/trace-context/) standard for propagating trace information across services via HTTP headers.

### The `traceparent` Header

Format: `{version}-{trace_id}-{parent_span_id}-{trace_flags}`

Example:
```
traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
             │   │                                │                │
             │   │                                │                └─ trace flags (01 = sampled)
             │   │                                └─ parent span id (16 hex chars)
             │   └─ trace id (32 hex chars)
             └─ version (always 00)
```

### The `tracestate` Header (Optional)

Vendor-specific trace data. Example:
```
tracestate: vendor1=value1,vendor2=value2
```

## Implementation Details

### 1. Telemetry Setup (Both Services)

Both services initialize OpenTelemetry with **`github.com/ubgo/otelkit`** (the v2 bootstrap — `lace/gotel` is gone). One `otelkit.Init` call wires traces, metrics, and logs from `appcfg.Otel`, then `SetGlobal()` registers the global tracer, logger provider, and propagator. See the full bootstrap in the [Bootstrap section](#bootstrap--initializing-telemetry-logging--error-reporting) below.

```go
// modules/core/pkg/volt — bootstrap (both services)

tel, err := otelkit.Init(ctx,
    otelkit.WithConfig(otelkit.Config{
        ServiceName:    appcfg.Otel.ServiceName,
        ServiceVersion: appcfg.Otel.ServiceVersion,
        Traces:         signal(appcfg.Otel.EnableTraces),
        Metrics:        signal(appcfg.Otel.EnableMetrics),
        Logs:           signal(appcfg.Otel.EnableLogs),
    }),
    otelkit.WithEnvOverrides(false),
    otelkit.WithTLS(tlsMode),
)

tel.SetGlobal()  // global tracer, logger provider, and propagator
```

The critical part for cross-service tracing is the global propagator (otelkit registers a composite W3C propagator on `SetGlobal`):

```go
propagator := propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},  // W3C Trace Context
    propagation.Baggage{},       // W3C Baggage
)

otel.SetTextMapPropagator(propagator)
```

### 2. Service A: githubapp (CLI Command)

Location: `webapp/cmd/cli/main.go`

#### Step 1: Create Parent Span

```go
// Get the telemetry instance

// Start a trace span - this creates the parent span
ctx := context.Background()
ctx, span := otel.Tracer("app").Start(ctx, "testotel-cli-command")
span.SetAttributes(
    attribute.String("test.type", "cross-service-trace"),
    attribute.String("source.service", "githubapp"),
    attribute.String("target.service", "wildaichat"),
)
defer span.End()
```

#### Step 2: Log with Trace Context

```go
// applog.FromCtx extracts trace_id and span_id from context
applog.FromCtx(ctx).Info("Demo log from githubapp",
    ulog.String("step", "1"),
    ulog.String("source", "githubapp-cli"),
)
```

The `FromCtx` function automatically adds trace context to logs:

```go
// core/pkg/applog/applog.go

func FromCtx(ctx context.Context) *ulog.Logger {
    spanContext := trace.SpanContextFromContext(ctx)
    if !spanContext.IsValid() {
        return GetLogger()
    }

    fields := []ulog.Field{
        ulog.String("trace_id", spanContext.TraceID().String()),
        ulog.String("span_id", spanContext.SpanID().String()),
    }

    return GetLogger().With(fields...)
}
```

#### Step 3: Make HTTP Request with Automatic Trace Propagation

Using `otelhttp.NewTransport` is the cleanest approach - it automatically handles trace context injection and creates HTTP client spans:

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

req, err := http.NewRequestWithContext(ctx, http.MethodGet, wildaichatURL+"/test/otel", nil)

// Use otelhttp.NewTransport for automatic trace context propagation
// This automatically:
// - Injects traceparent/tracestate headers
// - Creates an HTTP client span
// - Records http.method, http.url, http.status_code attributes
client := &http.Client{
    Timeout:   10 * time.Second,
    Transport: otelhttp.NewTransport(nil),
}
resp, err := client.Do(req)
```

The transport automatically adds headers:
```
GET /test/otel HTTP/1.1
Host: localhost:2310
traceparent: 00-{trace_id}-{span_id}-01
tracestate: (if any vendor-specific data)
```

**Alternative: Manual Injection**

If you can't use `otelhttp`, you can manually inject:

```go
// Manual approach using headerCarrier adapter
otel.GetTextMapPropagator().Inject(ctx, headerCarrier(req.Header))

// headerCarrier adapts http.Header to propagation.TextMapCarrier
type headerCarrier http.Header

func (c headerCarrier) Get(key string) string {
    return http.Header(c).Get(key)
}

func (c headerCarrier) Set(key, value string) {
    http.Header(c).Set(key, value)
}

func (c headerCarrier) Keys() []string {
    keys := make([]string, 0, len(c))
    for k := range c {
        keys = append(keys, k)
    }
    return keys
}
```

### 3. Service B: wildaichat (HTTP Handler)

Location: `webapp/pkg/apphandlers/apphandlers.go`

#### Step 1: Extract Trace Context from Request

```go
func (pkg AppHandlers) testOtelHandler(c *gin.Context) {
    // Extract trace context from incoming request headers
    ctx := otel.GetTextMapPropagator().Extract(
        c.Request.Context(),
        headerCarrier(c.Request.Header),
    )
    // ctx now contains the trace context from githubapp
```

#### Step 2: Create Child Span

```go
    // Get telemetry and create a child span
    ctx, span := otel.Tracer("app").Start(ctx, "testotel-wildaichat-handler")
    span.SetAttributes(
        attribute.String("test.type", "cross-service-trace"),
        attribute.String("service.name", "wildaichat"),
        attribute.String("handler", "testOtelHandler"),
    )
    defer span.End()
```

This span automatically becomes a child of the githubapp span because:
1. We extracted the trace context into `ctx`
2. We passed that `ctx` to `otel.Tracer(...).Start`
3. The tracer reads the parent span info from the context

#### Step 3: Log with Same Trace Context

```go
    // Log demo message - will have same trace_id as githubapp
    applog.FromCtx(ctx).Info("Demo log from wildaichat",
        ulog.String("step", "2"),
        ulog.String("source", "wildaichat-handler"),
    )
```

## Data Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              GITHUBAPP CLI                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. otel.Tracer("app").Start(ctx, "testotel-cli-command")                           │
│     └─► Creates span with new trace_id: abc123, span_id: span001            │
│                                                                              │
│  2. applog.FromCtx(ctx).Info("Demo log from githubapp")                      │
│     └─► Log sent to observability platform with trace_id: abc123            │
│                                                                              │
│  3. otel.GetTextMapPropagator().Inject(ctx, headerCarrier(req.Header))      │
│     └─► Adds header: traceparent: 00-abc123-span001-01                      │
│                                                                              │
│  4. client.Do(req) ─────────────────────────────────────────────────────────┼──┐
│                                                                              │  │
└──────────────────────────────────────────────────────────────────────────────┘  │
                                                                                   │
                              HTTP Request with traceparent header                 │
                                                                                   │
┌──────────────────────────────────────────────────────────────────────────────┐  │
│                              WILDAICHAT API                                   │  │
├──────────────────────────────────────────────────────────────────────────────┤  │
│                                                                              │  │
│  1. otel.GetTextMapPropagator().Extract(ctx, headerCarrier(headers))    ◄───┼──┘
│     └─► Extracts trace_id: abc123, parent_span_id: span001 into ctx         │
│                                                                              │
│  2. otel.Tracer("app").Start(ctx, "testotel-wildaichat-handler")                    │
│     └─► Creates child span with trace_id: abc123, span_id: span002          │
│         parent_span_id: span001                                              │
│                                                                              │
│  3. applog.FromCtx(ctx).Info("Demo log from wildaichat")                     │
│     └─► Log sent to observability platform with trace_id: abc123            │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Observability Platform View

In HyperDX/Grafana/Jaeger, you'll see:

### Trace View
```
Trace: abc123
├── testotel-cli-command (githubapp) ─────────────────────────────────
│   Duration: 150ms
│   Attributes:
│     test.type: cross-service-trace
│     source.service: githubapp
│
└── testotel-wildaichat-handler (wildaichat) ─────────────────────────
    Duration: 50ms
    Parent: testotel-cli-command
    Attributes:
      test.type: cross-service-trace
      service.name: wildaichat
```

### Logs View
Both logs will be linked by the same `trace_id`:

| Timestamp | Service    | Message                    | trace_id | span_id  |
|-----------|------------|----------------------------|----------|----------|
| 10:00:01  | githubapp  | Demo log from githubapp    | abc123   | span001  |
| 10:00:01  | wildaichat | Demo log from wildaichat   | abc123   | span002  |

## Testing

### Run the Test

```bash
# Terminal 1: Start wildaichat
cd /Volumes/D/www/projects/thewildai/thewildai_chat_go
go run ./webapp/cmd/api

# Terminal 2: Run the test
cd /Volumes/D/www/projects/thewildai/wildai_githubapp
go run ./webapp/cmd/cli testotel
```

### Expected CLI Output

```
=== Cross-Service OpenTelemetry Test ===

Trace ID:  abc123def456...
Span ID:   111222333...

--- Step 1: Logging from githubapp ---
  [OK] Logged demo message from githubapp

--- Step 2: Calling wildaichat API ---
  URL: http://localhost:2310/test/otel
  traceparent: 00-abc123def456...-111222333...-01
  [OK] Response status: 200 OK

=== Test Complete ===

Check your observability platform (HyperDX/Grafana) for:
  - Trace ID: abc123def456...
  - Two linked logs: one from githubapp, one from wildaichat
  - Parent-child span relationship

Waiting for telemetry to flush...
```

## Key Components

| Component | Package | Purpose |
|-----------|---------|---------|
| `otelkit.Init` | `github.com/ubgo/otelkit` | Initialize OpenTelemetry (traces/metrics/logs) + set globals |
| `otel.Tracer(name).Start` | `go.opentelemetry.io/otel` | Create new spans (standard OTel API) |
| `applog.FromCtx` | `core/pkg/applog` (`github.com/ubgo/logger`) | Log with trace context |
| `otelhttp.NewTransport` | `go.opentelemetry.io/contrib/.../otelhttp` | Auto trace propagation for HTTP clients |
| `otel.GetTextMapPropagator().Inject` | `go.opentelemetry.io/otel` | Manual: inject trace context into headers |
| `otel.GetTextMapPropagator().Extract` | `go.opentelemetry.io/otel` | Extract trace context from headers (server) |
| `headerCarrier` | Local adapter | Adapts `http.Header` to `TextMapCarrier` |

## Configuration

Both services need matching OTEL configuration. The fields live on the `Otel` struct in `modules/config/config.go`; each is bound to an environment variable, so a deployment sets them without touching code:

```sh
OTEL_SERVICE_NAME=githubapp        # or wildaichat; defaults to the app key
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENABLE_LOGS=true
OTEL_ENABLE_TRACES=true
OTEL_ENABLE_METRICS=false
OTEL_ENABLE_STDOUT=false
OTEL_ENABLE_HTTP=true
OTEL_ENDPOINT=https://your-otel-endpoint
OTEL_API_KEY=your-api-key           # secret: masked in config explain
```

`OTEL_API_KEY` is tagged `secret:"true"`, so it never appears in `task config:explain` output or in an error message. Every key above is listed with its default in the generated `.env.example` — see [Configuration](configuration.md).

## Troubleshooting

### Traces Not Linked

1. **Check propagator is set globally**: Ensure `telem.SetGlobal()` is called
2. **Check header injection**: Print `req.Header.Get("traceparent")` before sending
3. **Check header extraction**: Log the incoming headers in wildaichat

### Logs Missing trace_id

1. **Use `applog.FromCtx(ctx)`**: Regular `applog.GetLogger()` won't include trace context
2. **Ensure span is active**: The context must have an active span

### Spans Not Appearing

1. **Check telemetry flush**: Add `time.Sleep(3 * time.Second)` before exit
2. **Verify OTEL endpoint**: Check the endpoint URL and API key
3. **Enable stdout exporter**: Set `OTEL_ENABLE_STDOUT=true` to see local output


## Bootstrap — initializing telemetry, logging & error reporting


We have 3 packages (all migrated to `github.com/ubgo/*` in v2):

- **otelkit** (`github.com/ubgo/otelkit`) — OpenTelemetry instrumentation (traces, metrics, logs pipeline). Replaces the old `gotel`.
- **logger** (`github.com/ubgo/logger`, wired via `core/pkg/applog`) — structured logging to console / OTEL / file. Replaces the old `core/pkg/applog`.
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
    otelkit.WithEnvOverrides(false), // appcfg is the single authority
    otelkit.WithTLS(tlsMode),
)
if err != nil {
    panic(err)
}

// register the tracer + logger provider as the OpenTelemetry globals
tel.SetGlobal()
```

`signal(enabled)` maps `appcfg.Otel` onto an `otelkit.SignalConfig` (HTTP vs stdout transport, endpoint, headers). otelkit is single-transport-per-signal, so when both HTTP and stdout are enabled, HTTP wins.

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
