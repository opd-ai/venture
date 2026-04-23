package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
)

func TestNewGuildVehicleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.manager == nil {
		t.Error("Fleet manager not initialized")
	}
	if sys.logger == nil {
		t.Error("Logger not initialized")
	}
}

func TestGuildVehicleSystem_Update_NoComponents(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	entity := world.CreateEntity()
	entities := []*Entity{entity}

	// Should not panic with no guild_vehicle_fleet component
	sys.Update(entities, 0.016)
}

func TestGuildVehicleSystem_Update_WithComponents(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	// Create fleet
	guildID := "test_guild"
	fleetID := "test_fleet"
	err := sys.manager.CreateFleet(guildID, fleetID, "commander1")
	if err != nil {
		t.Fatalf("Failed to create fleet: %v", err)
	}

	// Create vehicle entity with components
	vehicleEntity := world.CreateEntity()
	vehicleEntity.AddComponent(&guild_vehicle.GuildVehicleFleetComponent{
		GuildID:           guildID,
		FleetID:           fleetID,
		SiegeType:         guild_vehicle.SiegeNone,
		FormationPosition: 0,
	})
	vehicleEntity.AddComponent(&VehicleCombatComponent{
		RammingDamage: 100.0,
		WeaponMounted: true,
		WeaponDamage:  50.0,
		ArmorRating:   10.0,
	})

	// Add vehicle to fleet
	err = sys.manager.AddVehicle(guildID, vehicleEntity.ID, fleetID)
	if err != nil {
		t.Fatalf("Failed to add vehicle to fleet: %v", err)
	}

	// Set formation
	err = sys.manager.SetFormation(guildID, fleetID, guild_vehicle.FormationWedge)
	if err != nil {
		t.Fatalf("Failed to set formation: %v", err)
	}

	entities := []*Entity{vehicleEntity}

	// Update system
	sys.Update(entities, 0.016)

	// Check that damage was modified by formation bonus
	comp, ok := vehicleEntity.GetComponent("vehicle_combat")
	if !ok {
		t.Fatal("Expected vehicle_combat component")
	}
	combatComp := comp.(*VehicleCombatComponent)

	// Wedge formation: 1.07x damage
	expectedRammingDamage := 100.0 * 1.07
	if combatComp.RammingDamage != expectedRammingDamage {
		t.Errorf("Expected ramming damage %.2f, got %.2f", expectedRammingDamage, combatComp.RammingDamage)
	}
	expectedWeaponDamage := 50.0 * 1.07
	if combatComp.WeaponDamage != expectedWeaponDamage {
		t.Errorf("Expected weapon damage %.2f, got %.2f", expectedWeaponDamage, combatComp.WeaponDamage)
	}
}

func TestGuildVehicleSystem_SiegeEngineMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	tests := []struct {
		name               string
		siegeType          guild_vehicle.SiegeEngineType
		baseDamage         float64
		formation          guild_vehicle.FormationType
		expectedMultiplier float64
	}{
		{
			name:               "battering_ram_no_formation",
			siegeType:          guild_vehicle.SiegeBatteringRam,
			baseDamage:         100.0,
			formation:          guild_vehicle.FormationNone,
			expectedMultiplier: 3.0, // 1.0 * 3.0
		},
		{
			name:               "catapult_wedge_formation",
			siegeType:          guild_vehicle.SiegeCatapult,
			baseDamage:         100.0,
			formation:          guild_vehicle.FormationWedge,
			expectedMultiplier: 5.35, // 1.07 * 5.0
		},
		{
			name:               "siege_tower_line_formation",
			siegeType:          guild_vehicle.SiegeTower,
			baseDamage:         100.0,
			formation:          guild_vehicle.FormationLine,
			expectedMultiplier: 2.1, // 1.05 * 2.0
		},
		{
			name:               "ballista_no_formation",
			siegeType:          guild_vehicle.SiegeBallistaBattery,
			baseDamage:         100.0,
			formation:          guild_vehicle.FormationNone,
			expectedMultiplier: 4.0, // 1.0 * 4.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildID := "guild_" + tt.name
			fleetID := "fleet_" + tt.name

			// Create fleet
			err := sys.manager.CreateFleet(guildID, fleetID, "commander1")
			if err != nil {
				t.Fatalf("Failed to create fleet: %v", err)
			}

			// Set formation
			err = sys.manager.SetFormation(guildID, fleetID, tt.formation)
			if err != nil {
				t.Fatalf("Failed to set formation: %v", err)
			}

			// Create vehicle entity
			vehicleEntity := world.CreateEntity()
			vehicleEntity.AddComponent(&VehicleCombatComponent{
				RammingDamage: tt.baseDamage,
				WeaponMounted: false,
			})

			// Add vehicle to fleet with the correct siege type
			err = sys.manager.AddVehicleWithType(guildID, vehicleEntity.ID, fleetID, tt.siegeType, 100)
			if err != nil {
				t.Fatalf("Failed to add vehicle to fleet: %v", err)
			}

			entities := []*Entity{vehicleEntity}

			// Update system
			sys.Update(entities, 0.016)

			// Check damage multiplier
			comp, ok := vehicleEntity.GetComponent("vehicle_combat")
			if !ok {
				t.Fatal("Expected vehicle_combat component")
			}
			combatComp := comp.(*VehicleCombatComponent)

			expectedDamage := tt.baseDamage * tt.expectedMultiplier
			if combatComp.RammingDamage != expectedDamage {
				t.Errorf("Expected ramming damage %.2f, got %.2f", expectedDamage, combatComp.RammingDamage)
			}
		})
	}
}

