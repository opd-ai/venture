# Venture - Procedural Action RPG

A fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, gameplay content—is generated at runtime with no external asset files.

## Overview

Venture is a top-down action-RPG that combines the deep procedural generation of modern roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda and Anodyne.

**Key Features:**

### Core Gameplay (Version 1.1)
- 🎮 Real-time action-RPG combat and exploration
- 🌐 **Play in browser** - WebAssembly build available on [GitHub Pages](https://opd-ai.github.io/venture/)
- 📱 **Native mobile support** - iOS and Android with touch-optimized controls
- 🎲 100% procedurally generated content (maps, items, monsters, abilities, quests)
- 🎨 Runtime-generated graphics using procedural techniques
- 💡 **Dynamic lighting** - Atmospheric lighting with flickering torches, spell lights, and genre-specific presets
- 🌦️ **Weather effects** - Procedural rain, snow, fog, and genre-appropriate weather
- 🎵 Procedural audio synthesis for music and sound effects
- 🌐 Multiplayer co-op supporting high-latency connections (200-5000ms, onion services)
- 🎭 Multiple genres (fantasy, sci-fi, post-apocalyptic, horror, cyberpunk)
- 📦 Single binary distribution - no external asset files required

### Advanced Features (Version 2.0)
- 🎯 **360° Rotation & Mouse Aim** - Dual-stick shooter controls with independent aim direction
- 💥 **Projectile Physics** - Realistic projectile trajectories with piercing, bouncing, and explosive effects
- 📐 **Multi-Layer Terrain** - Platforms, bridges, pits, and diagonal walls for complex spatial puzzles
- 🧩 **Procedural Puzzles** - Constraint-solving puzzles (pressure plates, lever sequences, block pushing)
- 🤖 **Behavior Tree AI** - Intelligent enemies with squad tactics and coordinated behaviors
- 🏛️ **Faction System** - Reputation-based relationships affecting NPC behavior and commerce
- 📖 **Dynamic Narratives** - Procedurally generated story arcs that adapt to player actions
- 🎼 **Adaptive Music** - Context-aware composition with character and location leitmotifs
- ⚡ **Screen Shake & Impact** - Visceral combat feedback with hit-stop and visual effects
- 🎬 **Animated Sprites** - Frame-by-frame animations with distance-based LOD optimization
- 🌓 **Dynamic Shadows** - Real-time shadow casting with ambient occlusion for depth
- 💥 **Environmental Destruction** - Destructible objects, carry/throw mechanics, and interactive hazards

## Project Status

**Version:** 2.0 Beta (Phase 14 Complete) - Built on 1.1 Production Foundation ✅

Core features implemented, tested, and production-ready. Version 2.0 (advanced features) complete. All 14 phases implemented. See [Development Roadmap](docs/ROADMAP.md) for detailed progress and milestones.

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
# Start playing
./venture-client

# Or with custom settings
./venture-client -width 1024 -height 768 -seed 12345 -genre fantasy

# Blend two genres for unique experiences (Version 2.0 feature)
./venture-client -genre-blend fantasy-scifi -blend-ratio 0.7

# Enable dynamic lighting and weather effects
./venture-client -enable-lighting -enable-weather
```

**Genre Options:**
- `-genre <id>`: Choose single genre: fantasy, scifi, horror, cyberpunk, postapoc
- `-genre-blend <genre1-genre2>`: Blend two genres (e.g., fantasy-scifi, horror-cyberpunk)
- `-blend-ratio <0.5-1.0>`: Primary genre weight (default: 0.7 = 70% primary, 30% secondary)

**Visual Features:**
- `-enable-lighting`: Adds atmospheric lighting with torches, spell lights, and dynamic shadows
- `-enable-weather`: Enables genre-appropriate weather (rain, snow, fog, etc.)
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
./venture-client --host-and-play

# Other players: join the host
./venture-client -multiplayer -server <host-ip>:8080
```

**Host gets IP address:** `ip addr show` (Linux) / `ipconfig` (Windows) / `ifconfig` (macOS)  
**For LAN access:** Add `--host-lan` flag to bind to all interfaces (default is localhost only)

#### Traditional Setup
```bash
# Start a dedicated server
./venture-server -port 8080 -max-players 4

# Connect clients
./venture-client -multiplayer -server localhost:8080
```

**Port Fallback:** If port 8080 is occupied, the system automatically tries ports 8081-8089. Use `-port <num>` to specify a different starting port.

**For complete setup instructions, gameplay guide, and all features, see:**
- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Installation and first steps (5 minutes)
- **[User Manual](docs/USER_MANUAL.md)** - Complete gameplay documentation

## Platform Support

Venture runs on multiple platforms:

- **🖥️ Desktop:** Linux, macOS, Windows (x64/ARM64) - Native builds
- **🌐 Web:** Play in browser via [GitHub Pages](https://opd-ai.github.io/venture/)  
  (WebAssembly with full touch support)
- **📱 Mobile:** iOS and Android - Touch-optimized  
  (see [Mobile Build Guide](docs/MOBILE_BUILD.md))

**Touch Input:**
The WebAssembly build fully supports touch input for mobile browsers and touch-capable devices. Virtual controls appear automatically when you touch the screen. See [Touch Input Testing Guide](docs/TESTING_TOUCH_INPUT.md) for details.

**WebAssembly Deployment:**
The game automatically deploys to GitHub Pages on every push to main. See [GitHub Pages Guide](docs/GITHUB_PAGES.md) for details.

## Documentation

### Quick Access
**New Players:** [Getting Started Guide](docs/GETTING_STARTED.md) (5 minutes) → [User Manual](docs/USER_MANUAL.md)  
**Developers:** [Development Guide](docs/DEVELOPMENT.md) → [API Reference](docs/API_REFERENCE.md)  
**Contributors:** [Contributing Guide](docs/CONTRIBUTING.md)

### Project Information
- **[Roadmap](docs/ROADMAP.md)** - Development roadmap and current status
- **[Roadmap V2](docs/ROADMAP_V2.md)** - Extended Version 2.0 roadmap with enhanced mechanics
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
- **[Testing Touch Input](docs/TESTING_TOUCH_INPUT.md)** - Touch input testing procedures
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
- **[WASM Touch Testing](docs/WASM_TOUCH_TESTING.md)** - WebAssembly touch testing guide
- **[GAPS](docs/GAPS.md)** - Identified gaps and planned improvements
- **[Release Notes V1.1](docs/RELEASE_NOTES_V1.1.md)** - Version 1.1 release notes

## Contributing

Contributions welcome! See [Contributing Guide](docs/CONTRIBUTING.md) for guidelines and [Development Guide](docs/DEVELOPMENT.md) for setup.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) - A dead simple 2D game library for Go
- Inspired by roguelikes like Dungeon Crawl Stone Soup and Cataclysm DDA
- Gameplay inspired by classic action-RPGs like The Legend of Zelda and Anodyne
