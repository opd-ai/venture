# Project Overview

Venture is a fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, and gameplay content—is generated at runtime with no external asset files, resulting in a single binary distribution. The game combines deep procedural generation inspired by roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda.

The project targets game developers, contributors, and hobbyists interested in procedural content generation, ECS architecture, and multiplayer game networking. It supports desktop (Linux, macOS, Windows), WebAssembly (browser), and mobile (iOS, Android) platforms.

## Key Features

- **100% Procedural Content**: All graphics, audio, terrain, items, quests, and NPCs generated at runtime
- **Player Housing & Guild Systems**: Persistent housing with furniture placement, guildhalls, and guild management
- **Advanced Physics**: Vehicle physics (suspension, collision, weight transfer), fluid simulation (buoyancy, flooding, swimming), and environmental destruction
- **Cross-Server Federation**: Federated server architecture with WebRTC support, portal systems, and cross-server guilds
- **High-Latency Multiplayer**: Designed for 200-5000ms latency (supports Tor/onion services)
- **Genre-Based Theming**: Dynamic content generation based on genre (fantasy, sci-fi, horror, cyberpunk)
- **VR/Stereoscopic Support**: VR controller integration and stereoscopic rendering
- **Modding System**: Sandboxed, JSON-based rule mods for data-driven balance/content tweaks (validated with no executable code)

The codebase follows an Entity-Component-System (ECS) architecture where entities are unique identifiers with component collections, components are pure data structures with no behavior, and systems contain all logic operating on entities with specific components. This separation enables data-oriented design, efficient caching, and easy testing.

## Technical Stack

- **Primary Language**: Go 1.24.5+
- **Game Framework**: Ebiten v2.9.3 (2D game engine with cross-platform support including WASM)
- **Logging**: Logrus v1.9.3 (structured logging with JSON/text output)
- **UUID Generation**: google/uuid v1.6.0 (entity and network IDs)
- **Image Processing**: golang.org/x/image v0.32.0 (procedural sprite generation)
- **Dialogs**: ncruces/zenity v0.10.14 (native dialogs)
- **Testing**: Go's built-in testing package with table-driven tests and benchmarks
- **Build/Deploy**: Go build, GitHub Actions CI/CD, WebAssembly deployment to GitHub Pages

## Code Assistance Guidelines

1. **Maintain ECS Architecture Strictly**: Components must be pure data structures with only a `Type() string` method. Never add behavior/logic to components. All game logic belongs in Systems that operate on entities with specific components.
   ```go
   // ✅ GOOD: Component is pure data
   type PositionComponent struct {
       X, Y float64
   }
   func (p *PositionComponent) Type() string { return "position" }
   
   // ❌ BAD: Logic in component
   func (p *PositionComponent) Move(dx, dy float64) { p.X += dx; p.Y += dy }
   ```

