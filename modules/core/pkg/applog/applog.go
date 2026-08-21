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

// CtxLogger is a context-bound view of the global logger. Every level method
// emits through the *Context variants so the log Record carries the caller's
// ctx — that ctx is what the OTEL sink hands to the SDK, which extracts the
// span and populates the BINARY TraceId/SpanId columns on the OTLP record.
// Without it (the pre-fix FromCtx returned a plain logger whose .Error() path
// hardcodes context.Background()), trace ids reached HyperDX only as string
// attributes and the log row's "Trace" tab showed "No trace id found for
// event". Same class of bug as the old gozap otelzap ctx-field fix.
type CtxLogger struct {
	ctx context.Context
	lg  *logger.Logger
}

func (c *CtxLogger) Trace(msg string, f ...logger.Field) { c.lg.TraceContext(c.ctx, msg, f...) }
func (c *CtxLogger) Debug(msg string, f ...logger.Field) { c.lg.DebugContext(c.ctx, msg, f...) }
func (c *CtxLogger) Info(msg string, f ...logger.Field)  { c.lg.InfoContext(c.ctx, msg, f...) }
func (c *CtxLogger) Warn(msg string, f ...logger.Field)  { c.lg.WarnContext(c.ctx, msg, f...) }
func (c *CtxLogger) Error(msg string, f ...logger.Field) { c.lg.ErrorContext(c.ctx, msg, f...) }

// With returns a child CtxLogger that prepends fields to every record,
// preserving the bound ctx.
func (c *CtxLogger) With(f ...logger.Field) *CtxLogger {
	return &CtxLogger{ctx: c.ctx, lg: c.lg.With(f...)}
}

// FromCtx returns the global logger bound to ctx: emissions carry ctx for the
// OTEL sink (binary TraceId/SpanId on OTLP records) AND, when ctx has a valid
// span, string trace_id/span_id fields so console/file sinks stay
// trace-correlated too. Mirrors the old gozap.FromCtx ergonomic.
func FromCtx(ctx context.Context) *CtxLogger {
	lg := global
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		lg = lg.With(
			logger.String("trace_id", sc.TraceID().String()),
			logger.String("span_id", sc.SpanID().String()),
		)
	}
	return &CtxLogger{ctx: ctx, lg: lg}
}
