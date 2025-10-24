// Package engine provides inventory_ui for game UI.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
}

// NewInventoryUI creates a new inventory UI.
func NewEbitenInventoryUI(world *World, screenWidth, screenHeight int) *EbitenInventoryUI {
	return &EbitenInventoryUI{
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
	}
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
	// Always check for toggle key, even when not visible
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		ui.Toggle()
		return // Don't process other input on the same frame as toggle
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Get inventory component
	invComp, ok := ui.playerEntity.GetComponent("inventory")
	if !ok {
		return
	}
	inventory := invComp.(*InventoryComponent)

	// Calculate inventory window position
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 100 // Extra for equipment/stats
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()
	mousePressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	mouseReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	// Check if mouse is over inventory grid
	if mouseX >= windowX+ui.padding && mouseX < windowX+windowWidth-ui.padding &&
		mouseY >= windowY+ui.padding+60 && mouseY < windowY+windowHeight-ui.padding {

		// Calculate which slot is hovered
		relX := mouseX - (windowX + ui.padding)
		relY := mouseY - (windowY + ui.padding + 60)
		col := relX / ui.slotSize
		row := relY / ui.slotSize

		if col >= 0 && col < ui.gridCols && row >= 0 && row < ui.gridRows {
			slotIndex := row*ui.gridCols + col
			ui.hoveredSlot = slotIndex

			// Handle click
			if mousePressed {
				if slotIndex < len(inventory.Items) {
					item := inventory.Items[slotIndex]
					if item != nil {
						// Start dragging
						ui.dragging = true
						ui.draggedIndex = slotIndex
						ui.selectedSlot = slotIndex

						// Generate drag preview
						ui.dragPreview = ui.generateItemPreview(item)
					}
				}
			}
		} else {
			ui.hoveredSlot = -1
		}
	} else {
		ui.hoveredSlot = -1
	}

	// Handle drag release
	if mouseReleased && ui.dragging {
		if ui.hoveredSlot >= 0 && ui.hoveredSlot != ui.draggedIndex {
			// Check if hovering over equipment slot (future enhancement)
			// For now, only handle inventory-to-inventory swaps

			// Swap items (simple implementation)
			// In full implementation, would use InventorySystem methods
			if ui.hoveredSlot < len(inventory.Items) && ui.draggedIndex < len(inventory.Items) {
				// Swap
				inventory.Items[ui.hoveredSlot], inventory.Items[ui.draggedIndex] = inventory.Items[ui.draggedIndex], inventory.Items[ui.hoveredSlot]
			}
		}
		ui.dragging = false
		ui.draggedIndex = -1
		ui.dragPreview = nil // Clear preview
	}

	// Handle keyboard shortcuts
	if inpututil.IsKeyJustPressed(ebiten.KeyE) && ui.selectedSlot >= 0 && ui.inventorySystem != nil {
		// Use/equip selected item
		if ui.selectedSlot < len(inventory.Items) {
			item := inventory.Items[ui.selectedSlot]
			if item != nil {
				if item.IsEquippable() {
					// Try to equip the item
					if err := ui.inventorySystem.EquipItem(ui.playerEntity.ID, ui.selectedSlot); err != nil {
						// Failed to equip (could show error message in UI)
						_ = err
					}
				} else if item.IsConsumable() {
					// Try to use consumable
					if err := ui.inventorySystem.UseConsumable(ui.playerEntity.ID, ui.selectedSlot); err != nil {
						// Failed to use (could show error message in UI)
						_ = err
					}
				}
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) && ui.selectedSlot >= 0 && ui.inventorySystem != nil {
		// Drop selected item
		if ui.selectedSlot < len(inventory.Items) {
			if err := ui.inventorySystem.DropItem(ui.playerEntity.ID, ui.selectedSlot); err != nil {
				// Failed to drop (could show error message in UI)
				_ = err
			}
			// Deselect after dropping
			ui.selectedSlot = -1
		}
	}
}

// Draw renders the inventory UI.
func (ui *EbitenInventoryUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Get inventory component
	invComp, ok := ui.playerEntity.GetComponent("inventory")
	if !ok {
		return
	}
	inventory := invComp.(*InventoryComponent)

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	img.DrawImage(overlay, nil)

	// Calculate window position
	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
	windowHeight := ui.gridRows*ui.slotSize + ui.padding*2 + 100
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Draw window background
	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{40, 40, 50, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	// Draw title
	ebitenutil.DebugPrintAt(img, "INVENTORY", windowX+10, windowY+10)

	// Draw capacity info
	capacityText := fmt.Sprintf("Weight: %.1f / %.1f", inventory.GetCurrentWeight(), inventory.MaxWeight)
	ebitenutil.DebugPrintAt(img, capacityText, windowX+windowWidth-150, windowY+10)

	goldText := fmt.Sprintf("Gold: %d", inventory.Gold)
	ebitenutil.DebugPrintAt(img, goldText, windowX+windowWidth-150, windowY+30)

	// Draw inventory grid
	startY := windowY + 60
	for row := 0; row < ui.gridRows; row++ {
		for col := 0; col < ui.gridCols; col++ {
			slotIndex := row*ui.gridCols + col
			slotX := windowX + ui.padding + col*ui.slotSize
			slotY := startY + row*ui.slotSize

			// Draw slot background
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

			// Draw item if present
			if slotIndex < len(inventory.Items) {
				item := inventory.Items[slotIndex]
				if item != nil {
					// Draw item icon (simplified - just show first letter of name)
					itemText := string(item.Name[0])
					ebitenutil.DebugPrintAt(img, itemText, slotX+16, slotY+16)

					// Draw item name on hover
					if slotIndex == ui.hoveredSlot {
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
				}
			}
		}
	}

	// Draw equipment slots
	equipY := startY + ui.gridRows*ui.slotSize + 20
	ebitenutil.DebugPrintAt(img, "Equipment:", windowX+10, equipY)

	// Get equipment component if exists
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

		// Draw slot
		slotBg := ebiten.NewImage(90, 40)
		slotBg.Fill(color.RGBA{60, 60, 70, 255})
		slotOpts := &ebiten.DrawImageOptions{}
		slotOpts.GeoM.Translate(float64(slotX), float64(slotY))
		img.DrawImage(slotBg, slotOpts)

		ebitenutil.DebugPrintAt(img, slotInfo.name, slotX+5, slotY+5)

		// Show equipped item if present
		if hasEquipment {
			equipment := equipComp.(*EquipmentComponent)
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

	// Draw controls hint
	controlsY := windowY + windowHeight - 20
	ebitenutil.DebugPrintAt(img, "I: Close | E: Use/Equip | D: Drop | Click+Drag: Move", windowX+10, controlsY)

	// Draw drag preview (if dragging)
	if ui.dragging && ui.dragPreview != nil {
		mouseX, mouseY := ebiten.CursorPosition()
		previewOpts := &ebiten.DrawImageOptions{}
		// Center preview on cursor
		previewOpts.GeoM.Translate(float64(mouseX-ui.slotSize/2), float64(mouseY-ui.slotSize/2))
		// Make slightly transparent to show it's being dragged
		previewOpts.ColorScale.ScaleAlpha(0.7)
		img.DrawImage(ui.dragPreview, previewOpts)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateItemPreview creates a visual preview image for an item being dragged.
// This provides better visual feedback during drag-and-drop operations.
func (ui *EbitenInventoryUI) generateItemPreview(item interface{}) *ebiten.Image {
	// Create preview image with same size as slot
	size := ui.slotSize - 2
	preview := ebiten.NewImage(size, size)

	// Determine item color based on rarity/type
	// For now, use a simple color scheme
	itemColor := color.RGBA{120, 120, 180, 255} // Default blue-ish

	// Fill with item color
	preview.Fill(itemColor)

	// Draw border
	borderColor := color.RGBA{200, 200, 220, 255}
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

	// TODO: In future enhancement, could draw actual item icon/sprite here
	// For now, the colored square with border provides clear visual feedback

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
