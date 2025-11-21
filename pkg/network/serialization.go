// Package network provides JSON serialization for network protocol.
// This file implements JSONProtocol which serializes game state and commands
// using JSON for human-readable network debugging.
package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// BinaryProtocol implements the Protocol interface using binary encoding.
// This provides efficient, compact serialization suitable for real-time multiplayer.
type BinaryProtocol struct{}

// NewBinaryProtocol creates a new binary protocol encoder/decoder.
func NewBinaryProtocol() *BinaryProtocol {
	return &BinaryProtocol{}
}

// EncodeStateUpdate serializes a state update to binary format.
// Binary format:
//   - Timestamp: 8 bytes (uint64)
//   - EntityID: 8 bytes (uint64)
//   - Priority: 1 byte (uint8)
//   - SequenceNumber: 4 bytes (uint32)
//   - ComponentCount: 2 bytes (uint16)
//   - For each component:
//   - TypeLength: 2 bytes (uint16)
//   - Type: variable (string bytes)
//   - DataLength: 4 bytes (uint32)
//   - Data: variable (byte array)
func (p *BinaryProtocol) EncodeStateUpdate(update *StateUpdate) ([]byte, error) {
	if update == nil {
		return nil, fmt.Errorf("cannot encode nil state update")
	}

	buf := new(bytes.Buffer)

	// Write fixed-size header
	if err := binary.Write(buf, binary.LittleEndian, update.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to write timestamp: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, update.EntityID); err != nil {
		return nil, fmt.Errorf("failed to write entity ID: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, update.Priority); err != nil {
		return nil, fmt.Errorf("failed to write priority: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, update.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to write sequence number: %w", err)
	}

	// Write component count
	componentCount := uint16(len(update.Components))
	if err := binary.Write(buf, binary.LittleEndian, componentCount); err != nil {
		return nil, fmt.Errorf("failed to write component count: %w", err)
	}

	// Write each component
	for _, comp := range update.Components {
		// Write type string
		typeBytes := []byte(comp.Type)
		typeLength := uint16(len(typeBytes))
		if err := binary.Write(buf, binary.LittleEndian, typeLength); err != nil {
			return nil, fmt.Errorf("failed to write type length: %w", err)
		}
		buf.Write(typeBytes)

		// Write data
		dataLength := uint32(len(comp.Data))
		if err := binary.Write(buf, binary.LittleEndian, dataLength); err != nil {
			return nil, fmt.Errorf("failed to write data length: %w", err)
		}
		buf.Write(comp.Data)
	}

	return buf.Bytes(), nil
}

// DecodeStateUpdate deserializes a state update from binary format.
func (p *BinaryProtocol) DecodeStateUpdate(data []byte) (*StateUpdate, error) {
	if len(data) < 23 {
		return nil, fmt.Errorf("data too short for state update: %d bytes", len(data))
	}

	buf := bytes.NewReader(data)
	update := &StateUpdate{}

	if err := decodeStateUpdateHeader(buf, update); err != nil {
		return nil, err
	}

	componentCount, err := decodeComponentCount(buf)
	if err != nil {
		return nil, err
	}

	if err := decodeComponents(buf, update, componentCount); err != nil {
		return nil, err
	}

	return update, nil
}

// decodeStateUpdateHeader reads the fixed-size header fields.
func decodeStateUpdateHeader(buf *bytes.Reader, update *StateUpdate) error {
	if err := binary.Read(buf, binary.LittleEndian, &update.Timestamp); err != nil {
		return fmt.Errorf("failed to read timestamp: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.EntityID); err != nil {
		return fmt.Errorf("failed to read entity ID: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.Priority); err != nil {
		return fmt.Errorf("failed to read priority: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.SequenceNumber); err != nil {
		return fmt.Errorf("failed to read sequence number: %w", err)
	}
	return nil
}

// decodeComponentCount reads and validates the component count.
func decodeComponentCount(buf *bytes.Reader) (uint16, error) {
	var componentCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &componentCount); err != nil {
		return 0, fmt.Errorf("failed to read component count: %w", err)
	}

	const maxComponentCount = 1000
	if componentCount > maxComponentCount {
		return 0, fmt.Errorf("component count too large: %d (max %d)", componentCount, maxComponentCount)
	}

	return componentCount, nil
}

// decodeComponents reads all component data from the buffer.
func decodeComponents(buf *bytes.Reader, update *StateUpdate, componentCount uint16) error {
	update.Components = make([]ComponentData, componentCount)
	for i := uint16(0); i < componentCount; i++ {
		if err := decodeComponentData(buf, &update.Components[i], i); err != nil {
			return err
		}
	}
	return nil
}

// decodeComponentData reads a single component's type and data.
func decodeComponentData(buf *bytes.Reader, component *ComponentData, index uint16) error {
	var typeLength uint16
	if err := binary.Read(buf, binary.LittleEndian, &typeLength); err != nil {
		return fmt.Errorf("failed to read type length for component %d: %w", index, err)
	}

	typeBytes := make([]byte, typeLength)
	if _, err := buf.Read(typeBytes); err != nil {
		return fmt.Errorf("failed to read type bytes for component %d: %w", index, err)
	}
	component.Type = string(typeBytes)

	var dataLength uint32
	if err := binary.Read(buf, binary.LittleEndian, &dataLength); err != nil {
		return fmt.Errorf("failed to read data length for component %d: %w", index, err)
	}

	if dataLength > 0 {
		dataBytes := make([]byte, dataLength)
		if _, err := buf.Read(dataBytes); err != nil {
			return fmt.Errorf("failed to read data bytes for component %d: %w", index, err)
		}
		component.Data = dataBytes
	}

	return nil
}

// EncodeInputCommand serializes an input command to binary format.
// Binary format:
//   - PlayerID: 8 bytes (uint64)
//   - Timestamp: 8 bytes (uint64)
//   - SequenceNumber: 4 bytes (uint32)
//   - InputTypeLength: 2 bytes (uint16)
//   - InputType: variable (string bytes)
//   - DataLength: 4 bytes (uint32)
//   - Data: variable (byte array)
func (p *BinaryProtocol) EncodeInputCommand(cmd *InputCommand) ([]byte, error) {
	if cmd == nil {
		return nil, fmt.Errorf("cannot encode nil input command")
	}

	buf := new(bytes.Buffer)

	// Write fixed-size fields
	if err := binary.Write(buf, binary.LittleEndian, cmd.PlayerID); err != nil {
		return nil, fmt.Errorf("failed to write player ID: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, cmd.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to write timestamp: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, cmd.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to write sequence number: %w", err)
	}

	// Write input type
	typeBytes := []byte(cmd.InputType)
	typeLength := uint16(len(typeBytes))
	if err := binary.Write(buf, binary.LittleEndian, typeLength); err != nil {
		return nil, fmt.Errorf("failed to write type length: %w", err)
	}
	buf.Write(typeBytes)

	// Write data
	dataLength := uint32(len(cmd.Data))
	if err := binary.Write(buf, binary.LittleEndian, dataLength); err != nil {
		return nil, fmt.Errorf("failed to write data length: %w", err)
	}
	buf.Write(cmd.Data)

	return buf.Bytes(), nil
}

// DecodeInputCommand deserializes an input command from binary format.
func (p *BinaryProtocol) DecodeInputCommand(data []byte) (*InputCommand, error) {
	if len(data) < 26 { // Minimum size: PlayerID (8) + Timestamp (8) + SequenceNumber (4) + TypeLength (2) + DataLength (4)
		return nil, fmt.Errorf("data too short for input command: %d bytes", len(data))
	}

	buf := bytes.NewReader(data)
	cmd := &InputCommand{}

	// Read fixed-size fields
	if err := binary.Read(buf, binary.LittleEndian, &cmd.PlayerID); err != nil {
		return nil, fmt.Errorf("failed to read player ID: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &cmd.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &cmd.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to read sequence number: %w", err)
	}

	// Read input type
	var typeLength uint16
	if err := binary.Read(buf, binary.LittleEndian, &typeLength); err != nil {
		return nil, fmt.Errorf("failed to read type length: %w", err)
	}

	typeBytes := make([]byte, typeLength)
	if _, err := buf.Read(typeBytes); err != nil {
		return nil, fmt.Errorf("failed to read type bytes: %w", err)
	}
	cmd.InputType = string(typeBytes)

	// Read data
	var dataLength uint32
	if err := binary.Read(buf, binary.LittleEndian, &dataLength); err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	if dataLength > 0 {
		dataBytes := make([]byte, dataLength)
		if _, err := buf.Read(dataBytes); err != nil {
			return nil, fmt.Errorf("failed to read data bytes: %w", err)
		}
		cmd.Data = dataBytes
	}

	return cmd, nil
}
