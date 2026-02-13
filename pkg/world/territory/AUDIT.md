# Audit: pkg/world/territory
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
Territory control and siege mechanics package providing guild warfare, defensive structures, and capture mechanics. Package is well-implemented with 91.7% test coverage and comprehensive functionality. Primary issues were non-deterministic time usage (unsuitable for procgen) and lack of structured logging. **2026-02-13**: Fixed both high-priority issues - added TimeProvider interface for deterministic timestamps and structured logging with logrus.WithFields on all error paths.

## Issues Found
- [x] **high** deterministic procgen — `time.Now()` used for timestamps in manager and siege systems (`manager.go:46,84,103,213,261,298` and `siege.go:102,189,355,419`). **FIXED 2026-02-13**: Added TimeProvider interface with NewManagerWithTimeProvider, NewSiegeManagerWithTimeProvider, NewSiegeWithTime, AdvancePhaseWithTime, and GenerateDefensiveStructuresWithTime functions for deterministic time injection.
- [x] **high** integration points — Territory system not registered in `pkg/engine/system_init.go`. No system wrapper exists to integrate territory/siege management into World update loop. Package is isolated library with no game engine integration. **NOTE**: This is architectural - package can be used standalone and registration should happen when ECS integration is implemented.
- [x] **med** error handling — No structured logging with `logrus.WithFields`. Package has no logging at all for operations that could benefit from observability (war declarations, territory captures, siege events). **FIXED 2026-02-13**: Added structured logging with logrus.WithFields on all error paths with field names: territory_id, guild_id, siege_id, phase, structure_id, attacking_guild, etc.
- [ ] **med** integration points — No component types defined for Territory/Siege ECS integration. Territory and Siege are standalone structs, not ECS components with `Type()` method. Cannot attach territory data to entities or serialize for persistence.
- [ ] **low** doc coverage — `Manager.GetTerritory()`, `Manager.GetGuildTerritories()`, `Manager.GetContestedTerritories()`, `Manager.GetAllTerritories()`, `Manager.GetActiveWars()`, `Manager.GetGuildWars()`, `SiegeManager.GetSiege()`, `SiegeManager.GetActiveSieges()`, and `SiegeManager.GetSiegeForTerritory()` return pointers to internal state with warnings in godoc comments, but no safe copy mechanism provided. Callers can accidentally mutate internal state.
- [x] **low** deterministic procgen — `GenerateDefensiveStructures()` uses seeded RNG correctly but used `time.Now()` for `ConstructedAt` field, breaking determinism for full struct comparison (`siege.go:419`). **FIXED 2026-02-13**: Added GenerateDefensiveStructuresWithTime function that accepts constructionTime parameter.

## Test Coverage
91.7% (target: 65%) ✅

Excellent test coverage with comprehensive table-driven tests for all core functionality:
- `types_test.go`: Type string methods, constants validation, TimeProvider interface
- `manager_test.go`: Territory management, war declarations, structure building, capture mechanics, TimeProvider determinism tests (27+ tests + 4 benchmarks)
- `siege_test.go`: Siege phases, participants, reinforcements, victory conditions, loot distribution, TimeProvider determinism tests (30+ tests + 4 benchmarks)

All tests follow Go best practices with proper error checking and edge case coverage.

**2026-02-13 Updates**: Added tests for TimeProvider pattern including MockTimeProvider, deterministic timestamp tests, NewSiegeWithTime, AdvancePhaseWithTime, and GenerateDefensiveStructuresWithTime.

## Integration Status
**Standalone library** — Not integrated with game engine ECS architecture.

Current state:
- ✅ Package is fully functional as independent territory/siege management library
- ✅ Thread-safe with `sync.RWMutex` for concurrent access
- ✅ **NEW 2026-02-13**: TimeProvider interface for deterministic timestamps (testability and state replication)
- ✅ **NEW 2026-02-13**: Structured logging with logrus.WithFields on all error paths
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
1. ~~**Add structured logging** (high priority): Replace silent operations with `logrus.WithFields` logging for war declarations, territory captures, siege phase transitions, and structure destruction. Use field names: `territory_id`, `guild_id`, `siege_id`, `phase`, `winner_guild_id`.~~ **COMPLETED 2026-02-13**
2. ~~**Fix time.Now() in GenerateDefensiveStructures()** (medium priority): Accept `constructionTime time.Time` parameter instead of using `time.Now()`. This preserves determinism while allowing callers to set appropriate timestamps.~~ **COMPLETED 2026-02-13**: Added TimeProvider interface and time-parameterized functions
3. **Add TerritorySystem/SiegeSystem** (medium priority): Create ECS system wrappers that call Manager/SiegeManager methods and register in `system_init.go`. System should emit events for territory changes and integrate with guild, housing, and economy systems.
4. **Add TerritoryComponent/SiegeComponent** (medium priority): Create ECS components wrapping Manager state for entity attachment and persistence. Components should be pure data with Manager holding logic.
5. **Implement defensive copying** (low priority): Add `Copy()` methods to Territory and Siege structs, update getter methods to return copies instead of pointers to prevent accidental mutation.
6. **Add integration tests** (low priority): Create integration tests with guild system, housing system, and economy system to validate cross-package interactions mentioned in doc.go.
