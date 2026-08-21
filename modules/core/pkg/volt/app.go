package volt

import (
	"config"
	coreenv "config/gen/coreconfig/environment"
	"config/gen/databaseconfig"
	"config/gen/loggerconfig"
	"config/gen/redisconfig"
	"config/gen/rootconfig"
	"context"
	"core/pkg/applog"
	"core/pkg/redisclient"
	"dbent"
	"encoding/json"
	"log"

	"lace/db"

	"lace/cli"

	"github.com/redis/go-redis/v9"
	"github.com/ubgo/cache"
	rediscache "github.com/ubgo/cache-redis"
	"github.com/ubgo/errx"
	"github.com/ubgo/health"
	healthredis "github.com/ubgo/health/contrib/health-redis"
	ulog "github.com/ubgo/logger"
	"github.com/ubgo/otelkit"

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
	applog.FromCtx(ctx).Error(r.Message, ulog.String("result", string(data)))
	return nil
}

type Kernel struct {
	Cli         cli.Cli
	DB          db.DB
	EntDB       dbent.EntDB
	Cache       cache.Cache
	RootConfig  *rootconfig.RootConfig
	RedisClient *redis.Client
	// Health is the dependency health-check registry. Live checkers for the
	// database and Redis are registered during New(); the api binary mounts
	// /healthz, /readyz, /startupz over it and runs a background re-check loop.
	Health *health.Registry
}

var kernel *Kernel

// telemetry holds the initialized OTEL bootstrap handle (nil when telemetry is
// disabled), used by ShutdownHandler to flush on exit.
var telemetry *otelkit.Telemetry

func New(options ...Option) *Kernel {
	cfg := apppkgconfig{}

	cfg.options(options...)

	cli := cli.New()

	configroot := config.Get()

	kernel = &Kernel{
		Cli:        cli,
		RootConfig: configroot,
		Health:     health.NewRegistry(),
	}

	err := initDB(configroot.Database, kernel)
	if err != nil {
		applog.GetLogger().Error("Error initializing DB", ulog.Err(err))
		panic(err)
	}

	err = initRedis(&configroot.Redis, kernel)
	if err != nil {
		applog.GetLogger().Error("Error initializing Redis", ulog.Err(err))
		panic(err)
	}

	// entClient := kernel.EntDB.Client()

	kernel.Cache = rediscache.New(kernel.RedisClient)

	// Fail closed: a non-dev deployment must never encrypt real data with an
	// empty or well-known dev-default key (the PKL template placeholders are
	// public knowledge). Refuse to boot rather than silently using a guessable
	// key — discovering this after rows are encrypted means a rekey migration.
	if configroot.Core.Environment != coreenv.Dev && IsWeakOrDefaultKey(configroot.Core.EncryptionKey) {
		panic("core.encryptionKey is empty or a dev-default placeholder in a non-dev environment (" + string(configroot.Core.Environment) + ") — refusing to boot; set a real 16/24/32-byte key via config or env")
	}

	// pubsub.New()

	// pubsublisten.SetupSubscribers(entClient)

	err = initGlobalLogger(configroot.Otel, configroot.Logger, string(configroot.Core.Environment))
	if err != nil {
		panic("Error initializing global logger: " + err.Error())
	}

	return kernel
}

func ShutdownHandler(ctx context.Context) error {
	// Step 1: Stop receiving HTTP traffic
	// if err := serverServer.Shutdown(shutdownCtx); err != nil {
	// 	log.Printf("HTTP server shutdown failed: %v", err)
	// }

	// DB close
	err := kernel.EntDB.Client().Close()
	if err != nil {
		log.Printf("DB close failed: %v", err)
	}

	// redis close
	err = kernel.RedisClient.Close()
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

func initGlobalLogger(otelconfig loggerconfig.Otel, loggerconfig loggerconfig.Setting, environment string) error {
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
			// deployment.environment.name — a SECONDARY discriminator. The
			// PRIMARY environment separation is the per-env service.name
			// convention (each env emits its own name, e.g. myapp-api vs
			// myapp-api-stag), because shared observability backends like
			// HyperDX have no tenant separation and alert rules scoped to a
			// prod service name must be structurally unable to match staging
			// noise. See volt/docs/guides/otel-service-naming.md.
			Environment: environment,
			Traces:      signal(otelconfig.EnableTraces),
			Metrics:     signal(otelconfig.EnableMetrics),
			Logs:        signal(otelconfig.EnableLogs),
		}),
		otelkit.WithEnvOverrides(false),
		otelkit.WithTLS(tlsMode),
	)
	if err != nil {
		panic(err)
	}

	telemetry = tel
	tel.SetGlobal()

	// applog's OTEL sink reads the global LoggerProvider that tel.SetGlobal()
	// just registered, so the logger ships logs through the same pipeline.
	applog.Init(applog.Config{
		EnableStdout: loggerconfig.EnableStdout,
		EnableOtel:   loggerconfig.EnableOtel,
		EnableFile:   loggerconfig.EnableFile && loggerconfig.LocalConfig.Enabled,
		ServiceName:  otelconfig.ServiceName,
		LogFile:      loggerconfig.LocalConfig.LogFile,
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

func initRedis(rediscfg *redisconfig.Setting, kernel *Kernel) error {
	if !rediscfg.Status {
		return nil
	}

	redisClient, err := redisclient.New(rediscfg.Url, rediscfg.Db)

	// if error is related to the connection string then panic
	if err != nil {
		applog.GetLogger().Error("Error initializing Redis", ulog.Err(err))
		return err
	}

	// error is on the redis side maybe permission related errors
	err = redisclient.PingRedis(redisClient)
	if err != nil {
		return err
	}
	kernel.RedisClient = redisClient

	// Register a live readiness checker; the background loop re-pings Redis
	// rather than relying on this boot-time outcome.
	_ = kernel.Health.Register(healthredis.New("redis", redisClient))
	return nil
}

func initDB(databaseconfig databaseconfig.Setting, kernel *Kernel) error {
	db := dbent.InitDB(databaseconfig)

	entClient := dbent.New(db)

	kernel.DB = db
	kernel.EntDB = entClient

	// Register a live readiness checker over the *sql.DB pool. Unlike the old
	// boot-time ping, the background loop re-checks it continuously.
	_ = kernel.Health.Register(newSQLDBChecker("database", db))

	return nil
}
