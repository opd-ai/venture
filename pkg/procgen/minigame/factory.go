package minigame

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
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

// GenerateAndCreateGame combines procedural generation with game instantiation.
// This is a convenience function that generates minigame metadata and creates
// a ready-to-play game instance in a single call.
//
// Parameters:
//   - seed: Seed for deterministic generation
//   - params: Generation parameters (difficulty, depth, genre, etc.)
//
// Returns:
//   - *MiniGame: Generated minigame metadata
//   - engine.MiniGame: Initialized game instance ready to play
//   - error: An error if generation or instantiation fails
//
// Example usage:
//
//	generator := minigame.NewGenerator()
//	metadata, instance, err := minigame.GenerateAndCreateGame(generator, seed, params)
//	if err != nil {
//	    return err
//	}
//	// Use metadata.Difficulty, metadata.TimeLimit, etc. for setup
//	// Use instance for gameplay via instance.Update(), instance.HandleInput(), etc.
func GenerateAndCreateGame(generator *Generator, seed int64, params procgen.GenerationParams) (*MiniGame, engine.MiniGame, error) {
	// Generate minigame metadata
	result, err := generator.Generate(seed, params)
	if err != nil {
		return nil, nil, fmt.Errorf("generation failed: %w", err)
	}

	metadata, ok := result.(*MiniGame)
	if !ok {
		return nil, nil, fmt.Errorf("generator returned invalid type")
	}

	// Validate generated minigame
	if err := generator.Validate(metadata); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create game instance
	instance, err := CreateGameInstance(metadata.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiation failed: %w", err)
	}

	// Initialize the instance with seed and difficulty
	if err := instance.Initialize(seed, metadata.Difficulty); err != nil {
		return nil, nil, fmt.Errorf("initialization failed: %w", err)
	}

	return metadata, instance, nil
}
