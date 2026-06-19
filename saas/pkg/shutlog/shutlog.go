// Package shutlog adapts a *zap.Logger to the github.com/ubgo/shutdown
// Logger interface, so graceful-shutdown logs flow through the same zap
// pipeline (and gozap/OTEL wiring) as the rest of the application.
package shutlog

import (
	"github.com/ubgo/shutdown"
	"go.uber.org/zap"
)

// Logger implements shutdown.Logger on top of a zap SugaredLogger. The
// shutdown package logs with alternating key/value pairs, which maps
// directly onto SugaredLogger's *w methods.
type Logger struct {
	sugar *zap.SugaredLogger
}

// New wraps l as a shutdown.Logger. A nil logger falls back to a no-op so
// callers never have to guard against it.
func New(l *zap.Logger) shutdown.Logger {
	if l == nil {
		l = zap.NewNop()
	}
	return &Logger{sugar: l.Sugar()}
}

func (z *Logger) Info(msg string, fields ...any)  { z.sugar.Infow(msg, fields...) }
func (z *Logger) Warn(msg string, fields ...any)  { z.sugar.Warnw(msg, fields...) }
func (z *Logger) Error(msg string, fields ...any) { z.sugar.Errorw(msg, fields...) }
