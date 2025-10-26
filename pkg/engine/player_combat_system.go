// Package engine provides player combat action handling.
// This file implements PlayerCombatSystem which connects player input (Space key)
// to combat actions via the CombatSystem.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// PlayerCombatSystem processes player combat input and triggers attacks.
// It bridges the InputSystem (which captures Space key) and CombatSystem (which applies damage).
type PlayerCombatSystem struct {
	combatSystem *CombatSystem
	world        *World
	logger       *logrus.Entry
}

// NewPlayerCombatSystem creates a new player combat system.
func NewPlayerCombatSystem(combatSystem *CombatSystem, world *World) *PlayerCombatSystem {
	return NewPlayerCombatSystemWithLogger(combatSystem, world, nil)
}

// NewPlayerCombatSystemWithLogger creates a new player combat system with a logger.
func NewPlayerCombatSystemWithLogger(combatSystem *CombatSystem, world *World, logger *logrus.Logger) *PlayerCombatSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "player_combat",
		})
	}
	return &PlayerCombatSystem{
		combatSystem: combatSystem,
		world:        world,
		logger:       logEntry,
	}
}

// Update processes player combat input for all player-controlled entities.
// This system must run AFTER InputSystem but BEFORE MovementSystem.
func (s *PlayerCombatSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Skip dead entities - they cannot attack (Category 1.1)
		if entity.HasComponent("dead") {
			continue
		}

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

		// Get attack component
		attackComp, ok := entity.GetComponent("attack")
		if !ok {
			continue // Entity can't attack
		}
		attack := attackComp.(*AttackComponent)

		// Check if attack is ready (cooldown)
		if !attack.CanAttack() {
			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entityID":         entity.ID,
					"cooldownRemaining": attack.CooldownTimer,
				}).Debug("attack on cooldown")
			}
			continue // Still on cooldown
		}

		// Consume the input immediately to prevent multiple triggers
		input.SetActionPressed(false)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entityID": entity.ID,
				"cooldown": attack.Cooldown,
			}).Debug("attack triggered")
		}

		// ALWAYS trigger attack animation, even if no target
		// This provides visual feedback that the attack button was pressed
		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
			anim := animComp.(*AnimationComponent)
			anim.SetState(AnimationStateAttack)

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

		// Find nearest enemy within attack range for damage
		maxRange := attack.Range
		target := FindNearestEnemy(s.world, entity, maxRange)

		if target == nil {
			// No enemy in range - attack animation plays but no damage
			// Start cooldown even if no target (player swung at air)
			attack.ResetCooldown()
			continue
		}

		// Perform attack through combat system (only if target exists)
		// Note: CombatSystem.Attack() handles cooldown reset internally
		hit := s.combatSystem.Attack(entity, target)

		if hit && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entityID": entity.ID,
				"targetID": target.ID,
			}).Debug("attack hit target")
		}

		if hit {
			// Attack successful - could trigger effects here
			// - Hit sound effect
			// - Screen shake
			// - Tutorial progress tracking
		}
	}
}
