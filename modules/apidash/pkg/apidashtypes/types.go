package apidashtypes

import (
	"config"
	"webapp"
)

type Resolver struct {
	Circuit    *webapp.Circuit
	RootConfig *config.Root
}
