# Merchant NPC Generation System

**Status**: ✅ Complete  
**Package**: `pkg/procgen/entity`  
**Test Coverage**: 92.0%  
**Date**: October 26, 2025

## Overview

The merchant NPC generation system extends the entity generator with specialized functionality for creating merchant NPCs with procedurally generated inventories. Merchants support both fixed (stationary shopkeepers in settlements) and nomadic (wandering traders) behavior patterns.

## Files

- **`pkg/procgen/entity/merchant.go`** (282 lines)
  - `GenerateMerchant()` - Main merchant generation function
  - `MerchantData` - Merchant-specific data structure
  - `GenerateMerchantSpawnPoints()` - Deterministic spawn location generation
  - `MerchantNameTemplates` - Genre-specific merchant names

- **`pkg/procgen/entity/merchant_test.go`** (430+ lines)
  - 9 comprehensive test functions
  - 2 benchmark functions
  - Tests cover all genres, merchant types, determinism, pricing, stats

- **`cmd/merchanttest/main.go`** (120 lines)
  - CLI tool for testing merchant generation
  - Supports all genres and merchant types
  - Verbose mode shows full inventory

- **`pkg/procgen/entity/types.go`** (Updated)
  - Added NPC templates for scifi, horror, cyberpunk, postapoc genres
  - Each genre now has merchant-appropriate names and stats

## Key Features

### 1. Merchant Types

**Fixed Merchants**:
- Stationary shopkeepers in settlements
- Lower price markup (1.5x base price)
- Spawn at deterministic safe locations
- Restock inventory periodically

**Nomadic Merchants**:
- Wandering traders
- Higher price markup (1.8x base price)
- Random spawn locations
- Represent traveling convenience

### 2. Genre Support

All 5 genres supported with theme-appropriate merchant names:

- **Fantasy**: "Aldric the Trader", "Mirena's Goods", "Thorn's Supplies"
- **Sci-Fi**: "Tech Trader Station", "Quantum Supplies", "Nexus Exchange"
- **Horror**: "The Bone Trader", "Cursed Goods", "Shadow Market"
- **Cyberpunk**: "Chrome Exchange", "Neural Market", "Augment Shop"
- **Post-Apocalyptic**: "Wasteland Trader", "Salvage Market", "Scrap Exchange"

### 3. Inventory Generation

Merchants stock 15-24 procedurally generated items with genre-appropriate types:

- **60% Consumables**: Potions, scrolls, food, ammunition
- **30% Equipment**: Weapons, armor, accessories
- **10% Rare Items**: Higher quality, any type

Inventory uses existing `item.Generator` with merchant-specific parameters.

### 4. Pricing System

- **Price Multiplier**: Controls markup on sold items (default 1.5x-1.8x)
- **Buyback Percentage**: How much merchants pay for player items (default 50%)
- **Rarity Scaling**: Higher rarity items have proportionally higher prices

### 5. Deterministic Generation

- Same seed produces identical merchants (name, stats, inventory)
- Spawn points are deterministic based on world seed
- Critical for multiplayer synchronization

## API Usage

### Generating a Merchant

```go
import (
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/entity"
)

gen := entity.NewEntityGenerator()

params := procgen.GenerationParams{
    GenreID:    "fantasy",
    Difficulty: 0.5,
    Depth:      1,
}

merchant, err := gen.GenerateMerchant(12345, params, entity.MerchantFixed)
if err != nil {
    // Handle error
}

// Access merchant data
name := merchant.Entity.Name
inventory := merchant.Inventory
priceMultiplier := merchant.PriceMultiplier
```

### Generating Spawn Points

```go
worldSeed := int64(67890)
worldWidth, worldHeight := 1000, 800
merchantCount := 3

spawnPoints := entity.GenerateMerchantSpawnPoints(
    worldSeed,
    worldWidth,
    worldHeight,
    entity.MerchantFixed,
    merchantCount,
)

// Use spawn points to place merchants
for i, pt := range spawnPoints {
    // Place merchant[i] at (pt.X, pt.Y)
}
```

## Integration with Commerce System

The `MerchantData` structure is designed to be converted into engine components:

