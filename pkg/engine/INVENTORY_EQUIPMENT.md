# Inventory & Equipment System

**Package:** `github.com/opd-ai/venture/pkg/engine`  
**Status:** ✅ Complete  
**Test Coverage:** 85.1%  
**Phase:** 5.3 - Core Gameplay Systems

---

## Overview

The Inventory & Equipment System provides comprehensive item management for Venture's action-RPG gameplay. It integrates seamlessly with the existing procedural item generation system (Phase 2) and combat system (Phase 5.2), allowing players to collect, manage, equip, and use generated items.

### Key Features

- **Inventory Management**: Capacity-limited storage with weight and item count restrictions
- **Equipment System**: 10 equipment slots for weapons, armor, and accessories
- **Automatic Stat Calculation**: Equipment bonuses automatically update entity combat stats
- **Consumable Usage**: Potions, scrolls, and other consumables with immediate effects
- **Inventory Operations**: Add, remove, transfer, sort, and find items
- **Equipment Swapping**: Seamless swapping between inventory and equipment
- **Gold/Currency Tracking**: Built-in wealth management

---

## Architecture

### Components

#### InventoryComponent

Manages an entity's item collection with capacity constraints.

```go
type InventoryComponent struct {
    Items     []*item.Item  // Items stored in inventory
    MaxItems  int           // Maximum number of items
    MaxWeight float64       // Maximum weight capacity (kg)
    Gold      int           // Currency amount
}
```

**Methods:**
- `GetCurrentWeight() float64` - Calculate total weight of all items
- `CanAddItem(item) bool` - Check if item can be added
- `AddItem(item) bool` - Add item to inventory
- `RemoveItem(index) *item.Item` - Remove item by index
- `RemoveItemByReference(item) bool` - Remove specific item instance
- `FindItem(name) *item.Item` - Search for item by name
- `GetItemCount() int` - Get number of items
- `IsFull() bool` - Check if inventory is full
- `Clear()` - Remove all items

#### EquipmentComponent

Manages equipped items in specific slots with stat calculation.

```go
type EquipmentComponent struct {
    Slots       map[EquipmentSlot]*item.Item  // Equipped items by slot
    CachedStats item.Stats                    // Total bonuses from equipment
    StatsDirty  bool                          // Indicates if stats need recalculation
}
```

**Equipment Slots:**
- `SlotMainHand` - Primary weapon
- `SlotOffHand` - Secondary weapon or shield
- `SlotHead` - Helmet
- `SlotChest` - Body armor
- `SlotLegs` - Leg armor
- `SlotBoots` - Footwear
- `SlotGloves` - Hand armor
- `SlotAccessory1-3` - Rings, amulets, etc.

**Methods:**
- `CanEquip(item, slot) bool` - Check if item fits in slot
- `Equip(item, slot) *item.Item` - Equip item, returns previous item
- `Unequip(slot) *item.Item` - Remove item from slot
- `GetEquipped(slot) *item.Item` - Get item in slot
- `IsEquipped(item) bool` - Check if item is equipped
- `GetSlotForItem(item) (EquipmentSlot, bool)` - Determine appropriate slot
- `RecalculateStats()` - Update cached stat bonuses
- `GetStats() item.Stats` - Get total equipment bonuses
- `GetTotalDefense() int` - Get sum of all armor defense
- `GetWeaponDamage() int` - Get main hand weapon damage
- `GetWeaponSpeed() float64` - Get main hand weapon speed
- `UnequipAll() []*item.Item` - Remove all equipment

### Systems

#### InventorySystem

Manages inventory and equipment operations for entities.

```go
type InventorySystem struct {
    world *World
}
```

**Inventory Operations:**
- `AddItemToInventory(entityID, item) (bool, error)` - Add item to entity's inventory
- `RemoveItemFromInventory(entityID, index) (*item.Item, error)` - Remove item by index
- `GetInventoryValue(entityID) (int, error)` - Calculate total inventory value
- `TransferItem(fromID, toID, index) error` - Move item between entities
- `DropItem(entityID, index) error` - Drop item from inventory

**Equipment Operations:**
- `EquipItem(entityID, inventoryIndex) error` - Equip item from inventory
- `UnequipItem(entityID, slot) error` - Unequip item to inventory
- `UseConsumable(entityID, index) error` - Use consumable item

