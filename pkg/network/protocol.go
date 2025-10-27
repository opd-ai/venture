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

// TileDamageMessage represents terrain damage notification from client to server.
// Client sends this when player attacks terrain with a weapon.
// Server validates the action and broadcasts terrain state changes.
// Phase 4: Environmental Manipulation System
type TileDamageMessage struct {
	// PlayerID identifies the player performing the action
	PlayerID uint64

	// TileX is the X coordinate of the tile being damaged
	TileX int

	// TileY is the Y coordinate of the tile being damaged
	TileY int

	// Damage is the amount of damage to apply
	Damage float64

	// WeaponID is the entity ID of the weapon used (0 if spell/ability)
	WeaponID uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// TileDestroyedMessage represents terrain destruction notification from server to clients.
// Server broadcasts this when a tile is destroyed (health reaches 0).
// Clients update local terrain state to match.
// Phase 4: Environmental Manipulation System
type TileDestroyedMessage struct {
	// TileX is the X coordinate of the destroyed tile
	TileX int

	// TileY is the Y coordinate of the destroyed tile
	TileY int

	// TimeOfDestruction is the server timestamp when destruction occurred
	TimeOfDestruction float64

	// DestroyedByPlayerID identifies the player that destroyed the tile (0 if environmental)
	DestroyedByPlayerID uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// TileConstructMessage represents construction request from client to server.
// Client sends this when player attempts to build a wall or structure.
// Server validates materials, location, and broadcasts construction start.
// Phase 4: Environmental Manipulation System
type TileConstructMessage struct {
	// PlayerID identifies the player performing construction
	PlayerID uint64

	// TileX is the X coordinate where construction is attempted
	TileX int

	// TileY is the Y coordinate where construction is attempted
	TileY int

	// TileType is the type of tile to construct (wall, door, etc.)
	TileType uint8

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// ConstructionStartedMessage represents construction start notification from server to clients.
// Server broadcasts this after validating TileConstructMessage.
// Clients create buildable entities and start progress tracking.
// Phase 4: Environmental Manipulation System
type ConstructionStartedMessage struct {
	// TileX is the X coordinate of construction
	TileX int

	// TileY is the Y coordinate of construction
	TileY int

	// BuilderPlayerID identifies the player performing construction
	BuilderPlayerID uint64

	// TileType is the type of tile being constructed
	TileType uint8

	// ConstructionTime is the duration in seconds to complete
	ConstructionTime float64

	// TimeStarted is the server timestamp when construction began
	TimeStarted float64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// ConstructionCompletedMessage represents construction completion notification from server to clients.
// Server broadcasts this when construction timer reaches completion.
// Clients place the final tile and remove buildable entities.
// Phase 4: Environmental Manipulation System
type ConstructionCompletedMessage struct {
	// TileX is the X coordinate of completed construction
	TileX int

	// TileY is the Y coordinate of completed construction
	TileY int

	// TileType is the type of tile constructed
	TileType uint8

	// TimeCompleted is the server timestamp when construction finished
	TimeCompleted float64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// FireIgniteMessage represents fire ignition request from client to server.
// Client sends this when player uses fire spell/ability or ignites terrain.
// Server validates and broadcasts fire spread.
// Phase 4: Environmental Manipulation System
type FireIgniteMessage struct {
	// PlayerID identifies the player causing the ignition
	PlayerID uint64

	// TileX is the X coordinate where fire starts
	TileX int

	// TileY is the Y coordinate where fire starts
	TileY int

	// Intensity is the initial fire intensity (0.0-1.0)
	Intensity float64

	// SourceType indicates ignition source ("spell", "explosion", "environmental")
	SourceType string

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// FireSpreadMessage represents fire spread notification from server to clients.
// Server broadcasts this when fire spreads to adjacent tiles.
// Clients create fire components on affected entities.
// Phase 4: Environmental Manipulation System
type FireSpreadMessage struct {
	// TileX is the X coordinate where fire spread
	TileX int

	// TileY is the Y coordinate where fire spread
	TileY int

	// Intensity is the fire intensity at this tile (0.0-1.0)
	Intensity float64

	// Duration is how long the fire will burn (seconds)
	Duration float64

	// TimeIgnited is the server timestamp when fire started
	TimeIgnited float64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// FireExtinguishedMessage represents fire extinguishment notification from server to clients.
// Server broadcasts this when fire burns out or is manually extinguished.
// Clients remove fire components from affected entities.
// Phase 4: Environmental Manipulation System
type FireExtinguishedMessage struct {
	// TileX is the X coordinate where fire was extinguished
	TileX int

	// TileY is the Y coordinate where fire was extinguished
	TileY int

	// TimeExtinguished is the server timestamp when fire ended
	TimeExtinguished float64

	// Reason indicates why fire ended ("burned_out", "extinguished", "tile_destroyed")
	Reason string

	// SequenceNumber for message ordering
	SequenceNumber uint32
}
