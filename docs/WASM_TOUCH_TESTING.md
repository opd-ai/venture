# WASM Touch Control Testing Guide

## Quick Start

```bash
# Build and verify
./scripts/verify-wasm-touch.sh

# Or manually:
make build-wasm
make serve-wasm
# Open http://localhost:8080 in Chrome with mobile emulation
```

## Expected Behavior

### On Touch-Capable Devices/Browsers
When you open the WASM build on a touch-capable device (mobile phone, tablet) or browser with mobile emulation:

1. **Virtual D-Pad (Bottom-Left)**
   - Translucent circle control
   - Touch and drag to move player
   - Inner circle follows your touch
   - Color changes when active

2. **Action Buttons (Bottom-Right)**
   - Button "A": Primary action (attack)
   - Button "B": Secondary action (use item)
   - Buttons light up when touched
   - Trigger haptic feedback (on supported devices)

3. **Menu Button (Top-Right)**
   - Small button with "☰" symbol
   - Opens pause/settings menu when tapped

### On Desktop Browsers
When you open the WASM build on a desktop browser without mobile emulation:

1. **No Virtual Controls Visible**
   - Keyboard controls work (WASD for movement, Space for attack, E for use item)
   - Mouse controls work (click to aim/attack)

2. **Automatic Switch to Touch**
   - If you touch the screen on a touch-screen desktop, virtual controls appear
   - Controls remain visible while touching

## Testing Checklist

### Build Verification
- [ ] `make build-wasm` completes without errors
- [ ] `build/wasm/venture.wasm` exists and is >5MB
- [ ] `build/wasm/wasm_exec.js` exists
- [ ] `build/wasm/index.html` and `build/wasm/game.html` exist

