# Commerce Network Protocol

**Status**: ✅ Complete  
**Package**: `pkg/network`  
**Test Coverage**: 100% (protocol structures)  
**Date**: October 26, 2025

## Overview

The commerce network protocol extends Venture's client-server communication with message types for shop transactions, enabling server-authoritative commerce in multiplayer. The protocol supports shop opening, inventory browsing, buy/sell transactions, and transaction validation.

## Message Types

### 1. OpenShopMessage (Client → Server)

Sent when a player interacts with a merchant NPC to open the shop interface.

```go
type OpenShopMessage struct {
    PlayerID       uint64  // Player requesting to open shop
    MerchantID     uint64  // Target merchant entity
    SequenceNumber uint32  // Message ordering
}
```

**Purpose**: Initiates shop interaction  
**Response**: ShopInventoryMessage  
**Validation**: Server checks proximity, merchant exists, merchant not busy

**Example Flow**:
```go
// Client: Player presses S key near merchant
openMsg := OpenShopMessage{
    PlayerID:       player.ID,
    MerchantID:     nearestMerchant.ID,
    SequenceNumber: nextSeq(),
}
client.Send(openMsg)
```

### 2. ShopInventoryMessage (Server → Client)

Server response containing merchant's inventory, pricing, and metadata.

```go
type ShopInventoryMessage struct {
    MerchantID        uint64    // Merchant entity ID
    MerchantName      string    // Display name
    PriceMultiplier   float64   // Markup (1.5 = 50% markup)
    BuyBackPercentage float64   // Buyback rate (0.5 = 50%)
    ItemIDs           []uint64  // Item entity IDs
    ItemPrices        []int     // Sell prices (parallel to ItemIDs)
    SequenceNumber    uint32    // Message ordering
}
```

**Purpose**: Provides shop data to client UI  
**Triggered By**: OpenShopMessage  
**Constraints**: `len(ItemIDs) == len(ItemPrices)`

**Example Flow**:
```go
// Server: Respond to shop open request
inventoryMsg := ShopInventoryMessage{
    MerchantID:        merchant.ID,
    MerchantName:      "Aldric the Trader",
    PriceMultiplier:   1.5,
    BuyBackPercentage: 0.5,
    ItemIDs:           []uint64{1001, 1002, 1003},
    ItemPrices:        []int{100, 200, 300},
    SequenceNumber:    nextSeq(),
}
server.SendTo(playerID, inventoryMsg)
```

### 3. BuyItemMessage (Client → Server)

Request to purchase an item from merchant inventory.

```go
type BuyItemMessage struct {
    PlayerID      uint64  // Player making purchase
    MerchantID    uint64  // Merchant entity
    ItemIndex     int     // Index in merchant inventory (0-based)
    ExpectedPrice int     // Price client expects to pay (validation)
    SequenceNumber uint32 // Message ordering
}
```

**Purpose**: Purchase transaction request  
**Response**: TransactionResultMessage  
**Validation**: 
- Player has sufficient gold
- Item index in range
- Price matches (prevents race conditions)
- Inventory has space

**Example Flow**:
```go
// Client: Player confirms purchase
buyMsg := BuyItemMessage{
    PlayerID:       player.ID,
    MerchantID:     merchant.ID,
    ItemIndex:      selectedIndex,
    ExpectedPrice:  displayedPrice,
    SequenceNumber: nextSeq(),
}
client.Send(buyMsg)
```

### 4. SellItemMessage (Client → Server)

Request to sell an item from player inventory to merchant.

```go
type SellItemMessage struct {
    PlayerID      uint64  // Player making sale
    MerchantID    uint64  // Merchant entity
    ItemIndex     int     // Index in player inventory (0-based)
    ExpectedPrice int     // Price client expects to receive (validation)
    SequenceNumber uint32 // Message ordering
}
```

**Purpose**: Sale transaction request  
**Response**: TransactionResultMessage  
**Validation**:
- Item exists in player inventory
- Item is sellable (not quest item, etc.)
- Price matches expected buyback value
- Merchant has gold (optional check)

**Example Flow**:
```go
// Client: Player confirms sale
sellMsg := SellItemMessage{
    PlayerID:       player.ID,
    MerchantID:     merchant.ID,
    ItemIndex:      selectedInventorySlot,
    ExpectedPrice:  calculatedBuyback,
    SequenceNumber: nextSeq(),
}
client.Send(sellMsg)
```

### 5. TransactionResultMessage (Server → Client)

Server response indicating transaction outcome (success or failure).

```go
type TransactionResultMessage struct {
    PlayerID          uint64  // Player involved
    MerchantID        uint64  // Merchant entity
    Success           bool    // Transaction succeeded
    ErrorMessage      string  // Failure reason (empty if success)
    TransactionType   string  // "buy" or "sell"
    ItemID            uint64  // Item entity ID (0 if failed)
    GoldAmount        int     // Gold change (negative = spent, positive = earned)
    UpdatedPlayerGold int     // New gold total
    UpdatedInventory  bool    // Client should refresh inventory
    SequenceNumber    uint32  // Message ordering
}
```

