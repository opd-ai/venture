//go:build !js
// +build !js

package mobile

// ShowKeyboard is a no-op on non-WASM platforms where the keyboard is always
// available through the operating system's input system.
//
// This function exists to maintain API compatibility across all platforms.
func ShowKeyboard() {
	// No-op on desktop/native mobile platforms
	// Desktop: Keyboard always available
	// Native mobile: Handled by OS keyboard APIs (not implemented in this build)
}

// HideKeyboard is a no-op on non-WASM platforms.
//
// This function exists to maintain API compatibility across all platforms.
func HideKeyboard() {
	// No-op on desktop/native mobile platforms
}

// IsKeyboardSupported returns false on non-WASM platforms to indicate that
// the JavaScript keyboard bridge is not available. Desktop platforms have
// keyboard support through the OS, but this function specifically refers to
// the WASM keyboard bridge feature.
func IsKeyboardSupported() bool {
	return false // JavaScript keyboard bridge not available on non-WASM platforms
}
