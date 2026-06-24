package coreapi

import (
	"config/gen/coreconfig"
	"dbent/gen/ent"
	"lace/bootinfo"
	"lace/ginserver"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubgo/buildinfo"
)

type Config struct {
	HttpEngine     *gin.Engine
	ApiRouterGroup *gin.RouterGroup
	EntClient      *ent.Client
	CoreConfig     *coreconfig.Setting
}

type coreApi struct {
	Config
}

func New(config Config) {
	handler := &coreApi{Config: config}

	apiRouterGroup := config.ApiRouterGroup
	apiRouterGroup.GET("/version", handler.versionHandler)

	apiInternalGroup := apiRouterGroup.Group("/i")
	apiInternalGroup.Use(ginserver.MiddlewareInternalApiGuard(config.CoreConfig.InternalKey))

	// http://localhost:9001/api/i/bootinfo?x-internal-key=1234
	apiInternalGroup.GET("/bootinfo", handler.bootinfoHandler)
}

// versionHandler returns build information
func (pkg coreApi) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, buildinfo.Map())
}

// bootinfoHandler returns boot information
func (pkg coreApi) bootinfoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, bootinfo.GetBootInfo())
}
