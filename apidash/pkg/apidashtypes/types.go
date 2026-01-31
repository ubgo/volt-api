package apidashtypes

import (
	"config/gen/rootconfig"
	"saas/pkg/app"
)

type Resolver struct {
	Plugin     *app.Plugin
	RootConfig *rootconfig.RootConfig
}
