package filesystem

import (
	"fmt"
	"regexp"
	"strings"
)

type WhitespacePathNormalizer struct{}

func (n *WhitespacePathNormalizer) NormalizePath(path string) (string, error) {
	path = strings.ReplaceAll(path, "\\", "/")

	if err := n.rejectFunkyWhiteSpace(path); err != nil {
		return "", err
	}

	return n.normalizeRelativePath(path)
}

func (n *WhitespacePathNormalizer) rejectFunkyWhiteSpace(path string) error {
	if matched, _ := regexp.MatchString(`\p{C}+`, path); matched {
		return fmt.Errorf("corrupted path detected: %s", path)
	}
	return nil
}

func (n *WhitespacePathNormalizer) normalizeRelativePath(path string) (string, error) {
	parts := []string{}

	for _, part := range strings.Split(path, "/") {
		switch part {
		case "", ".":
			// Do nothing
		case "..":
			if len(parts) == 0 {
				return "", fmt.Errorf("path traversal detected: %s", path)
			}
			parts = parts[:len(parts)-1]
		default:
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, "/"), nil
}
