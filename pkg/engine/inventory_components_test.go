package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestInventoryComponent_NewInventoryComponent(t *testing.T) {
	inv := NewInventoryComponent(20, 100.0)
	
	if inv == nil {
		t.Fatal("NewInventoryComponent returned nil")
	}
	
	if inv.MaxCapacity != 20 {
		t.Errorf("Expected MaxCapacity=20, got %d", inv.MaxCapacity)
	}
	
	if inv.MaxWeight != 100.0 {
		t.Errorf("Expected MaxWeight=100.0, got %f", inv.MaxWeight)
	}
	
	if len(inv.Items) != 0 {
		t.Errorf("Expected empty inventory, got %d items", len(inv.Items))
	}
}

func TestInventoryComponent_AddItem(t *testing.T) {
	inv := NewInventoryComponent(3, 50.0)
	
	// Create test items
	item1 := &item.Item{
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{Weight: 5.0},
	}
	item2 := &item.Item{
		Name: "Heavy Armor",
		Type: item.TypeArmor,
		Stats: item.Stats{Weight: 30.0},
	}
	
	// Add first item
	if !inv.AddItem(item1) {
		t.Error("Failed to add first item")
	}
	
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(inv.Items))
	}
	
	// Add second item
	if !inv.AddItem(item2) {
		t.Error("Failed to add second item")
	}
	
	if len(inv.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(inv.Items))
	}
}

func TestInventoryComponent_CapacityLimit(t *testing.T) {
	inv := NewInventoryComponent(2, 100.0)
	
	item1 := &item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}}
	item2 := &item.Item{Name: "Item2", Stats: item.Stats{Weight: 5.0}}
	item3 := &item.Item{Name: "Item3", Stats: item.Stats{Weight: 5.0}}
	
	inv.AddItem(item1)
	inv.AddItem(item2)
	
	// Inventory should be full
	if !inv.IsFull() {
		t.Error("Inventory should be full")
	}
	
	// Should not be able to add third item
	if inv.AddItem(item3) {
		t.Error("Should not be able to add item to full inventory")
	}
}

func TestInventoryComponent_WeightLimit(t *testing.T) {
	inv := NewInventoryComponent(10, 50.0)
	
	item1 := &item.Item{Name: "Heavy1", Stats: item.Stats{Weight: 30.0}}
	item2 := &item.Item{Name: "Heavy2", Stats: item.Stats{Weight: 30.0}}
	
	inv.AddItem(item1)
	
	// Should not be able to add item that would exceed weight limit
	if inv.AddItem(item2) {
		t.Error("Should not be able to add item that exceeds weight limit")
	}
	
	if inv.IsOverweight() {
		t.Error("Inventory should not be overweight")
	}
}

func TestInventoryComponent_RemoveItem(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)
	
	item1 := &item.Item{Name: "Sword", Stats: item.Stats{Weight: 5.0}}
	item2 := &item.Item{Name: "Shield", Stats: item.Stats{Weight: 8.0}}
	
	inv.AddItem(item1)
	inv.AddItem(item2)
	
	// Remove first item
	removed := inv.RemoveItem(0)
	if removed == nil {
		t.Fatal("RemoveItem returned nil")
	}
	
	if removed.Name != "Sword" {
		t.Errorf("Expected to remove 'Sword', got '%s'", removed.Name)
	}
	
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item remaining, got %d", len(inv.Items))
	}
	
	// Remaining item should be Shield
	if inv.Items[0].Name != "Shield" {
		t.Errorf("Expected 'Shield' to remain, got '%s'", inv.Items[0].Name)
	}
}

func TestInventoryComponent_GetCurrentWeight(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)
	
	inv.AddItem(&item.Item{Stats: item.Stats{Weight: 5.0}})
	inv.AddItem(&item.Item{Stats: item.Stats{Weight: 10.0}})
	inv.AddItem(&item.Item{Stats: item.Stats{Weight: 7.5}})
	
	expected := 22.5
	weight := inv.GetCurrentWeight()
	
	if weight != expected {
		t.Errorf("Expected weight %f, got %f", expected, weight)
	}
}

func TestInventoryComponent_FindItem(t *testing.T) {
	inv := NewInventoryComponent(10, 100.0)
	
	sword := &item.Item{Name: "Iron Sword", Type: item.TypeWeapon}
	potion := &item.Item{Name: "Health Potion", Type: item.TypeConsumable}
	
	inv.AddItem(sword)
	inv.AddItem(potion)
	
	// Find weapon
	idx := inv.FindItem(func(i *item.Item) bool {
		return i.Type == item.TypeWeapon
	})
	
	if idx != 0 {
		t.Errorf("Expected to find weapon at index 0, got %d", idx)
	}
	
	// Find consumable
	idx = inv.FindItem(func(i *item.Item) bool {
		return i.Type == item.TypeConsumable
	})
	
	if idx != 1 {
		t.Errorf("Expected to find consumable at index 1, got %d", idx)
	}
	
	// Find non-existent armor
	idx = inv.FindItem(func(i *item.Item) bool {
		return i.Type == item.TypeArmor
	})
	
	if idx != -1 {
		t.Errorf("Expected -1 for non-existent item, got %d", idx)
	}
}

