# Implementation Gaps — 2026-04-21

> This file supersedes the previous `GAPS.md`.  
> Gaps are ordered: open HIGH → open MEDIUM → open LOW.  
> Gaps from the prior file that are now closed appear in the **Resolved** section at the bottom.

---

## G1 — Story Fragment Generator: `"postapoc"` genre key produces default themes

- **Intended Behavior**: When the game is launched with `-genre=postapoc` (the canonical post-apocalyptic genre ID used by the CLI, `pkg/procgen/genre/predefined.go:86`, quest templates, and terrain generators), `story.FragmentGenerator` should generate post-apocalyptic story themes ("last_survivors", "resource_war", etc.) and matching title suffixes ("Last Stand", "Wasteland", etc.).
- **Current State**: `genreThemes` and `genreTitleSuffixes` in `pkg/procgen/story/generator.go:16-31` use `"postapocalyptic"` as the map key. `getThemesForGenre("postapoc")` at line 335 does not match; the `ok` check fails; a `Warn` log fires; and `defaultThemes` (fantasy/generic) are returned. The unit tests in `generator_test.go:275` also use `"postapocalyptic"`, so the mismatch is not caught by the test suite. `timeline.go` case labels (`:240`, `:361`, `:378`, `:395`, `:412`, `:429`, `:446`, `:463`, `:480`, `:532`, `:549`, `:577`) also use `"postapocalyptic"`.
- **Blocked Goal**: Genre-specific story content for the post-apocalyptic theme; correct lore displayed to players choosing the `postapoc` genre.
- **Implementation Path**:
  1. In `pkg/procgen/story/generator.go`, change both map-literal keys from `"postapocalyptic"` to `"postapoc"` (lines 21 and 30).
  2. In `pkg/procgen/story/timeline.go`, change all `case "postapocalyptic":` labels (10 locations) to `case "postapoc":`.
  3. In `pkg/procgen/story/generator_test.go:275`, `archaeology_test.go:238`, `crossdungeon_test.go:265`, change `"postapocalyptic"` to `"postapoc"`.
  4. Run `go test ./pkg/procgen/story/...` and `go test ./pkg/procgen/...` to confirm all genre-coverage tests pass.
- **Dependencies**: None.
- **Effort**: Small (< 1 hour).
- **Severity**: HIGH

---

## G2 — `AnimationSyncManager`: receive/playback path (`BufferState`/`GetNextState`) has zero callers

- **Intended Behavior**: In multiplayer mode, received animation-state packets from remote players should be buffered with jitter compensation and played back at a steady rate so remote-player animations are smooth under 200–5000 ms latency.
- **Current State**: The send side is fully wired (`cmd/client/handlers.go:676-677` creates the manager; `pkg/engine/animation_system.go:823-828` gates outgoing sends through `ShouldSync`/`RecordSync`). But `AnimationSyncManager.BufferState` (`pkg/network/animation_sync.go:234`) and `GetNextState` (`:251`) have zero callers outside the test suite. Received animation packets are decoded and discarded without entering the jitter buffer.
- **Blocked Goal**: "High-latency multiplayer … smooth remote-player animations" — core multiplayer experience.
- **Implementation Path**:
  1. In `pkg/network/client.go`, locate the `StateUpdate` handler that routes `ComponentData` to local entity components. When `ComponentData.Type == "animation_state"` (or whatever constant is used for animation state payloads), deserialise the packet into `AnimationStatePacket` and call `animSyncMgr.BufferState(packet)`.
  2. In `pkg/engine/animation_system.go`, add a `drainRemoteBuffer` helper: for each remote entity (those without a local `InputComponent`), call `syncMgr.GetNextState(entityID)` per `Update` tick; if non-nil, apply the returned state to `AnimationComponent.CurrentState`.
  3. Pass the manager reference to the client-side systems; since `animSyncMgr` is created in `cmd/client/handlers.go`, thread it through or expose via `AnimationSystem.GetSyncManager()`.
  4. Add `TestAnimationSyncReceivePath` exercising buffer→drain roundtrip. Run `go test ./pkg/network/ ./pkg/engine/`.
- **Dependencies**: Requires identifying the correct `ComponentData.Type` constant for animation state (search `pkg/network/` for animation payload encoding).
- **Effort**: Medium (2–4 hours).
- **Severity**: HIGH

