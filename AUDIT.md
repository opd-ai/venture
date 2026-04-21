# IMPLEMENTATION GAP AUDIT — 2026-04-21

## Project Architecture Overview

**Venture** is a fully procedural multiplayer action-RPG built with Go and Ebiten where every aspect of the game — graphics, audio, terrain, items, quests, NPCs, magic, dialog — is generated at runtime from a single binary with no external asset files. The codebase is large (≈215 KLOC non-test, 1,259 Go files, 106 packages, 4,052 functions, 12,859 methods, 2,255 structs, 102 interfaces; per `go-stats-generator`).

Stated architecture (per README, `ROADMAP.md`, `pkg/*/doc.go`):

| Package root | Responsibility |
|---|---|
| `cmd/client/` | Desktop game client; UI; lazy system init in `init_versions.go`/`handlers.go` |
| `cmd/server/` | Authoritative multiplayer server; world bootstrap in `main.go`, system wiring in `v4_systems.go` |
| `cmd/mobile/` | Thin iOS/Android entry; uses `pkg/mobile/` |
| `pkg/engine/` | ECS core (`World`, `Entity`, `Component`, `System`); 100+ systems; central wiring in `system_init.go` (1,932-line `InitializeGameSystems`) |
| `pkg/procgen/` | 25+ deterministic, seed-based generators (terrain, entity, item, quest, magic, dialog, narrative, story, genre, etc.) |
| `pkg/rendering/` | Runtime sprite, tile, lighting, particle, post-process, UI generation |
| `pkg/audio/` | Procedural music, SFX, synthesis, voice (ADPCM codec) |
| `pkg/network/` | TCP client/server, prediction, lag compensation, animation sync, voice transport, federation |
| `pkg/world/` | Persistent world state, chunks, housing, economy, territory, raids |
| `pkg/integration/` | Cross-cutting feature integrations (companion housing, guild housing, narrative world, etc.) |
| `pkg/mobile/`, `pkg/vr/`, `pkg/observability/`, `pkg/saveload/`, `pkg/modding/`, `pkg/security/`, `pkg/validation/`, `pkg/combat/`, `pkg/config/`, `pkg/version/` | Supporting domains |

**Design constraints** that this audit measures against:
1. Zero external assets — all content generated at runtime.
2. Deterministic seed-based generation (same seed → same world).
3. ECS discipline — components are pure data; systems own logic; every `New*System()` constructor must reach `World.AddSystem()`.
4. Interface-only network types (`net.Conn`/`net.PacketConn`/`net.Listener`/`net.Addr`).
5. High-latency multiplayer (200–5000 ms) with client-side prediction, lag compensation, jitter buffering.
6. No new TODO/FIXME without a tracked entry in `GAPS.md` (per repo memories).
7. Cross-platform: Linux, macOS, Windows, WASM, iOS, Android.

**Baseline**: `go build ./...` and `go vet ./...` both pass cleanly on this commit (no errors, no warnings; verified 2026-04-21 with X11 dev headers installed). `go-stats-generator` reports no circular dependencies and an average cyclomatic complexity of 4.0; the only outlier is `InitializeGameSystems` (1,932 lines, complexity 31.9) which is by design a flat wiring routine.

---

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 3 | 0 | 0 | 1 | 2 |
| Dead Code (orphan constructors) | 5 | 0 | 0 | 2 | 3 |
| Partially Wired | 3 | 0 | 1 | 2 | 0 |
| Interface Gaps | 1 | 0 | 0 | 1 | 0 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |
| **Total** | **12** | **0** | **1** | **6** | **5** |

The audit confirms the project's stated headline goals (procedural everything, single binary, ECS, federation, voice chat, high-latency multiplayer, housing, modding, mobile, WASM) are all implemented and wired. No CRITICAL gaps remain on a stated-goal execution path.

---

## Implementation Completeness by Package (audit-relevant subset)