```go
// In your world/entity spawning code:
merchantData, _ := entityGen.GenerateMerchant(seed, params, merchantType)

// Create ECS entity
merchantEntity := world.CreateEntity()

// Add merchant component (from pkg/engine/commerce_components.go)
merchantComp := &engine.MerchantComponent{
    Inventory:         merchantData.Inventory,
    MaxInventory:      len(merchantData.Inventory),
    MerchantType:      engine.MerchantType(merchantData.MerchantType),
    PriceMultiplier:   merchantData.PriceMultiplier,
    BuyBackPercentage: merchantData.BuyBackPercentage,
    MerchantName:      merchantData.Entity.Name,
}
merchantEntity.AddComponent(merchantComp)

// Add position component
positionComp := &engine.PositionComponent{
    X: merchantData.SpawnX,
    Y: merchantData.SpawnY,
}
merchantEntity.AddComponent(positionComp)

// Add other components (sprite, stats, etc.)
```

## CLI Testing Tool

The `merchanttest` utility provides quick testing without running the full game:

```bash
# Generate fantasy merchants
./merchanttest -genre fantasy -count 3 -type fixed -verbose

# Generate cyberpunk nomadic merchants
./merchanttest -genre cyberpunk -count 2 -type nomadic

# Custom seed and depth
./merchanttest -seed 99999 -depth 5 -genre scifi
```

**Flags**:
- `-genre`: fantasy, scifi, horror, cyberpunk, postapoc (default: fantasy)
- `-count`: Number of merchants to generate (default: 3)
- `-type`: fixed or nomadic (default: fixed)
- `-depth`: Affects inventory quality (default: 1)
- `-difficulty`: 0.0-1.0 multiplier (default: 0.5)
- `-seed`: Generation seed (default: 12345)
- `-verbose`: Show full inventory listings

## Testing

Run merchant tests:
```bash
go test ./pkg/procgen/entity -run TestGenerateMerchant
```

Run with coverage:
```bash
go test -cover ./pkg/procgen/entity
# Output: 92.0% coverage
```

Run benchmarks:
```bash
go test -bench=Merchant ./pkg/procgen/entity
# BenchmarkGenerateMerchant: ~200µs per merchant
# BenchmarkGenerateMerchantSpawnPoints: ~9µs for 5 spawn points
```

## Performance

- **Merchant Generation**: ~200 microseconds per merchant (including inventory)
- **Spawn Point Generation**: ~9 microseconds for 5 points
- **Memory**: ~5KB per merchant with 20-item inventory
- **Determinism**: 100% reproducible with same seed

## Next Steps

The merchant generation system is complete and ready for integration. Remaining Phase 3 tasks:

1. **Network Protocol Support**: Add commerce message types to `pkg/network/protocol.go`
2. **Client Integration**: Wire up merchant spawning, proximity detection, and shop UI in `cmd/client/main.go`

See `docs/PLAN.md` for detailed integration requirements.

## Design Decisions

### Why Separate MerchantData from Engine Components?

**Rationale**: Procgen packages should be engine-agnostic. `MerchantData` is pure generation output that can be converted to engine components at integration time. This maintains clean separation between procedural generation (data) and game logic (behavior).

### Why Two Merchant Types?

**Rationale**: Provides gameplay variety. Fixed merchants offer stable shops players can return to. Nomadic merchants add exploration incentive and create "lucky find" moments. Price difference balances convenience vs. cost.

### Why Genre-Specific Names?

**Rationale**: Merchant names significantly impact immersion. "Aldric the Trader" fits fantasy but breaks cyberpunk atmosphere. Genre-appropriate names cost minimal implementation but greatly enhance thematic consistency.

### Why 60/30/10 Inventory Distribution?

**Rationale**: Merchants need consumables (gameplay staples) more than equipment (less frequently needed). 10% rare items provide excitement without overwhelming common stock. Distribution mirrors typical RPG shop inventories.

## References

- **Commerce System**: `pkg/engine/commerce_components.go`, `pkg/engine/commerce_system.go`
- **Dialog System**: `pkg/engine/dialog_system.go`
- **Shop UI**: `pkg/engine/shop_ui.go`
- **Item Generator**: `pkg/procgen/item/generator.go`
- **Entity Generator**: `pkg/procgen/entity/generator.go`
- **PLAN.md**: Phase 3 implementation tracking
