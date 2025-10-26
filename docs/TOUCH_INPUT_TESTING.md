# Touch Input Testing Guide

This guide provides instructions for testing the touch input implementation on various devices and browsers.

## Build and Deploy

### Build WASM Binary

```bash
# Build the WASM binary
make build-wasm

# Or manually:
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/wasm/venture.wasm ./cmd/client
cp $(go env GOROOT)/lib/wasm/wasm_exec.js build/wasm/
```

### Local Testing

```bash
# Serve locally
make serve-wasm

# Or manually:
cd build/wasm
python3 -m http.server 8080
```

Then open http://localhost:8080 in your browser.

## Testing Checklist

### Basic Touch Input

- [ ] **Touch Detection**
  - Touch the screen
  - Game should respond to touch (no keyboard needed)
  - Console should show touch events if verbose mode enabled

- [ ] **No Unwanted Behaviors**
  - Pinch gestures should NOT zoom the page
  - Scrolling should NOT move the page
  - Pull-to-refresh should NOT trigger
  - Long press should NOT show context menu
  - Double-tap should NOT zoom

### Touch Gestures

- [ ] **Tap Gesture**
  - Quick tap on empty space
  - Should be recognized as tap action
  - Could trigger attack or interaction depending on game state

- [ ] **Swipe Gesture**
  - Swipe across the screen
  - Should detect direction and distance
  - Could be used for camera pan or quick navigation

- [ ] **Long Press**
  - Touch and hold for 1+ second
  - Should recognize as long press
  - Could show info or context menu

- [ ] **Double Tap**
  - Quickly tap twice in same location
  - Should recognize as double tap
  - Could be used for special actions

- [ ] **Pinch Gesture**
  - Two-finger pinch/spread
  - Should detect zoom scale
  - Could be used for camera zoom

### Input Method Switching

- [ ] **Touch → Keyboard**
  - Start with touch input
  - Press a key
  - Game should accept keyboard input

- [ ] **Keyboard → Touch**
  - Start with keyboard
  - Touch the screen
  - Game should accept touch input

### Virtual Controls (Optional)

Note: Virtual controls are disabled by default on WASM. To test:

1. Modify client code to enable:
   ```go
   inputSystem.SetMobileEnabled(true)
   inputSystem.InitializeVirtualControls(screenWidth, screenHeight)
   ```

2. Rebuild and test:
   - [ ] D-pad appears bottom left
   - [ ] Action button appears bottom right
   - [ ] Secondary button appears right side
   - [ ] Menu button appears top right
   - [ ] Controls respond to touch
   - [ ] Movement works via D-pad
   - [ ] Actions trigger via buttons

## Device Testing

### iOS Devices (Safari)

**iPhone/iPad (Safari):**
- [ ] Page loads correctly
- [ ] Touch input works immediately
- [ ] No zoom on double-tap
- [ ] No scroll on swipe
- [ ] Standalone mode works (Add to Home Screen)
- [ ] Notch areas handled correctly (viewport-fit=cover)

**Testing Tip:** Open Safari DevTools from Mac via USB for console logging.

### Android Devices (Chrome)

**Phone/Tablet (Chrome):**
- [ ] Page loads correctly
- [ ] Touch input works immediately
- [ ] No zoom on pinch
- [ ] No scroll on drag
- [ ] Chrome DevTools remote debugging shows touch events

**Testing Tip:** Use `chrome://inspect` on desktop Chrome for remote debugging.

### Desktop Browsers with Touch

**Chrome/Edge with Touch Screen:**
- [ ] Touch input works on touch-capable laptops
- [ ] Keyboard/mouse also works
- [ ] Can switch between input methods seamlessly

**Testing Tip:** Open DevTools, enable "Toggle device toolbar" (Ctrl+Shift+M), select a mobile device, and use mouse to simulate touch.

### Desktop Browsers without Touch

**Chrome/Firefox/Safari (no touch):**
- [ ] Game loads and works with keyboard/mouse
- [ ] No console errors about touch
- [ ] Virtual controls not shown (expected)

## Browser DevTools Testing

### Chrome DevTools

