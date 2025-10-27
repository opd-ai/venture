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

// Add systems with proper constructors (parameters required)
world.AddSystem(engine.NewMovementSystem(200.0))  // Max speed parameter required
world.AddSystem(engine.NewCollisionSystem(32.0))  // Cell size parameter required
world.AddSystem(engine.NewInputSystem())          // No parameters

// Update world (call every frame)
deltaTime := 0.016 // 60 FPS
world.Update(deltaTime)

// Query entities
entities := world.GetEntities()

// Query specific entities
playersAndEnemies := world.GetEntitiesWith("position", "health")
```

**Methods:**
- `NewWorld() *World` - Create new world instance
- `AddSystem(system System)` - Register a system
- `Update(deltaTime float64)` - Update all systems
- `GetEntities() []*Entity` - Get all entities (returns cached list)
- `AddEntity(entity *Entity)` - Add entity to world
- `RemoveEntity(entityID uint64)` - Remove entity by ID
- `GetEntity(entityID uint64) (*Entity, bool)` - Find entity by ID
- `CreateEntity() *Entity` - Create new entity with auto-assigned ID
- `GetEntitiesWith(componentTypes ...string) []*Entity` - Get entities with specific components

#### Entity

Entities are containers for components.

```go
// Create entity
entity := engine.NewEntity(123) // ID: 123

// Add components
entity.AddComponent(&engine.PositionComponent{X: 100, Y: 50})
entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 0})

// Get component by type string
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
- `GetComponent(componentType string) (Component, bool)` - Get component by type
- `HasComponent(componentType string) bool` - Check component existence
- `RemoveComponent(componentType string)` - Remove component

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
- `PositionComponent` - X, Y coordinates (float64)
- `VelocityComponent` - VX, VY velocity (float64)
- `ColliderComponent` - Collision box (Width, Height, Solid, IsTrigger, Layer, OffsetX, OffsetY)
- `BoundsComponent` - World boundaries (MinX, MinY, MaxX, MaxY, Wrap)
- `SpriteComponent` - Visual representation (Width, Height, Color, Image, Layer)
- `HealthComponent` - HP tracking (Current, Max float64)
- `StatsComponent` - RPG stats (Attack, Defense, MagicPower, MagicDefense, CritChance, CritDamage, Evasion, Resistances)
- `AttackComponent` - Combat abilities (Damage, DamageType, Range, Cooldown, CooldownTimer)
- `StatusEffectComponent` - Temporary effects (EffectType, Duration, Magnitude, TickInterval, NextTick)
- `TeamComponent` - Team/faction ID (TeamID int)
- `InventoryComponent` - Item storage (Items, MaxItems, MaxWeight, Gold)
- `EquipmentComponent` - Equipped items (Slots map[EquipmentSlot]*Item, CachedStats, StatsDirty)
- `AIComponent` - AI behavior state machine (State, Data, TargetEntity, LastUpdate, StateDuration)
- `InputComponent` - Player input state (Up, Down, Left, Right, Attack, Use, various key states)
- `NetworkComponent` - Network synchronization data (OwnerID, Priority, LastSync)
- `ExperienceComponent` - XP and leveling (Level, CurrentXP, RequiredXP, SkillPoints, StatPoints)
- `QuestTrackerComponent` - Quest progress tracking (ActiveQuests, MaxActiveQuests)
- `ParticleEmitterComponent` - Particle effects (Particles, MaxParticles, EmissionRate, LastEmission)
- `CameraComponent` - Camera targeting (TargetX, TargetY, Smoothing, Shake)
- `ManaComponent` - Magic resource (Current, Max int, Regen float64)
- `SpellSlotComponent` - Spell management (Slots [5]*Spell)
- `VisualFeedbackComponent` - Visual effects (FlashTimer, FlashColor, TintTimer, TintColor)
- `HotbarComponent` - Quick-use item slots (Slots, MaxSlots)

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
- `MovementSystem` - Applies velocity to position (`NewMovementSystem(maxSpeed float64)`)
- `CollisionSystem` - Detects collisions (`NewCollisionSystem(cellSize float64)`)
- `CombatSystem` - Damage calculation (`NewCombatSystem(seed int64)`)
- `PlayerCombatSystem` - Player-specific combat handling (`NewPlayerCombatSystem(combatSystem, world)`)
- `AISystem` - Enemy behavior AI state machine (`NewAISystem(world *World)`)
- `ProgressionSystem` - XP and leveling (`NewProgressionSystem(world *World)`)
- `SkillProgressionSystem` - Skill tree progression (`NewSkillProgressionSystem()`)
- `InventorySystem` - Item management (`NewInventorySystem(world *World)`)
- `InputSystem` - Player input handling (`NewInputSystem()`)
- `RenderSystem` - Visual rendering (`NewRenderSystem(cameraSystem *CameraSystem)`)
- `TerrainRenderSystem` - Terrain/tile rendering (`NewTerrainRenderSystem(tileWidth, tileHeight int, genreID string, seed int64)`)
- `CameraSystem` - Camera control (`NewCameraSystem(screenWidth, screenHeight int)`)
- `HUDSystem` - UI overlay (`NewHUDSystem(screenWidth, screenHeight int)`)
- `ParticleSystem` - Particle effects (`NewParticleSystem()`)
- `ObjectiveTrackerSystem` - Quest objective tracking (`NewObjectiveTrackerSystem()`)
- `ItemPickupSystem` - Automatic item collection (`NewItemPickupSystem(world *World)`)
- `SpellCastingSystem` - Magic spell execution (`NewSpellCastingSystem(world *World)`)
- `PlayerSpellCastingSystem` - Player spell casting (`NewPlayerSpellCastingSystem(spellCasting, world)`)
- `PlayerItemUseSystem` - Player item usage (`NewPlayerItemUseSystem(inventory, world)`)
- `ManaRegenSystem` - Mana regeneration (`&ManaRegenSystem{}`)
- `TutorialSystem` - Tutorial guidance (`NewTutorialSystem()`)
- `MenuSystem` - Game menus and save/load (`NewMenuSystem(world, screenWidth, screenHeight, saveDir)`)
- `VisualFeedbackSystem` - Hit flashes and visual effects (`NewVisualFeedbackSystem()`)
- `AudioManagerSystem` - Audio playback (`NewAudioManagerSystem(audioManager)`)

