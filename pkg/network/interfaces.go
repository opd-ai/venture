// Package network provides network protocol interfaces.
// This file defines interfaces for network communication including
// Protocol for serialization and connection interfaces for testability.
package network

import "time"

// Protocol defines the interface for network protocol implementations.
// Originally from: protocol.go
type Protocol interface {
	// EncodeStateUpdate serializes a state update for transmission
	EncodeStateUpdate(update *StateUpdate) ([]byte, error)

	// DecodeStateUpdate deserializes a state update from network data
	DecodeStateUpdate(data []byte) (*StateUpdate, error)

	// EncodeInputCommand serializes an input command for transmission
	EncodeInputCommand(cmd *InputCommand) ([]byte, error)

	// DecodeInputCommand deserializes an input command from network data
	DecodeInputCommand(data []byte) (*InputCommand, error)
}

// ClientConnection defines the interface for network client operations.
// Implementations: TCPClient (production), MockClient (testing)
type ClientConnection interface {
	// Connect establishes connection to the server
	Connect() error

	// Disconnect closes the connection to the server
	Disconnect() error

	// IsConnected returns whether the client is currently connected
	IsConnected() bool

	// GetPlayerID returns the client's assigned player ID
	GetPlayerID() uint64

	// SetPlayerID sets the client's player ID
	SetPlayerID(id uint64)

	// GetLatency returns the current network latency
	GetLatency() time.Duration

	// SendInput sends an input command to the server
	SendInput(inputType string, data []byte) error

	// ReceiveStateUpdate returns a channel for receiving state updates
	ReceiveStateUpdate() <-chan *StateUpdate

	// ReceiveError returns a channel for receiving errors
	ReceiveError() <-chan error
}

// ServerConnection defines the interface for network server operations.
// Implementations: TCPServer (production), MockServer (testing)
type ServerConnection interface {
	// Start begins listening for client connections
	Start() error

	// Stop shuts down the server
	Stop() error

	// IsRunning returns whether the server is currently running
	IsRunning() bool

	// GetPlayerCount returns the number of connected players
	GetPlayerCount() int

	// GetPlayers returns a list of all connected player IDs
	GetPlayers() []uint64

	// BroadcastStateUpdate sends a state update to all connected clients
	BroadcastStateUpdate(update *StateUpdate)

	// SendStateUpdate sends a state update to a specific player
	SendStateUpdate(playerID uint64, update *StateUpdate) error

	// ReceiveInputCommand returns a channel for receiving input commands
	ReceiveInputCommand() <-chan *InputCommand

	// ReceivePlayerJoin returns a channel for player connection events
	ReceivePlayerJoin() <-chan uint64

	// ReceivePlayerLeave returns a channel for player disconnection events
	ReceivePlayerLeave() <-chan uint64

	// ReceiveError returns a channel for receiving errors
	ReceiveError() <-chan error
}
