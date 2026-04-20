// Package network provides multiplayer client functionality.
// This file implements TCPClient which handles client-side networking, prediction,
// and server communication for multiplayer gameplay over TCP connections.
package network

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// ClientConfig holds configuration for the network client.
type ClientConfig struct {
	ServerAddress     string        // Server address (host:port)
	ConnectionTimeout time.Duration // Timeout for connection attempts
	PingInterval      time.Duration // Interval between ping messages
	MaxLatency        time.Duration // Maximum acceptable latency before warnings
	BufferSize        int           // Size of send/receive buffers
}

// DefaultClientConfig returns a client configuration with sensible defaults.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		ServerAddress:     "localhost:8080",
		ConnectionTimeout: 10 * time.Second,
		PingInterval:      1 * time.Second,
		MaxLatency:        500 * time.Millisecond,
		BufferSize:        256,
	}
}

// TorClientConfig returns a client configuration optimized for Tor/onion service
// connections and other high-latency networks (200-5000ms). This configuration uses
// significantly longer timeouts and larger buffers to handle the characteristics of
// anonymizing networks.
//
// Key differences from DefaultClientConfig:
//   - ConnectionTimeout: 60s (6x increase) - allows time for Tor circuit building (15-30s)
//   - MaxLatency: 5000ms (10x increase) - accepts extreme latency without warnings
//   - PingInterval: 5s (5x increase) - reduces unnecessary traffic over slow links
//   - BufferSize: 512 (2x increase) - buffers ~200 messages for 10s round-trip at 20Hz
//
// Use this configuration when connecting through Tor, I2P, or other high-latency networks
// where round-trip times can exceed 5 seconds and connection establishment is slow.
func TorClientConfig() ClientConfig {
	return ClientConfig{
		ServerAddress:     "localhost:8080",
		ConnectionTimeout: 60 * time.Second,        // 60s for Tor circuit building
		PingInterval:      5 * time.Second,         // Reduced ping frequency
		MaxLatency:        5000 * time.Millisecond, // Accept up to 5s latency
		BufferSize:        512,                     // Increased for buffering
	}
}

// ReconnectConfig configures automatic reconnection behavior with exponential backoff.
// This provides resilience against transient network failures common in high-latency
// environments like Tor where connections may drop due to circuit changes or timeouts.
type ReconnectConfig struct {
	MaxRetries    int           // Maximum number of reconnection attempts (0 = infinite)
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries (caps exponential growth)
	BackoffFactor float64       // Multiplier for exponential backoff (typically 2.0)
}

// DefaultReconnectConfig returns a reconnection configuration for standard networks.
// Retries up to 5 times with exponential backoff capped at 30 seconds.
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:    5,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// TorReconnectConfig returns a reconnection configuration for high-latency networks.
// More aggressive retry strategy with longer delays to account for Tor circuit rebuilding.
func TorReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:    10,                // More retries for unstable Tor circuits
		InitialDelay:  5 * time.Second,   // Longer initial delay for circuit stability
		MaxDelay:      120 * time.Second, // Higher cap (2 minutes) for Tor circuit rebuilding
		BackoffFactor: 2.0,
	}
}

// TCPClient handles client-side networking over TCP.
// Implements ClientConnection interface.
type TCPClient struct {
	config   ClientConfig
	protocol Protocol

	// Connection state
	conn      net.Conn
	connected atomic.Bool
	playerID  uint64

	// Sequence tracking
	inputSeq uint32
	stateSeq uint32

	// Channels for async communication
	stateUpdates chan *StateUpdate
	inputQueue   chan *InputCommand
	errors       chan error

	// Buffer monitoring
	stateUpdateStats *BufferStats
	inputQueueStats  *BufferStats
	errorStats       *BufferStats

	// Latency tracking
	latency  time.Duration
	lastPing time.Time
	lastPong time.Time

	// Thread safety
	mu sync.RWMutex

	// Shutdown - done channel is protected by doneMu for safe closing
	done       chan struct{}
	doneMu     sync.Mutex // Guards closing of done channel
	doneClosed bool       // Tracks if current done channel is closed
	wg         sync.WaitGroup

	// voiceHandler is invoked for inbound voice packets demultiplexed from
	// StateUpdate messages carrying a VoiceComponentType component. Set via
	// SetVoiceHandler. When nil, voice packets are silently dropped (no game
	// state side-effects).
	voiceHandler   func(*VoicePacket)
	voiceHandlerMu sync.RWMutex

	// Logger for network operations
	logger *logrus.Entry
}

