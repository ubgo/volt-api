package shutdown

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNew_DefaultConfig(t *testing.T) {
	s := New()

	if s.cfg.Timeout != 5*time.Second {
		t.Errorf("expected default Timeout to be 5s, got %v", s.cfg.Timeout)
	}
	if s.cfg.GracefulShutdownTimeout != 30*time.Second {
		t.Errorf("expected default GracefulShutdownTimeout to be 30s, got %v", s.cfg.GracefulShutdownTimeout)
	}
	if s.cfg.MsgShuttingDown != "Shutting down..." {
		t.Errorf("expected default MsgShuttingDown, got %q", s.cfg.MsgShuttingDown)
	}
	if s.cfg.MsgShutdownCompleted != "Shutdown completed" {
		t.Errorf("expected default MsgShutdownCompleted, got %q", s.cfg.MsgShutdownCompleted)
	}
	if s.cfg.MsgShutdownFailed != "Shutdown failed" {
		t.Errorf("expected default MsgShutdownFailed, got %q", s.cfg.MsgShutdownFailed)
	}
	if s.cfg.Logger == nil {
		t.Error("expected Logger to be initialized")
	}
}

func TestNew_WithOptions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	timeout := 10 * time.Second
	gracefulTimeout := 60 * time.Second

	s := New(
		WithLogger(logger),
		WithTimeout(timeout),
		WithGracefulShutdownTimeout(gracefulTimeout),
		WithMsgShuttingDown("Custom shutting down"),
		WithMsgShutdownCompleted("Custom completed"),
		WithMsgShutdownFailed("Custom failed"),
	)

	if s.cfg.Timeout != timeout {
		t.Errorf("expected Timeout %v, got %v", timeout, s.cfg.Timeout)
	}
	if s.cfg.GracefulShutdownTimeout != gracefulTimeout {
		t.Errorf("expected GracefulShutdownTimeout %v, got %v", gracefulTimeout, s.cfg.GracefulShutdownTimeout)
	}
	if s.cfg.MsgShuttingDown != "Custom shutting down" {
		t.Errorf("expected custom MsgShuttingDown, got %q", s.cfg.MsgShuttingDown)
	}
	if s.cfg.MsgShutdownCompleted != "Custom completed" {
		t.Errorf("expected custom MsgShutdownCompleted, got %q", s.cfg.MsgShutdownCompleted)
	}
	if s.cfg.MsgShutdownFailed != "Custom failed" {
		t.Errorf("expected custom MsgShutdownFailed, got %q", s.cfg.MsgShutdownFailed)
	}
	if s.cfg.Logger != logger {
		t.Error("expected custom Logger to be set")
	}
}

func TestWithShutdownHandler(t *testing.T) {
	handler1 := func(ctx context.Context) error { return nil }
	handler2 := func(ctx context.Context) error { return nil }

	s := New(
		WithShutdownHandler(handler1, handler2),
	)

	if len(s.cfg.ShutdownHandlers) != 2 {
		t.Errorf("expected 2 shutdown handlers, got %d", len(s.cfg.ShutdownHandlers))
	}
}

func TestAddShutdownHandler(t *testing.T) {
	s := New()

	handler := func(ctx context.Context) error { return nil }
	s.AddShutdownHandler(handler)

	if len(s.cfg.ShutdownHandlers) != 1 {
		t.Errorf("expected 1 shutdown handler, got %d", len(s.cfg.ShutdownHandlers))
	}
}

