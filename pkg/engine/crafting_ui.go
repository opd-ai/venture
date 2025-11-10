package engine

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/opd-ai/venture/pkg/mobile"
	"golang.org/x/image/font/basicfont"
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

	// Issue #12 FIX: Search and filter functionality
	searchQuery     string     // Search text for recipe names
	filterCategory  RecipeType // Filter by recipe type (use -1 for "All")
	sortMode        int        // 0=Name, 1=Tier, 2=Craftable
	showOnlyCrafted bool       // Show only recipes player can craft

	// Mobile keyboard state (WASM/mobile platforms)
	keyboardShown bool // Tracks whether mobile keyboard is currently shown

	// Crafting feedback
	craftingMessage     string
	craftingMessageTime float64 // Time remaining to show message
	showingProgress     bool    // Whether currently crafting

	// H-002 FIX: Error feedback
	errorState *UIErrorState
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
		searchQuery:         "", // Issue #12: Initialize search/filter
		filterCategory:      -1, // -1 means "All categories"
		sortMode:            0,  // 0=Name (default)
		showOnlyCrafted:     false,
		errorState:          NewUIErrorState(), // H-002 FIX
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
	// Issue #12: Reset search/filter on open
	ui.searchQuery = ""
	ui.filterCategory = -1
	ui.sortMode = 0
	ui.showOnlyCrafted = false

	// MOBILE/WASM: Show keyboard when opening crafting UI (search field active)
	// The crafting UI search is always active, so we show keyboard immediately.
	// Note: We track keyboard state to avoid redundant calls. The keyboard is
	// managed by this UI component and should only be shown/hidden by Open/Close.
	if !ui.keyboardShown && mobile.IsWASM() {
		mobile.ShowKeyboard()
		ui.keyboardShown = true
	}
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

	// MOBILE/WASM: Hide keyboard when closing crafting UI
	if ui.keyboardShown && mobile.IsWASM() {
		mobile.HideKeyboard()
		ui.keyboardShown = false
	}
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
// Handles dual-exit navigation (R key + ESC), recipe selection (mouse/keyboard),
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

	// Standardized dual-exit menu navigation: toggle key (R) OR Escape
	if shouldClose, shouldToggle := HandleMenuInput(MenuKeys.Crafting, ui.visible); shouldClose {
		if shouldToggle {
			ui.Toggle()
		} else {
			ui.Close()
		}
		return // Don't process other input on the same frame as toggle/close
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Check if player is currently crafting
	if progressComp, ok := ui.playerEntity.GetComponent("crafting_progress"); ok {
		if progress, ok := progressComp.(*CraftingProgressComponent); ok {
			if progress != nil {
				ui.showingProgress = true
				return // Don't allow new crafts while one is in progress
			}
		}
	}
	ui.showingProgress = false

	// Get player's known recipes
	knowledgeComp, hasKnowledge := ui.playerEntity.GetComponent("recipe_knowledge")
	if !hasKnowledge {
		ui.showMessage("You don't know any recipes yet")
		return
	}
	knowledge, ok := knowledgeComp.(*RecipeKnowledgeComponent)
	if !ok {
		return
	}

	// Convert map to slice for ordered iteration
	var recipeList []*Recipe
	for _, recipe := range knowledge.KnownRecipes {
		recipeList = append(recipeList, recipe)
	}

	if len(recipeList) == 0 {
		ui.showMessage("You don't know any recipes yet")
		return
	}

	// Issue #12 FIX: Handle search/filter input
	// Tab key cycles through categories: All -> Potion -> Enchanting -> MagicItem -> All
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		ui.filterCategory++
		if ui.filterCategory > RecipeMagicItem {
			ui.filterCategory = -1 // Back to "All"
		}
		ui.scrollOffset = 0
		ui.selectedRecipeIndex = -1
	}

	// F key cycles sort modes: Name -> Tier -> Craftable -> Name
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ui.sortMode = (ui.sortMode + 1) % 3
		ui.scrollOffset = 0
	}

	// C key toggles craftable-only filter
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		ui.showOnlyCrafted = !ui.showOnlyCrafted
		ui.scrollOffset = 0
		ui.selectedRecipeIndex = -1
	}

	// Backspace removes last character from search
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(ui.searchQuery) > 0 {
			ui.searchQuery = ui.searchQuery[:len(ui.searchQuery)-1]
			ui.scrollOffset = 0
		}
	}

	// Handle text input for search (only letters, numbers, spaces)
	inputChars := ebiten.AppendInputChars(nil)
	for _, char := range inputChars {
		// Allow alphanumeric and space
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == ' ' {
			if len(ui.searchQuery) < 30 { // Max 30 chars
				ui.searchQuery += string(char)
				ui.scrollOffset = 0
			}
		}
	}

	// Apply search/filter/sort to get filtered recipe list
	recipeList = ui.filterAndSortRecipes(recipeList)

	if len(recipeList) == 0 {
		ui.showMessage("No recipes match your search/filter")
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

	// Handle mouse and touch input (Touch support for WASM/mobile)
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	mousePressed := IsTouchOrMouseJustPressed()

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
		// H-002 FIX: Use error state for consistent feedback
		ui.errorState.ShowError("Crafting system not available")
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
		// H-002 FIX: Use error state for consistent feedback
		ui.errorState.ShowError(fmt.Sprintf("Crafting failed: %v", err))
		return
	}

	if !result.Success {
		// H-002 FIX: Use error state for consistent feedback
		ui.errorState.ShowError(result.ErrorMessage)
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
// craftingComponents holds the player's crafting-related components.
type craftingComponents struct {
	knowledge *RecipeKnowledgeComponent
	skill     *CraftingSkillComponent
	inv       *InventoryComponent
}

// getPlayerCraftingComponents retrieves and validates the player's crafting components.
func (ui *CraftingUI) getPlayerCraftingComponents() (*craftingComponents, bool) {
	if ui.playerEntity == nil {
		return nil, false
	}

	knowledgeComp, hasKnowledge := ui.playerEntity.GetComponent("recipe_knowledge")
	skillComp, hasSkill := ui.playerEntity.GetComponent("crafting_skill")
	invComp, hasInv := ui.playerEntity.GetComponent("inventory")

	if !hasKnowledge || !hasSkill || !hasInv {
		return nil, false
	}

	knowledge, ok := knowledgeComp.(*RecipeKnowledgeComponent)
	if !ok {
		return nil, false
	}
	skill, ok := skillComp.(*CraftingSkillComponent)
	if !ok {
		return nil, false
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil, false
	}

	return &craftingComponents{
		knowledge: knowledge,
		skill:     skill,
		inv:       inv,
	}, true
}

// drawWindowBackground draws the semi-transparent overlay and window background.
// Returns the window X and Y coordinates.
func (ui *CraftingUI) drawWindowBackground(img *ebiten.Image) (int, int) {
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

	return windowX, windowY
}

// drawTitleBar draws the title text and exit hint.
func (ui *CraftingUI) drawTitleBar(img *ebiten.Image, windowX, windowY int) {
	titleText := "CRAFTING RECIPES"
	if ui.stationEntity != nil {
		if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
			if station, ok := stationComp.(*CraftingStationComponent); ok {
				titleText = fmt.Sprintf("CRAFTING - %s Station (+5%% success, 25%% faster)", station.StationType.String())
			}
		}
	}
	ebitenutil.DebugPrintAt(img, titleText, windowX+10, windowY+10)

	// Draw exit hint (standardized dual-exit navigation)
	exitHint := GetExitHint(MenuKeys.Crafting)
	ebitenutil.DebugPrintAt(img, exitHint, windowX+10, windowY+30)
}

