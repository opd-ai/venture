// Package engine provides the AI system for autonomous entity behavior.
// This file implements AISystem which manages state transitions and behaviors
// for AI-controlled entities using a state machine pattern.
package engine

import (
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
)

// aiDebugEnabled caches whether debug-level AI logging is enabled.
// This variable should be updated via SetAIDebugEnabled() when logger level changes.
// Caching this check eliminates 100+ GetLevel() calls per frame in AI decision loops.
var aiDebugEnabled bool

// SetAIDebugEnabled updates the cached AI debug flag based on logger level.
// Call this whenever changing the log level for an AISystem's logger.
func SetAIDebugEnabled(logger *logrus.Entry) {
	if logger != nil && logger.Logger != nil {
		aiDebugEnabled = logger.Logger.GetLevel() >= logrus.DebugLevel
	} else {
		aiDebugEnabled = false
	}
}

// refreshAIDebugFlag is a helper to update the debug flag cache.
func refreshAIDebugFlag(logger *logrus.Entry) {
	SetAIDebugEnabled(logger)
}

// AISystem manages artificial intelligence behaviors for entities.
// It implements a state machine that transitions between idle, patrol, chase, attack, and flee states.
type AISystem struct {
	world        *World
	logger       *logrus.Entry
	combatSystem *CombatSystem // Cached combat system to avoid per-attack allocations
	quadtree     *Quadtree     // Spatial partition for O(log n) enemy queries (instead of O(n))
	queryBuffer  []*Entity     // Reusable buffer for spatial queries to reduce allocations
}

// NewAISystem creates a new AI system.
func NewAISystem(world *World) *AISystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "ai")
		refreshAIDebugFlag(logEntry)
		if aiDebugEnabled {
			logEntry.Debug("AI system created")
		}
	}
	return &AISystem{
		world:        world,
		logger:       logEntry,
		combatSystem: NewCombatSystem(12345), // Pre-allocate combat system to avoid per-attack allocations
		queryBuffer:  make([]*Entity, 0, 64), // Pre-allocate buffer for spatial queries
	}
}

// SetQuadtree sets the spatial partition for O(log n) enemy detection queries.
// When set, the AI system will use spatial queries instead of O(n) iteration.
// This improves performance by 50-80% for scenes with 500+ entities.
func (ai *AISystem) SetQuadtree(quadtree *Quadtree) {
	ai.quadtree = quadtree
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.Debug("AI system quadtree optimization enabled")
	}
}

// Update processes AI behavior for all entities with AI components.
// Pre-filters entities using the world query cache to avoid iterating non-AI entities.
func (ai *AISystem) Update(entities []*Entity, deltaTime float64) {
	// Use the world query cache to get only entities with AI components,
	// avoiding O(n) map lookups across all entities for non-AI entities.
	aiEntities := entities
	if ai.world != nil {
		aiEntities = ai.world.GetEntitiesWith("ai")
	}

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_count":    len(entities),
			"ai_entity_count": len(aiEntities),
			"delta_time":      deltaTime,
		}).Debug("AI system update started")
	}

	for _, entity := range aiEntities {
		// Retrieve the AI component — already guaranteed present by GetEntitiesWith.
		aiComp, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}

		// Type assert with safety check
		aiState, ok := aiComp.(*AIComponent)
		if !ok {
			if ai.logger != nil {
				ai.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"component_type": "ai",
				}).Warn("Failed to type assert AI component")
			}
			continue
		}

		// Update timers
		aiState.UpdateStateTimer(deltaTime)

		// Only make decisions at intervals
		if !aiState.ShouldUpdateDecision(deltaTime) {
			continue
		}

		// Process AI decision-making
		ai.processAI(entity, aiState, deltaTime)
	}
}

