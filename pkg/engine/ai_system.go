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
	world  *World
	logger *logrus.Entry
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
		world:  world,
		logger: logEntry,
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
	}
}

// processPatrol handles the patrol state - move between waypoints.
func (ai *AISystem) processPatrol(entity *Entity, aiComp *AIComponent, pos *PositionComponent, deltaTime float64) {
	// Look for enemies in range
	target := ai.findNearestEnemy(entity, pos, aiComp.DetectionRange)

	if target != nil {
		aiComp.Target = target
		aiComp.ChangeState(AIStateDetect)
		return
	}

	// Check if patrol route is configured
	if !aiComp.HasPatrolRoute() {
		// No patrol route, behave like idle
		return
	}

	// Get current waypoint
	waypoint := aiComp.GetCurrentWaypoint()
	if waypoint == nil {
		return
	}

	// Check if waiting at waypoint
	if aiComp.IsWaitingAtWaypoint(deltaTime) {
		// Stop movement while waiting
		if velComp, ok := entity.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX = 0
				vel.VY = 0
			}
		}
		return
	}

	// Calculate distance to waypoint
	dx := waypoint.X - pos.X
	dy := waypoint.Y - pos.Y
	distToWaypoint := math.Sqrt(dx*dx + dy*dy)

	// Check if reached waypoint
	if distToWaypoint <= aiComp.WaypointReachDistance {
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
		return
	}

	// Move towards waypoint
	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			// Use default base speed (velocity component doesn't store speed)
			baseSpeed := 100.0 // Default pixels per second

			// Normalize direction and apply patrol speed multiplier
			if distToWaypoint > 0 {
				dirX := dx / distToWaypoint
				dirY := dy / distToWaypoint
				speed := baseSpeed * aiComp.GetSpeedMultiplier()

				vel.VX = dirX * speed
				vel.VY = dirY * speed
			}
		}
	}
}

// processDetect handles the detect state - confirm target and start chase.
func (ai *AISystem) processDetect(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
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
	}
}

// processChase handles the chase state - pursue the target.
func (ai *AISystem) processChase(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	// Verify target is still valid
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.5) {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"state_transition": "Chase->Return",
				"reason":           "target_invalid",
			}).Debug("AI lost target during chase")
		}
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		return
	}

	// Check if too far from spawn
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
		return
	}

	// Get attack component to check range
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return
	}

	// Check if in attack range
	targetPos, ok := aiComp.Target.GetComponent("position")
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}
	targetP, ok := targetPos.(*PositionComponent)
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}

	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)

	if distance <= attack.Range {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        aiComp.Target.ID,
				"distance":         distance,
				"attack_range":     attack.Range,
				"state_transition": "Chase->Attack",
			}).Info("AI in attack range")
		}
		aiComp.ChangeState(AIStateAttack)
		return
	}

	// Move towards target
	ai.moveTowards(entity, pos, targetP.X, targetP.Y, aiComp.GetSpeedMultiplier())
}

// processAttack handles the attack state - attack the target.
func (ai *AISystem) processAttack(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	// Verify target is still valid
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.5) {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		return
	}

	// Get attack component
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return
	}

	// Check if in attack range
	targetPos, ok := aiComp.Target.GetComponent("position")
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}
	targetP, ok := targetPos.(*PositionComponent)
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}

	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)

	// If target moved out of range, chase again
	if distance > attack.Range {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":        entity.ID,
				"target_id":        aiComp.Target.ID,
				"distance":         distance,
				"attack_range":     attack.Range,
				"state_transition": "Attack->Chase",
			}).Debug("Target moved out of attack range")
		}
		aiComp.ChangeState(AIStateChase)
		return
	}

	// Attack if cooldown is ready
	if attack.CanAttack() {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"target_id": aiComp.Target.ID,
				"distance":  distance,
			}).Info("AI executing attack")
		}
		// GAP-018 REPAIR: Set animation to attack when attacking
		if animComp, ok := entity.GetComponent("animation"); ok {
			if anim, ok := animComp.(*AnimationComponent); ok {
				if anim.CurrentState != AnimationStateAttack {
					anim.SetState(AnimationStateAttack)
				}
			}
		}

		// Create a combat system to perform the attack
		// In a real game, this would be a separate system call
		combatSystem := NewCombatSystem(12345) // Use fixed seed for now
		combatSystem.Attack(entity, aiComp.Target)
	}
}

