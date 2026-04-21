// Package engine provides pre-built behavior trees for common enemy archetypes.
// This file implements standard behavior trees for Melee, Ranged, Tank, Support, and Stealth enemies.
package engine

// EnemyArchetype represents a type of enemy AI behavior.
type EnemyArchetype int

const (
	// ArchetypeMelee - aggressive close-range fighter
	ArchetypeMelee EnemyArchetype = iota
	// ArchetypeRanged - maintains distance, uses ranged attacks
	ArchetypeRanged
	// ArchetypeTank - protects allies, draws aggro
	ArchetypeTank
	// ArchetypeSupport - heals/buffs allies, debuffs enemies
	ArchetypeSupport
	// ArchetypeStealth - ambush tactics, backstab
	ArchetypeStealth
)

// String returns the string representation of an archetype.
func (a EnemyArchetype) String() string {
	switch a {
	case ArchetypeMelee:
		return "Melee"
	case ArchetypeRanged:
		return "Ranged"
	case ArchetypeTank:
		return "Tank"
	case ArchetypeSupport:
		return "Support"
	case ArchetypeStealth:
		return "Stealth"
	default:
		return "Unknown"
	}
}

// BuildBehaviorTree creates a behavior tree for the specified archetype.
func BuildBehaviorTree(archetype EnemyArchetype, world *World) *BehaviorTreeComponent {
	var root BehaviorNode
	var name string

	switch archetype {
	case ArchetypeMelee:
		root = buildMeleeBehaviorTree(world)
		name = "Melee AI"
	case ArchetypeRanged:
		root = buildRangedBehaviorTree(world)
		name = "Ranged AI"
	case ArchetypeTank:
		root = buildTankBehaviorTree(world)
		name = "Tank AI"
	case ArchetypeSupport:
		root = buildSupportBehaviorTree(world)
		name = "Support AI"
	case ArchetypeStealth:
		root = buildStealthBehaviorTree(world)
		name = "Stealth AI"
	default:
		root = buildMeleeBehaviorTree(world)
		name = "Default AI"
	}

	return NewBehaviorTreeComponent(root, name)
}

// buildMeleeBehaviorTree creates a behavior tree for melee enemies.
// Melee enemies are aggressive and chase targets relentlessly.
func buildMeleeBehaviorTree(world *World) BehaviorNode {
	const detectionRange = 250.0
	const attackRange = 40.0
	const moveSpeed = 100.0

	return NewSelectorNode("MeleeRoot",
		// Check if should flee (low health)
		NewSequenceNode("FleeSequence",
			NewHealthBelowCondition(0.2), // Below 20% health
			NewFleeFromTargetAction(moveSpeed),
		),

		// Combat sequence - highest priority
		NewSequenceNode("CombatSequence",
			NewHasTargetCondition(),
			NewSelectorNode("CombatActions",
				// Attack if in range
				NewSequenceNode("AttackSequence",
					NewTargetInRangeCondition(attackRange),
					NewAttackTargetAction(),
				),
				// Chase if out of range
				NewMoveToTargetAction(moveSpeed),
			),
		),

		// Search for target
		NewSequenceNode("SearchSequence",
			NewFindTargetAction(detectionRange, world),
		),

		// Idle - wander around
		NewSequenceNode("IdleSequence",
			NewWanderAction(moveSpeed),
			NewWaitAction(1.0), // Wait 1 second after wandering
		),
	)
}

// buildRangedBehaviorTree creates a behavior tree for ranged enemies.
// Ranged enemies maintain distance and use ranged attacks.
func buildRangedBehaviorTree(world *World) BehaviorNode {
	const detectionRange = 300.0
	const minRange = 80.0  // Minimum distance to maintain
	const maxRange = 200.0 // Maximum attack range
	const moveSpeed = 90.0

	return NewSelectorNode("RangedRoot",
		// Flee if low health
		NewSequenceNode("FleeSequence",
			NewHealthBelowCondition(0.25),
			NewFleeFromTargetAction(moveSpeed),
		),

		// Combat sequence
		NewSequenceNode("CombatSequence",
			NewHasTargetCondition(),
			NewSelectorNode("CombatActions",
				// Too close - back away
				NewSequenceNode("BackAwaySequence",
					NewTargetInRangeCondition(minRange),
					NewFleeFromTargetAction(moveSpeed*0.8), // Back away slower
				),
				// In optimal range - attack
				NewSequenceNode("AttackSequence",
					NewTargetInRangeCondition(maxRange),
					NewAttackTargetAction(),
				),
				// Too far - move closer
				NewMoveToTargetAction(moveSpeed*0.6), // Approach slowly
			),
		),

		// Search for target
		NewFindTargetAction(detectionRange, world),

		// Idle - wander
		NewSequenceNode("IdleSequence",
			NewWanderAction(moveSpeed),
			NewWaitAction(2.0),
		),
	)
}