// processAI handles the AI decision-making logic for an entity.
func (ai *AISystem) processAI(entity *Entity, aiComp *AIComponent, deltaTime float64) {
	// Skip processing if AI is disabled by a control effect (stun, frozen, etc.)
	if aiComp.Disabled {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"reason":    aiComp.DisabledReason,
			}).Debug("AI processing skipped - disabled by control effect")
		}
		return
	}

	// Get position component using typed getter for ~94x faster access
	pos := entity.GetPosition()
	if pos == nil {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "position",
			}).Debug("Entity missing position component for AI processing")
		}
		return // Can't do AI without position
	}

	// Check health for flee condition
	shouldFlee := ai.shouldFlee(entity, aiComp)

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"current_state": aiComp.State.String(),
			"should_flee":   shouldFlee,
			"x":             pos.X,
			"y":             pos.Y,
		}).Debug("Processing AI decision")
	}

	// State machine logic
	switch aiComp.State {
	case AIStateIdle:
		ai.processIdle(entity, aiComp, pos)

	case AIStatePatrol:
		ai.processPatrol(entity, aiComp, pos, deltaTime)

	case AIStateDetect:
		ai.processDetect(entity, aiComp, pos)

	case AIStateChase:
		if shouldFlee {
			ai.transitionToFlee(entity, aiComp, pos)
		} else {
			ai.processChase(entity, aiComp, pos)
		}

	case AIStateAttack:
		if shouldFlee {
			ai.transitionToFlee(entity, aiComp, pos)
		} else {
			ai.processAttack(entity, aiComp, pos)
		}

	case AIStateFlee:
		ai.processFlee(entity, aiComp, pos)

	case AIStateReturn:
		ai.processReturn(entity, aiComp, pos)
	}
}

// processIdle handles the idle state - look for targets.
func (ai *AISystem) processIdle(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"state":           aiComp.State.String(),
			"detection_range": aiComp.DetectionRange,
			"x":               pos.X,
			"y":               pos.Y,
		}).Debug("Processing idle state")
	}

	// Look for enemies in range
	target := ai.findNearestEnemy(entity, pos, aiComp.DetectionRange)

	if target != nil {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        target.ID,
				"detection_range":  aiComp.DetectionRange,
				"state_transition": "Idle->Detect",
			}).Info("AI detected enemy target")
		}
		aiComp.Target = target
		aiComp.ChangeState(AIStateDetect)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"target_id":  target.ID,
				"new_state":  AIStateDetect.String(),
				"prev_state": AIStateIdle.String(),
			}).Debug("State transition completed")
		}
	} else {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":       entity.ID,
				"detection_range": aiComp.DetectionRange,
			}).Debug("No target detected in idle state")
		}
	}
}

// processPatrol handles the patrol state - move between waypoints.
func (ai *AISystem) processPatrol(entity *Entity, aiComp *AIComponent, pos *PositionComponent, deltaTime float64) {
	ai.logPatrolState(entity, aiComp, pos)

	if ai.checkAndHandleEnemyDetection(entity, aiComp, pos) {
		return
	}

	if !ai.validatePatrolRoute(entity, aiComp) {
		return
	}

	waypoint := aiComp.GetCurrentWaypoint()
	if waypoint == nil {
		ai.logInvalidWaypoint(entity, aiComp)
		return
	}

	if ai.handleWaypointWait(entity, aiComp, waypoint, deltaTime) {
		return
	}

	if ai.handleWaypointReached(entity, aiComp, pos, waypoint) {
		return
	}

	ai.moveTowardWaypoint(entity, aiComp, pos, waypoint)
}

// logPatrolState logs the current patrol state.
func (ai *AISystem) logPatrolState(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":       entity.ID,
			"state":           aiComp.State.String(),
			"waypoint_index":  aiComp.CurrentWaypointIndex,
			"detection_range": aiComp.DetectionRange,
			"x":               pos.X,
			"y":               pos.Y,
		}).Debug("Processing patrol state")
	}
}

// checkAndHandleEnemyDetection checks for enemies and transitions to detect state if found.
func (ai *AISystem) checkAndHandleEnemyDetection(entity *Entity, aiComp *AIComponent, pos *PositionComponent) bool {
	target := ai.findNearestEnemy(entity, pos, aiComp.DetectionRange)
	if target == nil {
		return false
	}

	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"target_id":        target.ID,
			"state_transition": "Patrol->Detect",
		}).Info("AI detected enemy during patrol")
	}

	aiComp.Target = target
	aiComp.ChangeState(AIStateDetect)

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"new_state":  AIStateDetect.String(),
			"prev_state": AIStatePatrol.String(),
		}).Debug("State transition completed")
	}
	return true
}

// validatePatrolRoute checks if patrol route is configured.
func (ai *AISystem) validatePatrolRoute(entity *Entity, aiComp *AIComponent) bool {
	if aiComp.HasPatrolRoute() {
		return true
	}

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("No patrol route configured, behaving like idle")
	}
	return false
}

