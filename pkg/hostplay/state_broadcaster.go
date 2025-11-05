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
	priorityRadius float64 // Radius for high-priority entities
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
	return &StateBroadcaster{
		world:          world,
		broadcastRate:  broadcastRate,
		lastBroadcast:  time.Now(),
		logger:         logger,
		priorityRadius: 1000.0, // Default 1000 pixels
	}
}

// SetPriorityRadius sets the radius for high-priority entities.
func (b *StateBroadcaster) SetPriorityRadius(radius float64) {
	b.priorityRadius = radius
}

// ShouldBroadcast checks if it's time to broadcast based on rate.
func (b *StateBroadcaster) ShouldBroadcast() bool {
	interval := time.Duration(1000/b.broadcastRate) * time.Millisecond
	return time.Since(b.lastBroadcast) >= interval
}

// CreateSnapshot creates a snapshot of the current world state.
func (b *StateBroadcaster) CreateSnapshot() (*WorldState, error) {
	entities := b.world.GetEntities()

	snapshot := &WorldState{
		Timestamp: time.Now().UnixMilli(),
		Entities:  make([]EntityState, 0, len(entities)),
	}

	for _, entity := range entities {
		entityState := b.serializeEntity(entity)
		if entityState != nil {
			snapshot.Entities = append(snapshot.Entities, *entityState)
		}
	}

	b.lastBroadcast = time.Now()

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

	// Serialize position
	if posComp, ok := entity.GetComponent("position"); ok {
		if pos, ok := posComp.(*engine.PositionComponent); ok {
			state.Position = &PositionState{
				X: pos.X,
				Y: pos.Y,
			}
		}
	}

	// Serialize velocity
	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*engine.VelocityComponent); ok {
			state.Velocity = &VelocityState{
				VX: vel.VX,
				VY: vel.VY,
			}
		}
	}

	// Serialize health
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*engine.HealthComponent); ok {
			state.Health = &HealthState{
				Current: health.Current,
				Max:     health.Max,
			}
		}
	}

	// Serialize rotation
	if rotComp, ok := entity.GetComponent("rotation"); ok {
		if rot, ok := rotComp.(*engine.RotationComponent); ok {
			state.Rotation = &RotationState{
				Angle: rot.Angle,
			}
		}
	}

	// Only include entities with at least position
	if state.Position == nil {
		return nil
	}

	return state
}

// CreateDeltaSnapshot creates a delta snapshot (only changed entities).
// This is an optimization for future use.
func (b *StateBroadcaster) CreateDeltaSnapshot(previousSnapshot *WorldState) (*WorldState, error) {
	// For now, just return a full snapshot
	// Future: implement delta compression by comparing with previous state
	return b.CreateSnapshot()
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
