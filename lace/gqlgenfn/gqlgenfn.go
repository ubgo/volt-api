package gqlgenfn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	PlaygroundKey                = "pkey"
	isPlaygroundAllwedContextKey = "isplayground" //
)

func (c contextKey) String() string {
	return string(c)
}

// https://github.com/99designs/gqlgen/blob/master/docs/content/recipes/gin.md
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), contextKey("GinContextKey"), c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// check if ctx is already ginCtx if not then may be it's gqlCtx and extract ginCtx from gqlCtx
func GinContextFromContext(ctx context.Context) (*gin.Context, error) {
	ginCtx, isGin := ctx.(*gin.Context)

	var err error

	if isGin {
		return ginCtx, err
	}

	ginContext := ctx.Value(contextKey("GinContextKey"))
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return nil, err
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return nil, err
	}
	return gc, nil
}

// protect this route http://localhost:2303/sa/gql?pkey=1234
func PlaygroundAccessMiddleware(configkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok := setPlaygroundContext(configkey, c)

		if !ok {
			err := fmt.Errorf("playground access denied")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": err.Error(),
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func setPlaygroundContext(configkey string, gc *gin.Context) bool {
	key := gc.Query(PlaygroundKey)

	referer := gc.Request.Header.Get("Referer")
	if len(referer) != 0 && len(key) == 0 {
		u, err := url.Parse(referer)
		if err != nil {
			panic(err)
		}

		q := u.Query()
		key = q.Get(PlaygroundKey)
	}

	if len(key) == 0 {
		key = gc.Query(PlaygroundKey)
	}

	if key == configkey {
		gc.Set(isPlaygroundAllwedContextKey, true)
		return true
	}

	gc.Set(isPlaygroundAllwedContextKey, false)
	return false
}

func IsPlaygroundGqlCtx(ctx context.Context) bool {
	gc, err := GinContextFromContext(ctx)
	if err != nil {
		return false
	}

	return IsPlaygroundGinCtx(gc)
}

func IsPlaygroundGinCtx(c *gin.Context) bool {
	record, _ := c.Get(isPlaygroundAllwedContextKey)

	if record == nil {
		return false
	}

	value, _ := record.(bool)
	return value

}

type SentryMiddleware struct{}

func (a SentryMiddleware) ExtensionName() string {
	return "SentryMiddleware"
}

func (a SentryMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (a SentryMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
			log.Println(string(debug.Stack()))
			hub := sentry.CurrentHub().Clone()
			hub.Recover(r)
			hub.Flush(time.Second * 2)
			err = fmt.Errorf("internal server error")
		}
	}()
	return next(ctx)
}

// extract any key from the request like x-key, token etc
func KeyFromRequest(ctx context.Context, keyname string) (string, error) {
	gc, err := GinContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	apikey := gc.Query(keyname)

	xkey := gc.Request.Header.Get(keyname)
	if len(xkey) > 0 {
		apikey = xkey
	}

	apikey = strings.TrimSpace(apikey)

	if len(apikey) == 0 {
		return "", fmt.Errorf("no %s provided", keyname)
	}

	return apikey, nil
}
