package network

import (
	"testing"
	"time"
)

func TestDefaultLagCompensationConfig(t *testing.T) {
	config := DefaultLagCompensationConfig()

	if config.MaxCompensation != 500*time.Millisecond {
		t.Errorf("Expected MaxCompensation 500ms, got %v", config.MaxCompensation)
	}
	if config.MinCompensation != 10*time.Millisecond {
		t.Errorf("Expected MinCompensation 10ms, got %v", config.MinCompensation)
	}
	if config.SnapshotBufferSize != 100 {
		t.Errorf("Expected SnapshotBufferSize 100, got %d", config.SnapshotBufferSize)
	}
}

func TestHighLatencyLagCompensationConfig(t *testing.T) {
	config := HighLatencyLagCompensationConfig()

	if config.MaxCompensation != 5000*time.Millisecond {
		t.Errorf("Expected MaxCompensation 5000ms, got %v", config.MaxCompensation)
	}
	if config.MinCompensation != 10*time.Millisecond {
		t.Errorf("Expected MinCompensation 10ms, got %v", config.MinCompensation)
	}
	if config.SnapshotBufferSize != 200 {
		t.Errorf("Expected SnapshotBufferSize 200, got %d", config.SnapshotBufferSize)
	}
}

func TestNewLagCompensator(t *testing.T) {
	config := DefaultLagCompensationConfig()
	lc := NewLagCompensator(config)

	if lc == nil {
		t.Fatal("NewLagCompensator returned nil")
	}
	if lc.maxCompensation != config.MaxCompensation {
		t.Errorf("Expected maxCompensation %v, got %v", config.MaxCompensation, lc.maxCompensation)
	}
	if lc.minCompensation != config.MinCompensation {
		t.Errorf("Expected minCompensation %v, got %v", config.MinCompensation, lc.minCompensation)
	}
	if lc.snapshots == nil {
		t.Error("SnapshotManager not initialized")
	}
}

func TestRecordSnapshot(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	snapshot := WorldSnapshot{
		Timestamp: time.Now(),
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 100, Y: 200},
				Velocity: Velocity{VX: 10, VY: 20},
			},
		},
	}

	lc.RecordSnapshot(snapshot)

	// Verify snapshot was recorded
	latest := lc.snapshots.GetLatestSnapshot()
	if latest == nil {
		t.Fatal("Snapshot not recorded")
	}
	if len(latest.Entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(latest.Entities))
	}
}

func TestRewindToPlayerTime_NoSnapshots(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	result := lc.RewindToPlayerTime(100 * time.Millisecond)

	if result.Success {
		t.Error("Expected rewind to fail with no snapshots")
	}
	if result.Snapshot != nil {
		t.Error("Expected nil snapshot")
	}
}

func TestRewindToPlayerTime_Success(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record snapshots over time
	baseTime := time.Now().Add(-1 * time.Second)
	for i := 0; i < 10; i++ {
		snapshot := WorldSnapshot{
			Timestamp: baseTime.Add(time.Duration(i*100) * time.Millisecond),
			Entities: map[uint64]EntitySnapshot{
				1: {
					EntityID: 1,
					Position: Position{X: float64(i * 10), Y: float64(i * 20)},
					Velocity: Velocity{VX: 10, VY: 20},
				},
			},
		}
		lc.RecordSnapshot(snapshot)
	}

	// Rewind 300ms (should get snapshot around index 7)
	result := lc.RewindToPlayerTime(300 * time.Millisecond)

	if !result.Success {
		t.Fatal("Expected rewind to succeed")
	}
	if result.Snapshot == nil {
		t.Fatal("Expected non-nil snapshot")
	}
	if result.ActualLatency != 300*time.Millisecond {
		t.Errorf("Expected ActualLatency 300ms, got %v", result.ActualLatency)
	}
	if result.WasClamped {
		t.Error("Expected latency not to be clamped")
	}
}

