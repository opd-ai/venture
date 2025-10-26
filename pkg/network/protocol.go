// Package network provides network protocol data structures.
// This file defines core protocol types for state updates, input commands,
// and network messages used in client-server communication.
package network

// ComponentData represents serialized component data for network transmission.
type ComponentData struct {
	Type string
	Data []byte
}

// StateUpdate represents a network packet containing entity state changes.
type StateUpdate struct {
	// Timestamp of when this update was created (server time)
	Timestamp uint64

	// EntityID identifies which entity this update is for
	EntityID uint64

	// Components contains the updated component data
	Components []ComponentData

	// Priority determines update ordering (higher = more important)
	// 0 = low priority, 255 = critical
	Priority uint8

	// SequenceNumber for ordering and detecting packet loss
	SequenceNumber uint32
}

// InputCommand represents a player input sent from client to server.
type InputCommand struct {
	// PlayerID identifies which player sent this input
	PlayerID uint64

	// Timestamp when the input was generated (client time)
	Timestamp uint64

	// SequenceNumber for input ordering
	SequenceNumber uint32

	// InputType identifies the type of input (move, attack, use item, etc.)
	InputType string

	// Data contains the input-specific data (serialized)
	Data []byte
}

// ConnectionInfo contains information about a network connection.
type ConnectionInfo struct {
	// PlayerID uniquely identifies the player
	PlayerID uint64

	// Address is the network address (IP:port)
	Address string

	// Latency is the round-trip time in milliseconds
	Latency float64

	// Connected indicates if the connection is active
	Connected bool
}

// DeathMessage represents entity death notification from server to clients.
// Server broadcasts this message when an entity dies to synchronize death state.
// Category 1.1: Death & Revival System
type DeathMessage struct {
	// EntityID identifies the entity that died
	EntityID uint64

	// TimeOfDeath is the server timestamp when death occurred
	TimeOfDeath float64

	// KillerID identifies the entity that caused the death (0 if environmental)
	KillerID uint64

	// DroppedItemIDs contains entity IDs of items spawned from death
	DroppedItemIDs []uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// RevivalMessage represents player revival notification from server to clients.
// Server broadcasts this message when a player is revived by a teammate.
// Category 1.1: Death & Revival System
type RevivalMessage struct {
	// EntityID identifies the entity being revived
	EntityID uint64

	// ReviverID identifies the entity that performed the revival
	ReviverID uint64

	// TimeOfRevival is the server timestamp when revival occurred
	TimeOfRevival float64

	// RestoredHealth is the health amount restored (as fraction of max)
	RestoredHealth float64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// OpenShopMessage represents a client request to open a merchant shop.
// Client sends this when player interacts with a merchant NPC.
// Server responds with ShopInventoryMessage containing merchant's stock.
// Phase 3: Commerce & NPC Interaction System
type OpenShopMessage struct {
	// PlayerID identifies the player opening the shop
	PlayerID uint64

	// MerchantID identifies the merchant entity being interacted with
	MerchantID uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// ShopInventoryMessage represents server response with merchant inventory.
// Server sends this in response to OpenShopMessage to show merchant's stock.
// Phase 3: Commerce & NPC Interaction System
type ShopInventoryMessage struct {
	// MerchantID identifies the merchant entity
	MerchantID uint64

	// MerchantName is the display name of the merchant
	MerchantName string

	// PriceMultiplier affects buy prices (1.5 = 50% markup)
	PriceMultiplier float64

	// BuyBackPercentage is the percentage paid when buying from player
	BuyBackPercentage float64

	// ItemIDs contains the entity IDs of items in merchant inventory
	ItemIDs []uint64

	// ItemPrices contains the sell prices for each item (parallel to ItemIDs)
	ItemPrices []int

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// BuyItemMessage represents a client request to purchase an item from merchant.
// Client sends this when player confirms purchase in shop UI.
// Server validates gold, transfers item, and responds with TransactionResultMessage.
// Phase 3: Commerce & NPC Interaction System
type BuyItemMessage struct {
	// PlayerID identifies the player making the purchase
	PlayerID uint64

	// MerchantID identifies the merchant entity
	MerchantID uint64

	// ItemIndex is the index in the merchant's inventory (0-based)
	ItemIndex int

	// ExpectedPrice is the price the client expects to pay (validation)
	ExpectedPrice int

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// SellItemMessage represents a client request to sell an item to merchant.
// Client sends this when player confirms sale in shop UI.
// Server validates item, calculates price, and responds with TransactionResultMessage.
// Phase 3: Commerce & NPC Interaction System
type SellItemMessage struct {
	// PlayerID identifies the player making the sale
	PlayerID uint64

	// MerchantID identifies the merchant entity
	MerchantID uint64

	// ItemIndex is the index in the player's inventory (0-based)
	ItemIndex int

	// ExpectedPrice is the price the client expects to receive (validation)
	ExpectedPrice int

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// TransactionResultMessage represents server response to buy/sell transaction.
// Server sends this after processing BuyItemMessage or SellItemMessage.
// Contains transaction outcome (success/failure) and updated state.
// Phase 3: Commerce & NPC Interaction System
type TransactionResultMessage struct {
	// PlayerID identifies the player involved in transaction
	PlayerID uint64

	// MerchantID identifies the merchant entity
	MerchantID uint64

	// Success indicates if transaction completed successfully
	Success bool

	// ErrorMessage contains failure reason if Success is false
	ErrorMessage string

	// TransactionType indicates "buy" or "sell"
	TransactionType string

	// ItemID is the entity ID of the item that was transacted
	ItemID uint64

	// GoldAmount is the gold spent (negative) or received (positive)
	GoldAmount int

	// UpdatedPlayerGold is the player's gold after transaction
	UpdatedPlayerGold int

	// UpdatedInventory indicates if client should refresh inventory
	UpdatedInventory bool

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// CloseShopMessage represents a client notification that shop UI was closed.
// Client sends this when player exits shop (optional, for server tracking).
// Server can use this to clean up shop state or log analytics.
// Phase 3: Commerce & NPC Interaction System
type CloseShopMessage struct {
	// PlayerID identifies the player closing the shop
	PlayerID uint64

	// MerchantID identifies the merchant entity
	MerchantID uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}