// processFlee handles the flee state - run away from target.
func (ai *AISystem) processFlee(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
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
	}
}

// processReturn handles the return state - go back to spawn point.
func (ai *AISystem) processReturn(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	distance := aiComp.GetDistanceFromSpawn(pos.X, pos.Y)

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
}

// shouldFlee checks if the entity should flee based on health.
func (ai *AISystem) shouldFlee(entity *Entity, aiComp *AIComponent) bool {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return false
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return false
	}
	if health.Max <= 0 {
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
	teamComp, ok := entity.GetComponent("team")
	if !ok {
		if ai.logger != nil {
			ai.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "team",
			}).Debug("Entity missing team component for enemy detection")
		}
		return nil // No team component, can't determine enemies
	}
	team, ok := teamComp.(*TeamComponent)
	if !ok {
		return nil
	}

	var nearest *Entity
	nearestDist := detectionRange
	candidatesChecked := 0

	for _, other := range ai.world.entities {
		candidatesChecked++
		if other == entity {
			continue
		}

		// Check if other is an enemy
		otherTeam, ok := other.GetComponent("team")
		if !ok {
			continue
		}
		otherT, ok := otherTeam.(*TeamComponent)
		if !ok {
			continue
		}

		if !team.IsEnemy(otherT.TeamID) {
			continue
		}

		// Check if alive
		otherHealth, ok := other.GetComponent("health")
		if ok {
			if h, ok := otherHealth.(*HealthComponent); ok {
				if h.IsDead() {
					continue
				}
			}
		}

		// Check distance
		otherPos, ok := other.GetComponent("position")
		if !ok {
			continue
		}
		otherP, ok := otherPos.(*PositionComponent)
		if !ok {
			continue
		}

		dist := ai.getDistance(pos.X, pos.Y, otherP.X, otherP.Y)
		if dist < nearestDist {
			nearest = other
			nearestDist = dist
		}
	}

	if ai.logger != nil {
		ai.logger.WithFields(logrus.Fields{
			"entity_id":          entity.ID,
			"detection_range":    detectionRange,
			"candidates_checked": candidatesChecked,
			"enemy_found":        nearest != nil,
			"nearest_distance":   nearestDist,
		}).Debug("Enemy detection scan completed")
	}

	return nearest
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
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return // No velocity component, can't move
	}
	vel, ok := velComp.(*VelocityComponent)
	if !ok {
		return
	}

	// Calculate direction
	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0 {
		// Normalize and apply speed (use fixed speed since VelocityComponent doesn't have MaxSpeed)
		speed := 100.0 * speedMultiplier // Default speed
		vel.VX = (dx / dist) * speed
		vel.VY = (dy / dist) * speed

		// Phase 10.1: Update aim component to face movement direction
		// This makes enemies visibly rotate towards their target
		if aimComp, ok := entity.GetComponent("aim"); ok {
			if aim, ok := aimComp.(*AimComponent); ok {
				aim.SetAimTarget(targetX, targetY)
			}
		}

		// GAP-018 REPAIR: Update animation state to walk when moving
		if animComp, ok := entity.GetComponent("animation"); ok {
			if anim, ok := animComp.(*AnimationComponent); ok {
				if anim.CurrentState != AnimationStateWalk {
					anim.SetState(AnimationStateWalk)
				}
			}
		}
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
		return AIStateIdle
	}
	if aiC, ok := aiComp.(*AIComponent); ok {
		return aiC.State
	}
	return AIStateIdle
}
