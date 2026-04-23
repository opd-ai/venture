// Package engine provides action and condition nodes for behavior trees.
// This file implements concrete actions and conditions for AI behavior.
package engine

import (
	"fmt"
	"math"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

// ActionNode is a leaf node that performs an action.
type ActionNode struct {
	name   string
	action func(*Entity, *Blackboard, float64) NodeStatus
}

// NewActionNode creates a new action node.
func NewActionNode(name string, action func(*Entity, *Blackboard, float64) NodeStatus) *ActionNode {
	return &ActionNode{
		name:   name,
		action: action,
	}
}

// Tick executes the action.
// The entity parameter satisfies aitypes.EntityContext; the concrete *Entity is
// recovered via type assertion so the stored closure keeps its *Entity signature.
func (a *ActionNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entity.(*Entity)
	if !ok {
		return NodeFailure
	}
	return a.action(e, blackboard, deltaTime)
}

// Reset does nothing for action nodes.
func (a *ActionNode) Reset() {}

// String returns a string representation of the node.
func (a *ActionNode) String() string {
	return fmt.Sprintf("Action(%s)", a.name)
}

// ConditionNode is a leaf node that checks a condition.
type ConditionNode struct {
	name      string
	condition func(*Entity, *Blackboard) bool
}

// NewConditionNode creates a new condition node.
func NewConditionNode(name string, condition func(*Entity, *Blackboard) bool) *ConditionNode {
	return &ConditionNode{
		name:      name,
		condition: condition,
	}
}

// Tick checks the condition.
// The entity parameter satisfies aitypes.EntityContext; the concrete *Entity is
// recovered via type assertion so the stored closure keeps its *Entity signature.
func (c *ConditionNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entity.(*Entity)
	if !ok {
		return NodeFailure
	}
	if c.condition(e, blackboard) {
		return NodeSuccess
	}
	return NodeFailure
}

// Reset does nothing for condition nodes.
func (c *ConditionNode) Reset() {}

// String returns a string representation of the node.
func (c *ConditionNode) String() string {
	return fmt.Sprintf("Condition(%s)", c.name)
}

// Common Condition Builders

// NewHasTargetCondition creates a condition that checks if entity has a target.
func NewHasTargetCondition() *ConditionNode {
	return NewConditionNode("HasTarget", func(entity *Entity, blackboard *Blackboard) bool {
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		return ok && target != nil
	})
}

// NewHealthBelowCondition creates a condition that checks if health is below threshold.
func NewHealthBelowCondition(threshold float64) *ConditionNode {
	name := fmt.Sprintf("HealthBelow(%.2f)", threshold)
	return NewConditionNode(name, func(entity *Entity, blackboard *Blackboard) bool {
		healthComp, ok := entity.GetComponent("health")
		if !ok {
			return false
		}
		health, ok := healthComp.(*HealthComponent)
		if !ok {
			return false
		}
		return float64(health.Current)/float64(health.Max) < threshold
	})
}

// NewTargetInRangeCondition creates a condition that checks if target is within range.
func NewTargetInRangeCondition(range_ float64) *ConditionNode {
	name := fmt.Sprintf("TargetInRange(%.0f)", range_)
	return NewConditionNode(name, func(entity *Entity, blackboard *Blackboard) bool {
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		if !ok || target == nil {
			return false
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return false
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return false
		}

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return false
		}
		tPos, ok := targetPos.(*PositionComponent)
		if !ok {
			return false
		}

		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		return distance <= range_
	})
}

// Common Action Builders

// getEntityTeamID extracts team ID from entity.
func getEntityTeamID(entity *Entity) (int, bool) {
	teamComp, hasTeam := entity.GetComponent("team")
	if !hasTeam {
		return 0, false
	}
	if team, ok := teamComp.(*TeamComponent); ok {
		return team.TeamID, true
	}
	return 0, false
}

// isEnemyEntity checks if other entity is an enemy to the given team.
func isEnemyEntity(other *Entity, teamID int) bool {
	otherTeamID, ok := getEntityTeamID(other)
	if !ok {
		return false
	}
	return otherTeamID != teamID
}

// calculateEntityDistance computes distance between two position components.
func calculateEntityDistance(pos1, pos2 *PositionComponent) float64 {
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// findNearestEnemy finds the nearest enemy entity within detection range.
func findNearestEnemy(entity *Entity, pos *PositionComponent, teamID int, detectionRange float64, world *World) (*Entity, float64) {
	var nearestEnemy *Entity
	nearestDistance := detectionRange + 1.0

	if world == nil {
		return nil, nearestDistance
	}

	entities := world.GetEntitiesWith("position", "team")
	for _, other := range entities {
		if other.ID == entity.ID {
			continue
		}

		if !isEnemyEntity(other, teamID) {
			continue
		}

		otherPosComp, ok := other.GetComponent("position")
		if !ok {
			continue
		}
		oPos, ok := otherPosComp.(*PositionComponent)
		if !ok {
			continue
		}

		distance := calculateEntityDistance(pos, oPos)
		if distance <= detectionRange && distance < nearestDistance {
			nearestEnemy = other
			nearestDistance = distance
		}
	}

	return nearestEnemy, nearestDistance
}

// NewFindTargetAction creates an action that searches for the nearest enemy.
func NewFindTargetAction(detectionRange float64, world *World) *ActionNode {
	return NewActionNode("FindTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		teamID, _ := getEntityTeamID(entity)
		nearestEnemy, nearestDistance := findNearestEnemy(entity, pos, teamID, detectionRange, world)

		if nearestEnemy != nil {
			blackboard.Set("target", nearestEnemy)
			blackboard.Set("targetDistance", nearestDistance)
			return NodeSuccess
		}

		return NodeFailure
	})
}

