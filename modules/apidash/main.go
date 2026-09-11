package apidash

import (
	"apidash/internal/directives"
	"apidash/internal/graph/generated"
	"apidash/internal/graph/resolver"
	"apidash/internal/handlers"
	"apidash/pkg/apidashtypes"
	"config"
	"lace/gqlgenfn"
	"lace/gqlgin"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
)

func Boot(routerGroup *gin.RouterGroup, gqlResolver *apidashtypes.Resolver) {
	prefix := "/dash"
	rg := routerGroup.Group(prefix)

	c := generated.Config{Resolvers: (*resolver.Resolver)(gqlResolver)}
	setDirectives(&c, gqlResolver.RootConfig)

	schm := generated.NewExecutableSchema(c)
	gserver := handler.New(schm)

	// @tags('dnd')
	// Add the otelgqlgen middleware to add OpenTelemetry tracing this will track graphql queries and mutations
	// if resolver.RootConfig.Otel.Enable {
	// 	gserver.Use(otelgqlgen.Middleware())
	// }

	gserver.AddTransport(transport.Options{})
	gserver.AddTransport(transport.GET{})
	gserver.AddTransport(transport.POST{})

	gqlgin.New(
		gqlgin.WithRouterGroup(rg),
		gqlgin.WithGqlServer(gserver),
		gqlgin.WithPlaygroundKey(gqlResolver.RootConfig.Graphql.Key),
		gqlgin.WithRouteGroupPrefix(rg.BasePath()),
		gqlgin.WithMiddleware(
			gqlgenfn.GinContextToContextMiddleware(),
		),
	)

	handlers.New(handlers.Config{
		RouterGroup: rg,
		EntClient:   gqlResolver.Circuit.EntDB.Client(),
		RootConfig:  gqlResolver.Circuit.RootConfig,
	})
}

func setDirectives(c *generated.Config, appcofnig *config.Root) {
	c.Directives.Internal = directives.InternalApiDirective(appcofnig.Core.InternalKey)
}
