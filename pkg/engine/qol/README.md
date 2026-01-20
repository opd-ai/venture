# QoL Package - Quality of Life Features

Package `pkg/engine/qol` provides Quality of Life (QoL) features for enhanced player convenience in the Venture game.

## Package Structure

After reorganization (2026-01-20), the package is organized as follows:

### Implementation Files
- **`autoloot.go`** - AutoLootManager for companion auto-collection
- **`craftqueue.go`** - CraftQueueManager for smart crafting queues
- **`guildinvitation.go`** - GuildInvitationManager for offline invitations
- **`mountwhistle.go`** - MountWhistleManager for vehicle summoning
- **`recipetracker.go`** - RecipeTracker for material tracking
- **`storagesorter.go`** - StorageSorter for inventory organization
- **`system.go`** - Unified Manager coordinating all subsystems
- **`types.go`** - All type definitions, constants, and helpers
- **`doc.go`** - Package-level documentation

### Test Files
- **`manager_test.go`** - Tests for all manager implementations
- **`system_test.go`** - Tests for unified Manager
- **`types_test.go`** - Tests for types and helper functions

## Features

### Auto-Loot System
Companions automatically collect nearby items with configurable preferences:
```go
autoLoot := qol.NewAutoLootManager()
autoLoot.SetRadius(companionID, 8.0)  // 8 tiles radius
if autoLoot.ShouldCollect(companionID, itemRarity, itemType) {
    // Collect item
}
```

### Smart Crafting Queue
Queue multiple recipes for sequential crafting:
```go
craftQueue := qol.NewCraftQueueManager()
craftQueue.AddRecipe(playerID, "iron_sword", 5)  // Queue 5 iron swords
queue := craftQueue.GetQueue(playerID)
```

### Guild Invitations
Offline guild invitation system with expiration:
```go
guildInvites := qol.NewGuildInvitationManager()
guildInvites.SendInvitation(&qol.GuildInvitation{
    InvitationID: "inv123",
    GuildID:      "guild456",
    InviteeID:    "player789",
})
```

### Mount Whistle
Summon vehicles to player location:
```go
mountWhistle := qol.NewMountWhistleManager()
mountWhistle.SummonMount(&qol.MountSummon{
    PlayerID:    playerID,
    VehicleID:   vehicleID,
    TargetPos:   [2]float64{x, y},
    CurrentPos:  [2]float64{vx, vy},
})
```

### Storage Sorting
Organize inventory by multiple criteria:
```go
sorter := qol.NewStorageSorter()
sorter.SortItems(items, qol.SortByRarity)
preset := sorter.GetPreset("value")  // Sort by value preset
```

### Recipe Tracking
Track recipe availability and missing materials:
```go
tracker := qol.NewRecipeTracker()
tracker.TrackRecipe(playerID, &qol.RecipeTrackingInfo{
    RecipeID:     "iron_sword",
    RequiredMats: map[string]int{"iron": 5, "wood": 1},
    AvailableMats: map[string]int{"iron": 3, "wood": 2},
})
```

## Unified Manager

The `Manager` type coordinates all QoL subsystems:
```go
config := qol.Config{
    AutoLoot:     true,
    AutoSort:     true,
    QuickDeposit: true,
}
manager := qol.NewManager(config)

// Access individual subsystems
autoLoot := manager.AutoLoot()
craftQueue := manager.CraftQueue()
guildInvites := manager.GuildInvites()
```

## Thread Safety

All managers use `sync.RWMutex` for concurrent access protection. Safe to use from multiple goroutines.

## Performance

- **Auto-loot**: <1ms per collection cycle
- **Craft queue**: <5ms per recipe processing
- **Storage sort**: <10ms for 100 items
- **Mount summon**: <100ms pathfinding

## Test Coverage

- **Coverage**: 94.2% of statements
- **Tests**: 35 tests (all passing)
- **Benchmarks**: 6 benchmark functions

## Integration Points

- **V4 Companions** - Auto-loot collection behavior
- **V8 Crafting** - Smart queue system
- **V8 Guilds** - Offline invitation acceptance
- **V4 Vehicles** - Mount whistle summoning
- **ECS System** - QoLComponent for entity attachment

## Documentation

For detailed information, see:
- `doc.go` - Comprehensive package documentation
- `AUDIT.md` - Implementation audit and quality metrics
- Go documentation: `go doc github.com/opd-ai/venture/pkg/engine/qol`

## Recent Changes

**2026-01-20 Reorganization:**
- Split monolithic `manager.go` (539 lines) into 6 focused files
- Added file-level documentation for all implementation files
- Improved Item struct documentation
- Created comprehensive AUDIT.md
- Maintained 100% test compatibility (35/35 tests passing)
- Zero regressions, zero build errors
