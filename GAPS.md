# Implementation Gaps — 2026-04-21

> This file supersedes the previous `GAPS.md`.  
> All prior gaps that are now **closed** appear in the Resolved section at the bottom.  
> Open gaps are ordered: HIGH → MEDIUM → LOW.

---

## N1 — `"postapoc"` genre key not recognised by five packages

- **Intended Behavior**: When the game is launched with `-genre=postapoc` (the canonical post-apocalyptic genre ID from `pkg/procgen/genre/predefined.go:86` and the CLI `--genre` flag), all content generators should produce post-apocalyptic themed output: post-apocalyptic skill trees, class names (Raider, Scavenger, Mutant, …), terrain composition biased toward cellular/city/forest ruin patterns, environment hazards including fire pits, acid pools, and poison gas, and procedural patterns with a weathered/ruined aesthetic.
- **Current State**: Five packages use `"postapocalyptic"` (a key that is never passed by the runtime) rather than the canonical `"postapoc"`:
  1. `pkg/procgen/terrain/composite.go:227` — `genrePrefs` map has a `"postapocalyptic"` key; `"postapoc"` is not present; the map look-up fails silently and generic terrain composition is used instead.
  2. `pkg/procgen/skills/generator.go:129` — `normalizeGenre()` has a switch case for `"postapocalyptic"` but not `"postapoc"`; the function returns `"fantasy"` for the canonical ID, so post-apocalyptic players receive fantasy skill trees.
  3. `pkg/procgen/class/generator.go:389-394` — six `addGenreTheme("postapocalyptic", …)` calls register themed class names (Raider, Scavenger, Mutant, Healer, Outrider, Protector) under a key that never matches; players receive generic class names.
  4. `cmd/client/util.go:621,748` — two map literals key hazard pools and prop tables under `"postapocalyptic"`; the lookup always misses and falls back to the fallback (typically fantasy).
  5. `pkg/rendering/patterns/generator.go:118` — `case "postapocalyptic":` is unreachable; post-apocalyptic pattern generation is never triggered.
- **Blocked Goal**: Genre-based theming for the post-apocalyptic genre — one of five explicitly stated genres in the README.
- **Implementation Path**:
  1. `pkg/procgen/terrain/composite.go:227` — change the map key from `"postapocalyptic"` to `"postapoc"`.
  2. `pkg/procgen/skills/generator.go:129` — add `"postapoc"` to the switch case alongside `"postapocalyptic"`: `case "scifi", "fantasy", "horror", "cyberpunk", "postapocalyptic", "postapoc":`. Also add `"postapoc"` to the case at line 105.
  3. `pkg/procgen/class/generator.go:389-394` — change all six `addGenreTheme("postapocalyptic", …)` calls to `"postapoc"`.
  4. `cmd/client/util.go:621,748` — change the two `"postapocalyptic"` map keys to `"postapoc"`.
  5. `pkg/rendering/patterns/generator.go:118` — change `case "postapocalyptic":` to `case "postapoc":`.
  6. Add regression tests in each package that pass `"postapoc"` (not `"postapocalyptic"`) and assert that non-default (non-fantasy) output is produced.
- **Dependencies**: None; each fix is independent.
- **Effort**: Small (< 2 hours total across all five files).
- **Severity**: HIGH

---

## N2 — Client-side V9 integration managers not wired into consuming systems

- **Intended Behavior**: In solo-play mode (the client auto-starts an embedded server), the `CraftingSystem` should apply housing-station bonuses when the player crafts at a housed workbench, and `CompanionLoyaltySystem` should apply pet-home comfort bonuses to companion loyalty. These bonuses are intended to reward players who invest in player housing.
- **Current State**: `cmd/client/init_versions.go:469-479` (the `initializeV9SystemsClient` function) creates all three V9 managers (`sys.stationManager`, `sys.petHomeManager`, `sys.guildHousingManager`) but never calls the two injection setters that connect them to the ECS systems:
  - `craftingSystem.SetStationManager(sys.stationManager)` — called on the server at `cmd/server/main.go:432` but missing from the client init path.
  - `companionLoyaltySys.SetPetHomeProvider(sys.petHomeManager)` — called on the server at `cmd/server/main.go:438` but missing from the client init path.
  - `sys.guildHousingManager` is created but never passed to any validation or permission service on the client, so guild-housing permission checks in solo play always resolve against an empty manager.
