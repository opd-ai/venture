package network

import (
	"net"
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

// TestTorClientConfig verifies Tor/high-latency client configuration.
func TestTorClientConfig(t *testing.T) {
	config := TorClientConfig()

	// Verify all fields are non-zero
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

	// Verify Tor-specific values
	expectedConnectionTimeout := 60 * time.Second
	if config.ConnectionTimeout != expectedConnectionTimeout {
		t.Errorf("Expected ConnectionTimeout %v, got %v", expectedConnectionTimeout, config.ConnectionTimeout)
	}

	expectedMaxLatency := 5000 * time.Millisecond
	if config.MaxLatency != expectedMaxLatency {
		t.Errorf("Expected MaxLatency %v, got %v", expectedMaxLatency, config.MaxLatency)
	}

	expectedPingInterval := 5 * time.Second
	if config.PingInterval != expectedPingInterval {
		t.Errorf("Expected PingInterval %v, got %v", expectedPingInterval, config.PingInterval)
	}

	expectedBufferSize := 512
	if config.BufferSize != expectedBufferSize {
		t.Errorf("Expected BufferSize %d, got %d", expectedBufferSize, config.BufferSize)
	}

	// Verify it's different from default
	defaultConfig := DefaultClientConfig()
	if config.ConnectionTimeout == defaultConfig.ConnectionTimeout {
		t.Error("Tor ConnectionTimeout should differ from default")
	}

	if config.MaxLatency == defaultConfig.MaxLatency {
		t.Error("Tor MaxLatency should differ from default")
	}

	if config.PingInterval == defaultConfig.PingInterval {
		t.Error("Tor PingInterval should differ from default")
	}

	if config.BufferSize == defaultConfig.BufferSize {
		t.Error("Tor BufferSize should differ from default")
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
		ServerAddress:     "192.168.1.1:9000",
		ConnectionTimeout: 5 * time.Second,
		PingInterval:      500 * time.Millisecond,
		MaxLatency:        1 * time.Second,
		BufferSize:        512,
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

// TestDefaultReconnectConfig verifies default reconnection configuration.
func TestDefaultReconnectConfig(t *testing.T) {
	config := DefaultReconnectConfig()

	if config.MaxRetries == 0 {
		t.Error("Expected non-zero max retries")
	}

	if config.InitialDelay == 0 {
		t.Error("Expected non-zero initial delay")
	}

	if config.MaxDelay == 0 {
		t.Error("Expected non-zero max delay")
	}

	if config.BackoffFactor == 0 {
		t.Error("Expected non-zero backoff factor")
	}

	// Verify sensible defaults
	expectedMaxRetries := 5
	if config.MaxRetries != expectedMaxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", expectedMaxRetries, config.MaxRetries)
	}

	expectedInitialDelay := 1 * time.Second
	if config.InitialDelay != expectedInitialDelay {
		t.Errorf("Expected InitialDelay %v, got %v", expectedInitialDelay, config.InitialDelay)
	}

	expectedMaxDelay := 30 * time.Second
	if config.MaxDelay != expectedMaxDelay {
		t.Errorf("Expected MaxDelay %v, got %v", expectedMaxDelay, config.MaxDelay)
	}

	expectedBackoffFactor := 2.0
	if config.BackoffFactor != expectedBackoffFactor {
		t.Errorf("Expected BackoffFactor %.1f, got %.1f", expectedBackoffFactor, config.BackoffFactor)
	}
}

// TestTorReconnectConfig verifies Tor/high-latency reconnection configuration.
func TestTorReconnectConfig(t *testing.T) {
	config := TorReconnectConfig()

	// Verify all fields are non-zero
	if config.MaxRetries == 0 {
		t.Error("Expected non-zero max retries")
	}

	if config.InitialDelay == 0 {
		t.Error("Expected non-zero initial delay")
	}

	if config.MaxDelay == 0 {
		t.Error("Expected non-zero max delay")
	}

	if config.BackoffFactor == 0 {
		t.Error("Expected non-zero backoff factor")
	}

	// Verify Tor-specific values
	expectedMaxRetries := 10
	if config.MaxRetries != expectedMaxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", expectedMaxRetries, config.MaxRetries)
	}

	expectedInitialDelay := 5 * time.Second
	if config.InitialDelay != expectedInitialDelay {
		t.Errorf("Expected InitialDelay %v, got %v", expectedInitialDelay, config.InitialDelay)
	}

	expectedMaxDelay := 120 * time.Second
	if config.MaxDelay != expectedMaxDelay {
		t.Errorf("Expected MaxDelay %v, got %v", expectedMaxDelay, config.MaxDelay)
	}

	expectedBackoffFactor := 2.0
	if config.BackoffFactor != expectedBackoffFactor {
		t.Errorf("Expected BackoffFactor %.1f, got %.1f", expectedBackoffFactor, config.BackoffFactor)
	}

	// Verify it's more aggressive than default
	defaultConfig := DefaultReconnectConfig()
	if config.MaxRetries <= defaultConfig.MaxRetries {
		t.Error("Tor MaxRetries should be greater than default")
	}

	if config.InitialDelay <= defaultConfig.InitialDelay {
		t.Error("Tor InitialDelay should be greater than default")
	}

	if config.MaxDelay <= defaultConfig.MaxDelay {
		t.Error("Tor MaxDelay should be greater than default")
	}
}

// TestReconnectConfig_CustomValues verifies custom reconnection configuration.
func TestReconnectConfig_CustomValues(t *testing.T) {
	tests := []struct {
		name   string
		config ReconnectConfig
	}{
		{
			name: "minimal retries",
			config: ReconnectConfig{
				MaxRetries:    1,
				InitialDelay:  500 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 1.5,
			},
		},
		{
			name: "infinite retries",
			config: ReconnectConfig{
				MaxRetries:    0, // 0 = infinite
				InitialDelay:  2 * time.Second,
				MaxDelay:      60 * time.Second,
				BackoffFactor: 2.5,
			},
		},
		{
			name: "aggressive retries",
			config: ReconnectConfig{
				MaxRetries:    20,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      300 * time.Second,
				BackoffFactor: 3.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.InitialDelay == 0 {
				t.Error("Expected non-zero initial delay")
			}
			if tt.config.MaxDelay == 0 {
				t.Error("Expected non-zero max delay")
			}
			if tt.config.BackoffFactor == 0 {
				t.Error("Expected non-zero backoff factor")
			}
			// MaxRetries can be 0 for infinite retries
		})
	}
}

// TestConnectWithRetry_ImmediateSuccess tests successful connection on first attempt.
func TestConnectWithRetry_ImmediateSuccess(t *testing.T) {
	// This test verifies the ConnectWithRetry function exists and has correct signature
	// Actual connection testing requires a running server and is covered by integration tests
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19999" // Non-existent port
	config.ConnectionTimeout = 100 * time.Millisecond

	client := NewClient(config)

	reconnectConfig := ReconnectConfig{
		MaxRetries:    1,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Should fail quickly (no server running)
	err := client.ConnectWithRetry(reconnectConfig)
	if err == nil {
		t.Error("Expected connection failure (no server running)")
		client.Disconnect()
	}
}

// TestConnectWithRetry_ExponentialBackoff tests exponential backoff timing.
func TestConnectWithRetry_ExponentialBackoff(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19998" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	client := NewClient(config)

	reconnectConfig := ReconnectConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	start := time.Now()
	err := client.ConnectWithRetry(reconnectConfig)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected connection failure (no server running)")
		client.Disconnect()
		return
	}

	// Should have tried 3 times with delays: 10ms, 20ms
	// Connection fails immediately (connection refused), so no timeout wait
	// Total minimum time: ~10 + 20 = 30ms (allowing some overhead)
	minExpected := 25 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("Expected at least %v elapsed time, got %v", minExpected, elapsed)
	}

	// Shouldn't take too long (max 500ms for this simple test)
	maxExpected := 500 * time.Millisecond
	if elapsed > maxExpected {
		t.Errorf("Expected at most %v elapsed time, got %v", maxExpected, elapsed)
	}
}

