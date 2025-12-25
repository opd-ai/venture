//go:build !headless
// +build !headless

// Package engine provides the statistics UI for displaying player statistics.
// This file implements StatisticsUI which handles rendering and interaction
// for the player statistics dashboard, showing lifetime and session stats
// across all game categories.
//
// Phase 86: Statistics UI (V15.0)
package engine

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/mobile"
	"golang.org/x/image/font/basicfont"
)

// StatisticsUI handles rendering and interaction for the statistics dashboard.
type StatisticsUI struct {
	visible      bool
	world        *World
	playerEntity *Entity
	screenWidth  int
	screenHeight int

	// Current category being viewed
	selectedCategory StatCategory
	showingSession   bool // false = lifetime, true = session

	// Touch support
	touchHandler *mobile.TouchInputHandler
	closeButton  *mobile.TouchButton
	scrollOffset float64
}

// NewStatisticsUI creates a new statistics UI system.
func NewStatisticsUI(world *World, screenWidth, screenHeight int) *StatisticsUI {
	ui := &StatisticsUI{
		visible:          false,
		world:            world,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		selectedCategory: StatCategoryCombat, // Default to combat
		showingSession:   false,              // Default to lifetime
		touchHandler:     mobile.NewTouchInputHandler(),
	}

	ui.closeButton = mobile.NewTouchButton(
		float64(screenWidth-64),
		10,
		44, 44,
		"✕",
		func() { ui.Hide() },
	)

	return ui
}

