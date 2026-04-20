// Package network provides multiplayer server functionality.
// This file implements TCPServer which handles authoritative game state,
// client connections, and state synchronization for multiplayer games over TCP.
package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
	"github.com/sirupsen/logrus"
)

const (
	// errorBufferSize is the size of the error channel buffer.
	// This buffer holds async errors from network operations (connection failures,
	// protocol errors, etc.) before they can be consumed by the game logic.
	// A size of 64 provides sufficient buffering for error bursts while preventing
	// unbounded memory growth from unconsumed errors.
	errorBufferSize = 64
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
//
// Resource Management:
// The server implements comprehensive resource tracking and cleanup to prevent leaks:
// - Context-based cancellation for all long-running operations
// - Graceful shutdown with configurable timeout (default 30s)
// - Periodic cleanup of idle client connections
// - Resource tracking (active connections, goroutines)
// - Automatic cleanup of abandoned sessions
//
// Lifecycle:
// 1. Start() - Begins accepting connections and starts cleanup goroutine
// 2. Operations - Handles clients with context-aware operations
// 3. Stop() or StopWithContext() - Gracefully shuts down with timeout
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

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Shutdown
	done chan struct{}
	wg   sync.WaitGroup

	// Resource tracking
	idleTimeout     time.Duration
	cleanupInterval time.Duration
	shutdownTimeout time.Duration

	// State tracking
	stateSeq uint32
	stateMu  sync.Mutex

	// Metrics tracking
	metricsMu        sync.RWMutex
	totalBytesSent   uint64
	totalBytesRecv   uint64
	totalPacketsSent uint64
	totalPacketsRecv uint64

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

	// Create context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	server := &TCPServer{
		config:        config,
		protocol:      NewBinaryProtocol(),
		clients:       make(map[uint64]*clientConnection),
		nextPlayerID:  1,
		inputCommands: make(chan *InputCommand, config.BufferSize*config.MaxPlayers),
		playerJoins:   make(chan uint64, config.MaxPlayers),
		playerLeaves:  make(chan uint64, config.MaxPlayers),
		errors:        make(chan error, errorBufferSize),
		ctx:           ctx,
		cancel:        cancel,
		done:          make(chan struct{}),
		logger:        logEntry,
		// Resource management defaults
		idleTimeout:     5 * time.Minute,  // Disconnect clients idle for 5 minutes
		cleanupInterval: 30 * time.Second, // Check for idle clients every 30 seconds
		shutdownTimeout: 30 * time.Second, // Wait up to 30 seconds for graceful shutdown
	}

	// Initialize buffer monitoring
	server.inputCommandStats = NewBufferStats("server_input_commands", config.BufferSize*config.MaxPlayers, logEntry)
	server.playerJoinStats = NewBufferStats("server_player_joins", config.MaxPlayers, logEntry)
	server.playerLeaveStats = NewBufferStats("server_player_leaves", config.MaxPlayers, logEntry)
	server.errorStats = NewBufferStats("server_errors", errorBufferSize, logEntry)

	return server
}

// Start begins listening for client connections.
// Starts the accept loop and periodic cleanup of idle connections.
func (s *TCPServer) Start() error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.config.Address, err)
	}

	s.listener = listener
	s.running = true

	// Start accept loop (unlock before spawning goroutine to prevent race)
	s.clientsMu.Unlock()
	s.wg.Add(1)
	go s.acceptLoop()

	// Start cleanup goroutine for idle connection management
	s.wg.Add(1)
	go s.cleanupLoop()
	s.clientsMu.Lock()

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"idle_timeout":     s.idleTimeout,
			"cleanup_interval": s.cleanupInterval,
		}).Debug("started idle connection cleanup")
	}

	return nil
}

// Stop stops the server and closes all connections with default timeout.
// Uses the configured shutdownTimeout (default 30s) for graceful shutdown.
func (s *TCPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.StopWithContext(ctx)
}

