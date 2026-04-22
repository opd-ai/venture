// Package webrtc peer connection implementation.
// This file implements individual WebRTC peer connections with connection
// establishment, state management, and message send/receive operations.
// Note: This is a stub implementation for testing; real WebRTC integration
// requires github.com/pion/webrtc/v3.
package webrtc

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/recovery"
)

var errSendTimeout = errors.New("send timeout")

// NewPeer creates a new WebRTC federation peer with the given ID and configuration.
func NewPeer(id string, config *Config) (*Peer, error) {
	if id == "" {
		return nil, fmt.Errorf("peer ID cannot be empty")
	}
	if config == nil {
		config = DefaultConfig()
	}

	p := &Peer{
		ID:              id,
		Config:          config,
		state:           StateNew,
		sendChan:        make(chan []byte, 100),
		recvChan:        make(chan []byte, 100),
		closeChan:       make(chan struct{}),
		stateChangeChan: make(chan ConnectionState, 10),
		timeProvider:    DefaultTimeProvider(),
		stats: PeerStats{
			State: StateNew,
		},
	}

	return p, nil
}

// Connect establishes a connection to a remote peer.
// This initiates the WebRTC offer/answer exchange via the signaling server.
func (p *Peer) Connect(remotePeerID string) error {
	p.mu.Lock()
	if p.state != StateNew {
		currentState := p.state
		p.mu.Unlock()
		return fmt.Errorf("peer already in state %s", currentState)
	}
	p.remotePeerID = remotePeerID
	p.state = StateConnecting
	p.stats.State = StateConnecting
	p.mu.Unlock()

	// In a real implementation, this would:
	// 1. Create WebRTC PeerConnection
	// 2. Create data channel
	// 3. Generate SDP offer
	// 4. Send offer via signaling server
	// 5. Wait for answer and ICE candidates
	// 6. Establish P2P connection
	//
	// For testing, we simulate the connection process.

	ctx, cancel := context.WithTimeout(context.Background(), p.Config.ConnectionTimeout)
	defer cancel()

	// Simulate connection establishment
	go func(ctx context.Context) {
		defer recovery.RecoverPanicWithLogger("webrtc_peer", "simulate connection", nil)()
		p.simulateConnection(ctx)
	}(ctx)

	// Wait for connection or timeout
	for {
		select {
		case newState := <-p.stateChangeChan:
			if newState == StateConnected {
				p.mu.Lock()
				p.stats.ConnectedAt = p.timeProvider.Now()
				p.mu.Unlock()
				log.WithFields(log.Fields{
					"peer_id":        p.ID,
					"remote_peer_id": remotePeerID,
				}).Debug("peer connected")
				return nil
			}
			if newState == StateFailed {
				log.WithFields(log.Fields{
					"peer_id":        p.ID,
					"remote_peer_id": remotePeerID,
				}).Warn("peer connection failed")
				return fmt.Errorf("connection failed")
			}
			// Continue waiting for StateConnected
		case <-ctx.Done():
			p.mu.Lock()
			p.state = StateFailed
			p.stats.State = StateFailed
			p.mu.Unlock()
			return fmt.Errorf("connection timeout: %w", ctx.Err())
		}
	}
}

// simulateConnection simulates the WebRTC connection process for testing.
// In production, this would be replaced with actual WebRTC negotiation.
func (p *Peer) simulateConnection(ctx context.Context) {
	// Simulate minimal connection delay for testing
	// In production, signaling (100-500ms) + ICE gathering (500-2000ms)
	time.Sleep(10 * time.Millisecond)

	// Check for cancellation
	select {
	case <-ctx.Done():
		p.stateChangeChan <- StateFailed
		return
	default:
	}

	// Connection successful
	p.mu.Lock()
	p.state = StateConnected
	p.stats.State = StateConnected
	p.mu.Unlock()

	// Notify state change
	p.stateChangeChan <- StateConnected

	// Start message processing
	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_peer", "process messages", nil)()
		p.processMessages()
	}()
}

// processMessages handles sending and receiving messages on the data channel.
func (p *Peer) processMessages() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.closeChan:
			return
		case <-p.sendChan:
			// In production, this would send via WebRTC data channel
			// Stats are updated in Send() method
		case <-ticker.C:
			// Periodic housekeeping
		}
	}
}

// Send sends a message to the connected remote peer.
func (p *Peer) Send(data []byte) error {
	p.mu.RLock()
	state := p.state
	maxSize := p.Config.MaxMessageSize
	p.mu.RUnlock()

	if state != StateConnected {
		return ErrNotConnected
	}

	if len(data) > maxSize {
		return ErrMessageTooLarge
	}

	// Update stats immediately for testing (in production, this happens in processMessages)
	p.mu.Lock()
	p.stats.BytesSent += uint64(len(data))
	p.stats.MessagesSent++
	p.stats.LastActivity = p.timeProvider.Now()
	p.mu.Unlock()

	return sendWithTimeoutOrDone(
		p.sendChan,
		data,
		p.closeChan,
		5*time.Second,
		ErrConnectionClosed,
		errSendTimeout,
	)
}

// Receive returns a channel for receiving messages from the remote peer.
func (p *Peer) Receive() <-chan []byte {
	return p.recvChan
}

// GetState returns the current connection state.
func (p *Peer) GetState() ConnectionState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetStats returns current connection statistics.
func (p *Peer) GetStats() PeerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// GetRemotePeerID returns the ID of the connected remote peer.
func (p *Peer) GetRemotePeerID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.remotePeerID
}

// setState updates the connection state and notifies listeners.
func (p *Peer) setState(state ConnectionState) {
	p.mu.Lock()
	oldState := p.state
	p.state = state
	p.stats.State = state
	p.mu.Unlock()

	if oldState != state {
		// Non-blocking state change notification
		select {
		case p.stateChangeChan <- state:
		default:
		}
	}
}

// Close gracefully closes the peer connection.
func (p *Peer) Close() error {
	p.mu.Lock()
	if p.state == StateClosed {
		p.mu.Unlock()
		return nil
	}
	wasConnected := p.state == StateConnected || p.state == StateConnecting
	p.state = StateClosed
	p.stats.State = StateClosed
	p.mu.Unlock()

	log.WithFields(log.Fields{
		"peer_id":       p.ID,
		"was_connected": wasConnected,
	}).Debug("peer connection closed")

	// Only close closeChan if processMessages goroutine is running
	if wasConnected {
		close(p.closeChan)
	}
	return nil
}
