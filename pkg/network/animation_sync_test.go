package network

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestAnimationStatePacket_Encode tests packet encoding.
func TestAnimationStatePacket_Encode(t *testing.T) {
	tests := []struct {
		name    string
		packet  AnimationStatePacket
		wantErr bool
	}{
		{
			name: "valid idle state",
			packet: AnimationStatePacket{
				EntityID:   12345,
				State:      engine.AnimationStateIdle,
				FrameIndex: 0,
				Timestamp:  1234567890,
				Loop:       true,
			},
			wantErr: false,
		},
		{
			name: "valid walk state with frame",
			packet: AnimationStatePacket{
				EntityID:   67890,
				State:      engine.AnimationStateWalk,
				FrameIndex: 5,
				Timestamp:  9876543210,
				Loop:       true,
			},
			wantErr: false,
		},
		{
			name: "attack state no loop",
			packet: AnimationStatePacket{
				EntityID:   99999,
				State:      engine.AnimationStateAttack,
				FrameIndex: 3,
				Timestamp:  1111111111,
				Loop:       false,
			},
			wantErr: false,
		},
		{
			name: "frame index too large",
			packet: AnimationStatePacket{
				EntityID:   12345,
				State:      engine.AnimationStateIdle,
				FrameIndex: 70000, // > 65535
				Timestamp:  1234567890,
				Loop:       true,
			},
			wantErr: true,
		},
		{
			name: "negative frame index",
			packet: AnimationStatePacket{
				EntityID:   12345,
				State:      engine.AnimationStateIdle,
				FrameIndex: -1,
				Timestamp:  1234567890,
				Loop:       true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.packet.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) != 20 {
				t.Errorf("Encode() packet size = %d, want 20", len(data))
			}
		})
	}
}

// TestAnimationStatePacket_Decode tests packet decoding.
func TestAnimationStatePacket_Decode(t *testing.T) {
	tests := []struct {
		name       string
		packet     AnimationStatePacket
		wantErr    bool
		checkEqual bool
	}{
		{
			name: "decode idle state",
			packet: AnimationStatePacket{
				EntityID:   12345,
				State:      engine.AnimationStateIdle,
				FrameIndex: 0,
				Timestamp:  1234567890,
				Loop:       true,
			},
			checkEqual: true,
		},
		{
			name: "decode walk state",
			packet: AnimationStatePacket{
				EntityID:   67890,
				State:      engine.AnimationStateWalk,
				FrameIndex: 7,
				Timestamp:  9876543210,
				Loop:       false,
			},
			checkEqual: true,
		},
		{
			name: "decode all states",
			packet: AnimationStatePacket{
				EntityID:   11111,
				State:      engine.AnimationStateDeath,
				FrameIndex: 15,
				Timestamp:  5555555555,
				Loop:       false,
			},
			checkEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			data, err := tt.packet.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			var decoded AnimationStatePacket
			err = decoded.Decode(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkEqual {
				if decoded.EntityID != tt.packet.EntityID {
					t.Errorf("EntityID = %d, want %d", decoded.EntityID, tt.packet.EntityID)
				}
				if decoded.State != tt.packet.State {
					t.Errorf("State = %v, want %v", decoded.State, tt.packet.State)
				}
				if decoded.FrameIndex != tt.packet.FrameIndex {
					t.Errorf("FrameIndex = %d, want %d", decoded.FrameIndex, tt.packet.FrameIndex)
				}
				if decoded.Timestamp != tt.packet.Timestamp {
					t.Errorf("Timestamp = %d, want %d", decoded.Timestamp, tt.packet.Timestamp)
				}
				if decoded.Loop != tt.packet.Loop {
					t.Errorf("Loop = %v, want %v", decoded.Loop, tt.packet.Loop)
				}
			}
		})
	}
}

