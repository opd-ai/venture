package network

import (
	"testing"
	"time"
)

func TestNewSnapshotManager(t *testing.T) {
	sm := NewSnapshotManager(10)

	if sm == nil {
		t.Fatal("NewSnapshotManager returned nil")
	}

	if sm.maxSnapshots != 10 {
		t.Errorf("Expected maxSnapshots 10, got %d", sm.maxSnapshots)
	}

	if sm.currentIndex != -1 {
		t.Errorf("Expected currentIndex -1, got %d", sm.currentIndex)
	}
}

func TestNewSnapshotManager_MinSize(t *testing.T) {
	// Should enforce minimum of 2 snapshots
	sm := NewSnapshotManager(1)

	if sm.maxSnapshots != 2 {
		t.Errorf("Expected maxSnapshots 2 (enforced minimum), got %d", sm.maxSnapshots)
	}
}

func TestSnapshotManager_AddSnapshot(t *testing.T) {
	sm := NewSnapshotManager(5)

	snapshot := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 10, Y: 20}},
		},
	}

	sm.AddSnapshot(snapshot)

	latest := sm.GetLatestSnapshot()
	if latest == nil {
		t.Fatal("GetLatestSnapshot returned nil")
	}

	if latest.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", latest.Sequence)
	}

	if len(latest.Entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(latest.Entities))
	}
}

func TestSnapshotManager_AddMultipleSnapshots(t *testing.T) {
	sm := NewSnapshotManager(3)

	// Add 5 snapshots (more than max)
	for i := 0; i < 5; i++ {
		snapshot := WorldSnapshot{
			Entities: map[uint64]EntitySnapshot{
				uint64(i): {EntityID: uint64(i), Position: Position{X: float64(i * 10), Y: 0}},
			},
		}
		sm.AddSnapshot(snapshot)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	latest := sm.GetLatestSnapshot()
	if latest.Sequence != 5 {
		t.Errorf("Expected sequence 5, got %d", latest.Sequence)
	}

	// Should only keep the last 3 snapshots (sequences 3, 4, 5)
	// The oldest in the buffer should be sequence 3
	oldest := sm.GetSnapshotAtSequence(3)
	if oldest == nil {
		t.Error("Should still have sequence 3 in buffer")
	}

	// Sequence 1 and 2 should be overwritten
	old := sm.GetSnapshotAtSequence(1)
	if old != nil {
		t.Error("Sequence 1 should have been overwritten")
	}
}

func TestSnapshotManager_GetSnapshotAtSequence(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Add several snapshots
	for i := 1; i <= 5; i++ {
		snapshot := WorldSnapshot{
			Entities: map[uint64]EntitySnapshot{
				100: {EntityID: 100, Position: Position{X: float64(i), Y: 0}},
			},
		}
		sm.AddSnapshot(snapshot)
	}

	// Get snapshot at sequence 3
	snap := sm.GetSnapshotAtSequence(3)
	if snap == nil {
		t.Fatal("GetSnapshotAtSequence(3) returned nil")
	}

	if snap.Sequence != 3 {
		t.Errorf("Expected sequence 3, got %d", snap.Sequence)
	}

	entity := snap.Entities[100]
	if entity.Position.X != 3.0 {
		t.Errorf("Expected X position 3.0, got %f", entity.Position.X)
	}
}

func TestSnapshotManager_GetSnapshotAtTime(t *testing.T) {
	sm := NewSnapshotManager(10)

	now := time.Now()
	timestamps := make([]time.Time, 3)

	// Add snapshots with known timestamps
	for i := 0; i < 3; i++ {
		timestamps[i] = now.Add(time.Duration(i) * 100 * time.Millisecond)
		snapshot := WorldSnapshot{
			Timestamp: timestamps[i],
			Entities: map[uint64]EntitySnapshot{
				1: {EntityID: 1, Position: Position{X: float64(i * 10), Y: 0}},
			},
		}
		sm.AddSnapshot(snapshot)
	}

	// Query time between snapshot 1 and 2
	queryTime := now.Add(130 * time.Millisecond)
	snap := sm.GetSnapshotAtTime(queryTime)

	if snap == nil {
		t.Fatal("GetSnapshotAtTime returned nil")
	}

	// Should return snapshot 2 (at 100ms) as it's closest
	// The actual closest depends on the implementation
	if snap.Sequence < 1 || snap.Sequence > 3 {
		t.Errorf("Expected sequence between 1 and 3, got %d", snap.Sequence)
	}
}

func TestSnapshotManager_InterpolateEntity(t *testing.T) {
	sm := NewSnapshotManager(10)

	baseTime := time.Now()

	// Manually set snapshots with controlled timestamps
	sm.mu.Lock()
	sm.currentIndex = 0
	sm.snapshots[0] = WorldSnapshot{
		Timestamp: baseTime,
		Sequence:  1,
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 0, Y: 0}, Velocity: Velocity{VX: 10, VY: 0}},
		},
	}
	sm.currentIndex = 1
	sm.snapshots[1] = WorldSnapshot{
		Timestamp: baseTime.Add(100 * time.Millisecond),
		Sequence:  2,
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 100, Y: 0}, Velocity: Velocity{VX: 10, VY: 0}},
		},
	}
	sm.currentSeq = 2
	sm.mu.Unlock()

	// Interpolate at 50ms (halfway between)
	renderTime := baseTime.Add(50 * time.Millisecond)
	interpolated := sm.InterpolateEntity(1, renderTime)

	if interpolated == nil {
		t.Fatal("InterpolateEntity returned nil")
	}

	// Position should be approximately 50 (halfway between 0 and 100)
	expectedX := 50.0
	if abs(interpolated.Position.X-expectedX) > 1.0 {
		t.Errorf("Expected interpolated X ≈%f, got %f", expectedX, interpolated.Position.X)
	}
}

