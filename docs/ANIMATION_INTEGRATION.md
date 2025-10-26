# Animation System Integration Guide

## Overview

The Venture animation system provides frame-based sprite animations that integrate seamlessly with gameplay systems. Animations respond automatically to player actions, creating a dynamic visual experience without manual intervention.

## Architecture

### Components

**AnimationComponent** (`pkg/engine/components.go`)
- Stores animation state, frame data, and playback information
- Tracks current frame index, elapsed time, and loop settings
- Maintains cached sprite frames for efficient rendering
- Supports animation callbacks for state transitions

**AnimationSystem** (`pkg/engine/animation_system.go`)
- Updates all entities with `AnimationComponent` each frame
- Manages frame transitions based on frame time and delta time
- Implements LRU caching (100 sequences) to optimize frame generation
- Regenerates frames when animation parameters change (marked "dirty")

**Frame Generation** (`pkg/rendering/sprites/animation.go`)
- Generates individual animation frames with state-specific transformations
- Applies offsets, rotations, and scaling based on animation state
- Creates smooth transitions between frames using interpolation

### Animation States

The system supports 10 animation states:

| State | Trigger | Visual Effect |
|-------|---------|---------------|
| **Idle** | No movement or action | Minimal bobbing, default pose |
| **Walk** | Slow movement (< 70% max speed) | Vertical bobbing, gentle sway |
| **Run** | Fast movement (>= 70% max speed) | Increased bobbing, faster cycle |
| **Attack** | Combat attack action | Forward thrust, weapon swing |
| **Cast** | Spell casting | Raised arms, magical gesture |
| **Hit** | Taking damage | Knockback, recoil animation |
| **Death** | Health reaches 0 | Fall, rotation, fade |
| **Jump** | Jumping action | Parabolic arc, compressed landing |
| **Crouch** | Crouching | Reduced height, defensive pose |
| **Use** | Using items | Item interaction motion |

## Integration Points

### Movement System Integration

**File**: `pkg/engine/movement.go`  
**Lines**: 162-184 (added in Phase 5 integration)

The `MovementSystem.Update()` method analyzes entity velocity to determine the appropriate animation state:

```go
// Update animation state based on movement
if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
    anim := animComp.(*AnimationComponent)
    speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
    
    if speed > 0.1 {
        if speed > s.MaxSpeed*0.7 && s.MaxSpeed > 0 {
            // Fast movement - running
            if anim.CurrentState != AnimationStateRun {
                anim.SetState(AnimationStateRun)
            }
        } else {
            // Normal movement - walking
            if anim.CurrentState != AnimationStateWalk {
                anim.SetState(AnimationStateWalk)
            }
        }
    } else {
        // Not moving - idle
        if anim.CurrentState == AnimationStateWalk || anim.CurrentState == AnimationStateRun {
            anim.SetState(AnimationStateIdle)
        }
    }
}
```

**Behavior**:
- Velocity magnitude < 0.1: Switch to `AnimationStateIdle`
- Velocity < MaxSpeed * 0.7: Switch to `AnimationStateWalk`
- Velocity >= MaxSpeed * 0.7: Switch to `AnimationStateRun`
- Only transitions from Walk/Run states to Idle (preserves other animations like Attack)

### Combat System Integration

**File**: `pkg/engine/combat_system.go`  
**Lines**: 270-283 (added in Phase 5 integration)

The `CombatSystem.Attack()` method triggers attack and hit animations:

```go
// Trigger attack animation for attacker
if animComp, hasAnim := attacker.GetComponent("animation"); hasAnim {
    anim := animComp.(*AnimationComponent)
    anim.SetState(AnimationStateAttack)
}

// Trigger hurt animation for target
if animComp, hasAnim := target.GetComponent("animation"); hasAnim {
    anim := animComp.(*AnimationComponent)
    anim.SetState(AnimationStateHit)
    // Set callback to return to idle after hurt animation
    anim.OnComplete = func() {
        anim.SetState(AnimationStateIdle)
    }
}
```

