package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// AnimationStatePacket represents a networked animation state change.
// This packet is sent when an entity's animation state changes (e.g., idle -> walk).
// Only state changes are transmitted to minimize bandwidth usage.
type AnimationStatePacket struct {
	EntityID   uint64                  // Entity being animated
	State      engine.AnimationState   // New animation state
	FrameIndex int                     // Current frame in animation
	Timestamp  int64                   // Server timestamp (for interpolation)
	Loop       bool                    // Whether animation should loop
}

// AnimationStateBatch groups multiple animation state changes into a single packet.
// This reduces protocol overhead when many entities change state simultaneously.
type AnimationStateBatch struct {
	States    []AnimationStatePacket
	Timestamp int64 // Batch timestamp
}

// PacketType constants for animation packets
const (
	PacketTypeAnimationState      uint8 = 0x20 // Single state change
	PacketTypeAnimationStateBatch uint8 = 0x21 // Batch of state changes
)

// Encode serializes an AnimationStatePacket to bytes.
// Format: [EntityID:8][State:1][FrameIndex:2][Timestamp:8][Loop:1]
// Total: 20 bytes per packet (compact for bandwidth efficiency)
func (p *AnimationStatePacket) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write EntityID (8 bytes)
	if err := binary.Write(buf, binary.LittleEndian, p.EntityID); err != nil {
		return nil, fmt.Errorf("failed to encode EntityID: %w", err)
	}

	// Write State as uint8 (1 byte)
	stateID := animationStateToID(p.State)
	if err := binary.Write(buf, binary.LittleEndian, stateID); err != nil {
		return nil, fmt.Errorf("failed to encode State: %w", err)
	}

	// Write FrameIndex as uint16 (2 bytes, supports up to 65535 frames)
	if p.FrameIndex < 0 || p.FrameIndex > 65535 {
		return nil, fmt.Errorf("FrameIndex out of range: %d", p.FrameIndex)
	}
	if err := binary.Write(buf, binary.LittleEndian, uint16(p.FrameIndex)); err != nil {
		return nil, fmt.Errorf("failed to encode FrameIndex: %w", err)
	}

	// Write Timestamp (8 bytes)
	if err := binary.Write(buf, binary.LittleEndian, p.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to encode Timestamp: %w", err)
	}

	// Write Loop flag (1 byte)
	var loopByte uint8
	if p.Loop {
		loopByte = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, loopByte); err != nil {
		return nil, fmt.Errorf("failed to encode Loop: %w", err)
	}

	return buf.Bytes(), nil
}

// Decode deserializes bytes into an AnimationStatePacket.
func (p *AnimationStatePacket) Decode(data []byte) error {
	if len(data) < 20 {
		return fmt.Errorf("packet too short: expected 20 bytes, got %d", len(data))
	}

	buf := bytes.NewReader(data)

	// Read EntityID
	if err := binary.Read(buf, binary.LittleEndian, &p.EntityID); err != nil {
		return fmt.Errorf("failed to decode EntityID: %w", err)
	}

	// Read State
	var stateID uint8
	if err := binary.Read(buf, binary.LittleEndian, &stateID); err != nil {
		return fmt.Errorf("failed to decode State: %w", err)
	}
	p.State = idToAnimationState(stateID)

	// Read FrameIndex
	var frameIdx uint16
	if err := binary.Read(buf, binary.LittleEndian, &frameIdx); err != nil {
		return fmt.Errorf("failed to decode FrameIndex: %w", err)
	}
	p.FrameIndex = int(frameIdx)

	// Read Timestamp
	if err := binary.Read(buf, binary.LittleEndian, &p.Timestamp); err != nil {
		return fmt.Errorf("failed to decode Timestamp: %w", err)
	}

	// Read Loop flag
	var loopByte uint8
	if err := binary.Read(buf, binary.LittleEndian, &loopByte); err != nil {
		return fmt.Errorf("failed to decode Loop: %w", err)
	}
	p.Loop = loopByte == 1

	return nil
}

// Encode serializes an AnimationStateBatch to bytes.
// Format: [Count:2][Timestamp:8][States...]
func (b *AnimationStateBatch) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write count (2 bytes, supports up to 65535 states per batch)
	count := uint16(len(b.States))
	if err := binary.Write(buf, binary.LittleEndian, count); err != nil {
		return nil, fmt.Errorf("failed to encode count: %w", err)
	}

	// Write batch timestamp
	if err := binary.Write(buf, binary.LittleEndian, b.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to encode timestamp: %w", err)
	}

	// Write each state
	for i, state := range b.States {
		stateBytes, err := state.Encode()
		if err != nil {
			return nil, fmt.Errorf("failed to encode state %d: %w", i, err)
		}
		if _, err := buf.Write(stateBytes); err != nil {
			return nil, fmt.Errorf("failed to write state %d: %w", i, err)
		}
	}

	return buf.Bytes(), nil
}

// Decode deserializes bytes into an AnimationStateBatch.
func (b *AnimationStateBatch) Decode(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("batch packet too short: expected at least 10 bytes, got %d", len(data))
	}

	buf := bytes.NewReader(data)

	// Read count
	var count uint16
	if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
		return fmt.Errorf("failed to decode count: %w", err)
	}

	// Read timestamp
	if err := binary.Read(buf, binary.LittleEndian, &b.Timestamp); err != nil {
		return fmt.Errorf("failed to decode timestamp: %w", err)
	}

	// Read states
	b.States = make([]AnimationStatePacket, count)
	stateBytes := make([]byte, 20)
	for i := 0; i < int(count); i++ {
		n, err := buf.Read(stateBytes)
		if err != nil {
			return fmt.Errorf("failed to read state %d: %w", i, err)
		}
		if n != 20 {
			return fmt.Errorf("incomplete state %d: expected 20 bytes, got %d", i, n)
		}
		if err := b.States[i].Decode(stateBytes); err != nil {
			return fmt.Errorf("failed to decode state %d: %w", i, err)
		}
	}

	return nil
}

