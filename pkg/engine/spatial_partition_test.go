package engine

import (
	"testing"
)

func TestBoundsContains(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 100, Height: 100}

	tests := []struct {
		name     string
		x, y     float64
		expected bool
	}{
		{"inside", 50, 50, true},
		{"top-left corner", 0, 0, true},
		{"outside left", -1, 50, false},
		{"outside right", 100, 50, false},
		{"outside top", 50, -1, false},
		{"outside bottom", 50, 100, false},
		{"just inside", 99.9, 99.9, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bounds.Contains(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("Contains(%f, %f) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestBoundsIntersects(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 100, Height: 100}

	tests := []struct {
		name     string
		other    Bounds
		expected bool
	}{
		{"completely inside", Bounds{25, 25, 50, 50}, true},
		{"overlapping", Bounds{50, 50, 100, 100}, true},
		{"adjacent", Bounds{100, 0, 100, 100}, false},
		{"separate", Bounds{200, 200, 100, 100}, false},
		{"same bounds", Bounds{0, 0, 100, 100}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bounds.Intersects(tt.other)
			if result != tt.expected {
				t.Errorf("Intersects() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQuadtreeInsert(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 1000, 1000}, 4)

	// Create test entities
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i * 100),
			Y: float64(i * 100),
		})
		entities[i] = entity

		if !qt.Insert(entity) {
			t.Errorf("Failed to insert entity %d", i)
		}
	}

	if qt.Count() != 10 {
		t.Errorf("Count() = %d, want 10", qt.Count())
	}
}

func TestQuadtreeInsertOutOfBounds(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 100, 100}, 4)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 200, Y: 200})

	if qt.Insert(entity) {
		t.Error("Insert() succeeded for out-of-bounds entity, want failure")
	}
}

func TestQuadtreeQuery(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 1000, 1000}, 4)

	// Insert entities in different regions
	positions := []struct{ x, y float64 }{
		{100, 100}, {200, 200}, {800, 800}, {900, 900},
	}

	for i, pos := range positions {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: pos.x, Y: pos.y})
		qt.Insert(entity)
	}

	// Query top-left quadrant
	result := qt.Query(Bounds{0, 0, 500, 500})
	if len(result) != 2 {
		t.Errorf("Query() returned %d entities, want 2", len(result))
	}

	// Query bottom-right quadrant
	result = qt.Query(Bounds{500, 500, 500, 500})
	if len(result) != 2 {
		t.Errorf("Query() returned %d entities, want 2", len(result))
	}
}

func TestQuadtreeQueryRadius(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 1000, 1000}, 4)

	// Insert entities in a cluster
	positions := []struct{ x, y float64 }{
		{500, 500}, {510, 510}, {520, 520}, {600, 600},
	}

	for i, pos := range positions {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: pos.x, Y: pos.y})
		qt.Insert(entity)
	}

	// Query with small radius
	result := qt.QueryRadius(500, 500, 30)
	if len(result) != 3 {
		t.Errorf("QueryRadius(small) returned %d entities, want 3", len(result))
	}

	// Query with large radius
	result = qt.QueryRadius(500, 500, 200)
	if len(result) != 4 {
		t.Errorf("QueryRadius(large) returned %d entities, want 4", len(result))
	}
}

func TestQuadtreeSubdivision(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 100, 100}, 2) // Small capacity

	// Insert more entities than capacity
	for i := 0; i < 5; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i * 10),
			Y: float64(i * 10),
		})
		qt.Insert(entity)
	}

	if !qt.divided {
		t.Error("Quadtree should be divided after exceeding capacity")
	}

	if qt.Count() != 5 {
		t.Errorf("Count() = %d, want 5", qt.Count())
	}
}

func TestQuadtreeClear(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 100, 100}, 4)

	// Insert entities
	for i := 0; i < 5; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		qt.Insert(entity)
	}

	qt.Clear()

	if qt.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", qt.Count())
	}

	if qt.divided {
		t.Error("Quadtree should not be divided after Clear()")
	}
}

