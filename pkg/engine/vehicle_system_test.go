package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestVehicleSystem(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	// Create a mount vehicle
	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create a rider
	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 100, Y: 100})
	rider.AddComponent(&VelocityComponent{})
	rider.AddComponent(&InputComponent{MoveX: 1, MoveY: 0})

	// Test mounting
	err := system.Mount(rider, vehicle)
	if err != nil {
		t.Fatalf("Mount() failed: %v", err)
	}

	if !system.IsMounted(rider) {
		t.Error("IsMounted() = false, want true after mounting")
	}

	mountedVehicle := system.GetMountedVehicle(rider)
	if mountedVehicle == nil || mountedVehicle.ID != vehicle.ID {
		t.Error("GetMountedVehicle() did not return correct vehicle")
	}

	// Test vehicle update with rider
	system.Update(0.016) // ~60 FPS frame

	// Verify vehicle has started moving
	if vehicleComp.Speed == 0 {
		t.Error("Vehicle speed should be > 0 after update with rider input")
	}

	// Test dismounting
	err = system.Dismount(rider, vehicle)
	if err != nil {
		t.Fatalf("Dismount() failed: %v", err)
	}

	if system.IsMounted(rider) {
		t.Error("IsMounted() = true, want false after dismounting")
	}

	// Verify vehicle stopped
	if vehicleComp.Speed != 0 {
		t.Error("Vehicle speed should be 0 after dismounting")
	}
}

func TestVehicleSystemCapacity(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	// Create a vehicle with capacity of 1
	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	// Create two riders
	rider1 := world.CreateEntity()
	rider2 := world.CreateEntity()

	// Mount first rider - should succeed
	err := system.Mount(rider1, vehicle)
	if err != nil {
		t.Fatalf("Mount() first rider failed: %v", err)
	}

	// Try to mount second rider - should fail due to capacity
	err = system.Mount(rider2, vehicle)
	if err == nil {
		t.Error("Mount() second rider should fail due to capacity, but succeeded")
	}

	if vehicleComp.CurrentPassengers != 1 {
		t.Errorf("CurrentPassengers = %d, want 1", vehicleComp.CurrentPassengers)
	}
}

func TestVehicleSystemFuelConsumption(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Speed = 100.0 // Set initial speed
	vehicle.AddComponent(vehicleComp)

	initialFuel := vehicleComp.FuelAmount

	// Run updates to consume fuel
	for i := 0; i < 100; i++ {
		system.Update(0.016)
	}

	// Verify fuel was consumed
	if vehicleComp.FuelAmount >= initialFuel {
		t.Errorf("FuelAmount should decrease, got %f, initial %f", vehicleComp.FuelAmount, initialFuel)
	}
}

func TestVehicleSystemOutOfFuel(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Speed = 100.0
	vehicleComp.FuelAmount = 1.0 // Very low fuel
	vehicle.AddComponent(vehicleComp)

	// Run updates until fuel depleted
	for i := 0; i < 200; i++ {
		system.Update(0.1)
	}

	// Verify vehicle stopped when out of fuel
	if vehicleComp.FuelAmount != 0 {
		t.Errorf("FuelAmount should be 0, got %f", vehicleComp.FuelAmount)
	}

	if vehicleComp.Speed != 0 {
		t.Errorf("Speed should be 0 when out of fuel, got %f", vehicleComp.Speed)
	}
}

func TestVehicleSystemSpeedDecay(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Speed = 100.0
	vehicleComp.FuelAmount = 1000.0 // Plenty of fuel
	vehicle.AddComponent(vehicleComp)

	initialSpeed := vehicleComp.Speed

	// Run updates without rider input
	for i := 0; i < 50; i++ {
		system.Update(0.1)
	}

	// Verify speed decayed
	if vehicleComp.Speed >= initialSpeed {
		t.Errorf("Speed should decay over time, got %f, initial %f", vehicleComp.Speed, initialSpeed)
	}
}

func TestVehicleSystemCanTraverse(t *testing.T) {
	system := NewVehicleSystem(NewWorld())

	tests := []struct {
		name         string
		vehicleType  VehicleType
		terrainType  terrain.TileType
		wantTraverse bool
	}{
		{
			name:         "mount can traverse floor",
			vehicleType:  VehicleMount,
			terrainType:  terrain.TileFloor,
			wantTraverse: true,
		},
		{
			name:         "mount cannot traverse deep water",
			vehicleType:  VehicleMount,
			terrainType:  terrain.TileWaterDeep,
			wantTraverse: false,
		},
		{
			name:         "boat can traverse deep water",
			vehicleType:  VehicleBoat,
			terrainType:  terrain.TileWaterDeep,
			wantTraverse: true,
		},
		{
			name:         "boat cannot traverse floor",
			vehicleType:  VehicleBoat,
			terrainType:  terrain.TileFloor,
			wantTraverse: false,
		},
		{
			name:         "cart can traverse corridor",
			vehicleType:  VehicleCart,
			terrainType:  terrain.TileCorridor,
			wantTraverse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleComp := NewVehicleComponent(tt.vehicleType)
			got := system.CanTraverse(vehicleComp, tt.terrainType)
			if got != tt.wantTraverse {
				t.Errorf("CanTraverse() = %v, want %v", got, tt.wantTraverse)
			}
		})
	}
}

func TestVehicleSystemRiderVelocitySync(t *testing.T) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)
	vehicle.AddComponent(&PositionComponent{X: 0, Y: 0})

	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 0, Y: 0})
	velocityComp := &VelocityComponent{}
	rider.AddComponent(velocityComp)
	rider.AddComponent(&InputComponent{MoveX: 1, MoveY: 0})

	system.Mount(rider, vehicle)

	// Update several times to accelerate
	for i := 0; i < 10; i++ {
		system.Update(0.016)
	}

	// Verify rider velocity reflects vehicle speed
	if velocityComp.Vx == 0 && velocityComp.Vy == 0 {
		t.Error("Rider velocity should be non-zero when mounted on moving vehicle")
	}
}

func BenchmarkVehicleSystemUpdate(b *testing.B) {
	world := NewWorld()
	system := NewVehicleSystem(world)

	// Create 100 vehicles with riders
	for i := 0; i < 100; i++ {
		vehicle := world.CreateEntity()
		vehicleComp := NewVehicleComponent(VehicleMount)
		vehicle.AddComponent(vehicleComp)
		vehicle.AddComponent(&PositionComponent{X: float64(i * 10), Y: 0})

		rider := world.CreateEntity()
		rider.AddComponent(&PositionComponent{X: float64(i * 10), Y: 0})
		rider.AddComponent(&VelocityComponent{})
		rider.AddComponent(&InputComponent{MoveX: 1, MoveY: 0})

		system.Mount(rider, vehicle)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}
