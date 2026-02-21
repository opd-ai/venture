// Code relocated from: original platform.go (Platform type and constants moved to constants.go)
package parity

import (
	"runtime"
	"strings"
)

// String returns the human-readable platform name for this Platform constant.
func (p Platform) String() string {
	switch p {
	case PlatformLinux:
		return "Linux"
	case PlatformMacOS:
		return "macOS"
	case PlatformWindows:
		return "Windows"
	case PlatformWASM:
		return "WebAssembly"
	case PlatformIOS:
		return "iOS"
	case PlatformAndroid:
		return "Android"
	default:
		return "Unknown"
	}
}

// IsDesktop returns true if this Platform represents a desktop operating system
// (Linux, macOS, or Windows).
func (p Platform) IsDesktop() bool {
	return p == PlatformLinux || p == PlatformMacOS || p == PlatformWindows
}

// IsMobile returns true if this Platform represents a mobile operating system
// (iOS or Android).
func (p Platform) IsMobile() bool {
	return p == PlatformIOS || p == PlatformAndroid
}

// IsWeb returns true if this Platform represents a web/browser environment
// (WebAssembly).
func (p Platform) IsWeb() bool {
	return p == PlatformWASM
}

// DetectPlatform returns the current platform based on GOOS
func DetectPlatform() Platform {
	goos := strings.ToLower(runtime.GOOS)
	goarch := strings.ToLower(runtime.GOARCH)

	// WebAssembly detection
	if goos == "js" && goarch == "wasm" {
		return PlatformWASM
	}

	// Mobile platform detection
	if goos == "android" {
		return PlatformAndroid
	}
	if goos == "ios" {
		return PlatformIOS
	}

	// Desktop platform detection
	switch goos {
	case "linux":
		return PlatformLinux
	case "darwin":
		return PlatformMacOS
	case "windows":
		return PlatformWindows
	default:
		return PlatformUnknown
	}
}

// PlatformInfo contains detailed information about the current platform
type PlatformInfo struct {
	Platform           Platform
	GOOS               string
	GOARCH             string
	NumCPU             int
	SupportsTouch      bool
	SupportsFullscreen bool
	SupportsWebGL      bool
}

// GetPlatformInfo returns detailed information about the current platform
func GetPlatformInfo() PlatformInfo {
	platform := DetectPlatform()
	return PlatformInfo{
		Platform:           platform,
		GOOS:               runtime.GOOS,
		GOARCH:             runtime.GOARCH,
		NumCPU:             runtime.NumCPU(),
		SupportsTouch:      platform.IsMobile(),
		SupportsFullscreen: platform.IsDesktop(),
		SupportsWebGL:      platform.IsWeb(),
	}
}
