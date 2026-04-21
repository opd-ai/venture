package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ManaComponent tracks entity's magical energy.
type ManaComponent struct {
	Current int
	Max     int
	Regen   float64 // Mana regenerated per second
}

// Type implements Component interface.
func (m *ManaComponent) Type() string {
	return "mana"
}

// SpellSlotComponent stores equipped spells in slots 1-5.
type SpellSlotComponent struct {
	Slots      [5]*magic.Spell
	Cooldowns  [5]float64 // Remaining cooldown time for each slot
	CastingBar float64    // Progress of current cast (0.0 to 1.0)
	Casting    int        // Which slot is being cast (-1 = none)
}

// Type implements Component interface.
func (s *SpellSlotComponent) Type() string {
	return "spell_slots"
}

// GetSlot returns the spell in the given slot (0-4), or nil if empty.
func (s *SpellSlotComponent) GetSlot(slot int) *magic.Spell {
	if slot < 0 || slot >= 5 {
		return nil
	}
	return s.Slots[slot]
}

// SetSlot assigns a spell to a slot.
func (s *SpellSlotComponent) SetSlot(slot int, spell *magic.Spell) {
	if slot >= 0 && slot < 5 {
		s.Slots[slot] = spell
	}
}

// IsOnCooldown returns true if the slot is on cooldown.
func (s *SpellSlotComponent) IsOnCooldown(slot int) bool {
	if slot < 0 || slot >= 5 {
		return true
	}
	return s.Cooldowns[slot] > 0
}

// IsCasting returns true if currently casting a spell.
func (s *SpellSlotComponent) IsCasting() bool {
	return s.Casting >= 0
}

// SpellCastingSystem handles spell execution and cooldowns.
type SpellCastingSystem struct {
	world                 *World
	statusEffectSys       *StatusEffectSystem
	particleSys           *ParticleSystem           // For visual effects
	audioMgr              *AudioManager             // For sound effects
	tutorialSys           *EbitenTutorialSystem     // For notifications
	comboSys              *SpellCombinationSystem   // For combo detection
	weatherSpellDamageSys *WeatherSpellDamageSystem // For weather-element synergies
	logger                *logrus.Entry
}

// NewSpellCastingSystem creates a new spell casting system.
func NewSpellCastingSystem(world *World, statusEffectSys *StatusEffectSystem) *SpellCastingSystem {
	return NewSpellCastingSystemWithLogger(world, statusEffectSys, nil)
}

// NewSpellCastingSystemWithLogger creates a new spell casting system with a logger.
func NewSpellCastingSystemWithLogger(world *World, statusEffectSys *StatusEffectSystem, logger *logrus.Logger) *SpellCastingSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logger.ReportCaller = true
		logEntry = logger.WithField("system", "spell_casting")
		logEntry.Debug("Creating new spell casting system")
	}
	sys := &SpellCastingSystem{
		world:           world,
		statusEffectSys: statusEffectSys,
		particleSys:     NewParticleSystem(),
		audioMgr:        nil, // Will be set via SetAudioManager()
		logger:          logEntry,
	}
	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"has_status_effect_sys": statusEffectSys != nil,
		}).Debug("Spell casting system created")
	}
	return sys
}

// SetAudioManager sets the audio manager for sound effects.
// This allows deferred initialization when audio system is ready.
func (s *SpellCastingSystem) SetAudioManager(audioMgr *AudioManager) {
	if s.logger != nil {
		s.logger.Debug("AudioManager set for spell casting system")
	}
	s.audioMgr = audioMgr
}

// SetParticleSystem sets the particle system for visual effects.
// This allows using a shared particle system if desired.
func (s *SpellCastingSystem) SetParticleSystem(particleSys *ParticleSystem) {
	if s.logger != nil {
		s.logger.Debug("ParticleSystem set for spell casting system")
	}
	s.particleSys = particleSys
}

// SetTutorialSystem sets the tutorial system for notifications.
// This allows displaying feedback messages to the player.
func (s *SpellCastingSystem) SetTutorialSystem(tutorialSys *EbitenTutorialSystem) {
	if s.logger != nil {
		s.logger.Debug("TutorialSystem set for spell casting system")
	}
	s.tutorialSys = tutorialSys
}

// SetComboSystem sets the spell combination system for combo detection.
// This allows using a shared combo system if desired.
func (s *SpellCastingSystem) SetComboSystem(comboSys *SpellCombinationSystem) {
	if s.logger != nil {
		s.logger.Debug("SpellCombinationSystem set for spell casting system")
	}
	s.comboSys = comboSys
}

// SetWeatherSpellDamageSystem sets the weather-spell damage modifier system.
// This enables elemental spell damage bonuses based on current weather conditions
// (e.g., lightning damage +25% in rain, fire damage -25% in rain).
func (s *SpellCastingSystem) SetWeatherSpellDamageSystem(weatherSys *WeatherSpellDamageSystem) {
	if s.logger != nil {
		s.logger.Debug("WeatherSpellDamageSystem set for spell casting system")
	}
	s.weatherSpellDamageSys = weatherSys
}

// Update processes spell casting and cooldowns.
func (s *SpellCastingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("SpellCastingSystem update started")
	}

	for _, entity := range entities {
		slots := s.getAndValidateSpellSlots(entity)
		if slots == nil {
			continue
		}

		s.updateCooldowns(slots, deltaTime)
		s.updateCastingProgress(entity, slots, deltaTime)
	}
}

// getAndValidateSpellSlots retrieves and validates the spell slots component.
func (s *SpellCastingSystem) getAndValidateSpellSlots(entity *Entity) *SpellSlotComponent {
	spellComp, hasSpells := entity.GetComponent("spell_slots")
	if !hasSpells {
		return nil
	}
	slots, ok := spellComp.(*SpellSlotComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "spell_slots",
			}).Warn("Failed to cast spell_slots component to SpellSlotComponent")
		}
		return nil
	}
	return slots
}

// updateCooldowns decrements all active spell cooldowns.
func (s *SpellCastingSystem) updateCooldowns(slots *SpellSlotComponent, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"delta_time": deltaTime,
		}).Debug("Updating spell cooldowns")
	}

	for i := range slots.Cooldowns {
		if slots.Cooldowns[i] > 0 {
			oldCooldown := slots.Cooldowns[i]
			slots.Cooldowns[i] -= deltaTime
			if slots.Cooldowns[i] < 0 {
				slots.Cooldowns[i] = 0
			}
			if s.logger != nil && oldCooldown > 0 && slots.Cooldowns[i] == 0 {
				s.logger.WithFields(logrus.Fields{
					"slot_index": i,
				}).Debug("Spell cooldown completed")
			}
		}
	}
}

// updateCastingProgress advances spell casting progress and completes casts.
func (s *SpellCastingSystem) updateCastingProgress(entity *Entity, slots *SpellSlotComponent, deltaTime float64) {
	if !slots.IsCasting() {
		return
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"casting_slot": slots.Casting,
			"progress":     slots.CastingBar,
			"delta_time":   deltaTime,
		}).Debug("Updating casting progress")
	}

	spell := slots.GetSlot(slots.Casting)
	if spell == nil {
		s.cancelInvalidCast(entity, slots)
		return
	}

	s.advanceCastingBar(slots, spell, deltaTime)

	if slots.CastingBar >= 1.0 {
		s.completeCast(entity, slots, spell)
	}
}

// cancelInvalidCast cancels a cast when the spell is no longer valid.
func (s *SpellCastingSystem) cancelInvalidCast(entity *Entity, slots *SpellSlotComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"slot_index": slots.Casting,
		}).Warn("Casting spell is nil, canceling cast")
	}
	slots.Casting = -1
	slots.CastingBar = 0
}

// advanceCastingBar progresses the casting bar based on cast time.
func (s *SpellCastingSystem) advanceCastingBar(slots *SpellSlotComponent, spell *magic.Spell, deltaTime float64) {
	oldProgress := slots.CastingBar
	if spell.Stats.CastTime > 0 {
		slots.CastingBar += deltaTime / spell.Stats.CastTime
	} else {
		slots.CastingBar = 1.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"spell_name": spell.Name,
			}).Debug("Instant cast spell")
		}
	}

	if s.logger != nil && oldProgress != slots.CastingBar {
		s.logger.WithFields(logrus.Fields{
			"spell_name":   spell.Name,
			"old_progress": oldProgress,
			"new_progress": slots.CastingBar,
			"cast_time":    spell.Stats.CastTime,
		}).Debug("Casting bar advanced")
	}
}

// completeCast finalizes a spell cast and resets casting state.
func (s *SpellCastingSystem) completeCast(entity *Entity, slots *SpellSlotComponent, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"spell_name": spell.Name,
			"spell_type": spell.Type.String(),
			"slot_index": slots.Casting,
			"mana_cost":  spell.Stats.ManaCost,
		}).Info("Spell cast completed")
	}
	s.executeCast(entity, spell, slots.Casting)
	slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown
	slots.Casting = -1
	slots.CastingBar = 0
}

// executeCast performs the spell effect.
func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) {
	s.logSpellExecution(caster, spell, slotIndex)

	mana := s.validateAndConsumeMana(caster, spell)
	if mana == nil {
		return
	}

	pos := s.getCasterPosition(caster, spell)
	if pos == nil {
		return
	}

	s.dispatchSpellByType(caster, spell, pos)
	s.applySpellEffects(caster, spell, pos, slotIndex)
}

// logSpellExecution logs the initial spell cast event with relevant details.
func (s *SpellCastingSystem) logSpellExecution(caster *Entity, spell *magic.Spell, slotIndex int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"spell_type": spell.Type.String(),
			"element":    spell.Element.String(),
			"slot_index": slotIndex,
		}).Debug("Executing spell cast")
	}
}

// validateAndConsumeMana checks if the caster has sufficient mana and deducts the spell cost.
// Returns the ManaComponent if successful, nil otherwise.
func (s *SpellCastingSystem) validateAndConsumeMana(caster *Entity, spell *magic.Spell) *ManaComponent {
	mana := s.getManaComponent(caster, spell.Name)
	if mana == nil {
		return nil
	}

	if !s.hasEnoughMana(mana, spell) {
		return nil
	}

	s.consumeMana(caster.ID, mana, spell.Stats.ManaCost)
	return mana
}

