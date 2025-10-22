package network

import (
	"testing"
	"time"
)

// TestDefaultClientConfig verifies default client configuration.
func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()
	
	if config.ServerAddress == "" {
		t.Error("Expected non-empty server address")
	}
	
	if config.ConnectionTimeout == 0 {
		t.Error("Expected non-zero connection timeout")
	}
	
	if config.PingInterval == 0 {
		t.Error("Expected non-zero ping interval")
	}
	
	if config.MaxLatency == 0 {
		t.Error("Expected non-zero max latency")
	}
	
	if config.BufferSize == 0 {
		t.Error("Expected non-zero buffer size")
	}
}

// TestNewClient verifies client creation.
func TestNewClient(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	
	if client.IsConnected() {
		t.Error("Expected new client to not be connected")
	}
	
	if client.GetPlayerID() != 0 {
		t.Error("Expected initial player ID to be 0")
	}
}

// TestClient_SetPlayerID verifies player ID management.
func TestClient_SetPlayerID(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	testID := uint64(12345)
	client.SetPlayerID(testID)
	
	if client.GetPlayerID() != testID {
		t.Errorf("Expected player ID %d, got %d", testID, client.GetPlayerID())
	}
}

// TestClient_GetLatency verifies latency tracking.
func TestClient_GetLatency(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	latency := client.GetLatency()
	if latency < 0 {
		t.Error("Expected non-negative latency")
	}
}

// TestClient_IsConnected verifies connection state.
func TestClient_IsConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	if client.IsConnected() {
		t.Error("Expected new client to not be connected")
	}
}

// TestClient_SendInput_NotConnected verifies error when not connected.
func TestClient_SendInput_NotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	err := client.SendInput("move", []byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error when sending input while not connected")
	}
}

// TestClient_ReceiveStateUpdate verifies state update channel.
func TestClient_ReceiveStateUpdate(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	ch := client.ReceiveStateUpdate()
	if ch == nil {
		t.Error("Expected non-nil state update channel")
	}
}

// TestClient_ReceiveError verifies error channel.
func TestClient_ReceiveError(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	ch := client.ReceiveError()
	if ch == nil {
		t.Error("Expected non-nil error channel")
	}
}

// TestClientConfig_CustomValues verifies custom configuration.
func TestClientConfig_CustomValues(t *testing.T) {
	config := ClientConfig{
		ServerAddress:   "192.168.1.1:9000",
		ConnectionTimeout: 5 * time.Second,
		PingInterval:    500 * time.Millisecond,
		MaxLatency:      1 * time.Second,
		BufferSize:      512,
	}
	
	if config.ServerAddress != "192.168.1.1:9000" {
		t.Errorf("Expected server address '192.168.1.1:9000', got %s", config.ServerAddress)
	}
	
	if config.ConnectionTimeout != 5*time.Second {
		t.Error("Expected 5 second connection timeout")
	}
	
	if config.PingInterval != 500*time.Millisecond {
		t.Error("Expected 500ms ping interval")
	}
	
	if config.MaxLatency != 1*time.Second {
		t.Error("Expected 1 second max latency")
	}
	
	if config.BufferSize != 512 {
		t.Errorf("Expected buffer size 512, got %d", config.BufferSize)
	}
}

// TestClient_MultipleInstances verifies multiple client instances.
func TestClient_MultipleInstances(t *testing.T) {
	config1 := DefaultClientConfig()
	client1 := NewClient(config1)
	
	config2 := DefaultClientConfig()
	config2.ServerAddress = "different:8081"
	client2 := NewClient(config2)
	
	if client1 == client2 {
		t.Error("Expected different client instances")
	}
	
	client1.SetPlayerID(1)
	client2.SetPlayerID(2)
	
	if client1.GetPlayerID() == client2.GetPlayerID() {
		t.Error("Expected different player IDs")
	}
}

// TestClient_Disconnect_NotConnected verifies disconnect when not connected.
func TestClient_Disconnect_NotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Expected no error when disconnecting non-connected client, got: %v", err)
	}
}

// TestClientConfig_ZeroValue verifies zero-value config behavior.
func TestClientConfig_ZeroValue(t *testing.T) {
	var config ClientConfig
	
	if config.ServerAddress != "" {
		t.Error("Expected empty server address for zero value")
	}
	
	if config.ConnectionTimeout != 0 {
		t.Error("Expected zero connection timeout for zero value")
	}
	
	if config.PingInterval != 0 {
		t.Error("Expected zero ping interval for zero value")
	}
	
	if config.MaxLatency != 0 {
		t.Error("Expected zero max latency for zero value")
	}
	
	if config.BufferSize != 0 {
		t.Error("Expected zero buffer size for zero value")
	}
}

// TestClient_SequenceTracking verifies input sequence tracking.
func TestClient_SequenceTracking(t *testing.T) {
	config := DefaultClientConfig()
	config.BufferSize = 10
	client := NewClient(config)
	client.SetPlayerID(1)
	
	// Can't actually send without connection, but verify sequence increments
	// are handled in the SendInput method
	initialSeq := client.inputSeq
	
	// Even though these will fail (not connected), sequence should not change
	// until we're connected
	_ = client.SendInput("move", []byte{1})
	
	if client.inputSeq != initialSeq {
		t.Error("Expected sequence to remain unchanged when not connected")
	}
}

// TestClient_Channels verifies channel creation and capacity.
func TestClient_Channels(t *testing.T) {
	config := DefaultClientConfig()
	config.BufferSize = 100
	client := NewClient(config)
	
	// Verify channels exist
	if client.stateUpdates == nil {
		t.Error("Expected non-nil state updates channel")
	}
	
	if client.inputQueue == nil {
		t.Error("Expected non-nil input queue channel")
	}
	
	if client.errors == nil {
		t.Error("Expected non-nil errors channel")
	}
	
	// Verify buffer capacity
	if cap(client.stateUpdates) != config.BufferSize {
		t.Errorf("Expected state updates buffer size %d, got %d", config.BufferSize, cap(client.stateUpdates))
	}
	
	if cap(client.inputQueue) != config.BufferSize {
		t.Errorf("Expected input queue buffer size %d, got %d", config.BufferSize, cap(client.inputQueue))
	}
}

// TestClient_GetLatency_InitialValue verifies initial latency.
func TestClient_GetLatency_InitialValue(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	latency := client.GetLatency()
	
	// Initial latency should be 0 (no connection yet)
	if latency != 0 {
		t.Errorf("Expected initial latency 0, got %v", latency)
	}
}

// TestClient_Protocol verifies protocol initialization.
func TestClient_Protocol(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)
	
	if client.protocol == nil {
		t.Error("Expected non-nil protocol")
	}
	
	// Verify it's a BinaryProtocol
	if _, ok := client.protocol.(*BinaryProtocol); !ok {
		t.Error("Expected protocol to be BinaryProtocol")
	}
}
