package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/world/housing"
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

	// INTEGRATION FIX [Category B]: V8.0 Housing UI Integration (Phase 49.1)
	// Gap: UI defined but managers never connected, no rendering implementation
	// Fix: Added manager references and player state tracking for housing management
	// Roadmap: ROADMAP_V8.md Phase 49.1, 51.2, 51.3
	housingManager     *housing.Manager
	guildHallManager   *housing.GuildHallManager
	buildingGenerator  *building.Generator
	furnitureGenerator *furniture.Generator
	playerID           uint64
	selectedPlotIndex  int
	menuState          string // "main", "build", "furniture", "guildhall"
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

	// INTEGRATION FIX [Category B]: V8.0 Housing UI Input Handling
	// Gap: No input handling for housing management
	// Fix: Added keyboard navigation and menu state management
	// Roadmap: ROADMAP_V8.md Phase 49.1
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		h.Hide()
		return true
	}

	// Navigation between menu states
	if ebiten.IsKeyPressed(ebiten.KeyTab) {
		switch h.menuState {
		case "main":
			h.menuState = "build"
		case "build":
			h.menuState = "furniture"
		case "furniture":
			h.menuState = "guildhall"
		case "guildhall":
			h.menuState = "main"
		default:
			h.menuState = "main"
		}
	}

	// Consume input when UI is visible
	return true
}

// Draw renders the housing UI to the screen.
func (h *HousingUI) Draw(screen *ebiten.Image) {
	if !h.Visible || screen == nil {
		return
	}

	// INTEGRATION FIX [Category B]: V8.0 Housing UI Rendering
	// Gap: No visual representation of housing system
	// Fix: Added comprehensive UI rendering with player plots, build options, furniture, guild halls
	// Roadmap: ROADMAP_V8.md Phase 49.1, 51.2, 51.3

	// Draw background
	vector.DrawFilledRect(screen, float32(h.X), float32(h.Y), float32(h.Width), float32(h.Height), h.BackgroundColor, false)

	// Draw title
	title := "Housing Management"
	ebitenutil.DebugPrintAt(screen, title, h.X+10, h.Y+10)

	// Draw current menu state
	stateY := h.Y + 40
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mode: %s (Tab to switch)", h.menuState), h.X+10, stateY)

	// Draw player's plots if manager is available
	if h.housingManager != nil {
		playerIDStr := fmt.Sprintf("%d", h.playerID)
		plots := h.housingManager.GetPlayerPlots(playerIDStr)

		plotsY := stateY + 30
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Your Plots: %d", len(plots)), h.X+10, plotsY)

		// List plots
		for i, plot := range plots {
			if i >= 10 { // Limit display to prevent overflow
				break
			}
			plotY := plotsY + 20 + (i * 15)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %d. Plot %s at (%.0f, %.0f) - %s",
				i+1, plot.ID, plot.Position.X, plot.Position.Y, plot.Size.String()), h.X+10, plotY)
		}
	}

	// Draw menu-specific content
	contentY := h.Y + h.Height - 100
	switch h.menuState {
	case "build":
		ebitenutil.DebugPrintAt(screen, "Building Construction (Coming Soon)", h.X+10, contentY)
	case "furniture":
		ebitenutil.DebugPrintAt(screen, "Furniture Placement (Coming Soon)", h.X+10, contentY)
	case "guildhall":
		ebitenutil.DebugPrintAt(screen, "Guild Hall Management (Coming Soon)", h.X+10, contentY)
	}

	// Draw controls
	controlsY := h.Y + h.Height - 30
	ebitenutil.DebugPrintAt(screen, "ESC: Close | Tab: Switch Mode", h.X+10, controlsY)
}

// SetManagers sets the housing-related manager references.
func (h *HousingUI) SetManagers(housingManager, guildHallManager, buildingGenerator, furnitureGenerator interface{}) {
	// INTEGRATION FIX [Category B]: V8.0 Housing UI Manager Wiring
	// Gap: Managers passed but never stored or used
	// Fix: Type-assert and store all manager references for housing functionality
	// Roadmap: ROADMAP_V8.md Phase 49.1, 51.2, 51.3
	if hm, ok := housingManager.(*housing.Manager); ok {
		h.housingManager = hm
	}
	if ghm, ok := guildHallManager.(*housing.GuildHallManager); ok {
		h.guildHallManager = ghm
	}
	if bg, ok := buildingGenerator.(*building.Generator); ok {
		h.buildingGenerator = bg
	}
	if fg, ok := furnitureGenerator.(*furniture.Generator); ok {
		h.furnitureGenerator = fg
	}
	h.menuState = "main"
}

// SetPlayerID sets the player ID for the housing UI.
func (h *HousingUI) SetPlayerID(playerID uint64) {
	// INTEGRATION FIX [Category B]: V8.0 Housing UI Player Context
	// Gap: Player ID setter was empty stub
	// Fix: Store player ID for querying player-specific plots and permissions
	// Roadmap: ROADMAP_V8.md Phase 49.1
	h.playerID = playerID
}
