// Package engine provides vehicle system tests.
// Tests verify vehicle movement, durability, and mounting functionality.
//
// Phase 21.1: Vehicle Foundation
package engine

import (
	"math"
	"testing"
)

// TestVehicleMovementSystem tests vehicle physics and movement
func TestVehicleMovementSystem(t *testing.T) {
	world := NewWorld()
	vms := NewVehicleMovementSystem(world)

	// Create vehicle entity with position, rotation, and vehicle component
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicle.AddComponent(&RotationComponent{Angle: 0})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	// Initial state
	if vehicleComp.Speed != 0 {
		t.Errorf("Initial speed should be 0, got %f", vehicleComp.Speed)
	}

	// Update without control - should not move
	entities := []*Entity{vehicle}
	vms.Update(entities, 0.1)

	posComp, _ := vehicle.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X != 100 || pos.Y != 100 {
		t.Errorf("Vehicle moved without control: (%f, %f)", pos.X, pos.Y)
	}
}

func TestVehicleMovementSystem_WithStubInput(t *testing.T) {
	world := NewWorld()
	vms := NewVehicleMovementSystem(world)

	// Create vehicle with input
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicle.AddComponent(&RotationComponent{Angle: 0})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	// Add stub input provider
	input := NewStubInput()
	input.SetMovement(1.0, 0.0) // Move right
	vehicle.AddComponent(input)

	entities := []*Entity{vehicle}

	// Update multiple times to build up speed
	for i := 0; i < 10; i++ {
		vms.Update(entities, 0.1)
	}

	// Should have accelerated
	if vehicleComp.Speed <= 0 {
		t.Errorf("Vehicle should have accelerated, speed = %f", vehicleComp.Speed)
	}

	// Should have moved right
	posComp, _ := vehicle.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 100 {
		t.Errorf("Vehicle should have moved right, X = %f", pos.X)
	}
}

func TestVehicleMovementSystem_FuelDepletion(t *testing.T) {
	world := NewWorld()
	vms := NewVehicleMovementSystem(world)

	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicle.AddComponent(&RotationComponent{Angle: 0})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.FuelAmount = 5.0 // Low fuel but enough to run a few updates
	vehicle.AddComponent(vehicleComp)

	input := NewStubInput()
	input.SetMovement(1.0, 0.0)
	vehicle.AddComponent(input)

	entities := []*Entity{vehicle}

	// Accelerate to build speed
	for i := 0; i < 10; i++ {
		vms.Update(entities, 0.1)
	}

	initialSpeed := vehicleComp.Speed
	initialFuel := vehicleComp.FuelAmount

	// Run until fuel depletes or max iterations
	depleted := false
	for i := 0; i < 1000; i++ {
		vms.Update(entities, 0.1)
		if vehicleComp.IsFuelDepleted() {
			depleted = true
			break
		}
	}

	// Should have stopped when fuel depleted
	if !depleted {
		t.Errorf("Fuel should be depleted: started with %f, now have %f, speed %f",
			initialFuel, vehicleComp.FuelAmount, vehicleComp.Speed)
	}
	if vehicleComp.Speed != 0 {
		t.Errorf("Speed should be 0 when fuel depleted, got %f", vehicleComp.Speed)
	}

	// Initial speed should have been greater than zero
	if initialSpeed <= 0 {
		t.Error("Vehicle should have accelerated before fuel depletion")
	}
}

func TestVehicleMovementSystem_Deceleration(t *testing.T) {
	world := NewWorld()
	vms := NewVehicleMovementSystem(world)

	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicle.AddComponent(&RotationComponent{Angle: 0})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Speed = 100.0 // Start at speed
	vehicle.AddComponent(vehicleComp)

	input := NewStubInput()
	input.SetMovement(0, 0) // No input
	vehicle.AddComponent(input)

	entities := []*Entity{vehicle}

	// Update - should decelerate
	for i := 0; i < 50; i++ {
		vms.Update(entities, 0.1)
		if vehicleComp.Speed <= 0 {
			break
		}
	}

	// Should have stopped or be very slow
	if vehicleComp.Speed > 1.0 {
		t.Errorf("Vehicle should have decelerated, speed = %f", vehicleComp.Speed)
	}
}

