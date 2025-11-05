//go:build js
// +build js

package mobile

import (
	"syscall/js"
)

// keyboardElement holds a reference to the hidden input element used to trigger
// the native mobile keyboard. This element is created lazily on first use.
var keyboardElement js.Value

// inputEventListener holds the JavaScript callback for input events
var inputEventListener js.Func

// lastInputValue tracks the previous input value to detect changes
var lastInputValue string

// initKeyboardElement creates a hidden HTML input element that can trigger
// the native mobile keyboard when focused. This is necessary because canvas
// elements don't automatically trigger the keyboard on mobile browsers.
//
// The element is positioned off-screen but remains accessible to the browser's
// input system. This is a standard technique for canvas-based applications
// that need text input on mobile devices.
//
// CRITICAL FIX: This function also sets up event forwarding. When the hidden
// input receives keyboard input on mobile, those events must be forwarded to
// the document so that Ebiten's AppendInputChars can capture them. Without
// this forwarding, the mobile keyboard appears but typed text never reaches
// the game.
func initKeyboardElement() {
	if !keyboardElement.IsUndefined() {
		return // Already initialized
	}

	doc := js.Global().Get("document")
	input := doc.Call("createElement", "input")

	// Set input type to text for general text input
	input.Set("type", "text")
	input.Set("id", "venture-keyboard-input")

	// Enable autocomplete and autocorrect for better mobile UX
	input.Set("autocomplete", "off")
	input.Set("autocorrect", "off")
	input.Set("autocapitalize", "off")
	input.Set("spellcheck", false)

	// Set inputmode to optimize mobile keyboard layout
	// "text" mode provides standard keyboard with letters, numbers, symbols
	input.Set("inputmode", "text")

	// Set enterkeyhint to show appropriate Enter button label on mobile
	// "done" shows a "Done" button which is intuitive for completing input
	input.Set("enterkeyhint", "done")

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

	// MOBILE KEYBOARD FIX: Forward input events to document for Ebiten
	// When the mobile keyboard types into our hidden input, we need to dispatch
	// those characters as keyboard events so Ebiten's AppendInputChars picks them up.
	// This is the critical piece that makes mobile text input actually work.
	inputEventListener = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get current input value
		currentValue := input.Get("value").String()

		// Compare with last value to find new characters
		if len(currentValue) > len(lastInputValue) {
			// New characters added - dispatch them to document
			newChars := currentValue[len(lastInputValue):]
			for _, ch := range newChars {
				dispatchKeyboardEvent(doc, string(ch))
			}
		} else if len(currentValue) < len(lastInputValue) {
			// Characters deleted (backspace) - dispatch backspace event
			for i := 0; i < len(lastInputValue)-len(currentValue); i++ {
				dispatchBackspaceEvent(doc)
			}
		}

		// Update last value
		lastInputValue = currentValue
		return nil
	})

	// MOBILE KEYBOARD FIX: Forward special keys (Enter, Tab, Escape, etc.)
	// Mobile keyboards generate keydown events for special keys that need forwarding
	keydownListener := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()

			// Forward special keys that the game uses for navigation/completion
			// Enter: Complete input, Tab: Next field, Escape: Cancel
			if key == "Enter" || key == "Tab" || key == "Escape" {
				dispatchSpecialKeyEvent(doc, key, event)
			}
		}
		return nil
	})

	// Attach event listeners
	input.Call("addEventListener", "input", inputEventListener)
	input.Call("addEventListener", "keydown", keydownListener)

	// Add to DOM
	body := doc.Get("body")
	body.Call("appendChild", input)

	keyboardElement = input
	lastInputValue = ""
}

// dispatchKeyboardEvent dispatches a synthetic keyboard event to the document
// for a typed character. This allows Ebiten's AppendInputChars to capture
// characters typed on the mobile keyboard.
//
// WASM/Mobile Fix: Mobile keyboards type into the focused input element,
// but Ebiten listens for keyboard events on the document. We bridge this gap
// by manually dispatching events.
func dispatchKeyboardEvent(doc js.Value, char string) {
	// Create a KeyboardEvent for the character
	eventInit := js.Global().Get("Object").New()
	eventInit.Set("key", char)
	eventInit.Set("code", "")
	eventInit.Set("keyCode", 0)
	eventInit.Set("which", 0)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", true)

	// Dispatch both keydown and keypress for maximum compatibility
	keydownEvent := js.Global().Get("KeyboardEvent").New("keydown", eventInit)
	doc.Call("dispatchEvent", keydownEvent)

	keypressEvent := js.Global().Get("KeyboardEvent").New("keypress", eventInit)
	doc.Call("dispatchEvent", keypressEvent)
}

