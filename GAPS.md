# Implementation Gaps — 2026-04-19

This document tracks gaps between the project's stated goals (README.md) and actual implementation.

---

## Gap 1: Voice Chat Network Transport Not Implemented

- **Stated Goal**: README claims "Voice chat is integrated with party, guild, proximity, and private channels using a built-in codec with spatial audio support."
- **Current State**: 
  - ✅ Voice codec exists (`pkg/audio/voice.go` — ADPCM, 3 quality presets)
  - ✅ Channel management exists (`pkg/engine/voice_channel_system.go` — 4 channel types with moderation)
  - ✅ Spatial audio exists (`pkg/engine/spatial_voice_system.go` — distance-based volume/panning)
  - ❌ `VoiceTransport` interface has no TCP/WebRTC implementation
  - ❌ Voice packets are never sent over the network
  - ❌ No jitter buffer or packet loss concealment
- **Impact**: Voice chat is **non-functional in multiplayer**. Players cannot hear each other. The feature is advertised but does not work over the network.
- **Closing the Gap**:
  1. Implement `VoiceTransport` in `pkg/network/client.go` that serializes voice frames with sequence numbers and channel IDs
  2. Add voice packet routing in `pkg/network/server.go` that forwards to channel members
  3. Implement jitter buffer (50-100ms) in `VoiceProcessor.ProcessOutput()`
  4. Wire `VoiceProcessor` to network client in `cmd/client/main.go`
  5. Add integration tests: `TestVoiceEndToEnd`, `TestSpatialVoiceRouting`
  - **Estimated effort**: 2-3 days
  - **Validation**: `go test -v ./pkg/network/... -run Voice && go test -v ./pkg/engine/... -run Voice`

---

## Gap 2: Quest Generator Missing Post-Apocalyptic Genre Support

- **Status**: ✅ RESOLVED
- **Stated Goal**: README claims "genre system supporting fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic themes" with procedural quest generation.
- **Current State**:
  - ✅ `pkg/procgen/quest/generator.go:selectTemplates()` has `case "postapoc":` (lines 107-111)
  - ✅ `GetPostApocKillTemplates()` — implemented in types.go:645
  - ✅ `GetPostApocCollectTemplates()` — implemented in types.go:667
  - ✅ `GetPostApocBossTemplates()` — implemented in types.go:689
  - ✅ `GetPostApocExploreTemplates()` — implemented in types.go:712
- **Resolution**: Already fully implemented with genre-appropriate templates (Raiders, Mutants, Wasteland themes).

---

## Gap 3: README Claims "100+ Systems" but Actual Count is 66

- **Status**: ✅ RESOLVED
- **Stated Goal**: README states "100+ game systems" in project description.
- **Current State**:
  - ✅ README.md line 33 already states: "ECS core, 66 game systems"
  - `pkg/engine/system_init.go` explicitly logs: `"initializing game systems (66 total)"`
  - Documentation is accurate
- **Resolution**: Already fixed. README correctly states "66 game systems".

---

## Gap 4: Territory Bonuses Not Displayed in HUD

- **Status**: ✅ RESOLVED
- **Stated Goal**: Territory control provides gameplay bonuses.
- **Implementation**:
  - ✅ `TerritoryBonusProvider` interface defined in `pkg/engine/hud_system.go:14-19`
  - ✅ `drawTerritoryBonuses()` renders bonus panel in HUD (hud_system.go:389-449)
  - ✅ `GetBonusesForGuild()` method added to `pkg/world/territory/manager.go:499-512`
  - ✅ `TerritorySystem` implements `TerritoryBonusProvider` (territory_system.go:217-224)
  - ✅ HUD wired to TerritorySystem in `cmd/client/handlers.go:2194`
  - ✅ Test added: `TestGetBonusesForGuild` in manager_test.go
- **Resolution**: HUD now displays resource and XP bonuses for players in guilds with controlled territories.

---

## Gap 5: FPS Benchmark Scope Too Narrow

- **Stated Goal**: "60 FPS minimum on mid-range hardware"
- **Current State**:
  - `pkg/benchmark/fps/benchmark_test.go` tests MovementSystem with 2000 entities
  - Results show ~16,234 ns/op (well under 16.67ms target)
  - ❌ Does not test all 66 systems together
  - ❌ Does not include collision detection
  - ❌ Does not include rendering pipeline
