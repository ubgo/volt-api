package webapp

import "core/pkg/volt"

type config struct {
	coreOptions []volt.Option
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

func WithCoreOptions(options ...volt.Option) Option {
	return func(c *config) {
		c.coreOptions = append(c.coreOptions, options...)
	}
}
