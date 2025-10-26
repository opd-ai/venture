# Attack Animation Debug - Second Attack Not Working

## Problem Statement

User reports: "attacks and possibly other animations now trigger only the first time, and never trigger again."

## Expected Behavior (First Attack)

When pressing Space the FIRST time:

```
[PLAYER COMBAT] Entity 37 attack ready! Cooldown: 0.50, Timer: 0.00
[ANIM COMPONENT] SetState called: idle → attack (dirty will be: true)
[ANIM COMPONENT] Action animation - Loop=false, OnComplete will fire
[PLAYER COMBAT] Triggering attack animation (current state: attack)
[PLAYER COMBAT] Cooldown reset to 0.50s
[ANIMATION] Entity 37: Generating 6 frames for state=attack (sprite=28x28)

... 6 frames play over ~0.9 seconds (6 frames × 0.15s) ...

[ANIM SYSTEM] Animation complete - calling OnComplete callback (exists: true)
[ANIM COMPONENT] SetState called: attack → idle (dirty will be: true)
[ANIM COMPONENT] Movement animation - Loop=true
[ANIM SYSTEM] OnComplete callback executed
```

Then cooldown decrements:
```
[COMBAT SYSTEM] Entity 37 cooldown: 0.50 → 0.48 (delta: 0.022)
[COMBAT SYSTEM] Entity 37 cooldown: 0.48 → 0.46 (delta: 0.022)
... continues until 0.00 ...
```

## Expected Behavior (Second Attack)

When pressing Space the SECOND time (after cooldown expires):

```
[PLAYER COMBAT] Entity 37 attack ready! Cooldown: 0.50, Timer: 0.00
[ANIM COMPONENT] SetState called: idle → attack (dirty will be: true)
[ANIM COMPONENT] Action animation - Loop=false, OnComplete will fire
[PLAYER COMBAT] Triggering attack animation (current state: attack)
[PLAYER COMBAT] Cooldown reset to 0.50s
[ANIMATION] Entity 37: Generating 6 frames for state=attack (sprite=28x28)

... animation plays again ...
```

## Possible Failure Scenarios

### Scenario 1: Cooldown Not Resetting
**Symptom**: Second Space press shows:
```
[PLAYER COMBAT] Entity 37 attack on cooldown (0.50s remaining)
```

**Cause**: `attack.ResetCooldown()` not being called or not working
**Fix**: Verify ResetCooldown() implementation

### Scenario 2: Cooldown Not Decrementing
**Symptom**: No `[COMBAT SYSTEM]` cooldown messages after first attack
**Cause**: CombatSystem not running or player marked as dead
**Fix**: Check if player has "dead" component, verify CombatSystem in World.systems

### Scenario 3: Animation Not Resetting
**Symptom**: Second Space press shows:
```
[ANIM COMPONENT] SetState called: attack → attack (dirty will be: false)
```

**Cause**: State didn't change back to idle after first attack (OnComplete didn't fire)
**Fix**: Check OnComplete callback execution, verify Loop=false for attack animation

### Scenario 4: OnComplete Never Fires
**Symptom**: After first attack, no OnComplete messages:
```
[ANIM SYSTEM] Animation complete - calling OnComplete callback (exists: true)
[ANIM COMPONENT] SetState called: attack → idle (dirty will be: true)
[ANIM SYSTEM] OnComplete callback executed
```

**Cause**: Animation stuck in attack state, Playing=false but not calling callback
**Fix**: Verify animation reaches FrameIndex >= len(Frames), check Loop property

### Scenario 5: Input Not Being Captured
**Symptom**: No `[PLAYER COMBAT]` messages at all after first attack
**Cause**: Input system not registering Space key or input being consumed
**Fix**: Check InputSystem, verify IsActionPressed() returns true

## Testing Instructions

1. **Run the client**:
   ```bash
   ./client -seed 12345 -genre fantasy 2>&1 | tee attack_debug.log
   ```

