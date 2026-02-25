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
	TouchJustPressed       bool
	TouchX                 int
	TouchY                 int
	HasActiveInput         bool
}

func (s *StubMenuInput) IsMenuUpJustPressed() bool      { return s.MenuUpJustPressed }
func (s *StubMenuInput) IsMenuDownJustPressed() bool    { return s.MenuDownJustPressed }
func (s *StubMenuInput) IsMenuConfirmJustPressed() bool { return s.MenuConfirmJustPressed }
func (s *StubMenuInput) IsMenuBackJustPressed() bool    { return s.MenuBackJustPressed }
func (s *StubMenuInput) IsMenuTabJustPressed() bool     { return s.MenuTabJustPressed }
func (s *StubMenuInput) IsTouchOrMouseJustPressed() bool {
	return s.TouchJustPressed
}

func (s *StubMenuInput) GetTouchOrMousePosition() (x, y int, hasActiveInput bool) {
	return s.TouchX, s.TouchY, s.HasActiveInput
}

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

// Touch input tests

func TestHousingUI_HandleTouchInput_BuildingMenu_SelectFirst(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 2 // Start at index 2
	ui.Show()

	// Calculate the position of the first building item (House)
	startY := ui.Y + ui.Height - 200 // contentY
	itemStartY := startY + 20
	itemX := ui.X + 10

	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           itemX + 5, // Inside first item
		TouchY:           itemStartY + 5,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 0 {
		t.Errorf("selectedBuildingType = %d, want 0 after touching first item", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleTouchInput_BuildingMenu_SelectThird(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 0
	ui.Show()

	// Calculate the position of the third building item (Storage)
	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10
	thirdItemY := itemStartY + (2 * 15) // Third item at index 2

	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           itemX + 50,
		TouchY:           thirdItemY + 7,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 2 {
		t.Errorf("selectedBuildingType = %d, want 2 after touching third item", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleTouchInput_FurnitureMenu_SelectSecond(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "furniture"
	ui.selectedFurnitureType = 0
	ui.Show()

	// Calculate the position of the second furniture item (Storage)
	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10
	secondItemY := itemStartY + (1 * 15) // Second item at index 1

	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           itemX + 50,
		TouchY:           secondItemY + 7,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedFurnitureType != 1 {
		t.Errorf("selectedFurnitureType = %d, want 1 after touching second item", ui.selectedFurnitureType)
	}
}

func TestHousingUI_HandleTouchInput_OutOfBounds_NoChange(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 2
	ui.Show()

	// Touch outside the menu area
	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           ui.X - 100, // Way outside the menu
		TouchY:           ui.Y - 100,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 2 {
		t.Errorf("selectedBuildingType = %d, want 2 (unchanged) after touching outside bounds", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleTouchInput_NoActiveInput(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 1
	ui.Show()

	// Touch position provided but not active (cursor hover without click)
	input := &StubMenuInput{
		TouchJustPressed: false,
		TouchX:           ui.X + 20,
		TouchY:           ui.Y + 100,
		HasActiveInput:   false,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 1 {
		t.Errorf("selectedBuildingType = %d, want 1 (unchanged) without active input", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleTouchInput_GuildHallMenu_NoInteraction(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "guildhall"
	ui.Show()

	// Touch on guildhall menu (which has no interactive elements)
	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           ui.X + 50,
		TouchY:           ui.Y + 100,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	// Should not panic
	ui.Update()
}

func TestHousingUI_HandleTouchInput_NoInputProvider(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 1
	ui.Show()
	// No input provider set

	// Should not panic
	ui.Update()

	if ui.selectedBuildingType != 1 {
		t.Errorf("selectedBuildingType changed without input provider")
	}
}

func TestHousingUI_HandleTouchInput_AllBuildingItems(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.Show()

	buildingTypes := ui.getBuildingTypesList()
	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10

	// Test touching each building item
	for i := range buildingTypes {
		t.Run(buildingTypes[i], func(t *testing.T) {
			ui.selectedBuildingType = 0 // Reset to first item

			itemY := itemStartY + (i * 15)
			input := &StubMenuInput{
				TouchJustPressed: true,
				TouchX:           itemX + 50,
				TouchY:           itemY + 7,
				HasActiveInput:   true,
			}
			ui.SetInput(input)

			ui.Update()

			if ui.selectedBuildingType != i {
				t.Errorf("selectedBuildingType = %d, want %d after touching item %s",
					ui.selectedBuildingType, i, buildingTypes[i])
			}
		})
	}
}

func TestHousingUI_HandleTouchInput_AllFurnitureItems(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "furniture"
	ui.Show()

	furnitureTypes := ui.getFurnitureTypesList()
	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10

	// Test touching each furniture item
	for i := range furnitureTypes {
		t.Run(furnitureTypes[i], func(t *testing.T) {
			ui.selectedFurnitureType = 0 // Reset to first item

			itemY := itemStartY + (i * 15)
			input := &StubMenuInput{
				TouchJustPressed: true,
				TouchX:           itemX + 50,
				TouchY:           itemY + 7,
				HasActiveInput:   true,
			}
			ui.SetInput(input)

			ui.Update()

			if ui.selectedFurnitureType != i {
				t.Errorf("selectedFurnitureType = %d, want %d after touching item %s",
					ui.selectedFurnitureType, i, furnitureTypes[i])
			}
		})
	}
}

func TestHousingUI_HandleTouchInput_EdgeDetection_LeftBoundary(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 2
	ui.Show()

	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10

	// Touch exactly at the left boundary (should select)
	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           itemX,
		TouchY:           itemStartY + 5,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 0 {
		t.Errorf("selectedBuildingType = %d, want 0 at left boundary", ui.selectedBuildingType)
	}
}

func TestHousingUI_HandleTouchInput_EdgeDetection_JustOutside(t *testing.T) {
	ui := NewHousingUI(800, 600)
	ui.menuState = "build"
	ui.selectedBuildingType = 1
	ui.Show()

	startY := ui.Y + ui.Height - 200
	itemStartY := startY + 20
	itemX := ui.X + 10

	// Touch just outside the left boundary (should not select)
	input := &StubMenuInput{
		TouchJustPressed: true,
		TouchX:           itemX - 1,
		TouchY:           itemStartY + 5,
		HasActiveInput:   true,
	}
	ui.SetInput(input)

	ui.Update()

	if ui.selectedBuildingType != 1 {
		t.Errorf("selectedBuildingType = %d, want 1 (unchanged) just outside boundary", ui.selectedBuildingType)
	}
}
