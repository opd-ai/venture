//go:build !headless
// +build !headless

// Package engine provides the achievement UI for browsing player achievements.
// This file implements AchievementUI which handles rendering and interaction
// for the achievement browser, showing achievement progress across all categories.
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

// AchievementUI handles rendering and interaction for the achievement browser.
type AchievementUI struct {
	visible      bool
	world        *World
	playerEntity *Entity
	screenWidth  int
	screenHeight int

	// Current category being viewed
	selectedCategory AchievementCategory
	selectedIndex    int // Index of selected achievement in current category

	// Touch support
	touchHandler *mobile.TouchInputHandler
	closeButton  *mobile.TouchButton
	scrollOffset float64
}

// NewAchievementUI creates a new achievement UI system.
func NewAchievementUI(world *World, screenWidth, screenHeight int) *AchievementUI {
	ui := &AchievementUI{
		visible:          false,
		world:            world,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		selectedCategory: AchievementCategoryCombat,
		selectedIndex:    0,
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

// SetPlayerEntity sets the player entity whose achievements to display.
func (ui *AchievementUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the achievement UI.
func (ui *AchievementUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the achievement UI is currently shown.
func (ui *AchievementUI) IsVisible() bool {
	return ui.visible
}

// Show displays the achievement UI.
func (ui *AchievementUI) Show() {
	ui.visible = true
}

// Hide hides the achievement UI.
func (ui *AchievementUI) Hide() {
	ui.visible = false
}

// Update processes input for the achievement UI.
func (ui *AchievementUI) Update(entities []*Entity, deltaTime float64) {
	if ui.touchHandler != nil {
		ui.touchHandler.Update()
	}
	if ui.closeButton != nil {
		ui.closeButton.Update()
	}

	if shouldClose, shouldToggle := HandleMenuInputWithTouch(MenuKeys.Achievements, ui.visible, ui.touchHandler); shouldClose {
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
	ui.handleAchievementNavigation()
	ui.handleScrolling()
}

// handleCategoryNavigation handles left/right navigation for categories.
func (ui *AchievementUI) handleCategoryNavigation() {
	if IsKeyJustPressed(ebiten.KeyLeft) {
		ui.previousCategory()
	}
	if IsKeyJustPressed(ebiten.KeyRight) {
		ui.nextCategory()
	}
}

// handleAchievementNavigation handles up/down navigation for achievements.
func (ui *AchievementUI) handleAchievementNavigation() {
	if IsKeyJustPressed(ebiten.KeyUp) {
		ui.previousAchievement()
	}
	if IsKeyJustPressed(ebiten.KeyDown) {
		ui.nextAchievement()
	}
}

// handleScrolling processes touch scrolling gestures.
func (ui *AchievementUI) handleScrolling() {
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
func (ui *AchievementUI) previousCategory() {
	if ui.selectedCategory > AchievementCategoryCombat {
		ui.selectedCategory--
	} else {
		ui.selectedCategory = AchievementCategoryPvP
	}
	ui.selectedIndex = 0
	ui.scrollOffset = 0
}

// nextCategory switches to the next category.
func (ui *AchievementUI) nextCategory() {
	if ui.selectedCategory < AchievementCategoryPvP {
		ui.selectedCategory++
	} else {
		ui.selectedCategory = AchievementCategoryCombat
	}
	ui.selectedIndex = 0
	ui.scrollOffset = 0
}

// previousAchievement selects the previous achievement.
func (ui *AchievementUI) previousAchievement() {
	if ui.selectedIndex > 0 {
		ui.selectedIndex--
	}
}

// nextAchievement selects the next achievement.
func (ui *AchievementUI) nextAchievement() {
	defs := GetAchievementDefinitionsByCategory(ui.selectedCategory)
	if ui.selectedIndex < len(defs)-1 {
		ui.selectedIndex++
	}
}

// Draw renders the achievement UI overlay.
func (ui *AchievementUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok || !ui.visible || ui.playerEntity == nil {
		return
	}

	panelX, panelY, panelWidth, panelHeight := ui.calculatePanelDimensions()
	ui.drawBackground(img, panelX, panelY, panelWidth, panelHeight)
	ui.drawTitleBar(img, panelX, panelY, panelWidth)
	ui.drawCategoryTabs(img, panelX, panelY, panelWidth)
	ui.drawAchievements(img, panelX, panelY, panelWidth, panelHeight)
	ui.drawSelectedAchievementDetails(img, panelX, panelY, panelWidth, panelHeight)
	ui.drawFooter(img, panelX, panelY, panelWidth, panelHeight)
}

// calculatePanelDimensions determines panel position and size.
func (ui *AchievementUI) calculatePanelDimensions() (x, y, width, height int) {
	width = 800
	height = 550
	if ui.screenWidth < 800 {
		width = ui.screenWidth - 40
	}
	if ui.screenHeight < 550 {
		height = ui.screenHeight - 40
	}
	x = (ui.screenWidth - width) / 2
	y = (ui.screenHeight - height) / 2
	return x, y, width, height
}

// drawBackground renders the overlay and panel background.
func (ui *AchievementUI) drawBackground(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	vector.DrawFilledRect(img, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 180}, false)
	vector.DrawFilledRect(img, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)
	vector.StrokeRect(img, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)
}

// drawTitleBar renders the panel title and total points.
func (ui *AchievementUI) drawTitleBar(img *ebiten.Image, panelX, panelY, panelWidth int) {
	titleText := "ACHIEVEMENTS"
	titleX := panelX + panelWidth/2 - len(titleText)*3
	titleY := panelY + 20
	text.Draw(img, titleText, basicfont.Face7x13, titleX, titleY+13, color.RGBA{255, 255, 100, 255})

	// Draw total points
	achieveComp := ui.getPlayerAchievements()
	totalPoints := 0
	if achieveComp != nil {
		totalPoints = achieveComp.GetTotalPoints()
	}
	pointsText := fmt.Sprintf("Total: %d pts", totalPoints)
	text.Draw(img, pointsText, basicfont.Face7x13, panelX+panelWidth-120, titleY+13, color.RGBA{255, 215, 0, 255})

	if ui.closeButton != nil {
		ui.closeButton.SetPosition(float64(panelX+panelWidth-54), float64(panelY+10))
		ui.closeButton.Draw(img)
	}
}

// drawCategoryTabs renders the category navigation tabs.
func (ui *AchievementUI) drawCategoryTabs(img *ebiten.Image, panelX, panelY, panelWidth int) {
	tabY := panelY + 50
	tabHeight := 25

	categories := []AchievementCategory{
		AchievementCategoryCombat,
		AchievementCategoryQuest,
		AchievementCategoryCrafting,
		AchievementCategoryExploration,
		AchievementCategorySocial,
		AchievementCategoryPvP,
	}

	tabWidth := (panelWidth - 20) / len(categories)
	x := panelX + 10

	achieveComp := ui.getPlayerAchievements()

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

		// Tab text with category points
		catPoints := 0
		if achieveComp != nil {
			catPoints = achieveComp.GetCategoryPoints(cat)
		}
		tabText := fmt.Sprintf("%s (%d)", getAchievementCategoryAbbrev(cat), catPoints)
		textColor := color.RGBA{150, 150, 150, 255}
		if isSelected {
			textColor = color.RGBA{255, 255, 255, 255}
		}
		textX := x + 5
		text.Draw(img, tabText, basicfont.Face7x13, textX, tabY+17, textColor)

		x += tabWidth
	}
}

// getAchievementCategoryAbbrev returns an abbreviated name for an achievement category.
func getAchievementCategoryAbbrev(cat AchievementCategory) string {
	abbrevs := []string{"CMB", "QST", "CRF", "EXP", "SOC", "PVP"}
	if int(cat) < len(abbrevs) {
		return abbrevs[cat]
	}
	return "???"
}

// drawAchievements renders the achievement list for the selected category.
func (ui *AchievementUI) drawAchievements(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	// List area (left 60%)
	listX := panelX + 10
	listY := panelY + 90
	listWidth := int(float64(panelWidth) * 0.55)
	listHeight := panelHeight - 140

	// Draw list background
	vector.DrawFilledRect(img, float32(listX), float32(listY), float32(listWidth), float32(listHeight),
		color.RGBA{30, 30, 40, 200}, false)
	vector.StrokeRect(img, float32(listX), float32(listY), float32(listWidth), float32(listHeight),
		1, color.RGBA{60, 60, 80, 255}, false)

	// Get achievements for selected category
	defs := GetAchievementDefinitionsByCategory(ui.selectedCategory)

	// Sort by ID for consistent ordering
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].ID < defs[j].ID
	})

	achieveComp := ui.getPlayerAchievements()

	y := listY + 5 - int(ui.scrollOffset)
	itemHeight := 35

	for i, def := range defs {
		// Skip if out of view
		if y < listY || y > listY+listHeight-itemHeight {
			y += itemHeight
			continue
		}

		isSelected := i == ui.selectedIndex
		ui.drawAchievementItem(img, listX+5, y, listWidth-10, itemHeight-5, def, achieveComp, isSelected)
		y += itemHeight
	}
}

// drawAchievementItem renders a single achievement item.
func (ui *AchievementUI) drawAchievementItem(img *ebiten.Image, x, y, width, height int, def AchievementDefinition, achieveComp *ExtendedAchievementComponent, isSelected bool) {
	// Item background
	bgColor := color.RGBA{35, 35, 45, 255}
	if isSelected {
		bgColor = color.RGBA{50, 60, 80, 255}
	}
	vector.DrawFilledRect(img, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Selection indicator
	if isSelected {
		vector.StrokeRect(img, float32(x), float32(y), float32(width), float32(height), 2, color.RGBA{100, 150, 255, 255}, false)
	}

	// Get current tier and progress
	currentTier := AchievementTierNone
	var progress int64
	if achieveComp != nil {
		currentTier = achieveComp.GetTier(def.ID)
		progress = achieveComp.GetProgress(def.ID)
	}

	// Draw tier icon (colored circle)
	tierColor := getTierColor(currentTier)
	iconX := x + 15
	iconY := y + height/2
	vector.DrawFilledCircle(img, float32(iconX), float32(iconY), 10, tierColor, false)
	vector.StrokeCircle(img, float32(iconX), float32(iconY), 10, 1, color.RGBA{255, 255, 255, 128}, false)

	// Draw tier initial
	tierInitial := getTierInitial(currentTier)
	text.Draw(img, tierInitial, basicfont.Face7x13, iconX-3, iconY+4, color.RGBA{255, 255, 255, 255})

	// Achievement name
	text.Draw(img, def.Name, basicfont.Face7x13, x+35, y+12, color.RGBA{220, 220, 220, 255})

	// Progress bar
	nextTier := currentTier + 1
	if nextTier > AchievementTierPlatinum {
		nextTier = AchievementTierPlatinum
	}
	nextThreshold := def.Thresholds[0] // Bronze threshold
	if currentTier >= AchievementTierBronze && int(currentTier) < len(def.Thresholds) {
		nextThreshold = def.Thresholds[currentTier]
	}

	barWidth := width - 140
	barHeight := 8
	barX := x + 35
	barY := y + 18

	// Bar background
	vector.DrawFilledRect(img, float32(barX), float32(barY), float32(barWidth), float32(barHeight), color.RGBA{50, 50, 60, 255}, false)

	// Bar fill
	progressPct := float64(progress) / float64(nextThreshold)
	if progressPct > 1.0 {
		progressPct = 1.0
	}
	fillWidth := int(float64(barWidth) * progressPct)
	vector.DrawFilledRect(img, float32(barX), float32(barY), float32(fillWidth), float32(barHeight), tierColor, false)

	// Progress text
	progressText := fmt.Sprintf("%d/%d", progress, nextThreshold)
	text.Draw(img, progressText, basicfont.Face7x13, x+width-70, y+20, color.RGBA{150, 150, 150, 255})
}

// getTierColor returns the color for an achievement tier.
func getTierColor(tier AchievementTier) color.Color {
	switch tier {
	case AchievementTierNone:
		return color.RGBA{80, 80, 80, 255}
	case AchievementTierBronze:
		return color.RGBA{205, 127, 50, 255}
	case AchievementTierSilver:
		return color.RGBA{192, 192, 192, 255}
	case AchievementTierGold:
		return color.RGBA{255, 215, 0, 255}
	case AchievementTierPlatinum:
		return color.RGBA{229, 228, 226, 255}
	default:
		return color.RGBA{80, 80, 80, 255}
	}
}

// getTierInitial returns the first letter of a tier name.
func getTierInitial(tier AchievementTier) string {
	switch tier {
	case AchievementTierNone:
		return "-"
	case AchievementTierBronze:
		return "B"
	case AchievementTierSilver:
		return "S"
	case AchievementTierGold:
		return "G"
	case AchievementTierPlatinum:
		return "P"
	default:
		return "?"
	}
}

// drawSelectedAchievementDetails renders details for the selected achievement.
func (ui *AchievementUI) drawSelectedAchievementDetails(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	// Details area (right 40%)
	detailsX := panelX + int(float64(panelWidth)*0.57)
	detailsY := panelY + 90
	detailsWidth := int(float64(panelWidth)*0.40) - 10
	detailsHeight := panelHeight - 140

	// Draw details background
	vector.DrawFilledRect(img, float32(detailsX), float32(detailsY), float32(detailsWidth), float32(detailsHeight),
		color.RGBA{30, 35, 45, 200}, false)
	vector.StrokeRect(img, float32(detailsX), float32(detailsY), float32(detailsWidth), float32(detailsHeight),
		1, color.RGBA{60, 60, 80, 255}, false)

	// Get selected achievement
	defs := GetAchievementDefinitionsByCategory(ui.selectedCategory)
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].ID < defs[j].ID
	})

	if ui.selectedIndex >= len(defs) {
		return
	}

	def := defs[ui.selectedIndex]
	achieveComp := ui.getPlayerAchievements()

	currentTier := AchievementTierNone
	var progress int64
	if achieveComp != nil {
		currentTier = achieveComp.GetTier(def.ID)
		progress = achieveComp.GetProgress(def.ID)
	}

	y := detailsY + 20

	// Achievement name (header)
	text.Draw(img, def.Name, basicfont.Face7x13, detailsX+10, y, color.RGBA{255, 255, 100, 255})
	y += 25

	// Description
	desc := def.Description
	if len(desc) > 35 {
		// Word wrap
		text.Draw(img, desc[:35], basicfont.Face7x13, detailsX+10, y, color.RGBA{200, 200, 200, 255})
		y += 15
		text.Draw(img, desc[35:], basicfont.Face7x13, detailsX+10, y, color.RGBA{200, 200, 200, 255})
	} else {
		text.Draw(img, desc, basicfont.Face7x13, detailsX+10, y, color.RGBA{200, 200, 200, 255})
	}
	y += 30

	// Current tier
	tierText := fmt.Sprintf("Current Tier: %s", currentTier.String())
	tierColor := getTierColor(currentTier)
	text.Draw(img, tierText, basicfont.Face7x13, detailsX+10, y, tierColor)
	y += 20

	// Current progress
	progressText := fmt.Sprintf("Progress: %d", progress)
	text.Draw(img, progressText, basicfont.Face7x13, detailsX+10, y, color.RGBA{180, 180, 180, 255})
	y += 30

	// Tier thresholds
	text.Draw(img, "Tier Thresholds:", basicfont.Face7x13, detailsX+10, y, color.RGBA{150, 150, 200, 255})
	y += 18

	tiers := []struct {
		tier  AchievementTier
		name  string
		index int
	}{
		{AchievementTierBronze, "Bronze", 0},
		{AchievementTierSilver, "Silver", 1},
		{AchievementTierGold, "Gold", 2},
		{AchievementTierPlatinum, "Platinum", 3},
	}

	for _, t := range tiers {
		threshold := def.Thresholds[t.index]
		achieved := progress >= threshold
		checkMark := "[ ]"
		var textCol color.Color = color.RGBA{120, 120, 120, 255}
		if achieved {
			checkMark = "[✓]"
			textCol = getTierColor(t.tier)
		}
		thresholdText := fmt.Sprintf("%s %s: %d", checkMark, t.name, threshold)
		text.Draw(img, thresholdText, basicfont.Face7x13, detailsX+15, y, textCol)
		y += 18
	}

	// Points for this achievement
	y += 10
	totalPts := 0
	for tier := AchievementTierBronze; tier <= currentTier; tier++ {
		totalPts += tier.Points()
	}
	pointsText := fmt.Sprintf("Points Earned: %d", totalPts)
	text.Draw(img, pointsText, basicfont.Face7x13, detailsX+10, y, color.RGBA{255, 215, 0, 255})
}

