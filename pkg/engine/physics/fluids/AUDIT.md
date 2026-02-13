# Audit: github.com/opd-ai/venture/pkg/engine/physics/fluids
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The fluids package provides comprehensive fluid dynamics simulation with buoyancy, swimming, and flooding mechanics. Implementation is technically sound with excellent test coverage (95.3%) and proper ECS component design. However, the package lacks integration with the main engine (no System implementation), has no serialization support for persistent components, and missing structured logging for debugging production issues.

## Issues Found
- [ ] **high** Integration — No System implementation that uses BuoyancyComponent, SwimmingComponent, or FloodingComponent in `/pkg/engine/` (components defined but unused by ECS)
- [ ] **high** Serialization — BuoyancyComponent, SwimmingComponent, and FloodingComponent lack Serialize/Deserialize methods for save/load persistence (`types.go:68-112`)
- [ ] **med** Logging — No structured logging with logrus.WithFields for error paths or state changes (entire package)
- [ ] **med** Error handling — AddFluid and RemoveFluid errors not logged when bounds violations occur (`simulator.go:276-310`)
- [ ] **low** Documentation — Exported function UpdateDensity lacks godoc comment (`buoyancy_calculator.go:54`)
- [ ] **low** Documentation — GetNetForce exported method lacks godoc explaining net force calculation (`buoyancy_calculator.go:48`)

## Test Coverage
95.3% (target: 65%) ✅

Excellent coverage with comprehensive table-driven tests for:
- Buoyancy calculations (various fluid types, partial/full submersion)
- Swimming mechanics (stamina drain, drowning, speed multipliers)
- Flooding system (sources, level tracking, percentage calculations)
- Simulator (fluid flow, pressure, viscosity, advection, boundary conditions)
- Type conversions and properties

## Integration Status
**Current State**: Standalone library with no engine integration

**Components Defined**:
- `BuoyancyComponent` - Buoyancy properties for entities (mass, volume, density, buoyant force)
- `SwimmingComponent` - Swimming mechanics (stamina, drowning, speed)
- `FloodingComponent` - Flooding state for enclosed areas

**Managers/Calculators** (not Systems):
- `BuoyancyCalculator` - Calculates buoyancy forces
- `SwimmingManager` - Manages swimming state and stamina
- `FloodingManager` - Manages flood sources and levels
- `Simulator` - Grid-based fluid dynamics simulation

**Missing Integration**:
1. No System in `/pkg/engine/` that updates entities with these components
2. Components not referenced in `pkg/engine/system_init.go`
3. LayerComponent references swimming capability but doesn't integrate with SwimmingComponent
4. Movement systems (movement_system.go, terrain_movement_speed_system.go) don't use fluid mechanics
5. No terrain-fluid interaction (terrain cells don't store fluid type/amount)

**Expected Systems** (not implemented):
- `FluidPhysicsSystem` - Updates buoyancy calculations for entities in fluid cells
- `SwimmingSystem` - Manages swimming state, stamina drain, drowning damage
- `FloodingSystem` - Updates flooding areas and integrates with terrain/simulator
- Terrain integration - Store fluid data per terrain cell, sync with Simulator grid

## Recommendations
1. **HIGH PRIORITY**: Implement FluidPhysicsSystem, SwimmingSystem, FloodingSystem in `/pkg/engine/` to integrate components with ECS architecture
2. **HIGH PRIORITY**: Add Serialize/Deserialize methods to all three components for save/load support (follow pattern in `pkg/engine/components.go`)
3. **MEDIUM PRIORITY**: Add structured logging with logrus.WithFields to Simulator.AddFluid, RemoveFluid, Update for debugging fluid behavior (field names: `x`, `y`, `fluid_type`, `amount`, `error`)
4. **MEDIUM PRIORITY**: Integrate fluid data with terrain system - add FluidAmount and FluidType to terrain cells, sync with Simulator grid
5. **LOW PRIORITY**: Add godoc comments to exported functions UpdateDensity and GetNetForce
6. **LOW PRIORITY**: Register fluid systems in `pkg/engine/system_init.go` once implemented
