// Package engine provides tests for the statistics UI system.
//
// Phase 86: Statistics UI (V15.0)
package engine

import (
	"testing"
)

// TestNewStatisticsUI tests UI creation.
func TestNewStatisticsUI(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)

	if ui == nil {
		t.Fatal("NewStatisticsUI returned nil")
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
	if ui.selectedCategory != StatCategoryCombat {
		t.Error("Default category should be Combat")
	}
	if ui.showingSession {
		t.Error("Default mode should be Lifetime")
	}
	if ui.touchHandler == nil {
		t.Error("Touch handler should be initialized")
	}
	if ui.closeButton == nil {
		t.Error("Close button should be initialized")
	}
}

// TestStatisticsUIVisibility tests show/hide/toggle functionality.
func TestStatisticsUIVisibility(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)

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

// TestStatisticsUISetPlayerEntity tests entity assignment.
func TestStatisticsUISetPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)
	player := world.CreateEntity()

	ui.SetPlayerEntity(player)
	if ui.playerEntity != player {
		t.Error("Player entity not set correctly")
	}
}

// TestStatisticsUICategoryNavigation tests category switching.
func TestStatisticsUICategoryNavigation(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)

	// Initial category
	if ui.selectedCategory != StatCategoryCombat {
		t.Error("Initial category should be Combat")
	}

	// Next category
	ui.nextCategory()
	if ui.selectedCategory != StatCategoryQuest {
		t.Errorf("Expected Quest category, got %v", ui.selectedCategory)
	}

	// Continue through all categories
	ui.nextCategory() // Crafting
	ui.nextCategory() // Exploration
	ui.nextCategory() // Social
	ui.nextCategory() // PvP
	ui.nextCategory() // Economy
	ui.nextCategory() // General
	if ui.selectedCategory != StatCategoryGeneral {
		t.Errorf("Expected General category, got %v", ui.selectedCategory)
	}

	// Wrap around to Combat
	ui.nextCategory()
	if ui.selectedCategory != StatCategoryCombat {
		t.Errorf("Expected Combat category after wrap, got %v", ui.selectedCategory)
	}

	// Previous category (should wrap to General)
	ui.previousCategory()
	if ui.selectedCategory != StatCategoryGeneral {
		t.Errorf("Expected General category, got %v", ui.selectedCategory)
	}

	// Go back to PvP
	ui.previousCategory()
	if ui.selectedCategory != StatCategoryEconomy {
		t.Errorf("Expected Economy category, got %v", ui.selectedCategory)
	}
}

// TestStatisticsUIScrollOffset tests scroll offset handling.
func TestStatisticsUIScrollOffset(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)

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

// TestStatisticsUIGetPlayerStatistics tests component retrieval.
func TestStatisticsUIGetPlayerStatistics(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)

	// No player entity
	if ui.getPlayerStatistics() != nil {
		t.Error("Should return nil with no player entity")
	}

	// Player without statistics component
	player := world.CreateEntity()
	ui.SetPlayerEntity(player)
	if ui.getPlayerStatistics() != nil {
		t.Error("Should return nil without statistics component")
	}

	// Player with statistics component
	statsComp := NewPlayerStatisticsComponent()
	player.AddComponent(statsComp)
	result := ui.getPlayerStatistics()
	if result == nil {
		t.Error("Should return statistics component")
	}
	if result != statsComp {
		t.Error("Returned wrong component")
	}
}

// TestStatisticsUIPanelDimensions tests panel size calculation.
func TestStatisticsUIPanelDimensions(t *testing.T) {
	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
		wantWidth    int
		wantHeight   int
	}{
		{"Large screen", 1920, 1080, 700, 500},
		{"Medium screen", 800, 600, 700, 500},
		{"Small width", 600, 600, 560, 500},
		{"Small height", 800, 400, 700, 360},
		{"Small both", 400, 300, 360, 260},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := NewStatisticsUI(nil, tt.screenWidth, tt.screenHeight)
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

// TestFormatStatDisplayValue tests stat value formatting.
func TestFormatStatDisplayValue(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		def    StatDefinition
		expect string
	}{
		{"Zero", 0, StatDefinition{IsTime: false}, "0"},
		{"Small number", 42, StatDefinition{IsTime: false}, "42"},
		{"Thousands", 1500, StatDefinition{IsTime: false}, "1.5K"},
		{"Millions", 2500000, StatDefinition{IsTime: false}, "2.5M"},
		{"Time seconds", 45, StatDefinition{IsTime: true}, "45s"},
		{"Time minutes", 125, StatDefinition{IsTime: true}, "2m 5s"},
		{"Time hours", 3725, StatDefinition{IsTime: true}, "1h 2m 5s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatDisplayValue(tt.value, tt.def)
			if result != tt.expect {
				t.Errorf("Got %q, want %q", result, tt.expect)
			}
		})
	}
}

// TestGetCategoryAbbrev tests category abbreviation function.
func TestGetCategoryAbbrev(t *testing.T) {
	tests := []struct {
		category StatCategory
		expect   string
	}{
		{StatCategoryCombat, "CMB"},
		{StatCategoryQuest, "QST"},
		{StatCategoryCrafting, "CRF"},
		{StatCategoryExploration, "EXP"},
		{StatCategorySocial, "SOC"},
		{StatCategoryPvP, "PVP"},
		{StatCategoryEconomy, "ECO"},
		{StatCategoryGeneral, "GEN"},
		{StatCategory(99), "???"},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			result := getCategoryAbbrev(tt.category)
			if result != tt.expect {
				t.Errorf("Got %q, want %q", result, tt.expect)
			}
		})
	}
}

// TestStatisticsUIUpdate tests update cycle.
func TestStatisticsUIUpdate(t *testing.T) {
	world := NewWorld()
	ui := NewStatisticsUI(world, 1280, 720)
	player := world.CreateEntity()
	ui.SetPlayerEntity(player)

	// Should not panic when called
	ui.Update(nil, 0.016)
	ui.Show()
	ui.Update(nil, 0.016)
}

// TestStatisticsUIInterface tests UISystem interface compliance.
func TestStatisticsUIInterface(t *testing.T) {
	var _ UISystem = (*StatisticsUI)(nil)
}
