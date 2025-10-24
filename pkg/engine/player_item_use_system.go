// Package engine provides player item use handling.
// This file implements PlayerItemUseSystem which connects player input (E key)
// to inventory item usage for consumables and equipment.
package engine

import "log"

// PlayerItemUseSystem processes player item use input (E key).
// It bridges InputSystem and InventorySystem to consume/equip items.
type PlayerItemUseSystem struct {
	inventorySystem *InventorySystem
	world           *World
}

// NewPlayerItemUseSystem creates a new player item use system.
func NewPlayerItemUseSystem(inventorySystem *InventorySystem, world *World) *PlayerItemUseSystem {
	return &PlayerItemUseSystem{
		inventorySystem: inventorySystem,
		world:           world,
	}
}

// Update processes player item use input for all player-controlled entities.
// This system must run AFTER InputSystem.
func (s *PlayerItemUseSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check for input component (player-controlled entities only)
		inputComp, ok := entity.GetComponent("input")
		if !ok {
			continue
		}
		input := inputComp.(*EbitenInput)

		// Check if player pressed use item button
		if !input.UseItemPressed {
			continue
		}

		// Get inventory component
		invComp, ok := entity.GetComponent("inventory")
		if !ok {
			continue // Entity has no inventory
		}
		inventory := invComp.(*InventoryComponent)

		// Get selected item index (for now, use first consumable)
		// TODO: Implement hotbar/selection system
		selectedIndex := s.findFirstUsableItem(inventory)

		if selectedIndex == -1 {
			// No usable item found
			log.Println("No usable items in inventory")
			input.UseItemPressed = false
			continue
		}

		// Use the item through inventory system
		err := s.inventorySystem.UseConsumable(entity.ID, selectedIndex)

		if err == nil {
			log.Printf("Used item at index %d", selectedIndex)
			// Could trigger effects here:
			// - Use animation
			// - Sound effect
			// - Visual feedback
			// - Tutorial progress tracking
		} else {
			log.Printf("Failed to use item: %v", err)
		}

		// Consume the input so it doesn't trigger multiple times
		input.UseItemPressed = false
	}
}

// findFirstUsableItem finds the first consumable item in the inventory.
// Returns -1 if no usable item is found.
func (s *PlayerItemUseSystem) findFirstUsableItem(inventory *InventoryComponent) int {
	for i, item := range inventory.Items {
		// Check if item is a consumable
		if item.IsConsumable() {
			return i
		}
	}
	return -1
}

// SetSelectedItem sets the currently selected item index for quick use.
// This will be used when hotbar system is implemented.
func (s *PlayerItemUseSystem) SetSelectedItem(entity *Entity, index int) {
	// TODO: Store selected index in a component when hotbar is added
	// For now, this is a placeholder for future implementation
	_ = entity
	_ = index
}