#### Game

The Game struct integrates with Ebiten.

```go
// Create game
game := engine.NewGame(800, 600)

// Add systems to the world
game.World.AddSystem(engine.NewMovementSystem(200.0))
game.World.AddSystem(engine.NewInputSystem())

// Set up a player entity
player := game.World.CreateEntity() // Use CreateEntity() for auto-ID assignment
player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
game.SetPlayerEntity(player)

// Run game loop
err := game.Run("Venture")
if err != nil {
    log.Fatal(err)
}
```

**Methods:**
- `NewGame(screenWidth, screenHeight int) *Game` - Create game instance
- `Update() error` - Called every frame (implements ebiten.Game)
- `Draw(screen *ebiten.Image)` - Render frame (implements ebiten.Game)
- `Layout(outsideWidth, outsideHeight int) (int, int)` - Screen size (implements ebiten.Game)
- `Run(title string) error` - Start game loop
- `SetPlayerEntity(entity *Entity)` - Set player entity for UI systems
- `SetInventorySystem(system *InventorySystem)` - Connect inventory system to UI
- `SetupInputCallbacks(inputSystem *InputSystem, objectiveTracker *ObjectiveTrackerSystem)` - Setup UI callbacks

---

## Entity-Component-System

### Creating Entities

```go
// Player entity
player := world.CreateEntity() // Use CreateEntity() for auto-ID assignment
player.AddComponent(&engine.PositionComponent{X: 400, Y: 300})
player.AddComponent(&engine.VelocityComponent{})
playerSprite := engine.NewSpriteComponent(32, 32, color.RGBA{100, 150, 255, 255})
player.AddComponent(playerSprite)
player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
playerStats := engine.NewStatsComponent()
playerStats.Attack = 10
playerStats.Defense = 5
player.AddComponent(playerStats)
player.AddComponent(&engine.InputComponent{})
player.AddComponent(engine.NewInventoryComponent(20, 100.0)) // 20 slots, 100kg capacity
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
        } else if tile == terrain.TileDoor {
            // Doorway
        } else if tile == terrain.TileCorridor {
            // Corridor connecting rooms
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

// Generate entities (returns slice)
result, err := gen.Generate(seed, params)
if err != nil {
    log.Fatal(err)
}
entities := result.([]*entity.Entity)
generatedEntity := entities[0] // First entity

fmt.Printf("Name: %s\n", generatedEntity.Name)
fmt.Printf("Type: %s\n", generatedEntity.Type)
fmt.Printf("Level: %d\n", generatedEntity.Stats.Level)
fmt.Printf("HP: %d, Attack: %d, Defense: %d\n",
    generatedEntity.Stats.Health, generatedEntity.Stats.Damage, generatedEntity.Stats.Defense)

// Generate multiple entities with different seeds
rng := rand.New(rand.NewSource(seed))
for i := 0; i < 20; i++ {
    result, _ := gen.Generate(rng.Int63(), params)
    entities := result.([]*entity.Entity)
    entity := entities[0]
    // ... use entity
}
```

