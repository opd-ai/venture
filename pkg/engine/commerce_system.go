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

// BuyItem handles a player purchasing an item from a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) BuyItem(playerID, merchantID uint64, merchantItemIndex int) (*TransactionResult, error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":         "buy_item",
			"player_id":         playerID,
			"merchant_id":       merchantID,
			"merchant_item_idx": merchantItemIndex,
		}).Debug("Starting buy transaction")
	}

	// Get player entity
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "buy_item",
				"player_id": playerID,
				"error":     "entity_not_found",
			}).Error("Player entity not found")
		}
		return nil, fmt.Errorf("player entity %d not found", playerID)
	}

	// Get merchant entity
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "buy_item",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	// Get player inventory
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "buy_item",
				"player_id": playerID,
				"error":     "inventory_component_missing",
			}).Error("Failed to get player inventory component")
		}
		return nil, fmt.Errorf("player inventory: %w", err)
	}

	// Get merchant component
	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "buy_item",
				"merchant_id": merchantID,
				"error":       "merchant_component_missing",
			}).Error("Failed to get merchant component")
		}
		return nil, fmt.Errorf("merchant component: %w", err)
	}

	// Validate merchant item index
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
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}, nil
	}

	// Get the item
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
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}, nil
	}

	// Calculate price (merchant sell price)
	price := merchantComp.GetSellPrice(itm)

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

	// Validate transaction
	canBuy, errMsg := s.validator.CanBuyItem(
		playerInvComp.Gold,
		price,
		playerInvComp.IsFull(),
	)

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
		}, nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "buy_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"price":       price,
		}).Debug("Executing buy transaction")
	}

	// Execute transaction (atomic operations)
	// 1. Remove item from merchant
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

	// 2. Deduct gold from player
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

	// 3. Add item to player inventory
	success := playerInvComp.AddItem(itm)
	if !success {
		// Rollback: return item to merchant and refund gold
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

// SellItem handles a player selling an item to a merchant.
// Returns a TransactionResult with success status and details.
func (s *CommerceSystem) SellItem(playerID, merchantID uint64, playerItemIndex int) (*TransactionResult, error) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":       "sell_item",
			"player_id":       playerID,
			"merchant_id":     merchantID,
			"player_item_idx": playerItemIndex,
		}).Debug("Starting sell transaction")
	}

	// Get player entity
	playerEntity, ok := s.world.GetEntity(playerID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "sell_item",
				"player_id": playerID,
				"error":     "entity_not_found",
			}).Error("Player entity not found")
		}
		return nil, fmt.Errorf("player entity %d not found", playerID)
	}

	// Get merchant entity
	merchantEntity, ok := s.world.GetEntity(merchantID)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "sell_item",
				"merchant_id": merchantID,
				"error":       "entity_not_found",
			}).Error("Merchant entity not found")
		}
		return nil, fmt.Errorf("merchant entity %d not found", merchantID)
	}

	// Get player inventory
	playerInvComp, err := s.getInventoryComponent(playerEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "sell_item",
				"player_id": playerID,
				"error":     "inventory_component_missing",
			}).Error("Failed to get player inventory component")
		}
		return nil, fmt.Errorf("player inventory: %w", err)
	}

	// Get merchant component
	merchantComp, err := s.getMerchantComponent(merchantEntity)
	if err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"operation":   "sell_item",
				"merchant_id": merchantID,
				"error":       "merchant_component_missing",
			}).Error("Failed to get merchant component")
		}
		return nil, fmt.Errorf("merchant component: %w", err)
	}

	// Validate player item index
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
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Invalid item index",
		}, nil
	}

	// Get the item
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
		return &TransactionResult{
			Success:      false,
			ErrorMessage: "Item not found",
		}, nil
	}

	// Calculate price (merchant buy price)
	price := merchantComp.GetBuyPrice(itm)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "sell_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"item_type":   itm.Type.String(),
			"item_rarity": itm.Rarity.String(),
			"price":       price,
			"player_gold": playerInvComp.Gold,
		}).Debug("Validating sell transaction")
	}

	// Validate transaction
	// Note: merchants have infinite gold in current implementation
	canSell, errMsg := s.validator.CanSellItem(
		0, // merchant gold (not checked in default validator)
		price,
		!merchantComp.CanAddItem(),
	)

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
		}, nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "sell_item",
			"player_id":   playerID,
			"merchant_id": merchantID,
			"item_name":   itm.Name,
			"price":       price,
		}).Debug("Executing sell transaction")
	}

	// Execute transaction (atomic operations)
	// 1. Remove item from player
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

	// 2. Add gold to player
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

	// 3. Add item to merchant inventory
	success := merchantComp.AddItem(removedItem)
	if !success {
		// Rollback: return item to player and deduct gold
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
