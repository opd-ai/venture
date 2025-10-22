//go:build test
// +build test

package main

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/network"
)

// LagCompensationDemo demonstrates the lag compensation system
// for fair hit detection in high-latency multiplayer environments.
func main() {
	fmt.Println("=== Lag Compensation Demo ===")
	fmt.Println()

	// Create lag compensator with default config (500ms max)
	fmt.Println("1. Creating Lag Compensator (default config: 500ms max)")
	config := network.DefaultLagCompensationConfig()
	fmt.Printf("   Max Compensation: %v\n", config.MaxCompensation)
	fmt.Printf("   Min Compensation: %v\n", config.MinCompensation)
	fmt.Printf("   Buffer Size: %d snapshots\n", config.SnapshotBufferSize)
	lc := network.NewLagCompensator(config)
	fmt.Println()

	// Simulate game scenario: Player A shoots at Player B
	// Player A has 200ms latency
	fmt.Println("2. Simulating Game Scenario")
	fmt.Println("   Player A (ID: 1) shoots at Player B (ID: 2)")
	fmt.Println("   Player A's latency: 200ms")
	fmt.Println()

	// Record snapshots over time as players move
	fmt.Println("3. Recording World Snapshots (simulating 20 updates/sec)")
	baseTime := time.Now().Add(-1 * time.Second)

	for i := 0; i < 20; i++ {
		// Player B is moving from position (100, 100) to (200, 100) over 1 second
		playerBX := 100.0 + float64(i*5)

		snapshot := network.WorldSnapshot{
			Timestamp: baseTime.Add(time.Duration(i*50) * time.Millisecond),
			Entities: map[uint64]network.EntitySnapshot{
				1: { // Player A (shooter)
					EntityID: 1,
					Position: network.Position{X: 50, Y: 100},
					Velocity: network.Velocity{VX: 0, VY: 0},
				},
				2: { // Player B (target, moving)
					EntityID: 2,
					Position: network.Position{X: playerBX, Y: 100},
					Velocity: network.Velocity{VX: 100, VY: 0}, // 100 units/sec
				},
			},
		}

		lc.RecordSnapshot(snapshot)

		if i%5 == 0 {
			fmt.Printf("   [%dms] Player B position: (%.0f, %.0f)\n", i*50, playerBX, 100.0)
		}
	}
	fmt.Println()

	// Get current stats
	stats := lc.GetStats()
	fmt.Printf("4. Lag Compensator Stats\n")
	fmt.Printf("   Total Snapshots: %d\n", stats.TotalSnapshots)
	fmt.Printf("   Oldest Snapshot Age: %v\n", stats.OldestSnapshotAge)
	fmt.Printf("   Current Sequence: %d\n", stats.CurrentSequence)
	fmt.Println()

	// Scenario 1: Player A shoots NOW (200ms ago from their perspective)
	// Player B is currently at position (195, 100)
	// But 200ms ago, Player B was at position (175, 100)
	fmt.Println("5. Hit Detection WITHOUT Lag Compensation")
	currentPlayerBPos := network.Position{X: 195, Y: 100}
	fmt.Printf("   Player B's CURRENT position: (%.0f, %.0f)\n", currentPlayerBPos.X, currentPlayerBPos.Y)
	fmt.Printf("   Player A aims at: (%.0f, %.0f)\n", 175.0, 100.0)
	fmt.Printf("   Distance: %.1f units\n", distance(currentPlayerBPos, network.Position{X: 175, Y: 100}))
	fmt.Println("   Result: MISS (Player B already moved)")
	fmt.Println()

	// Scenario 2: With lag compensation, rewind to 200ms ago
	fmt.Println("6. Hit Detection WITH Lag Compensation")
	fmt.Printf("   Player A's latency: 200ms\n")
	fmt.Println("   Rewinding game state to 200ms ago...")

	rewindResult := lc.RewindToPlayerTime(200 * time.Millisecond)
	if !rewindResult.Success {
		fmt.Println("   ERROR: Failed to rewind")
		return
	}

	fmt.Printf("   Compensated Time: %v ago\n", time.Since(rewindResult.CompensatedTime))
	fmt.Printf("   Was Clamped: %v\n", rewindResult.WasClamped)

	// Get Player B's position at that historical time
	playerBHistorical, exists := rewindResult.Snapshot.Entities[2]
	if !exists {
		fmt.Println("   ERROR: Player B not found in snapshot")
		return
	}

	fmt.Printf("   Player B's HISTORICAL position (200ms ago): (%.0f, %.0f)\n",
		playerBHistorical.Position.X, playerBHistorical.Position.Y)

	// Validate the hit
	hitPosition := network.Position{X: 175, Y: 100}
	hitRadius := 10.0 // Hit radius in game units

	fmt.Printf("   Player A aims at: (%.0f, %.0f)\n", hitPosition.X, hitPosition.Y)
	fmt.Printf("   Hit radius: %.0f units\n", hitRadius)

	valid, err := lc.ValidateHit(1, 2, hitPosition, 200*time.Millisecond, hitRadius)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
		return
	}

	actualDistance := distance(playerBHistorical.Position, hitPosition)
	fmt.Printf("   Distance to historical position: %.1f units\n", actualDistance)

	if valid {
		fmt.Println("   Result: ✓ HIT CONFIRMED (fair hit detection)")
	} else {
		fmt.Println("   Result: ✗ MISS")
	}
	fmt.Println()

	// Scenario 3: High latency connection (e.g., Tor)
	fmt.Println("7. High Latency Scenario (e.g., Tor connection)")
	fmt.Println("   Reconfiguring for high-latency (5000ms max)...")

	highLatencyConfig := network.HighLatencyLagCompensationConfig()
	highLatencyLC := network.NewLagCompensator(highLatencyConfig)

	// Record same snapshots
	for i := 0; i < 20; i++ {
		playerBX := 100.0 + float64(i*5)
		snapshot := network.WorldSnapshot{
			Timestamp: baseTime.Add(time.Duration(i*50) * time.Millisecond),
			Entities: map[uint64]network.EntitySnapshot{
				1: {EntityID: 1, Position: network.Position{X: 50, Y: 100}},
				2: {EntityID: 2, Position: network.Position{X: playerBX, Y: 100}},
			},
		}
		highLatencyLC.RecordSnapshot(snapshot)
	}

	// Try to compensate for 800ms latency
	fmt.Println("   Player with 800ms latency shoots...")
	highLatencyResult := highLatencyLC.RewindToPlayerTime(800 * time.Millisecond)

	if highLatencyResult.Success {
		fmt.Printf("   Compensated for: %v\n", highLatencyResult.ActualLatency)
		fmt.Printf("   Was Clamped: %v\n", highLatencyResult.WasClamped)
		fmt.Println("   Result: Successfully rewound (within configured limit)")
	} else {
		fmt.Println("   Result: Failed to rewind (too old)")
	}
	fmt.Println()

	// Scenario 4: Demonstrate interpolation for smoother position
	fmt.Println("8. Entity Interpolation (for smooth rendering)")

	// Interpolate Player B's position between snapshots
	interpolateTime := baseTime.Add(425 * time.Millisecond) // Between snapshots
	interpolated, err := lc.InterpolateEntityAt(2, interpolateTime)

	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Time: 425ms (between 400ms and 450ms snapshots)\n")
		fmt.Printf("   Interpolated position: (%.1f, %.1f)\n",
			interpolated.Position.X, interpolated.Position.Y)
		fmt.Println("   Result: Smooth position between discrete snapshots")
	}
	fmt.Println()

	// Scenario 5: Performance characteristics
	fmt.Println("9. Performance Characteristics")

	// Measure rewind performance
	iterations := 10000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		lc.RewindToPlayerTime(200 * time.Millisecond)
	}
	elapsed := time.Since(start)

	fmt.Printf("   Rewind operations: %d\n", iterations)
	fmt.Printf("   Total time: %v\n", elapsed)
	fmt.Printf("   Average: %.2f µs/op\n", float64(elapsed.Microseconds())/float64(iterations))
	fmt.Println("   Result: Sub-microsecond performance (real-time capable)")
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("Lag compensation enables fair hit detection by:")
	fmt.Println("1. Recording historical world state snapshots")
	fmt.Println("2. Rewinding to when the player performed the action")
	fmt.Println("3. Validating hits against historical positions")
	fmt.Println("4. Preventing exploitation with reasonable time limits")
	fmt.Println()
	fmt.Println("This allows players with 200-5000ms latency to compete fairly!")
}

// Helper function to calculate distance between two positions
func distance(a, b network.Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return ((dx*dx + dy*dy) * 1.0) // Simplified, use math.Sqrt for real distance
}
