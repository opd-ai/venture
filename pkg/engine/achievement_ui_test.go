// Package engine provides tests for the achievement UI system.
//
// Phase 86: Statistics UI (V15.0)
package engine

import (
	"testing"
)

// TestNewAchievementUI tests UI creation.
func TestNewAchievementUI(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	if ui == nil {
		t.Fatal("NewAchievementUI returned nil")
	}
	if ui.world != world {
		t.Error("World not set correctly")
	}
	if ui.screenWidth != 1280 || ui.screenHeight != 720 {
		t.Error("Screen dimensions not set correctly")
	}
	if ui.visible {
		t.Error("UI should start hidden")
	}
	if ui.selectedCategory != AchievementCategoryCombat {
		t.Error("Default category should be Combat")
	}
	if ui.selectedIndex != 0 {
		t.Error("Default selected index should be 0")
	}
	if ui.touchHandler == nil {
		t.Error("Touch handler should be initialized")
	}
	if ui.closeButton == nil {
		t.Error("Close button should be initialized")
	}
}

// TestAchievementUIVisibility tests show/hide/toggle functionality.
func TestAchievementUIVisibility(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("UI should start hidden")
	}
	if ui.IsActive() {
		t.Error("UI should not be active initially")
	}

	// Show
	ui.Show()
	if !ui.IsVisible() {
		t.Error("UI should be visible after Show()")
	}
	if !ui.IsActive() {
		t.Error("UI should be active after Show()")
	}

	// Hide
	ui.Hide()
	if ui.IsVisible() {
		t.Error("UI should be hidden after Hide()")
	}

	// Toggle on
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("UI should be visible after Toggle()")
	}

	// Toggle off
	ui.Toggle()
	if ui.IsVisible() {
		t.Error("UI should be hidden after second Toggle()")
	}

	// SetActive
	ui.SetActive(true)
	if !ui.IsActive() {
		t.Error("UI should be active after SetActive(true)")
	}
	ui.SetActive(false)
	if ui.IsActive() {
		t.Error("UI should not be active after SetActive(false)")
	}
}

// TestAchievementUISetPlayerEntity tests entity assignment.
func TestAchievementUISetPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)
	player := world.CreateEntity()

	ui.SetPlayerEntity(player)
	if ui.playerEntity != player {
		t.Error("Player entity not set correctly")
	}
}

// TestAchievementUICategoryNavigation tests category switching.
func TestAchievementUICategoryNavigation(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	// Initial category
	if ui.selectedCategory != AchievementCategoryCombat {
		t.Error("Initial category should be Combat")
	}

	// Next category
	ui.nextCategory()
	if ui.selectedCategory != AchievementCategoryQuest {
		t.Errorf("Expected Quest category, got %v", ui.selectedCategory)
	}

	// Continue through all categories
	ui.nextCategory() // Crafting
	ui.nextCategory() // Exploration
	ui.nextCategory() // Social
	ui.nextCategory() // PvP
	if ui.selectedCategory != AchievementCategoryPvP {
		t.Errorf("Expected PvP category, got %v", ui.selectedCategory)
	}

	// Wrap around to Combat
	ui.nextCategory()
	if ui.selectedCategory != AchievementCategoryCombat {
		t.Errorf("Expected Combat category after wrap, got %v", ui.selectedCategory)
	}

	// Previous category (should wrap to PvP)
	ui.previousCategory()
	if ui.selectedCategory != AchievementCategoryPvP {
		t.Errorf("Expected PvP category, got %v", ui.selectedCategory)
	}

	// Go back to Social
	ui.previousCategory()
	if ui.selectedCategory != AchievementCategorySocial {
		t.Errorf("Expected Social category, got %v", ui.selectedCategory)
	}
}

// TestAchievementUIAchievementNavigation tests achievement selection.
func TestAchievementUIAchievementNavigation(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	// Initial index
	if ui.selectedIndex != 0 {
		t.Error("Initial selected index should be 0")
	}

	// Can't go previous from 0
	ui.previousAchievement()
	if ui.selectedIndex != 0 {
		t.Error("Selected index should stay at 0")
	}

	// Next achievement
	ui.nextAchievement()
	if ui.selectedIndex != 1 {
		t.Errorf("Expected index 1, got %d", ui.selectedIndex)
	}

	// More next
	ui.nextAchievement()
	ui.nextAchievement()
	if ui.selectedIndex != 3 {
		t.Errorf("Expected index 3, got %d", ui.selectedIndex)
	}

	// Previous
	ui.previousAchievement()
	if ui.selectedIndex != 2 {
		t.Errorf("Expected index 2, got %d", ui.selectedIndex)
	}

	// Category change resets index
	ui.nextCategory()
	if ui.selectedIndex != 0 {
		t.Error("Category change should reset selected index")
	}
}

// TestAchievementUIScrollOffset tests scroll offset handling.
func TestAchievementUIScrollOffset(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	// Initial scroll offset
	if ui.scrollOffset != 0 {
		t.Error("Initial scroll offset should be 0")
	}

	// Set scroll offset manually
	ui.scrollOffset = 100

	// Switching category should reset scroll
	ui.nextCategory()
	if ui.scrollOffset != 0 {
		t.Error("Scroll offset should reset on category change")
	}
}

