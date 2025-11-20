// Package network provides snapshot management for state history.
// This file implements snapshot buffering and interpolation for smooth
// entity movement and lag compensation.
package network

import (
	"sync"
	"time"
)

// EntitySnapshot represents a snapshot of an entity's state at a specific time
type EntitySnapshot struct {
	EntityID   uint64
	Timestamp  time.Time
	Sequence   uint32
	Position   Position
	Velocity   Velocity
	Components map[string][]byte // Additional component data
}

// WorldSnapshot represents a complete snapshot of the world state
type WorldSnapshot struct {
	Timestamp time.Time
	Sequence  uint32
	Entities  map[uint64]EntitySnapshot
}

// SnapshotManager manages world state snapshots for synchronization
type SnapshotManager struct {
	mu sync.RWMutex

	// Circular buffer of snapshots
	snapshots []WorldSnapshot

	// Current index in the circular buffer
	currentIndex int

	// Maximum number of snapshots to keep
	maxSnapshots int

	// Current sequence number
	currentSeq uint32

	// DeltaCompressionEpsilon is the threshold for detecting entity changes
	// in delta compression. Smaller values = more sensitive to changes = more
	// bandwidth usage. Larger values = less sensitive = better compression but
	// may miss small movements. Typical: 0.001 (default), 0.01 (high-latency).
	DeltaCompressionEpsilon float64
}

// NewSnapshotManager creates a new snapshot manager with default epsilon (0.001).
// For custom epsilon, use NewSnapshotManagerWithEpsilon.
func NewSnapshotManager(maxSnapshots int) *SnapshotManager {
	return NewSnapshotManagerWithEpsilon(maxSnapshots, 0.001)
}

// NewSnapshotManagerWithEpsilon creates a new snapshot manager with custom epsilon.
// The epsilon parameter controls delta compression sensitivity:
// - Smaller values (e.g., 0.001): More sensitive, higher bandwidth, accurate for low-latency
// - Larger values (e.g., 0.01): Less sensitive, lower bandwidth, suitable for high-latency
func NewSnapshotManagerWithEpsilon(maxSnapshots int, epsilon float64) *SnapshotManager {
	if maxSnapshots < 2 {
		maxSnapshots = 2
	}
	if epsilon <= 0 {
		epsilon = 0.001 // Default to prevent division by zero or invalid values
	}

	return &SnapshotManager{
		snapshots:               make([]WorldSnapshot, maxSnapshots),
		currentIndex:            -1,
		maxSnapshots:            maxSnapshots,
		currentSeq:              0,
		DeltaCompressionEpsilon: epsilon,
	}
}

// AddSnapshot adds a new snapshot to the manager
func (sm *SnapshotManager) AddSnapshot(snapshot WorldSnapshot) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentSeq++
	snapshot.Sequence = sm.currentSeq
	snapshot.Timestamp = time.Now()

	// Move to next index in circular buffer
	sm.currentIndex = (sm.currentIndex + 1) % sm.maxSnapshots
	sm.snapshots[sm.currentIndex] = snapshot
}

// GetLatestSnapshot returns the most recent snapshot
func (sm *SnapshotManager) GetLatestSnapshot() *WorldSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentIndex < 0 {
		return nil
	}

	snapshot := sm.snapshots[sm.currentIndex]
	return &snapshot
}

// GetSnapshotAtSequence retrieves a snapshot by sequence number
func (sm *SnapshotManager) GetSnapshotAtSequence(seq uint32) *WorldSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentIndex < 0 {
		return nil
	}

	// Search backwards from current index
	for i := 0; i < sm.maxSnapshots; i++ {
		idx := (sm.currentIndex - i + sm.maxSnapshots) % sm.maxSnapshots
		if sm.snapshots[idx].Sequence == seq {
			snapshot := sm.snapshots[idx]
			return &snapshot
		}

		// Stop if we reach an uninitialized snapshot (sequence 0)
		if sm.snapshots[idx].Sequence == 0 {
			break
		}
	}

	return nil
}

