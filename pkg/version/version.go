// Package version provides version information for Venture.
package version

const (
	// Version is the current version of Venture.
	Version = "7.0.0"

	// Release indicates the release status.
	Release = "Production"

	// FullVersion returns the complete version string.
	FullVersion = Version + " " + Release
)
