package mobile

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
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

// BUG FIX: Phase 1.2 - Test back button unified navigation API
// Resolution: Verify GetBackButtonKey and IsBackButtonPressed work correctly
// Platform: All platforms

func TestGetBackButtonKey(t *testing.T) {
	key := GetBackButtonKey()

	// All platforms should return ESC key
	// (Android's system back button is mapped to ESC in Ebiten)
	if key != ebiten.KeyEscape {
		t.Errorf("GetBackButtonKey() = %v, want %v", key, ebiten.KeyEscape)
	}

	t.Logf("GetBackButtonKey() = %v (ESC)", key)
}

func TestGetBackButtonName(t *testing.T) {
	name := GetBackButtonName()

	// Check that name is appropriate for platform
	if SupportsBackButton() {
		// Android should show "Back Button"
		if name != "Back Button" {
			t.Errorf("GetBackButtonName() on Android = %q, want %q", name, "Back Button")
		}
		t.Logf("GetBackButtonName() on Android = %q", name)
	} else {
		// Desktop/iOS/WASM should show "ESC"
		if name != "ESC" {
			t.Errorf("GetBackButtonName() on non-Android = %q, want %q", name, "ESC")
		}
		t.Logf("GetBackButtonName() on non-Android = %q", name)
	}
}

func TestIsBackButtonPressed(t *testing.T) {
	// This test verifies the function exists and returns a bool
	// Actual key press testing requires Ebiten runtime
	pressed := IsBackButtonPressed()
	if pressed {
		t.Log("IsBackButtonPressed() = true (key is pressed)")
	} else {
		t.Log("IsBackButtonPressed() = false (key not pressed)")
	}
}

func TestIsBackButtonDown(t *testing.T) {
	// This test verifies the function exists and returns a bool
	// Actual key press testing requires Ebiten runtime
	down := IsBackButtonDown()
	if down {
		t.Log("IsBackButtonDown() = true (key is held down)")
	} else {
		t.Log("IsBackButtonDown() = false (key not held)")
	}
}
