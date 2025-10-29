# Project Overview

Venture is a fully procedural multiplayer action-RPG built with Go 1.24 and the Ebiten 2.9 game engine. The project represents an ambitious technical achievement: a complete action-RPG where every aspect—graphics, audio, terrain, items, enemies, and abilities—is generated procedurally at runtime with zero external asset files. The game combines deep roguelike-style procedural generation (inspired by Dungeon Crawl Stone Soup and Cataclysm DDA) with real-time action gameplay reminiscent of The Legend of Zelda and Anodyne.

The architecture is built on an Entity-Component-System (ECS) pattern for maximum flexibility and performance. The system supports multiple genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic), each with distinct visual palettes, entity types, and thematic elements. Multiplayer functionality is designed to support high-latency connections (200-5000ms), including slow connections like onion services (Tor), through client-side prediction and authoritative server architecture.

Currently in Phase 9 (Post-Beta Enhancement), the project has completed Phases 1-8, establishing a robust foundation with comprehensive systems: procedural generation (terrain, entities, items, magic, skills, quests, recipes, stations, environment), visual rendering (sprites, tiles, particles, UI, lighting, patterns, caching, pooling), audio synthesis (waveforms, music, SFX), core gameplay (combat, movement, collision, inventory, progression, AI, death/revival), networking (client-server, prediction, lag compensation), save/load system, and genre blending. The project has achieved Beta status with all core features implemented and tested. All generation systems are deterministic using seed-based algorithms, ensuring reproducible content across clients and sessions. Cross-platform support includes desktop (Linux, macOS, Windows), WebAssembly for browsers, and native mobile builds (iOS, Android).

## Technical Stack

- **Primary Language**: Go 1.24.5+ (developed with 1.24.7)
- **Frameworks**: 
  - Ebiten v2.9.2 (2D game engine with cross-platform graphics, input, and audio)
  - Standard library for most functionality
  - Minimal external dependencies: logrus v1.9.3 (structured logging), zenity v0.10.14 (native dialogs), golang.org/x/image v0.32.0 (extended image processing)
- **Testing**: 
  - Go's built-in testing package (no build tags required)
  - Interface-based dependency injection with stub implementations (StubInput, StubSprite, etc.)
  - Table-driven tests for comprehensive scenario coverage
  - Benchmark tests for performance-critical paths
  - Race detection with `go test -race`
  - Average test coverage: 82.4% across all packages
- **Build/Deploy**: 
  - Single binary distribution via `go build`
  - Cross-platform builds: Linux (x64/ARM64), macOS (x64/ARM64), Windows (x64)
  - WebAssembly build for browser deployment (auto-deploys to GitHub Pages)
  - Mobile builds: Android (AAR/APK/AAB), iOS (XCFramework/IPA)
  - Build flags: `-ldflags="-s -w"` for release builds (reduced binary size)
  - Makefile with comprehensive targets for building, testing, profiling, and deployment

## Code Assistance Guidelines

1. **Maintain Deterministic Generation**: All procedural generation MUST use seed-based deterministic algorithms. Never use `time.Now()` or `math/rand` without seeding. Use `rand.New(rand.NewSource(seed))` for isolated RNG instances. Same seed with same parameters must always produce identical output for multiplayer synchronization and testing reproducibility. Example:
   ```go
   rng := rand.New(rand.NewSource(seed))
   // Use rng instead of global rand functions
   value := rng.Intn(100)
   ```

2. **Follow ECS Architecture Patterns**: Separate concerns into Entities (IDs with component collections), Components (pure data structures with no behavior), and Systems (pure logic operating on entities). Components should only contain data fields, never methods beyond simple getters and the `Type()` method. Systems should be stateless or maintain minimal state. Avoid putting logic in components or storing complex state in entities. Example component:
   ```go
   type PositionComponent struct {
       X, Y float64
   }
   
   func (p PositionComponent) Type() string { return "position" }
   ```

3. **Package Structure and Dependencies**: Follow the established `pkg/` organization with clear boundaries. Packages under `pkg/procgen/` should have minimal external dependencies. The `engine` package is foundational and should not import domain-specific packages. Use interfaces in `interfaces.go` files for public contracts. Avoid circular dependencies by keeping dependency flow one-directional (engine ← procgen ← rendering). Each package must have a `doc.go` file with comprehensive package documentation explaining purpose, key concepts, and usage examples.

