# Implementation Gaps — 2026-04-21

This document is the detailed companion to `AUDIT.md`. Each gap was verified against the working tree on 2026-04-21 with `go build ./...` and `go vet ./...` both clean. Citations are file:line references that resolve in the current `HEAD`.

Severity scale: 🔴 CRITICAL, 🟠 HIGH, 🟡 MEDIUM, 🟢 LOW.

Tiebreaker order applied: stubs on critical paths → partially wired features → dead code on maintained paths → interface gaps → dependency gaps → tracked TODOs.

---

## 🟠 Gap 1 — `AnimationSyncManager` Defined But Never Instantiated

- **Intended Behavior**: README — "high-latency multiplayer (200–5000 ms) … client-side prediction, lag compensation, and snapshot synchronization" — and the doc comment at `pkg/engine/animation_system.go:19` ("multiplayer animation states sync … with throttling and jitter buffering, producing smooth remote-player animations under high latency"). The mechanism exists in `pkg/network/animation_sync.go:188-271`: `AnimationSyncManager` provides `ShouldSync`, `RecordSync`, `BufferState`, `GetNextState`, per-entity buffers, and bandwidth/jitter stats.
- **Current State**: `NewAnimationSyncManager` (`pkg/network/animation_sync.go:206`) has **zero non-test, non-self references** anywhere in `cmd/` or `pkg/`. `AnimationSystem` has no `syncManager` field and no setter. The only external mention is the aspirational doc comment in `pkg/engine/animation_system.go:19` and the placeholder note in `cmd/client/handlers.go:674` ("AnimationSyncManager is instantiated when the animation-state network send/receive path is ready to consume it").
- **Blocked Goal**: Smooth multiplayer animation under 200–5000 ms latency (a stated subset of the headline high-latency multiplayer goal). Without the sync manager the project ships either over-frequent animation packets (bandwidth waste) or unbuffered packets (visible stutter), depending on what code paths a sender choose ad-hoc.
- **Implementation Path**:
  1. In `pkg/engine/animation_system.go`, add `syncManager *network.AnimationSyncManager` field plus `SetSyncManager(*network.AnimationSyncManager)`.
  2. In `cmd/client/handlers.go` (replace the placeholder at `:674`), construct `mgr := network.NewAnimationSyncManager()` and call `sys.animationSystem.SetSyncManager(mgr)`.
  3. In the network sender (search for the AnimationStatePacket emission site in `pkg/network/`), gate sends on `mgr.ShouldSync(entityID, state)` and follow successful sends with `mgr.RecordSync(entityID, state, n)`.
  4. In the network receiver, call `mgr.BufferState(packet)` and pull with `mgr.GetNextState(entityID)` from the playback frame in `AnimationSystem.Update`.
  5. Add `TestAnimationSyncRoundTrip` under `pkg/network/` covering buffering, ordered playback, and throttle behavior.
- **Dependencies**: None.
- **Effort**: small (1–2 days)

---

## 🟡 Gap 2 — `SkillPointGainParticleSystem` Never Registered; `OnSkillPointGain` Never Invoked

- **Intended Behavior**: When an entity gains skill points (level-up or quest reward), spawn celebratory particles at the entity's position. The system in `pkg/engine/skillpoint_gain_particle_system.go:18-180` is fully implemented: `SetParticleSystem`, `SetGenre`, `Update` (callback-driven, no per-frame work), `OnSkillPointGain`, `spawnSkillPointParticles`, `getParticleTypes`, `SpawnSkillPointEffect`. Tests cover the type.
- **Current State**: `NewSkillPointGainParticleSystem` (`pkg/engine/skillpoint_gain_particle_system.go:35`) has **zero non-test references**. Skill points are awarded in `pkg/engine/skills_ui.go:625-626` (refund), `pkg/engine/objective_tracker_system.go:753` (quest reward), and via `progression_system.go` leveling — none of these call `OnSkillPointGain`. Players therefore never see the particle effect.
- **Blocked Goal**: Procedural visual feedback completeness (subset of the rendering-pipeline goal). Other VFX systems still fire, so this is feature-incomplete rather than feature-absent.
- **Implementation Path**:
  1. In `pkg/engine/system_init.go`, near the existing skill/progression block, instantiate the system and register: `skillPointVFX := NewSkillPointGainParticleSystem(game.World, config.Seed); skillPointVFX.SetParticleSystem(particleSystem); skillPointVFX.SetGenre(config.Genre); game.World.AddSystem(skillPointVFX); result.SkillPointGainParticleSystem = skillPointVFX`.
  2. Add `SkillPointGainParticleSystem *SkillPointGainParticleSystem` to `SystemInitResult`.
  3. In `pkg/engine/progression_system.go`, at the call site that grants `exp.SkillPoints += n`, also call `result.SkillPointGainParticleSystem.OnSkillPointGain(entity, n)`.
  4. Mirror in `pkg/engine/objective_tracker_system.go:753` and `skills_ui.go:625-626`.
  5. Add `TestSkillPointGainParticleSystem_TriggeredOnAward` (table-driven, asserts particle spawn count).
