# Economy Package

**Package**: `github.com/opd-ai/venture/pkg/world/economy`

**Purpose**: Federated marketplace and guild banking system for cross-server item trading and guild resource management.

**Status**: Production-ready (87.3% test coverage)

## Package Structure

This package implements a complete in-game economy with federated marketplace and guild banking:

### Core Files

- **doc.go** (4.7K) - Comprehensive package documentation
- **interfaces.go** (0.4K) - Minimal ECS interfaces (World, Entity)
- **constants.go** (0.7K) - Enum constants (SortCriteria, DeliveryMethod)
- **types.go** (3.2K) - Core data structures (Listing, Transaction, ItemQuery, PriceTrend)

### Implementation Files

- **marketplace.go** (9.7K) - Federated marketplace with cross-server search and transactions
- **pricing_engine.go** (3.3K) - Dynamic pricing based on supply/demand
- **guild_bank.go** (15K) - Guild vaults with gold storage, item storage, and interest
- **system.go** (2.7K) - ECS system wrapper for periodic updates and cleanup

### Documentation

- **AUDIT.md** (8.7K) - Comprehensive quality assessment (zero gaps identified)

## Quick Start

**Note**: Examples use `log.Fatal` and `fmt.Printf` for simplicity. Production code should use `logrus.WithError()` and `logrus.WithFields()` for structured logging per project guidelines.

### Creating the Economy System

```go
import "github.com/opd-ai/venture/pkg/world/economy"

// Create economy system with ECS world
economySystem := economy.NewSystemWithServerID(world, "server-001")

// Update in game loop
economySystem.Update(deltaTime)

// Access marketplace
marketplace := economySystem.GetMarketplace()
```

### Marketplace Operations

```go
// Create a listing
listing := &economy.Listing{
    ListingID:      "listing-123",
    ItemID:         "sword-001",
    ItemName:       "Iron Sword",
    ItemType:       "weapon",
    SellerID:       "player-alice",
    SellerName:     "Alice",
    ServerID:       "server-001",
    Price:          100,
    Quantity:       5,
    MaxStackSize:   1,
    CreatedAt:      time.Now(),
    ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
    DeliveryMethod: economy.DeliveryMail,
    EstimatedHops:  0,
}

err := economySystem.CreateListing(listing)
if err != nil {
    log.Fatalf("Failed to create listing: %v", err)
}

// Search for items
query := economy.ItemQuery{
    ItemType: "weapon",
    MinPrice: 50,
    MaxPrice: 200,
    SortBy:   economy.SortByPrice,
    Limit:    10,
}

listings, err := economySystem.SearchItems(query)
if err != nil {
    log.Fatalf("Search failed: %v", err)
}

for _, l := range listings {
    fmt.Printf("%s: %d gold (server: %s)\n", l.ItemName, l.Price, l.ServerID)
}

// Purchase an item
err = economySystem.PurchaseItem("listing-123", "player-bob", 2)
if err != nil {
    log.Fatalf("Purchase failed: %v", err)
}

// Get price trends
trend := economySystem.GetPriceTrend("weapon")
fmt.Printf("Average weapon price: %d gold\n", trend.AveragePrice)
```

### Guild Banking

```go
// Access guild bank
guildBank := marketplace.GetGuildBank()

// Create guild vault
err := guildBank.CreateVault("guild-001", 0.005) // 0.5% monthly interest
if err != nil {
    log.Fatalf("Failed to create vault: %v", err)
}

// Deposit gold
err = guildBank.DepositGold("guild-001", 10000)
if err != nil {
    log.Fatalf("Deposit failed: %v", err)
}

// Withdraw gold
err = guildBank.WithdrawGold("guild-001", "player-alice", 500, "Buy guild supplies")
if err != nil {
    log.Fatalf("Withdrawal failed: %v", err)
}

// Deposit items
err = guildBank.DepositItem("guild-001", "potion-001", "Health Potion", 20)
if err != nil {
    log.Fatalf("Item deposit failed: %v", err)
}

// Check vault balance
vault, err := guildBank.GetVault("guild-001")
if err != nil {
    log.Fatalf("Failed to get vault: %v", err)
}
fmt.Printf("Vault balance: %d gold, %d items\n", vault.GoldBalance, len(vault.Items))
```

### Dynamic Pricing

```go
// Get pricing engine
pricingEngine := marketplace.GetPricingEngine()

// Suggest price based on market conditions
suggestedPrice := pricingEngine.SuggestPrice("weapon", 100)
fmt.Printf("Suggested price: %d gold (base: 100)\n", suggestedPrice)

// Price adjusts based on:
// - Supply (lower supply = higher price)
// - Demand (higher demand = higher price)  
// - Historical average (dampens volatility)
```

## Architecture

### Federated Marketplace

The marketplace supports cross-server trading:

1. **Seller** creates listing on their local server
2. **Federation** synchronizes listings across servers
3. **Buyer** searches all servers (or specific server)
4. **Transaction** executes with server hop-based fees
5. **Delivery** uses mail (instant) or courier (delayed by distance)

### Transaction Fees

Fees prevent marketplace abuse and create gold sinks:

- **Base fee**: 5% of purchase price
- **Hop fee**: +2% per server hop (for cross-server trades)
- **Maximum**: 15% total fee

Examples:
- Local trade (0 hops): 5% fee
- 1-server hop: 7% fee
- 2-server hops: 9% fee
- 5+ server hops: 15% fee (capped)

