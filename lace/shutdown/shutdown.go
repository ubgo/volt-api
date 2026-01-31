package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type config struct {
	Timeout                 time.Duration
	GracefulShutdownTimeout time.Duration
	MsgShuttingDown         string
	MsgShutdownCompleted    string
	MsgShutdownFailed       string
	Logger                  *zap.Logger
	ShutdownHandlers        []ShutdownHandler
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

type ShutdownHandler func(ctx context.Context) error

func WithShutdownHandler(handlers ...ShutdownHandler) Option {
	return func(c *config) {
		c.ShutdownHandlers = append(c.ShutdownHandlers, handlers...)
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(c *config) {
		c.Logger = logger
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.Timeout = timeout
	}
}

func WithGracefulShutdownTimeout(gracefulShutdownTimeout time.Duration) Option {
	return func(c *config) {
		c.GracefulShutdownTimeout = gracefulShutdownTimeout
	}
}

func WithMsgShuttingDown(msgShuttingDown string) Option {
	return func(c *config) {
		c.MsgShuttingDown = msgShuttingDown
	}
}

func WithMsgShutdownCompleted(msgShutdownCompleted string) Option {
	return func(c *config) {
		c.MsgShutdownCompleted = msgShutdownCompleted
	}
}

func WithMsgShutdownFailed(msgShutdownFailed string) Option {
	return func(c *config) {
		c.MsgShutdownFailed = msgShutdownFailed
	}
}

type Shutdown struct {
	cfg *config
}

func New(opts ...Option) *Shutdown {
	cfg := &config{
		Timeout:                 5 * time.Second,
		GracefulShutdownTimeout: 30 * time.Second,
		MsgShuttingDown:         "Shutting down...",
		MsgShutdownCompleted:    "Shutdown completed",
		MsgShutdownFailed:       "Shutdown failed",
	}
	cfg.options(opts...)

	if cfg.Logger == nil {
		logger, _ := zap.NewProduction()
		cfg.Logger = logger
	}

	return &Shutdown{cfg: cfg}
}

func (s *Shutdown) AddShutdownHandler(handler func(ctx context.Context) error) {
	s.cfg.ShutdownHandlers = append(s.cfg.ShutdownHandlers, handler)
}

func (s *Shutdown) runHandlers(ctx context.Context) error {
	for _, handler := range s.cfg.ShutdownHandlers {
		if err := handler(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Shutdown) Listen() {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)

	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.cfg.Logger.Info(s.cfg.MsgShuttingDown)

	// Graceful shutdown begins
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer cancel()

	err := s.runHandlers(shutdownCtx)
	if err != nil {
		s.cfg.Logger.Error(s.cfg.MsgShutdownFailed, zap.Error(err))
		return
	}

	// Wait an additional 10 seconds to ensure all NATS messages and database calls are fully processed.
	// Currently, there is no reliable way to confirm that all NATS messages have been handled.
	time.Sleep(s.cfg.GracefulShutdownTimeout)

	s.cfg.Logger.Info(s.cfg.MsgShutdownCompleted)
}
