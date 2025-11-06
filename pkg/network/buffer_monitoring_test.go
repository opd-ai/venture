package network

import (
	"testing"
	"time"
)

func TestClient_BufferMonitoring(t *testing.T) {
	config := DefaultClientConfig()
	config.BufferSize = 10 // Small buffer for testing
	client := NewClient(config)

	// Verify stats are initialized
	stats := client.GetBufferStats()
	if len(stats) != 3 {
		t.Errorf("Expected 3 buffer stats, got %d", len(stats))
	}

	// Verify initial state
	for name, snapshot := range stats {
		if snapshot.Sent != 0 {
			t.Errorf("%s: Expected sent=0, got %d", name, snapshot.Sent)
		}
		if snapshot.Dropped != 0 {
			t.Errorf("%s: Expected dropped=0, got %d", name, snapshot.Dropped)
		}
		if snapshot.CurrentSize != 0 {
			t.Errorf("%s: Expected size=0, got %d", name, snapshot.CurrentSize)
		}
	}
}

func TestClient_InputQueueMonitoring(t *testing.T) {
	config := DefaultClientConfig()
	config.BufferSize = 5 // Very small buffer
	client := NewClient(config)

	// Try to fill the input queue
	for i := 0; i < 10; i++ {
		err := client.SendInput("test", []byte("data"))
		if i < 5 {
			// First 5 should succeed (not connected, so queued)
			if err == nil {
				t.Errorf("Send %d: Expected error (not connected), got nil", i)
			}
		}
	}

	// Check stats
	stats := client.GetBufferStats()
	inputStats := stats["input_queue"]

	// Note: Since we're not connected, SendInput returns error before queuing
	// So we won't actually see sent/dropped here. This test verifies the stats exist.
	if inputStats.Capacity != 5 {
		t.Errorf("Expected capacity=5, got %d", inputStats.Capacity)
	}
}

func TestClient_StateUpdateMonitoring(t *testing.T) {
	config := DefaultClientConfig()
	config.BufferSize = 5
	client := NewClient(config)

	// Simulate receiving state updates (fill the channel)
	for i := 0; i < 10; i++ {
		update := &StateUpdate{
			SequenceNumber: uint32(i),
			Timestamp:      uint64(time.Now().UnixNano()),
		}
		select {
		case client.stateUpdates <- update:
			client.stateUpdateStats.RecordSend()
		default:
			client.stateUpdateStats.RecordDrop()
		}
	}

	// Check stats
	stats := client.GetBufferStats()
	stateStats := stats["state_updates"]

	expectedSent := 5
	expectedDropped := 5
	if int(stateStats.Sent) != expectedSent {
		t.Errorf("Expected sent=%d, got %d", expectedSent, stateStats.Sent)
	}
	if int(stateStats.Dropped) != expectedDropped {
		t.Errorf("Expected dropped=%d, got %d", expectedDropped, stateStats.Dropped)
	}
	if stateStats.Utilization != 1.0 {
		t.Errorf("Expected utilization=1.0 (full), got %f", stateStats.Utilization)
	}
	if stateStats.DropRate != 0.5 {
		t.Errorf("Expected drop_rate=0.5, got %f", stateStats.DropRate)
	}
}

func TestServer_BufferMonitoring(t *testing.T) {
	config := DefaultServerConfig()
	config.BufferSize = 10
	config.MaxPlayers = 2
	server := NewServer(config)

	// Verify stats are initialized
	stats := server.GetBufferStats()
	if len(stats) != 4 {
		t.Errorf("Expected 4 buffer stats, got %d", len(stats))
	}

	// Verify initial state
	for name, snapshot := range stats {
		if snapshot.Sent != 0 {
			t.Errorf("%s: Expected sent=0, got %d", name, snapshot.Sent)
		}
		if snapshot.Dropped != 0 {
			t.Errorf("%s: Expected dropped=0, got %d", name, snapshot.Dropped)
		}
	}
}

func TestServer_InputCommandMonitoring(t *testing.T) {
	config := DefaultServerConfig()
	config.BufferSize = 3
	config.MaxPlayers = 1
	server := NewServer(config)

	// Fill input commands channel
	totalCapacity := config.BufferSize * config.MaxPlayers
	for i := 0; i < totalCapacity+5; i++ {
		cmd := &InputCommand{
			PlayerID:       1,
			SequenceNumber: uint32(i),
			InputType:      "test",
		}
		select {
		case server.inputCommands <- cmd:
			server.inputCommandStats.RecordSend()
		default:
			server.inputCommandStats.RecordDrop()
		}
	}

	// Check stats
	stats := server.GetBufferStats()
	inputStats := stats["input_commands"]

	expectedSent := totalCapacity
	expectedDropped := 5
	if int(inputStats.Sent) != expectedSent {
		t.Errorf("Expected sent=%d, got %d", expectedSent, inputStats.Sent)
	}
	if int(inputStats.Dropped) != expectedDropped {
		t.Errorf("Expected dropped=%d, got %d", expectedDropped, inputStats.Dropped)
	}
}

func TestServer_PlayerEventMonitoring(t *testing.T) {
	config := DefaultServerConfig()
	config.MaxPlayers = 3
	server := NewServer(config)

	// Send player join events
	for i := uint64(1); i <= 5; i++ {
		select {
		case server.playerJoins <- i:
			server.playerJoinStats.RecordSend()
		default:
			server.playerJoinStats.RecordDrop()
		}
	}

	// Send player leave events
	for i := uint64(1); i <= 5; i++ {
		select {
		case server.playerLeaves <- i:
			server.playerLeaveStats.RecordSend()
		default:
			server.playerLeaveStats.RecordDrop()
		}
	}

	// Check stats
	stats := server.GetBufferStats()

	joinStats := stats["player_joins"]
	expectedJoinSent := 3
	expectedJoinDropped := 2
	if int(joinStats.Sent) != expectedJoinSent {
		t.Errorf("Joins: Expected sent=%d, got %d", expectedJoinSent, joinStats.Sent)
	}
	if int(joinStats.Dropped) != expectedJoinDropped {
		t.Errorf("Joins: Expected dropped=%d, got %d", expectedJoinDropped, joinStats.Dropped)
	}

	leaveStats := stats["player_leaves"]
	expectedLeaveSent := 3
	expectedLeaveDropped := 2
	if int(leaveStats.Sent) != expectedLeaveSent {
		t.Errorf("Leaves: Expected sent=%d, got %d", expectedLeaveSent, leaveStats.Sent)
	}
	if int(leaveStats.Dropped) != expectedLeaveDropped {
		t.Errorf("Leaves: Expected dropped=%d, got %d", expectedLeaveDropped, leaveStats.Dropped)
	}
}

func TestBufferMonitoring_HighUtilizationWarning(t *testing.T) {
	// This test verifies that warnings are logged at 80% threshold.
	// Since we can't easily capture log output in tests, we verify the
	// mechanism works by checking that stats are properly tracked.

	config := DefaultClientConfig()
	config.BufferSize = 10
	client := NewClient(config)

	// Fill to 80%
	for i := 0; i < 8; i++ {
		update := &StateUpdate{SequenceNumber: uint32(i)}
		select {
		case client.stateUpdates <- update:
			client.stateUpdateStats.RecordSend()
		default:
			client.stateUpdateStats.RecordDrop()
		}
	}

	stats := client.GetBufferStats()
	stateStats := stats["state_updates"]

	if stateStats.Utilization != 0.8 {
		t.Errorf("Expected utilization=0.8, got %f", stateStats.Utilization)
	}

	// Utilization at exactly 80% should trigger warning (verified in buffer_stats_test.go)
}
