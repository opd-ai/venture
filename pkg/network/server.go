// Package network provides multiplayer server functionality.
// This file implements TCPServer which handles authoritative game state,
// client connections, and state synchronization for multiplayer games over TCP.
package network

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ServerConfig holds configuration for the network server.
type ServerConfig struct {
	Address      string        // Listen address (host:port)
	MaxPlayers   int           // Maximum number of concurrent players
	ReadTimeout  time.Duration // Timeout for reading from clients
	WriteTimeout time.Duration // Timeout for writing to clients
	UpdateRate   int           // State updates per second
	BufferSize   int           // Size of send/receive buffers per client
}

// DefaultServerConfig returns a server configuration with sensible defaults.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Address:      ":8080",
		MaxPlayers:   32,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Second,
		UpdateRate:   20, // 20 updates/second
		BufferSize:   256,
	}
}

// HighLatencyServerConfig returns a server configuration optimized for high-latency
// connections (200-5000ms), including Tor/onion services. This configuration uses
// significantly longer timeouts and larger buffers to prevent premature disconnections
// and message drops during latency spikes.
//
// Key differences from DefaultServerConfig:
//   - ReadTimeout: 60s (6x increase) - prevents timeout during extreme latency
//   - WriteTimeout: 30s (6x increase) - allows writes to complete over slow links
//   - BufferSize: 512 (2x increase) - buffers 200+ messages for 10s round-trip at 20Hz
//
// Use this configuration when expecting connections with >1000ms latency or when
// supporting Tor connections which can have 5000ms+ latency with high variance.
func HighLatencyServerConfig() ServerConfig {
	return ServerConfig{
		Address:      ":8080",
		MaxPlayers:   32,
		ReadTimeout:  60 * time.Second, // 60s for extreme latency tolerance
		WriteTimeout: 30 * time.Second, // 30s for write operations over slow links
		UpdateRate:   20,               // 20 updates/second (unchanged)
		BufferSize:   512,              // Increased for buffering during latency spikes
	}
}

// TCPServer handles server-side networking for multiplayer over TCP.
// Implements ServerConnection interface.
type TCPServer struct {
	config   ServerConfig
	protocol Protocol

	// Network state
	listener net.Listener
	running  bool

	// Client management
	clients      map[uint64]*clientConnection
	clientsMu    sync.RWMutex
	nextPlayerID uint64

	// Channels for game logic
	inputCommands chan *InputCommand
	playerJoins   chan uint64 // Player connection events
	playerLeaves  chan uint64 // Player disconnection events
	errors        chan error

	// Buffer monitoring
	inputCommandStats *BufferStats
	playerJoinStats   *BufferStats
	playerLeaveStats  *BufferStats
	errorStats        *BufferStats

	// Shutdown
	done chan struct{}
	wg   sync.WaitGroup

	// State tracking
	stateSeq uint32
	stateMu  sync.Mutex

	// Logger for network operations
	logger *logrus.Entry
}

// clientConnection represents a connected client.
type clientConnection struct {
	playerID   uint64
	conn       net.Conn
	address    string
	connected  bool
	lastActive time.Time

	// Priority queue for state updates (replaces simple channel)
	stateUpdateQueue *StateUpdatePriorityQueue
	// Signal channel to notify send goroutine of new updates
	updateSignal chan struct{}

	// Buffer monitoring
	stateUpdateStats *BufferStats

	// Thread safety
	mu sync.RWMutex
}

// NewServer creates a new network server.
func NewServer(config ServerConfig) *TCPServer {
	return NewServerWithLogger(config, nil)
}

