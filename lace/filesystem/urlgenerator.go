package filesystem

type PrefixPublicUrlGenerator struct {
	prefixer PathPrefixer
}

func NewPrefixPublicUrlGenerator(url string) *PrefixPublicUrlGenerator {
	prefixer := NewPathPrefixer(url, "/")
	return &PrefixPublicUrlGenerator{*prefixer}
}

func (g *PrefixPublicUrlGenerator) PublicURL(path string) (string, error) {
	return g.prefixer.PrefixPath(path), nil
}
