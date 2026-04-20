//go:build vr

// Package engine provides the conditional-compilation framework for real VR hardware
// SDK integration. This file is compiled only when building with -tags vr and
// establishes the adapter types and cgo integration points for OpenXR.
//
// # OpenXR SDK Integration Research (ROADMAP item 1)
//
// The cross-platform standard for VR hardware integration in Go via cgo is OpenXR 1.x,
// supported by all major headset vendors (SteamVR, Oculus/Meta, Pico, WMR).
//
// Recommended Go binding approach (no mature native Go binding exists as of 2026):
//
//  1. Use cgo directly with the Khronos OpenXR Loader:
//     - Install: https://github.com/KhronosGroup/OpenXR-SDK-Source
//     - Header:  openxr/openxr.h (from khronos loader or distro package)
//     - Linker:  -lopenxr_loader (or platform-specific path)
//     - Linux:   sudo apt install libopenxr-loader1 libopenxr-dev
//     - macOS:   brew install openxr (SteamVR is the only runtime; prefer OpenVR)
//     - Windows: download from https://www.khronos.org/openxr/
//
//  2. Alternative: Use OpenVR (SteamVR-specific) for desktop:
//     - Repository: https://github.com/ValveSoftware/openvr
//     - Go wrapper example: https://github.com/mewpkg/openvr (unmaintained)
//     - Header: openvr.h, Link: -lopenvr_api
//
//  3. WebXR for WASM builds:
//     - Use syscall/js to call navigator.xr APIs (WebXR Device API W3C spec)
//     - Entry point: navigator.xr.requestSession("immersive-vr")
//     - No additional Go dependencies needed; integrated via JS bridge
//     - Reference: https://developer.mozilla.org/en-US/docs/Web/API/WebXR_Device_API
//
// # Implementation Path
//
// To complete OpenXR integration, replace the commented cgo sections below with
// the actual header includes and implement each method using the OpenXR C API.
// The interface contracts (VRHeadsetAdapter, VRControllerAdapter) are already
// defined in pkg/engine/interfaces.go and must be satisfied.
//
// Build this file with:
//
//	go build -tags vr ./...
//	go test -tags vr ./pkg/engine/...
package engine

import (
	log "github.com/sirupsen/logrus"
)

// OpenXRHeadsetAdapter implements VRHeadsetAdapter using the OpenXR 1.x API.
//
// # Cgo Integration (TODO: uncomment when OpenXR SDK is available)
//
//	/*
//	#cgo linux LDFLAGS: -lopenxr_loader
//	#cgo windows LDFLAGS: -lopenxr_loader
//	#include <openxr/openxr.h>
//	#include <stdlib.h>
//	...
//	*/
//	import "C"
//
// The cgo section above requires the OpenXR loader library to be installed.
// On Ubuntu/Debian: sudo apt install libopenxr-loader1 libopenxr-dev
// On Fedora/RHEL: sudo dnf install openxr-loader openxr-loader-devel
// On Windows: download the Khronos OpenXR SDK and set CGO_LDFLAGS.
//
// Implements: VRHeadsetAdapter
type OpenXRHeadsetAdapter struct {
	// sessionHandle holds the XrSession handle once the OpenXR session is created.
	// Field name matches the OpenXR spec type XrSession (uint64 handle).
	sessionHandle uint64 // XrSession handle (set after xrCreateSession)

	// systemID holds the XrSystemId for the HMD system selected at startup.
	systemID uint64 // XrSystemId (set after xrGetSystem)

	// connected tracks whether the OpenXR runtime and headset are available.
	connected bool

	// ipd is the interpupillary distance in millimeters, read from the runtime.
	ipd float64
}

// NewOpenXRHeadsetAdapter creates an OpenXR headset adapter and initialises the
// OpenXR runtime. Returns the adapter regardless of whether hardware is detected;
// call IsConnected() to determine availability.
func NewOpenXRHeadsetAdapter() *OpenXRHeadsetAdapter {
	a := &OpenXRHeadsetAdapter{ipd: 63.0} // Default IPD (mm) until runtime provides it

	// TODO(vr-sdk): Call xrCreateInstance, xrGetSystem, xrCreateSession here.
	// Example flow:
	//   xrCreateInstance(&instanceInfo, &instance)
	//   xrGetSystem(instance, &systemGetInfo, &systemID)
	//   xrCreateSession(instance, &sessionInfo, &session)
	//   Read XrViewConfigurationView[0].recommendedSwapchainSampleCount for IPD

	log.WithFields(log.Fields{
		"adapter": "openxr_headset",
		"status":  "sdk_not_integrated",
	}).Warn("OpenXR headset adapter created; SDK integration pending — no hardware connected")

	return a
}

// IsConnected returns true when the OpenXR runtime has confirmed a headset session.
// Returns false until the SDK is fully integrated.
func (a *OpenXRHeadsetAdapter) IsConnected() bool {
	return a.connected
}

