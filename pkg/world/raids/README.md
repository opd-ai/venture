# Package: raids

## Overview

The `raids` package provides complete procedural raid dungeon generation and instance management for Venture's endgame multiplayer content. Raids are designed for groups of 5-10 players and feature multi-boss encounters with procedurally generated mechanics, scaling difficulty tiers, and epic loot rewards.

**Integration**: V2 Terrain + V2 Entities + V8 Guilds + V9 Economy

## Package Structure

```
pkg/world/raids/
├── doc.go           # Comprehensive package documentation
├── README.md        # This file
├── AUDIT.md         # Implementation gap audit
│
├── types.go         # All type definitions, enums, and constants
│                    # - RaidTier (5 tiers), MechanicType (7 types), RoomType (6 types)
│                    # - RaidDungeon, RaidBoss, BossMechanic, BossPhase
│                    # - RaidRoom, LootTable, RaidInstance, PlayerLockout
│
├── manager.go       # Manager: Unified API for raids
│                    # - GenerateRaid, CreateInstance, CompleteRaid
│                    # - Lockout checking, cleanup operations
│
├── generator.go     # Generator: Procedural raid dungeon creation
│                    # - implements procgen.Generator interface
│                    # - Boss generation, mechanic assignment, loot tables
│
├── instance.go      # InstanceManager: Active instance tracking
│                    # - 4-hour instance timeout
│                    # - Group isolation, cleanup operations
│
├── lockout.go       # LockoutManager: Player lockout tracking
│                    # - 7-day lockout periods per tier
│                    # - Reset scheduling, expiration cleanup
│
├── mechanic.go      # MechanicGenerator: Boss mechanic generation
│                    # - 7 mechanic types (Summon, GroundEffect, Debuff, etc.)
│                    # - Tier-based scaling
│
├── names.go         # BossNameGenerator: Name generation
│                    # - Genre-appropriate raid/boss names
│                    # - Procedural title and name combination
│
└── *_test.go        # Comprehensive test suite (87.6% coverage)
```

## Raid Tiers

Five difficulty tiers with scaling rewards:

| Tier | Difficulty | Players | Lockout | Rewards |
|------|------------|---------|---------|---------|
| Normal | 2.0x | 5-8 | 7 days | Common/Uncommon loot |
| Heroic | 4.0x | 6-9 | 7 days | Uncommon/Rare loot |
| Mythic | 6.0x | 7-10 | 7 days | Rare/Epic loot |
| Legendary | 8.0x | 8-10 | 7 days | Epic/Legendary loot |
| Nightmare | 10.0x | 10 | 7 days | Legendary loot + titles |

## Core Features

### 1. Procedural Generation
- **Deterministic**: Same seed + parameters = identical raid
- **Boss Count**: 3-5 bosses per raid (tier-dependent)
- **Room Layout**: 10-20 rooms using BSP terrain generation
- **Boss Mechanics**: 7 mechanic types with tier scaling
- **Loot Tables**: Procedurally generated with rarity distribution

### 2. Instance System
- **Group Isolation**: Each group gets a separate dungeon instance
- **4-Hour Timeout**: Instances expire 4 hours after creation
- **Auto-Cleanup**: Expired instances removed automatically
- **Concurrent Safe**: Thread-safe with sync.RWMutex

### 3. Lockout System
- **Per-Tier Lockouts**: Players can run each tier once per week
- **7-Day Reset**: Lockouts reset every 7 days
- **Group Validation**: CreateInstance checks all player lockouts
- **Automatic Reset**: Expired lockouts cleaned up automatically

## Quick Start

### Basic Usage

```go
import "github.com/opd-ai/venture/pkg/world/raids"

// Create raid manager
manager := raids.NewManager(12345) // seed

// Generate a raid (without creating instance)
raid, err := manager.GenerateRaid(raids.TierHeroic, 15)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Generated: %s with %d bosses\n", raid.Name, len(raid.Bosses))

// Create instance for a group
groupID := "guild-123"
playerIDs := []string{"player1", "player2", "player3", "player4", "player5"}

instance, err := manager.CreateInstance(raids.TierNormal, 10, groupID, playerIDs)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Instance created: %s\n", instance.InstanceID)

// Check if players can participate
canParticipate, lockedPlayers := manager.CanParticipate(playerIDs, raids.TierNormal)
if !canParticipate {
    fmt.Printf("Locked players: %v\n", lockedPlayers)
}

// Complete raid (applies lockouts)
err = manager.CompleteRaid(instance.InstanceID)
if err != nil {
    log.Fatal(err)
}

// Cleanup expired instances and lockouts
instancesRemoved, lockoutsReset := manager.CleanupExpired()
fmt.Printf("Cleaned: %d instances, %d lockouts\n", instancesRemoved, lockoutsReset)
```

