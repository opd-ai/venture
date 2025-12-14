package engine

import (
	"math/rand"
	"testing"
)

// BenchmarkStatusEffectSystem_Update benchmarks the optimized Update method
// with buffer reuse to eliminate per-entity allocations.
func BenchmarkStatusEffectSystem_Update(b *testing.B) {
	// Create world and system
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewStatusEffectSystem(world, rng)

	// Create entities with status effects
	entityCount := 100
	entities := make([]*Entity, entityCount)
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{Attack: 10, Defense: 5})

		// Add multiple status effects to each entity
		// Use unique effect types to avoid component key collision
		entity.AddComponent(NewStatusEffectComponent("burning", 5.0, 10.0, 1.0))
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // 60 FPS deltaTime
	}
}

// BenchmarkStatusEffectSystem_UpdateWithExpiredEffects benchmarks the Update method
// when effects are expiring (testing the buffer reuse for expired effects).
func BenchmarkStatusEffectSystem_UpdateWithExpiredEffects(b *testing.B) {
	// Create world and system
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewStatusEffectSystem(world, rng)

	// Create entities with expiring status effects
	entityCount := 100
	entities := make([]*Entity, entityCount)
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{Attack: 10, Defense: 5})

		// Add effect that will expire soon
		effect := NewStatusEffectComponent("burning", 5.0, 0.01, 0.005)
		entity.AddComponent(effect)
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Re-add effects that expired
		for _, entity := range entities {
			if !entity.HasComponent("status_effect") {
				effect := NewStatusEffectComponent("burning", 5.0, 0.01, 0.005)
				entity.AddComponent(effect)
			}
		}
		system.Update(entities, 0.016) // 60 FPS deltaTime
	}
}

// BenchmarkStatusEffectSystem_UpdateNoEffects benchmarks the Update method
// when entities have no status effects (baseline case).
func BenchmarkStatusEffectSystem_UpdateNoEffects(b *testing.B) {
	// Create world and system
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewStatusEffectSystem(world, rng)

	// Create entities without status effects
	entityCount := 100
	entities := make([]*Entity, entityCount)
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // 60 FPS deltaTime
	}
}

// BenchmarkStatusEffectSystem_BaselineWithExpiry simulates the old behavior
// with effects that are expiring to measure allocation reduction.
func BenchmarkStatusEffectSystem_BaselineWithExpiry(b *testing.B) {
	// Create world and system
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewStatusEffectSystem(world, rng)

	// Create entities with nearly-expired effects
	entityCount := 100
	entities := make([]*Entity, entityCount)
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{Attack: 10, Defense: 5})
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Add fresh effects that will expire
		for _, entity := range entities {
			effect := NewStatusEffectComponent("burning", 5.0, 0.001, 0.0001)
			entity.AddComponent(effect)
		}
		// Old behavior: allocate a new slice per entity
		for _, entity := range entities {
			effectsToRemove := system.processStatusEffects(entity, 0.016)
			system.removeExpiredEffects(entity, effectsToRemove)
		}
	}
}

// BenchmarkStatusEffectSystem_OptimizedWithExpiry uses the new Update method
// with effects that are expiring to measure allocation reduction.
func BenchmarkStatusEffectSystem_OptimizedWithExpiry(b *testing.B) {
	// Create world and system
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewStatusEffectSystem(world, rng)

	// Create entities
	entityCount := 100
	entities := make([]*Entity, entityCount)
	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&StatsComponent{Attack: 10, Defense: 5})
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Add fresh effects that will expire
		for _, entity := range entities {
			effect := NewStatusEffectComponent("burning", 5.0, 0.001, 0.0001)
			entity.AddComponent(effect)
		}
		system.Update(entities, 0.016)
	}
}
