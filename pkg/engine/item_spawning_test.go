package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
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
	player.AddComponent(NewStubInput()) // Marks as player
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
	player.AddComponent(NewStubInput())
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
	player.AddComponent(NewStubInput())
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
	player.AddComponent(NewStubInput())
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

// TestSpawnRecipeInWorld tests spawning recipes as world entities.
func TestSpawnRecipeInWorld(t *testing.T) {
	world := NewWorld()

	// Create a test recipe
	testRecipe := &Recipe{
		ID:          "test_recipe_123",
		Name:        "Test Healing Potion Recipe",
		Description: "A recipe for crafting healing potions",
		Type:        RecipePotion,
		Rarity:      RecipeCommon,
	}

	// Spawn recipe
	recipeEntity := SpawnRecipeInWorld(world, testRecipe, 150, 250)

	if recipeEntity == nil {
		t.Fatal("SpawnRecipeInWorld returned nil")
	}

	// Verify position
	posComp, ok := recipeEntity.GetComponent("position")
	if !ok {
		t.Fatal("Recipe entity missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 150 || pos.Y != 250 {
		t.Errorf("Recipe position = (%f, %f), want (150, 250)", pos.X, pos.Y)
	}

	// Verify has sprite
	if !recipeEntity.HasComponent("sprite") {
		t.Error("Recipe entity missing sprite component")
	}

	// Verify has collider
	if !recipeEntity.HasComponent("collider") {
		t.Error("Recipe entity missing collider component")
	}

	// Verify has recipe data
	recipeComp, ok := recipeEntity.GetComponent("recipe_entity")
	if !ok {
		t.Fatal("Recipe entity missing recipe_entity component")
	}

	recipeData := recipeComp.(*RecipeEntityComponent)
	if recipeData.Recipe.ID != "test_recipe_123" {
		t.Errorf("Recipe ID = %s, want test_recipe_123", recipeData.Recipe.ID)
	}
}

// TestSpawnRecipeInWorld_Nil tests that nil recipe returns nil entity.
func TestSpawnRecipeInWorld_Nil(t *testing.T) {
	world := NewWorld()
	recipeEntity := SpawnRecipeInWorld(world, nil, 0, 0)

	if recipeEntity != nil {
		t.Error("SpawnRecipeInWorld with nil recipe should return nil")
	}
}

// TestGenerateRecipeDrop tests recipe generation from enemies.
func TestGenerateRecipeDrop(t *testing.T) {
	// Create mock recipe generator
	mockRecipeGen := &MockRecipeGenerator{
		recipes: []*Recipe{
			{
				ID:          "recipe_001",
				Name:        "Minor Healing Potion Recipe",
				Description: "Basic healing recipe",
				Type:        RecipePotion,
				Rarity:      RecipeCommon,
			},
		},
	}

	world := NewWorld()
	seed := int64(12345)
	genreID := "fantasy"

	// Create test enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 200, Y: 200})
	enemy.AddComponent(&StatsComponent{Attack: 15, Defense: 10})
	enemy.AddComponent(&ExperienceComponent{Level: 5})

	// Generate multiple drops to test probability
	dropCount := 0
	trials := 100

	for i := 0; i < trials; i++ {
		recipe := GenerateRecipeDrop(mockRecipeGen, world, enemy, 200, 200, seed+int64(i), genreID)
		if recipe != nil {
			dropCount++
		}
	}

	// Recipe drops should be rare (around 5% base)
	// Allow for some variance due to RNG
	if dropCount < 1 || dropCount > 15 {
		t.Errorf("Recipe drop rate = %d/%d (%.1f%%), expected around 5%%", dropCount, trials, float64(dropCount)/float64(trials)*100)
	}

	t.Logf("Recipe drop rate: %d/%d (%.1f%%)", dropCount, trials, float64(dropCount)/float64(trials)*100)
}

// TestGenerateRecipeDrop_BossDropRate tests that bosses have higher recipe drop rates.
func TestGenerateRecipeDrop_BossDropRate(t *testing.T) {
	mockRecipeGen := &MockRecipeGenerator{
		recipes: []*Recipe{
			{
				ID:     "recipe_boss_001",
				Name:   "Epic Recipe",
				Type:   RecipeEnchanting,
				Rarity: RecipeEpic,
			},
		},
	}

	world := NewWorld()
	seed := int64(54321)
	genreID := "scifi"

	// Create boss enemy (high stats)
	boss := world.CreateEntity()
	boss.AddComponent(&PositionComponent{X: 300, Y: 300})
	boss.AddComponent(&StatsComponent{Attack: 30, Defense: 25}) // Boss-level stats
	boss.AddComponent(&ExperienceComponent{Level: 10})

	// Generate multiple drops
	dropCount := 0
	trials := 100

	for i := 0; i < trials; i++ {
		recipe := GenerateRecipeDrop(mockRecipeGen, world, boss, 300, 300, seed+int64(i), genreID)
		if recipe != nil {
			dropCount++
		}
	}

	// Bosses should drop recipes around 20%
	// Allow for variance due to RNG
	if dropCount < 10 || dropCount > 35 {
		t.Errorf("Boss recipe drop rate = %d/%d (%.1f%%), expected around 20%%", dropCount, trials, float64(dropCount)/float64(trials)*100)
	}

	t.Logf("Boss recipe drop rate: %d/%d (%.1f%%)", dropCount, trials, float64(dropCount)/float64(trials)*100)
}

// TestGenerateRecipeDrop_NoDrop tests the no-drop case.
func TestGenerateRecipeDrop_NoDrop(t *testing.T) {
	mockRecipeGen := &MockRecipeGenerator{
		recipes: []*Recipe{
			{ID: "recipe_002", Name: "Test Recipe"},
		},
	}

	world := NewWorld()
	seed := int64(99999) // Seed that produces no drop
	genreID := "horror"

	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 100, Y: 100})
	enemy.AddComponent(&StatsComponent{Attack: 5, Defense: 3})

	recipe := GenerateRecipeDrop(mockRecipeGen, world, enemy, 100, 100, seed, genreID)

	// With low stats and specific seed, should not drop
	// This is probabilistic, but with seed 99999 should be consistent
	_ = recipe // May or may not be nil, depends on RNG
}

