# IMPLEMENTATION GAP AUDIT — 2026-04-21

Audited commit: HEAD on branch `copilot/audit-implementation-gaps-again`.  
Build baseline: `go build ./...` — clean for all pure-Go packages; Ebiten requires X11 headers not present in this CI environment, so the game binaries themselves are not compiled here but the package graph is intact.  
`go vet ./...` — clean on all pure-Go packages.  
Prior gap file (GAPS.md, dated 2026-04-21) reviewed; prior findings re-verified and updated below.

---

## Project Architecture Overview

**Venture** is a fully procedural multiplayer action-RPG distributed as a single binary with zero external asset files. All graphics, audio, terrain, items, quests, NPCs, and UI are generated at runtime from seed-based deterministic algorithms.

- **ECS core** (`pkg/engine/`) — 100+ systems, ~2 206 struct types, 355 `New*System` constructors, 105 interfaces, ≈386 000 non-test LOC.
- **Procedural generators** (`pkg/procgen/`) — 25 sub-packages: terrain (BSP/cellular/city/forest/composite/grammar/maze/L-system), entity, item, quest, magic, dialog, narrative (fragment, branching, archaeology, timeline, cross-dungeon), story, minigame, puzzle, book, and more.
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
| Stubs / TODOs      |     1 |        0 |    0 |      0 |   1 |
| Dead Code          |     2 |        0 |    0 |      2 |   0 |
| Partially Wired    |     4 |        0 |    1 |      2 |   1 |
| Interface Gaps     |     1 |        0 |    0 |      1 |   0 |
| Genre Key Mismatch |     1 |        0 |    1 |      0 |   0 |
| **TOTAL**          | **9** |      **0**| **2**|    **5**| **2**|

---

## Implementation Completeness by Package

Key packages audited. Coverage percentages from existing per-package `AUDIT.md` files where present; otherwise estimated from code review.

| Package                          | Exported Fns | Impl. | Stubs | Dead | Notes |
|----------------------------------|-------------|-------|-------|------|-------|
| `pkg/engine`                     | ~2 800      | ✅    | 0     | 2    | N3, N4 (dead objects in systemsContainer) |
| `pkg/procgen/story`              | 47          | ✅    | 0     | 0    | All generators now wired in client |
| `pkg/procgen/skills`             | ~40         | ⚠️    | 0     | 0    | N1: `normalizeGenre` misses `"postapoc"` |
| `pkg/procgen/class`              | ~30         | ⚠️    | 0     | 0    | N1: genre theme keys use `"postapocalyptic"` |
| `pkg/procgen/terrain`            | ~80         | ⚠️    | 0     | 0    | N1: composite.go map key mismatch |
| `pkg/network`                    | ~180        | ✅    | 0     | 0    | G2 receive path now wired |
| `pkg/rendering/ui`               | ~160        | ✅    | 0     | 0    | StoryJournalUI now instantiated |
| `pkg/integration/guild_vehicle`  | 12          | ✅    | 0     | 0    | Formation + VehicleSyncer now wired |
| `pkg/engine/merchant_spawn`      | 5           | ⚠️    | 0     | 0    | N5: No `GenreComponent` on merchants |
| `cmd/client`                     | ~120        | ⚠️    | 0     | 2    | N2, N3, N4, N7 |
| `pkg/vr`                         | ~30         | ⚠️    | 16    | 0    | Explicitly experimental (OpenXR TODO stubs) |

---

## Findings

### HIGH

- [ ] **N1 — Multi-package `"postapocalyptic"` key inconsistency breaks post-apocalyptic genre at runtime** — Multiple files — When the game is started with `-genre=postapoc` (the canonical genre ID from `pkg/procgen/genre/predefined.go:86` and the CLI flag), at least five packages use `"postapocalyptic"` as map keys or case labels and therefore silently fall back to fantasy/generic content instead of generating post-apocalyptic content. Specifically: (a) `pkg/procgen/terrain/composite.go:227` — `genrePrefs["postapocalyptic"]` is never hit; composite terrain falls back to generic `{"bsp", "cellular", "maze", "forest", "city"}`; (b) `pkg/procgen/skills/generator.go:129` — `normalizeGenre("postapoc")` falls to the default case and returns `"fantasy"`, so all post-apocalyptic skill trees are generated as fantasy trees; (c) `pkg/procgen/class/generator.go:389-394` — six `addGenreTheme("postapocalyptic", …)` calls are never matched for `"postapoc"`, so character classes spawn with fantasy names instead of Raider/Scavenger/Mutant etc.; (d) `cmd/client/util.go:621` and `:748` — environment hazard pools and spawn-prop tables keyed by `"postapocalyptic"` are not populated; rooms spawn with fantasy props; (e) `pkg/rendering/patterns/generator.go:118` — pattern case `"postapocalyptic"` is unreachable. — **Blocked goal**: Genre-based theming for the post-apocalyptic genre (one of five stated genres). — **Remediation**: In each of the five files, add `"postapoc"` as an alias or replace `"postapocalyptic"` with `"postapoc"`. (1) `pkg/procgen/terrain/composite.go:227` — change map key. (2) `pkg/procgen/skills/generator.go:129` — add `"postapoc"` to the switch case beside `"postapocalyptic"`, or normalise the key before the switch. (3) `pkg/procgen/class/generator.go:389-394` — change all six `addGenreTheme` calls to use `"postapoc"`. (4) `cmd/client/util.go:621,748` — change map keys. (5) `pkg/rendering/patterns/generator.go:118` — change the case label. Add cross-genre coverage tests that pass `"postapoc"` (not `"postapocalyptic"`) and assert non-default output. Validate with `go test ./pkg/procgen/... ./cmd/client/...`.

