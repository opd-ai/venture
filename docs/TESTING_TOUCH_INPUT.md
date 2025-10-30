# Testing Touch Input on WASM Build

This guide explains how to test touch input functionality in the WebAssembly/browser build of Venture.

## Prerequisites

- Touch-capable device (smartphone, tablet, or touch-capable laptop)
- Modern web browser with WebAssembly support
- Local development environment with Go 1.24.5+

## Building and Testing Locally

### 1. Build the WASM Version

```bash
# From repository root
make build-wasm
```

This compiles the client to WebAssembly and copies necessary files to `build/wasm/`.

### 2. Start Local Server

```bash
# Serve the WASM build
make serve-wasm
```

This starts a local HTTP server on port 8080.

### 3. Test on Device

#### Option A: Test on Local Device
If your development machine has a touchscreen:
1. Open http://localhost:8080 in your browser
2. Tap the screen to test touch input

#### Option B: Test on Mobile Device (Same Network)
1. Get your computer's local IP address:
   ```bash
   # Linux/Mac
   ifconfig | grep "inet "
   # Or
   ip addr show
   ```
2. On your mobile device, open browser and navigate to:
   ```
   http://<your-local-ip>:8080
   ```
   Example: `http://192.168.1.100:8080`

#### Option C: Test Deployed Version
If deployed to GitHub Pages:
```
https://<username>.github.io/<repository>/
```

## Expected Touch Behaviors

### Touch Input Detection

When you first touch the screen:
1. **Virtual controls appear** automatically (D-pad on left, action buttons on right)
2. **Touch input activates** (you should see the `useTouchInput` flag become true)
3. **Game responds** to touch events

### Gesture Testing

Test each gesture type:

#### 1. Tap (Attack/Interact)
- **Action**: Quickly tap screen (outside virtual controls)
- **Expected**: Character performs attack or interacts with nearby object
- **Criteria**: Touch duration <300ms, movement <20px

#### 2. Swipe (Movement/Navigation)
- **Action**: Swipe across screen
- **Expected**: Character moves in swipe direction, or menu navigates
- **Criteria**: Movement >50px, quick gesture

#### 3. Long Press (Context Menu/Info)
- **Action**: Touch and hold for 500ms+
- **Expected**: Context menu appears or item info displays
- **Criteria**: Hold 500ms+, movement <20px

#### 4. Double Tap (Special Action)
- **Action**: Tap twice quickly
- **Expected**: Special action or zoom toggle
- **Criteria**: Two taps within 300ms

#### 5. Pinch (Zoom)
- **Action**: Two-finger pinch/spread
- **Expected**: Camera zoom in/out
- **Criteria**: Distance change between two touch points

### Virtual Controls Testing

Test each virtual control element:

#### D-Pad (Bottom Left)
- **Action**: Touch and drag on circular D-pad
- **Expected**: Character moves in direction you drag
- **Verification**: Movement should be smooth and responsive

#### Action Button (Bottom Right, labeled "A")
- **Action**: Tap the circular "A" button
- **Expected**: Character performs primary action (attack)
- **Verification**: Button should visually respond (color change)

#### Secondary Button (Right Side, labeled "B")
- **Action**: Tap the circular "B" button  
- **Expected**: Character performs secondary action (use item)
- **Verification**: Item in inventory should be used

#### Menu Button (Top Right, "☰" icon)
- **Action**: Tap the menu button
- **Expected**: Pause menu opens
- **Verification**: Game pauses, menu appears

## Debugging Touch Input

### Enable Verbose Logging

Add debug logging to see touch events:

1. Open browser developer console (F12 or right-click → Inspect)
2. Watch for console messages indicating touch events

### Check Touch Detection

Verify touch input is detected:
```javascript
// In browser console
document.addEventListener('touchstart', (e) => {
    console.log('Touch detected:', e.touches.length, 'touches');
});
```

### Verify WASM Platform Detection

The game should log platform detection at startup:
```
Platform: WASM
Touch capability: true
```

