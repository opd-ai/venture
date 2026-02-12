# Audit: pkg/integration/companion_housing
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The companion_housing package integrates companion AI with player housing for bedding, training, and storage features. Test coverage is excellent (93.3%), ECS compliance is solid, but the package has 4 issues: non-deterministic time usage in manager methods, missing component serialization for persistence, deprecated system code still present, and lack of structured logging.

## Issues Found
- [ ] **severity:high** deterministic procgen — `time.Now()` used in `RecordRest()` method violates deterministic gameplay requirement (`pet_home_manager.go:128`)
- [ ] **severity:high** deterministic procgen — `time.Now()` used in `StartTrainingSession()` method violates deterministic gameplay requirement (`pet_home_manager.go:172`)
- [ ] **severity:med** integration points — `CompanionHousingComponent` missing `Serialize()`/`Deserialize()` methods for save/load persistence (`component.go:11-24`)
- [ ] **severity:low** stub/incomplete code — `CompanionHousingSystem` marked deprecated but still present in codebase, should be removed or un-deprecated (`companion_housing_system.go:12-17`)
- [ ] **severity:low** error handling — No structured logging with `logrus.WithFields` on error paths in manager methods (`pet_home_manager.go:68-83, 122-133, 163-174`)

## Test Coverage
93.3% (target: 65%) ✅

## Integration Status
The package is actively integrated into the game:
- Imported by `cmd/client/handlers.go` for client-side UI
- Imported by `cmd/server/v9_systems.go` for server-side companion management
- Used in V9 validation tests (`cmd/server/v9_validation_test.go`)
- Integration tests present (`cmd/server/companion_housing_integration_test.go`)

**Component Registration:** `CompanionHousingComponent` implements `Type()` method correctly for ECS integration ✅

**Serialization Status:** Component lacks `Serialize()`/`Deserialize()` methods - housing assignments will not persist across save/load ❌

**System Registration:** `CompanionHousingSystem` is deprecated; `PetHomeManager` is the primary interface used by other systems ✅

## Recommendations
1. **Fix time.Now() violations** — Replace `time.Now()` calls with injected time parameter or use deterministic time source from game clock. Change `RecordRest(companionID uint64)` to `RecordRest(companionID uint64, now time.Time)` and `StartTrainingSession(companionID, furnitureID)` to `StartTrainingSession(companionID, furnitureID string, now time.Time)`.
2. **Add component serialization** — Implement `Serialize()` and `Deserialize()` methods on `CompanionHousingComponent` to support save/load persistence. Include fields: `OwnerHouseID`, `BeddingID`, `LastRestTime`, `ActiveTraining`, `SharedChestAccess`.
3. **Remove deprecated system** — Either remove `CompanionHousingSystem` (lines 16-71 in companion_housing_system.go) or update documentation to clarify its purpose. Current deprecated warning suggests removal.
4. **Add structured logging** — Use `logrus.WithFields()` in error return paths: `AssignCompanionToBed` (line 74), `RecordRest` (line 132), `StartTrainingSession` (line 169).
