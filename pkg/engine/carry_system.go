// Package engine provides the carry system for Phase 11.3.
// Environmental Destruction & Manipulation
//
// This file implements CarrySystem which handles pickup and throw mechanics
// for carriable objects. Players can pick up objects with F key, carry them
// (object follows player), and throw them with attack button.
//
// Design Philosophy:
// - One object carried at a time per player
// - Carried objects follow player position (with small offset)
// - Throwing uses aim direction and weight-based velocity
// - Thrown objects deal impact damage on collision
// - Server-authoritative for multiplayer synchronization
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// CarrySystem manages pickup and throw mechanics for carriable objects.
type CarrySystem struct {
	world  *World
	logger *logrus.Entry

	// Track which player is carrying which object
	carriedObjects map[uint64]uint64 // playerID -> objectEntityID
}

// NewCarrySystem creates a new carry system.
func NewCarrySystem() *CarrySystem {
	return NewCarrySystemWithLogger(nil)
}

// NewCarrySystemWithLogger creates a system with a logger.
func NewCarrySystemWithLogger(logger *logrus.Logger) *CarrySystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "carry",
		})
		logEntry.Debug("carry system created")
	}

	return &CarrySystem{
		logger:         logEntry,
		carriedObjects: make(map[uint64]uint64),
	}
}

// SetWorld sets the ECS world reference.
func (s *CarrySystem) SetWorld(world *World) {
	s.world = world
}

// Update implements the System interface.
// Updates carried object positions to follow player.
func (s *CarrySystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Update positions of carried objects
	for playerID, objectID := range s.carriedObjects {
		// Get player entity
		player, ok := s.world.GetEntity(playerID)
		if !ok || player == nil {
			// Player no longer exists, drop object
			s.dropObject(objectID)
			continue
		}

		// Get player position
		playerPosComp, ok := player.GetComponent("position")
		if !ok {
			continue
		}
		playerPos, ok := playerPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Get object entity
		object, ok := s.world.GetEntity(objectID)
		if !ok || object == nil {
			// Object no longer exists
			delete(s.carriedObjects, playerID)
			continue
		}

		// Get object position
		objPosComp, ok := object.GetComponent("position")
		if !ok {
			continue
		}
		objPos, ok := objPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Update object position to follow player (slightly offset above/in front)
		objPos.X = playerPos.X
		objPos.Y = playerPos.Y - 16.0 // Offset above player
	}
}

// TryPickup attempts to pick up a carriable object.
// Returns true if pickup was successful.
func (s *CarrySystem) TryPickup(playerID, objectID uint64) bool {
	if s.world == nil {
		return false
	}

	// Check if player is already carrying something
	if _, carrying := s.carriedObjects[playerID]; carrying {
		if s.logger != nil {
			s.logger.WithField("playerID", playerID).Debug("player already carrying object")
		}
		return false
	}

	// Get object entity
	object, ok := s.world.GetEntity(objectID)
	if !ok || object == nil {
		return false
	}

	// Get carriable component
	carrComp, ok := object.GetComponent("carriable")
	if !ok {
		return false
	}
	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return false
	}

	// Check if object can be picked up
	if !carriable.CanPickUp || carriable.IsCarried {
		return false
	}

	// Pickup object
	carriable.Pickup(playerID)
	s.carriedObjects[playerID] = objectID

	// Remove velocity if object was moving
	if velComp, ok := object.GetComponent("velocity"); ok {
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			return true // Object picked up, velocity just won't be zeroed
		}
		vel.VX = 0
		vel.VY = 0
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"objectID": objectID,
		}).Debug("object picked up")
	}

	return true
}

// DropObject drops the object the player is carrying.
func (s *CarrySystem) DropObject(playerID uint64) {
	objectID, carrying := s.carriedObjects[playerID]
	if !carrying {
		return
	}

	s.dropObject(objectID)
	delete(s.carriedObjects, playerID)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"objectID": objectID,
		}).Debug("object dropped")
	}
}

