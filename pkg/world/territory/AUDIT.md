# Audit: github.com/opd-ai/venture/pkg/world/territory
**Date**: 2026-02-26
**Auditor**: GitHub Copilot (META_AUDIT v2 — Re-audit #2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `pkg/world/territory` package provides guild territory control and siege mechanics with thread-safe operations, deterministic time handling via `TimeProvider` interface, and excellent test coverage (90.8%). Territory UI exists (`pkg/engine/territory_ui.go`, 451 LOC) with keyboard/touch navigation integrated into the client. **However, critical architecture violations prevent server-authoritative multiplayer**: TerritorySystem and TerritorySiegeSystem run **client-only** (not in `cmd/server/`), enabling client-side territory capture exploits. Package lacks serialization (territory state lost on restart), network snapshot encoding (no replication), and InputProvider abstraction (TerritoryUI violates Coding Guideline #2 with direct `ebiten.KeyUp` / `inpututil.IsKeyJustPressed` calls). These gaps make territory warfare a client-only feature despite appearing production-ready.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 90.8% (target: 40%) |
| `go test -race` | ✅ Pass |
| WASM vet | N/A (server-only package) |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 (siege.go:488 is seed-based) |
| Concrete net types | 0 |

## Issues Found

### High Severity
- [ ] **Server Integration Gap** — TerritorySystem and TerritorySiegeSystem are registered in the **client only** (`cmd/client/handlers.go:486,557`, `cmd/client/init_versions.go:633,649`) but **entirely absent from the server**. Search results: 0 matches for `TerritorySystem|TerritorySiegeSystem` in `cmd/server/*.go`. This violates authoritative server architecture (custom instruction: "server-authoritative gameplay"): territory capture, war declarations (1000g cost), and siege progression (3-phase: Preparation 1h → Assault 2h → Resolution → Ended) run client-side only, enabling trivial exploits: (1) client sends fake 100% capture progress packet, (2) bypasses war declaration cost by manipulating local state, (3) declares war during peacetime, (4) creates phantom sieges. Multiplayer desync inevitable: each client holds independent territory state with no server validation. (`cmd/server/main.go`, `cmd/client/system_wrappers.go:375-377`)
- [ ] **No Persistence Layer** — Manager, Siege, Territory, WarDeclaration, DefensiveStructure structs lack `Serialize()`/`Deserialize()` methods. Grep result: 0 matches for `func.*Serialize|func.*Deserialize` in `pkg/world/territory/*.go` (excluding tests). Territory state lost on server restart: ownership (string guildID), capture progress (0.0-1.0 float64 taking 60s to complete), war declarations (time.Time DeclaredAt/EndsAt, 1000g cost, 7-day duration), defensive structures (HP: wall 1000, tower/guard 500; damage: tower 100; level: guard 30; X/Y positions), siege participants (Attackers/Defenders/Reinforcements maps), siege phase timestamps (PhaseStartTime for 1h/2h transitions), loot (DefenderTreasury int, 30% LootPercentage). No `ComponentSerializer` interface implementation. No integration with `pkg/saveload/manager.go` or `pkg/saveload/migrator.go`. (`manager.go:13-20`, `siege.go:73-100`, `types.go:56-67,92-103,106-114`)
- [ ] **No Network Sync** — Territory and Siege structs absent from network replication. Grep result: 0 matches for `Territory|Siege` in `pkg/network/component_serialization.go` (380 lines) or `pkg/network/*.go`. Critical state changes never replicate: (1) Territory ownership (OwnerGuildID), (2) capture progress (CaptureProgress float64, updated 1-60s via `UpdateCaptureProgress`), (3) war declarations (DeclaredAt, EndsAt, Active bool), (4) siege phase transitions (Preparation→Assault after 1h, Assault→Resolution after 2h, tracked via PhaseStartTime), (5) siege participants joining/leaving (Attackers/Defenders maps), (6) control point capture (ControlPointsCaptured int), (7) guild hall damage (GuildHallHP float64). Result: server can hold authoritative territory state (if systems were added) but clients would never receive updates, creating two divergent states. Client UI displays stale local data while server computes different ownership. (`manager.go:108-142`, `siege.go:217-258`, `pkg/network/component_serialization.go:1-380`)

### Medium Severity
- [ ] **Direct Ebiten Input Calls** — `TerritoryUI` violates Coding Guideline #2 (Input Interface Abstraction) by calling `ebiten.IsKeyPressed`, `ebiten.KeyUp`, `ebiten.KeyDown`, `inpututil.IsKeyJustPressed` directly instead of using the `InputProvider` interface. This breaks testability (cannot use `StubInput` in tests) and prevents input rebinding. (`pkg/engine/territory_ui.go:94-100`, `pkg/engine/territory_ui.go:1-451`)
- [ ] **No Cross-Server Federation** — Package lacks integration with `pkg/network/federation/` for cross-server territory control and guild wars. Documentation mentions "cross-server territory synchronization support" (`doc.go:14`) but no federation code exists. Cross-server guilds cannot declare wars on guilds from other servers or attack/defend territories across server boundaries. (`manager.go:1-499`, `doc.go:14`)
- [ ] **Hard-Coded Capture Radius** — Manager initializes `captureRadius: 50.0` without configuration, constructor parameter, or genre-based adjustment. This fixed value cannot adapt to map scale (large open-world vs. small arena) or genre (fantasy castles vs. sci-fi installations). (`manager.go:34`)
- [ ] **time.Now() in Production Code** — `RealTimeProvider.Now()` calls `time.Now()` directly in production code (non-test path). While abstracted behind `TimeProvider` interface for testing, this violates deterministic generation principles: two servers with same seed and same inputs can produce different territory state due to clock skew. Should use `GameClock` from `pkg/engine/game_clock.go` for simulation time. (`types.go:18`)

### Low Severity
- [ ] **Unused Update Return** — `SiegeManager.Update(deltaTime float64)` does not return errors or metrics despite processing multiple sieges with phase transitions (Preparation → Assault after 1 hour, Assault → Resolution after 2 hours) and victory calculations (timeout, capture points, guild hall destruction). Silent failures in `AdvancePhaseWithTime()` are unobservable. (`siege.go:448-483`)
- [ ] **Defensive Copy Overhead** — All getter methods (`GetTerritory`, `GetSiege`, `GetGuildTerritories`, etc.) return deep copies via `copyTerritory(t *Territory)` and `copySiege(s *Siege)`. While thread-safe, this allocates on every read: UI rendering 100 territories at 60 FPS = 6000 allocations/second + GC pressure. Consider read-only interfaces (TerritoryView, SiegeView) or copy-on-write patterns. (`manager.go:71-83`, `manager.go:483-492`, `siege.go:381-393`, `siege.go:427-444`)
- [ ] **Missing Victory Condition** — `VictorySurrender` condition is defined (`siege.go:53`) but never set by any code path. No `Surrender()` method exists on `Siege` struct. `String()` implementation exists (`siege.go:58-69`) but the condition is unreachable, creating dead code. (`siege.go:53`)
- [ ] **Missing Godoc Examples** — Package `doc.go` has comprehensive overview (106 lines) but lacks runnable `Example*` functions. Territory creation, capture progress updates, war declarations, and siege flow are documented in prose but not as testable examples. (`doc.go:1-106`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Data-only package; no direct input handling |
| Mouse | N/A | Data-only package; no direct input handling |
| Gamepad | N/A | Data-only package; no direct input handling |
| Touch | N/A | Data-only package; no direct input handling |
| VR | N/A | Data-only package; no direct input handling |
| Stub/Test | ✅ | Tests use `MockTimeProvider` for deterministic time (types_test.go) |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| Territory Map View | ✅ | ⚠️ | ✅ | **UPDATED**: `TerritoryUI` exists (`pkg/engine/territory_ui.go:16-451`). Displays territory list, ownership, status, capture progress. Reachable via keybind (not specified in code). **Issue**: Direct `ebiten.KeyUp`/`KeyDown` calls instead of `InputProvider` interface (violates Coding Guideline #2). No controller/gamepad support. |
| War Declaration UI | ⚠️ | ⚠️ | ✅ | **UPDATED**: UI exists but functionality unclear. `TerritoryUI` has no explicit war declaration dialog. May be embedded in territory detail view. No validation that player is guild officer before showing war options. (`pkg/engine/territory_ui.go:1-451`) |
| Defensive Structure UI | ⚠️ | ⚠️ | ✅ | **UPDATED**: UI exists but structure placement workflow unclear. `TerritoryUI` does not show structure placement grid or building menu. `BuildDefensiveStructure()` requires X/Y coordinates but UI provides no coordinate picker. (`pkg/engine/territory_ui.go:1-451`, `manager.go:198-258`) |
| Siege Participation | ⚠️ | ⚠️ | ✅ | **UPDATED**: Siege viewing exists in `TerritoryUI` but joining workflow unclear. No "Join Siege" prompt when player enters contested territory. `Siege.JoinSiege()` requires isAttacker boolean but UI does not distinguish attacker/defender sides. (`pkg/engine/territory_ui.go:1-451`, `siege.go:160-173`) |
| Territory Benefits Display | ❌ | ❌ | ⚠️ | **UNCHANGED**: No HUD element showing +10% resource / +5% XP bonuses. `Manager.GetResourceBonus()` and `GetXPBonus()` exist but are not called by HUD system. Bonuses are calculable but invisible to players. (`manager.go:392-418`, `pkg/engine/hud_system.go:1-500`) |

## Test Coverage
**Coverage**: 90.8% (target: 40%)
- Missing test areas: None identified; coverage exceeds target by 50.8 percentage points
- Missing benchmarks: None; package focuses on correctness over performance optimization
- Table-driven test compliance: ✅ (manager_test.go, siege_test.go, types_test.go use table-driven patterns)

## Documentation Coverage
- Package `doc.go`: ✅ (106 lines with comprehensive overview, examples, and performance targets)
- Exported symbols documented: 20/20 (100%)
- Complex algorithms commented: ✅ (capture progress formula, phase transitions, loot distribution)

## Integration Status
The package is **client-only integrated**, violating authoritative server architecture for multiplayer.

- System registration: ❌ — `TerritorySystem` and `TerritorySiegeSystem` are registered in `cmd/client/handlers.go:486-557` (client fields) and `cmd/client/init_versions.go:633-649` (initialization). **Both systems are absent from `cmd/server/v9_systems.go` (83 lines, only 5 V9 managers)** and `cmd/server/main.go` (system init lines 200-600). Server-authoritative gameplay requires these systems to run on the server. **Current architecture allows client to manipulate territory state without server validation.**
- Component registration: N/A — Package defines data structures (Manager, Siege, Territory, WarDeclaration), not ECS components. Wrapper systems (`TerritorySystem`, `TerritorySiegeSystem`) exist in `pkg/engine/territory_system.go` and `pkg/engine/territory_siege_system.go` but are only added to client world, not server world.
- Serialize/Deserialize: ❌ — No `Serialize()`/`Deserialize()` methods on any structs. Territory ownership (string), capture progress (float64), war declarations (time.Time, int), structures ([]*DefensiveStructure), and siege state (maps, slices) are not serializable. No `ComponentSerializer` interface implementation. No references to `pkg/saveload/`.
- Network sync: ❌ — No snapshot encoding/decoding. Territory updates (ownership, capture progress, war status, siege phase) are not replicated. Server-authoritative state exists in memory but clients never receive it. **Results in two independent territory states: server (authoritative but silent) and client (non-authoritative but active).**
- Genre theming: N/A — Territory mechanics are genre-agnostic. Constants (BaseCaptureTime=60s, WarDeclarationCost=1000 gold, structure HP) are fixed across all genres (fantasy, sci-fi, horror, cyberpunk). No `GenreID` parameter in any function signature.
- Mod compatibility: ❌ — All constants hard-coded in `types.go:117-131` with no `ModRuleProvider` integration. Mods cannot adjust capture speed, war duration (7 days), costs (war: 1000, peace: 500), or structure stats (wall: 1000 HP, tower: 500 HP + 100 damage, guard: 500 HP + level 30) without forking code.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code; standard Go with time.Time and sync.RWMutex |
| WASM | ✅ | No syscall dependencies; compatible with browser environment |
| Mobile | ✅ | No desktop-specific APIs; works on iOS/Android |

## Recommendations
1. **[HIGH]** Add TerritorySystem and TerritorySiegeSystem to server's ECS world in `cmd/server/v9_systems.go` after `initializeV9SystemsServer()` (line 34). Register in `cmd/server/main.go` system init block (lines 200-600) after guild manager creation. **Critical for server-authoritative multiplayer**: capture progress must be validated server-side to prevent exploits (fake 100% capture, instant wars). Example: `world.AddSystem(engine.NewTerritorySystem(territoryManager, serverLogger))`.
2. **[HIGH]** Implement `Serialize()`/`Deserialize()` methods on Territory, Manager, Siege, WarDeclaration, and DefensiveStructure structs. Use `encoding/gob` for binary serialization or `encoding/json` for human-readable format. Integrate with `pkg/saveload/manager.go` to save territory state alongside world chunks (existing pattern: `SaveWorld()` → component serialization). **Data to persist**: ownership, capture progress, war declarations (DeclaredAt, EndsAt, Active), structures (HP, damage, position), siege participants (Attackers map, Defenders map, Reinforcements map), siege phase + timestamps.
3. **[HIGH]** Add network snapshot encoding in `pkg/network/snapshot.go`. Create `EncodeTerritoryState()` and `DecodeTerritoryState()` functions. Include in snapshot packet: territory ID, owner guild ID, status (neutral/owned/contested), capture progress, capturing guild, last update time. **Priority tiers**: High (siege start/end, war declaration) = immediate broadcast; Medium (capture progress) = every 5 seconds; Low (structure HP) = on-damage events. Use delta compression: send only changed fields.
4. **[MED]** Refactor `TerritoryUI` to use `InputProvider` interface instead of direct `ebiten.KeyUp`/`inpututil.IsKeyJustPressed` calls. Replace `handleKeyboardInput()` (lines 92-108) with `HandleInput(input InputProvider) bool` method. Use `input.IsMenuUpJustPressed()`, `input.IsMenuDownJustPressed()` for navigation. Add gamepad/controller support via `InputProvider` abstraction. Create `StubInput`-based unit tests for UI navigation logic.
5. **[MED]** Add cross-server federation support. Create `pkg/network/federation/territory/` package with `SyncManager` struct. Implement `BroadcastTerritoryUpdate(territoryID, ownerGuildID, status)`, `BroadcastWarDeclaration(attackerGuild, defenderGuild)`, `HandleIncomingSiegeJoin(playerID, siegeID, isAttacker)`. Use existing federation handshake pattern from `pkg/network/federation/sync.go`. Add circuit breaker for failed cross-server territory sync (fallback: local-only sieges).
6. **[MED]** Expose territory constants to mod system. Integrate `ModRuleProvider` interface (defined in `pkg/engine/interfaces.go:694-708`) into Manager constructor. Replace `BaseCaptureTime` with `modRules.GetRuleFloat64("territory.baseCaptureTime", 60.0)`. Create `mods/example_territory_rules.json` with schema: `{"rules": {"territory.baseCaptureTime": 120, "territory.warCost": 5000, "territory.wallHP": 2000}}`. Add validation: capture time > 0, costs ≥ 0.
7. **[MED]** Replace `RealTimeProvider.Now()` direct `time.Now()` call with `GameClock` interface (`pkg/engine/game_clock.go`). Pass `GameClock` to Manager constructor instead of `TimeProvider`. Use simulation time for deterministic replays and testing. Benefits: (1) servers with same seed + same inputs = identical territory state, (2) fast-forward time in tests without sleep, (3) rollback/replay for debugging desync.
8. **[LOW]** Add `Surrender()` method to Siege struct: `func (s *Siege) Surrender(surrenderingGuild string, now time.Time) error`. Validate: (1) surrenderingGuild is attacker or defender, (2) phase is Preparation or Assault, (3) guild officer authorized surrender (requires guild system integration). Cost: `PeaceDeclarationCost / 2 = 250 gold`. Set `s.VictoryCondition = VictorySurrender`, `s.WinnerGuildID = oppositeGuild`, `s.Phase = PhaseResolution`. Log surrender with structured logging: `log.WithFields(log.Fields{"siege_id": s.ID, "surrendering_guild": surrenderingGuild, "winner_guild": s.WinnerGuildID}).Info("siege ended by surrender")`.
9. **[LOW]** Make `captureRadius` configurable. Add parameter to `NewManagerWithTimeProvider(tp TimeProvider, captureRadius float64)`. Default: 50.0. Add per-territory override: `Territory.CustomCaptureRadius *float64` (pointer for optional). Update `UpdateCaptureProgress()` to check `if territory.CustomCaptureRadius != nil { radius = *territory.CustomCaptureRadius } else { radius = m.captureRadius }`. Use genre-specific defaults: fantasy (75.0), sci-fi (40.0), horror (30.0).
10. **[LOW]** Add godoc examples in `example_test.go`. Functions: `ExampleManager_CreateTerritory()`, `ExampleManager_UpdateCaptureProgress()`, `ExampleManager_DeclareWar()`, `ExampleSiege_JoinSiege()`, `ExampleSiegeManager_CreateSiege()`. Use `MockTimeProvider` for deterministic timestamps. Include output comments for testable examples.
