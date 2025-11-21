// Package engine provides squad-aware behavior tree builders.
// This file implements behavior trees for squad members with tactical coordination.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// BuildSquadBehaviorTree creates a behavior tree for squad members with coordination.
// This extends the base archetype with squad-specific behaviors.
func BuildSquadBehaviorTree(archetype EnemyArchetype, world *World, isLeader bool) *BehaviorTreeComponent {
	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"is_leader": isLeader,
	}).Debug("BuildSquadBehaviorTree entry")

	var root BehaviorNode
	var name string

	if isLeader {
		log.WithFields(log.Fields{
			"archetype": archetype.String(),
			"role":      "leader",
		}).Debug("Building squad leader behavior tree")
		root = buildSquadLeaderTree(archetype, world)
		name = archetype.String() + " Squad Leader"
	} else {
		log.WithFields(log.Fields{
			"archetype": archetype.String(),
			"role":      "member",
		}).Debug("Building squad member behavior tree")
		root = buildSquadMemberTree(archetype, world)
		name = archetype.String() + " Squad Member"
	}

	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"is_leader": isLeader,
		"tree_name": name,
		"has_root":  root != nil,
	}).Debug("BuildSquadBehaviorTree exit")

	return NewBehaviorTreeComponent(root, name)
}

// buildSquadLeaderTree creates a behavior tree for squad leaders.
// Leaders make tactical decisions for the squad and coordinate member actions.
func buildSquadLeaderTree(archetype EnemyArchetype, world *World) BehaviorNode {
	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"role":      "leader",
	}).Debug("buildSquadLeaderTree entry")

	// Get base parameters from archetype
	detectionRange, attackRange, moveSpeed := getArchetypeParams(archetype)

	log.WithFields(log.Fields{
		"archetype":       archetype.String(),
		"detection_range": detectionRange,
		"attack_range":    attackRange,
		"move_speed":      moveSpeed,
	}).Debug("Retrieved archetype parameters for squad leader")

	root := NewSelectorNode("SquadLeaderRoot",
		// Leader makes tactical decisions first
		NewSequenceNode("LeaderDecisions",
			NewIsSquadLeaderCondition(),
			NewSquadLeaderDecisionAction(),
		),

		// Check if squad should retreat
		NewSequenceNode("SquadRetreatSequence",
			NewRetreatOrderedCondition(),
			NewCoordinatedRetreatAction(),
		),

		// Check if should flee individually (emergency)
		NewSequenceNode("EmergencyFleeSequence",
			NewHealthBelowCondition(0.1), // Below 10% health
			NewFleeFromTargetAction(moveSpeed*1.2),
		),

		// Combat sequence with squad coordination
		NewSequenceNode("SquadCombatSequence",
			NewHasTargetCondition(),
			NewSelectorNode("SquadCombatActions",
				// Call for help if needed (health below 40%)
				NewSequenceNode("CallForHelpSequence",
					NewHealthBelowCondition(0.4),
					NewCallForHelpAction(world),
				),

				// Attack if in range
				NewSequenceNode("AttackSequence",
					NewTargetInRangeCondition(attackRange),
					NewAttackTargetAction(),
				),

				// Move towards target
				NewMoveToTargetAction(moveSpeed),
			),
		),

		// Search for target
		NewSequenceNode("SearchSequence",
			NewFindTargetAction(detectionRange, world),
		),

		// Idle behavior
		NewWanderAction(moveSpeed*0.5), // Wander at half speed
	)

	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"role":      "leader",
		"has_root":  root != nil,
	}).Debug("buildSquadLeaderTree exit")

	return root
}

