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
	cs := NewCommerceSystemWithLogger(world, inventorySystem, nil)
	if cs.logger != nil {
		cs.logger.Debug("Created commerce system without explicit logger")
	}
	return cs
}

// NewCommerceSystemWithLogger creates a new commerce system with a logger.
func NewCommerceSystemWithLogger(world *World, inventorySystem *InventorySystem, logger *logrus.Logger) *CommerceSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "commerce")
		logEntry.WithFields(logrus.Fields{
			"has_world":     world != nil,
			"has_inventory": inventorySystem != nil,
		}).Debug("Initializing commerce system")
	}

	cs := &CommerceSystem{
		world:     world,
		inventory: inventorySystem,
		validator: NewDefaultTransactionValidator(),
		logger:    logEntry,
	}

	if logEntry != nil {
		logEntry.Info("Commerce system initialized successfully")
	}

	return cs
}

// SetValidator sets a custom transaction validator.
func (s *CommerceSystem) SetValidator(validator TransactionValidator) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"validator_type": fmt.Sprintf("%T", validator),
		}).Debug("Setting custom transaction validator")
	}
	s.validator = validator
	if s.logger != nil {
		s.logger.Info("Transaction validator updated")
	}
}

// logBuyStart logs the start of a buy transaction.
func (s *CommerceSystem) logBuyStart(playerID, merchantID uint64, merchantItemIndex int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":         "buy_item",
			"player_id":         playerID,
			"merchant_id":       merchantID,
			"merchant_item_idx": merchantItemIndex,
		}).Debug("Starting buy transaction")
	}
}

// retrieveBuyEntities retrieves and validates player and merchant entities for buy transaction.
func (s *CommerceSystem) retrieveBuyEntities(playerID, merchantID uint64) (*Entity, *Entity, error) {
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "buy_item",
				"player_id": playerID,
				"error":     "entity_not_found",
			}).Error("Player entity not found")
		}
		return nil, nil, fmt.Errorf("player entity %d not found", playerID)
	}

	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "buy_item",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return nil, nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	return playerEntity, merchantEntity, nil
}

// retrieveBuyComponents retrieves inventory and merchant components for buy transaction.
func (s *CommerceSystem) retrieveBuyComponents(playerEntity, merchantEntity *Entity, playerID, merchantID uint64) (*InventoryComponent, *MerchantComponent, error) {
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "buy_item",
				"player_id": playerID,
				"error":     "inventory_component_missing",
			}).Error("Failed to get player inventory component")
		}
		return nil, nil, fmt.Errorf("player inventory: %w", err)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "buy_item",
				"merchant_id": merchantID,
				"error":       "merchant_component_missing",
			}).Error("Failed to get merchant component")
		}
		return nil, nil, fmt.Errorf("merchant component: %w", err)
	}

	return playerInvComp, merchantComp, nil
}

// validateMerchantItem validates merchant item index and retrieves the item.
func (s *CommerceSystem) validateMerchantItem(merchantComp *MerchantComponent, merchantItemIndex int, merchantID uint64) (*item.Item, *TransactionResult) {
	if merchantItemIndex < 0 || merchantItemIndex >= len(merchantComp.Inventory) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":         "buy_item",
				"merchant_id":       merchantID,
				"merchant_item_idx": merchantItemIndex,
				"inventory_size":    len(merchantComp.Inventory),
				"error":             "invalid_index",
			}).Warn("Invalid merchant item index")
		}
		return nil, &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}
	}

	itm := merchantComp.Inventory[merchantItemIndex]
	if itm == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":         "buy_item",
				"merchant_id":       merchantID,
				"merchant_item_idx": merchantItemIndex,
				"error":             "item_null",
			}).Warn("Item at index is null")
		}
		return nil, &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}
	}

	return itm, nil
}

// validateBuyTransaction validates if the buy transaction can proceed.
func (s *CommerceSystem) validateBuyTransaction(playerID, merchantID uint64, itm *item.Item, price int, playerInvComp *InventoryComponent) *TransactionResult {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "buy_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"item_type":   itm.Type.String(),
			"item_rarity": itm.Rarity.String(),
			"price":       price,
			"player_gold": playerInvComp.Gold,
		}).Debug("Validating buy transaction")
	}

	canBuy, errMsg := s.validator.CanBuyItem(playerInvComp.Gold, price, playerInvComp.IsFull())
	if !canBuy {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":      "buy_item",
				"player_id":      playerID,
				"merchant_id":    merchantID,
				"item_name":      itm.Name,
				"price":          price,
				"player_gold":    playerInvComp.Gold,
				"inventory_full": playerInvComp.IsFull(),
				"reason":         errMsg,
			}).Warn("Buy transaction validation failed")
		}
		return &TransactionResult{
			Success:      false,
			ErrorMessage: errMsg,
			ItemName:     itm.Name,
		}
	}

	return nil
}

