package network

import (
	"testing"
	"time"
)

// TestSnapshotManager_SequenceWrapAround tests sequence number wrap-around at UINT32_MAX
func TestSnapshotManager_SequenceWrapAround(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Set the sequence to near UINT32_MAX to trigger wrap-around
	sm.mu.Lock()
	sm.currentSeq = 0xFFFFFFF0 // 16 before max
	sm.mu.Unlock()

	// Add snapshots that will wrap around
	for i := 0; i < 20; i++ {
		snap := WorldSnapshot{
			Entities: map[uint64]EntitySnapshot{
				uint64(i): {
					EntityID: uint64(i),
					Position: Position{X: float64(i), Y: float64(i)},
				},
			},
		}
		sm.AddSnapshot(snap)
	}

	// Get current sequence - should have wrapped
	currentSeq := sm.GetCurrentSequence()
	// 0xFFFFFFF0 + 20 wraps around: calculate at runtime to allow overflow
	start := uint32(0xFFFFFFF0)
	expectedSeq := start + 20 // Wraps to 0x00000004
	if currentSeq != expectedSeq {
		t.Errorf("Expected sequence %d (0x%X), got %d (0x%X)", expectedSeq, expectedSeq, currentSeq, currentSeq)
	}

	// Verify we can retrieve snapshots by sequence, including wrapped ones
	// The last snapshot should be at the current sequence
	latestSnap := sm.GetSnapshotAtSequence(currentSeq)
	if latestSnap == nil {
		t.Fatal("Failed to retrieve latest snapshot at current sequence")
	}
	if latestSnap.Sequence != currentSeq {
		t.Errorf("Latest snapshot has wrong sequence: expected %d, got %d", currentSeq, latestSnap.Sequence)
	}

	// Test retrieving a snapshot from before the wrap (should still be in buffer)
	// With buffer size 10, we should be able to get the last 10 snapshots
	oldSeq := uint32(0xFFFFFFF0 + 11) // 11th snapshot (might still be in buffer)
	oldSnap := sm.GetSnapshotAtSequence(oldSeq)
	if oldSnap != nil && oldSnap.Sequence != oldSeq {
		t.Errorf("Old snapshot has wrong sequence: expected %d, got %d", oldSeq, oldSnap.Sequence)
	}

	// Test retrieving snapshots across the wrap boundary
	wrapSeq := uint32(0x00000002) // A few after wrap (0xFFFFFFF0 + 18 = 0x00000002)
	wrapSnap := sm.GetSnapshotAtSequence(wrapSeq)
	if wrapSnap == nil {
		t.Error("Failed to retrieve snapshot after wrap-around")
	} else if wrapSnap.Sequence != wrapSeq {
		t.Errorf("Wrapped snapshot has wrong sequence: expected %d, got %d", wrapSeq, wrapSnap.Sequence)
	}
}

// TestSnapshotManager_WrapAroundEdgeCases tests edge cases around UINT32_MAX
func TestSnapshotManager_WrapAroundEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		startSeq   uint32
		addCount   int
		checkSeqs  []uint32 // Sequences we expect to find
	}{
		{
			name:       "wrap at exact UINT32_MAX",
			bufferSize: 10,
			startSeq:   0xFFFFFFFF,
			addCount:   5,
			checkSeqs:  []uint32{0xFFFFFFFF, 0x00000000, 0x00000001, 0x00000002, 0x00000003}, // All 5 should be in buffer
		},
		{
			name:       "wrap just before UINT32_MAX",
			bufferSize: 10,
			startSeq:   0xFFFFFFFE,
			addCount:   3,
			checkSeqs:  []uint32{0xFFFFFFFE, 0xFFFFFFFF, 0x00000000}, // All 3 should fit
		},
		{
			name:       "multiple wraps (edge case)",
			bufferSize: 10,
			startSeq:   0xFFFFFFF8,
			addCount:   12,
			checkSeqs:  []uint32{0x00000001, 0x00000002, 0x00000003}, // Last ones in buffer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSnapshotManager(tt.bufferSize)

			// Set starting sequence
			sm.mu.Lock()
			sm.currentSeq = tt.startSeq - 1 // Will increment to startSeq on first add
			sm.mu.Unlock()

			// Add snapshots
			for i := 0; i < tt.addCount; i++ {
				snap := WorldSnapshot{
					Entities: map[uint64]EntitySnapshot{
						uint64(i): {EntityID: uint64(i)},
					},
				}
				sm.AddSnapshot(snap)
			}

			// Check that expected sequences exist
			for _, checkSeq := range tt.checkSeqs {
				snap := sm.GetSnapshotAtSequence(checkSeq)
				if snap == nil {
					t.Errorf("Failed to retrieve snapshot at sequence %d (0x%X)", checkSeq, checkSeq)
				} else if snap.Sequence != checkSeq {
					t.Errorf("Retrieved snapshot has wrong sequence: expected %d (0x%X), got %d (0x%X)",
						checkSeq, checkSeq, snap.Sequence, snap.Sequence)
				}
			}
		})
	}
}