### Advanced Usage

```go
// Use procgen.Generator interface directly
gen := raids.NewGenerator(12345)
params := procgen.GenerationParams{
    Difficulty: 0.6, // Mythic tier (6.0 / 10.0)
    Depth:      20,
    GenreID:    "fantasy",
    Custom: map[string]interface{}{
        "tier":       raids.TierMythic,
        "group_id":   "elite-raiders",
        "group_size": 10,
    },
}

result, err := gen.Generate(67890, params)
if err != nil {
    log.Fatal(err)
}

raid := result.(*raids.RaidDungeon)

// Validate generated raid
if err := gen.Validate(raid); err != nil {
    log.Fatal("Invalid raid:", err)
}

// Access raid details
for i, boss := range raid.Bosses {
    fmt.Printf("Boss %d: %s\n", i+1, boss.Entity.Name)
    fmt.Printf("  Mechanics: %d\n", len(boss.Mechanics))
    fmt.Printf("  Phases: %d\n", len(boss.Phases))
    fmt.Printf("  Loot items: %d\n", len(boss.LootTable.PossibleItems))
}

// Access instance details
instanceMgr := raids.NewInstanceManager()
instance, err := instanceMgr.CreateInstance(raid, "group-456", playerIDs)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Instance ID: %s\n", instance.InstanceID)
fmt.Printf("Expires: %s\n", instance.ExpiresAt)
fmt.Printf("Players: %d\n", len(instance.PlayerIDs))
```

## API Reference

### Manager (Unified API)

**Recommended entry point** for all raid operations:

```go
// Creation
manager := NewManager(seed int64)

// Raid generation
raid, err := manager.GenerateRaid(tier RaidTier, depth int)

// Instance management
instance, err := manager.CreateInstance(tier RaidTier, depth int, groupID string, playerIDs []string)
instance, exists := manager.GetInstance(instanceID string)
err := manager.CompleteRaid(instanceID string)
instance, exists := manager.GetGroupInstance(groupID string, tier RaidTier)

// Lockout checking
canParticipate, lockedPlayers := manager.CanParticipate(playerIDs []string, tier RaidTier)
lockouts := manager.GetPlayerLockouts(playerID string)

// Cleanup
instanceCount, lockoutCount := manager.CleanupExpired()
activeInstances := manager.GetActiveInstanceCount()
activeLockouts := manager.GetActiveLockoutCount()
```

### Generator (Procedural Generation)

Low-level generation interface:

```go
gen := NewGenerator(baseSeed int64)
result, err := gen.Generate(seed int64, params procgen.GenerationParams)
err := gen.Validate(result interface{})
```

### InstanceManager (Instance Lifecycle)

Low-level instance management:

```go
mgr := NewInstanceManager()                                    // Default 4-hour timeout
mgr := NewInstanceManagerWithTimeout(timeout time.Duration)   // Custom timeout

instance, err := mgr.CreateInstance(raid *RaidDungeon, groupID string, playerIDs []string)
instance, exists := mgr.GetInstance(instanceID string)
instance, exists := mgr.GetGroupInstance(groupID string, tier RaidTier)
err := mgr.CompleteInstance(instanceID string)
removed := mgr.CleanupExpired()
count := mgr.GetActiveInstanceCount()
```

### LockoutManager (Lockout Tracking)

Low-level lockout management:

```go
mgr := NewLockoutManager()                                   // Default 7-day period
mgr := NewLockoutManagerWithPeriod(period time.Duration)    // Custom period

mgr.RecordClear(playerID string, tier RaidTier)
locked := mgr.IsLockedOut(playerID string, tier RaidTier)
lockouts := mgr.GetPlayerLockouts(playerID string)
duration := mgr.TimeUntilReset(playerID string, tier RaidTier)
reset := mgr.ResetExpiredLockouts()
count := mgr.GetActiveLockoutCount()
```

## Type Definitions