**Behavior**:
- Attacker: Plays attack animation (forward thrust)
- Target: Plays hit animation (knockback), then auto-returns to idle via callback
- Attack animations override movement animations during combat
- Hit animation completion callback ensures smooth transition back to idle

### Client Integration

**File**: `cmd/client/main.go`  
**Lines**: 293 (instantiation), 525 (system registration)

The client creates and registers the animation system at startup:

```go
// Create animation system
animationSystem := engine.NewAnimationSystem(spriteGenerator)

// Register with world via wrapper
type animationSystemWrapper struct {
    *engine.AnimationSystem
}

func (w animationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) error {
    return w.AnimationSystem.Update(entities, deltaTime)
}

world.AddSystem(animationSystemWrapper{animationSystem})
```

## Frame Caching

The animation system implements an LRU (Least Recently Used) cache with 100 sequence capacity to optimize performance:

**Cache Key Format**: `seed_state_width_height_frameIndex`

**Cache Behavior**:
- First access: Generate frames on-the-fly, store in cache
- Subsequent access: Retrieve pre-generated frames instantly
- Cache full: Evict least recently used sequence
- State change: Mark animation "dirty", regenerate on next update

**Performance Impact**:
- Initial frame generation: ~5-10ms per animation sequence
- Cached frame access: <1ms per frame
- Cache hit rate: 95%+ in typical gameplay
- Memory usage: ~50-100MB for 100 cached sequences

## Adding Animation Triggers

To add animation triggers to other gameplay systems:

### 1. Identify the Action Point

Find where the action occurs in the code (e.g., spell casting, item usage, death).

### 2. Check for AnimationComponent

Use the component pattern to safely check if the entity has animations:

```go
if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
    anim := animComp.(*AnimationComponent)
    // Trigger animation
}
```

### 3. Set Animation State

Choose the appropriate state and set it:

```go
anim.SetState(engine.AnimationStateCast)  // For spell casting
anim.SetState(engine.AnimationStateUse)   // For item usage
anim.SetState(engine.AnimationStateDeath) // For death
```

### 4. Add Completion Callback (Optional)

For animations that should transition to another state:

```go
anim.SetState(engine.AnimationStateHit)
anim.OnComplete = func() {
    anim.SetState(engine.AnimationStateIdle)
}
```

### Example: Spell Casting Integration

```go
// In magic system's CastSpell() method
func (s *MagicSystem) CastSpell(caster *Entity, spell *Spell) error {
    // Trigger cast animation
    if animComp, hasAnim := caster.GetComponent("animation"); hasAnim {
        anim := animComp.(*AnimationComponent)
        anim.SetState(AnimationStateCast)
        
        // Return to idle after casting
        anim.OnComplete = func() {
            anim.SetState(AnimationStateIdle)
        }
    }
    
    // ... rest of spell casting logic
    return nil
}
```

## Testing Animation Integration

### Animation Demo Tool

Run the standalone animation demo to visualize all animation states:

```bash
cd /home/user/go/src/github.com/opd-ai/venture
./animationdemo -seed 12345 -genre fantasy
```

**Controls**:
- `1` - Idle animation demo
- `2` - Walk/Run animation demo
- `3` - Attack/Hit animation demo
- `H` - Toggle help
- `ESC` - Quit

**Demo Modes**:
1. **Idle Demo**: Both entities remain stationary, showing idle animation
2. **Walk Demo**: Entities move in patterns, demonstrating walk/run transitions
3. **Attack Demo**: Entities attack each other, showing attack/hit animations

### In-Game Testing

Test animations in the actual game client:

```bash
cd /home/user/go/src/github.com/opd-ai/venture
./client
```

**Test Cases**:
1. **Movement**: Walk around, observe walk animation. Sprint (if implemented), observe run animation.
2. **Combat**: Attack enemies, verify attack animation plays. Get hit, verify hit animation plays.
3. **State Transitions**: Ensure smooth transitions between idle, walk, run, and attack states.
4. **Callback Behavior**: Verify hit animation returns to idle after completion.

## Performance Considerations

### Frame Generation Cost