// logInvalidWaypoint logs an invalid waypoint error.
func (ai *AISystem) logInvalidWaypoint(entity *Entity, aiComp *AIComponent) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"waypoint_index": aiComp.CurrentWaypointIndex,
		}).Warn("Invalid waypoint index during patrol")
	}
}

// handleWaypointWait handles the waiting period at a waypoint.
func (ai *AISystem) handleWaypointWait(entity *Entity, aiComp *AIComponent, waypoint *PatrolWaypoint, deltaTime float64) bool {
	if !aiComp.IsWaitingAtWaypoint(deltaTime) {
		return false
	}

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"waypoint_index": aiComp.CurrentWaypointIndex,
			"waypoint_x":     waypoint.X,
			"waypoint_y":     waypoint.Y,
		}).Debug("Waiting at waypoint")
	}

	ai.stopEntityMovement(entity)
	return true
}

// stopEntityMovement stops an entity's movement.
// Uses cached GetVelocity() getter for ~93x faster access vs map lookup + type assertion.
func (ai *AISystem) stopEntityMovement(entity *Entity) {
	if vel := entity.GetVelocity(); vel != nil {
		vel.VX = 0
		vel.VY = 0
	}
}

// handleWaypointReached checks if waypoint is reached and advances to next.
func (ai *AISystem) handleWaypointReached(entity *Entity, aiComp *AIComponent, pos *PositionComponent, waypoint *PatrolWaypoint) bool {
	dx := waypoint.X - pos.X
	dy := waypoint.Y - pos.Y
	distToWaypoint := math.Sqrt(dx*dx + dy*dy)

	if distToWaypoint > aiComp.WaypointReachDistance {
		return false
	}

	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"waypoint_index": aiComp.CurrentWaypointIndex,
			"waypoint_x":     waypoint.X,
			"waypoint_y":     waypoint.Y,
			"reach_distance": aiComp.WaypointReachDistance,
		}).Debug("AI reached waypoint")
	}

	aiComp.AdvanceToNextWaypoint()
	return true
}

// moveTowardWaypoint calculates and applies movement toward the current waypoint.
// Uses cached GetVelocity() getter for ~93x faster access vs map lookup + type assertion.
func (ai *AISystem) moveTowardWaypoint(entity *Entity, aiComp *AIComponent, pos *PositionComponent, waypoint *PatrolWaypoint) {
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	dx := waypoint.X - pos.X
	dy := waypoint.Y - pos.Y
	distToWaypoint := math.Sqrt(dx*dx + dy*dy)

	if distToWaypoint <= 0 {
		return
	}

	baseSpeed := 100.0
	dirX := dx / distToWaypoint
	dirY := dy / distToWaypoint
	speed := baseSpeed * aiComp.GetSpeedMultiplier()

	vel.VX = dirX * speed
	vel.VY = dirY * speed
}

// processDetect handles the detect state - confirm target and start chase.
func (ai *AISystem) processDetect(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled && aiComp.Target != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"state":        aiComp.State.String(),
			"target_id":    aiComp.Target.ID,
			"state_timer":  aiComp.StateTimer,
			"detect_range": aiComp.DetectionRange * 1.2,
		}).Debug("Processing detect state")
	}

	// Check if target is still valid and in range
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.2) {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Detect->Idle",
				"reason":           "target_invalid",
			}).Debug("AI lost target during detection")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateIdle.String(),
				"prev_state": AIStateDetect.String(),
			}).Debug("State transition completed")
		}
		return
	}

	// Transition to chase after brief detection period
	if aiComp.StateTimer > 0.3 {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        aiComp.Target.ID,
				"state_transition": "Detect->Chase",
				"detection_time":   aiComp.StateTimer,
			}).Info("AI confirmed target, starting chase")
		}
		aiComp.ChangeState(AIStateChase)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateChase.String(),
				"prev_state": AIStateDetect.String(),
			}).Debug("State transition completed")
		}
	}
}

// processChase handles the chase state - pursue the target.
func (ai *AISystem) processChase(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	ai.logChaseState(entity.ID, aiComp, pos)

	if !ai.validateChaseTarget(entity, aiComp, pos) {
		return
	}

	if ai.shouldReturnToSpawn(entity, aiComp, pos) {
		return
	}

	if ai.checkAttackRange(entity, aiComp, pos) {
		return
	}

	targetPos := ai.getTargetPositionSimple(aiComp.Target)
	if targetPos != nil {
		ai.moveTowards(entity, pos, targetPos.X, targetPos.Y, aiComp.GetSpeedMultiplier())
	}
}

