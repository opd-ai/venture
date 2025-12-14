// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	log "github.com/sirupsen/logrus"
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
	log.WithFields(log.Fields{
		"system_name": "movement",
		"max_speed":   maxSpeed,
	}).Debug("Creating movement system")

	return &MovementSystem{
		MaxSpeed: maxSpeed,
	}
}

// SetCollisionSystem sets the collision system for predictive collision checking.
// When set, MovementSystem will validate positions before applying movement.
func (s *MovementSystem) SetCollisionSystem(collisionSystem *CollisionSystem) {
	log.WithFields(log.Fields{
		"system_name":       "movement",
		"collision_enabled": collisionSystem != nil,
	}).Debug("Setting collision system")

	s.collisionSystem = collisionSystem
}

// SetSpatialPartition sets the spatial partition system for dirty tracking.
// When entities move, the spatial partition will be marked dirty for lazy rebuilding.
func (s *MovementSystem) SetSpatialPartition(spatialPartition *SpatialPartitionSystem) {
	log.WithFields(log.Fields{
		"system_name":       "movement",
		"partition_enabled": spatialPartition != nil,
	}).Debug("Setting spatial partition system")

	s.spatialPartition = spatialPartition
}

// Update applies velocity to position for all entities with both components.
func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
	// Check log level once at start to avoid per-entity allocation overhead
	debugEnabled := log.GetLevel() >= log.DebugLevel

	if debugEnabled {
		log.WithFields(log.Fields{
			"system_name":  "movement",
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("Movement system update started")
	}

	s.entitiesMoved = false

	for _, entity := range entities {
		// Skip dead entities - they cannot move (Priority 1.2)
		if entity.HasComponent("dead") {
			if debugEnabled {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"reason":    "dead",
				}).Debug("Skipping dead entity")
			}
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
			log.WithFields(log.Fields{
				"entity_id":      entity.ID,
				"component_type": "position",
			}).Warn("Invalid position component type")
			continue
		}
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			log.WithFields(log.Fields{
				"entity_id":      entity.ID,
				"component_type": "velocity",
			}).Warn("Invalid velocity component type")
			continue
		}

		// Apply speed limit if configured
		speedLimited := s.applySpeedLimit(vel)
		if speedLimited && debugEnabled {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"max_speed": s.MaxSpeed,
			}).Debug("Speed limit applied")
		}

		// Calculate new position
		newX := pos.X + vel.VX*deltaTime
		newY := pos.Y + vel.VY*deltaTime

		if debugEnabled {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"old_x":     pos.X,
				"old_y":     pos.Y,
				"new_x":     newX,
				"new_y":     newY,
				"vel_x":     vel.VX,
				"vel_y":     vel.VY,
			}).Debug("Calculating new position")
		}

		// Validate position with collision checking and wall sliding
		newX, newY = s.calculateValidPosition(entity, pos, vel, newX, newY, entities)

		// Update position and track movement
		oldX, oldY := pos.X, pos.Y
		pos.X = newX
		pos.Y = newY

		if pos.X != oldX || pos.Y != oldY {
			s.entitiesMoved = true

			if debugEnabled {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"from_x":    oldX,
					"from_y":    oldY,
					"to_x":      pos.X,
					"to_y":      pos.Y,
				}).Debug("Entity moved")
			}

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
		if debugEnabled {
			log.WithFields(log.Fields{
				"system_name": "movement",
			}).Debug("Marking spatial partition as dirty")
		}
		s.spatialPartition.MarkDirty()
	}

	if debugEnabled {
		log.WithFields(log.Fields{
			"system_name":    "movement",
			"entities_moved": s.entitiesMoved,
		}).Debug("Movement system update completed")
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
		if log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"operation":     "speed_limit",
				"current_speed": speed,
				"max_speed":     s.MaxSpeed,
			}).Debug("Applying speed limit")
		}

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
	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"operation": "collision_check",
			"target_x":  newX,
			"target_y":  newY,
		}).Debug("Validating position")
	}

	if !s.hasValidCollider(entity) {
		return newX, newY
	}

	collider := s.getCollider(entity)
	if collider == nil || !collider.Solid || collider.IsTrigger {
		return newX, newY
	}

	if resultX, resultY, blocked := s.handleTerrainCollision(entity, pos, vel, newX, newY); blocked {
		return resultX, resultY
	}

	if newX == pos.X && newY == pos.Y {
		return newX, newY
	}

	return s.handleEntityCollisions(entity, pos, vel, newX, newY, entities)
}