// logBuyExecution logs the execution of a buy transaction.
func (s *CommerceSystem) logBuyExecution(playerID, merchantID uint64, itemName string, price int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "buy_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itemName,
			"price":       price,
		}).Debug("Executing buy transaction")
	}
}

// executeBuyTransaction executes the atomic buy transaction with rollback on failure.
func (s *CommerceSystem) executeBuyTransaction(playerID, merchantID uint64, merchantItemIndex int, itm *item.Item, price int, playerInvComp *InventoryComponent, merchantComp *MerchantComponent) (*TransactionResult, error) {
	removedItem := merchantComp.RemoveItem(merchantItemIndex)
	if removedItem == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":         "buy_item",
				"merchant_id":       merchantID,
				"merchant_item_idx": merchantItemIndex,
				"error":             "removal_failed",
			}).Error("Failed to remove item from merchant inventory")
		}
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Failed to remove item from merchant",
		}, nil
	}

	oldGold := playerInvComp.Gold
	playerInvComp.Gold -= price

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "buy_item",
			"player_id":   playerID,
			"old_gold":    oldGold,
			"new_gold":    playerInvComp.Gold,
			"gold_change": -price,
		}).Debug("Player gold deducted")
	}

	success := playerInvComp.AddItem(itm)
	if !success {
		merchantComp.AddItem(removedItem)
		playerInvComp.Gold += price
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "buy_item",
				"player_id":   playerID,
				"merchant_id": merchantID,
				"item_name":   itm.Name,
				"error":       "add_to_inventory_failed",
			}).Error("Buy transaction rolled back - failed to add item to player inventory")
		}
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

// BuyItem handles a player purchasing an item from a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) BuyItem(playerID, merchantID uint64, merchantItemIndex int) (*TransactionResult, error) {
	s.logBuyStart(playerID, merchantID, merchantItemIndex)

	playerEntity, merchantEntity, err := s.retrieveBuyEntities(playerID, merchantID)
	if err != nil {
		return nil, err
	}

	playerInvComp, merchantComp, err := s.retrieveBuyComponents(playerEntity, merchantEntity, playerID, merchantID)
	if err != nil {
		return nil, err
	}

	itm, failResult := s.validateMerchantItem(merchantComp, merchantItemIndex, merchantID)
	if failResult != nil {
		return failResult, nil
	}

	price := merchantComp.GetSellPrice(itm)

	if failResult := s.validateBuyTransaction(playerID, merchantID, itm, price, playerInvComp); failResult != nil {
		return failResult, nil
	}

	s.logBuyExecution(playerID, merchantID, itm.Name, price)

	return s.executeBuyTransaction(playerID, merchantID, merchantItemIndex, itm, price, playerInvComp, merchantComp)
}

// SellItem handles a player selling an item to a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) SellItem(playerID, merchantID uint64, playerItemIndex int) (*TransactionResult, error) {
	s.logSellStart(playerID, merchantID, playerItemIndex)

	playerEntity, merchantEntity, err := s.retrieveSellEntities(playerID, merchantID)
	if err != nil {
		return nil, err
	}

	playerInvComp, merchantComp, err := s.retrieveSellComponents(playerEntity, merchantEntity, playerID, merchantID)
	if err != nil {
		return nil, err
	}

	itm, result := s.validatePlayerItem(playerInvComp, playerItemIndex, playerID)
	if result != nil {
		return result, nil
	}

	price := merchantComp.GetBuyPrice(itm)
	result = s.validateSellTransaction(playerID, merchantID, itm, price, playerInvComp.Gold, merchantComp)
	if result != nil {
		return result, nil
	}

	s.logSellExecution(playerID, merchantID, itm.Name, price)

	return s.executeSellTransaction(playerID, merchantID, playerItemIndex, itm, price, playerInvComp, merchantComp)
}

// logSellStart logs the start of a sell transaction.
func (s *CommerceSystem) logSellStart(playerID, merchantID uint64, playerItemIndex int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":       "sell_item",
			"player_id":       playerID,
			"merchant_id":     merchantID,
			"player_item_idx": playerItemIndex,
		}).Debug("Starting sell transaction")
	}
}