// drawPlayerStats draws the player's crafting stats (skill, gold, XP).
func (ui *CraftingUI) drawPlayerStats(img *ebiten.Image, components *craftingComponents, windowX, windowY int) {
	statsText := fmt.Sprintf("Crafting Skill: %d | Gold: %d | XP: %d/%d",
		components.skill.SkillLevel, components.inv.Gold, components.skill.Experience, components.skill.ExperienceToNextLevel)
	ebitenutil.DebugPrintAt(img, statsText, windowX+10, windowY+50)
}

// drawCraftingMessage draws the active crafting message with appropriate color.
func (ui *CraftingUI) drawCraftingMessage(img *ebiten.Image, windowX, windowY int) {
	if ui.craftingMessageTime <= 0 || ui.craftingMessage == "" {
		return
	}

	// Determine message color based on content
	msgColor := color.RGBA{100, 255, 100, 255} // Green for success
	if strings.Contains(ui.craftingMessage, "failed") ||
		strings.Contains(ui.craftingMessage, "cannot") ||
		strings.Contains(ui.craftingMessage, "not available") {
		msgColor = color.RGBA{255, 100, 100, 255} // Red for errors
	} else if strings.Contains(ui.craftingMessage, "progress") {
		msgColor = color.RGBA{255, 255, 100, 255} // Yellow for in-progress
	}

	// Use text.Draw with colored font instead of DebugPrintAt
	text.Draw(img, ui.craftingMessage, basicfont.Face7x13, windowX+10, windowY+80, msgColor)
}

