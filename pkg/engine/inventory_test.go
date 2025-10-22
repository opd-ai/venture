package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// Helper function to create a test item
func createTestItem(name string, itemType item.ItemType, weight float64, value, damage, defense int) *item.Item {
	return &item.Item{
		Name: name,
		Type: itemType,
		Stats: item.Stats{
			Weight:  weight,
			Value:   value,
			Damage:  damage,
			Defense: defense,
		},
	}
}

// Helper function to create a test weapon
func createTestWeapon(name string, weaponType item.WeaponType, damage int, weight, speed float64) *item.Item {
	return &item.Item{
		Name:       name,
		Type:       item.TypeWeapon,
		WeaponType: weaponType,
		Stats: item.Stats{
			Damage:      damage,
			Weight:      weight,
			AttackSpeed: speed,
			Value:       damage * 10,
		},
	}
}

// Helper function to create test armor
func createTestArmor(name string, armorType item.ArmorType, defense int, weight float64) *item.Item {
	return &item.Item{
		Name:      name,
		Type:      item.TypeArmor,
		ArmorType: armorType,
		Stats: item.Stats{
			Defense: defense,
			Weight:  weight,
			Value:   defense * 10,
		},
	}
}

// TestInventoryComponent_Basic tests basic inventory operations.
func TestInventoryComponent_Basic(t *testing.T) {
	tests := []struct {
		name      string
		maxItems  int
		maxWeight float64
	}{
		{"small inventory", 10, 50.0},
		{"medium inventory", 20, 100.0},
		{"large inventory", 50, 200.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventoryComponent(tt.maxItems, tt.maxWeight)

			if inv.MaxItems != tt.maxItems {
				t.Errorf("MaxItems = %d, want %d", inv.MaxItems, tt.maxItems)
			}
			if inv.MaxWeight != tt.maxWeight {
				t.Errorf("MaxWeight = %f, want %f", inv.MaxWeight, tt.maxWeight)
			}
			if inv.GetItemCount() != 0 {
				t.Errorf("Initial item count = %d, want 0", inv.GetItemCount())
			}
		})
	}
}

// TestInventoryComponent_AddRemove tests adding and removing items.
func TestInventoryComponent_AddRemove(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)

	// Add items
	item1 := createTestItem("Sword", item.TypeWeapon, 5.0, 100, 10, 0)
	item2 := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)

	if !inv.AddItem(item1) {
		t.Error("Failed to add item1")
	}
	if !inv.AddItem(item2) {
		t.Error("Failed to add item2")
	}

	if inv.GetItemCount() != 2 {
		t.Errorf("Item count = %d, want 2", inv.GetItemCount())
	}

	// Remove item by index
	removed := inv.RemoveItem(0)
	if removed != item1 {
		t.Error("Removed wrong item")
	}
	if inv.GetItemCount() != 1 {
		t.Errorf("Item count after removal = %d, want 1", inv.GetItemCount())
	}

	// Remove by reference
	if !inv.RemoveItemByReference(item2) {
		t.Error("Failed to remove item by reference")
	}
	if inv.GetItemCount() != 0 {
		t.Errorf("Item count after removal = %d, want 0", inv.GetItemCount())
	}
}

// TestInventoryComponent_WeightLimit tests weight-based capacity.
func TestInventoryComponent_WeightLimit(t *testing.T) {
	inv := NewInventoryComponent(10, 20.0) // 20kg weight limit

	// Add items until weight limit is reached
	heavyItem := createTestItem("Heavy Armor", item.TypeArmor, 15.0, 200, 0, 20)
	mediumItem := createTestItem("Medium Shield", item.TypeArmor, 8.0, 100, 0, 10)

	if !inv.AddItem(heavyItem) {
		t.Error("Failed to add heavy item")
	}

	// This should fail due to weight limit
	if inv.AddItem(mediumItem) {
		t.Error("Should not be able to add item exceeding weight limit")
	}

	// Weight should be 15.0
	if weight := inv.GetCurrentWeight(); weight != 15.0 {
		t.Errorf("Current weight = %f, want 15.0", weight)
	}
}

