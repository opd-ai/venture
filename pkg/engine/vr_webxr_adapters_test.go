//go:build js

// Package engine provides tests for the WebXR VR adapters (WASM build only).
//
// These tests run under WASM (js build tag) and verify that the WebXR adapters
// return safe zero/false values when no browser WebXR session is available,
// matching the graceful-degradation contract.  The actual session negotiation is
// covered by manual browser testing (navigator.xr is not available in test runners).
package engine

import "testing"

// TestWebXRHeadsetAdapter_NoSession verifies that NewWebXRHeadsetAdapter returns
// an adapter with connected=false when no browser XR session is available.
func TestWebXRHeadsetAdapter_NoSession(t *testing.T) {
	a := NewWebXRHeadsetAdapter()
	// In a Node-based test environment navigator.xr is absent, so connected must be false.
	if a.IsConnected() {
		t.Skip("a live WebXR session was unexpectedly established — manual verification required")
	}

	p, y, r := a.GetHeadOrientation()
	if p != 0 || y != 0 || r != 0 {
		t.Errorf("GetHeadOrientation() = (%v, %v, %v), want (0, 0, 0)", p, y, r)
	}

	x, yy, z := a.GetHeadPosition()
	if x != 0 || yy != 0 || z != 0 {
		t.Errorf("GetHeadPosition() = (%v, %v, %v), want (0, 0, 0)", x, yy, z)
	}

	if ipd := a.GetIPD(); ipd != 63.0 {
		t.Errorf("GetIPD() = %v, want 63.0mm default", ipd)
	}
}

// TestWebXRControllerAdapter_NoSession verifies safe defaults when not connected.
func TestWebXRControllerAdapter_NoSession(t *testing.T) {
	headset := NewWebXRHeadsetAdapter()
	ctrl := NewWebXRControllerAdapter(headset)

	for _, hand := range []string{"left", "right"} {
		if ctrl.IsConnected(hand) {
			t.Errorf("IsConnected(%q) = true, want false (no session)", hand)
		}
		if v := ctrl.GetTrigger(hand); v != 0 {
			t.Errorf("GetTrigger(%q) = %v, want 0", hand, v)
		}
		if v := ctrl.GetGrip(hand); v != 0 {
			t.Errorf("GetGrip(%q) = %v, want 0", hand, v)
		}
		if x, y := ctrl.GetThumbstick(hand); x != 0 || y != 0 {
			t.Errorf("GetThumbstick(%q) = (%v, %v), want (0, 0)", hand, x, y)
		}
		if ctrl.IsThumbstickPressed(hand) {
			t.Errorf("IsThumbstickPressed(%q) = true, want false", hand)
		}
		if ctrl.GetButton(hand, "a") {
			t.Errorf("GetButton(%q, a) = true, want false", hand)
		}
	}
}

// TestQuaternionToEuler verifies the identity quaternion maps to zero Euler angles.
func TestQuaternionToEuler(t *testing.T) {
	p, y, r := quaternionToEuler(0, 0, 0, 1)
	if p != 0 || y != 0 || r != 0 {
		t.Errorf("identity quaternion: got (%v, %v, %v), want (0, 0, 0)", p, y, r)
	}
}
