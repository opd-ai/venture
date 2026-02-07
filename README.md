# Venture

A fully procedural multiplayer action-RPG built with Go and Ebiten where all graphics, audio, and gameplay content are generated at runtime from a single binary with no external asset files.

## Description

Venture is a top-down action-RPG that uses deterministic, seed-based procedural generation to create all game content at runtime. The game uses an Entity-Component-System (ECS) architecture where entities are unique identifiers with component collections, components are pure data structures, and systems contain all game logic. Terrain is generated using BSP dungeons, cellular automata caves, L-system forests, and Voronoi biomes. Items, quests, NPCs, spells, and dialog are all procedurally generated based on a genre system supporting fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic themes.

The multiplayer networking layer supports high-latency connections (200–5000ms) suitable for Tor/onion service routing, with client-side prediction, lag compensation, and snapshot synchronization. A federation system enables cross-server travel, shared marketplaces, and multi-server guilds. On desktop/native builds, the client automatically starts a localhost server for solo play, requiring no manual server setup; on WebAssembly builds, embedded server/host-and-play is disabled and the client must connect to an existing server.

The rendering pipeline generates sprites with equipment overlays, tiles with transitions, dynamic lighting with bloom and ambient occlusion, particle effects, and post-processing—all at runtime. Audio synthesis generates music and sound effects procedurally. The result is a single distributable binary per platform with no external asset files.

## Tech Stack

