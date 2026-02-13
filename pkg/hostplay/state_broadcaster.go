// Package hostplay provides dedicated server state synchronization.
package hostplay

import (
	"encoding/json"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// StateBroadcaster manages game state synchronization to clients.
type StateBroadcaster struct {
	world          *engine.World
	broadcastRate  int // Hz (updates per second)
	lastBroadcast  time.Time
	logger         *logrus.Entry
	priorityRadius float64      // Radius for high-priority entities
	timeProvider   TimeProvider // TimeProvider for deterministic timestamps
}

// EntityState represents serialized entity state for network transmission.
type EntityState struct {
	ID       uint64         `json:"id"`
	Position *PositionState `json:"position,omitempty"`
	Velocity *VelocityState `json:"velocity,omitempty"`
	Health   *HealthState   `json:"health,omitempty"`
	Rotation *RotationState `json:"rotation,omitempty"`
}

// PositionState represents entity position.
type PositionState struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// VelocityState represents entity velocity.
type VelocityState struct {
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
}

// HealthState represents entity health.
type HealthState struct {
	Current float64 `json:"current"`
	Max     float64 `json:"max"`
}

// RotationState represents entity rotation.
type RotationState struct {
	Angle float64 `json:"angle"`
}

// WorldState represents the complete world state snapshot.
type WorldState struct {
	Timestamp int64         `json:"timestamp"`
	Entities  []EntityState `json:"entities"`
}

// NewStateBroadcaster creates a new state broadcaster.
func NewStateBroadcaster(world *engine.World, broadcastRate int, logger *logrus.Entry) *StateBroadcaster {
	return NewStateBroadcasterWithTimeProvider(world, broadcastRate, logger, DefaultTimeProvider())
}

// NewStateBroadcasterWithTimeProvider creates a new state broadcaster with a custom time provider.
// Use this constructor in tests to inject a mock TimeProvider for deterministic timestamps.
func NewStateBroadcasterWithTimeProvider(world *engine.World, broadcastRate int, logger *logrus.Entry, tp TimeProvider) *StateBroadcaster {
	return &StateBroadcaster{
		world:          world,
		broadcastRate:  broadcastRate,
		lastBroadcast:  time.Time{}, // Zero time ensures first broadcast happens immediately
		logger:         logger,
		priorityRadius: 1000.0, // Default 1000 pixels
		timeProvider:   tp,
	}
}

// SetPriorityRadius sets the radius for high-priority entities.
func (b *StateBroadcaster) SetPriorityRadius(radius float64) {
	b.priorityRadius = radius
}

// ShouldBroadcast checks if it's time to broadcast based on rate.
func (b *StateBroadcaster) ShouldBroadcast() bool {
	interval := time.Duration(1000/b.broadcastRate) * time.Millisecond
	return b.timeProvider.Now().Sub(b.lastBroadcast) >= interval
}

// CreateSnapshot creates a snapshot of the current world state.
func (b *StateBroadcaster) CreateSnapshot() (*WorldState, error) {
	entities := b.world.GetEntities()

	now := b.timeProvider.Now()
	snapshot := &WorldState{
		Timestamp: now.UnixMilli(),
		Entities:  make([]EntityState, 0, len(entities)),
	}

	for _, entity := range entities {
		entityState := b.serializeEntity(entity)
		if entityState != nil {
			snapshot.Entities = append(snapshot.Entities, *entityState)
		}
	}

	b.lastBroadcast = now

	return snapshot, nil
} // SerializeSnapshot serializes a world state to JSON bytes.
func (b *StateBroadcaster) SerializeSnapshot(snapshot *WorldState) ([]byte, error) {
	return json.Marshal(snapshot)
}

// Broadcast creates and returns a serialized snapshot if it's time to broadcast.
func (b *StateBroadcaster) Broadcast() ([]byte, bool, error) {
	if !b.ShouldBroadcast() {
		return nil, false, nil
	}

	snapshot, err := b.CreateSnapshot()
	if err != nil {
		return nil, false, err
	}

	data, err := b.SerializeSnapshot(snapshot)
	if err != nil {
		return nil, false, err
	}

	b.logger.WithFields(logrus.Fields{
		"entity_count": len(snapshot.Entities),
		"size_bytes":   len(data),
	}).Debug("state broadcast created")

	return data, true, nil
}

// serializeEntity converts an entity to EntityState for transmission.
func (b *StateBroadcaster) serializeEntity(entity *engine.Entity) *EntityState {
	state := &EntityState{
		ID: entity.ID,
	}

	serializePosition(entity, state)
	serializeVelocity(entity, state)
	serializeHealth(entity, state)
	serializeRotation(entity, state)

	// Only include entities with at least position
	if state.Position == nil {
		return nil
	}

	return state
}

// serializePosition extracts position component to state.
func serializePosition(entity *engine.Entity, state *EntityState) {
	if posComp, ok := entity.GetComponent("position"); ok {
		if pos, ok := posComp.(*engine.PositionComponent); ok {
			state.Position = &PositionState{
				X: pos.X,
				Y: pos.Y,
			}
		}
	}
}

// serializeVelocity extracts velocity component to state.
func serializeVelocity(entity *engine.Entity, state *EntityState) {
	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*engine.VelocityComponent); ok {
			state.Velocity = &VelocityState{
				VX: vel.VX,
				VY: vel.VY,
			}
		}
	}
}