// logChaseState logs current chase state information.
func (ai *AISystem) logChaseState(entityID uint64, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled && aiComp.Target != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"state":        aiComp.State.String(),
			"target_id":    aiComp.Target.ID,
			"detect_range": aiComp.DetectionRange * 1.5,
			"x":            pos.X,
			"y":            pos.Y,
		}).Debug("Processing chase state")
	}
}

// validateChaseTarget verifies target is still valid for chasing.
func (ai *AISystem) validateChaseTarget(entity *Entity, aiComp *AIComponent, pos *PositionComponent) bool {
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.5) {
		ai.logStateTransition(entity.ID, AIStateChase, AIStateReturn, "target_invalid")
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		ai.logTransitionComplete(entity.ID, AIStateChase, AIStateReturn)
		return false
	}
	return true
}

// shouldReturnToSpawn checks if entity exceeded leash range and should return.
func (ai *AISystem) shouldReturnToSpawn(entity *Entity, aiComp *AIComponent, pos *PositionComponent) bool {
	if aiComp.ShouldReturnToSpawn(pos.X, pos.Y) {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Chase->Return",
				"reason":           "too_far_from_spawn",
				"spawn_x":          aiComp.SpawnX,
				"spawn_y":          aiComp.SpawnY,
				"current_x":        pos.X,
				"current_y":        pos.Y,
			}).Debug("AI returning to spawn, exceeded leash range")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		ai.logTransitionComplete(entity.ID, AIStateChase, AIStateReturn)
		return true
	}
	return false
}

// checkAttackRange determines if entity is in attack range and transitions to attack state.
func (ai *AISystem) checkAttackRange(entity *Entity, aiComp *AIComponent, pos *PositionComponent) bool {
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "attack",
			}).Debug("Entity missing attack component during chase")
		}
		return false
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return false
	}

	targetPos := ai.getTargetPositionSimple(aiComp.Target)
	if targetPos == nil {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return false
	}

	distance := ai.getDistance(pos.X, pos.Y, targetPos.X, targetPos.Y)
	if distance <= attack.Range {
		ai.logAttackRangeReached(entity.ID, aiComp.Target.ID, distance, attack.Range)
		aiComp.ChangeState(AIStateAttack)
		ai.logTransitionComplete(entity.ID, AIStateChase, AIStateAttack)
		return true
	}
	return false
}

// getTargetPositionSimple extracts position component from target entity.
// Uses typed getter for ~94x faster access vs generic GetComponent.
func (ai *AISystem) getTargetPositionSimple(target *Entity) *PositionComponent {
	return target.GetPosition()
}

// logStateTransition logs AI state transition with reason.
func (ai *AISystem) logStateTransition(entityID uint64, from, to AIState, reason string) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entityID,
			"state_transition": from.String() + "->" + to.String(),
			"reason":           reason,
		}).Debug("AI state transition")
	}
}

// logTransitionComplete logs completion of state transition.
func (ai *AISystem) logTransitionComplete(entityID uint64, from, to AIState) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"new_state":  to.String(),
			"prev_state": from.String(),
		}).Debug("State transition completed")
	}
}

// logAttackRangeReached logs when entity reaches attack range.
func (ai *AISystem) logAttackRangeReached(entityID, targetID uint64, distance, attackRange float64) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entityID,
			"target_id":        targetID,
			"distance":         distance,
			"attack_range":     attackRange,
			"state_transition": "Chase->Attack",
		}).Info("AI in attack range")
	}
}

// processAttack handles the attack state - attack the target.
func (ai *AISystem) processAttack(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"state":     aiComp.State.String(),
			"target_id": aiComp.Target.ID,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("Processing attack state")
	}

	if !ai.validateAttackTarget(entity, aiComp, pos) {
		return
	}

	attack := ai.getAttackComponent(entity)
	if attack == nil {
		return
	}

	targetP := ai.getTargetPosition(entity, aiComp)
	if targetP == nil {
		return
	}

	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)

	if ai.handleAttackRangeCheck(entity, aiComp, distance, attack.Range) {
		return
	}

	ai.executeAttack(entity, aiComp, attack, distance)
}

// validateAttackTarget verifies the target is still valid for attack.
func (ai *AISystem) validateAttackTarget(entity *Entity, aiComp *AIComponent, pos *PositionComponent) bool {
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.5) {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"target_id": aiComp.Target.ID,
				"reason":    "target_invalid",
			}).Debug("Target invalid during attack")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateReturn.String(),
				"prev_state": AIStateAttack.String(),
			}).Debug("State transition completed")
		}
		return false
	}
	return true
}