// TestAnimationStatePacket_Decode_InvalidData tests decoding with invalid data.
func TestAnimationStatePacket_Decode_InvalidData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "too short",
			data:    []byte{1, 2, 3, 4, 5},
			wantErr: true,
		},
		{
			name:    "19 bytes (one short)",
			data:    make([]byte, 19),
			wantErr: true,
		},
		{
			name:    "exactly 20 bytes (valid)",
			data:    make([]byte, 20),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var packet AnimationStatePacket
			err := packet.Decode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAnimationStateBatch_Encode tests batch encoding.
func TestAnimationStateBatch_Encode(t *testing.T) {
	tests := []struct {
		name       string
		batch      AnimationStateBatch
		wantErr    bool
		expectSize int
	}{
		{
			name: "empty batch",
			batch: AnimationStateBatch{
				States:    []AnimationStatePacket{},
				Timestamp: 1234567890,
			},
			wantErr:    false,
			expectSize: 10, // 2 (count) + 8 (timestamp)
		},
		{
			name: "single state",
			batch: AnimationStateBatch{
				States: []AnimationStatePacket{
					{EntityID: 1, State: engine.AnimationStateIdle, FrameIndex: 0, Timestamp: 100, Loop: true},
				},
				Timestamp: 1234567890,
			},
			wantErr:    false,
			expectSize: 30, // 10 (header) + 20 (state)
		},
		{
			name: "multiple states",
			batch: AnimationStateBatch{
				States: []AnimationStatePacket{
					{EntityID: 1, State: engine.AnimationStateIdle, FrameIndex: 0, Timestamp: 100, Loop: true},
					{EntityID: 2, State: engine.AnimationStateWalk, FrameIndex: 5, Timestamp: 200, Loop: true},
					{EntityID: 3, State: engine.AnimationStateAttack, FrameIndex: 2, Timestamp: 300, Loop: false},
				},
				Timestamp: 1234567890,
			},
			wantErr:    false,
			expectSize: 70, // 10 (header) + 60 (3 * 20)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.batch.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) != tt.expectSize {
				t.Errorf("Encode() size = %d, want %d", len(data), tt.expectSize)
			}
		})
	}
}

// TestAnimationStateBatch_Decode tests batch decoding.
func TestAnimationStateBatch_Decode(t *testing.T) {
	tests := []struct {
		name       string
		batch      AnimationStateBatch
		checkEqual bool
	}{
		{
			name: "decode empty batch",
			batch: AnimationStateBatch{
				States:    []AnimationStatePacket{},
				Timestamp: 1234567890,
			},
			checkEqual: true,
		},
		{
			name: "decode batch with states",
			batch: AnimationStateBatch{
				States: []AnimationStatePacket{
					{EntityID: 1, State: engine.AnimationStateIdle, FrameIndex: 0, Timestamp: 100, Loop: true},
					{EntityID: 2, State: engine.AnimationStateWalk, FrameIndex: 5, Timestamp: 200, Loop: true},
				},
				Timestamp: 9876543210,
			},
			checkEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			data, err := tt.batch.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			var decoded AnimationStateBatch
			err = decoded.Decode(data)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}

			if tt.checkEqual {
				if decoded.Timestamp != tt.batch.Timestamp {
					t.Errorf("Timestamp = %d, want %d", decoded.Timestamp, tt.batch.Timestamp)
				}
				if len(decoded.States) != len(tt.batch.States) {
					t.Errorf("States count = %d, want %d", len(decoded.States), len(tt.batch.States))
					return
				}
				for i := range tt.batch.States {
					if decoded.States[i].EntityID != tt.batch.States[i].EntityID {
						t.Errorf("State %d EntityID = %d, want %d", i, decoded.States[i].EntityID, tt.batch.States[i].EntityID)
					}
					if decoded.States[i].State != tt.batch.States[i].State {
						t.Errorf("State %d State = %v, want %v", i, decoded.States[i].State, tt.batch.States[i].State)
					}
				}
			}
		})
	}
}

// TestAnimationSyncManager_ShouldSync tests delta compression logic.
func TestAnimationSyncManager_ShouldSync(t *testing.T) {
	manager := NewAnimationSyncManager()

	entityID := uint64(12345)

	// First sync should always return true
	if !manager.ShouldSync(entityID, engine.AnimationStateIdle) {
		t.Error("First sync should return true")
	}

	// Record the sync
	manager.RecordSync(entityID, engine.AnimationStateIdle, 20)

	// Same state should return false (delta compression)
	if manager.ShouldSync(entityID, engine.AnimationStateIdle) {
		t.Error("Same state should not sync (delta compression)")
	}

	// Different state should return true
	if !manager.ShouldSync(entityID, engine.AnimationStateWalk) {
		t.Error("Different state should sync")
	}

	// Record new state
	manager.RecordSync(entityID, engine.AnimationStateWalk, 20)

	// Back to idle should return true
	if !manager.ShouldSync(entityID, engine.AnimationStateIdle) {
		t.Error("State change back should sync")
	}
}

