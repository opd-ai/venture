package mobile

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDetectCapabilities(t *testing.T) {
	caps := DetectCapabilities()

	if caps == nil {
		t.Fatal("DetectCapabilities returned nil")
	}

	// All platforms should have HTTP and WebSocket
	if !caps.HTTPAvailable {
		t.Error("HTTPAvailable should be true on all platforms")
	}

	if !caps.WebSocketAvailable {
		t.Error("WebSocketAvailable should be true on all platforms")
	}

	// Platform should match runtime
	if caps.Platform != runtime.GOOS {
		t.Errorf("Platform = %s, want %s", caps.Platform, runtime.GOOS)
	}

	// WebRTC availability depends on platform
	switch runtime.GOOS {
	case "js", "android", "ios":
		if !caps.WebRTCAvailable {
			t.Errorf("WebRTCAvailable should be true on %s", runtime.GOOS)
		}
	default:
		if caps.WebRTCAvailable {
			t.Errorf("WebRTCAvailable should be false on %s", runtime.GOOS)
		}
		if len(caps.Restrictions) == 0 {
			t.Error("Expected restrictions on non-mobile platform")
		}
	}
}

func TestSupportsWebRTC(t *testing.T) {
	supported := SupportsWebRTC()

	// Should match current platform
	switch runtime.GOOS {
	case "js", "android", "ios":
		if !supported {
			t.Errorf("SupportsWebRTC() = false on %s, want true", runtime.GOOS)
		}
	default:
		if supported {
			t.Errorf("SupportsWebRTC() = true on %s, want false", runtime.GOOS)
		}
	}
}

func TestGetFallbackTransport(t *testing.T) {
	fallback := GetFallbackTransport()

	// Should always have a fallback
	if fallback == "" {
		t.Error("GetFallbackTransport returned empty string")
	}

	// Should prefer WebSocket
	if fallback != "websocket" {
		t.Errorf("GetFallbackTransport() = %s, want websocket", fallback)
	}
}

func TestLogCapabilities(t *testing.T) {
	// Create a logger with buffer to capture output
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)

	entry := logrus.NewEntry(logger)

	// Should not panic
	LogCapabilities(entry)

	// Should have logged something
	if buf.Len() == 0 {
		t.Error("LogCapabilities did not log anything")
	}

	output := buf.String()

	// Should contain platform info
	if !contains(output, "platform") {
		t.Error("Log output missing platform info")
	}

	// Should contain capability info
	if !contains(output, "webrtc") && !contains(output, "websocket") {
		t.Error("Log output missing capability info")
	}
}

func TestPlatformCapabilitiesRestrictions(t *testing.T) {
	caps := DetectCapabilities()

	// WASM should have restrictions
	if runtime.GOOS == "js" {
		if len(caps.Restrictions) == 0 {
			t.Error("WASM platform should have restrictions listed")
		}

		hasHTTPSRestriction := false
		for _, r := range caps.Restrictions {
			if contains(r, "HTTPS") {
				hasHTTPSRestriction = true
				break
			}
		}

		if !hasHTTPSRestriction {
			t.Error("WASM should list HTTPS restriction for WebRTC")
		}
	}

	// Desktop should mention library requirement
	if runtime.GOOS != "js" && runtime.GOOS != "android" && runtime.GOOS != "ios" {
		if len(caps.Restrictions) == 0 {
			t.Error("Desktop platform should list library requirement")
		}

		hasLibraryRestriction := false
		for _, r := range caps.Restrictions {
			if contains(r, "library") || contains(r, "pion") {
				hasLibraryRestriction = true
				break
			}
		}

		if !hasLibraryRestriction {
			t.Error("Desktop should mention WebRTC library requirement")
		}
	}
}

func TestNetworkCapabilityConstants(t *testing.T) {
	// Ensure capability flags are distinct
	caps := []NetworkCapability{
		CapabilityWebRTC,
		CapabilityWebSocket,
		CapabilityHTTP,
	}

	seen := make(map[NetworkCapability]bool)
	for _, cap := range caps {
		if seen[cap] {
			t.Errorf("Duplicate capability value: %d", cap)
		}
		seen[cap] = true
	}

	// Ensure they are power of 2 (bitflags)
	if CapabilityWebRTC&CapabilityWebSocket != 0 {
		t.Error("Capability flags should not overlap")
	}
	if CapabilityWebRTC&CapabilityHTTP != 0 {
		t.Error("Capability flags should not overlap")
	}
	if CapabilityWebSocket&CapabilityHTTP != 0 {
		t.Error("Capability flags should not overlap")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
