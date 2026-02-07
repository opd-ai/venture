// Package games contains tests for minigame Render implementations (Phase 27.3).
package games

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// stubScreen implements engine.ImageProvider for testing.
type stubScreen struct {
	width, height int
}

func (s *stubScreen) GetSize() (int, int)            { return s.width, s.height }
func (s *stubScreen) GetPixel(x, y int) color.Color  { return color.Transparent }

// Compile-time interface check.
var _ engine.ImageProvider = (*stubScreen)(nil)

// TestRender_NilScreen verifies all games return error for nil screen.
func TestRender_NilScreen(t *testing.T) {
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
		g := g // rebind to avoid capturing loop variable in closure
		t.Run(g.name, func(t *testing.T) {
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			err := g.game.Render(nil)
			if err == nil {
				t.Error("expected error for nil screen, got nil")
			}
		})
	}
}

// TestRender_Uninitialized verifies all games return error when not initialized.
func TestRender_Uninitialized(t *testing.T) {
	screen := &stubScreen{width: 320, height: 240}

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
		g := g // rebind to avoid capturing loop variable in closure
		t.Run(g.name, func(t *testing.T) {
			// Don't call Initialize
			err := g.game.Render(screen)
			if err == nil {
				t.Error("expected error for uninitialized game, got nil")
			}
		})
	}
}

// TestRender_ZeroDimensionScreen verifies error for zero-dimension screen.
func TestRender_ZeroDimensionScreen(t *testing.T) {
	screens := []struct {
		name   string
		screen *stubScreen
	}{
		{"zero width", &stubScreen{width: 0, height: 240}},
		{"zero height", &stubScreen{width: 320, height: 0}},
		{"both zero", &stubScreen{width: 0, height: 0}},
	}

	for _, s := range screens {
		s := s // rebind to avoid capturing loop variable in closure
		t.Run(s.name, func(t *testing.T) {
			game := NewCardGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}
			err := game.Render(s.screen)
			if err == nil {
				t.Error("expected error for invalid screen dimensions")
			}
		})
	}
}

// TestRender_ValidScreen verifies all games render successfully with valid screen.
func TestRender_ValidScreen(t *testing.T) {
	screen := &stubScreen{width: 320, height: 240}

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
		g := g // rebind to avoid capturing loop variable in closure
		t.Run(g.name, func(t *testing.T) {
			if err := g.game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			if err := g.game.Render(screen); err != nil {
				t.Errorf("Render() returned error: %v", err)
			}
		})
	}
}

// TestCardGame_RenderOutput verifies card game render output content.
func TestCardGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewCardGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Card Game" {
		t.Errorf("Title = %q, want %q", r.Title, "Card Game")
	}
	if r.Status != "Playing" {
		t.Errorf("Status = %q, want %q", r.Status, "Playing")
	}
	if r.Width != 640 || r.Height != 480 {
		t.Errorf("Dimensions = %dx%d, want 640x480", r.Width, r.Height)
	}
	if len(r.Elements) == 0 {
		t.Error("expected render elements, got none")
	}

	// Should have score text + player cards + opponent cards
	hasScore := false
	cardCount := 0
	for _, e := range r.Elements {
		if e.Type == "text" {
			hasScore = true
		}
		if e.Type == "card" {
			cardCount++
		}
	}
	if !hasScore {
		t.Error("missing score text element")
	}
	if cardCount != len(game.playerHand)+len(game.opponentHand) {
		t.Errorf("card elements = %d, want %d", cardCount, len(game.playerHand)+len(game.opponentHand))
	}
}

// TestDiceGame_RenderOutput verifies dice game render output content.
func TestDiceGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewDiceGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Dice Game" {
		t.Errorf("Title = %q, want %q", r.Title, "Dice Game")
	}

	dieCount := 0
	for _, e := range r.Elements {
		if e.Type == "die" {
			dieCount++
		}
	}
	if dieCount != game.numDice {
		t.Errorf("die elements = %d, want %d", dieCount, game.numDice)
	}
}

// TestPuzzleGame_RenderOutput verifies puzzle game render output content.
func TestPuzzleGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewPuzzleGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Puzzle" {
		t.Errorf("Title = %q, want %q", r.Title, "Puzzle")
	}

	tileCount := 0
	for _, e := range r.Elements {
		if e.Type == "tile" {
			tileCount++
		}
	}
	expectedTiles := game.gridSize * game.gridSize
	if tileCount != expectedTiles {
		t.Errorf("tile elements = %d, want %d", tileCount, expectedTiles)
	}
}

// TestMemoryGame_RenderOutput verifies memory game render output content.
func TestMemoryGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewMemoryGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Memory Game" {
		t.Errorf("Title = %q, want %q", r.Title, "Memory Game")
	}

	cardCount := 0
	for _, e := range r.Elements {
		if e.Type == "card" {
			cardCount++
		}
	}
	if cardCount != game.numPairs {
		t.Errorf("card elements = %d, want %d", cardCount, game.numPairs)
	}
}

// TestLockPickingGame_RenderOutput verifies lock-picking render output content.
func TestLockPickingGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewLockPickingGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Lock Picking" {
		t.Errorf("Title = %q, want %q", r.Title, "Lock Picking")
	}

	pinCount := 0
	for _, e := range r.Elements {
		if e.Type == "pin" {
			pinCount++
		}
	}
	if pinCount != game.numPins {
		t.Errorf("pin elements = %d, want %d", pinCount, game.numPins)
	}
}

