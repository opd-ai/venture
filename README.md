# Venture - Procedural Action RPG

A fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, gameplay content—is generated at runtime with no external asset files.

## Overview

Venture is a top-down action-RPG that combines the deep procedural generation of modern roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda and Anodyne.

**Key Features:**
- 🎮 Real-time action-RPG combat and exploration
- 🌐 **Play in browser** - WebAssembly build available on [GitHub Pages](https://opd-ai.github.io/venture/)
- 📱 **Native mobile support** - iOS and Android with touch-optimized controls
- 🎲 100% procedurally generated content (maps, items, monsters, abilities, quests)
- 💬 **V5.0 Social Systems** - Player-to-player chat (E2E encrypted), dynamic NPC dialog, image sharing, secure item trading
- 🎨 **V3.0 Enhanced Graphics** - Professional-grade visuals with advanced sprites, lighting, particles, and post-processing
- 💡 **Sophisticated Lighting** - Soft shadows, colored lighting, bloom effects, and genre-specific ambience
- 🌦️ **Rich Weather Systems** - Fluid simulation with rain, snow, fog, and environmental interactions
- 🎵 Procedural audio synthesis for music and sound effects
- 🌐 Multiplayer co-op supporting high-latency connections (200-5000ms, onion services)
- 🎭 Multiple genres (fantasy, sci-fi, post-apocalyptic, horror, cyberpunk)
- 📦 Single binary distribution - no external asset files required

## Project Status

**Current Version:** 5.0.0 Production ✅  
**In Development:** V5.0 Social Systems (Phase 26 In Progress)

Version 5.0 introduces comprehensive social systems for multiplayer communication and interaction while maintaining the game's core focus on procedural generation. All core features (chat, NPC dialog, image sharing, trading) operational. See [Social Systems Guide](docs/SOCIAL_SYSTEMS.md) and [Migration Guide](docs/MIGRATION_V5.md) for details.

**Version 5.0 Progress (Social Systems & Multiplayer Messaging):**
- ✅ **Phase 21 Complete**: Chat System Foundation with E2E encryption, ACK/NACK reliability, profanity filtering, and chat UI
- ✅ **Phase 22 Complete**: NPC Dialog System with Markov chains, genre-specific corpora, personality traits
- ✅ **Phase 23 Complete**: Image Sharing System with chunked transfer, thumbnails, moderation hooks
- ✅ **Phase 24 Complete**: Item Trading System with two-phase commit, proximity validation, trust mechanics
- ✅ **Phase 25 Complete**: Concurrency & Integration with multi-party conversations, message ordering
- ⏳ **Phase 26 In Progress**: Polish & Beta Release (trade UI, documentation, integration tests)

### Version 3.0.0 Achievements

**Enhanced Visual Quality (Phases 15-20):**
- **Enhanced Sprites**: 40% more anatomical detail with pixel-perfect dimensions, facial features, anti-aliasing, and genre-specific variations
- **Advanced Tiles**: Rich texture patterns (stone, wood, metal, organic), smooth transitions, multi-layer depth effects
- **Sophisticated Lighting**: Soft shadows, colored lighting, bloom effects, advanced ambient occlusion
- **Rich Particles**: Comprehensive weather systems (rain, snow, fog, dust), fluid simulation, environmental interactions
- **Polished UI**: Dynamic color palettes, smooth transitions, visual hierarchy, procedural decorations
- **Environmental Detail**: Parallax backgrounds, time-of-day systems, post-processing effects, visual polish rivaling hand-crafted games

**Performance Maintained:**
- 106 FPS with 2000 entities (70% above 60 FPS target)
- 73MB memory (86% below 500MB budget)
- 82.4% test coverage (26% above 65% requirement)
- Sprite cache hit rate: 95.9%

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Build the game
go build -o venture-client ./cmd/client
go build -o venture-server ./cmd/server
```

**Prerequisites:** Go 1.24.5+. Platform-specific dependencies required (Linux: X11 libraries, macOS: Xcode tools, Windows: none). See [Getting Started Guide](docs/GETTING_STARTED.md) for installation commands.

### 2. First Game

```bash
# Start playing (default 1920x1080)
./venture-client

# Or with custom settings
./venture-client -width 2560 -height 1440 -fullscreen -seed 12345 -genre fantasy

# Supported resolutions: 1280x720 (HD), 1920x1080 (Full HD), 2560x1440 (QHD), 3840x2160 (4K)
```

**Visual Features (V3.0 Enhanced Graphics, V7.0 Display Foundation):**
- **Display Scaling (V7.0)**: Dynamic resolution support (1280x720 to 3840x2160) with UI scaling and fullscreen mode
- **Enhanced Sprites**: 40% more detail with anatomical accuracy, facial features, anti-aliasing, and genre variations
- **Advanced Tiles**: Rich procedural textures with smooth transitions and depth effects
- **Sophisticated Lighting**: Soft shadows, colored lighting, bloom effects, and advanced ambient occlusion (disable with `-enable-lighting=false`)
- **Rich Weather**: Comprehensive weather systems with fluid simulation and environmental interactions (disable with `-enable-weather=false`)
- **Polished UI**: Dynamic color palettes with smooth transitions and visual hierarchy
- `-weather <type>`: Choose specific weather: rain, snow, fog, dust, ash (sci-fi: neonrain, smog, radiation)
- `-weather-intensity <level>`: Set intensity: light, medium, heavy, extreme

**Controls:** WASD (move), Space (attack), E (use item), F (interact with merchants/NPCs), 1-5 (cast spells), I (inventory), J (quests), K (skill tree), M (map), C (character), R (crafting), ESC (close menus/pause), F5 (save), F9 (load), H or F1 (help)

**All In-Game Menus** (Dual-Exit: Each menu's key OR ESC):

| Menu | Key | Description |
|------|-----|-------------|
| Inventory | I | Manage items and equipment |
| Character Stats | C | View stats, equipment, attributes |
| Skill Tree | K | Spend skill points, unlock abilities |
| Quest Log | J | Track active and completed quests |
| World Map | M | View explored areas and navigation |
| Crafting | R | Brew potions, enchant items, craft equipment |
| Shop | F | Buy/sell items (when near merchant) |
| Help | H or F1 | View controls and game information |

**Menu Navigation:** All menus support dual-exit: press the menu's letter key again (e.g., I for inventory) OR press ESC. No menu traps!

**Gameplay Systems:**
- **Crafting (R key)**: Brew potions, enchant equipment, and create magic items from gathered materials
- **Commerce (F key)**: Trade with merchants, sell loot, and purchase equipment in settlements
- **Skills & Progression**: Unlock new abilities through the skill tree, gain experience from combat and quests

### 3. Multiplayer

#### Quick Start (LAN Party Mode)
```bash
# Host player: start server and auto-connect (one command!)
./venture-client -host-and-play

# Other players: join the host
./venture-client -multiplayer -server <host-ip>:8080
```

**Host gets IP address:** `ip addr show` (Linux) / `ipconfig` (Windows) / `ifconfig` (macOS)  
**For LAN access:** Add `-host-lan` flag to bind to all interfaces (default is localhost only)

#### Traditional Setup
```bash
# Start a dedicated server
./venture-server -port 8080 -max-players 4

# Connect clients
./venture-client -multiplayer -server localhost:8080
```

**Port Fallback:** If port 8080 is occupied, the system automatically tries ports 8081-8089. Use `-port <num>` to specify a different starting port.

**High-Latency Networks (Tor/Onion Services):** For connections over Tor or other high-latency networks (200-5000ms), use the `-high-latency` flag when starting the server. This optimizes timeouts and buffers for extreme latency conditions. See [docs/TOR_SETUP.md](docs/TOR_SETUP.md) for complete Tor setup instructions or [docs/MULTIPLAYER.md](docs/MULTIPLAYER.md) for general network configuration details.

```bash
# Server optimized for Tor/high-latency connections
./venture-server -high-latency -port 8080
```

**For complete setup instructions, gameplay guide, and all features, see:**
- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Installation and first steps (5 minutes)
- **[User Manual](docs/USER_MANUAL.md)** - Complete gameplay documentation
- **[Multiplayer Guide](docs/MULTIPLAYER.md)** - Network configurations and Tor support
- **[Tor Setup Guide](docs/TOR_SETUP.md)** - Complete Tor/onion service configuration

## Platform Support

Venture runs on multiple platforms:

- **🖥️ Desktop:** Linux, macOS, Windows (x64/ARM64) - Native builds
- **🌐 Web:** Play in browser via [GitHub Pages](https://opd-ai.github.io/venture/)  
  (WebAssembly with full touch support)
- **📱 Mobile:** iOS and Android - Touch-optimized  
  (see [Mobile Build Guide](docs/MOBILE_BUILD.md))

**Touch Input:**
The WebAssembly build fully supports touch input for mobile browsers and touch-capable devices. Virtual controls appear automatically when you touch the screen. See [Touch Input (WASM)](docs/TOUCH_INPUT_WASM.md) for details.

**WebAssembly Deployment:**
The game automatically deploys to GitHub Pages on every push to main. See [GitHub Pages Guide](docs/GITHUB_PAGES.md) for details.

## Documentation

### Quick Access
**New Players:** [Getting Started Guide](docs/GETTING_STARTED.md) (5 minutes) → [User Manual](docs/USER_MANUAL.md)  
**Developers:** [Development Guide](docs/DEVELOPMENT.md) → [API Reference](docs/API_REFERENCE.md)  
**Contributors:** [Contributing Guide](docs/CONTRIBUTING.md)

### Project Information
- **[Roadmap V3](docs/ROADMAP_V3.md)** - Development roadmap and current status
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture and design patterns
- **[Technical Spec](docs/TECHNICAL_SPEC.md)** - Technical specifications and implementation details

### Build & Deployment Guides
- **[Mobile Build Guide](docs/MOBILE_BUILD.md)** - iOS and Android build instructions
- **[GitHub Pages Guide](docs/GITHUB_PAGES.md)** - WebAssembly deployment to GitHub Pages
- **[Cross-Platform Builds](docs/CROSS_PLATFORM_BUILDS.md)** - Building for multiple platforms
- **[CI/CD](docs/CI_CD.md)** - Continuous integration and deployment pipeline
- **[Production Deployment](docs/PRODUCTION_DEPLOYMENT.md)** - Production deployment checklist

### Testing & Quality
- **[Testing Guide](docs/TESTING.md)** - Testing strategy and practices
- **[Performance Guide](docs/PERFORMANCE.md)** - Performance optimization and profiling

### System Documentation
- **[Lighting System](docs/LIGHTING_SYSTEM.md)** - Dynamic lighting implementation
- **[Shadow System](docs/SHADOW_SYSTEM.md)** - Shadow casting and ambient occlusion
- **[Rotation System](docs/ROTATION_SYSTEM_SPEC.md)** - Entity rotation specification
- **[Rotation User Guide](docs/ROTATION_USER_GUIDE.md)** - User guide for rotation controls
- **[Structured Logging](docs/STRUCTURED_LOGGING_GUIDE.md)** - Logging best practices
- **[System Interaction Map](docs/SYSTEM_INTERACTION_MAP.md)** - System dependencies and interactions

### Specialized Topics
- **[Accessibility](docs/ACCESSIBILITY.md)** - Accessibility features and guidelines
- **[Ebiten Guide](docs/EBITEN.md)** - Ebiten engine integration notes
- **[Touch Input (WASM)](docs/TOUCH_INPUT_WASM.md)** - WebAssembly touch input implementation
- **[Release Notes V1.1](docs/RELEASE_NOTES_V1.1.md)** - Version 1.1 release notes

## Contributing

Contributions welcome! See [Contributing Guide](docs/CONTRIBUTING.md) for guidelines and [Development Guide](docs/DEVELOPMENT.md) for setup.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) - A dead simple 2D game library for Go
- Inspired by roguelikes like Dungeon Crawl Stone Soup and Cataclysm DDA
- Gameplay inspired by classic action-RPGs like The Legend of Zelda and Anodyne
