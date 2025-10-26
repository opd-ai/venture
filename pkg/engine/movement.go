// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import (
	"fmt"
	"math"
)

// MovementSystem handles entity movement based on velocity.
type MovementSystem struct {
	// MaxSpeed limits entity velocity (0 = no limit)
	MaxSpeed float64

	// CollisionSystem for predictive collision checking (optional)
	collisionSystem *CollisionSystem
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

// Update applies velocity to position for all entities with both components.
func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
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
		pos.X = newX
		pos.Y = newY

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
				if entity.HasComponent("input") {
					// DEBUG: Log when we skip overriding attack animations
					fmt.Printf("[MOVEMENT] Skipping animation update - entity in %s state\n", anim.CurrentState)
				}
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
			} else {
				// Not moving - idle (only if currently in a movement state)
				if anim.CurrentState == AnimationStateWalk || anim.CurrentState == AnimationStateRun {
					anim.SetState(AnimationStateIdle)
				}
			}
		}
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
