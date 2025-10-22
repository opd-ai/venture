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
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(maxSnapshots int) *SnapshotManager {
	if maxSnapshots < 2 {
		maxSnapshots = 2
	}

	return &SnapshotManager{
		snapshots:    make([]WorldSnapshot, maxSnapshots),
		currentIndex: -1,
		maxSnapshots: maxSnapshots,
		currentSeq:   0,
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

	// Find two snapshots that bracket the render time
	var before, after *WorldSnapshot

	for i := 0; i < sm.maxSnapshots-1; i++ {
		idx := (sm.currentIndex - i + sm.maxSnapshots) % sm.maxSnapshots
		snapshot := &sm.snapshots[idx]

		if snapshot.Sequence == 0 {
			break
		}

		if snapshot.Timestamp.Before(renderTime) || snapshot.Timestamp.Equal(renderTime) {
			before = snapshot
			// Look for the next snapshot after this one
			nextIdx := (idx + 1) % sm.maxSnapshots
			if sm.snapshots[nextIdx].Sequence != 0 && 
			   (sm.snapshots[nextIdx].Timestamp.After(renderTime) || sm.snapshots[nextIdx].Timestamp.Equal(renderTime)) {
				after = &sm.snapshots[nextIdx]
			}
			break
		}
	}

	// If we don't have bracketing snapshots, return the latest
	if before == nil {
		snapshot := sm.snapshots[sm.currentIndex]
		if entity, exists := snapshot.Entities[entityID]; exists {
			return &entity
		}
		return nil
	}

	// If we only have before, use it
	if after == nil {
		if entity, exists := before.Entities[entityID]; exists {
			return &entity
		}
		return nil
	}

	// Interpolate between before and after
	beforeEntity, beforeExists := before.Entities[entityID]
	afterEntity, afterExists := after.Entities[entityID]

	if !beforeExists || !afterExists {
		// Entity doesn't exist in one of the snapshots
		if beforeExists {
			return &beforeEntity
		}
		if afterExists {
			return &afterEntity
		}
		return nil
	}

	// Calculate interpolation factor (0.0 to 1.0)
	totalDuration := after.Timestamp.Sub(before.Timestamp).Seconds()
	if totalDuration <= 0 {
		return &afterEntity
	}

	elapsed := renderTime.Sub(before.Timestamp).Seconds()
	t := elapsed / totalDuration

	// Clamp t to [0, 1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Interpolate position
	interpolated := EntitySnapshot{
		EntityID:  entityID,
		Timestamp: renderTime,
		Sequence:  afterEntity.Sequence,
		Position: Position{
			X: lerp(beforeEntity.Position.X, afterEntity.Position.X, t),
			Y: lerp(beforeEntity.Position.Y, afterEntity.Position.Y, t),
		},
		Velocity: Velocity{
			VX: lerp(beforeEntity.Velocity.VX, afterEntity.Velocity.VX, t),
			VY: lerp(beforeEntity.Velocity.VY, afterEntity.Velocity.VY, t),
		},
		Components: afterEntity.Components, // Use latest component data
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
		} else if !entityEquals(fromEntity, toEntity) {
			delta.Changed[entityID] = toEntity
		}
	}

	return delta
}

// SnapshotDelta represents the difference between two snapshots
type SnapshotDelta struct {
	FromSequence uint32
	ToSequence   uint32
	Added        []uint64                    // Entity IDs that were added
	Removed      []uint64                    // Entity IDs that were removed
	Changed      map[uint64]EntitySnapshot   // Entities that changed
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

func entityEquals(a, b EntitySnapshot) bool {
	const epsilon = 0.001
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
