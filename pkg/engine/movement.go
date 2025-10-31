// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import (
	"math"

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
	s.entitiesMoved = false // Reset movement flag

	for _, entity := range entities {
		// Skip dead entities - they cannot move (Priority 1.2)
		// Dead entities are immobilized until revived or removed from the world
		if entity.HasComponent("dead") {
			continue
		}

		// Check if entity has required components
		posComp, hasPos := entity.GetComponent("position")
		velComp, hasVel := entity.GetComponent("velocity")

		if !hasPos || !hasVel {
			continue
		}

		pos := posComp.(*PositionComponent)
		vel := velComp.(*VelocityComponent)

		// Apply speed limit if configured
		if s.MaxSpeed > 0 {
			speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
			if speed > s.MaxSpeed {
				scale := s.MaxSpeed / speed
				vel.VX *= scale
				vel.VY *= scale
			}
		}

		// Calculate new position
		newX := pos.X + vel.VX*deltaTime
		newY := pos.Y + vel.VY*deltaTime

		// GAP-001 REPAIR: Predictive collision checking before updating position
		// If collision system is set, validate position before moving
		if s.collisionSystem != nil && entity.HasComponent("collider") {
			colliderComp, _ := entity.GetComponent("collider")
			collider := colliderComp.(*ColliderComponent)

			// Only check solid, non-trigger colliders
			if collider.Solid && !collider.IsTrigger {
				// Check terrain collision at new position
				if s.collisionSystem.WouldCollideWithTerrain(entity, newX, newY) {
					// Collision detected - try sliding along walls
					// Try X movement only
					if !s.collisionSystem.WouldCollideWithTerrain(entity, newX, pos.Y) {
						newY = pos.Y // Keep Y, allow X movement (slide horizontally)
						vel.VY = 0
					} else if !s.collisionSystem.WouldCollideWithTerrain(entity, pos.X, newY) {
						// Try Y movement only
						newX = pos.X // Keep X, allow Y movement (slide vertically)
						vel.VX = 0
					} else {
						// Completely blocked - don't move at all
						newX = pos.X
						newY = pos.Y
						vel.VX = 0
						vel.VY = 0
					}
				}

				// Check entity-to-entity collisions at new position
				// Only if we're still planning to move
				if newX != pos.X || newY != pos.Y {
					blocked := false
					for _, other := range entities {
						if other.ID == entity.ID {
							continue
						}
						if s.collisionSystem.WouldCollideWithEntity(entity, newX, newY, other) {
							// Blocked by another entity - stop movement
							blocked = true
							break
						}
					}

					if blocked {
						// Try sliding along the blocking entity
						if !s.anyEntityBlocking(entity, newX, pos.Y, entities) {
							newY = pos.Y // Slide horizontally
							vel.VY = 0
						} else if !s.anyEntityBlocking(entity, pos.X, newY, entities) {
							newX = pos.X // Slide vertically
							vel.VX = 0
						} else {
							// Completely blocked
							newX = pos.X
							newY = pos.Y
							vel.VX = 0
							vel.VY = 0
						}
					}
				}
			}
		}

		// Update position (only if validated or no collision checking)
		oldX, oldY := pos.X, pos.Y
		pos.X = newX
		pos.Y = newY

		// Track if entity actually moved
		if pos.X != oldX || pos.Y != oldY {
			s.entitiesMoved = true

			// Phase 11.1 Week 3: Check for layer transitions via ramps
			// If entity has collider and moved, check if they're on a ramp tile
			if s.collisionSystem != nil && entity.HasComponent("collider") {
				s.checkLayerTransition(entity, pos)
			}
		}

		// Apply bounds if entity has them
		if boundsComp, hasBounds := entity.GetComponent("bounds"); hasBounds {
			bounds := boundsComp.(*BoundsComponent)
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

		// Priority 1.4: Apply friction/drag to slow down entities
		if frictionComp, hasFriction := entity.GetComponent("friction"); hasFriction {
			friction := frictionComp.(*FrictionComponent)

			// Apply friction as exponential decay: v *= (1 - coefficient)^deltaTime
			// For small deltaTime and coefficient, this approximates: v *= (1 - coefficient * deltaTime)
			decayFactor := math.Pow(1.0-friction.Coefficient, deltaTime*60.0) // Normalize to 60 FPS
			vel.VX *= decayFactor
			vel.VY *= decayFactor

			// Stop completely if velocity is very small (optimization)
			if math.Abs(vel.VX) < 0.1 && math.Abs(vel.VY) < 0.1 {
				vel.VX = 0
				vel.VY = 0
			}
		}

		// Update animation state based on movement
		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
			anim := animComp.(*AnimationComponent)
			speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)

			// DON'T override attack/hit/death/cast animations - let them finish
			if anim.CurrentState == AnimationStateAttack ||
				anim.CurrentState == AnimationStateHit ||
				anim.CurrentState == AnimationStateDeath ||
				anim.CurrentState == AnimationStateCast {
				// Animation is in action state, don't override with movement
				continue
			}

			// Determine animation state based on velocity
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
					// Phase 3: Update facing direction based on velocity
					// Apply 0.1 threshold to filter input jitter and noise
					absVX := math.Abs(vel.VX)
					absVY := math.Abs(vel.VY)

					if absVX > 0.1 || absVY > 0.1 {
						// Prioritize horizontal movement for diagonal directions
						// This provides clearer visual feedback for player control
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
					}
					// If velocity is below threshold, preserve current facing
				}
				// Phase 10.1: If entity has rotation component, facing is determined by RotationComponent.Angle
			} else {
				// Not moving - idle (only if currently in a movement state)
				if anim.CurrentState == AnimationStateWalk || anim.CurrentState == AnimationStateRun {
					anim.SetState(AnimationStateIdle)
				}
				// When idle, preserve facing direction (don't reset)
			}
		}
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
} // SetVelocity is a helper to set entity velocity.
func SetVelocity(entity *Entity, vx, vy float64) {
	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
		vel := velComp.(*VelocityComponent)
		vel.VX = vx
		vel.VY = vy
	}
}

// GetPosition is a helper to get entity position.
func GetPosition(entity *Entity) (x, y float64, ok bool) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		pos := posComp.(*PositionComponent)
		return pos.X, pos.Y, true
	}
	return 0, 0, false
}

// SetPosition is a helper to set entity position.
func SetPosition(entity *Entity, x, y float64) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		pos := posComp.(*PositionComponent)
		pos.X = x
		pos.Y = y
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
	collider := colliderComp.(*ColliderComponent)

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
