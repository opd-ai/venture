// Package engine provides the talent definitions for the talent point system.
// This file contains all available talents organized by category.
package engine

import "sync"

// talentRegistry holds all defined talents.
var talentRegistry = make(map[string]*TalentDefinition)

// categoryTalents groups talents by category for UI.
var categoryTalents = make(map[TalentCategory][]*TalentDefinition)

// talentMu protects talentRegistry and categoryTalents from concurrent
// access. The maps are populated once in init() and are effectively
// read-only during normal gameplay, but the mod system may call
// registerTalent() at runtime, so all accesses must hold the lock.
var talentMu sync.RWMutex

// init initializes the talent registry with all defined talents.
func init() {
	registerOffenseTalents()
	registerDefenseTalents()
	registerUtilityTalents()
	registerMasteryTalents()
}

// GetTalentDefinition returns a talent by ID.
func GetTalentDefinition(id string) *TalentDefinition {
	talentMu.RLock()
	defer talentMu.RUnlock()
	return talentRegistry[id]
}

// GetAllTalentDefinitions returns all defined talents.
func GetAllTalentDefinitions() []*TalentDefinition {
	talentMu.RLock()
	defer talentMu.RUnlock()
	result := make([]*TalentDefinition, 0, len(talentRegistry))
	for _, def := range talentRegistry {
		result = append(result, def)
	}
	return result
}

// GetTalentsByCategory returns all talents in a category.
func GetTalentsByCategory(category TalentCategory) []*TalentDefinition {
	talentMu.RLock()
	defer talentMu.RUnlock()
	return categoryTalents[category]
}

// registerTalent adds a talent to the registry.
func registerTalent(talent *TalentDefinition) {
	talentMu.Lock()
	talentRegistry[talent.ID] = talent
	categoryTalents[talent.Category] = append(categoryTalents[talent.Category], talent)
	talentMu.Unlock()
}

// registerOffenseTalents defines all offense talents.
func registerOffenseTalents() {
	// Tier 1 - No prerequisites
	registerTalent(&TalentDefinition{
		ID:            "offense_raw_power",
		Name:          "Raw Power",
		Description:   "Increases physical damage dealt.",
		Category:      TalentCategoryOffense,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			DamagePercent: 0.02, // +2% per rank = +10% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:            "offense_precision",
		Name:          "Precision",
		Description:   "Increases critical hit chance.",
		Category:      TalentCategoryOffense,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			CritChanceBonus: 0.01, // +1% per rank = +5% at max
		},
	})

	// Tier 2 - Requires 5 points in Offense
	registerTalent(&TalentDefinition{
		ID:                           "offense_brutal_strikes",
		Name:                         "Brutal Strikes",
		Description:                  "Increases critical hit damage.",
		Category:                     TalentCategoryOffense,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			CritDamageBonus: 0.05, // +5% crit damage per rank = +25% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                           "offense_arcane_fury",
		Name:                         "Arcane Fury",
		Description:                  "Increases magic damage dealt.",
		Category:                     TalentCategoryOffense,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			MagicPowerPercent: 0.02, // +2% per rank = +10% at max
		},
	})

	// Tier 3 - Requires specific talents
	registerTalent(&TalentDefinition{
		ID:                   "offense_bloodlust",
		Name:                 "Bloodlust",
		Description:          "Attacks restore health based on damage dealt.",
		Category:             TalentCategoryOffense,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "offense_brutal_strikes",
		BonusPerRank: TalentBonus{
			LifestealPercent: 0.02, // +2% lifesteal per rank = +6% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                   "offense_spell_surge",
		Name:                 "Spell Surge",
		Description:          "Reduces spell cooldowns.",
		Category:             TalentCategoryOffense,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "offense_arcane_fury",
		BonusPerRank: TalentBonus{
			CooldownReduction: 0.05, // +5% CDR per rank = +15% at max
		},
	})

	// Tier 4 - Capstone
	registerTalent(&TalentDefinition{
		ID:                           "offense_devastation",
		Name:                         "Devastation",
		Description:                  "Massively increases all damage output.",
		Category:                     TalentCategoryOffense,
		MaxRanks:                     1,
		RequiredLevel:                20,
		PrerequisitePointsInCategory: 15,
		BonusPerRank: TalentBonus{
			DamagePercent:     0.08,
			MagicPowerPercent: 0.08,
			CritDamageBonus:   0.15,
		},
	})
}

