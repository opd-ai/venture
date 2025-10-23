//go:build test
// +build test

// Package engine provides the character UI types for testing.
package engine

import "fmt"

// CharacterUI stub for testing (full implementation in character_ui.go).
type CharacterUI struct {
	visible      bool
	world        *World
	playerEntity *Entity
	screenWidth  int
	screenHeight int

	// Layout sections
	statsPanel      Rectangle
	equipmentPanel  Rectangle
	attributesPanel Rectangle
}

// Rectangle defines a UI panel bounds.
type Rectangle struct {
	X, Y, Width, Height int
}

// NewCharacterUI creates a new character UI system stub for testing.
func NewCharacterUI(world *World, screenWidth, screenHeight int) *CharacterUI {
	ui := &CharacterUI{
		visible:      false,
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
	ui.calculateLayout()
	return ui
}

// SetPlayerEntity sets the player entity whose stats to display.
func (ui *CharacterUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the character UI.
func (ui *CharacterUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the character UI is currently shown.
func (ui *CharacterUI) IsVisible() bool {
	return ui.visible
}

// Show displays the character UI.
func (ui *CharacterUI) Show() {
	ui.visible = true
}

// Hide hides the character UI.
func (ui *CharacterUI) Hide() {
	ui.visible = false
}

// Update processes input for the character UI (stub).
func (ui *CharacterUI) Update(deltaTime float64) {
	// Stub for testing
}

// calculateLayout computes panel positions based on screen size.
func (ui *CharacterUI) calculateLayout() {
	panelWidth := 800
	panelHeight := 600
	if ui.screenWidth < 800 {
		panelWidth = ui.screenWidth - 40
	}
	if ui.screenHeight < 600 {
		panelHeight = ui.screenHeight - 40
	}

	panelX := (ui.screenWidth - panelWidth) / 2
	panelY := (ui.screenHeight - panelHeight) / 2

	// Layout: Left 30%, Center 40%, Right 30%
	leftWidth := int(float64(panelWidth) * 0.30)
	centerWidth := int(float64(panelWidth) * 0.40)
	rightWidth := panelWidth - leftWidth - centerWidth

	contentY := panelY + 60 // Below title
	contentHeight := panelHeight - 100

	ui.statsPanel = Rectangle{
		X:      panelX + 10,
		Y:      contentY,
		Width:  leftWidth - 20,
		Height: contentHeight,
	}

	ui.equipmentPanel = Rectangle{
		X:      panelX + leftWidth + 10,
		Y:      contentY,
		Width:  centerWidth - 20,
		Height: contentHeight,
	}

	ui.attributesPanel = Rectangle{
		X:      panelX + leftWidth + centerWidth + 10,
		Y:      contentY,
		Width:  rightWidth - 20,
		Height: contentHeight,
	}
}

// calculateDerivedStats computes crit chance, evasion, etc. from base stats.
func (ui *CharacterUI) calculateDerivedStats(stats *StatsComponent) map[string]float64 {
	derived := make(map[string]float64)

	// Critical chance: base + bonus from attack
	derived["crit_chance"] = stats.CritChance * 100 // Convert to percentage

	// Critical damage multiplier
	derived["crit_damage"] = stats.CritDamage

	// Evasion: base + bonus from speed (if speed stat exists, otherwise just base)
	derived["evasion"] = stats.Evasion * 100 // Convert to percentage

	return derived
}

// getResistanceColor returns a color based on resistance value (stub for testing).
func (ui *CharacterUI) getResistanceColor(resist float64) interface{} {
	// Return a simple value for testing
	if resist >= 0.5 {
		return "green"
	} else if resist >= 0.25 {
		return "yellow"
	} else if resist > 0 {
		return "orange"
	}
	return "gray"
}

// formatStatValue formats a stat value for display (e.g., "42" or "42.5%")
func formatStatValue(value float64, isPercentage bool) string {
	if isPercentage {
		return fmt.Sprintf("%.1f%%", value)
	}
	// Round to nearest integer for non-percentage values
	return fmt.Sprintf("%.0f", value)
}
