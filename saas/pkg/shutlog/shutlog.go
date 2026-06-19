// Package shutlog adapts a *logger.Logger to the github.com/ubgo/shutdown
// Logger interface, so graceful-shutdown logs flow through the same
// ubgo/logger pipeline (and OTEL wiring) as the rest of the application.
package shutlog

import (
	"io"

	ulog "github.com/ubgo/logger"
	"github.com/ubgo/shutdown"
)

// Logger implements shutdown.Logger on top of ubgo/logger. The shutdown
// package logs with alternating key/value pairs, which are converted to
// logger fields.
type Logger struct {
	lg *ulog.Logger
}

// New wraps lg as a shutdown.Logger. A nil logger falls back to a no-op so
// callers never have to guard against it.
func New(lg *ulog.Logger) shutdown.Logger {
	if lg == nil {
		lg = ulog.New(ulog.WithSink(ulog.NewWriterSink(io.Discard, ulog.NewJSONEncoder(), ulog.LevelError)))
	}
	return &Logger{lg: lg}
}

func (z *Logger) Info(msg string, fields ...any)  { z.lg.Info(msg, pairsToFields(fields)...) }
func (z *Logger) Warn(msg string, fields ...any)  { z.lg.Warn(msg, pairsToFields(fields)...) }
func (z *Logger) Error(msg string, fields ...any) { z.lg.Error(msg, pairsToFields(fields)...) }

// pairsToFields converts alternating key/value pairs to logger fields.
func pairsToFields(args []any) []ulog.Field {
	fields := make([]ulog.Field, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			fields = append(fields, ulog.Any(key, args[i+1]))
		}
	}
	return fields
}
