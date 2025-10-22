package network

import "testing"

// TestComponentData_Structure verifies ComponentData struct initialization and fields.
func TestComponentData_Structure(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		data          []byte
	}{
		{"position_component", "position", []byte{1, 2, 3, 4}},
		{"velocity_component", "velocity", []byte{5, 6, 7, 8}},
		{"empty_data", "empty", []byte{}},
		{"nil_data", "nil", nil},
		{"large_data", "large", make([]byte, 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := ComponentData{
				Type: tt.componentType,
				Data: tt.data,
			}

			if cd.Type != tt.componentType {
				t.Errorf("Expected type %s, got %s", tt.componentType, cd.Type)
			}

			if tt.data == nil {
				if cd.Data != nil {
					t.Error("Expected nil data")
				}
			} else {
				if len(cd.Data) != len(tt.data) {
					t.Errorf("Expected data length %d, got %d", len(tt.data), len(cd.Data))
				}
			}
		})
	}
}

// TestStateUpdate_Structure verifies StateUpdate struct initialization.
func TestStateUpdate_Structure(t *testing.T) {
	components := []ComponentData{
		{Type: "position", Data: []byte{1, 2, 3}},
		{Type: "velocity", Data: []byte{4, 5, 6}},
	}

	update := StateUpdate{
		Timestamp:      1234567890,
		EntityID:       42,
		Components:     components,
		Priority:       128,
		SequenceNumber: 100,
	}

	if update.Timestamp != 1234567890 {
		t.Errorf("Expected timestamp 1234567890, got %d", update.Timestamp)
	}

	if update.EntityID != 42 {
		t.Errorf("Expected entity ID 42, got %d", update.EntityID)
	}

	if len(update.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(update.Components))
	}

	if update.Priority != 128 {
		t.Errorf("Expected priority 128, got %d", update.Priority)
	}

	if update.SequenceNumber != 100 {
		t.Errorf("Expected sequence number 100, got %d", update.SequenceNumber)
	}
}

// TestStateUpdate_PriorityLevels verifies priority value ranges.
func TestStateUpdate_PriorityLevels(t *testing.T) {
	tests := []struct {
		name     string
		priority uint8
	}{
		{"low_priority", 0},
		{"medium_priority", 128},
		{"high_priority", 200},
		{"critical_priority", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := StateUpdate{
				Priority: tt.priority,
			}

			if update.Priority != tt.priority {
				t.Errorf("Expected priority %d, got %d", tt.priority, update.Priority)
			}
		})
	}
}

// TestStateUpdate_EmptyComponents verifies behavior with no components.
func TestStateUpdate_EmptyComponents(t *testing.T) {
	update := StateUpdate{
		Timestamp:      100,
		EntityID:       1,
		Components:     []ComponentData{},
		Priority:       0,
		SequenceNumber: 0,
	}

	if len(update.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(update.Components))
	}
}

// TestStateUpdate_MultipleComponents verifies handling multiple components.
func TestStateUpdate_MultipleComponents(t *testing.T) {
	components := []ComponentData{
		{Type: "position", Data: []byte{1}},
		{Type: "velocity", Data: []byte{2}},
		{Type: "health", Data: []byte{3}},
		{Type: "sprite", Data: []byte{4}},
		{Type: "collider", Data: []byte{5}},
	}

	update := StateUpdate{
		EntityID:   1,
		Components: components,
	}

	if len(update.Components) != 5 {
		t.Errorf("Expected 5 components, got %d", len(update.Components))
	}

	// Verify component types
	expectedTypes := []string{"position", "velocity", "health", "sprite", "collider"}
	for i, comp := range update.Components {
		if comp.Type != expectedTypes[i] {
			t.Errorf("Component %d: expected type %s, got %s", i, expectedTypes[i], comp.Type)
		}
	}
}

// TestInputCommand_Structure verifies InputCommand struct initialization.
func TestInputCommand_Structure(t *testing.T) {
	cmd := InputCommand{
		PlayerID:       999,
		Timestamp:      1111111111,
		SequenceNumber: 50,
		InputType:      "move",
		Data:           []byte{10, 20, 30},
	}

	if cmd.PlayerID != 999 {
		t.Errorf("Expected player ID 999, got %d", cmd.PlayerID)
	}

	if cmd.Timestamp != 1111111111 {
		t.Errorf("Expected timestamp 1111111111, got %d", cmd.Timestamp)
	}

	if cmd.SequenceNumber != 50 {
		t.Errorf("Expected sequence number 50, got %d", cmd.SequenceNumber)
	}

	if cmd.InputType != "move" {
		t.Errorf("Expected input type 'move', got %s", cmd.InputType)
	}

	if len(cmd.Data) != 3 {
		t.Errorf("Expected data length 3, got %d", len(cmd.Data))
	}
}