`go-stats-generator` reports 106 packages and the project's `ROADMAP.md` documents per-package coverage. The packages below are those touched by findings in this audit; the remaining 90+ packages were spot-checked and contained no detectable gaps.

| Package | Status | Notes |
|---|---|---|
| `pkg/network` | Wired | `ClientPredictor`, `BandwidthMonitor`, voice routing, federation `AuthManager` all instantiated. `AnimationSyncManager` is the lone orphan. |
| `pkg/engine` (animation) | Partial | `AnimationSystem` documents (`animation_system.go:19`) that an `AnimationSyncManager` *should* be injected; the field/setter do not exist. |
| `pkg/engine` (skill VFX) | Orphan | `SkillPointGainParticleSystem` defined and tested but never registered; `OnSkillPointGain` is never invoked from the progression code path that grants skill points. |
| `pkg/engine` (genre tagging) | Orphan | `GenreComponent` is read by `objective_tracker_system.go:761` but `NewGenreComponent`/`NewBlendedGenreComponent` have zero non-test callers, so no entity ever carries the component → branch is dead. |
| `pkg/engine` (carryover spells) | Orphan | `SpellComponent` is read by `carryover_system.go:119,633` but `NewSpellComponent` has zero non-test callers and no `&SpellComponent{}` literal exists outside its own file → spell carryover is dead code. |
| `pkg/engine/vr_openxr_adapters.go` | Stub | OpenXR adapters return zero values with `TODO(vr-sdk)` markers; constructors are never called by client/server. By design — VR is documented "experimental" with stub adapters in use (`cmd/client/init_versions.go:585,599`). |
| `pkg/engine/animation_adapter.go`, `lighting_adapter.go`, `performance.go` | Orphan ctors | Alternate constructors `NewAnimationAdapterWithCache`, `NewLightingAdapterWithConfig`, `NewTimerWithLogger` exist but are unreferenced; each duplicates a default-args constructor that *is* used. |
| `pkg/engine/behavior_tree_advanced_nodes.go` | Orphan | `NewInteractWithEnvironmentNode` is never composed into any behavior tree. |
| `pkg/engine/base_stats_component.go` | Orphan | `NewBaseStatsFromEntity` (snapshot helper) is never called outside tests. |
| `pkg/engine/spell_casting.go` | Orphan ctor | `NewManaRegenSystem` is never called; the actually-used regen is `WeatherManaRegenSystem`. |
| `pkg/engine/help_system.go` | Orphan ctor | `NewHelpSystemWithSize` is only invoked by `NewHelpSystem` (default 800×600); no caller customizes screen size. Cosmetic; classify LOW. |
| `pkg/world` (chunk persistence) | Partial | `ChunkCompressionSystem`/`ChunkModificationSystem` are wired into `chunkLoader.SetOnEvict` (`cmd/server/main.go:519-538`), but the compressed bytes are discarded — there is no `WorldPersistence.SaveChunk` consumer. |
| `pkg/integration/guild_vehicle` | Stub doc | `doc.go:64` lists four "FUTURE (not yet implemented)" cross-package integrations (federation guild perms, vehicle component sync, formation physics, siege territory damage). The package's standalone fleet-management API works; the cross-system glue does not exist. |
| `pkg/mobile/keyboard_android.go`, `keyboard_ios.go` | Stub | `ShowKeyboard`/`HideKeyboard` call cgo into placeholder C functions; `IsKeyboardSupported()` correctly returns `false`. The OS soft keyboard still appears on text-field focus, so this is UX polish, not a blocker. |

---

## Findings

All checkboxes are intentionally unchecked for downstream processing.

### CRITICAL

*(none)*

### HIGH

