// Package engine provides the combat system for damage and status effects.
// This file implements CombatSystem which handles damage calculation, combat
// interactions, and status effect management using the combat package.
package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/sirupsen/logrus"
)

// CombatSystem handles combat interactions, damage calculation, and status effects.
type CombatSystem struct {
	rng *rand.Rand

	// Camera reference for screen shake feedback (GAP-012)
	camera *CameraSystem

	// GAP-016 REPAIR: Particle system for hit effects
	particleSystem *ParticleSystem
	world          *World
	seed           int64
	genreID        string

	// Callback for when an entity dies
	onDeathCallback func(entity *Entity)

	// Callback for when damage is dealt
	onDamageCallback func(attacker, target *Entity, damage float64)

	// Logger for combat events
	logger *logrus.Entry
}

// NewCombatSystem creates a new combat system with a given random seed.
func NewCombatSystem(seed int64) *CombatSystem {
	return NewCombatSystemWithLogger(seed, nil)
}

// NewCombatSystemWithLogger creates a new combat system with a logger.
func NewCombatSystemWithLogger(seed int64, logger *logrus.Logger) *CombatSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "combat",
			"seed":   seed,
		})
		logEntry.Debug("combat system created")
	}

	return &CombatSystem{
		rng:    rand.New(rand.NewSource(seed)),
		seed:   seed,
		logger: logEntry,
	}
}

// SetCamera sets the camera reference for screen shake feedback (GAP-012).
func (s *CombatSystem) SetCamera(camera *CameraSystem) {
	s.camera = camera
}

// GAP-016 REPAIR: SetParticleSystem sets the particle system reference for hit effects.
func (s *CombatSystem) SetParticleSystem(ps *ParticleSystem, world *World, genreID string) {
	s.particleSystem = ps
	s.world = world
	s.genreID = genreID
}

// Update implements the System interface.
// Updates attack cooldowns and processes status effects.
func (s *CombatSystem) Update(entities []*Entity, deltaTime float64) {
	// Update attack cooldowns and status effects
	for _, entity := range entities {
		// Priority 1.3: Dead entities don't progress attack cooldowns
		// but status effects continue (poison doesn't stop at death)
		isDead := entity.HasComponent("dead")

		// DEBUG: Log if player is somehow marked as dead
		if entity.HasComponent("input") && isDead {
			fmt.Printf("[COMBAT SYSTEM] WARNING: Player entity %d has 'dead' component!\n", entity.ID)
		}

		if !isDead {
			// Update attack cooldowns only for living entities
			if attackComp, ok := entity.GetComponent("attack"); ok {
				attack := attackComp.(*AttackComponent)
				beforeCooldown := attack.CooldownTimer
				attack.UpdateCooldown(deltaTime)

				// DEBUG: Log cooldown updates for player (entity with input component)
				if entity.HasComponent("input") && beforeCooldown > 0 {
					fmt.Printf("[COMBAT SYSTEM] Entity %d cooldown: %.2f → %.2f (delta: %.3f)\n",
						entity.ID, beforeCooldown, attack.CooldownTimer, deltaTime)
				}
			}
		}

		// Process status effects (for both living and dead entities)
		if statusComp, ok := entity.GetComponent("status_effect"); ok {
			status := statusComp.(*StatusEffectComponent)

			// Update status effect
			if ticked := status.Update(deltaTime); ticked {
				s.applyStatusEffectTick(entity, status)
			}

			// Remove expired effects
			if status.IsExpired() {
				entity.RemoveComponent("status_effect")
			}
		}
	}

	// Clean up dead entities
	for _, entity := range entities {
		if healthComp, ok := entity.GetComponent("health"); ok {
			health := healthComp.(*HealthComponent)
			if health.IsDead() {
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
					s.logger.WithFields(logrus.Fields{
						"entityID":      entity.ID,
						"currentHealth": health.Current,
					}).Info("entity death")
				}
				if s.onDeathCallback != nil {
					s.onDeathCallback(entity)
				}
			}
		}
	}
}

