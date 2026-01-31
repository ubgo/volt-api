package medialibrary

type UrlGenerator interface {
	IsPublicllyAccessible() bool
	GetAbsolutePath() string
	GetURL() string
	GetTemporaryURL() string
}

type PublicUrlGenerator interface {
	PublicURL(path string) string
}
