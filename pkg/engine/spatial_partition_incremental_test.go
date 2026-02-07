package engine

import (
	"testing"
)

// TestQuadtreeRemove tests the Remove method for selective entity removal.
func TestQuadtreeRemove(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	qt := NewQuadtree(bounds, 4)

	// Create test entities
	entities := make([]*Entity, 5)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{X: float64(100 + i*100), Y: float64(100 + i*100)}
		entities[i].AddComponent(pos)
	}

	// Insert all entities
	for _, e := range entities {
		if !qt.Insert(e) {
			t.Fatalf("Failed to insert entity %d", e.ID)
		}
	}

	// Verify all inserted
	if count := qt.Count(); count != 5 {
		t.Errorf("Expected 5 entities, got %d", count)
	}

	// Remove middle entity
	removed := qt.Remove(entities[2])
	if !removed {
		t.Error("Expected entity to be removed")
	}

	// Verify count decreased
	if count := qt.Count(); count != 4 {
		t.Errorf("Expected 4 entities after removal, got %d", count)
	}

	// Verify removed entity is not in query results
	queryBounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	results := qt.Query(queryBounds)
	for _, e := range results {
		if e.ID == entities[2].ID {
			t.Error("Removed entity should not appear in query results")
		}
	}

	// Verify other entities still present
	foundCount := 0
	for _, e := range results {
		for i, expected := range entities {
			if i == 2 {
				continue // Skip removed entity
			}
			if e.ID == expected.ID {
				foundCount++
			}
		}
	}
	if foundCount != 4 {
		t.Errorf("Expected 4 remaining entities in results, got %d", foundCount)
	}
}

// TestQuadtreeRemove_NotFound tests removing an entity that doesn't exist.
func TestQuadtreeRemove_NotFound(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	qt := NewQuadtree(bounds, 4)

	entity := NewEntity(1)
	pos := &PositionComponent{X: 500, Y: 500}
	entity.AddComponent(pos)

	// Try to remove entity that was never inserted
	removed := qt.Remove(entity)
	if removed {
		t.Error("Expected false when removing non-existent entity")
	}
}

// TestQuadtreeRemove_OutOfBounds tests removing an entity outside bounds.
func TestQuadtreeRemove_OutOfBounds(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	qt := NewQuadtree(bounds, 4)

	entity := NewEntity(1)
	pos := &PositionComponent{X: 2000, Y: 2000} // Outside bounds
	entity.AddComponent(pos)

	// Try to remove entity outside bounds
	removed := qt.Remove(entity)
	if removed {
		t.Error("Expected false when removing out-of-bounds entity")
	}
}

// TestQuadtreeRemove_WithSubdivision tests removal from subdivided quadtree.
func TestQuadtreeRemove_WithSubdivision(t *testing.T) {
	bounds := Bounds{X: 0, Y: 0, Width: 1000, Height: 1000}
	qt := NewQuadtree(bounds, 2) // Low capacity to force subdivision

	// Create entities that will force subdivision
	entities := make([]*Entity, 10)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		// Spread entities across quadrants
		x := float64(100 + (i%2)*400)
		y := float64(100 + (i/2)*200)
		pos := &PositionComponent{X: x, Y: y}
		entities[i].AddComponent(pos)
		qt.Insert(entities[i])
	}

	// Verify subdivision occurred
	if !qt.divided {
		t.Error("Expected quadtree to subdivide with 10 entities and capacity 2")
	}

	// Remove entities from different quadrants
	removed1 := qt.Remove(entities[0]) // Should be in one quadrant
	removed2 := qt.Remove(entities[5]) // Should be in another quadrant

	if !removed1 || !removed2 {
		t.Error("Expected both entities to be removed successfully")
	}

	if count := qt.Count(); count != 8 {
		t.Errorf("Expected 8 entities after removing 2, got %d", count)
	}
}

// TestSpatialPartitionSystem_IncrementalUpdate tests incremental update functionality.
func TestSpatialPartitionSystem_IncrementalUpdate(t *testing.T) {
	system := NewSpatialPartitionSystem(1000, 1000)

	// Create test entities
	entities := make([]*Entity, 10)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{X: float64(100 + i*50), Y: 100}
		entities[i].AddComponent(pos)
	}

	// Initial rebuild
	system.Rebuild(entities)
	initialCount := system.quadtree.Count()
	if initialCount != 10 {
		t.Errorf("Expected 10 entities, got %d", initialCount)
	}

	// Track movement of a few entities
	system.TrackEntityMovement(entities[0])
	system.TrackEntityMovement(entities[1])
	system.TrackEntityMovement(entities[2])

	// Move the tracked entities
	entities[0].GetPosition().X = 200
	entities[1].GetPosition().X = 250
	entities[2].GetPosition().X = 300

	// Advance frames to trigger incremental update
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify entities are still in quadtree
	if count := system.quadtree.Count(); count != 10 {
		t.Errorf("Expected 10 entities after incremental update, got %d", count)
	}

	// Verify moved entities can be queried at new positions
	results := system.QueryRadius(200, 100, 50)
	found := false
	for _, e := range results {
		if e.ID == entities[0].ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Entity at new position should be findable via query")
	}

	// Check statistics
	stats := system.GetStatistics()
	if stats["incremental_updates"].(int) == 0 {
		t.Error("Expected at least one incremental update")
	}
}

