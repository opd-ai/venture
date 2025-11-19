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
	s.frameCounter++

	// Only update periodically to avoid excessive recalculation
	if s.frameCounter < s.updateInterval {
		return
	}
	s.frameCounter = 0

	log.WithFields(log.Fields{
		"system_name":   "skill_progression",
		"entity_count":  len(entities),
		"frame_counter": s.frameCounter,
	}).Debug("Starting skill progression update cycle")

	// Apply skill bonuses to all entities with skill trees
	entitiesProcessed := 0
	for _, entity := range entities {
		if !entity.HasComponent("skill_tree") {
			continue
		}

		s.applySkillBonuses(entity)
		entitiesProcessed++
	}

	log.WithFields(log.Fields{
		"system_name":        "skill_progression",
		"entities_processed": entitiesProcessed,
		"total_entities":     len(entities),
	}).Debug("Completed skill progression update cycle")
}

// applySkillBonuses calculates and applies all learned skill effects to an entity.
func (s *SkillProgressionSystem) applySkillBonuses(entity *Entity) {
	log.WithFields(log.Fields{
		"entity_id":      entity.ID,
		"component_type": "skill_tree",
	}).Debug("Applying skill bonuses to entity")

	comp, ok := entity.GetComponent("skill_tree")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "skill_tree",
		}).Debug("Entity missing skill_tree component")
		return
	}
	// Type assert with safety check
	treeComp, ok := comp.(*SkillTreeComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "skill_tree",
		}).Error("Failed to type assert skill_tree component")
		return
	}

	// Get stats component
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "stats",
		}).Debug("Entity missing stats component, no stats to modify")
		return // No stats to modify
	}
	// Type assert with safety check
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "stats",
		}).Error("Failed to type assert stats component")
		return
	}

	// Reset bonus stats (we'll recalculate from scratch)
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

	// Accumulate bonuses from all learned skills
	skillsProcessed := 0
	effectsApplied := 0
	for skillID := range treeComp.LearnedSkills {
		skill := treeComp.Tree.GetSkillByID(skillID)
		if skill == nil {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"skill_id":  skillID,
			}).Warn("Learned skill not found in tree")
			continue
		}

		// Get skill level for scaling
		skillLevel := treeComp.GetSkillLevel(skillID)
		if skillLevel == 0 {
			log.WithFields(log.Fields{
				"entity_id": entity.ID,
				"skill_id":  skillID,
			}).Debug("Skill has zero level, skipping")
			continue
		}

		log.WithFields(log.Fields{
			"entity_id":    entity.ID,
			"skill_id":     skillID,
			"skill_level":  skillLevel,
			"effect_count": len(skill.Effects),
		}).Debug("Processing skill effects")

		// Apply each effect
		for _, effect := range skill.Effects {
			s.applyEffect(bonuses, effect, float64(skillLevel))
			effectsApplied++
		}
		skillsProcessed++
	}

	log.WithFields(log.Fields{
		"entity_id":        entity.ID,
		"skills_processed": skillsProcessed,
		"effects_applied":  effectsApplied,
		"damage_bonus":     bonuses.DamageBonus,
		"defense_bonus":    bonuses.DefenseBonus,
		"health_bonus":     bonuses.HealthBonus,
	}).Debug("Completed bonus calculation")

	// Apply accumulated bonuses to stats
	s.applyBonusesToStats(entity, stats, bonuses)
}

