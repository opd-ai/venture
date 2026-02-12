// Package engine provides stub implementations of VR adapters for production use.
//
// These stub adapters provide the foundation for future VR hardware SDK integration
// (OpenVR/OpenXR). They are production-ready but intentionally return no-hardware-detected
// states to enable graceful degradation in VR mode.
//
// EXPERIMENTAL: VR support is currently experimental. These stub adapters will be
// replaced with real SDK implementations in future releases.

package engine

import (
	log "github.com/sirupsen/logrus"
)

// StubHeadsetAdapter is a production stub implementation of VRHeadsetAdapter.
// It provides the interface contract for VR headset hardware but reports
// no hardware detected, enabling graceful degradation to mouse fallback.
//
// Future work: Replace with OpenVR/OpenXR SDK integration.
type StubHeadsetAdapter struct{}

// NewStubHeadsetAdapter creates a production stub headset adapter.
// This adapter reports no hardware connected and is used when VR runtime
// is requested but no VR SDK is integrated.
func NewStubHeadsetAdapter() *StubHeadsetAdapter {
	log.WithFields(log.Fields{
		"adapter_type": "vr_headset",
		"status":       "stub",
	}).Debug("Creating stub VR headset adapter (no hardware SDK)")
	return &StubHeadsetAdapter{}
}

// IsConnected reports false, indicating no VR hardware detected.
// This enables the head tracking system to fall back to mouse input.
func (s *StubHeadsetAdapter) IsConnected() bool {
	return false
}

// GetHeadOrientation returns zero orientation since no hardware is connected.
// Returns (pitch=0, yaw=0, roll=0) in radians.
func (s *StubHeadsetAdapter) GetHeadOrientation() (pitch, yaw, roll float64) {
	return 0, 0, 0
}

// GetHeadPosition returns zero position since no hardware is connected.
// Returns (x=0, y=0, z=0) in meters.
func (s *StubHeadsetAdapter) GetHeadPosition() (x, y, z float64) {
	return 0, 0, 0
}

// GetIPD returns a standard interpupillary distance of 63mm.
// This is a typical adult IPD value used for stereoscopic rendering
// even when no hardware is present.
func (s *StubHeadsetAdapter) GetIPD() float64 {
	return 63.0
}

// StubControllerAdapter is a production stub implementation of VRControllerAdapter.
// It provides the interface contract for VR controller hardware but reports
// no controllers detected, enabling graceful degradation to keyboard/mouse input.
//
// Future work: Replace with OpenVR/OpenXR SDK integration.
type StubControllerAdapter struct{}

// NewStubControllerAdapter creates a production stub controller adapter.
// This adapter reports no controllers connected and is used when VR runtime
// is requested but no VR SDK is integrated.
func NewStubControllerAdapter() *StubControllerAdapter {
	log.WithFields(log.Fields{
		"adapter_type": "vr_controller",
		"status":       "stub",
	}).Debug("Creating stub VR controller adapter (no hardware SDK)")
	return &StubControllerAdapter{}
}

// IsConnected reports false for all hands, indicating no controllers detected.
// This enables the VR controller system to degrade gracefully.
func (s *StubControllerAdapter) IsConnected(hand string) bool {
	return false
}

// GetTrigger returns 0.0 (not pressed) since no hardware is connected.
func (s *StubControllerAdapter) GetTrigger(hand string) float64 {
	return 0.0
}

// GetGrip returns 0.0 (not pressed) since no hardware is connected.
func (s *StubControllerAdapter) GetGrip(hand string) float64 {
	return 0.0
}

// GetThumbstick returns (0, 0) center position since no hardware is connected.
func (s *StubControllerAdapter) GetThumbstick(hand string) (x, y float64) {
	return 0, 0
}

// IsThumbstickPressed returns false since no hardware is connected.
func (s *StubControllerAdapter) IsThumbstickPressed(hand string) bool {
	return false
}

// GetButton returns false for all buttons since no hardware is connected.
func (s *StubControllerAdapter) GetButton(hand, button string) bool {
	return false
}

// SetHaptic is a no-op since no hardware is connected.
// Future SDK implementations will send haptic feedback to the controller.
func (s *StubControllerAdapter) SetHaptic(hand string, intensity, duration float64) {
	// No-op: no hardware to send haptic feedback to
}
