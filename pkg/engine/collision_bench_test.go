package engine

import (
	"testing"
)

// BenchmarkCollisionGetNearbyEntities benchmarks the getNearbyEntities method
// which is called for every entity with a collider every frame.
func BenchmarkCollisionGetNearbyEntities(b *testing.B) {
	system := NewCollisionSystem(100.0)

	// Create a dense grid of entities (simulating crowded dungeon)
	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%25) * 40,
			Y: float64(i/25) * 40,
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Layer:  0,
		})
		entities[i] = entity
	}

	// Build grid
	system.Update(entities, 0.016)

	// Test entity in the middle
	testEntity := entities[250]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.getNearbyEntities(testEntity)
	}
}

// BenchmarkCollisionSystemUpdate benchmarks full collision system update
func BenchmarkCollisionSystemUpdate(b *testing.B) {
	system := NewCollisionSystem(100.0)

	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%20) * 50,
			Y: float64(i/20) * 50,
		})
		entity.AddComponent(&VelocityComponent{
			VX: 10.0,
			VY: 10.0,
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Layer:  0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// BenchmarkCollisionSystemUpdateWithQuadtree benchmarks collision with quadtree optimization
func BenchmarkCollisionSystemUpdateWithQuadtree(b *testing.B) {
	system := NewCollisionSystem(100.0)

	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%20) * 50,
			Y: float64(i/20) * 50,
		})
		entity.AddComponent(&VelocityComponent{
			VX: 10.0,
			VY: 10.0,
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Layer:  0,
		})
		entities[i] = entity
	}

	// Create and set quadtree
	bounds := Bounds{X: 0, Y: 0, Width: 2000, Height: 2000}
	quadtree := NewQuadtree(bounds, 32)
	quadtree.Rebuild(entities)
	system.SetQuadtree(quadtree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// BenchmarkCollisionSystemUpdateWithQuadtree_Dense benchmarks collision with 500 entities
func BenchmarkCollisionSystemUpdateWithQuadtree_Dense(b *testing.B) {
	system := NewCollisionSystem(100.0)

	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%25) * 40,
			Y: float64(i/25) * 40,
		})
		entity.AddComponent(&VelocityComponent{
			VX: 10.0,
			VY: 10.0,
		})
		entity.AddComponent(&ColliderComponent{
			Width:  32,
			Height: 32,
			Layer:  0,
		})
		entities[i] = entity
	}

	// Create and set quadtree
	bounds := Bounds{X: 0, Y: 0, Width: 2000, Height: 2000}
	quadtree := NewQuadtree(bounds, 32)
	quadtree.Rebuild(entities)
	system.SetQuadtree(quadtree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