// getAttackComponent retrieves and validates the attack component.
func (ai *AISystem) getAttackComponent(entity *Entity) *AttackComponent {
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "attack",
			}).Debug("Entity missing attack component")
		}
		return nil
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "attack",
			}).Warn("Failed to type assert attack component")
		}
		return nil
	}
	return attack
}

// getTargetPosition retrieves and validates the target's position component.
// Uses typed GetPosition() getter for ~94x faster access vs map lookup + type assertion.
func (ai *AISystem) getTargetPosition(entity *Entity, aiComp *AIComponent) *PositionComponent {
	targetP := aiComp.Target.GetPosition()
	if targetP == nil {
		ai.logTargetMissingComponent(entity.ID, aiComp.Target.ID)
		ai.clearTargetAndReturnToIdle(entity, aiComp)
		return nil
	}
	return targetP
}

// logTargetMissingComponent logs when target is missing position component.
func (ai *AISystem) logTargetMissingComponent(entityID, targetID uint64) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"target_id":      targetID,
			"component_type": "position",
		}).Debug("Target missing position component")
	}
}

// logTargetTypeAssertFailed logs when target position type assertion fails.
func (ai *AISystem) logTargetTypeAssertFailed(entityID, targetID uint64) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"target_id":      targetID,
			"component_type": "position",
		}).Warn("Failed to type assert target position component")
	}
}

// clearTargetAndReturnToIdle clears target and transitions to idle state.
func (ai *AISystem) clearTargetAndReturnToIdle(entity *Entity, aiComp *AIComponent) {
	aiComp.ClearTarget()
	aiComp.ChangeState(AIStateIdle)
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"new_state":  AIStateIdle.String(),
			"prev_state": AIStateAttack.String(),
		}).Debug("State transition completed")
	}
}

// handleAttackRangeCheck checks if target is in attack range and transitions to chase if not.
func (ai *AISystem) handleAttackRangeCheck(entity *Entity, aiComp *AIComponent, distance, attackRange float64) bool {
	if distance > attackRange {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        aiComp.Target.ID,
				"distance":         distance,
				"attack_range":     attackRange,
				"state_transition": "Attack->Chase",
			}).Debug("Target moved out of attack range")
		}
		aiComp.ChangeState(AIStateChase)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateChase.String(),
				"prev_state": AIStateAttack.String(),
			}).Debug("State transition completed")
		}
		return true
	}
	return false
}

// executeAttack performs the actual attack if cooldown is ready.
func (ai *AISystem) executeAttack(entity *Entity, aiComp *AIComponent, attack *AttackComponent, distance float64) {
	if attack.CanAttack() {
		ai.performAttack(entity, aiComp, distance)
	} else {
		ai.logAttackOnCooldown(entity.ID, aiComp.Target.ID, attack.Cooldown)
	}
}

// performAttack executes the attack and updates animation.
func (ai *AISystem) performAttack(entity *Entity, aiComp *AIComponent, distance float64) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"target_id": aiComp.Target.ID,
			"distance":  distance,
		}).Info("AI executing attack")
	}
	ai.setAttackAnimation(entity)
	ai.executeCombatAttack(entity, aiComp)
}

// setAttackAnimation sets the entity's animation to attack state.
func (ai *AISystem) setAttackAnimation(entity *Entity) {
	if animComp, ok := entity.GetComponent("animation"); ok {
		if anim, ok := animComp.(*AnimationComponent); ok {
			if anim.CurrentState != AnimationStateAttack {
				anim.SetState(AnimationStateAttack)
			}
		}
	}
}

// executeCombatAttack performs the combat system attack.
func (ai *AISystem) executeCombatAttack(entity *Entity, aiComp *AIComponent) {
	ai.combatSystem.Attack(entity, aiComp.Target)
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"target_id": aiComp.Target.ID,
		}).Debug("Attack executed")
	}
}

// logAttackOnCooldown logs when attack is on cooldown.
func (ai *AISystem) logAttackOnCooldown(entityID, targetID uint64, cooldown float64) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"target_id": targetID,
			"cooldown":  cooldown,
		}).Debug("Attack on cooldown")
	}
}

