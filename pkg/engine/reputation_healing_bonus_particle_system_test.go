package engine

import (
	"testing"
)

func TestNewReputationHealingBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.spawnInterval != 1.0 {
		t.Errorf("expected spawnInterval 1.0, got %f", sys.spawnInterval)
	}
	if sys.particleCount != 4 {
		t.Errorf("expected particleCount 4, got %d", sys.particleCount)
	}
}

func TestReputationHealingBonusParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("expected genre cyberpunk, got %s", sys.genreID)
	}
}

func TestReputationHealingBonusParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("expected particle system to be set")
	}
}

func TestReputationHealingBonusParticleSystem_SetHealingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	hs := NewReputationHealingBonusSystem(world, 42)
	sys.SetHealingSystem(hs)
	if sys.healingSystem != hs {
		t.Error("expected healing system to be set")
	}
}

func TestReputationHealingBonusParticleSystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	// Should not panic with no dependencies set
	sys.Update([]*Entity{}, 0.5)
	if sys.timeSinceSpawn != 0.5 {
		t.Errorf("expected timeSinceSpawn 0.5, got %f", sys.timeSinceSpawn)
	}
}

func TestReputationHealingBonusParticleSystem_UpdateNoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	hs := NewReputationHealingBonusSystem(world, 42)
	sys.SetHealingSystem(hs)
	// Should not panic with nil particle system
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationHealingBonusParticleSystem_SkipsLowRegenRate(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	hs := NewReputationHealingBonusSystem(world, 42)
	sys.SetHealingSystem(hs)

	entity := NewEntity(1)
	entity.AddComponent(&StubInput{})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// regenRate for entity is 0 (default), so particles should not spawn
	sys.Update([]*Entity{entity}, 2.0)
}

func TestReputationHealingBonusParticleSystem_GetParticleTypeForGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)

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
			sys.SetGenre(tt.genre)
			pt := sys.getParticleTypeForGenre()
			// Just verify it returns a valid type without panicking
			if pt < 0 {
				t.Errorf("expected valid particle type for genre %s", tt.genre)
			}
		})
	}
}

func TestReputationHealingBonusParticleSystem_SkipsNonPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationHealingBonusParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	hs := NewReputationHealingBonusSystem(world, 42)
	hs.regenRates[1] = 2.0
	sys.SetHealingSystem(hs)

	// Entity without input should be skipped
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 2.0)
}