2. **Enforce Deterministic Generation**: All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()`, global `math/rand` functions, or system-dependent randomness. Always use `rand.New(rand.NewSource(seed))` to ensure same seed = same output.
   ```go
   // ✅ GOOD: Deterministic
   func Generate(seed int64) {
       rng := rand.New(rand.NewSource(seed))
       value := rng.Intn(100)
   }
   
   // ❌ BAD: Non-deterministic
   func Generate() {
       value := rand.Intn(100) // Uses global random state
   }
   ```

3. **Use Structured Logging with Logrus**: Always use `logrus.Fields` for contextual logging instead of string formatting. Use standard field names: `seed`, `genre`, `entityID`, `playerID`, `system_name`, `component_type`.
   ```go
   logger.WithFields(logrus.Fields{
       "entityID": id,
       "x": x, "y": y,
   }).Info("player moved")
   ```

4. **Follow Interface-Based Network Design**: Use interface types for network variables to enhance testability. Use `net.Addr` (not `net.UDPAddr`/`net.TCPAddr`), `net.PacketConn` (not `net.UDPConn`), `net.Conn` (not `net.TCPConn`), and `net.Listener` (not specific listener types). Avoid type switches/assertions to concrete types.

5. **Maintain Performance Targets**: Target 60 FPS minimum, <500MB client memory, <1GB server memory (4 players). Use spatial partitioning for collision detection, sprite caching for rendering, and object pooling for frequently allocated objects.

6. **Write Table-Driven Tests**: Target ≥40% code coverage per package (≥30% for packages depending on X11/Wayland/Ebiten). Use Go's built-in testing with table-driven test patterns. Include benchmarks for performance-critical code. Use stub implementations (StubInput, StubSprite) for testing without Ebiten runtime.
   ```go
   func TestGenerator(t *testing.T) {
       tests := []struct {
           name    string
           seed    int64
           wantErr bool
       }{
           {"valid", 12345, false},
           {"edge case", 0, false},
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test implementation
           })
       }
   }
   ```

7. **No External Assets Allowed**: All visual and audio content must be generated at runtime using procedural algorithms. This ensures single binary distribution and infinite content variety within generation rules.

## Project Context

- **Domain**: Procedural action-RPG with multiplayer support. Key concepts include ECS entities/components/systems, deterministic procedural generation with seeds, genre-based theming (fantasy, sci-fi, horror, cyberpunk), and authoritative server networking with client-side prediction.

- **Architecture**: Entity-Component-System (ECS) pattern with:
  - `World`: Central ECS container managing entities and systems
  - `Entity`: Lightweight containers with unique IDs and component collections (with hot-path caching for position, velocity, health, collider, sprite, rotation, etc.)
  - `Component`: Pure data structures implementing `Type() string`
  - `System`: Logic processors with `Update(entities []*Entity, deltaTime float64)`

- **Configuration**: Use CLI flags for runtime configuration (`-width`, `-height`, `-seed`, `-genre`, `-port`, `-high-latency`). Environment variables for logging: `LOG_LEVEL` (debug/info/warn/error), `LOG_FORMAT` (json/text).

## Directory Architecture

### Root Structure

```
venture/
├── cmd/           # Application entry points
├── pkg/           # Core packages (30+ domains)
├── docs/          # Documentation (60+ guides)
├── examples/      # Demo programs
├── scripts/       # Build & deployment scripts
├── mods/          # Example mod configurations
├── web/           # WASM deployment assets
├── build/         # Platform-specific build configs
└── Formula/       # Homebrew formula
```

### Command Packages (`cmd/`)

| Directory | Description |
|-----------|-------------|
| `cmd/client/` | Desktop game client entry point. Implements `EbitenGame`, state management (main menu, gameplay, settings), all UI systems (inventory, quest, map, housing, guild, crafting, trade, mail). Handles lazy initialization of systems and WASM-specific WebRTC. |
| `cmd/server/` | Dedicated multiplayer server. Manages player connections, entity spawning, authoritative game state, network snapshots. Supports V4-V9 system architectures with validation layers. |
| `cmd/mobile/` | Mobile platform entry point for iOS/Android. Thin wrapper using Ebiten's mobile build support. |

### Engine Package (`pkg/engine/`)

The core game engine containing 400+ files with ECS implementation, all game systems, and components.

#### Core ECS
- `ecs.go` - Entity/World management with component caching for hot-path optimization
- `components.go` - Core component definitions (Position, Velocity, Health, Collider, Stats, etc.)
- `interfaces.go` - System, Component, and Input interfaces
- `spatial_partition.go` - Spatial hash grid for efficient collision/range queries

#### Game Systems (100+ systems)
| System Category | Key Systems |
|-----------------|-------------|
| **Movement & Physics** | `movement.go`, `collision.go`, `collision_precise.go`, `projectile_system.go`, `vehicle_system.go`, `mounting_system.go` |
| **Combat** | `combat_system.go`, `player_combat_system.go`, `spell_casting.go`, `spell_effect_system.go`, `spell_combination_system.go`, `status_effect_system.go` |
| **AI & Behavior** | `ai_system.go`, `behavior_tree_system.go`, `behavior_tree_nodes.go`, `squad_system.go`, `companion_ai_system.go` |
| **Rendering** | `render_system.go`, `animation_system.go`, `particle_system.go`, `lighting_system.go`, `shadow_system.go`, `post_processor.go` |
| **UI Systems** | `menu_system.go`, `hud_system.go`, `inventory_ui.go`, `quest_ui.go`, `shop_ui.go`, `crafting_ui.go`, `trade_ui.go`, `guild_ui.go`, housing UI (`pkg/world/housing/ui.go` via `HousingUIProvider` in `interfaces.go`) |
| **Progression** | `progression_system.go`, `skill_progression_system.go`, `achievement.go`, `class_progression_system.go`, `reputation_system.go` |
| **Social** | `chat_system.go`, `mail_system.go`, `trade_system.go`, `guild_system.go`, `faction_system.go` |
| **World** | `weather_system.go`, `terrain_modification_system.go`, `world_events_system.go`, `city_evolution_system.go`, `economy_system.go` |
| **Narrative** | `narrative_system.go`, `dialog_system.go`, `branching_narrative_system.go`, `quest_tracker.go`, `investigation_system.go` |
| **Multiplayer** | `network_components.go`, `matchmaking_system.go`, `pvp_rating_system.go`, `tournament_system.go` |
| **Quality of Life** | `qol/` - Auto-loot, craft queue, mount whistle, recipe tracker, storage sorter |
| **Prestige** | `prestige/` - New Game+ and prestige progression systems |

#### Physics Subsystems (`pkg/engine/physics/`)
| Subsystem | Description |
|-----------|-------------|
| `physics/fluids/` | Fluid simulation with buoyancy calculator, flooding mechanics, swimming system |
| `physics/destruction/` | Environmental destruction system with debris and damage propagation |
| `physics/vehicle/` | Vehicle physics: suspension, weight transfer, terrain deformation, collision response |

### Procedural Generation (`pkg/procgen/`)

Deterministic content generation with seed-based algorithms.

| Subdirectory | Description |
|--------------|-------------|
| `procgen/terrain/` | Terrain generation: BSP dungeons, cellular automata caves, L-system forests, Voronoi biomes, city generation, composite multi-level dungeons, async loading |
| `procgen/entity/` | NPC and creature generation with templates, merchants, and genre-specific variants |
| `procgen/item/` | Item generation with class restrictions, rarity tiers, stat scaling |
| `procgen/quest/` | Quest generation with objectives, rewards, and progression curves |
| `procgen/magic/` | Spell and magic system generation with balance calculations |
| `procgen/skills/` | Skill tree generation with templates and progression |
| `procgen/building/` | Building and structure generation |
| `procgen/furniture/` | Furniture generation with placement algorithms |
| `procgen/genre/` | Genre system: blending, predefined genres, registry |
| `procgen/dialog/` | Dialog generation with Markov chains, personality, corpus management |
| `procgen/narrative/` | Story beat and narrative arc generation |
| `procgen/story/` | Story generators: archaeology, branching paths, cross-dungeon stories, timelines |
| `procgen/faction/` | Faction generation with relationships |
| `procgen/companion/` | Companion/pet generation |
| `procgen/environment/` | Environmental detail generation and placement |
| `procgen/vehicle/` | Vehicle generation with combat and visual variants |
| `procgen/legendary/` | Legendary item and quest generation |
| `procgen/minigame/` | Mini-game generation (games/, factory, state machine) |
| `procgen/puzzle/` | Puzzle generation with solver |
| `procgen/class/` | Class and multiclass generation |
| `procgen/book/` | In-game book content generation |
| `procgen/station/` | Crafting station generation |
| `procgen/recipe/` | Recipe generation |

### Rendering Pipeline (`pkg/rendering/`)

Runtime procedural graphics generation.

| Subdirectory | Description |
|--------------|-------------|
| `rendering/sprites/` | Sprite generation: anatomy templates, equipment overlays, silhouettes, projectiles, animation, caching, pooling |
| `rendering/animation/` | Animation system: articulation, caching, controller, directional variants |
| `rendering/tiles/` | Tile generation: parallax, wall variants, transitions |
| `rendering/lighting/` | Lighting system: bloom, ambient occlusion, dynamic lights |
| `rendering/postprocess/` | Post-processing: chromatic aberration, color grading, depth blur, motion blur, vignette |
| `rendering/particles/` | Particle system: behaviors, physics, LOD, weather effects, pooling |
| `rendering/ui/` | UI generation: chat, decorations, hierarchy, notifications, quick travel, transitions, tutorial |
| `rendering/palette/` | Color palette generation: gradients, time-of-day |
| `rendering/patterns/` | Pattern generation for textures |
| `rendering/cache/` | Sprite caching, predictive warming, pre-generation, memory monitoring |
| `rendering/pool/` | Resource pooling for sprites and images |
| `rendering/parallel/` | Parallel rendering utilities |
| `rendering/quality/` | Quality settings and LOD management |
| `rendering/display/` | Display configuration |

### Audio Pipeline (`pkg/audio/`)

Runtime procedural audio synthesis.

| Subdirectory | Description |
|--------------|-------------|
| `audio/music/` | Music generation: adaptive soundtrack, motifs, theory-based composition |
| `audio/sfx/` | Sound effects: generator, variety manager, processing |
| `audio/synthesis/` | Audio synthesis: oscillators, envelopes, engine |

### Network Layer (`pkg/network/`)

Multiplayer networking with high-latency support.

| Subdirectory | Description |
|--------------|-------------|
| `network/` | Core: client/server, protocol, packets, compression, crypto, lag compensation, prediction, snapshot system, desync detection |
| `network/federation/` | Cross-server federation: discovery, auth, handshake, sync, transfer, portal, circuit breaker, connection pooling, retry logic |
| `network/federation/guild/` | Cross-server guild management |
| `network/federation/mobile/` | Mobile-specific federation |
| `network/federation/webrtc/` | WebRTC peer connections |
| `network/federation/market/` | Cross-server marketplace |
| `network/chat/` | Chat system with channels |
| `network/trade/` | Trade system between players |
| `network/resilience/` | Network resilience: metrics, simulator |

### World Management (`pkg/world/`)

Persistent world state and territory.

| Subdirectory | Description |
|--------------|-------------|
| `world/` | Core: state, persistence, chunk loading/compression/modification, metagame, ranking |
| `world/housing/` | Housing system: blueprints, guildhalls, spatial management, persistence, UI |
| `world/economy/` | Economy: marketplace, pricing engine, guild bank |
| `world/territory/` | Territory control: manager, siege mechanics |
| `world/raids/` | Raid system: generator, instances, lockouts, mechanics, manager |

### Integration Packages (`pkg/integration/`)

Cross-system feature integrations.

| Subdirectory | Description |
|--------------|-------------|
| `integration/companion_housing/` | Companion/pet home system: bedding, training areas, storage |
| `integration/guild_housing/` | Guild housing: permissions, transactions, upgrades |
| `integration/guild_vehicle/` | Guild fleet management |
| `integration/housing_crafting/` | Housing + crafting integration |
| `integration/choice_consequences/` | Narrative choice tracking and consequences |
| `integration/narrative_world/` | Narrative + world state integration |
| `integration/political_warfare/` | Political and faction warfare |
| `integration/trade_routes/` | Trade route management |
| `integration/world_events/` | World event management |

### Supporting Packages

| Package | Description |
|---------|-------------|
| `pkg/combat/` | Combat resolver: damage calculation, interfaces, validation |
| `pkg/saveload/` | Save/load system: manager, migrator, recovery, WASM storage support |
| `pkg/config/` | Configuration types and validation |
| `pkg/validation/` | Input validation: chat, rate limiting, trade |
| `pkg/errors/` | Error types, correlation IDs, helpers |
| `pkg/logging/` | Structured logging utilities |
| `pkg/recovery/` | Panic recovery handlers |
| `pkg/stability/` | Stability monitoring |
| `pkg/observability/` | Metrics and observability |
| `pkg/security/` | Security audit and persistence |
| `pkg/version/` | Version management |
| `pkg/migration/` | Data migration validation |
| `pkg/modding/` | Mod system: loader, manager, sandboxed execution |
| `pkg/narrative/` | Branching narrative types |
| `pkg/ux/` | UX validation and user journeys |
| `pkg/balance/` | Game balance: combat, economic |
| `pkg/class/` | Class system: advanced multiclassing |
| `pkg/companion/` | Companion: learning system |
| `pkg/social/` | Social system persistence |
| `pkg/hostplay/` | Host-and-play (local server + client) |
| `pkg/mobile/` | Mobile platform: controls, touch input, dual joystick, keyboard |
| `pkg/audit/` | Code audit utilities |
| `pkg/visualtest/` | Visual testing: benchmarks, snapshots, genre tests, regression |

### Examples (`examples/`)

Demo programs for testing individual systems:
- `animation_timing_demo.go` - Animation timing showcase
- `bloom_demo.go` - Bloom effect demonstration
- `genre_ui_palettes_demo.go` - Genre-based UI palette testing
- `momentum_scrolling_demo.go` - Touch scrolling demo
- `mouse_delta_demo.go` - Mouse input testing
- `soft_shadow_demo.go` - Shadow system demo
- `sprite_antialiasing_demo.go` - Sprite antialiasing
- `weather_cli_demo.go` - Weather system CLI
- `virtual_controls_wasm_demo/` - WASM virtual controls

### Scripts (`scripts/`)

Build, test, and deployment automation:
- `build-*.sh` - Platform-specific builds (Linux, macOS, Windows, Android, iOS)
- `package-*.sh` - Packaging (deb, rpm, Docker, Windows, release)
- `test-*.sh` - Platform testing scripts
- `validate-*.sh` - Validation scripts
- `benchmark-*.sh` - Performance benchmarking
- `profile_cpu.sh` - CPU profiling
- `sign-binaries.sh` - Binary signing

### Documentation (`docs/`)

60+ documentation files covering:
- Architecture and technical specs
- Platform-specific builds (Android, iOS, WASM)
- System documentation (lighting, shadows, magic, rotation, post-processing)
- Deployment guides (GitHub Pages, production, Tor)
- Performance optimization guides
- Runbooks for operations

## Quality Standards

- **Test Coverage**: Minimum 40% per package (30% for packages depending on X11/Wayland/Ebiten). Run `go test -cover ./pkg/...` to verify.
- **Code Quality**: All code must pass `go fmt`, `go vet`, and ideally `golangci-lint run`.
- **Documentation**: All exported functions, types, and packages must have godoc comments. Each package should have a `doc.go` file explaining its purpose.
- **Commit Messages**: Use conventional format: `feat:`, `fix:`, `docs:`, `test:`, `perf:`, `refactor:`.
- **Performance Validation**: Run benchmarks for performance-critical changes. Maintain 60+ FPS with 2000 entities.

## Networking Best Practices

When declaring network variables, always use interface types:
- Never use `net.UDPAddr`, `net.IPAddr`, or `net.TCPAddr`. Use `net.Addr` only instead.
- Never use `net.UDPConn`, use `net.PacketConn` instead.
- Never use `net.TCPConn`, use `net.Conn` instead.
- Never use `net.UDPListener` or `net.TCPListener`, use `net.Listener` instead.
- Never use a type switch or type assertion to convert from an interface type to a concrete type. Use the interface methods instead.

This approach enhances testability and flexibility when working with different network implementations or mocks.

## Generator Pattern

All procedural generators implement the `Generator` interface:
```go
type Generator interface {
    Generate(seed int64, params GenerationParams) (interface{}, error)
    Validate(result interface{}) error
}
```

Use `GenerationParams` with `Difficulty` (0.0-1.0), `Depth` (game progression), `GenreID` (theme), and `Custom` map for additional parameters. Validate parameters before generation using `ValidateParams()`.

## System Pattern

Systems follow this structure:
```go
type MySystem struct {
    world *World
    // Optional dependencies
}