// retrieveSellEntities retrieves and validates player and merchant entities for sell transaction.
func (s *CommerceSystem) retrieveSellEntities(playerID, merchantID uint64) (*Entity, *Entity, error) {
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "sell_item",
				"player_id": playerID,
				"error":     "entity_not_found",
			}).Error("Player entity not found")
		}
		return nil, nil, fmt.Errorf("player entity %d not found", playerID)
	}

	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "sell_item",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return nil, nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	return playerEntity, merchantEntity, nil
}

// retrieveSellComponents retrieves inventory and merchant components for sell transaction.
func (s *CommerceSystem) retrieveSellComponents(playerEntity, merchantEntity *Entity, playerID, merchantID uint64) (*InventoryComponent, *MerchantComponent, error) {
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "sell_item",
				"player_id": playerID,
				"error":     "inventory_component_missing",
			}).Error("Failed to get player inventory component")
		}
		return nil, nil, fmt.Errorf("player inventory: %w", err)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "sell_item",
				"merchant_id": merchantID,
				"error":       "merchant_component_missing",
			}).Error("Failed to get merchant component")
		}
		return nil, nil, fmt.Errorf("merchant component: %w", err)
	}

	return playerInvComp, merchantComp, nil
}

// validatePlayerItem validates player item index and retrieves the item.
func (s *CommerceSystem) validatePlayerItem(playerInvComp *InventoryComponent, playerItemIndex int, playerID uint64) (*item.Item, *TransactionResult) {
	if playerItemIndex < 0 || playerItemIndex >= len(playerInvComp.Items) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":       "sell_item",
				"player_id":       playerID,
				"player_item_idx": playerItemIndex,
				"inventory_size":  len(playerInvComp.Items),
				"error":           "invalid_index",
			}).Warn("Invalid player item index")
		}
		return nil, &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}
	}

	itm := playerInvComp.Items[playerItemIndex]
	if itm == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":       "sell_item",
				"player_id":       playerID,
				"player_item_idx": playerItemIndex,
				"error":           "item_null",
			}).Warn("Item at index is null")
		}
		return nil, &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}
	}

	return itm, nil
}

// validateSellTransaction validates if the sell transaction can proceed.
func (s *CommerceSystem) validateSellTransaction(playerID, merchantID uint64, itm *item.Item, price, playerGold int, merchantComp *MerchantComponent) *TransactionResult {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "sell_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"item_type":   itm.Type.String(),
			"item_rarity": itm.Rarity.String(),
			"price":       price,
			"player_gold": playerGold,
		}).Debug("Validating sell transaction")
	}

	canSell, errMsg := s.validator.CanSellItem(0, price, !merchantComp.CanAddItem())
	if !canSell {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":        "sell_item",
				"player_id":        playerID,
				"merchant_id":      merchantID,
				"item_name":        itm.Name,
				"price":            price,
				"merchant_can_add": merchantComp.CanAddItem(),
				"reason":           errMsg,
			}).Warn("Sell transaction validation failed")
		}
		return &TransactionResult{
			Success:      false,
			ErrorMessage: errMsg,
			ItemName:     itm.Name,
		}
	}

	return nil
}

// logSellExecution logs the execution of a sell transaction.
func (s *CommerceSystem) logSellExecution(playerID, merchantID uint64, itemName string, price int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "sell_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itemName,
			"price":       price,
		}).Debug("Executing sell transaction")
	}
}

// executeSellTransaction executes the atomic sell transaction with rollback on failure.
func (s *CommerceSystem) executeSellTransaction(playerID, merchantID uint64, playerItemIndex int, itm *item.Item, price int, playerInvComp *InventoryComponent, merchantComp *MerchantComponent) (*TransactionResult, error) {
	removedItem := playerInvComp.RemoveItem(playerItemIndex)
	if removedItem == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":       "sell_item",
				"player_id":       playerID,
				"player_item_idx": playerItemIndex,
				"error":           "removal_failed",
			}).Error("Failed to remove item from player inventory")
		}
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Failed to remove item from inventory",
		}, nil
	}

	oldGold := playerInvComp.Gold
	playerInvComp.Gold += price

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "sell_item",
			"player_id":   playerID,
			"old_gold":    oldGold,
			"new_gold":    playerInvComp.Gold,
			"gold_change": price,
		}).Debug("Player gold increased")
	}

	success := merchantComp.AddItem(removedItem)
	if !success {
		playerInvComp.AddItem(removedItem)
		playerInvComp.Gold -= price
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "sell_item",
				"player_id":   playerID,
				"merchant_id": merchantID,
				"item_name":   itm.Name,
				"error":       "merchant_inventory_full",
			}).Error("Sell transaction rolled back - merchant inventory full")
		}
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
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "get_merchant_inventory",
			"merchant_id": merchantID,
		}).Debug("Retrieving merchant inventory")
	}

	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "get_merchant_inventory",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "get_merchant_inventory",
				"merchant_id": merchantID,
				"error":       "component_retrieval_failed",
			}).Error("Failed to retrieve merchant component")
		}
		return nil, err
	}

	// Return a copy to prevent external modification
	inventory := make([]*item.Item, len(merchantComp.Inventory))
	copy(inventory, merchantComp.Inventory)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":      "get_merchant_inventory",
			"merchant_id":    merchantID,
			"inventory_size": len(inventory),
		}).Debug("Merchant inventory retrieved successfully")
	}

	return inventory, nil
}