// TestHackingGame_RenderOutput verifies hacking game render output content.
func TestHackingGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewHackingGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Run a few updates to generate guesses
	for i := 0; i < 3 && !game.IsComplete(); i++ {
		if err := game.Update(1.0); err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Hacking" {
		t.Errorf("Title = %q, want %q", r.Title, "Hacking")
	}

	termCount := 0
	for _, e := range r.Elements {
		if e.Type == "terminal" {
			termCount++
		}
	}
	if termCount != len(game.guesses) {
		t.Errorf("terminal elements = %d, want %d", termCount, len(game.guesses))
	}
}

// TestRitualGame_RenderOutput verifies ritual game render output content.
func TestRitualGame_RenderOutput(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewRitualGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	r := game.LastRender
	if r == nil {
		t.Fatal("LastRender is nil after Render()")
	}
	if r.Title != "Ritual" {
		t.Errorf("Title = %q, want %q", r.Title, "Ritual")
	}

	symbolCount := 0
	for _, e := range r.Elements {
		if e.Type == "symbol" {
			symbolCount++
		}
	}
	if symbolCount != game.numSymbols {
		t.Errorf("symbol elements = %d, want %d", symbolCount, game.numSymbols)
	}
}

// TestRender_CompletedGameStatus verifies render output shows correct status after completion.
func TestRender_CompletedGameStatus(t *testing.T) {
	screen := &stubScreen{width: 320, height: 240}

	tests := []struct {
		name      string
		setupGame func(t *testing.T) *CardGame
		wantStatus string
	}{
		{
			name: "won",
			setupGame: func(t *testing.T) *CardGame {
				t.Helper()
				g := NewCardGame()
				if err := g.Initialize(12345, 0.5); err != nil {
					t.Fatalf("Initialize() failed: %v", err)
				}
				g.completed = true
				g.playerWon = true
				return g
			},
			wantStatus: "Won",
		},
		{
			name: "lost",
			setupGame: func(t *testing.T) *CardGame {
				t.Helper()
				g := NewCardGame()
				if err := g.Initialize(12345, 0.5); err != nil {
					t.Fatalf("Initialize() failed: %v", err)
				}
				g.completed = true
				g.playerWon = false
				return g
			},
			wantStatus: "Lost",
		},
		{
			name: "playing",
			setupGame: func(t *testing.T) *CardGame {
				t.Helper()
				g := NewCardGame()
				if err := g.Initialize(12345, 0.5); err != nil {
					t.Fatalf("Initialize() failed: %v", err)
				}
				return g
			},
			wantStatus: "Playing",
		},
	}

	for _, tt := range tests {
		tt := tt // rebind to avoid capturing loop variable in closure
		t.Run(tt.name, func(t *testing.T) {
			game := tt.setupGame(t)
			if err := game.Render(screen); err != nil {
				t.Fatalf("Render() failed: %v", err)
			}
			if game.LastRender.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", game.LastRender.Status, tt.wantStatus)
			}
		})
	}
}

// TestRender_DifferentScreenSizes verifies rendering adapts to screen dimensions.
func TestRender_DifferentScreenSizes(t *testing.T) {
	screens := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 160, 120},
		{"medium", 640, 480},
		{"large", 1920, 1080},
	}

	for _, s := range screens {
		s := s // rebind to avoid capturing loop variable in closure
		t.Run(s.name, func(t *testing.T) {
			screen := &stubScreen{width: s.width, height: s.height}
			game := NewPuzzleGame()
			if err := game.Initialize(12345, 0.5); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			if err := game.Render(screen); err != nil {
				t.Fatalf("Render() failed: %v", err)
			}

			r := game.LastRender
			if r.Width != s.width || r.Height != s.height {
				t.Errorf("Dimensions = %dx%d, want %dx%d", r.Width, r.Height, s.width, s.height)
			}
		})
	}
}

// TestRender_AfterUpdates verifies render output reflects game state changes.
func TestRender_AfterUpdates(t *testing.T) {
	screen := &stubScreen{width: 640, height: 480}
	game := NewMemoryGame()
	if err := game.Initialize(12345, 0.3); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Render before any updates
	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() before updates failed: %v", err)
	}
	initialStatus := game.LastRender.Status

	// Run some updates
	for i := 0; i < 5 && !game.IsComplete(); i++ {
		if err := game.Update(1.0); err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	// Render after updates
	if err := game.Render(screen); err != nil {
		t.Fatalf("Render() after updates failed: %v", err)
	}

	// Status should still be valid
	if game.LastRender.Status != "Playing" && game.LastRender.Status != "Won" && game.LastRender.Status != "Lost" {
		t.Errorf("unexpected status after updates: %q", game.LastRender.Status)
	}

	// Initial status should have been "Playing"
	if initialStatus != "Playing" {
		t.Errorf("initial status = %q, want %q", initialStatus, "Playing")
	}
}

// BenchmarkRender benchmarks render performance across all game types.
func BenchmarkRender(b *testing.B) {
	screen := &stubScreen{width: 640, height: 480}

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
		g := g // rebind to avoid capturing loop variable in closure
		if err := g.game.Initialize(12345, 0.5); err != nil {
			b.Fatalf("Initialize() failed for %s: %v", g.name, err)
		}
		b.Run(g.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := g.game.Render(screen); err != nil {
					b.Fatalf("Render() failed for %s: %v", g.name, err)
				}
			}
		})
	}
}
