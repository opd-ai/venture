// Constants for the advanced class system.
//
// This file contains all class identifiers, prestige class identifiers,
// talent identifiers, and category constants used throughout the package.
// Code relocated from: types.go
package advanced

// ClassID identifies a base character class
type ClassID string

const (
	// Warrior classes
	ClassWarrior   ClassID = "warrior"
	ClassBerserker ClassID = "berserker"
	ClassPaladin   ClassID = "paladin"
	ClassKnight    ClassID = "knight"

	// Rogue classes
	ClassRogue    ClassID = "rogue"
	ClassAssassin ClassID = "assassin"
	ClassRanger   ClassID = "ranger"
	ClassNinja    ClassID = "ninja"

	// Mage classes
	ClassMage         ClassID = "mage"
	ClassElementalist ClassID = "elementalist"
	ClassNecromancer  ClassID = "necromancer"
	ClassEnchanter    ClassID = "enchanter"

	// Support classes
	ClassCleric ClassID = "cleric"
	ClassBard   ClassID = "bard"
	ClassDruid  ClassID = "druid"
)

// PrestigeClassID identifies a prestige class (unlocked at level 20+)
type PrestigeClassID string

const (
	// Warrior prestige classes
	PrestigeBladeMaster  PrestigeClassID = "blade_master"
	PrestigeChampion     PrestigeClassID = "champion"
	PrestigeDragonKnight PrestigeClassID = "dragon_knight"
	PrestigeDreadnought  PrestigeClassID = "dreadnought"

	// Rogue prestige classes
	PrestigeShadowDancer PrestigeClassID = "shadow_dancer"
	PrestigeDeadeye      PrestigeClassID = "deadeye"
	PrestigeDuelist      PrestigeClassID = "duelist"
	PrestigePhantom      PrestigeClassID = "phantom"

	// Mage prestige classes
	PrestigeArchmage      PrestigeClassID = "archmage"
	PrestigeSoulReaper    PrestigeClassID = "soul_reaper"
	PrestigeElementalLord PrestigeClassID = "elemental_lord"
	PrestigeTimeMage      PrestigeClassID = "time_mage"

	// Support prestige classes
	PrestigeHighPriest PrestigeClassID = "high_priest"
	PrestigeMaestro    PrestigeClassID = "maestro"
	PrestigeArchdruid  PrestigeClassID = "archdruid"
	PrestigeOracle     PrestigeClassID = "oracle"

	// Hybrid prestige classes
	PrestigeBattlemage PrestigeClassID = "battlemage"
	PrestigeSpellblade PrestigeClassID = "spellblade"
	PrestigeVoidwalker PrestigeClassID = "voidwalker"
	PrestigeRunesmith  PrestigeClassID = "runesmith"
)

// TalentID identifies a talent within a talent tree
type TalentID string

// TalentCategory groups related talents
type TalentCategory string

const (
	CategoryOffensive TalentCategory = "offensive"
	CategoryDefensive TalentCategory = "defensive"
	CategoryUtility   TalentCategory = "utility"
)
