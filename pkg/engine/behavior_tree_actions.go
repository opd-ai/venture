// Package engine provides action and condition nodes for behavior trees.
// This file implements concrete actions and conditions for AI behavior.
package engine

import (
	"fmt"
	"math"
	"math/rand"
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
func (a *ActionNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	return a.action(entity, blackboard, deltaTime)
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
func (c *ConditionNode) Tick(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
	if c.condition(entity, blackboard) {
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
		target, ok := blackboard.GetEntity("target")
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
		health := healthComp.(*HealthComponent)
		return float64(health.Current)/float64(health.Max) < threshold
	})
}

// NewTargetInRangeCondition creates a condition that checks if target is within range.
func NewTargetInRangeCondition(range_ float64) *ConditionNode {
	name := fmt.Sprintf("TargetInRange(%.0f)", range_)
	return NewConditionNode(name, func(entity *Entity, blackboard *Blackboard) bool {
		target, ok := blackboard.GetEntity("target")
		if !ok || target == nil {
			return false
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return false
		}
		pos := posComp.(*PositionComponent)

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return false
		}
		tPos := targetPos.(*PositionComponent)

		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		return distance <= range_
	})
}

// Common Action Builders

// NewFindTargetAction creates an action that searches for the nearest enemy.
func NewFindTargetAction(detectionRange float64, world *World) *ActionNode {
	return NewActionNode("FindTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos := posComp.(*PositionComponent)

		// Get team component to identify enemies
		teamComp, hasTeam := entity.GetComponent("team")
		var teamID int
		if hasTeam {
			teamID = teamComp.(*TeamComponent).TeamID
		}

		// Find nearest enemy
		var nearestEnemy *Entity
		nearestDistance := detectionRange + 1.0

		// Query all entities with position and team components
		if world != nil {
			entities := world.GetEntitiesWith("position", "team")

			for _, other := range entities {
				if other.ID == entity.ID {
					continue
				}

				// Check if enemy (different team)
				otherTeamComp, hasOtherTeam := other.GetComponent("team")
				if !hasOtherTeam {
					continue
				}
				otherTeamID := otherTeamComp.(*TeamComponent).TeamID
				if otherTeamID == teamID {
					continue // Same team
				}

				// Check distance
				otherPos, ok := other.GetComponent("position")
				if !ok {
					continue
				}
				oPos := otherPos.(*PositionComponent)

				dx := oPos.X - pos.X
				dy := oPos.Y - pos.Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance <= detectionRange && distance < nearestDistance {
					nearestEnemy = other
					nearestDistance = distance
				}
			}
		}

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
		target, ok := blackboard.GetEntity("target")
		if !ok || target == nil {
			return NodeFailure
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos := posComp.(*PositionComponent)

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		tPos := targetPos.(*PositionComponent)

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
		vel := velComp.(*VelocityComponent)
		vel.VX = dx
		vel.VY = dy

		return NodeRunning
	})
}

// NewAttackTargetAction creates an action that attacks the target.
func NewAttackTargetAction() *ActionNode {
	return NewActionNode("AttackTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		target, ok := blackboard.GetEntity("target")
		if !ok || target == nil {
			return NodeFailure
		}

		// Check if we have attack component
		attackComp, ok := entity.GetComponent("attack")
		if !ok {
			return NodeFailure
		}
		attack := attackComp.(*AttackComponent)

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
		target, ok := blackboard.GetEntity("target")
		if !ok || target == nil {
			return NodeFailure
		}

		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos := posComp.(*PositionComponent)

		targetPos, ok := target.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		tPos := targetPos.(*PositionComponent)

		// Calculate direction (away from target)
		dx := pos.X - tPos.X
		dy := pos.Y - tPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < 1.0 {
			// Pick random direction if too close
			dx = (rand.Float64() - 0.5) * 2
			dy = (rand.Float64() - 0.5) * 2
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
		vel := velComp.(*VelocityComponent)
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
			// Pick random direction
			angle := rand.Float64() * 2 * math.Pi
			dx := math.Cos(angle) * speed * 0.5 // Wander slower
			dy := math.Sin(angle) * speed * 0.5

			blackboard.Set("wanderDX", dx)
			blackboard.Set("wanderDY", dy)
			blackboard.Set("wanderTime", 2.0+rand.Float64()*3.0) // Wander for 2-5 seconds
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
				vel := velComp.(*VelocityComponent)
				vel.VX = 0
				vel.VY = 0
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
		vel := velComp.(*VelocityComponent)
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