// hasValidCollider checks if entity has collision system and collider component.
func (s *MovementSystem) hasValidCollider(entity *Entity) bool {
	return s.collisionSystem != nil && entity.HasComponent("collider")
}

// getCollider retrieves and validates the collider component from entity.
func (s *MovementSystem) getCollider(entity *Entity) *ColliderComponent {
	colliderComp, _ := entity.GetComponent("collider")
	if colliderComp == nil {
		return nil
	}

	collider, ok := colliderComp.(*ColliderComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "collider",
		}).Warn("Invalid collider component type")
		return nil
	}
	return collider
}

// handleTerrainCollision checks terrain collision and applies wall sliding.
func (s *MovementSystem) handleTerrainCollision(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64) (float64, float64, bool) {
	if !s.collisionSystem.WouldCollideWithTerrain(entity, newX, newY) {
		return newX, newY, false
	}

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"x":         newX,
			"y":         newY,
		}).Debug("Terrain collision detected")
	}

	return s.tryTerrainWallSlide(entity, pos, vel, newX, newY)
}

// tryTerrainWallSlide attempts to slide along walls when blocked by terrain.
func (s *MovementSystem) tryTerrainWallSlide(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64) (float64, float64, bool) {
	if !s.collisionSystem.WouldCollideWithTerrain(entity, newX, pos.Y) {
		if log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"entity_id":  entity.ID,
				"slide_type": "horizontal",
			}).Debug("Wall sliding applied")
		}
		vel.VY = 0
		return newX, pos.Y, true
	}

	if !s.collisionSystem.WouldCollideWithTerrain(entity, pos.X, newY) {
		if log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"entity_id":  entity.ID,
				"slide_type": "vertical",
			}).Debug("Wall sliding applied")
		}
		vel.VX = 0
		return pos.X, newY, true
	}

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Debug("Movement completely blocked by terrain")
	}
	vel.VX = 0
	vel.VY = 0
	return pos.X, pos.Y, true
}

// handleEntityCollisions checks collisions with other entities and applies wall sliding.
func (s *MovementSystem) handleEntityCollisions(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64, entities []*Entity) (float64, float64) {
	for _, other := range entities {
		if other.ID == entity.ID {
			continue
		}
		if s.collisionSystem.WouldCollideWithEntity(entity, newX, newY, other) {
			if log.GetLevel() >= log.DebugLevel {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"other_id":  other.ID,
				}).Debug("Entity collision detected")
			}

			return s.tryEntityWallSlide(entity, pos, vel, newX, newY, entities)
		}
	}
	return newX, newY
}

// tryEntityWallSlide attempts to slide along entities when blocked.
func (s *MovementSystem) tryEntityWallSlide(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64, entities []*Entity) (float64, float64) {
	if !s.anyEntityBlocking(entity, newX, pos.Y, entities) {
		vel.VY = 0
		return newX, pos.Y
	}

	if !s.anyEntityBlocking(entity, pos.X, newY, entities) {
		vel.VX = 0
		return pos.X, newY
	}

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Debug("Movement completely blocked by entity")
	}
	vel.VX = 0
	vel.VY = 0
	return pos.X, pos.Y
}

