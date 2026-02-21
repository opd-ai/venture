# Consolidated Audit Report

**Generated**: 2026-02-21  
**Source**: 108 subpackage AUDIT.md files  
**Repository**: github.com/opd-ai/venture

## Summary

- **Total issues**: 223 (91 open, 132 resolved)
- **Critical**: 0 (0 open) | **High**: 43 (0 open) | **Medium**: 42 (3 open) | **Low**: 138 (88 open)
- **Affected subpackages**: 41 of 108 audited packages have open issues
- **Resolution rate**: 132/223 (59%)

| Severity | Total | Open | Resolved |
|----------|-------|------|----------|
| Critical | 0 | 0 | 0 |
| High | 43 | 0 | 43 |
| Medium | 42 | 3 | 39 |
| Low | 138 | 88 | 50 |
| **Total** | **223** | **91** | **132** |

## Priority Resolution Order

### Phase 1: Critical Issues

_No open critical issues._

### Phase 2: High Priority

_No open high-priority issues._

### Phase 3: Medium Priority

- **pkg/network/resilience** (1 issue)
  - [ ] deterministic procgen
- **pkg/procgen/skills** (1 issue)
  - [ ] ECS: compliance — `Skill` and `SkillTree` types have business logic methods (IsUnlocked, CanLevelUp, Tota
- **pkg/procgen/story** (1 issue)
  - [ ] ECS compliance: `StoryJournalComponent` in `pkg/engine/story_fragment_component.go` has behavior methods (`AddDiscov

### Phase 4: Low Priority

- **cmd/client** (2 issues)
  - [ ] doc.go references Go 1.24 and Ebiten 2.9 (doc.go)
  - [ ] Large function count in handlers.go
- **cmd/mobile** (3 issues)
  - [ ] Nil terrain access
  - [ ] Hardcoded screen dimensions
  - [ ] Silent item generation failures
- **cmd/server** (1 issue)
  - [ ] Global `v9ValidationService` synchronization
- **pkg/audio/music** (2 issues)
  - [ ] determinism
  - [ ] doc coverage
- **pkg/audio/sfx** (1 issue)
  - [ ] doc coverage
- **pkg/audit/features** (2 issues)
  - [ ] Doc coverage
  - [ ] Doc coverage
- **pkg/class** (3 issues)
  - [ ] Silent error handling in `addPrimaryClassStats`, `addSecondaryClassStats`, `addPrestigeClassStats` — errors from `GetCla...
  - [ ] Missing method-level documentation on several exported functions (`GetPrestigeClassDefinition`, `GetAllClasses`, `GetAll...
  - [ ] `RespecCost.MaxCost` cap behavior (10,000g) not documented; respec cost formula not tested at boundary.
- **pkg/companion** (2 issues)
  - [ ] Hardcoded magic numbers for skill decay (0.1), trait clamping (0.0-1.0), LRU limits (1000), and XP values are not config...
  - [ ] `Deserialize()` rebuilds prerequisites as empty and defaults cost to 1 — deserialized skill trees have incomplete struct...
- **pkg/companion/learning** (1 issue)
  - [ ] Documentation
- **pkg/config** (3 issues)
  - [ ] documentation: Missing benchmark tests for validation performance (`validator_test.go:N/A`)
  - [ ] architecture: Genre list dependency on pkg/procgen/dialog creates coupling; consider extracting genre definitions 
  - [ ] maintainability: Magic numbers (1024, 65535, 100, 60) could be extracted as named constants for clarity (`validator.g
- **pkg/engine/performance** (3 issues)
  - [ ] doc: style — LOD constants lack block comment (`types.go:12`)
  - [ ] naming: style — PerformanceConfig stutters with package name; consider renaming to Config (`types.go:108`)
  - [ ] naming: style — PerformanceMonitor stutters with package name; consider renaming to Monitor (`types.go:157`)
- **pkg/integration/choice_consequences** (4 issues)
  - [ ] doc coverage
  - [ ] doc coverage
  - [ ] test coverage
  - [ ] integration
- **pkg/integration/guild_housing** (2 issues)
  - [ ] Serialization
  - [ ] API design
- **pkg/modding** (1 issue)
  - [ ] Deterministic procgen
- **pkg/narrative/branching** (1 issue)
  - [ ] error handling: No structured logging with `logrus.WithFields`; package has no logging at all (acceptable for pure d
- **pkg/network/resilience** (2 issues)
  - [ ] error handling
  - [ ] doc coverage
- **pkg/procgen/book** (4 issues)
  - [ ] Missing genre-specific quest grammar for horror, cyberpunk, post-apocalyptic
  - [ ] Missing genre-specific history grammar for horror, cyberpunk, post-apocalyptic
  - [ ] Missing genre-specific recipe grammar for cyberpunk, post-apocalyptic
  - [ ] Stub implementations for getSeriesName/getVolumeNumber
- **pkg/procgen/building** (2 issues)
  - [ ] doc
  - [ ] integration
- **pkg/procgen/companion** (1 issue)
  - [ ] error handling
- **pkg/procgen/dialog** (4 issues)
  - [ ] doc: `GetGreeting()` method in personality.go does not randomize greetings (returns first greeting), comm
  - [ ] performance: `selectWeightedWord` temperature weighting could use cached power calculations for common temperatur
  - [ ] doc: `hash64` function lacks godoc comment explaining fallback usage (`utils.go:184`)
  - [ ] testing: No benchmark for `GenerateWithPersonality` method (only benchmarks for Generate, GenerateDeterminist
- **pkg/procgen/entity** (4 issues)
  - [ ] **Doc coverage**
  - [ ] **Doc coverage**
  - [ ] **Performance**
  - [ ] **Error handling**
- **pkg/procgen/faction** (3 issues)
  - [ ] ECS compliance
  - [ ] Documentation
  - [ ] Error handling
- **pkg/procgen/furniture** (7 issues)
  - [ ] **Stub/incomplete code**
  - [ ] **Error handling**
  - [ ] **Error handling**
  - [ ] **Doc coverage**
  - [ ] Implement depth-based furniture weighting
  - [ ] Add input parameter validation
  - [ ] Document helper functions
- **pkg/procgen/genre** (1 issue)
  - [ ] doc: coverage — Missing benchmark documentation for blend operations (`blender_test.go:637-649`)
- **pkg/procgen/magic** (2 issues)
  - [ ] documentation: Missing package-level doc.go example import path (`pkg/procgen/magic` should be `github.com/opd-ai/v
  - [ ] documentation: Balance system formulas in doc.go could reference specific functions for easier navigation (`doc.go:
- **pkg/procgen/minigame/games** (4 issues)
  - [ ] error handling: No structured logging with logrus.WithFields; errors returned but not logged for observability (`car
  - [ ] test coverage: `System.Update()` method 0% coverage; documented as no-op but should have explicit test (`system.go:
  - [ ] test coverage: `determineGameStatus()` function 40% coverage; not all game state transitions tested (`memory.go:151
  - [ ] test coverage: `GetRenderOutput()` methods 66.7% coverage across all games; nil LastRender edge case undertested (a
- **pkg/procgen/narrative** (2 issues)
  - [ ] error handling
  - [ ] documentation
- **pkg/procgen/skills** (2 issues)
  - [ ] deprecated: api — Uses deprecated `strings.Title` which should be replaced with `cases.Title(language.English)` 
  - [ ] doc: coverage — README.md exists and is comprehensive; all exported types/functions have godoc; package d
- **pkg/procgen/terrain** (1 issue)
  - [ ] determinism
- **pkg/procgen/vehicle** (1 issue)
  - [ ] documentation: VehicleGenerator struct fields `templates` and `logger` lack individual godoc comments (`generator.g
- **pkg/rendering/parallel** (1 issue)
  - [ ] stub/incomplete code
- **pkg/rendering/particles** (2 issues)
  - [ ] deterministic procgen
  - [ ] doc coverage
- **pkg/rendering/sprites** (1 issue)
  - [ ] Doc coverage
- **pkg/saveload** (3 issues)
  - [ ] doc: `types.go` exported types lack individual godoc comments (all types documented in package doc instea
  - [ ] integration: WASM migration not supported; incompatible saves rejected rather than migrated (documented limitatio
  - [ ] test: `animation_test.go` contains only one test case; could expand coverage of animation state serializat
- **pkg/security** (2 issues)
  - [ ] deterministic procgen
  - [ ] performance
- **pkg/social** (3 issues)
  - [ ] `ChatHistory.GetDelta()` uses a simple heuristic for delta sync rather than true change tracking — documented as a known...
  - [ ] No concurrent Save/Load tests — while thread-safety via `RWMutex` is present, no stress tests validate concurrent persis...
  - [ ] Missing explicit thread-safety guarantees in documentation for public types.
- **pkg/social/persistence** (1 issue)
  - [ ] test: Delta synchronization in `ChatHistory.GetDelta()` uses version-based heuristic rather than true chan
- **pkg/visualtest/parity** (1 issue)
  - [ ] documentation: Platform type methods (String, IsDesktop, IsMobile, IsWeb) lack godoc comments (`platform.go:10,30,3
- **pkg/world/housing** (2 issues)
  - [ ] stub/incomplete
  - [ ] stub/incomplete
- **pkg/world/territory** (1 issue)
  - [ ] error handling

## Issues by Subpackage

### Commands (cmd/)

#### `cmd/client`
- Source: `cmd/client/AUDIT.md`
- Issues: 2 open, 5 resolved (Low: 2)
  - [Low] doc.go references Go 1.24 and Ebiten 2.9 (doc.go)
  - [Low] Large function count in handlers.go
  - ~~[High] Nil pointer dereference in lazy initialization (handlers.go)~~ ✅ Resolved
  - ~~[Medium] Unchecked type assertions on Generator results (util.go)~~ ✅ Resolved
  - ~~[Medium] Unsafe component type assertions (util.go, handlers.go)~~ ✅ Resolved
  - ~~[Medium] Mutex unlock without defer in lazy init (handlers.go:747-749)~~ ✅ Resolved
  - ~~[Medium] WeatherXPBonusSystem receives nil progressionSystem (handlers.go:1186)~~ ✅ Resolved

#### `cmd/mobile`
- Source: `cmd/mobile/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] Nil terrain access
  - [Low] Hardcoded screen dimensions
  - [Low] Silent item generation failures

#### `cmd/mobile/config`
- Source: `cmd/mobile/config/AUDIT.md`
- Issues: 0 — _No issues found_

#### `cmd/server`
- Source: `cmd/server/AUDIT.md`
- Issues: 1 open, 3 resolved (Low: 1)
  - [Low] Global `v9ValidationService` synchronization
  - ~~[High] Nil pointer dereference in `createPlayerEntity`~~ ✅ Resolved
  - ~~[High] Build error in `v9_validation_test.go`~~ ✅ Resolved
  - ~~[Medium] Manual byte encoding in `putFloat64`~~ ✅ Resolved

### Audio (pkg/audio/)

#### `pkg/audio`
- Source: `pkg/audio/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/audio/music`
- Source: `pkg/audio/music/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] determinism
  - [Low] doc coverage

#### `pkg/audio/sfx`
- Source: `pkg/audio/sfx/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] doc coverage

#### `pkg/audio/synthesis`
- Source: `pkg/audio/synthesis/AUDIT.md`
- Issues: 0 — _No issues found_

### Audit (pkg/audit/)

#### `pkg/audit`
- Source: `pkg/audit/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/audit/features`
- Source: `pkg/audit/features/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] Doc coverage
  - [Low] Doc coverage

### Class System (pkg/class/)

#### `pkg/class`
- Source: `pkg/class/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] Silent error handling in `addPrimaryClassStats`, `addSecondaryClassStats`, `addPrestigeClassStats` — errors from `GetCla...
  - [Low] Missing method-level documentation on several exported functions (`GetPrestigeClassDefinition`, `GetAllClasses`, `GetAll...
  - [Low] `RespecCost.MaxCost` cap behavior (10,000g) not documented; respec cost formula not tested at boundary.

#### `pkg/class/advanced`
- Source: `pkg/class/advanced/AUDIT.md`
- Issues: 0 — _No issues found_

### Companion (pkg/companion/)

#### `pkg/companion`
- Source: `pkg/companion/AUDIT.md`
- Issues: 2 open, 2 resolved (Low: 2)
  - [Low] Hardcoded magic numbers for skill decay (0.1), trait clamping (0.0-1.0), LRU limits (1000), and XP values are not config...
  - [Low] `Deserialize()` rebuilds prerequisites as empty and defaults cost to 1 — deserialized skill trees have incomplete struct...
  - ~~[Medium] Race condition in `CompanionLearningSystem.Update()` — accessed `s.manager.companions` map without holding a lock while ...~~ ✅ Resolved
  - ~~[Low] Nil pointer dereference risk in `ProcessCombatAction`, `ProcessSocialInteraction`, `ProcessExploration`, `AdaptBehaviorT...~~ ✅ Resolved

#### `pkg/companion/learning`
- Source: `pkg/companion/learning/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] Documentation

### Configuration (pkg/config/)

#### `pkg/config`
- Source: `pkg/config/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] documentation: Missing benchmark tests for validation performance (`validator_test.go:N/A`)
  - [Low] architecture: Genre list dependency on pkg/procgen/dialog creates coupling; consider extracting genre definitions 
  - [Low] maintainability: Magic numbers (1024, 65535, 100, 60) could be extracted as named constants for clarity (`validator.g

### Engine (pkg/engine/)

#### `pkg/engine/performance`
- Source: `pkg/engine/performance/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] doc: style — LOD constants lack block comment (`types.go:12`)
  - [Low] naming: style — PerformanceConfig stutters with package name; consider renaming to Config (`types.go:108`)
  - [Low] naming: style — PerformanceMonitor stutters with package name; consider renaming to Monitor (`types.go:157`)

#### `pkg/engine/physics/destruction`
- Source: `pkg/engine/physics/destruction/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/engine/physics/fluids`
- Source: `pkg/engine/physics/fluids/AUDIT.md`
- Issues: 2 resolved
  - ~~[Medium] Error handling~~ ✅ Resolved
  - ~~[Low] Doc coverage~~ ✅ Resolved

#### `pkg/engine/physics/vehicle`
- Source: `pkg/engine/physics/vehicle/AUDIT.md`
- Issues: 8 resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[Medium] ECS compliance~~ ✅ Resolved
  - ~~[Medium] ECS compliance~~ ✅ Resolved
  - ~~[Low] Doc coverage~~ ✅ Resolved

#### `pkg/engine/prestige`
- Source: `pkg/engine/prestige/AUDIT.md`
- Issues: 2 resolved
  - ~~[Medium] ECS compliance~~ ✅ Resolved
  - ~~[Low] Deterministic procgen~~ ✅ Resolved

#### `pkg/engine/qol`
- Source: `pkg/engine/qol/AUDIT.md`
- Issues: 0 — _No issues found_

### Errors (pkg/errors/)

#### `pkg/errors`
- Source: `pkg/errors/AUDIT.md`
- Issues: 1 resolved
  - ~~[High] Integration points~~ ✅ Resolved

### Host & Play (pkg/hostplay/)

#### `pkg/hostplay`
- Source: `pkg/hostplay/AUDIT.md`
- Issues: 3 resolved
  - ~~[Medium] — Inconsistent logrus logging in `server_manager.go`: 8 instances used positional args instead of `WithFields`, causing ...~~ ✅ Resolved
  - ~~[Low] — Dead timeout branch in `Shutdown()` (`host_and_play.go`): After calling `s.cancel()`, the `select` on `s.ctx.Done()` f...~~ ✅ Resolved
  - ~~[Low] — Missing blank line before `SerializeSnapshot` comment in `state_broadcaster.go`: The closing brace of `CreateSnapshot`...~~ ✅ Resolved

### Integration (pkg/integration/)

#### `pkg/integration`
- Source: `pkg/integration/AUDIT.md`
- Issues: 8 resolved
  - ~~[High] deterministic procgen~~ ✅ Resolved
  - ~~[High] deterministic procgen~~ ✅ Resolved
  - ~~[High] deterministic procgen~~ ✅ Resolved
  - ~~[High] deterministic procgen~~ ✅ Resolved
  - ~~[High] deterministic procgen~~ ✅ Resolved
  - ~~[High] test coverage~~ ✅ Resolved
  - ~~[Medium] integration points~~ ✅ Resolved
  - ~~[Low] doc coverage~~ ✅ Resolved

#### `pkg/integration/choice_consequences`
- Source: `pkg/integration/choice_consequences/AUDIT.md`
- Issues: 4 open (Low: 4)
  - [Low] doc coverage
  - [Low] doc coverage
  - [Low] test coverage
  - [Low] integration

#### `pkg/integration/companion_housing`
- Source: `pkg/integration/companion_housing/AUDIT.md`
- Issues: 4 resolved
  - ~~[High] error-handling~~ ✅ Resolved
  - ~~[Medium] test-coverage~~ ✅ Resolved
  - ~~[Low] documentation~~ ✅ Resolved
  - ~~[Low] documentation~~ ✅ Resolved

#### `pkg/integration/guild_housing`
- Source: `pkg/integration/guild_housing/AUDIT.md`
- Issues: 2 open, 5 resolved (Low: 2)
  - [Low] Serialization
  - [Low] API design
  - ~~[High] Concurrency~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[Medium] Input validation~~ ✅ Resolved
  - ~~[Medium] Input validation~~ ✅ Resolved
  - ~~[Medium] Input validation~~ ✅ Resolved

#### `pkg/integration/guild_vehicle`
- Source: `pkg/integration/guild_vehicle/AUDIT.md`
- Issues: 4 resolved
  - ~~[Medium] persistence~~ ✅ Resolved
  - ~~[Low] validation~~ ✅ Resolved
  - ~~[Low] testability~~ ✅ Resolved
  - ~~[Low] persistence~~ ✅ Resolved

#### `pkg/integration/housing_crafting`
- Source: `pkg/integration/housing_crafting/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/integration/narrative_world`
- Source: `pkg/integration/narrative_world/AUDIT.md`
- Issues: 4 resolved
  - ~~[High] Stub/incomplete code~~ ✅ Resolved
  - ~~[Low] Stub/incomplete code~~ ✅ Resolved
  - ~~[Low] Test coverage~~ ✅ Resolved
  - ~~[Low] Integration points~~ ✅ Resolved

#### `pkg/integration/political_warfare`
- Source: `pkg/integration/political_warfare/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/integration/trade_routes`
- Source: `pkg/integration/trade_routes/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/integration/world_events`
- Source: `pkg/integration/world_events/AUDIT.md`
- Issues: 0 — _No issues found_

### Memory Profiling (pkg/memprofile/)

#### `pkg/memprofile`
- Source: `pkg/memprofile/AUDIT.md`
- Issues: 1 resolved
  - ~~[Low] documentation: doc.go has duplicate package description comments (`profile.go:1` and `doc.go:1`) — **FIXED**: Remov~~ ✅ Resolved

### Modding (pkg/modding/)

#### `pkg/modding`
- Source: `pkg/modding/AUDIT.md`
- Issues: 1 open, 1 resolved (Low: 1)
  - [Low] Deterministic procgen
  - ~~[Medium] Error handling~~ ✅ Resolved

### Narrative (pkg/narrative/)

#### `pkg/narrative`
- Source: `pkg/narrative/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/narrative/branching`
- Source: `pkg/narrative/branching/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] error handling: No structured logging with `logrus.WithFields`; package has no logging at all (acceptable for pure d

### Network (pkg/network/)

#### `pkg/network/chat`
- Source: `pkg/network/chat/AUDIT.md`
- Issues: 3 resolved
  - ~~[Low] documentation: `ChatSystem` struct fields lack godoc comments (`system.go:31-35`) — **Fixed 2026-02-16**: Added god~~ ✅ Resolved
  - ~~[Low] documentation: `generateMessageID` function is unexported but complex enough to warrant a comment explaining collis~~ ✅ Resolved
  - ~~[Low] error handling: Rate limit error at `system.go:60` hard-codes the limit value "10" instead of using `ChatRateLimit` ~~ ✅ Resolved

#### `pkg/network/federation`
- Source: `pkg/network/federation/AUDIT.md`
- Issues: 7 resolved
  - ~~[High] Test coverage~~ ✅ Resolved
  - ~~[High] Stub/incomplete code~~ ✅ Resolved
  - ~~[High] Stub/incomplete code~~ ✅ Resolved
  - ~~[Medium] Deterministic procgen~~ ✅ Resolved
  - ~~[Medium] Network interfaces~~ ✅ Resolved
  - ~~[Low] Integration points~~ ✅ Resolved
  - ~~[Low] Error handling~~ ✅ Resolved

#### `pkg/network/federation/guild`
- Source: `pkg/network/federation/guild/AUDIT.md`
- Issues: 9 resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[Medium] Deterministic procgen~~ ✅ Resolved
  - ~~[Medium] Integration status~~ ✅ Resolved
  - ~~[Medium] Integration status~~ ✅ Resolved
  - ~~[Low] Error handling~~ ✅ Resolved
  - ~~[Low] Documentation~~ ✅ Resolved
  - ~~[Low] Thread safety~~ ✅ Resolved
  - ~~[Low] TimeProvider abstraction~~ ✅ Resolved

#### `pkg/network/federation/mobile`
- Source: `pkg/network/federation/mobile/AUDIT.md`
- Issues: 7 resolved
  - ~~[High] Stub/incomplete code~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[Medium] Error handling~~ ✅ Resolved
  - ~~[Low] Integration points~~ ✅ Resolved

#### `pkg/network/federation/webrtc`
- Source: `pkg/network/federation/webrtc/AUDIT.md`
- Issues: 8 resolved
  - ~~[High] Integration~~ ✅ Resolved
  - ~~[High] Non-deterministic timestamps~~ ✅ Resolved
  - ~~[Medium] Stub implementation~~ ✅ Resolved
  - ~~[Medium] No structured logging~~ ✅ Resolved
  - ~~[Low] Health check race~~ ✅ Resolved
  - ~~[Low] Round-robin counter overflow~~ ✅ Resolved
  - ~~[Low] Missing context propagation~~ ✅ Resolved
  - ~~[Low] Unbounded channel~~ ✅ Resolved

#### `pkg/network/resilience`
- Source: `pkg/network/resilience/AUDIT.md`
- Issues: 3 open (Medium: 1, Low: 2)
  - [Medium] deterministic procgen
  - [Low] error handling
  - [Low] doc coverage

#### `pkg/network/trade`
- Source: `pkg/network/trade/AUDIT.md`
- Issues: 9 resolved
  - ~~[High] architecture~~ ✅ Resolved
  - ~~[High] deterministic-procgen~~ ✅ Resolved
  - ~~[Medium] test-coverage~~ ✅ Resolved
  - ~~[Medium] error-handling~~ ✅ Resolved
  - ~~[Medium] integration~~ ✅ Resolved
  - ~~[Low] test-coverage~~ ✅ Resolved
  - ~~[Low] test-coverage~~ ✅ Resolved
  - ~~[Low] test-coverage~~ ✅ Resolved
  - ~~[Low] doc-coverage~~ ✅ Resolved

### Procedural Generation (pkg/procgen/)

#### `pkg/procgen`
- Source: `pkg/procgen/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/audit`
- Source: `pkg/procgen/audit/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/book`
- Source: `pkg/procgen/book/AUDIT.md`
- Issues: 4 open, 1 resolved (Low: 4)
  - [Low] Missing genre-specific quest grammar for horror, cyberpunk, post-apocalyptic
  - [Low] Missing genre-specific history grammar for horror, cyberpunk, post-apocalyptic
  - [Low] Missing genre-specific recipe grammar for cyberpunk, post-apocalyptic
  - [Low] Stub implementations for getSeriesName/getVolumeNumber
  - ~~[Medium] No recursion depth limit in Grammar.Expand~~ ✅ Resolved

#### `pkg/procgen/building`
- Source: `pkg/procgen/building/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] doc
  - [Low] integration

#### `pkg/procgen/class`
- Source: `pkg/procgen/class/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/companion`
- Source: `pkg/procgen/companion/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] error handling

#### `pkg/procgen/dialog`
- Source: `pkg/procgen/dialog/AUDIT.md`
- Issues: 4 open (Low: 4)
  - [Low] doc: `GetGreeting()` method in personality.go does not randomize greetings (returns first greeting), comm
  - [Low] performance: `selectWeightedWord` temperature weighting could use cached power calculations for common temperatur
  - [Low] doc: `hash64` function lacks godoc comment explaining fallback usage (`utils.go:184`)
  - [Low] testing: No benchmark for `GenerateWithPersonality` method (only benchmarks for Generate, GenerateDeterminist

#### `pkg/procgen/entity`
- Source: `pkg/procgen/entity/AUDIT.md`
- Issues: 4 open (Low: 4)
  - [Low] **Doc coverage**
  - [Low] **Doc coverage**
  - [Low] **Performance**
  - [Low] **Error handling**

#### `pkg/procgen/environment`
- Source: `pkg/procgen/environment/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/faction`
- Source: `pkg/procgen/faction/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] ECS compliance
  - [Low] Documentation
  - [Low] Error handling

#### `pkg/procgen/furniture`
- Source: `pkg/procgen/furniture/AUDIT.md`
- Issues: 7 open (Low: 7)
  - [Low] **Stub/incomplete code**
  - [Low] **Error handling**
  - [Low] **Error handling**
  - [Low] **Doc coverage**
  - [Low] Implement depth-based furniture weighting
  - [Low] Add input parameter validation
  - [Low] Document helper functions

#### `pkg/procgen/genre`
- Source: `pkg/procgen/genre/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] doc: coverage — Missing benchmark documentation for blend operations (`blender_test.go:637-649`)

#### `pkg/procgen/legendary`
- Source: `pkg/procgen/legendary/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/magic`
- Source: `pkg/procgen/magic/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] documentation: Missing package-level doc.go example import path (`pkg/procgen/magic` should be `github.com/opd-ai/v
  - [Low] documentation: Balance system formulas in doc.go could reference specific functions for easier navigation (`doc.go:

#### `pkg/procgen/minigame`
- Source: `pkg/procgen/minigame/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/minigame/games`
- Source: `pkg/procgen/minigame/games/AUDIT.md`
- Issues: 4 open (Low: 4)
  - [Low] error handling: No structured logging with logrus.WithFields; errors returned but not logged for observability (`car
  - [Low] test coverage: `System.Update()` method 0% coverage; documented as no-op but should have explicit test (`system.go:
  - [Low] test coverage: `determineGameStatus()` function 40% coverage; not all game state transitions tested (`memory.go:151
  - [Low] test coverage: `GetRenderOutput()` methods 66.7% coverage across all games; nil LastRender edge case undertested (a

#### `pkg/procgen/narrative`
- Source: `pkg/procgen/narrative/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] error handling
  - [Low] documentation

#### `pkg/procgen/puzzle`
- Source: `pkg/procgen/puzzle/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/quest`
- Source: `pkg/procgen/quest/AUDIT.md`
- Issues: 5 resolved
  - ~~[High] ECS compliance~~ ✅ Resolved
  - ~~[High] Integration~~ ✅ Resolved
  - ~~[Medium] Genre support~~ ✅ Resolved
  - ~~[Low] Documentation~~ ✅ Resolved
  - ~~[Low] Validation~~ ✅ Resolved

#### `pkg/procgen/recipe`
- Source: `pkg/procgen/recipe/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/skills`
- Source: `pkg/procgen/skills/AUDIT.md`
- Issues: 3 open (Medium: 1, Low: 2)
  - [Medium] ECS: compliance — `Skill` and `SkillTree` types have business logic methods (IsUnlocked, CanLevelUp, Tota
  - [Low] deprecated: api — Uses deprecated `strings.Title` which should be replaced with `cases.Title(language.English)` 
  - [Low] doc: coverage — README.md exists and is comprehensive; all exported types/functions have godoc; package d

#### `pkg/procgen/station`
- Source: `pkg/procgen/station/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/procgen/story`
- Source: `pkg/procgen/story/AUDIT.md`
- Issues: 1 open (Medium: 1)
  - [Medium] ECS compliance: `StoryJournalComponent` in `pkg/engine/story_fragment_component.go` has behavior methods (`AddDiscov

#### `pkg/procgen/terrain`
- Source: `pkg/procgen/terrain/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] determinism

#### `pkg/procgen/vehicle`
- Source: `pkg/procgen/vehicle/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] documentation: VehicleGenerator struct fields `templates` and `logger` lack individual godoc comments (`generator.g

### Rendering (pkg/rendering/)

#### `pkg/rendering`
- Source: `pkg/rendering/AUDIT.md`
- Issues: 4 resolved
  - ~~[High] Interface orphaning~~ ✅ Resolved
  - ~~[High] Architecture inconsistency~~ ✅ Resolved
  - ~~[Medium] No integration points~~ ✅ Resolved
  - ~~[Low] Test coverage misleading~~ ✅ Resolved

#### `pkg/rendering/animation`
- Source: `pkg/rendering/animation/AUDIT.md`
- Issues: 3 resolved
  - ~~[Medium] ECS Integration~~ ✅ Resolved
  - ~~[Low] Performance Monitoring~~ ✅ Resolved
  - ~~[Low] Test Environment~~ ✅ Resolved

#### `pkg/rendering/cache`
- Source: `pkg/rendering/cache/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/display`
- Source: `pkg/rendering/display/AUDIT.md`
- Issues: 4 resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[High] Deterministic procgen~~ ✅ Resolved
  - ~~[Medium] Integration points~~ ✅ Resolved
  - ~~[Low] Test coverage~~ ✅ Resolved

#### `pkg/rendering/lighting`
- Source: `pkg/rendering/lighting/AUDIT.md`
- Issues: 3 resolved
  - ~~[Low] **Doc coverage**~~ ✅ Resolved
  - ~~[Low] **Incomplete feature**~~ ✅ Resolved
  - ~~[Low] **Test isolation**~~ ✅ Resolved

#### `pkg/rendering/palette`
- Source: `pkg/rendering/palette/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/parallel`
- Source: `pkg/rendering/parallel/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] stub/incomplete code

#### `pkg/rendering/particles`
- Source: `pkg/rendering/particles/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] deterministic procgen
  - [Low] doc coverage

#### `pkg/rendering/patterns`
- Source: `pkg/rendering/patterns/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/pool`
- Source: `pkg/rendering/pool/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/postprocess`
- Source: `pkg/rendering/postprocess/AUDIT.md`
- Issues: 3 resolved
  - ~~[Medium] **Division by zero in motion blur**~~ ✅ Resolved
  - ~~[Low] **GPU tests crash headless environments**~~ ✅ Resolved
  - ~~[Low] **Missing config validation**~~ ✅ Resolved

#### `pkg/rendering/quality`
- Source: `pkg/rendering/quality/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/shapes`
- Source: `pkg/rendering/shapes/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/sprites`
- Source: `pkg/rendering/sprites/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] Doc coverage

#### `pkg/rendering/tiles`
- Source: `pkg/rendering/tiles/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/rendering/ui`
- Source: `pkg/rendering/ui/AUDIT.md`
- Issues: 0 — _No issues found_

### Save/Load (pkg/saveload/)

#### `pkg/saveload`
- Source: `pkg/saveload/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] doc: `types.go` exported types lack individual godoc comments (all types documented in package doc instea
  - [Low] integration: WASM migration not supported; incompatible saves rejected rather than migrated (documented limitatio
  - [Low] test: `animation_test.go` contains only one test case; could expand coverage of animation state serializat

### Security (pkg/security/)

#### `pkg/security`
- Source: `pkg/security/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] deterministic procgen
  - [Low] performance

### Social (pkg/social/)

#### `pkg/social`
- Source: `pkg/social/AUDIT.md`
- Issues: 3 open (Low: 3)
  - [Low] `ChatHistory.GetDelta()` uses a simple heuristic for delta sync rather than true change tracking — documented as a known...
  - [Low] No concurrent Save/Load tests — while thread-safety via `RWMutex` is present, no stress tests validate concurrent persis...
  - [Low] Missing explicit thread-safety guarantees in documentation for public types.

#### `pkg/social/persistence`
- Source: `pkg/social/persistence/AUDIT.md`
- Issues: 1 open, 2 resolved (Low: 1)
  - [Low] test: Delta synchronization in `ChatHistory.GetDelta()` uses version-based heuristic rather than true chan
  - ~~[Low] doc: `TimeProvider` abstraction added to all managers; `ChatHistory`, `TrustManager`, and `ReputationMana~~ ✅ Resolved
  - ~~[Medium] doc: `types.go` TimeProvider documentation updated to explain intentional time.Now() usage for server-sid~~ ✅ Resolved

### UX (pkg/ux/)

#### `pkg/ux`
- Source: `pkg/ux/AUDIT.md`
- Issues: 1 resolved
  - ~~[Low] Redundant code~~ ✅ Resolved

### Visual Testing (pkg/visualtest/)

#### `pkg/visualtest`
- Source: `pkg/visualtest/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/visualtest/parity`
- Source: `pkg/visualtest/parity/AUDIT.md`
- Issues: 1 open (Low: 1)
  - [Low] documentation: Platform type methods (String, IsDesktop, IsMobile, IsWeb) lack godoc comments (`platform.go:10,30,3

### World (pkg/world/)

#### `pkg/world`
- Source: `pkg/world/AUDIT.md`
- Issues: 0 — _No issues found_

#### `pkg/world/economy`
- Source: `pkg/world/economy/AUDIT.md`
- Issues: 3 resolved
  - ~~[High] deterministic-procgen~~ ✅ Resolved
  - ~~[Medium] error-handling~~ ✅ Resolved
  - ~~[Low] logging~~ ✅ Resolved

#### `pkg/world/housing`
- Source: `pkg/world/housing/AUDIT.md`
- Issues: 2 open (Low: 2)
  - [Low] stub/incomplete
  - [Low] stub/incomplete

#### `pkg/world/raids`
- Source: `pkg/world/raids/AUDIT.md`
- Issues: 3 resolved
  - ~~[Medium] Deterministic procgen~~ ✅ Resolved
  - ~~[Low] Error handling~~ ✅ Resolved
  - ~~[Low] Doc coverage~~ ✅ Resolved

#### `pkg/world/territory`
- Source: `pkg/world/territory/AUDIT.md`
- Issues: 1 open, 4 resolved (Low: 1)
  - [Low] error handling
  - ~~[High] error handling~~ ✅ Resolved
  - ~~[High] data race~~ ✅ Resolved
  - ~~[High] data race~~ ✅ Resolved
  - ~~[Medium] deterministic procgen~~ ✅ Resolved

## Cross-Package Dependencies

- Issues in **pkg/config** should be resolved before **cmd/client** — Configuration validation fixes in pkg/config must land before client consumption
- Issues in **pkg/config** should be resolved before **cmd/server** — Configuration validation fixes in pkg/config must land before server consumption
- Issues in **pkg/config** should be resolved before **cmd/mobile** — Configuration validation fixes in pkg/config must land before mobile consumption
- Issues in **pkg/saveload** should be resolved before **pkg/social/persistence** — Save/load system improvements enable better social data persistence

### Dependency Graph

```
  pkg/config --> cmd/client
  pkg/config --> cmd/mobile
  pkg/config --> cmd/server
  pkg/saveload --> pkg/social/persistence
```

## Recommendations

### Strategic Approach

1. **Resolve Critical Issues First**: Address all critical-severity issues before moving to high-priority items. These represent potential security vulnerabilities, data loss risks, or system stability concerns.

2. **Fix Core Packages Before Dependents**: Resolve issues in foundational packages (`pkg/errors`, `pkg/config`, `pkg/engine`) before their dependents to avoid rework and ensure consistent patterns propagate downstream.

3. **Batch by Domain**: Group related fixes within the same domain (e.g., all `pkg/procgen/*` issues together) to maintain context and reduce review overhead.

4. **Prioritize ECS Compliance**: Several packages have components with behavior methods that violate the pure-data ECS pattern. Refactoring these into dedicated systems improves architecture consistency and testability.

5. **Add Structured Logging**: Multiple packages lack `logrus.WithFields` structured logging. Adding consistent logging across all packages improves operational visibility and debugging capability.

6. **Enhance Test Coverage**: While average coverage is strong (82.4%), some packages fall below the 65% target. Prioritize test additions for packages with high-severity open issues.

### Domain-Specific Guidance

- **Commands (cmd/)**: 6 open issues (Low: 6)
- **Audio (pkg/audio/)**: 3 open issues (Low: 3)
- **Audit (pkg/audit/)**: 2 open issues (Low: 2)
- **Class System (pkg/class/)**: 3 open issues (Low: 3)
- **Companion (pkg/companion/)**: 3 open issues (Low: 3)
- **Configuration (pkg/config/)**: 3 open issues (Low: 3)
- **Engine (pkg/engine/)**: 3 open issues (Low: 3)
- **Integration (pkg/integration/)**: 6 open issues (Low: 6)
- **Modding (pkg/modding/)**: 1 open issues (Low: 1)
- **Narrative (pkg/narrative/)**: 1 open issues (Low: 1)
- **Network (pkg/network/)**: 3 open issues (Medium: 1, Low: 2)
- **Procedural Generation (pkg/procgen/)**: 40 open issues (Medium: 2, Low: 38)
- **Rendering (pkg/rendering/)**: 4 open issues (Low: 4)
- **Save/Load (pkg/saveload/)**: 3 open issues (Low: 3)
- **Security (pkg/security/)**: 2 open issues (Low: 2)
- **Social (pkg/social/)**: 4 open issues (Low: 4)
- **Visual Testing (pkg/visualtest/)**: 1 open issues (Low: 1)
- **World (pkg/world/)**: 3 open issues (Low: 3)

### Estimated Effort

- **Phase 1 (Critical)**: 0 issues — Immediate attention required
- **Phase 2 (High)**: 0 issues — Target for next sprint
- **Phase 3 (Medium)**: 3 issues — Ongoing improvement backlog
- **Phase 4 (Low)**: 88 issues — Address opportunistically

## Resolved Issues Summary

132 issues have been resolved across 33 packages.

| Severity | Resolved |
|----------|----------|
| High | 43 |
| Medium | 39 |
| Low | 50 |
| **Total** | **132** |

---

*This report was generated by consolidating all subpackage AUDIT.md files in the repository. Each issue links back to its source audit file for detailed context and resolution guidance.*