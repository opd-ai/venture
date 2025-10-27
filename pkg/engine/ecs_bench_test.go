package engine

import (
	"testing"
)

// BenchmarkGetEntitiesWith benchmarks entity query performance.
func BenchmarkGetEntitiesWith(b *testing.B) {
	world := NewWorld()

	// Create 1000 entities with various component combinations
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		}
	}

	// Process pending entity additions
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = world.GetEntitiesWith("position", "velocity")
	}
}

// BenchmarkGetEntitiesWithMultipleQueries benchmarks multiple queries per frame (realistic scenario).
func BenchmarkGetEntitiesWithMultipleQueries(b *testing.B) {
	world := NewWorld()

	// Create 2000 entities
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		}
		if i%5 == 0 {
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		}
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate typical frame with multiple system queries
		_ = world.GetEntitiesWith("position", "velocity")    // Movement system
		_ = world.GetEntitiesWith("position", "collider")    // Collision system
		_ = world.GetEntitiesWith("position", "health")      // Health bar render
		_ = world.GetEntitiesWith("position")                // General position queries
	}
}

// BenchmarkWorldUpdate benchmarks the full world update cycle.
func BenchmarkWorldUpdate(b *testing.B) {
	world := NewWorld()

	// Create 1000 entities
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
	}

	// Add a simple test system
	world.AddSystem(&TestSystem{})

	b.ResetTimer()
	b.ReportAllocs()

	deltaTime := 0.016 // 60 FPS
	for i := 0; i < b.N; i++ {
		world.Update(deltaTime)
	}
}

// TestSystem is a simple system for benchmarking.
type TestSystem struct{}

func (ts *TestSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if pos, ok := entity.GetComponent("position"); ok {
			position := pos.(*PositionComponent)
			if vel, ok := entity.GetComponent("velocity"); ok {
				velocity := vel.(*VelocityComponent)
				position.X += velocity.VX * deltaTime
				position.Y += velocity.VY * deltaTime
			}
		}
	}
}

// BenchmarkGetEntitiesWithCacheHit measures cache hit performance.
func BenchmarkGetEntitiesWithCacheHit(b *testing.B) {
	world := NewWorld()

	// Create 2000 entities
	for i := 0; i < 2000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
	}

	world.Update(0)

	// Prime the cache
	_ = world.GetEntitiesWith("position", "velocity")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Should hit cache every time
		_ = world.GetEntitiesWith("position", "velocity")
	}
}
