// types.go defines configuration data structures.
// This file contains the Config type used for validation input.
//
// Package config provides configuration validation utilities for server and client.
package config

// Config holds configuration values to validate.
// Originally from: validator.go
type Config struct {
	// Port is the server port number in string format.
	// Valid range: "1024" to "65535" (ports < 1024 require root privileges)
	Port string

	// MaxPlayers is the maximum number of concurrent players.
	// Valid range: 1 to 100 (performance degrades above 100)
	MaxPlayers int

	// ValidateMaxPlayers indicates whether MaxPlayers field should be validated.
	ValidateMaxPlayers bool

	// TickRate is the server tick rate in Hz (updates per second).
	// Valid range: 1 to 60 Hz (diminishing returns above 60 Hz)
	TickRate int

	// ValidateTickRate indicates whether TickRate field should be validated.
	ValidateTickRate bool

	// Genre is the game genre identifier.
	// Valid values can be retrieved via Validator.GetAvailableGenres()
	Genre string

	// SaveDir is the directory path for game save files.
	// Will be created if CreateDirs is true and directory doesn't exist.
	SaveDir string

	// LogDir is the directory path for log files.
	// Will be created if CreateDirs is true and directory doesn't exist.
	LogDir string

	// ModsDir is the directory path for game modifications.
	// Will be created if CreateDirs is true and directory doesn't exist.
	ModsDir string

	// CreateDirs indicates whether to create missing directories during validation.
	CreateDirs bool
}

// CharacterClass represents a player archetype with specific stat distributions.
// Moved from pkg/engine to avoid circular dependencies and allow usage in non-engine packages.
type CharacterClass int

const (
	// Base Classes (6 original)
	// ClassWarrior is a high HP, melee-focused class
	ClassWarrior CharacterClass = iota
	// ClassMage is a high mana, magic-focused class
	ClassMage
	// ClassRogue is a balanced, agility-focused class
	ClassRogue
	// ClassRanger is a ranged combat class with pet bonding abilities (V4 Phase 25)
	ClassRanger
	// ClassCleric is a support class with healing and buffs (V4 Phase 25)
	ClassCleric
	// ClassNecromancer is a summoning class with life drain and debuffs (V4 Phase 25)
	ClassNecromancer

	// Hybrid Classes (15 combinations) - Phase 25.2 Extension
	// ClassBattlemage combines Warrior melee prowess with Mage spellcasting
	ClassBattlemage
	// ClassSpellblade combines Rogue agility with Mage magic
	ClassSpellblade
	// ClassPaladin combines Warrior strength with Cleric holy powers
	ClassPaladin
	// ClassMonk combines Rogue speed with Cleric spiritual discipline
	ClassMonk
	// ClassDeathKnight combines Warrior combat with Necromancer dark magic
	ClassDeathKnight
	// ClassWitchHunter combines Ranger precision with Cleric divine power
	ClassWitchHunter
	// ClassBeastlord combines Warrior might with Ranger beast mastery
	ClassBeastlord
	// ClassArcaneArcher combines Ranger marksmanship with Mage arcane arts
	ClassArcaneArcher
	// ClassShadowPriest combines Rogue shadows with Necromancer dark arts
	ClassShadowPriest
	// ClassDruid combines Ranger nature affinity with Mage elemental magic
	ClassDruid
	// ClassInquisitor combines Cleric faith with Rogue investigation
	ClassInquisitor
	// ClassBloodKnight combines Warrior combat with Necromancer blood magic
	ClassBloodKnight
	// ClassMystic combines Mage arcane knowledge with Cleric divine wisdom
	ClassMystic
	// ClassWarlock combines Mage magic with Necromancer dark pacts
	ClassWarlock
	// ClassNinja combines Rogue stealth with Ranger precision strikes
	ClassNinja
)

