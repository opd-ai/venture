# Housing Crafting Integration Audit

**Date**: 2026-02-16
**Package**: `pkg/integration/housing_crafting/`
**Coverage**: 98.5%

## Summary

Audited the housing_crafting integration package. Found 4 issues (2 high, 2 medium) and fixed all of them. No remaining issues.

## Issues Found & Fixed

### HIGH — SyncFromStation shared mutable references (race condition)

**File**: `housing_crafting_system.go`
**Description**: `SyncFromStation` copied `SkillBonus` map and `ActiveRecipes` slice by reference from the station to the component. Modifications to the station would silently corrupt the component data, and concurrent access could cause data races.
**Fix**: Deep copy both the map and slice during sync.

### HIGH — RegisterFacility missing HouseID validation

**File**: `station_manager.go`
**Description**: `RegisterStation` validated that HouseID was non-empty, but `RegisterFacility` did not, despite facilities also being tied to houses. This allowed orphaned facilities with no house context.
**Fix**: Added HouseID validation to `RegisterFacility` consistent with `RegisterStation`.

### MEDIUM — GetStationsByOwner/GetStationsByHouse returned internal slice references

**File**: `station_manager.go`
**Description**: These methods returned the internal slice directly. Callers could modify the returned slice (e.g., set elements to nil) and corrupt the manager's internal state.
**Fix**: Return a copy of the slice instead of the internal reference.

### MEDIUM — Missing UnregisterFacility method (API gap)

**File**: `station_manager.go`
**Description**: `RegisterFacility` existed but no `UnregisterFacility` counterpart. Facilities could only accumulate, never be removed, causing a memory leak for players who reorganize their houses.
**Fix**: Added `UnregisterFacility` method with proper cleanup from both `facilities` and `facilitiesByOwner` maps.

## Remaining Issues

None.

## Test Coverage

- 98.5% statement coverage (28 tests, 7 benchmarks)
- Deep copy behavior verified with mutation tests
- Slice copy isolation verified
- Facility lifecycle (register/unregister) fully tested
- All validation paths tested including new HouseID check
