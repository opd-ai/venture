// Package network provides lag compensation for fair hit detection.
// This file implements lag compensation using snapshot history to rewind
// game state for accurate hit detection on the server.
package network

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LagCompensator provides server-side lag compensation for hit detection
// and other time-critical operations in multiplayer games.
//
// Lag compensation works by "rewinding" the game world to the time when
// the player performed an action (e.g., fired a shot), accounting for
// their latency. This ensures fair hit detection even with high latency.
type LagCompensator struct {
	mu sync.RWMutex

	// Snapshot manager for historical world states
	snapshots *SnapshotManager

	// Maximum latency to compensate for (prevents abuse)
	maxCompensation time.Duration

	// Minimum latency to compensate for (ignores trivial delays)
	minCompensation time.Duration
}

// LagCompensationConfig configures the lag compensation system
type LagCompensationConfig struct {
	// MaxCompensation is the maximum latency to compensate for
	// Typical: 500ms for normal play, 5000ms for high-latency (Tor)
	MaxCompensation time.Duration

	// MinCompensation is the minimum latency to compensate for
	// Latencies below this are considered negligible
	MinCompensation time.Duration

	// SnapshotBufferSize is the number of snapshots to keep
	// Should be large enough to cover MaxCompensation
	// At 20 updates/sec, 100 snapshots = 5 seconds
	SnapshotBufferSize int

	// DeltaCompressionEpsilon controls sensitivity for delta compression
	// Smaller values = more sensitive = higher bandwidth
	// Larger values = less sensitive = lower bandwidth
	// Typical: 0.001 (default), 0.01 (high-latency)
	DeltaCompressionEpsilon float64
}

// DefaultLagCompensationConfig returns a default configuration
// for typical internet play (up to 500ms latency)
func DefaultLagCompensationConfig() LagCompensationConfig {
	return LagCompensationConfig{
		MaxCompensation:         500 * time.Millisecond,
		MinCompensation:         10 * time.Millisecond,
		SnapshotBufferSize:      100,
		DeltaCompressionEpsilon: 0.001, // Default: high sensitivity for accurate low-latency play
	}
}

// HighLatencyLagCompensationConfig returns a configuration
// for high-latency connections (e.g., Tor, up to 5000ms).
// Uses 300 snapshots at 20Hz = 15 seconds of history, providing adequate
// buffer for 5000ms MaxCompensation with 10s round-trip time plus safety margin.
// Also uses higher epsilon (0.01) for better bandwidth efficiency under high latency.
func HighLatencyLagCompensationConfig() LagCompensationConfig {
	return LagCompensationConfig{
		MaxCompensation:         5000 * time.Millisecond,
		MinCompensation:         10 * time.Millisecond,
		SnapshotBufferSize:      300,  // 15s at 20Hz (was 200 = 10s)
		DeltaCompressionEpsilon: 0.01, // 10x less sensitive for bandwidth efficiency
	}
}

// NewLagCompensator creates a new lag compensator with the given configuration
func NewLagCompensator(config LagCompensationConfig) *LagCompensator {
	epsilon := config.DeltaCompressionEpsilon
	if epsilon <= 0 {
		epsilon = 0.001 // Default if not specified
	}
	return &LagCompensator{
		snapshots:       NewSnapshotManagerWithEpsilon(config.SnapshotBufferSize, epsilon),
		maxCompensation: config.MaxCompensation,
		minCompensation: config.MinCompensation,
	}
}

// RecordSnapshot records a world snapshot for future lag compensation
func (lc *LagCompensator) RecordSnapshot(snapshot WorldSnapshot) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.snapshots.AddSnapshot(snapshot)
}

// RewindResult contains the results of a lag compensation rewind
type RewindResult struct {
	// Success indicates if the rewind was successful
	Success bool

	// Snapshot is the historical world state at the compensated time
	Snapshot *WorldSnapshot

	// CompensatedTime is the actual time we rewound to
	CompensatedTime time.Time

	// ActualLatency is the latency we compensated for
	ActualLatency time.Duration

	// WasClamped indicates if the latency was clamped to max/min
	WasClamped bool
}

