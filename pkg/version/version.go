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
	"strconv"
	"strings"
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

	// ProtocolVersion is the network federation protocol version.
	// Protocol versions follow semantic versioning independently of application version.
	// Compatibility is determined by major version: 6.x.x is compatible with 6.y.z.
	ProtocolVersion = "6.0.0"

	// ProtocolMajor is the major version of the federation protocol.
	ProtocolMajor = 6

	// ProtocolMinor is the minor version of the federation protocol.
	ProtocolMinor = 0

	// ProtocolPatch is the patch version of the federation protocol.
	ProtocolPatch = 0
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

// ParseVersion parses a semantic version string into major, minor, patch components.
// Returns an error if the version string is not valid semver format (e.g., "1.2.3").
func ParseVersion(v string) (major, minor, patch int, err error) {
	if v == "" {
		return 0, 0, 0, fmt.Errorf("empty version string")
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: expected MAJOR.MINOR.PATCH, got %q", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	if major < 0 {
		return 0, 0, 0, fmt.Errorf("negative major version: %d", major)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	if minor < 0 {
		return 0, 0, 0, fmt.Errorf("negative minor version: %d", minor)
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}
	if patch < 0 {
		return 0, 0, 0, fmt.Errorf("negative patch version: %d", patch)
	}

	return major, minor, patch, nil
}

// Compare compares two semantic version strings.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// Returns an error if either version string is invalid.
func Compare(v1, v2 string) (int, error) {
	maj1, min1, pat1, err := ParseVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid first version: %w", err)
	}

	maj2, min2, pat2, err := ParseVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid second version: %w", err)
	}

	if maj1 != maj2 {
		if maj1 < maj2 {
			return -1, nil
		}
		return 1, nil
	}

	if min1 != min2 {
		if min1 < min2 {
			return -1, nil
		}
		return 1, nil
	}

	if pat1 != pat2 {
		if pat1 < pat2 {
			return -1, nil
		}
		return 1, nil
	}

	return 0, nil
}

// IsCompatible checks if two semantic version strings are compatible.
// Two versions are compatible if they share the same major version number.
// This is used for network protocol version negotiation.
// Returns an error if either version string is invalid.
func IsCompatible(v1, v2 string) (bool, error) {
	maj1, _, _, err := ParseVersion(v1)
	if err != nil {
		return false, fmt.Errorf("invalid first version: %w", err)
	}

	maj2, _, _, err := ParseVersion(v2)
	if err != nil {
		return false, fmt.Errorf("invalid second version: %w", err)
	}

	return maj1 == maj2, nil
}
