// Package engine provides squad-aware behavior tree builders.
// This file implements behavior trees for squad members with tactical coordination.
package engine

// BuildSquadBehaviorTree creates a behavior tree for squad members with coordination.
// This extends the base archetype with squad-specific behaviors.
func BuildSquadBehaviorTree(archetype EnemyArchetype, world *World, isLeader bool) *BehaviorTreeComponent {
	var root BehaviorNode
	var name string

	if isLeader {
		root = buildSquadLeaderTree(archetype, world)
		name = archetype.String() + " Squad Leader"
	} else {
		root = buildSquadMemberTree(archetype, world)
		name = archetype.String() + " Squad Member"
	}

	return NewBehaviorTreeComponent(root, name)
}

// buildSquadLeaderTree creates a behavior tree for squad leaders.
// Leaders make tactical decisions for the squad and coordinate member actions.
func buildSquadLeaderTree(archetype EnemyArchetype, world *World) BehaviorNode {
	// Get base parameters from archetype
	detectionRange, attackRange, moveSpeed := getArchetypeParams(archetype)

	return NewSelectorNode("SquadLeaderRoot",
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
}

// buildSquadMemberTree creates a behavior tree for squad members.
// Members follow squad tactics and maintain formation.
func buildSquadMemberTree(archetype EnemyArchetype, world *World) BehaviorNode {
	// Get base parameters from archetype
	detectionRange, attackRange, moveSpeed := getArchetypeParams(archetype)

	return NewSelectorNode("SquadMemberRoot",
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
							return archetype == ArchetypeMelee || archetype == ArchetypeStealth
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
}

// getArchetypeParams returns common parameters for each archetype.
func getArchetypeParams(archetype EnemyArchetype) (detectionRange, attackRange, moveSpeed float64) {
	switch archetype {
	case ArchetypeMelee:
		return 250.0, 40.0, 100.0
	case ArchetypeRanged:
		return 300.0, 200.0, 90.0
	case ArchetypeTank:
		return 200.0, 50.0, 70.0
	case ArchetypeSupport:
		return 280.0, 150.0, 95.0
	case ArchetypeStealth:
		return 220.0, 35.0, 120.0
	default:
		return 250.0, 40.0, 100.0
	}
}
