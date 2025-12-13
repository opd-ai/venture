# Venture API Reference

Developer documentation for the Venture procedural action-RPG engine.

**Version:** 8.0  
**Last Updated:** December 2025

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
9. [Social Systems (v5.0)](#social-systems-v50)

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
gen := terrain.NewBSPGenerator()
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
gen := entity.NewEntityGenerator()
result, _ := gen.Generate(seed, params)
e := result.(*entity.Entity) // Type, Stats, Behavior
```

### Item Generation

**Package:** `pkg/procgen/item`

Weapons, armor, consumables with rarity system (Common, Uncommon, Rare, Epic, Legendary).

```go
gen := item.NewItemGenerator()
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

#### Sprite Generation (V3.0 Enhanced)

**Package:** `pkg/rendering/sprites`

**V3.0 Enhancements:**
- Enhanced anatomical templates with pixel-perfect dimensions
- Genre-specific variations (organic, geometric, distorted, augmented, weathered)
- Facial features for close-up views (eyes, mouth)
- 40% more anatomical detail

```go
// Create sprite generator
gen := sprites.NewGenerator()

// Basic sprite generation
config := sprites.Config{
    Type:       sprites.SpriteEntity,
    Width:      32,
    Height:     32,
    Seed:       12345,
    GenreID:    "fantasy",
    Complexity: 0.8,
}
sprite, err := gen.Generate(config)

// Generate directional sprites (for 8-way movement)
directions, err := gen.GenerateDirectionalSprites(config)
// Returns map[int]*ebiten.Image with sprites for 8 directions
```

**Performance:**
- Generation time: 3-5ms per sprite (with V3.0 enhancements)
- Cache hit rate: 95.9% (maintained from V2.0)
- Target: <5ms ✅

#### Tile Rendering (V3.0 Enhanced)

**Package:** `pkg/rendering/tiles`

**V3.0 Enhancements:**
- Procedural texture patterns (stone, wood, metal, organic)
- 50+ unique patterns per genre
- Smooth transitions with automated edge detection
- Multi-layer depth effects
- Detail layers and normal mapping simulation

```go
// Create tile generator
gen := tiles.NewGenerator()

// Generate tile with pattern
config := tiles.Config{
    Type:       tiles.TypeFloor,
    Size:       32,
    Seed:       12345,
    GenreID:    "fantasy",
    Complexity: 0.7,
}
tile, err := gen.Generate(config)
```

**Tile Types:** Floor, Wall, Door, Corridor, Water, Lava, Trap, Stairs

**Pattern Types (V3.0):**
- Solid, Checkerboard, Dots, Lines, Brick, Grain (procedurally varied per tile)

#### Lighting System (V3.0 Enhanced)

**Package:** `pkg/engine` (system), `pkg/rendering/lighting` (utilities)

**V3.0 Enhancements:**
- Soft shadows with gradient edges
- Colored lighting matching light sources
- Bloom effects for magical/technological lights
- Advanced ambient occlusion
- Dynamic flickering animations
- Genre-specific presets
- <5% frame time overhead

```go
// Lighting system in engine
lightingSystem := engine.NewLightingSystem(world)

// Add light source to an entity
lightComp := &engine.LightComponent{
    Color:     color.RGBA{255, 200, 150, 255}, // Warm torch color
    Intensity: 1.0,
    Radius:    150.0,
    Flicker:   true, // Dynamic flickering
}
entity.AddComponent(lightComp)

// Ambient occlusion (rendering/lighting package)
config := lighting.DefaultAOConfig()
config.Intensity = 0.8
enhancedImage := lighting.ApplyAmbientOcclusion(img, depthMap, config)

// Bloom effects
bloomImage := lighting.ApplyBloom(img, bloomRadius, bloomIntensity)
```

**Performance:**
- Frame time overhead: <5% with all lighting active
- Shadow quality: Gradient-based soft edges
- Bloom quality: Radius-based glow effects

#### Particle & Weather Systems (V3.0 New)

**Package:** `pkg/rendering/particles`

**V3.0 Weather System:**
- Comprehensive weather types: rain, snow, fog, dust, ash, and genre-specific variations
- Fluid simulation for realistic behavior
- Intensity levels: light, medium, heavy, extreme
- Environmental interactions (puddles, snow accumulation, visibility)

```go
// Generate weather system
config := particles.WeatherConfig{
    Type:      particles.WeatherRain,
    Intensity: particles.IntensityMedium,
    GenreID:   "fantasy",
    Width:     800,
    Height:    600,
    Seed:      12345,
}
weatherSystem, err := particles.GenerateWeather(config)

// Update weather each frame
weatherSystem.Update(deltaTime)

// Get genre-specific weather types
genreWeathers := particles.GetGenreWeather("scifi")
// Returns: [WeatherNeonRain, WeatherSmog, WeatherRadiation]

// Check environmental effects
```

**Weather Types:**
- Standard: Rain, Snow, Fog, Dust, Ash, Sandstorm
- Genre-specific: NeonRain (sci-fi), Smog (cyberpunk), Radiation (post-apocalyptic), BloodRain (horror)

**Weather Types:**
- **Rain:** Water droplets with splash effects
- **Snow:** Snowflakes with accumulation
- **Fog:** Volumetric fog particles
- **Dust:** Swirling dust motes
- **Ash:** Falling ash particles

**Genre Variations:**
- Fantasy: Natural precipitation
- Sci-Fi: Neon particles, energy effects
- Horror: Blood, toxic substances
- Cyberpunk: Acid rain, pollution smog
- Post-Apocalyptic: Radiation, fallout

#### UI Enhancement (V3.0 Enhanced)

**Package:** `pkg/rendering/ui`

**V3.0 Enhancements:**
- Dynamic color palettes per genre
- Improved visual hierarchy
- Smooth transitions and animations
- Procedural UI decorations

```go
// V3.0: Dynamic UI colors based on genre
uiColors := ui.GetGenreColorScheme(genreID)
primaryColor := uiColors.Primary
secondaryColor := uiColors.Secondary
accentColor := uiColors.Accent

// V3.0: Smooth menu transitions
ui.TransitionToMenu(
    fromMenu, toMenu,
    transitionType, // ui.FadeIn, SlideLeft, SlideRight, etc.
    duration,
)

// V3.0: Visual hierarchy
ui.RenderWithHierarchy(
    elements,
    primaryLevel,    // Main focus elements
    secondaryLevel,  // Supporting information
    tertiaryLevel,   // Background/decorative
```

#### Post-Processing (V3.0 New)

**Package:** `pkg/rendering/postprocess`

**V3.0 Post-Processing Effects:**
- Parallax backgrounds with multi-layer depth
- Time-of-day system with dynamic lighting
- Screen-space enhancements
- Genre-specific visual filters

```go
// V3.0: Parallax background
parallax := postprocess.NewParallaxBackground(
    genreID, seed,
    layerCount, // Number of parallax layers (2-5)
)
parallax.Update(cameraX, cameraY)

// V3.0: Time-of-day system
timeOfDay := postprocess.NewTimeOfDaySystem(
    startHour, // 0-23
    dayLength, // Duration of full day cycle
)
timeOfDay.Update(deltaTime)
ambientColor := timeOfDay.GetAmbientColor()
lightIntensity := timeOfDay.GetLightIntensity()

// V3.0: Screen-space effects
postprocess.ApplyBloom(screen, bloomIntensity, bloomRadius)
postprocess.ApplyVignette(screen, vignetteIntensity)
```

#### Quality Settings (V3.0 New)

**Package:** `pkg/rendering/quality`

**V3.0 Quality Levels:**
- Anti-aliasing quality (Off, Low, Medium, High)
- Lighting quality settings
- Particle density levels
- Performance vs quality trade-offs

```go
// V3.0: Anti-aliasing quality
quality.SetAntiAliasing(quality.High) // Off, Low, Medium, High
// Off: No AA, fastest
// Low: 2x2 super-sampling
// Medium: 4x4 super-sampling
// High: 8x8 super-sampling

// V3.0: Lighting quality
quality.SetLighting(quality.High)
// Low: Hard shadows, basic lighting
// Medium: Soft shadows, colored lights
// High: Bloom effects, advanced ambient occlusion

// V3.0: Particle density
quality.SetParticleDensity(quality.Medium)
// Low: Fewer particles, better performance
// Medium: Balanced particle count
// High: Maximum particle density

```

#### Color Palettes

**Package:** `pkg/rendering/palette`

Genre-specific color schemes.

```go
pal := palette.Generate("fantasy", seed)
primary := pal.GetPrimaryColor()
```

#### Particles (Legacy, Enhanced in V3.0)

**Package:** `pkg/rendering/particles`

Object pooling for performance.

```go
emitter := particles.NewEmitter(x, y, particleType)
emitter.Emit(count)
```

#### Lighting (Legacy, Enhanced in V3.0)

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
- [Social Systems Guide](SOCIAL_SYSTEMS.md) - Chat, NPC dialog, trading (v5.0)
- [Migration Guide v5.0](MIGRATION_V5.md) - Upgrade from v4.0

**Repository:** https://github.com/opd-ai/venture

---

## Social Systems (v5.0)

### Package: `github.com/opd-ai/venture/pkg/network`

Multiplayer communication and E2E encryption.

#### Chat System

**ChatComponent** tracks chat state for player entities.

```go
// Create chat component
chat := engine.NewChatComponent()

// Add message
msg := engine.ChatMessage{
    ID:         "msg-uuid",
    SenderID:   playerEntity.ID,
    SenderName: "Alice",
    Channel:    engine.ChatGlobal,
    Content:    "Hello world!",
    Timestamp:  time.Now(),
}
chat.AddMessage(msg)

// Check rate limiting
canSend := chat.CanSendMessage(engine.ChatGlobal, time.Now())

// Apply mute for rate limit violation
chat.ApplyMute(time.Now(), 3) // 3 violations = 120s mute
```

**Key Methods:**
- `NewChatComponent()` - Create with default settings
- `AddMessage(msg)` - Add message to history (max 100)
- `MarkAllRead()` - Clear unread count
- `CanSendMessage(channel, now)` - Check rate limit
- `ApplyMute(now, count)` - Apply progressive mute (30s → 60s → 120s)
- `IsMuted(now)` - Check if currently muted
- `ActivateMegaphone()` - Extend local radius to 30 tiles
- `ActivateWalkieTalkie()` - Unlimited range for local chat

**Chat Channels:**
```go
engine.ChatGlobal   // All players, 3s cooldown
engine.ChatLocal    // 10-tile radius, 1s cooldown
engine.ChatParty    // Party members, 0.5s cooldown
engine.ChatWhisper  // Direct message, 0.5s cooldown
```

#### Encryption System

**E2E encryption** using Diffie-Hellman + AES-256-GCM.

```go
// Generate key pair (2048-bit DH)
keyPair := crypto.GenerateKeyPair()

// Perform key exchange
sharedSecret := crypto.ComputeSharedSecret(keyPair.Private, partnerPublicKey)

// Encrypt message
ciphertext, err := crypto.EncryptMessage(plaintext, sharedSecret)

// Decrypt message
plaintext, err := crypto.DecryptMessage(ciphertext, sharedSecret)
```

**Performance:**
- Key generation: ~50ms (one-time per connection)
- Encryption: ~50µs per message
- Decryption: ~40µs per message

**Security:**
- 2048-bit DH modulus (RFC 3526 Group 14)
- AES-256-GCM authenticated encryption
- Random 12-byte IV per message
- SHA-256 for key derivation

#### ChatSystem

Processes chat messages and enforces rate limits.

```go
// Create chat system
chatSystem := engine.NewChatSystem(world)

// Enable/disable system
chatSystem.Enable()
chatSystem.Disable()

// Update (called every frame)
chatSystem.Update(deltaTime)
```

**Responsibilities:**
- Message delivery (global, local, party, whisper)
- Range checking for local chat
- Rate limit enforcement
- Mute application
- Message encryption/decryption coordination

### Package: `github.com/opd-ai/venture/pkg/procgen/dialog`

NPC dialog generation using Markov chains.

#### NPCDialogComponent

Tracks conversation state for NPCs.

```go
// Create NPC dialog component
npcDialog := &engine.NPCDialogComponent{
    NPCPersonality: &dialog.Personality{
        Friendliness: 0.7,
        Formality:    0.5,
        Verbosity:    0.8,
    },
    GenreID:           "fantasy",
    DeterministicMode: false,
}

// Component methods
npcDialog.Type() // Returns "npcdialog"
```

**Key Fields:**
- `NPCPersonality` - Personality traits (friendliness, formality, verbosity)
- `ConversationHistory` - Last 10 player inputs
- `ResponseHistory` - Last 10 NPC responses
- `DialogState` - Current state ("greeting", "trading", "questing", etc.)
- `TopicMemory` - Topics discussed (avoid repetition)
- `Generator` - Cached Markov generator
- `DeterministicMode` - Use templates instead of Markov

#### Markov Generator

**MarkovGenerator** creates dynamic dialog responses.

```go
// Create generator
gen := dialog.NewMarkovGenerator(genreID, order)

// Train on corpus (automatic on creation)
// Corpus loaded from pkg/procgen/dialog/corpora/<genre>.txt

// Generate response
response, err := gen.Generate(seed, playerInput, personality)
```

**Configuration:**
- `order` - Markov chain order (2-3 recommended)
- `genreID` - Determines corpus ("fantasy", "scifi", "horror", etc.)
- `personality` - Influences word selection probabilities

**Performance:**
- Generation: <50ms per response
- Corpus loading: ~100ms (one-time per genre)
- Memory: ~50KB per trained generator

#### NPCDialogSystem

Generates NPC responses and manages conversation state.

```go
// Create dialog system
dialogSystem := engine.NewNPCDialogSystem()

// Update (processes dialog requests)
dialogSystem.Update(deltaTime)
```

**Features:**
- Automatic Markov generator creation
- Personality-based response variation
- Conversation history tracking
- Genre-appropriate vocabulary
- Fallback to templates on failure

### Package: `github.com/opd-ai/venture/pkg/network/trade`

Item trading with two-phase commit protocol.

#### TradeComponent

Tracks trading state and history.

```go
// TradeComponent fields
type TradeComponent struct {
    ActiveTrade  *TradeProposal // Current trade (nil if none)
    TradeHistory []TradeRecord  // Completed trades
    TrustScore   float64        // 0.0-1.0 (default 0.5)
}

// Component method
component.Type() // Returns "trade"
```

#### TradeProposal

Represents an active trade proposal.

```go
type TradeProposal struct {
    ProposerID     uint64   // Entity proposing trade
    RecipientID    uint64   // Entity receiving proposal
    OfferedItems   []string // Item IDs offered
    RequestedItems []string // Item IDs requested
    Status         string   // "pending", "accepted", "rejected", etc.
    ProposalTime   int64    // Unix timestamp
    FailureReason  string   // Reason if failed
}
```

**Statuses:**
- `"pending"` - Awaiting recipient response
- `"accepted"` - Recipient accepted, awaiting server validation
- `"rejected"` - Recipient declined
- `"committed"` - Server completed transfer
- `"cancelled"` - Proposer cancelled
- `"failed"` - Validation failed (proximity, trust, ownership)

#### Trade System

**Two-phase commit protocol:**
1. **Propose**: `CreateProposal(proposer, recipient, offered, requested)`
2. **Review**: `AcceptProposal()` / `RejectProposal()`
3. **Validate**: Server checks proximity, trust, ownership
4. **Commit**: `CommitTrade()` atomically transfers items
5. **Rollback**: `RollbackTrade(reason)` if any failure

**Trust Mechanics:**
```go
// Update trust after trade
UpdateTrust(success bool) {
    if success {
        trustScore += 0.05 // Max 1.0
    } else {
        trustScore -= 0.1  // Min 0.0
    }
}

// Check trust requirements
CanTradeItem(rarity string) bool {
    if trustScore > 0.8 {
        return true // Can trade any rarity
    }
    if trustScore < 0.3 {
        return rarity == "Common" || rarity == "Uncommon"
    }
    return rarity != "Legendary"
}
```

**Proximity Validation:**
- Initiate: Within 5 tiles
- During trade: Max 10 tiles
- Server uses lag compensation for fair checks

### Package: `github.com/opd-ai/venture/pkg/rendering/ui`

Social UI components.

#### ChatUI

Renders chat interface with message history and input field.

```go
// Create chat UI
chatUI := ui.NewChatUI(x, y, width, height)

// Set colors
chatUI.SetColors(bg, text, inputBG, border)

// Add message
msg := ui.ChatMessage{
    SenderName: "Bob",
    Content:    "Hello!",
    Channel:    0, // Global
    Timestamp:  time.Now(),
    IsSystem:   false,
}
chatUI.AddMessage(msg)

// Update and draw
chatUI.Update() // Handle input
chatUI.Draw(screen) // Render
```

**Features:**
- 4 channel tabs (Global, Local, Party, Whisper)
- Message history (last 100 messages)
- Scrolling support
- Unread count indicators
- Cursor blinking for input field
- Keyboard shortcuts

#### TradeUI

Renders trade proposal and acceptance interface.

```go
// Create trade UI
tradeUI := ui.NewTradeUI(x, y, width, height)

// Set proposal
proposal := &ui.TradeProposal{
    ProposerName:   "Alice",
    RecipientName:  "Bob",
    OfferedItems:   []ui.TradeItem{...},
    RequestedItems: []ui.TradeItem{...},
    Status:         "pending",
    ProposalTime:   time.Now(),
}
tradeUI.SetProposal(proposal)

// Update and draw
tradeUI.Update()
tradeUI.Draw(screen)

// Check for button clicks
```

**Features:**
- Two-panel item display (offered vs. requested)
- Rarity color indicators
- Accept/Reject buttons
- Status messages (accepted, rejected, failed, etc.)
- Hover effects
- Automatic visibility management

---

## Additional Resources

---

## Housing System (V8.0)

### Package: `github.com/opd-ai/venture/pkg/world/housing`

Player housing with procedural buildings and cross-server persistence.

#### Manager

```go
manager := housing.NewManager()

// Place plot
plot := &housing.Plot{
    OwnerID:  "player123",
    Location: housing.Coords{X: 100, Y: 200},
    Size:     housing.SizeMedium, // 16×16
}
manager.PlacePlot(plot)

// Query
nearby := manager.GetPlotsInRadius(housing.Coords{X: 100, Y: 200}, 50.0)
ownerPlots := manager.GetPlotsByOwner("player123")

// Save/load
manager.Save("world_housing.json.gz")
manager.Load("world_housing.json.gz")
```

**Plot Sizes:** Small (8×8), Medium (16×16), Large (24×24), Estate (32×32)

**Permissions:** None, Visit, Friend, CoOwner

**Performance:** <1ms placement, <0.1ms collision (1000 plots)

---

## Guild System (V8.0)

### Package: `github.com/opd-ai/venture/pkg/network/federation/guild`

Multi-server guild management.

```go
manager := guild.NewManager()

// Create guild
identity := guild.GenerateIdentity(12345, "fantasy")
g := &guild.Guild{ID: "guild123", Name: identity.Name, OwnerID: "player123"}
manager.CreateGuild(g)

// Members
manager.AddMember("guild123", "player456", guild.RankMember)
manager.DepositTreasury("guild123", 1000)
```

**Ranks:** Recruit, Member, Officer, Leader

**Permissions:** Invite, Kick, Promote, Treasury, GuildHall, Territory, MOTD, Diplomacy

---

## Social Persistence (V8.0)

### Package: `github.com/opd-ai/venture/pkg/social/persistence`

Trust, reputation, chat history, images.

```go
// Trust
trustMgr := persistence.NewTrustManager()
trustMgr.UpdateTrust("player1", "player2", 0.1)
trust := trustMgr.GetTrust("player1", "player2")
tier := trust.GetTier() // Stranger, Acquaintance, Friend, Trusted

// Chat
history := persistence.NewChatHistory("player1", 1000)
history.AddMessage(&persistence.Message{Sender: "player1", Content: "Hi"})
delta := history.GetDelta(lastVersion) // Delta sync

// Images
gallery := persistence.NewImageGallery("player1", 100, 50*1024*1024)
gallery.AddImage(&persistence.StoredImage{Data: imageBytes, Tags: []string{"combat"}})
```

---

## Advanced Physics (V8.0)

### Package: `github.com/opd-ai/venture/pkg/engine/physics`

Vehicle physics, fluid dynamics, destruction.

```go
// Vehicle suspension
suspension := &vehicle.SuspensionComponent{
    Wheels: []vehicle.Wheel{{Stiffness: 100, Damping: 20}},
}

// Fluid simulation
sim := fluids.NewSimulator(&fluids.SimulatorConfig{Width: 100, Height: 100})
sim.AddFluid(50, 50, fluids.FluidWater, 10.0)
sim.Update(1.0/30.0) // 30 Hz

// Building destruction
system := destruction.NewSystem(30.0)
system.ApplyDamage("building123", position, 500.0, 5.0)
```

---

## Deep Gameplay (V8.0)

### Companion AI

Package: `github.com/opd-ai/venture/pkg/companion/learning`

```go
manager := learning.NewManager()
companion := manager.RegisterCompanion("companion123", "player456")
manager.AddExperience("companion123", learning.SkillSwordMastery, 50)
manager.AdjustTrait("companion123", learning.TraitBrave, 0.1)
```

**24 Skills:** 8 categories (Combat, Defense, Utility, Social, Healing, Magic, Crafting, Stealth)

**10 Personality Traits:** Cautious/Brave, Shy/Outgoing, Aggressive/Pacifist, etc.

### Branching Narratives

Package: `github.com/opd-ai/venture/pkg/narrative/branching`

```go
gen := branching.NewGenerator()
arc := gen.Generate(12345, params).(*branching.StoryArc)

manager := branching.NewManager()
manager.StartArc("player123", arc)
manager.MakeChoice("player123", arc.ID, choiceID)
```

**6 Ending Types:** Heroic, Tragic, Neutral, Mystery, Triumph, Betrayal

### Advanced Classes

Package: `github.com/opd-ai/venture/pkg/class/advanced`

```go
manager := advanced.NewManager()
manager.SetPrimaryClass("player123", advanced.ClassWarrior)
manager.SetSecondaryClass("player123", advanced.ClassMage) // Level 20+
manager.SetPrestigeClass("player123", advanced.PrestigeSpellblade) // Level 30+
manager.AllocateTalent("player123", talentID)
```

**15 Base Classes + 20 Prestige Classes + 90+ Talents**

---

## Modding System (V8.0)

### Package: `github.com/opd-ai/venture/pkg/modding`

```go
loader := modding.NewLoader()
mod, _ := loader.LoadFromFile("mods/hardcore-mode.json")
manager := modding.NewManager()
manager.AddMod(mod)
manager.ApplyRules(world)
```

See [MODDING_GUIDE.md](MODDING_GUIDE.md) for complete API.

---

## Additional Resources (Updated)

