package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"core/pkg/plugin"
	"lace/gocron/gocronlogger"
	"lace/gocron/gocronmonitor"
	"lace/gocron/gocronui"
	"lace/gozap"
	"lace/shutdown"

	"saas/pkg/app"

	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const tracerName = "cron"

type registerFunc func(scheduler gocron.Scheduler, p *app.CustomApp)

// allJobs defines all available cron jobs.
// The map key is the job name used in pkl config's enabledJobs list.
var allJobs = map[string]registerFunc{
	"exampleJob": registerExampleJob,
}

func main() {
	customApp := app.New(app.WithCoreOptions(plugin.WithDisableNats(true)))

	scheduler, err := gocron.NewScheduler(
		gocron.WithLogger(gocronlogger.New(gozap.GetLogger())),
		gocron.WithMonitor(gocronmonitor.New(gozap.GetLogger())),
		gocron.WithMonitorStatus(gocronmonitor.New(gozap.GetLogger())),
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
	cronUICfg := customApp.RootConfig.App.CronUI
	var httpSrv *http.Server
	if cronUICfg.Enable {
		httpSrv = gocronui.NewServer(scheduler, gocronui.Config{
			Port:        cronUICfg.Port,
			Title:       customApp.RootConfig.Core.AppName + " Cron",
			InternalKey: customApp.RootConfig.Core.InternalKey,
		})
		gocronui.ListenAndLog(httpSrv, cronUICfg.Port, func(err error) {
			gozap.GetLogger().Error("gocron-ui server failed", zap.Error(err))
		})
	}

	fmt.Println("Cron scheduler started. Press Ctrl+C to exit.")

	shutdown.New(
		shutdown.WithMsgShuttingDown("Shutting down scheduler..."),
		shutdown.WithMsgShutdownCompleted("Scheduler stopped."),
		shutdown.WithMsgShutdownFailed("Error shutting down scheduler"),
		shutdown.WithLogger(gozap.GetLogger()),
		shutdown.WithGracefulShutdownTimeout(0),
		shutdown.WithShutdownHandler(
			func(ctx context.Context) error {
				// Shutdown the gocron-ui HTTP server
				if httpSrv != nil {
					if err := httpSrv.Shutdown(ctx); err != nil {
						gozap.GetLogger().Error("Error shutting down gocron-ui server", zap.Error(err))
					}
				}
				if err := scheduler.StopJobs(); err != nil {
					gozap.GetLogger().Error("Error stopping scheduler jobs")
					return err
				}
				if err := scheduler.Shutdown(); err != nil {
					gozap.GetLogger().Error("Error shutting down scheduler")
					return err
				}
				gozap.GetLogger().Info("Scheduler jobs stopped gracefully")
				return nil
			},
		),
	).Listen()
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
			gozap.FromCtx(ctx).Info("Example job completed")
		}),
		gocron.WithName("exampleJob"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		fmt.Printf("Error scheduling exampleJob: %v\n", err)
	}
}
