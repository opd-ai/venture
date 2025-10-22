# Project Overview

Venture is a fully procedural multiplayer action-RPG built with Go 1.24 and the Ebiten 2.9 game engine. The project represents an ambitious technical achievement: a complete action-RPG where every aspect—graphics, audio, terrain, items, enemies, and abilities—is generated procedurally at runtime with zero external asset files. The game combines deep roguelike-style procedural generation (inspired by Dungeon Crawl Stone Soup and Cataclysm DDA) with real-time action gameplay reminiscent of The Legend of Zelda and Chrono Trigger.

The architecture is built on an Entity-Component-System (ECS) pattern for maximum flexibility and performance. The system supports multiple genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic), each with distinct visual palettes, entity types, and thematic elements. Multiplayer functionality is designed to support high-latency connections (200-5000ms), including slow connections like onion services (Tor), through client-side prediction and authoritative server architecture.

Currently in Phase 3 (Visual Rendering System), the project has completed Phases 1-2, establishing a robust foundation with comprehensive procedural generation systems for terrain (BSP and cellular automata algorithms), entities (monsters, NPCs, bosses), items (weapons, armor, consumables), magic spells, and skill trees. All generation systems are deterministic using seed-based algorithms, ensuring reproducible content across clients and sessions.

## Technical Stack

- **Primary Language**: Go 1.24.7
- **Frameworks**: 
  - Ebiten v2.9.2 (2D game engine with cross-platform graphics and input)
  - Standard library for most functionality
  - No external dependencies beyond Ebiten ecosystem
- **Testing**: 
  - Go's built-in testing package with `-tags test` flag
  - Table-driven tests for comprehensive scenario coverage
  - Benchmark tests for performance-critical paths
  - Target coverage: 80%+ (current average: 94.3% for procgen packages)
  - Race detection with `go test -race`
- **Build/Deploy**: 
  - Single binary distribution via `go build`
  - Cross-platform builds: Linux, macOS, Windows (x64/ARM64)
  - No external dependencies at runtime
  - Build flags: `-ldflags="-s -w"` for release builds

## Code Assistance Guidelines

