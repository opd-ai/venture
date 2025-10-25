package engine

import (
	"testing"
)

// TestNewEquipmentVisualComponent tests component initialization.
func TestNewEquipmentVisualComponent(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	if comp == nil {
		t.Fatal("Expected non-nil component")
	}

	if comp.Type() != "equipment_visual" {
		t.Errorf("Expected type 'equipment_visual', got '%s'", comp.Type())
	}

	if !comp.Dirty {
		t.Error("Expected component to be dirty initially")
	}

	if comp.HasWeapon() {
		t.Error("Expected no weapon initially")
	}

	if comp.HasArmor() {
		t.Error("Expected no armor initially")
	}

	if comp.HasAccessories() {
		t.Error("Expected no accessories initially")
	}
}

// TestEquipmentVisualComponent_SetWeapon tests weapon setting.
func TestEquipmentVisualComponent_SetWeapon(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.Dirty = false

	comp.SetWeapon("sword_001", 12345)

	if !comp.HasWeapon() {
		t.Error("Expected weapon to be set")
	}

	if comp.WeaponID != "sword_001" {
		t.Errorf("Expected weapon ID 'sword_001', got '%s'", comp.WeaponID)
	}

	if comp.WeaponSeed != 12345 {
		t.Errorf("Expected weapon seed 12345, got %d", comp.WeaponSeed)
	}

	if !comp.Dirty {
		t.Error("Expected component to be marked dirty after setting weapon")
	}
}

// TestEquipmentVisualComponent_SetArmor tests armor setting.
func TestEquipmentVisualComponent_SetArmor(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.Dirty = false

	comp.SetArmor("plate_armor_001", 54321)

	if !comp.HasArmor() {
		t.Error("Expected armor to be set")
	}

	if comp.ArmorID != "plate_armor_001" {
		t.Errorf("Expected armor ID 'plate_armor_001', got '%s'", comp.ArmorID)
	}

	if comp.ArmorSeed != 54321 {
		t.Errorf("Expected armor seed 54321, got %d", comp.ArmorSeed)
	}

	if !comp.Dirty {
		t.Error("Expected component to be marked dirty after setting armor")
	}
}

// TestEquipmentVisualComponent_AddAccessory tests accessory addition.
func TestEquipmentVisualComponent_AddAccessory(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	comp.AddAccessory("ring_001", 11111)
	comp.AddAccessory("amulet_001", 22222)

	if !comp.HasAccessories() {
		t.Error("Expected accessories to be set")
	}

	if len(comp.AccessoryIDs) != 2 {
		t.Errorf("Expected 2 accessories, got %d", len(comp.AccessoryIDs))
	}

	if comp.AccessoryIDs[0] != "ring_001" {
		t.Errorf("Expected first accessory 'ring_001', got '%s'", comp.AccessoryIDs[0])
	}

	if comp.AccessorySeeds[1] != 22222 {
		t.Errorf("Expected second accessory seed 22222, got %d", comp.AccessorySeeds[1])
	}
}

// TestEquipmentVisualComponent_RemoveAccessory tests accessory removal.
func TestEquipmentVisualComponent_RemoveAccessory(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	comp.AddAccessory("ring_001", 11111)
	comp.AddAccessory("amulet_001", 22222)
	comp.AddAccessory("cape_001", 33333)

	// Remove middle accessory
	comp.RemoveAccessory(1)

	if len(comp.AccessoryIDs) != 2 {
		t.Errorf("Expected 2 accessories after removal, got %d", len(comp.AccessoryIDs))
	}

	if comp.AccessoryIDs[1] != "cape_001" {
		t.Errorf("Expected second accessory 'cape_001', got '%s'", comp.AccessoryIDs[1])
	}

	// Test out of bounds removal
	comp.RemoveAccessory(-1)
	if len(comp.AccessoryIDs) != 2 {
		t.Error("Removal with negative index should be no-op")
	}

	comp.RemoveAccessory(10)
	if len(comp.AccessoryIDs) != 2 {
		t.Error("Removal with out of bounds index should be no-op")
	}
}

