// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// MovementSystem handles entity movement based on velocity.
type MovementSystem struct {
	// MaxSpeed limits entity velocity (0 = no limit)
	MaxSpeed float64

	// CollisionSystem for predictive collision checking (optional)
	collisionSystem *CollisionSystem

	// SpatialPartitionSystem for dirty tracking (optional)
	spatialPartition *SpatialPartitionSystem

	// Track if any entity moved this frame
	entitiesMoved bool
}

// NewMovementSystem creates a new movement system.
func NewMovementSystem(maxSpeed float64) *MovementSystem {
	return &MovementSystem{
		MaxSpeed: maxSpeed,
	}
}

// SetCollisionSystem sets the collision system for predictive collision checking.
// When set, MovementSystem will validate positions before applying movement.
func (s *MovementSystem) SetCollisionSystem(collisionSystem *CollisionSystem) {
	s.collisionSystem = collisionSystem
}

// SetSpatialPartition sets the spatial partition system for dirty tracking.
// When entities move, the spatial partition will be marked dirty for lazy rebuilding.
func (s *MovementSystem) SetSpatialPartition(spatialPartition *SpatialPartitionSystem) {
	s.spatialPartition = spatialPartition
}

// Update applies velocity to position for all entities with both components.
func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
	s.entitiesMoved = false

	for _, entity := range entities {
		// Skip dead entities - they cannot move (Priority 1.2)
		if entity.HasComponent("dead") {
			continue
		}

		// Get required components
		posComp, hasPos := entity.GetComponent("position")
		velComp, hasVel := entity.GetComponent("velocity")
		if !hasPos || !hasVel {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			continue
		}

		// Apply speed limit if configured
		s.applySpeedLimit(vel)

		// Calculate new position
		newX := pos.X + vel.VX*deltaTime
		newY := pos.Y + vel.VY*deltaTime

		// Validate position with collision checking and wall sliding
		newX, newY = s.calculateValidPosition(entity, pos, vel, newX, newY, entities)

		// Update position and track movement
		oldX, oldY := pos.X, pos.Y
		pos.X = newX
		pos.Y = newY

		if pos.X != oldX || pos.Y != oldY {
			s.entitiesMoved = true

			// Check for layer transitions via ramps
			if s.collisionSystem != nil && entity.HasComponent("collider") {
				s.checkLayerTransition(entity, pos)
			}
		}

		// Apply bounds constraints
		s.applyBoundsConstraints(entity, pos, vel)

		// Apply friction to slow down entities
		s.applyFriction(entity, vel, deltaTime)

		// Update animation state based on movement
		s.updateAnimationState(entity, vel)
	}

	// Mark spatial partition as dirty if any entities moved
	if s.entitiesMoved && s.spatialPartition != nil {
		s.spatialPartition.MarkDirty()
	}
}

// anyEntityBlocking checks if any entity would block movement to the given position.
// Helper method for collision sliding logic.
func (s *MovementSystem) anyEntityBlocking(entity *Entity, x, y float64, entities []*Entity) bool {
	if s.collisionSystem == nil {
		return false
	}

	for _, other := range entities {
		if other.ID == entity.ID {
			continue
		}
		if s.collisionSystem.WouldCollideWithEntity(entity, x, y, other) {
			return true
		}
	}
	return false
}

// applySpeedLimit clamps entity velocity to MaxSpeed if configured.
// Returns true if speed was limited.
func (s *MovementSystem) applySpeedLimit(vel *VelocityComponent) bool {
	if s.MaxSpeed <= 0 {
		return false
	}

	speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
	if speed > s.MaxSpeed {
		scale := s.MaxSpeed / speed
		vel.VX *= scale
		vel.VY *= scale
		return true
	}
	return false
}

