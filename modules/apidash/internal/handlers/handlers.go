package handlers

import (
	"config"
	"dbent/gen/ent"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	RouterGroup *gin.RouterGroup
	EntClient   *ent.Client
	RootConfig  *config.Root
}

type RouteHandler struct {
	Config
}

func New(config Config) {
	router := config.RouterGroup

	handler := &RouteHandler{Config: config}

	router.GET("/health", handler.HealthHandler)
}

// HealthHandler handles health check endpoint
func (pkg RouteHandler) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
