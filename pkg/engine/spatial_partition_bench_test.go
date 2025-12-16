package engine

import (
	"testing"
)

// BenchmarkSpatialPartitionWithoutDirtyTracking measures rebuild performance without optimization
func BenchmarkSpatialPartitionWithoutDirtyTracking(b *testing.B) {
	system := NewSpatialPartitionSystem(1000, 1000)
	system.SetRebuildInterval(1) // Rebuild every frame (old behavior)

	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%50) * 20,
			Y: float64(i/50) * 20,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016) // 60 FPS delta
	}
}

// BenchmarkSpatialPartitionWithDirtyTracking measures rebuild performance with lazy optimization
func BenchmarkSpatialPartitionWithDirtyTracking(b *testing.B) {
	system := NewSpatialPartitionSystem(1000, 1000)
	system.SetRebuildInterval(60) // Check every 60 frames (1 second)

	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%50) * 20,
			Y: float64(i/50) * 20,
		})
		entities[i] = entity
	}

	// Simulate scenario where entities don't move most frames
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark dirty only 10% of the time (typical game scenario)
		if i%10 == 0 {
			system.MarkDirty()
		}
		system.Update(entities, 0.016)
	}
}

// BenchmarkQuadtreeCapacity8vs16 compares capacity=8 vs capacity=16
func BenchmarkQuadtreeCapacity8vs16(b *testing.B) {
	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	entities := make([]*Entity, 1000)

	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
		entities[i] = entity
	}

	b.Run("capacity=8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			qt := NewQuadtree(bounds, 8)
			for _, entity := range entities {
				qt.Insert(entity)
			}
			_ = qt.Query(Bounds{X: 200, Y: 200, Width: 300, Height: 300})
		}
	})

	b.Run("capacity=16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			qt := NewQuadtree(bounds, 16)
			for _, entity := range entities {
				qt.Insert(entity)
			}
			_ = qt.Query(Bounds{X: 200, Y: 200, Width: 300, Height: 300})
		}
	})

	b.Run("capacity=32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			qt := NewQuadtree(bounds, 32)
			for _, entity := range entities {
				qt.Insert(entity)
			}
			_ = qt.Query(Bounds{X: 200, Y: 200, Width: 300, Height: 300})
		}
	})
}

// BenchmarkMovementSystemWithSpatialPartition measures impact of dirty tracking
func BenchmarkMovementSystemWithSpatialPartition(b *testing.B) {
	movementSys := NewMovementSystem(200.0)
	spatialSys := NewSpatialPartitionSystem(1000, 1000)
	movementSys.SetSpatialPartition(spatialSys)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i%10) * 100,
			Y: float64(i/10) * 100,
		})
		entity.AddComponent(&VelocityComponent{
			VX: 50.0,
			VY: 30.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movementSys.Update(entities, 0.016)
	}
}

// BenchmarkQueryRadiusInto benchmarks the zero-allocation radius query method.
func BenchmarkQueryRadiusInto(b *testing.B) {
	qt := NewQuadtree(Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}, 8)

	// Pre-populate with entities in a tight cluster
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: 500 + float64(i%10), // Cluster around (500, 505)
			Y: 500 + float64(i/10),
		})
		qt.Insert(entity)
	}

	// Pre-allocate buffer for zero-allocation queries
	buffer := make([]*Entity, 0, 100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buffer = buffer[:0] // Reset buffer length, keep capacity
		buffer = qt.QueryRadiusInto(505, 505, 50, buffer)
	}
}

// BenchmarkQueryRadius_Comparison compares allocating vs zero-allocation radius queries.
func BenchmarkQueryRadius_Comparison(b *testing.B) {
	qt := NewQuadtree(Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}, 8)

	// Pre-populate with entities in a tight cluster
	for i := 0; i < 100; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: 500 + float64(i%10),
			Y: 500 + float64(i/10),
		})
		qt.Insert(entity)
	}

	b.Run("QueryRadius_allocating", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = qt.QueryRadius(505, 505, 50)
		}
	})

	b.Run("QueryRadiusInto_zero_alloc", func(b *testing.B) {
		buffer := make([]*Entity, 0, 100)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buffer = buffer[:0]
			buffer = qt.QueryRadiusInto(505, 505, 50, buffer)
		}
	})
}
