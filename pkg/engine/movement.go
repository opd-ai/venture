// Package engine provides movement mechanics for entities.
// This file implements movement logic with velocity, friction, and boundary
// checking for entity position updates.
package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	log "github.com/sirupsen/logrus"
)

// movementDebugEnabled caches whether debug-level movement logging is enabled.
// This variable is set at system creation to avoid per-frame GetLevel() checks.
// Eliminates ~1µs/frame overhead (mutex acquisition in GetLevel()).
var movementDebugEnabled bool

// SetMovementDebugEnabled updates the cached movement debug flag based on logger level.
// Call this whenever changing the log level to refresh the cached flag.
func SetMovementDebugEnabled(enabled bool) {
	movementDebugEnabled = enabled
}

// RefreshMovementDebugFlag updates the cached debug flag from the current log level.
func RefreshMovementDebugFlag() {
	movementDebugEnabled = log.GetLevel() >= log.DebugLevel
}

// MovementSystem handles entity movement based on velocity.
type MovementSystem struct {
	// MaxSpeed limits entity velocity (0 = no limit)
	MaxSpeed float64

	// CollisionSystem for predictive collision checking (optional)
	collisionSystem *CollisionSystem

	// SpatialPartitionSystem for dirty tracking (optional)
	spatialPartition *SpatialPartitionSystem

	// StatisticsSystem for tracking player movement stats (optional)
	statisticsSystem *StatisticsSystem

	// Track if any entity moved this frame
	entitiesMoved bool

	// Reusable buffer for nearby entity queries (reduces allocations)
	nearbyBuffer []*Entity

	// Track visited grid cells per entity for area exploration (grid cell size: 200x200).
	// Uses struct{} value to avoid storing a bool for set membership.
	visitedCells map[uint64]map[int64]struct{}
}

// NewMovementSystem creates a new movement system.
func NewMovementSystem(maxSpeed float64) *MovementSystem {
	// Cache debug flag once to avoid per-frame GetLevel() checks
	movementDebugEnabled = log.GetLevel() >= log.DebugLevel

	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"system_name": "movement",
			"max_speed":   maxSpeed,
		}).Debug("Creating movement system")
	}

	return &MovementSystem{
		MaxSpeed:     maxSpeed,
		nearbyBuffer: make([]*Entity, 0, 64), // Pre-allocate for typical nearby count
		visitedCells: make(map[uint64]map[int64]struct{}),
	}
}

// SetCollisionSystem sets the collision system for predictive collision checking.
// When set, MovementSystem will validate positions before applying movement.
func (s *MovementSystem) SetCollisionSystem(collisionSystem *CollisionSystem) {
	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"system_name":       "movement",
			"collision_enabled": collisionSystem != nil,
		}).Debug("Setting collision system")
	}

	s.collisionSystem = collisionSystem
}

// SetSpatialPartition sets the spatial partition system for dirty tracking.
// When entities move, the spatial partition will be marked dirty for lazy rebuilding.
func (s *MovementSystem) SetSpatialPartition(spatialPartition *SpatialPartitionSystem) {
	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"system_name":       "movement",
			"partition_enabled": spatialPartition != nil,
		}).Debug("Setting spatial partition system")
	}

	s.spatialPartition = spatialPartition
}

// SetStatisticsSystem sets the statistics system for tracking player movement stats.
// When set, player movement distance and area exploration will be tracked.
func (s *MovementSystem) SetStatisticsSystem(statisticsSystem *StatisticsSystem) {
	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"system_name":        "movement",
			"statistics_enabled": statisticsSystem != nil,
		}).Debug("Setting statistics system")
	}

	s.statisticsSystem = statisticsSystem
}

// Update applies velocity to position for all entities with both components.
func (s *MovementSystem) Update(entities []*Entity, deltaTime float64) {
	debugEnabled := s.logUpdateStart(len(entities), deltaTime)
	s.entitiesMoved = false

	for _, entity := range entities {
		if s.shouldSkipEntity(entity, debugEnabled) {
			continue
		}

		pos, vel := s.getMovementComponents(entity)
		if pos == nil || vel == nil {
			continue
		}

		s.processEntityMovement(entity, pos, vel, deltaTime, debugEnabled, entities)
	}

	s.finalizeUpdate(debugEnabled)
}

