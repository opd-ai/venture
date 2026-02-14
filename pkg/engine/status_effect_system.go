package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectTickCallback is called when a status effect ticks (deals damage/healing).
type StatusEffectTickCallback func(entity *Entity, effectType string, magnitude float64)

// StatusEffectSystem manages status effects on entities.
type StatusEffectSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	// Reusable buffer for expired effects to reduce per-entity allocations
	expiredEffectsBuffer []Component

	// Callback for status effect tick events (for visual feedback)
	tickCallback StatusEffectTickCallback
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
		world:                world,
		rng:                  rng,
		logger:               logEntry,
		expiredEffectsBuffer: make([]Component, 0, 8), // Pre-allocate for typical max expired effects per entity
	}
}

// SetTickCallback sets the callback invoked when status effects tick (deal damage/healing).
func (s *StatusEffectSystem) SetTickCallback(callback StatusEffectTickCallback) {
	s.tickCallback = callback
	if s.logger != nil {
		s.logger.Debug("Status effect tick callback set")
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
		// Reuse buffer to avoid per-entity allocations
		s.expiredEffectsBuffer = s.expiredEffectsBuffer[:0]
		s.processStatusEffectsInto(entity, deltaTime)
		s.removeExpiredEffects(entity, s.expiredEffectsBuffer)
		s.updateShieldComponent(entity, deltaTime)
	}

	if s.logger != nil {
		s.logger.Debug("Status effect system update completed")
	}
}

// processStatusEffectsInto processes status effects for an entity and appends expired effects to the system buffer.
func (s *StatusEffectSystem) processStatusEffectsInto(entity *Entity, deltaTime float64) {
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
			s.handleExpiredEffectBuffer(entity, effect)
		} else if ticked {
			s.handleTickedEffect(entity, effect)
		}
	}
}

// processStatusEffects processes status effects for an entity and returns expired effects.
// Deprecated: Use processStatusEffectsInto for better performance with buffer reuse.
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

// handleExpiredEffectBuffer handles an expired status effect using the system buffer.
func (s *StatusEffectSystem) handleExpiredEffectBuffer(entity *Entity, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType,
		}).Debug("Status effect expired")
	}
	s.expiredEffectsBuffer = append(s.expiredEffectsBuffer, effect)
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
	s.logPeriodicEffect(entity.ID, effect)

	health := s.getHealthComponent(entity, effect.EffectType)
	if health == nil {
		return
	}

	s.applyEffectToHealth(entity, effect, health)
}

// logPeriodicEffect logs the periodic effect application.
func (s *StatusEffectSystem) logPeriodicEffect(entityID uint64, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Applying periodic effect to entity")
	}
}

// getHealthComponent retrieves and validates the health component.
func (s *StatusEffectSystem) getHealthComponent(entity *Entity, effectType string) *HealthComponent {
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		s.logMissingHealth(entity.ID, effectType)
		return nil
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		s.logHealthTypeError(entity.ID)
		return nil
	}

	return health
}

// logMissingHealth logs when an entity lacks a health component.
func (s *StatusEffectSystem) logMissingHealth(entityID uint64, effectType string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"effect_type": effectType,
		}).Warn("Entity has no health component for periodic effect")
	}
}

// logHealthTypeError logs when health component type assertion fails.
func (s *StatusEffectSystem) logHealthTypeError(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "health",
		}).Warn("Failed to type assert health component")
	}
}

// applyEffectToHealth applies the specific effect type to health.
func (s *StatusEffectSystem) applyEffectToHealth(entity *Entity, effect *StatusEffectComponent, health *HealthComponent) {
	switch effect.EffectType {
	case "burning":
		s.applyBurning(entity, effect.Magnitude, health)
	case "poisoned":
		s.applyPoison(entity, effect.Magnitude, health)
	case "regeneration":
		s.applyRegeneration(entity, effect.Magnitude, health)
	}
}

// applyBurning applies fire damage over time.
func (s *StatusEffectSystem) applyBurning(entity *Entity, magnitude float64, health *HealthComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"damage":    magnitude,
		}).Debug("Applying burning damage")
	}
	health.TakeDamage(magnitude)
	s.invokeTickCallback(entity, "burning", magnitude)
}

// applyPoison applies poison damage that ignores armor.
func (s *StatusEffectSystem) applyPoison(entity *Entity, magnitude float64, health *HealthComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"damage":    magnitude,
		}).Debug("Applying poison damage")
	}
	health.TakeDamage(magnitude)
	s.invokeTickCallback(entity, "poisoned", magnitude)
}

// applyRegeneration applies healing over time.
func (s *StatusEffectSystem) applyRegeneration(entity *Entity, magnitude float64, health *HealthComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"healing":   magnitude,
		}).Debug("Applying regeneration healing")
	}
	health.Heal(magnitude)
	s.invokeTickCallback(entity, "regeneration", magnitude)
}

