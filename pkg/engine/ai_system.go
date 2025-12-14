// Package engine provides the AI system for autonomous entity behavior.
// This file implements AISystem which manages state transitions and behaviors
// for AI-controlled entities using a state machine pattern.
package engine

import (
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
)

// AISystem manages artificial intelligence behaviors for entities.
// It implements a state machine that transitions between idle, patrol, chase, attack, and flee states.
type AISystem struct {
	world        *World
	logger       *logrus.Entry
	combatSystem *CombatSystem // Cached combat system to avoid per-attack allocations
}

// NewAISystem creates a new AI system.
func NewAISystem(world *World) *AISystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "ai")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("AI system created")
		}
	}
	return &AISystem{
		world:        world,
		logger:       logEntry,
		combatSystem: NewCombatSystem(12345), // Pre-allocate combat system to avoid per-attack allocations
	}
}

// Update processes AI behavior for all entities with AI components.
func (ai *AISystem) Update(entities []*Entity, deltaTime float64) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("AI system update started")
	}

	for _, entity := range entities {
		// Check if entity has AI component
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
	// Get position component
	posComp, ok := entity.GetComponent("position")
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "position",
			}).Debug("Entity missing position component for AI processing")
		}
		return // Can't do AI without position
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Check health for flee condition
	shouldFlee := ai.shouldFlee(entity, aiComp)

	if ai.logger != nil {
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
	if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"target_id":  target.ID,
				"new_state":  AIStateDetect.String(),
				"prev_state": AIStateIdle.String(),
			}).Debug("State transition completed")
		}
	} else {
		if ai.logger != nil {
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
	if ai.logger != nil {
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

	if ai.logger != nil {
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

	if ai.logger != nil {
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

	if ai.logger != nil {
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
func (ai *AISystem) stopEntityMovement(entity *Entity) {
	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			vel.VX = 0
			vel.VY = 0
		}
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

	if ai.logger != nil {
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
func (ai *AISystem) moveTowardWaypoint(entity *Entity, aiComp *AIComponent, pos *PositionComponent, waypoint *PatrolWaypoint) {
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return
	}

	vel, ok := velComp.(*VelocityComponent)
	if !ok {
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
	if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Detect->Idle",
				"reason":           "target_invalid",
			}).Debug("AI lost target during detection")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		if ai.logger != nil {
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
		if ai.logger != nil {
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
	if ai.logger != nil {
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
		if ai.logger != nil {
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
		if ai.logger != nil {
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
func (ai *AISystem) getTargetPositionSimple(target *Entity) *PositionComponent {
	targetPos, ok := target.GetComponent("position")
	if !ok {
		return nil
	}
	pos, ok := targetPos.(*PositionComponent)
	if !ok {
		return nil
	}
	return pos
}

// logStateTransition logs AI state transition with reason.
func (ai *AISystem) logStateTransition(entityID uint64, from, to AIState, reason string) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":        entityID,
			"state_transition": from.String() + "->" + to.String(),
			"reason":           reason,
		}).Debug("AI state transition")
	}
}

// logTransitionComplete logs completion of state transition.
func (ai *AISystem) logTransitionComplete(entityID uint64, from, to AIState) {
	if ai.logger != nil {
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
	if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"target_id": aiComp.Target.ID,
				"reason":    "target_invalid",
			}).Debug("Target invalid during attack")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		if ai.logger != nil {
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
		if ai.logger != nil {
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
func (ai *AISystem) getTargetPosition(entity *Entity, aiComp *AIComponent) *PositionComponent {
	targetPos, ok := aiComp.Target.GetComponent("position")
	if !ok {
		ai.logTargetMissingComponent(entity.ID, aiComp.Target.ID)
		ai.clearTargetAndReturnToIdle(entity, aiComp)
		return nil
	}
	targetP, ok := targetPos.(*PositionComponent)
	if !ok {
		ai.logTargetTypeAssertFailed(entity.ID, aiComp.Target.ID)
		ai.clearTargetAndReturnToIdle(entity, aiComp)
		return nil
	}
	return targetP
}

// logTargetMissingComponent logs when target is missing position component.
func (ai *AISystem) logTargetMissingComponent(entityID, targetID uint64) {
	if ai.logger != nil {
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
	if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        aiComp.Target.ID,
				"distance":         distance,
				"attack_range":     attackRange,
				"state_transition": "Attack->Chase",
			}).Debug("Target moved out of attack range")
		}
		aiComp.ChangeState(AIStateChase)
		if ai.logger != nil {
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
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"target_id": aiComp.Target.ID,
		}).Debug("Attack executed")
	}
}

// logAttackOnCooldown logs when attack is on cooldown.
func (ai *AISystem) logAttackOnCooldown(entityID, targetID uint64, cooldown float64) {
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"target_id": targetID,
			"cooldown":  cooldown,
		}).Debug("Attack on cooldown")
	}
}

// processFlee handles the flee state - run away from target.
func (ai *AISystem) processFlee(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	if ai.logger != nil {
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
		if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Flee->Return",
				"reason":           "reached_spawn",
			}).Debug("AI reached spawn while fleeing")
		}
		aiComp.ChangeState(AIStateReturn)
		if ai.logger != nil {
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

	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"state":     aiComp.State.String(),
			"distance":  distance,
			"spawn_x":   aiComp.SpawnX,
			"spawn_y":   aiComp.SpawnY,
			"current_x": pos.X,
			"current_y": pos.Y,
		}).Debug("Processing return state")
	}

	// If close enough to spawn, go idle
	if distance < 10.0 {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Return->Idle",
				"distance":         distance,
			}).Debug("AI returned to spawn")
		}
		aiComp.ChangeState(AIStateIdle)
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"new_state":  AIStateIdle.String(),
				"prev_state": AIStateReturn.String(),
			}).Debug("State transition completed")
		}
		// Stop movement
		velComp, ok := entity.GetComponent("velocity")
		if ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX = 0
				vel.VY = 0
			}
		}

		// GAP-018 REPAIR: Set animation to idle when stopped
		if animComp, ok := entity.GetComponent("animation"); ok {
			if anim, ok := animComp.(*AnimationComponent); ok {
				if anim.CurrentState != AnimationStateIdle {
					anim.SetState(AnimationStateIdle)
				}
			}
		}
		return
	}

	// Move towards spawn
	ai.moveTowards(entity, pos, aiComp.SpawnX, aiComp.SpawnY, aiComp.GetSpeedMultiplier())
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
	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"new_state":  AIStateFlee.String(),
			"prev_state": aiComp.State.String(),
		}).Debug("State transition completed")
	}
}

