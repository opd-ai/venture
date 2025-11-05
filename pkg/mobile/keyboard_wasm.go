//go:build js
// +build js

package mobile

import (
	"syscall/js"
)

// keyboardElement holds a reference to the hidden input element used to trigger
// the native mobile keyboard. This element is created lazily on first use.
var keyboardElement js.Value

// initKeyboardElement creates a hidden HTML input element that can trigger
// the native mobile keyboard when focused. This is necessary because canvas
// elements don't automatically trigger the keyboard on mobile browsers.
//
// The element is positioned off-screen but remains accessible to the browser's
// input system. This is a standard technique for canvas-based applications
// that need text input on mobile devices.
func initKeyboardElement() {
	if !keyboardElement.IsUndefined() {
		return // Already initialized
	}

	doc := js.Global().Get("document")
	input := doc.Call("createElement", "input")

	// Set input type to text for general text input
	input.Set("type", "text")
	input.Set("id", "venture-keyboard-input")

	// Style the input to be invisible but functional
	// Position it off-screen but keep it in the DOM so keyboard triggers work
	style := input.Get("style")
	style.Set("position", "absolute")
	style.Set("left", "-9999px")
	style.Set("top", "-9999px")
	style.Set("width", "1px")
	style.Set("height", "1px")
	style.Set("opacity", "0")
	style.Set("pointerEvents", "none")

	// Add to DOM
	body := doc.Get("body")
	body.Call("appendChild", input)

	keyboardElement = input
}

// ShowKeyboard displays the native mobile keyboard by focusing the hidden input element.
// This function should be called when the game enters a text input state (e.g., character
// name entry, server address input).
//
// On mobile browsers, focusing an input element triggers the native keyboard to appear.
// The input element is hidden from view but remains functional.
//
// This is a no-op on desktop browsers where keyboard is always available.
func ShowKeyboard() {
	// Ensure keyboard element exists
	initKeyboardElement()

	// Focus the input element to trigger keyboard
	// Mobile browsers will show the native keyboard when an input is focused
	if !keyboardElement.IsUndefined() {
		keyboardElement.Call("focus")
	}
}

// HideKeyboard dismisses the native mobile keyboard by blurring the hidden input element.
// This function should be called when text input is complete or cancelled.
//
// On mobile browsers, blurring an input element signals that text input is complete
// and the keyboard can be dismissed.
//
// This is a no-op on desktop browsers.
func HideKeyboard() {
	// Blur the input element to dismiss keyboard
	if !keyboardElement.IsUndefined() {
		keyboardElement.Call("blur")

		// Clear the hidden input value (game manages its own text state)
		keyboardElement.Set("value", "")
	}
}

// IsKeyboardSupported returns true on WASM builds where the keyboard bridge is available.
// This can be used to conditionally show keyboard-related UI hints.
func IsKeyboardSupported() bool {
	return true // Always supported in WASM builds
}
