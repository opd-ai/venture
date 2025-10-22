package network

import (
	"testing"
	"time"
)

// TestDefaultServerConfig verifies default server configuration.
func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()
	
	if config.Address == "" {
		t.Error("Expected non-empty address")
	}
	
	if config.MaxPlayers == 0 {
		t.Error("Expected non-zero max players")
	}
	
	if config.ReadTimeout == 0 {
		t.Error("Expected non-zero read timeout")
	}
	
	if config.WriteTimeout == 0 {
		t.Error("Expected non-zero write timeout")
	}
	
	if config.UpdateRate == 0 {
		t.Error("Expected non-zero update rate")
	}
	
	if config.BufferSize == 0 {
		t.Error("Expected non-zero buffer size")
	}
}

// TestNewServer verifies server creation.
func TestNewServer(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	
	if server.IsRunning() {
		t.Error("Expected new server to not be running")
	}
	
	if server.GetPlayerCount() != 0 {
		t.Error("Expected initial player count to be 0")
	}
}

// TestServer_IsRunning verifies running state.
func TestServer_IsRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	if server.IsRunning() {
		t.Error("Expected new server to not be running")
	}
}

// TestServer_GetPlayerCount verifies player count.
func TestServer_GetPlayerCount(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	count := server.GetPlayerCount()
	if count != 0 {
		t.Errorf("Expected player count 0, got %d", count)
	}
}

// TestServer_GetPlayers verifies player list.
func TestServer_GetPlayers(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	players := server.GetPlayers()
	if len(players) != 0 {
		t.Errorf("Expected empty player list, got %d players", len(players))
	}
}

// TestServer_ReceiveInputCommand verifies input command channel.
func TestServer_ReceiveInputCommand(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	ch := server.ReceiveInputCommand()
	if ch == nil {
		t.Error("Expected non-nil input command channel")
	}
}

// TestServer_ReceiveError verifies error channel.
func TestServer_ReceiveError(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	ch := server.ReceiveError()
	if ch == nil {
		t.Error("Expected non-nil error channel")
	}
}

// TestServer_Stop_NotRunning verifies stop when not running.
func TestServer_Stop_NotRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	err := server.Stop()
	if err != nil {
		t.Errorf("Expected no error when stopping non-running server, got: %v", err)
	}
}

// TestServerConfig_CustomValues verifies custom configuration.
func TestServerConfig_CustomValues(t *testing.T) {
	config := ServerConfig{
		Address:      ":9000",
		MaxPlayers:   64,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 3 * time.Second,
		UpdateRate:   30,
		BufferSize:   512,
	}
	
	if config.Address != ":9000" {
		t.Errorf("Expected address ':9000', got %s", config.Address)
	}
	
	if config.MaxPlayers != 64 {
		t.Errorf("Expected max players 64, got %d", config.MaxPlayers)
	}
	
	if config.ReadTimeout != 15*time.Second {
		t.Error("Expected 15 second read timeout")
	}
	
	if config.WriteTimeout != 3*time.Second {
		t.Error("Expected 3 second write timeout")
	}
	
	if config.UpdateRate != 30 {
		t.Errorf("Expected update rate 30, got %d", config.UpdateRate)
	}
	
	if config.BufferSize != 512 {
		t.Errorf("Expected buffer size 512, got %d", config.BufferSize)
	}
}

// TestServer_MultipleInstances verifies multiple server instances.
func TestServer_MultipleInstances(t *testing.T) {
	config1 := DefaultServerConfig()
	config1.Address = ":0" // Random port
	server1 := NewServer(config1)
	
	config2 := DefaultServerConfig()
	config2.Address = ":0" // Random port
	server2 := NewServer(config2)
	
	if server1 == server2 {
		t.Error("Expected different server instances")
	}
}

// TestServer_SendStateUpdate_NoPlayers verifies behavior with no players.
func TestServer_SendStateUpdate_NoPlayers(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	update := &StateUpdate{
		Timestamp: 12345,
		EntityID:  1,
		Priority:  128,
	}
	
	err := server.SendStateUpdate(1, update)
	if err == nil {
		t.Error("Expected error when sending to non-existent player")
	}
}

// TestServer_BroadcastStateUpdate_NoPlayers verifies broadcast with no players.
func TestServer_BroadcastStateUpdate_NoPlayers(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	update := &StateUpdate{
		Timestamp: 12345,
		EntityID:  1,
		Priority:  128,
	}
	
	// Should not error, just no-op (no players to send to)
	server.BroadcastStateUpdate(update)
	
	// Verify sequence was assigned (first sequence is 0, so it was assigned)
	// The server assigns sequence numbers starting from 0
	expectedSeq := uint32(0)
	if update.SequenceNumber != expectedSeq {
		t.Errorf("Expected sequence number %d after first broadcast, got %d", expectedSeq, update.SequenceNumber)
	}
}

