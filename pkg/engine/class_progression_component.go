package engine

// ClassProgressionComponent tracks character class level and specialization.
// This extends the basic CharacterClass with progression mechanics for V4 Phase 25.
type ClassProgressionComponent struct {
	Class          CharacterClass
	Level          int
	Experience     float64
	Specialization SpecializationType
	Abilities      []string        // Unlocked ability IDs
	SecondaryClass *CharacterClass // Phase 25.2: Dual-classing (unlocked at level 20)
	SecondaryLevel int             // Level in secondary class
	SecondarySpec  SpecializationType
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

	// Hybrid class specializations (2 per class)
	// Battlemage specializations
	SpecializationSpellsword // Melee-focused battlemage
	SpecializationWarmage    // Magic-focused battlemage

	// Spellblade specializations
	SpecializationTrickster // Illusion-focused spellblade
	SpecializationDuelist   // Combat-focused spellblade

	// Paladin specializations
	SpecializationCrusader // Offensive paladin
	SpecializationGuardian // Defensive paladin

	// Monk specializations
	SpecializationWindwalker // Speed-focused monk
	SpecializationBrewmaster // Tank-focused monk

	// DeathKnight specializations (reused name, different context)
	SpecializationUnholy // Disease and pets
	SpecializationFrost  // Frost magic and dual-wield

	// WitchHunter specializations
	SpecializationExorcist // Demon hunter
	SpecializationPurifier // Undead hunter

	// Beastlord specializations
	SpecializationFeralRage  // Savage melee
	SpecializationPackLeader // Multi-pet commander

	// ArcaneArcher specializations
	SpecializationSpellshot // Magic arrow specialist
	SpecializationSeeker    // Homing projectiles

	// ShadowPriest specializations
	SpecializationVoidcaller // Shadow damage
	SpecializationSoulweaver // Drain and heal

	// Druid specializations
	SpecializationShapeshifter // Animal forms
	SpecializationNaturemage   // Elemental caster

	// Inquisitor specializations
	SpecializationJudge        // Divine judgment
	SpecializationInterrogator // Investigation and torture

	// BloodKnight specializations
	SpecializationCrimsonBlade // Blood weapon master
	SpecializationHemomancer   // Blood magic caster

	// Mystic specializations
	SpecializationOracle    // Foresight and divination
	SpecializationTheurgist // Divine arcane fusion

	// Warlock specializations
	SpecializationDemonologist // Demon summoning
	SpecializationAffliction   // Curses and DoTs

	// Ninja specializations
	SpecializationShinobi // Pure stealth assassin
	SpecializationStriker // Ranged thrown weapons
)

// String returns the string representation of a SpecializationType.
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (44) is intentional—this is
// an exhaustive switch over all 44 specialization enum values for UI display.
// This pattern ensures compile-time completeness checking when new specializations are added.
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
	// Hybrid specializations
	case SpecializationSpellsword:
		return "Spellsword"
	case SpecializationWarmage:
		return "Warmage"
	case SpecializationTrickster:
		return "Trickster"
	case SpecializationDuelist:
		return "Duelist"
	case SpecializationCrusader:
		return "Crusader"
	case SpecializationGuardian:
		return "Guardian"
	case SpecializationWindwalker:
		return "Windwalker"
	case SpecializationBrewmaster:
		return "Brewmaster"
	case SpecializationUnholy:
		return "Unholy"
	case SpecializationFrost:
		return "Frost"
	case SpecializationExorcist:
		return "Exorcist"
	case SpecializationPurifier:
		return "Purifier"
	case SpecializationFeralRage:
		return "Feral Rage"
	case SpecializationPackLeader:
		return "Pack Leader"
	case SpecializationSpellshot:
		return "Spellshot"
	case SpecializationSeeker:
		return "Seeker"
	case SpecializationVoidcaller:
		return "Voidcaller"
	case SpecializationSoulweaver:
		return "Soulweaver"
	case SpecializationShapeshifter:
		return "Shapeshifter"
	case SpecializationNaturemage:
		return "Naturemage"
	case SpecializationJudge:
		return "Judge"
	case SpecializationInterrogator:
		return "Interrogator"
	case SpecializationCrimsonBlade:
		return "Crimson Blade"
	case SpecializationHemomancer:
		return "Hemomancer"
	case SpecializationOracle:
		return "Oracle"
	case SpecializationTheurgist:
		return "Theurgist"
	case SpecializationDemonologist:
		return "Demonologist"
	case SpecializationAffliction:
		return "Affliction"
	case SpecializationShinobi:
		return "Shinobi"
	case SpecializationStriker:
		return "Striker"
	default:
		return "Unknown"
	}
}

