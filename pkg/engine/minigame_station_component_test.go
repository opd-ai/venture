// Package engine provides tests for mini-game station components (Phase 27.3).
package engine

import (
	"testing"
)

func TestMiniGameStationComponent_Type(t *testing.T) {
	station := NewMiniGameStationComponent(MiniGameCard, 0.5)
	if station.Type() != "minigameStation" {
		t.Errorf("expected type 'minigameStation', got '%s'", station.Type())
	}
}

func TestNewMiniGameStationComponent(t *testing.T) {
	tests := []struct {
		name       string
		gameType   MiniGameType
		difficulty float64
		wantDiff   float64
	}{
		{
			name:       "card game medium difficulty",
			gameType:   MiniGameCard,
			difficulty: 0.5,
			wantDiff:   0.5,
		},
		{
			name:       "dice game easy",
			gameType:   MiniGameDice,
			difficulty: 0.2,
			wantDiff:   0.2,
		},
		{
			name:       "difficulty clamped low",
			gameType:   MiniGamePuzzle,
			difficulty: -0.5,
			wantDiff:   0.0,
		},
		{
			name:       "difficulty clamped high",
			gameType:   MiniGameHacking,
			difficulty: 1.5,
			wantDiff:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			station := NewMiniGameStationComponent(tt.gameType, tt.difficulty)

			if station.GameType != tt.gameType {
				t.Errorf("expected game type %v, got %v", tt.gameType, station.GameType)
			}

			if station.Difficulty != tt.wantDiff {
				t.Errorf("expected difficulty %v, got %v", tt.wantDiff, station.Difficulty)
			}

			if station.IsOccupied {
				t.Error("new station should not be occupied")
			}

			if station.OccupiedByPlayerID != 0 {
				t.Error("new station should have no player ID")
			}

			if station.StationName == "" {
				t.Error("station should have a name")
			}
		})
	}
}

func TestMiniGameStationComponent_CanPlay(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() *MiniGameStationComponent
		playerLevel int
		playerGold  int
		wantCanPlay bool
	}{
		{
			name: "player meets all requirements",
			setupFunc: func() *MiniGameStationComponent {
				return NewMiniGameStationComponent(MiniGameCard, 0.5)
			},
			playerLevel: 5,
			playerGold:  100,
			wantCanPlay: true,
		},
		{
			name: "station occupied",
			setupFunc: func() *MiniGameStationComponent {
				station := NewMiniGameStationComponent(MiniGameCard, 0.5)
				station.IsOccupied = true
				return station
			},
			playerLevel: 5,
			playerGold:  100,
			wantCanPlay: false,
		},
		{
			name: "player level too low",
			setupFunc: func() *MiniGameStationComponent {
				station := NewMiniGameStationComponent(MiniGameCard, 0.5)
				station.RequiresLevel = 10
				return station
			},
			playerLevel: 5,
			playerGold:  100,
			wantCanPlay: false,
		},
		{
			name: "not enough gold",
			setupFunc: func() *MiniGameStationComponent {
				station := NewMiniGameStationComponent(MiniGameCard, 0.5)
				station.EntryCost = 50
				return station
			},
			playerLevel: 5,
			playerGold:  30,
			wantCanPlay: false,
		},
		{
			name: "free game no restrictions",
			setupFunc: func() *MiniGameStationComponent {
				station := NewMiniGameStationComponent(MiniGamePuzzle, 0.5)
				station.EntryCost = -1 // Free
				station.RequiresLevel = 0
				return station
			},
			playerLevel: 1,
			playerGold:  0,
			wantCanPlay: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			station := tt.setupFunc()
			canPlay := station.CanPlay(tt.playerLevel, tt.playerGold)

			if canPlay != tt.wantCanPlay {
				t.Errorf("expected CanPlay=%v, got %v", tt.wantCanPlay, canPlay)
			}
		})
	}
}

func TestMiniGameStationComponent_StartAndEndGame(t *testing.T) {
	station := NewMiniGameStationComponent(MiniGameCard, 0.5)
	playerID := uint64(123)

	// Initially not occupied
	if station.IsOccupied {
		t.Error("station should not be occupied initially")
	}

	// Start game
	station.StartGame(playerID)

	if !station.IsOccupied {
		t.Error("station should be occupied after StartGame")
	}

	if station.OccupiedByPlayerID != playerID {
		t.Errorf("expected player ID %d, got %d", playerID, station.OccupiedByPlayerID)
	}

	// End game
	station.EndGame()

	if station.IsOccupied {
		t.Error("station should not be occupied after EndGame")
	}

	if station.OccupiedByPlayerID != 0 {
		t.Error("player ID should be reset after EndGame")
	}
}

func TestGetDefaultStationName(t *testing.T) {
	tests := []struct {
		gameType MiniGameType
		wantName string
	}{
		{MiniGameCard, "Card Table"},
		{MiniGameDice, "Dice Den"},
		{MiniGamePuzzle, "Puzzle Board"},
		{MiniGameMemory, "Memory Game"},
		{MiniGameLockPicking, "Practice Lock"},
		{MiniGameHacking, "Terminal"},
		{MiniGameRitual, "Ritual Circle"},
		{MiniGameType(999), "Game Station"}, // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			name := getDefaultStationName(tt.gameType)
			if name != tt.wantName {
				t.Errorf("expected name '%s', got '%s'", tt.wantName, name)
			}
		})
	}
}

func TestGetDefaultEntryCost(t *testing.T) {
	tests := []struct {
		gameType MiniGameType
		wantCost int
	}{
		{MiniGameCard, 20},
		{MiniGameDice, 15},
		{MiniGamePuzzle, -1},      // Free
		{MiniGameMemory, -1},      // Free
		{MiniGameLockPicking, -1}, // Free
		{MiniGameHacking, 10},
		{MiniGameRitual, 5},
		{MiniGameType(999), -1}, // Unknown type is free
	}

	for _, tt := range tests {
		t.Run(tt.gameType.String(), func(t *testing.T) {
			cost := getDefaultEntryCost(tt.gameType)
			if cost != tt.wantCost {
				t.Errorf("expected cost %d, got %d", tt.wantCost, cost)
			}
		})
	}
}

func TestClampDifficulty(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"normal value", 0.5, 0.5},
		{"low value", 0.1, 0.1},
		{"high value", 0.9, 0.9},
		{"below minimum", -0.5, 0.0},
		{"above maximum", 1.5, 1.0},
		{"exactly minimum", 0.0, 0.0},
		{"exactly maximum", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampDifficulty(tt.input)
			if got != tt.want {
				t.Errorf("clampDifficulty(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