// StopWithContext stops the server with a custom context for shutdown timeout control.
// Returns an error if shutdown doesn't complete within the context deadline.
// All client connections are closed immediately, but goroutines are given time to cleanup.
func (s *TCPServer) StopWithContext(ctx context.Context) error {
	if !s.initiateShutdown() {
		return nil
	}

	clientCount := s.disconnectAllClients()
	s.logShutdownStart(clientCount)

	return s.waitForGoroutines(ctx)
}

// initiateShutdown begins the shutdown process.
func (s *TCPServer) initiateShutdown() bool {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if !s.running {
		return false
	}

	s.running = false
	s.cancel()
	close(s.done)

	if s.listener != nil {
		s.listener.Close()
	}

	return true
}

// disconnectAllClients disconnects all connected clients.
func (s *TCPServer) disconnectAllClients() int {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clientCount := len(s.clients)
	for _, client := range s.clients {
		client.disconnect()
	}

	return clientCount
}

// logShutdownStart logs the beginning of shutdown process.
func (s *TCPServer) logShutdownStart(clientCount int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"client_count": clientCount,
			"timeout":      s.shutdownTimeout,
		}).Info("shutting down server, waiting for goroutines")
	}
}

// waitForGoroutines waits for all goroutines to finish with timeout.
func (s *TCPServer) waitForGoroutines(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logShutdownComplete()
		return nil
	case <-ctx.Done():
		s.logShutdownTimeout()
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// logShutdownComplete logs successful shutdown completion.
func (s *TCPServer) logShutdownComplete() {
	if s.logger != nil {
		s.logger.Info("server shutdown complete")
	}
}

// logShutdownTimeout logs when shutdown times out.
func (s *TCPServer) logShutdownTimeout() {
	if s.logger != nil {
		s.logger.Warn("server shutdown timed out, some goroutines may still be running")
	}
}

// IsRunning returns whether the server is running.
func (s *TCPServer) IsRunning() bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.running
}

// Address returns the actual network address the server is listening on, or
// an empty string if the server has not started. Useful for tests that bind
// to ":0" (an OS-assigned ephemeral port) and need to discover the port.
func (s *TCPServer) Address() string {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// cleanupLoop periodically checks for and disconnects idle clients.
// This prevents resource leaks from abandoned connections.
func (s *TCPServer) cleanupLoop() {
	defer s.wg.Done()
	defer recovery.RecoverPanic(s.logger, "cleanup loop", nil)()

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// Server context cancelled, exit cleanup loop
			return
		case <-ticker.C:
			s.cleanupIdleClients()
		}
	}
}

// cleanupIdleClients disconnects clients that have been idle beyond the timeout.
func (s *TCPServer) cleanupIdleClients() {
	now := time.Now()
	var idleClients []uint64

	// Find idle clients
	s.clientsMu.RLock()
	for playerID, client := range s.clients {
		client.mu.RLock()
		lastActive := client.lastActive
		client.mu.RUnlock()

		if now.Sub(lastActive) > s.idleTimeout {
			idleClients = append(idleClients, playerID)
		}
	}
	s.clientsMu.RUnlock()

	// Disconnect idle clients
	if len(idleClients) > 0 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"count":        len(idleClients),
				"idle_timeout": s.idleTimeout,
			}).Info("disconnecting idle clients")
		}

		for _, playerID := range idleClients {
			s.disconnectClient(playerID)
		}
	}
}

