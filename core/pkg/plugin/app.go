package plugin

import (
	"config"
	"config/gen/databaseconfig"
	"config/gen/loggerconfig"
	"config/gen/redisconfig"
	"config/gen/rootconfig"
	"context"
	"core/pkg/redisclient"
	"dbent"
	"encoding/json"
	"lace/bootinfo"
	"lace/crypt"
	"lace/gozap"
	"log"

	"lace/db"

	"lace/cli"

	"github.com/redis/go-redis/v9"
	"github.com/ubgo/cache"
	rediscache "github.com/ubgo/cache-redis"
	"github.com/ubgo/errx"
	"github.com/ubgo/otelkit"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	_ "dbent/gen/ent/runtime"
)

// errReporter fans structured error Reports out to registered sinks. It
// replaces goerr's package-global AddCallback. Reporting is explicit:
// call ReportError(ctx, err).
var errReporter = errx.NewRegistry()

// ReportError projects err into an errx.Report and emits it to every
// registered sink (no-op if err is not an *errx.Error).
func ReportError(ctx context.Context, err error) {
	errReporter.Report(ctx, err)
}

// errxGozapSink logs each error Report as structured JSON via gozap, so the
// key/values stay expandable in HyperDX instead of collapsing to a string.
type errxGozapSink struct{}

func (errxGozapSink) Name() string { return "gozap" }

func (errxGozapSink) Emit(ctx context.Context, r *errx.Report) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	gozap.FromCtx(ctx).Sugar().Errorw(r.Message, "result", string(data))
	return nil
}

type Plugin struct {
	Cli         cli.Cli
	DB          db.DB
	EntDB       dbent.EntDB
	Cache       cache.Cache
	RootConfig  *rootconfig.RootConfig
	RedisClient *redis.Client
}

var plugins *Plugin

// telemetry holds the initialized OTEL bootstrap handle (nil when telemetry is
// disabled), used by ShutdownHandler to flush on exit.
var telemetry *otelkit.Telemetry

func New(options ...AppPkgOption) *Plugin {
	cfg := apppkgconfig{}

	cfg.options(options...)

	cli := cli.New()

	configroot := config.Get()

	plugins = &Plugin{
		Cli:        cli,
		RootConfig: configroot,
	}

	err := initDB(configroot.Database, plugins)
	if err != nil {
		gozap.GetLogger().Sugar().Errorf("Error initializing DB: %v", err)
		panic(err)
	}

	err = initRedis(&configroot.Redis, plugins)
	if err != nil {
		gozap.GetLogger().Sugar().Errorf("Error initializing Redis: %v", err)
		panic(err)
	}

	// entClient := plugins.EntDB.Client()

	plugins.Cache = rediscache.New(plugins.RedisClient)

	crypt.LoadKey(configroot.Core.EncryptionKey)

	// pubsub.New()

	// pubsublisten.SetupSubscribers(entClient)

	err = initGlobalLogger(configroot.Otel, configroot.Logger)
	if err != nil {
		panic("Error initializing global logger: " + err.Error())
	}

	return plugins
}

func ShutdownHandler(ctx context.Context) error {
	// Step 1: Stop receiving HTTP traffic
	// if err := serverServer.Shutdown(shutdownCtx); err != nil {
	// 	log.Printf("HTTP server shutdown failed: %v", err)
	// }

	// DB close
	err := plugins.EntDB.Client().Close()
	if err != nil {
		log.Printf("DB close failed: %v", err)
	}

	// redis close
	err = plugins.RedisClient.Close()
	if err != nil {
		log.Printf("Redis close failed: %v", err)
	}

	// Step 2: Flush and shut down telemetry (ordered: logs, metrics, traces).
	if telemetry != nil {
		if err := telemetry.Shutdown(ctx); err != nil {
			log.Printf("OpenTelemetry shutdown failed: %v", err)
		}
	}
	return nil
}

