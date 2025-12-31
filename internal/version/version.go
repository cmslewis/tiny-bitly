package version

// Version information is set at build time using -ldflags.
// These variables default to "dev" values if not set during build.

var (
	// Version is the semantic version of the application (e.g., "1.0.0").
	// Set at build time with: -ldflags "-X tiny-bitly/internal/version.Version=1.0.0"
	Version = "dev"

	// Commit is the Git commit hash of the build.
	// Set at build time with: -ldflags "-X tiny-bitly/internal/version.Commit=$(git rev-parse HEAD)"
	Commit = "unknown"

	// BuildTime is the timestamp when the binary was built (RFC3339 format).
	// Set at build time with: -ldflags "-X tiny-bitly/internal/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildTime = "unknown"
)

// Info returns a map of all version information.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildTime": BuildTime,
	}
}