// getManaComponent retrieves and validates the mana component.
func (s *SpellCastingSystem) getManaComponent(caster *Entity, spellName string) *ManaComponent {
	manaComp, hasMana := caster.GetComponent("mana")
	if !hasMana {
		s.logMissingMana(caster.ID, spellName)
		return nil
	}

	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		s.logManaTypeError(caster.ID)
		return nil
	}

	return mana
}

// logMissingMana logs when an entity lacks a mana component.
func (s *SpellCastingSystem) logMissingMana(entityID uint64, spellName string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"spell_name": spellName,
		}).Warn("Entity has no mana component, cannot cast spell")
	}
}

// logManaTypeError logs when mana component type assertion fails.
func (s *SpellCastingSystem) logManaTypeError(entityID uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "mana",
		}).Error("Failed to cast mana component to ManaComponent")
	}
}

// hasEnoughMana checks if the caster has sufficient mana.
func (s *SpellCastingSystem) hasEnoughMana(mana *ManaComponent, spell *magic.Spell) bool {
	if mana.Current >= spell.Stats.ManaCost {
		return true
	}

	s.logInsufficientMana(mana, spell)
	if s.tutorialSys != nil {
		s.tutorialSys.ShowNotification("Not enough mana!", 1.5)
	}
	return false
}

// logInsufficientMana logs when insufficient mana is available.
func (s *SpellCastingSystem) logInsufficientMana(mana *ManaComponent, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     "unknown",
			"spell_name":    spell.Name,
			"mana_current":  mana.Current,
			"mana_required": spell.Stats.ManaCost,
		}).Debug("Insufficient mana for spell cast")
	}
}

// consumeMana deducts mana cost and logs the consumption.
func (s *SpellCastingSystem) consumeMana(entityID uint64, mana *ManaComponent, manaCost int) {
	previousMana := mana.Current
	mana.Current -= manaCost
	if mana.Current < 0 {
		mana.Current = 0
	}

	s.logManaConsumption(entityID, previousMana, manaCost, mana.Current)
}

// logManaConsumption logs mana consumption details.
func (s *SpellCastingSystem) logManaConsumption(entityID uint64, before, deducted, remaining int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"mana_before":    before,
			"mana_deducted":  deducted,
			"mana_remaining": remaining,
		}).Info("Mana consumed for spell cast")
	}
}

// getCasterPosition retrieves and validates the caster's position component.
// Returns the PositionComponent if successful, nil otherwise.
func (s *SpellCastingSystem) getCasterPosition(caster *Entity, spell *magic.Spell) *PositionComponent {
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Warn("Caster has no position component")
		}
		return nil
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      caster.ID,
				"component_type": "position",
			}).Error("Failed to cast position component to PositionComponent")
		}
		return nil
	}
	return pos
}

// dispatchSpellByType routes the spell to the appropriate casting method based on spell type.
func (s *SpellCastingSystem) dispatchSpellByType(caster *Entity, spell *magic.Spell, pos *PositionComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_type": spell.Type.String(),
			"spell_name": spell.Name,
			"position_x": pos.X,
			"position_y": pos.Y,
		}).Debug("Applying spell effect")
	}

	switch spell.Type {
	case magic.TypeOffensive:
		s.castOffensiveSpell(caster, spell, pos.X, pos.Y)
	case magic.TypeHealing:
		s.castHealingSpell(caster, spell)
	case magic.TypeDefensive:
		s.castDefensiveSpell(caster, spell)
	case magic.TypeBuff:
		s.castBuffSpell(caster, spell)
	case magic.TypeDebuff:
		s.castDebuffSpell(caster, spell, pos.X, pos.Y)
	case magic.TypeUtility:
		s.castUtilitySpell(caster, spell)
	}
}

// applySpellEffects handles audio, visual particles, lighting, and combo tracking after spell dispatch.
func (s *SpellCastingSystem) applySpellEffects(caster *Entity, spell *magic.Spell, pos *PositionComponent, slotIndex int) {
	if s.audioMgr != nil {
		effectType := "magic"
		if err := s.audioMgr.PlaySFX(effectType, int64(caster.ID)); err != nil {
			if s.logger != nil {
				s.logger.Debugf("Failed to play spell cast sound: %v", err)
			}
		}
	}

	if s.particleSys != nil {
		s.particleSys.SpawnMagicParticles(s.world, pos.X, pos.Y, int64(caster.ID), "fantasy")
	}

	s.spawnSpellLight(pos.X, pos.Y, spell, 2.5)

	if s.comboSys != nil {
		s.comboSys.OnSpellCast(caster, spell, slotIndex)
	}
}

// castOffensiveSpell deals damage to enemies in range.
func (s *SpellCastingSystem) castOffensiveSpell(caster *Entity, spell *magic.Spell, x, y float64) {
	s.logSpellCasting(caster, spell)
	targets := s.findTargets(caster, spell, x, y)
	s.logTargetsFound(caster, spell, targets)

	for _, target := range targets {
		s.applySpellDamageToTarget(caster, target, spell)
	}
}

// logSpellCasting logs the spell casting event.
func (s *SpellCastingSystem) logSpellCasting(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   caster.ID,
			"spell_name":  spell.Name,
			"target_type": spell.Target.String(),
			"damage":      spell.Stats.Damage,
		}).Debug("Casting offensive spell")
	}
}

// logTargetsFound logs the number of targets found.
func (s *SpellCastingSystem) logTargetsFound(caster *Entity, spell *magic.Spell, targets []*Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    caster.ID,
			"spell_name":   spell.Name,
			"target_count": len(targets),
		}).Debug("Targets found for offensive spell")
	}
}

// applySpellDamageToTarget applies damage, effects, and feedback to a single target.
func (s *SpellCastingSystem) applySpellDamageToTarget(caster, target *Entity, spell *magic.Spell) {
	health := s.getTargetHealth(target)
	if health == nil {
		return
	}

	damage, comboMultiplier := s.calculateSpellDamage(caster, spell)
	previousHealth := health.Current
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}

	if s.logger != nil && previousHealth > 0 && health.Current == 0 {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"target_id": target.ID,
		}).Info("Target killed by spell damage")
	}

	s.logDamageApplied(caster, target, spell, damage, comboMultiplier, health)
	s.applySpellEffectsAndFeedback(target, spell)
}

// getTargetHealth retrieves the health component from a target entity.
func (s *SpellCastingSystem) getTargetHealth(target *Entity) *HealthComponent {
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
			}).Debug("Target has no health component")
		}
		return nil
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id":      target.ID,
				"component_type": "health",
			}).Warn("Failed to cast health component")
		}
		return nil
	}
	return health
}

// calculateSpellDamage computes final damage with combo and weather multipliers.
func (s *SpellCastingSystem) calculateSpellDamage(caster *Entity, spell *magic.Spell) (float64, float64) {
	damage := float64(spell.Stats.Damage)
	comboMultiplier := 1.0

	// Apply combo multiplier if available
	if s.comboSys != nil {
		comboMultiplier = s.comboSys.GetActiveComboMultiplier(caster)
		damage *= comboMultiplier
		if s.logger != nil && comboMultiplier > 1.0 {
			s.logger.WithFields(logrus.Fields{
				"entity_id":        caster.ID,
				"base_damage":      spell.Stats.Damage,
				"combo_multiplier": comboMultiplier,
				"final_damage":     damage,
			}).Debug("Combo multiplier applied to spell damage")
		}
	}

	// Apply weather-element damage modifier if available
	if s.weatherSpellDamageSys != nil {
		weatherMod := s.weatherSpellDamageSys.GetDamageModifier(spell.Element)
		if weatherMod != 1.0 {
			damage *= weatherMod
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":        caster.ID,
					"spell_element":    spell.Element.String(),
					"weather_modifier": weatherMod,
					"final_damage":     damage,
				}).Debug("Weather modifier applied to spell damage")
			}
		}
	}

	return damage, comboMultiplier
}

// logDamageApplied logs the damage application details.
func (s *SpellCastingSystem) logDamageApplied(caster, target *Entity, spell *magic.Spell,
	damage, comboMultiplier float64, health *HealthComponent,
) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":        caster.ID,
			"target_id":        target.ID,
			"spell_name":       spell.Name,
			"base_damage":      spell.Stats.Damage,
			"combo_multiplier": comboMultiplier,
			"final_damage":     damage,
			"health_remaining": health.Current,
		}).Info("Offensive spell damage applied")
	}
}

// applySpellEffectsAndFeedback applies elemental effects, visuals, and audio.
func (s *SpellCastingSystem) applySpellEffectsAndFeedback(target *Entity, spell *magic.Spell) {
	s.logSpellApplication(target, spell)
	s.applyTargetSpellEffects(target, spell)
	s.createSpellFeedback(target, spell)
}

// logSpellApplication logs the spell application event.
func (s *SpellCastingSystem) logSpellApplication(target *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":  target.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Applying spell effects and feedback")
	}
}

// applyTargetSpellEffects applies the elemental effects of the spell to the target.
func (s *SpellCastingSystem) applyTargetSpellEffects(target *Entity, spell *magic.Spell) {
	if s.statusEffectSys != nil {
		s.applyElementalEffect(target, spell)
	}
}

// createSpellFeedback creates visual and audio feedback for the spell.
func (s *SpellCastingSystem) createSpellFeedback(target *Entity, spell *magic.Spell) {
	s.createVisualFeedback(target, spell)
	s.createAudioFeedback(target)
}

// createVisualFeedback spawns visual particle effects for the spell hit.
func (s *SpellCastingSystem) createVisualFeedback(target *Entity, spell *magic.Spell) {
	if s.particleSys == nil {
		return
	}

	targetPos, hasPos := target.GetComponent("position")
	if !hasPos {
		return
	}

	if pos, ok := targetPos.(*PositionComponent); ok {
		s.spawnElementalHitEffect(pos.X, pos.Y, spell.Element, target.ID)
	}
}

// createAudioFeedback plays impact sound effect.
func (s *SpellCastingSystem) createAudioFeedback(target *Entity) {
	if s.audioMgr == nil {
		return
	}

	if err := s.audioMgr.PlaySFX("impact", int64(target.ID)); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
				"error":     err.Error(),
			}).Debug("Failed to play impact sound")
		}
	}
}

