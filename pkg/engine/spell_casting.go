package engine

import (
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
	world           *World
	statusEffectSys *StatusEffectSystem
	particleSys     *ParticleSystem         // For visual effects
	audioMgr        *AudioManager           // For sound effects
	tutorialSys     *EbitenTutorialSystem   // For notifications
	comboSys        *SpellCombinationSystem // For combo detection
	logger          *logrus.Entry
}

// NewSpellCastingSystem creates a new spell casting system.
func NewSpellCastingSystem(world *World, statusEffectSys *StatusEffectSystem) *SpellCastingSystem {
	return NewSpellCastingSystemWithLogger(world, statusEffectSys, nil)
}

// NewSpellCastingSystemWithLogger creates a new spell casting system with a logger.
func NewSpellCastingSystemWithLogger(world *World, statusEffectSys *StatusEffectSystem, logger *logrus.Logger) *SpellCastingSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "spell_casting")
	}
	return &SpellCastingSystem{
		world:           world,
		statusEffectSys: statusEffectSys,
		particleSys:     NewParticleSystem(),
		audioMgr:        nil, // Will be set via SetAudioManager()
		logger:          logEntry,
	}
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
	for i := range slots.Cooldowns {
		if slots.Cooldowns[i] > 0 {
			slots.Cooldowns[i] -= deltaTime
			if slots.Cooldowns[i] < 0 {
				slots.Cooldowns[i] = 0
			}
		}
	}
}

// updateCastingProgress advances spell casting progress and completes casts.
func (s *SpellCastingSystem) updateCastingProgress(entity *Entity, slots *SpellSlotComponent, deltaTime float64) {
	if !slots.IsCasting() {
		return
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
	if spell.Stats.CastTime > 0 {
		slots.CastingBar += deltaTime / spell.Stats.CastTime
	} else {
		slots.CastingBar = 1.0
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
	manaComp, hasMana := caster.GetComponent("mana")
	if !hasMana {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  caster.ID,
				"spell_name": spell.Name,
			}).Warn("Entity has no mana component, cannot cast spell")
		}
		return nil
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      caster.ID,
				"component_type": "mana",
			}).Error("Failed to cast mana component to ManaComponent")
		}
		return nil
	}

	if mana.Current < spell.Stats.ManaCost {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":     caster.ID,
				"spell_name":    spell.Name,
				"mana_current":  mana.Current,
				"mana_required": spell.Stats.ManaCost,
			}).Debug("Insufficient mana for spell cast")
		}
		if s.tutorialSys != nil {
			s.tutorialSys.ShowNotification("Not enough mana!", 1.5)
		}
		return nil
	}

	mana.Current -= spell.Stats.ManaCost
	if mana.Current < 0 {
		mana.Current = 0
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      caster.ID,
			"mana_deducted":  spell.Stats.ManaCost,
			"mana_remaining": mana.Current,
		}).Debug("Mana cost deducted")
	}

	return mana
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
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}

	s.logDamageApplied(caster, target, spell, damage, comboMultiplier, health)
	s.applySpellEffectsAndFeedback(target, spell)
}

// getTargetHealth retrieves the health component from a target entity.
func (s *SpellCastingSystem) getTargetHealth(target *Entity) *HealthComponent {
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return nil
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return nil
	}
	return health
}

// calculateSpellDamage computes final damage with combo multiplier.
func (s *SpellCastingSystem) calculateSpellDamage(caster *Entity, spell *magic.Spell) (float64, float64) {
	damage := float64(spell.Stats.Damage)
	comboMultiplier := 1.0
	if s.comboSys != nil {
		comboMultiplier = s.comboSys.GetActiveComboMultiplier(caster)
		damage *= comboMultiplier
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
	if s.statusEffectSys != nil {
		s.applyElementalEffect(target, spell)
	}

	if s.particleSys != nil {
		targetPos, hasPos := target.GetComponent("position")
		if hasPos {
			if pos, ok := targetPos.(*PositionComponent); ok {
				s.spawnElementalHitEffect(pos.X, pos.Y, spell.Element, target.ID)
			}
		}
	}

	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("impact", int64(target.ID))
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
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Calculate healing with combo multiplier (Phase 24.2)
	healing := float64(spell.Stats.Healing)
	comboMultiplier := 1.0
	if s.comboSys != nil {
		comboMultiplier = s.comboSys.GetActiveComboMultiplier(caster)
		healing *= comboMultiplier
	}

	health.Current += healing
	if health.Current > health.Max {
		health.Current = health.Max
	}

	if s.logger != nil {
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

	// Spawn healing visual effect (green/gold particles rising upward)
	if s.particleSys != nil {
		targetPos, hasPos := target.GetComponent("position")
		if hasPos {
			if pos, ok := targetPos.(*PositionComponent); ok {
				config := particles.Config{
					Type:     particles.ParticleMagic,
					Count:    20,
					GenreID:  "fantasy",
					Seed:     int64(target.ID),
					Duration: 1.0,
					SpreadX:  60.0,
					SpreadY:  60.0,
					Gravity:  -80.0, // Rise upward for healing
					MinSize:  4.0,
					MaxSize:  8.0,
					Custom:   map[string]interface{}{"color": "healing"},
				}
				s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
			}
		}
	}

	// Play healing sound effect
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(target.ID))
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
		}
	}
	return casterTeamID
}

