package engine

import (
	"testing"
)

func TestNewReputationSpellDamageBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("expected updateInterval 1.0, got %f", sys.updateInterval)
	}
}

func TestReputationSpellDamageBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("expected genre scifi, got %s", sys.genreID)
	}
}

func TestReputationSpellDamageBonusSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	if sys.factionSystem != fs {
		t.Error("expected faction system to be set")
	}
}

func TestReputationSpellDamageBonusSystem_BonusForReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)

	tests := []struct {
		name       string
		reputation int
		expected   float64
	}{
		{"hostile", -100, 0.0},
		{"suspicious", -25, 0.0},
		{"zero", 0, 0.0},
		{"low_neutral", 10, 2.0},
		{"high_neutral", 50, 2.0},
		{"friendly", 60, 5.0},
		{"high_friendly", 75, 5.0},
		{"honored", 76, 10.0},
		{"max_honored", 100, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.bonusForReputation(tt.reputation)
			if result != tt.expected {
				t.Errorf("bonusForReputation(%d) = %f, want %f", tt.reputation, result, tt.expected)
			}
		})
	}
}

func TestReputationSpellDamageBonusSystem_GetGenreMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)

	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.15},
		{"horror", 0.75},
		{"cyberpunk", 1.10},
		{"postapoc", 0.90},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			result := sys.GetGenreMultiplier()
			if result != tt.expected {
				t.Errorf("GetGenreMultiplier() for %s = %f, want %f", tt.genre, result, tt.expected)
			}
		})
	}
}

func TestReputationSpellDamageBonusSystem_UpdateNoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	// Should not panic with nil faction system
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationSpellDamageBonusSystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	sys.Update([]*Entity{}, 0.3)
	if sys.timeSinceCheck != 0.3 {
		t.Errorf("expected timeSinceCheck 0.3, got %f", sys.timeSinceCheck)
	}
}

func TestReputationSpellDamageBonusSystem_GetBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)

	if bonus := sys.GetBonus(1); bonus != 0.0 {
		t.Errorf("expected 0.0 for unknown entity, got %f", bonus)
	}

	sys.appliedBonuses[1] = 5.0
	if bonus := sys.GetBonus(1); bonus != 5.0 {
		t.Errorf("expected 5.0, got %f", bonus)
	}
}

func TestReputationSpellDamageBonusSystem_SkipsNonPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	entity := NewEntity(1)
	entity.AddComponent(NewStatsComponent())

	sys.Update([]*Entity{entity}, 2.0)

	if bonus := sys.GetBonus(entity.ID); bonus != 0.0 {
		t.Errorf("expected 0 bonus for non-player entity, got %f", bonus)
	}
}

func TestReputationSpellDamageBonusSystem_SkipsEntitiesWithoutStats(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	entity := NewEntity(1)
	entity.AddComponent(&StubInput{})

	sys.Update([]*Entity{entity}, 2.0)

	if bonus := sys.GetBonus(entity.ID); bonus != 0.0 {
		t.Errorf("expected 0 bonus for entity without stats, got %f", bonus)
	}
}

func TestReputationSpellDamageBonusSystem_RemovesBonusWhenLost(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusSystem(world, 42)

	entity := NewEntity(1)
	stats := NewStatsComponent()
	stats.MagicPower = 20.0
	entity.AddComponent(stats)

	// Simulate a previously applied bonus
	sys.appliedBonuses[entity.ID] = 5.0
	stats.MagicPower += 5.0 // was 20, now 25

	// Process without input component triggers removeBonus
	sys.removeBonus(entity)

	if stats.MagicPower != 20.0 {
		t.Errorf("expected MagicPower 20.0 after removal, got %f", stats.MagicPower)
	}
	if _, exists := sys.appliedBonuses[entity.ID]; exists {
		t.Error("expected bonus to be removed from cache")
	}
}

func TestReputationSpellDamageBonusSystem_NilWorld(t *testing.T) {
	sys := NewReputationSpellDamageBonusSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	sys.Update([]*Entity{}, 2.0)
}