---

### MEDIUM

- [ ] **N2 — Client-side V9 integration managers instantiated but not wired into consuming systems** — `cmd/client/init_versions.go:469-479` — `initializeV9SystemsClient()` creates three managers: `sys.stationManager` (housing crafting), `sys.petHomeManager` (companion housing), and `sys.guildHousingManager` (guild housing). The dedicated server correctly wires these at `cmd/server/main.go:432` (`craftingSystem.SetStationManager(stationMgr)`) and `cmd/server/main.go:438` (`companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)`) and sets up a `V9ValidationService`. No equivalent wiring exists in the client path. Consequently, in solo-play mode (client auto-starts an embedded server), housing-based crafting bonuses and companion loyalty bonuses are silently zero, even when the player owns a house with crafting stations or companion bedding. — **Blocked goal**: Player housing benefits in solo play. — **Remediation**: After the crafting and companion loyalty systems are created in the client init path, add the two missing calls: `sys.craftingSystem.SetStationManager(sys.stationManager)` (immediately after `sys.craftingSystem` is constructed at `cmd/client/handlers.go:1036`) and `sys.companionLoyaltySys.SetPetHomeProvider(sys.petHomeManager)` (after `sys.companionLoyaltySys` is constructed at `cmd/client/init_versions.go:63`). For `guildHousingManager`, either create a client-local `V9ValidationService` or expose the housing permission check directly. Validate with `go build ./cmd/client/ && go test ./pkg/integration/housing_crafting/ ./pkg/integration/companion_housing/`.

- [ ] **N3 — `sys.choiceTracker` dead object: duplicate of `choiceConsequencesSystem`'s internal tracker** — `cmd/client/init_versions.go:523` — `sys.choiceTracker = choice_consequences.NewChoiceTracker()` creates a standalone `ChoiceTracker` that is never accessed after this line. The active tracker is the one created internally by `engine.NewChoiceConsequencesSystem()` at `pkg/engine/choice_consequences_system.go:21`. Any code calling through `sys.choiceConsequencesSystem.RecordChoice()` writes to the system's internal tracker; `sys.choiceTracker` is permanently empty and orphaned. — **Blocked goal**: None directly blocked, but the duplicate creates confusion and maintenance debt (future developers may attempt to wire systems to the wrong tracker). — **Remediation**: Remove `sys.choiceTracker` from `systemsContainer` and delete the initialization at `cmd/client/init_versions.go:523`. If external code needs direct tracker access, expose `sys.choiceConsequencesSystem.GetTracker()` (add this accessor if not present). Validate with `go build ./cmd/client/ && go vet ./cmd/client/`.

- [ ] **N4 — `sys.guildFleetManager` dead object: duplicate of `guildVehicleSystem`'s internal manager** — `cmd/client/init_versions.go:527` — `sys.guildFleetManager = guild_vehicle.NewFleetManager()` creates a standalone `FleetManager` that is never accessed after this line. The active manager is the one owned by `sys.guildVehicleSystem`, created inside `engine.NewGuildVehicleSystem()` at `pkg/engine/guild_vehicle_system.go:20-26`. `GuildVehicleSystem` already exposes `GetFleetManager()` for external access. The standalone `sys.guildFleetManager` has no fleets, no vehicles, and no membership validator; it is permanently empty. — **Blocked goal**: None directly blocked, but any code that accidentally wires against `sys.guildFleetManager` instead of `sys.guildVehicleSystem.GetFleetManager()` will observe empty state. — **Remediation**: Remove `sys.guildFleetManager` from `systemsContainer` and delete the initialization at `cmd/client/init_versions.go:527`. Update any references to use `sys.guildVehicleSystem.GetFleetManager()`. Validate with `go build ./cmd/client/ && go vet ./cmd/client/`.

