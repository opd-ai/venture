# Package: games

## Overview

The `games` package provides complete implementations of 7 playable mini-game types for the Venture action-RPG. All games implement the `engine.MiniGame` interface with deterministic gameplay, difficulty scaling, and genre-appropriate theming.

**Phase**: 27.2 - Mini-Game Types (ROADMAP_V4.md)

## Package Structure

```
pkg/procgen/minigame/games/
├── doc.go              # Package documentation
├── README.md           # This file
├── AUDIT.md            # Implementation gap audit
│
├── card.go             # CardGame: Deck-based card game with AI opponent
├── dice.go             # DiceGame: Multi-die rolling with betting mechanics
├── puzzle.go           # PuzzleGame: Sliding tile and pattern matching
├── memory.go           # MemoryGame: Card pair matching challenges
├── lockpicking.go      # LockPickingGame: Timing-based lock picking
├── hacking.go          # HackingGame: Code-breaking terminal puzzles
├── ritual.go           # RitualGame: Spell pattern drawing (Symbol, Point types)
│
├── system.go           # System: ECS wrapper and factory methods
│
└── *_test.go           # Comprehensive test suite (86% coverage)
```

## File Organization

### Single Responsibility Principle
Each file contains exactly one game implementation with its related types and methods:

- **One struct per file**: Each game type has its own file
- **Co-located helpers**: Helper functions are in the same file as the struct that uses them
- **Shared types only when necessary**: Symbol and Point types are in ritual.go (only used there)

### Naming Conventions
- Game files: `{gametype}.go` (e.g., `card.go`, `dice.go`)
- Test files: `{gametype}_test.go` or `games_test.go` for integration tests
- System file: `system.go` for factory and ECS wrapper

## Game Implementations

### 1. CardGame (`card.go`)
**Duration**: 5-10 minutes  
**Description**: Procedural card game where players compete against an AI opponent across multiple rounds. The player with the most round wins reaches the target first.

**Key Features**:
- Deterministic deck shuffling
- AI difficulty scaling (opponent bonus increases with difficulty)
- Variable deck size (20-52 cards) and hand size (5-7 cards)

**Methods**:
- `NewCardGame()` - Constructor
- `Initialize(seed, difficulty)` - Setup with difficulty 0.0-1.0
- `Update(deltaTime)` - Play one round per update
- `PrepareRender(screenWidth, screenHeight)` - Prepare visual state (stub for Phase 27.3)
- `GetRenderOutput()` - Get computed render data
- `IsComplete()` - Check if game finished
- `GetReward()` - Return gold/XP rewards

### 2. DiceGame (`dice.go`)
**Duration**: 2-5 minutes  
**Description**: Custom dice rules with betting mechanics. Players roll multiple dice and compete to win rounds.

**Key Features**:
- Configurable dice count (2-5) and sides (6-12)
- Betting system with difficulty-scaled amounts
- Target round wins (5-10)

### 3. PuzzleGame (`puzzle.go`)
**Duration**: 3-7 minutes  
**Description**: Sliding tile puzzle where players rearrange a shuffled grid to match the solution.

**Key Features**:
- Grid size scales 3x3 to 5x5 with difficulty
- Move limit based on grid size
- Bonus rewards for efficient solutions

### 4. MemoryGame (`memory.go`)
**Duration**: 2-4 minutes  
**Description**: Card pair matching with attempt limits. Can switch to sequence mode for higher difficulties.

**Key Features**:
- Pair count scales 4-12 with difficulty
- Sequence mode activated at difficulty >0.6
- Attempt limit: 20-30 based on difficulty

### 5. LockPickingGame (`lockpicking.go`)
**Duration**: 0.5-2 minutes  
**Description**: Timing-based challenge where players align pins within timing windows.

**Key Features**:
- Pin count: 3-7 pins
- Timing window: 0.5s to 0.1s
- Failure limit: 5 to 2 allowed failures

### 6. HackingGame (`hacking.go`)
**Duration**: 1-3 minutes  
**Description**: Terminal puzzle where players deduce an alphanumeric code using feedback hints.

**Key Features**:
- Code length: 4-8 characters
- Attempt limit: 10-6 attempts
- Hint system (correct position vs. misplaced)

### 7. RitualGame (`ritual.go`)
**Duration**: 2-5 minutes  
**Description**: Spell pattern drawing for fantasy/horror themes. Players draw procedural symbols accurately.

**Key Features**:
- Symbol count: 3-7 symbols
- Accuracy requirement: 0.7 to 0.95
- Procedural symbol generation (Circle, Star, Triangle, Spiral, Pentagram, Cross, Rune)

**Special Types**:
- `Symbol`: Pattern definition (Name, Points)
- `Point`: 2D coordinate (X, Y float64)

### System Integration (`system.go`)

The `System` struct provides factory methods for creating game instances:

```go
system := NewSystem(world)
game, err := system.CreateGame(engine.MiniGameCard)
if err != nil {
    log.Fatal(err)
}
```

