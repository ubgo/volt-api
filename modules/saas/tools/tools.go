//go:build tools
// +build tools

// tools is a dummy package that will be ignored for builds, but included for dependencies
package tools

import (
	_ "lace/gqlmodel"
	// _ "github.com/99designs/gqlgen"
)