### Item Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/item`

```go
gen := item.NewItemGenerator()

// Generate random items (returns []*item.Item)
result, err := gen.Generate(seed, params)
items := result.([]*item.Item)
generatedItem := items[0] // First item

fmt.Printf("%s (%s)\n", generatedItem.Name, generatedItem.Rarity)
fmt.Printf("Type: %s\n", generatedItem.Type)
fmt.Printf("Value: %d gold\n", generatedItem.Stats.Value)

// Display stats
if generatedItem.Stats.Damage > 0 {
    fmt.Printf("  +%d Damage\n", generatedItem.Stats.Damage)
}
if generatedItem.Stats.Defense > 0 {
    fmt.Printf("  +%d Defense\n", generatedItem.Stats.Defense)
}
if generatedItem.Stats.AttackSpeed > 0 {
    fmt.Printf("  %.1fx Attack Speed\n", generatedItem.Stats.AttackSpeed)
}

// Generate specific type
// Generate specific type
params.Custom = map[string]interface{}{
    "type": "weapon",
    "count": 1,
}
result, _ = gen.Generate(seed, params)
weapons := result.([]*item.Item)
weapon := weapons[0]
```

### Magic Generation

**Package:** `github.com/opd-ai/venture/pkg/procgen/magic`

```go
gen := magic.NewSpellGenerator()

result, err := gen.Generate(seed, params)
spells := result.([]*magic.Spell)
spell := spells[0] // First spell

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
gen := skills.NewSkillTreeGenerator()

result, err := gen.Generate(seed, params)
trees := result.([]*skills.SkillTree)
tree := trees[0] // First tree

fmt.Printf("Tree: %s\n", tree.Name)
fmt.Printf("Skills: %d\n", len(tree.Nodes))

for _, skill := range tree.Nodes {
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
gen := quest.NewQuestGenerator()

result, err := gen.Generate(seed, params)
quests := result.([]*quest.Quest)
generatedQuest := quests[0] // First quest

fmt.Printf("Quest: %s\n", generatedQuest.Name)
fmt.Printf("Type: %s\n", generatedQuest.Type)
fmt.Printf("Description: %s\n", generatedQuest.Description)

// Objectives
for _, obj := range generatedQuest.Objectives {
    fmt.Printf("  - %s (%d/%d)\n", 
        obj.Description, obj.Current, obj.Target)
}

// Rewards
fmt.Printf("Rewards:\n")
fmt.Printf("  XP: %d\n", generatedQuest.RewardXP)
fmt.Printf("  Gold: %d\n", generatedQuest.RewardGold)
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

#### Basic Sprite Generation

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

#### Directional Sprite Generation

Generate 4-directional sprites for top-down gameplay with automatic facing support.

```go
gen := sprites.NewGenerator()

// Generate 4-directional sprite sheet (Up, Down, Left, Right)
config := sprites.GenerationConfig{
    Width:    32,
    Height:   32,
    Seed:     12345,
    GenreID:  "fantasy",
    EntityType: "humanoid",
    UseAerial: true,  // Enable aerial-view perspective
}

sprites, err := gen.GenerateDirectionalSprites(config)
if err != nil {
    log.Fatal(err)
}

// sprites is map[Direction]*ebiten.Image
upSprite := sprites[sprites.DirUp]
downSprite := sprites[sprites.DirDown]
leftSprite := sprites[sprites.DirLeft]
rightSprite := sprites[sprites.DirRight]

// Render based on entity facing direction
currentDirection := entity.GetComponent("animation").(*engine.AnimationComponent).Facing
screen.DrawImage(sprites[currentDirection], opts)
```

#### Aerial-View Templates

Aerial-view templates provide top-down character perspectives optimized for overhead gameplay cameras. All templates maintain consistent 35/50/15 proportions (head/torso/legs).

**Base Aerial Template:**

```go
// Get base humanoid aerial template
template := sprites.HumanoidAerial()

// Template maintains 35/50/15 proportions:
// - Head: 35% of total height (top of sprite)
// - Torso: 50% of total height (middle section)
// - Legs: 15% of total height (bottom of sprite)

// Each body part has:
// - RelativeWidth/RelativeHeight (percentage of sprite dimensions)
// - RelativeX/RelativeY (position from sprite center, 0.5 = center)
// - Role (primary, secondary, accent, detail)
// - Shape (rectangle, ellipse, polygon)
```

**Genre-Specific Aerial Templates:**

```go
// Get genre-specific template with thematic variations
fantasyTemplate := sprites.FantasyHumanoidAerial()
// - Broader shoulders (0.55 width)
// - Helmet/head detail (accent role)
// - Shield/weapon positioning

