package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func OtelMiddleware(enable bool, serviceName string, fns func(ctx *gin.Context) []oteltrace.SpanStartOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enable {
			c.Next()
			return
		}

		var spanStartOptions []oteltrace.SpanStartOption

		if fns != nil {
			spanStartOptions = append(spanStartOptions, fns(c)...)
		}

		// @doc: inject custom attributes to the span
		// workspaceID, _ := GetWorkspaceID(c.Request.Context())
		// if len(workspaceID) > 0 {
		// 	spanStartOptions = append(spanStartOptions, oteltrace.WithAttributes(attribute.KeyValue{
		// 		Key:   constants.LogWorkspaceIdKey,
		// 		Value: attribute.StringValue(workspaceID),
		// 	}))
		// }

		otelgin.Middleware(serviceName, otelgin.WithSpanStartOptions(spanStartOptions...))(c)
		c.Next()
	}
}
