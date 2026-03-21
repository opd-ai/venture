# Implementation Gaps — 2026-03-21

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

- **Stated Goal**: README claims "genre system supporting fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic themes" with procedural quest generation.
- **Current State**:
  - `pkg/procgen/quest/generator.go:selectTemplates()` has switch cases for:
    - ✅ `"fantasy"` (line 107)
    - ✅ `"scifi"` (lines 92-96)
    - ✅ `"horror"` (lines 97-101)
    - ✅ `"cyberpunk"` (lines 102-106)
    - ❌ `"postapoc"` — **MISSING**
  - Default case falls through to fantasy templates (line 110)
  - No `GetPostApocKillTemplates()`, `GetPostApocCollectTemplates()` functions defined
- **Impact**: Post-apocalyptic genre quests use fantasy templates (e.g., "Slay the Dragon" instead of "Clear the Raider Camp"). Breaks genre immersion.
- **Closing the Gap**:
  1. Add `case "postapoc":` to `selectTemplates()` switch statement after line 106
  2. Create post-apocalyptic quest template functions in `pkg/procgen/quest/templates.go`:
     - `GetPostApocKillTemplates()` — e.g., "Eliminate the Raider Boss", "Hunt the Mutant Pack"
     - `GetPostApocCollectTemplates()` — e.g., "Scavenge Medical Supplies", "Find Clean Water"
     - `GetPostApocBossTemplates()` — e.g., "Defeat the Warlord", "Destroy the Rogue AI"
     - `GetPostApocExploreTemplates()` — e.g., "Map the Ruins", "Scout the Wasteland"
  3. Add test: `TestQuestGenerator_PostApocGenre`
  - **Estimated effort**: 4-6 hours
  - **Validation**: `go test -v ./pkg/procgen/quest/... -run Postapoc`

---

## Gap 3: README Claims "100+ Systems" but Actual Count is 66

- **Stated Goal**: README states "100+ game systems" in project description.
- **Current State**:
  - `pkg/engine/system_init.go` explicitly logs: `"initializing game systems (66 total)"`
  - 343 files ending in `_system.go` exist but include tests, stubs, and specialized variants
  - Actual registered systems via `world.AddSystem()`: 66
- **Impact**: Documentation inaccuracy. Sets incorrect expectations for contributors and users.
- **Closing the Gap**:
  1. Update README.md to state "66 game systems" instead of "100+"
  2. Alternatively, if the intent is to count all system-related code, clarify as "66 core systems with 340+ specialized variants"
  - **Estimated effort**: 15 minutes
  - **Validation**: `grep -n "66 total" pkg/engine/system_init.go`

---

## Gap 4: Territory Bonuses Not Displayed in HUD

- **Stated Goal**: Territory control provides gameplay bonuses.
- **Current State**:
  - `pkg/world/territory/manager.go` calculates bonuses (+10% resource, +5% XP)
  - Bonuses are applied to gameplay calculations
  - ❌ No visual indicator in player HUD
  - Players cannot see what bonuses they're receiving from controlled territory
- **Impact**: Players unaware of territory benefits. Reduces motivation to engage with territory system.
- **Closing the Gap**:
  1. Add `TerritoryBonusIndicator` to `pkg/engine/hud_system.go`
  2. Query `TerritoryManager.GetBonusesForPlayer(playerID)` each frame
  3. Render bonus icons/text when player is in controlled territory
  - **Estimated effort**: 2-4 hours
  - **Validation**: Visual inspection in-game; add `TestHUD_TerritoryBonuses`

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

- **Stated Goal**: Server should gracefully shut down on SIGTERM.
- **Current State**:
  - `cmd/server/main.go:106` uses `signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)`
  - Context cancellation propagates to subsystems
  - ❌ No integration test validates exit code 0 on SIGTERM
- **Impact**: Graceful shutdown can only be validated manually. Regressions could go undetected.
- **Closing the Gap**:
  1. Create `TestServerGracefulShutdown` in `cmd/server/` or `pkg/integration/`
  2. Start server as subprocess
  3. Send `SIGTERM`
  4. Assert exit code 0 within 5-second timeout
  - **Estimated effort**: 2-4 hours
  - **Validation**: `go test -v ./cmd/server/... -run GracefulShutdown`

---

## Gap 7: Territory System Lacks Mod Support

- **Stated Goal**: Modding system allows server customization.
- **Current State**:
  - `pkg/modding/` provides sandboxed JSON mod loading
  - Territory/siege constants are hard-coded:
    - `BaseCaptureTime = 60` seconds
    - `GuildHallHP = 1000`
    - `SiegePreparationDuration = 1 hour`
  - ❌ No `ModRuleProvider` integration for territory system