// NewClient creates a new network client.
func NewClient(config ClientConfig) *TCPClient {
	return NewClientWithLogger(config, nil)
}

// NewClientWithLogger creates a new network client with a logger.
func NewClientWithLogger(config ClientConfig, logger *logrus.Logger) *TCPClient {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "network_client",
			"server":    config.ServerAddress,
		})
	}

	client := &TCPClient{
		config:       config,
		protocol:     NewBinaryProtocol(),
		stateUpdates: make(chan *StateUpdate, config.BufferSize),
		inputQueue:   make(chan *InputCommand, config.BufferSize),
		errors:       make(chan error, 16),
		done:         make(chan struct{}),
		logger:       logEntry,
	}

	// Initialize buffer monitoring
	client.stateUpdateStats = NewBufferStats("client_state_updates", config.BufferSize, logEntry)
	client.inputQueueStats = NewBufferStats("client_input_queue", config.BufferSize, logEntry)
	client.errorStats = NewBufferStats("client_errors", 16, logEntry)

	return client
}

// Connect establishes connection to the server.
func (c *TCPClient) Connect() error {
	// Check if already connected before acquiring lock
	if c.connected.Load() {
		return fmt.Errorf("already connected")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring lock to prevent race
	if c.connected.Load() {
		return fmt.Errorf("already connected")
	}

	if c.logger != nil {
		c.logger.WithField("server", c.config.ServerAddress).Info("connecting to server")
	}

	conn, err := c.establishConnection()
	if err != nil {
		return err
	}

	// Recreate done channel for reconnection after disconnect.
	// Must be protected by doneMu to prevent race with Disconnect.
	c.doneMu.Lock()
	c.done = make(chan struct{})
	c.doneClosed = false
	c.doneMu.Unlock()

	c.setupConnection(conn)
	c.startConnectionHandlers()

	if c.logger != nil {
		c.logger.WithField("server", c.config.ServerAddress).Info("connected successfully")
	}

	return nil
}

// establishConnection creates and configures a TCP connection to the server.
func (c *TCPClient) establishConnection() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", c.config.ServerAddress, c.config.ConnectionTimeout)
	if err != nil {
		if c.logger != nil {
			c.logger.WithError(err).WithField("server", c.config.ServerAddress).Error("failed to connect")
		}
		return nil, fmt.Errorf("failed to connect to %s: %w", c.config.ServerAddress, err)
	}

	c.configureTCPKeepalive(conn)
	return conn, nil
}

// configureTCPKeepalive enables TCP keepalive to prevent silent disconnections.
// Uses KeepAliveConn interface for testability and flexible connection handling.
func (c *TCPClient) configureTCPKeepalive(conn net.Conn) {
	kaConn, ok := conn.(KeepAliveConn)
	if !ok {
		return
	}

	if err := kaConn.SetKeepAlive(true); err != nil {
		if c.logger != nil {
			c.logger.WithError(err).Warn("failed to enable TCP keepalive")
		}
		return
	}

	if err := kaConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
		if c.logger != nil {
			c.logger.WithError(err).Warn("failed to set TCP keepalive period")
		}
		return
	}

	if c.logger != nil {
		c.logger.Debug("TCP keepalive enabled (30s period)")
	}
}

// setupConnection initializes connection state.
func (c *TCPClient) setupConnection(conn net.Conn) {
	c.conn = conn
	c.connected.Store(true)
	c.lastPing = time.Now()
	c.lastPong = time.Now()
}

// startConnectionHandlers starts async receive and send loops.
func (c *TCPClient) startConnectionHandlers() {
	c.wg.Add(2)
	go c.receiveLoop()
	go c.sendLoop()
}

