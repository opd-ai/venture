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

// BenchmarkSpriteAccessGeneric benchmarks generic sprite component access.
func BenchmarkSpriteAccessGeneric(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Generic access pattern (old way)
		if comp, ok := entity.GetComponent("sprite"); ok {
			sprite := comp.(*EbitenSprite)
			_ = sprite.Layer
		}
	}
}

// BenchmarkSpriteAccessTyped benchmarks typed sprite getter.
func BenchmarkSpriteAccessTyped(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})
	entity.AddComponent(NewSpriteComponent(32, 32, nil))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Typed access pattern (new way)
		if sprite := entity.GetSprite(); sprite != nil {
			_ = sprite.Layer
		}
	}
}

// BenchmarkRenderValidationGeneric benchmarks render entity validation with generic access.
func BenchmarkRenderValidationGeneric(b *testing.B) {
	// Create 2000 entities (typical game count)
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(NewSpriteComponent(32, 32, nil))
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate render system validation with generic access
		for _, entity := range entities {
			pos := entity.GetPosition()
			if pos == nil {
				continue
			}
			if comp, ok := entity.GetComponent("sprite"); ok {
				sprite := comp.(*EbitenSprite)
				if sprite.Visible {
					_ = sprite.Layer
				}
			}
		}
	}
}

// BenchmarkRenderValidationTyped benchmarks render entity validation with typed getters.
func BenchmarkRenderValidationTyped(b *testing.B) {
	// Create 2000 entities (typical game count)
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(NewSpriteComponent(32, 32, nil))
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate render system validation with typed getters
		for _, entity := range entities {
			pos := entity.GetPosition()
			sprite := entity.GetSprite()
			if pos != nil && sprite != nil && sprite.Visible {
				_ = sprite.Layer
			}
		}
	}
}

// BenchmarkLayerAccessGeneric benchmarks generic layer component access in collision-like patterns.
func BenchmarkLayerAccessGeneric(b *testing.B) {
	// Create entities with layer components (simulating collision system)
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		layerComp := NewLayerComponent()
		entity.AddComponent(&layerComp)
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate areLayersCompatible with generic access (old way)
		for _, e1 := range entities[:50] {
			for _, e2 := range entities[50:100] {
				layer1Comp, hasLayer1 := e1.GetComponent("layer")
				layer2Comp, hasLayer2 := e2.GetComponent("layer")
				if hasLayer1 && hasLayer2 {
					l1, ok1 := layer1Comp.(*LayerComponent)
					l2, ok2 := layer2Comp.(*LayerComponent)
					if ok1 && ok2 {
						_ = l1.GetEffectiveLayer() == l2.GetEffectiveLayer()
					}
				}
			}
		}
	}
}

// BenchmarkLayerAccessTyped benchmarks typed layer getter in collision-like patterns.
func BenchmarkLayerAccessTyped(b *testing.B) {
	// Create entities with layer components (simulating collision system)
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		layerComp := NewLayerComponent()
		entity.AddComponent(&layerComp)
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate areLayersCompatible with typed getter (new way)
		for _, e1 := range entities[:50] {
			for _, e2 := range entities[50:100] {
				l1 := e1.GetLayer()
				l2 := e2.GetLayer()
				if l1 != nil && l2 != nil {
					_ = l1.GetEffectiveLayer() == l2.GetEffectiveLayer()
				}
			}
		}
	}
}

// BenchmarkParticleEmitterAccessGeneric benchmarks generic particle emitter access in render-like patterns.
func BenchmarkParticleEmitterAccessGeneric(b *testing.B) {
	// Create 2000 entities (typical game count), 100 with particle emitters
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		// Only 5% of entities have particle emitters (typical)
		if i%20 == 0 {
			emitter := NewParticleEmitterComponent(0, DefaultParticleConfig(), 10)
			entity.AddComponent(emitter)
		}
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate drawParticles with generic access (old way)
		for _, entity := range entities {
			comp, ok := entity.GetComponent("particle_emitter")
			if !ok {
				continue
			}
			emitter, ok := comp.(*ParticleEmitterComponent)
			if !ok {
				continue
			}
			_ = len(emitter.Systems)
		}
	}
}

// BenchmarkParticleEmitterAccessTyped benchmarks typed particle emitter getter in render-like patterns.
func BenchmarkParticleEmitterAccessTyped(b *testing.B) {
	// Create 2000 entities (typical game count), 100 with particle emitters
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		// Only 5% of entities have particle emitters (typical)
		if i%20 == 0 {
			emitter := NewParticleEmitterComponent(0, DefaultParticleConfig(), 10)
			entity.AddComponent(emitter)
		}
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate drawParticles with typed getter (new way)
		for _, entity := range entities {
			emitter := entity.GetParticleEmitter()
			if emitter == nil {
				continue
			}
			_ = len(emitter.Systems)
		}
	}
}
