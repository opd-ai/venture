// Package webrtc constants consolidation.
// This file contains all enum constant definitions used throughout the webrtc package,
// consolidated from types.go, nat_traversal.go, relay.go, and stun.go for centralized
// constant management and easier discovery of available enum values.
package webrtc

// Constants for WebRTC federation.
// Code relocated from: types.go, nat_traversal.go, relay.go, stun.go

// ConnectionState constants
// Originally from: types.go
const (
	// StateNew indicates the connection has been created but not started.
	StateNew ConnectionState = iota
	// StateConnecting indicates ICE/DTLS negotiation in progress.
	StateConnecting
	// StateConnected indicates the P2P connection is established and ready.
	StateConnected
	// StateDisconnected indicates the connection was lost but may reconnect.
	StateDisconnected
	// StateFailed indicates the connection failed permanently.
	StateFailed
	// StateClosed indicates the connection was closed deliberately.
	StateClosed
)

// ICECandidateType constants
// Originally from: types.go
const (
	// CandidateHost indicates a local network address.
	CandidateHost ICECandidateType = iota
	// CandidateServerReflexive indicates a public IP from STUN.
	CandidateServerReflexive
	// CandidateRelay indicates a TURN relay address.
	CandidateRelay
)

// TraversalMethod constants
// Originally from: nat_traversal.go
const (
	// MethodDirect indicates direct P2P connection (no NAT).
	MethodDirect TraversalMethod = iota
	// MethodSTUN indicates STUN-assisted connection.
	MethodSTUN
	// MethodTURN indicates TURN relay connection.
	MethodTURN
	// MethodFailed indicates all methods failed.
	MethodFailed
)

// SelectionStrategy constants
// Originally from: relay.go
const (
	// StrategyLowestLatency selects relay with lowest latency.
	StrategyLowestLatency SelectionStrategy = iota
	// StrategyHighestBandwidth selects relay with highest bandwidth.
	StrategyHighestBandwidth
	// StrategyLowestUtilization selects relay with most available capacity.
	StrategyLowestUtilization
	// StrategyRoundRobin rotates through available relays.
	StrategyRoundRobin
)

// NATType constants
// Originally from: stun.go
const (
	// NATTypeUnknown indicates NAT type could not be determined.
	NATTypeUnknown NATType = iota
	// NATTypeNone indicates no NAT (direct internet connection).
	NATTypeNone
	// NATTypeFullCone allows any external host to send packets.
	NATTypeFullCone
	// NATTypeRestrictedCone allows packets from hosts we've sent to.
	NATTypeRestrictedCone
	// NATTypePortRestrictedCone requires matching port and IP.
	NATTypePortRestrictedCone
	// NATTypeSymmetric uses different mappings for different destinations.
	NATTypeSymmetric
)