// registerDefenseTalents defines all defense talents.
func registerDefenseTalents() {
	// Tier 1 - No prerequisites
	registerTalent(&TalentDefinition{
		ID:            "defense_fortitude",
		Name:          "Fortitude",
		Description:   "Increases maximum health.",
		Category:      TalentCategoryDefense,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			HealthPercent: 0.03, // +3% per rank = +15% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:            "defense_tough_skin",
		Name:          "Tough Skin",
		Description:   "Increases physical defense.",
		Category:      TalentCategoryDefense,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			DefensePercent: 0.02, // +2% per rank = +10% at max
		},
	})

	// Tier 2 - Requires 5 points in Defense
	registerTalent(&TalentDefinition{
		ID:                           "defense_shield_mastery",
		Name:                         "Shield Mastery",
		Description:                  "Increases block chance.",
		Category:                     TalentCategoryDefense,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			BlockChanceBonus: 0.02, // +2% per rank = +10% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                           "defense_arcane_barrier",
		Name:                         "Arcane Barrier",
		Description:                  "Increases magic defense.",
		Category:                     TalentCategoryDefense,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			MagicDefensePercent: 0.02, // +2% per rank = +10% at max
		},
	})

	// Tier 3 - Requires specific talents
	registerTalent(&TalentDefinition{
		ID:                   "defense_resilience",
		Name:                 "Resilience",
		Description:          "Increases healing received from all sources.",
		Category:             TalentCategoryDefense,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "defense_fortitude",
		BonusPerRank: TalentBonus{
			HealingReceivedBonus: 0.08, // +8% per rank = +24% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                   "defense_evasion",
		Name:                 "Evasion",
		Description:          "Chance to dodge attacks completely.",
		Category:             TalentCategoryDefense,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "defense_tough_skin",
		BonusPerRank: TalentBonus{
			DodgeChanceBonus: 0.02, // +2% per rank = +6% at max
		},
	})

	// Tier 4 - Capstone
	registerTalent(&TalentDefinition{
		ID:                           "defense_indomitable",
		Name:                         "Indomitable",
		Description:                  "Massively increases survivability.",
		Category:                     TalentCategoryDefense,
		MaxRanks:                     1,
		RequiredLevel:                20,
		PrerequisitePointsInCategory: 15,
		BonusPerRank: TalentBonus{
			HealthPercent:        0.10,
			DefensePercent:       0.10,
			MagicDefensePercent:  0.10,
			StatusResistBonus:    0.15,
			HealingReceivedBonus: 0.10,
		},
	})
}

// registerUtilityTalents defines all utility talents.
func registerUtilityTalents() {
	// Tier 1 - No prerequisites
	registerTalent(&TalentDefinition{
		ID:            "utility_mana_pool",
		Name:          "Mana Pool",
		Description:   "Increases maximum mana.",
		Category:      TalentCategoryUtility,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			ManaPercent: 0.04, // +4% per rank = +20% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:            "utility_swift_feet",
		Name:          "Swift Feet",
		Description:   "Increases movement speed.",
		Category:      TalentCategoryUtility,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			SpeedPercent: 0.02, // +2% per rank = +10% at max
		},
	})

	// Tier 2 - Requires 5 points in Utility
	registerTalent(&TalentDefinition{
		ID:                           "utility_efficient_casting",
		Name:                         "Efficient Casting",
		Description:                  "Reduces mana cost of abilities.",
		Category:                     TalentCategoryUtility,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			ManaCostReduction: 0.03, // +3% per rank = +15% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                           "utility_treasure_hunter",
		Name:                         "Treasure Hunter",
		Description:                  "Increases gold gained from all sources.",
		Category:                     TalentCategoryUtility,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			GoldBonusPercent: 0.04, // +4% per rank = +20% at max
		},
	})

	// Tier 3 - Requires specific talents
	registerTalent(&TalentDefinition{
		ID:                   "utility_arcane_wisdom",
		Name:                 "Arcane Wisdom",
		Description:          "Further reduces cooldowns and mana costs.",
		Category:             TalentCategoryUtility,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "utility_efficient_casting",
		BonusPerRank: TalentBonus{
			CooldownReduction: 0.03,
			ManaCostReduction: 0.03,
		},
	})
	registerTalent(&TalentDefinition{
		ID:                   "utility_veteran",
		Name:                 "Veteran",
		Description:          "Increases experience gained.",
		Category:             TalentCategoryUtility,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "utility_treasure_hunter",
		BonusPerRank: TalentBonus{
			XPBonusPercent: 0.05, // +5% per rank = +15% at max
		},
	})

	// Tier 4 - Capstone
	registerTalent(&TalentDefinition{
		ID:                           "utility_resourceful",
		Name:                         "Resourceful",
		Description:                  "Maximizes resource efficiency and rewards.",
		Category:                     TalentCategoryUtility,
		MaxRanks:                     1,
		RequiredLevel:                20,
		PrerequisitePointsInCategory: 15,
		BonusPerRank: TalentBonus{
			ManaPercent:       0.15,
			ManaCostReduction: 0.10,
			CooldownReduction: 0.10,
			XPBonusPercent:    0.10,
			GoldBonusPercent:  0.10,
		},
	})
}

