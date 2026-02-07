// Package engine provides tests for movement system spatial partition optimization.
// This file tests the O(log n) query optimization for entity collision checking.
package engine

import (
	"testing"
)

// TestMovementSystem_QueryNearbyEntities verifies spatial partition optimization.
func TestMovementSystem_QueryNearbyEntities(t *testing.T) {
	tests := []struct {
		name                 string
		withSpatialPartition bool
		entityCount          int
		expectNearbyCount    int
	}{
		{
			name:                 "fallback to full list without spatial partition",
			withSpatialPartition: false,
			entityCount:          100,
			expectNearbyCount:    100, // Should return all entities
		},
		{
			name:                 "query nearby with spatial partition",
			withSpatialPartition: true,
			entityCount:          100,
			expectNearbyCount:    -1, // Variable, depends on entity distribution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			movementSys := NewMovementSystem(200.0)
			collisionSys := NewCollisionSystem(32.0)
			movementSys.SetCollisionSystem(collisionSys)

			var spatialSys *SpatialPartitionSystem
			if tt.withSpatialPartition {
				spatialSys = NewSpatialPartitionSystem(1000, 1000)
				movementSys.SetSpatialPartition(spatialSys)
			}

			// Create test entity with position and collider
			testEntity := world.CreateEntity()
			testEntity.AddComponent(&PositionComponent{X: 500, Y: 500})
			testEntity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

			// Create additional entities spread across the world
			entities := []*Entity{testEntity}
			for i := 0; i < tt.entityCount-1; i++ {
				e := world.CreateEntity()
				x := float64((i % 10) * 100)
				y := float64((i / 10) * 100)
				e.AddComponent(&PositionComponent{X: x, Y: y})
				e.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
				entities = append(entities, e)
			}

			world.Update(0) // Process entity additions

			if tt.withSpatialPartition {
				spatialSys.Rebuild(entities)
			}

			// Query nearby entities
			nearbyEntities := movementSys.queryNearbyEntities(testEntity, 500, 500, entities)

			if !tt.withSpatialPartition {
				// Without spatial partition, should return full list
				if len(nearbyEntities) != tt.expectNearbyCount {
					t.Errorf("Expected %d entities (full list), got %d", tt.expectNearbyCount, len(nearbyEntities))
				}
			} else {
				// With spatial partition, should return fewer entities
				if len(nearbyEntities) >= tt.entityCount {
					t.Errorf("Spatial partition should reduce entity count, got %d (same as total %d)", len(nearbyEntities), tt.entityCount)
				}
				t.Logf("Spatial partition reduced entities from %d to %d (%.1f%% reduction)",
					tt.entityCount, len(nearbyEntities),
					100.0*(1.0-float64(len(nearbyEntities))/float64(tt.entityCount)))
			}
		})
	}
}

// TestMovementSystem_BufferInitialization verifies nearby buffer is properly initialized.
func TestMovementSystem_BufferInitialization(t *testing.T) {
	movementSys := NewMovementSystem(200.0)

	if movementSys.nearbyBuffer == nil {
		t.Error("nearbyBuffer should be initialized")
	}

	if cap(movementSys.nearbyBuffer) < 64 {
		t.Errorf("nearbyBuffer capacity should be at least 64, got %d", cap(movementSys.nearbyBuffer))
	}

	if len(movementSys.nearbyBuffer) != 0 {
		t.Errorf("nearbyBuffer length should be 0, got %d", len(movementSys.nearbyBuffer))
	}
}