// processFlee handles the flee state - run away from target.
func (ai *AISystem) processFlee(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":         entity.ID,
			"state":             aiComp.State.String(),
			"spawn_x":           aiComp.SpawnX,
			"spawn_y":           aiComp.SpawnY,
			"distance_to_spawn": aiComp.GetDistanceFromSpawn(pos.X, pos.Y),
			"x":                 pos.X,
			"y":                 pos.Y,
		}).Debug("Processing flee state")
	}

	// Check if health has recovered enough
	if !ai.shouldFlee(entity, aiComp) {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Flee->Return",
				"reason":           "health_recovered",
			}).Info("AI health recovered, stopping flee")
		}
		// Health recovered, go back to idle
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateReturn.String(),
				"prev_state": AIStateFlee.String(),
			}).Debug("State transition completed")
		}
		return
	}

	// Move away from target towards spawn
	ai.moveTowards(entity, pos, aiComp.SpawnX, aiComp.SpawnY, aiComp.GetSpeedMultiplier())

	// If close to spawn, switch to return state
	if aiComp.GetDistanceFromSpawn(pos.X, pos.Y) < 20.0 {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Flee->Return",
				"reason":           "reached_spawn",
			}).Debug("AI reached spawn while fleeing")
		}
		aiComp.ChangeState(AIStateReturn)
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateReturn.String(),
				"prev_state": AIStateFlee.String(),
			}).Debug("State transition completed")
		}
	}
}

// processReturn handles the return state - go back to spawn point.
func (ai *AISystem) processReturn(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	distance := aiComp.GetDistanceFromSpawn(pos.X, pos.Y)
	ai.logReturnState(entity.ID, aiComp, pos, distance)

	if distance < 10.0 {
		ai.handleReturnComplete(entity, aiComp)
		return
	}

	ai.moveTowards(entity, pos, aiComp.SpawnX, aiComp.SpawnY, aiComp.GetSpeedMultiplier())
}

// logReturnState logs the return state processing details.
func (ai *AISystem) logReturnState(entityID uint64, aiComp *AIComponent, pos *PositionComponent, distance float64) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"state":     aiComp.State.String(),
			"distance":  distance,
			"spawn_x":   aiComp.SpawnX,
			"spawn_y":   aiComp.SpawnY,
			"current_x": pos.X,
			"current_y": pos.Y,
		}).Debug("Processing return state")
	}
}

// handleReturnComplete transitions entity to idle when return is complete.
func (ai *AISystem) handleReturnComplete(entity *Entity, aiComp *AIComponent) {
	ai.logReturnCompletion(entity.ID)
	aiComp.ChangeState(AIStateIdle)
	ai.logStateTransition(entity.ID, AIStateReturn, AIStateIdle, "returned_to_spawn")
	ai.stopEntityMovement(entity)
	ai.setIdleAnimation(entity)
}

// logReturnCompletion logs successful return to spawn point.
func (ai *AISystem) logReturnCompletion(entityID uint64) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entityID,
			"state_transition": "Return->Idle",
		}).Debug("AI returned to spawn")
	}
}

// setIdleAnimation sets entity animation to idle state.
func (ai *AISystem) setIdleAnimation(entity *Entity) {
	if animComp, ok := entity.GetComponent("animation"); ok {
		if anim, ok := animComp.(*AnimationComponent); ok {
			if anim.CurrentState != AnimationStateIdle {
				anim.SetState(AnimationStateIdle)
			}
		}
	}
}

// transitionToFlee switches to flee state and sets up retreat.
func (ai *AISystem) transitionToFlee(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"state_transition": fmt.Sprintf("%s->Flee", aiComp.State.String()),
			"reason":           "low_health",
		}).Warn("AI fleeing due to low health")
	}
	aiComp.ChangeState(AIStateFlee)
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"new_state":  AIStateFlee.String(),
			"prev_state": aiComp.State.String(),
		}).Debug("State transition completed")
	}
}

// shouldFlee checks if the entity should flee based on health.
// Uses typed GetHealth() getter for ~93x faster access vs map lookup + type assertion.
func (ai *AISystem) shouldFlee(entity *Entity, aiComp *AIComponent) bool {
	health := entity.GetHealth()
	if health == nil {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "health",
			}).Debug("Entity missing health component for flee check")
		}
		return false
	}
	if health.Max <= 0 {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"max_health": health.Max,
			}).Warn("Invalid max health for flee check")
		}
		return false
	}

	healthPercent := health.Current / health.Max
	shouldFlee := healthPercent < aiComp.FleeHealthThreshold

	if shouldFlee && ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"health_percent": healthPercent,
			"flee_threshold": aiComp.FleeHealthThreshold,
			"current_health": health.Current,
			"max_health":     health.Max,
		}).Debug("AI health check for flee condition")
	}

	return shouldFlee
}

