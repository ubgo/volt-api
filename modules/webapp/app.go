package webapp

import (
	"sync"

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

// Lazy returns an accessor that builds the Circuit on first call and returns
// the same one afterwards.
//
// Why it exists: New loads the configuration, pings Redis and refuses a
// placeholder encryption key outside dev. That is right for a service and
// wrong for a CLI, where it ran unconditionally in main and made `ping` and
// `--help` depend on infrastructure — and made a `CONFIG_ENV=prod` invocation
// exit before cobra could pick a subcommand at all. Commands that need the
// Circuit call the accessor; commands that do not never trigger the build.
//
// Invariant: at most one Circuit per process, so a lazy CLI still matches the
// api binary's "configuration is read once" guarantee. Safe for concurrent use.
func Lazy(options ...Option) func() *Circuit {
	return sync.OnceValue(func() *Circuit { return New(options...) })
}
