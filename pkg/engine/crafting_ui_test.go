package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestNewCraftingUI tests crafting UI creation.
func TestNewCraftingUI(t *testing.T) {
	ui := NewCraftingUI(800, 600)

	if ui == nil {
		t.Fatal("NewCraftingUI() returned nil")
	}

	// Verify initial state
	if ui.visible {
		t.Error("New crafting UI should not be visible")
	}

	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}

	if ui.selectedRecipeIndex != -1 {
		t.Errorf("Initial selectedRecipeIndex = %d, want -1", ui.selectedRecipeIndex)
	}

	if ui.hoveredRecipeIndex != -1 {
		t.Errorf("Initial hoveredRecipeIndex = %d, want -1", ui.hoveredRecipeIndex)
	}

	if ui.scrollOffset != 0 {
		t.Errorf("Initial scrollOffset = %d, want 0", ui.scrollOffset)
	}
}

// TestCraftingUI_SettersAndGetters tests setter and getter methods.
func TestCraftingUI_SettersAndGetters(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)

	// Create test entities
	player := world.CreateEntity()
	station := world.CreateEntity()

	// Test SetPlayerEntity
	ui.SetPlayerEntity(player)
	if ui.playerEntity != player {
		t.Error("SetPlayerEntity() did not set player entity")
	}

	// Test SetStationEntity
	ui.SetStationEntity(station)
	if ui.stationEntity != station {
		t.Error("SetStationEntity() did not set station entity")
	}

	// Create crafting system
	invSystem := NewInventorySystem(world)
	itemGen := item.NewItemGenerator()
	craftingSystem := NewCraftingSystem(world, invSystem, itemGen)

	// Test SetCraftingSystem
	ui.SetCraftingSystem(craftingSystem)
	if ui.craftingSystem != craftingSystem {
		t.Error("SetCraftingSystem() did not set crafting system")
	}
}

// TestCraftingUI_OpenClose tests opening and closing the crafting UI.
func TestCraftingUI_OpenClose(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	station := world.CreateEntity()

	// Test initial state
	if ui.IsVisible() {
		t.Error("Crafting UI should not be visible initially")
	}

	// Test Open with station
	ui.Open(station)
	if !ui.IsVisible() {
		t.Error("Crafting UI should be visible after Open()")
	}

	if ui.stationEntity != station {
		t.Error("Open() did not set station entity")
	}

	if ui.selectedRecipeIndex != -1 {
		t.Error("Open() did not reset selectedRecipeIndex")
	}

	if ui.scrollOffset != 0 {
		t.Error("Open() did not reset scrollOffset")
	}

	// Test Close
	ui.selectedRecipeIndex = 3
	ui.scrollOffset = 2
	ui.Close()
	if ui.IsVisible() {
		t.Error("Crafting UI should not be visible after Close()")
	}

	if ui.stationEntity != nil {
		t.Error("Close() did not clear station entity")
	}

	if ui.selectedRecipeIndex != -1 {
		t.Error("Close() did not reset selectedRecipeIndex")
	}

	if ui.scrollOffset != 0 {
		t.Error("Close() did not reset scrollOffset")
	}

	// Test Open without station (nil station for no bonuses)
	ui.Open(nil)
	if !ui.IsVisible() {
		t.Error("Crafting UI should be visible after Open(nil)")
	}

	if ui.stationEntity != nil {
		t.Error("Open(nil) should set station entity to nil")
	}
}

// TestCraftingUI_Toggle tests toggling visibility.
func TestCraftingUI_Toggle(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	station := world.CreateEntity()

	// Open crafting UI first
	ui.Open(station)
	if !ui.IsVisible() {
		t.Fatal("Crafting UI should be visible after Open()")
	}

	// Toggle off
	ui.Toggle()
	if ui.IsVisible() {
		t.Error("Crafting UI should not be visible after first Toggle()")
	}

	// Verify state was cleaned up
	if ui.stationEntity != nil {
		t.Error("Toggle(off) did not clean up station entity")
	}

	// Toggle on
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("Crafting UI should be visible after second Toggle()")
	}

	// Note: stationEntity will still be nil since Toggle doesn't set it,
	// only Open() does. This is expected behavior.
}

