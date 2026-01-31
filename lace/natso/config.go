package natso

type config struct {
	Enabled   bool
	AppKey    string // namespacing the nats app unique identifier
	URL       string
	CredsFile *string
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

func WithAppKey(appKey string) Option {
	return func(c *config) {
		c.AppKey = appKey
	}
}

func WithURL(natsurl string) Option {
	return func(c *config) {
		c.URL = natsurl
	}
}

func WithEnabled(enabled bool) Option {
	return func(c *config) {
		c.Enabled = enabled
	}
}

func WithCredsFile(credsFile *string) Option {
	return func(c *config) {
		c.CredsFile = credsFile
	}
}