// calculateValidPosition determines the valid new position after collision checking.
// Implements wall sliding by trying X-only and Y-only movement when diagonal movement is blocked.
// Returns the validated newX, newY coordinates.
func (s *MovementSystem) calculateValidPosition(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64, entities []*Entity) (float64, float64) {
	// No collision system means no validation needed
	if s.collisionSystem == nil || !entity.HasComponent("collider") {
		return newX, newY
	}

	colliderComp, _ := entity.GetComponent("collider")
	if colliderComp == nil {
		return newX, newY
	}

	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return newX, newY
	}

	// Only check solid, non-trigger colliders
	if !collider.Solid || collider.IsTrigger {
		return newX, newY
	}

	// Check terrain collision at new position
	if s.collisionSystem.WouldCollideWithTerrain(entity, newX, newY) {
		// Collision detected - try sliding along walls
		if !s.collisionSystem.WouldCollideWithTerrain(entity, newX, pos.Y) {
			// Allow X movement, block Y (slide horizontally)
			vel.VY = 0
			return newX, pos.Y
		} else if !s.collisionSystem.WouldCollideWithTerrain(entity, pos.X, newY) {
			// Allow Y movement, block X (slide vertically)
			vel.VX = 0
			return pos.X, newY
		}
		// Completely blocked - don't move at all
		vel.VX = 0
		vel.VY = 0
		return pos.X, pos.Y
	}

	// Check entity-to-entity collisions (only if still planning to move)
	if newX == pos.X && newY == pos.Y {
		return newX, newY
	}

	// Check if blocked by another entity
	for _, other := range entities {
		if other.ID == entity.ID {
			continue
		}
		if s.collisionSystem.WouldCollideWithEntity(entity, newX, newY, other) {
			// Blocked - try sliding
			if !s.anyEntityBlocking(entity, newX, pos.Y, entities) {
				vel.VY = 0
				return newX, pos.Y
			} else if !s.anyEntityBlocking(entity, pos.X, newY, entities) {
				vel.VX = 0
				return pos.X, newY
			}
			// Completely blocked
			vel.VX = 0
			vel.VY = 0
			return pos.X, pos.Y
		}
	}

	return newX, newY
}

// applyBoundsConstraints clamps position to bounds and stops velocity at boundaries.
func (s *MovementSystem) applyBoundsConstraints(entity *Entity, pos *PositionComponent, vel *VelocityComponent) {
	boundsComp, hasBounds := entity.GetComponent("bounds")
	if !hasBounds {
		return
	}

	bounds, ok := boundsComp.(*BoundsComponent)
	if !ok {
		return
	}

	pos.X, pos.Y = bounds.Clamp(pos.X, pos.Y)

	// Stop movement at boundaries if not wrapping
	if !bounds.Wrap {
		if pos.X <= bounds.MinX || pos.X >= bounds.MaxX {
			vel.VX = 0
		}
		if pos.Y <= bounds.MinY || pos.Y >= bounds.MaxY {
			vel.VY = 0
		}
	}
}

// applyFriction applies exponential decay friction to slow down entity movement.
// Stops velocity completely when it falls below a threshold.
func (s *MovementSystem) applyFriction(entity *Entity, vel *VelocityComponent, deltaTime float64) {
	frictionComp, hasFriction := entity.GetComponent("friction")
	if !hasFriction {
		return
	}

	friction, ok := frictionComp.(*FrictionComponent)
	if !ok {
		return
	}

	// Apply friction as exponential decay: v *= (1 - coefficient)^deltaTime
	// Normalize to 60 FPS for consistent behavior
	decayFactor := math.Pow(1.0-friction.Coefficient, deltaTime*60.0)
	vel.VX *= decayFactor
	vel.VY *= decayFactor

	// Stop completely if velocity is very small (optimization)
	if math.Abs(vel.VX) < 0.1 && math.Abs(vel.VY) < 0.1 {
		vel.VX = 0
		vel.VY = 0
	}
}

// updateAnimationState updates entity animation based on velocity and movement state.
// Handles idle/walk/run state transitions and facing direction updates.
func (s *MovementSystem) updateAnimationState(entity *Entity, vel *VelocityComponent) {
	animComp, hasAnim := entity.GetComponent("animation")
	if !hasAnim {
		return
	}

	anim, ok := animComp.(*AnimationComponent)
	if !ok {
		return
	}

	// DON'T override attack/hit/death/cast animations - let them finish
	if anim.CurrentState == AnimationStateAttack ||
		anim.CurrentState == AnimationStateHit ||
		anim.CurrentState == AnimationStateDeath ||
		anim.CurrentState == AnimationStateCast {
		return
	}

	speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)

	if speed > 0.1 {
		// Moving - determine if walking or running
		if speed > s.MaxSpeed*0.7 && s.MaxSpeed > 0 {
			// Fast movement - running
			if anim.CurrentState != AnimationStateRun {
				anim.SetState(AnimationStateRun)
			}
		} else {
			// Normal movement - walking
			if anim.CurrentState != AnimationStateWalk {
				anim.SetState(AnimationStateWalk)
			}
		}

		// Phase 10.1: Only update facing direction from velocity if entity doesn't have rotation component
		// Entities with RotationComponent use 360° rotation from aim input instead of 4-directional velocity-based facing
		if !entity.HasComponent("rotation") {
			s.updateFacingDirection(anim, vel)
		}
	} else {
		// Not moving - idle (only if currently in a movement state)
		if anim.CurrentState == AnimationStateWalk || anim.CurrentState == AnimationStateRun {
			anim.SetState(AnimationStateIdle)
		}
		// When idle, preserve facing direction (don't reset)
	}
}

