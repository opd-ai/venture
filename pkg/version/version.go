// Package version provides version information for Venture.
// Venture follows Semantic Versioning (https://semver.org/):
//   - MAJOR: Incompatible API changes or breaking save file format changes
//   - MINOR: New functionality in a backward-compatible manner
//   - PATCH: Backward-compatible bug fixes
//
// API Stability Guarantees:
//   - Network protocol is stable within a MAJOR version
//   - Save file format is backward-compatible within a MAJOR version
//   - CLI flags are stable within a MAJOR version
//   - Deprecation requires 2 MINOR version notice before removal
package version

import (
	"fmt"
	"runtime"
)

const (
	// Major is the major version number (breaking changes).
	Major = 1

	// Minor is the minor version number (new features).
	Minor = 0

	// Patch is the patch version number (bug fixes).
	Patch = 0

	// Version is the semantic version string (MAJOR.MINOR.PATCH).
	Version = "1.0.0"

	// Release indicates the release status (Production, Beta, Alpha, Development).
	Release = "Production"

	// FullVersion returns the complete version string.
	FullVersion = Version + " " + Release
)

// BuildInfo returns build information including Go version and platform.
func BuildInfo() string {
	return fmt.Sprintf("Venture %s (%s/%s, %s)",
		FullVersion,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version())
}

// ShortVersion returns just the version number without release status.
func ShortVersion() string {
	return Version
}

// PrintVersion prints version information to stdout.
func PrintVersion() {
	fmt.Println(BuildInfo())
}
