# Touch Input Support for WebAssembly/Browser Build

This document describes the touch input implementation for the Venture WASM build.

## Overview

Touch input is fully supported in the WebAssembly/browser build, allowing the game to be played on:
- Mobile browsers (iOS Safari, Chrome, Firefox)
- Touch-capable laptops and tablets
- Devices with stylus input

The implementation leverages Ebiten's cross-platform touch APIs and the existing mobile touch infrastructure.

## Architecture

### Platform Detection

The `pkg/mobile/platform.go` module detects three touch-capable platforms:

- **iOS** (`GOOS=ios`): Native iOS devices
- **Android** (`GOOS=android`): Native Android devices  
- **WASM** (`GOOS=js`): WebAssembly/browser environment

Key functions:
```go
IsTouchCapable() bool    // Returns true for iOS, Android, WASM
IsMobilePlatform() bool  // Returns true ONLY for iOS, Android
IsWASM() bool            // Returns true for WASM
```

This distinction allows:
- Touch input to work on all platforms
- Virtual controls to appear only on true mobile platforms
- WASM to use keyboard/mouse OR touch based on available input

### Input System Integration

The `InputSystem` (`pkg/engine/input_system.go`) handles touch input:

1. **Initialization**: Sets `useTouchInput = true` for touch-capable platforms
2. **Auto-detection**: Detects touch events at runtime and switches input modes
3. **Processing**: Reads touch coordinates via `ebiten.TouchIDs()` and `ebiten.TouchPosition()`
4. **Gestures**: Recognizes tap, swipe, pinch, long-press, and double-tap

### Touch Handler

The `TouchInputHandler` (`pkg/mobile/touch.go`) processes raw touch events:

- Tracks multiple simultaneous touches
- Maintains touch lifecycle (start, move, end)
- Feeds data to `GestureDetector` for pattern recognition

### Gesture Detection

The `GestureDetector` recognizes common touch patterns:

| Gesture | Detection Criteria | Use Case |
|---------|-------------------|----------|
| **Tap** | Quick touch/release, <20px movement | Attack, interact, select |
| **Double Tap** | Two taps within 300ms | Special actions, zoom |
| **Long Press** | Hold 500ms+, <20px movement | Context menu, info |
| **Swipe** | Movement >50px | Navigation, quick actions |
| **Pinch** | Two-finger distance change | Camera zoom, map scale |

### Virtual Controls

Virtual on-screen controls are **automatically enabled on WASM when touch input is detected**:

- **D-Pad**: Directional movement (bottom left)
- **Action Button**: Primary action (bottom right)
- **Secondary Button**: Secondary action (right side)
- **Menu Button**: Pause/menu (top right)

Virtual controls are initialized automatically when:
1. The platform is touch-capable (iOS, Android, or WASM)
2. Touch input is enabled (`useTouchInput = true`)
3. The first touch event is detected

Manual initialization (optional):
```go
inputSystem.InitializeVirtualControls(screenWidth, screenHeight)
```

## Browser Integration

### HTML Configuration

The `build/wasm/game.html` file includes:

**Viewport Meta Tags:**
```html
<meta name="viewport" content="width=device-width, initial-scale=1.0, 
      maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
```

**Web App Meta Tags:**
```html
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="mobile-web-app-capable" content="yes">
```

### CSS Touch Handling

```css
body {
    touch-action: none;           /* Disable default touch behaviors */
    -webkit-user-select: none;    /* Prevent text selection */
    user-select: none;
    overflow: hidden;             /* Prevent scrolling */
}
```

### JavaScript Event Prevention

The HTML includes event handlers to prevent:
- Pinch zoom
- Pull-to-refresh
- Context menu on long press
- Double-tap zoom
- Scroll/pan gestures

## Input Flow

1. **Browser Touch Event** → Touch captured by browser
2. **Ebiten Integration** → `ebiten.TouchIDs()` provides active touches
3. **Touch Handler** → `TouchInputHandler.Update()` processes touch data
4. **Gesture Detection** → `GestureDetector` analyzes patterns
5. **Input System** → Provides input via `InputProvider` interface
6. **Game Response** → Game systems read and respond to input

## Testing

Run tests to verify touch input integration:

```bash
# Platform detection tests
go test ./pkg/mobile/ -v -run TestPlatform

# Integration documentation tests  
go test ./pkg/mobile/ -v -run TestTouch

# Build WASM binary
make build-wasm
```

## Usage Example

Touch input "just works" on WASM without configuration:

```go
// InputSystem automatically detects touch capability
inputSystem := engine.NewInputSystem()

// Touch input activates automatically when touches detected
// Virtual controls appear automatically on first touch
// No special configuration needed for WASM

// Optional: Pre-initialize virtual controls before first touch
inputSystem.InitializeVirtualControls(screenWidth, screenHeight)
```

## Browser Compatibility

Touch input requires:
- WebAssembly support (all modern browsers)
- Touch Events API
- Canvas support

Tested on:
- iOS Safari 14+
- Chrome/Edge 90+ (mobile and desktop)
- Firefox 88+ (mobile and desktop)
- Android Chrome 90+

## Debugging

Enable verbose logging to see touch input detection:

```bash
./venture-client -verbose
```

Console output will show:
- Platform detection: "Platform: WASM"
- Touch capability: "Touch input enabled"
- Touch events: Active touch IDs and positions

## Implementation Details

### Auto-Detection Logic

```go
// From input_system.go (updated for WASM support)
if len(ebiten.TouchIDs()) > 0 {
    s.useTouchInput = true
    // Auto-initialize virtual controls when touch detected
    if s.virtualControls == nil && mobile.IsTouchCapable() {
        screenW, screenH := ebiten.WindowSize()
        s.InitializeVirtualControls(screenW, screenH)
    }
} else if !s.mobileEnabled && len(ebiten.TouchIDs()) == 0 {
    s.useTouchInput = false
}
```

This allows seamless switching between:
- Touch input when user touches screen (virtual controls appear automatically)
- Keyboard/mouse when user uses those instead (virtual controls hidden)
- Dual-input devices can use both methods simultaneously

### Virtual Controls Visibility

```go
// Virtual controls shown on all touch-capable platforms when touch input is active
if s.useTouchInput && s.virtualControls != nil {
    s.virtualControls.Draw(screen)
}
```

Virtual controls are shown dynamically based on input method:
- **Touch detected**: Virtual controls appear automatically
- **Keyboard/mouse only**: Virtual controls hidden
- **Dual-input devices**: Controls appear when touch is used, hidden otherwise

This provides the best experience for all device types:
- Desktop browsers: Keyboard/mouse by default, touch when tapped
- Touch-capable laptops: Can use either input method
- Mobile devices: Touch input with virtual controls always available

## Mobile Keyboard for Text Input (WASM)

### Overview

On mobile browsers, the native keyboard must be explicitly triggered for text input since the game runs in a canvas element. The keyboard bridge in `pkg/mobile/keyboard_wasm.go` handles this automatically.

### How It Works

1. **Hidden Input Element**: A hidden HTML input element is created off-screen
2. **Focus/Blur Control**: Focusing the input shows the mobile keyboard, blurring hides it
3. **Event Forwarding**: Keyboard events from the hidden input are forwarded to Ebiten's `AppendInputChars`

### Text Input Components

All text input UI components follow the mobile keyboard lifecycle pattern:

#### Character Creation (`pkg/engine/character_creation.go`)
```go
// Show keyboard when entering name input step
if !cc.keyboardShown && mobile.IsWASM() {
    mobile.ShowKeyboard()
    cc.keyboardShown = true
}

// Hide keyboard when leaving text input
if cc.keyboardShown && mobile.IsWASM() {
    mobile.HideKeyboard()
    cc.keyboardShown = false
}
```

**Lifecycle:**
- **Name Entry**: Keyboard shown automatically on first Update() in name input step
- **Class Selection**: Keyboard hidden (no text input needed)
- **Portrait Selection**: Keyboard not shown automatically (file path not practical on mobile)
- **Confirmation**: Keyboard remains hidden
- **Completion**: Cleanup() ensures keyboard is hidden

#### Server Address Input (`pkg/engine/server_address_input.go`)
```go
// Show keyboard when input becomes visible
if mobile.IsWASM() {
    mobile.ShowKeyboard()
    s.keyboardShown = true
}

// Hide keyboard on Enter (connect) or Escape (cancel)
if s.keyboardShown && mobile.IsWASM() {
    mobile.HideKeyboard()
    s.keyboardShown = false
}
```

#### Crafting UI Search (`pkg/engine/crafting_ui.go`)
```go
// Show keyboard when opening crafting UI (search active)
if !ui.keyboardShown && mobile.IsWASM() {
    mobile.ShowKeyboard()
    ui.keyboardShown = true
}

// Hide keyboard when closing crafting UI
if ui.keyboardShown && mobile.IsWASM() {
    mobile.HideKeyboard()
    ui.keyboardShown = false
}
```