func initGlobalLogger(otelconfig loggerconfig.Otel, loggerconfig loggerconfig.Setting) error {
	if !otelconfig.Enable {
		return nil
	}

	// Map the PKL otel config onto an otelkit Config. PKL is authoritative
	// (WithEnvOverrides(false)). gotel could tee stdout+http per signal; otelkit
	// is single-transport-per-signal, so HTTP wins when both are enabled.
	auth := otelconfig.ApiKey
	if otelconfig.PrefixBasic {
		auth = "Basic " + auth
	}
	headers := map[string]string{}
	if auth != "" {
		headers["authorization"] = auth
	}
	signal := func(enabled bool) otelkit.SignalConfig {
		sc := otelkit.SignalConfig{Enabled: enabled}
		switch {
		case otelconfig.EnableHttp:
			sc.Transport = otelkit.TransportHTTP
			sc.Endpoint = otelconfig.Endpoint
			sc.EndpointIsURL = otelconfig.IsEndpoingUrl
			sc.Headers = headers
		case otelconfig.EnableStdout:
			sc.Transport = otelkit.TransportStdout
		}
		return sc
	}
	tlsMode := otelkit.TLSModeTLS
	if otelconfig.Insecure {
		tlsMode = otelkit.TLSModePlaintext
	}

	tel, err := otelkit.Init(context.Background(),
		otelkit.WithConfig(otelkit.Config{
			ServiceName:    otelconfig.ServiceName,
			ServiceVersion: otelconfig.ServiceVersion,
			Traces:         signal(otelconfig.EnableTraces),
			Metrics:        signal(otelconfig.EnableMetrics),
			Logs:           signal(otelconfig.EnableLogs),
		}),
		otelkit.WithEnvOverrides(false),
		otelkit.WithTLS(tlsMode),
	)
	if err != nil {
		bootinfo.GetBootInfo().SetServiceStatus("otel", bootinfo.ServiceStatus{
			Status:  "error",
			Error:   err.Error(),
			Message: "OpenTelemetry initialization failed",
		})
		panic(err)
	}

	telemetry = tel
	tel.SetGlobal()

	// gozap's OTEL bridge wants the concrete SDK LoggerProvider; otelkit
	// exposes the API interface. Type-assert (nil when logs are disabled —
	// gozap guards against a nil provider).
	sdkLP, _ := tel.LoggerProvider().(*sdklog.LoggerProvider)

	gozap.InitGlobalLogger(gozap.Config{
		EnableStdout: loggerconfig.EnableStdout,
		EnableOtel:   loggerconfig.EnableOtel,
		EnableFile:   loggerconfig.EnableFile,
		LocalConfig: gozap.LocalConfig{
			Enabled: loggerconfig.LocalConfig.Enabled,
			LogFile: loggerconfig.LocalConfig.LogFile,
		},
		ServiceName:    otelconfig.ServiceName,
		LoggerProvider: sdkLP,
	})

	// gozap.FromCtxInitCb(func(ctx context.Context) []zap.Field {

	// 	fields := []zap.Field{}

	// 	return fields
	// })

	// Wire the structured error-reporting sink. Unlike goerr's old global
	// AddCallback (which auto-fired on Error()), errx is explicit: callers
	// report via ReportError(ctx, err), which fans out to this gozap sink.
	errReporter.Add(errxGozapSink{})

	return nil
}

func initRedis(rediscfg *redisconfig.Setting, plugins *Plugin) error {
	if !rediscfg.Status {
		return nil
	}

	redisClient, err := redisclient.New(rediscfg.Url, rediscfg.Db)

	// if error is related to the connection string then panic
	if err != nil {
		gozap.GetLogger().Sugar().Errorf("Error initializing Redis: %v", err)
		return err
	}

	// error is on the redis side maybe permission related errors
	err = redisclient.PingRedis(redisClient)
	if err != nil {
		bootinfo.GetBootInfo().SetServiceStatus("redis", bootinfo.ServiceStatus{
			Status:  "error",
			Error:   err.Error(),
			Message: "Redis connection failed",
		})
		return err
	}
	plugins.RedisClient = redisClient
	bootinfo.GetBootInfo().SetServiceStatus("redis", bootinfo.ServiceStatus{
		Status:  "ok",
		Error:   "",
		Message: "Redis connection successful",
	})
	return nil
}

func initDB(databaseconfig databaseconfig.Setting, plugins *Plugin) error {
	db := dbent.InitDB(databaseconfig)

	errdb := db.Ping()
	if errdb != nil {
		bootinfo.GetBootInfo().SetServiceStatus("database", bootinfo.ServiceStatus{
			Status:  "error",
			Error:   errdb.Error(),
			Message: "Database connection failed",
		})
	} else {
		bootinfo.GetBootInfo().SetServiceStatus("database", bootinfo.ServiceStatus{
			Status:  "ok",
			Error:   "",
			Message: "Database connection successful",
		})
	}

	entClient := dbent.New(db)

	plugins.DB = db
	plugins.EntDB = entClient

	return nil
}