- **Blocked Goal**: Player housing benefits in solo-play mode; parity between hosted-server play and solo-play for housing-based bonuses.
- **Implementation Path**:
  1. In `cmd/client/handlers.go`, after `sys.craftingSystem` is constructed at line 1036, add: `sys.craftingSystem.SetStationManager(sys.stationManager)`.
  2. In `cmd/client/init_versions.go`, after `sys.companionLoyaltySys` is constructed at line 63, add: `sys.companionLoyaltySys.SetPetHomeProvider(sys.petHomeManager)`.
  3. For `guildHousingManager`: decide whether to create a client-local `V9ValidationService` (mirroring `cmd/server/main.go:419`) or expose housing permission checks through the existing guild-housing UI path.
  4. Add integration tests asserting that a client-side world with a registered crafting station returns non-zero crafting bonuses.
- **Dependencies**: None; all managers and setter methods already exist.
- **Effort**: Small (< 1 hour for items 1 and 2; medium for item 3).
- **Severity**: MEDIUM

---

## N3 — `sys.choiceTracker` is a dead duplicate of `choiceConsequencesSystem`'s internal tracker

- **Intended Behavior**: Choice consequences (story-branching decisions, NPC attitude shifts, class-quest gating) should be recorded to a single, authoritative `ChoiceTracker` that is queried by all systems that need to test choice history.
- **Current State**: Two separate `ChoiceTracker` instances exist on the client:
  1. `sys.choiceConsequencesSystem.tracker` — created inside `engine.NewChoiceConsequencesSystem()` at `pkg/engine/choice_consequences_system.go:21`. This is the active tracker. It is populated by `RecordChoice()` calls on the system, which is registered with `game.World` at `cmd/client/handlers.go:2282`.
  2. `sys.choiceTracker` — created at `cmd/client/init_versions.go:523` as a standalone `choice_consequences.NewChoiceTracker()`. It is never accessed again after this line; no choices are recorded to it; no system reads from it.
  The duplicate creates confusion: a future developer writing code against `sys.choiceTracker` will observe an always-empty tracker and may incorrectly conclude the feature is non-functional.
- **Blocked Goal**: No current feature is blocked, but the duplicate creates maintenance debt and potential correctness bugs if code is added that writes to the wrong tracker.
- **Implementation Path**:
  1. Delete the `choiceTracker *choice_consequences.ChoiceTracker` field from `systemsContainer` in `cmd/client/handlers.go:584`.
  2. Delete the initialisation at `cmd/client/init_versions.go:523`.
  3. If direct tracker access is needed externally, add `func (s *ChoiceConsequencesSystem) GetTracker() *choice_consequences.ChoiceTracker { return s.tracker }` to `pkg/engine/choice_consequences_system.go` and use `sys.choiceConsequencesSystem.GetTracker()` at call sites.
  4. Validate with `go build ./cmd/client/ && go vet ./cmd/client/`.
- **Dependencies**: None.
- **Effort**: Small (< 30 minutes).
- **Severity**: MEDIUM

---

## N4 — `sys.guildFleetManager` is a dead duplicate of `guildVehicleSystem`'s internal manager

- **Intended Behavior**: Guild vehicle fleet operations (adding vehicles, setting formations, granting access) should flow through a single, authoritative `FleetManager` that is owned by and in sync with `GuildVehicleSystem`.
- **Current State**: Two separate `FleetManager` instances exist on the client:
  1. `sys.guildVehicleSystem.manager` — created inside `engine.NewGuildVehicleSystem()` at `pkg/engine/guild_vehicle_system.go:20`. This is the active manager. It is wired as a `VehicleSyncer`, registered in the ECS world, and processes formation each frame. Exposed via `sys.guildVehicleSystem.GetFleetManager()`.
  2. `sys.guildFleetManager` — created at `cmd/client/init_versions.go:527` as a standalone `guild_vehicle.NewFleetManager()`. It is never accessed again after this line; no vehicles are ever added to it; no membership validator is injected; no system reads from it.
- **Blocked Goal**: No current feature is blocked, but the duplicate creates the same confusion and correctness risk as N3.
- **Implementation Path**:
  1. Delete the `guildFleetManager *guild_vehicle.FleetManager` field from `systemsContainer` in `cmd/client/handlers.go:585`.
  2. Delete the initialisation at `cmd/client/init_versions.go:527`.
  3. At any future call site needing direct fleet-manager access, use `sys.guildVehicleSystem.GetFleetManager()`.
  4. Validate with `go build ./cmd/client/ && go vet ./cmd/client/`.
