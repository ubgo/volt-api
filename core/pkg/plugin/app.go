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
	"lace/bootinfo"
	"lace/crypt"
	"lace/goerr"
	"lace/gotel"
	"lace/gozap"
	"log"

	"lace/db"

	"lace/cli"

	"github.com/redis/go-redis/v9"
	"github.com/ubgo/cache"
	rediscache "github.com/ubgo/cache-redis"

	_ "dbent/gen/ent/runtime"
)

type Plugin struct {
	Cli         cli.Cli
	DB          db.DB
	EntDB       dbent.EntDB
	Cache       cache.Cache
	RootConfig  *rootconfig.RootConfig
	RedisClient *redis.Client
}

var plugins *Plugin

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

	// Step 2: Wait for service-level background tasks to complete
	if err := gotel.MustGetTelemetry().Shutdown(ctx); err != nil {
		log.Printf("OpenTelemetry shutdown failed: %v", err)
	}
	return nil
}

func initGlobalLogger(otelconfig loggerconfig.Otel, loggerconfig loggerconfig.Setting) error {
	if !otelconfig.Enable {
		return nil
	}

	telem, err := gotel.NewTelemetryWithPreExporters(context.Background(), gotel.PreExporterConfig{
		Config: gotel.Config{
			ServiceName:    otelconfig.ServiceName,
			ServiceVersion: otelconfig.ServiceVersion,
			EnableLogs:     otelconfig.EnableLogs,
			EnableTraces:   otelconfig.EnableTraces,
			EnableMetrics:  otelconfig.EnableMetrics,
		},
		EnableStdout:  otelconfig.EnableStdout,
		EnableHttp:    otelconfig.EnableHttp,
		Endpoint:      otelconfig.Endpoint,
		IsEndpoingUrl: otelconfig.IsEndpoingUrl,
		PrefixBasic:   otelconfig.PrefixBasic,
		ApiKey:        otelconfig.ApiKey,
		Insecure:      otelconfig.Insecure,
	})
	if err != nil {
		bootinfo.GetBootInfo().SetServiceStatus("otel", bootinfo.ServiceStatus{
			Status:  "error",
			Error:   err.Error(),
			Message: "OpenTelemetry initialization failed",
		})
		panic(err)
	}

	telem.SetGlobal()

	gozap.InitGlobalLogger(gozap.Config{
		EnableStdout: loggerconfig.EnableStdout,
		EnableOtel:   loggerconfig.EnableOtel,
		EnableFile:   loggerconfig.EnableFile,
		LocalConfig: gozap.LocalConfig{
			Enabled: loggerconfig.LocalConfig.Enabled,
			LogFile: loggerconfig.LocalConfig.LogFile,
		},
		ServiceName:    otelconfig.ServiceName,
		LoggerProvider: telem.GetLoggerProvider(),
	})

	// gozap.FromCtxInitCb(func(ctx context.Context) []zap.Field {

	// 	fields := []zap.Field{}

	// 	return fields
	// })

	goerr.AddCallback(func(ctx context.Context, resp *goerr.ErrorResult) {
		// send as JSON to hyperdx so it we can expand the key, values otherwise it will be a string
		gozap.FromCtx(ctx).Sugar().Errorw(resp.Message, "result", resp.ToJSON())
	})

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
