# WASM Touch Input - Manual Testing Guide (Quick Reference)

This is a quick reference for manually testing the touch input implementation on WASM. For comprehensive details, see [TESTING_TOUCH_INPUT.md](TESTING_TOUCH_INPUT.md).

## Quick Start

```bash
# From repository root
./scripts/test-wasm-touch.sh
```

This script will:
1. Build the WASM version
2. Start a local server
3. Display your local IP for mobile testing
4. Show a testing checklist

## What Changed

The implementation now enables touch input on WASM/browser builds by:
1. Using `useTouchInput` flag instead of `mobileEnabled` for virtual controls
2. Auto-detecting touch events and initializing virtual controls dynamically
3. Supporting seamless switching between touch and keyboard/mouse input

## Key Files Modified

- `pkg/engine/input_system.go` - Core touch input logic
- `pkg/engine/input_system_extended_test.go` - Unit tests
- `docs/TOUCH_INPUT_WASM.md` - Technical documentation
- `docs/TESTING_TOUCH_INPUT.md` - Comprehensive testing guide

## Expected Behavior

### On Desktop Browser (with mouse/keyboard)
- Game starts with keyboard/mouse input
- Touch screen → virtual controls appear
- Use keyboard again → virtual controls disappear
- Touch again → virtual controls reappear

### On Mobile Browser
- Game starts in touch mode
- Virtual controls appear on first touch
- D-pad on left, action buttons on right
- Menu button on top right

### Virtual Controls Layout
```
┌─────────────────────────────────┐
│                          [☰]    │  ← Menu
│                                 │
│                                 │
│                                 │
│                              [B]│  ← Secondary
│                                 │
│                           [A]   │  ← Action
│  (D-PAD)                        │
│     ↑                           │
│   ←   →                         │
│     ↓                           │
└─────────────────────────────────┘
```

## Test Priorities

### Critical (Must Work)
1. ✅ Virtual controls appear on touch
2. ✅ D-pad moves character
3. ✅ Action button works
4. ✅ Touch detection activates correctly

### Important (Should Work)
5. ✅ Secondary button works
6. ✅ Menu button works
7. ✅ Tap gesture detected
8. ✅ Input method switching

### Nice-to-Have (Good to Test)
9. Swipe gesture
10. Long press gesture
11. Double tap gesture
12. Pinch zoom gesture

## Quick Test Procedure

### 5-Minute Smoke Test
1. Open game in mobile browser
2. Touch screen → controls appear ✓
3. Drag D-pad → character moves ✓
4. Tap [A] button → character attacks ✓
5. Tap [☰] button → menu opens ✓

**If these 5 things work, the implementation is successful!**

### 10-Minute Full Test
Follow the smoke test, then:
6. Tap [B] button → item used ✓
7. Quick tap empty space → attack ✓
8. Connect keyboard, press W → controls vanish ✓
9. Touch again → controls reappear ✓
10. Check browser console → no errors ✓

## Common Issues & Solutions

### Issue: Controls Don't Appear
**Solution**: Touch anywhere on screen. Controls initialize on first touch.

### Issue: Controls Too Small/Large
**Solution**: Refresh page. Controls auto-size based on screen height.

### Issue: Touch Not Detected
**Check**:
1. Browser supports Touch Events API (Chrome, Firefox, Safari do)
2. HTML has `touch-action: none` CSS (it does)
3. Browser console for errors (F12)

### Issue: Controls Block View
**Note**: This is expected. Virtual controls are semi-transparent and positioned in corners to minimize obstruction.

## Performance Targets

- **FPS**: 30+ on mobile, 60 on desktop
- **Touch Latency**: <100ms from touch to action
- **Memory**: <200MB on mobile
- **Load Time**: <10 seconds on 4G

## Browser Compatibility

### Supported ✅
- Chrome 90+ (mobile & desktop)
- Firefox 88+ (mobile & desktop)
- Safari 14+ (mobile & desktop)
- Edge 90+

### Not Supported ❌
- Internet Explorer
- Very old browsers (<2020)

## Debugging Commands

```bash
# Check if WASM file exists
ls -lh build/wasm/venture.wasm

# Check if server is running
curl -I http://localhost:8080

# View server logs
# (If using make serve-wasm, logs appear in terminal)
```

## Browser Console Checks

Open DevTools (F12) and check:
1. **Console**: No errors related to touch events
2. **Network**: venture.wasm loaded successfully (10-15MB)
3. **Performance**: Frame rate stays above 30 FPS

## Reporting Results

When reporting test results, include:
1. Device type (phone/tablet/laptop)
2. Browser and version
3. Which tests passed/failed
4. Any console errors
5. Screenshots if possible

## Success Criteria

The implementation is successful if:
- ✅ Touch input works on mobile browsers
- ✅ Virtual controls appear automatically
- ✅ All critical tests pass
- ✅ No regressions on desktop (keyboard/mouse still works)
- ✅ No console errors related to touch

## Next Steps After Testing

1. **If all tests pass**: Mark PR as ready for merge
2. **If minor issues**: Create follow-up issues for non-critical bugs
3. **If major issues**: Fix critical bugs before merging

## Questions?

See [TESTING_TOUCH_INPUT.md](TESTING_TOUCH_INPUT.md) for comprehensive testing guide.