func NewMySystem(params) *MySystem {
    log.WithFields(log.Fields{"system_name": "mysystem"}).Debug("Creating system")
    return &MySystem{...}
}

func (s *MySystem) Update(entities []*Entity, deltaTime float64) {
    for _, entity := range entities {
        if !entity.HasComponent("required_component") {
            continue
        }
        // Process entity
    }
}
```

## Component Pattern

Components are pure data with Type() method:
```go
type MyComponent struct {
    Field1 float64
    Field2 string
}

func (c *MyComponent) Type() string { return "mycomponent" }

// Optional: Serialize/Deserialize for persistence
func (c *MyComponent) Serialize() ([]byte, error) { ... }
func (c *MyComponent) Deserialize(data []byte) error { ... }
```

## Package Integration Status

The project has 90+ active packages organized into 30 domain areas. All packages are fully integrated as of the latest version.

### Package Statistics by Domain

| Domain | Packages | LOC (approx) | Description |
|--------|----------|--------------|-------------|
| `engine/` | 1 (400+ files) | 240K+ | Core ECS, all game systems, components |
| `procgen/` | 25+ subdirs | 50K+ | Procedural content generators |
| `rendering/` | 15+ subdirs | 30K+ | Graphics pipeline and sprite generation |
| `network/` | 10+ subdirs | 25K+ | Multiplayer networking |
| `world/` | 5+ subdirs | 15K+ | World state and persistence |
| `integration/` | 10+ subdirs | 10K+ | Cross-system integrations |
| `audio/` | 4 subdirs | 8K+ | Audio synthesis |
| Supporting | 20+ packages | 15K+ | Utilities, validation, security |

### Key Active Systems

**Engine Systems (100+)**: All ECS systems fully operational including combat, AI, rendering, physics, UI, progression, social, narrative, and multiplayer systems.

**Procedural Generators (25+)**: Terrain (BSP, cellular, L-system, Voronoi, city), entities, items, quests, magic, skills, dialog, narrative, factions, companions, vehicles, legendary items, minigames, puzzles.

**Rendering Pipeline**: Sprites with equipment overlays, animations with articulation, tiles with transitions, lighting with bloom/AO, post-processing effects, particles with physics, UI generation.

**Network Layer**: Client-server architecture, federation with discovery/sync, WebRTC support, chat/trade systems, resilience patterns.

**World Systems**: Chunk-based persistence, housing with guildhalls, economy with marketplace, territory control with siege mechanics, raid generation.

### Test/Infrastructure Packages

Used by CI/CD for quality validation:
- `pkg/audit/features/` - Feature audit tests
- `pkg/procgen/audit/` - Procgen audit tests
- `pkg/visualtest/` - Visual regression testing
- `pkg/visualtest/parity/` - Cross-platform parity tests