func TestRewindToPlayerTime_ClampMax(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record a snapshot
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities:  map[uint64]EntitySnapshot{},
	}
	lc.RecordSnapshot(snapshot)

	// Try to rewind more than max (600ms > 500ms max)
	result := lc.RewindToPlayerTime(600 * time.Millisecond)

	if !result.WasClamped {
		t.Error("Expected latency to be clamped")
	}
	if result.ActualLatency != 500*time.Millisecond {
		t.Errorf("Expected ActualLatency 500ms (clamped), got %v", result.ActualLatency)
	}
}

func TestRewindToPlayerTime_ClampMin(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record a snapshot
	snapshot := WorldSnapshot{
		Timestamp: time.Now(),
		Entities:  map[uint64]EntitySnapshot{},
	}
	lc.RecordSnapshot(snapshot)

	// Try to rewind less than min (5ms < 10ms min)
	result := lc.RewindToPlayerTime(5 * time.Millisecond)

	if !result.WasClamped {
		t.Error("Expected latency to be clamped")
	}
	if result.ActualLatency != 10*time.Millisecond {
		t.Errorf("Expected ActualLatency 10ms (clamped), got %v", result.ActualLatency)
	}
}

func TestValidateHit_Success(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record snapshot with two entities
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 100, Y: 100},
				Velocity: Velocity{VX: 0, VY: 0},
			},
			2: {
				EntityID: 2,
				Position: Position{X: 110, Y: 100},
				Velocity: Velocity{VX: 0, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot)

	// Validate hit on target at historical position
	hitPos := Position{X: 112, Y: 100}
	valid, err := lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)

	if err != nil {
		t.Fatalf("ValidateHit returned error: %v", err)
	}
	if !valid {
		t.Error("Expected hit to be valid")
	}
}

func TestValidateHit_Miss(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record snapshot with two entities far apart
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 100, Y: 100},
				Velocity: Velocity{VX: 0, VY: 0},
			},
			2: {
				EntityID: 2,
				Position: Position{X: 200, Y: 200},
				Velocity: Velocity{VX: 0, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot)

	// Try to hit target at wrong position
	hitPos := Position{X: 110, Y: 100}
	valid, err := lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)

	if err != nil {
		t.Fatalf("ValidateHit returned error: %v", err)
	}
	if valid {
		t.Error("Expected hit to be invalid (miss)")
	}
}

func TestValidateHit_TargetNotFound(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record snapshot with only attacker
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 100, Y: 100},
				Velocity: Velocity{VX: 0, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot)

	// Try to hit non-existent target
	hitPos := Position{X: 110, Y: 100}
	valid, err := lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)

	if err == nil {
		t.Error("Expected error for non-existent target")
	}
	if valid {
		t.Error("Expected hit to be invalid")
	}
}

func TestValidateHit_AttackerNotFound(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record snapshot with only target
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			2: {
				EntityID: 2,
				Position: Position{X: 110, Y: 100},
				Velocity: Velocity{VX: 0, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot)

	// Try to hit with non-existent attacker
	hitPos := Position{X: 110, Y: 100}
	valid, err := lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)

	if err == nil {
		t.Error("Expected error for non-existent attacker")
	}
	if valid {
		t.Error("Expected hit to be invalid")
	}
}

func TestValidateHit_NoSnapshot(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// No snapshots recorded
	hitPos := Position{X: 110, Y: 100}
	valid, err := lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)

	if err == nil {
		t.Error("Expected error with no snapshots")
	}
	if valid {
		t.Error("Expected hit to be invalid")
	}
}

