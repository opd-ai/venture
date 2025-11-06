package network

import (
	"testing"
)

// TestClientConnection_PriorityOrdering verifies high-priority messages are sent before low-priority.
func TestClientConnection_PriorityOrdering(t *testing.T) {
	client := &clientConnection{
		playerID:         1,
		connected:        true,
		stateUpdateQueue: NewStateUpdatePriorityQueue(10),
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats("test_priority", 10, nil),
	}

	// Send updates in mixed priority order
	updates := []*StateUpdate{
		{EntityID: 1, Priority: PriorityNormal},
		{EntityID: 2, Priority: PriorityCritical},
		{EntityID: 3, Priority: PriorityLow},
		{EntityID: 4, Priority: PriorityHigh},
		{EntityID: 5, Priority: PriorityNormal},
	}

	for _, update := range updates {
		client.sendStateUpdate(update)
	}

	// Verify queue has all updates
	if client.stateUpdateQueue.Len() != 5 {
		t.Fatalf("Expected 5 updates in queue, got %d", client.stateUpdateQueue.Len())
	}

	// Pop and verify priority ordering (high to low)
	expectedOrder := []uint64{2, 4, 1, 5, 3} // EntityIDs in priority order
	for i, expectedID := range expectedOrder {
		popped := client.stateUpdateQueue.Pop()
		if popped == nil {
			t.Fatalf("Expected update at position %d, got nil", i)
		}
		if popped.EntityID != expectedID {
			t.Errorf("At position %d: expected EntityID %d (priority order), got %d", i, expectedID, popped.EntityID)
		}
	}
}

// TestClientConnection_PriorityBufferFull verifies behavior when priority queue is full.
func TestClientConnection_PriorityBufferFull(t *testing.T) {
	capacity := 3
	client := &clientConnection{
		playerID:         1,
		connected:        true,
		stateUpdateQueue: NewStateUpdatePriorityQueue(capacity),
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats("test_full", capacity, nil),
	}

	// Fill the queue
	for i := 0; i < capacity; i++ {
		update := &StateUpdate{
			EntityID: uint64(i),
			Priority: PriorityNormal,
		}
		client.sendStateUpdate(update)
	}

	// Try to send one more (should be dropped)
	overflow := &StateUpdate{
		EntityID: 999,
		Priority: PriorityCritical,
	}
	client.sendStateUpdate(overflow)

	// Verify queue is still at capacity
	if client.stateUpdateQueue.Len() != capacity {
		t.Errorf("Expected queue length %d, got %d", capacity, client.stateUpdateQueue.Len())
	}

	// Verify the overflow was counted as a drop
	stats := client.stateUpdateStats.Snapshot()
	if stats.Dropped == 0 {
		t.Error("Expected at least one drop to be recorded")
	}
}

// TestClientConnection_CriticalPriorityFirst verifies critical messages beat normal ones.
func TestClientConnection_CriticalPriorityFirst(t *testing.T) {
	client := &clientConnection{
		playerID:         1,
		connected:        true,
		stateUpdateQueue: NewStateUpdatePriorityQueue(10),
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats("test_critical", 10, nil),
	}

	// Send normal priority first
	normal := &StateUpdate{
		EntityID: 1,
		Priority: PriorityNormal,
	}
	client.sendStateUpdate(normal)

	// Send critical priority second
	critical := &StateUpdate{
		EntityID: 2,
		Priority: PriorityCritical,
	}
	client.sendStateUpdate(critical)

	// Critical should come out first despite being sent second
	first := client.stateUpdateQueue.Pop()
	if first == nil || first.EntityID != 2 {
		t.Errorf("Expected critical update (EntityID=2) first, got EntityID=%v", first)
	}

	// Normal should come out second
	second := client.stateUpdateQueue.Pop()
	if second == nil || second.EntityID != 1 {
		t.Errorf("Expected normal update (EntityID=1) second, got EntityID=%v", second)
	}
}

// TestBroadcastStateUpdate_PriorityPreserved verifies broadcast preserves priority.
func TestBroadcastStateUpdate_PriorityPreserved(t *testing.T) {
	config := DefaultServerConfig()
	config.BufferSize = 10
	server := NewServer(config)

	// Create test client connections (simulating connected clients)
	client1 := &clientConnection{
		playerID:         1,
		connected:        true,
		stateUpdateQueue: NewStateUpdatePriorityQueue(10),
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats("client_1", 10, nil),
	}
	client2 := &clientConnection{
		playerID:         2,
		connected:        true,
		stateUpdateQueue: NewStateUpdatePriorityQueue(10),
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats("client_2", 10, nil),
	}

	server.clients[1] = client1
	server.clients[2] = client2

	// Broadcast a critical update
	update := &StateUpdate{
		EntityID: 100,
		Priority: PriorityCritical,
	}
	server.BroadcastStateUpdate(update)

	// Verify both clients received the update with correct priority
	pop1 := client1.stateUpdateQueue.Pop()
	if pop1 == nil || pop1.Priority != PriorityCritical {
		t.Error("Client 1 did not receive critical priority update")
	}

	pop2 := client2.stateUpdateQueue.Pop()
	if pop2 == nil || pop2.Priority != PriorityCritical {
		t.Error("Client 2 did not receive critical priority update")
	}
}
