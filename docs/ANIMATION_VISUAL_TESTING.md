# Animation System - Visual Testing Guide

**Purpose**: Verify that all recent animation system fixes are working correctly and visible in-game.

## Fixed Issues (Ready for Verification)

✅ **Issue 1**: Sprites morphing between frames → FIXED (single base sprite + transformations)  
✅ **Issue 2**: No directional facing → FIXED (velocity-based facing detection)  
✅ **Issue 3**: Attack animations not visible → FIXED (always trigger, no enemy requirement)  
✅ **Issue 4**: Animation amplitudes too subtle → FIXED (doubled all amplitudes to 200%)  
✅ **Issue 5**: Attacks only work once → FIXED (SetState same-state restart handling)

## How to Run the Test

```bash
cd /home/user/go/src/github.com/opd-ai/venture
./client -seed 12345 -genre fantasy
```

## What to Test

### Test 1: Directional Facing ✓

**What to do**:
1. Press **W** (move up) for 1-2 seconds
2. Press **S** (move down) for 1-2 seconds
3. Press **A** (move left) for 1-2 seconds
4. Press **D** (move right) for 1-2 seconds

**What to expect**:
- Sprite should visibly **turn to face the movement direction**
- Each direction should have distinct visual appearance:
  - **Up**: Character facing away/showing back
  - **Down**: Character facing toward camera/showing front
  - **Left**: Character facing left (may show left profile)
  - **Right**: Character facing right (may show right profile)

**What you're verifying**:
- Velocity-based facing detection works
- Direction changes are smooth and immediate
- Sprite appearance changes based on direction

**Failure signs**:
- ❌ Sprite always faces same direction regardless of movement
- ❌ No visual difference between up/down/left/right
- ❌ Sprite rotates instead of changing appearance

---

### Test 2: Idle Direction Persistence ✓

**What to do**:
1. Press **D** (move right) for 1 second
2. **Release D** and wait 2 seconds (idle)
3. Press **A** (move left) for 1 second
4. **Release A** and wait 2 seconds (idle)

**What to expect**:
- When moving right, sprite faces right
- When stopped after moving right, sprite **continues facing right** during idle
- When moving left, sprite faces left
- When stopped after moving left, sprite **continues facing left** during idle

**What you're verifying**:
- `LastFacing` field persists direction during idle state
- Sprite doesn't snap back to default "down" facing when stopped
- Idle animations respect last movement direction

**Failure signs**:
- ❌ Sprite always faces down when idle (ignores last direction)
- ❌ Direction resets after stopping

---

### Test 3: Attack Animation Visibility ✓

**What to do**:
1. Stand still (idle)
2. Press **Space** once
3. Observe the sprite carefully

**What to expect**:
- Sprite should perform a **visible forward lunge** of 16 pixels
- Sprite should **rotate** by 3.0 radians (about 172 degrees)
- Attack animation lasts 0.9 seconds (6 frames × 0.15s each)
- Animation should be **clearly visible** and **dramatic**

**What you're verifying**:
- Attack amplitudes are doubled (16px lunge, not 8px)
- Rotation is doubled (3.0 radians, not 1.5)
- Animation plays without requiring an enemy nearby
- Movement is noticeable, not subtle

**Failure signs**:
- ❌ Sprite barely moves (< 8 pixels)
- ❌ No visible rotation/swing
- ❌ Attack animation too subtle to notice
- ❌ Animation doesn't play at all

---

### Test 4: Attack Animation Responsiveness ✓ (CRITICAL - Fixes regression)

**What to do**:
1. Press **Space** once → wait 1 second
2. Press **Space** twice rapidly (tap-tap)
3. Press **Space** once → wait 1 second
4. Press **Space** three times rapidly (tap-tap-tap)
5. **Hold Space down** for 3 seconds

**What to expect**:
- **Every press triggers attack animation** (limited by 0.5s cooldown)
- Rapid presses should **restart the animation** each time
- Holding Space should repeat attacks every 0.5 seconds
- No stuck states or frozen animations
- Attacks continue working indefinitely (no "first time only" bug)

**What you're verifying**:
- SetState(attack → attack) restarts animation correctly
- Rapid input doesn't cause stuck states
- OnComplete callback fires reliably every time
- Cooldown system works (0.5s between attacks)

**Failure signs**:
- ❌ Only first attack works, subsequent attacks do nothing
- ❌ Sprite gets stuck in attack pose
- ❌ Animation doesn't restart on rapid presses
- ❌ Holding Space causes freeze or single attack

---

### Test 5: Walk Animation + Direction Changes ✓

**What to do**:
1. Press **D** and hold for 2 seconds (walk right)
2. **Without releasing D**, also press **W** (walk diagonal up-right)
3. Release **D**, continue holding **W** (walk up)
4. Release **W**

**What to expect**:
- Sprite shows walking animation (8px vertical bob)
- Direction changes are smooth
- Sprite faces the **primary direction** based on velocity magnitude
- Walking animation loops continuously while moving
- Returns to idle animation when stopped

**What you're verifying**:
- Walk animation plays during movement
- Direction detection works with diagonal movement
- Animation transitions smoothly: idle → walk → idle

**Failure signs**:
- ❌ No walking animation (sprite glides without bobbing)
- ❌ Direction doesn't update during diagonal movement
- ❌ Animation stutters or freezes during direction changes

