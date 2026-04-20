# Implementation Gaps — 2026-05-02

This document tracks open implementation gaps between stated project goals and actual runtime
behavior. Resolved gaps are retained for history; new gaps are appended.

---

## Gap 1: Voice Chat Network Transport Not Wired

- **Intended Behavior**: README claims "Voice chat is integrated with party, guild, proximity, and
  private channels using a built-in codec with spatial audio support."
- **Current State**:
  - ✅ ADPCM voice codec — `pkg/audio/voice.go`
  - ✅ `TCPVoiceTransport` implementing `audio.VoiceTransport` — `pkg/network/voice_transport.go:94`
  - ✅ Jitter buffer with heap-ordered packet reordering — `pkg/network/voice_transport.go:57`
  - ✅ `VoiceChannelSystem` (moderation, 4 channel types) — `pkg/engine/voice_channel_system.go`
  - ✅ `SpatialVoiceSystem` (distance-based volume/pan) — `pkg/engine/spatial_voice_system.go`
  - ✅ `VoiceAudioSystem`, `VoiceChannelSystem` wired in `cmd/client/handlers.go:2215`
  - ❌ `InitializeVoice()` never called from `cmd/client/` or `cmd/server/`
  - ❌ `NewTCPVoiceTransport` has zero non-test, non-self-file callers
  - ❌ `VoiceSettingsSystem` not instantiated (see Gap 14)
  - ❌ Server-side voice routing (forward packets to channel members) not implemented
- **Blocked Goal**: Multiplayer voice chat is completely non-functional. Players cannot hear each other.
- **Implementation Path**:
  1. In `cmd/client/handlers.go`, after network client connects, create transport:
     `transport := network.NewTCPVoiceTransport(network.DefaultVoiceTransportConfig(), playerID, sendFunc)`
  2. Call `sys.audioManager.InitializeVoice(audio.VoiceQualityMedium, transport)` 
  3. Add server-side handler in `pkg/network/server.go` to decode VoicePacket type and broadcast to channel members
  4. Instantiate `VoiceSettingsSystem` in `cmd/client/handlers.go` alongside `VoiceAudioSystem`
  5. Add integration tests: `TestVoiceEndToEnd`, `TestSpatialVoiceRouting`
- **Dependencies**: Network client connection must be established before voice initialization
- **Effort**: medium (2–3 days)

---

## Gap 2: Quest Generator Post-Apocalyptic Templates

- **Status**: ✅ RESOLVED — `pkg/procgen/quest/types.go` has full `postapoc` case with all template functions.

---

## Gap 3: README System Count Mismatch

- **Status**: ✅ RESOLVED — README correctly states 66 game systems; `system_init.go` logs this value.

---

## Gap 4: Territory Bonuses Not Displayed in HUD

- **Status**: ✅ RESOLVED — `drawTerritoryBonuses()` in `pkg/engine/hud_system.go:389`; wired at `cmd/client/handlers.go:2194`.

---

## Gap 5: FPS Benchmark Scope Too Narrow

- **Intended Behavior**: "60 FPS minimum on mid-range hardware" validated by CI.
- **Current State**:
  - `pkg/benchmark/fps/benchmark_test.go` tests `MovementSystem` with 2,000 entities only
  - CI gate using `scripts/benchmark-regression.sh` validates this single benchmark
  - Collision, rendering, and 64 other systems are unrepresented
- **Blocked Goal**: CI cannot detect FPS regressions introduced by any system other than movement.
- **Implementation Path**:
  1. Create `BenchmarkFullSystemSuite` calling `InitializeGameSystems()` with all 66 systems active
  2. Add `BenchmarkCollisionSystem` with realistic entity density (100–500 entities)
  3. Add `BenchmarkRenderPipeline` (requires xvfb or headless guard)
  4. Update `scripts/benchmark-baseline.json` with new baseline entries
- **Dependencies**: Requires xvfb in CI for rendering benchmarks
- **Effort**: small (1–2 days)

---

## Gap 6: Signal Handler Integration Test Missing

- **Status**: ✅ RESOLVED — `cmd/server/shutdown_test.go` contains five comprehensive tests including SIGINT/SIGTERM.

---