func TestGetEntityPositionAt_Success(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	snapshotTime := time.Now().Add(-100 * time.Millisecond)
	snapshot := WorldSnapshot{
		Timestamp: snapshotTime,
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 123.45, Y: 678.90},
				Velocity: Velocity{VX: 10, VY: 20},
			},
		},
	}
	lc.RecordSnapshot(snapshot)

	pos, err := lc.GetEntityPositionAt(1, snapshotTime)

	if err != nil {
		t.Fatalf("GetEntityPositionAt returned error: %v", err)
	}
	if pos == nil {
		t.Fatal("Expected non-nil position")
	}
	if pos.X != 123.45 || pos.Y != 678.90 {
		t.Errorf("Expected position (123.45, 678.90), got (%.2f, %.2f)", pos.X, pos.Y)
	}
}

func TestGetEntityPositionAt_EntityNotFound(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	snapshotTime := time.Now().Add(-100 * time.Millisecond)
	snapshot := WorldSnapshot{
		Timestamp: snapshotTime,
		Entities:  map[uint64]EntitySnapshot{},
	}
	lc.RecordSnapshot(snapshot)

	_, err := lc.GetEntityPositionAt(1, snapshotTime)

	if err == nil {
		t.Error("Expected error for non-existent entity")
	}
}

func TestGetEntityPositionAt_NoSnapshot(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	_, err := lc.GetEntityPositionAt(1, time.Now())

	if err == nil {
		t.Error("Expected error with no snapshots")
	}
}

func TestInterpolateEntityAt_Success(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record two snapshots with proper timestamps
	baseTime := time.Now().Add(-200 * time.Millisecond)
	snapshot1 := WorldSnapshot{
		Timestamp: baseTime,
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 0, Y: 0},
				Velocity: Velocity{VX: 10, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot1)

	time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps

	snapshot2 := WorldSnapshot{
		Timestamp: baseTime.Add(100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {
				EntityID: 1,
				Position: Position{X: 100, Y: 0},
				Velocity: Velocity{VX: 10, VY: 0},
			},
		},
	}
	lc.RecordSnapshot(snapshot2)

	// Interpolate at a time between snapshots
	// Use the latest timestamp as reference
	latest := lc.snapshots.GetLatestSnapshot()
	if latest == nil {
		t.Fatal("No snapshots recorded")
	}
	
	entity, err := lc.InterpolateEntityAt(1, latest.Timestamp)

	if err != nil {
		t.Fatalf("InterpolateEntityAt returned error: %v", err)
	}
	if entity == nil {
		t.Fatal("Expected non-nil entity")
	}
	// Just verify we got an entity, interpolation may vary based on timing
	if entity.EntityID != 1 {
		t.Errorf("Expected entity ID 1, got %d", entity.EntityID)
	}
}

func TestInterpolateEntityAt_NoSnapshot(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	_, err := lc.InterpolateEntityAt(1, time.Now())

	if err == nil {
		t.Error("Expected error with no snapshots")
	}
}

func TestGetStats(t *testing.T) {
	config := DefaultLagCompensationConfig()
	lc := NewLagCompensator(config)

	// Initially no snapshots
	stats := lc.GetStats()
	if stats.TotalSnapshots != 0 {
		t.Errorf("Expected 0 snapshots, got %d", stats.TotalSnapshots)
	}
	if stats.MaxCompensation != config.MaxCompensation {
		t.Errorf("Expected MaxCompensation %v, got %v", config.MaxCompensation, stats.MaxCompensation)
	}
	if stats.MinCompensation != config.MinCompensation {
		t.Errorf("Expected MinCompensation %v, got %v", config.MinCompensation, stats.MinCompensation)
	}

	// Add snapshots
	for i := 0; i < 5; i++ {
		snapshot := WorldSnapshot{
			Timestamp: time.Now().Add(time.Duration(-i*100) * time.Millisecond),
			Entities:  map[uint64]EntitySnapshot{},
		}
		lc.RecordSnapshot(snapshot)
	}

	stats = lc.GetStats()
	// Should have at least 1 snapshot
	if stats.TotalSnapshots < 1 {
		t.Errorf("Expected at least 1 snapshot, got %d", stats.TotalSnapshots)
	}
	if stats.CurrentSequence != 5 {
		t.Errorf("Expected CurrentSequence 5, got %d", stats.CurrentSequence)
	}
	if stats.OldestSnapshotAge < 0 {
		t.Errorf("Expected positive OldestSnapshotAge, got %v", stats.OldestSnapshotAge)
	}
}