4. **Testing Requirements**: Write table-driven tests for functions with multiple scenarios. Test both success and error paths, including validation failures. Verify determinism by generating content twice with the same seed and comparing results. Tests use interface-based dependency injection with stub implementations, enabling testing without Ebiten initialization in CI environments. Target minimum 65% code coverage per package, excluding functions that require Ebiten runtime initialization (e.g., `ebiten.NewImage()`, rendering operations, audio playback). These Ebiten-dependent functions should be isolated and minimized where possible. Include benchmarks for performance-critical generation functions. Example table-driven test:
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

5. **Genre System Integration**: All content generators should support genre-based theming through the `GenreID` parameter in `GenerationParams`. Use genre-specific templates, naming conventions, and visual styles. The genre system provides five core genres: fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic. Reference `pkg/procgen/genre.Registry` for genre definitions. Genre themes should influence entity names, color palettes, item prefixes/suffixes, and ability descriptions. Support for cross-genre blending is implemented.

6. **Performance Targets and Optimization**: All code must meet performance targets: 60 FPS minimum (achieved 106 FPS with 2000 entities), <500MB client memory (achieved 73MB), <2s generation time for world areas, <100KB/s per player network bandwidth. Profile before optimizing using `go test -cpuprofile` and `go test -memprofile`. Use object pooling for frequently allocated objects (sync.Pool). Implement spatial partitioning (quadtrees/grids) for entity queries. Cache generated content where determinism allows (sprite caching achieves 95.9% hit rate). Avoid allocations in hot paths (game loop, rendering). Run benchmarks to verify performance: `go test -bench=. -benchmem`. The rendering system implements four primary optimizations: viewport culling (1,635x speedup), batch rendering (1,667x speedup), sprite caching (37x speedup), and object pooling (2x speedup) for a combined 1,625x performance improvement.

7. **Error Handling and Validation**: Return errors rather than panicking (except for programmer errors in init). Wrap errors with context using `fmt.Errorf("context: %w", err)`. Implement validation methods for all generators that verify output meets quality thresholds. Check all error returns—no unchecked errors. Use structured logging with logrus for error context and debugging. Validation example:
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

- **Domain**: Procedural action-RPG with real-time combat, multiplayer co-op, and infinite content generation. Core gameplay loop involves exploring procedurally generated dungeons, fighting generated enemies, collecting generated loot, progressing through generated skill trees, completing generated quests, and interacting with crafting stations. The technical challenge is creating engaging, balanced content entirely through algorithms without artist-created assets. Target users are action-RPG and roguelike enthusiasts seeking infinite replayability with procedurally generated content.

- **Architecture**: Modular package-based design with clear separation of concerns. Core packages: `engine` (ECS framework, game loop, systems for movement, collision, combat, inventory, progression, AI, death/revival, audio management), `procgen` (all generation systems including terrain, entities, items, magic, skills, quests, recipes, stations, environment, genre registry), `rendering` (visual generation including sprites, tiles, particles, UI, lighting, patterns, caching, pooling), `audio` (sound synthesis including waveforms, music composition, SFX), `network` (multiplayer with client-server architecture, prediction, lag compensation), `combat` (combat mechanics), `world` (state management), `saveload` (persistent game state), `logging` (structured logging), `hostplay` (LAN party mode), `mobile` (touch input for mobile platforms). Generation systems are independent and composable. Client-server network architecture with authoritative server for multiplayer.

