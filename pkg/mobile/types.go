package mobile

// types.go - Consolidated type definitions for mobile package
// This file contains all enum types and their methods.

// Platform represents the target platform.
// Originally from: platform.go
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

// Orientation represents screen orientation.
// Originally from: platform.go
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

// HapticFeedback represents haptic feedback intensity levels.
// Originally from: platform.go
type HapticFeedback int

const (
	// HapticLight represents light haptic feedback.
	HapticLight HapticFeedback = iota
	// HapticMedium represents medium haptic feedback.
	HapticMedium
	// HapticHeavy represents heavy haptic feedback.
	HapticHeavy
)

// AppLifecycleState represents the application lifecycle state.
// Platform parity fix: Critical for handling system interruptions on mobile
// Originally from: platform.go
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
// Originally from: platform.go
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

// WASMSecurityRestriction represents WASM security constraints.
// Originally from: platform.go
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

// TouchState represents the state of a touch input.
// Originally from: touch.go
type TouchState int

const (
	// TouchStateStarted indicates touch just began this frame
	TouchStateStarted TouchState = iota
	// TouchStateMoved indicates touch is active and moving
	TouchStateMoved
	// TouchStateStationary indicates touch is active but not moving
	TouchStateStationary
	// TouchStateEnded indicates touch ended normally this frame
	TouchStateEnded
	// TouchStateCancelled indicates touch was interrupted (system event, app backgrounded)
	// Platform parity fix: Critical for mobile - handles calls, notifications, app switching
	TouchStateCancelled
)

// String returns human-readable touch state name.
func (ts TouchState) String() string {
	switch ts {
	case TouchStateStarted:
		return "Started"
	case TouchStateMoved:
		return "Moved"
	case TouchStateStationary:
		return "Stationary"
	case TouchStateEnded:
		return "Ended"
	case TouchStateCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// FocusState represents the focus state of the UI.
// Platform parity fix: Track focus/blur states for input filtering
// Originally from: touch.go
type FocusState int

const (
	// FocusStateNormal indicates normal input processing
	FocusStateNormal FocusState = iota
	// FocusStateBlurred indicates UI has lost focus (tab backgrounded on web, app minimized on mobile)
	// Platform parity fix: Critical for preserving state during interruptions
	FocusStateBlurred
	// FocusStateFocused indicates UI has explicit focus (text input active on mobile/web)
	// Platform parity fix: Used to prevent game input when keyboard is active
	FocusStateFocused
)

// String returns human-readable focus state name.
func (fs FocusState) String() string {
	switch fs {
	case FocusStateNormal:
		return "Normal"
	case FocusStateBlurred:
		return "Blurred"
	case FocusStateFocused:
		return "Focused"
	default:
		return "Unknown"
	}
}

// CancelGesture represents different methods to cancel an action.
// Originally from: controls.go
type CancelGesture int

const (
	// GestureTwoFingerTap - Two fingers tap simultaneously (touch)
	// Platform parity fix: Equivalent to Ctrl+Z or right-click cancel
	GestureTwoFingerTap CancelGesture = iota

	// GestureSwipeDown - Quick downward swipe (touch)
	// Platform parity fix: Common mobile pattern for dismiss/close
	GestureSwipeDown

	// GestureEdgeSwipe - Swipe from screen edge (touch)
	// Platform parity fix: iOS/Android back gesture equivalent
	GestureEdgeSwipe

	// GestureEscape - Escape key press (keyboard)
	// Platform parity fix: Standard desktop cancel
	GestureEscape

	// GestureRightClick - Right mouse button (mouse)
	// Platform parity fix: Standard desktop context menu/cancel
	GestureRightClick
)

// String returns human-readable gesture name.
func (g CancelGesture) String() string {
	switch g {
	case GestureTwoFingerTap:
		return "TwoFingerTap"
	case GestureSwipeDown:
		return "SwipeDown"
	case GestureEdgeSwipe:
		return "EdgeSwipe"
	case GestureEscape:
		return "Escape"
	case GestureRightClick:
		return "RightClick"
	default:
		return "Unknown"
	}
}

// JoystickType represents the type of virtual joystick.
// Originally from: dual_joystick.go
type JoystickType int

const (
	JoystickTypeMovement JoystickType = iota // Left joystick for WASD movement
	JoystickTypeAim                          // Right joystick for mouse aim
)