## Gap 7: Territory System Lacks Mod Support

- **Status**: ✅ RESOLVED — `TerritoryConfig` struct; `Manager.SetConfig()` allows runtime override.

---

## Gap 8: Trade Validation Per-Item Quantity

- **Status**: ✅ RESOLVED (Design Decision) — Items are unique instances, not stackable commodities.

---

## Gap 9: memprofile Uses fmt.Printf

- **Status**: ✅ RESOLVED — Exempt with documentation comment; `ExportJSON()` available for structured output.

---

## Gap 10: No Automated CI Performance Benchmark Gates

- **Status**: ✅ RESOLVED — `scripts/benchmark-regression.sh` and `scripts/benchmark-memory.sh` integrated into `.github/workflows/test.yml`.

---

## Gap 11: ECS Race Conditions on Entity Staging Buffers

- **Status**: ✅ RESOLVED — `entityMu sync.Mutex` added to `pkg/engine/ecs.go`; `TestCreateEntityConcurrentSafety` passes under `-race`.

---

## Gap 12: Unprotected Global Talent Registry Maps

- **Status**: ✅ RESOLVED — `talentMu sync.RWMutex` added to `pkg/engine/talent_definitions.go`.

---

## Gap 13: ChatUI HandleClick Double-Activation

- **Status**: ✅ RESOLVED — Second `if` changed to `else if` in `pkg/rendering/ui/chat.go`.

---

## Gap 14: XP Never Awarded on Enemy Death

- **Intended Behavior**: Players earn experience points for killing enemies and progress through levels.
- **Current State**:
  - ✅ `ProgressionSystem.AwardXP()` fully implemented at `pkg/engine/progression_system.go:97`
  - ✅ `FactionXPBonusSystem` calls `AwardXP` for faction bonuses at `pkg/engine/faction_xp_bonus_system.go:81`
  - ❌ `createDeathCallback` in `cmd/client/util.go:1444` does NOT call `AwardXP`
  - ❌ `progressionSystem` is not a parameter of `createDeathCallback` (5 parameters; none is progression)
  - ❌ `system_init.go:971` `SetKillCallback` awards weapon mastery and class affinity XP but not base combat XP
- **Blocked Goal**: Players cannot level up from combat. Character progression is broken.
- **Implementation Path**:
  1. Add `progressionSystem *engine.ProgressionSystem` as a parameter to `createDeathCallback` signature at `cmd/client/util.go:1430`
  2. After the entity is marked dead (after `enemy.AddComponent(deadComp)` at line ~1460), call:
     `if progressionSystem != nil && playerEntity != nil { _ = progressionSystem.AwardXP(*playerEntity, calculateXP(enemy)) }`
  3. Implement `calculateXP(enemy *engine.Entity) int` using enemy stats or a `BaseXPReward` component
  4. Thread `sys.progressionSystem` into `configureDeathCallback` call at `cmd/client/handlers.go:3482`
- **Dependencies**: None; `AwardXP` is already thread-safe
- **Effort**: small (< 1 day)

---

## Gap 15: GPU Memory Leak in LoadingUI

- **Intended Behavior**: Loading screen renders without accumulating GPU memory.
- **Current State**:
  - `pkg/engine/loading_ui.go:87` — `drawRect()` calls `ebiten.NewImage(width, height)` and immediately discards the reference after `screen.DrawImage`
  - Called 3× per frame (lines 75, 76, 81) for the full duration of async terrain generation (5–30 s)
  - No `defer img.Dispose()` call; Ebiten GPU images are not GC'd by Go's runtime without explicit disposal
- **Blocked Goal**: "<500 MB client memory target" and "Mobile support" (mobile OOM crashes before gameplay).
- **Implementation Path**:
  Option A (preferred — zero per-frame allocation): Add three `*ebiten.Image` fields to `LoadingUI` (`bgRect`, `borderRect`, `fillRect`). Initialize lazily in `Draw()` if nil. Reuse each frame, calling `img.Fill(c)` before draw.
  Option B (minimal change): Add `defer img.Dispose()` at line 88 in `drawRect()`. Correct but allocates per frame.
- **Dependencies**: None
- **Effort**: small (< 1 day)

---

