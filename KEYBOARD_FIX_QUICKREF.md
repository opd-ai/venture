# WebAssembly Virtual Keyboard - Quick Reference

## TL;DR - The Fix

**Problem:** Mobile keyboard didn't appear on touch devices in WASM build.

**Root Cause:** CSS `touch-action: none` on body blocked input element from receiving focus.

**Solution:** Allow touch events on input elements while blocking on canvas.

---

## What Changed

### 1. HTML/CSS (build/wasm/game.html)

```css
/* BEFORE - ❌ BROKEN */
body {
    touch-action: none;  /* Blocked ALL touches */
}

/* AFTER - ✅ FIXED */
body {
    /* touch-action removed - allows input interaction */
}
input {
    touch-action: auto !important;      /* Explicitly allow */
    pointer-events: auto !important;    /* Explicitly allow */
}
```

### 2. JavaScript (build/wasm/game.html)

```javascript
// BEFORE - ❌ BLOCKED INPUT
if (e.target.tagName !== 'CANVAS') {
    e.preventDefault();
}

// AFTER - ✅ ALLOWS INPUT
if (e.target.tagName !== 'CANVAS' && e.target.tagName !== 'INPUT') {
    e.preventDefault();
}
```

### 3. Go Code (pkg/mobile/keyboard_wasm.go)

```go
// BEFORE - ❌ OFF-SCREEN ONLY
style.Set("left", "-9999px")
style.Set("pointerEvents", "none")

// AFTER - ✅ DYNAMIC POSITIONING
// OFF-SCREEN by default
style.Set("left", "-9999px")
// (no pointerEvents: none)

// ON-SCREEN when ShowKeyboard() called
style.Set("left", "50%")
style.Set("bottom", "80px")
style.Set("transform", "translateX(-50%)")
```

---

## How It Works

1. **ShowKeyboard() called** → Input moves on-screen (bottom-center)
2. **User taps screen** → Input receives focus → Keyboard appears
3. **User types** → Characters forwarded to Ebiten
4. **HideKeyboard() called** → Input moves off-screen

---

## For Users

**Before:**
- ❌ Keyboard never appeared on mobile
- ❌ Text input impossible

**After:**
- ✅ Tap screen when entering text
- ✅ Native mobile keyboard appears
- ✅ Type normally with autocomplete/suggestions

---

## For Developers

### Using the Keyboard

```go
import "github.com/opd-ai/venture/pkg/mobile"

// Show keyboard (e.g., entering character name)
if mobile.IsWASM() {
    mobile.ShowKeyboard()
}

// Handle input
chars := ebiten.AppendInputChars(nil)
for _, r := range chars {
    name += string(r)
}

// Hide keyboard (e.g., input complete)
if mobile.IsWASM() {
    mobile.HideKeyboard()
}
```

### API (No Changes)

- `mobile.ShowKeyboard()` - Shows mobile keyboard
- `mobile.HideKeyboard()` - Hides mobile keyboard  
- `mobile.IsWASM()` - Check if running in WASM
- `mobile.IsKeyboardSupported()` - Returns true on WASM

### Debugging

Check browser console for `[VentureKeyboard]` logs:

```
[VentureKeyboard] ShowKeyboard() called
[VentureKeyboard] Keyboard element moved on-screen and focused
[VentureKeyboard] Focus successful - active element is venture-keyboard-input
[VentureKeyboard] Input event: new chars added: 'H'
```

### Testing

**Build keyboard test app:**
```bash
bash scripts/build-keyboardtest.sh
cd build/keyboardtest
python3 -m http.server 8080
# Open on mobile: http://your-ip:8080
```

**Test checklist:**
- [ ] Tap screen → keyboard appears
- [ ] Type text → appears in app
- [ ] Backspace works
- [ ] Enter dismisses keyboard
- [ ] No console errors

---

## Files Modified

1. `build/wasm/game.html` - CSS and touch handler fixes
2. `pkg/mobile/keyboard_wasm.go` - Dynamic positioning
3. `cmd/keyboardtest/keyboardtest.html` - Same fixes

Total lines changed: ~90 across 3 files

---

## Common Issues

### "Keyboard still doesn't appear"

**Check:**
1. Browser console - any errors?
2. DevTools Elements - is `#venture-keyboard-input` in DOM?
3. When tapping, does input move on-screen? (check styles)
4. Is focus on the input? (check activeElement)

**Solutions:**
- Some browsers need user tap (can't show via programmatic focus)
- Input moved on-screen at bottom-center - tap there
- Check browser-specific keyboard restrictions

### "Input interferes with game controls"

**Check:**
- Is `HideKeyboard()` being called when input complete?
- Input should move off-screen when keyboard hidden
- Check z-index isn't too high (should be 999)

---

## References

- Full details: `KEYBOARD_FIX_NOVEMBER_2025.md`
- Original implementation: `KEYBOARD_SUMMARY.md`
- Mobile package docs: `pkg/mobile/README.md`

---

*Last Updated: 2025-11-17*  
*Fix Version: November 2025*  
*Status: ✅ Resolved*
