// Package engine provides base stat tracking for skill bonuses.
// This file implements BaseStatsComponent which stores unmodified base stats
// before skill bonuses and equipment modifiers are applied.
package engine

// BaseStatsComponent stores the entity's base stats before any modifiers.
// This is essential for correctly applying percentage bonuses from skills
// without compound multiplication bugs.
//
// Example:
//
//	Base Attack = 10
//	Skill gives +20% attack
//	Modified Attack = 10 * 1.20 = 12
//
// Without base tracking, reapplying bonuses would compound:
//
//	First apply: 10 * 1.20 = 12
//	Reapply: 12 * 1.20 = 14.4 (incorrect!)
//
// With base tracking:
//
//	Modified Attack = BaseAttack * (1.0 + TotalAttackBonus)
type BaseStatsComponent struct {
	// Base attack damage before modifiers
	BaseAttack float64

	// Base defense before modifiers
	BaseDefense float64

	// Base magic power before modifiers
	BaseMagicPower float64

	// Base max health before modifiers
	BaseMaxHealth float64

	// Base movement speed before modifiers
	BaseSpeed float64

	// Base mana regeneration before modifiers
	BaseManaRegen float64
}

// Type implements Component interface.
func (b *BaseStatsComponent) Type() string {
	return "base_stats"
}

// NewBaseStatsComponent creates a new base stats component with default values.
func NewBaseStatsComponent() *BaseStatsComponent {
	return &BaseStatsComponent{
		BaseAttack:     10.0,
		BaseDefense:    5.0,
		BaseMagicPower: 5.0,
		BaseMaxHealth:  100.0,
		BaseSpeed:      100.0,
		BaseManaRegen:  1.0,
	}
}

// NewBaseStatsFromEntity extracts current stats from an entity as base stats.
// This should be called once when an entity is created to capture initial values.
func NewBaseStatsFromEntity(entity *Entity) *BaseStatsComponent {
	base := NewBaseStatsComponent()

	// Extract attack from AttackComponent
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			base.BaseAttack = attack.Damage
		}
	}

	// Extract defense from StatsComponent
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			base.BaseDefense = stats.Defense
			base.BaseMagicPower = stats.MagicPower
		}
	}

	// Extract max health from HealthComponent
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			base.BaseMaxHealth = health.Max
		}
	}

	// Extract mana regen from ManaComponent
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			base.BaseManaRegen = mana.Regen
		}
	}

	return base
}

// ApplyToEntity applies base stats back to an entity (resetting to unmodified values).
// This should be called before applying skill bonuses to ensure clean calculations.
func (b *BaseStatsComponent) ApplyToEntity(entity *Entity) {
	// Apply base attack
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			attack.Damage = b.BaseAttack
		}
	}

	// Apply base defense and magic power
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Defense = b.BaseDefense
			stats.MagicPower = b.BaseMagicPower
		}
	}

	// Apply base max health
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			// Store current health percentage
			healthPercent := health.Current / health.Max

			// Update max health
			health.Max = b.BaseMaxHealth

			// Restore health percentage
			health.Current = health.Max * healthPercent
			if health.Current < 1 && healthPercent > 0 {
				health.Current = 1 // Ensure at least 1 HP if was alive
			}
		}
	}

	// Apply base mana regen
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			mana.Regen = b.BaseManaRegen
		}
	}
}