// registerMasteryTalents defines all mastery talents.
func registerMasteryTalents() {
	// Tier 1 - No prerequisites
	registerTalent(&TalentDefinition{
		ID:            "mastery_status_resist",
		Name:          "Iron Will",
		Description:   "Increases resistance to status effects.",
		Category:      TalentCategoryMastery,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			StatusResistBonus: 0.04, // +4% per rank = +20% at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:            "mastery_combat_prowess",
		Name:          "Combat Prowess",
		Description:   "Increases both damage and defense slightly.",
		Category:      TalentCategoryMastery,
		MaxRanks:      5,
		RequiredLevel: 1,
		BonusPerRank: TalentBonus{
			DamagePercent:  0.01,
			DefensePercent: 0.01,
		},
	})

	// Tier 2 - Requires 5 points in Mastery
	registerTalent(&TalentDefinition{
		ID:                           "mastery_battle_mage",
		Name:                         "Battle Mage",
		Description:                  "Increases both physical and magical power.",
		Category:                     TalentCategoryMastery,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			FlatDamage:     3.0, // +3 flat damage per rank = +15 at max
			FlatMagicPower: 3.0, // +3 flat magic power per rank = +15 at max
		},
	})
	registerTalent(&TalentDefinition{
		ID:                           "mastery_constitution",
		Name:                         "Constitution",
		Description:                  "Increases both health and mana.",
		Category:                     TalentCategoryMastery,
		MaxRanks:                     5,
		RequiredLevel:                5,
		PrerequisitePointsInCategory: 5,
		BonusPerRank: TalentBonus{
			FlatHealth: 10.0, // +10 flat HP per rank = +50 at max
			FlatMana:   8.0,  // +8 flat mana per rank = +40 at max
		},
	})

	// Tier 3 - Requires specific talents
	registerTalent(&TalentDefinition{
		ID:                   "mastery_dual_wield",
		Name:                 "Dual Wield Mastery",
		Description:          "Enhances attack speed and critical chance.",
		Category:             TalentCategoryMastery,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "mastery_combat_prowess",
		BonusPerRank: TalentBonus{
			SpeedPercent:    0.03,
			CritChanceBonus: 0.02,
		},
	})
	registerTalent(&TalentDefinition{
		ID:                   "mastery_elemental_attunement",
		Name:                 "Elemental Attunement",
		Description:          "Enhances magic damage and reduces spell costs.",
		Category:             TalentCategoryMastery,
		MaxRanks:             3,
		RequiredLevel:        10,
		PrerequisiteTalentID: "mastery_battle_mage",
		BonusPerRank: TalentBonus{
			MagicPowerPercent: 0.03,
			ManaCostReduction: 0.02,
		},
	})

	// Tier 4 - Capstone
	registerTalent(&TalentDefinition{
		ID:                           "mastery_transcendence",
		Name:                         "Transcendence",
		Description:                  "Ultimate mastery of all aspects of combat.",
		Category:                     TalentCategoryMastery,
		MaxRanks:                     1,
		RequiredLevel:                20,
		PrerequisitePointsInCategory: 15,
		BonusPerRank: TalentBonus{
			HealthPercent:     0.05,
			ManaPercent:       0.05,
			DamagePercent:     0.05,
			DefensePercent:    0.05,
			MagicPowerPercent: 0.05,
			CritChanceBonus:   0.03,
			CritDamageBonus:   0.10,
			LifestealPercent:  0.03,
		},
	})
}