// TestInventoryComponent_ItemLimit tests item count limit.
func TestInventoryComponent_ItemLimit(t *testing.T) {
	inv := NewInventoryComponent(3, 100.0) // Only 3 items allowed

	lightItem := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)

	// Add 3 items
	for i := 0; i < 3; i++ {
		if !inv.AddItem(lightItem) {
			t.Errorf("Failed to add item %d", i)
		}
	}

	// Fourth item should fail
	if inv.AddItem(lightItem) {
		t.Error("Should not be able to add 4th item with 3-item limit")
	}

	if !inv.IsFull() {
		t.Error("Inventory should be full")
	}
}

// TestInventoryComponent_FindItem tests finding items by name.
func TestInventoryComponent_FindItem(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)

	item1 := createTestItem("Sword", item.TypeWeapon, 5.0, 100, 10, 0)
	item2 := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)

	inv.AddItem(item1)
	inv.AddItem(item2)

	found := inv.FindItem("Sword")
	if found != item1 {
		t.Error("FindItem returned wrong item")
	}

	notFound := inv.FindItem("Nonexistent")
	if notFound != nil {
		t.Error("FindItem should return nil for nonexistent item")
	}
}

// TestInventoryComponent_Clear tests clearing the inventory.
func TestInventoryComponent_Clear(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)

	inv.AddItem(createTestItem("Item1", item.TypeWeapon, 5.0, 100, 10, 0))
	inv.AddItem(createTestItem("Item2", item.TypeArmor, 10.0, 200, 0, 15))

	inv.Clear()

	if inv.GetItemCount() != 0 {
		t.Errorf("Item count after clear = %d, want 0", inv.GetItemCount())
	}
	if inv.GetCurrentWeight() != 0.0 {
		t.Errorf("Weight after clear = %f, want 0.0", inv.GetCurrentWeight())
	}
}

