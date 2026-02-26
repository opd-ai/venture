# Audit: github.com/opd-ai/venture/pkg/engine/physics/vehicle
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
The `pkg/engine/physics/vehicle` package provides advanced vehicle physics simulation including suspension dynamics (spring-damper model), weight transfer during acceleration/braking/turning, terrain deformation (tire tracks), and collision response. The package demonstrates **excellent engineering quality** with 94.8% test coverage, zero race conditions, and comprehensive documentation. However, **one critical integration gap exists**: `EnhancedVehicleSystem` is initialized in `cmd/client/init_versions.go:423` but **never registered with the ECS World**, rendering the entire vehicle physics subsystem dead code. Additionally, multiple components violate ECS purity guidelines by containing getter/setter methods that should reside in the system layer.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 94.8% (target: 40%, or 30% for X11/Wayland/Ebiten-dependent packages) |
| `go test -race` | ✅ Pass |
| WASM vet | ✅ Pass |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 occurrences |
| Concrete net types | 0 occurrences |

## Issues Found

### High Severity
- [ ] **Integration** — `EnhancedVehicleSystem` initialized but never registered in World (`cmd/client/handlers.go:457` declares field, `cmd/client/init_versions.go:423` initializes it, but no `game.World.AddSystem(sys.enhancedVehicleSys)` call exists anywhere in `cmd/client/`) — **CRITICAL: Entire vehicle physics subsystem is dead code, system Update() never ticks**

### Medium Severity
- [ ] **ECS Purity** — `SuspensionComponent` contains getter/setter methods (`GetWheelLoad`, `GetWheelCompression`, `IsWheelGrounded`, `GetGroundedWheelCount`, `SetWheelLoad` in `suspension.go:85-125`) violating ECS guideline that components must be pure data structures — should be direct field access or system methods
- [ ] **ECS Purity** — `WeightTransferComponent` contains getter methods (`GetWheelWeights`, `GetFrontAxleWeight`, `GetRearAxleWeight`, `GetLeftSideWeight`, `GetRightSideWeight`, `GetTransferMagnitude`, `Reset` in `weight_transfer.go:61-109`) violating ECS guideline — should be direct field access or system methods
- [ ] **ECS Purity** — `CollisionResponseComponent` contains getter/setter methods (`GetDamageMultiplier`, `IsDestroyed`, `GetIntegrity`, `Repair`, `Reset`, `GetCollisionCount`, `GetLastImpactForce`, `GetLastImpactVelocity`, `ShouldCauseDamage` in `collision_response.go:42-95`) violating ECS guideline — should be direct field access or system methods
- [ ] **ECS Purity** — `TerrainDeformationComponent` contains methods with logic (`GetVisibleTracks`, `GetTrackAlpha`, `Clear`, `GetTrackCount` in `terrain_deformation.go:71-115`) violating ECS guideline — should be pure data + system methods handle logic

### Low Severity
- [ ] **Documentation** — Package doc.go is comprehensive but `GetTerrainTypeFromTile()` helper function in `terrain_deformation.go:119-134` lacks usage examples in doc
- [ ] **Performance** — `GetVisibleTracks()` uses AABB culling (`terrain_deformation.go:82-86`) but could benefit from spatial partitioning for >1000 tracks (current max is 200, acceptable for now)
- [ ] **Integration Gap** — No integration with `pkg/procgen/vehicle` generator to automatically add physics components to generated vehicles — vehicle entities created without physics components
- [ ] **Test Gap** — No integration test verifying all four components work together in a real ECS World with actual entity lifecycle (current tests use mock entities)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Vehicle physics system does not handle input directly (movement system consumes input, vehicle physics responds to velocity changes) |
| Mouse | N/A | No direct mouse input handling |
| Gamepad | N/A | No direct gamepad input handling |
| Touch | N/A | No direct touch input handling |
| VR | N/A | No VR-specific input handling |
| Stub/Test | ✅ | Uses mock components (MockPositionComponent, MockVelocityComponent, MockRotationComponent) in `system_test.go:348-372` for testing without Ebiten |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Vehicle physics system has no UI — visual feedback handled by render system showing suspension compression and track marks |

