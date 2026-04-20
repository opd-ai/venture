//go:build !js && !(android && cgo && ebitenmobilebind) && !(ios && cgo && ebitenmobilebind)

package mobile

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ShowKeyboard is a no-op on non-WASM platforms where the keyboard is always
// available through the operating system's input system.
//
// This function exists to maintain API compatibility across all platforms.
func ShowKeyboard() {
	// No-op on desktop platforms — the OS keyboard is always available.
	// Native mobile (Android/iOS) keyboard control is implemented in
	// keyboard_android.go and keyboard_ios.go (ebitenmobilebind builds).
}

// HideKeyboard is a no-op on non-WASM platforms.
//
// This function exists to maintain API compatibility across all platforms.
func HideKeyboard() {
	// No-op on desktop/native mobile platforms
}

// IsKeyboardSupported returns false on non-WASM, non-mobile platforms to indicate that
// the JavaScript keyboard bridge is not available. Desktop platforms have
// keyboard support through the OS, but this function specifically refers to
// programmatic keyboard show/hide control.
// On Android (ebitenmobilebind) and iOS (ebitenmobilebind) builds,
// keyboard_android.go and keyboard_ios.go provide platform-specific
// implementations that return true.
func IsKeyboardSupported() bool {
	return false // Programmatic keyboard control not available on desktop/generic builds
}

// BUG FIX: Phase 1.2 - Complete Android back button dual-exit pattern
// Resolution: Add GetBackButtonKey and IsBackButtonPressed for unified navigation
// Platform: Mobile (Android back button, iOS/Desktop ESC key)

// GetBackButtonKey returns the appropriate back navigation key for the platform.
// Platform parity fix: Android back button maps to ESC key in Ebiten
func GetBackButtonKey() ebiten.Key {
	// In Ebiten, Android's system back button is mapped to ebiten.KeyEscape
	// This provides unified handling across all platforms
	return ebiten.KeyEscape
}

// IsBackButtonPressed checks if the platform-appropriate back button was pressed.
// Platform parity fix: Unified back navigation across platforms
// - Android: System back button (mapped to ESC)
// - iOS: No hardware back button (ESC key or swipe gestures)
// - Desktop: ESC key
// - WASM: ESC key or browser back button
func IsBackButtonPressed() bool {
	return inpututil.IsKeyJustPressed(GetBackButtonKey())
}

// IsBackButtonDown checks if the back button is currently held down.
// Platform parity fix: Allows long-press detection on back button
func IsBackButtonDown() bool {
	return ebiten.IsKeyPressed(GetBackButtonKey())
}

// GetBackButtonName returns a human-readable name for the back button.
// Platform parity fix: UI labels show correct button name per platform
func GetBackButtonName() string {
	if SupportsBackButton() {
		return "Back Button" // Android
	}
	return "ESC" // Desktop/iOS/WASM
}
