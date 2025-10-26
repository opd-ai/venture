// Package engine provides commerce and dialog components for the ECS.
// This file defines components and interfaces for merchant NPCs, shop interactions,
// dialog systems, and transaction handling. The commerce system supports both
// fixed-location shopkeepers and nomadic merchants with deterministic spawning.
//
// Design Philosophy:
// - Components contain only data, no behavior
// - Interfaces enable extensibility for future dialog/pricing systems
// - Server-authoritative transactions prevent exploitation in multiplayer
// - Deterministic merchant spawning ensures consistency across clients
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// MerchantType represents the behavior pattern of a merchant NPC.
type MerchantType int

const (
	// MerchantFixed represents stationary shopkeepers in settlements
	MerchantFixed MerchantType = iota
	// MerchantNomadic represents wandering merchants that spawn periodically
	MerchantNomadic
)

// String returns the string representation of a merchant type.
func (m MerchantType) String() string {
	switch m {
	case MerchantFixed:
		return "fixed"
	case MerchantNomadic:
		return "nomadic"
	default:
		return "unknown"
	}
}

// MerchantComponent marks an entity as a merchant and manages their inventory.
// Merchants have their own inventory separate from the player's, with items
// available for purchase. Price multipliers scale based on merchant type,
// location, and player reputation (future enhancement).
type MerchantComponent struct {
	// Inventory of items available for purchase
	Inventory []*item.Item

	// MaxInventory is the maximum number of items the merchant can stock
	MaxInventory int

	// MerchantType determines spawn behavior and characteristics
	MerchantType MerchantType

	// PriceMultiplier affects buy/sell prices (1.0 = base price, 1.5 = 50% markup)
	PriceMultiplier float64

	// BuyBackPercentage is the percentage of value paid when buying from player (0.0-1.0)
	BuyBackPercentage float64

	// RestockTimeSec is how often inventory refreshes (0 = never)
	RestockTimeSec float64

	// LastRestockTime tracks when inventory was last regenerated
	LastRestockTime float64

	// MerchantName is the display name for this merchant
	MerchantName string
}

// Type returns the component type identifier.
func (m *MerchantComponent) Type() string {
	return "merchant"
}

// NewMerchantComponent creates a merchant with default values.
// maxInventory: number of item slots (default 20)
// merchantType: fixed or nomadic behavior
// priceMultiplier: markup on items (default 1.5 = 50% markup)
func NewMerchantComponent(maxInventory int, merchantType MerchantType, priceMultiplier float64) *MerchantComponent {
	if maxInventory <= 0 {
		maxInventory = 20
	}
	if priceMultiplier <= 0 {
		priceMultiplier = 1.5
	}

	return &MerchantComponent{
		Inventory:         make([]*item.Item, 0, maxInventory),
		MaxInventory:      maxInventory,
		MerchantType:      merchantType,
		PriceMultiplier:   priceMultiplier,
		BuyBackPercentage: 0.5, // Merchants buy at 50% of item value
		RestockTimeSec:    300, // Restock every 5 minutes
		LastRestockTime:   0,
		MerchantName:      "Merchant",
	}
}

// GetSellPrice calculates the price to sell an item to a player.
// Applies rarity multipliers and merchant price markup.
func (m *MerchantComponent) GetSellPrice(itm *item.Item) int {
	basePrice := float64(itm.Stats.Value)
	return int(basePrice * m.PriceMultiplier)
}

// GetBuyPrice calculates the price to buy an item from a player.
// Players receive a percentage of the item's base value.
func (m *MerchantComponent) GetBuyPrice(itm *item.Item) int {
	basePrice := float64(itm.Stats.Value)
	return int(basePrice * m.BuyBackPercentage)
}

// CanAddItem checks if merchant can stock another item.
func (m *MerchantComponent) CanAddItem() bool {
	return len(m.Inventory) < m.MaxInventory
}

// AddItem adds an item to merchant inventory.
func (m *MerchantComponent) AddItem(itm *item.Item) bool {
	if !m.CanAddItem() {
		return false
	}
	m.Inventory = append(m.Inventory, itm)
	return true
}

// RemoveItem removes an item from merchant inventory by index.
func (m *MerchantComponent) RemoveItem(index int) *item.Item {
	if index < 0 || index >= len(m.Inventory) {
		return nil
	}
	itm := m.Inventory[index]
	m.Inventory = append(m.Inventory[:index], m.Inventory[index+1:]...)
	return itm
}

// NeedsRestock checks if merchant should refresh their inventory.
func (m *MerchantComponent) NeedsRestock(currentTime float64) bool {
	if m.RestockTimeSec <= 0 {
		return false
	}
	return currentTime-m.LastRestockTime >= m.RestockTimeSec
}

// DialogOption represents a choice in a dialog interaction.
type DialogOption struct {
	// Text displayed to the player
	Text string

	// Action triggered when option is selected
	Action DialogAction

	// Enabled determines if option can be selected
	Enabled bool
}

