// Package version provides version information for claude-loop.
package version

var (
	// Version is the semantic version of claude-loop.
	// This is set at build time via ldflags.
	Version = "v0.18.0"

	// GitCommit is the git commit hash.
	// This is set at build time via ldflags.
	GitCommit = "unknown"

	// BuildDate is the build date.
	// This is set at build time via ldflags.
	BuildDate = "unknown"
)

// Info returns version information as a formatted string.
func Info() string {
	return "claude-loop version " + Version
}

// Full returns detailed version information.
func Full() string {
	return "claude-loop version " + Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
