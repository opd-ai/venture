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

// ShopMode represents whether the player is buying or selling.
type ShopMode int

const (
	// ShopModeBuy is when player purchases from merchant
	ShopModeBuy ShopMode = iota
	// ShopModeSell is when player sells to merchant
	ShopModeSell
)

// String returns the string representation of shop mode.
func (m ShopMode) String() string {
	switch m {
	case ShopModeBuy:
		return "Buy"
	case ShopModeSell:
		return "Sell"
	default:
		return "Unknown"
	}
}

// ShopUI handles rendering and interaction for the shop screen.
// Displays merchant inventory for purchasing and player inventory for selling.
// Follows the same patterns as EbitenInventoryUI for consistency.
type ShopUI struct {
	visible bool
	mode    ShopMode

	// Entity references
	playerEntity   *Entity
	merchantEntity *Entity

	// System references
	commerceSystem *CommerceSystem
	dialogSystem   *DialogSystem

	// Layout
	screenWidth  int
	screenHeight int
	gridCols     int
	gridRows     int
	slotSize     int
	padding      int

	// Selection
	selectedSlot int // Selected item index in current inventory
	hoveredSlot  int // Hovered item index

	// Transaction feedback
	lastTransactionMessage string
	transactionMessageTime float64 // Time remaining to show message

	// H-002 FIX: Error feedback
	errorState *UIErrorState

	// Touch support
	touchHandler  *mobile.TouchInputHandler
	closeButton   *mobile.TouchButton
	buyTabButton  *mobile.TouchButton
	sellTabButton *mobile.TouchButton
	buyButton     *mobile.TouchButton
	sellButton    *mobile.TouchButton
}

// NewShopUI creates a new shop UI.
// Parameters match the pattern used by NewEbitenInventoryUI.
func NewShopUI(screenWidth, screenHeight int) *ShopUI {
	ui := &ShopUI{
		visible:      false,
		mode:         ShopModeBuy,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		gridCols:     6,
		gridRows:     3,
		slotSize:     64,
		padding:      15,
		selectedSlot: -1,
		hoveredSlot:  -1,
		errorState:   NewUIErrorState(), // H-002 FIX
		touchHandler: mobile.NewTouchInputHandler(),
	}

	// Window dimensions for button positioning
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2 + 200 // Extra for info panel
	windowX := (screenWidth - windowWidth) / 2

	// Create close button (top-right)
	ui.closeButton = mobile.NewTouchButton(
		float64(windowX+windowWidth-54),
		10,
		44, 44,
		"✕",
		func() { ui.Close() },
	)

	// Create Buy/Sell tab buttons (top of window)
	ui.buyTabButton = mobile.NewTouchButton(
		float64(windowX+20),
		60,
		100, 44,
		"Buy",
		func() { ui.mode = ShopModeBuy; ui.selectedSlot = -1 },
	)

	ui.sellTabButton = mobile.NewTouchButton(
		float64(windowX+130),
		60,
		100, 44,
		"Sell",
		func() { ui.mode = ShopModeSell; ui.selectedSlot = -1 },
	)

	// Create Buy/Sell action buttons (bottom-right)
	ui.buyButton = mobile.NewTouchButton(
		float64(screenWidth-164),
		float64(screenHeight-64),
		120, 44,
		"Buy Item",
		func() { ui.attemptBuy() },
	)

	ui.sellButton = mobile.NewTouchButton(
		float64(screenWidth-164),
		float64(screenHeight-64),
		120, 44,
		"Sell Item",
		func() { ui.attemptSell() },
	)

	return ui
}

