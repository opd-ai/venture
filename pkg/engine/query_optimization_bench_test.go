package engine

import (
	"testing"
)

// BenchmarkGetEntitiesWith_CacheMiss benchmarks entity queries on cache misses
// to measure allocation overhead in the filtering path.
func BenchmarkGetEntitiesWith_CacheMiss(b *testing.B) {
	world := NewWorld()

	// Create 1000 entities with various components
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		}
		if i%5 == 0 {
			entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		}
	}
	world.processPendingEntityAdditions()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Force cache miss by invalidating cache each iteration
		world.InvalidateQueryCache()

		// Query entities - this will trigger filterEntitiesByComponents
		_ = world.GetEntitiesWith("position", "velocity")
	}
}

// BenchmarkGetEntitiesWith_MultipleQueries benchmarks multiple different queries
// to measure allocation patterns with various query patterns.
func BenchmarkGetEntitiesWith_MultipleQueries(b *testing.B) {
	world := NewWorld()

	// Create 1000 entities
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		}
	}
	world.processPendingEntityAdditions()

	queries := [][]string{
		{"position"},
		{"position", "velocity"},
		{"position", "health"},
		{"position", "velocity", "health"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		world.InvalidateQueryCache()
		for _, query := range queries {
			_ = world.GetEntitiesWith(query...)
		}
	}
}
