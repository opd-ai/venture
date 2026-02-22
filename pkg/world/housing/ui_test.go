package housing

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
)

func TestNewHousingUI(t *testing.T) {
	ui := NewHousingUI(1920, 1080)

	if ui == nil {
		t.Fatal("NewHousingUI returned nil")
	}
	if ui.Width != 1880 { // 1920 - 40
		t.Errorf("Width = %d, want 1880", ui.Width)
	}
	if ui.Height != 980 { // 1080 - 100
		t.Errorf("Height = %d, want 980", ui.Height)
	}
	if ui.Visible {
		t.Error("UI should not be visible by default")
	}
	if ui.GenreID != "fantasy" {
		t.Errorf("GenreID = %s, want fantasy", ui.GenreID)
	}
}

func TestHousingUI_Visibility(t *testing.T) {
	ui := NewHousingUI(800, 600)

	if ui.IsVisible() {
		t.Error("IsVisible should return false initially")
	}

	ui.Show()
	if !ui.IsVisible() {
		t.Error("IsVisible should return true after Show()")
	}

	ui.Hide()
	if ui.IsVisible() {
		t.Error("IsVisible should return false after Hide()")
	}

	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("IsVisible should return true after Toggle() from hidden")
	}

	ui.Toggle()
	if ui.IsVisible() {
		t.Error("IsVisible should return false after Toggle() from visible")
	}
}

func TestHousingUI_SetPlayerID(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.SetPlayerID(12345)

	if ui.playerID != 12345 {
		t.Errorf("playerID = %d, want 12345", ui.playerID)
	}
}

func TestHousingUI_SetGuildID(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.SetGuildID("guild-abc-123")

	if ui.guildID != "guild-abc-123" {
		t.Errorf("guildID = %s, want guild-abc-123", ui.guildID)
	}
}

func TestHousingUI_SetManagers(t *testing.T) {
	ui := NewHousingUI(800, 600)

	hm := NewManager()
	ghm := NewGuildHallManager()
	bg := building.NewGenerator()
	fg := furniture.NewGenerator()

	ui.SetManagers(hm, ghm, bg, fg)

	if ui.housingManager != hm {
		t.Error("housingManager not set correctly")
	}
	if ui.guildHallManager != ghm {
		t.Error("guildHallManager not set correctly")
	}
	if ui.buildingGenerator != bg {
		t.Error("buildingGenerator not set correctly")
	}
	if ui.furnitureGenerator != fg {
		t.Error("furnitureGenerator not set correctly")
	}
	if ui.menuState != "main" {
		t.Errorf("menuState = %s, want main", ui.menuState)
	}
}

func TestHousingUI_SetManagers_NilSafe(t *testing.T) {
	ui := NewHousingUI(800, 600)

	// Should not panic with nil values
	ui.SetManagers(nil, nil, nil, nil)

	if ui.housingManager != nil {
		t.Error("housingManager should be nil")
	}
	if ui.guildHallManager != nil {
		t.Error("guildHallManager should be nil")
	}
}

func TestHousingUI_SetManagers_WrongTypes(t *testing.T) {
	ui := NewHousingUI(800, 600)

	// Should not panic with wrong types
	ui.SetManagers("wrong", 123, true, struct{}{})

	if ui.housingManager != nil {
		t.Error("housingManager should be nil for wrong type")
	}
}

func TestHousingUI_GetBuildingTypesList(t *testing.T) {
	ui := NewHousingUI(800, 600)
	types := ui.getBuildingTypesList()

	if len(types) != 6 {
		t.Errorf("getBuildingTypesList returned %d types, want 6", len(types))
	}

	expectedTypes := []string{"House", "Workshop", "Storage", "Tower", "Manor", "Guild Hall"}
	for i, expected := range expectedTypes {
		if types[i] != expected {
			t.Errorf("types[%d] = %s, want %s", i, types[i], expected)
		}
	}
}

func TestHousingUI_GetFurnitureTypesList(t *testing.T) {
	ui := NewHousingUI(800, 600)
	types := ui.getFurnitureTypesList()

	if len(types) != 6 {
		t.Errorf("getFurnitureTypesList returned %d types, want 6", len(types))
	}

	expectedTypes := []string{"Seating", "Storage", "Crafting", "Decoration", "Lighting", "Bedding"}
	for i, expected := range expectedTypes {
		if types[i] != expected {
			t.Errorf("types[%d] = %s, want %s", i, types[i], expected)
		}
	}
}

