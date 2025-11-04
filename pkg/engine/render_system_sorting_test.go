// Package engine provides tests for deterministic sprite layer sorting.
// Tests verify that entities with the same layer value maintain consistent rendering order.
package engine

import (
	"testing"
)

// TestRenderSystem_DeterministicSorting_SameLayer verifies stable sorting for same-layer entities.
func TestRenderSystem_DeterministicSorting_SameLayer(t *testing.T) {
	// Create render system
	renderSystem := &EbitenRenderSystem{}

	// Create world with multiple entities on same layer
	world := NewWorld()
	
	// Create 5 entities with same layer but different IDs and positions
	entityIDs := []uint64{100, 50, 200, 25, 150}
	yPositions := []float64{100.0, 50.0, 200.0, 25.0, 150.0}
	
	entities := make([]*Entity, len(entityIDs))
	for i, id := range entityIDs {
		e := &Entity{ID: id}
		e.components = make(map[string]interface{})
		e.AddComponent(&PositionComponent{X: 100, Y: yPositions[i]})
		// All entities have same layer value (1)
		sprite := &EbitenSprite{Layer: 1}
		e.AddComponent(sprite)
		entities[i] = e
		world.entities = append(world.entities, e)
	}
	
	// Sort multiple times and verify consistency
	sorted1 := renderSystem.sortEntitiesByLayer(entities)
	sorted2 := renderSystem.sortEntitiesByLayer(entities)
	sorted3 := renderSystem.sortEntitiesByLayer(entities)
	
	// Verify same order across all sorts
	if len(sorted1) != len(sorted2) || len(sorted1) != len(sorted3) {
		t.Fatal("Sorted arrays have different lengths")
	}
	
	for i := 0; i < len(sorted1); i++ {
		if sorted1[i].ID != sorted2[i].ID || sorted1[i].ID != sorted3[i].ID {
			t.Errorf("Sort order not deterministic at position %d: run1=%d, run2=%d, run3=%d",
				i, sorted1[i].ID, sorted2[i].ID, sorted3[i].ID)
		}
	}
	
	// Verify entities are sorted by Y position (secondary sort)
	// Expected order by Y: 25.0, 50.0, 100.0, 150.0, 200.0
	// Expected IDs: 25, 50, 100, 150, 200
	expectedIDs := []uint64{25, 50, 100, 150, 200}
	for i := 0; i < len(sorted1); i++ {
		if sorted1[i].ID != expectedIDs[i] {
			t.Errorf("Entity at position %d has ID %d, expected %d", i, sorted1[i].ID, expectedIDs[i])
		}
	}
}

// TestRenderSystem_DeterministicSorting_DifferentLayers verifies primary layer sort works.
func TestRenderSystem_DeterministicSorting_DifferentLayers(t *testing.T) {
	renderSystem := &EbitenRenderSystem{}
	world := NewWorld()
	
	// Create entities with different layers
	// Layer order: 3, 1, 2, 0, 1 (with different Y positions)
	layersAndYPos := []struct {
		layer int
		y     float64
		id    uint64
	}{
		{3, 100.0, 1},
		{1, 50.0, 2},
		{2, 75.0, 3},
		{0, 25.0, 4},
		{1, 150.0, 5}, // Same layer as entity 2, but higher Y
	}
	
	entities := make([]*Entity, len(layersAndYPos))
	for i, spec := range layersAndYPos {
		e := &Entity{ID: spec.id}
		e.components = make(map[string]interface{})
		e.AddComponent(&PositionComponent{X: 100, Y: spec.y})
		sprite := &EbitenSprite{Layer: spec.layer}
		e.AddComponent(sprite)
		entities[i] = e
		world.entities = append(world.entities, e)
	}
	
	// Sort entities
	sorted := renderSystem.sortEntitiesByLayer(entities)
	
	// Expected order by layer (primary), then Y (secondary):
	// Layer 0: ID 4 (y=25)
	// Layer 1: ID 2 (y=50), ID 5 (y=150)
	// Layer 2: ID 3 (y=75)
	// Layer 3: ID 1 (y=100)
	expectedIDs := []uint64{4, 2, 5, 3, 1}
	
	if len(sorted) != len(expectedIDs) {
		t.Fatalf("Expected %d entities, got %d", len(expectedIDs), len(sorted))
	}
	
	for i := 0; i < len(sorted); i++ {
		if sorted[i].ID != expectedIDs[i] {
			t.Errorf("Entity at position %d has ID %d, expected %d", i, sorted[i].ID, expectedIDs[i])
		}
	}
}