1. Open DevTools (F12)
2. Go to "Device Mode" (Ctrl+Shift+M)
3. Select device: iPhone, iPad, or Android
4. Test touch interactions with mouse

**Console Commands:**
```javascript
// Check touch capability
console.log('Touch events supported:', 'ontouchstart' in window);

// Monitor touch events
document.addEventListener('touchstart', e => console.log('Touch start:', e.touches.length));
document.addEventListener('touchmove', e => console.log('Touch move:', e.touches.length));
document.addEventListener('touchend', e => console.log('Touch end'));
```

### Expected Console Output

With verbose mode:
```
Platform: WASM
Touch input enabled: true
Mobile controls enabled: false
Touch handler initialized
[Touch detected] IDs: [0]
[Touch position] ID 0: (234, 567)
[Gesture detected] Tap at (234, 567)
```

## Performance Testing

- [ ] **Frame Rate**
  - Touch input should maintain 60 FPS
  - No lag when processing touch events
  - Smooth gesture recognition

- [ ] **Responsiveness**
  - Touch should register within 16ms
  - No delayed reactions
  - Gesture detection feels immediate

- [ ] **Memory**
  - No memory leaks with repeated touches
  - Touch handler properly cleans up
  - GestureDetector resets state correctly

## Edge Cases

- [ ] **Multi-Touch**
  - Two-finger gestures work correctly
  - Three+ finger touches handled gracefully
  - Doesn't interfere with single-touch input

- [ ] **Touch Near Edges**
  - Touches near screen edges work
  - No interference from OS gestures
  - Fullscreen mode works correctly

- [ ] **Orientation Change**
  - Works in portrait mode
  - Works in landscape mode
  - Handles orientation changes smoothly

- [ ] **Background/Foreground**
  - Resumes correctly after backgrounding
  - Cleans up touches on background
  - No stuck touch states

## Debugging Tips

### Touch Not Working

1. Check console for errors
2. Verify GOOS=js build
3. Check viewport meta tags present
4. Ensure JavaScript loaded (wasm_exec.js)
5. Test on different browser

### Unwanted Zoom/Scroll

1. Check viewport meta tag: `maximum-scale=1.0, user-scalable=no`
2. Verify CSS: `touch-action: none`
3. Check JavaScript event handlers loaded
4. Test in incognito mode (no extensions)

### Virtual Controls Not Showing

This is expected! Virtual controls are disabled by default on WASM.
To enable: modify client code and rebuild.

### Gestures Not Detected

1. Check touch threshold values in GestureDetector
2. Verify TouchInputHandler.Update() is called
3. Test with larger/longer gestures
4. Check console for gesture events

## Automated Testing

While manual testing is required for true touch input, you can verify the build:

```bash
# Verify platform detection
go test ./pkg/mobile/ -run TestPlatform -v

# Verify integration points
go test ./pkg/mobile/ -run TestTouch -v

# Build test
GOOS=js GOARCH=wasm go build -o /tmp/test.wasm ./cmd/client
echo "Build status: $?"
```

## Reporting Issues

When reporting touch input issues, include:

1. **Device:** iPhone 12, Pixel 5, etc.
2. **Browser:** Safari 14.5, Chrome 90, etc.
3. **OS Version:** iOS 14.5, Android 11, etc.
4. **Issue:** Description of problem
5. **Console Output:** Copy/paste any errors
6. **Steps to Reproduce:** Detailed steps
7. **Expected vs Actual:** What should happen vs what happens

## Success Criteria

✅ Touch input implementation is successful if:

1. Touch events are detected on mobile browsers
2. No unwanted zoom/scroll/pull-to-refresh
3. Basic gestures work (tap, swipe, pinch)
4. Can switch between touch and keyboard/mouse
5. No console errors related to touch
6. Performance remains at 60 FPS
7. Works on iOS Safari and Android Chrome

## Next Steps

After successful testing:

1. Test on multiple device types
2. Gather user feedback
3. Consider enabling virtual controls for pure-touch devices
4. Implement touch-specific camera controls
5. Add haptic feedback via Web Vibration API
6. Optimize touch performance for low-end devices
