//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/integration/world_events"
)

// TestWorldEventManagerWrapper verifies the wrapper correctly adapts the EventManager.
func TestWorldEventManagerWrapper(t *testing.T) {
	seed := int64(99999)
	manager := world_events.NewEventManager(seed)

	wrapper := &worldEventManagerWrapper{system: manager}

	// Create empty entity slice (EventManager doesn't use entities)
	entities := make([]*engine.Entity, 0)

	// Should not panic
	wrapper.Update(entities, 0.016)
	wrapper.Update(entities, 0.032)
	wrapper.Update(entities, 0.016)

	t.Log("World event manager wrapper executed successfully")
}

// TestWorldEventManagerDeterminism verifies that the event manager
// produces deterministic results with the same seed.
func TestWorldEventManagerDeterminism(t *testing.T) {
	seed := int64(77777)

	// Create two managers with the same seed
	manager1 := world_events.NewEventManager(seed)
	manager2 := world_events.NewEventManager(seed)

	// Generate events with the same trigger
	params1 := world_events.TriggerParams{
		TriggerType: world_events.TriggerWeatherChange,
		Severity:    world_events.SeverityModerate,
		Location:    "test-location",
		ServerID:    "test-server",
		PlayerID:    "player1",
	}

	params2 := world_events.TriggerParams{
		TriggerType: world_events.TriggerWeatherChange,
		Severity:    world_events.SeverityModerate,
		Location:    "test-location",
		ServerID:    "test-server",
		PlayerID:    "player1",
	}

	event1, err1 := manager1.GenerateEvent(world_events.TriggerWeatherChange, params1)
	event2, err2 := manager2.GenerateEvent(world_events.TriggerWeatherChange, params2)

	if err1 != nil {
		t.Fatalf("Manager 1 failed to generate event: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Manager 2 failed to generate event: %v", err2)
	}

	// Verify events have the same type
	if event1.Type != event2.Type {
		t.Errorf("Event types differ: %v vs %v", event1.Type, event2.Type)
	}

	// Both should have generated weather disaster events
	if event1.Type != world_events.EventWeatherDisaster {
		t.Errorf("Expected EventWeatherDisaster, got %v", event1.Type)
	}

	t.Log("World event manager determinism verified")
}

// TestWorldEventManagerConcurrentUpdates verifies the event manager
// is safe for concurrent access through the wrapper.
func TestWorldEventManagerConcurrentUpdates(t *testing.T) {
	seed := int64(11111)
	manager := world_events.NewEventManager(seed)
	wrapper := &worldEventManagerWrapper{system: manager}

	entities := make([]*engine.Entity, 0)

	// Run multiple updates concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				wrapper.Update(entities, 0.016)
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
