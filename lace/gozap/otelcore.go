package gozap

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap/zapcore"
)

func WithOtel(serviceName string, lp *log.LoggerProvider) zapcore.Core {
	return otelzap.NewCore(serviceName, otelzap.WithLoggerProvider(lp))
}