**Key Methods**:
- `CreateGame(gameType)` - Factory method for all 7 game types
- `GetAvailableGames()` - List all game types
- `Update(entities, deltaTime)` - No-op (MiniGameSystem handles updates)

## Usage Example

```go
import "github.com/opd-ai/venture/pkg/procgen/minigame/games"

// Create a dice game
game := games.NewDiceGame()

// Initialize with seed and difficulty (0.0 = easy, 1.0 = hard)
if err := game.Initialize(12345, 0.5); err != nil {
    log.Fatal(err)
}

// Game loop
for !game.IsComplete() {
    // Update game state (plays one round)
    if err := game.Update(deltaTime); err != nil {
        log.Fatal(err)
    }
    
    // Prepare render state (stub in Phase 27.2)
    if err := game.PrepareRender(screenWidth, screenHeight); err != nil {
        log.Fatal(err)
    }
    
    // Get render output for drawing
    renderOutput := game.GetRenderOutput()
    // Use renderOutput for drawing...
}

// Award rewards
if reward := game.GetReward(); reward != nil {
    player.Gold += reward.Gold
    player.XP += reward.XP
}
```

## Testing

**Coverage**: 86.0% of statements

Run tests:
```bash
go test ./pkg/procgen/minigame/games
go test -cover ./pkg/procgen/minigame/games
go test -v ./pkg/procgen/minigame/games  # verbose output
```

### Test Categories

1. **Determinism Tests**: Verify same seed produces identical results
2. **Difficulty Scaling**: Validate parameters scale correctly
3. **Completion Detection**: Ensure games end properly
4. **Reward Calculation**: Verify gold/XP formulas
5. **Interface Compliance**: Confirm all games implement engine.MiniGame
6. **Factory Tests**: Validate System.CreateGame for all types

### Coverage Details

- **Well-tested** (>90%): Initialize, Update, IsComplete, factory methods
- **Partially tested** (25-40%): GetReward (loss conditions)
- **Intentionally untested** (0%): PrepareRender/GetRenderOutput methods (stubs for Phase 27.3)

## Design Patterns

### Deterministic Generation
All games use seeded RNG for reproducible gameplay:
```go
rng := rand.New(rand.NewSource(seed))
```

### Difficulty Scaling
All parameters scale linearly from easy to hard:
```go
// Easy: 3 items, Hard: 7 items
numItems := 3 + int(difficulty*4)

// Easy: 0.7 accuracy, Hard: 0.95 accuracy
accuracy := 0.7 + difficulty*0.25
```

### Reward Calculation
Rewards scale with difficulty and performance:
```go
goldReward := baseGold + int(difficulty*bonusGold)
xpReward := baseXP + (difficulty * bonusXP)
```

### Factory Pattern
System provides centralized game creation:
```go
func (s *System) CreateGame(gameType engine.MiniGameType) (engine.MiniGame, error)
```

## Dependencies

### External Dependencies
- `math/rand` - Deterministic RNG
- `fmt` - Error formatting
- `math` - Trigonometry for symbol generation (ritual.go only)
- `strings` - String manipulation (hacking.go only)

### Internal Dependencies
- `pkg/engine` - MiniGame interface, Reward type, World, Entity, MiniGameType

**No circular dependencies**

## Performance Characteristics

All games meet target performance:
- **Initialize**: <1ms per game
- **Update**: <0.1ms per frame
- **Memory**: <1KB per game instance

Designed for 60 FPS gameplay with minimal overhead.

## Future Work (Phase 27.3)

### Render Implementations
All PrepareRender()/GetRenderOutput() methods are currently stubs. Phase 27.3 will implement:
- Visual representation for each game type
- Genre-appropriate styling (fantasy, sci-fi, horror themes)
- Animation and effects
- Performance-optimized rendering

### Optional Enhancements
- Item rewards for hard difficulty completions
- Multiplayer state synchronization (Serialize/Deserialize)
- Save/Load support for pause/resume
- Additional game types (Archery, Racing, Crafting puzzles)

## Quality Metrics

- ✅ Test Coverage: 86.0% (target: >65%)
- ✅ Documentation: 100% of exported symbols
- ✅ Build: No errors or warnings
- ✅ Determinism: All games verified
- ✅ Interface Compliance: All games implement engine.MiniGame
- ⚠️ Render: Stubs only (deferred to Phase 27.3)

See `AUDIT.md` for detailed implementation gap analysis.

## Contributing

When adding new game types:

1. Create new file: `{gametype}.go`
2. Add file-level comment describing the game
3. Implement all engine.MiniGame methods
4. Add factory case to System.CreateGame()
5. Add to System.GetAvailableGames()
6. Write comprehensive tests (aim for >85% coverage)
7. Document difficulty scaling and reward formulas
8. Verify determinism with seed-based tests

## References

- Package documentation: `doc.go`
- Implementation audit: `AUDIT.md`
- Engine interface: `pkg/engine/minigame.go`
- Phase roadmap: `ROADMAP_V4.md` (Phase 27.2)
- Project conventions: `PLAN.md`, `AUDIT.md` (root)
