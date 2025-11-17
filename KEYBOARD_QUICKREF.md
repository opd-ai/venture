# Virtual Keyboard Quick Reference

## Quick Start - Testing

### Test Keyboard in Isolation
```bash
./scripts/build-keyboardtest.sh
cd build/keyboardtest
python3 -m http.server 8080
# Open http://localhost:8080 on mobile browser
```

### Test Full Game
```bash
make build-wasm
cd build/wasm
python3 -m http.server 8080
# Open http://localhost:8080/game.html
```

## Quick Checks

### ✅ Is Keyboard Working?
1. Open browser console (F12 or remote debugging)
2. Look for: `[VentureKeyboard] ShowKeyboard() called`
3. Look for: `[VentureKeyboard] Focus successful`
4. Type characters - look for: `[VentureKeyboard] Input event: new chars added: 'X'`

### ✅ Using Fallback (If Keyboard Doesn't Work)
In character creation screen:
- Tap **"Warrior"** button → name set to "Warrior"
- Tap **"Mage"** button → name set to "Mage"
- Tap **"Rogue"** button → name set to "Rogue"
- Tap **"Ranger"** button → name set to "Ranger"
- Tap **"Auto"** button → random name generated (e.g., "Braveblade")

## Console Log Reference

### Success Pattern
```
[VentureKeyboard] Initializing virtual keyboard element
[VentureKeyboard] Virtual keyboard element created and added to DOM
[VentureKeyboard] ShowKeyboard() called
[VentureKeyboard] Focus successful - active element is venture-keyboard-input
[VentureKeyboard] Input event: new chars added: 'a'
[VentureKeyboard] Input event: new chars added: 'b'
```

### Failure Pattern
```
[VentureKeyboard] ShowKeyboard() called
[VentureKeyboard] Focus failed - active element is: CANVAS
```
→ Use preset name buttons instead

## Browser Remote Debugging

### iOS Safari
1. iPhone: Settings > Safari > Advanced > Enable Web Inspector
2. Connect iPhone to Mac via USB
3. Mac Safari: Develop > [iPhone Name] > [Page]

### Android Chrome
1. Android: Settings > Developer Options > Enable USB Debugging
2. Connect to computer via USB
3. Computer Chrome: `chrome://inspect`
4. Click "inspect" on your device

## File Locations

- Keyboard implementation: `pkg/mobile/keyboard_wasm.go`
- Character creation (with presets): `pkg/engine/character_creation.go`
- Test app: `cmd/keyboardtest/main.go`
- Test HTML: `cmd/keyboardtest/keyboardtest.html`
- Build script: `scripts/build-keyboardtest.sh`

## Common Issues

**Keyboard doesn't appear:**
→ Use preset name buttons (Warrior, Mage, Rogue, Ranger, Auto)

**Characters don't appear:**
→ Check console for input event logs
→ Use preset name buttons

**Can't complete character creation:**
→ Should never happen - preset buttons always work

## Success Criteria

✅ Character creation completable without physical keyboard
✅ Console shows keyboard lifecycle
✅ Preset buttons work as fallback
✅ No JavaScript errors in console

---

For detailed testing instructions, see: `KEYBOARD_TESTING.md`
For technical analysis, see: `KEYBOARD_DEBUG_REPORT.md`
