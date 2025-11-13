package minigame

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/minigame/games"
)

// CreateGameInstance creates an engine.MiniGame instance from a GameType.
// This factory function instantiates the actual game implementations.
//
// Phase 27.2: Mini-Game Types
func CreateGameInstance(gameType GameType) (engine.MiniGame, error) {
	switch gameType {
	case GameTypeCard:
		return games.NewCardGame(), nil
	case GameTypeDice:
		return games.NewDiceGame(), nil
	case GameTypePuzzle:
		return games.NewPuzzleGame(), nil
	case GameTypeMemory:
		return games.NewMemoryGame(), nil
	case GameTypeLockPicking:
		return games.NewLockPickingGame(), nil
	case GameTypeHacking:
		return games.NewHackingGame(), nil
	case GameTypeRitual:
		return games.NewRitualGame(), nil
	default:
		return nil, fmt.Errorf("unknown game type: %d", gameType)
	}
}

// GameTypeToEngineType converts a procgen GameType to an engine MiniGameType.
// This is needed for compatibility between the generator and engine systems.
func GameTypeToEngineType(gameType GameType) engine.MiniGameType {
	switch gameType {
	case GameTypeCard:
		return engine.MiniGameCard
	case GameTypeDice:
		return engine.MiniGameDice
	case GameTypePuzzle:
		return engine.MiniGamePuzzle
	case GameTypeMemory:
		return engine.MiniGameMemory
	case GameTypeLockPicking:
		return engine.MiniGameLockPicking
	case GameTypeHacking:
		return engine.MiniGameHacking
	case GameTypeRitual:
		return engine.MiniGameRitual
	default:
		return engine.MiniGameCard // fallback
	}
}

// EngineTypeToGameType converts an engine MiniGameType to a procgen GameType.
func EngineTypeToGameType(engineType engine.MiniGameType) GameType {
	switch engineType {
	case engine.MiniGameCard:
		return GameTypeCard
	case engine.MiniGameDice:
		return GameTypeDice
	case engine.MiniGamePuzzle:
		return GameTypePuzzle
	case engine.MiniGameMemory:
		return GameTypeMemory
	case engine.MiniGameLockPicking:
		return GameTypeLockPicking
	case engine.MiniGameHacking:
		return GameTypeHacking
	case engine.MiniGameRitual:
		return GameTypeRitual
	default:
		return GameTypeCard // fallback
	}
}
