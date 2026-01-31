package main

import (
	"apidash"
	"apidash/pkg/apidashtypes"
	"config"
	"config/gen/coreconfig/environment"
	"context"
	"lace/ginserver"
	"lace/gozap"
	"lace/shutdown"
	"lace/util"
	"saas/cmd/api/natshandlers"
	"saas/pkg/app"
	"saas/pkg/apphandlers"
	"saas/pkg/httppkg/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	appcfg := config.Get()

	plugins := app.New(app.WithAfterNatsInitiated(func(plugins *app.Plugin) {
		natshandlers.New(natshandlers.Config{
			Natso: *plugins.Natso,
		})
	}))

	handlerFuncs := []gin.HandlerFunc{
		util.CORSMiddleware(),
		middlewares.OtelMiddleware(appcfg.Otel.Enable, appcfg.Core.AppKey),
	}

	serverServer := ginserver.New(
		ginserver.WithBeforeHandler(handlerFuncs...),
		ginserver.WithAppName(appcfg.Server.Hostname),
		ginserver.WithPort(appcfg.Server.Port),
		ginserver.WithIsProd(appcfg.Core.Environment == environment.Prod),
	)

	apiGroup := serverServer.Router.Group("api")

	apidash.Boot(apiGroup, &apidashtypes.Resolver{
		Plugin:     app.GetPlugins(),
		RootConfig: appcfg,
	})

	apphandlers.New(apphandlers.Config{
		HttpEngine:     serverServer.Router,
		ApiRouterGroup: apiGroup,
		EntClient:      plugins.EntDB.Client(),
		RootConfig:     plugins.RootConfig,
	})

	defer shutdown.New(
		shutdown.WithMsgShuttingDown("Shutting down API..."),
		shutdown.WithMsgShutdownCompleted("API stopped."),
		shutdown.WithMsgShutdownFailed("Error shutting down API"),
		shutdown.WithLogger(gozap.GetLogger()),
		shutdown.WithShutdownHandler(app.ShutdownHandler, func(ctx context.Context) error {
			// Step 1: Stop receiving HTTP traffic
			if err := serverServer.Shutdown(ctx); err != nil {
				gozap.FromCtx(ctx).Error("HTTP server shutdown failed", zap.Error(err))
				return err
			}
			return nil
		}),
	).Listen()

	go serverServer.Start()
}