// invokeTickCallback safely invokes the tick callback if set.
func (s *StatusEffectSystem) invokeTickCallback(entity *Entity, effectType string, magnitude float64) {
	if s.tickCallback != nil {
		s.tickCallback(entity, effectType, magnitude)
	}
}

// removeEffectModifiers removes stat modifications when effect expires.
func (s *StatusEffectSystem) removeEffectModifiers(entity *Entity, effect *StatusEffectComponent) {
	s.logModifierRemoval(entity, effect)

	stats := s.getStatsComponent(entity, effect)
	if stats == nil {
		return
	}

	s.removeStatModifier(entity, stats, effect)
}

// logModifierRemoval logs the start of modifier removal.
func (s *StatusEffectSystem) logModifierRemoval(entity *Entity, effect *StatusEffectComponent) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"effect_type": effect.EffectType,
		"magnitude":   effect.Magnitude,
	}).Debug("Removing effect modifiers from entity")
}

// getStatsComponent retrieves and validates the stats component.
func (s *StatusEffectSystem) getStatsComponent(entity *Entity, effect *StatusEffectComponent) *StatsComponent {
	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		s.logMissingStats(entity, effect)
		return nil
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		s.logInvalidStats(entity)
		return nil
	}

	return stats
}

// logMissingStats logs when entity has no stats component.
func (s *StatusEffectSystem) logMissingStats(entity *Entity, effect *StatusEffectComponent) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"effect_type": effect.EffectType,
	}).Debug("Entity has no stats component, skipping modifier removal")
}

// logInvalidStats logs stats component type assertion failure.
func (s *StatusEffectSystem) logInvalidStats(entity *Entity) {
	if s.logger == nil {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"component_type": "stats",
	}).Warn("Failed to type assert stats component")
}

// removeStatModifier removes the stat modifier based on effect type.
func (s *StatusEffectSystem) removeStatModifier(entity *Entity, stats *StatsComponent, effect *StatusEffectComponent) {
	switch effect.EffectType {
	case "strength":
		s.removeStrengthModifier(entity, stats, effect)
	case "weakness":
		s.removeWeaknessModifier(entity, stats, effect)
	case "fortify":
		s.removeFortifyModifier(entity, stats, effect)
	case "vulnerability":
		s.removeVulnerabilityModifier(entity, stats, effect)
	}
}

// removeStrengthModifier removes attack boost from strength effect.
func (s *StatusEffectSystem) removeStrengthModifier(entity *Entity, stats *StatsComponent, effect *StatusEffectComponent) {
	oldAttack := stats.Attack
	stats.Attack /= (1.0 + effect.Magnitude)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"old_attack": oldAttack,
			"new_attack": stats.Attack,
		}).Debug("Removed strength effect modifier")
	}
}

// removeWeaknessModifier removes attack penalty from weakness effect.
func (s *StatusEffectSystem) removeWeaknessModifier(entity *Entity, stats *StatsComponent, effect *StatusEffectComponent) {
	oldAttack := stats.Attack
	if effect.Magnitude > 0 {
		stats.Attack /= effect.Magnitude
	}
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"old_attack": oldAttack,
			"new_attack": stats.Attack,
		}).Debug("Removed weakness effect modifier")
	}
}

// removeFortifyModifier removes defense boost from fortify effect.
func (s *StatusEffectSystem) removeFortifyModifier(entity *Entity, stats *StatsComponent, effect *StatusEffectComponent) {
	oldDefense := stats.Defense
	stats.Defense /= (1.0 + effect.Magnitude)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"old_defense": oldDefense,
			"new_defense": stats.Defense,
		}).Debug("Removed fortify effect modifier")
	}
}

// removeVulnerabilityModifier removes defense penalty from vulnerability effect.
func (s *StatusEffectSystem) removeVulnerabilityModifier(entity *Entity, stats *StatsComponent, effect *StatusEffectComponent) {
	oldDefense := stats.Defense
	if effect.Magnitude > 0 {
		stats.Defense /= effect.Magnitude
	}
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"old_defense": oldDefense,
			"new_defense": stats.Defense,
		}).Debug("Removed vulnerability effect modifier")
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
	s.logEffectApplication(entity.ID, effect)

	stats := s.getAndValidateStatsComponent(entity, effect.EffectType)
	if stats == nil {
		return
	}

	switch effect.EffectType {
	case "strength":
		s.applyStrengthEffect(entity.ID, stats, effect.Magnitude)
	case "weakness":
		s.applyWeaknessEffect(entity.ID, stats, effect.Magnitude)
	case "fortify":
		s.applyFortifyEffect(entity.ID, stats, effect.Magnitude)
	case "vulnerability":
		s.applyVulnerabilityEffect(entity.ID, stats, effect.Magnitude)
	}
}