// TestRenderSystem_DeterministicSorting_TertiarySort verifies ID-based tertiary sort.
func TestRenderSystem_DeterministicSorting_TertiarySort(t *testing.T) {
	renderSystem := &EbitenRenderSystem{}
	world := NewWorld()
	
	// Create entities with same layer and same Y position
	// Only difference is entity ID
	ids := []uint64{500, 100, 300, 200, 400}
	
	entities := make([]*Entity, len(ids))
	for i, id := range ids {
		e := &Entity{ID: id}
		e.components = make(map[string]interface{})
		e.AddComponent(&PositionComponent{X: 100, Y: 100}) // Same Y for all
		sprite := &EbitenSprite{Layer: 1} // Same layer for all
		e.AddComponent(sprite)
		entities[i] = e
		world.entities = append(world.entities, e)
	}
	
	// Sort entities
	sorted := renderSystem.sortEntitiesByLayer(entities)
	
	// Expected order by ID (tertiary sort): 100, 200, 300, 400, 500
	expectedIDs := []uint64{100, 200, 300, 400, 500}
	
	if len(sorted) != len(expectedIDs) {
		t.Fatalf("Expected %d entities, got %d", len(expectedIDs), len(sorted))
	}
	
	for i := 0; i < len(sorted); i++ {
		if sorted[i].ID != expectedIDs[i] {
			t.Errorf("Entity at position %d has ID %d, expected %d", i, sorted[i].ID, expectedIDs[i])
		}
	}
	
	// Run sort multiple times to verify consistency
	for run := 0; run < 10; run++ {
		sorted = renderSystem.sortEntitiesByLayer(entities)
		for i := 0; i < len(sorted); i++ {
			if sorted[i].ID != expectedIDs[i] {
				t.Errorf("Run %d: Entity at position %d has ID %d, expected %d", 
					run, i, sorted[i].ID, expectedIDs[i])
			}
		}
	}
}

// TestRenderSystem_DeterministicSorting_NoPosition verifies behavior without position component.
func TestRenderSystem_DeterministicSorting_NoPosition(t *testing.T) {
	renderSystem := &EbitenRenderSystem{}
	world := NewWorld()
	
	// Create entities with same layer but no position component
	ids := []uint64{300, 100, 200}
	
	entities := make([]*Entity, len(ids))
	for i, id := range ids {
		e := &Entity{ID: id}
		e.components = make(map[string]interface{})
		// No position component
		sprite := &EbitenSprite{Layer: 1}
		e.AddComponent(sprite)
		entities[i] = e
		world.entities = append(world.entities, e)
	}
	
	// Sort entities
	sorted := renderSystem.sortEntitiesByLayer(entities)
	
	// Should fall back to ID-based sort
	// Expected order: 100, 200, 300
	expectedIDs := []uint64{100, 200, 300}
	
	if len(sorted) != len(expectedIDs) {
		t.Fatalf("Expected %d entities, got %d", len(expectedIDs), len(sorted))
	}
	
	for i := 0; i < len(sorted); i++ {
		if sorted[i].ID != expectedIDs[i] {
			t.Errorf("Entity at position %d has ID %d, expected %d", i, sorted[i].ID, expectedIDs[i])
		}
	}
}
