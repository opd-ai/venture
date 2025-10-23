//go:build test
// +build test

// Package engine provides test stubs for InputSystem.
package engine

// InputSystem processes input (test stub).
type InputSystem struct {
	MoveSpeed float64
}

// NewInputSystem creates a new input system (test stub).
func NewInputSystem() *InputSystem {
	return &InputSystem{
		MoveSpeed: 100.0,
	}
}

// Update processes input for all entities (test stub).
func (s *InputSystem) Update(entities []*Entity, deltaTime float64) {
	// Stub - no op in tests
}

// SetHelpSystem connects the help system (test stub).
func (s *InputSystem) SetHelpSystem(helpSystem *HelpSystem) {
	// Stub - no op in tests
}

// SetTutorialSystem connects the tutorial system (test stub).
func (s *InputSystem) SetTutorialSystem(tutorialSystem *TutorialSystem) {
	// Stub - no op in tests
}

// SetQuickSaveCallback sets the callback function for quick save (test stub).
func (s *InputSystem) SetQuickSaveCallback(callback func() error) {
	// Stub - no op in tests
}

// SetQuickLoadCallback sets the callback function for quick load (test stub).
func (s *InputSystem) SetQuickLoadCallback(callback func() error) {
	// Stub - no op in tests
}

// SetInventoryCallback sets the callback function for opening inventory (test stub).
func (s *InputSystem) SetInventoryCallback(callback func()) {
	// Stub - no op in tests
}

// SetCharacterCallback sets the callback function for opening character screen (test stub).
func (s *InputSystem) SetCharacterCallback(callback func()) {
	// Stub - no op in tests
}

// SetSkillsCallback sets the callback function for opening skills screen (test stub).
func (s *InputSystem) SetSkillsCallback(callback func()) {
	// Stub - no op in tests
}

// SetQuestsCallback sets the callback function for opening quest log (test stub).
func (s *InputSystem) SetQuestsCallback(callback func()) {
	// Stub - no op in tests
}

// SetMapCallback sets the callback function for opening map (test stub).
func (s *InputSystem) SetMapCallback(callback func()) {
	// Stub - no op in tests
}

// SetMenuToggleCallback sets the callback function for toggling pause menu (test stub).
func (s *InputSystem) SetMenuToggleCallback(callback func()) {
	// Stub - no op in tests
}