### Keyboard Lifecycle Pattern

All text input components must follow this pattern:

1. **Show on Entry**: When entering a text input state:
   ```go
   if !keyboardShown && mobile.IsWASM() {
       mobile.ShowKeyboard()
       keyboardShown = true
   }
   ```

2. **Hide on Exit**: When leaving text input state:
   ```go
   if keyboardShown && mobile.IsWASM() {
       mobile.HideKeyboard()
       keyboardShown = false
   }
   ```

3. **Reset Flag on Transition**: When moving between UI states:
   ```go
   keyboardShown = false  // Will be shown by next state if needed
   ```

4. **Cleanup on Completion**: Final cleanup when UI closes completely:
   ```go
   if keyboardShown && mobile.IsWASM() {
       mobile.HideKeyboard()
       keyboardShown = false
   }
   ```

### Implementation Details

**Hidden Input Element Attributes:**
```javascript
input.type = "text"
input.inputmode = "text"          // Optimized keyboard layout
input.autocomplete = "off"        // Disable autocomplete
input.enterkeyhint = "done"       // Show "Done" button
```

**Event Forwarding:**
- Regular keys → Dispatched as `keydown` and `keypress` events
- Backspace → Dispatched as `keydown` with keyCode 8
- Special keys (Enter, Tab, Escape) → Forwarded with correct keyCodes

**Character Processing:**
```go
// Text input handling in UI components
runes := ebiten.AppendInputChars(nil)
for _, r := range runes {
    // Process character (validation, length check, etc.)
    nameInput += string(r)
}
```

### Testing on Mobile WASM

1. Build WASM: `make build-wasm`
2. Serve locally: `python3 -m http.server 8000 -d web/`
3. Open on mobile: `http://<your-ip>:8000`
4. Test scenarios:
   - Character name entry: Keyboard appears automatically
   - Server address input: Keyboard appears when field shown
   - Crafting search: Keyboard appears when UI opens
   - Navigation: Keyboard disappears when leaving text input

### Common Pitfalls to Avoid

❌ **Don't**: Show keyboard before UI is ready
```go
// BAD: Shows keyboard immediately in Reset()
mobile.ShowKeyboard()
```

✅ **Do**: Let Update() show keyboard when state is ready
```go
// GOOD: Reset flag, let updateNameInput() show keyboard
keyboardShown = false
```

❌ **Don't**: Forget to hide keyboard when UI closes
```go
// BAD: Keyboard stays visible after UI closes
```

✅ **Do**: Always hide keyboard in cleanup
```go
// GOOD: Hide keyboard in Hide() or Cleanup()
if keyboardShown && mobile.IsWASM() {
    mobile.HideKeyboard()
    keyboardShown = false
}
```

❌ **Don't**: Show keyboard for non-text input steps
```go
// BAD: Showing keyboard on file path selection (not practical on mobile)
if !keyboardShown && mobile.IsWASM() {
    mobile.ShowKeyboard()
}
```

✅ **Do**: Only show keyboard when text input is actually needed
```go
// GOOD: Only show keyboard on name entry, not file selection
// Portrait selection: keyboard not shown automatically
```

### Debugging Keyboard Issues

Enable verbose logging to debug keyboard visibility:
```bash
./venture-client -verbose
```

Check for:
- "ShowKeyboard called" / "HideKeyboard called" log messages
- Keyboard state flag (`keyboardShown`) matches actual visibility
- Keyboard appears/disappears at correct UI state transitions

### Browser Compatibility

The keyboard bridge works on all modern mobile browsers:
- ✅ iOS Safari 14+ (iPhone, iPad)
- ✅ Android Chrome 90+
- ✅ Android Firefox 88+
- ✅ Mobile Edge, Brave, Samsung Internet

Desktop browsers ignore keyboard calls (keyboard always available).

## Future Enhancements

Possible improvements:
- [ ] Add haptic feedback via Web Vibration API
- [ ] Implement touch-specific camera controls
- [ ] Add swipe gestures for inventory/menu navigation
- [ ] Support for stylus/pen input properties
- [ ] Touch-optimized UI scaling for small screens
- [ ] Custom keyboard types (email, URL, numeric) per input field
- [ ] Predictive text and autocorrect support for mobile typing

## References

- [Ebiten Touch API](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#TouchIDs)
- [Touch Events API](https://developer.mozilla.org/en-US/docs/Web/API/Touch_events)
- [Viewport Configuration](https://developer.mozilla.org/en-US/docs/Web/HTML/Viewport_meta_tag)
