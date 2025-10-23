# Mobile Package

Touch input, gesture detection, and mobile-optimized UI components for Venture.

## Overview

This package provides everything needed to run Venture on iOS and Android:
- Touch input handling via Ebiten's touch API
- Gesture detection (tap, swipe, pinch, long-press)
- Virtual on-screen controls (D-pad, buttons)
- Mobile-optimized UI components
- Platform detection utilities

## Components

### Touch Input Handler

Processes raw touch events from Ebiten and provides a higher-level API:

```go
handler := mobile.NewTouchHandler()

// Update each frame with current touch state
touchIDs := ebiten.AppendTouchIDs(nil)
handler.Update(touchIDs)

// Query active touches
for _, touch := range handler.GetActiveTouches() {
    fmt.Printf("Touch %d at (%d, %d)\n", touch.ID, touch.X, touch.Y)
}
```

### Gesture Detector

Recognizes common touch gestures:

```go
detector := mobile.NewGestureDetector()
detector.Update(touchIDs)

// Check for gestures
if detector.IsTap() {
    pos := detector.TapPosition()
    // Handle tap
}

if detector.IsSwipe() {
    direction := detector.SwipeDirection()
    distance := detector.SwipeDistance()
    // Handle swipe
}

if detector.IsPinch() {
    scale := detector.PinchScale()
    center := detector.PinchCenter()
    // Handle pinch zoom
}

if detector.IsLongPress() {
    pos := detector.LongPressPosition()
    // Handle long press (context menu, etc.)
}
```

### Virtual Controls

On-screen controls for touch input:

```go
manager := mobile.NewVirtualControlManager(screenWidth, screenHeight)

// Update each frame
manager.Update()

// Check D-pad
if manager.IsDPadPressed(mobile.DirectionUp) {
    player.MoveUp()
}

// Check buttons
if manager.IsButtonPressed(mobile.ButtonA) {
    player.Attack()
}

// Draw controls
manager.Draw(screen)

// Configure appearance
manager.SetOpacity(0.7)
manager.SetControlSize(80) // Radius in pixels
```

### Mobile UI Components

Touch-optimized UI widgets:

```go
button := mobile.NewTouchButton(
    100, 100,    // Position
    200, 60,     // Size
    "START",     // Text
)

button.OnTap = func() {
    game.Start()
}

// Update and draw
button.Update(touchHandler)
button.Draw(screen)

// Slider for settings
slider := mobile.NewTouchSlider(
    100, 200, 300, // Position and width
    0.0, 1.0,      // Min/max values
    0.5,           // Initial value
)

slider.OnChange = func(value float64) {
    audio.SetVolume(value)
}
```

### Platform Detection

Runtime detection of platform and capabilities:

```go
if mobile.IsMobile() {
    // Show virtual controls
    controls.Show()
} else {
    // Hide virtual controls
    controls.Hide()
}

if mobile.IsAndroid() {
    // Android-specific code
    mobile.RequestPermission("android.permission.VIBRATE")
}

if mobile.IsIOS() {
    // iOS-specific code
}

if mobile.SupportsHaptics() {
    mobile.Vibrate(mobile.HapticLight)
}

if mobile.IsLowEndDevice() {
    // Reduce quality settings
}
```

## Usage in Game

### Initialization

```go
// In game initialization
func NewGame() *Game {
    game := &Game{
        touchHandler: mobile.NewTouchHandler(),
        gestures:     mobile.NewGestureDetector(),
    }
    
    if mobile.IsMobile() {
        game.virtualControls = mobile.NewVirtualControlManager(
            screenWidth, screenHeight,
        )
    }
    
    return game
}
```

### Update Loop

```go
func (g *Game) Update() error {
    touchIDs := ebiten.AppendTouchIDs(nil)
    
    // Update touch input
    g.touchHandler.Update(touchIDs)
    g.gestures.Update(touchIDs)
    
    // Update virtual controls
    if g.virtualControls != nil {
        g.virtualControls.Update()
    }
    
    // Process input
    if g.virtualControls != nil {
        g.processVirtualInput()
    } else {
        g.processKeyboardMouse()
    }
    
    // Handle gestures
    if g.gestures.IsTap() {
        g.handleTap(g.gestures.TapPosition())
    }
    
    return nil
}
```

### Draw Loop

