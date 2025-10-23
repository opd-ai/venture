// Package engine provides player combat action handling.
// This file implements PlayerCombatSystem which connects player input (Space key)
// to combat actions via the CombatSystem.
package engine

// PlayerCombatSystem processes player combat input and triggers attacks.
// It bridges the InputSystem (which captures Space key) and CombatSystem (which applies damage).
type PlayerCombatSystem struct {
	combatSystem *CombatSystem
	world        *World
}

// NewPlayerCombatSystem creates a new player combat system.
func NewPlayerCombatSystem(combatSystem *CombatSystem, world *World) *PlayerCombatSystem {
	return &PlayerCombatSystem{
		combatSystem: combatSystem,
		world:        world,
	}
}

// Update processes player combat input for all player-controlled entities.
// This system must run AFTER InputSystem but BEFORE MovementSystem.
func (s *PlayerCombatSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check for input component (player-controlled entities only)
		inputComp, ok := entity.GetComponent("input")
		if !ok {
			continue
		}
		input := inputComp.(*InputComponent)

		// Check if player pressed attack button
		if !input.ActionPressed {
			continue
		}

		// Get attack component
		attackComp, ok := entity.GetComponent("attack")
		if !ok {
			continue // Entity can't attack
		}
		attack := attackComp.(*AttackComponent)

		// Check if attack is ready (cooldown)
		if !attack.CanAttack() {
			continue // Still on cooldown
		}

		// Find nearest enemy within attack range
		maxRange := attack.Range
		target := FindNearestEnemy(s.world, entity, maxRange)

		// Consume the input immediately to prevent multiple triggers
		input.ActionPressed = false

		if target == nil {
			// No enemy in range - attack fails silently
			// Could add feedback later (swing animation, miss sound)
			continue
		}

		// Perform attack through combat system
		hit := s.combatSystem.Attack(entity, target)

		if hit {
			// Attack successful - could trigger effects here
			// - Attack animation
			// - Hit sound effect
			// - Screen shake
			// - Tutorial progress tracking
		}
	}
}
