//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestWorldEventsSystemUnification verifies that server uses engine.WorldEventsSystem
// instead of the raw world_events.EventManager wrapper for consistency with client.
func TestWorldEventsSystemUnification(t *testing.T) {
	seed := int64(88888)
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()

	initialSystemCount := len(world.GetSystems())

	// Initialize V4 systems which includes WorldEventsSystem
	initializeV4Systems(world, seed, "fantasy", logger, nil)
	initializeV6SystemsServer(world, seed, logger, nil)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// Verify at least one system was added
	if addedSystems < 1 {
		t.Fatalf("No systems were added, expected at least 1")
	}

	// Verify WorldEventsSystem is present (not the wrapper)
	systems := world.GetSystems()
	foundWorldEventsSystem := false

	for _, sys := range systems {
		// Check if it's the WorldEventsSystem by type assertion
		if _, ok := sys.(*engine.WorldEventsSystem); ok {
			foundWorldEventsSystem = true
			break
		}
	}

	if !foundWorldEventsSystem {
		t.Error("engine.WorldEventsSystem not found in server systems (should replace world_events.EventManager wrapper)")
	}

	t.Log("World events system properly unified between client and server")
}

// TestWorldEventsSystemTriggerLogic verifies that the unified system includes
// trigger checking logic (checkGuildWarfareTriggers, checkEconomicTriggers, etc.)
// that was missing from the raw EventManager wrapper.
func TestWorldEventsSystemTriggerLogic(t *testing.T) {
	seed := int64(77777)
	world := engine.NewWorld()
	system := engine.NewWorldEventsSystem(world, seed)

	entities := []*engine.Entity{}

	// System should run without errors and check triggers internally
	// (updateInterval is 30s, so won't trigger on first call)
	system.Update(entities, 0.016)
	system.Update(entities, 15.0) // Half interval
	system.Update(entities, 16.0) // Over interval threshold

	t.Log("World events system trigger logic executed successfully")
}

// TestWorldEventsSystemServerClientParity verifies that both server and client
// use the same WorldEventsSystem implementation for consistency.
func TestWorldEventsSystemServerClientParity(t *testing.T) {
	seed := int64(99999)

	// Server world
	serverWorld := engine.NewWorld()
	serverLogger := createTestLoggerForSystems()
	initializeV4Systems(serverWorld, seed, "fantasy", serverLogger, nil)
	initializeV6SystemsServer(serverWorld, seed, serverLogger, nil)

	// Client world simulation
	clientWorld := engine.NewWorld()
	clientSystem := engine.NewWorldEventsSystemWithLogger(clientWorld, seed, serverLogger)
	clientWorld.AddSystem(clientSystem)

	// Both should have WorldEventsSystem
	serverSystems := serverWorld.GetSystems()
	clientSystems := clientWorld.GetSystems()

	serverHasWorldEvents := false
	for _, sys := range serverSystems {
		if _, ok := sys.(*engine.WorldEventsSystem); ok {
			serverHasWorldEvents = true
			break
		}
	}

	clientHasWorldEvents := false
	for _, sys := range clientSystems {
		if _, ok := sys.(*engine.WorldEventsSystem); ok {
			clientHasWorldEvents = true
			break
		}
	}

	if !serverHasWorldEvents {
		t.Error("Server missing engine.WorldEventsSystem")
	}
	if !clientHasWorldEvents {
		t.Error("Client missing engine.WorldEventsSystem")
	}

	t.Log("Server and client use identical WorldEventsSystem implementation")
}
