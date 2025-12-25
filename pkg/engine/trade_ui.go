//go:build !headless
// +build !headless

package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TradeUIState represents the current state of the trade UI.
type TradeUIState int

const (
	// TradeUIStateIdle means no trade is active
	TradeUIStateIdle TradeUIState = iota
	// TradeUIStateSelectingPartner means player is selecting who to trade with
	TradeUIStateSelectingPartner
	// TradeUIStateSelectingItems means player is selecting items to offer/request
	TradeUIStateSelectingItems
	// TradeUIStatePending means trade proposal sent, waiting for response
	TradeUIStatePending
	// TradeUIStateNegotiating means both parties are reviewing the trade
	TradeUIStateNegotiating
	// TradeUIStateCompleted means trade was successfully completed
	TradeUIStateCompleted
	// TradeUIStateCancelled means trade was cancelled or rejected
	TradeUIStateCancelled
)

// String returns the string representation of trade UI state.
func (s TradeUIState) String() string {
	switch s {
	case TradeUIStateIdle:
		return "Idle"
	case TradeUIStateSelectingPartner:
		return "Selecting Partner"
	case TradeUIStateSelectingItems:
		return "Selecting Items"
	case TradeUIStatePending:
		return "Pending"
	case TradeUIStateNegotiating:
		return "Negotiating"
	case TradeUIStateCompleted:
		return "Completed"
	case TradeUIStateCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// TradeUI handles rendering and interaction for player-to-player trading.
// Implements two-phase commit protocol with item selection UI.
type TradeUI struct {
	visible bool
	state   TradeUIState

	// Entity references
	playerEntity  *Entity
	partnerEntity *Entity

	// System references
	world       *World
	tradeSystem *TradeSystem

	// Layout
	screenWidth  int
	screenHeight int
	gridCols     int
	gridRows     int
	slotSize     int
	padding      int

	// Item selection
	selectedOfferedSlots   []int // Indices of items player is offering
	selectedRequestedSlots []int // Indices of items player is requesting
	hoveredSlot            int   // Currently hovered slot

	// Grid navigation (keyboard)
	focusedPanel int // 0 = offer panel, 1 = request panel
	cursorRow    int // Current cursor row in focused panel
	cursorCol    int // Current cursor column in focused panel
	gridStartX   int // X position of grid (calculated from center)

	// UI panels
	offerPanelY    int
	requestPanelY  int
	partnerListY   int
	statusMessageY int
	statusMessage  string
	statusColor    color.Color
	messageTimeout float64

	// Partner selection
	nearbyPartners       []*Entity
	selectedPartnerIndex int

	// Touch support
	touchHandler       *mobile.TouchInputHandler
	closeButton        *mobile.TouchButton
	proposeButton      *mobile.TouchButton
	acceptButton       *mobile.TouchButton
	rejectButton       *mobile.TouchButton
	cancelButton       *mobile.TouchButton
	confirmOfferButton *mobile.TouchButton
	partnerButtons     []*mobile.TouchButton
}

// NewTradeUI creates a new trade UI.
func NewTradeUI(world *World, tradeSystem *TradeSystem, screenWidth, screenHeight int) *TradeUI {
	ui := &TradeUI{
		visible:                false,
		state:                  TradeUIStateIdle,
		world:                  world,
		tradeSystem:            tradeSystem,
		screenWidth:            screenWidth,
		screenHeight:           screenHeight,
		gridCols:               6,
		gridRows:               2,
		slotSize:               64,
		padding:                15,
		hoveredSlot:            -1,
		selectedOfferedSlots:   make([]int, 0),
		selectedRequestedSlots: make([]int, 0),
		selectedPartnerIndex:   -1,
		statusColor:            color.White,
		touchHandler:           mobile.NewTouchInputHandler(),
		nearbyPartners:         make([]*Entity, 0),
		partnerButtons:         make([]*mobile.TouchButton, 0),
		focusedPanel:           0,
		cursorRow:              0,
		cursorCol:              0,
	}

	// Calculate panel positions
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowX := (screenWidth - windowWidth) / 2
	ui.gridStartX = (screenWidth - ui.gridCols*ui.slotSize) / 2

	ui.offerPanelY = 140
	ui.requestPanelY = ui.offerPanelY + ui.gridRows*ui.slotSize + 80
	ui.partnerListY = 180
	ui.statusMessageY = screenHeight - 100

	// Create close button (top-right)
	ui.closeButton = mobile.NewTouchButton(
		float64(windowX+windowWidth-54),
		10,
		44, 44,
		"✕",
		func() { ui.Close() },
	)

	// Create propose button (bottom center)
	ui.proposeButton = mobile.NewTouchButton(
		float64(screenWidth/2-120),
		float64(screenHeight-120),
		110, 44,
		"Propose",
		func() { ui.proposeTrade() },
	)

	// Create confirm offer button (for selecting items)
	ui.confirmOfferButton = mobile.NewTouchButton(
		float64(screenWidth/2-120),
		float64(screenHeight-120),
		110, 44,
		"Confirm",
		func() { ui.confirmItemSelection() },
	)

	// Create accept button
	ui.acceptButton = mobile.NewTouchButton(
		float64(screenWidth/2-120),
		float64(screenHeight-120),
		110, 44,
		"Accept",
		func() { ui.acceptTrade() },
	)

	// Create reject button
	ui.rejectButton = mobile.NewTouchButton(
		float64(screenWidth/2+10),
		float64(screenHeight-120),
		110, 44,
		"Reject",
		func() { ui.rejectTrade() },
	)

	// Create cancel button
	ui.cancelButton = mobile.NewTouchButton(
		float64(screenWidth/2+10),
		float64(screenHeight-120),
		110, 44,
		"Cancel",
		func() { ui.cancelTrade() },
	)

	return ui
}

// SetPlayerEntity sets the player entity.
func (ui *TradeUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Open opens the trade UI and finds nearby trade partners.
func (ui *TradeUI) Open() {
	ui.visible = true
	ui.state = TradeUIStateSelectingPartner
	ui.findNearbyPartners()
	ui.resetSelection()
	ui.setStatusMessage("Select a player to trade with", color.White)
}

// Close closes the trade UI and cancels any active trade.
func (ui *TradeUI) Close() {
	if ui.state == TradeUIStatePending || ui.state == TradeUIStateNegotiating {
		ui.cancelTrade()
	}
	ui.visible = false
	ui.state = TradeUIStateIdle
	ui.resetSelection()
	ui.partnerEntity = nil
}

// Toggle toggles the visibility of the trade UI.
func (ui *TradeUI) Toggle() {
	if ui.visible {
		ui.Close()
	} else {
		ui.Open()
	}
}

// IsVisible returns whether the trade UI is visible.
func (ui *TradeUI) IsVisible() bool {
	return ui.visible
}

// Update updates the trade UI state and handles input.
func (ui *TradeUI) Update(deltaTime float64) {
	if !ui.visible {
		return
	}

	// Update message timeout
	if ui.messageTimeout > 0 {
		ui.messageTimeout -= deltaTime
		if ui.messageTimeout <= 0 {
			ui.statusMessage = ""
		}
	}

	// Check for active trade updates
	ui.updateTradeState()

	// Handle keyboard input based on state
	switch ui.state {
	case TradeUIStateSelectingPartner:
		ui.handlePartnerSelectionInput()
	case TradeUIStateSelectingItems:
		ui.handleItemSelectionInput()
	case TradeUIStateNegotiating:
		ui.handleNegotiationInput()
	}

	// Handle touch input
	ui.handleTouchInput()
}

// Draw renders the trade UI.
func (ui *TradeUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Draw semi-transparent background
	overlayImg := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlayImg.Fill(color.RGBA{0, 0, 0, 200})
	screen.DrawImage(overlayImg, nil)

	// Draw UI based on state
	switch ui.state {
	case TradeUIStateSelectingPartner:
		ui.drawPartnerSelection(screen)
	case TradeUIStateSelectingItems:
		ui.drawItemSelection(screen)
	case TradeUIStatePending:
		ui.drawPendingState(screen)
	case TradeUIStateNegotiating:
		ui.drawNegotiationState(screen)
	case TradeUIStateCompleted:
		ui.drawCompletedState(screen)
	case TradeUIStateCancelled:
		ui.drawCancelledState(screen)
	}

	// Draw status message
	ui.drawStatusMessage(screen)

	// Draw close button
	ui.closeButton.Draw(screen)
}

// drawPartnerSelection draws the partner selection screen.
func (ui *TradeUI) drawPartnerSelection(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2

	// Title
	ebitenutil.DebugPrintAt(screen, "Trade - Select Partner", centerX-100, 40)

	// Instructions
	ebitenutil.DebugPrintAt(screen, "Use UP/DOWN to select, ENTER to confirm", centerX-150, 80)

	// Partner list
	y := ui.partnerListY
	for i, partner := range ui.nearbyPartners {
		x := centerX - 200

		// Highlight selected partner
		bgColor := color.RGBA{50, 50, 50, 255}
		if i == ui.selectedPartnerIndex {
			bgColor = color.RGBA{100, 100, 150, 255}
		}

		// Draw partner slot background
		slotImg := ebiten.NewImage(400, 50)
		slotImg.Fill(bgColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(slotImg, op)

		// Get partner info
		partnerName := ui.getEntityName(partner)
		distance := ui.getDistance(ui.playerEntity, partner)

		// Draw partner info
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s (%.1f tiles away)", partnerName, distance), x+10, y+18)

		y += 60
	}

	if len(ui.nearbyPartners) == 0 {
		ebitenutil.DebugPrintAt(screen, "No nearby players (must be within 5 tiles)", centerX-150, y)
	}
}

// drawItemSelection draws the item selection screen.
func (ui *TradeUI) drawItemSelection(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2

	// Title
	partnerName := ui.getEntityName(ui.partnerEntity)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Trading with %s", partnerName), centerX-100, 40)

	// Instructions
	ebitenutil.DebugPrintAt(screen, "Arrow keys: navigate | Tab: switch panel | Space: select | Enter: confirm", centerX-280, 80)

	// Draw offer panel (your items)
	ui.drawOfferPanel(screen)

	// Draw request panel (their items)
	ui.drawRequestPanel(screen)

	// Draw confirm button
	ui.confirmOfferButton.Draw(screen)
	ui.cancelButton.Draw(screen)
}

// drawOfferPanel draws the player's items to offer.
func (ui *TradeUI) drawOfferPanel(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2

	// Show focused indicator
	label := "Your Items (OFFER):"
	if ui.focusedPanel == 0 {
		label = "► Your Items (OFFER):"
	}
	ebitenutil.DebugPrintAt(screen, label, centerX-280, ui.offerPanelY-30)

	inv := ui.getInventoryComponent(ui.playerEntity)
	if inv == nil {
		return
	}

	ui.drawItemGrid(screen, inv.Items, ui.offerPanelY, ui.selectedOfferedSlots, "offer")
}

// drawRequestPanel draws the partner's items to request.
func (ui *TradeUI) drawRequestPanel(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2

	// Show focused indicator
	label := "Their Items (REQUEST):"
	if ui.focusedPanel == 1 {
		label = "► Their Items (REQUEST):"
	}
	ebitenutil.DebugPrintAt(screen, label, centerX-280, ui.requestPanelY-30)

	partnerInv := ui.getInventoryComponent(ui.partnerEntity)
	if partnerInv == nil {
		return
	}

	ui.drawItemGrid(screen, partnerInv.Items, ui.requestPanelY, ui.selectedRequestedSlots, "request")
}

// drawItemGrid draws a grid of items with selection highlighting.
func (ui *TradeUI) drawItemGrid(screen *ebiten.Image, items []*item.Item, startY int, selectedSlots []int, panelType string) {
	centerX := ui.screenWidth / 2
	gridStartX := centerX - (ui.gridCols*ui.slotSize)/2

	// Determine if this is the focused panel
	isFocusedPanel := (panelType == "offer" && ui.focusedPanel == 0) || (panelType == "request" && ui.focusedPanel == 1)

	for i := 0; i < len(items) && i < ui.gridCols*ui.gridRows; i++ {
		row := i / ui.gridCols
		col := i % ui.gridCols

		x := gridStartX + col*ui.slotSize
		y := startY + row*ui.slotSize

		// Determine if slot is selected
		isSelected := false
		for _, sel := range selectedSlots {
			if sel == i {
				isSelected = true
				break
			}
		}

		// Determine if this is the cursor position
		isCursor := isFocusedPanel && row == ui.cursorRow && col == ui.cursorCol

		// Draw slot background
		bgColor := color.RGBA{50, 50, 50, 255}
		if isSelected {
			bgColor = color.RGBA{100, 150, 100, 255}
		} else if isCursor {
			bgColor = color.RGBA{100, 100, 150, 255} // Blue for cursor
		} else if ui.hoveredSlot == i {
			bgColor = color.RGBA{80, 80, 80, 255}
		}

		slotImg := ebiten.NewImage(ui.slotSize-4, ui.slotSize-4)
		slotImg.Fill(bgColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+2), float64(y+2))
		screen.DrawImage(slotImg, op)

		// Draw cursor border if focused
		if isCursor {
			borderColor := color.RGBA{200, 200, 255, 255}
			// Top border
			borderImg := ebiten.NewImage(ui.slotSize-4, 2)
			borderImg.Fill(borderColor)
			opBorder := &ebiten.DrawImageOptions{}
			opBorder.GeoM.Translate(float64(x+2), float64(y+2))
			screen.DrawImage(borderImg, opBorder)
			// Bottom border
			opBorder2 := &ebiten.DrawImageOptions{}
			opBorder2.GeoM.Translate(float64(x+2), float64(y+ui.slotSize-6))
			screen.DrawImage(borderImg, opBorder2)
		}

		// Draw item info
		itm := items[i]
		nameLen := len(itm.Name)
		if nameLen > 10 {
			nameLen = 10
		}
		ebitenutil.DebugPrintAt(screen, itm.Name[:nameLen], x+5, y+20)

		// Draw rarity indicator
		rarityColor := ui.getRarityColor(itm.Rarity)
		rarityImg := ebiten.NewImage(ui.slotSize-8, 3)
		rarityImg.Fill(rarityColor)
		opRarity := &ebiten.DrawImageOptions{}
		opRarity.GeoM.Translate(float64(x+4), float64(y+ui.slotSize-8))
		screen.DrawImage(rarityImg, opRarity)
	}
}

// drawPendingState draws the pending state (waiting for response).
func (ui *TradeUI) drawPendingState(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2
	centerY := ui.screenHeight / 2

	partnerName := ui.getEntityName(ui.partnerEntity)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Waiting for %s to respond...", partnerName), centerX-150, centerY-40)
	ebitenutil.DebugPrintAt(screen, "Press ESCAPE to cancel", centerX-100, centerY)

	ui.cancelButton.Draw(screen)
}

// drawNegotiationState draws the negotiation state (reviewing trade).
func (ui *TradeUI) drawNegotiationState(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2

	partnerName := ui.getEntityName(ui.partnerEntity)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s wants to trade", partnerName), centerX-100, 40)

	// Show offered items
	ebitenutil.DebugPrintAt(screen, "They offer:", centerX-280, 100)
	ui.drawTradeItems(screen, ui.selectedOfferedSlots, ui.getInventoryComponent(ui.partnerEntity), 130)

	// Show requested items
	ebitenutil.DebugPrintAt(screen, "They request:", centerX-280, 280)
	ui.drawTradeItems(screen, ui.selectedRequestedSlots, ui.getInventoryComponent(ui.playerEntity), 310)

	// Draw buttons
	ui.acceptButton.Draw(screen)
	ui.rejectButton.Draw(screen)
}

// drawTradeItems draws a list of items in a trade.
func (ui *TradeUI) drawTradeItems(screen *ebiten.Image, itemIndices []int, inv *InventoryComponent, startY int) {
	centerX := ui.screenWidth / 2

	if inv == nil || len(itemIndices) == 0 {
		ebitenutil.DebugPrintAt(screen, "  (nothing)", centerX-260, startY)
		return
	}

	y := startY
	for _, idx := range itemIndices {
		if idx >= 0 && idx < len(inv.Items) {
			itm := inv.Items[idx]
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  • %s (%s)", itm.Name, itm.Rarity.String()), centerX-260, y)
			y += 20
		}
	}
}

// drawCompletedState draws the completed state.
func (ui *TradeUI) drawCompletedState(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2
	centerY := ui.screenHeight / 2

	ebitenutil.DebugPrintAt(screen, "Trade completed successfully!", centerX-120, centerY-20)
	ebitenutil.DebugPrintAt(screen, "Press ESCAPE to close", centerX-100, centerY+20)
}

// drawCancelledState draws the cancelled state.
func (ui *TradeUI) drawCancelledState(screen *ebiten.Image) {
	centerX := ui.screenWidth / 2
	centerY := ui.screenHeight / 2

	ebitenutil.DebugPrintAt(screen, "Trade cancelled", centerX-80, centerY-20)
	ebitenutil.DebugPrintAt(screen, "Press ESCAPE to close", centerX-100, centerY+20)
}

// drawStatusMessage draws the status message at the bottom.
func (ui *TradeUI) drawStatusMessage(screen *ebiten.Image) {
	if ui.statusMessage == "" {
		return
	}

	centerX := ui.screenWidth / 2
	// Draw with color (simplified - just draw text)
	ebitenutil.DebugPrintAt(screen, ui.statusMessage, centerX-len(ui.statusMessage)*3, ui.statusMessageY)
}

// handlePartnerSelectionInput handles keyboard input for partner selection.
func (ui *TradeUI) handlePartnerSelectionInput() {
	if len(ui.nearbyPartners) == 0 {
		return
	}

	// Arrow keys to navigate
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		ui.selectedPartnerIndex--
		if ui.selectedPartnerIndex < 0 {
			ui.selectedPartnerIndex = len(ui.nearbyPartners) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		ui.selectedPartnerIndex++
		if ui.selectedPartnerIndex >= len(ui.nearbyPartners) {
			ui.selectedPartnerIndex = 0
		}
	}

	// Enter to confirm
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if ui.selectedPartnerIndex >= 0 && ui.selectedPartnerIndex < len(ui.nearbyPartners) {
			ui.selectPartner(ui.nearbyPartners[ui.selectedPartnerIndex])
		}
	}
}

// handleItemSelectionInput handles keyboard input for item selection.
func (ui *TradeUI) handleItemSelectionInput() {
	// Get item count for current panel
	itemCount := ui.getItemCountForPanel(ui.focusedPanel)
	if itemCount == 0 {
		itemCount = 1 // Prevent division by zero
	}

	// Arrow key navigation within grid
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		ui.cursorCol--
		if ui.cursorCol < 0 {
			ui.cursorCol = ui.gridCols - 1
			ui.cursorRow--
			if ui.cursorRow < 0 {
				ui.cursorRow = ui.gridRows - 1
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		ui.cursorCol++
		if ui.cursorCol >= ui.gridCols {
			ui.cursorCol = 0
			ui.cursorRow++
			if ui.cursorRow >= ui.gridRows {
				ui.cursorRow = 0
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		ui.cursorRow--
		if ui.cursorRow < 0 {
			ui.cursorRow = ui.gridRows - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		ui.cursorRow++
		if ui.cursorRow >= ui.gridRows {
			ui.cursorRow = 0
		}
	}

	// Tab to switch between offer/request panels
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		ui.focusedPanel = 1 - ui.focusedPanel // Toggle between 0 and 1
		ui.cursorRow = 0
		ui.cursorCol = 0
	}

	// Space to toggle item selection
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		slotIndex := ui.cursorRow*ui.gridCols + ui.cursorCol
		if slotIndex < itemCount {
			ui.toggleSlotSelection(ui.focusedPanel, slotIndex)
		}
	}

	// Enter to confirm selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ui.confirmItemSelection()
	}

	// Escape to cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.cancelTrade()
	}
}

// getItemCountForPanel returns the number of items in the specified panel.
func (ui *TradeUI) getItemCountForPanel(panel int) int {
	var entity *Entity
	if panel == 0 {
		entity = ui.playerEntity
	} else {
		entity = ui.partnerEntity
	}

	inv := ui.getInventoryComponent(entity)
	if inv == nil {
		return 0
	}
	return len(inv.Items)
}

// toggleSlotSelection toggles the selection state of a slot.
func (ui *TradeUI) toggleSlotSelection(panel, slotIndex int) {
	var selectedSlots *[]int
	if panel == 0 {
		selectedSlots = &ui.selectedOfferedSlots
	} else {
		selectedSlots = &ui.selectedRequestedSlots
	}

	// Check if already selected
	for i, sel := range *selectedSlots {
		if sel == slotIndex {
			// Remove from selection
			*selectedSlots = append((*selectedSlots)[:i], (*selectedSlots)[i+1:]...)
			return
		}
	}

	// Add to selection
	*selectedSlots = append(*selectedSlots, slotIndex)
}

// handleNegotiationInput handles keyboard input during negotiation.
func (ui *TradeUI) handleNegotiationInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ui.acceptTrade()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.rejectTrade()
	}
}

