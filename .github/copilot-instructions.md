# Project Overview

Venture is a fully procedural multiplayer action-RPG built with Go and Ebiten. Every aspect of the game—graphics, audio, and gameplay content—is generated at runtime with no external asset files, resulting in a single binary distribution. The game combines deep procedural generation inspired by roguelikes (Dungeon Crawl Stone Soup, Cataclysm DDA) with real-time action gameplay inspired by classics like The Legend of Zelda.

The project targets game developers, contributors, and hobbyists interested in procedural content generation, ECS architecture, and multiplayer game networking. It supports desktop (Linux, macOS, Windows), WebAssembly (browser), and mobile (iOS, Android) platforms. Key features include 100% procedural content, player housing and guild systems, advanced physics (vehicles, fluids, destruction), cross-server federation, and high-latency multiplayer support (200-5000ms for Tor/onion services).

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

6. **Write Table-Driven Tests**: Target ≥65% code coverage per package (current average: 82.4%). Use Go's built-in testing with table-driven test patterns. Include benchmarks for performance-critical code. Use stub implementations (StubInput, StubSprite) for testing without Ebiten runtime.
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
  - `Entity`: Lightweight containers with unique IDs and component collections
  - `Component`: Pure data structures implementing `Type() string`
  - `System`: Logic processors with `Update(entities []*Entity, deltaTime float64)`

- **Key Directories**:
  - `cmd/client/` - Game client entry point
  - `cmd/server/` - Dedicated server entry point
  - `cmd/mobile/` - Mobile platform entry point
  - `pkg/engine/` - Core ECS framework, all game systems, and components
  - `pkg/procgen/` - All procedural generators (terrain, entity, item, quest, magic, skills, etc.)
  - `pkg/rendering/` - Sprite generation, animation, lighting, particles, caching
  - `pkg/audio/` - Audio synthesis for music and sound effects
  - `pkg/network/` - Client-server networking, federation, chat, trade
  - `pkg/combat/` - Combat mechanics and damage calculation
  - `pkg/world/` - World state, housing, territory management
  - `pkg/saveload/` - Game state persistence
  - `examples/` - Test programs and demos for individual systems

- **Configuration**: Use CLI flags for runtime configuration (`-width`, `-height`, `-seed`, `-genre`, `-port`, `-high-latency`). Environment variables for logging: `LOG_LEVEL` (debug/info/warn/error), `LOG_FORMAT` (json/text).

## Quality Standards

- **Test Coverage**: Minimum 65% per package (target 80%+). Run `go test -cover ./pkg/...` to verify.
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

The project has 87 active packages (89.7%) and 10 test/infrastructure-only packages (10.3%). All priority packages have been integrated as of V19.0. See `docs/INTEGRATION_AUDIT.md` for detailed package status.

Key active packages include: `engine` (240K+ LOC), `network` (22K+ LOC), `procgen/*` (all generators active), `rendering/*` (all systems active), `world/*` (housing, economy, territory), `combat`, `saveload`, and all integration packages.

Test/infrastructure packages: `pkg/audit/features`, `pkg/procgen/audit`, `pkg/visualtest`, `pkg/visualtest/parity` - used by CI/CD for quality validation.
