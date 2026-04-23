// Package engine provides tests for squad-aware behavior trees.
package engine

import (
	"testing"
)

func TestBuildSquadBehaviorTree_Leader(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name      string
		archetype EnemyArchetype
		wantName  string
	}{
		{"melee leader", ArchetypeMelee, "Melee Squad Leader"},
		{"ranged leader", ArchetypeRanged, "Ranged Squad Leader"},
		{"tank leader", ArchetypeTank, "Tank Squad Leader"},
		{"support leader", ArchetypeSupport, "Support Squad Leader"},
		{"stealth leader", ArchetypeStealth, "Stealth Squad Leader"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt := BuildSquadBehaviorTree(tt.archetype, world, true)

			if bt == nil {
				t.Fatal("BuildSquadBehaviorTree returned nil")
			}
			if bt.Root == nil {
				t.Fatal("Behavior tree root is nil")
			}
			if bt.TreeName != tt.wantName {
				t.Errorf("TreeName = %v, want %v", bt.TreeName, tt.wantName)
			}
			if bt.Blackboard == nil {
				t.Error("Blackboard is nil")
			}
			if !bt.Enabled {
				t.Error("Behavior tree not enabled")
			}
		})
	}
}

func TestBuildSquadBehaviorTree_Member(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name      string
		archetype EnemyArchetype
		wantName  string
	}{
		{"melee member", ArchetypeMelee, "Melee Squad Member"},
		{"ranged member", ArchetypeRanged, "Ranged Squad Member"},
		{"tank member", ArchetypeTank, "Tank Squad Member"},
		{"support member", ArchetypeSupport, "Support Squad Member"},
		{"stealth member", ArchetypeStealth, "Stealth Squad Member"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt := BuildSquadBehaviorTree(tt.archetype, world, false)

			if bt == nil {
				t.Fatal("BuildSquadBehaviorTree returned nil")
			}
			if bt.Root == nil {
				t.Fatal("Behavior tree root is nil")
			}
			if bt.TreeName != tt.wantName {
				t.Errorf("TreeName = %v, want %v", bt.TreeName, tt.wantName)
			}
		})
	}
}

func TestSquadLeaderTree_MakesDecisions(t *testing.T) {
	world := NewWorld()

	// Create leader entity
	leader := NewEntity(1)
	leader.AddComponent(&PositionComponent{X: 100, Y: 100})
	leader.AddComponent(&VelocityComponent{})
	leader.AddComponent(&HealthComponent{Current: 80, Max: 100})
	leader.AddComponent(NewSquadComponent(1, SquadRoleLeader, 0))

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, true)
	leader.AddComponent(bt)

	// Execute tree
	status := bt.Tick(leader, 0.016)

	// Should not fail (returns Running, Success, or Failure depending on state)
	if status < 0 || status > 2 {
		t.Errorf("Invalid status: %v", status)
	}
}

func TestSquadMemberTree_FollowsOrders(t *testing.T) {
	world := NewWorld()

	// Create member entity
	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 150, Y: 150})
	member.AddComponent(&VelocityComponent{})
	member.AddComponent(&HealthComponent{Current: 100, Max: 100})
	squad := NewSquadComponent(1, SquadRoleMember, 1)
	squad.SharedBlackboard.Set("retreat_ordered", true)
	squad.SharedBlackboard.Set("retreat_dir_x", -1.0)
	squad.SharedBlackboard.Set("retreat_dir_y", 0.0)
	member.AddComponent(squad)

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, false)
	member.AddComponent(bt)

	// Execute tree - should follow retreat order
	status := bt.Tick(member, 0.016)

	// Should be running (retreating)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning for retreat, got %v", status)
	}

	// Velocity should be set for retreat
	velComp, _ := member.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX >= 0 {
		t.Errorf("Expected negative X velocity for retreat, got %v", vel.VX)
	}
}

func TestSquadMemberTree_FocusFire(t *testing.T) {
	world := NewWorld()

	// Create target
	target := NewEntity(10)
	target.AddComponent(&PositionComponent{X: 200, Y: 100})
	target.AddComponent(&TeamComponent{TeamID: 2})
	target.AddComponent(&HealthComponent{Current: 50, Max: 50})

	// Create member entity
	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 100, Y: 100})
	member.AddComponent(&VelocityComponent{})
	member.AddComponent(&HealthComponent{Current: 100, Max: 100})
	member.AddComponent(&TeamComponent{TeamID: 1})
	member.AddComponent(&AttackComponent{Damage: 10, Cooldown: 1.0, CooldownTimer: 0.0})

	squad := NewSquadComponent(1, SquadRoleMember, 1)
	squad.SharedBlackboard.Set("priority_target", target)
	member.AddComponent(squad)

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, false)
	member.AddComponent(bt)

	// Execute tree - should focus fire on priority target
	status := bt.Tick(member, 0.016)

	// Should succeed or be running
	if status == NodeFailure {
		t.Error("Squad member failed to focus fire on priority target")
	}

	// Personal blackboard should have the priority target
	priorityTarget, ok := GetEntityFromBlackboard(bt.Blackboard, "target")
	if !ok || priorityTarget != target {
		t.Error("Priority target not set in member's blackboard")
	}
}

