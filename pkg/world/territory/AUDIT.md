# Audit: pkg/world/territory
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
Territory control and siege mechanics package providing guild warfare, defensive structures, and capture mechanics. Package is well-implemented with 93.9% test coverage and comprehensive functionality. Primary issues are non-deterministic time usage (unsuitable for procgen), missing ECS component registration, and lack of structured logging. No stub/incomplete code or critical functionality gaps detected.

## Issues Found
- [ ] **high** deterministic procgen — `time.Now()` used for timestamps in manager and siege systems (`manager.go:46,84,103,213,261,298` and `siege.go:102,189,355,419`). While timestamps for game state are appropriate, package lacks deterministic seed support for `GenerateDefensiveStructures()` which uses `time.Now()` for `ConstructedAt` field. This prevents predictable structure generation replays.
- [ ] **high** integration points — Territory system not registered in `pkg/engine/system_init.go`. No system wrapper exists to integrate territory/siege management into World update loop. Package is isolated library with no game engine integration.
- [ ] **med** error handling — No structured logging with `logrus.WithFields`. Package has no logging at all for operations that could benefit from observability (war declarations, territory captures, siege events). Only example usage in `doc.go` shows `log.Fatal()` pattern.
- [ ] **med** integration points — No component types defined for Territory/Siege ECS integration. Territory and Siege are standalone structs, not ECS components with `Type()` method. Cannot attach territory data to entities or serialize for persistence.
- [ ] **low** doc coverage — `Manager.GetTerritory()`, `Manager.GetGuildTerritories()`, `Manager.GetContestedTerritories()`, `Manager.GetAllTerritories()`, `Manager.GetActiveWars()`, `Manager.GetGuildWars()`, `SiegeManager.GetSiege()`, `SiegeManager.GetActiveSieges()`, and `SiegeManager.GetSiegeForTerritory()` return pointers to internal state with warnings in godoc comments, but no safe copy mechanism provided. Callers can accidentally mutate internal state.
- [ ] **low** deterministic procgen — `GenerateDefensiveStructures()` uses seeded RNG correctly but uses `time.Now()` for `ConstructedAt` field, breaking determinism for full struct comparison (`siege.go:419`).

## Test Coverage
93.9% (target: 65%) ✅

Excellent test coverage with comprehensive table-driven tests for all core functionality:
- `types_test.go`: Type string methods and constants validation
- `manager_test.go`: Territory management, war declarations, structure building, capture mechanics (27 tests + 4 benchmarks)
- `siege_test.go`: Siege phases, participants, reinforcements, victory conditions, loot distribution (21 tests + 4 benchmarks)

All tests follow Go best practices with proper error checking and edge case coverage.

## Integration Status
**Standalone library** — Not integrated with game engine ECS architecture.

Current state:
- ✅ Package is fully functional as independent territory/siege management library
- ✅ Thread-safe with `sync.RWMutex` for concurrent access
- ❌ No registration in `pkg/engine/system_init.go` (no TerritorySystem or SiegeSystem)
- ❌ No ECS component wrappers (TerritoryComponent, SiegeComponent)
- ❌ No serialization support for save/load
- ⚠️ Referenced by `pkg/integration/political_warfare` and `pkg/integration/guild_vehicle` but as planned dependency, not yet connected
- ⚠️ Referenced in `pkg/world/economy/README.md` as related package

Missing integrations:
1. **System registration**: Need `TerritorySystem` and `SiegeSystem` implementing `System` interface with `Update()` method
2. **Component types**: Need `TerritoryComponent` and `SiegeComponent` with `Type()` string and `Serialize()`/`Deserialize()` methods
3. **Entity linkage**: Territories should be attachable to world grid entities or zone entities
4. **Persistence**: No save/load support for territory ownership or active sieges
5. **Event system**: No events fired for territory capture, war declarations, siege outcomes

## Recommendations
1. **Add TerritorySystem/SiegeSystem** (high priority): Create ECS system wrappers that call Manager/SiegeManager methods and register in `system_init.go`. System should emit events for territory changes and integrate with guild, housing, and economy systems.
2. **Add structured logging** (high priority): Replace silent operations with `logrus.WithFields` logging for war declarations, territory captures, siege phase transitions, and structure destruction. Use field names: `territory_id`, `guild_id`, `siege_id`, `phase`, `winner_guild_id`.
3. **Fix time.Now() in GenerateDefensiveStructures()** (medium priority): Accept `constructionTime time.Time` parameter instead of using `time.Now()`. This preserves determinism while allowing callers to set appropriate timestamps.
4. **Add TerritoryComponent/SiegeComponent** (medium priority): Create ECS components wrapping Manager state for entity attachment and persistence. Components should be pure data with Manager holding logic.
5. **Implement defensive copying** (low priority): Add `Copy()` methods to Territory and Siege structs, update getter methods to return copies instead of pointers to prevent accidental mutation.
6. **Add integration tests** (low priority): Create integration tests with guild system, housing system, and economy system to validate cross-package interactions mentioned in doc.go.