func TestQuadtreeRebuild(t *testing.T) {
	qt := NewQuadtree(Bounds{0, 0, 1000, 1000}, 4)

	// Create entities
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i * 100),
			Y: float64(i * 100),
		})
		entities[i] = entity
	}

	// Rebuild
	qt.Rebuild(entities)

	if qt.Count() != 10 {
		t.Errorf("Count() after Rebuild() = %d, want 10", qt.Count())
	}

	// Verify query works after rebuild
	result := qt.Query(Bounds{0, 0, 500, 500})
	if len(result) < 1 {
		t.Error("Query after Rebuild() returned no results")
	}
}

func TestSpatialPartitionSystem(t *testing.T) {
	sps := NewSpatialPartitionSystem(1000, 1000)

	// Create entities
	entities := make([]*Entity, 20)
	for i := 0; i < 20; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i * 50),
			Y: float64(i * 50),
		})
		entities[i] = entity
	}

	// Update system (builds quadtree) - need to do this before query
	sps.Update(entities, 0.016)
	// Manually rebuild to ensure quadtree is populated
	sps.quadtree.Rebuild(entities)

	// Query
	results := sps.QueryRadius(250, 250, 100)
	if len(results) < 1 {
		t.Error("QueryRadius returned no results")
	}

	// Check statistics
	stats := sps.GetStatistics()
	if stats["entity_count"].(int) != 20 {
		t.Errorf("Statistics entity_count = %v, want 20", stats["entity_count"])
	}
}

func TestSpatialPartitionSystemPeriodicRebuild(t *testing.T) {
	sps := NewSpatialPartitionSystem(1000, 1000)
	sps.rebuildEvery = 5 // Rebuild every 5 frames

	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 50})
		entities[i] = entity
	}

	// Update multiple frames
	for frame := 0; frame < 10; frame++ {
		sps.Update(entities, 0.016)
	}

	// Should have rebuilt at least once
	if sps.frameCount >= sps.rebuildEvery {
		t.Error("Frame count should reset after rebuild")
	}
}

func TestDistanceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		x1, y1   float64
		x2, y2   float64
		expected float64
	}{
		{"zero distance", 0, 0, 0, 0, 0},
		{"unit distance", 0, 0, 1, 0, 1},
		{"pythagorean triple", 0, 0, 3, 4, 5},
		{"negative coords", -10, -10, 0, 0, 14.142135623730951},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Distance(tt.x1, tt.y1, tt.x2, tt.y2)
			if abs(result-tt.expected) > 0.0001 {
				t.Errorf("Distance() = %f, want %f", result, tt.expected)
			}

			// Also test DistanceSquared
			distSq := DistanceSquared(tt.x1, tt.y1, tt.x2, tt.y2)
			expectedSq := tt.expected * tt.expected
			if abs(distSq-expectedSq) > 0.0001 {
				t.Errorf("DistanceSquared() = %f, want %f", distSq, expectedSq)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmarks

func BenchmarkQuadtreeInsert(b *testing.B) {
	qt := NewQuadtree(Bounds{0, 0, 10000, 10000}, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10000),
			Y: float64((i * 7) % 10000),
		})
		qt.Insert(entity)
	}
}

func BenchmarkQuadtreeQuery(b *testing.B) {
	qt := NewQuadtree(Bounds{0, 0, 10000, 10000}, 8)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10000),
			Y: float64((i * 7) % 10000),
		})
		qt.Insert(entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.Query(Bounds{1000, 1000, 2000, 2000})
	}
}

func BenchmarkQuadtreeQueryRadius(b *testing.B) {
	qt := NewQuadtree(Bounds{0, 0, 10000, 10000}, 8)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10000),
			Y: float64((i * 7) % 10000),
		})
		qt.Insert(entity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.QueryRadius(5000, 5000, 500)
	}
}

func BenchmarkQuadtreeRebuild(b *testing.B) {
	qt := NewQuadtree(Bounds{0, 0, 10000, 10000}, 8)

	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10000),
			Y: float64((i * 7) % 10000),
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.Rebuild(entities)
	}
}
