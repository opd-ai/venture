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
		squad := squadComp.(*SquadComponent)
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
		squad := squadComp.(*SquadComponent)
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
	leaderPosComp, ok := leader.GetComponent("position")
	if !ok {
		return
	}
	leaderPos := leaderPosComp.(*PositionComponent)

	// Get leader's rotation for oriented formations
	leaderAngle := 0.0
	if rotComp, ok := leader.GetComponent("rotation"); ok {
		rot := rotComp.(*RotationComponent)
		leaderAngle = rot.Angle
	}

	// Calculate formation positions for each member
	memberIndex := 0
	for _, member := range members {
		if member.ID == leader.ID {
			continue // Skip leader
		}

		squadComp, _ := member.GetComponent("squad")
		squad := squadComp.(*SquadComponent)

		// Calculate target position based on formation type
		targetX, targetY := s.calculateFormationPosition(
			leaderSquad.Formation,
			memberIndex,
			leaderPos.X,
			leaderPos.Y,
			leaderAngle,
			squad.FormationSpacing,
		)

		// Store target position in member's blackboard
		if btComp, ok := member.GetComponent("behaviortree"); ok {
			bt := btComp.(*BehaviorTreeComponent)
			bt.Blackboard.Set("formation_target_x", targetX)
			bt.Blackboard.Set("formation_target_y", targetY)
		}

		memberIndex++
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
		offset := (float64(memberIndex) - 0.5) * spacing
		if memberIndex%2 == 0 {
			offset = float64(memberIndex/2) * spacing
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
		row := (memberIndex + 1) / 2
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
		squad := squadComp.(*SquadComponent)
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
		squad := squadComp.(*SquadComponent)
		squad.SharedBlackboard = sharedBlackboard
	}
}
