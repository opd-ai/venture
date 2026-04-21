# IMPLEMENTATION GAP AUDIT — 2026-04-21

Audited commit: HEAD on branch `copilot/audit-implementation-gaps-again`.  
Build baseline: `go build ./...` — clean (Ebiten requires X11 headers not present in CI; all pure-Go packages build without errors). `go vet ./...` — clean.  
Previous gap file (GAPS.md, dated 2026-04-21) reviewed; all 12 prior gaps re-verified; status noted below.

---

## Project Architecture Overview

**Venture** is a fully procedural multiplayer action-RPG distributed as a single binary with zero external asset files. All graphics, audio, terrain, items, quests, NPCs, and UI are generated at runtime from seed-based deterministic algorithms. The codebase implements:

- **ECS core** (`pkg/engine/`) — 100+ systems, 2 206 struct types, 355 New*System constructors, 105 interfaces, ≈210 000 non-test LOC.
- **Procedural generators** (`pkg/procgen/`) — 25 sub-packages: terrain (BSP/cellular/city/forest), entities, items, quests, magic, dialog, narrative (fragment, branching, archaeology, timeline, cross-dungeon), story, minigames, puzzles, books, and more.
- **Rendering pipeline** (`pkg/rendering/`) — sprites, tiles, lighting, particles, post-processing, UI, animation, cache.
- **Audio synthesis** (`pkg/audio/`) — adaptive music, SFX, voice (ADPCM), synthesis engine.
- **Multiplayer networking** (`pkg/network/`) — authoritative server, client-side prediction, lag compensation, federation, voice chat.
- **Persistent world** (`pkg/world/`) — housing, economy, territory, raids, chunk-based persistence.
- **Integration layer** (`pkg/integration/`) — 9 cross-package integration adapters.

**Stated platforms**: Linux, macOS, Windows, WebAssembly, iOS, Android.  
**Stated multiplayer design**: 200–5000 ms latency tolerance (Tor/onion services).

---

## Gap Summary

| Category           | Count | Critical | High | Medium | Low |
|--------------------|-------|----------|------|--------|-----|
| Stubs / TODOs      |     3 |        0 |    1 |      1 |   1 |
| Dead Code          |     2 |        0 |    0 |      1 |   1 |
| Partially Wired    |     4 |        0 |    1 |      2 |   1 |
| Interface Gaps     |     1 |        0 |    0 |      1 |   0 |
| Dependency Gaps    |     0 |        0 |    0 |      0 |   0 |
| **TOTAL**          | **10**|      **0**| **2**|    **5**| **3**|

---

## Implementation Completeness by Package

Key packages audited. Coverage percentages are from existing `pkg/*/AUDIT.md` files where present; otherwise estimated from code structure.

| Package                     | Exported Fns | Impl. | Stubs | Dead | Notes |
|-----------------------------|-------------|-------|-------|------|-------|
| `pkg/engine`                | ~2 800      | ✅    | 1     | 1    | See G1, G8 |
| `pkg/procgen/story`         | 47          | ✅    | 0     | 3    | See G4, G5, G6 |
| `pkg/network`               | ~180        | ✅    | 0     | 0    | G2 receive path missing |
| `pkg/rendering/ui`          | ~160        | ✅    | 0     | 1    | See G6 |
| `pkg/mobile`                | ~40         | ✅    | 0     | 0    | G9 resolved |
| `pkg/integration/guild_vehicle` | 12      | ⚠️    | 0     | 0    | G3 partial |
| `pkg/procgen/terrain`       | ~80         | ✅    | 0     | 0    | 94% coverage |
| `pkg/audio`                 | ~60         | ✅    | 0     | 0    | 91–98% coverage |
| `pkg/world`                 | ~120        | ✅    | 0     | 0    | Chunk persistence now wired |
| `pkg/vr`                    | ~30         | ⚠️    | 16    | 0    | Explicitly experimental |

---

## Findings

### HIGH