// TestEquipmentSlot_String tests equipment slot string representation.
func TestEquipmentSlot_String(t *testing.T) {
	tests := []struct {
		slot EquipmentSlot
		want string
	}{
		{SlotMainHand, "main_hand"},
		{SlotOffHand, "off_hand"},
		{SlotHead, "head"},
		{SlotChest, "chest"},
		{SlotLegs, "legs"},
		{SlotBoots, "boots"},
		{SlotGloves, "gloves"},
		{SlotAccessory1, "accessory_1"},
		{SlotAccessory2, "accessory_2"},
		{SlotAccessory3, "accessory_3"},
		{EquipmentSlot(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.slot.String(); got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestEquipmentComponent_Basic tests basic equipment operations.
func TestEquipmentComponent_Basic(t *testing.T) {
	equip := NewEquipmentComponent()

	if len(equip.Slots) != 0 {
		t.Errorf("Initial slots count = %d, want 0", len(equip.Slots))
	}

	if !equip.StatsDirty {
		t.Error("Stats should be dirty initially")
	}
}

// TestEquipmentComponent_EquipWeapon tests equipping weapons.
func TestEquipmentComponent_EquipWeapon(t *testing.T) {
	equip := NewEquipmentComponent()

	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)

	// Equip sword in main hand
	previous := equip.Equip(sword, SlotMainHand)
	if previous != nil {
		t.Error("Should have no previous item")
	}

	equipped := equip.GetEquipped(SlotMainHand)
	if equipped != sword {
		t.Error("Sword not equipped properly")
	}

	// Equip another weapon
	axe := createTestWeapon("Axe", item.WeaponAxe, 20, 8.0, 0.9)
	previous = equip.Equip(axe, SlotMainHand)
	if previous != sword {
		t.Error("Should return previously equipped sword")
	}

	equipped = equip.GetEquipped(SlotMainHand)
	if equipped != axe {
		t.Error("Axe not equipped properly")
	}
}

// TestEquipmentComponent_EquipArmor tests equipping armor pieces.
func TestEquipmentComponent_EquipArmor(t *testing.T) {
	equip := NewEquipmentComponent()

	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)
	chest := createTestArmor("Chestplate", item.ArmorChest, 20, 10.0)
	legs := createTestArmor("Leggings", item.ArmorLegs, 15, 6.0)

	// Equip armor pieces
	equip.Equip(helmet, SlotHead)
	equip.Equip(chest, SlotChest)
	equip.Equip(legs, SlotLegs)

	if equip.GetEquipped(SlotHead) != helmet {
		t.Error("Helmet not equipped")
	}
	if equip.GetEquipped(SlotChest) != chest {
		t.Error("Chest armor not equipped")
	}
	if equip.GetEquipped(SlotLegs) != legs {
		t.Error("Leg armor not equipped")
	}
}

// TestEquipmentComponent_CanEquip tests item-slot compatibility.
func TestEquipmentComponent_CanEquip(t *testing.T) {
	equip := NewEquipmentComponent()

	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)
	potion := createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0)

	tests := []struct {
		name string
		item *item.Item
		slot EquipmentSlot
		want bool
	}{
		{"sword in main hand", sword, SlotMainHand, true},
		{"sword in off hand", sword, SlotOffHand, true},
		{"sword in head", sword, SlotHead, false},
		{"helmet in head", helmet, SlotHead, true},
		{"helmet in chest", helmet, SlotChest, false},
		{"potion in main hand", potion, SlotMainHand, false},
		{"potion in head", potion, SlotHead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equip.CanEquip(tt.item, tt.slot); got != tt.want {
				t.Errorf("CanEquip() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEquipmentComponent_GetSlotForItem tests automatic slot detection.
func TestEquipmentComponent_GetSlotForItem(t *testing.T) {
	equip := NewEquipmentComponent()

	tests := []struct {
		name     string
		item     *item.Item
		wantSlot EquipmentSlot
		wantOK   bool
	}{
		{
			"weapon",
			createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2),
			SlotMainHand,
			true,
		},
		{
			"helmet",
			createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0),
			SlotHead,
			true,
		},
		{
			"chest armor",
			createTestArmor("Chest", item.ArmorChest, 20, 10.0),
			SlotChest,
			true,
		},
		{
			"shield",
			createTestArmor("Shield", item.ArmorShield, 15, 5.0),
			SlotOffHand,
			true,
		},
		{
			"consumable",
			createTestItem("Potion", item.TypeConsumable, 0.5, 20, 0, 0),
			0,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlot, gotOK := equip.GetSlotForItem(tt.item)
			if gotOK != tt.wantOK {
				t.Errorf("GetSlotForItem() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if tt.wantOK && gotSlot != tt.wantSlot {
				t.Errorf("GetSlotForItem() slot = %v, want %v", gotSlot, tt.wantSlot)
			}
		})
	}
}

// TestEquipmentComponent_UnequipAll tests removing all equipment.
func TestEquipmentComponent_UnequipAll(t *testing.T) {
	equip := NewEquipmentComponent()

	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)
	chest := createTestArmor("Chest", item.ArmorChest, 20, 10.0)

	equip.Equip(sword, SlotMainHand)
	equip.Equip(helmet, SlotHead)
	equip.Equip(chest, SlotChest)

	items := equip.UnequipAll()

	if len(items) != 3 {
		t.Errorf("UnequipAll returned %d items, want 3", len(items))
	}

	if len(equip.Slots) != 0 {
		t.Errorf("Slots count after UnequipAll = %d, want 0", len(equip.Slots))
	}
}

// TestEquipmentComponent_RecalculateStats tests stat calculation from equipment.
func TestEquipmentComponent_RecalculateStats(t *testing.T) {
	equip := NewEquipmentComponent()

	// Equip items with known stats
	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)
	chest := createTestArmor("Chest", item.ArmorChest, 20, 10.0)

	equip.Equip(sword, SlotMainHand)
	equip.Equip(helmet, SlotHead)
	equip.Equip(chest, SlotChest)

	stats := equip.GetStats()

	// Total damage should be 15 from sword
	if stats.Damage != 15 {
		t.Errorf("Total damage = %d, want 15", stats.Damage)
	}

	// Total defense should be 30 (10 + 20)
	if stats.Defense != 30 {
		t.Errorf("Total defense = %d, want 30", stats.Defense)
	}

	// Attack speed should be from weapon (1.2)
	if stats.AttackSpeed != 1.2 {
		t.Errorf("Attack speed = %f, want 1.2", stats.AttackSpeed)
	}

	// Total weight should be 17.0 (5.0 + 2.0 + 10.0)
	if stats.Weight != 17.0 {
		t.Errorf("Total weight = %f, want 17.0", stats.Weight)
	}
}

// TestEquipmentComponent_GetTotalDefense tests defense calculation.
func TestEquipmentComponent_GetTotalDefense(t *testing.T) {
	equip := NewEquipmentComponent()

	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)
	chest := createTestArmor("Chest", item.ArmorChest, 25, 10.0)
	legs := createTestArmor("Legs", item.ArmorLegs, 15, 6.0)

	equip.Equip(helmet, SlotHead)
	equip.Equip(chest, SlotChest)
	equip.Equip(legs, SlotLegs)

	totalDefense := equip.GetTotalDefense()
	expectedDefense := 10 + 25 + 15

	if totalDefense != expectedDefense {
		t.Errorf("GetTotalDefense() = %d, want %d", totalDefense, expectedDefense)
	}
}

