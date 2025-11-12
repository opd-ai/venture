package engine

// ClassProgressionComponent tracks character class level and specialization.
// This extends the basic CharacterClass with progression mechanics for V4 Phase 25.
type ClassProgressionComponent struct {
	Class          CharacterClass
	Level          int
	Experience     float64
	Specialization SpecializationType
	Abilities      []string // Unlocked ability IDs
}

// Type returns the component type identifier.
func (c ClassProgressionComponent) Type() string {
	return "class_progression"
}

// SpecializationType represents a class specialization path (unlocked at level 10).
type SpecializationType int

const (
	// SpecializationNone means no specialization chosen yet
	SpecializationNone SpecializationType = iota

	// Warrior specializations
	SpecializationBerserker // High damage, rage mechanics
	SpecializationDefender  // Tank, defensive focus

	// Mage specializations
	SpecializationElementalist // Elemental mastery
	SpecializationArcanist     // Raw magic power

	// Rogue specializations
	SpecializationAssassin     // Burst damage, poison
	SpecializationShadowdancer // Stealth, evasion

	// Ranger specializations
	SpecializationBeastmaster // Pet focus
	SpecializationMarksman    // Ranged precision

	// Cleric specializations
	SpecializationHealer  // Pure healing
	SpecializationTemplar // Holy warrior

	// Necromancer specializations
	SpecializationDeathKnight // Undead army
	SpecializationBloodMage   // Life drain
)

// String returns the string representation of a SpecializationType.
func (s SpecializationType) String() string {
	switch s {
	case SpecializationNone:
		return "None"
	case SpecializationBerserker:
		return "Berserker"
	case SpecializationDefender:
		return "Defender"
	case SpecializationElementalist:
		return "Elementalist"
	case SpecializationArcanist:
		return "Arcanist"
	case SpecializationAssassin:
		return "Assassin"
	case SpecializationShadowdancer:
		return "Shadowdancer"
	case SpecializationBeastmaster:
		return "Beastmaster"
	case SpecializationMarksman:
		return "Marksman"
	case SpecializationHealer:
		return "Healer"
	case SpecializationTemplar:
		return "Templar"
	case SpecializationDeathKnight:
		return "Death Knight"
	case SpecializationBloodMage:
		return "Blood Mage"
	default:
		return "Unknown"
	}
}

// GetClassAbilities returns starting abilities for a character class.
// Phase 25.1: Extended to 8+ abilities per class (previously 4).
func GetClassAbilities(class CharacterClass) []string {
	switch class {
	case ClassWarrior:
		return []string{
			"power_strike",     // Basic heavy attack
			"shield_bash",      // Stun attack
			"battle_cry",       // AOE buff
			"cleave",           // Multi-target attack
			"charge",           // Gap closer
			"defensive_stance", // Defense buff
			"execute",          // Finishing move
			"taunt",            // Threat generation
		}
	case ClassRogue:
		return []string{
			"backstab",     // High damage from behind
			"dual_wield",   // Attack with both weapons
			"stealth",      // Invisibility
			"poison_blade", // DOT attack
			"evade",        // Dodge attacks
			"ambush",       // Surprise attack from stealth
			"shadow_step",  // Blink behind enemy
			"disarm",       // Remove enemy weapon
		}
	case ClassMage:
		return []string{
			"fireball",       // Fire damage spell
			"ice_shard",      // Slow + damage
			"magic_missile",  // Auto-hit projectile
			"mana_shield",    // Absorb damage with mana
			"lightning_bolt", // Chain lightning
			"frost_nova",     // AOE freeze
			"teleport",       // Blink away
			"arcane_barrage", // Rapid fire spells
		}
	case ClassRanger:
		return []string{
			"aimed_shot",     // High damage ranged
			"rapid_fire",     // Multiple shots
			"tame_beast",     // Recruit companion
			"track",          // Reveal enemies
			"explosive_shot", // AOE ranged
			"multi_shot",     // Hit multiple targets
			"camouflage",     // Stealth in nature
			"hunters_mark",   // Increase damage on target
		}
	case ClassCleric:
		return []string{
			"heal",          // Restore HP
			"smite",         // Holy damage
			"divine_shield", // Damage immunity
			"prayer",        // HP regeneration buff
			"resurrection",  // Revive dead ally
			"holy_light",    // AOE heal
			"purify",        // Remove debuffs
			"blessing",      // All stats buff
		}
	case ClassNecromancer:
		return []string{
			"raise_dead",       // Summon skeleton
			"life_drain",       // Steal HP
			"curse",            // Debuff enemy
			"bone_armor",       // Shield
			"death_coil",       // Damage or heal
			"fear",             // Make enemy flee
			"corpse_explosion", // AOE from corpse
			"soul_harvest",     // Gain power from kills
		}
	default:
		return []string{}
	}
}

// GetSpecializationAbilities returns additional abilities for a specialization.
func GetSpecializationAbilities(spec SpecializationType) []string {
	switch spec {
	case SpecializationBerserker:
		return []string{"rage", "whirlwind", "reckless_attack"}
	case SpecializationDefender:
		return []string{"shield_wall", "taunt", "iron_skin"}
	case SpecializationElementalist:
		return []string{"lightning_bolt", "frost_nova", "pyroblast"}
	case SpecializationArcanist:
		return []string{"arcane_blast", "time_warp", "spell_steal"}
	case SpecializationAssassin:
		return []string{"shadow_strike", "deadly_poison", "vanish"}
	case SpecializationShadowdancer:
		return []string{"shadow_step", "smoke_bomb", "evasion"}
	case SpecializationBeastmaster:
		return []string{"call_of_wild", "bestial_wrath", "mend_pet"}
	case SpecializationMarksman:
		return []string{"snipe", "explosive_shot", "kill_shot"}
	case SpecializationHealer:
		return []string{"greater_heal", "resurrect", "dispel_magic"}
	case SpecializationTemplar:
		return []string{"holy_strike", "sacred_ground", "judgment"}
	case SpecializationDeathKnight:
		return []string{"death_coil", "army_of_dead", "corpse_explosion"}
	case SpecializationBloodMage:
		return []string{"blood_boil", "vampiric_touch", "siphon_life"}
	default:
		return []string{}
	}
}

// GetAvailableSpecializations returns valid specializations for a class.
func GetAvailableSpecializations(class CharacterClass) []SpecializationType {
	switch class {
	case ClassWarrior:
		return []SpecializationType{SpecializationBerserker, SpecializationDefender}
	case ClassMage:
		return []SpecializationType{SpecializationElementalist, SpecializationArcanist}
	case ClassRogue:
		return []SpecializationType{SpecializationAssassin, SpecializationShadowdancer}
	case ClassRanger:
		return []SpecializationType{SpecializationBeastmaster, SpecializationMarksman}
	case ClassCleric:
		return []SpecializationType{SpecializationHealer, SpecializationTemplar}
	case ClassNecromancer:
		return []SpecializationType{SpecializationDeathKnight, SpecializationBloodMage}
	default:
		return []SpecializationType{}
	}
}
