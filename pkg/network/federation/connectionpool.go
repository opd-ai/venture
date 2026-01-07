// Package federation provides server-to-server federation protocol.
package federation

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ConnectionConfig holds configuration for connection pool management
type ConnectionConfig struct {
	// MaxIdleTime is how long a connection can be idle before being closed
	MaxIdleTime time.Duration
	// MaxLifetime is the maximum lifetime of a connection
	MaxLifetime time.Duration
	// CleanupInterval is how often to run cleanup of stale connections
	CleanupInterval time.Duration
}

// DefaultConnectionConfig returns a connection configuration with sensible defaults
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		MaxIdleTime:     5 * time.Minute,  // Close idle connections after 5 minutes
		MaxLifetime:     30 * time.Minute, // Close connections after 30 minutes regardless of use
		CleanupInterval: 1 * time.Minute,  // Run cleanup every minute
	}
}

// PooledConnection represents a connection in the pool with metadata
type PooledConnection struct {
	Conn           net.Conn
	ServerID       string
	Address        string
	CreatedAt      time.Time
	LastUsedAt     time.Time
	circuitBreaker *CircuitBreaker
}

// IsStale checks if the connection should be closed based on idle time or lifetime
func (pc *PooledConnection) IsStale(config ConnectionConfig) bool {
	now := time.Now()

	// Check if connection exceeded max lifetime
	if now.Sub(pc.CreatedAt) > config.MaxLifetime {
		return true
	}

	// Check if connection has been idle too long
	if now.Sub(pc.LastUsedAt) > config.MaxIdleTime {
		return true
	}

	return false
}

// ConnectionPool manages a pool of connections to remote servers
type ConnectionPool struct {
	mu            sync.RWMutex
	connections   map[string]*PooledConnection // serverID -> connection
	config        ConnectionConfig
	logger        *logrus.Entry
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	cleanupWg     sync.WaitGroup
}

// NewConnectionPool creates a new connection pool with the given configuration
func NewConnectionPool(config ConnectionConfig) *ConnectionPool {
	pool := &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		config:      config,
		logger: logrus.WithFields(logrus.Fields{
			"component": "connection_pool",
		}),
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	pool.startCleanup()

	return pool
}

// Get retrieves a connection from the pool or creates a new one
func (cp *ConnectionPool) Get(serverID, address string) (net.Conn, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Check if we have a valid connection
	if conn, exists := cp.connections[serverID]; exists {
		// Check if connection is stale
		if conn.IsStale(cp.config) {
			cp.logger.WithFields(logrus.Fields{
				"server_id": serverID,
				"address":   address,
			}).Debug("Connection is stale, removing from pool")
			conn.Conn.Close()
			delete(cp.connections, serverID)
		} else {
			// Update last used time and return existing connection
			conn.LastUsedAt = time.Now()
			cp.logger.WithFields(logrus.Fields{
				"server_id": serverID,
				"address":   address,
			}).Debug("Reusing existing connection from pool")
			return conn.Conn, nil
		}
	}

	// No valid connection exists, create a new one
	cp.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"address":   address,
	}).Debug("Creating new connection")

	return nil, fmt.Errorf("connection not found in pool")
}

// Add adds a new connection to the pool
func (cp *ConnectionPool) Add(serverID, address string, conn net.Conn) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Close existing connection if present
	if existing, exists := cp.connections[serverID]; exists {
		cp.logger.WithFields(logrus.Fields{
			"server_id": serverID,
			"address":   address,
		}).Debug("Replacing existing connection in pool")
		existing.Conn.Close()
	}

	now := time.Now()
	pooledConn := &PooledConnection{
		Conn:           conn,
		ServerID:       serverID,
		Address:        address,
		CreatedAt:      now,
		LastUsedAt:     now,
		circuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
	}

	cp.connections[serverID] = pooledConn

	cp.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"address":   address,
	}).Info("Added connection to pool")

	return nil
}

// Remove removes a connection from the pool
func (cp *ConnectionPool) Remove(serverID string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conn, exists := cp.connections[serverID]; exists {
		conn.Conn.Close()
		delete(cp.connections, serverID)
		cp.logger.WithFields(logrus.Fields{
			"server_id": serverID,
		}).Info("Removed connection from pool")
		return nil
	}

	return fmt.Errorf("connection not found for server %s", serverID)
}

// GetCircuitBreaker returns the circuit breaker for a specific server connection
func (cp *ConnectionPool) GetCircuitBreaker(serverID string) (*CircuitBreaker, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if conn, exists := cp.connections[serverID]; exists {
		return conn.circuitBreaker, true
	}

	return nil, false
}

// Stats returns statistics about the connection pool
func (cp *ConnectionPool) Stats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	totalConnections := len(cp.connections)
	staleCount := 0

	for _, conn := range cp.connections {
		if conn.IsStale(cp.config) {
			staleCount++
		}
	}

	return map[string]interface{}{
		"total_connections":  totalConnections,
		"stale_connections":  staleCount,
		"active_connections": totalConnections - staleCount,
	}
}

// startCleanup starts the background cleanup goroutine
func (cp *ConnectionPool) startCleanup() {
	cp.cleanupTicker = time.NewTicker(cp.config.CleanupInterval)

	cp.cleanupWg.Add(1)
	go func() {
		defer cp.cleanupWg.Done()
		for {
			select {
			case <-cp.cleanupTicker.C:
				cp.cleanup()
			case <-cp.stopCleanup:
				cp.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes stale connections from the pool
func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	removedCount := 0
	for serverID, conn := range cp.connections {
		if conn.IsStale(cp.config) {
			cp.logger.WithFields(logrus.Fields{
				"server_id":  serverID,
				"created_at": conn.CreatedAt,
				"last_used":  conn.LastUsedAt,
			}).Debug("Cleaning up stale connection")
			conn.Conn.Close()
			delete(cp.connections, serverID)
			removedCount++
		}
	}

	if removedCount > 0 {
		cp.logger.WithFields(logrus.Fields{
			"removed_count": removedCount,
		}).Info("Cleaned up stale connections")
	}
}

// Close closes all connections and stops the cleanup goroutine
func (cp *ConnectionPool) Close() error {
	// Stop cleanup goroutine and wait for it to exit
	close(cp.stopCleanup)
	cp.cleanupWg.Wait()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Close all connections
	for serverID, conn := range cp.connections {
		cp.logger.WithFields(logrus.Fields{
			"server_id": serverID,
		}).Debug("Closing connection during pool shutdown")
		conn.Conn.Close()
	}

	// Clear the map
	cp.connections = make(map[string]*PooledConnection)

	cp.logger.Info("Connection pool closed")
	return nil
}