// String returns the human-readable class name
func (c CharacterClass) String() string {
	switch c {
	case ClassWarrior:
		return "Warrior"
	case ClassMage:
		return "Mage"
	case ClassRogue:
		return "Rogue"
	case ClassRanger:
		return "Ranger"
	case ClassCleric:
		return "Cleric"
	case ClassNecromancer:
		return "Necromancer"
	case ClassBattlemage:
		return "Battlemage"
	case ClassSpellblade:
		return "Spellblade"
	case ClassPaladin:
		return "Paladin"
	case ClassMonk:
		return "Monk"
	case ClassDeathKnight:
		return "Death Knight"
	case ClassWitchHunter:
		return "Witch Hunter"
	case ClassBeastlord:
		return "Beastlord"
	case ClassArcaneArcher:
		return "Arcane Archer"
	case ClassShadowPriest:
		return "Shadow Priest"
	case ClassDruid:
		return "Druid"
	case ClassInquisitor:
		return "Inquisitor"
	case ClassBloodKnight:
		return "Blood Knight"
	case ClassMystic:
		return "Mystic"
	case ClassWarlock:
		return "Warlock"
	case ClassNinja:
		return "Ninja"
	default:
		return "Unknown"
	}
}

// Description returns a short description of the class
func (c CharacterClass) Description() string {
	switch c {
	case ClassWarrior:
		return "High HP melee fighter with strong defensive capabilities"
	case ClassMage:
		return "High mana spellcaster with devastating magical attacks"
	case ClassRogue:
		return "Balanced agility-focused class with critical strike bonuses"
	case ClassRanger:
		return "Ranged combat specialist with pet bonding abilities"
	case ClassCleric:
		return "Support class specializing in healing and divine buffs"
	case ClassNecromancer:
		return "Dark summoner with life drain and debuff abilities"
	case ClassBattlemage:
		return "Warrior-Mage hybrid blending melee combat with spellcasting"
	case ClassSpellblade:
		return "Rogue-Mage hybrid combining agility with arcane magic"
	case ClassPaladin:
		return "Warrior-Cleric hybrid with holy powers and strong defense"
	case ClassMonk:
		return "Rogue-Cleric hybrid mastering spiritual discipline and speed"
	case ClassDeathKnight:
		return "Warrior-Necromancer hybrid wielding dark combat magic"
	case ClassWitchHunter:
		return "Ranger-Cleric hybrid using divine power for precision strikes"
	case ClassBeastlord:
		return "Warrior-Ranger hybrid with enhanced beast mastery"
	case ClassArcaneArcher:
		return "Ranger-Mage hybrid infusing arrows with arcane energy"
	case ClassShadowPriest:
		return "Rogue-Necromancer hybrid manipulating shadows and dark magic"
	case ClassDruid:
		return "Ranger-Mage hybrid channeling nature and elemental forces"
	case ClassInquisitor:
		return "Cleric-Rogue hybrid combining faith with investigation"
	case ClassBloodKnight:
		return "Warrior-Necromancer hybrid mastering blood magic"
	case ClassMystic:
		return "Mage-Cleric hybrid blending arcane and divine wisdom"
	case ClassWarlock:
		return "Mage-Necromancer hybrid wielding dark pacts and corruption"
	case ClassNinja:
		return "Rogue-Ranger hybrid mastering stealth and precision"
	default:
		return "Unknown class"
	}
}

// LowerName returns the lowercase name of the class for matching with item restrictions.
// Phase 25.2: Used for class-specific equipment restrictions.
func (c CharacterClass) LowerName() string {
	switch c {
	case ClassWarrior:
		return "warrior"
	case ClassMage:
		return "mage"
	case ClassRogue:
		return "rogue"
	case ClassRanger:
		return "ranger"
	case ClassCleric:
		return "cleric"
	case ClassNecromancer:
		return "necromancer"
	case ClassBattlemage:
		return "battlemage"
	case ClassSpellblade:
		return "spellblade"
	case ClassPaladin:
		return "paladin"
	case ClassMonk:
		return "monk"
	case ClassDeathKnight:
		return "deathknight"
	case ClassWitchHunter:
		return "witchhunter"
	case ClassBeastlord:
		return "beastlord"
	case ClassArcaneArcher:
		return "arcanearcher"
	case ClassShadowPriest:
		return "shadowpriest"
	case ClassDruid:
		return "druid"
	case ClassInquisitor:
		return "inquisitor"
	case ClassBloodKnight:
		return "bloodknight"
	case ClassMystic:
		return "mystic"
	case ClassWarlock:
		return "warlock"
	case ClassNinja:
		return "ninja"
	default:
		return "unknown"
	}
}
