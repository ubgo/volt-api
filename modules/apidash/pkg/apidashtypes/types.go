package apidashtypes

import (
	"config/gen/rootconfig"
	"webapp"
)

type Resolver struct {
	Circuit    *webapp.Circuit
	RootConfig *rootconfig.RootConfig
}
