package engine

import (
	"testing"
)

// BenchmarkQuadtreeRemove benchmarks entity removal from quadtree.
func BenchmarkQuadtreeRemove(b *testing.B) {
	bounds := Bounds{X: 0, Y: 0, Width: 10000, Height: 10000}
	qt := NewQuadtree(bounds, 32)

	// Create and insert test entities
	entities := make([]*Entity, 1000)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i % 100 * 100),
			Y: float64(i / 100 * 100),
		}
		entities[i].AddComponent(pos)
		qt.Insert(entities[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Remove and re-insert to maintain benchmark state
		idx := i % len(entities)
		qt.Remove(entities[idx])
		qt.Insert(entities[idx])
	}
}

// BenchmarkIncrementalUpdate benchmarks incremental updates vs full rebuilds.
func BenchmarkIncrementalUpdate(b *testing.B) {
	entityCounts := []int{50, 200, 500, 1000, 2000}
	movePercentages := []float64{0.05, 0.10, 0.25} // 5%, 10%, 25% of entities move

	for _, entityCount := range entityCounts {
		for _, movePct := range movePercentages {
			moveCount := int(float64(entityCount) * movePct)

			// Benchmark incremental update
			b.Run(benchName("Incremental", entityCount, movePct), func(b *testing.B) {
				benchmarkUpdateStrategy(b, entityCount, moveCount, true)
			})

			// Benchmark full rebuild
			b.Run(benchName("FullRebuild", entityCount, movePct), func(b *testing.B) {
				benchmarkUpdateStrategy(b, entityCount, moveCount, false)
			})
		}
	}
}

// benchName creates a consistent benchmark name.
func benchName(strategy string, entities int, movePct float64) string {
	return sprintf("%s_%dEntities_%.0fPctMove", strategy, entities, movePct*100)
}

// sprintf is a simple string formatter for benchmark names.
func sprintf(format, strategy string, entities int, pct float64) string {
	// Simple concatenation for benchmark names
	pctInt := int(pct)
	return strategy + "_" + itoa(entities) + "Entities_" + itoa(pctInt) + "PctMove"
}

// itoa converts int to string (simple implementation for benchmark names).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	// Handle up to 5 digits
	buf := make([]byte, 0, 5)
	if n < 0 {
		buf = append(buf, '-')
		n = -n
	}

	// Build digits in reverse
	var tmp [5]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}

	buf = append(buf, tmp[i:]...)
	return string(buf)
}

// benchmarkUpdateStrategy benchmarks a specific update strategy.
func benchmarkUpdateStrategy(b *testing.B, entityCount, moveCount int, useIncremental bool) {
	system := NewSpatialPartitionSystem(10000, 10000)
	system.SetIncrementalUpdate(useIncremental)
	system.rebuildEvery = 1 // Update every frame for consistent benchmark

	// Create entities
	entities := make([]*Entity, entityCount)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i%100) * 100,
			Y: float64(i/100) * 100,
		}
		entities[i].AddComponent(pos)
	}

	// Initial rebuild
	system.Rebuild(entities)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Move a subset of entities
		for j := 0; j < moveCount; j++ {
			idx := (i*moveCount + j) % entityCount
			if useIncremental {
				system.TrackEntityMovement(entities[idx])
			} else {
				system.MarkDirty()
			}
			// Simulate movement
			pos := entities[idx].GetPosition()
			pos.X += 10
			if pos.X > 10000 {
				pos.X = 0
			}
		}

		// Trigger update
		system.Update(entities, 0.016)
	}
}

// BenchmarkIncrementalUpdate_SmallMoves benchmarks with minimal entity movement.
func BenchmarkIncrementalUpdate_SmallMoves(b *testing.B) {
	system := NewSpatialPartitionSystem(10000, 10000)
	system.SetIncrementalUpdate(true)
	system.rebuildEvery = 1

	// 2000 entities (target entity count)
	entities := make([]*Entity, 2000)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i%100) * 100,
			Y: float64(i/100) * 100,
		}
		entities[i].AddComponent(pos)
	}

	system.Rebuild(entities)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Move only 5% of entities (100 entities)
		for j := 0; j < 100; j++ {
			idx := (i*100 + j) % 2000
			system.TrackEntityMovement(entities[idx])
			entities[idx].GetPosition().X += 5
		}
		system.Update(entities, 0.016)
	}
}

// BenchmarkFullRebuild_SmallMoves benchmarks full rebuild with minimal movement.
func BenchmarkFullRebuild_SmallMoves(b *testing.B) {
	system := NewSpatialPartitionSystem(10000, 10000)
	system.SetIncrementalUpdate(false)
	system.rebuildEvery = 1

	// 2000 entities (target entity count)
	entities := make([]*Entity, 2000)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i%100) * 100,
			Y: float64(i/100) * 100,
		}
		entities[i].AddComponent(pos)
	}

	system.Rebuild(entities)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Move only 5% of entities (100 entities)
		system.MarkDirty()
		for j := 0; j < 100; j++ {
			idx := (i*100 + j) % 2000
			entities[idx].GetPosition().X += 5
		}
		system.Update(entities, 0.016)
	}
}

// BenchmarkIncrementalUpdate_Memory measures memory allocations.
func BenchmarkIncrementalUpdate_Memory(b *testing.B) {
	system := NewSpatialPartitionSystem(10000, 10000)
	system.SetIncrementalUpdate(true)
	system.rebuildEvery = 1

	entities := make([]*Entity, 500)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i * 20),
			Y: float64(i * 20),
		}
		entities[i].AddComponent(pos)
	}

	system.Rebuild(entities)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Move 10% of entities
		for j := 0; j < 50; j++ {
			idx := (i*50 + j) % 500
			system.TrackEntityMovement(entities[idx])
			entities[idx].GetPosition().X += 10
		}
		system.Update(entities, 0.016)
	}
}

// BenchmarkFullRebuild_Memory measures memory allocations for full rebuild.
func BenchmarkFullRebuild_Memory(b *testing.B) {
	system := NewSpatialPartitionSystem(10000, 10000)
	system.SetIncrementalUpdate(false)
	system.rebuildEvery = 1

	entities := make([]*Entity, 500)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i * 20),
			Y: float64(i * 20),
		}
		entities[i].AddComponent(pos)
	}

	system.Rebuild(entities)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Move 10% of entities
		system.MarkDirty()
		for j := 0; j < 50; j++ {
			idx := (i*50 + j) % 500
			entities[idx].GetPosition().X += 10
		}
		system.Update(entities, 0.016)
	}
}

// BenchmarkQuadtreeRemove_Subdivided benchmarks removal from subdivided tree.
func BenchmarkQuadtreeRemove_Subdivided(b *testing.B) {
	bounds := Bounds{X: 0, Y: 0, Width: 10000, Height: 10000}
	qt := NewQuadtree(bounds, 8) // Force subdivision with low capacity

	// Create many entities to force deep subdivision
	entities := make([]*Entity, 2000)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{
			X: float64(i%100) * 100,
			Y: float64(i/100) * 100,
		}
		entities[i].AddComponent(pos)
		qt.Insert(entities[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(entities)
		qt.Remove(entities[idx])
		qt.Insert(entities[idx])
	}
}
