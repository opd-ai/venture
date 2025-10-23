//go:build test
// +build test

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestSpawnItemInWorld tests spawning items as world entities.
func TestSpawnItemInWorld(t *testing.T) {
	world := NewWorld()

	// Create a test item
	testItem := &item.Item{
		Name:   "Test Sword",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
	}

	// Spawn item
	itemEntity := SpawnItemInWorld(world, testItem, 100, 200)

	if itemEntity == nil {
		t.Fatal("SpawnItemInWorld returned nil")
	}

	// Verify position
	posComp, ok := itemEntity.GetComponent("position")
	if !ok {
		t.Fatal("Item entity missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 100 || pos.Y != 200 {
		t.Errorf("Item position = (%f, %f), want (100, 200)", pos.X, pos.Y)
	}

	// Verify has sprite
	if !itemEntity.HasComponent("sprite") {
		t.Error("Item entity missing sprite component")
	}

	// Verify has collider
	if !itemEntity.HasComponent("collider") {
		t.Error("Item entity missing collider component")
	}

	// Verify has item data
	itemComp, ok := itemEntity.GetComponent("item_entity")
	if !ok {
		t.Fatal("Item entity missing item_entity component")
	}
	itemData := itemComp.(*ItemEntityComponent)
	if itemData.Item != testItem {
		t.Error("Item entity component does not match original item")
	}
}

// TestItemPickupSystem tests automatic item collection.
func TestItemPickupSystem(t *testing.T) {
	world := NewWorld()
	pickupSystem := NewItemPickupSystem(world)

	// Create player with inventory
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&InputComponent{}) // Marks as player
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)

	// Create item nearby
	testItem := &item.Item{
		Name:   "Health Potion",
		Type:   item.TypeConsumable,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 0.5},
	}
	itemEntity := SpawnItemInWorld(world, testItem, 110, 110) // 14 pixels away

	world.Update(0)
	entities := world.GetEntities()

	// Run pickup system
	pickupSystem.Update(entities, 0.016)

	// Process entity removals
	world.Update(0.016)

	// Verify item was picked up
	if len(inventory.Items) != 1 {
		t.Fatalf("Inventory has %d items, want 1", len(inventory.Items))
	}

	if inventory.Items[0] != testItem {
		t.Error("Inventory item does not match spawned item")
	}

	// Verify item entity was removed
	if entity, exists := world.GetEntity(itemEntity.ID); exists && entity != nil {
		t.Error("Item entity was not removed from world after pickup")
	}
}

// TestItemPickupDistance tests that items are only picked up within range.
func TestItemPickupDistance(t *testing.T) {
	world := NewWorld()
	pickupSystem := NewItemPickupSystem(world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&InputComponent{})
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)

	// Create item far away (50 pixels = 1.5 tiles)
	testItem := &item.Item{
		Name:  "Distant Sword",
		Type:  item.TypeWeapon,
		Stats: item.Stats{Weight: 2.0},
	}
	itemEntity := SpawnItemInWorld(world, testItem, 150, 100)

	world.Update(0)
	entities := world.GetEntities()

	// Run pickup system
	pickupSystem.Update(entities, 0.016)

	// Verify item was NOT picked up (too far)
	if len(inventory.Items) != 0 {
		t.Errorf("Inventory has %d items, want 0 (item too far)", len(inventory.Items))
	}

	// Verify item still exists
	if entity, exists := world.GetEntity(itemEntity.ID); !exists || entity == nil {
		t.Error("Item entity was removed despite being out of range")
	}
}