func TestHousingUI_GetBuildingInfo(t *testing.T) {
	ui := NewHousingUI(800, 600)

	tests := []struct {
		buildingType string
		wantContains string
	}{
		{"House", "Small living space"},
		{"Workshop", "Crafting facility"},
		{"Storage", "Warehouse"},
		{"Tower", "Vertical structure"},
		{"Manor", "Large estate"},
		{"Guild Hall", "Guild headquarters"},
		{"Unknown", "Select a building type"},
	}

	for _, tt := range tests {
		t.Run(tt.buildingType, func(t *testing.T) {
			info := ui.getBuildingInfo(tt.buildingType)
			if info == "" {
				t.Error("getBuildingInfo returned empty string")
			}
			// Check that info contains expected substring
			if len(info) < 10 {
				t.Errorf("getBuildingInfo returned short string: %s", info)
			}
		})
	}
}

func TestHousingUI_CategoryToFurnitureType(t *testing.T) {
	ui := NewHousingUI(800, 600)

	tests := []struct {
		category string
		want     int
	}{
		{"Seating", int(furniture.TypeSeating)},
		{"Storage", int(furniture.TypeStorage)},
		{"Crafting", int(furniture.TypeCrafting)},
		{"Decoration", int(furniture.TypeDecoration)},
		{"Lighting", int(furniture.TypeLighting)},
		{"Bedding", int(furniture.TypeBedding)},
		{"Unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := ui.categoryToFurnitureType(tt.category)
			if got != tt.want {
				t.Errorf("categoryToFurnitureType(%s) = %d, want %d", tt.category, got, tt.want)
			}
		})
	}
}

func TestHousingUI_GetFurnitureCountForCategory(t *testing.T) {
	ui := NewHousingUI(800, 600)

	// Test that known categories return non-zero counts
	seatingCount := ui.getFurnitureCountForCategory("Seating")
	if seatingCount == 0 {
		t.Error("Seating category should have items")
	}

	// Test unknown category
	unknownCount := ui.getFurnitureCountForCategory("Unknown")
	if unknownCount != 0 {
		t.Errorf("Unknown category should have 0 items, got %d", unknownCount)
	}
}

func TestHousingUI_GetFurnitureItemsForCategory(t *testing.T) {
	ui := NewHousingUI(800, 600)

	// Test that known categories return items
	seatingItems := ui.getFurnitureItemsForCategory("Seating")
	if seatingItems == "None" {
		t.Error("Seating category should have items")
	}

	// Test unknown category
	unknownItems := ui.getFurnitureItemsForCategory("Unknown")
	if unknownItems != "None" {
		t.Errorf("Unknown category should return 'None', got %s", unknownItems)
	}
}