// findNearestEnemy finds the closest enemy within the detection range.
// Uses spatial partition (quadtree) when available for O(log n) queries,
// otherwise falls back to O(n) iteration over all entities.
func (ai *AISystem) findNearestEnemy(entity *Entity, pos *PositionComponent, detectionRange float64) *Entity {
	team, ok := ai.validateEntityTeam(entity)
	if !ok {
		return nil
	}

	var nearest *Entity
	nearestDist := detectionRange
	candidatesChecked := 0

	// Use spatial partition for efficient enemy detection when available
	var candidates []*Entity
	if ai.quadtree != nil {
		// O(log n) spatial query - reuse buffer to avoid allocations
		ai.queryBuffer = ai.queryBuffer[:0]
		ai.queryBuffer = ai.quadtree.QueryInto(Bounds{
			X:      pos.X - detectionRange,
			Y:      pos.Y - detectionRange,
			Width:  detectionRange * 2,
			Height: detectionRange * 2,
		}, ai.queryBuffer)
		candidates = ai.queryBuffer
	} else {
		// Fallback to O(n) iteration when quadtree not set
		candidates = ai.world.cachedEntityList
	}

	for _, other := range candidates {
		candidatesChecked++

		if !ai.isValidEnemyTarget(entity, other, team) {
			continue
		}

		otherPos, ok := ai.getEntityPosition(other)
		if !ok {
			continue
		}

		dist := ai.getDistance(pos.X, pos.Y, otherPos.X, otherPos.Y)
		if dist < nearestDist {
			nearest = other
			nearestDist = dist
		}
	}

	ai.logEnemyDetection(entity, detectionRange, nearestDist, candidatesChecked, nearest)
	return nearest
}

// validateEntityTeam checks if an entity has a valid team component.
// Uses typed GetTeam() getter for ~93x faster access in AI hot path.
func (ai *AISystem) validateEntityTeam(entity *Entity) (*TeamComponent, bool) {
	team := entity.GetTeam()
	if team == nil {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "team",
			}).Debug("Entity missing team component for enemy detection")
		}
		return nil, false
	}
	return team, true
}

// isValidEnemyTarget checks if an entity is a valid enemy target.
// Uses typed GetTeam() and GetHealth() getters for ~93x faster access.
func (ai *AISystem) isValidEnemyTarget(entity, other *Entity, myTeam *TeamComponent) bool {
	if other == entity {
		return false
	}
	otherTeam := other.GetTeam()
	if otherTeam == nil || !myTeam.IsEnemy(otherTeam.TeamID) {
		return false
	}
	// Use typed getter for health check
	if h := other.GetHealth(); h != nil && h.IsDead() {
		return false
	}
	return true
}

// getEntityPosition retrieves the position component from an entity.
// Uses typed getter for ~94x faster access vs generic GetComponent.
func (ai *AISystem) getEntityPosition(entity *Entity) (*PositionComponent, bool) {
	pos := entity.GetPosition()
	return pos, pos != nil
}

// logEnemyDetection logs the results of an enemy detection scan.
func (ai *AISystem) logEnemyDetection(entity *Entity, detectionRange, nearestDist float64, candidatesChecked int, nearest *Entity) {
	if ai.logger != nil && aiDebugEnabled {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":          entity.ID,
			"detection_range":    detectionRange,
			"candidates_checked": candidatesChecked,
			"enemy_found":        nearest != nil,
			"nearest_distance":   nearestDist,
		}).Debug("Enemy detection scan completed")
	}
}

// isValidTarget checks if a target is still valid (alive, in range, etc.).
// Uses typed GetHealth() and GetPosition() getters for ~93x faster access vs map lookup + type assertion.
func (ai *AISystem) isValidTarget(target, entity *Entity, pos *PositionComponent, maxRange float64) bool {
	if target == nil {
		return false
	}

	// Check if target is alive using typed getter (~93x faster)
	if h := target.GetHealth(); h != nil {
		if h.IsDead() {
			if ai.logger != nil && aiDebugEnabled {
				ai.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"target_id": target.ID,
					"reason":    "target_dead",
				}).Debug("Target validation failed")
			}
			return false
		}
	}

	// Check if target is in range using typed getter (~93x faster)
	targetP := target.GetPosition()
	if targetP == nil {
		return false
	}

	dist := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
	isValid := dist <= maxRange

	if !isValid && ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"target_id": target.ID,
			"distance":  dist,
			"max_range": maxRange,
			"reason":    "out_of_range",
		}).Debug("Target validation failed")
	}

	return isValid
}

