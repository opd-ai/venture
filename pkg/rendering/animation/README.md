# Animation Fluidity System - Phase 46

The animation package provides advanced 8-frame animation with 8-directional support, body part articulation, and intelligent caching for the Venture game engine.

## Features

### 8-Frame Animation Cycles
- Smooth animation at 60 FPS with 8 frames per cycle
- State-specific frame counts (idle: 8, walk: 8, run: 8, attack: 8, hit: 4, death: 8, jump: 6)
- Configurable frame times (idle: 8 FPS, walk: 12 FPS, run: 16 FPS, attack: 16 FPS)

### 8-Directional Movement
- Full 8-direction support: N, NE, E, SE, S, SW, W, NW
- Automatic direction calculation from velocity vectors
- Smooth direction transitions for natural movement
- Conversion utilities for legacy 4-direction systems

### Body Part Articulation
- Precise pixel-level articulation constraints:
  - Arms: ±3px offset, ±0.3 radians rotation
  - Legs: ±4px offset, ±0.4 radians rotation
  - Head: ±2px offset, ±0.2 radians rotation
  - Tail: ±5px offset, ±0.5 radians rotation (for quadrupeds)
- State-specific articulation patterns:
  - Idle: Subtle breathing animation
  - Walk: Natural arm/leg swing with opposite phase
  - Run: Exaggerated motion (1.5x amplitude)
  - Attack: Wind-up → strike → follow-through
  - Hit: Knockback with recoil
  - Death: Falling/collapsing with rotation
  - Cast: Spell-casting gesture
  - Jump: Crouch → jump arc → landing

### Animation Caching
- LRU (Least Recently Used) cache eviction
- Configurable size limits (default: 50MB, 1000 entries)
- Target ≥85% cache hit rate
- Thread-safe concurrent access
- Cache statistics and performance metrics
- Dynamic trimming for memory management

### Pre-computation
- Pre-generate common animation sequences
- Front-load generation work for runtime performance
- Configurable seed lists for player and enemy animations
- Common states: idle, walk, run, attack
- Primary directions: N, E, S, W (diagonals generated on-demand)

### Frame Interpolation
- Sub-frame smoothness for 60 FPS rendering
- Alpha-blended interpolation between frames
- Enables smooth motion even with 8-12 FPS animations

## Usage

### Basic Animation

```go
import (
    "github.com/opd-ai/venture/pkg/rendering/animation"
    "github.com/opd-ai/venture/pkg/rendering/sprites"
)

// Create controller
gen := sprites.NewGenerator()
controller := animation.NewController(gen)

// Configure sprite
config := sprites.Config{
    Width:  64,
    Height: 64,
    Seed:   12345,
}

// Generate animation frame
frame, err := controller.GenerateFrame(
    12345,                    // seed
    "walk",                   // state
    0,                        // frame index
    8,                        // frame count
    animation.Dir8North,      // direction
    config,                   // sprite config
)
```

### Pre-computation

```go
// Pre-compute common animations for faster runtime
seeds := []int64{12345, 67890} // Player and enemy seeds
controller.PrecomputeCommon(seeds, config)

// Later frames will be served from cache
```

### Direction Calculation

```go
// Calculate direction from velocity
vx, vy := 1.0, -1.0
direction := animation.FromVelocity(vx, vy) // Returns Dir8NorthEast

// Convert to legacy 4-direction
dir4 := direction.To4Direction() // Returns "up"
```

### Custom Articulation

```go
// Create custom articulation config
config := animation.ArticulationConfig{
    ArmOffsetMax:    5.0,  // Increased arm range
    LegOffsetMax:    6.0,  // Increased leg range
    HeadOffsetMax:   3.0,
    TailOffsetMax:   7.0,
    ArmRotationMax:  0.5,
    LegRotationMax:  0.6,
    HeadRotationMax: 0.3,
    TailRotationMax: 0.7,
}
controller.SetArticulationConfig(config)
```

### Performance Monitoring

```go
// Get performance metrics
metrics := controller.GetPerformanceMetrics()
fmt.Printf("Cache Hit Rate: %.1f%%\n", metrics.CacheHitRate)
fmt.Printf("Cache Size: %d bytes\n", metrics.CacheSize)
fmt.Printf("Frame Generation Time: %v\n", metrics.FrameGenerationTime)
```

## Package Structure

- `direction.go` - 8-directional movement system
- `articulation.go` - Body part articulation calculations
- `cache.go` - LRU animation frame cache
- `controller.go` - Main animation controller
- `doc.go` - Package documentation

## Performance

- Direction calculation: ~1µs per call
- Articulation calculation: <10µs per call
- Cache get operations: <1µs per call
- Target: 60 FPS with 100 animated entities
- Cache hit rate target: ≥85%

## Testing

Run tests with:
```bash
go test ./pkg/rendering/animation/...
```

Run with coverage:
```bash
go test -cover ./pkg/rendering/animation/...
```

Current coverage: **71.6%** (exceeds 65% minimum requirement)

## Examples

See `examples/animation_fluidity_demo/` for a complete interactive demonstration.

## Integration

The animation system integrates with the existing `engine.AnimationComponent` but can also be used standalone. It's designed to work with the procedural sprite generation system in `pkg/rendering/sprites/`.

## License

Part of the Venture game engine. See LICENSE file in repository root.