// GetSnapshotAtTime retrieves the snapshot closest to a specific time
func (sm *SnapshotManager) GetSnapshotAtTime(t time.Time) *WorldSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentIndex < 0 {
		return nil
	}

	var closest *WorldSnapshot
	var minDiff time.Duration

	for i := 0; i < sm.maxSnapshots; i++ {
		idx := (sm.currentIndex - i + sm.maxSnapshots) % sm.maxSnapshots
		snapshot := &sm.snapshots[idx]

		// Skip uninitialized snapshots
		if snapshot.Sequence == 0 {
			break
		}

		diff := absTimeDiff(snapshot.Timestamp, t)
		if closest == nil || diff < minDiff {
			closest = snapshot
			minDiff = diff
		}
	}

	return closest
}

// InterpolateEntity interpolates an entity's position between two snapshots
func (sm *SnapshotManager) InterpolateEntity(entityID uint64, renderTime time.Time) *EntitySnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentIndex < 0 {
		return nil
	}

	before, after := sm.findBracketingSnapshots(renderTime)
	return sm.computeInterpolation(entityID, renderTime, before, after)
}

// findBracketingSnapshots finds two snapshots that bracket the render time
func (sm *SnapshotManager) findBracketingSnapshots(renderTime time.Time) (*WorldSnapshot, *WorldSnapshot) {
	var before, after *WorldSnapshot

	for i := 0; i < sm.maxSnapshots-1; i++ {
		idx := (sm.currentIndex - i + sm.maxSnapshots) % sm.maxSnapshots
		snapshot := &sm.snapshots[idx]

		if snapshot.Sequence == 0 {
			break
		}

		if snapshot.Timestamp.Before(renderTime) || snapshot.Timestamp.Equal(renderTime) {
			before = snapshot
			after = sm.findNextSnapshot(idx, renderTime)
			break
		}
	}

	return before, after
}

// findNextSnapshot finds the snapshot after the given index that's after render time
func (sm *SnapshotManager) findNextSnapshot(currentIdx int, renderTime time.Time) *WorldSnapshot {
	nextIdx := (currentIdx + 1) % sm.maxSnapshots
	nextSnapshot := &sm.snapshots[nextIdx]

	if nextSnapshot.Sequence != 0 &&
		(nextSnapshot.Timestamp.After(renderTime) || nextSnapshot.Timestamp.Equal(renderTime)) {
		return nextSnapshot
	}
	return nil
}

// computeInterpolation calculates interpolated entity state between snapshots
func (sm *SnapshotManager) computeInterpolation(entityID uint64, renderTime time.Time, before, after *WorldSnapshot) *EntitySnapshot {
	if before == nil {
		return sm.getLatestEntity(entityID)
	}

	if after == nil {
		return sm.getEntityFromSnapshot(entityID, before)
	}

	return sm.interpolateBetweenSnapshots(entityID, renderTime, before, after)
}

// getLatestEntity retrieves entity from the most recent snapshot
func (sm *SnapshotManager) getLatestEntity(entityID uint64) *EntitySnapshot {
	snapshot := sm.snapshots[sm.currentIndex]
	if entity, exists := snapshot.Entities[entityID]; exists {
		return &entity
	}
	return nil
}

// getEntityFromSnapshot retrieves entity from a specific snapshot
func (sm *SnapshotManager) getEntityFromSnapshot(entityID uint64, snapshot *WorldSnapshot) *EntitySnapshot {
	if entity, exists := snapshot.Entities[entityID]; exists {
		return &entity
	}
	return nil
}

// interpolateBetweenSnapshots performs linear interpolation between two snapshots
func (sm *SnapshotManager) interpolateBetweenSnapshots(entityID uint64, renderTime time.Time, before, after *WorldSnapshot) *EntitySnapshot {
	beforeEntity, beforeExists := before.Entities[entityID]
	afterEntity, afterExists := after.Entities[entityID]

	if !beforeExists && !afterExists {
		return nil
	}
	if !beforeExists {
		return &afterEntity
	}
	if !afterExists {
		return &beforeEntity
	}

	t := sm.calculateInterpolationFactor(renderTime, before, after)
	return sm.createInterpolatedSnapshot(entityID, renderTime, beforeEntity, afterEntity, t)
}

