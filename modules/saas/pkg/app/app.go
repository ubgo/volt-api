package app

import (
	"core/pkg/plugin"
	_ "dbent/gen/ent/runtime"
)

type CustomApp struct {
	plugin.Plugin
}

var plugins *CustomApp

func New(options ...Option) *CustomApp {
	cfg := config{
		coreOptions: []plugin.AppPkgOption{},
	}

	cfg.options(options...)

	corePlugin := plugin.New(cfg.coreOptions...)

	customApp := &CustomApp{
		Plugin: *corePlugin,
	}

	// Add your custom app deps here

	return customApp
}
