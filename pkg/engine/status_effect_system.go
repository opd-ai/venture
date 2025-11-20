package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectSystem manages status effects on entities.
type StatusEffectSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
}

// NewStatusEffectSystem creates a new status effect system.
func NewStatusEffectSystem(world *World, rng *rand.Rand) *StatusEffectSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "status_effect")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Status effect system created")
		}
	}
	return &StatusEffectSystem{
		world:  world,
		rng:    rng,
		logger: logEntry,
	}
}

// Update processes all status effects.
func (s *StatusEffectSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("Status effect system update started")
	}

	for _, entity := range entities {
		effectsToRemove := s.processStatusEffects(entity, deltaTime)
		s.removeExpiredEffects(entity, effectsToRemove)
		s.updateShieldComponent(entity, deltaTime)
	}

	if s.logger != nil {
		s.logger.Debug("Status effect system update completed")
	}
}

// processStatusEffects processes status effects for an entity and returns expired effects.
func (s *StatusEffectSystem) processStatusEffects(entity *Entity, deltaTime float64) []Component {
	var effectsToRemove []Component

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"effect_type":   effect.EffectType,
				"magnitude":     effect.Magnitude,
				"duration":      effect.Duration,
				"tick_interval": effect.TickInterval,
			}).Debug("Processing status effect")
		}

		ticked := effect.Update(deltaTime)

		if effect.IsExpired() {
			s.handleExpiredEffect(entity, effect, &effectsToRemove)
		} else if ticked {
			s.handleTickedEffect(entity, effect)
		}
	}

	return effectsToRemove
}

// handleExpiredEffect handles an expired status effect.
func (s *StatusEffectSystem) handleExpiredEffect(entity *Entity, effect *StatusEffectComponent, effectsToRemove *[]Component) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
		}).Debug("Status effect expired")
	}
	*effectsToRemove = append(*effectsToRemove, effect)
	s.removeEffectModifiers(entity, effect)
}

// handleTickedEffect handles a ticked status effect.
func (s *StatusEffectSystem) handleTickedEffect(entity *Entity, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Applying periodic effect")
	}
	s.applyPeriodicEffect(entity, effect)
}

// removeExpiredEffects removes expired effects from entity and returns them to pool.
func (s *StatusEffectSystem) removeExpiredEffects(entity *Entity, effectsToRemove []Component) {
	for _, effect := range effectsToRemove {
		entity.RemoveComponent(effect.Type())
		statusEffect, ok := effect.(*StatusEffectComponent)
		if !ok {
			continue
		}
		ReleaseStatusEffect(statusEffect)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"effect_type": statusEffect.EffectType,
			}).Debug("Status effect removed and returned to pool")
		}
	}
}

// updateShieldComponent updates the shield component for an entity.
func (s *StatusEffectSystem) updateShieldComponent(entity *Entity, deltaTime float64) {
	shieldComp, hasShield := entity.GetComponent("shield")
	if !hasShield {
		return
	}

	shield, ok := shieldComp.(*ShieldComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "shield",
			}).Warn("Failed to type assert shield component")
		}
		return
	}

	shield.Update(deltaTime)

	if !shield.IsActive() {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Shield depleted, removing component")
		}
		entity.RemoveComponent("shield")
	}
}

// applyPeriodicEffect handles tick-based status effects.
func (s *StatusEffectSystem) applyPeriodicEffect(entity *Entity, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Applying periodic effect to entity")
	}

	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"effect_type": effect.EffectType,
			}).Warn("Entity has no health component for periodic effect")
		}
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "health",
			}).Warn("Failed to type assert health component")
		}
		return
	}

	switch effect.EffectType {
	case "burning":
		// Fire DoT (damage over time)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"damage":    effect.Magnitude,
			}).Debug("Applying burning damage")
		}
		health.TakeDamage(effect.Magnitude)

	case "poisoned":
		// Poison DoT (ignores armor)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"damage":    effect.Magnitude,
			}).Debug("Applying poison damage")
		}
		health.TakeDamage(effect.Magnitude)

	case "regeneration":
		// Healing over time
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"healing":   effect.Magnitude,
			}).Debug("Applying regeneration healing")
		}
		health.Heal(effect.Magnitude)
	}
}