func TestGuildVehicleSystem_AddVehicleToFleet(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	guildID := "test_guild"
	fleetID := "test_fleet"
	vehicleID := uint64(12345)

	err := sys.AddVehicleToFleet(guildID, vehicleID, fleetID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify vehicle was added
	fleetIDResult, err := sys.manager.GetVehicleFleetID(guildID, vehicleID)
	if err != nil {
		t.Errorf("Expected vehicle in fleet, got error: %v", err)
	}
	if fleetIDResult != fleetID {
		t.Errorf("Expected fleet ID %s, got %s", fleetID, fleetIDResult)
	}
}

func TestGuildVehicleSystem_SetFormation(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	guildID := "test_guild"
	fleetID := "test_fleet"

	// Create fleet first
	err := sys.manager.CreateFleet(guildID, fleetID, "commander1")
	if err != nil {
		t.Fatalf("Failed to create fleet: %v", err)
	}

	formations := []guild_vehicle.FormationType{
		guild_vehicle.FormationLine,
		guild_vehicle.FormationWedge,
		guild_vehicle.FormationColumn,
		guild_vehicle.FormationCircle,
	}

	for _, formation := range formations {
		err := sys.SetFormation(guildID, fleetID, formation)
		if err != nil {
			t.Errorf("Expected no error for formation %v, got %v", formation, err)
		}

		// Verify formation was set
		bonuses := sys.manager.GetFleetBonuses(guildID, fleetID)
		if bonuses.Formation != formation {
			t.Errorf("Expected formation %v, got %v", formation, bonuses.Formation)
		}
	}
}

func TestGuildVehicleSystem_GrantAccess(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	guildID := "test_guild"
	vehicleID := uint64(12345)
	playerID := "player_123"
	fleetID := "test_fleet"

	// Add vehicle to fleet
	err := sys.AddVehicleToFleet(guildID, vehicleID, fleetID)
	if err != nil {
		t.Fatalf("Failed to add vehicle: %v", err)
	}

	// Grant access
	err = sys.GrantAccess(guildID, vehicleID, playerID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify access
	hasAccess := sys.CheckAccess(guildID, vehicleID, playerID)
	if !hasAccess {
		t.Error("Expected player to have access")
	}
}

func TestGuildVehicleSystem_CheckAccess(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	guildID := "test_guild"
	vehicleID := uint64(12345)
	playerID := "player_123"
	fleetID := "test_fleet"

	// Add vehicle to fleet
	err := sys.AddVehicleToFleet(guildID, vehicleID, fleetID)
	if err != nil {
		t.Fatalf("Failed to add vehicle: %v", err)
	}

	// Check access before granting (should be false)
	hasAccess := sys.CheckAccess(guildID, vehicleID, playerID)
	if hasAccess {
		t.Error("Expected player to not have access before granting")
	}

	// Grant access
	err = sys.GrantAccess(guildID, vehicleID, playerID)
	if err != nil {
		t.Fatalf("Failed to grant access: %v", err)
	}

	// Check access after granting (should be true)
	hasAccess = sys.CheckAccess(guildID, vehicleID, playerID)
	if !hasAccess {
		t.Error("Expected player to have access after granting")
	}
}

func TestGuildVehicleSystem_GetFleetManager(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	manager := sys.GetFleetManager()
	if manager == nil {
		t.Error("Expected non-nil fleet manager")
	}
	if manager != sys.manager {
		t.Error("Expected same fleet manager instance")
	}
}

func TestGuildVehicleSystem_InvalidComponentType(t *testing.T) {
	world := NewWorld()
	sys := NewGuildVehicleSystem(world)

	// Create entity with wrong component type
	entity := world.CreateEntity()
	entity.Components["guild_vehicle_fleet"] = &PositionComponent{} // Wrong type

	entities := []*Entity{entity}

	// Should not panic, just log warning
	sys.Update(entities, 0.016)
}
