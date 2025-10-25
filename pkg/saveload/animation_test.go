package saveload

import (
	"encoding/json"
	"testing"
)

// TestPlayerStateAnimationSerialization tests PlayerState with animation data serializes correctly.
func TestPlayerStateAnimationSerialization(t *testing.T) {
	tests := []struct {
		name           string
		animationState *AnimationStateData
		wantNil        bool
	}{
		{
			name: "with animation data",
			animationState: &AnimationStateData{
				State:          "walk",
				FrameIndex:     3,
				Loop:           true,
				LastUpdateTime: 1.5,
			},
			wantNil: false,
		},
		{
			name:           "without animation data",
			animationState: nil,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create player state with animation
			player := &PlayerState{
				EntityID:       12345,
				X:              100.0,
				Y:              200.0,
				Level:          5,
				AnimationState: tt.animationState,
			}

			// Serialize to JSON
			data, err := json.Marshal(player)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			// Deserialize from JSON
			var loaded PlayerState
			err = json.Unmarshal(data, &loaded)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			// Verify basic fields
			if loaded.EntityID != player.EntityID {
				t.Errorf("EntityID = %d, want %d", loaded.EntityID, player.EntityID)
			}

			// Verify animation state
			if tt.wantNil {
				if loaded.AnimationState != nil {
					t.Errorf("Expected nil animation state, got %+v", loaded.AnimationState)
				}
			} else {
				if loaded.AnimationState == nil {
					t.Fatal("Animation state is nil, expected non-nil")
				}
				if loaded.AnimationState.State != tt.animationState.State {
					t.Errorf("State = %s, want %s", loaded.AnimationState.State, tt.animationState.State)
				}
				if loaded.AnimationState.FrameIndex != tt.animationState.FrameIndex {
					t.Errorf("FrameIndex = %d, want %d", loaded.AnimationState.FrameIndex, tt.animationState.FrameIndex)
				}
				if loaded.AnimationState.Loop != tt.animationState.Loop {
					t.Errorf("Loop = %v, want %v", loaded.AnimationState.Loop, tt.animationState.Loop)
				}
			}
		})
	}
}

// TestModifiedEntityAnimationSerialization tests ModifiedEntity with animation data serializes correctly.
func TestModifiedEntityAnimationSerialization(t *testing.T) {
	tests := []struct {
		name           string
		animationState *AnimationStateData
		wantNil        bool
	}{
		{
			name: "entity with animation",
			animationState: &AnimationStateData{
				State:          "attack",
				FrameIndex:     5,
				Loop:           false,
				LastUpdateTime: 2.75,
			},
			wantNil: false,
		},
		{
			name:           "entity without animation",
			animationState: nil,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create modified entity with animation
			entity := ModifiedEntity{
				EntityID:       67890,
				X:              50.5,
				Y:              75.3,
				Health:         80.0,
				IsAlive:        true,
				AnimationState: tt.animationState,
			}

			// Serialize to JSON
			data, err := json.Marshal(entity)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			// Deserialize from JSON
			var loaded ModifiedEntity
			err = json.Unmarshal(data, &loaded)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			// Verify basic fields
			if loaded.EntityID != entity.EntityID {
				t.Errorf("EntityID = %d, want %d", loaded.EntityID, entity.EntityID)
			}
			if loaded.Health != entity.Health {
				t.Errorf("Health = %f, want %f", loaded.Health, entity.Health)
			}

			// Verify animation state
			if tt.wantNil {
				if loaded.AnimationState != nil {
					t.Errorf("Expected nil animation state, got %+v", loaded.AnimationState)
				}
			} else {
				if loaded.AnimationState == nil {
					t.Fatal("Animation state is nil, expected non-nil")
				}
				if loaded.AnimationState.State != tt.animationState.State {
					t.Errorf("State = %s, want %s", loaded.AnimationState.State, tt.animationState.State)
				}
				if loaded.AnimationState.FrameIndex != tt.animationState.FrameIndex {
					t.Errorf("FrameIndex = %d, want %d", loaded.AnimationState.FrameIndex, tt.animationState.FrameIndex)
				}
			}
		})
	}
}

// TestBackwardCompatibility tests loading saves without animation data.
func TestBackwardCompatibility(t *testing.T) {
	// Simulate old save format without animation_state field
	oldSaveJSON := `{
		"entity_id": 12345,
		"x": 100.0,
		"y": 200.0,
		"level": 5,
		"current_health": 100.0,
		"max_health": 100.0
	}`

	var player PlayerState
	err := json.Unmarshal([]byte(oldSaveJSON), &player)
	if err != nil {
		t.Fatalf("Failed to unmarshal old format: %v", err)
	}

	// Verify animation state is nil (backward compatible)
	if player.AnimationState != nil {
		t.Errorf("Expected nil animation state for old save format, got %+v", player.AnimationState)
	}

	// Verify basic fields loaded correctly
	if player.EntityID != 12345 {
		t.Errorf("EntityID = %d, want 12345", player.EntityID)
	}
	if player.Level != 5 {
		t.Errorf("Level = %d, want 5", player.Level)
	}
}