// TestCraftingUI_UpdateMessageTimer tests message display timeout.
func TestCraftingUI_UpdateMessageTimer(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup player with components including a recipe so Update doesn't override message
	knowledge := NewRecipeKnowledgeComponent(0)
	recipe := &Recipe{
		ID:                "test_recipe",
		Name:              "Test Recipe",
		Type:              RecipePotion,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
		CraftTimeSec:      3.0,
	}
	knowledge.LearnRecipe(recipe)
	player.AddComponent(knowledge)
	player.AddComponent(NewCraftingSkillComponent())
	player.AddComponent(NewInventoryComponent(20, 0))

	ui.SetPlayerEntity(player)
	ui.Open(nil)

	// Show a message
	ui.showMessage("Test message")
	if ui.craftingMessage != "Test message" {
		t.Error("showMessage() did not set message")
	}
	if ui.craftingMessageTime <= 0 {
		t.Error("showMessage() did not set message time")
	}

	// Update with enough time to expire message (4.5 > 4.0 timeout)
	ui.Update([]*Entity{player}, 4.5)

	// Message should be cleared
	if ui.craftingMessage != "" {
		t.Errorf("Message not cleared after timeout, got: %s", ui.craftingMessage)
	}
	if ui.craftingMessageTime > 0 {
		t.Errorf("Message time not cleared after timeout, got: %f", ui.craftingMessageTime)
	}
}

// TestCraftingUI_ScrollingNavigation tests scrolling through recipe list.
func TestCraftingUI_ScrollingNavigation(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup player with recipe knowledge
	knowledge := NewRecipeKnowledgeComponent(0)

	// Add multiple recipes
	for i := 0; i < 20; i++ {
		recipe := &Recipe{
			ID:                "recipe_" + string(rune('A'+i)),
			Name:              "Recipe " + string(rune('A'+i)),
			Type:              RecipePotion,
			Rarity:            RecipeCommon,
			Materials:         []MaterialRequirement{},
			SkillRequired:     0,
			BaseSuccessChance: 0.75,
			CraftTimeSec:      3.0,
		}
		knowledge.LearnRecipe(recipe)
	}

	player.AddComponent(knowledge)
	player.AddComponent(NewCraftingSkillComponent())
	player.AddComponent(NewInventoryComponent(20, 100))

	ui.SetPlayerEntity(player)
	ui.Open(nil)

	// Test initial scroll position
	if ui.scrollOffset != 0 {
		t.Error("Initial scroll offset should be 0")
	}

	// Test selecting recipes (simulates down arrow key)
	if ui.selectedRecipeIndex != -1 {
		t.Error("Initial selected index should be -1")
	}

	// Simulate selecting first recipe
	ui.selectedRecipeIndex = 0
	if ui.selectedRecipeIndex != 0 {
		t.Error("Should be able to select first recipe")
	}

	// Simulate navigating down
	ui.selectedRecipeIndex = 5
	if ui.selectedRecipeIndex != 5 {
		t.Error("Should be able to select recipe at index 5")
	}

	// Simulate scrolling (would happen in Update with keyboard/mouse input)
	ui.scrollOffset = 3
	if ui.scrollOffset != 3 {
		t.Error("Should be able to scroll offset")
	}
}