// removeEffectModifiers removes stat modifications when effect expires.
func (s *StatusEffectSystem) removeEffectModifiers(entity *Entity, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Removing effect modifiers from entity")
	}

	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"effect_type": effect.EffectType,
			}).Debug("Entity has no stats component, skipping modifier removal")
		}
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "stats",
			}).Warn("Failed to type assert stats component")
		}
		return
	}

	switch effect.EffectType {
	case "strength":
		oldAttack := stats.Attack
		// Remove attack boost
		stats.Attack /= (1.0 + effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"old_attack": oldAttack,
				"new_attack": stats.Attack,
			}).Debug("Removed strength effect modifier")
		}

	case "weakness":
		oldAttack := stats.Attack
		// Remove attack penalty
		stats.Attack /= effect.Magnitude
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"old_attack": oldAttack,
				"new_attack": stats.Attack,
			}).Debug("Removed weakness effect modifier")
		}

	case "fortify":
		oldDefense := stats.Defense
		// Remove defense boost
		stats.Defense /= (1.0 + effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"old_defense": oldDefense,
				"new_defense": stats.Defense,
			}).Debug("Removed fortify effect modifier")
		}

	case "vulnerability":
		oldDefense := stats.Defense
		// Remove defense penalty
		stats.Defense /= effect.Magnitude
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"old_defense": oldDefense,
				"new_defense": stats.Defense,
			}).Debug("Removed vulnerability effect modifier")
		}
	}
}

// ApplyStatusEffect applies a new status effect to an entity.
func (s *StatusEffectSystem) ApplyStatusEffect(entity *Entity, effectType string, magnitude, duration, tickInterval float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"effect_type":   effectType,
			"magnitude":     magnitude,
			"duration":      duration,
			"tick_interval": tickInterval,
		}).Debug("Applying status effect to entity")
	}

	// Check if effect already exists
	for _, comp := range entity.Components {
		if existing, ok := comp.(*StatusEffectComponent); ok {
			if existing.EffectType == effectType {
				// Refresh duration if same effect type
				if duration > existing.Duration {
					if s.logger != nil {
						s.logger.WithFields(logrus.Fields{
							"entity_id":    entity.ID,
							"effect_type":  effectType,
							"old_duration": existing.Duration,
							"new_duration": duration,
						}).Debug("Refreshing existing status effect duration")
					}
					existing.Duration = duration
				}
				return
			}
		}
	}

	// Create new status effect from pool
	effect := NewStatusEffectComponent(effectType, magnitude, duration, tickInterval)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"effect_type":   effectType,
			"magnitude":     magnitude,
			"duration":      duration,
			"tick_interval": tickInterval,
		}).Debug("Created new status effect component")
	}

	entity.AddComponent(effect)

	// Apply immediate stat modifications
	s.applyEffectModifiers(entity, effect)
}

// applyEffectModifiers applies stat modifications when effect is added.
func (s *StatusEffectSystem) applyEffectModifiers(entity *Entity, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Applying effect modifiers to entity")
	}

	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"effect_type": effect.EffectType,
			}).Debug("Entity has no stats component, skipping modifier application")
		}
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "stats",
			}).Warn("Failed to type assert stats component")
		}
		return
	}

	switch effect.EffectType {
	case "strength":
		oldAttack := stats.Attack
		// Attack boost (magnitude is percentage: 0.3 = +30%)
		stats.Attack *= (1.0 + effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"old_attack": oldAttack,
				"new_attack": stats.Attack,
				"boost_pct":  effect.Magnitude * 100,
			}).Debug("Applied strength effect modifier")
		}

	case "weakness":
		oldAttack := stats.Attack
		// Attack penalty (magnitude is fraction: 0.7 = 70% attack)
		stats.Attack *= effect.Magnitude
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"old_attack":  oldAttack,
				"new_attack":  stats.Attack,
				"penalty_pct": (1.0 - effect.Magnitude) * 100,
			}).Debug("Applied weakness effect modifier")
		}

	case "fortify":
		oldDefense := stats.Defense
		// Defense boost
		stats.Defense *= (1.0 + effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"old_defense": oldDefense,
				"new_defense": stats.Defense,
				"boost_pct":   effect.Magnitude * 100,
			}).Debug("Applied fortify effect modifier")
		}

	case "vulnerability":
		oldDefense := stats.Defense
		// Defense penalty
		stats.Defense *= effect.Magnitude
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"old_defense": oldDefense,
				"new_defense": stats.Defense,
				"penalty_pct": (1.0 - effect.Magnitude) * 100,
			}).Debug("Applied vulnerability effect modifier")
		}
	}
}

// ApplyShield creates a shield on the entity.
func (s *StatusEffectSystem) ApplyShield(entity *Entity, amount, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"amount":    amount,
			"duration":  duration,
		}).Debug("Applying shield to entity")
	}

	// Check if shield already exists
	if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
		// Add to existing shield
		shield, ok := shieldComp.(*ShieldComponent)
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"component_type": "shield",
				}).Warn("Failed to type assert shield component, creating new shield")
			}
			// Create new shield if type assertion fails (fall through)
		} else {
			oldAmount := shield.Amount
			shield.Amount += amount
			if shield.Amount > shield.MaxAmount {
				shield.MaxAmount = shield.Amount
			}
			if duration > shield.Duration {
				shield.Duration = duration
				shield.MaxDuration = duration
			}
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":    entity.ID,
					"old_amount":   oldAmount,
					"new_amount":   shield.Amount,
					"max_amount":   shield.MaxAmount,
					"new_duration": shield.Duration,
				}).Debug("Enhanced existing shield")
			}
			return
		}
	}

	// Create new shield
	shield := &ShieldComponent{
		Amount:      amount,
		MaxAmount:   amount,
		Duration:    duration,
		MaxDuration: duration,
	}
	entity.AddComponent(shield)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"amount":    amount,
			"duration":  duration,
		}).Debug("Created new shield component")
	}
}