// SetPlayerEntity sets the player entity for transactions.
func (ui *ShopUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// SetMerchantEntity sets the merchant entity for transactions.
func (ui *ShopUI) SetMerchantEntity(entity *Entity) {
	ui.merchantEntity = entity
}

// SetCommerceSystem sets the commerce system for transactions.
func (ui *ShopUI) SetCommerceSystem(system *CommerceSystem) {
	ui.commerceSystem = system
}

// SetDialogSystem sets the dialog system for integration.
func (ui *ShopUI) SetDialogSystem(system *DialogSystem) {
	ui.dialogSystem = system
}

// Open displays the shop UI for a specific merchant.
// This is typically called when the player selects "Browse your wares" in dialog.
func (ui *ShopUI) Open(merchantEntity *Entity) {
	ui.merchantEntity = merchantEntity
	ui.visible = true
	ui.mode = ShopModeBuy
	ui.selectedSlot = -1
	ui.hoveredSlot = -1
	ui.lastTransactionMessage = ""
	ui.transactionMessageTime = 0
}

// Close hides the shop UI and cleans up state.
func (ui *ShopUI) Close() {
	ui.visible = false
	ui.merchantEntity = nil
	ui.selectedSlot = -1
	ui.hoveredSlot = -1
	ui.lastTransactionMessage = ""
	ui.transactionMessageTime = 0
}

// IsVisible returns whether the shop is currently shown.
func (ui *ShopUI) IsVisible() bool {
	return ui.visible
}

// Toggle shows or hides the shop UI.
// Note: Shop typically opened via dialog, not toggled directly.
func (ui *ShopUI) Toggle() {
	ui.visible = !ui.visible
	if !ui.visible {
		ui.Close()
	}
}

// GetMode returns the current shop mode (buy/sell).
func (ui *ShopUI) GetMode() ShopMode {
	return ui.mode
}

// SetMode sets the shop mode (buy/sell).
func (ui *ShopUI) SetMode(mode ShopMode) {
	ui.mode = mode
	ui.selectedSlot = -1 // Clear selection when switching modes
}

// Update processes input for the shop UI.
// Handles dual-exit navigation (F key + ESC), mode switching (TAB),
// item selection (mouse/keyboard), and transaction confirmation (ENTER/click).
func (ui *ShopUI) Update(entities []*Entity, deltaTime float64) {
	// Update transaction message timer
	if ui.transactionMessageTime > 0 {
		ui.transactionMessageTime -= deltaTime
		if ui.transactionMessageTime < 0 {
			ui.transactionMessageTime = 0
			ui.lastTransactionMessage = ""
		}
	}

	// Update touch handler
	if ui.touchHandler != nil {
		ui.touchHandler.Update()
	}

	// Update all touch buttons
	if ui.closeButton != nil {
		ui.closeButton.Update()
	}
	if ui.buyTabButton != nil {
		ui.buyTabButton.Update()
	}
	if ui.sellTabButton != nil {
		ui.sellTabButton.Update()
	}
	if ui.mode == ShopModeBuy && ui.buyButton != nil {
		ui.buyButton.Update()
	}
	if ui.mode == ShopModeSell && ui.sellButton != nil {
		ui.sellButton.Update()
	}

	// Dual-exit navigation: F key (toggle) OR ESC (close only)
	// Note: Shop uses F key to match merchant interaction semantics
	if shouldClose, shouldToggle := HandleMenuInput(MenuKeys.Shop, ui.visible); shouldClose {
		if shouldToggle {
			ui.Toggle()
		} else {
			ui.Close()
		}
		// Also end dialog if dialog system is set
		if ui.dialogSystem != nil {
			ui.dialogSystem.EndDialog()
		}
		return
	}

	if !ui.visible || ui.playerEntity == nil || ui.merchantEntity == nil {
		return
	}

	// Handle mode switching (TAB key)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ui.mode == ShopModeBuy {
			ui.mode = ShopModeSell
		} else {
			ui.mode = ShopModeBuy
		}
		ui.selectedSlot = -1
		return
	}

	// Get current inventory based on mode
	var currentInventory []*item.Item
	if ui.mode == ShopModeBuy {
		// Show merchant inventory
		if merchantComp, ok := ui.merchantEntity.GetComponent("merchant"); ok {
			// Type assert with safety check
			if merchant, ok := merchantComp.(*MerchantComponent); ok {
				currentInventory = merchant.Inventory
			}
		}
	} else {
		// Show player inventory
		if invComp, ok := ui.playerEntity.GetComponent("inventory"); ok {
			// Type assert with safety check
			if inv, ok := invComp.(*InventoryComponent); ok {
				currentInventory = inv.Items
			}
		}
	}

	// Calculate shop window position
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 150 // Extra for header/footer
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Handle mouse and touch input (Touch support for WASM/mobile)
	mouseX, mouseY, _ := GetTouchOrMousePosition()
	mousePressed := IsTouchOrMouseJustPressed()

	// Check if mouse is over item grid
	gridStartY := windowY + 100 // Below header
	if mouseX >= windowX+ui.padding && mouseX < windowX+windowWidth-ui.padding &&
		mouseY >= gridStartY && mouseY < gridStartY+ui.gridRows*ui.slotSize {

		// Calculate which slot is hovered
		relX := mouseX - (windowX + ui.padding)
		relY := mouseY - gridStartY
		col := relX / ui.slotSize
		row := relY / ui.slotSize

		if col >= 0 && col < ui.gridCols && row >= 0 && row < ui.gridRows {
			slotIndex := row*ui.gridCols + col
			ui.hoveredSlot = slotIndex

			// Select slot on click
			if mousePressed && slotIndex < len(currentInventory) {
				ui.selectedSlot = slotIndex
			}
		}
	} else {
		ui.hoveredSlot = -1
	}

	// Handle keyboard navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		if ui.selectedSlot > 0 {
			ui.selectedSlot--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		if ui.selectedSlot < len(currentInventory)-1 {
			ui.selectedSlot++
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if ui.selectedSlot >= ui.gridCols {
			ui.selectedSlot -= ui.gridCols
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if ui.selectedSlot+ui.gridCols < len(currentInventory) {
			ui.selectedSlot += ui.gridCols
		}
	}

	// Handle transaction confirmation (ENTER or double-click)
	confirmPressed := inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	if (confirmPressed || mousePressed) && ui.selectedSlot >= 0 && ui.selectedSlot < len(currentInventory) {
		ui.executeTransaction()
	}
}

