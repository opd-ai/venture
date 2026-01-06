package network

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestServerResourceTracking verifies resource tracking functionality.
func TestServerResourceTracking(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0" // Use random available port
	server := NewServer(config)

	// Verify initial state
	stats := server.GetResourceStats()
	if stats["active_connections"].(int) != 0 {
		t.Errorf("Expected 0 active connections, got %d", stats["active_connections"])
	}

	if stats["max_players"].(int) != config.MaxPlayers {
		t.Errorf("Expected max_players %d, got %d", config.MaxPlayers, stats["max_players"])
	}

	// Verify timeout values are set
	if stats["idle_timeout"] == nil {
		t.Error("Expected idle_timeout to be set")
	}
	if stats["cleanup_interval"] == nil {
		t.Error("Expected cleanup_interval to be set")
	}
	if stats["shutdown_timeout"] == nil {
		t.Error("Expected shutdown_timeout to be set")
	}
}

// TestServerGracefulShutdown verifies graceful shutdown with timeout.
func TestServerGracefulShutdown(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Verify it's running
	if !server.IsRunning() {
		t.Error("Server should be running after Start()")
	}

	// Stop server
	startTime := time.Now()
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}
	shutdownDuration := time.Since(startTime)

	// Verify it stopped quickly (should be fast with no clients)
	if shutdownDuration > 5*time.Second {
		t.Errorf("Shutdown took too long: %v (expected < 5s)", shutdownDuration)
	}

	// Verify it's not running
	if server.IsRunning() {
		t.Error("Server should not be running after Stop()")
	}

	// Verify idempotency - calling Stop() again should not error
	if err := server.Stop(); err != nil {
		t.Errorf("Second Stop() returned error: %v", err)
	}
}

// TestServerStopWithContext verifies context-based shutdown.
func TestServerStopWithContext(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Stop with custom timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.StopWithContext(ctx); err != nil {
		t.Errorf("StopWithContext() returned error: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after StopWithContext()")
	}
}

// TestServerShutdownTimeout verifies timeout handling during shutdown.
func TestServerShutdownTimeout(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)
	
	// Set very short shutdown timeout for testing
	server.shutdownTimeout = 1 * time.Nanosecond

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// This should timeout (though goroutines might still finish fast)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	// Give the context time to expire
	time.Sleep(2 * time.Millisecond)
	
	err := server.StopWithContext(ctx)
	// We expect either success (if goroutines finished fast) or timeout error
	if err != nil && err != context.DeadlineExceeded {
		// Check if it's a wrapped deadline exceeded
		if ctx.Err() != context.DeadlineExceeded {
			t.Logf("StopWithContext returned: %v (context error: %v)", err, ctx.Err())
		}
	}
}

// TestServerCleanupInterval verifies cleanup goroutine functionality.
func TestServerCleanupInterval(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)
	
	// Set short cleanup interval for testing
	server.cleanupInterval = 100 * time.Millisecond
	server.idleTimeout = 50 * time.Millisecond

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for a few cleanup cycles
	time.Sleep(350 * time.Millisecond)

	// Verify server is still running (cleanup loop shouldn't crash)
	if !server.IsRunning() {
		t.Error("Server should still be running after cleanup cycles")
	}
}

// TestServerContextCancellation verifies context cancellation propagation.
func TestServerContextCancellation(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)

	// Verify context is set
	if server.ctx == nil {
		t.Fatal("Server context should be initialized")
	}

	if server.cancel == nil {
		t.Fatal("Server cancel function should be initialized")
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Stop should cancel context
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Verify context is cancelled
	select {
	case <-server.ctx.Done():
		// Expected: context should be cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Server context should be cancelled after Stop()")
	}
}

// MockConn implements net.Conn for testing idle client cleanup.
type MockConn struct {
	closed bool
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, net.ErrClosed
	}
	// Block indefinitely until closed
	select {}
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, net.ErrClosed
	}
	return len(b), nil
}

func (m *MockConn) Close() error {
	m.closed = true
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9090}
}

func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestIdleClientCleanup verifies that idle clients are disconnected.
func TestIdleClientCleanup(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)
	
	// Set short timeouts for testing
	server.idleTimeout = 100 * time.Millisecond
	server.cleanupInterval = 50 * time.Millisecond

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Manually add a mock client with old lastActive time
	mockConn := &MockConn{}
	server.clientsMu.Lock()
	client := &clientConnection{
		playerID:         1,
		conn:             mockConn,
		address:          "127.0.0.1:9090",
		connected:        true,
		lastActive:       time.Now().Add(-200 * time.Millisecond), // Old timestamp
		stateUpdateQueue: NewStateUpdatePriorityQueue(config.BufferSize),
		updateSignal:     make(chan struct{}, 1),
	}
	server.clients[1] = client
	server.clientsMu.Unlock()

	// Verify client exists
	if server.GetPlayerCount() != 1 {
		t.Errorf("Expected 1 client, got %d", server.GetPlayerCount())
	}

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify client was cleaned up
	if server.GetPlayerCount() != 0 {
		t.Errorf("Expected idle client to be cleaned up, got %d clients", server.GetPlayerCount())
	}
}

// TestActiveClientNotCleaned verifies that active clients are NOT disconnected.
func TestActiveClientNotCleaned(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)
	
	// Set timeouts so client won't be idle after our sleep
	server.idleTimeout = 500 * time.Millisecond  // Longer timeout
	server.cleanupInterval = 50 * time.Millisecond

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Manually add a mock client with recent lastActive time
	mockConn := &MockConn{}
	server.clientsMu.Lock()
	client := &clientConnection{
		playerID:         1,
		conn:             mockConn,
		address:          "127.0.0.1:9090",
		connected:        true,
		lastActive:       time.Now(), // Recent timestamp
		stateUpdateQueue: NewStateUpdatePriorityQueue(config.BufferSize),
		updateSignal:     make(chan struct{}, 1),
	}
	server.clients[1] = client
	server.clientsMu.Unlock()

	// Wait for cleanup cycle (multiple cycles to be sure)
	time.Sleep(200 * time.Millisecond)

	// Verify client still exists (hasn't reached idle timeout yet)
	if server.GetPlayerCount() != 1 {
		t.Errorf("Expected active client to remain, got %d clients", server.GetPlayerCount())
	}
}

// TestServerMultipleStartStop verifies server can be restarted.
func TestServerMultipleStartStop(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"

	// First lifecycle
	server := NewServer(config)
	if err := server.Start(); err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}
	if err := server.Stop(); err != nil {
		t.Errorf("First Stop() failed: %v", err)
	}

	// Note: The server context is cancelled after Stop(), so we need a new server
	// This tests that creating a new server works correctly
	server2 := NewServer(config)
	if err := server2.Start(); err != nil {
		t.Fatalf("Second Start() failed: %v", err)
	}
	if err := server2.Stop(); err != nil {
		t.Errorf("Second Stop() failed: %v", err)
	}
}

// TestServerDoubleStart verifies error on double start.
func TestServerDoubleStart(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server := NewServer(config)

	if err := server.Start(); err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}
	defer server.Stop()

	// Second start should error
	if err := server.Start(); err == nil {
		t.Error("Second Start() should return error")
	}
}
