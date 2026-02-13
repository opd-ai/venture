# Audit: github.com/opd-ai/venture/pkg/engine/physics
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The `pkg/engine/physics` parent package serves as a namespace and documentation hub for specialized physics sub-packages (vehicle, fluids, destruction). While well-documented with excellent coverage in sub-packages, the parent package contains documentation inaccuracies that misrepresent the actual architecture, and none of the physics systems are integrated into the main engine initialization.

## Issues Found
- [ ] high doc — Documentation claims non-existent subdirectories `terrain/` and `collision/` in Architecture section (`doc.go:12-14`)
- [ ] high integration — Physics systems (VehiclePhysicsSystem, FluidSimulationSystem, DestructionSystem) are not registered in `pkg/engine/system_init.go` despite being fully implemented
- [ ] med doc — Example code references non-existent constructors `physics.NewSuspensionComponent()` and `physics.NewWeightTransferComponent()` which actually exist in `vehicle` sub-package (`doc.go:23-24`)
- [ ] low doc — Documentation states "terrain/" subdirectory exists but terrain deformation is actually implemented in `vehicle/terrain_deformation.go` (`doc.go:13`)
- [ ] low doc — Documentation states "collision/" subdirectory exists but collision response is actually implemented in `vehicle/collision_response.go` (`doc.go:14`)

## Test Coverage
N/A (parent package only - no test files expected)
- `pkg/engine/physics/destruction`: 81.4%
- `pkg/engine/physics/fluids`: 95.3%
- `pkg/engine/physics/vehicle`: 94.1%

## Integration Status
**Current**: The physics package is a parent namespace with only `doc.go`. All implementation is in sub-packages (vehicle, fluids, destruction).

**Integration Points**:
- ❌ Physics systems NOT registered in `pkg/engine/system_init.go`
- ✅ Sub-packages implement proper ECS components and systems
- ✅ Sub-packages have been audited separately with minor issues
- ❌ No usage of physics components found in main engine code outside physics package

**Missing Registrations**:
- `vehicle.VehiclePhysicsSystem` should be added to system initialization
- Fluid simulation system should be added (if exists)
- Destruction system should be added (if exists)

## Recommendations
1. **HIGH PRIORITY**: Correct `doc.go` Architecture section to reflect actual structure (vehicle/, fluids/, destruction/ only - remove terrain/ and collision/ references)
2. **HIGH PRIORITY**: Register all three physics systems in `pkg/engine/system_init.go` to enable physics features in the game
3. **MEDIUM PRIORITY**: Update example code in `doc.go` to use correct import path `vehicle.NewSuspensionComponent()` instead of `physics.NewSuspensionComponent()`
4. **LOW PRIORITY**: Clarify in documentation that terrain deformation and collision response are part of vehicle physics, not separate subsystems
