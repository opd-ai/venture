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
// Note: This is a placeholder. Actual implementation requires platform-specific code
// or additional libraries for iOS/Android haptic APIs.
func TriggerHaptic(feedback HapticFeedback) {
	// TODO: Implement platform-specific haptic feedback
	// iOS: Use Core Haptics or UIImpactFeedbackGenerator
	// Android: Use Vibrator service
	// For now, this is a no-op
}
