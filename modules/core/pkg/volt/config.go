package volt

type apppkgconfig struct{}

func (c *apppkgconfig) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*apppkgconfig)
