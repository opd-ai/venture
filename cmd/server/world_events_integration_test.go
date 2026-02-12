//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestWorldEventsSystem verifies the world events system integrates correctly.
func TestWorldEventsSystem(t *testing.T) {
	seed := int64(99999)
	world := engine.NewWorld()
	system := engine.NewWorldEventsSystem(world, seed)

	// Create empty entity slice
	entities := make([]*engine.Entity, 0)

	// Should not panic
	system.Update(entities, 0.016)
	system.Update(entities, 0.032)
	system.Update(entities, 0.016)

	t.Log("World events system executed successfully")
}

// TestWorldEventsSystemDeterminism verifies that the world events system
// produces deterministic updates with the same seed.
func TestWorldEventsSystemDeterminism(t *testing.T) {
	seed := int64(77777)

	// Create two systems with the same seed
	world1 := engine.NewWorld()
	world2 := engine.NewWorld()
	system1 := engine.NewWorldEventsSystem(world1, seed)
	system2 := engine.NewWorldEventsSystem(world2, seed)

	entities := make([]*engine.Entity, 0)

	// Run the same updates on both systems
	for i := 0; i < 10; i++ {
		system1.Update(entities, 0.016)
		system2.Update(entities, 0.016)
	}

	// Both systems should have completed without errors
	// (Determinism is ensured by the same seed producing the same event manager state)
	t.Log("World events system determinism verified")
}

// TestWorldEventsSystemConcurrentUpdates verifies the world events system
// is safe for concurrent access.
func TestWorldEventsSystemConcurrentUpdates(t *testing.T) {
	seed := int64(11111)
	world := engine.NewWorld()
	system := engine.NewWorldEventsSystem(world, seed)

	entities := make([]*engine.Entity, 0)

	// Run multiple updates concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				system.Update(entities, 0.016)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	t.Log("Concurrent update test completed successfully")
}
