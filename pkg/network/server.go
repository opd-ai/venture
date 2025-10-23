// Package network provides multiplayer server functionality.
// This file implements Server which handles authoritative game state,
// client connections, and state synchronization for multiplayer games.
package network

import (
	"fmt"
	"net"
	"sync"
	"time"
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

// Server handles server-side networking for multiplayer.
type Server struct {
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

	// Shutdown
	done chan struct{}
	wg   sync.WaitGroup

	// State tracking
	stateSeq uint32
	stateMu  sync.Mutex
}

// clientConnection represents a connected client.
type clientConnection struct {
	playerID   uint64
	conn       net.Conn
	address    string
	connected  bool
	lastActive time.Time

	// Channels
	stateUpdates chan *StateUpdate

	// Thread safety
	mu sync.RWMutex
}

// NewServer creates a new network server.
func NewServer(config ServerConfig) *Server {
	return &Server{
		config:        config,
		protocol:      NewBinaryProtocol(),
		clients:       make(map[uint64]*clientConnection),
		nextPlayerID:  1,
		inputCommands: make(chan *InputCommand, config.BufferSize*config.MaxPlayers),
		playerJoins:   make(chan uint64, config.MaxPlayers),
		playerLeaves:  make(chan uint64, config.MaxPlayers),
		errors:        make(chan error, 64),
		done:          make(chan struct{}),
	}
}

// Start begins listening for client connections.
func (s *Server) Start() error {
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
func (s *Server) Stop() error {
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
func (s *Server) IsRunning() bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.running
}

// GetPlayerCount returns the number of connected players.
func (s *Server) GetPlayerCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// GetPlayers returns a list of connected player IDs.
func (s *Server) GetPlayers() []uint64 {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	players := make([]uint64, 0, len(s.clients))
	for playerID := range s.clients {
		players = append(players, playerID)
	}
	return players
}

// BroadcastStateUpdate sends a state update to all connected clients.
func (s *Server) BroadcastStateUpdate(update *StateUpdate) {
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
func (s *Server) SendStateUpdate(playerID uint64, update *StateUpdate) error {
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
func (s *Server) ReceiveInputCommand() <-chan *InputCommand {
	return s.inputCommands
}

// ReceivePlayerJoin returns a channel for receiving player join events.
func (s *Server) ReceivePlayerJoin() <-chan uint64 {
	return s.playerJoins
}

// ReceivePlayerLeave returns a channel for receiving player leave events.
func (s *Server) ReceivePlayerLeave() <-chan uint64 {
	return s.playerLeaves
}

// ReceiveError returns a channel for receiving errors.
func (s *Server) ReceiveError() <-chan error {
	return s.errors
}

// acceptLoop accepts incoming client connections.
func (s *Server) acceptLoop() {
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

		// Create client connection
		s.clientsMu.Lock()
		playerID := s.nextPlayerID
		s.nextPlayerID++

		client := &clientConnection{
			playerID:     playerID,
			conn:         conn,
			address:      conn.RemoteAddr().String(),
			connected:    true,
			lastActive:   time.Now(),
			stateUpdates: make(chan *StateUpdate, s.config.BufferSize),
		}

		s.clients[playerID] = client
		s.clientsMu.Unlock()

		// Notify game logic of new player
		select {
		case s.playerJoins <- playerID:
		case <-s.done:
			return
		default:
			s.errors <- fmt.Errorf("player join channel full, dropped event for player %d", playerID)
		}

		// Start client handlers
		s.wg.Add(2)
		go s.handleClientReceive(client)
		go s.handleClientSend(client)
	}
}

// handleClientReceive receives data from a client.
func (s *Server) handleClientReceive(client *clientConnection) {
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
		case <-s.done:
			return
		default:
			// Drop if full
		}
	}
}

// handleClientSend sends state updates to a client.
func (s *Server) handleClientSend(client *clientConnection) {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return

		case update := <-client.stateUpdates:
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
	}
}

// disconnectClient removes a client from the server.
func (s *Server) disconnectClient(playerID uint64) {
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
		case <-s.done:
		default:
			s.errors <- fmt.Errorf("player leave channel full, dropped event for player %d", playerID)
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
		close(c.stateUpdates)
	}
}

func (c *clientConnection) sendStateUpdate(update *StateUpdate) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return
	}

	select {
	case c.stateUpdates <- update:
	default:
		// Drop if full (prioritize fresh updates)
	}
}
