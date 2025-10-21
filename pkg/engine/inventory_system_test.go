package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// Helper function to get inventory component safely
func getInventory(entity *Entity) *InventoryComponent {
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, _ := comp.(*InventoryComponent)
	return inv
}

// Helper function to get equipment component safely
func getEquipment(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok {
		return nil
	}
	equip, _ := comp.(*EquipmentComponent)
	return equip
}

// Helper function to get health component safely
func getHealth(entity *Entity) *HealthComponent {
	comp, ok := entity.GetComponent("health")
	if !ok {
		return nil
	}
	health, _ := comp.(*HealthComponent)
	return health
}

// Helper function to get stats component safely
func getStats(entity *Entity) *StatsComponent {
	comp, ok := entity.GetComponent("stats")
	if !ok {
		return nil
	}
	stats, _ := comp.(*StatsComponent)
	return stats
}

// Helper function to get attack component safely
func getAttack(entity *Entity) *AttackComponent {
	comp, ok := entity.GetComponent("attack")
	if !ok {
		return nil
	}
	attack, _ := comp.(*AttackComponent)
	return attack
}

func TestInventorySystem_AddItemToInventory(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	// Create entity with inventory
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	
	testItem := &item.Item{
		Name: "Test Sword",
		Stats: item.Stats{Weight: 5.0},
	}
	
	// Test callback
	callbackCalled := false
	system.OnItemAdded = func(entityID uint64, itm *item.Item) {
		callbackCalled = true
		if itm.Name != "Test Sword" {
			t.Errorf("Callback received wrong item: %s", itm.Name)
		}
	}
	
	success := system.AddItemToInventory(world, entity.ID, testItem)
	if !success {
		t.Error("Failed to add item to inventory")
	}
	
	if !callbackCalled {
		t.Error("OnItemAdded callback was not called")
	}
	
	inv := getInventory(entity)
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(inv.Items))
	}
}

func TestInventorySystem_RemoveItemFromInventory(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	inv := NewInventoryComponent(10, 100.0)
	entity.AddComponent(inv)
	
	testItem := &item.Item{Name: "Test Item"}
	inv.AddItem(testItem)
	
	// Test callback
	callbackCalled := false
	system.OnItemRemoved = func(entityID uint64, itm *item.Item) {
		callbackCalled = true
	}
	
	removed := system.RemoveItemFromInventory(world, entity.ID, 0)
	if removed == nil {
		t.Fatal("RemoveItemFromInventory returned nil")
	}
	
	if removed.Name != "Test Item" {
		t.Errorf("Expected 'Test Item', got '%s'", removed.Name)
	}
	
	if !callbackCalled {
		t.Error("OnItemRemoved callback was not called")
	}
	
	if len(inv.Items) != 0 {
		t.Errorf("Expected 0 items in inventory, got %d", len(inv.Items))
	}
}

func TestInventorySystem_EquipItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(NewStatsComponent())
	
	sword := &item.Item{
		Name: "Iron Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{Damage: 15},
	}
	
	inv := getInventory(entity)
	inv.AddItem(sword)
	
	// Test callback
	callbackCalled := false
	system.OnItemEquipped = func(entityID uint64, itm *item.Item, slot EquipmentSlot) {
		callbackCalled = true
		if slot != SlotWeapon {
			t.Errorf("Expected SlotWeapon, got %v", slot)
		}
	}
	
	success := system.EquipItem(world, entity.ID, 0)
	if !success {
		t.Fatal("Failed to equip item")
	}
	
	if !callbackCalled {
		t.Error("OnItemEquipped callback was not called")
	}
	
	// Item should be removed from inventory
	if len(inv.Items) != 0 {
		t.Errorf("Expected 0 items in inventory, got %d", len(inv.Items))
	}
	
	// Item should be equipped
	equip := getEquipment(entity)
	equipped := equip.GetEquipped(SlotWeapon)
	if equipped == nil {
		t.Fatal("Item not equipped")
	}
	
	if equipped.Name != "Iron Sword" {
		t.Errorf("Expected 'Iron Sword' equipped, got '%s'", equipped.Name)
	}
}