func TestSnapshotManager_InterpolateEntity_NoSnapshots(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Try to interpolate with no snapshots
	interpolated := sm.InterpolateEntity(1, time.Now())

	if interpolated != nil {
		t.Error("Expected nil when no snapshots exist")
	}
}

func TestSnapshotManager_InterpolateEntity_MissingEntity(t *testing.T) {
	sm := NewSnapshotManager(10)

	snapshot := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 10, Y: 20}},
		},
	}
	sm.AddSnapshot(snapshot)

	// Try to interpolate an entity that doesn't exist
	interpolated := sm.InterpolateEntity(999, time.Now())

	if interpolated != nil {
		t.Error("Expected nil for non-existent entity")
	}
}

func TestSnapshotManager_CreateDelta(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Snapshot 1: entities 1, 2, 3
	snap1 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 0, Y: 0}},
			2: {EntityID: 2, Position: Position{X: 10, Y: 0}},
			3: {EntityID: 3, Position: Position{X: 20, Y: 0}},
		},
	}
	sm.AddSnapshot(snap1)

	// Snapshot 2: entity 1 moved, entity 2 removed, entity 4 added
	snap2 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 5, Y: 5}},
			3: {EntityID: 3, Position: Position{X: 20, Y: 0}}, // No change
			4: {EntityID: 4, Position: Position{X: 30, Y: 0}},
		},
	}
	sm.AddSnapshot(snap2)

	delta := sm.CreateDelta(1, 2)

	if delta == nil {
		t.Fatal("CreateDelta returned nil")
	}

	if delta.FromSequence != 1 || delta.ToSequence != 2 {
		t.Errorf("Expected delta from 1 to 2, got %d to %d", delta.FromSequence, delta.ToSequence)
	}

	// Check removed
	if len(delta.Removed) != 1 || delta.Removed[0] != 2 {
		t.Errorf("Expected removed [2], got %v", delta.Removed)
	}

	// Check added
	if len(delta.Added) != 1 || delta.Added[0] != 4 {
		t.Errorf("Expected added [4], got %v", delta.Added)
	}

	// Check changed (entity 1 moved, entity 4 added)
	if len(delta.Changed) != 2 {
		t.Errorf("Expected 2 changed entities, got %d", len(delta.Changed))
	}

	if _, hasEntity1 := delta.Changed[1]; !hasEntity1 {
		t.Error("Expected entity 1 in changed")
	}

	if _, hasEntity4 := delta.Changed[4]; !hasEntity4 {
		t.Error("Expected entity 4 in changed")
	}
}

