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

		// Get player position using the hot-path cached accessor.
		playerPos := player.GetPosition()
		if playerPos == nil {
			continue
		}

		// Get object entity
		object, ok := s.world.GetEntity(objectID)
		if !ok || object == nil {
			// Object no longer exists
			delete(s.carriedObjects, playerID)
			continue
		}

		// Get object position using the hot-path cached accessor.
		objPos := object.GetPosition()
		if objPos == nil {
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

	if s.isPlayerCarrying(playerID) {
		return false
	}

	carriable, ok := s.getCarriableComponent(objectID)
	if !ok {
		return false
	}

	if !s.canPickupObject(carriable) {
		return false
	}

	s.executePickup(playerID, objectID, carriable)
	s.logPickupSuccess(playerID, objectID)

	return true
}

// isPlayerCarrying checks if the player is already carrying an object.
func (s *CarrySystem) isPlayerCarrying(playerID uint64) bool {
	if _, carrying := s.carriedObjects[playerID]; carrying {
		if s.logger != nil {
			s.logger.WithField("playerID", playerID).Debug("player already carrying object")
		}
		return true
	}
	return false
}

// getCarriableComponent retrieves and validates the carriable component from an object.
func (s *CarrySystem) getCarriableComponent(objectID uint64) (*CarriableComponent, bool) {
	object, ok := s.world.GetEntity(objectID)
	if !ok || object == nil {
		return nil, false
	}

	carrComp, ok := object.GetComponent("carriable")
	if !ok {
		return nil, false
	}

	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return nil, false
	}

	return carriable, true
}

// canPickupObject checks if the object is eligible for pickup.
func (s *CarrySystem) canPickupObject(carriable *CarriableComponent) bool {
	return carriable.CanPickUp && !carriable.IsCarried
}

// executePickup performs the pickup operation and clears object velocity.
func (s *CarrySystem) executePickup(playerID, objectID uint64, carriable *CarriableComponent) {
	carriable.Pickup(playerID)
	s.carriedObjects[playerID] = objectID

	object, ok := s.world.GetEntity(objectID)
	if ok && object != nil {
		s.clearObjectVelocity(object)
	}
}

// clearObjectVelocity removes velocity from the picked-up object.
func (s *CarrySystem) clearObjectVelocity(object *Entity) {
	if velComp, ok := object.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			vel.VX = 0
			vel.VY = 0
		}
	}
}

// logPickupSuccess logs successful object pickup.
func (s *CarrySystem) logPickupSuccess(playerID, objectID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"objectID": objectID,
		}).Debug("object picked up")
	}
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
	if !carrying || s.world == nil {
		return
	}

	object, carriable := s.getObjectAndCarriable(playerID, objectID)
	if object == nil || carriable == nil {
		return
	}

	throwVel := s.calculateThrowVelocity(carriable)
	aimX, aimY = s.normalizeAimDirection(aimX, aimY)
	s.applyThrowVelocity(object, aimX, aimY, throwVel)
	s.finalizeThrow(playerID, objectID, carriable, aimX, aimY, throwVel)
}

// getObjectAndCarriable retrieves the object entity and its carriable component.
func (s *CarrySystem) getObjectAndCarriable(playerID, objectID uint64) (*Entity, *CarriableComponent) {
	object, ok := s.world.GetEntity(objectID)
	if !ok || object == nil {
		delete(s.carriedObjects, playerID)
		return nil, nil
	}

	carrComp, ok := object.GetComponent("carriable")
	if !ok {
		return nil, nil
	}

	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return nil, nil
	}

	return object, carriable
}

// calculateThrowVelocity computes the throw velocity based on object weight.
func (s *CarrySystem) calculateThrowVelocity(carriable *CarriableComponent) float64 {
	baseVelocity := 300.0
	return baseVelocity * carriable.ThrowVelocityMultiplier
}

// normalizeAimDirection normalizes the aim direction vector.
func (s *CarrySystem) normalizeAimDirection(aimX, aimY float64) (float64, float64) {
	aimLen := math.Sqrt(aimX*aimX + aimY*aimY)
	if aimLen > 0 {
		return aimX / aimLen, aimY / aimLen
	}
	return 1.0, 0.0
}

// applyThrowVelocity sets the velocity component on the thrown object.
func (s *CarrySystem) applyThrowVelocity(object *Entity, aimX, aimY, throwVel float64) {
	velocityX := aimX * throwVel
	velocityY := aimY * throwVel

	if velComp, ok := object.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			vel.VX = velocityX
			vel.VY = velocityY
			return
		}
	}

	object.AddComponent(&VelocityComponent{
		VX: velocityX,
		VY: velocityY,
	})
}

// finalizeThrow completes the throw operation by dropping the object and logging.
func (s *CarrySystem) finalizeThrow(playerID, objectID uint64, carriable *CarriableComponent, aimX, aimY, throwVel float64) {
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

	closestID, closestDist := s.searchNearestCarriable(x, y, maxDistance)
	if closestID != 0 {
		return closestID, closestDist
	}
	return 0, 0
}

// searchNearestCarriable searches for the nearest carriable object within range.
func (s *CarrySystem) searchNearestCarriable(x, y, maxDistance float64) (uint64, float64) {
	var closestID uint64
	closestDist := maxDistance + 1.0

	entities := s.world.GetEntitiesWith("carriable")
	for _, entity := range entities {
		carriable, pos := s.validateCarriableEntity(entity)
		if carriable == nil || pos == nil {
			continue
		}

		if s.shouldSkipCarriable(carriable) {
			continue
		}

		dist := s.calculateDistance(x, y, pos.X, pos.Y)
		if dist < closestDist {
			closestID = entity.ID
			closestDist = dist
		}
	}

	return closestID, closestDist
}

// validateCarriableEntity validates and extracts carriable and position components.
func (s *CarrySystem) validateCarriableEntity(entity *Entity) (*CarriableComponent, *PositionComponent) {
	carrComp, ok := entity.GetComponent("carriable")
	if !ok {
		return nil, nil
	}
	carriable, ok := carrComp.(*CarriableComponent)
	if !ok {
		return nil, nil
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil, nil
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil
	}

	return carriable, pos
}

// shouldSkipCarriable checks if a carriable object should be excluded from search.
func (s *CarrySystem) shouldSkipCarriable(carriable *CarriableComponent) bool {
	return carriable.IsCarried || !carriable.CanPickUp
}

// calculateDistance computes Euclidean distance between two points.
func (s *CarrySystem) calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