// NewServerWithLogger creates a new network server with a logger.
func NewServerWithLogger(config ServerConfig, logger *logrus.Logger) *TCPServer {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "network_server",
			"address":   config.Address,
		})
	}

	server := &TCPServer{
		config:        config,
		protocol:      NewBinaryProtocol(),
		clients:       make(map[uint64]*clientConnection),
		nextPlayerID:  1,
		inputCommands: make(chan *InputCommand, config.BufferSize*config.MaxPlayers),
		playerJoins:   make(chan uint64, config.MaxPlayers),
		playerLeaves:  make(chan uint64, config.MaxPlayers),
		errors:        make(chan error, 64),
		done:          make(chan struct{}),
		logger:        logEntry,
	}

	// Initialize buffer monitoring
	server.inputCommandStats = NewBufferStats("server_input_commands", config.BufferSize*config.MaxPlayers, logEntry)
	server.playerJoinStats = NewBufferStats("server_player_joins", config.MaxPlayers, logEntry)
	server.playerLeaveStats = NewBufferStats("server_player_leaves", config.MaxPlayers, logEntry)
	server.errorStats = NewBufferStats("server_errors", 64, logEntry)

	return server
}

// Start begins listening for client connections.
func (s *TCPServer) Start() error {
	s.clientsMu.Lock()
	if s.running {
		s.clientsMu.Unlock()
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		s.clientsMu.Unlock()
		return fmt.Errorf("failed to listen on %s: %w", s.config.Address, err)
	}

	s.listener = listener
	s.running = true
	s.clientsMu.Unlock()

	// Start accept loop
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop shuts down the server.
func (s *TCPServer) Stop() error {
	s.clientsMu.Lock()
	if !s.running {
		s.clientsMu.Unlock()
		return nil
	}

	s.running = false
	close(s.done)

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Disconnect all clients
	for _, client := range s.clients {
		client.disconnect()
	}
	s.clientsMu.Unlock()

	// Wait for goroutines
	s.wg.Wait()

	return nil
}

// IsRunning returns whether the server is running.
func (s *TCPServer) IsRunning() bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.running
}

// GetPlayerCount returns the number of connected players.
func (s *TCPServer) GetPlayerCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// GetPlayers returns a list of connected player IDs.
func (s *TCPServer) GetPlayers() []uint64 {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	players := make([]uint64, 0, len(s.clients))
	for playerID := range s.clients {
		players = append(players, playerID)
	}
	return players
}

// BroadcastStateUpdate sends a state update to all connected clients.
func (s *TCPServer) BroadcastStateUpdate(update *StateUpdate) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	// Assign sequence number
	s.stateMu.Lock()
	update.SequenceNumber = s.stateSeq
	s.stateSeq++
	s.stateMu.Unlock()

	// Send to all clients
	for _, client := range s.clients {
		client.sendStateUpdate(update)
	}
}

// SendStateUpdate sends a state update to a specific client.
func (s *TCPServer) SendStateUpdate(playerID uint64, update *StateUpdate) error {
	s.clientsMu.RLock()
	client, exists := s.clients[playerID]
	s.clientsMu.RUnlock()

	if !exists {
		return fmt.Errorf("player %d not connected", playerID)
	}

	// Assign sequence number
	s.stateMu.Lock()
	update.SequenceNumber = s.stateSeq
	s.stateSeq++
	s.stateMu.Unlock()

	client.sendStateUpdate(update)
	return nil
}

// ReceiveInputCommand returns a channel for receiving input commands from clients.
func (s *TCPServer) ReceiveInputCommand() <-chan *InputCommand {
	return s.inputCommands
}

// ReceivePlayerJoin returns a channel for receiving player join events.
func (s *TCPServer) ReceivePlayerJoin() <-chan uint64 {
	return s.playerJoins
}

// ReceivePlayerLeave returns a channel for receiving player leave events.
func (s *TCPServer) ReceivePlayerLeave() <-chan uint64 {
	return s.playerLeaves
}

// ReceiveError returns a channel for receiving errors.
func (s *TCPServer) ReceiveError() <-chan error {
	return s.errors
}