---

## G3 — `guild_vehicle` integration: formation physics and `VehicleComponent.FleetID` sync not wired

- **Intended Behavior**: When a player is in a guild fleet, their vehicle should (a) adopt a formation position relative to the fleet leader each physics tick and (b) have `VehicleComponent.FleetID` set to the guild fleet's ID so other systems can identify fleet membership.
- **Current State**: `pkg/integration/guild_vehicle/integrations.go:57-116` defines `SetVehicleSyncer` (`:74`) and `GetFormationOffsets` (`:102`) but neither has external callers that feed results into `pkg/engine/physics/vehicle/`. Guild membership validation (`SetMembershipValidator`) is wired per prior audit. Fleet siege damage (`ApplySiegeDamage`) is implemented but `SetStructureDamager` caller binding is incomplete in `cmd/server/main.go`. `VehicleComponent.FleetID` remains empty for all vehicles.
- **Blocked Goal**: Guild vehicle fleets operating in formation; guild siege mechanics via vehicle weaponry.
- **Implementation Path**:
  1. Create a `vehicleSyncer` adapter (implements `guild_vehicle.VehicleSyncer`) that writes `FleetID` onto `VehicleComponent` when `AddVehicle`/`RemoveVehicle` are called.
  2. After fleet manager construction in `cmd/server/main.go`, call `fleetManager.SetVehicleSyncer(vehicleSyncer)`.
  3. In `pkg/engine/physics/vehicle/system.go`, add an `updateFormation` step: query `FleetManager.GetFormationOffsets(fleetID)` for entities whose `VehicleComponent.FleetID != ""` and apply the offset as target position.
  4. Wire `SetStructureDamager` with the territory manager (compare existing wiring in `cmd/server/main.go` post `initializeTerritorySystemsServer`).
  5. Validate with `go test ./pkg/integration/guild_vehicle/ ./pkg/engine/physics/vehicle/`.
- **Dependencies**: G-vehicle physics system must be running (it is, registered in `system_init.go`).
- **Effort**: Medium (3–5 hours).
- **Severity**: MEDIUM

---

## G4 — `ArchaeologyGenerator`, `TimelineGenerator`, `CrossDungeonGenerator` implemented but not in game loop

- **Intended Behavior**: Players should be able to discover archaeological excavation sites, historical timelines, and multi-dungeon story arcs as part of world exploration, complementing the already-wired `FragmentGenerator` output.
- **Current State**: Three fully tested generators exist in `pkg/procgen/story/` (archaeology: 301 LOC, timeline: 580 LOC, crossdungeon: 320 LOC) with 88.7% test coverage and comprehensive documentation. None are called from `cmd/client/util.go` or any spawner. Explicitly deferred in `pkg/procgen/story/AUDIT.md` as "Phase 2 integration sprint."
- **Blocked Goal**: "100% procedural content … story depth / environmental narrative" (README).
- **Implementation Path**:
  1. Define ECS components: `ArchaeologicalSiteComponent`, `TimelineComponent`, `CrossDungeonStoryComponent` in `pkg/engine/`, following the structure of `StoryFragmentComponent`.
  2. Add spawner functions in `cmd/client/util.go`: `spawnArchaeologySites`, `spawnTimelines`, `spawnCrossDungeonArcs`. Call them from `spawnWorldEntities` in `cmd/client/init_spawning.go`.
  3. Add discovery-system logic to handle the three new component types (analogous to `StoryFragmentComponent` handling).
  4. Implement `Serialize`/`Deserialize` so excavation progress persists across sessions.
  5. Validate with `go test ./pkg/procgen/story/ ./pkg/engine/`.
- **Dependencies**: G5 (StoryJournalUI) should be completed first so discovered content is visible.
- **Effort**: Large (2–4 days).
- **Severity**: MEDIUM

---

## G5 — `StoryJournalUI` fully implemented but never instantiated