func TestInventorySystem_EquipItem_SwapWeapon(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(NewStatsComponent())
	
	// Equip first sword
	sword1 := &item.Item{Name: "Iron Sword", Type: item.TypeWeapon, Stats: item.Stats{Damage: 10}}
	equip := getEquipment(entity)
	equip.Equip(sword1)
	
	// Add second sword to inventory
	sword2 := &item.Item{Name: "Steel Sword", Type: item.TypeWeapon, Stats: item.Stats{Damage: 15}}
	inv := getInventory(entity)
	inv.AddItem(sword2)
	
	// Equip second sword (should swap)
	success := system.EquipItem(world, entity.ID, 0)
	if !success {
		t.Fatal("Failed to swap weapon")
	}
	
	// Steel Sword should be equipped
	equipped := equip.GetEquipped(SlotWeapon)
	if equipped.Name != "Steel Sword" {
		t.Errorf("Expected 'Steel Sword' equipped, got '%s'", equipped.Name)
	}
	
	// Iron Sword should be back in inventory
	if len(inv.Items) != 1 {
		t.Fatalf("Expected 1 item in inventory, got %d", len(inv.Items))
	}
	
	if inv.Items[0].Name != "Iron Sword" {
		t.Errorf("Expected 'Iron Sword' in inventory, got '%s'", inv.Items[0].Name)
	}
}

func TestInventorySystem_UnequipItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(NewStatsComponent())
	
	helmet := &item.Item{
		Name: "Iron Helmet",
		Type: item.TypeArmor,
		ArmorType: item.ArmorHelmet,
		Stats: item.Stats{Defense: 10, Weight: 3.0},
	}
	
	equip := getEquipment(entity)
	equip.Equip(helmet)
	
	// Test callback
	callbackCalled := false
	system.OnItemUnequipped = func(entityID uint64, itm *item.Item, slot EquipmentSlot) {
		callbackCalled = true
	}
	
	success := system.UnequipItem(world, entity.ID, SlotHead)
	if !success {
		t.Fatal("Failed to unequip item")
	}
	
	if !callbackCalled {
		t.Error("OnItemUnequipped callback was not called")
	}
	
	// Slot should be empty
	if !equip.IsSlotEmpty(SlotHead) {
		t.Error("Slot should be empty after unequip")
	}
	
	// Item should be in inventory
	inv := getInventory(entity)
	if len(inv.Items) != 1 {
		t.Fatalf("Expected 1 item in inventory, got %d", len(inv.Items))
	}
	
	if inv.Items[0].Name != "Iron Helmet" {
		t.Errorf("Expected 'Iron Helmet' in inventory, got '%s'", inv.Items[0].Name)
	}
}

func TestInventorySystem_UnequipItem_InventoryFull(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(1, 100.0)) // Only 1 slot
	entity.AddComponent(NewEquipmentComponent())
	
	// Fill inventory
	inv := getInventory(entity)
	inv.AddItem(&item.Item{Name: "Potion"})
	
	// Equip helmet
	helmet := &item.Item{
		Name: "Helmet",
		Type: item.TypeArmor,
		ArmorType: item.ArmorHelmet,
	}
	equip := getEquipment(entity)
	equip.Equip(helmet)
	
	// Try to unequip (should fail, inventory full)
	success := system.UnequipItem(world, entity.ID, SlotHead)
	if success {
		t.Error("Should not be able to unequip when inventory is full")
	}
	
	// Helmet should still be equipped
	if equip.IsSlotEmpty(SlotHead) {
		t.Error("Helmet should still be equipped")
	}
}

func TestInventorySystem_UseConsumable(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	entity.AddComponent(&HealthComponent{Current: 50.0, Max: 100.0})
	
	potion := &item.Item{
		Name: "Health Potion",
		Type: item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
		Stats: item.Stats{Value: 100}, // Healing = value / 10 = 10
	}
	
	inv := getInventory(entity)
	inv.AddItem(potion)
	
	// Use the potion
	success := system.UseConsumable(world, entity.ID, 0)
	if !success {
		t.Fatal("Failed to use consumable")
	}
	
	// Item should be removed from inventory
	if len(inv.Items) != 0 {
		t.Errorf("Expected 0 items in inventory, got %d", len(inv.Items))
	}
	
	// Health should be increased
	health := getHealth(entity)
	expectedHealth := 50.0 + 10.0 // 50 + (100 / 10)
	if health.Current != expectedHealth {
		t.Errorf("Expected health %f, got %f", expectedHealth, health.Current)
	}
}

