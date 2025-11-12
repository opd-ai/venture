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

	// Hybrid classes - balanced between parent classes
	case ClassBattlemage: // Warrior + Mage
		return &StatGrowth{
			HPPerLevel:            10.5, // Average of Warrior and Mage
			ManaPerLevel:          7.0,
			AttackPerLevel:        1.0,
			DefensePerLevel:       0.85,
			MagicPowerPerLevel:    1.4,
			SpeedPerLevel:         0.65,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.05,
			AttackPercentPerLevel: 0.02,
		}

	case ClassSpellblade: // Rogue + Mage
		return &StatGrowth{
			HPPerLevel:            8.0,
			ManaPerLevel:          8.0,
			AttackPerLevel:        1.25,
			DefensePerLevel:       0.65,
			MagicPowerPerLevel:    1.5,
			SpeedPerLevel:         1.15,
			HPPercentPerLevel:     0.025,
			ManaPercentPerLevel:   0.055,
			AttackPercentPerLevel: 0.03,
		}

	case ClassPaladin: // Warrior + Cleric
		return &StatGrowth{
			HPPerLevel:            14.0,
			ManaPerLevel:          6.0,
			AttackPerLevel:        1.25,
			DefensePerLevel:       1.15,
			MagicPowerPerLevel:    1.15,
			SpeedPerLevel:         0.65,
			HPPercentPerLevel:     0.045,
			ManaPercentPerLevel:   0.04,
			AttackPercentPerLevel: 0.025,
		}

	case ClassMonk: // Rogue + Cleric
		return &StatGrowth{
			HPPerLevel:            11.5,
			ManaPerLevel:          7.0,
			AttackPerLevel:        1.5,
			DefensePerLevel:       0.95,
			MagicPowerPerLevel:    1.25,
			SpeedPerLevel:         1.15,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.045,
			AttackPercentPerLevel: 0.035,
		}

	case ClassDeathKnight: // Warrior + Necromancer
		return &StatGrowth{
			HPPerLevel:            11.0,
			ManaPerLevel:          6.5,
			AttackPerLevel:        1.15,
			DefensePerLevel:       0.9,
			MagicPowerPerLevel:    1.25,
			SpeedPerLevel:         0.6,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.045,
			AttackPercentPerLevel: 0.025,
		}

	case ClassWitchHunter: // Ranger + Cleric
		return &StatGrowth{
			HPPerLevel:            12.5,
			ManaPerLevel:          7.5,
			AttackPerLevel:        1.4,
			DefensePerLevel:       1.05,
			MagicPowerPerLevel:    1.4,
			SpeedPerLevel:         1.0,
			HPPercentPerLevel:     0.04,
			ManaPercentPerLevel:   0.05,
			AttackPercentPerLevel: 0.03,
		}

	case ClassBeastlord: // Warrior + Ranger
		return &StatGrowth{
			HPPerLevel:            13.5,
			ManaPerLevel:          3.5,
			AttackPerLevel:        1.65,
			DefensePerLevel:       1.1,
			MagicPowerPerLevel:    0.55,
			SpeedPerLevel:         0.85,
			HPPercentPerLevel:     0.045,
			ManaPercentPerLevel:   0.03,
			AttackPercentPerLevel: 0.035,
		}

	case ClassArcaneArcher: // Ranger + Mage
		return &StatGrowth{
			HPPerLevel:            9.0,
			ManaPerLevel:          8.5,
			AttackPerLevel:        1.15,
			DefensePerLevel:       0.75,
			MagicPowerPerLevel:    1.65,
			SpeedPerLevel:         1.0,
			HPPercentPerLevel:     0.03,
			ManaPercentPerLevel:   0.06,
			AttackPercentPerLevel: 0.025,
		}

	case ClassShadowPriest: // Rogue + Necromancer
		return &StatGrowth{
			HPPerLevel:            8.5,
			ManaPerLevel:          7.5,
			AttackPerLevel:        1.4,
			DefensePerLevel:       0.7,
			MagicPowerPerLevel:    1.35,
			SpeedPerLevel:         1.1,
			HPPercentPerLevel:     0.025,
			ManaPercentPerLevel:   0.05,
			AttackPercentPerLevel: 0.035,
		}

	case ClassDruid: // Ranger + Mage (nature themed)
		return &StatGrowth{
			HPPerLevel:            9.0,
			ManaPerLevel:          8.5,
			AttackPerLevel:        1.15,
			DefensePerLevel:       0.75,
			MagicPowerPerLevel:    1.65,
			SpeedPerLevel:         1.0,
			HPPercentPerLevel:     0.03,
			ManaPercentPerLevel:   0.06,
			AttackPercentPerLevel: 0.025,
		}

	case ClassInquisitor: // Cleric + Rogue
		return &StatGrowth{
			HPPerLevel:            11.5,
			ManaPerLevel:          7.0,
			AttackPerLevel:        1.5,
			DefensePerLevel:       0.95,
			MagicPowerPerLevel:    1.25,
			SpeedPerLevel:         1.15,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.045,
			AttackPercentPerLevel: 0.035,
		}

	case ClassBloodKnight: // Warrior + Necromancer (blood themed)
		return &StatGrowth{
			HPPerLevel:            11.0,
			ManaPerLevel:          6.5,
			AttackPerLevel:        1.15,
			DefensePerLevel:       0.9,
			MagicPowerPerLevel:    1.25,
			SpeedPerLevel:         0.6,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.045,
			AttackPercentPerLevel: 0.025,
		}

	case ClassMystic: // Mage + Cleric
		return &StatGrowth{
			HPPerLevel:            9.5,
			ManaPerLevel:          11.0,
			AttackPerLevel:        0.75,
			DefensePerLevel:       0.8,
			MagicPowerPerLevel:    2.25,
			SpeedPerLevel:         0.8,
			HPPercentPerLevel:     0.03,
			ManaPercentPerLevel:   0.07,
			AttackPercentPerLevel: 0.015,
		}

	case ClassWarlock: // Mage + Necromancer
		return &StatGrowth{
			HPPerLevel:            6.5,
			ManaPerLevel:          11.5,
			AttackPerLevel:        0.65,
			DefensePerLevel:       0.55,
			MagicPowerPerLevel:    2.35,
			SpeedPerLevel:         0.75,
			HPPercentPerLevel:     0.02,
			ManaPercentPerLevel:   0.075,
			AttackPercentPerLevel: 0.015,
		}

	case ClassNinja: // Rogue + Ranger
		return &StatGrowth{
			HPPerLevel:            11.0,
			ManaPerLevel:          4.5,
			AttackPerLevel:        1.9,
			DefensePerLevel:       0.9,
			MagicPowerPerLevel:    0.65,
			SpeedPerLevel:         1.35,
			HPPercentPerLevel:     0.035,
			ManaPercentPerLevel:   0.035,
			AttackPercentPerLevel: 0.045,
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