// calculateInterpolationFactor computes normalized time between snapshots
func (sm *SnapshotManager) calculateInterpolationFactor(renderTime time.Time, before, after *WorldSnapshot) float64 {
	totalDuration := after.Timestamp.Sub(before.Timestamp).Seconds()
	if totalDuration <= 0 {
		return 1.0
	}

	elapsed := renderTime.Sub(before.Timestamp).Seconds()
	t := elapsed / totalDuration

	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// createInterpolatedSnapshot builds interpolated entity snapshot from two states
func (sm *SnapshotManager) createInterpolatedSnapshot(entityID uint64, renderTime time.Time, before, after EntitySnapshot, t float64) *EntitySnapshot {
	interpolated := EntitySnapshot{
		EntityID:  entityID,
		Timestamp: renderTime,
		Sequence:  after.Sequence,
		Position: Position{
			X: lerp(before.Position.X, after.Position.X, t),
			Y: lerp(before.Position.Y, after.Position.Y, t),
		},
		Velocity: Velocity{
			VX: lerp(before.Velocity.VX, after.Velocity.VX, t),
			VY: lerp(before.Velocity.VY, after.Velocity.VY, t),
		},
		Components: after.Components,
	}

	return &interpolated
}

// CreateDelta creates a delta between two snapshots (for compression)
func (sm *SnapshotManager) CreateDelta(fromSeq, toSeq uint32) *SnapshotDelta {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	from := sm.GetSnapshotAtSequence(fromSeq)
	to := sm.GetSnapshotAtSequence(toSeq)

	if from == nil || to == nil {
		return nil
	}

	delta := &SnapshotDelta{
		FromSequence: fromSeq,
		ToSequence:   toSeq,
		Added:        make([]uint64, 0),
		Removed:      make([]uint64, 0),
		Changed:      make(map[uint64]EntitySnapshot),
	}

	// Find removed entities
	for entityID := range from.Entities {
		if _, exists := to.Entities[entityID]; !exists {
			delta.Removed = append(delta.Removed, entityID)
		}
	}

	// Find added and changed entities
	for entityID, toEntity := range to.Entities {
		fromEntity, existed := from.Entities[entityID]
		if !existed {
			delta.Added = append(delta.Added, entityID)
			delta.Changed[entityID] = toEntity
		} else if !sm.entityEquals(fromEntity, toEntity) {
			delta.Changed[entityID] = toEntity
		}
	}

	return delta
}

// SnapshotDelta represents the difference between two snapshots
type SnapshotDelta struct {
	FromSequence uint32
	ToSequence   uint32
	Added        []uint64                  // Entity IDs that were added
	Removed      []uint64                  // Entity IDs that were removed
	Changed      map[uint64]EntitySnapshot // Entities that changed
}

// ApplyDelta applies a delta to a snapshot to produce a new snapshot
func (sm *SnapshotManager) ApplyDelta(baseSeq uint32, delta *SnapshotDelta) *WorldSnapshot {
	sm.mu.RLock()
	base := sm.GetSnapshotAtSequence(baseSeq)
	sm.mu.RUnlock()

	if base == nil {
		return nil
	}

	// Create new snapshot
	newSnapshot := WorldSnapshot{
		Timestamp: time.Now(),
		Sequence:  delta.ToSequence,
		Entities:  make(map[uint64]EntitySnapshot),
	}

	// Copy entities from base, excluding removed ones
	for entityID, entity := range base.Entities {
		if !contains(delta.Removed, entityID) {
			newSnapshot.Entities[entityID] = entity
		}
	}

	// Apply changes (includes both changed and added)
	for entityID, entity := range delta.Changed {
		newSnapshot.Entities[entityID] = entity
	}

	return &newSnapshot
}

// GetCurrentSequence returns the current sequence number
func (sm *SnapshotManager) GetCurrentSequence() uint32 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentSeq
}

// Clear removes all snapshots
func (sm *SnapshotManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.snapshots = make([]WorldSnapshot, sm.maxSnapshots)
	sm.currentIndex = -1
	sm.currentSeq = 0
}

// Helper functions

func absTimeDiff(t1, t2 time.Time) time.Duration {
	diff := t1.Sub(t2)
	if diff < 0 {
		return -diff
	}
	return diff
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// entityEquals compares two entity snapshots using the configured epsilon threshold.
// Returns true if the entities are considered equal (differences below epsilon).
func (sm *SnapshotManager) entityEquals(a, b EntitySnapshot) bool {
	epsilon := sm.DeltaCompressionEpsilon
	return abs(a.Position.X-b.Position.X) < epsilon &&
		abs(a.Position.Y-b.Position.Y) < epsilon &&
		abs(a.Velocity.VX-b.Velocity.VX) < epsilon &&
		abs(a.Velocity.VY-b.Velocity.VY) < epsilon
}

func contains(slice []uint64, val uint64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
