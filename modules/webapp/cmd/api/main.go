package main

import (
	"apidash"
	"apidash/pkg/apidashtypes"
	"config"
	"context"
	"core/pkg/applog"
	"core/pkg/coreapi"
	"core/pkg/volt"
	"lace/ginserver"
	"lace/httppkg/middlewares"
	"lace/util"
	"time"
	"webapp"
	"webapp/pkg/apphandlers"
	"webapp/pkg/shutlog"

	"github.com/gin-gonic/gin"
	healthgin "github.com/ubgo/health/contrib/health-gin"
	ulog "github.com/ubgo/logger"
	"github.com/ubgo/shutdown"
)

func main() {
	// The kernel loads the configuration during New, and Circuit embeds it, so
	// the config is read ONCE per process. Calling config.GetOrDie() here as
	// well would load a second, independent snapshot — wasted work, and two
	// values that could in principle disagree about what the environment said.
	circuit := webapp.New()
	appcfg := circuit.RootConfig

	handlerFuncs := []gin.HandlerFunc{
		util.CORSMiddleware(),
		middlewares.OtelMiddleware(appcfg.Otel.Enable, appcfg.Core.AppKey, nil),
	}

	serverServer := ginserver.New(
		ginserver.WithBeforeHandler(handlerFuncs...),
		ginserver.WithAppName(appcfg.Server.Hostname),
		ginserver.WithPort(appcfg.Server.Port),
		ginserver.WithIsProd(appcfg.Core.Environment == config.Prod),
	)

	apiGroup := serverServer.Router.Group("api")

	apidash.Boot(apiGroup, &apidashtypes.Resolver{
		Circuit:    circuit,
		RootConfig: appcfg,
	})

	coreapi.New(coreapi.Config{
		HttpEngine:     serverServer.Router,
		ApiRouterGroup: apiGroup,
		EntClient:      circuit.EntDB.Client(),
		CoreConfig:     &appcfg.Core,
	})

	// Health probes: /healthz (liveness), /readyz (readiness), /startupz —
	// all behind the internal-key guard. A background loop re-checks the
	// registered dependency checkers (db, redis) every 30s; the handlers
	// render the latest snapshot.
	healthgin.Mount(serverServer.Router, circuit.Health,
		healthgin.WithMiddleware(ginserver.MiddlewareInternalApiGuard(appcfg.Core.InternalKey)),
	)
	circuit.Health.Start(context.Background(), 30*time.Second)

	apphandlers.New(apphandlers.Config{
		HttpEngine:     serverServer.Router,
		ApiRouterGroup: apiGroup,
		EntClient:      circuit.EntDB.Client(),
		RootConfig:     circuit.RootConfig,
	})

	mgr := shutdown.New(
		shutdown.WithLogger(shutlog.New(applog.GetLogger())),
		shutdown.WithBudget(45*time.Second),
	)

	// Stop the background health re-check loop before closing clients, so a
	// re-check never pings an already-closed DB/Redis connection.
	_ = mgr.Register("health-checks", func(ctx context.Context) error {
		circuit.Health.Stop()
		return nil
	}, shutdown.WithPhase(shutdown.PhaseStopAccepting))

	// Stop accepting HTTP traffic first, then close downstream clients
	// (NATS drain, DB, Redis, OTEL flush) in a later phase.
	_ = mgr.Register("http-server", func(ctx context.Context) error {
		if err := serverServer.Shutdown(ctx); err != nil {
			applog.FromCtx(ctx).Error("HTTP server shutdown failed", ulog.Err(err))
			return err
		}
		return nil
	}, shutdown.WithPhase(shutdown.PhaseStopAccepting))

	_ = mgr.Register("core-circuit", volt.ShutdownHandler,
		shutdown.WithPhase(shutdown.PhaseCloseClients),
		shutdown.WithTimeout(30*time.Second))

	go serverServer.Start()

	if err := mgr.Listen(context.Background()); err != nil {
		applog.GetLogger().Error("shutdown completed with errors", ulog.Err(err))
	}
}
