// Package engine provides the commerce transaction system.
// This file implements CommerceSystem which handles buy/sell transactions
// between players and merchant NPCs. The system validates transactions,
// transfers items and gold, and supports server-authoritative validation
// for multiplayer environments.
//
// Design Philosophy:
// - Server-authoritative: all transactions must be validated server-side
// - Atomic operations: transactions either complete fully or fail with rollback
// - Extensible validation: uses TransactionValidator interface for custom rules
// - Integration with existing inventory system for item management
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// TransactionType represents the type of commerce transaction.
type TransactionType int

const (
	// TransactionBuy represents player buying from merchant
	TransactionBuy TransactionType = iota
	// TransactionSell represents player selling to merchant
	TransactionSell
)

// String returns the string representation of a transaction type.
func (t TransactionType) String() string {
	switch t {
	case TransactionBuy:
		return "buy"
	case TransactionSell:
		return "sell"
	default:
		return "unknown"
	}
}

// TransactionResult contains the outcome of a commerce transaction.
type TransactionResult struct {
	Success      bool
	ErrorMessage string
	GoldChanged  int    // Positive = gained, negative = spent
	ItemName     string // Name of item transacted
}

// CommerceSystem manages buy/sell transactions between players and merchants.
type CommerceSystem struct {
	world     *World
	inventory *InventorySystem
	validator TransactionValidator
	logger    *logrus.Entry
}

// NewCommerceSystem creates a new commerce system.
func NewCommerceSystem(world *World, inventorySystem *InventorySystem) *CommerceSystem {
	return NewCommerceSystemWithLogger(world, inventorySystem, nil)
}

// NewCommerceSystemWithLogger creates a new commerce system with a logger.
func NewCommerceSystemWithLogger(world *World, inventorySystem *InventorySystem, logger *logrus.Logger) *CommerceSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "commerce")
	}

	return &CommerceSystem{
		world:     world,
		inventory: inventorySystem,
		validator: NewDefaultTransactionValidator(),
		logger:    logEntry,
	}
}

// SetValidator sets a custom transaction validator.
func (s *CommerceSystem) SetValidator(validator TransactionValidator) {
	s.validator = validator
}

// BuyItem handles a player purchasing an item from a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) BuyItem(playerID, merchantID uint64, merchantItemIndex int) (*TransactionResult, error) {
	// Get player entity
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		return nil, fmt.Errorf("player entity %d not found", playerID)
	}

	// Get merchant entity
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	// Get player inventory
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		return nil, fmt.Errorf("player inventory: %w", err)
	}

	// Get merchant component
	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		return nil, fmt.Errorf("merchant component: %w", err)
	}

	// Validate merchant item index
	if merchantItemIndex < 0 || merchantItemIndex >= len(merchantComp.Inventory) {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}, nil
	}

	// Get the item
	itm := merchantComp.Inventory[merchantItemIndex]
	if itm == nil {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}, nil
	}

	// Calculate price (merchant sell price)
	price := merchantComp.GetSellPrice(itm)

	// Validate transaction
	canBuy, errMsg := s.validator.CanBuyItem(
		playerInvComp.Gold,
		price,
		playerInvComp.IsFull(),
	)

	if !canBuy {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: errMsg,
			ItemName:     itm.Name,
		}, nil
	}

	// Execute transaction (atomic operations)
	// 1. Remove item from merchant
	removedItem := merchantComp.RemoveItem(merchantItemIndex)
	if removedItem == nil {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Failed to remove item from merchant",
		}, nil
	}

	// 2. Deduct gold from player
	playerInvComp.Gold -= price

	// 3. Add item to player inventory
	success := playerInvComp.AddItem(itm)
	if !success {
		// Rollback: return item to merchant and refund gold
		merchantComp.AddItem(removedItem)
		playerInvComp.Gold += price
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Failed to add item to inventory (rollback performed)",
		}, nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":   playerID,
			"merchantID": merchantID,
			"itemName":   itm.Name,
			"price":      price,
			"playerGold": playerInvComp.Gold,
		}).Info("buy transaction completed")
	}

	return &TransactionResult{
		Success:     true,
		GoldChanged: -price,
		ItemName:    itm.Name,
	}, nil
}

