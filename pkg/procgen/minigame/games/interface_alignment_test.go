// Package games contains implementations of mini-game types for Venture.
// This file tests the MiniGame interface alignment with ECS rendering pattern.
package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestMiniGameInterfaceCompliance verifies all game types implement the updated MiniGame interface.
// This test ensures PrepareRender and GetRenderOutput methods are properly implemented.
func TestMiniGameInterfaceCompliance(t *testing.T) {
	tests := []struct {
		name     string
		gameType engine.MiniGameType
	}{
		{"CardGame", engine.MiniGameCard},
		{"DiceGame", engine.MiniGameDice},
		{"PuzzleGame", engine.MiniGamePuzzle},
		{"MemoryGame", engine.MiniGameMemory},
		{"LockPickingGame", engine.MiniGameLockPicking},
		{"HackingGame", engine.MiniGameHacking},
		{"RitualGame", engine.MiniGameRitual},
	}

	world := engine.NewWorld()
	sys := NewSystem(world)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := sys.CreateGame(tt.gameType)
			if err != nil {
				t.Fatalf("CreateGame() error = %v", err)
			}

			// Initialize the game
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			// Test PrepareRender with valid dimensions
			if err := game.PrepareRender(800, 600); err != nil {
				t.Errorf("PrepareRender() error = %v, want nil", err)
			}

			// Test GetRenderOutput returns non-nil after PrepareRender
			output := game.GetRenderOutput()
			if output == nil {
				t.Error("GetRenderOutput() = nil, want non-nil after PrepareRender")
			}

			// Verify the output implements the interface properly
			if output != nil {
				title := output.GetTitle()
				if title == "" {
					t.Error("GetTitle() = empty string, want non-empty")
				}

				status := output.GetStatus()
				if status == "" {
					t.Error("GetStatus() = empty string, want non-empty")
				}

				w, h := output.GetDimensions()
				if w != 800 || h != 600 {
					t.Errorf("GetDimensions() = (%d, %d), want (800, 600)", w, h)
				}

				elements := output.GetElements()
				if elements == nil {
					t.Error("GetElements() = nil, want non-nil")
				}
			}
		})
	}
}

// TestPrepareRenderValidation verifies PrepareRender validates screen dimensions properly.
func TestPrepareRenderValidation(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"ValidDimensions", 800, 600, false},
		{"ZeroWidth", 0, 600, true},
		{"ZeroHeight", 800, 0, true},
		{"NegativeWidth", -800, 600, true},
		{"NegativeHeight", 800, -600, true},
		{"BothZero", 0, 0, true},
	}

	world := engine.NewWorld()
	sys := NewSystem(world)

	// Test with one game type (all games share same validation logic)
	game, err := sys.CreateGame(engine.MiniGameCard)
	if err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}

	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := game.PrepareRender(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareRender(%d, %d) error = %v, wantErr %v", tt.width, tt.height, err, tt.wantErr)
			}
		})
	}
}

// TestGetRenderOutputBeforePrepare verifies GetRenderOutput returns nil before PrepareRender is called.
func TestGetRenderOutputBeforePrepare(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	game, err := sys.CreateGame(engine.MiniGameCard)
	if err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}

	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// GetRenderOutput should return nil before PrepareRender
	output := game.GetRenderOutput()
	if output != nil {
		t.Error("GetRenderOutput() before PrepareRender = non-nil, want nil")
	}
}

// TestBackwardCompatibilityRender verifies the deprecated Render method still works on concrete types.
func TestBackwardCompatibilityRender(t *testing.T) {
	// Test with concrete type to access deprecated Render method
	game := NewMemoryGame()

	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Create a stub screen from existing render_test.go
	screen := &stubScreen{width: 800, height: 600}

	// Old Render method should still work for backward compatibility
	if err := game.Render(screen); err != nil {
		t.Errorf("Render() error = %v, want nil (backward compatibility)", err)
	}

	// Verify it populated the render output
	output := game.GetRenderOutput()
	if output == nil {
		t.Error("GetRenderOutput() after Render() = nil, want non-nil")
	}
}

// TestRenderOutputDataConsistency verifies PrepareRender and GetRenderOutput produce consistent data.
func TestRenderOutputDataConsistency(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	games := []engine.MiniGameType{
		engine.MiniGameCard,
		engine.MiniGameDice,
		engine.MiniGamePuzzle,
		engine.MiniGameMemory,
		engine.MiniGameLockPicking,
		engine.MiniGameHacking,
		engine.MiniGameRitual,
	}

	for _, gameType := range games {
		t.Run(gameType.String(), func(t *testing.T) {
			game, err := sys.CreateGame(gameType)
			if err != nil {
				t.Fatalf("CreateGame() error = %v", err)
			}

			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			// Prepare render multiple times to ensure consistency
			for i := 0; i < 3; i++ {
				if err := game.PrepareRender(800, 600); err != nil {
					t.Fatalf("PrepareRender() iteration %d error = %v", i, err)
				}

				output := game.GetRenderOutput()
				if output == nil {
					t.Fatalf("GetRenderOutput() iteration %d = nil", i)
				}

				// Verify dimensions match input
				w, h := output.GetDimensions()
				if w != 800 || h != 600 {
					t.Errorf("GetDimensions() iteration %d = (%d, %d), want (800, 600)", i, w, h)
				}

				// Verify status is valid
				status := output.GetStatus()
				validStatuses := map[string]bool{"Playing": true, "Won": true, "Lost": true}
				if !validStatuses[status] {
					t.Errorf("GetStatus() iteration %d = %q, want one of %v", i, status, []string{"Playing", "Won", "Lost"})
				}
			}
		})
	}
}