// moveTowards moves an entity towards a target position.
func (ai *AISystem) moveTowards(entity *Entity, pos *PositionComponent, targetX, targetY, speedMultiplier float64) {
	vel := ai.getVelocityComponent(entity)
	if vel == nil {
		return
	}

	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0 {
		ai.calculateMovementVelocity(vel, dx, dy, dist, speedMultiplier)
		ai.logMovement(entity, targetX, targetY, dist, speedMultiplier, vel)
		ai.updateEntityDirection(entity, targetX, targetY)
		ai.updateWalkAnimation(entity)
	}
}

// getVelocityComponent retrieves and validates the velocity component.
// Uses cached GetVelocity() getter for ~93x faster access vs map lookup + type assertion.
func (ai *AISystem) getVelocityComponent(entity *Entity) *VelocityComponent {
	vel := entity.GetVelocity()
	if vel == nil && ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "velocity",
		}).Debug("Entity missing velocity component for movement")
	}
	return vel
}

// calculateMovementVelocity calculates and sets the velocity for movement.
func (ai *AISystem) calculateMovementVelocity(vel *VelocityComponent, dx, dy, dist, speedMultiplier float64) {
	speed := 100.0 * speedMultiplier
	vel.VX = (dx / dist) * speed
	vel.VY = (dy / dist) * speed
}

// logMovement logs movement details for debugging.
func (ai *AISystem) logMovement(entity *Entity, targetX, targetY, dist, speedMultiplier float64, vel *VelocityComponent) {
	if ai.logger == nil {
		return
	}
	ai.logger.WithFields(logrus.Fields{
		"entity_id":        entity.ID,
		"target_x":         targetX,
		"target_y":         targetY,
		"distance":         dist,
		"speed":            100.0 * speedMultiplier,
		"speed_multiplier": speedMultiplier,
		"vx":               vel.VX,
		"vy":               vel.VY,
	}).Debug("Entity moving towards target")
}

// updateEntityDirection updates the aim component to face the movement direction.
func (ai *AISystem) updateEntityDirection(entity *Entity, targetX, targetY float64) {
	aimComp, ok := entity.GetComponent("aim")
	if !ok {
		return
	}
	aim, ok := aimComp.(*AimComponent)
	if !ok {
		return
	}
	aim.SetAimTarget(targetX, targetY)
}

// updateWalkAnimation sets the animation state to walk when moving.
func (ai *AISystem) updateWalkAnimation(entity *Entity) {
	animComp, ok := entity.GetComponent("animation")
	if !ok {
		return
	}
	anim, ok := animComp.(*AnimationComponent)
	if !ok {
		return
	}
	if anim.CurrentState != AnimationStateWalk {
		anim.SetState(AnimationStateWalk)
	}
}

// getDistance calculates the distance between two points.
func (ai *AISystem) getDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// SetDetectionRange sets the detection range for all AI entities.
func (ai *AISystem) SetDetectionRange(entity *Entity, detectionRange float64) {
	aiComp, ok := entity.GetComponent("ai")
	if ok {
		if aiC, ok := aiComp.(*AIComponent); ok {
			if ai.logger != nil && aiDebugEnabled {
				ai.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"old_range": aiC.DetectionRange,
					"new_range": detectionRange,
					"operation": "set_detection_range",
				}).Debug("AI detection range updated")
			}
			aiC.DetectionRange = detectionRange
		}
	}
}

// GetState returns the current AI state of an entity.
func (ai *AISystem) GetState(entity *Entity) AIState {
	aiComp, ok := entity.GetComponent("ai")
	if !ok {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "ai",
			}).Debug("Entity missing AI component for state query")
		}
		return AIStateIdle
	}
	if aiC, ok := aiComp.(*AIComponent); ok {
		if ai.logger != nil && aiDebugEnabled {
			ai.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"state":     aiC.State.String(),
			}).Debug("AI state retrieved")
		}
		return aiC.State
	}
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "ai",
		}).Warn("Failed to type assert AI component for state query")
	}
	return AIStateIdle
}