- [x] **G1 — Story generator uses `"postapocalyptic"` but genre registry uses `"postapoc"`** — `pkg/procgen/story/generator.go:16-22` — `genreThemes` and `genreTitleSuffixes` are keyed by `"postapocalyptic"`; the CLI flag (`-genre=postapoc`), genre registry (`pkg/procgen/genre/predefined.go:86`), quest templates (`pkg/procgen/quest/templates.go:278`), and terrain genre mapping all use `"postapoc"`. At runtime, `story.FragmentGenerator.getThemesForGenre("postapoc")` finds no match (`:334-340`), logs a warning, and silently returns `defaultThemes`. Players selecting the post-apocalyptic genre receive generic story themes and titles instead of theme-appropriate ones. The unit tests (`generator_test.go:275`) also use `"postapocalyptic"`, masking the mismatch. — **Blocked goal**: "genre-based theming (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)". — **Remediation**: In `pkg/procgen/story/generator.go`, change the two map literal keys on lines 21 and 30 from `"postapocalyptic"` to `"postapoc"`. Update `generator_test.go:275` and `archaeology_test.go:238`, `crossdungeon_test.go:265`, `timeline.go` case labels (`:240`, `:361`, `:378`, `:395`, `:412`, `:429`, `:446`, `:463`, `:480`, `:532`, `:549`, `:577`) to use `"postapoc"` consistently. Validate with `go test ./pkg/procgen/story/... ./pkg/procgen/...`.

- [x] **G2 — `AnimationSyncManager` receive path (`BufferState`/`GetNextState`) has zero callers** — `pkg/network/animation_sync.go:234-265` — The send side is fully wired: `cmd/client/handlers.go:676-677` creates `animSyncMgr` and calls `animationSystem.SetSyncManager(animSyncMgr)`; `pkg/engine/animation_system.go:823-828` gates outgoing sends through `ShouldSync`/`RecordSync`. However, `AnimationSyncManager.BufferState` and `GetNextState` (the receive/playback path) have zero callers outside tests anywhere in `cmd/` or `pkg/`. Received animation packets from remote players are therefore not buffered or jitter-compensated. — **Blocked goal**: "high-latency multiplayer (200–5000 ms) … smooth remote-player animations". — **Remediation**: In the network client state-update handler (search `pkg/network/client.go` for the animation state component type), call `mgr.BufferState(packet)` on receipt. In `pkg/engine/animation_system.go`, add a `drainRemoteBuffer(entityID uint64)` helper that calls `mgr.GetNextState(entityID)` each `Update` tick and applies the result to the remote entity's `AnimationComponent`. Gate on `syncManager != nil`. Add `TestAnimationSyncReceivePath` covering buffer → drain round-trip. Validate with `go test ./pkg/network/ ./pkg/engine/`.

---

### MEDIUM

- [x] **G3 — `pkg/integration/guild_vehicle`: Formation physics and `VehicleComponent.FleetID` sync not wired** — `pkg/integration/guild_vehicle/integrations.go:57-116` — The four-interface injection layer (membership, vehicle sync, structure damage, formation offset) is fully defined. Guild membership validation (`SetMembershipValidator`) is wired from `cmd/server/main.go`. However `SetVehicleSyncer` (`:74`) and the `FormationOffset` return path (`:102`) have no callers that feed results into `pkg/engine/physics/vehicle/`. Vehicles in a fleet do not move to formation positions each frame, and fleet membership is not reflected on `VehicleComponent`. — **Blocked goal**: Guild × vehicle interplay. — **Remediation**: (1) After vehicle spawning, call `fleet.SetVehicleSyncer(vehicleSyncer)` with an adapter that writes `VehicleComponent.FleetID`. (2) In `pkg/engine/physics/vehicle/`, read `GetFormationOffsets` each frame and apply as target position for non-leader vehicles. Validate with `go test ./pkg/integration/guild_vehicle/ ./pkg/engine/physics/vehicle/`.

- [x] **G4 — `ArchaeologyGenerator`, `TimelineGenerator`, `CrossDungeonGenerator` fully implemented but never used at runtime** — `pkg/procgen/story/archaeology.go:67`, `timeline.go:73`, `crossdungeon.go:47` — Three complete generators with tests and doc coverage of 88.7% exist but are never called from `cmd/client/util.go` or any spawner. Only `FragmentGenerator` is wired (`cmd/client/util.go:2397`). This is documented as "DEFERRED Phase 2" in `pkg/procgen/story/AUDIT.md`. — **Blocked goal**: Story depth stated in README ("environmental story fragments … branching narrative"). — **Remediation**: Implement ECS components (`ArchaeologicalSiteComponent`, `TimelineComponent`, `CrossDungeonStoryComponent`) mirroring `StoryFragmentComponent`. Add spawner functions in `cmd/client/util.go` analogous to `spawnStoryFragments`. Register systems that advance excavation/timeline state. This is a medium–large effort already scoped in the story AUDIT.md. Validate with `go test ./pkg/procgen/story/ ./pkg/engine/` after integration.