- **Impact**: Benchmark validates ECS overhead but not real-world gameplay performance. Performance regressions in other systems could go undetected.
- **Closing the Gap**:
  1. Create `BenchmarkFullSystemSuite` that calls `InitializeGameSystems()` with all systems
  2. Add `BenchmarkCollisionSystem` with realistic entity density
  3. Add `BenchmarkRenderPipeline` (requires xvfb or headless mode)
  4. Add CI gate that fails on >20% regression from baseline
  - **Estimated effort**: 1-2 days
  - **Validation**: `go test -bench=BenchmarkFull ./pkg/benchmark/fps/`

---

## Gap 6: Signal Handler Integration Test Missing

- **Status**: ✅ RESOLVED
- **Stated Goal**: Server should gracefully shut down on SIGTERM.
- **Current State**:
  - ✅ `cmd/server/shutdown_test.go` exists with comprehensive tests:
    - `TestGracefulShutdown_SignalHandling` — Tests SIGINT and SIGTERM
    - `TestGracefulShutdown_DeadlineEnforcement` — Tests shutdown deadline
    - `TestGracefulShutdown_ContextPropagation` — Tests context propagation
    - `TestRunGameLoop_ContextCancellation` — Tests game loop stops on cancellation
    - `TestShutdownSequence_AllComponentsStop` — Tests all components shut down cleanly
- **Resolution**: Integration tests already exist and validate graceful shutdown behavior.

---

## Gap 7: Territory System Lacks Mod Support

- **Status**: ✅ RESOLVED
- **Stated Goal**: Modding system allows server customization.
- **Current State**:
  - ✅ `TerritoryConfig` struct allows runtime configuration (types.go:133-163)
  - ✅ `Manager.SetConfig()` method allows programmatic override (manager.go:41-58)
  - ✅ All territory mechanics use `m.config.*` values instead of constants
  - ✅ Example mod created: `mods/fast-sieges.json` 
  - ✅ Documentation created: `mods/README.md` with territory rules section
- **Resolution**: Territory system is fully configurable via `TerritoryConfig` struct. Mods can be loaded and applied via `SetConfig()`. Constants in types.go serve as defaults but all runtime logic uses the config.

---

## Gap 8: Trade Validation Lacks Per-Item Quantity

- **Status**: ✅ RESOLVED (Design Decision)
- **Stated Goal**: Reject "negative-quantity and zero-value trades" (from previous GAPS.md)
- **Current State**:
  - `pkg/validation/trade.go:ValidateTradeQuantity()` exists and validates individual quantities
  - Trade data model (`TradeProposal.OfferedItems`) uses `[]string` item IDs
  - Items in `pkg/procgen/item/types.go:Item` are unique instances (not stacked)
  - Inventory (`pkg/engine/inventory_components.go`) holds `[]*item.Item` where each item is unique
- **Resolution**: The current design intentionally treats items as unique instances (like equipment in most ARPGs), not stackable commodities. Each item has a unique ID, and trades transfer ownership of specific item instances. The `ValidateTradeQuantity()` function is available for future use if stackable items are added. No code changes needed.
- **Impact**: None. Current design is consistent with the ARPG genre where equipment is unique. If stackable resources (ores, potions) are needed later, the data model can be extended.

---

## Gap 9: memprofile Uses fmt.Printf Instead of Structured Logging

