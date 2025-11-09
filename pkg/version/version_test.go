package version

import (
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