// castHealingSpell restores health to caster or allies.
func (s *SpellCastingSystem) castHealingSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   caster.ID,
			"spell_name":  spell.Name,
			"target_type": spell.Target.String(),
			"healing":     spell.Stats.Healing,
		}).Debug("Casting healing spell")
	}

	target := caster
	if spell.Target == magic.TargetSingle {
		// Find nearest injured ally in range
		ally := s.findNearestInjuredAlly(caster, spell.Stats.Range)
		if ally != nil {
			target = ally
		}
	} else if spell.Target == magic.TargetArea || spell.Target == magic.TargetAllAllies {
		// Heal multiple allies
		allies := s.findAlliesInRange(caster, spell.Stats.AreaSize)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
				"ally_count": len(allies),
			}).Debug("Healing multiple allies")
		}
		for _, ally := range allies {
			s.healTarget(caster, ally, spell)
		}
		return
	}

	s.healTarget(caster, target, spell)
}

// healTarget applies healing to a single target.
func (s *SpellCastingSystem) healTarget(caster, target *Entity, spell *magic.Spell) {
	s.logHealingAttempt(caster, target, spell)

	health := s.getHealthComponent(target)
	if health == nil {
		return
	}

	healing := s.calculateHealing(caster, spell)
	s.applyHealing(health, healing)
	s.logHealingApplied(caster, target, spell, healing, health)
	s.spawnHealingParticles(target)
	s.playHealingSound(target)
}

// logHealingAttempt logs the start of a healing action.
func (s *SpellCastingSystem) logHealingAttempt(caster, target *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"target_id":  target.ID,
			"spell_name": spell.Name,
		}).Debug("Healing target")
	}
}

// getHealthComponent retrieves and validates the health component from target.
func (s *SpellCastingSystem) getHealthComponent(target *Entity) *HealthComponent {
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
			}).Debug("Target has no health component for healing")
		}
		return nil
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id":      target.ID,
				"component_type": "health",
			}).Warn("Failed to cast health component for healing")
		}
		return nil
	}

	return health
}

// calculateHealing calculates the total healing amount including combo multiplier.
func (s *SpellCastingSystem) calculateHealing(caster *Entity, spell *magic.Spell) float64 {
	healing := float64(spell.Stats.Healing)
	if s.comboSys != nil {
		comboMultiplier := s.comboSys.GetActiveComboMultiplier(caster)
		healing *= comboMultiplier
	}
	return healing
}

// applyHealing applies healing to the health component.
func (s *SpellCastingSystem) applyHealing(health *HealthComponent, healing float64) {
	health.Current += healing
	if health.Current > health.Max {
		health.Current = health.Max
	}
}

// logHealingApplied logs successful healing application.
func (s *SpellCastingSystem) logHealingApplied(caster, target *Entity, spell *magic.Spell, healing float64, health *HealthComponent) {
	if s.logger == nil {
		return
	}

	comboMultiplier := 1.0
	if s.comboSys != nil {
		comboMultiplier = s.comboSys.GetActiveComboMultiplier(caster)
	}

	s.logger.WithFields(logrus.Fields{
		"caster_id":        caster.ID,
		"target_id":        target.ID,
		"spell_name":       spell.Name,
		"base_healing":     spell.Stats.Healing,
		"combo_multiplier": comboMultiplier,
		"final_healing":    healing,
		"health_after":     health.Current,
		"health_max":       health.Max,
	}).Info("Healing applied")
}

// spawnHealingParticles creates visual healing particle effects.
func (s *SpellCastingSystem) spawnHealingParticles(target *Entity) {
	if s.particleSys == nil {
		return
	}

	targetPos, hasPos := target.GetComponent("position")
	if !hasPos {
		return
	}

	pos, ok := targetPos.(*PositionComponent)
	if !ok {
		return
	}

	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    20,
		GenreID:  "fantasy",
		Seed:     int64(target.ID),
		Duration: 1.0,
		SpreadX:  60.0,
		SpreadY:  60.0,
		Gravity:  -80.0,
		MinSize:  4.0,
		MaxSize:  8.0,
		Custom:   map[string]interface{}{"color": "healing"},
	}
	s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":  target.ID,
			"position_x": pos.X,
			"position_y": pos.Y,
		}).Debug("Spawned healing particles")
	}
}

// playHealingSound plays the healing sound effect.
func (s *SpellCastingSystem) playHealingSound(target *Entity) {
	if s.audioMgr == nil {
		return
	}

	if err := s.audioMgr.PlaySFX("powerup", int64(target.ID)); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
				"error":     err.Error(),
			}).Debug("Failed to play healing sound")
		}
	}
}

// findNearestInjuredAlly finds the nearest ally that needs healing.
func (s *SpellCastingSystem) findNearestInjuredAlly(caster *Entity, maxRange float64) *Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"max_range": maxRange,
		}).Debug("Searching for nearest injured ally")
	}

	casterTeamID := s.getCasterTeamID(caster)
	entities := s.world.GetEntities()
	nearestAlly, minDist := s.searchNearestInjuredAlly(caster, entities, casterTeamID, maxRange)

	s.logSearchResult(caster, nearestAlly, minDist)
	return nearestAlly
}

// getCasterTeamID extracts the team ID from the caster entity.
func (s *SpellCastingSystem) getCasterTeamID(caster *Entity) int {
	var casterTeamID int
	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
		if team, ok := teamComp.(*TeamComponent); ok {
			casterTeamID = team.TeamID
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"caster_id": caster.ID,
					"team_id":   casterTeamID,
				}).Debug("Retrieved caster team ID")
			}
		}
	}
	return casterTeamID
}

// searchNearestInjuredAlly searches for the nearest injured ally within range.
func (s *SpellCastingSystem) searchNearestInjuredAlly(caster *Entity, entities []*Entity, casterTeamID int, maxRange float64) (*Entity, float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"team_id":      casterTeamID,
			"max_range":    maxRange,
			"entity_count": len(entities),
		}).Debug("Searching nearest injured ally")
	}

	var nearestAlly *Entity
	minDist := maxRange

	for _, entity := range entities {
		if entity == caster {
			continue
		}

		if !s.isValidInjuredAlly(entity, casterTeamID) {
			continue
		}

		dist := GetDistance(caster, entity)
		if dist <= minDist {
			nearestAlly = entity
			minDist = dist
		}
	}

	return nearestAlly, minDist
}

// isValidInjuredAlly checks if entity is an ally and injured.
func (s *SpellCastingSystem) isValidInjuredAlly(entity *Entity, casterTeamID int) bool {
	teamComp, hasTeam := entity.GetComponent("team")
	if !hasTeam {
		return false
	}

	team, ok := teamComp.(*TeamComponent)
	if !ok || !team.IsAlly(casterTeamID) {
		return false
	}

	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return false
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return false
	}

	isInjured := health.Current < health.Max
	if s.logger != nil && isInjured {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"team_id":        team.TeamID,
			"health_current": health.Current,
			"health_max":     health.Max,
		}).Debug("Found valid injured ally")
	}
	return isInjured
}

// logSearchResult logs the result of the nearest ally search.
func (s *SpellCastingSystem) logSearchResult(caster, nearestAlly *Entity, minDist float64) {
	if s.logger != nil {
		if nearestAlly != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
				"ally_id":   nearestAlly.ID,
				"distance":  minDist,
			}).Debug("Found injured ally")
		} else {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
			}).Debug("No injured ally found in range")
		}
	}
}

// findAlliesInRange finds all allies within range.
func (s *SpellCastingSystem) findAlliesInRange(caster *Entity, maxRange float64) []*Entity {
	s.logAllySearchStart(caster, maxRange)
	casterTeamID := s.getCasterTeamID(caster)
	allies := s.filterAlliesInRange(caster, casterTeamID, maxRange)
	s.logAllySearchResult(caster, allies)
	return allies
}

// logAllySearchStart logs the beginning of ally search.
func (s *SpellCastingSystem) logAllySearchStart(caster *Entity, maxRange float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"max_range": maxRange,
		}).Debug("Searching for allies in range")
	}
}

// filterAlliesInRange filters entities to find allies within range.
func (s *SpellCastingSystem) filterAlliesInRange(caster *Entity, casterTeamID int, maxRange float64) []*Entity {
	entities := s.world.GetEntities()
	var allies []*Entity

	for _, entity := range entities {
		if !s.isAlly(entity, caster, casterTeamID) {
			continue
		}
		if !entity.HasComponent("health") {
			continue
		}
		if GetDistance(caster, entity) <= maxRange {
			allies = append(allies, entity)
		}
	}

	return allies
}

// isAlly checks if an entity is an ally of the caster.
func (s *SpellCastingSystem) isAlly(entity, caster *Entity, casterTeamID int) bool {
	if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
		team, ok := teamComp.(*TeamComponent)
		if !ok || !team.IsAlly(casterTeamID) {
			return false
		}
	} else if entity != caster {
		return false
	}
	return true
}

// logAllySearchResult logs the number of allies found.
func (s *SpellCastingSystem) logAllySearchResult(caster *Entity, allies []*Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"ally_count": len(allies),
		}).Debug("Found allies in range")
	}
}

// castDefensiveSpell applies shields or defensive buffs.
func (s *SpellCastingSystem) castDefensiveSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
		}).Debug("Casting defensive spell")
	}

	// Apply shield using the damage stat as shield strength
	if s.statusEffectSys != nil {
		shieldAmount := float64(spell.Stats.Damage)
		if shieldAmount <= 0 {
			shieldAmount = 50.0 // Default shield if no damage stat
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  caster.ID,
					"spell_name": spell.Name,
				}).Debug("Using default shield amount")
			}
		}

		duration := spell.Stats.Duration
		if duration <= 0 {
			duration = 30.0 // Default duration
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  caster.ID,
					"spell_name": spell.Name,
				}).Debug("Using default shield duration")
			}
		}

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     caster.ID,
				"spell_name":    spell.Name,
				"shield_amount": shieldAmount,
				"duration":      duration,
			}).Info("Shield applied")
		}

		s.statusEffectSys.ApplyShield(caster, shieldAmount, duration)
	}
}

// castBuffSpell applies stat boosts.
func (s *SpellCastingSystem) castBuffSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Casting buff spell")
	}

	if !s.validateBuffSpellCaster() {
		return
	}

	duration := s.determineBuffDuration(caster, spell)
	s.applyElementalBuff(caster, spell, duration)
}

// validateBuffSpellCaster checks if the caster can cast buff spells.
func (s *SpellCastingSystem) validateBuffSpellCaster() bool {
	if s.statusEffectSys == nil {
		if s.logger != nil {
			s.logger.Warn("StatusEffectSystem is nil, cannot apply buff")
		}
		return false
	}
	return true
}

