package apidash

import (
	"apidash/internal/directives"
	"apidash/internal/graph/generated"
	"apidash/internal/graph/resolver"
	"apidash/internal/handlers"
	"apidash/pkg/apidashtypes"
	"config/gen/rootconfig"
	"fmt"
	"lace/gqlgenfn"
	"lace/gqlgin"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

func Boot(routerGroup *gin.RouterGroup, gqlResolver *apidashtypes.Resolver) {
	prefix := "/sa"
	rg := routerGroup.Group(prefix)

	c := generated.Config{Resolvers: (*resolver.Resolver)(gqlResolver)}
	setDirectives(&c, gqlResolver.RootConfig)

	schm := generated.NewExecutableSchema(c)
	// extendSchema(&schm)
	gserver := handler.New(schm)

	// Add the otelgqlgen middleware to add OpenTelemetry tracing this will track graphql queries and mutations
	// if resolver.RootConfig.Otel.Enable {
	// 	gserver.Use(otelgqlgen.Middleware())
	// }

	gserver.AddTransport(transport.Options{})
	gserver.AddTransport(transport.GET{})
	gserver.AddTransport(transport.POST{})

	gqlgin.New(gqlgin.Config{
		RouterGroup:      rg,
		GqlServer:        gserver,
		PlaygroundKey:    gqlResolver.RootConfig.Graphql.Key,
		RouteGroupPrefix: prefix,
		Middleware: []gin.HandlerFunc{
			gqlgenfn.GinContextToContextMiddleware(),
		},
	})

	handlers.New(handlers.Config{
		RouterGroup: rg,
		EntClient:   gqlResolver.Plugin.EntDB.Client(),
		RootConfig:  gqlResolver.Plugin.RootConfig,
	})

	fmt.Println("boot apidash")
}

func setDirectives(c *generated.Config, appcofnig *rootconfig.RootConfig) {
	c.Directives.Internal = directives.InternalApiDirective(appcofnig.Core.InternalKey)
}

// Custom schema extension
func extendSchema(schema *graphql.ExecutableSchema) {
	// fields, err := fetchDynamicFields()
	// if err != nil {
	// 	log.Fatalf("failed to fetch dynamic fields: %v", err)
	// }

	aa := *schema
	// fmt.Println(aa.Schema().Query.Fields)
	for _, field := range aa.Schema().Types {
		if field.Name == "Student" {

			fmt.Println(field.Name, reflect.TypeOf(field))

			field.Fields = append(field.Fields, &ast.FieldDefinition{
				Name: "class",
				Type: ast.NamedType("String", nil),
			})

		}

	}
	// // Add dynamic fields to the ExampleType
	// for _, field := range fields {
	// 	fieldName := field.FieldName
	// 	fieldType := field.FieldType

	// 	// Add field to introspection
	// 	schema.Query().Fields()["example"].Type.(*graphql.Object).AddFieldConfig(fieldName, &graphql.Field{
	// 		Type: graphql.String, // Adjust the type based on fieldType
	// 		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
	// 			example := p.Source.(*ExampleType)
	// 			if value, ok := example.DynamicFields[fieldName]; ok {
	// 				return value, nil
	// 			}
	// 			return nil, fmt.Errorf("field %s not found", fieldName)
	// 		},
	// 	})
	// }
}
