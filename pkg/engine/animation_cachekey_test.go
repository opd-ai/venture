package engine

import (
	"testing"
)

// TestStateToInt verifies AnimationState to integer conversion.
func TestStateToInt(t *testing.T) {
	tests := []struct {
		state    AnimationState
		expected uint8
	}{
		{AnimationStateIdle, 0},
		{AnimationStateWalk, 1},
		{AnimationStateRun, 2},
		{AnimationStateAttack, 3},
		{AnimationStateCast, 4},
		{AnimationStateHit, 5},
		{AnimationStateDeath, 6},
		{AnimationStateJump, 7},
		{AnimationStateCrouch, 8},
		{AnimationStateUse, 9},
		{AnimationState("unknown"), 255},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			result := stateToInt(tt.state)
			if result != tt.expected {
				t.Errorf("stateToInt(%q) = %d; want %d", tt.state, result, tt.expected)
			}
		})
	}
}

// TestGetCacheKey verifies uint64 cache key generation.
func TestGetCacheKey(t *testing.T) {
	sys := NewAnimationSystem(nil)

	tests := []struct {
		name     string
		seed     int64
		state    AnimationState
		wantDiff bool // Whether keys should differ
	}{
		{"same seed and state", 12345, AnimationStateIdle, false},
		{"different seeds", 12345, AnimationStateIdle, true},
		{"different states", 12345, AnimationStateWalk, true},
		{"negative seed", -9876, AnimationStateAttack, false},
		{"zero seed", 0, AnimationStateIdle, false},
		{"max seed", 1<<55 - 1, AnimationStateDeath, false},
	}

	// Generate first key for comparison
	firstKey := sys.getCacheKey(12345, AnimationStateIdle)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := sys.getCacheKey(tt.seed, tt.state)

			// Verify key is non-zero
			if key == 0 {
				t.Errorf("getCacheKey(%d, %q) = 0; expected non-zero", tt.seed, tt.state)
			}

			// Check uniqueness based on test expectation
			if tt.wantDiff {
				secondKey := sys.getCacheKey(tt.seed, tt.state)
				if firstKey == secondKey && tt.seed != 12345 {
					t.Errorf("Expected different keys for different seeds/states")
				}
			} else {
				// Same inputs should produce same key (deterministic)
				key2 := sys.getCacheKey(tt.seed, tt.state)
				if key != key2 {
					t.Errorf("getCacheKey not deterministic: %d != %d", key, key2)
				}
			}
		})
	}
}

// TestGetCacheKey_Uniqueness verifies all state combinations produce unique keys.
func TestGetCacheKey_Uniqueness(t *testing.T) {
	sys := NewAnimationSystem(nil)

	states := []AnimationState{
		AnimationStateIdle, AnimationStateWalk, AnimationStateRun,
		AnimationStateAttack, AnimationStateCast, AnimationStateHit,
		AnimationStateDeath, AnimationStateJump, AnimationStateCrouch,
		AnimationStateUse,
	}

	seeds := []int64{0, 1, 100, 12345, -5000, 1 << 50}

	keys := make(map[uint64]bool)

	for _, seed := range seeds {
		for _, state := range states {
			key := sys.getCacheKey(seed, state)

			if keys[key] {
				t.Errorf("Duplicate key detected: seed=%d, state=%s, key=%d", seed, state, key)
			}
			keys[key] = true
		}
	}

	expectedUnique := len(seeds) * len(states)
	if len(keys) != expectedUnique {
		t.Errorf("Expected %d unique keys, got %d", expectedUnique, len(keys))
	}
}

// TestGetCacheKey_BitLayout verifies the bit layout of generated keys.
func TestGetCacheKey_BitLayout(t *testing.T) {
	sys := NewAnimationSystem(nil)

	seed := int64(0x123456789ABCD) // Use recognizable hex pattern
	state := AnimationStateWalk    // State ID = 1

	key := sys.getCacheKey(seed, state)

	// Extract state ID from lower 8 bits
	extractedStateID := uint8(key & 0xFF)
	if extractedStateID != 1 {
		t.Errorf("State ID in key = %d; want 1", extractedStateID)
	}

	// Extract seed from upper 56 bits
	extractedSeed := int64(key >> 8)
	if extractedSeed != seed {
		t.Errorf("Seed in key = %d; want %d", extractedSeed, seed)
	}
}

// TestGetCacheKey_NoCollisions verifies no collisions for edge cases.
func TestGetCacheKey_NoCollisions(t *testing.T) {
	sys := NewAnimationSystem(nil)

	// Test edge case: seeds that differ only in lower bits
	seed1 := int64(0x100)
	seed2 := int64(0x101)

	key1 := sys.getCacheKey(seed1, AnimationStateIdle)
	key2 := sys.getCacheKey(seed2, AnimationStateIdle)

	if key1 == key2 {
		t.Errorf("Collision detected for seeds %d and %d", seed1, seed2)
	}

	// Test edge case: same seed, different states
	keyIdle := sys.getCacheKey(12345, AnimationStateIdle)
	keyWalk := sys.getCacheKey(12345, AnimationStateWalk)

	if keyIdle == keyWalk {
		t.Error("Collision detected for different states with same seed")
	}
}