// applyStatusEffectTick applies periodic status effect damage/healing.
func (s *CombatSystem) applyStatusEffectTick(entity *Entity, effect *StatusEffectComponent) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}

	health := healthComp.(*HealthComponent)

	switch effect.EffectType {
	case "poison", "burn":
		// Damage over time
		health.TakeDamage(effect.Magnitude)
	case "regeneration":
		// Healing over time
		health.Heal(effect.Magnitude)
	}
}

// Attack performs an attack from attacker to target.
// Returns true if the attack hit, false if it missed or was invalid.
func (s *CombatSystem) Attack(attacker, target *Entity) bool {
	// Priority 1.3: Dead entities cannot attack
	if attacker.HasComponent("dead") {
		return false
	}

	// Priority 1.3: Dead entities cannot be targeted for attacks
	if target.HasComponent("dead") {
		return false
	}

	// Validate entities have required components
	attackComp, ok := attacker.GetComponent("attack")
	if !ok {
		return false
	}
	attack := attackComp.(*AttackComponent)

	// Check cooldown
	if !attack.CanAttack() {
		return false
	}

	targetHealth, ok := target.GetComponent("health")
	if !ok {
		return false
	}
	health := targetHealth.(*HealthComponent)

	// Check if target is already dead
	if health.IsDead() {
		return false
	}

	// Check range
	_, attackerHasPos := attacker.GetComponent("position")
	_, targetHasPos := target.GetComponent("position")
	if attackerHasPos && targetHasPos {
		distance := GetDistance(attacker, target)
		if distance > attack.Range {
			return false
		}
	}

	// Get attacker stats
	attackerStatsComp, _ := attacker.GetComponent("stats")
	var attackerStats *StatsComponent
	if attackerStatsComp != nil {
		attackerStats = attackerStatsComp.(*StatsComponent)
	}

	// Get target stats
	targetStatsComp, _ := target.GetComponent("stats")
	var targetStats *StatsComponent
	if targetStatsComp != nil {
		targetStats = targetStatsComp.(*StatsComponent)
	}

	// Check for evasion
	if targetStats != nil && s.rollChance(targetStats.Evasion) {
		// Attack missed
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"attackerID": attacker.ID,
				"targetID":   target.ID,
				"evasion":    targetStats.Evasion,
			}).Debug("attack evaded")
		}
		attack.ResetCooldown()
		return false
	}

	// Calculate damage
	baseDamage := attack.Damage
	isCrit := false

	// Apply attacker stats
	if attackerStats != nil {
		if attack.DamageType == combat.DamageMagical {
			baseDamage += attackerStats.MagicPower
		} else {
			baseDamage += attackerStats.Attack
		}

		// Check for critical hit
		if s.rollChance(attackerStats.CritChance) {
			baseDamage *= attackerStats.CritDamage
			isCrit = true
		}
	}

	// Apply target defense and resistances
	finalDamage := baseDamage
	if targetStats != nil {
		// Apply defense
		if attack.DamageType == combat.DamageMagical {
			finalDamage -= targetStats.MagicDefense
		} else {
			finalDamage -= targetStats.Defense
		}

		// Apply resistance
		resistance := targetStats.GetResistance(attack.DamageType)
		finalDamage *= (1.0 - resistance)
	}

	// Minimum damage
	if finalDamage < 1.0 {
		finalDamage = 1.0
	}

	// Check for shield first
	if shieldComp, hasShield := target.GetComponent("shield"); hasShield {
		shield := shieldComp.(*ShieldComponent)
		if shield.IsActive() {
			// Shield absorbs damage
			absorbed := shield.AbsorbDamage(finalDamage)
			finalDamage -= absorbed

			// If shield absorbed all damage, no health damage
			if finalDamage <= 0 {
				attack.ResetCooldown()
				return true
			}
		}
	}

	// Apply remaining damage to health
	health.TakeDamage(finalDamage)

	// Trigger attack animation for attacker
	if animComp, hasAnim := attacker.GetComponent("animation"); hasAnim {
		anim := animComp.(*AnimationComponent)

		// DEBUG: Log animation trigger
		if attacker.HasComponent("input") {
			fmt.Printf("[ATTACK ANIM] Player attacking - setting state to ATTACK (was %s)\n", anim.CurrentState)
		}

		anim.SetState(AnimationStateAttack)
		// Set callback to return to idle after attack animation completes
		anim.OnComplete = func() {
			// Check if entity is moving to set appropriate idle/walk state
			if velComp, hasVel := attacker.GetComponent("velocity"); hasVel {
				vel := velComp.(*VelocityComponent)
				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
				if speed > 0.1 {
					anim.SetState(AnimationStateWalk)
				} else {
					anim.SetState(AnimationStateIdle)
				}
			} else {
				anim.SetState(AnimationStateIdle)
			}

			if attacker.HasComponent("input") {
				fmt.Printf("[ATTACK ANIM] Player attack complete - returning to idle/walk\n")
			}
		}
	}

	// Trigger hurt animation for target
	if animComp, hasAnim := target.GetComponent("animation"); hasAnim {
		anim := animComp.(*AnimationComponent)
		anim.SetState(AnimationStateHit)
		// Set a callback to return to idle after hurt animation
		anim.OnComplete = func() {
			// Check if entity is moving to set appropriate idle/walk state
			if velComp, hasVel := target.GetComponent("velocity"); hasVel {
				vel := velComp.(*VelocityComponent)
				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
				if speed > 0.1 {
					anim.SetState(AnimationStateWalk)
				} else {
					anim.SetState(AnimationStateIdle)
				}
			} else {
				anim.SetState(AnimationStateIdle)
			}
		}
	}

	// Log damage event
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"attackerID":   attacker.ID,
			"targetID":     target.ID,
			"damage":       finalDamage,
			"baseDamage":   baseDamage,
			"damageType":   attack.DamageType,
			"critical":     isCrit,
			"targetHealth": health.Current,
		}).Info("damage dealt")
	}

	// GAP-016 REPAIR: Spawn hit particles at target position
	if s.particleSystem != nil && s.world != nil {
		if posComp, ok := target.GetComponent("position"); ok {
			pos := posComp.(*PositionComponent)
			// Use timestamp for particle seed variation
			particleSeed := s.seed + int64(pos.X*1000) + int64(pos.Y*1000)
			s.particleSystem.SpawnHitSparks(s.world, pos.X, pos.Y, particleSeed, s.genreID)
		}
	}

	// GAP-012 REPAIR: Trigger hit flash on damage
	if feedbackComp, ok := target.GetComponent("visual_feedback"); ok {
		feedback := feedbackComp.(*VisualFeedbackComponent)
		// Flash intensity scales with damage (0.3-1.0 range)
		flashIntensity := 0.3 + (finalDamage / 100.0)
		if flashIntensity > 1.0 {
			flashIntensity = 1.0
		}
		feedback.TriggerFlash(flashIntensity)
	}

	// GAP-012 REPAIR: Trigger screen shake on damage
	if s.camera != nil {
		// Shake intensity scales with damage (0.1-0.5 range for subtlety)
		shakeIntensity := (finalDamage / 100.0) * 5.0
		if shakeIntensity > 5.0 {
			shakeIntensity = 5.0
		}
		s.camera.Shake(shakeIntensity)
	}

	// Reset cooldown
	attack.ResetCooldown()

	// Trigger callback
	if s.onDamageCallback != nil {
		s.onDamageCallback(attacker, target, finalDamage)
	}

	return true
}

