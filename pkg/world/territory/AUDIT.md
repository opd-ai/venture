# Audit: pkg/world/territory
**Date**: 2026-02-16
**Status**: Complete

## Summary
The `pkg/world/territory` package provides guild territory control and warfare mechanics with siege systems. Overall health is excellent with 91.7% test coverage, comprehensive godoc, thread-safe operations, and full engine integration. The package correctly uses TimeProvider abstraction for deterministic state. Critical risk: data race potential from returning pointers to internal state without defensive copying.

## Issues Found
- [ ] <severity:high> error handling — War duration calculation uses incorrect time duration math: `now.Add(WarDurationDays * 24 * 3600 * 1e9)` should use `time.Duration` constants instead of raw nanoseconds. This will produce incorrect war duration. (`manager.go:321`)
- [ ] <severity:high> data race — `GetTerritory()`, `GetGuildTerritories()`, `GetAllTerritories()`, `GetContestedTerritories()`, `GetActiveWars()`, and `GetGuildWars()` return pointers to internal state with RLock but callers can mutate after lock release, causing race conditions. Should return defensive copies or add WARNING comments are already present but insufficient protection. (`manager.go:70-81,379-390,422-433,437-446,450-461,465-481`)
- [ ] <severity:high> data race — `GetSiege()` and `GetActiveSieges()` in SiegeManager return pointers to internal Siege state without defensive copying, allowing external mutations during concurrent access. (`siege.go:392-404,408-420`)
- [ ] <severity:med> deterministic procgen — Deprecated functions `NewSiege()`, `AdvancePhase()`, and `GenerateDefensiveStructures()` use `time.Now()` directly, violating deterministic generation standards. While properly deprecated with recommendations to use `*WithTime` variants, they should be removed to prevent accidental misuse. (`siege.go:104-106,223-225,476-478`)
- [ ] <severity:low> error handling — `RealTimeProvider.Now()` returns `time.Now()` which is expected, but package lacks validation that production code never uses deprecated non-deterministic functions. Consider adding build tags or linter rules. (`types.go:18`)

## Test Coverage
91.7% (target: 65%)

**Status**: ✅ Exceeds target by 26.7 percentage points

**Test Quality Indicators**:
- Comprehensive table-driven tests for all manager operations
- Thread safety validated via concurrent access tests
- TimeProvider abstraction enables deterministic testing
- Edge cases covered: duplicate territories, war conflicts, structure limits
- Siege phase transitions fully tested with mock time
- Test files: `manager_test.go` (382 LOC), `siege_test.go` (587 LOC), `types_test.go` (57 LOC)

## Integration Status
**Full Integration** — Package is actively used in engine and properly registered.

### Engine Integration (`pkg/engine/`)
- **TerritorySystem** (`territory_system.go`) — Main system managing territory capture and guild warfare
- **TerritorySiegeSystem** (`territory_siege_system.go`) — Wraps `SiegeManager` for ECS integration
- **TerritoryUI** (`territory_ui.go`) — UI components for territory management display

All systems properly instantiate managers and provide thread-safe access to territory state.

### Integration Pattern
```go
// Engine creates managers during initialization
territoryMgr := territory.NewManager()
siegeMgr := territory.NewSiegeManager()

// Systems wrap managers for ECS integration
territorySys := engine.NewTerritorySystem(territoryMgr, logger)
siegeSys := engine.NewTerritorySiegeSystem(siegeMgr)
```

### Missing Registrations
**None identified**. Systems are properly integrated into engine system lists.

### Persistence Integration
**Status**: Not audited (would require checking `pkg/saveload/` and `pkg/world/` persistence layer)

**Recommendation**: Verify Territory and Siege structures have Serialize/Deserialize methods for world state persistence. Current types use standard Go types (string, float64, time.Time, maps) which should serialize cleanly with encoding/json or encoding/gob.

## Deterministic Generation ⚠️
**Partial Compliance** — Uses deterministic TimeProvider pattern, but deprecated functions violate standards.

### Compliant
- ✅ TimeProvider abstraction enables deterministic timestamps (`types.go:5-24`)
- ✅ `GenerateDefensiveStructuresWithTime()` uses seed-based `rand.New(rand.NewSource(seed))` (`siege.go:483`)
- ✅ No global `rand` calls in deterministic code paths
- ✅ All production code paths use `*WithTime` variants with injected time

### Non-Compliant
- ❌ `NewSiege()` calls `time.Now()` directly (`siege.go:105`)
- ❌ `AdvancePhase()` calls `time.Now()` directly (`siege.go:224`)
- ❌ `GenerateDefensiveStructures()` calls `time.Now()` directly (`siege.go:477`)
- ⚠️ `RealTimeProvider.Now()` returns `time.Now()` but this is expected for production use (`types.go:18`)

**Impact**: Medium. Deprecated functions are clearly marked but still exist in codebase. If called by production code, they violate deterministic generation standards for multiplayer sync and testing.

**Recommendation**: Remove deprecated functions entirely or use build tags to exclude them from production builds. Add linter rule to prevent `time.Now()` calls outside TimeProvider implementations.

## Network Interface Compliance ✅
**Not Applicable** — Package does not use network types.

No network communication in this package. All networking logic is in `pkg/network/`.

## ECS Compliance ✅
**Not Applicable** — Package does not define ECS components.

This is a world management package providing territory/warfare services to systems. It does not define components or systems itself. Engine systems (`TerritorySystem`, `TerritorySiegeSystem`) that *use* this package are defined in `pkg/engine/` and maintain proper ECS architecture with component-based entity management.

