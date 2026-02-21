// Package engine provides squad coordination components for AI.
// This file defines SquadComponent for group tactics and coordination.
package engine

// SquadRole defines the role of an entity within a squad.
type SquadRole int

const (
	// SquadRoleLeader is the squad leader who makes tactical decisions.
	SquadRoleLeader SquadRole = iota
	// SquadRoleMember is a regular squad member who follows leader.
	SquadRoleMember
)

// String returns the string representation of a squad role.
func (r SquadRole) String() string {
	switch r {
	case SquadRoleLeader:
		return "Leader"
	case SquadRoleMember:
		return "Member"
	default:
		return "Unknown"
	}
}

// FormationType defines the type of formation the squad maintains.
type FormationType int

const (
	// FormationNone means no specific formation.
	FormationNone FormationType = iota
	// FormationLine means squad members form a horizontal line.
	FormationLine
	// FormationColumn means squad members form a vertical column behind leader.
	FormationColumn
	// FormationWedge means squad members form a V-shape with leader at front.
	FormationWedge
	// FormationCircle means squad members form a circle around center.
	FormationCircle
	// FormationScatter means squad members spread out randomly.
	FormationScatter
)

// String returns the string representation of a formation type.
func (f FormationType) String() string {
	switch f {
	case FormationNone:
		return "None"
	case FormationLine:
		return "Line"
	case FormationColumn:
		return "Column"
	case FormationWedge:
		return "Wedge"
	case FormationCircle:
		return "Circle"
	case FormationScatter:
		return "Scatter"
	default:
		return "Unknown"
	}
}

// SquadComponent represents squad membership and coordination data.
type SquadComponent struct {
	// SquadID uniquely identifies the squad.
	SquadID int

	// Role defines whether this entity is the leader or a member.
	Role SquadRole

	// LeaderID is the entity ID of the squad leader (0 if this is the leader).
	LeaderID int

	// Formation defines the formation type the squad should maintain.
	Formation FormationType

	// FormationPosition is the relative position within the formation.
	// For example, in a line formation, this might be the index (0, 1, 2...).
	FormationPosition int

	// FormationSpacing is the distance between squad members in pixels.
	FormationSpacing float64

	// SharedBlackboard is a pointer to the squad's shared blackboard for coordination.
	// All members of the squad share this blackboard to communicate.
	SharedBlackboard *Blackboard

	// LastAlertTime tracks when the squad was last alerted (for cooldown).
	LastAlertTime float64

	// AlertCooldown is the minimum time in seconds between alerts.
	AlertCooldown float64
}

// Type returns the component type identifier.
func (s SquadComponent) Type() string {
	return "squad"
}

// NewSquadComponent creates a new squad component.
func NewSquadComponent(squadID int, role SquadRole, leaderID int) *SquadComponent {
	return &SquadComponent{
		SquadID:           squadID,
		Role:              role,
		LeaderID:          leaderID,
		Formation:         FormationNone,
		FormationPosition: 0,
		FormationSpacing:  50.0, // Default spacing in pixels
		SharedBlackboard:  NewBlackboard(),
		LastAlertTime:     0.0,
		AlertCooldown:     2.0, // 2 second cooldown between alerts
	}
}

// IsLeader returns true if this entity is the squad leader.
func (s *SquadComponent) IsLeader() bool {
	return s.Role == SquadRoleLeader
}

// CanAlert returns true if enough time has passed since the last alert.
func (s *SquadComponent) CanAlert(currentTime float64) bool {
	return currentTime-s.LastAlertTime >= s.AlertCooldown
}

// SetAlerted marks the squad as alerted at the given time.
func (s *SquadComponent) SetAlerted(currentTime float64) {
	s.LastAlertTime = currentTime
}