// handleTouchInput handles touch/mouse input.
func (ui *TradeUI) handleTouchInput() {
	// Handle close button
	ui.closeButton.Update()

	// Get mouse/touch position
	mouseX, mouseY := ebiten.CursorPosition()

	// Handle state-specific buttons and clicks
	switch ui.state {
	case TradeUIStateSelectingPartner:
		ui.handlePartnerListClick(mouseX, mouseY)
	case TradeUIStateSelectingItems:
		ui.confirmOfferButton.Update()
		ui.cancelButton.Update()
		ui.handleItemSlotClick(mouseX, mouseY)
		ui.updateHoveredSlot(mouseX, mouseY)
	case TradeUIStatePending:
		ui.cancelButton.Update()
	case TradeUIStateNegotiating:
		ui.acceptButton.Update()
		ui.rejectButton.Update()
	}
}

// handlePartnerListClick handles clicks on the partner selection list.
func (ui *TradeUI) handlePartnerListClick(mouseX, mouseY int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	centerX := ui.screenWidth / 2
	slotX := centerX - 200
	slotWidth := 400
	slotHeight := 50

	y := ui.partnerListY
	for i, partner := range ui.nearbyPartners {
		// Check if click is within this partner slot
		if mouseX >= slotX && mouseX < slotX+slotWidth &&
			mouseY >= y && mouseY < y+slotHeight {
			ui.selectedPartnerIndex = i
			ui.selectPartner(partner)
			return
		}
		y += 60
	}
}

