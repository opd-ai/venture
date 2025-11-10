# Testing the WASM Build Fix

## Quick Start

```bash
# Build and serve WASM version
make serve-wasm
```

Then open http://localhost:8080 in your browser.

## What to Test

### 1. Initial Load
- ✅ Main menu should appear without crash
- ✅ No errors in browser console (F12 → Console)

### 2. Character Creation Flow
1. Click "New Game" → Single Player → Select Genre
2. **Name Input Step**
   - Type a character name
   - Press ENTER to continue
   
3. **Class Selection Step**
   - Choose Warrior, Mage, or Rogue
   - Press ENTER or click to continue
   
4. **Portrait Step (Optional)**
   - Press TAB to skip (recommended for quick test)
   - Or press ENTER with empty input
   
5. **Confirmation Step**
   - Review your character
   - Press ENTER to begin adventure

### 3. Verify Gameplay
- ✅ Game transitions to dungeon view
- ✅ Player sprite renders (or colored rectangle initially)
- ✅ Player can move with WASD
- ✅ No crashes or errors

## Expected Behavior

### Before Fix
```
Character creation → Click "BEGIN ADVENTURE"
→ CRASH: "invalid memory address or nil pointer dereference"
→ Blank screen or error message
```

### After Fix
```
Character creation → Click "BEGIN ADVENTURE"
→ Smooth transition to gameplay
→ Player appears in dungeon
→ Can move and interact
```

## Browser Console Check

Open Developer Tools (F12) and check Console:
- ✅ No errors (red messages)
- ✅ May see info/debug messages (blue) - this is normal
- ✅ No "panic" or "undefined" errors

## Common Issues

### Issue: Blank screen after loading
**Solution:** 
- Check browser console for errors
- Try hard refresh (Ctrl+Shift+R)
- Ensure you're using a modern browser (Chrome, Firefox, Edge)

### Issue: Character creation UI not showing
**Solution:**
- Wait a few seconds for WASM to load
- Check browser console for load errors
- Verify build/wasm/venture.wasm exists and is ~19MB

### Issue: Input doesn't work
**Solution:**
- Click on the game canvas to focus it
- Try keyboard and mouse inputs
- On mobile, tap to show virtual keyboard

## Performance Notes

- First load may take 5-10 seconds (WASM is ~19MB)
- Subsequent loads are faster (browser cache)
- Character creation should be responsive
- Gameplay should run at 60 FPS

## Files to Check

If issues occur, verify these files exist:
```
build/wasm/
├── venture.wasm       (~19MB - the game binary)
├── wasm_exec.js       (~17KB - Go runtime)
├── index.html         (main page)
└── game.html          (fullscreen mode)
```

## Manual Test Checklist

- [ ] Main menu loads without crash
- [ ] Can navigate menus
- [ ] Character creation starts
- [ ] Name input accepts text
- [ ] Class selection works
- [ ] Portrait selection can be skipped
- [ ] Confirmation shows character summary
- [ ] "BEGIN ADVENTURE" button works
- [ ] Game transitions to dungeon view
- [ ] Player sprite appears
- [ ] Player can move with WASD
- [ ] No browser console errors
- [ ] No crashes throughout flow

## Success!

If all checks pass, the WASM fix is working correctly. The game should:
1. Load without errors
2. Complete character creation smoothly
3. Transition to gameplay seamlessly
4. Run without crashes or freezes

## Need Help?

- Check `WASM_FIX_SUMMARY.md` for technical details
- Review browser console for specific errors
- Verify WASM build with: `ls -lh build/wasm/venture.wasm`
- Rebuild if needed: `make build-wasm`