## Error Handling
**Good** — Structured logging with logrus, proper error propagation, comprehensive validation.

### Strengths
- ✅ Uses `logrus.WithFields` for structured logging with contextual fields (`manager.go:44-46,76-78`)
- ✅ All public API methods return errors with clear messages (`manager.go:47,79,95`)
- ✅ Validation functions prevent invalid state transitions (`manager.go:144-149`)
- ✅ Thread-safe error handling with mutex protection
- ✅ Edge cases logged at Debug level for troubleshooting

### Gaps
- Time duration calculation uses unsafe arithmetic instead of `time.Duration` constants (HIGH severity - see Issues Found)
- No error wrapping with `%w` verb for error chain composition (could improve debugging)
- Some validation errors don't log context before returning (e.g., `validateAttackingGuild`)

**Impact**: Low. Error messages are clear and logging is comprehensive. The time duration bug is a functional correctness issue, not an error handling issue.

## Documentation Coverage ✅
**Excellent** — Comprehensive godoc with package guide and all exports documented.

- ✅ Package doc (`doc.go`) — 106 lines with detailed usage examples, mechanics explanation, thread safety notes
- ✅ All exported types have godoc comments
- ✅ All exported functions have godoc comments
- ✅ Constants documented with inline comments (`types.go:117-132`)
- ✅ Performance targets documented (`doc.go:99-105`)
- ✅ Example usage included (`doc.go:54-92`)

**Documentation Highlights**:
- Territory system overview with chunk-based zones
- Capture mechanics with formula explanation
- Defensive structures with stats
- War declaration system with costs and duration
- Thread safety guarantees
- Performance targets for Phase 50.2

**Missing Documentation**:
- Data race warnings on getter methods (WARNING comments exist but could be more prominent)
- Serialization/persistence contract (if applicable)

## Code Quality
**Excellent** — Clean architecture, thread-safe, well-tested, clear separation of concerns.

### Architecture Strengths
- Clear separation: Manager (territory state), SiegeManager (siege state), types (data structures)
- TimeProvider abstraction enables deterministic testing and state replication
- Thread-safe operations with `sync.RWMutex` for concurrent access
- Private helper methods for complex state transitions (e.g., `applyAttackerProgress`, `resetCaptureProgress`)
- Comprehensive constants for game balance tunability

### Thread Safety
- **Status**: Good with caveats
- All Manager/SiegeManager methods properly use mutex locks
- Read operations use `RLock()`, write operations use `Lock()`
- ⚠️ Returning pointers to internal state creates race condition potential (see Issues Found)

### Performance Characteristics
- Lock contention: Medium (RWMutex allows concurrent reads)
- Memory: Low (maps with pointer storage, no defensive copying)
- Complexity: O(n) for guild territory queries, O(1) for direct lookups

### Code Organization
- 6 Go files, 1,238 LOC (excluding tests)
- Logical grouping: types, manager operations, siege operations
- Helper functions properly scoped (private for implementation details)
- Clear naming conventions (Manager, SiegeManager, TimeProvider)

## Recommendations
1. **Fix war duration calculation (HIGH priority)** — Change `manager.go:321` from `now.Add(WarDurationDays * 24 * 3600 * 1e9)` to `now.Add(time.Duration(WarDurationDays) * 24 * time.Hour)` to use proper time.Duration constants and avoid nanosecond math errors.

2. **Add defensive copying for getter methods (HIGH priority)** — Modify `GetTerritory()`, `GetGuildTerritories()`, `GetAllTerritories()`, `GetSiege()`, etc. to return deep copies instead of pointers to internal state. Alternative: Add mutex-protected accessors for individual fields.
   ```go
   // Example defensive copy
   func (m *Manager) GetTerritory(id string) (*Territory, error) {
       m.mu.RLock()
       defer m.mu.RUnlock()
       territory, exists := m.territories[id]
       if !exists {
           return nil, fmt.Errorf("territory not found: %s", id)
       }
       // Deep copy to prevent external mutations
       territoryCopy := *territory
       territoryCopy.Structures = make([]*DefensiveStructure, len(territory.Structures))
       copy(territoryCopy.Structures, territory.Structures)
       return &territoryCopy, nil
   }
   ```

3. **Remove deprecated non-deterministic functions (MED priority)** — Delete `NewSiege()`, `AdvancePhase()`, and `GenerateDefensiveStructures()` entirely. Callers should use `*WithTime` variants with explicit time injection. This enforces deterministic design and prevents accidental misuse.

4. **Add linter rule for time.Now() (MED priority)** — Configure golangci-lint to forbid `time.Now()` calls outside `TimeProvider` implementations. Use `forbidigo` linter with pattern `time\.Now\(\)` restricted to `types.go`.

5. **Implement Serialize/Deserialize methods (LOW priority)** — Add persistence methods to Territory, Siege, WarDeclaration types for world state save/load. Coordinate with `pkg/saveload/` team to ensure consistent serialization format (JSON/gob).

6. **Add performance benchmarks (LOW priority)** — Create `manager_bench_test.go` with benchmarks for:
   - `CreateTerritory` / `GetTerritory` lookup performance
   - `UpdateCaptureProgress` with varying attacker/defender counts
   - `GetGuildTerritories` with varying territory counts (10, 100, 1000)
   - Lock contention under concurrent access (100 goroutines)

7. **Document data race mitigation strategy (LOW priority)** — Add section to `doc.go` explaining caller responsibilities when using returned pointers. Document that callers must not mutate returned territories/sieges and should use Manager methods for all state changes.
