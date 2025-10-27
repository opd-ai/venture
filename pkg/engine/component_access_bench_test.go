package engine

import (
	"testing"
)

// BenchmarkComponentAccessGeneric benchmarks generic GetComponent + type assertion.
func BenchmarkComponentAccessGeneric(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Generic access pattern (old way)
		if comp, ok := entity.GetComponent("position"); ok {
			pos := comp.(*PositionComponent)
			_ = pos.X
		}
		if comp, ok := entity.GetComponent("velocity"); ok {
			vel := comp.(*VelocityComponent)
			_ = vel.VX
		}
		if comp, ok := entity.GetComponent("health"); ok {
			health := comp.(*HealthComponent)
			_ = health.Current
		}
	}
}

// BenchmarkComponentAccessTyped benchmarks typed getters.
func BenchmarkComponentAccessTyped(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Typed access pattern (new way)
		if pos := entity.GetPosition(); pos != nil {
			_ = pos.X
		}
		if vel := entity.GetVelocity(); vel != nil {
			_ = vel.VX
		}
		if health := entity.GetHealth(); health != nil {
			_ = health.Current
		}
	}
}

// BenchmarkSystemUpdateGeneric simulates a system update with generic component access.
func BenchmarkSystemUpdateGeneric(b *testing.B) {
	// Create 1000 entities
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		entities[i] = entity
	}

	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate movement system with generic access
		for _, entity := range entities {
			if posComp, ok := entity.GetComponent("position"); ok {
				pos := posComp.(*PositionComponent)
				if velComp, ok := entity.GetComponent("velocity"); ok {
					vel := velComp.(*VelocityComponent)
					pos.X += vel.VX * deltaTime
					pos.Y += vel.VY * deltaTime
				}
			}
		}
	}
}

// BenchmarkSystemUpdateTyped simulates a system update with typed getters.
func BenchmarkSystemUpdateTyped(b *testing.B) {
	// Create 1000 entities
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		entities[i] = entity
	}

	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate movement system with typed getters
		for _, entity := range entities {
			pos := entity.GetPosition()
			vel := entity.GetVelocity()
			if pos != nil && vel != nil {
				pos.X += vel.VX * deltaTime
				pos.Y += vel.VY * deltaTime
			}
		}
	}
}
