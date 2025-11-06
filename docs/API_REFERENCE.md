# Venture API Reference

Developer documentation for the Venture procedural action-RPG engine.

**Version:** 1.0  
**Last Updated:** October 2025

**New to development?** Start with [Development Guide](DEVELOPMENT.md) and [Contributing Guide](CONTRIBUTING.md).

---

## Table of Contents

1. [Core Engine](#core-engine)
2. [Entity-Component-System](#entity-component-system)
3. [Procedural Generation](#procedural-generation)
4. [Rendering System](#rendering-system)
5. [Audio System](#audio-system)
6. [Networking](#networking)
7. [Save/Load System](#saveload-system)
8. [UI Systems](#ui-systems)

---

## Core Engine

### Package: `github.com/opd-ai/venture/pkg/engine`

The engine package provides the ECS framework and core game systems.

#### World

Central ECS container managing entities and systems.

```go
world := engine.NewWorld()
world.AddSystem(engine.NewMovementSystem(200.0))
world.Update(0.016) // 60 FPS
entities := world.GetEntitiesWith("position", "health")
```

**Key Methods:** `NewWorld()`, `AddSystem()`, `Update()`, `GetEntities()`, `AddEntity()`, `CreateEntity()`, `GetEntitiesWith()`

#### Entity

Container for components.

```go
entity := engine.NewEntity(123)
entity.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
pos, ok := entity.GetComponent("position")
hasVelocity := entity.HasComponent("velocity")
```

**Key Methods:** `NewEntity()`, `AddComponent()`, `GetComponent()`, `HasComponent()`, `RemoveComponent()`

#### Component

Pure data structures implementing `Type() string`.

**Core Components:**
- `PositionComponent` - X, Y coordinates
- `VelocityComponent` - VX, VY velocity
- `ColliderComponent` - Collision box
- `HealthComponent`, `ManaComponent` - Resources
- `StatsComponent` - RPG stats (Attack, Defense, etc.)
- `InventoryComponent`, `EquipmentComponent` - Item management
- `AIComponent` - AI state machine
- `NetworkComponent` - Network sync

**Visual/Audio:**
- Sprite (via `SpriteProvider` interface)
- `ParticleEmitterComponent`, `AnimationComponent`

**See full component list in source:** `pkg/engine/*.go`

#### System

Game logic processors.

```go
type MySystem struct {}
func (s *MySystem) Update(entities []*engine.Entity, deltaTime float64) {
    // Process entities with required components
}
world.AddSystem(&MySystem{})
```

**Core Systems:** MovementSystem, CollisionSystem, CombatSystem, AISystem, ProgressionSystem, InventorySystem, InputSystem, RenderSystem, CameraSystem, ParticleSystem

**See system constructors for parameters.**

---

## Entity-Component-System

### ECS Pattern

**Entities:** IDs with component collections  
**Components:** Data only, no logic  
**Systems:** Logic only, process entities

### Component Examples

```go
type PositionComponent struct {
    X, Y float64
}
func (p PositionComponent) Type() string { return "position" }

type HealthComponent struct {
    Current, Max float64
}
func (h HealthComponent) Type() string { return "health" }
```

### System Examples

```go
// MovementSystem applies velocity
func (m *MovementSystem) Update(entities []*Entity, dt float64) {
    for _, e := range entities {
        if !e.HasComponent("position") || !e.HasComponent("velocity") { continue }
        pos := e.GetComponent("position").(*PositionComponent)
        vel := e.GetComponent("velocity").(*VelocityComponent)
        pos.X += vel.VX * dt
        pos.Y += vel.VY * dt
    }
}
```

---

## Procedural Generation

### Package: `github.com/opd-ai/venture/pkg/procgen`

All generators implement:
```go
type Generator interface {
    Generate(seed int64, params GenerationParams) (interface{}, error)
    Validate(result interface{}) error
}
```

### GenerationParams

```go
type GenerationParams struct {
    Difficulty float64  // 0.0-1.0
    Depth      int      // Dungeon level
    GenreID    string   // "fantasy", "scifi", etc.
    Custom     map[string]interface{}
}
```

### Terrain Generation

**Package:** `pkg/procgen/terrain`

```go
gen := terrain.NewGenerator()
result, _ := gen.Generate(12345, procgen.GenerationParams{
    Difficulty: 0.5,
    Depth: 5,
    GenreID: "fantasy",
})
t := result.(*terrain.Terrain)
```

**Room Types:** Start, Boss, Treasure, Shop, Standard, Shrine, Library, Trap

### Entity Generation

**Package:** `pkg/procgen/entity`

Generates monsters, NPCs, bosses with genre-themed names and stats.

```go
gen := entity.NewGenerator()
result, _ := gen.Generate(seed, params)
e := result.(*entity.Entity) // Type, Stats, Behavior
```

### Item Generation

**Package:** `pkg/procgen/item`

Weapons, armor, consumables with rarity system (Common, Uncommon, Rare, Epic, Legendary).

```go
gen := item.NewGenerator()
result, _ := gen.Generate(seed, params)
item := result.(*item.Item)
```

### Magic & Skills

**Packages:** `pkg/procgen/magic`, `pkg/procgen/skills`

Spell generation with elemental types, skill trees with prerequisites.

### Quest & Environment

**Packages:** `pkg/procgen/quest`, `pkg/procgen/environment`

Quest objectives/rewards, environmental effects/ambience.

---

## Rendering System

### Package: `github.com/opd-ai/venture/pkg/rendering`

#### Sprite Generation

**Package:** `pkg/rendering/sprites`

```go
sprite := sprites.Generate(width, height, color, seed, entityType)
```

#### Color Palettes

**Package:** `pkg/rendering/palette`

Genre-specific color schemes.

```go
pal := palette.Generate("fantasy", seed)
primary := pal.GetPrimaryColor()
```

#### Particles

**Package:** `pkg/rendering/particles`

Object pooling for performance.

```go
emitter := particles.NewEmitter(x, y, particleType)
emitter.Emit(count)
```

#### Lighting

**Package:** `pkg/rendering/lighting`

Dynamic lighting with intensity/color/falloff.

```go
light := lighting.NewLight(x, y, intensity, color)
```

#### Cache & Pool

**Packages:** `pkg/rendering/cache`, `pkg/rendering/pool`

Sprite caching (95.9% hit rate), object pooling for performance.

---

## Audio System

### Package: `github.com/opd-ai/venture/pkg/audio`

#### Synthesis

**Package:** `pkg/audio/synthesis`

Waveform generation (sine, square, triangle, sawtooth).

```go
wave := synthesis.Sine(frequency, duration, sampleRate)
```

#### Music

**Package:** `pkg/audio/music`

Procedural composition with genre themes and motifs.

```go
track := music.Generate(seed, genreID, params)
```

#### Sound Effects

**Package:** `pkg/audio/sfx`

Combat, movement, UI sound generation.

---

## Networking

### Package: `github.com/opd-ai/venture/pkg/network`

#### Client-Server Architecture

Authoritative server, client-side prediction, lag compensation (200-5000ms support).

#### Protocol

```go
// Message types
type MessageType uint8
const (
    MsgConnect MessageType = iota
    MsgDisconnect
    MsgPlayerInput
    MsgWorldState
)
```

#### Prediction

```go
// Client predicts locally, reconciles with server
client.PredictMovement(input)
client.ReconcileState(serverState)
```

#### Interpolation

```go
// Smooth remote entity movement
interpolator.AddSnapshot(state, timestamp)
interpolatedState := interpolator.Get(renderTime)
```

---

## Save/Load System

### Package: `github.com/opd-ai/venture/pkg/saveload`

JSON serialization of game state.

```go
// Save
saveload.Save(world, "save.json")

// Load
world := saveload.Load("save.json")
```

**Saved Data:** Entities, components, world seed, player progress

---

## UI Systems

Menu navigation, HUD, inventory screens, character sheets, skill trees, quest logs, crafting interfaces.

**Package:** `pkg/rendering/ui`

Standard dual-exit pattern: ESC or dedicated close button.

---

## Examples

See `examples/` directory for standalone demos:
- `complete_dungeon_generation/` - Full generation pipeline
- `genre_blending_demo/` - Cross-genre blending
- `multiplayer_demo/` - Client-server integration
- `optimization_demo/` - Performance techniques

**Run:** `go run ./examples/<name>` or `go build ./examples/<name>`

---

## Additional Resources

- [Architecture](ARCHITECTURE.md) - System design
- [Development Guide](DEVELOPMENT.md) - Setup and workflow
- [User Manual](USER_MANUAL.md) - Gameplay guide
- [Performance Guide](PERFORMANCE.md) - Optimization
- [Testing Guide](TESTING.md) - Test infrastructure

**Repository:** https://github.com/opd-ai/venture
