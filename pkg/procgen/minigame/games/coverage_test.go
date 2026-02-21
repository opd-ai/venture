// Package games contains additional coverage tests for minigame implementations.
// These tests address audit items for improved test coverage of edge cases and
// state transitions identified in AUDIT.md (2026-02-21).
package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestSystemUpdate_NoOp verifies System.Update is explicitly a no-op.
// This documents the intent that minigame systems delegate updates to MiniGameSystem.
func TestSystemUpdate_NoOp(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(world)

	// Create some entities to pass to Update
	entities := []*engine.Entity{
		engine.NewEntity(1),
		engine.NewEntity(2),
	}

	// Update should not panic or error with nil entities
	sys.Update(nil, 0.016)

	// Update should not panic or error with empty entities
	sys.Update([]*engine.Entity{}, 0.016)

	// Update should not panic or error with actual entities
	sys.Update(entities, 0.016)

	// Update should work with various delta times
	sys.Update(entities, 0.0)
	sys.Update(entities, 1.0)
	sys.Update(entities, 100.0)
}

// TestGetRenderOutput_NilBeforePrepare verifies GetRenderOutput returns nil
// before PrepareRender has been called. This is expected behavior per the
// MiniGame interface contract.
func TestGetRenderOutput_NilBeforePrepare(t *testing.T) {
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
			// Initialize the game (required for normal operation)
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// GetRenderOutput should return nil before any PrepareRender call
			output := g.game.GetRenderOutput()
			if output != nil {
				t.Errorf("GetRenderOutput() before PrepareRender = %v, want nil", output)
			}
		})
	}
}

// TestGetRenderOutput_NilWithoutInit verifies GetRenderOutput returns nil
// for uninitialized games.
func TestGetRenderOutput_NilWithoutInit(t *testing.T) {
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
			// Don't initialize - GetRenderOutput should still return nil safely
			output := g.game.GetRenderOutput()
			if output != nil {
				t.Errorf("GetRenderOutput() without Initialize = %v, want nil", output)
			}
		})
	}
}

// TestMemoryGame_DetermineGameStatus verifies all state transitions for
// the determineGameStatus helper function in memory.go.
func TestMemoryGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
		// Edge case: completed with playerWon flag set but should use playerWon value
		{"playing_won_flag_ignored", false, true, "Playing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewMemoryGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// Set game state directly for testing
			game.completed = tt.completed
			game.playerWon = tt.playerWon

			// Call PrepareRender to trigger determineGameStatus
			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestCardGame_DetermineGameStatus verifies all state transitions for card game.
func TestCardGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewCardGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestDiceGame_DetermineGameStatus verifies all state transitions for dice game.
func TestDiceGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewDiceGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestPuzzleGame_DetermineGameStatus verifies all state transitions for puzzle game.
func TestPuzzleGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewPuzzleGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestLockPickingGame_DetermineGameStatus verifies all state transitions.
func TestLockPickingGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewLockPickingGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestHackingGame_DetermineGameStatus verifies all state transitions.
func TestHackingGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewHackingGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestRitualGame_DetermineGameStatus verifies all state transitions.
func TestRitualGame_DetermineGameStatus(t *testing.T) {
	tests := []struct {
		name       string
		completed  bool
		playerWon  bool
		wantStatus string
	}{
		{"playing", false, false, "Playing"},
		{"won", true, true, "Won"},
		{"lost", true, false, "Lost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewRitualGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			game.completed = tt.completed
			game.playerWon = tt.playerWon

			if err := game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestMemoryGame_AttemptExhaustion verifies game ends when attempts are exhausted.
func TestMemoryGame_AttemptExhaustion(t *testing.T) {
	game := NewMemoryGame()
	// Use high difficulty for more attempts to test
	if err := game.Initialize(12345, 0.0); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Run updates until game completes (either win or loss)
	for i := 0; i < 200 && !game.IsComplete(); i++ {
		if err := game.Update(1.0); err != nil {
			t.Fatalf("Update() failed at iteration %d: %v", i, err)
		}
	}

	if !game.IsComplete() {
		t.Error("game should complete after sufficient updates")
	}

	// Verify status is either Won or Lost
	if err := game.PrepareRender(640, 480); err != nil {
		t.Fatalf("PrepareRender() failed: %v", err)
	}

	if game.LastRender.Status != "Won" && game.LastRender.Status != "Lost" {
		t.Errorf("unexpected final status: %q", game.LastRender.Status)
	}
}

// TestGetRenderOutput_AfterPrepareRender verifies GetRenderOutput returns
// non-nil after a successful PrepareRender call.
func TestGetRenderOutput_AfterPrepareRender(t *testing.T) {
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

			// GetRenderOutput should be nil before PrepareRender
			if output := g.game.GetRenderOutput(); output != nil {
				t.Errorf("GetRenderOutput() before PrepareRender = %v, want nil", output)
			}

			// Call PrepareRender
			if err := g.game.PrepareRender(640, 480); err != nil {
				t.Fatalf("PrepareRender() failed: %v", err)
			}

			// GetRenderOutput should now return non-nil
			output := g.game.GetRenderOutput()
			if output == nil {
				t.Error("GetRenderOutput() after PrepareRender = nil, want non-nil")
			}

			// Verify output has expected properties
			if output.GetTitle() == "" {
				t.Error("GetTitle() returned empty string")
			}
			if output.GetStatus() == "" {
				t.Error("GetStatus() returned empty string")
			}
		})
	}
}