// RewindToPlayerTime rewinds the world to the time when the player
// performed an action, accounting for their latency.
//
// This is the core of lag compensation:
// 1. Calculate when the player saw the world (now - latency)
// 2. Retrieve the world snapshot from that time
// 3. Perform hit detection against that historical state
// 4. Validate that the result is fair (within compensation limits)
func (lc *LagCompensator) RewindToPlayerTime(playerLatency time.Duration) *RewindResult {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return lc.rewindToPlayerTimeUnlocked(playerLatency)
}

// GAP-002 REPAIR: Internal unlocked version to prevent recursive locking
// This method assumes the caller already holds the appropriate lock.
// Used by methods like ValidateHit that need to call rewind logic while
// already holding a lock, avoiding deadlock.
func (lc *LagCompensator) rewindToPlayerTimeUnlocked(playerLatency time.Duration) *RewindResult {
	now := time.Now()
	result := &RewindResult{
		Success:    false,
		WasClamped: false,
	}

	// Clamp latency to configured bounds
	compensatedLatency := playerLatency
	if compensatedLatency > lc.maxCompensation {
		compensatedLatency = lc.maxCompensation
		result.WasClamped = true
	}
	if compensatedLatency < lc.minCompensation {
		compensatedLatency = lc.minCompensation
		result.WasClamped = true
	}

	result.ActualLatency = compensatedLatency

	// Calculate the time the player saw the world
	compensatedTime := now.Add(-compensatedLatency)
	result.CompensatedTime = compensatedTime

	// Retrieve historical snapshot
	snapshot := lc.snapshots.GetSnapshotAtTime(compensatedTime)
	if snapshot == nil {
		return result
	}

	result.Snapshot = snapshot
	result.Success = true
	return result
}

// ValidateHit checks if a hit is valid given the player's latency
// and the current world state. This prevents exploits where players
// might manipulate latency to gain an advantage.
//
// Returns true if the hit is valid, false otherwise.
//
// GAP-002 REPAIR: Uses unlocked internal method to avoid recursive locking deadlock
func (lc *LagCompensator) ValidateHit(
	attackerID uint64,
	targetID uint64,
	hitPosition Position,
	playerLatency time.Duration,
	hitRadius float64,
) (bool, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	// GAP-002 FIX: Use internal unlocked version to avoid recursive lock
	rewind := lc.rewindToPlayerTimeUnlocked(playerLatency)
	if !rewind.Success {
		logrus.WithFields(logrus.Fields{
			"system_name":    "lag_compensator",
			"attackerID":     attackerID,
			"targetID":       targetID,
			"player_latency": playerLatency.String(),
		}).Warn("failed to rewind to player time")
		return false, fmt.Errorf("failed to rewind to player time: no snapshot available")
	}

	// Check if target existed at that time
	targetSnapshot, exists := rewind.Snapshot.Entities[targetID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"system_name":      "lag_compensator",
			"attackerID":       attackerID,
			"targetID":         targetID,
			"compensated_time": rewind.CompensatedTime.Format(time.RFC3339Nano),
		}).Debug("target entity did not exist at compensated time")
		return false, fmt.Errorf("target entity %d did not exist at compensated time", targetID)
	}

	// Check if attacker existed at that time
	_, exists = rewind.Snapshot.Entities[attackerID]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"system_name":      "lag_compensator",
			"attackerID":       attackerID,
			"targetID":         targetID,
			"compensated_time": rewind.CompensatedTime.Format(time.RFC3339Nano),
		}).Debug("attacker entity did not exist at compensated time")
		return false, fmt.Errorf("attacker entity %d did not exist at compensated time", attackerID)
	}

	// Calculate distance between hit position and target's historical position
	dx := hitPosition.X - targetSnapshot.Position.X
	dy := hitPosition.Y - targetSnapshot.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Hit is valid if within radius
	if distance <= hitRadius {
		logrus.WithFields(logrus.Fields{
			"system_name":      "lag_compensator",
			"attackerID":       attackerID,
			"targetID":         targetID,
			"distance":         distance,
			"hit_radius":       hitRadius,
			"compensated_time": rewind.CompensatedTime.Format(time.RFC3339Nano),
			"latency_ms":       playerLatency.Milliseconds(),
		}).Debug("hit validated")
		return true, nil
	}

	logrus.WithFields(logrus.Fields{
		"system_name":      "lag_compensator",
		"attackerID":       attackerID,
		"targetID":         targetID,
		"distance":         distance,
		"hit_radius":       hitRadius,
		"compensated_time": rewind.CompensatedTime.Format(time.RFC3339Nano),
		"latency_ms":       playerLatency.Milliseconds(),
	}).Debug("hit rejected: outside radius")

	return false, nil
}