// logUpdateStart logs the movement system start and returns debug state.
func (s *MovementSystem) logUpdateStart(entityCount int, deltaTime float64) bool {
	// Use cached debug flag to avoid per-frame GetLevel() call
	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"system_name":  "movement",
			"entity_count": entityCount,
			"delta_time":   deltaTime,
		}).Debug("Movement system update started")
	}
	return movementDebugEnabled
}

// shouldSkipEntity checks if an entity should be skipped during movement processing.
func (s *MovementSystem) shouldSkipEntity(entity *Entity, debugEnabled bool) bool {
	if !entity.HasComponent("dead") {
		return false
	}
	if debugEnabled {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"reason":    "dead",
		}).Debug("Skipping dead entity")
	}
	return true
}

// getMovementComponents retrieves position and velocity components from entity.
func (s *MovementSystem) getMovementComponents(entity *Entity) (*PositionComponent, *VelocityComponent) {
	return entity.GetPosition(), entity.GetVelocity()
}

// processEntityMovement handles the complete movement logic for a single entity.
func (s *MovementSystem) processEntityMovement(entity *Entity, pos *PositionComponent, vel *VelocityComponent, deltaTime float64, debugEnabled bool, entities []*Entity) {
	s.applySpeedLimitWithLogging(entity, vel, debugEnabled)
	newX, newY := s.calculateNewPosition(entity, pos, vel, deltaTime, debugEnabled)
	moved := s.validateAndUpdatePosition(entity, pos, vel, newX, newY, debugEnabled, entities)
	s.handlePostMovement(entity, pos, vel, deltaTime, moved)
}

// applySpeedLimitWithLogging applies speed limit and logs if needed.
func (s *MovementSystem) applySpeedLimitWithLogging(entity *Entity, vel *VelocityComponent, debugEnabled bool) {
	if s.applySpeedLimit(vel) && debugEnabled {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"max_speed": s.MaxSpeed,
		}).Debug("Speed limit applied")
	}
}

// calculateNewPosition computes the target position based on velocity.
func (s *MovementSystem) calculateNewPosition(entity *Entity, pos *PositionComponent, vel *VelocityComponent, deltaTime float64, debugEnabled bool) (float64, float64) {
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

	return newX, newY
}

// validateAndUpdatePosition validates and applies the new position, returning true if entity moved.
func (s *MovementSystem) validateAndUpdatePosition(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64, debugEnabled bool, entities []*Entity) bool {
	newX, newY = s.calculateValidPosition(entity, pos, vel, newX, newY, entities)
	oldX, oldY := pos.X, pos.Y

	// Save previous position for render interpolation (fixes visual jitter between Update/Draw)
	pos.PrevX = oldX
	pos.PrevY = oldY

	// Track entity movement before updating position (for incremental spatial updates)
	if s.spatialPartition != nil && (newX != oldX || newY != oldY) {
		s.spatialPartition.TrackEntityMovement(entity)
	}

	pos.X = newX
	pos.Y = newY

	if pos.X == oldX && pos.Y == oldY {
		return false
	}

	s.entitiesMoved = true
	s.logMovement(entity, oldX, oldY, pos.X, pos.Y, debugEnabled)
	s.checkLayerTransition(entity, pos)

	// Track statistics for player entities (entities with input component)
	s.trackPlayerMovementStats(entity, oldX, oldY, newX, newY)

	return true
}

// logMovement logs entity movement if debug is enabled.
func (s *MovementSystem) logMovement(entity *Entity, fromX, fromY, toX, toY float64, debugEnabled bool) {
	if !debugEnabled {
		return
	}
	log.WithFields(log.Fields{
		"entity_id": entity.ID,
		"from_x":    fromX,
		"from_y":    fromY,
		"to_x":      toX,
		"to_y":      toY,
	}).Debug("Entity moved")
}

