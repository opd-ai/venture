// Package webrtc core types and data structures.
// This file defines all core types, structs, and configurations used throughout
// the webrtc package, including peer connection state, SDP structures, and statistics.
package webrtc

import (
	"sync"
	"time"
)

// ConnectionState represents the current state of a WebRTC peer connection.
type ConnectionState int

// String returns the string representation of ConnectionState.
func (s ConnectionState) String() string {
	switch s {
	case StateNew:
		return "New"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateDisconnected:
		return "Disconnected"
	case StateFailed:
		return "Failed"
	case StateClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// ICECandidateType represents the type of ICE candidate.
type ICECandidateType int

// String returns the string representation of ICECandidateType.
func (t ICECandidateType) String() string {
	switch t {
	case CandidateHost:
		return "Host"
	case CandidateServerReflexive:
		return "ServerReflexive"
	case CandidateRelay:
		return "Relay"
	default:
		return "Unknown"
	}
}

// TURNServer represents a TURN relay server configuration.
type TURNServer struct {
	// URLs are the TURN server URLs (turn:host:port or turns:host:port).
	URLs []string
	// Username for TURN authentication.
	Username string
	// Credential for TURN authentication.
	Credential string
}

// Config holds configuration for WebRTC federation.
type Config struct {
	// SignalingURL is the WebSocket URL for the signaling server (optional).
	// If empty, signaling must be done manually via SetRemoteOffer/SetRemoteAnswer.
	SignalingURL string

	// STUNServers are STUN server URLs for NAT traversal.
	// Default: ["stun:stun.l.google.com:19302"]
	STUNServers []string

	// TURNServers are TURN relay servers for symmetric NAT (optional).
	TURNServers []TURNServer

	// DataChannelLabel is the label for the federation data channel.
	// Default: "venture-federation"
	DataChannelLabel string

	// ICETimeout is the timeout for ICE candidate gathering.
	// Default: 10 seconds
	ICETimeout time.Duration

	// ConnectionTimeout is the timeout for establishing a connection.
	// Default: 30 seconds
	ConnectionTimeout time.Duration

	// ReconnectAttempts is the number of reconnection attempts.
	// Default: 3
	ReconnectAttempts int

	// ReconnectDelay is the delay between reconnection attempts.
	// Default: 5 seconds
	ReconnectDelay time.Duration

	// MaxMessageSize is the maximum size of a federation message.
	// Default: 1 MB
	MaxMessageSize int

	// EnableFallback enables automatic fallback to WebSocket.
	// Default: true
	EnableFallback bool

	// FallbackURL is the WebSocket URL for fallback (optional).
	FallbackURL string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		STUNServers:       []string{"stun:stun.l.google.com:19302"},
		TURNServers:       []TURNServer{},
		DataChannelLabel:  "venture-federation",
		ICETimeout:        10 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		ReconnectAttempts: 3,
		ReconnectDelay:    5 * time.Second,
		MaxMessageSize:    1024 * 1024, // 1 MB
		EnableFallback:    true,
	}
}

// SDPOffer represents a WebRTC session description (offer).
type SDPOffer struct {
	// Type is always "offer".
	Type string
	// SDP is the session description protocol string.
	SDP string
}

// SDPAnswer represents a WebRTC session description (answer).
type SDPAnswer struct {
	// Type is always "answer".
	Type string
	// SDP is the session description protocol string.
	SDP string
}

// ICECandidate represents an ICE candidate for NAT traversal.
type ICECandidate struct {
	// Candidate is the ICE candidate string.
	Candidate string
	// SDPMid is the media stream ID.
	SDPMid string
	// SDPMLineIndex is the m-line index.
	SDPMLineIndex int
}

// SignalingMessage represents a message exchanged via the signaling server.
type SignalingMessage struct {
	// Type is the message type: "offer", "answer", "candidate", "bye".
	Type string
	// From is the sender peer ID.
	From string
	// To is the recipient peer ID.
	To string
	// Offer is the SDP offer (for type="offer").
	Offer *SDPOffer `json:"offer,omitempty"`
	// Answer is the SDP answer (for type="answer").
	Answer *SDPAnswer `json:"answer,omitempty"`
	// Candidate is the ICE candidate (for type="candidate").
	Candidate *ICECandidate `json:"candidate,omitempty"`
	// Timestamp is when the message was created.
	Timestamp time.Time
}

// PeerStats holds statistics about a peer connection.
type PeerStats struct {
	// State is the current connection state.
	State ConnectionState
	// BytesSent is the total bytes sent on the data channel.
	BytesSent uint64
	// BytesReceived is the total bytes received on the data channel.
	BytesReceived uint64
	// MessagesSent is the total messages sent.
	MessagesSent uint64
	// MessagesReceived is the total messages received.
	MessagesReceived uint64
	// RTT is the estimated round-trip time in milliseconds.
	RTT time.Duration
	// ConnectedAt is when the connection was established.
	ConnectedAt time.Time
	// LastActivity is when the last message was sent/received.
	LastActivity time.Time
	// ICECandidatesUsed is the number of ICE candidates used.
	ICECandidatesUsed int
	// UsingTURN indicates if a TURN relay is being used.
	UsingTURN bool
}

// Peer represents a WebRTC federation peer connection.
type Peer struct {
	// ID is the unique identifier for this peer.
	ID string

	// Config holds the peer configuration.
	Config *Config

	// mu protects peer state.
	mu sync.RWMutex

	// state is the current connection state.
	state ConnectionState

	// remotePeerID is the ID of the connected remote peer.
	remotePeerID string

	// stats holds connection statistics.
	stats PeerStats

	// sendChan is the channel for outgoing messages.
	sendChan chan []byte

	// recvChan is the channel for incoming messages.
	recvChan chan []byte

	// closeChan signals the peer is closing.
	closeChan chan struct{}

	// stateChangeChan signals state changes.
	stateChangeChan chan ConnectionState

	// timeProvider abstracts time access for deterministic testing.
	timeProvider TimeProvider

	// peerConn holds the underlying WebRTC peer connection.
	// Value is *webrtc.PeerConnection (WASM builds via pion) or nil (native builds).
	peerConn interface{}

	// dataChannel holds the WebRTC data channel used for game federation messages.
	// Value is *webrtc.DataChannel (WASM builds via pion) or nil (native builds).
	dataChannel interface{}
}

// webrtcCloseable is implemented by underlying WebRTC handles that require
// explicit shutdown, such as peer connections and data channels.
type webrtcCloseable interface {
	Close() error
}

// closeTransportHandles releases any stored underlying WebRTC transport handles.
// It type-asserts peerConn/dataChannel to webrtcCloseable under the mutex and
// nils the fields so repeated shutdown paths are idempotent.
func (p *Peer) closeTransportHandles() {
	p.mu.Lock()
	dc := p.dataChannel
	pc := p.peerConn
	p.dataChannel = nil
	p.peerConn = nil
	p.mu.Unlock()

	if c, ok := dc.(webrtcCloseable); ok && c != nil {
		_ = c.Close()
	}
	if c, ok := pc.(webrtcCloseable); ok && c != nil {
		_ = c.Close()
	}
}

// ConnectionMetrics holds aggregated metrics for all peer connections.
type ConnectionMetrics struct {
	// TotalConnections is the total number of connections attempted.
	TotalConnections int
	// ActiveConnections is the number of currently active connections.
	ActiveConnections int
	// FailedConnections is the number of failed connection attempts.
	FailedConnections int
	// TotalBytesSent is the total bytes sent across all connections.
	TotalBytesSent uint64
	// TotalBytesReceived is the total bytes received across all connections.
	TotalBytesReceived uint64
	// AverageRTT is the average RTT across active connections.
	AverageRTT time.Duration
	// TURNUsageRate is the percentage of connections using TURN relay.
	TURNUsageRate float64
}