// searchNearestInjuredAlly searches for the nearest injured ally within range.
func (s *SpellCastingSystem) searchNearestInjuredAlly(caster *Entity, entities []*Entity, casterTeamID int, maxRange float64) (*Entity, float64) {
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

	return health.Current < health.Max
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
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
			"max_range": maxRange,
		}).Debug("Searching for allies in range")
	}

	entities := s.world.GetEntities()
	var allies []*Entity

	// Get caster's team
	var casterTeamID int
	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
		if team, ok := teamComp.(*TeamComponent); ok {
			casterTeamID = team.TeamID
		}
	}

	for _, entity := range entities {
		// Check if ally (including self)
		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
			team, ok := teamComp.(*TeamComponent)
			if !ok || !team.IsAlly(casterTeamID) {
				continue
			}
		} else if entity != caster {
			continue
		}

		// Check if has health
		if !entity.HasComponent("health") {
			continue
		}

		// Check distance
		dist := GetDistance(caster, entity)
		if dist <= maxRange {
			allies = append(allies, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"ally_count": len(allies),
		}).Debug("Found allies in range")
	}

	return allies
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
		}

		duration := spell.Stats.Duration
		if duration <= 0 {
			duration = 30.0 // Default duration
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

	if s.statusEffectSys == nil {
		if s.logger != nil {
			s.logger.Warn("StatusEffectSystem is nil, cannot apply buff")
		}
		return
	}

	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 30.0 // Default duration
	}

	// Determine buff type based on spell element
	switch spell.Element {
	case magic.ElementWind:
		// Haste - increased attack speed (represented as attack boost)
		s.statusEffectSys.ApplyStatusEffect(caster, "haste", 0.5, duration, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"buff_type": "haste",
				"magnitude": 0.5,
				"duration":  duration,
			}).Info("Buff applied")
		}
	case magic.ElementLight:
		// Strength - increased attack
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.3, duration, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"buff_type": "strength",
				"magnitude": 0.3,
				"duration":  duration,
			}).Info("Buff applied")
		}
	case magic.ElementEarth:
		// Fortify - increased defense
		s.statusEffectSys.ApplyStatusEffect(caster, "fortify", 0.3, duration, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"buff_type": "fortify",
				"magnitude": 0.3,
				"duration":  duration,
			}).Info("Buff applied")
		}
	default:
		// Generic buff - small attack and defense boost
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.2, duration, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
				"buff_type": "strength",
				"magnitude": 0.2,
				"duration":  duration,
			}).Info("Generic buff applied")
		}
	}
}

// castDebuffSpell applies stat reductions to enemies.
func (s *SpellCastingSystem) castDebuffSpell(caster *Entity, spell *magic.Spell, x, y float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Casting debuff spell")
	}

	targets := s.findTargets(caster, spell, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    caster.ID,
			"spell_name":   spell.Name,
			"target_count": len(targets),
		}).Debug("Targets found for debuff spell")
	}

	for _, target := range targets {
		// Apply minor damage if any
		if spell.Stats.Damage > 0 {
			healthComp, hasHealth := target.GetComponent("health")
			if hasHealth {
				if health, ok := healthComp.(*HealthComponent); ok {
					health.Current -= float64(spell.Stats.Damage)
					if health.Current < 0 {
						health.Current = 0
					}
				}
			}
		}

		// Apply debuff effects
		if s.statusEffectSys != nil {
			duration := spell.Stats.Duration
			if duration <= 0 {
				duration = 10.0 // Default duration
			}

			// Determine debuff type based on spell element
			switch spell.Element {
			case magic.ElementDark:
				// Weakness - reduced attack
				s.statusEffectSys.ApplyStatusEffect(target, "weakness", 0.7, duration, 0)
			case magic.ElementEarth:
				// Vulnerability - reduced defense
				s.statusEffectSys.ApplyStatusEffect(target, "vulnerability", 0.7, duration, 0)
			default:
				// Generic debuff - small attack reduction
				s.statusEffectSys.ApplyStatusEffect(target, "weakness", 0.8, duration, 0)
			}
		}
	}
}

