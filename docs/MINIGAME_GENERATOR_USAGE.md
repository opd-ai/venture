# Minigame Generator Usage Guide

This guide demonstrates the proper usage patterns for procedural minigame generation in Venture.

## Overview

The minigame system uses a two-phase approach:
1. **Generation**: Create minigame metadata (type, difficulty, rules, etc.) using `minigame.Generator`
2. **Instantiation**: Create playable game instances using factory functions

## Quick Start

### Combined Generation + Instantiation

The simplest way to create a ready-to-play minigame:

```go
import (
    "github.com/opd-ai/venture/pkg/procgen"
    "github.com/opd-ai/venture/pkg/procgen/minigame"
)

func createMinigame(seed int64, difficulty float64, genreID string) error {
    generator := minigame.NewGenerator()
    
    params := procgen.GenerationParams{
        Difficulty: difficulty,
        Depth:      10,
        GenreID:    genreID,
        Custom:     make(map[string]interface{}),
    }
    
    // Generate metadata and create instance in one call
    metadata, instance, err := minigame.GenerateAndCreateGame(generator, seed, params)
    if err != nil {
        return fmt.Errorf("failed to create minigame: %w", err)
    }
    
    // metadata contains:
    //   - Type (GameType)
    //   - Name (string)
    //   - Difficulty (float64)
    //   - TimeLimit (float64)
    //   - Rules (string)
    //   - State (interface{})
    
    // instance is ready to play:
    //   - Initialize() already called
    //   - Update(deltaTime) to advance game
    //   - HandleInput(input) for player actions
    //   - IsComplete() to check win/loss
    //   - GetReward() for completion rewards
    
    return nil
}
```

## Step-by-Step Usage

### Phase 1: Generate Minigame Metadata

```go
generator := minigame.NewGenerator()
params := procgen.GenerationParams{
    Difficulty: 0.5,  // 0.0 = easy, 1.0 = hard
    Depth:      10,   // Game progression depth
    GenreID:    "fantasy",
    Custom:     make(map[string]interface{}),
}

result, err := generator.Generate(seed, params)
if err != nil {
    return err
}

metadata := result.(*minigame.MiniGame)
// metadata.Type = GameTypeCard, GameTypeDice, etc.
// metadata.Difficulty = 0.5
// metadata.TimeLimit = 600.0 (seconds)
// metadata.Name = "Dragon Hold'em" (genre-specific)
```

### Phase 2: Create Game Instance

#### Option A: Direct Factory Call

```go
instance, err := minigame.CreateGameInstance(metadata.Type)
if err != nil {
    return err
}

// Initialize with seed and difficulty
if err := instance.Initialize(seed, metadata.Difficulty); err != nil {
    return err
}
```

#### Option B: Using games.System (ECS integration)

```go
import "github.com/opd-ai/venture/pkg/procgen/minigame/games"

gamesSystem := games.NewSystem(world)

// Convert procgen type to engine type
engineType := minigame.GameTypeToEngineType(metadata.Type)

instance, err := gamesSystem.CreateGame(engineType)
if err != nil {
    return err
}

if err := instance.Initialize(seed, metadata.Difficulty); err != nil {
    return err
}
```

## Type Conversion

The system uses two type enums that need conversion:

```go
// Procgen type -> Engine type
procgenType := minigame.GameTypeCard
engineType := minigame.GameTypeToEngineType(procgenType)
// engineType = engine.MiniGameCard

// Engine type -> Procgen type
engineType := engine.MiniGameDice
procgenType := minigame.EngineTypeToGameType(engineType)
// procgenType = minigame.GameTypeDice
```

**Always use the conversion functions** instead of manual switch-case to avoid duplication.

## Common Patterns

### Pattern 1: Adding Minigames to Merchants

```go
// Generate minigame metadata
metadata, instance, err := minigame.GenerateAndCreateGame(generator, seed, params)
if err != nil {
    return err
}

// Add component to entity
gameType := minigame.GameTypeToEngineType(metadata.Type)
entity.AddComponent(&engine.MiniGameComponent{
    GameType:     gameType,
    Active:       false,
    State:        metadata.State,
    Difficulty:   metadata.Difficulty,
    TimeLimit:    metadata.TimeLimit,
    GameInstance: nil, // Set when player starts game
})
```

