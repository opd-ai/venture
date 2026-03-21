# Project Overview

Venture is a fully procedural multiplayer action-RPG built with Go and Ebiten where **every aspect of the game—graphics, audio, and gameplay content—is generated at runtime from a single binary with no external asset files**. This "zero-asset" philosophy enables infinite content variety, single-binary distribution, and cross-platform consistency.

The game combines deep procedural generation inspired by roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by The Legend of Zelda. It targets game developers, contributors, and hobbyists interested in procedural content generation, ECS architecture, and high-latency multiplayer game networking. Platforms: Linux, macOS, Windows, WebAssembly (browser), iOS, and Android.

Key features include: 100% procedural content (graphics, audio, terrain, items, quests, NPCs), player housing & guild systems with furniture placement and guildhalls, advanced physics (vehicle suspension/collision/weight transfer, fluid buoyancy/flooding/swimming, environmental destruction), cross-server federation with WebRTC/portal systems/cross-server guilds, high-latency multiplayer designed for 200–5000ms latency supporting Tor/onion services, genre-based theming (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic), experimental VR/stereoscopic support, and a sandboxed JSON-based modding system.

## Sibling Repository Context

Venture is part of the **opd-ai Procedural Game Suite**—8 sibling repositories sharing the same architectural patterns, conventions, and eventually shared library packages. All repos follow the zero-asset philosophy where every game generates all graphics, audio, and content at runtime from a single binary.

| Repo | Genre | Ebiten Version | Description |
|------|-------|----------------|-------------|
| `opd-ai/venture` | Co-op action-RPG | v2.9.3 | Top-down multiplayer action-RPG (this repo) |
| `opd-ai/vania` | Metroidvania | v2.6.3 | Procedural platformer with exploration |
| `opd-ai/velocity` | Galaga-like shooter | v2.9.8 | Vertical scrolling space shooter |
| `opd-ai/violence` | Raycasting FPS | v2.8.8 | First-person shooter with multiplayer |
| `opd-ai/way` | Battle-cart racer | v2.x | Racing game with combat |
| `opd-ai/wyrm` | Survival RPG | v2.x | First-person survival RPG |
| `opd-ai/where` | Wilderness survival | v2.x | Open-world survival game |
| `opd-ai/whack` | Arena battle | v2.x | Arena combat game |

When implementing features, follow patterns compatible with all sibling repos so code can eventually be extracted into shared libraries. The repositories share: ECS architecture patterns, deterministic seed-based generation, interface-only networking, structured logging with logrus, and the same naming conventions.

## Technical Stack

- **Primary Language**: Go 1.24.5+
- **Game Framework**: Ebiten v2.9.3 — 2D game engine with cross-platform + WASM support
- **Structured Logging**: Logrus v1.9.3 — JSON/text output with contextual fields
- **UUID Generation**: google/uuid v1.6.0 — entity and network identifiers
- **Image Processing**: golang.org/x/image v0.32.0 — procedural sprite generation
- **Text Rendering**: golang.org/x/text v0.30.0 — internationalization support
- **Native Dialogs**: ncruces/zenity v0.10.14 — file dialogs and notifications
- **Testing**: Go standard `testing` package with table-driven tests and benchmarks
- **Build/Deploy**: GNU Make, GitHub Actions CI/CD, WASM deployment to GitHub Pages

## Project Structure

This repo uses **venture-style layout**: `cmd/client`, `cmd/server`, `cmd/mobile` entry points with `pkg/` containing 30+ public library packages.