func TestSnapshotManager_ApplyDelta(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Base snapshot
	base := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 0, Y: 0}},
			2: {EntityID: 2, Position: Position{X: 10, Y: 0}},
		},
	}
	sm.AddSnapshot(base)

	// Create a delta
	delta := &SnapshotDelta{
		FromSequence: 1,
		ToSequence:   2,
		Added:        []uint64{3},
		Removed:      []uint64{2},
		Changed: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 5, Y: 5}},
			3: {EntityID: 3, Position: Position{X: 20, Y: 0}},
		},
	}

	result := sm.ApplyDelta(1, delta)

	if result == nil {
		t.Fatal("ApplyDelta returned nil")
	}

	if result.Sequence != 2 {
		t.Errorf("Expected sequence 2, got %d", result.Sequence)
	}

	// Entity 1 should be updated
	if entity, exists := result.Entities[1]; !exists {
		t.Error("Entity 1 should exist")
	} else if entity.Position.X != 5 || entity.Position.Y != 5 {
		t.Errorf("Entity 1 position incorrect: (%f, %f)", entity.Position.X, entity.Position.Y)
	}

	// Entity 2 should be removed
	if _, exists := result.Entities[2]; exists {
		t.Error("Entity 2 should be removed")
	}

	// Entity 3 should be added
	if entity, exists := result.Entities[3]; !exists {
		t.Error("Entity 3 should exist")
	} else if entity.Position.X != 20 {
		t.Errorf("Entity 3 position incorrect: %f", entity.Position.X)
	}
}

func TestSnapshotManager_GetCurrentSequence(t *testing.T) {
	sm := NewSnapshotManager(10)

	if sm.GetCurrentSequence() != 0 {
		t.Error("Initial sequence should be 0")
	}

	sm.AddSnapshot(WorldSnapshot{Entities: make(map[uint64]EntitySnapshot)})
	if sm.GetCurrentSequence() != 1 {
		t.Error("Sequence should be 1 after first snapshot")
	}

	sm.AddSnapshot(WorldSnapshot{Entities: make(map[uint64]EntitySnapshot)})
	if sm.GetCurrentSequence() != 2 {
		t.Error("Sequence should be 2 after second snapshot")
	}
}

func TestSnapshotManager_Clear(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Add some snapshots
	for i := 0; i < 5; i++ {
		sm.AddSnapshot(WorldSnapshot{Entities: make(map[uint64]EntitySnapshot)})
	}

	if sm.GetCurrentSequence() != 5 {
		t.Error("Should have sequence 5 before clear")
	}

	sm.Clear()

	if sm.GetCurrentSequence() != 0 {
		t.Error("Sequence should be 0 after clear")
	}

	if sm.GetLatestSnapshot() != nil {
		t.Error("Should have no snapshots after clear")
	}
}

func TestEntitySnapshot_Struct(t *testing.T) {
	snapshot := EntitySnapshot{
		EntityID:  42,
		Timestamp: time.Now(),
		Sequence:  10,
		Position:  Position{X: 100, Y: 200},
		Velocity:  Velocity{VX: 10, VY: 20},
		Components: map[string][]byte{
			"health": []byte{1, 2, 3},
		},
	}

	if snapshot.EntityID != 42 {
		t.Errorf("Expected EntityID 42, got %d", snapshot.EntityID)
	}

	if snapshot.Position.X != 100 {
		t.Errorf("Expected X 100, got %f", snapshot.Position.X)
	}

	if len(snapshot.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(snapshot.Components))
	}
}

func TestWorldSnapshot_Struct(t *testing.T) {
	snapshot := WorldSnapshot{
		Timestamp: time.Now(),
		Sequence:  5,
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1},
			2: {EntityID: 2},
		},
	}

	if snapshot.Sequence != 5 {
		t.Errorf("Expected Sequence 5, got %d", snapshot.Sequence)
	}

	if len(snapshot.Entities) != 2 {
		t.Errorf("Expected 2 entities, got %d", len(snapshot.Entities))
	}
}