// executeTransaction performs the buy or sell transaction.
// This is an internal helper called by Update when player confirms a transaction.
func (ui *ShopUI) executeTransaction() {
	if ui.commerceSystem == nil {
		// H-002 FIX: Use error state for consistent feedback
		ui.errorState.ShowError("Commerce system not available")
		return
	}

	var result *TransactionResult
	var err error

	if ui.mode == ShopModeBuy {
		// Buy from merchant
		result, err = ui.commerceSystem.BuyItem(
			ui.playerEntity.ID,
			ui.merchantEntity.ID,
			ui.selectedSlot,
		)
	} else {
		// Sell to merchant
		result, err = ui.commerceSystem.SellItem(
			ui.playerEntity.ID,
			ui.merchantEntity.ID,
			ui.selectedSlot,
		)
	}

	if err != nil {
		// H-002 FIX: Use error state for consistent feedback
		ui.errorState.ShowError(fmt.Sprintf("Transaction failed: %v", err))
		return
	}

	if result.Success {
		if ui.mode == ShopModeBuy {
			ui.showMessage(fmt.Sprintf("Bought %s for %d gold", result.ItemName, -result.GoldChanged))
		} else {
			ui.showMessage(fmt.Sprintf("Sold %s for %d gold", result.ItemName, result.GoldChanged))
		}
		ui.selectedSlot = -1 // Clear selection after successful transaction
	} else {
		// H-002 FIX: Use error state for transaction failures
		ui.errorState.ShowError(result.ErrorMessage)
	}
}

// showMessage displays a transaction message for 3 seconds.
func (ui *ShopUI) showMessage(message string) {
	ui.lastTransactionMessage = message
	ui.transactionMessageTime = 3.0
}

// attemptBuy is called by the buy button to purchase the selected item.
func (ui *ShopUI) attemptBuy() {
	if ui.mode != ShopModeBuy || ui.selectedSlot < 0 {
		return
	}
	ui.executeTransaction()
}

