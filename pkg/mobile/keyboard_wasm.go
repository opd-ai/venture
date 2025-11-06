//go:build js
// +build js

package mobile

import (
	"syscall/js"
)

// keyboardElement holds a reference to the hidden input element used to trigger
// the native mobile keyboard. This element is created lazily on first use.
var keyboardElement js.Value

// inputEventListener holds the JavaScript callback for input events.
// Note: This is a persistent function that lives for the application's duration.
// We intentionally do not call Release() as it needs to remain active for the
// entire time the game is running. The browser will clean it up when the page unloads.
var inputEventListener js.Func

// keydownEventListener holds the JavaScript callback for keydown events.
// Note: Like inputEventListener, this is persistent for the application's duration.
var keydownEventListener js.Func

// lastInputValue tracks the previous input value to detect changes
var lastInputValue string

// keyCodeMap maps special key names to their keyboard codes for event dispatch.
// Defined at package level to avoid repeated map allocation during event handling.
var keyCodeMap = map[string]int{
	"Enter":  13,
	"Tab":    9,
	"Escape": 27,
}

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

	// CRITICAL: Ensure input can receive and maintain focus
	// Without this, some browsers may dismiss keyboard when input is "invisible"
	input.Set("tabIndex", 0)     // Make focusable
	input.Set("readOnly", false) // Ensure it's editable

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

			// CRITICAL: Refocus input after dispatching events to keep keyboard open
			// Dispatching events can cause the canvas to steal focus
			input.Call("focus")
		} else if len(currentValue) < len(lastInputValue) {
			// Characters deleted (backspace) - dispatch backspace event
			for i := 0; i < len(lastInputValue)-len(currentValue); i++ {
				dispatchBackspaceEvent(doc)
			}

			// CRITICAL: Refocus input after dispatching events to keep keyboard open
			input.Call("focus")
		}

		// Update last value
		lastInputValue = currentValue
		return nil
	})

	// MOBILE KEYBOARD FIX: Forward special keys (Enter, Tab, Escape, etc.)
	// Mobile keyboards generate keydown events for special keys that need forwarding
	keydownEventListener = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
	input.Call("addEventListener", "keydown", keydownEventListener)

	// CRITICAL FIX: Prevent canvas from stealing focus while keyboard is active
	// When events are dispatched to canvas, it can steal focus and hide the keyboard
	// This listener immediately refocuses the input to keep the keyboard visible
	focusGuard := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// If canvas gains focus while input has content, refocus input
		if len(args) > 0 {
			event := args[0]
			target := event.Get("target")

			// Check if a canvas element gained focus
			if target.Get("tagName").String() == "CANVAS" {
				// Check if our input has value (keyboard is in use)
				if !keyboardElement.IsUndefined() && keyboardElement.Get("value").String() != "" {
					// Prevent canvas from taking focus
					event.Call("preventDefault")
					// Refocus input
					keyboardElement.Call("focus")
				}
			}
		}
		return nil
	})

	// Listen for focus events on the entire document
	doc.Call("addEventListener", "focusin", focusGuard, true) // Use capture phase

	// Add to DOM
	body := doc.Get("body")
	body.Call("appendChild", input)

	keyboardElement = input
	lastInputValue = ""
}

// dispatchKeyboardEvent dispatches a synthetic keyboard event to the canvas element
// for a typed character. This allows Ebiten's AppendInputChars to capture
// characters typed on the mobile keyboard.
//
// WASM/Mobile Fix: Mobile keyboards type into the focused input element,
// but Ebiten listens for keyboard events on its canvas element. We bridge this gap
// by manually dispatching events to the canvas.
func dispatchKeyboardEvent(doc js.Value, char string) {
	// Find Ebiten's canvas element (it's the first canvas in the document)
	canvasList := doc.Call("getElementsByTagName", "canvas")
	if canvasList.Get("length").Int() == 0 {
		// Canvas not ready yet, try document as fallback
		dispatchToTarget(doc, char, false)
		return
	}

	canvas := canvasList.Index(0)

	// CRITICAL FIX: Dispatch input event which Ebiten's AppendInputChars actually uses
	// KeyboardEvent alone doesn't populate inpututil.AppendInputChars
	dispatchInputEvent(canvas, char)

	// Also dispatch keyboard events for compatibility with IsKeyPressed
	dispatchToTarget(canvas, char, false)
	dispatchToTarget(doc, char, false)
}