// areaCellSize defines the size of grid cells used for area exploration tracking.
// Each 200x200 pixel region counts as a distinct "area" for tutorial purposes.
const areaCellSize = 200.0

// trackPlayerMovementStats updates statistics for player entities when they move.
// Tracks distance traveled and unique areas visited for tutorial conditions.
func (s *MovementSystem) trackPlayerMovementStats(entity *Entity, oldX, oldY, newX, newY float64) {
	// Only track stats if statistics system is set and entity is a player
	if s.statisticsSystem == nil || !entity.HasComponent("input") {
		return
	}

	// Calculate distance traveled (Euclidean distance)
	dx := newX - oldX
	dy := newY - oldY
	distance := math.Sqrt(dx*dx + dy*dy)

	// Track distance traveled (convert to int64, minimum 1 if moved)
	if distance > 0 {
		distInt := int64(distance)
		if distInt < 1 {
			distInt = 1
		}
		s.statisticsSystem.OnDistanceTraveled(entity.ID, distInt)
	}

	// Track unique area visits using grid cells
	s.trackAreaVisit(entity, newX, newY)
}

// trackAreaVisit tracks whether the entity has entered a new grid cell area.
// Uses a 200x200 pixel grid to define areas for exploration tracking.
func (s *MovementSystem) trackAreaVisit(entity *Entity, x, y float64) {
	// Calculate grid cell coordinates
	cellX := int64(x / areaCellSize)
	cellY := int64(y / areaCellSize)
	cellKey := cellX<<32 | (cellY & 0xFFFFFFFF)

	// Get or create visited cells map for this entity.
	// Pre-allocate with a reasonable capacity to amortize future insertions.
	cells, exists := s.visitedCells[entity.ID]
	if !exists {
		cells = make(map[int64]struct{}, 64)
		s.visitedCells[entity.ID] = cells
	}

	// Check if this is a new cell
	if _, visited := cells[cellKey]; !visited {
		cells[cellKey] = struct{}{}
		s.statisticsSystem.OnAreaVisited(entity.ID)

		if movementDebugEnabled {
			log.WithFields(log.Fields{
				"entity_id":   entity.ID,
				"cell_x":      cellX,
				"cell_y":      cellY,
				"total_areas": len(cells),
			}).Debug("New area visited")
		}
	}
}

// ClearVisitedCells clears the visited cells tracking for an entity.
// Called automatically via the World entity-removal hook registered by
// RegisterRemovalHook; can also be called directly to reset exploration
// state (e.g. when starting a new session or resetting the tutorial).
func (s *MovementSystem) ClearVisitedCells(entityID uint64) {
	delete(s.visitedCells, entityID)
}

// RegisterRemovalHook registers a World entity-removal hook so that visitedCells
// entries are reclaimed when an entity is removed. Call this once after
// AddSystem(movementSystem) so that the hook fires on the game-loop goroutine
// whenever an entity is drained from the removal queue.
func (s *MovementSystem) RegisterRemovalHook(world *World) {
	world.AddEntityRemovalHook(s.ClearVisitedCells)
}

// handlePostMovement applies bounds, friction, and animation updates after movement.
func (s *MovementSystem) handlePostMovement(entity *Entity, pos *PositionComponent, vel *VelocityComponent, deltaTime float64, moved bool) {
	s.applyBoundsConstraints(entity, pos, vel)
	s.applyFriction(entity, vel, deltaTime)
	s.updateAnimationState(entity, vel)
}