// applyEffect adds a single effect to the bonus accumulator.
func (s *SkillProgressionSystem) applyEffect(bonuses *SkillBonuses, effect skills.Effect, skillLevel float64) {
	// Scale effect value by skill level
	value := effect.Value * skillLevel

	log.WithFields(log.Fields{
		"effect_type":  effect.Type,
		"effect_value": effect.Value,
		"skill_level":  skillLevel,
		"scaled_value": value,
		"is_percent":   effect.IsPercent,
	}).Debug("Applying skill effect")

	// Map effect types to stat bonuses
	switch effect.Type {
	case "damage", "attack_power", "strength":
		if effect.IsPercent {
			bonuses.DamageBonus += value
		} else {
			bonuses.DamageBonus += value / 100 // Convert flat bonus to percentage
		}

	case "defense", "armor", "toughness":
		if effect.IsPercent {
			bonuses.DefenseBonus += value
		} else {
			bonuses.DefenseBonus += value / 100
		}

	case "health", "vitality", "max_health":
		if effect.IsPercent {
			bonuses.HealthBonus += value
		} else {
			bonuses.HealthBonus += value / 100
		}

	case "speed", "movement_speed", "agility":
		if effect.IsPercent {
			bonuses.SpeedBonus += value
		} else {
			bonuses.SpeedBonus += value / 100
		}

	case "crit_chance", "critical_chance":
		bonuses.CritChanceBonus += value

	case "crit_damage", "critical_damage":
		if effect.IsPercent {
			bonuses.CritDamageBonus += value
		} else {
			bonuses.CritDamageBonus += value / 100
		}

	case "magic_power", "spell_power", "intelligence":
		if effect.IsPercent {
			bonuses.MagicPowerBonus += value
		} else {
			bonuses.MagicPowerBonus += value / 100
		}

	case "mana_regen", "mana_regeneration":
		if effect.IsPercent {
			bonuses.ManaRegenBonus += value
		} else {
			bonuses.ManaRegenBonus += value / 100
		}

	case "cooldown_reduction", "haste":
		bonuses.CooldownReduction += value

	default:
		log.WithFields(log.Fields{
			"effect_type": effect.Type,
		}).Warn("Unknown effect type, not applied")
	}
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

	// Apply direct bonuses to critical hit stats (already in correct units)
	if bonuses.CritChanceBonus != 0 {
		// Reset crit chance to base (5%) and add bonuses
		baseCritChance := 0.05
		oldCritChance := stats.CritChance
		stats.CritChance = baseCritChance + bonuses.CritChanceBonus
		// Cap at 100%
		if stats.CritChance > 1.0 {
			stats.CritChance = 1.0
		}
		log.WithFields(log.Fields{
			"entity_id":       entity.ID,
			"old_crit_chance": oldCritChance,
			"new_crit_chance": stats.CritChance,
			"bonus":           bonuses.CritChanceBonus,
		}).Debug("Updated crit chance")
	}

	if bonuses.CritDamageBonus != 0 {
		// Reset crit damage to base (2.0x) and add bonuses
		baseCritDamage := 2.0
		oldCritDamage := stats.CritDamage
		stats.CritDamage = baseCritDamage + bonuses.CritDamageBonus
		log.WithFields(log.Fields{
			"entity_id":       entity.ID,
			"old_crit_damage": oldCritDamage,
			"new_crit_damage": stats.CritDamage,
			"bonus":           bonuses.CritDamageBonus,
		}).Debug("Updated crit damage")
	}

	// Apply attack/defense/magic bonuses using base stats
	baseStatsComp, hasBaseStats := entity.GetComponent("base_stats")
	if hasBaseStats {
		// Type assert with safety check
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

		// Apply attack bonus
		if bonuses.DamageBonus != 0 {
			if attackComp, ok := entity.GetComponent("attack"); ok {
				// Type assert with safety check
				if attack, ok := attackComp.(*AttackComponent); ok {
					oldDamage := attack.Damage
					attack.Damage = baseStats.BaseAttack * (1.0 + bonuses.DamageBonus)
					log.WithFields(log.Fields{
						"entity_id":  entity.ID,
						"old_damage": oldDamage,
						"new_damage": attack.Damage,
						"bonus":      bonuses.DamageBonus,
					}).Debug("Updated attack damage")
				}
			}
		}

		// Apply defense bonus
		if bonuses.DefenseBonus != 0 {
			oldDefense := stats.Defense
			stats.Defense = baseStats.BaseDefense * (1.0 + bonuses.DefenseBonus)
			log.WithFields(log.Fields{
				"entity_id":   entity.ID,
				"old_defense": oldDefense,
				"new_defense": stats.Defense,
				"bonus":       bonuses.DefenseBonus,
			}).Debug("Updated defense")
		}

		// Apply magic power bonus
		if bonuses.MagicPowerBonus != 0 {
			oldMagicPower := stats.MagicPower
			stats.MagicPower = baseStats.BaseMagicPower * (1.0 + bonuses.MagicPowerBonus)
			log.WithFields(log.Fields{
				"entity_id":       entity.ID,
				"old_magic_power": oldMagicPower,
				"new_magic_power": stats.MagicPower,
				"bonus":           bonuses.MagicPowerBonus,
			}).Debug("Updated magic power")
		}

		// Apply health bonus
		if bonuses.HealthBonus != 0 {
			if healthComp, ok := entity.GetComponent("health"); ok {
				// Type assert with safety check
				if health, ok := healthComp.(*HealthComponent); ok {
					oldMax := health.Max
					oldCurrent := health.Current
					health.Max = baseStats.BaseMaxHealth * (1.0 + bonuses.HealthBonus)
					// Scale current health proportionally
					if oldMax > 0 {
						health.Current = health.Current * (health.Max / oldMax)
					}
					log.WithFields(log.Fields{
						"entity_id":      entity.ID,
						"old_max_health": oldMax,
						"new_max_health": health.Max,
						"old_current":    oldCurrent,
						"new_current":    health.Current,
						"bonus":          bonuses.HealthBonus,
					}).Debug("Updated health")
				}
			}
		}

		// Apply mana regen bonus
		if bonuses.ManaRegenBonus != 0 {
			if manaComp, ok := entity.GetComponent("mana"); ok {
				// Type assert with safety check
				if mana, ok := manaComp.(*ManaComponent); ok {
					oldRegen := mana.Regen
					mana.Regen = baseStats.BaseManaRegen * (1.0 + bonuses.ManaRegenBonus)
					log.WithFields(log.Fields{
						"entity_id": entity.ID,
						"old_regen": oldRegen,
						"new_regen": mana.Regen,
						"bonus":     bonuses.ManaRegenBonus,
					}).Debug("Updated mana regen")
				}
			}
		}
	} else {
		log.WithFields(log.Fields{
			"entity_id": entity.ID,
		}).Debug("Entity missing base_stats component, cannot apply percentage bonuses")
	}

	// Note: Speed bonus not applied here since it's managed by separate MovementComponent
	// Cooldown reduction applied during spell casting
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