// TestEquipmentComponent_GetWeaponDamage tests weapon damage retrieval.
func TestEquipmentComponent_GetWeaponDamage(t *testing.T) {
	equip := NewEquipmentComponent()

	// No weapon equipped
	if damage := equip.GetWeaponDamage(); damage != 0 {
		t.Errorf("Damage with no weapon = %d, want 0", damage)
	}

	// Equip weapon
	sword := createTestWeapon("Sword", item.WeaponSword, 20, 5.0, 1.2)
	equip.Equip(sword, SlotMainHand)

	if damage := equip.GetWeaponDamage(); damage != 20 {
		t.Errorf("Damage with sword = %d, want 20", damage)
	}
}

// TestEquipmentComponent_GetWeaponSpeed tests weapon speed retrieval.
func TestEquipmentComponent_GetWeaponSpeed(t *testing.T) {
	equip := NewEquipmentComponent()

	// No weapon equipped (should return default)
	if speed := equip.GetWeaponSpeed(); speed != 1.0 {
		t.Errorf("Speed with no weapon = %f, want 1.0", speed)
	}

	// Equip fast weapon
	dagger := createTestWeapon("Dagger", item.WeaponDagger, 8, 1.0, 1.8)
	equip.Equip(dagger, SlotMainHand)

	if speed := equip.GetWeaponSpeed(); speed != 1.8 {
		t.Errorf("Speed with dagger = %f, want 1.8", speed)
	}
}

// TestEquipmentComponent_IsEquipped tests checking if an item is equipped.
func TestEquipmentComponent_IsEquipped(t *testing.T) {
	equip := NewEquipmentComponent()

	sword := createTestWeapon("Sword", item.WeaponSword, 15, 5.0, 1.2)
	helmet := createTestArmor("Helmet", item.ArmorHelmet, 10, 2.0)

	// Not equipped yet
	if equip.IsEquipped(sword) {
		t.Error("Sword should not be equipped yet")
	}

	// Equip sword
	equip.Equip(sword, SlotMainHand)
	if !equip.IsEquipped(sword) {
		t.Error("Sword should be equipped")
	}

	// Helmet not equipped
	if equip.IsEquipped(helmet) {
		t.Error("Helmet should not be equipped")
	}
}
