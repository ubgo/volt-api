## Intrumentation

We have 3 packages
- gotel - opentelemetry instrumentation pkg
- gozap - for logging to file, otel etc
- goerr - for hiding senstive errors and send custom errors to the user and log the actual errors to the otel


## init telementry
```go

	telem, err := gotel.NewTelemetryWithPreExporters(context.Background(), gotel.PreExporterConfig{
		Config: gotel.Config{
			ServiceName:    appcfg.Otel.ServiceName,
			ServiceVersion: appcfg.Otel.ServiceVersion,
			EnableLogs:     appcfg.Otel.EnableLogs,
			EnableTraces:   appcfg.Otel.EnableTraces,
			EnableMetrics:  appcfg.Otel.EnableMetrics,
		},
		EnableStdout:  appcfg.Otel.EnableStdout,
		EnableHttp:    appcfg.Otel.EnableHttp,
		Endpoint:      appcfg.Otel.Endpoint,
		IsEndpoingUrl: appcfg.Otel.IsEndpoingUrl,
		PrefixBasic:   appcfg.Otel.PrefixBasic,
		ApiKey:        appcfg.Otel.ApiKey,
		Insecure:      appcfg.Otel.Insecure,
	})
	if err != nil {
		panic(err)
	}

  // this will setn the logger, tracer to global vars in opentelemetry go
	telem.SetGlobal()
```


## init gozap
```go
	gozap.InitGlobalLogger(gozap.Config{
		EnableStdout: appcfg.Logger.EnableStdout,
		EnableOtel:   appcfg.Logger.EnableOtel,
		EnableFile:   appcfg.Logger.EnableFile,
		LocalConfig: gozap.LocalConfig{
			Enabled: appcfg.Logger.LocalConfig.Enabled,
			LogFile: appcfg.Logger.LocalConfig.LogFile,
		},
		ServiceName:    appcfg.Otel.ServiceName,
		LoggerProvider: telem.GetLoggerProvider(),
	})

  // let say we want to apped some global values every time send log to otel so this is how you do it
  // extracting the values from gin gonic context
	gozap.FromCtxInitCb(func(ctx context.Context) []zap.Field {
		workspaceID, _ := middlewares.GetWorkspaceID(ctx)

		fields := []zap.Field{}

		if len(workspaceID) > 0 {
			fields = append(fields, zap.String(constants.LogWorkspaceIdKey, workspaceID))
		}

		return fields
	})

  // integrating the goerr with gozap let say everytime we send an error to api response we also want to log it to otel also
	goerr.AddCallback(func(ctx context.Context, resp *goerr.ErrorResult) {
		// send as JSON to hyperdx so it we can expand the key, values otherwise it will be a string
		gozap.FromCtx(ctx).Sugar().Errorw(resp.Message, "result", resp.ToJSON())
	})
``


## Few other things

### add otel tracing middelware to graphql gqlgen
use `"github.com/ravilushqa/otelgqlgen"` pkg
```go
	gserver := handler.New(schm)
  gserver.Use(otelgqlgen.Middleware())

```

### for gin gonic
use `"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"`
or check OtelMiddleware how i implemented it