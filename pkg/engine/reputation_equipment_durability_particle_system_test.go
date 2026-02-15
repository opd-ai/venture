package engine

import (
	"testing"
)

func TestNewReputationEquipmentDurabilityParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.emitInterval != 1.5 {
		t.Errorf("expected emitInterval=1.5, got %f", sys.emitInterval)
	}
}

func TestReputationEquipmentDurabilityParticleSystem_SetDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	durabSys := NewReputationEquipmentDurabilitySystem(world, 42)
	sys.SetDurabilitySystem(durabSys)
	if sys.durabSys != durabSys {
		t.Error("durability system not set")
	}

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got '%s'", sys.genreID)
	}
}

func TestReputationEquipmentDurabilityParticleSystem_UpdateNilDeps(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	// Should not panic with nil dependencies
	sys.Update([]*Entity{}, 2.0)
}

func TestReputationEquipmentDurabilityParticleSystem_UpdateBelowInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	durabSys := NewReputationEquipmentDurabilitySystem(world, 42)
	partSys := NewParticleSystem()
	sys.SetDurabilitySystem(durabSys)
	sys.SetParticleSystem(partSys)

	// Delta below emit interval - should not emit
	entity := NewEntity(1)
	entity.AddComponent(NewStubInput())
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	sys.Update([]*Entity{entity}, 0.5)
	// No panic = pass
}

func TestReputationEquipmentDurabilityParticleSystem_SkipsNonPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	durabSys := NewReputationEquipmentDurabilitySystem(world, 42)
	partSys := NewParticleSystem()
	sys.SetDurabilitySystem(durabSys)
	sys.SetParticleSystem(partSys)

	// Entity without input - should be skipped
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	sys.Update([]*Entity{entity}, 2.0)
	// No panic = pass
}

func TestReputationEquipmentDurabilityParticleSystem_ProtectionParticleTypes(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}
	for _, g := range genres {
		sys.SetGenre(g)
		pt := sys.getProtectionParticleType()
		if pt == 0 {
			t.Errorf("expected non-zero particle type for genre %s", g)
		}
	}
}

func TestReputationEquipmentDurabilityParticleSystem_CorrosionParticleTypes(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}
	for _, g := range genres {
		sys.SetGenre(g)
		pt := sys.getCorrosionParticleType()
		if pt == 0 {
			t.Errorf("expected non-zero particle type for genre %s", g)
		}
	}
}

func TestReputationEquipmentDurabilityParticleSystem_SkipsBelowThreshold(t *testing.T) {
	world := NewWorld()
	sys := NewReputationEquipmentDurabilityParticleSystem(world, 42)

	durabSys := NewReputationEquipmentDurabilitySystem(world, 42)
	partSys := NewParticleSystem()
	sys.SetDurabilitySystem(durabSys)
	sys.SetParticleSystem(partSys)
	sys.SetGenre("fantasy")

	// Entity with no modifier applied (0.0) - should be skipped
	entity := NewEntity(1)
	entity.AddComponent(NewStubInput())
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 2.0)
	// No panic = pass
}
