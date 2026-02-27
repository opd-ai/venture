# Audit: github.com/opd-ai/venture/pkg/engine/physics
**Date**: 2026-02-25
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
The `pkg/engine/physics` package provides advanced physics simulation for vehicles (suspension, weight transfer, terrain deformation, collision response), fluid dynamics (buoyancy, swimming, flooding), and environmental destruction (structural integrity, debris). The package achieves excellent test coverage (94.8%-96.3%) and passes all automated checks (vet, race detector). However, **critical integration gaps** exist: the `EnhancedVehicleSystem` is initialized in `cmd/client/init_versions.go` but **never registered with the ECS World**, making it dead code. Additionally, multiple components contain getter/setter methods that violate ECS purity guidelines.

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass |
| `go test -cover` | 95.4% average (destruction: 96.3%, fluids: 95.2%, vehicle: 94.8%) — exceeds 40% target |
| `go test -race` | ✅ Pass |
| WASM vet | N/A (no WASM-specific code in this package) |
| TODO/FIXME count | 0 |
| Non-deterministic rand | 0 (all use `rand.New(rand.NewSource(seed))`) |
| Concrete net types | 0 |

## Issues Found

### High Severity
- [x] **Integration** — `EnhancedVehicleSystem` initialized but never registered in World (`cmd/client/handlers.go` line 457 declares field, `cmd/client/init_versions.go` initializes it, but no `game.World.AddSystem(sys.enhancedVehicleSys)` call exists) — **CRITICAL: Dead code, system never ticks**
- [x] **ECS Compliance** — Components contain getter/setter methods beyond `Type()`, violating ECS purity (see Medium Severity for full list) — While these are read-only getters, the guideline states components should be "pure data structures with only a `Type() string` method"

### Medium Severity
- [x] **ECS Compliance** — `SuspensionComponent` has 5 methods beyond `Type()`: `GetWheelLoad()`, `GetWheelCompression()`, `IsWheelGrounded()`, `GetGroundedWheelCount()`, `SetWheelLoad()` (`vehicle/suspension.go:85-125`)
- [x] **ECS Compliance** — `CollisionResponseComponent` has 9 methods beyond `Type()`: `GetDamageMultiplier()`, `IsDestroyed()`, `GetIntegrity()`, `Repair()`, `Reset()`, `GetCollisionCount()`, `GetLastImpactForce()`, `GetLastImpactVelocity()`, `ShouldCauseDamage()` (`vehicle/collision_response.go:42-95`)
- [x] **ECS Compliance** — `WeightTransferComponent` has 6 methods beyond `Type()`: `GetWheelWeights()`, `GetFrontAxleWeight()`, `GetRearAxleWeight()`, `GetLeftSideWeight()`, `GetRightSideWeight()`, `GetTransferMagnitude()`, `Reset()` (`vehicle/weight_transfer.go:61-109`)
- [x] **ECS Compliance** — `TerrainDeformationComponent` has 4 methods beyond `Type()`: `GetVisibleTracks()`, `GetTrackAlpha()`, `Clear()`, `GetTrackCount()` (`vehicle/terrain_deformation.go:71-115`)
- [x] **Documentation** — Example code in `doc.go` references `log.Println()` instead of structured logging (`destruction/doc.go:64`)

### Low Severity
- [x] **API Naming** — `GetTerrainTypeFromTile()` is a package-level function that should be a method on a terrain manager or system (`vehicle/terrain_deformation.go:119-134`)
- [x] **Component Initialization** — `NewSuspensionComponent()`, `NewCollisionResponseComponent()`, `NewWeightTransferComponent()`, `NewTerrainDeformationComponent()` are constructor functions for components, which is acceptable but could be simplified by having the system create default components
- [x] **Code Organization** — `EnhancedVehicleSystem.Update()` accepts `[]Entity` but defines its own `Entity` interface locally instead of importing from `pkg/engine` (`vehicle/system.go:20-23`) — This is intentional to avoid import cycles but could be confusing

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | N/A | Physics package has no direct input handling |
| Mouse | N/A | Physics package has no direct input handling |
| Gamepad | N/A | Physics package has no direct input handling |
| Touch | N/A | Physics package has no direct input handling |
| VR | N/A | Physics package has no direct input handling |
| Stub/Test | ✅ | Tests use direct component/system instantiation without Ebiten dependencies |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| N/A | N/A | N/A | N/A | Physics package provides simulation systems with no UI components |

## Documentation Coverage
- Package `doc.go`: ✅ (comprehensive package-level documentation exists)
- Exported symbols documented: 95/100 (estimated 95%)
- Complex algorithms commented: ✅ (suspension spring-damper, weight transfer, collision response all explained)

## Integration Status
The physics package provides three specialized subsystems (vehicle, fluids, destruction) with well-defined component-system architecture.

- System registration: ❌ — **EnhancedVehicleSystem initialized but never registered in World** (see High Severity); destruction system registered via wrapper (`cmd/client/handlers.go:2031`); fluid managers initialized but not registered as systems (`cmd/client/init_versions.go`)
- Component registration: ✅ — Components implement `Type() string` and are discoverable via component type strings
- Serialize/Deserialize: ✅ — `BuoyancyComponent`, `SwimmingComponent`, `FloodingComponent` implement serialization (`fluids/types.go`)
- Network sync: ⚠️ — Vehicle and destruction components not currently included in network snapshot system (client-side visual physics only per `cmd/server/v8_systems.go:81-84`)
- Genre theming: N/A — Physics simulation is deterministic and genre-agnostic
- Mod compatibility: ✅ — Physics parameters (suspension stiffness, damage thresholds, fluid properties) exposed as component fields, modifiable via mod system

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅ | All systems work on desktop platforms |
| WASM | ✅ | No WASM-specific code; physics calculations are platform-agnostic |
| Mobile | ✅ | No mobile-specific concerns; math-only package |

## Recommendations
1. **[HIGH]** Register `EnhancedVehicleSystem` with World in `cmd/client/handlers.go` after initialization (add `game.World.AddSystem(sys.enhancedVehicleSys)` around line 2045 after other vehicle systems)
2. **[HIGH]** Refactor component getter/setter methods to systems or separate manager types to enforce ECS purity — Move `GetWheelLoad()`, `GetDamageMultiplier()`, etc. to system methods or create separate manager types (`SuspensionManager`, `CollisionManager`, etc.) that operate on pure data components
3. **[MED]** Fix documentation example in `destruction/doc.go:64` to use structured logging instead of `log.Println()`
4. **[MED]** Consider registering fluid managers (`BuoyancyCalculator`, `SwimmingManager`, `FloodingManager`) as formal systems in World if they need per-frame updates (currently they are utility classes called on-demand)
5. **[LOW]** Move `GetTerrainTypeFromTile()` to a terrain integration system or manager instead of package-level function
6. **[LOW]** Add server-side EnhancedVehicleSystem validation for multiplayer physics sync (noted as TODO in `cmd/server/v8_systems.go:81-84`)
7. **[LOW]** Document in `INTEGRATION.md` that vehicle physics is currently client-side visual only and server uses basic physics (this is by design but should be clearly documented)