// TestVehicleDurabilitySystem tests damage and destruction
func TestVehicleDurabilitySystem(t *testing.T) {
	world := NewWorld()
	vds := NewVehicleDurabilitySystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	entities := []*Entity{vehicle}

	// Initial durability
	initialDurability := vehicleComp.Durability

	// Update - should not take damage without collisions
	vds.Update(entities, 0.1)

	if vehicleComp.Durability != initialDurability {
		t.Errorf("Durability changed without reason: %f -> %f", initialDurability, vehicleComp.Durability)
	}
}

func TestVehicleDurabilitySystem_ApplyDamage(t *testing.T) {
	world := NewWorld()
	vds := NewVehicleDurabilitySystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	initialDurability := vehicleComp.Durability

	// Apply damage
	destroyed := vds.ApplyDamage(vehicle, 20.0)

	if destroyed {
		t.Error("Vehicle should not be destroyed by 20 damage")
	}

	if vehicleComp.Durability >= initialDurability {
		t.Error("Durability should have decreased")
	}

	// Apply lethal damage
	destroyed = vds.ApplyDamage(vehicle, 1000.0)

	if !destroyed {
		t.Error("Vehicle should be destroyed by 1000 damage")
	}

	if !vehicleComp.IsDestroyed() {
		t.Error("Vehicle should report as destroyed")
	}

	if vehicleComp.Speed != 0 {
		t.Error("Destroyed vehicle should have 0 speed")
	}
}

func TestVehicleDurabilitySystem_Repair(t *testing.T) {
	world := NewWorld()
	vds := NewVehicleDurabilitySystem(world)

	vehicle := world.CreateEntity()
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Durability = 50.0 // Damaged
	vehicle.AddComponent(vehicleComp)

	// Repair
	success := vds.RepairVehicle(vehicle, 30.0)

	if !success {
		t.Error("Repair should succeed")
	}

	if vehicleComp.Durability != 80.0 {
		t.Errorf("Durability should be 80.0, got %f", vehicleComp.Durability)
	}
}

// TestMountingSystem tests rider-vehicle relationships
func TestMountingSystem_Mount(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Create vehicle
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	// Create rider
	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 95, Y: 95})

	// Process entity additions
	world.Update(0)

	// Mount rider on vehicle
	err := ms.Mount(rider, vehicle)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	// Check mount component was added
	if !ms.IsMounted(rider) {
		t.Error("Rider should be mounted")
	}

	// Check passenger count increased
	if vehicleComp.CurrentPassengers != 1 {
		t.Errorf("CurrentPassengers should be 1, got %d", vehicleComp.CurrentPassengers)
	}

	// Check mounted vehicle
	mountedVehicle := ms.GetMountedVehicle(rider)
	if mountedVehicle == nil || mountedVehicle.ID != vehicle.ID {
		t.Error("GetMountedVehicle returned wrong vehicle")
	}
}

func TestMountingSystem_Dismount(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Setup mounted rider
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicle.AddComponent(vehicleComp)

	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Process entity additions
	world.Update(0)

	ms.Mount(rider, vehicle)

	// Dismount
	err := ms.Dismount(rider)
	if err != nil {
		t.Fatalf("Dismount failed: %v", err)
	}

	// Check mount component removed
	if ms.IsMounted(rider) {
		t.Error("Rider should not be mounted after dismount")
	}

	// Check passenger count decreased
	if vehicleComp.CurrentPassengers != 0 {
		t.Errorf("CurrentPassengers should be 0, got %d", vehicleComp.CurrentPassengers)
	}
}