- **Intended Behavior**: Players should be able to open a journal (keybind e.g. `L`) to review story fragments, series completion state, and discovered narrative content.
- **Current State**: `pkg/rendering/ui/story_journal.go` contains a complete, genre-themed journal UI (`StoryJournalUI`, 230 LOC) with full navigation, series grouping, and input handling. `NewStoryJournalUI` has zero callers in `cmd/`. `DiscoverySystem` populates `StoryJournalComponent` but there is no in-game path to render it. Explicitly deferred in `pkg/procgen/story/AUDIT.md`.
- **Blocked Goal**: Players experiencing story content they have discovered during exploration.
- **Implementation Path**:
  1. In `cmd/client/handlers.go`, add a `storyJournalUI *ui.StoryJournalUI` field to `systemsContainer`.
  2. In the lazy init phase (`scheduleLazyInit` or group 3 rendering), instantiate with `ui.NewStoryJournalUI(x, y, w, h, *genreID)`.
  3. Bind a keybind (suggest `ebiten.KeyL`) in the input-handler to toggle journal visibility.
  4. In the `Draw` path, call `storyJournalUI.Draw(screen)` when visible.
  5. On journal open, call `storyJournalUI.LoadFromJournal(playerJournal, game.World)`.
  6. Validate with `go build ./cmd/client/`.
- **Dependencies**: None; the journal UI is self-contained.
- **Effort**: Small (2–3 hours).
- **Severity**: MEDIUM

---

## G6 — `GenreComponent` attached only to player; NPC entities carry none

- **Intended Behavior**: Per `pkg/engine/objective_tracker_system.go:761`, quest-reward scaling reads `entity.GetComponent("genre")` for any entity. Genre-specific NPC/enemy theming also relies on `GenreComponent` being present.
- **Current State**: `cmd/client/handlers.go:2753` attaches `NewGenreComponent(*genreID)` to the player entity. All NPC and enemy entities spawned by `SpawnEnemiesInTerrain`, `SpawnMerchantsInTerrain`, etc. receive no `GenreComponent`, so the `entity.GetComponent("genre")` call in the objective tracker always returns nil for them, and the genre-aware scaling branch is unreachable. `NewBlendedGenreComponent` has zero callers.
- **Blocked Goal**: Consistent genre theming and reward scaling for all entities.
- **Implementation Path**:
  1. In the NPC/enemy spawn helper (pkg/engine area called by `SpawnEnemiesInTerrain`), add `entity.AddComponent(engine.NewGenreComponent(genreID))` using the `params.GenreID` already available at spawn time.
  2. For multi-genre worlds (when `params.Custom["secondary_genres"]` is non-nil), use `NewBlendedGenreComponent` instead.
  3. Validate with `go test ./pkg/engine/` (cover the objective tracker reward-scaling path).
- **Dependencies**: None.
- **Effort**: Small (1–2 hours).
- **Severity**: MEDIUM

---

## G7 — `pkg/engine/equipment_durability_particle_system.go` excluded with `//go:build ignore`

- **Intended Behavior**: The file was the base class for equipment-degradation particles; tests and implementation exist.
- **Current State**: Both `equipment_durability_particle_system.go` and `equipment_durability_particle_system_test.go` carry `//go:build ignore` at line 1. The four specialised variants (terrain, combat, reputation, weather-sheen) are registered in `system_init.go` and fully tested. The base file's 279 LOC and its test are never compiled, dead code that still consumes reader attention.
- **Blocked Goal**: None active (maintenance debt only).
- **Implementation Path**: Delete `pkg/engine/equipment_durability_particle_system.go` and `pkg/engine/equipment_durability_particle_system_test.go`. If a future base class is desired, re-derive from the four specialised variants when the pattern is proven. Run `go build ./... && go vet ./...`.
- **Dependencies**: None.
- **Effort**: Trivial (< 15 minutes).
- **Severity**: LOW

---

## G8 — `NewInteractWithEnvironmentNode` never composed into any behavior tree

