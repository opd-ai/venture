package network

import (
	"bytes"
	"testing"
)

// TestBinaryProtocol_EncodeStateUpdate verifies state update encoding.
func TestBinaryProtocol_EncodeStateUpdate(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name    string
		update  *StateUpdate
		wantErr bool
	}{
		{
			name: "valid_update_single_component",
			update: &StateUpdate{
				Timestamp:      12345,
				EntityID:       100,
				Priority:       128,
				SequenceNumber: 42,
				Components: []ComponentData{
					{Type: "position", Data: []byte{1, 2, 3, 4}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_update_multiple_components",
			update: &StateUpdate{
				Timestamp:      99999,
				EntityID:       200,
				Priority:       255,
				SequenceNumber: 100,
				Components: []ComponentData{
					{Type: "position", Data: []byte{1, 2, 3, 4}},
					{Type: "velocity", Data: []byte{5, 6, 7, 8}},
					{Type: "health", Data: []byte{9, 10}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_update_no_components",
			update: &StateUpdate{
				Timestamp:      1000,
				EntityID:       1,
				Priority:       0,
				SequenceNumber: 1,
				Components:     []ComponentData{},
			},
			wantErr: false,
		},
		{
			name: "valid_update_empty_data",
			update: &StateUpdate{
				Timestamp:      5000,
				EntityID:       50,
				Priority:       50,
				SequenceNumber: 5,
				Components: []ComponentData{
					{Type: "tag", Data: []byte{}},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil_update",
			update:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := protocol.EncodeStateUpdate(tt.update)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("Expected non-empty encoded data")
			}
		})
	}
}

// TestBinaryProtocol_DecodeStateUpdate verifies state update decoding.
func TestBinaryProtocol_DecodeStateUpdate(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name    string
		update  *StateUpdate
		wantErr bool
	}{
		{
			name: "valid_decode_single_component",
			update: &StateUpdate{
				Timestamp:      12345,
				EntityID:       100,
				Priority:       128,
				SequenceNumber: 42,
				Components: []ComponentData{
					{Type: "position", Data: []byte{1, 2, 3, 4}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_decode_multiple_components",
			update: &StateUpdate{
				Timestamp:      99999,
				EntityID:       200,
				Priority:       255,
				SequenceNumber: 100,
				Components: []ComponentData{
					{Type: "position", Data: []byte{1, 2, 3, 4}},
					{Type: "velocity", Data: []byte{5, 6}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_decode_no_components",
			update: &StateUpdate{
				Timestamp:      1000,
				EntityID:       1,
				Priority:       0,
				SequenceNumber: 1,
				Components:     []ComponentData{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			encoded, err := protocol.EncodeStateUpdate(tt.update)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Decode
			decoded, err := protocol.DecodeStateUpdate(encoded)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify fields match
			if decoded.Timestamp != tt.update.Timestamp {
				t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, tt.update.Timestamp)
			}
			if decoded.EntityID != tt.update.EntityID {
				t.Errorf("EntityID mismatch: got %d, want %d", decoded.EntityID, tt.update.EntityID)
			}
			if decoded.Priority != tt.update.Priority {
				t.Errorf("Priority mismatch: got %d, want %d", decoded.Priority, tt.update.Priority)
			}
			if decoded.SequenceNumber != tt.update.SequenceNumber {
				t.Errorf("SequenceNumber mismatch: got %d, want %d", decoded.SequenceNumber, tt.update.SequenceNumber)
			}
			if len(decoded.Components) != len(tt.update.Components) {
				t.Errorf("Component count mismatch: got %d, want %d", len(decoded.Components), len(tt.update.Components))
			}

			// Verify each component
			for i, comp := range decoded.Components {
				if i >= len(tt.update.Components) {
					break
				}
				if comp.Type != tt.update.Components[i].Type {
					t.Errorf("Component %d type mismatch: got %s, want %s", i, comp.Type, tt.update.Components[i].Type)
				}
				if !bytes.Equal(comp.Data, tt.update.Components[i].Data) {
					t.Errorf("Component %d data mismatch", i)
				}
			}
		})
	}
}

// TestBinaryProtocol_DecodeStateUpdate_InvalidData verifies error handling for invalid data.
func TestBinaryProtocol_DecodeStateUpdate_InvalidData(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty_data", []byte{}},
		{"too_short", []byte{1, 2, 3}},
		{"truncated_header", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.DecodeStateUpdate(tt.data)
			if err == nil {
				t.Error("Expected error for invalid data but got none")
			}
		})
	}
}

// TestBinaryProtocol_EncodeInputCommand verifies input command encoding.
func TestBinaryProtocol_EncodeInputCommand(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name    string
		cmd     *InputCommand
		wantErr bool
	}{
		{
			name: "valid_command_with_data",
			cmd: &InputCommand{
				PlayerID:       123,
				Timestamp:      5555,
				SequenceNumber: 10,
				InputType:      "move",
				Data:           []byte{1, 2, 3, 4, 5},
			},
			wantErr: false,
		},
		{
			name: "valid_command_empty_data",
			cmd: &InputCommand{
				PlayerID:       456,
				Timestamp:      7777,
				SequenceNumber: 20,
				InputType:      "ping",
				Data:           []byte{},
			},
			wantErr: false,
		},
		{
			name: "valid_command_long_type",
			cmd: &InputCommand{
				PlayerID:       789,
				Timestamp:      9999,
				SequenceNumber: 30,
				InputType:      "use_item_from_inventory",
				Data:           []byte{10, 20, 30},
			},
			wantErr: false,
		},
		{
			name:    "nil_command",
			cmd:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := protocol.EncodeInputCommand(tt.cmd)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("Expected non-empty encoded data")
			}
		})
	}
}

// TestBinaryProtocol_DecodeInputCommand verifies input command decoding.
func TestBinaryProtocol_DecodeInputCommand(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name    string
		cmd     *InputCommand
		wantErr bool
	}{
		{
			name: "valid_decode_with_data",
			cmd: &InputCommand{
				PlayerID:       123,
				Timestamp:      5555,
				SequenceNumber: 10,
				InputType:      "move",
				Data:           []byte{1, 2, 3, 4, 5},
			},
			wantErr: false,
		},
		{
			name: "valid_decode_empty_data",
			cmd: &InputCommand{
				PlayerID:       456,
				Timestamp:      7777,
				SequenceNumber: 20,
				InputType:      "attack",
				Data:           []byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			encoded, err := protocol.EncodeInputCommand(tt.cmd)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Decode
			decoded, err := protocol.DecodeInputCommand(encoded)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify fields match
			if decoded.PlayerID != tt.cmd.PlayerID {
				t.Errorf("PlayerID mismatch: got %d, want %d", decoded.PlayerID, tt.cmd.PlayerID)
			}
			if decoded.Timestamp != tt.cmd.Timestamp {
				t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, tt.cmd.Timestamp)
			}
			if decoded.SequenceNumber != tt.cmd.SequenceNumber {
				t.Errorf("SequenceNumber mismatch: got %d, want %d", decoded.SequenceNumber, tt.cmd.SequenceNumber)
			}
			if decoded.InputType != tt.cmd.InputType {
				t.Errorf("InputType mismatch: got %s, want %s", decoded.InputType, tt.cmd.InputType)
			}
			if !bytes.Equal(decoded.Data, tt.cmd.Data) {
				t.Error("Data mismatch")
			}
		})
	}
}

// TestBinaryProtocol_DecodeInputCommand_InvalidData verifies error handling for invalid data.
func TestBinaryProtocol_DecodeInputCommand_InvalidData(t *testing.T) {
	protocol := NewBinaryProtocol()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty_data", []byte{}},
		{"too_short", []byte{1, 2, 3}},
		{"truncated_header", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.DecodeInputCommand(tt.data)
			if err == nil {
				t.Error("Expected error for invalid data but got none")
			}
		})
	}
}

// TestBinaryProtocol_RoundTrip_StateUpdate verifies full encode-decode cycle for state updates.
func TestBinaryProtocol_RoundTrip_StateUpdate(t *testing.T) {
	protocol := NewBinaryProtocol()

	original := &StateUpdate{
		Timestamp:      999888777,
		EntityID:       42,
		Priority:       200,
		SequenceNumber: 12345,
		Components: []ComponentData{
			{Type: "position", Data: []byte{10, 20, 30, 40, 50}},
			{Type: "velocity", Data: []byte{1, 2, 3, 4}},
			{Type: "health", Data: []byte{100, 200}},
		},
	}

	// Encode
	encoded, err := protocol.EncodeStateUpdate(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	decoded, err := protocol.DecodeStateUpdate(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all fields
	if decoded.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, original.Timestamp)
	}
	if decoded.EntityID != original.EntityID {
		t.Errorf("EntityID mismatch: got %d, want %d", decoded.EntityID, original.EntityID)
	}
	if decoded.Priority != original.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", decoded.Priority, original.Priority)
	}
	if decoded.SequenceNumber != original.SequenceNumber {
		t.Errorf("SequenceNumber mismatch: got %d, want %d", decoded.SequenceNumber, original.SequenceNumber)
	}
	if len(decoded.Components) != len(original.Components) {
		t.Fatalf("Component count mismatch: got %d, want %d", len(decoded.Components), len(original.Components))
	}

	for i, comp := range decoded.Components {
		if comp.Type != original.Components[i].Type {
			t.Errorf("Component %d type mismatch: got %s, want %s", i, comp.Type, original.Components[i].Type)
		}
		if !bytes.Equal(comp.Data, original.Components[i].Data) {
			t.Errorf("Component %d data mismatch", i)
		}
	}
}

