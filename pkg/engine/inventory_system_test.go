package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestInventorySystem_AddRemoveItem tests adding and removing items from inventory.
func TestInventorySystem_AddRemoveItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create entity with inventory
	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	world.Update(0.0) // Process entity additions

	// Add item
	testItem := createTestItem("Sword", item.TypeWeapon, 5.0, 100, 10, 0)
	success, err := system.AddItemToInventory(entity.ID, testItem)
	if err != nil {
		t.Fatalf("AddItemToInventory failed: %v", err)
	}
	if !success {
		t.Error("AddItemToInventory returned false")
	}

	// Verify item was added
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)
	if inv.GetItemCount() != 1 {
		t.Errorf("Item count = %d, want 1", inv.GetItemCount())
	}

	// Remove item
	removed, err := system.RemoveItemFromInventory(entity.ID, 0)
	if err != nil {
		t.Fatalf("RemoveItemFromInventory failed: %v", err)
	}
	if removed != testItem {
		t.Error("Removed wrong item")
	}
	if inv.GetItemCount() != 0 {
		t.Errorf("Item count after removal = %d, want 0", inv.GetItemCount())
	}
}

// TestInventorySystem_EquipItem tests equipping items from inventory.
func TestInventorySystem_EquipItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create entity with inventory and equipment
	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&StatsComponent{})
	entity.AddComponent(&AttackComponent{})

	world.Update(0.0) // Process entity additions
	// Add weapon to inventory
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)
	inv.AddItem(sword)

	// Equip weapon
	err := system.EquipItem(entity.ID, 0)
	if err != nil {
		t.Fatalf("EquipItem failed: %v", err)
	}

	// Verify item was equipped and removed from inventory
	if inv.GetItemCount() != 0 {
		t.Errorf("Inventory count = %d, want 0 after equipping", inv.GetItemCount())
	}

	comp2, _ := entity.GetComponent("equipment")
	equip := comp2.(*EquipmentComponent)
	equipped := equip.GetEquipped(SlotMainHand)
	if equipped != sword {
		t.Error("Sword not equipped in main hand")
	}
}

// TestInventorySystem_EquipWithSwap tests equipping when slot already has an item.
func TestInventorySystem_EquipWithSwap(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&StatsComponent{})
	entity.AddComponent(&AttackComponent{})

	world.Update(0.0) // Process entity additions
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)
	comp2, _ := entity.GetComponent("equipment")
	equip := comp2.(*EquipmentComponent)

	// Equip first weapon
	sword1 := createTestWeapon("Sword 1", item.WeaponSword, 15, 5.0, 1.2)
	equip.Equip(sword1, SlotMainHand)

	// Add second weapon to inventory
	sword2 := createTestWeapon("Sword 2", item.WeaponSword, 18, 6.0, 1.1)
	inv.AddItem(sword2)

	// Equip second weapon (should swap with first)
	err := system.EquipItem(entity.ID, 0)
	if err != nil {
		t.Fatalf("EquipItem with swap failed: %v", err)
	}

	// Verify sword2 is equipped
	equipped := equip.GetEquipped(SlotMainHand)
	if equipped != sword2 {
		t.Error("Sword 2 not equipped")
	}

	// Verify sword1 is back in inventory
	if inv.GetItemCount() != 1 {
		t.Errorf("Inventory count = %d, want 1", inv.GetItemCount())
	}
	if inv.Items[0] != sword1 {
		t.Error("Sword 1 not returned to inventory")
	}
}

// TestInventorySystem_UnequipItem tests unequipping items.
func TestInventorySystem_UnequipItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&StatsComponent{})
	entity.AddComponent(&AttackComponent{})

	world.Update(0.0) // Process entity additions
	comp2, _ := entity.GetComponent("equipment")
	equip := comp2.(*EquipmentComponent)
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)

	// Equip weapon
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	equip.Equip(sword, SlotMainHand)

	// Unequip weapon
	err := system.UnequipItem(entity.ID, SlotMainHand)
	if err != nil {
		t.Fatalf("UnequipItem failed: %v", err)
	}

	// Verify weapon is unequipped
	if equip.GetEquipped(SlotMainHand) != nil {
		t.Error("Weapon still equipped")
	}

	// Verify weapon is in inventory
	if inv.GetItemCount() != 1 {
		t.Errorf("Inventory count = %d, want 1", inv.GetItemCount())
	}
	if inv.Items[0] != sword {
		t.Error("Sword not in inventory")
	}
}

// TestInventorySystem_UnequipFullInventory tests unequipping when inventory is full.
func TestInventorySystem_UnequipFullInventory(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(1, 100.0)) // Only 1 slot
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(&StatsComponent{})
	entity.AddComponent(&AttackComponent{})

	world.Update(0.0) // Process entity additions
	comp2, _ := entity.GetComponent("equipment")
	equip := comp2.(*EquipmentComponent)
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)

	// Fill inventory
	potion := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)
	inv.AddItem(potion)

	// Equip weapon
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	equip.Equip(sword, SlotMainHand)

	// Try to unequip (should fail, inventory full)
	err := system.UnequipItem(entity.ID, SlotMainHand)
	if err == nil {
		t.Error("UnequipItem should fail with full inventory")
	}

	// Verify weapon is still equipped
	if equip.GetEquipped(SlotMainHand) != sword {
		t.Error("Weapon should still be equipped")
	}
}