// GetHeadOrientation returns head pose orientation (pitch, yaw, roll) in radians.
//
// TODO(vr-sdk): Call xrLocateViews and extract XrPosef.orientation as XrQuaternionf,
// then convert quaternion to Euler angles (pitch/yaw/roll).
func (a *OpenXRHeadsetAdapter) GetHeadOrientation() (pitch, yaw, roll float64) {
	// Placeholder: returns zero until SDK integration provides XrPosef
	return 0, 0, 0
}

// GetHeadPosition returns head position (x, y, z) in metres relative to the play origin.
//
// TODO(vr-sdk): Call xrLocateViews and extract XrPosef.position as XrVector3f.
func (a *OpenXRHeadsetAdapter) GetHeadPosition() (x, y, z float64) {
	// Placeholder: returns origin until SDK integration provides XrPosef
	return 0, 0, 0
}

// GetIPD returns the interpupillary distance in millimetres.
//
// TODO(vr-sdk): Read from XrViewConfigurationView recommended values returned by
// xrEnumerateViewConfigurationViews for XR_VIEW_CONFIGURATION_TYPE_PRIMARY_STEREO.
func (a *OpenXRHeadsetAdapter) GetIPD() float64 {
	return a.ipd
}

// OpenXRControllerAdapter implements VRControllerAdapter using the OpenXR 1.x API.
//
// OpenXR controller input uses the interaction profile system:
//   - Bind action paths (e.g. /user/hand/left/input/trigger/value)
//   - Call xrSyncActions each frame
//   - Read state via xrGetActionStateFloat, xrGetActionStateBoolean, etc.
//
// TODO(vr-sdk): Implement the full action-set and binding setup in NewOpenXRControllerAdapter.
//
// Implements: VRControllerAdapter
type OpenXRControllerAdapter struct {
	// connected tracks whether the OpenXR runtime reports controllers active.
	connected bool
}

// NewOpenXRControllerAdapter creates an OpenXR controller adapter.
// Returns the adapter regardless of hardware availability; call IsConnected to check.
func NewOpenXRControllerAdapter() *OpenXRControllerAdapter {
	a := &OpenXRControllerAdapter{}

	// TODO(vr-sdk): Create XrActionSet and XrAction objects for:
	//   trigger, grip, thumbstick (x/y/click), buttons (A/B/X/Y)
	// Then call xrSuggestInteractionProfileBindings for common profiles:
	//   KHR/simple, Valve Index, Oculus Touch, Windows Mixed Reality

	log.WithFields(log.Fields{
		"adapter": "openxr_controller",
		"status":  "sdk_not_integrated",
	}).Warn("OpenXR controller adapter created; SDK integration pending — no hardware connected")

	return a
}

// IsConnected returns true when the OpenXR runtime confirms controllers are active
// for the given hand ("left" or "right").
//
// TODO(vr-sdk): Query XR_TYPE_INTERACTION_PROFILE_STATE via xrGetCurrentInteractionProfile.
func (a *OpenXRControllerAdapter) IsConnected(hand string) bool {
	return a.connected
}

// GetTrigger returns the trigger axis value [0, 1] for the given hand.
//
// TODO(vr-sdk): Call xrGetActionStateFloat for the trigger action bound to
// /user/hand/{hand}/input/trigger/value.
func (a *OpenXRControllerAdapter) GetTrigger(hand string) float64 {
	return 0
}

// GetGrip returns the grip axis value [0, 1] for the given hand.
//
// TODO(vr-sdk): Call xrGetActionStateFloat for the grip/squeeze action bound to
// /user/hand/{hand}/input/squeeze/value.
func (a *OpenXRControllerAdapter) GetGrip(hand string) float64 {
	return 0
}

// GetThumbstick returns the thumbstick position [-1, 1] × [-1, 1] for the given hand.
//
// TODO(vr-sdk): Call xrGetActionStateVector2f for the thumbstick action bound to
// /user/hand/{hand}/input/thumbstick.
func (a *OpenXRControllerAdapter) GetThumbstick(hand string) (x, y float64) {
	return 0, 0
}

// IsThumbstickPressed returns whether the thumbstick click is active.
//
// TODO(vr-sdk): Call xrGetActionStateBoolean for /user/hand/{hand}/input/thumbstick/click.
func (a *OpenXRControllerAdapter) IsThumbstickPressed(hand string) bool {
	return false
}

// GetButton returns whether the named face button is pressed.
// Common button names: "a", "b", "x", "y", "menu", "system".
//
// TODO(vr-sdk): Map button name to the corresponding XrAction and call
// xrGetActionStateBoolean.
func (a *OpenXRControllerAdapter) GetButton(hand, button string) bool {
	return false
}

// SetHaptic triggers haptic feedback on the given controller.
// intensity is [0, 1] and duration is in seconds.
//
// TODO(vr-sdk): Call xrApplyHapticFeedback with XrHapticVibration:
//   amplitude = intensity, duration (ns), frequency (Hz, 0 = default).
func (a *OpenXRControllerAdapter) SetHaptic(hand string, intensity, duration float64) {
	// No-op: OpenXR SDK not yet integrated.
	// TODO(vr-sdk): Call xrApplyHapticFeedback with XrHapticVibration:
	//   amplitude = intensity, duration (ns), frequency (Hz, 0 = default).
}
