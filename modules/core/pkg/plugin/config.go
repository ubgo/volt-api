package plugin

type apppkgconfig struct{}

func (c *apppkgconfig) options(opts ...AppPkgOption) {
	for _, opt := range opts {
		opt(c)
	}
}

type AppPkgOption func(*apppkgconfig)
