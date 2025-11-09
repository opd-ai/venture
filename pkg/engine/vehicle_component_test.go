// Package engine provides vehicle component tests.
// Tests verify vehicle stats, fuel/durability management,
// terrain traversal, and mount component relationships.
//
// Phase 21.1: Vehicle Foundation
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestVehicleTypeString(t *testing.T) {
	tests := []struct {
		vehicleType VehicleType
		expected    string
	}{
		{VehicleMount, "Mount"},
		{VehicleCart, "Cart"},
		{VehicleBoat, "Boat"},
		{VehicleGlider, "Glider"},
		{VehicleMech, "Mech"},
		{VehicleType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.vehicleType.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewVehicleComponent(t *testing.T) {
	tests := []struct {
		name         string
		vehicleType  VehicleType
		wantMaxSpeed float64
		wantFuelType string
		wantCapacity int
	}{
		{"Mount", VehicleMount, 200.0, "Stamina", 1},
		{"Cart", VehicleCart, 120.0, "Stamina", 4},
		{"Boat", VehicleBoat, 150.0, "Energy", 2},
		{"Glider", VehicleGlider, 250.0, "Energy", 1},
		{"Mech", VehicleMech, 180.0, "Energy", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc := NewVehicleComponent(tt.vehicleType)

			if vc.Type() != "vehicle" {
				t.Errorf("Type() = %q, want %q", vc.Type(), "vehicle")
			}

			if vc.MaxSpeed != tt.wantMaxSpeed {
				t.Errorf("MaxSpeed = %f, want %f", vc.MaxSpeed, tt.wantMaxSpeed)
			}

			if vc.FuelType != tt.wantFuelType {
				t.Errorf("FuelType = %q, want %q", vc.FuelType, tt.wantFuelType)
			}

			if vc.Capacity != tt.wantCapacity {
				t.Errorf("Capacity = %d, want %d", vc.Capacity, tt.wantCapacity)
			}

			if vc.Speed != 0.0 {
				t.Errorf("Speed should start at 0.0, got %f", vc.Speed)
			}

			if vc.Durability != vc.MaxDurability {
				t.Errorf("Durability = %f, want MaxDurability %f", vc.Durability, vc.MaxDurability)
			}

			if vc.FuelAmount != vc.FuelCapacity {
				t.Errorf("FuelAmount = %f, want FuelCapacity %f", vc.FuelAmount, vc.FuelCapacity)
			}

			if vc.CurrentPassengers != 0 {
				t.Errorf("CurrentPassengers should start at 0, got %d", vc.CurrentPassengers)
			}
		})
	}
}

func TestVehicleComponent_Getters(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	vc.Speed = 100.0

	if got := vc.GetSpeed(); got != 100.0 {
		t.Errorf("GetSpeed() = %f, want 100.0", got)
	}

	if got := vc.GetMaxSpeed(); got != vc.MaxSpeed {
		t.Errorf("GetMaxSpeed() = %f, want %f", got, vc.MaxSpeed)
	}

	if got := vc.GetAcceleration(); got != vc.Acceleration {
		t.Errorf("GetAcceleration() = %f, want %f", got, vc.Acceleration)
	}

	if got := vc.GetHandling(); got != vc.Handling {
		t.Errorf("GetHandling() = %f, want %f", got, vc.Handling)
	}
}

func TestVehicleComponent_CanTraverse(t *testing.T) {
	tests := []struct {
		name         string
		vehicleType  VehicleType
		terrainType  terrain.TileType
		canTraverse  bool
	}{
		// Mount can traverse most ground terrain
		{"Mount on floor", VehicleMount, terrain.TileFloor, true},
		{"Mount on shallow water", VehicleMount, terrain.TileWaterShallow, true},
		{"Mount on deep water", VehicleMount, terrain.TileWaterDeep, false},
		{"Mount on wall", VehicleMount, terrain.TileWall, false},

		// Boat can only traverse water
		{"Boat on deep water", VehicleBoat, terrain.TileWaterDeep, true},
		{"Boat on shallow water", VehicleBoat, terrain.TileWaterShallow, true},
		{"Boat on floor", VehicleBoat, terrain.TileFloor, false},

		// Glider can fly over most terrain
		{"Glider on floor", VehicleGlider, terrain.TileFloor, true},
		{"Glider on water", VehicleGlider, terrain.TileWaterDeep, true},
		{"Glider on tree", VehicleGlider, terrain.TileTree, true},
		{"Glider on wall", VehicleGlider, terrain.TileWall, false},

		// Cart has limited terrain
		{"Cart on floor", VehicleCart, terrain.TileFloor, true},
		{"Cart on corridor", VehicleCart, terrain.TileCorridor, true},
		{"Cart on shallow water", VehicleCart, terrain.TileWaterShallow, false},

		// Mech can traverse almost anything
		{"Mech on floor", VehicleMech, terrain.TileFloor, true},
		{"Mech on water", VehicleMech, terrain.TileWaterDeep, true},
		{"Mech on tree", VehicleMech, terrain.TileTree, true},
		{"Mech on wall", VehicleMech, terrain.TileWall, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc := NewVehicleComponent(tt.vehicleType)
			got := vc.CanTraverse(int(tt.terrainType))
			if got != tt.canTraverse {
				t.Errorf("CanTraverse(%v) = %v, want %v", tt.terrainType, got, tt.canTraverse)
			}
		})
	}
}

func TestVehicleComponent_GetFuelCost(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	
	// Test fuel cost scales with speed
	vc.Speed = 0.0
	if cost := vc.GetFuelCost(); cost != 0.0 {
		t.Errorf("GetFuelCost() at speed 0 = %f, want 0.0", cost)
	}

	vc.Speed = 100.0
	if cost := vc.GetFuelCost(); cost != 1.0 {
		t.Errorf("GetFuelCost() at speed 100 = %f, want 1.0", cost)
	}

	vc.Speed = 200.0
	if cost := vc.GetFuelCost(); cost != 2.0 {
		t.Errorf("GetFuelCost() at speed 200 = %f, want 2.0", cost)
	}
}

func TestVehicleComponent_ConsumeFuel(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	initialFuel := vc.FuelAmount

	// Consume valid amount
	if !vc.ConsumeFuel(10.0) {
		t.Error("ConsumeFuel(10.0) returned false, expected true")
	}
	if vc.FuelAmount != initialFuel-10.0 {
		t.Errorf("FuelAmount = %f, want %f", vc.FuelAmount, initialFuel-10.0)
	}

	// Try to consume more than available
	vc.FuelAmount = 5.0
	if vc.ConsumeFuel(10.0) {
		t.Error("ConsumeFuel(10.0) with 5.0 fuel returned true, expected false")
	}
	if vc.FuelAmount != 5.0 {
		t.Errorf("FuelAmount should remain 5.0, got %f", vc.FuelAmount)
	}

	// Consume all remaining fuel
	if !vc.ConsumeFuel(5.0) {
		t.Error("ConsumeFuel(5.0) returned false, expected true")
	}
	if vc.FuelAmount != 0.0 {
		t.Errorf("FuelAmount = %f, want 0.0", vc.FuelAmount)
	}
}

func TestVehicleComponent_RefillFuel(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	vc.FuelAmount = 20.0

	// Refill partial
	vc.RefillFuel(30.0)
	if vc.FuelAmount != 50.0 {
		t.Errorf("FuelAmount = %f, want 50.0", vc.FuelAmount)
	}

	// Refill beyond capacity (should cap at FuelCapacity)
	vc.RefillFuel(1000.0)
	if vc.FuelAmount != vc.FuelCapacity {
		t.Errorf("FuelAmount = %f, want FuelCapacity %f", vc.FuelAmount, vc.FuelCapacity)
	}
}

func TestVehicleComponent_IsFuelDepleted(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)

	if vc.IsFuelDepleted() {
		t.Error("IsFuelDepleted() = true for full fuel, want false")
	}

	vc.FuelAmount = 0.0
	if !vc.IsFuelDepleted() {
		t.Error("IsFuelDepleted() = false for 0 fuel, want true")
	}
}

