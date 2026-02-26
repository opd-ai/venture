# Audit: github.com/opd-ai/venture/pkg/world/territory
**Date**: 2026-02-26
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `pkg/world/territory` package provides guild territory control and siege mechanics with thread-safe operations, deterministic time handling, and strong test coverage (90.8%). However, critical integration gaps exist: the package lacks serialization/deserialization for persistence, the TerritorySystem and TerritorySiegeSystem are not registered in the server's ECS loop (client-only), and there is no UI integration for players to interact with territory mechanics in-game. These gaps prevent territory warfare from functioning in multiplayer gameplay.

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
- [ ] **Server Integration Gap** — TerritorySystem and TerritorySiegeSystem are not registered in `cmd/server/main.go` or `cmd/server/v9_systems.go`. The systems exist and are registered in the client (`cmd/client/handlers.go:2116`, `cmd/client/init_versions.go:649`) but are missing from the server's ECS update loop. This prevents server-authoritative territory control and siege mechanics from functioning in multiplayer. (`cmd/server/v9_systems.go:1-100`)
- [ ] **No Persistence Layer** — Manager and Siege structs lack `Serialize()`/`Deserialize()` methods required by `ComponentSerializer` interface. Territory ownership, war declarations, defensive structures, and siege state will be lost on server restart. No integration with `pkg/saveload/` found. (`manager.go:1-499`, `siege.go:1-525`)
- [ ] **No UI Integration** — No territory UI system exists in `pkg/engine/territory_ui.go` for players to view controlled territories, declare wars, build structures, or participate in sieges. Package documentation examples show API usage but not how players interact with the feature in-game. (`doc.go:54-92`, `pkg/engine/territory_ui.go:1-200`)

### Medium Severity
- [ ] **Missing Network Sync** — Territory and Siege structs are not referenced in `pkg/network/` snapshot or replication systems. Changes to territory ownership and siege state will not propagate to connected clients in real-time, causing desynchronization in multiplayer. (`pkg/network/snapshot.go:1-500`)
- [ ] **No Cross-Server Federation** — Package lacks integration with `pkg/network/federation/` for cross-server territory control and guild wars. Documentation mentions "cross-server territory synchronization support" (doc.go:14) but no federation code exists. (`manager.go:1-499`)
- [ ] **Hard-Coded Capture Radius** — Manager initializes `captureRadius: 50.0` without configuration or genre-based adjustment. This fixed value may not suit all gameplay contexts or map scales. (`manager.go:34`)

### Low Severity
- [ ] **Unused Update Return** — `SiegeManager.Update()` does not return errors or metrics despite processing multiple sieges with phase transitions and victory calculations. Silent failures are unobservable. (`siege.go:448-483`)
- [ ] **Defensive Copy Overhead** — All getter methods (`GetTerritory`, `GetSiege`, etc.) return deep copies via `copyTerritory`/`copySiege`. While thread-safe, this adds allocation overhead for read-heavy operations (e.g., UI rendering territory lists). Consider read-only views or COW patterns. (`manager.go:71-83`, `siege.go:381-393`)
- [ ] **Missing Victory Condition** — `VictorySurrender` condition is defined (siege.go:53) but never set by any code path. No `Surrender()` method exists on Siege. (`siege.go:53`)

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
| Territory Map View | ❌ | ❌ | ❌ | No UI system exists to display controlled territories, contested zones, or siege status |
| War Declaration UI | ❌ | ❌ | ❌ | No menu for guild officers to declare/end wars or view war costs |
| Defensive Structure UI | ❌ | ❌ | ❌ | No interface for placing walls, towers, or guards in owned territories |
| Siege Participation | ❌ | ❌ | ❌ | No UI for joining active sieges or viewing capture progress |
| Territory Benefits Display | ❌ | ❌ | ❌ | No HUD element showing +10% resource / +5% XP bonuses from controlled territories |

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
The package is partially integrated into the client but critically absent from the server.

- System registration: ⚠️ — `TerritorySystem` and `TerritorySiegeSystem` are registered in `cmd/client/handlers.go:2116` and added to the client's ECS world. However, **both systems are missing from the server** (`cmd/server/main.go`, `cmd/server/v9_systems.go`). Server-authoritative gameplay requires these systems to run on the server, not the client. Client-side territory systems can cause exploits and desync.
- Component registration: N/A — Package defines data structures (Manager, Siege, Territory), not ECS components. Wrapper systems exist in `pkg/engine/territory_system.go` and `pkg/engine/territory_siege_system.go`.
- Serialize/Deserialize: ❌ — No `Serialize()`/`Deserialize()` methods on Territory, Manager, Siege, or WarDeclaration structs. Territory state will not persist across server restarts. No references to `pkg/saveload/` or `ComponentSerializer` interface.
- Network sync: ❌ — No snapshot encoding/decoding for territory state. No references in `pkg/network/snapshot.go` or replication code. Multiplayer clients will not receive territory updates.
- Genre theming: N/A — Territory mechanics are genre-agnostic (no procedural generation of content based on genre). Constants (BaseCaptureTime, WarDeclarationCost) are fixed across all genres.
- Mod compatibility: ⚠️ — Hard-coded constants (BaseCaptureTime=60, WarDeclarationCost=1000) in types.go:117-131 are not exposed to mod system. Mods cannot adjust capture speed, war costs, or structure HP without code changes.

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code; standard Go with time.Time and sync.RWMutex |
| WASM | ✅ | No syscall dependencies; compatible with browser environment |
| Mobile | ✅ | No desktop-specific APIs; works on iOS/Android |

## Recommendations
1. **[HIGH]** Add TerritorySystem and TerritorySiegeSystem to server's ECS world in `cmd/server/v9_systems.go` after guild system initialization. Territory warfare must be server-authoritative to prevent client-side cheating and ensure consistent state across all connected players.
2. **[HIGH]** Implement `Serialize()`/`Deserialize()` methods on Territory, Manager, Siege, and WarDeclaration structs. Add persistence integration in `pkg/saveload/` to save territory state alongside world chunks and guild data.
3. **[HIGH]** Create territory UI system in `pkg/engine/territory_ui.go` with:
   - Map overlay showing controlled/contested territories color-coded by guild
   - War declaration dialog for guild officers (with cost validation)
   - Defensive structure placement interface (drag-and-drop or grid placement)
   - Siege participation prompts ("Join Siege?" when entering contested territory)
   - HUD indicator showing active territory bonuses (+10% resources, +5% XP)
4. **[MED]** Add network snapshot encoding for Territory and Siege state in `pkg/network/snapshot.go`. Territory ownership changes, capture progress updates, and siege phase transitions must replicate to clients in real-time (target: <100ms latency for high-priority updates like siege start/end).
5. **[MED]** Add cross-server federation support in `pkg/network/federation/territory/` for multi-server guild wars and territory ownership sync. Implement handshake protocol for territory data exchange between federated servers.
6. **[MED]** Expose territory constants (BaseCaptureTime, WarDeclarationCost, structure HP) to mod system. Create `territory_rules.json` schema in `mods/` directory for balance customization without code changes.
7. **[LOW]** Add `Surrender()` method to Siege struct and implement surrender logic (requires guild officer confirmation, costs PeaceDeclarationCost/2, immediately ends siege with defending guild as winner).
8. **[LOW]** Make `captureRadius` configurable via Manager constructor parameter with default 50.0. Allow per-territory override for map-specific balance (e.g., larger radius for open plains, smaller for dense forests).
9. **[LOW]** Consider read-only Territory/Siege views (interfaces with getters only) to avoid deep copy overhead in UI rendering loops. Implement copy-on-write for mutable operations.
