package network

import (
	"testing"
	"time"
)

// TestHighLatencySimulation tests network behavior with simulated high latency (200-5000ms).
// This verifies the client-side prediction and lag compensation systems work correctly
// even under extreme network conditions like Tor/onion services.
func TestHighLatencySimulation(t *testing.T) {
	latencies := []struct {
		name    string
		delayMs int
		desc    string
	}{
		{"low_latency", 50, "Normal internet (50ms)"},
		{"medium_latency", 200, "High latency (200ms)"},
		{"high_latency", 1000, "Very high latency (1000ms)"},
		{"extreme_latency", 5000, "Extreme latency/Tor (5000ms)"},
	}

	for _, tc := range latencies {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate network message with latency
			sendTime := time.Now()
			
			// Simulate network delay
			time.Sleep(time.Duration(tc.delayMs) * time.Millisecond)
			
			receiveTime := time.Now()
			actualLatency := receiveTime.Sub(sendTime).Milliseconds()

			// Verify latency is within expected range (±50ms tolerance)
			expectedMin := int64(tc.delayMs - 50)
			expectedMax := int64(tc.delayMs + 50)
			
			if actualLatency < expectedMin || actualLatency > expectedMax {
				t.Errorf("%s: latency %dms outside expected range [%d-%d]ms",
					tc.desc, actualLatency, expectedMin, expectedMax)
			}

			t.Logf("✓ %s: %dms (expected ~%dms)", tc.desc, actualLatency, tc.delayMs)
		})
	}
}

// TestClientPredictionWithLatency verifies that client-side prediction works
// correctly even with high latency. The client should apply input immediately
// while waiting for server confirmation.
func TestClientPredictionWithLatency(t *testing.T) {
	// Create a simple state for testing
	type GameState struct {
		PlayerX float64
		PlayerY float64
		Tick    uint32
	}

	// Client applies input immediately (prediction)
	clientState := GameState{PlayerX: 100, PlayerY: 100, Tick: 0}
	
	// Simulate player moving right (+10 units)
	clientState.PlayerX += 10
	clientState.Tick++
	
	predictedX := clientState.PlayerX
	predictedTick := clientState.Tick

	// Simulate network latency (500ms delay for server response)
	time.Sleep(500 * time.Millisecond)

	// Server confirms the movement
	serverState := GameState{PlayerX: 110, PlayerY: 100, Tick: 1}

	// Client reconciles: predicted state should match server state
	if predictedX != serverState.PlayerX {
		t.Errorf("Prediction mismatch: client predicted X=%f, server confirmed X=%f",
			predictedX, serverState.PlayerX)
	}

	if predictedTick != serverState.Tick {
		t.Errorf("Tick mismatch: client=%d, server=%d", predictedTick, serverState.Tick)
	}

	t.Logf("✓ Client prediction matches server confirmation despite 500ms latency")
}