// handleItemSlotClick handles clicks on item slots in offer/request panels.
func (ui *TradeUI) handleItemSlotClick(mouseX, mouseY int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	// Check offer panel
	slotIndex := ui.getSlotIndexAt(mouseX, mouseY, ui.offerPanelY)
	if slotIndex >= 0 {
		itemCount := ui.getItemCountForPanel(0)
		if slotIndex < itemCount {
			ui.toggleSlotSelection(0, slotIndex)
			ui.focusedPanel = 0
		}
		return
	}

	// Check request panel
	slotIndex = ui.getSlotIndexAt(mouseX, mouseY, ui.requestPanelY)
	if slotIndex >= 0 {
		itemCount := ui.getItemCountForPanel(1)
		if slotIndex < itemCount {
			ui.toggleSlotSelection(1, slotIndex)
			ui.focusedPanel = 1
		}
		return
	}
}

// getSlotIndexAt returns the slot index at the given coordinates, or -1 if none.
func (ui *TradeUI) getSlotIndexAt(mouseX, mouseY, panelY int) int {
	// Check if within panel Y range
	if mouseY < panelY || mouseY >= panelY+ui.gridRows*ui.slotSize {
		return -1
	}

	// Check if within panel X range
	if mouseX < ui.gridStartX || mouseX >= ui.gridStartX+ui.gridCols*ui.slotSize {
		return -1
	}

	// Calculate row and column
	col := (mouseX - ui.gridStartX) / ui.slotSize
	row := (mouseY - panelY) / ui.slotSize

	if col < 0 || col >= ui.gridCols || row < 0 || row >= ui.gridRows {
		return -1
	}

	return row*ui.gridCols + col
}