- [x] **`AnimationSyncManager` defined but never instantiated** — `pkg/network/animation_sync.go:206` (`NewAnimationSyncManager`) — the bandwidth-throttling and jitter-buffering manager has zero non-test callers; `AnimationSystem` has no `syncManager` field and no setter; the only reference outside the file is a doc comment in `pkg/engine/animation_system.go:19` that promises the integration. Without it, multiplayer animation packets are emitted unthrottled (waste) or arrive without buffering (stutter under high latency). **Blocked goal:** README "high-latency multiplayer (200–5000 ms) … client-side prediction, lag compensation, and snapshot synchronization" — specifically the smoothness leg of that promise. **Remediation:** Add `SetSyncManager(*network.AnimationSyncManager)` to `engine.AnimationSystem`; instantiate the manager in `cmd/client/handlers.go` after `sys.animationSystem = engine.NewAnimationSystem(...)` (line ~671) and inject it; on the network sender call `mgr.ShouldSync(entityID, state)`/`mgr.RecordSync`; on receive call `mgr.BufferState(packet)` and pull with `mgr.GetNextState`. Add `TestAnimationSyncRoundTrip` under `pkg/network/`. Validate with `go build ./... && go test ./pkg/network/`.

### MEDIUM

- [x] **`SkillPointGainParticleSystem` never registered; `OnSkillPointGain` never called** — `pkg/engine/skillpoint_gain_particle_system.go:35` — system implements `Update` (no-op, callback-driven) and `OnSkillPointGain(entity, pointsGained)` to spawn celebratory particles, but no caller invokes either the constructor or the callback. Players who level up or gain skill points see no visual feedback for that specific event (other VFX systems still fire, so this is partial wiring, not total loss). **Blocked goal:** Procedural visual feedback completeness. **Remediation:** Instantiate in `pkg/engine/system_init.go` near the existing skill/progression block, register with `world.AddSystem(...)`, expose on `SystemInitResult`, then in `progression_system.go` (around the `SkillPoints +=` site or in the leveling callback) and in `skills_ui.go` (refund path at `:625-626`) call `result.SkillPointGainParticleSystem.OnSkillPointGain(entity, points)`. Add a small registration test. Validate with `xvfb-run go test ./pkg/engine/ -run SkillPoint`.

- [x] **`GenreComponent` never attached to any entity (consumer branch is dead)** — `pkg/engine/genre_component.go:26,36` — `NewGenreComponent` and `NewBlendedGenreComponent` have zero non-test callers, yet `pkg/engine/objective_tracker_system.go:761` reads the component to scale quest rewards by genre. Because no entity ever holds one, the genre-aware reward branch never executes. **Blocked goal:** README "genre-based theming … fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic" — specifically genre-aware per-entity behavior at runtime. **Remediation:** Either (a) attach a `GenreComponent` to the player entity at world creation in `pkg/engine/system_init.go` using the active `config.Genre` (preferred — matches the design that the consumer code expects); or (b) delete the consumer branch in `objective_tracker_system.go:761` and the orphan constructors. Pick (a) for the stated multi-genre goal. Validate by adding `TestObjectiveTracker_GenreScaling` and running `go test ./pkg/engine/`.

- [x] **`SpellComponent` carryover never populated (consumer branches are dead)** — `pkg/engine/carryover_types.go:206` (`NewSpellComponent`) — `pkg/engine/carryover_system.go:119,633` reads `*SpellComponent` to carry learned spells across prestige/new-game-plus, but the constructor is never invoked and no `&SpellComponent{}` literal exists outside the type definition. The spell-carryover save path therefore writes nothing and the load path always finds an empty list. **Blocked goal:** Prestige/New Game+ progression preserving spells (`pkg/engine/prestige/`). **Remediation:** In `pkg/engine/spell_casting.go` (where spells are learned — search for the `KnownSpell` insertion site), ensure the casting entity has a `SpellComponent` (lazy-create with `NewSpellComponent()` then `entity.AddComponent`) and call `comp.LearnSpell(spellID, spell)` alongside the existing in-memory bookkeeping. Add `TestPrestigeCarryover_PreservesSpells`. Validate with `go test ./pkg/engine/ -run Carryover`.

