package filesystem

import (
	"strings"
)

type PathPrefixer struct {
	prefix    string
	separator string
}

func NewPathPrefixer(prefix, separator string) *PathPrefixer {
	p := &PathPrefixer{separator: separator}
	p.prefix = strings.TrimRight(prefix, "\\/")

	if p.prefix != "" || prefix == separator {
		p.prefix += separator
	}
	return p
}

func (p *PathPrefixer) PrefixPath(path string) string {
	return p.prefix + strings.TrimLeft(path, "\\/")
}

func (p *PathPrefixer) StripPrefix(path string) string {
	if strings.HasPrefix(path, p.prefix) {
		return path[len(p.prefix):]
	}
	return path
}

func (p *PathPrefixer) StripDirectoryPrefix(path string) string {
	return strings.TrimRight(p.StripPrefix(path), "\\/")
}

func (p *PathPrefixer) PrefixDirectoryPath(path string) string {
	prefixedPath := p.PrefixPath(strings.TrimRight(path, "\\/"))

	if prefixedPath == "" || strings.HasSuffix(prefixedPath, p.separator) {
		return prefixedPath
	}

	return prefixedPath + p.separator
}
