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
func GetClassAbilities(class CharacterClass) []string {
	switch class {
	case ClassWarrior:
		return []string{"power_strike", "shield_bash", "battle_cry", "cleave"}
	case ClassRogue:
		return []string{"backstab", "dual_wield", "stealth", "poison_blade"}
	case ClassMage:
		return []string{"fireball", "ice_shard", "magic_missile", "mana_shield"}
	case ClassRanger:
		return []string{"aimed_shot", "rapid_fire", "tame_beast", "track"}
	case ClassCleric:
		return []string{"heal", "smite", "divine_shield", "prayer"}
	case ClassNecromancer:
		return []string{"raise_dead", "life_drain", "curse", "bone_armor"}
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