// ChainLightning applies chain lightning damage to nearby enemies.
func (s *StatusEffectSystem) ChainLightning(source, initialTarget *Entity, damage float64, chains int, range_ float64) {
	s.logChainStart(source, initialTarget, damage, chains, range_)
	if chains <= 0 {
		s.logChainTerminated()
		return
	}
	s.applyLightningDamage(initialTarget, damage)
	s.continueChain(source, initialTarget, damage, chains, range_)
}

// logChainStart logs the initiation of chain lightning
func (s *StatusEffectSystem) logChainStart(source, initialTarget *Entity, damage float64, chains int, range_ float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"source_id": source.ID,
			"target_id": initialTarget.ID,
			"damage":    damage,
			"chains":    chains,
			"range":     range_,
		}).Debug("Chain lightning initiated")
	}
}

// logChainTerminated logs when chain lightning ends due to no remaining chains
func (s *StatusEffectSystem) logChainTerminated() {
	if s.logger != nil {
		s.logger.Debug("Chain lightning terminated, no chains remaining")
	}
}

// applyLightningDamage applies damage and shocked effect to target
func (s *StatusEffectSystem) applyLightningDamage(target *Entity, damage float64) {
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		s.logHealthTypeAssertFailed(target)
		return
	}
	health.TakeDamage(damage)
	s.logDamageApplied(target, damage)
	s.ApplyStatusEffect(target, "shocked", 0, 2.0, 0)
}

// logHealthTypeAssertFailed logs failure to convert health component
func (s *StatusEffectSystem) logHealthTypeAssertFailed(target *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":      target.ID,
			"component_type": "health",
		}).Warn("Failed to type assert health component for chain lightning")
	}
}

// logDamageApplied logs successful damage application
func (s *StatusEffectSystem) logDamageApplied(target *Entity, damage float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"damage":    damage,
		}).Debug("Applied chain lightning damage to target")
	}
}

// continueChain finds next target and recursively chains lightning
func (s *StatusEffectSystem) continueChain(source, currentTarget *Entity, damage float64, chains int, range_ float64) {
	nextTarget := s.findNextChainTarget(source, currentTarget, range_)
	if nextTarget != nil {
		s.logChainContinue(nextTarget, damage, chains)
		s.ChainLightning(source, nextTarget, damage*0.7, chains-1, range_)
	} else {
		s.logNoChainTarget(chains)
	}
}

// findNextChainTarget searches for the nearest valid enemy to chain to
func (s *StatusEffectSystem) findNextChainTarget(source, currentTarget *Entity, range_ float64) *Entity {
	s.logSearchingForTarget(range_)
	entities := s.world.GetEntities()
	var nextTarget *Entity
	minDist := range_

	for _, entity := range entities {
		if entity == source || entity == currentTarget {
			continue
		}
		if !isEnemyTarget(source, entity) {
			continue
		}
		dist := GetDistance(currentTarget, entity)
		if dist <= minDist {
			nextTarget = entity
			minDist = dist
		}
	}
	return nextTarget
}

// logSearchingForTarget logs the start of next target search
func (s *StatusEffectSystem) logSearchingForTarget(range_ float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"total_entities": len(s.world.GetEntities()),
			"search_range":   range_,
		}).Debug("Searching for next chain lightning target")
	}
}

// logChainContinue logs successful chain to next target
func (s *StatusEffectSystem) logChainContinue(nextTarget *Entity, damage float64, chains int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"next_target_id":   nextTarget.ID,
			"remaining_chains": chains - 1,
			"reduced_damage":   damage * 0.7,
		}).Debug("Chaining lightning to next target")
	}
}

// logNoChainTarget logs when no valid target is found for chaining
func (s *StatusEffectSystem) logNoChainTarget(chains int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"remaining_chains": chains - 1,
		}).Debug("No valid chain lightning target found, chain terminated")
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
			ct, ok := casterTeam.(*TeamComponent)
			if !ok {
				return true // Default to enemy if type assertion fails
			}
			tt, ok := targetTeam.(*TeamComponent)
			if !ok {
				return true
			}
			return ct.IsEnemy(tt.TeamID)
		}
	}

	return true
}
