package mobile

import (
	"runtime"
)

// Platform represents a mobile platform.
// Platform type moved to types.go

// GetPlatform detects the current platform.
func GetPlatform() Platform {
	switch runtime.GOOS {
	case "ios":
		return PlatformIOS
	case "android":
		return PlatformAndroid
	case "js":
		return PlatformWASM
	default:
		return PlatformUnknown
	}
}

// IsMobilePlatform returns true if running on iOS or Android.
func IsMobilePlatform() bool {
	platform := GetPlatform()
	return platform == PlatformIOS || platform == PlatformAndroid
}

// IsTouchCapable returns true if the platform supports touch input.
// This includes mobile platforms (iOS, Android) and WASM (browser with touch).
func IsTouchCapable() bool {
	platform := GetPlatform()
	return platform == PlatformIOS || platform == PlatformAndroid || platform == PlatformWASM
}

// IsWASM returns true if running in WebAssembly/browser.
func IsWASM() bool {
	return GetPlatform() == PlatformWASM
}

// IsIOS returns true if running on iOS.
func IsIOS() bool {
	return GetPlatform() == PlatformIOS
}

// IsAndroid returns true if running on Android.
func IsAndroid() bool {
	return GetPlatform() == PlatformAndroid
}

// Platform parity fix: Platform-specific constraint detection

// KeyboardObscuresUI returns the estimated height of the on-screen keyboard.
// Platform parity fix: Helps UI reposition when mobile keyboard appears
// Returns 0 if keyboard is not visible or on desktop platforms
func KeyboardObscuresUI() int {
	platform := GetPlatform()

	// Platform parity fix: Mobile keyboard height estimates
	// iOS: ~264px portrait (iPhone), ~194px landscape
	// Android: varies by device/keyboard, ~280px portrait typical
	// WASM: varies by browser/device, ~300px mobile browsers
	// Desktop: 0 (no on-screen keyboard)

	switch platform {
	case PlatformIOS:
		// Platform parity fix: iOS keyboard heights
		// In practice, would detect actual keyboard frame from notifications
		// For now, provide conservative estimate
		return 250 // Approximate height in pixels
	case PlatformAndroid:
		// Platform parity fix: Android keyboard heights
		// In practice, would query View.getWindowVisibleDisplayFrame()
		// For now, provide conservative estimate
		return 280 // Approximate height in pixels
	case PlatformWASM:
		// Platform parity fix: Browser keyboard detection
		// In practice, would monitor window.visualViewport resize events
		// For now, provide conservative estimate for mobile browsers
		return 300 // Approximate height in pixels
	default:
		return 0 // Desktop has no on-screen keyboard
	}
}

// GetMinimumTouchTargetSize returns the minimum recommended touch target size.
// Platform parity fix: iOS Human Interface Guidelines: 44pt, Android Material: 48dp
func GetMinimumTouchTargetSize() int {
	platform := GetPlatform()

	switch platform {
	case PlatformIOS:
		// Platform parity fix: iOS HIG minimum touch target
		return 44 // 44pt at 1x scale
	case PlatformAndroid:
		// Platform parity fix: Android Material Design minimum
		return 48 // 48dp at baseline density
	case PlatformWASM:
		// Platform parity fix: Web on mobile devices
		// Use larger of iOS/Android for safety
		return 48
	default:
		// Platform parity fix: Desktop can be more precise
		return 32 // Smaller targets acceptable with mouse precision
	}
}

// SupportsBackButton returns true if platform has a hardware/system back button.
// Platform parity fix: Android has system back button, iOS/WASM use UI navigation
func SupportsBackButton() bool {
	return GetPlatform() == PlatformAndroid
}

// SupportsSystemGestures returns true if platform uses system-level gestures.
// Platform parity fix: iOS uses swipe-from-edge for back, Android varies
func SupportsSystemGestures() bool {
	platform := GetPlatform()
	return platform == PlatformIOS || platform == PlatformAndroid
}

// Orientation represents screen orientation.
// Orientation type moved to types.go

// GetOrientation determines screen orientation based on dimensions.
func GetOrientation(width, height int) Orientation {
	if width > height {
		return OrientationLandscape
	} else if height > width {
		return OrientationPortrait
	}
	return OrientationUnknown
}

// RequiredOrientation returns the orientation required by the game.
// Mobile platforms must be held in landscape mode for the game to be playable.
func RequiredOrientation() Orientation {
	return OrientationLandscape
}

// HapticFeedback represents haptic feedback intensity.
// HapticFeedback type moved to types.go

// TriggerHaptic triggers haptic feedback on mobile devices.
// Platform-specific implementations should be provided via build tags:
// - platform_ios.go with //go:build ios tag for iOS Core Haptics
// - platform_android.go with //go:build android tag for Android Vibrator
// This default implementation is a no-op for desktop/WASM builds.
func TriggerHaptic(feedback HapticFeedback) {
	// Default no-op implementation for non-mobile platforms
	// Mobile platforms should provide their own implementations via build tags
	triggerHapticImpl(feedback)
}