**Sorting Operations:**
- `SortInventoryByValue(entityID) error` - Sort by value (descending)
- `SortInventoryByWeight(entityID) error` - Sort by weight (ascending)
- `SortInventoryByType(entityID) error` - Sort by item type

---

## Usage Examples

### Basic Inventory Management

```go
// Create entity with inventory
world := engine.NewWorld()
player := world.CreateEntity()
player.AddComponent(engine.NewInventoryComponent(20, 100.0)) // 20 items, 100kg
world.Update(0.0) // Process additions

// Create inventory system
invSystem := engine.NewInventorySystem(world)

// Add items to inventory
sword := &item.Item{Name: "Iron Sword", Type: item.TypeWeapon, /* ... */}
success, err := invSystem.AddItemToInventory(player.ID, sword)
if !success {
    fmt.Println("Inventory full!")
}

// Remove item
removedItem, err := invSystem.RemoveItemFromInventory(player.ID, 0)
```

### Equipment Management

```go
// Add equipment component
player.AddComponent(engine.NewEquipmentComponent())
player.AddComponent(&engine.StatsComponent{})
player.AddComponent(&engine.AttackComponent{})
world.Update(0.0)

// Add weapon to inventory
sword := &item.Item{
    Name: "Steel Sword",
    Type: item.TypeWeapon,
    WeaponType: item.WeaponSword,
    Stats: item.Stats{Damage: 15, AttackSpeed: 1.2},
}
invSystem.AddItemToInventory(player.ID, sword)

// Equip weapon from inventory (index 0)
err := invSystem.EquipItem(player.ID, 0)
if err != nil {
    fmt.Printf("Cannot equip: %v\n", err)
}

// Equipment stats are automatically applied to entity's combat components
```

### Using Consumables

```go
// Add health potion to inventory
potion := &item.Item{
    Name: "Health Potion",
    Type: item.TypeConsumable,
    ConsumableType: item.ConsumablePotion,
    Stats: item.Stats{Value: 50},
}
invSystem.AddItemToInventory(player.ID, potion)

// Use consumable (index 0)
// Automatically applies effects and removes from inventory
err := invSystem.UseConsumable(player.ID, 0)
```

### Inventory Sorting

```go
// Sort inventory by value (most valuable first)
invSystem.SortInventoryByValue(player.ID)

// Sort by weight (lightest first)
invSystem.SortInventoryByWeight(player.ID)

// Sort by type (weapons, armor, consumables, accessories)
invSystem.SortInventoryByType(player.ID)
```

### Item Transfer

```go
// Transfer item from one entity to another
merchant := world.CreateEntity()
merchant.AddComponent(engine.NewInventoryComponent(50, 200.0))
world.Update(0.0)

// Transfer item at index 0 from player to merchant
err := invSystem.TransferItem(player.ID, merchant.ID, 0)
if err != nil {
    fmt.Printf("Transfer failed: %v\n", err)
}
```

---

## Integration with Other Systems

### Combat System Integration

Equipment stats automatically update combat components:

```go
// When equipping a weapon:
// - AttackComponent.Damage is set to weapon damage
// - AttackComponent.Cooldown is set based on weapon speed
// - AttackComponent.DamageType is set based on weapon type

// When equipping armor:
// - StatsComponent.Defense is updated with total armor defense
```

### Item Generation Integration

Works seamlessly with the procedural item generation system:

```go
// Generate items
itemGen := item.NewItemGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,
    Depth: 5,
    GenreID: "fantasy",
    Custom: map[string]interface{}{"count": 10},
}

result, _ := itemGen.Generate(seed, params)
items := result.([]*item.Item)

// Add generated items to inventory
for _, itm := range items {
    invSystem.AddItemToInventory(player.ID, itm)
}
```

---

## Design Decisions

### Capacity Limits

Two types of capacity limits are enforced:
1. **Item Count**: Maximum number of items (prevents UI clutter)
2. **Weight**: Maximum weight capacity (adds strategic resource management)

Both limits must be satisfied to add an item.

### Equipment Swapping

When equipping an item in an occupied slot:
1. New item is equipped
2. Previous item is automatically returned to inventory
3. If inventory is full, operation fails and original item stays equipped

### Stat Calculation

Equipment stats are cached and only recalculated when:
- Items are equipped or unequipped
- Manual recalculation is requested

This optimization prevents redundant calculations during gameplay.

### Consumable Effects

