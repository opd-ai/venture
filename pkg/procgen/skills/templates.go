// Package skills provides skill tree templates.
// This file defines genre-specific skill tree template data used by the skill generator
// to create structured progression systems. Type definitions have been moved to types.go.
//
// # Custom Genre Template Structure
//
// To add custom genre support, create a template function returning []SkillTreeTemplate:
//
//	func GetCyberpunkTreeTemplates() []SkillTreeTemplate {
//	    return []SkillTreeTemplate{
//	        {
//	            Name:        "Netrunner",
//	            Description: "Master of digital infiltration and hacking",
//	            Category:    CategoryMagic, // Repurposed for hacking
//	            SkillTemplates: []SkillTemplate{
//	                {
//	                    BaseType:          TypeActive,
//	                    BaseCategory:      CategoryMagic,
//	                    NamePrefixes:      []string{"Neural", "Cyber", "Digital"},
//	                    NameSuffixes:      []string{"Breach", "Exploit", "Override"},
//	                    DescriptionFormat: "Hack %s systems remotely",
//	                    EffectTypes:       []string{"hack_power", "ice_break"},
//	                    ValueRanges: map[string][2]float64{
//	                        "hack_power": {10.0, 50.0},
//	                        "ice_break":  {0.1, 0.4},
//	                    },
//	                    Tags:          []string{"hacking", "netrunner"},
//	                    TierRange:     [2]int{0, 6},
//	                    MaxLevelRange: [2]int{1, 5},
//	                },
//	            },
//	        },
//	    }
//	}
//
// Then register in generator.go's getTemplates() switch statement.
package skills

