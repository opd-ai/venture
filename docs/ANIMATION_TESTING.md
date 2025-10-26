# Animation System Manual Testing Guide

## Quick Start

```bash
cd /home/user/go/src/github.com/opd-ai/venture
./client -seed 12345 -genre fantasy
```

## What to Test

### 1. Directionality (NEW! ✅)

**Test**: Move with WASD keys  
**Expected**: Player sprite should face the direction you're moving
- Press **W** (up) → sprite faces up
- Press **S** (down) → sprite faces down  
- Press **A** (left) → sprite faces left
- Press **D** (right) → sprite faces right

**Test**: Stop moving (release all keys)  
**Expected**: Player sprite maintains the last facing direction during idle animation

### 2. Walk/Idle Animations ✅

**Test**: Walk around with WASD  
**Expected**: 8-frame walk animation cycles (should see subtle bobbing/stepping)

**Test**: Stop moving  
**Expected**: 4-frame idle animation cycles (should see subtle breathing/swaying)

### 3. Attack Animations

**Test**: Walk close to an enemy, press **Space**  
**Expected**: 
- Console output: `[PLAYER COMBAT] Entity 37 pressing attack button`
- Console output: `[ATTACK ANIM] Player attacking - setting state to ATTACK (was idle/walk)`
- Visual: Player sprite performs attack animation (6 frames)
- After animation: `[ATTACK ANIM] Player attack complete - returning to idle/walk`

**Troubleshooting**:
- If no `[PLAYER COMBAT]` message: Space key not being captured
- If message but no `[ATTACK ANIM]`: No enemy within range (50 pixels)
- If no visual change: Animations too subtle (see "If Animations Too Subtle" below)

### 4. Frame Generation

**Check Console Output**:
```
[ANIMATION] Entity 37: Generating 4 frames for state=idle (sprite=28x28)
[ANIMATION] Entity 37: Successfully generated 4 frames, now showing frame 0
```

This confirms frames are being generated with the anatomical template system.

### 5. State Transitions

**Test**: Walk, then stop, then walk again  
**Expected Console Output**:
```
[ANIMATION] Entity 37: Generating 8 frames for state=walk (sprite=28x28)
[ANIMATION] Entity 37: Generating 4 frames for state=idle (sprite=28x28)
[ANIMATION] Entity 37: Generating 8 frames for state=walk (sprite=28x28)
```

## If Animations Too Subtle

The animations are generated but may be hard to see on a 28x28 sprite. To increase visibility:

### Option 1: Increase Animation Amplitudes (Recommended)

Edit `pkg/engine/animation_system.go`:

**Current values** (lines ~207-300):
```go
case AnimationStateWalk:
    offset := math.Sin(progress * 2 * math.Pi) * 4  // ← Change 4 to 8
case AnimationStateAttack:
    offset := math.Sin(progress * math.Pi) * 8      // ← Change 8 to 16
```

**After changing, rebuild**:
```bash
go build -o client ./cmd/client
```

### Option 2: Increase Sprite Size

Edit `cmd/client/main.go` (line ~690):
```go
playerSprite := &engine.EbitenSprite{
    Image:   ebiten.NewImage(56, 56),  // ← Change from 28, 28
    Width:   56,                        // ← Change from 28
    Height:  56,                        // ← Change from 28
    Visible: true,
    Layer:   10,
}
```

**Also update collider** (line ~793):
```go
player.AddComponent(&engine.ColliderComponent{
    Width:     56,    // ← Match sprite size
    Height:    56,    // ← Match sprite size
    Solid:     true,
    IsTrigger: false,
    Layer:     1,
    OffsetX:   -28,  // ← Width/2
    OffsetY:   -28,  // ← Height/2
})
```

### Option 3: Zoom In

Use camera zoom (if implemented) or reduce window size to make sprites appear larger.

## Console Output Reference

