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

### Virtual Controls (Optional)

Virtual on-screen controls are available but **disabled by default on WASM**:

- **D-Pad**: Directional movement (bottom left)
- **Action Button**: Primary action (bottom right)
- **Secondary Button**: Secondary action (right side)
- **Menu Button**: Pause/menu (top right)

To enable virtual controls on WASM:
```go
inputSystem.SetMobileEnabled(true)
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
go test -tags test ./pkg/mobile/ -v -run TestPlatform

# Integration documentation tests  
go test -tags test ./pkg/mobile/ -v -run TestTouch

# Build WASM binary
make build-wasm
```

## Usage Example

Touch input "just works" on WASM without configuration:

```go
// InputSystem automatically detects touch capability
inputSystem := engine.NewInputSystem()

// Touch input activates when touches detected
// No special configuration needed for WASM

// Optional: Enable virtual controls for pure touch devices
if needsVirtualControls {
    inputSystem.SetMobileEnabled(true)
    inputSystem.InitializeVirtualControls(800, 600)
}
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
// From input_system.go:392-397
if s.mobileEnabled && len(ebiten.TouchIDs()) > 0 {
    s.useTouchInput = true
} else if !s.mobileEnabled && len(ebiten.TouchIDs()) == 0 {
    s.useTouchInput = false
}
```

This allows seamless switching between:
- Touch input when user touches screen
- Keyboard/mouse when user uses those instead

### Virtual Controls Visibility

```go
// Virtual controls only shown on true mobile platforms
if s.mobileEnabled && s.virtualControls != nil {
    s.virtualControls.Draw(screen)
}
```

WASM doesn't show virtual controls by default because:
- Desktop browsers have keyboard/mouse
- Touch-capable laptops can use either input method
- User can enable if needed for touch-only devices

## Future Enhancements

Possible improvements:
- [ ] Add haptic feedback via Web Vibration API
- [ ] Implement touch-specific camera controls
- [ ] Add swipe gestures for inventory/menu navigation
- [ ] Support for stylus/pen input properties
- [ ] Touch-optimized UI scaling for small screens

## References

- [Ebiten Touch API](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#TouchIDs)
- [Touch Events API](https://developer.mozilla.org/en-US/docs/Web/API/Touch_events)
- [Viewport Configuration](https://developer.mozilla.org/en-US/docs/Web/HTML/Viewport_meta_tag)