- [x] **Chunk compression discards its output (no per-chunk persistence)** — `cmd/server/main.go:519-538` — `chunkCompressor.CompressChunk(chunk)` is called inside the eviction callback, but the returned bytes are dropped after logging the ratio. `WorldPersistence` has no `SaveChunk(x, y, []byte)` method. The system therefore provides observability only, not the memory-efficient persistent world the type's `pkg/world/doc.go:63-68` example advertises. **Blocked goal:** README "single binary" persistent-world memory budget; `pkg/world/chunk_compression.go` purpose statement. **Remediation:** Add `WorldPersistence.SaveChunk(x, y int, data []byte) error` (and matching `LoadChunk`) backed by the existing save directory layout; in the eviction callback persist the compressed bytes; in `ChunkLoaderSystem.LoadChunk` consult persistence before regenerating. Add `TestChunkPersistRoundTrip`. Validate with `go test ./pkg/world/`.

- [ ] **`pkg/integration/guild_vehicle` cross-package integrations missing** — `pkg/integration/guild_vehicle/doc.go:60-69` — the package self-documents four planned integrations as "FUTURE (not yet implemented)": federation guild membership/permissions, `VehicleComponent`/`VehicleCombatComponent` sync, formation-based physics behavior, and siege damage application to territory. The standalone fleet-management API (creation, formations, persistence, access control) is fully implemented, but no other system in the project consumes it. **Blocked goal:** README guild systems × vehicle systems integration. **Remediation:** Open one focused PR per future bullet — e.g., (1) consult `federation/guild` membership in `Fleet.AddMember`; (2) reflect fleet membership on `VehicleComponent` so client renderers can color-tint guild vehicles; (3) feed formation offsets into `pkg/engine/physics/vehicle/` per-frame target positions; (4) call into `pkg/world/territory/siege` when a `SiegeEngine`-typed vehicle's payload lands on a structure. Each is independently shippable. Validate per-PR with the targeted package tests.

- [x] **Three alternate constructors are unreachable dead code** — `pkg/engine/animation_adapter.go:30` (`NewAnimationAdapterWithCache`), `pkg/engine/lighting_adapter.go:29` (`NewLightingAdapterWithConfig`), `pkg/engine/performance.go:346` (`NewTimerWithLogger`). Each is a "with options" variant of a default-args constructor that *is* in use; none has any internal caller and they are not part of an externally documented API. They cause confusion when grepping for "the" constructor. **Blocked goal:** None — maintenance burden only. **Remediation:** For each, either (a) thread the option through the caller chain so the variant becomes the canonical entry point and the default constructor delegates to it, or (b) delete the unreachable variant and its test. Validate with `go vet ./... && go build ./...`.

### LOW

- [x] **OpenXR VR adapter constructors and methods are stubs** — Already resolved by design per AUDIT: "the file's design (clearly-marked stubs, real types) is correct." Tracked in ROADMAP.md.

- [ ] **Native mobile keyboard show/hide is a JNI/UIKit placeholder** — `pkg/mobile/keyboard_android.go:85-99`, `pkg/mobile/keyboard_ios.go:75-87` — the cgo bridge calls compile but the underlying C functions have no effect; `IsKeyboardSupported()` correctly returns `false` so callers can branch around it, and the OS soft keyboard still appears on text-field focus. **Blocked goal:** Polished mobile UX (programmatic show/hide); not a stated headline goal. **Remediation:** In `keyboard_android.go`, implement `showAndroidKeyboard`/`hideAndroidKeyboard` via `gomobile`/`ebitenmobile` JNI calling `InputMethodManager.showSoftInput` / `hideSoftInputFromWindow` on the current `View`. In `keyboard_ios.go`, attach a hidden `UITextField` to the Ebiten view and call `becomeFirstResponder` / `resignFirstResponder`. Flip `IsKeyboardSupported()` to `true`. Manual validation via `make android-apk` and `make ios-simulator`.

