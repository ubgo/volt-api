package gqlfn

import (
	"context"
	"lace/gqlgenfn"
)

// TokenFromRequest extracts the authentication token from the request context
func TokenFromRequest(ctx context.Context) (string, error) {
	return gqlgenfn.KeyFromRequest(ctx, "Authorization")
}