// TestMovementSystem_EntityCollisionWithSpatialPartition tests collision checking uses spatial queries.
func TestMovementSystem_EntityCollisionWithSpatialPartition(t *testing.T) {
	world := NewWorld()
	movementSys := NewMovementSystem(200.0)
	collisionSys := NewCollisionSystem(32.0)
	spatialSys := NewSpatialPartitionSystem(1000, 1000)

	movementSys.SetCollisionSystem(collisionSys)
	movementSys.SetSpatialPartition(spatialSys)

	// Create entity that will move
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&VelocityComponent{VX: 50, VY: 0}) // Moving right
	entity1.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true, OffsetX: -16, OffsetY: -16})

	// Create many far-away entities (should be filtered out by spatial query)
	entities := []*Entity{entity1}
	for i := 0; i < 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: 500 + float64(i*50), Y: 500})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entities = append(entities, e)
	}

	world.Update(0) // Process entity additions
	spatialSys.Rebuild(entities)

	// Record starting position
	pos1 := entity1.GetPosition()
	startX := pos1.X

	// Test that queryNearbyEntities reduces the search space
	nearbyEntities := movementSys.queryNearbyEntities(entity1, pos1.X, pos1.Y, entities)

	// Should return far fewer entities than total (spatial filtering)
	if len(nearbyEntities) >= len(entities) {
		t.Errorf("Spatial query should reduce entities, got %d out of %d", len(nearbyEntities), len(entities))
	}

	// Update movement
	movementSys.Update(entities, 0.016)

	// Entity1 should have moved (no obstacles nearby)
	finalX := pos1.X
	if finalX <= startX {
		t.Error("Entity1 should have moved from starting position")
	}

	t.Logf("Spatial query reduced entities from %d to %d (%.1f%% reduction)",
		len(entities), len(nearbyEntities),
		100.0*(1.0-float64(len(nearbyEntities))/float64(len(entities))))
	t.Logf("Entity1 moved from X=%.2f to X=%.2f with spatial partition optimization", startX, finalX)
}

// BenchmarkMovementSystem_WithoutSpatialPartition benchmarks movement without optimization.
func BenchmarkMovementSystem_WithoutSpatialPartition(b *testing.B) {
	world := NewWorld()
	movementSys := NewMovementSystem(200.0)
	collisionSys := NewCollisionSystem(32.0)
	movementSys.SetCollisionSystem(collisionSys)

	// Create entities
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		e := world.CreateEntity()
		x := float64((i % 20) * 50)
		y := float64((i / 20) * 50)
		e.AddComponent(&PositionComponent{X: x, Y: y})
		e.AddComponent(&VelocityComponent{VX: 10, VY: 10})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entities[i] = e
	}

	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movementSys.Update(entities, 0.016)
	}
}

// BenchmarkMovementSystem_WithSpatialPartition benchmarks movement with optimization.
func BenchmarkMovementSystem_WithSpatialPartition(b *testing.B) {
	world := NewWorld()
	movementSys := NewMovementSystem(200.0)
	collisionSys := NewCollisionSystem(32.0)
	spatialSys := NewSpatialPartitionSystem(1000, 1000)

	movementSys.SetCollisionSystem(collisionSys)
	movementSys.SetSpatialPartition(spatialSys)

	// Create entities
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		e := world.CreateEntity()
		x := float64((i % 20) * 50)
		y := float64((i / 20) * 50)
		e.AddComponent(&PositionComponent{X: x, Y: y})
		e.AddComponent(&VelocityComponent{VX: 10, VY: 10})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
		entities[i] = e
	}

	world.Update(0)
	spatialSys.Rebuild(entities)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movementSys.Update(entities, 0.016)
		spatialSys.MarkDirty()
		spatialSys.Update(entities, 0.016)
	}
}

// BenchmarkMovementSystem_AllocationsComparison benchmarks allocation reduction.
func BenchmarkMovementSystem_AllocationsComparison(b *testing.B) {
	benchmarks := []struct {
		name                 string
		withSpatialPartition bool
		entityCount          int
	}{
		{"NoSpatial_50entities", false, 50},
		{"WithSpatial_50entities", true, 50},
		{"NoSpatial_200entities", false, 200},
		{"WithSpatial_200entities", true, 200},
		{"NoSpatial_500entities", false, 500},
		{"WithSpatial_500entities", true, 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			world := NewWorld()
			movementSys := NewMovementSystem(200.0)
			collisionSys := NewCollisionSystem(32.0)
			movementSys.SetCollisionSystem(collisionSys)

			var spatialSys *SpatialPartitionSystem
			if bm.withSpatialPartition {
				spatialSys = NewSpatialPartitionSystem(2000, 2000)
				movementSys.SetSpatialPartition(spatialSys)
			}

			// Create entities
			entities := make([]*Entity, bm.entityCount)
			for i := 0; i < bm.entityCount; i++ {
				e := world.CreateEntity()
				x := float64((i % 20) * 50)
				y := float64((i / 20) * 50)
				e.AddComponent(&PositionComponent{X: x, Y: y})
				e.AddComponent(&VelocityComponent{VX: 10, VY: 10})
				e.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})
				entities[i] = e
			}

			world.Update(0)
			if spatialSys != nil {
				spatialSys.Rebuild(entities)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				movementSys.Update(entities, 0.016)
				if spatialSys != nil {
					spatialSys.MarkDirty()
					spatialSys.Update(entities, 0.016)
				}
			}
		})
	}
}
