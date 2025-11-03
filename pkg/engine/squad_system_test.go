// Package engine provides tests for squad system.
package engine

import (
	"math"
	"testing"
)

func TestNewSquadSystem(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	if system == nil {
		t.Fatal("NewSquadSystem returned nil")
	}
	if system.world != world {
		t.Error("SquadSystem world not set correctly")
	}
	if system.logger == nil {
		t.Error("SquadSystem logger is nil")
	}
}

func TestSquadSystem_OrganizeSquads(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	// Create entities with different squad IDs
	e1 := NewEntity(1)
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	e1.AddComponent(NewSquadComponent(1, SquadRoleLeader, 0))

	e2 := NewEntity(2)
	e2.AddComponent(&PositionComponent{X: 10, Y: 10})
	e2.AddComponent(NewSquadComponent(1, SquadRoleMember, 1))

	e3 := NewEntity(3)
	e3.AddComponent(&PositionComponent{X: 20, Y: 20})
	e3.AddComponent(NewSquadComponent(2, SquadRoleLeader, 0))

	entities := []*Entity{e1, e2, e3}
	squads := system.organizeSquads(entities)

	// Should have 2 squads
	if len(squads) != 2 {
		t.Errorf("Got %d squads, want 2", len(squads))
	}

	// Squad 1 should have 2 members
	if len(squads[1]) != 2 {
		t.Errorf("Squad 1 has %d members, want 2", len(squads[1]))
	}

	// Squad 2 should have 1 member
	if len(squads[2]) != 1 {
		t.Errorf("Squad 2 has %d members, want 1", len(squads[2]))
	}
}

func TestSquadSystem_CalculateFormationPosition_Line(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	leaderX := 100.0
	leaderY := 100.0
	leaderAngle := 0.0 // Facing right
	spacing := 50.0

	tests := []struct {
		name        string
		memberIndex int
		wantApproxX float64
		wantApproxY float64
	}{
		{"member 0", 0, 100.0, 150.0}, // First member goes to one side
		{"member 1", 1, 100.0, 50.0},  // Second member goes to other side
		{"member 2", 2, 100.0, 200.0}, // Third continues pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := system.calculateFormationPosition(
				FormationLine,
				tt.memberIndex,
				leaderX,
				leaderY,
				leaderAngle,
				spacing,
			)

			// Allow some tolerance for floating point
			tolerance := 5.0
			if math.Abs(x-tt.wantApproxX) > tolerance {
				t.Errorf("X position = %v, want approximately %v", x, tt.wantApproxX)
			}
			if math.Abs(y-tt.wantApproxY) > tolerance {
				t.Errorf("Y position = %v, want approximately %v", y, tt.wantApproxY)
			}
		})
	}
}

func TestSquadSystem_CalculateFormationPosition_Wedge(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	leaderX := 100.0
	leaderY := 100.0
	leaderAngle := 0.0 // Facing right
	spacing := 50.0

	// Wedge formation should place members behind and to the sides
	x0, y0 := system.calculateFormationPosition(FormationWedge, 0, leaderX, leaderY, leaderAngle, spacing)
	x1, y1 := system.calculateFormationPosition(FormationWedge, 1, leaderX, leaderY, leaderAngle, spacing)

	// Both should be behind the leader (lower X since facing right and behind is -X direction)
	if x0 >= leaderX {
		t.Errorf("Member 0 X position %v should be behind leader X %v", x0, leaderX)
	}
	if x1 >= leaderX {
		t.Errorf("Member 1 X position %v should be behind leader X %v", x1, leaderX)
	}

	// They should be on opposite sides
	if (y0 > leaderY && y1 > leaderY) || (y0 < leaderY && y1 < leaderY) {
		t.Errorf("Members should be on opposite sides, got Y0=%v Y1=%v LeaderY=%v", y0, y1, leaderY)
	}
}

func TestSquadSystem_CalculateFormationPosition_Circle(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	leaderX := 100.0
	leaderY := 100.0
	leaderAngle := 0.0
	spacing := 50.0 // This is the radius

	// Get position for first member
	x, y := system.calculateFormationPosition(FormationCircle, 0, leaderX, leaderY, leaderAngle, spacing)

	// Check distance from leader is approximately the spacing (radius)
	dx := x - leaderX
	dy := y - leaderY
	distance := math.Sqrt(dx*dx + dy*dy)

	tolerance := 1.0
	if math.Abs(distance-spacing) > tolerance {
		t.Errorf("Distance from leader = %v, want %v", distance, spacing)
	}
}