// getPlayerAchievements returns the player's achievement component.
func (ui *AchievementUI) getPlayerAchievements() *ExtendedAchievementComponent {
	if ui.playerEntity == nil {
		return nil
	}

	comp, exists := ui.playerEntity.GetComponent("extended_achievement")
	if !exists {
		return nil
	}

	achieveComp, ok := comp.(*ExtendedAchievementComponent)
	if !ok {
		return nil
	}

	return achieveComp
}

// drawFooter renders controls hint.
func (ui *AchievementUI) drawFooter(img *ebiten.Image, panelX, panelY, panelWidth, panelHeight int) {
	exitHint := GetExitHint(MenuKeys.Achievements)
	controlsText := exitHint + " | [←/→] Categories | [↑/↓] Select"
	controlsX := panelX + panelWidth/2 - len(controlsText)*3
	controlsY := panelY + panelHeight - 15
	text.Draw(img, controlsText, basicfont.Face7x13, controlsX, controlsY, color.RGBA{180, 180, 180, 255})
}

// IsActive returns whether the achievement UI is currently visible.
func (ui *AchievementUI) IsActive() bool {
	return ui.visible
}

// SetActive sets whether the achievement UI is visible.
func (ui *AchievementUI) SetActive(active bool) {
	ui.visible = active
}

// Compile-time check that AchievementUI implements UISystem
var _ UISystem = (*AchievementUI)(nil)
