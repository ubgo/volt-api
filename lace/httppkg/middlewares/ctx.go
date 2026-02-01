package middlewares

import (
	"context"
	"fmt"
)

/*
type userID struct{}
type requestID struct{}

// Store values
ctx = WithValue(ctx, userID{}, "user123")
ctx = WithValue(ctx, requestID{}, "req456")

// Retrieve values
userID, err := GetValue[string](ctx, userID{})
requestID, err := GetValue[string](ctx, requestID{})
*/

// WithValue associates the value with the key in the given context.
func WithValue[K comparable, V any](ctx context.Context, key K, value V) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetValue returns the value associated with the key in the given context.
func GetValue[K comparable, V any](ctx context.Context, key K) (V, error) {
	value, ok := ctx.Value(key).(V)
	if !ok {
		var zero V
		return zero, fmt.Errorf("missing")
	}

	return value, nil
}
