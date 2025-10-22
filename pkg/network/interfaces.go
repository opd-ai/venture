package network

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
