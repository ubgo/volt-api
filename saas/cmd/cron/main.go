package main

import (
	"context"
	"fmt"
	"time"

	"lace/gocron/gocronlogger"
	"lace/gocron/gocronmonitor"
	"lace/gozap"
	"lace/shutdown"

	"saas/pkg/app"

	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "cron"

func main() {
	plugin := app.New(app.WithDisableNats(true))

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

	// Register all jobs
	registerExampleJob(scheduler, plugin)

	// Start the scheduler
	scheduler.Start()

	fmt.Println("Cron scheduler started. Press Ctrl+C to exit.")

	shutdown.New(
		shutdown.WithMsgShuttingDown("Shutting down scheduler..."),
		shutdown.WithMsgShutdownCompleted("Scheduler stopped."),
		shutdown.WithMsgShutdownFailed("Error shutting down scheduler"),
		shutdown.WithLogger(gozap.GetLogger()),
		shutdown.WithGracefulShutdownTimeout(0),
		shutdown.WithShutdownHandler(
			func(ctx context.Context) error {
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

func registerExampleJob(scheduler gocron.Scheduler, plugin *app.Plugin) {
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
			_ = plugin.EntDB.Client()

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