## Common Issues and Solutions

### Issue: Virtual Controls Don't Appear

**Symptoms**: Touch events work but no on-screen controls visible

**Causes**:
1. Virtual controls not initialized
2. Touch input not detected yet

**Solutions**:
- Ensure you've touched the screen at least once
- Check that `useTouchInput` flag is true
- Verify `InitializeVirtualControls()` was called

### Issue: Touch Events Not Detected

**Symptoms**: Touching screen has no effect

**Causes**:
1. Browser doesn't support Touch Events API
2. WASM platform detection failed
3. Touch events are being consumed by default handlers

**Solutions**:
- Test in a different browser (Chrome, Firefox, Safari)
- Check browser console for errors
- Verify HTML has `touch-action: none` CSS
- Ensure `preventDefault()` is called on touch events

### Issue: Controls Too Small or Large

**Symptoms**: Virtual controls are wrong size for screen

**Causes**:
1. Screen size detection incorrect
2. Virtual controls initialized with wrong dimensions

**Solutions**:
- Check that `InitializeVirtualControls(screenWidth, screenHeight)` uses correct values
- Virtual controls auto-size based on screen height (15% for D-pad, 8% for buttons)
- Refresh page to reinitialize with correct size

### Issue: Touch and Mouse Conflict

**Symptoms**: Both touch and mouse input active simultaneously, causing erratic behavior

**Causes**:
1. Some browsers generate both touch and mouse events for compatibility

**Solutions**:
- This is handled automatically by the InputSystem
- Touch input takes priority when detected
- Mouse events are ignored when `useTouchInput` is true

## Performance Testing

### Frame Rate
- **Target**: 60 FPS
- **Acceptable**: 30 FPS minimum
- **Tool**: Browser Performance tab (F12 → Performance)

### Touch Latency
- **Target**: <50ms from touch to action
- **Acceptable**: <100ms
- **Test**: Tap attack button and observe response time

### Memory Usage
- **Target**: <100MB RAM
- **Acceptable**: <200MB
- **Tool**: Browser Memory tab (F12 → Memory)

## Test Checklist

Use this checklist when testing touch input:

- [ ] Virtual controls appear on first touch
- [ ] D-pad controls character movement (8 directions)
- [ ] Action button triggers attack/interact
- [ ] Secondary button uses item from inventory
- [ ] Menu button opens pause menu
- [ ] Tap gesture detected correctly
- [ ] Swipe gesture moves character
- [ ] Long press shows context menu
- [ ] Double tap performs special action
- [ ] Pinch zoom works for camera
- [ ] Touch input switches off when using keyboard
- [ ] Touch input switches back on when touching screen
- [ ] No conflicts between touch and mouse input
- [ ] Controls are appropriately sized for screen
- [ ] Game runs at acceptable frame rate (30+ FPS)
- [ ] No console errors related to touch events

## Automated Testing

While manual testing is essential for touch input, some aspects can be automated:

### Unit Tests

Run the touch input unit tests:
```bash
go test ./pkg/engine -run TestInputSystem_Touch -v
go test ./pkg/mobile -v
```

### Integration Tests

Test the full touch input pipeline:
```bash
go test ./pkg/mobile -run TestTouch -v
```

## Reporting Issues

When reporting touch input issues, include:

1. **Device information**:
   - Device type (phone, tablet, laptop)
   - Operating system and version
   - Browser and version

2. **Expected behavior**: What should happen

3. **Actual behavior**: What actually happens

4. **Steps to reproduce**: Exact actions to trigger the issue

5. **Console logs**: Any error messages from browser console

6. **Screenshots/video**: If possible, visual evidence of the issue

## Additional Resources

- [Touch Input WASM Documentation](TOUCH_INPUT_WASM.md)
- [Ebiten Touch API](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#TouchIDs)
- [Touch Events API](https://developer.mozilla.org/en-US/docs/Web/API/Touch_events)
- [Mobile Build Guide](MOBILE_BUILD.md)