```go
// Calculate fee for purchase
basePrice := 1000 * quantity
fee := economy.CalculateTransactionFee(basePrice, 2) // 2 hops
totalCost := basePrice + fee
```

### Guild Banking

Guild vaults provide shared storage:

- **Gold storage** with interest (0.1-1% monthly)
- **Item storage** (up to 5000 items)
- **Daily withdrawal limits** (configurable per vault)
- **Transaction history** (all deposits/withdrawals logged)
- **Interest compounding** (updated on System.Update())

Interest calculation:
```
monthlyRate = 0.005 (0.5%)
dailyRate = monthlyRate / 30
interest = balance * dailyRate
newBalance = balance + interest
```

### Dynamic Pricing

The pricing engine adjusts prices based on market conditions:

```
supplyFactor = 1.0 / (1.0 + listings * 0.1)  // More supply = lower price
demandFactor = 1.0 + (volume * 0.01)         // More demand = higher price
priceFactor = supplyFactor * demandFactor

avgPrice = (currentPrice * 0.7) + (historicalAvg * 0.3)  // 70/30 blend
adjustedPrice = avgPrice * priceFactor

// Clamp to ±10% to prevent extreme swings
finalPrice = clamp(adjustedPrice, basePrice * 0.9, basePrice * 1.1)
```

## Configuration

### Search Parameters

```go
query := economy.ItemQuery{
    ItemType: "weapon",          // Filter by type
    ItemName: "Sword",           // Filter by name (partial match)
    MinPrice: 100,               // Minimum price
    MaxPrice: 1000,              // Maximum price
    ServerID: "server-001",      // Specific server (empty = all servers)
    SellerID: "player-alice",    // Specific seller (empty = all sellers)
    SortBy:   economy.SortByPrice,
    Limit:    50,                // Max results
}
```

### Sort Options

- `SortByPrice` - Price ascending (cheapest first)
- `SortByPriceDesc` - Price descending (most expensive first)
- `SortByQuantity` - Quantity available
- `SortByDeliveryTime` - Fastest delivery first
- `SortByRelevance` - Search relevance (name match quality)

### Delivery Methods

- `DeliveryMail` - Instant delivery via in-game mail system
- `DeliveryCourier` - NPC courier (10-60 minutes based on server distance)

### Listing Expiry

Default expiry: **7 days** from creation

Cleanup runs every **5 minutes** via System.Update()

## Error Handling

All operations return descriptive errors:

```go
// Validation errors
err := guildBank.DepositGold("guild-001", -100)
// Returns: "deposit amount must be positive, got -100"

// Business logic errors
err := guildBank.WithdrawGold("guild-001", "player", 99999)
// Returns: "insufficient funds: vault has 1000 gold, requested 99999"

// Limit enforcement
err := guildBank.WithdrawGold("guild-001", "player", 6000)
// Returns: "daily withdrawal limit exceeded: 4000/5000 gold used, requested 6000"

// Capacity limits
err := guildBank.DepositItem("guild-001", "item", "name", 1)
// Returns: "vault capacity exceeded: 5000 items"
```

## Thread Safety

All operations are thread-safe:

- Marketplace protected by `sync.RWMutex`
- Guild bank protected by `sync.RWMutex`
- Read operations use `RLock()` for concurrency
- Write operations use `Lock()` for exclusivity

Safe for concurrent access from multiple goroutines.

## Performance Characteristics

- **Listing search**: O(n) with early termination on limit
- **Vault lookup**: O(1) by guild ID
- **Price trend**: O(1) map lookup
- **Transaction history**: Append-only (efficient)
- **Memory**: Bounded by listing expiry and vault capacity

## Test Coverage

**Coverage**: 87.3% (exceeds project minimum of 40%)

Run tests:
```bash
go test ./pkg/world/economy/...
go test -cover ./pkg/world/economy/...
go test -race ./pkg/world/economy/...  # Check for race conditions
```

## Dependencies

**Standard Library**:
- `fmt` - Error formatting
- `sort` - Listing sorting
- `strings` - String operations
- `sync` - Thread safety
- `time` - Timestamps

**Internal**: None (fully self-contained)

## Related Packages

- `pkg/world` - World management
- `pkg/world/housing` - Player housing
- `pkg/world/territory` - Territory control
- `pkg/engine` - ECS world and entities

## Integration Example

```go
// In your game setup
economySystem := economy.NewSystemWithServerID(world, serverID)

// Register with ECS
world.AddSystem(economySystem)

// In game loop
economySystem.Update(deltaTime)

// In UI code
listings, _ := economySystem.SearchItems(economy.ItemQuery{
    ItemType: selectedType,
    SortBy:   economy.SortByPrice,
    Limit:    20,
})

// Render listings in UI
for _, listing := range listings {
    ui.ShowListing(listing)
}

// On purchase button click
err := economySystem.PurchaseItem(listingID, playerID, quantity)
if err != nil {
    ui.ShowError(err.Error())
} else {
    ui.ShowSuccess("Purchase complete!")
}
```

## Contributing

When modifying this package:

1. **Maintain test coverage** above 40% (currently 87.3%)
2. **Document all exports** with godoc comments
3. **Add tests** for new functionality (table-driven preferred)
4. **Update AUDIT.md** if adding incomplete features
5. **Preserve thread safety** (use sync.RWMutex)
6. **Run full test suite** before committing

## License

See repository LICENSE file.
