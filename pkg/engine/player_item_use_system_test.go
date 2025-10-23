//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestPlayerItemUseSystem_UseConsumable tests using a healing potion.
func TestPlayerItemUseSystem_UseConsumable(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player with damaged health
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&InputComponent{UseItemPressed: true})
	player.AddComponent(&HealthComponent{Current: 50, Max: 100})

	inventory := NewInventoryComponent(10, 50.0)

	// Add healing potion
	potion := &item.Item{
		Name:           "Minor Health Potion",
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
		Rarity:         item.RarityCommon,
		Stats: item.Stats{
			Value:  10,
			Weight: 0.2,
		},
	}
	inventory.Items = append(inventory.Items, potion)
	player.AddComponent(inventory)

	world.Update(0)

	initialHealth := 50.0
	itemCount := len(inventory.Items)

	// Run item use system
	itemUseSys.Update(world.GetEntities(), 0.016)

	// Verify health increased (consumable effect)
	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Health should increase after using potion
	// Note: Actual healing amount depends on consumable implementation
	if health.Current == initialHealth {
		t.Error("Health should have increased after using potion")
	}

	// Verify item was removed from inventory
	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if len(inv.Items) != itemCount-1 {
		t.Errorf("Expected %d items after use, got %d", itemCount-1, len(inv.Items))
	}

	// Verify input was consumed
	inputComp, _ := player.GetComponent("input")
	input := inputComp.(*InputComponent)
	if input.UseItemPressed {
		t.Error("UseItemPressed should be false after use")
	}
}

// TestPlayerItemUseSystem_NoUsableItems tests when inventory has no consumables.
func TestPlayerItemUseSystem_NoUsableItems(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: true})

	inventory := NewInventoryComponent(10, 50.0)

	// Add non-consumable item (weapon)
	weapon := &item.Item{
		Name:       "Sword",
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
		Stats:      item.Stats{Damage: 10},
	}
	inventory.Items = append(inventory.Items, weapon)
	player.AddComponent(inventory)

	world.Update(0)

	// Run item use system
	itemUseSys.Update(world.GetEntities(), 0.016)

	// Verify weapon was not removed (not consumable)
	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item (weapon not consumed), got %d", len(inv.Items))
	}

	// Input should still be consumed
	inputComp, _ := player.GetComponent("input")
	input := inputComp.(*InputComponent)
	if input.UseItemPressed {
		t.Error("UseItemPressed should be consumed even if no usable item")
	}
}

// TestPlayerItemUseSystem_EmptyInventory tests using items with empty inventory.
func TestPlayerItemUseSystem_EmptyInventory(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player with empty inventory
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: true})
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)

	world.Update(0)

	// Should not panic
	itemUseSys.Update(world.GetEntities(), 0.016)

	// Input should be consumed
	inputComp, _ := player.GetComponent("input")
	input := inputComp.(*InputComponent)
	if input.UseItemPressed {
		t.Error("UseItemPressed should be consumed")
	}
}

// TestPlayerItemUseSystem_NoInputComponent tests system with non-player entity.
func TestPlayerItemUseSystem_NoInputComponent(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create entity without input (not player-controlled)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	inventory := NewInventoryComponent(10, 50.0)
	entity.AddComponent(inventory)

	world.Update(0)

	// Should not panic
	itemUseSys.Update(world.GetEntities(), 0.016)
}

// TestPlayerItemUseSystem_NoInventoryComponent tests player without inventory.
func TestPlayerItemUseSystem_NoInventoryComponent(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player without inventory
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: true})

	world.Update(0)

	// Should not panic
	itemUseSys.Update(world.GetEntities(), 0.016)
}

// TestPlayerItemUseSystem_MultipleConsumables tests that first consumable is used.
func TestPlayerItemUseSystem_MultipleConsumables(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: true})
	player.AddComponent(&HealthComponent{Current: 50, Max: 100})

	inventory := NewInventoryComponent(10, 50.0)

	// Add multiple potions
	potion1 := &item.Item{
		Name:           "Potion 1",
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
	}
	potion2 := &item.Item{
		Name:           "Potion 2",
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
	}
	inventory.Items = append(inventory.Items, potion1, potion2)
	player.AddComponent(inventory)

	world.Update(0)

	// Run item use system
	itemUseSys.Update(world.GetEntities(), 0.016)

	// Should have one less item
	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item remaining, got %d", len(inv.Items))
	}
}

// TestFindFirstUsableItem tests the helper function.
func TestFindFirstUsableItem(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	tests := []struct {
		name     string
		items    []*item.Item
		expected int
	}{
		{
			name: "First item consumable",
			items: []*item.Item{
				{Type: item.TypeConsumable, ConsumableType: item.ConsumablePotion},
				{Type: item.TypeWeapon},
			},
			expected: 0,
		},
		{
			name: "Second item consumable",
			items: []*item.Item{
				{Type: item.TypeWeapon},
				{Type: item.TypeConsumable, ConsumableType: item.ConsumablePotion},
			},
			expected: 1,
		},
		{
			name: "No consumables",
			items: []*item.Item{
				{Type: item.TypeWeapon},
				{Type: item.TypeArmor},
			},
			expected: -1,
		},
		{
			name:     "Empty inventory",
			items:    []*item.Item{},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := NewInventoryComponent(10, 50.0)
			inventory.Items = tt.items

			result := itemUseSys.findFirstUsableItem(inventory)
			if result != tt.expected {
				t.Errorf("Expected index %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestPlayerItemUseSystem_InputNotPressed tests that system does nothing without input.
func TestPlayerItemUseSystem_InputNotPressed(t *testing.T) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player with input not pressed
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: false})

	inventory := NewInventoryComponent(10, 50.0)
	potion := &item.Item{
		Type:           item.TypeConsumable,
		ConsumableType: item.ConsumablePotion,
	}
	inventory.Items = append(inventory.Items, potion)
	player.AddComponent(inventory)

	world.Update(0)

	// Run item use system
	itemUseSys.Update(world.GetEntities(), 0.016)

	// Item should not be used
	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*InventoryComponent)
	if len(inv.Items) != 1 {
		t.Error("Item should not be used when input not pressed")
	}
}

// Benchmark for player item use system performance
func BenchmarkPlayerItemUseSystem(b *testing.B) {
	world := NewWorld()
	inventorySys := NewInventorySystem(world)
	itemUseSys := NewPlayerItemUseSystem(inventorySys, world)

	// Create player with inventory
	player := world.CreateEntity()
	player.AddComponent(&InputComponent{UseItemPressed: true})
	player.AddComponent(&HealthComponent{Current: 50, Max: 100})

	inventory := NewInventoryComponent(10, 50.0)
	for i := 0; i < 5; i++ {
		potion := &item.Item{
			Type:           item.TypeConsumable,
			ConsumableType: item.ConsumablePotion,
		}
		inventory.Items = append(inventory.Items, potion)
	}
	player.AddComponent(inventory)

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset input for each iteration
		inputComp, _ := player.GetComponent("input")
		input := inputComp.(*InputComponent)
		input.UseItemPressed = true

		itemUseSys.Update(entities, 0.016)
	}
}
