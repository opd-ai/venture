package housing

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
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
	housingManager     *Manager
	guildHallManager   *GuildHallManager
	buildingGenerator  *building.Generator
	furnitureGenerator *furniture.Generator
	playerID           uint64
	selectedPlotIndex  int
	menuState          string // "main", "build", "furniture", "guildhall"

	// Selection tracking for submenus
	selectedBuildingType  int // Index into building types list
	selectedFurnitureType int // Index into furniture types list
	guildID               string
	tabCooldown           int // Prevent rapid tab switching
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

// IsVisible returns true if the housing UI is currently visible.
// INTEGRATION FIX [Category B]: V8.0 Housing UI Integration (Phase 49.1)
// Gap: InputSystem needs IsVisible method for ESC key handling
// Fix: Added IsVisible method to satisfy HousingUIProvider interface
func (h *HousingUI) IsVisible() bool {
	return h.Visible
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

	// Cooldown for tab switching to prevent rapid cycling
	if h.tabCooldown > 0 {
		h.tabCooldown--
	}

	// Navigation between menu states
	if ebiten.IsKeyPressed(ebiten.KeyTab) && h.tabCooldown == 0 {
		h.tabCooldown = 15 // ~250ms at 60 FPS
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

	// Handle submenu navigation with Up/Down arrows
	h.handleSubmenuInput()

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
	contentY := h.Y + h.Height - 200
	switch h.menuState {
	case "build":
		h.drawBuildingMenu(screen, contentY)
	case "furniture":
		h.drawFurnitureMenu(screen, contentY)
	case "guildhall":
		h.drawGuildHallMenu(screen, contentY)
	}

	// Draw controls
	controlsY := h.Y + h.Height - 30
	ebitenutil.DebugPrintAt(screen, "ESC: Close | Tab: Switch Mode | Up/Down: Select", h.X+10, controlsY)
}

// SetManagers sets the housing-related manager references.
func (h *HousingUI) SetManagers(housingManager, guildHallManager, buildingGenerator, furnitureGenerator interface{}) {
	// INTEGRATION FIX [Category B]: V8.0 Housing UI Manager Wiring
	// Gap: Managers passed but never stored or used
	// Fix: Type-assert and store all manager references for housing functionality
	// Roadmap: ROADMAP_V8.md Phase 49.1, 51.2, 51.3
	if hm, ok := housingManager.(*Manager); ok {
		h.housingManager = hm
	}
	if ghm, ok := guildHallManager.(*GuildHallManager); ok {
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

// SetGuildID sets the guild ID for guild hall management.
func (h *HousingUI) SetGuildID(guildID string) {
	h.guildID = guildID
}

// handleSubmenuInput handles Up/Down arrow navigation within submenus.
func (h *HousingUI) handleSubmenuInput() {
	// Simple cooldown using static variable pattern via field check
	switch h.menuState {
	case "build":
		buildingTypes := h.getBuildingTypesList()
		if ebiten.IsKeyPressed(ebiten.KeyDown) && h.tabCooldown == 0 {
			h.tabCooldown = 10
			h.selectedBuildingType = (h.selectedBuildingType + 1) % len(buildingTypes)
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) && h.tabCooldown == 0 {
			h.tabCooldown = 10
			h.selectedBuildingType--
			if h.selectedBuildingType < 0 {
				h.selectedBuildingType = len(buildingTypes) - 1
			}
		}
	case "furniture":
		furnitureTypes := h.getFurnitureTypesList()
		if ebiten.IsKeyPressed(ebiten.KeyDown) && h.tabCooldown == 0 {
			h.tabCooldown = 10
			h.selectedFurnitureType = (h.selectedFurnitureType + 1) % len(furnitureTypes)
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) && h.tabCooldown == 0 {
			h.tabCooldown = 10
			h.selectedFurnitureType--
			if h.selectedFurnitureType < 0 {
				h.selectedFurnitureType = len(furnitureTypes) - 1
			}
		}
	}
}

// getBuildingTypesList returns the list of available building types.
func (h *HousingUI) getBuildingTypesList() []string {
	return []string{
		"House",
		"Workshop",
		"Storage",
		"Tower",
		"Manor",
		"Guild Hall",
	}
}

// getFurnitureTypesList returns the list of furniture categories.
func (h *HousingUI) getFurnitureTypesList() []string {
	return []string{
		"Seating",
		"Storage",
		"Crafting",
		"Decoration",
		"Lighting",
		"Bedding",
	}
}

// drawBuildingMenu renders the building construction submenu.
func (h *HousingUI) drawBuildingMenu(screen *ebiten.Image, startY int) {
	ebitenutil.DebugPrintAt(screen, "=== Building Construction ===", h.X+10, startY)

	buildingTypes := h.getBuildingTypesList()
	for i, bt := range buildingTypes {
		y := startY + 20 + (i * 15)
		prefix := "  "
		if i == h.selectedBuildingType {
			prefix = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s", prefix, bt), h.X+10, y)
	}

	// Show building info for selected type
	if h.selectedBuildingType >= 0 && h.selectedBuildingType < len(buildingTypes) {
		infoY := startY + 20 + (len(buildingTypes) * 15) + 10
		selected := buildingTypes[h.selectedBuildingType]
		info := h.getBuildingInfo(selected)
		ebitenutil.DebugPrintAt(screen, info, h.X+10, infoY)
	}
}

// getBuildingInfo returns descriptive info for a building type.
func (h *HousingUI) getBuildingInfo(buildingType string) string {
	switch buildingType {
	case "House":
		return "Small living space (2-4 rooms). Perfect for starting adventurers."
	case "Workshop":
		return "Crafting facility (3-6 rooms). Includes workbenches and storage."
	case "Storage":
		return "Warehouse (4-8 rooms). Maximum item capacity for collectors."
	case "Tower":
		return "Vertical structure (1-3 floors). Great views, magical ambiance."
	case "Manor":
		return "Large estate (6-12 rooms). Multiple floors, gardens, prestige."
	case "Guild Hall":
		return "Guild headquarters (8-16 rooms). Meeting halls, training areas."
	default:
		return "Select a building type."
	}
}

// drawFurnitureMenu renders the furniture placement submenu.
func (h *HousingUI) drawFurnitureMenu(screen *ebiten.Image, startY int) {
	ebitenutil.DebugPrintAt(screen, "=== Furniture Catalog ===", h.X+10, startY)

	furnitureTypes := h.getFurnitureTypesList()
	for i, ft := range furnitureTypes {
		y := startY + 20 + (i * 15)
		prefix := "  "
		if i == h.selectedFurnitureType {
			prefix = "> "
		}
		// Get item count for this category
		count := h.getFurnitureCountForCategory(ft)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s (%d items)", prefix, ft, count), h.X+10, y)
	}

	// Show furniture items for selected category
	if h.selectedFurnitureType >= 0 && h.selectedFurnitureType < len(furnitureTypes) {
		itemsY := startY + 20 + (len(furnitureTypes) * 15) + 10
		selected := furnitureTypes[h.selectedFurnitureType]
		items := h.getFurnitureItemsForCategory(selected)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Items: %s", items), h.X+10, itemsY)
	}
}

// getFurnitureCountForCategory returns the number of items in a category.
func (h *HousingUI) getFurnitureCountForCategory(category string) int {
	furnitureType := h.categoryToFurnitureType(category)
	if furnitureType < 0 {
		return 0
	}
	return len(furniture.GetSubTypesByCategory(furniture.FurnitureType(furnitureType)))
}

// getFurnitureItemsForCategory returns a comma-separated list of items.
func (h *HousingUI) getFurnitureItemsForCategory(category string) string {
	furnitureType := h.categoryToFurnitureType(category)
	if furnitureType < 0 {
		return "None"
	}
	items := furniture.GetSubTypesByCategory(furniture.FurnitureType(furnitureType))
	if len(items) == 0 {
		return "None"
	}
	// Limit displayed items to prevent overflow
	if len(items) > 5 {
		return fmt.Sprintf("%s, %s, %s, %s, %s...", items[0], items[1], items[2], items[3], items[4])
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += ", " + items[i]
	}
	return result
}

// categoryToFurnitureType converts category name to furniture.FurnitureType.
func (h *HousingUI) categoryToFurnitureType(category string) int {
	switch category {
	case "Seating":
		return int(furniture.TypeSeating)
	case "Storage":
		return int(furniture.TypeStorage)
	case "Crafting":
		return int(furniture.TypeCrafting)
	case "Decoration":
		return int(furniture.TypeDecoration)
	case "Lighting":
		return int(furniture.TypeLighting)
	case "Bedding":
		return int(furniture.TypeBedding)
	default:
		return -1
	}
}

// drawGuildHallMenu renders the guild hall management submenu.
func (h *HousingUI) drawGuildHallMenu(screen *ebiten.Image, startY int) {
	ebitenutil.DebugPrintAt(screen, "=== Guild Hall Management ===", h.X+10, startY)

	if h.guildHallManager == nil || h.guildID == "" {
		ebitenutil.DebugPrintAt(screen, "No guild hall manager or guild ID set.", h.X+10, startY+20)
		ebitenutil.DebugPrintAt(screen, "Join a guild to manage guild halls.", h.X+10, startY+35)
		return
	}

	guildHall, found := h.guildHallManager.GetGuildHall(h.guildID)
	if !found {
		ebitenutil.DebugPrintAt(screen, "Your guild does not have a hall yet.", h.X+10, startY+20)
		ebitenutil.DebugPrintAt(screen, "A guild leader can create one.", h.X+10, startY+35)
		return
	}

	// Display guild hall info
	y := startY + 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Guild: %s", guildHall.OwnerGuildName), h.X+10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Size: %s", guildHall.Size.String()), h.X+10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Phase: %s", guildHall.Phase.String()), h.X+10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Floors: %d", guildHall.Floors), h.X+10, y)
	y += 15

	// Show construction progress if not complete
	if guildHall.Phase != PhaseComplete {
		ebitenutil.DebugPrintAt(screen, "--- Construction Progress ---", h.X+10, y)
		y += 15
		for matType, required := range guildHall.RequiredMaterials {
			contributed := guildHall.Materials[matType]
			percent := 0
			if required > 0 {
				percent = (contributed * 100) / required
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %s: %d/%d (%d%%)",
				matType.String(), contributed, required, percent), h.X+10, y)
			y += 15
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "Construction Complete!", h.X+10, y)
	}
}
