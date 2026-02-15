package engine

import (
	"testing"
)

func TestNewReputationEquipmentDurabilitySystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.updateInterval != 2.0 {
		t.Errorf("expected updateInterval=2.0, got %f", sys.updateInterval)
	}
	if sys.rng == nil {
		t.Error("expected rng to be initialized")
	}
	if len(sys.appliedModifiers) != 0 {
		t.Error("expected empty applied modifiers")
	}
}

func TestReputationEquipmentDurabilitySystem_ModifierForReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	tests := []struct {
		name     string
		bestRep  int
		worstRep int
		want     float64
	}{
		{"honored_rep", 80, 0, 0.20},
		{"friendly_rep", 60, 0, 0.12},
		{"neutral_positive", 30, 0, 0.05},
		{"zero_rep", 0, 0, 0.0},
		{"hostile_rep", 0, -60, -0.15},
		{"mixed_honored", 90, -80, 0.20},
		{"slightly_negative", 0, -30, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.modifierForReputation(tt.bestRep, tt.worstRep)
			if got != tt.want {
				t.Errorf("modifierForReputation(%d, %d) = %f, want %f", tt.bestRep, tt.worstRep, got, tt.want)
			}
		})
	}
}

func TestReputationEquipmentDurabilitySystem_GenreMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	tests := []struct {
		genre string
		want  float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.2},
		{"horror", 0.6},
		{"cyberpunk", 1.1},
		{"postapoc", 0.8},
		{"unknown", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			got := sys.GetGenreMultiplier()
			if got != tt.want {
				t.Errorf("genre %s: got %f, want %f", tt.genre, got, tt.want)
			}
		})
	}
}

func TestReputationEquipmentDurabilitySystem_UpdateNoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(NewStubInput())
	// Should not panic with nil faction system
	sys.Update([]*Entity{entity}, 3.0)
}

func TestReputationEquipmentDurabilitySystem_UpdateSkipsNonPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	equipComp := NewEquipmentComponent()
	entity.AddComponent(equipComp)
	// No input component - should be skipped
	sys.Update([]*Entity{entity}, 3.0)
}

func TestReputationEquipmentDurabilitySystem_DamageStateChanged(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	tests := []struct {
		name    string
		oldDur  int
		newDur  int
		maxDur  int
		changed bool
	}{
		{"no_change", 80, 78, 100, false},
		{"cross_75_down", 76, 74, 100, true},
		{"cross_50_down", 51, 49, 100, true},
		{"cross_25_down", 26, 24, 100, true},
		{"cross_75_up", 74, 76, 100, true},
		{"zero_max", 50, 40, 0, false},
		{"same_value", 80, 80, 100, false},
		{"cross_50_up", 49, 51, 100, true},
		{"cross_25_up", 24, 26, 100, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.damageStateChanged(tt.oldDur, tt.newDur, tt.maxDur)
			if got != tt.changed {
				t.Errorf("damageStateChanged(%d,%d,%d) = %v, want %v", tt.oldDur, tt.newDur, tt.maxDur, got, tt.changed)
			}
		})
	}
}

func TestReputationEquipmentDurabilitySystem_GetModifier(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	if got := sys.GetDurabilityModifier(999); got != 0.0 {
		t.Errorf("expected 0.0 for unknown entity, got %f", got)
	}
	if got := sys.GetDurabilityModifierPercent(999); got != 0.0 {
		t.Errorf("expected 0.0 for unknown entity percent, got %f", got)
	}
}

func TestReputationEquipmentDurabilitySystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	if sys.factionSystem != nil {
		t.Error("expected nil faction system initially")
	}

	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	if sys.factionSystem != fs {
		t.Error("faction system not set correctly")
	}
}

func TestReputationEquipmentDurabilitySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got '%s'", sys.genreID)
	}
}

func TestReputationEquipmentDurabilitySystem_CalculateDurabilityModifierNilFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	modifier := sys.calculateDurabilityModifier()
	if modifier != 0.0 {
		t.Errorf("expected 0.0 with nil faction system, got %f", modifier)
	}
}

func TestReputationEquipmentDurabilitySystem_CalculateDurabilityModifierEmptyFactions(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)
	sys.SetGenre("fantasy")

	modifier := sys.calculateDurabilityModifier()
	if modifier != 0.0 {
		t.Errorf("expected 0.0 with no factions, got %f", modifier)
	}
}

func TestReputationEquipmentDurabilitySystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)
	fs := NewFactionSystem(world, nil)
	sys.SetFactionSystem(fs)

	entity := NewEntity(1)
	entity.AddComponent(NewStubInput())

	// Delta below update interval - should skip
	sys.Update([]*Entity{entity}, 0.5)
	if sys.timeSinceCheck != 0.5 {
		t.Errorf("expected timeSinceCheck=0.5, got %f", sys.timeSinceCheck)
	}
}

func TestReputationEquipmentDurabilitySystem_MarkVisualDirtyNoComponent(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	entity := NewEntity(1)
	// No equipment_visual component - should not panic
	sys.markVisualDirty(entity)
}

func TestReputationEquipmentDurabilitySystem_GetEquipmentComponentNone(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilitySystem(world, 42)

	entity := NewEntity(1)
	result := sys.getEquipmentComponent(entity)
	if result != nil {
		t.Error("expected nil for entity without equipment component")
	}
}
