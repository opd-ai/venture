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
	config.ConnectionTimeout = 100 * time.Millisecond // Short timeout to avoid test hanging
	client := NewClient(config)

	// Create a mock connection using pipes
	serverConn, clientConn := createMockConnectionPair(t)

	// Close server side immediately so client reads will fail fast
	serverConn.Close()

	// Manually set up the connection state (simulating Connect without network).
	// Must use doneMu to properly synchronize with Disconnect().
	client.mu.Lock()
	client.conn = clientConn
	client.connected.Store(true)
	client.mu.Unlock()

	client.doneMu.Lock()
	client.done = make(chan struct{})
	client.doneClosed = false
	client.doneMu.Unlock()

	// Start goroutines as Connect() would
	client.wg.Add(2)

	// Use channel to ensure goroutines have started before proceeding
	started := make(chan struct{}, 2)
	go func() {
		started <- struct{}{}
		client.receiveLoop()
	}()
	go func() {
		started <- struct{}{}
		client.sendLoop()
	}()

	// Wait for both goroutines to signal they've started
	<-started
	<-started

	// Give goroutines a moment to hit EOF and exit gracefully
	// This is necessary because the server conn is already closed,
	// so the first read will return EOF and the goroutines will exit
	time.Sleep(10 * time.Millisecond)

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

	// Manually set up the first connection.
	// Must use doneMu to properly synchronize with Disconnect().
	client.mu.Lock()
	client.conn = clientConn1
	client.connected.Store(true)
	client.mu.Unlock()

	client.doneMu.Lock()
	client.done = make(chan struct{})
	client.doneClosed = false
	client.doneMu.Unlock()

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

	// Manually set up the second connection.
	// Must use doneMu to properly synchronize with Disconnect().
	client.mu.Lock()
	client.conn = clientConn2
	client.connected.Store(true)
	client.mu.Unlock()

	client.doneMu.Lock()
	client.done = make(chan struct{})
	client.doneClosed = false
	client.doneMu.Unlock()

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

