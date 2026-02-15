package engine

import (
	"testing"
)

func TestReputationCompanionBonusSystem_ReputationTier(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)

	tests := []struct {
		name       string
		reputation int
		wantTier   int
	}{
		{"hostile", -50, 0},
		{"zero", 0, 0},
		{"low_neutral", 1, 1},
		{"mid_neutral", 50, 1},
		{"friendly", 51, 2},
		{"high_friendly", 75, 2},
		{"honored", 76, 3},
		{"max_honored", 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.reputationTier(tt.reputation)
			if got != tt.wantTier {
				t.Errorf("reputationTier(%d) = %d, want %d", tt.reputation, got, tt.wantTier)
			}
		})
	}
}

func TestReputationCompanionBonusSystem_CalculateBonus(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		tier      int
		wantAttGt float64
		wantDefGt float64
		wantSpdGt float64
	}{
		{"no_bonus", "fantasy", 0, 0.99, 0.99, 0.99},
		{"neutral_fantasy", "fantasy", 1, 1.04, 1.04, 1.02},
		{"friendly_fantasy", "fantasy", 2, 1.09, 1.09, 1.06},
		{"honored_fantasy", "fantasy", 3, 1.17, 1.14, 1.11},
		{"honored_scifi", "scifi", 3, 1.18, 1.15, 1.12},
		{"honored_horror", "horror", 3, 1.12, 1.10, 1.08},
		{"honored_cyberpunk", "cyberpunk", 3, 1.19, 1.16, 1.12},
		{"honored_postapoc", "postapoc", 3, 1.15, 1.12, 1.10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewReputationCompanionBonusSystem(nil, 42)
			sys.SetGenre(tt.genre)
			bonus := sys.calculateBonus(tt.tier)
			if bonus.attackMult < tt.wantAttGt {
				t.Errorf("attack mult = %f, want > %f", bonus.attackMult, tt.wantAttGt)
			}
			if bonus.defenseMult < tt.wantDefGt {
				t.Errorf("defense mult = %f, want > %f", bonus.defenseMult, tt.wantDefGt)
			}
			if bonus.speedMult < tt.wantSpdGt {
				t.Errorf("speed mult = %f, want > %f", bonus.speedMult, tt.wantSpdGt)
			}
		})
	}
}

func TestReputationCompanionBonusSystem_GetCompanionBonus_NoBonus(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)
	attack, defense, speed := sys.GetCompanionBonus(999)
	if attack != 1.0 || defense != 1.0 || speed != 1.0 {
		t.Errorf("expected 1.0/1.0/1.0 for unknown entity, got %f/%f/%f", attack, defense, speed)
	}
}

func TestReputationCompanionBonusSystem_HasActiveBonus(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)
	if sys.HasActiveBonus(999) {
		t.Error("expected no active bonus for unknown entity")
	}
}

func TestReputationCompanionBonusSystem_GenreMultiplier(t *testing.T) {
	tests := []struct {
		genre string
		want  float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.1},
		{"horror", 0.75},
		{"cyberpunk", 1.15},
		{"postapoc", 0.9},
		{"unknown", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewReputationCompanionBonusSystem(nil, 42)
			sys.SetGenre(tt.genre)
			got := sys.GetGenreMultiplier()
			if got != tt.want {
				t.Errorf("GetGenreMultiplier() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestReputationCompanionBonusSystem_UpdateSkipsWithoutFactionSystem(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)
	// Should not panic with nil faction system
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationCompanionBonusSystem_UpdateThrottles(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)
	sys.SetFactionSystem(&FactionSystem{Factions: make(map[string]*Faction)})
	// First call with small delta should be throttled
	sys.Update([]*Entity{}, 0.1)
	// timeSinceCheck should accumulate but not reset
	if sys.timeSinceCheck == 0 {
		t.Error("expected timeSinceCheck to accumulate")
	}
}

func TestReputationCompanionBonusSystem_ReverseBonus(t *testing.T) {
	sys := NewReputationCompanionBonusSystem(nil, 42)
	stats := &CompanionStatsComponent{Attack: 10.0, Defense: 8.0, Speed: 5.0}
	bonus := &reputationCompanionBonus{attackMult: 1.1, defenseMult: 1.05, speedMult: 1.03}
	sys.reverseBonus(stats, bonus)
	if stats.Attack < 9.0 || stats.Attack > 9.1 {
		t.Errorf("expected attack ~9.09 after reverse, got %f", stats.Attack)
	}
}

func TestReputationCompanionBonusParticleSystem_Creation(t *testing.T) {
	sys := NewReputationCompanionBonusParticleSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.spawnInterval != 1.5 {
		t.Errorf("expected spawnInterval 1.5, got %f", sys.spawnInterval)
	}
	if sys.particleCount != 3 {
		t.Errorf("expected particleCount 3, got %d", sys.particleCount)
	}
}

func TestReputationCompanionBonusParticleSystem_SetGenre(t *testing.T) {
	sys := NewReputationCompanionBonusParticleSystem(nil, 42)
	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("expected genreID scifi, got %s", sys.genreID)
	}
}

func TestReputationCompanionBonusParticleSystem_UpdateSkipsWithoutDeps(t *testing.T) {
	sys := NewReputationCompanionBonusParticleSystem(nil, 42)
	// Should not panic with nil dependencies
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationCompanionBonusParticleSystem_GenreParticleTypes(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewReputationCompanionBonusParticleSystem(nil, 42)
			sys.SetGenre(tt.genre)
			pt := sys.getParticleTypeForGenre()
			if pt < 0 {
				t.Errorf("unexpected negative particle type for genre %s", tt.genre)
			}
		})
	}
}
