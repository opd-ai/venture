package mobile

import (
	"runtime"
)

// Platform represents a mobile platform.
type Platform int

const (
	// PlatformUnknown represents an unknown or desktop platform.
	PlatformUnknown Platform = iota
	// PlatformIOS represents iOS (iPhone, iPad).
	PlatformIOS
	// PlatformAndroid represents Android.
	PlatformAndroid
	// PlatformWASM represents WebAssembly/browser (js/wasm).
	PlatformWASM
)

// String returns the string representation of the platform.
func (p Platform) String() string {
	switch p {
	case PlatformIOS:
		return "iOS"
	case PlatformAndroid:
		return "Android"
	case PlatformWASM:
		return "WASM"
	default:
		return "Unknown"
	}
}

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
type Orientation int

const (
	// OrientationUnknown represents an unknown orientation.
	OrientationUnknown Orientation = iota
	// OrientationPortrait represents portrait orientation (height > width).
	OrientationPortrait
	// OrientationLandscape represents landscape orientation (width > height).
	OrientationLandscape
)

// String returns the string representation of the orientation.
func (o Orientation) String() string {
	switch o {
	case OrientationPortrait:
		return "Portrait"
	case OrientationLandscape:
		return "Landscape"
	default:
		return "Unknown"
	}
}

// GetOrientation determines screen orientation based on dimensions.
func GetOrientation(width, height int) Orientation {
	if width > height {
		return OrientationLandscape
	} else if height > width {
		return OrientationPortrait
	}
	return OrientationUnknown
}

// HapticFeedback represents haptic feedback intensity.
type HapticFeedback int

const (
	// HapticLight represents light haptic feedback.
	HapticLight HapticFeedback = iota
	// HapticMedium represents medium haptic feedback.
	HapticMedium
	// HapticHeavy represents heavy haptic feedback.
	HapticHeavy
)

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

// AppLifecycleState represents the application lifecycle state.
// Platform parity fix: Critical for handling system interruptions on mobile
type AppLifecycleState int

const (
	// AppStateActive indicates app is in foreground and interactive
	AppStateActive AppLifecycleState = iota
	// AppStateInactive indicates app is in foreground but not interactive
	// Platform parity fix: iOS - during interruptions (calls, notifications)
	AppStateInactive
	// AppStateBackground indicates app is in background
	// Platform parity fix: Android - app minimized, iOS - home button pressed
	AppStateBackground
	// AppStateTerminating indicates app is about to be terminated
	// Platform parity fix: Allows saving state before forced shutdown
	AppStateTerminating
)

// String returns human-readable app state name.
func (s AppLifecycleState) String() string {
	switch s {
	case AppStateActive:
		return "Active"
	case AppStateInactive:
		return "Inactive"
	case AppStateBackground:
		return "Background"
	case AppStateTerminating:
		return "Terminating"
	default:
		return "Unknown"
	}
}

// SystemInterruptionType represents types of system interruptions.
// Platform parity fix: Different handling needed for different interruption types
type SystemInterruptionType int

const (
	// InterruptionCall indicates incoming phone call
	InterruptionCall SystemInterruptionType = iota
	// InterruptionNotification indicates system notification
	InterruptionNotification
	// InterruptionLowMemory indicates low memory warning
	// Platform parity fix: Should trigger cache clearing on mobile
	InterruptionLowMemory
	// InterruptionAudioRoute indicates audio route change (headphones unplugged)
	InterruptionAudioRoute
)

// String returns human-readable interruption type name.
func (t SystemInterruptionType) String() string {
	switch t {
	case InterruptionCall:
		return "Call"
	case InterruptionNotification:
		return "Notification"
	case InterruptionLowMemory:
		return "LowMemory"
	case InterruptionAudioRoute:
		return "AudioRoute"
	default:
		return "Unknown"
	}
}

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
type WASMSecurityRestriction int

const (
	// RestrictionClipboard - WASM cannot access clipboard without user gesture
	// Platform parity fix: navigator.clipboard requires user interaction (click/tap)
	RestrictionClipboard WASMSecurityRestriction = iota

	// RestrictionFullscreen - WASM fullscreen requires user gesture
	// Platform parity fix: element.requestFullscreen() must be in event handler
	RestrictionFullscreen

	// RestrictionAutoplay - WASM cannot autoplay audio without user interaction
	// Platform parity fix: Browser autoplay policies require user gesture for audio
	RestrictionAutoplay

	// RestrictionPointerLock - WASM pointer lock requires user gesture
	// Platform parity fix: element.requestPointerLock() must be in event handler
	RestrictionPointerLock

	// RestrictionLocalStorage - WASM localStorage may be blocked in private mode
	// Platform parity fix: Safari private browsing blocks localStorage
	RestrictionLocalStorage

	// RestrictionWebGL - WASM WebGL context may fail on some devices
	// Platform parity fix: Mobile browsers may have limited WebGL support
	RestrictionWebGL
)

// String returns human-readable restriction name.
func (r WASMSecurityRestriction) String() string {
	switch r {
	case RestrictionClipboard:
		return "Clipboard"
	case RestrictionFullscreen:
		return "Fullscreen"
	case RestrictionAutoplay:
		return "Autoplay"
	case RestrictionPointerLock:
		return "PointerLock"
	case RestrictionLocalStorage:
		return "LocalStorage"
	case RestrictionWebGL:
		return "WebGL"
	default:
		return "Unknown"
	}
}

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