func TestVehicleComponent_TakeDamage(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	initialDurability := vc.Durability

	// Take damage but not destroyed
	if vc.TakeDamage(20.0) {
		t.Error("TakeDamage(20.0) returned true (destroyed), expected false")
	}
	if vc.Durability != initialDurability-20.0 {
		t.Errorf("Durability = %f, want %f", vc.Durability, initialDurability-20.0)
	}

	// Take damage to destroy
	vc.Durability = 10.0
	if !vc.TakeDamage(15.0) {
		t.Error("TakeDamage(15.0) with 10.0 durability returned false, expected true (destroyed)")
	}
	if vc.Durability != 0.0 {
		t.Errorf("Durability = %f, want 0.0", vc.Durability)
	}
}

func TestVehicleComponent_Repair(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)
	vc.Durability = 30.0

	// Repair partial
	vc.Repair(20.0)
	if vc.Durability != 50.0 {
		t.Errorf("Durability = %f, want 50.0", vc.Durability)
	}

	// Repair beyond maximum (should cap)
	vc.Repair(1000.0)
	if vc.Durability != vc.MaxDurability {
		t.Errorf("Durability = %f, want MaxDurability %f", vc.Durability, vc.MaxDurability)
	}
}

func TestVehicleComponent_IsDestroyed(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount)

	if vc.IsDestroyed() {
		t.Error("IsDestroyed() = true for full durability, want false")
	}

	vc.Durability = 0.0
	if !vc.IsDestroyed() {
		t.Error("IsDestroyed() = false for 0 durability, want true")
	}
}