scifiTemplate := sprites.SciFiHumanoidAerial()
// - Angular shapes
// - Jetpack/tech details
// - Streamlined proportions

horrorTemplate := sprites.HorrorHumanoidAerial()
// - Narrow, elongated head (0.28 width)
// - Reduced shadow opacity (0.2)
// - Unsettling asymmetry

cyberpunkTemplate := sprites.CyberpunkHumanoidAerial()
// - Compact build (0.45 torso width)
// - Tech accents (neon glows)
// - Urban aesthetic

postapocTemplate := sprites.PostApocalypticHumanoidAerial()
// - Ragged organic shapes
// - Makeshift appearance
// - Survival theme
```

**Boss Scaling:**

```go
// Scale any aerial template for boss entities
baseTemplate := sprites.FantasyHumanoidAerial()
bossTemplate := sprites.BossAerialTemplate(baseTemplate, 2.5)

// Boss scaling:
// - Uniformly scales all body part dimensions by 2.5x
// - Preserves 35/50/15 proportions
// - Maintains directional asymmetry
// - Scales position offsets from center (asymmetry preservation)

// Use scaled template for boss sprite generation
config := sprites.GenerationConfig{
    Width:      64,  // Larger canvas for scaled boss
    Height:     64,
    Template:   &bossTemplate,
    UseAerial:  true,
}
```

**Directional Asymmetry:**

Aerial templates include directional variations for visual clarity:

- **Up (North)**: Head offset upward, arms positioned high
- **Down (South)**: Head centered, arms at sides
- **Left (West)**: Head offset left, left arm forward
- **Right (East)**: Head offset right, right arm forward

```go
// Templates automatically create directional variants
// Movement system updates facing based on velocity
// Render system selects correct directional sprite

// No manual direction handling required - it's automatic!
```

**Proportion Ratios:**

All aerial templates follow strict proportion guidelines:

| Body Part | Height % | Purpose |
|-----------|----------|---------|
| Head      | 35%      | Character recognition, facial features |
| Torso     | 50%      | Main body mass, equipment visibility |
| Legs      | 15%      | Ground contact, movement indication |
| **Total** | **100%** | Complete sprite height |

**Color Roles:**

- `Primary`: Main body color (torso, legs)
- `Secondary`: Accent elements (clothing, armor)
- `Accent`: Highlights (eyes, weapons, glows)
- `Detail`: Fine features (shadows, outlines)

**Integration with Movement System:**

```go
// Movement system automatically updates facing direction
// Based on velocity vector (see pkg/engine/movement.go)

// Horizontal priority: |VX| >= |VY| chooses left/right
// Vertical movement: |VY| > |VX| chooses up/down
// Jitter filtering: velocities < 0.1 don't change facing
// Action preservation: attack/hit/death states preserve facing

// Example: Entity moving right with velocity (5.0, 0.0)
// → AnimationComponent.Facing automatically set to DirRight
// → RenderSystem displays sprites[DirRight]
// No manual coordination needed!
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
mgr, err := saveload.NewSaveManager("./saves")
if err != nil {
    log.Fatal(err)
}

// Build game save
save := &saveload.GameSave{
    PlayerState: &saveload.PlayerState{
        EntityID:      1,
        X:             100,
        Y:             50,
        CurrentHealth: 80,
        MaxHealth:     100,
        Level:         5,
        Experience:    1200,
    },
    WorldState: &saveload.WorldState{
        Seed:       12345,
        GenreID:    "fantasy",
        Depth:      10,
        TimePlayed: 3600,
    },
    Settings: &saveload.GameSettings{
        MusicVolume: 0.8,
        SfxVolume:   1.0,
        WindowWidth: 800,
        WindowHeight: 600,
    },
}

// Save
err = mgr.SaveGame("mysave", save)
if err != nil {
    log.Fatal(err)
}
```

### Loading Game

```go
// Load save
save, err := mgr.LoadGame("mysave")
if err != nil {
    log.Fatal(err)
}

// Restore game state
playerX := save.PlayerState.X
playerY := save.PlayerState.Y
worldSeed := save.WorldState.Seed

// List saves
saves, err := mgr.ListSaves()
if err != nil {
    log.Fatal(err)
}
for _, save := range saves {
    fmt.Printf("%s - Level %d\n", save.Name, save.PlayerLevel)
}
```

### Save Management

```go
// Delete save
err := mgr.DeleteSave("mysave")
if err != nil {
    log.Fatal(err)
}

// Check if save exists
exists := mgr.SaveExists("mysave")

// Get save metadata
metadata, err := mgr.GetSaveMetadata("mysave")
if err != nil {
    log.Fatal(err)
}
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