### Pattern 2: Starting a Minigame

```go
// Player activates minigame
miniGameSystem.StartGame(entityID, gameType, difficulty)

// Create and set instance
instance, err := minigame.CreateGameInstance(procgenType)
if err != nil {
    return err
}

if err := instance.Initialize(seed, difficulty); err != nil {
    return err
}

miniGameSystem.SetGameInstance(entityID, instance)
```

### Pattern 3: Genre-Based Generation

```go
genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

for _, genre := range genres {
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        GenreID:    genre,
        // ... other params
    }
    
    metadata, instance, err := minigame.GenerateAndCreateGame(generator, seed, params)
    // Generates genre-appropriate names and styling
}
```

## Testing

### Unit Tests

```go
func TestMinigameGeneration(t *testing.T) {
    generator := minigame.NewGenerator()
    params := procgen.GenerationParams{
        Difficulty: 0.5,
        GenreID:    "fantasy",
    }
    
    // Test determinism
    metadata1, instance1, _ := minigame.GenerateAndCreateGame(generator, 12345, params)
    metadata2, instance2, _ := minigame.GenerateAndCreateGame(generator, 12345, params)
    
    if metadata1.Type != metadata2.Type {
        t.Error("Non-deterministic generation")
    }
}
```

### Integration Tests

```go
func TestMinigamePlaythrough(t *testing.T) {
    _, instance, err := minigame.GenerateAndCreateGame(generator, seed, params)
    if err != nil {
        t.Fatal(err)
    }
    
    // Simulate gameplay
    for i := 0; i < 100; i++ {
        if instance.IsComplete() {
            break
        }
        instance.Update(0.016) // 60 FPS
    }
    
    reward := instance.GetReward()
    if reward == nil {
        t.Error("Expected reward on completion")
    }
}
```

## Performance Considerations

- **Generator**: Lightweight, creates metadata in <1ms
- **Factory**: Fast instantiation, creates game object in <0.1ms
- **Combined function**: Total time <2ms for generation + instantiation
- **Caching**: Consider caching generated metadata for merchants/NPCs

## Migration from Direct Construction

**Before** (manual type mapping):
```go
var gameType engine.MiniGameType
switch mg.Type {
case minigame.GameTypeCard:
    gameType = engine.MiniGameCard
case minigame.GameTypeDice:
    gameType = engine.MiniGameDice
// ... 5 more cases
}
```

**After** (using factory function):
```go
gameType := minigame.GameTypeToEngineType(mg.Type)
```

## Best Practices

1. ✅ **Use `GenerateAndCreateGame()`** for new minigames
2. ✅ **Use type conversion functions** instead of switch-case
3. ✅ **Validate difficulty** (0.0 to 1.0 range)
4. ✅ **Use deterministic seeds** for multiplayer sync
5. ✅ **Cache metadata** for static content (merchants)
6. ❌ **Don't create instances manually** - use factory
7. ❌ **Don't skip Initialize()** - required for game state

## Architecture

```
┌─────────────────┐
│ Client/Server   │
│   Game Logic    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────────┐
│   Generator     │─────▶│   Metadata       │
│  (procgen)      │      │  (Type, Rules,   │
└─────────────────┘      │   Difficulty)    │
         │               └──────────────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────────┐
│    Factory      │─────▶│  Game Instance   │
│ (instantiation) │      │  (Playable impl) │
└─────────────────┘      └──────────────────┘
         │
         ▼
┌─────────────────┐
│  MiniGameSystem │
│   (ECS runtime) │
└─────────────────┘
```

## Further Reading

- `pkg/procgen/minigame/generator.go` - Procedural generation logic
- `pkg/procgen/minigame/factory.go` - Factory functions
- `pkg/procgen/minigame/games/` - Concrete game implementations
- `pkg/engine/minigame_system.go` - ECS system integration