func TestVehicleComponent_PassengerManagement(t *testing.T) {
	vc := NewVehicleComponent(VehicleMount) // Capacity = 1

	// Can add passenger
	if !vc.CanAddPassenger() {
		t.Error("CanAddPassenger() = false for empty vehicle, want true")
	}

	// Add passenger
	if !vc.AddPassenger() {
		t.Error("AddPassenger() returned false, expected true")
	}
	if vc.CurrentPassengers != 1 {
		t.Errorf("CurrentPassengers = %d, want 1", vc.CurrentPassengers)
	}

	// At capacity
	if vc.CanAddPassenger() {
		t.Error("CanAddPassenger() = true when at capacity, want false")
	}

	// Try to add beyond capacity
	if vc.AddPassenger() {
		t.Error("AddPassenger() returned true when at capacity, expected false")
	}

	// Remove passenger
	vc.RemovePassenger()
	if vc.CurrentPassengers != 0 {
		t.Errorf("CurrentPassengers = %d, want 0", vc.CurrentPassengers)
	}

	// Remove from empty (should not go negative)
	vc.RemovePassenger()
	if vc.CurrentPassengers != 0 {
		t.Errorf("CurrentPassengers = %d, should stay at 0", vc.CurrentPassengers)
	}
}

func TestVehicleComponent_CartCapacity(t *testing.T) {
	vc := NewVehicleComponent(VehicleCart) // Capacity = 4

	// Add multiple passengers
	for i := 0; i < 4; i++ {
		if !vc.AddPassenger() {
			t.Errorf("AddPassenger() failed at passenger %d", i+1)
		}
	}

	if vc.CurrentPassengers != 4 {
		t.Errorf("CurrentPassengers = %d, want 4", vc.CurrentPassengers)
	}

	// Should be at capacity
	if vc.AddPassenger() {
		t.Error("AddPassenger() succeeded when at capacity 4")
	}
}

func TestMountComponent(t *testing.T) {
	vehicleID := uint64(12345)
	offsetX, offsetY := 10.0, 15.0

	mc := NewMountComponent(vehicleID, offsetX, offsetY)

	if mc.Type() != "mount" {
		t.Errorf("Type() = %q, want %q", mc.Type(), "mount")
	}

	if mc.MountedEntityID != vehicleID {
		t.Errorf("MountedEntityID = %d, want %d", mc.MountedEntityID, vehicleID)
	}

	if mc.MountOffset.X != offsetX {
		t.Errorf("MountOffset.X = %f, want %f", mc.MountOffset.X, offsetX)
	}

	if mc.MountOffset.Y != offsetY {
		t.Errorf("MountOffset.Y = %f, want %f", mc.MountOffset.Y, offsetY)
	}
}

// Benchmark vehicle operations
func BenchmarkNewVehicleComponent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewVehicleComponent(VehicleMount)
	}
}

func BenchmarkVehicleComponent_CanTraverse(b *testing.B) {
	vc := NewVehicleComponent(VehicleMount)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vc.CanTraverse(int(terrain.TileFloor))
	}
}

func BenchmarkVehicleComponent_ConsumeFuel(b *testing.B) {
	vc := NewVehicleComponent(VehicleMount)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.FuelAmount = 100.0 // Reset to avoid running out
		_ = vc.ConsumeFuel(1.0)
	}
}