// acceptLoop accepts incoming client connections.
func (s *TCPServer) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				s.errors <- fmt.Errorf("accept error: %w", err)
				continue
			}
		}

		// Check player limit
		s.clientsMu.RLock()
		playerCount := len(s.clients)
		s.clientsMu.RUnlock()

		if playerCount >= s.config.MaxPlayers {
			conn.Close()
			s.errors <- fmt.Errorf("server full, rejected connection from %s", conn.RemoteAddr())
			continue
		}

		// Enable TCP keepalive for long-duration connections
		// This prevents silent disconnections at NAT/proxy boundaries (typical 5-15 min timeout)
		// and helps detect dead connections when network fails without proper FIN/RST
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				if s.logger != nil {
					s.logger.WithError(err).WithField("client", conn.RemoteAddr()).Warn("failed to enable TCP keepalive")
				}
			} else if err := tcpConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
				if s.logger != nil {
					s.logger.WithError(err).WithField("client", conn.RemoteAddr()).Warn("failed to set TCP keepalive period")
				}
			} else if s.logger != nil {
				s.logger.WithField("client", conn.RemoteAddr()).Debug("TCP keepalive enabled (30s period)")
			}
		}

		// Create client connection
		s.clientsMu.Lock()
		playerID := s.nextPlayerID
		s.nextPlayerID++

		// Initialize per-client buffer stats
		var clientLogEntry *logrus.Entry
		if s.logger != nil {
			clientLogEntry = s.logger.WithField("player_id", playerID)
		}

		client := &clientConnection{
			playerID:         playerID,
			conn:             conn,
			address:          conn.RemoteAddr().String(),
			connected:        true,
			lastActive:       time.Now(),
			stateUpdateQueue: NewStateUpdatePriorityQueue(s.config.BufferSize),
			updateSignal:     make(chan struct{}, 1), // Buffered to avoid blocking
			stateUpdateStats: NewBufferStats(fmt.Sprintf("client_%d_state_updates", playerID), s.config.BufferSize, clientLogEntry),
		}

		s.clients[playerID] = client
		s.clientsMu.Unlock()

		// Notify game logic of new player
		select {
		case s.playerJoins <- playerID:
			s.playerJoinStats.RecordSend()
		case <-s.done:
			return
		default:
			// Non-blocking error send
			s.playerJoinStats.RecordDrop()
			select {
			case s.errors <- fmt.Errorf("player join channel full, dropped event for player %d", playerID):
				s.errorStats.RecordSend()
			case <-s.done:
				return
			default:
				// Both channels full - continue without notification
				s.errorStats.RecordDrop()
			}
		}

		// Start client handlers
		s.wg.Add(2)
		go s.handleClientReceive(client)
		go s.handleClientSend(client)
	}
}

// handleClientReceive receives data from a client.
func (s *TCPServer) handleClientReceive(client *clientConnection) {
	defer s.wg.Done()
	defer s.disconnectClient(client.playerID)

	buf := make([]byte, 4096)
	for {
		select {
		case <-s.done:
			return
		default:
		}

		// Set read deadline
		client.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

		// Read message length (4 bytes)
		if _, err := client.conn.Read(buf[:4]); err != nil {
			if s.IsRunning() && client.isConnected() {
				s.errors <- fmt.Errorf("player %d read length error: %w", client.playerID, err)
			}
			return
		}

		// Decode length
		msgLen := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
		if msgLen > uint32(len(buf)) {
			s.errors <- fmt.Errorf("player %d message too large: %d bytes", client.playerID, msgLen)
			return
		}

		// Read message data
		if _, err := client.conn.Read(buf[:msgLen]); err != nil {
			if s.IsRunning() && client.isConnected() {
				s.errors <- fmt.Errorf("player %d read data error: %w", client.playerID, err)
			}
			return
		}

		// Update last active
		client.mu.Lock()
		client.lastActive = time.Now()
		client.mu.Unlock()

		// Decode input command
		cmd, err := s.protocol.DecodeInputCommand(buf[:msgLen])
		if err != nil {
			s.errors <- fmt.Errorf("player %d decode error: %w", client.playerID, err)
			continue
		}

		// Send to game logic (non-blocking)
		select {
		case s.inputCommands <- cmd:
			s.inputCommandStats.RecordSend()
		case <-s.done:
			return
		default:
			// Drop if full
			s.inputCommandStats.RecordDrop()
		}
	}
}

