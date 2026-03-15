package vr

import (
	"os"
	"runtime"
	"testing"
)

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	if d == nil {
		t.Fatal("NewDetector returned nil")
	}

	if d.detectionRun {
		t.Error("expected detectionRun to be false initially")
	}
	if d.headsetDetected {
		t.Error("expected headsetDetected to be false initially")
	}
	if d.controllerDetected {
		t.Error("expected controllerDetected to be false initially")
	}
}

func TestDetectorForceEnable(t *testing.T) {
	d := NewDetector()
	d.SetForceEnable(true)

	available := d.DetectHardware()
	if !available {
		t.Error("expected VR to be available when force enabled")
	}

	if !d.IsHeadsetDetected() {
		t.Error("expected headset to be detected when force enabled")
	}
	if !d.IsControllerDetected() {
		t.Error("expected controller to be detected when force enabled")
	}
}

func TestDetectorForceDisable(t *testing.T) {
	d := NewDetector()
	d.SetForceDisable(true)

	available := d.DetectHardware()
	if available {
		t.Error("expected VR to be unavailable when force disabled")
	}

	if d.IsHeadsetDetected() {
		t.Error("expected headset not to be detected when force disabled")
	}
	if d.IsControllerDetected() {
		t.Error("expected controller not to be detected when force disabled")
	}
}

func TestDetectorForceDisableTakesPrecedence(t *testing.T) {
	d := NewDetector()
	d.SetForceEnable(true)
	d.SetForceDisable(true) // Disable should override enable

	available := d.DetectHardware()
	if available {
		t.Error("expected force disable to take precedence over force enable")
	}
}

func TestDetectorForceDisableClearsDetectionState(t *testing.T) {
	d := NewDetector()
	// First, force enable and detect
	d.SetForceEnable(true)
	d.DetectHardware()

	// Verify detection state is set
	if !d.IsHeadsetDetected() {
		t.Error("expected headset to be detected after force enable")
	}
	if !d.IsControllerDetected() {
		t.Error("expected controller to be detected after force enable")
	}

	// Now force disable - this should clear detection state
	d.SetForceDisable(true)

	// All detection methods should now return false for consistency
	if d.DetectHardware() {
		t.Error("expected DetectHardware to return false after force disable")
	}
	if d.IsHeadsetDetected() {
		t.Error("expected IsHeadsetDetected to return false after force disable")
	}
	if d.IsControllerDetected() {
		t.Error("expected IsControllerDetected to return false after force disable")
	}
}

func TestDetectorCaching(t *testing.T) {
	d := NewDetector()
	d.SetForceEnable(true)

	// First call
	result1 := d.DetectHardware()
	if !result1 {
		t.Error("expected first detection to return true")
	}

	// Change force enable (should use cached result)
	d.mu.Lock()
	d.forceEnable = false
	d.mu.Unlock()

	// Second call should return cached result
	result2 := d.DetectHardware()
	if !result2 {
		t.Error("expected second detection to return cached result (true)")
	}
}

func TestDetectorReset(t *testing.T) {
	d := NewDetector()
	d.SetForceEnable(true)

	// First detection
	result1 := d.DetectHardware()
	if !result1 {
		t.Error("expected first detection to return true")
	}

	// Reset and change config
	d.Reset()
	d.SetForceEnable(false)
	d.SetForceDisable(true)

	// Should re-detect and return false
	result2 := d.DetectHardware()
	if result2 {
		t.Error("expected detection after reset to return false")
	}
}

func TestDetectHeadsetEnvironmentVariables(t *testing.T) {
	// Skip on platforms that don't support VR
	if runtime.GOOS == "js" || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skip("skipping VR detection test on unsupported platform")
	}

	tests := []struct {
		name   string
		envVar string
		envVal string
		want   bool
	}{
		{"SteamVR", "STEAMVR_LH_ENABLE", "1", true},
		{"Oculus", "OVR_SDK_PATH", "/opt/oculus", true},
		{"OpenVR", "OPENVR_PATH", "/usr/lib/openvr", true},
		{"NoEnvVar", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			originalVal := os.Getenv(tt.envVar)
			defer func() {
				if originalVal != "" {
					os.Setenv(tt.envVar, originalVal)
				} else {
					os.Unsetenv(tt.envVar)
				}
			}()

			// Clear all VR env vars
			os.Unsetenv("STEAMVR_LH_ENABLE")
			os.Unsetenv("OVR_SDK_PATH")
			os.Unsetenv("OPENVR_PATH")

			// Set test env var
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
			}

			d := NewDetector()
			result := d.detectHeadset()

			if result != tt.want {
				t.Errorf("detectHeadset() with %s=%s = %v, want %v",
					tt.envVar, tt.envVal, result, tt.want)
			}
		})
	}
}