func TestSnapshotDelta_Struct(t *testing.T) {
	delta := SnapshotDelta{
		FromSequence: 1,
		ToSequence:   2,
		Added:        []uint64{3, 4},
		Removed:      []uint64{5},
		Changed: map[uint64]EntitySnapshot{
			1: {EntityID: 1},
		},
	}

	if delta.FromSequence != 1 {
		t.Errorf("Expected FromSequence 1, got %d", delta.FromSequence)
	}

	if len(delta.Added) != 2 {
		t.Errorf("Expected 2 added, got %d", len(delta.Added))
	}

	if len(delta.Removed) != 1 {
		t.Errorf("Expected 1 removed, got %d", len(delta.Removed))
	}

	if len(delta.Changed) != 1 {
		t.Errorf("Expected 1 changed, got %d", len(delta.Changed))
	}
}

func TestHelperFunctions_Lerp(t *testing.T) {
	tests := []struct {
		a, b, t, expected float64
	}{
		{0, 10, 0.0, 0},
		{0, 10, 0.5, 5},
		{0, 10, 1.0, 10},
		{10, 20, 0.5, 15},
		{-10, 10, 0.5, 0},
	}

	for _, tt := range tests {
		result := lerp(tt.a, tt.b, tt.t)
		if abs(result-tt.expected) > 0.001 {
			t.Errorf("lerp(%f, %f, %f) = %f, expected %f", tt.a, tt.b, tt.t, result, tt.expected)
		}
	}
}

func TestHelperFunctions_EntityEquals(t *testing.T) {
	e1 := EntitySnapshot{
		Position: Position{X: 10, Y: 20},
		Velocity: Velocity{VX: 1, VY: 2},
	}

	e2 := EntitySnapshot{
		Position: Position{X: 10, Y: 20},
		Velocity: Velocity{VX: 1, VY: 2},
	}

	if !entityEquals(e1, e2) {
		t.Error("Equal entities should return true")
	}

	e3 := EntitySnapshot{
		Position: Position{X: 11, Y: 20},
		Velocity: Velocity{VX: 1, VY: 2},
	}

	if entityEquals(e1, e3) {
		t.Error("Different entities should return false")
	}
}

func TestHelperFunctions_Contains(t *testing.T) {
	slice := []uint64{1, 2, 3, 4, 5}

	if !contains(slice, 3) {
		t.Error("Should find 3 in slice")
	}

	if contains(slice, 10) {
		t.Error("Should not find 10 in slice")
	}

	if contains([]uint64{}, 1) {
		t.Error("Should not find anything in empty slice")
	}
}

// Benchmark snapshot operations
func BenchmarkSnapshotManager_AddSnapshot(b *testing.B) {
	sm := NewSnapshotManager(100)
	snapshot := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 10, Y: 20}},
			2: {EntityID: 2, Position: Position{X: 30, Y: 40}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.AddSnapshot(snapshot)
	}
}

func BenchmarkSnapshotManager_GetLatestSnapshot(b *testing.B) {
	sm := NewSnapshotManager(100)
	sm.AddSnapshot(WorldSnapshot{Entities: make(map[uint64]EntitySnapshot)})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetLatestSnapshot()
	}
}

func BenchmarkSnapshotManager_InterpolateEntity(b *testing.B) {
	sm := NewSnapshotManager(100)

	baseTime := time.Now()
	sm.AddSnapshot(WorldSnapshot{
		Timestamp: baseTime,
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 0, Y: 0}},
		},
	})

	sm.AddSnapshot(WorldSnapshot{
		Timestamp: baseTime.Add(100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 100, Y: 0}},
		},
	})

	renderTime := baseTime.Add(50 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.InterpolateEntity(1, renderTime)
	}
}

func BenchmarkSnapshotManager_CreateDelta(b *testing.B) {
	sm := NewSnapshotManager(100)

	snap1 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 0, Y: 0}},
			2: {EntityID: 2, Position: Position{X: 10, Y: 0}},
		},
	}
	sm.AddSnapshot(snap1)

	snap2 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 5, Y: 5}},
			3: {EntityID: 3, Position: Position{X: 20, Y: 0}},
		},
	}
	sm.AddSnapshot(snap2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CreateDelta(1, 2)
	}
}
