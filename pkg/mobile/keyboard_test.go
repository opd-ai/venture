package mobile

import (
	"testing"
)

// TestShowHideKeyboard tests that keyboard functions don't panic
// Note: These are mainly smoke tests since actual keyboard behavior requires a browser
func TestShowHideKeyboard(t *testing.T) {
	// Should not panic on non-WASM builds (no-op functions)
	ShowKeyboard()
	HideKeyboard()

	// Test can be called multiple times
	ShowKeyboard()
	ShowKeyboard()
	HideKeyboard()
	HideKeyboard()
}

// TestIsKeyboardSupported tests the keyboard support detection
func TestIsKeyboardSupported(t *testing.T) {
	// On WASM builds, this should return true
	// On other builds, this should return false
	// We can't test the actual value here since it depends on build tags
	// But we can ensure the function doesn't panic
	_ = IsKeyboardSupported()
}
