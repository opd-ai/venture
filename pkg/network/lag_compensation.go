// Package network provides lag compensation for fair hit detection.
// This file implements lag compensation using snapshot history to rewind
// game state for accurate hit detection on the server.
package network

import (
	"fmt"
	"math"
	"sync"
	"time"
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
}

// DefaultLagCompensationConfig returns a default configuration
// for typical internet play (up to 500ms latency)
func DefaultLagCompensationConfig() LagCompensationConfig {
	return LagCompensationConfig{
		MaxCompensation:    500 * time.Millisecond,
		MinCompensation:    10 * time.Millisecond,
		SnapshotBufferSize: 100,
	}
}

// HighLatencyLagCompensationConfig returns a configuration
// for high-latency connections (e.g., Tor, up to 5000ms)
func HighLatencyLagCompensationConfig() LagCompensationConfig {
	return LagCompensationConfig{
		MaxCompensation:    5000 * time.Millisecond,
		MinCompensation:    10 * time.Millisecond,
		SnapshotBufferSize: 200,
	}
}

// NewLagCompensator creates a new lag compensator with the given configuration
func NewLagCompensator(config LagCompensationConfig) *LagCompensator {
	return &LagCompensator{
		snapshots:       NewSnapshotManager(config.SnapshotBufferSize),
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
		return false, fmt.Errorf("failed to rewind to player time: no snapshot available")
	}

	// Check if target existed at that time
	targetSnapshot, exists := rewind.Snapshot.Entities[targetID]
	if !exists {
		return false, fmt.Errorf("target entity %d did not exist at compensated time", targetID)
	}

	// Check if attacker existed at that time
	_, exists = rewind.Snapshot.Entities[attackerID]
	if !exists {
		return false, fmt.Errorf("attacker entity %d did not exist at compensated time", attackerID)
	}

	// Calculate distance between hit position and target's historical position
	dx := hitPosition.X - targetSnapshot.Position.X
	dy := hitPosition.Y - targetSnapshot.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Hit is valid if within radius
	if distance <= hitRadius {
		return true, nil
	}

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

		// Search for oldest snapshot
		for seq := latest.Sequence - 1; seq > 0 && seq >= latest.Sequence-200; seq-- {
			snapshot := lc.snapshots.GetSnapshotAtSequence(seq)
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