// dispatchBackspaceEvent dispatches a synthetic backspace keyboard event.
// This handles the case where the user deletes characters on mobile.
//
// WASM/Mobile Fix: When user presses backspace on mobile keyboard,
// we need to forward that to Ebiten as well.
func dispatchBackspaceEvent(doc js.Value) {
	eventInit := js.Global().Get("Object").New()
	eventInit.Set("key", "Backspace")
	eventInit.Set("code", "Backspace")
	eventInit.Set("keyCode", 8)
	eventInit.Set("which", 8)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", true)

	keydownEvent := js.Global().Get("KeyboardEvent").New("keydown", eventInit)
	doc.Call("dispatchEvent", keydownEvent)
}

// dispatchSpecialKeyEvent forwards special key events (Enter, Tab, Escape) to document.
// These keys are used for navigation and completing text input in the game.
//
// WASM/Mobile Fix: Mobile keyboards generate these events on the focused input,
// but we need to forward them to Ebiten for game control.
func dispatchSpecialKeyEvent(doc js.Value, key string, originalEvent js.Value) {
	// Map key names to keyCodes for compatibility
	keyCodeMap := map[string]int{
		"Enter":  13,
		"Tab":    9,
		"Escape": 27,
	}

	keyCode := keyCodeMap[key]

	eventInit := js.Global().Get("Object").New()
	eventInit.Set("key", key)
	eventInit.Set("code", key)
	eventInit.Set("keyCode", keyCode)
	eventInit.Set("which", keyCode)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", true)

	// Preserve modifier keys from original event
	if !originalEvent.IsUndefined() {
		eventInit.Set("shiftKey", originalEvent.Get("shiftKey"))
		eventInit.Set("ctrlKey", originalEvent.Get("ctrlKey"))
		eventInit.Set("altKey", originalEvent.Get("altKey"))
		eventInit.Set("metaKey", originalEvent.Get("metaKey"))
	}

	keydownEvent := js.Global().Get("KeyboardEvent").New("keydown", eventInit)
	doc.Call("dispatchEvent", keydownEvent)
}

// ShowKeyboard displays the native mobile keyboard by focusing the hidden input element.
// This function should be called when the game enters a text input state (e.g., character
// name entry, server address input).
//
// On mobile browsers, focusing an input element triggers the native keyboard to appear.
// The input element is hidden from view but remains functional.
//
// MOBILE FIX: The hidden input captures keyboard events and forwards them to Ebiten
// via synthetic keyboard events dispatched to the document.
//
// This is a no-op on desktop browsers where keyboard is always available.
func ShowKeyboard() {
	// Ensure keyboard element exists (with event forwarding)
	initKeyboardElement()

	// Clear any previous input value
	if !keyboardElement.IsUndefined() {
		keyboardElement.Set("value", "")
		lastInputValue = ""

		// Focus the input element to trigger keyboard
		// Mobile browsers will show the native keyboard when an input is focused
		keyboardElement.Call("focus")
	}
}

// HideKeyboard dismisses the native mobile keyboard by blurring the hidden input element.
// This function should be called when text input is complete or cancelled.
//
// On mobile browsers, blurring an input element signals that text input is complete
// and the keyboard can be dismissed.
//
// MOBILE FIX: Clears the input value and resets state tracking to ensure clean state
// for the next time keyboard is shown.
//
// This is a no-op on desktop browsers.
func HideKeyboard() {
	// Blur the input element to dismiss keyboard
	if !keyboardElement.IsUndefined() {
		keyboardElement.Call("blur")

		// Clear the hidden input value (game manages its own text state)
		keyboardElement.Set("value", "")
		lastInputValue = ""
	}
}

// IsKeyboardSupported returns true on WASM builds where the keyboard bridge is available.
// This can be used to conditionally show keyboard-related UI hints.
func IsKeyboardSupported() bool {
	return true // Always supported in WASM builds
}
