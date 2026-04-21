// Package engine provides the core game systems including skill progression.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/skills"
	log "github.com/sirupsen/logrus"
)

// SkillProgressionSystem applies learned skill effects to entities.
// This system handles:
// - Passive skill bonuses (damage, defense, health, etc.)
// - Stat modifications based on learned skills
// - Dynamic updates when skills are learned/unlearned
type SkillProgressionSystem struct {
	updateInterval int // Frames between stat recalculations
	frameCounter   int
}

// NewSkillProgressionSystem creates a new skill progression system.
// Updates stat bonuses every 60 frames (1 second at 60 FPS).
func NewSkillProgressionSystem() *SkillProgressionSystem {
	log.WithFields(log.Fields{
		"system_name":     "skill_progression",
		"update_interval": 60,
	}).Debug("Creating skill progression system")

	return &SkillProgressionSystem{
		updateInterval: 60,
		frameCounter:   0,
	}
}

// Update applies skill effects to entities with skill trees.
// This recalculates stat bonuses based on all learned skills.
func (s *SkillProgressionSystem) Update(entities []*Entity, deltaTime float64) {
	if !s.shouldUpdateThisFrame() {
		return
	}

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"system_name":   "skill_progression",
			"entity_count":  len(entities),
			"frame_counter": s.frameCounter,
		}).Debug("Starting skill progression update cycle")
	}

	entitiesProcessed := s.processSkillEntities(entities)

	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"system_name":        "skill_progression",
			"entities_processed": entitiesProcessed,
			"total_entities":     len(entities),
		}).Debug("Completed skill progression update cycle")
	}
}

// shouldUpdateThisFrame checks if enough frames have passed for an update.
func (s *SkillProgressionSystem) shouldUpdateThisFrame() bool {
	s.frameCounter++
	if s.frameCounter < s.updateInterval {
		return false
	}
	s.frameCounter = 0
	return true
}

// processSkillEntities applies skill bonuses to all entities with skill trees.
func (s *SkillProgressionSystem) processSkillEntities(entities []*Entity) int {
	entitiesProcessed := 0
	for _, entity := range entities {
		if !entity.HasComponent("skill_tree") {
			continue
		}
		s.applySkillBonuses(entity)
		entitiesProcessed++
	}
	return entitiesProcessed
}

// applySkillBonuses calculates and applies all learned skill effects to an entity.
func (s *SkillProgressionSystem) applySkillBonuses(entity *Entity) {
	log.WithFields(log.Fields{
		"entity_id":      entity.ID,
		"component_type": "skill_tree",
	}).Debug("Applying skill bonuses to entity")

	treeComp := s.validateSkillTreeComponent(entity)
	if treeComp == nil {
		return
	}

	stats := s.validateStatsComponent(entity)
	if stats == nil {
		return
	}

	bonuses := s.calculateSkillBonuses(entity, treeComp)
	s.applyBonusesToStats(entity, stats, bonuses)
}

// validateSkillTreeComponent retrieves and validates the skill tree component.
func (s *SkillProgressionSystem) validateSkillTreeComponent(entity *Entity) *SkillTreeComponent {
	comp, ok := entity.GetComponent("skill_tree")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "skill_tree",
		}).Debug("Entity missing skill_tree component")
		return nil
	}

	treeComp, ok := comp.(*SkillTreeComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "skill_tree",
		}).Error("Failed to type assert skill_tree component")
		return nil
	}

	return treeComp
}

// validateStatsComponent retrieves and validates the stats component.
func (s *SkillProgressionSystem) validateStatsComponent(entity *Entity) *StatsComponent {
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "stats",
		}).Debug("Entity missing stats component, no stats to modify")
		return nil
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "stats",
		}).Error("Failed to type assert stats component")
		return nil
	}

	return stats
}

// calculateSkillBonuses calculates total bonuses from all learned skills.
func (s *SkillProgressionSystem) calculateSkillBonuses(entity *Entity, treeComp *SkillTreeComponent) *SkillBonuses {
	bonuses := &SkillBonuses{
		DamageBonus:       0,
		DefenseBonus:      0,
		HealthBonus:       0,
		SpeedBonus:        0,
		CritChanceBonus:   0,
		CritDamageBonus:   0,
		MagicPowerBonus:   0,
		ManaRegenBonus:    0,
		CooldownReduction: 0,
	}

	log.WithFields(log.Fields{
		"entity_id":      entity.ID,
		"learned_skills": len(treeComp.LearnedSkills),
	}).Debug("Starting bonus calculation from learned skills")

	skillsProcessed := 0
	effectsApplied := 0
	for skillID := range treeComp.LearnedSkills {
		effects := s.processLearnedSkill(entity, treeComp, skillID, bonuses)
		if effects > 0 {
			skillsProcessed++
			effectsApplied += effects
		}
	}

	log.WithFields(log.Fields{
		"entity_id":        entity.ID,
		"skills_processed": skillsProcessed,
		"effects_applied":  effectsApplied,
		"damage_bonus":     bonuses.DamageBonus,
		"defense_bonus":    bonuses.DefenseBonus,
		"health_bonus":     bonuses.HealthBonus,
	}).Debug("Completed bonus calculation")

	return bonuses
}

