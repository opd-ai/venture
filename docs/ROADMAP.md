# Development Roadmap

## Overview

This document outlines the development plan for Venture, a fully procedural multiplayer action-RPG built with Go 1.24 and Ebiten 2.9. The project successfully completed its initial 8 development phases (Foundation through Polish & Optimization) in October 2025. This roadmap presents the post-Beta enhancement strategy based on comprehensive technical audits and community feedback priorities.

## Project Status: BETA TO PRODUCTION TRANSITION 🚀

**Current Phase:** Post-Beta Enhancement (Phase 9)  
**Version:** 1.0 Beta → 1.1 Production  
**Timeline Horizon:** 6-8 months (January - August 2026)  
**Status:** Active Development - Production Hardening

### Current State Strengths ✅

- **Mature Architecture**: Clean ECS design with 48.9% engine coverage, 100% combat/procgen coverage
- **Deterministic Generation**: Fully seed-based with multiplayer synchronization proven at 200-5000ms latency
- **Cross-Platform**: Desktop (Linux/macOS/Windows), WebAssembly, mobile (iOS/Android) all operational
- **Comprehensive Testing**: 82.4% average test coverage, table-driven tests throughout
- **Production Monitoring**: Structured logging with logrus integrated across all major packages
- **Visual Fidelity**: Phase 5 sprite system complete with silhouette analysis, cache optimization (65-75% hit rate)
- **Performance**: 106 FPS achieved with 2000 entities (exceeds 60 FPS target)

### Technical Debt Priorities 🔧

1. **Test Coverage Gaps**: engine (48.9%), rendering/sprites (50.0%), rendering/patterns (0.0%), saveload (66.9%), network (57.1%)
2. **Death/Revival Mechanics**: Incomplete implementation - player death doesn't disable actions or drop items
3. **Menu UX Consistency**: Not all menus support dual-exit (toggle key + ESC)
4. **Performance Profiling**: Visual sluggishness reported despite 106 FPS metric (frame time variance likely)
5. **LAN Party Experience**: No single-command "host-and-play" mode for local co-op

### Active Development Focus (January 2026) 🎯

Per `IMPLEMENTATION_REPORT.md`, currently completing **GAP-001 through GAP-003** (Spell System):
- ✅ Elemental status effects (Fire/Ice/Lightning/Poison DoT)
- ✅ Shield absorption mechanics
- ✅ Buff/debuff stat modification system
- Coverage improvement: 45.7% → 46.4% (engine package)

**Next milestone:** Complete remaining high-priority gaps (GAP-007, GAP-009, GAP-015), then proceed to Phase 9 enhancements.

---

## Enhancement Categories

Based on comprehensive analysis of `docs/auditors/` and codebase audit, enhancements are organized into 6 categories prioritized by impact, effort, and risk.

---

### Category 1: Core Gameplay Mechanics (MUST HAVE)

**Priority**: Critical  
**Estimated Effort**: Large (3-4 weeks)  
**Dependencies**: None - foundational work

These enhancements address game-breaking gaps and incomplete mechanics identified in `docs/auditors/AUTO_AUDIT.md` and `docs/auditors/EVENT.md`.

#### 1.1: Complete Death & Revival System

**Description**: Implement comprehensive death mechanics including entity immobilization, action disabling, item dropping, and multiplayer revival system.

**Rationale**: Currently referenced in `docs/auditors/EVENT.md:L15-L34` as critical gap. Player death doesn't properly disable actions, items don't drop, and revival mechanics are non-functional. This creates confusing gameplay states and breaks multiplayer cooperation expectations.

**Technical Approach**:
1. Add `DeadComponent` to `pkg/engine/combat_components.go` with `timeOfDeath` field
2. Modify `CombatSystem.Update()` in `pkg/engine/combat_system.go:L123-L187` to check health and apply `DeadComponent`
3. Gate all action systems (movement, attack, spell casting, item use) on absence of `DeadComponent`
4. Implement item drop spawning in death handler: create entity with `PositionComponent`, `SpriteComponent`, and `PickupComponent`
5. Add proximity detection in multiplayer: check for nearby ally entities (distance < 32 units)
6. On revival: remove `DeadComponent`, restore 20% health, trigger resurrection animation

**Success Criteria**:
- Zero-health entities cannot move, attack, cast spells, or use items
- Items from inventory spawn at death location within 64-unit radius
- In multiplayer, ally touch triggers revival restoring 20% max health
- Death and revival properly synchronized across clients via network protocol
- Comprehensive test suite covering single-player death, multiplayer revival, edge cases (simultaneous deaths)

**Risks**:
- Network desync if death/revival not properly synchronized (mitigate with authoritative server approach)
- Item drop physics causing items to spawn inside walls (mitigate with collision validation)
- Revival griefing potential (mitigate with cooldown or animation lock)

**Reference Files**:
- `pkg/engine/combat_system.go` (death detection logic)
- `pkg/engine/combat_components.go` (new DeadComponent)
- `pkg/engine/movement_system.go` (gate on !hasDeadComponent)
- `pkg/engine/spell_casting.go` (gate on !hasDeadComponent)
- `pkg/network/protocol.go` (add death/revival message types)

---

#### 1.2: Menu Navigation Standardization

**Description**: Enforce consistent dual-exit navigation (toggle key + ESC) across all in-game menus as specified in `docs/auditors/MENUS.md`.

**Rationale**: Referenced in `docs/auditors/MENUS.md:L1-L40`. Current implementation has "menu traps" where only specific exit methods work. Inconsistent UX reduces polish and frustrates players accustomed to standard escape patterns.

**Technical Approach**:
1. Audit all menu handlers in `cmd/client/main.go:L600-L1360` (inventory, character, skills, quests, map, settings)
2. Create `MenuState` abstraction in `pkg/engine/menu_system.go` with `openKey`, `isOpen`, `Toggle()`, `Close()` methods
3. Modify input handling to check for both `ebiten.Key{openKey}` and `ebiten.KeyEscape` on every frame
4. Ensure `game.menuOpen` state properly tracks active menu for mutual exclusion
5. Add visual indicators: "[I] Inventory - Press I or ESC to close" in menu headers
6. Implement menu stack for nested menus (e.g., settings submenu)

**Success Criteria**:
- All 8 menus (Inventory, Character, Skills, Quests, Map, Settings, Help, Debug) support both toggle key and ESC
- No menu can become "trapped" requiring specific exit action
- Menu toggle keys documented in central constants file (`pkg/engine/menu_keys.go`)
- Visual indicators present in all menu headers
- Test cases verify both exit methods from all game states (combat, exploration, pause)

**Risks**:
- Key conflict with existing bindings (mitigate with conflict detection system)
- Nested menu confusion (mitigate with breadcrumb navigation)

**Reference Files**:
- `cmd/client/main.go:L1000-L1100` (input handling)
- `pkg/engine/ui_system.go` (menu rendering)
- New file: `pkg/engine/menu_system.go` (abstraction layer)

