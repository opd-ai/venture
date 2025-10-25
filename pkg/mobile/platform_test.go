package mobile

import (
	"runtime"
	"testing"
)

func TestGetPlatform(t *testing.T) {
	// Test current platform detection
	platform := GetPlatform()

	// Verify it returns a valid platform
	switch runtime.GOOS {
	case "ios":
		if platform != PlatformIOS {
			t.Errorf("Expected PlatformIOS for ios, got %v", platform)
		}
	case "android":
		if platform != PlatformAndroid {
			t.Errorf("Expected PlatformAndroid for android, got %v", platform)
		}
	case "js":
		if platform != PlatformWASM {
			t.Errorf("Expected PlatformWASM for js, got %v", platform)
		}
	default:
		if platform != PlatformUnknown {
			t.Errorf("Expected PlatformUnknown for %s, got %v", runtime.GOOS, platform)
		}
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		platform Platform
		expected string
	}{
		{PlatformIOS, "iOS"},
		{PlatformAndroid, "Android"},
		{PlatformWASM, "WASM"},
		{PlatformUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.platform.String()
			if got != tt.expected {
				t.Errorf("Platform.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsMobilePlatform(t *testing.T) {
	// IsMobilePlatform should only return true for iOS and Android, not WASM
	result := IsMobilePlatform()

	switch runtime.GOOS {
	case "ios", "android":
		if !result {
			t.Errorf("IsMobilePlatform() = false for %s, want true", runtime.GOOS)
		}
	default:
		if result {
			t.Errorf("IsMobilePlatform() = true for %s, want false", runtime.GOOS)
		}
	}
}

func TestIsTouchCapable(t *testing.T) {
	// IsTouchCapable should return true for iOS, Android, and WASM
	result := IsTouchCapable()

	switch runtime.GOOS {
	case "ios", "android", "js":
		if !result {
			t.Errorf("IsTouchCapable() = false for %s, want true", runtime.GOOS)
		}
	default:
		if result {
			t.Errorf("IsTouchCapable() = true for %s, want false", runtime.GOOS)
		}
	}
}

func TestIsWASM(t *testing.T) {
	result := IsWASM()

	if runtime.GOOS == "js" {
		if !result {
			t.Error("IsWASM() = false for js, want true")
		}
	} else {
		if result {
			t.Errorf("IsWASM() = true for %s, want false", runtime.GOOS)
		}
	}
}

func TestIsIOS(t *testing.T) {
	result := IsIOS()

	if runtime.GOOS == "ios" {
		if !result {
			t.Error("IsIOS() = false for ios, want true")
		}
	} else {
		if result {
			t.Errorf("IsIOS() = true for %s, want false", runtime.GOOS)
		}
	}
}

func TestIsAndroid(t *testing.T) {
	result := IsAndroid()

	if runtime.GOOS == "android" {
		if !result {
			t.Error("IsAndroid() = false for android, want true")
		}
	} else {
		if result {
			t.Errorf("IsAndroid() = true for %s, want false", runtime.GOOS)
		}
	}
}

func TestGetOrientation(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		expected Orientation
	}{
		{"Landscape", 800, 600, OrientationLandscape},
		{"Portrait", 600, 800, OrientationPortrait},
		{"Square", 600, 600, OrientationUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOrientation(tt.width, tt.height)
			if got != tt.expected {
				t.Errorf("GetOrientation(%d, %d) = %v, want %v",
					tt.width, tt.height, got, tt.expected)
			}
		})
	}
}

func TestOrientationString(t *testing.T) {
	tests := []struct {
		orientation Orientation
		expected    string
	}{
		{OrientationPortrait, "Portrait"},
		{OrientationLandscape, "Landscape"},
		{OrientationUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.orientation.String()
			if got != tt.expected {
				t.Errorf("Orientation.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