// processLearnedSkill processes effects from a single learned skill and returns the number of effects applied.
func (s *SkillProgressionSystem) processLearnedSkill(entity *Entity, treeComp *SkillTreeComponent, skillID string, bonuses *SkillBonuses) int {
	skill := treeComp.Tree.GetSkillByID(skillID)
	if skill == nil {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"skill_id":  skillID,
		}).Warn("Learned skill not found in tree")
		return 0
	}

	skillLevel := treeComp.GetSkillLevel(skillID)
	if skillLevel == 0 {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
			"skill_id":  skillID,
		}).Debug("Skill has zero level, skipping")
		return 0
	}

	log.WithFields(log.Fields{
		"entity_id":    entity.ID,
		"skill_id":     skillID,
		"skill_level":  skillLevel,
		"effect_count": len(skill.Effects),
	}).Debug("Processing skill effects")

	for _, effect := range skill.Effects {
		s.applyEffect(bonuses, effect, float64(skillLevel))
	}

	return len(skill.Effects)
}

// applyEffect adds a single effect to the bonus accumulator.
func (s *SkillProgressionSystem) applyEffect(bonuses *SkillBonuses, effect skills.Effect, skillLevel float64) {
	value := effect.Value * skillLevel
	s.logEffectApplication(effect, skillLevel, value)

	switch effect.Type {
	case "damage", "attack_power", "strength":
		s.applyDamageBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "defense", "armor", "toughness":
		s.applyDefenseBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "health", "vitality", "max_health":
		s.applyHealthBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "speed", "movement_speed", "agility":
		s.applySpeedBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "crit_chance", "critical_chance":
		bonuses.CritChanceBonus += value
	case "crit_damage", "critical_damage":
		s.applyCritDamageBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "magic_power", "spell_power", "intelligence":
		s.applyMagicPowerBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "mana_regen", "mana_regeneration":
		s.applyManaRegenBonusToSkillBonuses(bonuses, value, effect.IsPercent)
	case "cooldown_reduction", "haste":
		bonuses.CooldownReduction += value
	default:
		s.logUnknownEffect(effect.Type)
	}
}

// logEffectApplication logs the application of a skill effect.
func (s *SkillProgressionSystem) logEffectApplication(effect skills.Effect, skillLevel, scaledValue float64) {
	log.WithFields(log.Fields{
		"effect_type":  effect.Type,
		"effect_value": effect.Value,
		"skill_level":  skillLevel,
		"scaled_value": scaledValue,
		"is_percent":   effect.IsPercent,
	}).Debug("Applying skill effect")
}

// applyDamageBonusToSkillBonuses applies damage/attack power bonus to skill bonuses.
func (s *SkillProgressionSystem) applyDamageBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.DamageBonus += value
	} else {
		bonuses.DamageBonus += value / 100
	}
}

// applyDefenseBonusToSkillBonuses applies defense/armor bonus to skill bonuses.
func (s *SkillProgressionSystem) applyDefenseBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.DefenseBonus += value
	} else {
		bonuses.DefenseBonus += value / 100
	}
}

// applyHealthBonusToSkillBonuses applies health/vitality bonus to skill bonuses.
func (s *SkillProgressionSystem) applyHealthBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.HealthBonus += value
	} else {
		bonuses.HealthBonus += value / 100
	}
}

// applySpeedBonusToSkillBonuses applies speed/movement bonus to skill bonuses.
func (s *SkillProgressionSystem) applySpeedBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.SpeedBonus += value
	} else {
		bonuses.SpeedBonus += value / 100
	}
}

// applyCritDamageBonusToSkillBonuses applies critical damage bonus to skill bonuses.
func (s *SkillProgressionSystem) applyCritDamageBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.CritDamageBonus += value
	} else {
		bonuses.CritDamageBonus += value / 100
	}
}

// applyMagicPowerBonusToSkillBonuses applies magic/spell power bonus to skill bonuses.
func (s *SkillProgressionSystem) applyMagicPowerBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.MagicPowerBonus += value
	} else {
		bonuses.MagicPowerBonus += value / 100
	}
}

