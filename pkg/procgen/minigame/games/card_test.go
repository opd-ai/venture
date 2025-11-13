package games

import (
	"testing"
)

func TestCardGame_Initialize(t *testing.T) {
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
			game := NewCardGame()
			err := game.Initialize(tt.seed, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify game state
				if game.deckSize <= 0 {
					t.Error("deckSize should be positive")
				}
				if game.handSize <= 0 {
					t.Error("handSize should be positive")
				}
				if game.targetWins <= 0 {
					t.Error("targetWins should be positive")
				}
				if len(game.playerHand) != game.handSize {
					t.Errorf("playerHand size = %d, want %d", len(game.playerHand), game.handSize)
				}
				if len(game.opponentHand) != game.handSize {
					t.Errorf("opponentHand size = %d, want %d", len(game.opponentHand), game.handSize)
				}
				if game.completed {
					t.Error("game should not be completed initially")
				}
			}
		})
	}
}

func TestCardGame_Determinism(t *testing.T) {
	seed := int64(42)
	difficulty := 0.5

	// Create two games with same seed
	game1 := NewCardGame()
	game2 := NewCardGame()

	if err := game1.Initialize(seed, difficulty); err != nil {
		t.Fatalf("game1 Initialize() failed: %v", err)
	}
	if err := game2.Initialize(seed, difficulty); err != nil {
		t.Fatalf("game2 Initialize() failed: %v", err)
	}

	// Verify identical initial state
	if game1.deckSize != game2.deckSize {
		t.Errorf("deckSize mismatch: %d != %d", game1.deckSize, game2.deckSize)
	}
	if game1.handSize != game2.handSize {
		t.Errorf("handSize mismatch: %d != %d", game1.handSize, game2.handSize)
	}
	if len(game1.playerHand) != len(game2.playerHand) {
		t.Errorf("playerHand length mismatch: %d != %d", len(game1.playerHand), len(game2.playerHand))
	}

	// Verify identical hands
	for i := range game1.playerHand {
		if game1.playerHand[i] != game2.playerHand[i] {
			t.Errorf("playerHand[%d] mismatch: %d != %d", i, game1.playerHand[i], game2.playerHand[i])
		}
	}
}

func TestCardGame_Update(t *testing.T) {
	game := NewCardGame()
	if err := game.Initialize(12345, 0.3); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Play game until completion
	maxRounds := 100 // Prevent infinite loop
	for i := 0; i < maxRounds && !game.IsComplete(); i++ {
		if err := game.Update(1.0); err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
	}

	if !game.IsComplete() {
		t.Error("game should be complete after max rounds")
	}

	// Verify winner is determined
	if game.playerWins < game.targetWins && game.opponentWins < game.targetWins {
		t.Error("neither player reached target wins")
	}
}

func TestCardGame_IsComplete(t *testing.T) {
	game := NewCardGame()
	if err := game.Initialize(12345, 0.5); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if game.IsComplete() {
		t.Error("game should not be complete initially")
	}

	// Play until complete
	for i := 0; i < 100 && !game.IsComplete(); i++ {
		game.Update(1.0)
	}

	if !game.IsComplete() {
		t.Error("game should be complete")
	}
}

func TestCardGame_GetReward(t *testing.T) {
	tests := []struct {
		name       string
		difficulty float64
		forceWin   bool
		wantReward bool
	}{
		{"easy win", 0.0, true, true},
		{"medium win", 0.5, true, true},
		{"hard win", 1.0, true, true},
		{"loss", 0.5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewCardGame()
			if err := game.Initialize(12345, tt.difficulty); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}

			// Force game completion
			game.completed = true
			game.playerWon = tt.forceWin

			reward := game.GetReward()
			if tt.wantReward {
				if reward == nil {
					t.Error("expected reward, got nil")
					return
				}
				if reward.Gold <= 0 {
					t.Error("reward Gold should be positive")
				}
				if reward.XP <= 0 {
					t.Error("reward XP should be positive")
				}
			} else {
				if reward != nil {
					t.Error("expected no reward, got reward")
				}
			}
		})
	}
}

func TestCardGame_DifficultyScaling(t *testing.T) {
	difficulties := []float64{0.0, 0.3, 0.5, 0.7, 1.0}
	seed := int64(42)

	prevDeckSize := 0
	prevHandSize := 0
	prevTargetWins := 0

	for _, diff := range difficulties {
		game := NewCardGame()
		if err := game.Initialize(seed, diff); err != nil {
			t.Fatalf("Initialize(%.1f) failed: %v", diff, err)
		}

		// Verify parameters increase with difficulty
		if game.deckSize < prevDeckSize {
			t.Errorf("deckSize should increase with difficulty: %d < %d at %.1f", game.deckSize, prevDeckSize, diff)
		}
		if game.handSize < prevHandSize {
			t.Errorf("handSize should increase with difficulty: %d < %d at %.1f", game.handSize, prevHandSize, diff)
		}
		if game.targetWins < prevTargetWins {
			t.Errorf("targetWins should increase with difficulty: %d < %d at %.1f", game.targetWins, prevTargetWins, diff)
		}

		prevDeckSize = game.deckSize
		prevHandSize = game.handSize
		prevTargetWins = game.targetWins
	}
}

func BenchmarkCardGame_Initialize(b *testing.B) {
	game := NewCardGame()
	for i := 0; i < b.N; i++ {
		game.Initialize(int64(i), 0.5)
	}
}

func BenchmarkCardGame_Update(b *testing.B) {
	game := NewCardGame()
	game.Initialize(12345, 0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if game.IsComplete() {
			game.Initialize(int64(i), 0.5)
		}
		game.Update(1.0)
	}
}
