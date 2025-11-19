// Package engine provides player combat action handling.
// This file implements PlayerCombatSystem which connects player input (Space key)
// to combat actions via the CombatSystem.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// DefaultAimCone is the default aim cone width for forgiving aim (45° = π/4 radians).
// This provides a balance between precision and ease of use, particularly for mobile controls.
const DefaultAimCone = math.Pi / 4

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
		if entity.HasComponent("dead") {
			continue
		}

		input, attack := s.validateCombatComponents(entity)
		if input == nil || attack == nil {
			continue
		}

		if !input.IsActionPressed() {
			continue
		}

		if !s.checkAttackReadiness(entity, attack) {
			continue
		}

		input.SetActionPressed(false)
		s.logAttackTriggered(entity, attack)
		s.triggerAttackAnimation(entity)

		target := s.findAttackTarget(entity, attack)
		if target == nil {
			attack.ResetCooldown()
			continue
		}

		hit := s.combatSystem.Attack(entity, target)
		s.logAttackHit(entity, target, hit)
	}
}

// validateCombatComponents checks if entity has required components for combat and returns them.
func (s *PlayerCombatSystem) validateCombatComponents(entity *Entity) (InputProvider, *AttackComponent) {
	inputComp, ok := entity.GetComponent("input")
	if !ok {
		return nil, nil
	}
	input, ok := inputComp.(InputProvider)
	if !ok {
		return nil, nil
	}

	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return nil, nil
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return nil, nil
	}

	return input, attack
}

// checkAttackReadiness verifies if the attack can proceed based on cooldown status.
func (s *PlayerCombatSystem) checkAttackReadiness(entity *Entity, attack *AttackComponent) bool {
	if !attack.CanAttack() {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entityID":          entity.ID,
				"cooldownRemaining": attack.CooldownTimer,
			}).Debug("attack on cooldown")
		}
		return false
	}
	return true
}

// logAttackTriggered logs when an attack is successfully triggered.
func (s *PlayerCombatSystem) logAttackTriggered(entity *Entity, attack *AttackComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID": entity.ID,
			"cooldown": attack.Cooldown,
		}).Debug("attack triggered")
	}
}

// triggerAttackAnimation starts the attack animation and sets up the completion callback.
func (s *PlayerCombatSystem) triggerAttackAnimation(entity *Entity) {
	animComp, hasAnim := entity.GetComponent("animation")
	if !hasAnim {
		return
	}

	anim, ok := animComp.(*AnimationComponent)
	if !ok {
		return
	}

	anim.SetState(AnimationStateAttack)
	anim.OnComplete = func() {
		s.setPostAttackAnimation(entity, anim)
	}
}

// setPostAttackAnimation determines the animation state after attack completes.
func (s *PlayerCombatSystem) setPostAttackAnimation(entity *Entity, anim *AnimationComponent) {
	velComp, hasVel := entity.GetComponent("velocity")
	if !hasVel {
		return
	}

	vel, ok := velComp.(*VelocityComponent)
	if !ok {
		return
	}

	speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
	if speed > 0.1 {
		anim.SetState(AnimationStateWalk)
	} else {
		anim.SetState(AnimationStateIdle)
	}
}

// findAttackTarget locates an enemy target using aim-based or proximity-based selection.
func (s *PlayerCombatSystem) findAttackTarget(entity *Entity, attack *AttackComponent) *Entity {
	maxRange := attack.Range
	aimComp, hasAim := entity.GetComponent("aim")
	if !hasAim {
		return FindNearestEnemy(s.world, entity, maxRange)
	}

	aim, ok := aimComp.(*AimComponent)
	if !ok {
		return FindNearestEnemy(s.world, entity, maxRange)
	}

	target := FindEnemyInAimDirection(s.world, entity, aim.AimAngle, maxRange, DefaultAimCone)
	s.logAimTargetSelection(entity, aim, maxRange, target)
	return target
}

// logAimTargetSelection logs the aim-based target selection process.
func (s *PlayerCombatSystem) logAimTargetSelection(entity *Entity, aim *AimComponent, maxRange float64, target *Entity) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	targetID := uint64(0)
	if target != nil {
		targetID = target.ID
	}

	s.logger.WithFields(logrus.Fields{
		"entityID":  entity.ID,
		"aimAngle":  aim.AimAngle,
		"aimDegree": aim.AimAngle * 180 / math.Pi,
		"range":     maxRange,
		"aimCone":   DefaultAimCone * 180 / math.Pi,
		"targetID":  targetID,
	}).Debug("aim-based attack target selection")
}

// logAttackHit logs when an attack successfully hits a target.
func (s *PlayerCombatSystem) logAttackHit(entity, target *Entity, hit bool) {
	if hit && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID": entity.ID,
			"targetID": target.ID,
		}).Debug("attack hit target")
	}
}