// serializeHealth extracts health component to state.
func serializeHealth(entity *engine.Entity, state *EntityState) {
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*engine.HealthComponent); ok {
			state.Health = &HealthState{
				Current: health.Current,
				Max:     health.Max,
			}
		}
	}
}

// serializeRotation extracts rotation component to state.
func serializeRotation(entity *engine.Entity, state *EntityState) {
	if rotComp, ok := entity.GetComponent("rotation"); ok {
		if rot, ok := rotComp.(*engine.RotationComponent); ok {
			state.Rotation = &RotationState{
				Angle: rot.Angle,
			}
		}
	}
}

// CreateDeltaSnapshot creates a delta snapshot containing only entities that have
// changed since the previous snapshot. This reduces network bandwidth by not
// retransmitting unchanged entity state.
//
// If previousSnapshot is nil, returns a full snapshot (all entities).
// An entity is considered changed if:
//   - It is new (not in the previous snapshot)
//   - Its position, velocity, health, or rotation has changed
//
// Removed entities are not explicitly tracked; clients should use full snapshots
// periodically to reconcile state.
func (b *StateBroadcaster) CreateDeltaSnapshot(previousSnapshot *WorldState) (*WorldState, error) {
	// If no previous snapshot, return full snapshot
	if previousSnapshot == nil {
		return b.CreateSnapshot()
	}

	// Build map of previous entity states for fast lookup
	previousStates := make(map[uint64]*EntityState, len(previousSnapshot.Entities))
	for i := range previousSnapshot.Entities {
		previousStates[previousSnapshot.Entities[i].ID] = &previousSnapshot.Entities[i]
	}

	entities := b.world.GetEntities()
	now := b.timeProvider.Now()

	snapshot := &WorldState{
		Timestamp: now.UnixMilli(),
		Entities:  make([]EntityState, 0),
	}

	for _, entity := range entities {
		currentState := b.serializeEntity(entity)
		if currentState == nil {
			continue // Entity has no position, skip
		}

		prevState, exists := previousStates[entity.ID]
		if !exists || entityStateChanged(prevState, currentState) {
			snapshot.Entities = append(snapshot.Entities, *currentState)
		}
	}

	b.lastBroadcast = now

	return snapshot, nil
}

// entityStateChanged compares two entity states and returns true if any
// component has changed.
func entityStateChanged(prev, current *EntityState) bool {
	// Check position changes
	if !positionEqual(prev.Position, current.Position) {
		return true
	}
	// Check velocity changes
	if !velocityEqual(prev.Velocity, current.Velocity) {
		return true
	}
	// Check health changes
	if !healthEqual(prev.Health, current.Health) {
		return true
	}
	// Check rotation changes
	if !rotationEqual(prev.Rotation, current.Rotation) {
		return true
	}
	return false
}

// positionEqual returns true if two position states are equal.
func positionEqual(a, b *PositionState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.X == b.X && a.Y == b.Y
}

// velocityEqual returns true if two velocity states are equal.
func velocityEqual(a, b *VelocityState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.VX == b.VX && a.VY == b.VY
}

// healthEqual returns true if two health states are equal.
func healthEqual(a, b *HealthState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Current == b.Current && a.Max == b.Max
}

// rotationEqual returns true if two rotation states are equal.
func rotationEqual(a, b *RotationState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Angle == b.Angle
}

// GetBroadcastRate returns the current broadcast rate in Hz.
func (b *StateBroadcaster) GetBroadcastRate() int {
	return b.broadcastRate
}

// SetBroadcastRate updates the broadcast rate.
func (b *StateBroadcaster) SetBroadcastRate(rate int) {
	if rate > 0 && rate <= 60 {
		b.broadcastRate = rate
		b.logger.WithField("rate", rate).Debug("broadcast rate updated")
	}
}