func TestRunHandlers_Success(t *testing.T) {
	called := false
	s := New(
		WithShutdownHandler(func(ctx context.Context) error {
			called = true
			return nil
		}),
	)

	ctx := context.Background()
	err := s.runHandlers(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRunHandlers_Error(t *testing.T) {
	expectedErr := errors.New("shutdown error")
	s := New(
		WithShutdownHandler(func(ctx context.Context) error {
			return expectedErr
		}),
	)

	ctx := context.Background()
	err := s.runHandlers(ctx)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRunHandlers_MultipleHandlers_StopsOnError(t *testing.T) {
	handler1Called := false
	handler2Called := false
	expectedErr := errors.New("first handler error")

	s := New(
		WithShutdownHandler(
			func(ctx context.Context) error {
				handler1Called = true
				return expectedErr
			},
			func(ctx context.Context) error {
				handler2Called = true
				return nil
			},
		),
	)

	ctx := context.Background()
	err := s.runHandlers(ctx)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if !handler1Called {
		t.Error("expected first handler to be called")
	}
	if handler2Called {
		t.Error("expected second handler NOT to be called after first handler error")
	}
}

func TestRunHandlers_MultipleHandlers_AllSuccess(t *testing.T) {
	callOrder := []int{}

	s := New()
	s.AddShutdownHandler(func(ctx context.Context) error {
		callOrder = append(callOrder, 1)
		return nil
	})
	s.AddShutdownHandler(func(ctx context.Context) error {
		callOrder = append(callOrder, 2)
		return nil
	})
	s.AddShutdownHandler(func(ctx context.Context) error {
		callOrder = append(callOrder, 3)
		return nil
	})

	ctx := context.Background()
	err := s.runHandlers(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(callOrder) != 3 {
		t.Errorf("expected 3 handlers called, got %d", len(callOrder))
	}
	// Verify handlers are called in order
	for i, v := range callOrder {
		if v != i+1 {
			t.Errorf("expected handler %d at position %d, got %d", i+1, i, v)
		}
	}
}

func TestRunHandlers_ContextCancellation(t *testing.T) {
	s := New(
		WithShutdownHandler(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := s.runHandlers(ctx)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestRunHandlers_NoHandlers(t *testing.T) {
	s := New()

	ctx := context.Background()
	err := s.runHandlers(ctx)

	if err != nil {
		t.Errorf("expected no error with no handlers, got %v", err)
	}
}

func TestConfig_Options(t *testing.T) {
	cfg := &config{}

	opts := []Option{
		WithTimeout(15 * time.Second),
		WithGracefulShutdownTimeout(45 * time.Second),
	}

	cfg.options(opts...)

	if cfg.Timeout != 15*time.Second {
		t.Errorf("expected Timeout 15s, got %v", cfg.Timeout)
	}
	if cfg.GracefulShutdownTimeout != 45*time.Second {
		t.Errorf("expected GracefulShutdownTimeout 45s, got %v", cfg.GracefulShutdownTimeout)
	}
}

func TestWithLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config{}

	opt := WithLogger(logger)
	opt(cfg)

	if cfg.Logger != logger {
		t.Error("expected logger to be set")
	}
}

func TestWithTimeout(t *testing.T) {
	cfg := &config{}
	timeout := 20 * time.Second

	opt := WithTimeout(timeout)
	opt(cfg)

	if cfg.Timeout != timeout {
		t.Errorf("expected Timeout %v, got %v", timeout, cfg.Timeout)
	}
}

func TestWithGracefulShutdownTimeout(t *testing.T) {
	cfg := &config{}
	timeout := 120 * time.Second

	opt := WithGracefulShutdownTimeout(timeout)
	opt(cfg)

	if cfg.GracefulShutdownTimeout != timeout {
		t.Errorf("expected GracefulShutdownTimeout %v, got %v", timeout, cfg.GracefulShutdownTimeout)
	}
}

func TestWithMsgShuttingDown(t *testing.T) {
	cfg := &config{}
	msg := "Server is going down"

	opt := WithMsgShuttingDown(msg)
	opt(cfg)

	if cfg.MsgShuttingDown != msg {
		t.Errorf("expected MsgShuttingDown %q, got %q", msg, cfg.MsgShuttingDown)
	}
}

func TestWithMsgShutdownCompleted(t *testing.T) {
	cfg := &config{}
	msg := "Server shutdown complete"

	opt := WithMsgShutdownCompleted(msg)
	opt(cfg)

	if cfg.MsgShutdownCompleted != msg {
		t.Errorf("expected MsgShutdownCompleted %q, got %q", msg, cfg.MsgShutdownCompleted)
	}
}

func TestWithMsgShutdownFailed(t *testing.T) {
	cfg := &config{}
	msg := "Server shutdown error"

	opt := WithMsgShutdownFailed(msg)
	opt(cfg)

	if cfg.MsgShutdownFailed != msg {
		t.Errorf("expected MsgShutdownFailed %q, got %q", msg, cfg.MsgShutdownFailed)
	}
}

func TestAddShutdownHandler_Multiple(t *testing.T) {
	s := New()

	s.AddShutdownHandler(func(ctx context.Context) error { return nil })
	s.AddShutdownHandler(func(ctx context.Context) error { return nil })
	s.AddShutdownHandler(func(ctx context.Context) error { return nil })

	if len(s.cfg.ShutdownHandlers) != 3 {
		t.Errorf("expected 3 shutdown handlers, got %d", len(s.cfg.ShutdownHandlers))
	}
}

func TestWithShutdownHandler_CombinedWithAddHandler(t *testing.T) {
	s := New(
		WithShutdownHandler(func(ctx context.Context) error { return nil }),
	)

	s.AddShutdownHandler(func(ctx context.Context) error { return nil })

	if len(s.cfg.ShutdownHandlers) != 2 {
		t.Errorf("expected 2 shutdown handlers, got %d", len(s.cfg.ShutdownHandlers))
	}
}

func TestRunHandlers_ContextTimeout(t *testing.T) {
	s := New(
		WithShutdownHandler(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
				return nil
			}
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := s.runHandlers(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded error, got %v", err)
	}
}