### Chrome DevTools Mobile Emulation
- [ ] Open http://localhost:8080 in Chrome
- [ ] Press F12, click device toolbar icon
- [ ] Select "iPad" or "iPhone 12 Pro"
- [ ] Reload page
- [ ] Virtual controls appear immediately (don't need to touch first)
- [ ] D-pad visible in bottom-left (translucent circle)
- [ ] Action buttons visible in bottom-right (A and B)
- [ ] Menu button visible in top-right (☰)
- [ ] Click-drag D-pad moves player character
- [ ] Click action button A triggers attack animation
- [ ] Click action button B uses item (if in inventory)
- [ ] Click menu button opens pause menu
- [ ] No console errors (check Console tab)

### Real Mobile Device Testing (iOS)
- [ ] Open http://YOUR_LOCAL_IP:8080 on iPhone Safari
- [ ] Wait for game to load (may take 10-30 seconds)
- [ ] Virtual controls appear on screen
- [ ] Touch D-pad and drag - player moves smoothly
- [ ] Lift finger - player stops moving
- [ ] Tap action button - attack triggers
- [ ] Tap menu button - menu opens
- [ ] Haptic feedback occurs (vibration) when touching controls
- [ ] No visual glitches or stuttering
- [ ] Game runs at reasonable framerate (30+ FPS)

### Real Mobile Device Testing (Android)
- [ ] Open http://YOUR_LOCAL_IP:8080 on Android Chrome
- [ ] Wait for game to load
- [ ] Virtual controls appear
- [ ] Touch D-pad - player moves
- [ ] Tap buttons - actions trigger
- [ ] Pinch-to-zoom disabled (page should not zoom)
- [ ] Pull-to-refresh disabled (page should not refresh)
- [ ] Game responsive to touch input

### Desktop Browser Testing (Regression Check)
- [ ] Open http://localhost:8080 in Chrome (no mobile emulation)
- [ ] Virtual controls do NOT appear
- [ ] WASD keys move player
- [ ] Space bar attacks
- [ ] E key uses items
- [ ] Mouse click attacks
- [ ] All keyboard shortcuts work (I for inventory, J for quests, etc.)
- [ ] No console errors

### Console Verification
Open browser console (F12 → Console) and verify:
- [ ] Log message: "virtual controls initialized for touch-capable platform" (mobile only)
- [ ] Log message: "Starting Venture - Procedural Action RPG"
- [ ] No errors: "Failed to load WASM"
- [ ] No errors: "WebAssembly instantiation failed"
- [ ] No errors: "touch is not defined"
- [ ] No errors: "Cannot read property 'TouchIDs' of undefined"

### Performance Verification
- [ ] Game loads in <30 seconds on mobile (4G/WiFi)
- [ ] Frame rate is playable (>30 FPS on mobile)
- [ ] Touch response is immediate (<100ms latency)
- [ ] Virtual controls don't lag or stutter
- [ ] Memory usage reasonable (<200MB on mobile)

## Troubleshooting

### Controls Don't Appear

**Symptoms**: Page loads but no virtual controls visible

**Possible Causes**:
1. Not running on touch-capable device
2. Mobile emulation not enabled in Chrome DevTools
3. JavaScript error preventing initialization

**Fix**:
```bash
# Check console for errors
# Open F12 → Console tab
# Look for errors in red

# Verify platform detection
# In browser console, run:
navigator.maxTouchPoints
# Should return >0 on touch devices

# Check if controls initialized
# Look for log message:
# "virtual controls initialized for touch-capable platform"
```

### Controls Appear but Don't Respond

**Symptoms**: Controls visible but touching does nothing

**Possible Causes**:
1. Touch events blocked by CSS
2. JavaScript error in touch handler
3. Controls rendered but not updating

**Fix**:
```bash
# Check game.html has:
touch-action: none;

# Verify no overlays blocking touches
# In console:
document.elementsFromPoint(100, 500)
# Should include canvas element
```

### Build Fails

**Symptoms**: `make build-wasm` errors

**Fix**:
```bash
# Clean and rebuild
make clean-wasm
make build-wasm

# Check Go version
go version
# Should be 1.24.5 or higher

# Verify dependencies
go mod download
go mod verify
```

### Controls Appear on Desktop

**Symptoms**: Virtual controls show on desktop browser when they shouldn't

**Possible Cause**: Browser reports touch capability

**Check**:
```javascript
// In browser console:
navigator.maxTouchPoints
// If >0, browser claims touch support
// This is correct behavior for touch-screen laptops
```

## Manual Testing Scenarios

### Scenario 1: First-Time Mobile User
1. User opens game on mobile browser
2. Game loads, shows loading screen
3. **Expected**: Virtual controls appear immediately on game screen
4. User sees D-pad and buttons
5. User touches D-pad
6. **Expected**: Player character moves
7. **SUCCESS**: Controls work on first interaction

### Scenario 2: Desktop to Mobile Transition
1. User opens game on desktop (keyboard works)
2. User rotates laptop with touchscreen
3. User touches screen
4. **Expected**: Virtual controls fade in
5. **Expected**: Touch and keyboard both work
6. **SUCCESS**: Seamless input method switching

### Scenario 3: Mobile Portrait Orientation
1. User opens game on mobile in portrait mode
2. **Expected**: Controls scale to screen size
3. **Expected**: D-pad in bottom-left, buttons in bottom-right
4. **Expected**: Controls don't overlap with game UI
5. **SUCCESS**: Controls usable in portrait mode

### Scenario 4: Mobile Landscape Orientation
1. User rotates device to landscape
2. **Expected**: Controls reposition to screen edges
3. **Expected**: Controls scale appropriately
4. **Expected**: Full game area visible
5. **SUCCESS**: Controls usable in landscape mode

## Automated Testing (Future)

### Unit Tests
```go
// Test virtual control initialization
func TestVirtualControlsInitialization(t *testing.T) {
    // Simulate WASM environment
    is := engine.NewInputSystem()
    is.InitializeVirtualControls(800, 600)
    
    if !is.HasVirtualControls() {
        t.Error("Virtual controls should be initialized")
    }
}

// Test platform detection
func TestPlatformDetection(t *testing.T) {
    platform := mobile.GetPlatform()
    if platform == mobile.PlatformWASM {
        if !mobile.IsTouchCapable() {
            t.Error("WASM should be touch-capable")
        }
    }
}
```

### Integration Tests (Browser Automation)
```javascript
// Puppeteer test example
const puppeteer = require('puppeteer');

test('virtual controls appear on mobile', async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  
  // Emulate mobile device
  await page.emulate(puppeteer.devices['iPhone 12']);
  
  // Load game
  await page.goto('http://localhost:8080');
  await page.waitForSelector('canvas');
  
  // Check for virtual controls
  // (Would need to add data attributes to controls for testing)
  const hasControls = await page.evaluate(() => {
    // Check if controls are rendered
    // This requires exposing control state via DOM
    return true; // Placeholder
  });
  
  expect(hasControls).toBe(true);
  await browser.close();
});
```

## Performance Benchmarks

### Target Metrics
- **Load Time**: <30 seconds on 4G
- **Frame Rate**: >30 FPS on mobile devices
- **Touch Latency**: <100ms from touch to visual response
- **Memory**: <200MB on mobile devices
- **File Size**: venture.wasm <25MB

### Current Metrics (Post-Fix)
- **Build Size**: ~19MB venture.wasm ✓
- **wasm_exec.js**: ~17KB ✓
- **Total Download**: ~19MB ✓

## Deployment Testing

### GitHub Pages Verification
Once deployed to GitHub Pages:

```bash
# Check MIME type
curl -I https://opd-ai.github.io/venture/venture.wasm | grep -i content-type
# Expected: Content-Type: application/wasm

# Check wasm_exec.js
curl -I https://opd-ai.github.io/venture/wasm_exec.js | grep -i content-type
# Expected: Content-Type: application/javascript

# Test loading
curl -I https://opd-ai.github.io/venture/ | grep -i http
# Expected: HTTP/2 200
```

### Cross-Browser Testing
- [ ] Chrome Android (latest)
- [ ] Safari iOS (latest)
- [ ] Firefox Android (latest)
- [ ] Samsung Internet (latest)
- [ ] Edge Mobile (latest)

## Success Criteria Summary

The WASM touch control fix is successful if:

✅ WASM builds without errors  
✅ Virtual controls appear on mobile browsers  
✅ D-pad moves player when touched  
✅ Action buttons trigger game actions  
✅ Desktop keyboard/mouse still works  
✅ No console errors on mobile  
✅ Controls visible immediately (no delay)  
✅ Performance is acceptable (30+ FPS)  
✅ Works on iOS Safari and Android Chrome  
✅ GitHub Pages deployment succeeds  

## Contact & Support

If issues are found during testing:

1. **Check console logs** - Most issues show in browser console
2. **Verify platform detection** - Run `navigator.maxTouchPoints` in console
3. **Test on multiple devices** - Issue may be device-specific
4. **Check GitHub Issues** - Report bugs with console logs and device info

## Additional Resources

- [Ebiten Touch Input Documentation](https://ebiten.org/en/documents/mobile.html)
- [WebAssembly Browser Compatibility](https://caniuse.com/wasm)
- [Touch Events Specification](https://www.w3.org/TR/touch-events/)
- [Chrome DevTools Device Mode](https://developer.chrome.com/docs/devtools/device-mode/)