// TestAnimationSyncManager_BufferState tests client-side state buffering.
func TestAnimationSyncManager_BufferState(t *testing.T) {
	manager := NewAnimationSyncManager()
	manager.bufferSize = 3

	entityID := uint64(12345)

	// Add first state - should not apply immediately
	packet1 := AnimationStatePacket{
		EntityID:   entityID,
		State:      engine.AnimationStateIdle,
		FrameIndex: 0,
		Timestamp:  100,
		Loop:       true,
	}
	if manager.BufferState(packet1) {
		t.Error("First buffered state should not apply immediately")
	}

	// Add second state
	packet2 := AnimationStatePacket{
		EntityID:   entityID,
		State:      engine.AnimationStateWalk,
		FrameIndex: 0,
		Timestamp:  200,
		Loop:       true,
	}
	if manager.BufferState(packet2) {
		t.Error("Second buffered state should not apply immediately")
	}

	// Add third state - buffer full, should apply
	packet3 := AnimationStatePacket{
		EntityID:   entityID,
		State:      engine.AnimationStateRun,
		FrameIndex: 0,
		Timestamp:  300,
		Loop:       true,
	}
	if !manager.BufferState(packet3) {
		t.Error("Third buffered state should trigger application (buffer full)")
	}

	// Get buffered states
	state1 := manager.GetNextState(entityID)
	if state1 == nil || state1.State != engine.AnimationStateIdle {
		t.Error("First state should be idle")
	}

	state2 := manager.GetNextState(entityID)
	if state2 == nil || state2.State != engine.AnimationStateWalk {
		t.Error("Second state should be walk")
	}

	state3 := manager.GetNextState(entityID)
	if state3 == nil || state3.State != engine.AnimationStateRun {
		t.Error("Third state should be run")
	}

	// Buffer should be empty now
	state4 := manager.GetNextState(entityID)
	if state4 != nil {
		t.Error("Buffer should be empty")
	}
}

// TestAnimationSyncManager_Stats tests statistics tracking.
func TestAnimationSyncManager_Stats(t *testing.T) {
	manager := NewAnimationSyncManager()

	// Record some syncs
	manager.RecordSync(1, engine.AnimationStateIdle, 20)
	manager.RecordSync(2, engine.AnimationStateWalk, 20)
	manager.RecordSync(3, engine.AnimationStateAttack, 20)

	// Receive some states
	manager.BufferState(AnimationStatePacket{EntityID: 4, State: engine.AnimationStateIdle})
	manager.BufferState(AnimationStatePacket{EntityID: 5, State: engine.AnimationStateWalk})

	stats := manager.Stats()

	if stats.StateChangesSent != 3 {
		t.Errorf("StateChangesSent = %d, want 3", stats.StateChangesSent)
	}
	if stats.StatesReceived != 2 {
		t.Errorf("StatesReceived = %d, want 2", stats.StatesReceived)
	}
	if stats.BytesTransmitted != 60 {
		t.Errorf("BytesTransmitted = %d, want 60", stats.BytesTransmitted)
	}
	if stats.BufferedEntities != 2 {
		t.Errorf("BufferedEntities = %d, want 2", stats.BufferedEntities)
	}

	// Test bandwidth calculation
	bandwidth := stats.Bandwidth(1 * time.Second)
	if bandwidth != 60.0 {
		t.Errorf("Bandwidth = %f, want 60.0", bandwidth)
	}
}

// TestAnimationSyncManager_ClearEntity tests entity cleanup.
func TestAnimationSyncManager_ClearEntity(t *testing.T) {
	manager := NewAnimationSyncManager()

	entityID := uint64(12345)

	// Add some data
	manager.RecordSync(entityID, engine.AnimationStateIdle, 20)
	manager.BufferState(AnimationStatePacket{EntityID: entityID, State: engine.AnimationStateWalk})

	// Verify data exists
	if !manager.ShouldSync(entityID, engine.AnimationStateIdle) {
		// Expected - same state
	}
	if manager.GetNextState(entityID) == nil {
		t.Error("Buffer should have data before clear")
	}

	// Clear entity
	manager.ClearEntity(entityID)

	// Verify data cleared
	if !manager.ShouldSync(entityID, engine.AnimationStateIdle) {
		t.Error("After clear, should sync (no tracking data)")
	}
	if manager.GetNextState(entityID) != nil {
		t.Error("Buffer should be empty after clear")
	}
}

