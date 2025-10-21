# Venture - Procedural Action RPG

A fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, gameplay content—is generated at runtime with no external asset files.

## Overview

Venture is a top-down action-RPG that combines the deep procedural generation of modern roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda and Chrono Trigger.

**Key Features:**
- 🎮 Real-time action-RPG combat and exploration
- 🎲 100% procedurally generated content (maps, items, monsters, abilities, quests)
- 🎨 Runtime-generated graphics using procedural techniques
- 🎵 Procedural audio synthesis for music and sound effects
- 🌐 Multiplayer co-op supporting high-latency connections (200-500ms)
- 🎭 Multiple genres (fantasy, sci-fi, post-apocalyptic, horror, cyberpunk)
- 📦 Single binary distribution - no external asset files required

## Project Status

**Current Phase:** Phase 2 - Procedural Generation Core (In Progress) 🚧

Phase 1 (Architecture & Foundation) is complete. We are now implementing Phase 2 with terrain generation as the first deliverable.

### Phase 2 Progress

- [x] **Terrain/Dungeon Generation**
  - [x] BSP (Binary Space Partitioning) algorithm
  - [x] Cellular Automata algorithm
  - [x] Comprehensive test suite (91.5% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [x] **Entity Generator (monsters, NPCs)**
  - [x] Entity type system (Monster, Boss, Minion, NPC)
  - [x] Stats and rarity system
  - [x] Fantasy and Sci-Fi templates
  - [x] Deterministic generation with level scaling
  - [x] Comprehensive test suite (87.8% coverage)
  - [x] CLI tool for visualization
  - [x] Complete documentation
- [ ] Item generation system
- [ ] Magic/spell generation
- [ ] Skill tree generation
- [ ] Genre definition system

See the [Phase 2 Terrain Implementation](docs/PHASE2_TERRAIN_IMPLEMENTATION.md) for complete details.

### Development Roadmap

- [x] **Phase 1: Architecture & Foundation** (Weeks 1-2) ✅
  - [x] Project structure and Go module setup
  - [x] Core ECS (Entity-Component-System) framework
  - [x] Base interfaces for all major systems
  - [x] Basic Ebiten game loop
  - [x] Architecture Decision Records

- [ ] **Phase 2: Procedural Generation Core** (Weeks 3-5) 🚧
  - [x] Terrain/dungeon generation (BSP, cellular automata)
  - [x] Entity generator (monsters, NPCs)
  - [ ] Item generation system
  - [ ] Magic/spell generation
  - [ ] Skill tree generation
  - [ ] Genre definition system

- [ ] **Phase 3: Visual Rendering System** (Weeks 6-7)
  - [ ] Procedural shape generation
  - [ ] Runtime sprite generation
  - [ ] Tile rendering system
  - [ ] Particle effects
  - [ ] UI rendering
  - [ ] Genre-based color palettes

- [ ] **Phase 4: Audio Synthesis** (Weeks 8-9)
  - [ ] Waveform generation
  - [ ] Procedural music composition
  - [ ] Sound effect generation
  - [ ] Audio mixing

- [ ] **Phase 5: Core Gameplay Systems** (Weeks 10-13)
  - [ ] Movement and collision detection
  - [ ] Combat system (melee, ranged, magic)
  - [ ] Inventory and equipment
  - [ ] Character progression
  - [ ] Monster AI
  - [ ] Quest generation

- [ ] **Phase 6: Networking & Multiplayer** (Weeks 14-16)
  - [ ] Network protocol
  - [ ] Authoritative game server
  - [ ] Client-side prediction
  - [ ] State synchronization
  - [ ] Lag compensation

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

- Go 1.21 or later
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

### Running

```bash
# Start the client (single-player or connecting to server)
./venture-client -width 1024 -height 768 -seed 12345

# Start a dedicated server
./venture-server -port 8080 -max-players 4
```

## Project Structure

```
venture/
├── cmd/
│   ├── client/          # Client application
│   └── server/          # Server application
├── pkg/
│   ├── engine/          # Core game loop and ECS framework
│   ├── procgen/         # Procedural generation systems
│   │   ├── terrain/     # Map/dungeon generation
│   │   ├── entity/      # Monster/NPC generation
│   │   ├── items/       # Weapon/armor/item generation
│   │   ├── magic/       # Spell/ability generation
│   │   ├── skills/      # Skill tree generation
│   │   └── genre/       # Genre definition system
│   ├── rendering/       # Visual generation
│   │   ├── primitives/  # Shape generation
│   │   ├── sprites/     # Sprite generation
│   │   ├── tiles/       # Tile rendering
│   │   ├── particles/   # Particle effects
│   │   ├── ui/          # UI rendering
│   │   └── palette/     # Color scheme generation
│   ├── audio/           # Sound synthesis
│   │   ├── synthesis/   # Waveform generation
│   │   ├── music/       # Music composition
│   │   ├── sfx/         # Sound effects
│   │   └── mixer/       # Audio mixing
│   ├── network/         # Multiplayer systems
│   │   ├── protocol/    # Network protocol
│   │   ├── server/      # Game server
│   │   ├── client/      # Client networking
│   │   ├── sync/        # State synchronization
│   │   └── lag/         # Lag compensation
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
