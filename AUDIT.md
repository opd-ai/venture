# IMPLEMENTATION GAP AUDIT — 2026-05-02

## Project Architecture Overview

**Venture** is a fully procedural multiplayer action-RPG built on Ebiten v2.9.3 with a zero-asset
philosophy: all graphics, audio, terrain, items, quests, and NPCs are generated at runtime from a
single binary. The architecture is:

| Layer | Responsibility |
|-------|---------------|
| `cmd/client/` | Desktop game client; calls `setupAllGameSystems()` in `handlers.go` |
| `cmd/server/` | Authoritative dedicated server; initializes systems in `main.go` + `v*_systems.go` |
| `cmd/mobile/` | iOS/Android entry point; calls `engine.InitializeGameSystems()` |
| `pkg/engine/` | ECS core (World/Entity/Component/System), 250+ game systems, 1,170 Go files |
| `pkg/procgen/` | 25+ deterministic, seed-based content generators |
| `pkg/rendering/` | Runtime sprite, tile, lighting, particle, UI pipeline |
| `pkg/audio/` | Procedural music, SFX, voice codec |
| `pkg/network/` | Multiplayer with high-latency (200–5000ms) and federation support |
| `pkg/world/` | Persistent world state — housing, economy, territory, raids, chunk loading |
| `pkg/integration/` | 9 cross-system integration packages |

**Codebase metrics** (non-test):
- Go files: 1,256 | Test files: 1,069
- Total LoC: ~400,000 | Packages: 123
- World.AddSystem calls in system_init.go: 250 | Additional in handlers.go: 266

**CI quality gates:** `go vet`, `gofmt`, `govulncheck`, `go test -race`, `scripts/validate-network-types.sh`, benchmark regression, memory budget.

---

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs / TODOs | 5 | 0 | 1 | 2 | 2 |
| Dead Code / Dangling Systems | 9 | 1 | 4 | 4 | 0 |
| Partially Wired | 5 | 3 | 2 | 0 | 0 |
| Interface Gaps | 1 | 1 | 0 | 0 | 0 |
| Dependency Gaps (nil injection) | 1 | 1 | 0 | 0 | 0 |
| **Total** | **21** | **6** | **7** | **6** | **2** |

---

## Implementation Completeness by Package

| Package | Exported Symbols (est.) | Implemented | Stubs / TODOs | Dangling | Notes |
|---------|------------------------|-------------|---------------|----------|-------|
| `pkg/engine` | ~1 800 | ~98% | 0 | 9 systems | 250 systems wired; 9 constructors never called |
| `pkg/audio` | ~80 | ~90% | 1 | 0 | VoiceTransport not wired to network |
| `pkg/network` | ~120 | ~95% | 0 | 0 | TCPVoiceTransport exists but never used |
| `pkg/procgen` | ~300 | ~99% | 2 | 0 | Refactoring TODOs only |
| `pkg/world` | ~80 | ~80% | 0 | 3 | Chunk systems never instantiated |
| `pkg/rendering` | ~200 | ~100% | 0 | 0 | |
| `pkg/integration` | ~60 | ~100% | 0 | 0 | All 9 packages wired in cmd/ |
| `cmd/client` | ~500 | ~98% | 2 | 0 | XP-on-kill and respec-gold TODOs |
| `cmd/mobile` | ~50 | ~90% | 1 | 0 | CraftingSystem nil itemGen |

---

## Findings

### CRITICAL

- [x] **XP never awarded on enemy death** — `cmd/client/util.go:1444` — `createDeathCallback` does not call `progressionSystem.AwardXP()` and `progressionSystem` is not a parameter of the callback. `handlers.go` sets `SetDeathCallback` (not `SetKillCallback`), so the kill-to-XP path is completely severed on the desktop client. `AwardXP` is defined at `pkg/engine/progression_system.go:97` but only reaches entities via `pkg/engine/faction_xp_bonus_system.go:81` (faction bonus, not base combat XP). Players cannot level up from combat. — **Blocked goal:** "100% Procedural Content" and "ECS Architecture Discipline" (the six-link integration chain: Update→Output→Consumer→Player Effect is broken). — **Remediation:** Add `progressionSystem *engine.ProgressionSystem` parameter to `createDeathCallback`; call `progressionSystem.AwardXP(playerEntity, enemy.BaseXP())` after marking the entity dead. Wire from `configureDeathCallback` in `cmd/client/handlers.go:3482`. — **FIXED:** Added `calculateKillXP()` and `progressionSystem.AwardXP()` call to `SetKillCallback` in `pkg/engine/system_init.go`.