// determineBuffDuration calculates the duration for a buff spell.
func (s *SpellCastingSystem) determineBuffDuration(caster *Entity, spell *magic.Spell) float64 {
	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 30.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Debug("Using default buff duration")
		}
	}
	return duration
}

// applyElementalBuff applies a buff based on spell element.
func (s *SpellCastingSystem) applyElementalBuff(caster *Entity, spell *magic.Spell, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
			"duration":   duration,
		}).Debug("Determining buff type by element")
	}

	switch spell.Element {
	case magic.ElementWind:
		s.statusEffectSys.ApplyStatusEffect(caster, "haste", 0.5, duration, 0)
		s.logBuffApplied(caster.ID, "haste", 0.5, duration)
	case magic.ElementLight:
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.3, duration, 0)
		s.logBuffApplied(caster.ID, "strength", 0.3, duration)
	case magic.ElementEarth:
		s.statusEffectSys.ApplyStatusEffect(caster, "fortify", 0.3, duration, 0)
		s.logBuffApplied(caster.ID, "fortify", 0.3, duration)
	default:
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.2, duration, 0)
		s.logGenericBuffApplied(caster.ID, duration)
	}
}

// logBuffApplied logs when a buff is applied.
func (s *SpellCastingSystem) logBuffApplied(entityID uint64, buffType string, magnitude, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"buff_type": buffType,
			"magnitude": magnitude,
			"duration":  duration,
		}).Info("Buff applied")
	}
}

// logGenericBuffApplied logs when a generic buff is applied.
func (s *SpellCastingSystem) logGenericBuffApplied(entityID uint64, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"buff_type": "strength",
			"magnitude": 0.2,
			"duration":  duration,
		}).Info("Generic buff applied")
	}
}

// castDebuffSpell applies stat reductions to enemies.
func (s *SpellCastingSystem) castDebuffSpell(caster *Entity, spell *magic.Spell, x, y float64) {
	s.logDebuffCast(caster, spell)
	targets := s.findTargets(caster, spell, x, y)
	s.logDebuffTargets(caster, spell, targets)

	for _, target := range targets {
		s.applyDebuffDamage(target, spell)
		s.applyDebuffEffect(target, spell)
	}
}

// logDebuffCast logs the initiation of a debuff spell cast.
func (s *SpellCastingSystem) logDebuffCast(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Casting debuff spell")
	}
}

// logDebuffTargets logs the number of targets found for debuff.
func (s *SpellCastingSystem) logDebuffTargets(caster *Entity, spell *magic.Spell, targets []*Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    caster.ID,
			"spell_name":   spell.Name,
			"target_count": len(targets),
		}).Debug("Targets found for debuff spell")
	}
}

// applyDebuffDamage applies minor damage from debuff spell if any.
func (s *SpellCastingSystem) applyDebuffDamage(target *Entity, spell *magic.Spell) {
	if spell.Stats.Damage <= 0 {
		return
	}

	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	oldHealth := health.Current
	health.Current -= float64(spell.Stats.Damage)
	if health.Current < 0 {
		health.Current = 0
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":     target.ID,
			"damage":        spell.Stats.Damage,
			"health_before": oldHealth,
			"health_after":  health.Current,
		}).Debug("Debuff spell damage applied")
	}
}

// applyDebuffEffect applies debuff status effect based on spell element.
func (s *SpellCastingSystem) applyDebuffEffect(target *Entity, spell *magic.Spell) {
	if s.statusEffectSys == nil {
		return
	}

	duration := s.getDebuffDuration(target, spell)
	effectName, magnitude := s.determineDebuffType(spell.Element)
	s.statusEffectSys.ApplyStatusEffect(target, effectName, magnitude, duration, 0)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"effect":    effectName,
			"magnitude": magnitude,
			"duration":  duration,
		}).Info("Debuff applied")
	}
}

// getDebuffDuration returns the duration for a debuff effect.
func (s *SpellCastingSystem) getDebuffDuration(target *Entity, spell *magic.Spell) float64 {
	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 10.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id":  target.ID,
				"spell_name": spell.Name,
			}).Debug("Using default debuff duration")
		}
	}
	return duration
}

// determineDebuffType returns the debuff type and magnitude based on element.
func (s *SpellCastingSystem) determineDebuffType(element magic.ElementType) (string, float64) {
	switch element {
	case magic.ElementDark:
		return "weakness", 0.7
	case magic.ElementEarth:
		return "vulnerability", 0.7
	default:
		return "weakness", 0.8
	}
}

// applyElementalEffect applies status effects based on spell element.
func (s *SpellCastingSystem) applyElementalEffect(target *Entity, spell *magic.Spell) {
	if s.statusEffectSys == nil {
		if s.logger != nil {
			s.logger.Warn("StatusEffectSystem is nil, cannot apply elemental effects")
		}
		return
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":  target.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Applying elemental effect")
	}

	switch spell.Element {
	case magic.ElementFire:
		s.applyBurningEffect(target)
	case magic.ElementIce:
		s.applyFrozenEffect(target)
	case magic.ElementLightning:
		s.applyLightningEffect(target, spell)
	case magic.ElementEarth:
		s.applyPoisonEffect(target)
	}
}

// applyBurningEffect applies burning status (fire damage over time).
func (s *SpellCastingSystem) applyBurningEffect(target *Entity) {
	s.statusEffectSys.ApplyStatusEffect(target, "burning", 10.0, 3.0, 1.0)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"effect":    "burning",
			"dps":       10.0,
			"duration":  3.0,
		}).Info("Burning effect applied")
	}
}

// applyFrozenEffect applies frozen status (movement slow).
func (s *SpellCastingSystem) applyFrozenEffect(target *Entity) {
	s.statusEffectSys.ApplyStatusEffect(target, "frozen", 0.5, 2.0, 0)
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"effect":    "frozen",
			"slow":      0.5,
			"duration":  2.0,
		}).Info("Frozen effect applied")
	}
}

// applyLightningEffect applies lightning chaining and shocked status.
func (s *SpellCastingSystem) applyLightningEffect(target *Entity, spell *magic.Spell) {
	if spell.Target == magic.TargetSingle || spell.Target == magic.TargetArea {
		s.statusEffectSys.ChainLightning(nil, target, float64(spell.Stats.Damage)*0.5, 2, 15.0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id":    target.ID,
				"chain_damage": float64(spell.Stats.Damage) * 0.5,
				"max_chain":    2,
				"chain_range":  15.0,
			}).Info("Lightning chain initiated")
		}
	}
	s.statusEffectSys.ApplyStatusEffect(target, "shocked", 0, 2.0, 0)
}

// applyPoisonEffect applies poison status if proc check passes.
func (s *SpellCastingSystem) applyPoisonEffect(target *Entity) {
	if s.shouldApplyPoison() {
		s.statusEffectSys.ApplyStatusEffect(target, "poisoned", 5.0, 5.0, 1.0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
				"effect":    "poisoned",
				"dps":       5.0,
				"duration":  5.0,
			}).Info("Poison effect applied")
		}
	}
}

// shouldApplyPoison returns true 30% of the time for Earth spells.
func (s *SpellCastingSystem) shouldApplyPoison() bool {
	// Use status effect system's RNG if available
	if s.statusEffectSys != nil && s.statusEffectSys.rng != nil {
		roll := s.statusEffectSys.rng.Float64()
		applied := roll < 0.3
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"roll":        roll,
				"threshold":   0.3,
				"poison_proc": applied,
			}).Debug("Poison proc check for Earth spell")
		}
		return applied
	}
	return true
}

// castUtilitySpell handles non-combat spells.
func (s *SpellCastingSystem) castUtilitySpell(caster *Entity, spell *magic.Spell) {
	logUtilitySpellCast(s.logger, caster.ID, spell)

	if tryTagBasedUtility(s, caster, spell) {
		return
	}

	castElementBasedUtility(s, caster, spell)
}

// logUtilitySpellCast logs the initial utility spell cast.
func logUtilitySpellCast(logger *logrus.Entry, casterID uint64, spell *magic.Spell) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entity_id":  casterID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
			"tags":       spell.Tags,
		}).Debug("Casting utility spell")
	}
}

// tryTagBasedUtility attempts to cast utility based on spell tags.
func tryTagBasedUtility(s *SpellCastingSystem, caster *Entity, spell *magic.Spell) bool {
	switch {
	case containsTag(spell.Tags, "teleport"):
		logSpecificUtility(s.logger, caster.ID, spell, "teleport")
		s.castTeleportSpell(caster, spell)
		return true
	case containsTag(spell.Tags, "light"), containsTag(spell.Tags, "reveal"):
		logSpecificUtility(s.logger, caster.ID, spell, "reveal")
		s.castRevealSpell(caster, spell)
		return true
	case containsTag(spell.Tags, "speed"), containsTag(spell.Tags, "haste"):
		logSpecificUtility(s.logger, caster.ID, spell, "speed boost")
		s.castSpeedBoostSpell(caster, spell)
		return true
	}
	return false
}

// castElementBasedUtility routes utility spell based on element.
func castElementBasedUtility(s *SpellCastingSystem, caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Using element-based utility spell routing")
	}

	switch spell.Element {
	case magic.ElementLight:
		s.castRevealSpell(caster, spell)
	case magic.ElementWind:
		s.castSpeedBoostSpell(caster, spell)
	case magic.ElementArcane:
		s.castTeleportSpell(caster, spell)
	default:
		logUnmatchedUtility(s.logger, caster.ID, spell)
	}
}

// logSpecificUtility logs a specific utility spell type being cast.
func logSpecificUtility(logger *logrus.Entry, casterID uint64, spell *magic.Spell, utilityType string) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entity_id":  casterID,
			"spell_name": spell.Name,
		}).Debug("Casting " + utilityType + " utility spell")
	}
}

// logUnmatchedUtility logs when no utility implementation matches the spell.
func logUnmatchedUtility(logger *logrus.Entry, casterID uint64, spell *magic.Spell) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entity_id":  casterID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("No matching utility spell implementation for element")
	}
}

// castTeleportSpell teleports the caster to a nearby safe location.
// Teleport distance is based on spell range, and landing spot must be walkable.
// getTeleportDirection calculates the teleport direction from entity velocity or default.
func getTeleportDirection(caster *Entity) (float64, float64) {
	dirX, dirY := 0.0, 1.0
	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			if vel.VX != 0 || vel.VY != 0 {
				mag := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
				if mag > 0 {
					dirX = vel.VX / mag
					dirY = vel.VY / mag
				}
			}
		}
	}
	return dirX, dirY
}