func TestMountingSystem_PositionSync(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Setup mounted rider
	vehicle := world.CreateEntity()
	vehiclePos := &PositionComponent{X: 100, Y: 100}
	vehicle.AddComponent(vehiclePos)
	vehicle.AddComponent(&RotationComponent{Angle: 0})
	vehicle.AddComponent(NewVehicleComponent(VehicleMount))

	rider := world.CreateEntity()
	riderPos := &PositionComponent{X: 100, Y: 100}
	rider.AddComponent(riderPos)
	rider.AddComponent(&RotationComponent{Angle: 0})

	ms.Mount(rider, vehicle)

	// Move vehicle
	vehiclePos.X = 200
	vehiclePos.Y = 200

	// Update mounting system
	entities := []*Entity{vehicle, rider}
	ms.Update(entities, 0.1)

	// Rider position should sync with vehicle
	if math.Abs(riderPos.X-vehiclePos.X) > 0.1 {
		t.Errorf("Rider X not synced: rider=%f, vehicle=%f", riderPos.X, vehiclePos.X)
	}
	if math.Abs(riderPos.Y-vehiclePos.Y) > 0.1 {
		t.Errorf("Rider Y not synced: rider=%f, vehicle=%f", riderPos.Y, vehiclePos.Y)
	}
}

func TestMountingSystem_MountErrors(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicle.AddComponent(NewVehicleComponent(VehicleMount))

	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Test nil rider
	err := ms.Mount(nil, vehicle)
	if err == nil {
		t.Error("Mount should fail with nil rider")
	}

	// Test nil vehicle
	err = ms.Mount(rider, nil)
	if err == nil {
		t.Error("Mount should fail with nil vehicle")
	}

	// Mount rider
	ms.Mount(rider, vehicle)

	// Test already mounted
	err = ms.Mount(rider, vehicle)
	if err == nil {
		t.Error("Mount should fail when already mounted")
	}
}

func TestMountingSystem_CapacityLimit(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Create vehicle with capacity 1
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicleComp := NewVehicleComponent(VehicleMount) // Capacity = 1
	vehicle.AddComponent(vehicleComp)

	// Create two riders
	rider1 := world.CreateEntity()
	rider1.AddComponent(&PositionComponent{X: 100, Y: 100})

	rider2 := world.CreateEntity()
	rider2.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Mount first rider - should succeed
	err := ms.Mount(rider1, vehicle)
	if err != nil {
		t.Fatalf("First mount failed: %v", err)
	}

	// Try to mount second rider - should fail (capacity full)
	err = ms.Mount(rider2, vehicle)
	if err == nil {
		t.Error("Second mount should fail (vehicle at capacity)")
	}
}

func TestMountingSystem_DestroyedVehicle(t *testing.T) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Create destroyed vehicle
	vehicle := world.CreateEntity()
	vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
	vehicleComp := NewVehicleComponent(VehicleMount)
	vehicleComp.Durability = 0 // Destroyed
	vehicle.AddComponent(vehicleComp)

	rider := world.CreateEntity()
	rider.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Try to mount on destroyed vehicle
	err := ms.Mount(rider, vehicle)
	if err == nil {
		t.Error("Mount should fail on destroyed vehicle")
	}
}

// Benchmark system operations
func BenchmarkVehicleMovementSystem_Update(b *testing.B) {
	world := NewWorld()
	vms := NewVehicleMovementSystem(world)

	// Create 10 vehicles
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		vehicle := world.CreateEntity()
		vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
		vehicle.AddComponent(&RotationComponent{Angle: 0})
		vehicle.AddComponent(NewVehicleComponent(VehicleMount))
		vehicle.AddComponent(NewStubInput())
		entities[i] = vehicle
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vms.Update(entities, 0.016) // 60 FPS delta
	}
}

func BenchmarkMountingSystem_Update(b *testing.B) {
	world := NewWorld()
	ms := NewMountingSystem(world)

	// Create 10 mounted riders
	entities := make([]*Entity, 20) // 10 vehicles + 10 riders
	for i := 0; i < 10; i++ {
		vehicle := world.CreateEntity()
		vehicle.AddComponent(&PositionComponent{X: 100, Y: 100})
		vehicle.AddComponent(&RotationComponent{Angle: 0})
		vehicle.AddComponent(NewVehicleComponent(VehicleMount))
		entities[i*2] = vehicle

		rider := world.CreateEntity()
		rider.AddComponent(&PositionComponent{X: 100, Y: 100})
		rider.AddComponent(&RotationComponent{Angle: 0})
		entities[i*2+1] = rider

		ms.Mount(rider, vehicle)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.Update(entities, 0.016)
	}
}