// rollChance returns true if a random roll succeeds based on the given chance (0.0 to 1.0).
func (s *CombatSystem) rollChance(chance float64) bool {
	if chance <= 0 {
		return false
	}
	if chance >= 1.0 {
		return true
	}
	return s.rng.Float64() < chance
}

// CanAttackTarget checks if an attacker can attack a target (range and cooldown check).
func (s *CombatSystem) CanAttackTarget(attacker, target *Entity) bool {
	attackComp, ok := attacker.GetComponent("attack")
	if !ok {
		return false
	}
	attack := attackComp.(*AttackComponent)

	if !attack.CanAttack() {
		return false
	}

	targetHealth, ok := target.GetComponent("health")
	if !ok || targetHealth.(*HealthComponent).IsDead() {
		return false
	}

	// Check range if both have positions
	_, attackerHasPos := attacker.GetComponent("position")
	_, targetHasPos := target.GetComponent("position")
	if attackerHasPos && targetHasPos {
		distance := GetDistance(attacker, target)
		if distance > attack.Range {
			return false
		}
	}

	return true
}

// ApplyStatusEffect applies a status effect to an entity.
func (s *CombatSystem) ApplyStatusEffect(target *Entity, effectType string, duration, magnitude, tickInterval float64) {
	effect := &StatusEffectComponent{
		EffectType:   effectType,
		Duration:     duration,
		Magnitude:    magnitude,
		TickInterval: tickInterval,
		NextTick:     tickInterval,
	}

	// Replace any existing status effect (simplification)
	target.AddComponent(effect)
}

