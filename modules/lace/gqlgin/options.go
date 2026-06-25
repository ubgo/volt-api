package gqlgin

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
)

// Default route paths
const (
	DefaultQueryRoute           = "/query"
	DefaultQueryPlaygroundRoute = "/query_playground"
	DefaultPlaygroundRoute      = "/gql"
	DefaultPlaygroundKey        = "pkey"
)

// Config holds the configuration for the Gql instance.
type Config struct {
	routerGroup      *gin.RouterGroup
	gqlServer        *handler.Server
	middleware       []gin.HandlerFunc
	routeGroupPrefix string
	playgroundKey    string

	// Route paths (configurable)
	queryRoute           string
	queryPlaygroundRoute string
	playgroundRoute      string

	// DisablePlayground controls whether the GraphQL playground routes are registered.
	// When set to true, the playground UI and introspection endpoint routes
	// will not be registered. This is useful for external API distributions (e.g., partner APIs)
	// where you want to hide the playground from end users who only have API keys and shouldn't
	// know about the playground key mechanism. Defaults to false (playground enabled).
	disablePlayground bool

	// EnableDefaultIntrospection controls whether introspection is enabled by default on the query route.
	// When set to true, introspection will be automatically enabled on the query endpoint, allowing
	// external API consumers (e.g., partners with API keys) to introspect the schema without needing
	// to use the query_playground endpoint or know about the playground key. The introspection will
	// be protected by whatever middleware is configured in the Middleware chain (e.g., API key middleware).
	// This is particularly useful for external API distributions where you want introspection available
	// but protected by standard authentication mechanisms rather than the playground key.
	// Defaults to false (introspection disabled on query route, only available via query_playground with pkey).
	enableDefaultIntrospection bool
}

// Option is a function that configures the Gql instance.
type Option func(*Config)

// WithRouterGroup sets the Gin router group.
func WithRouterGroup(rg *gin.RouterGroup) Option {
	return func(c *Config) {
		c.routerGroup = rg
	}
}

// WithGqlServer sets the GraphQL server handler.
func WithGqlServer(server *handler.Server) Option {
	return func(c *Config) {
		c.gqlServer = server
	}
}

// WithMiddleware sets the middleware chain.
func WithMiddleware(middleware ...gin.HandlerFunc) Option {
	return func(c *Config) {
		c.middleware = middleware
	}
}

// WithRouteGroupPrefix sets the route group prefix for playground handler.
func WithRouteGroupPrefix(prefix string) Option {
	return func(c *Config) {
		c.routeGroupPrefix = prefix
	}
}

// WithPlaygroundKey sets the playground access key.
func WithPlaygroundKey(key string) Option {
	return func(c *Config) {
		c.playgroundKey = key
	}
}

// WithQueryRoute sets the query endpoint path. Default: "/query"
func WithQueryRoute(route string) Option {
	return func(c *Config) {
		c.queryRoute = route
	}
}

// WithQueryPlaygroundRoute sets the query playground endpoint path. Default: "/query_playground"
func WithQueryPlaygroundRoute(route string) Option {
	return func(c *Config) {
		c.queryPlaygroundRoute = route
	}
}

// WithPlaygroundRoute sets the playground UI endpoint path. Default: "/gql"
func WithPlaygroundRoute(route string) Option {
	return func(c *Config) {
		c.playgroundRoute = route
	}
}

// WithDisablePlayground disables the playground routes.
func WithDisablePlayground(disable bool) Option {
	return func(c *Config) {
		c.disablePlayground = disable
	}
}

// WithEnableDefaultIntrospection enables introspection on the query route.
func WithEnableDefaultIntrospection(enable bool) Option {
	return func(c *Config) {
		c.enableDefaultIntrospection = enable
	}
}
