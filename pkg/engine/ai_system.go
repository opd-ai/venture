// Package engine provides the AI system for autonomous entity behavior.
// This file implements AISystem which manages state transitions and behaviors
// for AI-controlled entities using a state machine pattern.
package engine

import (
	"math"
)

// AISystem manages artificial intelligence behaviors for entities.
// It implements a state machine that transitions between idle, patrol, chase, attack, and flee states.
type AISystem struct {
	world *World
}

// NewAISystem creates a new AI system.
func NewAISystem(world *World) *AISystem {
	return &AISystem{
		world: world,
	}
}

// Update processes AI behavior for all entities with AI components.
func (ai *AISystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has AI component
		aiComp, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}

		aiState := aiComp.(*AIComponent)

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
		return // Can't do AI without position
	}
	pos := posComp.(*PositionComponent)

	// Check health for flee condition
	shouldFlee := ai.shouldFlee(entity, aiComp)

	// State machine logic
	switch aiComp.State {
	case AIStateIdle:
		ai.processIdle(entity, aiComp, pos)

	case AIStatePatrol:
		ai.processPatrol(entity, aiComp, pos)

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
		aiComp.Target = target
		aiComp.ChangeState(AIStateDetect)
	}
}

// processPatrol handles the patrol state - similar to idle for now.
func (ai *AISystem) processPatrol(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	// Look for enemies in range
	target := ai.findNearestEnemy(entity, pos, aiComp.DetectionRange)

	if target != nil {
		aiComp.Target = target
		aiComp.ChangeState(AIStateDetect)
	}

	// TODO: Implement actual patrol movement along a route
}

// processDetect handles the detect state - confirm target and start chase.
func (ai *AISystem) processDetect(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	// Check if target is still valid and in range
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.2) {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}

	// Transition to chase after brief detection period
	if aiComp.StateTimer > 0.3 {
		aiComp.ChangeState(AIStateChase)
	}
}

// processChase handles the chase state - pursue the target.
func (ai *AISystem) processChase(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	// Verify target is still valid
	if !ai.isValidTarget(aiComp.Target, entity, pos, aiComp.DetectionRange*1.5) {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		return
	}

	// Check if too far from spawn
	if aiComp.ShouldReturnToSpawn(pos.X, pos.Y) {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		return
	}

	// Get attack component to check range
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return
	}
	attack := attackComp.(*AttackComponent)

	// Check if in attack range
	targetPos, ok := aiComp.Target.GetComponent("position")
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}
	targetP := targetPos.(*PositionComponent)

	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)

	if distance <= attack.Range {
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
	attack := attackComp.(*AttackComponent)

	// Check if in attack range
	targetPos, ok := aiComp.Target.GetComponent("position")
	if !ok {
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateIdle)
		return
	}
	targetP := targetPos.(*PositionComponent)

	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)

	// If target moved out of range, chase again
	if distance > attack.Range {
		aiComp.ChangeState(AIStateChase)
		return
	}

	// Attack if cooldown is ready
	if attack.CanAttack() {
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
		// Health recovered, go back to idle
		aiComp.ClearTarget()
		aiComp.ChangeState(AIStateReturn)
		return
	}

	// Move away from target towards spawn
	ai.moveTowards(entity, pos, aiComp.SpawnX, aiComp.SpawnY, aiComp.GetSpeedMultiplier())

	// If close to spawn, switch to return state
	if aiComp.GetDistanceFromSpawn(pos.X, pos.Y) < 20.0 {
		aiComp.ChangeState(AIStateReturn)
	}
}

// processReturn handles the return state - go back to spawn point.
func (ai *AISystem) processReturn(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	distance := aiComp.GetDistanceFromSpawn(pos.X, pos.Y)

	// If close enough to spawn, go idle
	if distance < 10.0 {
		aiComp.ChangeState(AIStateIdle)
		// Stop movement
		velComp, ok := entity.GetComponent("velocity")
		if ok {
			vel := velComp.(*VelocityComponent)
			vel.VX = 0
			vel.VY = 0
		}
		return
	}

	// Move towards spawn
	ai.moveTowards(entity, pos, aiComp.SpawnX, aiComp.SpawnY, aiComp.GetSpeedMultiplier())
}

// transitionToFlee switches to flee state and sets up retreat.
func (ai *AISystem) transitionToFlee(entity *Entity, aiComp *AIComponent, pos *PositionComponent) {
	aiComp.ChangeState(AIStateFlee)
}

// shouldFlee checks if the entity should flee based on health.
func (ai *AISystem) shouldFlee(entity *Entity, aiComp *AIComponent) bool {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return false
	}

	health := healthComp.(*HealthComponent)
	if health.Max <= 0 {
		return false
	}

	healthPercent := health.Current / health.Max
	return healthPercent < aiComp.FleeHealthThreshold
}

// findNearestEnemy finds the closest enemy within the detection range.
func (ai *AISystem) findNearestEnemy(entity *Entity, pos *PositionComponent, detectionRange float64) *Entity {
	teamComp, ok := entity.GetComponent("team")
	if !ok {
		return nil // No team component, can't determine enemies
	}
	team := teamComp.(*TeamComponent)

	var nearest *Entity
	nearestDist := detectionRange

	for _, other := range ai.world.entities {
		if other == entity {
			continue
		}

		// Check if other is an enemy
		otherTeam, ok := other.GetComponent("team")
		if !ok {
			continue
		}
		otherT := otherTeam.(*TeamComponent)

		if !team.IsEnemy(otherT.TeamID) {
			continue
		}

		// Check if alive
		otherHealth, ok := other.GetComponent("health")
		if ok {
			h := otherHealth.(*HealthComponent)
			if h.IsDead() {
				continue
			}
		}

		// Check distance
		otherPos, ok := other.GetComponent("position")
		if !ok {
			continue
		}
		otherP := otherPos.(*PositionComponent)

		dist := ai.getDistance(pos.X, pos.Y, otherP.X, otherP.Y)
		if dist < nearestDist {
			nearest = other
			nearestDist = dist
		}
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
		h := targetHealth.(*HealthComponent)
		if h.IsDead() {
			return false
		}
	}

	// Check if target is in range
	targetPos, ok := target.GetComponent("position")
	if !ok {
		return false
	}
	targetP := targetPos.(*PositionComponent)

	dist := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
	return dist <= maxRange
}

// moveTowards moves an entity towards a target position.
func (ai *AISystem) moveTowards(entity *Entity, pos *PositionComponent, targetX, targetY, speedMultiplier float64) {
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return // No velocity component, can't move
	}
	vel := velComp.(*VelocityComponent)

	// Calculate direction
	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0 {
		// Normalize and apply speed (use fixed speed since VelocityComponent doesn't have MaxSpeed)
		speed := 100.0 * speedMultiplier // Default speed
		vel.VX = (dx / dist) * speed
		vel.VY = (dy / dist) * speed
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
		aiC := aiComp.(*AIComponent)
		aiC.DetectionRange = detectionRange
	}
}

// GetState returns the current AI state of an entity.
func (ai *AISystem) GetState(entity *Entity) AIState {
	aiComp, ok := entity.GetComponent("ai")
	if !ok {
		return AIStateIdle
	}
	return aiComp.(*AIComponent).State
}