## Gap 16: Mobile CraftingSystem Initialized with nil itemGenerator

- **Intended Behavior**: Crafting works on all platforms including iOS/Android.
- **Current State**:
  - `pkg/engine/system_init.go:1168` — `NewCraftingSystem(game.World, inventorySystem, nil)` with comment `// itemGen set later`
  - No exported setter exists on `CraftingSystem` for `itemGenerator` (only `SetStationManager` at line 75)
  - `pkg/engine/crafting_system.go:579` dereferences `s.itemGenerator.Generate(...)` without a nil guard
  - `cmd/mobile/mobile.go:449` creates `itemGen := item.NewItemGenerator()` but never assigns it to the CraftingSystem
- **Blocked Goal**: "Mobile support" — crafting panics on iOS/Android.
- **Implementation Path**:
  1. Add to `pkg/engine/crafting_system.go`: `func (s *CraftingSystem) SetItemGenerator(gen *item.ItemGenerator) { s.itemGenerator = gen }`
  2. Add nil guard in `CraftingSystem.Craft()` at line 579: `if s.itemGenerator == nil { return nil, errors.New("item generator not configured") }`
  3. In `cmd/mobile/mobile.go` after `InitializeGameSystems()` returns, call: `systemsInitResult.CraftingSystem.SetItemGenerator(item.NewItemGenerator())`
- **Dependencies**: None
- **Effort**: small (< 1 day)

---

## Gap 17: ScriptingSystem Not Instantiated — Mod Scripts Cannot Execute

- **Intended Behavior**: JSON mods that attach `ScriptingComponent` to entities can run sandboxed scripts via the ECS update loop.
- **Current State**:
  - `pkg/engine/scripting_system.go:52` — Full implementation with builtins, execution limits, statistics
  - Zero callers of `NewScriptingSystem` outside the definition file
  - `pkg/modding/` manager loads and validates mods but has no path to execute entity scripts
- **Blocked Goal**: "Sandboxed JSON-based modding system" — entity scripts silently do nothing.
- **Implementation Path**:
  1. Add `scriptingSystem := NewScriptingSystem(game.World); game.World.AddSystem(scriptingSystem)` to `pkg/engine/system_init.go` in the modding/scripting section
  2. Expose via `SystemInitResult.ScriptingSystem *ScriptingSystem`
  3. In `cmd/client/handlers.go`, after mod loading, call `scriptingSystem.RegisterBuiltin(...)` for any game-specific builtins
- **Dependencies**: Modding manager must be initialized first
- **Effort**: small (< 1 day)

---

## Gap 18: WorldPersistenceSystem Not Instantiated

- **Intended Behavior**: NPC state, city state, world events, and player reputation persist across sessions without full world regeneration.
- **Current State**:
  - `pkg/engine/world_persistence_system.go:22` — Complete implementation with `SaveWorldState`, `LoadWorldState`, city/NPC state serialization
  - Zero callers of `NewWorldPersistenceSystem` outside the definition file
  - Save/load infrastructure in `pkg/saveload/` saves player state but not world state
- **Blocked Goal**: Persistent world state — NPCs and cities reset to generated state on every session.
- **Implementation Path**:
  1. Instantiate in `pkg/engine/system_init.go` and expose via `SystemInitResult`
  2. In `cmd/server/main.go`, call `persistenceSystem.LoadWorldState(ctx, worldID)` after terrain generation
  3. On graceful shutdown (SIGTERM handler), call `persistenceSystem.SaveWorldState(ctx, worldID)`
- **Dependencies**: Requires `WorldMemoryComponent` to be attached to relevant entities during generation
- **Effort**: medium (2–3 days)

---

## Gap 19: HousingUI Missing from shouldUpdateWorld

- **Intended Behavior**: World simulation pauses (enemies freeze, time stops) while the player interacts with the housing UI, consistent with all other gameplay UIs.
- **Current State**:
  - `pkg/engine/game.go:1358` — `shouldUpdateWorld()` checks 14 UIs; `HousingUI` is absent
  - `pkg/engine/game.go:1323` — `updateVirtualControlsVisibility()`'s `anyUIOpen` also omits `HousingUI`
  - `pkg/world/housing/ui.go:101` — `HousingUI.IsVisible() bool` exists
  - `HousingUI` is stored as `g.HousingUI` (game.go:77) and updated/drawn each frame