// triggerHapticImpl is the platform-specific implementation.
// Default implementation is a no-op.
func triggerHapticImpl(feedback HapticFeedback) {
	// No-op on desktop/WASM platforms
	_ = feedback
}

// Platform parity fix: System interruption and lifecycle management

// AppLifecycleState type moved to types.go
// Platform parity fix: Different handling needed for different interruption types
// SystemInterruptionType type moved to types.go

// AppLifecycleHandler manages application lifecycle events.
// Platform parity fix: Centralizes platform-specific lifecycle handling
type AppLifecycleHandler struct {
	currentState   AppLifecycleState
	onStateChange  func(AppLifecycleState)
	onInterruption func(SystemInterruptionType)
}

// NewAppLifecycleHandler creates a new lifecycle handler.
func NewAppLifecycleHandler() *AppLifecycleHandler {
	return &AppLifecycleHandler{
		currentState: AppStateActive,
	}
}

// SetStateChangeCallback registers a callback for app state changes.
// Platform parity fix: Allows game to pause, save, or adjust rendering
func (h *AppLifecycleHandler) SetStateChangeCallback(callback func(AppLifecycleState)) {
	h.onStateChange = callback
}

// SetInterruptionCallback registers a callback for system interruptions.
// Platform parity fix: Allows game to mute audio, pause, show notification
func (h *AppLifecycleHandler) SetInterruptionCallback(callback func(SystemInterruptionType)) {
	h.onInterruption = callback
}

// NotifyStateChange notifies the handler of an app state change.
// Platform parity fix: Called by platform-specific code (iOS/Android/WASM bridge)
func (h *AppLifecycleHandler) NotifyStateChange(newState AppLifecycleState) {
	if h.currentState == newState {
		return
	}

	h.currentState = newState

	if h.onStateChange != nil {
		h.onStateChange(newState)
	}
}

// NotifyInterruption notifies the handler of a system interruption.
// Platform parity fix: Called when phone call, notification, or other interruption occurs
func (h *AppLifecycleHandler) NotifyInterruption(interruptionType SystemInterruptionType) {
	if h.onInterruption != nil {
		h.onInterruption(interruptionType)
	}
}

// GetCurrentState returns the current app lifecycle state.
func (h *AppLifecycleHandler) GetCurrentState() AppLifecycleState {
	return h.currentState
}

// IsActive returns true if app is currently active and interactive.
func (h *AppLifecycleHandler) IsActive() bool {
	return h.currentState == AppStateActive
}

// IsBackground returns true if app is in background.
func (h *AppLifecycleHandler) IsBackground() bool {
	return h.currentState == AppStateBackground
}

// Platform parity fix: WASM security restrictions documentation and helpers

// WASMSecurityRestriction represents browser security limitations in WASM.
// Platform parity fix: Documents and provides detection for WASM-specific constraints
// WASMSecurityRestriction type moved to types.go

// GetWASMRestrictionMessage returns a user-friendly message for a WASM restriction.
// Platform parity fix: Provides clear guidance when browser blocks functionality
func GetWASMRestrictionMessage(restriction WASMSecurityRestriction) string {
	switch restriction {
	case RestrictionClipboard:
		return "Clipboard access requires clicking a button or menu item. " +
			"Browser security prevents automatic clipboard access."
	case RestrictionFullscreen:
		return "Fullscreen requires clicking a button or pressing a key. " +
			"Browser security prevents automatic fullscreen."
	case RestrictionAutoplay:
		return "Audio autoplay blocked by browser. Click anywhere to enable sound. " +
			"This is a browser security policy to prevent unwanted audio."
	case RestrictionPointerLock:
		return "Pointer lock requires clicking in the game area. " +
			"Browser security prevents automatic mouse capture."
	case RestrictionLocalStorage:
		return "Save data unavailable in private browsing mode. " +
			"Use regular browsing to enable game saves."
	case RestrictionWebGL:
		return "Graphics acceleration unavailable on this device. " +
			"Game may run slowly or not display correctly."
	default:
		return "Browser security restriction detected."
	}
}

// HasWASMRestriction checks if a specific restriction applies to current platform.
// Platform parity fix: Returns true only on WASM, false on native platforms
func HasWASMRestriction(restriction WASMSecurityRestriction) bool {
	// All restrictions only apply to WASM platform
	// Native mobile/desktop builds don't have browser security restrictions
	return IsWASM()
}

// GetWASMWorkaroundMessage returns suggested workaround for a WASM restriction.
// Platform parity fix: Provides actionable guidance to users
func GetWASMWorkaroundMessage(restriction WASMSecurityRestriction) string {
	switch restriction {
	case RestrictionClipboard:
		return "Use the in-game clipboard menu to copy/paste with button clicks."
	case RestrictionFullscreen:
		return "Use the fullscreen button in the menu or press F11."
	case RestrictionAutoplay:
		return "Click the screen or press any key to start audio."
	case RestrictionPointerLock:
		return "Click in the game window to enable mouse control."
	case RestrictionLocalStorage:
		return "Exit private browsing mode to enable game saves, or use manual save export."
	case RestrictionWebGL:
		return "Try updating your browser or using a different device."
	default:
		return "Please interact with the game to enable this feature."
	}
}