// TestServerConfig_ZeroValue verifies zero-value config behavior.
func TestServerConfig_ZeroValue(t *testing.T) {
	var config ServerConfig
	
	if config.Address != "" {
		t.Error("Expected empty address for zero value")
	}
	
	if config.MaxPlayers != 0 {
		t.Error("Expected zero max players for zero value")
	}
	
	if config.ReadTimeout != 0 {
		t.Error("Expected zero read timeout for zero value")
	}
	
	if config.WriteTimeout != 0 {
		t.Error("Expected zero write timeout for zero value")
	}
	
	if config.UpdateRate != 0 {
		t.Error("Expected zero update rate for zero value")
	}
	
	if config.BufferSize != 0 {
		t.Error("Expected zero buffer size for zero value")
	}
}

// TestServer_Protocol verifies protocol initialization.
func TestServer_Protocol(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	if server.protocol == nil {
		t.Error("Expected non-nil protocol")
	}
	
	// Verify it's a BinaryProtocol
	if _, ok := server.protocol.(*BinaryProtocol); !ok {
		t.Error("Expected protocol to be BinaryProtocol")
	}
}

// TestServer_Channels verifies channel creation and capacity.
func TestServer_Channels(t *testing.T) {
	config := DefaultServerConfig()
	config.BufferSize = 100
	config.MaxPlayers = 10
	server := NewServer(config)
	
	// Verify channels exist
	if server.inputCommands == nil {
		t.Error("Expected non-nil input commands channel")
	}
	
	if server.errors == nil {
		t.Error("Expected non-nil errors channel")
	}
	
	// Verify buffer capacity for input commands (buffer * max players)
	expectedCapacity := config.BufferSize * config.MaxPlayers
	if cap(server.inputCommands) != expectedCapacity {
		t.Errorf("Expected input commands buffer size %d, got %d", expectedCapacity, cap(server.inputCommands))
	}
}

// TestServer_SequenceTracking verifies state sequence tracking.
func TestServer_SequenceTracking(t *testing.T) {
	config := DefaultServerConfig()
	server := NewServer(config)
	
	// Broadcast multiple updates and verify sequence increments
	for i := 0; i < 5; i++ {
		update := &StateUpdate{
			Timestamp: uint64(i),
			EntityID:  1,
			Priority:  128,
		}
		server.BroadcastStateUpdate(update)
		
		expectedSeq := uint32(i)
		if update.SequenceNumber != expectedSeq {
			t.Errorf("Update %d: expected sequence %d, got %d", i, expectedSeq, update.SequenceNumber)
		}
	}
}

// TestClientConnection_IsConnected verifies client connection state.
func TestClientConnection_IsConnected(t *testing.T) {
	client := &clientConnection{
		playerID:  1,
		connected: true,
	}
	
	if !client.isConnected() {
		t.Error("Expected client to be connected")
	}
	
	client.connected = false
	if client.isConnected() {
		t.Error("Expected client to not be connected")
	}
}

// TestClientConnection_SendStateUpdate verifies sending state to client.
func TestClientConnection_SendStateUpdate(t *testing.T) {
	client := &clientConnection{
		playerID:     1,
		connected:    true,
		stateUpdates: make(chan *StateUpdate, 10),
	}
	
	update := &StateUpdate{
		Timestamp: 12345,
		EntityID:  1,
		Priority:  128,
	}
	
	client.sendStateUpdate(update)
	
	// Verify update was queued
	select {
	case received := <-client.stateUpdates:
		if received.Timestamp != update.Timestamp {
			t.Error("Received update does not match sent update")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for state update")
	}
}

// TestClientConnection_SendStateUpdate_Disconnected verifies no-op when disconnected.
func TestClientConnection_SendStateUpdate_Disconnected(t *testing.T) {
	client := &clientConnection{
		playerID:     1,
		connected:    false,
		stateUpdates: make(chan *StateUpdate, 10),
	}
	
	update := &StateUpdate{
		Timestamp: 12345,
		EntityID:  1,
		Priority:  128,
	}
	
	// Should not panic or block
	client.sendStateUpdate(update)
	
	// Verify no update was queued
	select {
	case <-client.stateUpdates:
		t.Error("Expected no state update when disconnected")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}
