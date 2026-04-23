// Package engine provides squad-aware behavior tree actions and conditions.
// This file implements tactical behaviors for squad coordination.
package engine

import (
	"fmt"
	"math"
)

// NewMaintainFormationAction creates an action that moves towards formation position.
func NewMaintainFormationAction() *ActionNode {
	return NewActionNode("MaintainFormation", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		targetX, targetY, ok := getFormationTarget(blackboard)
		if !ok {
			return NodeFailure
		}

		pos, ok := getEntityPosition(entity)
		if !ok {
			return NodeFailure
		}

		if isAtFormationPosition(pos, targetX, targetY) {
			stopEntityMovement(entity)
			return NodeSuccess
		}

		moveToFormation(entity, pos, targetX, targetY)
		return NodeRunning
	})
}

// getFormationTarget retrieves formation target coordinates from blackboard.
func getFormationTarget(blackboard *Blackboard) (float64, float64, bool) {
	targetX, okX := blackboard.GetFloat64("formation_target_x")
	targetY, okY := blackboard.GetFloat64("formation_target_y")
	return targetX, targetY, okX && okY
}

// getEntityPosition retrieves and validates the position component for formations.
func getEntityPosition(entity *Entity) (*PositionComponent, bool) {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	return pos, ok
}

// isAtFormationPosition checks if entity is within formation tolerance.
func isAtFormationPosition(pos *PositionComponent, targetX, targetY float64) bool {
	dx := targetX - pos.X
	dy := targetY - pos.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < 5.0
}

// stopEntityMovement sets entity velocity to zero.
func stopEntityMovement(entity *Entity) {
	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			vel.VX = 0
			vel.VY = 0
		}
	}
}

// moveToFormation directs entity movement toward formation position.
func moveToFormation(entity *Entity, pos *PositionComponent, targetX, targetY float64) {
	dx := targetX - pos.X
	dy := targetY - pos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if velComp, ok := entity.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			speed := 80.0
			vel.VX = (dx / distance) * speed
			vel.VY = (dy / distance) * speed
		}
	}
}

// NewFlankTargetAction creates an action that attempts to flank the current target.
func NewFlankTargetAction(world *World) *ActionNode {
	return NewActionNode("FlankTarget", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
		// Get target from blackboard
		target, ok := GetEntityFromBlackboard(blackboard, "target")
		if !ok || target == nil {
			return NodeFailure
		}

		// Validate and extract components
		squad, pos, targetPos, failed := extractFlankingComponents(entity, target)
		if failed {
			return NodeFailure
		}

		// Calculate flanking position
		flankX, flankY := calculateFlankingPosition(squad, pos, targetPos)

		// Move towards flanking position
		return applyFlankingMovement(entity, pos, flankX, flankY)
	})
}

// extractFlankingComponents validates and extracts required components for flanking.
func extractFlankingComponents(entity, target *Entity) (*SquadComponent, *PositionComponent, *PositionComponent, bool) {
	// Get squad component
	squadComp, ok := entity.GetComponent("squad")
	if !ok {
		return nil, nil, nil, true
	}
	squad, ok := squadComp.(*SquadComponent)
	if !ok {
		return nil, nil, nil, true
	}

	// Get entity position
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil, nil, nil, true
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil, nil, true
	}

	// Get target position
	targetPosComp, ok := target.GetComponent("position")
	if !ok {
		return nil, nil, nil, true
	}
	targetPos, ok := targetPosComp.(*PositionComponent)
	if !ok {
		return nil, nil, nil, true
	}

	return squad, pos, targetPos, false
}

// calculateFlankingPosition determines the flanking position around target.
func calculateFlankingPosition(squad *SquadComponent, pos, targetPos *PositionComponent) (float64, float64) {
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

	return flankX, flankY
}

// applyFlankingMovement moves entity towards flanking position.
func applyFlankingMovement(entity *Entity, pos *PositionComponent, flankX, flankY float64) NodeStatus {
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
		priorityTarget, ok := GetEntityFromBlackboard(squad.SharedBlackboard, "priority_target")
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
		squad, pos, err := getSquadAndPosition(entity, blackboard)
		if err != nil {
			return NodeFailure
		}

		allSquadEntities := world.GetEntitiesWith("squad", "position")
		alertedCount := alertNearbySquads(entity, squad, pos, allSquadEntities, 500.0)

		if alertedCount > 0 {
			currentTime, _ := blackboard.GetFloat64("game_time")
			squad.SetAlerted(currentTime)
			return NodeSuccess
		}

		return NodeFailure
	})
}

