// Package engine provides tests for squad components.
package engine

import (
	"testing"
)

func TestSquadRole_String(t *testing.T) {
	tests := []struct {
		name string
		role SquadRole
		want string
	}{
		{"leader", SquadRoleLeader, "Leader"},
		{"member", SquadRoleMember, "Member"},
		{"unknown", SquadRole(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("SquadRole.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormationType_String(t *testing.T) {
	tests := []struct {
		name      string
		formation FormationType
		want      string
	}{
		{"none", FormationNone, "None"},
		{"line", FormationLine, "Line"},
		{"wedge", FormationWedge, "Wedge"},
		{"circle", FormationCircle, "Circle"},
		{"scatter", FormationScatter, "Scatter"},
		{"unknown", FormationType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.formation.String(); got != tt.want {
				t.Errorf("FormationType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSquadComponent_Type(t *testing.T) {
	comp := NewSquadComponent(1, SquadRoleLeader, 0)
	if got := comp.Type(); got != "squad" {
		t.Errorf("SquadComponent.Type() = %v, want squad", got)
	}
}

func TestNewSquadComponent(t *testing.T) {
	tests := []struct {
		name     string
		squadID  int
		role     SquadRole
		leaderID int
	}{
		{"leader", 1, SquadRoleLeader, 0},
		{"member", 1, SquadRoleMember, 42},
		{"different squad", 2, SquadRoleMember, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSquadComponent(tt.squadID, tt.role, tt.leaderID)

			if comp.SquadID != tt.squadID {
				t.Errorf("SquadID = %v, want %v", comp.SquadID, tt.squadID)
			}
			if comp.Role != tt.role {
				t.Errorf("Role = %v, want %v", comp.Role, tt.role)
			}
			if comp.LeaderID != tt.leaderID {
				t.Errorf("LeaderID = %v, want %v", comp.LeaderID, tt.leaderID)
			}
			if comp.Formation != FormationNone {
				t.Errorf("Formation = %v, want FormationNone", comp.Formation)
			}
			if comp.FormationSpacing != 50.0 {
				t.Errorf("FormationSpacing = %v, want 50.0", comp.FormationSpacing)
			}
			if comp.SharedBlackboard == nil {
				t.Error("SharedBlackboard is nil, want non-nil")
			}
			if comp.AlertCooldown != 2.0 {
				t.Errorf("AlertCooldown = %v, want 2.0", comp.AlertCooldown)
			}
		})
	}
}

func TestSquadComponent_IsLeader(t *testing.T) {
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
			comp := NewSquadComponent(1, tt.role, 0)
			if got := comp.IsLeader(); got != tt.want {
				t.Errorf("IsLeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSquadComponent_CanAlert(t *testing.T) {
	tests := []struct {
		name          string
		lastAlertTime float64
		currentTime   float64
		cooldown      float64
		want          bool
	}{
		{"never alerted", 0.0, 5.0, 2.0, true},
		{"within cooldown", 5.0, 6.0, 2.0, false},
		{"after cooldown", 5.0, 7.5, 2.0, true},
		{"exactly at cooldown", 5.0, 7.0, 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSquadComponent(1, SquadRoleLeader, 0)
			comp.LastAlertTime = tt.lastAlertTime
			comp.AlertCooldown = tt.cooldown

			if got := comp.CanAlert(tt.currentTime); got != tt.want {
				t.Errorf("CanAlert(%v) = %v, want %v", tt.currentTime, got, tt.want)
			}
		})
	}
}

func TestSquadComponent_SetAlerted(t *testing.T) {
	comp := NewSquadComponent(1, SquadRoleLeader, 0)
	comp.SetAlerted(10.0)

	if comp.LastAlertTime != 10.0 {
		t.Errorf("LastAlertTime = %v, want 10.0", comp.LastAlertTime)
	}

	// Set again with different time
	comp.SetAlerted(15.0)
	if comp.LastAlertTime != 15.0 {
		t.Errorf("LastAlertTime = %v, want 15.0", comp.LastAlertTime)
	}
}

func TestSquadComponent_FormationDefaults(t *testing.T) {
	comp := NewSquadComponent(1, SquadRoleMember, 10)

	// Test default formation values
	if comp.Formation != FormationNone {
		t.Errorf("Default Formation = %v, want FormationNone", comp.Formation)
	}
	if comp.FormationPosition != 0 {
		t.Errorf("Default FormationPosition = %v, want 0", comp.FormationPosition)
	}
	if comp.FormationSpacing != 50.0 {
		t.Errorf("Default FormationSpacing = %v, want 50.0", comp.FormationSpacing)
	}
}

func TestSquadComponent_SharedBlackboard(t *testing.T) {
	// Create two squad members with different components
	leader := NewSquadComponent(1, SquadRoleLeader, 0)
	member := NewSquadComponent(1, SquadRoleMember, 1)

	// They should have separate blackboards initially
	leader.SharedBlackboard.Set("test", "leader_value")

	// Verify member doesn't have this value yet
	if val, ok := member.SharedBlackboard.Get("test"); ok {
		t.Errorf("Member has leader's value %v, should have separate blackboard", val)
	}

	// Now make them share the same blackboard (simulating squad system)
	sharedBB := NewBlackboard()
	leader.SharedBlackboard = sharedBB
	member.SharedBlackboard = sharedBB

	// Now changes should be visible to both
	leader.SharedBlackboard.Set("shared", "value")
	if val, ok := member.SharedBlackboard.Get("shared"); !ok || val != "value" {
		t.Errorf("Member doesn't see shared value, got %v", val)
	}
}
