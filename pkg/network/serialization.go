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
	binary.Write(buf, binary.LittleEndian, update.Timestamp)
	binary.Write(buf, binary.LittleEndian, update.EntityID)
	binary.Write(buf, binary.LittleEndian, update.Priority)
	binary.Write(buf, binary.LittleEndian, update.SequenceNumber)

	// Write component count
	componentCount := uint16(len(update.Components))
	binary.Write(buf, binary.LittleEndian, componentCount)

	// Write each component
	for _, comp := range update.Components {
		// Write type string
		typeBytes := []byte(comp.Type)
		typeLength := uint16(len(typeBytes))
		binary.Write(buf, binary.LittleEndian, typeLength)
		buf.Write(typeBytes)

		// Write data
		dataLength := uint32(len(comp.Data))
		binary.Write(buf, binary.LittleEndian, dataLength)
		buf.Write(comp.Data)
	}

	return buf.Bytes(), nil
}

// DecodeStateUpdate deserializes a state update from binary format.
func (p *BinaryProtocol) DecodeStateUpdate(data []byte) (*StateUpdate, error) {
	if len(data) < 23 { // Minimum size: header (21) + component count (2)
		return nil, fmt.Errorf("data too short for state update: %d bytes", len(data))
	}

	buf := bytes.NewReader(data)
	update := &StateUpdate{}

	// Read fixed-size header
	if err := binary.Read(buf, binary.LittleEndian, &update.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.EntityID); err != nil {
		return nil, fmt.Errorf("failed to read entity ID: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.Priority); err != nil {
		return nil, fmt.Errorf("failed to read priority: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &update.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to read sequence number: %w", err)
	}

	// Read component count
	var componentCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &componentCount); err != nil {
		return nil, fmt.Errorf("failed to read component count: %w", err)
	}

	// Read components
	update.Components = make([]ComponentData, componentCount)
	for i := uint16(0); i < componentCount; i++ {
		// Read type string
		var typeLength uint16
		if err := binary.Read(buf, binary.LittleEndian, &typeLength); err != nil {
			return nil, fmt.Errorf("failed to read type length for component %d: %w", i, err)
		}

		typeBytes := make([]byte, typeLength)
		if _, err := buf.Read(typeBytes); err != nil {
			return nil, fmt.Errorf("failed to read type bytes for component %d: %w", i, err)
		}
		update.Components[i].Type = string(typeBytes)

		// Read data
		var dataLength uint32
		if err := binary.Read(buf, binary.LittleEndian, &dataLength); err != nil {
			return nil, fmt.Errorf("failed to read data length for component %d: %w", i, err)
		}

		if dataLength > 0 {
			dataBytes := make([]byte, dataLength)
			if _, err := buf.Read(dataBytes); err != nil {
				return nil, fmt.Errorf("failed to read data bytes for component %d: %w", i, err)
			}
			update.Components[i].Data = dataBytes
		}
	}

	return update, nil
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
	binary.Write(buf, binary.LittleEndian, cmd.PlayerID)
	binary.Write(buf, binary.LittleEndian, cmd.Timestamp)
	binary.Write(buf, binary.LittleEndian, cmd.SequenceNumber)

	// Write input type
	typeBytes := []byte(cmd.InputType)
	typeLength := uint16(len(typeBytes))
	binary.Write(buf, binary.LittleEndian, typeLength)
	buf.Write(typeBytes)

	// Write data
	dataLength := uint32(len(cmd.Data))
	binary.Write(buf, binary.LittleEndian, dataLength)
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