// buildSquadMemberTree creates a behavior tree for squad members.
// Members follow squad tactics and maintain formation.
func buildSquadMemberTree(archetype EnemyArchetype, world *World) BehaviorNode {
	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"role":      "member",
	}).Debug("buildSquadMemberTree entry")

	// Get base parameters from archetype
	detectionRange, attackRange, moveSpeed := getArchetypeParams(archetype)

	log.WithFields(log.Fields{
		"archetype":       archetype.String(),
		"detection_range": detectionRange,
		"attack_range":    attackRange,
		"move_speed":      moveSpeed,
	}).Debug("Retrieved archetype parameters for squad member")

	root := NewSelectorNode("SquadMemberRoot",
		// Follow retreat orders from leader
		NewSequenceNode("SquadRetreatSequence",
			NewRetreatOrderedCondition(),
			NewCoordinatedRetreatAction(),
		),

		// Emergency individual flee (very low health)
		NewSequenceNode("EmergencyFleeSequence",
			NewHealthBelowCondition(0.1),
			NewFleeFromTargetAction(moveSpeed*1.2),
		),

		// Squad combat with priority target
		NewSequenceNode("SquadCombatSequence",
			NewSquadHasPriorityTargetCondition(),
			NewSequenceNode("FocusFireSequence",
				NewFocusFireAction(), // Get priority target from squad
				NewSelectorNode("SquadTactics",
					// Attack if in range
					NewSequenceNode("AttackSequence",
						NewTargetInRangeCondition(attackRange),
						NewAttackTargetAction(),
					),

					// Flank target if archetype is Melee or Stealth
					NewSequenceNode("FlankSequence",
						NewConditionNode("IsMeleeOrStealth", func(e *Entity, b *Blackboard) bool {
							// Check archetype from behavior tree name or component
							isFlanking := archetype == ArchetypeMelee || archetype == ArchetypeStealth
							log.WithFields(log.Fields{
								"archetype":   archetype.String(),
								"is_flanking": isFlanking,
							}).Debug("Evaluating flanking behavior")
							return isFlanking
						}),
						NewFlankTargetAction(world),
					),

					// Otherwise move towards target
					NewMoveToTargetAction(moveSpeed),
				),
			),
		),

		// Individual combat (no priority target)
		NewSequenceNode("IndividualCombatSequence",
			NewHasTargetCondition(),
			NewSelectorNode("IndividualActions",
				NewSequenceNode("AttackSequence",
					NewTargetInRangeCondition(attackRange),
					NewAttackTargetAction(),
				),
				NewMoveToTargetAction(moveSpeed),
			),
		),

		// Maintain formation when idle
		NewSequenceNode("FormationSequence",
			NewHasFormationTargetCondition(),
			NewMaintainFormationAction(),
		),

		// Search for target
		NewSequenceNode("SearchSequence",
			NewFindTargetAction(detectionRange, world),
		),

		// Idle wander
		NewWanderAction(moveSpeed*0.5), // Wander at half speed
	)

	log.WithFields(log.Fields{
		"archetype": archetype.String(),
		"role":      "member",
		"has_root":  root != nil,
	}).Debug("buildSquadMemberTree exit")

	return root
}

// getArchetypeParams returns common parameters for each archetype.
func getArchetypeParams(archetype EnemyArchetype) (detectionRange, attackRange, moveSpeed float64) {
	log.WithFields(log.Fields{
		"archetype": archetype.String(),
	}).Debug("getArchetypeParams entry")

	var dr, ar, ms float64

	switch archetype {
	case ArchetypeMelee:
		dr, ar, ms = 250.0, 40.0, 100.0
		log.WithFields(log.Fields{
			"archetype":       "Melee",
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Debug("Retrieved Melee archetype parameters")
	case ArchetypeRanged:
		dr, ar, ms = 300.0, 200.0, 90.0
		log.WithFields(log.Fields{
			"archetype":       "Ranged",
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Debug("Retrieved Ranged archetype parameters")
	case ArchetypeTank:
		dr, ar, ms = 200.0, 50.0, 70.0
		log.WithFields(log.Fields{
			"archetype":       "Tank",
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Debug("Retrieved Tank archetype parameters")
	case ArchetypeSupport:
		dr, ar, ms = 280.0, 150.0, 95.0
		log.WithFields(log.Fields{
			"archetype":       "Support",
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Debug("Retrieved Support archetype parameters")
	case ArchetypeStealth:
		dr, ar, ms = 220.0, 35.0, 120.0
		log.WithFields(log.Fields{
			"archetype":       "Stealth",
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Debug("Retrieved Stealth archetype parameters")
	default:
		dr, ar, ms = 250.0, 40.0, 100.0
		log.WithFields(log.Fields{
			"archetype":       archetype.String(),
			"detection_range": dr,
			"attack_range":    ar,
			"move_speed":      ms,
		}).Warn("Unknown archetype, using default parameters")
	}

	log.WithFields(log.Fields{
		"archetype":       archetype.String(),
		"detection_range": dr,
		"attack_range":    ar,
		"move_speed":      ms,
	}).Debug("getArchetypeParams exit")

	return dr, ar, ms
}
