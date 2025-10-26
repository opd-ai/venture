// Package engine provides player item use handling.
// This file implements PlayerItemUseSystem which connects player input (E key)
// to inventory item usage for consumables and equipment.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// PlayerItemUseSystem processes player item use input (E key).
// It bridges InputSystem and InventorySystem to consume/equip items.
type PlayerItemUseSystem struct {
	inventorySystem *InventorySystem
	world           *World
	logger          *logrus.Entry
}

// NewPlayerItemUseSystem creates a new player item use system.
func NewPlayerItemUseSystem(inventorySystem *InventorySystem, world *World) *PlayerItemUseSystem {
	return &PlayerItemUseSystem{
		inventorySystem: inventorySystem,
		world:           world,
		logger:          nil,
	}
}

// NewPlayerItemUseSystemWithLogger creates a new player item use system with structured logging.
func NewPlayerItemUseSystemWithLogger(inventorySystem *InventorySystem, world *World, logger *logrus.Logger) *PlayerItemUseSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "playerItemUse",
		})
	}
	return &PlayerItemUseSystem{
		inventorySystem: inventorySystem,
		world:           world,
		logger:          logEntry,
	}
}

// Update processes player item use input for all player-controlled entities.
// This system must run AFTER InputSystem.
func (s *PlayerItemUseSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Skip dead entities - they cannot use items (Category 1.1)
		if entity.HasComponent("dead") {
			continue
		}

		// Check for input component (player-controlled entities only)
		inputComp, ok := entity.GetComponent("input")
		if !ok {
			continue
		}
		input, ok := inputComp.(InputProvider)
		if !ok {
			continue // Not an InputProvider
		}

		// Check if player pressed use item button
		if !input.IsUseItemPressed() {
			continue
		}

		// Get inventory component
		invComp, ok := entity.GetComponent("inventory")
		if !ok {
			continue // Entity has no inventory
		}
		inventory := invComp.(*InventoryComponent)

		// Get hotbar component for selected item (if available)
		var selectedIndex int
		if hotbarComp, hasHotbar := entity.GetComponent("hotbar"); hasHotbar {
			hotbar := hotbarComp.(*HotbarComponent)
			selectedIndex = hotbar.LastUsedIndex
			// Check if the slot has an item
			if selectedIndex == -1 || hotbar.GetSlot(selectedIndex) == nil {
				// No selected item, fall back to first consumable
				selectedIndex = s.findFirstUsableItem(inventory)
			} else {
				// Find the hotbar item in inventory
				targetItem := hotbar.GetSlot(selectedIndex)
				selectedIndex = s.findItemInInventory(inventory, targetItem)
			}
		} else {
			// No hotbar, use first consumable
			selectedIndex = s.findFirstUsableItem(inventory)
		}

		if selectedIndex == -1 {
			// No usable item found
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithField("entityID", entity.ID).Debug("no usable items in inventory")
			}
			// Note: Input flag will be cleared by InputSystem on next frame
			continue
		}

		// Use the item through inventory system
		err := s.inventorySystem.UseConsumable(entity.ID, selectedIndex)

		if err == nil {
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
				// Get item name for logging
				var itemName string
				if selectedIndex < len(inventory.Items) {
					itemName = inventory.Items[selectedIndex].Name
				}
				s.logger.WithFields(logrus.Fields{
					"entityID":  entity.ID,
					"itemIndex": selectedIndex,
					"itemName":  itemName,
				}).Info("item used")
			}
			// Could trigger effects here:
			// - Use animation
			// - Sound effect
			// - Visual feedback
			// - Tutorial progress tracking
		} else {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entityID":  entity.ID,
					"itemIndex": selectedIndex,
				}).WithError(err).Warn("failed to use item")
			}
		}

		// Note: Input flag will be cleared by InputSystem on next frame
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

// findItemInInventory finds the inventory index of a specific item.
// Returns -1 if not found.
func (s *PlayerItemUseSystem) findItemInInventory(inventory *InventoryComponent, targetItem *item.Item) int {
	if targetItem == nil {
		return -1
	}
	for i, invItem := range inventory.Items {
		// Compare item references (assuming same pointer means same item)
		if invItem == targetItem {
			return i
		}
	}
	return -1
}

// SetSelectedItem sets the currently selected hotbar slot for quick use.
func (s *PlayerItemUseSystem) SetSelectedItem(entity *Entity, slotIndex int) {
	hotbarComp, hasHotbar := entity.GetComponent("hotbar")
	if !hasHotbar {
		return
	}
	hotbar := hotbarComp.(*HotbarComponent)
	hotbar.LastUsedIndex = slotIndex
}