// handleClientSend sends state updates to a client.
// Pops updates from the priority queue, sending higher priority updates first.
func (s *TCPServer) handleClientSend(client *clientConnection) {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return

		case <-client.updateSignal:
			// Process available updates from the priority queue
			// Limit batch size to prevent blocking goroutine for too long
			const maxBatchSize = 20
			batchCount := 0

			for batchCount < maxBatchSize {
				update := client.stateUpdateQueue.Pop()
				if update == nil {
					break // Queue is empty
				}
				batchCount++

				client.stateUpdateStats.RecordReceive()

				// Encode state update
				data, err := s.protocol.EncodeStateUpdate(update)
				if err != nil {
					s.errors <- fmt.Errorf("player %d encode error: %w", client.playerID, err)
					continue
				}

				// Send length prefix
				msgLen := uint32(len(data))
				lenBuf := []byte{
					byte(msgLen),
					byte(msgLen >> 8),
					byte(msgLen >> 16),
					byte(msgLen >> 24),
				}

				// Set write deadline
				client.conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

				// Send length + data
				if _, err := client.conn.Write(lenBuf); err != nil {
					if s.IsRunning() && client.isConnected() {
						s.errors <- fmt.Errorf("player %d write length error: %w", client.playerID, err)
					}
					return
				}
				if _, err := client.conn.Write(data); err != nil {
					if s.IsRunning() && client.isConnected() {
						s.errors <- fmt.Errorf("player %d write data error: %w", client.playerID, err)
					}
					return
				}
			}

			// If we processed a full batch and there are more updates,
			// re-signal to continue processing
			if batchCount == maxBatchSize && !client.stateUpdateQueue.IsEmpty() {
				select {
				case client.updateSignal <- struct{}{}:
				default:
					// Already signaled
				}
			}
		}
	}
}

// disconnectClient removes a client from the server.
func (s *TCPServer) disconnectClient(playerID uint64) {
	s.clientsMu.Lock()
	client, exists := s.clients[playerID]
	if exists {
		client.disconnect()
		delete(s.clients, playerID)
	}
	s.clientsMu.Unlock()

	// Notify game logic of player leave
	if exists {
		select {
		case s.playerLeaves <- playerID:
			s.playerLeaveStats.RecordSend()
		case <-s.done:
		default:
			s.playerLeaveStats.RecordDrop()
			select {
			case s.errors <- fmt.Errorf("player leave channel full, dropped event for player %d", playerID):
				s.errorStats.RecordSend()
			default:
				s.errorStats.RecordDrop()
			}
		}
	}
}

// clientConnection methods

func (c *clientConnection) isConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *clientConnection) disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		close(c.updateSignal)
	}
}

func (c *clientConnection) sendStateUpdate(update *StateUpdate) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return
	}

	// Push to priority queue
	if c.stateUpdateQueue.Push(update) {
		c.stateUpdateStats.RecordSend()

		// Signal the send goroutine (non-blocking)
		select {
		case c.updateSignal <- struct{}{}:
		default:
			// Channel buffer full, goroutine will process existing signal
		}
	} else {
		// Queue is full, drop the update
		c.stateUpdateStats.RecordDrop()
	}
}

// GetBufferStats returns snapshots of all server buffer statistics.
// This provides visibility into channel utilization and potential congestion.
func (s *TCPServer) GetBufferStats() map[string]BufferSnapshot {
	stats := map[string]BufferSnapshot{
		"input_commands": s.inputCommandStats.Snapshot(),
		"player_joins":   s.playerJoinStats.Snapshot(),
		"player_leaves":  s.playerLeaveStats.Snapshot(),
		"errors":         s.errorStats.Snapshot(),
	}

	// Add per-client stats
	s.clientsMu.RLock()
	for playerID, client := range s.clients {
		key := fmt.Sprintf("client_%d_state_updates", playerID)
		stats[key] = client.stateUpdateStats.Snapshot()
	}
	s.clientsMu.RUnlock()

	return stats
}

// Compile-time interface check
var _ ServerConnection = (*TCPServer)(nil)