func TestInventorySystem_UseConsumable_NotConsumable(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	
	sword := &item.Item{
		Name: "Sword",
		Type: item.TypeWeapon,
	}
	
	inv := getInventory(entity)
	inv.AddItem(sword)
	
	// Try to use weapon as consumable (should fail)
	success := system.UseConsumable(world, entity.ID, 0)
	if success {
		t.Error("Should not be able to use non-consumable item")
	}
	
	// Item should still be in inventory
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(inv.Items))
	}
}

func TestInventorySystem_UpdateStatsFromEquipment(t *testing.T) {
	entity := NewWorld().CreateEntity()
	entity.AddComponent(NewEquipmentComponent())
	entity.AddComponent(NewStatsComponent())
	entity.AddComponent(&AttackComponent{})
	
	system := NewInventorySystem()
	
	// Equip weapon with 20 damage
	weapon := &item.Item{
		Type: item.TypeWeapon,
		Stats: item.Stats{Damage: 20},
	}
	equip := getEquipment(entity)
	equip.Equip(weapon)
	
	// Equip armor with 15 defense
	armor := &item.Item{
		Type: item.TypeArmor,
		ArmorType: item.ArmorChest,
		Stats: item.Stats{Defense: 15},
	}
	equip.Equip(armor)
	
	// Update stats
	system.UpdateStatsFromEquipment(entity)
	
	stats := getStats(entity)
	
	// Base attack is 10, weapon adds 20
	expectedAttack := 10.0 + 20.0
	if stats.Attack != expectedAttack {
		t.Errorf("Expected attack %f, got %f", expectedAttack, stats.Attack)
	}
	
	// Base defense is 5, armor adds 15
	expectedDefense := 5.0 + 15.0
	if stats.Defense != expectedDefense {
		t.Errorf("Expected defense %f, got %f", expectedDefense, stats.Defense)
	}
	
	// Attack component should be updated
	attack := getAttack(entity)
	if attack.Damage != expectedAttack {
		t.Errorf("Expected attack damage %f, got %f", expectedAttack, attack.Damage)
	}
}

func TestInventorySystem_GetInventoryInfo(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 50.0))
	
	inv := getInventory(entity)
	inv.AddItem(&item.Item{Stats: item.Stats{Weight: 10.0}})
	inv.AddItem(&item.Item{Stats: item.Stats{Weight: 15.0}})
	
	count, maxCap, weight, maxWeight, ok := system.GetInventoryInfo(world, entity.ID)
	
	if !ok {
		t.Fatal("GetInventoryInfo failed")
	}
	
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	
	if maxCap != 10 {
		t.Errorf("Expected maxCapacity 10, got %d", maxCap)
	}
	
	if weight != 25.0 {
		t.Errorf("Expected weight 25.0, got %f", weight)
	}
	
	if maxWeight != 50.0 {
		t.Errorf("Expected maxWeight 50.0, got %f", maxWeight)
	}
}

func TestInventorySystem_GetEquipmentInfo(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewEquipmentComponent())
	
	equip := getEquipment(entity)
	equip.Equip(&item.Item{
		Type: item.TypeWeapon,
		Stats: item.Stats{Damage: 25},
	})
	equip.Equip(&item.Item{
		Type: item.TypeArmor,
		ArmorType: item.ArmorChest,
		Stats: item.Stats{Defense: 30},
	})
	
	damage, defense, ok := system.GetEquipmentInfo(world, entity.ID)
	
	if !ok {
		t.Fatal("GetEquipmentInfo failed")
	}
	
	if damage != 25 {
		t.Errorf("Expected damage 25, got %d", damage)
	}
	
	if defense != 30 {
		t.Errorf("Expected defense 30, got %d", defense)
	}
}

func TestInventorySystem_DropItem(t *testing.T) {
	world := NewWorld()
	system := NewInventorySystem()
	
	entity := world.CreateEntity()
	world.Update(0)
	entity.AddComponent(NewInventoryComponent(10, 100.0))
	
	testItem := &item.Item{Name: "Dropped Item"}
	inv := getInventory(entity)
	inv.AddItem(testItem)
	
	dropped := system.DropItem(world, entity.ID, 0)
	
	if dropped == nil {
		t.Fatal("DropItem returned nil")
	}
	
	if dropped.Name != "Dropped Item" {
		t.Errorf("Expected 'Dropped Item', got '%s'", dropped.Name)
	}
	
	if len(inv.Items) != 0 {
		t.Errorf("Expected 0 items in inventory, got %d", len(inv.Items))
	}
}
