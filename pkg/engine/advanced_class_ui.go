package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/opd-ai/venture/pkg/class/advanced"
	"golang.org/x/image/font/basicfont"
)

// AdvancedClassUI provides the user interface for advanced class features:
// multi-classing, prestige classes, and talent trees.
type AdvancedClassUI struct {
	visible          bool
	world            *World
	system           *AdvancedClassSystem
	screenWidth      int
	screenHeight     int
	selectedTab      int // 0=talents, 1=classes, 2=prestige
	selectedTalent   advanced.TalentID
	selectedCategory advanced.TalentCategory
	currentPlayerID  string
	scrollOffset     int
	confirmRespec    bool
}

// NewAdvancedClassUI creates a new advanced class UI
func NewAdvancedClassUI(world *World, system *AdvancedClassSystem, screenWidth, screenHeight int) *AdvancedClassUI {
	return &AdvancedClassUI{
		world:            world,
		system:           system,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		selectedCategory: advanced.CategoryOffensive,
	}
}

// SetVisible sets the UI visibility
func (ui *AdvancedClassUI) SetVisible(visible bool) {
	ui.visible = visible
}

// IsVisible returns whether the UI is visible
func (ui *AdvancedClassUI) IsVisible() bool {
	return ui.visible
}

// Toggle toggles the UI visibility
func (ui *AdvancedClassUI) Toggle() {
	ui.visible = !ui.visible
}

// SetPlayerEntity sets the current player entity for the UI
func (ui *AdvancedClassUI) SetPlayerEntity(entity *Entity) {
	if entity != nil {
		ui.currentPlayerID = fmt.Sprintf("%d", entity.ID)
	}
}

// Update processes input for the advanced class UI
func (ui *AdvancedClassUI) Update() error {
	if !ui.visible {
		return nil
	}

	ui.handleTabSwitching()
	ui.handleCategoryAndScrolling()
	ui.handleRespecConfirmation()

	return nil
}

// Platform parity fix: Uses edge-triggered detection for tab switching
func (ui *AdvancedClassUI) handleTabSwitching() {
	// Handle touch/mouse tab selection
	// Match coordinates from drawTabs(): tabWidth=120, startX=70, spacing=10, y=60
	if IsTouchOrMouseJustPressed() {
		mouseX, mouseY, _ := GetTouchOrMousePosition()
		tabWidth := 120
		tabHeight := 30
		startX := 70
		tabY := 60
		tabSpacing := 10

		if mouseY >= tabY && mouseY <= tabY+tabHeight {
			for i := 0; i < 3; i++ {
				tabX := startX + i*(tabWidth+tabSpacing)
				if mouseX >= tabX && mouseX <= tabX+tabWidth {
					ui.selectedTab = i
					break
				}
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		ui.selectedTab = 0
	} else if inpututil.IsKeyJustPressed(ebiten.Key2) {
		ui.selectedTab = 1
	} else if inpututil.IsKeyJustPressed(ebiten.Key3) {
		ui.selectedTab = 2
	}
}

func (ui *AdvancedClassUI) handleCategoryAndScrolling() {
	if ui.selectedTab == 0 {
		ui.handleCategorySwitching()
	}
	ui.handleScrollInput()
}

// Platform parity fix: Uses edge-triggered detection for category switching
func (ui *AdvancedClassUI) handleCategorySwitching() {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		ui.selectedCategory = advanced.CategoryOffensive
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		ui.selectedCategory = advanced.CategoryDefensive
	} else if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		ui.selectedCategory = advanced.CategoryUtility
	}
}

// Platform parity fix: Uses edge-triggered detection for scroll input
func (ui *AdvancedClassUI) handleScrollInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		ui.scrollOffset = max(0, ui.scrollOffset-1)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		ui.scrollOffset++
	}
}

// Platform parity fix: Uses edge-triggered detection for respec confirmation
func (ui *AdvancedClassUI) handleRespecConfirmation() {
	if ui.selectedTab != 0 {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if ui.confirmRespec {
			ui.handleRespec()
			ui.confirmRespec = false
		} else {
			ui.confirmRespec = true
		}
	}
}

