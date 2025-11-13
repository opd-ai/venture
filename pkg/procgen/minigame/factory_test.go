package minigame

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

func TestCreateGameInstance(t *testing.T) {
	tests := []struct {
		name     string
		gameType GameType
		wantErr  bool
	}{
		{"card game", GameTypeCard, false},
		{"dice game", GameTypeDice, false},
		{"puzzle game", GameTypePuzzle, false},
		{"memory game", GameTypeMemory, false},
		{"lock-picking game", GameTypeLockPicking, false},
		{"hacking game", GameTypeHacking, false},
		{"ritual game", GameTypeRitual, false},
		{"invalid type", GameType(999), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := CreateGameInstance(tt.gameType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGameInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if game == nil {
					t.Error("CreateGameInstance() returned nil game")
				}

				// Test that game can be initialized
				if err := game.Initialize(12345, 0.5); err != nil {
					t.Errorf("game.Initialize() failed: %v", err)
				}
			}
		})
	}
}

func TestGameTypeConversion(t *testing.T) {
	gameTypes := []GameType{
		GameTypeCard,
		GameTypeDice,
		GameTypePuzzle,
		GameTypeMemory,
		GameTypeLockPicking,
		GameTypeHacking,
		GameTypeRitual,
	}

	for _, gt := range gameTypes {
		t.Run(gt.String(), func(t *testing.T) {
			// Convert to engine type and back
			engineType := GameTypeToEngineType(gt)
			backToGameType := EngineTypeToGameType(engineType)

			if backToGameType != gt {
				t.Errorf("Conversion roundtrip failed: %v -> %v -> %v", gt, engineType, backToGameType)
			}
		})
	}
}

func TestGameTypeToEngineType(t *testing.T) {
	tests := []struct {
		gameType   GameType
		engineType engine.MiniGameType
	}{
		{GameTypeCard, engine.MiniGameCard},
		{GameTypeDice, engine.MiniGameDice},
		{GameTypePuzzle, engine.MiniGamePuzzle},
		{GameTypeMemory, engine.MiniGameMemory},
		{GameTypeLockPicking, engine.MiniGameLockPicking},
		{GameTypeHacking, engine.MiniGameHacking},
		{GameTypeRitual, engine.MiniGameRitual},
	}

	for _, tt := range tests {
		t.Run(tt.gameType.String(), func(t *testing.T) {
			result := GameTypeToEngineType(tt.gameType)
			if result != tt.engineType {
				t.Errorf("GameTypeToEngineType() = %v, want %v", result, tt.engineType)
			}
		})
	}
}

func TestEngineTypeToGameType(t *testing.T) {
	tests := []struct {
		engineType engine.MiniGameType
		gameType   GameType
	}{
		{engine.MiniGameCard, GameTypeCard},
		{engine.MiniGameDice, GameTypeDice},
		{engine.MiniGamePuzzle, GameTypePuzzle},
		{engine.MiniGameMemory, GameTypeMemory},
		{engine.MiniGameLockPicking, GameTypeLockPicking},
		{engine.MiniGameHacking, GameTypeHacking},
		{engine.MiniGameRitual, GameTypeRitual},
	}

	for _, tt := range tests {
		t.Run(tt.engineType.String(), func(t *testing.T) {
			result := EngineTypeToGameType(tt.engineType)
			if result != tt.gameType {
				t.Errorf("EngineTypeToGameType() = %v, want %v", result, tt.gameType)
			}
		})
	}
}

func TestAllGameTypesHaveImplementations(t *testing.T) {
	// Ensure all GameType values can create instances
	for i := GameTypeCard; i <= GameTypeRitual; i++ {
		game, err := CreateGameInstance(i)
		if err != nil {
			t.Errorf("GameType %d (%s) has no implementation", i, i.String())
			continue
		}
		if game == nil {
			t.Errorf("GameType %d (%s) returned nil game", i, i.String())
		}
	}
}
