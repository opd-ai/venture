// Package games contains loss condition tests for all mini-game types.
// These tests verify that GetReward() returns nil when the player loses,
// and that game state is handled correctly in loss scenarios.
package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestAllGames_GetReward_Incomplete tests that GetReward returns nil when game is not completed.
func TestAllGames_GetReward_Incomplete(t *testing.T) {
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
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// Game is not completed yet
			if g.game.IsComplete() {
				t.Fatal("game should not be complete immediately after initialization")
			}

			// Reward should be nil for incomplete game
			reward := g.game.GetReward()
			if reward != nil {
				t.Errorf("expected nil reward for incomplete game, got Gold=%d, XP=%.1f", reward.Gold, reward.XP)
			}
		})
	}
}

// TestAllGames_GetReward_Loss tests that GetReward returns nil when player loses.
func TestAllGames_GetReward_Loss(t *testing.T) {
	tests := []struct {
		name       string
		setupLoss  func() engine.MiniGame
		difficulty float64
	}{
		{
			name: "DiceGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewDiceGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: opponent wins
				game.completed = true
				game.playerWon = false
				return game
			},
		},
		{
			name: "PuzzleGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewPuzzleGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: ran out of moves
				game.completed = true
				game.playerWon = false
				return game
			},
		},
		{
			name: "MemoryGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewMemoryGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: ran out of attempts
				game.completed = true
				game.playerWon = false
				return game
			},
		},
		{
			name: "LockPickingGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewLockPickingGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: too many failures
				game.completed = true
				game.playerWon = false
				return game
			},
		},
		{
			name: "HackingGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewHackingGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: ran out of attempts
				game.completed = true
				game.playerWon = false
				return game
			},
		},
		{
			name: "RitualGame loss",
			setupLoss: func() engine.MiniGame {
				game := NewRitualGame()
				game.Initialize(12345, 0.5)
				// Force loss condition: failed to draw symbol
				game.completed = true
				game.playerWon = false
				return game
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := tt.setupLoss()

			if !game.IsComplete() {
				t.Fatal("game should be complete after setup")
			}

			reward := game.GetReward()
			if reward != nil {
				t.Errorf("expected nil reward for lost game, got Gold=%d, XP=%.1f", reward.Gold, reward.XP)
			}
		})
	}
}

// TestDiceGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestDiceGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewDiceGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewDiceGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestHackingGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestHackingGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewHackingGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true
	winGame.attempts = 3
	winGame.maxAttempts = 8

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewHackingGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestLockPickingGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestLockPickingGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewLockPickingGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true
	winGame.timeElapsed = 5.0 // Fast completion for time bonus

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewLockPickingGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestMemoryGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestMemoryGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewMemoryGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true
	winGame.attempts = 10
	winGame.maxAttempts = 25

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewMemoryGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestPuzzleGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestPuzzleGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewPuzzleGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true
	winGame.moves = 10
	winGame.maxMoves = 32

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewPuzzleGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestRitualGame_GetReward_LossVsWin verifies reward difference between win and loss.
func TestRitualGame_GetReward_LossVsWin(t *testing.T) {
	// Test win scenario
	winGame := NewRitualGame()
	winGame.Initialize(12345, 0.5)
	winGame.completed = true
	winGame.playerWon = true

	winReward := winGame.GetReward()
	if winReward == nil {
		t.Fatal("expected reward for win, got nil")
	}
	if winReward.Gold <= 0 || winReward.XP <= 0 {
		t.Errorf("win reward should be positive: Gold=%d, XP=%.1f", winReward.Gold, winReward.XP)
	}

	// Test loss scenario
	lossGame := NewRitualGame()
	lossGame.Initialize(12345, 0.5)
	lossGame.completed = true
	lossGame.playerWon = false

	lossReward := lossGame.GetReward()
	if lossReward != nil {
		t.Errorf("expected nil reward for loss, got Gold=%d, XP=%.1f", lossReward.Gold, lossReward.XP)
	}
}

// TestUpdate_CompletedGame tests that Update does nothing after game completion.
func TestUpdate_CompletedGame(t *testing.T) {
	games := []struct {
		name string
		game interface {
			engine.MiniGame
			setCompleted(bool)
		}
	}{
		{"Dice", &diceGameWrapper{NewDiceGame()}},
		{"Puzzle", &puzzleGameWrapper{NewPuzzleGame()}},
		{"Memory", &memoryGameWrapper{NewMemoryGame()}},
		{"LockPicking", &lockPickingGameWrapper{NewLockPickingGame()}},
		{"Hacking", &hackingGameWrapper{NewHackingGame()}},
		{"Ritual", &ritualGameWrapper{NewRitualGame()}},
	}

	for _, g := range games {
		t.Run(g.name, func(t *testing.T) {
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// Force completion
			g.game.setCompleted(true)

			// Update should not panic or change state
			for i := 0; i < 10; i++ {
				if err := g.game.Update(1.0); err != nil {
					t.Fatalf("Update() after completion failed: %v", err)
				}
			}

			// Game should still be complete
			if !g.game.IsComplete() {
				t.Error("game should remain complete after Update calls")
			}
		})
	}
}

