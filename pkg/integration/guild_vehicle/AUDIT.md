# Audit: pkg/integration/guild_vehicle
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The guild_vehicle package provides fleet management for guild-owned vehicles with formation bonuses, siege engines, and access control. The implementation is largely complete with excellent test coverage (97.5%) and thread-safe operations. However, it violates the deterministic procgen requirement by using `time.Now()` for timestamp management, and the component lacks Serialize/Deserialize methods for persistence support.

## Issues Found
- [ ] **high** Deterministic Procgen — Uses `time.Now()` for timestamps which violates determinism requirement. While this package is an integration package (not procgen), timestamps should ideally use injected time providers for testability. (`fleet_manager.go:44`, `fleet_manager.go:45`, `fleet_manager.go:72`, `fleet_manager.go:73`, `fleet_manager.go:91`, `fleet_manager.go:92`, `fleet_manager.go:96`, `fleet_manager.go:117`, `fleet_manager.go:135`, `fleet_manager.go:156`, `fleet_manager.go:194`, `fleet_manager.go:211`)
- [ ] **med** Integration Points — GuildVehicleFleetComponent lacks Serialize/Deserialize methods for persistence support. Core engine components (PositionComponent, VelocityComponent, ColliderComponent) implement these methods for save/load functionality. (`types.go:209-225`)
- [ ] **low** Error Handling — No structured logging with logrus.WithFields for error paths or important operations. While errors are properly returned with context using fmt.Errorf, operational logging would improve observability. (All files)
- [ ] **low** Doc Coverage — FleetManager methods have no godoc comments. Only package-level documentation exists in doc.go. Public methods should document parameters, return values, and error conditions. (`fleet_manager.go:14-367`)

## Test Coverage
97.5% (target: 65%) ✅

Excellent coverage with comprehensive table-driven tests and benchmarks. Tests include:
- Unit tests for all types (FormationType, SiegeEngineType, FleetBonus, Fleet methods)
- FleetManager CRUD operations (create, add, remove, grant/revoke access)
- Formation and maintenance cost calculations
- Persistence (Save/Load with gzip compression)
- Edge cases (empty manager, invalid paths, concurrent access)
- Benchmarks for performance-critical operations

## Integration Status
The package is integrated with the engine layer via `pkg/engine/guild_vehicle_system.go` and `pkg/engine/guild_vehicle_system_test.go`. The GuildVehicleFleetComponent is used by the engine system to apply formation bonuses and manage guild vehicle entities.

**Integration Points:**
- ✅ Engine system exists: `pkg/engine/guild_vehicle_system.go`
- ✅ Engine tests exist: `pkg/engine/guild_vehicle_system_test.go`
- ⚠️ Component persistence: Missing Serialize/Deserialize methods
- ⚠️ System registration: Not found in `pkg/engine/system_init.go` (may need to verify registration)

**Planned Integrations (from doc.go):**
- ❌ pkg/network/federation/guild: Guild membership validation and permissions
- ❌ pkg/engine: VehicleComponent and VehicleCombatComponent synchronization
- ❌ pkg/engine/physics/vehicle: Formation-based physics behavior
- ❌ pkg/world/territory: Siege damage application to territory structures

## Recommendations
1. **High Priority**: Add time provider injection to FleetManager for testability and potential determinism (if fleet creation needs to be replicated across servers). Pattern: `type TimeProvider interface { Now() time.Time }` with default implementation and test mocks.
2. **Medium Priority**: Implement Serialize/Deserialize methods on GuildVehicleFleetComponent for persistence support, following the pattern in `pkg/engine/components.go`.
3. **Low Priority**: Add godoc comments to all exported FleetManager methods with parameter descriptions, return value semantics, and error conditions.
4. **Low Priority**: Add structured logging with logrus.WithFields for key operations (fleet creation, vehicle addition/removal, formation changes) with fields like `guildID`, `fleetID`, `vehicleID`, `formation`.
