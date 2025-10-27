# Aerial-View Sprite Migration Guide

**Version:** 1.0  
**Date:** October 26, 2025  
**Target Audience:** Developers integrating directional aerial-view sprites

---

## Table of Contents

1. [Overview](#overview)
2. [What Changed](#what-changed)
3. [Backward Compatibility](#backward-compatibility)
4. [Migration Steps](#migration-steps)
5. [Visual Comparison](#visual-comparison)
6. [Code Examples](#code-examples)
7. [Troubleshooting](#troubleshooting)
8. [Performance Considerations](#performance-considerations)

---

## Overview

The Character Avatar Enhancement introduces **aerial-view perspective sprites** optimized for top-down gameplay cameras. This enhancement provides:

- **4-directional facing** (Up, Down, Left, Right)
- **Aerial-view perspective** (top-down anatomical templates)
- **Automatic direction updates** from entity velocity
- **Genre-specific visual themes** (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
- **Boss scaling support** (preserves proportions and asymmetry)

### Why Aerial-View?

The original side-view sprites were designed for side-scrolling gameplay with vertical body proportions. Aerial-view sprites use **35/50/15 proportions** (head/torso/legs) optimized for overhead camera angles, providing better visual clarity in top-down games.

### Key Benefits

✅ **Automatic facing updates** - Movement system handles direction changes  
✅ **Zero manual coordination** - ECS integration is seamless  
✅ **Negligible performance impact** - 61.85 ns/op direction updates  
✅ **Genre consistency** - All 5 genres supported  
✅ **Boss scaling** - 2.5× scale with asymmetry preservation  
✅ **Backward compatible** - Optional via `UseAerial` flag  

---

## What Changed

### Architecture Changes

**New Components:**
- `AnimationComponent.Facing` - Current facing direction (0-3)
- `AnimationComponent.DirectionalImages` - Map of 4 directional sprites

**New Templates:**
- `HumanoidAerial()` - Base aerial template
- `FantasyHumanoidAerial()` - Fantasy genre
- `SciFiHumanoidAerial()` - Sci-fi genre
- `HorrorHumanoidAerial()` - Horror genre
- `CyberpunkHumanoidAerial()` - Cyberpunk genre
- `PostApocalypticHumanoidAerial()` - Post-apocalyptic genre
- `BossAerialTemplate(base, scale)` - Boss scaling function

**New Functions:**
- `GenerateDirectionalSprites(config)` - Generate 4-directional sprite sheet
- Direction enum: `DirUp`, `DirDown`, `DirLeft`, `DirRight`

**System Integration:**
- `MovementSystem` - Automatically updates `Facing` based on velocity
- `RenderSystem` - Syncs `CurrentDirection` with `Facing` before drawing

### File Changes

**Modified Files:**
- `pkg/rendering/sprites/anatomy_template.go` - Added 7 new template functions
- `pkg/rendering/sprites/generator.go` - Added `GenerateDirectionalSprites()`
- `pkg/engine/animation.go` - Added `Facing` and `DirectionalImages` fields
- `pkg/engine/movement.go` - Added automatic facing updates
- `pkg/engine/render_system.go` - Added direction sync logic

**New Files:**
- `pkg/engine/movement_direction_test.go` - Movement direction tests (439 lines)
- `pkg/rendering/sprites/generator_directional_test.go` - Sprite generation tests (374 lines)
- `pkg/rendering/sprites/aerial_validation_test.go` - Visual consistency tests (356 lines)
- `PHASE1_COMPLETE.md` through `PHASE6_COMPLETE.md` - Implementation summaries

---

## Backward Compatibility

### UseAerial Flag

The `UseAerial` flag in `GenerationConfig` controls sprite perspective:

```go
config := sprites.GenerationConfig{
    UseAerial: true,   // Aerial-view (new, recommended)
    UseAerial: false,  // Side-view (legacy, default)
}
```

**Default Behavior:** `UseAerial` defaults to `false` for backward compatibility. Existing code continues to work without changes.

### Gradual Migration

You can migrate incrementally:

1. **Phase 1**: Test aerial sprites alongside side-view sprites
2. **Phase 2**: Migrate specific entity types (e.g., humanoids only)
3. **Phase 3**: Enable globally via server/client config flags

### Breaking Changes

**None.** All changes are additive. The original side-view sprite generation remains fully functional.

---

## Migration Steps

### Step 1: Update Entity Generation (Client/Server)

**Before (Side-View):**

```go
// Original side-view sprite generation
result, err := gen.Generate(seed, procgen.GenerationParams{
    GenreID: "fantasy",
    Custom: map[string]interface{}{
        "width":  32,
        "height": 32,
        "type":   "humanoid",
    },
})
sprite := result.(*sprites.Sprite)
entity.AddComponent(&engine.SpriteComponent{
    Image: sprite.Image,
})
```

**After (Aerial-View with Directional Sprites):**

```go
// New aerial-view directional sprite generation
config := sprites.GenerationConfig{
    Width:      32,
    Height:     32,
    Seed:       seed,
    GenreID:    "fantasy",
    EntityType: "humanoid",
    UseAerial:  true,  // Enable aerial-view
}

directionalSprites, err := gen.GenerateDirectionalSprites(config)
if err != nil {
    return err
}

// Store in AnimationComponent
entity.AddComponent(engine.NewAnimationComponent(
    seed,  // Required seed parameter
    directionalSprites,
))

// Sprite component uses first direction initially
entity.AddComponent(&engine.SpriteComponent{
    Image: directionalSprites[sprites.DirDown],
})
```

### Step 2: Update Render System (If Custom)

If you have a custom render system, ensure it syncs direction before drawing:

```go
func (rs *CustomRenderSystem) Draw(entity *engine.Entity, screen *ebiten.Image) {
    sprite, ok := entity.GetComponent("sprite")
    if !ok {
        return
    }
    spriteComp := sprite.(*engine.SpriteComponent)
    
    // Sync direction with animation component
    anim, ok := entity.GetComponent("animation")
    if ok {
        animation := anim.(*engine.AnimationComponent)
        spriteComp.CurrentDirection = animation.Facing
        
        // Use directional sprite if available
        if animation.DirectionalImages != nil {
            if dirImage, exists := animation.DirectionalImages[animation.Facing]; exists {
                spriteComp.Image = dirImage
            }
        }
    }
    
    // Draw sprite
    screen.DrawImage(spriteComp.Image, opts)
}
```

**Note:** If using `engine.RenderSystem`, this is already handled automatically.

### Step 3: Verify Movement Integration

The movement system automatically updates facing. Verify your entities have required components:

```go
// Ensure entities have VelocityComponent for automatic facing
entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})

// Ensure entities have AnimationComponent with DirectionalImages
entity.AddComponent(engine.NewAnimationComponent(seed, directionalSprites))

// Movement system will automatically update Facing based on velocity
// No manual direction handling needed!
```

### Step 4: Test Directional Behavior

Run your game and verify:

- ✅ Character faces correct direction when moving
- ✅ Diagonal movement chooses horizontal direction (horizontal priority)
- ✅ Stationary entities preserve last facing direction
- ✅ Action states (attack/hit/death) don't change facing
- ✅ No flickering from velocity jitter (<0.1 threshold filtered)

### Step 5: Update Boss Entities (Optional)

For boss entities, use the boss scaling system:

```go
// Generate boss sprite with 2.5× scale
baseTemplate := sprites.FantasyHumanoidAerial()
bossTemplate := sprites.BossAerialTemplate(baseTemplate, 2.5)

config := sprites.GenerationConfig{
    Width:     64,  // Larger canvas for scaled boss
    Height:    64,
    Seed:      seed,
    GenreID:   "fantasy",
    EntityType: "boss",
    Template:  &bossTemplate,
    UseAerial: true,
}

bossSprites, err := gen.GenerateDirectionalSprites(config)
```

---

## Visual Comparison

### Side-View vs. Aerial-View

**Side-View (Legacy):**
```
┌─────────┐
│    O    │  Head: 30% (circular, centered)
│   /|\   │  Torso: 40% (vertical rectangle)
│    |    │  
│   / \   │  Legs: 30% (angled lines)
└─────────┘
- Vertical orientation
- Designed for side-scrolling
- Single facing direction
```

**Aerial-View (New):**
```
   Up (North)
┌─────────┐
│    ○    │  Head: 35% (offset upward)
│   ━╋━   │  Torso: 50% (main body)
│    ┃    │  
│   ╱ ╲   │  Legs: 15% (ground contact)
└─────────┘
- Top-down orientation
- Optimized for overhead camera
- 4 distinct directions
- Head offset for visual clarity
```

### Proportion Comparison

| Body Part | Side-View | Aerial-View | Change |
|-----------|-----------|-------------|--------|
| Head      | 30%       | 35%         | +5%    |
| Torso     | 40%       | 50%         | +10%   |
| Legs      | 30%       | 15%         | -15%   |
| **Total** | **100%**  | **100%**    | -      |

**Rationale:** Aerial-view emphasizes torso (main visual focus from above) and reduces leg visibility (minimal from overhead camera).

### Directional Asymmetry

Aerial templates use directional asymmetry for visual clarity:

- **Up (North)**: Head offset upward (Y: 0.35), arms positioned high
- **Down (South)**: Head centered (Y: 0.5), arms at sides
- **Left (West)**: Head offset left (X: 0.45), left arm forward
- **Right (East)**: Head offset right (X: 0.55), right arm forward

This ensures each direction is visually distinct even at small sprite sizes (16×16 pixels).

---

## Code Examples

### Example 1: Basic Aerial Sprite Generation

```go
package main

import (
    "log"
    "github.com/opd-ai/venture/pkg/rendering/sprites"
)

func main() {
    gen := sprites.NewGenerator()
    
    config := sprites.GenerationConfig{
        Width:      32,
        Height:     32,
        Seed:       42,
        GenreID:    "fantasy",
        EntityType: "humanoid",
        UseAerial:  true,
    }
    
    directionalSprites, err := gen.GenerateDirectionalSprites(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access sprites by direction
    upSprite := directionalSprites[sprites.DirUp]
    downSprite := directionalSprites[sprites.DirDown]
    leftSprite := directionalSprites[sprites.DirLeft]
    rightSprite := directionalSprites[sprites.DirRight]
    
    // Use in your game...
}
```

### Example 2: Genre-Specific Template

```go
// Use a specific genre template
horrorTemplate := sprites.HorrorHumanoidAerial()

config := sprites.GenerationConfig{
    Width:      32,
    Height:     32,
    Seed:       12345,
    GenreID:    "horror",
    EntityType: "humanoid",
    Template:   &horrorTemplate,  // Override default template
    UseAerial:  true,
}

sprites, err := gen.GenerateDirectionalSprites(config)
```

### Example 3: Boss with Custom Scale

```go
// Create boss with 3.0× scale
baseTemplate := sprites.SciFiHumanoidAerial()
bossTemplate := sprites.BossAerialTemplate(baseTemplate, 3.0)

config := sprites.GenerationConfig{
    Width:      96,   // 3× larger canvas
    Height:     96,
    Seed:       99999,
    GenreID:    "scifi",
    EntityType: "boss",
    Template:   &bossTemplate,
    UseAerial:  true,
}

bossSprites, err := gen.GenerateDirectionalSprites(config)
```

### Example 4: Complete Entity Setup

```go
// Complete entity creation with aerial sprites
func CreatePlayer(world *engine.World, x, y float64, seed int64) *engine.Entity {
    gen := sprites.NewGenerator()
    
    // Generate directional sprites
    config := sprites.GenerationConfig{
        Width:      32,
        Height:     32,
        Seed:       seed,
        GenreID:    "fantasy",
        EntityType: "humanoid",
        UseAerial:  true,
    }
    
    directionalSprites, err := gen.GenerateDirectionalSprites(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create entity
    entity := world.CreateEntity()
    
    // Add components
    entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
    entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
    entity.AddComponent(engine.NewAnimationComponent(seed, directionalSprites))
    entity.AddComponent(&engine.SpriteComponent{
        Image: directionalSprites[sprites.DirDown],  // Initial direction
    })
    entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
    
    world.AddEntity(entity)
    return entity
}
```

### Example 5: Custom Render Loop

```go
// Custom render loop with directional sprite support
func RenderEntities(world *engine.World, screen *ebiten.Image, camera *Camera) {
    for _, entity := range world.GetEntities() {
        pos, ok := entity.GetComponent("position")
        if !ok {
            continue
        }
        position := pos.(*engine.PositionComponent)
        
        sprite, ok := entity.GetComponent("sprite")
        if !ok {
            continue
        }
        spriteComp := sprite.(*engine.SpriteComponent)
        
        // Sync direction with animation
        anim, ok := entity.GetComponent("animation")
        if ok {
            animation := anim.(*engine.AnimationComponent)
            
            // Update sprite image based on facing direction
            if animation.DirectionalImages != nil {
                if dirImage, exists := animation.DirectionalImages[animation.Facing]; exists {
                    spriteComp.Image = dirImage
                    spriteComp.CurrentDirection = animation.Facing
                }
            }
        }
        
        // Calculate screen position
        screenX, screenY := camera.WorldToScreen(position.X, position.Y)
        
        // Draw sprite
        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Translate(screenX, screenY)
        screen.DrawImage(spriteComp.Image, opts)
    }
}
```

---

## Troubleshooting

### Issue: Character not changing direction

**Symptoms:** Character sprite doesn't update when moving

**Cause:** Entity missing `VelocityComponent` or `AnimationComponent`

**Solution:**
```go
// Ensure entity has both components
entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
entity.AddComponent(engine.NewAnimationComponent(seed, directionalSprites))
```

### Issue: Direction flickers during slow movement

**Symptoms:** Direction rapidly changes when entity is nearly stationary

**Cause:** Velocity jitter below threshold

**Solution:** Movement system already filters velocities < 0.1. If still occurring, check:
```go
// Ensure friction is applied correctly
entity.AddComponent(&engine.VelocityComponent{
    VX: vx,
    VY: vy,
    Friction: 0.85,  // Friction coefficient
})
```

### Issue: Wrong direction priority for diagonals

**Symptoms:** Diagonal movement shows unexpected direction

**Expected Behavior:** Horizontal priority - `|VX| >= |VY|` chooses left/right

**Solution:** This is by design. To change priority, modify `pkg/engine/movement.go`:
```go
// Current: Horizontal priority
if absVX >= absVY {
    // Choose left/right
}

// Alternative: Vertical priority
if absVY >= absVX {
    // Choose up/down
}
```

### Issue: Attack animation changes facing direction

**Symptoms:** Character turns during attack animation

**Cause:** Velocity not zeroed during action states

**Solution:**
```go
// When starting attack, zero velocity
entity.GetComponent("velocity").(*engine.VelocityComponent).VX = 0
entity.GetComponent("velocity").(*engine.VelocityComponent).VY = 0

// Movement system preserves facing when velocity < 0.1
```

### Issue: Boss sprite proportions look wrong

**Symptoms:** Boss sprite appears stretched or distorted

**Cause:** Canvas size doesn't match scale factor

**Solution:**
```go
// Ensure canvas size matches scale
scale := 2.5
baseSize := 32
canvasSize := int(float64(baseSize) * scale)  // 80 pixels

config := sprites.GenerationConfig{
    Width:  canvasSize,
    Height: canvasSize,
    // ... rest of config
}
```

### Issue: Aerial sprites not being used

**Symptoms:** Side-view sprites still appear

**Cause:** `UseAerial` flag not set or sprite component not updated

**Solution:**
```go
// Verify UseAerial is true
config := sprites.GenerationConfig{
    UseAerial: true,  // Must be explicit
    // ... rest of config
}

// Verify AnimationComponent has DirectionalImages
anim, ok := entity.GetComponent("animation")
if ok {
    animation := anim.(*engine.AnimationComponent)
    if animation.DirectionalImages == nil {
        log.Println("Warning: DirectionalImages not set")
    }
}
```

### Issue: Performance degradation

**Symptoms:** Frame rate drops after migration

**Cause:** Generating sprites every frame instead of caching

**Solution:**
```go
// Generate once during entity creation
directionalSprites, err := gen.GenerateDirectionalSprites(config)

// Store in AnimationComponent (cached)
entity.AddComponent(engine.NewAnimationComponent(seed, directionalSprites))

// Never regenerate during gameplay
// Direction switching uses map lookups (5ns overhead)
```

---

## Performance Considerations

### Generation Performance

**Single Sprite (Side-View):**
- Generation time: ~85 µs per sprite
- Memory: ~30 KB per sprite

**Directional Sprites (Aerial-View):**
- Generation time: ~172 µs for 4 sprites (43 µs per sprite)
- Memory: ~121 KB for 4-sprite sheet (30 KB per sprite)

**Recommendation:** Generate sprites once during entity creation, never during gameplay.

### Runtime Performance

**Direction Updates:**
- Movement system overhead: 61.85 ns per entity per frame
- Zero memory allocations
- Frame budget @ 60 FPS: 0.0004%

**Direction Switching:**
- Map lookup: <5 ns overhead
- No allocations
- Negligible impact

**Scalability:**
- 100 entities: 6.7 µs per frame (0.04% of frame budget)
- 1000 entities: 67 µs per frame (0.4% of frame budget)

### Memory Footprint

**Per Entity:**
- Side-view: 1 sprite × 30 KB = 30 KB
- Aerial-view: 4 sprites × 30 KB = 120 KB

**Memory Trade-off:**
- **4× memory usage** for 4-directional sprites
- **Zero runtime cost** for direction switching
- Cached sprites, no regeneration needed

**Recommendation:** For memory-constrained environments, consider:
- Generate only 2 directions (horizontal + vertical)
- Use smaller sprite sizes (16×16 instead of 32×32)
- Share sprites between similar entities

### Optimization Tips

1. **Batch sprite generation during loading screens**
   ```go
   // Generate all entity sprites during level load
   for _, entityDef := range level.Entities {
       sprites := generateDirectionalSprites(entityDef)
       entityCache[entityDef.ID] = sprites
   }
   ```

2. **Share sprites for identical entity types**
   ```go
   // Use same sprites for all goblins
   goblinSprites := generateOnce("goblin", seed)
   for _, goblinEntity := range goblins {
       entity.AddComponent(goblinSprites)
   }
   ```

3. **Profile before optimizing**
   ```bash
   go test -bench=. -benchmem -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   ```

---

## Testing Your Migration

### Validation Checklist

- [ ] All entity types generate aerial sprites correctly
- [ ] Characters face correct direction when moving
- [ ] Diagonal movement shows expected direction (horizontal priority)
- [ ] Stationary entities preserve facing direction
- [ ] Action states (attack/hit/death) don't change facing
- [ ] No flickering during slow movement (jitter filtering works)
- [ ] Boss entities scale correctly (proportions preserved)
- [ ] All 5 genres generate correctly (fantasy, sci-fi, horror, cyberpunk, post-apoc)
- [ ] Performance meets targets (<100ns direction updates)
- [ ] Memory usage acceptable (120KB per entity)

### Test Commands

```bash
# Run all directional sprite tests
go test ./pkg/rendering/sprites/ -run="Directional|Aerial|Boss" -v

# Run movement direction tests
go test ./pkg/engine/ -run="MovementSystem_Direction" -v

# Performance benchmarks
go test ./pkg/engine/ -bench="MovementSystem_DirectionUpdate" -benchmem
go test ./pkg/rendering/sprites/ -bench="GenerateDirectionalSprites" -benchmem

# Visual validation (requires X11)
go run ./cmd/rendertest/ --aerial --genre=fantasy
```

### Integration Testing

Create a test level with entities moving in all directions:

```go
func TestAerialSpriteIntegration(t *testing.T) {
    world := engine.NewWorld()
    world.AddSystem(engine.NewMovementSystem(200.0))
    
    // Create entity moving right
    entity := createTestEntity(world, 0, 0, 5.0, 0.0)  // VX=5, VY=0
    
    // Update world
    world.Update(0.016)
    
    // Verify facing direction
    anim, _ := entity.GetComponent("animation")
    animation := anim.(*engine.AnimationComponent)
    
    if animation.Facing != sprites.DirRight {
        t.Errorf("Expected DirRight, got %v", animation.Facing)
    }
}
```

---

## Additional Resources

- **API Reference**: See `docs/API_REFERENCE.md` for complete API documentation
- **Implementation Details**: See `PHASE1_COMPLETE.md` through `PHASE6_COMPLETE.md`
- **Architecture**: See `docs/ARCHITECTURE.md` for ECS design patterns
- **Performance**: See `docs/PERFORMANCE.md` for profiling guides

---

## Summary

**Migration is straightforward:**

1. Set `UseAerial: true` in `GenerationConfig`
2. Use `GenerateDirectionalSprites()` instead of `Generate()`
3. Store directional sprites in `AnimationComponent`
4. Let movement system handle facing updates automatically
5. Render system syncs direction before drawing

**No breaking changes.** Side-view sprites remain fully functional. Migrate at your own pace.

**Questions?** See `docs/CONTRIBUTING.md` for community support channels.

---

**Last Updated:** October 26, 2025  
**Document Version:** 1.0  
**Related:** Phase 7 - Documentation & Migration (Character Avatar Enhancement Plan)