// updateHoveredSlot updates the hovered slot based on mouse position.
func (ui *TradeUI) updateHoveredSlot(mouseX, mouseY int) {
	// Check offer panel
	slotIndex := ui.getSlotIndexAt(mouseX, mouseY, ui.offerPanelY)
	if slotIndex >= 0 {
		ui.hoveredSlot = slotIndex
		return
	}

	// Check request panel
	slotIndex = ui.getSlotIndexAt(mouseX, mouseY, ui.requestPanelY)
	if slotIndex >= 0 {
		ui.hoveredSlot = slotIndex + ui.gridCols*ui.gridRows // Offset for request panel
		return
	}

	ui.hoveredSlot = -1
}

// findNearbyPartners finds all nearby players within trade range.
func (ui *TradeUI) findNearbyPartners() {
	ui.nearbyPartners = make([]*Entity, 0)
	ui.selectedPartnerIndex = -1

	if ui.playerEntity == nil {
		return
	}

	playerPos := ui.getPositionComponent(ui.playerEntity)
	if playerPos == nil {
		return
	}

	// Find all entities with player component (or name containing "Player")
	entities := ui.world.GetEntities() // Get all entities
	for _, entity := range entities {
		if entity.ID == ui.playerEntity.ID {
			continue // Skip self
		}

		// Check if entity is a player (has input component or specific tag)
		if !ui.isPlayer(entity) {
			continue
		}

		// Check proximity (use squared distance for efficiency)
		distance := ui.getDistance(ui.playerEntity, entity)
		if distance <= ProposalProximity*ProposalProximity {
			ui.nearbyPartners = append(ui.nearbyPartners, entity)
		}
	}

	if len(ui.nearbyPartners) > 0 {
		ui.selectedPartnerIndex = 0
	}
}

