package federation

import (
	"net"
	"testing"
	"time"
)

// mockConn is a mock implementation of net.Conn for testing
type mockConn struct {
	closed bool
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNewConnectionPool(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	if pool == nil {
		t.Fatal("NewConnectionPool returned nil")
	}

	stats := pool.Stats()
	if stats["total_connections"].(int) != 0 {
		t.Error("Expected empty pool initially")
	}
}

func TestConnectionPoolAdd(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	conn := &mockConn{}
	err := pool.Add("server1", "localhost:8080", conn)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	stats := pool.Stats()
	if stats["total_connections"].(int) != 1 {
		t.Error("Expected 1 connection in pool")
	}
}

func TestConnectionPoolGet(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add a connection
	conn1 := &mockConn{}
	pool.Add("server1", "localhost:8080", conn1)

	// Get the connection
	conn2, err := pool.Get("server1", "localhost:8080")
	if err != nil {
		t.Errorf("Expected to get connection, got error: %v", err)
	}

	if conn2 != conn1 {
		t.Error("Expected to get the same connection")
	}
}

func TestConnectionPoolGetNonExistent(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	_, err := pool.Get("server1", "localhost:8080")
	if err == nil {
		t.Error("Expected error for non-existent connection")
	}
}

func TestConnectionPoolRemove(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add a connection
	conn := &mockConn{}
	pool.Add("server1", "localhost:8080", conn)

	// Remove it
	err := pool.Remove("server1")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !conn.closed {
		t.Error("Expected connection to be closed")
	}

	stats := pool.Stats()
	if stats["total_connections"].(int) != 0 {
		t.Error("Expected empty pool after removal")
	}
}

func TestConnectionPoolRemoveNonExistent(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	err := pool.Remove("server1")
	if err == nil {
		t.Error("Expected error when removing non-existent connection")
	}
}

func TestConnectionPoolReplaceExisting(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add first connection
	conn1 := &mockConn{}
	pool.Add("server1", "localhost:8080", conn1)

	// Add second connection with same server ID
	conn2 := &mockConn{}
	pool.Add("server1", "localhost:8080", conn2)

	// First connection should be closed
	if !conn1.closed {
		t.Error("Expected first connection to be closed when replaced")
	}

	// Pool should have one connection
	stats := pool.Stats()
	if stats["total_connections"].(int) != 1 {
		t.Error("Expected 1 connection in pool")
	}

	// Get should return second connection
	conn3, err := pool.Get("server1", "localhost:8080")
	if err != nil {
		t.Errorf("Expected to get connection, got error: %v", err)
	}

	if conn3 != conn2 {
		t.Error("Expected to get the second connection")
	}
}

func TestPooledConnectionIsStale(t *testing.T) {
	config := ConnectionConfig{
		MaxIdleTime: 100 * time.Millisecond,
		MaxLifetime: 200 * time.Millisecond,
	}

	tests := []struct {
		name      string
		createdAt time.Time
		lastUsed  time.Time
		expected  bool
	}{
		{
			name:      "fresh_connection",
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			expected:  false,
		},
		{
			name:      "idle_too_long",
			createdAt: time.Now(),
			lastUsed:  time.Now().Add(-150 * time.Millisecond),
			expected:  true,
		},
		{
			name:      "exceeded_max_lifetime",
			createdAt: time.Now().Add(-250 * time.Millisecond),
			lastUsed:  time.Now(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PooledConnection{
				Conn:       &mockConn{},
				ServerID:   "test",
				Address:    "localhost:8080",
				CreatedAt:  tt.createdAt,
				LastUsedAt: tt.lastUsed,
			}

			if pc.IsStale(config) != tt.expected {
				t.Errorf("IsStale() = %v, expected %v", pc.IsStale(config), tt.expected)
			}
		})
	}
}

func TestConnectionPoolGetStaleConnection(t *testing.T) {
	config := ConnectionConfig{
		MaxIdleTime:     50 * time.Millisecond,
		MaxLifetime:     200 * time.Millisecond,
		CleanupInterval: 1 * time.Second, // High interval so cleanup doesn't interfere
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add a connection
	conn := &mockConn{}
	pool.Add("server1", "localhost:8080", conn)

	// Wait for connection to become stale
	time.Sleep(100 * time.Millisecond)

	// Try to get the stale connection - should fail and close the connection
	_, err := pool.Get("server1", "localhost:8080")
	if err == nil {
		t.Error("Expected error when getting stale connection")
	}

	if !conn.closed {
		t.Error("Expected stale connection to be closed")
	}
}

func TestConnectionPoolCleanup(t *testing.T) {
	config := ConnectionConfig{
		MaxIdleTime:     50 * time.Millisecond,
		MaxLifetime:     200 * time.Millisecond,
		CleanupInterval: 100 * time.Millisecond,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add connections
	conn1 := &mockConn{}
	conn2 := &mockConn{}
	pool.Add("server1", "localhost:8080", conn1)
	pool.Add("server2", "localhost:8081", conn2)

	stats := pool.Stats()
	if stats["total_connections"].(int) != 2 {
		t.Error("Expected 2 connections initially")
	}

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Connections should be cleaned up
	stats = pool.Stats()
	if stats["total_connections"].(int) != 0 {
		t.Errorf("Expected 0 connections after cleanup, got %d", stats["total_connections"].(int))
	}

	if !conn1.closed || !conn2.closed {
		t.Error("Expected connections to be closed by cleanup")
	}
}

func TestConnectionPoolClose(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)

	// Add connections
	conn1 := &mockConn{}
	conn2 := &mockConn{}
	pool.Add("server1", "localhost:8080", conn1)
	pool.Add("server2", "localhost:8081", conn2)

	// Close pool
	err := pool.Close()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// All connections should be closed
	if !conn1.closed || !conn2.closed {
		t.Error("Expected all connections to be closed")
	}

	stats := pool.Stats()
	if stats["total_connections"].(int) != 0 {
		t.Error("Expected empty pool after close")
	}
}

func TestConnectionPoolGetCircuitBreaker(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	// Add a connection
	conn := &mockConn{}
	pool.Add("server1", "localhost:8080", conn)

	// Get circuit breaker
	cb, exists := pool.GetCircuitBreaker("server1")
	if !exists {
		t.Error("Expected to find circuit breaker")
	}

	if cb == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	if cb.State() != CircuitStateClosed {
		t.Error("Expected circuit breaker to be in Closed state initially")
	}
}

func TestConnectionPoolGetCircuitBreakerNonExistent(t *testing.T) {
	config := DefaultConnectionConfig()
	pool := NewConnectionPool(config)
	defer pool.Close()

	cb, exists := pool.GetCircuitBreaker("server1")
	if exists {
		t.Error("Expected not to find circuit breaker for non-existent connection")
	}

	if cb != nil {
		t.Error("Expected nil circuit breaker for non-existent connection")
	}
}

func TestDefaultConnectionConfig(t *testing.T) {
	config := DefaultConnectionConfig()

	if config.MaxIdleTime <= 0 {
		t.Error("MaxIdleTime should be positive")
	}

	if config.MaxLifetime <= 0 {
		t.Error("MaxLifetime should be positive")
	}

	if config.CleanupInterval <= 0 {
		t.Error("CleanupInterval should be positive")
	}

	if config.MaxLifetime <= config.MaxIdleTime {
		t.Error("MaxLifetime should be greater than MaxIdleTime")
	}
}