// Heal heals a target entity by the given amount.
func (s *CombatSystem) Heal(target *Entity, amount float64) {
	healthComp, ok := target.GetComponent("health")
	if !ok {
		return
	}

	health := healthComp.(*HealthComponent)
	health.Heal(amount)
}

// SetDeathCallback sets the callback function for entity deaths.
func (s *CombatSystem) SetDeathCallback(callback func(entity *Entity)) {
	s.onDeathCallback = callback
}

// SetDamageCallback sets the callback function for damage dealt.
func (s *CombatSystem) SetDamageCallback(callback func(attacker, target *Entity, damage float64)) {
	s.onDamageCallback = callback
}

// FindEnemiesInRange finds all enemy entities within the given range of the attacker.
func FindEnemiesInRange(world *World, attacker *Entity, maxRange float64) []*Entity {
	_, ok := attacker.GetComponent("position")
	if !ok {
		return nil
	}

	attackerTeam, _ := attacker.GetComponent("team")
	var attackerTeamID int
	if attackerTeam != nil {
		attackerTeamID = attackerTeam.(*TeamComponent).TeamID
	}

	enemies := make([]*Entity, 0)

	for _, entity := range world.GetEntities() {
		if entity.ID == attacker.ID {
			continue
		}

		// Priority 1.3: Skip dead entities - they cannot be targeted
		if entity.HasComponent("dead") {
			continue
		}

		// Check team
		targetTeam, hasTeam := entity.GetComponent("team")
		if hasTeam {
			team := targetTeam.(*TeamComponent)
			if !team.IsEnemy(attackerTeamID) {
				continue
			}
		}

		// Check health
		healthComp, hasHealth := entity.GetComponent("health")
		if !hasHealth || healthComp.(*HealthComponent).IsDead() {
			continue
		}

		// Check range
		_, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		distance := GetDistance(attacker, entity)
		if distance <= maxRange {
			enemies = append(enemies, entity)
		}
	}

	return enemies
}

// FindNearestEnemy finds the closest enemy to the attacker within the given range.
func FindNearestEnemy(world *World, attacker *Entity, maxRange float64) *Entity {
	enemies := FindEnemiesInRange(world, attacker, maxRange)
	if len(enemies) == 0 {
		return nil
	}

	var nearest *Entity
	nearestDistance := math.MaxFloat64

	for _, enemy := range enemies {
		distance := GetDistance(attacker, enemy)
		if distance < nearestDistance {
			nearestDistance = distance
			nearest = enemy
		}
	}

	return nearest
}