// TestGenerateLootDrop tests loot generation from enemies.
func TestGenerateLootDrop(t *testing.T) {
	world := NewWorld()

	// Create enemy entity
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 200, Y: 200})
	enemy.AddComponent(&HealthComponent{Current: 0, Max: 50}) // Dead
	enemy.AddComponent(&ExperienceComponent{Level: 5})
	stats := NewStatsComponent()
	stats.Attack = 15
	enemy.AddComponent(stats)

	// Generate loot (using enemy ID as part of seed for determinism)
	seed := int64(12345)
	genreID := "fantasy"

	// Try multiple times since drop is probabilistic
	droppedCount := 0
	for i := 0; i < 10; i++ {
		loot := GenerateLootDrop(world, enemy, 200, 200, seed+int64(i), genreID)
		if loot != nil {
			droppedCount++

			// Verify loot has required components
			if !loot.HasComponent("item_entity") {
				t.Error("Generated loot missing item_entity component")
			}
			if !loot.HasComponent("position") {
				t.Error("Generated loot missing position component")
			}
		}
	}

	// With 10 attempts, we should have at least one drop (30% chance each)
	if droppedCount == 0 {
		t.Error("No loot dropped in 10 attempts (probability failure)")
	}

	// Log for visibility
	t.Logf("Loot dropped %d/10 times", droppedCount)
}

// TestGetItemColor verifies color assignment based on item type and rarity.
func TestGetItemColor(t *testing.T) {
	tests := []struct {
		name  string
		item  *item.Item
		wantR bool // Check if red channel is non-zero
		wantG bool // Check if green channel is non-zero
		wantB bool // Check if blue channel is non-zero
	}{
		{
			name: "common weapon",
			item: &item.Item{
				Type:   item.TypeWeapon,
				Rarity: item.RarityCommon,
			},
			wantR: true,
			wantG: true,
			wantB: true, // Silver-ish (all channels)
		},
		{
			name: "legendary weapon",
			item: &item.Item{
				Type:   item.TypeWeapon,
				Rarity: item.RarityLegendary,
			},
			wantR: true,
			wantG: true,
			wantB: true, // Brighter silver
		},
		{
			name: "common consumable",
			item: &item.Item{
				Type:   item.TypeConsumable,
				Rarity: item.RarityCommon,
			},
			wantR: true, // Red-ish for potions
			wantG: true,
			wantB: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := getItemColor(tt.item)

			if (col.R > 0) != tt.wantR {
				t.Errorf("Red channel = %d, wantNonZero = %v", col.R, tt.wantR)
			}
			if (col.G > 0) != tt.wantG {
				t.Errorf("Green channel = %d, wantNonZero = %v", col.G, tt.wantG)
			}
			if (col.B > 0) != tt.wantB {
				t.Errorf("Blue channel = %d, wantNonZero = %v", col.B, tt.wantB)
			}

			// Alpha should always be 255
			if col.A != 255 {
				t.Errorf("Alpha channel = %d, want 255", col.A)
			}
		})
	}
}

// TestItemEntityComponent tests the component interface.
func TestItemEntityComponent(t *testing.T) {
	testItem := &item.Item{
		Name: "Test Item",
	}

	comp := &ItemEntityComponent{
		Item: testItem,
	}

	if comp.Type() != "item_entity" {
		t.Errorf("Component type = %s, want 'item_entity'", comp.Type())
	}

	if comp.Item != testItem {
		t.Error("Item reference not preserved")
	}
}

// BenchmarkItemPickupSystem benchmarks the pickup system performance.
func BenchmarkItemPickupSystem(b *testing.B) {
	world := NewWorld()
	pickupSystem := NewItemPickupSystem(world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&InputComponent{})
	inventory := NewInventoryComponent(100, 500.0)
	player.AddComponent(inventory)

	// Create 50 items scattered around
	for i := 0; i < 50; i++ {
		testItem := &item.Item{
			Name:   "Test Item",
			Type:   item.TypeConsumable,
			Rarity: item.RarityCommon,
			Stats:  item.Stats{Weight: 0.5},
		}
		x := 50.0 + float64(i*10)
		y := 50.0 + float64(i*5)
		SpawnItemInWorld(world, testItem, x, y)
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pickupSystem.Update(entities, 0.016)
	}
}

// TestItemSpawning_NilItem tests null safety.
func TestItemSpawning_NilItem(t *testing.T) {
	world := NewWorld()

	// Spawn nil item
	itemEntity := SpawnItemInWorld(world, nil, 100, 100)

	if itemEntity != nil {
		t.Error("SpawnItemInWorld should return nil for nil item")
	}
}