// NewMoveToTargetAction creates an action that moves towards the target.
func NewMoveToTargetAction(speed float64) *ActionNode {
	return NewActionNode("MoveToTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		if !ok || target == nil {
			return NodeFailure
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		tPos, ok := targetPos.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Calculate direction
		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < 1.0 {
			return NodeSuccess // Already at target
		}

		// Normalize and apply speed
		dx = (dx / distance) * speed
		dy = (dy / distance) * speed

		// Update velocity
		velComp, ok := entity.GetComponent("velocity")
		if !ok {
			return NodeFailure
		}
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			return NodeFailure
		}
		vel.VX = dx
		vel.VY = dy

		return NodeRunning
	})
}

// NewAttackTargetAction creates an action that attacks the target.
func NewAttackTargetAction() *ActionNode {
	return NewActionNode("AttackTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		if !ok || target == nil {
			return NodeFailure
		}

		// Check if we have attack component
		attackComp, ok := entity.GetComponent("attack")
		if !ok {
			return NodeFailure
		}
		attack, ok := attackComp.(*AttackComponent)
		if !ok {
			return NodeFailure
		}

		// Check attack cooldown
		if attack.CooldownTimer > 0 {
			attack.CooldownTimer -= deltaTime
			return NodeRunning
		}

		// Perform attack
		attack.CooldownTimer = attack.Cooldown

		// Set attack target (combat system will handle damage)
		blackboard.Set("attackTarget", target)

		return NodeSuccess
	})
}

// NewFleeFromTargetAction creates an action that moves away from the target.
func NewFleeFromTargetAction(speed float64) *ActionNode {
	return NewActionNode("FleeFromTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		if !ok || target == nil {
			return NodeFailure
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		tPos, ok := targetPos.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Calculate direction (away from target)
		dx := pos.X - tPos.X
		dy := pos.Y - tPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < 1.0 {
			// Pick random direction if too close using seeded RNG
			rng := blackboard.GetRNG()
			dx = (rng.Float64() - 0.5) * 2
			dy = (rng.Float64() - 0.5) * 2
			distance = 1.0
		}

		// Normalize and apply speed
		dx = (dx / distance) * speed * 1.5 // Flee faster
		dy = (dy / distance) * speed * 1.5

		// Update velocity
		velComp, ok := entity.GetComponent("velocity")
		if !ok {
			return NodeFailure
		}
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			return NodeFailure
		}
		vel.VX = dx
		vel.VY = dy

		return NodeRunning
	})
}

// NewWanderAction creates an action that makes the entity wander randomly.
func NewWanderAction(speed float64) *ActionNode {
	return NewActionNode("Wander", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Check if we have a wander target
		hasWanderTarget, ok := blackboard.GetBool("hasWanderTarget")
		if !ok || !hasWanderTarget {
			// Pick random direction using seeded RNG
			rng := blackboard.GetRNG()
			angle := rng.Float64() * 2 * math.Pi
			dx := math.Cos(angle) * speed * 0.5 // Wander slower
			dy := math.Sin(angle) * speed * 0.5

			blackboard.Set("wanderDX", dx)
			blackboard.Set("wanderDY", dy)
			blackboard.Set("wanderTime", 2.0+rng.Float64()*3.0) // Wander for 2-5 seconds
			blackboard.Set("hasWanderTarget", true)
		}

		// Update wander time
		wanderTime, _ := blackboard.GetFloat64("wanderTime")
		wanderTime -= deltaTime
		blackboard.Set("wanderTime", wanderTime)

		if wanderTime <= 0 {
			// Stop wandering
			blackboard.Set("hasWanderTarget", false)
			velComp, ok := entity.GetComponent("velocity")
			if ok {
				if vel, ok := velComp.(*VelocityComponent); ok {
					vel.VX = 0
					vel.VY = 0
				}
			}
			return NodeSuccess
		}

		// Apply wander velocity
		dx, _ := blackboard.GetFloat64("wanderDX")
		dy, _ := blackboard.GetFloat64("wanderDY")

		velComp, ok := entity.GetComponent("velocity")
		if !ok {
			return NodeFailure
		}
		vel, ok := velComp.(*VelocityComponent)
		if !ok {
			return NodeFailure
		}
		vel.VX = dx
		vel.VY = dy

		return NodeRunning
	})
}

// NewWaitAction creates an action that waits for a specified duration.
func NewWaitAction(duration float64) *ActionNode {
	return NewActionNode("Wait", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Check if we have started waiting
		waitTimer, ok := blackboard.GetFloat64("waitTimer")
		if !ok {
			blackboard.Set("waitTimer", duration)
			return NodeRunning
		}

		waitTimer -= deltaTime
		if waitTimer <= 0 {
			blackboard.Set("waitTimer", 0.0)
			return NodeSuccess
		}

		blackboard.Set("waitTimer", waitTimer)
		return NodeRunning
	})
}

// NewClearTargetAction creates an action that clears the current target.
func NewClearTargetAction() *ActionNode {
	return NewActionNode("ClearTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		blackboard.Set("target", nil)
		return NodeSuccess
	})
}