// GetResourceStats returns current resource usage statistics.
// Useful for monitoring and detecting resource leaks.
func (s *TCPServer) GetResourceStats() map[string]interface{} {
	s.clientsMu.RLock()
	clientCount := len(s.clients)
	s.clientsMu.RUnlock()

	return map[string]interface{}{
		"active_connections": clientCount,
		"max_players":        s.config.MaxPlayers,
		"idle_timeout":       s.idleTimeout.String(),
		"cleanup_interval":   s.cleanupInterval.String(),
		"shutdown_timeout":   s.shutdownTimeout.String(),
	}
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
// Automatically assigns a monotonically increasing sequence number to the update
// before broadcasting, enabling clients to detect out-of-order or missing updates.
//
// The function is non-blocking: updates are queued for async transmission per client.
// If a client's send queue is full, that client will skip this update (client-side
// interpolation and server reconciliation handle missing updates).
//
// Thread-safety: Safe to call concurrently. Uses RLock for client iteration and
// separate lock for sequence number assignment.
//
// Performance: O(n) where n is number of connected clients. For high player counts
// (>100), consider using interest management or spatial partitioning to reduce
// broadcast scope.
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
	defer recovery.RecoverPanic(s.logger, "accept loop", nil)()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.handleAcceptError(err) {
				return
			}
			continue
		}

		if s.isServerFull(conn) {
			continue
		}

		s.configureTCPKeepalive(conn)

		client := s.createClientConnection(conn)

		if s.notifyPlayerJoin(client.playerID) {
			return
		}

		s.startClientHandlers(client)
	}
}

// handleAcceptError processes errors from the accept loop and returns true if the loop should exit.
func (s *TCPServer) handleAcceptError(err error) bool {
	select {
	case <-s.ctx.Done():
		// Server context cancelled, exit accept loop
		return true
	default:
		s.errors <- fmt.Errorf("accept error: %w", err)
		return false
	}
}

// isServerFull checks if the server has reached max player capacity and rejects the connection if so.
func (s *TCPServer) isServerFull(conn net.Conn) bool {
	s.clientsMu.RLock()
	playerCount := len(s.clients)
	s.clientsMu.RUnlock()

	if playerCount >= s.config.MaxPlayers {
		conn.Close()
		s.errors <- fmt.Errorf("server full, rejected connection from %s", conn.RemoteAddr())
		return true
	}
	return false
}

// configureTCPKeepalive enables TCP keepalive for long-duration connections to prevent silent disconnections.
// Uses KeepAliveConn interface for testability and flexible connection handling.
func (s *TCPServer) configureTCPKeepalive(conn net.Conn) {
	kaConn, ok := conn.(KeepAliveConn)
	if !ok {
		return
	}

	if err := kaConn.SetKeepAlive(true); err != nil {
		if s.logger != nil {
			s.logger.WithError(err).WithField("client", conn.RemoteAddr()).Warn("failed to enable TCP keepalive")
		}
		return
	}

	if err := kaConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
		if s.logger != nil {
			s.logger.WithError(err).WithField("client", conn.RemoteAddr()).Warn("failed to set TCP keepalive period")
		}
		return
	}

	if s.logger != nil {
		s.logger.WithField("client", conn.RemoteAddr()).Debug("TCP keepalive enabled (30s period)")
	}
}

// createClientConnection creates and registers a new client connection with the server.
func (s *TCPServer) createClientConnection(conn net.Conn) *clientConnection {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	playerID := s.nextPlayerID
	s.nextPlayerID++

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
		updateSignal:     make(chan struct{}, 1),
		stateUpdateStats: NewBufferStats(fmt.Sprintf("client_%d_state_updates", playerID), s.config.BufferSize, clientLogEntry),
	}

	s.clients[playerID] = client
	return client
}

// notifyPlayerJoin notifies the game logic of a new player join and returns true if the server should exit.
func (s *TCPServer) notifyPlayerJoin(playerID uint64) bool {
	select {
	case s.playerJoins <- playerID:
		s.playerJoinStats.RecordSend()
		return false
	case <-s.done:
		return true
	default:
		return s.handlePlayerJoinDrop(playerID)
	}
}

// handlePlayerJoinDrop handles the case where the player join channel is full.
func (s *TCPServer) handlePlayerJoinDrop(playerID uint64) bool {
	s.playerJoinStats.RecordDrop()
	select {
	case s.errors <- fmt.Errorf("player join channel full, dropped event for player %d", playerID):
		s.errorStats.RecordSend()
		return false
	case <-s.done:
		return true
	default:
		s.errorStats.RecordDrop()
		return false
	}
}

// startClientHandlers starts the receive and send goroutines for a client connection.
func (s *TCPServer) startClientHandlers(client *clientConnection) {
	s.wg.Add(2)
	go s.handleClientReceive(client)
	go s.handleClientSend(client)
}

