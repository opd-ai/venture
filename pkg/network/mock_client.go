package network

import (
	"fmt"
	"sync"
	"time"
)

// MockClient is a test implementation of ClientConnection without real network I/O.
// Use for unit testing code that depends on network clients.
type MockClient struct {
	// Configuration
	ConnectError    error // Error to return from Connect()
	DisconnectError error // Error to return from Disconnect()
	SendInputError  error // Error to return from SendInput()

	// State
	Connected bool
	PlayerID  uint64
	Latency   time.Duration

	// Recording
	ConnectCalls    int
	DisconnectCalls int
	SendInputCalls  int
	SentInputs      []struct {
		Type string
		Data []byte
	}

	// Channels
	stateUpdates chan *StateUpdate
	errors       chan error

	// Thread safety
	mu sync.RWMutex
}

// NewMockClient creates a new mock client for testing.
func NewMockClient() *MockClient {
	return &MockClient{
		stateUpdates: make(chan *StateUpdate, 16),
		errors:       make(chan error, 16),
		Latency:      50 * time.Millisecond, // Default simulated latency
	}
}

// Connect implements ClientConnection.
func (m *MockClient) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectCalls++

	if m.ConnectError != nil {
		return m.ConnectError
	}

	if m.Connected {
		return fmt.Errorf("already connected")
	}

	m.Connected = true
	return nil
}

// Disconnect implements ClientConnection.
func (m *MockClient) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DisconnectCalls++

	if m.DisconnectError != nil {
		return m.DisconnectError
	}

	if !m.Connected {
		return fmt.Errorf("not connected")
	}

	m.Connected = false
	return nil
}

// IsConnected implements ClientConnection.
func (m *MockClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Connected
}

// GetPlayerID implements ClientConnection.
func (m *MockClient) GetPlayerID() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.PlayerID
}

// SetPlayerID implements ClientConnection.
func (m *MockClient) SetPlayerID(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PlayerID = id
}

// GetLatency implements ClientConnection.
func (m *MockClient) GetLatency() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Latency
}

// SendInput implements ClientConnection.
func (m *MockClient) SendInput(inputType string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SendInputCalls++

	if m.SendInputError != nil {
		return m.SendInputError
	}

	if !m.Connected {
		return fmt.Errorf("not connected")
	}

	// Record the input
	m.SentInputs = append(m.SentInputs, struct {
		Type string
		Data []byte
	}{
		Type: inputType,
		Data: append([]byte(nil), data...), // Copy data
	})

	return nil
}

// ReceiveStateUpdate implements ClientConnection.
func (m *MockClient) ReceiveStateUpdate() <-chan *StateUpdate {
	return m.stateUpdates
}

// ReceiveError implements ClientConnection.
func (m *MockClient) ReceiveError() <-chan error {
	return m.errors
}

// SimulateStateUpdate injects a state update for testing.
// Use this to simulate server messages in tests.
func (m *MockClient) SimulateStateUpdate(update *StateUpdate) {
	select {
	case m.stateUpdates <- update:
	default:
		// Channel full - drop update (simulates network conditions)
	}
}

// SimulateError injects an error for testing.
// Use this to simulate network errors in tests.
func (m *MockClient) SimulateError(err error) {
	select {
	case m.errors <- err:
	default:
		// Channel full - drop error
	}
}

// SetLatency sets the simulated network latency.
// Use this to test behavior under different network conditions.
func (m *MockClient) SetLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Latency = latency
}

// GetSentInputCount returns the number of inputs sent.
func (m *MockClient) GetSentInputCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.SentInputs)
}

// GetSentInput returns a specific sent input by index.
func (m *MockClient) GetSentInput(index int) (inputType string, data []byte, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.SentInputs) {
		return "", nil, false
	}

	input := m.SentInputs[index]
	return input.Type, input.Data, true
}

// Reset clears all state and counters.
// Use this between test cases to reset the mock.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Connected = false
	m.PlayerID = 0
	m.ConnectCalls = 0
	m.DisconnectCalls = 0
	m.SendInputCalls = 0
	m.SentInputs = nil
	m.ConnectError = nil
	m.DisconnectError = nil
	m.SendInputError = nil

	// Drain channels
	for len(m.stateUpdates) > 0 {
		<-m.stateUpdates
	}
	for len(m.errors) > 0 {
		<-m.errors
	}
}

// Compile-time interface check
var _ ClientConnection = (*MockClient)(nil)
