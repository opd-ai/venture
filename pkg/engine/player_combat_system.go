// Package engine provides player combat action handling.
// This file implements PlayerCombatSystem which connects player input (Space key)
// to combat actions via the CombatSystem.
package engine

import (
	"fmt"
	"math"
)

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
		input, ok := inputComp.(InputProvider)
		if !ok {
			continue // Not an InputProvider
		}

		// Check if player pressed attack button
		if !input.IsActionPressed() {
			continue
		}

		// DEBUG: Player is trying to attack
		fmt.Printf("[PLAYER COMBAT] Entity %d pressing attack button\n", entity.ID)

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

		// Consume the input immediately to prevent multiple triggers
		input.SetActionPressed(false)

		// ALWAYS trigger attack animation, even if no target
		// This provides visual feedback that the attack button was pressed
		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
			anim := animComp.(*AnimationComponent)
			anim.SetState(AnimationStateAttack)
			fmt.Printf("[PLAYER COMBAT] Triggering attack animation (current state: %s)\n", anim.CurrentState)

			// Set OnComplete callback to return to idle/walk
			anim.OnComplete = func() {
				if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
					vel := velComp.(*VelocityComponent)
					speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
					if speed > 0.1 {
						anim.SetState(AnimationStateWalk)
					} else {
						anim.SetState(AnimationStateIdle)
					}
				}
			}
		}

		// Start cooldown even if no target (player swung at air)
		attack.ResetCooldown()

		// Find nearest enemy within attack range for damage
		maxRange := attack.Range
		target := FindNearestEnemy(s.world, entity, maxRange)

		if target == nil {
			// No enemy in range - attack animation plays but no damage
			fmt.Printf("[PLAYER COMBAT] Attack animation playing, but no target in range\n")
			continue
		}

		// Perform attack through combat system (only if target exists)
		hit := s.combatSystem.Attack(entity, target)

		if hit {
			fmt.Printf("[PLAYER COMBAT] Attack hit target entity %d\n", target.ID)
			// Attack successful - could trigger effects here
			// - Hit sound effect
			// - Screen shake
			// - Tutorial progress tracking
		}
	}
}