### RaidDungeon
Complete raid instance with all components:
```go
type RaidDungeon struct {
    ID          string           // Unique raid ID
    Name        string           // Procedural name (e.g., "Temple of the Forgotten")
    Description string           // Generated description
    Tier        RaidTier         // Difficulty tier
    Terrain     *terrain.Terrain // Dungeon layout (BSP-generated)
    Bosses      []*RaidBoss      // 3-5 bosses
    Rooms       []*RaidRoom      // 10-20 rooms
    CreatedAt   time.Time        // Generation timestamp
    Seed        int64            // Generation seed
}
```

### RaidBoss
Boss encounter with mechanics and loot:
```go
type RaidBoss struct {
    Entity    *entity.Entity   // Boss entity (stats, AI)
    RoomID    string           // Room containing boss
    Mechanics []BossMechanic   // Boss abilities
    Phases    []BossPhase      // Health-based phases
    Position  Position         // Initial position
    LootTable *LootTable       // Reward drops
}
```

### BossMechanic
Procedurally generated boss ability:
```go
type BossMechanic struct {
    ID          string           // Unique mechanic ID
    Name        string           // Mechanic name
    Description string           // What it does
    Type        MechanicType     // Summon, GroundEffect, Debuff, etc.
    Cooldown    time.Duration    // Cooldown between uses
    Damage      int              // Base damage
    AoE         bool             // Area-of-effect?
    Radius      float64          // AoE radius (if applicable)
}
```

### MechanicType Constants
```go
const (
    MechanicSummon       // Spawn additional enemies
    MechanicGroundEffect // Create hazard zones
    MechanicDebuff       // Apply negative effects
    MechanicBuff         // Self-buff
    MechanicChanneled    // Continuous ability
    MechanicInstant      // Immediate effect
    MechanicPeriodic     // Damage-over-time
)
```

## Boss Mechanics

Each raid tier generates bosses with varied mechanics:

| Mechanic Type | Effect | Example |
|---------------|--------|---------|
| Summon | Spawns adds | "Calls 3 shadow minions every 30s" |
| GroundEffect | Creates hazard zones | "Poison pools deal 500 dmg/sec" |
| Debuff | Negative player effects | "Curses reduce healing by 50%" |
| Buff | Boss self-buffs | "Enrages at 25% health (+100% dmg)" |
| Channeled | Continuous ability | "Channels beam for 5s (interruptible)" |
| Instant | Immediate damage | "Cleaves for 2000 damage (frontal)" |
| Periodic | Damage-over-time | "Bleeds all players for 300/sec" |

Mechanics scale with raid tier:
- Normal: 2-3 mechanics per boss
- Heroic: 3-4 mechanics per boss
- Mythic: 4-5 mechanics per boss
- Legendary: 5-6 mechanics per boss
- Nightmare: 6-7 mechanics per boss

## Boss Phases

Bosses transition between phases at health thresholds:

```go
type BossPhase struct {
    Number       int      // Phase number (1, 2, 3)
    HealthThresh float64  // Health % to trigger (0.75, 0.50, 0.25)
    Mechanics    []string // Active mechanic IDs in this phase
    AddSpawns    int      // Number of adds spawned
}
```

Example: A mythic boss might have:
- **Phase 1** (100-75%): 2 mechanics active
- **Phase 2** (75-50%): 3 mechanics + 2 adds
- **Phase 3** (50-25%): 4 mechanics + 4 adds
- **Phase 4** (<25%): All mechanics + enrage

## Loot System

Each boss has a procedurally generated loot table:

```go
type LootTable struct {
    GuaranteedItems int           // Always drop this many items
    PossibleItems   []LootItem    // Pool of potential drops
    CurrencyMin     int           // Minimum gold
    CurrencyMax     int           // Maximum gold
}

type LootItem struct {
    ItemID   string   // Item reference
    Rarity   string   // Common, Uncommon, Rare, Epic, Legendary
    DropRate float64  // 0.0-1.0 probability
}
```

Loot rarity distribution by tier:
- **Normal**: 60% Common, 30% Uncommon, 10% Rare
- **Heroic**: 40% Uncommon, 40% Rare, 20% Epic
- **Mythic**: 30% Rare, 50% Epic, 20% Legendary
- **Legendary**: 20% Epic, 60% Legendary, 20% Unique
- **Nightmare**: 100% Legendary + guaranteed title/mount

## Testing

**Coverage**: 87.6% of statements

Run tests:
```bash
go test ./pkg/world/raids
go test -cover ./pkg/world/raids
go test -v ./pkg/world/raids        # verbose output
go test -race ./pkg/world/raids     # race detection
```

### Test Categories