- [x] **GPU memory leak in LoadingUI — ebiten.Image created 3× per frame, never disposed** — `pkg/engine/loading_ui.go:87` — `drawRect()` calls `ebiten.NewImage(width, height)` and never calls `img.Dispose()`. It is invoked 3 times per frame (lines 75, 76, 81) for the duration of world generation (5–30 seconds), leaking hundreds to thousands of GPU-backed images. On low-memory targets (mobile, WASM) this causes OOM crashes before gameplay begins. — **Blocked goal:** "<500 MB client memory target" and "Mobile support". — **Remediation:** Promote the three fixed-size rect images to `LoadingUI` fields initialized in `NewLoadingUI`, reusing them each frame; or add `defer img.Dispose()` inside `drawRect` after line 87. — **FIXED:** Added `defer img.Dispose()` in `drawRect()`.

- [x] **Mobile CraftingSystem initialized with nil itemGenerator — panics on craft attempt** — `pkg/engine/system_init.go:1168` — `InitializeGameSystems()` (called by `cmd/mobile/mobile.go:164`) creates the CraftingSystem with `nil` for itemGen: `NewCraftingSystem(game.World, inventorySystem, nil) // itemGen set later`. No setter for `itemGenerator` exists on `CraftingSystem` (only `SetStationManager` at `pkg/engine/crafting_system.go:75`), and no subsequent code on the mobile path sets it. `pkg/engine/crafting_system.go:579` dereferences `s.itemGenerator.Generate(...)` without a nil check, causing a nil pointer dereference (panic) on any craft attempt on mobile. — **Blocked goal:** "Mobile support". — **Remediation:** Add `func (s *CraftingSystem) SetItemGenerator(gen *item.ItemGenerator)` to `pkg/engine/crafting_system.go`; call it with `item.NewItemGenerator()` immediately after `InitializeGameSystems` returns in `cmd/mobile/mobile.go`. — **FIXED:** Added `SetItemGenerator()` method, nil guard in crafting path, and passed `item.NewItemGenerator()` directly in `system_init.go`.

- [ ] **VoiceTransport never wired — voice chat non-functional in multiplayer** — `pkg/audio/manager.go:310` — `InitializeVoice()` is defined at line 313 but never called from `cmd/client/` or `cmd/server/`. `TCPVoiceTransport` (a complete implementation of `audio.VoiceTransport`) exists at `pkg/network/voice_transport.go:94` but `NewTCPVoiceTransport` has zero non-test callers outside that file. The comment `// TODO(integration): Wire VoiceTransport to network/chat subsystem` at `pkg/audio/manager.go:310` confirms this is a known gap. Voice codec, spatial audio, and channel systems all exist but produce no network packets. — **Blocked goal:** README claims "Voice chat is integrated with party, guild, proximity, and private channels." — **Remediation:** See GAPS.md Gap 1 for detailed implementation path.

- [ ] **VoiceChannelSystem and VoiceAudioSystem wired on client but VoiceTransport never provided** — `cmd/client/handlers.go:1067,1069` — Both systems are instantiated and added to the world at lines 2215–2216, but without a network transport the codec frames they generate are silently discarded. This is a consequence of the VoiceTransport gap above. — **Remediation:** Resolved by fixing Gap 1 above.

