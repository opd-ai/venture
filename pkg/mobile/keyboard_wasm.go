//go:build js
// +build js

package mobile

import (
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Console logging helper for debugging keyboard issues
var console = js.Global().Get("console")

// logInfo logs an informational message to browser console
func logInfo(msg string) {
	console.Call("log", "[VentureKeyboard] "+msg)
}

// logError logs an error message to browser console
func logError(msg string) {
	console.Call("error", "[VentureKeyboard] "+msg)
}

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

// initializationAttempted tracks whether we've tried to initialize the keyboard
// to avoid excessive retry attempts
var initializationAttempted bool

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
		logInfo("Keyboard element already initialized, skipping")
		return
	}

	markInitializationAttempt()

	doc := js.Global().Get("document")
	if !validateDOM(doc) {
		return
	}

	input := createKeyboardInput(doc)
	setupInputElement(input)
	attachEventListeners(input, doc)
	setupFocusGuard(doc)

	if !appendInputToBody(doc, input) {
		return
	}

	keyboardElement = input
	lastInputValue = ""
	logInfo("Virtual keyboard element created and added to DOM")
	logInfo("Element ID: venture-keyboard-input, Type: text, InputMode: text")
	logInfo("Canvas element detected - keyboard ready for use")
}

// markInitializationAttempt marks that initialization has been attempted.
func markInitializationAttempt() {
	if !initializationAttempted {
		initializationAttempted = true
		logInfo("First keyboard initialization attempt")
	} else {
		logInfo("Retrying keyboard initialization")
	}
	logInfo("Initializing virtual keyboard element")
}

// validateDOM verifies document and body are available.
func validateDOM(doc js.Value) bool {
	if doc.IsUndefined() || doc.IsNull() {
		logError("Document is undefined or null - DOM not ready")
		logInfo("Initialization will be retried on next ShowKeyboard() call")
		return false
	}
	return true
}

// createKeyboardInput creates the input element with basic attributes.
func createKeyboardInput(doc js.Value) js.Value {
	input := doc.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("id", "venture-keyboard-input")
	input.Set("autocomplete", "off")
	input.Set("autocorrect", "off")
	input.Set("autocapitalize", "off")
	input.Set("spellcheck", false)
	input.Set("inputmode", "text")
	input.Set("enterkeyhint", "done")
	return input
}

// setupInputElement applies styling and properties to the input element.
func setupInputElement(input js.Value) {
	style := input.Get("style")
	style.Set("position", "fixed")
	style.Set("left", "-9999px")
	style.Set("top", "-9999px")
	style.Set("width", "200px")
	style.Set("height", "50px")
	style.Set("opacity", "0.01")
	style.Set("zIndex", "999")
	style.Set("border", "none")
	style.Set("background", "transparent")
	style.Set("color", "transparent")
	style.Set("fontSize", "16px")
	style.Set("outline", "none")

	input.Set("tabIndex", 0)
	input.Set("readOnly", false)
}

// attachEventListeners attaches input and keydown event listeners.
func attachEventListeners(input, doc js.Value) {
	inputEventListener = createInputListener(input, doc)
	keydownEventListener = createKeydownListener(doc)

	input.Call("addEventListener", "input", inputEventListener)
	input.Call("addEventListener", "keydown", keydownEventListener)
}

// createInputListener creates the input event listener function.
func createInputListener(input, doc js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		currentValue := input.Get("value").String()

		if len(currentValue) > len(lastInputValue) {
			newChars := currentValue[len(lastInputValue):]
			logInfo("Input event: new chars added: '" + newChars + "'")
			for _, ch := range newChars {
				dispatchKeyboardEvent(doc, string(ch))
			}
			input.Call("focus")
		} else if len(currentValue) < len(lastInputValue) {
			deletedCount := len(lastInputValue) - len(currentValue)
			logInfo("Input event: backspace pressed (" + string(rune('0'+deletedCount)) + " chars deleted)")
			for i := 0; i < deletedCount; i++ {
				dispatchBackspaceEvent(doc)
			}
			input.Call("focus")
		}

		lastInputValue = currentValue
		return nil
	})
}

// createKeydownListener creates the keydown event listener function.
func createKeydownListener(doc js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()

			if key == "Enter" || key == "Tab" || key == "Escape" {
				dispatchSpecialKeyEvent(doc, key, event)
			}
		}
		return nil
	})
}

// setupFocusGuard prevents canvas from stealing focus while keyboard is active.
func setupFocusGuard(doc js.Value) {
	focusGuard := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			target := event.Get("target")

			if target.Get("tagName").String() == "CANVAS" {
				if !keyboardElement.IsUndefined() && keyboardElement.Get("value").String() != "" {
					event.Call("preventDefault")
					keyboardElement.Call("focus")
				}
			}
		}
		return nil
	})

	doc.Call("addEventListener", "focusin", focusGuard, true)
}