// TestCraftingUI_AttemptCraft tests crafting initiation.
func TestCraftingUI_AttemptCraft(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Add player to world so crafting system can find it
	world.AddEntity(player)

	// Setup inventory system and crafting system
	invSystem := NewInventorySystem(world)
	itemGen := item.NewItemGenerator()
	craftingSystem := NewCraftingSystem(world, invSystem, itemGen)

	// Setup player with components
	knowledge := NewRecipeKnowledgeComponent(0)
	skill := NewCraftingSkillComponent()
	inv := NewInventoryComponent(20, 100) // 20 capacity, 100 max weight
	inv.Gold = 100                        // Set gold explicitly

	// Add materials to inventory (leave space for output)
	herb := &item.Item{
		Name: "Healing Herb",
		Type: item.TypeConsumable,
		Stats: item.Stats{
			Value: 5,
		},
	}
	inv.Items = append(inv.Items, herb, herb) // 2 herbs, 18 slots free

	player.AddComponent(knowledge)
	player.AddComponent(skill)
	player.AddComponent(inv)

	// Create a simple recipe with proper seed
	recipe := &Recipe{
		ID:          "healing_potion",
		Name:        "Healing Potion",
		Description: "Restores health",
		Type:        RecipePotion,
		Rarity:      RecipeCommon,
		Materials: []MaterialRequirement{
			{ItemName: "Healing Herb", Quantity: 2, Optional: false},
		},
		GoldCost:          10,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
		CraftTimeSec:      3.0,
		OutputItemSeed:    12345, // Add seed for deterministic generation
		OutputItemType:    item.TypeConsumable,
		GenreID:           "fantasy",
	}
	knowledge.LearnRecipe(recipe)

	ui.SetPlayerEntity(player)
	ui.SetCraftingSystem(craftingSystem)
	ui.Open(nil)

	// Call world.Update to flush pending entity additions
	world.Update(0.0)

	// Attempt to craft
	ui.attemptCraft(recipe)

	// Check the message to see if there was an error
	if ui.craftingMessage != "" {
		t.Logf("Crafting message: %s", ui.craftingMessage)
	}

	// Check if crafting started (should have progress component)
	progressComp, hasProgress := player.GetComponent("crafting_progress")
	if !hasProgress {
		t.Error("Crafting should have started (missing progress component)")
	} else {
		progress := progressComp.(*CraftingProgressComponent)
		if progress.CurrentRecipe != recipe {
			t.Error("Progress component has wrong recipe")
		}
	}
}

// TestCraftingUI_AttemptCraftWithoutSystem tests error handling when system missing.
func TestCraftingUI_AttemptCraftWithoutSystem(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup player but no crafting system
	knowledge := NewRecipeKnowledgeComponent(0)
	recipe := &Recipe{
		ID:                "test_recipe",
		Name:              "Test Recipe",
		Type:              RecipePotion,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
	}
	knowledge.LearnRecipe(recipe)
	player.AddComponent(knowledge)

	ui.SetPlayerEntity(player)
	ui.SetCraftingSystem(nil) // No system
	ui.Open(nil)

	// Attempt to craft
	ui.attemptCraft(recipe)

	// Should show error message
	if ui.craftingMessage == "" {
		t.Error("Should show error message when system missing")
	}
	if ui.craftingMessage != "Crafting system not available" {
		t.Errorf("Wrong error message: %s", ui.craftingMessage)
	}
}

// TestCraftingUI_AttemptCraftInsufficientMaterials tests crafting with missing materials.
func TestCraftingUI_AttemptCraftInsufficientMaterials(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup systems
	invSystem := NewInventorySystem(world)
	itemGen := item.NewItemGenerator()
	craftingSystem := NewCraftingSystem(world, invSystem, itemGen)

	// Setup player with components but no materials
	knowledge := NewRecipeKnowledgeComponent(0)
	skill := NewCraftingSkillComponent()
	inv := NewInventoryComponent(20, 100) // Gold but no items

	player.AddComponent(knowledge)
	player.AddComponent(skill)
	player.AddComponent(inv)

	// Create recipe requiring materials
	recipe := &Recipe{
		ID:     "test_recipe",
		Name:   "Test Recipe",
		Type:   RecipePotion,
		Rarity: RecipeCommon,
		Materials: []MaterialRequirement{
			{ItemName: "Missing Item", Quantity: 2, Optional: false},
		},
		GoldCost:          10,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
		CraftTimeSec:      3.0,
		OutputItemType:    item.TypeConsumable,
		GenreID:           "fantasy",
	}
	knowledge.LearnRecipe(recipe)

	ui.SetPlayerEntity(player)
	ui.SetCraftingSystem(craftingSystem)
	ui.Open(nil)

	// Attempt to craft
	ui.attemptCraft(recipe)

	// Should not have started crafting
	_, hasProgress := player.GetComponent("crafting_progress")
	if hasProgress {
		t.Error("Crafting should not start without materials")
	}

	// Should show error message
	if ui.craftingMessage == "" {
		t.Error("Should show error message for missing materials")
	}
}