// selectPartner selects a trade partner and moves to item selection.
func (ui *TradeUI) selectPartner(partner *Entity) {
	ui.partnerEntity = partner
	ui.state = TradeUIStateSelectingItems
	ui.resetSelection()
	ui.setStatusMessage("Select items to offer and request", color.White)
}

// confirmItemSelection confirms the item selection and proposes the trade.
func (ui *TradeUI) confirmItemSelection() {
	if ui.partnerEntity == nil {
		ui.setStatusMessage("No partner selected", color.RGBA{255, 100, 100, 255})
		return
	}

	// Get item IDs from selected slots
	playerInv := ui.getInventoryComponent(ui.playerEntity)
	partnerInv := ui.getInventoryComponent(ui.partnerEntity)

	if playerInv == nil {
		ui.setStatusMessage("No inventory", color.RGBA{255, 100, 100, 255})
		return
	}

	offeredItemIDs := make([]string, 0)
	for _, idx := range ui.selectedOfferedSlots {
		if idx >= 0 && idx < len(playerInv.Items) {
			offeredItemIDs = append(offeredItemIDs, playerInv.Items[idx].ID)
		}
	}

	requestedItemIDs := make([]string, 0)
	if partnerInv != nil {
		for _, idx := range ui.selectedRequestedSlots {
			if idx >= 0 && idx < len(partnerInv.Items) {
				requestedItemIDs = append(requestedItemIDs, partnerInv.Items[idx].ID)
			}
		}
	}

	ui.proposeTrade()
}

