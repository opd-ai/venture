// Package network provides network protocol data structures.
// This file defines core protocol types for state updates, input commands,
// and network messages used in client-server communication.
package network

// ComponentData represents serialized component data for network transmission.
type ComponentData struct {
	Type string
	Data []byte
}

// StateUpdate represents a network packet containing entity state changes.
type StateUpdate struct {
	// Timestamp of when this update was created (server time)
	Timestamp uint64

	// EntityID identifies which entity this update is for
	EntityID uint64

	// Components contains the updated component data
	Components []ComponentData

	// Priority determines update ordering (higher = more important)
	// 0 = low priority, 255 = critical
	Priority uint8

	// SequenceNumber for ordering and detecting packet loss
	SequenceNumber uint32
}

// InputCommand represents a player input sent from client to server.
type InputCommand struct {
	// PlayerID identifies which player sent this input
	PlayerID uint64

	// Timestamp when the input was generated (client time)
	Timestamp uint64

	// SequenceNumber for input ordering
	SequenceNumber uint32

	// InputType identifies the type of input (move, attack, use item, etc.)
	InputType string

	// Data contains the input-specific data (serialized)
	Data []byte
}

// ConnectionInfo contains information about a network connection.
type ConnectionInfo struct {
	// PlayerID uniquely identifies the player
	PlayerID uint64

	// Address is the network address (IP:port)
	Address string

	// Latency is the round-trip time in milliseconds
	Latency float64

	// Connected indicates if the connection is active
	Connected bool
}

// DeathMessage represents entity death notification from server to clients.
// Server broadcasts this message when an entity dies to synchronize death state.
// Category 1.1: Death & Revival System
type DeathMessage struct {
	// EntityID identifies the entity that died
	EntityID uint64

	// TimeOfDeath is the server timestamp when death occurred
	TimeOfDeath float64

	// KillerID identifies the entity that caused the death (0 if environmental)
	KillerID uint64

	// DroppedItemIDs contains entity IDs of items spawned from death
	DroppedItemIDs []uint64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}

// RevivalMessage represents player revival notification from server to clients.
// Server broadcasts this message when a player is revived by a teammate.
// Category 1.1: Death & Revival System
type RevivalMessage struct {
	// EntityID identifies the entity being revived
	EntityID uint64

	// ReviverID identifies the entity that performed the revival
	ReviverID uint64

	// TimeOfRevival is the server timestamp when revival occurred
	TimeOfRevival float64

	// RestoredHealth is the health amount restored (as fraction of max)
	RestoredHealth float64

	// SequenceNumber for message ordering
	SequenceNumber uint32
}