// TestSnapshotManager_WrapAroundDeltaCompression tests delta compression across wrap
func TestSnapshotManager_WrapAroundDeltaCompression(t *testing.T) {
	sm := NewSnapshotManager(10)

	// Set sequence near max
	sm.mu.Lock()
	sm.currentSeq = 0xFFFFFFFE
	sm.mu.Unlock()

	// Add snapshot before wrap
	snap1 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 10, Y: 10}},
			2: {EntityID: 2, Position: Position{X: 20, Y: 20}},
		},
	}
	sm.AddSnapshot(snap1)
	seq1 := sm.GetCurrentSequence() // Should be 0xFFFFFFFF

	// Add snapshot after wrap
	snap2 := WorldSnapshot{
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 15, Y: 15}}, // Changed
			2: {EntityID: 2, Position: Position{X: 20, Y: 20}}, // Unchanged
			3: {EntityID: 3, Position: Position{X: 30, Y: 30}}, // New
		},
	}
	sm.AddSnapshot(snap2)
	seq2 := sm.GetCurrentSequence() // Should be 0x00000000

	// Verify sequences wrapped correctly
	if seq1 != 0xFFFFFFFF {
		t.Errorf("First sequence should be UINT32_MAX, got %d (0x%X)", seq1, seq1)
	}
	if seq2 != 0x00000000 {
		t.Errorf("Second sequence should be 0, got %d (0x%X)", seq2, seq2)
	}

	// Create delta across wrap boundary
	// First verify both snapshots exist
	snap1Check := sm.GetSnapshotAtSequence(seq1)
	snap2Check := sm.GetSnapshotAtSequence(seq2)
	if snap1Check == nil {
		t.Fatalf("Snapshot at seq %d (0x%X) not found before creating delta", seq1, seq1)
	}
	if snap2Check == nil {
		t.Fatalf("Snapshot at seq %d (0x%X) not found before creating delta", seq2, seq2)
	}

	delta := sm.CreateDelta(seq1, seq2)
	if delta == nil {
		t.Fatal("Failed to create delta across wrap boundary")
	}

	// Verify delta contents
	if delta.FromSequence != seq1 {
		t.Errorf("Delta FromSequence wrong: expected %d, got %d", seq1, delta.FromSequence)
	}
	if delta.ToSequence != seq2 {
		t.Errorf("Delta ToSequence wrong: expected %d, got %d", seq2, delta.ToSequence)
	}

	// Should have 1 added entity (ID 3)
	if len(delta.Added) != 1 || delta.Added[0] != 3 {
		t.Errorf("Expected 1 added entity (ID 3), got %v", delta.Added)
	}

	// Should have changes for entity 1 and 3 (entity 2 unchanged due to epsilon)
	if len(delta.Changed) < 1 {
		t.Error("Expected at least 1 changed entity")
	}
}

// TestLagCompensator_SequenceWrapAroundInStats tests GetStats() with sequence wrap-around
func TestLagCompensator_SequenceWrapAroundInStats(t *testing.T) {
	config := DefaultLagCompensationConfig()
	config.SnapshotBufferSize = 20
	lc := NewLagCompensator(config)

	// Set sequence near UINT32_MAX to trigger wrap
	lc.mu.Lock()
	lc.snapshots.mu.Lock()
	lc.snapshots.currentSeq = 0xFFFFFFF0 // 16 before max
	lc.snapshots.mu.Unlock()
	lc.mu.Unlock()

	// Add snapshots that will wrap around
	baseTime := time.Now()
	for i := 0; i < 25; i++ {
		snapshot := WorldSnapshot{
			Timestamp: baseTime.Add(time.Duration(i) * 50 * time.Millisecond),
			Entities: map[uint64]EntitySnapshot{
				1: {
					EntityID: 1,
					Position: Position{X: float64(i * 10), Y: float64(i * 10)},
				},
			},
		}
		lc.RecordSnapshot(snapshot)
	}

	// Get stats - this previously had the wrap-around bug
	stats := lc.GetStats()

	// Current sequence should have wrapped
	// 0xFFFFFFF0 + 25 wraps around: calculate at runtime to allow overflow
	start := uint32(0xFFFFFFF0)
	expectedSeq := start + 25 // Wraps to 0x00000009
	if stats.CurrentSequence != expectedSeq {
		t.Errorf("Expected CurrentSequence %d (0x%X), got %d (0x%X)",
			expectedSeq, expectedSeq, stats.CurrentSequence, stats.CurrentSequence)
	}

	// Should count snapshots correctly (limited by buffer size)
	if stats.TotalSnapshots > config.SnapshotBufferSize {
		t.Errorf("TotalSnapshots (%d) exceeds buffer size (%d)",
			stats.TotalSnapshots, config.SnapshotBufferSize)
	}

	// Should have a reasonable number of snapshots
	if stats.TotalSnapshots < 1 {
		t.Error("Expected at least 1 snapshot in stats")
	}

	// OldestSnapshotAge should be reasonable (not negative or zero)
	if stats.OldestSnapshotAge <= 0 {
		t.Errorf("OldestSnapshotAge should be positive, got %v", stats.OldestSnapshotAge)
	}
}

