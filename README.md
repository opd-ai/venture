# Venture

A fully procedural multiplayer action-RPG built with Go and Ebiten where all graphics, audio, and gameplay content are generated at runtime from a single binary with no external asset files.

## Description

Venture is a top-down action-RPG that uses deterministic, seed-based procedural generation to create all game content at runtime. The game uses an Entity-Component-System (ECS) architecture where entities are unique identifiers with component collections, components are primarily data structures with convenience accessors for common queries, and systems contain all game logic. Terrain is generated using BSP dungeons, cellular automata caves, L-system forests, and Voronoi biomes. Items, quests, NPCs, spells, and dialog are all procedurally generated based on a genre system supporting fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic themes.

The multiplayer networking layer supports high-latency connections (200–5000ms) suitable for Tor/onion service routing, with client-side prediction, lag compensation, and snapshot synchronization. Voice chat is integrated with party, guild, proximity, and private channels using a built-in codec with spatial audio support. A federation system enables cross-server travel, shared marketplaces, and multi-server guilds. On desktop/native builds, the client automatically starts a localhost server for solo play, requiring no manual server setup; on WebAssembly builds, embedded server/host-and-play is disabled and the client must connect to an existing server.

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
│   ├── engine/             # ECS core, 66 game systems, components, spatial partitioning
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
│   │   ├── synthesis/      # Oscillators, envelopes, synthesis engine
│   │   └── voice.go        # Voice codec for multiplayer voice chat (ADPCM encoding)
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

For detailed performance optimization, see **[Performance Tuning Guide](docs/PERFORMANCE_TUNING.md)**.

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
| `--postprocess-preset` | `cinematic` | Post-processing preset (`fantasy`, `sci-fi`, `horror`, `cyberpunk`, `post-apocalyptic`, `neutral`, `cinematic`) |
| `--postprocess-color-grading` | `true` | Enable color grading effect |
| `--postprocess-vignette` | `true` | Enable vignette effect |
| `--postprocess-chromatic` | `true` | Enable chromatic aberration effect |
| `--postprocess-saturation` | `1.1` | Color grading saturation (0.0–2.0) |
| `--postprocess-contrast` | `1.05` | Color grading contrast (0.0–2.0) |
| `--postprocess-brightness` | `0.02` | Color grading brightness (-1.0 to 1.0) |
| `--postprocess-vignette-intensity` | `0.6` | Vignette intensity (0.0–1.0) |
| `--postprocess-vignette-softness` | `0.4` | Vignette softness (0.0–1.0) |
| `--postprocess-chromatic-intensity` | `0.3` | Chromatic aberration intensity (0.0–1.0) |
| `--palette-harmony` | `triadic` | Color harmony type (`complementary`, `analogous`, `triadic`, `tetradic`, `split-complementary`, `monochromatic`) |
| `--palette-mood` | `vibrant` | Palette mood (`normal`, `bright`, `dark`, `saturated`, `muted`, `vibrant`, `pastel`, `tense`, `calm`, `victorious`, `melancholic`, `energetic`, `mystical`, `ominous`, `serene`, `aggressive`, `playful`, `somber`, `ethereal`, `dangerous`, `peaceful`, `chaotic`, `regal`, `desolate`) |
| `--palette-rarity` | `epic` | Palette rarity/intensity (`common`, `uncommon`, `rare`, `epic`, `legendary`) |
| `-profile` | `true` | Enable performance profiling with frame time tracking |
| `-no-tutorial` | `false` | Disable tutorial for experienced players |
| `--vr` | `false` | Enable experimental VR mode with auto-detection of VR runtime paths (activates stereoscopic rendering, head tracking with mouse fallback, VR controllers, and VR UI) **[Experimental: default build uses stub adapters; build with `-tags vr` for OpenXR hardware integration]** |
| `--force-vr` | `false` | Force experimental VR mode even without detected runtime paths (for testing VR systems without VR software installed) |
| `--verbose` | `true` | Enable verbose debug logging (sets log level to `debug` when `LOG_LEVEL` not set) |

### Server Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Server port |
| `-max-players` | `8` | Maximum concurrent players |
| `-seed` | `12345` | World generation seed |
| `-genre` | `fantasy` | Genre ID for world generation |
| `-terrain-type` | `bsp` | Terrain generator type: `bsp`, `cellular`, `city`, `forest`, `composite`, `grammar`, `maze` |
| `-tick-rate` | `30` | Server update rate in ticks per second |
| `-aerial-sprites` | `true` | Enable aerial-view perspective sprites for top-down gameplay |
| `-high-latency` | `false` | Optimize for Tor/high-latency connections (200–5000ms) |
| `-server-name` | `venture-server` | Server name for federation identity (uses `SERVER_NAME` env if set) |
| `-security-audit` | `true` | Run security audit at startup |
| `-stability-monitor` | `true` | Enable stability monitoring |
| `-simulate-network` | `""` | Simulate network conditions for testing (`low`, `medium`, `high`, `very-high`, `extreme`) |
| `-resilience-metrics` | `true` | Enable network resilience metrics collection |
| `-balance-validate` | `false` | Run combat and economic balance validation at startup |
| `-migration-validate` | `false` | Run save file migration validation at startup |
| `-ux-validate` | `false` | Run user experience journey validation at startup |
| `-enable-mods` | `true` | Enable mod system with sandbox security |
| `-mods-dir` | `mods` | Directory to load mods from |
| `-metrics-port` | `9090` | Port for Prometheus metrics HTTP endpoint |
| `-enable-metrics` | `true` | Enable Prometheus metrics export at /metrics endpoint |
| `-verbose` | `true` | Enable verbose debug logging (sets log level to `debug` when `LOG_LEVEL` not set) |
| `-version` | `false` | Print version information and exit |

### Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error`, `fatal` | Logging verbosity (unknown values default to `info`). **Note:** Takes precedence over `--verbose` flag. When `LOG_LEVEL` is not set, `--verbose=true` (default) sets level to `debug`, `--verbose=false` sets level to `info`. |
| `LOG_FORMAT` | `json`, `text` | Log output format |
| `SERVER_NAME` | Any non-empty string | Server name for federation identity. Overrides the `-server-name` flag default. Used to identify the server in cross-server guild federation and generates a unique ed25519 keypair for secure server authentication. |

### Terrain Generation Types

The `-terrain-type` flag controls which procedural algorithm generates the world layout:

| Type | Description | Best For |
|------|-------------|----------|
| `bsp` | Binary Space Partitioning — Creates dungeon-style layouts with rectangular rooms connected by corridors | Traditional dungeon crawling, fantasy genres |
| `cellular` | Cellular Automata — Generates organic cave systems with natural-looking formations | Cave exploration, underground environments |
| `city` | City Generator — Creates urban layouts with roads, buildings, and districts | Urban exploration, cyberpunk/modern settings |
| `forest` | L-System Trees — Generates forest layouts with procedural tree placement | Natural outdoor environments, wilderness exploration |
| `composite` | Multi-Biome Composite — Combines multiple terrain types using Voronoi regions | Large varied worlds with diverse environments |
| `grammar` | Graph Grammar — Uses L-systems to create dungeons with narrative flow and structured room connections | Story-driven dungeons with meaningful layouts and progression |
| `maze` | Recursive Backtracking Maze — Creates complex winding corridors with optional rooms at dead ends | Puzzle-focused dungeons, labyrinth exploration |

Example usage:
```bash
# Generate a story-driven dungeon with graph grammar
./venture-server -terrain-type grammar -genre fantasy -seed 42

# Generate an organic cave system
./venture-server -terrain-type cellular -genre horror -seed 99

# Generate a cyberpunk city
./venture-server -terrain-type city -genre cyberpunk -seed 2077

# Generate a complex maze with winding corridors
./venture-server -terrain-type maze -genre fantasy -seed 1337
```

### VR Mode (Experimental)

The `--vr` flag enables **experimental** virtual reality mode with support for VR development and testing. When enabled, the client activates four VR-specific systems:

| System | Description |
|--------|-------------|
| **Stereoscopic Rendering** | Renders separate images for each eye with proper eye separation for depth perception |
| **Head Tracking** | Tracks camera orientation (uses mouse fallback when no VR runtime detected) |
| **VR Controller Input** | Supports VR controller input mapping (OpenXR with `-tags vr`, stub otherwise) |
| **VR UI** | Adapts the user interface for VR display with spatial positioning |

**⚠️ Current Limitations:**

VR mode is currently **experimental**. Two adapter tiers are provided:

| Build | Adapter | Head Tracking | Controllers | Haptics |
|-------|---------|---------------|-------------|---------|
| Default (`go build`) | Stub | Mouse fallback | Keyboard fallback | No-op |
| VR build (`go build -tags vr`) | OpenXR 1.x | ✅ xrLocateViews | ✅ Action-input system | ✅ xrApplyHapticFeedback |

The default build uses stub adapters for graceful degradation on systems without a VR runtime.  The `-tags vr` build links against the Khronos OpenXR Loader and provides full hardware head tracking, controller input, and haptic feedback through any OpenXR-compatible runtime (SteamVR, Monado, Meta Link, Windows Mixed Reality).

**What works (default build):**
- VR runtime path detection (SteamVR, Oculus installation paths)
- Stereoscopic dual-eye rendering with configurable eye separation
- Head tracking simulation with mouse fallback
- VR UI layout and spatial positioning
- Stub controller input for development/testing

**What works (VR build, `-tags vr`):**
- All of the above, plus:
- Real-time head orientation and position via `xrLocateViews`
- Physical VR controller axes (trigger, grip, thumbstick) and buttons
- Haptic feedback to controllers

**Runtime Detection:**
- By default (`--vr`), the client checks for VR runtime installations at startup
- If no VR runtime paths are detected, VR mode will not activate
- Use `--force-vr` to enable VR systems without runtime detection (useful for testing stereoscopic rendering)

Example usage:
```bash
# Default build — stub adapters (no SDK required)
./venture-client --vr

# VR build — OpenXR hardware adapters (requires OpenXR loader)
# Linux: sudo apt install libopenxr-loader1 libopenxr-dev
# Windows: install the Khronos OpenXR SDK and set CGO_LDFLAGS
go build -tags vr ./cmd/client && ./venture-client --vr

# Force VR mode for testing stereoscopic rendering
./venture-client --force-vr

# Combine with other settings
./venture-client --force-vr --genre scifi --seed 2077 --fullscreen
```

**Development Status:** OpenXR 1.x hardware adapters are implemented (`pkg/engine/vr_openxr_adapters.go`). Validation on real VR headsets is pending; the feature remains experimental until that validation is complete.

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