// proposeTrade proposes the trade to the partner.
func (ui *TradeUI) proposeTrade() {
	if ui.playerEntity == nil || ui.partnerEntity == nil {
		ui.setStatusMessage("Invalid trade participants", color.RGBA{255, 100, 100, 255})
		return
	}

	playerInv := ui.getInventoryComponent(ui.playerEntity)
	partnerInv := ui.getInventoryComponent(ui.partnerEntity)

	if playerInv == nil {
		ui.setStatusMessage("No inventory", color.RGBA{255, 100, 100, 255})
		return
	}

	// Build item ID lists
	offeredItemIDs := make([]string, 0)
	for _, idx := range ui.selectedOfferedSlots {
		if idx >= 0 && idx < len(playerInv.Items) {
			offeredItemIDs = append(offeredItemIDs, playerInv.Items[idx].ID)
		}
	}

	requestedItemIDs := make([]string, 0)
	if partnerInv != nil {
		for _, idx := range ui.selectedRequestedSlots {
			if idx >= 0 && idx < len(partnerInv.Items) {
				requestedItemIDs = append(requestedItemIDs, partnerInv.Items[idx].ID)
			}
		}
	}

	// Propose trade
	err := ui.tradeSystem.ProposeTrade(ui.playerEntity.ID, ui.partnerEntity.ID, offeredItemIDs, requestedItemIDs)
	if err != nil {
		ui.setStatusMessage(fmt.Sprintf("Trade failed: %v", err), color.RGBA{255, 100, 100, 255})
		return
	}

	ui.state = TradeUIStatePending
	ui.setStatusMessage("Trade proposed, waiting for response...", color.RGBA{200, 200, 100, 255})
}