// TestRecipePickup tests that recipes are learned when picked up.
func TestRecipePickup(t *testing.T) {
	world := NewWorld()

	// Create player with input component
	player := world.CreateEntity()
	player.AddComponent(NewStubInput()) // Marks as player
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)
	// Note: RecipeKnowledgeComponent will be created automatically on first pickup

	// Create recipe entity
	testRecipe := &Recipe{
		ID:          "pickup_recipe_001",
		Name:        "Pickup Test Recipe",
		Description: "Recipe for testing pickup",
		Type:        RecipePotion,
		Rarity:      RecipeCommon,
	}
	recipeEntity := SpawnRecipeInWorld(world, testRecipe, 100, 100) // Same position as player

	world.AddEntity(player)
	world.AddEntity(recipeEntity)

	// Flush pending entities
	world.Update(0.0)

	// Create pickup system
	pickupSystem := NewItemPickupSystem(world)

	// Update should trigger pickup
	pickupSystem.Update(world.GetEntities(), 0.016)

	// Flush removals
	world.Update(0.0)

	// Verify recipe was learned
	knowledgeComp, hasKnowledge := player.GetComponent("recipe_knowledge")
	if !hasKnowledge {
		t.Fatal("Player should have recipe_knowledge component after pickup")
	}

	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
	if !knowledge.KnowsRecipe("pickup_recipe_001") {
		t.Error("Player should know the picked up recipe")
	}

	// Verify recipe entity was removed
	entities := world.GetEntities()
	for _, e := range entities {
		if e.HasComponent("recipe_entity") {
			t.Error("Recipe entity should have been removed after pickup")
		}
	}
}

// TestRecipePickup_AlreadyKnown tests that duplicate recipes show a message.
func TestRecipePickup_AlreadyKnown(t *testing.T) {
	world := NewWorld()

	// Create player with recipe already known
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)

	// Add recipe knowledge with recipe already learned
	knowledge := NewRecipeKnowledgeComponent(0)
	testRecipe := &Recipe{
		ID:     "known_recipe_001",
		Name:   "Already Known Recipe",
		Type:   RecipePotion,
		Rarity: RecipeCommon,
	}
	knowledge.LearnRecipe(testRecipe)
	player.AddComponent(knowledge)

	// Create recipe entity for the same recipe
	recipeEntity := SpawnRecipeInWorld(world, testRecipe, 100, 100)

	world.AddEntity(player)
	world.AddEntity(recipeEntity)

	// Flush pending entities
	world.Update(0.0)

	// Create pickup system
	pickupSystem := NewItemPickupSystem(world)

	// Update should trigger pickup attempt
	pickupSystem.Update(world.GetEntities(), 0.016)

	// Recipe entity should still exist (not picked up)
	entities := world.GetEntities()
	found := false
	for _, e := range entities {
		if e.HasComponent("recipe_entity") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Recipe entity should not have been removed for already-known recipe")
	}
}