- **Impact**: Server operators cannot customize siege mechanics without code changes.
- **Closing the Gap**:
  1. Replace hard-coded constants with `world.GetModRuleFloat64()` calls:
     ```go
     captureTime := world.GetModRuleFloat64("siege.captureTime", 60.0)
     guildHallHP := world.GetModRuleFloat64("siege.guildHallHP", 1000.0)
     ```
  2. Document territory mod rules in `mods/README.md`
  3. Add example mod: `mods/fast-sieges.json`
  - **Estimated effort**: 4-6 hours
  - **Validation**: Create test mod and verify constants change

---

## Gap 8: Trade Validation Lacks Per-Item Quantity

- **Stated Goal**: Reject "negative-quantity and zero-value trades" (from previous GAPS.md)
- **Current State**:
  - `pkg/validation/trade.go` validates item IDs and item counts
  - Trade data model passes item IDs as string slices with no associated quantity field
  - Cannot validate quantities like "5x Iron Ore" because model lacks quantity
- **Impact**: `ValidateTradeQuantity` method exists but is never called from trade flow because data model doesn't support it.
- **Closing the Gap**:
  1. Decide: Should trade system support per-item quantities (e.g., "5x Iron Ore") or is each item ID unique?
  2. If quantities needed: Update `pkg/network/trade/` and `pkg/engine/trade_system.go` data structures to include quantity field
  3. Wire `ValidateTradeQuantity()` into trade flow
  - **Estimated effort**: 4-8 hours depending on scope
  - **Validation**: `go test -v ./pkg/validation/... -run TradeQuantity`

---

## Gap 9: memprofile Uses fmt.Printf Instead of Structured Logging

- **Stated Goal**: Structured logging throughout codebase (logrus).
- **Current State**:
  - `pkg/memprofile/profile.go:Print()` uses 6+ `fmt.Printf` calls
  - Outputs human-readable memory profile report
  - May be intentional CLI/debug tool output
- **Impact**: Violates "zero unstructured logging" if classified as production code. Acceptable if classified as debugging tool.
- **Closing the Gap**:
  1. **Option A**: Classify `pkg/memprofile` as CLI debugging tool (exempt from structured logging). Document exemption.
  2. **Option B**: Migrate to `logrus.WithFields()` for consistency. Changes output format.
  - **Estimated effort**: 1-2 hours
  - **Validation**: `grep -n "fmt.Print" pkg/memprofile/`

---

## Gap 10: No Automated CI Gate for Performance Benchmarks

- **Stated Goal**: Maintain 60 FPS and <500MB memory.
- **Current State**:
  - Benchmarks exist in `pkg/benchmark/fps/` and `pkg/benchmark/memory/`
  - Scripts exist: `scripts/benchmark-regression.sh`, `scripts/benchmark-memory.sh`
  - ❌ Not integrated into CI (`.github/workflows/test.yml`)
  - Performance regressions can ship undetected
- **Impact**: Performance degradation could reach users before detection.
- **Closing the Gap**:
  1. Add CI step to `test.yml`: `xvfb-run go test -bench=BenchmarkFPS2000Entities ./pkg/benchmark/fps/`
  2. Parse output, compare against baseline (store in `scripts/benchmark-baseline.json`)
  3. Fail CI on >20% regression
  4. Similarly for memory: fail if heap >500MB
  - **Estimated effort**: 4-8 hours
  - **Validation**: CI fails when performance-degrading PR is submitted

---

## Summary

| Gap | Severity | Effort | Status |
|-----|----------|--------|--------|
| Gap 1: Voice network transport | 🔴 CRITICAL | 2-3 days | Open |
| Gap 2: Quest postapoc templates | 🟡 HIGH | 4-6 hours | Open |
| Gap 3: README system count | 🟡 HIGH | 15 min | Open |
| Gap 4: Territory HUD display | 🟡 MEDIUM | 2-4 hours | Open |
| Gap 5: FPS benchmark scope | 🟡 MEDIUM | 1-2 days | Open |
| Gap 6: Signal handler test | 🟡 MEDIUM | 2-4 hours | Open |
| Gap 7: Territory mod support | 🟢 LOW | 4-6 hours | Open |
| Gap 8: Trade quantity model | 🟢 LOW | 4-8 hours | Open |
| Gap 9: memprofile logging | 🟢 LOW | 1-2 hours | Open |
| Gap 10: Performance CI gate | 🟡 MEDIUM | 4-8 hours | Open |

---

*Generated: 2026-03-21 | Source: Functional Audit comparing README claims to implementation*