```
venture/
├── cmd/                    # Application entry points
│   ├── client/             # Desktop game client with UI systems
│   │                       # - main.go: Entry point, flag parsing, EbitenGame
│   │                       # - handlers.go: Game event handlers
│   │                       # - webrtc_wasm.go: WASM-specific WebRTC
│   ├── server/             # Dedicated multiplayer server
│   │                       # - main.go: Server initialization, system registration
│   └── mobile/             # iOS/Android entry point (thin wrapper)
├── pkg/                    # Core library packages (30+ domains)
│   ├── engine/             # ECS core, 100+ systems, 400+ files
│   │   ├── ecs.go          # Entity, World, component management
│   │   ├── components.go   # Core components (Position, Velocity, Health, etc.)
│   │   ├── interfaces.go   # Component, System, GameRunner interfaces
│   │   ├── system_init.go  # InitializeGameSystems() — critical integration point
│   │   ├── spatial_partition.go  # Spatial hash grid for entity queries
│   │   ├── physics/        # Vehicle, fluid, destruction subsystems
│   │   ├── prestige/       # New Game+ progression
│   │   └── qol/            # Quality-of-life (auto-loot, craft queue, etc.)
│   ├── procgen/            # Procedural generators (25+ subdirs)
│   │   ├── terrain/        # BSP, cellular, L-system, Voronoi, city, composite
│   │   ├── entity/         # NPC/creature generation with templates
│   │   ├── item/           # Item generation with rarity tiers
│   │   ├── quest/          # Quest generation with objectives/rewards
│   │   ├── dialog/         # Markov chain dialog generation
│   │   ├── narrative/      # Story beat and narrative arc generation
│   │   ├── genre/          # Genre blending, registry, predefined genres
│   │   ├── magic/          # Spell generation with balance calculations
│   │   ├── skills/         # Skill tree generation
│   │   └── ...             # building, furniture, vehicle, faction, legendary, etc.
│   ├── rendering/          # Runtime graphics pipeline
│   │   ├── sprites/        # Sprite generation, anatomy, equipment overlays
│   │   ├── animation/      # Articulation, caching, directional variants
│   │   ├── tiles/          # Tile transitions, parallax, wall variants
│   │   ├── lighting/       # Bloom, ambient occlusion, dynamic lights
│   │   ├── particles/      # Particle physics, weather, LOD, pooling
│   │   ├── postprocess/    # Color grading, vignette, chromatic aberration
│   │   ├── ui/             # Chat, notifications, tutorials, transitions
│   │   ├── cache/          # Sprite caching, predictive warming, memory monitor
│   │   └── pool/           # Resource pooling for sprites/images
│   ├── audio/              # Procedural audio synthesis
│   │   ├── music/          # Adaptive soundtrack, motifs, theory-based
│   │   ├── sfx/            # Sound effect generation and processing
│   │   ├── synthesis/      # Oscillators, envelopes, synthesis engine
│   │   └── voice.go        # Voice codec (ADPCM) for multiplayer
│   ├── network/            # Multiplayer networking
│   │   ├── client.go       # Game client networking
│   │   ├── server.go       # Authoritative server
│   │   ├── prediction.go   # Client-side prediction
│   │   ├── lag_compensation.go  # High-latency compensation
│   │   ├── federation/     # Cross-server discovery, auth, WebRTC, portals
│   │   ├── chat/           # Chat channels
│   │   ├── trade/          # Player trade system
│   │   └── resilience/     # Network resilience metrics/simulation
│   ├── world/              # Persistent world state
│   │   ├── housing/        # Player housing, blueprints, guildhalls
│   │   ├── economy/        # Marketplace, pricing engine, guild bank
│   │   ├── territory/      # Territory control and siege mechanics
│   │   └── raids/          # Raid generation, instances, lockouts
│   ├── integration/        # Cross-system integrations (10 subdirs)
│   │   ├── companion_housing/   # Companion home system
│   │   ├── guild_housing/       # Guild housing permissions/upgrades
│   │   ├── housing_crafting/    # Housing + crafting integration
│   │   ├── choice_consequences/ # Narrative choice tracking
│   │   └── ...
│   ├── combat/             # Damage calculation, resolver, validation
│   ├── saveload/           # Save/load with migration, WASM storage
│   ├── config/             # Configuration types and validation
│   ├── validation/         # Input validation, chat filter, rate limiting
│   ├── modding/            # JSON mod loader, sandbox execution
│   ├── security/           # Security audit (30 checks), persistence
│   └── observability/      # Prometheus metrics at /metrics endpoint
├── docs/                   # 60+ documentation files
├── examples/               # Demo programs (bloom, weather, sprites, controls)
├── scripts/                # Build, test, deployment automation
│   ├── build-*.sh          # Platform builds (Linux, macOS, Windows, mobile)
│   ├── validate-network-types.sh  # CI: Network interface enforcement
│   └── test-integration.sh # Integration test runner
├── mods/                   # Example mod configurations (JSON)
├── web/                    # WebAssembly deployment assets
└── build/                  # Platform-specific build configs
```

---

## ⚠️ CRITICAL: Complete Feature Integration (Zero Dangling Features)

**This is the single most important rule for this codebase.** Every feature, system, component, generator, and integration MUST be fully wired into the runtime. Dangling features are a maintenance nightmare, a source of deep frustration, and actively degrade code quality.

### The Dangling Feature Problem

In complex procedural game codebases with 100+ systems and 25+ generators, it is extremely common for features to be:

1. **Defined but never instantiated** — A system struct exists but `NewXxxSystem()` is never called in `cmd/client/main.go`, `cmd/server/main.go`, or `pkg/engine/system_init.go`

