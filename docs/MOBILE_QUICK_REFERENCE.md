# Mobile Quick Reference

Quick commands and tips for mobile development.

## Quick Commands

### Setup
```bash
# Install mobile dependencies
make mobile-deps
```

### Android
```bash
# Debug build and install
make android-install

# Release build
export VENTURE_KEYSTORE_FILE=/path/to/keystore.jks
export VENTURE_KEYSTORE_PASSWORD=your_password
export VENTURE_KEY_ALIAS=venture
export VENTURE_KEY_PASSWORD=your_password
make android-apk-release

# Play Store bundle
make android-aab
```

### iOS (macOS only)
```bash
# Simulator
make ios-simulator

# Device
export IOS_SIGNING_IDENTITY="Apple Development: Name (TEAM)"
export IOS_TEAM_ID="TEAM123"
make ios-device

# App Store
make ios-ipa
```

## Touch Controls API

### Detecting Touch
```go
import "github.com/opd-ai/venture/pkg/mobile"

// In your Update() method
touchIDs := ebiten.AppendTouchIDs(nil)
for _, id := range touchIDs {
    x, y := ebiten.TouchPosition(id)
    // Handle touch at (x, y)
}
```

### Virtual Controls
```go
// Get virtual control manager
controls := mobile.NewVirtualControlManager(screenWidth, screenHeight)

// Update each frame
controls.Update()

// Get input
if controls.IsDPadPressed(mobile.DirectionUp) {
    // Move up
}

if controls.IsButtonPressed(mobile.ButtonA) {
    // Primary action
}

// Draw controls
controls.Draw(screen)
```

### Gesture Detection
```go
detector := mobile.NewGestureDetector()

// Update with touch events
detector.Update(touchIDs)

// Check gestures
if detector.IsTap() {
    pos := detector.TapPosition()
    // Handle tap at pos
}

if detector.IsSwipe() {
    dir := detector.SwipeDirection()
    // Handle swipe in direction
}

if detector.IsPinch() {
    scale := detector.PinchScale()
    // Handle pinch zoom
}
```

## Platform Detection

```go
import "github.com/opd-ai/venture/pkg/mobile"

if mobile.IsMobile() {
    // Mobile-specific code
    controls.Show()
} else {
    // Desktop-specific code
    controls.Hide()
}

if mobile.IsAndroid() {
    // Android-specific
}

if mobile.IsIOS() {
    // iOS-specific
}
```

## Screen Sizes

### Common Resolutions
- **iPhone 15 Pro**: 1179x2556 (portrait), 2556x1179 (landscape)
- **iPhone SE**: 750x1334 (portrait), 1334x750 (landscape)
- **iPad Pro 12.9"**: 2048x2732 (portrait), 2732x2048 (landscape)
- **Pixel 7**: 1080x2400 (portrait), 2400x1080 (landscape)
- **Samsung Galaxy S23**: 1080x2340 (portrait), 2340x1080 (landscape)

### Safe Areas
Always account for notches and home indicators:
```go
// Typical safe area insets
topInset := 44    // Status bar / notch
bottomInset := 34 // Home indicator
sideInset := 0    // Usually 0, unless notched

// Position UI within safe area
uiY := topInset + margin
```

## Performance Tips

### Battery Life
- Reduce frame rate when idle: `ebiten.SetMaxTPS(30)` instead of 60
- Disable haptics in battery saver mode
- Reduce particle effects on low-end devices

### Memory
- Keep texture atlas under 2048x2048 for older devices
- Pool frequently allocated objects
- Profile memory on device, not simulator

### Optimization
```go
// Check if device is low-end
if mobile.IsLowEndDevice() {
    // Reduce quality settings
    particleCount /= 2
    shadowQuality = "low"
}
```

## Testing Checklist

### Pre-Release Testing
- [ ] Test on physical iPhone and Android device
- [ ] Test in portrait and landscape modes
- [ ] Test with different screen sizes (small phone, tablet)
- [ ] Test touch controls feel responsive
- [ ] Test with poor network (airplane mode toggle)
- [ ] Test battery usage (should last 2+ hours)
- [ ] Test with VoiceOver/TalkBack enabled
- [ ] Test with reduced motion settings
- [ ] Verify no crashes in background/foreground transitions
- [ ] Test app store screenshots on actual devices

## Common Issues

### Build Fails
```bash
# Clear build cache
go clean -cache
rm -rf build/android/app/build
rm -rf build/ios/DerivedData

# Rebuild
make android-apk  # or ios-simulator
```

### Touch Not Working
- Ensure `mobile.SetGame()` is called in init()
- Check virtual controls are created with correct screen size
- Verify touch input isn't blocked by UI layers

### Poor Performance
- Profile on device, not simulator
- Check entity count (use spatial partitioning)
- Reduce draw calls (batch rendering)
- Lower particle/effect quality on mobile

## File Locations

```
cmd/mobile/mobile.go              - Mobile entry point
pkg/mobile/touch.go               - Touch input handling
pkg/mobile/controls.go            - Virtual controls
pkg/mobile/ui.go                  - Mobile-optimized UI
pkg/mobile/platform.go            - Platform detection
pkg/engine/input_system.go        - Input system with mobile support
build/android/AndroidManifest.xml - Android configuration
build/ios/Info.plist              - iOS configuration
scripts/build-android.sh          - Android build script
scripts/build-ios.sh              - iOS build script
docs/MOBILE_BUILD.md              - Complete build guide
```

## Resources

- [Full Mobile Build Guide](MOBILE_BUILD.md)
- [Ebiten Mobile Docs](https://ebitengine.org/en/documents/mobile.html)
- [Touch Input API](../pkg/mobile/README.md)
- [GitHub Actions Workflows](../.github/workflows/)

## Support

For issues or questions:
- Check [MOBILE_BUILD.md](MOBILE_BUILD.md) troubleshooting section
- Search [GitHub Issues](https://github.com/opd-ai/venture/issues)
- Ask in [Discussions](https://github.com/opd-ai/venture/discussions)
