//go:build !android && !ios
// +build !android,!ios

/*
Package main provides the dedicated multiplayer game server for Venture.

The server implements an authoritative game server architecture that manages
multiplayer game sessions with support for high-latency connections (200-5000ms),
including Tor/onion services. It runs a headless game world simulation and
broadcasts state updates to connected clients.

# Architecture

The server follows a client-server authoritative model where:
  - Server maintains the canonical game state
  - Clients send input commands to the server
  - Server validates and processes all actions
  - Server broadcasts authoritative state updates to clients
  - Lag compensation ensures fair hit detection despite network latency

# Key Systems

The server initializes all core gameplay systems for authoritative simulation:
  - Movement and collision detection
  - Combat system with damage calculation
  - AI behavior for NPCs and enemies
  - Inventory and equipment management
  - Quest and progression tracking
  - Crafting and skill systems
  - Housing and companion systems
  - Vehicle physics and mounting

# Network Protocol

The server uses a TCP-based protocol with the following features:
  - Client-side prediction support with reconciliation
  - Entity interpolation with snapshot buffering
  - Lag compensation with server-side rewinding
  - Delta compression for bandwidth efficiency
  - Spatial culling (only send visible/nearby entities)
  - Component filtering (prioritize critical data)

# World Generation

The server generates a procedural world on startup using:
  - Configurable seed for reproducible worlds
  - Genre-based theming (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
  - BSP dungeon generation with room types
  - Procedural entity spawning (enemies, merchants, NPCs)
  - Dynamic item and loot generation

# Configuration

The server accepts the following command-line flags:

	-port <string>           Server port (default: "8080")
	-max-players <int>       Maximum number of players (default: 4)
	-tick-rate <int>         Server update rate in Hz (default: 20)
	-seed <int64>            World generation seed (default: 12345)
	-genre <string>          Genre ID for world generation (default: "fantasy")
	-verbose                 Enable verbose debug logging
	-aerial-sprites          Enable aerial-view perspective sprites (default: true)
	-high-latency            Use configuration optimized for Tor/onion services

# Performance Targets

The server is optimized for:
  - 60+ TPS (ticks per second) with 100+ entities
  - <500MB memory usage per server instance
  - <100KB/s bandwidth per connected player
  - Sub-second response time for player actions
  - Support for 200-5000ms client latency

# Example Usage

Start a server with default settings:

	./venture-server

Start a server with custom configuration:

	./venture-server -port 9000 -max-players 8 -seed 42 -genre sci-fi -verbose

Start a server optimized for high-latency connections:

	./venture-server -high-latency -tick-rate 10

# Platform Support

The server is supported on:
  - Linux (x64, ARM64)
  - macOS (x64, ARM64)
  - Windows (x64)

The server is NOT supported on mobile platforms (Android, iOS).
Build tags ensure mobile builds are excluded.

# See Also

For client implementation, see cmd/client package.
For network protocol details, see pkg/network package.
For ECS framework, see pkg/engine package.
For procedural generation, see pkg/procgen package.
*/
package main