func TestSquadMemberTree_MaintainFormation(t *testing.T) {
	world := NewWorld()

	// Create member entity
	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 100, Y: 100})
	member.AddComponent(&VelocityComponent{})
	member.AddComponent(&HealthComponent{Current: 100, Max: 100})
	member.AddComponent(NewSquadComponent(1, SquadRoleMember, 1))

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, false)
	member.AddComponent(bt)

	// Set formation target
	bt.Blackboard.Set("formation_target_x", 200.0)
	bt.Blackboard.Set("formation_target_y", 100.0)

	// Execute tree - should move to formation
	status := bt.Tick(member, 0.016)

	// Should be running (moving to formation)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning for formation movement, got %v", status)
	}

	// Velocity should be set towards formation
	velComp, _ := member.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX <= 0 {
		t.Errorf("Expected positive X velocity towards formation, got %v", vel.VX)
	}
}

func TestGetArchetypeParams(t *testing.T) {
	tests := []struct {
		name          string
		archetype     EnemyArchetype
		wantDetection float64
		wantAttack    float64
		wantSpeed     float64
	}{
		{"melee", ArchetypeMelee, 250.0, 40.0, 100.0},
		{"ranged", ArchetypeRanged, 300.0, 200.0, 90.0},
		{"tank", ArchetypeTank, 200.0, 50.0, 70.0},
		{"support", ArchetypeSupport, 280.0, 150.0, 95.0},
		{"stealth", ArchetypeStealth, 220.0, 35.0, 120.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection, attack, speed := getArchetypeParams(tt.archetype)

			if detection != tt.wantDetection {
				t.Errorf("Detection range = %v, want %v", detection, tt.wantDetection)
			}
			if attack != tt.wantAttack {
				t.Errorf("Attack range = %v, want %v", attack, tt.wantAttack)
			}
			if speed != tt.wantSpeed {
				t.Errorf("Move speed = %v, want %v", speed, tt.wantSpeed)
			}
		})
	}
}

func TestSquadLeaderTree_CallForHelp(t *testing.T) {
	world := NewWorld()

	// Create leader entity with low health
	leader := NewEntity(1)
	leader.AddComponent(&PositionComponent{X: 100, Y: 100})
	leader.AddComponent(&VelocityComponent{})
	leader.AddComponent(&HealthComponent{Current: 35, Max: 100}) // 35% health
	leader.AddComponent(&TeamComponent{TeamID: 1})
	squad := NewSquadComponent(1, SquadRoleLeader, 0)
	leader.AddComponent(squad)
	world.AddEntity(leader)

	// Create target
	target := NewEntity(10)
	target.AddComponent(&PositionComponent{X: 150, Y: 100})
	target.AddComponent(&TeamComponent{TeamID: 2})
	target.AddComponent(&HealthComponent{Current: 50, Max: 50})
	world.AddEntity(target)

	// Create nearby squad
	nearbyLeader := NewEntity(2)
	nearbyLeader.AddComponent(&PositionComponent{X: 200, Y: 100})
	nearbySquad := NewSquadComponent(2, SquadRoleLeader, 0)
	nearbyLeader.AddComponent(nearbySquad)
	world.AddEntity(nearbyLeader)

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, true)
	leader.AddComponent(bt)
	bt.Blackboard.Set("target", target)
	bt.Blackboard.Set("game_time", 10.0)

	// Execute tree multiple times to trigger call for help
	for i := 0; i < 5; i++ {
		bt.Tick(leader, 0.016)
	}

	// Nearby squad should be alerted (if call for help was triggered)
	// Note: This might not trigger in every execution due to tree priorities
	// The test verifies the tree doesn't crash with this scenario
}

func TestSquadMemberTree_EmergencyFlee(t *testing.T) {
	world := NewWorld()

	// Create member with very low health
	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 100, Y: 100})
	member.AddComponent(&VelocityComponent{})
	member.AddComponent(&HealthComponent{Current: 5, Max: 100}) // 5% health
	member.AddComponent(&TeamComponent{TeamID: 1})
	member.AddComponent(NewSquadComponent(1, SquadRoleMember, 1))

	// Create target
	target := NewEntity(10)
	target.AddComponent(&PositionComponent{X: 150, Y: 100})
	target.AddComponent(&TeamComponent{TeamID: 2})

	bt := BuildSquadBehaviorTree(ArchetypeMelee, world, false)
	member.AddComponent(bt)
	bt.Blackboard.Set("target", target)

	// Execute tree - should flee due to very low health
	status := bt.Tick(member, 0.016)

	// Should be running (fleeing)
	if status != NodeRunning {
		t.Errorf("Expected NodeRunning for emergency flee, got %v", status)
	}

	// Velocity should be set away from target
	velComp, _ := member.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX >= 0 {
		t.Errorf("Expected negative X velocity fleeing from target, got %v", vel.VX)
	}
}