1. **Maintain Deterministic Generation**: All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()` or `math/rand` without seeding. Use `rand.New(rand.NewSource(seed))` for isolated RNG instances. Same seed with same parameters must always produce identical output for multiplayer synchronization and testing reproducibility. Example:
   ```go
   rng := rand.New(rand.NewSource(seed))
   // Use rng instead of global rand functions
   value := rng.Intn(100)
   ```

2. **Follow ECS Architecture Patterns**: Separate concerns into Entities (IDs with component collections), Components (pure data structures with no behavior), and Systems (pure logic operating on entities). Components should only contain data fields, never methods beyond simple getters. Systems should be stateless or maintain minimal state. Avoid putting logic in components or storing complex state in entities. Example component:
   ```go
   type PositionComponent struct {
       X, Y float64
   }
   
   func (p PositionComponent) Type() string { return "position" }
   ```

3. **Package Structure and Dependencies**: Follow the established `pkg/` organization with clear boundaries. Packages under `pkg/procgen/` should have minimal external dependencies. The `engine` package is foundational and should not import domain-specific packages. Use interfaces in `interfaces.go` files for public contracts. Avoid circular dependencies by keeping dependency flow one-directional (engine ← procgen ← rendering). Each package must have a `doc.go` file with comprehensive package documentation.

4. **Testing Requirements**: Write table-driven tests for functions with multiple scenarios. Test both success and error paths, including validation failures. Verify determinism by generating content twice with the same seed and comparing results. All tests must use `-tags test` to exclude Ebiten initialization in CI environments. Target minimum 80% code coverage per package. Include benchmarks for performance-critical generation functions. Example table-driven test:
   ```go
   func TestGenerator(t *testing.T) {
       tests := []struct {
           name    string
           seed    int64
           params  procgen.GenerationParams
           wantErr bool
       }{
           {"valid params", 12345, validParams, false},
           {"invalid depth", 12345, invalidParams, true},
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // test implementation
           })
       }
   }
   ```

5. **Genre System Integration**: All content generators should support genre-based theming through the `GenreID` parameter in `GenerationParams`. Use genre-specific templates, naming conventions, and visual styles. The genre system provides five core genres: fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic. Reference `pkg/procgen/genre.Registry` for genre definitions. Genre themes should influence entity names, color palettes, item prefixes/suffixes, and ability descriptions.

6. **Performance Targets and Optimization**: All code must meet performance targets: 60 FPS minimum, <500MB client memory, <2s generation time for world areas, <100KB/s per player network bandwidth. Profile before optimizing using `go test -cpuprofile` and `go test -memprofile`. Use object pooling for frequently allocated objects. Implement spatial partitioning (quadtrees/grids) for entity queries. Cache generated content where determinism allows. Avoid allocations in hot paths (game loop, rendering). Run benchmarks to verify performance: `go test -bench=. -benchmem`.

7. **Error Handling and Validation**: Return errors rather than panicking (except for programmer errors in init). Wrap errors with context using `fmt.Errorf("context: %w", err)`. Implement validation methods for all generators that verify output meets quality thresholds. Check all error returns—no unchecked errors. Validation example:
   ```go
   func (g *Generator) Validate(result interface{}) error {
       terrain := result.(*Terrain)
       walkable := terrain.CountWalkableTiles()
       if float64(walkable) < 0.3 * float64(terrain.Width * terrain.Height) {
           return fmt.Errorf("terrain has insufficient walkable tiles: %d", walkable)
       }
       return nil
   }
   ```

## Project Context

- **Domain**: Procedural action-RPG with real-time combat, multiplayer co-op, and infinite content generation. Core gameplay loop involves exploring procedurally generated dungeons, fighting generated enemies, collecting generated loot, and progressing through generated skill trees. The technical challenge is creating engaging, balanced content entirely through algorithms without artist-created assets.

- **Architecture**: Modular package-based design with clear separation of concerns. Core packages: `engine` (ECS framework), `procgen` (all generation systems), `rendering` (visual generation), `audio` (sound synthesis), `network` (multiplayer), `combat` (mechanics), `world` (state management). Generation systems are independent and composable. Client-server network architecture with authoritative server for multiplayer.

- **Key Directories**:
  - `cmd/client/` - Game client application with Ebiten integration
  - `cmd/server/` - Dedicated server application (multiplayer)
  - `cmd/terraintest/`, `cmd/entitytest/`, etc. - CLI tools for testing generators offline
  - `pkg/engine/` - ECS framework, game loop, entity management
  - `pkg/procgen/terrain/` - BSP and cellular automata dungeon generation
  - `pkg/procgen/entity/` - Monster, NPC, boss generation with stats
  - `pkg/procgen/item/` - Weapon, armor, consumable generation
  - `pkg/procgen/magic/` - Spell and ability generation
  - `pkg/procgen/skills/` - Skill tree generation with prerequisites
  - `pkg/procgen/genre/` - Genre definitions and registry
  - `pkg/rendering/palette/` - Color palette generation (98.4% coverage)
  - `pkg/rendering/shapes/` - Procedural shape generation (100% coverage)
  - `pkg/rendering/sprites/` - Runtime sprite generation (100% coverage)
  - `docs/` - Architecture decisions, technical specs, development guide

- **Configuration**: Generators use `procgen.GenerationParams` struct with fields: `Difficulty` (0.0-1.0 scaling), `Depth` (dungeon level/progression), `GenreID` (theme selector), `Custom` (map[string]interface{} for generator-specific params). Tests use `-tags test` build tag to exclude Ebiten/X11 dependencies. Development on Linux requires X11 libraries: `libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`.

## Quality Standards

- **Testing Requirements**: Maintain minimum 80% code coverage per package (current: engine 81.0%, procgen 100%, entity 95.9%, terrain 96.4%, magic 91.9%, item 93.8%, skills 90.6%, palette 98.4%). All tests must pass with `-tags test` flag. Use table-driven tests for multiple scenarios. Test both success and error paths. Verify deterministic generation by comparing outputs from same seed. Include benchmarks for generation functions. Run race detector: `go test -race ./...`. Example benchmark:
  ```go
  func BenchmarkGenerate(b *testing.B) {
      gen := NewGenerator()
      params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
      for i := 0; i < b.N; i++ {
          gen.Generate(12345, params)
      }
  }
  ```

- **Code Review Criteria**: All exported functions, types, and constants must have godoc comments starting with the element name. Packages must have `doc.go` files. Use `go fmt` for formatting. Pass `go vet` checks. No circular package dependencies. Interfaces in `interfaces.go` files. Follow Go naming conventions (MixedCaps, not snake_case). Keep functions focused and small (<50 lines when possible). Error messages should be lowercase without ending punctuation.

- **Documentation Standards**: Every package must have a `doc.go` file explaining purpose, key concepts, and usage examples. Public APIs need comprehensive godoc comments. Complex algorithms should have inline comments explaining the approach. Update README.md files in package directories when adding features. Maintain ARCHITECTURE.md and TECHNICAL_SPEC.md when making architectural changes. CLI tools need help text via `-help` flag. Document generation parameters and their effects.

## Development Workflow

- **Building**: Use `go build ./cmd/client` and `go build ./cmd/server` for development. Use CLI test tools (`terraintest`, `entitytest`, `itemtest`, `magictest`, `skilltest`, `genretest`, `rendertest`) for rapid iteration on generators without running full game. Release builds use: `go build -ldflags="-s -w"` for binary size reduction.

- **Testing**: Run `go test -tags test ./...` for all tests. Use `go test -tags test -cover ./pkg/procgen/...` for coverage reports. Generate HTML coverage: `go test -tags test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`. Use `go test -tags test -race ./...` to detect race conditions. Run benchmarks: `go test -tags test -bench=. -benchmem ./...`.

- **Profiling**: Use `go test -tags test -cpuprofile=cpu.prof -bench=.` then `go tool pprof cpu.prof` for CPU profiling. Use `go test -tags test -memprofile=mem.prof -bench=.` then `go tool pprof mem.prof` for memory profiling. Profile before optimizing to identify actual bottlenecks.

- **Code Quality**: Run `go fmt ./...` before committing. Use `go vet ./...` to catch common mistakes. Optional: use `golangci-lint run` for comprehensive linting. Ensure no build warnings.

## Common Patterns and Conventions

- **Generator Interface**: All generators implement `procgen.Generator` interface with `Generate(seed int64, params GenerationParams) (interface{}, error)` and `Validate(result interface{}) error` methods. Generators should be stateless or use minimal state. Return typed results that callers can type-assert.

- **Seed Derivation**: Use `procgen.SeedGenerator` to derive deterministic sub-seeds from a base world seed. Never use the same seed for different generator types. Pattern: `seedGen := procgen.NewSeedGenerator(worldSeed)` then `entitySeed := seedGen.GetSeed("entity", roomIndex)`.

- **Entity Templates**: Entity, item, magic, and skill generators use template-based generation. Templates define base ranges and patterns. Actual values are generated within template constraints using RNG. Templates are organized by genre.

- **Rarity System**: Consistent rarity levels across all content: Common (1.0x multiplier), Uncommon (1.2x), Rare (1.5x), Epic (2.0x), Legendary (3.0x). Rarity affects stats, drop chances, and visual presentation. Higher depth increases rarity chances.

- **Stat Scaling**: Stats scale with level and depth. Formula pattern: `baseStat * (1.0 + 0.15 * (level - 1)) * rarityMultiplier`. Difficulty parameter affects level calculation. Maintain balance between offense (damage) and defense (health, armor).

- **Enum String Methods**: All enum types should have `String()` methods returning human-readable names. Include "Unknown" case for invalid values. Example:
  ```go
  func (e EntityType) String() string {
      switch e {
      case TypeMonster: return "Monster"
      case TypeBoss: return "Boss"
      default: return "Unknown"
      }
  }
  ```

## Genre-Specific Guidelines

- **Fantasy Genre**: Medieval fantasy theme with magic, dungeons, dragons. Entity names use prefixes: "Ancient", "Dark", "Elder", "Fire", "Shadow". Suffixes: "Drake", "Lord", "Wyrm", "Knight". Color palette: earth tones, magical glows. Item types: swords, bows, staffs, plate armor, leather armor, potions.

- **Sci-Fi Genre**: Futuristic technology theme. Entity names: "Combat", "Security", "Battle", "Titan", "Omega" + "Android", "Cyborg", "Mech", "Unit", "Destroyer". Color palette: neon blues, metallics, energy glows. Item types: laser rifles, plasma guns, powered armor, nanobots, energy shields.

- **Horror Genre**: Dark, scary atmosphere. Entity names should evoke fear and dread. Color palette: dark tones, blood reds, sickly greens. Limited visibility, fog effects.

- **Cyberpunk Genre**: Urban future with hacking and neon. Emphasis on technology and corporate themes. Color palette: neon pinks, purples, blues against dark backgrounds.

- **Post-Apocalyptic Genre**: Survival in wasteland. Entity names reference mutation and decay. Color palette: browns, grays, rust tones. Emphasis on scarcity and makeshift equipment.

## Anti-Patterns to Avoid

- **Non-Deterministic Generation**: Never use `time.Now()`, `math/rand` (global), or system randomness in generators. Always use seeded RNG instances.

- **Logic in Components**: Components should only hold data. All behavior belongs in systems. Don't add methods to components beyond simple getters and the `Type()` method.

- **Global State**: Avoid package-level variables that hold mutable state. Generators and systems should be stateless or explicitly manage state through structs.

- **Ignoring Errors**: Always check returned errors. Use `if err != nil` checks. Don't use blank identifiers `_` for errors unless there's a clear reason.

- **Premature Optimization**: Profile before optimizing. Don't sacrifice code clarity for micro-optimizations without benchmarks proving value.

- **Breaking Determinism**: Never modify generation algorithms without verifying determinism is preserved. Test with multiple runs of same seed.

- **Tight Coupling**: Keep packages loosely coupled. Use interfaces for dependencies. Don't import higher-level packages from lower-level ones.

## Future Phase Awareness

- **Phase 3 (Current)**: Implementing visual rendering system with procedural sprites, tiles, particles, and UI elements. Focus on genre-specific color palettes and visual themes.

- **Phase 4 (Next)**: Audio synthesis with waveform generation, procedural music composition, and sound effects. Will use similar deterministic generation patterns.

- **Phase 5**: Core gameplay systems including combat, inventory, progression, AI, and quests. Will integrate all generation systems into playable mechanics.

- **Phase 6**: Networking and multiplayer with client-side prediction, authoritative server, and lag compensation for high-latency support.

When adding code, consider how it will integrate with upcoming phases. Keep network synchronization in mind (determinism critical). Design with multiplayer state sync needs. Maintain performance targets as systems layer together.
