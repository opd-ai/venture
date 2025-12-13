package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestNewSystem verifies system creation.
func TestNewSystem(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.world != world {
		t.Error("System world not set correctly")
	}
}

// TestSystemUpdate verifies the Update method is a no-op.
func TestSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	// Should not panic or error
	sys.Update(nil, 0.016)
	sys.Update([]*engine.Entity{}, 1.0)
}

// TestCreateGame verifies all game types can be created.
func TestCreateGame(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	tests := []struct {
		name     string
		gameType engine.MiniGameType
		wantErr  bool
	}{
		{"Card game", engine.MiniGameCard, false},
		{"Dice game", engine.MiniGameDice, false},
		{"Puzzle game", engine.MiniGamePuzzle, false},
		{"Memory game", engine.MiniGameMemory, false},
		{"Lock picking", engine.MiniGameLockPicking, false},
		{"Hacking game", engine.MiniGameHacking, false},
		{"Ritual game", engine.MiniGameRitual, false},
		{"Invalid type", engine.MiniGameType(999), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := sys.CreateGame(tt.gameType)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid game type")
				}
				if game != nil {
					t.Error("Expected nil game for invalid type")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if game == nil {
				t.Error("Expected non-nil game")
			}
		})
	}
}

// TestCreateGameImplementsInterface verifies created games implement MiniGame.
func TestCreateGameImplementsInterface(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	gameTypes := []engine.MiniGameType{
		engine.MiniGameCard,
		engine.MiniGameDice,
		engine.MiniGamePuzzle,
		engine.MiniGameMemory,
		engine.MiniGameLockPicking,
		engine.MiniGameHacking,
		engine.MiniGameRitual,
	}

	for _, gameType := range gameTypes {
		t.Run(gameType.String(), func(t *testing.T) {
			game, err := sys.CreateGame(gameType)
			if err != nil {
				t.Fatalf("Failed to create game: %v", err)
			}

			// Verify interface methods exist (will panic if not)
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Errorf("Initialize failed: %v", err)
			}
			_ = game.IsComplete()
			_ = game.GetReward()
		})
	}
}

// TestGetAvailableGames verifies all 7 game types are returned.
func TestGetAvailableGames(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	games := sys.GetAvailableGames()

	expectedCount := 7
	if len(games) != expectedCount {
		t.Errorf("Expected %d games, got %d", expectedCount, len(games))
	}

	// Verify all expected types are present
	expectedTypes := map[engine.MiniGameType]bool{
		engine.MiniGameCard:        true,
		engine.MiniGameDice:        true,
		engine.MiniGamePuzzle:      true,
		engine.MiniGameMemory:      true,
		engine.MiniGameLockPicking: true,
		engine.MiniGameHacking:     true,
		engine.MiniGameRitual:      true,
	}

	for _, gameType := range games {
		if !expectedTypes[gameType] {
			t.Errorf("Unexpected game type: %v", gameType)
		}
		delete(expectedTypes, gameType)
	}

	if len(expectedTypes) > 0 {
		t.Errorf("Missing game types: %v", expectedTypes)
	}
}

// TestGameDeterminism verifies games produce consistent results with same seed.
func TestGameDeterminism(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	const testSeed = int64(42)
	const difficulty = 0.5

	gameTypes := []engine.MiniGameType{
		engine.MiniGameCard,
		engine.MiniGameDice,
		engine.MiniGamePuzzle,
		engine.MiniGameMemory,
		engine.MiniGameLockPicking,
		engine.MiniGameHacking,
		engine.MiniGameRitual,
	}

	for _, gameType := range gameTypes {
		t.Run(gameType.String(), func(t *testing.T) {
			// Create and initialize two instances with same seed
			game1, err := sys.CreateGame(gameType)
			if err != nil {
				t.Fatalf("Failed to create game1: %v", err)
			}
			if err := game1.Initialize(testSeed, difficulty); err != nil {
				t.Fatalf("Failed to initialize game1: %v", err)
			}

			game2, err := sys.CreateGame(gameType)
			if err != nil {
				t.Fatalf("Failed to create game2: %v", err)
			}
			if err := game2.Initialize(testSeed, difficulty); err != nil {
				t.Fatalf("Failed to initialize game2: %v", err)
			}

			// Update both games for a few frames
			for i := 0; i < 5; i++ {
				if err := game1.Update(0.016); err != nil {
					t.Fatalf("game1.Update failed: %v", err)
				}
				if err := game2.Update(0.016); err != nil {
					t.Fatalf("game2.Update failed: %v", err)
				}

				// Both should be in same state
				if game1.IsComplete() != game2.IsComplete() {
					t.Errorf("Frame %d: IsComplete mismatch", i)
				}
			}
		})
	}
}

// TestGameDifficultyScaling verifies games respect difficulty parameter.
func TestGameDifficultyScaling(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	const testSeed = int64(12345)

	// Test with easy and hard difficulties
	difficulties := []struct {
		name  string
		value float64
	}{
		{"Easy", 0.0},
		{"Medium", 0.5},
		{"Hard", 1.0},
	}

	for _, gameType := range sys.GetAvailableGames() {
		t.Run(gameType.String(), func(t *testing.T) {
			for _, diff := range difficulties {
				t.Run(diff.name, func(t *testing.T) {
					game, err := sys.CreateGame(gameType)
					if err != nil {
						t.Fatalf("Failed to create game: %v", err)
					}

					// Should accept valid difficulty values
					if err := game.Initialize(testSeed, diff.value); err != nil {
						t.Errorf("Initialize failed with difficulty %.1f: %v", diff.value, err)
					}
				})
			}
		})
	}
}

// TestGameInvalidDifficulty verifies games reject invalid difficulty values.
func TestGameInvalidDifficulty(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	const testSeed = int64(12345)

	invalidDifficulties := []float64{-0.1, -1.0, 1.1, 2.0}

	for _, gameType := range sys.GetAvailableGames() {
		t.Run(gameType.String(), func(t *testing.T) {
			for _, diff := range invalidDifficulties {
				game, err := sys.CreateGame(gameType)
				if err != nil {
					t.Fatalf("Failed to create game: %v", err)
				}

				// Should reject invalid difficulty
				if err := game.Initialize(testSeed, diff); err == nil {
					t.Errorf("Expected error for difficulty %.1f, got nil", diff)
				}
			}
		})
	}
}

// BenchmarkCreateGame benchmarks game creation performance.
func BenchmarkCreateGame(b *testing.B) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sys.CreateGame(engine.MiniGameCard)
	}
}

// BenchmarkGetAvailableGames benchmarks listing available games.
func BenchmarkGetAvailableGames(b *testing.B) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GetAvailableGames()
	}
}