Consumable effects are applied immediately based on type:
- **Potions**: Restore health (value / 10)
- **Food**: Restore health over time (value / 20)
- **Scrolls**: Placeholder for spell effects
- **Bombs**: Placeholder for area damage

---

## Testing

### Test Coverage: 85.1%

**Component Tests (19 tests):**
- Inventory capacity management (weight & item limits)
- Add/remove operations
- Item finding and clearing
- Equipment slot validation
- Equipment/unequip operations
- Stat calculation from equipment
- Slot detection for items

**System Tests (16 tests):**
- Add/remove from inventory
- Equipment with swapping
- Unequip with full inventory
- Consumable usage
- Item transfer between entities
- Inventory sorting (value, weight, type)
- Error handling for invalid operations

### Running Tests

```bash
# Run inventory tests
go test -tags test ./pkg/engine -run "TestInventory|TestEquipment" -v

# Run with coverage
go test -tags test ./pkg/engine -cover

# Run benchmarks
go test -tags test ./pkg/engine -bench=. -benchmem
```

---

## Demo Tool

### inventorytest

A CLI tool to demonstrate inventory and equipment functionality.

**Location:** `cmd/inventorytest/main.go`

**Usage:**
```bash
go run ./cmd/inventorytest -seed 12345 -count 10 -depth 5

# Options:
#   -seed   int64   Generation seed (default: 12345)
#   -count  int     Number of items to generate (default: 10)
#   -depth  int     Dungeon depth for scaling (default: 5)
```

**Features:**
- Generates procedural items
- Creates player with inventory and equipment
- Adds items to inventory
- Automatically equips items
- Uses consumables
- Sorts inventory
- Shows equipment stats

---

## Performance Considerations

### Memory Usage

- Inventory items are stored as pointers (minimal overhead)
- Equipment slots use a map (O(1) access)
- Cached stats prevent repeated calculations

### Typical Memory Footprint

- `InventoryComponent`: ~200 bytes + (pointers to items)
- `EquipmentComponent`: ~300 bytes + (pointers to items)
- Per-entity overhead: ~500 bytes

### Optimization Tips

1. **Pre-allocate inventory capacity** during entity creation
2. **Batch inventory operations** when possible
3. **Cache equipment stats** (already implemented)
4. **Limit inventory size** to reasonable values (20-50 items)

---

## Future Enhancements

### Potential Additions

1. **Stack Management**: Stack consumables (e.g., 99 potions per slot)
2. **Quick Slots**: Hotbar for fast consumable access
3. **Item Sets**: Bonuses for wearing matching equipment
4. **Durability System**: Item degradation over time
5. **Enchantments**: Additional item properties beyond stats
6. **Trading System**: Enhanced transfer with gold exchange
7. **Container Items**: Bags that increase inventory capacity
8. **Item Comparison**: UI helpers for comparing equipment

### Network Synchronization

For Phase 6 (Multiplayer), inventory operations will need:
- Reliable state synchronization
- Conflict resolution for simultaneous modifications
- Efficient delta updates for inventory changes
- Server-side validation for anti-cheat

---

## API Reference

### Type Constants

```go
type EquipmentSlot int

const (
    SlotMainHand    EquipmentSlot = iota
    SlotOffHand
    SlotHead
    SlotChest
    SlotLegs
    SlotBoots
    SlotGloves
    SlotAccessory1
    SlotAccessory2
    SlotAccessory3
)
```

### Component Constructors

```go
func NewInventoryComponent(maxItems int, maxWeight float64) *InventoryComponent
func NewEquipmentComponent() *EquipmentComponent
```

### System Constructor

```go
func NewInventorySystem(world *World) *InventorySystem
```

---

## Related Documentation

- [Item Generation System](../procgen/item/README.md)
- [Combat System](COMBAT_SYSTEM.md)
- [Movement & Collision](MOVEMENT_COLLISION.md)
- [ECS Architecture](doc.go)

---

## Changelog

### v1.0.0 (Current)
- Initial implementation
- Inventory management with capacity limits
- Equipment system with 10 slots
- Consumable usage
- Inventory sorting utilities
- Integration with combat system
- Comprehensive test suite (85.1% coverage)
- CLI demo tool

---

**Last Updated:** October 21, 2025  
**Author:** Venture Development Team  
**Phase:** 5.3 - Core Gameplay Systems