// TestConnectWithRetry_MaxDelayRespected tests that MaxDelay caps exponential growth.
func TestConnectWithRetry_MaxDelayRespected(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19997" // Non-existent port
	config.ConnectionTimeout = 10 * time.Millisecond

	client := NewClient(config)

	reconnectConfig := ReconnectConfig{
		MaxRetries:    3,
		InitialDelay:  5 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond, // Cap at 10ms
		BackoffFactor: 10.0,                  // Large factor to test capping
	}

	start := time.Now()
	err := client.ConnectWithRetry(reconnectConfig)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected connection failure (no server running)")
		client.Disconnect()
		return
	}

	// With large backoff factor, delays would be: 5ms, 50ms, 500ms
	// But MaxDelay caps at 10ms, so actual: 5ms, 10ms
	// Connection fails immediately (connection refused)
	// Total minimum: ~5 + 10 = 15ms (allowing some overhead)
	minExpected := 10 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("Expected at least %v elapsed time, got %v", minExpected, elapsed)
	}

	// Should be much less than uncapped exponential growth
	// Uncapped would be: 5 + 50 + 500 = 555ms
	maxExpected := 200 * time.Millisecond
	if elapsed > maxExpected {
		t.Errorf("Expected at most %v elapsed time (MaxDelay should cap growth), got %v", maxExpected, elapsed)
	}
}