// attemptSell is called by the sell button to sell the selected item.
func (ui *ShopUI) attemptSell() {
	if ui.mode != ShopModeSell || ui.selectedSlot < 0 {
		return
	}
	ui.executeTransaction()
}

// Draw renders the shop UI.
// Displays merchant/player inventory grid, prices, gold, and transaction feedback.
func (ui *ShopUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}

	if !ui.visible || ui.playerEntity == nil || ui.merchantEntity == nil {
		return
	}

	playerInv, merchant, ok := ui.getComponents()
	if !ok {
		return
	}

	windowX, windowY, windowWidth := ui.drawWindowBackground(img)
	ui.drawHeader(img, playerInv, merchant, windowX, windowY, windowWidth)
	ui.drawModeInstructions(img, windowX, windowY)

	currentInventory := ui.getCurrentInventory(merchant, playerInv)
	gridStartY := windowY + 100
	ui.drawItemGrid(img, currentInventory, merchant, playerInv, windowX, windowY, gridStartY)

	ui.errorState.DrawError(img)

	// Draw touch buttons
	if ui.closeButton != nil {
		ui.closeButton.Draw(img)
	}
	if ui.buyTabButton != nil {
		ui.buyTabButton.Draw(img)
	}
	if ui.sellTabButton != nil {
		ui.sellTabButton.Draw(img)
	}

	// Draw action button only when item is selected
	if ui.selectedSlot >= 0 {
		if ui.mode == ShopModeBuy && ui.buyButton != nil {
			ui.buyButton.Draw(img)
		} else if ui.mode == ShopModeSell && ui.sellButton != nil {
			ui.sellButton.Draw(img)
		}
	}
}

// getComponents retrieves and validates the player inventory and merchant components.
// Returns the inventory component, merchant component, and a boolean indicating success.
func (ui *ShopUI) getComponents() (*InventoryComponent, *MerchantComponent, bool) {
	playerInvComp, hasPlayerInv := ui.playerEntity.GetComponent("inventory")
	merchantComp, hasMerchant := ui.merchantEntity.GetComponent("merchant")
	if !hasPlayerInv || !hasMerchant {
		return nil, nil, false
	}

	playerInv, ok := playerInvComp.(*InventoryComponent)
	if !ok {
		return nil, nil, false
	}
	merchant, ok := merchantComp.(*MerchantComponent)
	if !ok {
		return nil, nil, false
	}

	return playerInv, merchant, true
}

