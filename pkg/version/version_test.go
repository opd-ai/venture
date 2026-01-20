package version

import (
	"fmt"
	"strings"
	"testing"
)

func TestVersionConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		check func(string) bool
	}{
		{
			name:  "Version is not empty",
			value: Version,
			check: func(v string) bool { return v != "" },
		},
		{
			name:  "Version follows semver format",
			value: Version,
			check: func(v string) bool {
				parts := strings.Split(v, ".")
				return len(parts) == 3
			},
		},
		{
			name:  "Release is not empty",
			value: Release,
			check: func(v string) bool { return v != "" },
		},
		{
			name:  "FullVersion contains Version",
			value: FullVersion,
			check: func(v string) bool { return strings.Contains(v, Version) },
		},
		{
			name:  "FullVersion contains Release",
			value: FullVersion,
			check: func(v string) bool { return strings.Contains(v, Release) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.value) {
				t.Errorf("%s failed: value = %q", tt.name, tt.value)
			}
		})
	}
}

func TestSemanticVersionComponents(t *testing.T) {
	// Verify that constants match the Version string
	expectedVersion := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	if Version != expectedVersion {
		t.Errorf("Version constant %q does not match components %d.%d.%d",
			Version, Major, Minor, Patch)
	}
}

func TestBuildInfo(t *testing.T) {
	info := BuildInfo()

	// BuildInfo should contain version
	if !strings.Contains(info, Version) {
		t.Errorf("BuildInfo() = %q, want to contain %q", info, Version)
	}

	// BuildInfo should contain "Venture"
	if !strings.Contains(info, "Venture") {
		t.Errorf("BuildInfo() = %q, want to contain 'Venture'", info)
	}

	// BuildInfo should contain OS/arch info (e.g., linux/amd64)
	if !strings.Contains(info, "/") {
		t.Errorf("BuildInfo() = %q, want to contain OS/arch separator '/'", info)
	}
}

func TestShortVersion(t *testing.T) {
	short := ShortVersion()
	if short != Version {
		t.Errorf("ShortVersion() = %q, want %q", short, Version)
	}
}

func TestVersionFormat(t *testing.T) {
	// Version should be v1.0.0 for production release
	if Version != "1.0.0" {
		t.Logf("Version is %q (expected 1.0.0 for production release)", Version)
	}

	// Release should be one of the valid statuses
	validReleases := []string{"Production", "Beta", "Alpha", "Development"}
	valid := false
	for _, r := range validReleases {
		if Release == r {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("Release = %q, want one of %v", Release, validReleases)
	}
}