- [x] **G5 — `StoryJournalUI` fully implemented but never instantiated** — `pkg/rendering/ui/story_journal.go:17-68` — `NewStoryJournalUI` and its full navigation + input + genre-theming implementation exist but have zero callers in `cmd/`. `DiscoverySystem` populates `StoryJournalComponent` but players have no way to open the journal. Documented as "DEFERRED Phase 2" in `pkg/procgen/story/AUDIT.md`. — **Blocked goal**: Story discoverability / lore UX. — **Remediation**: In `cmd/client/handlers.go`, instantiate `NewStoryJournalUI(x, y, w, h, *genreID)`, bind a keybind (e.g., `L`), and render it in the HUD draw path. Validate with `go build ./cmd/client/`.

- [x] **G6 — `GenreComponent` attached only to the player entity; NPC entities still carry no `GenreComponent`** — `cmd/client/handlers.go:2753` — `NewGenreComponent(*genreID)` is now added to the player entity. However `pkg/engine/objective_tracker_system.go:761` reads `entity.GetComponent("genre")` for any entity when computing quest reward scaling — for all non-player entities this returns nil and the scaling branch is dead. `NewBlendedGenreComponent` (genre blending for multi-genre worlds) has zero callers. — **Blocked goal**: Per-entity genre-based theming for NPCs and enemies. — **Remediation**: In the NPC/entity spawner (e.g., `pkg/engine/` spawn helpers called from `cmd/client/util.go:SpawnEnemiesInTerrain`), attach `NewGenreComponent(params.GenreID)` to spawned entities. If multi-genre blending is desired, use `NewBlendedGenreComponent` where secondary genres are available. Validate with `go test ./pkg/engine/`.

---

### LOW

- [x] **G7 — `pkg/engine/equipment_durability_particle_system.go` excluded from all builds (`//go:build ignore`)** — `pkg/engine/equipment_durability_particle_system.go:1` — The base `EquipmentDurabilityParticleSystem` (279 LOC, 4 methods) and its test file are both tagged `//go:build ignore`, meaning they are compiled into no binary and the test suite never runs them. Four specialised variants (terrain, combat, reputation, plus `WeatherEquipmentSheenSystem`) superseded this base class and are all registered in `system_init.go`. — **Blocked goal**: None — this is maintenance debt. — **Remediation**: Option A (preferred): Delete `pkg/engine/equipment_durability_particle_system.go` and `pkg/engine/equipment_durability_particle_system_test.go`; the specialised variants are complete. Option B: Remove the `//go:build ignore` tag and wire the base class alongside the specialised systems if a general-purpose equipment particle trigger is still wanted. Validate with `go build ./... && go vet ./...`.

- [x] **G8 — `NewInteractWithEnvironmentNode` never composed into any behavior tree** — `pkg/engine/behavior_tree_advanced_nodes.go:302` — The node type is fully implemented (Execute, tick, logging) but no behavior-tree builder in `pkg/engine/` or `cmd/` composes it. NPC gatherers, crafters, and interactive-object users therefore cannot use environment interaction via the behavior-tree path. — **Blocked goal**: NPC procedural behavior depth. — **Remediation**: In the NPC behavior-tree construction helper (search `pkg/engine/` for `BehaviorTree{` or `NewSelectorNode`), add an `InteractWithEnvironmentNode` branch for crafter/gatherer NPC archetypes, guarded by a `HasNearbyInteractable` condition node. Add a tree-construction test asserting the branch is reachable. Validate with `go test ./pkg/engine/`.

- [x] **G9 — `NewHelpSystemWithSize` exported but has zero external callers** — `pkg/engine/help_system.go:50` — `NewHelpSystem` delegates to `NewHelpSystemWithSize(800, 600)`. No external caller passes custom dimensions, making the parameterised variant effectively private API surface. — **Blocked goal**: None — API surface clarity. — **Remediation**: Either (a) lower-case to `newHelpSystemWithSize` so it is not part of the public API, or (b) pass actual screen dimensions from the renderer through `NewHelpSystemWithSize` so different display resolutions are handled correctly. Validate with `go build ./... && go vet ./...`.

- [x] **G10 — OpenXR VR adapters are documented stubs** — `pkg/engine/vr_openxr_adapters.go:56-230` (build tag `//go:build vr && !js`) — 16 `TODO(vr-sdk):` markers describe the OpenXR SDK calls needed. All methods return zero values. `cmd/client/init_versions.go:583,595` uses `NewStubHeadsetAdapter`/`NewStubControllerAdapter` unconditionally (not gated on the `vr` build tag path to OpenXR). README explicitly marks VR as "experimental with mock adapters only; no hardware SDK integration." — **Blocked goal**: None at stated-goal level — classified LOW per README. — **Remediation**: Tracked in ROADMAP.md Priority 4. When a Go OpenXR binding is available, replace each `TODO(vr-sdk)` stub per the inline pseudocode and switch `init_versions.go:583,595` to the OpenXR constructors when `--vr` is set without `--force-stub`. No action required until SDK is available.

