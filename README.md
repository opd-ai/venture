# Venture - Procedural Action RPG

A fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, gameplay content—is generated at runtime with no external asset files.

## Overview

Venture is a top-down action-RPG that combines the deep procedural generation of modern roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda and Chrono Trigger.

**Key Features:**
- 🎮 Real-time action-RPG combat and exploration
- 🎲 100% procedurally generated content (maps, items, monsters, abilities, quests)
- 🎨 Runtime-generated graphics using procedural techniques
- 🎵 Procedural audio synthesis for music and sound effects
- 🌐 Multiplayer co-op supporting high-latency connections (200-5000ms, onion services)
- 🎭 Multiple genres (fantasy, sci-fi, post-apocalyptic, horror, cyberpunk)
- 📦 Single binary distribution - no external asset files required

## Project Status

**Current Phase:** Phase 6 - Networking & Multiplayer 🚧 IN PROGRESS

Phases 1-5 complete (Architecture, Procedural Generation, Visual Rendering, Audio Synthesis, Core Gameplay). Phase 6.1-6.2 complete with binary serialization, client/server communication, client-side prediction, and state synchronization. Network package at 63.1% test coverage.

### Phase 2 Progress

- [x] **Terrain/Dungeon Generation**
  - [x] BSP (Binary Space Partitioning) algorithm
  - [x] Cellular Automata algorithm
  - [x] Comprehensive test suite (96.4% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Entity Generator (monsters, NPCs)**
  - [x] Entity type system (Monster, Boss, Minion, NPC)
  - [x] Stats and rarity system
  - [x] Fantasy and Sci-Fi templates
  - [x] Deterministic generation with level scaling
  - [x] Comprehensive test suite (95.9% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Item Generation System**
  - [x] Item type system (Weapon, Armor, Consumable, Accessory)
  - [x] Rarity and stat generation with depth scaling
  - [x] Fantasy and Sci-Fi item templates
  - [x] Deterministic generation with seed support
  - [x] Comprehensive test suite (93.8% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Magic/Spell Generation System**
  - [x] Spell type system (Offensive, Defensive, Healing, Buff, Debuff, Utility, Summon)
  - [x] Element system (Fire, Ice, Lightning, Earth, Wind, Light, Dark, Arcane)
  - [x] Target patterns (Self, Single, Area, Cone, Line, All Allies, All Enemies)
  - [x] Fantasy and Sci-Fi spell templates
  - [x] Deterministic generation with power scaling
  - [x] Comprehensive test suite (91.9% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Skill Tree Generation System**
  - [x] Skill type system (Passive, Active, Ultimate, Synergy)
  - [x] Tier-based progression (Basic, Intermediate, Advanced, Master)
  - [x] Prerequisite and dependency system
  - [x] Multiple skill trees per genre (Warrior, Mage, Rogue, Soldier, Engineer, Biotic)
  - [x] Fantasy and Sci-Fi skill templates
  - [x] Deterministic generation with level scaling
  - [x] Comprehensive test suite (90.6% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Genre Definition System**
  - [x] Genre type with ID, name, description, themes, colors, and prefixes
  - [x] Registry for centralized genre management
  - [x] Five predefined genres (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic)
  - [x] Validation and type safety
  - [x] Comprehensive test suite (100.0% coverage)
  - [x] CLI tool for genre exploration
  - [x] Complete documentation

See the [Phase 2 Terrain Implementation](docs/PHASE2_TERRAIN_IMPLEMENTATION.md) for complete details.

### Development Roadmap

- [x] **Phase 1: Architecture & Foundation** (Weeks 1-2) ✅
  - [x] Project structure and Go module setup
  - [x] Core ECS (Entity-Component-System) framework
  - [x] Base interfaces for all major systems
  - [x] Basic Ebiten game loop
  - [x] Architecture Decision Records

- [ ] **Phase 2: Procedural Generation Core** (Weeks 3-5) ✅
  - [x] Terrain/dungeon generation (BSP, cellular automata)
  - [x] Entity generator (monsters, NPCs)
  - [x] Item generation system
  - [x] Magic/spell generation
  - [x] Skill tree generation
  - [x] Genre definition system

- [ ] **Phase 3: Visual Rendering System** (Weeks 6-7) ✅
  - [x] Genre-based color palettes (98.4% coverage)
  - [x] Procedural shape generation (100% coverage)
  - [x] Runtime sprite generation (100% coverage)
  - [x] Tile rendering system (92.6% coverage)
  - [x] Particle effects (98.0% coverage)
  - [x] UI rendering (94.8% coverage)

- [ ] **Phase 4: Audio Synthesis** (Weeks 8-9) ✅
  - [x] Waveform generation (5 types, 94.2% coverage)
  - [x] Procedural music composition (100% coverage)
  - [x] Sound effect generation (9 types, 99.1% coverage)
  - [x] Audio mixing and processing
  - [x] Genre-aware audio themes
  - [x] CLI testing tool (audiotest)

- [x] **Phase 5: Core Gameplay Systems** (Weeks 10-13) ✅
  - [x] Movement and collision detection (95.4% coverage)
  - [x] Combat system (melee, ranged, magic) (90.1% coverage)
  - [x] Inventory and equipment (85.1% coverage)
  - [x] Character progression (100% coverage)
  - [x] Monster AI (100% coverage)
  - [x] Quest generation (96.6% coverage)

- [ ] **Phase 6: Networking & Multiplayer** (Weeks 14-16) 🚧 IN PROGRESS
  - [x] Binary protocol serialization (100% coverage)
  - [x] Network client layer (45% coverage*)
  - [x] Authoritative game server (35% coverage*)
  - [x] Client-side prediction (100% coverage)
  - [x] State synchronization (100% coverage)
  - [ ] Lag compensation

*Note: Client/server require integration tests for full coverage (I/O operations)

- [ ] **Phase 7: Genre System** (Weeks 17-18)
  - [ ] Genre templates (5+ genres)
  - [ ] Cross-genre modifiers
  - [ ] Theme-appropriate content generation

- [ ] **Phase 8: Polish & Optimization** (Weeks 19-20)
  - [ ] Performance optimization
  - [ ] Game balance
  - [ ] Tutorial system
  - [ ] Save/load functionality
  - [ ] Complete documentation

## Quick Start

### Prerequisites

- Go 1.24.7 or later
- Platform-specific dependencies for Ebiten:
  - **Linux:** `apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`
  - **macOS:** Xcode command line tools
  - **Windows:** No additional dependencies

### Building

```bash
# Clone the repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Build the client (requires X11 libraries on Linux)
go build -o venture-client ./cmd/client

# Build the server
go build -o venture-server ./cmd/server

# Build the terrain test tool (no graphics dependencies)
go build -o terraintest ./cmd/terraintest

# Build the entity test tool (no graphics dependencies)
go build -o entitytest ./cmd/entitytest

# Build the item test tool (no graphics dependencies)
go build -o itemtest ./cmd/itemtest

# Build the magic test tool (no graphics dependencies)
go build -o magictest ./cmd/magictest

# Build the skill test tool (no graphics dependencies)
go build -o skilltest ./cmd/skilltest

# Build the genre test tool (no graphics dependencies)
go build -o genretest ./cmd/genretest

# Build the rendering test tool (no graphics dependencies)
go build -o rendertest ./cmd/rendertest

# Build the audio test tool (no graphics dependencies)
go build -o audiotest ./cmd/audiotest

# Build the movement test tool
go build -o movementtest ./cmd/movementtest

# Build the inventory test tool (no graphics dependencies)
go build -o inventorytest ./cmd/inventorytest

# Build the tile test tool (no graphics dependencies)
go build -o tiletest ./cmd/tiletest
```

### Testing Terrain Generation

Try out the procedural terrain generation:

```bash
# Generate a BSP dungeon
./terraintest -algorithm bsp -width 80 -height 50 -seed 12345

# Generate cellular automata caves
./terraintest -algorithm cellular -width 80 -height 50 -seed 54321

# Save to file
./terraintest -algorithm bsp -output dungeon.txt
```

See [pkg/procgen/terrain/README.md](pkg/procgen/terrain/README.md) for more details on terrain generation.

### Testing Entity Generation

Try out the procedural entity generation:

```bash
# Generate fantasy entities
./entitytest -genre fantasy -count 20 -depth 5 -seed 12345

# Generate sci-fi entities with verbose details
./entitytest -genre scifi -count 15 -depth 10 -verbose

# Save to file
./entitytest -genre fantasy -count 100 -output entities.txt
```

See [pkg/procgen/entity/README.md](pkg/procgen/entity/README.md) for more details on entity generation.

### Testing Item Generation

Try out the procedural item generation:

```bash
# Generate fantasy weapons
./itemtest -genre fantasy -count 20 -type weapon -seed 12345

# Generate sci-fi armor with verbose details
./itemtest -genre scifi -count 15 -type armor -depth 10 -verbose

# Generate mixed items at high depth
./itemtest -genre fantasy -count 50 -depth 20

# Save to file
./itemtest -genre fantasy -count 100 -output items.txt
```

See [pkg/procgen/item/README.md](pkg/procgen/item/README.md) for more details on item generation.

### Testing Magic Generation

Try out the procedural spell generation:

```bash
# Generate fantasy spells
./magictest -genre fantasy -count 20 -depth 5 -seed 12345

# Generate sci-fi spells with verbose details
./magictest -genre scifi -count 15 -depth 10 -verbose

# Filter by spell type
./magictest -type offensive -count 30 -depth 15

# Save to file
./magictest -genre fantasy -count 100 -output spells.txt
```

See [pkg/procgen/magic/README.md](pkg/procgen/magic/README.md) for more details on magic generation.

### Testing Skill Tree Generation

Try out the procedural skill tree generation:

```bash
# Generate fantasy skill trees
./skilltest -genre fantasy -count 3 -depth 5 -seed 12345

# Generate sci-fi skill trees with verbose details
./skilltest -genre scifi -count 3 -depth 10 -verbose

# Generate trees at high depth
./skilltest -genre fantasy -count 3 -depth 20

# Save to file
./skilltest -genre fantasy -count 5 -output skills.txt
```

See [pkg/procgen/skills/README.md](pkg/procgen/skills/README.md) for more details on skill tree generation.

### Testing Genre System

Explore the genre definition system:

```bash
# List all available genres
./genretest -list

# Show details for a specific genre
./genretest -genre fantasy

# Show all genres with detailed information
./genretest -all

# Validate a genre ID
./genretest -validate horror
```

See [pkg/procgen/genre/README.md](pkg/procgen/genre/README.md) for more details on the genre system.

### Testing Rendering System

Try out the procedural color palette generation:

```bash
# Generate fantasy-themed color palette
./rendertest -genre fantasy -seed 12345

# Generate sci-fi palette with verbose details
./rendertest -genre scifi -seed 54321 -verbose

# Generate and save cyberpunk palette
./rendertest -genre cyberpunk -output palette.txt
```

See [pkg/rendering/palette/README.md](pkg/rendering/palette/README.md) for more details on the rendering system.

### Testing Audio System

Try out the procedural audio generation:

```bash
# Test waveform synthesis
./audiotest -type oscillator -waveform sine -frequency 440 -duration 1.0 -verbose

# Generate sound effects
./audiotest -type sfx -effect explosion -verbose

# Generate music tracks
./audiotest -type music -genre fantasy -context combat -duration 5.0 -verbose

# Try different genres
./audiotest -type music -genre horror -context ambient -duration 10.0
```

See [pkg/audio/README.md](pkg/audio/README.md) for more details on the audio synthesis system.

### Testing Movement and Collision

Try out the movement and collision systems:

```bash
# Run the interactive example (requires -tags test)
go run -tags test ./examples/movement_collision_demo.go

# Or build and run the CLI tool (requires display currently)
./movementtest -count 50 -duration 3.0 -verbose

# Options:
#   -count N        Number of entities (default: 10)
#   -duration N     Simulation duration in seconds (default: 5.0)
#   -verbose        Show detailed output
#   -seed N         Random seed
```

See [pkg/engine/MOVEMENT_COLLISION.md](pkg/engine/MOVEMENT_COLLISION.md) for more details on movement and collision systems.

### Testing Combat System

Try out the combat system with damage calculation, status effects, and team mechanics:

```bash
# Run the interactive combat demo (requires -tags test)
go run -tags test ./examples/combat_demo.go

# Demonstrates:
#   - Basic melee combat with stats
#   - Magic combat with resistances
#   - Status effects (poison over time)
#   - Critical hit mechanics
#   - Team-based enemy detection
```

See [pkg/engine/COMBAT_SYSTEM.md](pkg/engine/COMBAT_SYSTEM.md) for more details on the combat system.

### Testing Networking System

Try out the networking and multiplayer systems:

```bash
# Run the networking demo (requires -tags test)
go run -tags test ./examples/network_demo.go

# Demonstrates:
#   - Binary protocol serialization
#   - Client/server configuration
#   - State broadcasting
#   - Performance characteristics

# Run the multiplayer integration demo (requires -tags test)
go run -tags test ./examples/multiplayer_demo.go

# Demonstrates:
#   - Complete client-server setup
#   - Component serialization with ECS
#   - Simulated multiplayer game loop
#   - Multiple concurrent players

# Run the prediction and synchronization demo (requires -tags test)
go run -tags test ./examples/prediction_demo.go

# Demonstrates:
#   - Client-side prediction for responsive controls
#   - Server reconciliation with error correction
#   - Entity interpolation for smooth movement
#   - State synchronization techniques
```

See [pkg/network/README.md](pkg/network/README.md) for more details on the networking system.

### Running

```bash
# Start the client (single-player or connecting to server)
./venture-client -width 1024 -height 768 -seed 12345

# Start a dedicated server
./venture-server -port 8080 -max-players 4
```

## Documentation

### Root Documentation
- **README.md** - This file, project overview and quick start guide

### Core Documentation (docs/)
- **ARCHITECTURE.md** - Architecture Decision Records (ADRs)
- **TECHNICAL_SPEC.md** - Complete technical specification
- **ROADMAP.md** - Detailed 8-phase development roadmap
- **DEVELOPMENT.md** - Development guide and best practices
- **CLEANUP_REPORT_2025-10-22.md** - Repository cleanup summary (Oct 2025)

### Phase Implementation Reports (docs/)
- **PHASE1_SUMMARY.md** - Phase 1: Architecture & Foundation
- **PHASE2_TERRAIN_IMPLEMENTATION.md** - Terrain/dungeon generation (BSP, cellular automata)
- **PHASE2_ENTITY_IMPLEMENTATION.md** - Monster and NPC generation
- **PHASE2_ITEM_IMPLEMENTATION.md** - Weapons, armor, and consumables
- **PHASE2_MAGIC_IMPLEMENTATION.md** - Spell and ability generation
- **PHASE2_SKILLS_IMPLEMENTATION.md** - Skill tree generation
- **PHASE2_GENRE_IMPLEMENTATION.md** - Genre definition system
- **PHASE3_RENDERING_IMPLEMENTATION.md** - Visual rendering systems
- **PHASE4_AUDIO_IMPLEMENTATION.md** - Audio synthesis implementation
- **PHASE5_COMBAT_IMPLEMENTATION.md** - Combat system implementation
- **PHASE5_MOVEMENT_COLLISION_REPORT.md** - Movement and collision systems
- **PHASE5_PROGRESSION_AI_REPORT.md** - Character progression and AI systems
- **PHASE5_QUEST_IMPLEMENTATION.md** - Quest generation system
- **PHASE6_NETWORKING_IMPLEMENTATION.md** - Networking and multiplayer foundation
- **PHASE6_2_PREDICTION_SYNC_IMPLEMENTATION.md** - Client-side prediction and state synchronization

### Package-Specific Documentation
Each package contains detailed technical documentation:

**Procedural Generation (pkg/procgen/):**
- [terrain/README.md](pkg/procgen/terrain/README.md) - Terrain/dungeon generation
- [entity/README.md](pkg/procgen/entity/README.md) - Monster and NPC generation
- [item/README.md](pkg/procgen/item/README.md) - Item generation
- [magic/README.md](pkg/procgen/magic/README.md) - Spell generation
- [skills/README.md](pkg/procgen/skills/README.md) - Skill tree generation
- [genre/README.md](pkg/procgen/genre/README.md) - Genre system
- [quest/README.md](pkg/procgen/quest/README.md) - Quest generation

**Rendering (pkg/rendering/):**
- [palette/README.md](pkg/rendering/palette/README.md) - Color palette generation
- [tiles/README.md](pkg/rendering/tiles/README.md) - Tile rendering
- [particles/README.md](pkg/rendering/particles/README.md) - Particle effects

**Audio (pkg/audio/):**
- [README.md](pkg/audio/README.md) - Audio synthesis system

**Networking (pkg/network/):**
- [README.md](pkg/network/README.md) - Multiplayer networking system

**Game Engine (pkg/engine/):**
- [MOVEMENT_COLLISION.md](pkg/engine/MOVEMENT_COLLISION.md) - Movement and collision system
- [COMBAT_SYSTEM.md](pkg/engine/COMBAT_SYSTEM.md) - Combat system
- [INVENTORY_EQUIPMENT.md](pkg/engine/INVENTORY_EQUIPMENT.md) - Inventory and equipment system
- [PROGRESSION_SYSTEM.md](pkg/engine/PROGRESSION_SYSTEM.md) - Character progression system
- [AI_SYSTEM.md](pkg/engine/AI_SYSTEM.md) - AI behavior system

## Project Structure

```
venture/
├── cmd/
│   ├── client/          # Client application
│   ├── server/          # Server application
│   ├── movementtest/    # Movement/collision demo tool
│   └── ... (other test tools)
├── pkg/
│   ├── engine/          # Core game loop and ECS framework
│   │   ├── ecs.go       # Entity-Component-System
│   │   ├── components.go # Movement/collision components
│   │   ├── movement.go  # Movement system
│   │   ├── collision.go # Collision detection system
│   │   └── game.go      # Ebiten integration
│   ├── procgen/         # Procedural generation systems
│   │   ├── terrain/     # Map/dungeon generation
│   │   ├── entity/      # Monster/NPC generation
│   │   ├── item/        # Weapon/armor/item generation
│   │   ├── magic/       # Spell/ability generation
│   │   ├── skills/      # Skill tree generation
│   │   └── genre/       # Genre definition system
│   ├── rendering/       # Visual generation
│   │   ├── shapes/      # Shape generation
│   │   ├── sprites/     # Sprite generation
│   │   ├── tiles/       # Tile rendering
│   │   ├── particles/   # Particle effects
│   │   ├── ui/          # UI rendering
│   │   └── palette/     # Color scheme generation
│   ├── audio/           # Sound synthesis
│   │   ├── synthesis/   # Waveform generation
│   │   ├── music/       # Music composition
│   │   └── sfx/         # Sound effects
│   ├── network/         # Multiplayer systems
│   ├── combat/          # Combat mechanics
│   └── world/           # World state management
├── docs/
│   └── ARCHITECTURE.md  # Architectural decisions
├── go.mod
└── README.md
```

## Architecture

The game uses an Entity-Component-System (ECS) architecture for maximum flexibility and performance. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architectural decisions.

### Core Concepts

**Entities:** Game objects represented by unique IDs with attached components
**Components:** Pure data structures (Position, Health, Sprite, etc.)
**Systems:** Behavior logic that operates on entities with specific components

This architecture allows for easy composition of complex behaviors and efficient data processing.

## Performance Targets

- **FPS:** 60 minimum on modest hardware (Intel i5/Ryzen 5, 8GB RAM, integrated graphics)
- **Memory:** <500MB client, <1GB server (4 players)
- **Generation:** <2 seconds for new world areas
- **Network:** <100KB/s per player at 20 updates/second

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Development

### Code Quality

- All packages must include `doc.go` with package documentation
- Public interfaces defined in dedicated files
- Comprehensive unit tests (target: 80%+ coverage)
- Follow Go best practices and conventions

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Contributing

This is an active development project following a structured roadmap. Contributions are welcome! Please:

1. Review the current phase in the roadmap
2. Check existing issues and pull requests
3. Follow the code quality standards
4. Include tests for new functionality
5. Update documentation as needed

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) - A dead simple 2D game library for Go
- Inspired by roguelikes like Dungeon Crawl Stone Soup and Cataclysm DDA
- Gameplay inspired by classic action-RPGs like The Legend of Zelda and Chrono Trigger