// acceptTrade accepts the current trade proposal.
func (ui *TradeUI) acceptTrade() {
	if ui.playerEntity == nil {
		return
	}

	err := ui.tradeSystem.AcceptTrade(ui.playerEntity.ID)
	if err != nil {
		ui.setStatusMessage(fmt.Sprintf("Accept failed: %v", err), color.RGBA{255, 100, 100, 255})
		return
	}

	// Attempt to commit the trade
	err = ui.tradeSystem.CommitTrade(ui.playerEntity.ID)
	if err != nil {
		ui.setStatusMessage(fmt.Sprintf("Trade failed: %v", err), color.RGBA{255, 100, 100, 255})
		ui.state = TradeUIStateCancelled
		return
	}

	ui.state = TradeUIStateCompleted
	ui.setStatusMessage("Trade completed!", color.RGBA{100, 255, 100, 255})
}

// rejectTrade rejects the current trade proposal.
func (ui *TradeUI) rejectTrade() {
	if ui.playerEntity == nil {
		return
	}

	err := ui.tradeSystem.RejectTrade(ui.playerEntity.ID)
	if err != nil {
		ui.setStatusMessage(fmt.Sprintf("Reject failed: %v", err), color.RGBA{255, 100, 100, 255})
		return
	}

	ui.state = TradeUIStateCancelled
	ui.setStatusMessage("Trade rejected", color.RGBA{200, 200, 100, 255})
}