// getSquadAndPosition retrieves squad component and position from entity with validation.
func getSquadAndPosition(entity *Entity, blackboard *Blackboard) (*SquadComponent, *PositionComponent, error) {
	squadComp, ok := entity.GetComponent("squad")
	if !ok {
		return nil, nil, fmt.Errorf("no squad component")
	}
	squad, ok := squadComp.(*SquadComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid squad component")
	}

	currentTime, _ := blackboard.GetFloat64("game_time")
	if !squad.CanAlert(currentTime) {
		return nil, nil, fmt.Errorf("cannot alert yet")
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return nil, nil, fmt.Errorf("no position component")
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid position component")
	}

	return squad, pos, nil
}

// alertNearbySquads alerts all nearby squads within range and returns count alerted.
func alertNearbySquads(entity *Entity, squad *SquadComponent, pos *PositionComponent, allSquadEntities []*Entity, alertRange float64) int {
	alertedCount := 0

	for _, other := range allSquadEntities {
		if other.ID == entity.ID {
			continue
		}

		otherSquad, otherPos, ok := getOtherSquadAndPosition(other, squad.SquadID)
		if !ok {
			continue
		}

		if isWithinRange(pos, otherPos, alertRange) {
			alertSquad(otherSquad, pos)
			alertedCount++
		}
	}

	return alertedCount
}

// getOtherSquadAndPosition retrieves squad and position from another entity with validation.
func getOtherSquadAndPosition(other *Entity, skipSquadID int) (*SquadComponent, *PositionComponent, bool) {
	otherSquadComp, _ := other.GetComponent("squad")
	if otherSquadComp == nil {
		return nil, nil, false
	}
	otherSquad, ok := otherSquadComp.(*SquadComponent)
	if !ok || otherSquad.SquadID == skipSquadID {
		return nil, nil, false
	}

	otherPosComp, _ := other.GetComponent("position")
	if otherPosComp == nil {
		return nil, nil, false
	}
	otherPos, ok := otherPosComp.(*PositionComponent)
	if !ok {
		return nil, nil, false
	}

	return otherSquad, otherPos, true
}

// alertSquad sets alert status in squad's shared blackboard.
func alertSquad(squad *SquadComponent, alertPos *PositionComponent) {
	if squad.SharedBlackboard != nil {
		squad.SharedBlackboard.Set("alerted", true)
		squad.SharedBlackboard.Set("alert_position_x", alertPos.X)
		squad.SharedBlackboard.Set("alert_position_y", alertPos.Y)
	}
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
		squad := getSquadLeaderComponent(entity)
		if squad == nil {
			return NodeFailure
		}

		evaluateRetreatDecision(entity, squad, blackboard)
		setPriorityTarget(squad, blackboard)

		return NodeSuccess
	})
}

// getSquadLeaderComponent retrieves and validates the squad component for a leader.
func getSquadLeaderComponent(entity *Entity) *SquadComponent {
	squadComp, ok := entity.GetComponent("squad")
	if !ok {
		return nil
	}
	squad, ok := squadComp.(*SquadComponent)
	if !ok || !squad.IsLeader() {
		return nil
	}
	return squad
}

// evaluateRetreatDecision determines whether the squad should retreat based on leader health.
func evaluateRetreatDecision(entity *Entity, squad *SquadComponent, blackboard *Blackboard) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	healthPercent := float64(health.Current) / float64(health.Max)

	if healthPercent < 0.3 {
		squad.SharedBlackboard.Set("retreat_ordered", true)
		setRetreatDirection(entity, squad, blackboard)
	} else {
		squad.SharedBlackboard.Set("retreat_ordered", false)
	}
}

// setRetreatDirection calculates and sets the retreat direction away from the target.
func setRetreatDirection(entity *Entity, squad *SquadComponent, blackboard *Blackboard) {
	target, ok := GetEntityFromBlackboard(blackboard, "target")
	if !ok {
		return
	}

	posComp, _ := entity.GetComponent("position")
	targetPosComp, _ := target.GetComponent("position")
	if posComp == nil || targetPosComp == nil {
		return
	}

	pos, ok1 := posComp.(*PositionComponent)
	targetPos, ok2 := targetPosComp.(*PositionComponent)
	if !ok1 || !ok2 {
		return
	}

	dx := pos.X - targetPos.X
	dy := pos.Y - targetPos.Y
	squad.SharedBlackboard.Set("retreat_dir_x", dx)
	squad.SharedBlackboard.Set("retreat_dir_y", dy)
}

// setPriorityTarget sets the squad's priority target from the leader's current target.
func setPriorityTarget(squad *SquadComponent, blackboard *Blackboard) {
	if target, ok := GetEntityFromBlackboard(blackboard, "target"); ok {
		squad.SharedBlackboard.Set("priority_target", target)
	}
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

		target, ok := GetEntityFromBlackboard(squad.SharedBlackboard, "priority_target")
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
