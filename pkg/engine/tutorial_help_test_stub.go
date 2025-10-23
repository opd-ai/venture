//go:build test
// +build test

// Package engine provides test stubs for Tutorial and Help systems.
package engine

// TutorialSystem manages tutorial steps (test stub).
type TutorialSystem struct {
	Enabled bool
	ShowUI  bool
}

// NewTutorialSystem creates a new tutorial system (test stub).
func NewTutorialSystem() *TutorialSystem {
	return &TutorialSystem{
		Enabled: false,
		ShowUI:  false,
	}
}

// Update processes tutorial logic (test stub).
func (t *TutorialSystem) Update(entities []*Entity, deltaTime float64) {
	// Stub - no op in tests
}

// Draw renders the tutorial UI (test stub).
func (t *TutorialSystem) Draw(screen interface{}) {
	// Stub - no op in tests
}

// Skip skips the current tutorial step (test stub).
func (t *TutorialSystem) Skip() {
	// Stub - no op in tests
}

// ShowNotification displays a notification message (test stub).
func (t *TutorialSystem) ShowNotification(message string, duration float64) {
	// Stub - no op in tests
}

// HelpSystem manages context-sensitive help (test stub).
type HelpSystem struct {
	Visible bool
}

// NewHelpSystem creates a new help system (test stub).
func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		Visible: false,
	}
}

// Toggle toggles help visibility (test stub).
func (h *HelpSystem) Toggle() {
	h.Visible = !h.Visible
}

// Draw renders the help UI (test stub).
func (h *HelpSystem) Draw(screen interface{}) {
	// Stub - no op in tests
}

// ShowTopic shows a specific help topic (test stub).
func (h *HelpSystem) ShowTopic(topicID string) {
	// Stub - no op in tests
}
