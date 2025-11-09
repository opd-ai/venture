// Package engine provides the core game systems including skill progression.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/skills"
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

	// Apply skill bonuses to all entities with skill trees
	for _, entity := range entities {
		if !entity.HasComponent("skill_tree") {
			continue
		}

		s.applySkillBonuses(entity)
	}
}

// applySkillBonuses calculates and applies all learned skill effects to an entity.
func (s *SkillProgressionSystem) applySkillBonuses(entity *Entity) {
	comp, ok := entity.GetComponent("skill_tree")
	if !ok {
		return
	}
	// Type assert with safety check
	treeComp, ok := comp.(*SkillTreeComponent)
	if !ok {
		return
	}

	// Get stats component
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return // No stats to modify
	}
	// Type assert with safety check
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
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

	// Accumulate bonuses from all learned skills
	for skillID := range treeComp.LearnedSkills {
		skill := treeComp.Tree.GetSkillByID(skillID)
		if skill == nil {
			continue
		}

		// Get skill level for scaling
		skillLevel := treeComp.GetSkillLevel(skillID)
		if skillLevel == 0 {
			continue
		}

		// Apply each effect
		for _, effect := range skill.Effects {
			s.applyEffect(bonuses, effect, float64(skillLevel))
		}
	}

	// Apply accumulated bonuses to stats
	s.applyBonusesToStats(entity, stats, bonuses)
}

// applyEffect adds a single effect to the bonus accumulator.
func (s *SkillProgressionSystem) applyEffect(bonuses *SkillBonuses, effect skills.Effect, skillLevel float64) {
	// Scale effect value by skill level
	value := effect.Value * skillLevel

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
	}
}

// applyBonusesToStats modifies stats based on calculated bonuses.
// GAP-008 REPAIR: Store base stats and reapply bonuses from scratch to avoid compounding.
// This system uses BaseStatsComponent to store original stat values and recalculates
// derived stats from the base values each update, preventing multiplicative compounding.
func (s *SkillProgressionSystem) applyBonusesToStats(entity *Entity, stats *StatsComponent, bonuses *SkillBonuses) {
	// Apply direct bonuses to critical hit stats (already in correct units)
	if bonuses.CritChanceBonus != 0 {
		// Reset crit chance to base (5%) and add bonuses
		baseCritChance := 0.05
		stats.CritChance = baseCritChance + bonuses.CritChanceBonus
		// Cap at 100%
		if stats.CritChance > 1.0 {
			stats.CritChance = 1.0
		}
	}

	if bonuses.CritDamageBonus != 0 {
		// Reset crit damage to base (2.0x) and add bonuses
		baseCritDamage := 2.0
		stats.CritDamage = baseCritDamage + bonuses.CritDamageBonus
	}

	// Apply attack/defense/magic bonuses using base stats
	baseStatsComp, hasBaseStats := entity.GetComponent("base_stats")
	if hasBaseStats {
		// Type assert with safety check
		baseStats, ok := baseStatsComp.(*BaseStatsComponent)
		if !ok {
			return
		}

		// Apply attack bonus
		if bonuses.DamageBonus != 0 {
			if attackComp, ok := entity.GetComponent("attack"); ok {
				// Type assert with safety check
				if attack, ok := attackComp.(*AttackComponent); ok {
					attack.Damage = baseStats.BaseAttack * (1.0 + bonuses.DamageBonus)
				}
			}
		}

		// Apply defense bonus
		if bonuses.DefenseBonus != 0 {
			stats.Defense = baseStats.BaseDefense * (1.0 + bonuses.DefenseBonus)
		}

		// Apply magic power bonus
		if bonuses.MagicPowerBonus != 0 {
			stats.MagicPower = baseStats.BaseMagicPower * (1.0 + bonuses.MagicPowerBonus)
		}

		// Apply health bonus
		if bonuses.HealthBonus != 0 {
			if healthComp, ok := entity.GetComponent("health"); ok {
				// Type assert with safety check
				if health, ok := healthComp.(*HealthComponent); ok {
					oldMax := health.Max
					health.Max = baseStats.BaseMaxHealth * (1.0 + bonuses.HealthBonus)
					// Scale current health proportionally
					if oldMax > 0 {
						health.Current = health.Current * (health.Max / oldMax)
					}
				}
			}
		}

		// Apply mana regen bonus
		if bonuses.ManaRegenBonus != 0 {
			if manaComp, ok := entity.GetComponent("mana"); ok {
				// Type assert with safety check
				if mana, ok := manaComp.(*ManaComponent); ok {
					mana.Regen = baseStats.BaseManaRegen * (1.0 + bonuses.ManaRegenBonus)
				}
			}
		}
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
	system := NewSkillProgressionSystem()
	system.applySkillBonuses(entity)
}