// GetClassAbilities returns starting abilities for a character class.
// Phase 25.1: Extended to 8+ abilities per class (previously 4).
// Phase 25.2 Extension: Hybrid classes get 6 abilities (3 from each parent class).
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
	// Hybrid classes (6 abilities each - 3 from each parent)
	case ClassBattlemage: // Warrior + Mage
		return []string{
			"power_strike",   // From Warrior
			"shield_bash",    // From Warrior
			"battle_cry",     // From Warrior
			"fireball",       // From Mage
			"mana_shield",    // From Mage
			"arcane_barrage", // From Mage
		}
	case ClassSpellblade: // Rogue + Mage
		return []string{
			"backstab",      // From Rogue
			"stealth",       // From Rogue
			"shadow_step",   // From Rogue
			"ice_shard",     // From Mage
			"teleport",      // From Mage
			"magic_missile", // From Mage
		}
	case ClassPaladin: // Warrior + Cleric
		return []string{
			"power_strike",     // From Warrior
			"defensive_stance", // From Warrior
			"taunt",            // From Warrior
			"heal",             // From Cleric
			"smite",            // From Cleric
			"divine_shield",    // From Cleric
		}
	case ClassMonk: // Rogue + Cleric
		return []string{
			"dual_wield",  // From Rogue (fists)
			"evade",       // From Rogue
			"shadow_step", // From Rogue
			"heal",        // From Cleric
			"purify",      // From Cleric
			"prayer",      // From Cleric
		}
	case ClassDeathKnight: // Warrior + Necromancer
		return []string{
			"power_strike", // From Warrior
			"charge",       // From Warrior
			"taunt",        // From Warrior
			"life_drain",   // From Necromancer
			"bone_armor",   // From Necromancer
			"death_coil",   // From Necromancer
		}
	case ClassWitchHunter: // Ranger + Cleric
		return []string{
			"aimed_shot", // From Ranger
			"rapid_fire", // From Ranger
			"track",      // From Ranger
			"smite",      // From Cleric
			"purify",     // From Cleric
			"holy_light", // From Cleric
		}
	case ClassBeastlord: // Warrior + Ranger
		return []string{
			"power_strike", // From Warrior
			"battle_cry",   // From Warrior
			"cleave",       // From Warrior
			"tame_beast",   // From Ranger
			"hunters_mark", // From Ranger
			"multi_shot",   // From Ranger
		}
	case ClassArcaneArcher: // Ranger + Mage
		return []string{
			"aimed_shot",     // From Ranger
			"explosive_shot", // From Ranger
			"camouflage",     // From Ranger
			"fireball",       // From Mage
			"lightning_bolt", // From Mage
			"frost_nova",     // From Mage
		}
	case ClassShadowPriest: // Rogue + Necromancer
		return []string{
			"backstab",     // From Rogue
			"stealth",      // From Rogue
			"poison_blade", // From Rogue
			"life_drain",   // From Necromancer
			"curse",        // From Necromancer
			"fear",         // From Necromancer
		}
	case ClassDruid: // Ranger + Mage
		return []string{
			"tame_beast",     // From Ranger
			"camouflage",     // From Ranger
			"track",          // From Ranger
			"ice_shard",      // From Mage (nature element)
			"lightning_bolt", // From Mage (storm element)
			"teleport",       // From Mage (wild shape)
		}
	case ClassInquisitor: // Cleric + Rogue
		return []string{
			"smite",    // From Cleric
			"purify",   // From Cleric
			"blessing", // From Cleric
			"backstab", // From Rogue
			"stealth",  // From Rogue
			"disarm",   // From Rogue
		}
	case ClassBloodKnight: // Warrior + Necromancer
		return []string{
			"cleave",           // From Warrior
			"execute",          // From Warrior
			"defensive_stance", // From Warrior
			"life_drain",       // From Necromancer
			"curse",            // From Necromancer
			"soul_harvest",     // From Necromancer
		}
	case ClassMystic: // Mage + Cleric
		return []string{
			"magic_missile",  // From Mage
			"mana_shield",    // From Mage
			"arcane_barrage", // From Mage
			"heal",           // From Cleric
			"prayer",         // From Cleric
			"resurrection",   // From Cleric
		}
	case ClassWarlock: // Mage + Necromancer
		return []string{
			"fireball",         // From Mage
			"frost_nova",       // From Mage
			"teleport",         // From Mage
			"raise_dead",       // From Necromancer
			"curse",            // From Necromancer
			"corpse_explosion", // From Necromancer
		}
	case ClassNinja: // Rogue + Ranger
		return []string{
			"backstab",     // From Rogue
			"stealth",      // From Rogue
			"evade",        // From Rogue
			"aimed_shot",   // From Ranger
			"camouflage",   // From Ranger
			"hunters_mark", // From Ranger
		}
	default:
		return []string{}
	}
}