// finalizeUpdate marks spatial partition dirty and logs completion.
func (s *MovementSystem) finalizeUpdate(debugEnabled bool) {
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

// queryNearbyEntities returns entities near the given position using spatial partition if available.
// Falls back to full entity list if spatial partition is not set.
// Uses reusable buffer to minimize allocations.
func (s *MovementSystem) queryNearbyEntities(entity *Entity, x, y float64, entities []*Entity) []*Entity {
	if s.spatialPartition == nil {
		return entities
	}

	// Get entity collision size for query bounds (default 32x32 if no collider)
	queryRadius := 64.0 // Default search radius
	if collider := entity.GetCollider(); collider != nil {
		// Use max dimension + margin for collision detection
		maxDim := collider.Width
		if collider.Height > maxDim {
			maxDim = collider.Height
		}
		queryRadius = maxDim * 2 // 2x entity size for safe collision checking
	}

	// Create query bounds centered on target position
	queryBounds := Bounds{
		X:      x - queryRadius,
		Y:      y - queryRadius,
		Width:  queryRadius * 2,
		Height: queryRadius * 2,
	}

	// Reset buffer and query nearby entities
	s.nearbyBuffer = s.nearbyBuffer[:0]
	s.nearbyBuffer = s.spatialPartition.QueryBoundsInto(queryBounds, s.nearbyBuffer)

	return s.nearbyBuffer
}

// anyEntityBlocking checks if any entity would block movement to the given position.
// Helper method for collision sliding logic.
// Uses spatial partition for O(log n) queries when available.
func (s *MovementSystem) anyEntityBlocking(entity *Entity, x, y float64, entities []*Entity) bool {
	if s.collisionSystem == nil {
		return false
	}

	// Query nearby entities using spatial partition if available
	nearbyEntities := s.queryNearbyEntities(entity, x, y, entities)

	for _, other := range nearbyEntities {
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
		if movementDebugEnabled {
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
	if movementDebugEnabled {
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
// Uses typed getter for zero-overhead access.
func (s *MovementSystem) getCollider(entity *Entity) *ColliderComponent {
	return entity.GetCollider()
}

// handleTerrainCollision checks terrain collision and applies wall sliding.
func (s *MovementSystem) handleTerrainCollision(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64) (float64, float64, bool) {
	if !s.collisionSystem.WouldCollideWithTerrain(entity, newX, newY) {
		return newX, newY, false
	}

	if movementDebugEnabled {
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
		if movementDebugEnabled {
			log.WithFields(log.Fields{
				"entity_id":  entity.ID,
				"slide_type": "horizontal",
			}).Debug("Wall sliding applied")
		}
		vel.VY = 0
		return newX, pos.Y, true
	}

	if !s.collisionSystem.WouldCollideWithTerrain(entity, pos.X, newY) {
		if movementDebugEnabled {
			log.WithFields(log.Fields{
				"entity_id":  entity.ID,
				"slide_type": "vertical",
			}).Debug("Wall sliding applied")
		}
		vel.VX = 0
		return pos.X, newY, true
	}

	if movementDebugEnabled {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Debug("Movement completely blocked by terrain")
	}
	vel.VX = 0
	vel.VY = 0
	return pos.X, pos.Y, true
}

// handleEntityCollisions checks collisions with other entities and applies wall sliding.
// Uses spatial partition for O(log n) queries when available.
func (s *MovementSystem) handleEntityCollisions(entity *Entity, pos *PositionComponent, vel *VelocityComponent, newX, newY float64, entities []*Entity) (float64, float64) {
	// Query nearby entities using spatial partition if available
	nearbyEntities := s.queryNearbyEntities(entity, newX, newY, entities)

	for _, other := range nearbyEntities {
		if other.ID == entity.ID {
			continue
		}
		if s.collisionSystem.WouldCollideWithEntity(entity, newX, newY, other) {
			if movementDebugEnabled {
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

	if movementDebugEnabled {
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
		if movementDebugEnabled {
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

		if (oldVX != 0 || oldVY != 0) && movementDebugEnabled {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
			}).Debug("Velocity stopped by friction threshold")
		}
	}
}

// updateAnimationState updates entity animation based on velocity and movement state.
// Handles idle/walk/run state transitions and facing direction updates.
func (s *MovementSystem) updateAnimationState(entity *Entity, vel *VelocityComponent) {
	anim := s.getAnimationComponent(entity)
	if anim == nil || s.shouldPreserveAnimationState(anim) {
		return
	}

	debugEnabled := movementDebugEnabled
	speed := s.calculateSpeed(vel)

	if speed > 0.1 {
		s.handleMovingState(entity, anim, vel, speed, debugEnabled)
	} else {
		s.handleIdleState(entity, anim, debugEnabled)
	}
}

// getAnimationComponent retrieves and validates the animation component.
func (s *MovementSystem) getAnimationComponent(entity *Entity) *AnimationComponent {
	animComp, hasAnim := entity.GetComponent("animation")
	if !hasAnim {
		return nil
	}

	anim, ok := animComp.(*AnimationComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "animation",
		}).Warn("Invalid animation component type")
		return nil
	}
	return anim
}

// shouldPreserveAnimationState checks if current animation should not be interrupted.
func (s *MovementSystem) shouldPreserveAnimationState(anim *AnimationComponent) bool {
	return anim.CurrentState == AnimationStateAttack ||
		anim.CurrentState == AnimationStateHit ||
		anim.CurrentState == AnimationStateDeath ||
		anim.CurrentState == AnimationStateCast
}

// calculateSpeed computes the magnitude of velocity vector.
func (s *MovementSystem) calculateSpeed(vel *VelocityComponent) float64 {
	return math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
}

// handleMovingState updates animation for moving entities (walk/run).
func (s *MovementSystem) handleMovingState(entity *Entity, anim *AnimationComponent, vel *VelocityComponent, speed float64, debugEnabled bool) {
	if s.shouldRun(speed) {
		s.transitionToRun(entity, anim, speed, debugEnabled)
	} else {
		s.transitionToWalk(entity, anim, speed, debugEnabled)
	}

	if !entity.HasComponent("rotation") {
		s.updateFacingDirection(anim, vel)
	}
}

// shouldRun determines if entity should be running based on speed.
func (s *MovementSystem) shouldRun(speed float64) bool {
	return speed > s.MaxSpeed*0.7 && s.MaxSpeed > 0
}

// transitionToRun changes animation to run state if needed.
func (s *MovementSystem) transitionToRun(entity *Entity, anim *AnimationComponent, speed float64, debugEnabled bool) {
	if anim.CurrentState == AnimationStateRun {
		return
	}
	if debugEnabled {
		log.WithFields(log.Fields{
			"entity_id":       entity.ID,
			"animation_state": "run",
			"speed":           speed,
		}).Debug("Animation state changed to run")
	}
	anim.SetState(AnimationStateRun)
}

// transitionToWalk changes animation to walk state if needed.
func (s *MovementSystem) transitionToWalk(entity *Entity, anim *AnimationComponent, speed float64, debugEnabled bool) {
	if anim.CurrentState == AnimationStateWalk {
		return
	}
	if debugEnabled {
		log.WithFields(log.Fields{
			"entity_id":       entity.ID,
			"animation_state": "walk",
			"speed":           speed,
		}).Debug("Animation state changed to walk")
	}
	anim.SetState(AnimationStateWalk)
}

// handleIdleState updates animation for idle entities.
func (s *MovementSystem) handleIdleState(entity *Entity, anim *AnimationComponent, debugEnabled bool) {
	if anim.CurrentState != AnimationStateWalk && anim.CurrentState != AnimationStateRun {
		return
	}
	if debugEnabled {
		log.WithFields(log.Fields{
			"entity_id":       entity.ID,
			"animation_state": "idle",
		}).Debug("Animation state changed to idle")
	}
	anim.SetState(AnimationStateIdle)
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

	if oldFacing != anim.Facing && movementDebugEnabled {
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
			if movementDebugEnabled {
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
			if movementDebugEnabled {
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
			if movementDebugEnabled {
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

	if movementDebugEnabled {
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
	debugEnabled := movementDebugEnabled
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

	// Get entity's collider using typed getter for zero-overhead access
	collider := entity.GetCollider()
	if collider == nil {
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
