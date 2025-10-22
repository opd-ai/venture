// Package engine provides the combat system for damage and status effects.
// This file implements CombatSystem which handles damage calculation, combat
// interactions, and status effect management using the combat package.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/combat"
)

// CombatSystem handles combat interactions, damage calculation, and status effects.
type CombatSystem struct {
	rng *rand.Rand

	// Callback for when an entity dies
	onDeathCallback func(entity *Entity)

	// Callback for when damage is dealt
	onDamageCallback func(attacker, target *Entity, damage float64)
}

// NewCombatSystem creates a new combat system with a given random seed.
func NewCombatSystem(seed int64) *CombatSystem {
	return &CombatSystem{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Update implements the System interface.
// Updates attack cooldowns and processes status effects.
func (s *CombatSystem) Update(entities []*Entity, deltaTime float64) {
	// Update attack cooldowns
	for _, entity := range entities {
		if attackComp, ok := entity.GetComponent("attack"); ok {
			attack := attackComp.(*AttackComponent)
			attack.UpdateCooldown(deltaTime)
		}

		// Process status effects
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
		attack.ResetCooldown()
		return false
	}

	// Calculate damage
	baseDamage := attack.Damage

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

	// Apply damage
	health.TakeDamage(finalDamage)

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
