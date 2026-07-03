# New Cron Job Rules

**Read before:** Adding a scheduled background job.

## Checklist

- [ ] Add job function in `webapp/pkg/cronfn/` or inline in cron main
- [ ] Register in `webapp/cmd/cron/main.go`
- [ ] Add OpenTelemetry tracing
- [ ] Use singleton mode to prevent overlap
- [ ] Test manually before deploying

## File Locations

| Type | Location |
|------|----------|
| Cron scheduler | `webapp/cmd/cron/main.go` |
| Reusable job functions | `webapp/pkg/cronfn/` |

## Do

- Add tracing with `otel.Tracer().Start()`
- Use `gocron.WithSingletonMode()` to prevent overlapping runs
- Log job start/completion
- Handle errors gracefully (don't crash scheduler)
- Set appropriate job names for monitoring

## Don't

- Use `panic` in job functions
- Run jobs without tracing (invisible in observability)
- Forget singleton mode for long-running jobs

## Job Pattern

```go
const tracerName = "cron"

func registerMyJob(scheduler gocron.Scheduler, plugin *volt.Kernel) {
    _, err := scheduler.NewJob(
        // Schedule: every 5 minutes
        gocron.DurationJob(5*time.Minute),

        // Or cron expression: gocron.CronJob("0 */5 * * * *", false)
        // Or daily: gocron.DurationJob(24*time.Hour)

        gocron.NewTask(func() {
            ctx, span := otel.Tracer(tracerName).Start(
                context.Background(),
                "cron.myJob",
                trace.WithAttributes(
                    attribute.String("job.name", "myJob"),
                    attribute.String("job.type", "cleanup"),
                ),
            )
            defer span.End()

            span.AddEvent("job_started")

            // Job logic
            err := doJobWork(ctx, plugin)
            if err != nil {
                span.SetStatus(codes.Error, err.Error())
                span.RecordError(err)
                applog.FromCtx(ctx).Error("job failed", ulog.Err(err))
                return
            }

            span.SetStatus(codes.Ok, "success")
            span.AddEvent("job_completed")
            applog.FromCtx(ctx).Info("myJob completed")
        }),
        gocron.WithName("myJob"),
        gocron.WithSingletonMode(gocron.LimitModeReschedule),
    )
    if err != nil {
        fmt.Printf("Error scheduling myJob: %v\n", err)
    }
}
```

## Schedule Options

| Schedule | Code |
|----------|------|
| Every 5 minutes | `gocron.DurationJob(5*time.Minute)` |
| Every hour | `gocron.DurationJob(time.Hour)` |
| Daily at midnight | `gocron.CronJob("0 0 * * *", false)` |
| Weekly Monday 9am | `gocron.CronJob("0 9 * * 1", false)` |
| Every 30 seconds | `gocron.DurationJob(30*time.Second)` |

## Singleton Modes

| Mode | Behavior |
|------|----------|
| `LimitModeReschedule` | Skip if previous still running, reschedule |
| `LimitModeWait` | Queue and wait for previous to finish |

## Reusable Job Function

```go
// webapp/pkg/cronfn/cleanup.go
package cronfn

import (
    "context"
    "fmt"

    "dbent/gen/ent"
)

func CleanupOldRecords(ctx context.Context, client *ent.Client) error {
    deleted, err := client.TempRecord.
        Delete().
        Where(temprecord.CreatedAtLT(time.Now().AddDate(0, 0, -30))).
        Exec(ctx)
    if err != nil {
        return fmt.Errorf("cleanup old records: %w", err)
    }

    applog.FromCtx(ctx).Info("cleanup completed", ulog.Any("deleted", deleted))
    return nil
}
```

## Registration in main.go

```go
func main() {
    circuit := webapp.New(webapp.WithDisableNats(true))

    scheduler, err := gocron.NewScheduler(
        gocron.WithLogger(gocronlogger.New(applog.GetLogger())),
        gocron.WithMonitor(gocronmonitor.New(applog.GetLogger())),
        gocron.WithStopTimeout(24*time.Hour),
    )
    if err != nil {
        panic(err)
    }

    // Register jobs
    registerMyJob(scheduler, plugin)
    registerAnotherJob(scheduler, plugin)

    scheduler.Start()
    // ... shutdown handling
}
```
