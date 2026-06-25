package ginserver

import (
	"context"
	"errors"
	"lace/gqlgenfn"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config
	Router       *gin.Engine
	RouterGroups map[string]*gin.RouterGroup
	srv          *http.Server
}

func (server Server) AddRouterGroup(name string, path string) *gin.RouterGroup {
	server.RouterGroups[name] = server.Router.Group(path)
	return server.RouterGroups[name]
}

func (server Server) GetRouterGroup(name string) (*gin.RouterGroup, error) {
	group := server.RouterGroups[name]
	if group == nil {
		return nil, errors.New("no route group found")
	}
	return server.RouterGroups[name], nil
}

func (server *Server) Start() {
	port := server.config.Port
	url := "http://localhost:" + port
	log.Println("Http Sever started at " + color.CyanString(url))
	// server.Router.Run(":" + port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: server.Router.Handler(),
	}

	server.srv = srv

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

// https://gin-gonic.com/en/docs/examples/graceful-restart-or-stop/
func (server *Server) Shutdown(ctx context.Context) error {
	if err := server.srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
		return err
	}
	return nil
}

func (server Server) pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"app":     server.config.AppName,
	})
}

func New(opts ...Option) Server {
	cfg := config{
		AppName: "app",
		Port:    "2131",
		IsProd:  true,
	}

	cfg.options(opts...)

	if cfg.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	if cfg.BeforeHandlers != nil {
		for _, handler := range cfg.BeforeHandlers {
			router.Use(handler)
		}
	}

	server := Server{
		config:       cfg,
		Router:       router,
		RouterGroups: make(map[string]*gin.RouterGroup),
	}

	router.GET("/api/ping", server.pingHandler)

	return server
}

func MiddlewareInternalApiGuard(internalKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(internalKey) == 0 {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized key"})
			return
		}

		internalKeyFromRequest, err := gqlgenfn.KeyFromRequest(c, "x-internal-key")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized, no internal key provided"})
			return
		}

		if len(internalKeyFromRequest) == 0 {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized, no internal key provided"})
			return
		}

		if internalKeyFromRequest != internalKey {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized, invalid internal key"})
			return
		}
		c.Next()
	}
}
