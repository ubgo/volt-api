package coreapi

import (
	"config"
	"dbent/gen/ent"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubgo/buildinfo"
)

type Config struct {
	HttpEngine     *gin.Engine
	ApiRouterGroup *gin.RouterGroup
	EntClient      *ent.Client
	CoreConfig     *config.Core
}

type coreApi struct {
	Config
}

func New(config Config) {
	handler := &coreApi{Config: config}

	apiRouterGroup := config.ApiRouterGroup
	apiRouterGroup.GET("/version", handler.versionHandler)
}

// versionHandler returns build information
func (pkg coreApi) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, buildinfo.Map())
}
