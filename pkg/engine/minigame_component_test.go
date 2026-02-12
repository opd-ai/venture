package engine

import (
	"testing"
)

func TestMiniGameType_String(t *testing.T) {
	tests := []struct {
		name     string
		gameType MiniGameType
		want     string
	}{
		{"Card Game", MiniGameCard, "Card Game"},
		{"Dice Game", MiniGameDice, "Dice Game"},
		{"Puzzle", MiniGamePuzzle, "Puzzle"},
		{"Memory", MiniGameMemory, "Memory"},
		{"Lock-Picking", MiniGameLockPicking, "Lock-Picking"},
		{"Hacking", MiniGameHacking, "Hacking"},
		{"Ritual", MiniGameRitual, "Ritual"},
		{"Unknown", MiniGameType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gameType.String()
			if got != tt.want {
				t.Errorf("MiniGameType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMiniGameComponent_Type(t *testing.T) {
	comp := MiniGameComponent{}
	if comp.Type() != "minigame" {
		t.Errorf("MiniGameComponent.Type() = %v, want minigame", comp.Type())
	}
}

func TestMiniGameComponent_Creation(t *testing.T) {
	reward := &Reward{
		Gold:  100,
		XP:    50.0,
		Items: []uint64{1, 2, 3},
	}

	comp := MiniGameComponent{
		GameType:    MiniGameCard,
		Active:      true,
		State:       map[string]interface{}{"test": "value"},
		Difficulty:  0.5,
		TimeLimit:   600.0,
		TimeElapsed: 10.0,
		Reward:      reward,
	}

	if comp.GameType != MiniGameCard {
		t.Errorf("GameType = %v, want %v", comp.GameType, MiniGameCard)
	}
	if !comp.Active {
		t.Error("Active should be true")
	}
	if comp.Difficulty != 0.5 {
		t.Errorf("Difficulty = %v, want 0.5", comp.Difficulty)
	}
	if comp.TimeLimit != 600.0 {
		t.Errorf("TimeLimit = %v, want 600.0", comp.TimeLimit)
	}
	if comp.TimeElapsed != 10.0 {
		t.Errorf("TimeElapsed = %v, want 10.0", comp.TimeElapsed)
	}
	if comp.Reward == nil {
		t.Fatal("Reward should not be nil")
	}
	if comp.Reward.Gold != 100 {
		t.Errorf("Reward.Gold = %v, want 100", comp.Reward.Gold)
	}
	if comp.Reward.XP != 50.0 {
		t.Errorf("Reward.XP = %v, want 50.0", comp.Reward.XP)
	}
	if len(comp.Reward.Items) != 3 {
		t.Errorf("len(Reward.Items) = %v, want 3", len(comp.Reward.Items))
	}
}

func TestReward_EmptyItems(t *testing.T) {
	reward := &Reward{
		Gold:  0,
		XP:    0.0,
		Items: []uint64{},
	}

	if reward.Gold != 0 {
		t.Errorf("Gold = %v, want 0", reward.Gold)
	}
	if reward.XP != 0.0 {
		t.Errorf("XP = %v, want 0.0", reward.XP)
	}
	if len(reward.Items) != 0 {
		t.Errorf("len(Items) = %v, want 0", len(reward.Items))
	}
}

func TestReward_NilItems(t *testing.T) {
	reward := &Reward{
		Gold:  50,
		XP:    25.0,
		Items: nil,
	}

	if reward.Items != nil {
		t.Error("Items should be nil")
	}
	// Should not panic when accessing nil slice
	if len(reward.Items) != 0 {
		t.Errorf("len(nil Items) = %v, want 0", len(reward.Items))
	}
}

func TestMiniGameComponent_GameInstance(t *testing.T) {
	// Create a stub mini-game instance
	stub := &StubMiniGame{
		isComplete: false,
		reward:     &Reward{Gold: 50, XP: 25.0},
	}

	comp := MiniGameComponent{
		GameType:     MiniGamePuzzle,
		Active:       true,
		GameInstance: stub,
	}

	if comp.GameInstance == nil {
		t.Fatal("GameInstance should not be nil")
	}

	// Test that we can call interface methods
	if comp.GameInstance.IsComplete() {
		t.Error("GameInstance.IsComplete() should be false initially")
	}

	reward := comp.GameInstance.GetReward()
	if reward == nil {
		t.Fatal("GameInstance.GetReward() should not be nil")
	}
	if reward.Gold != 50 {
		t.Errorf("reward.Gold = %v, want 50", reward.Gold)
	}
}

// StubMiniGame is a test implementation of MiniGame interface
type StubMiniGame struct {
	initError    error
	updateErr    error
	prepareErr   error
	isComplete   bool
	reward       *Reward
	renderOutput MiniGameRenderOutput
}

func (s *StubMiniGame) Initialize(seed int64, difficulty float64) error {
	return s.initError
}

func (s *StubMiniGame) Update(deltaTime float64) error {
	return s.updateErr
}

func (s *StubMiniGame) PrepareRender(screenWidth, screenHeight int) error {
	return s.prepareErr
}

func (s *StubMiniGame) GetRenderOutput() MiniGameRenderOutput {
	return s.renderOutput
}

func (s *StubMiniGame) IsComplete() bool {
	return s.isComplete
}

func (s *StubMiniGame) GetReward() *Reward {
	return s.reward
}
