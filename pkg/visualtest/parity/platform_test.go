package parity

import (
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	// Should detect a valid platform
	if platform == PlatformUnknown {
		t.Error("DetectPlatform returned Unknown platform")
	}

	// Platform should match GOOS
	expectedPlatforms := map[string]Platform{
		"linux":   PlatformLinux,
		"darwin":  PlatformMacOS,
		"windows": PlatformWindows,
		"android": PlatformAndroid,
		"ios":     PlatformIOS,
	}

	if runtime.GOOS == "js" && runtime.GOARCH == "wasm" {
		if platform != PlatformWASM {
			t.Errorf("expected PlatformWASM for js/wasm, got %v", platform)
		}
	} else if expected, ok := expectedPlatforms[runtime.GOOS]; ok {
		if platform != expected {
			t.Errorf("expected %v for GOOS=%s, got %v", expected, runtime.GOOS, platform)
		}
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{PlatformLinux, "Linux"},
		{PlatformMacOS, "macOS"},
		{PlatformWindows, "Windows"},
		{PlatformWASM, "WebAssembly"},
		{PlatformIOS, "iOS"},
		{PlatformAndroid, "Android"},
		{PlatformUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.platform.String()
			if got != tt.want {
				t.Errorf("Platform.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformIsDesktop(t *testing.T) {
	tests := []struct {
		platform Platform
		want     bool
	}{
		{PlatformLinux, true},
		{PlatformMacOS, true},
		{PlatformWindows, true},
		{PlatformWASM, false},
		{PlatformIOS, false},
		{PlatformAndroid, false},
		{PlatformUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.String(), func(t *testing.T) {
			got := tt.platform.IsDesktop()
			if got != tt.want {
				t.Errorf("Platform.IsDesktop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformIsMobile(t *testing.T) {
	tests := []struct {
		platform Platform
		want     bool
	}{
		{PlatformLinux, false},
		{PlatformMacOS, false},
		{PlatformWindows, false},
		{PlatformWASM, false},
		{PlatformIOS, true},
		{PlatformAndroid, true},
		{PlatformUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.String(), func(t *testing.T) {
			got := tt.platform.IsMobile()
			if got != tt.want {
				t.Errorf("Platform.IsMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformIsWeb(t *testing.T) {
	tests := []struct {
		platform Platform
		want     bool
	}{
		{PlatformLinux, false},
		{PlatformMacOS, false},
		{PlatformWindows, false},
		{PlatformWASM, true},
		{PlatformIOS, false},
		{PlatformAndroid, false},
		{PlatformUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.String(), func(t *testing.T) {
			got := tt.platform.IsWeb()
			if got != tt.want {
				t.Errorf("Platform.IsWeb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	// Basic validation
	if info.Platform == PlatformUnknown {
		t.Error("GetPlatformInfo returned Unknown platform")
	}

	if info.GOOS == "" {
		t.Error("GetPlatformInfo returned empty GOOS")
	}

	if info.GOARCH == "" {
		t.Error("GetPlatformInfo returned empty GOARCH")
	}

	if info.NumCPU <= 0 {
		t.Errorf("GetPlatformInfo returned invalid NumCPU: %d", info.NumCPU)
	}

	// Validate feature flags match platform type
	if info.Platform.IsDesktop() && !info.SupportsFullscreen {
		t.Error("Desktop platform should support fullscreen")
	}

	if info.Platform.IsMobile() && !info.SupportsTouch {
		t.Error("Mobile platform should support touch")
	}

	if info.Platform.IsWeb() && !info.SupportsWebGL {
		t.Error("Web platform should support WebGL")
	}
}

func BenchmarkDetectPlatform(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DetectPlatform()
	}
}

func BenchmarkGetPlatformInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetPlatformInfo()
	}
}