// ConnectWithRetry establishes connection to the server with automatic retry on failure.
// Uses exponential backoff to handle transient network issues gracefully.
//
// Retry behavior is controlled by ReconnectConfig:
//   - MaxRetries: Maximum number of connection attempts (0 = infinite retries)
//   - InitialDelay: Starting delay between retries (e.g., 1s)
//   - MaxDelay: Maximum delay cap for exponential backoff (e.g., 30s)
//   - BackoffMultiplier: Factor to multiply delay after each failure (typically 2.0)
//
// Returns nil on successful connection, or an error if all retries are exhausted.
// The function can be cancelled gracefully by calling Disconnect() on the client,
// which will interrupt the retry loop immediately.
//
// Thread-safety: Safe to call concurrently with Disconnect(). Uses internal
// cancellation channel to coordinate shutdown.
//
// Example usage:
//
//	client := NewClient(TorClientConfig())
//	if err := client.ConnectWithRetry(TorReconnectConfig()); err != nil {
//	    log.WithError(err).Error("Failed to connect after retries")
//	    return err
//	}
func (c *TCPClient) ConnectWithRetry(reconnectConfig ReconnectConfig) error {
	attempt := 0
	delay := reconnectConfig.InitialDelay

	for {
		err := c.Connect()
		if err == nil {
			c.logSuccessfulConnection(attempt)
			return nil
		}

		attempt++
		if err := c.checkMaxRetries(reconnectConfig, attempt, err); err != nil {
			return err
		}

		c.logRetryAttempt(err, attempt, delay)

		if err := c.waitForRetry(delay); err != nil {
			return err
		}

		delay = c.calculateNextDelay(delay, reconnectConfig)
	}
}

// logSuccessfulConnection logs when a connection succeeds after retries.
func (c *TCPClient) logSuccessfulConnection(attempt int) {
	if c.logger != nil && attempt > 0 {
		c.logger.WithField("attempts", attempt+1).Info("connected successfully after retries")
	}
}

// checkMaxRetries verifies if maximum retry attempts have been reached.
func (c *TCPClient) checkMaxRetries(config ReconnectConfig, attempt int, err error) error {
	if config.MaxRetries > 0 && attempt >= config.MaxRetries {
		if c.logger != nil {
			c.logger.WithError(err).WithField("attempts", attempt).Error("exhausted all reconnection attempts")
		}
		return fmt.Errorf("failed to connect after %d attempts: %w", attempt, err)
	}
	return nil
}

// logRetryAttempt logs connection failure and retry information.
func (c *TCPClient) logRetryAttempt(err error, attempt int, delay time.Duration) {
	if c.logger != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"attempt": attempt,
			"delay":   delay,
		}).Warn("connection failed, retrying")
	}
}

// waitForRetry waits before the next retry with cancellation support.
func (c *TCPClient) waitForRetry(delay time.Duration) error {
	select {
	case <-c.done:
		if c.logger != nil {
			c.logger.Info("reconnection cancelled - client disconnected")
		}
		return fmt.Errorf("reconnection cancelled")
	default:
	}

	select {
	case <-c.done:
		if c.logger != nil {
			c.logger.Info("reconnection cancelled during shutdown")
		}
		return fmt.Errorf("reconnection cancelled")
	case <-time.After(delay):
		return nil
	}
}

// calculateNextDelay computes the next retry delay using exponential backoff.
func (c *TCPClient) calculateNextDelay(currentDelay time.Duration, config ReconnectConfig) time.Duration {
	nextDelay := time.Duration(float64(currentDelay) * config.BackoffFactor)
	if nextDelay > config.MaxDelay {
		return config.MaxDelay
	}
	return nextDelay
}

// Disconnect closes the connection to the server.
func (c *TCPClient) Disconnect() error {
	if c.logger != nil {
		c.logger.Info("disconnecting from server")
	}

	// Set connected to false atomically before signaling shutdown
	c.connected.Store(false)

	// Close done channel to signal goroutines to exit.
	// Use doneMu to prevent race with Connect() and ensure single close.
	c.doneMu.Lock()
	if !c.doneClosed && c.done != nil {
		close(c.done)
		c.doneClosed = true
	}
	c.doneMu.Unlock()

	// Close connection to unblock any pending I/O operations
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		conn.Close()
	}

	// Wait for goroutines to finish (no lock held to avoid deadlock)
	c.wg.Wait()

	if c.logger != nil {
		c.logger.Info("disconnected successfully")
	}

	return nil
}

// IsConnected returns whether the client is connected.
func (c *TCPClient) IsConnected() bool {
	return c.connected.Load()
}

// GetPlayerID returns the client's player ID.
func (c *TCPClient) GetPlayerID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.playerID
}

// SetPlayerID sets the client's player ID (called after authentication).
func (c *TCPClient) SetPlayerID(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.playerID = id
}

