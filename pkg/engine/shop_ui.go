package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
}

// NewShopUI creates a new shop UI.
// Parameters match the pattern used by NewEbitenInventoryUI.
func NewShopUI(screenWidth, screenHeight int) *ShopUI {
	return &ShopUI{
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
	}
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
// Handles dual-exit navigation (S key + ESC), mode switching (TAB),
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

	// Dual-exit navigation: S key (toggle) OR ESC (close only)
	// Note: Shop uses S key by convention, matching inventory (I), character (C), etc.
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
			merchant := merchantComp.(*MerchantComponent)
			currentInventory = merchant.Inventory
		}
	} else {
		// Show player inventory
		if invComp, ok := ui.playerEntity.GetComponent("inventory"); ok {
			inv := invComp.(*InventoryComponent)
			currentInventory = inv.Items
		}
	}

	// Calculate shop window position
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 150 // Extra for header/footer
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()
	mousePressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

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
		ui.showMessage("Commerce system not available")
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
		ui.showMessage(fmt.Sprintf("Error: %v", err))
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
		ui.showMessage(result.ErrorMessage)
	}
}

// showMessage displays a transaction message for 3 seconds.
func (ui *ShopUI) showMessage(message string) {
	ui.lastTransactionMessage = message
	ui.transactionMessageTime = 3.0
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

	// Get components
	playerInvComp, hasPlayerInv := ui.playerEntity.GetComponent("inventory")
	merchantComp, hasMerchant := ui.merchantEntity.GetComponent("merchant")
	if !hasPlayerInv || !hasMerchant {
		return
	}

	playerInv := playerInvComp.(*InventoryComponent)
	merchant := merchantComp.(*MerchantComponent)

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	img.DrawImage(overlay, nil)

	// Calculate window position
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 150
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Draw window background
	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{30, 30, 40, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	// Draw title and merchant name
	titleText := fmt.Sprintf("SHOP - %s", merchant.MerchantName)
	if merchant.MerchantName == "" {
		titleText = "SHOP"
	}
	ebitenutil.DebugPrintAt(img, titleText, windowX+10, windowY+10)

	// Draw exit hint (standardized dual-exit navigation)
	exitHint := "Press [S] or [ESC] to close"
	ebitenutil.DebugPrintAt(img, exitHint, windowX+10, windowY+30)

	// Draw mode indicator and switch hint
	modeText := fmt.Sprintf("Mode: %s (TAB to switch)", ui.mode.String())
	ebitenutil.DebugPrintAt(img, modeText, windowX+windowWidth-200, windowY+10)

	// Draw player gold
	goldText := fmt.Sprintf("Your Gold: %d", playerInv.Gold)
	ebitenutil.DebugPrintAt(img, goldText, windowX+windowWidth-150, windowY+30)

	// Draw transaction message if active
	if ui.transactionMessageTime > 0 && ui.lastTransactionMessage != "" {
		ebitenutil.DebugPrintAt(img, ui.lastTransactionMessage, windowX+10, windowY+50)
	}

	// Draw mode-specific instructions
	instructionY := windowY + 70
	if ui.mode == ShopModeBuy {
		ebitenutil.DebugPrintAt(img, "Select item to purchase (ENTER to confirm)", windowX+10, instructionY)
	} else {
		ebitenutil.DebugPrintAt(img, "Select item to sell (ENTER to confirm)", windowX+10, instructionY)
	}

	// Draw item grid
	var currentInventory []*item.Item
	if ui.mode == ShopModeBuy {
		currentInventory = merchant.Inventory
	} else {
		currentInventory = playerInv.Items
	}

	gridStartY := windowY + 100
	for row := 0; row < ui.gridRows; row++ {
		for col := 0; col < ui.gridCols; col++ {
			slotIndex := row*ui.gridCols + col
			slotX := windowX + ui.padding + col*ui.slotSize
			slotY := gridStartY + row*ui.slotSize

			// Draw slot background with selection/hover highlighting
			slotColor := color.RGBA{50, 50, 60, 255}
			if slotIndex == ui.hoveredSlot {
				slotColor = color.RGBA{70, 70, 90, 255}
			}
			if slotIndex == ui.selectedSlot {
				slotColor = color.RGBA{90, 90, 120, 255}
			}

			slot := ebiten.NewImage(ui.slotSize-4, ui.slotSize-4)
			slot.Fill(slotColor)
			slotOpts := &ebiten.DrawImageOptions{}
			slotOpts.GeoM.Translate(float64(slotX), float64(slotY))
			img.DrawImage(slot, slotOpts)

			// Draw item if present
			if slotIndex < len(currentInventory) {
				itm := currentInventory[slotIndex]
				if itm != nil {
					// Draw item icon (first letter of name - simplified)
					itemText := string(itm.Name[0])
					ebitenutil.DebugPrintAt(img, itemText, slotX+24, slotY+24)

					// Calculate and draw price
					var price int
					if ui.mode == ShopModeBuy {
						price = merchant.GetSellPrice(itm)
					} else {
						price = merchant.GetBuyPrice(itm)
					}
					priceText := fmt.Sprintf("%dg", price)
					ebitenutil.DebugPrintAt(img, priceText, slotX+5, slotY+ui.slotSize-15)

					// Draw tooltip on hover
					if slotIndex == ui.hoveredSlot {
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
							ebitenutil.DebugPrintAt(img, fmt.Sprintf("Buy Price: %d gold", price), tooltipX+5, tooltipY+35)
						} else {
							ebitenutil.DebugPrintAt(img, fmt.Sprintf("Sell Price: %d gold", price), tooltipX+5, tooltipY+35)
						}
					}
				}
			}
		}
	}
}
