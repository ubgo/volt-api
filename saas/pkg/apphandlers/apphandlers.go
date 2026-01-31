package apphandlers

import (
	"config/gen/rootconfig"
	"dbent/gen/ent"
	"lace/bootinfo"
	"lace/buildinfo"
	"lace/ginserver"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	HttpEngine     *gin.Engine
	ApiRouterGroup *gin.RouterGroup
	EntClient      *ent.Client
	RootConfig     *rootconfig.RootConfig
}

type AppHandlers struct {
	Config
}

func New(config Config) {
	handler := &AppHandlers{Config: config}

	apiRouterGroup := config.ApiRouterGroup
	apiRouterGroup.GET("/version", handler.versionHandler)

	apiInternalGroup := apiRouterGroup.Group("/i")
	apiInternalGroup.Use(ginserver.MiddlewareInternalApiGuard(config.RootConfig.Core.InternalKey))

	// http://localhost:9001/api/i/bootinfo?x-internal-key=1234
	apiInternalGroup.GET("/bootinfo", handler.bootinfoHandler)
}

// versionHandler returns build information
func (pkg AppHandlers) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, buildinfo.Info())
}

// bootinfoHandler returns boot information
func (pkg AppHandlers) bootinfoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, bootinfo.GetBootInfo())
}
