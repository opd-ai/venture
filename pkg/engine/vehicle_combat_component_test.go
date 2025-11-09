// Package engine provides tests for vehicle combat components.
// Phase 21.2: Advanced Vehicle Features
package engine

import (
	"testing"
)

func TestVehicleCombatComponent_Type(t *testing.T) {
	comp := NewVehicleCombatComponent()
	if comp.Type() != "vehicle_combat" {
		t.Errorf("Expected type 'vehicle_combat', got '%s'", comp.Type())
	}
}

func TestVehicleCombatComponent_CanRam(t *testing.T) {
	tests := []struct {
		name         string
		currentSpeed float64
		minSpeed     float64
		cooldown     float64
		expected     bool
	}{
		{"sufficient speed and ready", 100.0, 50.0, 0.0, true},
		{"insufficient speed", 30.0, 50.0, 0.0, false},
		{"on cooldown", 100.0, 50.0, 0.5, false},
		{"exactly min speed", 50.0, 50.0, 0.0, true},
		{"zero speed", 0.0, 50.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewVehicleCombatComponent()
			comp.MinRammingSpeed = tt.minSpeed
			comp.CurrentRammingCooldown = tt.cooldown

			if got := comp.CanRam(tt.currentSpeed); got != tt.expected {
				t.Errorf("CanRam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVehicleCombatComponent_CanShoot(t *testing.T) {
	tests := []struct {
		name          string
		weaponMounted bool
		cooldown      float64
		expected      bool
	}{
		{"mounted and ready", true, 0.0, true},
		{"not mounted", false, 0.0, false},
		{"on cooldown", true, 0.5, false},
		{"not mounted and on cooldown", false, 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewVehicleCombatComponent()
			comp.WeaponMounted = tt.weaponMounted
			comp.CurrentWeaponCooldown = tt.cooldown

			if got := comp.CanShoot(); got != tt.expected {
				t.Errorf("CanShoot() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVehicleCombatComponent_UpdateCooldowns(t *testing.T) {
	comp := NewVehicleCombatComponent()
	comp.CurrentRammingCooldown = 1.0
	comp.CurrentWeaponCooldown = 0.5

	comp.UpdateCooldowns(0.3)

	if comp.CurrentRammingCooldown != 0.7 {
		t.Errorf("Expected ramming cooldown 0.7, got %f", comp.CurrentRammingCooldown)
	}
	if comp.CurrentWeaponCooldown != 0.2 {
		t.Errorf("Expected weapon cooldown 0.2, got %f", comp.CurrentWeaponCooldown)
	}

	// Update past zero
	comp.UpdateCooldowns(1.0)

	if comp.CurrentRammingCooldown != 0.0 {
		t.Errorf("Expected ramming cooldown 0.0, got %f", comp.CurrentRammingCooldown)
	}
	if comp.CurrentWeaponCooldown != 0.0 {
		t.Errorf("Expected weapon cooldown 0.0, got %f", comp.CurrentWeaponCooldown)
	}
}

func TestVehicleCombatComponent_ExecuteRam(t *testing.T) {
	comp := NewVehicleCombatComponent()
	comp.RammingCooldown = 2.0
	comp.CurrentRammingCooldown = 0.0

	comp.ExecuteRam()

	if comp.CurrentRammingCooldown != 2.0 {
		t.Errorf("Expected cooldown set to 2.0, got %f", comp.CurrentRammingCooldown)
	}
}

func TestVehicleCombatComponent_ExecuteShot(t *testing.T) {
	comp := NewVehicleCombatComponent()
	comp.WeaponCooldown = 0.5
	comp.CurrentWeaponCooldown = 0.0

	comp.ExecuteShot()

	if comp.CurrentWeaponCooldown != 0.5 {
		t.Errorf("Expected cooldown set to 0.5, got %f", comp.CurrentWeaponCooldown)
	}
}

func TestVehicleCombatComponent_CalculateRammingDamage(t *testing.T) {
	tests := []struct {
		name         string
		baseDamage   float64
		minSpeed     float64
		currentSpeed float64
		expected     float64
	}{
		{"below min speed", 20.0, 50.0, 30.0, 0.0},
		{"exactly min speed", 20.0, 50.0, 50.0, 20.0},
		{"double min speed", 20.0, 50.0, 100.0, 40.0},
		{"triple min speed", 20.0, 50.0, 150.0, 60.0},
		{"zero speed", 20.0, 50.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewVehicleCombatComponent()
			comp.RammingDamage = tt.baseDamage
			comp.MinRammingSpeed = tt.minSpeed

			got := comp.CalculateRammingDamage(tt.currentSpeed)
			if got != tt.expected {
				t.Errorf("CalculateRammingDamage(%f) = %f, want %f",
					tt.currentSpeed, got, tt.expected)
			}
		})
	}
}

func TestCargoComponent_Type(t *testing.T) {
	comp := NewCargoComponent(10, 100.0)
	if comp.Type() != "cargo" {
		t.Errorf("Expected type 'cargo', got '%s'", comp.Type())
	}
}

func TestCargoComponent_AddRemoveItem(t *testing.T) {
	comp := NewCargoComponent(3, 100.0)

	// Add items
	if !comp.AddItem(1, 10.0) {
		t.Error("Failed to add first item")
	}
	if comp.GetItemCount() != 1 {
		t.Errorf("Expected 1 item, got %d", comp.GetItemCount())
	}
	if comp.CurrentWeight != 10.0 {
		t.Errorf("Expected weight 10.0, got %f", comp.CurrentWeight)
	}

	if !comp.AddItem(2, 20.0) {
		t.Error("Failed to add second item")
	}
	if comp.GetItemCount() != 2 {
		t.Errorf("Expected 2 items, got %d", comp.GetItemCount())
	}
	if comp.CurrentWeight != 30.0 {
		t.Errorf("Expected weight 30.0, got %f", comp.CurrentWeight)
	}

	// Remove item
	if !comp.RemoveItem(1, 10.0) {
		t.Error("Failed to remove item")
	}
	if comp.GetItemCount() != 1 {
		t.Errorf("Expected 1 item after removal, got %d", comp.GetItemCount())
	}
	if comp.CurrentWeight != 20.0 {
		t.Errorf("Expected weight 20.0 after removal, got %f", comp.CurrentWeight)
	}

	// Try to remove non-existent item
	if comp.RemoveItem(999, 10.0) {
		t.Error("Should not remove non-existent item")
	}
}

func TestCargoComponent_CapacityLimits(t *testing.T) {
	comp := NewCargoComponent(2, 50.0)

	// Fill to capacity
	comp.AddItem(1, 20.0)
	comp.AddItem(2, 20.0)

	if !comp.IsFull() {
		t.Error("Should be full after adding 2 items (max 2)")
	}

	// Try to add beyond capacity
	if comp.AddItem(3, 5.0) {
		t.Error("Should not add item beyond capacity")
	}

	// Test weight limit
	comp2 := NewCargoComponent(10, 50.0)
	comp2.AddItem(1, 30.0)

	// Try to add item that would exceed weight
	if comp2.AddItem(2, 25.0) {
		t.Error("Should not add item that exceeds weight capacity")
	}

	// Add item within weight limit
	if !comp2.AddItem(2, 15.0) {
		t.Error("Should add item within weight capacity")
	}
}

func TestCargoComponent_IsEmpty(t *testing.T) {
	comp := NewCargoComponent(10, 100.0)

	if !comp.IsEmpty() {
		t.Error("New cargo should be empty")
	}

	comp.AddItem(1, 10.0)
	if comp.IsEmpty() {
		t.Error("Cargo should not be empty after adding item")
	}

	comp.RemoveItem(1, 10.0)
	if !comp.IsEmpty() {
		t.Error("Cargo should be empty after removing all items")
	}
}

func TestUpgradeSlotComponent_Type(t *testing.T) {
	comp := NewUpgradeSlotComponent(5)
	if comp.Type() != "upgrade_slots" {
		t.Errorf("Expected type 'upgrade_slots', got '%s'", comp.Type())
	}
}

func TestUpgradeSlotComponent_AddRemoveUpgrade(t *testing.T) {
	comp := NewUpgradeSlotComponent(3)

	upgrade1 := &VehicleUpgrade{
		Name:             "Turbo Engine",
		Type:             UpgradeSpeed,
		Value:            1.5,
		IsMultiplicative: true,
	}

	// Add upgrade
	if !comp.AddUpgrade(upgrade1) {
		t.Error("Failed to add upgrade")
	}
	if comp.GetUpgradeCount() != 1 {
		t.Errorf("Expected 1 upgrade, got %d", comp.GetUpgradeCount())
	}
	if !comp.HasUpgrade("Turbo Engine") {
		t.Error("Should have 'Turbo Engine' upgrade")
	}

	// Add more upgrades
	upgrade2 := &VehicleUpgrade{Name: "Heavy Armor", Type: UpgradeArmor, Value: 10.0}
	upgrade3 := &VehicleUpgrade{Name: "Fuel Tank", Type: UpgradeFuelCapacity, Value: 50.0}

	comp.AddUpgrade(upgrade2)
	comp.AddUpgrade(upgrade3)

	if comp.GetUpgradeCount() != 3 {
		t.Errorf("Expected 3 upgrades, got %d", comp.GetUpgradeCount())
	}

	// Try to add beyond capacity
	upgrade4 := &VehicleUpgrade{Name: "Extra", Type: UpgradeSpeed, Value: 1.0}
	if comp.AddUpgrade(upgrade4) {
		t.Error("Should not add upgrade beyond capacity")
	}

	// Remove upgrade
	if !comp.RemoveUpgrade("Turbo Engine") {
		t.Error("Failed to remove upgrade")
	}
	if comp.GetUpgradeCount() != 2 {
		t.Errorf("Expected 2 upgrades after removal, got %d", comp.GetUpgradeCount())
	}
	if comp.HasUpgrade("Turbo Engine") {
		t.Error("Should not have removed upgrade")
	}

	// Try to remove non-existent upgrade
	if comp.RemoveUpgrade("NonExistent") {
		t.Error("Should not remove non-existent upgrade")
	}
}

func TestUpgradeType_String(t *testing.T) {
	tests := []struct {
		upgradeType UpgradeType
		expected    string
	}{
		{UpgradeSpeed, "Speed"},
		{UpgradeAcceleration, "Acceleration"},
		{UpgradeHandling, "Handling"},
		{UpgradeDurability, "Durability"},
		{UpgradeArmor, "Armor"},
		{UpgradeCapacity, "Capacity"},
		{UpgradeFuelCapacity, "FuelCapacity"},
		{UpgradeWeaponDamage, "WeaponDamage"},
		{UpgradeType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.upgradeType.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyUpgrades(t *testing.T) {
	vehicle := NewVehicleComponent(VehicleMount)
	combat := NewVehicleCombatComponent()
	upgrades := NewUpgradeSlotComponent(5)

	// Record initial stats
	initialMaxSpeed := vehicle.MaxSpeed
	initialAcceleration := vehicle.Acceleration
	initialArmor := combat.ArmorRating

	// Add multiplicative speed upgrade
	upgrades.AddUpgrade(&VehicleUpgrade{
		Name:             "Turbo",
		Type:             UpgradeSpeed,
		Value:            1.5,
		IsMultiplicative: true,
	})

	// Add additive acceleration upgrade
	upgrades.AddUpgrade(&VehicleUpgrade{
		Name:             "Better Engine",
		Type:             UpgradeAcceleration,
		Value:            20.0,
		IsMultiplicative: false,
	})

	// Add armor upgrade
	upgrades.AddUpgrade(&VehicleUpgrade{
		Name:             "Heavy Plates",
		Type:             UpgradeArmor,
		Value:            15.0,
		IsMultiplicative: false,
	})

	ApplyUpgrades(vehicle, combat, upgrades)

	// Verify multiplicative speed upgrade
	expectedSpeed := initialMaxSpeed * 1.5
	if vehicle.MaxSpeed != expectedSpeed {
		t.Errorf("Expected max speed %f, got %f", expectedSpeed, vehicle.MaxSpeed)
	}

	// Verify additive acceleration upgrade
	expectedAccel := initialAcceleration + 20.0
	if vehicle.Acceleration != expectedAccel {
		t.Errorf("Expected acceleration %f, got %f", expectedAccel, vehicle.Acceleration)
	}

	// Verify armor upgrade
	expectedArmor := initialArmor + 15.0
	if combat.ArmorRating != expectedArmor {
		t.Errorf("Expected armor %f, got %f", expectedArmor, combat.ArmorRating)
	}
}