---

### Test 6: Attack While Moving ✓

**What to do**:
1. Press **D** and hold (walk right)
2. While holding **D**, press **Space** (attack while moving)
3. Continue holding **D** after attack finishes

**What to expect**:
- Movement continues during attack input
- Attack animation plays (sprite lunges/rotates)
- After attack completes, **returns to walk animation** (not idle)
- Sprite continues facing right during and after attack

**What you're verifying**:
- Attack animations work during movement
- OnComplete callback detects velocity and returns to walk state
- Movement isn't cancelled by attacking
- State transitions: walk → attack → walk

**Failure signs**:
- ❌ Attack doesn't trigger while moving
- ❌ Returns to idle after attack (ignores velocity)
- ❌ Movement stops when attacking
- ❌ Sprite changes direction incorrectly

---

### Test 7: Animation Frame Consistency ✓

**What to do**:
1. Press **A** and hold (walk left)
2. Watch the sprite for 5 seconds
3. Stop, wait 3 seconds (idle)
4. Watch the sprite for 5 seconds

**What to expect**:
- Walking animation loops smoothly (8 frames)
- Sprite appearance remains **consistent across frames** (no morphing)
- Colors don't shift between frames
- Shapes don't change randomly
- Idle animation also loops consistently (4 frames)

**What you're verifying**:
- Single base sprite generation (no per-frame seed variation)
- Frame transformations only affect position/rotation/scale
- No color/shape morphing bug

**Failure signs**:
- ❌ Sprite "morphs" into different shapes/colors each frame
- ❌ Colors shift or flicker between frames
- ❌ Limbs/features appear/disappear randomly

---

## Expected System Behavior

### Animation States
- **Idle**: 4 frames, looping, minimal movement
- **Walk**: 8 frames, looping, 8px vertical bob
- **Run**: 8 frames, looping, faster bob
- **Attack**: 6 frames, non-looping, 16px forward lunge + 3.0 radian rotation
- **Hit**: 4 frames, non-looping, 12px knockback
- **Death**: 6 frames, non-looping, 20px fall

### Cooldown System
- Attack cooldown: **0.5 seconds**
- After pressing Space, must wait 0.5s before next attack triggers
- Visual feedback: None yet (HUD not implemented), but attacks won't trigger if cooldown active

### Direction Detection
- **Up**: abs(VY) > abs(VX) and VY < 0
- **Down**: abs(VY) > abs(VX) and VY > 0
- **Left**: abs(VX) > abs(VY) and VX < 0
- **Right**: abs(VX) > abs(VY) and VX > 0

### Performance Targets
- **60 FPS** minimum (check game window title for FPS)
- **No frame drops** during animation transitions
- **< 100ms** input latency (attack should feel instant)

---

## Debug Output (Optional)

If you want to see technical details, run with debug output:

```bash
./client -seed 12345 -genre fantasy 2>&1 | grep -E "\[PLAYER COMBAT\]|\[ANIM COMPONENT\]|\[ANIM SYSTEM\]"
```

**What you'll see**:
```
[PLAYER COMBAT] Entity 37 attack ready! Cooldown: 0.50, Timer: 0.00
[ANIM COMPONENT] SetState called: idle → attack (dirty will be: true)
[ANIMATION] Entity 37 facing: left (velocity: -100.0, 0.0)
[ANIM SYSTEM] OnComplete callback executed
[ANIM COMPONENT] SetState called: attack → idle (dirty will be: true)
[COMBAT SYSTEM] Entity 37 attack cooldown: 0.48
[COMBAT SYSTEM] Entity 37 attack cooldown: 0.46
...
[COMBAT SYSTEM] Entity 37 attack cooldown: 0.00
```

---

## Success Criteria

All tests should **PASS** with the following observed:

✅ Directional facing clearly visible (sprite changes based on direction)  
✅ Idle direction persists (sprite faces last movement direction when stopped)  
✅ Attack animation dramatic and visible (16px lunge, 3.0 radian rotation)  
✅ Attack animation works every time (no "first time only" bug)  
✅ Walk animation smooth with 8px vertical bob  
✅ Attack while moving returns to walk state (not idle)  
✅ Sprite frames consistent (no morphing/color shifting)  

---

## Reporting Issues

If any test **FAILS**, please report:

1. **Which test failed** (1-7)
2. **What you expected** (from "What to expect" section)
3. **What actually happened** (describe the behavior)
4. **Failure signs observed** (from "Failure signs" section)
5. **Screenshot or video** (if possible)

---

## Next Steps After Testing

1. **If all tests pass**: Proceed with debug logging cleanup (see docs/ATTACK_ANIMATION_FIX.md)
2. **If any tests fail**: Report findings and we'll investigate further
3. **Performance issues**: Run `go test -bench=. -benchmem ./pkg/rendering/sprites/` and report results

---

## Additional Notes

- **Genre**: Tests use `fantasy` genre, but animations work identically across all genres
- **Seed**: Using seed `12345` ensures reproducible results
- **Window size**: Default 800×600, can be changed with `-width` and `-height` flags
- **Exit**: Press ESC or close window to exit

---

**Last Updated**: 2025-10-25  
**Status**: All animation fixes complete, awaiting visual verification ✅
