package network

// NewStateUpdate creates a state update with the specified priority.
func NewStateUpdate(entityID uint64, priority uint8) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: priority,
	}
}

// NewCriticalUpdate creates a critical priority state update.
// Use for death events, revival events, and other game-critical state changes.
func NewCriticalUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityCritical,
	}
}

// NewHighPriorityUpdate creates a high priority state update.
// Use for combat events, damage notifications, and important gameplay changes.
func NewHighPriorityUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityHigh,
	}
}

// NewNormalUpdate creates a normal priority state update.
// Use for regular entity position/velocity updates and general state changes.
func NewNormalUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityNormal,
	}
}

// NewLowPriorityUpdate creates a low priority state update.
// Use for cosmetic changes like animations, particles, and visual effects.
func NewLowPriorityUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityLow,
	}
}

// sequenceLessThan compares two sequence numbers accounting for uint32 wrap-around.
// Returns true if seq1 is "less than" seq2 in circular sequence number space.
//
// This uses the "half-range" comparison algorithm: if the difference between
// two sequence numbers is less than half the total range (2^31), then the
// smaller number comes before the larger. This handles wrap-around correctly
// as long as sequence numbers don't differ by more than 2^31 (2.14 billion).
//
// For a 20Hz update rate, 2^31 updates = ~3.4 years, which is well beyond
// any reasonable lag compensation window.
func sequenceLessThan(seq1, seq2 uint32) bool {
	// If sequences are equal, seq1 is not less than seq2
	if seq1 == seq2 {
		return false
	}

	// Calculate the difference, accounting for wrap-around
	diff := seq2 - seq1

	// If the difference is less than half the range, seq1 < seq2
	// Otherwise, seq1 > seq2 (it wrapped around)
	return diff < (1 << 31)
}

// sequenceDifference calculates the difference between two sequence numbers,
// accounting for uint32 wrap-around. Returns newer - older.
//
// This assumes that 'newer' is the more recent sequence number and 'older'
// is the earlier one, and that they are within 2^31 updates of each other.
func sequenceDifference(newer, older uint32) uint32 {
	// Subtraction naturally handles wrap-around in uint32 arithmetic
	return newer - older
}

// sequenceInRange checks if a sequence number is within a range of another,
// accounting for wrap-around. Returns true if seq is within 'rangeVal' of ref.
//
// For example, sequenceInRange(100, 95, 10) returns true because 100 is
// within 10 of 95. Also handles wrap-around: sequenceInRange(5, UINT32_MAX-3, 10)
// returns true because 5 is within 10 of (UINT32_MAX-3) in circular space.
func sequenceInRange(seq, ref, rangeVal uint32) bool {
	diff := sequenceDifference(seq, ref)

	// Check if seq is ahead of ref by at most rangeVal
	if diff <= rangeVal {
		return true
	}

	// Check if ref is ahead of seq by at most rangeVal (seq is "behind")
	diff = sequenceDifference(ref, seq)
	return diff <= rangeVal
}