// logEffectApplication logs the start of effect modifier application.
func (s *StatusEffectSystem) logEffectApplication(entityID uint64, effect *StatusEffectComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"effect_type": effect.EffectType,
			"magnitude":   effect.Magnitude,
		}).Debug("Applying effect modifiers to entity")
	}
}

// getAndValidateStatsComponent retrieves and validates the stats component.
func (s *StatusEffectSystem) getAndValidateStatsComponent(entity *Entity, effectType string) *StatsComponent {
	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		s.logMissingStatsComponent(entity.ID, effectType)
		return nil
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		s.logStatsTypeAssertionFailed(entity.ID)
		return nil
	}

	return stats
}

// logMissingStatsComponent logs when an entity lacks a stats component.
func (s *StatusEffectSystem) logMissingStatsComponent(entityID uint64, effectType string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"effect_type": effectType,
		}).Debug("Entity has no stats component, skipping modifier application")
	}
}

// logStatsTypeAssertionFailed logs when stats component type assertion fails.
func (s *StatusEffectSystem) logStatsTypeAssertionFailed(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "stats",
		}).Warn("Failed to type assert stats component")
	}
}

// applyStrengthEffect applies an attack boost modifier.
func (s *StatusEffectSystem) applyStrengthEffect(entityID uint64, stats *StatsComponent, magnitude float64) {
	oldAttack := stats.Attack
	stats.Attack *= (1.0 + magnitude)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"old_attack": oldAttack,
			"new_attack": stats.Attack,
			"boost_pct":  magnitude * 100,
		}).Debug("Applied strength effect modifier")
	}
}

// applyWeaknessEffect applies an attack penalty modifier.
func (s *StatusEffectSystem) applyWeaknessEffect(entityID uint64, stats *StatsComponent, magnitude float64) {
	oldAttack := stats.Attack
	stats.Attack *= magnitude

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"old_attack":  oldAttack,
			"new_attack":  stats.Attack,
			"penalty_pct": (1.0 - magnitude) * 100,
		}).Debug("Applied weakness effect modifier")
	}
}

// applyFortifyEffect applies a defense boost modifier.
func (s *StatusEffectSystem) applyFortifyEffect(entityID uint64, stats *StatsComponent, magnitude float64) {
	oldDefense := stats.Defense
	stats.Defense *= (1.0 + magnitude)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"old_defense": oldDefense,
			"new_defense": stats.Defense,
			"boost_pct":   magnitude * 100,
		}).Debug("Applied fortify effect modifier")
	}
}

// applyVulnerabilityEffect applies a defense penalty modifier.
func (s *StatusEffectSystem) applyVulnerabilityEffect(entityID uint64, stats *StatsComponent, magnitude float64) {
	oldDefense := stats.Defense
	stats.Defense *= magnitude

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"old_defense": oldDefense,
			"new_defense": stats.Defense,
			"penalty_pct": (1.0 - magnitude) * 100,
		}).Debug("Applied vulnerability effect modifier")
	}
}

// ApplyShield creates a shield on the entity.
func (s *StatusEffectSystem) ApplyShield(entity *Entity, amount, duration float64) {
	logShieldApplication(s.logger, entity.ID, amount, duration)

	if enhanceExistingShield(s.logger, entity, amount, duration) {
		return
	}

	createNewShield(s.logger, entity, amount, duration)
}

// logShieldApplication logs the shield application attempt.
func logShieldApplication(logger *logrus.Entry, entityID uint64, amount, duration float64) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"amount":    amount,
			"duration":  duration,
		}).Debug("Applying shield to entity")
	}
}

// enhanceExistingShield attempts to enhance an existing shield.
func enhanceExistingShield(logger *logrus.Entry, entity *Entity, amount, duration float64) bool {
	shieldComp, hasShield := entity.GetComponent("shield")
	if !hasShield {
		return false
	}

	shield, ok := shieldComp.(*ShieldComponent)
	if !ok {
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "shield",
			}).Warn("Failed to type assert shield component, creating new shield")
		}
		return false
	}

	oldAmount := shield.Amount
	shield.Amount += amount
	if shield.Amount > shield.MaxAmount {
		shield.MaxAmount = shield.Amount
	}
	if duration > shield.Duration {
		shield.Duration = duration
		shield.MaxDuration = duration
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"old_amount":   oldAmount,
			"new_amount":   shield.Amount,
			"max_amount":   shield.MaxAmount,
			"new_duration": shield.Duration,
		}).Debug("Enhanced existing shield")
	}
	return true
}

// createNewShield creates a new shield component for the entity.
func createNewShield(logger *logrus.Entry, entity *Entity, amount, duration float64) {
	shield := &ShieldComponent{
		Amount:      amount,
		MaxAmount:   amount,
		Duration:    duration,
		MaxDuration: duration,
	}
	entity.AddComponent(shield)

	if logger != nil {
		logger.WithFields(logrus.Fields{
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