// TestInventorySystem_UseConsumable tests using consumable items.
func TestInventorySystem_UseConsumable(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(&HealthComponent{Current: 50.0, Max: 100.0})

	world.Update(0.0) // Process entity additions
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)
	comp3, _ := entity.GetComponent("health")
	health := comp3.(*HealthComponent)

	// Add healing potion
	potion := createTestItem("Health Potion", item.TypeConsumable, 0.5, 50, 0, 0)
	potion.ConsumableType = item.ConsumablePotion
	inv.AddItem(potion)

	initialHealth := health.Current

	// Use potion
	err := system.UseConsumable(entity.ID, 0)
	if err != nil {
		t.Fatalf("UseConsumable failed: %v", err)
	}

	// Verify potion was consumed
	if inv.GetItemCount() != 0 {
		t.Error("Potion should be removed from inventory after use")
	}

	// Verify health was restored
	if health.Current <= initialHealth {
		t.Error("Health should increase after using potion")
	}
}

// TestInventorySystem_UseNonConsumable tests attempting to use non-consumable items.
func TestInventorySystem_UseNonConsumable(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Process entity additions
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)

	// Add weapon (non-consumable)
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	inv.AddItem(sword)

	// Try to use weapon (should fail)
	err := system.UseConsumable(entity.ID, 0)
	if err == nil {
		t.Error("UseConsumable should fail for non-consumable items")
	}

	// Verify weapon is still in inventory
	if inv.GetItemCount() != 1 {
		t.Error("Weapon should still be in inventory")
	}
}

// TestInventorySystem_TransferItem tests transferring items between entities.
func TestInventorySystem_TransferItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create two entities with inventories
	entity1 := world.CreateEntity()
	entity1.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Process entity additions
	entity2 := world.CreateEntity()
	entity2.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Process entity additions
	comp1, _ := entity1.GetComponent("inventory")
	inv1 := comp1.(*InventoryComponent)
	comp2, _ := entity2.GetComponent("inventory")
	inv2 := comp2.(*InventoryComponent)

	// Add item to entity1
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	inv1.AddItem(sword)

	// Transfer item to entity2
	err := system.TransferItem(entity1.ID, entity2.ID, 0)
	if err != nil {
		t.Fatalf("TransferItem failed: %v", err)
	}

	// Verify item was transferred
	if inv1.GetItemCount() != 0 {
		t.Error("Item should be removed from entity1")
	}
	if inv2.GetItemCount() != 1 {
		t.Error("Item should be added to entity2")
	}
	if inv2.Items[0] != sword {
		t.Error("Wrong item in entity2 inventory")
	}
}

// TestInventorySystem_TransferFullInventory tests transferring to a full inventory.
func TestInventorySystem_TransferFullInventory(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity1 := world.CreateEntity()
	entity1.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Process entity additions
	entity2 := world.CreateEntity()
	entity2.AddComponent(NewInventoryComponent(1, 10.0)) // Small inventory

	world.Update(0.0) // Process entity additions
	comp1, _ := entity1.GetComponent("inventory")
	inv1 := comp1.(*InventoryComponent)
	comp2, _ := entity2.GetComponent("inventory")
	inv2 := comp2.(*InventoryComponent)

	// Fill entity2's inventory
	potion := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)
	inv2.AddItem(potion)

	// Try to transfer heavy item (should fail due to weight or capacity)
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 20.0, 1.2)
	inv1.AddItem(sword)

	err := system.TransferItem(entity1.ID, entity2.ID, 0)
	if err == nil {
		t.Error("TransferItem should fail when destination is full")
	}

	// Verify item is still in entity1
	if inv1.GetItemCount() != 1 {
		t.Error("Item should still be in entity1")
	}
}

// TestInventorySystem_GetInventoryValue tests calculating total inventory value.
func TestInventorySystem_GetInventoryValue(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	inv := NewInventoryComponent(10, 100.0)
	inv.Gold = 100
	entity.AddComponent(inv)

	world.Update(0.0) // Process entity additions
	// Add items with known values
	inv.AddItem(createTestItem("Sword", item.TypeWeapon, 5.0, 150, 10, 0))
	inv.AddItem(createTestItem("Potion", item.TypeConsumable, 0.5, 25, 0, 0))
	inv.AddItem(createTestItem("Armor", item.TypeArmor, 10.0, 200, 0, 15))

	totalValue, err := system.GetInventoryValue(entity.ID)
	if err != nil {
		t.Fatalf("GetInventoryValue failed: %v", err)
	}

	expectedValue := 100 + 150 + 25 + 200
	if totalValue != expectedValue {
		t.Errorf("Total value = %d, want %d", totalValue, expectedValue)
	}
}