- **Language:** [Go](https://go.dev/) 1.24.5+
- **Game Framework:** [Ebiten](https://ebiten.org/) v2.9.3 — 2D game engine with cross-platform support including WebAssembly
- **Structured Logging:** [Logrus](https://github.com/sirupsen/logrus) v1.9.3
- **UUID Generation:** [google/uuid](https://github.com/google/uuid) v1.6.0
- **Image Processing:** [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) v0.32.0
- **Native Dialogs:** [ncruces/zenity](https://github.com/ncruces/zenity) v0.10.14
- **Testing:** Go standard `testing` package with table-driven tests and benchmarks
- **Build System:** GNU Make, Go build toolchain, GitHub Actions CI/CD

## Project Structure

```
venture/
├── cmd/                    # Application entry points
│   ├── client/             # Desktop game client (main.go)
│   ├── server/             # Dedicated multiplayer server (main.go)
│   └── mobile/             # Mobile entry point for iOS/Android (mobile.go)
├── pkg/                    # Core library packages
│   ├── engine/             # ECS core, 100+ game systems, components, spatial partitioning
│   │   ├── physics/        # Vehicle physics, fluid simulation, environmental destruction
│   │   ├── prestige/       # New Game+ and prestige progression
│   │   └── qol/            # Quality-of-life systems (auto-loot, craft queue, etc.)
│   ├── procgen/            # Procedural content generators (25+ subdirectories)
│   │   ├── terrain/        # BSP dungeons, cellular automata, L-systems, Voronoi biomes
│   │   ├── entity/         # NPC and creature generation
│   │   ├── item/           # Item generation with rarity tiers
│   │   ├── quest/          # Quest generation with objectives and rewards
│   │   ├── magic/          # Spell and magic system generation
│   │   ├── dialog/         # Dialog generation with Markov chains
│   │   ├── narrative/      # Story beat and narrative arc generation
│   │   └── ...             # genre, skills, building, furniture, vehicle, faction, etc.
│   ├── rendering/          # Runtime graphics generation pipeline
│   │   ├── sprites/        # Sprite generation, anatomy templates, equipment overlays
│   │   ├── animation/      # Animation system with articulation and directional variants
│   │   ├── tiles/          # Tile generation with transitions and parallax
│   │   ├── lighting/       # Bloom, ambient occlusion, dynamic lights
│   │   ├── postprocess/    # Chromatic aberration, color grading, depth blur, vignette
│   │   ├── particles/      # Particle physics, weather effects, LOD
│   │   └── ui/             # UI generation, chat, notifications, tutorials
│   ├── audio/              # Procedural audio synthesis
│   │   ├── music/          # Adaptive soundtrack, motifs, theory-based composition
│   │   ├── sfx/            # Sound effect generation and processing
│   │   └── synthesis/      # Oscillators, envelopes, synthesis engine
│   ├── network/            # Multiplayer networking
│   │   ├── federation/     # Cross-server discovery, auth, sync, WebRTC, portals
│   │   ├── chat/           # Chat system with channels
│   │   ├── trade/          # Player-to-player trade system
│   │   └── resilience/     # Network resilience metrics and simulation
│   ├── world/              # Persistent world state
│   │   ├── housing/        # Player housing, blueprints, guildhalls
│   │   ├── economy/        # Marketplace, pricing engine, guild bank
│   │   ├── territory/      # Territory control and siege mechanics
│   │   └── raids/          # Raid generation, instances, lockouts
│   ├── integration/        # Cross-system feature integrations (10 subdirectories)
│   ├── combat/             # Damage calculation and combat resolution
│   ├── config/             # Configuration types and validation
│   ├── modding/            # JSON-based mod loader and sandboxed execution
│   ├── saveload/           # Save/load with migration and WASM storage support
│   ├── validation/         # Input validation, chat filtering, rate limiting
│   ├── security/           # Security audit and persistence
│   └── version/            # Semantic version management (1.0.0)
├── docs/                   # 60+ documentation files
├── examples/               # Demo programs (bloom, shadows, weather, sprites, etc.)
├── scripts/                # Build, test, packaging, and deployment scripts
├── mods/                   # Example mod configurations (JSON)
├── web/                    # WebAssembly deployment assets
├── build/                  # Platform-specific build configs (Android, WASM)
├── Formula/                # Homebrew formula (venture.rb)
├── Makefile                # Build automation (build, test, lint, release, Docker)
└── go.mod                  # Go module definition and dependencies
```

## Installation

### Option A: Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/opd-ai/venture/releases/latest):

- Linux: `venture-linux-amd64.tar.gz` or `venture-linux-arm64.tar.gz`
- macOS: `venture-darwin-amd64.tar.gz` or `venture-darwin-arm64.tar.gz`
- Windows: `venture-windows-amd64.zip`

```bash
tar -xzf venture-linux-amd64.tar.gz
./venture-client-linux-amd64
```

### Option B: Package Managers

```bash
# macOS (Homebrew)
brew tap opd-ai/tap
brew install venture

# Debian/Ubuntu
curl -LO https://github.com/opd-ai/venture/releases/latest/download/venture_1.0.0_amd64.deb
sudo dpkg -i venture_1.0.0_amd64.deb

# RHEL/Fedora
curl -LO https://github.com/opd-ai/venture/releases/latest/download/venture-1.0.0-1.x86_64.rpm
sudo rpm -i venture-1.0.0-1.x86_64.rpm

# Docker (server only)
docker run -d -p 8080:8080 ghcr.io/opd-ai/venture-server:1.0.0
```

### Option C: Build from Source

1. Install [Go 1.24.5+](https://go.dev/dl/)
2. Install platform-specific dependencies:
   - **Linux:** `sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config xvfb`
   - **macOS:** Xcode command-line tools (`xcode-select --install`)
   - **Windows:** No additional dependencies
3. Clone and build:

```bash
git clone https://github.com/opd-ai/venture.git
cd venture
make build
```

This produces `build/venture-client` and `build/venture-server`. Alternatively, build directly with Go:

```bash
go build -o venture-client ./cmd/client
go build -o venture-server ./cmd/server
```

## Usage

### Running the Client

```bash
# Start the game (auto-starts a localhost server for solo play)
# After `make build`, binaries are in build/
./build/venture-client

# Or if built directly with `go build -o venture-client ./cmd/client`:
./venture-client

# Custom settings
./venture-client -width 2560 -height 1440 -fullscreen -seed 12345 -genre fantasy
```

### Running a Dedicated Server

```bash
# Start dedicated server
./venture-server -port 8080 -max-players 8 -seed 12345 -genre fantasy

# Connect a client to the server
./venture-client --multiplayer --server <address>:8080
```

### LAN Co-op

```bash
# Host: start a local server and allow LAN connections
./venture-client --host-and-play --host-lan

# Other players: join the host
./venture-client --multiplayer --server <host-ip>:8080
```

### WebAssembly Build

```bash
make build-wasm      # Build WASM binary to build/wasm/
make serve-wasm      # Build and serve locally at http://localhost:8080
```

The WASM build is also deployed to [GitHub Pages](https://opd-ai.github.io/venture/) on every push to `main`.

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build client and server for the current platform |
| `make build-all` | Build for Linux, Windows, and macOS |
| `make build-wasm` | Build WebAssembly version |
| `make test` | Run all tests (`go test -v ./...`) |
| `make test-coverage` | Run tests with coverage report |
| `make test-race` | Run tests with race detection |
| `make bench` | Run benchmarks |
| `make lint` | Run `go vet` and network type validation |
| `make fmt` | Format code with `gofumpt` |
| `make clean` | Remove build artifacts |
| `make release VERSION=x.y.z` | Build all platforms, package, and generate checksums |
| `make docker-build` | Build Docker image for the server |
| `make dev-setup` | Install all dependencies for development |
| `make quality` | Run all quality validation tools |

### Mobile Builds

```bash
make android-apk     # Build Android debug APK
make ios-simulator    # Build for iOS Simulator
```

See [Mobile Build Guide](docs/MOBILE_BUILD.md) for full instructions.

## Configuration

### Client Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-width` | `1920` | Window width in pixels |
| `-height` | `1080` | Window height in pixels |
| `-fullscreen` | `false` | Enable fullscreen mode |
| `-seed` | `random` | World generation seed (uses a random seed if not specified) |
| `-genre` | `random` | Genre (`random`, `fantasy`, `scifi`, `horror`, `cyberpunk`, `postapoc`) |
| `--multiplayer` | `false` | Connect to a remote server instead of hosting locally |
| `--server` | `localhost:8080` | Server address (e.g., `192.168.1.5:8080`) |
| `-high-latency` | `false` | Optimize for Tor/high-latency connections (200–5000ms) |
| `--host-and-play` | `false` | Explicitly enable host-and-play mode (default behavior when `--multiplayer` not specified) |
| `--host-lan` | `false` | Bind server to `0.0.0.0` for LAN access instead of `127.0.0.1` (requires host-and-play mode) |
| `-weather` | — | Weather type (`rain`, `snow`, `fog`, `dust`, `ash`, `neonrain`, `smog`, `radiation`) |
| `-weather-intensity` | `heavy` | Weather intensity (`light`, `medium`, `heavy`, `extreme`) |

### Server Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Server port |
| `-max-players` | `8` | Maximum concurrent players |
| `-seed` | `12345` | World generation seed |
| `-genre` | `fantasy` | Genre ID for world generation |
| `-tick-rate` | `30` | Server update rate in ticks per second |
| `-high-latency` | `false` | Optimize for Tor/high-latency connections (200–5000ms) |
| `-security-audit` | `true` | Run security audit at startup |
| `-stability-monitor` | `true` | Enable stability monitoring |

### Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `fatal` | Logging verbosity (unknown values default to `info`) |
| `LOG_FORMAT` | `json`, `text` | Log output format |

### Mod Configuration

JSON-based mods are placed in the `mods/` directory. Example mods are provided:
- `mods/custom-spawns.json` — Custom entity spawn rules
- `mods/hardcore-mode.json` — Hardcore difficulty settings
- `mods/pvp-zones.json` — PvP zone configuration

## Contributing

Contributions are welcome. See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines and [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for development setup.

## License

MIT License. See [LICENSE](LICENSE) for the full text.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) — a 2D game library for Go
- Inspired by roguelikes such as Dungeon Crawl Stone Soup and Cataclysm DDA
- Gameplay influenced by classic action-RPGs like The Legend of Zelda