// calculateTeleportTarget calculates the target position for teleportation.
func calculateTeleportTarget(pos *PositionComponent, dirX, dirY, maxDist float64) (float64, float64) {
	targetX := pos.X + dirX*maxDist
	targetY := pos.Y + dirY*maxDist
	return targetX, targetY
}

// executeTeleport performs the actual teleportation if position is valid.
func (s *SpellCastingSystem) executeTeleport(caster *Entity, pos *PositionComponent, targetX, targetY, maxDist float64) bool {
	if !s.isPositionWalkable(targetX, targetY, caster) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"target_x":  targetX,
				"target_y":  targetY,
			}).Warn("Teleport failed: position not walkable")
		}
		if s.tutorialSys != nil {
			s.tutorialSys.ShowNotification("Cannot teleport there!", 1.5)
		}
		return false
	}

	previousX, previousY := pos.X, pos.Y
	pos.X = targetX
	pos.Y = targetY

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": caster.ID,
			"from_x":    previousX,
			"from_y":    previousY,
			"to_x":      targetX,
			"to_y":      targetY,
			"distance":  maxDist,
		}).Info("Entity position teleported")
	}
	return true
}

// spawnTeleportEffects spawns visual and audio effects for teleportation.
func (s *SpellCastingSystem) spawnTeleportEffects(caster *Entity, x, y float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"position_x": x,
			"position_y": y,
		}).Debug("Spawning teleport effects")
	}

	if s.particleSys != nil {
		config := particles.Config{
			Type:     particles.ParticleMagic,
			Count:    30,
			GenreID:  "fantasy",
			Seed:     int64(caster.ID),
			Duration: 0.5,
			SpreadX:  80.0,
			SpreadY:  80.0,
			Gravity:  0.0,
			MinSize:  4.0,
			MaxSize:  8.0,
			Custom:   map[string]interface{}{"color": "teleport"},
		}
		s.particleSys.SpawnParticles(s.world, config, x, y)
	}

	if s.audioMgr != nil {
		if err := s.audioMgr.PlaySFX("magic", int64(caster.ID)); err != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": caster.ID,
					"error":     err.Error(),
				}).Debug("Failed to play teleport sound")
			}
		}
	}
}

// castTeleportSpell casts a teleport spell for the specified entity.
func (s *SpellCastingSystem) castTeleportSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
		}).Debug("Casting teleport spell")
	}

	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
			}).Warn("Caster has no position component for teleport")
		}
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      caster.ID,
				"component_type": "position",
			}).Warn("Failed to cast position component for teleport")
		}
		return
	}

	maxDist := spell.Stats.Range
	if maxDist <= 0 {
		maxDist = 100.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Debug("Using default teleport distance")
		}
	}

	dirX, dirY := getTeleportDirection(caster)
	targetX, targetY := calculateTeleportTarget(pos, dirX, dirY, maxDist)

	if s.executeTeleport(caster, pos, targetX, targetY, maxDist) {
		s.spawnTeleportEffects(caster, pos.X, pos.Y)
	}
}

// castRevealSpell reveals fog of war in an area around the caster.
// Useful for exploration and finding hidden enemies/items.
func (s *SpellCastingSystem) castRevealSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
		}).Debug("Casting reveal spell")
	}

	pos, ok := s.validateRevealSpellPosition(caster)
	if !ok {
		return
	}

	revealRadius := s.calculateRevealRadius(caster, spell)
	s.applyRevealEffects(caster, pos, revealRadius)
}

// validateRevealSpellPosition validates the caster has a valid position.
func (s *SpellCastingSystem) validateRevealSpellPosition(caster *Entity) (*PositionComponent, bool) {
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
			}).Warn("Caster has no position component for reveal spell")
		}
		return nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, false
	}
	return pos, true
}

// calculateRevealRadius determines the reveal radius from spell stats.
func (s *SpellCastingSystem) calculateRevealRadius(caster *Entity, spell *magic.Spell) float64 {
	revealRadius := spell.Stats.AreaSize
	if revealRadius <= 0 {
		revealRadius = 200.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Debug("Using default reveal radius")
		}
	}
	return revealRadius
}

// applyRevealEffects applies status effect and spawns visual/audio effects.
func (s *SpellCastingSystem) applyRevealEffects(caster *Entity, pos *PositionComponent, revealRadius float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     caster.ID,
			"reveal_radius": revealRadius,
			"position_x":    pos.X,
			"position_y":    pos.Y,
		}).Info("Reveal spell effect applied")
	}

	if s.statusEffectSys != nil {
		duration := 0.1
		s.statusEffectSys.ApplyStatusEffect(caster, "revealing", revealRadius, duration, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"radius":    revealRadius,
				"duration":  duration,
			}).Debug("Revealing status effect applied")
		}
	}

	s.spawnRevealParticles(caster, pos, revealRadius)
	s.playRevealSound(caster)
}

// spawnRevealParticles spawns light particles for reveal effect.
func (s *SpellCastingSystem) spawnRevealParticles(caster *Entity, pos *PositionComponent, revealRadius float64) {
	if s.particleSys != nil {
		config := particles.Config{
			Type:     particles.ParticleSpark,
			Count:    40,
			GenreID:  "fantasy",
			Seed:     int64(caster.ID),
			Duration: 1.5,
			SpreadX:  revealRadius,
			SpreadY:  revealRadius,
			Gravity:  -20.0,
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   map[string]interface{}{"color": "light"},
		}
		s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
	}
}

// playRevealSound plays the reveal sound effect.
func (s *SpellCastingSystem) playRevealSound(caster *Entity) {
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(caster.ID))
	}
}

// castSpeedBoostSpell applies a temporary speed boost to the caster.
// Increases movement speed for exploration or combat mobility.
func (s *SpellCastingSystem) castSpeedBoostSpell(caster *Entity, spell *magic.Spell) {
	s.logSpeedBoostCast(caster, spell)

	if s.statusEffectSys == nil {
		s.logMissingStatusSystem()
		return
	}

	duration := s.getSpeedBoostDuration(caster, spell)
	speedMultiplier := s.calculateSpeedMultiplier(caster, spell)
	s.logSpeedBoostApplication(caster, spell, speedMultiplier, duration)

	s.statusEffectSys.ApplyStatusEffect(caster, "speed_boost", speedMultiplier, duration, 0)
	s.spawnSpeedBoostParticles(caster, duration)
	s.playSpeedBoostSound(caster)
}

// logSpeedBoostCast logs the initiation of speed boost spell.
func (s *SpellCastingSystem) logSpeedBoostCast(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
		}).Debug("Casting speed boost spell")
	}
}

// logMissingStatusSystem logs warning when status effect system is missing.
func (s *SpellCastingSystem) logMissingStatusSystem() {
	if s.logger != nil {
		s.logger.Warn("StatusEffectSystem is nil, cannot apply speed boost")
	}
}

// getSpeedBoostDuration returns the duration for speed boost effect.
func (s *SpellCastingSystem) getSpeedBoostDuration(caster *Entity, spell *magic.Spell) float64 {
	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 10.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Debug("Using default speed boost duration")
		}
	}
	return duration
}

// calculateSpeedMultiplier determines the speed multiplier based on spell rarity.
func (s *SpellCastingSystem) calculateSpeedMultiplier(caster *Entity, spell *magic.Spell) float64 {
	speedMultiplier := 1.5
	if spell.Rarity >= magic.RarityRare {
		speedMultiplier = 2.0
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
				"rarity":     spell.Rarity.String(),
			}).Debug("Using increased speed multiplier for rare spell")
		}
	}
	return speedMultiplier
}

// logSpeedBoostApplication logs successful speed boost application.
func (s *SpellCastingSystem) logSpeedBoostApplication(caster *Entity, spell *magic.Spell, speedMultiplier, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        caster.ID,
			"speed_multiplier": speedMultiplier,
			"duration":         duration,
			"rarity":           spell.Rarity.String(),
		}).Info("Speed boost applied")
	}
}

// spawnSpeedBoostParticles creates visual speed boost particle effects.
func (s *SpellCastingSystem) spawnSpeedBoostParticles(caster *Entity, duration float64) {
	if s.particleSys == nil {
		return
	}

	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	config := particles.Config{
		Type:     particles.ParticleDust,
		Count:    25,
		GenreID:  "fantasy",
		Seed:     int64(caster.ID),
		Duration: duration,
		SpreadX:  100.0,
		SpreadY:  60.0,
		Gravity:  10.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   map[string]interface{}{"color": "wind"},
	}
	s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"position_x": pos.X,
			"position_y": pos.Y,
		}).Debug("Spawned speed boost particles")
	}
}

// playSpeedBoostSound plays the speed boost sound effect.
func (s *SpellCastingSystem) playSpeedBoostSound(caster *Entity) {
	if s.audioMgr == nil {
		return
	}

	if err := s.audioMgr.PlaySFX("powerup", int64(caster.ID)); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"error":     err.Error(),
			}).Debug("Failed to play speed boost sound")
		}
	}
}

// isPositionWalkable checks if a position is valid for teleportation.
// Returns true if the position has no solid colliders.
func (s *SpellCastingSystem) isPositionWalkable(x, y float64, caster *Entity) bool {
	s.logWalkableCheck(caster.ID, x, y, "Checking if position is walkable")

	for _, entity := range s.world.GetEntities() {
		if entity == caster {
			continue
		}

		collider, pos := s.extractSolidColliderAndPosition(entity)
		if collider == nil || pos == nil {
			continue
		}

		if s.checkCollisionAtPosition(x, y, caster, entity, collider, pos) {
			return false
		}
	}

	s.logWalkableCheck(caster.ID, x, y, "Position is walkable")
	return true
}

// extractSolidColliderAndPosition extracts solid collider and position from entity.
func (s *SpellCastingSystem) extractSolidColliderAndPosition(entity *Entity) (*ColliderComponent, *PositionComponent) {
	colliderComp, hasCollider := entity.GetComponent("collider")
	if !hasCollider {
		return nil, nil
	}
	collider, ok := colliderComp.(*ColliderComponent)
	if !ok || !collider.Solid {
		return nil, nil
	}

	entityPos, hasPos := entity.GetComponent("position")
	if !hasPos {
		return nil, nil
	}
	pos, ok := entityPos.(*PositionComponent)
	if !ok {
		return nil, nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"solid":     collider.Solid,
		}).Debug("Extracted solid collider and position")
	}
	return collider, pos
}

