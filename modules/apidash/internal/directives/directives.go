package directives

import (
	"apidash/internal/apidashfn"
	"context"

	"github.com/99designs/gqlgen/graphql"
)

func InternalApiDirective(internalKey string) func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		err := apidashfn.ValidateInternalKeyFromRequest(ctx, internalKey)

		if err != nil {
			return nil, err
		}

		return next(ctx)
	}
}