// buildTankBehaviorTree creates a behavior tree for tank enemies.
// Tanks are durable and protect allies, less concerned with fleeing.
func buildTankBehaviorTree(world *World) BehaviorNode {
	const detectionRange = 200.0
	const attackRange = 50.0
	const moveSpeed = 70.0 // Slower than melee

	return NewSelectorNode("TankRoot",
		// Tanks rarely flee, only at very low health
		NewSequenceNode("FleeSequence",
			NewHealthBelowCondition(0.1), // Below 10% health
			NewFleeFromTargetAction(moveSpeed*0.8),
		),

		// Combat sequence
		NewSequenceNode("CombatSequence",
			NewHasTargetCondition(),
			NewSelectorNode("CombatActions",
				// Attack if in range
				NewSequenceNode("AttackSequence",
					NewTargetInRangeCondition(attackRange),
					NewAttackTargetAction(),
				),
				// Chase slowly but persistently
				NewMoveToTargetAction(moveSpeed),
			),
		),

		// Search for target
		NewFindTargetAction(detectionRange, world),

		// Idle - hold position mostly
		NewSequenceNode("IdleSequence",
			NewWaitAction(3.0),             // Wait longer
			NewWanderAction(moveSpeed*0.5), // Wander slower if at all
		),
	)
}

// buildSupportBehaviorTree creates a behavior tree for support enemies.
// Support enemies prioritize staying alive and helping allies.
func buildSupportBehaviorTree(world *World) BehaviorNode {
	const detectionRange = 280.0
	const safeDistance = 150.0
	const moveSpeed = 95.0

	return NewSelectorNode("SupportRoot",
		// Flee eagerly - support should stay alive
		NewSequenceNode("FleeSequence",
			NewHealthBelowCondition(0.4), // Below 40% health
			NewFleeFromTargetAction(moveSpeed*1.2),
		),

		// Maintain safe distance
		NewSequenceNode("MaintainDistanceSequence",
			NewHasTargetCondition(),
			NewSequenceNode("KeepDistanceActions",
				NewTargetInRangeCondition(safeDistance),
				NewFleeFromTargetAction(moveSpeed),
			),
		),

		// Search for threats
		NewFindTargetAction(detectionRange, world),

		// Idle - wander and interact with the environment (e.g. resource nodes,
		// allies in need) when no threat is detected.
		NewSequenceNode("IdleSequence",
			NewInteractWithEnvironmentNode("EnvironmentInteract", "resource", 60.0, 3.0),
			NewWanderAction(moveSpeed*0.6),
			NewWaitAction(2.5),
		),
	)
}

// buildStealthBehaviorTree creates a behavior tree for stealth enemies.
// Stealth enemies use hit-and-run tactics.
func buildStealthBehaviorTree(world *World) BehaviorNode {
	const detectionRange = 220.0
	const attackRange = 35.0 // Close for backstab
	const moveSpeed = 120.0  // Fastest archetype

	return NewSelectorNode("StealthRoot",
		// Disengage after attack (hit and run)
		NewSequenceNode("DisengageSequence",
			NewHasTargetCondition(),
			// After attacking, back away
			NewSequenceNode("HitAndRunSequence",
				NewTargetInRangeCondition(attackRange*1.5),
				NewFleeFromTargetAction(moveSpeed),
				NewWaitAction(0.5), // Brief pause
				NewClearTargetAction(),
			),
		),

		// Flee if detected at low health
		NewSequenceNode("FleeSequence",
			NewHealthBelowCondition(0.3),
			NewFleeFromTargetAction(moveSpeed*1.3),
		),

		// Ambush sequence
		NewSequenceNode("AmbushSequence",
			NewHasTargetCondition(),
			NewSelectorNode("AmbushActions",
				// Attack if in range
				NewSequenceNode("StrikeSequence",
					NewTargetInRangeCondition(attackRange),
					NewAttackTargetAction(),
				),
				// Approach quickly
				NewMoveToTargetAction(moveSpeed),
			),
		),

		// Search for target
		NewFindTargetAction(detectionRange, world),

		// Idle - patrol and hide
		NewSequenceNode("IdleSequence",
			NewWanderAction(moveSpeed*0.4), // Slow, stealthy movement
			NewWaitAction(3.0),             // Wait longer between movements
		),
	)
}
