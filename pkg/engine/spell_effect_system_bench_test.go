package engine

import (
	"math/rand"
	"testing"
)

// BenchmarkSpellEffectSystem_Update benchmarks the Update method.
func BenchmarkSpellEffectSystem_Update(b *testing.B) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entities[i] = entity
	}
	world.FlushPendingEntities()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// BenchmarkSpellEffectSystem_UpdateWithEffects benchmarks with active effects.
func BenchmarkSpellEffectSystem_UpdateWithEffects(b *testing.B) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create 100 entities with spell effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		// Add active spell effect
		entity.AddComponent(&SpellEffectComponent{
			EffectType: EffectTerrainManipulation,
			Magnitude:  5.0,
			Duration:   10.0,
			Active:     true,
		})
		entities[i] = entity
	}
	world.FlushPendingEntities()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// BenchmarkSpellEffectSystem_UpdateWithExpiredEffects benchmarks with expired effects.
func BenchmarkSpellEffectSystem_UpdateWithExpiredEffects(b *testing.B) {
	world := NewWorld()
	rng := rand.New(rand.NewSource(12345))
	system := NewSpellEffectSystem(world, rng)

	// Create 100 entities with spell effects that will expire
	entities := make([]*Entity, 100)
	effects := make([]*SpellEffectComponent, 100) // Pre-allocate effects
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		effects[i] = &SpellEffectComponent{
			EffectType:  EffectTerrainManipulation,
			Magnitude:   5.0,
			Duration:    0.001, // Will expire immediately
			ElapsedTime: 0.0,
			Active:      true,
		}
		entities[i] = entity
	}
	world.FlushPendingEntities()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Reset effects and add them back
		for j, entity := range entities {
			effects[j].ElapsedTime = 0.001 // Make it expired
			effects[j].Duration = 0.001
			effects[j].Active = true
			entity.AddComponent(effects[j])
		}
		system.Update(entities, 0.016)
	}
}