- **Key Directories**:
  - `cmd/client/` - Game client application with Ebiten integration and WebAssembly support
  - `cmd/server/` - Dedicated server application for multiplayer
  - `cmd/mobile/` - Mobile application entry point for iOS and Android
  - `cmd/*test/` - CLI tools for testing generators offline (terraintest, entitytest, itemtest, magictest, skilltest, questtest, genretest, genreblend, rendertest, audiotest, movementtest, inventorytest, tiletest, humanoidtest, monstertest, merchanttest, silhouettetest, anatomytest, perftest, cachetest, etc.)
  - `pkg/engine/` - ECS framework with entities, components, systems (movement, collision, combat, inventory, progression, AI, input, render, audio manager, death/revival), quadtree spatial partitioning
  - `pkg/procgen/` - Root procedural generation package with generator interface and parameters
  - `pkg/procgen/terrain/` - BSP and cellular automata dungeon generation with room types
  - `pkg/procgen/entity/` - Monster, NPC, boss generation with stats and behavior
  - `pkg/procgen/item/` - Weapon, armor, consumable generation with rarity system
  - `pkg/procgen/magic/` - Spell and ability generation with elemental types
  - `pkg/procgen/skills/` - Skill tree generation with prerequisites and unlocks
  - `pkg/procgen/quest/` - Quest generation with objectives and rewards
  - `pkg/procgen/recipe/` - Crafting recipe generation
  - `pkg/procgen/station/` - Crafting station generation
  - `pkg/procgen/environment/` - Environmental effects and ambience
  - `pkg/procgen/genre/` - Genre definitions, registry, and cross-genre blending
  - `pkg/rendering/` - Root rendering package with interfaces and types
  - `pkg/rendering/palette/` - Color palette generation per genre
  - `pkg/rendering/shapes/` - Procedural shape generation
  - `pkg/rendering/sprites/` - Runtime sprite generation with silhouette analysis
  - `pkg/rendering/tiles/` - Tile rendering system
  - `pkg/rendering/particles/` - Particle effects with pooling
  - `pkg/rendering/ui/` - UI rendering (menus, HUD, inventory, character sheet, skills, quests, map, crafting)
  - `pkg/rendering/lighting/` - Dynamic lighting system
  - `pkg/rendering/patterns/` - Texture pattern generation
  - `pkg/rendering/cache/` - Sprite caching system (95.9% hit rate)
  - `pkg/rendering/pool/` - Object pooling for rendering
  - `pkg/audio/synthesis/` - Waveform generation (sine, square, triangle, sawtooth)
  - `pkg/audio/music/` - Procedural music composition with genre themes
  - `pkg/audio/sfx/` - Sound effect generation for combat, movement, UI
  - `pkg/network/` - Multiplayer networking with client-server, prediction, lag compensation, state synchronization
  - `pkg/combat/` - Combat mechanics (damage calculation, status effects)
  - `pkg/world/` - World state management
  - `pkg/saveload/` - Save/load system with JSON serialization
  - `pkg/logging/` - Structured logging with logrus
  - `pkg/hostplay/` - LAN party "host-and-play" mode
  - `pkg/mobile/` - Touch input handling for mobile platforms
  - `pkg/visualtest/` - Visual testing utilities
  - `docs/` - Comprehensive documentation (ARCHITECTURE, TECHNICAL_SPEC, CONTRIBUTING, DEVELOPMENT, TESTING, API_REFERENCE, ROADMAP, USER_MANUAL, GETTING_STARTED, PERFORMANCE, MOBILE_BUILD, GITHUB_PAGES, and more)
  - `examples/` - Standalone demonstration applications for major systems (complete_dungeon_generation, genre_blending_demo, audio_demo, prediction_demo, phase3_demo, movement_collision_demo, combat_demo, network_demo, multiplayer_demo, lag_compensation_demo, terrain_entity_integration, optimization_demo, animation_demo, color_demo, environment_demo)
  - `scripts/` - Build scripts for cross-platform and mobile builds
  - `web/` - WebAssembly deployment files for GitHub Pages
  - `.github/workflows/` - CI/CD pipelines for testing, building, and deployment

- **Configuration**: Generators use `procgen.GenerationParams` struct with fields: `Difficulty` (0.0-1.0 scaling), `Depth` (dungeon level/progression), `GenreID` (theme selector), `Custom` (map[string]interface{} for generator-specific params). Client accepts command-line flags: `-width`, `-height`, `-seed`, `-genre`, `-multiplayer`, `-server`, `--host-and-play`, `--host-lan`, `-port`. Server accepts: `-port`, `-max-players`. Development dependencies: Go 1.24.5+; Linux requires X11 libraries (libc6-dev, libgl1-mesa-dev, libxcursor-dev, libxi-dev, libxinerama-dev, libxrandr-dev, libxxf86vm-dev, libasound2-dev, pkg-config, libx11-dev); macOS requires Xcode tools; Windows has no additional requirements.

## Quality Standards