// handleClientReceive receives data from a client.
// handleClientReceive processes incoming commands from a client.
// Performance optimization: The done channel is NOT checked in the hot path to minimize
// overhead on every packet receive. Instead, we rely on SetReadDeadline and connection
// closure (triggered by server shutdown) to exit the loop. This eliminates ~1-2% CPU
// overhead at high packet rates by avoiding select statement in the hot path.
func (s *TCPServer) handleClientReceive(client *clientConnection) {
	defer s.wg.Done()
	defer s.disconnectClient(client.playerID)
	defer recovery.RecoverPanic(s.logger, "client receive handler", nil)()

	buf := make([]byte, 4096)
	for {
		// Hot path optimization: No done channel check here.
		// Server shutdown closes all client connections, causing reads to fail and exit.

		msgLen, err := s.readMessageLength(client, buf)
		if err != nil {
			return
		}

		if err := s.readMessageData(client, buf, msgLen); err != nil {
			return
		}

		client.updateLastActive()

		cmd, err := s.decodeCommand(client, buf, msgLen)
		if err != nil {
			continue
		}

		// Voice packets are routed by connection-bound playerID rather than
		// the client-supplied cmd.PlayerID (which may be 0 or spoofed). All
		// other inputs are forwarded to game logic unchanged.
		if cmd != nil && cmd.InputType == VoiceInputType {
			s.routeVoiceCommand(cmd, client.playerID)
			continue
		}

		s.sendCommandToGameLogic(cmd)
	}
}

func (s *TCPServer) readMessageLength(client *clientConnection, buf []byte) (uint32, error) {
	client.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

	if _, err := client.conn.Read(buf[:4]); err != nil {
		if s.IsRunning() && client.isConnected() {
			s.errors <- fmt.Errorf("player %d read length error: %w", client.playerID, err)
		}
		return 0, err
	}

	msgLen := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
	if msgLen > uint32(len(buf)) {
		err := fmt.Errorf("player %d message too large: %d bytes", client.playerID, msgLen)
		s.errors <- err
		return 0, err
	}

	return msgLen, nil
}

func (s *TCPServer) readMessageData(client *clientConnection, buf []byte, msgLen uint32) error {
	n, err := client.conn.Read(buf[:msgLen])
	if err != nil {
		if s.IsRunning() && client.isConnected() {
			s.errors <- fmt.Errorf("player %d read data error: %w", client.playerID, err)
		}
		return err
	}
	// Record bytes received for metrics
	s.recordBytesReceived(uint64(n))
	return nil
}

func (client *clientConnection) updateLastActive() {
	client.mu.Lock()
	client.lastActive = time.Now()
	client.mu.Unlock()
}

func (s *TCPServer) decodeCommand(client *clientConnection, buf []byte, msgLen uint32) (*InputCommand, error) {
	cmd, err := s.protocol.DecodeInputCommand(buf[:msgLen])
	if err != nil {
		s.errors <- fmt.Errorf("player %d decode error: %w", client.playerID, err)
		return nil, err
	}
	return cmd, nil
}

func (s *TCPServer) sendCommandToGameLogic(cmd *InputCommand) {
	select {
	case s.inputCommands <- cmd:
		s.inputCommandStats.RecordSend()
	case <-s.done:
		return
	default:
		s.inputCommandStats.RecordDrop()
	}
}

