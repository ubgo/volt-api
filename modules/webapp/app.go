package webapp

import (
	"core/pkg/volt"
	_ "dbent/gen/ent/runtime"
)

type Circuit struct {
	volt.Kernel
}

var circuit *Circuit

func New(options ...Option) *Circuit {
	cfg := config{
		coreOptions: []volt.Option{},
	}

	cfg.options(options...)

	corePlugin := volt.New(cfg.coreOptions...)

	customApp := &Circuit{
		Kernel: *corePlugin,
	}

	// Add your custom app deps here

	return customApp
}