- **Dependencies**: None.
- **Effort**: Small (< 30 minutes).
- **Severity**: MEDIUM

---

## N5 — Merchant entities do not receive `GenreComponent`

- **Intended Behavior**: All entities in the world should carry a `GenreComponent` so that systems reading `entity.GetComponent("genre")` — notably the `ObjectiveTrackerSystem` quest reward scaler — can compute genre-appropriate bonuses regardless of entity type.
- **Current State**: Enemy entities receive a `GenreComponent` via `addAdvancedComponents()` at `pkg/engine/entity_spawning.go:273-275`. Merchant entities created by `SpawnMerchantFromData()` in `pkg/engine/merchant_spawn.go` do not; the function adds position, velocity, health, team, sprite, animation, collider, merchant, and dialog components but no genre component. At `pkg/engine/objective_tracker_system.go:761`, the code path `if genre, ok := genreComp.(*GenreComponent); ok { … }` is therefore always skipped for merchants, so genre-aware quest reward scaling for merchant-related objectives is always zero.
- **Blocked Goal**: Consistent genre-based reward scaling for all entity types; per-entity genre theming.
- **Implementation Path**:
  1. In `pkg/engine/merchant_spawn.go`, within `SpawnMerchantFromData()`, after the dialog components are added (near the end of the function), add: `if params.GenreID != "" { merchant.AddComponent(NewGenreComponent(params.GenreID)) }`. The `params procgen.GenerationParams` argument is already available in the function signature.
  2. Add a table-driven unit test in `pkg/engine/merchant_spawn_test.go` asserting that `SpawnMerchantFromData()` attaches a `GenreComponent` with the correct genre ID.
  3. Validate with `go test ./pkg/engine/`.
- **Dependencies**: None.
- **Effort**: Small (< 1 hour).
- **Severity**: MEDIUM

---

## N6 — `NewBlendedGenreComponent` exported but has zero callers

- **Intended Behavior**: Multi-genre worlds (e.g., a horror/cyberpunk blend) should use `NewBlendedGenreComponent` to attach a `GenreComponent` with a primary genre, secondary genres, and a blend ratio to entities, enabling blended theming and content generation.
- **Current State**: `NewBlendedGenreComponent` is defined and exported at `pkg/engine/genre_component.go:36` with a complete implementation (sets `SecondaryGenres`, `BlendRatio`, validates ratio). There are zero callers in `cmd/`, `pkg/`, or examples. The blend-ratio and secondary-genre fields on `GenreComponent` are never populated at runtime.
- **Blocked Goal**: No current feature is blocked; the API exists for a multi-genre feature that has not yet been wired to any content path.
- **Implementation Path**: Two options — (a) **Activate**: In the spawn helpers, check if `params.Custom["secondary_genres"]` is non-nil and use `NewBlendedGenreComponent` instead of `NewGenreComponent`. Wire the secondary genres into at least one content generator (e.g., the terrain composite generator) so the blend ratio influences generator selection. (b) **Defer**: Lower-case to `newBlendedGenreComponent` (unexported) to signal this is internal API awaiting future integration, preventing it from appearing in public API documentation as a usable constructor. Option (b) is low-effort; option (a) is medium-effort.
- **Dependencies**: N1 should be completed first to ensure `"postapoc"` is consistently recognised before blending it with other genres.
- **Effort**: Small for option (b); Medium for option (a).
- **Severity**: LOW

---

## N7 — `tradeRouteManager.Stop()` never called on client shutdown

- **Intended Behavior**: Background goroutines started by `tradeRouteManager.Start()` should be cleanly terminated when the game client exits, preventing goroutine and ticker resource leaks. The server correctly calls `Stop()` as part of its shutdown sequence.
- **Current State**: `cmd/client/init_versions.go:644-645` calls `sys.tradeRouteManager.Start()`, which launches a background goroutine (protected by `sync.Once`) with a `10 * time.Second` `time.Ticker`. The currently referenced handler location in earlier notes is stale: `cmd/client/handlers.go:3576-3602` is `handleHostAndPlay`, not a client shutdown/cleanup path, and there is no `initializeV19SystemsClient()` function in the current code. The goroutine and ticker therefore appear to lack a documented client-side shutdown hook and run until the OS reclaims the process (acceptable for production exit) but still cause false positives in `-race` testing and complicate graceful shutdown testing.
- **Blocked Goal**: No user-visible goal is blocked; this is a resource management correctness issue.
- **Implementation Path**: Wire `sys.tradeRouteManager.Stop()` into an actual client shutdown/cleanup path rather than deferring it inside a short-lived init helper. In particular, after `setupAllGameSystems(...)` returns in `cmd/client/main.go`, defer `sys.tradeRouteManager.Stop()` there, or register/return a cleanup closure from initialization and invoke that closure during client shutdown. Do **not** add `defer sys.tradeRouteManager.Stop()` inside an init helper such as `initializePhase3Systems`, because that would stop the manager immediately when the helper returns. `Stop()` is idempotent (uses `sync.Once`). Validate with `go test -race ./pkg/integration/trade_routes/`.
- **Dependencies**: None.
- **Effort**: Trivial (< 15 minutes).
- **Severity**: LOW

