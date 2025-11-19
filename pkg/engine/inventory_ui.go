package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// InventoryUI handles rendering and interaction for the inventory screen.
type EbitenInventoryUI struct {
	visible      bool
	world        *World
	playerEntity *Entity

	// Layout
	screenWidth  int
	screenHeight int
	gridCols     int
	gridRows     int
	slotSize     int
	padding      int

	// Selection
	selectedSlot int
	hoveredSlot  int

	// Dragging
	dragging     bool
	draggedIndex int
	dragPreview  *ebiten.Image // Preview image for dragged item

	// System reference for item actions
	inventorySystem *InventorySystem

	// H-002 FIX: Error feedback
	errorState *UIErrorState

	// Touch support
	touchHandler *mobile.TouchInputHandler
	closeButton  *mobile.TouchButton
	scrollOffset float64 // For touch scrolling
}

// NewInventoryUI creates a new inventory UI.
func NewEbitenInventoryUI(world *World, screenWidth, screenHeight int) *EbitenInventoryUI {
	ui := &EbitenInventoryUI{
		visible:      false,
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		gridCols:     8,
		gridRows:     4,
		slotSize:     48,
		padding:      10,
		selectedSlot: -1,
		hoveredSlot:  -1,
		draggedIndex: -1,
		errorState:   NewUIErrorState(), // H-002 FIX
		touchHandler: mobile.NewTouchInputHandler(),
	}

	// Create close button (top-right of window)
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 100
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	ui.closeButton = mobile.NewTouchButton(
		float64(windowX+windowWidth-54),
		float64(windowY+10),
		44, 44,
		"✕",
		func() { ui.Hide() },
	)

	return ui
}