// checkCollisionAtPosition checks if position collides with entity.
func (s *SpellCastingSystem) checkCollisionAtPosition(x, y float64, caster, entity *Entity, collider *ColliderComponent, pos *PositionComponent) bool {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"entity_id": entity.ID,
			"target_x":  x,
			"target_y":  y,
		}).Debug("Checking collision at position")
	}

	casterCollider, hasCasterCollider := caster.GetComponent("collider")
	if hasCasterCollider {
		return s.checkBoxCollision(x, y, caster.ID, entity.ID, casterCollider, collider, pos)
	}
	return s.checkPointCollision(x, y, caster.ID, entity.ID, collider, pos)
}

// checkBoxCollision checks box-to-box collision using caster's collider bounds.
func (s *SpellCastingSystem) checkBoxCollision(x, y float64, casterID, entityID uint64, casterCollider Component, collider *ColliderComponent, pos *PositionComponent) bool {
	cc, ok := casterCollider.(*ColliderComponent)
	if !ok {
		return false
	}

	tempCollider := &ColliderComponent{
		Width:   cc.Width,
		Height:  cc.Height,
		OffsetX: cc.OffsetX,
		OffsetY: cc.OffsetY,
		Solid:   true,
		Layer:   cc.Layer,
	}

	if tempCollider.Intersects(x, y, collider, pos.X, pos.Y) {
		s.logCollisionDetected(casterID, entityID, x, y, "collision detected")
		return true
	}
	return false
}

// checkPointCollision checks point-to-box collision.
func (s *SpellCastingSystem) checkPointCollision(x, y float64, casterID, entityID uint64, collider *ColliderComponent, pos *PositionComponent) bool {
	minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)
	if x >= minX && x <= maxX && y >= minY && y <= maxY {
		s.logCollisionDetected(casterID, entityID, x, y, "point collision")
		return true
	}
	return false
}

// logWalkableCheck logs walkable position checks.
func (s *SpellCastingSystem) logWalkableCheck(casterID uint64, x, y float64, message string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": casterID,
			"target_x":  x,
			"target_y":  y,
		}).Debug(message)
	}
}

// logCollisionDetected logs collision detection events.
func (s *SpellCastingSystem) logCollisionDetected(casterID, entityID uint64, x, y float64, reason string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":       casterID,
			"blocking_entity": entityID,
			"target_x":        x,
			"target_y":        y,
		}).Debug("Position not walkable: " + reason)
	}
}

// containsTag checks if a spell has a specific tag.
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// findTargets returns entities affected by the spell.
func (s *SpellCastingSystem) findTargets(caster *Entity, spell *magic.Spell, x, y float64) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":   caster.ID,
			"spell_name":  spell.Name,
			"target_type": spell.Target.String(),
		}).Debug("Finding targets for spell")
	}

	var targets []*Entity
	switch spell.Target {
	case magic.TargetSelf:
		targets = []*Entity{caster}
	case magic.TargetSingle:
		targets = s.findNearestEnemyTarget(caster, spell.Stats.Range)
	case magic.TargetArea:
		targets = s.findAreaTargets(caster, spell.Stats.AreaSize)
	case magic.TargetAllEnemies:
		targets = s.findAllEnemyTargets(caster)
	case magic.TargetCone:
		targets = s.findConeTargets(caster, spell, x, y)
	case magic.TargetLine:
		targets = s.findLineTargets(caster, spell, x, y)
	default:
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id":   caster.ID,
				"target_type": spell.Target.String(),
			}).Warn("Unknown target type for spell")
		}
		return nil
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"spell_name":   spell.Name,
			"target_count": len(targets),
		}).Debug("Targets found")
	}

	return targets
}

// isEnemyTarget checks if an entity is a valid enemy target for spell effects.
func (s *SpellCastingSystem) isEnemyTarget(caster, entity *Entity) bool {
	if entity == caster {
		return false
	}
	// Player has input component
	if entity.HasComponent("input") {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity rejected as target: is player")
		}
		return false
	}
	// Must have health to be a valid target
	if !entity.HasComponent("health") {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity rejected as target: no health component")
		}
		return false
	}
	return true
}

// findNearestEnemyTarget finds the nearest enemy within range.
func (s *SpellCastingSystem) findNearestEnemyTarget(caster *Entity, maxRange float64) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"max_range": maxRange,
		}).Debug("Searching for nearest enemy target")
	}

	entities := s.world.GetEntities()
	var nearest *Entity
	nearestDist := maxRange

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		dist := GetDistance(caster, entity)
		if dist <= nearestDist {
			nearest = entity
			nearestDist = dist
		}
	}

	if nearest != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
				"target_id": nearest.ID,
				"distance":  nearestDist,
			}).Debug("Found nearest enemy target")
		}
		return []*Entity{nearest}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
		}).Debug("No enemy targets found in range")
	}
	return nil
}

// findAreaTargets finds all enemies within area range.
func (s *SpellCastingSystem) findAreaTargets(caster *Entity, areaSize float64) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"area_size": areaSize,
		}).Debug("Searching for enemies in area")
	}

	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		dist := GetDistance(caster, entity)
		if dist <= areaSize {
			targets = append(targets, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found enemies in area")
	}

	return targets
}

// findAllEnemyTargets finds all enemies regardless of distance.
func (s *SpellCastingSystem) findAllEnemyTargets(caster *Entity) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
		}).Debug("Searching for all enemy targets")
	}

	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}
		targets = append(targets, entity)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found all enemy targets")
	}

	return targets
}

// findConeTargets finds entities within a cone-shaped area from the caster.
func (s *SpellCastingSystem) findConeTargets(caster *Entity, spell *magic.Spell, x, y float64) []*Entity {
	s.logConeTargetSearch(caster, spell)

	casterPosComp := s.getCasterPositionForCone(caster)
	if casterPosComp == nil {
		return nil
	}

	dirX, dirY := s.getNormalizedConeDirection(caster, x, y)
	targets := s.filterTargetsInCone(caster, spell, casterPosComp, dirX, dirY)

	s.logConeTargetResults(caster, targets)
	return targets
}

// logConeTargetSearch logs the start of cone target search.
func (s *SpellCastingSystem) logConeTargetSearch(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"spell_name": spell.Name,
			"range":      spell.Stats.Range,
		}).Debug("Searching for enemies in cone")
	}
}

// getCasterPositionForCone retrieves and validates caster position component for cone targeting.
func (s *SpellCastingSystem) getCasterPositionForCone(caster *Entity) *PositionComponent {
	casterPos, hasCasterPos := caster.GetComponent("position")
	if !hasCasterPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
			}).Warn("Caster has no position component for cone targeting")
		}
		return nil
	}

	casterPosComp, ok := casterPos.(*PositionComponent)
	if !ok {
		return nil
	}

	return casterPosComp
}

// getNormalizedConeDirection calculates and normalizes the direction vector for cone attacks.
func (s *SpellCastingSystem) getNormalizedConeDirection(caster *Entity, x, y float64) (float64, float64) {
	dirX, dirY := s.getCasterDirection(caster, x, y)
	if dirX == 0 && dirY == 0 {
		dirX = 1.0
	}

	dirLength := math.Sqrt(dirX*dirX + dirY*dirY)
	return dirX / dirLength, dirY / dirLength
}

// filterTargetsInCone finds all valid targets within the cone area.
func (s *SpellCastingSystem) filterTargetsInCone(caster *Entity, spell *magic.Spell, casterPos *PositionComponent, dirX, dirY float64) []*Entity {
	coneAngle := 45.0 * math.Pi / 180.0
	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		entityPosComp := s.getConeEntityPosition(entity)
		if entityPosComp == nil {
			continue
		}

		if s.isInCone(casterPos, entityPosComp, dirX, dirY, spell.Stats.Range, coneAngle) {
			targets = append(targets, entity)
		}
	}

	return targets
}

// getConeEntityPosition retrieves entity position component for cone targeting.
func (s *SpellCastingSystem) getConeEntityPosition(entity *Entity) *PositionComponent {
	entityPos, hasPos := entity.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity has no position component for cone targeting")
		}
		return nil
	}

	entityPosComp, ok := entityPos.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "position",
			}).Warn("Failed to cast position component for cone targeting")
		}
		return nil
	}

	return entityPosComp
}

// isInCone checks if an entity position is within the cone area.
func (s *SpellCastingSystem) isInCone(casterPos, entityPos *PositionComponent, dirX, dirY, maxRange, coneAngle float64) bool {
	toEntityX := entityPos.X - casterPos.X
	toEntityY := entityPos.Y - casterPos.Y
	dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

	if dist > maxRange || dist < 0.1 {
		return false
	}

	toEntityX /= dist
	toEntityY /= dist

	dotProduct := dirX*toEntityX + dirY*toEntityY
	angle := math.Acos(math.Max(-1.0, math.Min(1.0, dotProduct)))

	inCone := angle <= coneAngle
	if s.logger != nil && inCone {
		s.logger.WithFields(logrus.Fields{
			"distance":   dist,
			"angle":      angle,
			"cone_angle": coneAngle,
		}).Debug("Entity is within cone area")
	}
	return inCone
}

// logConeTargetResults logs the cone target search results.
func (s *SpellCastingSystem) logConeTargetResults(caster *Entity, targets []*Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found enemies in cone")
	}
}

// findLineTargets finds entities along a line from the caster in the facing direction.
func (s *SpellCastingSystem) findLineTargets(caster *Entity, spell *magic.Spell, x, y float64) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"spell_name": spell.Name,
			"range":      spell.Stats.Range,
		}).Debug("Searching for enemies in line")
	}

	casterPos, ok := s.validateCasterPosition(caster)
	if !ok {
		return nil
	}

	dirX, dirY := s.getNormalizedConeDirection(caster, x, y)
	lineWidth := 32.0
	targets := s.filterEntitiesInLine(caster, casterPos, dirX, dirY, spell.Stats.Range, lineWidth)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found enemies in line")
	}

	return targets
}