// applyManaRegenBonusToSkillBonuses applies mana regeneration bonus to skill bonuses.
func (s *SkillProgressionSystem) applyManaRegenBonusToSkillBonuses(bonuses *SkillBonuses, value float64, isPercent bool) {
	if isPercent {
		bonuses.ManaRegenBonus += value
	} else {
		bonuses.ManaRegenBonus += value / 100
	}
}

// logUnknownEffect logs a warning for unknown effect types.
func (s *SkillProgressionSystem) logUnknownEffect(effectType string) {
	log.WithFields(log.Fields{
		"effect_type": effectType,
	}).Warn("Unknown effect type, not applied")
}

// applyBonusesToStats modifies stats based on calculated bonuses.
// GAP-008 REPAIR: Store base stats and reapply bonuses from scratch to avoid compounding.
// This system uses BaseStatsComponent to store original stat values and recalculates
// derived stats from the base values each update, preventing multiplicative compounding.
func (s *SkillProgressionSystem) applyBonusesToStats(entity *Entity, stats *StatsComponent, bonuses *SkillBonuses) {
	log.WithFields(log.Fields{
		"entity_id":          entity.ID,
		"damage_bonus":       bonuses.DamageBonus,
		"defense_bonus":      bonuses.DefenseBonus,
		"health_bonus":       bonuses.HealthBonus,
		"crit_chance_bonus":  bonuses.CritChanceBonus,
		"crit_damage_bonus":  bonuses.CritDamageBonus,
		"magic_power_bonus":  bonuses.MagicPowerBonus,
		"mana_regen_bonus":   bonuses.ManaRegenBonus,
		"cooldown_reduction": bonuses.CooldownReduction,
	}).Debug("Applying bonuses to entity stats")

	s.applyCriticalHitBonuses(entity, stats, bonuses)
	s.applyPercentageBonuses(entity, stats, bonuses)
}

// applyCriticalHitBonuses applies critical hit chance and damage bonuses.
func (s *SkillProgressionSystem) applyCriticalHitBonuses(entity *Entity, stats *StatsComponent, bonuses *SkillBonuses) {
	if bonuses.CritChanceBonus != 0 {
		s.applyCritChanceBonus(entity, stats, bonuses.CritChanceBonus)
	}
	if bonuses.CritDamageBonus != 0 {
		s.applyCritDamageBonus(entity, stats, bonuses.CritDamageBonus)
	}
}

// applyCritChanceBonus updates critical hit chance with capping at 100%.
func (s *SkillProgressionSystem) applyCritChanceBonus(entity *Entity, stats *StatsComponent, bonus float64) {
	baseCritChance := 0.05
	oldCritChance := stats.CritChance
	stats.CritChance = baseCritChance + bonus
	if stats.CritChance > 1.0 {
		stats.CritChance = 1.0
	}
	log.WithFields(log.Fields{
		"entity_id":       entity.ID,
		"old_crit_chance": oldCritChance,
		"new_crit_chance": stats.CritChance,
		"bonus":           bonus,
	}).Debug("Updated crit chance")
}

// applyCritDamageBonus updates critical damage multiplier.
func (s *SkillProgressionSystem) applyCritDamageBonus(entity *Entity, stats *StatsComponent, bonus float64) {
	baseCritDamage := 2.0
	oldCritDamage := stats.CritDamage
	stats.CritDamage = baseCritDamage + bonus
	log.WithFields(log.Fields{
		"entity_id":       entity.ID,
		"old_crit_damage": oldCritDamage,
		"new_crit_damage": stats.CritDamage,
		"bonus":           bonus,
	}).Debug("Updated crit damage")
}

// applyPercentageBonuses applies all percentage-based stat bonuses using base stats.
func (s *SkillProgressionSystem) applyPercentageBonuses(entity *Entity, stats *StatsComponent, bonuses *SkillBonuses) {
	baseStatsComp, hasBaseStats := entity.GetComponent("base_stats")
	if !hasBaseStats {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Debug("Entity missing base_stats component, cannot apply percentage bonuses")
		return
	}

	baseStats, ok := baseStatsComp.(*BaseStatsComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "base_stats",
		}).Error("Failed to type assert base_stats component")
		return
	}

	log.WithFields(log.Fields{
		"entity_id":        entity.ID,
		"base_attack":      baseStats.BaseAttack,
		"base_defense":     baseStats.BaseDefense,
		"base_magic_power": baseStats.BaseMagicPower,
		"base_max_health":  baseStats.BaseMaxHealth,
		"base_mana_regen":  baseStats.BaseManaRegen,
	}).Debug("Applying bonuses using base stats")

	s.applyAttackBonus(entity, baseStats, bonuses.DamageBonus)
	s.applyDefenseBonus(entity, stats, baseStats, bonuses.DefenseBonus)
	s.applyMagicPowerBonus(entity, stats, baseStats, bonuses.MagicPowerBonus)
	s.applyHealthBonus(entity, baseStats, bonuses.HealthBonus)
	s.applyManaRegenBonus(entity, baseStats, bonuses.ManaRegenBonus)
}

