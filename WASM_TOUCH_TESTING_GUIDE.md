# WASM Touch Input - Manual Testing Guide

## Quick Test Checklist

### Prerequisites
- ✅ WASM build completed (19MB binary)
- ✅ All code committed and pushed
- ✅ Code review feedback addressed

### Desktop Testing (5 minutes)
Test that mouse/keyboard functionality is unchanged:

1. **Build and serve:**
   ```bash
   cd /home/runner/work/venture/venture
   make build-wasm
   make serve-wasm
   ```

2. **Test in Chrome (Desktop mode):**
   - Open http://localhost:8080
   - Verify main menu appears
   - Navigate with arrow keys ✓
   - Click menu items with mouse ✓
   - Start new game ✓
   - Open inventory (I key) ✓
   - Click inventory items with mouse ✓
   - Open skills (K key) ✓
   - Click skills with mouse ✓
   - Press ESC for menu ✓

### Mobile Emulation Testing (10 minutes)
Test touch functionality in Chrome DevTools:

1. **Enable mobile emulation:**
   - Press F12 in Chrome
   - Click device toolbar icon (Ctrl+Shift+M)
   - Select "iPad" or "iPhone 12 Pro" from dropdown

2. **Test main menu:**
   - ✓ Virtual controls visible (D-pad bottom-left, buttons bottom-right)
   - ✓ Tap "Single-Player" - should highlight and select
   - ✓ Tap "New Game" - should start character creation
   - ✓ Complete character creation with taps

3. **Test gameplay:**
   - ✓ Touch D-pad - character moves
   - ✓ Touch action button (A) - character attacks
   - ✓ Touch use item button (B) - uses item
   - ✓ Touch menu button - opens menu

4. **Test inventory:**
   - ✓ Tap inventory button or press I
   - ✓ Tap item - selects it
   - ✓ Drag item - starts drag
   - ✓ Drop on another slot - swaps items
   - ✓ Tap close button or ESC - closes inventory

5. **Test skills:**
   - ✓ Tap skills button or press K
   - ✓ Tap skill node - purchases skill (if points available)
   - ✓ Hover over skill - shows tooltip
   - ✓ Tap close or ESC - closes skills UI

6. **Test shop (if merchant present):**
   - ✓ Tap merchant
   - ✓ Tap items in shop inventory
   - ✓ Tap buy/sell buttons
   - ✓ Close shop with tap

### Real Device Testing (15 minutes)
**iPhone Safari:**
1. Get your computer's local IP: `ip addr show | grep "inet "`
2. On iPhone, open Safari: `http://YOUR_IP:8080`
3. Run through all tests above
4. Check console for errors: Settings > Safari > Advanced > Web Inspector

**Android Chrome:**
1. Same IP address as above
2. Open Chrome: `http://YOUR_IP:8080`
3. Run through all tests above
4. Check console: Menu > More tools > Developer tools

### Performance Testing
Monitor performance in DevTools:
- FPS should be 60 (or close)
- Memory stable (not growing)
- No frame drops during touch
- Touch response <100ms

### Known Issues to Watch For
- **Drag-and-drop:** May be finicky on some devices
- **Double-tap:** Should not zoom (prevented in HTML)
- **Scroll:** Page should not scroll (prevented in HTML)
- **Context menu:** Should not appear on long-press (prevented in HTML)

### Success Criteria
All tests above should pass. If any fail:
1. Check browser console for errors
2. Verify WASM file loaded (Network tab)
3. Verify touch events firing (Console > event listeners)
4. Report issue with device/browser details

### Bug Report Template
If you find issues:
```
**Device:** iPhone 13 Pro / Samsung Galaxy S21 / etc.
**Browser:** Safari 16.5 / Chrome 118 / etc.
**Issue:** Touch on skill tree doesn't work
**Steps to reproduce:**
1. Open skills UI
2. Tap on skill node
3. Nothing happens

**Console errors:**
[paste any errors from console]

**Expected:** Skill should be purchased
**Actual:** No response
```

### Deployment Verification
After deploying to GitHub Pages:
1. Visit https://opd-ai.github.io/venture/
2. Run all tests above
3. Verify MIME type: `Content-Type: application/wasm`
4. Test on multiple devices/browsers

## Quick Reference

### Virtual Controls Layout
```
Screen Layout:
┌─────────────────────────────────┐
│         GAME VIEW               │
│                                 │
│                                 │
│  D-Pad          Buttons         │
│  (move)         A B M           │
│   ◯              ● ● ●          │
└─────────────────────────────────┘

D-Pad (bottom-left):
- Touch and drag to move character
- Outer circle = boundaries
- Inner circle = dead zone

Buttons (bottom-right):
- A = Action/Attack
- B = Use Item
- M = Menu
```

### Touch Gestures Supported
- **Tap:** Select menu items, click buttons
- **Touch-and-hold:** Drag items (inventory)
- **Drag:** Move on D-pad, drag items
- **Double-tap:** Prevented (no zoom)
- **Pinch:** Not implemented
- **Swipe:** Not implemented

### Keyboard Shortcuts (Desktop)
- WASD / Arrow keys: Move
- Space: Attack
- E: Use item
- I: Inventory
- K: Skills
- J: Quests
- C: Character
- M: Map
- R: Crafting
- ESC: Menu
- F5: Quick save
- F9: Quick load

## Troubleshooting

### Virtual Controls Not Visible
1. Check platform detection: Should be WASM
2. Check initialization: Look for "virtual controls initialized" in console
3. Check rendering: Controls drawn last (on top)
4. Try refreshing page

### Touch Not Working
1. Verify touch events firing: Add console.log in touch handlers
2. Check if touch-action CSS applied: Inspect element
3. Verify helper functions called: Add debug logs
4. Test with different browsers

### Performance Issues
1. Check FPS in DevTools performance tab
2. Monitor memory usage
3. Verify batch rendering active
4. Check for console warnings

### Build Issues
1. Verify WASM file size (~19MB)
2. Check for compilation errors
3. Verify all dependencies present
4. Try clean build: `make clean && make build-wasm`

## Contact
If you encounter issues during testing, provide:
- Device and browser details
- Console error messages
- Steps to reproduce
- Screenshots if applicable
