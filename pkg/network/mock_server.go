package network

import (
	"fmt"
	"sync"
)

// MockServer is a test implementation of ServerConnection without real network I/O.
// Use for unit testing code that depends on network servers.
type MockServer struct {
	// Configuration
	StartError error // Error to return from Start()
	StopError  error // Error to return from Stop()
	SendError  error // Error to return from SendStateUpdate()

	// State
	Running bool
	Players map[uint64]bool // Set of connected players

	// Recording
	StartCalls     int
	StopCalls      int
	BroadcastCalls int
	SendCalls      int
	SentUpdates    []struct {
		PlayerID uint64        // 0 for broadcasts
		Update   *StateUpdate
	}

	// Channels
	inputCommands chan *InputCommand
	playerJoins   chan uint64
	playerLeaves  chan uint64
	errors        chan error

	// Thread safety
	mu sync.RWMutex
}

// NewMockServer creates a new mock server for testing.
func NewMockServer() *MockServer {
	return &MockServer{
		Players:       make(map[uint64]bool),
		inputCommands: make(chan *InputCommand, 64),
		playerJoins:   make(chan uint64, 16),
		playerLeaves:  make(chan uint64, 16),
		errors:        make(chan error, 16),
	}
}

// Start implements ServerConnection.
func (m *MockServer) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartCalls++

	if m.StartError != nil {
		return m.StartError
	}

	if m.Running {
		return fmt.Errorf("server already running")
	}

	m.Running = true
	return nil
}

// Stop implements ServerConnection.
func (m *MockServer) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StopCalls++

	if m.StopError != nil {
		return m.StopError
	}

	if !m.Running {
		return fmt.Errorf("server not running")
	}

	m.Running = false
	// Disconnect all players
	m.Players = make(map[uint64]bool)
	return nil
}

// IsRunning implements ServerConnection.
func (m *MockServer) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Running
}

// GetPlayerCount implements ServerConnection.
func (m *MockServer) GetPlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Players)
}

// GetPlayers implements ServerConnection.
func (m *MockServer) GetPlayers() []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	players := make([]uint64, 0, len(m.Players))
	for id := range m.Players {
		players = append(players, id)
	}
	return players
}

// BroadcastStateUpdate implements ServerConnection.
func (m *MockServer) BroadcastStateUpdate(update *StateUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.BroadcastCalls++

	// Record the broadcast
	m.SentUpdates = append(m.SentUpdates, struct {
		PlayerID uint64
		Update   *StateUpdate
	}{
		PlayerID: 0, // 0 indicates broadcast
		Update:   update,
	})
}

// SendStateUpdate implements ServerConnection.
func (m *MockServer) SendStateUpdate(playerID uint64, update *StateUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SendCalls++

	if m.SendError != nil {
		return m.SendError
	}

	if !m.Running {
		return fmt.Errorf("server not running")
	}

	if !m.Players[playerID] {
		return fmt.Errorf("player %d not connected", playerID)
	}

	// Record the send
	m.SentUpdates = append(m.SentUpdates, struct {
		PlayerID uint64
		Update   *StateUpdate
	}{
		PlayerID: playerID,
		Update:   update,
	})

	return nil
}

// ReceiveInputCommand implements ServerConnection.
func (m *MockServer) ReceiveInputCommand() <-chan *InputCommand {
	return m.inputCommands
}

// ReceivePlayerJoin implements ServerConnection.
func (m *MockServer) ReceivePlayerJoin() <-chan uint64 {
	return m.playerJoins
}

// ReceivePlayerLeave implements ServerConnection.
func (m *MockServer) ReceivePlayerLeave() <-chan uint64 {
	return m.playerLeaves
}

// ReceiveError implements ServerConnection.
func (m *MockServer) ReceiveError() <-chan error {
	return m.errors
}

// SimulatePlayerJoin simulates a player connecting.
// Use this to test player connection handling.
func (m *MockServer) SimulatePlayerJoin(playerID uint64) {
	m.mu.Lock()
	m.Players[playerID] = true
	m.mu.Unlock()

	select {
	case m.playerJoins <- playerID:
	default:
		// Channel full
	}
}

// SimulatePlayerLeave simulates a player disconnecting.
// Use this to test player disconnection handling.
func (m *MockServer) SimulatePlayerLeave(playerID uint64) {
	m.mu.Lock()
	delete(m.Players, playerID)
	m.mu.Unlock()

	select {
	case m.playerLeaves <- playerID:
	default:
		// Channel full
	}
}

// SimulateInputCommand injects an input command for testing.
// Use this to simulate client input in tests.
func (m *MockServer) SimulateInputCommand(cmd *InputCommand) {
	select {
	case m.inputCommands <- cmd:
	default:
		// Channel full - drop command
	}
}

// SimulateError injects an error for testing.
// Use this to simulate network errors in tests.
func (m *MockServer) SimulateError(err error) {
	select {
	case m.errors <- err:
	default:
		// Channel full - drop error
	}
}

// GetSentUpdateCount returns the number of updates sent (broadcast or targeted).
func (m *MockServer) GetSentUpdateCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.SentUpdates)
}

// GetSentUpdate returns a specific sent update by index.
// playerID of 0 indicates a broadcast.
func (m *MockServer) GetSentUpdate(index int) (playerID uint64, update *StateUpdate, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.SentUpdates) {
		return 0, nil, false
	}

	sent := m.SentUpdates[index]
	return sent.PlayerID, sent.Update, true
}

// Reset clears all state and counters.
// Use this between test cases to reset the mock.
func (m *MockServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Running = false
	m.Players = make(map[uint64]bool)
	m.StartCalls = 0
	m.StopCalls = 0
	m.BroadcastCalls = 0
	m.SendCalls = 0
	m.SentUpdates = nil
	m.StartError = nil
	m.StopError = nil
	m.SendError = nil

	// Drain channels
	for len(m.inputCommands) > 0 {
		<-m.inputCommands
	}
	for len(m.playerJoins) > 0 {
		<-m.playerJoins
	}
	for len(m.playerLeaves) > 0 {
		<-m.playerLeaves
	}
	for len(m.errors) > 0 {
		<-m.errors
	}
}

// Compile-time interface check
var _ ServerConnection = (*MockServer)(nil)