## Test Coverage
**Coverage**: 94.8% (target: 40%, or 30% for X11/Wayland/Ebiten-dependent packages)
- Missing test areas: None significant — all major code paths covered
- Missing benchmarks: Benchmarks exist for all hot-path functions (suspension, weight transfer, collision, terrain)
- Table-driven test compliance: ✅ — All tests use table-driven patterns (e.g., `system_test.go:14-47`, `suspension_test.go:14-54`, `weight_transfer_test.go:14-62`)

## Documentation Coverage
- Package `doc.go`: ✅ — Comprehensive 210-line package documentation with usage examples, component descriptions, terrain types, performance considerations
- Exported symbols documented: 43/43 (100%) — All exported types, functions, constants have godoc comments
- Complex algorithms commented: ✅ — Spring-damper physics (`system.go:320-375`), weight transfer calculations (`system.go:533-646`), collision response with force calculation (`system.go:378-443`), velocity reflection formula (`system.go:445-460`)

## Integration Status
The vehicle physics package integrates with ECS architecture but has a **critical registration failure**.

- System registration: ❌ — **EnhancedVehicleSystem initialized but never registered in World** (see High Severity issue) — `cmd/client/handlers.go:457` declares field, `cmd/client/init_versions.go:423` initializes it, but no `game.World.AddSystem()` call
- Component registration: ⚠️ — Components define `Type()` string returning unique IDs (`suspension`, `weight_transfer`, `terrain_deformation`, `collision_response`), but **no verification that these are registered in ECS component type registry**
- Serialize/Deserialize: ❌ — None of the four components implement `Serialize()/Deserialize()` methods, making them **non-persistent** — vehicle physics state lost on save/load
- Network sync: ❌ — Components not listed in snapshot system, no delta compression, no client-side prediction handling — **multiplayer physics will desync**
- Genre theming: N/A — Vehicle physics is genre-agnostic (same suspension model for fantasy horses, sci-fi mechs, cyberpunk cars)
- Mod compatibility: ✅ — Physics parameters exposed as component fields (SpringStiffness, DamperStrength, DamageThreshold, etc.) — mods can override values via JSON

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | No platform-specific code, pure Go math and data structures |
| WASM | ✅ | Passes WASM vet check, no syscalls or filesystem access |
| Mobile | ✅ | No touch-specific handling needed (vehicle physics runs on server/client tick) |

## Recommendations
1. **[HIGH]** Register `EnhancedVehicleSystem` with World in `cmd/client/main.go` after `sys.enhancedVehicleSys = vehicle.NewEnhancedVehicleSystem()` initialization — add `game.World.AddSystem(sys.enhancedVehicleSys)` in the critical systems registration section (around line 150 near other physics systems)
2. **[HIGH]** Implement `Serialize()/Deserialize()` for all four components to enable save/load persistence — use encoding/gob or JSON for Wheel arrays, TrackMark slices, and scalar fields
3. **[HIGH]** Add network sync support: list components in snapshot system (`pkg/network/snapshot.go`), implement delta compression for large Tracks slices, add client-side prediction awareness
4. **[MED]** Refactor component getter/setter methods to comply with ECS purity guidelines — convert methods to direct field access for simple getters, move logic methods (Reset, GetVisibleTracks, GetTrackAlpha, ShouldCauseDamage) into EnhancedVehicleSystem as helper functions
5. **[MED]** Add integration with `pkg/procgen/vehicle` generator to automatically attach physics components (suspension, weight transfer, collision response, terrain deformation) to generated vehicle entities based on vehicle type (mount, cart, boat, glider, mech)
6. **[LOW]** Add integration test creating real World, spawning entity with vehicle components, running system Update(), verifying state changes — current tests use mocks
7. **[LOW]** Add spatial partitioning to `GetVisibleTracks()` if MaxTracks increases beyond 200 (current AABB culling O(n) is acceptable for ≤200 tracks)
8. **[LOW]** Document `GetTerrainTypeFromTile()` helper with usage examples showing terrain system integration
