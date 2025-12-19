# Music Trigger System

## Overview

The Music Trigger System integrates adaptive music with gameplay events in Venture. It automatically adjusts the music based on combat state, boss encounters, quest completions, exploration milestones, and reputation changes.

## Architecture

The system consists of two main components:

### MusicTriggerComponent

A component attached to entities (typically the player) that tracks music-relevant game state:

- **Combat State**: Whether combat is active, when it last occurred
- **Boss State**: Whether a boss is nearby, when it last appeared
- **Exploration**: Number of exploration milestones reached
- **Reputation**: Current reputation tier affecting ambient tension
- **Pending Transitions**: Temporary music changes (e.g., victory fanfares)

### MusicTriggerSystem

An ECS system that processes music trigger events and updates the adaptive music system:

- Maintains an event queue for music-relevant gameplay events
- Processes events at regular intervals (0.5s) to avoid excessive updates
- Applies context changes to the AdaptiveMusicSystem
- Handles temporary music transitions (victory music, etc.)

## Event Types

### TriggerCombatStart / TriggerCombatEnd
Triggers when combat begins or ends. Sets danger level to 0.6 during combat, 0.2 after combat.

```go
// Start combat
system.OnCombatStart(playerEntityID)

// End combat  
system.OnCombatEnd(playerEntityID)
```

### TriggerBossAppear / TriggerBossDefeated
Triggers when a boss appears or is defeated. Sets danger to maximum (1.0) during boss fights, triggers victory music when defeated.

```go
// Boss appears
system.OnBossAppear(playerEntityID)

// Boss defeated
system.OnBossDefeated(playerEntityID) // Also triggers victory music
```

### TriggerQuestComplete
Triggers temporary victory music when a quest is completed.

```go
system.OnQuestComplete(playerEntityID)
```

### TriggerExplorationMilestone
Triggers when the player discovers a new area or reaches an exploration milestone.

```go
// New area discovered
system.OnExplorationMilestone(playerEntityID, true)

// Same area milestone
system.OnExplorationMilestone(playerEntityID, false)
```

### TriggerReputationChange
Triggers when player reputation changes significantly. Affects ambient danger level.

```go
system.OnReputationChange(playerEntityID, "revered") // Low danger
system.OnReputationChange(playerEntityID, "hated")   // High danger
```

Reputation tiers (from hostile to friendly):
- `"hated"` - Danger 1.0
- `"hostile"` - Danger 0.8
- `"unfriendly"` - Danger 0.5
- `"neutral"` - Danger 0.3
- `"friendly"` - Danger 0.1
- `"honored"` - Danger 0.0
- `"revered"` - Danger 0.0

## Integration Guide

### Step 1: Create Music Trigger Component

```go
// Add to player entity
player := world.CreateEntity()
musicTrigger := engine.NewMusicTriggerComponent()
player.AddComponent(musicTrigger)
```

### Step 2: Initialize Music Trigger System

```go
// Create adaptive music manager
musicManager := music.NewAdaptiveMusicManager(44100, worldSeed)
musicManager.Initialize(genreID, bpm)

// Create trigger system
musicTriggerSystem := engine.NewMusicTriggerSystem(world, musicManager)

// Add to world systems
world.AddSystem(musicTriggerSystem)
```

### Step 3: Trigger Events from Gameplay Systems

```go
// In combat system
if combatStarted {
    musicTriggerSystem.OnCombatStart(player.ID)
}

// In boss AI
if bossSpawned {
    musicTriggerSystem.OnBossAppear(player.ID)
}

// In quest system
if questCompleted {
    musicTriggerSystem.OnQuestComplete(player.ID)
}
```

### Step 4: Update in Game Loop

```go
func (g *Game) Update() error {
    deltaTime := 1.0 / 60.0
    
    // World update processes all systems including music trigger system
    g.world.Update(deltaTime)
    
    return nil
}
```

## Music Context Mapping

The system maps game state to music contexts:

| Game State | Location | Combat | BossNearby | Danger | TimeOfDay |
|------------|----------|--------|------------|--------|-----------|
| Exploration | "exploration" | false | false | 0.0-0.3 | current |
| Combat | "exploration" | true | false | 0.6 | current |
| Boss Fight | "exploration" | true | true | 1.0 | current |
| Victory | "victory" | false | false | 0.0 | current |
| Post-Combat | "exploration" | false | false | 0.2 | current |

## Performance

- **Event Queue**: Events are queued and processed in batches
- **Update Interval**: Music context updated every 0.5 seconds (configurable)
- **Immediate Processing**: Events are processed immediately, but entity loop runs at intervals
- **Memory**: Minimal overhead (~200 bytes per component)

## Testing

The system includes comprehensive tests:

```bash
# Run all music trigger tests
go test -run=TestMusicTrigger ./pkg/engine/

# Run specific test
go test -run=TestMusicTriggerSystem_OnCombatStart ./pkg/engine/

# Check coverage
go test -cover ./pkg/engine/
```

Coverage: ~90% for music trigger components and system.

## Example: Complete Integration

```go
package main

import (
    "github.com/opd-ai/venture/pkg/engine"
    "github.com/opd-ai/venture/pkg/audio/music"
)

func main() {
    // Create world
    world := engine.NewWorld()
    
    // Create music system
    musicManager := music.NewAdaptiveMusicManager(44100, 12345)
    musicManager.Initialize("fantasy", 120)
    
    // Create trigger system
    musicTriggerSystem := engine.NewMusicTriggerSystem(world, musicManager)
    world.AddSystem(musicTriggerSystem)
    
    // Create player
    player := world.CreateEntity()
    player.AddComponent(engine.NewMusicTriggerComponent())
    player.AddComponent(engine.NewPositionComponent(100, 100))
    
    // Process pending additions
    world.Update(0.0)
    
    // Simulate gameplay events
    musicTriggerSystem.OnCombatStart(player.ID)
    world.Update(0.016) // ~60 FPS
    
    musicTriggerSystem.OnBossAppear(player.ID)
    world.Update(0.016)
    
    // ... continue game loop ...
}
```

## Troubleshooting

### Music doesn't change when events occur

**Cause**: Entities created with `CreateEntity()` aren't added to the world until `World.Update()` is called.

**Solution**: Call `world.Update(0.0)` after creating entities and before triggering events.

```go
entity := world.CreateEntity()
entity.AddComponent(musicTrigger)
world.Update(0.0) // Commit entity to world
system.OnCombatStart(entity.ID) // Now works correctly
```

### Music updates lag behind gameplay

**Cause**: Default update interval is 0.5 seconds.

**Solution**: Access the system's updateInterval field (requires making it public) or accept the slight delay as intended for performance.

### Events are lost

**Cause**: Event queue is processed and cleared each update.

**Solution**: Ensure `MusicTriggerSystem.Update()` is called regularly (every frame via `World.Update()`).

## Future Enhancements

Planned improvements for future phases:

1. **Priority System**: High-priority events (boss music) override lower-priority events
2. **Crossfade Control**: Configurable transition times between contexts
3. **Context Stack**: Push/pop contexts for nested scenarios (boss fight during quest)
4. **Location Awareness**: Different music for different dungeon areas
5. **Dynamic BPM**: Adjust tempo based on action intensity

## Related Documentation

- [Adaptive Music System](../pkg/audio/music/doc.go) - Core music generation
- [Music Interfaces](../pkg/audio/interfaces.go) - AdaptiveMusicSystem interface
- [ECS Architecture](ARCHITECTURE.md#entity-component-system) - ECS pattern overview
- [Roadmap V8](ROADMAP_V8.md) - Full phase plan
