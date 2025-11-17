# Virtual Keyboard Testing Guide

## Overview

This document provides comprehensive testing instructions for the virtual keyboard implementation in the Venture WASM build.

## What Was Fixed/Added

### 1. Enhanced Logging
- Added comprehensive console logging to `pkg/mobile/keyboard_wasm.go`
- All keyboard operations now log to browser console for debugging
- Logs include: initialization, focus events, character input, errors

### 2. Preset Name Fallback (Character Creation)
- Added 5 preset name buttons: "Warrior", "Mage", "Rogue", "Ranger", "Auto"
- "Auto" button generates a random name based on selected class
- Provides fallback if virtual keyboard doesn't work
- Only visible in WASM builds (mobile.IsWASM() check)

### 3. Keyboard Test Application
- Created standalone test app: `cmd/keyboardtest/main.go`
- Minimal UI to isolate keyboard functionality
- Real-time event logging display
- Browser diagnostic information

## Testing Instructions

### Test 1: Keyboard Test Application (Recommended First)

**Purpose:** Verify keyboard functionality in isolation

**Steps:**
1. Build the test app:
   ```bash
   ./scripts/build-keyboardtest.sh
   ```

2. Serve locally:
   ```bash
   cd build/keyboardtest
   python3 -m http.server 8080
   ```

3. Open on mobile device or desktop browser:
   - Desktop: `http://localhost:8080`
   - Mobile: `http://YOUR_IP:8080` (get IP with `hostname -I`)

4. **Testing procedure:**
   - Tap anywhere in the app to trigger keyboard
   - Browser console should show: `[VentureKeyboard] ShowKeyboard() called`
   - Type characters on keyboard (virtual or physical)
   - Console should show: `[VentureKeyboard] Input event: new chars added: 'X'`
   - Characters should appear in the text box
   - Press Enter to hide keyboard
   - Press Escape to clear text

5. **Check browser console** (F12 or mobile remote debugging):
   - Look for `[VentureKeyboard]` prefixed messages
   - Verify keyboard element is created
   - Check for focus success/failure messages

### Test 2: Full Game - Character Creation

**Purpose:** Test keyboard in actual game context with fallback options

**Steps:**
1. Build main game:
   ```bash
   make build-wasm
   ```

2. Serve locally:
   ```bash
   cd build/wasm
   python3 -m http.server 8080
   ```

3. Open `game.html` in browser (mobile or desktop)

4. Navigate to character creation

5. **Test keyboard input:**
   - Tap the name input field
   - Check console for `[VentureKeyboard] ShowKeyboard() called`
   - Type your character name
   - Verify text appears in input field

6. **Test fallback (if keyboard doesn't work):**
   - Tap one of the preset name buttons:
     - "Warrior" - sets name to "Warrior"
     - "Mage" - sets name to "Mage"
     - "Rogue" - sets name to "Rogue"
     - "Ranger" - sets name to "Ranger"
     - "Auto" - generates random name (e.g., "Braveblade")
   - Verify name is set correctly
   - Tap "Next" button to proceed

7. **Verify completion:**
   - You should be able to complete character creation even without virtual keyboard
   - Preset buttons provide alternative input method

### Test 3: Browser Console Debugging

**Purpose:** Diagnose keyboard issues via console logs

**Desktop Testing:**
1. Open browser DevTools (F12)
2. Navigate to Console tab
3. Load the game/test app
4. Look for these key messages:
   ```
   [VentureKeyboard] Initializing virtual keyboard element
   [VentureKeyboard] Virtual keyboard element created and added to DOM
   [VentureKeyboard] ShowKeyboard() called
   [VentureKeyboard] Keyboard element focused - mobile keyboard should appear
   [VentureKeyboard] Focus successful - active element is venture-keyboard-input
   [VentureKeyboard] Input event: new chars added: 'X'
   ```

**Mobile Testing (Remote Debugging):**

iOS Safari:
1. Enable Web Inspector on iPhone: Settings > Safari > Advanced > Web Inspector
2. Connect iPhone to Mac via USB
3. On Mac: Safari > Develop > [Your iPhone] > [Page]
4. View console logs in Mac Safari DevTools

Android Chrome:
1. Enable USB debugging on Android
2. Connect to computer via USB
3. On computer: Open Chrome, navigate to `chrome://inspect`
4. Find your device and click "inspect"
5. View console logs in Chrome DevTools

### Test 4: Cross-Browser Testing

**Purpose:** Verify compatibility across different browsers

Test on:
- ✅ Desktop Chrome (should work with physical keyboard)
- ✅ Desktop Firefox (should work with physical keyboard)
- ✅ Desktop Safari (should work with physical keyboard)
- ✅ iOS Safari (virtual keyboard test)
- ✅ Android Chrome (virtual keyboard test)

Expected results:
- Desktop: Physical keyboard input works
- Mobile: Virtual keyboard appears when tapping input field
- All: Preset buttons work as fallback
- All: Console logs appear showing keyboard operations

## Known Issues and Workarounds

### Issue 1: Keyboard Doesn't Appear on Mobile
**Symptoms:** Tapping doesn't show keyboard, console shows focus failed
**Workaround:** Use preset name buttons (Warrior, Mage, Rogue, Ranger, Auto)
**Debug:** Check console for error messages, verify element ID

### Issue 2: Characters Don't Appear
**Symptoms:** Keyboard appears but typing doesn't show text
**Workaround:** Use preset name buttons
**Debug:** Check console for input event logs, verify event forwarding

### Issue 3: Browser Security Blocks Focus
**Symptoms:** Console shows "Focus failed - active element is: CANVAS"
**Cause:** Some browsers require user gesture before programmatic focus
**Workaround:** Ensure ShowKeyboard() is called in response to touch/click event

## Success Criteria

✅ **Keyboard Test App:**
- Tap triggers ShowKeyboard() (verified in console)
- Keyboard element is created (verified in console)
- Focus successful (verified in console)
- Typing produces characters in text box
- Enter hides keyboard
- Escape clears text

✅ **Character Creation:**
- Tapping name field shows keyboard (or preset buttons work)
- Typing produces text in name field (or preset selection works)
- Can proceed to next step with valid name
- Character creation completable without physical keyboard

✅ **Browser Console:**
- No JavaScript errors
- Keyboard lifecycle logged correctly
- Input events logged when typing
- Focus state changes logged

## File Locations

- Virtual keyboard implementation: `pkg/mobile/keyboard_wasm.go`
- Character creation (with presets): `pkg/engine/character_creation.go`
- Keyboard test app: `cmd/keyboardtest/main.go`
- Test HTML: `cmd/keyboardtest/keyboardtest.html`
- Build script: `scripts/build-keyboardtest.sh`

## Reporting Issues

When reporting keyboard issues, include:
1. Browser and version (e.g., "iOS Safari 16.5")
2. Device (e.g., "iPhone 13")
3. Full browser console log
4. Screenshot of keyboard test app
5. Whether preset buttons work
6. Network tab showing resources loaded

## Next Steps

If keyboard still doesn't work after these tests:
1. Check browser console for errors
2. Verify wasm_exec.js is loaded
3. Test with different browser
4. Use preset name buttons as fallback
5. Consider browser-specific keyboard API differences