// TestSpatialPartitionSystem_TrackEntityMovement tests movement tracking.
func TestSpatialPartitionSystem_TrackEntityMovement(t *testing.T) {
	system := NewSpatialPartitionSystem(1000, 1000)

	entity := NewEntity(1)
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)

	// Track movement
	system.TrackEntityMovement(entity)

	// Verify entity is tracked
	if len(system.movedEntities) != 1 {
		t.Errorf("Expected 1 tracked entity, got %d", len(system.movedEntities))
	}

	// Verify position is recorded
	if _, exists := system.lastKnownPositions[entity.ID]; !exists {
		t.Error("Expected entity position to be recorded")
	}

	recordedPos := system.lastKnownPositions[entity.ID]
	if recordedPos.X != 100 || recordedPos.Y != 100 {
		t.Errorf("Expected position (100, 100), got (%.0f, %.0f)", recordedPos.X, recordedPos.Y)
	}

	// Verify dirty flag is set
	if !system.IsDirty() {
		t.Error("Expected system to be marked dirty after tracking movement")
	}
}

// TestSpatialPartitionSystem_FullRebuildInterval tests periodic full rebuilds.
func TestSpatialPartitionSystem_FullRebuildInterval(t *testing.T) {
	system := NewSpatialPartitionSystem(1000, 1000)
	system.SetIncrementalUpdate(true)

	entities := make([]*Entity, 5)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{X: float64(100 + i*100), Y: 100}
		entities[i].AddComponent(pos)
	}

	// Initial rebuild
	system.Rebuild(entities)
	initialFrame := system.frameCount // Should be 0

	// Don't track movement - just advance frames to trigger forced rebuild
	// Need to advance fullRebuildInterval (300) frames from last rebuild
	for i := 0; i < 305; i++ {
		system.Update(entities, 0.016)
	}

	// Verify full rebuild occurred
	stats := system.GetStatistics()
	if stats["forced_rebuilds"].(int) == 0 {
		t.Errorf("Expected at least one forced full rebuild after 305 frames (initial frame: %d, current frame: %d, forced: %d, lazy: %d, incremental: %d)",
			initialFrame, system.frameCount, stats["forced_rebuilds"].(int), stats["lazy_rebuilds"].(int), stats["incremental_updates"].(int))
	}
}

// TestSpatialPartitionSystem_SetIncrementalUpdate tests enabling/disabling incremental updates.
func TestSpatialPartitionSystem_SetIncrementalUpdate(t *testing.T) {
	system := NewSpatialPartitionSystem(1000, 1000)

	// Default should be enabled
	if !system.useIncrementalUpdate {
		t.Error("Expected incremental updates to be enabled by default")
	}

	// Disable
	system.SetIncrementalUpdate(false)
	if system.useIncrementalUpdate {
		t.Error("Expected incremental updates to be disabled")
	}

	// Test that tracking falls back to MarkDirty when disabled
	entity := NewEntity(1)
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)

	system.TrackEntityMovement(entity)

	// Should set dirty but not track in map
	if !system.IsDirty() {
		t.Error("Expected system to be marked dirty")
	}
	if len(system.movedEntities) != 0 {
		t.Error("Expected no entities tracked when incremental updates disabled")
	}
}

// TestSpatialPartitionSystem_IncrementalVsFullRebuild compares update strategies.
func TestSpatialPartitionSystem_IncrementalVsFullRebuild(t *testing.T) {
	// Create two identical systems
	systemIncremental := NewSpatialPartitionSystem(1000, 1000)
	systemIncremental.SetIncrementalUpdate(true)

	systemFull := NewSpatialPartitionSystem(1000, 1000)
	systemFull.SetIncrementalUpdate(false)

	// Create test entities
	entities := make([]*Entity, 100)
	for i := range entities {
		entities[i] = NewEntity(uint64(i))
		pos := &PositionComponent{X: float64(i * 10), Y: float64(i * 10)}
		entities[i].AddComponent(pos)
	}

	// Initial rebuild for both
	systemIncremental.Rebuild(entities)
	systemFull.Rebuild(entities)

	// Move a small subset (10%)
	for i := 0; i < 10; i++ {
		systemIncremental.TrackEntityMovement(entities[i])
		systemFull.MarkDirty()
		entities[i].GetPosition().X += 50
	}

	// Update both systems
	for i := 0; i < 60; i++ {
		systemIncremental.Update(entities, 0.016)
		systemFull.Update(entities, 0.016)
	}

	// Verify both have same entity count
	if systemIncremental.quadtree.Count() != systemFull.quadtree.Count() {
		t.Error("Incremental and full rebuild should have same entity count")
	}

	// Verify incremental used incremental updates
	statsIncremental := systemIncremental.GetStatistics()
	if statsIncremental["incremental_updates"].(int) == 0 {
		t.Error("Expected incremental system to use incremental updates")
	}

	// Verify full used full rebuilds
	statsFull := systemFull.GetStatistics()
	if statsFull["lazy_rebuilds"].(int) == 0 {
		t.Error("Expected full system to use full rebuilds")
	}
}
