package gocronlogger

import (
	ulog "github.com/ubgo/logger"
)

// Logger is a gocron.Logger adapter backed by github.com/ubgo/logger.
type Logger struct {
	lg *ulog.Logger
}

// New creates a gocron.Logger backed by ubgo/logger.
func New(lg *ulog.Logger) *Logger {
	return &Logger{lg: lg}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.lg.Debug(msg, argsToFields(args)...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.lg.Error(msg, argsToFields(args)...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.lg.Info(msg, argsToFields(args)...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.lg.Warn(msg, argsToFields(args)...)
}

// argsToFields converts gocron's key-value pairs to logger fields.
func argsToFields(args []any) []ulog.Field {
	fields := make([]ulog.Field, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			fields = append(fields, ulog.Any(key, args[i+1]))
		}
	}
	return fields
}
