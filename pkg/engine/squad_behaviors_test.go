// Package engine provides tests for squad behavior tree actions.
package engine

import (
	"testing"
)

func TestMaintainFormationAction_NoTarget(t *testing.T) {
	action := NewMaintainFormationAction()
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{})
	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure when no formation target, got %v", status)
	}
}

func TestMaintainFormationAction_AtTarget(t *testing.T) {
	action := NewMaintainFormationAction()
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	blackboard := NewBlackboard()
	blackboard.Set("formation_target_x", 101.0) // Very close
	blackboard.Set("formation_target_y", 101.0)

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected NodeSuccess when at target, got %v", status)
	}

	// Velocity should be stopped
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX != 0 || vel.VY != 0 {
		t.Errorf("Expected zero velocity, got (%v, %v)", vel.VX, vel.VY)
	}
}

func TestMaintainFormationAction_Moving(t *testing.T) {
	action := NewMaintainFormationAction()
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{})

	blackboard := NewBlackboard()
	blackboard.Set("formation_target_x", 200.0)
	blackboard.Set("formation_target_y", 100.0)

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning when moving, got %v", status)
	}

	// Velocity should be set towards target
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX <= 0 {
		t.Errorf("Expected positive X velocity towards target, got %v", vel.VX)
	}
	if vel.VY != 0 {
		t.Errorf("Expected zero Y velocity for horizontal movement, got %v", vel.VY)
	}
}

func TestFlankTargetAction_NoTarget(t *testing.T) {
	world := NewWorld()
	action := NewFlankTargetAction(world)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(NewSquadComponent(1, SquadRoleMember, 10))

	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure when no target, got %v", status)
	}
}

func TestFlankTargetAction_WithTarget(t *testing.T) {
	world := NewWorld()
	action := NewFlankTargetAction(world)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{})
	squadComp := NewSquadComponent(1, SquadRoleMember, 10)
	squadComp.FormationPosition = 0 // Even - flank left
	entity.AddComponent(squadComp)

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 200, Y: 100})

	blackboard := NewBlackboard()
	blackboard.Set("target", target)

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning when flanking, got %v", status)
	}

	// Velocity should be set
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("Expected non-zero velocity when flanking")
	}
}

func TestFocusFireAction_NoSquad(t *testing.T) {
	action := NewFocusFireAction()
	entity := NewEntity(1)
	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure when not in squad, got %v", status)
	}
}

func TestFocusFireAction_NoPriorityTarget(t *testing.T) {
	action := NewFocusFireAction()
	entity := NewEntity(1)
	entity.AddComponent(NewSquadComponent(1, SquadRoleMember, 10))
	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure when no priority target, got %v", status)
	}
}

func TestFocusFireAction_WithPriorityTarget(t *testing.T) {
	action := NewFocusFireAction()
	entity := NewEntity(1)
	squad := NewSquadComponent(1, SquadRoleMember, 10)
	entity.AddComponent(squad)

	target := NewEntity(2)
	squad.SharedBlackboard.Set("priority_target", target)

	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected NodeSuccess with priority target, got %v", status)
	}

	// Check target was set in personal blackboard
	savedTarget, ok := GetEntityFromBlackboard(blackboard, "target")
	if !ok || savedTarget != target {
		t.Error("Priority target not set in blackboard")
	}
}

func TestCallForHelpAction_WithinCooldown(t *testing.T) {
	world := NewWorld()
	action := NewCallForHelpAction(world)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	squad := NewSquadComponent(1, SquadRoleLeader, 0)
	squad.LastAlertTime = 5.0
	squad.AlertCooldown = 2.0
	entity.AddComponent(squad)

	blackboard := NewBlackboard()
	blackboard.Set("game_time", 6.0) // Within cooldown

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure within cooldown, got %v", status)
	}
}

func TestCallForHelpAction_AlertsNearbySquads(t *testing.T) {
	world := NewWorld()
	action := NewCallForHelpAction(world)

	// Calling entity
	caller := NewEntity(1)
	caller.AddComponent(&PositionComponent{X: 100, Y: 100})
	callerSquad := NewSquadComponent(1, SquadRoleLeader, 0)
	caller.AddComponent(callerSquad)
	world.AddEntity(caller)

	// Nearby entity in different squad
	nearby := NewEntity(2)
	nearby.AddComponent(&PositionComponent{X: 200, Y: 100}) // 100 pixels away
	nearbySquad := NewSquadComponent(2, SquadRoleLeader, 0)
	nearby.AddComponent(nearbySquad)
	world.AddEntity(nearby)

	// Far entity (should not be alerted)
	far := NewEntity(3)
	far.AddComponent(&PositionComponent{X: 700, Y: 100}) // 600 pixels away
	farSquad := NewSquadComponent(3, SquadRoleLeader, 0)
	far.AddComponent(farSquad)
	world.AddEntity(far)

	blackboard := NewBlackboard()
	blackboard.Set("game_time", 10.0)

	// Process pending entity additions
	world.Update(0)

	status := action.Tick(caller, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected NodeSuccess after alerting, got %v", status)
	}

	// Nearby squad should be alerted
	alerted, _ := nearbySquad.SharedBlackboard.GetBool("alerted")
	if !alerted {
		t.Error("Nearby squad was not alerted")
	}

	// Far squad should not be alerted
	farAlerted, _ := farSquad.SharedBlackboard.GetBool("alerted")
	if farAlerted {
		t.Error("Far squad should not be alerted")
	}

	// Caller's last alert time should be updated
	if callerSquad.LastAlertTime != 10.0 {
		t.Errorf("LastAlertTime = %v, want 10.0", callerSquad.LastAlertTime)
	}
}

