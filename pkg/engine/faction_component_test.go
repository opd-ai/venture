package engine

import (
	"testing"
)

func TestFactionComponent_Type(t *testing.T) {
	fc := FactionComponent{}
	if fc.Type() != "faction" {
		t.Errorf("Expected type 'faction', got '%s'", fc.Type())
	}
}

func TestFactionComponent_GetReputationLevel(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   string
	}{
		{"Hostile lower bound", -100, "Hostile"},
		{"Hostile upper bound", -50, "Hostile"},
		{"Suspicious lower bound", -49, "Suspicious"},
		{"Suspicious zero", 0, "Suspicious"},
		{"Neutral lower bound", 1, "Neutral"},
		{"Neutral mid", 25, "Neutral"},
		{"Neutral upper bound", 50, "Neutral"},
		{"Friendly lower bound", 51, "Friendly"},
		{"Friendly max", 100, "Friendly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			if got := fc.GetReputationLevel(); got != tt.expected {
				t.Errorf("GetReputationLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFactionComponent_IsHostile(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   bool
	}{
		{"Hostile at -100", -100, true},
		{"Hostile at -50", -50, true},
		{"Not hostile at -49", -49, false},
		{"Not hostile at 0", 0, false},
		{"Not hostile at 50", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			if got := fc.IsHostile(); got != tt.expected {
				t.Errorf("IsHostile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFactionComponent_IsSuspicious(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   bool
	}{
		{"Not suspicious at -50", -50, false},
		{"Suspicious at -49", -49, true},
		{"Suspicious at -25", -25, true},
		{"Suspicious at 0", 0, true},
		{"Not suspicious at 1", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			if got := fc.IsSuspicious(); got != tt.expected {
				t.Errorf("IsSuspicious() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFactionComponent_IsNeutral(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   bool
	}{
		{"Not neutral at 0", 0, false},
		{"Neutral at 1", 1, true},
		{"Neutral at 25", 25, true},
		{"Neutral at 50", 50, true},
		{"Not neutral at 51", 51, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			if got := fc.IsNeutral(); got != tt.expected {
				t.Errorf("IsNeutral() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFactionComponent_IsFriendly(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   bool
	}{
		{"Not friendly at 50", 50, false},
		{"Friendly at 51", 51, true},
		{"Friendly at 75", 75, true},
		{"Friendly at 100", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			if got := fc.IsFriendly(); got != tt.expected {
				t.Errorf("IsFriendly() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFactionComponent_GetPriceMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		reputation int
		expected   float64
		tolerance  float64
	}{
		{"Hostile no trade", -100, 0.0, 0.001},
		{"Hostile boundary", -50, 0.0, 0.001},
		{"Suspicious markup", -25, 1.5, 0.001},
		{"Suspicious zero", 0, 1.5, 0.001},
		{"Neutral normal", 1, 1.0, 0.001},
		{"Neutral mid", 25, 1.0, 0.001},
		{"Neutral high", 50, 1.0, 0.001},
		{"Friendly slight discount", 51, 0.995, 0.001}, // ~0.5% discount
		{"Friendly mid discount", 75, 0.875, 0.001},    // ~12.5% discount
		{"Friendly max discount", 100, 0.75, 0.001},    // 25% discount
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FactionComponent{Reputation: tt.reputation}
			got := fc.GetPriceMultiplier()
			diff := got - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("GetPriceMultiplier() = %v, want %v (±%v)", got, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestFactionType_String(t *testing.T) {
	tests := []struct {
		name        string
		factionType FactionType
		expectedStr string
	}{
		{"Kingdom", FactionTypeKingdom, "kingdom"},
		{"Guild", FactionTypeGuild, "guild"},
		{"Cult", FactionTypeCult, "cult"},
		{"Corporation", FactionTypeCorporation, "corporation"},
		{"Gang", FactionTypeGang, "gang"},
		{"Rebels", FactionTypeRebels, "rebels"},
		{"Merchants", FactionTypeMerchants, "merchants"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.factionType.String(); got != tt.expectedStr {
				t.Errorf("String() = %v, want %v", got, tt.expectedStr)
			}
		})
	}
}

func TestFaction_GetRelationship(t *testing.T) {
	faction := &Faction{
		ID: "faction1",
		Relationships: map[string]int{
			"faction2": 75,
			"faction3": -60,
			"faction4": 0,
		},
	}

	tests := []struct {
		name     string
		otherID  string
		expected int
	}{
		{"Allied faction", "faction2", 75},
		{"Enemy faction", "faction3", -60},
		{"Neutral faction", "faction4", 0},
		{"Unknown faction defaults neutral", "faction5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := faction.GetRelationship(tt.otherID); got != tt.expected {
				t.Errorf("GetRelationship() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFaction_IsEnemy(t *testing.T) {
	faction := &Faction{
		ID: "faction1",
		Relationships: map[string]int{
			"faction2": 75,
			"faction3": -60,
			"faction4": -50,
			"faction5": -49,
		},
	}

	tests := []struct {
		name     string
		otherID  string
		expected bool
	}{
		{"Allied faction is not enemy", "faction2", false},
		{"Enemy faction", "faction3", true},
		{"Enemy at boundary", "faction4", true},
		{"Not enemy at -49", "faction5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := faction.IsEnemy(tt.otherID); got != tt.expected {
				t.Errorf("IsEnemy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFaction_IsAlly(t *testing.T) {
	faction := &Faction{
		ID: "faction1",
		Relationships: map[string]int{
			"faction2": 75,
			"faction3": 51,
			"faction4": 50,
			"faction5": -60,
		},
	}

	tests := []struct {
		name     string
		otherID  string
		expected bool
	}{
		{"Allied faction", "faction2", true},
		{"Allied at boundary", "faction3", true},
		{"Not allied at 50", "faction4", false},
		{"Enemy is not ally", "faction5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := faction.IsAlly(tt.otherID); got != tt.expected {
				t.Errorf("IsAlly() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReputationChange_Constants(t *testing.T) {
	// Verify reputation change constants are reasonable
	tests := []struct {
		name     string
		constant int
		minVal   int
		maxVal   int
	}{
		{"Kill member penalty", ReputationKillMember, -20, -5},
		{"Complete quest reward", ReputationCompleteQuest, 10, 25},
		{"Betray penalty", ReputationBetray, -100, -30},
		{"Rescue reward", ReputationRescue, 10, 30},
		{"Steal penalty", ReputationSteal, -10, -1},
		{"Donate reward", ReputationDonate, 1, 10},
		{"Kill enemy reward", ReputationKillEnemy, 5, 20},
		{"Kill ally penalty", ReputationKillAlly, -30, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant < tt.minVal || tt.constant > tt.maxVal {
				t.Errorf("%s = %d, should be between %d and %d",
					tt.name, tt.constant, tt.minVal, tt.maxVal)
			}
		})
	}
}