// TestInputCommand_InputTypes verifies different input types.
func TestInputCommand_InputTypes(t *testing.T) {
	inputTypes := []string{
		"move",
		"attack",
		"use_item",
		"interact",
		"jump",
		"crouch",
		"inventory",
	}

	for _, inputType := range inputTypes {
		t.Run(inputType, func(t *testing.T) {
			cmd := InputCommand{
				PlayerID:  1,
				InputType: inputType,
			}

			if cmd.InputType != inputType {
				t.Errorf("Expected input type %s, got %s", inputType, cmd.InputType)
			}
		})
	}
}

// TestInputCommand_SequenceOrdering verifies sequence number ordering.
func TestInputCommand_SequenceOrdering(t *testing.T) {
	commands := []InputCommand{
		{SequenceNumber: 1},
		{SequenceNumber: 2},
		{SequenceNumber: 3},
		{SequenceNumber: 4},
		{SequenceNumber: 5},
	}

	for i, cmd := range commands {
		expectedSeq := uint32(i + 1)
		if cmd.SequenceNumber != expectedSeq {
			t.Errorf("Command %d: expected sequence %d, got %d", i, expectedSeq, cmd.SequenceNumber)
		}
	}
}

// TestConnectionInfo_Structure verifies ConnectionInfo struct initialization.
func TestConnectionInfo_Structure(t *testing.T) {
	conn := ConnectionInfo{
		PlayerID:  12345,
		Address:   "192.168.1.100:8080",
		Latency:   45.5,
		Connected: true,
	}

	if conn.PlayerID != 12345 {
		t.Errorf("Expected player ID 12345, got %d", conn.PlayerID)
	}

	if conn.Address != "192.168.1.100:8080" {
		t.Errorf("Expected address '192.168.1.100:8080', got %s", conn.Address)
	}

	if conn.Latency != 45.5 {
		t.Errorf("Expected latency 45.5, got %f", conn.Latency)
	}

	if !conn.Connected {
		t.Error("Expected connected to be true")
	}
}

// TestConnectionInfo_DisconnectedState verifies disconnected state.
func TestConnectionInfo_DisconnectedState(t *testing.T) {
	conn := ConnectionInfo{
		PlayerID:  1,
		Address:   "0.0.0.0:0",
		Latency:   0,
		Connected: false,
	}

	if conn.Connected {
		t.Error("Expected connected to be false")
	}

	if conn.Latency != 0 {
		t.Errorf("Expected latency 0 for disconnected, got %f", conn.Latency)
	}
}

// TestConnectionInfo_LatencyValues verifies various latency values.
func TestConnectionInfo_LatencyValues(t *testing.T) {
	tests := []struct {
		name    string
		latency float64
	}{
		{"excellent", 10.0},
		{"good", 50.0},
		{"moderate", 100.0},
		{"poor", 200.0},
		{"very_poor", 500.0},
		{"zero", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := ConnectionInfo{
				Latency: tt.latency,
			}

			if conn.Latency != tt.latency {
				t.Errorf("Expected latency %f, got %f", tt.latency, conn.Latency)
			}
		})
	}
}

// TestConnectionInfo_AddressFormats verifies different address formats.
func TestConnectionInfo_AddressFormats(t *testing.T) {
	addresses := []string{
		"127.0.0.1:8080",
		"192.168.1.1:9000",
		"10.0.0.1:3000",
		"example.com:8080",
		"[::1]:8080",
		"localhost:3000",
	}

	for _, addr := range addresses {
		t.Run(addr, func(t *testing.T) {
			conn := ConnectionInfo{
				Address: addr,
			}

			if conn.Address != addr {
				t.Errorf("Expected address %s, got %s", addr, conn.Address)
			}
		})
	}
}

// TestStateUpdate_SequenceNumberOverflow verifies large sequence numbers.
func TestStateUpdate_SequenceNumberOverflow(t *testing.T) {
	tests := []struct {
		name     string
		sequence uint32
	}{
		{"zero", 0},
		{"small", 100},
		{"medium", 10000},
		{"large", 1000000},
		{"max", 4294967295}, // uint32 max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := StateUpdate{
				SequenceNumber: tt.sequence,
			}

			if update.SequenceNumber != tt.sequence {
				t.Errorf("Expected sequence %d, got %d", tt.sequence, update.SequenceNumber)
			}
		})
	}
}