- **Blocked Goal**: Player safety during housing management; consistent pause behavior across all UIs.
- **Implementation Path**:
  1. In `shouldUpdateWorld()` (game.go:1358), add: `(g.HousingUI == nil || !g.HousingUI.IsVisible()) &&`
  2. In `updateVirtualControlsVisibility()` (game.go:1323), add `(g.HousingUI != nil && g.HousingUI.IsVisible())` to the `anyUIOpen` expression
- **Dependencies**: None
- **Effort**: trivial (< 30 minutes)

---

## Gap 20: Chunk World Systems Never Instantiated

- **Intended Behavior**: Large worlds are streamed in chunks; chunk state is compressed and persisted; player modifications to the world are tracked by chunk.
- **Current State**:
  - `pkg/world/chunk_loader.go:33` — `NewChunkLoaderSystem` — zero non-world-pkg callers
  - `pkg/world/chunk_compression.go:14` — `NewChunkCompressionSystem` — zero non-world-pkg callers
  - `pkg/world/chunk_modification.go:15` — `NewChunkModificationSystem` — zero non-world-pkg callers
  - Only referenced in `pkg/world/doc.go` code examples
- **Blocked Goal**: Persistent world; memory-efficient large worlds. Without chunking, entire terrain is held in RAM.
- **Implementation Path**:
  1. Fix Gap 18 (WorldPersistenceSystem) first to provide the `WorldPersistence` dependency
  2. In `cmd/server/main.go` world initialization: `loader := world.NewChunkLoaderSystem(seed, persistence, generator)`; register with the game loop
  3. Add `NewChunkCompressionSystem()` to compress chunks on eviction
  4. Add `NewChunkModificationSystem(state)` to track and dirty-flag modified chunks
- **Dependencies**: Gap 18 (WorldPersistenceSystem), `ChunkGenerator` interface implementation
- **Effort**: large (1–2 weeks — requires chunk boundary handling and streaming integration)

---

## Summary

| Gap | Severity | Effort | Status |
|-----|----------|--------|--------|
| Gap 1: Voice network transport | 🔴 CRITICAL | medium | Open |
| Gap 2: Quest postapoc templates | 🟢 LOW | N/A | ✅ Resolved |
| Gap 3: README system count | 🟢 LOW | N/A | ✅ Resolved |
| Gap 4: Territory HUD display | 🟢 LOW | N/A | ✅ Resolved |
| Gap 5: FPS benchmark scope | 🟡 MEDIUM | small | Open |
| Gap 6: Signal handler test | 🟢 LOW | N/A | ✅ Resolved |
| Gap 7: Territory mod support | 🟢 LOW | N/A | ✅ Resolved |
| Gap 8: Trade quantity model | 🟢 LOW | N/A | ✅ Resolved |
| Gap 9: memprofile logging | 🟢 LOW | N/A | ✅ Resolved |
| Gap 10: Performance CI gate | 🟢 LOW | N/A | ✅ Resolved |
| Gap 11: ECS entity staging races | 🔴 CRITICAL | N/A | ✅ Resolved |
| Gap 12: Talent registry unprotected | 🟡 MEDIUM | N/A | ✅ Resolved |
| Gap 13: ChatUI double-activation | 🟡 MEDIUM | N/A | ✅ Resolved |
| Gap 14: XP never awarded on kill | 🔴 CRITICAL | small | Open |
| Gap 15: GPU memory leak (LoadingUI) | 🔴 CRITICAL | small | Open |
| Gap 16: Mobile nil itemGenerator | 🔴 CRITICAL | small | Open |
| Gap 17: ScriptingSystem not wired | 🔴 CRITICAL | small | Open |
| Gap 18: WorldPersistenceSystem not wired | 🟠 HIGH | medium | Open |
| Gap 19: HousingUI missing from pause | 🟠 HIGH | trivial | Open |
| Gap 20: Chunk systems not wired | 🟠 HIGH | large | Open |

---

*Updated: 2026-05-02 | 10 of 20 gaps resolved*