2. **Instantiated but never integrated** — A system runs but its output is never consumed by other systems. Example: A weather system updates internal state but no render system reads weather to apply visual effects.

3. **Partially integrated** — A system works for fantasy genre but silently no-ops for cyberpunk or horror. The generator exists but the genre dispatch table doesn't include it.

4. **Tested in isolation but broken in context** — Unit tests pass because they test the system in isolation, but the system was never wired into the actual game loop or the component it depends on is never attached to entities.

5. **Events emitted with no listeners** — An event bus emits "player.crafted_item" but no system ever registered a handler for it.

6. **Seeds not propagated** — A parent generator accepts a seed but calls sub-generators with `rand.Int63()` instead of derived deterministic seeds.

### The Integration Chain: Six Links That Must All Connect

**Before writing ANY new code, verify the full integration chain:**

```
Definition → Instantiation → Registration → Update Loop → Output → Consumer → Player Effect
```

1. **Definition → Instantiation**: Is the struct/system created at runtime?
   - Check: `grep -rn 'NewYourSystem' cmd/ pkg/engine/system_init.go`
   - Bad: Constructor exists in `your_system.go` but is never called

2. **Instantiation → Registration**: Is the system registered with the World?
   - Check: Look for `world.AddSystem(yourSystem)` in `InitializeGameSystems()`
   - Bad: System is created but `world.AddSystem()` is never called

3. **Registration → Update Loop**: Does `Update()` actually get called each frame?
   - Check: Add a log statement in `Update()` and verify it fires 60x/sec
   - Bad: System is registered but World.Update() doesn't iterate over it

4. **Update → Output**: Does the system produce observable outputs?
   - Check: Does `Update()` modify entity components, emit events, or change world state?
   - Bad: `Update()` calculates values but stores them in private fields nobody reads

5. **Output → Consumer**: Is there at least one other system that reads this output?
   - Check: `grep -rn 'GetYourComponent\|yourSystem\.' pkg/engine/`
   - Bad: System sets `WeatherComponent.RainIntensity` but no system reads it

6. **Consumer → Player Effect**: Does the chain produce something the player perceives?
   - Check: Trace from consumer to render/audio/input effect
   - Bad: Consumer reads the value but only logs it; player never sees/hears anything

**If ANY link in this chain is missing, the feature is dangling. Do not submit dangling features.**

### Specific Anti-Patterns to Reject

#### Anti-Pattern 1: System Defined But Never Instantiated

```go
// ❌ BAD: System defined but never added to the game world
// File: pkg/engine/weather_system.go
type WeatherSystem struct {
    world *World
    seed  int64
}

func NewWeatherSystem(seed int64) *WeatherSystem {
    return &WeatherSystem{seed: seed}
}

func (w *WeatherSystem) Update(entities []*Entity, dt float64) {
    // This code NEVER runs because NewWeatherSystem is never called!
}

// ✅ GOOD: System instantiated and registered in system_init.go
// File: pkg/engine/system_init.go
func InitializeGameSystems(world *World, seed int64) {
    // ...other systems...
    weatherSystem := NewWeatherSystem(seed)
    world.AddSystem(weatherSystem)
    
    // AND other systems consume weather state:
    renderSystem.SetWeatherProvider(weatherSystem)
    audioSystem.SetAmbientProvider(weatherSystem)
}
```

#### Anti-Pattern 2: Generator Never Called Outside Tests

```go
// ❌ BAD: Generator implements interface but is never called in runtime
// File: pkg/procgen/terrain/cyberpunk.go
type CyberpunkTerrainGen struct{}

func (g *CyberpunkTerrainGen) Generate(seed int64, params GenParams) *Terrain {
    // Great implementation...but only called in cyberpunk_test.go
}

// File: pkg/procgen/terrain/generator.go - Genre dispatch
var terrainGenerators = map[string]TerrainGenerator{
    "fantasy": &FantasyTerrainGen{},
    "scifi":   &SciFiTerrainGen{},
    // MISSING: "cyberpunk": &CyberpunkTerrainGen{},  // Dangling!
}

// ✅ GOOD: Generator registered in dispatch table
var terrainGenerators = map[string]TerrainGenerator{
    "fantasy":   &FantasyTerrainGen{},
    "scifi":     &SciFiTerrainGen{},
    "cyberpunk": &CyberpunkTerrainGen{},  // Properly registered
    "horror":    &HorrorTerrainGen{},
    "postapoc":  &PostApocTerrainGen{},
}
```

#### Anti-Pattern 3: Events Emitted Without Handlers