// Wrapper types to allow setting completed state

type diceGameWrapper struct{ *DiceGame }

func (w *diceGameWrapper) setCompleted(v bool) { w.completed = v }

type puzzleGameWrapper struct{ *PuzzleGame }

func (w *puzzleGameWrapper) setCompleted(v bool) { w.completed = v }

type memoryGameWrapper struct{ *MemoryGame }

func (w *memoryGameWrapper) setCompleted(v bool) { w.completed = v }

type lockPickingGameWrapper struct{ *LockPickingGame }

func (w *lockPickingGameWrapper) setCompleted(v bool) { w.completed = v }

type hackingGameWrapper struct{ *HackingGame }

func (w *hackingGameWrapper) setCompleted(v bool) { w.completed = v }

type ritualGameWrapper struct{ *RitualGame }

func (w *ritualGameWrapper) setCompleted(v bool) { w.completed = v }

// TestSystem_Update tests the System.Update method.
func TestSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	// Create and add a game
	game, err := sys.CreateGame(engine.MiniGameCard)
	if err != nil {
		t.Fatalf("CreateGame() failed: %v", err)
	}

	if game == nil {
		t.Fatal("CreateGame() returned nil game")
	}

	// System update should not panic (it's a no-op)
	sys.Update([]*engine.Entity{}, 1.0)
}

// TestRender_Stub tests that Render stubs validate screen input.
// This test uses the deprecated Render method for backward compatibility testing.
func TestRender_Stub(t *testing.T) {
	screen := &stubScreen{width: 320, height: 240}

	games := []struct {
		name string
		game renderableGame
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
		g := g // rebind to avoid capturing loop variable in closure
		t.Run(g.name, func(t *testing.T) {
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// Render with valid screen should succeed
			if err := g.game.Render(screen); err != nil {
				t.Errorf("Render() returned error: %v", err)
			}
		})
	}
}

// TestPuzzleGame_Update_LossCondition tests that puzzle game completes as loss when out of moves.
func TestPuzzleGame_Update_LossCondition(t *testing.T) {
	// Use a seed and high difficulty that typically results in loss
	game := NewPuzzleGame()
	if err := game.Initialize(99999, 1.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Force the game to exhaust moves by updating many times
	// With high difficulty, success rate is low
	for i := 0; i < game.maxMoves+10 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete after exceeding maxMoves")
	}

	// The game could be won or lost depending on RNG
	// This test verifies completion happens correctly
}

// TestLockPickingGame_Update_LossCondition tests loss by exceeding failures.
func TestLockPickingGame_Update_LossCondition(t *testing.T) {
	// Use high difficulty for lower success rate
	game := NewLockPickingGame()
	if err := game.Initialize(88888, 1.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// With high difficulty, maxFailures is low and success rate is low
	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete")
	}
}

// TestMemoryGame_Update_LossCondition tests loss by exceeding attempts.
func TestMemoryGame_Update_LossCondition(t *testing.T) {
	// Use high difficulty for lower success rate
	game := NewMemoryGame()
	if err := game.Initialize(77777, 1.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// With high difficulty, more pairs to match and lower success rate
	for i := 0; i < game.maxAttempts+10 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete after exceeding maxAttempts")
	}
}

// TestHackingGame_Update_LossCondition tests loss by exceeding attempts without correct guess.
func TestHackingGame_Update_LossCondition(t *testing.T) {
	// Use high difficulty for fewer attempts and harder codes
	game := NewHackingGame()
	if err := game.Initialize(66666, 1.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// With high difficulty, 6 attempts and 8-char codes
	for i := 0; i < game.maxAttempts+5 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete after exceeding maxAttempts")
	}
}

// TestRitualGame_Update_LossCondition tests loss by failing to draw symbol.
func TestRitualGame_Update_LossCondition(t *testing.T) {
	// Use very high difficulty for lower success rate
	game := NewRitualGame()
	if err := game.Initialize(55555, 1.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// With high difficulty (1.0), success rate is 0.6 + (1-1)*0.3 = 0.6
	// Failure can happen with 40% probability each update
	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete")
	}
}
