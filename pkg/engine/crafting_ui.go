package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// CraftingUI handles rendering and interaction for the crafting screen.
// Displays available recipes with material requirements and crafting progress.
// Follows the same patterns as ShopUI for consistency.
type CraftingUI struct {
	visible bool

	// Entity references
	playerEntity *Entity

	// System references
	craftingSystem *CraftingSystem

	// Station reference (optional - can craft without station)
	stationEntity *Entity

	// Layout
	screenWidth    int
	screenHeight   int
	listItemHeight int
	padding        int

	// Selection
	selectedRecipeIndex int // Selected recipe in list
	hoveredRecipeIndex  int // Hovered recipe
	scrollOffset        int // For scrolling through long recipe lists

	// Crafting feedback
	craftingMessage     string
	craftingMessageTime float64 // Time remaining to show message
	showingProgress     bool    // Whether currently crafting
}

// NewCraftingUI creates a new crafting UI.
// Parameters match the pattern used by NewShopUI and NewEbitenInventoryUI.
func NewCraftingUI(screenWidth, screenHeight int) *CraftingUI {
	return &CraftingUI{
		visible:             false,
		screenWidth:         screenWidth,
		screenHeight:        screenHeight,
		listItemHeight:      80,
		padding:             15,
		selectedRecipeIndex: -1,
		hoveredRecipeIndex:  -1,
		scrollOffset:        0,
	}
}

// SetPlayerEntity sets the player entity for crafting.
func (ui *CraftingUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// SetCraftingSystem sets the crafting system for recipe execution.
func (ui *CraftingUI) SetCraftingSystem(system *CraftingSystem) {
	ui.craftingSystem = system
}

// SetStationEntity sets the crafting station entity (optional - can be nil).
func (ui *CraftingUI) SetStationEntity(entity *Entity) {
	ui.stationEntity = entity
}

// Open displays the crafting UI, optionally at a specific crafting station.
// stationEntity can be nil for crafting without station bonuses.
func (ui *CraftingUI) Open(stationEntity *Entity) {
	ui.stationEntity = stationEntity
	ui.visible = true
	ui.selectedRecipeIndex = -1
	ui.hoveredRecipeIndex = -1
	ui.scrollOffset = 0
	ui.craftingMessage = ""
	ui.craftingMessageTime = 0
	ui.showingProgress = false
}

// Close hides the crafting UI and cleans up state.
func (ui *CraftingUI) Close() {
	ui.visible = false
	ui.stationEntity = nil
	ui.selectedRecipeIndex = -1
	ui.hoveredRecipeIndex = -1
	ui.scrollOffset = 0
	ui.craftingMessage = ""
	ui.craftingMessageTime = 0
	ui.showingProgress = false
}

// IsVisible returns whether the crafting UI is currently shown.
func (ui *CraftingUI) IsVisible() bool {
	return ui.visible
}

// Toggle shows or hides the crafting UI.
func (ui *CraftingUI) Toggle() {
	ui.visible = !ui.visible
	if !ui.visible {
		ui.Close()
	}
}

// Update processes input for the crafting UI.
// Handles dual-exit navigation (C key + ESC), recipe selection (mouse/keyboard),
// and crafting initiation (ENTER/click).
func (ui *CraftingUI) Update(entities []*Entity, deltaTime float64) {
	// Update crafting message timer
	if ui.craftingMessageTime > 0 {
		ui.craftingMessageTime -= deltaTime
		if ui.craftingMessageTime < 0 {
			ui.craftingMessageTime = 0
			ui.craftingMessage = ""
		}
	}

	// Dual-exit navigation: R key (toggle) OR ESC (close only)
	// Note: Crafting uses R key (R for Recipe)
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		ui.Toggle()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && ui.visible {
		ui.Close()
		return
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Check if player is currently crafting
	if progressComp, ok := ui.playerEntity.GetComponent("crafting_progress"); ok {
		progress := progressComp.(*CraftingProgressComponent)
		if progress != nil {
			ui.showingProgress = true
			return // Don't allow new crafts while one is in progress
		}
	}
	ui.showingProgress = false

	// Get player's known recipes
	knowledgeComp, hasKnowledge := ui.playerEntity.GetComponent("recipe_knowledge")
	if !hasKnowledge {
		ui.showMessage("You don't know any recipes yet")
		return
	}
	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)

	// Convert map to slice for ordered iteration
	var recipeList []*Recipe
	for _, recipe := range knowledge.KnownRecipes {
		recipeList = append(recipeList, recipe)
	}

	if len(recipeList) == 0 {
		ui.showMessage("You don't know any recipes yet")
		return
	}

	// Calculate visible area
	windowWidth := 800
	windowHeight := 600
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	listAreaY := windowY + 120           // Below header
	listAreaHeight := windowHeight - 180 // Leave space for footer
	maxVisibleRecipes := listAreaHeight / ui.listItemHeight

	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()
	mousePressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// Check if mouse is over recipe list
	if mouseX >= windowX+ui.padding && mouseX < windowX+windowWidth-ui.padding &&
		mouseY >= listAreaY && mouseY < listAreaY+listAreaHeight {

		// Calculate which recipe is hovered
		relY := mouseY - listAreaY
		listIndex := relY / ui.listItemHeight
		recipeIndex := ui.scrollOffset + listIndex

		if recipeIndex >= 0 && recipeIndex < len(recipeList) {
			ui.hoveredRecipeIndex = recipeIndex

			// Select recipe on click
			if mousePressed {
				ui.selectedRecipeIndex = recipeIndex
			}
		}
	} else {
		ui.hoveredRecipeIndex = -1
	}

	// Handle keyboard navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if ui.selectedRecipeIndex > 0 {
			ui.selectedRecipeIndex--
			// Scroll up if needed
			if ui.selectedRecipeIndex < ui.scrollOffset {
				ui.scrollOffset = ui.selectedRecipeIndex
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if ui.selectedRecipeIndex < len(recipeList)-1 {
			ui.selectedRecipeIndex++
			// Scroll down if needed
			if ui.selectedRecipeIndex >= ui.scrollOffset+maxVisibleRecipes {
				ui.scrollOffset = ui.selectedRecipeIndex - maxVisibleRecipes + 1
			}
		} else if ui.selectedRecipeIndex == -1 && len(recipeList) > 0 {
			// Start selection at first recipe
			ui.selectedRecipeIndex = 0
		}
	}

	// Handle scrolling with mouse wheel
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 && ui.scrollOffset > 0 {
		ui.scrollOffset--
	} else if wheelY < 0 && ui.scrollOffset < len(recipeList)-maxVisibleRecipes {
		ui.scrollOffset++
	}

	// Handle crafting initiation (ENTER key or double-click)
	if ui.selectedRecipeIndex >= 0 {
		if ui.selectedRecipeIndex < len(recipeList) {
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				ui.attemptCraft(recipeList[ui.selectedRecipeIndex])
			}
		}
	}
}

