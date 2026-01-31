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

type Gql struct {
	Config
}

const PlaygroundKey = "pkey"

type Config struct {
	RouterGroup      *gin.RouterGroup
	GqlServer        *handler.Server
	Middleware       []gin.HandlerFunc
	RouteGroupPrefix string
	PlaygroundKey    string

	// DisablePlayground controls whether the GraphQL playground routes are registered.
	// When set to true, the /gql (playground UI) and /query_playground (introspection endpoint) routes
	// will not be registered. This is useful for external API distributions (e.g., partner APIs)
	// where you want to hide the playground from end users who only have API keys and shouldn't
	// know about the playground key mechanism. Defaults to false (playground enabled).
	DisablePlayground bool

	// EnableDefaultIntrospection controls whether introspection is enabled by default on the /query route.
	// When set to true, introspection will be automatically enabled on the /query endpoint, allowing
	// external API consumers (e.g., partners with API keys) to introspect the schema without needing
	// to use the /query_playground endpoint or know about the playground key. The introspection will
	// be protected by whatever middleware is configured in the Middleware chain (e.g., API key middleware).
	// This is particularly useful for external API distributions where you want introspection available
	// but protected by standard authentication mechanisms rather than the playground key.
	// Defaults to false (introspection disabled on /query, only available via /query_playground with pkey).
	EnableDefaultIntrospection bool
}

// Defining the Graphql handler
// instrospectionMiddleware enables introspection when the playground key is valid.
// This is used for the /query_playground route which requires the pkey query parameter.
func (cls *Gql) instrospectionMiddleware(gserver *handler.Server) gin.HandlerFunc {
	// https://github.com/99designs/gqlgen/blob/master/docs/content/reference/introspection.md
	gserver.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// isAllowedPlayground := gqlgenfn.IsPlaygroundAllowedOperation(cls.PlaygroundKey, ctx)
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

// defaultIntrospectionMiddleware enables introspection by default on the /query route.
// This is used when EnableDefaultIntrospection is true, allowing introspection to be
// protected by standard middleware (e.g., API key middleware) rather than the playground key.
func (cls *Gql) defaultIntrospectionMiddleware(gserver *handler.Server) gin.HandlerFunc {
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

func (cls *Gql) serverHttpHandler(gserver *handler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		gserver.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func playgroundHandler(routeGroupPrefix string) gin.HandlerFunc {
	h := playground.Handler("GraphQL", fmt.Sprintf("%s/query_playground", routeGroupPrefix))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func New(config Config) Gql {
	gqlgin := Gql{
		Config: config,
	}

	routeGroupPrefix := ""
	if len(config.RouteGroupPrefix) > 0 {
		routeGroupPrefix = config.RouteGroupPrefix
	}

	rg := config.RouterGroup

	// disable the introspection by default
	// This will be overridden by introspection middlewares if they are enabled
	config.GqlServer.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		graphql.GetOperationContext(ctx).DisableIntrospection = true
		return next(ctx)
	})

	// middleware orders is very important so do not change it
	// for making actual request and can have userAuth and other Middleware in between
	config.Middleware = append(config.Middleware, gqlgin.serverHttpHandler(config.GqlServer))

	// Build handlers for the /query route
	queryHandlers := []gin.HandlerFunc{
		gqlgenfn.GinContextToContextMiddleware(),
	}

	// If EnableDefaultIntrospection is true, add the introspection middleware
	// This allows introspection on /query, protected by the configured middleware chain
	if config.EnableDefaultIntrospection {
		queryHandlers = append(queryHandlers, gqlgin.defaultIntrospectionMiddleware(config.GqlServer))
	}

	// Append user-defined middleware (e.g., API key auth, user auth, etc.)
	queryHandlers = append(queryHandlers, config.Middleware...)

	rg.POST("/query", queryHandlers...)

	// Register playground routes only if DisablePlayground is false
	if !config.DisablePlayground {
		// only used for introspection, playground and protected by the ?pkey=xxx
		// also we do not use the /query for playground otherwise it unecessary runs the  PlaygroundAccessMiddleware
		// so we want to hyper optimize the speed and for /query we should not run this below middlewares to speed up
		introspectionHandlers := []gin.HandlerFunc{
			gqlgenfn.GinContextToContextMiddleware(),
			gqlgenfn.PlaygroundAccessMiddleware(config.PlaygroundKey),
			gqlgin.instrospectionMiddleware(config.GqlServer),
		}
		introspectionHandlers = append(introspectionHandlers, config.Middleware...)

		rg.POST("/query_playground", introspectionHandlers...)

		rg.GET("/gql", gqlgenfn.PlaygroundAccessMiddleware(config.PlaygroundKey), playgroundHandler(routeGroupPrefix))
	}

	return gqlgin
}