---

## N8 — OpenXR VR adapters are documented stubs (ROADMAP item)

- **Intended Behavior**: When the game is built with the `vr` tag and run with `--vr` on hardware with an OpenXR-compatible headset, the headset and controller adapters should delegate to the OpenXR SDK for head tracking, controller input, and haptic feedback.
- **Current State**: `pkg/engine/vr_openxr_adapters.go` (build tag `//go:build vr && !js`) contains 16 `TODO(vr-sdk):` markers. All adapter methods return zero values or mock responses. `cmd/client/init_versions.go:582,594` unconditionally uses `NewStubHeadsetAdapter` / `NewStubControllerAdapter` regardless of the `vr` build tag. README explicitly documents VR as "experimental with mock adapters only; no hardware SDK integration."
- **Blocked Goal**: No currently stated goal requires real VR hardware; README explicitly marks VR as experimental. Tracked in ROADMAP.md Priority 4.
- **Implementation Path**: When a stable Go OpenXR binding is available, implement each `TODO(vr-sdk):` stub per the inline pseudocode comments (which describe the exact OpenXR API calls required: `xrCreateInstance`, `xrLocateViews`, action-set setup, `xrGetActionStateFloat`, etc.). Switch `init_versions.go:582,594` to instantiate the OpenXR adapters when `--vr` is set and `--force-stub` is not. No action required until the SDK is available.
- **Dependencies**: External: Go OpenXR SDK / cgo binding for the target platform.
- **Effort**: Large (requires C/cgo integration, SDK bindings, and hardware testing).
- **Severity**: LOW

---

## Resolved Gaps (from prior GAPS.md, 2026-04-21)

| Prior ID | Description | Resolution |
|----------|-------------|------------|
| G1 | Story generator `"postapocalyptic"` key mismatch | Fixed: `generator.go:21,30` and all `timeline.go` cases use `"postapoc"`. |
| G2 | `AnimationSyncManager` receive path (`BufferState`/`GetNextState`) never called | Fixed: `network/client.go:726` calls `mgr.BufferState(pkt)`; `engine/animation_system.go:853` calls `DrainRemoteState`. |
| G3 | Guild vehicle formation physics and `FleetID` sync not wired | Fixed: `GuildVehicleSystem.Update()` (guild_vehicle_system.go:35–105) performs two-pass formation steering; `SetVehicleSyncer(sys)` called at construction. |
| G4 | `ArchaeologyGenerator`, `TimelineGenerator`, `CrossDungeonGenerator` not in game loop | Fixed: `cmd/client/util.go:2474, 2505, 2535` now calls all three generators. |
| G5 | `StoryJournalUI` instantiated but not wired | Fixed: `cmd/client/handlers.go:3371-3372` instantiates `NewStoryJournalUI`; toggle keybind and draw path wired. |
| G6 | `GenreComponent` attached only to player entity | Partially fixed: enemies now receive `GenreComponent` via `entity_spawning.go:275`. Merchants still missing → **N5**. |
| G7 | `equipment_durability_particle_system.go` with `//go:build ignore` | Fixed: File deleted; specialised variants fully replace it. |
| G8 | `InteractWithEnvironmentNode` never composed into any behavior tree | Fixed: `behavior_tree_archetypes.go:222` composes the node for crafter/gatherer archetypes. |
| G9 | `NewHelpSystemWithSize` exported but effectively private | Fixed: Renamed to unexported `newHelpSystemWithSize` (`help_system.go:53`). |
| G10 (prior) | OpenXR VR adapter stubs | Carried forward as **N8** (LOW, ROADMAP item). |