// SetPlayerEntity sets the player entity whose stats to display.
func (ui *StatisticsUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the statistics UI.
func (ui *StatisticsUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the statistics UI is currently shown.
func (ui *StatisticsUI) IsVisible() bool {
	return ui.visible
}

// Show displays the statistics UI.
func (ui *StatisticsUI) Show() {
	ui.visible = true
}

// Hide hides the statistics UI.
func (ui *StatisticsUI) Hide() {
	ui.visible = false
}

// Update processes input for the statistics UI.
func (ui *StatisticsUI) Update(entities []*Entity, deltaTime float64) {
	if ui.touchHandler != nil {
		ui.touchHandler.Update()
	}
	if ui.closeButton != nil {
		ui.closeButton.Update()
	}

	if shouldClose, shouldToggle := HandleMenuInputWithTouch(MenuKeys.Statistics, ui.visible, ui.touchHandler); shouldClose {
		if shouldToggle {
			ui.Toggle()
		} else {
			ui.Hide()
		}
		return
	}

	if !ui.visible {
		return
	}

	ui.handleCategoryNavigation()
	ui.handleScrolling()
}

// handleCategoryNavigation handles left/right arrow keys for category switching.
func (ui *StatisticsUI) handleCategoryNavigation() {
	if IsKeyJustPressed(ebiten.KeyLeft) || IsKeyJustPressed(ebiten.KeyA) {
		ui.previousCategory()
	}
	if IsKeyJustPressed(ebiten.KeyRight) || IsKeyJustPressed(ebiten.KeyD) {
		ui.nextCategory()
	}
	// Tab to toggle lifetime/session
	if IsKeyJustPressed(ebiten.KeyTab) {
		ui.showingSession = !ui.showingSession
	}
}

// handleScrolling processes touch scrolling gestures.
func (ui *StatisticsUI) handleScrolling() {
	if ui.touchHandler == nil {
		return
	}

	if direction, distance, detected := ui.touchHandler.GetSwipe(); detected {
		if direction > 1.0 || direction < -1.0 {
			if direction < 0 {
				ui.scrollOffset += distance * 0.5
			} else {
				ui.scrollOffset -= distance * 0.5
			}
			if ui.scrollOffset < 0 {
				ui.scrollOffset = 0
			}
		}
	}
}

// previousCategory switches to the previous category.
func (ui *StatisticsUI) previousCategory() {
	if ui.selectedCategory > StatCategoryCombat {
		ui.selectedCategory--
	} else {
		ui.selectedCategory = StatCategoryGeneral
	}
	ui.scrollOffset = 0
}

// nextCategory switches to the next category.
func (ui *StatisticsUI) nextCategory() {
	if ui.selectedCategory < StatCategoryGeneral {
		ui.selectedCategory++
	} else {
		ui.selectedCategory = StatCategoryCombat
	}
	ui.scrollOffset = 0
}

// Draw renders the statistics UI overlay.
func (ui *StatisticsUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok || !ui.visible || ui.playerEntity == nil {
		return
	}

	panelX, panelY, panelWidth, panelHeight := ui.calculatePanelDimensions()
	ui.drawBackground(img, panelX, panelY, panelWidth, panelHeight)
	ui.drawTitleBar(img, panelX, panelY, panelWidth)
	ui.drawCategoryTabs(img, panelX, panelY, panelWidth)
	ui.drawStatistics(img, panelX, panelY, panelWidth, panelHeight)
	ui.drawFooter(img, panelX, panelY, panelWidth, panelHeight)
}

// calculatePanelDimensions determines panel position and size.
func (ui *StatisticsUI) calculatePanelDimensions() (x, y, width, height int) {
	width = 700
	height = 500
	if ui.screenWidth < 700 {
		width = ui.screenWidth - 40
	}
	if ui.screenHeight < 500 {
		height = ui.screenHeight - 40
	}
	x = (ui.screenWidth - width) / 2
	y = (ui.screenHeight - height) / 2
	return x, y, width, height
}

// drawBackground renders the overlay and panel background.
func (ui *StatisticsUI) drawBackground(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	vector.DrawFilledRect(img, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 180}, false)
	vector.DrawFilledRect(img, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)
	vector.StrokeRect(img, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)
}

// drawTitleBar renders the panel title and mode toggle.
func (ui *StatisticsUI) drawTitleBar(img *ebiten.Image, panelX, panelY, panelWidth int) {
	titleText := "PLAYER STATISTICS"
	titleX := panelX + panelWidth/2 - len(titleText)*3
	titleY := panelY + 20
	text.Draw(img, titleText, basicfont.Face7x13, titleX, titleY+13,
		color.RGBA{255, 255, 100, 255})

	// Draw lifetime/session toggle indicator
	modeText := "Lifetime"
	modeColor := color.RGBA{100, 200, 100, 255}
	if ui.showingSession {
		modeText = "Session"
		modeColor = color.RGBA{100, 100, 255, 255}
	}
	text.Draw(img, "[TAB] "+modeText, basicfont.Face7x13, panelX+panelWidth-120, titleY+13, modeColor)

	if ui.closeButton != nil {
		ui.closeButton.SetPosition(float64(panelX+panelWidth-54), float64(panelY+10))
		ui.closeButton.Draw(img)
	}
}

// drawCategoryTabs renders the category navigation tabs.
func (ui *StatisticsUI) drawCategoryTabs(img *ebiten.Image, panelX, panelY, panelWidth int) {
	tabY := panelY + 50
	tabHeight := 25

	categories := []StatCategory{
		StatCategoryCombat,
		StatCategoryQuest,
		StatCategoryCrafting,
		StatCategoryExploration,
		StatCategorySocial,
		StatCategoryPvP,
		StatCategoryEconomy,
		StatCategoryGeneral,
	}

	tabWidth := (panelWidth - 20) / len(categories)
	x := panelX + 10

	for _, cat := range categories {
		isSelected := cat == ui.selectedCategory

		// Tab background
		bgColor := color.RGBA{40, 40, 50, 255}
		if isSelected {
			bgColor = color.RGBA{60, 80, 120, 255}
		}
		vector.DrawFilledRect(img, float32(x), float32(tabY), float32(tabWidth-2), float32(tabHeight), bgColor, false)

		// Tab border
		borderColor := color.RGBA{80, 80, 100, 255}
		if isSelected {
			borderColor = color.RGBA{100, 150, 200, 255}
		}
		vector.StrokeRect(img, float32(x), float32(tabY), float32(tabWidth-2), float32(tabHeight), 1, borderColor, false)

		// Tab text (abbreviated)
		tabText := getCategoryAbbrev(cat)
		textColor := color.RGBA{150, 150, 150, 255}
		if isSelected {
			textColor = color.RGBA{255, 255, 255, 255}
		}
		textX := x + tabWidth/2 - len(tabText)*3
		text.Draw(img, tabText, basicfont.Face7x13, textX, tabY+17, textColor)

		x += tabWidth
	}

	// Navigation hint
	navHint := "< [A/Left]  [D/Right] >"
	text.Draw(img, navHint, basicfont.Face7x13, panelX+10, tabY+tabHeight+15, color.RGBA{120, 120, 140, 255})
}

// getCategoryAbbrev returns an abbreviated name for a category.
func getCategoryAbbrev(cat StatCategory) string {
	abbrevs := []string{"CMB", "QST", "CRF", "EXP", "SOC", "PVP", "ECO", "GEN"}
	if int(cat) < len(abbrevs) {
		return abbrevs[cat]
	}
	return "???"
}

// drawStatistics renders the statistics for the selected category.
func (ui *StatisticsUI) drawStatistics(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	statsComp := ui.getPlayerStatistics()
	if statsComp == nil {
		noDataText := "No statistics available"
		text.Draw(img, noDataText, basicfont.Face7x13, panelX+panelWidth/2-len(noDataText)*3, panelY+panelHeight/2, color.RGBA{150, 150, 150, 255})
		return
	}

	// Content area
	contentX := panelX + 20
	contentY := panelY + 100
	contentWidth := panelWidth - 40
	contentHeight := panelHeight - 160

	// Draw content background
	vector.DrawFilledRect(img, float32(contentX), float32(contentY), float32(contentWidth), float32(contentHeight),
		color.RGBA{30, 30, 40, 200}, false)

	// Get stats for selected category
	defs := GetStatDefinitionsByCategory(ui.selectedCategory)

	// Sort by ID for consistent ordering
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].ID < defs[j].ID
	})

	// Draw category header
	headerText := ui.selectedCategory.String() + " Statistics"
	text.Draw(img, headerText, basicfont.Face7x13, contentX+10, contentY+20, color.RGBA{200, 200, 255, 255})

	// Draw stats
	y := contentY + 45 - int(ui.scrollOffset)
	lineHeight := 22

	for _, def := range defs {
		// Skip if out of view
		if y < contentY+30 || y > contentY+contentHeight-10 {
			y += lineHeight
			continue
		}

		var value int64
		if ui.showingSession {
			value = statsComp.GetSessionStat(def.ID)
		} else {
			value = statsComp.GetLifetimeStat(def.ID)
		}

		ui.drawStatLine(img, contentX+10, y, def, value, contentWidth-20)
		y += lineHeight
	}
}

