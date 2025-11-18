package persistence

import (
	"testing"
)

func TestTrustLevelString(t *testing.T) {
	tests := []struct {
		level    TrustLevel
		expected string
	}{
		{TrustLevelStranger, "Stranger"},
		{TrustLevelAcquaintance, "Acquaintance"},
		{TrustLevelFriend, "Friend"},
		{TrustLevelTrusted, "Trusted"},
		{TrustLevel(99), "Unknown"}, // Invalid value
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("TrustLevel.String() = %s, want %s", result, tt.expected)
		}
	}
}

func TestTypesGetTrustLevel(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected TrustLevel
	}{
		{"stranger low", 0.0, TrustLevelStranger},
		{"stranger high", 0.29, TrustLevelStranger},
		{"acquaintance low", 0.3, TrustLevelAcquaintance},
		{"acquaintance mid", 0.45, TrustLevelAcquaintance},
		{"acquaintance high", 0.59, TrustLevelAcquaintance},
		{"friend low", 0.6, TrustLevelFriend},
		{"friend mid", 0.7, TrustLevelFriend},
		{"friend high", 0.79, TrustLevelFriend},
		{"trusted low", 0.8, TrustLevelTrusted},
		{"trusted mid", 0.9, TrustLevelTrusted},
		{"trusted max", 1.0, TrustLevelTrusted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTrustLevel(tt.score)
			if result != tt.expected {
				t.Errorf("GetTrustLevel(%f) = %s, want %s", tt.score, result, tt.expected)
			}
		})
	}
}

func TestCanTradeRarity(t *testing.T) {
	tests := []struct {
		name     string
		level    TrustLevel
		rarity   string
		expected bool
	}{
		// Stranger (common only)
		{"stranger common", TrustLevelStranger, "common", true},
		{"stranger uncommon", TrustLevelStranger, "uncommon", false},
		{"stranger rare", TrustLevelStranger, "rare", false},
		{"stranger epic", TrustLevelStranger, "epic", false},
		{"stranger legendary", TrustLevelStranger, "legendary", false},
		
		// Acquaintance (common + uncommon)
		{"acquaintance common", TrustLevelAcquaintance, "common", true},
		{"acquaintance uncommon", TrustLevelAcquaintance, "uncommon", true},
		{"acquaintance rare", TrustLevelAcquaintance, "rare", false},
		{"acquaintance epic", TrustLevelAcquaintance, "epic", false},
		{"acquaintance legendary", TrustLevelAcquaintance, "legendary", false},
		
		// Friend (up to rare)
		{"friend common", TrustLevelFriend, "common", true},
		{"friend uncommon", TrustLevelFriend, "uncommon", true},
		{"friend rare", TrustLevelFriend, "rare", true},
		{"friend epic", TrustLevelFriend, "epic", false},
		{"friend legendary", TrustLevelFriend, "legendary", false},
		
		// Trusted (all items)
		{"trusted common", TrustLevelTrusted, "common", true},
		{"trusted uncommon", TrustLevelTrusted, "uncommon", true},
		{"trusted rare", TrustLevelTrusted, "rare", true},
		{"trusted epic", TrustLevelTrusted, "epic", true},
		{"trusted legendary", TrustLevelTrusted, "legendary", true},
		
		// Invalid rarity
		{"invalid rarity", TrustLevelTrusted, "invalid", false},
		{"empty rarity", TrustLevelStranger, "", false},
		
		// Invalid trust level
		{"invalid level", TrustLevel(99), "common", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTradeRarity(tt.level, tt.rarity)
			if result != tt.expected {
				t.Errorf("CanTradeRarity(%s, %s) = %v, want %v", tt.level, tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestTrustRecordFields(t *testing.T) {
	record := TrustRecord{
		PlayerA:      "alice",
		PlayerB:      "bob",
		Score:        0.75,
		Interactions: 5,
	}

	if record.PlayerA != "alice" {
		t.Errorf("PlayerA = %s, want alice", record.PlayerA)
	}
	if record.PlayerB != "bob" {
		t.Errorf("PlayerB = %s, want bob", record.PlayerB)
	}
	if record.Score != 0.75 {
		t.Errorf("Score = %f, want 0.75", record.Score)
	}
	if record.Interactions != 5 {
		t.Errorf("Interactions = %d, want 5", record.Interactions)
	}
}

func TestDecayConstants(t *testing.T) {
	if DecayRatePerDay != 0.01 {
		t.Errorf("DecayRatePerDay = %f, want 0.01", DecayRatePerDay)
	}
	if MinTrustScore != 0.0 {
		t.Errorf("MinTrustScore = %f, want 0.0", MinTrustScore)
	}
	if MaxTrustScore != 1.0 {
		t.Errorf("MaxTrustScore = %f, want 1.0", MaxTrustScore)
	}
}

func TestTrustLevelProgression(t *testing.T) {
	// Test progression through trust levels with incremental scores
	scores := []float64{0.0, 0.3, 0.6, 0.8, 1.0}
	expectedLevels := []TrustLevel{
		TrustLevelStranger,
		TrustLevelAcquaintance,
		TrustLevelFriend,
		TrustLevelTrusted,
		TrustLevelTrusted,
	}

	for i, score := range scores {
		level := GetTrustLevel(score)
		if level != expectedLevels[i] {
			t.Errorf("Score %f should be level %s, got %s", score, expectedLevels[i], level)
		}
	}
}

func TestTradeLimitsByTrustLevel(t *testing.T) {
	// Test that trade limits increase with trust level
	rarities := []string{"common", "uncommon", "rare", "epic", "legendary"}
	
	for i, rarity := range rarities {
		// Each trust level should allow i or fewer rarity levels
		for level := TrustLevelStranger; level <= TrustLevelTrusted; level++ {
			canTrade := CanTradeRarity(level, rarity)
			
			// Calculate expected result based on trust level and rarity index
			maxRarityForLevel := map[TrustLevel]int{
				TrustLevelStranger:     0,
				TrustLevelAcquaintance: 1,
				TrustLevelFriend:       2,
				TrustLevelTrusted:      4,
			}
			
			expected := i <= maxRarityForLevel[level]
			if canTrade != expected {
				t.Errorf("Level %s, rarity %s: got %v, want %v", level, rarity, canTrade, expected)
			}
		}
	}
}

// Benchmark trust level computation
func BenchmarkGetTrustLevel(b *testing.B) {
	scores := []float64{0.1, 0.4, 0.7, 0.9}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetTrustLevel(scores[i%len(scores)])
	}
}

// Benchmark trade rarity check
func BenchmarkCanTradeRarity(b *testing.B) {
	levels := []TrustLevel{TrustLevelStranger, TrustLevelAcquaintance, TrustLevelFriend, TrustLevelTrusted}
	rarities := []string{"common", "uncommon", "rare", "epic", "legendary"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		rarity := rarities[i%len(rarities)]
		_ = CanTradeRarity(level, rarity)
	}
}
