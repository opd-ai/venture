package advanced

import (
	"fmt"
	"image/color"
)

// GetClassDefinition returns the definition for a base class
func GetClassDefinition(id ClassID) (ClassDefinition, error) {
	def, exists := classDefinitions[id]
	if !exists {
		return ClassDefinition{}, fmt.Errorf("unknown class: %s", id)
	}
	return def, nil
}

// GetPrestigeClassDefinition returns the definition for a prestige class
func GetPrestigeClassDefinition(id PrestigeClassID) (PrestigeClassDefinition, error) {
	def, exists := prestigeDefinitions[id]
	if !exists {
		return PrestigeClassDefinition{}, fmt.Errorf("unknown prestige class: %s", id)
	}
	return def, nil
}

// GetAllClasses returns all available base classes
func GetAllClasses() []ClassDefinition {
	classes := make([]ClassDefinition, 0, len(classDefinitions))
	for _, def := range classDefinitions {
		classes = append(classes, def)
	}
	return classes
}

// GetAllPrestigeClasses returns all available prestige classes
func GetAllPrestigeClasses() []PrestigeClassDefinition {
	classes := make([]PrestigeClassDefinition, 0, len(prestigeDefinitions))
	for _, def := range prestigeDefinitions {
		classes = append(classes, def)
	}
	return classes
}

// Base class definitions (15 total)
var classDefinitions = map[ClassID]ClassDefinition{
	ClassWarrior: {
		ID:          ClassWarrior,
		Name:        "Warrior",
		Description: "Master of melee combat and physical prowess",
		BaseStats: StatBonuses{
			Health:   100,
			Stamina:  50,
			Strength: 10,
			Defense:  5,
		},
		Color: color.RGBA{R: 200, G: 50, B: 50, A: 255},
	},
	ClassBerserker: {
		ID:          ClassBerserker,
		Name:        "Berserker",
		Description: "Fierce warrior sacrificing defense for raw damage",
		BaseStats: StatBonuses{
			Health:     80,
			Stamina:    60,
			Strength:   15,
			CritDamage: 0.5,
		},
		Color: color.RGBA{R: 180, G: 30, B: 30, A: 255},
	},
	ClassPaladin: {
		ID:          ClassPaladin,
		Name:        "Paladin",
		Description: "Holy warrior combining strength with divine magic",
		BaseStats: StatBonuses{
			Health:       120,
			Mana:         40,
			Strength:     8,
			Wisdom:       6,
			Defense:      8,
			MagicDefense: 5,
		},
		Color: color.RGBA{R: 220, G: 180, B: 50, A: 255},
	},
	ClassKnight: {
		ID:          ClassKnight,
		Name:        "Knight",
		Description: "Heavily armored defender specializing in protection",
		BaseStats: StatBonuses{
			Health:       150,
			Stamina:      40,
			Strength:     7,
			Defense:      12,
			MagicDefense: 3,
		},
		Color: color.RGBA{R: 100, G: 100, B: 150, A: 255},
	},
	ClassRogue: {
		ID:          ClassRogue,
		Name:        "Rogue",
		Description: "Agile combatant focused on precision and evasion",
		BaseStats: StatBonuses{
			Health:     70,
			Stamina:    70,
			Dexterity:  12,
			CritChance: 0.1,
			Speed:      1.2,
		},
		Color: color.RGBA{R: 80, G: 80, B: 80, A: 255},
	},
	ClassAssassin: {
		ID:          ClassAssassin,
		Name:        "Assassin",
		Description: "Master of stealth and deadly critical strikes",
		BaseStats: StatBonuses{
			Health:     60,
			Stamina:    60,
			Dexterity:  15,
			CritChance: 0.2,
			CritDamage: 1.0,
			Speed:      1.3,
		},
		Color: color.RGBA{R: 60, G: 40, B: 80, A: 255},
	},
	ClassRanger: {
		ID:          ClassRanger,
		Name:        "Ranger",
		Description: "Wilderness expert skilled in ranged combat and survival",
		BaseStats: StatBonuses{
			Health:    80,
			Stamina:   80,
			Dexterity: 10,
			Wisdom:    5,
			Speed:     1.1,
		},
		Color: color.RGBA{R: 100, G: 150, B: 80, A: 255},
	},
	ClassNinja: {
		ID:          ClassNinja,
		Name:        "Ninja",
		Description: "Shadow warrior blending martial arts with deception",
		BaseStats: StatBonuses{
			Health:     65,
			Mana:       30,
			Dexterity:  13,
			CritChance: 0.15,
			Speed:      1.4,
		},
		Color: color.RGBA{R: 50, G: 50, B: 100, A: 255},
	},
	ClassMage: {
		ID:          ClassMage,
		Name:        "Mage",
		Description: "Arcane spellcaster wielding raw magical power",
		BaseStats: StatBonuses{
			Health:       50,
			Mana:         100,
			Intelligence: 12,
			MagicDefense: 5,
		},
		Color: color.RGBA{R: 100, G: 100, B: 200, A: 255},
	},
	ClassElementalist: {
		ID:          ClassElementalist,
		Name:        "Elementalist",
		Description: "Master of elemental forces",
		BaseStats: StatBonuses{
			Health:       50,
			Mana:         110,
			Intelligence: 15,
			Wisdom:       5,
		},
		Color: color.RGBA{R: 80, G: 120, B: 200, A: 255},
	},
	ClassNecromancer: {
		ID:          ClassNecromancer,
		Name:        "Necromancer",
		Description: "Dark mage commanding death and decay",
		BaseStats: StatBonuses{
			Health:       60,
			Mana:         90,
			Intelligence: 13,
			Wisdom:       6,
		},
		Color: color.RGBA{R: 100, G: 50, B: 150, A: 255},
	},
	ClassEnchanter: {
		ID:          ClassEnchanter,
		Name:        "Enchanter",
		Description: "Spellcaster specializing in buffs and debuffs",
		BaseStats: StatBonuses{
			Health:       55,
			Mana:         95,
			Intelligence: 10,
			Wisdom:       8,
			Charisma:     5,
		},
		Color: color.RGBA{R: 150, G: 100, B: 200, A: 255},
	},
	ClassCleric: {
		ID:          ClassCleric,
		Name:        "Cleric",
		Description: "Divine healer and support specialist",
		BaseStats: StatBonuses{
			Health:       90,
			Mana:         80,
			Wisdom:       12,
			Charisma:     5,
			MagicDefense: 6,
		},
		Color: color.RGBA{R: 220, G: 220, B: 150, A: 255},
	},
	ClassBard: {
		ID:          ClassBard,
		Name:        "Bard",
		Description: "Charismatic performer buffing allies through music",
		BaseStats: StatBonuses{
			Health:    75,
			Mana:      70,
			Dexterity: 6,
			Charisma:  12,
			Speed:     1.1,
		},
		Color: color.RGBA{R: 200, G: 150, B: 100, A: 255},
	},
	ClassDruid: {
		ID:          ClassDruid,
		Name:        "Druid",
		Description: "Nature caster balancing magic and physical forms",
		BaseStats: StatBonuses{
			Health:       85,
			Mana:         75,
			Wisdom:       10,
			Intelligence: 7,
			Defense:      4,
		},
		Color: color.RGBA{R: 120, G: 180, B: 80, A: 255},
	},
}

