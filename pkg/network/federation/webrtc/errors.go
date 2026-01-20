// Package webrtc errors consolidation.
// This file contains all error variables used throughout the webrtc package,
// consolidated from peer.go, nat_traversal.go, relay.go, signaling.go, and stun.go
// for easier error discovery and maintenance.
package webrtc

import "errors"

// Error variables for WebRTC federation.
// Code relocated from: peer.go, nat_traversal.go, relay.go, signaling.go, stun.go

// Peer errors
// Originally defined in: peer.go
var (
	// ErrNotConnected indicates the peer is not connected.
	ErrNotConnected = errors.New("peer not connected")
	// ErrConnectionClosed indicates the connection was closed.
	ErrConnectionClosed = errors.New("connection closed")
	// ErrConnectionFailed indicates the connection failed.
	ErrConnectionFailed = errors.New("connection failed")
	// ErrSignalingFailed indicates signaling failed.
	ErrSignalingFailed = errors.New("signaling failed")
	// ErrICETimeout indicates ICE negotiation timed out.
	ErrICETimeout = errors.New("ICE negotiation timeout")
	// ErrMessageTooLarge indicates a message exceeds MaxMessageSize.
	ErrMessageTooLarge = errors.New("message too large")
	// ErrInvalidSDP indicates invalid SDP offer/answer.
	ErrInvalidSDP = errors.New("invalid SDP")
)

// NAT traversal errors
// Originally defined in: nat_traversal.go
var (
	// ErrNATTraversalFailed indicates NAT traversal failed.
	ErrNATTraversalFailed = errors.New("NAT traversal failed")
	// ErrAllMethodsFailed indicates all traversal methods failed.
	ErrAllMethodsFailed = errors.New("all NAT traversal methods failed")
)

// Relay errors
// Originally defined in: relay.go
var (
	// ErrNoRelayAvailable indicates no TURN relay is available.
	ErrNoRelayAvailable = errors.New("no relay available")
	// ErrRelayTimeout indicates relay connection timed out.
	ErrRelayTimeout = errors.New("relay connection timeout")
	// ErrRelayFull indicates the relay has reached capacity.
	ErrRelayFull = errors.New("relay at full capacity")
)

// Signaling errors
// Originally defined in: signaling.go
var (
	// ErrSignalingNotConnected indicates the signaling connection is not active.
	ErrSignalingNotConnected = errors.New("signaling not connected")
	// ErrPeerNotFound indicates the remote peer was not found.
	ErrPeerNotFound = errors.New("peer not found")
)

// STUN errors
// Originally defined in: stun.go
var (
	// ErrSTUNTimeout indicates STUN request timed out.
	ErrSTUNTimeout = errors.New("STUN request timeout")
	// ErrSTUNServerUnreachable indicates STUN server is unreachable.
	ErrSTUNServerUnreachable = errors.New("STUN server unreachable")
)
