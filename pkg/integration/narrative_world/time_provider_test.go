package narrative_world

import "testing"

func TestTimeProvider(t *testing.T) {
	// Test RealTimeProvider
	real := RealTimeProvider{}
	ts1 := real.Now()
	ts2 := real.Now()
	if ts1 > ts2 {
		t.Error("real time should not go backwards")
	}

	// Test FixedTimeProvider
	fixed := FixedTimeProvider{Timestamp: 1234567890}
	if fixed.Now() != 1234567890 {
		t.Errorf("expected 1234567890, got %d", fixed.Now())
	}
	if fixed.Now() != 1234567890 {
		t.Error("fixed time should return same value")
	}

	// Test IncrementingTimeProvider
	inc := &IncrementingTimeProvider{Current: 1000, Step: 10}
	if inc.Now() != 1000 {
		t.Errorf("expected 1000, got %d", inc.Now())
	}
	if inc.Now() != 1010 {
		t.Errorf("expected 1010, got %d", inc.Now())
	}
	if inc.Now() != 1020 {
		t.Errorf("expected 1020, got %d", inc.Now())
	}
}

func TestSetTimeProvider(t *testing.T) {
	// Save original state
	original := defaultTimeProvider
	defer func() {
		defaultTimeProvider = original
	}()

	// Set fixed provider
	SetTimeProvider(FixedTimeProvider{Timestamp: 9999})
	if now() != 9999 {
		t.Errorf("expected 9999, got %d", now())
	}

	// Reset to real
	ResetTimeProvider()
	ts := now()
	if ts < 1000000 {
		t.Error("expected real timestamp after reset")
	}
}

func TestDeterministicMemoryRecording(t *testing.T) {
	// Use fixed time provider for determinism
	SetTimeProvider(FixedTimeProvider{Timestamp: 1000000})
	defer ResetTimeProvider()

	manager := NewStoryEventManager(12345)
	companionID := uint64(100)

	manager.RecordMemory(companionID, EventTypeCombat, "Test event")

	memory := manager.memories[companionID]
	if len(memory.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(memory.Events))
	}

	if memory.Events[0].Timestamp != 1000000 {
		t.Errorf("expected timestamp 1000000, got %d", memory.Events[0].Timestamp)
	}
}

func TestDeterministicPruning(t *testing.T) {
	// Use incrementing provider to simulate time progression
	inc := &IncrementingTimeProvider{Current: 1000000, Step: 3600} // 1 hour increments
	SetTimeProvider(inc)
	defer ResetTimeProvider()

	manager := NewStoryEventManager(12345)
	manager.SetMaxMemoryEvents(5)
	companionID := uint64(100)

	// Record 10 events with different timestamps
	for i := 0; i < 10; i++ {
		eventType := EventTypeCombat
		if i == 0 || i == 9 {
			eventType = EventTypeSacrifice // High importance (oldest and newest)
		}
		manager.RecordMemory(companionID, eventType, "Test event")
	}

	// Should be pruned to 5 events
	count := manager.GetMemoryCount(companionID)
	if count != 5 {
		t.Errorf("expected 5 events after pruning, got %d", count)
	}

	// Verify high importance events are retained
	memory := manager.memories[companionID]
	hasHighImportance := false
	for _, event := range memory.Events {
		if event.Type == EventTypeSacrifice {
			hasHighImportance = true
			break
		}
	}
	if !hasHighImportance {
		t.Error("expected high importance events to be retained")
	}
}
