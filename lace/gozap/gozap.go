package gozap

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var fromCtxInitCb func(ctx context.Context) []zap.Field

// allow to add custom fields let say extract header keys from Gin Context so ctx can be gin, graphql any
// it's upto the provider how he handles the ctx
// call this globally in bootstrap once
func FromCtxInitCb(fn func(ctx context.Context) []zap.Field) {
	fromCtxInitCb = fn
}

/*
traceCtx, span := telemetry.TraceStart(ctx, "testtrace")
span.SetAttributes(attribute.String("name", "otel"))
span.End()
gozap.FromCtx(traceCtx).Info("test gozap1")
*/
func FromCtx(ctx context.Context) *zap.Logger {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		// ctx does not contain a valid span.
		// There is no trace metadata to add.
		return GetLogger()
	}

	fmt.Println("trace_id", spanContext.TraceID().String())
	fmt.Println("span_id", spanContext.SpanID().String())

	fields := []zap.Field{
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	}

	if fromCtxInitCb != nil {
		fields = append(fields, fromCtxInitCb(ctx)...)
	}

	return GetLogger().With(fields...)
}

func GetLogger() *zap.Logger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return logger
}

type Config struct {
	EnableStdout   bool
	EnableOtel     bool
	EnableFile     bool
	ServiceName    string
	LoggerProvider *log.LoggerProvider
	LocalConfig    LocalConfig
	Core           []zapcore.Core
}

func NewLogger(cfg Config) *zap.Logger {
	cores := []zapcore.Core{}
	if cfg.EnableStdout {
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), zapcore.InfoLevel))
	}

	if cfg.EnableOtel && cfg.LoggerProvider != nil {
		// fmt.Println("enable otel")
		cores = append(cores, WithOtel(cfg.ServiceName, cfg.LoggerProvider))
	}

	if cfg.EnableFile {
		cores = append(cores, WithLocal(cfg.LocalConfig))
	}

	if len(cfg.Core) > 0 {
		cores = append(cores, cfg.Core...)
	}

	// AddCaller() will add the caller information to the log like file name, line number, function name
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller())
}

func InitGlobalLogger(cfg Config) {
	logger = NewLogger(cfg)
}