// GetEntityPositionAt retrieves an entity's position at a specific time
// in the past. This is useful for debugging or visualization.
func (lc *LagCompensator) GetEntityPositionAt(entityID uint64, t time.Time) (*Position, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	snapshot := lc.snapshots.GetSnapshotAtTime(t)
	if snapshot == nil {
		return nil, fmt.Errorf("no snapshot available at time %v", t)
	}

	entity, exists := snapshot.Entities[entityID]
	if !exists {
		return nil, fmt.Errorf("entity %d not found in snapshot at time %v", entityID, t)
	}

	return &entity.Position, nil
}

// InterpolateEntityAt interpolates an entity's position at a specific
// time, using surrounding snapshots for smooth movement.
func (lc *LagCompensator) InterpolateEntityAt(entityID uint64, t time.Time) (*EntitySnapshot, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	entity := lc.snapshots.InterpolateEntity(entityID, t)
	if entity == nil {
		return nil, fmt.Errorf("failed to interpolate entity %d at time %v", entityID, t)
	}

	return entity, nil
}

// CompensationStats contains statistics about lag compensation.
type CompensationStats struct {
	// TotalSnapshots is the number of snapshots stored
	TotalSnapshots int

	// OldestSnapshotAge is the age of the oldest snapshot
	OldestSnapshotAge time.Duration

	// MaxCompensation is the maximum compensation configured
	MaxCompensation time.Duration

	// MinCompensation is the minimum compensation configured
	MinCompensation time.Duration

	// CurrentSequence is the current snapshot sequence number
	CurrentSequence uint32
}

// GetStats returns statistics about the lag compensation system
func (lc *LagCompensator) GetStats() CompensationStats {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	stats := CompensationStats{
		MaxCompensation: lc.maxCompensation,
		MinCompensation: lc.minCompensation,
		CurrentSequence: lc.snapshots.GetCurrentSequence(),
	}

	// Count valid snapshots and find oldest
	latest := lc.snapshots.GetLatestSnapshot()
	if latest != nil {
		stats.TotalSnapshots = 1
		oldestTime := latest.Timestamp

		// Search for oldest snapshot using wrap-around-safe sequence iteration
		// We look back up to 200 sequences from the latest
		// Using uint32 subtraction naturally handles wrap-around
		for i := uint32(1); i <= 200; i++ {
			// Calculate the sequence to check, handling wrap-around
			checkSeq := latest.Sequence - i

			// Try to get the snapshot at this sequence
			snapshot := lc.snapshots.GetSnapshotAtSequence(checkSeq)
			if snapshot != nil {
				stats.TotalSnapshots++
				if snapshot.Timestamp.Before(oldestTime) {
					oldestTime = snapshot.Timestamp
				}
			}
		}

		stats.OldestSnapshotAge = time.Since(oldestTime)
	}

	return stats
}

// Clear clears all stored snapshots
func (lc *LagCompensator) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.snapshots.Clear()
}
