// Package engine provides tests for merchant gambling component (Phase 27.3).
package engine

import (
	"testing"
)

func TestMerchantGamblingComponent_Type(t *testing.T) {
	comp := NewMerchantGamblingComponent()
	if comp.Type() != "merchantGambling" {
		t.Errorf("expected type 'merchantGambling', got '%s'", comp.Type())
	}
}

func TestNewMerchantGamblingComponent(t *testing.T) {
	comp := NewMerchantGamblingComponent()

	if !comp.OffersGambling {
		t.Error("new merchant should offer gambling")
	}

	if len(comp.GamesOffered) == 0 {
		t.Error("new merchant should offer at least one game")
	}

	if comp.MinimumBet <= 0 {
		t.Error("minimum bet should be positive")
	}

	if comp.MaximumBet <= comp.MinimumBet {
		t.Error("maximum bet should be greater than minimum")
	}

	if comp.HouseEdge < 0.0 || comp.HouseEdge > 1.0 {
		t.Errorf("house edge %.2f should be in range [0.0, 1.0]", comp.HouseEdge)
	}

	if comp.WinMultiplier <= 0.0 {
		t.Error("win multiplier should be positive")
	}

	if comp.CurrentBet != 0 {
		t.Error("new merchant should have no active bet")
	}
}

func TestMerchantGamblingComponent_CanPlaceBet(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *MerchantGamblingComponent
		betAmount  int
		wantCanBet bool
	}{
		{
			name: "valid bet within range",
			setupFunc: func() *MerchantGamblingComponent {
				return NewMerchantGamblingComponent()
			},
			betAmount:  50,
			wantCanBet: true,
		},
		{
			name: "bet too low",
			setupFunc: func() *MerchantGamblingComponent {
				return NewMerchantGamblingComponent()
			},
			betAmount:  5,
			wantCanBet: false,
		},
		{
			name: "bet too high",
			setupFunc: func() *MerchantGamblingComponent {
				return NewMerchantGamblingComponent()
			},
			betAmount:  200,
			wantCanBet: false,
		},
		{
			name: "merchant doesn't offer gambling",
			setupFunc: func() *MerchantGamblingComponent {
				comp := NewMerchantGamblingComponent()
				comp.OffersGambling = false
				return comp
			},
			betAmount:  50,
			wantCanBet: false,
		},
		{
			name: "already has active bet",
			setupFunc: func() *MerchantGamblingComponent {
				comp := NewMerchantGamblingComponent()
				comp.CurrentBet = 30
				return comp
			},
			betAmount:  50,
			wantCanBet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.setupFunc()
			canBet := comp.CanPlaceBet(tt.betAmount)

			if canBet != tt.wantCanBet {
				t.Errorf("expected CanPlaceBet=%v, got %v", tt.wantCanBet, canBet)
			}
		})
	}
}

func TestMerchantGamblingComponent_PlaceBet(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *MerchantGamblingComponent
		playerID   uint64
		betAmount  int
		gameType   MiniGameType
		wantAccept bool
	}{
		{
			name: "valid bet accepted",
			setupFunc: func() *MerchantGamblingComponent {
				return NewMerchantGamblingComponent()
			},
			playerID:   123,
			betAmount:  50,
			gameType:   MiniGameCard,
			wantAccept: true,
		},
		{
			name: "bet too low rejected",
			setupFunc: func() *MerchantGamblingComponent {
				return NewMerchantGamblingComponent()
			},
			playerID:   123,
			betAmount:  5,
			gameType:   MiniGameCard,
			wantAccept: false,
		},
		{
			name: "game not offered rejected",
			setupFunc: func() *MerchantGamblingComponent {
				comp := NewMerchantGamblingComponent()
				comp.GamesOffered = []MiniGameType{MiniGameCard}
				return comp
			},
			playerID:   123,
			betAmount:  50,
			gameType:   MiniGamePuzzle,
			wantAccept: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.setupFunc()
			accepted := comp.PlaceBet(tt.playerID, tt.betAmount, tt.gameType)

			if accepted != tt.wantAccept {
				t.Errorf("expected PlaceBet=%v, got %v", tt.wantAccept, accepted)
			}

			if accepted {
				if comp.CurrentBet != tt.betAmount {
					t.Errorf("expected current bet %d, got %d", tt.betAmount, comp.CurrentBet)
				}
				if comp.BettingPlayerID != tt.playerID {
					t.Errorf("expected player ID %d, got %d", tt.playerID, comp.BettingPlayerID)
				}
				if comp.CurrentGameType != tt.gameType {
					t.Errorf("expected game type %v, got %v", tt.gameType, comp.CurrentGameType)
				}
			}
		})
	}
}

func TestMerchantGamblingComponent_CalculateWinnings(t *testing.T) {
	comp := NewMerchantGamblingComponent()
	comp.CurrentBet = 100

	// Player wins
	winnings := comp.CalculateWinnings(true)
	expectedWin := 100 + int(100*comp.WinMultiplier) // Original bet + winnings
	if winnings != expectedWin {
		t.Errorf("expected winnings %d, got %d", expectedWin, winnings)
	}

	// Player loses
	losses := comp.CalculateWinnings(false)
	if losses != 0 {
		t.Errorf("expected 0 for losing, got %d", losses)
	}
}

func TestMerchantGamblingComponent_ClearBet(t *testing.T) {
	comp := NewMerchantGamblingComponent()
	comp.PlaceBet(123, 50, MiniGameCard)

	// Verify bet is active
	if !comp.HasActiveBet() {
		t.Error("expected active bet before clear")
	}

	// Clear bet
	comp.ClearBet()

	// Verify bet is cleared
	if comp.HasActiveBet() {
		t.Error("expected no active bet after clear")
	}

	if comp.CurrentBet != 0 {
		t.Errorf("expected CurrentBet=0, got %d", comp.CurrentBet)
	}

	if comp.BettingPlayerID != 0 {
		t.Errorf("expected BettingPlayerID=0, got %d", comp.BettingPlayerID)
	}
}

func TestMerchantGamblingComponent_HasActiveBet(t *testing.T) {
	comp := NewMerchantGamblingComponent()

	// Initially no bet
	if comp.HasActiveBet() {
		t.Error("new merchant should have no active bet")
	}

	// Place bet
	comp.PlaceBet(123, 50, MiniGameCard)

	// Should have active bet
	if !comp.HasActiveBet() {
		t.Error("merchant should have active bet after PlaceBet")
	}

	// Clear bet
	comp.ClearBet()

	// Should have no bet again
	if comp.HasActiveBet() {
		t.Error("merchant should have no active bet after ClearBet")
	}
}
