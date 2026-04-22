// Package webrtc signaling coordination.
// This file implements WebSocket-based signaling for WebRTC connection establishment,
// including both client (peer) and server (relay) components for SDP/ICE exchange.
package webrtc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/recovery"
)

// SignalingClient manages WebSocket connection to signaling server.
type SignalingClient struct {
	mu        sync.RWMutex
	url       string
	peerID    string
	connected bool

	// Message channels with bounded capacity for backpressure.
	// Senders use select with timeout to avoid blocking indefinitely
	// when the channel is full (see SendOffer, SendAnswer, etc.).
	sendChan  chan *SignalingMessage
	recvChan  chan *SignalingMessage
	closeChan chan struct{}

	// Peer registry (for signaling server)
	peers map[string]time.Time // peerID -> last seen

	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider
}

// DefaultSignalingChannelCapacity is the default capacity for signaling message channels.
// Senders apply backpressure via select-with-timeout when channels are full.
const DefaultSignalingChannelCapacity = 50

// NewSignalingClient creates a new signaling client.
func NewSignalingClient(url, peerID string) *SignalingClient {
	return NewSignalingClientWithCapacity(url, peerID, DefaultSignalingChannelCapacity)
}

// NewSignalingClientWithCapacity creates a new signaling client with configurable channel capacity.
// Higher capacity absorbs message bursts; lower capacity detects backpressure sooner.
func NewSignalingClientWithCapacity(url, peerID string, channelCapacity int) *SignalingClient {
	if channelCapacity <= 0 {
		channelCapacity = DefaultSignalingChannelCapacity
	}
	return &SignalingClient{
		url:          url,
		peerID:       peerID,
		sendChan:     make(chan *SignalingMessage, channelCapacity),
		recvChan:     make(chan *SignalingMessage, channelCapacity),
		closeChan:    make(chan struct{}),
		peers:        make(map[string]time.Time),
		timeProvider: DefaultTimeProvider(),
	}
}

// Connect establishes connection to the signaling server.
func (s *SignalingClient) Connect() error {
	s.mu.Lock()
	if s.connected {
		s.mu.Unlock()
		return nil
	}
	s.connected = true
	s.mu.Unlock()

	// In production, this would establish WebSocket connection
	// For testing, we simulate the connection
	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_signaling", "process messages", nil)()
		s.processMessages()
	}()

	return nil
}

// processMessages handles signaling message routing.
func (s *SignalingClient) processMessages() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.closeChan:
			return
		case msg := <-s.sendChan:
			// In production, send via WebSocket
			s.handleSend(msg)
		case <-ticker.C:
			// Clean up stale peers
			s.cleanupPeers()
		}
	}
}

// handleSend processes outgoing signaling messages.
func (s *SignalingClient) handleSend(msg *SignalingMessage) {
	// Simulate message relay to recipient
	// In production, this goes through WebSocket to server
	msg.Timestamp = s.timeProvider.Now()

	// Validate recipient exists
	s.mu.RLock()
	_, exists := s.peers[msg.To]
	s.mu.RUnlock()

	if !exists && msg.Type != "bye" {
		// Peer not registered, message dropped
		return
	}
}

// cleanupPeers removes peers that haven't been seen in 5 minutes.
func (s *SignalingClient) cleanupPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	threshold := s.timeProvider.Now().Add(-5 * time.Minute)
	for peerID, lastSeen := range s.peers {
		if lastSeen.Before(threshold) {
			delete(s.peers, peerID)
		}
	}
}

// SendOffer sends an SDP offer to a remote peer.
func (s *SignalingClient) SendOffer(remotePeerID string, offer *SDPOffer) error {
	s.mu.RLock()
	connected := s.connected
	s.mu.RUnlock()

	if !connected {
		return ErrSignalingNotConnected
	}

	msg := &SignalingMessage{
		Type:  "offer",
		From:  s.peerID,
		To:    remotePeerID,
		Offer: offer,
	}

	if !sendWithTimeout(s.sendChan, msg, 5*time.Second) {
		return fmt.Errorf("send offer timeout")
	}
	return nil
}

// SendAnswer sends an SDP answer to a remote peer.
func (s *SignalingClient) SendAnswer(remotePeerID string, answer *SDPAnswer) error {
	s.mu.RLock()
	connected := s.connected
	s.mu.RUnlock()

	if !connected {
		return ErrSignalingNotConnected
	}

	msg := &SignalingMessage{
		Type:   "answer",
		From:   s.peerID,
		To:     remotePeerID,
		Answer: answer,
	}

	if !sendWithTimeout(s.sendChan, msg, 5*time.Second) {
		return fmt.Errorf("send answer timeout")
	}
	return nil
}

