//go:build test
// +build test

// Package engine provides test stubs for HUDSystem.
package engine

// HUDSystem renders heads-up display elements (test stub).
type HUDSystem struct {
	ScreenWidth  int
	ScreenHeight int
}

// NewHUDSystem creates a new HUD system (test stub).
func NewHUDSystem(screenWidth, screenHeight int) *HUDSystem {
	return &HUDSystem{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// SetPlayerEntity sets the player entity to display stats for (test stub).
func (h *HUDSystem) SetPlayerEntity(entity *Entity) {
	// Stub - no op in tests
}

// Draw renders HUD elements (test stub).
func (h *HUDSystem) Draw(screen interface{}) {
	// Stub - no op in tests
}