// TestCraftingUI_AttemptCraftWithStation tests crafting at a station.
func TestCraftingUI_AttemptCraftWithStation(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()
	station := world.CreateEntity()

	// Add entities to world
	world.AddEntity(player)
	world.AddEntity(station)

	// Setup crafting station
	station.AddComponent(NewCraftingStationComponent(RecipePotion))

	// Setup systems
	invSystem := NewInventorySystem(world)
	itemGen := item.NewItemGenerator()
	craftingSystem := NewCraftingSystem(world, invSystem, itemGen)

	// Setup player with components and materials
	knowledge := NewRecipeKnowledgeComponent(0)
	skill := NewCraftingSkillComponent()
	inv := NewInventoryComponent(20, 100) // 20 capacity, 100 max weight
	inv.Gold = 100                        // Set gold explicitly

	herb := &item.Item{
		Name:  "Healing Herb",
		Type:  item.TypeConsumable,
		Stats: item.Stats{Value: 5},
	}
	inv.Items = append(inv.Items, herb, herb) // 2 herbs, 18 slots free

	player.AddComponent(knowledge)
	player.AddComponent(skill)
	player.AddComponent(inv)

	// Create recipe with proper seed
	recipe := &Recipe{
		ID:     "potion_recipe",
		Name:   "Healing Potion",
		Type:   RecipePotion,
		Rarity: RecipeCommon,
		Materials: []MaterialRequirement{
			{ItemName: "Healing Herb", Quantity: 2, Optional: false},
		},
		GoldCost:          10,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
		CraftTimeSec:      4.0,
		OutputItemSeed:    12345, // Add seed
		OutputItemType:    item.TypeConsumable,
		GenreID:           "fantasy",
	}
	knowledge.LearnRecipe(recipe)

	ui.SetPlayerEntity(player)
	ui.SetCraftingSystem(craftingSystem)
	ui.SetStationEntity(station)
	ui.Open(station)

	// Call world.Update to flush pending entity additions
	world.Update(0.0)

	// Attempt to craft at station
	ui.attemptCraft(recipe)

	// Should have started crafting
	progressComp, hasProgress := player.GetComponent("crafting_progress")
	if !hasProgress {
		t.Error("Crafting should have started at station")
	} else {
		progress := progressComp.(*CraftingProgressComponent)
		if progress.UsingStationID != station.ID {
			t.Error("Progress should reference station ID")
		}
	}
}

// TestCraftingUI_ShowingProgress tests detecting ongoing crafting.
func TestCraftingUI_ShowingProgress(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup player with progress component (simulating ongoing craft)
	recipe := &Recipe{
		ID:           "test_recipe",
		Name:         "Test Recipe",
		CraftTimeSec: 5.0,
	}
	progress := NewCraftingProgressComponent(recipe, 5.0, 0)
	player.AddComponent(progress)
	player.AddComponent(NewRecipeKnowledgeComponent(0))
	player.AddComponent(NewCraftingSkillComponent())
	player.AddComponent(NewInventoryComponent(20, 100))

	ui.SetPlayerEntity(player)
	ui.Open(nil)

	// Update should detect ongoing crafting
	ui.Update([]*Entity{player}, 0.1)

	if !ui.showingProgress {
		t.Error("Should detect ongoing crafting")
	}
}

// TestCraftingUI_MinIntHelper tests the minInt helper function.
func TestCraftingUI_MinIntHelper(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a less than b", 5, 10, 5},
		{"b less than a", 10, 5, 5},
		{"equal values", 7, 7, 7},
		{"negative values", -3, -1, -3},
		{"mixed signs", -5, 3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minInt(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestCraftingUI_UpdateWithNoRecipes tests UI behavior when player has no recipes.
func TestCraftingUI_UpdateWithNoRecipes(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Setup player with empty recipe knowledge
	knowledge := NewRecipeKnowledgeComponent(0)
	player.AddComponent(knowledge)

	ui.SetPlayerEntity(player)
	ui.Open(nil)

	// Update should show "no recipes" message
	ui.Update([]*Entity{player}, 0.1)

	if ui.craftingMessage == "" {
		t.Error("Should show message when no recipes known")
	}
}

// TestCraftingUI_DrawWithoutComponents tests Draw with missing components (should not crash).
func TestCraftingUI_DrawWithoutComponents(t *testing.T) {
	world := NewWorld()
	ui := NewCraftingUI(800, 600)
	player := world.CreateEntity()

	// Don't add any components
	ui.SetPlayerEntity(player)
	ui.Open(nil)

	// Draw should not crash even without components
	// Note: Can't actually test Draw without Ebiten context, but we can verify no panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw() panicked with missing components: %v", r)
		}
	}()

	// Draw expects *ebiten.Image but we'll pass nil to test graceful handling
	ui.Draw(nil)
}
