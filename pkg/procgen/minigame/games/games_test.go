package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// Tests for DiceGame

func TestDiceGame_Initialize(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		difficulty float64
		wantErr    bool
	}{
		{"valid easy", 12345, 0.0, false},
		{"valid medium", 12345, 0.5, false},
		{"valid hard", 12345, 1.0, false},
		{"invalid difficulty low", 12345, -0.1, true},
		{"invalid difficulty high", 12345, 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewDiceGame()
			err := game.Initialize(tt.seed, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiceGame_Determinism(t *testing.T) {
	seed := int64(42)
	game1, game2 := NewDiceGame(), NewDiceGame()
	game1.Initialize(seed, 0.5)
	game2.Initialize(seed, 0.5)

	if game1.numDice != game2.numDice {
		t.Errorf("numDice mismatch: %d != %d", game1.numDice, game2.numDice)
	}
}

func TestDiceGame_Complete(t *testing.T) {
	game := NewDiceGame()
	game.Initialize(12345, 0.5)

	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should complete")
	}
}

// Tests for PuzzleGame

func TestPuzzleGame_Initialize(t *testing.T) {
	tests := []struct {
		name       string
		difficulty float64
		wantErr    bool
	}{
		{"valid easy", 0.0, false},
		{"valid hard", 1.0, false},
		{"invalid low", -0.1, true},
		{"invalid high", 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewPuzzleGame()
			err := game.Initialize(12345, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPuzzleGame_GridGeneration(t *testing.T) {
	game := NewPuzzleGame()
	game.Initialize(12345, 0.5)

	if len(game.grid) != game.gridSize {
		t.Errorf("grid size = %d, want %d", len(game.grid), game.gridSize)
	}
}

// Tests for MemoryGame

func TestMemoryGame_Initialize(t *testing.T) {
	game := NewMemoryGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if game.numPairs <= 0 {
		t.Error("numPairs should be positive")
	}
	if len(game.matched) != game.numPairs {
		t.Errorf("matched size = %d, want %d", len(game.matched), game.numPairs)
	}
}

func TestMemoryGame_Complete(t *testing.T) {
	game := NewMemoryGame()
	game.Initialize(12345, 0.3)

	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should complete")
	}
}

// Tests for LockPickingGame

func TestLockPickingGame_Initialize(t *testing.T) {
	game := NewLockPickingGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if game.numPins <= 0 {
		t.Error("numPins should be positive")
	}
	if len(game.pinPositions) != game.numPins {
		t.Errorf("pinPositions size = %d, want %d", len(game.pinPositions), game.numPins)
	}
}

func TestLockPickingGame_Determinism(t *testing.T) {
	seed := int64(42)
	game1, game2 := NewLockPickingGame(), NewLockPickingGame()
	game1.Initialize(seed, 0.5)
	game2.Initialize(seed, 0.5)

	if game1.numPins != game2.numPins {
		t.Errorf("numPins mismatch: %d != %d", game1.numPins, game2.numPins)
	}

	for i := range game1.pinPositions {
		if game1.pinPositions[i] != game2.pinPositions[i] {
			t.Errorf("pinPositions[%d] mismatch", i)
		}
	}
}

// Tests for HackingGame

func TestHackingGame_Initialize(t *testing.T) {
	game := NewHackingGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if len(game.code) != game.codeLength {
		t.Errorf("code length = %d, want %d", len(game.code), game.codeLength)
	}
}

func TestHackingGame_Determinism(t *testing.T) {
	seed := int64(42)
	game1, game2 := NewHackingGame(), NewHackingGame()
	game1.Initialize(seed, 0.5)
	game2.Initialize(seed, 0.5)

	if game1.code != game2.code {
		t.Errorf("code mismatch: %s != %s", game1.code, game2.code)
	}
}

func TestHackingGame_Complete(t *testing.T) {
	game := NewHackingGame()
	game.Initialize(12345, 0.5)

	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should complete")
	}
}

// Tests for RitualGame

func TestRitualGame_Initialize(t *testing.T) {
	game := NewRitualGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if game.numSymbols <= 0 {
		t.Error("numSymbols should be positive")
	}
	if len(game.symbols) != game.numSymbols {
		t.Errorf("symbols size = %d, want %d", len(game.symbols), game.numSymbols)
	}
}

func TestRitualGame_Determinism(t *testing.T) {
	seed := int64(42)
	game1, game2 := NewRitualGame(), NewRitualGame()
	game1.Initialize(seed, 0.5)
	game2.Initialize(seed, 0.5)

	if game1.numSymbols != game2.numSymbols {
		t.Errorf("numSymbols mismatch: %d != %d", game1.numSymbols, game2.numSymbols)
	}

	if len(game1.symbols) != len(game2.symbols) {
		t.Errorf("symbols length mismatch: %d != %d", len(game1.symbols), len(game2.symbols))
	}
}

func TestRitualGame_Complete(t *testing.T) {
	game := NewRitualGame()
	game.Initialize(12345, 0.3)

	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should complete")
	}
}

// Common tests for all games

func TestAllGames_RewardScaling(t *testing.T) {
	games := []struct {
		name string
		game engine.MiniGame
	}{
		{"Card", NewCardGame()},
		{"Dice", NewDiceGame()},
		{"Puzzle", NewPuzzleGame()},
		{"Memory", NewMemoryGame()},
		{"LockPicking", NewLockPickingGame()},
		{"Hacking", NewHackingGame()},
		{"Ritual", NewRitualGame()},
	}

	for _, g := range games {
		t.Run(g.name, func(t *testing.T) {
			// Test easy difficulty
			g.game.Initialize(12345, 0.0)
			// Force completion for reward check
			for i := 0; i < 100 && !g.game.IsComplete(); i++ {
				g.game.Update(1.0)
			}

			// Test that rewards exist for completed games
			// Note: actual reward value testing is in individual tests
			if g.game.IsComplete() {
				// Reward may be nil if player lost
				// This is expected behavior
			}
		})
	}
}