1. **Generation Tests**: Verify raid creation for all tiers
2. **Instance Tests**: Test instance lifecycle and expiration
3. **Lockout Tests**: Validate lockout tracking and resets
4. **Manager Tests**: Integration tests for unified API
5. **Concurrency Tests**: Thread-safety verification
6. **Validation Tests**: Parameter and structure validation

## Performance

From doc.go targets (all met):
- ✅ Generation time: <5s per raid dungeon
- ✅ Memory usage: <50MB per instance
- ✅ Boss count: 3-5 per raid
- ✅ Room count: 10-20 per raid

Concurrency:
- Thread-safe operations with sync.RWMutex
- Read operations allow concurrent access
- Write operations are serialized

## Integration

### Required Systems
- **pkg/procgen/terrain**: BSPGenerator for dungeon layouts
- **pkg/procgen/entity**: EntityGenerator for boss stats

### Planned Integrations
- **V8 Guilds**: Group coordination, lockout tracking (referenced in doc.go)
- **V9 Economy**: Epic/Legendary loot distribution (referenced in doc.go)

### Usage in Game
```go
// Server-side raid manager
raidMgr := raids.NewManager(worldSeed)

// Guild initiates raid
guildID := guild.ID
playerIDs := guild.GetOnlineMemberIDs()

// Check eligibility
canJoin, locked := raidMgr.CanParticipate(playerIDs, raids.TierHeroic)
if !canJoin {
    return fmt.Errorf("players locked out: %v", locked)
}

// Create instance
instance, err := raidMgr.CreateInstance(raids.TierHeroic, depth, guildID, playerIDs)
if err != nil {
    return err
}

// Teleport players to raid
for _, playerID := range playerIDs {
    player := getPlayer(playerID)
    player.Teleport(instance.Dungeon.Terrain.SpawnPoint)
}

// On raid completion
raidMgr.CompleteRaid(instance.InstanceID)  // Applies lockouts

// Periodic cleanup (every hour)
raidMgr.CleanupExpired()
```

## Best Practices

1. **Use Manager for all operations** - Don't interact with Generator/InstanceManager/LockoutManager directly
2. **Check lockouts before creating instances** - Use CanParticipate() first
3. **Run CleanupExpired() periodically** - Prevents memory leaks from expired instances
4. **Validate group size** - Check tier.MinPlayers() and tier.MaxPlayers()
5. **Handle lockout errors gracefully** - Inform players when they're locked out
6. **Use deterministic seeds** - Same group + tier should get same dungeon
7. **Don't modify raid structures** - Raids are immutable after generation

## Common Patterns

### Check and Create Instance
```go
// Check eligibility first
canJoin, lockedPlayers := manager.CanParticipate(playerIDs, tier)
if !canJoin {
    return fmt.Errorf("players locked: %v", lockedPlayers)
}

// Create instance
instance, err := manager.CreateInstance(tier, depth, groupID, playerIDs)
if err != nil {
    return err
}
```

### Reuse Group Instance
```go
// Check if group already has an instance
instance, exists := manager.GetGroupInstance(groupID, tier)
if exists {
    // Teleport to existing instance
    return instance
}

// Create new instance
instance, err = manager.CreateInstance(tier, depth, groupID, playerIDs)
```

### Lockout Checking
```go
lockouts := manager.GetPlayerLockouts(playerID)
for _, lockout := range lockouts {
    fmt.Printf("Tier %s: Next reset in %s\n", 
        lockout.Tier, 
        lockout.NextReset.Sub(time.Now()))
}
```

## Troubleshooting

### "Player is locked out"
- Player completed this tier within the last 7 days
- Check GetPlayerLockouts() for reset time
- Wait for weekly reset or try different tier

### "Instance not found"
- Instance may have expired (4-hour timeout)
- Check GetGroupInstance() for active instance
- Create new instance if needed

### "Invalid difficulty"
- Difficulty must be 0.0-1.0 in GenerationParams
- Use tier.DifficultyMultiplier() / 10.0 for conversion

### High memory usage
- Run CleanupExpired() regularly
- Check GetActiveInstanceCount()
- Verify instance timeout is reasonable

## Future Enhancements

See AUDIT.md Priority 3 for optional improvements:
- Boss mechanic balance testing
- Loot table verification
- Instance stress testing
- Save/load support for persistence

## References

- Package documentation: `doc.go`
- Implementation audit: `AUDIT.md`
- Type definitions: `types.go`
- Project architecture: `PLAN.md` (root)
- Integration roadmap: `ROADMAP_V4.md`
