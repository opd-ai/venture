package engine

import (
	"testing"
)

func TestNewFearFleeParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFearFleeParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", sys.seed)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
	if sys.emitInterval <= 0 {
		t.Error("emitInterval should be positive")
	}
	if sys.baseCount <= 0 {
		t.Error("baseCount should be positive")
	}
}

func TestFearFleeParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("Particle system not set correctly")
	}
}

func TestFearFleeParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewFearFleeParticleSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestFearFleeParticleSystem_Update_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)

	// Should not panic without particle system
	sys.Update([]*Entity{}, 0.1)
}

func TestFearFleeParticleSystem_Update_NoWorld(t *testing.T) {
	sys := NewFearFleeParticleSystem(nil, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic without world
	sys.Update([]*Entity{}, 0.1)
}

func TestFearFleeParticleSystem_Update_WithFearedEntity(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity with fear effect and flee state
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&VelocityComponent{VX: -50, VY: 0})
	aiComp := &AIComponent{State: AIStateFlee}
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   5.0,
	})

	// First update - accumulate time
	sys.Update([]*Entity{entity}, 0.1)
	if sys.GetActiveCount() != 0 {
		t.Error("Should not emit before interval")
	}

	// Update past interval - should spawn particles
	sys.Update([]*Entity{entity}, 0.2)
	if sys.GetActiveCount() != 1 {
		t.Errorf("Active count = %d, want 1", sys.GetActiveCount())
	}
	if !sys.IsEntityFleeing(entity.ID) {
		t.Error("Entity should be marked as fleeing")
	}
}

func TestFearFleeParticleSystem_Update_NoFearEffect(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity in flee state but without fear effect
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	aiComp := &AIComponent{State: AIStateFlee}
	entity.AddComponent(aiComp)

	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Should not track entity without fear effect")
	}
}

func TestFearFleeParticleSystem_Update_NotFleeing(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity with fear effect but not in flee state
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	aiComp := &AIComponent{State: AIStateIdle}
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   5.0,
	})

	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Should not track entity not in flee state")
	}
}

func TestFearFleeParticleSystem_Update_FearExpired(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	aiComp := &AIComponent{State: AIStateFlee}
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   0.0,
	})

	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Should not track entity with expired fear effect")
	}
}

func TestFearFleeParticleSystem_Update_FearedVariant(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Test "feared" variant (vs "fear")
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	aiComp := &AIComponent{State: AIStateFlee}
	entity.AddComponent(aiComp)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "feared",
		Duration:   5.0,
	})

	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 1 {
		t.Error("Should track entity with 'feared' effect type")
	}
}

func TestFearFleeParticleSystem_Update_StopFleeing(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	aiComp := &AIComponent{State: AIStateFlee}
	entity.AddComponent(aiComp)
	effect := &StatusEffectComponent{
		EffectType: "fear",
		Duration:   5.0,
	}
	entity.AddComponent(effect)

	// Start fleeing
	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 1 {
		t.Fatal("Entity should be active")
	}

	// Stop fleeing (change state)
	aiComp.State = AIStateIdle

	// Next update should remove from active
	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Entity should no longer be active after stopping flee")
	}
}

func TestFearFleeParticleSystem_GetPrimaryParticleType(t *testing.T) {
	tests := []struct {
		genreID      string
		expectNonNil bool
	}{
		{"fantasy", true},
		{"scifi", true},
		{"horror", true},
		{"cyberpunk", true},
		{"postapoc", true},
		{"unknown", true}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			world := NewWorld()
			sys := NewFearFleeParticleSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			// Call private method via reflection or use public test
			// Since getPrimaryParticleType is private, test indirectly
			// by verifying Update doesn't panic
			ps := NewParticleSystem()
			sys.SetParticleSystem(ps)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 200})
			entity.AddComponent(&AIComponent{State: AIStateFlee})
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "fear",
				Duration:   5.0,
			})

			// Should not panic
			sys.Update([]*Entity{entity}, 0.3)
		})
	}
}

func TestFearFleeParticleSystem_EntityNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity without position
	entity := world.CreateEntity()
	entity.AddComponent(&AIComponent{State: AIStateFlee})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   5.0,
	})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Entity without position should not be tracked")
	}
}

func TestFearFleeParticleSystem_EntityNoAI(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entity without AI component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "fear",
		Duration:   5.0,
	})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.3)
	if sys.GetActiveCount() != 0 {
		t.Error("Entity without AI should not be tracked")
	}
}

func TestFearFleeParticleSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entities := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 200})
		entity.AddComponent(&AIComponent{State: AIStateFlee})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "fear",
			Duration:   5.0,
		})
		entities[i] = entity
	}

	sys.Update(entities, 0.3)
	if sys.GetActiveCount() != 3 {
		t.Errorf("Active count = %d, want 3", sys.GetActiveCount())
	}
}

func BenchmarkFearFleeParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create 100 feared entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		entity.AddComponent(&VelocityComponent{VX: -50, VY: 0})
		entity.AddComponent(&AIComponent{State: AIStateFlee})
		entity.AddComponent(&StatusEffectComponent{
			EffectType: "fear",
			Duration:   5.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // ~60fps
	}
}

func BenchmarkFearFleeParticleSystem_UpdateNoFear(b *testing.B) {
	world := NewWorld()
	sys := NewFearFleeParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create 100 entities without fear
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		entity.AddComponent(&AIComponent{State: AIStateIdle})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