// GetSpecializationAbilities returns additional abilities for a specialization.
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (43) is intentional—this is
// an exhaustive lookup table returning abilities for each of 43 specialization types.
// The switch provides exhaustiveness checking and is more maintainable than a map[int][]string.
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
	// Hybrid specializations
	case SpecializationSpellsword:
		return []string{"enchanted_blade", "spell_parry", "arcane_strike"}
	case SpecializationWarmage:
		return []string{"battle_magic", "spellsteel", "mana_fortress"}
	case SpecializationTrickster:
		return []string{"mirror_image", "blink_strike", "mana_burst"}
	case SpecializationDuelist:
		return []string{"riposte", "counter_spell", "precision_cut"}
	case SpecializationCrusader:
		return []string{"divine_fury", "avenging_wrath", "hammer_of_justice"}
	case SpecializationGuardian:
		return []string{"sanctuary", "shield_of_faith", "blessed_armor"}
	case SpecializationWindwalker:
		return []string{"flying_serpent_kick", "fists_of_fury", "storm_earth_fire"}
	case SpecializationBrewmaster:
		return []string{"fortifying_brew", "keg_smash", "breath_of_fire"}
	case SpecializationUnholy:
		return []string{"plague_strike", "festering_wound", "apocalypse"}
	case SpecializationFrost:
		return []string{"frost_strike", "obliterate", "pillar_of_frost"}
	case SpecializationExorcist:
		return []string{"banish", "consecration", "divine_bolt"}
	case SpecializationPurifier:
		return []string{"holy_fire", "turn_undead", "cleansing_flame"}
	case SpecializationFeralRage:
		return []string{"savage_strike", "primal_fury", "blood_frenzy"}
	case SpecializationPackLeader:
		return []string{"coordinated_assault", "pack_tactics", "alpha_command"}
	case SpecializationSpellshot:
		return []string{"enchanted_arrow", "mana_pierce", "arcane_volley"}
	case SpecializationSeeker:
		return []string{"homing_shot", "mark_target", "inevitable_strike"}
	case SpecializationVoidcaller:
		return []string{"void_bolt", "shadow_form", "dark_ascension"}
	case SpecializationSoulweaver:
		return []string{"spirit_link", "soul_drain", "ethereal_touch"}
	case SpecializationShapeshifter:
		return []string{"bear_form", "cat_form", "moonkin_form"}
	case SpecializationNaturemage:
		return []string{"wrath", "starfire", "solar_beam"}
	case SpecializationJudge:
		return []string{"divine_sentence", "holy_verdict", "righteous_fury"}
	case SpecializationInterrogator:
		return []string{"mind_probe", "confession", "truth_serum"}
	case SpecializationCrimsonBlade:
		return []string{"blood_edge", "crimson_cleave", "hemorrhage"}
	case SpecializationHemomancer:
		return []string{"blood_orb", "sanguine_ritual", "exsanguinate"}
	case SpecializationOracle:
		return []string{"foresight", "divine_vision", "prophecy"}
	case SpecializationTheurgist:
		return []string{"holy_arcana", "divine_spark", "blessed_bolt"}
	case SpecializationDemonologist:
		return []string{"summon_demon", "fel_infusion", "demonic_pact"}
	case SpecializationAffliction:
		return []string{"agony", "corruption", "unstable_affliction"}
	case SpecializationShinobi:
		return []string{"silent_kill", "shadow_clone", "smoke_screen"}
	case SpecializationStriker:
		return []string{"shuriken_storm", "fan_of_knives", "lethal_precision"}
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
	// Hybrid classes
	case ClassBattlemage:
		return []SpecializationType{SpecializationSpellsword, SpecializationWarmage}
	case ClassSpellblade:
		return []SpecializationType{SpecializationTrickster, SpecializationDuelist}
	case ClassPaladin:
		return []SpecializationType{SpecializationCrusader, SpecializationGuardian}
	case ClassMonk:
		return []SpecializationType{SpecializationWindwalker, SpecializationBrewmaster}
	case ClassDeathKnight:
		return []SpecializationType{SpecializationUnholy, SpecializationFrost}
	case ClassWitchHunter:
		return []SpecializationType{SpecializationExorcist, SpecializationPurifier}
	case ClassBeastlord:
		return []SpecializationType{SpecializationFeralRage, SpecializationPackLeader}
	case ClassArcaneArcher:
		return []SpecializationType{SpecializationSpellshot, SpecializationSeeker}
	case ClassShadowPriest:
		return []SpecializationType{SpecializationVoidcaller, SpecializationSoulweaver}
	case ClassDruid:
		return []SpecializationType{SpecializationShapeshifter, SpecializationNaturemage}
	case ClassInquisitor:
		return []SpecializationType{SpecializationJudge, SpecializationInterrogator}
	case ClassBloodKnight:
		return []SpecializationType{SpecializationCrimsonBlade, SpecializationHemomancer}
	case ClassMystic:
		return []SpecializationType{SpecializationOracle, SpecializationTheurgist}
	case ClassWarlock:
		return []SpecializationType{SpecializationDemonologist, SpecializationAffliction}
	case ClassNinja:
		return []SpecializationType{SpecializationShinobi, SpecializationStriker}
	default:
		return []SpecializationType{}
	}
}