// drawInstructions draws the crafting instructions or progress indicator.
func (ui *CraftingUI) drawInstructions(img *ebiten.Image, windowX, windowY int) {
	instructionY := windowY + 90
	if ui.showingProgress {
		progressComp, _ := ui.playerEntity.GetComponent("crafting_progress")
		if progressComp != nil {
			if progress, ok := progressComp.(*CraftingProgressComponent); ok {
				if progress != nil {
					progressPercent := (progress.ElapsedTimeSec / progress.RequiredTimeSec) * 100
					ebitenutil.DebugPrintAt(img, fmt.Sprintf("Crafting in progress... %.0f%%", progressPercent),
						windowX+10, instructionY)
				}
			}
		}
	} else {
		ebitenutil.DebugPrintAt(img, "Select recipe and press ENTER/SPACE to craft", windowX+10, instructionY)
	}
}

// drawSearchFilterUI draws the search bar and filter controls.
func (ui *CraftingUI) drawSearchFilterUI(img *ebiten.Image, windowX, windowY int) {
	filterY := windowY + 110

	// Search bar
	searchText := fmt.Sprintf("Search: %s_", ui.searchQuery)
	if len(ui.searchQuery) == 0 {
		searchText = "Search: (type to search)"
	}
	ebitenutil.DebugPrintAt(img, searchText, windowX+10, filterY)

	// Category filter
	categoryName := "All"
	if ui.filterCategory == RecipePotion {
		categoryName = "Potions"
	} else if ui.filterCategory == RecipeEnchanting {
		categoryName = "Enchanting"
	} else if ui.filterCategory == RecipeMagicItem {
		categoryName = "Magic Items"
	}
	filterText := fmt.Sprintf("[TAB] Category: %s", categoryName)
	ebitenutil.DebugPrintAt(img, filterText, windowX+280, filterY)

	// Sort mode
	sortName := "Name"
	if ui.sortMode == 1 {
		sortName = "Tier"
	} else if ui.sortMode == 2 {
		sortName = "Craftable"
	}
	sortText := fmt.Sprintf("[F] Sort: %s", sortName)
	ebitenutil.DebugPrintAt(img, sortText, windowX+480, filterY)

	// Craftable filter
	craftableText := "[C] Show All"
	if ui.showOnlyCrafted {
		craftableText = "[C] Only Craftable"
	}
	ebitenutil.DebugPrintAt(img, craftableText, windowX+620, filterY)
}

// drawRecipeList draws the list of recipes with all their details.
func (ui *CraftingUI) drawRecipeList(img *ebiten.Image, recipes map[string]*Recipe, components *craftingComponents, windowX, windowY, windowWidth, windowHeight int) {
	listAreaY := windowY + 140
	listAreaHeight := windowHeight - 200
	maxVisibleRecipes := listAreaHeight / ui.listItemHeight

	if len(recipes) == 0 {
		ebitenutil.DebugPrintAt(img, "No recipes known. Explore the world to discover recipes!",
			windowX+windowWidth/2-150, windowY+windowHeight/2)
		return
	}

	// Convert map to slice and apply filters/sorting
	var recipeList []*Recipe
	for _, recipe := range recipes {
		recipeList = append(recipeList, recipe)
	}

	// Apply search/filter/sort
	recipeList = ui.filterAndSortRecipes(recipeList)

	if len(recipeList) == 0 {
		ui.drawEmptyRecipeListMessage(img, windowX, windowY, windowWidth, windowHeight)
		return
	}

	// Draw each visible recipe
	for i := 0; i < maxVisibleRecipes && (ui.scrollOffset+i) < len(recipeList); i++ {
		recipeIndex := ui.scrollOffset + i
		recipe := recipeList[recipeIndex]
		itemY := listAreaY + i*ui.listItemHeight
		ui.drawRecipeItem(img, recipe, recipeIndex, components, windowX, itemY, windowWidth)
	}

	// Draw scroll indicator and footer
	ui.drawRecipeListFooter(img, recipeList, maxVisibleRecipes, windowX, windowY, windowWidth, windowHeight)
}

