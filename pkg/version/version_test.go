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

func TestPrintVersion(t *testing.T) {
	// Capture stdout to verify PrintVersion output
	// Since PrintVersion uses fmt.Println which writes to os.Stdout,
	// we test that it doesn't panic and produces consistent output
	// by verifying it matches BuildInfo
	expectedOutput := BuildInfo()

	// Call PrintVersion to ensure it doesn't panic and exercises the code path
	// We can't easily capture stdout in Go tests without modifying the function,
	// but we can verify the function runs without error
	PrintVersion()

	// Verify the output format matches what we expect
	if !strings.Contains(expectedOutput, "Venture") {
		t.Errorf("BuildInfo() = %q, want to contain 'Venture'", expectedOutput)
	}
	if !strings.Contains(expectedOutput, Version) {
		t.Errorf("BuildInfo() = %q, want to contain version %q", expectedOutput, Version)
	}
}

func TestMajorMinorPatchConsistency(t *testing.T) {
	// Verify Major, Minor, Patch constants are non-negative
	if Major < 0 {
		t.Errorf("Major = %d, want non-negative", Major)
	}
	if Minor < 0 {
		t.Errorf("Minor = %d, want non-negative", Minor)
	}
	if Patch < 0 {
		t.Errorf("Patch = %d, want non-negative", Patch)
	}

	// Verify for production release, major should be at least 1
	if Release == "Production" && Major < 1 {
		t.Errorf("Production release should have Major >= 1, got %d", Major)
	}
}

func TestProtocolVersionConstants(t *testing.T) {
	// Verify that ProtocolVersion follows semver format
	parts := strings.Split(ProtocolVersion, ".")
	if len(parts) != 3 {
		t.Errorf("ProtocolVersion = %q, want MAJOR.MINOR.PATCH format", ProtocolVersion)
	}

	// Verify components match the string
	expectedVersion := fmt.Sprintf("%d.%d.%d", ProtocolMajor, ProtocolMinor, ProtocolPatch)
	if ProtocolVersion != expectedVersion {
		t.Errorf("ProtocolVersion = %q, want %q (from components)", ProtocolVersion, expectedVersion)
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{
			name:      "valid version",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "application version",
			version:   Version,
			wantMajor: Major,
			wantMinor: Minor,
			wantPatch: Patch,
			wantErr:   false,
		},
		{
			name:      "protocol version",
			version:   ProtocolVersion,
			wantMajor: ProtocolMajor,
			wantMinor: ProtocolMinor,
			wantPatch: ProtocolPatch,
			wantErr:   false,
		},
		{
			name:      "zero version",
			version:   "0.0.0",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "large numbers",
			version:   "100.200.300",
			wantMajor: 100,
			wantMinor: 200,
			wantPatch: 300,
			wantErr:   false,
		},
		{
			name:    "empty string",
			version: "",
			wantErr: true,
		},
		{
			name:    "missing patch",
			version: "1.2",
			wantErr: true,
		},
		{
			name:    "missing minor and patch",
			version: "1",
			wantErr: true,
		},
		{
			name:    "too many parts",
			version: "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "non-numeric major",
			version: "a.2.3",
			wantErr: true,
		},
		{
			name:    "non-numeric minor",
			version: "1.b.3",
			wantErr: true,
		},
		{
			name:    "non-numeric patch",
			version: "1.2.c",
			wantErr: true,
		},
		{
			name:    "negative major",
			version: "-1.0.0",
			wantErr: true,
		},
		{
			name:    "negative minor",
			version: "1.-2.0",
			wantErr: true,
		},
		{
			name:    "negative patch",
			version: "1.0.-3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor {
					t.Errorf("ParseVersion(%q) major = %d, want %d", tt.version, major, tt.wantMajor)
				}
				if minor != tt.wantMinor {
					t.Errorf("ParseVersion(%q) minor = %d, want %d", tt.version, minor, tt.wantMinor)
				}
				if patch != tt.wantPatch {
					t.Errorf("ParseVersion(%q) patch = %d, want %d", tt.version, patch, tt.wantPatch)
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name    string
		v1      string
		v2      string
		want    int
		wantErr bool
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: 0,
		},
		{
			name: "v1 major less than v2",
			v1:   "1.0.0",
			v2:   "2.0.0",
			want: -1,
		},
		{
			name: "v1 major greater than v2",
			v1:   "2.0.0",
			v2:   "1.0.0",
			want: 1,
		},
		{
			name: "v1 minor less than v2",
			v1:   "1.1.0",
			v2:   "1.2.0",
			want: -1,
		},
		{
			name: "v1 minor greater than v2",
			v1:   "1.3.0",
			v2:   "1.2.0",
			want: 1,
		},
		{
			name: "v1 patch less than v2",
			v1:   "1.2.1",
			v2:   "1.2.3",
			want: -1,
		},
		{
			name: "v1 patch greater than v2",
			v1:   "1.2.5",
			v2:   "1.2.3",
			want: 1,
		},
		{
			name: "compare with application version",
			v1:   Version,
			v2:   Version,
			want: 0,
		},
		{
			name: "compare with protocol version",
			v1:   ProtocolVersion,
			v2:   ProtocolVersion,
			want: 0,
		},
		{
			name:    "invalid first version",
			v1:      "invalid",
			v2:      "1.0.0",
			wantErr: true,
		},
		{
			name:    "invalid second version",
			v1:      "1.0.0",
			v2:      "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compare(tt.v1, tt.v2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compare(%q, %q) error = %v, wantErr %v", tt.v1, tt.v2, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name    string
		v1      string
		v2      string
		want    bool
		wantErr bool
	}{
		{
			name:    "same version",
			v1:      "6.0.0",
			v2:      "6.0.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "same major different minor",
			v1:      "6.0.0",
			v2:      "6.1.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "same major different patch",
			v1:      "6.0.0",
			v2:      "6.0.1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "same major different minor and patch",
			v1:      "6.0.0",
			v2:      "6.5.3",
			want:    true,
			wantErr: false,
		},
		{
			name:    "different major versions",
			v1:      "5.0.0",
			v2:      "6.0.0",
			want:    false,
			wantErr: false,
		},
		{
			name:    "application version compatible with itself",
			v1:      Version,
			v2:      Version,
			want:    true,
			wantErr: false,
		},
		{
			name:    "protocol version compatible with itself",
			v1:      ProtocolVersion,
			v2:      ProtocolVersion,
			want:    true,
			wantErr: false,
		},
		{
			name:    "invalid first version",
			v1:      "invalid",
			v2:      "1.0.0",
			want:    false,
			wantErr: true,
		},
		{
			name:    "invalid second version",
			v1:      "1.0.0",
			v2:      "invalid",
			want:    false,
			wantErr: true,
		},
		{
			name:    "empty first version",
			v1:      "",
			v2:      "1.0.0",
			want:    false,
			wantErr: true,
		},
		{
			name:    "empty second version",
			v1:      "1.0.0",
			v2:      "",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsCompatible(tt.v1, tt.v2)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsCompatible(%q, %q) error = %v, wantErr %v", tt.v1, tt.v2, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsCompatible(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
