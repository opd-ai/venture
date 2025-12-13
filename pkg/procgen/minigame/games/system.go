// Package games contains implementations of mini-game types for Phase 27.2.
// This file provides an ECS system wrapper for minigame implementations.
package games

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
)

// System is an ECS wrapper that provides factory methods for creating
// concrete minigame implementations. It doesn't need Update() logic since
// the MiniGameSystem handles the actual game lifecycle.
//
// Phase 3.4: Minigame Implementations (PLAN.md)
type System struct {
	world *engine.World
}

// NewSystem creates a new minigame games system.
// This system provides factory methods to create concrete game instances.
func NewSystem(world *engine.World) *System {
	return &System{
		world: world,
	}
}

// Update implements the engine.System interface but is a no-op.
// The actual game updates are handled by MiniGameSystem.
func (s *System) Update(entities []*engine.Entity, deltaTime float64) {
	// No-op: MiniGameSystem handles game updates
}

// CreateGame creates a concrete minigame implementation based on the game type.
// The returned game implements the engine.MiniGame interface.
//
// Parameters:
//   - gameType: The type of minigame to create (e.g., MiniGameCard, MiniGameDice)
//
// Returns:
//   - engine.MiniGame: The created game instance, or nil if type is unknown
//   - error: An error if the game type is invalid or creation fails
func (s *System) CreateGame(gameType engine.MiniGameType) (engine.MiniGame, error) {
	switch gameType {
	case engine.MiniGameCard:
		return NewCardGame(), nil
	case engine.MiniGameDice:
		return NewDiceGame(), nil
	case engine.MiniGamePuzzle:
		return NewPuzzleGame(), nil
	case engine.MiniGameMemory:
		return NewMemoryGame(), nil
	case engine.MiniGameLockPicking:
		return NewLockPickingGame(), nil
	case engine.MiniGameHacking:
		return NewHackingGame(), nil
	case engine.MiniGameRitual:
		return NewRitualGame(), nil
	default:
		return nil, fmt.Errorf("unknown minigame type: %v", gameType)
	}
}

// GetAvailableGames returns a list of all available minigame types.
// This is useful for UI or random selection.
func (s *System) GetAvailableGames() []engine.MiniGameType {
	return []engine.MiniGameType{
		engine.MiniGameCard,
		engine.MiniGameDice,
		engine.MiniGamePuzzle,
		engine.MiniGameMemory,
		engine.MiniGameLockPicking,
		engine.MiniGameHacking,
		engine.MiniGameRitual,
	}
}