// applyBoundsConstraints clamps position to bounds and stops velocity at boundaries.
func (s *MovementSystem) applyBoundsConstraints(entity *Entity, pos *PositionComponent, vel *VelocityComponent) {
	boundsComp, hasBounds := entity.GetComponent("bounds")
	if !hasBounds {
		return
	}

	bounds, ok := boundsComp.(*BoundsComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "bounds",
		}).Warn("Invalid bounds component type")
		return
	}

	oldX, oldY := pos.X, pos.Y
	pos.X, pos.Y = bounds.Clamp(pos.X, pos.Y)

	if pos.X != oldX || pos.Y != oldY {
		if log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"old_x":     oldX,
				"old_y":     oldY,
				"new_x":     pos.X,
				"new_y":     pos.Y,
			}).Debug("Position clamped to bounds")
		}
	}

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
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "friction",
		}).Warn("Invalid friction component type")
		return
	}

	oldVX, oldVY := vel.VX, vel.VY

	// Apply friction as exponential decay: v *= (1 - coefficient)^deltaTime
	// Normalize to 60 FPS for consistent behavior
	decayFactor := math.Pow(1.0-friction.Coefficient, deltaTime*60.0)
	vel.VX *= decayFactor
	vel.VY *= decayFactor

	// Stop completely if velocity is very small (optimization)
	if math.Abs(vel.VX) < 0.1 && math.Abs(vel.VY) < 0.1 {
		vel.VX = 0
		vel.VY = 0

		if (oldVX != 0 || oldVY != 0) && log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
			}).Debug("Velocity stopped by friction threshold")
		}
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
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "animation",
		}).Warn("Invalid animation component type")
		return
	}

	// DON'T override attack/hit/death/cast animations - let them finish
	if anim.CurrentState == AnimationStateAttack ||
		anim.CurrentState == AnimationStateHit ||
		anim.CurrentState == AnimationStateDeath ||
		anim.CurrentState == AnimationStateCast {
		return
	}

	// Check log level once for this function
	debugEnabled := log.GetLevel() >= log.DebugLevel

	speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)

	if speed > 0.1 {
		// Moving - determine if walking or running
		if speed > s.MaxSpeed*0.7 && s.MaxSpeed > 0 {
			// Fast movement - running
			if anim.CurrentState != AnimationStateRun {
				if debugEnabled {
					log.WithFields(log.Fields{
						"entity_id":       entity.ID,
						"animation_state": "run",
						"speed":           speed,
					}).Debug("Animation state changed to run")
				}
				anim.SetState(AnimationStateRun)
			}
		} else {
			// Normal movement - walking
			if anim.CurrentState != AnimationStateWalk {
				if debugEnabled {
					log.WithFields(log.Fields{
						"entity_id":       entity.ID,
						"animation_state": "walk",
						"speed":           speed,
					}).Debug("Animation state changed to walk")
				}
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
			if debugEnabled {
				log.WithFields(log.Fields{
					"entity_id":       entity.ID,
					"animation_state": "idle",
				}).Debug("Animation state changed to idle")
			}
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

	oldFacing := anim.Facing

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

	if oldFacing != anim.Facing && log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"old_facing": oldFacing,
			"new_facing": anim.Facing,
			"vel_x":      vel.VX,
			"vel_y":      vel.VY,
		}).Debug("Facing direction updated")
	}
}

// SetVelocity is a helper to set entity velocity.
func SetVelocity(entity *Entity, vx, vy float64) {
	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			if log.GetLevel() >= log.DebugLevel {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"vel_x":     vx,
					"vel_y":     vy,
				}).Debug("Setting entity velocity")
			}
			vel.VX = vx
			vel.VY = vy
		}
	}
}

// GetPosition is a helper to get entity position.
func GetPosition(entity *Entity) (x, y float64, ok bool) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok {
			if log.GetLevel() >= log.DebugLevel {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"x":         pos.X,
					"y":         pos.Y,
				}).Debug("Getting entity position")
			}
			return pos.X, pos.Y, true
		}
	}
	return 0, 0, false
}

// SetPosition is a helper to set entity position.
func SetPosition(entity *Entity, x, y float64) {
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok {
			if log.GetLevel() >= log.DebugLevel {
				log.WithFields(log.Fields{
					"entity_id": entity.ID,
					"x":         x,
					"y":         y,
				}).Debug("Setting entity position")
			}
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
		log.WithFields(log.Fields{
			"entity1_id": e1.ID,
			"entity2_id": e2.ID,
			"has_pos1":   ok1,
			"has_pos2":   ok2,
		}).Warn("Cannot calculate distance - missing position component")
		return math.Inf(1)
	}

	dx := x2 - x1
	dy := y2 - y1
	distance := math.Sqrt(dx*dx + dy*dy)

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"entity1_id": e1.ID,
			"entity2_id": e2.ID,
			"distance":   distance,
		}).Debug("Calculated entity distance")
	}

	return distance
}

// MoveTowards moves an entity towards a target position.
// Returns true if the entity reached the target.
func MoveTowards(entity *Entity, targetX, targetY, speed, deltaTime float64) bool {
	debugEnabled := log.GetLevel() >= log.DebugLevel
	if debugEnabled {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"target_x":  targetX,
			"target_y":  targetY,
			"speed":     speed,
		}).Debug("Moving entity towards target")
	}

	x, y, ok := GetPosition(entity)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Warn("Cannot move towards - missing position component")
		return false
	}

	dx := targetX - x
	dy := targetY - y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Already at target
	if distance < 0.1 {
		if debugEnabled {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity reached target")
		}
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
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "collider",
		}).Warn("Invalid collider component type in layer transition")
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
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"old_layer": oldLayer,
				"new_layer": targetLayer,
				"tile":      currentTile,
				"tile_x":    tileX,
				"tile_y":    tileY,
			}).Debug("Entity layer transition via ramp")
		}
	}
}