// ThrowObject throws the carried object in the player's aim direction.
func (s *CarrySystem) ThrowObject(playerID uint64, aimX, aimY float64) {
	objectID, carrying := s.carriedObjects[playerID]
	if !carrying {
		return
	}

	if s.world == nil {
		return
	}

	// Get object entity
	object, ok := s.world.GetEntity(objectID)
	if !ok || object == nil {
		delete(s.carriedObjects, playerID)
		return
	}

	// Get carriable component
	carrComp, ok := object.GetComponent("carriable")
	if !ok {
		return
	}
	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return
	}

	// Calculate throw velocity based on weight
	baseVelocity := 300.0 // pixels per second
	throwVel := baseVelocity * carriable.ThrowVelocityMultiplier

	// Normalize aim direction
	aimLen := math.Sqrt(aimX*aimX + aimY*aimY)
	if aimLen > 0 {
		aimX /= aimLen
		aimY /= aimLen
	} else {
		// Default to right if no aim direction
		aimX = 1.0
		aimY = 0.0
	}

	// Set velocity
	if velComp, ok := object.GetComponent("velocity"); ok {
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			// Add velocity component if type assertion fails
			velComp := &VelocityComponent{
				VX: aimX * throwVel,
				VY: aimY * throwVel,
			}
			object.AddComponent(velComp)
			return
		}
		vel.VX = aimX * throwVel
		vel.VY = aimY * throwVel
	} else {
		// Add velocity component if not present
		velComp := &VelocityComponent{
			VX: aimX * throwVel,
			VY: aimY * throwVel,
		}
		object.AddComponent(velComp)
	}

	// Mark as not carried
	carriable.Drop()
	delete(s.carriedObjects, playerID)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID":  playerID,
			"objectID":  objectID,
			"velocityX": aimX * throwVel,
			"velocityY": aimY * throwVel,
		}).Debug("object thrown")
	}
}

// dropObject drops an object (internal helper).
func (s *CarrySystem) dropObject(objectID uint64) {
	if s.world == nil {
		return
	}

	object, ok := s.world.GetEntity(objectID)
	if !ok || object == nil {
		return
	}

	// Get carriable component
	carrComp, ok := object.GetComponent("carriable")
	if !ok {
		return
	}
	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return
	}

	// Mark as not carried
	carriable.Drop()
}

// IsCarrying returns true if the player is carrying an object.
func (s *CarrySystem) IsCarrying(playerID uint64) bool {
	_, carrying := s.carriedObjects[playerID]
	return carrying
}

// GetCarriedObject returns the entity ID of the object being carried, or 0 if none.
func (s *CarrySystem) GetCarriedObject(playerID uint64) uint64 {
	objectID, carrying := s.carriedObjects[playerID]
	if !carrying {
		return 0
	}
	return objectID
}

// FindNearbyCarriableObject finds a carriable object near the given position.
// Returns the entity ID and distance, or 0 if none found.
func (s *CarrySystem) FindNearbyCarriableObject(x, y, maxDistance float64) (uint64, float64) {
	if s.world == nil {
		return 0, 0
	}

	var closestID uint64
	closestDist := maxDistance + 1.0

	// Find all carriable objects
	entities := s.world.GetEntitiesWith("carriable")
	for _, entity := range entities {
		// Get carriable component
		carrComp, ok := entity.GetComponent("carriable")
		if !ok {
			continue
		}
		carriable, ok := carrComp.(*CarriableComponent)
		if !ok {
			continue
		}

		// Skip if already carried or not pickupable
		if carriable.IsCarried || !carriable.CanPickUp {
			continue
		}

		// Get position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance
		dx := pos.X - x
		dy := pos.Y - y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Check if closer than current closest
		if dist < closestDist {
			closestID = entity.ID
			closestDist = dist
		}
	}

	if closestID != 0 {
		return closestID, closestDist
	}
	return 0, 0
}