// SetPlayerEntity sets the player entity whose inventory to display.
func (ui *EbitenInventoryUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// SetInventorySystem sets the inventory system for item actions.
func (ui *EbitenInventoryUI) SetInventorySystem(system *InventorySystem) {
	ui.inventorySystem = system
}

// Toggle shows or hides the inventory UI.
func (ui *EbitenInventoryUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the inventory is currently shown.
func (ui *EbitenInventoryUI) IsVisible() bool {
	return ui.visible
}

// Show displays the inventory UI.
func (ui *EbitenInventoryUI) Show() {
	ui.visible = true
}

// Hide hides the inventory UI.
func (ui *EbitenInventoryUI) Hide() {
	ui.visible = false
}

// Update processes input for the inventory UI.
func (ui *EbitenInventoryUI) Update(entities []*Entity, deltaTime float64) {
	ui.updateTouchComponents()

	if ui.handleMenuNavigation() {
		return
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	inventory := ui.getInventoryComponent()
	if inventory == nil {
		return
	}

	windowX, windowY, windowWidth, windowHeight := ui.calculateWindowBounds()
	ui.handleTouchScrolling()

	mouseX, mouseY, _ := GetTouchOrMousePosition()
	mousePressed := IsTouchOrMouseJustPressed()
	mouseReleased := IsTouchOrMouseJustReleased()

	ui.handleMouseHover(mouseX, mouseY, windowX, windowY, windowWidth, windowHeight, inventory, mousePressed)
	ui.handleDragRelease(mouseReleased, inventory)
	ui.handleKeyboardShortcuts(inventory)
}

// updateTouchComponents updates touch-related UI components.
func (ui *EbitenInventoryUI) updateTouchComponents() {
	if ui.touchHandler != nil {
		ui.touchHandler.Update()
	}
	if ui.closeButton != nil {
		ui.closeButton.Update()
	}
}

// handleMenuNavigation processes menu navigation input and returns whether to exit early.
func (ui *EbitenInventoryUI) handleMenuNavigation() bool {
	if shouldClose, shouldToggle := HandleMenuInput(MenuKeys.Inventory, ui.visible); shouldClose {
		if shouldToggle {
			ui.Toggle()
		} else {
			ui.Hide()
		}
		return true
	}
	return false
}

// getInventoryComponent retrieves and validates the inventory component.
func (ui *EbitenInventoryUI) getInventoryComponent() *InventoryComponent {
	invComp, ok := ui.playerEntity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inventory, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil
	}
	return inventory
}

// calculateWindowBounds computes the inventory window dimensions and position.
func (ui *EbitenInventoryUI) calculateWindowBounds() (windowX, windowY, windowWidth, windowHeight int) {
	windowWidth = ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight = ui.gridRows*ui.slotSize + ui.padding*2 + 100
	windowX = (ui.screenWidth - windowWidth) / 2
	windowY = (ui.screenHeight - windowHeight) / 2
	return windowX, windowY, windowWidth, windowHeight
}

// handleTouchScrolling processes touch swipe gestures for scrolling.
func (ui *EbitenInventoryUI) handleTouchScrolling() {
	if ui.touchHandler == nil {
		return
	}

	direction, distance, detected := ui.touchHandler.GetSwipe()
	if !detected {
		return
	}

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

// handleMouseHover processes mouse hover and click events on inventory slots.
func (ui *EbitenInventoryUI) handleMouseHover(mouseX, mouseY, windowX, windowY, windowWidth, windowHeight int, inventory *InventoryComponent, mousePressed bool) {
	if !ui.isMouseOverGrid(mouseX, mouseY, windowX, windowY, windowWidth, windowHeight) {
		ui.hoveredSlot = -1
		return
	}

	slotIndex := ui.calculateHoveredSlot(mouseX, mouseY, windowX, windowY)
	if slotIndex < 0 {
		ui.hoveredSlot = -1
		return
	}

	ui.hoveredSlot = slotIndex

	if mousePressed {
		ui.handleSlotClick(slotIndex, inventory)
	}
}

// isMouseOverGrid checks if the mouse is within the inventory grid area.
func (ui *EbitenInventoryUI) isMouseOverGrid(mouseX, mouseY, windowX, windowY, windowWidth, windowHeight int) bool {
	return mouseX >= windowX+ui.padding && mouseX < windowX+windowWidth-ui.padding &&
		mouseY >= windowY+ui.padding+60 && mouseY < windowY+windowHeight-ui.padding
}

// calculateHoveredSlot determines which inventory slot is under the mouse cursor.
func (ui *EbitenInventoryUI) calculateHoveredSlot(mouseX, mouseY, windowX, windowY int) int {
	relX := mouseX - (windowX + ui.padding)
	relY := mouseY - (windowY + ui.padding + 60)
	col := relX / ui.slotSize
	row := relY / ui.slotSize

	if col >= 0 && col < ui.gridCols && row >= 0 && row < ui.gridRows {
		return row*ui.gridCols + col
	}
	return -1
}

// handleSlotClick initiates drag-and-drop for the clicked inventory slot.
func (ui *EbitenInventoryUI) handleSlotClick(slotIndex int, inventory *InventoryComponent) {
	if slotIndex >= len(inventory.Items) {
		return
	}

	item := inventory.Items[slotIndex]
	if item != nil {
		ui.dragging = true
		ui.draggedIndex = slotIndex
		ui.selectedSlot = slotIndex
		ui.dragPreview = ui.generateItemPreview(item)
	}
}

// handleDragRelease completes drag-and-drop operations when mouse is released.
func (ui *EbitenInventoryUI) handleDragRelease(mouseReleased bool, inventory *InventoryComponent) {
	if !mouseReleased || !ui.dragging {
		return
	}

	if ui.hoveredSlot >= 0 && ui.hoveredSlot != ui.draggedIndex {
		if ui.hoveredSlot < len(inventory.Items) && ui.draggedIndex < len(inventory.Items) {
			inventory.Items[ui.hoveredSlot], inventory.Items[ui.draggedIndex] = inventory.Items[ui.draggedIndex], inventory.Items[ui.hoveredSlot]
		}
	}

	ui.dragging = false
	ui.draggedIndex = -1
	ui.dragPreview = nil
}

// handleKeyboardShortcuts processes keyboard input for item actions.
func (ui *EbitenInventoryUI) handleKeyboardShortcuts(inventory *InventoryComponent) {
	if ui.selectedSlot < 0 || ui.inventorySystem == nil {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		ui.handleEquipOrUseKey(inventory)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		ui.handleDropKey()
	}
}

// handleEquipOrUseKey processes the E key to equip or use an item.
func (ui *EbitenInventoryUI) handleEquipOrUseKey(inventory *InventoryComponent) {
	if ui.selectedSlot >= len(inventory.Items) {
		return
	}

	item := inventory.Items[ui.selectedSlot]
	if item == nil {
		return
	}

	if item.IsEquippable() {
		if err := ui.inventorySystem.EquipItem(ui.playerEntity.ID, ui.selectedSlot); err != nil {
			ui.errorState.ShowError(fmt.Sprintf("Cannot equip: %v", err))
		}
	} else if item.IsConsumable() {
		if err := ui.inventorySystem.UseConsumable(ui.playerEntity.ID, ui.selectedSlot); err != nil {
			ui.errorState.ShowError(fmt.Sprintf("Cannot use: %v", err))
		}
	}
}

// handleDropKey processes the D key to drop an item.
func (ui *EbitenInventoryUI) handleDropKey() {
	invComp, ok := ui.playerEntity.GetComponent("inventory")
	if !ok {
		return
	}
	inventory, ok := invComp.(*InventoryComponent)
	if !ok {
		return
	}

	if ui.selectedSlot >= len(inventory.Items) {
		return
	}

	if err := ui.inventorySystem.DropItem(ui.playerEntity.ID, ui.selectedSlot); err != nil {
		ui.errorState.ShowError(fmt.Sprintf("Cannot drop: %v", err))
	}
	ui.selectedSlot = -1
}

// Draw renders the inventory UI.
func (ui *EbitenInventoryUI) Draw(screen interface{}) {
	img, inventory := ui.validateAndPrepare(screen)
	if img == nil || inventory == nil {
		return
	}

	windowX, windowY, windowWidth, windowHeight := ui.calculateWindowBounds()
	ui.drawOverlayAndBackground(img, windowX, windowY, windowWidth, windowHeight)
	ui.drawHeader(img, windowX, windowY, windowWidth, inventory)
	ui.drawInventoryGrid(img, windowX, windowY, windowWidth, windowHeight, inventory)
	ui.drawEquipmentSlots(img, windowX, windowY, windowWidth, windowHeight)
	ui.drawFooterAndExtras(img, windowX, windowY, windowWidth, windowHeight)
}

// validateAndPrepare validates the screen and retrieves the inventory component.
func (ui *EbitenInventoryUI) validateAndPrepare(screen interface{}) (*ebiten.Image, *InventoryComponent) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return nil, nil
	}
	if !ui.visible || ui.playerEntity == nil {
		return nil, nil
	}

	invComp, ok := ui.playerEntity.GetComponent("inventory")
	if !ok {
		return nil, nil
	}
	inventory, ok := invComp.(*InventoryComponent)
	if !ok {
		return nil, nil
	}

	return img, inventory
}

// drawOverlayAndBackground renders the semi-transparent overlay and window background.
func (ui *EbitenInventoryUI) drawOverlayAndBackground(img *ebiten.Image, windowX, windowY, windowWidth, windowHeight int) {
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	img.DrawImage(overlay, nil)

	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{40, 40, 50, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)
}

// drawHeader renders the title, exit hint, capacity info, and gold display.
func (ui *EbitenInventoryUI) drawHeader(img *ebiten.Image, windowX, windowY, windowWidth int, inventory *InventoryComponent) {
	ebitenutil.DebugPrintAt(img, "INVENTORY", windowX+10, windowY+10)

	exitHint := GetExitHint(MenuKeys.Inventory)
	ebitenutil.DebugPrintAt(img, exitHint, windowX+10, windowY+30)

	capacityText := fmt.Sprintf("Weight: %.1f / %.1f", inventory.GetCurrentWeight(), inventory.MaxWeight)
	ebitenutil.DebugPrintAt(img, capacityText, windowX+windowWidth-150, windowY+10)

	goldText := fmt.Sprintf("Gold: %d", inventory.Gold)
	ebitenutil.DebugPrintAt(img, goldText, windowX+windowWidth-150, windowY+30)
}

// drawInventoryGrid renders all inventory slots with items and tooltips.
func (ui *EbitenInventoryUI) drawInventoryGrid(img *ebiten.Image, windowX, windowY, windowWidth, windowHeight int, inventory *InventoryComponent) {
	startY := windowY + 60
	for row := 0; row < ui.gridRows; row++ {
		for col := 0; col < ui.gridCols; col++ {
			slotIndex := row*ui.gridCols + col
			slotX := windowX + ui.padding + col*ui.slotSize
			slotY := startY + row*ui.slotSize

			ui.drawInventorySlot(img, slotX, slotY, slotIndex, windowY, inventory)
		}
	}
}

// drawInventorySlot renders a single inventory slot with its item and tooltip.
func (ui *EbitenInventoryUI) drawInventorySlot(img *ebiten.Image, slotX, slotY, slotIndex, windowY int, inventory *InventoryComponent) {
	slotColor := color.RGBA{60, 60, 70, 255}
	if slotIndex == ui.hoveredSlot {
		slotColor = color.RGBA{80, 80, 100, 255}
	}
	if slotIndex == ui.selectedSlot {
		slotColor = color.RGBA{100, 100, 120, 255}
	}

	slot := ebiten.NewImage(ui.slotSize-2, ui.slotSize-2)
	slot.Fill(slotColor)
	slotOpts := &ebiten.DrawImageOptions{}
	slotOpts.GeoM.Translate(float64(slotX), float64(slotY))
	img.DrawImage(slot, slotOpts)

	if slotIndex < len(inventory.Items) {
		item := inventory.Items[slotIndex]
		if item != nil {
			itemText := string(item.Name[0])
			ebitenutil.DebugPrintAt(img, itemText, slotX+16, slotY+16)

			if slotIndex == ui.hoveredSlot {
				ui.drawItemTooltip(img, slotX, slotY, windowY, item)
			}
		}
	}
}

// drawItemTooltip renders the tooltip for a hovered item.
func (ui *EbitenInventoryUI) drawItemTooltip(img *ebiten.Image, slotX, slotY, windowY int, item *item.Item) {
	tooltipX := slotX
	tooltipY := slotY - 40
	if tooltipY < windowY {
		tooltipY = slotY + ui.slotSize + 5
	}

	tooltipBg := ebiten.NewImage(200, 35)
	tooltipBg.Fill(color.RGBA{20, 20, 30, 240})
	tooltipOpts := &ebiten.DrawImageOptions{}
	tooltipOpts.GeoM.Translate(float64(tooltipX), float64(tooltipY))
	img.DrawImage(tooltipBg, tooltipOpts)

	ebitenutil.DebugPrintAt(img, item.Name, tooltipX+5, tooltipY+5)
	ebitenutil.DebugPrintAt(img, fmt.Sprintf("Value: %d", item.Stats.Value), tooltipX+5, tooltipY+20)
}

// drawEquipmentSlots renders the equipment slots section.
func (ui *EbitenInventoryUI) drawEquipmentSlots(img *ebiten.Image, windowX, windowY, windowWidth, windowHeight int) {
	startY := windowY + 60
	equipY := startY + ui.gridRows*ui.slotSize + 20
	ebitenutil.DebugPrintAt(img, "Equipment:", windowX+10, equipY)

	equipComp, hasEquipment := ui.playerEntity.GetComponent("equipment")
	equipSlots := []struct {
		name string
		slot EquipmentSlot
	}{
		{"Weapon", SlotMainHand},
		{"Chest", SlotChest},
		{"Accessory", SlotAccessory1},
	}

	for i, slotInfo := range equipSlots {
		slotX := windowX + ui.padding + i*100
		slotY := equipY + 20
		ui.drawEquipmentSlot(img, slotX, slotY, slotInfo, equipComp, hasEquipment)
	}
}

// drawEquipmentSlot renders a single equipment slot with its equipped item.
func (ui *EbitenInventoryUI) drawEquipmentSlot(img *ebiten.Image, slotX, slotY int, slotInfo struct {
	name string
	slot EquipmentSlot
}, equipComp interface{}, hasEquipment bool,
) {
	slotBg := ebiten.NewImage(90, 40)
	slotBg.Fill(color.RGBA{60, 60, 70, 255})
	slotOpts := &ebiten.DrawImageOptions{}
	slotOpts.GeoM.Translate(float64(slotX), float64(slotY))
	img.DrawImage(slotBg, slotOpts)

	ebitenutil.DebugPrintAt(img, slotInfo.name, slotX+5, slotY+5)

	if hasEquipment {
		equipment, ok := equipComp.(*EquipmentComponent)
		if !ok {
			return
		}
		equipped := equipment.GetEquipped(slotInfo.slot)
		if equipped != nil {
			itemName := equipped.Name
			if len(itemName) > 10 {
				itemName = itemName[:10]
			}
			ebitenutil.DebugPrintAt(img, itemName, slotX+5, slotY+20)
		}
	}
}

// drawFooterAndExtras renders controls hint, drag preview, close button, and error feedback.
func (ui *EbitenInventoryUI) drawFooterAndExtras(img *ebiten.Image, windowX, windowY, windowWidth, windowHeight int) {
	controlsY := windowY + windowHeight - 20
	ebitenutil.DebugPrintAt(img, "I: Close | E: Use/Equip | D: Drop | Click+Drag: Move", windowX+10, controlsY)

	if ui.dragging && ui.dragPreview != nil {
		mouseX, mouseY := ebiten.CursorPosition()
		previewOpts := &ebiten.DrawImageOptions{}
		previewOpts.GeoM.Translate(float64(mouseX-ui.slotSize/2), float64(mouseY-ui.slotSize/2))
		previewOpts.ColorScale.ScaleAlpha(0.7)
		img.DrawImage(ui.dragPreview, previewOpts)
	}

	if ui.closeButton != nil {
		ui.closeButton.Draw(img)
	}

	ui.errorState.DrawError(img)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateItemPreview creates a visual preview image for an item being dragged.
// This provides better visual feedback during drag-and-drop operations.
func (ui *EbitenInventoryUI) generateItemPreview(itm interface{}) *ebiten.Image {
	// Create preview image with same size as slot
	size := ui.slotSize - 2
	preview := ebiten.NewImage(size, size)

	// Determine item color and icon based on item type
	itemColor := color.RGBA{120, 120, 180, 255}   // Default blue-ish
	iconColor := color.RGBA{220, 220, 240, 255}   // Lighter shade for icon
	borderColor := color.RGBA{200, 200, 220, 255} // Border color

	// Try to extract item type
	if itemPtr, ok := itm.(*item.Item); ok {
		// Color based on item type
		switch itemPtr.Type {
		case item.TypeWeapon:
			itemColor = color.RGBA{180, 60, 60, 255} // Red for weapons
			iconColor = color.RGBA{240, 120, 120, 255}
		case item.TypeArmor:
			itemColor = color.RGBA{80, 140, 80, 255} // Green for armor
			iconColor = color.RGBA{140, 200, 140, 255}
		case item.TypeConsumable:
			itemColor = color.RGBA{180, 140, 60, 255} // Orange for consumables
			iconColor = color.RGBA{240, 200, 120, 255}
		case item.TypeAccessory:
			itemColor = color.RGBA{140, 80, 180, 255} // Purple for accessories
			iconColor = color.RGBA{200, 140, 240, 255}
		}

		// Adjust shade based on rarity
		switch itemPtr.Rarity {
		case item.RarityLegendary:
			borderColor = color.RGBA{255, 200, 50, 255} // Gold border
		case item.RarityEpic:
			borderColor = color.RGBA{200, 100, 255, 255} // Purple border
		case item.RarityRare:
			borderColor = color.RGBA{100, 150, 255, 255} // Blue border
		}
	}

	// Fill with item color
	preview.Fill(itemColor)

	// Draw a simple icon shape in the center
	centerX := float32(size / 2)
	centerY := float32(size / 2)
	iconSize := float32(size) * 0.4

	// Draw a circle icon (could be enhanced to draw different shapes per item type)
	vector.DrawFilledCircle(preview, centerX, centerY, iconSize, iconColor, true)

	// Draw icon border/outline
	vector.StrokeCircle(preview, centerX, centerY, iconSize, 1.5, borderColor, true)

	// Draw border around entire preview
	// Top border
	topBorder := ebiten.NewImage(size, 2)
	topBorder.Fill(borderColor)
	preview.DrawImage(topBorder, nil)

	// Bottom border
	bottomOpts := &ebiten.DrawImageOptions{}
	bottomOpts.GeoM.Translate(0, float64(size-2))
	preview.DrawImage(topBorder, bottomOpts)

	// Left border
	leftBorder := ebiten.NewImage(2, size)
	leftBorder.Fill(borderColor)
	preview.DrawImage(leftBorder, nil)

	// Right border
	rightOpts := &ebiten.DrawImageOptions{}
	rightOpts.GeoM.Translate(float64(size-2), 0)
	preview.DrawImage(leftBorder, rightOpts)

	return preview
}

// IsActive returns whether the inventory UI is currently visible.
// Implements UISystem interface.
func (i *EbitenInventoryUI) IsActive() bool {
	return i.visible
}

// SetActive sets whether the inventory UI is visible.
// Implements UISystem interface.
func (i *EbitenInventoryUI) SetActive(active bool) {
	i.visible = active
}

// Compile-time check that EbitenInventoryUI implements UISystem
var _ UISystem = (*EbitenInventoryUI)(nil)
