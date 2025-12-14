package engine

import (
	"testing"
)

// TestQueryDeterminism verifies that query results are deterministic and consistent
// after the allocation optimization.
func TestQueryDeterminism(t *testing.T) {
	world := NewWorld()

	// Create entities with predictable IDs
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
	}
	world.processPendingEntityAdditions()

	// Query multiple times and verify results are identical
	result1 := world.GetEntitiesWith("position", "velocity")
	world.InvalidateQueryCache()
	result2 := world.GetEntitiesWith("position", "velocity")
	world.InvalidateQueryCache()
	result3 := world.GetEntitiesWith("position", "velocity")

	if len(result1) != len(result2) || len(result1) != len(result3) {
		t.Fatalf("Query lengths differ: %d, %d, %d", len(result1), len(result2), len(result3))
	}

	// Verify entity IDs are consistent
	for i := 0; i < len(result1); i++ {
		if result1[i].ID != result2[i].ID || result1[i].ID != result3[i].ID {
			t.Fatalf("Entity ID mismatch at index %d: %d, %d, %d", i, result1[i].ID, result2[i].ID, result3[i].ID)
		}
	}
}

// TestQueryCacheConsistency verifies that cached and uncached queries return identical results.
func TestQueryCacheConsistency(t *testing.T) {
	world := NewWorld()

	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%3 == 0 {
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		}
	}
	world.processPendingEntityAdditions()

	// First query (cache miss)
	world.InvalidateQueryCache()
	uncached := world.GetEntitiesWith("position", "health")

	// Second query (cache hit)
	cached := world.GetEntitiesWith("position", "health")

	if len(uncached) != len(cached) {
		t.Fatalf("Cached vs uncached length mismatch: %d vs %d", len(uncached), len(cached))
	}

	for i := 0; i < len(uncached); i++ {
		if uncached[i].ID != cached[i].ID {
			t.Fatalf("Entity ID mismatch at index %d: uncached=%d, cached=%d", i, uncached[i].ID, cached[i].ID)
		}
	}
}