// applyElementalEffect applies status effects based on spell element.
func (s *SpellCastingSystem) applyElementalEffect(target *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":  target.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
		}).Debug("Applying elemental effect")
	}

	switch spell.Element {
	case magic.ElementFire:
		// Burning: 10 damage per second for 3 seconds
		s.statusEffectSys.ApplyStatusEffect(target, "burning", 10.0, 3.0, 1.0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
				"effect":    "burning",
				"dps":       10.0,
				"duration":  3.0,
			}).Info("Burning effect applied")
		}

	case magic.ElementIce:
		// Frozen: 50% movement slow for 2 seconds (visual indicator only, actual movement handled by AI)
		s.statusEffectSys.ApplyStatusEffect(target, "frozen", 0.5, 2.0, 0)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"target_id": target.ID,
				"effect":    "frozen",
				"slow":      0.5,
				"duration":  2.0,
			}).Info("Frozen effect applied")
		}

	case magic.ElementLightning:
		// Shocked: chain to nearby enemies
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
		// Apply shocked marker for visual effects
		s.statusEffectSys.ApplyStatusEffect(target, "shocked", 0, 2.0, 0)

	case magic.ElementEarth:
		// Earth spells can apply poison effect
		// Poison: 5 damage per second ignoring armor for 5 seconds
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
}

// shouldApplyPoison returns true 30% of the time for Earth spells.
func (s *SpellCastingSystem) shouldApplyPoison() bool {
	// Use status effect system's RNG if available
	if s.statusEffectSys != nil && s.statusEffectSys.rng != nil {
		return s.statusEffectSys.rng.Float64() < 0.3
	}
	return true
}

// castUtilitySpell handles non-combat spells.
func (s *SpellCastingSystem) castUtilitySpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
			"tags":       spell.Tags,
		}).Debug("Casting utility spell")
	}

	// Determine utility spell type based on tags and element
	switch {
	case containsTag(spell.Tags, "teleport"):
		s.castTeleportSpell(caster, spell)
	case containsTag(spell.Tags, "light"), containsTag(spell.Tags, "reveal"):
		s.castRevealSpell(caster, spell)
	case containsTag(spell.Tags, "speed"), containsTag(spell.Tags, "haste"):
		s.castSpeedBoostSpell(caster, spell)
	default:
		// Generic utility effect based on element
		switch spell.Element {
		case magic.ElementLight:
			s.castRevealSpell(caster, spell)
		case magic.ElementWind:
			s.castSpeedBoostSpell(caster, spell)
		case magic.ElementArcane:
			s.castTeleportSpell(caster, spell)
		default:
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":  caster.ID,
					"spell_name": spell.Name,
					"element":    spell.Element.String(),
				}).Debug("No matching utility spell implementation for element")
			}
		}
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
		return false
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": caster.ID,
			"from_x":    pos.X,
			"from_y":    pos.Y,
			"to_x":      targetX,
			"to_y":      targetY,
			"distance":  maxDist,
		}).Info("Teleport successful")
	}
	pos.X = targetX
	pos.Y = targetY
	return true
}

// spawnTeleportEffects spawns visual and audio effects for teleportation.
func (s *SpellCastingSystem) spawnTeleportEffects(caster *Entity, x, y float64) {
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
		_ = s.audioMgr.PlaySFX("magic", int64(caster.ID))
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
		return
	}

	maxDist := spell.Stats.Range
	if maxDist <= 0 {
		maxDist = 100.0
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

	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": caster.ID,
			}).Warn("Caster has no position component for reveal spell")
		}
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Determine reveal radius from spell stats
	revealRadius := spell.Stats.AreaSize
	if revealRadius <= 0 {
		revealRadius = 200.0 // Default reveal radius
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     caster.ID,
			"reveal_radius": revealRadius,
			"position_x":    pos.X,
			"position_y":    pos.Y,
		}).Info("Reveal spell effect applied")
	}

	// Reveal fog of war (requires access to map UI system)
	// This is handled at a higher level in the game loop
	// For now, we mark this with a temporary status effect that the map system can detect
	if s.statusEffectSys != nil {
		duration := 0.1 // Brief marker effect
		s.statusEffectSys.ApplyStatusEffect(caster, "revealing", revealRadius, duration, 0)
	}

	// Spawn light particles to indicate reveal
	if s.particleSys != nil {
		config := particles.Config{
			Type:     particles.ParticleSpark,
			Count:    40,
			GenreID:  "fantasy",
			Seed:     int64(caster.ID),
			Duration: 1.5,
			SpreadX:  revealRadius,
			SpreadY:  revealRadius,
			Gravity:  -20.0, // Slow rise
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   map[string]interface{}{"color": "light"},
		}
		s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
	}

	// Play light sound
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(caster.ID))
	}
}