// drawEmptyRecipeListMessage draws a message when the recipe list is empty due to filters.
func (ui *CraftingUI) drawEmptyRecipeListMessage(img *ebiten.Image, windowX, windowY, windowWidth, windowHeight int) {
	emptyMsg := "No recipes match your search/filter criteria"
	if ui.searchQuery != "" {
		emptyMsg = fmt.Sprintf("No recipes match search: \"%s\"", ui.searchQuery)
	} else if ui.filterCategory != -1 {
		emptyMsg = fmt.Sprintf("No %s recipes available", ui.filterCategory.String())
	} else if ui.showOnlyCrafted {
		emptyMsg = "No craftable recipes (missing materials or gold)"
	}

	// Center the message
	msgLen := len(emptyMsg)
	msgX := windowX + (windowWidth-msgLen*7)/2
	ebitenutil.DebugPrintAt(img, emptyMsg, msgX, windowY+windowHeight/2)

	// Show hint about clearing filters
	hintMsg := "Press [Backspace] to clear search or [C] to show all recipes"
	hintX := windowX + (windowWidth-len(hintMsg)*7)/2
	ebitenutil.DebugPrintAt(img, hintMsg, hintX, windowY+windowHeight/2+20)
}

// drawRecipeItem draws a single recipe item in the list.
func (ui *CraftingUI) drawRecipeItem(img *ebiten.Image, recipe *Recipe, recipeIndex int, components *craftingComponents, windowX, itemY, windowWidth int) {
	itemX := windowX + ui.padding

	// Draw recipe background with highlighting
	isCraftable := ui.canCraftRecipe(recipe)
	itemColor := ui.getRecipeItemColor(recipeIndex, isCraftable)

	itemBg := ebiten.NewImage(windowWidth-ui.padding*2, ui.listItemHeight-5)
	itemBg.Fill(itemColor)
	itemOpts := &ebiten.DrawImageOptions{}
	itemOpts.GeoM.Translate(float64(itemX), float64(itemY))
	img.DrawImage(itemBg, itemOpts)

	// Draw recipe details
	ui.drawRecipeItemDetails(img, recipe, components, itemX, itemY)
}

// getRecipeItemColor determines the background color for a recipe item.
func (ui *CraftingUI) getRecipeItemColor(recipeIndex int, isCraftable bool) color.RGBA {
	itemColor := color.RGBA{50, 50, 60, 255}
	if isCraftable {
		itemColor = color.RGBA{50, 70, 50, 255} // Green tint for craftable
	}
	if recipeIndex == ui.hoveredRecipeIndex {
		itemColor = color.RGBA{70, 70, 90, 255}
		if isCraftable {
			itemColor = color.RGBA{70, 90, 70, 255} // Green-tinted hover
		}
	}
	if recipeIndex == ui.selectedRecipeIndex {
		itemColor = color.RGBA{90, 90, 120, 255}
		if isCraftable {
			itemColor = color.RGBA{90, 120, 90, 255} // Green-tinted selection
		}
	}
	return itemColor
}