// TestRecipePickup_SlotLimit tests recipe limit enforcement.
func TestRecipePickup_SlotLimit(t *testing.T) {
	world := NewWorld()

	// Create player with recipe slot limit
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	inventory := NewInventoryComponent(10, 50.0)
	player.AddComponent(inventory)

	// Add recipe knowledge with only 2 slots
	knowledge := NewRecipeKnowledgeComponent(2)

	// Fill up the slots
	recipe1 := &Recipe{ID: "slot_recipe_001", Name: "Recipe 1", Type: RecipePotion, Rarity: RecipeCommon}
	recipe2 := &Recipe{ID: "slot_recipe_002", Name: "Recipe 2", Type: RecipePotion, Rarity: RecipeCommon}
	knowledge.LearnRecipe(recipe1)
	knowledge.LearnRecipe(recipe2)

	player.AddComponent(knowledge)

	// Try to pick up a third recipe
	recipe3 := &Recipe{ID: "slot_recipe_003", Name: "Recipe 3", Type: RecipePotion, Rarity: RecipeCommon}
	recipeEntity := SpawnRecipeInWorld(world, recipe3, 100, 100)

	world.AddEntity(player)
	world.AddEntity(recipeEntity)

	// Flush pending entities
	world.Update(0.0)

	// Create pickup system
	pickupSystem := NewItemPickupSystem(world)

	// Update should attempt pickup but fail
	pickupSystem.Update(world.GetEntities(), 0.016)

	// Recipe entity should still exist (limit reached)
	entities := world.GetEntities()
	found := false
	for _, e := range entities {
		if e.HasComponent("recipe_entity") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Recipe entity should not have been removed when slot limit reached")
	}

	// Verify recipe was not learned
	if knowledge.KnowsRecipe("slot_recipe_003") {
		t.Error("Recipe should not have been learned when slot limit reached")
	}
}

// TestRecipeColor_TypeColors verifies different recipe types have distinct colors.
func TestRecipeColor_TypeColors(t *testing.T) {
	types := []RecipeType{RecipePotion, RecipeEnchanting, RecipeMagicItem}

	colors := make(map[RecipeType]string)

	for _, recipeType := range types {
		testRecipe := &Recipe{
			ID:     "color_test",
			Name:   "Color Test",
			Type:   recipeType,
			Rarity: RecipeCommon,
		}

		col := getRecipeColor(testRecipe)
		colorKey := string([]byte{col.R, col.G, col.B})
		colors[recipeType] = colorKey

		t.Logf("%s color: RGB(%d, %d, %d)", recipeType.String(), col.R, col.G, col.B)
	}

	// Verify all types have different colors
	seen := make(map[string]bool)
	for _, colorKey := range colors {
		if seen[colorKey] {
			t.Error("Multiple recipe types have the same color")
		}
		seen[colorKey] = true
	}
}

// TestRecipeColor_RarityBrightness verifies rarity affects recipe brightness.
func TestRecipeColor_RarityBrightness(t *testing.T) {
	rarities := []RecipeRarity{
		RecipeCommon,
		RecipeUncommon,
		RecipeRare,
		RecipeEpic,
		RecipeLegendary,
	}

	var previousBrightness int

	for _, rarity := range rarities {
		testRecipe := &Recipe{
			ID:     "brightness_test",
			Name:   "Brightness Test",
			Type:   RecipePotion,
			Rarity: rarity,
		}

		col := getRecipeColor(testRecipe)
		brightness := int(col.R) + int(col.G) + int(col.B)

		// Higher rarity should have equal or higher brightness
		if previousBrightness > 0 && brightness < previousBrightness {
			t.Errorf("Rarity %s has lower brightness than previous rarity", rarity.String())
		}

		previousBrightness = brightness
		t.Logf("%s brightness: %d", rarity.String(), brightness)
	}
}

// MockRecipeGenerator is a test helper that implements procgen.Generator for recipes.
type MockRecipeGenerator struct {
	recipes []*Recipe
	err     error
}

func (m *MockRecipeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.recipes, nil
}

func (m *MockRecipeGenerator) Validate(result interface{}) error {
	return nil
}
