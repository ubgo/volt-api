package filesystem

type configFileSystem struct {
	PublicURLGenerator func(path string) (string, error)
	PublicURL          []string
}

func (c *configFileSystem) options(opts ...FileSystemOption) {
	for _, opt := range opts {
		opt(c)
	}
}

type FileSystemOption func(*configFileSystem)

func WithPublicURLGenerator(fn func(path string) (string, error)) FileSystemOption {
	return func(c *configFileSystem) {
		c.PublicURLGenerator = fn
	}
}

func WithPublicURL(urls []string) FileSystemOption {
	return func(c *configFileSystem) {
		c.PublicURL = urls
	}
}
