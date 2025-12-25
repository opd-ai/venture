//go:build !headless
// +build !headless

// Package engine provides quest_ui for game UI.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"golang.org/x/image/font/basicfont"
)

// QuestUI handles rendering and interaction for the quest log.
type EbitenQuestUI struct {
	visible      bool
	world        *World
	playerEntity *Entity

	// Layout
	screenWidth  int
	screenHeight int

	// Tab selection
	currentTab int // 0 = Active, 1 = Completed

	// Scrolling support for long quest lists
	scrollOffset int // Vertical scroll offset in pixels
	maxScroll    int // Maximum scroll offset based on content height

	// H-002 FIX: Error feedback
	errorState *UIErrorState

	// M-003 FIX: Quest list height caching
	cachedQuestListHeight int  // Cached total content height
	cacheValid            bool // Whether cache is valid
	lastQuestCount        int  // Quest count when cache was built

	// Touch support
	touchHandler    *mobile.TouchInputHandler
	closeButton     *mobile.TouchButton
	activeTabButton *mobile.TouchButton
	doneTabButton   *mobile.TouchButton
}

// NewQuestUI creates a new quest UI.
func NewEbitenQuestUI(world *World, screenWidth, screenHeight int) *EbitenQuestUI {
	ui := &EbitenQuestUI{
		visible:      false,
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		currentTab:   0,
		errorState:   NewUIErrorState(), // H-002 FIX
		cacheValid:   false,             // M-003 FIX
		touchHandler: mobile.NewTouchInputHandler(),
	}

	// Create close button
	ui.closeButton = mobile.NewTouchButton(
		float64(screenWidth-64),
		10,
		44, 44,
		"✕",
		func() { ui.Hide() },
	)

	// Create tab buttons
	windowWidth := 800
	if screenWidth < 800 {
		windowWidth = screenWidth - 40
	}
	windowX := (screenWidth - windowWidth) / 2

	ui.activeTabButton = mobile.NewTouchButton(
		float64(windowX+20),
		60,
		120, 44,
		"Active",
		func() { ui.currentTab = 0; ui.scrollOffset = 0 },
	)

	ui.doneTabButton = mobile.NewTouchButton(
		float64(windowX+150),
		60,
		120, 44,
		"Completed",
		func() { ui.currentTab = 1; ui.scrollOffset = 0 },
	)

	return ui
}

