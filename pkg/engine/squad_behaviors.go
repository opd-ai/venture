// Package engine provides squad-aware behavior tree actions and conditions.
// This file implements tactical behaviors for squad coordination.
package engine

import (
	"math"
)

// NewMaintainFormationAction creates an action that moves towards formation position.
func NewMaintainFormationAction() *ActionNode {
	return NewActionNode("MaintainFormation", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get formation target position from blackboard
		targetX, okX := blackboard.GetFloat64("formation_target_x")
		targetY, okY := blackboard.GetFloat64("formation_target_y")
		if !okX || !okY {
			return NodeFailure // No formation target set
		}

		// Get current position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Calculate distance to formation position
		dx := targetX - pos.X
		dy := targetY - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// If already at formation position (within 5 pixels), success
		if distance < 5.0 {
			// Stop movement
			if velComp, ok := entity.GetComponent("velocity"); ok {
				if vel, ok := velComp.(*VelocityComponent); ok {
					vel.VX = 0
					vel.VY = 0
				}
			}
			return NodeSuccess
		}

		// Move towards formation position
		if velComp, ok := entity.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				speed := 80.0 // Formation movement speed

				// Normalize direction
				vel.VX = (dx / distance) * speed
				vel.VY = (dy / distance) * speed
			}
		}

		return NodeRunning
	})
}

// NewFlankTargetAction creates an action that attempts to flank the current target.
func NewFlankTargetAction(world *World) *ActionNode {
	return NewActionNode("FlankTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get target from blackboard
		target, ok := blackboard.GetEntity("target")
		if !ok || target == nil {
			return NodeFailure
		}

		// Get squad component
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return NodeFailure // Not in a squad
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return NodeFailure
		}

		// Get positions
		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		targetPosComp, ok := target.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		targetPos, ok := targetPosComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Determine flanking position based on squad position index
		// Even indices flank left, odd indices flank right
		flankAngle := math.Pi / 3 // 60 degrees to the side
		if squad.FormationPosition%2 == 1 {
			flankAngle = -flankAngle
		}

		// Calculate flanking position around target
		toTarget := math.Atan2(targetPos.Y-pos.Y, targetPos.X-pos.X)
		flankDirection := toTarget + flankAngle
		flankDistance := 100.0 // Distance from target

		flankX := targetPos.X + math.Cos(flankDirection)*flankDistance
		flankY := targetPos.Y + math.Sin(flankDirection)*flankDistance

		// Move towards flanking position
		dx := flankX - pos.X
		dy := flankY - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < 10.0 {
			// At flanking position
			return NodeSuccess
		}

		if velComp, ok := entity.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				speed := 90.0
				vel.VX = (dx / distance) * speed
				vel.VY = (dy / distance) * speed
			}
		}

		return NodeRunning
	})
}

// NewFocusFireAction creates an action that prioritizes the squad's shared target.
func NewFocusFireAction() *ActionNode {
	return NewActionNode("FocusFire", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get squad component
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return NodeFailure
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return NodeFailure
		}

		// Check if squad has a shared priority target
		priorityTarget, ok := squad.SharedBlackboard.GetEntity("priority_target")
		if !ok || priorityTarget == nil {
			return NodeFailure
		}

		// Set this as our target for focus fire
		blackboard.Set("target", priorityTarget)
		return NodeSuccess
	})
}

// NewCallForHelpAction creates an action that alerts nearby squads.
func NewCallForHelpAction(world *World) *ActionNode {
	return NewActionNode("CallForHelp", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get squad component
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return NodeFailure
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return NodeFailure
		}

		// Check if we can alert (cooldown)
		currentTime, _ := blackboard.GetFloat64("game_time")
		if !squad.CanAlert(currentTime) {
			return NodeFailure
		}

		// Get our position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Find nearby squads within alert range (500 pixels)
		alertRange := 500.0
		allSquadEntities := world.GetEntitiesWith("squad", "position")

		for _, other := range allSquadEntities {
			if other.ID == entity.ID {
				continue
			}

			otherSquadComp, _ := other.GetComponent("squad")
			if otherSquadComp == nil {
				continue
			}
			otherSquad, ok := otherSquadComp.(*SquadComponent)
			if !ok {
				continue
			}

			// Skip same squad
			if otherSquad.SquadID == squad.SquadID {
				continue
			}

			otherPosComp, _ := other.GetComponent("position")
			if otherPosComp == nil {
				continue
			}
			otherPos, ok := otherPosComp.(*PositionComponent)
			if !ok {
				continue
			}

			// Check distance
			dx := otherPos.X - pos.X
			dy := otherPos.Y - pos.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance <= alertRange {
				// Alert this squad by setting alert status in their shared blackboard
				if otherSquad.SharedBlackboard != nil {
					otherSquad.SharedBlackboard.Set("alerted", true)
					otherSquad.SharedBlackboard.Set("alert_position_x", pos.X)
					otherSquad.SharedBlackboard.Set("alert_position_y", pos.Y)
				}
			}
		}

		// Mark as alerted
		squad.SetAlerted(currentTime)
		return NodeSuccess
	})
}