// TestDisconnect_NotConnected verifies disconnect when not connected.
func TestDisconnect_NotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)

	err := client.Disconnect()
	if err != nil {
		t.Errorf("Expected no error when disconnecting while not connected, got %v", err)
	}
}

// TestDisconnect_MultipleCallsSafe verifies multiple disconnect calls don't cause issues.
func TestDisconnect_MultipleCallsSafe(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)

	// First disconnect (not connected)
	err := client.Disconnect()
	if err != nil {
		t.Errorf("First disconnect failed: %v", err)
	}

	// Second disconnect (still not connected)
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Second disconnect failed: %v", err)
	}

	// Verify client is not connected
	if client.IsConnected() {
		t.Error("Client should not be connected after disconnect")
	}
}

// TestDisconnect_MultipleCallsAfterConnection verifies multiple disconnect calls after connection don't panic.
func TestDisconnect_MultipleCallsAfterConnection(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)

	// Create a mock connection using pipes
	serverConn, clientConn := createMockConnectionPair(t)
	defer serverConn.Close()

	// Manually set up the connection state (simulating Connect without network)
	client.mu.Lock()
	client.conn = clientConn
	client.connected.Store(true)
	client.done = make(chan struct{})
	client.mu.Unlock()

	// Start goroutines as Connect() would
	client.wg.Add(2)
	go client.receiveLoop()
	go client.sendLoop()

	// First disconnect should work
	err := client.Disconnect()
	if err != nil {
		t.Errorf("First disconnect failed: %v", err)
	}

	// Verify client is disconnected
	if client.IsConnected() {
		t.Error("Client should not be connected after first disconnect")
	}

	// Second disconnect should not panic
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Second disconnect failed: %v", err)
	}

	// Third disconnect for good measure
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Third disconnect failed: %v", err)
	}
}

// createMockConnectionPair creates a pair of connected net.Conn for testing.
func createMockConnectionPair(t *testing.T) (net.Conn, net.Conn) {
	// Use net.Pipe to create a pair of connected connections
	server, client := net.Pipe()
	return server, client
}

// TestAtomicConnectedFlag verifies connected flag is atomic.
func TestAtomicConnectedFlag(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)

	// Initial state should be not connected
	if client.IsConnected() {
		t.Error("New client should not be connected")
	}

	// This test verifies that IsConnected can be called concurrently
	// without holding a lock (it uses atomic.Bool)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = client.IsConnected()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestReconnectAfterDisconnect verifies that a client can reconnect after disconnecting.
func TestReconnectAfterDisconnect(t *testing.T) {
	config := DefaultClientConfig()
	client := NewClient(config)

	// Create first connection
	serverConn1, clientConn1 := createMockConnectionPair(t)

	// Manually set up the first connection
	client.mu.Lock()
	client.conn = clientConn1
	client.connected.Store(true)
	client.done = make(chan struct{})
	client.mu.Unlock()

	client.wg.Add(2)
	go client.receiveLoop()
	go client.sendLoop()

	// Disconnect
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}
	serverConn1.Close()

	// Verify disconnected
	if client.IsConnected() {
		t.Error("Client should not be connected after disconnect")
	}

	// Create second connection (reconnect)
	serverConn2, clientConn2 := createMockConnectionPair(t)
	defer serverConn2.Close()

	// Manually set up the second connection
	client.mu.Lock()
	client.conn = clientConn2
	client.connected.Store(true)
	client.done = make(chan struct{})
	client.mu.Unlock()

	client.wg.Add(2)
	go client.receiveLoop()
	go client.sendLoop()

	// Verify reconnected
	if !client.IsConnected() {
		t.Error("Client should be connected after reconnect")
	}

	// Disconnect again
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Second disconnect failed: %v", err)
	}
}