// drawWindowBackground renders the semi-transparent overlay and window background.
// Returns the window position (x, y) and width for use in subsequent rendering.
func (ui *ShopUI) drawWindowBackground(img *ebiten.Image) (int, int, int) {
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	img.DrawImage(overlay, nil)

	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 150
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{30, 30, 40, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	return windowX, windowY, windowWidth
}

// drawHeader renders the shop title, player gold, exit hint, mode indicator, and transaction message.
// Displays all header information including merchant name, player resources, and navigation hints.
func (ui *ShopUI) drawHeader(img *ebiten.Image, playerInv *InventoryComponent, merchant *MerchantComponent, windowX, windowY, windowWidth int) {
	titleText := fmt.Sprintf("SHOP - %s", merchant.MerchantName)
	if merchant.MerchantName == "" {
		titleText = "SHOP"
	}
	ebitenutil.DebugPrintAt(img, titleText, windowX+10, windowY+10)

	goldText := fmt.Sprintf("Your Gold: %d", playerInv.Gold)
	ebitenutil.DebugPrintAt(img, goldText, windowX+10, windowY+30)

	exitHint := GetExitHint(MenuKeys.Shop)
	ebitenutil.DebugPrintAt(img, exitHint, windowX+windowWidth-200, windowY+10)

	modeText := fmt.Sprintf("Mode: %s (TAB to switch)", ui.mode.String())
	ebitenutil.DebugPrintAt(img, modeText, windowX+windowWidth-200, windowY+30)

	if ui.transactionMessageTime > 0 && ui.lastTransactionMessage != "" {
		ebitenutil.DebugPrintAt(img, ui.lastTransactionMessage, windowX+10, windowY+50)
	}
}

// drawModeInstructions renders the mode-specific instruction text for buying or selling.
// Provides clear user guidance on current interaction mode.
func (ui *ShopUI) drawModeInstructions(img *ebiten.Image, windowX, windowY int) {
	instructionY := windowY + 70
	if ui.mode == ShopModeBuy {
		ebitenutil.DebugPrintAt(img, "Select item to purchase (ENTER to confirm)", windowX+10, instructionY)
	} else {
		ebitenutil.DebugPrintAt(img, "Select item to sell (ENTER to confirm)", windowX+10, instructionY)
	}
}

// getCurrentInventory returns the appropriate inventory based on the current shop mode.
// Returns merchant inventory for buy mode and player inventory for sell mode.
func (ui *ShopUI) getCurrentInventory(merchant *MerchantComponent, playerInv *InventoryComponent) []*item.Item {
	if ui.mode == ShopModeBuy {
		return merchant.Inventory
	}
	return playerInv.Items
}

// drawItemGrid renders the complete item grid with slots, items, prices, and tooltips.
// Handles affordability color-coding, hover/selection states, and tooltip display.
func (ui *ShopUI) drawItemGrid(img *ebiten.Image, currentInventory []*item.Item, merchant *MerchantComponent, playerInv *InventoryComponent, windowX, windowY, gridStartY int) {
	for row := 0; row < ui.gridRows; row++ {
		for col := 0; col < ui.gridCols; col++ {
			slotIndex := row*ui.gridCols + col
			slotX := windowX + ui.padding + col*ui.slotSize
			slotY := gridStartY + row*ui.slotSize

			ui.drawItemSlot(img, slotIndex, slotX, slotY, currentInventory, merchant, playerInv, windowY)
		}
	}
}

// drawItemSlot renders a single inventory slot including background, item, price, and tooltip.
// Applies color-coding based on affordability and hover/selection states.
func (ui *ShopUI) drawItemSlot(img *ebiten.Image, slotIndex, slotX, slotY int, currentInventory []*item.Item, merchant *MerchantComponent, playerInv *InventoryComponent, windowY int) {
	slotColor := ui.calculateSlotColor(slotIndex, currentInventory, merchant, playerInv)

	slot := ebiten.NewImage(ui.slotSize-4, ui.slotSize-4)
	slot.Fill(slotColor)
	slotOpts := &ebiten.DrawImageOptions{}
	slotOpts.GeoM.Translate(float64(slotX), float64(slotY))
	img.DrawImage(slot, slotOpts)

	if slotIndex < len(currentInventory) {
		itm := currentInventory[slotIndex]
		if itm != nil {
			ui.drawItemContent(img, itm, merchant, playerInv, slotIndex, slotX, slotY, windowY)
		}
	}
}

// calculateSlotColor determines the slot background color based on affordability and selection state.
// Returns appropriate color for default, hover, selection, affordable, and unaffordable states.
func (ui *ShopUI) calculateSlotColor(slotIndex int, currentInventory []*item.Item, merchant *MerchantComponent, playerInv *InventoryComponent) color.RGBA {
	slotColor := color.RGBA{50, 50, 60, 255}

	if ui.mode == ShopModeBuy && slotIndex < len(currentInventory) {
		itm := currentInventory[slotIndex]
		if itm != nil {
			price := merchant.GetSellPrice(itm)
			if price > playerInv.Gold {
				slotColor = color.RGBA{80, 40, 40, 255}
			} else {
				slotColor = color.RGBA{40, 70, 40, 255}
			}
		}
	}

	if slotIndex == ui.hoveredSlot {
		slotColor = ui.getHoverColor(slotIndex, currentInventory, merchant, playerInv)
	}
	if slotIndex == ui.selectedSlot {
		slotColor = ui.getSelectionColor(slotIndex, currentInventory, merchant, playerInv)
	}

	return slotColor
}

// getHoverColor returns the appropriate color for a hovered slot based on affordability.
// Provides visual feedback for item affordability on hover in buy mode.
func (ui *ShopUI) getHoverColor(slotIndex int, currentInventory []*item.Item, merchant *MerchantComponent, playerInv *InventoryComponent) color.RGBA {
	if ui.mode == ShopModeBuy && slotIndex < len(currentInventory) {
		itm := currentInventory[slotIndex]
		if itm != nil {
			price := merchant.GetSellPrice(itm)
			if price > playerInv.Gold {
				return color.RGBA{100, 60, 60, 255}
			}
			return color.RGBA{60, 100, 60, 255}
		}
	}
	return color.RGBA{70, 70, 90, 255}
}

// getSelectionColor returns the appropriate color for a selected slot based on affordability.
// Provides strong visual feedback for item affordability on selection in buy mode.
func (ui *ShopUI) getSelectionColor(slotIndex int, currentInventory []*item.Item, merchant *MerchantComponent, playerInv *InventoryComponent) color.RGBA {
	if ui.mode == ShopModeBuy && slotIndex < len(currentInventory) {
		itm := currentInventory[slotIndex]
		if itm != nil {
			price := merchant.GetSellPrice(itm)
			if price > playerInv.Gold {
				return color.RGBA{120, 70, 70, 255}
			}
			return color.RGBA{70, 120, 70, 255}
		}
	}
	return color.RGBA{90, 90, 120, 255}
}

// drawItemContent renders the item icon, price label, and tooltip for a populated slot.
// Handles mode-specific pricing and affordability indicators.
func (ui *ShopUI) drawItemContent(img *ebiten.Image, itm *item.Item, merchant *MerchantComponent, playerInv *InventoryComponent, slotIndex, slotX, slotY, windowY int) {
	itemText := string(itm.Name[0])
	ebitenutil.DebugPrintAt(img, itemText, slotX+24, slotY+24)

	var price int
	if ui.mode == ShopModeBuy {
		price = merchant.GetSellPrice(itm)
	} else {
		price = merchant.GetBuyPrice(itm)
	}
	priceText := fmt.Sprintf("%dg", price)
	ebitenutil.DebugPrintAt(img, priceText, slotX+5, slotY+ui.slotSize-15)

	if slotIndex == ui.hoveredSlot {
		ui.drawTooltip(img, itm, price, playerInv, slotX, slotY, windowY)
	}
}

// drawTooltip renders a detailed tooltip showing item name, value, and purchase/sell price.
// Displays affordability indicators in buy mode and adjusts position to stay within window bounds.
func (ui *ShopUI) drawTooltip(img *ebiten.Image, itm *item.Item, price int, playerInv *InventoryComponent, slotX, slotY, windowY int) {
	tooltipX := slotX
	tooltipY := slotY - 60
	if tooltipY < windowY {
		tooltipY = slotY + ui.slotSize + 5
	}

	tooltipBg := ebiten.NewImage(220, 50)
	tooltipBg.Fill(color.RGBA{20, 20, 30, 250})
	tooltipOpts := &ebiten.DrawImageOptions{}
	tooltipOpts.GeoM.Translate(float64(tooltipX), float64(tooltipY))
	img.DrawImage(tooltipBg, tooltipOpts)

	ebitenutil.DebugPrintAt(img, itm.Name, tooltipX+5, tooltipY+5)
	ebitenutil.DebugPrintAt(img, fmt.Sprintf("Value: %d", itm.Stats.Value), tooltipX+5, tooltipY+20)

	if ui.mode == ShopModeBuy {
		priceText := fmt.Sprintf("Buy Price: %d gold", price)
		if price > playerInv.Gold {
			priceText += " (TOO EXPENSIVE)"
		} else {
			priceText += " (CAN AFFORD)"
		}
		ebitenutil.DebugPrintAt(img, priceText, tooltipX+5, tooltipY+35)
	} else {
		ebitenutil.DebugPrintAt(img, fmt.Sprintf("Sell Price: %d gold", price), tooltipX+5, tooltipY+35)
	}
}
