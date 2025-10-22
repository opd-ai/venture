// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import "math"

// MovementSystem handles entity movement based on velocity.
type MovementSystem struct {
	// MaxSpeed limits entity velocity (0 = no limit)
	MaxSpeed float64
}

// NewMovementSystem creates a new movement system.
func NewMovementSystem(maxSpeed float64) *MovementSystem {
	return &MovementSystem{
		MaxSpeed: maxSpeed,
	}
}

// Update applies velocity to position for all entities with both components.
func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
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

		// Update position based on velocity
		pos.X += vel.VX * deltaTime
		pos.Y += vel.VY * deltaTime

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
	}
}

// SetVelocity is a helper to set entity velocity.
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