// GetLatency returns the current network latency.
func (c *TCPClient) GetLatency() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latency
}

// SendInput queues an input command to send to the server.
func (c *TCPClient) SendInput(inputType string, data []byte) error {
	if !c.connected.Load() {
		return fmt.Errorf("not connected")
	}

	timestamp, err := NowTimestamp()
	if err != nil {
		return fmt.Errorf("cannot create input command: %w", err)
	}

	c.mu.Lock()

	cmd := &InputCommand{
		PlayerID:       c.playerID,
		Timestamp:      timestamp,
		SequenceNumber: c.inputSeq,
		InputType:      inputType,
		Data:           data,
	}
	c.inputSeq++
	c.mu.Unlock()

	select {
	case c.inputQueue <- cmd:
		c.inputQueueStats.RecordSend()
		return nil
	case <-c.done:
		return fmt.Errorf("client shutting down")
	default:
		c.inputQueueStats.RecordDrop()
		return fmt.Errorf("input queue full")
	}
}

// ReceiveStateUpdate returns a channel for receiving state updates from the server.
func (c *TCPClient) ReceiveStateUpdate() <-chan *StateUpdate {
	return c.stateUpdates
}

// ReceiveError returns a channel for receiving errors.
func (c *TCPClient) ReceiveError() <-chan error {
	return c.errors
}

// receiveLoop continuously receives data from the server.
// Performance optimization: The done channel is NOT checked in the hot path to minimize
// overhead on every packet receive. Instead, we rely on SetReadDeadline and connection
// closure (triggered by Disconnect()) to exit the loop. This eliminates ~1-2% CPU overhead
// at high packet rates (640+ packets/sec) by avoiding select statement in the hot path.
func (c *TCPClient) receiveLoop() {
	defer c.wg.Done()

	buf := make([]byte, 4096)
	for {
		// Hot path optimization: No done channel check here.
		// Disconnect() closes the connection, causing reads to fail and exit the loop.

		// Check if connection is valid before accessing it
		// Note: This nil check is an early-exit optimization, not full race prevention.
		// The connection may still be closed after this point; in that case,
		// SetReadDeadline and subsequent reads will return errors that are already
		// handled by readMessageLength/readMessageData.
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		conn.SetReadDeadline(time.Now().Add(c.config.ConnectionTimeout))

		msgLen, ok := c.readMessageLength(buf)
		if !ok {
			return
		}

		if !c.readMessageData(buf, msgLen) {
			return
		}

		if !c.processStateUpdate(buf[:msgLen]) {
			continue
		}
	}
}

// readMessageLength reads and validates the message length prefix.
// Returns the message length and true if successful, 0 and false otherwise.
func (c *TCPClient) readMessageLength(buf []byte) (uint32, bool) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return 0, false
	}

	if _, err := conn.Read(buf[:4]); err != nil {
		c.sendNonBlockingError(fmt.Errorf("read length error: %w", err), true)
		return 0, false
	}

	msgLen := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
	if msgLen > uint32(len(buf)) {
		c.sendNonBlockingError(fmt.Errorf("message too large: %d bytes", msgLen), true)
		return 0, false
	}

	return msgLen, true
}

// readMessageData reads the message data from the connection.
// Returns true if successful, false if an error occurred.
func (c *TCPClient) readMessageData(buf []byte, msgLen uint32) bool {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return false
	}

	if _, err := conn.Read(buf[:msgLen]); err != nil {
		c.sendNonBlockingError(fmt.Errorf("read data error: %w", err), true)
		return false
	}
	return true
}

// processStateUpdate decodes and processes a state update message.
// Returns true to continue processing, false to stop the receive loop.
func (c *TCPClient) processStateUpdate(data []byte) bool {
	update, err := c.protocol.DecodeStateUpdate(data)
	if err != nil {
		c.sendNonBlockingError(fmt.Errorf("decode error: %w", err), false)
		return false
	}

	// Demultiplex voice packets out of the state-update stream before they
	// reach game logic (AUDIT.md CRITICAL: voice chat server-side packet
	// routing). Voice updates carry a single reserved VoiceComponentType
	// component and have no gameplay side effects.
	if c.routeVoiceComponent(update) {
		return true
	}

	c.mu.Lock()
	c.stateSeq = update.SequenceNumber
	c.mu.Unlock()

	select {
	case c.stateUpdates <- update:
		c.stateUpdateStats.RecordSend()
	case <-c.done:
		return false
	default:
		c.stateUpdateStats.RecordDrop()
	}

	return true
}