// validateCasterPosition retrieves and validates the caster's position component.
func (s *SpellCastingSystem) validateCasterPosition(caster *Entity) (*PositionComponent, bool) {
	casterPos, hasCasterPos := caster.GetComponent("position")
	if !hasCasterPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
			}).Warn("Caster has no position component for line targeting")
		}
		return nil, false
	}
	casterPosComp, ok := casterPos.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id":      caster.ID,
				"component_type": "position",
			}).Warn("Failed to cast position component for line targeting")
		}
		return nil, false
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"position_x": casterPosComp.X,
			"position_y": casterPosComp.Y,
		}).Debug("Caster position validated for line targeting")
	}
	return casterPosComp, true
}

// filterEntitiesInLine finds all entities within the line attack area.
func (s *SpellCastingSystem) filterEntitiesInLine(caster *Entity, casterPos *PositionComponent, dirX, dirY, maxRange, lineWidth float64) []*Entity {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"max_range":  maxRange,
			"line_width": lineWidth,
			"dir_x":      dirX,
			"dir_y":      dirY,
		}).Debug("Filtering entities in line area")
	}

	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		entityPosComp := s.getConeEntityPosition(entity)
		if entityPosComp == nil {
			continue
		}

		if s.isEntityInLineArea(casterPos, entityPosComp, dirX, dirY, maxRange, lineWidth) {
			targets = append(targets, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Completed line area filtering")
	}
	return targets
}

// isEntityInLineArea checks if an entity is within the line attack boundaries.
func (s *SpellCastingSystem) isEntityInLineArea(casterPos, entityPos *PositionComponent, dirX, dirY, maxRange, lineWidth float64) bool {
	toEntityX := entityPos.X - casterPos.X
	toEntityY := entityPos.Y - casterPos.Y
	dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

	if dist > maxRange || dist < 0.1 {
		return false
	}

	projection := toEntityX*dirX + toEntityY*dirY
	if projection < 0 {
		return false
	}

	perpX := toEntityX - projection*dirX
	perpY := toEntityY - projection*dirY
	perpDist := math.Sqrt(perpX*perpX + perpY*perpY)

	inLine := perpDist <= lineWidth
	if s.logger != nil && inLine {
		s.logger.WithFields(logrus.Fields{
			"distance":         dist,
			"projection":       projection,
			"perpendicular":    perpDist,
			"line_width_limit": lineWidth,
		}).Debug("Entity is within line area")
	}
	return inLine
}

// getCasterDirection determines the caster's facing direction for directional spells.
// Uses velocity if moving, otherwise uses direction towards target point (x, y).
func (s *SpellCastingSystem) getCasterDirection(caster *Entity, targetX, targetY float64) (dirX, dirY float64) {
	if dirX, dirY, ok := s.getDirectionFromVelocity(caster); ok {
		return dirX, dirY
	}

	if dirX, dirY, ok := s.getDirectionFromTarget(caster, targetX, targetY); ok {
		return dirX, dirY
	}

	return s.getDefaultDirection(caster)
}

// getDirectionFromVelocity attempts to get direction from entity velocity.
func (s *SpellCastingSystem) getDirectionFromVelocity(caster *Entity) (dirX, dirY float64, ok bool) {
	velComp, hasVel := caster.GetComponent("velocity")
	if !hasVel {
		return 0, 0, false
	}

	vel, ok := velComp.(*VelocityComponent)
	if !ok || (vel.VX == 0 && vel.VY == 0) {
		return 0, 0, false
	}

	s.logDirection(caster.ID, vel.VX, vel.VY, "velocity")
	return vel.VX, vel.VY, true
}

// getDirectionFromTarget calculates direction towards target point.
func (s *SpellCastingSystem) getDirectionFromTarget(caster *Entity, targetX, targetY float64) (dirX, dirY float64, ok bool) {
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		return 0, 0, false
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return 0, 0, false
	}

	dirX = targetX - pos.X
	dirY = targetY - pos.Y
	s.logDirection(caster.ID, dirX, dirY, "target_point")
	return dirX, dirY, true
}

// getDefaultDirection returns default direction when no other method works.
func (s *SpellCastingSystem) getDefaultDirection(caster *Entity) (dirX, dirY float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
		}).Debug("Using default direction (right)")
	}
	return 1.0, 0.0
}

// logDirection logs the computed spell direction.
func (s *SpellCastingSystem) logDirection(casterID uint64, dirX, dirY float64, source string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": casterID,
			"dir_x":     dirX,
			"dir_y":     dirY,
			"source":    source,
		}).Debug("Using velocity for spell direction")
	}
}

// StartCast initiates casting a spell from a slot.
func (s *SpellCastingSystem) StartCast(entity *Entity, slotIndex int) bool {
	s.logCastAttempt(entity.ID, slotIndex)

	slots := s.getSpellSlots(entity)
	if slots == nil {
		return false
	}

	if !s.validateCastPreconditions(entity, slots, slotIndex) {
		return false
	}

	spell := slots.GetSlot(slotIndex)
	if !s.checkManaAvailability(entity, spell) {
		return false
	}

	s.initiateCast(entity, slots, slotIndex, spell)
	return true
}

// logCastAttempt logs spell cast attempt.
func (s *SpellCastingSystem) logCastAttempt(entityID uint64, slotIndex int) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"slot_index": slotIndex,
		}).Debug("Attempting to start spell cast")
	}
}

// getSpellSlots retrieves spell slots component from entity.
func (s *SpellCastingSystem) getSpellSlots(entity *Entity) *SpellSlotComponent {
	spellComp, hasSpells := entity.GetComponent("spell_slots")
	if !hasSpells {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity has no spell_slots component")
		}
		return nil
	}
	slots, ok := spellComp.(*SpellSlotComponent)
	if !ok {
		return nil
	}
	return slots
}

// validateCastPreconditions checks if entity can cast (not already casting, valid slot, not on cooldown).
func (s *SpellCastingSystem) validateCastPreconditions(entity *Entity, slots *SpellSlotComponent, slotIndex int) bool {
	if slots.IsCasting() {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity is already casting a spell")
		}
		return false
	}

	spell := slots.GetSlot(slotIndex)
	if spell == nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"slot_index": slotIndex,
			}).Debug("Spell slot is empty")
		}
		return false
	}

	if slots.IsOnCooldown(slotIndex) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":    entity.ID,
				"slot_index":   slotIndex,
				"spell_name":   spell.Name,
				"remaining_cd": slots.Cooldowns[slotIndex],
			}).Debug("Spell is on cooldown")
		}
		return false
	}
	return true
}

// checkManaAvailability verifies entity has sufficient mana for spell.
func (s *SpellCastingSystem) checkManaAvailability(entity *Entity, spell *magic.Spell) bool {
	manaComp, hasMana := entity.GetComponent("mana")
	if !hasMana {
		return false
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return false
	}

	if mana.Current < spell.Stats.ManaCost {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"spell_name":    spell.Name,
				"mana_current":  mana.Current,
				"mana_required": spell.Stats.ManaCost,
			}).Debug("Insufficient mana to start cast")
		}
		return false
	}
	return true
}

// initiateCast starts the spell casting process.
func (s *SpellCastingSystem) initiateCast(entity *Entity, slots *SpellSlotComponent, slotIndex int, spell *magic.Spell) {
	slots.Casting = slotIndex
	slots.CastingBar = 0

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"slot_index": slotIndex,
			"spell_name": spell.Name,
			"cast_time":  spell.Stats.CastTime,
		}).Info("Spell cast started")
	}
}

// CancelCast interrupts current spell cast.
func (s *SpellCastingSystem) CancelCast(entity *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("Attempting to cancel spell cast")
	}

	spellComp, hasSpells := entity.GetComponent("spell_slots")
	if !hasSpells {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("Entity has no spell_slots component to cancel")
		}
		return
	}
	slots, ok := spellComp.(*SpellSlotComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "spell_slots",
			}).Warn("Failed to cast spell_slots component for cancel")
		}
		return
	}

	if slots.IsCasting() {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     entity.ID,
				"slot_index":    slots.Casting,
				"cast_progress": slots.CastingBar,
			}).Info("Spell cast canceled")
		}
		slots.Casting = -1
		slots.CastingBar = 0
	} else {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Debug("No active cast to cancel")
		}
	}
}

// spawnElementalHitEffect creates element-specific particle effects for spell hits.
func (s *SpellCastingSystem) spawnElementalHitEffect(x, y float64, element magic.ElementType, seed uint64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"element":    element.String(),
			"position_x": x,
			"position_y": y,
		}).Debug("Spawning elemental hit effect")
	}

	if s.particleSys == nil {
		if s.logger != nil {
			s.logger.Debug("ParticleSystem is nil, skipping hit effect")
		}
		return
	}

	var config particles.Config
	switch element {
	case magic.ElementFire:
		// Fire: Orange/red flames rising upward
		config = particles.Config{
			Type:     particles.ParticleFlame,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.8,
			SpreadX:  80.0,
			SpreadY:  80.0,
			Gravity:  -100.0, // Rise upward
			MinSize:  3.0,
			MaxSize:  7.0,
			Custom:   make(map[string]interface{}),
		}

	case magic.ElementIce:
		// Ice: Blue/white crystals with slower movement
		config = particles.Config{
			Type:     particles.ParticleMagic, // Use magic particles with blue tint
			Count:    15,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.2,
			SpreadX:  60.0,
			SpreadY:  60.0,
			Gravity:  50.0, // Slow fall
			MinSize:  4.0,
			MaxSize:  8.0,
			Custom:   map[string]interface{}{"color": "ice"},
		}

	case magic.ElementLightning:
		// Lightning: Fast yellow/white sparks
		config = particles.Config{
			Type:     particles.ParticleSpark,
			Count:    25,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.4,
			SpreadX:  120.0,
			SpreadY:  120.0,
			Gravity:  0.0, // No gravity, pure energy
			MinSize:  2.0,
			MaxSize:  5.0,
			Custom:   map[string]interface{}{"color": "lightning"},
		}

	case magic.ElementEarth:
		// Earth: Brown/green dust/rock particles (can apply poison)
		config = particles.Config{
			Type:     particles.ParticleDust,
			Count:    18,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.2,
			SpreadX:  70.0,
			SpreadY:  70.0,
			Gravity:  100.0, // Fall to ground
			MinSize:  3.0,
			MaxSize:  7.0,
			Custom:   map[string]interface{}{"color": "earth"},
		}

	case magic.ElementWind:
		// Wind: Fast-moving dust particles
		config = particles.Config{
			Type:     particles.ParticleDust,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.6,
			SpreadX:  150.0,
			SpreadY:  80.0,
			Gravity:  20.0,
			MinSize:  2.0,
			MaxSize:  4.0,
			Custom:   map[string]interface{}{"color": "wind"},
		}

	case magic.ElementLight:
		// Light: Bright white/yellow particles
		config = particles.Config{
			Type:     particles.ParticleSpark,
			Count:    22,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.0,
			SpreadX:  100.0,
			SpreadY:  100.0,
			Gravity:  -30.0, // Slow rise
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   map[string]interface{}{"color": "light"},
		}

	case magic.ElementDark:
		// Dark: Purple/black smoke particles
		config = particles.Config{
			Type:     particles.ParticleSmoke,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.5,
			SpreadX:  90.0,
			SpreadY:  90.0,
			Gravity:  -10.0, // Slow rise
			MinSize:  4.0,
			MaxSize:  8.0,
			Custom:   map[string]interface{}{"color": "dark"},
		}

	default:
		// Generic magic effect for other elements (none, arcane)
		config = particles.Config{
			Type:     particles.ParticleMagic,
			Count:    15,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.8,
			SpreadX:  90.0,
			SpreadY:  90.0,
			Gravity:  -50.0,
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   make(map[string]interface{}),
		}
	}

	s.particleSys.SpawnParticles(s.world, config, x, y)
}

