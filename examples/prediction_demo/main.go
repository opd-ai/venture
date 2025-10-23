//go:build test
// +build test

package main

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/network"
)

func main() {
	fmt.Println("=== Client-Side Prediction Demo ===")
	fmt.Println()

	demonstratePrediction()
	fmt.Println()
	demonstrateReconciliation()
	fmt.Println()
	demonstrateInterpolation()
}

func demonstratePrediction() {
	fmt.Println("1. Client-Side Prediction")
	fmt.Println("   Predicts local player movement before server confirmation")
	fmt.Println()

	predictor := network.NewClientPredictor()
	predictor.SetInitialState(network.Position{X: 0, Y: 0}, network.Velocity{VX: 0, VY: 0})

	fmt.Println("   Initial State: Position (0, 0), Velocity (0, 0)")
	fmt.Println()

	// Simulate player inputs over several frames
	fmt.Println("   Simulating 5 frames of rightward movement:")
	deltaTime := 0.016 // 60 FPS

	for i := 1; i <= 5; i++ {
		predicted := predictor.PredictInput(100, 0, deltaTime) // Move right
		fmt.Printf("   Frame %d - Seq %d: Position (%.2f, %.2f), Velocity (%.2f, %.2f)\n",
			i, predicted.Sequence, predicted.Position.X, predicted.Position.Y,
			predicted.Velocity.VX, predicted.Velocity.VY)
	}

	finalState := predictor.GetCurrentState()
	fmt.Printf("\n   Final Predicted Position: (%.2f, %.2f)\n", finalState.Position.X, finalState.Position.Y)
	fmt.Println("   ✓ Client sees immediate response to inputs")
}

func demonstrateReconciliation() {
	fmt.Println("2. Server Reconciliation")
	fmt.Println("   Corrects prediction errors when server updates arrive")
	fmt.Println()

	predictor := network.NewClientPredictor()
	predictor.SetInitialState(network.Position{X: 0, Y: 0}, network.Velocity{VX: 0, VY: 0})

	deltaTime := 0.016

	// Make several predictions
	fmt.Println("   Client predicts 5 frames:")
	for i := 1; i <= 5; i++ {
		predicted := predictor.PredictInput(100, 0, deltaTime)
		fmt.Printf("   Predicted Seq %d: Position (%.2f, %.2f)\n",
			predicted.Sequence, predicted.Position.X, predicted.Position.Y)
	}

	// Server acknowledges sequence 3 with a different position (prediction error)
	fmt.Println("\n   Server acknowledges Seq 3 with corrected position:")
	serverPos := network.Position{X: 1.0, Y: 0.0}
	serverVel := network.Velocity{VX: 10, VY: 0}
	fmt.Printf("   Server Position: (%.2f, %.2f) - differs from client!\n", serverPos.X, serverPos.Y)

	// Reconcile
	corrected := predictor.ReconcileServerState(3, serverPos, serverVel)
	fmt.Printf("\n   After reconciliation (replaying inputs 4, 5):\n")
	fmt.Printf("   Corrected Position: (%.2f, %.2f)\n", corrected.Position.X, corrected.Position.Y)
	fmt.Printf("   ✓ Client adjusts smoothly to server authority\n")
}

func demonstrateInterpolation() {
	fmt.Println("3. Entity Interpolation")
	fmt.Println("   Smoothly interpolates remote entities between snapshots")
	fmt.Println()

	sm := network.NewSnapshotManager(100)

	baseTime := time.Now()

	// Create two snapshots 100ms apart
	fmt.Println("   Snapshot 1 (t=0ms): Entity at (0, 0)")
	sm.AddSnapshot(network.WorldSnapshot{
		Timestamp: baseTime,
		Entities: map[uint64]network.EntitySnapshot{
			100: {
				EntityID: 100,
				Position: network.Position{X: 0, Y: 0},
				Velocity: network.Velocity{VX: 100, VY: 0},
			},
		},
	})

	time.Sleep(10 * time.Millisecond) // Small delay for timestamp difference

	fmt.Println("   Snapshot 2 (t=100ms): Entity at (100, 0)")
	sm.AddSnapshot(network.WorldSnapshot{
		Timestamp: baseTime.Add(100 * time.Millisecond),
		Entities: map[uint64]network.EntitySnapshot{
			100: {
				EntityID: 100,
				Position: network.Position{X: 100, Y: 0},
				Velocity: network.Velocity{VX: 100, VY: 0},
			},
		},
	})

	// Interpolate at various times
	fmt.Println("\n   Interpolated positions:")
	times := []struct {
		offset   time.Duration
		expected float64
	}{
		{0, 0.0},
		{25 * time.Millisecond, 25.0},
		{50 * time.Millisecond, 50.0},
		{75 * time.Millisecond, 75.0},
		{100 * time.Millisecond, 100.0},
	}

	for _, tc := range times {
		renderTime := baseTime.Add(tc.offset)
		interpolated := sm.InterpolateEntity(100, renderTime)
		if interpolated != nil {
			fmt.Printf("   t=%3dms: Position (%.1f, %.1f) - expected ~%.1f\n",
				tc.offset.Milliseconds(), interpolated.Position.X, interpolated.Position.Y, tc.expected)
		}
	}

	fmt.Println("\n   ✓ Smooth movement between server updates")
}
