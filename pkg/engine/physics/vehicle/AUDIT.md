# Audit: github.com/opd-ai/venture/pkg/engine/physics/vehicle
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The vehicle physics package implements advanced suspension, weight transfer, collision response, and terrain deformation systems with excellent test coverage (94.1%) and well-structured code. However, it suffers from **critical ECS architecture violations** where components contain business logic methods instead of being pure data structures, and the entire package appears to be **not integrated** into the actual game engine - components are never instantiated and the EnhancedVehicleSystem is never invoked in production code.

## Issues Found
- [x] **high** ECS Compliance — Components violate ECS purity by containing logic methods beyond Type(). All four components (SuspensionComponent, WeightTransferComponent, CollisionResponseComponent, TerrainDeformationComponent) have Update(), ProcessCollision(), AddTrack(), GetWheelLoad(), etc. methods. These should be pure data with logic moved to systems. (`collision_response.go:48`, `suspension.go:90`, `weight_transfer.go:67`, `terrain_deformation.go:71`)
- [x] **high** Integration Status — EnhancedVehicleSystem is created in cmd/client/handlers.go:1191 but NEVER called in any Update loop. Components are NEVER instantiated in production code (only in tests/doc examples). Package is effectively unused despite being fully implemented. (`cmd/client/handlers.go:296`, `cmd/client/handlers.go:1191`)
- [x] **high** Integration Status — No component type registration in engine's entity spawning systems. Components with types "suspension", "weight_transfer", "collision_response", "terrain_deformation" are not referenced anywhere in pkg/engine/ or cmd/. (no file references found)
- [x] **med** ECS Compliance — SuspensionComponent.Update() performs physics calculations directly on component state. This should be in a SuspensionSystem that operates on entities with suspension components. (`suspension.go:90-163`)
- [x] **med** ECS Compliance — WeightTransferComponent.Update() calculates and stores weight distribution. This business logic should be in a WeightTransferSystem. (`weight_transfer.go:67-97`)
- [x] **med** ECS Compliance — CollisionResponseComponent.ProcessCollision() contains collision physics logic including damage calculation and velocity reflection. Should be in CollisionSystem. (`collision_response.go:48-124`)
- [x] **med** ECS Compliance — TerrainDeformationComponent.AddTrack() and Update() manage track marks directly. Should be in TerrainDeformationSystem. (`terrain_deformation.go:71-151`)
- [x] **low** Deterministic Procgen — TerrainDeformationComponent correctly uses seeded rand.New(rand.NewSource(seed)) for deterministic track noise. No issues found. (`terrain_deformation.go:44`)
- [x] **low** Error Handling — No error returns in package API. Components use defensive programming (bounds checks, nil checks) but errors are silently handled. Consider returning errors for invalid inputs. (`suspension.go:91-94`, `weight_transfer.go:69`)
- [x] **low** Doc Coverage — Package has excellent doc.go with comprehensive examples. All exported types/functions have godoc comments. (`doc.go:1-201`)

## Test Coverage
94.1% (target: 65%) ✅ **EXCEEDS TARGET**

Test files provide comprehensive coverage with table-driven tests for all components:
- `suspension_test.go`: Tests suspension physics, wheel grounding, compression
- `weight_transfer_test.go`: Tests weight distribution during acceleration/braking/turning
- `collision_response_test.go`: Tests damage calculation, velocity reflection, structural integrity
- `terrain_deformation_test.go`: Tests track mark creation, fading, culling, determinism
- `system_test.go`: Tests integrated vehicle physics system

## Integration Status
**NOT INTEGRATED** ⚠️

The package is created but not used:
1. ✅ EnhancedVehicleSystem instantiated in `cmd/client/handlers.go:1191`
2. ❌ System never called in Update loops (neither client nor server)
3. ❌ Components never created in entity spawning code
4. ❌ No integration with existing VehicleSystem (`pkg/engine/vehicle_system.go`) which handles mounting/fuel but not physics
5. ❌ Component types ("suspension", "weight_transfer", etc.) not registered in system_init.go

**Integration Path**: 
- Components need to be added to vehicle entities in `pkg/procgen/vehicle/` or `pkg/engine/vehicle_spawning.go`
- EnhancedVehicleSystem needs to be called in game update loop (likely in VehicleSystem.Update())
- Alternatively, if component-based approach is kept, systems should be registered in pkg/engine/system_init.go

## Recommendations
1. **[CRITICAL] Refactor to ECS purity**: Move all logic methods from components to systems. Components should only have Type() and serialization methods. Create separate systems: SuspensionSystem, WeightTransferSystem, CollisionSystem, TerrainDeformationSystem that operate on entities with these components.

2. **[CRITICAL] Integrate into engine**: Wire up EnhancedVehicleSystem to be called in the game update loop. Add component creation to vehicle entity spawning. Register component types in system_init.go for proper ECS integration.

3. **[HIGH] Decide on architecture**: Choose between:
   - **Option A**: Keep EnhancedVehicleSystem as a facade that directly operates on components (current design), but components should still be pure data
   - **Option B**: Break into separate ECS systems (SuspensionSystem, WeightTransferSystem, etc.) that each process entities with specific components following standard ECS pattern

4. **[MEDIUM] Add error handling**: Return errors from component methods instead of silent failures (e.g., Update() with invalid terrain heights array, AddTrack() with invalid terrain type).

5. **[LOW] Add integration tests**: Create tests that verify components work with actual Entity/World from pkg/engine/ecs.go, not just unit tests of components in isolation.