// castSpeedBoostSpell applies a temporary speed boost to the caster.
// Increases movement speed for exploration or combat mobility.
func (s *SpellCastingSystem) castSpeedBoostSpell(caster *Entity, spell *magic.Spell) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  caster.ID,
			"spell_name": spell.Name,
		}).Debug("Casting speed boost spell")
	}

	if s.statusEffectSys == nil {
		if s.logger != nil {
			s.logger.Warn("StatusEffectSystem is nil, cannot apply speed boost")
		}
		return
	}

	// Determine duration from spell stats
	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 10.0 // Default speed boost duration
	}

	// Speed multiplier based on spell power
	speedMultiplier := 1.5 // 50% speed increase
	if spell.Rarity >= magic.RarityRare {
		speedMultiplier = 2.0 // 100% speed increase for rare+ spells
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        caster.ID,
			"speed_multiplier": speedMultiplier,
			"duration":         duration,
			"rarity":           spell.Rarity.String(),
		}).Info("Speed boost applied")
	}

	// Apply speed boost as a status effect
	// The movement system will need to check for this effect
	s.statusEffectSys.ApplyStatusEffect(caster, "speed_boost", speedMultiplier, duration, 0)

	// Spawn speed particles (fast-moving wind particles)
	if s.particleSys != nil {
		posComp, hasPos := caster.GetComponent("position")
		if hasPos {
			if pos, ok := posComp.(*PositionComponent); ok {
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
			}
		}
	}

	// Play speed boost sound
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(caster.ID))
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
	return collider, pos
}

// checkCollisionAtPosition checks if position collides with entity.
func (s *SpellCastingSystem) checkCollisionAtPosition(x, y float64, caster, entity *Entity, collider *ColliderComponent, pos *PositionComponent) bool {
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
		return false
	}
	// Must have health to be a valid target
	if !entity.HasComponent("health") {
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
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":  caster.ID,
			"spell_name": spell.Name,
			"range":      spell.Stats.Range,
		}).Debug("Searching for enemies in cone")
	}

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

	dirX, dirY := s.getCasterDirection(caster, x, y)
	if dirX == 0 && dirY == 0 {
		dirX = 1.0
	}

	dirLength := math.Sqrt(dirX*dirX + dirY*dirY)
	dirX /= dirLength
	dirY /= dirLength

	coneAngle := 45.0 * math.Pi / 180.0
	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		entityPos, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		entityPosComp, ok := entityPos.(*PositionComponent)
		if !ok {
			continue
		}

		toEntityX := entityPosComp.X - casterPosComp.X
		toEntityY := entityPosComp.Y - casterPosComp.Y
		dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

		if dist > spell.Stats.Range || dist < 0.1 {
			continue
		}

		toEntityX /= dist
		toEntityY /= dist

		dotProduct := dirX*toEntityX + dirY*toEntityY
		angle := math.Acos(math.Max(-1.0, math.Min(1.0, dotProduct)))

		if angle <= coneAngle {
			targets = append(targets, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found enemies in cone")
	}

	return targets
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

	casterPos, hasCasterPos := caster.GetComponent("position")
	if !hasCasterPos {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"caster_id": caster.ID,
			}).Warn("Caster has no position component for line targeting")
		}
		return nil
	}
	casterPosComp, ok := casterPos.(*PositionComponent)
	if !ok {
		return nil
	}

	dirX, dirY := s.getCasterDirection(caster, x, y)
	if dirX == 0 && dirY == 0 {
		dirX = 1.0
	}

	dirLength := math.Sqrt(dirX*dirX + dirY*dirY)
	dirX /= dirLength
	dirY /= dirLength

	lineWidth := 32.0
	entities := s.world.GetEntities()
	var targets []*Entity

	for _, entity := range entities {
		if !s.isEnemyTarget(caster, entity) {
			continue
		}

		entityPos, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		entityPosComp, ok := entityPos.(*PositionComponent)
		if !ok {
			continue
		}

		toEntityX := entityPosComp.X - casterPosComp.X
		toEntityY := entityPosComp.Y - casterPosComp.Y
		dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

		if dist > spell.Stats.Range || dist < 0.1 {
			continue
		}

		projection := toEntityX*dirX + toEntityY*dirY
		if projection < 0 {
			continue
		}

		perpX := toEntityX - projection*dirX
		perpY := toEntityY - projection*dirY
		perpDist := math.Sqrt(perpX*perpX + perpY*perpY)

		if perpDist <= lineWidth {
			targets = append(targets, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id":    caster.ID,
			"target_count": len(targets),
		}).Debug("Found enemies in line")
	}

	return targets
}