// drawRecipeItemDetails draws all the text details for a recipe item.
func (ui *CraftingUI) drawRecipeItemDetails(img *ebiten.Image, recipe *Recipe, components *craftingComponents, itemX, itemY int) {
	// Recipe name and rarity
	nameText := fmt.Sprintf("%s [%s]", recipe.Name, recipe.Rarity.String())
	ebitenutil.DebugPrintAt(img, nameText, itemX+5, itemY+5)

	// Recipe description
	descText := recipe.Description
	if len(descText) > 60 {
		descText = descText[:57] + "..."
	}
	ebitenutil.DebugPrintAt(img, descText, itemX+5, itemY+20)

	// Skill requirement
	skillText := fmt.Sprintf("Skill Required: %d", recipe.SkillRequired)
	if components.skill.SkillLevel < recipe.SkillRequired {
		skillText += " (TOO LOW)"
	}
	ebitenutil.DebugPrintAt(img, skillText, itemX+5, itemY+35)

	// Success chance
	successText := ui.getSuccessChanceText(recipe, components.skill)
	ebitenutil.DebugPrintAt(img, successText, itemX+200, itemY+35)

	// Gold cost
	goldText := fmt.Sprintf("Gold: %d", recipe.GoldCost)
	if components.inv.Gold < recipe.GoldCost {
		goldText += " (NOT ENOUGH)"
	}
	ebitenutil.DebugPrintAt(img, goldText, itemX+350, itemY+35)

	// Materials
	materialsText := ui.getMaterialsText(recipe, components.inv)
	ebitenutil.DebugPrintAt(img, materialsText, itemX+5, itemY+50)

	// Craft time
	craftTimeText := fmt.Sprintf("Time: %.1fs", recipe.CraftTimeSec)
	if ui.stationEntity != nil {
		craftTimeText = fmt.Sprintf("Time: %.1fs (station bonus)", recipe.CraftTimeSec*0.75)
	}
	ebitenutil.DebugPrintAt(img, craftTimeText, itemX+5, itemY+65)
}

