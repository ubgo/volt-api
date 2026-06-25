package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"core/pkg/applog"
	"lace/gocron/gocronlogger"
	"lace/gocron/gocronmonitor"
	"lace/gocron/gocronui"

	"saas/pkg/app"
	"saas/pkg/shutlog"

	"github.com/go-co-op/gocron/v2"
	ulog "github.com/ubgo/logger"
	"github.com/ubgo/shutdown"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "cron"

type registerFunc func(scheduler gocron.Scheduler, p *app.CustomApp)

// allJobs defines all available cron jobs.
// The map key is the job name used in pkl config's enabledJobs list.
var allJobs = map[string]registerFunc{
	"exampleJob": registerExampleJob,
}

func main() {
	customApp := app.New()

	scheduler, err := gocron.NewScheduler(
		gocron.WithLogger(gocronlogger.New(applog.GetLogger())),
		gocron.WithMonitor(gocronmonitor.New(applog.GetLogger())),
		gocron.WithMonitorStatus(gocronmonitor.New(applog.GetLogger())),
		gocron.WithStopTimeout(24*time.Hour), // Wait up to 24h for jobs to complete on shutdown
	)
	if err != nil {
		fmt.Printf("Error creating scheduler: %v\n", err)
		return
	}

	// Register jobs based on config.
	// If enabledJobs is empty, all jobs are registered.
	gocronCfg := customApp.RootConfig.App.Gocron
	enabledJobs := gocronCfg.EnabledJobs

	if len(enabledJobs) == 0 {
		for name, register := range allJobs {
			register(scheduler, customApp)
			fmt.Printf("Registered job: %s\n", name)
		}
	} else {
		for _, name := range enabledJobs {
			register, ok := allJobs[name]
			if !ok {
				fmt.Printf("WARNING: unknown job name in enabledJobs: %q\n", name)
				continue
			}
			register(scheduler, customApp)
			fmt.Printf("Registered job: %s\n", name)
		}
	}

	// Start the scheduler
	scheduler.Start()

	// Start gocron-ui dashboard server
	var httpSrv *http.Server
	if gocronCfg.UiEnable {
		httpSrv = gocronui.NewServer(scheduler, gocronui.Config{
			Port:        gocronCfg.UiPort,
			Title:       customApp.RootConfig.Core.AppName + " Cron",
			InternalKey: customApp.RootConfig.Core.InternalKey,
		})
		gocronui.ListenAndLog(httpSrv, gocronCfg.UiPort, func(err error) {
			applog.GetLogger().Error("gocron-ui server failed", ulog.Err(err))
		})
	}

	fmt.Println("Cron scheduler started. Press Ctrl+C to exit.")

	mgr := shutdown.New(
		shutdown.WithLogger(shutlog.New(applog.GetLogger())),
		// Jobs may run long; allow a generous drain window (mirrors the old
		// gocron WithStopTimeout(24h)) before the watchdog force-exits.
		shutdown.WithBudget(24*time.Hour),
	)

	// Stop serving the dashboard first, then drain in-flight jobs.
	_ = mgr.Register("gocron-ui", func(ctx context.Context) error {
		if httpSrv == nil {
			return nil
		}
		return httpSrv.Shutdown(ctx)
	}, shutdown.WithPhase(shutdown.PhaseStopAccepting))

	_ = mgr.Register("scheduler", func(ctx context.Context) error {
		if err := scheduler.StopJobs(); err != nil {
			return err
		}
		if err := scheduler.Shutdown(); err != nil {
			return err
		}
		applog.FromCtx(ctx).Info("Scheduler jobs stopped gracefully")
		return nil
	}, shutdown.WithPhase(shutdown.PhaseDrainTraffic), shutdown.WithTimeout(24*time.Hour))

	if err := mgr.Listen(context.Background()); err != nil {
		applog.GetLogger().Error("shutdown completed with errors", ulog.Err(err))
	}
}

func registerExampleJob(scheduler gocron.Scheduler, customApp *app.CustomApp) {
	_, err := scheduler.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(func() {
			ctx, span := otel.Tracer(tracerName).Start(context.Background(), "cron.exampleJob",
				trace.WithAttributes(
					attribute.String("job.name", "exampleJob"),
					attribute.String("job.type", "example"),
				),
			)
			defer span.End()

			span.AddEvent("job_started")

			// Add your cron job logic here
			_ = customApp.EntDB.Client()

			span.SetStatus(codes.Ok, "success")
			span.AddEvent("job_completed")
			applog.FromCtx(ctx).Info("Example job completed")
		}),
		gocron.WithName("exampleJob"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		fmt.Printf("Error scheduling exampleJob: %v\n", err)
	}
}