- [ ] **ScriptingSystem defined but never instantiated — mod scripts cannot execute** — `pkg/engine/scripting_system.go:52` — `NewScriptingSystem` has zero callers outside its definition file. The modding system loads JSON mods via `pkg/modding/` (which is wired), but `ScriptingSystem` is the ECS-integrated sandbox that evaluates `Script` components on entities. Without it, any entity with a `ScriptingComponent` silently no-ops on every `Update`. — **Blocked goal:** "Sandboxed JSON-based modding system." — **Remediation:** Add `scriptingSystem := NewScriptingSystem(game.World); game.World.AddSystem(scriptingSystem)` in `pkg/engine/system_init.go` after the modding block; wire to `modding.Manager.SetScriptingSystem(scriptingSystem)` if such an interface is available.

---

### HIGH

- [ ] **WorldPersistenceSystem defined but never instantiated — world state not persisted** — `pkg/engine/world_persistence_system.go:22` — `NewWorldPersistenceSystem()` has zero callers outside its own file. The system manages city state, NPC state, and world events across sessions via `SaveWorldState`/`LoadWorldState`. Without it, the world is regenerated from scratch on every session despite the save/load infrastructure. — **Remediation:** Instantiate in `pkg/engine/system_init.go` and expose via `SystemInitResult`; call `WorldPersistenceSystem.LoadWorldState()` after terrain generation and `SaveWorldState()` on graceful shutdown in `cmd/server/main.go`.

- [ ] **ChallengeSystem defined but never instantiated — daily/weekly challenges disabled** — `pkg/engine/challenge_system.go:42` — `NewChallengeSystem` has zero callers outside its definition file. The struct manages daily and weekly challenge lifecycle and reset. Players have no challenge system active. — **Remediation:** Add `challengeSystem := NewChallengeSystem(game.World); game.World.AddSystem(challengeSystem)` in `pkg/engine/system_init.go`.

- [ ] **CollectionSystem defined but never instantiated — collectible tracking disabled** — `pkg/engine/collection_system.go:44` — `NewCollectionSystem` has zero callers outside its definition file. The system processes collectible discovery and completion. — **Remediation:** Instantiate in `pkg/engine/system_init.go`.

- [ ] **EconomyTerritoryIntegrationSystem defined but never instantiated — territory economic modifiers not applied** — `pkg/engine/economy_territory_system.go:81` — `NewEconomyTerritoryIntegrationSystem` has zero callers outside its file. The system bridges territory control bonuses with marketplace pricing. `EconomyTerritoryComponent` (defined at `pkg/engine/economy_territory_component.go`) is never populated with live data. — **Remediation:** Instantiate after `TerritorySystem` and `EconomySystem` are created, using `NewEconomyTerritoryIntegrationSystem(game.World, EconomyTerritoryConfig{})`.

- [ ] **HousingUI missing from shouldUpdateWorld — world runs while player places furniture** — `pkg/engine/game.go:1358` — `shouldUpdateWorld()` checks 14 UIs for pause, but `HousingUI` (`pkg/world/housing/ui.go:101` implements `IsVisible() bool`) is absent. When the housing UI is open, enemies continue attacking, NPCs pathfind, and time-of-day advances. The `updateVirtualControlsVisibility` function at line 1315 has the same omission. — **Remediation:** Add `(g.HousingUI == nil || !g.HousingUI.IsVisible())` to the `shouldUpdateWorld` boolean expression and the `anyUIOpen` expression in `updateVirtualControlsVisibility`.

- [ ] **Chunk world systems (Loader/Compression/Modification) never instantiated — persistent world chunking not active** — `pkg/world/chunk_loader.go:33`, `chunk_compression.go:14`, `chunk_modification.go:15` — `NewChunkLoaderSystem`, `NewChunkCompressionSystem`, and `NewChunkModificationSystem` have zero callers outside `pkg/world/` itself (only referenced in `doc.go` examples). The persistent chunk infrastructure exists in full but is never activated, so worlds are not streamed or persisted by chunk. — **Remediation:** Instantiate in `cmd/server/main.go` using the persistence layer that is already wired (`WorldPersistence`). Requires fixing WorldPersistenceSystem first.

