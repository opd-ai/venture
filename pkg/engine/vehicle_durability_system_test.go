// Package engine provides tests for VehicleDurabilitySystem.
// Phase 21.2: Terrain Hazard Integration - Gap A6 Resolution
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestVehicleDurabilitySystem_NewVehicleDurabilitySystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	if system == nil {
		t.Fatal("NewVehicleDurabilitySystem returned nil")
	}
	if system.world != world {
		t.Error("world reference not set correctly")
	}
}

func TestVehicleDurabilitySystem_NewVehicleDurabilitySystem_NilWorld(t *testing.T) {
	system := NewVehicleDurabilitySystem(nil)
	if system == nil {
		t.Fatal("NewVehicleDurabilitySystem returned nil for nil world")
	}
}

func TestVehicleDurabilitySystem_SetTerrainCollisionChecker(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)
	checker := NewTerrainCollisionChecker(32, 32)

	system.SetTerrainCollisionChecker(checker)

	if system.terrainCollisionChecker != checker {
		t.Error("SetTerrainCollisionChecker did not set terrain collision checker correctly")
	}
}

func TestVehicleDurabilitySystem_SetTerrainCollisionChecker_Nil(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Should not panic when setting nil
	system.SetTerrainCollisionChecker(nil)

	if system.terrainCollisionChecker != nil {
		t.Error("Expected terrainCollisionChecker to be nil")
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_LavaFlow(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with lava
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileLavaFlow)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle entity on lava tile
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16} // Center of tile (5,5)
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleMount)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify damage was applied
	expectedDamage := LavaFlowDamagePerSecond
	expectedDurability := initialDurability - expectedDamage
	if vehicleComp.Durability != expectedDurability {
		t.Errorf("Durability = %f, want %f (expected %f damage from lava)",
			vehicleComp.Durability, expectedDurability, expectedDamage)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_Pit(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with pit
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TilePit)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle entity on pit tile
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleCart)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify damage was applied
	expectedDamage := PitDamagePerSecond
	expectedDurability := initialDurability - expectedDamage
	if vehicleComp.Durability != expectedDurability {
		t.Errorf("Durability = %f, want %f (expected %f damage from pit)",
			vehicleComp.Durability, expectedDurability, expectedDamage)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_Pit_GliderImmune(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with pit
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TilePit)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create glider entity on pit tile
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleGlider)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify NO damage was applied (gliders fly over pits)
	if vehicleComp.Durability != initialDurability {
		t.Errorf("Glider took pit damage: Durability = %f, want %f (gliders should be immune)",
			vehicleComp.Durability, initialDurability)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_DeepWater(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with deep water
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileWaterDeep)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create cart entity on deep water tile (carts can't traverse water)
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleCart)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify damage was applied
	expectedDamage := DeepWaterDamagePerSecond
	expectedDurability := initialDurability - expectedDamage
	if vehicleComp.Durability != expectedDurability {
		t.Errorf("Durability = %f, want %f (expected %f damage from deep water)",
			vehicleComp.Durability, expectedDurability, expectedDamage)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_DeepWater_BoatImmune(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with deep water
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileWaterDeep)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create boat entity on deep water tile
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleBoat)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify NO damage was applied (boats can traverse water)
	if vehicleComp.Durability != initialDurability {
		t.Errorf("Boat took deep water damage: Durability = %f, want %f (boats should be immune)",
			vehicleComp.Durability, initialDurability)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_NoTerrainChecker(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// No terrain checker set

	// Create vehicle entity
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 160, Y: 160}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleMount)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage - should not panic
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify no damage was applied (graceful degradation)
	if vehicleComp.Durability != initialDurability {
		t.Errorf("Vehicle took damage without terrain checker: Durability = %f, want %f",
			vehicleComp.Durability, initialDurability)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileLavaFlow)

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle entity WITHOUT position
	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage - should not panic
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify no damage was applied
	if vehicleComp.Durability != initialDurability {
		t.Errorf("Vehicle without position took damage: Durability = %f, want %f",
			vehicleComp.Durability, initialDurability)
	}
}

func TestVehicleDurabilitySystem_CheckEnvironmentalDamage_SafeTerrain(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with floor (safe)
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileFloor)

	// Set up terrain checker
	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle entity on safe tile
	vehicle := world.CreateEntity()
	pos := &PositionComponent{X: 5*32 + 16, Y: 5*32 + 16}
	vehicle.AddComponent(pos)

	vehicleComp := NewVehicleComponent(VehicleMount)
	initialDurability := vehicleComp.Durability
	vehicle.AddComponent(vehicleComp)

	// Apply environmental damage for 1 second
	system.checkEnvironmentalDamage(vehicle, vehicleComp, 1.0)

	// Verify NO damage was applied on safe terrain
	if vehicleComp.Durability != initialDurability {
		t.Errorf("Vehicle took damage on safe terrain: Durability = %f, want %f",
			vehicleComp.Durability, initialDurability)
	}
}