- **Initial Generation**: 5-10ms per animation sequence (8 frames for Walk, 6 for Attack)
- **Frame Update**: <1ms per entity per frame
- **Cache Hit**: <0.1ms per entity per frame

### Optimization Strategies

1. **Pre-warm Cache**: Generate common animations during loading screen
2. **Batch Updates**: AnimationSystem processes all entities in single pass
3. **Dirty Marking**: Only regenerate frames when parameters change
4. **LRU Eviction**: Automatically manage memory usage

### Memory Usage

- **Per Animation Sequence**: 500KB - 1MB (8 frames @ 32x32 sprites, scaled 3x)
- **Cache Capacity**: 100 sequences = 50-100MB
- **Entity Component**: 200-400 bytes per entity (metadata only, frames cached separately)

## Troubleshooting

### Animation Not Playing

**Symptom**: Entity sprite doesn't animate  
**Causes**:
1. Entity missing `AnimationComponent`
2. AnimationSystem not registered with world
3. Animation marked not playing (`Playing = false`)

**Solution**:
```go
// Check component exists
if _, hasAnim := entity.GetComponent("animation"); !hasAnim {
    log.Println("Entity missing AnimationComponent")
}

// Check system registered
// In client main.go, verify world.AddSystem(animationSystemWrapper{...})

// Check playing flag
anim.Playing = true
```

### Wrong Animation Playing

**Symptom**: Incorrect animation for action  
**Causes**:
1. State not set correctly
2. Multiple systems competing for state
3. Callback overriding state

**Solution**:
```go
// Log state changes for debugging
anim.SetState(AnimationStateAttack)
log.Printf("Set animation state to %s", anim.CurrentState)

// Check for competing state changes
// Search codebase for SetState calls affecting same entity
```

### Animation Stuttering

**Symptom**: Choppy or frozen animations  
**Causes**:
1. Frame time too high
2. Delta time issues (large jumps)
3. Cache eviction (regenerating frequently)

**Solution**:
```go
// Adjust frame time for smoother playback
anim.FrameTime = 0.08  // 12.5 FPS
anim.FrameTime = 0.05  // 20 FPS (smoother)

// Ensure delta time clamping
if deltaTime > 0.1 {
    deltaTime = 0.1
}

// Check cache statistics
cacheSize := animationSystem.GetCacheSize()
log.Printf("Animation cache size: %d/100", cacheSize)
```

### Memory Leaks

**Symptom**: Memory usage grows over time  
**Causes**:
1. Entities not cleaned up properly
2. Callbacks holding entity references
3. Cache not evicting

**Solution**:
```go
// Clear callbacks when removing entities
anim.OnComplete = nil

// Verify cache eviction
// AnimationSystem automatically evicts LRU entries at 100 capacity

// Monitor cache size over time
if cacheSize > 150 {
    log.Println("WARNING: Animation cache exceeding capacity")
}
```

## Future Enhancements

### Planned Features

1. **Animation Blending**: Smooth transitions between states (e.g., walk → attack)
2. **Layered Animations**: Separate upper/lower body animations (walk while casting)
3. **IK System**: Inverse kinematics for foot placement on terrain
4. **Animation Events**: Trigger sound effects at specific frames (footsteps, sword swings)
5. **Directional Sprites**: Different sprite directions (8-way movement)

### Integration Roadmap

- **Phase 8.2**: Input & Rendering - Additional visual polish
- **Phase 8.3**: Save/Load - Persist animation state in save files
- **Phase 8.4**: Performance - Optimize frame generation and caching
- **Phase 8.5**: Tutorial - In-game animation system documentation

## References

- **AnimationComponent**: `pkg/engine/components.go` lines 208-250
- **AnimationSystem**: `pkg/engine/animation_system.go` lines 1-260
- **Frame Generation**: `pkg/rendering/sprites/animation.go` lines 1-150+
- **Movement Integration**: `pkg/engine/movement.go` lines 162-184
- **Combat Integration**: `pkg/engine/combat_system.go` lines 270-283
- **Client Setup**: `cmd/client/main.go` lines 293, 525
- **Demo Tool**: `cmd/animationdemo/main.go`
