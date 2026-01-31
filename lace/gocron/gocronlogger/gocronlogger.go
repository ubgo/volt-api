package gocronlogger

import (
	"go.uber.org/zap"
)

// ZapLogger is a gocron.Logger adapter for zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger
func New(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: logger}
}

func (l *ZapLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, argsToZapFields(args)...)
}

func (l *ZapLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, argsToZapFields(args)...)
}

func (l *ZapLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, argsToZapFields(args)...)
}

func (l *ZapLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, argsToZapFields(args)...)
}

// argsToZapFields converts key-value pairs to zap fields
func argsToZapFields(args []any) []zap.Field {
	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, args[i+1]))
			}
		}
	}
	return fields
}