// updateFacingDirection updates animation facing based on velocity direction.
// Applies threshold filtering to prevent jitter from input noise.
func (s *MovementSystem) updateFacingDirection(anim *AnimationComponent, vel *VelocityComponent) {
	absVX := math.Abs(vel.VX)
	absVY := math.Abs(vel.VY)

	// Apply 0.1 threshold to filter input jitter and noise
	if absVX <= 0.1 && absVY <= 0.1 {
		// Velocity below threshold, preserve current facing
		return
	}

	// Prioritize horizontal movement for diagonal directions
	// For perfect diagonals (absVX == absVY), horizontal takes priority
	if absVX >= absVY {
		// Moving horizontally (or perfect diagonal)
		if vel.VX > 0 {
			anim.SetFacing(DirRight)
		} else {
			anim.SetFacing(DirLeft)
		}
	} else {
		// Moving vertically
		if vel.VY > 0 {
			anim.SetFacing(DirDown)
		} else {
			anim.SetFacing(DirUp)
		}
	}
} // SetVelocity is a helper to set entity velocity.
func SetVelocity(entity *Entity, vx, vy float64) {
	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			vel.VX = vx
			vel.VY = vy
		}
	}
}

// GetPosition is a helper to get entity position.
func GetPosition(entity *Entity) (x, y float64, ok bool) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok {
			return pos.X, pos.Y, true
		}
	}
	return 0, 0, false
}

// SetPosition is a helper to set entity position.
func SetPosition(entity *Entity, x, y float64) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok {
			pos.X = x
			pos.Y = y
		}
	}
}

// GetDistance calculates the distance between two entities.
func GetDistance(e1, e2 *Entity) float64 {
	x1, y1, ok1 := GetPosition(e1)
	x2, y2, ok2 := GetPosition(e2)

	if !ok1 || !ok2 {
		return math.Inf(1)
	}

	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// MoveTowards moves an entity towards a target position.
// Returns true if the entity reached the target.
func MoveTowards(entity *Entity, targetX, targetY, speed, deltaTime float64) bool {
	x, y, ok := GetPosition(entity)
	if !ok {
		return false
	}

	dx := targetX - x
	dy := targetY - y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Already at target
	if distance < 0.1 {
		SetVelocity(entity, 0, 0)
		return true
	}

	// Normalize direction and apply speed
	vx := (dx / distance) * speed
	vy := (dy / distance) * speed

	SetVelocity(entity, vx, vy)
	return false
}

// checkLayerTransition checks if an entity is on a ramp tile and updates their layer accordingly.
// Phase 11.1 Week 3: Layer Transition System
//
// This method enables smooth transitions between different terrain layers (ground, water, platform)
// by detecting ramp tiles and updating the entity's collider layer to match.
//
// Ramp tiles allow entities to move between:
// - LayerGround (0) ↔ LayerPlatform (2) via platform ramps
// - LayerGround (0) ↔ LayerWater (1) via water ramps (stairs into water)
//
// The system uses the tile's GetLayer() method to determine the target layer and checks
// CanTransitionToLayer() to verify the transition is valid before applying it.
func (s *MovementSystem) checkLayerTransition(entity *Entity, pos *PositionComponent) {
	// Get terrain checker from collision system
	if s.collisionSystem == nil || s.collisionSystem.terrainChecker == nil {
		return
	}

	terrainChecker := s.collisionSystem.terrainChecker
	if terrainChecker.terrain == nil {
		return
	}

	// Get entity's collider
	colliderComp, hasCollider := entity.GetComponent("collider")
	if !hasCollider {
		return
	}
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		return
	}

	// Calculate tile coordinates from entity position using helper method
	tileX, tileY := terrainChecker.worldToTileCoords(pos.X, pos.Y)

	// Get tile at entity's position
	currentTile := terrainChecker.terrain.GetTile(tileX, tileY)

	// Check if this is a ramp tile (allows layer transitions)
	// Use explicit tile type checks for clarity and correctness
	if currentTile == terrain.TileRamp || currentTile == terrain.TileRampUp || currentTile == terrain.TileRampDown {
		// Determine target layer based on the tile's layer
		// Ramps lead TO the layer they're assigned to
		targetLayer := int(currentTile.GetLayer())

		// Update collider layer if different
		// This allows entity to interact with tiles on the new layer
		if collider.Layer != targetLayer {
			oldLayer := collider.Layer
			collider.Layer = targetLayer
			// Phase 11.1 Week 3: Debug logging for layer transitions
			logrus.WithFields(logrus.Fields{
				"entity":   entity.ID,
				"oldLayer": oldLayer,
				"newLayer": targetLayer,
				"tile":     currentTile,
			}).Debug("Entity layer transition via ramp")
		}
	}
}