// ManaRegenSystem regenerates mana over time.
type ManaRegenSystem struct {
	logger *logrus.Entry
}

// NewManaRegenSystem creates a new mana regeneration system.
func NewManaRegenSystem() *ManaRegenSystem {
	return &ManaRegenSystem{
		logger: logrus.WithField("system", "mana_regen"),
	}
}

// Update regenerates mana for all entities.
func (s *ManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("ManaRegenSystem update started")
	}

	regenCount, totalRegenerated := s.regenerateManaForEntities(entities, deltaTime)
	s.logRegenerationStats(regenCount, totalRegenerated, deltaTime)
}

// regenerateManaForEntities processes mana regeneration for all entities with mana.
func (s *ManaRegenSystem) regenerateManaForEntities(entities []*Entity, deltaTime float64) (int, float64) {
	regenCount := 0
	totalRegenerated := 0.0

	for _, entity := range entities {
		mana := s.getManaComponent(entity)
		if mana == nil {
			continue
		}

		regenerated := s.regenerateMana(mana, deltaTime)
		if regenerated > 0 {
			regenCount++
			totalRegenerated += regenerated
		}
	}

	return regenCount, totalRegenerated
}

// getManaComponent retrieves and validates the mana component for an entity.
func (s *ManaRegenSystem) getManaComponent(entity *Entity) *ManaComponent {
	manaComp, hasMana := entity.GetComponent("mana")
	if !hasMana {
		return nil
	}

	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
			}).Warn("Failed to cast mana component in mana regen system")
		}
		return nil
	}

	return mana
}

// regenerateMana applies mana regeneration to a component and returns amount regenerated.
func (s *ManaRegenSystem) regenerateMana(mana *ManaComponent, deltaTime float64) float64 {
	if mana.Current >= mana.Max {
		return 0
	}

	oldMana := mana.Current
	regenAmount := mana.Regen * deltaTime
	mana.Current += int(regenAmount)
	if mana.Current > mana.Max {
		mana.Current = mana.Max
	}

	return float64(mana.Current - oldMana)
}

// logRegenerationStats logs mana regeneration statistics if any regeneration occurred.
func (s *ManaRegenSystem) logRegenerationStats(regenCount int, totalRegenerated, deltaTime float64) {
	if s.logger != nil && regenCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_count":      regenCount,
			"total_regenerated": totalRegenerated,
			"delta_time":        deltaTime,
		}).Debug("Mana regenerated for entities")
	}
}

// LoadPlayerSpells generates and equips spells for the player.
func LoadPlayerSpells(player *Entity, seed int64, genreID string, depth int) error {
	log := logrus.WithFields(logrus.Fields{
		"entity_id": player.ID,
		"seed":      seed,
		"genre_id":  genreID,
		"depth":     depth,
	})
	log.Info("Loading player spells initiated")

	// Generate spells using procgen system
	generator := magic.NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 5, // Generate 5 spells for the 5 slots
		},
	}

	log.WithFields(logrus.Fields{
		"difficulty": params.Difficulty,
		"count":      5,
	}).Debug("Generating spells with parameters")

	result, err := generator.Generate(seed, params)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Spell generation failed")
		return err
	}

	spells := result.([]*magic.Spell)
	log.WithFields(logrus.Fields{
		"spell_count": len(spells),
	}).Debug("Spells generated successfully")

	// Create spell slots component if doesn't exist
	var slots *SpellSlotComponent
	componentCreated := false
	if !player.HasComponent("spell_slots") {
		slots = &SpellSlotComponent{
			Casting: -1,
		}
		player.AddComponent(slots)
		componentCreated = true
		log.WithFields(logrus.Fields{
			"component_type": "spell_slots",
		}).Debug("Created new spell_slots component")
	} else {
		slotsComp, _ := player.GetComponent("spell_slots")
		slots = slotsComp.(*SpellSlotComponent)
		log.WithFields(logrus.Fields{
			"component_type": "spell_slots",
		}).Debug("Using existing spell_slots component")
	}

	// Equip spells to slots
	equippedCount := 0
	for i := 0; i < 5 && i < len(spells); i++ {
		slots.SetSlot(i, spells[i])
		log.WithFields(logrus.Fields{
			"slot_index": i,
			"spell_name": spells[i].Name,
			"spell_type": spells[i].Type.String(),
			"element":    spells[i].Element.String(),
			"rarity":     spells[i].Rarity.String(),
			"mana_cost":  spells[i].Stats.ManaCost,
		}).Debug("Spell equipped to slot")
		equippedCount++
	}

	// Populate SpellComponent so prestige/new-game-plus carryover can persist
	// the player's learned spells across runs.
	var spellComp *SpellComponent
	if comp, ok := player.GetComponent("spell"); ok {
		spellComp, _ = comp.(*SpellComponent)
	}
	if spellComp == nil {
		spellComp = NewSpellComponent()
		player.AddComponent(spellComp)
	}
	for i, spell := range spells {
		if i >= 5 {
			break
		}
		spellID := fmt.Sprintf("slot_%d_%s_%d", i, spell.Type.String(), spell.Seed)
		known := &KnownSpell{
			Name:     spell.Name,
			Type:     spell.Element.String(),
			ManaCost: spell.Stats.ManaCost,
			CastTime: spell.Stats.CastTime,
			Cooldown: spell.Stats.Cooldown,
			Damage:   spell.Stats.Damage,
			Healing:  spell.Stats.Healing,
		}
		spellComp.LearnSpell(spellID, known)
	}

	log.WithFields(logrus.Fields{
		"equipped_count":    equippedCount,
		"component_created": componentCreated,
	}).Info("Player spells loaded successfully")
	return nil
}

// spawnSpellLight creates a temporary light entity at the spell cast position.
// The light color and intensity are based on the spell's elemental type.
// This function is part of Phase 5.3: Dynamic Lighting System Integration.
func (s *SpellCastingSystem) spawnSpellLight(x, y float64, spell *magic.Spell, duration float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
			"position_x": x,
			"position_y": y,
			"duration":   duration,
		}).Debug("Spawning spell light")
	}

	// Get light color based on spell element
	lightColor := getElementLightColor(spell.Element)

	// Create light entity
	lightEntity := s.world.CreateEntity()

	// Add position component
	lightEntity.AddComponent(&PositionComponent{X: x, Y: y})

	// Create spell light with appropriate radius and color
	// Radius scaled by spell power (damage/healing amount)
	baseRadius := 100.0
	spellPower := float64(spell.Stats.Damage + spell.Stats.Healing)
	powerScale := math.Min(spellPower/50.0, 2.0)
	radius := baseRadius * powerScale

	spellLight := NewSpellLight(radius, lightColor)
	spellLight.Pulsing = true    // Spells have pulsing lights
	spellLight.PulseSpeed = 4.0  // Fast pulse for dramatic effect
	spellLight.PulseAmount = 0.3 // Moderate pulse intensity
	lightEntity.AddComponent(spellLight)

	// Add lifetime component so light despawns automatically
	lightEntity.AddComponent(&LifetimeComponent{
		Duration: duration,
		Elapsed:  0,
	})

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"light_entity_id": lightEntity.ID,
			"spell_power":     spellPower,
			"radius":          radius,
			"duration":        duration,
			"color_r":         lightColor.R,
			"color_g":         lightColor.G,
			"color_b":         lightColor.B,
		}).Info("Spell light entity created")
	}
}

// getElementLightColor returns the appropriate light color for a spell element.
// Colors are chosen to match the visual theme of each element while providing
// good visibility and atmosphere.
func getElementLightColor(element magic.ElementType) color.RGBA {
	switch element {
	case magic.ElementFire:
		return color.RGBA{255, 100, 0, 255} // Orange-red (warm fire)
	case magic.ElementIce:
		return color.RGBA{100, 200, 255, 255} // Cyan (cold ice)
	case magic.ElementLightning:
		return color.RGBA{255, 255, 150, 255} // Bright yellow (electric)
	case magic.ElementEarth:
		return color.RGBA{139, 90, 43, 255} // Brown (earthy)
	case magic.ElementWind:
		return color.RGBA{200, 230, 255, 255} // Light cyan (airy)
	case magic.ElementLight:
		return color.RGBA{255, 255, 220, 255} // Bright white-yellow (holy)
	case magic.ElementDark:
		return color.RGBA{100, 50, 150, 255} // Purple (shadowy)
	case magic.ElementArcane:
		return color.RGBA{200, 100, 255, 255} // Magenta (pure magic)
	case magic.ElementNone:
		return color.RGBA{180, 180, 200, 255} // Neutral grey-blue
	default:
		return color.RGBA{180, 180, 200, 255} // Default grey-blue
	}
}

// LifetimeComponent marks an entity for automatic despawn after a duration.
// Used for temporary entities like spell lights and particle effects.
type LifetimeComponent struct {
	Duration float64 // Total lifetime in seconds
	Elapsed  float64 // Time elapsed since creation
}

// Type implements Component interface.
func (l *LifetimeComponent) Type() string {
	return "lifetime"
}