// TestAchievementUIGetPlayerAchievements tests component retrieval.
func TestAchievementUIGetPlayerAchievements(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)

	// No player entity
	if ui.getPlayerAchievements() != nil {
		t.Error("Should return nil with no player entity")
	}

	// Player without achievement component
	player := world.CreateEntity()
	ui.SetPlayerEntity(player)
	if ui.getPlayerAchievements() != nil {
		t.Error("Should return nil without achievement component")
	}

	// Player with achievement component
	achieveComp := NewExtendedAchievementComponent()
	player.AddComponent(achieveComp)
	result := ui.getPlayerAchievements()
	if result == nil {
		t.Error("Should return achievement component")
	}
	if result != achieveComp {
		t.Error("Returned wrong component")
	}
}

// TestAchievementUIPanelDimensions tests panel size calculation.
func TestAchievementUIPanelDimensions(t *testing.T) {
	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
		wantWidth    int
		wantHeight   int
	}{
		{"Large screen", 1920, 1080, 800, 550},
		{"Medium screen", 900, 700, 800, 550},
		{"Small width", 600, 700, 560, 550},
		{"Small height", 900, 400, 800, 360},
		{"Small both", 400, 300, 360, 260},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := NewAchievementUI(nil, tt.screenWidth, tt.screenHeight)
			_, _, width, height := ui.calculatePanelDimensions()

			if width != tt.wantWidth {
				t.Errorf("Width: got %d, want %d", width, tt.wantWidth)
			}
			if height != tt.wantHeight {
				t.Errorf("Height: got %d, want %d", height, tt.wantHeight)
			}
		})
	}
}

// TestGetTierColor tests tier color function.
func TestGetTierColor(t *testing.T) {
	tiers := []AchievementTier{
		AchievementTierNone,
		AchievementTierBronze,
		AchievementTierSilver,
		AchievementTierGold,
		AchievementTierPlatinum,
	}

	for _, tier := range tiers {
		result := getTierColor(tier)
		if result == nil {
			t.Errorf("getTierColor(%v) returned nil", tier)
		}
	}
}

// TestGetTierInitial tests tier initial function.
func TestGetTierInitial(t *testing.T) {
	tests := []struct {
		tier   AchievementTier
		expect string
	}{
		{AchievementTierNone, "-"},
		{AchievementTierBronze, "B"},
		{AchievementTierSilver, "S"},
		{AchievementTierGold, "G"},
		{AchievementTierPlatinum, "P"},
		{AchievementTier(99), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			result := getTierInitial(tt.tier)
			if result != tt.expect {
				t.Errorf("Got %q, want %q", result, tt.expect)
			}
		})
	}
}

// TestGetAchievementCategoryAbbrev tests category abbreviation function.
func TestGetAchievementCategoryAbbrev(t *testing.T) {
	tests := []struct {
		category AchievementCategory
		expect   string
	}{
		{AchievementCategoryCombat, "CMB"},
		{AchievementCategoryQuest, "QST"},
		{AchievementCategoryCrafting, "CRF"},
		{AchievementCategoryExploration, "EXP"},
		{AchievementCategorySocial, "SOC"},
		{AchievementCategoryPvP, "PVP"},
		{AchievementCategory(99), "???"},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			result := getAchievementCategoryAbbrev(tt.category)
			if result != tt.expect {
				t.Errorf("Got %q, want %q", result, tt.expect)
			}
		})
	}
}

// TestAchievementUIUpdate tests update cycle.
func TestAchievementUIUpdate(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)
	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	// Should not panic when called
	ui.Update(nil, 0.016)
	ui.Show()
	ui.Update(nil, 0.016)
}

// TestAchievementUIInterface tests UISystem interface compliance.
func TestAchievementUIInterface(t *testing.T) {
	var _ UISystem = (*AchievementUI)(nil)
}

// TestAchievementUIWithProgress tests UI with actual achievement progress.
func TestAchievementUIWithProgress(t *testing.T) {
	world := NewWorld()
	ui := NewAchievementUI(world, 1280, 720)
	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	// Add achievement component with some progress
	achieveComp := NewExtendedAchievementComponent()
	player.AddComponent(achieveComp)

	// Set some progress
	achieveComp.SetProgress("combat_first_blood", AchievementCategoryCombat, 50, [4]int64{1, 10, 100, 1000}, 1000)

	// Verify we can retrieve it
	result := ui.getPlayerAchievements()
	if result == nil {
		t.Fatal("Should return achievement component")
	}

	progress := result.GetProgress("combat_first_blood")
	if progress != 50 {
		t.Errorf("Expected progress 50, got %d", progress)
	}

	tier := result.GetTier("combat_first_blood")
	if tier != AchievementTierSilver {
		t.Errorf("Expected Silver tier, got %v", tier)
	}
}