// NewCoordinatedRetreatAction creates an action for squad retreat.
func NewCoordinatedRetreatAction() *ActionNode {
	return NewActionNode("CoordinatedRetreat", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get squad component
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return NodeFailure
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return NodeFailure
		}

		// Check if squad leader ordered retreat
		shouldRetreat, ok := squad.SharedBlackboard.GetBool("retreat_ordered")
		if !ok || !shouldRetreat {
			return NodeFailure
		}

		// Get retreat direction from shared blackboard
		retreatX, okX := squad.SharedBlackboard.GetFloat64("retreat_dir_x")
		retreatY, okY := squad.SharedBlackboard.GetFloat64("retreat_dir_y")
		if !okX || !okY {
			// No retreat direction set, retreat away from average threat position
			return NodeFailure
		}

		// Move in retreat direction
		if velComp, ok := entity.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				speed := 120.0 // Retreat faster

				// Normalize direction
				length := math.Sqrt(retreatX*retreatX + retreatY*retreatY)
				if length > 0 {
					vel.VX = (retreatX / length) * speed
					vel.VY = (retreatY / length) * speed
				}
			}
		}

		return NodeRunning
	})
}

// NewSquadLeaderDecisionAction creates an action for squad leader decision-making.
func NewSquadLeaderDecisionAction() *ActionNode {
	return NewActionNode("SquadLeaderDecision", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get squad component
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return NodeFailure
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok || !squad.IsLeader() {
			return NodeFailure
		}

		// Leader makes tactical decisions for the squad
		// Check if squad should retreat (simple heuristic: if leader health low)
		healthComp, ok := entity.GetComponent("health")
		if ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				healthPercent := float64(health.Current) / float64(health.Max)

				if healthPercent < 0.3 {
					// Order retreat
					squad.SharedBlackboard.Set("retreat_ordered", true)

					// Calculate retreat direction (away from current target)
					if target, ok := blackboard.GetEntity("target"); ok {
						posComp, _ := entity.GetComponent("position")
						targetPosComp, _ := target.GetComponent("position")
						if posComp != nil && targetPosComp != nil {
							if pos, ok := posComp.(*PositionComponent); ok {
								if targetPos, ok := targetPosComp.(*PositionComponent); ok {
									// Retreat direction is away from target
									dx := pos.X - targetPos.X
									dy := pos.Y - targetPos.Y
									squad.SharedBlackboard.Set("retreat_dir_x", dx)
									squad.SharedBlackboard.Set("retreat_dir_y", dy)
								}
							}
						}
					}
				} else {
					// Continue fighting
					squad.SharedBlackboard.Set("retreat_ordered", false)
				}
			}
		}

		// Set priority target if we have one
		if target, ok := blackboard.GetEntity("target"); ok {
			squad.SharedBlackboard.Set("priority_target", target)
		}

		return NodeSuccess
	})
}

// Squad-related Conditions

// NewIsSquadLeaderCondition creates a condition that checks if entity is squad leader.
func NewIsSquadLeaderCondition() *ConditionNode {
	return NewConditionNode("IsSquadLeader", func(entity *Entity, blackboard *Blackboard) bool {
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return false
		}
		return squadComp.(*SquadComponent).IsLeader()
	})
}

// NewSquadHasPriorityTargetCondition creates a condition checking for shared target.
func NewSquadHasPriorityTargetCondition() *ConditionNode {
	return NewConditionNode("SquadHasPriorityTarget", func(entity *Entity, blackboard *Blackboard) bool {
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return false
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return false
		}

		target, ok := squad.SharedBlackboard.GetEntity("priority_target")
		return ok && target != nil
	})
}

// NewRetreatOrderedCondition creates a condition checking for retreat order.
func NewRetreatOrderedCondition() *ConditionNode {
	return NewConditionNode("RetreatOrdered", func(entity *Entity, blackboard *Blackboard) bool {
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			return false
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			return false
		}

		retreat, ok := squad.SharedBlackboard.GetBool("retreat_ordered")
		return ok && retreat
	})
}

// NewHasFormationTargetCondition creates a condition checking for formation target.
func NewHasFormationTargetCondition() *ConditionNode {
	return NewConditionNode("HasFormationTarget", func(entity *Entity, blackboard *Blackboard) bool {
		_, okX := blackboard.GetFloat64("formation_target_x")
		_, okY := blackboard.GetFloat64("formation_target_y")
		return okX && okY
	})
}