// getSuccessChanceText calculates and formats the success chance text.
func (ui *CraftingUI) getSuccessChanceText(recipe *Recipe, skill *CraftingSkillComponent) string {
	successChance := recipe.GetEffectiveSuccessChance(skill.SkillLevel)
	successText := fmt.Sprintf("Success: %.0f%%", successChance*100)

	// Show station bonus if applicable
	if ui.stationEntity != nil {
		if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
			if station, ok := stationComp.(*CraftingStationComponent); ok {
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
	}

	if successChance == 0 {
		successText = "Success: Impossible (low skill)"
	}
	return successText
}

// getMaterialsText formats the materials requirements text.
func (ui *CraftingUI) getMaterialsText(recipe *Recipe, inv *InventoryComponent) string {
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
	return materialsText
}

// drawRecipeListFooter draws the scroll indicator and footer hints.
func (ui *CraftingUI) drawRecipeListFooter(img *ebiten.Image, recipeList []*Recipe, maxVisibleRecipes, windowX, windowY, windowWidth, windowHeight int) {
	// Scroll indicator
	if len(recipeList) > maxVisibleRecipes {
		scrollText := fmt.Sprintf("Scroll: %d-%d / %d recipes",
			ui.scrollOffset+1,
			minInt(ui.scrollOffset+maxVisibleRecipes, len(recipeList)),
			len(recipeList))
		ebitenutil.DebugPrintAt(img, scrollText, windowX+windowWidth-200, windowY+windowHeight-30)
	}

	// Footer hints
	footerY := windowY + windowHeight - 30
	ebitenutil.DebugPrintAt(img, "Arrow Keys: Navigate | ENTER/SPACE: Craft | Mouse Wheel: Scroll",
		windowX+10, footerY)

	// Nearby station hint
	ui.drawNearbyStationHint(img, windowX, footerY)
}

// drawNearbyStationHint draws a hint about nearby crafting stations.
func (ui *CraftingUI) drawNearbyStationHint(img *ebiten.Image, windowX, footerY int) {
	if ui.stationEntity != nil || ui.playerEntity == nil {
		return
	}

	posComp, ok := ui.playerEntity.GetComponent("position")
	if !ok {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Find nearest station within 100 pixels
	nearestStation, distance := ui.findNearestStation(pos.X, pos.Y, 100)
	if nearestStation == nil {
		return
	}

	stationComp, ok := nearestStation.GetComponent("crafting_station")
	if !ok {
		return
	}
	station, ok := stationComp.(*CraftingStationComponent)
	if !ok {
		return
	}

	stationHint := fmt.Sprintf("Nearby: %s (%.0f units away) - Move closer to use station bonuses",
		station.StationType.String(), distance)
	ebitenutil.DebugPrintAt(img, stationHint, windowX+10, footerY-20)
}

func (ui *CraftingUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}

	if !ui.visible {
		return
	}

	components, ok := ui.getPlayerCraftingComponents()
	if !ok {
		return
	}

	recipes := components.knowledge.KnownRecipes

	// Draw overlay and window background
	windowX, windowY := ui.drawWindowBackground(img)

	// Draw title bar
	ui.drawTitleBar(img, windowX, windowY)

	// Draw player stats
	ui.drawPlayerStats(img, components, windowX, windowY)

	// Draw crafting message if active
	ui.drawCraftingMessage(img, windowX, windowY)

	// Draw instructions
	ui.drawInstructions(img, windowX, windowY)

	// Draw search/filter UI
	ui.drawSearchFilterUI(img, windowX, windowY)

	// Draw recipe list
	windowWidth := 800
	windowHeight := 600
	ui.drawRecipeList(img, recipes, components, windowX, windowY, windowWidth, windowHeight)

	// H-002 FIX: Draw error feedback
	ui.errorState.DrawError(img)
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

// Issue #12 FIX: Helper functions for search/filter functionality

// filterAndSortRecipes applies search, filter, and sort to recipe list.
// Returns filtered and sorted recipes that match current UI state.
func (ui *CraftingUI) filterAndSortRecipes(recipes []*Recipe) []*Recipe {
	// Step 1: Apply search filter
	var filtered []*Recipe
	for _, recipe := range recipes {
		if ui.matchesSearch(recipe) && ui.matchesCategory(recipe) {
			filtered = append(filtered, recipe)
		}
	}

	// Step 2: Apply craftable filter if enabled
	if ui.showOnlyCrafted {
		var craftable []*Recipe
		for _, recipe := range filtered {
			if ui.canCraftRecipe(recipe) {
				craftable = append(craftable, recipe)
			}
		}
		filtered = craftable
	}

	// Step 3: Sort based on current sort mode
	ui.sortRecipes(filtered)

	return filtered
}

// matchesSearch returns true if recipe matches current search query.
func (ui *CraftingUI) matchesSearch(recipe *Recipe) bool {
	if ui.searchQuery == "" {
		return true
	}

	// Case-insensitive substring match on recipe name
	query := ui.searchQuery
	name := recipe.Name

	// Simple case-insensitive contains check
	queryLen := len(query)
	nameLen := len(name)
	if queryLen > nameLen {
		return false
	}

	for i := 0; i <= nameLen-queryLen; i++ {
		match := true
		for j := 0; j < queryLen; j++ {
			c1 := name[i+j]
			c2 := query[j]
			// Convert to lowercase for comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 = c1 + 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 = c2 + 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// matchesCategory returns true if recipe matches current category filter.
func (ui *CraftingUI) matchesCategory(recipe *Recipe) bool {
	if ui.filterCategory == -1 {
		return true // "All" category
	}
	return recipe.Type == ui.filterCategory
}

// canCraftRecipe checks if player has materials to craft recipe.
func (ui *CraftingUI) canCraftRecipe(recipe *Recipe) bool {
	if ui.playerEntity == nil {
		return false
	}

	invComp, hasInv := ui.playerEntity.GetComponent("inventory")
	if !hasInv {
		return false
	}
	inventory, ok := invComp.(*InventoryComponent)
	if !ok {
		return false
	}

	// Check if player has all required materials
	for _, mat := range recipe.Materials {
		hasQuantity := 0
		for _, item := range inventory.Items {
			if item != nil && item.Name == mat.ItemName {
				hasQuantity++
			}
		}
		if hasQuantity < mat.Quantity {
			return false // Missing required material
		}
	}

	// Check gold requirement
	if recipe.GoldCost > 0 {
		if inventory.Gold < recipe.GoldCost {
			return false
		}
	}

	return true
}

// sortRecipes sorts recipes in place based on current sort mode.
func (ui *CraftingUI) sortRecipes(recipes []*Recipe) {
	if len(recipes) <= 1 {
		return
	}

	// Bubble sort (simple for small lists)
	n := len(recipes)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			shouldSwap := false

			switch ui.sortMode {
			case 0: // Sort by name (alphabetical)
				shouldSwap = recipes[j].Name > recipes[j+1].Name
			case 1: // Sort by tier (rarity)
				shouldSwap = recipes[j].Rarity < recipes[j+1].Rarity
			case 2: // Sort by craftable (craftable first)
				canJ := ui.canCraftRecipe(recipes[j])
				canJPlus1 := ui.canCraftRecipe(recipes[j+1])
				shouldSwap = !canJ && canJPlus1
			}

			if shouldSwap {
				recipes[j], recipes[j+1] = recipes[j+1], recipes[j]
			}
		}
	}
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