// routeVoiceCommand forwards a voice InputCommand to all connected clients
// except the sender by wrapping the voice payload in a StateUpdate carrying a
// reserved VoiceComponentType component.
//
// The cmd.Data field is expected to begin with a single PacketTypeVoice byte
// (added by TCPVoiceTransport.SendVoice) followed by a serialized VoicePacket;
// the leading byte is stripped before wrapping so the receiving client only
// sees the bytes that DeserializeVoicePacket understands.
//
// senderID is the connection-bound player ID of the client that sent the
// packet; it is authoritative regardless of the (potentially spoofed or zero)
// cmd.PlayerID field. The packet is fanned out to every other connected
// client; channel-membership filtering is performed client-side by
// TCPVoiceTransport.HandleReceivedPacket (which checks IsInChannel for the
// embedded ChannelID). A future enhancement may add server-side membership
// filtering by querying engine.VoiceChannelSystem.
func (s *TCPServer) routeVoiceCommand(cmd *InputCommand, senderID uint64) {
	if cmd == nil || len(cmd.Data) < 1 {
		if s.logger != nil {
			s.logger.WithField("sender_id", senderID).Debug("voice routing: dropping empty voice command")
		}
		return
	}
	// Strip the leading packet-type byte. The remaining bytes are the
	// serialized VoicePacket as understood by DeserializeVoicePacket.
	payload := cmd.Data[1:]
	if len(payload) < VoicePacketHeaderSize {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"sender_id":   senderID,
				"payload_len": len(payload),
				"min_len":     VoicePacketHeaderSize,
			}).Warn("voice routing: dropping undersized voice payload")
		}
		return
	}

	// Wrap in a StateUpdate so the existing per-client priority queue,
	// batching, framing, and write paths handle delivery uniformly. EntityID
	// is set to the sender's player ID for diagnostic correlation; the
	// payload's own SenderID field is the authoritative source for the
	// receiving voice transport.
	update := &StateUpdate{
		EntityID: senderID,
		Priority: PriorityHigh,
		Components: []ComponentData{{
			Type: VoiceComponentType,
			Data: payload,
		}},
	}

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	// Assign sequence number once for the broadcast.
	s.stateMu.Lock()
	update.SequenceNumber = s.stateSeq
	s.stateSeq++
	s.stateMu.Unlock()

	for playerID, client := range s.clients {
		if playerID == senderID {
			continue
		}
		client.sendStateUpdate(update)
	}
}

// handleClientSend sends state updates to a client.
// Pops updates from the priority queue, sending higher priority updates first.
func (s *TCPServer) handleClientSend(client *clientConnection) {
	defer s.wg.Done()
	defer recovery.RecoverPanic(s.logger, "client send handler", nil)()

	for {
		select {
		case <-s.done:
			return

		case <-client.updateSignal:
			batchCount := s.processBatchUpdates(client)
			s.requeueIfMoreUpdates(client, batchCount)
		}
	}
}

// processBatchUpdates processes a batch of state updates for a client.
func (s *TCPServer) processBatchUpdates(client *clientConnection) int {
	const maxBatchSize = 20
	batchCount := 0

	for batchCount < maxBatchSize {
		update := client.stateUpdateQueue.Pop()
		if update == nil {
			break
		}
		batchCount++

		client.stateUpdateStats.RecordReceive()

		if !s.sendStateUpdate(client, update) {
			return batchCount
		}
	}

	return batchCount
}

// sendStateUpdate encodes and sends a single state update to a client.
func (s *TCPServer) sendStateUpdate(client *clientConnection, update *StateUpdate) bool {
	data, err := s.protocol.EncodeStateUpdate(update)
	if err != nil {
		s.errors <- fmt.Errorf("player %d encode error: %w", client.playerID, err)
		return true
	}

	lenBuf := s.createLengthPrefix(len(data))
	client.conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

	if !s.writeToClient(client, lenBuf) || !s.writeToClient(client, data) {
		return false
	}

	return true
}

// createLengthPrefix creates a 4-byte little-endian length prefix.
func (s *TCPServer) createLengthPrefix(length int) []byte {
	msgLen := uint32(length)
	return []byte{
		byte(msgLen),
		byte(msgLen >> 8),
		byte(msgLen >> 16),
		byte(msgLen >> 24),
	}
}

// writeToClient writes data to a client connection with error handling.
func (s *TCPServer) writeToClient(client *clientConnection, data []byte) bool {
	n, err := client.conn.Write(data)
	if err != nil {
		if s.IsRunning() && client.isConnected() {
			s.errors <- fmt.Errorf("player %d write error: %w", client.playerID, err)
		}
		return false
	}
	// Record bytes sent for metrics
	s.recordBytesSent(uint64(n))
	return true
}