// TestInventorySystem_SortByValue tests sorting inventory by value.
func TestInventorySystem_SortByValue(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	inv := NewInventoryComponent(10, 100.0)
	entity.AddComponent(inv)

	world.Update(0.0) // Process entity additions
	// Add items with different values
	item1 := createTestItem("Cheap", item.TypeConsumable, 0.5, 10, 0, 0)
	item2 := createTestItem("Expensive", item.TypeWeapon, 5.0, 200, 15, 0)
	item3 := createTestItem("Medium", item.TypeArmor, 8.0, 75, 0, 10)

	inv.AddItem(item1)
	inv.AddItem(item2)
	inv.AddItem(item3)

	// Sort by value
	err := system.SortInventoryByValue(entity.ID)
	if err != nil {
		t.Fatalf("SortInventoryByValue failed: %v", err)
	}

	// Verify order (descending by value)
	if inv.Items[0] != item2 {
		t.Error("Most valuable item should be first")
	}
	if inv.Items[1] != item3 {
		t.Error("Medium value item should be second")
	}
	if inv.Items[2] != item1 {
		t.Error("Least valuable item should be last")
	}
}

// TestInventorySystem_SortByWeight tests sorting inventory by weight.
func TestInventorySystem_SortByWeight(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	inv := NewInventoryComponent(10, 100.0)
	entity.AddComponent(inv)

	world.Update(0.0) // Process entity additions
	// Add items with different weights
	item1 := createTestItem("Light", item.TypeConsumable, 0.5, 10, 0, 0)
	item2 := createTestItem("Heavy", item.TypeArmor, 15.0, 100, 0, 20)
	item3 := createTestItem("Medium", item.TypeWeapon, 5.0, 75, 10, 0)

	inv.AddItem(item2)
	inv.AddItem(item1)
	inv.AddItem(item3)

	// Sort by weight
	err := system.SortInventoryByWeight(entity.ID)
	if err != nil {
		t.Fatalf("SortInventoryByWeight failed: %v", err)
	}

	// Verify order (ascending by weight)
	if inv.Items[0] != item1 {
		t.Error("Lightest item should be first")
	}
	if inv.Items[1] != item3 {
		t.Error("Medium weight item should be second")
	}
	if inv.Items[2] != item2 {
		t.Error("Heaviest item should be last")
	}
}

// TestInventorySystem_SortByType tests sorting inventory by item type.
func TestInventorySystem_SortByType(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	inv := NewInventoryComponent(10, 100.0)
	entity.AddComponent(inv)

	world.Update(0.0) // Process entity additions
	// Add items of different types
	weapon := createTestItem("Sword", item.TypeWeapon, 5.0, 100, 10, 0)
	armor := createTestItem("Helmet", item.TypeArmor, 2.0, 50, 0, 8)
	consumable := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)

	// Add in random order
	inv.AddItem(consumable)
	inv.AddItem(armor)
	inv.AddItem(weapon)

	// Sort by type
	err := system.SortInventoryByType(entity.ID)
	if err != nil {
		t.Fatalf("SortInventoryByType failed: %v", err)
	}

	// Verify order (by ItemType enum value: Weapon < Armor < Consumable)
	if inv.Items[0] != weapon {
		t.Error("Weapon should be first")
	}
	if inv.Items[1] != armor {
		t.Error("Armor should be second")
	}
	if inv.Items[2] != consumable {
		t.Error("Consumable should be last")
	}
}

// TestInventorySystem_InvalidEntity tests operations on non-existent entities.
func TestInventorySystem_InvalidEntity(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	invalidID := uint64(9999)
	testItem := createTestItem("Item", item.TypeWeapon, 5.0, 100, 10, 0)

	// Test all operations with invalid entity
	_, err := system.AddItemToInventory(invalidID, testItem)
	if err == nil {
		t.Error("AddItemToInventory should fail with invalid entity")
	}

	_, err = system.RemoveItemFromInventory(invalidID, 0)
	if err == nil {
		t.Error("RemoveItemFromInventory should fail with invalid entity")
	}

	err = system.EquipItem(invalidID, 0)
	if err == nil {
		t.Error("EquipItem should fail with invalid entity")
	}

	err = system.UnequipItem(invalidID, SlotMainHand)
	if err == nil {
		t.Error("UnequipItem should fail with invalid entity")
	}
}

// TestInventorySystem_MissingComponents tests operations on entities without required components.
func TestInventorySystem_MissingComponents(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	// Create entity without components
	entity := world.CreateEntity()
	testItem := createTestItem("Item", item.TypeWeapon, 5.0, 100, 10, 0)

	_, err := system.AddItemToInventory(entity.ID, testItem)
	if err == nil {
		t.Error("AddItemToInventory should fail without inventory component")
	}
}

// TestInventorySystem_DropItem tests dropping items from inventory.
func TestInventorySystem_DropItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Process entity additions
	comp, _ := entity.GetComponent("inventory")
	inv := comp.(*InventoryComponent)
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	inv.AddItem(sword)

	// Drop item
	err := system.DropItem(entity.ID, 0)
	if err != nil {
		t.Fatalf("DropItem failed: %v", err)
	}

	// Verify item was removed
	if inv.GetItemCount() != 0 {
		t.Error("Item should be removed from inventory after dropping")
	}
}
