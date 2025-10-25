package engine

import (
	"math/rand"
)

// StatusEffectSystem manages status effects on entities.
type StatusEffectSystem struct {
	world *World
	rng   *rand.Rand
}

// NewStatusEffectSystem creates a new status effect system.
func NewStatusEffectSystem(world *World, rng *rand.Rand) *StatusEffectSystem {
	return &StatusEffectSystem{
		world: world,
		rng:   rng,
	}
}

// Update processes all status effects.
func (s *StatusEffectSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Collect all status effect components
		var effectsToRemove []Component

		for _, comp := range entity.Components {
			if effect, ok := comp.(*StatusEffectComponent); ok {
				// Update effect duration and check for ticks
				ticked := effect.Update(deltaTime)

				if effect.IsExpired() {
					// Remove expired effects
					effectsToRemove = append(effectsToRemove, effect)
					s.removeEffectModifiers(entity, effect)
				} else if ticked {
					// Apply periodic effect
					s.applyPeriodicEffect(entity, effect)
				}
			}
		}

		// Remove expired effects
		for _, effect := range effectsToRemove {
			entity.RemoveComponent(effect.Type())
		}

		// Update shield duration
		if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
			shield := shieldComp.(*ShieldComponent)
			shield.Update(deltaTime)

			// Remove depleted shields
			if !shield.IsActive() {
				entity.RemoveComponent("shield")
			}
		}
	}
}

// applyPeriodicEffect handles tick-based status effects.
func (s *StatusEffectSystem) applyPeriodicEffect(entity *Entity, effect *StatusEffectComponent) {
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return
	}
	health := healthComp.(*HealthComponent)

	switch effect.EffectType {
	case "burning":
		// Fire DoT (damage over time)
		health.TakeDamage(effect.Magnitude)

	case "poisoned":
		// Poison DoT (ignores armor)
		health.TakeDamage(effect.Magnitude)

	case "regeneration":
		// Healing over time
		health.Heal(effect.Magnitude)
	}
}

// removeEffectModifiers removes stat modifications when effect expires.
func (s *StatusEffectSystem) removeEffectModifiers(entity *Entity, effect *StatusEffectComponent) {
	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		return
	}
	stats := statsComp.(*StatsComponent)

	switch effect.EffectType {
	case "strength":
		// Remove attack boost
		stats.Attack /= (1.0 + effect.Magnitude)

	case "weakness":
		// Remove attack penalty
		stats.Attack /= effect.Magnitude

	case "fortify":
		// Remove defense boost
		stats.Defense /= (1.0 + effect.Magnitude)

	case "vulnerability":
		// Remove defense penalty
		stats.Defense /= effect.Magnitude
	}
}

// ApplyStatusEffect applies a new status effect to an entity.
func (s *StatusEffectSystem) ApplyStatusEffect(entity *Entity, effectType string, magnitude, duration, tickInterval float64) {
	// Check if effect already exists
	for _, comp := range entity.Components {
		if existing, ok := comp.(*StatusEffectComponent); ok {
			if existing.EffectType == effectType {
				// Refresh duration if same effect type
				if duration > existing.Duration {
					existing.Duration = duration
				}
				return
			}
		}
	}

	// Create new status effect
	effect := &StatusEffectComponent{
		EffectType:   effectType,
		Duration:     duration,
		Magnitude:    magnitude,
		TickInterval: tickInterval,
		NextTick:     tickInterval,
	}

	entity.AddComponent(effect)

	// Apply immediate stat modifications
	s.applyEffectModifiers(entity, effect)
}

// applyEffectModifiers applies stat modifications when effect is added.
func (s *StatusEffectSystem) applyEffectModifiers(entity *Entity, effect *StatusEffectComponent) {
	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		return
	}
	stats := statsComp.(*StatsComponent)

	switch effect.EffectType {
	case "strength":
		// Attack boost (magnitude is percentage: 0.3 = +30%)
		stats.Attack *= (1.0 + effect.Magnitude)

	case "weakness":
		// Attack penalty (magnitude is fraction: 0.7 = 70% attack)
		stats.Attack *= effect.Magnitude

	case "fortify":
		// Defense boost
		stats.Defense *= (1.0 + effect.Magnitude)

	case "vulnerability":
		// Defense penalty
		stats.Defense *= effect.Magnitude
	}
}

// ApplyShield creates a shield on the entity.
func (s *StatusEffectSystem) ApplyShield(entity *Entity, amount, duration float64) {
	// Check if shield already exists
	if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
		// Add to existing shield
		shield := shieldComp.(*ShieldComponent)
		shield.Amount += amount
		if shield.Amount > shield.MaxAmount {
			shield.MaxAmount = shield.Amount
		}
		if duration > shield.Duration {
			shield.Duration = duration
			shield.MaxDuration = duration
		}
	} else {
		// Create new shield
		shield := &ShieldComponent{
			Amount:      amount,
			MaxAmount:   amount,
			Duration:    duration,
			MaxDuration: duration,
		}
		entity.AddComponent(shield)
	}
}

// ChainLightning applies chain lightning damage to nearby enemies.
func (s *StatusEffectSystem) ChainLightning(source, initialTarget *Entity, damage float64, chains int, range_ float64) {
	if chains <= 0 {
		return
	}

	// Apply damage to initial target
	if healthComp, hasHealth := initialTarget.GetComponent("health"); hasHealth {
		health := healthComp.(*HealthComponent)
		health.TakeDamage(damage)

		// Apply shocked effect
		s.ApplyStatusEffect(initialTarget, "shocked", 0, 2.0, 0)
	}

	// Find next chain target
	entities := s.world.GetEntities()
	var nextTarget *Entity
	minDist := range_

	for _, entity := range entities {
		if entity == source || entity == initialTarget {
			continue
		}

		// Must be an enemy
		if !isEnemyTarget(source, entity) {
			continue
		}

		// Check distance from current target
		dist := GetDistance(initialTarget, entity)
		if dist <= minDist {
			nextTarget = entity
			minDist = dist
		}
	}

	// Recursively chain to next target with reduced damage
	if nextTarget != nil {
		s.ChainLightning(source, nextTarget, damage*0.7, chains-1, range_)
	}
}

// isEnemyTarget checks if an entity is a valid enemy target.
func isEnemyTarget(caster, target *Entity) bool {
	if target == caster {
		return false
	}

	// Player has input component
	if target.HasComponent("input") {
		return false
	}

	// Must have health
	if !target.HasComponent("health") {
		return false
	}

	// Check team if available
	if casterTeam, hasCasterTeam := caster.GetComponent("team"); hasCasterTeam {
		if targetTeam, hasTargetTeam := target.GetComponent("team"); hasTargetTeam {
			ct := casterTeam.(*TeamComponent)
			tt := targetTeam.(*TeamComponent)
			return ct.IsEnemy(tt.TeamID)
		}
	}

	return true
}
