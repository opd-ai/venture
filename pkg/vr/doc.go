// Package vr provides VR (Virtual Reality) hardware detection and configuration utilities
// for the Venture game engine.
//
// This package enables conditional initialization of VR systems (stereoscopic rendering,
// head tracking, controller input) based on hardware detection or explicit user configuration.
//
// # Hardware Detection
//
// The Detector type performs platform-specific VR hardware detection by checking:
//   - Environment variables for VR runtime paths (STEAMVR_LH_ENABLE, OVR_SDK_PATH, OPENVR_PATH)
//   - Common VR installation directories (SteamVR, Oculus)
//   - Platform restrictions (mobile and WASM always return false)
//
// # Controller Detection Strategy
//
// Controller detection is intentionally conservative: [IsControllerDetected] always returns
// false unless VR mode is force-enabled via [Detector.SetForceEnable]. This design decision
// ensures that VR controller systems are only initialized when a VR headset runtime is
// explicitly detected or enabled. The rationale:
//
//  1. Controllers require a VR runtime (SteamVR, OpenVR, Oculus) to function properly
//  2. Detecting controllers without a headset would result in non-functional VR input
//  3. Users who want VR controller support should have a working headset setup first
//  4. This prevents accidental VR system initialization on machines with generic HID devices
//
// To enable VR systems without physical hardware for testing, use SetForceEnable(true).
//
// # Usage
//
//	detector := vr.NewDetector()
//	if detector.DetectHardware() {
//	    // Initialize VR systems
//	}
//
// # Force Enable/Disable
//
// VR mode can be forced on or off for testing purposes:
//
//	detector.SetForceEnable(true)  // Enable VR without hardware
//	detector.SetForceDisable(true) // Disable VR even with hardware
//
// # CLI Integration
//
// The package integrates with command-line flags:
//   - --vr: Enable VR mode with hardware auto-detection
//   - --force-vr: Force VR mode without hardware (for testing)
//
// # Architecture
//
// VR systems are initialized conditionally in cmd/client/handlers.go:
//   - StereoscopicSystem: Dual-eye stereoscopic rendering
//   - HeadTrackingSystem: Head orientation tracking (with mouse fallback)
//   - VRControllerSystem: VR controller input handling
//   - VRUISystem: VR-optimized UI rendering
//
// Detection results are cached for performance. Use Reset() to clear the cache
// and force re-detection.
package vr