- **Testing Requirements**: Maintain minimum 65% code coverage per package, excluding functions that require Ebiten runtime initialization (e.g., `ebiten.NewImage()`, rendering operations, audio playback). Ebiten-dependent functions cannot be tested in CI environments without X11/graphics context and should be isolated to minimize untestable surface area. Current coverage by package: engine 50.0%, procgen 100%, procgen/entity 92.0%, procgen/environment 95.1%, procgen/genre 100%, procgen/item 91.3%, procgen/magic 88.8%, procgen/quest 91.9%, procgen/skills 85.4%, procgen/station 94.2%, procgen/terrain 93.4%, rendering/lighting 90.9%, rendering/palette 96.3%, rendering/particles 91.6%, rendering/patterns 100%, rendering/tiles 92.2%, rendering/ui 85.4%, audio/synthesis (implicitly high via music/sfx), audio/music (implicitly high), audio/sfx (implicitly high), saveload 66.9%, combat 100%, world 100%, logging 77.8%, visualtest 91.5%. Overall average: 82.4%. All tests must pass without build tags. Use table-driven tests for multiple scenarios. Test both success and error paths. Verify deterministic generation by comparing outputs from same seed. Include benchmarks for generation functions. Run race detector: `go test -race ./...`. Example benchmark:
  ```go
  func BenchmarkGenerate(b *testing.B) {
      gen := NewGenerator()
      params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
      for i := 0; i < b.N; i++ {
          gen.Generate(12345, params)
      }
  }
  ```

- **Code Review Criteria**: All exported functions, types, and constants must have godoc comments starting with the element name. Packages must have `doc.go` files explaining purpose, key concepts, and usage examples. Use `go fmt` for formatting. Pass `go vet` checks. No circular package dependencies. Interfaces in `interfaces.go` files with corresponding `*_test.go` files for interface tests. Follow Go naming conventions (MixedCaps, not snake_case). Keep functions focused and small (<50 lines when possible). Error messages should be lowercase without ending punctuation. Use structured logging with logrus for all significant events, errors, and debugging information. Follow existing logging patterns in the codebase.

- **Documentation Standards**: Every package must have a `doc.go` file explaining purpose, key concepts, and usage examples. Public APIs need comprehensive godoc comments. Complex algorithms should have inline comments explaining the approach. Update README.md files in package directories when adding features. Maintain ARCHITECTURE.md and TECHNICAL_SPEC.md when making architectural changes. CLI tools need help text via `-help` flag. Document generation parameters and their effects. User-facing documentation in docs/ directory includes GETTING_STARTED.md (5-minute quickstart), USER_MANUAL.md (complete gameplay guide), DEVELOPMENT.md (developer setup), CONTRIBUTING.md (contribution guidelines), API_REFERENCE.md (API documentation), and specialized guides for mobile builds, WebAssembly deployment, and performance optimization.

## Development Workflow

- **Building**: Use `go build ./cmd/client` and `go build ./cmd/server` for development. Use Makefile targets: `make build` (client+server), `make build-all` (all platforms), `make build-wasm` (WebAssembly), `make android-apk` / `make ios-ipa` (mobile). Use CLI test tools (terraintest, entitytest, itemtest, magictest, skilltest, genretest, genreblend, rendertest, audiotest, movementtest, inventorytest, tiletest, questtest, etc.) for rapid iteration on generators without running full game. Release builds use: `go build -ldflags="-s -w"` for binary size reduction. Mobile builds require ebitenmobile: `go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest`.

