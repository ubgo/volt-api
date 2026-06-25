package gocronmonitor

import (
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	ulog "github.com/ubgo/logger"
)

// Monitor implements gocron.MonitorStatus, logging job metrics through
// github.com/ubgo/logger (and on to OpenTelemetry via its sinks).
type Monitor struct {
	lg *ulog.Logger
}

// New creates a Monitor backed by ubgo/logger.
func New(lg *ulog.Logger) *Monitor {
	return &Monitor{lg: lg.With(ulog.String("component", "cron_monitor"))}
}

// IncrementJob logs job execution counts.
func (m *Monitor) IncrementJob(id uuid.UUID, name string, tags []string, status gocron.JobStatus) {
	m.lg.Info("job executed",
		ulog.String("job_id", id.String()),
		ulog.String("job_name", name),
		ulog.String("tags", strings.Join(tags, ",")),
		ulog.String("status", string(status)),
	)
}

// RecordJobTiming logs job timing metrics.
func (m *Monitor) RecordJobTiming(startTime, endTime time.Time, id uuid.UUID, name string, tags []string) {
	duration := endTime.Sub(startTime)
	m.lg.Info("job timing",
		ulog.String("job_id", id.String()),
		ulog.String("job_name", name),
		ulog.String("tags", strings.Join(tags, ",")),
		ulog.Time("start_time", startTime),
		ulog.Time("end_time", endTime),
		ulog.Dur("duration", duration),
		ulog.Float("duration_ms", float64(duration.Milliseconds())),
	)
}

// RecordJobTimingWithStatus logs job timing with status and error details
// (the gocron.MonitorStatus interface).
func (m *Monitor) RecordJobTimingWithStatus(startTime, endTime time.Time, id uuid.UUID, name string, tags []string, status gocron.JobStatus, err error) {
	duration := endTime.Sub(startTime)
	fields := []ulog.Field{
		ulog.String("job_id", id.String()),
		ulog.String("job_name", name),
		ulog.String("tags", strings.Join(tags, ",")),
		ulog.Time("start_time", startTime),
		ulog.Time("end_time", endTime),
		ulog.Dur("duration", duration),
		ulog.Float("duration_ms", float64(duration.Milliseconds())),
		ulog.String("status", string(status)),
	}

	if err != nil {
		fields = append(fields, ulog.Err(err))
		m.lg.Error("job completed with error", fields...)
	} else {
		m.lg.Info("job completed", fields...)
	}
}