// drawStatLine renders a single statistic line.
func (ui *StatisticsUI) drawStatLine(img *ebiten.Image, x, y int, def StatDefinition, value int64, width int) {
	// Stat name
	text.Draw(img, def.Name+":", basicfont.Face7x13, x, y, color.RGBA{180, 180, 180, 255})

	// Stat value (right-aligned)
	valueStr := formatStatDisplayValue(value, def)
	valueX := x + width - len(valueStr)*7
	text.Draw(img, valueStr, basicfont.Face7x13, valueX, y, ui.getStatValueColor(value, def))
}

// formatStatDisplayValue formats a stat value for display.
func formatStatDisplayValue(value int64, def StatDefinition) string {
	if def.IsTime {
		// Format as hours:minutes:seconds
		hours := value / 3600
		minutes := (value % 3600) / 60
		seconds := value % 60
		if hours > 0 {
			return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
		} else if minutes > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%ds", seconds)
	}

	// Format large numbers with suffixes
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(value)/1000000)
	} else if value >= 1000 {
		return fmt.Sprintf("%.1fK", float64(value)/1000)
	}
	return fmt.Sprintf("%d", value)
}

// getStatValueColor returns a color based on stat value (higher = greener).
func (ui *StatisticsUI) getStatValueColor(value int64, _ StatDefinition) color.Color {
	if value == 0 {
		return color.RGBA{100, 100, 100, 255}
	}
	if value >= 1000000 {
		return color.RGBA{255, 215, 0, 255} // Gold for very high
	}
	if value >= 10000 {
		return color.RGBA{100, 255, 100, 255} // Green for high
	}
	if value >= 100 {
		return color.RGBA{150, 200, 255, 255} // Blue for medium
	}
	return color.RGBA{200, 200, 200, 255} // White for low
}

// getPlayerStatistics returns the player's statistics component.
func (ui *StatisticsUI) getPlayerStatistics() *PlayerStatisticsComponent {
	if ui.playerEntity == nil {
		return nil
	}

	comp, exists := ui.playerEntity.GetComponent("player_statistics")
	if !exists {
		return nil
	}

	statsComp, ok := comp.(*PlayerStatisticsComponent)
	if !ok {
		return nil
	}

	return statsComp
}

// drawFooter renders controls hint.
func (ui *StatisticsUI) drawFooter(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	exitHint := GetExitHint(MenuKeys.Statistics)
	controlsText := exitHint + " | [TAB] Toggle Mode | [←/→] Categories"
	controlsX := panelX + panelWidth/2 - len(controlsText)*3
	controlsY := panelY + panelHeight - 15
	text.Draw(img, controlsText, basicfont.Face7x13, controlsX, controlsY, color.RGBA{180, 180, 180, 255})
}

// IsActive returns whether the statistics UI is currently visible.
func (ui *StatisticsUI) IsActive() bool {
	return ui.visible
}

// SetActive sets whether the statistics UI is visible.
func (ui *StatisticsUI) SetActive(active bool) {
	ui.visible = active
}

// Compile-time check that StatisticsUI implements UISystem
var _ UISystem = (*StatisticsUI)(nil)