- [ ] **Item pickup bypasses InventorySystem.AddItemToInventory — inconsistent behavior** — `pkg/engine/item_spawning.go:716` — `attemptItemPickup` directly appends to `inventory.Items` (`inventory.Items = append(inventory.Items, itemData.Item)`). `InventorySystem.AddItemToInventory()` is the canonical insertion path that enforces capacity, stacking rules, logging, and event emission. This creates a silent second code path for item insertion that diverges from the expected behavior. — **Remediation:** Replace the direct append with a call to `s.inventorySystem.AddItemToInventory(player, itemData.Item)` and nil-guard the `inventorySystem` reference.

---

### MEDIUM

- [ ] **PvPRewardSystem defined but never instantiated — PvP honor points disabled** — `pkg/engine/pvp_reward_system.go:78` — `NewPvPRewardSystem` has zero callers outside its file. Honor points, seasonal rewards, and tournament PvP rewards are completely disabled. `PvPRatingSystem` (the ranking half) is wired in `cmd/server/v4_systems.go:161`, so ratings accumulate but rewards are never distributed. — **Remediation:** Instantiate in `cmd/server/v4_systems.go` alongside `PvPRatingSystem` and call `world.AddSystem(pvpRewardSystem)`.

- [ ] **CityEvolutionSystem defined but never instantiated — cities never evolve** — `pkg/engine/city_evolution_system.go:25` — `NewCityEvolutionSystem` has zero callers outside its file. City evolution triggers defined in `pkg/engine/city_evolution_triggers.go` are queued but never processed. — **Remediation:** Instantiate in `pkg/engine/system_init.go` (requires a `GameClock` — pass the same clock used by other scheduled systems).

- [ ] **PostOfficeSpawner defined but never instantiated — post offices never appear in cities** — `pkg/engine/postoffice_spawning.go:19` — `NewPostOfficeSpawner` has zero callers outside its file. Courier system is wired (`cmd/server/v4_systems.go`) but has no post office buildings to spawn from, making deliveries impossible. — **Remediation:** Instantiate after `CourierSystem` in `cmd/server/v4_systems.go`; call `spawner.SpawnPostOffice(terrain)` after world generation.

- [ ] **VoiceSettingsSystem defined but never instantiated — voice settings changes not applied** — `pkg/engine/voice_settings_system.go:20` — `NewVoiceSettingsSystem` has zero callers outside its file. Player voice settings (volume, codec quality, mic device) stored in `VoiceSettingsComponent` are never applied at runtime. Note: VoiceAudioSystem and VoiceChannelSystem are wired, making this a secondary gap that only matters once VoiceTransport (Gap 1) is fixed. — **Remediation:** Instantiate in `cmd/client/handlers.go` alongside `VoiceAudioSystem` and `VoiceChannelSystem`.

- [ ] **Scene litBuffer not reset on window resize — one-frame lighting misalignment** — `pkg/engine/game.go:1722` — `Layout()` disposes and recreates `sceneBuffer` on resize but does NOT nil `litBuffer`. The next call to `drawLitScene()` detects the size mismatch at line 1546 and recreates it, but the _first_ frame after resize renders lighting to a stale-sized buffer, causing visible clipping or misalignment. — **Remediation:** Add `g.litBuffer = nil` in the resize branch of `Layout()` at line 1737, immediately after `g.sceneBuffer = ebiten.NewImage(...)`.

- [ ] **FPS benchmark too narrow — 65 systems unrepresented** — `pkg/benchmark/fps/benchmark_test.go` — Benchmark tests only `MovementSystem` with 2 000 entities. All other 65 systems, collision detection, and the rendering pipeline are absent. CI regression gate therefore only catches regressions in ECS dispatch overhead. (Existing Gap 5 — open.) — **Remediation:** See GAPS.md Gap 5.

---

### LOW