// attemptCraft tries to start crafting the selected recipe.
func (ui *CraftingUI) attemptCraft(recipe *Recipe) {
	if ui.craftingSystem == nil || ui.playerEntity == nil {
		ui.showMessage("Crafting system not available")
		return
	}

	// Determine station ID (0 if no station)
	var stationID uint64 = 0
	if ui.stationEntity != nil {
		stationID = ui.stationEntity.ID
	}

	// Attempt to start crafting
	result, err := ui.craftingSystem.StartCraft(ui.playerEntity.ID, recipe, stationID)
	if err != nil {
		ui.showMessage(fmt.Sprintf("Error: %v", err))
		return
	}

	if !result.Success {
		ui.showMessage(result.ErrorMessage)
		return
	}

	// Crafting started successfully
	craftTime := recipe.CraftTimeSec
	if stationID != 0 {
		// Apply station speed bonus
		craftTime *= 0.75 // 25% faster
	}
	ui.showMessage(fmt.Sprintf("Crafting %s... (%.1fs)", recipe.Name, craftTime))
	ui.showingProgress = true
}

// showMessage displays a crafting message for 4 seconds.
func (ui *CraftingUI) showMessage(message string) {
	ui.craftingMessage = message
	ui.craftingMessageTime = 4.0
}

// Draw renders the crafting UI.
// Displays recipe list with material requirements, skill levels, and success chances.
func (ui *CraftingUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Get components
	knowledgeComp, hasKnowledge := ui.playerEntity.GetComponent("recipe_knowledge")
	skillComp, hasSkill := ui.playerEntity.GetComponent("crafting_skill")
	invComp, hasInv := ui.playerEntity.GetComponent("inventory")

	if !hasKnowledge || !hasSkill || !hasInv {
		return
	}

	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
	skill := skillComp.(*CraftingSkillComponent)
	inv := invComp.(*InventoryComponent)
	recipes := knowledge.KnownRecipes

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	img.DrawImage(overlay, nil)

	// Calculate window position
	windowWidth := 800
	windowHeight := 600
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Draw window background
	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{30, 30, 40, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	// Draw title
	titleText := "CRAFTING RECIPES"
	if ui.stationEntity != nil {
		if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
			station := stationComp.(*CraftingStationComponent)
			titleText = fmt.Sprintf("CRAFTING - %s Station (+5%% success, 25%% faster)", station.StationType.String())
		}
	}
	ebitenutil.DebugPrintAt(img, titleText, windowX+10, windowY+10)

	// Draw exit hint (standardized dual-exit navigation)
	exitHint := "Press [C] or [ESC] to close"
	ebitenutil.DebugPrintAt(img, exitHint, windowX+10, windowY+30)

	// Draw player stats
	statsText := fmt.Sprintf("Crafting Skill: %d | Gold: %d | XP: %d/%d",
		skill.SkillLevel, inv.Gold, skill.Experience, skill.ExperienceToNextLevel)
	ebitenutil.DebugPrintAt(img, statsText, windowX+10, windowY+50)

	// Draw crafting message if active
	if ui.craftingMessageTime > 0 && ui.craftingMessage != "" {
		msgColor := color.RGBA{255, 255, 100, 255}
		ebitenutil.DebugPrintAt(img, ui.craftingMessage, windowX+10, windowY+70)
		_ = msgColor // TODO: Use colored text when available
	}

	// Draw instructions
	instructionY := windowY + 90
	if ui.showingProgress {
		progressComp, _ := ui.playerEntity.GetComponent("crafting_progress")
		progress := progressComp.(*CraftingProgressComponent)
		if progress != nil {
			progressPercent := (progress.ElapsedTimeSec / progress.RequiredTimeSec) * 100
			ebitenutil.DebugPrintAt(img, fmt.Sprintf("Crafting in progress... %.0f%%", progressPercent),
				windowX+10, instructionY)
		}
	} else {
		ebitenutil.DebugPrintAt(img, "Select recipe and press ENTER/SPACE to craft", windowX+10, instructionY)
	}

	// Draw recipe list
	listAreaY := windowY + 120
	listAreaHeight := windowHeight - 180
	maxVisibleRecipes := listAreaHeight / ui.listItemHeight

	if len(recipes) == 0 {
		ebitenutil.DebugPrintAt(img, "No recipes known. Explore the world to discover recipes!",
			windowX+windowWidth/2-150, windowY+windowHeight/2)
		return
	}

	// Draw visible recipes
	// Convert map to slice for ordered iteration
	var recipeList []*Recipe
	for _, recipe := range recipes {
		recipeList = append(recipeList, recipe)
	}

	for i := 0; i < maxVisibleRecipes && (ui.scrollOffset+i) < len(recipeList); i++ {
		recipeIndex := ui.scrollOffset + i
		recipe := recipeList[recipeIndex]

		itemY := listAreaY + i*ui.listItemHeight
		itemX := windowX + ui.padding

		// Draw recipe background with selection/hover highlighting
		itemColor := color.RGBA{50, 50, 60, 255}
		if recipeIndex == ui.hoveredRecipeIndex {
			itemColor = color.RGBA{70, 70, 90, 255}
		}
		if recipeIndex == ui.selectedRecipeIndex {
			itemColor = color.RGBA{90, 90, 120, 255}
		}

		itemBg := ebiten.NewImage(windowWidth-ui.padding*2, ui.listItemHeight-5)
		itemBg.Fill(itemColor)
		itemOpts := &ebiten.DrawImageOptions{}
		itemOpts.GeoM.Translate(float64(itemX), float64(itemY))
		img.DrawImage(itemBg, itemOpts)

		// Draw recipe name and type
		nameText := fmt.Sprintf("%s [%s]", recipe.Name, recipe.Rarity.String())
		ebitenutil.DebugPrintAt(img, nameText, itemX+5, itemY+5)

		// Draw recipe description (truncated if too long)
		descText := recipe.Description
		if len(descText) > 60 {
			descText = descText[:57] + "..."
		}
		ebitenutil.DebugPrintAt(img, descText, itemX+5, itemY+20)

		// Draw skill requirement
		skillText := fmt.Sprintf("Skill Required: %d", recipe.SkillRequired)
		if skill.SkillLevel < recipe.SkillRequired {
			skillText += " (TOO LOW)"
		}
		ebitenutil.DebugPrintAt(img, skillText, itemX+5, itemY+35)

		// Draw success chance
		successChance := recipe.GetEffectiveSuccessChance(skill.SkillLevel)
		successText := fmt.Sprintf("Success: %.0f%%", successChance*100)

		// Show station bonus in success chance
		if ui.stationEntity != nil {
			if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
				station := stationComp.(*CraftingStationComponent)
				// Check if station type matches recipe type
				if station.StationType == recipe.Type {
					bonusChance := successChance + station.BonusSuccessChance
					if bonusChance > 0.95 {
						bonusChance = 0.95 // Cap at 95%
					}
					successText = fmt.Sprintf("Success: %.0f%% → %.0f%% (station +%.0f%%)",
						successChance*100, bonusChance*100, station.BonusSuccessChance*100)
				}
			}
		}

		if successChance == 0 {
			successText = "Success: Impossible (low skill)"
		}
		ebitenutil.DebugPrintAt(img, successText, itemX+200, itemY+35)

		// Draw gold cost
		goldText := fmt.Sprintf("Gold: %d", recipe.GoldCost)
		if inv.Gold < recipe.GoldCost {
			goldText += " (NOT ENOUGH)"
		}
		ebitenutil.DebugPrintAt(img, goldText, itemX+350, itemY+35)

		// Draw materials requirements
		materialsText := "Materials: "
		for j, mat := range recipe.Materials {
			// Count available materials
			available := 0
			for _, invItem := range inv.Items {
				if invItem != nil && invItem.Name == mat.ItemName {
					available++
				}
			}

			matText := fmt.Sprintf("%s (%d/%d)", mat.ItemName, available, mat.Quantity)
			if available < mat.Quantity {
				matText += "!"
			}
			materialsText += matText
			if j < len(recipe.Materials)-1 {
				materialsText += ", "
			}
		}
		// Truncate if too long
		if len(materialsText) > 75 {
			materialsText = materialsText[:72] + "..."
		}
		ebitenutil.DebugPrintAt(img, materialsText, itemX+5, itemY+50)

		// Draw craft time
		craftTimeText := fmt.Sprintf("Time: %.1fs", recipe.CraftTimeSec)
		if ui.stationEntity != nil {
			craftTimeText = fmt.Sprintf("Time: %.1fs (station bonus)", recipe.CraftTimeSec*0.75)
		}
		ebitenutil.DebugPrintAt(img, craftTimeText, itemX+5, itemY+65)
	}

	// Draw scroll indicator if needed
	if len(recipeList) > maxVisibleRecipes {
		scrollText := fmt.Sprintf("Scroll: %d-%d / %d recipes",
			ui.scrollOffset+1,
			minInt(ui.scrollOffset+maxVisibleRecipes, len(recipeList)),
			len(recipeList))
		ebitenutil.DebugPrintAt(img, scrollText, windowX+windowWidth-200, windowY+windowHeight-30)
	}

	// Draw footer hints
	footerY := windowY + windowHeight - 30
	ebitenutil.DebugPrintAt(img, "Arrow Keys: Navigate | ENTER/SPACE: Craft | Mouse Wheel: Scroll",
		windowX+10, footerY)

	// Draw nearby station hint if not at a station
	if ui.stationEntity == nil && ui.playerEntity != nil {
		if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
			pos := posComp.(*PositionComponent)
			// Find nearest station within 100 pixels
			nearestStation, distance := ui.findNearestStation(pos.X, pos.Y, 100)
			if nearestStation != nil {
				if stationComp, ok := nearestStation.GetComponent("crafting_station"); ok {
					station := stationComp.(*CraftingStationComponent)
					stationHint := fmt.Sprintf("Nearby: %s (%.0f units away) - Move closer to use station bonuses",
						station.StationType.String(), distance)
					ebitenutil.DebugPrintAt(img, stationHint, windowX+10, footerY-20)
				}
			}
		}
	}
}

// findNearestStation finds the nearest crafting station within maxDistance.
// Returns nil if no station is found within range.
// Uses the same logic as FindClosestStation from station_spawn.go but operates on World entities.
func (ui *CraftingUI) findNearestStation(centerX, centerY, maxDistance float64) (*Entity, float64) {
	if ui.craftingSystem == nil || ui.craftingSystem.world == nil {
		return nil, 0
	}

	entities := ui.craftingSystem.world.GetEntities()

	// Convert []*Entity to []Entity for FindClosestStation
	entitySlice := make([]Entity, len(entities))
	for i, e := range entities {
		if e != nil {
			entitySlice[i] = *e
		}
	}

	return FindClosestStation(entitySlice, centerX, centerY, maxDistance)
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