func TestVehicleDurabilitySystem_CalculateTerrainHazardDamage(t *testing.T) {
	system := NewVehicleDurabilitySystem(nil)

	tests := []struct {
		name        string
		tile        terrain.TileType
		vehicleType VehicleType
		deltaTime   float64
		wantDamage  float64
	}{
		{
			name:        "lava_flow_mount",
			tile:        terrain.TileLavaFlow,
			vehicleType: VehicleMount,
			deltaTime:   1.0,
			wantDamage:  LavaFlowDamagePerSecond,
		},
		{
			name:        "lava_flow_boat",
			tile:        terrain.TileLavaFlow,
			vehicleType: VehicleBoat,
			deltaTime:   1.0,
			wantDamage:  LavaFlowDamagePerSecond, // Lava damages all vehicles
		},
		{
			name:        "lava_flow_half_second",
			tile:        terrain.TileLavaFlow,
			vehicleType: VehicleMount,
			deltaTime:   0.5,
			wantDamage:  LavaFlowDamagePerSecond * 0.5,
		},
		{
			name:        "pit_cart",
			tile:        terrain.TilePit,
			vehicleType: VehicleCart,
			deltaTime:   1.0,
			wantDamage:  PitDamagePerSecond,
		},
		{
			name:        "pit_glider_immune",
			tile:        terrain.TilePit,
			vehicleType: VehicleGlider,
			deltaTime:   1.0,
			wantDamage:  0, // Gliders fly over pits
		},
		{
			name:        "deep_water_cart",
			tile:        terrain.TileWaterDeep,
			vehicleType: VehicleCart,
			deltaTime:   1.0,
			wantDamage:  DeepWaterDamagePerSecond,
		},
		{
			name:        "deep_water_boat_immune",
			tile:        terrain.TileWaterDeep,
			vehicleType: VehicleBoat,
			deltaTime:   1.0,
			wantDamage:  0, // Boats can traverse water
		},
		{
			name:        "deep_water_glider_immune",
			tile:        terrain.TileWaterDeep,
			vehicleType: VehicleGlider,
			deltaTime:   1.0,
			wantDamage:  0, // Gliders fly over water
		},
		{
			name:        "floor_safe",
			tile:        terrain.TileFloor,
			vehicleType: VehicleMount,
			deltaTime:   1.0,
			wantDamage:  0,
		},
		{
			name:        "corridor_safe",
			tile:        terrain.TileCorridor,
			vehicleType: VehicleMount,
			deltaTime:   1.0,
			wantDamage:  0,
		},
		{
			name:        "wall_safe",
			tile:        terrain.TileWall,
			vehicleType: VehicleMount,
			deltaTime:   1.0,
			wantDamage:  0, // Walls don't damage; they block
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle := NewVehicleComponent(tt.vehicleType)
			damage := system.calculateTerrainHazardDamage(tt.tile, vehicle, tt.deltaTime)
			if damage != tt.wantDamage {
				t.Errorf("calculateTerrainHazardDamage(%s, %s, %f) = %f, want %f",
					tt.tile.String(), tt.vehicleType.String(), tt.deltaTime, damage, tt.wantDamage)
			}
		})
	}
}

func TestVehicleDurabilitySystem_Update_WithTerrainHazards(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with lava
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileLavaFlow)
	testTerrain.SetTile(6, 6, terrain.TileFloor)

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle on lava
	vehicleOnLava := world.CreateEntity()
	vehicleOnLava.AddComponent(&PositionComponent{X: 5*32 + 16, Y: 5*32 + 16})
	vehicleCompLava := NewVehicleComponent(VehicleMount)
	initialLavaDurability := vehicleCompLava.Durability
	vehicleOnLava.AddComponent(vehicleCompLava)

	// Create vehicle on safe tile
	vehicleSafe := world.CreateEntity()
	vehicleSafe.AddComponent(&PositionComponent{X: 6*32 + 16, Y: 6*32 + 16})
	vehicleCompSafe := NewVehicleComponent(VehicleMount)
	initialSafeDurability := vehicleCompSafe.Durability
	vehicleSafe.AddComponent(vehicleCompSafe)

	entities := []*Entity{vehicleOnLava, vehicleSafe}

	// Run update for 1 second
	system.Update(entities, 1.0)

	// Vehicle on lava should have taken damage
	expectedLavaDurability := initialLavaDurability - LavaFlowDamagePerSecond
	if vehicleCompLava.Durability != expectedLavaDurability {
		t.Errorf("Vehicle on lava: Durability = %f, want %f",
			vehicleCompLava.Durability, expectedLavaDurability)
	}

	// Vehicle on safe terrain should have no damage
	if vehicleCompSafe.Durability != initialSafeDurability {
		t.Errorf("Vehicle on safe terrain: Durability = %f, want %f",
			vehicleCompSafe.Durability, initialSafeDurability)
	}
}