```go
// ❌ BAD: Event emitted but no listener handles it
// File: pkg/engine/progression_system.go
func (s *ProgressionSystem) OnLevelUp(entity *Entity, newLevel int) {
    s.eventBus.Emit("player.levelup", entity.ID, newLevel)
    // No system ever calls eventBus.On("player.levelup", ...)
}

// ✅ GOOD: Event has both emitter and handler
// File: pkg/engine/progression_system.go
func (s *ProgressionSystem) OnLevelUp(entity *Entity, newLevel int) {
    s.eventBus.Emit("player.levelup", entity.ID, newLevel)
}

// File: pkg/engine/achievement_system.go
func (s *AchievementSystem) init() {
    s.eventBus.On("player.levelup", func(entityID uint64, level int) {
        s.CheckLevelAchievements(entityID, level)
    })
}

// File: pkg/engine/hud_system.go
func (s *HUDSystem) init() {
    s.eventBus.On("player.levelup", func(entityID uint64, level int) {
        s.ShowLevelUpAnimation(entityID, level)
    })
}
```

#### Anti-Pattern 4: Seed Not Propagated

```go
// ❌ BAD: Seed accepted but not forwarded (causes non-determinism)
func GenerateWorld(seed int64) *World {
    terrain := generateTerrain(rand.Int63())  // BUG: Ignores input seed!
    items := generateItems(rand.Int63())      // BUG: Different every run!
    return &World{terrain, items}
}

// ✅ GOOD: Derived seeds for deterministic hierarchy
func GenerateWorld(seed int64) *World {
    // XOR with magic bytes creates unique but deterministic sub-seeds
    terrainSeed := seed ^ 0x54455252  // "TERR" in hex
    itemSeed := seed ^ 0x4954454D     // "ITEM" in hex
    entitySeed := seed ^ 0x454E5459   // "ENTY" in hex
    questSeed := seed ^ 0x51554553    // "QUES" in hex
    
    terrain := generateTerrain(terrainSeed)
    items := generateItems(itemSeed)
    entities := generateEntities(entitySeed, terrain)
    quests := generateQuests(questSeed, entities)
    
    return &World{terrain, items, entities, quests}
}
```

### Integration Verification Checklist

Run these checks before every PR:

```bash
# 1. Every constructor has at least one non-test caller
grep -rn 'func New' --include='*.go' pkg/ | grep -v _test.go | \
  while read line; do
    func=$(echo "$line" | sed 's/.*func \(New[^(]*\).*/\1/')
    callers=$(grep -rn "$func" --include='*.go' pkg/ cmd/ | grep -v _test.go | grep -v "func $func" | wc -l)
    if [ "$callers" -eq 0 ]; then echo "DANGLING: $func"; fi
  done

# 2. All TODOs are tracked in GAPS.md or ROADMAP.md
grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' pkg/

# 3. No empty method bodies in non-test files
grep -Pzo 'func.*\{\s*\}' --include='*.go' pkg/ | grep -v _test.go

# 4. Run feature audit
make feature-audit

# 5. Verify system registration
grep -c 'world.AddSystem' pkg/engine/system_init.go
# Should match number of systems in codebase

# 6. Check for seeds not propagated
grep -rn 'rand.Int63()' --include='*.go' pkg/procgen/
# Each hit should be reviewed for determinism
```

### Known Gaps (see GAPS.md and ROADMAP.md)

Current documented gaps that need resolution:

- **Gap 1**: Signal handler integration test infrastructure missing — no test validates graceful shutdown on SIGTERM
- **Gap 3**: Trade validation lacks per-item quantity concept — `TradeValidator` exists but trade model has no quantity field
- **Gap 4**: `pkg/memprofile/profile.go` uses `fmt.Printf` — decision needed on exemption from structured logging
- **Gap 5**: No automated flag-to-documentation sync — new CLI flags can be added without updating docs
- **ROADMAP Priority 3**: No automated FPS regression testing in CI
- **ROADMAP Priority 4**: No automated memory budget validation in CI
- **ROADMAP Priority 2**: Health check endpoints (`/healthz`, `/readyz`) not implemented

---

## Networking Best Practices (MANDATORY)

### Interface-Only Network Types (Hard Constraint)

When declaring network variables, **ALWAYS** use interface types. This is a **non-negotiable project rule** enforced by `scripts/validate-network-types.sh` in CI.

