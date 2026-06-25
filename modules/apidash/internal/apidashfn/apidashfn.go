package apidashfn

import (
	"context"
	"fmt"
	"lace/gqlgenfn"
)

func ValidateInternalKeyFromRequest(ctx context.Context, internalKey string) error {

	keyName := "x-internal-key"

	apikey, err := gqlgenfn.KeyFromRequest(ctx, keyName)

	if err != nil {
		return err
	}

	if apikey != internalKey {
		return fmt.Errorf("invalid key value for %s", keyName)
	}

	return nil
}