// Draw renders the advanced class UI
func (ui *AdvancedClassUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Draw background panel
	ui.drawPanel(screen, 50, 50, ui.screenWidth-100, ui.screenHeight-100, color.RGBA{20, 20, 30, 220})

	// Draw tabs
	ui.drawTabs(screen)

	// Draw content based on selected tab
	switch ui.selectedTab {
	case 0:
		ui.drawTalentsTab(screen)
	case 1:
		ui.drawClassesTab(screen)
	case 2:
		ui.drawPrestigeTab(screen)
	}

	// Draw help text
	ui.drawHelpText(screen)
}

// drawTabs draws the tab navigation
func (ui *AdvancedClassUI) drawTabs(screen *ebiten.Image) {
	tabs := []string{"Talents [1]", "Classes [2]", "Prestige [3]"}
	tabWidth := 120
	startX := 70

	for i, tabName := range tabs {
		x := startX + (i * (tabWidth + 10))
		y := 60

		bgColor := color.RGBA{40, 40, 50, 220}
		if i == ui.selectedTab {
			bgColor = color.RGBA{60, 60, 80, 220}
		}

		ui.drawPanel(screen, x, y, tabWidth, 30, bgColor)
		text.Draw(screen, tabName, basicfont.Face7x13, x+10, y+20, color.White)
	}
}

