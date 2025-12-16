package engine

import (
	"testing"
)

// BenchmarkGetRotationCached benchmarks the cached GetRotation() getter
func BenchmarkGetRotationCached(b *testing.B) {
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entity.AddComponent(NewRotationComponent(0.5, 3.0))
		entities[i] = entity
	}

	b.ResetTimer()
	var angle float64
	for i := 0; i < b.N; i++ {
		for _, entity := range entities {
			if rot := entity.GetRotation(); rot != nil {
				angle = rot.Angle
			}
		}
	}
	_ = angle
}

// BenchmarkGetRotationUncached benchmarks the old GetComponent pattern for comparison
func BenchmarkGetRotationUncached(b *testing.B) {
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entity.AddComponent(NewRotationComponent(0.5, 3.0))
		entities[i] = entity
	}

	b.ResetTimer()
	var angle float64
	for i := 0; i < b.N; i++ {
		for _, entity := range entities {
			if rotComp, hasRot := entity.GetComponent("rotation"); hasRot {
				if rot, ok := rotComp.(*RotationComponent); ok {
					angle = rot.Angle
				}
			}
		}
	}
	_ = angle
}

// BenchmarkDetectIntersectionWithRotation benchmarks the optimized collision detection
func BenchmarkDetectIntersectionWithRotation(b *testing.B) {
	system := NewCollisionSystem(32)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entity.AddComponent(NewRotationComponent(float64(i)*0.1, 3.0))
		entities[i] = entity
	}

	b.ResetTimer()
	intersects := false
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(entities)-1; j++ {
			e1, e2 := entities[j], entities[j+1]
			pos1, pos2 := e1.GetPosition(), e2.GetPosition()
			coll1, coll2 := e1.GetCollider(), e2.GetCollider()
			if pos1 != nil && pos2 != nil && coll1 != nil && coll2 != nil {
				intersects = system.detectIntersection(e1, pos1, coll1, e2, pos2, coll2)
			}
		}
	}
	_ = intersects
}

// BenchmarkRenderSyncSpriteState benchmarks the sprite state sync in render system
func BenchmarkRenderSyncSpriteState(b *testing.B) {
	renderSystem := NewRenderSystem(nil)

	entities := make([]*Entity, 1000)
	sprites := make([]*EbitenSprite, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&AnimationComponent{CurrentState: AnimationStateIdle, Facing: DirDown})
		entity.AddComponent(NewRotationComponent(0.5, 3.0))
		sprite := NewSpriteComponent(32, 32, nil)
		entity.AddComponent(sprite)
		entities[i] = entity
		sprites[i] = sprite
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, entity := range entities {
			renderSystem.syncSpriteState(entity, sprites[j])
		}
	}
}