// getCasterDirection determines the caster's facing direction for directional spells.
// Uses velocity if moving, otherwise uses direction towards target point (x, y).
func (s *SpellCastingSystem) getCasterDirection(caster *Entity, targetX, targetY float64) (dirX, dirY float64) {
	// Try to use velocity for moving entities
	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			if vel.VX != 0 || vel.VY != 0 {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"caster_id": caster.ID,
						"dir_x":     vel.VX,
						"dir_y":     vel.VY,
						"source":    "velocity",
					}).Debug("Using velocity for spell direction")
				}
				return vel.VX, vel.VY
			}
		}
	}

	// Fall back to direction towards target point
	if posComp, hasPos := caster.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok {
			dirX = targetX - pos.X
			dirY = targetY - pos.Y
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"caster_id": caster.ID,
					"dir_x":     dirX,
					"dir_y":     dirY,
					"source":    "target_point",
				}).Debug("Using target point for spell direction")
			}
			return dirX, dirY
		}
	}

	// Default to facing right if no position
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"caster_id": caster.ID,
		}).Debug("Using default direction (right)")
	}
	return 1.0, 0.0
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
		return
	}
	slots, ok := spellComp.(*SpellSlotComponent)
	if !ok {
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
	}

	slots.Casting = -1
	slots.CastingBar = 0
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
	regenCount := 0
	for _, entity := range entities {
		manaComp, hasMana := entity.GetComponent("mana")
		if !hasMana {
			continue
		}
		mana, ok := manaComp.(*ManaComponent)
		if !ok {
			continue
		}

		// Only regenerate if not at max
		if mana.Current < mana.Max {
			oldMana := mana.Current
			mana.Current += int(mana.Regen * deltaTime)
			if mana.Current > mana.Max {
				mana.Current = mana.Max
			}

			if mana.Current != oldMana {
				regenCount++
			}
		}
	}

	if s.logger != nil && regenCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_count": regenCount,
			"delta_time":   deltaTime,
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
	log.Debug("Loading player spells")

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

	result, err := generator.Generate(seed, params)
	if err != nil {
		log.WithError(err).Error("Failed to generate player spells")
		return err
	}

	spells := result.([]*magic.Spell)
	log.WithFields(logrus.Fields{
		"spell_count": len(spells),
	}).Debug("Spells generated successfully")

	// Create spell slots component if doesn't exist
	var slots *SpellSlotComponent
	if !player.HasComponent("spell_slots") {
		slots = &SpellSlotComponent{
			Casting: -1,
		}
		player.AddComponent(slots)
	} else {
		slotsComp, _ := player.GetComponent("spell_slots")
		slots = slotsComp.(*SpellSlotComponent)
	}

	// Equip spells to slots
	for i := 0; i < 5 && i < len(spells); i++ {
		slots.SetSlot(i, spells[i])
		log.WithFields(logrus.Fields{
			"slot_index": i,
			"spell_name": spells[i].Name,
			"spell_type": spells[i].Type.String(),
			"element":    spells[i].Element.String(),
		}).Debug("Spell equipped to slot")
	}

	log.Info("Player spells loaded successfully")
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
	powerScale := math.Min(float64(spell.Stats.Damage+spell.Stats.Healing)/50.0, 2.0)
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
			"radius":          radius,
			"color_r":         lightColor.R,
			"color_g":         lightColor.G,
			"color_b":         lightColor.B,
		}).Debug("Spell light spawned")
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