---

#### 1.3: Commerce & NPC Interaction System

**Description**: Implement shop mechanics with fixed and nomadic merchants, dialog interface, and transaction system as outlined in `docs/auditors/EXPAND.md:L23-L42`.

**Rationale**: Provides gameplay depth, economic loops, and NPC interaction missing from current implementation. Essential for long-term engagement and resource management strategy.

**Technical Approach**:
1. Create `MerchantComponent` in `pkg/engine/commerce_components.go` with `inventory`, `priceMultiplier`, `merchantType` (fixed/nomadic)
2. Add `DialogComponent` with `currentDialog`, `availableOptions []DialogOption`, `state` fields
3. Implement `DialogSystem` in `pkg/engine/dialog_system.go` using interface `DialogProvider` for extensibility
4. Generate merchant inventory using existing `item.Generator` with genre-appropriate stock
5. Add shop UI in `pkg/rendering/ui/shop.go` showing merchant inventory grid and player gold
6. Implement transaction logic: validate gold, transfer items, apply price calculations
7. Spawn nomadic merchants deterministically: use world seed + time cycle to place in random rooms

**Success Criteria**:
- Fixed shopkeepers spawn in town areas (terrain type = settlement)
- Nomadic merchants spawn every 10 minutes at deterministic locations
- Dialog system supports simple text-based interaction with "Buy", "Sell", "Leave" options
- Transactions validate gold availability and inventory space
- Prices scale with item rarity: Common 1.0x, Uncommon 1.5x, Rare 3.0x, Epic 8.0x, Legendary 25.0x
- Interface designed for future extensions (branching dialogs, voice, animations)
- Multiplayer synchronization ensures only one player can transact at a time

**Risks**:
- Economic balance issues (mitigate with configurable price multipliers and testing)
- Merchant spawn determinism breaking in multiplayer (mitigate with server-authoritative spawn logic)

**Reference Files**:
- New files: `pkg/engine/commerce_components.go`, `pkg/engine/dialog_system.go`, `pkg/rendering/ui/shop.go`
- `pkg/procgen/entity/generator.go:L200-L250` (NPC generation)
- `pkg/procgen/terrain/generator.go:L150-L200` (settlement detection)

---

### Category 2: User Experience & Polish (SHOULD HAVE)

**Priority**: High  
**Estimated Effort**: Medium (2-3 weeks)  
**Dependencies**: Category 1 completion for consistent base

#### 2.1: LAN Party "Host-and-Play" Mode

**Description**: Single-command mode that starts authoritative server and auto-connects local client, as specified in `docs/auditors/LAN_PARTY.md`.

**Rationale**: Current workflow requires two terminals and manual connection steps. Ideal LAN party/casual co-op experience should be "player 1 runs one command, shares IP, others join". Reduces barrier to entry for multiplayer.

**Technical Approach**:
1. Add `--host-and-play` flag to `cmd/client/main.go`
2. When flag set, start `server.Run()` in goroutine before game initialization
3. Bind server to `127.0.0.1:8080` by default (security: localhost only)
4. Add `--host-lan` flag to override bind address to `0.0.0.0` with warning
5. Wait for server readiness via channel: `serverReady := make(chan struct{})`
6. Auto-set `-server localhost:8080` flag and enable `-multiplayer`
7. Implement graceful shutdown: defer `server.Shutdown()` and wait for goroutine completion
8. Add port conflict handling: try ports 8080-8089, fail with clear error if all in use

**Success Criteria**:
- Single command `./venture-client --host-and-play` starts server and connects client
- Default bind to localhost (secure) with explicit `--host-lan` for 0.0.0.0
- Server logs clearly indicate "Server ready on localhost:8080"
- Graceful shutdown on client exit: server goroutine terminates cleanly
- Documentation in README.md shows LAN party workflow: host runs command, shares IP, others join
- Integration test verifies server start, client connection, and clean shutdown

**Risks**:
- Port conflicts on shared machines (mitigated by auto-fallback)
- Firewall blocking LAN connections (document firewall rules)
- Resource cleanup issues (mitigated by defer pattern and channel synchronization)

**Reference Files**:
- `cmd/client/main.go:L220-L280` (initialization)
- `cmd/server/main.go:L100-L150` (server startup)
- New file: `pkg/engine/host_and_play.go` (goroutine management)

---

#### 2.2: Character Creation & Tutorial Integration

**Description**: Transform existing tutorial into unified character creation onboarding flow as outlined in `docs/auditors/EXPAND.md:L9-L21`.

**Rationale**: Current tutorial system (`pkg/engine/tutorial.go`) is separate from game start. Modern RPGs integrate character creation with tutorial for seamless onboarding. Improves first-time user experience and reduces player drop-off.

**Technical Approach**:
1. Create `CharacterCreationState` game state in `pkg/engine/game_states.go`
2. Modify main menu to show "New Game" → CharacterCreationState transition
3. Design creation flow: Name input → Class selection (Warrior/Mage/Rogue) → Confirm
4. During creation, show tutorial prompts: "Warriors excel at melee combat", "Use WASD to move"
5. Generate starting stats based on class: Warrior (high HP, low mana), Mage (low HP, high mana), Rogue (balanced)
6. Transition to gameplay after creation with tutorial quest auto-added
7. Ensure creation works in both single-player and multiplayer (sync character data to server)

**Success Criteria**:
- New game flow: Main menu → Character creation → Gameplay (no interruption)
- Tutorial information integrated naturally into creation steps
- Class selection affects starting stats and equipment
- Character data persists through save/load system
- Multiplayer: character creation happens client-side, synced to server on connection
- Extensible for future customization: appearance, stat allocation, ability selection

**Risks**:
- Complexity increase for new players (mitigate with clear step-by-step UI)
- Save/load compatibility with existing saves (mitigate with optional character data in save format)

**Reference Files**:
- `pkg/engine/tutorial.go:L1-L100` (existing tutorial system)
- `cmd/client/main.go:L190-L250` (player entity creation)
- New file: `pkg/engine/character_creation.go`

---

#### 2.3: Main Menu & Game Modes System

**Description**: Implement splash screen menu system with Single-Player/Multi-Player selection and sub-menus as described in `docs/auditors/EXPAND.md:L1-L7`.

**Rationale**: Current implementation starts directly into gameplay with CLI flags. Professional games need proper main menu with save management, server connection UI, and settings before gameplay starts.

**Technical Approach**:
1. Create `MainMenuState` in `pkg/engine/game_states.go` as initial game state
2. Render menu using `pkg/rendering/ui/menu.go` with vertical option list
3. Main menu options: "Single-Player", "Multi-Player", "Settings", "Quit"
4. Single-Player submenu: "New Game" (→ CharacterCreation), "Load Game" (→ save file picker), "Back"
5. Multi-Player submenu: Text field for server address, "Connect" button, "Back"
6. Store menu state in `game.currentState` and route `Update()`/`Draw()` calls
7. Implement smooth transitions: fade out → state change → fade in