// GetFantasyTreeTemplates returns skill tree templates themed for fantasy genre worlds,
// including warrior, mage, ranger, and healer trees with magic and combat abilities.
func GetFantasyTreeTemplates() []SkillTreeTemplate {
	return []SkillTreeTemplate{
		{
			Name:        "Warrior",
			Description: "Master of melee combat and physical prowess",
			Category:    CategoryCombat,
			SkillTemplates: []SkillTemplate{
				// Combat passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Weapon", "Combat", "Battle", "Melee"},
					NameSuffixes:      []string{"Mastery", "Training", "Expertise", "Proficiency"},
					DescriptionFormat: "Improves %s effectiveness in combat",
					EffectTypes:       []string{"damage", "crit_chance", "attack_speed"},
					ValueRanges: map[string][2]float64{
						"damage":       {0.05, 0.15},
						"crit_chance":  {0.02, 0.08},
						"attack_speed": {0.03, 0.10},
					},
					Tags:          []string{"combat", "passive", "weapon"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active combat skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Cleave", "Slash", "Smash", "Charge"},
					NameSuffixes:      []string{"Strike", "Blow", "Attack", "Assault"},
					DescriptionFormat: "Powerful %s that damages enemies",
					EffectTypes:       []string{"damage", "aoe_damage", "stun_chance"},
					ValueRanges: map[string][2]float64{
						"damage":      {0.50, 1.50},
						"aoe_damage":  {0.30, 0.80},
						"stun_chance": {0.10, 0.30},
					},
					Tags:          []string{"combat", "active", "aoe"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Defensive skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Iron", "Stone", "Hardened", "Fortified"},
					NameSuffixes:      []string{"Skin", "Defense", "Will", "Resolve"},
					DescriptionFormat: "Increases %s and survivability",
					EffectTypes:       []string{"armor", "health", "damage_reduction"},
					ValueRanges: map[string][2]float64{
						"armor":            {0.05, 0.15},
						"health":           {0.10, 0.25},
						"damage_reduction": {0.03, 0.10},
					},
					Tags:          []string{"defense", "passive", "tank"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Titan's", "Berserker's", "Warlord's", "Champion's"},
					NameSuffixes:      []string{"Fury", "Rage", "Wrath", "Rampage"},
					DescriptionFormat: "Unleash %s for devastating damage",
					EffectTypes:       []string{"damage", "lifesteal", "damage_reduction"},
					ValueRanges: map[string][2]float64{
						"damage":           {2.00, 4.00},
						"lifesteal":        {0.20, 0.50},
						"damage_reduction": {0.30, 0.50},
					},
					Tags:          []string{"ultimate", "combat", "burst"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Mage",
			Description: "Master of arcane arts and elemental magic",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Magic passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Arcane", "Mystic", "Ethereal", "Magical"},
					NameSuffixes:      []string{"Focus", "Attunement", "Resonance", "Mastery"},
					DescriptionFormat: "Enhances %s magical abilities",
					EffectTypes:       []string{"spell_damage", "mana_regen", "cast_speed"},
					ValueRanges: map[string][2]float64{
						"spell_damage": {0.08, 0.20},
						"mana_regen":   {0.05, 0.15},
						"cast_speed":   {0.05, 0.12},
					},
					Tags:          []string{"magic", "passive", "caster"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active spells
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Fireball", "Ice", "Lightning", "Arcane"},
					NameSuffixes:      []string{"Blast", "Storm", "Nova", "Missile"},
					DescriptionFormat: "Launch %s at enemies",
					EffectTypes:       []string{"spell_damage", "mana_cost", "cooldown_reduction"},
					ValueRanges: map[string][2]float64{
						"spell_damage":       {0.60, 1.80},
						"mana_cost":          {-0.10, -0.05},
						"cooldown_reduction": {0.05, 0.15},
					},
					Tags:          []string{"magic", "active", "spell"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Mana skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Deep", "Vast", "Ancient", "Boundless"},
					NameSuffixes:      []string{"Mana", "Reserves", "Knowledge", "Power"},
					DescriptionFormat: "Increases %s and magical capacity",
					EffectTypes:       []string{"max_mana", "mana_regen", "spell_efficiency"},
					ValueRanges: map[string][2]float64{
						"max_mana":         {0.10, 0.30},
						"mana_regen":       {0.08, 0.20},
						"spell_efficiency": {0.05, 0.12},
					},
					Tags:          []string{"magic", "passive", "mana"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Cataclysm", "Meteor", "Apocalypse", "Oblivion"},
					NameSuffixes:      []string{"", "", "", ""},
					DescriptionFormat: "Summon %s to devastate all enemies",
					EffectTypes:       []string{"spell_damage", "aoe_radius", "pierce_resistance"},
					ValueRanges: map[string][2]float64{
						"spell_damage":      {3.00, 6.00},
						"aoe_radius":        {0.50, 1.00},
						"pierce_resistance": {0.30, 0.60},
					},
					Tags:          []string{"ultimate", "magic", "aoe"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Rogue",
			Description: "Master of stealth, speed, and precision",
			Category:    CategoryUtility,
			SkillTemplates: []SkillTemplate{
				// Speed/Agility passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Swift", "Nimble", "Agile", "Fleet"},
					NameSuffixes:      []string{"Feet", "Movement", "Reflexes", "Steps"},
					DescriptionFormat: "Improves %s and mobility",
					EffectTypes:       []string{"move_speed", "dodge_chance", "evasion"},
					ValueRanges: map[string][2]float64{
						"move_speed":   {0.05, 0.15},
						"dodge_chance": {0.03, 0.10},
						"evasion":      {0.05, 0.12},
					},
					Tags:          []string{"utility", "passive", "mobility"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Stealth/Crit skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Shadow", "Backstab", "Assassinate", "Ambush"},
					NameSuffixes:      []string{"Strike", "Technique", "", "Attack"},
					DescriptionFormat: "Execute %s from stealth for massive damage",
					EffectTypes:       []string{"crit_damage", "crit_chance", "stealth_bonus"},
					ValueRanges: map[string][2]float64{
						"crit_damage":   {0.50, 1.50},
						"crit_chance":   {0.10, 0.30},
						"stealth_bonus": {0.20, 0.60},
					},
					Tags:          []string{"combat", "active", "stealth"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Utility skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Lockpick", "Trap", "Sneak", "Pickpocket"},
					NameSuffixes:      []string{"Expert", "Master", "Specialist", "Training"},
					DescriptionFormat: "Enhances %s abilities",
					EffectTypes:       []string{"loot_chance", "gold_find", "detection_range"},
					ValueRanges: map[string][2]float64{
						"loot_chance":     {0.05, 0.15},
						"gold_find":       {0.10, 0.30},
						"detection_range": {-0.10, -0.20},
					},
					Tags:          []string{"utility", "passive", "thief"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Blade", "Death", "Shadow", "Assassin's"},
					NameSuffixes:      []string{"Dance", "Mark", "Cloak", "Calling"},
					DescriptionFormat: "Enter %s mode for incredible speed and damage",
					EffectTypes:       []string{"attack_speed", "crit_chance", "dodge_chance"},
					ValueRanges: map[string][2]float64{
						"attack_speed": {0.80, 1.50},
						"crit_chance":  {0.40, 0.70},
						"dodge_chance": {0.30, 0.50},
					},
					Tags:          []string{"ultimate", "combat", "burst"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
	}
}

// GetHorrorTreeTemplates returns skill tree templates for horror genre.
// Archetypes include Necromancer, Blood Mage, Monster Hunter, and Cultist.
func GetHorrorTreeTemplates() []SkillTreeTemplate {
	return []SkillTreeTemplate{
		{
			Name:        "Necromancer",
			Description: "Master of death and undead summoning",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Dark magic passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Dark", "Death", "Grave", "Corpse"},
					NameSuffixes:      []string{"Affinity", "Mastery", "Attunement", "Knowledge"},
					DescriptionFormat: "Enhances %s and necromantic power",
					EffectTypes:       []string{"dark_damage", "summon_health", "soul_harvest"},
					ValueRanges: map[string][2]float64{
						"dark_damage":   {0.06, 0.18},
						"summon_health": {0.10, 0.25},
						"soul_harvest":  {0.05, 0.12},
					},
					Tags:          []string{"magic", "passive", "dark"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Summoning active skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Raise", "Summon", "Animate", "Call"},
					NameSuffixes:      []string{"Undead", "Skeleton", "Ghoul", "Zombie"},
					DescriptionFormat: "Conjure %s to fight for you",
					EffectTypes:       []string{"summon_count", "summon_damage", "summon_duration"},
					ValueRanges: map[string][2]float64{
						"summon_count":    {1.00, 3.00},
						"summon_damage":   {0.40, 1.00},
						"summon_duration": {15.0, 45.0},
					},
					Tags:          []string{"magic", "active", "summon"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Life drain skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Soul", "Life", "Vitality", "Essence"},
					NameSuffixes:      []string{"Siphon", "Drain", "Leech", "Absorption"},
					DescriptionFormat: "Steals %s from enemies",
					EffectTypes:       []string{"lifesteal", "mana_steal", "health_regen"},
					ValueRanges: map[string][2]float64{
						"lifesteal":    {0.05, 0.15},
						"mana_steal":   {0.03, 0.10},
						"health_regen": {0.08, 0.18},
					},
					Tags:          []string{"defense", "passive", "drain"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Army of", "Legion of", "Horde of", "Plague of"},
					NameSuffixes:      []string{"Darkness", "the Dead", "Undeath", "Shadows"},
					DescriptionFormat: "Raise %s to overwhelm enemies",
					EffectTypes:       []string{"summon_count", "dark_damage", "fear_chance"},
					ValueRanges: map[string][2]float64{
						"summon_count": {5.00, 10.00},
						"dark_damage":  {2.00, 4.00},
						"fear_chance":  {0.30, 0.60},
					},
					Tags:          []string{"ultimate", "magic", "summon"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Blood Mage",
			Description: "Wielder of forbidden blood magic and sacrifice",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Blood magic passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Crimson", "Blood", "Sanguine", "Vital"},
					NameSuffixes:      []string{"Mastery", "Rites", "Binding", "Focus"},
					DescriptionFormat: "Enhances %s power through blood",
					EffectTypes:       []string{"blood_damage", "health_cost_reduction", "spell_power"},
					ValueRanges: map[string][2]float64{
						"blood_damage":          {0.08, 0.22},
						"health_cost_reduction": {0.05, 0.15},
						"spell_power":           {0.06, 0.14},
					},
					Tags:          []string{"magic", "passive", "blood"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Blood sacrifice active skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Blood", "Crimson", "Hemorrhage", "Exsanguinate"},
					NameSuffixes:      []string{"Bolt", "Strike", "Rupture", "Cascade"},
					DescriptionFormat: "Sacrifice health to unleash %s",
					EffectTypes:       []string{"blood_damage", "bleed_chance", "health_cost"},
					ValueRanges: map[string][2]float64{
						"blood_damage": {0.80, 2.00},
						"bleed_chance": {0.30, 0.60},
						"health_cost":  {0.05, 0.15},
					},
					Tags:          []string{"combat", "active", "sacrifice"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Self-sustain through blood
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Transfusion", "Bloodwell", "Vitae", "Regenerative"},
					NameSuffixes:      []string{"", "Reserve", "Pool", "Flow"},
					DescriptionFormat: "Recover through %s",
					EffectTypes:       []string{"lifesteal", "max_health", "health_on_kill"},
					ValueRanges: map[string][2]float64{
						"lifesteal":      {0.08, 0.20},
						"max_health":     {0.10, 0.25},
						"health_on_kill": {0.05, 0.12},
					},
					Tags:          []string{"defense", "passive", "sustain"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Sanguine", "Blood", "Crimson", "Vital"},
					NameSuffixes:      []string{"Apocalypse", "Storm", "Tempest", "Offering"},
					DescriptionFormat: "Unleash %s for devastating damage",
					EffectTypes:       []string{"blood_damage", "lifesteal", "aoe_radius"},
					ValueRanges: map[string][2]float64{
						"blood_damage": {3.00, 6.00},
						"lifesteal":    {0.40, 0.80},
						"aoe_radius":   {0.50, 1.00},
					},
					Tags:          []string{"ultimate", "magic", "burst"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Monster Hunter",
			Description: "Specialist in tracking and slaying supernatural creatures",
			Category:    CategoryCombat,
			SkillTemplates: []SkillTemplate{
				// Tracking passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Beast", "Monster", "Creature", "Prey"},
					NameSuffixes:      []string{"Tracker", "Sense", "Detection", "Knowledge"},
					DescriptionFormat: "Improves %s against supernatural foes",
					EffectTypes:       []string{"detection_range", "crit_vs_monsters", "damage_vs_monsters"},
					ValueRanges: map[string][2]float64{
						"detection_range":    {0.10, 0.30},
						"crit_vs_monsters":   {0.05, 0.15},
						"damage_vs_monsters": {0.08, 0.20},
					},
					Tags:          []string{"utility", "passive", "tracking"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Combat active skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Silver", "Holy", "Blessed", "Consecrated"},
					NameSuffixes:      []string{"Strike", "Blade", "Arrow", "Bolt"},
					DescriptionFormat: "Attack with %s for bonus damage",
					EffectTypes:       []string{"holy_damage", "stun_chance", "weakness_exploit"},
					ValueRanges: map[string][2]float64{
						"holy_damage":      {0.60, 1.40},
						"stun_chance":      {0.15, 0.35},
						"weakness_exploit": {0.20, 0.50},
					},
					Tags:          []string{"combat", "active", "holy"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Preparation skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Warded", "Protected", "Resistant", "Hardened"},
					NameSuffixes:      []string{"Mind", "Soul", "Body", "Spirit"},
					DescriptionFormat: "Grants %s against supernatural attacks",
					EffectTypes:       []string{"fear_resistance", "curse_resistance", "dark_resistance"},
					ValueRanges: map[string][2]float64{
						"fear_resistance":  {0.15, 0.40},
						"curse_resistance": {0.10, 0.30},
						"dark_resistance":  {0.10, 0.25},
					},
					Tags:          []string{"defense", "passive", "resistance"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Purge", "Exorcism", "Banishment", "Annihilation"},
					NameSuffixes:      []string{"Protocol", "Rite", "Ritual", "Strike"},
					DescriptionFormat: "Execute %s to devastate supernatural foes",
					EffectTypes:       []string{"holy_damage", "instant_kill_chance", "aoe_radius"},
					ValueRanges: map[string][2]float64{
						"holy_damage":         {2.50, 5.00},
						"instant_kill_chance": {0.10, 0.25},
						"aoe_radius":          {0.40, 0.80},
					},
					Tags:          []string{"ultimate", "combat", "execution"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Cultist",
			Description: "Channeler of eldritch powers and cosmic horror",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Eldritch passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Eldritch", "Cosmic", "Void", "Abyssal"},
					NameSuffixes:      []string{"Knowledge", "Insight", "Whispers", "Attunement"},
					DescriptionFormat: "Deepens %s connection to the beyond",
					EffectTypes:       []string{"void_damage", "madness_resistance", "spell_power"},
					ValueRanges: map[string][2]float64{
						"void_damage":        {0.08, 0.20},
						"madness_resistance": {0.10, 0.25},
						"spell_power":        {0.06, 0.14},
					},
					Tags:          []string{"magic", "passive", "eldritch"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Madness active skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Mind", "Sanity", "Reality", "Perception"},
					NameSuffixes:      []string{"Shatter", "Break", "Rend", "Twist"},
					DescriptionFormat: "Inflict %s on enemies",
					EffectTypes:       []string{"madness_damage", "confusion_chance", "fear_chance"},
					ValueRanges: map[string][2]float64{
						"madness_damage":   {0.50, 1.30},
						"confusion_chance": {0.20, 0.45},
						"fear_chance":      {0.15, 0.35},
					},
					Tags:          []string{"magic", "active", "control"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Tentacle/corruption skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Otherworldly", "Aberrant", "Twisted", "Corrupted"},
					NameSuffixes:      []string{"Form", "Flesh", "Body", "Vessel"},
					DescriptionFormat: "Transform into %s for protection",
					EffectTypes:       []string{"damage_reduction", "tentacle_counter", "regeneration"},
					ValueRanges: map[string][2]float64{
						"damage_reduction": {0.08, 0.18},
						"tentacle_counter": {0.10, 0.25},
						"regeneration":     {0.05, 0.12},
					},
					Tags:          []string{"defense", "passive", "transform"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Call of", "Gate to", "Avatar of", "Manifestation of"},
					NameSuffixes:      []string{"Cthulhu", "the Void", "Madness", "the Beyond"},
					DescriptionFormat: "Invoke %s to devastate all",
					EffectTypes:       []string{"void_damage", "madness_aoe", "summon_horror"},
					ValueRanges: map[string][2]float64{
						"void_damage":   {2.50, 5.50},
						"madness_aoe":   {0.60, 1.00},
						"summon_horror": {1.00, 2.00},
					},
					Tags:          []string{"ultimate", "magic", "eldritch"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
	}
}

// GetCyberpunkTreeTemplates returns skill tree templates for cyberpunk genre.
// Archetypes include Netrunner, Street Samurai, Technomancer, and Corporate Infiltrator.
func GetCyberpunkTreeTemplates() []SkillTreeTemplate {
	return []SkillTreeTemplate{
		{
			Name:        "Netrunner",
			Description: "Master hacker and digital warfare specialist",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Hacking passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Neural", "ICE", "Daemon", "Protocol"},
					NameSuffixes:      []string{"Interface", "Mastery", "Optimization", "Bypass"},
					DescriptionFormat: "Enhances %s hacking capabilities",
					EffectTypes:       []string{"hack_damage", "breach_speed", "firewall_bypass"},
					ValueRanges: map[string][2]float64{
						"hack_damage":     {0.06, 0.18},
						"breach_speed":    {0.08, 0.20},
						"firewall_bypass": {0.05, 0.15},
					},
					Tags:          []string{"magic", "passive", "hacking"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active hacking skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"System", "Neural", "Data", "Memory"},
					NameSuffixes:      []string{"Crash", "Spike", "Wipe", "Overload"},
					DescriptionFormat: "Execute %s against enemy systems",
					EffectTypes:       []string{"hack_damage", "disable_duration", "spread_chance"},
					ValueRanges: map[string][2]float64{
						"hack_damage":      {0.60, 1.60},
						"disable_duration": {2.0, 6.0},
						"spread_chance":    {0.15, 0.40},
					},
					Tags:          []string{"magic", "active", "offensive"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Defensive hacking
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Firewall", "Encryption", "Counter", "Trace"},
					NameSuffixes:      []string{"Matrix", "Protocol", "Defense", "Shield"},
					DescriptionFormat: "Protects against %s cyber attacks",
					EffectTypes:       []string{"hack_resistance", "trace_immunity", "data_integrity"},
					ValueRanges: map[string][2]float64{
						"hack_resistance": {0.10, 0.25},
						"trace_immunity":  {0.08, 0.20},
						"data_integrity":  {0.06, 0.15},
					},
					Tags:          []string{"defense", "passive", "security"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Blackout", "Total", "Grid", "Network"},
					NameSuffixes:      []string{"Protocol", "Collapse", "Shutdown", "Apocalypse"},
					DescriptionFormat: "Unleash %s to devastate all connected systems",
					EffectTypes:       []string{"hack_damage", "aoe_disable", "system_corruption"},
					ValueRanges: map[string][2]float64{
						"hack_damage":       {2.50, 5.00},
						"aoe_disable":       {5.0, 12.0},
						"system_corruption": {0.40, 0.80},
					},
					Tags:          []string{"ultimate", "magic", "aoe"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Street Samurai",
			Description: "Cybernetically enhanced close combat specialist",
			Category:    CategoryCombat,
			SkillTemplates: []SkillTemplate{
				// Combat passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Wired", "Chrome", "Cyber", "Reflex"},
					NameSuffixes:      []string{"Reflexes", "Enhancement", "Boost", "Augmentation"},
					DescriptionFormat: "Enhances %s through cybernetic upgrades",
					EffectTypes:       []string{"attack_speed", "crit_chance", "reaction_time"},
					ValueRanges: map[string][2]float64{
						"attack_speed":  {0.08, 0.20},
						"crit_chance":   {0.05, 0.15},
						"reaction_time": {0.06, 0.16},
					},
					Tags:          []string{"combat", "passive", "cybernetic"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active combat skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Mantis", "Mono", "Vibro", "Plasma"},
					NameSuffixes:      []string{"Blade Strike", "Wire Slice", "Edge Combo", "Claw Assault"},
					DescriptionFormat: "Execute %s with cybernetic weapons",
					EffectTypes:       []string{"damage", "bleed_chance", "armor_pierce"},
					ValueRanges: map[string][2]float64{
						"damage":       {0.70, 1.80},
						"bleed_chance": {0.20, 0.50},
						"armor_pierce": {0.15, 0.40},
					},
					Tags:          []string{"combat", "active", "melee"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Defensive augmentations
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Subdermal", "Titanium", "Nano", "Reinforced"},
					NameSuffixes:      []string{"Armor", "Plating", "Weave", "Skeleton"},
					DescriptionFormat: "Provides %s cybernetic protection",
					EffectTypes:       []string{"armor", "damage_reduction", "health_boost"},
					ValueRanges: map[string][2]float64{
						"armor":            {0.08, 0.20},
						"damage_reduction": {0.05, 0.15},
						"health_boost":     {0.10, 0.25},
					},
					Tags:          []string{"defense", "passive", "augmentation"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Sandevistan", "Kerenzikov", "Berserk", "Combat"},
					NameSuffixes:      []string{"Overdrive", "Protocol", "Mode", "Surge"},
					DescriptionFormat: "Activate %s for superhuman combat speed",
					EffectTypes:       []string{"time_dilation", "damage_boost", "attack_speed"},
					ValueRanges: map[string][2]float64{
						"time_dilation": {0.50, 1.00},
						"damage_boost":  {1.50, 3.00},
						"attack_speed":  {1.00, 2.00},
					},
					Tags:          []string{"ultimate", "combat", "burst"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Technomancer",
			Description: "Controller of drones, robots, and smart technology",
			Category:    CategoryUtility,
			SkillTemplates: []SkillTemplate{
				// Drone control passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Drone", "Bot", "AI", "Smart"},
					NameSuffixes:      []string{"Link", "Control", "Mastery", "Network"},
					DescriptionFormat: "Improves %s and autonomous systems",
					EffectTypes:       []string{"drone_damage", "drone_count", "control_range"},
					ValueRanges: map[string][2]float64{
						"drone_damage":  {0.08, 0.20},
						"drone_count":   {0.5, 1.5},
						"control_range": {0.10, 0.30},
					},
					Tags:          []string{"utility", "passive", "drone"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active drone deployment
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Attack", "Recon", "Medical", "Shield"},
					NameSuffixes:      []string{"Drone", "Bot", "Unit", "Swarm"},
					DescriptionFormat: "Deploy %s to assist in combat",
					EffectTypes:       []string{"drone_damage", "drone_health", "special_ability"},
					ValueRanges: map[string][2]float64{
						"drone_damage":    {0.50, 1.30},
						"drone_health":    {0.60, 1.40},
						"special_ability": {0.20, 0.50},
					},
					Tags:          []string{"utility", "active", "summon"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Tech crafting
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCrafting,
					NamePrefixes:      []string{"Advanced", "Nano", "Smart", "Modular"},
					NameSuffixes:      []string{"Engineering", "Fabrication", "Assembly", "Design"},
					DescriptionFormat: "Enhances %s crafting abilities",
					EffectTypes:       []string{"craft_speed", "component_efficiency", "upgrade_bonus"},
					ValueRanges: map[string][2]float64{
						"craft_speed":          {0.10, 0.25},
						"component_efficiency": {0.08, 0.20},
						"upgrade_bonus":        {0.05, 0.15},
					},
					Tags:          []string{"crafting", "passive", "tech"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Mech", "Titan", "Warbot", "Assault"},
					NameSuffixes:      []string{"Deployment", "Summon", "Protocol", "Override"},
					DescriptionFormat: "Deploy %s for overwhelming firepower",
					EffectTypes:       []string{"mech_damage", "mech_armor", "mech_duration"},
					ValueRanges: map[string][2]float64{
						"mech_damage":   {2.00, 4.00},
						"mech_armor":    {1.00, 2.00},
						"mech_duration": {15.0, 30.0},
					},
					Tags:          []string{"ultimate", "utility", "summon"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Corporate Infiltrator",
			Description: "Master of deception, social engineering, and covert ops",
			Category:    CategoryUtility,
			SkillTemplates: []SkillTemplate{
				// Stealth passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Optical", "Sound", "Thermal", "Signature"},
					NameSuffixes:      []string{"Camo", "Dampening", "Masking", "Suppression"},
					DescriptionFormat: "Improves %s and infiltration",
					EffectTypes:       []string{"stealth_bonus", "detection_reduction", "move_speed"},
					ValueRanges: map[string][2]float64{
						"stealth_bonus":       {0.10, 0.25},
						"detection_reduction": {0.08, 0.20},
						"move_speed":          {0.05, 0.12},
					},
					Tags:          []string{"utility", "passive", "stealth"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active infiltration skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Silent", "Precision", "Lethal", "Executive"},
					NameSuffixes:      []string{"Takedown", "Elimination", "Strike", "Neutralization"},
					DescriptionFormat: "Execute %s from stealth",
					EffectTypes:       []string{"stealth_damage", "instant_kill_chance", "silence_duration"},
					ValueRanges: map[string][2]float64{
						"stealth_damage":      {1.00, 2.50},
						"instant_kill_chance": {0.10, 0.30},
						"silence_duration":    {3.0, 8.0},
					},
					Tags:          []string{"combat", "active", "assassination"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Social engineering
				{
					BaseType:          TypePassive,
					BaseCategory:      CategorySocial,
					NamePrefixes:      []string{"Corporate", "Street", "Black Market", "Elite"},
					NameSuffixes:      []string{"Connections", "Access", "Contacts", "Network"},
					DescriptionFormat: "Grants %s social advantages",
					EffectTypes:       []string{"bribe_discount", "intel_bonus", "reputation_gain"},
					ValueRanges: map[string][2]float64{
						"bribe_discount":  {0.10, 0.30},
						"intel_bonus":     {0.08, 0.20},
						"reputation_gain": {0.10, 0.25},
					},
					Tags:          []string{"social", "passive", "influence"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Ghost", "Phantom", "Shadow", "Specter"},
					NameSuffixes:      []string{"Protocol", "Mode", "Strike", "Operation"},
					DescriptionFormat: "Activate %s for perfect infiltration",
					EffectTypes:       []string{"invisibility_duration", "damage_bonus", "escape_chance"},
					ValueRanges: map[string][2]float64{
						"invisibility_duration": {8.0, 15.0},
						"damage_bonus":          {2.00, 4.00},
						"escape_chance":         {0.50, 1.00},
					},
					Tags:          []string{"ultimate", "utility", "stealth"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
	}
}

// GetPostApocalypticTreeTemplates returns skill tree templates for post-apocalyptic genre.
// Archetypes include Scavenger, Raider, Survivor, and Mutant.
func GetPostApocalypticTreeTemplates() []SkillTreeTemplate {
	return []SkillTreeTemplate{
		{
			Name:        "Scavenger",
			Description: "Expert in finding resources and crafting improvised equipment",
			Category:    CategoryUtility,
			SkillTemplates: []SkillTemplate{
				// Resource finding passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Keen", "Trained", "Wasteland", "Salvage"},
					NameSuffixes:      []string{"Eye", "Instinct", "Sense", "Expertise"},
					DescriptionFormat: "Improves %s resource discovery",
					EffectTypes:       []string{"loot_chance", "resource_find", "detection_range"},
					ValueRanges: map[string][2]float64{
						"loot_chance":     {0.08, 0.20},
						"resource_find":   {0.10, 0.25},
						"detection_range": {0.06, 0.15},
					},
					Tags:          []string{"utility", "passive", "scavenging"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active scavenging skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Quick", "Deep", "Thorough", "Expert"},
					NameSuffixes:      []string{"Search", "Salvage", "Scrounge", "Excavation"},
					DescriptionFormat: "Perform %s to find extra resources",
					EffectTypes:       []string{"bonus_loot", "rare_find_chance", "speed_bonus"},
					ValueRanges: map[string][2]float64{
						"bonus_loot":       {0.30, 0.80},
						"rare_find_chance": {0.10, 0.30},
						"speed_bonus":      {0.15, 0.40},
					},
					Tags:          []string{"utility", "active", "loot"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Crafting skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCrafting,
					NamePrefixes:      []string{"Improvised", "Jury-Rigged", "Makeshift", "Cobbled"},
					NameSuffixes:      []string{"Engineering", "Crafting", "Assembly", "Fabrication"},
					DescriptionFormat: "Enhances %s improvised crafting",
					EffectTypes:       []string{"craft_speed", "material_efficiency", "repair_bonus"},
					ValueRanges: map[string][2]float64{
						"craft_speed":         {0.10, 0.25},
						"material_efficiency": {0.15, 0.35},
						"repair_bonus":        {0.12, 0.28},
					},
					Tags:          []string{"crafting", "passive", "improvisation"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Master", "Legendary", "Ultimate", "Supreme"},
					NameSuffixes:      []string{"Scavenger", "Salvager", "Finder", "Collector"},
					DescriptionFormat: "Activate %s for incredible resource discovery",
					EffectTypes:       []string{"guaranteed_rare", "loot_multiplier", "area_reveal"},
					ValueRanges: map[string][2]float64{
						"guaranteed_rare": {1.00, 3.00},
						"loot_multiplier": {2.00, 4.00},
						"area_reveal":     {0.50, 1.00},
					},
					Tags:          []string{"ultimate", "utility", "loot"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Raider",
			Description: "Aggressive combatant focused on intimidation and quick strikes",
			Category:    CategoryCombat,
			SkillTemplates: []SkillTemplate{
				// Aggression passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Brutal", "Savage", "Vicious", "Relentless"},
					NameSuffixes:      []string{"Assault", "Aggression", "Fury", "Onslaught"},
					DescriptionFormat: "Enhances %s offensive capabilities",
					EffectTypes:       []string{"damage", "crit_chance", "attack_speed"},
					ValueRanges: map[string][2]float64{
						"damage":       {0.08, 0.20},
						"crit_chance":  {0.05, 0.14},
						"attack_speed": {0.06, 0.15},
					},
					Tags:          []string{"combat", "passive", "aggression"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active combat skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Charge", "Ambush", "Raid", "Blitz"},
					NameSuffixes:      []string{"Attack", "Strike", "Assault", "Rush"},
					DescriptionFormat: "Execute %s for devastating damage",
					EffectTypes:       []string{"damage", "stun_chance", "fear_chance"},
					ValueRanges: map[string][2]float64{
						"damage":      {0.70, 1.80},
						"stun_chance": {0.15, 0.40},
						"fear_chance": {0.10, 0.30},
					},
					Tags:          []string{"combat", "active", "burst"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Intimidation skills
				{
					BaseType:          TypePassive,
					BaseCategory:      CategorySocial,
					NamePrefixes:      []string{"Fearsome", "Terrifying", "Menacing", "Dreadful"},
					NameSuffixes:      []string{"Presence", "Reputation", "Aura", "Visage"},
					DescriptionFormat: "Grants %s intimidation bonuses",
					EffectTypes:       []string{"fear_aura", "enemy_damage_reduction", "barter_bonus"},
					ValueRanges: map[string][2]float64{
						"fear_aura":              {0.10, 0.25},
						"enemy_damage_reduction": {0.05, 0.15},
						"barter_bonus":           {0.08, 0.20},
					},
					Tags:          []string{"social", "passive", "intimidation"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Warlord's", "Marauder's", "Berserker's", "Devastator's"},
					NameSuffixes:      []string{"Rampage", "Fury", "Massacre", "Carnage"},
					DescriptionFormat: "Unleash %s for overwhelming destruction",
					EffectTypes:       []string{"damage", "fear_aoe", "lifesteal"},
					ValueRanges: map[string][2]float64{
						"damage":    {2.50, 5.00},
						"fear_aoe":  {0.60, 1.00},
						"lifesteal": {0.30, 0.60},
					},
					Tags:          []string{"ultimate", "combat", "burst"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Survivor",
			Description: "Master of endurance, adaptation, and environmental resistance",
			Category:    CategoryDefense,
			SkillTemplates: []SkillTemplate{
				// Endurance passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Hardened", "Weathered", "Resilient", "Tough"},
					NameSuffixes:      []string{"Constitution", "Endurance", "Fortitude", "Resolve"},
					DescriptionFormat: "Improves %s survival capabilities",
					EffectTypes:       []string{"max_health", "health_regen", "damage_reduction"},
					ValueRanges: map[string][2]float64{
						"max_health":       {0.10, 0.25},
						"health_regen":     {0.08, 0.18},
						"damage_reduction": {0.05, 0.14},
					},
					Tags:          []string{"defense", "passive", "survival"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active survival skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Emergency", "Desperate", "Last Stand", "Second Wind"},
					NameSuffixes:      []string{"Recovery", "Heal", "Surge", "Revival"},
					DescriptionFormat: "Activate %s for emergency healing",
					EffectTypes:       []string{"instant_heal", "heal_over_time", "damage_immunity"},
					ValueRanges: map[string][2]float64{
						"instant_heal":    {0.20, 0.50},
						"heal_over_time":  {0.15, 0.40},
						"damage_immunity": {2.0, 5.0},
					},
					Tags:          []string{"defense", "active", "healing"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Environmental resistance
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Radiation", "Toxin", "Heat", "Cold"},
					NameSuffixes:      []string{"Resistance", "Immunity", "Adaptation", "Tolerance"},
					DescriptionFormat: "Grants %s environmental protection",
					EffectTypes:       []string{"radiation_resist", "poison_resist", "elemental_resist"},
					ValueRanges: map[string][2]float64{
						"radiation_resist": {0.15, 0.40},
						"poison_resist":    {0.12, 0.35},
						"elemental_resist": {0.10, 0.28},
					},
					Tags:          []string{"defense", "passive", "resistance"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Undying", "Immortal", "Unkillable", "Invincible"},
					NameSuffixes:      []string{"Will", "Spirit", "Determination", "Resolve"},
					DescriptionFormat: "Activate %s for near-immortality",
					EffectTypes:       []string{"damage_immunity_duration", "full_heal", "buff_all_stats"},
					ValueRanges: map[string][2]float64{
						"damage_immunity_duration": {5.0, 10.0},
						"full_heal":                {1.00, 1.00},
						"buff_all_stats":           {0.30, 0.60},
					},
					Tags:          []string{"ultimate", "defense", "survival"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Mutant",
			Description: "Radiation-altered being with unnatural powers and resilience",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Mutation passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Unstable", "Evolved", "Mutated", "Altered"},
					NameSuffixes:      []string{"Genes", "DNA", "Cells", "Biology"},
					DescriptionFormat: "Enhances %s mutant abilities",
					EffectTypes:       []string{"mutation_power", "regeneration", "radiation_absorb"},
					ValueRanges: map[string][2]float64{
						"mutation_power":   {0.08, 0.20},
						"regeneration":     {0.10, 0.25},
						"radiation_absorb": {0.12, 0.30},
					},
					Tags:          []string{"magic", "passive", "mutation"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active mutation powers
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Radioactive", "Toxic", "Corrosive", "Necrotic"},
					NameSuffixes:      []string{"Blast", "Wave", "Pulse", "Eruption"},
					DescriptionFormat: "Unleash %s at enemies",
					EffectTypes:       []string{"radiation_damage", "poison_damage", "debuff_chance"},
					ValueRanges: map[string][2]float64{
						"radiation_damage": {0.60, 1.60},
						"poison_damage":    {0.40, 1.00},
						"debuff_chance":    {0.20, 0.50},
					},
					Tags:          []string{"magic", "active", "radiation"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Physical mutations
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Clawed", "Armored", "Enhanced", "Monstrous"},
					NameSuffixes:      []string{"Limbs", "Hide", "Physique", "Form"},
					DescriptionFormat: "Grants %s physical enhancements",
					EffectTypes:       []string{"melee_damage", "natural_armor", "strength_bonus"},
					ValueRanges: map[string][2]float64{
						"melee_damage":   {0.10, 0.25},
						"natural_armor":  {0.08, 0.20},
						"strength_bonus": {0.06, 0.16},
					},
					Tags:          []string{"combat", "passive", "physical"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Meltdown", "Unstable", "Critical", "Nuclear"},
					NameSuffixes:      []string{"Form", "Transformation", "Evolution", "Apotheosis"},
					DescriptionFormat: "Trigger %s for devastating power",
					EffectTypes:       []string{"radiation_aoe", "mutation_buff", "terror_aura"},
					ValueRanges: map[string][2]float64{
						"radiation_aoe": {2.50, 5.00},
						"mutation_buff": {0.50, 1.00},
						"terror_aura":   {0.60, 1.00},
					},
					Tags:          []string{"ultimate", "magic", "transformation"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
	}
}

// GetSciFiTreeTemplates returns skill tree templates for sci-fi genre.
func GetSciFiTreeTemplates() []SkillTreeTemplate {
	return []SkillTreeTemplate{
		{
			Name:        "Soldier",
			Description: "Combat specialist with advanced weaponry",
			Category:    CategoryCombat,
			SkillTemplates: []SkillTemplate{
				// Weapon passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Ballistic", "Plasma", "Rail", "Laser"},
					NameSuffixes:      []string{"Training", "Proficiency", "Expertise", "Mastery"},
					DescriptionFormat: "Improves %s weapon effectiveness",
					EffectTypes:       []string{"damage", "accuracy", "fire_rate"},
					ValueRanges: map[string][2]float64{
						"damage":    {0.06, 0.16},
						"accuracy":  {0.03, 0.10},
						"fire_rate": {0.05, 0.12},
					},
					Tags:          []string{"combat", "passive", "weapons"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active combat
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Grenade", "Rocket", "Missile", "Mine"},
					NameSuffixes:      []string{"Launcher", "Barrage", "Strike", "Deploy"},
					DescriptionFormat: "Deploy %s for explosive damage",
					EffectTypes:       []string{"explosion_damage", "aoe_radius", "armor_pierce"},
					ValueRanges: map[string][2]float64{
						"explosion_damage": {0.60, 1.60},
						"aoe_radius":       {0.20, 0.50},
						"armor_pierce":     {0.10, 0.30},
					},
					Tags:          []string{"combat", "active", "explosive"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Defense
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Kinetic", "Energy", "Reactive", "Ablative"},
					NameSuffixes:      []string{"Armor", "Shields", "Plating", "Barrier"},
					DescriptionFormat: "Enhances %s defensive capabilities",
					EffectTypes:       []string{"armor", "shield_capacity", "regen_rate"},
					ValueRanges: map[string][2]float64{
						"armor":           {0.06, 0.16},
						"shield_capacity": {0.12, 0.28},
						"regen_rate":      {0.08, 0.18},
					},
					Tags:          []string{"defense", "passive", "tank"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryCombat,
					NamePrefixes:      []string{"Tactical", "Orbital", "Nuclear", "Antimatter"},
					NameSuffixes:      []string{"Strike", "Bombardment", "Warhead", "Payload"},
					DescriptionFormat: "Call in %s for devastating destruction",
					EffectTypes:       []string{"damage", "aoe_radius", "armor_shred"},
					ValueRanges: map[string][2]float64{
						"damage":      {2.50, 5.00},
						"aoe_radius":  {0.80, 1.50},
						"armor_shred": {0.30, 0.60},
					},
					Tags:          []string{"ultimate", "combat", "aoe"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Engineer",
			Description: "Tech specialist with gadgets and turrets",
			Category:    CategoryUtility,
			SkillTemplates: []SkillTemplate{
				// Tech passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Tech", "Mechanical", "Electronic", "System"},
					NameSuffixes:      []string{"Affinity", "Expertise", "Knowledge", "Mastery"},
					DescriptionFormat: "Improves %s and gadget effectiveness",
					EffectTypes:       []string{"tech_bonus", "cooldown_reduction", "efficiency"},
					ValueRanges: map[string][2]float64{
						"tech_bonus":         {0.08, 0.18},
						"cooldown_reduction": {0.05, 0.12},
						"efficiency":         {0.06, 0.14},
					},
					Tags:          []string{"utility", "passive", "tech"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Turret skills
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Auto", "Plasma", "Laser", "Shock"},
					NameSuffixes:      []string{"Turret", "Sentry", "Drone", "Bot"},
					DescriptionFormat: "Deploy %s to attack enemies",
					EffectTypes:       []string{"turret_damage", "turret_health", "deploy_speed"},
					ValueRanges: map[string][2]float64{
						"turret_damage": {0.40, 1.20},
						"turret_health": {0.50, 1.50},
						"deploy_speed":  {-0.10, -0.30},
					},
					Tags:          []string{"utility", "active", "summon"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Crafting
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryCrafting,
					NamePrefixes:      []string{"Advanced", "Efficient", "Master", "Expert"},
					NameSuffixes:      []string{"Fabrication", "Crafting", "Engineering", "Assembly"},
					DescriptionFormat: "Enhances %s abilities",
					EffectTypes:       []string{"craft_speed", "resource_efficiency", "quality_bonus"},
					ValueRanges: map[string][2]float64{
						"craft_speed":         {0.10, 0.25},
						"resource_efficiency": {0.05, 0.15},
						"quality_bonus":       {0.05, 0.12},
					},
					Tags:          []string{"crafting", "passive", "utility"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryUtility,
					NamePrefixes:      []string{"Mech", "Power", "Combat", "Titan"},
					NameSuffixes:      []string{"Suit", "Armor", "Frame", "Exosuit"},
					DescriptionFormat: "Deploy %s for enhanced combat capabilities",
					EffectTypes:       []string{"damage", "armor", "ability_power"},
					ValueRanges: map[string][2]float64{
						"damage":        {1.00, 2.50},
						"armor":         {0.80, 1.50},
						"ability_power": {0.50, 1.00},
					},
					Tags:          []string{"ultimate", "utility", "transform"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
		{
			Name:        "Biotic",
			Description: "Psionic specialist with mind powers",
			Category:    CategoryMagic,
			SkillTemplates: []SkillTemplate{
				// Psionic passives
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Psionic", "Neural", "Mental", "Telepathic"},
					NameSuffixes:      []string{"Amplifier", "Enhancement", "Focus", "Discipline"},
					DescriptionFormat: "Boosts %s abilities",
					EffectTypes:       []string{"psi_power", "psi_regen", "mental_fortitude"},
					ValueRanges: map[string][2]float64{
						"psi_power":        {0.08, 0.20},
						"psi_regen":        {0.06, 0.15},
						"mental_fortitude": {0.05, 0.12},
					},
					Tags:          []string{"magic", "passive", "psionic"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Active powers
				{
					BaseType:          TypeActive,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Mind", "Psychic", "Telekinetic", "Neural"},
					NameSuffixes:      []string{"Blast", "Wave", "Storm", "Shock"},
					DescriptionFormat: "Release %s to damage and control enemies",
					EffectTypes:       []string{"psi_damage", "crowd_control", "shield_damage"},
					ValueRanges: map[string][2]float64{
						"psi_damage":    {0.50, 1.50},
						"crowd_control": {0.20, 0.50},
						"shield_damage": {0.30, 0.80},
					},
					Tags:          []string{"magic", "active", "control"},
					TierRange:     [2]int{1, 5},
					MaxLevelRange: [2]int{1, 3},
				},
				// Support
				{
					BaseType:          TypePassive,
					BaseCategory:      CategoryDefense,
					NamePrefixes:      []string{"Biotic", "Kinetic", "Protective", "Mental"},
					NameSuffixes:      []string{"Barrier", "Shield", "Ward", "Defense"},
					DescriptionFormat: "Creates %s for protection",
					EffectTypes:       []string{"shield_strength", "damage_absorption", "regen"},
					ValueRanges: map[string][2]float64{
						"shield_strength":   {0.15, 0.35},
						"damage_absorption": {0.08, 0.18},
						"regen":             {0.10, 0.20},
					},
					Tags:          []string{"defense", "passive", "shield"},
					TierRange:     [2]int{0, 4},
					MaxLevelRange: [2]int{3, 5},
				},
				// Ultimate
				{
					BaseType:          TypeUltimate,
					BaseCategory:      CategoryMagic,
					NamePrefixes:      []string{"Singularity", "Vortex", "Warp", "Stasis"},
					NameSuffixes:      []string{"Field", "Collapse", "Cascade", "Breach"},
					DescriptionFormat: "Create %s to devastate all enemies",
					EffectTypes:       []string{"psi_damage", "aoe_radius", "duration"},
					ValueRanges: map[string][2]float64{
						"psi_damage": {2.00, 5.00},
						"aoe_radius": {0.60, 1.20},
						"duration":   {3.00, 8.00},
					},
					Tags:          []string{"ultimate", "magic", "aoe"},
					TierRange:     [2]int{6, 6},
					MaxLevelRange: [2]int{1, 1},
				},
			},
		},
	}
}