---

## Re-verification of Previous GAPS.md Gaps

The following gaps from the prior `GAPS.md` (2026-04-21) were re-verified and are **closed**:

| Prior Gap | Prior Status | Current Status | Evidence |
|-----------|-------------|----------------|----------|
| Gap 2 — `SkillPointGainParticleSystem` unregistered | OPEN | **CLOSED** | `system_init.go:1499-1503` instantiates system; `AddSkillPointCallback` wires `OnSkillPointGain` |
| Gap 3 — `GenreComponent` never attached | OPEN | **PARTIALLY CLOSED** | `handlers.go:2753` adds to player; NPCs still missing → see G6 above |
| Gap 4 — `SpellComponent` carryover unpopulated | OPEN | **CLOSED** | `spell_casting.go:3061-3068` creates and populates `SpellComponent` |
| Gap 5 — Chunk compression bytes discarded | OPEN | **CLOSED** | `cmd/server/main.go:541-546` calls `worldPersistence.SaveChunk`; `pkg/world/persistence.go:427-454` implements `SaveChunk`/`LoadChunk` |
| Gap 7 — 3 alternate constructors dead | OPEN | **CLOSED** | `NewAnimationAdapter`, `NewLightingAdapter`, `NewTimer` each delegate to their with-options variants |
| Gap 9 — Native mobile keyboard JNI/UIKit | OPEN | **CLOSED (iOS)** | `keyboard_ios.go:111-112` returns `true`; Android returns based on `C.hasAndroidActivity()` |
| Gap 10 — `NewBaseStatsFromEntity` orphan | OPEN | **CLOSED** | `handlers.go:2746` calls `engine.NewBaseStatsFromEntity(player)` |

The following remain **open** (carried forward with updated analysis in the findings above):

| Prior Gap | Current Finding |
|-----------|----------------|
| Gap 1 — `AnimationSyncManager` unwired | → G2 (receive path still missing) |
| Gap 6 — `guild_vehicle` cross-pkg integration | → G3 (formation physics + FleetID sync missing) |
| Gap 8 — OpenXR VR stubs | → G10 (unchanged; explicitly LOW) |
| Gap 11 — `InteractWithEnvironmentNode` orphan | → G8 (unchanged) |
| Gap 12 — `NewHelpSystemWithSize` effectively private | → G9 (unchanged) |

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|-----------------|
| `memprofile.PrintProfile` uses `fmt.Printf` | File comment at `:195` explicitly states "NOTE: This function intentionally uses fmt.Printf for CLI/debug output." The function is part of a profiling CLI utility and `fmt` is the correct output channel. |
| `NewMockHeadset`/`NewMockController` in `vr_controller_system.go`, `head_tracking_system.go` | These are mock adapters for offline/CI use, not dead code. They are the production fallback path used by `init_versions.go:583,595`. |
| Stub constructors (`NewStubInput`, `NewStubGame`, etc.) in `_test.go` | All defined in `*_test.go` files; intentional test infrastructure. |
| `StaminaRegenSystem` in `pkg/engine/doc.go:208-217` | Only appears as a commented-out code example in package documentation, not as a real source file. No implementation exists; no gap. |
| `//go:build !headless` / `//go:build headless` paired files | Intentional conditional compilation for GPU-shader vs. headless CI environments. Both paths are complete implementations. |
| `//go:build vr && !js` on `vr_openxr_adapters.go` | Intentional gating for future VR SDK integration. Documented in ROADMAP. |
| Systems defined in `pkg/engine/` but not in `system_init.go` | Most are registered in `cmd/client/init_versions.go`, `cmd/client/handlers.go`, or `cmd/server/v*_systems.go`. The `comm -23` diff was cross-checked against all cmd/ callers; none are dangling. |
| `pkg/procgen/story/AUDIT.md` deferred items (serialization, Markov chains, terrain-aware positioning) | Explicitly scoped as Phase 2 deferred work in a dated audit file. Not undiscovered gaps. |
| Interfaces with single implementations | Most single-implementation interfaces in `pkg/engine/` are ECS component or adapter interfaces; the second implementation is the stub/mock. This is the intentional pattern for testability. |