- **Intended Behavior**: NPCs with crafter, gatherer, or interactive-object-user roles should have a behavior-tree branch that triggers the `InteractWithEnvironmentNode` during world traversal.
- **Current State**: `pkg/engine/behavior_tree_advanced_nodes.go:292-315` provides a complete implementation (`Execute`, tick, logging, cooldown). No behavior-tree builder in `pkg/engine/` or `cmd/` includes this node. The `GatheringSystem` and `CraftingSystem` handle similar logic through direct component polling rather than via the behavior tree, making this node unreachable from any NPC archetype.
- **Blocked Goal**: Data-driven NPC behavior depth through composable behavior trees.
- **Implementation Path**:
  1. Identify NPC behavior-tree build sites (search `pkg/engine/ai_system.go` or `pkg/engine/npc_*.go` for `BehaviorTree{` / `NewSelectorNode`).
  2. For crafter/gatherer archetypes, add an `InteractWithEnvironmentNode` child under a `HasNearbyInteractable` condition node.
  3. Add a unit test asserting the branch is reachable and `Execute` returns `Running` when a resource node is nearby.
  4. Validate with `go test ./pkg/engine/`.
- **Dependencies**: None.
- **Effort**: Small (2–3 hours).
- **Severity**: LOW

---

## G9 — `NewHelpSystemWithSize` exported but has no external callers

- **Intended Behavior**: Callers should be able to pass custom screen dimensions to the help system for non-standard display sizes.
- **Current State**: `pkg/engine/help_system.go:50` exports `NewHelpSystemWithSize(width, height int)` with zero callers. `NewHelpSystem` delegates to it with hardcoded `800, 600`. The parameterised variant is thus effectively dead public API surface.
- **Blocked Goal**: None active (API clarity only).
- **Implementation Path**: Either (a) make the function unexported (`newHelpSystemWithSize`) and update the delegation call, or (b) pass real screen dimensions from the display system (`cmd/client/main.go` has `screenWidth`/`screenHeight` flags) so the parameter has observable effect. Validate with `go vet ./pkg/engine/`.
- **Dependencies**: None.
- **Effort**: Trivial.
- **Severity**: LOW

---

## G10 — OpenXR VR adapter methods are documented stubs (16 `TODO(vr-sdk)`)

- **Intended Behavior**: When compiled with `-tags vr`, `OpenXRHeadsetAdapter` and `OpenXRControllerAdapter` should call the OpenXR SDK for head pose, controller state, haptics, etc.
- **Current State**: `pkg/engine/vr_openxr_adapters.go:56-230` contains 16 `TODO(vr-sdk):` markers describing required xrLocateViews, xrCreateInstance, xrGetActionStateFloat, xrApplyHapticFeedback calls. All methods return zero values. `cmd/client/init_versions.go:583,595` uses the stub adapters unconditionally regardless of whether the `vr` build tag is set. Explicitly marked "experimental VR/stereoscopic support … mock adapters only" in the README.
- **Blocked Goal**: Real VR headset / controller support.
- **Implementation Path**: When a Go OpenXR binding (e.g., `github.com/nicholasgasior/go-openxr` or cgo wrapper) is available: implement each TODO per the inline pseudocode, add CGo imports, switch `init_versions.go:583,595` to OpenXR constructors when `--vr` flag is set. Until then, no action required.
- **Dependencies**: External Go OpenXR SDK availability (ROADMAP Priority 4).
- **Effort**: Large (tracked in ROADMAP).
- **Severity**: LOW

---

## Resolved Gaps (closed since prior GAPS.md)

| ID | Description | Resolved Evidence |
|----|-------------|-------------------|
| Prior Gap 2 | `SkillPointGainParticleSystem` not registered | `system_init.go:1499-1503` instantiates; `AddSkillPointCallback` wires `OnSkillPointGain` |
| Prior Gap 3 | `GenreComponent` never attached | `handlers.go:2753`; NPC side → see G6 (still open) |
| Prior Gap 4 | `SpellComponent` carryover unpopulated | `spell_casting.go:3061-3068` creates and populates |
| Prior Gap 5 | Chunk compression bytes discarded on eviction | `cmd/server/main.go:541-546` calls `worldPersistence.SaveChunk`; `pkg/world/persistence.go:427-454` implements the method |
| Prior Gap 7 | Three alternate constructors were orphaned | `NewAnimationAdapter`, `NewLightingAdapter`, `NewTimer` now delegate to their with-options variants |
| Prior Gap 9 (iOS) | iOS keyboard `IsKeyboardSupported` returned false | `keyboard_ios.go:111-112` returns `true` |
| Prior Gap 10 | `NewBaseStatsFromEntity` had zero callers | `handlers.go:2746` calls it |
