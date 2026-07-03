# cronfn Package

**Path:** `webapp/pkg/cronfn/`
**Purpose:** Reusable cron/job functions that can be invoked from both scheduled cron jobs and CLI commands.

## Overview

The `cronfn` package provides a pattern for writing job functions that:
- Can be reused across cron scheduler and CLI commands (no code duplication)
- Include built-in OpenTelemetry tracing with spans and events
- Include structured logging with applog (github.com/ubgo/logger)
- Handle errors consistently
- Report stats on completion

## Pattern

Each function in `cronfn` follows this pattern:

```go
func JobName(ctx context.Context, plugin *volt.Kernel) error {
    // 1. Start trace span
    ctx, span := otel.Tracer(tracerName).Start(ctx, "cronfn.JobName",
        trace.WithAttributes(
            attribute.String("job.name", "jobName"),
            attribute.String("job.type", "sync"),
        ),
    )
    defer span.End()

    span.AddEvent("job_started")

    // 2. Initialize dependencies
    svc, err := github.NewService(&plugin.AppConfig.Github)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        applog.FromCtx(ctx).Error("Failed to init service", ulog.Err(err))
        return err
    }

    // 3. Execute job logic
    stats, err := doWork(ctx)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        applog.FromCtx(ctx).Error("Job failed", ulog.Err(err))
        return err
    }

    // 4. Record stats in span
    span.SetAttributes(
        attribute.Int("items_processed", stats.Processed),
        attribute.Int("items_success", stats.Success),
    )

    // 5. Log completion
    span.SetStatus(codes.Ok, "success")
    span.AddEvent("job_completed")
    applog.FromCtx(ctx).Info("Job completed", ulog.Any("processed", stats.Processed))

    return nil
}
```

## Available Functions

### SyncWildAI

Syncs pending files to WildAI service.

```go
func SyncWildAI(ctx context.Context, plugin *volt.Kernel) error
```

**What it does:**
1. Initializes GitHub service for fetching file contents
2. Creates `wildaisync.Syncer` instance
3. Calls `syncer.SyncPendingFiles(ctx)` to sync all pending files
4. Logs stats and errors

**Trace attributes:**
- `job.name`: "wildaiSync"
- `job.type`: "sync"
- `item_groups_processed`: count
- `files_processed`: count
- `files_success`: count
- `files_error`: count

### CleanDeletedFiles

Permanently removes soft-deleted item group files.

```go
func CleanDeletedFiles(ctx context.Context, plugin *volt.Kernel) error
```

**What it does:**
1. Finds all files with `status="deleted"`
2. For files **without** `wildai_source_id`: Deletes directly from database
3. For files **with** `wildai_source_id`:
   - Calls WildAI API (`ItemGroupSourceDelete`) to delete the source
   - If API call succeeds, deletes from database
   - If API call fails, skips the file (will retry next run)
4. Logs stats and errors

**Trace attributes:**
- `job.name`: "cleanDeletedFiles"
- `job.type`: "cleanup"
- `files_processed`: count
- `files_deleted`: count
- `files_deleted_direct`: count (files without wildai_source_id)
- `files_deleted_from_api`: count (files deleted via WildAI API first)
- `files_error`: count

## Usage

### From Cron Scheduler (`webapp/cmd/cron/main.go`)

```go
import "webapp/pkg/cronfn"

// Register WildAI sync job
_, err := scheduler.NewJob(
    gocron.CronJob("*/5 * * * *", false), // Every 5 minutes
    gocron.NewTask(func() {
        ctx := context.Background()
        _ = cronfn.SyncWildAI(ctx, plugin)
    }),
    gocron.WithName("wildai-sync"),
    gocron.WithSingletonMode(gocron.LimitModeReschedule),
)

// Register cleanup job
_, err = scheduler.NewJob(
    gocron.CronJob("*/10 * * * *", false), // Every 10 minutes
    gocron.NewTask(func() {
        ctx := context.Background()
        _ = cronfn.CleanDeletedFiles(ctx, plugin)
    }),
    gocron.WithName("clean-deleted-files"),
    gocron.WithSingletonMode(gocron.LimitModeReschedule),
)
```

### From CLI (`webapp/cmd/cli/main.go`)

```go
import "webapp/pkg/cronfn"

var syncCommand = &cobra.Command{
    Use:   "syncwildai",
    Short: "Sync pending files to WildAI",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.Background()
        err := cronfn.SyncWildAI(ctx, plugin)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    },
}
```

**Note:** The CLI currently has its own implementation in `syncwildaiCommand` that provides more detailed output. Consider migrating to use `cronfn.SyncWildAI` for consistency.

## Adding New Functions

When adding a new cron function:

1. Create the function in `webapp/pkg/cronfn/cronfn.go` (or a new file if complex)
2. Follow the pattern above with tracing, logging, and error handling
3. Register in cron scheduler if it should run on schedule
4. Optionally add CLI command for manual execution
5. Update this documentation

## Dependencies

- `webapp` - Plugin with config and database
- `webapp/pkg/github` - GitHub API service
- `webapp/pkg/wildaisync` - WildAI sync logic
- `core/pkg/applog` - Structured logging
- `go.opentelemetry.io/otel` - Tracing

## Related Files

| File | Purpose |
|------|---------|
| `webapp/pkg/cronfn/cronfn.go` | Cron function implementations |
| `webapp/cmd/cron/main.go` | Cron scheduler entry point |
| `webapp/cmd/cli/main.go` | CLI entry point |
| `webapp/pkg/wildaisync/` | WildAI sync logic |
| `webapp/` | Application plugin/config |