- **Testing**: Run `go test ./...` for all tests. Use `go test -cover ./pkg/...` for coverage reports. Generate HTML coverage: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`. Use `go test -race ./...` to detect race conditions. Run benchmarks: `go test -bench=. -benchmem ./...`. Use Makefile targets: `make test`, `make test-coverage`, `make test-race`, `make bench`. Tests use stub implementations (StubInput, StubSprite, StubAudio, etc.) to avoid Ebiten dependencies in CI environments.

- **Profiling**: Use `go test -cpuprofile=cpu.prof -bench=.` then `go tool pprof cpu.prof` for CPU profiling. Use `go test -memprofile=mem.prof -bench=.` then `go tool pprof mem.prof` for memory profiling. Profile before optimizing to identify actual bottlenecks. Use Makefile targets: `make profile-cpu`, `make profile-mem`. Interactive pprof commands: `top20` (show top functions), `list FuncName` (annotated source), `web` (call graph, requires graphviz). Target frame time: <16.67ms (60 FPS). Current performance: 0.02ms frame time with 2000 entities (50,000 FPS effective).

- **Code Quality**: Run `go fmt ./...` before committing. Use `go vet ./...` to catch common mistakes. Optional: use `golangci-lint run` for comprehensive linting. Ensure no build warnings. Use Makefile targets: `make fmt`, `make lint`. Verify all tests pass before committing.

## Common Patterns and Conventions

- **Generator Interface**: All generators implement `procgen.Generator` interface with `Generate(seed int64, params GenerationParams) (interface{}, error)` and `Validate(result interface{}) error` methods. Generators should be stateless or use minimal state. Return typed results that callers can type-assert. Example implementation in pkg/procgen/terrain/, pkg/procgen/entity/, etc.

- **Seed Derivation**: Use `procgen.SeedGenerator` to derive deterministic sub-seeds from a base world seed. Never use the same seed for different generator types. Pattern: `seedGen := procgen.NewSeedGenerator(worldSeed)` then `entitySeed := seedGen.GetSeed("entity", roomIndex)`. This ensures reproducible but varied content across different content types and locations.

- **Entity Templates**: Entity, item, magic, and skill generators use template-based generation. Templates define base ranges and patterns. Actual values are generated within template constraints using RNG. Templates are organized by genre. See pkg/procgen/entity/templates.go, pkg/procgen/item/templates.go, etc.

- **Rarity System**: Consistent rarity levels across all content: Common (1.0x multiplier), Uncommon (1.2x), Rare (1.5x), Epic (2.0x), Legendary (3.0x). Rarity affects stats, drop chances, and visual presentation (color tints, particle effects). Higher depth increases rarity chances. Rarity calculation uses depth-dependent probability curves.

- **Stat Scaling**: Stats scale with level and depth. Formula pattern: `baseStat * (1.0 + 0.15 * (level - 1)) * rarityMultiplier`. Difficulty parameter affects level calculation. Maintain balance between offense (damage, attack) and defense (health, armor, defense). Combat system uses these stats for damage calculation.

- **Enum String Methods**: All enum types should have `String()` methods returning human-readable names. Include "Unknown" case for invalid values. Example:
  ```go
  func (e EntityType) String() string {
      switch e {
      case TypeMonster: return "Monster"
      case TypeBoss: return "Boss"
      case TypeNPC: return "NPC"
      default: return "Unknown"
      }
  }
  ```

- **Component Type Methods**: All components must implement `Type() string` method returning their type identifier. Type identifiers should be lowercase, single-word strings matching the component name without "Component" suffix. Example: PositionComponent returns "position", VelocityComponent returns "velocity". Used for component lookup in entities.

- **System Update Pattern**: Systems implement `Update(deltaTime float64)` method that processes all relevant entities. Use `world.GetEntitiesWith(componentTypes...)` for efficient entity queries. Systems should operate on components, not entities directly. Example: MovementSystem queries entities with "position" and "velocity", updates positions based on velocities and deltaTime.

- **Logging Patterns**: Use structured logging with logrus throughout. Logger instances should be package-level variables initialized in init() or passed via constructors. Log levels: Debug (verbose internal state), Info (significant events), Warn (recoverable issues), Error (operation failures), Fatal (unrecoverable errors). Include context fields: entity IDs, component types, system names. See pkg/logging/ for logging utilities and patterns.

## Multiplayer and Networking

- **Client-Side Prediction**: Client immediately applies player input locally while sending to server. Server validates and sends authoritative state. Client reconciles prediction with server state and replays inputs if misprediction detected. Enables responsive gameplay despite network latency. See `pkg/network/prediction.go` and `examples/prediction_demo/` for implementation.

- **Entity Interpolation**: Server sends snapshots at 20 Hz (50ms intervals). Client buffers snapshots (100-200ms buffer) and interpolates between them for smooth movement despite network jitter. Remote entities are always slightly in the past relative to local player. See `pkg/network/interpolation.go`.

- **Lag Compensation**: Server uses snapshot history for hit detection. When processing player actions (combat, item use), server rewinds to the game state that the client saw when performing the action. Ensures fair hit detection even with high latency (200-5000ms supported). See `pkg/network/lag_compensation.go` and `examples/lag_compensation_demo/`.

- **State Synchronization**: Use delta compression (send only changes). Spatial culling (send only visible/nearby entities). Component filtering (prioritize position > velocity > cosmetic). Target bandwidth: <100KB/s per player at 20 updates/second. Server is authoritative for all game state. See `pkg/network/protocol.go` for message types and serialization.

- **Network Components**: Entities requiring network sync should have NetworkComponent with fields: `Owner` (player ID), `LastSync` (timestamp), `SyncPriority` (importance). Mark components as synced vs. local-only. Client and server use same ECS but different systems active (client has render system, server doesn't). See `pkg/engine/network_component.go`.

- **Multiplayer Setup**: Server started with `./venture-server -port 8080 -max-players 4`. Client connects with `./venture-client -multiplayer -server localhost:8080`. LAN party mode: `./venture-client --host-and-play` starts server and auto-connects. Use `--host-lan` to bind to all network interfaces (default is localhost only). Port fallback: if 8080 occupied, tries 8081-8089 automatically.

## Examples and Demonstrations

The `examples/` directory contains standalone demonstrations of major systems:
- `complete_dungeon_generation/` - Full dungeon generation pipeline with all content types
- `genre_blending_demo/` - Cross-genre blending system with visual comparison
- `audio_demo/` - Audio synthesis and music composition showcase
- `prediction_demo/` - Client-side prediction and reconciliation mechanics
- `phase3_demo/` - Visual rendering system showcase (sprites, tiles, particles)
- `movement_collision_demo/` - Movement system and collision detection
- `combat_demo/` - Combat system with damage calculation and status effects
- `network_demo/` - Networking protocol and serialization
- `multiplayer_demo/` - Complete multiplayer integration with client-server
- `lag_compensation_demo/` - Lag compensation techniques for high-latency scenarios
- `terrain_entity_integration/` - Terrain and entity integration with spawning
- `optimization_demo/` - Rendering optimization systems (culling, batching, caching, pooling)
- `animation_demo/` - Animation system showcase
- `color_demo/` - Color palette and genre-specific color schemes
- `environment_demo/` - Environmental effects and ambience

Run examples with: `go run ./examples/<example_name>` or build with `go build ./examples/<example_name>`

## Genre-Specific Guidelines

- **Fantasy Genre**: Medieval fantasy theme with magic, dungeons, dragons. Entity names use prefixes: "Ancient", "Dark", "Elder", "Fire", "Shadow". Suffixes: "Drake", "Lord", "Wyrm", "Knight", "Sorcerer". Color palette: earth tones, magical glows (blues, purples), warm torch lighting. Item types: swords, bows, staffs, plate armor, leather armor, health potions, mana potions. Magic: elemental spells (fireball, ice shard, lightning bolt).

- **Sci-Fi Genre**: Futuristic technology theme. Entity names: "Combat", "Security", "Battle", "Titan", "Omega" + "Android", "Cyborg", "Mech", "Unit", "Destroyer", "Bot". Color palette: neon blues, metallics (silver, chrome), energy glows (cyan, green), holographic effects. Item types: laser rifles, plasma guns, energy blades, powered armor, nanobots, energy shields, medkits. Technology: force fields, holograms, warp gates.

- **Horror Genre**: Dark, scary atmosphere. Entity names should evoke fear and dread: "Twisted", "Corrupted", "Unholy", "Cursed", "Nightmare" + "Fiend", "Horror", "Specter", "Ghoul", "Abomination". Color palette: dark tones (blacks, deep grays), blood reds, sickly greens, decaying browns. Limited visibility, fog effects, flickering lights. Item types: makeshift weapons, torn armor, tainted potions. Environmental effects: whispers, distant screams, blood trails.

- **Cyberpunk Genre**: Urban future with hacking and neon. Entity names: "Corp", "Enforcer", "Netrunner", "Street", "Shadow" + "Operative", "Assassin", "Hacker", "Mercenary", "Agent". Emphasis on technology and corporate themes. Color palette: neon pinks, purples, blues against dark backgrounds (blacks, dark grays). Rain effects, holographic advertisements. Item types: smart weapons, cyber implants, stealth gear, neural interfaces, hacking tools. Themes: corporate dystopia, digital underground.

- **Post-Apocalyptic Genre**: Survival in wasteland. Entity names reference mutation and decay: "Irradiated", "Mutant", "Feral", "Scrap", "Wasteland" + "Stalker", "Raider", "Survivor", "Cannibal", "Beast". Color palette: browns, grays, rust tones, dusty yellows, radioactive greens. Emphasis on scarcity and makeshift equipment. Item types: scavenged weapons, patched armor, dirty water, canned food, radiation meds. Environmental effects: dust storms, radiation zones, ruins.

## Anti-Patterns to Avoid

- **Non-Deterministic Generation**: Never use `time.Now()`, `math/rand` (global), or system randomness in generators. Always use seeded RNG instances created with `rand.New(rand.NewSource(seed))`. Test determinism by running generation multiple times with same seed and comparing outputs byte-for-byte.

- **Logic in Components**: Components should only hold data. All behavior belongs in systems. Don't add methods to components beyond simple getters and the `Type()` method. Components are data transfer objects (DTOs) for the ECS pattern. Logic in systems enables flexibility, testing, and optimization.

- **Global State**: Avoid package-level variables that hold mutable state. Generators and systems should be stateless or explicitly manage state through structs. Global state breaks testability and concurrency. Use dependency injection to pass state explicitly.

- **Ignoring Errors**: Always check returned errors. Use `if err != nil` checks. Don't use blank identifiers `_` for errors unless there's a clear reason (e.g., known-safe operations). Wrap errors with context using `fmt.Errorf("operation: %w", err)` for better debugging.

- **Premature Optimization**: Profile before optimizing. Don't sacrifice code clarity for micro-optimizations without benchmarks proving value. Focus on algorithmic improvements (O(n²) → O(n log n)) over micro-optimizations. Use Makefile profiling targets.

- **Breaking Determinism**: Never modify generation algorithms without verifying determinism is preserved. Test with multiple runs of same seed. Determinism is critical for multiplayer synchronization and debugging. Document any intentional non-deterministic features.

- **Tight Coupling**: Keep packages loosely coupled. Use interfaces for dependencies. Don't import higher-level packages from lower-level ones (e.g., engine shouldn't import procgen, procgen shouldn't import rendering). Follow dependency hierarchy: engine ← procgen ← rendering ← UI.

- **Direct Ebiten Types in Tests**: Tests should use stub implementations (StubInput, StubSprite, StubAudio) instead of actual Ebiten types. Ebiten types require display initialization which fails in CI environments. Isolate Ebiten dependencies to production code only.

- **Uncached Generation in Hot Paths**: Don't generate sprites, sounds, or other content in the game loop. Use caching systems (pkg/rendering/cache, object pooling). Generation should happen once during initialization or loading, not every frame.

## Future Phase Awareness

- **Phase 9 (Current)**: Post-Beta Enhancement - IN PROGRESS
  - **Completed**: Death/Revival System (GAP-001, GAP-002, GAP-003), Menu Navigation Standardization (GAP-004)
  - **In Progress**: Commerce & NPC Interaction System (GAP-005), Tutorial System (GAP-006)
  - **Upcoming**: Content expansion (crafting, multiplayer features, environmental effects), UI/UX polish, performance optimization, community features

- **Completed Phases**:
  - **Phase 1**: Architecture & Foundation (ECS framework, project structure) ✅
  - **Phase 2**: Procedural Generation Core (terrain, entities, items, magic, skills, quests) ✅
  - **Phase 3**: Visual Rendering System (palettes, shapes, sprites, tiles, particles, UI) ✅
  - **Phase 4**: Audio Synthesis (waveforms, music, SFX) ✅
  - **Phase 5**: Core Gameplay Systems (combat, movement, collision, inventory, progression, AI, sprite improvements) ✅
  - **Phase 6**: Networking & Multiplayer (client-server, prediction, lag compensation) ✅
  - **Phase 7**: Genre System Enhancement (cross-genre blending, genre gallery) ✅
  - **Phase 8**: Polish & Optimization (client/server integration, input/rendering, save/load, performance optimization, rendering optimizations, structured logging, LAN party mode) ✅

When adding code, consider:
- **Network Synchronization**: Maintain determinism for multiplayer. Use seed-based generation. Test with network latency simulation.
- **Cross-Platform Compatibility**: Test on multiple platforms (desktop, web, mobile). Use interfaces to abstract platform-specific code.
- **Performance**: Maintain 60 FPS target. Profile hot paths. Use caching and pooling where appropriate. Avoid allocations in game loop.
- **User Experience**: Focus on responsive controls, clear UI, helpful feedback. Support dual-exit pattern for menus. Provide keyboard shortcuts.
- **Testing**: Maintain coverage above 65% per package. Use stub implementations for Ebiten dependencies. Write benchmarks for performance-critical code.
- **Documentation**: Update relevant docs when changing architecture, APIs, or user-facing features. Keep inline comments current.
