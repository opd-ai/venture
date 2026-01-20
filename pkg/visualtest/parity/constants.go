// Package parity provides cross-platform visual parity validation.
package parity

// Platform represents a target deployment platform
// Originally from: platform.go
type Platform string

const (
	// PlatformLinux represents Linux desktop platform
	PlatformLinux Platform = "linux"
	// PlatformMacOS represents macOS desktop platform
	PlatformMacOS Platform = "darwin"
	// PlatformWindows represents Windows desktop platform
	PlatformWindows Platform = "windows"
	// PlatformWASM represents WebAssembly web platform
	PlatformWASM Platform = "wasm"
	// PlatformIOS represents iOS mobile platform
	PlatformIOS Platform = "ios"
	// PlatformAndroid represents Android mobile platform
	PlatformAndroid Platform = "android"
	// PlatformUnknown represents an unknown or unsupported platform
	PlatformUnknown Platform = "unknown"
)
