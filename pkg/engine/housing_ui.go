package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// HousingUI represents a player housing management interface.
// Phase 49.1, 51.2, 51.3
type HousingUI struct {
	X, Y          int
	Width, Height int
	GenreID       string
	Visible       bool

	// Colors
	BackgroundColor color.Color
	TextColor       color.Color
	HighlightColor  color.Color
}

// NewHousingUI creates a new housing UI instance.
func NewHousingUI(screenWidth, screenHeight int) *HousingUI {
	width := screenWidth - 40
	height := screenHeight - 100
	x := 20
	y := 50

	return &HousingUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		GenreID:         "fantasy",
		Visible:         false,
		BackgroundColor: color.RGBA{20, 20, 30, 230},
		TextColor:       color.RGBA{220, 220, 220, 255},
		HighlightColor:  color.RGBA{100, 150, 255, 255},
	}
}

// Toggle toggles the visibility of the housing UI.
func (h *HousingUI) Toggle() {
	h.Visible = !h.Visible
}

// Show makes the housing UI visible.
func (h *HousingUI) Show() {
	h.Visible = true
}

// Hide makes the housing UI invisible.
func (h *HousingUI) Hide() {
	h.Visible = false
}

// Update updates the housing UI state.
// Returns true if the UI consumed the input (blocking pass-through).
func (h *HousingUI) Update() bool {
	if !h.Visible {
		return false
	}
	// Minimal stub implementation - will be enhanced in Phase 49.1
	return true
}

// Draw renders the housing UI to the screen.
func (h *HousingUI) Draw(screen *ebiten.Image) {
	if !h.Visible || screen == nil {
		return
	}
	// Minimal stub implementation - will be enhanced in Phase 49.1
}

// SetManagers sets the housing-related manager references.
func (h *HousingUI) SetManagers(housingManager, guildHallManager, buildingGenerator, furnitureGenerator interface{}) {
	// Minimal stub implementation - will be enhanced in Phase 49.1
}

// SetPlayerID sets the player ID for the housing UI.
func (h *HousingUI) SetPlayerID(playerID uint64) {
	// Minimal stub implementation - will be enhanced in Phase 49.1
}