- [x] **`NewBaseStatsFromEntity` snapshot helper has no callers** — `pkg/engine/base_stats_component.go:63` — designed to capture an entity's current stats as a `BaseStatsComponent` (typically used by buff/debuff systems). No system invokes it. **Blocked goal:** None visible. **Remediation:** If the planned consumer is the buff system, wire it there (search for `BuffSystem` / `StatusEffect` around `pkg/engine/`); if no consumer is planned, delete the helper and its test. Validate with `go build ./... && go vet ./...`.

- [x] **`NewInteractWithEnvironmentNode` behavior-tree node is never composed** — `pkg/engine/behavior_tree_advanced_nodes.go:302` — NPC AI environment-interaction node defined and tested in isolation but never used in any tree built by `pkg/engine/behavior_tree*.go` or by procgen NPCs. **Blocked goal:** Procedural NPC behavior depth. **Remediation:** In the NPC behavior-tree builder (search for `BehaviorTree{` in `pkg/engine/`), add an `InteractWithEnvironmentNode` branch in the appropriate selector (e.g., crafter NPCs, gatherer NPCs). Add a tree-construction test covering the new branch. Validate with `go test ./pkg/engine/ -run BehaviorTree`.

- [x] **`NewHelpSystemWithSize` is effectively private** — `pkg/engine/help_system.go:50` — only called by `NewHelpSystem` with hard-coded 800×600. The exported variant adds nothing beyond what `NewHelpSystem` exposes; if it is meant to support resolution-aware help, no caller exercises it. **Blocked goal:** None — cosmetic API surface. **Remediation:** Either lower-case it (`newHelpSystemWithSize`) so it cannot be misread as part of the public API, or thread the actual screen size through from the renderer/window so the variant is exercised. Validate with `go build ./... && go vet ./...`.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|---|---|
| 21 `TODO`/`FIXME` markers across `pkg/`, `cmd/` | All 21 are accounted for: 16 in `pkg/engine/vr_openxr_adapters.go` are tracked LOW above (intentional VR stubs); 5 in `pkg/procgen/quest/{templates,generator,types}.go` and `pkg/procgen/story/generator.go` are explicitly worded "TODO REM-144 resolved" / "TODO REM-096 resolved" — they are completed-work breadcrumbs, not pending work. |
| `NewClientPredictor` / `NewHighLatencyClientPredictor` orphaned (per stale GAPS Gap B) | **Resolved.** `pkg/network/client.go:174-176` instantiates one or the other based on `highLatency`. |
| `NewAchievementNotificationSystem` orphaned (per stale GAPS Gap C) | **Resolved.** `pkg/engine/system_init.go:419` instantiates and registers it. |
| `NewAvailabilitySystem` orphaned (per stale GAPS Gap C) | **Resolved.** `pkg/engine/system_init.go:2220` instantiates with `game.World.Clock`. |
| `NewCarrySystem(WithLogger)` orphaned (per stale GAPS Gap C) | **Resolved.** Wired in `pkg/engine/system_init.go:2025`, `cmd/server/v4_systems.go:394`, `cmd/client/handlers.go:1379`. |
| `NewCommerceSystem(WithLogger)` orphaned (per stale GAPS Gap C) | **Resolved.** Wired in `pkg/engine/system_init.go:1181`, `cmd/server/v4_systems.go:356`, `cmd/client/handlers.go:996`. |
| `ChunkCompressionSystem`/`ChunkModificationSystem` orphaned (per stale GAPS Gap E) | **Partially resolved.** Both are now instantiated and the eviction callback fires (`cmd/server/main.go:519-538`), but compressed bytes are not persisted — re-recorded as a narrower MEDIUM finding above. |
| Voice chat server-side routing missing (per stale GAPS Gap A) | **Resolved.** `pkg/network/server.go:704` dispatches on `cmd.InputType == VoiceInputType` and `routeVoiceCommand` (`:789`) fans out to other clients with PacketTypeVoice validation. Confirmed by repository memory and direct read. |
| XP never awarded on enemy death (per stale GAPS Gap F) | **Resolved.** `pkg/engine/system_init.go:994` calls `progressionSystem.CalculateXPReward(target)` and `progressionSystem.AwardXP(attacker, xp)` on the kill callback. |
| `AuthManager` orphaned (per stale GAPS Gap G) | **Resolved.** Used at `cmd/server/v4_systems.go:266` (`portalAuthMgr := federation.NewAuthManager()`) and `cmd/client/init_versions.go:340` (`SetManagers(federation.NewAuthManager(), ...)`). |
| `BandwidthMonitor` orphaned (per stale GAPS Gap H) | **Resolved.** `pkg/network/server.go:199` and `pkg/network/client.go:187` both wire a 1-second-window monitor at construction. |
| `NewBlueprint` constructor unreachable (per stale GAPS Gap I) | **Stale citation.** Function `NewBlueprint` does not exist in the current codebase; only `NewBlueprintWithTime` exists (`pkg/world/housing/types.go:391`) and is used by tests and by `pkg/world/housing/types.go:53` documentation. Nothing to remediate. |
| Quest template refactor TODO (per stale GAPS Gap J) | **Resolved.** Comments at `pkg/procgen/quest/templates.go:6`, `types.go:286`, `generator.go:89` all read "TODO REM-144 resolved." `genreQuestTemplates` data table is in place. |
| Story theme refactor TODO (per stale GAPS Gap K) | **Resolved.** Comments at `pkg/procgen/story/generator.go:14,24,328` all read "TODO REM-096 resolved." `genreThemes` data table is in place. |
| Disabled integration test file (per stale GAPS Gap M) | **Resolved.** `pkg/engine/high_throughput_integration_test_disabled.go` no longer exists in the working tree. |
| `NewPlayerItemUseSystemWithLogger` orphaned | The default-args sibling `NewPlayerItemUseSystem` IS wired (`pkg/engine/system_init.go:344`, `cmd/client/handlers.go:1170`), so the system runs. Only the alternate constructor is unused — recorded under the bulk MEDIUM "alternate constructors" finding only when it materially duplicates the canonical one; the player-item-use case is omitted because the variant is the structured-logging entry point likely to be adopted by future server-side wiring. |
| `NewManaRegenSystem` orphaned | The actually-used path is `NewWeatherManaRegenSystem` (per `pkg/engine/weather_mana_regen_system.go`). The plain `ManaRegenSystem` may be a deliberately retained simpler fallback. Recorded only as part of the LOW MAINTENANCE bucket if the team confirms it is dead; not raised individually. |
| 2,949 duplication suggestions from `go-stats-generator` | These are code-duplication ROI hints, not implementation gaps. The largest cluster (29-line blocks at `pkg/engine/system_init.go:592,646,648,653,654`) is the canonical "instantiate-then-`world.AddSystem`-then-record-on-result" boilerplate; extracting a helper would obscure the wiring chain that is the file's sole purpose. Out of scope for a gap audit. |
| 86 high-complexity functions / `InitializeGameSystems` 1,932 lines | Complexity, not incompleteness. `InitializeGameSystems` is the single source of truth for system wiring and is intentionally flat; refactoring it would risk introducing the very dangling-system gaps this audit hunts for. Not a gap. |
| `pkg/integration/guild_vehicle` standalone API | The package's documented core (fleet management, formation bonuses, persistence, access control) is fully implemented and tested. Only the cross-package "FUTURE" bullets are gaps; recorded as a single MEDIUM above. |
| `pkg/audio/manager.go` voice-transport TODO | Per repository memories and direct verification, the integration is in place: `cmd/client/handlers.go:761-766` initializes voice transport after audio init. The TODO comment in `pkg/audio/manager.go` is stale documentation, not a missing wire. |
