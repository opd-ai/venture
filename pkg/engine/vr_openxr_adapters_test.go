//go:build vr && !js

package engine

import "testing"

// TestOpenXRHeadsetAdapter_NoRuntime verifies that NewOpenXRHeadsetAdapter
// completes without panicking when no OpenXR runtime is present (the
// common CI/development case) and that the adapter reports disconnected,
// satisfying the factory's fallback contract.
func TestOpenXRHeadsetAdapter_NoRuntime(t *testing.T) {
	a := NewOpenXRHeadsetAdapter()
	if a == nil {
		t.Fatal("NewOpenXRHeadsetAdapter returned nil")
	}
	// In CI there is no OpenXR runtime; connected must be false so the
	// factory (vr_adapter_factory_openxr.go) falls through to the stub.
	if a.connected {
		// Runtime is available — verify that IsConnected is consistent.
		if !a.IsConnected() {
			t.Error("IsConnected() returned false but connected field is true")
		}
		// Pose calls should not panic.
		p, y, r := a.GetHeadOrientation()
		t.Logf("head orientation (runtime present): pitch=%.4f yaw=%.4f roll=%.4f", p, y, r)
		x, yy, z := a.GetHeadPosition()
		t.Logf("head position    (runtime present): x=%.4f y=%.4f z=%.4f", x, yy, z)
		return
	}
	// No runtime: IsConnected must be false and pose calls must return zeros.
	if a.IsConnected() {
		t.Error("IsConnected() returned true but connected field is false")
	}
	p, y, r := a.GetHeadOrientation()
	if p != 0 || y != 0 || r != 0 {
		t.Errorf("GetHeadOrientation() on disconnected adapter = (%v,%v,%v); want (0,0,0)", p, y, r)
	}
	x, yy, z := a.GetHeadPosition()
	if x != 0 || yy != 0 || z != 0 {
		t.Errorf("GetHeadPosition() on disconnected adapter = (%v,%v,%v); want (0,0,0)", x, yy, z)
	}
	if got := a.GetIPD(); got <= 0 {
		t.Errorf("GetIPD() = %v; want > 0", got)
	}
}

// TestOpenXRControllerAdapter_NoRuntime verifies that NewOpenXRControllerAdapter
// completes without panicking when no OpenXR runtime is present and that all
// input methods return safe zero/false values.
func TestOpenXRControllerAdapter_NoRuntime(t *testing.T) {
	a := NewOpenXRControllerAdapter()
	if a == nil {
		t.Fatal("NewOpenXRControllerAdapter returned nil")
	}
	if a.connected {
		// Runtime present — check consistency only; we can't assert input values.
		if !a.IsConnected("left") && !a.IsConnected("right") {
			t.Error("connected=true but IsConnected returns false for both hands")
		}
		return
	}
	for _, hand := range []string{"left", "right"} {
		if a.IsConnected(hand) {
			t.Errorf("IsConnected(%q) returned true but no runtime is available", hand)
		}
		if v := a.GetTrigger(hand); v != 0 {
			t.Errorf("GetTrigger(%q) = %v; want 0 when disconnected", hand, v)
		}
		if v := a.GetGrip(hand); v != 0 {
			t.Errorf("GetGrip(%q) = %v; want 0 when disconnected", hand, v)
		}
		if x, y := a.GetThumbstick(hand); x != 0 || y != 0 {
			t.Errorf("GetThumbstick(%q) = (%v,%v); want (0,0) when disconnected", hand, x, y)
		}
		if a.IsThumbstickPressed(hand) {
			t.Errorf("IsThumbstickPressed(%q) returned true when disconnected", hand)
		}
		for _, btn := range []string{"a", "b", "x", "y", "menu", "system"} {
			if a.GetButton(hand, btn) {
				t.Errorf("GetButton(%q, %q) returned true when disconnected", hand, btn)
			}
		}
		// SetHaptic must not panic when disconnected.
		a.SetHaptic(hand, 0.5, 0.1)
	}
}

// TestOpenXRAdapters_FactoryFallback verifies that NewRuntimeHeadsetAdapter and
// NewRuntimeControllerAdapter always return non-nil adapters (either OpenXR or
// stub depending on runtime availability).
func TestOpenXRAdapters_FactoryFallback(t *testing.T) {
	h := NewRuntimeHeadsetAdapter()
	if h == nil {
		t.Error("NewRuntimeHeadsetAdapter() returned nil")
	}
	c := NewRuntimeControllerAdapter()
	if c == nil {
		t.Error("NewRuntimeControllerAdapter() returned nil")
	}
}
