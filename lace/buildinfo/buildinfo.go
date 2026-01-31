package buildinfo

import "go.opentelemetry.io/otel/attribute"

// These variables are set at build time via -ldflags
var (
	// Version is the semantic version (e.g., "1.0.0")
	Version = "dev"

	// Commit is the git commit SHA
	Commit = "unknown"

	// BuildTime is the UTC build timestamp
	BuildTime = "unknown"

	// Branch is the git branch name
	Branch = "unknown"
)

// Info returns build information as a map
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     Commit,
		"build_time": BuildTime,
		"branch":     Branch,
	}
}

func WrapOtelAttributes(attrs ...attribute.KeyValue) []attribute.KeyValue {
	return append(attrs,
		attribute.String("build.version", Version),
		attribute.String("build.commit", Commit),
		attribute.String("build.branch", Branch),
		attribute.String("build.time", BuildTime),
	)
}
