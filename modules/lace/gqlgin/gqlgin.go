package gqlgin

import (
	"context"
	"fmt"

	"lace/gqlgenfn"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// Gql handles GraphQL endpoint registration and configuration.
type Gql struct {
	config *Config
}

// instrospectionMiddleware enables introspection when the playground key is valid.
// This is used for the query_playground route which requires the pkey query parameter.
func (g *Gql) instrospectionMiddleware(gserver *handler.Server) gin.HandlerFunc {
	// https://github.com/99designs/gqlgen/blob/master/docs/content/reference/introspection.md
	gserver.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		isAllowedPlayground := gqlgenfn.IsPlaygroundGqlCtx(ctx)
		if isAllowedPlayground {
			graphql.GetOperationContext(ctx).DisableIntrospection = false
		}

		return next(ctx)
	})

	return func(c *gin.Context) {
		c.Next()
	}
}

// defaultIntrospectionMiddleware enables introspection by default on the query route.
// This is used when EnableDefaultIntrospection is true, allowing introspection to be
// protected by standard middleware (e.g., API key middleware) rather than the playground key.
func (g *Gql) defaultIntrospectionMiddleware(gserver *handler.Server) gin.HandlerFunc {
	// https://github.com/99designs/gqlgen/blob/master/docs/content/reference/introspection.md
	gserver.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Enable introspection by default when this middleware is applied
		graphql.GetOperationContext(ctx).DisableIntrospection = false
		return next(ctx)
	})

	return func(c *gin.Context) {
		c.Next()
	}
}

func (g *Gql) serverHttpHandler(gserver *handler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		gserver.ServeHTTP(c.Writer, c.Request)
	}
}

// playgroundHandler returns a handler for the GraphQL playground UI.
func playgroundHandler(routeGroupPrefix, queryPlaygroundRoute string) gin.HandlerFunc {
	h := playground.Handler("GraphQL", fmt.Sprintf("%s%s", routeGroupPrefix, queryPlaygroundRoute))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// New creates a new Gql instance with the provided options.
func New(opts ...Option) *Gql {
	cfg := &Config{
		queryRoute:           DefaultQueryRoute,
		queryPlaygroundRoute: DefaultQueryPlaygroundRoute,
		playgroundRoute:      DefaultPlaygroundRoute,
		playgroundKey:        DefaultPlaygroundKey,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	g := &Gql{
		config: cfg,
	}

	routeGroupPrefix := cfg.routeGroupPrefix

	rg := cfg.routerGroup

	// disable the introspection by default
	// This will be overridden by introspection middlewares if they are enabled
	cfg.gqlServer.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		graphql.GetOperationContext(ctx).DisableIntrospection = true
		return next(ctx)
	})

	// middleware orders is very important so do not change it
	// for making actual request and can have userAuth and other Middleware in between
	middleware := append(cfg.middleware, g.serverHttpHandler(cfg.gqlServer))

	// Build handlers for the query route
	queryHandlers := []gin.HandlerFunc{
		gqlgenfn.GinContextToContextMiddleware(),
	}

	// If EnableDefaultIntrospection is true, add the introspection middleware
	// This allows introspection on query route, protected by the configured middleware chain
	if cfg.enableDefaultIntrospection {
		queryHandlers = append(queryHandlers, g.defaultIntrospectionMiddleware(cfg.gqlServer))
	}

	// Append user-defined middleware (e.g., API key auth, user auth, etc.)
	queryHandlers = append(queryHandlers, middleware...)

	rg.POST(cfg.queryRoute, queryHandlers...)

	// Register playground routes only if DisablePlayground is false
	if !cfg.disablePlayground {
		// only used for introspection, playground and protected by the ?pkey=xxx
		// also we do not use the query route for playground otherwise it unnecessarily runs the PlaygroundAccessMiddleware
		// so we want to hyper optimize the speed and for query route we should not run this below middlewares to speed up
		introspectionHandlers := []gin.HandlerFunc{
			gqlgenfn.GinContextToContextMiddleware(),
			gqlgenfn.PlaygroundAccessMiddleware(cfg.playgroundKey),
			g.instrospectionMiddleware(cfg.gqlServer),
		}
		introspectionHandlers = append(introspectionHandlers, middleware...)

		rg.POST(cfg.queryPlaygroundRoute, introspectionHandlers...)

		rg.GET(cfg.playgroundRoute, gqlgenfn.PlaygroundAccessMiddleware(cfg.playgroundKey), playgroundHandler(routeGroupPrefix, cfg.queryPlaygroundRoute))
	}

	return g
}