// AnimationSyncManager manages animation state synchronization for multiplayer.
// Server-side: Tracks state changes and broadcasts to relevant clients
// Client-side: Applies received state changes with interpolation
type AnimationSyncManager struct {
	// State tracking
	lastState map[uint64]engine.AnimationState // EntityID -> last sent state
	
	// Interpolation buffer (client-side)
	stateBuffer map[uint64][]AnimationStatePacket // EntityID -> buffered states
	bufferSize  int                                // Number of states to buffer (default: 3)
	
	// Statistics
	stateChangesSent uint64
	statesReceived   uint64
	bytesTransmitted uint64
}

// NewAnimationSyncManager creates a new animation sync manager.
func NewAnimationSyncManager() *AnimationSyncManager {
	return &AnimationSyncManager{
		lastState:   make(map[uint64]engine.AnimationState),
		stateBuffer: make(map[uint64][]AnimationStatePacket),
		bufferSize:  3, // 150ms buffer at 20 updates/sec
	}
}

// ShouldSync determines if an animation state change should be transmitted.
// Returns true only if the state has changed (delta compression).
func (m *AnimationSyncManager) ShouldSync(entityID uint64, newState engine.AnimationState) bool {
	lastState, exists := m.lastState[entityID]
	if !exists {
		// First time seeing this entity - always sync
		return true
	}
	
	// Only sync if state changed
	return lastState != newState
}

// RecordSync records that a state was transmitted.
func (m *AnimationSyncManager) RecordSync(entityID uint64, state engine.AnimationState, bytesSent int) {
	m.lastState[entityID] = state
	m.stateChangesSent++
	m.bytesTransmitted += uint64(bytesSent)
}

// BufferState adds a received state to the interpolation buffer.
// Client-side only. Returns true if the state should be applied immediately.
func (m *AnimationSyncManager) BufferState(packet AnimationStatePacket) bool {
	entityID := packet.EntityID
	
	// Add to buffer
	buffer := m.stateBuffer[entityID]
	buffer = append(buffer, packet)
	m.stateBuffer[entityID] = buffer
	m.statesReceived++
	
	// Apply immediately if buffer full
	return len(buffer) >= m.bufferSize
}

// GetNextState retrieves the next state from the interpolation buffer.
// Client-side only. Returns nil if buffer is empty.
func (m *AnimationSyncManager) GetNextState(entityID uint64) *AnimationStatePacket {
	buffer := m.stateBuffer[entityID]
	if len(buffer) == 0 {
		return nil
	}
	
	// Pop first state (FIFO)
	state := buffer[0]
	m.stateBuffer[entityID] = buffer[1:]
	
	return &state
}

// ClearEntity removes tracking data for an entity (when entity destroyed).
func (m *AnimationSyncManager) ClearEntity(entityID uint64) {
	delete(m.lastState, entityID)
	delete(m.stateBuffer, entityID)
}

// Stats returns synchronization statistics.
func (m *AnimationSyncManager) Stats() AnimationSyncStats {
	return AnimationSyncStats{
		StateChangesSent: m.stateChangesSent,
		StatesReceived:   m.statesReceived,
		BytesTransmitted: m.bytesTransmitted,
		BufferedEntities: len(m.stateBuffer),
	}
}

// AnimationSyncStats contains synchronization statistics.
type AnimationSyncStats struct {
	StateChangesSent uint64
	StatesReceived   uint64
	BytesTransmitted uint64
	BufferedEntities int
}

// Bandwidth returns average bytes per second over the given duration.
func (s AnimationSyncStats) Bandwidth(duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	return float64(s.BytesTransmitted) / duration.Seconds()
}

// animationStateToID converts an AnimationState to a compact uint8 ID.
func animationStateToID(state engine.AnimationState) uint8 {
	switch state {
	case engine.AnimationStateIdle:
		return 0
	case engine.AnimationStateWalk:
		return 1
	case engine.AnimationStateRun:
		return 2
	case engine.AnimationStateAttack:
		return 3
	case engine.AnimationStateCast:
		return 4
	case engine.AnimationStateHit:
		return 5
	case engine.AnimationStateDeath:
		return 6
	case engine.AnimationStateJump:
		return 7
	case engine.AnimationStateCrouch:
		return 8
	case engine.AnimationStateUse:
		return 9
	default:
		return 0 // Default to idle for unknown states
	}
}

// idToAnimationState converts a uint8 ID back to an AnimationState.
func idToAnimationState(id uint8) engine.AnimationState {
	switch id {
	case 0:
		return engine.AnimationStateIdle
	case 1:
		return engine.AnimationStateWalk
	case 2:
		return engine.AnimationStateRun
	case 3:
		return engine.AnimationStateAttack
	case 4:
		return engine.AnimationStateCast
	case 5:
		return engine.AnimationStateHit
	case 6:
		return engine.AnimationStateDeath
	case 7:
		return engine.AnimationStateJump
	case 8:
		return engine.AnimationStateCrouch
	case 9:
		return engine.AnimationStateUse
	default:
		return engine.AnimationStateIdle
	}
}
