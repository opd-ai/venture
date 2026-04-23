// Package aitypes defines minimal shared types for the behavior tree AI subsystem.
package aitypes

import "math/rand"

// Blackboard is a shared data structure for behavior tree state.
// It stores key-value pairs that nodes can read and write to share information.
// Entity references must be stored as interface{} and retrieved via Get; callers
// in pkg/engine use GetEntityFromBlackboard to recover the concrete *Entity type.
type Blackboard struct {
	data map[string]interface{}
	rng  *rand.Rand
}

// NewBlackboard creates a new empty blackboard with a default RNG.
// For deterministic behavior, use NewBlackboardWithSeed instead.
func NewBlackboard() *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
		rng:  rand.New(rand.NewSource(0)),
	}
}

// NewBlackboardWithSeed creates a new blackboard with a seeded RNG.
// Using the same seed ensures deterministic AI behavior.
func NewBlackboardWithSeed(seed int64) *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// GetRNG returns the seeded random number generator for deterministic behavior.
// All randomness in behavior tree actions should use this RNG.
func (b *Blackboard) GetRNG() *rand.Rand {
	return b.rng
}

// SetRNG sets the random number generator for this blackboard.
func (b *Blackboard) SetRNG(rng *rand.Rand) {
	b.rng = rng
}

// Set stores a value in the blackboard.
func (b *Blackboard) Set(key string, value interface{}) {
	b.data[key] = value
}

// Get retrieves a value from the blackboard.
// Returns the value and true if found, nil and false otherwise.
func (b *Blackboard) Get(key string) (interface{}, bool) {
	val, ok := b.data[key]
	return val, ok
}

// GetFloat64 retrieves a float64 value from the blackboard.
// Returns the value and true if found and correct type, 0.0 and false otherwise.
func (b *Blackboard) GetFloat64(key string) (float64, bool) {
	val, ok := b.data[key]
	if !ok {
		return 0.0, false
	}
	floatVal, ok := val.(float64)
	return floatVal, ok
}

// GetBool retrieves a bool value from the blackboard.
func (b *Blackboard) GetBool(key string) (bool, bool) {
	val, ok := b.data[key]
	if !ok {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}

// Clear removes all data from the blackboard.
func (b *Blackboard) Clear() {
	b.data = make(map[string]interface{})
}
