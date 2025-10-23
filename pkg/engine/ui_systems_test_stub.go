//go:build test
// +build test

// Package engine provides test stubs for UI systems.
package engine

// InventoryUI manages the inventory interface (test stub).
type InventoryUI struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
}

// NewInventoryUI creates a new inventory UI (test stub).
func NewInventoryUI(world *World, screenWidth, screenHeight int) *InventoryUI {
	return &InventoryUI{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity (test stub).
func (ui *InventoryUI) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// SetInventorySystem connects the inventory system (test stub).
func (ui *InventoryUI) SetInventorySystem(system *InventorySystem) {
	// Stub - no op in tests
}

// Toggle toggles inventory visibility (test stub).
func (ui *InventoryUI) Toggle() {
	// Stub - no op in tests
}

// Update updates the inventory UI (test stub).
func (ui *InventoryUI) Update() {
	// Stub - no op in tests
}

// Draw renders the inventory UI (test stub).
func (ui *InventoryUI) Draw(screen interface{}) {
	// Stub - no op in tests
}

// IsVisible returns whether inventory is visible (test stub).
func (ui *InventoryUI) IsVisible() bool {
	return false
}

// QuestUI manages the quest log interface (test stub).
type QuestUI struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
}

// NewQuestUI creates a new quest UI (test stub).
func NewQuestUI(world *World, screenWidth, screenHeight int) *QuestUI {
	return &QuestUI{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity (test stub).
func (ui *QuestUI) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// Toggle toggles quest UI visibility (test stub).
func (ui *QuestUI) Toggle() {
	// Stub - no op in tests
}

// Update updates the quest UI (test stub).
func (ui *QuestUI) Update() {
	// Stub - no op in tests
}

// Draw renders the quest UI (test stub).
func (ui *QuestUI) Draw(screen interface{}) {
	// Stub - no op in tests
}

// IsVisible returns whether quest UI is visible (test stub).
func (ui *QuestUI) IsVisible() bool {
	return false
}