**Success Criteria**:
- Game launches to main menu, not directly into gameplay
- All menu options functional and lead to correct states
- Server address field supports standard input (keyboard entry, paste)
- Settings menu persists between sessions (stored in config file)
- Visual polish: title logo (procedurally generated), genre-themed background
- Responsive controls: mouse and keyboard navigation

**Risks**:
- Increased startup time perception (mitigate with quick initial load and background generation)
- UI complexity (mitigate with simple, clean design and consistent navigation)

**Reference Files**:
- New file: `pkg/engine/game_states.go` (state machine)
- New file: `pkg/rendering/ui/main_menu.go`
- `cmd/client/main.go:L1-L50` (CLI flag handling → migrate to menu)

---

#### 2.4: Dynamic Music Context Switching

**Description**: Implement adaptive music system that changes background music based on game context (exploration, combat, boss fight, victory).

**Rationale**: Identified in system integration audit (`docs/FINAL_AUDIT.md` Issue #5). AudioManager currently only plays exploration music without context awareness. Dynamic music significantly enhances player immersion and emotional engagement.

**Technical Approach**:
1. **Context Detection** (2 days):
   - Add music context enum: Exploration, Combat, Boss, Victory, Death
   - Implement context detection in AudioManagerSystem.Update():
     - Check for nearby enemies within 300-unit radius (Combat)
     - Check for boss entities with proximity (Boss)
     - Check player health < 20% (Danger variant)
     - Default to Exploration when no enemies nearby

2. **Music Transition System** (2 days):
   - Implement crossfade between tracks: fade out current over 2 seconds, fade in new
   - Add transition cooldown: prevent rapid switching (minimum 10 seconds in context)
   - Cache generated music tracks to avoid regeneration overhead
   - Priority system: Boss > Combat > Danger > Exploration

3. **Music Generation** (2 days):
   - Extend `pkg/audio/music` to generate context-appropriate compositions:
     - Exploration: slow tempo (80 BPM), major keys, ambient
     - Combat: fast tempo (140 BPM), minor keys, percussion emphasis
     - Boss: intense tempo (160 BPM), dramatic orchestration, bass drops
     - Victory: triumphant (120 BPM), major keys, fanfare elements
   - Use genre-specific instruments per context

4. **Integration & Testing** (1 day):
   - Test transitions in various scenarios: entering combat, boss encounters
   - Verify smooth crossfades without audio pops or clicks
   - Add configuration: `-music-dynamic` flag to enable/disable

**Success Criteria**:
- Music changes appropriately when entering/leaving combat
- Boss music triggers for boss entities and persists until defeat
- Smooth transitions with no audio artifacts
- Context persists for minimum duration (no rapid switching)
- Genre-appropriate instrumentation maintained across contexts

**Effort**: Small (7 days)  
**Priority**: Medium (UX enhancement)  
**Risk**: Low - audio system already functional, adding context detection

**Reference Files**:
- `pkg/engine/audio_manager.go:186` (AudioManagerSystem)
- `pkg/audio/music/generator.go` (music generation)
- `docs/FINAL_AUDIT.md` (Issue #5: Music Context Not Dynamic)

---

### Category 3: Gameplay Depth Expansion (COULD HAVE)

**Priority**: Medium  
**Estimated Effort**: Large (4-5 weeks)  
**Dependencies**: Category 1 and 2 for stable foundation

#### 3.1: Environmental Manipulation System

**Description**: Destructible and constructible terrain with fire propagation as outlined in `docs/auditors/EXPAND.md:L44-L65`.

**Rationale**: Adds strategic depth and emergent gameplay. Players can create shortcuts, block enemy paths, or suffer consequences of fire spells. Aligns with immersive sim design philosophy.

**Technical Approach**:
1. Add `Destructible` flag to `TileComponent` in `pkg/engine/terrain_components.go`
2. Implement `TerrainModificationSystem` in `pkg/engine/terrain_system.go`
3. Weapon-based destruction: check weapon type (pickaxe, bombs) and apply to adjacent tiles on attack
4. Spell-based destruction: fire/explosion spells check area of effect and destroy vulnerable tiles
5. Fire propagation: add `FireComponent` to tiles with `intensity`, `duration`, `spreadChance` fields
6. Fire system: each tick, check 4-connected neighbors, spread based on chance and tile material
7. Construction: add `BuildComponent` with required materials (stone, wood from inventory)
8. Networked sync: send tile modification messages to server, broadcast to all clients
9. Save/load: serialize modified tiles with world state

**Success Criteria**:
- Pickaxe weapons can destroy wall tiles (2-3 hits depending on material)
- Fire/explosion spells destroy tiles in 3×3 area
- Fire spreads to adjacent wooden/flammable tiles at 30% chance per second
- Fire burns for 10-15 seconds before extinguishing
- Players can build walls from inventory materials (10 stone per tile)
- Multiplayer: all clients see same terrain modifications deterministically
- Performance: <5ms per frame for fire propagation with 100 burning tiles

**Risks**:
- Performance impact with many simultaneous fires (mitigate with spatial culling and maximum fire entity limit)
- Map becoming unsolvable (mitigate by marking critical paths as indestructible)
- Network bandwidth for frequent modifications (mitigate with delta compression and batching)

**Reference Files**:
- `pkg/procgen/terrain/generator.go:L100-L200` (tile generation)
- `pkg/engine/combat_system.go:L200-L250` (damage application)
- New file: `pkg/engine/terrain_system.go`

---

#### 3.2: Crafting System (Potions, Enchanting, Magic Items)

**Description**: Comprehensive crafting mechanics for consumables, equipment enhancement, and magic item creation as described in `docs/auditors/EXPAND.md:L67-L88`.

**Rationale**: Provides long-term progression, resource management, and player agency. Integrates with existing item generation system to create hybrid player-crafted/procedurally-enhanced items.

**Technical Approach**:
1. Create `CraftingSystem` in `pkg/engine/crafting_system.go` with recipe management
2. Define `Recipe` struct: `inputs []ItemRequirement`, `output ItemTemplate`, `skillRequired`, `successChance`
3. Implement recipe discovery: found in world, unlocked via skill progression, learned from NPCs
4. Potion brewing: combine herbs (consumable items) + flask → healing/mana/buff potions
5. Enchanting: weapon/armor + enchantment scroll + gold → add stat bonuses or effects
6. Magic item crafting: base item + magic essence + crafting materials → wands/rings/amulets
7. Add crafting UI in `pkg/rendering/ui/crafting.go` showing available recipes and material availability
8. Integrate with skill system: crafting skill level affects success chance and available recipes
9. Deterministic results: use player seed + recipe ID + material seeds for output generation

**Success Criteria**:
- 15+ base recipes across 3 crafting types (potions, enchanting, magic items)
- Recipes discovered through gameplay (world drops, quest rewards, NPC teaching)
- Crafting success chance based on skill level (50% at level 1, 95% at max)
- Failed crafting consumes 50% of materials (player risk/reward decision)
- Crafted items retain procedural generation aesthetic while showing player customization
- Multiplayer: crafting synchronized, recipe knowledge shared in party

**Risks**:
- Economic balance disruption (mitigate with high material costs and failure chance)
- Complexity overwhelming players (mitigate with tutorial quest and gradual recipe introduction)
- Determinism issues with multiplayer crafting (mitigate with server-authoritative results)

**Reference Files**:
- `pkg/procgen/item/generator.go:L150-L300` (item generation)
- `pkg/engine/inventory_system.go:L100-L200` (inventory management)
- New file: `pkg/engine/crafting_system.go`
- New file: `pkg/rendering/ui/crafting.go`

---

### Category 4: Performance & Optimization (SHOULD HAVE)

**Priority**: High  
**Estimated Effort**: Medium (2-3 weeks)  
**Dependencies**: Profiling must happen early to inform all other work

#### 4.1: Visual Performance Optimization

**Description**: Eliminate reported sluggishness through comprehensive profiling and optimization as detailed in `docs/auditors/PERFORMANCE_AUDIT.md`.

**Rationale**: Despite 106 FPS average, players report visible lag. Likely culprit is frame time variance (jank) causing perceived stutter. Profiling required to identify actual bottlenecks rather than premature optimization.

**Technical Approach**:
1. **Assessment Phase** (3 days):
   - CPU profiling: `go test -cpuprofile=cpu.prof -bench=. ./pkg/...`
   - Memory profiling: `go test -memprofile=mem.prof -bench=. ./pkg/...`
   - Frame time tracking: add `FrameTimeTracker` to measure 1%, 0.1% lows
   - Network profiling: log packet sizes and timing in multiplayer

2. **Critical Path Optimization** (4 days):
   - Sprite batch rendering: group draws by texture atlas (reduce draw calls)
   - Entity query caching: cache `GetEntitiesWithComponents()` results until invalidation
   - Collision detection: optimize quadtree (tune cell size, implement lazy updates)
   - Component access: add fast-path type assertions in hot loops

3. **Memory Allocation Reduction** (3 days):
   - Audit allocations in `Update()` loops across all systems
   - Implement object pooling for `StatusEffectComponent`, `ParticleComponent`
   - Reuse slice buffers for entity queries: `entityBuffer []Entity` as system field
   - Profile-guided optimization: focus on top 5 allocation sites

4. **Validation** (2 days):
   - Before/after frame time distribution graphs
   - Performance regression test suite: benchmark all systems
   - 60 FPS validation with 5000 entities (2.5x stress test)

**Success Criteria**:
- 1% low frame time ≥ 16.67ms (no perceptible stutter)
- 0.1% low frame time ≥ 10ms (rare frames OK, no hard drops)
- Allocation rate < 10MB/s during typical gameplay
- Draw calls < 100 per frame (batch rendering effective)
- Frame time variance (std dev) < 2ms
- All optimizations verified with benchmarks showing ≥ 20% improvement

**Risks**:
- Premature optimization reducing code clarity (mitigate with profiling-first approach)
- Platform-specific issues (mitigate with testing on Linux/macOS/Windows)

**Reference Files**:
- `pkg/rendering/sprites/generator.go:L200-L300` (sprite generation)
- `pkg/engine/ecs.go:L50-L100` (entity queries)
- `pkg/engine/collision_system.go:L100-L200` (quadtree)
- `docs/PERFORMANCE.md` (existing performance guide)

---

#### 4.2: Test Coverage Improvement

**Description**: Increase test coverage in packages below 70% to meet project standard: engine (48.9%), rendering/sprites (50.0%), rendering/patterns (0.0%), saveload (66.9%), network (57.1%).

**Rationale**: Current coverage gaps identified in test run output. Low coverage in critical systems (engine, network) increases risk of regressions and production bugs. Project standard is 65%+ per package, 80%+ for critical paths.

**Technical Approach**:
1. **Coverage Analysis** (1 day):
   - Run `go test -coverprofile=coverage.out ./pkg/...`
   - Generate HTML: `go tool cover -html=coverage.out -o coverage.html`
   - Identify uncovered critical paths using coverage annotations

2. **Engine Package** (3 days, target 70%):
   - Focus on untested systems: `AnimationSystem`, `ParticleSystem`, `MenuSystem`
   - Add integration tests for system interactions (combat + status effects)
   - Mock Ebiten dependencies for CI-friendly tests

3. **Rendering Packages** (3 days, target 75%):
   - `rendering/sprites`: test cache miss paths, edge cases in generation
   - `rendering/patterns`: implement first tests (currently 0%), focus on pattern generation algorithms
   - Use interface-based mocks for `ebiten.Image` operations

4. **Network Package** (2 days, target 70%):
   - Add tests for error conditions: connection drops, malformed packets
   - Integration tests for prediction and reconciliation
   - Mock network I/O for deterministic testing

5. **Saveload Package** (1 day, target 75%):
   - Test edge cases: corrupted saves, missing files, version mismatches
   - Test equipment and fog-of-war persistence (currently missing per GAP-015)

**Success Criteria**:
- All packages ≥ 65% coverage
- Critical packages (engine, network, saveload) ≥ 70%
- No untestable code (0% coverage) except Ebiten-dependent rendering (pixel operations)
- New tests follow table-driven pattern matching project convention
- CI pipeline runs all tests with `-tags test` flag successfully

**Risks**:
- Time investment with limited user-facing benefit (mitigate by prioritizing critical paths)
- Mock complexity for Ebiten dependencies (mitigate with thin interface wrappers)

**Reference Files**:
- All `pkg/` directories with `*_test.go` files
- `.github/workflows/test.yml` (CI configuration)

---

#### 4.3: Spatial Partition System Integration

**Description**: Integrate SpatialPartitionSystem from perftest into main game client for improved entity query performance with large entity counts.

**Rationale**: Identified in system integration audit (`docs/FINAL_AUDIT.md`). SpatialPartitionSystem currently only used in `cmd/perftest` but could provide viewport culling optimization for main game when entity counts exceed 500. Performance tests show benefits with 2000+ entities.

**Technical Approach**:
1. **Integration Phase** (2 days):
   - Add SpatialPartitionSystem instantiation in `cmd/client/main.go` after world creation
   - Set world bounds based on terrain dimensions: `NewSpatialPartitionSystem(terrainWidth*32, terrainHeight*32)`
   - Connect to RenderSystem via existing `SetSpatialPartition()` method
   - Register in ECS World with `world.AddSystem(spatialSystem)`

2. **Configuration** (1 day):
   - Add config flag: `-enable-spatial-partition` for opt-in testing
   - Tune cell size parameter (default 64 units, test 32/128 for different map sizes)
   - Add performance metrics: track culled entities vs. rendered

3. **Validation** (1 day):
   - Benchmark with 500, 1000, 2000 entity scenarios
   - Verify viewport culling works correctly (entities outside view not rendered)
   - Ensure no visual glitches (entities popping in/out at viewport edge)

**Success Criteria**:
- Performance improvement: ≥10% frame time reduction with 500+ entities
- No visual artifacts or entity popping
- Graceful degradation if disabled (fallback to render all entities)
- Integration documented in system audit update

**Effort**: Small (4 days)  
**Priority**: Low (optional optimization)  
**Risk**: Low - existing system with proven implementation in perftest

**Reference Files**:
- `pkg/engine/spatial_partition.go:218` (SpatialPartitionSystem definition)
- `pkg/engine/render_system.go:183` (SetSpatialPartition method)
- `cmd/perftest/main.go:43` (current usage example)
- `docs/FINAL_AUDIT.md` (Issue #4)

---

### Category 5: Visual Fidelity Enhancement (COULD HAVE)

**Priority**: Medium  
**Estimated Effort**: Medium (3-4 weeks)  
**Dependencies**: Performance optimization (Category 4) to ensure budget for visual improvements

#### 5.1: Advanced Anatomical Sprite Generation

**Description**: Enhance sprite recognizability with improved anatomical accuracy as detailed in `docs/auditors/VISUAL.md` and `VISUAL_FIDELITY_SUMMARY.md`.

**Rationale**: While Phase 5 established foundation (silhouette analysis, cache, outlines), sprites still lack fine anatomical detail. Current 28×28 pixel constraint limits clarity. Goal is "good enough" visual recognition, not photorealism.

**Technical Approach**:
1. **Enhanced Body Part Templates** (1 week):
   - Refine humanoid templates in `pkg/rendering/sprites/templates.go`
   - Add sub-pixel rendering hints: anti-aliasing for diagonal edges
   - Implement proportional scaling: head = 4×4 pixels, torso = 4×6, legs = 4×8
   - Add facial feature generation: eyes (2 pixels), mouth (1-2 pixels) for close-up views

2. **Layered Composition Enhancement** (1 week):
   - Improve `pkg/rendering/sprites/composition.go` with blend modes
   - Add shadow layer: 50% opacity, 1-pixel offset for depth perception
   - Implement highlight layer: 25% opacity, top-edge for lighting effect
   - Clothing over skin: proper alpha blending and color tinting

3. **Genre-Specific Anatomy Variations** (1 week):
   - Fantasy: organic proportions, flowing robes, medieval armor plates
   - Sci-Fi: geometric augments, angular armor, cybernetic limbs
   - Horror: elongated/distorted proportions, visible bone/decay
   - Refine genre palettes for anatomical emphasis

4. **Silhouette Refinement** (1 week):
   - Improve silhouette scoring algorithm to penalize ambiguous shapes
   - Implement iterative refinement: regenerate sprites scoring < 0.6
   - Add A/B testing framework: compare old vs. new sprite generations

**Success Criteria**:
- Silhouette scores increase by average of 0.1 (e.g., 0.65 → 0.75)
- Player survey: 80%+ can identify entity type at a glance (warrior vs. mage vs. rogue)
- No performance regression: sprite generation still < 5ms per sprite (cached)
- Maintain 28×28 pixel constraint for player characters
- Enemy sprites can scale up to 64×64 for bosses with proportional detail increase

**Risks**:
- Diminishing returns on 28×28 canvas (mitigate with selective detail and color coding)
- Generation complexity increase (mitigate with caching and lazy evaluation)

**Reference Files**:
- `pkg/rendering/sprites/templates.go:L50-L200`
- `pkg/rendering/sprites/humanoid.go:L100-L300`
- `pkg/rendering/sprites/silhouette.go:L200-L400`
- `docs/auditors/VISUAL_FIDELITY_SUMMARY.md` (Phase 5 baseline)

---

#### 5.2: Equipment Visual System

**Description**: Implement visual representation of equipped items on character sprites, showing weapons, armor, and accessories dynamically.

**Rationale**: Identified in system integration audit (`docs/FINAL_AUDIT.md`). EquipmentVisualSystem is defined but never integrated. Players expect to see equipped gear on their character for visual feedback and immersion.

**Technical Approach**:
1. **System Integration** (2 days):
   - Instantiate EquipmentVisualSystem in `cmd/client/main.go` with sprite generator
   - Register with ECS World: `world.AddSystem(equipmentVisualSystem)`
   - Add EquipmentVisualComponent to player and NPC entities
   - Connect to EquipmentComponent for gear change detection

2. **Composite Sprite Generation** (3 days):
   - Implement `regenerateEquipmentLayers()` to create layered sprites
   - Generate equipment overlays: weapon sprites (8×8), armor tints, accessories
   - Composite layers onto base character sprite using alpha blending
   - Cache composited results with invalidation on equipment change

3. **Equipment Sprite Templates** (2 days):
   - Create weapon sprite templates: swords, bows, staves, guns (genre-specific)
   - Create armor visual styles: color tints for light/medium/heavy armor
   - Create accessory indicators: glowing effects for magic items

4. **Performance Optimization** (1 day):
   - Regenerate only when equipment changes (dirty flag tracking)
   - Cache composite sprites (reuse until invalidation)
   - Limit regeneration frequency (max 1/second to prevent spam)

**Success Criteria**:
- Equipped weapons visible in character sprite (held in hand position)
- Armor affects character color/tint (e.g., plate armor = metallic shine)
- Accessories show visual indicators (e.g., magic ring = particle glow)
- No performance regression: composite generation < 10ms per character
- Equipment changes reflected within 1 frame

**Effort**: Medium (8 days)  
**Priority**: Medium (visual polish)  
**Risk**: Medium - sprite composition complexity may require iteration

**Reference Files**:
- `pkg/engine/equipment_visual_system.go:14` (system definition)
- `pkg/rendering/sprites/composition.go` (composite sprite generation)
- `docs/FINAL_AUDIT.md` (Issue #1: Orphaned Systems)

---

#### 5.3: Dynamic Lighting System

**Description**: Implement dynamic lighting with point lights, ambient light, and falloff for enhanced visual atmosphere.

**Rationale**: Identified in system integration audit (`docs/FINAL_AUDIT.md`). lighting.System is defined but not integrated. Dynamic lighting would significantly enhance visual fidelity and atmosphere, especially for horror and dungeon scenarios.

**Technical Approach**:
1. **System Integration** (2 days):
   - Integrate `pkg/rendering/lighting` system into game loop
   - Create LightingSystem wrapper for ECS integration
   - Add LightComponent to entities (torches, spells, player)

2. **Lighting Calculation** (3 days):
   - Implement per-pixel lighting in post-processing pass
   - Support point lights with radius and falloff (linear, quadratic, inverse-square)
   - Add ambient light configuration (adjustable per genre/area)
   - Apply gamma correction for realistic appearance

3. **Light Sources** (2 days):
   - Player torch: 200-unit radius point light (follows player)
   - Spell effects: colored lights (fire=orange, ice=blue, lightning=white)
   - Environmental lights: wall torches, magic crystals, bioluminescence
   - Dynamic lights: flickering torches, pulsing magic

4. **Performance Optimization** (2 days):
   - Light culling: only calculate lights within viewport
   - Light limit: max 16 active lights per frame (configurable)
   - Deferred lighting: calculate lighting in separate pass
   - Add performance toggle: `-enable-lighting` flag for opt-in

**Success Criteria**:
- Visible light/shadow contrast enhances atmosphere
- No performance regression: maintain 60 FPS with up to 16 lights
- Genre-appropriate lighting: horror uses low ambient, fantasy uses warm tones
- Smooth light transitions: no popping or harsh cutoffs
- Configurable quality settings (low/medium/high light count)

**Effort**: Medium (9 days)  
**Priority**: Low (visual enhancement)  
**Risk**: High - performance impact may require significant optimization

**Reference Files**:
- `pkg/rendering/lighting/system.go:11` (lighting system)
- `docs/FINAL_AUDIT.md` (Issue #3: Future feature)
- Future: `pkg/engine/lighting_system.go` (ECS wrapper)

---

#### 5.4: Weather Particle System

**Description**: Implement procedural weather effects including rain, snow, fog, and wind using particle systems.

**Rationale**: Identified in system integration audit (`docs/FINAL_AUDIT.md`). particles.WeatherSystem is defined but not integrated. Weather effects would add environmental dynamism and genre-appropriate atmosphere.

**Technical Approach**:
1. **System Integration** (1 day):
   - Integrate `pkg/rendering/particles/weather.go` into game
   - Add WeatherComponent to world/area entities
   - Connect to existing ParticleSystem for rendering

2. **Weather Types** (3 days):
   - Rain: 1000+ particles, downward velocity, transparency
   - Snow: 500+ particles, slow downward drift, wind sway
   - Fog: Large particles with high alpha, slow movement
   - Sandstorm: Brown particles, horizontal movement, screen obscuration

3. **Environmental Effects** (2 days):
   - Wind: affects particle movement and player sprite tilt
   - Puddles: accumulate during rain (visual effect only)
   - Visibility: fog reduces draw distance (integration with render culling)
   - Audio integration: rain sounds, wind ambiance

4. **Genre-Specific Weather** (2 days):
   - Fantasy: gentle rain, magical sparkle snow
   - Sci-Fi: acid rain (green tint), radioactive fallout
   - Horror: oppressive fog, blood rain (red particles)
   - Post-apocalyptic: ash fall, toxic storms

**Success Criteria**:
- Weather visible and immersive without obscuring gameplay
- Performance: maintain 60 FPS with 1000 weather particles
- Weather transitions smoothly (fade in/out over 5 seconds)
- Genre-appropriate weather types and intensity
- Optional: weather affects gameplay (movement speed in wind)

**Effort**: Medium (8 days)  
**Priority**: Low (atmospheric enhancement)  
**Risk**: Medium - particle count may impact performance

**Reference Files**:
- `pkg/rendering/particles/weather.go:162` (WeatherSystem)
- `pkg/engine/particle_system.go:11` (particle rendering)
- `docs/FINAL_AUDIT.md` (Issue #3: Future feature)

---

### Category 6: Infrastructure & Tooling (SHOULD HAVE)

**Priority**: Medium-High  
**Estimated Effort**: Small-Medium (1-2 weeks)  
**Dependencies**: Should be implemented early to support ongoing development

#### 6.1: Comprehensive System Integration Audit ✅ COMPLETED

**Status**: ✅ **COMPLETED** (October 26, 2025)

**Description**: Complete audit and verification of all game systems as specified in `docs/auditors/SCAN_AUDIT.md` to ensure proper wiring and integration.

**Rationale**: As project matured through 8 phases, systems were added incrementally. Need to verify all systems are properly instantiated, registered with ECS world, and communicating correctly. Prevents "orphaned systems" and integration bugs.

**Completion Summary**:
- ✅ **38 total systems** identified and audited across engine, rendering, and audio packages
- ✅ **33 systems verified** as properly integrated (87% integration rate)
- ✅ **5 systems identified** as orphaned/future features (13% - see below)
- ✅ **Zero critical bugs** found - all core gameplay systems functional
- ✅ **Documentation complete**: `docs/FINAL_AUDIT.md` (810 lines) with comprehensive analysis

**Audit Results**:

*Systems Properly Integrated:*
- 22 ECS systems registered in World (Input, Movement, Collision, Combat, AI, etc.)
- 7 Rendering systems (Camera, Render, Terrain, HUD, Tutorial, Help, Menu)
- 5 UI systems (Inventory, Quest, Character, Skills, Map)
- Server correctly uses 6 authoritative ECS systems (no rendering/UI)

*Orphaned/Future Systems Identified:*
1. **RevivalSystem** - Defined but not integrated (addressed in Category 1.1)
2. **EquipmentVisualSystem** - Not integrated (NEW: addressed in Category 5.2)
3. **SpatialPartitionSystem** - Used only in perftest (NEW: addressed in Category 4.3)
4. **lighting.System** - Not integrated (NEW: addressed in Category 5.3)
5. **particles.WeatherSystem** - Not integrated (NEW: addressed in Category 5.4)

*Additional Recommendations from Audit:*
- Dynamic music context switching (NEW: addressed in Category 2.4)
- System lifecycle hooks (Init/Shutdown) - future enhancement
- Runtime system introspection methods - future enhancement

**Next Steps**:
All future features identified in the audit have been added to this roadmap with appropriate prioritization and technical approaches. See Categories 2.4, 4.3, 5.2-5.4 for implementation plans.

**Reference Files**:
- ✅ `docs/FINAL_AUDIT.md` (completed audit report)
- `cmd/client/main.go:L280-L450` (system initialization verified)
- `pkg/engine/ecs.go:L50-L150` (system registration verified)
- All `pkg/engine/*_system.go` files (audited)

---

#### 6.2: Continuous Logging Enhancement

**Description**: Complete structured logging implementation as specified in `docs/auditors/LOGGING_REQUIREMENTS.md`, ensuring all packages use logrus consistently.

**Rationale**: Logging partially implemented per commit history (commits 2059477, 089c972, 2de8c76). Need to ensure completeness, consistent field usage, and production-ready configuration for debugging and monitoring.

**Technical Approach**:
1. **Audit Current State**:
   - Check all packages for logger initialization
   - Verify log levels used appropriately (Debug for internal state, Info for lifecycle, Error for failures)
   - Ensure structured fields (logrus.Fields) used, not string concatenation

2. **Complete Missing Packages**:
   - Add logging to packages with 0 log statements
   - Focus on critical paths: rendering pipeline, network packet handling, save/load operations

3. **Standardize Field Names**:
   - Create `pkg/logging/fields.go` with standard field name constants
   - Entity operations: `entityID`, `componentType`, `systemName`
   - Procgen: `seed`, `genreID`, `depth`, `difficulty`
   - Network: `playerID`, `latency`, `packetSize`

4. **Performance Optimization**:
   - Add conditional debug logging: `if logger.GetLevel() >= logrus.DebugLevel`
   - Lazy evaluation for expensive field computation
   - Ensure no logging in game loop hot paths above Info level

**Success Criteria**:
- All packages in `pkg/` have logger integration
- Consistent field naming across codebase (using constants)
- No performance regression from logging (benchmark validation)
- Client uses text format with color, server uses JSON format
- Environment variable `LOG_LEVEL` controls verbosity without recompilation
- Documentation in `docs/STRUCTURED_LOGGING_GUIDE.md` updated

**Risks**:
- Performance impact from excessive logging (mitigate with conditional compilation and level checks)

**Reference Files**:
- `pkg/logging/logger.go:L1-L100` (logger implementation)
- `docs/STRUCTURED_LOGGING_GUIDE.md` (usage guide)
- All `pkg/` packages for audit

---

## Phased Implementation Plan

Based on MoSCoW prioritization and dependency analysis, the roadmap is organized into 4 phases over 6-8 months.

### Phase 9.1: Production Readiness (Months 1-2, January-February 2026)

**Focus**: Critical gameplay gaps and infrastructure foundation

**Must Have**:
- ✅ Complete current spell system work (GAP-001, GAP-002, GAP-003) - ONGOING
- [ ] **1.1: Death & Revival System** (2 weeks) - Game-breaking gap
- [ ] **1.2: Menu Navigation Standardization** (1 week) - UX polish
- [ ] **6.1: System Integration Audit** (1 week) - Foundation validation
- [ ] **6.2: Logging Enhancement** (3 days) - Production monitoring

**Should Have**:
- [ ] **4.2: Test Coverage Improvement** (1 week) - Focus on engine, network packages

**Deliverable**: Version 1.1 Alpha - Core mechanics complete, production-ready monitoring

---

### Phase 9.2: Player Experience Enhancement (Months 3-4, March-April 2026)

**Focus**: User onboarding and multiplayer accessibility

**Must Have**:
- [ ] **1.3: Commerce & NPC System** (2 weeks) - Gameplay depth

**Should Have**:
- [ ] **2.1: LAN Party Host-and-Play** (1 week) - Multiplayer UX
- [ ] **2.2: Character Creation Integration** (2 weeks) - Onboarding flow
- [ ] **2.3: Main Menu & Game Modes** (1 week) - Professional presentation

**Could Have**:
- [ ] **4.1: Visual Performance Optimization** (2 weeks) - Address sluggishness

**Deliverable**: Version 1.2 Beta - Complete player-facing experience

---

### Phase 9.3: Gameplay Depth Expansion (Months 5-6, May-June 2026)

**Focus**: Strategic depth and emergent gameplay

**Could Have**:
- [ ] **3.1: Environmental Manipulation** (3 weeks) - Destructible terrain, fire propagation
- [ ] **3.2: Crafting System** (3 weeks) - Potions, enchanting, magic items

**Should Have** (if time permits):
- [ ] **5.1: Advanced Anatomical Sprites** (2 weeks) - Visual fidelity improvement

**Deliverable**: Version 1.3 - Deep gameplay systems

---

### Phase 9.4: Polish & Long-term Support (Months 7-8, July-August 2026)

**Focus**: Final polish, optimization, and production release

**Must Have**:
- [ ] Complete remaining test coverage gaps (target 75%+ all packages)
- [ ] Performance validation and regression testing
- [ ] Documentation updates reflecting all new features
- [ ] Production deployment guide and server hosting documentation

**Should Have**:
- [ ] **5.1: Advanced Anatomical Sprites** (if not completed in Phase 9.3)
- [ ] Balance tuning based on playtesting feedback
- [ ] Accessibility features (colorblind modes, key rebinding)

**Could Have** (stretch goals):
- [ ] Mod support infrastructure
- [ ] Replay system for sharing gameplay moments
- [ ] Achievement system

**Deliverable**: Version 1.5 Production - Polished, production-ready release

---

## Dependencies & Blockers

### Cross-Cutting Dependencies

1. **Performance Budget**: All new features must maintain 60 FPS target
   - **Blocker**: Visual optimization (4.1) should complete before heavy gameplay additions (3.1, 3.2)
   - **Mitigation**: Performance regression tests in CI pipeline

2. **Network Protocol Stability**: Multiplayer features require protocol versioning
   - **Blocker**: Death/revival (1.1), commerce (1.3), environmental manipulation (3.1) all add new message types
   - **Mitigation**: Implement protocol version negotiation in Phase 9.1

3. **Save/Load Compatibility**: New components must serialize correctly
   - **Blocker**: Every system addition requires saveload integration
   - **Mitigation**: Automated serialization tests, version migration system

4. **Test Coverage**: Infrastructure improvements enable safer feature additions
   - **Blocker**: Low test coverage in engine (48.9%) increases regression risk
   - **Mitigation**: Prioritize coverage improvements (4.2) in Phase 9.1

### System Integration Dependencies

- **Death/Revival (1.1)** → Required for balanced commerce system (1.3) to prevent death-exploit loops
- **Menu System (2.3)** → Required before character creation (2.2) for navigation
- **Performance Optimization (4.1)** → Should complete before environmental manipulation (3.1) to ensure fire propagation budget
- **Crafting (3.2)** → Requires commerce system (1.3) for material sourcing and economy balance

---

## Metrics for Success

### Technical Metrics

**Code Quality**:
- Test coverage: ≥ 70% all packages, ≥ 80% critical packages (engine, network, combat)
- Zero critical bugs in production (severity: data loss, crashes, security)
- Build time: < 2 minutes for full build + tests

**Performance**:
- Frame rate: 60 FPS average, ≥ 55 FPS 1% lows (no perceptible stutter)
- Memory: < 500MB client, < 1GB server (4 players)
- Network: < 100KB/s per player, < 150ms perceived latency (with prediction)
- Load times: < 3s world generation, < 1s menu transitions

**Reliability**:
- Crash rate: < 0.1% of play sessions
- Save corruption: < 0.01% of save operations
- Multiplayer desync: < 1 per 100 player-hours

### User Experience Metrics

**Onboarding**:
- Tutorial completion rate: ≥ 80% of new players
- Time to first gameplay: < 2 minutes from launch (including character creation)
- Multiplayer setup success: ≥ 90% of attempts (LAN party mode)

**Engagement**:
- Average session length: ≥ 45 minutes (indicating engagement)
- Save file retention: ≥ 60% players create multiple characters
- Multiplayer adoption: ≥ 30% of playtime in co-op mode

**Quality Perception**:
- Player-reported bugs per hour: < 0.5 (based on issue tracker)
- "Game feels responsive": ≥ 85% positive in surveys
- "Graphics are clear and readable": ≥ 75% positive

### Development Velocity Metrics

**Phase Completion**:
- Phase 9.1: 100% Must Have items, ≥ 80% Should Have items
- Phase 9.2: 100% Must Have items, ≥ 70% Should Have/Could Have items
- Schedule adherence: < 2 weeks variance per phase

**Quality Gates**:
- All tests passing before phase completion
- No regressions in existing functionality (automated regression suite)
- Documentation updated within 3 days of feature completion

---

## Completed Phases (Reference)

### Phase 1-8: Foundation Through Beta (Weeks 1-20, 2025) ✅

All initial phases complete as documented in original roadmap sections. Key achievements:

- **Phase 1-2**: ECS architecture, procedural generation core (terrain, entities, items, magic, skills, quests)
- **Phase 3-4**: Visual rendering system, audio synthesis (100% procedural, zero external assets)
- **Phase 5-6**: Core gameplay systems, networking & multiplayer (200-5000ms latency support)
- **Phase 7**: Genre system with 5 base genres and cross-genre blending
- **Phase 8**: Polish & optimization (save/load, performance, tutorial, documentation)

**Metrics Achieved**:
- Test coverage: 82.4% average (procgen/combat 100%, audio 93%+, rendering 95%+)
- Performance: 106 FPS with 2000 entities (exceeds 60 FPS target)
- Multiplayer: 2-4 players with lag compensation and client-side prediction
- Platforms: Desktop (Linux/macOS/Windows), WebAssembly, mobile (iOS/Android)

---

## Roadmap Maintenance

### Review Cadence

- **Weekly**: Phase progress tracking, blocker identification
- **Bi-weekly**: Sprint planning, priority adjustments based on discoveries
- **Monthly**: Phase retrospectives, metrics review, community feedback integration
- **Quarterly**: Strategic roadmap review, long-term goal alignment

### Priority Adjustment Criteria

Priorities may shift based on:
1. **Critical bugs discovered**: Move to Must Have and address immediately
2. **Community feedback**: Adjust Could Have items based on player demand
3. **Technical discoveries**: Re-estimate effort if complexity higher than anticipated
4. **Performance issues**: Prioritize optimization if targets not met
5. **Platform requirements**: Platform-specific issues may require urgent attention

### Success Criteria for Production Release

Version 1.5 Production release requires:
- ✅ All Phase 9.1 and 9.2 Must Have items complete
- ✅ ≥ 75% of Should Have items complete
- ✅ All technical metrics met (performance, reliability, test coverage)
- ✅ Zero critical or high-severity bugs in issue tracker
- ✅ User manual and developer documentation complete and accurate
- ✅ Deployment guide and server hosting documentation available
- ✅ At least 100 hours of cumulative playtesting with feedback incorporated

---

## Auditor Coverage Analysis

### Auditor Documents Addressed

| Auditor Document | Primary Roadmap Items | Coverage |
|-----------------|----------------------|----------|
| `AUTO_AUDIT.md` | 1.1 (Death/Revival), 6.1 (System Audit ✅) | 100% |
| `AUTO_BUG_AUDIT.md` | 1.1 (Death/Revival), 4.2 (Test Coverage) | 100% |
| `EVENT.md` | 1.1 (Death/Revival mechanics) | 100% |
| `EXPAND.md` | 1.3 (Commerce), 2.2 (Character Creation), 2.3 (Main Menu), 3.1 (Environmental), 3.2 (Crafting) | 100% |
| `LAN_PARTY.md` | 2.1 (Host-and-Play mode) | 100% |
| `MENUS.md` | 1.2 (Menu Navigation) | 100% |
| `PERFORMANCE_AUDIT.md` | 4.1 (Visual Performance) | 100% |
| `VISUAL.md` | 5.1 (Anatomical Sprites) | 100% |
| `VISUAL_FIDELITY_SUMMARY.md` | 5.1 (Building on Phase 5 foundation) | 100% |
| `SCAN_AUDIT.md` | 6.1 (System Audit ✅), **NEW:** 2.4 (Music), 4.3 (Spatial), 5.2-5.4 (Visual) | 100% |
| `LOGGING_REQUIREMENTS.md` | 6.2 (Logging Enhancement) | 100% |
| **`FINAL_AUDIT.md`** ✅ | **6.1 (Completed)**, 2.4, 4.3, 5.2-5.4 | **100%** |

**Total Coverage**: 12/12 auditor documents fully addressed (100%)  
**Latest Audit**: `FINAL_AUDIT.md` (Oct 26, 2025) - comprehensive system integration verification

**New Features from FINAL_AUDIT.md**:
- **2.4**: Dynamic Music Context Switching (audio system enhancement)
- **4.3**: Spatial Partition System Integration (performance optimization)
- **5.2**: Equipment Visual System (visual polish)
- **5.3**: Dynamic Lighting System (atmospheric effects)
- **5.4**: Weather Particle System (environmental dynamics)

### Items NOT Included (With Rationale)

**None** - All auditor suggestions have been incorporated into the roadmap with appropriate prioritization. Items marked "Could Have" or placed in Phase 9.3/9.4 are still planned, just deferred based on:
- **Dependency chains**: Need foundation work first (e.g., performance optimization before heavy features)
- **Scope management**: Ensuring Phase 9.1-9.2 remain achievable within 4 months
- **Risk mitigation**: Core mechanics and UX take priority over advanced features

---

## Conclusion

This roadmap transforms Venture from a feature-complete Beta into a production-ready 1.5 release through 6-8 months of focused enhancement. The phased approach balances:

- **Critical gaps** (death/revival, menu UX) that impact core experience
- **Strategic depth** (commerce, crafting, environmental manipulation) for long-term engagement
- **Technical excellence** (test coverage, performance, system integration) for maintainability
- **Player accessibility** (LAN party mode, character creation, main menu) for wider adoption

Every enhancement is grounded in specific codebase analysis, auditor recommendations, and project architecture patterns. The roadmap is actionable, measurable, and aligned with the project's commitment to deterministic procedural generation, zero-asset architecture, and high-performance gameplay.

**Development team can begin Phase 9.1 work immediately** - all tasks have clear technical approaches, success criteria, and reference files. The structured priority system (MoSCoW) enables flexible response to discoveries while maintaining core milestone commitments.

---

**Document Version**: 2.1 (Post-Beta Enhancement Roadmap with System Audit Results)  
**Last Updated**: October 2025  
**Next Review**: January 15, 2026 (Phase 9.1 completion)  
**Maintained By**: Venture Development Team