func TestSquadSystem_CalculateFormationPosition_Scatter(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	leaderX := 100.0
	leaderY := 100.0
	leaderAngle := 0.0
	spacing := 50.0

	// Get several positions and verify they're different and deterministic
	positions := make([][2]float64, 3)
	for i := 0; i < 3; i++ {
		x, y := system.calculateFormationPosition(FormationScatter, i, leaderX, leaderY, leaderAngle, spacing)
		positions[i] = [2]float64{x, y}
	}

	// Verify they're all different
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[i][0] == positions[j][0] && positions[i][1] == positions[j][1] {
				t.Errorf("Positions %d and %d are identical: %v", i, j, positions[i])
			}
		}
	}

	// Verify determinism - calling again should give same results
	for i := 0; i < 3; i++ {
		x, y := system.calculateFormationPosition(FormationScatter, i, leaderX, leaderY, leaderAngle, spacing)
		if x != positions[i][0] || y != positions[i][1] {
			t.Errorf("Position %d not deterministic: got (%v,%v), want (%v,%v)",
				i, x, y, positions[i][0], positions[i][1])
		}
	}
}

func TestSquadSystem_UpdateFormation(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	// Create leader with line formation
	leader := NewEntity(1)
	leader.AddComponent(&PositionComponent{X: 100, Y: 100})
	leader.AddComponent(&RotationComponent{Angle: 0.0})
	leaderSquad := NewSquadComponent(1, SquadRoleLeader, 0)
	leaderSquad.Formation = FormationLine
	leaderSquad.FormationSpacing = 50.0
	leader.AddComponent(leaderSquad)

	// Create member
	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 50, Y: 50})
	memberBT := NewBehaviorTreeComponent(nil, "test")
	member.AddComponent(memberBT)
	memberSquad := NewSquadComponent(1, SquadRoleMember, 1)
	member.AddComponent(memberSquad)

	members := []*Entity{leader, member}

	// Update formation
	system.updateFormation(leader, leaderSquad, members)

	// Member's blackboard should have formation targets
	targetX, okX := memberBT.Blackboard.GetFloat64("formation_target_x")
	targetY, okY := memberBT.Blackboard.GetFloat64("formation_target_y")

	if !okX || !okY {
		t.Fatal("Formation target not set in member blackboard")
	}

	// Targets should be different from current position
	if targetX == 50.0 && targetY == 50.0 {
		t.Error("Formation target same as current position")
	}
}

func TestSquadSystem_SynchronizeBlackboards(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	// Create squad with shared blackboard
	sharedBB := NewBlackboard()
	sharedBB.Set("test_key", "test_value")

	e1 := NewEntity(1)
	squad1 := NewSquadComponent(1, SquadRoleLeader, 0)
	squad1.SharedBlackboard = sharedBB
	e1.AddComponent(squad1)

	e2 := NewEntity(2)
	squad2 := NewSquadComponent(1, SquadRoleMember, 1)
	squad2.SharedBlackboard = NewBlackboard() // Different initially
	e2.AddComponent(squad2)

	members := []*Entity{e1, e2}

	// Synchronize
	system.synchronizeBlackboards(members)

	// Both should now have the same shared blackboard
	if squad1.SharedBlackboard != squad2.SharedBlackboard {
		t.Error("Blackboards not synchronized")
	}

	// Both should see the test value
	val, ok := squad2.SharedBlackboard.Get("test_key")
	if !ok || val != "test_value" {
		t.Errorf("Member can't see shared value, got %v", val)
	}
}

func TestSquadSystem_Update_EmptyWorld(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	// Update with no entities should not panic
	system.Update(0.016)
}

func TestSquadSystem_Update_WithSquads(t *testing.T) {
	world := NewWorld()
	system := NewSquadSystem(world)

	// Create a squad
	leader := NewEntity(1)
	leader.AddComponent(&PositionComponent{X: 100, Y: 100})
	leaderSquad := NewSquadComponent(1, SquadRoleLeader, 0)
	leaderSquad.Formation = FormationLine
	leader.AddComponent(leaderSquad)
	world.AddEntity(leader)

	member := NewEntity(2)
	member.AddComponent(&PositionComponent{X: 80, Y: 80})
	member.AddComponent(NewBehaviorTreeComponent(nil, "test"))
	memberSquad := NewSquadComponent(1, SquadRoleMember, 1)
	memberSquad.SharedBlackboard = leaderSquad.SharedBlackboard
	member.AddComponent(memberSquad)
	world.AddEntity(member)

	// Process pending entity additions
	world.Update(0)

	// Update should not panic
	system.Update(0.016)

	// Member should have formation targets set
	btComp, _ := member.GetComponent("behaviortree")
	bt := btComp.(*BehaviorTreeComponent)
	_, okX := bt.Blackboard.GetFloat64("formation_target_x")
	_, okY := bt.Blackboard.GetFloat64("formation_target_y")

	if !okX || !okY {
		t.Error("Formation targets not set after Update")
	}
}