2. **Wait for game to start** (see "Game initialized successfully" message)

3. **Press Space ONCE** - Watch for first attack sequence

4. **Wait 1 second** - Allow cooldown to expire and animation to complete

5. **Press Space AGAIN** - This is where the bug occurs

6. **Exit the game** (Ctrl+C)

7. **Analyze the log**:
   ```bash
   grep -E "\[PLAYER COMBAT\]|\[ANIM COMPONENT\]|\[ANIM SYSTEM\]|\[COMBAT SYSTEM\]" attack_debug.log
   ```

## Diagnostic Checklist

Compare the log output against these checkpoints:

**First Attack**:
- [ ] `[PLAYER COMBAT] Entity 37 attack ready!` appears
- [ ] `[ANIM COMPONENT] SetState called: idle → attack` appears
- [ ] `[ANIM COMPONENT] Action animation - Loop=false` appears
- [ ] `[ANIMATION] Entity 37: Generating 6 frames for state=attack` appears
- [ ] `[ANIM SYSTEM] Animation complete` appears after ~0.9s
- [ ] `[ANIM SYSTEM] OnComplete callback executed` appears
- [ ] `[ANIM COMPONENT] SetState called: attack → idle` appears
- [ ] `[COMBAT SYSTEM] Entity 37 cooldown:` messages appear counting down from 0.50 to 0.00

**Second Attack** (should match first attack):
- [ ] `[PLAYER COMBAT] Entity 37 attack ready!` appears again
- [ ] `[ANIM COMPONENT] SetState called: idle → attack` appears again
- [ ] Rest of sequence repeats

## Known Issues to Check

1. **SetState() Early Exit**: If `CurrentState == state`, SetState() does nothing. Verify state changes to idle after attack.

2. **OnComplete Overwriting**: Each Space press sets a NEW OnComplete callback. Previous callback is lost. Verify this doesn't cause issues.

3. **Playing Flag**: After OnComplete, Playing=false. Verify next SetState() sets Playing=true again.

4. **Dirty Flag**: After OnComplete and state change, Dirty=true should regenerate frames. Verify frames are regenerated.

5. **Dead Component**: If player somehow gets "dead" component, cooldowns stop updating. Check for warning message.

## Code References

**PlayerCombatSystem.Update()** (`pkg/engine/player_combat_system.go:26-115`):
- Line 55: Checks `CanAttack()` - if false, cooldown blocking
- Line 72: Sets attack state
- Line 76: Sets OnComplete callback
- Line 90: Calls `ResetCooldown()`

**CombatSystem.Update()** (`pkg/engine/combat_system.go:75-104`):
- Line 86: Checks for "dead" component
- Line 91: Calls `UpdateCooldown(deltaTime)`

**AnimationComponent.SetState()** (`pkg/engine/animation_component.go:106-136`):
- Line 112: Early exit if state unchanged
- Line 123: Sets Loop=false for action animations

**AnimationSystem.updateFrame()** (`pkg/engine/animation_system.go:84-107`):
- Line 100: Calls OnComplete when animation finishes

## Resolution Steps

Based on log analysis, apply appropriate fix:

1. **If cooldown stuck at 0.50s**: 
   - Add Playing=true to AnimationComponent.SetState()
   - Or verify CombatSystem is calling UpdateCooldown()

2. **If state stuck at "attack"**:
   - Verify OnComplete callback is being set and executed
   - Check Loop property is correctly set to false

3. **If no second Space press detected**:
   - Check InputSystem.Update() is running
   - Verify IsActionPressed() returns true on subsequent presses

4. **If SetState called but dirty=false**:
   - Remove `if a.CurrentState != state` check
   - Or ensure state changes to idle between attacks

## Success Criteria

Attack animation should trigger EVERY time Space is pressed (after cooldown expires), not just the first time. Log should show identical sequences for first, second, third, etc. attacks.
