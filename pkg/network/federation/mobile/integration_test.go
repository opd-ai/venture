// Integration test for mobile federation with platform detection
// This test demonstrates platform detection and graceful degradation
package mobile

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestIntegration_PlatformDetectionAndFallback(t *testing.T) {
	// This integration test verifies the complete flow:
	// 1. Platform capabilities are detected
	// 2. Adapter created with appropriate configuration
	// 3. System operates correctly regardless of WebRTC availability

	// Detect capabilities
	caps := DetectCapabilities()
	t.Logf("Detected platform: %s", caps.Platform)
	t.Logf("WebRTC available: %v", caps.WebRTCAvailable)
	t.Logf("Fallback transport: %s", GetFallbackTransport())

	// Create adapter with default config
	config := DefaultConfig()
	adapter := NewAdapter(config)

	// Register a simple sync handler
	var syncCalled bool
	adapter.RegisterSyncHandler(func(ctx context.Context) error {
		syncCalled = true
		return nil
	})

	// Start adapter
	err := adapter.Start()
	if err != nil {
		t.Fatalf("Failed to start adapter: %v", err)
	}
	defer adapter.Stop()

	// Simulate battery changes
	adapter.UpdateBatteryLevel(0.8)
	time.Sleep(10 * time.Millisecond)

	// Verify state
	state := adapter.GetState()
	if state.BatteryLevel != 0.8 {
		t.Errorf("Battery level = %.2f, want 0.8", state.BatteryLevel)
	}

	// Log final state
	t.Logf("Final state - Battery: %.2f, Mode: %v, Status: %v, Sync called: %v",
		state.BatteryLevel, state.BatteryMode, state.SyncStatus, syncCalled)
}

func TestIntegration_WebRTCPlatformMatrix(t *testing.T) {
	// This test documents expected WebRTC availability across platforms
	caps := DetectCapabilities()

	platformMatrix := map[string]bool{
		"js":      true,  // WASM/browser
		"android": true,  // Android mobile
		"ios":     true,  // iOS mobile
		"linux":   false, // Desktop Linux
		"darwin":  false, // Desktop macOS
		"windows": false, // Desktop Windows
	}

	expectedAvailability, exists := platformMatrix[runtime.GOOS]
	if !exists {
		t.Logf("Platform %s not in test matrix, skipping validation", runtime.GOOS)
		return
	}

	if caps.WebRTCAvailable != expectedAvailability {
		t.Errorf("WebRTC availability = %v for %s, want %v",
			caps.WebRTCAvailable, runtime.GOOS, expectedAvailability)
	}

	t.Logf("Platform: %s, WebRTC: %v (expected: %v) ✓",
		runtime.GOOS, caps.WebRTCAvailable, expectedAvailability)
}

func TestIntegration_FallbackTransportSelection(t *testing.T) {
	// Test that fallback transport is correctly selected
	fallback := GetFallbackTransport()

	// Should always prefer WebSocket when available
	if fallback != "websocket" {
		t.Errorf("Fallback transport = %s, want websocket", fallback)
	}

	// Verify it matches capabilities
	caps := DetectCapabilities()
	if !caps.WebSocketAvailable {
		t.Error("WebSocket should always be available as fallback")
	}
}

func TestIntegration_GracefulDegradation(t *testing.T) {
	// Test complete graceful degradation flow
	caps := DetectCapabilities()

	// Create adapter
	config := DefaultConfig()
	config.SyncInterval = 100 * time.Millisecond // Fast sync for testing
	adapter := NewAdapter(config)

	// Should work regardless of WebRTC availability
	err := adapter.Start()
	if err != nil {
		t.Fatalf("Adapter should start regardless of WebRTC: %v", err)
	}
	defer adapter.Stop()

	// Perform operations that would use WebRTC if available
	adapter.UpdateBatteryLevel(0.5)
	time.Sleep(50 * time.Millisecond)

	state := adapter.GetState()
	if state.SyncStatus == SyncStatusError {
		t.Error("System should not error due to missing WebRTC")
	}

	if !caps.WebRTCAvailable {
		t.Logf("✓ System running in fallback mode (WebRTC unavailable on %s)", runtime.GOOS)
	} else {
		t.Logf("✓ System running with WebRTC (available on %s)", runtime.GOOS)
	}
}