// TestInputCommand_EmptyData verifies behavior with empty data.
func TestInputCommand_EmptyData(t *testing.T) {
	cmd := InputCommand{
		PlayerID:  1,
		InputType: "ping",
		Data:      []byte{},
	}

	if len(cmd.Data) != 0 {
		t.Errorf("Expected empty data, got length %d", len(cmd.Data))
	}
}

// TestInputCommand_NilData verifies behavior with nil data.
func TestInputCommand_NilData(t *testing.T) {
	cmd := InputCommand{
		PlayerID:  1,
		InputType: "disconnect",
		Data:      nil,
	}

	if cmd.Data != nil {
		t.Error("Expected nil data")
	}
}

// TestStateUpdate_TimestampProgression verifies timestamp ordering.
func TestStateUpdate_TimestampProgression(t *testing.T) {
	updates := []StateUpdate{
		{Timestamp: 1000},
		{Timestamp: 2000},
		{Timestamp: 3000},
		{Timestamp: 4000},
		{Timestamp: 5000},
	}

	for i := 1; i < len(updates); i++ {
		if updates[i].Timestamp <= updates[i-1].Timestamp {
			t.Errorf("Update %d timestamp not increasing: %d -> %d",
				i, updates[i-1].Timestamp, updates[i].Timestamp)
		}
	}
}

// TestConnectionInfo_MultipleConnections verifies handling multiple connections.
func TestConnectionInfo_MultipleConnections(t *testing.T) {
	connections := []ConnectionInfo{
		{PlayerID: 1, Address: "192.168.1.1:8080", Connected: true},
		{PlayerID: 2, Address: "192.168.1.2:8080", Connected: true},
		{PlayerID: 3, Address: "192.168.1.3:8080", Connected: false},
		{PlayerID: 4, Address: "192.168.1.4:8080", Connected: true},
	}

	if len(connections) != 4 {
		t.Errorf("Expected 4 connections, got %d", len(connections))
	}

	connectedCount := 0
	for _, conn := range connections {
		if conn.Connected {
			connectedCount++
		}
	}

	if connectedCount != 3 {
		t.Errorf("Expected 3 connected, got %d", connectedCount)
	}
}

// TestComponentData_ZeroValue verifies zero-value initialization.
func TestComponentData_ZeroValue(t *testing.T) {
	var cd ComponentData

	if cd.Type != "" {
		t.Errorf("Expected empty type, got %s", cd.Type)
	}

	if cd.Data != nil {
		t.Error("Expected nil data for zero value")
	}
}

// TestStateUpdate_ZeroValue verifies zero-value initialization.
func TestStateUpdate_ZeroValue(t *testing.T) {
	var update StateUpdate

	if update.Timestamp != 0 {
		t.Errorf("Expected timestamp 0, got %d", update.Timestamp)
	}

	if update.EntityID != 0 {
		t.Errorf("Expected entity ID 0, got %d", update.EntityID)
	}

	if update.Components != nil {
		t.Error("Expected nil components for zero value")
	}

	if update.Priority != 0 {
		t.Errorf("Expected priority 0, got %d", update.Priority)
	}

	if update.SequenceNumber != 0 {
		t.Errorf("Expected sequence number 0, got %d", update.SequenceNumber)
	}
}

// TestInputCommand_ZeroValue verifies zero-value initialization.
func TestInputCommand_ZeroValue(t *testing.T) {
	var cmd InputCommand

	if cmd.PlayerID != 0 {
		t.Errorf("Expected player ID 0, got %d", cmd.PlayerID)
	}

	if cmd.Timestamp != 0 {
		t.Errorf("Expected timestamp 0, got %d", cmd.Timestamp)
	}

	if cmd.SequenceNumber != 0 {
		t.Errorf("Expected sequence number 0, got %d", cmd.SequenceNumber)
	}

	if cmd.InputType != "" {
		t.Errorf("Expected empty input type, got %s", cmd.InputType)
	}

	if cmd.Data != nil {
		t.Error("Expected nil data for zero value")
	}
}

// TestConnectionInfo_ZeroValue verifies zero-value initialization.
func TestConnectionInfo_ZeroValue(t *testing.T) {
	var conn ConnectionInfo

	if conn.PlayerID != 0 {
		t.Errorf("Expected player ID 0, got %d", conn.PlayerID)
	}

	if conn.Address != "" {
		t.Errorf("Expected empty address, got %s", conn.Address)
	}

	if conn.Latency != 0 {
		t.Errorf("Expected latency 0, got %f", conn.Latency)
	}

	if conn.Connected {
		t.Error("Expected connected to be false for zero value")
	}
}