// TestLagCompensator_WrapAroundRewind tests rewinding across sequence wrap
func TestLagCompensator_WrapAroundRewind(t *testing.T) {
	config := DefaultLagCompensationConfig()
	config.MaxCompensation = 1 * time.Second
	config.SnapshotBufferSize = 30
	lc := NewLagCompensator(config)

	// Set sequence near wrap
	lc.mu.Lock()
	lc.snapshots.mu.Lock()
	lc.snapshots.currentSeq = 0xFFFFFFFD // 3 before max
	lc.snapshots.mu.Unlock()
	lc.mu.Unlock()

	// Add snapshots before and after wrap
	baseTime := time.Now()

	for i := 0; i < 10; i++ {
		pos := Position{X: float64(i * 5), Y: float64(i * 5)}
		snapshot := WorldSnapshot{
			Timestamp: baseTime.Add(time.Duration(i) * 100 * time.Millisecond),
			Entities: map[uint64]EntitySnapshot{
				1: {EntityID: 1, Position: pos},
			},
		}
		lc.RecordSnapshot(snapshot)
	}

	// Try to rewind to different latencies
	latencies := []time.Duration{
		100 * time.Millisecond,
		300 * time.Millisecond,
		500 * time.Millisecond,
	}

	for _, latency := range latencies {
		result := lc.RewindToPlayerTime(latency)
		if !result.Success {
			t.Errorf("Rewind with latency %v failed", latency)
			continue
		}

		// Verify we got a snapshot
		if result.Snapshot == nil {
			t.Errorf("Rewind with latency %v returned nil snapshot", latency)
			continue
		}

		// Verify snapshot has entity 1
		if _, exists := result.Snapshot.Entities[1]; !exists {
			t.Errorf("Snapshot from rewind with latency %v missing entity 1", latency)
		}
	}
}

// TestLagCompensator_WrapAroundValidateHit tests hit validation across sequence wrap
func TestLagCompensator_WrapAroundValidateHit(t *testing.T) {
	config := DefaultLagCompensationConfig()
	config.MaxCompensation = 500 * time.Millisecond
	lc := NewLagCompensator(config)

	// Set sequence to wrap
	lc.mu.Lock()
	lc.snapshots.mu.Lock()
	lc.snapshots.currentSeq = 0xFFFFFFFE
	lc.snapshots.mu.Unlock()
	lc.mu.Unlock()

	baseTime := time.Now()

	// Add snapshot before wrap (older)
	snapshot1 := WorldSnapshot{
		Timestamp: baseTime.Add(-300 * time.Millisecond),
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 100, Y: 100}}, // Attacker
			2: {EntityID: 2, Position: Position{X: 110, Y: 100}}, // Target
		},
	}
	lc.RecordSnapshot(snapshot1)
	seq1 := lc.snapshots.GetCurrentSequence()

	// Add snapshot after wrap (more recent)
	snapshot2 := WorldSnapshot{
		Timestamp: baseTime,
		Entities: map[uint64]EntitySnapshot{
			1: {EntityID: 1, Position: Position{X: 105, Y: 105}}, // Attacker moved
			2: {EntityID: 2, Position: Position{X: 115, Y: 105}}, // Target moved
		},
	}
	lc.RecordSnapshot(snapshot2)
	seq2 := lc.snapshots.GetCurrentSequence()

	// Verify wrap occurred
	if seq1 != 0xFFFFFFFF {
		t.Errorf("Expected first sequence to be UINT32_MAX, got %d (0x%X)", seq1, seq1)
	}
	if seq2 != 0x00000000 {
		t.Errorf("Expected second sequence to wrap to 0, got %d (0x%X)", seq2, seq2)
	}

	// Validate hit at the first snapshot's position (compensating for 300ms latency)
	hitPos := Position{X: 112, Y: 100} // Close to target's old position
	valid, err := lc.ValidateHit(1, 2, hitPos, 300*time.Millisecond, 5.0)

	if err != nil {
		t.Errorf("ValidateHit returned error: %v", err)
	}
	if !valid {
		t.Error("Expected hit to be valid at target's historical position")
	}
}