// TestAnimationStateToID tests state ID conversion.
func TestAnimationStateToID(t *testing.T) {
	tests := []struct {
		state      engine.AnimationState
		expectedID uint8
	}{
		{engine.AnimationStateIdle, 0},
		{engine.AnimationStateWalk, 1},
		{engine.AnimationStateRun, 2},
		{engine.AnimationStateAttack, 3},
		{engine.AnimationStateCast, 4},
		{engine.AnimationStateHit, 5},
		{engine.AnimationStateDeath, 6},
		{engine.AnimationStateJump, 7},
		{engine.AnimationStateCrouch, 8},
		{engine.AnimationStateUse, 9},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			id := animationStateToID(tt.state)
			if id != tt.expectedID {
				t.Errorf("animationStateToID(%v) = %d, want %d", tt.state, id, tt.expectedID)
			}

			// Test round-trip
			state := idToAnimationState(id)
			if state != tt.state {
				t.Errorf("Round-trip failed: got %v, want %v", state, tt.state)
			}
		})
	}
}

// TestAnimationStateToID_Unknown tests unknown state handling.
func TestAnimationStateToID_Unknown(t *testing.T) {
	unknownState := engine.AnimationState("unknown_state")
	id := animationStateToID(unknownState)
	if id != 0 {
		t.Errorf("Unknown state should map to 0 (idle), got %d", id)
	}
}

// TestIDToAnimationState_Unknown tests unknown ID handling.
func TestIDToAnimationState_Unknown(t *testing.T) {
	unknownID := uint8(255)
	state := idToAnimationState(unknownID)
	if state != engine.AnimationStateIdle {
		t.Errorf("Unknown ID should map to idle, got %v", state)
	}
}

// BenchmarkAnimationStatePacket_Encode benchmarks packet encoding performance.
func BenchmarkAnimationStatePacket_Encode(b *testing.B) {
	packet := AnimationStatePacket{
		EntityID:   12345,
		State:      engine.AnimationStateWalk,
		FrameIndex: 5,
		Timestamp:  1234567890,
		Loop:       true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := packet.Encode()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAnimationStatePacket_Decode benchmarks packet decoding performance.
func BenchmarkAnimationStatePacket_Decode(b *testing.B) {
	packet := AnimationStatePacket{
		EntityID:   12345,
		State:      engine.AnimationStateWalk,
		FrameIndex: 5,
		Timestamp:  1234567890,
		Loop:       true,
	}
	data, _ := packet.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded AnimationStatePacket
		err := decoded.Decode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAnimationStateBatch_Encode benchmarks batch encoding with multiple states.
func BenchmarkAnimationStateBatch_Encode(b *testing.B) {
	batch := AnimationStateBatch{
		States: []AnimationStatePacket{
			{EntityID: 1, State: engine.AnimationStateIdle, FrameIndex: 0, Timestamp: 100, Loop: true},
			{EntityID: 2, State: engine.AnimationStateWalk, FrameIndex: 5, Timestamp: 200, Loop: true},
			{EntityID: 3, State: engine.AnimationStateAttack, FrameIndex: 2, Timestamp: 300, Loop: false},
			{EntityID: 4, State: engine.AnimationStateCast, FrameIndex: 1, Timestamp: 400, Loop: false},
			{EntityID: 5, State: engine.AnimationStateRun, FrameIndex: 3, Timestamp: 500, Loop: true},
		},
		Timestamp: 1234567890,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := batch.Encode()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAnimationSyncManager_ShouldSync benchmarks delta compression check.
func BenchmarkAnimationSyncManager_ShouldSync(b *testing.B) {
	manager := NewAnimationSyncManager()
	entityID := uint64(12345)
	manager.RecordSync(entityID, engine.AnimationStateIdle, 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ShouldSync(entityID, engine.AnimationStateIdle)
	}
}
