package engine

import (
	"testing"
)

func TestNewReputationSpellDamageBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.spawnInterval != 1.2 {
		t.Errorf("expected spawnInterval 1.2, got %f", sys.spawnInterval)
	}
	if sys.particleCount != 5 {
		t.Errorf("expected particleCount 5, got %d", sys.particleCount)
	}
}

func TestReputationSpellDamageBonusParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("expected genre cyberpunk, got %s", sys.genreID)
	}
}

func TestReputationSpellDamageBonusParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("expected particle system to be set")
	}
}

func TestReputationSpellDamageBonusParticleSystem_SetSpellDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	sds := NewReputationSpellDamageBonusSystem(world, 42)
	sys.SetSpellDamageSystem(sds)
	if sys.spellDamageSystem != sds {
		t.Error("expected spell damage system to be set")
	}
}

func TestReputationSpellDamageBonusParticleSystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	sys.Update([]*Entity{}, 0.5)
	if sys.timeSinceSpawn != 0.5 {
		t.Errorf("expected timeSinceSpawn 0.5, got %f", sys.timeSinceSpawn)
	}
}

func TestReputationSpellDamageBonusParticleSystem_UpdateNilDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	// Should not panic with nil particle/spell damage systems
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationSpellDamageBonusParticleSystem_SkipsNonPlayerEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	sds := NewReputationSpellDamageBonusSystem(world, 42)
	sys.SetSpellDamageSystem(sds)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// Should not panic; entity has no input component so it's skipped
	sys.Update([]*Entity{entity}, 2.0)
}

func TestReputationSpellDamageBonusParticleSystem_SkipsLowBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)
	sds := NewReputationSpellDamageBonusSystem(world, 42)
	sys.SetSpellDamageSystem(sds)
	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&StubInput{})

	// Bonus is 0 (below minBonus), should be skipped
	sys.Update([]*Entity{entity}, 2.0)
}

func TestReputationSpellDamageBonusParticleSystem_GetParticleTypeForGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationSpellDamageBonusParticleSystem(world, 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		pt := sys.getParticleTypeForGenre()
		if pt < 0 {
			t.Errorf("expected valid particle type for genre %s", genre)
		}
	}
}

func TestReputationSpellDamageBonusParticleSystem_NilWorld(t *testing.T) {
	sys := NewReputationSpellDamageBonusParticleSystem(nil, 42)
	if sys == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	sys.Update([]*Entity{}, 2.0)
}
