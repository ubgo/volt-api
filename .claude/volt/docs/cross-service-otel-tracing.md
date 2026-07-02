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

Both services use the same `lace/gotel` package to initialize OpenTelemetry:

```go
// saas/pkg/app/app.go (both services)

telem, err := gotel.NewTelemetryWithPreExporters(context.Background(), gotel.PreExporterConfig{
    Config: gotel.Config{
        ServiceName:    appcfg.Otel.ServiceName,    // "githubapp" or "wildaichat"
        ServiceVersion: appcfg.Otel.ServiceVersion,
        EnableLogs:     appcfg.Otel.EnableLogs,
        EnableTraces:   appcfg.Otel.EnableTraces,
        EnableMetrics:  appcfg.Otel.EnableMetrics,
    },
    EnableStdout:  appcfg.Otel.EnableStdout,
    EnableHttp:    appcfg.Otel.EnableHttp,
    Endpoint:      appcfg.Otel.Endpoint,      // e.g., HyperDX or Grafana endpoint
    ApiKey:        appcfg.Otel.ApiKey,
    // ... other config
})

telem.SetGlobal()  // Sets global tracer, logger, and propagator
```

The critical part for cross-service tracing is setting the global propagator:

```go
// lace/gotel/gotel.go

propagator := propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},  // W3C Trace Context
    propagation.Baggage{},       // W3C Baggage
)

otel.SetTextMapPropagator(propagator)
```

### 2. Service A: githubapp (CLI Command)

Location: `saas/cmd/cli/main.go`

#### Step 1: Create Parent Span

```go
// Get the telemetry instance
telem := gotel.MustGetTelemetry()

// Start a trace span - this creates the parent span
ctx := context.Background()
ctx, span := telem.TraceStart(ctx, "testotel-cli-command")
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

Location: `saas/pkg/apphandlers/apphandlers.go`

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
    telem := gotel.MustGetTelemetry()
    ctx, span := telem.TraceStart(ctx, "testotel-wildaichat-handler")
    span.SetAttributes(
        attribute.String("test.type", "cross-service-trace"),
        attribute.String("service.name", "wildaichat"),
        attribute.String("handler", "testOtelHandler"),
    )
    defer span.End()
```

This span automatically becomes a child of the githubapp span because:
1. We extracted the trace context into `ctx`
2. We passed that `ctx` to `TraceStart`
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
│  1. telem.TraceStart(ctx, "testotel-cli-command")                           │
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
│  2. telem.TraceStart(ctx, "testotel-wildaichat-handler")                    │
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
go run ./saas/cmd/api

# Terminal 2: Run the test
cd /Volumes/D/www/projects/thewildai/wildai_githubapp
go run ./saas/cmd/cli testotel
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
| `gotel.NewTelemetryWithPreExporters` | `lace/gotel` | Initialize OpenTelemetry with exporters |
| `telem.TraceStart` | `lace/gotel` | Create new spans |
| `applog.FromCtx` | `core/pkg/applog` | Log with trace context |
| `otelhttp.NewTransport` | `go.opentelemetry.io/contrib/.../otelhttp` | Auto trace propagation for HTTP clients |
| `otel.GetTextMapPropagator().Inject` | `go.opentelemetry.io/otel` | Manual: inject trace context into headers |
| `otel.GetTextMapPropagator().Extract` | `go.opentelemetry.io/otel` | Extract trace context from headers (server) |
| `headerCarrier` | Local adapter | Adapts `http.Header` to `TextMapCarrier` |

## Configuration

Both services need matching OTEL configuration in their config files:

```pkl
Otel {
  ServiceName = "githubapp"  // or "wildaichat"
  ServiceVersion = "1.0.0"
  EnableLogs = true
  EnableTraces = true
  EnableMetrics = false
  EnableStdout = false
  EnableHttp = true
  Endpoint = "https://your-otel-endpoint"
  ApiKey = "your-api-key"
}
```

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
3. **Enable stdout exporter**: Set `EnableStdout = true` to see local output
