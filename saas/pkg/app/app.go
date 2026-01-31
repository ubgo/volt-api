package app

import (
	"config"
	"config/gen/databaseconfig"
	"config/gen/loggerconfig"
	"config/gen/natsconfig"
	"config/gen/redisconfig"
	"config/gen/rootconfig"
	"context"
	"dbent"
	"lace/bootinfo"
	"lace/cacheredis"
	"lace/crypt"
	"lace/goerr"
	"lace/gotel"
	"lace/gozap"
	"log"
	"saas/pkg/fm/redisclient"
	"saas/pkg/pubsub"
	"saas/pkg/pubsublisten"

	"lace/cache"
	"lace/db"
	"lace/natso"

	"lace/cli"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "dbent/gen/ent/runtime"
)

type Plugin struct {
	Cli         cli.Cli
	DB          db.DB
	EntDB       dbent.EntDB
	Cache       cache.Store
	RootConfig  *rootconfig.RootConfig
	RedisClient *redis.Client
	Natso       *natso.Natso
}

var plugins *Plugin

func GetPlugins() *Plugin {
	if plugins == nil {
		panic("Plugins not initialized")
	}
	return plugins
}

func New(options ...AppPkgOption) *Plugin {
	cfg := apppkgconfig{
		DisableNats: false,
	}

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

	entClient := plugins.EntDB.Client()

	plugins.Cache = cacheredis.New(plugins.RedisClient)

	crypt.LoadKey(configroot.Core.EncryptionKey)

	pubsub.New()

	pubsublisten.SetupSubscribers(entClient)

	err = initNats(cfg.DisableNats, configroot.Nats, plugins, cfg.AfterNatsInitiated)
	if err != nil {
		panic("Error initializing NATS: " + err.Error())
	}

	err = initGlobalLogger(configroot.Otel, configroot.Logger)
	if err != nil {
		panic("Error initializing global logger: " + err.Error())
	}

	err = CustomInit(plugins)
	if err != nil {
		panic("Error initializing custom: " + err.Error())
	}

	return plugins
}

func ShutdownHandler(ctx context.Context) error {
	// Nats gracefully close
	// TODO: it closes the connection but wont wait for the existing messages to be processed so we still need to find a way to wait for the existing messages to be processed
	if plugins.Natso != nil {
		err := plugins.Natso.GetConn().Drain()
		if err != nil {
			log.Printf("Nats drain failed: %v", err)
		}
	}

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

	gozap.FromCtxInitCb(func(ctx context.Context) []zap.Field {

		fields := []zap.Field{}

		return fields
	})

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

/*
for dev i do not want to wait for the nats connection to be established and then start the server
so i will start the nats connection in a go routine to speed up the development
*/
func initNats(disableNats bool, natsconfig natsconfig.Setting, plugins *Plugin, afterNatsInitiated func(plugins *Plugin)) error {
	if disableNats || !natsconfig.Status {
		return nil
	}

	// if appcfg.Environment == constants.EnvironmentProd {
	// 	natsnats, err := natso.New(natso.WithAppKey(appcfg.Nats.AppKey), natso.WithURL(appcfg.Nats.Url), natso.WithEnabled(true), natso.WithCredsFile(appcfg.Nats.CredsFile))
	// 	if err != nil {
	// 		gozap.GetLogger().Sugar().Errorf("FATAL NATS CONNECTION - %s", err)
	// 		return err
	// 	}
	// 	plugins.Natso = natsnats

	// 	if afterNatsInitiated != nil {
	// 		afterNatsInitiated(plugins)
	// 	}

	// 	return nil
	// }

	go func(plugins *Plugin) {
		natsnats, err := natso.New(natso.WithAppKey(natsconfig.AppKey), natso.WithURL(natsconfig.Url), natso.WithEnabled(true), natso.WithCredsFile(natsconfig.CredsFile))
		if err != nil {
			gozap.GetLogger().Sugar().Errorf("FATAL NATS CONNECTION - %s", err)
			bootinfo.GetBootInfo().SetServiceStatus("nats", bootinfo.ServiceStatus{
				Status:  "error",
				Error:   err.Error(),
				Message: "NATS connection failed",
			})
			return
		}

		bootinfo.GetBootInfo().SetServiceStatus("nats", bootinfo.ServiceStatus{
			Status:  "ok",
			Error:   "",
			Message: "NATS connection successful",
		})

		plugins.Natso = natsnats

		if afterNatsInitiated != nil {
			afterNatsInitiated(plugins)
		}

	}(plugins)

	return nil
}