// dispatchInputEvent dispatches an 'input' event which Ebiten uses for text input.
// This is the key to making AppendInputChars work on WASM.
func dispatchInputEvent(target js.Value, char string) {
	eventInit := js.Global().Get("Object").New()
	eventInit.Set("data", char)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", false)
	eventInit.Set("composed", true)

	inputEvent := js.Global().Get("InputEvent").New("input", eventInit)

	// Dispatch event but ensure canvas doesn't steal focus
	// Store current active element
	doc := js.Global().Get("document")
	activeElement := doc.Get("activeElement")

	// Dispatch the event
	target.Call("dispatchEvent", inputEvent)

	// If the active element changed (canvas stole focus), restore it
	if !activeElement.IsUndefined() && activeElement.Get("tagName").String() == "INPUT" {
		newActive := doc.Get("activeElement")
		if !newActive.Equal(activeElement) {
			activeElement.Call("focus")
		}
	}
}

// dispatchToTarget dispatches keyboard events to a specific target
func dispatchToTarget(target js.Value, char string, isSpecial bool) {
	eventInit := js.Global().Get("Object").New()
	eventInit.Set("key", char)
	eventInit.Set("code", "")
	eventInit.Set("keyCode", 0)
	eventInit.Set("which", 0)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", true)

	// Dispatch both keydown and keypress for maximum compatibility
	keydownEvent := js.Global().Get("KeyboardEvent").New("keydown", eventInit)
	target.Call("dispatchEvent", keydownEvent)

	if !isSpecial {
		keypressEvent := js.Global().Get("KeyboardEvent").New("keypress", eventInit)
		target.Call("dispatchEvent", keypressEvent)
	}
}

// dispatchBackspaceEvent dispatches a synthetic backspace keyboard event.
// This handles the case where the user deletes characters on mobile.
//
// WASM/Mobile Fix: When user presses backspace on mobile keyboard,
// we need to forward that to Ebiten's canvas as well.
func dispatchBackspaceEvent(doc js.Value) {
	// Find Ebiten's canvas element
	canvasList := doc.Call("getElementsByTagName", "canvas")

	eventInit := js.Global().Get("Object").New()
	eventInit.Set("key", "Backspace")
	eventInit.Set("code", "Backspace")
	eventInit.Set("keyCode", 8)
	eventInit.Set("which", 8)
	eventInit.Set("bubbles", true)
	eventInit.Set("cancelable", true)

	keydownEvent := js.Global().Get("KeyboardEvent").New("keydown", eventInit)

	// Dispatch to canvas if available
	if canvasList.Get("length").Int() > 0 {
		canvas := canvasList.Index(0)
		canvas.Call("dispatchEvent", keydownEvent)
	}

	// Also dispatch to document for compatibility
	doc.Call("dispatchEvent", keydownEvent)
}

// dispatchSpecialKeyEvent forwards special key events (Enter, Tab, Escape) to canvas.
// These keys are used for navigation and completing text input in the game.
//
// WASM/Mobile Fix: Mobile keyboards generate these events on the focused input,
// but we need to forward them to Ebiten's canvas for game control.
func dispatchSpecialKeyEvent(doc js.Value, key string, originalEvent js.Value) {
	// Use package-level keyCodeMap to avoid repeated allocations
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

	// Find Ebiten's canvas element
	canvasList := doc.Call("getElementsByTagName", "canvas")

	// Dispatch to canvas if available
	if canvasList.Get("length").Int() > 0 {
		canvas := canvasList.Index(0)
		canvas.Call("dispatchEvent", keydownEvent)
	}

	// Also dispatch to document for compatibility
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