// shouldFlee checks if the entity should flee based on health.
func (ai *AISystem) shouldFlee(entity *Entity, aiComp *AIComponent) bool {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "health",
			}).Debug("Entity missing health component for flee check")
		}
		return false
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "health",
			}).Warn("Failed to type assert health component for flee check")
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
func (ai *AISystem) findNearestEnemy(entity *Entity, pos *PositionComponent, detectionRange float64) *Entity {
	team, ok := ai.validateEntityTeam(entity)
	if !ok {
		return nil
	}

	var nearest *Entity
	nearestDist := detectionRange
	candidatesChecked := 0

	// Iterate over cached entity list (slice) instead of entities map for better performance
	// Slice iteration is significantly faster than map iteration
	for _, other := range ai.world.cachedEntityList {
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
func (ai *AISystem) validateEntityTeam(entity *Entity) (*TeamComponent, bool) {
	teamComp, ok := entity.GetComponent("team")
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "team",
			}).Debug("Entity missing team component for enemy detection")
		}
		return nil, false
	}
	team, ok := teamComp.(*TeamComponent)
	return team, ok
}

// isValidEnemyTarget checks if an entity is a valid enemy target.
func (ai *AISystem) isValidEnemyTarget(entity, other *Entity, myTeam *TeamComponent) bool {
	if other == entity {
		return false
	}
	otherTeamComp, ok := other.GetComponent("team")
	if !ok {
		return false
	}
	otherTeam, ok := otherTeamComp.(*TeamComponent)
	if !ok || !myTeam.IsEnemy(otherTeam.TeamID) {
		return false
	}
	if otherHealth, ok := other.GetComponent("health"); ok {
		if h, ok := otherHealth.(*HealthComponent); ok && h.IsDead() {
			return false
		}
	}
	return true
}

// getEntityPosition retrieves the position component from an entity.
func (ai *AISystem) getEntityPosition(entity *Entity) (*PositionComponent, bool) {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	return pos, ok
}

// logEnemyDetection logs the results of an enemy detection scan.
func (ai *AISystem) logEnemyDetection(entity *Entity, detectionRange, nearestDist float64, candidatesChecked int, nearest *Entity) {
	if ai.logger != nil {
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
func (ai *AISystem) isValidTarget(target, entity *Entity, pos *PositionComponent, maxRange float64) bool {
	if target == nil {
		return false
	}

	// Check if target is alive
	targetHealth, ok := target.GetComponent("health")
	if ok {
		if h, ok := targetHealth.(*HealthComponent); ok {
			if h.IsDead() {
				if ai.logger != nil {
					ai.logger.WithFields(logrus.Fields{
						"entity_id": entity.ID,
						"target_id": target.ID,
						"reason":    "target_dead",
					}).Debug("Target validation failed")
				}
				return false
			}
		}
	}

	// Check if target is in range
	targetPos, ok := target.GetComponent("position")
	if !ok {
		return false
	}
	targetP, ok := targetPos.(*PositionComponent)
	if !ok {
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
func (ai *AISystem) getVelocityComponent(entity *Entity) *VelocityComponent {
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "velocity",
			}).Debug("Entity missing velocity component for movement")
		}
		return nil
	}
	vel, ok := velComp.(*VelocityComponent)
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "velocity",
			}).Warn("Failed to type assert velocity component")
		}
		return nil
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
			if ai.logger != nil {
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
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "ai",
			}).Debug("Entity missing AI component for state query")
		}
		return AIStateIdle
	}
	if aiC, ok := aiComp.(*AIComponent); ok {
		if ai.logger != nil {
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