// drawTalentsTab draws the talent tree interface
func (ui *AdvancedClassUI) drawTalentsTab(screen *ebiten.Image) {
	player := ui.getPlayerEntity()
	if player == nil {
		text.Draw(screen, "No player entity", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	comp, ok := player.GetComponent("advanced_class")
	if !ok || comp == nil {
		text.Draw(screen, "No advanced class", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	advClass := comp.(*advanced.AdvancedClassComponent)

	// Draw points available
	available := advClass.TalentPoints.PointsTotal - advClass.TalentPoints.PointsSpent
	pointsText := fmt.Sprintf("Points: %d / %d", advClass.TalentPoints.PointsSpent, advClass.TalentPoints.PointsTotal)
	text.Draw(screen, pointsText, basicfont.Face7x13, 70, 130, color.RGBA{255, 215, 0, 255})

	// Draw category buttons
	ui.drawCategoryButtons(screen)

	// Draw talent tree
	tree, err := ui.system.GetTalentTree(advClass.PrimaryClass)
	if err != nil {
		text.Draw(screen, fmt.Sprintf("Error: %v", err), basicfont.Face7x13, 70, 180, color.RGBA{255, 0, 0, 255})
		return
	}

	ui.drawTalentList(screen, tree, advClass, available)

	// Draw respec info
	cost := ui.system.GetRespecCost(player)
	respecText := fmt.Sprintf("Press R to respec (%d gold)", cost)
	if ui.confirmRespec {
		respecText = "Press R again to confirm respec"
	}
	text.Draw(screen, respecText, basicfont.Face7x13, 70, ui.screenHeight-80, color.RGBA{200, 200, 100, 255})
}

// drawCategoryButtons draws the talent category selection buttons
func (ui *AdvancedClassUI) drawCategoryButtons(screen *ebiten.Image) {
	categories := []struct {
		cat  advanced.TalentCategory
		name string
		key  string
	}{
		{advanced.CategoryOffensive, "Offensive", "[Q]"},
		{advanced.CategoryDefensive, "Defensive", "[W]"},
		{advanced.CategoryUtility, "Utility", "[E]"},
	}

	startY := 150
	for i, cat := range categories {
		y := startY + (i * 40)
		bgColor := color.RGBA{40, 40, 50, 220}
		if cat.cat == ui.selectedCategory {
			bgColor = color.RGBA{80, 80, 100, 220}
		}

		ui.drawPanel(screen, 70, y, 150, 30, bgColor)
		text.Draw(screen, fmt.Sprintf("%s %s", cat.name, cat.key), basicfont.Face7x13, 80, y+20, color.White)
	}
}

// drawTalentList draws the list of talents in the selected category
func (ui *AdvancedClassUI) drawTalentList(screen *ebiten.Image, tree *advanced.TalentTree, advClass *advanced.AdvancedClassComponent, availablePoints int) {
	var talents []advanced.TalentDefinition

	switch ui.selectedCategory {
	case advanced.CategoryOffensive:
		talents = tree.Offensive
	case advanced.CategoryDefensive:
		talents = tree.Defensive
	case advanced.CategoryUtility:
		talents = tree.Utility
	}

	startX := 250
	startY := 150
	y := startY

	for i, talent := range talents {
		if i < ui.scrollOffset {
			continue
		}
		if y > ui.screenHeight-120 {
			break
		}

		currentRank := advClass.TalentPoints.Talents[talent.ID]
		rankText := fmt.Sprintf("%d/%d", currentRank, talent.MaxRank)

		// Color based on state
		textColor := color.RGBA{150, 150, 150, 255} // Not learned
		if currentRank > 0 {
			textColor = color.RGBA{0, 255, 0, 255} // Learned
		} else if availablePoints > 0 && ui.canLearnTalent(talent, advClass) {
			textColor = color.RGBA{255, 215, 0, 255} // Can learn
		}

		talentText := fmt.Sprintf("%s %s", talent.Name, rankText)
		text.Draw(screen, talentText, basicfont.Face7x13, startX, y, textColor)

		// Draw description
		descText := talent.Description
		if len(descText) > 50 {
			descText = descText[:50] + "..."
		}
		text.Draw(screen, descText, basicfont.Face7x13, startX+10, y+15, color.RGBA{180, 180, 180, 255})

		y += 45
	}
}

// canLearnTalent checks if a talent can be learned
func (ui *AdvancedClassUI) canLearnTalent(talent advanced.TalentDefinition, advClass *advanced.AdvancedClassComponent) bool {
	// Check rank limit
	currentRank := advClass.TalentPoints.Talents[talent.ID]
	if currentRank >= talent.MaxRank {
		return false
	}

	// Check prerequisites
	for _, prereq := range talent.Prerequisites {
		if advClass.TalentPoints.Talents[prereq] == 0 {
			return false
		}
	}

	return true
}

// drawClassesTab draws the multi-classing interface
func (ui *AdvancedClassUI) drawClassesTab(screen *ebiten.Image) {
	player := ui.getPlayerEntity()
	if player == nil {
		text.Draw(screen, "No player entity", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	comp, ok := player.GetComponent("advanced_class")
	if !ok || comp == nil {
		text.Draw(screen, "No advanced class", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	advClass := comp.(*advanced.AdvancedClassComponent)

	// Draw current classes
	y := 130
	text.Draw(screen, "Primary Class:", basicfont.Face7x13, 70, y, color.White)
	y += 20
	text.Draw(screen, fmt.Sprintf("  %s (Level %d)", advClass.PrimaryClass, advClass.Level), basicfont.Face7x13, 70, y, color.RGBA{0, 255, 0, 255})
	y += 30

	if advClass.SecondaryClass != "" {
		text.Draw(screen, "Secondary Class:", basicfont.Face7x13, 70, y, color.White)
		y += 20
		text.Draw(screen, fmt.Sprintf("  %s", advClass.SecondaryClass), basicfont.Face7x13, 70, y, color.RGBA{0, 200, 255, 255})
	} else {
		if advClass.Level >= 20 {
			text.Draw(screen, "Secondary class unlocked at level 20!", basicfont.Face7x13, 70, y, color.RGBA{255, 215, 0, 255})
		} else {
			text.Draw(screen, fmt.Sprintf("Secondary class unlocks at level 20 (current: %d)", advClass.Level), basicfont.Face7x13, 70, y, color.RGBA{150, 150, 150, 255})
		}
	}

	// Draw synergy bonuses
	y += 50
	ui.drawSynergyInfo(screen, y)
}

// drawSynergyInfo draws information about class synergies
func (ui *AdvancedClassUI) drawSynergyInfo(screen *ebiten.Image, startY int) {
	text.Draw(screen, "Class Synergies:", basicfont.Face7x13, 70, startY, color.White)
	startY += 20

	synergies := ui.system.GetAllSynergies()
	displayCount := 0
	for _, synergy := range synergies {
		if displayCount >= 5 {
			break
		}
		synergyText := fmt.Sprintf("%s + %s = %s", synergy.Primary, synergy.Secondary, synergy.Name)
		text.Draw(screen, synergyText, basicfont.Face7x13, 80, startY, color.RGBA{200, 200, 255, 255})
		startY += 20
		displayCount++
	}
}

// drawPrestigeTab draws the prestige class interface
func (ui *AdvancedClassUI) drawPrestigeTab(screen *ebiten.Image) {
	player := ui.getPlayerEntity()
	if player == nil {
		text.Draw(screen, "No player entity", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	comp, ok := player.GetComponent("advanced_class")
	if !ok || comp == nil {
		text.Draw(screen, "No advanced class", basicfont.Face7x13, 70, 130, color.White)
		return
	}

	advClass := comp.(*advanced.AdvancedClassComponent)

	y := 130
	if advClass.PrestigeClass != "" {
		text.Draw(screen, "Current Prestige Class:", basicfont.Face7x13, 70, y, color.White)
		y += 20
		text.Draw(screen, fmt.Sprintf("  %s", advClass.PrestigeClass), basicfont.Face7x13, 70, y, color.RGBA{255, 215, 0, 255})
	} else {
		if advClass.Level >= 20 {
			text.Draw(screen, "Prestige classes available!", basicfont.Face7x13, 70, y, color.RGBA{255, 215, 0, 255})
			y += 20
			text.Draw(screen, "Visit a class trainer to unlock", basicfont.Face7x13, 70, y, color.White)
		} else {
			text.Draw(screen, fmt.Sprintf("Prestige classes unlock at level 20 (current: %d)", advClass.Level), basicfont.Face7x13, 70, y, color.RGBA{150, 150, 150, 255})
		}
	}

	// Draw available prestige classes
	y += 40
	text.Draw(screen, "Available Prestige Classes:", basicfont.Face7x13, 70, y, color.White)
	y += 20

	prestigeClasses := []string{
		"Blade Master - Master of melee combat",
		"Champion - Defensive warrior",
		"Dragon Knight - Dragon-powered warrior",
		"Dreadnought - Unstoppable tank",
		"Archmage - Master of all magic",
		"Soul Reaper - Death magic specialist",
	}

	for i, pc := range prestigeClasses {
		if i >= 6 {
			break
		}
		text.Draw(screen, pc, basicfont.Face7x13, 80, y, color.RGBA{200, 200, 255, 255})
		y += 20
	}
}

// drawHelpText draws help text at the bottom
func (ui *AdvancedClassUI) drawHelpText(screen *ebiten.Image) {
	helpText := "ESC: Close | 1/2/3: Switch tabs | Arrow Keys: Scroll"
	if ui.selectedTab == 0 {
		helpText = "ESC: Close | Q/W/E: Categories | R: Respec | Arrow Keys: Scroll"
	}
	text.Draw(screen, helpText, basicfont.Face7x13, 70, ui.screenHeight-50, color.RGBA{150, 150, 150, 255})
}

// drawPanel draws a colored panel
func (ui *AdvancedClassUI) drawPanel(screen *ebiten.Image, x, y, width, height int, col color.Color) {
	img := ebiten.NewImage(width, height)
	img.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// getPlayerEntity finds the player entity
func (ui *AdvancedClassUI) getPlayerEntity() *Entity {
	if ui.currentPlayerID != "" {
		// Try to find entity by ID in world
		entities := ui.world.GetEntities()
		for _, e := range entities {
			if fmt.Sprintf("%d", e.ID) == ui.currentPlayerID {
				return e
			}
		}
	}
	return nil
}

// handleRespec performs talent respec
func (ui *AdvancedClassUI) handleRespec() {
	player := ui.getPlayerEntity()
	if player == nil {
		return
	}

	// In a real implementation, we'd check gold from an inventory component
	// For now, assume player has enough gold
	goldAmount := 10000

	err := ui.system.RespecTalents(player, goldAmount)
	if err != nil {
		// Could show error message in UI
		return
	}
}

// UpdateWithTouch processes touch input for mobile
func (ui *AdvancedClassUI) UpdateWithTouch(touches []ebiten.TouchID) error {
	if !ui.visible {
		return nil
	}

	for _, id := range touches {
		x, y := ebiten.TouchPosition(id)
		// Handle touch interactions
		// Tab selection
		if y >= 60 && y <= 90 {
			if x >= 70 && x < 190 {
				ui.selectedTab = 0
			} else if x >= 200 && x < 320 {
				ui.selectedTab = 1
			} else if x >= 330 && x < 450 {
				ui.selectedTab = 2
			}
		}
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