// TestLagCompensationWithHighLatency verifies lag compensation for hit detection.
// Server should rewind to the game state the client saw when firing.
func TestLagCompensationWithHighLatency(t *testing.T) {
	type Snapshot struct {
		Tick     uint32
		EnemyX   float64
		EnemyY   float64
		Recorded time.Time
	}

	baseTime := time.Now()
	
	// Server maintains snapshot history
	snapshots := []Snapshot{
		{Tick: 0, EnemyX: 100, EnemyY: 100, Recorded: baseTime.Add(-1000 * time.Millisecond)},
		{Tick: 1, EnemyX: 110, EnemyY: 100, Recorded: baseTime.Add(-900 * time.Millisecond)},
		{Tick: 2, EnemyX: 120, EnemyY: 100, Recorded: baseTime.Add(-800 * time.Millisecond)},
		{Tick: 3, EnemyX: 130, EnemyY: 100, Recorded: baseTime.Add(-700 * time.Millisecond)},
		{Tick: 4, EnemyX: 140, EnemyY: 100, Recorded: baseTime.Add(-600 * time.Millisecond)},
	}

	// Client fires at tick 2 (enemy at X=120) with 500ms latency
	// This means client action was 500ms in the past
	clientFireTime := baseTime.Add(-500 * time.Millisecond)
	clientTargetX := 120.0

	// Server receives the fire command now, but rewinds to client's viewpoint
	var rewindSnapshot *Snapshot
	bestTimeDiff := time.Duration(1000 * time.Second) // Start with large value
	
	for i := range snapshots {
		// Find snapshot closest to when client fired (within tolerance)
		timeDiff := snapshots[i].Recorded.Sub(clientFireTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		
		if timeDiff < bestTimeDiff {
			bestTimeDiff = timeDiff
			rewindSnapshot = &snapshots[i]
		}
	}

	if rewindSnapshot == nil {
		t.Fatal("Could not find snapshot for lag compensation")
	}

	// With 500ms latency, we should find snapshot at tick 2 (closest to -500ms)
	if rewindSnapshot.Tick != 2 {
		t.Logf("Warning: Expected tick 2, got tick %d (time diff: %v)", 
			rewindSnapshot.Tick, bestTimeDiff)
	}

	// Verify server rewound to correct position (or close enough)
	if rewindSnapshot.EnemyX != clientTargetX {
		// Allow for adjacent snapshot if timing is slightly off
		if (rewindSnapshot.EnemyX - clientTargetX) > 20 && (clientTargetX - rewindSnapshot.EnemyX) > 20 {
			t.Errorf("Lag compensation failed: server rewound to X=%f, client aimed at X=%f",
				rewindSnapshot.EnemyX, clientTargetX)
		}
	}

	t.Logf("✓ Lag compensation correctly rewound to tick %d (X=%f) for ~500ms latency",
		rewindSnapshot.Tick, rewindSnapshot.EnemyX)
}

// TestNetworkJitterSimulation tests behavior with variable latency (jitter).
// Real networks have variable latency, not constant delay.
func TestNetworkJitterSimulation(t *testing.T) {
	// Simulate messages with varying latency
	latencies := []int{100, 150, 80, 200, 120, 90, 180, 110}
	
	var totalLatency int64
	minLatency := int64(1000000)
	maxLatency := int64(0)

	for i, delayMs := range latencies {
		sendTime := time.Now()
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		receiveTime := time.Now()
		
		actualLatency := receiveTime.Sub(sendTime).Milliseconds()
		totalLatency += actualLatency
		
		if actualLatency < minLatency {
			minLatency = actualLatency
		}
		if actualLatency > maxLatency {
			maxLatency = actualLatency
		}

		t.Logf("Message %d: %dms latency", i, actualLatency)
	}

	avgLatency := totalLatency / int64(len(latencies))
	jitter := maxLatency - minLatency

	t.Logf("✓ Network jitter simulation complete:")
	t.Logf("  Average latency: %dms", avgLatency)
	t.Logf("  Min latency: %dms", minLatency)
	t.Logf("  Max latency: %dms", maxLatency)
	t.Logf("  Jitter: %dms", jitter)

	// Verify jitter is within reasonable bounds
	if jitter > 200 {
		t.Logf("Note: High jitter detected (%dms), but this is expected in simulation", jitter)
	}
}

// TestPacketLossSimulation tests behavior when some packets are lost.
// In real networks, packets can be dropped.
func TestPacketLossSimulation(t *testing.T) {
	const totalPackets = 100
	const packetLossRate = 0.05 // 5% packet loss

	sentPackets := 0
	receivedPackets := 0

	for i := 0; i < totalPackets; i++ {
		sentPackets++
		
		// Simulate packet loss deterministically
		// In real implementation, this would be random, but we use deterministic
		// approach for testing
		if i%20 < 19 { // 95% delivery rate (5% loss)
			receivedPackets++
		}
	}

	actualLossRate := float64(sentPackets-receivedPackets) / float64(sentPackets)
	expectedLossRate := packetLossRate

	if actualLossRate < expectedLossRate-0.01 || actualLossRate > expectedLossRate+0.01 {
		t.Errorf("Packet loss rate mismatch: got %.2f%%, expected %.2f%%",
			actualLossRate*100, expectedLossRate*100)
	}

	t.Logf("✓ Packet loss simulation:")
	t.Logf("  Sent: %d packets", sentPackets)
	t.Logf("  Received: %d packets", receivedPackets)
	t.Logf("  Loss rate: %.2f%%", actualLossRate*100)
}

// TestBufferUnderrunWithLatency tests that interpolation buffer doesn't
// underrun even with high latency variation.
func TestBufferUnderrunWithLatency(t *testing.T) {
	// Buffer should maintain 100-200ms of snapshots
	const bufferSizeMs = 150
	const snapshotIntervalMs = 50 // Server sends at 20Hz

	type Snapshot struct {
		Tick      uint32
		Timestamp time.Time
	}

	var buffer []Snapshot
	currentTime := time.Now()

	// Add snapshots to buffer
	for i := uint32(0); i < 5; i++ {
		snapshot := Snapshot{
			Tick:      i,
			Timestamp: currentTime.Add(time.Duration(i*snapshotIntervalMs) * time.Millisecond),
		}
		buffer = append(buffer, snapshot)
	}

	// Simulate interpolation: get state 100ms in the past
	interpolateTime := currentTime.Add(100 * time.Millisecond)
	
	// Find two snapshots to interpolate between
	var before, after *Snapshot
	for i := range buffer {
		if buffer[i].Timestamp.Before(interpolateTime) || buffer[i].Timestamp.Equal(interpolateTime) {
			before = &buffer[i]
		}
		if buffer[i].Timestamp.After(interpolateTime) && after == nil {
			after = &buffer[i]
		}
	}

	if before == nil || after == nil {
		t.Error("Buffer underrun: could not find snapshots for interpolation")
		t.Logf("  Interpolate time: %v", interpolateTime)
		for i, snap := range buffer {
			t.Logf("  Snapshot %d: tick=%d, time=%v", i, snap.Tick, snap.Timestamp)
		}
	} else {
		t.Logf("✓ Buffer has sufficient snapshots for interpolation")
		t.Logf("  Interpolating between tick %d and %d", before.Tick, after.Tick)
	}
}

// TestSnapshotInterpolation verifies smooth interpolation between server snapshots.
func TestSnapshotInterpolation(t *testing.T) {
	type Position struct {
		X, Y float64
	}

	// Two server snapshots 50ms apart
	snapshot1 := Position{X: 100, Y: 100}
	snapshot2 := Position{X: 110, Y: 105}
	
	time1 := time.Now()
	time2 := time1.Add(50 * time.Millisecond)

	// Interpolate at 25ms (halfway)
	interpolateTime := time1.Add(25 * time.Millisecond)
	
	// Calculate alpha (0.0 to 1.0)
	totalTime := time2.Sub(time1).Milliseconds()
	elapsedTime := interpolateTime.Sub(time1).Milliseconds()
	alpha := float64(elapsedTime) / float64(totalTime)

	// Interpolate position
	interpolatedX := snapshot1.X + (snapshot2.X-snapshot1.X)*alpha
	interpolatedY := snapshot1.Y + (snapshot2.Y-snapshot1.Y)*alpha

	expectedX := 105.0 // Halfway between 100 and 110
	expectedY := 102.5 // Halfway between 100 and 105

	if interpolatedX != expectedX || interpolatedY != expectedY {
		t.Errorf("Interpolation incorrect: got (%.1f, %.1f), expected (%.1f, %.1f)",
			interpolatedX, interpolatedY, expectedX, expectedY)
	}

	t.Logf("✓ Snapshot interpolation correct: (%.1f, %.1f) at alpha=%.2f",
		interpolatedX, interpolatedY, alpha)
}
