package goerr

import (
	"context"
	"encoding/json"
	"time"
)

var callbacks []callbackFn

type Error struct {
	resp *ErrorResult

	// otel trace ctx so we can extract trace_id and span_id
	// traceCtx, span := telemetry.TraceStart(ctx, "test-http1")
	// and pass CTX in callback so zap logger can use it
	// gozap.FromCtx(ctx).Info("chatbotSetFavorite", zap.String("id", id), zap.Bool("isFavorite", isFavorite))
	ctx context.Context
}

func (e *Error) Error() string {
	runCallbacks(e)
	return e.resp.Message
}

func (e *Error) ErrorUser() *ErrorUser {
	return &ErrorUser{resp: e.resp, err: e}
}

func (e *Error) Response() ErrorResult {
	return *e.resp
}

func (e *Error) SetMessage(message string) *Error {
	e.resp.Message = message
	return e
}

func (e *Error) SetMessageUser(message string) *Error {
	e.resp.MessageUser = message
	return e
}

func (e *Error) SetKV(key string, value any) *Error {
	e.resp.KV[key] = value
	return e
}

type ErrorUser struct {
	resp *ErrorResult
	err  *Error
}

func (e *ErrorUser) Error() string {
	runCallbacks(e.err)
	return e.resp.MessageUser
}

type ErrorResult struct {
	MessageUser string `json:"messageUser,omitempty"`
	Message     string `json:"message,omitempty"`
	Code        string `json:"code,omitempty"`
	Data        any    `json:"data,omitempty"`
	TraceID     string `json:"traceId,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Level       string `json:"level,omitempty"`

	// any extra information like line number, stack trace etc
	KV map[string]any `json:"kv,omitempty"`
}

// To JSON RAW
func (e *ErrorResult) ToJSON() string {
	json, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(json)
}

func New(message string, options ...ErrorOption) *Error {
	resp := &ErrorResult{Message: message, MessageUser: message, Timestamp: time.Now().Format(time.RFC3339), Level: "error"}
	e := &Error{resp: resp}

	for _, option := range options {
		option(e)
	}
	return e
}

// sometimes we want to create an instance withData and then use the same instance to log the error with different message
// so each error message will have the same data but different message
func NewDefault(options ...ErrorOption) *Error {
	resp := &ErrorResult{Message: "An error occurred", MessageUser: "An error occurred", Timestamp: time.Now().Format(time.RFC3339), Level: "error"}
	e := &Error{resp: resp}

	for _, option := range options {
		option(e)
	}
	return e
}

type ErrorOption func(*Error)

func WithMessageUser(message string) ErrorOption {
	return func(e *Error) {
		e.resp.MessageUser = message
	}
}

// func WithMessage(message string) ErrorOption {
// 	return func(e *Error) {
// 		e.resp.Message = message
// 	}
// }

func WithCode(code string) ErrorOption {
	return func(e *Error) {
		e.resp.Code = code
	}
}

func WithData(data any) ErrorOption {
	return func(e *Error) {
		e.resp.Data = data
	}
}

func WithTraceID(traceID string) ErrorOption {
	return func(e *Error) {
		e.resp.TraceID = traceID
	}
}

func WithLevel(level string) ErrorOption {
	return func(e *Error) {
		e.resp.Level = level
	}
}

func WithKVMap(kv map[string]any) ErrorOption {
	return func(e *Error) {
		for k, v := range kv {
			e.resp.KV[k] = v
		}
	}
}

func WithKV(key string, value any) ErrorOption {
	return func(e *Error) {
		e.resp.KV[key] = value
	}
}

func WithCtx(ctx context.Context) ErrorOption {
	return func(e *Error) {
		e.ctx = ctx
	}
}

// for any global callbacks let say send the erorr to loki or sentry
type callbackFn func(ctx context.Context, resp *ErrorResult)

func AddCallback(callback callbackFn) {
	callbacks = append(callbacks, callback)
}

func runCallbacks(e *Error) {
	for _, callback := range callbacks {
		callback(e.ctx, e.resp)
	}
}
