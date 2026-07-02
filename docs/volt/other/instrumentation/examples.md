## Some Examples

```go
applog.FromCtx(ctx).Info("chatbotSetFavorite", ulog.String("id", id), ulog.Bool("isFavorite", isFavorit

return false, goerr.New(err.Error(), goerr.WithMessageUser("chatbot not found")).ErrorUser()
```

## Traces using Global Otel

```go
import (
	"context"
	"fmt"
	"core/pkg/applog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	ulog "github.com/ubgo/logger"
)

func() {
	ctx, span := otel.Tracer(tracerName).Start(context.Background(), "cron.processFiles",
		trace.WithAttributes(
			attribute.String("job.name", "processFiles"),
			attribute.String("job.type", "file_processing"),
		),
	)
	// defer span.End()

	span.AddEvent("job_started")

	entClient := plugin.EntDB.Client()
	if err := chatbotsourcefn.ProcessFiles(ctx, entClient, plugin.RootConfig); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		applog.FromCtx(ctx).Error("Error processing files", ulog.Err(err))
		return
	}

	span.SetStatus(codes.Ok, "success")
	span.AddEvent("job_completed")

	applog.FromCtx(ctx).Info("Successfully processed pending files")
	span.End()

	time.Sleep(5 * time.Minute)
}
```

## Logging with TraceCtx

```go
	traceCtx, span := telemetry.TraceStart(ctx, "test-http1")

	span.SetAttributes(attribute.String("test11", "test"))
	span.End()

	testMulitpleMicroservice(traceCtx)

	// applog.GetLogger().Info("test log1")
	applog.FromCtx(traceCtx).Info("test log1")
```

```go
package main

import (
	"context"
	"fmt"
	"io"
	"lace/gotel"
	"core/pkg/applog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// testStdout()
	testHttp()
}

func testMulitpleMicroservice(ctx context.Context) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:9001/ping", nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(nil),
	}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func testHttp() {
	ctx := context.Background()

	cfg := gotel.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		EnableLogs:     true,
		EnableTraces:   true,
		EnableMetrics:  false,
	}

	telemetry, err := gotel.NewTelemetryWithPreExporters(ctx, gotel.PreExporterConfig{
		Config:       cfg,
		EnableStdout: true,
		EnableHttp:   true,
		Endpoint:     "in-otel.hyperdx.io",
		ApiKey:       "xxxx",
		Insecure:     false,
	})
	if err != nil {
		panic(err)
	}

	telemetry.SetGlobal()

	defer telemetry.Shutdown(ctx)

	// applog.Init runs after SetGlobal so its OTEL sink picks up the global LoggerProvider.
	applog.Init(applog.Config{
		EnableStdout: true,
		EnableOtel:   true,
		EnableFile:   true,
		ServiceName:  "test",
		LogFile:      "./logs/app.log",
	})

	traceCtx, span := telemetry.TraceStart(ctx, "test-http1")

	span.SetAttributes(attribute.String("test11", "test"))
	span.End()

	testMulitpleMicroservice(traceCtx)

	// applog.GetLogger().Info("test log1")
	applog.FromCtx(traceCtx).Info("test log1")

	time.Sleep(time.Second * 30)
}

func testStdout() {

	ctx := context.Background()

	cfg := gotel.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		EnableLogs:     true,
		EnableTraces:   false,
		EnableMetrics:  false,
	}

	telemetry, err := gotel.NewTelemetryWithPreExporters(ctx, gotel.PreExporterConfig{
		Config:       cfg,
		EnableStdout: true,
	})
	if err != nil {
		panic(err)
	}

	telemetry.SetGlobal()

	defer telemetry.Shutdown(ctx)

	_, span := telemetry.TraceStart(ctx, "test")

	span.SetAttributes(attribute.String("test", "test"))
	span.End()

	time.Sleep(time.Second * 10)
}
```
