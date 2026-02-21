/*
Package main provides the desktop client application for Venture, a fully procedural
multiplayer action-RPG built with Go 1.24.5+ and the Ebiten 2.9.3 game engine.

# Purpose

This package contains the client-side game application that connects players to
procedurally generated worlds. The client handles rendering, input processing,
audio playback, UI management, and network communication with the game server.
For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool.

# Architecture

The client is organized into several key modules:

  - main.go: Entry point, command-line flag parsing, and game loop startup
  - handlers.go: System initialization and orchestration (80+ game systems)
  - util.go: Helper functions for world generation, entity spawning, and callbacks
  - consts.go: Constants for game configuration and seed offsets

The client follows a layered architecture:
 1. Input Layer: Keyboard, mouse, gamepad, and touch input handling
 2. Systems Layer: ECS systems for movement, combat, AI, progression, etc.
 3. Rendering Layer: Sprite generation, tiles, particles, lighting, UI
 4. Audio Layer: Procedural music composition and sound effects
 5. Network Layer: Client-server communication with prediction and lag compensation

# Build Requirements

Desktop platforms:
  - Go 1.24.5 or later
  - Linux: X11 libraries (libc6-dev, libgl1-mesa-dev, libx* packages, libasound2-dev)
  - macOS: Xcode command-line tools
  - Windows: No additional dependencies

WebAssembly:
  - GOOS=js GOARCH=wasm go build

Build tags:
  - Excludes mobile platforms with: //go:build !android && !ios

# Usage Examples

Run with default settings:

	go run ./cmd/client

Run with custom seed and genre:

	go run ./cmd/client -seed 12345 -genre sci-fi -width 1920 -height 1080

Run in multiplayer mode:

	go run ./cmd/client -multiplayer -server localhost:8080

Host a local server and auto-connect (LAN party mode):

	go run ./cmd/client --host-and-play --host-lan

Enable all features for testing:

	go run ./cmd/client -verbose -profile -weather rain -weather-intensity heavy

# Command-Line Flags

Display:

	-width, -height: Window dimensions (default: 1920x1080)
	-fullscreen: Run in fullscreen mode

World Generation:

	-seed: World generation seed (default: random)
	-genre: Theme selection (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)

Weather:

	-weather: Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation)
	-weather-intensity: Weather strength (light, medium, heavy, extreme)

Multiplayer:

	-multiplayer: Enable multiplayer mode
	-server: Server address (e.g., localhost:8080)
	-host-and-play: Start embedded server and auto-connect
	-host-lan: Bind server to all interfaces (0.0.0.0) instead of localhost

Development:

	-verbose: Enable detailed logging
	-profile: Enable performance profiling
	-no-tutorial: Disable tutorial system

# Key Components

systemsContainer: Central dependency injection container holding references to
all 80+ game systems organized by version (V4-V9 systems).

Initialization Functions:
  - initializeCoreSystems: Core gameplay systems (movement, combat, collision)
  - initializeV4Systems: Vehicles, companions, books, spells, expressions
  - initializeV5Systems: Chat, trade, terrain modification, merchant caravans
  - initializeV6Systems: Federation, portals, bounties, politics, territories
  - initializeV7Systems: Display management and viewport optimization
  - initializeV8Systems: Housing, physics, fluid dynamics, buildings
  - initializeV9Systems: Integration managers (crafting stations, housing)

System Wrappers: Adapter pattern implementations to conform incompatible system
signatures to the ECS System interface (discoverySystemWrapper, etc.).

# Performance

The client is optimized for 60 FPS minimum with:
  - Viewport culling (1,635x speedup)
  - Batch rendering (1,667x speedup)
  - Sprite caching (37x speedup, 95.9% hit rate)
  - Object pooling (2x speedup)
  - Spatial partitioning (quadtree-based entity queries)

Current benchmarks: 106 FPS with 2,000 entities, 73MB memory usage.

# Network Integration

The client supports high-latency connections (200-5000ms) through:
  - Client-side prediction: Immediate local input application
  - Entity interpolation: Smooth remote entity movement
  - Lag compensation: Server-side rewind for fair hit detection
  - Delta compression: Bandwidth-efficient state synchronization

Target bandwidth: <100KB/s per player at 20 updates/second.

# Testing

Tests require Xvfb for CI environments:

	xvfb-run -a go test ./cmd/client

Integration tests verify:
  - System initialization order and dependencies
  - UI callback registration
  - Save/load functionality
  - Multiplayer client connection

Note: Test coverage is intentionally low (0.4%) as this package contains
initialization code with Ebiten dependencies that cannot be tested in CI
without a display server. Core game logic in pkg/ packages has 82.4% average
coverage.

# Related Packages

  - pkg/engine: ECS framework and game systems
  - pkg/procgen: Procedural content generation
  - pkg/rendering: Visual rendering systems
  - pkg/audio: Audio synthesis and music composition
  - pkg/network: Multiplayer networking
  - cmd/server: Dedicated server application
  - cmd/mobile: Mobile application entry point
*/
package main