func TestEquipmentComponent_NewEquipmentComponent(t *testing.T) {
	equip := NewEquipmentComponent()
	
	if equip == nil {
		t.Fatal("NewEquipmentComponent returned nil")
	}
	
	if equip.Slots == nil {
		t.Fatal("Slots map is nil")
	}
	
	if len(equip.Slots) != 0 {
		t.Errorf("Expected empty slots, got %d", len(equip.Slots))
	}
}

func TestEquipmentComponent_Equip(t *testing.T) {
	equip := NewEquipmentComponent()
	
	sword := &item.Item{
		Name: "Iron Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{Damage: 10},
	}
	
	previous := equip.Equip(sword)
	
	if previous != nil {
		t.Error("Expected no previous item")
	}
	
	equipped := equip.GetEquipped(SlotWeapon)
	if equipped == nil {
		t.Fatal("Item not equipped")
	}
	
	if equipped.Name != "Iron Sword" {
		t.Errorf("Expected 'Iron Sword', got '%s'", equipped.Name)
	}
}

func TestEquipmentComponent_EquipReplacement(t *testing.T) {
	equip := NewEquipmentComponent()
	
	sword1 := &item.Item{Name: "Iron Sword", Type: item.TypeWeapon}
	sword2 := &item.Item{Name: "Steel Sword", Type: item.TypeWeapon}
	
	equip.Equip(sword1)
	previous := equip.Equip(sword2)
	
	if previous == nil {
		t.Fatal("Expected previous item")
	}
	
	if previous.Name != "Iron Sword" {
		t.Errorf("Expected 'Iron Sword' as previous, got '%s'", previous.Name)
	}
	
	equipped := equip.GetEquipped(SlotWeapon)
	if equipped.Name != "Steel Sword" {
		t.Errorf("Expected 'Steel Sword' equipped, got '%s'", equipped.Name)
	}
}

func TestEquipmentComponent_Unequip(t *testing.T) {
	equip := NewEquipmentComponent()
	
	helmet := &item.Item{
		Name: "Iron Helmet",
		Type: item.TypeArmor,
		ArmorType: item.ArmorHelmet,
	}
	
	equip.Equip(helmet)
	
	unequipped := equip.Unequip(SlotHead)
	if unequipped == nil {
		t.Fatal("Unequip returned nil")
	}
	
	if unequipped.Name != "Iron Helmet" {
		t.Errorf("Expected 'Iron Helmet', got '%s'", unequipped.Name)
	}
	
	if !equip.IsSlotEmpty(SlotHead) {
		t.Error("Slot should be empty after unequip")
	}
}

func TestEquipmentComponent_GetTotalDamage(t *testing.T) {
	equip := NewEquipmentComponent()
	
	sword := &item.Item{
		Type: item.TypeWeapon,
		Stats: item.Stats{Damage: 15},
	}
	
	equip.Equip(sword)
	
	totalDamage := equip.GetTotalDamage()
	if totalDamage != 15 {
		t.Errorf("Expected total damage 15, got %d", totalDamage)
	}
}

func TestEquipmentComponent_GetTotalDefense(t *testing.T) {
	equip := NewEquipmentComponent()
	
	chest := &item.Item{
		Type: item.TypeArmor,
		ArmorType: item.ArmorChest,
		Stats: item.Stats{Defense: 20},
	}
	
	helmet := &item.Item{
		Type: item.TypeArmor,
		ArmorType: item.ArmorHelmet,
		Stats: item.Stats{Defense: 10},
	}
	
	equip.Equip(chest)
	equip.Equip(helmet)
	
	totalDefense := equip.GetTotalDefense()
	if totalDefense != 30 {
		t.Errorf("Expected total defense 30, got %d", totalDefense)
	}
}

func TestGetSlotForItem(t *testing.T) {
	tests := []struct {
		name      string
		item      *item.Item
		wantSlot  EquipmentSlot
	}{
		{
			name:     "Weapon",
			item:     &item.Item{Type: item.TypeWeapon},
			wantSlot: SlotWeapon,
		},
		{
			name:     "Helmet",
			item:     &item.Item{Type: item.TypeArmor, ArmorType: item.ArmorHelmet},
			wantSlot: SlotHead,
		},
		{
			name:     "Chest",
			item:     &item.Item{Type: item.TypeArmor, ArmorType: item.ArmorChest},
			wantSlot: SlotChest,
		},
		{
			name:     "Shield",
			item:     &item.Item{Type: item.TypeArmor, ArmorType: item.ArmorShield},
			wantSlot: SlotOffhand,
		},
		{
			name:     "Accessory",
			item:     &item.Item{Type: item.TypeAccessory},
			wantSlot: SlotAccessory1,
		},
		{
			name:     "Consumable (not equippable)",
			item:     &item.Item{Type: item.TypeConsumable},
			wantSlot: EquipmentSlot(-1),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot := GetSlotForItem(tt.item)
			if slot != tt.wantSlot {
				t.Errorf("Expected slot %v, got %v", tt.wantSlot, slot)
			}
		})
	}
}

func TestEquipmentSlot_String(t *testing.T) {
	tests := []struct {
		slot EquipmentSlot
		want string
	}{
		{SlotWeapon, "weapon"},
		{SlotHead, "head"},
		{SlotChest, "chest"},
		{SlotAccessory1, "accessory1"},
		{EquipmentSlot(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.slot.String()
			if got != tt.want {
				t.Errorf("Expected '%s', got '%s'", tt.want, got)
			}
		})
	}
}