// TestBinaryProtocol_RoundTrip_InputCommand verifies full encode-decode cycle for input commands.
func TestBinaryProtocol_RoundTrip_InputCommand(t *testing.T) {
	protocol := NewBinaryProtocol()

	original := &InputCommand{
		PlayerID:       987654321,
		Timestamp:      111222333444,
		SequenceNumber: 55555,
		InputType:      "use_ability",
		Data:           []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	// Encode
	encoded, err := protocol.EncodeInputCommand(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	decoded, err := protocol.DecodeInputCommand(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all fields
	if decoded.PlayerID != original.PlayerID {
		t.Errorf("PlayerID mismatch: got %d, want %d", decoded.PlayerID, original.PlayerID)
	}
	if decoded.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, original.Timestamp)
	}
	if decoded.SequenceNumber != original.SequenceNumber {
		t.Errorf("SequenceNumber mismatch: got %d, want %d", decoded.SequenceNumber, original.SequenceNumber)
	}
	if decoded.InputType != original.InputType {
		t.Errorf("InputType mismatch: got %s, want %s", decoded.InputType, original.InputType)
	}
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Error("Data mismatch")
	}
}

// BenchmarkEncodeStateUpdate measures state update encoding performance.
func BenchmarkEncodeStateUpdate(b *testing.B) {
	protocol := NewBinaryProtocol()
	update := &StateUpdate{
		Timestamp:      12345,
		EntityID:       100,
		Priority:       128,
		SequenceNumber: 42,
		Components: []ComponentData{
			{Type: "position", Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
			{Type: "velocity", Data: []byte{1, 2, 3, 4}},
			{Type: "health", Data: []byte{100, 200}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = protocol.EncodeStateUpdate(update)
	}
}

// BenchmarkDecodeStateUpdate measures state update decoding performance.
func BenchmarkDecodeStateUpdate(b *testing.B) {
	protocol := NewBinaryProtocol()
	update := &StateUpdate{
		Timestamp:      12345,
		EntityID:       100,
		Priority:       128,
		SequenceNumber: 42,
		Components: []ComponentData{
			{Type: "position", Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
			{Type: "velocity", Data: []byte{1, 2, 3, 4}},
			{Type: "health", Data: []byte{100, 200}},
		},
	}

	encoded, _ := protocol.EncodeStateUpdate(update)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = protocol.DecodeStateUpdate(encoded)
	}
}

// BenchmarkEncodeInputCommand measures input command encoding performance.
func BenchmarkEncodeInputCommand(b *testing.B) {
	protocol := NewBinaryProtocol()
	cmd := &InputCommand{
		PlayerID:       123,
		Timestamp:      5555,
		SequenceNumber: 10,
		InputType:      "move",
		Data:           []byte{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = protocol.EncodeInputCommand(cmd)
	}
}

// BenchmarkDecodeInputCommand measures input command decoding performance.
func BenchmarkDecodeInputCommand(b *testing.B) {
	protocol := NewBinaryProtocol()
	cmd := &InputCommand{
		PlayerID:       123,
		Timestamp:      5555,
		SequenceNumber: 10,
		InputType:      "move",
		Data:           []byte{1, 2, 3, 4, 5},
	}

	encoded, _ := protocol.EncodeInputCommand(cmd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = protocol.DecodeInputCommand(encoded)
	}
}