func TestVehicleDurabilitySystem_VehicleDestruction_FromTerrainHazard(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create terrain with lava
	testTerrain := terrain.NewTerrain(10, 10, 12345)
	testTerrain.SetTile(5, 5, terrain.TileLavaFlow)

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(testTerrain)
	system.SetTerrainCollisionChecker(checker)

	// Create vehicle with low durability
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 5*32 + 16, Y: 5*32 + 16})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Durability = 5.0 // Low durability - will be destroyed
	vehicleComp.Speed = 100.0
	vehicleComp.CurrentPassengers = 1
	vehicle.AddComponent(vehicleComp)

	// Run update for 1 second (10 damage from lava)
	system.Update([]*Entity{vehicle}, 1.0)

	// Vehicle should be destroyed
	if !vehicleComp.IsDestroyed() {
		t.Error("Vehicle should be destroyed after lava damage")
	}

	// Speed should be 0 (vehicle stopped)
	if vehicleComp.Speed != 0.0 {
		t.Errorf("Destroyed vehicle speed = %f, want 0.0", vehicleComp.Speed)
	}

	// Passengers should be 0 (dismounted)
	if vehicleComp.CurrentPassengers != 0 {
		t.Errorf("Destroyed vehicle passengers = %d, want 0", vehicleComp.CurrentPassengers)
	}
}

func TestVehicleDurabilitySystem_ApplyDamage_Extended(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)
	initialDurability := vehicleComp.Durability

	// Apply damage
	destroyed := system.ApplyDamage(vehicle, 25.0)

	if destroyed {
		t.Error("Vehicle should not be destroyed with 25 damage")
	}

	expectedDurability := initialDurability - 25.0
	if vehicleComp.Durability != expectedDurability {
		t.Errorf("Durability = %f, want %f", vehicleComp.Durability, expectedDurability)
	}
}

func TestVehicleDurabilitySystem_ApplyDamage_NoVehicle(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	// Create entity without vehicle component
	entity := world.CreateEntity()

	destroyed := system.ApplyDamage(entity, 25.0)

	if destroyed {
		t.Error("Should return false for entity without vehicle component")
	}
}

func TestVehicleDurabilitySystem_RepairVehicle_Extended(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Durability = 50.0 // Half health
	vehicle.AddComponent(vehicleComp)

	// Repair
	repaired := system.RepairVehicle(vehicle, 30.0)

	if !repaired {
		t.Error("RepairVehicle should return true")
	}

	if vehicleComp.Durability != 80.0 {
		t.Errorf("Durability = %f, want 80.0", vehicleComp.Durability)
	}
}

func TestVehicleDurabilitySystem_RepairVehicle_NoVehicle(t *testing.T) {
	world := NewWorld()
	system := NewVehicleDurabilitySystem(world)

	entity := world.CreateEntity()

	repaired := system.RepairVehicle(entity, 30.0)

	if repaired {
		t.Error("Should return false for entity without vehicle component")
	}
}

func TestVehicleDurabilitySystem_TerrainHazardDamageConstants(t *testing.T) {
	// Verify damage constants match expected values from PLAN.md
	// PLAN.md specifies: lava 10/sec, spikes 5/sec, acid 15/sec
	// We use: LavaFlow 10/sec, DeepWater 5/sec, Pit 15/sec

	if LavaFlowDamagePerSecond != 10.0 {
		t.Errorf("LavaFlowDamagePerSecond = %f, want 10.0", LavaFlowDamagePerSecond)
	}

	if DeepWaterDamagePerSecond != 5.0 {
		t.Errorf("DeepWaterDamagePerSecond = %f, want 5.0", DeepWaterDamagePerSecond)
	}

	if PitDamagePerSecond != 15.0 {
		t.Errorf("PitDamagePerSecond = %f, want 15.0", PitDamagePerSecond)
	}
}
