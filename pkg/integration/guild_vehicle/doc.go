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
// # Integration Status
//
// IMPLEMENTED:
//   - Thread-safe fleet and vehicle management
//   - Formation bonus calculations
//   - Siege engine type definitions with damage multipliers
//   - Maintenance cost calculations
//   - Gzip-compressed persistence (save/load)
//   - Access control for shared vehicle access
//   - pkg/network/federation/guild: Guild membership validation via MembershipValidator
//     interface; inject with SetMembershipValidator(guildManager).
//   - pkg/engine: VehicleComponent sync via VehicleSyncer interface; inject with
//     SetVehicleSyncer; engine receives GuildVehicleFleetComponent on AddVehicle/RemoveVehicle.
//   - pkg/engine/physics/vehicle: Formation target offsets via GetFormationOffsets;
//     feed per-slot FormationOffset values into vehicle physics target positions.
//   - pkg/world/territory: Siege damage via StructureDamager interface; inject with
//     SetStructureDamager(territoryManager); call ApplySiegeDamage on vehicle fire.
//
// # Thread Safety
//
// All operations are thread-safe using sync.RWMutex. Concurrent reads and writes
// to fleet data are safe.
//
// # Persistence
//
// Fleet data can be saved and loaded using gzip-compressed JSON:
//
//	err := manager.Save("fleets.json.gz")
//	err = manager.Load("fleets.json.gz")
package guild_vehicle
