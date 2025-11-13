package engine

import (
	"testing"
)

func TestFactionReactionSystem_GetReactionLevel(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	entity := world.CreateEntity()
	repComp := &ReputationComponent{
		Factions: map[string]float64{
			"Merchants": 0,
			"Guards":    50,
			"Thieves":   -80,
		},
		Alignment: Alignment{},
		KarmaDeed: []Deed{},
	}
	entity.AddComponent(repComp)
	world.Update(0) // Process entity addition to world

	tests := []struct {
		name     string
		faction  string
		expected string
	}{
		{"neutral faction", "Merchants", "neutral"},
		{"friendly faction", "Guards", "friendly"},
		{"hostile faction", "Thieves", "hostile"},
		{"unknown faction", "Unknown", "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.GetReactionLevel(entity.ID, tt.faction)
			if result != tt.expected {
				t.Errorf("GetReactionLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFactionReactionSystem_GetPriceModifier(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	entity := world.CreateEntity()
	repComp := &ReputationComponent{
		Factions:  make(map[string]float64),
		Alignment: Alignment{},
		KarmaDeed: []Deed{},
	}
	entity.AddComponent(repComp)
	world.Update(0) // Process entity addition to world

	tests := []struct {
		name       string
		reputation float64
		expected   float64
	}{
		{"hostile", -80, 2.0},
		{"unfriendly", -30, 1.5},
		{"neutral", 0, 1.0},
		{"friendly", 50, 0.85},
		{"honored", 80, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repComp.Factions["Test"] = tt.reputation
			result := system.GetPriceModifier(entity.ID, "Test")
			if result != tt.expected {
				t.Errorf("GetPriceModifier() = %.2f, want %.2f", result, tt.expected)
			}
		})
	}
}

func TestFactionReactionSystem_ShouldAttackOnSight(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	entity := world.CreateEntity()
	repComp := &ReputationComponent{
		Factions:  make(map[string]float64),
		Alignment: Alignment{},
		KarmaDeed: []Deed{},
	}
	entity.AddComponent(repComp)
	world.Update(0) // Process entity addition to world

	tests := []struct {
		name       string
		reputation float64
		expected   bool
	}{
		{"hostile", -80, true},
		{"unfriendly", -30, false},
		{"neutral", 0, false},
		{"friendly", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repComp.Factions["Test"] = tt.reputation
			result := system.ShouldAttackOnSight(entity.ID, "Test")
			if result != tt.expected {
				t.Errorf("ShouldAttackOnSight() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFactionReactionSystem_CanAcceptQuest(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	entity := world.CreateEntity()
	repComp := &ReputationComponent{
		Factions:  map[string]float64{"Test": 30},
		Alignment: Alignment{},
		KarmaDeed: []Deed{},
	}
	entity.AddComponent(repComp)
	world.Update(0) // Process entity addition to world

	tests := []struct {
		name          string
		minReputation float64
		expected      bool
	}{
		{"can accept low requirement", 10, true},
		{"can accept exact requirement", 30, true},
		{"cannot accept high requirement", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.CanAcceptQuest(entity.ID, "Test", tt.minReputation)
			if result != tt.expected {
				t.Errorf("CanAcceptQuest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFactionReactionSystem_GetDialogOptions(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	entity := world.CreateEntity()
	repComp := &ReputationComponent{
		Factions:  make(map[string]float64),
		Alignment: Alignment{},
		KarmaDeed: []Deed{},
	}
	entity.AddComponent(repComp)
	world.Update(0) // Process entity addition to world

	tests := []struct {
		name       string
		reputation float64
		minOptions int
		hasQuests  bool
	}{
		{"hostile", -80, 2, false},
		{"neutral", 0, 3, true},
		{"friendly", 50, 4, true},
		{"honored", 80, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repComp.Factions["Test"] = tt.reputation
			options := system.GetDialogOptions(entity.ID, "Test")

			if len(options) < tt.minOptions {
				t.Errorf("GetDialogOptions() len = %d, want >= %d", len(options), tt.minOptions)
			}

			hasQuests := false
			for _, opt := range options {
				if opt == "Quests" {
					hasQuests = true
					break
				}
			}

			if hasQuests != tt.hasQuests {
				t.Errorf("GetDialogOptions() has quests = %v, want %v", hasQuests, tt.hasQuests)
			}
		})
	}
}

func TestFactionReactionSystem_GetAlignmentDescription(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	tests := []struct {
		name     string
		lawAxis  float64
		goodAxis float64
		expected string
	}{
		{"true neutral", 0, 0, "True Neutral"},
		{"lawful good", 0.8, 0.8, "Lawful Good"},
		{"chaotic evil", -0.8, -0.8, "Chaotic Evil"},
		{"neutral good", 0, 0.8, "Good"},
		{"lawful neutral", 0.8, 0, "Lawful"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			repComp := &ReputationComponent{
				Factions: make(map[string]float64),
				Alignment: Alignment{
					LawAxis:  tt.lawAxis,
					GoodAxis: tt.goodAxis,
				},
				KarmaDeed: []Deed{},
			}
			entity.AddComponent(repComp)
			world.Update(0) // Process entity addition to world

			result := system.GetAlignmentDescription(entity.ID)
			if result != tt.expected {
				t.Errorf("GetAlignmentDescription() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFactionReactionSystem_GetReputationThreshold(t *testing.T) {
	world := NewWorld()
	system := NewFactionReactionSystem(world)

	tests := []struct {
		level    string
		expected float64
	}{
		{"hostile", -75.0},
		{"unfriendly", -25.0},
		{"neutral", 0.0},
		{"friendly", 25.0},
		{"honored", 75.0},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := system.GetReputationThreshold(tt.level)
			if result != tt.expected {
				t.Errorf("GetReputationThreshold(%s) = %.2f, want %.2f", tt.level, result, tt.expected)
			}
		})
	}
}