// appendInputToBody adds the input element to the document body after validation.
func appendInputToBody(doc, input js.Value) bool {
	body := doc.Get("body")
	if body.IsUndefined() || body.IsNull() {
		logError("Document body is undefined or null - DOM not ready")
		logInfo("This may occur if ShowKeyboard() is called too early")
		logInfo("Initialization will be retried on next ShowKeyboard() call")
		return false
	}

	canvasList := doc.Call("getElementsByTagName", "canvas")
	if canvasList.Get("length").Int() == 0 {
		logError("No canvas element found - Ebiten not fully initialized")
		logInfo("Waiting for Ebiten to create canvas element")
		logInfo("Initialization will be retried on next ShowKeyboard() call")
		return false
	}

	canvas := canvasList.Index(0)
	canvasStyle := canvas.Get("style")
	canvasStyle.Set("position", "relative")
	canvasStyle.Set("zIndex", "1")
	logInfo("Canvas z-index set to 1 (input is 999)")

	body.Call("appendChild", input)
	return true
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

// ShowKeyboard displays the native mobile keyboard by focusing the hidden input element
// and moving it to a tappable on-screen position.
//
// MOBILE FIX: The input element is moved from off-screen to on-screen (bottom-center)
// so users can tap it to trigger the keyboard. Many mobile browsers require a user
// gesture (touch) to show the keyboard, not just programmatic focus().
//
// The input is nearly invisible (opacity 0.01) but positioned where users naturally
// tap during text input screens. Programmatic focus() is also attempted, but the
// on-screen position provides a fallback tap target.
//
// This function should be called when the game enters a text input state (e.g., character
// name entry, server address input).
func ShowKeyboard() {
	logInfo("ShowKeyboard() called")

	// Ensure keyboard element exists (with event forwarding)
	initKeyboardElement()

	// Clear any previous input value
	if !keyboardElement.IsUndefined() {
		keyboardElement.Set("value", "")
		lastInputValue = ""

		// CRITICAL FIX: Move input ON-SCREEN to a tappable position
		// Position at bottom-center where users naturally tap for text input
		// This is required because mobile browsers may not show keyboard for
		// programmatic focus() alone - they often require a user gesture (touch)
		style := keyboardElement.Get("style")
		style.Set("left", "50%")
		style.Set("transform", "translateX(-50%)") // Center horizontally
		style.Set("bottom", "80px")                // Above mobile keyboard area
		style.Set("top", "auto")                   // Clear the off-screen top value

		// MOBILE FIX: Make input slightly more visible when active to help users find it
		// Increase opacity from 0.01 to 0.05 when keyboard is requested
		// Still mostly invisible but easier to tap
		style.Set("opacity", "0.05")

		// MOBILE FIX: Use requestAnimationFrame to ensure style changes are processed
		// before calling focus(). This gives the browser a chance to reflow/repaint.
		// On some mobile browsers, immediate focus() after style changes may fail.
		requestAnimationFrame := js.Global().Get("requestAnimationFrame")

		// Create callback that will be called once by requestAnimationFrame
		var focusCallback js.Func
		focusCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// CRITICAL: Release callback after execution to prevent memory leak
			// This is safe because requestAnimationFrame calls the callback exactly once
			defer focusCallback.Release()

			// Focus the input element to trigger keyboard
			// Mobile browsers will show the native keyboard when an input is focused
			// (especially if this focus happens during/after a user touch event)
			keyboardElement.Call("focus")

			logInfo("Keyboard element moved on-screen and focused")

			// Verify focus was successful
			doc := js.Global().Get("document")
			activeElement := doc.Get("activeElement")
			if activeElement.Get("id").String() == "venture-keyboard-input" {
				logInfo("Focus successful - active element is venture-keyboard-input")
			} else {
				logError("Focus failed - active element is: " + activeElement.Get("tagName").String())
				logInfo("User may need to tap the screen to trigger keyboard")
				logInfo("Input position: bottom-center, opacity: 0.05, size: 200x50px")
			}

			return nil
		})

		// Schedule focus for next animation frame
		requestAnimationFrame.Call("call", js.Global(), focusCallback)
	} else {
		logError("Keyboard element is undefined - initialization failed")
	}
}

// HideKeyboard dismisses the native mobile keyboard by blurring the hidden input element
// and moving it back off-screen.
//
// MOBILE FIX: The input element is moved back off-screen after blurring to ensure it
// doesn't intercept touch events meant for game controls. The input value is cleared
// to reset state for the next keyboard session.
//
// This function should be called when text input is complete or cancelled.
func HideKeyboard() {
	logInfo("HideKeyboard() called")

	// Blur the input element to dismiss keyboard
	if !keyboardElement.IsUndefined() {
		keyboardElement.Call("blur")

		// CRITICAL FIX: Move input back OFF-SCREEN
		// This ensures it doesn't intercept touch events during gameplay
		style := keyboardElement.Get("style")
		style.Set("left", "-9999px")
		style.Set("top", "-9999px")
		style.Set("bottom", "auto")    // Clear bottom positioning
		style.Set("transform", "none") // Clear transform
		style.Set("opacity", "0.01")   // Reset opacity to nearly invisible

		// Clear the hidden input value (game manages its own text state)
		keyboardElement.Set("value", "")
		lastInputValue = ""

		logInfo("Keyboard element blurred, cleared, and moved off-screen")
	} else {
		logInfo("HideKeyboard() called but keyboard element not initialized")
	}
}

// IsKeyboardSupported returns true on WASM builds where the keyboard bridge is available.
// This can be used to conditionally show keyboard-related UI hints.
func IsKeyboardSupported() bool {
	return true // Always supported in WASM builds
}

// BUG FIX: Phase 1.2 - Complete Android back button dual-exit pattern
// Resolution: Add GetBackButtonKey and IsBackButtonPressed for unified navigation
// Platform: WASM (browser back button, ESC key)

// GetBackButtonKey returns the appropriate back navigation key for the platform.
// Platform parity fix: WASM uses ESC key (browser back is not accessible)
func GetBackButtonKey() ebiten.Key {
	return ebiten.KeyEscape
}

// IsBackButtonPressed checks if the platform-appropriate back button was pressed.
// Platform parity fix: Unified back navigation across platforms
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
	return "ESC" // WASM uses ESC key
}
