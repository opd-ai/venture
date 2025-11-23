// Package engine provides squad coordination system for managing squad behaviors.
// This file implements SquadSystem which coordinates squad member actions.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// SquadSystem manages squad coordination and formation maintenance.
type SquadSystem struct {
	world  *World
	logger *logrus.Logger
}

// NewSquadSystem creates a new squad coordination system.
func NewSquadSystem(world *World) *SquadSystem {
	return &SquadSystem{
		world:  world,
		logger: logrus.New(),
	}
}

// Update processes all entities with squad components.
func (s *SquadSystem) Update(deltaTime float64) {
	// Get all entities with squad components
	entities := s.world.GetEntitiesWith("squad", "position")

	// Process each squad
	squads := s.organizeSquads(entities)
	for _, members := range squads {
		s.updateSquad(members, deltaTime)
	}
}

// organizeSquads groups entities by squad ID.
func (s *SquadSystem) organizeSquads(entities []*Entity) map[int][]*Entity {
	squads := make(map[int][]*Entity)
	for _, entity := range entities {
		squadComp, ok := entity.GetComponent("squad")
		if !ok {
			continue
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			continue
		}
		squads[squad.SquadID] = append(squads[squad.SquadID], entity)
	}
	return squads
}

// updateSquad updates a single squad's coordination.
func (s *SquadSystem) updateSquad(members []*Entity, deltaTime float64) {
	if len(members) == 0 {
		return
	}

	// Find the leader
	var leader *Entity
	var leaderSquad *SquadComponent
	for _, member := range members {
		squadComp, _ := member.GetComponent("squad")
		if squadComp == nil {
			continue
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			continue
		}
		if squad.IsLeader() {
			leader = member
			leaderSquad = squad
			break
		}
	}

	if leader == nil || leaderSquad == nil {
		return
	}

	// Update formation if needed
	if leaderSquad.Formation != FormationNone {
		s.updateFormation(leader, leaderSquad, members)
	}

	// Share information across squad via shared blackboard
	s.synchronizeBlackboards(members)
}

// updateFormation calculates and applies formation positions.
func (s *SquadSystem) updateFormation(leader *Entity, leaderSquad *SquadComponent, members []*Entity) {
	leaderPos, leaderAngle, ok := s.getLeaderComponents(leader)
	if !ok {
		return
	}

	memberIndex := 0
	for _, member := range members {
		if member.ID == leader.ID {
			continue // Skip leader
		}

		if s.updateMemberFormationTarget(member, leaderSquad, leaderPos, leaderAngle, memberIndex) {
			memberIndex++
		}
	}
}

// getLeaderComponents extracts position and rotation from leader entity.
func (s *SquadSystem) getLeaderComponents(leader *Entity) (*PositionComponent, float64, bool) {
	leaderPosComp, ok := leader.GetComponent("position")
	if !ok {
		return nil, 0, false
	}
	leaderPos, ok := leaderPosComp.(*PositionComponent)
	if !ok {
		return nil, 0, false
	}

	leaderAngle := 0.0
	if rotComp, ok := leader.GetComponent("rotation"); ok {
		if rot, ok := rotComp.(*RotationComponent); ok {
			leaderAngle = rot.Angle
		}
	}

	return leaderPos, leaderAngle, true
}

// updateMemberFormationTarget updates a single member's formation target.
func (s *SquadSystem) updateMemberFormationTarget(member *Entity, leaderSquad *SquadComponent, leaderPos *PositionComponent, leaderAngle float64, memberIndex int) bool {
	squadComp, _ := member.GetComponent("squad")
	if squadComp == nil {
		return false
	}
	squad, ok := squadComp.(*SquadComponent)
	if !ok {
		return false
	}

	targetX, targetY := s.calculateFormationPosition(
		leaderSquad.Formation,
		memberIndex,
		leaderPos.X,
		leaderPos.Y,
		leaderAngle,
		squad.FormationSpacing,
	)

	s.storeFormationTarget(member, targetX, targetY)
	return true
}

// storeFormationTarget stores target position in member's behavior tree blackboard.
func (s *SquadSystem) storeFormationTarget(member *Entity, targetX, targetY float64) {
	if btComp, ok := member.GetComponent("behaviortree"); ok {
		if bt, ok := btComp.(*BehaviorTreeComponent); ok {
			bt.Blackboard.Set("formation_target_x", targetX)
			bt.Blackboard.Set("formation_target_y", targetY)
		}
	}
}

// calculateFormationPosition returns the target position for a squad member.
func (s *SquadSystem) calculateFormationPosition(
	formation FormationType,
	memberIndex int,
	leaderX, leaderY, leaderAngle, spacing float64,
) (float64, float64) {
	switch formation {
	case FormationLine:
		// Form a horizontal line perpendicular to leader's facing
		perpAngle := leaderAngle + math.Pi/2
		var offset float64
		if memberIndex%2 == 0 {
			offset = float64(memberIndex/2+1) * spacing
		} else {
			offset = -float64((memberIndex+1)/2) * spacing
		}
		return leaderX + math.Cos(perpAngle)*offset, leaderY + math.Sin(perpAngle)*offset

	case FormationWedge:
		// Form a V-shape with leader at the front
		side := 1.0
		if memberIndex%2 == 0 {
			side = -1.0
		}
		row := (memberIndex / 2) + 1
		backOffset := float64(row) * spacing
		sideOffset := float64(row) * spacing * 0.6 * side

		angle := leaderAngle + math.Pi // Behind leader
		perpAngle := leaderAngle + math.Pi/2
		backX := leaderX + math.Cos(angle)*backOffset
		backY := leaderY + math.Sin(angle)*backOffset
		return backX + math.Cos(perpAngle)*sideOffset, backY + math.Sin(perpAngle)*sideOffset

	case FormationCircle:
		// Form a circle around the leader
		count := memberIndex + 1 // +1 to avoid division by zero
		angleStep := 2 * math.Pi / float64(count)
		angle := float64(memberIndex) * angleStep
		radius := spacing
		return leaderX + math.Cos(angle)*radius, leaderY + math.Sin(angle)*radius

	case FormationScatter:
		// Spread out in a semi-random pattern (deterministic based on index)
		// Use member index to create deterministic scatter
		angle := float64(memberIndex) * 2.618 // Golden angle for even distribution
		distance := spacing * (1.0 + float64(memberIndex%3)*0.5)
		return leaderX + math.Cos(angle)*distance, leaderY + math.Sin(angle)*distance

	default:
		return leaderX, leaderY
	}
}

// synchronizeBlackboards ensures all squad members share the same information.
func (s *SquadSystem) synchronizeBlackboards(members []*Entity) {
	// Find the shared blackboard (from any member)
	var sharedBlackboard *Blackboard
	for _, member := range members {
		squadComp, _ := member.GetComponent("squad")
		if squadComp == nil {
			continue
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			continue
		}
		if squad.SharedBlackboard != nil {
			sharedBlackboard = squad.SharedBlackboard
			break
		}
	}

	if sharedBlackboard == nil {
		return
	}

	// Ensure all members reference the same shared blackboard
	for _, member := range members {
		squadComp, _ := member.GetComponent("squad")
		if squadComp == nil {
			continue
		}
		squad, ok := squadComp.(*SquadComponent)
		if !ok {
			continue
		}
		squad.SharedBlackboard = sharedBlackboard
	}
}