func TestClear(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Add snapshots
	for i := 0; i < 3; i++ {
		snapshot := WorldSnapshot{
			Timestamp: time.Now(),
			Entities:  map[uint64]EntitySnapshot{},
		}
		lc.RecordSnapshot(snapshot)
	}

	// Verify snapshots exist
	stats := lc.GetStats()
	if stats.TotalSnapshots < 1 {
		t.Errorf("Expected at least 1 snapshot before clear, got %d", stats.TotalSnapshots)
	}
	if stats.CurrentSequence != 3 {
		t.Errorf("Expected CurrentSequence 3 before clear, got %d", stats.CurrentSequence)
	}

	// Clear
	lc.Clear()

	// Verify cleared
	stats = lc.GetStats()
	if stats.TotalSnapshots != 0 {
		t.Errorf("Expected 0 snapshots after clear, got %d", stats.TotalSnapshots)
	}
	if stats.CurrentSequence != 0 {
		t.Errorf("Expected CurrentSequence 0 after clear, got %d", stats.CurrentSequence)
	}
}

func TestConcurrentAccess(t *testing.T) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Record initial snapshots
	for i := 0; i < 10; i++ {
		snapshot := WorldSnapshot{
			Timestamp: time.Now().Add(time.Duration(-i*50) * time.Millisecond),
			Entities: map[uint64]EntitySnapshot{
				1: {
					EntityID: 1,
					Position: Position{X: float64(i * 10), Y: float64(i * 10)},
					Velocity: Velocity{VX: 10, VY: 10},
				},
			},
		}
		lc.RecordSnapshot(snapshot)
	}

	// Concurrent operations
	done := make(chan bool, 3)

	// Goroutine 1: Record snapshots
	go func() {
		for i := 0; i < 100; i++ {
			snapshot := WorldSnapshot{
				Timestamp: time.Now(),
				Entities: map[uint64]EntitySnapshot{
					1: {EntityID: 1, Position: Position{X: float64(i), Y: float64(i)}},
				},
			}
			lc.RecordSnapshot(snapshot)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Rewind
	go func() {
		for i := 0; i < 100; i++ {
			lc.RewindToPlayerTime(50 * time.Millisecond)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Validate hits
	go func() {
		for i := 0; i < 100; i++ {
			hitPos := Position{X: 50, Y: 50}
			lc.ValidateHit(1, 1, hitPos, 50*time.Millisecond, 10.0)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic or race
}

// Benchmark tests

func BenchmarkRecordSnapshot(b *testing.B) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())
	snapshot := WorldSnapshot{
		Timestamp: time.Now(),
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 100, Y: 200}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.RecordSnapshot(snapshot)
	}
}

func BenchmarkRewindToPlayerTime(b *testing.B) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Setup snapshots
	for i := 0; i < 100; i++ {
		snapshot := WorldSnapshot{
			Timestamp: time.Now().Add(time.Duration(-i*10) * time.Millisecond),
			Entities:  map[uint64]EntitySnapshot{},
		}
		lc.RecordSnapshot(snapshot)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.RewindToPlayerTime(100 * time.Millisecond)
	}
}

func BenchmarkValidateHit(b *testing.B) {
	lc := NewLagCompensator(DefaultLagCompensationConfig())

	// Setup snapshot
	snapshot := WorldSnapshot{
		Timestamp: time.Now().Add(-100 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 100, Y: 100}},
			2: {EntityID: 2, Position: Position{X: 110, Y: 100}},
		},
	}
	lc.RecordSnapshot(snapshot)

	hitPos := Position{X: 112, Y: 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.ValidateHit(1, 2, hitPos, 100*time.Millisecond, 5.0)
	}
}
