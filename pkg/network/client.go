// Package network provides multiplayer client functionality.
// This file implements TCPClient which handles client-side networking, prediction,
// and server communication for multiplayer gameplay over TCP connections.
package network

import (
	"fmt"
	"net"
	"sync"
	"time"
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

// TCPClient handles client-side networking over TCP.
// Implements ClientConnection interface.
type TCPClient struct {
	config   ClientConfig
	protocol Protocol

	// Connection state
	conn      net.Conn
	connected bool
	playerID  uint64

	// Sequence tracking
	inputSeq uint32
	stateSeq uint32

	// Channels for async communication
	stateUpdates chan *StateUpdate
	inputQueue   chan *InputCommand
	errors       chan error

	// Latency tracking
	latency  time.Duration
	lastPing time.Time
	lastPong time.Time

	// Thread safety
	mu sync.RWMutex

	// Shutdown
	done chan struct{}
	wg   sync.WaitGroup
}

// NewClient creates a new network client.
func NewClient(config ClientConfig) *TCPClient {
	return &TCPClient{
		config:       config,
		protocol:     NewBinaryProtocol(),
		stateUpdates: make(chan *StateUpdate, config.BufferSize),
		inputQueue:   make(chan *InputCommand, config.BufferSize),
		errors:       make(chan error, 16),
		done:         make(chan struct{}),
	}
}

// Connect establishes connection to the server.
func (c *TCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	// Set connection timeout
	conn, err := net.DialTimeout("tcp", c.config.ServerAddress, c.config.ConnectionTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.config.ServerAddress, err)
	}

	c.conn = conn
	c.connected = true
	c.lastPing = time.Now()
	c.lastPong = time.Now()

	// Start async handlers
	c.wg.Add(2)
	go c.receiveLoop()
	go c.sendLoop()

	return nil
}

// Disconnect closes the connection to the server.
func (c *TCPClient) Disconnect() error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return nil
	}

	c.connected = false
	close(c.done)

	// Close connection
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()

	// Wait for goroutines
	c.wg.Wait()

	return nil
}

// IsConnected returns whether the client is connected.
func (c *TCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
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
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return fmt.Errorf("not connected")
	}

	cmd := &InputCommand{
		PlayerID:       c.playerID,
		Timestamp:      uint64(time.Now().UnixNano()),
		SequenceNumber: c.inputSeq,
		InputType:      inputType,
		Data:           data,
	}
	c.inputSeq++
	c.mu.Unlock()

	select {
	case c.inputQueue <- cmd:
		return nil
	case <-c.done:
		return fmt.Errorf("client shutting down")
	default:
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
func (c *TCPClient) receiveLoop() {
	defer c.wg.Done()

	buf := make([]byte, 4096)
	for {
		select {
		case <-c.done:
			return
		default:
		}

		// Set read deadline
		c.conn.SetReadDeadline(time.Now().Add(c.config.ConnectionTimeout))

		// Read message length (4 bytes)
		if _, err := c.conn.Read(buf[:4]); err != nil {
			if c.IsConnected() {
				c.errors <- fmt.Errorf("read length error: %w", err)
			}
			return
		}

		// Decode length
		msgLen := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
		if msgLen > uint32(len(buf)) {
			c.errors <- fmt.Errorf("message too large: %d bytes", msgLen)
			return
		}

		// Read message data
		if _, err := c.conn.Read(buf[:msgLen]); err != nil {
			if c.IsConnected() {
				c.errors <- fmt.Errorf("read data error: %w", err)
			}
			return
		}

		// Decode state update
		update, err := c.protocol.DecodeStateUpdate(buf[:msgLen])
		if err != nil {
			c.errors <- fmt.Errorf("decode error: %w", err)
			continue
		}

		// Update sequence number
		c.mu.Lock()
		c.stateSeq = update.SequenceNumber
		c.mu.Unlock()

		// Send to channel (non-blocking)
		select {
		case c.stateUpdates <- update:
		case <-c.done:
			return
		default:
			// Drop if full (prioritize fresh updates)
		}
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
			// Send ping (empty input with type "ping")
			c.mu.Lock()
			c.lastPing = time.Now()
			c.mu.Unlock()

			// Update latency (time since last pong)
			c.mu.RLock()
			c.latency = time.Since(c.lastPong)
			c.mu.RUnlock()

		case cmd := <-c.inputQueue:
			// Encode input
			data, err := c.protocol.EncodeInputCommand(cmd)
			if err != nil {
				c.errors <- fmt.Errorf("encode error: %w", err)
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
			c.conn.SetWriteDeadline(time.Now().Add(c.config.ConnectionTimeout))

			// Send length + data
			if _, err := c.conn.Write(lenBuf); err != nil {
				if c.IsConnected() {
					c.errors <- fmt.Errorf("write length error: %w", err)
				}
				return
			}
			if _, err := c.conn.Write(data); err != nil {
				if c.IsConnected() {
					c.errors <- fmt.Errorf("write data error: %w", err)
				}
				return
			}
		}
	}
}

// Compile-time interface check
var _ ClientConnection = (*TCPClient)(nil)