- [ ] **N5 — Merchant entities do not receive `GenreComponent`, breaking genre-aware quest reward scaling for merchants** — `pkg/engine/merchant_spawn.go` (entire file) — `SpawnMerchantFromData()` assembles a merchant entity with position, sprite, health, collider, merchant, dialog, and NPC dialog components but never calls `entity.AddComponent(NewGenreComponent(...))`. Enemy entities correctly receive a `GenreComponent` via `addAdvancedComponents()` at `pkg/engine/entity_spawning.go:273-275`. The `objectiveTracker` at `pkg/engine/objective_tracker_system.go:761` reads `entity.GetComponent("genre")` for any entity type when computing quest reward scaling; for merchant entities this call always returns nil and the genre-scaling branch is dead. — **Blocked goal**: Consistent genre-based reward scaling for quest objectives involving merchants. — **Remediation**: Near the end of `SpawnMerchantFromData()` in `pkg/engine/merchant_spawn.go`, add `merchant.AddComponent(NewGenreComponent(params.GenreID))`. Use the `params.GenreID` value already available at the call site. Validate with `go test ./pkg/engine/`.

---

### LOW

- [ ] **N6 — `NewBlendedGenreComponent` defined and exported but has zero callers** — `pkg/engine/genre_component.go:36` — `NewBlendedGenreComponent(primaryGenre string, secondaryGenres []string, blendRatio float64)` is exported and fully implemented but is never called in `cmd/`, `pkg/`, or any example. The multi-genre blending data model exists (fields `SecondaryGenres []string` and `BlendRatio float64` on `GenreComponent`), but the feature is never activated at runtime. — **Blocked goal**: None at this scope; planned multi-genre worlds (e.g., horror + cyberpunk blend) are not yet connected. — **Remediation**: Either (a) wire `NewBlendedGenreComponent` where multi-genre `GenerationParams` are detected (check `params.Custom["secondary_genres"]` in spawn helpers), or (b) lower-case to `newBlendedGenreComponent` to signal it is internal API awaiting future integration. Validate with `go build ./... && go vet ./...`.

- [ ] **N7 — `tradeRouteManager.Stop()` never called on client shutdown** — `cmd/client/init_versions.go:644-645` — `sys.tradeRouteManager.Start()` launches a background goroutine (via `sync.Once`) with a `10 * time.Second` ticker. The goroutine runs until `Stop()` is called. No `Stop()` call exists in the client's shutdown or cleanup path (`cmd/client/handlers.go:3576-3602`). On clean exit this goroutine is leaked; on test runs with `t.Cleanup` it can cause `go test -race` false positives. The server correctly stops the manager (per `cmd/server/main.go:468`). — **Blocked goal**: None at stated goal level; this is a resource management issue. — **Remediation**: Do not add `defer sys.tradeRouteManager.Stop()` inside the short-lived initialization helper that calls `Start()`, because that would run as soon as setup returns and stop trade routes during gameplay. Instead, stop the manager from a long-lived shutdown scope (for example, defer it from `main()` after the full client setup succeeds) or register `sys.tradeRouteManager.Stop` in the explicit client cleanup/shutdown hook, such as the cleanup function returned by `startEmbeddedServer()`. `Stop()` is idempotent. Validate with `go test -race ./cmd/client/... ./pkg/integration/trade_routes/`.

- [ ] **N8 — OpenXR VR adapters remain documented stubs** — `pkg/engine/vr_openxr_adapters.go:56-230` (build tag `//go:build vr && !js`) — 16 `TODO(vr-sdk):` markers describe the OpenXR SDK calls needed. All methods return zero values or mock responses. `cmd/client/init_versions.go:582,594` unconditionally uses `NewStubHeadsetAdapter`/`NewStubControllerAdapter` regardless of whether the `vr` build tag is set. README explicitly states "experimental with mock adapters only; no hardware SDK integration." — **Blocked goal**: None at stated goal level — README marks VR as explicitly experimental. — **Remediation**: Tracked in ROADMAP.md Priority 4. When a stable Go OpenXR binding becomes available, implement each `TODO(vr-sdk)` stub per the inline pseudocode and switch `init_versions.go:582,594` to construct OpenXR adapters when `--vr` is set without `--force-stub`. No action required until SDK is available.

---

## Re-verification of Prior GAPS.md Findings

All gaps from the prior `GAPS.md` (dated 2026-04-21) were re-verified against the current HEAD. Status:

| Prior Gap | Prior Severity | Current Status | Evidence |
|-----------|---------------|----------------|----------|
| G1 — Story generator `"postapocalyptic"` key | HIGH | **CLOSED** | `generator.go:21,30` and all `timeline.go` case labels now use `"postapoc"`. Story tests updated. |
| G2 — `AnimationSyncManager` receive path unwired | HIGH | **CLOSED** | `pkg/network/client.go:726` calls `mgr.BufferState(pkt)`; `pkg/engine/animation_system.go:853` calls `DrainRemoteState`. Both paths fully wired. |
| G3 — Guild vehicle formation physics + `FleetID` sync | MEDIUM | **CLOSED** | `GuildVehicleSystem.Update()` (engine/guild_vehicle_system.go:35–105) runs two-pass formation steering; `SetVehicleSyncer(sys)` wired at construction. |
| G4 — `ArchaeologyGenerator` etc. never called | MEDIUM | **CLOSED** | `cmd/client/util.go:2474,2505,2535` now calls all three generators. |
| G5 — `StoryJournalUI` never instantiated | MEDIUM | **CLOSED** | `cmd/client/handlers.go:3371-3372` instantiates `NewStoryJournalUI`; keybind toggle and draw path wired. |
| G6 — `GenreComponent` only on player | MEDIUM | **PARTIALLY CLOSED** | Enemies now get `GenreComponent` via `entity_spawning.go:275`. Merchants still excluded → see **N5** above. |
| G7 — `equipment_durability_particle_system.go` (`//go:build ignore`) | LOW | **CLOSED** | File deleted; specialised variants fully cover the use case. |
| G8 — `InteractWithEnvironmentNode` orphan | LOW | **CLOSED** | `behavior_tree_archetypes.go:222` now composes the node in crafter/gatherer behavior trees. |
| G9 — `NewHelpSystemWithSize` effectively private API | LOW | **CLOSED** | Renamed to unexported `newHelpSystemWithSize` (`help_system.go:53`). |
| G10 — OpenXR VR stubs | LOW | **OPEN → N8** | Unchanged; ROADMAP Priority 4. |

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `memprofile.PrintProfile` uses `fmt.Printf` | Intentional CLI/debug output per comment at `profile.go:195`. Correct channel for a profiling utility. |
| `NewMockHeadset`/`NewMockController` in VR files | Production fallback path for offline/CI use; intentionally used by `init_versions.go:582,594`. Not dead code. |
| `choiceConsequencesSystem` ECS system fields | Fully registered at `handlers.go:2282` and wired. Not a gap. |
| `guildVehicleSystem` ECS system | Registered at `handlers.go:2283`; internal FleetManager is active. Not a gap. |
| `tradeRouteManager` on server (Start + AddSystem) | The wrapper calls `UpdateRoutes()` on every ECS tick _in addition_ to the background timer goroutine. Client lacks the wrapper; its routes are updated only on the 10s timer. The gap is the missing `Stop()`, not the missing wrapper (background timer is sufficient for solo play). |
| `RouteManager.SetPriceUpdateHandler` not called on client | The federated market pricing system (`worldEconomySystem`) is registered via `AddSystem` on the client. Cross-checking prices between solo-play caravans and the local economy is not a stated goal for the single-player path. |
| Interfaces with single implementations | The ECS component / adapter interface pattern intentionally uses one real + one mock implementation. Not a gap. |
| Terrain `genre_mapping.go:128` uses `"postapoc"` | The genre_mapping file is correct. Only `composite.go`, `grammar.go` (which accepts both), and `lsystem.go` (which accepts both) deviate. `grammar.go` and `lsystem.go` accept multi-alias cases so they are not broken. Only `composite.go:227` is a pure single-key map that silently misses `"postapoc"`. |
| `pkg/procgen/terrain/grammar.go:250` and `lsystem.go:476` | Both use compound case labels (`"post-apocalyptic", "postapocalyptic"`) that cover multiple spellings. They do NOT include `"postapoc"`, but the grammar generator is invoked via `CompositeGenerator.selectGenerators`, which itself falls back for `"postapoc"`. The root miss is `composite.go:227`. Flagged as part of N1 rather than separate findings. |
| `sys.stationManager` fields not in `AddSystem` | `StationManager` is not an ECS system; it is a service object injected into ECS systems. Not being added to `World.AddSystem()` is correct by design. The gap is the missing injection call. |
| `pkg/procgen/dialog/corpus.go:29` accepts both keys | `case "postapoc", "postapocalyptic"`: — this file correctly handles both spellings. Not a gap. |