// SendICECandidate sends an ICE candidate to a remote peer.
func (s *SignalingClient) SendICECandidate(remotePeerID string, candidate *ICECandidate) error {
	s.mu.RLock()
	connected := s.connected
	s.mu.RUnlock()

	if !connected {
		return ErrSignalingNotConnected
	}

	msg := &SignalingMessage{
		Type:      "candidate",
		From:      s.peerID,
		To:        remotePeerID,
		Candidate: candidate,
	}

	if !sendWithTimeout(s.sendChan, msg, 5*time.Second) {
		return fmt.Errorf("send candidate timeout")
	}
	return nil
}

// SendBye notifies a remote peer of disconnection.
func (s *SignalingClient) SendBye(remotePeerID string) error {
	s.mu.RLock()
	connected := s.connected
	s.mu.RUnlock()

	if !connected {
		return ErrSignalingNotConnected
	}

	msg := &SignalingMessage{
		Type: "bye",
		From: s.peerID,
		To:   remotePeerID,
	}

	if !sendWithTimeout(s.sendChan, msg, 5*time.Second) {
		return fmt.Errorf("send bye timeout")
	}
	return nil
}

// Receive returns channel for incoming signaling messages.
func (s *SignalingClient) Receive() <-chan *SignalingMessage {
	return s.recvChan
}

// RegisterPeer registers a peer as active on the signaling server.
func (s *SignalingClient) RegisterPeer(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[peerID] = s.timeProvider.Now()
	log.WithFields(log.Fields{
		"peer_id":     s.peerID,
		"register_id": peerID,
	}).Debug("peer registered on signaling server")
}

// UnregisterPeer removes a peer from the active registry.
func (s *SignalingClient) UnregisterPeer(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, peerID)
}

// GetActivePeers returns list of currently registered peers.
func (s *SignalingClient) GetActivePeers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peers := make([]string, 0, len(s.peers))
	for peerID := range s.peers {
		peers = append(peers, peerID)
	}
	return peers
}

// Close closes the signaling connection.
func (s *SignalingClient) Close() error {
	s.mu.Lock()
	if !s.connected {
		s.mu.Unlock()
		return nil
	}
	s.connected = false
	s.mu.Unlock()

	close(s.closeChan)
	return nil
}

// SignalingServer is a lightweight relay for WebRTC signaling.
// This runs as a separate process/server to coordinate initial connections.
type SignalingServer struct {
	mu      sync.RWMutex
	address string
	clients map[string]*SignalingClient

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSignalingServer creates a new signaling server.
func NewSignalingServer(address string) *SignalingServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &SignalingServer{
		address: address,
		clients: make(map[string]*SignalingClient),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the signaling server.
func (s *SignalingServer) Start() error {
	// In production, this would start HTTP/WebSocket server
	// For testing, we simulate server operations
	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_signaling_server", "run loop", nil)()
		s.run()
	}()
	return nil
}

// run is the main server loop.
func (s *SignalingServer) run() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupClients()
		}
	}
}

// cleanupClients removes inactive clients.
func (s *SignalingServer) cleanupClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove clients with no peers
	for clientID, client := range s.clients {
		if len(client.GetActivePeers()) == 0 {
			delete(s.clients, clientID)
		}
	}
}

// RegisterClient registers a client with the server.
func (s *SignalingServer) RegisterClient(clientID string) (*SignalingClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[clientID]; exists {
		return nil, fmt.Errorf("client %s already registered", clientID)
	}

	client := NewSignalingClient(s.address, clientID)
	s.clients[clientID] = client
	return client, nil
}

// RelayMessage relays a signaling message between clients.
func (s *SignalingServer) RelayMessage(msg *SignalingMessage) error {
	s.mu.RLock()
	toClient, exists := s.clients[msg.To]
	s.mu.RUnlock()

	if !exists {
		return ErrPeerNotFound
	}

	// Deliver message to recipient
	if !sendWithTimeout(toClient.recvChan, msg, 5*time.Second) {
		return fmt.Errorf("relay timeout")
	}
	return nil
}

// GetStats returns server statistics.
func (s *SignalingServer) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalPeers := 0
	for _, client := range s.clients {
		totalPeers += len(client.GetActivePeers())
	}

	return map[string]interface{}{
		"clients":     len(s.clients),
		"total_peers": totalPeers,
		"address":     s.address,
	}
}

// Stop stops the signaling server.
func (s *SignalingServer) Stop() error {
	s.cancel()

	s.mu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clients = make(map[string]*SignalingClient)
	s.mu.Unlock()

	return nil
}

// MarshalJSON serializes SignalingMessage to JSON.
func (m *SignalingMessage) MarshalJSON() ([]byte, error) {
	type Alias SignalingMessage
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(m),
		Timestamp: m.Timestamp.Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON deserializes SignalingMessage from JSON.
func (m *SignalingMessage) UnmarshalJSON(data []byte) error {
	type Alias SignalingMessage
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.Timestamp != "" {
		t, err := time.Parse(time.RFC3339Nano, aux.Timestamp)
		if err != nil {
			return err
		}
		m.Timestamp = t
	}

	return nil
}