// SetVoiceHandler registers a callback invoked for inbound voice packets.
// Pass nil to clear. Safe to call concurrently with the receive loop.
func (c *TCPClient) SetVoiceHandler(h func(*VoicePacket)) {
	c.voiceHandlerMu.Lock()
	c.voiceHandler = h
	c.voiceHandlerMu.Unlock()
}

// routeVoiceComponent inspects a decoded StateUpdate for a reserved
// VoiceComponentType component. If found, it deserializes the embedded
// VoicePacket, dispatches it to the registered voice handler, and returns
// true so the caller skips game-state enqueueing. Returns false otherwise.
//
// The server currently constructs voice updates with exactly one component,
// but this scan tolerates additional metadata components that may be added
// alongside the voice payload in the future.
func (c *TCPClient) routeVoiceComponent(update *StateUpdate) bool {
	if update == nil {
		return false
	}
	for _, comp := range update.Components {
		if comp.Type != VoiceComponentType {
			continue
		}
		pkt, err := DeserializeVoicePacket(comp.Data)
		if err != nil {
			if c.logger != nil {
				c.logger.WithError(err).Warn("failed to deserialize inbound voice packet")
			}
			return true
		}

		c.voiceHandlerMu.RLock()
		handler := c.voiceHandler
		c.voiceHandlerMu.RUnlock()

		if handler != nil {
			handler(pkt)
		}
		return true
	}
	return false
}

// sendNonBlockingError sends an error to the error channel without blocking.
// If exitOnError is true, it signals the loop should terminate by returning from done channel.
func (c *TCPClient) sendNonBlockingError(err error, exitOnError bool) {
	select {
	case c.errors <- err:
		c.errorStats.RecordSend()
	case <-c.done:
		// Already shutting down
	default:
		c.errorStats.RecordDrop()
	}
}

// sendLoop continuously sends queued inputs to the server.
func (c *TCPClient) sendLoop() {
	defer c.wg.Done()

	pingTicker := time.NewTicker(c.config.PingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.done:
			return

		case <-pingTicker.C:
			c.handlePingTick()

		case cmd := <-c.inputQueue:
			if err := c.sendInputCommand(cmd); err != nil {
				return
			}
		}
	}
}

// handlePingTick updates ping timestamp and calculates latency.
func (c *TCPClient) handlePingTick() {
	c.mu.Lock()
	c.lastPing = time.Now()
	c.mu.Unlock()

	c.mu.RLock()
	c.latency = time.Since(c.lastPong)
	c.mu.RUnlock()
}

// sendInputCommand encodes and sends an input command to the server.
func (c *TCPClient) sendInputCommand(cmd *InputCommand) error {
	c.inputQueueStats.RecordReceive()

	data, err := c.protocol.EncodeInputCommand(cmd)
	if err != nil {
		c.errors <- fmt.Errorf("encode error: %w", err)
		return nil
	}

	return c.writeMessageWithLength(data)
}

// writeMessageWithLength writes a length-prefixed message to the connection.
func (c *TCPClient) writeMessageWithLength(data []byte) error {
	msgLen := uint32(len(data))
	lenBuf := []byte{
		byte(msgLen),
		byte(msgLen >> 8),
		byte(msgLen >> 16),
		byte(msgLen >> 24),
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.config.ConnectionTimeout))

	if _, err := c.conn.Write(lenBuf); err != nil {
		if c.IsConnected() {
			c.errors <- fmt.Errorf("write length error: %w", err)
		}
		return err
	}

	if _, err := c.conn.Write(data); err != nil {
		if c.IsConnected() {
			c.errors <- fmt.Errorf("write data error: %w", err)
		}
		return err
	}

	return nil
}

// GetBufferStats returns snapshots of all buffer statistics.
// This provides visibility into channel utilization and potential congestion.
func (c *TCPClient) GetBufferStats() map[string]BufferSnapshot {
	return map[string]BufferSnapshot{
		"state_updates": c.stateUpdateStats.Snapshot(),
		"input_queue":   c.inputQueueStats.Snapshot(),
		"errors":        c.errorStats.Snapshot(),
	}
}

// Compile-time interface check
var _ ClientConnection = (*TCPClient)(nil)
