package games

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// BenchmarkUpdate_AllGames benchmarks the Update method for all game types
// Target: <0.1ms (100μs) per call per doc.go:51
func BenchmarkUpdate_AllGames(b *testing.B) {
	tests := []struct {
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

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tt.game.Initialize(12345, 0.5)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tt.game.Update(0.016) // 60 FPS
			}
		})
	}
}

// BenchmarkPrepareRender_AllGames benchmarks the PrepareRender method for all game types
// Target: <0.1ms (100μs) per call per doc.go:52
func BenchmarkPrepareRender_AllGames(b *testing.B) {
	tests := []struct {
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

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tt.game.Initialize(12345, 0.5)
			// Run a few updates to get into a stable state
			for i := 0; i < 10; i++ {
				tt.game.Update(0.016)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tt.game.PrepareRender(800, 600)
			}
		})
	}
}

// BenchmarkInitialize_AllGames benchmarks the Initialize method for all game types
// Target: <1ms per game per doc.go:50
func BenchmarkInitialize_AllGames(b *testing.B) {
	tests := []struct {
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

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tt.game.Initialize(int64(i), 0.5)
			}
		})
	}
}

// BenchmarkUpdate_DiceGame benchmarks the DiceGame Update method specifically
func BenchmarkUpdate_DiceGame(b *testing.B) {
	game := NewDiceGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkUpdate_PuzzleGame benchmarks the PuzzleGame Update method specifically
func BenchmarkUpdate_PuzzleGame(b *testing.B) {
	game := NewPuzzleGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkUpdate_MemoryGame benchmarks the MemoryGame Update method specifically
func BenchmarkUpdate_MemoryGame(b *testing.B) {
	game := NewMemoryGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkUpdate_LockPickingGame benchmarks the LockPickingGame Update method specifically
func BenchmarkUpdate_LockPickingGame(b *testing.B) {
	game := NewLockPickingGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkUpdate_HackingGame benchmarks the HackingGame Update method specifically
func BenchmarkUpdate_HackingGame(b *testing.B) {
	game := NewHackingGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkUpdate_RitualGame benchmarks the RitualGame Update method specifically
func BenchmarkUpdate_RitualGame(b *testing.B) {
	game := NewRitualGame()
	game.Initialize(12345, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Update(0.016)
	}
}

// BenchmarkPrepareRender_DiceGame benchmarks the DiceGame PrepareRender method specifically
func BenchmarkPrepareRender_DiceGame(b *testing.B) {
	game := NewDiceGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}

// BenchmarkPrepareRender_PuzzleGame benchmarks the PuzzleGame PrepareRender method specifically
func BenchmarkPrepareRender_PuzzleGame(b *testing.B) {
	game := NewPuzzleGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}

// BenchmarkPrepareRender_MemoryGame benchmarks the MemoryGame PrepareRender method specifically
func BenchmarkPrepareRender_MemoryGame(b *testing.B) {
	game := NewMemoryGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}

// BenchmarkPrepareRender_LockPickingGame benchmarks the LockPickingGame PrepareRender method specifically
func BenchmarkPrepareRender_LockPickingGame(b *testing.B) {
	game := NewLockPickingGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}

// BenchmarkPrepareRender_HackingGame benchmarks the HackingGame PrepareRender method specifically
func BenchmarkPrepareRender_HackingGame(b *testing.B) {
	game := NewHackingGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}

// BenchmarkPrepareRender_RitualGame benchmarks the RitualGame PrepareRender method specifically
func BenchmarkPrepareRender_RitualGame(b *testing.B) {
	game := NewRitualGame()
	game.Initialize(12345, 0.5)
	for i := 0; i < 10; i++ {
		game.Update(0.016)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.PrepareRender(800, 600)
	}
}
