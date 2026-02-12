package item

import "testing"

func TestRarity_Value(t *testing.T) {
	tests := []struct {
		name     string
		rarity   Rarity
		expected float64
	}{
		{"common", RarityCommon, 1.0},
		{"uncommon", RarityUncommon, 1.2},
		{"rare", RarityRare, 1.5},
		{"epic", RarityEpic, 2.0},
		{"legendary", RarityLegendary, 3.0},
		{"unknown", Rarity(99), 1.0}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rarity.Value()
			if result != tt.expected {
				t.Errorf("Rarity.Value() for %s = %.1f, want %.1f", tt.name, result, tt.expected)
			}
		})
	}
}

func TestRarity_ValueMatchesMultipliers(t *testing.T) {
	// Verify that Value() matches the documented stat multipliers
	// from the generator (see generator.go:373-378)
	rarityMultipliers := map[Rarity]float64{
		RarityCommon:    1.0,
		RarityUncommon:  1.2,
		RarityRare:      1.5,
		RarityEpic:      2.0,
		RarityLegendary: 3.0,
	}

	for rarity, expectedMultiplier := range rarityMultipliers {
		actualValue := rarity.Value()
		if actualValue != expectedMultiplier {
			t.Errorf("Rarity %s: Value()=%.1f, expected multiplier=%.1f",
				rarity.String(), actualValue, expectedMultiplier)
		}
	}
}
