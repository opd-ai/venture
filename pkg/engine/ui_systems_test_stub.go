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

// CharacterUI manages the character screen interface (test stub).
type CharacterUI struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
}

// NewCharacterUI creates a new character UI (test stub).
func NewCharacterUI(world *World, screenWidth, screenHeight int) *CharacterUI {
	return &CharacterUI{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity (test stub).
func (ui *CharacterUI) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// Toggle toggles character UI visibility (test stub).
func (ui *CharacterUI) Toggle() {
	// Stub - no op in tests
}

// Update updates the character UI (test stub).
func (ui *CharacterUI) Update(deltaTime float64) {
	// Stub - no op in tests
}

// Draw renders the character UI (test stub).
func (ui *CharacterUI) Draw(screen interface{}) {
	// Stub - no op in tests
}

// IsVisible returns whether character UI is visible (test stub).
func (ui *CharacterUI) IsVisible() bool {
	return false
}

// SkillsUI manages the skills screen interface (test stub).
type SkillsUI struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
}

// NewSkillsUI creates a new skills UI (test stub).
func NewSkillsUI(world *World, screenWidth, screenHeight int) *SkillsUI {
	return &SkillsUI{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity (test stub).
func (ui *SkillsUI) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// Toggle toggles skills UI visibility (test stub).
func (ui *SkillsUI) Toggle() {
	// Stub - no op in tests
}

// Update updates the skills UI (test stub).
func (ui *SkillsUI) Update(deltaTime float64) {
	// Stub - no op in tests
}

// Draw renders the skills UI (test stub).
func (ui *SkillsUI) Draw(screen interface{}) {
	// Stub - no op in tests
}

// IsVisible returns whether skills UI is visible (test stub).
func (ui *SkillsUI) IsVisible() bool {
	return false
}

// MapUI manages the map interface (test stub).
type MapUI struct {
	World        *World
	ScreenWidth  int
	ScreenHeight int
}

// NewMapUI creates a new map UI (test stub).
func NewMapUI(world *World, screenWidth, screenHeight int) *MapUI {
	return &MapUI{
		World:        world,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity (test stub).
func (ui *MapUI) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// ToggleFullScreen toggles map fullscreen mode (test stub).
func (ui *MapUI) ToggleFullScreen() {
	// Stub - no op in tests
}

// Update updates the map UI (test stub).
func (ui *MapUI) Update(deltaTime float64) {
	// Stub - no op in tests
}

// Draw renders the map UI (test stub).
func (ui *MapUI) Draw(screen interface{}) {
	// Stub - no op in tests
}

// IsFullScreen returns whether map is in fullscreen mode (test stub).
func (ui *MapUI) IsFullScreen() bool {
	return false
}
