package engine

import (
	"testing"
)

func TestNewReputationHealingBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("expected updateInterval 0.5, got %f", sys.updateInterval)
	}
}

func TestReputationHealingBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre horror, got %s", sys.genreID)
	}
}

func TestReputationHealingBonusSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	if sys.factionSystem != fs {
		t.Error("expected faction system to be set")
	}
}

func TestReputationHealingBonusSystem_RegenForReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)

	tests := []struct {
		name       string
		reputation int
		expected   float64
	}{
		{"hostile", -100, 0.0},
		{"suspicious", -25, 0.0},
		{"zero", 0, 0.0},
		{"low_neutral", 10, 0.5},
		{"high_neutral", 50, 0.5},
		{"friendly", 60, 1.0},
		{"high_friendly", 75, 1.0},
		{"honored", 76, 2.0},
		{"max_honored", 100, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.regenForReputation(tt.reputation)
			if result != tt.expected {
				t.Errorf("regenForReputation(%d) = %f, want %f", tt.reputation, result, tt.expected)
			}
		})
	}
}

func TestReputationHealingBonusSystem_GetGenreMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)

	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.1},
		{"horror", 0.7},
		{"cyberpunk", 1.15},
		{"postapoc", 0.85},
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

func TestReputationHealingBonusSystem_UpdateNoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	// Should not panic with nil faction system
	sys.Update([]*Entity{}, 1.0)
}

func TestReputationHealingBonusSystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	// Call with deltaTime less than updateInterval - should accumulate but not process
	sys.Update([]*Entity{}, 0.1)
	if sys.timeSinceCheck != 0.1 {
		t.Errorf("expected timeSinceCheck 0.1, got %f", sys.timeSinceCheck)
	}
}

func TestReputationHealingBonusSystem_GetRegenRate(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)

	// No rate set yet
	if rate := sys.GetRegenRate(1); rate != 0.0 {
		t.Errorf("expected 0.0 for unknown entity, got %f", rate)
	}

	// Manually set for testing
	sys.regenRates[1] = 1.5
	if rate := sys.GetRegenRate(1); rate != 1.5 {
		t.Errorf("expected 1.5, got %f", rate)
	}
}

func TestReputationHealingBonusSystem_SkipsNonPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	// Entity without input component should be skipped
	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})

	sys.Update([]*Entity{entity}, 1.0)

	if rate := sys.GetRegenRate(entity.ID); rate != 0.0 {
		t.Errorf("expected 0 regen for non-player entity, got %f", rate)
	}
}

func TestReputationHealingBonusSystem_SkipsFullHealth(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&StubInput{})

	sys.Update([]*Entity{entity}, 1.0)

	if rate := sys.GetRegenRate(entity.ID); rate != 0.0 {
		t.Errorf("expected 0 regen for full health, got %f", rate)
	}
}

func TestReputationHealingBonusSystem_SkipsDeadEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusSystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 0, Max: 100})
	entity.AddComponent(&StubInput{})

	sys.Update([]*Entity{entity}, 1.0)

	if rate := sys.GetRegenRate(entity.ID); rate != 0.0 {
		t.Errorf("expected 0 regen for dead entity, got %f", rate)
	}
}