func TestCoordinatedRetreatAction_NoRetreatOrder(t *testing.T) {
	action := NewCoordinatedRetreatAction()
	entity := NewEntity(1)
	entity.AddComponent(NewSquadComponent(1, SquadRoleMember, 10))
	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure when no retreat order, got %v", status)
	}
}

func TestCoordinatedRetreatAction_WithRetreatOrder(t *testing.T) {
	action := NewCoordinatedRetreatAction()
	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{})
	squad := NewSquadComponent(1, SquadRoleMember, 10)
	squad.SharedBlackboard.Set("retreat_ordered", true)
	squad.SharedBlackboard.Set("retreat_dir_x", -1.0)
	squad.SharedBlackboard.Set("retreat_dir_y", 0.0)
	entity.AddComponent(squad)

	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning when retreating, got %v", status)
	}

	// Velocity should be set in retreat direction
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX >= 0 {
		t.Errorf("Expected negative X velocity for retreat, got %v", vel.VX)
	}
}

func TestSquadLeaderDecisionAction_NotLeader(t *testing.T) {
	action := NewSquadLeaderDecisionAction()
	entity := NewEntity(1)
	entity.AddComponent(NewSquadComponent(1, SquadRoleMember, 10))
	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Expected NodeFailure for non-leader, got %v", status)
	}
}

func TestSquadLeaderDecisionAction_OrderRetreat(t *testing.T) {
	action := NewSquadLeaderDecisionAction()
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 20, Max: 100}) // 20% health
	squad := NewSquadComponent(1, SquadRoleLeader, 0)
	entity.AddComponent(squad)

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 200, Y: 100})

	blackboard := NewBlackboard()
	blackboard.Set("target", target)

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected NodeSuccess, got %v", status)
	}

	// Should order retreat when health low
	retreatOrdered, _ := squad.SharedBlackboard.GetBool("retreat_ordered")
	if !retreatOrdered {
		t.Error("Leader should order retreat at low health")
	}

	// Retreat direction should be set
	_, okX := squad.SharedBlackboard.GetFloat64("retreat_dir_x")
	_, okY := squad.SharedBlackboard.GetFloat64("retreat_dir_y")
	if !okX || !okY {
		t.Error("Retreat direction not set")
	}
}

func TestSquadLeaderDecisionAction_ContinueFighting(t *testing.T) {
	action := NewSquadLeaderDecisionAction()
	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 80, Max: 100}) // 80% health
	squad := NewSquadComponent(1, SquadRoleLeader, 0)
	entity.AddComponent(squad)

	blackboard := NewBlackboard()

	status := action.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Expected NodeSuccess, got %v", status)
	}

	// Should not order retreat when health good
	retreatOrdered, _ := squad.SharedBlackboard.GetBool("retreat_ordered")
	if retreatOrdered {
		t.Error("Leader should not order retreat at good health")
	}
}

func TestIsSquadLeaderCondition(t *testing.T) {
	tests := []struct {
		name string
		role SquadRole
		want bool
	}{
		{"leader", SquadRoleLeader, true},
		{"member", SquadRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := NewIsSquadLeaderCondition()
			entity := NewEntity(1)
			entity.AddComponent(NewSquadComponent(1, tt.role, 0))
			blackboard := NewBlackboard()

			result := condition.condition(entity, blackboard)
			if result != tt.want {
				t.Errorf("IsSquadLeaderCondition = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestSquadHasPriorityTargetCondition(t *testing.T) {
	condition := NewSquadHasPriorityTargetCondition()

	entity := NewEntity(1)
	squad := NewSquadComponent(1, SquadRoleMember, 10)
	entity.AddComponent(squad)
	blackboard := NewBlackboard()

	// No priority target
	if condition.condition(entity, blackboard) {
		t.Error("Expected false when no priority target")
	}

	// With priority target
	target := NewEntity(2)
	squad.SharedBlackboard.Set("priority_target", target)
	if !condition.condition(entity, blackboard) {
		t.Error("Expected true when priority target set")
	}
}

func TestRetreatOrderedCondition(t *testing.T) {
	condition := NewRetreatOrderedCondition()

	entity := NewEntity(1)
	squad := NewSquadComponent(1, SquadRoleMember, 10)
	entity.AddComponent(squad)
	blackboard := NewBlackboard()

	// No retreat order
	if condition.condition(entity, blackboard) {
		t.Error("Expected false when no retreat order")
	}

	// With retreat order
	squad.SharedBlackboard.Set("retreat_ordered", true)
	if !condition.condition(entity, blackboard) {
		t.Error("Expected true when retreat ordered")
	}
}

func TestHasFormationTargetCondition(t *testing.T) {
	condition := NewHasFormationTargetCondition()
	entity := NewEntity(1)
	blackboard := NewBlackboard()

	// No formation target
	if condition.condition(entity, blackboard) {
		t.Error("Expected false when no formation target")
	}

	// With formation target
	blackboard.Set("formation_target_x", 100.0)
	blackboard.Set("formation_target_y", 200.0)
	if !condition.condition(entity, blackboard) {
		t.Error("Expected true when formation target set")
	}
}