// applyAttackBonus updates attack damage based on base attack.
func (s *SkillProgressionSystem) applyAttackBonus(entity *Entity, baseStats *BaseStatsComponent, bonus float64) {
	if bonus == 0 {
		return
	}
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			oldDamage := attack.Damage
			attack.Damage = baseStats.BaseAttack * (1.0 + bonus)
			log.WithFields(log.Fields{
				"entity_id":  entity.ID,
				"old_damage": oldDamage,
				"new_damage": attack.Damage,
				"bonus":      bonus,
			}).Debug("Updated attack damage")
		}
	}
}

// applyDefenseBonus updates defense stat based on base defense.
func (s *SkillProgressionSystem) applyDefenseBonus(entity *Entity, stats *StatsComponent, baseStats *BaseStatsComponent, bonus float64) {
	if bonus == 0 {
		return
	}
	oldDefense := stats.Defense
	stats.Defense = baseStats.BaseDefense * (1.0 + bonus)
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"old_defense": oldDefense,
		"new_defense": stats.Defense,
		"bonus":       bonus,
	}).Debug("Updated defense")
}

// applyMagicPowerBonus updates magic power stat based on base magic power.
func (s *SkillProgressionSystem) applyMagicPowerBonus(entity *Entity, stats *StatsComponent, baseStats *BaseStatsComponent, bonus float64) {
	if bonus == 0 {
		return
	}
	oldMagicPower := stats.MagicPower
	stats.MagicPower = baseStats.BaseMagicPower * (1.0 + bonus)
	log.WithFields(log.Fields{
		"entity_id":       entity.ID,
		"old_magic_power": oldMagicPower,
		"new_magic_power": stats.MagicPower,
		"bonus":           bonus,
	}).Debug("Updated magic power")
}

// applyHealthBonus updates max health and scales current health proportionally.
func (s *SkillProgressionSystem) applyHealthBonus(entity *Entity, baseStats *BaseStatsComponent, bonus float64) {
	if bonus == 0 {
		return
	}
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			oldMax := health.Max
			oldCurrent := health.Current
			health.Max = baseStats.BaseMaxHealth * (1.0 + bonus)
			if oldMax > 0 {
				health.Current = health.Current * (health.Max / oldMax)
			}
			log.WithFields(log.Fields{
				"entity_id":      entity.ID,
				"old_max_health": oldMax,
				"new_max_health": health.Max,
				"old_current":    oldCurrent,
				"new_current":    health.Current,
				"bonus":          bonus,
			}).Debug("Updated health")
		}
	}
}

// applyManaRegenBonus updates mana regeneration based on base mana regen.
func (s *SkillProgressionSystem) applyManaRegenBonus(entity *Entity, baseStats *BaseStatsComponent, bonus float64) {
	if bonus == 0 {
		return
	}
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			oldRegen := mana.Regen
			mana.Regen = baseStats.BaseManaRegen * (1.0 + bonus)
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"old_regen": oldRegen,
				"new_regen": mana.Regen,
				"bonus":     bonus,
			}).Debug("Updated mana regen")
		}
	}
}

// SkillBonuses accumulates all skill effect bonuses.
type SkillBonuses struct {
	DamageBonus       float64 // Percentage bonus to attack
	DefenseBonus      float64 // Percentage bonus to defense
	HealthBonus       float64 // Percentage bonus to max health
	SpeedBonus        float64 // Percentage bonus to speed
	CritChanceBonus   float64 // Flat bonus to crit chance (0.0-1.0)
	CritDamageBonus   float64 // Percentage bonus to crit damage
	MagicPowerBonus   float64 // Percentage bonus to magic power
	ManaRegenBonus    float64 // Percentage bonus to mana regen
	CooldownReduction float64 // Percentage cooldown reduction
}

// RecalculateSkillBonuses immediately recalculates skill bonuses for an entity.
// Call this after learning/unlearning skills to update stats immediately.
func RecalculateSkillBonuses(entity *Entity) {
	if entity == nil {
		log.Warn("RecalculateSkillBonuses called with nil entity")
		return
	}

	log.WithFields(log.Fields{
		"entity_id": entity.ID,
		"operation": "recalculate_skill_bonuses",
	}).Info("Immediately recalculating skill bonuses")

	system := NewSkillProgressionSystem()
	system.applySkillBonuses(entity)

	log.WithFields(log.Fields{
		"entity_id": entity.ID,
		"operation": "recalculate_skill_bonuses",
	}).Debug("Completed immediate skill bonus recalculation")
}