- [ ] **Respec lacks gold check — players can respec for free** — `cmd/client/handlers.go:3260` — `TODO: Check if player has enough gold for respec`. The prestige UI `SetRespecCallback` always returns `true` regardless of player gold. This bypasses the intended economic cost of skill respecs. — **Remediation:** Retrieve player's `GoldComponent`, compare against `cost`, return `false` if insufficient; display feedback message.

- [ ] **Mobile keyboard not integrated — native on-screen keyboard not triggered** — `pkg/mobile/keyboard_default.go:19` — `ShowKeyboard()` is a no-op on native mobile builds with a `TODO: Integrate native mobile keyboard APIs (UIKeyboard on iOS, InputMethodManager on Android)`. Text input fields (chat, character naming) may not trigger the OS keyboard on some mobile devices. — **Remediation:** Implement via Ebiten's Android Java bridge or iOS bridge per platform build tag; tracked as future mobile enhancement.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `AdaptiveSoundtrackSystem` not in `system_init.go` | Instantiated in `cmd/client/init_versions.go:65`. |
| `DiscoverySystem` not in `system_init.go` | Instantiated in both `cmd/server/v4_systems.go` and `cmd/client/init_versions.go`. |
| `EconomySystem` not in `system_init.go` | Instantiated in `cmd/server/main.go` and `cmd/client/handlers.go:640`. |
| `PerformanceMonitoringSystem` not in `system_init.go` | Instantiated in `cmd/server/main.go`, `cmd/client/handlers.go`, and `cmd/mobile/mobile.go`. |
| `RaidSystem` not in `system_init.go` | Instantiated in `cmd/server/v4_systems.go` and `cmd/client/handlers.go`. |
| `ReputationSystem` not in `system_init.go` | Instantiated in `cmd/server/v4_systems.go` and `cmd/client/handlers.go`. |
| `TerritorySystem` not in `system_init.go` | Instantiated in `cmd/server/v9_systems.go` and `cmd/client/init_versions.go`. |
| `VehicleSystem` not in `system_init.go` | Instantiated in `cmd/server/v4_systems.go` and `cmd/client/init_versions.go`. |
| `WorldEventsSystem` not in `system_init.go` | Instantiated in `cmd/server/v4_systems.go:269` and `cmd/client/handlers.go:1311`; `world.AddSystem()` called at `cmd/client/handlers.go:2088`. |
| `VoiceAudioSystem`/`VoiceChannelSystem` not in `system_init.go` | Both instantiated and added to world in `cmd/client/handlers.go:1067–2216`. |
| `ReverbSystem` not in `system_init.go` | Instantiated in `cmd/client/handlers.go:1056`; added to world at line 2217. |
| `SpatialVoiceSystem` not in `system_init.go` | Instantiated in `cmd/client/handlers.go`. |
| `BranchingNarrativeSystem` not in `system_init.go` | Instantiated in `cmd/client/handlers.go:1310`; added to world at line 2087. |
| `CompanionLearningSystem` not in `system_init.go` | Instantiated in `cmd/client/init_versions.go:65`; added to world at `cmd/client/handlers.go:2104`. |
| `StereoscopicSystem` not in `system_init.go` | Instantiated in `cmd/client/init_versions.go`. |
| `drawRect` `ebiten.Image` never used after frame | The finding is valid — it _is_ used for drawing but then leaked. Not a false positive. |
| `pkg/procgen/story/generator.go:303` `TODO(REM-096)` | Refactoring/code-quality TODO only; current implementation is complete and tested. Low priority. |
| `pkg/procgen/quest/types.go:286` `TODO(REM-144)` | Same — function size refactoring todo; implementation is fully correct. |
| C-001 / C-002 ECS entity staging races | Already **resolved** in GAPS.md Gap 11; `entityMu` added to `ecs.go`. |
| H-008 talent registry maps unprotected | Already **resolved** in GAPS.md Gap 12. |
| H-009 ChatUI double-activation | Already **resolved** in GAPS.md Gap 13. |
| Health endpoints missing | Already **implemented** at `pkg/observability/metrics.go:170-175`. |
| Signal handler missing | Already **implemented** at `cmd/server/main.go:106`. |