- **Dependencies**: None (`ParticleSystem` is already constructed before `SystemInitResult` is populated).
- **Effort**: small (≈1 day)

---

## 🟡 Gap 3 — `GenreComponent` Read But Never Attached (Consumer Branch is Dead)

- **Intended Behavior**: Per-entity genre tagging so quest reward scaling, NPC behavior, and other systems can read `*GenreComponent` from an entity and adapt. `pkg/engine/objective_tracker_system.go:761` does exactly this: `if genre, ok := genreComp.(*GenreComponent); ok { ... }`.
- **Current State**: `NewGenreComponent` (`pkg/engine/genre_component.go:26`) and `NewBlendedGenreComponent` (`:36`) have **zero non-test callers**. No `&GenreComponent{}` literal exists outside the type definition. Consequently `entity.GetComponent("genre")` always returns nil and the consumer branch in `objective_tracker_system.go:761` is unreachable in production.
- **Blocked Goal**: README "genre-based theming … fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic." The global `config.Genre` is honored at world generation, but per-entity multi-genre blending (the very reason `NewBlendedGenreComponent` exists) cannot occur.
- **Implementation Path**:
  1. In `pkg/engine/system_init.go` (player-entity creation block), attach the component: `playerEntity.AddComponent(NewGenreComponent(config.Genre))`. If `config.SecondaryGenres` is populated, use `NewBlendedGenreComponent(config.Genre, config.SecondaryGenres, config.BlendRatio)`.
  2. Wherever NPCs/entities are spawned at runtime (e.g., `pkg/procgen/entity/`, spawner systems in `pkg/engine/`), add the same component using the active world genre.
  3. Add `TestObjectiveTracker_GenreScaling` covering both single-genre and blended-genre cases.
  4. Alternative if cross-system genre-per-entity is genuinely not wanted: delete the consumer branch at `objective_tracker_system.go:761` plus both orphan constructors. Do **not** ship a state where the consumer reads a component nothing produces.
- **Dependencies**: None.
- **Effort**: small (≈1 day)

---

## 🟡 Gap 4 — `SpellComponent` Carryover Never Populated (Consumer Branches are Dead)

