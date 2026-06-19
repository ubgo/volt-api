// Package applog is the application's logging bootstrap, built on
// github.com/ubgo/logger. It replaces the old lace/gozap zap wrapper:
// ubgo/logger is the single logging engine, and this package wires its sinks
// (console, OTEL log bridge, file) and exposes the global + a context-aware
// accessor that correlates logs with the active trace.
//
// The OTEL sink reads the global LoggerProvider, so Init must run after the
// telemetry pipeline has been registered (otelkit's SetGlobal).
package applog

import (
	"context"
	"os"

	"github.com/ubgo/logger"
	otelsink "github.com/ubgo/logger/contrib/otel"
	"go.opentelemetry.io/otel/trace"
)

// Config controls which sinks the global logger fans out to.
type Config struct {
	EnableStdout bool
	EnableOtel   bool
	EnableFile   bool
	ServiceName  string
	LogFile      string
}

// global is the process-wide logger. It defaults to a console logger so logging
// works before Init is called.
var global = logger.New(
	logger.WithSink(logger.NewWriterSink(os.Stdout, logger.NewConsoleEncoder(), logger.LevelInfo)),
	logger.WithLevel(logger.LevelInfo),
)

// Init builds the global logger from cfg, fanning out to the enabled sinks.
// Call it once during bootstrap, after the OTEL providers are registered so the
// OTEL sink picks up the configured LoggerProvider.
func Init(cfg Config) {
	var sinks []logger.Sink
	if cfg.EnableStdout {
		sinks = append(sinks, logger.NewWriterSink(os.Stdout, logger.NewJSONEncoder(), logger.LevelInfo))
	}
	if cfg.EnableOtel {
		sinks = append(sinks, otelsink.New(cfg.ServiceName, logger.LevelInfo))
	}
	if cfg.EnableFile && cfg.LogFile != "" {
		if f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			sinks = append(sinks, logger.NewWriterSink(f, logger.NewJSONEncoder(), logger.LevelInfo))
		}
	}

	var sink logger.Sink
	switch len(sinks) {
	case 0:
		sink = logger.NewWriterSink(os.Stdout, logger.NewJSONEncoder(), logger.LevelInfo)
	case 1:
		sink = sinks[0]
	default:
		sink = logger.NewFanout(sinks...)
	}

	global = logger.New(logger.WithSink(sink), logger.WithLevel(logger.LevelInfo))
}

// GetLogger returns the global logger.
func GetLogger() *logger.Logger { return global }

// FromCtx returns the global logger enriched with the active trace's trace_id
// and span_id when ctx carries a valid span, so logs correlate with traces.
// Mirrors the old gozap.FromCtx ergonomic.
func FromCtx(ctx context.Context) *logger.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return global
	}
	return global.With(
		logger.String("trace_id", sc.TraceID().String()),
		logger.String("span_id", sc.SpanID().String()),
	)
}