// requeueIfMoreUpdates re-signals the client if more updates are available.
func (s *TCPServer) requeueIfMoreUpdates(client *clientConnection, batchCount int) {
	const maxBatchSize = 20
	if batchCount == maxBatchSize && !client.stateUpdateQueue.IsEmpty() {
		select {
		case client.updateSignal <- struct{}{}:
		default:
		}
	}
}

// disconnectClient removes a client from the server.
func (s *TCPServer) disconnectClient(playerID uint64) {
	client := s.removeClient(playerID)
	if client != nil {
		s.notifyPlayerLeave(playerID)
	}
}

// removeClient removes a client from the server's client map.
func (s *TCPServer) removeClient(playerID uint64) *clientConnection {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	client, exists := s.clients[playerID]
	if exists {
		client.disconnect()
		delete(s.clients, playerID)
		return client
	}
	return nil
}

// notifyPlayerLeave sends player leave notification to game logic.
func (s *TCPServer) notifyPlayerLeave(playerID uint64) {
	select {
	case s.playerLeaves <- playerID:
		s.playerLeaveStats.RecordSend()
	case <-s.done:
	default:
		s.handlePlayerLeaveChannelFull(playerID)
	}
}

// handlePlayerLeaveChannelFull handles the case when player leave channel is full.
func (s *TCPServer) handlePlayerLeaveChannelFull(playerID uint64) {
	s.playerLeaveStats.RecordDrop()
	s.reportPlayerLeaveError(playerID)
}

// reportPlayerLeaveError attempts to report a player leave event drop error.
func (s *TCPServer) reportPlayerLeaveError(playerID uint64) {
	err := fmt.Errorf("player leave channel full, dropped event for player %d", playerID)
	select {
	case s.errors <- err:
		s.errorStats.RecordSend()
	default:
		s.errorStats.RecordDrop()
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
			if err := c.conn.Close(); err != nil {
				logrus.WithError(err).Warn("error closing client connection")
			}
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

	// Push to priority queue with atomic stats recording.
	// Using PushWithCallback ensures the stats are recorded while still
	// holding the queue lock, preventing race conditions between the push
	// result and the stats update under high concurrency.
	c.stateUpdateQueue.PushWithCallback(update,
		func() {
			// onSuccess: record send and signal the send goroutine
			c.stateUpdateStats.RecordSend()

			// Signal the send goroutine (non-blocking)
			select {
			case c.updateSignal <- struct{}{}:
			default:
				// Channel buffer full, goroutine will process existing signal
			}
		},
		func() {
			// onFail: queue is full, record drop
			c.stateUpdateStats.RecordDrop()
		},
	)
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

// GetConnectedClients returns the number of currently connected clients.
func (s *TCPServer) GetConnectedClients() int {
	return s.GetPlayerCount()
}

// GetTotalBytesSent returns the total number of bytes sent to all clients.
func (s *TCPServer) GetTotalBytesSent() uint64 {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()
	return s.totalBytesSent
}

// GetTotalBytesReceived returns the total number of bytes received from all clients.
func (s *TCPServer) GetTotalBytesReceived() uint64 {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()
	return s.totalBytesRecv
}

// GetPacketsSent returns the total number of packets sent to all clients.
func (s *TCPServer) GetPacketsSent() uint64 {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()
	return s.totalPacketsSent
}

// GetPacketsReceived returns the total number of packets received from all clients.
func (s *TCPServer) GetPacketsReceived() uint64 {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()
	return s.totalPacketsRecv
}

// recordBytesSent increments the total bytes sent counter.
func (s *TCPServer) recordBytesSent(bytes uint64) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	s.totalBytesSent += bytes
	s.totalPacketsSent++
}

// recordBytesReceived increments the total bytes received counter.
func (s *TCPServer) recordBytesReceived(bytes uint64) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	s.totalBytesRecv += bytes
	s.totalPacketsRecv++
}

// Compile-time interface check
var _ ServerConnection = (*TCPServer)(nil)
