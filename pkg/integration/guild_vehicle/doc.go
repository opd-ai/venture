// Package guild_vehicle provides guild vehicle fleet combat integration.
// This package connects V8 Guilds, V4 Vehicles, and V8 Vehicle Physics into
// coordinated fleet warfare mechanics for guild combat.
//
// Phase 56.1: Vehicle Fleet Combat (ROADMAP_V9.md)
//
// # Features
//
// - Guild vehicle ownership with shared access permissions
// - Fleet formations with coordinated movement bonuses
// - Siege engines for territory attacks (battering rams, catapults)
// - Automated maintenance costs from guild treasury
// - Fleet command system for multi-vehicle coordination
//
// # Fleet Formations
//
// Formations provide combat bonuses when vehicles move in coordinated patterns:
//
//	Line: 5% damage bonus, tight lateral spacing
//	Wedge: 7% damage bonus, frontal assault formation
//	Column: 10% defense bonus, single-file movement
//	Circle: 8% defense bonus, defensive perimeter
//	Scattered: No bonus, maximum maneuverability
//
// # Siege Engines
//
// Special vehicle types designed for territory attacks:
//
//	BatteringRam: 3x damage vs walls, slow movement
//	Catapult: 5x damage vs structures, long range
//	SiegeTower: 2x damage, enables wall climbing
//	BallistaBattery: 4x damage vs defensive towers
//
// # Usage Example
//
//	// Create fleet manager
//	manager := guild_vehicle.NewFleetManager()
//
//	// Add guild vehicle to fleet
//	vehicleID := uint64(12345)
//	guildID := "guild_warriors"
//	err := manager.AddVehicle(guildID, vehicleID, "primary_siege")
//
//	// Grant member access
//	err = manager.GrantAccess(guildID, vehicleID, "player_123")
//
//	// Set formation for combat bonuses
//	formation := guild_vehicle.FormationWedge
//	err = manager.SetFormation(guildID, "primary_siege", formation)
//
//	// Calculate formation bonuses
//	bonuses := manager.GetFleetBonuses(guildID, "primary_siege")
//	// bonuses.DamageMultiplier = 1.07 (7% bonus for Wedge)
//
//	// Process maintenance costs (daily)
//	cost := manager.CalculateMaintenanceCost(guildID, "primary_siege")
//	// Deduct from guild treasury
//
// # Integration
//
// This package integrates with:
//   - pkg/network/federation/guild: Guild management and permissions
//   - pkg/engine: VehicleComponent and VehicleCombatComponent
//   - pkg/engine/physics/vehicle: Enhanced vehicle physics
//   - pkg/world/territory: Territory control for siege mechanics
//
// All operations are thread-safe and support deterministic seed-based generation.
package guild_vehicle
