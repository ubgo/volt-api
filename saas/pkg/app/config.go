package app

type apppkgconfig struct {
	// for cron jobs we dont need to connect to nats
	DisableNats bool

	AfterNatsInitiated func(plugins *Plugin)
}

func (c *apppkgconfig) options(opts ...AppPkgOption) {
	for _, opt := range opts {
		opt(c)
	}
}

type AppPkgOption func(*apppkgconfig)

func WithDisableNats(disableNats bool) AppPkgOption {
	return func(c *apppkgconfig) {
		c.DisableNats = disableNats
	}
}

func WithAfterNatsInitiated(callback func(plugins *Plugin)) AppPkgOption {
	return func(c *apppkgconfig) {
		c.AfterNatsInitiated = callback
	}
}
