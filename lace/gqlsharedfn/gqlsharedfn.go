package gqlsharedfn

import (
	"context"
	"fmt"
	"lace/gqlgenfn"
	"strings"
)

// shared function which can be used in both apidash and gql

func TokenFromRequest(ctx context.Context) (string, error) {
	gc, err := gqlgenfn.GinContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	result := gc.Query("token")

	xat := gc.Request.Header.Get("token")
	if len(xat) > 0 {
		result = xat
	}

	result = strings.TrimSpace(result)

	if len(result) == 0 {
		return "", fmt.Errorf("no token provided")
	}

	return result, nil
}
