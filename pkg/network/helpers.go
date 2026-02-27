package network

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrSystemClockInvalid is returned when the system clock reports a time before the Unix epoch.
var ErrSystemClockInvalid = errors.New("network: system clock returned timestamp before Unix epoch")

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
	// Calculate differences in both directions
	forwardDiff := sequenceDifference(seq, ref)
	backwardDiff := sequenceDifference(ref, seq)

	// Check if either direction is within range
	return forwardDiff <= rangeVal || backwardDiff <= rangeVal
}

// NowTimestamp returns the current time as a uint64 nanosecond timestamp
// suitable for use in network protocol messages (StateUpdate, InputCommand).
//
// This function converts time.Now().UnixNano() (int64) to uint64 for network
// transmission. The conversion is safe because:
//  1. Current time is always after Unix epoch (1970), so UnixNano() is positive
//  2. UnixNano() is valid for dates between 1678 and 2262
//  3. Positive int64 values convert to identical uint64 values
//
// System Requirements:
//   - 64-bit system (Go's time.Time.UnixNano() is undefined on 32-bit systems after 2038)
//   - Current date between 1970-2262 (guaranteed for all practical use cases)
//
// The uint64 format is used in network protocol for:
//   - Efficient binary serialization (8 bytes, little-endian)
//   - Checksum computation in desync detection
//   - Nanosecond precision for accurate timing and lag compensation
//
// Note: This timestamp is used for protocol-level timing and checksums.
// For save/load functionality and time arithmetic, use time.Time instead.
//
// Returns ErrSystemClockInvalid if the system clock reports a time before the Unix epoch.
func NowTimestamp() (uint64, error) {
	nanos := time.Now().UnixNano()

	// Defensive check: ensure timestamp is positive (should always be true for current dates)
	// If this fails, it indicates a system clock issue or date before 1970
	if nanos < 0 {
		logrus.WithFields(logrus.Fields{
			"system_name": "network_helpers",
			"nanos":       nanos,
		}).Error("system clock returned timestamp before Unix epoch")
		return 0, ErrSystemClockInvalid
	}

	return uint64(nanos), nil
}