```go
func (g *Game) Draw(screen *ebiten.Image) {
    // Draw game content
    g.drawWorld(screen)
    g.drawEntities(screen)
    g.drawUI(screen)
    
    // Draw virtual controls on top
    if g.virtualControls != nil {
        g.virtualControls.Draw(screen)
    }
}
```

## Virtual Control Layout

Default layout for landscape mode:

```
+------------------------------------------+
|                              [Menu]      |  <- Top right menu button
|                                          |
|                                          |
|                                          |
|   (D-Pad)                      [Y]       |  <- Bottom area
|      ↑                                   |
|    ← + →              [X]     [A]       |  <- Action buttons (right)
|      ↓                         [B]       |  <- D-pad (left)
+------------------------------------------+
```

Portrait mode adjusts automatically to maintain reachability.

## Touch Gesture Specifications

### Tap
- Touch duration < 300ms
- Movement < 10 pixels
- Single finger

### Double Tap
- Two taps within 500ms
- At same location (±20 pixels)

### Long Press
- Touch duration > 500ms
- Movement < 10 pixels
- Single finger

### Swipe
- Touch duration < 500ms
- Movement > 50 pixels
- Single finger
- Direction: Up/Down/Left/Right

### Pinch
- Two fingers
- Distance change > 20 pixels
- Returns scale factor (1.0 = no change)

## Performance Considerations

### Touch Processing

Touch input processing is lightweight:
- O(n) where n = number of active touches (typically 1-2)
- No allocations in hot path
- Updates take < 0.1ms on mobile devices

### Virtual Controls

Virtual controls use cached sprites:
- One sprite per control type (D-pad, buttons)
- Minimal draw calls (batched rendering)
- Transparency handled efficiently

### Best Practices

```go
// ✅ Good - Cache touch state
type Game struct {
    lastTapPos Vec2
    lastTapTime time.Time
}

// ❌ Bad - Creating new objects each frame
func (g *Game) Update() {
    handler := mobile.NewTouchHandler() // Don't do this!
}

// ✅ Good - Reuse gesture detector
func (g *Game) Update() {
    g.gestures.Update(touchIDs)
}
```

## Testing

### Unit Tests

```bash
go test -tags test ./pkg/mobile/...
```

### On Device

Test on physical devices, not just simulator:
- Touch responsiveness varies by hardware
- Simulator doesn't reflect actual touch latency
- Test with different finger sizes (accessibility)

### Test Cases

```go
func TestTouchDetection(t *testing.T) {
    handler := NewTouchHandler()
    
    // Simulate touch down
    touch := Touch{ID: 0, X: 100, Y: 100, State: TouchStateBegan}
    handler.UpdateWithTouches([]Touch{touch})
    
    assert.Equal(t, 1, len(handler.GetActiveTouches()))
}
```

## API Reference

### TouchHandler

```go
type TouchHandler struct { }

func NewTouchHandler() *TouchHandler
func (h *TouchHandler) Update(touchIDs []ebiten.TouchID)
func (h *TouchHandler) GetActiveTouches() []Touch
func (h *TouchHandler) GetTouch(id ebiten.TouchID) (Touch, bool)
```

### GestureDetector

```go
type GestureDetector struct { }

func NewGestureDetector() *GestureDetector
func (d *GestureDetector) Update(touchIDs []ebiten.TouchID)
func (d *GestureDetector) IsTap() bool
func (d *GestureDetector) IsSwipe() bool
func (d *GestureDetector) IsPinch() bool
func (d *GestureDetector) IsLongPress() bool
```

### VirtualControlManager

```go
type VirtualControlManager struct { }

func NewVirtualControlManager(width, height int) *VirtualControlManager
func (m *VirtualControlManager) Update()
func (m *VirtualControlManager) Draw(screen *ebiten.Image)
func (m *VirtualControlManager) IsDPadPressed(dir Direction) bool
func (m *VirtualControlManager) IsButtonPressed(btn Button) bool
func (m *VirtualControlManager) SetOpacity(opacity float64)
```

## Examples

See:
- `examples/mobile_demo/` - Complete mobile demo
- `cmd/mobile/mobile.go` - Mobile entry point
- `pkg/engine/input_system.go` - Integration example

## Platform Support

- **iOS**: iOS 14.0+, arm64
- **Android**: API 21+ (Android 5.0+), armeabi-v7a, arm64-v8a
- **Ebiten**: v2.9.0+

## License

Same as Venture project (see LICENSE file).
