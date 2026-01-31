package ginserver

import "github.com/gin-gonic/gin"

type config struct {
	BeforeHandlers []gin.HandlerFunc
	Port           string
	IsProd         bool
	AppName        string
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

func WithBeforeHandler(handlers ...gin.HandlerFunc) Option {
	return func(c *config) {
		c.BeforeHandlers = handlers
	}
}

func WithPort(port string) Option {
	return func(c *config) {
		c.Port = port
	}
}

func WithIsProd(isProd bool) Option {
	return func(c *config) {
		c.IsProd = isProd
	}
}

func WithAppName(appName string) Option {
	return func(c *config) {
		c.AppName = appName
	}
}
