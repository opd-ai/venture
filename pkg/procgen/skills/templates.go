// Package skills provides skill tree templates.
// This file defines skill tree templates used by the skill generator
// to create structured progression systems.
package skills


// SkillTreeTemplate defines a template for generating a complete skill tree.
type SkillTreeTemplate struct {
	Name           string
	Description    string
	Category       SkillCategory
	SkillTemplates []SkillTemplate
}

// GetFantasyTreeTemplates returns skill tree templates for fantasy genre.
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