// TestGameSaveAnimationSerialization tests full GameSave with animation data.
func TestGameSaveAnimationSerialization(t *testing.T) {
	// Create full game save with animation data
	save := NewGameSave()
	save.PlayerState.EntityID = 12345
	save.PlayerState.Level = 10
	save.PlayerState.AnimationState = &AnimationStateData{
		State:          "walk",
		FrameIndex:     3,
		Loop:           true,
		LastUpdateTime: 1.5,
	}
	save.WorldState.Seed = 67890
	save.WorldState.ModifiedEntities = []ModifiedEntity{
		{
			EntityID: 100,
			X:        50.0,
			Y:        60.0,
			IsAlive:  true,
			AnimationState: &AnimationStateData{
				State:      "idle",
				FrameIndex: 0,
				Loop:       true,
			},
		},
		{
			EntityID: 101,
			X:        75.0,
			Y:        80.0,
			IsAlive:  false,
			AnimationState: &AnimationStateData{
				State:      "death",
				FrameIndex: 7,
				Loop:       false,
			},
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Deserialize from JSON
	var loaded GameSave
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Verify player animation state
	if loaded.PlayerState.AnimationState == nil {
		t.Fatal("Player animation state is nil")
	}
	if loaded.PlayerState.AnimationState.State != "walk" {
		t.Errorf("Player animation state = %s, want walk", loaded.PlayerState.AnimationState.State)
	}
	if loaded.PlayerState.AnimationState.FrameIndex != 3 {
		t.Errorf("Player frame index = %d, want 3", loaded.PlayerState.AnimationState.FrameIndex)
	}

	// Verify entity animation states
	if len(loaded.WorldState.ModifiedEntities) != 2 {
		t.Fatalf("Expected 2 entities, got %d", len(loaded.WorldState.ModifiedEntities))
	}

	entity1 := loaded.WorldState.ModifiedEntities[0]
	if entity1.AnimationState == nil {
		t.Fatal("Entity 1 animation state is nil")
	}
	if entity1.AnimationState.State != "idle" {
		t.Errorf("Entity 1 animation state = %s, want idle", entity1.AnimationState.State)
	}

	entity2 := loaded.WorldState.ModifiedEntities[1]
	if entity2.AnimationState == nil {
		t.Fatal("Entity 2 animation state is nil")
	}
	if entity2.AnimationState.State != "death" {
		t.Errorf("Entity 2 animation state = %s, want death", entity2.AnimationState.State)
	}
	if entity2.AnimationState.FrameIndex != 7 {
		t.Errorf("Entity 2 frame index = %d, want 7", entity2.AnimationState.FrameIndex)
	}
}

// TestAnimationStateDeterminism tests that animation state serialization is deterministic.
func TestAnimationStateDeterminism(t *testing.T) {
	// Create same animation state twice
	state1 := &AnimationStateData{
		State:          "run",
		FrameIndex:     2,
		Loop:           true,
		LastUpdateTime: 0.75,
	}
	state2 := &AnimationStateData{
		State:          "run",
		FrameIndex:     2,
		Loop:           true,
		LastUpdateTime: 0.75,
	}

	// Serialize both
	data1, err1 := json.Marshal(state1)
	data2, err2 := json.Marshal(state2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Serialization failed: %v, %v", err1, err2)
	}

	// Compare JSON outputs (should be identical)
	if string(data1) != string(data2) {
		t.Errorf("Determinism failure:\ndata1 = %s\ndata2 = %s", data1, data2)
	}
}

// TestAllAnimationStates tests all standard animation states serialize correctly.
func TestAllAnimationStates(t *testing.T) {
	states := []string{
		"idle", "walk", "run", "attack", "cast",
		"hit", "death", "jump", "crouch", "use",
	}

	for i, state := range states {
		t.Run(state, func(t *testing.T) {
			data := &AnimationStateData{
				State:          state,
				FrameIndex:     uint8(i),
				Loop:           i%2 == 0, // Alternate loop/no-loop
				LastUpdateTime: float64(i) * 0.5,
			}

			// Serialize
			jsonData, err := json.Marshal(data)
			if err != nil {
				t.Fatalf("Serialization failed for %s: %v", state, err)
			}

			// Deserialize
			var loaded AnimationStateData
			err = json.Unmarshal(jsonData, &loaded)
			if err != nil {
				t.Fatalf("Deserialization failed for %s: %v", state, err)
			}

			// Verify
			if loaded.State != state {
				t.Errorf("State = %s, want %s", loaded.State, state)
			}
			if loaded.FrameIndex != uint8(i) {
				t.Errorf("FrameIndex = %d, want %d", loaded.FrameIndex, i)
			}
			if loaded.Loop != (i%2 == 0) {
				t.Errorf("Loop = %v, want %v", loaded.Loop, i%2 == 0)
			}
		})
	}
}

// BenchmarkAnimationStateJSONMarshal benchmarks JSON serialization.
func BenchmarkAnimationStateJSONMarshal(b *testing.B) {
	data := &AnimationStateData{
		State:          "walk",
		FrameIndex:     3,
		Loop:           true,
		LastUpdateTime: 1.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

// BenchmarkAnimationStateJSONUnmarshal benchmarks JSON deserialization.
func BenchmarkAnimationStateJSONUnmarshal(b *testing.B) {
	jsonData := []byte(`{"state":"walk","frame_index":3,"loop":true,"last_update_time":1.5}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data AnimationStateData
		_ = json.Unmarshal(jsonData, &data)
	}
}

// BenchmarkFullGameSaveWithAnimation benchmarks serializing a complete game save with animation data.
func BenchmarkFullGameSaveWithAnimation(b *testing.B) {
	save := NewGameSave()
	save.PlayerState.AnimationState = &AnimationStateData{
		State:      "walk",
		FrameIndex: 3,
		Loop:       true,
	}
	save.WorldState.ModifiedEntities = []ModifiedEntity{
		{
			EntityID: 100,
			AnimationState: &AnimationStateData{
				State:      "idle",
				FrameIndex: 0,
				Loop:       true,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(save)
	}
}