// GetMerchantPrices returns buy and sell prices for an item from a specific merchant.
func (s *CommerceSystem) GetMerchantPrices(merchantID uint64, itm *item.Item) (sellPrice, buyPrice int, err error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "get_merchant_prices",
			"merchant_id": merchantID,
			"item_name":   itm.Name,
		}).Debug("Retrieving merchant prices for item")
	}

	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "get_merchant_prices",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return 0, 0, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "get_merchant_prices",
				"merchant_id": merchantID,
				"error":       "component_retrieval_failed",
			}).Error("Failed to retrieve merchant component")
		}
		return 0, 0, err
	}

	sellPrice = merchantComp.GetSellPrice(itm)
	buyPrice = merchantComp.GetBuyPrice(itm)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "get_merchant_prices",
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"sell_price":  sellPrice,
			"buy_price":   buyPrice,
		}).Debug("Merchant prices retrieved successfully")
	}

	return sellPrice, buyPrice, nil
}

// getInventoryComponent is a helper to retrieve and validate inventory component.
func (s *CommerceSystem) getInventoryComponent(entity *Entity) (*InventoryComponent, error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "get_inventory_component",
			"entity_id": entity.ID,
		}).Debug("Retrieving inventory component")
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "get_inventory_component",
				"entity_id": entity.ID,
				"error":     "component_missing",
			}).Error("Entity does not have inventory component")
		}
		return nil, fmt.Errorf("entity %d does not have inventory component", entity.ID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":      "get_inventory_component",
				"entity_id":      entity.ID,
				"component_type": fmt.Sprintf("%T", comp),
				"error":          "wrong_type",
			}).Error("Inventory component has wrong type")
		}
		return nil, fmt.Errorf("entity %d inventory component has wrong type", entity.ID)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":  "get_inventory_component",
			"entity_id":  entity.ID,
			"gold":       invComp.Gold,
			"item_count": len(invComp.Items),
		}).Debug("Inventory component retrieved successfully")
	}

	return invComp, nil
}

// getMerchantComponent is a helper to retrieve and validate merchant component.
func (s *CommerceSystem) getMerchantComponent(entity *Entity) (*MerchantComponent, error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "get_merchant_component",
			"entity_id": entity.ID,
		}).Debug("Retrieving merchant component")
	}

	comp, ok := entity.GetComponent("merchant")
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "get_merchant_component",
				"entity_id": entity.ID,
				"error":     "component_missing",
			}).Error("Entity does not have merchant component")
		}
		return nil, fmt.Errorf("entity %d does not have merchant component", entity.ID)
	}
	merchantComp, ok := comp.(*MerchantComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":      "get_merchant_component",
				"entity_id":      entity.ID,
				"component_type": fmt.Sprintf("%T", comp),
				"error":          "wrong_type",
			}).Error("Merchant component has wrong type")
		}
		return nil, fmt.Errorf("entity %d merchant component has wrong type", entity.ID)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":        "get_merchant_component",
			"entity_id":        entity.ID,
			"inventory_size":   len(merchantComp.Inventory),
			"price_multiplier": merchantComp.PriceMultiplier,
			"buyback_pct":      merchantComp.BuyBackPercentage,
		}).Debug("Merchant component retrieved successfully")
	}

	return merchantComp, nil
}

// Update satisfies the System interface but commerce transactions
// are handled via direct method calls (BuyItem, SellItem),
// not per-frame updates.
func (s *CommerceSystem) Update(entities []*Entity, deltaTime float64) {
	// Commerce transactions are event-driven, not tick-based
	// This method is a no-op to satisfy the System interface
}
