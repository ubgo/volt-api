package apphandlers

import (
	"config"
	"dbent/gen/ent"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	HttpEngine     *gin.Engine
	ApiRouterGroup *gin.RouterGroup
	EntClient      *ent.Client
	RootConfig     *config.Root
}

type AppHandlers struct {
	Config
}

func New(config Config) {
	handler := &AppHandlers{Config: config}

	apiRouterGroup := config.ApiRouterGroup
	apiRouterGroup.GET("/app_test", handler.appTestHandler)

}

// versionHandler returns build information
func (pkg AppHandlers) appTestHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "app test",
	})
}