- **Intended Behavior**: Prestige / New Game+ should preserve learned spells across runs. `pkg/engine/carryover_system.go:119` and `:633` read `*SpellComponent` from entities to serialize the spell list into the carryover save.
- **Current State**: `NewSpellComponent` (`pkg/engine/carryover_types.go:206`) has **zero non-test callers** and no `&SpellComponent{}` literal exists outside the type definition. Spells learned via `pkg/engine/spell_casting.go` are tracked in other structures but never written into a `SpellComponent`, so the carryover load path always finds an empty list and the prestige spell-preservation feature is silently a no-op.
- **Blocked Goal**: `pkg/engine/prestige/` New Game+ progression preserving spells (referenced by README's progression depth).
- **Implementation Path**:
  1. Locate the spell-learning path in `pkg/engine/spell_casting.go` (search for the `KnownSpell` insertion site and the public "learn spell" entry point).
  2. Lazy-attach a `SpellComponent` to the casting entity: if `entity.GetComponent("spells")` is nil, `entity.AddComponent(NewSpellComponent())`.
  3. Call `comp.LearnSpell(spellID, spell)` alongside the existing in-memory bookkeeping.
  4. Add `TestPrestigeCarryover_PreservesSpells` covering learn → save → reset → load → assert.
- **Dependencies**: None.
- **Effort**: small (≈1 day)

---

## 🟡 Gap 5 — Chunk Compression Discards Its Output (No Per-Chunk Persistence)

- **Intended Behavior**: `pkg/world/doc.go:63-68` advertises the example flow `mods := world.NewChunkModificationSystem(state); compressor := world.NewChunkCompressionSystem()` for memory-efficient persistent worlds. Compressed bytes from evicted chunks should be persisted so that re-loading does not require regeneration.
- **Current State**: `cmd/server/main.go:519-538` does instantiate both systems and wire them into `chunkLoader.SetOnEvict`. However the compressed `_, ratio, err := chunkCompressor.CompressChunk(chunk)` discards the bytes after logging the ratio: the comment at `:514-518` admits "CompressChunk returns compressed bytes for use when a per-chunk persistence layer is added (see WorldPersistence). Until then we call it to validate that chunks are compressible and log the ratio for memory observability." `WorldPersistence` has no `SaveChunk(x, y, []byte)` method.
- **Blocked Goal**: Persistent-world memory budget (README single-binary durability promise).
- **Implementation Path**:
  1. Extend `WorldPersistence` (in `pkg/world/`) with `SaveChunk(x, y int, data []byte) error` and `LoadChunk(x, y int) ([]byte, bool, error)`, backed by per-chunk files under the existing save directory layout (e.g., `chunks/x_y.gz`).
  2. In the eviction callback in `cmd/server/main.go:519-538`, write the compressed bytes via `worldPersistence.SaveChunk(chunk.X, chunk.Y, compressedBytes)` and only call `chunkMods.MarkDirty` if the save fails.
  3. In `ChunkLoaderSystem.LoadChunk` (`pkg/world/`), consult `LoadChunk` before regenerating; on hit, decompress with `chunkCompressor.DecompressChunk`.
  4. Add `TestChunkPersistRoundTrip` (generate → evict → load → assert tile equality).
- **Dependencies**: None.
- **Effort**: medium (3–4 days; touches save layout)

---

## 🟡 Gap 6 — `pkg/integration/guild_vehicle` Cross-Package Integrations Missing

- **Intended Behavior**: `pkg/integration/guild_vehicle/doc.go:60-69` self-documents four planned cross-system integrations between the guild and vehicle subsystems.
- **Current State**: The standalone fleet API (creation, formation bonuses, maintenance costs, gzip persistence, access control) is fully implemented. The four "FUTURE (not yet implemented)" bullets are not:
  1. `pkg/network/federation/guild` — Guild membership validation and permissions
  2. `pkg/engine` — `VehicleComponent` and `VehicleCombatComponent` synchronization
  3. `pkg/engine/physics/vehicle` — Formation-based physics behavior
  4. `pkg/world/territory` — Siege damage application to territory structures
- **Blocked Goal**: Guild × vehicle interplay implied by README's deep guild and vehicle systems.
- **Implementation Path**: One focused PR per bullet, in dependency order:
  1. **Guild membership in `Fleet.AddMember`** — call into `federation/guild` to validate that the joining member is in the fleet's owning guild.
  2. **Vehicle component sync** — when a vehicle joins a fleet, mark its `VehicleComponent.FleetID` so renderers can color-tint guild vehicles.
  3. **Formation physics** — feed `Fleet.GetFormationOffsets()` into `pkg/engine/physics/vehicle/` per-frame target positions for non-leader fleet members.
  4. **Siege damage to territory** — when a vehicle of `SiegeEngineType != ""` lands its payload on a territory structure, call `pkg/world/territory/siege` to apply damage with the guild as attacker.
- **Dependencies**: Item 4 depends on items 1-2 to know the attacker guild.
- **Effort**: medium per bullet; large aggregate (≈8-10 days)

---

## 🟡 Gap 7 — Three Alternate Constructors Are Unreachable Dead Code

- **Intended Behavior**: Each of `NewAnimationAdapterWithCache` (`pkg/engine/animation_adapter.go:30`), `NewLightingAdapterWithConfig` (`pkg/engine/lighting_adapter.go:29`), and `NewTimerWithLogger` (`pkg/engine/performance.go:346`) is presented as a "with options" variant of a default-args constructor.
- **Current State**: All three have **zero non-test, non-self references** in `cmd/` or `pkg/`. The default-args sibling constructors are wired and used. The alternate variants therefore exist as dead code that confuses anyone grepping for "the" constructor.
- **Blocked Goal**: None — code-clarity / maintenance burden.
- **Implementation Path**: Per constructor, choose one:
  - **Adopt**: thread the option through callers so the variant becomes the canonical entry point and the default-args version delegates to it (`func NewAnimationAdapter(g) { return NewAnimationAdapterWithCache(g, defaultCache, nil) }`).
  - **Delete**: remove the variant and any test that exercises it solely.
- **Dependencies**: None.
- **Effort**: trivial each (<1 hour)

---

## 🟢 Gap 8 — OpenXR VR Adapters Are Documented Stubs

- **Intended Behavior**: `pkg/engine/vr_openxr_adapters.go:46-230` would, if completed, drive a real OpenXR runtime: `xrCreateInstance`, `xrLocateViews`, action-set bindings for trigger/grip/thumbstick/buttons, `xrApplyHapticFeedback`.
- **Current State**: All methods return zero values; 16 `TODO(vr-sdk):` markers in the file describe the OpenXR calls that would replace each placeholder. `NewOpenXRHeadsetAdapter` and `NewOpenXRControllerAdapter` are never called — `cmd/client/init_versions.go:585,599` instantiates `NewStubVRHeadsetAdapter`/`NewStubVRControllerAdapter` instead.
- **Blocked Goal**: None at the stated-goal level. README and `ROADMAP.md` row "VR/Stereoscopic Support" explicitly classify VR as "experimental with mock adapters only; no hardware SDK integration."
- **Implementation Path**:
  1. Adopt a Go OpenXR binding (e.g., write a thin cgo wrapper, or use a community package once available; track in `ROADMAP.md`).
  2. Replace each `TODO(vr-sdk)` per the inline pseudocode.
  3. Switch `cmd/client/init_versions.go:585,599` to the OpenXR constructors when the `--vr` flag is set without `--force-stub`; keep the stub fallback for headless CI.
  4. Add `TestOpenXRAdapter_RoundTrip` behind a build tag (e.g., `//go:build openxr`).
- **Dependencies**: External — Go OpenXR binding availability.
- **Effort**: large (1-2 weeks once a binding is chosen)

---

## 🟢 Gap 9 — Native Mobile Keyboard Show/Hide Is a JNI/UIKit Placeholder

- **Intended Behavior**: Programmatic show/hide of the on-screen keyboard on Android and iOS, matching the WASM bridge in `pkg/mobile/keyboard_wasm.go`. `IsKeyboardSupported()` should return `true` on those platforms.
- **Current State**:
  - `pkg/mobile/keyboard_android.go:85-99`: cgo calls `C.showAndroidKeyboard()` / `C.hideAndroidKeyboard()`, but the underlying C functions are placeholders. `IsKeyboardSupported()` honestly returns `false`.
  - `pkg/mobile/keyboard_ios.go:75-87`: same pattern with `C.showIOSKeyboard()` / `C.hideIOSKeyboard()` and the same honest `false` return.
  - The OS soft keyboard still appears on text-field focus, so users can still type; only programmatic control is missing.
- **Blocked Goal**: Mobile UX polish — not a stated headline goal.
- **Implementation Path**:
  1. **Android**: implement `showAndroidKeyboard` / `hideAndroidKeyboard` via JNI — fetch the current `View` from the `ebitenmobile`/`gomobile` activity, then call `InputMethodManager.showSoftInput(view, SHOW_IMPLICIT)` / `hideSoftInputFromWindow(view.getWindowToken(), 0)`.
  2. **iOS**: add a hidden `UITextField` to the Ebiten `UIView`; call `becomeFirstResponder` / `resignFirstResponder`.
  3. Flip both `IsKeyboardSupported()` returns to `true`.
  4. Manual validation via `make android-apk` and `make ios-simulator`. Per repository memory, both files use the build-tag pair `(android|ios) && cgo && ebitenmobilebind`.
- **Dependencies**: None.
- **Effort**: medium (2-4 days; cgo/JNI plumbing)

---

## 🟢 Gap 10 — `NewBaseStatsFromEntity` Snapshot Helper Has No Callers

- **Intended Behavior**: Capture an entity's current stats as a `BaseStatsComponent` for later restoration (typical use: buff/debuff systems wanting to remember the unmodified baseline).
- **Current State**: `pkg/engine/base_stats_component.go:63` (`NewBaseStatsFromEntity`) has **zero non-test references**. The component type itself may be used elsewhere, but the snapshot helper specifically is dead.
- **Blocked Goal**: None visible.
- **Implementation Path**: If the planned consumer is a buff/status-effect system that needs to remember the pre-buff baseline, locate that system (search for `BuffSystem` / `StatusEffect` in `pkg/engine/`) and call `NewBaseStatsFromEntity` at buff application; restore on expiry. If no consumer is planned, delete the helper and any test that targets only it.
- **Dependencies**: None.
- **Effort**: trivial (<1 day either way)

---

## 🟢 Gap 11 — `NewInteractWithEnvironmentNode` Behavior-Tree Node Never Composed

- **Intended Behavior**: NPC AI behavior trees include an environment-interaction branch (chop trees, mine ore, open chests, etc.) realised by `InteractWithEnvironmentNode`.
- **Current State**: `pkg/engine/behavior_tree_advanced_nodes.go:302` (`NewInteractWithEnvironmentNode`) has **zero non-test references**. No behavior-tree builder in `pkg/engine/` composes this node into a tree.
- **Blocked Goal**: Procedural NPC behavior depth (subset of the engine's NPC system goal).
- **Implementation Path**: In the NPC behavior-tree builder (search for `BehaviorTree{` in `pkg/engine/`), add an `InteractWithEnvironmentNode` branch in the appropriate selector — likely on crafter and gatherer NPC archetypes. Add a tree-construction test that asserts the new branch is reachable when an interactable entity is in range.
- **Dependencies**: None.
- **Effort**: small (≈1 day)

---

## 🟢 Gap 12 — `NewHelpSystemWithSize` Is Effectively Private

- **Intended Behavior**: A resolution-aware help-system constructor that callers customize for non-default screen sizes.
- **Current State**: `pkg/engine/help_system.go:50` is only called by `NewHelpSystem` at `:46` with hard-coded 800×600. No external caller customizes the size, and the renderer/window plumbing does not feed actual screen dimensions through.
- **Blocked Goal**: None — cosmetic API surface.
- **Implementation Path**: Either (a) lower-case to `newHelpSystemWithSize` so it cannot be misread as part of the public API, or (b) thread the actual screen size from the renderer/window into `NewHelpSystemWithSize` so the variant is exercised. Validate with `go build ./... && go vet ./...`.
- **Dependencies**: None.
- **Effort**: trivial (<1 hour)

---

## Summary

| Gap | Severity | Effort | Stated Goal Affected |
|---|---|---|---|
| 1 — `AnimationSyncManager` unwired | 🟠 HIGH | small | High-latency multiplayer smoothness |
| 2 — `SkillPointGainParticleSystem` unregistered | 🟡 MEDIUM | small | Visual feedback completeness |
| 3 — `GenreComponent` never attached | 🟡 MEDIUM | small | Genre-based theming (per-entity) |
| 4 — `SpellComponent` carryover unpopulated | 🟡 MEDIUM | small | Prestige / New Game+ |
| 5 — Chunk compression bytes discarded | 🟡 MEDIUM | medium | Persistent-world memory budget |
| 6 — `guild_vehicle` cross-pkg integration | 🟡 MEDIUM | medium ×4 | Guild × vehicle interplay |
| 7 — 3 alternate constructors dead | 🟡 MEDIUM | trivial each | Code clarity |
| 8 — OpenXR adapters are stubs | 🟢 LOW | large | (none — explicitly experimental) |
| 9 — Native mobile keyboard JNI/UIKit | 🟢 LOW | medium | Mobile UX polish |
| 10 — `NewBaseStatsFromEntity` orphan | 🟢 LOW | trivial | (none) |
| 11 — `InteractWithEnvironmentNode` orphan | 🟢 LOW | small | NPC behavior depth |
| 12 — `NewHelpSystemWithSize` effectively private | 🟢 LOW | trivial | (none) |

**Totals**: 0 CRITICAL · 1 HIGH · 6 MEDIUM · 5 LOW = 12 verified open gaps.

*Validated 2026-04-21 against `go build ./...` and `go vet ./...` (both clean) and `grep`-verified for every constructor/symbol cited.*
