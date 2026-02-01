package app

import "core/pkg/plugin"

type config struct {
	// for cron jobs we dont need to connect to nats
	coreOptions []plugin.AppPkgOption

	// AfterNatsInitiated func(plugins *plugin.Plugin)
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

func WithCoreOptions(options ...plugin.AppPkgOption) Option {
	return func(c *config) {
		c.coreOptions = append(c.coreOptions, options...)
	}
}