// SetPlayerEntity sets the player entity whose quests to display.
func (ui *EbitenQuestUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the quest UI.
func (ui *EbitenQuestUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the quest log is currently shown.
func (ui *EbitenQuestUI) IsVisible() bool {
	return ui.visible
}

// Show displays the quest UI.
func (ui *EbitenQuestUI) Show() {
	ui.visible = true
}

// Hide hides the quest UI.
func (ui *EbitenQuestUI) Hide() {
	ui.visible = false
}

// Update processes input for the quest UI.
// Calculates scroll bounds based on current quest list content.
func (ui *EbitenQuestUI) Update(entities []*Entity, deltaTime float64) {
	ui.updateTouchInputs()

	if ui.handleMenuNavigation() {
		return
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	ui.calculateMaxScroll()
	ui.handleTouchScrolling()
	ui.handleTabSwitching()
	ui.handleKeyboardScrolling()
	ui.handleMouseWheelScrolling()
}

// updateTouchInputs updates all touch-related UI elements.
func (ui *EbitenQuestUI) updateTouchInputs() {
	if ui.touchHandler != nil {
		ui.touchHandler.Update()
	}

	if ui.closeButton != nil {
		ui.closeButton.Update()
	}
	if ui.activeTabButton != nil {
		ui.activeTabButton.Update()
	}
	if ui.doneTabButton != nil {
		ui.doneTabButton.Update()
	}
}

// handleMenuNavigation processes menu open/close input.
// Returns true if menu was closed and further input processing should stop.
func (ui *EbitenQuestUI) handleMenuNavigation() bool {
	// BUG FIX: Phase 1.2 - Mobile gesture support for closing quest log
	// Resolution: Use HandleMenuInputWithTouch to support swipe gestures on mobile
	// Platform: Mobile (Android/iOS)
	if shouldClose, shouldToggle := HandleMenuInputWithTouch(MenuKeys.Quests, ui.visible, ui.touchHandler); shouldClose {
		if shouldToggle {
			ui.Toggle()
		} else {
			ui.Hide()
		}
		ui.scrollOffset = 0
		return true
	}
	return false
}

// calculateMaxScroll computes the maximum scroll offset based on current quest list.
func (ui *EbitenQuestUI) calculateMaxScroll() {
	if ui.playerEntity == nil {
		return
	}

	trackerComp, ok := ui.playerEntity.GetComponent("questtracker")
	if !ok {
		return
	}

	tracker, ok := trackerComp.(*QuestTrackerComponent)
	if !ok {
		return
	}

	quests := ui.getQuestsForCurrentTab(tracker)
	estimatedHeight := ui.estimateContentHeight(quests)
	contentHeight := 500 - 90 - 40
	ui.maxScroll = estimatedHeight - contentHeight
	if ui.maxScroll < 0 {
		ui.maxScroll = 0
	}
}

// getQuestsForCurrentTab returns the quest list for the active tab.
func (ui *EbitenQuestUI) getQuestsForCurrentTab(tracker *QuestTrackerComponent) []*TrackedQuest {
	if ui.currentTab == 0 {
		return tracker.ActiveQuests
	}
	return tracker.CompletedQuests
}

// estimateContentHeight calculates total pixel height for quest list.
func (ui *EbitenQuestUI) estimateContentHeight(quests []*TrackedQuest) int {
	estimatedHeight := 0
	for _, tracked := range quests {
		baseHeight := 120
		baseHeight += len(tracked.Quest.Objectives) * 20
		estimatedHeight += baseHeight
	}
	return estimatedHeight
}

// handleTouchScrolling processes touch swipe gestures for scrolling.
func (ui *EbitenQuestUI) handleTouchScrolling() {
	if ui.touchHandler == nil {
		return
	}

	direction, distance, detected := ui.touchHandler.GetSwipe()
	if !detected {
		return
	}

	if direction > 1.0 || direction < -1.0 {
		scrollSpeed := int(distance * 0.5)
		if direction < 0 {
			ui.scrollOffset -= scrollSpeed
		} else {
			ui.scrollOffset += scrollSpeed
		}
		ui.clampScrollOffset()
	}
}

// handleTabSwitching processes tab change input.
func (ui *EbitenQuestUI) handleTabSwitching() {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		ui.switchToTab(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		ui.switchToTab(1)
	}
}

// switchToTab changes the current tab and resets scroll state.
func (ui *EbitenQuestUI) switchToTab(tabIndex int) {
	ui.currentTab = tabIndex
	ui.scrollOffset = 0
	ui.cacheValid = false
}

// handleKeyboardScrolling processes arrow key scrolling input.
func (ui *EbitenQuestUI) handleKeyboardScrolling() {
	scrollSpeed := 20
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		ui.scrollOffset -= scrollSpeed
		ui.clampScrollOffset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		ui.scrollOffset += scrollSpeed
		ui.clampScrollOffset()
	}
}

// handleMouseWheelScrolling processes mouse wheel scrolling input.
func (ui *EbitenQuestUI) handleMouseWheelScrolling() {
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		scrollSpeed := 20
		ui.scrollOffset -= int(wheelY * float64(scrollSpeed))
		ui.clampScrollOffset()
	}
}

// clampScrollOffset ensures scroll offset stays within valid bounds.
func (ui *EbitenQuestUI) clampScrollOffset() {
	if ui.scrollOffset < 0 {
		ui.scrollOffset = 0
	}
	if ui.scrollOffset > ui.maxScroll {
		ui.scrollOffset = ui.maxScroll
	}
}

// Draw renders the quest UI.
func (ui *EbitenQuestUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	if !ui.visible || ui.playerEntity == nil {
		return
	}

	tracker, ok := ui.getQuestTracker()
	if !ok {
		return
	}

	windowWidth := 600
	windowHeight := 500
	windowX, windowY := ui.drawWindowBackground(img, windowWidth, windowHeight)
	ui.drawHeader(img, windowX, windowY, windowWidth)
	tabY := ui.drawTabs(img, windowX, windowY)

	quests := ui.getQuestsForCurrentTab(tracker)
	listY := tabY + 40
	contentStartY := listY + 10
	contentMaxY := windowY + windowHeight - 20
	contentHeight := contentMaxY - contentStartY

	if len(quests) == 0 {
		ebitenutil.DebugPrintAt(img, "No quests", windowX+20, contentStartY)
	} else {
		totalContentHeight := ui.getCachedQuestListHeight(quests, windowWidth)
		ui.drawQuestList(img, quests, windowX, windowWidth, contentStartY, contentMaxY)
		ui.updateMaxScroll(totalContentHeight, contentHeight)
		ui.drawScrollbar(img, windowX, windowWidth, contentStartY, contentHeight, totalContentHeight)
	}

	ui.drawControlsHint(img, windowX, windowY, windowHeight)

	// Draw touch buttons
	if ui.closeButton != nil {
		// Position close button at top-right of window
		ui.closeButton.SetPosition(float64(windowX+windowWidth-54), float64(windowY+10))
		ui.closeButton.Draw(img)
	}

	// Draw tab buttons at correct position
	if ui.activeTabButton != nil {
		// Highlight active tab
		if ui.currentTab == 0 {
			ui.activeTabButton.BackgroundColor = color.RGBA{100, 150, 255, 255}
		} else {
			ui.activeTabButton.BackgroundColor = color.RGBA{50, 50, 70, 255}
		}
		ui.activeTabButton.SetPosition(float64(windowX+20), 60)
		ui.activeTabButton.Draw(img)
	}

	if ui.doneTabButton != nil {
		// Highlight active tab
		if ui.currentTab == 1 {
			ui.doneTabButton.BackgroundColor = color.RGBA{100, 150, 255, 255}
		} else {
			ui.doneTabButton.BackgroundColor = color.RGBA{50, 50, 70, 255}
		}
		ui.doneTabButton.SetPosition(float64(windowX+150), 60)
		ui.doneTabButton.Draw(img)
	}

	ui.errorState.DrawError(img)
}

// getQuestTracker retrieves the quest tracker component from the player entity.
func (ui *EbitenQuestUI) getQuestTracker() (*QuestTrackerComponent, bool) {
	trackerComp, ok := ui.playerEntity.GetComponent("questtracker")
	if !ok {
		return nil, false
	}
	tracker, ok := trackerComp.(*QuestTrackerComponent)
	return tracker, ok
}

// drawWindowBackground renders the semi-transparent overlay and window background.
// Returns the window x and y coordinates.
func (ui *EbitenQuestUI) drawWindowBackground(img *ebiten.Image, windowWidth, windowHeight int) (int, int) {
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	img.DrawImage(overlay, nil)

	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{40, 40, 50, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	return windowX, windowY
}

// drawHeader renders the title and exit hint at the top of the window.
func (ui *EbitenQuestUI) drawHeader(img *ebiten.Image, windowX, windowY, windowWidth int) {
	ebitenutil.DebugPrintAt(img, "QUEST LOG", windowX+10, windowY+10)
	exitHint := GetExitHint(MenuKeys.Quests)
	ebitenutil.DebugPrintAt(img, exitHint, windowX+windowWidth-200, windowY+10)
}

// drawTabs renders the Active and Completed tabs.
// Returns the y coordinate of the bottom of the tabs.
func (ui *EbitenQuestUI) drawTabs(img *ebiten.Image, windowX, windowY int) int {
	tabY := windowY + 40
	tabs := []string{"Active", "Completed"}
	for i, tabName := range tabs {
		tabX := windowX + 10 + i*100
		tabColor := color.RGBA{60, 60, 70, 255}
		if i == ui.currentTab {
			tabColor = color.RGBA{80, 80, 100, 255}
		}

		tabBg := ebiten.NewImage(90, 30)
		tabBg.Fill(tabColor)
		tabOpts := &ebiten.DrawImageOptions{}
		tabOpts.GeoM.Translate(float64(tabX), float64(tabY))
		img.DrawImage(tabBg, tabOpts)

		ebitenutil.DebugPrintAt(img, tabName, tabX+10, tabY+10)
	}
	return tabY
}

// getCachedQuestListHeight returns the total height of the quest list, using cache when valid.
func (ui *EbitenQuestUI) getCachedQuestListHeight(quests []*TrackedQuest, windowWidth int) int {
	if ui.cacheValid && ui.lastQuestCount == len(quests) {
		return ui.cachedQuestListHeight
	}
	totalContentHeight := ui.calculateQuestListHeight(quests, windowWidth)
	ui.cachedQuestListHeight = totalContentHeight
	ui.lastQuestCount = len(quests)
	ui.cacheValid = true
	return totalContentHeight
}

// drawQuestList renders all visible quests in the scrollable area.
func (ui *EbitenQuestUI) drawQuestList(img *ebiten.Image, quests []*TrackedQuest, windowX, windowWidth, contentStartY, contentMaxY int) {
	y := contentStartY - ui.scrollOffset
	maxTextWidth := windowWidth - 80

	for _, tracked := range quests {
		if y < contentStartY-200 {
			y += 120
			continue
		}
		if y > contentMaxY {
			break
		}
		y = ui.drawSingleQuest(img, tracked, windowX, maxTextWidth, y, contentStartY, contentMaxY)
	}
}

// drawSingleQuest renders a single quest entry with its details.
// Returns the updated y coordinate after drawing.
func (ui *EbitenQuestUI) drawSingleQuest(img *ebiten.Image, tracked *TrackedQuest, windowX, maxTextWidth, y, contentStartY, contentMaxY int) int {
	y = ui.drawQuestName(img, tracked, windowX, maxTextWidth, y, contentStartY, contentMaxY)
	y = ui.drawQuestInfo(img, tracked, windowX, y, contentStartY, contentMaxY)
	y = ui.drawQuestObjectives(img, tracked, windowX, maxTextWidth, y, contentStartY, contentMaxY)
	y = ui.drawQuestRewards(img, tracked, windowX, y, contentStartY, contentMaxY)
	return y + 30
}

// drawQuestName renders the quest name with text wrapping.
func (ui *EbitenQuestUI) drawQuestName(img *ebiten.Image, tracked *TrackedQuest, windowX, maxTextWidth, y, contentStartY, contentMaxY int) int {
	nameLines := WrapText(tracked.Quest.Name, maxTextWidth, basicfont.Face7x13)
	for _, line := range nameLines {
		if y >= contentStartY && y <= contentMaxY {
			ebitenutil.DebugPrintAt(img, line, windowX+20, y)
		}
		y += 15
	}
	return y
}

// drawQuestInfo renders the quest type and difficulty.
func (ui *EbitenQuestUI) drawQuestInfo(img *ebiten.Image, tracked *TrackedQuest, windowX, y, contentStartY, contentMaxY int) int {
	info := fmt.Sprintf("%s | %s", tracked.Quest.Type.String(), tracked.Quest.Difficulty.String())
	if y >= contentStartY && y <= contentMaxY {
		ebitenutil.DebugPrintAt(img, info, windowX+30, y)
	}
	return y + 20
}

// drawQuestObjectives renders all objectives for a quest with progress bars.
func (ui *EbitenQuestUI) drawQuestObjectives(img *ebiten.Image, tracked *TrackedQuest, windowX, maxTextWidth, y, contentStartY, contentMaxY int) int {
	for _, obj := range tracked.Quest.Objectives {
		y = ui.drawObjectiveText(img, obj, windowX, maxTextWidth, y, contentStartY, contentMaxY)
		ui.drawProgressBar(img, obj, windowX, y, contentStartY, contentMaxY)
		y += 5
	}
	return y + 15
}

// drawObjectiveText renders the objective description with progress.
func (ui *EbitenQuestUI) drawObjectiveText(img *ebiten.Image, obj quest.Objective, windowX, maxTextWidth, y, contentStartY, contentMaxY int) int {
	progressPrefix := fmt.Sprintf("  [%d/%d] ", obj.Current, obj.Required)
	descLines := WrapText(obj.Description, maxTextWidth-100, basicfont.Face7x13)

	for i, line := range descLines {
		prefix := progressPrefix
		if i > 0 {
			prefix = "       "
		}
		if y >= contentStartY && y <= contentMaxY {
			ebitenutil.DebugPrintAt(img, prefix+line, windowX+30, y)
		}
		y += 15
	}
	return y
}

// drawProgressBar renders a progress bar for a quest objective.
func (ui *EbitenQuestUI) drawProgressBar(img *ebiten.Image, obj quest.Objective, windowX, y, contentStartY, contentMaxY int) {
	barWidth := 200
	barHeight := 8
	barX := windowX + 240
	barY := y - 10

	if barY < contentStartY || barY > contentMaxY {
		return
	}

	barBg := ebiten.NewImage(barWidth, barHeight)
	barBg.Fill(color.RGBA{60, 60, 70, 255})
	barOpts := &ebiten.DrawImageOptions{}
	barOpts.GeoM.Translate(float64(barX), float64(barY))
	img.DrawImage(barBg, barOpts)

	progressPct := obj.Progress()
	if progressPct > 0 {
		progressWidth := int(float64(barWidth) * progressPct)
		barFill := ebiten.NewImage(progressWidth, barHeight)
		fillColor := color.RGBA{80, 180, 80, 255}
		if obj.IsComplete() {
			fillColor = color.RGBA{100, 220, 100, 255}
		}
		barFill.Fill(fillColor)
		img.DrawImage(barFill, barOpts)
	}
}

// drawQuestRewards renders the reward information for a quest.
func (ui *EbitenQuestUI) drawQuestRewards(img *ebiten.Image, tracked *TrackedQuest, windowX, y, contentStartY, contentMaxY int) int {
	rewards := fmt.Sprintf("  Rewards: %d XP, %d Gold", tracked.Quest.Reward.XP, tracked.Quest.Reward.Gold)
	if y >= contentStartY && y <= contentMaxY {
		ebitenutil.DebugPrintAt(img, rewards, windowX+30, y)
	}
	return y
}

// updateMaxScroll updates the maximum scroll offset based on content height.
func (ui *EbitenQuestUI) updateMaxScroll(totalContentHeight, contentHeight int) {
	ui.maxScroll = totalContentHeight - contentHeight
	if ui.maxScroll < 0 {
		ui.maxScroll = 0
	}
}

// drawScrollbar renders the scrollbar if content exceeds visible area.
func (ui *EbitenQuestUI) drawScrollbar(img *ebiten.Image, windowX, windowWidth, contentStartY, contentHeight, totalContentHeight int) {
	if ui.maxScroll <= 0 {
		return
	}

	scrollbarX := windowX + windowWidth - 15
	scrollbarY := contentStartY
	scrollbarHeight := contentHeight

	scrollbarBg := ebiten.NewImage(10, scrollbarHeight)
	scrollbarBg.Fill(color.RGBA{60, 60, 70, 255})
	scrollbarOpts := &ebiten.DrawImageOptions{}
	scrollbarOpts.GeoM.Translate(float64(scrollbarX), float64(scrollbarY))
	img.DrawImage(scrollbarBg, scrollbarOpts)

	handleHeight := max(20, (contentHeight*contentHeight)/totalContentHeight)
	handleY := scrollbarY + (ui.scrollOffset*scrollbarHeight)/totalContentHeight
	scrollbarHandle := ebiten.NewImage(10, handleHeight)
	scrollbarHandle.Fill(color.RGBA{150, 150, 170, 255})
	handleOpts := &ebiten.DrawImageOptions{}
	handleOpts.GeoM.Translate(float64(scrollbarX), float64(handleY))
	img.DrawImage(scrollbarHandle, handleOpts)

	scrollHint := "↑↓/Wheel: Scroll"
	ebitenutil.DebugPrintAt(img, scrollHint, windowX+windowWidth-140, contentStartY-20)
}

// drawControlsHint renders the controls hint at the bottom of the window.
func (ui *EbitenQuestUI) drawControlsHint(img *ebiten.Image, windowX, windowY, windowHeight int) {
	controlsY := windowY + windowHeight - 20
	ebitenutil.DebugPrintAt(img, "J: Close | 1: Active | 2: Completed", windowX+10, controlsY)
}

// IsActive returns whether the quest UI is currently visible.
// Implements UISystem interface.
func (q *EbitenQuestUI) IsActive() bool {
	return q.visible
}

// SetActive sets whether the quest UI is visible.
// Implements UISystem interface.
func (q *EbitenQuestUI) SetActive(active bool) {
	q.visible = active
}

// M-003 FIX: calculateQuestListHeight computes total height of all quests in list.
// This is cached to avoid O(n) recalculation every frame.
// Parameters:
//
//	quests - List of quests to measure
//	windowWidth - Width of window for text wrapping calculation
//
// Returns: Total height in pixels
func (ui *EbitenQuestUI) calculateQuestListHeight(quests []*TrackedQuest, windowWidth int) int {
	totalHeight := 0
	maxTextWidth := windowWidth - 80 // Leave margin for scrollbar

	for _, tracked := range quests {
		// Quest name height (with text wrapping)
		nameLines := WrapText(tracked.Quest.Name, maxTextWidth, basicfont.Face7x13)
		questHeight := len(nameLines) * 15 // 15px per line

		// Description height (with text wrapping)
		descLines := WrapText(tracked.Quest.Description, maxTextWidth, basicfont.Face7x13)
		questHeight += len(descLines) * 15

		// Objectives height
		questHeight += len(tracked.Quest.Objectives) * 20 // ~20px per objective

		// Rewards line
		questHeight += 30

		// Spacing between quests
		questHeight += 30

		totalHeight += questHeight
	}

	return totalHeight
}

// Compile-time check that EbitenQuestUI implements UISystem
var _ UISystem = (*EbitenQuestUI)(nil)