func TestDetectHeadsetPlatformRestrictions(t *testing.T) {
	// This test verifies that mobile/WASM platforms always return false
	// We can't actually change runtime.GOOS in tests, but we can verify
	// the logic by checking the current platform behavior

	d := NewDetector()

	// Set an environment variable that would normally trigger detection
	os.Setenv("STEAMVR_LH_ENABLE", "1")
	defer os.Unsetenv("STEAMVR_LH_ENABLE")

	result := d.detectHeadset()

	switch runtime.GOOS {
	case "js", "android", "ios":
		if result {
			t.Error("expected detectHeadset to return false on mobile/WASM platforms")
		}
	default:
		if !result {
			t.Error("expected detectHeadset to return true with STEAMVR_LH_ENABLE set")
		}
	}
}

func TestDetectController(t *testing.T) {
	d := NewDetector()
	result := d.detectController()

	// Controllers detection is currently conservative (always false)
	if result {
		t.Error("expected detectController to return false (conservative implementation)")
	}
}

func TestParseEnableVRFlag(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"1", true},
		{"on", true},
		{"On", true},
		{"ON", true},
		{"enable", true},
		{"Enable", true},
		{"ENABLE", true},
		{"enabled", true},
		{"Enabled", true},
		{"ENABLED", true},
		{"false", false},
		{"no", false},
		{"0", false},
		{"off", false},
		{"disable", false},
		{"", false},
		{"  true  ", true}, // Test trimming
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseEnableVRFlag(tt.input)
			if got != tt.want {
				t.Errorf("ParseEnableVRFlag(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsHeadsetDetected(t *testing.T) {
	d := NewDetector()

	// Before detection
	if d.IsHeadsetDetected() {
		t.Error("expected IsHeadsetDetected to return false before detection")
	}

	// After force enable
	d.SetForceEnable(true)
	d.DetectHardware()

	if !d.IsHeadsetDetected() {
		t.Error("expected IsHeadsetDetected to return true after force enable")
	}
}

func TestIsControllerDetected(t *testing.T) {
	d := NewDetector()

	// Before detection
	if d.IsControllerDetected() {
		t.Error("expected IsControllerDetected to return false before detection")
	}

	// After force enable
	d.SetForceEnable(true)
	d.DetectHardware()

	if !d.IsControllerDetected() {
		t.Error("expected IsControllerDetected to return true after force enable")
	}
}

func TestCheckVRRuntimePaths(t *testing.T) {
	d := NewDetector()

	// This test will pass/fail based on actual system state
	// We mainly verify it doesn't panic
	result := d.checkVRRuntimePaths()

	// Result depends on whether VR is actually installed
	// We just verify it returns a boolean without crashing
	t.Logf("checkVRRuntimePaths returned: %v (platform: %s)", result, runtime.GOOS)
}

func TestDetectorConcurrency(t *testing.T) {
	d := NewDetector()
	d.SetForceEnable(true)

	// Concurrent access test
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			d.DetectHardware()
			d.IsHeadsetDetected()
			d.IsControllerDetected()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkDetectHardware benchmarks hardware detection.
func BenchmarkDetectHardware(b *testing.B) {
	d := NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Reset()
		d.DetectHardware()
	}
}

// BenchmarkDetectHardwareCached benchmarks cached detection.
func BenchmarkDetectHardwareCached(b *testing.B) {
	d := NewDetector()
	d.DetectHardware() // Prime the cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.DetectHardware()
	}
}

// BenchmarkIsHeadsetDetected benchmarks the IsHeadsetDetected read-path under RWMutex.
func BenchmarkIsHeadsetDetected(b *testing.B) {
	d := NewDetector()
	d.DetectHardware() // Prime the cache so read-path is exercised

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsHeadsetDetected()
	}
}

// BenchmarkIsControllerDetected benchmarks the IsControllerDetected read-path under RWMutex.
func BenchmarkIsControllerDetected(b *testing.B) {
	d := NewDetector()
	d.DetectHardware() // Prime the cache so read-path is exercised

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsControllerDetected()
	}
}

// BenchmarkIsHeadsetDetectedParallel benchmarks concurrent read contention on RWMutex.
func BenchmarkIsHeadsetDetectedParallel(b *testing.B) {
	d := NewDetector()
	d.DetectHardware() // Prime the cache

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.IsHeadsetDetected()
		}
	})
}
