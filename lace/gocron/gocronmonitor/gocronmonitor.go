package gocronmonitor

import (
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ZapMonitor implements gocron.MonitorStatus for logging job metrics to zap/OpenTelemetry
type ZapMonitor struct {
	logger *zap.Logger
}

// NewZapMonitor creates a new ZapMonitor
func New(logger *zap.Logger) *ZapMonitor {
	return &ZapMonitor{logger: logger.With(zap.String("component", "cron_monitor"))}
}

// IncrementJob logs job execution counts
func (m *ZapMonitor) IncrementJob(id uuid.UUID, name string, tags []string, status gocron.JobStatus) {
	m.logger.Info("job executed",
		zap.String("job_id", id.String()),
		zap.String("job_name", name),
		zap.String("tags", strings.Join(tags, ",")),
		zap.String("status", string(status)),
	)
}

// RecordJobTiming logs job timing metrics
func (m *ZapMonitor) RecordJobTiming(startTime, endTime time.Time, id uuid.UUID, name string, tags []string) {
	duration := endTime.Sub(startTime)
	m.logger.Info("job timing",
		zap.String("job_id", id.String()),
		zap.String("job_name", name),
		zap.String("tags", strings.Join(tags, ",")),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.Duration("duration", duration),
		zap.Float64("duration_ms", float64(duration.Milliseconds())),
	)
}

// RecordJobTimingWithStatus logs job timing with status and error details (MonitorStatus interface)
func (m *ZapMonitor) RecordJobTimingWithStatus(startTime, endTime time.Time, id uuid.UUID, name string, tags []string, status gocron.JobStatus, err error) {
	duration := endTime.Sub(startTime)
	fields := []zap.Field{
		zap.String("job_id", id.String()),
		zap.String("job_name", name),
		zap.String("tags", strings.Join(tags, ",")),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.Duration("duration", duration),
		zap.Float64("duration_ms", float64(duration.Milliseconds())),
		zap.String("status", string(status)),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		m.logger.Error("job completed with error", fields...)
	} else {
		m.logger.Info("job completed", fields...)
	}
}