**Purpose**: Transaction confirmation or rejection  
**Triggered By**: BuyItemMessage or SellItemMessage  
**Success Cases**: Item transferred, gold updated, inventory modified  
**Failure Cases**: Insufficient gold, invalid item, price mismatch, inventory full

**Example Flow (Success)**:
```go
// Server: Process successful buy transaction
resultMsg := TransactionResultMessage{
    PlayerID:          player.ID,
    MerchantID:        merchant.ID,
    Success:           true,
    ErrorMessage:      "",
    TransactionType:   "buy",
    ItemID:            purchasedItem.ID,
    GoldAmount:        -100,
    UpdatedPlayerGold: 900,
    UpdatedInventory:  true,
    SequenceNumber:    nextSeq(),
}
server.SendTo(playerID, resultMsg)
```

**Example Flow (Failure)**:
```go
// Server: Reject transaction due to insufficient gold
resultMsg := TransactionResultMessage{
    PlayerID:          player.ID,
    MerchantID:        merchant.ID,
    Success:           false,
    ErrorMessage:      "insufficient gold",
    TransactionType:   "buy",
    ItemID:            0,
    GoldAmount:        0,
    UpdatedPlayerGold: player.Gold,
    UpdatedInventory:  false,
    SequenceNumber:    nextSeq(),
}
server.SendTo(playerID, resultMsg)
```

### 6. CloseShopMessage (Client → Server)

Optional notification that player closed shop UI.

```go
type CloseShopMessage struct {
    PlayerID       uint64  // Player closing shop
    MerchantID     uint64  // Merchant entity
    SequenceNumber uint32  // Message ordering
}
```

**Purpose**: Cleanup notification (optional)  
**Use Cases**: 
- Server analytics tracking
- Release merchant "busy" state
- Cancel pending transactions

**Example Flow**:
```go
// Client: Player presses S or ESC to close shop
closeMsg := CloseShopMessage{
    PlayerID:       player.ID,
    MerchantID:     merchant.ID,
    SequenceNumber: nextSeq(),
}
client.Send(closeMsg)
```

## Complete Transaction Flow

### Successful Purchase Flow

```
Client                          Server
  |                                |
  |--- OpenShopMessage ----------->|
  |                                | Validate proximity, merchant
  |<-- ShopInventoryMessage -------|
  |                                |
  | (Player browses, selects item) |
  |                                |
  |--- BuyItemMessage ------------>|
  |                                | Validate gold, space, price
  |                                | Deduct gold, transfer item
  |<-- TransactionResultMessage ---|
  | (Success=true, UpdatedGold)    |
  |                                |
  |--- CloseShopMessage ---------->|
  |                                |
```

### Failed Transaction Flow

```
Client                          Server
  |                                |
  |--- BuyItemMessage ------------>|
  |                                | Check gold: INSUFFICIENT
  |<-- TransactionResultMessage ---|
  | (Success=false, ErrorMessage)  |
  | Display error to player        |
  |                                |
```

### Sell Transaction Flow

```
Client                          Server
  |                                |
  |--- SellItemMessage ----------->|
  |                                | Validate item, calculate price
  |                                | Transfer item, add gold
  |<-- TransactionResultMessage ---|
  | (Success=true, GoldAmount > 0) |
  | Update UI                      |
  |                                |
```

## Error Handling

### Client-Side Validation (Pre-Send)

Before sending transaction messages, client should validate:
- Player has sufficient gold (buy)
- Player has item in inventory (sell)
- Item index is valid
- Price hasn't changed (race condition check)

### Server-Side Validation (Authoritative)

Server MUST validate all transactions:
1. **Proximity Check**: Player near merchant (<32 pixels)
2. **State Check**: Merchant exists, not busy, shop open
3. **Gold Check**: Player has gold >= item price (buy)
4. **Inventory Check**: Player has space (buy), item exists (sell)
5. **Price Validation**: ExpectedPrice matches calculated price
6. **Atomic Transaction**: All-or-nothing (rollback on failure)

### Common Error Codes

| Error Message | Cause | Client Action |
|--------------|-------|---------------|
| "insufficient gold" | Player gold < item price | Display error, stay in shop |
| "inventory full" | No empty slots | Display error, suggest selling |
| "item not found" | Invalid item index | Refresh inventory, retry |
| "price changed" | Price mismatch | Refresh shop inventory |
| "merchant busy" | Another player transacting | Display "Please wait", retry |
| "out of range" | Player too far | Close shop UI |

## Security Considerations

### Price Validation

Client sends `ExpectedPrice` to detect race conditions:
```go
// Client calculates expected price
expectedPrice := basePrice * merchant.PriceMultiplier

// Server validates
if msg.ExpectedPrice != actualPrice {
    return TransactionResultMessage{
        Success: false,
        ErrorMessage: "price changed",
    }
}
```

**Rationale**: Prevents exploits where client assumes old price during slow network conditions.

### Server Authority

All gold and inventory modifications occur **server-side only**:
- Client UI shows predicted state (client-side prediction)
- Server validates and sends authoritative TransactionResultMessage
- Client reconciles prediction with server result

**Rationale**: Prevents cheating via client modification.

### Sequence Numbers

