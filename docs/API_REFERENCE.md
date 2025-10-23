# Venture API Reference

Developer documentation for the Venture procedural action-RPG engine.

**Version:** 1.0  
**Last Updated:** October 2025

**New to development?** Start with [Development Guide](DEVELOPMENT.md) for setup and [Contributing Guide](CONTRIBUTING.md) for guidelines.

---

## Table of Contents

1. [Core Engine](#core-engine)
2. [Entity-Component-System](#entity-component-system)
3. [Procedural Generation](#procedural-generation)
4. [Rendering System](#rendering-system)
5. [Audio System](#audio-system)
6. [Networking](#networking)
7. [Save/Load System](#saveload-system)
8. [Examples](#examples)

---

## Core Engine

### Package: `github.com/opd-ai/venture/pkg/engine`

The engine package provides the ECS framework and core game systems.

#### World

The World is the central ECS container.

```go
// Create a new world
world := engine.NewWorld()

// Add systems
world.AddSystem(&engine.MovementSystem{})
world.AddSystem(engine.NewCollisionSystem(32.0))

// Update world (call every frame)
deltaTime := 0.016 // 60 FPS
world.Update(deltaTime)

// Query entities
entities := world.GetEntities()
```

**Methods:**
- `NewWorld() *World` - Create new world instance
- `AddSystem(system System)` - Register a system
- `Update(deltaTime float64)` - Update all systems
- `GetEntities() []*Entity` - Get all entities
- `AddEntity(entity *Entity)` - Add entity to world
- `RemoveEntity(id uint64)` - Remove entity by ID
- `GetEntity(id uint64) *Entity` - Find entity by ID

#### Entity

Entities are containers for components.

```go
// Create entity
entity := engine.NewEntity(123) // ID: 123

// Add components
entity.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 0})

// Get component
pos, ok := entity.GetComponent("position")
if ok {
    position := pos.(*engine.PositionComponent)
    fmt.Printf("Position: %.2f, %.2f\n", position.X, position.Y)
}

// Check for component
hasVelocity := entity.HasComponent("velocity")

// Remove component
entity.RemoveComponent("position")
```

**Methods:**
- `NewEntity(id uint64) *Entity` - Create entity with ID
- `AddComponent(comp Component)` - Add component
- `GetComponent(name string) (Component, bool)` - Get component by type
- `HasComponent(name string) bool` - Check component existence
- `RemoveComponent(name string)` - Remove component
- `GetComponents() map[string]Component` - Get all components

#### Component

Components are pure data structures.

```go
// Define custom component
type MyComponent struct {
    Value int
    Data  string
}

func (m MyComponent) Type() string {
    return "my_component"
}

// Usage
entity.AddComponent(&MyComponent{Value: 42, Data: "hello"})
```

**Built-in Components:**
- `PositionComponent` - X, Y coordinates
- `VelocityComponent` - VX, VY velocity
- `SpriteComponent` - Visual representation
- `ColliderComponent` - Collision box
- `HealthComponent` - HP tracking
- `StatsComponent` - RPG stats
- `InventoryComponent` - Item storage
- `EquipmentComponent` - Equipped items
- `AIComponent` - AI behavior
- `InputComponent` - Player input

#### System

Systems contain game logic.

```go
// Define custom system
type MySystem struct {
    // System state
}

func (s *MySystem) Update(entities []*engine.Entity, deltaTime float64) {
    // Process entities
    for _, entity := range entities {
        // Check for required components
        if !entity.HasComponent("position") {
            continue
        }
        
        pos := entity.GetComponent("position").(*engine.PositionComponent)
        // ... update logic
    }
}

// Register system
world.AddSystem(&MySystem{})
```

**Built-in Systems:**
- `MovementSystem` - Applies velocity to position
- `CollisionSystem` - Detects collisions
- `CombatSystem` - Damage calculation
- `AISystem` - Enemy behavior
- `ProgressionSystem` - XP and leveling
- `InventorySystem` - Item management
- `InputSystem` - Player input handling
- `RenderSystem` - Visual rendering
- `CameraSystem` - Camera control
- `HUDSystem` - UI overlay

#### Game

The Game struct integrates with Ebiten.

```go
// Create game
game := engine.NewGame(800, 600)

// Add systems
game.World.AddSystem(&engine.MovementSystem{})

// Run game loop
err := game.Run("Venture")
if err != nil {
    log.Fatal(err)
}
```

**Methods:**
- `NewGame(width, height int) *Game` - Create game instance
- `Update() error` - Called every frame (implements ebiten.Game)
- `Draw(screen *ebiten.Image)` - Render frame
- `Layout(w, h int) (int, int)` - Screen size
- `Run(title string) error` - Start game loop

---

## Entity-Component-System

### Creating Entities

```go
// Player entity
player := engine.NewEntity(1)
player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
player.AddComponent(&engine.VelocityComponent{})
player.AddComponent(&engine.SpriteComponent{Width: 32, Height: 32})
player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
player.AddComponent(&engine.StatsComponent{
    Attack:  10,
    Defense: 5,
    Speed:   100,
})
player.AddComponent(&engine.InputComponent{})
player.AddComponent(engine.NewInventoryComponent(20, 100))

world.AddEntity(player)
```

### Querying Entities

```go
// Find all enemies
var enemies []*engine.Entity
for _, entity := range world.GetEntities() {
    if entity.HasComponent("ai") && entity.HasComponent("combat") {
        enemies = append(enemies, entity)
    }
}

// Find player
var player *engine.Entity
for _, entity := range world.GetEntities() {
    if entity.HasComponent("input") {
        player = entity
        break
    }
}
```

### Component Access Patterns

```go
// Safe access
if pos, ok := entity.GetComponent("position"); ok {
    position := pos.(*engine.PositionComponent)
    position.X += 10
}

// Batch processing
for _, entity := range entities {
    // Skip entities without required components
    if !entity.HasComponent("position") || !entity.HasComponent("velocity") {
        continue
    }
    
    pos := entity.GetComponent("position").(*engine.PositionComponent)
    vel := entity.GetComponent("velocity").(*engine.VelocityComponent)
    
    // Apply velocity
    pos.X += vel.VX * deltaTime
    pos.Y += vel.VY * deltaTime
}
```

---

## Procedural Generation

### Package: `github.com/opd-ai/venture/pkg/procgen`

All generators implement the `Generator` interface:

```go
type Generator interface {
    Generate(seed int64, params GenerationParams) (interface{}, error)
    Validate(result interface{}) error
}
```

### Generation Parameters

```go
params := procgen.GenerationParams{
    Difficulty: 0.5,    // 0.0-1.0 (easy to hard)
    Depth:      10,     // Dungeon depth/level
    GenreID:    "fantasy", // Genre theme
    Custom: map[string]interface{}{
        "width":  80,
        "height": 50,
    },
}
```

### Terrain Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/terrain`

```go
// BSP Algorithm (structured dungeons)
bspGen := terrain.NewBSPGenerator()
result, err := bspGen.Generate(seed, params)
if err != nil {
    log.Fatal(err)
}
terrain := result.(*terrain.Terrain)

// Cellular Automata (organic caves)
cellGen := terrain.NewCellularGenerator()
result, err = cellGen.Generate(seed, params)
terrain = result.(*terrain.Terrain)

// Access terrain data
for y := 0; y < terrain.Height; y++ {
    for x := 0; x < terrain.Width; x++ {
        tile := terrain.GetTile(x, y)
        if tile == terrain.TileFloor {
            // Walkable floor
        } else if tile == terrain.TileWall {
            // Solid wall
        }
    }
}

// Get rooms
for _, room := range terrain.Rooms {
    fmt.Printf("Room: x=%d, y=%d, w=%d, h=%d\n",
        room.X, room.Y, room.Width, room.Height)
}
```

### Entity Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/entity`

```go
gen := entity.NewGenerator()

// Generate single entity
result, err := gen.Generate(seed, params)
if err != nil {
    log.Fatal(err)
}
entity := result.(*entity.Entity)

fmt.Printf("Name: %s\n", entity.Name)
fmt.Printf("Type: %s\n", entity.Type)
fmt.Printf("Level: %d\n", entity.Level)
fmt.Printf("HP: %d, Attack: %d, Defense: %d\n",
    entity.Health, entity.Attack, entity.Defense)

// Generate multiple entities
rng := rand.New(rand.NewSource(seed))
for i := 0; i < 20; i++ {
    result, _ := gen.Generate(rng.Int63(), params)
    e := result.(*entity.Entity)
    // ... use entity
}
```

### Item Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/item`

```go
gen := item.NewGenerator()

// Generate random item
result, err := gen.Generate(seed, params)
item := result.(*item.Item)

fmt.Printf("%s (%s)\n", item.Name, item.Rarity)
fmt.Printf("Type: %s\n", item.Type)
fmt.Printf("Value: %d gold\n", item.Value)

// Stats
for stat, value := range item.Stats {
    fmt.Printf("  +%d %s\n", value, stat)
}

// Generate specific type
params.Custom = map[string]interface{}{
    "type": "weapon",
}
result, _ = gen.Generate(seed, params)
weapon := result.(*item.Item)
```

### Magic Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/magic`

```go
gen := magic.NewGenerator()

result, err := gen.Generate(seed, params)
spell := result.(*magic.Spell)

fmt.Printf("Spell: %s\n", spell.Name)
fmt.Printf("Type: %s\n", spell.Type)
fmt.Printf("Element: %s\n", spell.Element)
fmt.Printf("Power: %d, Cost: %d MP\n", spell.Power, spell.ManaCost)
fmt.Printf("Target: %s\n", spell.TargetPattern)
fmt.Printf("Effect: %s\n", spell.Description)
```

### Skill Tree Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/skills`

```go
gen := skills.NewGenerator()

result, err := gen.Generate(seed, params)
tree := result.(*skills.SkillTree)

fmt.Printf("Tree: %s\n", tree.Name)
fmt.Printf("Skills: %d\n", len(tree.Skills))

for _, skill := range tree.Skills {
    fmt.Printf("  %s (Tier %d)\n", skill.Name, skill.Tier)
    fmt.Printf("    Type: %s\n", skill.Type)
    fmt.Printf("    Effect: %s\n", skill.Effect)
    
    if len(skill.Prerequisites) > 0 {
        fmt.Printf("    Requires: %v\n", skill.Prerequisites)
    }
}
```

### Quest Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/quest`

```go
gen := quest.NewGenerator()

result, err := gen.Generate(seed, params)
quest := result.(*quest.Quest)

fmt.Printf("Quest: %s\n", quest.Title)
fmt.Printf("Type: %s\n", quest.Type)
fmt.Printf("Description: %s\n", quest.Description)

// Objectives
for _, obj := range quest.Objectives {
    fmt.Printf("  - %s (%d/%d)\n", 
        obj.Description, obj.Current, obj.Target)
}

// Rewards
fmt.Printf("Rewards:\n")
fmt.Printf("  XP: %d\n", quest.RewardXP)
fmt.Printf("  Gold: %d\n", quest.RewardGold)
```

### Seed Management

```go
// Generate sub-seeds deterministically
seedGen := procgen.NewSeedGenerator(worldSeed)

terrainSeed := seedGen.GetSeed("terrain", roomID)
entitySeed := seedGen.GetSeed("entity", roomID)
itemSeed := seedGen.GetSeed("item", chestID)

// Use sub-seeds
terrainGen.Generate(terrainSeed, params)
entityGen.Generate(entitySeed, params)
itemGen.Generate(itemSeed, params)
```

---

## Rendering System

### Package: `github.com/opd-ai/venture/pkg/rendering`

### Color Palettes

**Package:** `github.com/opd-ai/venture/pkg/rendering/palette`

```go
gen := palette.NewGenerator()

result, err := gen.Generate(seed, procgen.GenerationParams{
    GenreID: "fantasy",
})
palette := result.(*palette.Palette)

// Access colors
primaryColor := palette.Primary
secondaryColor := palette.Secondary
accentColor := palette.Accent

// Get color by name
wall := palette.GetColor("wall")
floor := palette.GetColor("floor")
```

### Sprite Generation

**Package:** `github.com/opd-ai/venture/pkg/rendering/sprites`

```go
gen := sprites.NewGenerator()

result, err := gen.Generate(seed, procgen.GenerationParams{
    GenreID: "fantasy",
    Custom: map[string]interface{}{
        "width":  32,
        "height": 32,
        "type":   "monster",
    },
})
sprite := result.(*sprites.Sprite)

// sprite.Image is *ebiten.Image
// Render sprite
screen.DrawImage(sprite.Image, opts)
```

### Tile Rendering

**Package:** `github.com/opd-ai/venture/pkg/rendering/tiles`

```go
tileGen := tiles.NewGenerator()

// Generate tile
result, err := tileGen.Generate(seed, procgen.GenerationParams{
    GenreID: "fantasy",
    Custom: map[string]interface{}{
        "type": "floor",
    },
})
tile := result.(*tiles.Tile)

// Use tile image
screen.DrawImage(tile.Image, opts)
```

### Particle Effects

**Package:** `github.com/opd-ai/venture/pkg/rendering/particles`

```go
// Create emitter
emitter := particles.NewEmitter(x, y, "explosion")

// Update particles
emitter.Update(deltaTime)

// Render particles
for _, particle := range emitter.Particles {
    if particle.Active {
        // Draw particle at particle.X, particle.Y
    }
}
```

---

## Audio System

### Package: `github.com/opd-ai/venture/pkg/audio`

### Waveform Synthesis

**Package:** `github.com/opd-ai/venture/pkg/audio/synthesis`

```go
// Create oscillator
osc := synthesis.NewOscillator(synthesis.WaveSine, 440.0, 44100)

// Generate samples
duration := 1.0 // seconds
samples := osc.Generate(duration)

// samples is []float64, convert to audio format
```

### Sound Effects

**Package:** `github.com/opd-ai/venture/pkg/audio/sfx`

```go
gen := sfx.NewGenerator()

// Generate explosion sound
result, err := gen.Generate(seed, procgen.GenerationParams{
    Custom: map[string]interface{}{
        "type": "explosion",
    },
})
sound := result.(*sfx.Sound)

// Play sound
// sound.Samples is []float64
```

### Music Generation

**Package:** `github.com/opd-ai/venture/pkg/audio/music`

```go
gen := music.NewGenerator()

result, err := gen.Generate(seed, procgen.GenerationParams{
    GenreID: "fantasy",
    Custom: map[string]interface{}{
        "context":  "combat",
        "duration": 30.0, // seconds
    },
})
track := result.(*music.Track)

// track.Samples is []float64 audio data
// Play via audio library
```

---

## Networking

### Package: `github.com/opd-ai/venture/pkg/network`

### Server

```go
// Create server
config := network.DefaultServerConfig()
config.Address = ":8080"
config.MaxPlayers = 4
config.UpdateRate = 20 // Hz

// Server setup (network layer is work-in-progress)
// See cmd/server/main.go for full example
```

### Client

```go
// Create client
config := network.DefaultClientConfig()
config.ServerAddress = "localhost:8080"

// Client setup (network layer is work-in-progress)
// See cmd/client/main.go for full example
```

### State Synchronization

```go
// Create snapshot manager
snapMgr := network.NewSnapshotManager(100) // history size

// Record snapshot
snapshot := network.WorldSnapshot{
    Timestamp: time.Now(),
    Entities:  make(map[uint64]network.EntitySnapshot),
}

// Add entity states
snapshot.Entities[entityID] = network.EntitySnapshot{
    EntityID: entityID,
    Position: network.Position{X: x, Y: y},
    Velocity: network.Velocity{VX: vx, VY: vy},
}

snapMgr.AddSnapshot(snapshot)

// Get historical snapshot
past := snapMgr.GetSnapshot(timestamp)
```

### Client-Side Prediction

```go
predictor := network.NewPredictor()

// Apply input locally
input := network.Input{
    Timestamp: time.Now(),
    MoveX:     1.0,
    MoveY:     0.0,
}
predictor.ApplyInput(input, entity)

// Reconcile with server
serverState := network.EntitySnapshot{...}
predictor.Reconcile(serverState, entity)
```

---

## Save/Load System

### Package: `github.com/opd-ai/venture/pkg/saveload`

### Saving Game

```go
// Create save manager
mgr := saveload.NewManager("./saves")

// Build game state
state := &saveload.GameState{
    Version: "1.0",
    Player: saveload.PlayerState{
        Position: saveload.Position{X: 100, Y: 50},
        Health:   80,
        MaxHealth: 100,
        Level:    5,
        Experience: 1200,
    },
    World: saveload.WorldState{
        Seed:       12345,
        GenreID:    "fantasy",
        Depth:      10,
        TimePlayed: 3600,
    },
}

// Save
err := mgr.Save("mysave", state)
if err != nil {
    log.Fatal(err)
}
```

### Loading Game

```go
// Load save
state, err := mgr.Load("mysave")
if err != nil {
    log.Fatal(err)
}

// Restore game state
playerX := state.Player.Position.X
playerY := state.Player.Position.Y
worldSeed := state.World.Seed

// List saves
saves, err := mgr.List()
for _, save := range saves {
    fmt.Printf("%s - Level %d\n", save.Name, save.Metadata.PlayerLevel)
}
```

### Save Management

```go
// Delete save
err := mgr.Delete("mysave")

// Check if save exists
exists := mgr.Exists("mysave")

// Get save metadata
metadata, err := mgr.GetMetadata("mysave")
fmt.Printf("Saved at: %s\n", metadata.SavedAt)
fmt.Printf("Level: %d\n", metadata.PlayerLevel)
```

---

## Examples

### Complete Entity Creation

```go
func CreatePlayer(world *engine.World, x, y float64) *engine.Entity {
    player := engine.NewEntity(1)
    
    // Position
    player.AddComponent(&engine.PositionComponent{X: x, Y: y})
    
    // Movement
    player.AddComponent(&engine.VelocityComponent{})
    
    // Rendering
    player.AddComponent(engine.NewSpriteComponent(32, 32, color.RGBA{0, 255, 0, 255}))
    
    // Combat
    player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
    player.AddComponent(&engine.StatsComponent{
        Attack:     10,
        Defense:    5,
        MagicPower: 8,
        Speed:      100,
    })
    
    // RPG systems
    player.AddComponent(engine.NewExperienceComponent())
    player.AddComponent(engine.NewInventoryComponent(20, 100))
    player.AddComponent(engine.NewEquipmentComponent())
    
    // Input
    player.AddComponent(&engine.InputComponent{})
    
    // Camera target
    player.AddComponent(engine.NewCameraComponent())
    
    world.AddEntity(player)
    return player
}
```

### Simple Game Loop

```go
func main() {
    // Create game
    game := engine.NewGame(800, 600)
    
    // Add systems
    game.World.AddSystem(engine.NewInputSystem())
    game.World.AddSystem(&engine.MovementSystem{})
    game.World.AddSystem(engine.NewCollisionSystem(32.0))
    
    // Create player
    CreatePlayer(game.World, 400, 300)
    
    // Run
    if err := game.Run("My Game"); err != nil {
        log.Fatal(err)
    }
}
```

### Procedural Level Generation

```go
func GenerateLevel(seed int64, depth int) *terrain.Terrain {
    gen := terrain.NewBSPGenerator()
    
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        Depth:      depth,
        GenreID:    "fantasy",
        Custom: map[string]interface{}{
            "width":  80,
            "height": 50,
        },
    }
    
    result, err := gen.Generate(seed, params)
    if err != nil {
        log.Fatal(err)
    }
    
    return result.(*terrain.Terrain)
}
```

### Spawning Enemies

```go
func SpawnEnemies(world *engine.World, terrain *terrain.Terrain, seed int64, depth int) {
    entityGen := entity.NewGenerator()
    rng := rand.New(rand.NewSource(seed))
    
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        Depth:      depth,
        GenreID:    "fantasy",
    }
    
    for _, room := range terrain.Rooms {
        // Skip first room (spawn room)
        if room == terrain.Rooms[0] {
            continue
        }
        
        // Generate 1-3 enemies per room
        enemyCount := 1 + rng.Intn(3)
        
        for i := 0; i < enemyCount; i++ {
            // Generate enemy
            result, _ := entityGen.Generate(rng.Int63(), params)
            genEntity := result.(*entity.Entity)
            
            // Convert to ECS entity
            enemy := CreateEnemy(world, genEntity, room)
        }
    }
}
```

---

## Additional Resources

- **Package Documentation**: Each package has a `doc.go` file with details
- **Examples**: See `examples/` directory for standalone demos
- **Tests**: Package tests demonstrate usage patterns
- **User Manual**: [USER_MANUAL.md](USER_MANUAL.md) for gameplay info
- **Getting Started**: [GETTING_STARTED.md](GETTING_STARTED.md) for quick start

---

**Version:** 1.0  
**Last Updated:** October 2025  
**Repository:** https://github.com/opd-ai/venture