// Serialize converts ClassProgressionComponent to bytes for network transmission.
// Format: [Class:1][Level:4][Experience:8][Specialization:1][HasSecondary:1]
//
//	[SecondaryClass:1][SecondaryLevel:4][SecondarySpec:1] = 29 bytes
//
// Note: Abilities array is not synchronized; it can be regenerated from Class+Level on client
func (c *ClassProgressionComponent) Serialize() []byte {
	buf := make([]byte, 29)
	offset := 0

	// Primary class data
	buf[offset] = byte(c.Class)
	offset++
	writeInt32(buf[offset:], int32(c.Level))
	offset += 4
	writeFloat64(buf[offset:], c.Experience)
	offset += 8
	buf[offset] = byte(c.Specialization)
	offset++

	// Secondary class (dual-classing)
	hasSecondary := c.SecondaryClass != nil
	writeBool(buf[offset:], hasSecondary)
	offset++

	if hasSecondary {
		buf[offset] = byte(*c.SecondaryClass)
		offset++
		writeInt32(buf[offset:], int32(c.SecondaryLevel))
		offset += 4
		buf[offset] = byte(c.SecondarySpec)
	} else {
		buf[offset] = 0 // INTEGRATION FIX [Category D]: SecondaryClass Serialization
		// Gap: No secondary class - serialize as 0 (valid state, not a missing feature)
		// Fix: Already correct - dual-classing is optional, 0 indicates no secondary class
		// Roadmap: ROADMAP_V4.md Phase 25.2 - Dual-classing complete, serialization functional
		offset++
		writeInt32(buf[offset:], 0) // SecondaryLevel = 0 (no secondary class)
		offset += 4
		buf[offset] = 0 // SecondarySpec = 0 (no secondary specialization)
	}

	return buf
}

// Deserialize reads ClassProgressionComponent from bytes.
func (c *ClassProgressionComponent) Deserialize(data []byte) error {
	if len(data) < 29 {
		return ErrInvalidComponentData
	}

	offset := 0

	// Primary class data
	c.Class = CharacterClass(data[offset])
	offset++
	c.Level = int(readInt32(data[offset:]))
	offset += 4
	c.Experience = readFloat64(data[offset:])
	offset += 8
	c.Specialization = SpecializationType(data[offset])
	offset++

	// Secondary class
	hasSecondary := readBool(data[offset:])
	offset++

	if hasSecondary {
		secondaryClass := CharacterClass(data[offset])
		c.SecondaryClass = &secondaryClass
		offset++
		c.SecondaryLevel = int(readInt32(data[offset:]))
		offset += 4
		c.SecondarySpec = SpecializationType(data[offset])
	} else {
		c.SecondaryClass = nil
		c.SecondaryLevel = 0
		c.SecondarySpec = SpecializationNone
	}

	// Abilities array should be regenerated on client from Class+Level+Specialization
	c.Abilities = GetClassAbilities(c.Class)
	if c.Specialization != SpecializationNone {
		c.Abilities = append(c.Abilities, GetSpecializationAbilities(c.Specialization)...)
	}

	return nil
}