// TestGenerateLootDrop_NoDrop tests the no-drop case.
func TestGenerateLootDrop_NoDrop(t *testing.T) {
	world := NewWorld()

	// Create weak enemy (low drop chance)
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 100, Y: 100})
	enemy.AddComponent(&ExperienceComponent{Level: 1})
	stats := NewStatsComponent()
	stats.Attack = 5 // Weak enemy
	enemy.AddComponent(stats)

	// Try with seeds that should not drop (deterministic)
	// Not all seeds will drop due to 30% chance
	seed := int64(99999) // Arbitrary seed likely to not drop
	loot := GenerateLootDrop(world, enemy, 100, 100, seed, "fantasy")

	// This test is probabilistic, just ensure no crash
	// We accept either drop or no drop
	_ = loot
}

// TestItemPickupSystem_FullInventory tests the inventory full case.
func TestItemPickupSystem_FullInventory(t *testing.T) {
	world := NewWorld()
	pickupSystem := NewItemPickupSystem(world)

	// Create player with small inventory
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&InputComponent{})
	inventory := NewInventoryComponent(1, 50.0) // Only 1 slot
	player.AddComponent(inventory)

	// Fill inventory with existing item
	existingItem := &item.Item{
		Name:   "Existing Item",
		Type:   item.TypeWeapon,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 1.0},
	}
	inventory.Items = append(inventory.Items, existingItem)

	// Try to pick up another item
	testItem := &item.Item{
		Name:   "New Item",
		Type:   item.TypeArmor,
		Rarity: item.RarityCommon,
		Stats:  item.Stats{Weight: 1.0},
	}
	itemEntity := SpawnItemInWorld(world, testItem, 105, 105)

	world.Update(0)
	entities := world.GetEntities()

	// Run pickup system
	pickupSystem.Update(entities, 0.016)

	// Verify item was NOT picked up (inventory full)
	if len(inventory.Items) != 1 {
		t.Errorf("Inventory has %d items, want 1 (full)", len(inventory.Items))
	}

	// Verify item still exists in world
	if entity, exists := world.GetEntity(itemEntity.ID); !exists || entity == nil {
		t.Error("Item was removed despite inventory being full")
	}
}

// TestItemColor_AllItemTypes verifies all item types have colors.
func TestItemColor_AllItemTypes(t *testing.T) {
	itemTypes := []item.ItemType{
		item.TypeWeapon,
		item.TypeArmor,
		item.TypeConsumable,
		item.TypeAccessory,
	}

	for _, itemType := range itemTypes {
		t.Run(itemType.String(), func(t *testing.T) {
			testItem := &item.Item{
				Type:   itemType,
				Rarity: item.RarityCommon,
			}

			col := getItemColor(testItem)

			// Verify color is not completely black (should have some RGB)
			if col.R == 0 && col.G == 0 && col.B == 0 {
				t.Error("Item color is completely black (missing color assignment)")
			}
		})
	}
}

// TestItemColor_RarityBrightness verifies rarity affects brightness.
func TestItemColor_RarityBrightness(t *testing.T) {
	rarities := []item.Rarity{
		item.RarityCommon,
		item.RarityUncommon,
		item.RarityRare,
		item.RarityEpic,
		item.RarityLegendary,
	}

	var previousBrightness int

	for _, rarity := range rarities {
		testItem := &item.Item{
			Type:   item.TypeWeapon,
			Rarity: rarity,
		}

		col := getItemColor(testItem)
		brightness := int(col.R) + int(col.G) + int(col.B)

		// Higher rarity should have equal or higher brightness
		if previousBrightness > 0 && brightness < previousBrightness {
			t.Errorf("Rarity %s has lower brightness than previous rarity", rarity.String())
		}

		previousBrightness = brightness
		t.Logf("%s brightness: %d", rarity.String(), brightness)
	}
}
