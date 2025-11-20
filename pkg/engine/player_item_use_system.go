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
		if !s.canUseItems(entity) {
			continue
		}

		input, ok := s.getInputProvider(entity)
		if !ok || !input.IsUseItemPressed() {
			continue
		}

		inventory, ok := s.getInventory(entity)
		if !ok {
			continue
		}

		selectedIndex := s.determineSelectedItem(entity, inventory)
		if selectedIndex == -1 {
			s.logNoUsableItems(entity.ID)
			continue
		}

		s.executeItemUse(entity.ID, selectedIndex, inventory)
	}
}

func (s *PlayerItemUseSystem) canUseItems(entity *Entity) bool {
	return !entity.HasComponent("dead")
}

func (s *PlayerItemUseSystem) getInputProvider(entity *Entity) (InputProvider, bool) {
	inputComp, ok := entity.GetComponent("input")
	if !ok {
		return nil, false
	}
	input, ok := inputComp.(InputProvider)
	return input, ok
}

func (s *PlayerItemUseSystem) getInventory(entity *Entity) (*InventoryComponent, bool) {
	invComp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, false
	}
	inventory, ok := invComp.(*InventoryComponent)
	return inventory, ok
}

func (s *PlayerItemUseSystem) determineSelectedItem(entity *Entity, inventory *InventoryComponent) int {
	hotbarComp, hasHotbar := entity.GetComponent("hotbar")
	if !hasHotbar {
		return s.findFirstUsableItem(inventory)
	}

	hotbar, ok := hotbarComp.(*HotbarComponent)
	if !ok {
		return s.findFirstUsableItem(inventory)
	}

	selectedIndex := hotbar.LastUsedIndex
	if selectedIndex == -1 || hotbar.GetSlot(selectedIndex) == nil {
		return s.findFirstUsableItem(inventory)
	}

	targetItem := hotbar.GetSlot(selectedIndex)
	return s.findItemInInventory(inventory, targetItem)
}

func (s *PlayerItemUseSystem) logNoUsableItems(entityID uint64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("entityID", entityID).Debug("no usable items in inventory")
	}
}

func (s *PlayerItemUseSystem) executeItemUse(entityID uint64, selectedIndex int, inventory *InventoryComponent) {
	err := s.inventorySystem.UseConsumable(entityID, selectedIndex)
	if err == nil {
		s.logSuccessfulItemUse(entityID, selectedIndex, inventory)
	} else {
		s.logFailedItemUse(entityID, selectedIndex, err)
	}
}

func (s *PlayerItemUseSystem) logSuccessfulItemUse(entityID uint64, selectedIndex int, inventory *InventoryComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		var itemName string
		if selectedIndex < len(inventory.Items) {
			itemName = inventory.Items[selectedIndex].Name
		}
		s.logger.WithFields(logrus.Fields{
			"entityID":  entityID,
			"itemIndex": selectedIndex,
			"itemName":  itemName,
		}).Info("item used")
	}
}

func (s *PlayerItemUseSystem) logFailedItemUse(entityID uint64, selectedIndex int, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":  entityID,
			"itemIndex": selectedIndex,
		}).WithError(err).Warn("failed to use item")
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
	// Type assert with safety check
	hotbar, ok := hotbarComp.(*HotbarComponent)
	if !ok {
		return
	}
	hotbar.LastUsedIndex = slotIndex
}