| ❌ Never Use (Concrete Type) | ✅ Always Use (Interface Type) |
|------------------------------|-------------------------------|
| `*net.UDPAddr` | `net.Addr` |
| `*net.IPAddr` | `net.Addr` |
| `*net.TCPAddr` | `net.Addr` |
| `*net.UDPConn` | `net.PacketConn` |
| `*net.TCPConn` | `net.Conn` |
| `*net.UDPListener` | `net.Listener` |
| `*net.TCPListener` | `net.Listener` |
| `*net.UnixAddr` | `net.Addr` |
| `*net.UnixConn` | `net.Conn` |
| `*net.UnixListener` | `net.Listener` |

```go
// ✅ GOOD: Interface types throughout
func handleConnection(conn net.Conn, remoteAddr net.Addr) error {
    // Works with TCP, Unix, or any net.Conn implementation
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        return fmt.Errorf("read from %v: %w", remoteAddr, err)
    }
    return nil
}

func handlePacket(conn net.PacketConn) error {
    buf := make([]byte, 1024)
    n, addr, err := conn.ReadFrom(buf)  // Interface method
    if err != nil {
        return err
    }
    logrus.WithField("from", addr.String()).Debug("received packet")
    return nil
}

// ❌ BAD: Concrete types — will fail CI validation
func handleUDP(conn *net.UDPConn, addr *net.UDPAddr) {
    conn.ReadFromUDP(buf)  // Tied to UDP implementation
}
```

**Never use type switches or type assertions to access concrete network methods:**

```go
// ❌ BAD: Type assertion breaks interface abstraction and testability
if udpConn, ok := conn.(*net.UDPConn); ok {
    udpConn.ReadFromUDP(buf)
    udpConn.SetReadBuffer(256 * 1024)
}

// ✅ GOOD: Use interface methods; configure at creation time
n, addr, err := conn.ReadFrom(buf)  // PacketConn interface
```

This constraint enables:
- Mock implementations for unit testing without real network I/O
- Future transport swaps (WebRTC, QUIC) without code changes
- Consistent error handling across transport types

### High-Latency Network Design (200–5000ms)

All multiplayer networking code MUST function correctly under **200–5000ms round-trip latency**. The game explicitly targets Tor/onion services, satellite internet, and intercontinental connections.

#### Mandatory Design Principles

1. **Client-Side Prediction**: The client simulates game state locally and reconciles with server authoritative state when it arrives. Never block the game loop waiting for a server response.

2. **State Interpolation/Extrapolation**: Remote entity positions interpolate between known server states. When packets are delayed beyond the interpolation window, extrapolate using last-known velocity. See `pkg/network/prediction.go`.

3. **Jitter Buffers**: Incoming state updates buffer and play back at consistent rate, absorbing latency variance. Design for ±500ms jitter tolerance minimum.

4. **Idempotent Messages**: Every network message must be safe to process multiple times. Retransmission is expected, not exceptional. Use sequence numbers to detect/ignore duplicates.

5. **No Synchronous RPC in Game Loops**: Never issue a blocking network call inside `Update()` or `Draw()`. All network I/O is asynchronous via channels or callbacks.

6. **Graceful Degradation**: At 5000ms latency the game must remain playable—reduce update frequency, increase prediction windows, hide latency with animations.

7. **Timeout Tolerance**: Connection timeouts ≥10 seconds. Disconnect detection uses heartbeat absence over sliding window (≥3 missed heartbeats at expected interval), never a single missed packet.

```go
// ❌ BAD: Tight timeout drops players on satellite connections
conn.SetReadDeadline(time.Now().Add(1 * time.Second))

// ✅ GOOD: Generous timeout for high-latency environments
conn.SetReadDeadline(time.Now().Add(10 * time.Second))

// ❌ BAD: Blocking RPC in game loop freezes entire game
func (g *Game) Update() error {
    state, err := g.server.GetWorldState()  // BLOCKS until response!
    g.world = state
    return nil
}

// ✅ GOOD: Async receive with interpolation
func (g *Game) Update() error {
    // Non-blocking check for new server state
    select {
    case state := <-g.stateChannel:
        g.interpolator.PushServerState(state)
    default:
        // No new state this frame — continue with local prediction
    }
    
    // Always use interpolated/predicted state, never wait
    g.world = g.interpolator.GetInterpolatedState(time.Now())
    return nil
}
```

#### Latency Budget Allocation (60 FPS = 16.6ms per frame)

- Input processing: ≤1ms
- Local simulation/prediction: ≤4ms
- State interpolation: ≤1ms
- Network send (non-blocking enqueue): ≤0.5ms
- Rendering: ≤10ms
- Network I/O goroutines: Run independently, never counted against frame budget

---

## Code Assistance Guidelines

### 1. ECS Architecture Discipline

Components are **pure data structures** with only a `Type() string` method. **NO behavior or logic in components.**