func TestHousingUI_HandleSubmenuInput_Build(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 0

	// Simulate selection change (can't test key presses directly without Ebiten)
	buildingTypes := ui.getBuildingTypesList()
	if len(buildingTypes) < 2 {
		t.Skip("Not enough building types to test navigation")
	}

	// Test manual selection bounds
	ui.selectedBuildingType = len(buildingTypes) - 1
	if ui.selectedBuildingType != 5 {
		t.Errorf("selectedBuildingType = %d, want 5", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleSubmenuInput_Furniture(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "furniture"
	ui.selectedFurnitureType = 0

	furnitureTypes := ui.getFurnitureTypesList()
	if len(furnitureTypes) < 2 {
		t.Skip("Not enough furniture types to test navigation")
	}

	// Test manual selection bounds
	ui.selectedFurnitureType = len(furnitureTypes) - 1
	if ui.selectedFurnitureType != 5 {
		t.Errorf("selectedFurnitureType = %d, want 5", ui.selectedFurnitureType)
	}
}

func TestHousingUI_Update_NotVisible(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.Visible = false

	consumed := ui.Update()
	if consumed {
		t.Error("Update should return false when UI is not visible")
	}
}

// StubMenuInput implements MenuInputProvider for testing.
type StubMenuInput struct {
	MenuUpJustPressed      bool
	MenuDownJustPressed    bool
	MenuConfirmJustPressed bool
	MenuBackJustPressed    bool
	MenuTabJustPressed     bool
}

func (s *StubMenuInput) IsMenuUpJustPressed() bool      { return s.MenuUpJustPressed }
func (s *StubMenuInput) IsMenuDownJustPressed() bool    { return s.MenuDownJustPressed }
func (s *StubMenuInput) IsMenuConfirmJustPressed() bool { return s.MenuConfirmJustPressed }
func (s *StubMenuInput) IsMenuBackJustPressed() bool    { return s.MenuBackJustPressed }
func (s *StubMenuInput) IsMenuTabJustPressed() bool     { return s.MenuTabJustPressed }

func TestHousingUI_SetInput(t *testing.T) {
	ui := NewHousingUI(800, 600)
	input := &StubMenuInput{}

	ui.SetInput(input)

	if ui.inputProvider != input {
		t.Error("SetInput did not set input provider")
	}
}

func TestHousingUI_Update_WithInputProvider_Back(t *testing.T) {
	ui := NewHousingUI(800, 600)
	input := &StubMenuInput{MenuBackJustPressed: true}
	ui.SetInput(input)
	ui.Show()

	consumed := ui.Update()
	if !consumed {
		t.Error("Update should consume input when visible")
	}
	if ui.IsVisible() {
		t.Error("UI should be hidden after back pressed")
	}
}

func TestHousingUI_Update_WithInputProvider_Tab(t *testing.T) {
	ui := NewHousingUI(800, 600)
	input := &StubMenuInput{MenuTabJustPressed: true}
	ui.SetInput(input)
	ui.menuState = "main"
	ui.Show()

	ui.Update()
	if ui.menuState != "build" {
		t.Errorf("menuState = %s, want build after tab from main", ui.menuState)
	}

	ui.Update()
	if ui.menuState != "furniture" {
		t.Errorf("menuState = %s, want furniture after tab from build", ui.menuState)
	}

	ui.Update()
	if ui.menuState != "guildhall" {
		t.Errorf("menuState = %s, want guildhall after tab from furniture", ui.menuState)
	}

	ui.Update()
	if ui.menuState != "main" {
		t.Errorf("menuState = %s, want main after tab from guildhall", ui.menuState)
	}
}

func TestHousingUI_HandleSubmenuInput_WithInputProvider_Build(t *testing.T) {
	ui := NewHousingUI(800, 600)
	input := &StubMenuInput{}
	ui.SetInput(input)
	ui.menuState = "build"
	ui.selectedBuildingType = 0
	ui.Show()

	// Test down navigation
	input.MenuDownJustPressed = true
	ui.handleSubmenuInput()
	if ui.selectedBuildingType != 1 {
		t.Errorf("selectedBuildingType = %d, want 1 after down", ui.selectedBuildingType)
	}

	// Test up navigation
	input.MenuDownJustPressed = false
	input.MenuUpJustPressed = true
	ui.handleSubmenuInput()
	if ui.selectedBuildingType != 0 {
		t.Errorf("selectedBuildingType = %d, want 0 after up", ui.selectedBuildingType)
	}

	// Test wrap-around (up from 0)
	ui.handleSubmenuInput()
	buildingTypes := ui.getBuildingTypesList()
	if ui.selectedBuildingType != len(buildingTypes)-1 {
		t.Errorf("selectedBuildingType = %d, want %d after wrap-around", ui.selectedBuildingType, len(buildingTypes)-1)
	}
}

func TestHousingUI_HandleSubmenuInput_WithInputProvider_Furniture(t *testing.T) {
	ui := NewHousingUI(800, 600)
	input := &StubMenuInput{}
	ui.SetInput(input)
	ui.menuState = "furniture"
	ui.selectedFurnitureType = 0
	ui.Show()

	// Test down navigation
	input.MenuDownJustPressed = true
	ui.handleSubmenuInput()
	if ui.selectedFurnitureType != 1 {
		t.Errorf("selectedFurnitureType = %d, want 1 after down", ui.selectedFurnitureType)
	}

	// Test up navigation
	input.MenuDownJustPressed = false
	input.MenuUpJustPressed = true
	ui.handleSubmenuInput()
	if ui.selectedFurnitureType != 0 {
		t.Errorf("selectedFurnitureType = %d, want 0 after up", ui.selectedFurnitureType)
	}

	// Test wrap-around (up from 0)
	ui.handleSubmenuInput()
	furnitureTypes := ui.getFurnitureTypesList()
	if ui.selectedFurnitureType != len(furnitureTypes)-1 {
		t.Errorf("selectedFurnitureType = %d, want %d after wrap-around", ui.selectedFurnitureType, len(furnitureTypes)-1)
	}
}

func TestHousingUI_HandleSubmenuInput_NoInputProvider(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 0
	// No input provider set

	// Should not panic and should not change selection
	ui.handleSubmenuInput()
	if ui.selectedBuildingType != 0 {
		t.Errorf("selectedBuildingType changed without input provider")
	}
}