- **Status**: ✅ RESOLVED (Option A applied)
- **Stated Goal**: Structured logging throughout codebase (logrus).
- **Current State**:
  - ✅ `PrintProfile()` method has explicit documentation exemption (lines 195-197):
    > "NOTE: This function intentionally uses fmt.Printf for CLI/debug output.
    > It is exempt from the structured logging guideline (Coding Guideline #3)
    > as it's designed for human-readable console output in testing and debugging tools."
  - ✅ `ExportJSON()` method added for machine-readable structured export (lines 325-371)
- **Resolution**: Classified as CLI debugging tool. Human-readable output intentional. JSON export available for structured consumption.

---

## Gap 10: No Automated CI Gate for Performance Benchmarks

- **Status**: ✅ RESOLVED
- **Stated Goal**: Maintain 60 FPS and <500MB memory.
- **Current State**:
  - ✅ Benchmarks exist in `pkg/benchmark/fps/` and `pkg/benchmark/memory/`
  - ✅ Scripts exist: `scripts/benchmark-regression.sh`, `scripts/benchmark-memory.sh`
  - ✅ Integrated into CI (`.github/workflows/test.yml` lines 63-69):
    - FPS benchmark regression check via `xvfb-run ./scripts/benchmark-regression.sh`
    - Memory benchmark check via `xvfb-run ./scripts/benchmark-memory.sh`
  - ✅ Baseline stored in `scripts/benchmark-baseline.json`
- **Resolution**: CI gates fully implemented. Performance regressions will be detected before merge.

---

## Gap 11: ECS Race Conditions on Entity Staging Buffers (C-001/C-002)

- **Status**: ✅ RESOLVED
- **Stated Goal**: Thread-safe ECS for concurrent server-join and game-loop operations.
- **Root Cause**:
  - `World.CreateEntity()` read/incremented `nextEntityID` and appended to `entitiesToAdd` without any synchronization.
  - `World.AddEntity()` and `World.RemoveEntity()` appended to staging slices without synchronization.
  - `World.Update()` read and drained staging slices on the game-loop goroutine without locking, racing with concurrent writers.
- **Fix Applied** (`pkg/engine/ecs.go`):
  - Added `entityMu sync.Mutex` field to `World` struct (separate from `mu sync.RWMutex` used for system-list and metrics access).
  - `CreateEntity()`, `AddEntity()`, `RemoveEntity()` lock `entityMu` around all staging-buffer mutations.
  - `Update()` and `processPendingEntityAdditions()` snapshot the staging slices under `entityMu` (swap with `nil`) then process the snapshot outside the lock, keeping the critical section minimal.
- **Test Added**: `TestCreateEntityConcurrentSafety` in `pkg/engine/ecs_test.go` — spawns 20 goroutines each creating 50 entities; passes cleanly under `go test -race`.

---

## Gap 12: Unprotected Global Talent Registry Maps (H-008)

- **Status**: ✅ RESOLVED
- **Root Cause**: Package-level `talentRegistry` (`map[string]*TalentDefinition`) and `categoryTalents` (`map[TalentCategory][]*TalentDefinition`) had no mutex protection. Concurrent reads from the game loop and writes from the mod system (which can call `registerTalent()` at runtime) would cause a fatal Go map-concurrency panic.
- **Fix Applied** (`pkg/engine/talent_definitions.go`):
  - Added `talentMu sync.RWMutex` package-level variable.
  - `GetTalentDefinition()`, `GetAllTalentDefinitions()`, `GetTalentsByCategory()` acquire `talentMu.RLock()`.
  - `registerTalent()` acquires `talentMu.Lock()`.
  - `init()` is single-threaded by Go's runtime guarantee, so no lock needed there.

---

## Gap 13: ChatUI HandleClick Double-Activation on Overlapping Regions (H-009)

- **Status**: ✅ RESOLVED
- **Root Cause**: `ChatUI.HandleClick()` in `pkg/rendering/ui/chat.go` used two independent `if` statements for the tab row and input field regions. On a small window where the regions overlap, both handlers fired for a single click, simultaneously switching the channel and activating text input.
- **Fix Applied** (`pkg/rendering/ui/chat.go`):
  - Changed the second `if` to `else` — input field click is only tested when the tab row check did not match.
  - Added comment explaining the mutual-exclusion intent.

---

## Summary

| Gap | Severity | Effort | Status |
|-----|----------|--------|--------|
| Gap 1: Voice network transport | 🔴 CRITICAL | 2-3 days | Open |
| Gap 2: Quest postapoc templates | 🟢 LOW | N/A | ✅ Resolved |
| Gap 3: README system count | 🟢 LOW | N/A | ✅ Resolved |
| Gap 4: Territory HUD display | 🟢 LOW | N/A | ✅ Resolved |
| Gap 5: FPS benchmark scope | 🟡 MEDIUM | 1-2 days | Open |
| Gap 6: Signal handler test | 🟢 LOW | N/A | ✅ Resolved |
| Gap 7: Territory mod support | 🟢 LOW | N/A | ✅ Resolved |
| Gap 8: Trade quantity model | 🟢 LOW | N/A | ✅ Resolved (design decision) |
| Gap 9: memprofile logging | 🟢 LOW | N/A | ✅ Resolved |
| Gap 10: Performance CI gate | 🟢 LOW | N/A | ✅ Resolved |
| Gap 11: ECS entity staging races | 🔴 CRITICAL | Fixed | ✅ Resolved |
| Gap 12: Talent registry unprotected | 🟡 MEDIUM | Fixed | ✅ Resolved |
| Gap 13: ChatUI double-activation | 🟡 MEDIUM | Fixed | ✅ Resolved |

---

*Updated: 2026-04-19 | 11 of 13 gaps resolved*
