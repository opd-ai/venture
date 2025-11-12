// Package engine provides stat growth mechanics for character classes.
// This file implements the StatGrowth system which defines per-level scaling
// for character classes as part of Phase 25.1.
package engine

// StatGrowth defines per-level stat increases for a character class.
// Used by ClassProgressionSystem to scale stats as characters level up.
type StatGrowth struct {
	// Per-level increases (flat bonuses)
	HPPerLevel         float64
	ManaPerLevel       float64
	AttackPerLevel     float64
	DefensePerLevel    float64
	MagicPowerPerLevel float64
	SpeedPerLevel      float64

	// Percentage increases per level (applied after flat bonuses)
	HPPercentPerLevel     float64 // e.g., 0.05 = 5% more HP per level
	ManaPercentPerLevel   float64
	AttackPercentPerLevel float64
}

// CalculateStatAtLevel calculates the total stat value at a given level.
// Formula: BaseStat + (FlatPerLevel × (level - 1)) × (1 + PercentPerLevel × (level - 1))
//
// Example: Base HP 100, +10 per level, +5% per level
//
//	Level 1: 100 + 0 = 100
//	Level 5: 100 + (10 × 4) × (1 + 0.05 × 4) = 140 × 1.2 = 168
func (sg *StatGrowth) CalculateHP(baseHP float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	flat := baseHP + (sg.HPPerLevel * float64(level-1))
	multiplier := 1.0 + (sg.HPPercentPerLevel * float64(level-1))
	return flat * multiplier
}

// CalculateMana calculates mana at a given level using the same formula as HP.
func (sg *StatGrowth) CalculateMana(baseMana float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	flat := baseMana + (sg.ManaPerLevel * float64(level-1))
	multiplier := 1.0 + (sg.ManaPercentPerLevel * float64(level-1))
	return flat * multiplier
}

// CalculateAttack calculates attack at a given level.
func (sg *StatGrowth) CalculateAttack(baseAttack float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	flat := baseAttack + (sg.AttackPerLevel * float64(level-1))
	multiplier := 1.0 + (sg.AttackPercentPerLevel * float64(level-1))
	return flat * multiplier
}

// CalculateDefense calculates defense at a given level (flat only, no percentage).
func (sg *StatGrowth) CalculateDefense(baseDefense float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseDefense + (sg.DefensePerLevel * float64(level-1))
}

// CalculateMagicPower calculates magic power at a given level (flat only).
func (sg *StatGrowth) CalculateMagicPower(baseMagicPower float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseMagicPower + (sg.MagicPowerPerLevel * float64(level-1))
}

// CalculateSpeed calculates movement speed at a given level (flat only).
func (sg *StatGrowth) CalculateSpeed(baseSpeed float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseSpeed + (sg.SpeedPerLevel * float64(level-1))
}

// GetClassStatGrowth returns the stat growth configuration for a character class.
// Each class has distinct growth rates that define their scaling patterns.
func GetClassStatGrowth(class CharacterClass) *StatGrowth {
	switch class {
	case ClassWarrior:
		// High HP growth, moderate attack, low mana
		return &StatGrowth{
			HPPerLevel:            15.0,
			ManaPerLevel:          2.0,
			AttackPerLevel:        1.5,
			DefensePerLevel:       1.2,
			MagicPowerPerLevel:    0.3,
			SpeedPerLevel:         0.5,
			HPPercentPerLevel:     0.05, // 5% more HP per level
			ManaPercentPerLevel:   0.02, // 2% more mana
			AttackPercentPerLevel: 0.03, // 3% more attack
		}

	case ClassRogue:
		// Moderate HP, high attack and speed, moderate mana
		return &StatGrowth{
			HPPerLevel:            10.0,
			ManaPerLevel:          4.0,
			AttackPerLevel:        2.0,
			DefensePerLevel:       0.8,
			MagicPowerPerLevel:    0.5,
			SpeedPerLevel:         1.5,
			HPPercentPerLevel:     0.03,
			ManaPercentPerLevel:   0.03,
			AttackPercentPerLevel: 0.05, // High attack scaling
		}

	case ClassMage:
		// Low HP, very high mana and magic power, low attack
		return &StatGrowth{
			HPPerLevel:            6.0,
			ManaPerLevel:          12.0,
			AttackPerLevel:        0.5,
			DefensePerLevel:       0.5,
			MagicPowerPerLevel:    2.5,
			SpeedPerLevel:         0.8,
			HPPercentPerLevel:     0.02,
			ManaPercentPerLevel:   0.08, // 8% more mana per level
			AttackPercentPerLevel: 0.01,
		}

	case ClassRanger:
		// Balanced HP, high attack, moderate mana
		return &StatGrowth{
			HPPerLevel:            12.0,
			ManaPerLevel:          5.0,
			AttackPerLevel:        1.8,
			DefensePerLevel:       1.0,
			MagicPowerPerLevel:    0.8,
			SpeedPerLevel:         1.2,
			HPPercentPerLevel:     0.04,
			ManaPercentPerLevel:   0.04,
			AttackPercentPerLevel: 0.04,
		}

	case ClassCleric:
		// High HP and mana, moderate defense, magic power focused
		return &StatGrowth{
			HPPerLevel:            13.0,
			ManaPerLevel:          10.0,
			AttackPerLevel:        1.0,
			DefensePerLevel:       1.1,
			MagicPowerPerLevel:    2.0,
			SpeedPerLevel:         0.8,
			HPPercentPerLevel:     0.04,
			ManaPercentPerLevel:   0.06,
			AttackPercentPerLevel: 0.02,
		}

	case ClassNecromancer:
		// Low HP, high mana and magic power, very low defense
		return &StatGrowth{
			HPPerLevel:            7.0,
			ManaPerLevel:          11.0,
			AttackPerLevel:        0.8,
			DefensePerLevel:       0.6,
			MagicPowerPerLevel:    2.2,
			SpeedPerLevel:         0.7,
			HPPercentPerLevel:     0.02,
			ManaPercentPerLevel:   0.07,
			AttackPercentPerLevel: 0.02,
		}

	default:
		// Generic balanced growth
		return &StatGrowth{
			HPPerLevel:            10.0,
			ManaPerLevel:          5.0,
			AttackPerLevel:        1.0,
			DefensePerLevel:       1.0,
			MagicPowerPerLevel:    1.0,
			SpeedPerLevel:         1.0,
			HPPercentPerLevel:     0.03,
			ManaPercentPerLevel:   0.03,
			AttackPercentPerLevel: 0.03,
		}
	}
}