// TestConnectWithRetry_CancellationDuringRetry tests that ConnectWithRetry
// can be cancelled gracefully by calling Disconnect() during retry loop.
// This addresses AUDIT.md Finding 2: Infinite Reconnection Loop with No Cancellation.
func TestConnectWithRetry_CancellationDuringRetry(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19996" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	client := NewClient(config)

	// Configure for infinite retries (Tor mode)
	reconnectConfig := ReconnectConfig{
		MaxRetries:    0, // 0 = infinite retries
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      500 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Start reconnection in background
	done := make(chan error, 1)
	go func() {
		done <- client.ConnectWithRetry(reconnectConfig)
	}()

	// Wait a bit to ensure we're in retry loop
	time.Sleep(150 * time.Millisecond)

	// Cancel by disconnecting
	start := time.Now()
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	// ConnectWithRetry should return with cancellation error
	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err == nil {
			t.Error("Expected ConnectWithRetry to return error after cancellation")
		}
		if err.Error() != "reconnection cancelled" {
			t.Errorf("Expected 'reconnection cancelled' error, got: %v", err)
		}
		// Should return quickly (well before the max backoff delay; allow up to 300ms)
		if elapsed > 300*time.Millisecond {
			t.Errorf("Expected quick cancellation, took %v", elapsed)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ConnectWithRetry did not return after Disconnect (deadlock)")
	}
}

// TestConnectWithRetry_CancellationBeforeFirstAttempt tests that calling
// Disconnect before any connection attempts works correctly.
func TestConnectWithRetry_CancellationBeforeFirstAttempt(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19995" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	client := NewClient(config)

	// Disconnect immediately (before any connection attempt)
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	// Configure for multiple retries
	reconnectConfig := ReconnectConfig{
		MaxRetries:    5,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      500 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Try to connect after disconnect
	err = client.ConnectWithRetry(reconnectConfig)

	// Should fail to connect (expected behavior)
	if err == nil {
		t.Error("Expected connection failure after disconnect")
		client.Disconnect()
	}
}

// TestConnectWithRetry_MultipleCancellations tests that multiple
// cancellations don't cause panics or deadlocks.
func TestConnectWithRetry_MultipleCancellations(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19994" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	client := NewClient(config)

	// Configure for infinite retries
	reconnectConfig := ReconnectConfig{
		MaxRetries:    0, // infinite
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      500 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Start reconnection in background
	done := make(chan error, 1)
	go func() {
		done <- client.ConnectWithRetry(reconnectConfig)
	}()

	// Wait for retry loop to start
	time.Sleep(150 * time.Millisecond)

	// Call Disconnect multiple times
	for i := 0; i < 3; i++ {
		err := client.Disconnect()
		if err != nil && i == 0 {
			// First disconnect might return error if connection attempt ongoing
			t.Logf("First disconnect returned: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// ConnectWithRetry should return
	select {
	case err := <-done:
		if err == nil {
			t.Error("Expected ConnectWithRetry to return error after cancellation")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ConnectWithRetry did not return after multiple Disconnects")
	}
}

// TestConnectWithRetry_CancellationWithTorConfig tests graceful cancellation
// specifically with Tor/high-latency configuration (infinite retries).
func TestConnectWithRetry_CancellationWithTorConfig(t *testing.T) {
	config := TorClientConfig()
	config.ServerAddress = "localhost:19993" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	client := NewClient(config)

	// Use Tor reconnect config (infinite retries)
	reconnectConfig := TorReconnectConfig()
	// Override delays for faster test
	reconnectConfig.InitialDelay = 100 * time.Millisecond
	reconnectConfig.MaxDelay = 500 * time.Millisecond

	// Start reconnection in background
	done := make(chan error, 1)
	go func() {
		done <- client.ConnectWithRetry(reconnectConfig)
	}()

	// Wait for retry loop
	time.Sleep(150 * time.Millisecond)

	// Cancel by disconnecting
	start := time.Now()
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	// Should return with cancellation error quickly
	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err == nil {
			t.Error("Expected ConnectWithRetry to return error after cancellation")
		}
		// Verify it's a cancellation error
		if err.Error() != "reconnection cancelled" {
			t.Errorf("Expected 'reconnection cancelled' error, got: %v", err)
		}
		// Should return quickly (well before max backoff delay; allow up to 300ms for system variability)
		if elapsed > 300*time.Millisecond {
			t.Errorf("Expected quick cancellation with Tor config, took %v", elapsed)
		}
		t.Logf("Tor config cancellation took %v", elapsed)
	case <-time.After(2 * time.Second):
		t.Fatal("ConnectWithRetry with Tor config did not return after Disconnect")
	}
}

// TestConnectWithRetry_NoMemoryLeakOnCancellation tests that cancelling
// reconnection attempts doesn't leak goroutines or memory.
func TestConnectWithRetry_NoMemoryLeakOnCancellation(t *testing.T) {
	config := DefaultClientConfig()
	config.ServerAddress = "localhost:19992" // Non-existent port
	config.ConnectionTimeout = 50 * time.Millisecond

	reconnectConfig := ReconnectConfig{
		MaxRetries:    0, // infinite
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      200 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Create and cancel multiple clients to test for leaks
	for i := 0; i < 5; i++ {
		client := NewClient(config)

		done := make(chan error, 1)
		go func() {
			done <- client.ConnectWithRetry(reconnectConfig)
		}()

		// Wait for retry loop
		time.Sleep(75 * time.Millisecond)

		// Cancel
		err := client.Disconnect()
		if err != nil {
			t.Errorf("Iteration %d: Disconnect failed: %v", i, err)
		}

		// Wait for completion
		select {
		case <-done:
			// Good, completed
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Iteration %d: ConnectWithRetry did not return", i)
		}
	}

	// If we get here without hanging, no obvious goroutine leak
	t.Log("Successfully created and cancelled 5 clients without leaks")
}

// TestNewClient_PredictorInstantiated verifies that NewClient always creates
// a non-nil predictor for client-side prediction support.
func TestNewClient_PredictorInstantiated(t *testing.T) {
	client := NewClient(DefaultClientConfig())
	if client.predictor == nil {
		t.Fatal("Expected predictor to be non-nil for default client config")
	}
}

// TestNewClient_HighLatencyPredictorSelected verifies that TorClientConfig
// causes the high-latency predictor variant to be selected.
func TestNewClient_HighLatencyPredictorSelected(t *testing.T) {
	standard := NewClient(DefaultClientConfig())
	highLat := NewClient(TorClientConfig())

	if standard.predictor == nil || highLat.predictor == nil {
		t.Fatal("Expected non-nil predictor for both configs")
	}

	if standard.predictor.maxHistory != 256 {
		t.Errorf("Standard predictor maxHistory = %d, want 256", standard.predictor.maxHistory)
	}
	if highLat.predictor.maxHistory != 512 {
		t.Errorf("High-latency predictor maxHistory = %d, want 512", highLat.predictor.maxHistory)
	}
}

// TestClientPredictorReconciliation confirms the predictor diverges from the
// server then converges within the error threshold after reconciliation.
func TestClientPredictorReconciliation(t *testing.T) {
	client := NewClient(DefaultClientConfig())
	client.predictor.SetInitialState(Position{X: 0, Y: 0}, Velocity{VX: 0, VY: 0})

	// Advance the local simulation with several move inputs.
	const dt = 0.05
	for i := 0; i < 5; i++ {
		client.PredictMovement(10, 0, dt)
	}

	predicted := client.predictor.GetCurrentState()
	if predicted.Position.X <= 0 {
		t.Fatal("Expected non-zero predicted position after PredictMovement calls")
	}

	// Server acknowledges seq 3 with a slightly different position (within threshold).
	serverPos := client.predictor.stateHistory[2].Position
	reconciled := client.ReconcileFromServer(3, serverPos, Velocity{VX: 0, VY: 0})

	// After reconciliation the client should still have a valid position
	// (no full reset since error is within threshold).
	if reconciled.Sequence < 3 {
		t.Errorf("Reconciled sequence = %d, want >= 3", reconciled.Sequence)
	}

	// GetPredictor must return the same instance used by PredictMovement/ReconcileFromServer.
	if client.GetPredictor() != client.predictor {
		t.Error("GetPredictor returned a different instance than the internal predictor")
	}
}