// Prestige class definitions (20 total)
var prestigeDefinitions = map[PrestigeClassID]PrestigeClassDefinition{
	PrestigeBladeMaster: {
		ID:          PrestigeBladeMaster,
		Name:        "Blade Master",
		Description: "Supreme warrior with unmatched weapon mastery",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassWarrior, ClassBerserker},
			MinPrimaryStat:  15,
		},
		BaseStats: StatBonuses{
			Strength:   20,
			CritChance: 0.15,
			CritDamage: 0.5,
		},
		Color: color.RGBA{R: 220, G: 70, B: 70, A: 255},
	},
	PrestigeChampion: {
		ID:          PrestigeChampion,
		Name:        "Champion",
		Description: "Legendary hero inspiring allies with presence",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassPaladin, ClassKnight},
			MinPrimaryStat:  15,
		},
		BaseStats: StatBonuses{
			Health:   150,
			Defense:  15,
			Charisma: 10,
		},
		Color: color.RGBA{R: 240, G: 200, B: 70, A: 255},
	},
	PrestigeDragonKnight: {
		ID:          PrestigeDragonKnight,
		Name:        "Dragon Knight",
		Description: "Warrior infused with draconic power",
		Requirements: PrestigeRequirements{
			MinLevel:         20,
			RequiredPrimary:  []ClassID{ClassWarrior, ClassPaladin},
			MinSecondaryStat: 10,
		},
		BaseStats: StatBonuses{
			Health:       130,
			Strength:     15,
			Defense:      10,
			MagicDefense: 10,
		},
		Color: color.RGBA{R: 180, G: 50, B: 50, A: 255},
	},
	PrestigeDreadnought: {
		ID:          PrestigeDreadnought,
		Name:        "Dreadnought",
		Description: "Unstoppable armored juggernaut",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassKnight},
			MinPrimaryStat:  18,
		},
		BaseStats: StatBonuses{
			Health:       200,
			Defense:      20,
			MagicDefense: 8,
			Speed:        -0.2,
		},
		Color: color.RGBA{R: 120, G: 120, B: 170, A: 255},
	},
	PrestigeShadowDancer: {
		ID:          PrestigeShadowDancer,
		Name:        "Shadow Dancer",
		Description: "Phantom assassin moving through shadows",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassRogue, ClassAssassin, ClassNinja},
			MinPrimaryStat:  15,
		},
		BaseStats: StatBonuses{
			Dexterity:  20,
			CritChance: 0.25,
			Speed:      1.5,
		},
		Color: color.RGBA{R: 80, G: 60, B: 100, A: 255},
	},
	PrestigeDeadeye: {
		ID:          PrestigeDeadeye,
		Name:        "Deadeye",
		Description: "Unmatched sharpshooter with perfect precision",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassRanger, ClassRogue},
			MinPrimaryStat:  15,
		},
		BaseStats: StatBonuses{
			Dexterity:  18,
			CritChance: 0.3,
			CritDamage: 1.5,
		},
		Color: color.RGBA{R: 120, G: 170, B: 100, A: 255},
	},
	PrestigeDuelist: {
		ID:          PrestigeDuelist,
		Name:        "Duelist",
		Description: "Master of single combat and parrying",
		Requirements: PrestigeRequirements{
			MinLevel:         20,
			RequiredPrimary:  []ClassID{ClassRogue, ClassWarrior},
			MinPrimaryStat:   14,
			MinSecondaryStat: 12,
		},
		BaseStats: StatBonuses{
			Dexterity:  15,
			Strength:   10,
			CritChance: 0.2,
			Speed:      1.2,
		},
		Color: color.RGBA{R: 160, G: 100, B: 100, A: 255},
	},
	PrestigePhantom: {
		ID:          PrestigePhantom,
		Name:        "Phantom",
		Description: "Invisible killer striking from nowhere",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassAssassin, ClassNinja},
			MinPrimaryStat:  16,
		},
		BaseStats: StatBonuses{
			Dexterity:  22,
			CritDamage: 2.0,
			Speed:      1.6,
		},
		Color: color.RGBA{R: 70, G: 70, B: 120, A: 255},
	},
	PrestigeArchmage: {
		ID:          PrestigeArchmage,
		Name:        "Archmage",
		Description: "Supreme master of all magical schools",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassMage, ClassElementalist},
			MinPrimaryStat:  18,
		},
		BaseStats: StatBonuses{
			Mana:         150,
			Intelligence: 25,
			Wisdom:       10,
		},
		Color: color.RGBA{R: 120, G: 120, B: 220, A: 255},
	},
	PrestigeSoulReaper: {
		ID:          PrestigeSoulReaper,
		Name:        "Soul Reaper",
		Description: "Dark mage harvesting souls for power",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassNecromancer},
			MinPrimaryStat:  16,
		},
		BaseStats: StatBonuses{
			Mana:         130,
			Intelligence: 20,
			Wisdom:       8,
			CritDamage:   0.8,
		},
		Color: color.RGBA{R: 120, G: 70, B: 170, A: 255},
	},
	PrestigeElementalLord: {
		ID:          PrestigeElementalLord,
		Name:        "Elemental Lord",
		Description: "Avatar commanding all elemental forces",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassElementalist, ClassMage},
			MinPrimaryStat:  17,
		},
		BaseStats: StatBonuses{
			Mana:         140,
			Intelligence: 22,
			Wisdom:       7,
		},
		Color: color.RGBA{R: 100, G: 140, B: 220, A: 255},
	},
	PrestigeTimeMage: {
		ID:          PrestigeTimeMage,
		Name:        "Time Mage",
		Description: "Chronomancer manipulating time itself",
		Requirements: PrestigeRequirements{
			MinLevel:         20,
			RequiredPrimary:  []ClassID{ClassMage, ClassEnchanter},
			MinPrimaryStat:   15,
			MinSecondaryStat: 12,
		},
		BaseStats: StatBonuses{
			Mana:         120,
			Intelligence: 18,
			Wisdom:       12,
			Speed:        1.3,
		},
		Color: color.RGBA{R: 180, G: 120, B: 220, A: 255},
	},
	PrestigeHighPriest: {
		ID:          PrestigeHighPriest,
		Name:        "High Priest",
		Description: "Divine conduit with ultimate healing power",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassCleric},
			MinPrimaryStat:  18,
		},
		BaseStats: StatBonuses{
			Mana:         140,
			Wisdom:       20,
			Charisma:     10,
			MagicDefense: 12,
		},
		Color: color.RGBA{R: 240, G: 240, B: 170, A: 255},
	},
	PrestigeMaestro: {
		ID:          PrestigeMaestro,
		Name:        "Maestro",
		Description: "Musical virtuoso with overwhelming charisma",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassBard},
			MinPrimaryStat:  16,
		},
		BaseStats: StatBonuses{
			Mana:     110,
			Charisma: 22,
			Wisdom:   8,
		},
		Color: color.RGBA{R: 220, G: 170, B: 120, A: 255},
	},
	PrestigeArchdruid: {
		ID:          PrestigeArchdruid,
		Name:        "Archdruid",
		Description: "Nature's champion with ultimate shapeshifting",
		Requirements: PrestigeRequirements{
			MinLevel:        20,
			RequiredPrimary: []ClassID{ClassDruid},
			MinPrimaryStat:  16,
		},
		BaseStats: StatBonuses{
			Health:       120,
			Mana:         110,
			Wisdom:       18,
			Intelligence: 10,
		},
		Color: color.RGBA{R: 140, G: 200, B: 100, A: 255},
	},
	PrestigeOracle: {
		ID:          PrestigeOracle,
		Name:        "Oracle",
		Description: "Seer perceiving past, present, and future",
		Requirements: PrestigeRequirements{
			MinLevel:         20,
			RequiredPrimary:  []ClassID{ClassCleric, ClassMage},
			MinPrimaryStat:   14,
			MinSecondaryStat: 14,
		},
		BaseStats: StatBonuses{
			Mana:         130,
			Wisdom:       18,
			Intelligence: 12,
			Charisma:     8,
		},
		Color: color.RGBA{R: 200, G: 180, B: 220, A: 255},
	},
	PrestigeBattlemage: {
		ID:          PrestigeBattlemage,
		Name:        "Battlemage",
		Description: "Warrior-mage blending sword and spell",
		Requirements: PrestigeRequirements{
			MinLevel:          20,
			RequiredPrimary:   []ClassID{ClassWarrior, ClassPaladin},
			RequiredSecondary: []ClassID{ClassMage, ClassElementalist},
			MinPrimaryStat:    12,
			MinSecondaryStat:  12,
		},
		BaseStats: StatBonuses{
			Health:       100,
			Mana:         100,
			Strength:     12,
			Intelligence: 12,
		},
		Color: color.RGBA{R: 180, G: 100, B: 180, A: 255},
	},
	PrestigeSpellblade: {
		ID:          PrestigeSpellblade,
		Name:        "Spellblade",
		Description: "Agile fighter infusing weapons with magic",
		Requirements: PrestigeRequirements{
			MinLevel:          20,
			RequiredPrimary:   []ClassID{ClassRogue, ClassRanger},
			RequiredSecondary: []ClassID{ClassMage, ClassEnchanter},
			MinPrimaryStat:    12,
			MinSecondaryStat:  12,
		},
		BaseStats: StatBonuses{
			Mana:         90,
			Dexterity:    14,
			Intelligence: 10,
			Speed:        1.2,
		},
		Color: color.RGBA{R: 140, G: 140, B: 200, A: 255},
	},
	PrestigeVoidwalker: {
		ID:          PrestigeVoidwalker,
		Name:        "Voidwalker",
		Description: "Dark adept wielding void energies",
		Requirements: PrestigeRequirements{
			MinLevel:         20,
			RequiredPrimary:  []ClassID{ClassNecromancer, ClassAssassin},
			MinPrimaryStat:   14,
			MinSecondaryStat: 10,
		},
		BaseStats: StatBonuses{
			Mana:         110,
			Intelligence: 15,
			Dexterity:    10,
			CritDamage:   1.0,
		},
		Color: color.RGBA{R: 100, G: 60, B: 140, A: 255},
	},
	PrestigeRunesmith: {
		ID:          PrestigeRunesmith,
		Name:        "Runesmith",
		Description: "Craftsman enchanting items with powerful runes",
		Requirements: PrestigeRequirements{
			MinLevel:          20,
			RequiredPrimary:   []ClassID{ClassWarrior, ClassKnight},
			RequiredSecondary: []ClassID{ClassEnchanter, ClassCleric},
			MinPrimaryStat:    10,
			MinSecondaryStat:  14,
		},
		BaseStats: StatBonuses{
			Health:       110,
			Mana:         90,
			Strength:     10,
			Intelligence: 12,
			Defense:      8,
		},
		Color: color.RGBA{R: 160, G: 140, B: 120, A: 255},
	},
}