### Normal Operation
```
INFO[...] Starting Venture - Procedural Action RPG
INFO[...] client configuration
INFO[...] single-player mode
INFO[...] game initialized
INFO[...] BSP terrain generation complete
[ANIMATION] Entity 37: Generating 4 frames for state=idle (sprite=28x28)
[ANIMATION] Entity 37: Successfully generated 4 frames, now showing frame 0
```

### Walking Around
```
[ANIMATION] Entity 37: Generating 8 frames for state=walk (sprite=28x28)
[ANIMATION] Entity 37: Successfully generated 4 frames, now showing frame 0
```

(Note: No per-frame spam - only logs when state changes)

### Attacking Enemy
```
[PLAYER COMBAT] Entity 37 pressing attack button
[ATTACK ANIM] Player attacking - setting state to ATTACK (was idle)
[ANIMATION] Entity 37: Generating 6 frames for state=attack (sprite=28x28)
[ANIMATION] Entity 37: Successfully generated 6 frames, now showing frame 0
[ATTACK ANIM] Player attack complete - returning to idle/walk
[ANIMATION] Entity 37: Generating 4 frames for state=idle (sprite=28x28)
```

### Getting Hit
```
[ATTACK ANIM] Entity 37 hit - setting state to HIT (was idle)
[ANIMATION] Entity 37: Generating 4 frames for state=hit (sprite=28x28)
[ATTACK ANIM] Entity 37 hit complete - returning to idle/walk
```

## Performance Check

Open a terminal and run:
```bash
./client -seed 12345 -genre fantasy 2>&1 | grep "deltaTime" | head -20
```

**Expected**: deltaTime values between 0.011-0.023 (corresponding to ~45-90 FPS)

If deltaTime > 0.033 (below 30 FPS), performance optimization may be needed.

## Debug Mode

For more verbose output:
```bash
./client -verbose -seed 12345 -genre fantasy
```

This enables additional logging from all systems.

## Common Issues

### Issue: "Can't see any animation"
**Check**: 
1. Console shows `[ANIMATION] Entity 37: Generating X frames` messages?
2. If yes: Animations working but may be too subtle (see "If Animations Too Subtle")
3. If no: Animation system not running (file a bug report)

### Issue: "No directionality - sprite doesn't turn"
**Check**:
1. Facing detection needs velocity > 0.1 to trigger
2. Try walking for at least 1 second in each direction
3. Verify `config.Custom["facing"]` is being set in buildSpriteConfig()

### Issue: "Attack animation doesn't play"
**Check**:
1. Console shows `[PLAYER COMBAT]` message when pressing Space?
   - If no: InputSystem not capturing Space key
2. Console shows `[ATTACK ANIM] Player attacking`?
   - If no: No enemy within range (walk closer, or increase range in main.go line 785)
3. Messages appear but no visual change?
   - Animations too subtle (increase amplitudes)

### Issue: "Game crashes / panic"
**Check**:
1. Run with error output: `./client -seed 12345 -genre fantasy 2>&1 | grep -i "panic\|error"`
2. File bug report with full stack trace

## Success Criteria

✅ **System Working** if you see:
- Frame generation messages on state changes
- Sprite faces different directions when moving different ways
- Idle animation plays when standing still
- Walk animation plays when moving
- Attack animation plays when pressing Space near enemy
- No errors/panics in console

⚠️ **Needs Adjustment** if:
- Animations too subtle to notice
- Facing directions not distinct enough
- Attack animation too fast/slow

## Next Steps After Testing

1. Report results: Which animations work, which don't
2. If too subtle: We'll increase amplitudes or sprite size
3. If directionality issues: We'll enhance the template system
4. If attack issues: We'll debug the combat→animation pipeline

## Files to Watch

Animation implementation:
- `pkg/engine/animation_system.go` - Frame generation and updates
- `pkg/engine/animation_component.go` - Animation state and data
- `pkg/rendering/sprites/animation.go` - Frame transformation logic
- `pkg/engine/movement.go` - Walk/idle animation triggers
- `pkg/engine/combat_system.go` - Attack/hit animation triggers

Player setup:
- `cmd/client/main.go` - Player entity creation and components (lines 657-804)