```go
// ✅ GOOD: Component is pure data
type PositionComponent struct {
    X, Y         float64
    PrevX, PrevY float64  // Previous tick for render interpolation
}

func (p *PositionComponent) Type() string { return "position" }

// Optional: Serialize/Deserialize for save/load
func (p *PositionComponent) Serialize() ([]byte, error) { /* ... */ }
func (p *PositionComponent) Deserialize(data []byte) error { /* ... */ }

// ❌ BAD: Logic in component (violates ECS principle)
func (p *PositionComponent) Move(dx, dy float64) { 
    p.X += dx; p.Y += dy  // This belongs in MovementSystem!
}
```

Systems contain ALL game logic and operate on entity collections:

```go
type MovementSystem struct {
    world        *World
    spatialGrid  *SpatialPartition
}

func NewMovementSystem(world *World) *MovementSystem {
    logrus.WithField("system", "movement").Debug("Creating movement system")
    return &MovementSystem{
        world:       world,
        spatialGrid: NewSpatialPartition(64), // 64px cell size
    }
}

func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
    for _, entity := range entities {
        // Use hot-path cached accessors (~93x faster than map lookup)
        pos := entity.GetPosition()
        vel := entity.GetVelocity()
        if pos == nil || vel == nil {
            continue
        }
        
        // Store previous position for render interpolation
        pos.PrevX, pos.PrevY = pos.X, pos.Y
        
        // Apply velocity
        pos.X += vel.VX * deltaTime
        pos.Y += vel.VY * deltaTime
        
        // Update spatial partition for collision queries
        s.spatialGrid.Update(entity)
    }
}
```

Entity hot-path caching provides ~93x faster component access for critical paths:

```go
// Cached accessors (use these in hot paths)
entity.GetPosition()       // *PositionComponent
entity.GetVelocity()       // *VelocityComponent
entity.GetHealth()         // *HealthComponent
entity.GetCollider()       // *ColliderComponent
entity.GetSprite()         // *EbitenSprite
entity.GetRotation()       // *RotationComponent
entity.GetAnimation()      // *AnimationComponent
entity.GetVisualFeedback() // *VisualFeedbackComponent

// Generic accessor (use when cached version unavailable)
comp := entity.GetComponent("custom_component")
```

### 2. Deterministic Procedural Generation

All content generation MUST be deterministic and seed-based. **Same seed = identical output** across all platforms and runs.

```go
// ✅ GOOD: Explicit seed-based RNG, never global
func GenerateTerrain(seed int64, params TerrainParams) *Terrain {
    rng := rand.New(rand.NewSource(seed))
    
    tiles := make([][]Tile, params.Width)
    for x := 0; x < params.Width; x++ {
        tiles[x] = make([]Tile, params.Height)
        for y := 0; y < params.Height; y++ {
            tiles[x][y] = generateTile(rng, x, y, params.Biome)
        }
    }
    
    logrus.WithFields(logrus.Fields{
        "seed":   seed,
        "width":  params.Width,
        "height": params.Height,
        "biome":  params.Biome,
    }).Debug("Terrain generated")
    
    return &Terrain{Tiles: tiles, Seed: seed}
}

// ✅ GOOD: Derived seeds for sub-generators (deterministic hierarchy)
func GenerateWorld(seed int64, genre string) *World {
    // XOR with magic bytes creates unique but deterministic sub-seeds
    terrainSeed := seed ^ 0x54455252  // "TERR"
    entitySeed := seed ^ 0x454E5459   // "ENTY"
    itemSeed := seed ^ 0x4954454D     // "ITEM"
    questSeed := seed ^ 0x51554553    // "QUES"
    
    terrain := GenerateTerrain(terrainSeed, getTerrainParams(genre))
    entities := GenerateEntities(entitySeed, terrain, genre)
    items := GenerateItems(itemSeed, terrain, genre)
    quests := GenerateQuests(questSeed, entities, genre)
    
    return &World{
        Seed:     seed,
        Genre:    genre,
        Terrain:  terrain,
        Entities: entities,
        Items:    items,
        Quests:   quests,
    }
}

// ❌ BAD: Global rand (non-deterministic, not thread-safe)
func Generate() int {
    return rand.Intn(100)  // Uses global state — different each run!
}

// ❌ BAD: Time-based seeding breaks reproducibility
func GenerateBad() *World {
    rng := rand.New(rand.NewSource(time.Now().UnixNano()))
    // Cannot reproduce this world later!
}
```

### 3. Structured Logging with Logrus

Always use `logrus.WithFields()` for contextual, searchable logs:

```go
// ✅ GOOD: Structured logging with standard field names
logrus.WithFields(logrus.Fields{
    "system":       "terrain",
    "seed":         seed,
    "genre":        genre,
    "width":        width,
    "height":       height,
    "duration_ms":  elapsed.Milliseconds(),
}).Info("Terrain generation complete")

logrus.WithFields(logrus.Fields{
    "entity":         entity.ID,
    "component_type": "position",
    "x":              pos.X,
    "y":              pos.Y,
}).Debug("Entity moved")

logrus.WithFields(logrus.Fields{
    "player":  playerID,
    "item":    itemID,
    "slot":    slotIndex,
}).Info("Item equipped")

// ❌ BAD: Unstructured logging (hard to search, inconsistent format)
fmt.Printf("Generated terrain with seed %d in %dms\n", seed, elapsed)
log.Println("terrain done")
```

**Standard field names** (use consistently):
- `system` — system name (e.g., "terrain", "combat", "network")
- `entity` — entity ID
- `player` — player ID/name
- `seed` — generation seed
- `genre` — genre identifier
- `error` — error value
- `duration` / `duration_ms` — timing
- `count` — quantity
- `component_type` — component identifier

### 4. Performance Requirements

- **Target**: 60 FPS minimum on mid-range hardware
- **Memory**: <500MB client, <1GB server (8 players)
- **Entity queries**: Use `pkg/engine/spatial_partition.go` for collections >100 entities
- **Sprite caching**: Never regenerate the same sprite twice per session (see `pkg/rendering/cache/`)
- **Object pooling**: Use pools for bullets, particles, status effects (see `pkg/engine/projectile_pool.go`, `status_effect_pool.go`)
- **Benchmarks**: Run `go test -bench=. -benchmem` for hot paths before submitting performance-sensitive code

### 5. Zero External Assets

**ALL content is generated at runtime.** Never add asset files.

- **Graphics**: Procedurally generated via `pkg/rendering/sprites/`, `tiles/`, `particles/`
- **Audio**: Synthesized via `pkg/audio/synthesis/` oscillators, envelopes, effects
- **Levels/Maps**: Generated via `pkg/procgen/terrain/` (BSP, cellular, L-systems, Voronoi)
- **Items/NPCs/Quests**: Generated via `pkg/procgen/entity/`, `item/`, `quest/`
- **UI**: Built from code via `pkg/rendering/ui/`

### 6. Error Handling

Return errors up the call stack. **Never panic in game/library code.**

```go
// ✅ GOOD: Return errors with context
func GenerateTerrain(seed int64, params TerrainParams) (*Terrain, error) {
    if params.Width <= 0 || params.Height <= 0 {
        return nil, fmt.Errorf("invalid terrain dimensions: %dx%d", params.Width, params.Height)
    }
    if seed == 0 {
        return nil, errors.New("terrain generation requires non-zero seed")
    }
    // ... generation logic
    return terrain, nil
}

// ✅ GOOD: Handle errors gracefully with fallback
func (s *TerrainSystem) Update(entities []*Entity, dt float64) {
    if s.terrain == nil {
        terrain, err := GenerateTerrain(s.seed, s.params)
        if err != nil {
            logrus.WithError(err).Error("Terrain generation failed, using fallback")
            s.terrain = s.createFallbackTerrain()
            return
        }
        s.terrain = terrain
    }
}

// ❌ BAD: Panic in library code crashes the game
func GenerateTerrain(seed int64) *Terrain {
    if seed == 0 {
        panic("zero seed")  // NEVER panic in game logic!
    }
}
```

Panics are acceptable ONLY in `main()` for unrecoverable startup failures (missing required config, etc.).

### 7. Table-Driven Tests

Target ≥40% coverage per package (≥30% for Ebiten-dependent packages requiring xvfb):

```go
func TestGenerateTerrain(t *testing.T) {
    tests := []struct {
        name    string
        seed    int64
        width   int
        height  int
        wantErr bool
    }{
        {"valid params", 12345, 100, 100, false},
        {"zero seed", 0, 100, 100, true},
        {"negative seed allowed", -1, 100, 100, false},
        {"large world", 99999, 1000, 1000, false},
        {"zero width", 12345, 0, 100, true},
        {"zero height", 12345, 100, 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            terrain, err := GenerateTerrain(tt.seed, TerrainParams{
                Width: tt.width, Height: tt.height,
            })
            
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateTerrain() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                if terrain == nil {
                    t.Error("GenerateTerrain() returned nil without error")
                }
                if len(terrain.Tiles) != tt.width {
                    t.Errorf("terrain width = %d, want %d", len(terrain.Tiles), tt.width)
                }
            }
        })
    }
}

// Test determinism: same seed must produce identical output
func TestGenerateTerrain_Deterministic(t *testing.T) {
    seed := int64(42)
    params := TerrainParams{Width: 50, Height: 50}
    
    terrain1, _ := GenerateTerrain(seed, params)
    terrain2, _ := GenerateTerrain(seed, params)
    
    for x := 0; x < params.Width; x++ {
        for y := 0; y < params.Height; y++ {
            if terrain1.Tiles[x][y] != terrain2.Tiles[x][y] {
                t.Fatalf("Non-deterministic at (%d,%d): %v vs %v", 
                    x, y, terrain1.Tiles[x][y], terrain2.Tiles[x][y])
            }
        }
    }
}
```

