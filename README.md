# Venture - Procedural Action RPG

A fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, gameplay content—is generated at runtime with no external asset files.

## Overview

Venture is a top-down action-RPG that combines the deep procedural generation of modern roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda and Anodyne.

**Key Features:**
- 🎮 Real-time action-RPG combat and exploration
- 📱 **Native mobile support** - iOS and Android with touch-optimized controls
- 🎲 100% procedurally generated content (maps, items, monsters, abilities, quests)
- 🎨 Runtime-generated graphics using procedural techniques
- 🎵 Procedural audio synthesis for music and sound effects
- 🌐 Multiplayer co-op supporting high-latency connections (200-5000ms, onion services)
- 🎭 Multiple genres (fantasy, sci-fi, post-apocalyptic, horror, cyberpunk)
- 📦 Single binary distribution - no external asset files required

## Project Status

**Phase:** 8 (Polish & Optimization) - ✅ COMPLETE  
**Version:** 1.0 Beta  
**Status:** Ready for Beta Release 🎉

All major development phases complete with:
- ✅ 100% procedural content generation (graphics, audio, gameplay)
- ✅ Full multiplayer co-op support (2-4 players, high-latency tolerant)
- ✅ Native mobile support (iOS & Android)
- ✅ Five distinct genres with blending system
- ✅ Comprehensive tutorial and documentation
- ✅ Performance-optimized (106 FPS with 2000 entities)
- ✅ Production-ready save/load system
- ✅ 80%+ test coverage across all packages

**See [Development Roadmap](docs/ROADMAP.md) for complete phase details and milestones.**

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

**Prerequisites:** Go 1.24.5+. Platform dependencies vary (see [Getting Started Guide](docs/GETTING_STARTED.md) for details).

### 2. First Game

```bash
# Start playing
./venture-client

# Or with custom settings
./venture-client -width 1024 -height 768 -seed 12345 -genre fantasy
```

**Controls:** WASD (move), Space (attack), E (interact), I (inventory), ESC (menu)

### 3. Multiplayer

```bash
# Start a server
./venture-server -port 8080 -max-players 4

# Connect clients
./venture-client -server localhost:8080
```

**For complete setup instructions, gameplay guide, and all features, see:**
- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Installation and first steps (5 minutes)
- **[User Manual](docs/USER_MANUAL.md)** - Complete gameplay documentation

## Documentation

**For Players:**
- [Getting Started Guide](docs/GETTING_STARTED.md) - Quick installation and first game (5 minutes)
- [User Manual](docs/USER_MANUAL.md) - Complete gameplay guide and mechanics

**For Developers:**
- [Development Guide](docs/DEVELOPMENT.md) - Setup, workflow, testing, and best practices
- [Contributing Guide](docs/CONTRIBUTING.md) - How to contribute to the project
- [API Reference](docs/API_REFERENCE.md) - Complete API documentation with examples

**Project Information:**
- [Roadmap](docs/ROADMAP.md) - Development phases and milestones
- [Architecture](docs/ARCHITECTURE.md) - Architecture Decision Records (ADRs)
- [Technical Specification](docs/TECHNICAL_SPEC.md) - Complete technical details

## Contributing

Contributions welcome! Please see [Contributing Guide](docs/CONTRIBUTING.md) for guidelines and [Development Guide](docs/DEVELOPMENT.md) for setup instructions.

Quick verification commands:
```bash
# Run tests
go test -tags test ./...

# Format and check code
go fmt ./...
go vet ./...
```

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) - A dead simple 2D game library for Go
- Inspired by roguelikes like Dungeon Crawl Stone Soup and Cataclysm DDA
- Gameplay inspired by classic action-RPGs like The Legend of Zelda and Anodyne