// DialogAction represents the type of action triggered by a dialog choice.
type DialogAction int

const (
	// ActionNone does nothing (for informational dialogs)
	ActionNone DialogAction = iota
	// ActionOpenShop opens the merchant's shop interface
	ActionOpenShop
	// ActionCloseDialog exits the current dialog
	ActionCloseDialog
	// ActionStartQuest initiates a quest (future enhancement)
	ActionStartQuest
	// ActionGiveItem gives an item to the player (future enhancement)
	ActionGiveItem
)

// String returns the string representation of a dialog action.
func (d DialogAction) String() string {
	switch d {
	case ActionNone:
		return "none"
	case ActionOpenShop:
		return "open_shop"
	case ActionCloseDialog:
		return "close_dialog"
	case ActionStartQuest:
		return "start_quest"
	case ActionGiveItem:
		return "give_item"
	default:
		return "unknown"
	}
}

// DialogComponent manages NPC dialog state and available options.
type DialogComponent struct {
	// CurrentDialog is the text currently displayed
	CurrentDialog string

	// Options are the available choices for the player
	Options []DialogOption

	// IsActive indicates if dialog is currently open
	IsActive bool

	// DialogProvider generates dialog content (extensible for complex dialogs)
	Provider DialogProvider
}

// Type returns the component type identifier.
func (d *DialogComponent) Type() string {
	return "dialog"
}

// NewDialogComponent creates a dialog component with a provider.
func NewDialogComponent(provider DialogProvider) *DialogComponent {
	return &DialogComponent{
		CurrentDialog: "",
		Options:       make([]DialogOption, 0),
		IsActive:      false,
		Provider:      provider,
	}
}

// Activate starts the dialog interaction.
func (d *DialogComponent) Activate() {
	if d.Provider != nil {
		d.CurrentDialog, d.Options = d.Provider.GetDialog()
	}
	d.IsActive = true
}

// Deactivate closes the dialog.
func (d *DialogComponent) Deactivate() {
	d.IsActive = false
	d.CurrentDialog = ""
	d.Options = d.Options[:0]
}

// DialogProvider generates dialog content for NPCs.
// This interface enables extensibility for branching dialogs, quest dialogs,
// and dynamic content based on game state.
type DialogProvider interface {
	// GetDialog returns the current dialog text and available options.
	GetDialog() (text string, options []DialogOption)
}

// MerchantDialogProvider implements DialogProvider for merchant NPCs.
// Provides simple buy/sell/leave options.
type MerchantDialogProvider struct {
	MerchantName string
	GreetingText string
}

// NewMerchantDialogProvider creates a dialog provider for merchants.
func NewMerchantDialogProvider(merchantName string) *MerchantDialogProvider {
	return &MerchantDialogProvider{
		MerchantName: merchantName,
		GreetingText: "Welcome! What can I do for you?",
	}
}

// GetDialog returns merchant greeting and shop options.
func (m *MerchantDialogProvider) GetDialog() (string, []DialogOption) {
	text := m.GreetingText

	options := []DialogOption{
		{
			Text:    "Browse your wares",
			Action:  ActionOpenShop,
			Enabled: true,
		},
		{
			Text:    "Never mind",
			Action:  ActionCloseDialog,
			Enabled: true,
		},
	}

	return text, options
}

// TransactionValidator validates commerce transactions.
// This interface enables server-authoritative validation in multiplayer
// and extensibility for reputation systems, barter, and trade quests.
type TransactionValidator interface {
	// CanBuyItem checks if player can purchase an item from merchant.
	// Returns true and empty string if valid, false and error message if invalid.
	CanBuyItem(playerGold, itemPrice int, playerInventoryFull bool) (bool, string)

	// CanSellItem checks if player can sell an item to merchant.
	// Returns true and empty string if valid, false and error message if invalid.
	CanSellItem(merchantGold, itemPrice int, merchantInventoryFull bool) (bool, string)
}

// DefaultTransactionValidator implements basic gold and inventory checks.
type DefaultTransactionValidator struct{}

// NewDefaultTransactionValidator creates the default validator.
func NewDefaultTransactionValidator() *DefaultTransactionValidator {
	return &DefaultTransactionValidator{}
}

// CanBuyItem validates player purchase.
func (v *DefaultTransactionValidator) CanBuyItem(playerGold, itemPrice int, playerInventoryFull bool) (bool, string) {
	if playerGold < itemPrice {
		return false, "Not enough gold"
	}
	if playerInventoryFull {
		return false, "Inventory full"
	}
	return true, ""
}

// CanSellItem validates player sale.
// Note: Merchants always have infinite gold in this implementation.
// Future enhancement: limited merchant gold for economy simulation.
func (v *DefaultTransactionValidator) CanSellItem(merchantGold, itemPrice int, merchantInventoryFull bool) (bool, string) {
	if merchantInventoryFull {
		return false, "Merchant inventory full"
	}
	return true, ""
}