Use stub implementations for testing without Ebiten runtime:
- `StubInput` — mock input provider
- `StubSprite` — mock sprite without *ebiten.Image
- `StubImage` — mock image provider

---

## Cross-Repository Code Sharing Patterns

### Shared Pattern Catalog

When implementing features, follow these patterns for future extraction into shared libraries:

| Pattern | Package Location | Used By |
|---------|-----------------|---------|
| ECS core (World, Entity, Component, System) | `pkg/engine/ecs.go`, `interfaces.go` | All 8 repos |
| Procedural generation framework | `pkg/procgen/generator.go` | All 8 repos |
| Seed management & derivation | XOR patterns inline | All 8 repos |
| Sprite/tile generation | `pkg/rendering/sprites/`, `tiles/` | All 8 repos |
| Audio synthesis | `pkg/audio/synthesis/` | All 8 repos |
| Input handling | `pkg/engine/input_system.go` | All 8 repos |
| Camera systems | `pkg/engine/camera_system.go` | All 8 repos |
| Particle systems | `pkg/rendering/particles/` | venture, vania, violence |
| Save/load persistence | `pkg/saveload/` | All 8 repos |
| Configuration (CLI flags) | `pkg/config/` | venture, violence, velocity |
| Networking (multiplayer) | `pkg/network/` | venture, violence |

### Guidelines for Shareable Code

1. **Minimal dependencies**: Shared packages depend only on stdlib + Ebiten
2. **Interfaces at boundaries**: Define interfaces for game-specific behavior
3. **Parameterize, don't specialize**: Generators accept parameters for any genre
4. **Identical interfaces across repos**:
   - Component: `Type() string`
   - System: `Update(entities []*Entity, deltaTime float64)`

---

## Quality Standards

### Testing Requirements
- **Coverage**: ≥40% per package (≥30% for display-dependent packages needing xvfb)
- **Race detection**: `go test -race ./...` must pass
- **Benchmarks**: Required for rendering, physics, generation hot paths

### Code Review Quality Gates
- Build success (`make build`)
- All tests pass (`make test`)
- Race-free (`make test-race`)
- Static analysis (`go vet ./...`)
- Network type validation (`scripts/validate-network-types.sh`)
- No new TODO/FIXME without GAPS.md entry

### Makefile Targets
| Target | Description |
|--------|-------------|
| `make build` | Build client and server |
| `make test` | Run all tests |
| `make test-coverage` | Tests with coverage report |
| `make test-race` | Tests with race detection |
| `make lint` | `go vet` + network validation |
| `make bench` | Run benchmarks |
| `make quality` | All quality validations |
| `make feature-audit` | Check for instantiation gaps |
| `make visual-regression` | Visual regression tests |
| `make balance-validate` | Combat/economic balance check |

---

## Naming Conventions

- **Packages**: lowercase, single-word when possible (`engine`, `procgen`, `audio`, `render`)
- **Files**: snake_case (`terrain_generator.go`, `combat_system.go`)
- **Types**: PascalCase (`TerrainGenerator`, `CombatSystem`, `HealthComponent`)
- **Interfaces**: PascalCase, often ending in `-er` (`Generator`, `Renderer`, `Provider`)
- **Components**: PascalCase + "Component" suffix (`HealthComponent`, `PositionComponent`)
- **Systems**: PascalCase + "System" suffix (`CombatSystem`, `RenderSystem`, `MovementSystem`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Seeds**: Always `int64`, always named `seed` in function parameters

## GAPS.md Protocol

When identifying potential gaps during development:
1. Note it in your response with severity assessment
2. Suggest adding to GAPS.md with severity (Critical/High/Medium/Low)
3. Include file path and line number if applicable
4. Propose actionable fix or investigation steps

Always reference `GAPS.md` and `ROADMAP.md` for known issues before implementing related features.