// SellItem handles a player selling an item to a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) SellItem(playerID, merchantID uint64, playerItemIndex int) (*TransactionResult, error) {
	// Get player entity
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		return nil, fmt.Errorf("player entity %d not found", playerID)
	}

	// Get merchant entity
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	// Get player inventory
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		return nil, fmt.Errorf("player inventory: %w", err)
	}

	// Get merchant component
	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		return nil, fmt.Errorf("merchant component: %w", err)
	}

	// Validate player item index
	if playerItemIndex < 0 || playerItemIndex >= len(playerInvComp.Items) {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}, nil
	}

	// Get the item
	itm := playerInvComp.Items[playerItemIndex]
	if itm == nil {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}, nil
	}

	// Calculate price (merchant buy price)
	price := merchantComp.GetBuyPrice(itm)

	// Validate transaction
	// Note: merchants have infinite gold in current implementation
	canSell, errMsg := s.validator.CanSellItem(
		0, // merchant gold (not checked in default validator)
		price,
		!merchantComp.CanAddItem(),
	)

	if !canSell {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: errMsg,
			ItemName:     itm.Name,
		}, nil
	}

	// Execute transaction (atomic operations)
	// 1. Remove item from player
	removedItem := playerInvComp.RemoveItem(playerItemIndex)
	if removedItem == nil {
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Failed to remove item from inventory",
		}, nil
	}

	// 2. Add gold to player
	playerInvComp.Gold += price

	// 3. Add item to merchant inventory
	success := merchantComp.AddItem(removedItem)
	if !success {
		// Rollback: return item to player and deduct gold
		playerInvComp.AddItem(removedItem)
		playerInvComp.Gold -= price
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Merchant inventory full (rollback performed)",
		}, nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":   playerID,
			"merchantID": merchantID,
			"itemName":   itm.Name,
			"price":      price,
			"playerGold": playerInvComp.Gold,
		}).Info("sell transaction completed")
	}

	return &TransactionResult{
		Success:     true,
		GoldChanged: price,
		ItemName:    itm.Name,
	}, nil
}

// GetMerchantInventory returns a copy of the merchant's inventory for display.
func (s *CommerceSystem) GetMerchantInventory(merchantID uint64) ([]*item.Item, error) {
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		return nil, err
	}

	// Return a copy to prevent external modification
	inventory := make([]*item.Item, len(merchantComp.Inventory))
	copy(inventory, merchantComp.Inventory)
	return inventory, nil
}

// GetMerchantPrices returns buy and sell prices for an item from a specific merchant.
func (s *CommerceSystem) GetMerchantPrices(merchantID uint64, itm *item.Item) (sellPrice, buyPrice int, err error) {
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		return 0, 0, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		return 0, 0, err
	}

	return merchantComp.GetSellPrice(itm), merchantComp.GetBuyPrice(itm), nil
}

// getInventoryComponent is a helper to retrieve and validate inventory component.
func (s *CommerceSystem) getInventoryComponent(entity *Entity) (*InventoryComponent, error) {
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, fmt.Errorf("entity %d does not have inventory component", entity.ID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, fmt.Errorf("entity %d inventory component has wrong type", entity.ID)
	}
	return invComp, nil
}

// getMerchantComponent is a helper to retrieve and validate merchant component.
func (s *CommerceSystem) getMerchantComponent(entity *Entity) (*MerchantComponent, error) {
	comp, ok := entity.GetComponent("merchant")
	if !ok {
		return nil, fmt.Errorf("entity %d does not have merchant component", entity.ID)
	}
	merchantComp, ok := comp.(*MerchantComponent)
	if !ok {
		return nil, fmt.Errorf("entity %d merchant component has wrong type", entity.ID)
	}
	return merchantComp, nil
}
