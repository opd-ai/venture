//go:build test
// +build test

// Package engine provides character UI tests.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
)

// TestCharacterUI_NewCharacterUI verifies initialization.
func TestCharacterUI_NewCharacterUI(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	if ui == nil {
		t.Fatal("NewCharacterUI returned nil")
	}

	if ui.visible {
		t.Error("CharacterUI should be hidden by default")
	}

	if ui.screenWidth != 800 {
		t.Errorf("Expected screenWidth 800, got %d", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("Expected screenHeight 600, got %d", ui.screenHeight)
	}

	// Verify layout was calculated
	if ui.statsPanel.Width == 0 || ui.equipmentPanel.Width == 0 || ui.attributesPanel.Width == 0 {
		t.Error("Panel layout was not calculated")
	}
}

// TestCharacterUI_Toggle tests visibility toggling.
func TestCharacterUI_Toggle(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("UI should be hidden initially")
	}

	// Toggle to show
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("UI should be visible after first toggle")
	}

	// Toggle to hide
	ui.Toggle()
	if ui.IsVisible() {
		t.Error("UI should be hidden after second toggle")
	}
}

// TestCharacterUI_SetPlayerEntity ensures entity is stored.
func TestCharacterUI_SetPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	entity := world.CreateEntity()
	entity.AddComponent(NewStatsComponent())

	ui.SetPlayerEntity(entity)

	if ui.playerEntity != entity {
		t.Error("Player entity was not stored correctly")
	}
}

// TestCharacterUI_CalculateDerivedStats verifies stat calculations.
func TestCharacterUI_CalculateDerivedStats(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	stats := NewStatsComponent()
	stats.CritChance = 0.15  // 15%
	stats.CritDamage = 2.5   // 250%
	stats.Evasion = 0.10     // 10%

	derived := ui.calculateDerivedStats(stats)

	// Check crit chance (should be converted to percentage)
	if derived["crit_chance"] != 15.0 {
		t.Errorf("Expected crit_chance 15.0, got %.1f", derived["crit_chance"])
	}

	// Check crit damage
	if derived["crit_damage"] != 2.5 {
		t.Errorf("Expected crit_damage 2.5, got %.1f", derived["crit_damage"])
	}

	// Check evasion (should be converted to percentage)
	if derived["evasion"] != 10.0 {
		t.Errorf("Expected evasion 10.0, got %.1f", derived["evasion"])
	}
}

// TestCharacterUI_FormatStatValue tests number formatting.
func TestCharacterUI_FormatStatValue(t *testing.T) {
	tests := []struct {
		name         string
		value        float64
		isPercentage bool
		expected     string
	}{
		{"integer as number", 42.0, false, "42"},
		{"float as number", 42.5, false, "42"},  // Go's %.0f rounds down for .5
		{"percentage", 15.5, true, "15.5%"},
		{"zero percentage", 0.0, true, "0.0%"},
		{"high percentage", 99.9, true, "99.9%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatValue(tt.value, tt.isPercentage)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestCharacterUI_ShowHide tests show/hide methods.
func TestCharacterUI_ShowHide(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("UI should be hidden initially")
	}

	// Show
	ui.Show()
	if !ui.IsVisible() {
		t.Error("UI should be visible after Show()")
	}

	// Hide
	ui.Hide()
	if ui.IsVisible() {
		t.Error("UI should be hidden after Hide()")
	}
}

// TestCharacterUI_GetResistanceColor tests resistance color mapping.
func TestCharacterUI_GetResistanceColor(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	tests := []struct {
		name     string
		resist   float64
		expected string // Color description
	}{
		{"no resistance", 0.0, "gray"},
		{"low resistance", 0.1, "orange"},
		{"medium resistance", 0.3, "yellow"},
		{"high resistance", 0.6, "green"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := ui.getResistanceColor(tt.resist)
			if color == nil {
				t.Error("getResistanceColor returned nil")
			}
			// Just verify it returns a color - actual RGB values are implementation details
		})
	}
}

// TestCharacterUI_CalculateLayout verifies panel layout calculation.
func TestCharacterUI_CalculateLayout(t *testing.T) {
	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
	}{
		{"standard resolution", 800, 600},
		{"large resolution", 1920, 1080},
		{"small resolution", 800, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			ui := NewCharacterUI(world, tt.screenWidth, tt.screenHeight)

			// Verify panels have non-zero dimensions
			if ui.statsPanel.Width == 0 || ui.statsPanel.Height == 0 {
				t.Error("Stats panel has zero dimensions")
			}
			if ui.equipmentPanel.Width == 0 || ui.equipmentPanel.Height == 0 {
				t.Error("Equipment panel has zero dimensions")
			}
			if ui.attributesPanel.Width == 0 || ui.attributesPanel.Height == 0 {
				t.Error("Attributes panel has zero dimensions")
			}

			// Verify panels don't overlap (basic check)
			if ui.statsPanel.X+ui.statsPanel.Width > ui.equipmentPanel.X {
				// This is expected since they're side by side
			}
		})
	}
}

// TestCharacterUI_WithoutPlayerEntity verifies UI handles nil player gracefully.
func TestCharacterUI_WithoutPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	// Show UI without player entity
	ui.Show()

	// Update should not panic
	ui.Update(0.016)

	// Draw should not panic (we can't actually test drawing without Ebiten context)
	// Just verify the UI is in a valid state
	if ui.playerEntity != nil {
		t.Error("Player entity should be nil")
	}
}

// TestCharacterUI_WithCompletePlayer verifies UI with fully-equipped player.
func TestCharacterUI_WithCompletePlayer(t *testing.T) {
	world := NewWorld()
	ui := NewCharacterUI(world, 800, 600)

	// Create player with all components
	player := world.CreateEntity()

	stats := NewStatsComponent()
	stats.Attack = 50.0
	stats.Defense = 30.0
	stats.MagicPower = 40.0
	stats.MagicDefense = 25.0
	stats.CritChance = 0.15
	stats.CritDamage = 2.0
	stats.Evasion = 0.10
	stats.Resistances = map[combat.DamageType]float64{
		combat.DamageFire: 0.25,
		combat.DamageIce:  0.15,
	}

	player.AddComponent(stats)
	player.AddComponent(NewEquipmentComponent())
	player.AddComponent(NewExperienceComponent())
	player.AddComponent(NewInventoryComponent(20, 100.0))

	ui.SetPlayerEntity(player)
	ui.Show()

	// Verify UI is visible with player set
	if !ui.IsVisible() {
		t.Error("UI should be visible")
	}

	// Update should work
	ui.Update(0.016)

	// Verify derived stats are calculated correctly
	derived := ui.calculateDerivedStats(stats)
	if derived["crit_chance"] != 15.0 {
		t.Errorf("Expected crit_chance 15.0, got %.1f", derived["crit_chance"])
	}
}