// cancelTrade cancels the current trade.
func (ui *TradeUI) cancelTrade() {
	if ui.playerEntity == nil {
		return
	}

	err := ui.tradeSystem.CancelTrade(ui.playerEntity.ID)
	if err != nil {
		ui.setStatusMessage(fmt.Sprintf("Cancel failed: %v", err), color.RGBA{255, 100, 100, 255})
		return
	}

	ui.state = TradeUIStateCancelled
	ui.setStatusMessage("Trade cancelled", color.RGBA{200, 200, 100, 255})
}

// updateTradeState checks for trade state updates from the trade system.
func (ui *TradeUI) updateTradeState() {
	if ui.playerEntity == nil {
		return
	}

	comp, ok := ui.playerEntity.GetComponent("trade")
	if !ok {
		return
	}

	tradeComp, ok := comp.(*TradeComponent)
	if !ok || tradeComp.ActiveTrade == nil {
		// No active trade
		if ui.state == TradeUIStatePending {
			ui.state = TradeUIStateCancelled
			ui.setStatusMessage("Trade expired or cancelled", color.RGBA{200, 200, 100, 255})
		}
		return
	}

	proposal := tradeComp.ActiveTrade

	// Update UI state based on trade status
	switch proposal.Status {
	case "pending":
		if ui.state != TradeUIStatePending {
			// We are the proposer
			ui.state = TradeUIStatePending
		}
	case "accepted":
		// Partner accepted, move to negotiation
		if ui.state != TradeUIStateNegotiating {
			ui.state = TradeUIStateNegotiating
			ui.setStatusMessage("Partner accepted! Review trade", color.RGBA{100, 255, 100, 255})
		}
	case "rejected", "cancelled", "failed":
		ui.state = TradeUIStateCancelled
		reason := proposal.FailureReason
		if reason == "" {
			reason = "Trade " + proposal.Status
		}
		ui.setStatusMessage(reason, color.RGBA{255, 100, 100, 255})
	case "committed":
		ui.state = TradeUIStateCompleted
		ui.setStatusMessage("Trade completed!", color.RGBA{100, 255, 100, 255})
	}
}

// Helper functions

func (ui *TradeUI) resetSelection() {
	ui.selectedOfferedSlots = make([]int, 0)
	ui.selectedRequestedSlots = make([]int, 0)
	ui.hoveredSlot = -1
}

func (ui *TradeUI) setStatusMessage(message string, color color.Color) {
	ui.statusMessage = message
	ui.statusColor = color
	ui.messageTimeout = 5.0 // Show for 5 seconds
}

func (ui *TradeUI) getInventoryComponent(entity *Entity) *InventoryComponent {
	if entity == nil {
		return nil
	}
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return nil
	}
	return inv
}

func (ui *TradeUI) getPositionComponent(entity *Entity) *PositionComponent {
	if entity == nil {
		return nil
	}
	comp, ok := entity.GetComponent("position")
	if !ok {
		return nil
	}
	pos, ok := comp.(*PositionComponent)
	if !ok {
		return nil
	}
	return pos
}

func (ui *TradeUI) getEntityName(entity *Entity) string {
	if entity == nil {
		return "Unknown"
	}
	// Try to get name from a NameComponent or similar
	// For now, just use entity ID
	return fmt.Sprintf("Player %d", entity.ID)
}

func (ui *TradeUI) getDistance(e1, e2 *Entity) float64 {
	pos1 := ui.getPositionComponent(e1)
	pos2 := ui.getPositionComponent(e2)

	if pos1 == nil || pos2 == nil {
		return 999.0
	}

	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return (dx*dx + dy*dy) // Sqrt not needed for comparison
}

func (ui *TradeUI) isPlayer(entity *Entity) bool {
	// Check if entity has input component (indicating it's a player)
	_, hasInput := entity.GetComponent("input")
	return hasInput
}

func (ui *TradeUI) getRarityColor(rarity item.Rarity) color.Color {
	switch rarity {
	case item.RarityCommon:
		return color.RGBA{200, 200, 200, 255}
	case item.RarityUncommon:
		return color.RGBA{100, 255, 100, 255}
	case item.RarityRare:
		return color.RGBA{100, 100, 255, 255}
	case item.RarityEpic:
		return color.RGBA{200, 100, 255, 255}
	case item.RarityLegendary:
		return color.RGBA{255, 165, 0, 255}
	default:
		return color.White
	}
}
