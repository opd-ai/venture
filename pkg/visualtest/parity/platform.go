package parity

import (
	"runtime"
	"strings"
)

// Platform represents a target deployment platform
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformMacOS   Platform = "darwin"
	PlatformWindows Platform = "windows"
	PlatformWASM    Platform = "wasm"
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformUnknown Platform = "unknown"
)

// String returns the human-readable platform name
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

// IsDesktop returns true for desktop platforms
func (p Platform) IsDesktop() bool {
	return p == PlatformLinux || p == PlatformMacOS || p == PlatformWindows
}

// IsMobile returns true for mobile platforms
func (p Platform) IsMobile() bool {
	return p == PlatformIOS || p == PlatformAndroid
}

// IsWeb returns true for web platforms
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