// TestEquipmentVisualComponent_ClearWeapon tests weapon clearing.
func TestEquipmentVisualComponent_ClearWeapon(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.SetWeapon("sword_001", 12345)
	comp.Dirty = false

	comp.ClearWeapon()

	if comp.HasWeapon() {
		t.Error("Expected weapon to be cleared")
	}

	if comp.WeaponID != "" {
		t.Errorf("Expected empty weapon ID, got '%s'", comp.WeaponID)
	}

	if !comp.Dirty {
		t.Error("Expected component to be marked dirty after clearing weapon")
	}

	// Clearing already empty weapon should not mark dirty
	comp.Dirty = false
	comp.ClearWeapon()

	if comp.Dirty {
		t.Error("Clearing empty weapon should not mark dirty")
	}
}

// TestEquipmentVisualComponent_ClearArmor tests armor clearing.
func TestEquipmentVisualComponent_ClearArmor(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.SetArmor("plate_armor_001", 54321)
	comp.Dirty = false

	comp.ClearArmor()

	if comp.HasArmor() {
		t.Error("Expected armor to be cleared")
	}

	if comp.ArmorID != "" {
		t.Errorf("Expected empty armor ID, got '%s'", comp.ArmorID)
	}

	if !comp.Dirty {
		t.Error("Expected component to be marked dirty after clearing armor")
	}
}

// TestEquipmentVisualComponent_ClearAccessories tests accessory clearing.
func TestEquipmentVisualComponent_ClearAccessories(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.AddAccessory("ring_001", 11111)
	comp.AddAccessory("amulet_001", 22222)
	comp.Dirty = false

	comp.ClearAccessories()

	if comp.HasAccessories() {
		t.Error("Expected accessories to be cleared")
	}

	if len(comp.AccessoryIDs) != 0 {
		t.Errorf("Expected 0 accessories, got %d", len(comp.AccessoryIDs))
	}

	if !comp.Dirty {
		t.Error("Expected component to be marked dirty after clearing accessories")
	}
}

// TestEquipmentVisualComponent_MarkClean tests dirty flag management.
func TestEquipmentVisualComponent_MarkClean(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	if !comp.Dirty {
		t.Error("Expected component to start dirty")
	}

	comp.MarkClean()

	if comp.Dirty {
		t.Error("Expected component to be clean after MarkClean()")
	}
}

// TestEquipmentVisualComponent_MarkDirty tests forced dirty marking.
func TestEquipmentVisualComponent_MarkDirty(t *testing.T) {
	comp := NewEquipmentVisualComponent()
	comp.Dirty = false

	comp.MarkDirty()

	if !comp.Dirty {
		t.Error("Expected component to be dirty after MarkDirty()")
	}
}

// TestEquipmentVisualComponent_VisibilityFlags tests layer visibility.
func TestEquipmentVisualComponent_VisibilityFlags(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	if !comp.ShowWeapon {
		t.Error("Expected weapon to be visible by default")
	}

	if !comp.ShowArmor {
		t.Error("Expected armor to be visible by default")
	}

	if !comp.ShowAccessories {
		t.Error("Expected accessories to be visible by default")
	}

	// Test toggling visibility
	comp.ShowWeapon = false
	if comp.ShowWeapon {
		t.Error("Expected weapon visibility to be toggled")
	}
}

// TestEquipmentVisualComponent_SetSameItem tests setting same item twice.
func TestEquipmentVisualComponent_SetSameItem(t *testing.T) {
	comp := NewEquipmentVisualComponent()

	comp.SetWeapon("sword_001", 12345)
	comp.Dirty = false

	// Setting same weapon should not mark dirty
	comp.SetWeapon("sword_001", 12345)

	if comp.Dirty {
		t.Error("Setting same weapon should not mark dirty")
	}

	// Setting different weapon should mark dirty
	comp.SetWeapon("axe_001", 54321)

	if !comp.Dirty {
		t.Error("Setting different weapon should mark dirty")
	}
}

// BenchmarkEquipmentVisualComponent_SetWeapon benchmarks weapon setting.
func BenchmarkEquipmentVisualComponent_SetWeapon(b *testing.B) {
	comp := NewEquipmentVisualComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.SetWeapon("sword_001", 12345)
	}
}

// BenchmarkEquipmentVisualComponent_AddAccessory benchmarks accessory addition.
func BenchmarkEquipmentVisualComponent_AddAccessory(b *testing.B) {
	comp := NewEquipmentVisualComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddAccessory("ring_001", 11111)
	}
}