All messages include SequenceNumber for:
- Detecting duplicate messages (replay attacks)
- Ordering out-of-order UDP packets
- Correlating request/response pairs

**Implementation**:
```go
// Server tracks last processed sequence per player
if msg.SequenceNumber <= player.LastSeq {
    return // Ignore duplicate
}
player.LastSeq = msg.SequenceNumber
```

## Integration Guide

### Client Implementation

```go
// In client game loop
func (g *Game) Update() error {
    // Check for merchant interaction
    if inpututil.IsKeyJustPressed(ebiten.KeyS) {
        merchant := g.findNearestMerchant()
        if merchant != nil {
            g.openShop(merchant)
        }
    }
    
    // Handle shop UI input
    if g.shopUI.IsVisible() {
        g.shopUI.Update()
    }
    
    // Process network messages
    for _, msg := range g.network.ReceiveMessages() {
        switch m := msg.(type) {
        case ShopInventoryMessage:
            g.shopUI.LoadInventory(m)
        case TransactionResultMessage:
            g.handleTransactionResult(m)
        }
    }
}

func (g *Game) openShop(merchant *Entity) {
    msg := OpenShopMessage{
        PlayerID:   g.playerID,
        MerchantID: merchant.ID,
        SequenceNumber: g.nextSeq(),
    }
    g.network.Send(msg)
}
```

### Server Implementation

```go
// In server message handler
func (s *Server) HandleMessage(playerID uint64, msg interface{}) {
    switch m := msg.(type) {
    case OpenShopMessage:
        s.handleOpenShop(playerID, m)
    case BuyItemMessage:
        s.handleBuyItem(playerID, m)
    case SellItemMessage:
        s.handleSellItem(playerID, m)
    case CloseShopMessage:
        s.handleCloseShop(playerID, m)
    }
}

func (s *Server) handleBuyItem(playerID uint64, msg BuyItemMessage) {
    player := s.getPlayer(playerID)
    merchant := s.getMerchant(msg.MerchantID)
    
    // Validate transaction
    result := s.commerceSystem.BuyItem(playerID, msg.MerchantID, msg.ItemIndex)
    
    // Send result
    resultMsg := TransactionResultMessage{
        PlayerID:          playerID,
        MerchantID:        msg.MerchantID,
        Success:           result.Success,
        ErrorMessage:      result.ErrorMessage,
        TransactionType:   "buy",
        ItemID:            result.ItemID,
        GoldAmount:        result.GoldAmount,
        UpdatedPlayerGold: player.Gold,
        UpdatedInventory:  result.Success,
        SequenceNumber:    s.nextSeq(),
    }
    s.SendTo(playerID, resultMsg)
}
```

## Testing

### Unit Tests

Run protocol structure tests:
```bash
go test ./pkg/network -run TestOpenShopMessage
go test ./pkg/network -run TestBuyItemMessage
go test ./pkg/network -run TestSellItemMessage
go test ./pkg/network -run TestTransactionResultMessage
go test ./pkg/network -run TestCloseShopMessage
go test ./pkg/network -run TestShopInventoryMessage
```

### Integration Tests

Test complete workflow:
```bash
go test ./pkg/network -run TestCommerceProtocolWorkflow
go test ./pkg/network -run TestCommerceProtocolFailureScenarios
```

### Coverage

```bash
go test -cover ./pkg/network
# Result: 57.1% overall (100% for protocol.go structures)
```

## Performance Considerations

### Message Sizes

Estimated sizes for typical transactions:
- OpenShopMessage: 24 bytes
- ShopInventoryMessage: ~200-500 bytes (20 items)
- BuyItemMessage: 28 bytes
- SellItemMessage: 28 bytes
- TransactionResultMessage: 64 bytes
- CloseShopMessage: 24 bytes

### Bandwidth Usage

Typical shop interaction:
1. Open: 24 bytes (C→S) + 300 bytes (S→C) = 324 bytes
2. Buy: 28 bytes (C→S) + 64 bytes (S→C) = 92 bytes
3. Close: 24 bytes (C→S) = 24 bytes

**Total**: ~440 bytes per complete transaction

### Optimization Strategies

1. **Batch Updates**: Send multiple item updates in single ShopInventoryMessage
2. **Delta Compression**: Only send changed items on restock
3. **Caching**: Client caches merchant inventory, server sends "inventory unchanged" flag
4. **Lazy Loading**: Request full inventory only when opening shop, not on proximity

## Future Enhancements

Potential protocol extensions:
- **Bartering**: Add item-for-item trades
- **Bulk Transactions**: Buy/sell multiple items at once
- **Auction System**: Timed auctions with multiple bidders
- **Reputation**: Pricing based on player faction standing
- **Merchant AI**: Dynamic pricing based on supply/demand

## References

- **Commerce System**: `pkg/engine/commerce_system.go`, `pkg/engine/commerce_components.go`
- **Shop UI**: `pkg/engine/shop_ui.go`
- **Dialog System**: `pkg/engine/dialog_system.go`
- **Merchant Generation**: `pkg/procgen/entity/merchant.go`
- **PLAN.md**: Phase 3 tracking
