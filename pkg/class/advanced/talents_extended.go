package advanced

// createBerserkerTalentTree returns the talent tree for the Berserker class
func createBerserkerTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Berserker Talents",
		ClassID: ClassBerserker,
		Offensive: []TalentDefinition{
			{ID: "berserker_frenzy", Name: "Frenzy", Description: "Attack speed increases with damage taken", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 3, Speed: 0.05}},
			{ID: "berserker_reckless_blow", Name: "Reckless Blow", Description: "Devastating but dangerous attacks", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 4, CritDamage: 0.1}},
			{ID: "berserker_bloodrage", Name: "Blood Rage", Description: "Gain power from wounds", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_frenzy"}, Bonuses: StatBonuses{Strength: 5, CritChance: 0.03}},
			{ID: "berserker_savage_strikes", Name: "Savage Strikes", Description: "Brutal combo attacks", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 3, Dexterity: 2}},
			{ID: "berserker_rampage", Name: "Rampage", Description: "Chain kills refresh abilities", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_bloodrage"}, Bonuses: StatBonuses{Strength: 6, Stamina: 20}},
			{ID: "berserker_wild_fury", Name: "Wild Fury", Description: "Uncontrolled destruction", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_reckless_blow"}, Bonuses: StatBonuses{Strength: 5, CritDamage: 0.15}},
			{ID: "berserker_carnage", Name: "Carnage", Description: "Area damage from rage", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_rampage"}, Bonuses: StatBonuses{Strength: 7}},
			{ID: "berserker_war_cry", Name: "War Cry", Description: "Terrifying battle scream", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_savage_strikes"}, Bonuses: StatBonuses{Strength: 4, Charisma: 3}},
			{ID: "berserker_deathwish", Name: "Deathwish", Description: "More damage at low health", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_wild_fury"}, Bonuses: StatBonuses{Strength: 8, CritDamage: 0.25}},
			{ID: "berserker_unstoppable_fury", Name: "Unstoppable Fury", Description: "Ultimate berserk state", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_carnage", "berserker_deathwish"}, Bonuses: StatBonuses{Strength: 10, CritDamage: 0.3, Speed: 0.1}},
		},
		Defensive: []TalentDefinition{
			{ID: "berserker_thick_skin", Name: "Thick Skin", Description: "Natural armor from scars", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 2, Health: 15}},
			{ID: "berserker_pain_tolerance", Name: "Pain Tolerance", Description: "Ignore minor wounds", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Health: 25}},
			{ID: "berserker_battle_scars", Name: "Battle Scars", Description: "Each wound makes you stronger", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_thick_skin"}, Bonuses: StatBonuses{Health: 30, Defense: 3}},
			{ID: "berserker_blood_bond", Name: "Blood Bond", Description: "Heal from dealing damage", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_pain_tolerance"}, Bonuses: StatBonuses{Health: 40}},
			{ID: "berserker_ignore_pain", Name: "Ignore Pain", Description: "Delay damage taken", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_blood_bond"}, Bonuses: StatBonuses{Health: 50, Stamina: 20}},
			{ID: "berserker_enraged_regeneration", Name: "Enraged Regeneration", Description: "Heal faster in combat", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"berserker_battle_scars"}, Bonuses: StatBonuses{Health: 35, Stamina: 15}},
			{ID: "berserker_diehard", Name: "Diehard", Description: "Fight on at critical health", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"berserker_ignore_pain"}, Bonuses: StatBonuses{Health: 75}},
			{ID: "berserker_intimidating_presence", Name: "Intimidating Presence", Description: "Enemies hesitate to attack", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Charisma: 4, Defense: 2}},
			{ID: "berserker_fearless", Name: "Fearless", Description: "Immune to fear effects", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_intimidating_presence"}, Bonuses: StatBonuses{Wisdom: 3, Health: 20}},
			{ID: "berserker_undying_rage", Name: "Undying Rage", Description: "Cannot die while enraged", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"berserker_diehard", "berserker_fearless"}, Bonuses: StatBonuses{Health: 100, Strength: 5}},
		},
		Utility: []TalentDefinition{
			{ID: "berserker_battle_trance", Name: "Battle Trance", Description: "Extended rage duration", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Stamina: 20, Wisdom: 1}},
			{ID: "berserker_intimidating_shout", Name: "Intimidating Shout", Description: "Fear nearby enemies", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Charisma: 5}},
			{ID: "berserker_endless_rage", Name: "Endless Rage", Description: "Generate rage faster", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"berserker_battle_trance"}, Bonuses: StatBonuses{Stamina: 25, Strength: 2}},
			{ID: "berserker_relentless", Name: "Relentless", Description: "Cannot be slowed", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Speed: 0.1, Stamina: 15}},
			{ID: "berserker_war_stomp", Name: "War Stomp", Description: "Stun nearby enemies", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"berserker_intimidating_shout"}, Bonuses: StatBonuses{Strength: 3, Charisma: 3}},
			{ID: "berserker_primal_instinct", Name: "Primal Instinct", Description: "Sense nearby danger", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"berserker_relentless"}, Bonuses: StatBonuses{Wisdom: 4, Dexterity: 2}},
			{ID: "berserker_momentum", Name: "Momentum", Description: "Build speed in combat", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"berserker_endless_rage"}, Bonuses: StatBonuses{Speed: 0.15, Stamina: 20}},
			{ID: "berserker_bloodlust_aura", Name: "Bloodlust Aura", Description: "Allies gain attack speed", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"berserker_war_stomp"}, Bonuses: StatBonuses{Charisma: 6, Strength: 2}},
			{ID: "berserker_unbreakable_will", Name: "Unbreakable Will", Description: "Immune to crowd control", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"berserker_primal_instinct"}, Bonuses: StatBonuses{Wisdom: 5, Stamina: 25}},
			{ID: "berserker_avatar_of_war", Name: "Avatar of War", Description: "Embody pure combat", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"berserker_momentum", "berserker_bloodlust_aura"}, Bonuses: StatBonuses{Strength: 6, Speed: 0.2, Charisma: 5}},
		},
	}
}

// createPaladinTalentTree returns the talent tree for the Paladin class
func createPaladinTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Paladin Talents",
		ClassID: ClassPaladin,
		Offensive: []TalentDefinition{
			{ID: "paladin_holy_strike", Name: "Holy Strike", Description: "Weapon imbued with holy power", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 2, Wisdom: 2}},
			{ID: "paladin_consecration", Name: "Consecration", Description: "Holy ground damages enemies", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Intelligence: 1}},
			{ID: "paladin_divine_storm", Name: "Divine Storm", Description: "Area holy damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_holy_strike"}, Bonuses: StatBonuses{Strength: 3, Wisdom: 3}},
			{ID: "paladin_hammer_of_justice", Name: "Hammer of Justice", Description: "Stun with holy force", Category: CategoryOffensive, MaxRank: 3, Bonuses: StatBonuses{Strength: 4}},
			{ID: "paladin_righteous_fury", Name: "Righteous Fury", Description: "Holy wrath against evil", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_divine_storm"}, Bonuses: StatBonuses{Strength: 4, CritDamage: 0.1}},
			{ID: "paladin_seals", Name: "Seals of Power", Description: "Apply holy seals to attacks", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_consecration"}, Bonuses: StatBonuses{Wisdom: 4, Mana: 20}},
			{ID: "paladin_shield_bash", Name: "Shield Bash", Description: "Offensive shield technique", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"paladin_hammer_of_justice"}, Bonuses: StatBonuses{Strength: 5, Defense: 2}},
			{ID: "paladin_holy_wrath", Name: "Holy Wrath", Description: "Channel divine energy", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_righteous_fury"}, Bonuses: StatBonuses{Wisdom: 5, CritDamage: 0.15}},
			{ID: "paladin_execution_sentence", Name: "Execution Sentence", Description: "Delayed holy damage", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"paladin_seals"}, Bonuses: StatBonuses{Wisdom: 6, CritChance: 0.05}},
			{ID: "paladin_divine_champion", Name: "Divine Champion", Description: "Become holy warrior", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_holy_wrath", "paladin_execution_sentence"}, Bonuses: StatBonuses{Strength: 6, Wisdom: 6, CritDamage: 0.2}},
		},
		Defensive: []TalentDefinition{
			{ID: "paladin_divine_armor", Name: "Divine Armor", Description: "Holy-blessed protection", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 3, MagicDefense: 2}},
			{ID: "paladin_devotion_aura", Name: "Devotion Aura", Description: "Protection for all allies", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 2, MagicDefense: 2, Charisma: 2}},
			{ID: "paladin_blessing_of_protection", Name: "Blessing of Protection", Description: "Shield ally from harm", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"paladin_divine_armor"}, Bonuses: StatBonuses{Defense: 5, Health: 20}},
			{ID: "paladin_lay_on_hands", Name: "Lay on Hands", Description: "Emergency full heal", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"paladin_devotion_aura"}, Bonuses: StatBonuses{Health: 100, Wisdom: 5}},
			{ID: "paladin_sacred_shield", Name: "Sacred Shield", Description: "Damage absorption", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_blessing_of_protection"}, Bonuses: StatBonuses{Defense: 4, MagicDefense: 4}},
			{ID: "paladin_divine_intervention", Name: "Divine Intervention", Description: "Save ally from death", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"paladin_lay_on_hands"}, Bonuses: StatBonuses{Wisdom: 8, Charisma: 5}},
			{ID: "paladin_ardent_defender", Name: "Ardent Defender", Description: "Reduce fatal damage", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"paladin_sacred_shield"}, Bonuses: StatBonuses{Health: 60, Defense: 5}},
			{ID: "paladin_aegis_of_light", Name: "Aegis of Light", Description: "Party-wide shield", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_devotion_aura"}, Bonuses: StatBonuses{Defense: 3, MagicDefense: 3, Health: 30}},
			{ID: "paladin_divine_bulwark", Name: "Divine Bulwark", Description: "Become immovable", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_ardent_defender"}, Bonuses: StatBonuses{Defense: 8, Health: 50}},
			{ID: "paladin_avatar_of_protection", Name: "Avatar of Protection", Description: "Ultimate guardian", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"paladin_divine_bulwark", "paladin_divine_intervention"}, Bonuses: StatBonuses{Defense: 10, MagicDefense: 10, Health: 80}},
		},
		Utility: []TalentDefinition{
			{ID: "paladin_flash_of_light", Name: "Flash of Light", Description: "Quick healing spell", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Mana: 15}},
			{ID: "paladin_cleanse", Name: "Cleanse", Description: "Remove debuffs", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 2, Intelligence: 2}},
			{ID: "paladin_word_of_glory", Name: "Word of Glory", Description: "Holy power heal", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"paladin_flash_of_light"}, Bonuses: StatBonuses{Wisdom: 4, Mana: 20}},
			{ID: "paladin_blessing_of_might", Name: "Blessing of Might", Description: "Increase attack power", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Strength: 3, Charisma: 2}},
			{ID: "paladin_blessing_of_wisdom", Name: "Blessing of Wisdom", Description: "Increase mana regen", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"paladin_cleanse"}, Bonuses: StatBonuses{Wisdom: 4, Mana: 25}},
			{ID: "paladin_aura_mastery", Name: "Aura Mastery", Description: "Enhanced aura effects", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"paladin_blessing_of_might"}, Bonuses: StatBonuses{Charisma: 5, Wisdom: 3}},
			{ID: "paladin_divine_steed", Name: "Divine Steed", Description: "Summon holy mount", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Speed: 0.15, Stamina: 20}},
			{ID: "paladin_beacon_of_light", Name: "Beacon of Light", Description: "Healing duplicated to target", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"paladin_word_of_glory"}, Bonuses: StatBonuses{Wisdom: 5, Charisma: 4}},
			{ID: "paladin_holy_light", Name: "Holy Light", Description: "Powerful healing", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"paladin_beacon_of_light"}, Bonuses: StatBonuses{Wisdom: 6, Mana: 30}},
			{ID: "paladin_divine_purpose", Name: "Divine Purpose", Description: "Serve the light perfectly", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"paladin_holy_light", "paladin_aura_mastery"}, Bonuses: StatBonuses{Wisdom: 8, Charisma: 6, Mana: 40}},
		},
	}
}

// createKnightTalentTree returns the talent tree for the Knight class
func createKnightTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Knight Talents",
		ClassID: ClassKnight,
		Offensive: []TalentDefinition{
			{ID: "knight_lance_charge", Name: "Lance Charge", Description: "Devastating mounted charge", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 3, Speed: 0.05}},
			{ID: "knight_sword_mastery", Name: "Sword Mastery", Description: "Expert swordplay", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Strength: 4}},
			{ID: "knight_mortal_strike", Name: "Mortal Strike", Description: "Wound that won't heal", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"knight_sword_mastery"}, Bonuses: StatBonuses{Strength: 5, CritDamage: 0.1}},
			{ID: "knight_crushing_blow", Name: "Crushing Blow", Description: "Armor-piercing attack", Category: CategoryOffensive, MaxRank: 3, Bonuses: StatBonuses{Strength: 6}},
			{ID: "knight_blade_storm", Name: "Blade Storm", Description: "Whirlwind of steel", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"knight_mortal_strike"}, Bonuses: StatBonuses{Strength: 4, Dexterity: 3}},
			{ID: "knight_overpower", Name: "Overpower", Description: "Counter after dodge", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"knight_crushing_blow"}, Bonuses: StatBonuses{Strength: 5, CritChance: 0.04}},
			{ID: "knight_rend", Name: "Rend", Description: "Bleeding wounds", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"knight_lance_charge"}, Bonuses: StatBonuses{Strength: 3, Dexterity: 2}},
			{ID: "knight_colossus_smash", Name: "Colossus Smash", Description: "Shatter enemy defenses", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"knight_blade_storm"}, Bonuses: StatBonuses{Strength: 7, CritDamage: 0.2}},
			{ID: "knight_victory_rush", Name: "Victory Rush", Description: "Heal from kills", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"knight_overpower"}, Bonuses: StatBonuses{Strength: 4, Health: 30}},
			{ID: "knight_legendary_champion", Name: "Legendary Champion", Description: "Paragon of knighthood", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"knight_colossus_smash", "knight_victory_rush"}, Bonuses: StatBonuses{Strength: 8, CritDamage: 0.25, Health: 40}},
		},
		Defensive: []TalentDefinition{
			{ID: "knight_plate_armor", Name: "Plate Armor", Description: "Heavy armor expertise", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 4}},
			{ID: "knight_shield_wall", Name: "Shield Wall", Description: "Massive damage reduction", Category: CategoryDefensive, MaxRank: 3, Bonuses: StatBonuses{Defense: 6, Health: 20}},
			{ID: "knight_fortress", Name: "Fortress", Description: "Become immovable", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"knight_plate_armor"}, Bonuses: StatBonuses{Defense: 5, Health: 30}},
			{ID: "knight_spell_reflection", Name: "Spell Reflection", Description: "Return hostile spells", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"knight_shield_wall"}, Bonuses: StatBonuses{MagicDefense: 6}},
			{ID: "knight_shield_block", Name: "Shield Block", Description: "Block incoming attacks", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"knight_fortress"}, Bonuses: StatBonuses{Defense: 6, Dexterity: 2}},
			{ID: "knight_last_stand", Name: "Last Stand", Description: "Emergency health boost", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"knight_spell_reflection"}, Bonuses: StatBonuses{Health: 100}},
			{ID: "knight_impenetrable_defense", Name: "Impenetrable Defense", Description: "Reduce all damage", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"knight_shield_block"}, Bonuses: StatBonuses{Defense: 7, MagicDefense: 4}},
			{ID: "knight_vigilance", Name: "Vigilance", Description: "Protect an ally", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 3, Charisma: 3}},
			{ID: "knight_die_by_the_sword", Name: "Die by the Sword", Description: "Parry all attacks briefly", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"knight_impenetrable_defense"}, Bonuses: StatBonuses{Defense: 8, Dexterity: 4}},
			{ID: "knight_immortal_bastion", Name: "Immortal Bastion", Description: "Ultimate defensive stance", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"knight_last_stand", "knight_die_by_the_sword"}, Bonuses: StatBonuses{Defense: 12, MagicDefense: 8, Health: 80}},
		},
		Utility: []TalentDefinition{
			{ID: "knight_commanding_shout", Name: "Commanding Shout", Description: "Boost ally health", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Charisma: 4, Health: 20}},
			{ID: "knight_banner", Name: "Banner of Valor", Description: "Inspire allies", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Charisma: 5, Strength: 2}},
			{ID: "knight_intervene", Name: "Intervene", Description: "Intercept attacks for ally", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"knight_commanding_shout"}, Bonuses: StatBonuses{Speed: 0.1, Defense: 3}},
			{ID: "knight_heroic_leap", Name: "Heroic Leap", Description: "Leap to location", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Speed: 0.15, Stamina: 15}},
			{ID: "knight_rallying_cry", Name: "Rallying Cry", Description: "Mass ally buff", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_banner"}, Bonuses: StatBonuses{Charisma: 6, Health: 30}},
			{ID: "knight_taunt", Name: "Taunt", Description: "Force enemy attention", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_intervene"}, Bonuses: StatBonuses{Charisma: 4, Defense: 2}},
			{ID: "knight_mounted_combat", Name: "Mounted Combat", Description: "Fight from horseback", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_heroic_leap"}, Bonuses: StatBonuses{Speed: 0.2, Stamina: 25}},
			{ID: "knight_leadership", Name: "Leadership", Description: "Natural commander", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_rallying_cry"}, Bonuses: StatBonuses{Charisma: 7, Wisdom: 3}},
			{ID: "knight_inspiring_presence", Name: "Inspiring Presence", Description: "Allies fight harder", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_taunt"}, Bonuses: StatBonuses{Charisma: 5, Strength: 3}},
			{ID: "knight_legendary_commander", Name: "Legendary Commander", Description: "Ultimate leadership", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"knight_leadership", "knight_inspiring_presence"}, Bonuses: StatBonuses{Charisma: 10, Strength: 5, Wisdom: 5}},
		},
	}
}

// createAssassinTalentTree returns the talent tree for the Assassin class
func createAssassinTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Assassin Talents",
		ClassID: ClassAssassin,
		Offensive: []TalentDefinition{
			{ID: "assassin_assassination", Name: "Assassination", Description: "Deadly kill techniques", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, CritDamage: 0.15}},
			{ID: "assassin_venomous_blades", Name: "Venomous Blades", Description: "Poison-coated weapons", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 2, Intelligence: 2}},
			{ID: "assassin_garrote", Name: "Garrote", Description: "Silent takedown", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_assassination"}, Bonuses: StatBonuses{Dexterity: 4, CritChance: 0.05}},
			{ID: "assassin_mutilate", Name: "Mutilate", Description: "Dual-wield strikes", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 4, CritDamage: 0.1}},
			{ID: "assassin_envenom", Name: "Envenom", Description: "Concentrated poison burst", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_venomous_blades"}, Bonuses: StatBonuses{Intelligence: 4, CritDamage: 0.15}},
			{ID: "assassin_vendetta_mark", Name: "Vendetta", Description: "Mark for death", Category: CategoryOffensive, MaxRank: 1, Prerequisites: []TalentID{"assassin_garrote"}, Bonuses: StatBonuses{CritChance: 0.1, CritDamage: 0.3}},
			{ID: "assassin_rupture", Name: "Rupture", Description: "Severe bleeding", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_mutilate"}, Bonuses: StatBonuses{Dexterity: 3, Strength: 2}},
			{ID: "assassin_deadly_momentum", Name: "Deadly Momentum", Description: "Chain kills refresh", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"assassin_envenom"}, Bonuses: StatBonuses{Dexterity: 5, Stamina: 20}},
			{ID: "assassin_death_from_above", Name: "Death from Above", Description: "Aerial assassination", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"assassin_rupture"}, Bonuses: StatBonuses{Dexterity: 6, CritDamage: 0.25}},
			{ID: "assassin_master_killer", Name: "Master Killer", Description: "Perfect assassination", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_vendetta_mark", "assassin_death_from_above"}, Bonuses: StatBonuses{Dexterity: 8, CritChance: 0.12, CritDamage: 0.4}},
		},
		Defensive: []TalentDefinition{
			{ID: "assassin_fleet_footed", Name: "Fleet Footed", Description: "Enhanced movement", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Speed: 0.1, Dexterity: 2}},
			{ID: "assassin_elusiveness", Name: "Elusiveness", Description: "Hard to hit", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Defense: 2}},
			{ID: "assassin_shadow_dance", Name: "Shadow Dance", Description: "Move unseen", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"assassin_fleet_footed"}, Bonuses: StatBonuses{Speed: 0.15, Dexterity: 4}},
			{ID: "assassin_cloak_and_dagger", Name: "Cloak and Dagger", Description: "Teleport from stealth", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_elusiveness"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.1}},
			{ID: "assassin_leeching_poison", Name: "Leeching Poison", Description: "Heal from poison damage", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_shadow_dance"}, Bonuses: StatBonuses{Health: 30, Intelligence: 2}},
			{ID: "assassin_evasive_maneuvers", Name: "Evasive Maneuvers", Description: "Dodge area attacks", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"assassin_cloak_and_dagger"}, Bonuses: StatBonuses{Dexterity: 6}},
			{ID: "assassin_feint_death", Name: "Feign Death", Description: "Play dead to escape", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"assassin_leeching_poison"}, Bonuses: StatBonuses{Health: 50}},
			{ID: "assassin_uncanny_dodge", Name: "Uncanny Dodge", Description: "Avoid critical hits", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_evasive_maneuvers"}, Bonuses: StatBonuses{Dexterity: 5, Defense: 4}},
			{ID: "assassin_blindside", Name: "Blindside", Description: "Attack from blind spots", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, CritChance: 0.03}},
			{ID: "assassin_phantom_strike", Name: "Phantom Strike", Description: "Untouchable killer", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"assassin_feint_death", "assassin_uncanny_dodge"}, Bonuses: StatBonuses{Dexterity: 8, Speed: 0.2, Defense: 6}},
		},
		Utility: []TalentDefinition{
			{ID: "assassin_marked_for_death", Name: "Marked for Death", Description: "Generate combo points on target", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 2, Stamina: 15}},
			{ID: "assassin_silent_footsteps", Name: "Silent Footsteps", Description: "Move without sound", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Speed: 0.05}},
			{ID: "assassin_anticipation", Name: "Anticipation", Description: "Store extra combo points", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"assassin_marked_for_death"}, Bonuses: StatBonuses{Dexterity: 4, Wisdom: 2}},
			{ID: "assassin_quick_getaway", Name: "Quick Getaway", Description: "Escape after kills", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"assassin_silent_footsteps"}, Bonuses: StatBonuses{Speed: 0.15, Stamina: 20}},
			{ID: "assassin_master_poisoner", Name: "Master Poisoner", Description: "Enhanced poison effects", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 4, Dexterity: 2}},
			{ID: "assassin_cut_to_the_chase", Name: "Cut to the Chase", Description: "Refresh buffs on crit", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"assassin_anticipation"}, Bonuses: StatBonuses{CritChance: 0.05, Stamina: 15}},
			{ID: "assassin_internal_bleeding", Name: "Internal Bleeding", Description: "Bleeds from kidney shot", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"assassin_master_poisoner"}, Bonuses: StatBonuses{Dexterity: 4, Strength: 2}},
			{ID: "assassin_thuggery", Name: "Thuggery", Description: "Longer stuns", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"assassin_quick_getaway"}, Bonuses: StatBonuses{Dexterity: 5, Charisma: 2}},
			{ID: "assassin_elaborate_planning", Name: "Elaborate Planning", Description: "Bonus from finishers", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"assassin_cut_to_the_chase"}, Bonuses: StatBonuses{Dexterity: 5, Wisdom: 4}},
			{ID: "assassin_deathstalker", Name: "Deathstalker", Description: "Perfect predator", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"assassin_elaborate_planning", "assassin_thuggery"}, Bonuses: StatBonuses{Dexterity: 8, Speed: 0.2, Intelligence: 5}},
		},
	}
}

// createRangerTalentTree returns the talent tree for the Ranger class
func createRangerTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Ranger Talents",
		ClassID: ClassRanger,
		Offensive: []TalentDefinition{
			{ID: "ranger_aimed_shot", Name: "Aimed Shot", Description: "Precise long-range shot", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, CritChance: 0.03}},
			{ID: "ranger_multi_shot", Name: "Multi-Shot", Description: "Hit multiple targets", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Stamina: 10}},
			{ID: "ranger_steady_shot", Name: "Steady Shot", Description: "Reliable damage dealer", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_aimed_shot"}, Bonuses: StatBonuses{Dexterity: 4, Wisdom: 2}},
			{ID: "ranger_explosive_shot", Name: "Explosive Shot", Description: "Area damage arrows", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_multi_shot"}, Bonuses: StatBonuses{Dexterity: 4, Intelligence: 2}},
			{ID: "ranger_kill_shot", Name: "Kill Shot", Description: "Execute low health targets", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"ranger_steady_shot"}, Bonuses: StatBonuses{CritDamage: 0.25, Dexterity: 3}},
			{ID: "ranger_barrage", Name: "Barrage", Description: "Rapid fire arrows", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_explosive_shot"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.05}},
			{ID: "ranger_black_arrow", Name: "Black Arrow", Description: "Shadow-infused shot", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"ranger_aimed_shot"}, Bonuses: StatBonuses{Dexterity: 4, Intelligence: 3}},
			{ID: "ranger_chimera_shot", Name: "Chimera Shot", Description: "Healing and damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_barrage"}, Bonuses: StatBonuses{Dexterity: 5, Health: 20}},
			{ID: "ranger_powershot", Name: "Powershot", Description: "Penetrating shot", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"ranger_kill_shot"}, Bonuses: StatBonuses{Dexterity: 6, CritDamage: 0.2}},
			{ID: "ranger_sniper", Name: "Sniper", Description: "Ultimate marksman", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_powershot", "ranger_chimera_shot"}, Bonuses: StatBonuses{Dexterity: 8, CritChance: 0.1, CritDamage: 0.3}},
		},
		Defensive: []TalentDefinition{
			{ID: "ranger_aspect_of_the_hawk", Name: "Aspect of the Hawk", Description: "Enhanced ranged attacks", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Speed: 0.05}},
			{ID: "ranger_disengage", Name: "Disengage", Description: "Leap away from danger", Category: CategoryDefensive, MaxRank: 3, Bonuses: StatBonuses{Speed: 0.1, Stamina: 15}},
			{ID: "ranger_aspect_of_the_turtle", Name: "Aspect of the Turtle", Description: "Damage immunity", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"ranger_aspect_of_the_hawk"}, Bonuses: StatBonuses{Defense: 10}},
			{ID: "ranger_exhilaration", Name: "Exhilaration", Description: "Heal self and pet", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_disengage"}, Bonuses: StatBonuses{Health: 30, Stamina: 20}},
			{ID: "ranger_camouflage", Name: "Camouflage", Description: "Blend with environment", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_aspect_of_the_turtle"}, Bonuses: StatBonuses{Dexterity: 4, Speed: 0.1}},
			{ID: "ranger_misdirection", Name: "Misdirection", Description: "Transfer threat", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"ranger_exhilaration"}, Bonuses: StatBonuses{Charisma: 3, Dexterity: 3}},
			{ID: "ranger_survival_instincts", Name: "Survival Instincts", Description: "Emergency protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_camouflage"}, Bonuses: StatBonuses{Health: 40, Defense: 4}},
			{ID: "ranger_feign_death", Name: "Feign Death", Description: "Play dead", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"ranger_misdirection"}, Bonuses: StatBonuses{Wisdom: 5}},
			{ID: "ranger_deterrence", Name: "Deterrence", Description: "Deflect attacks", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_survival_instincts"}, Bonuses: StatBonuses{Defense: 6, MagicDefense: 4}},
			{ID: "ranger_survivalist", Name: "Survivalist", Description: "Ultimate wilderness survival", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ranger_deterrence", "ranger_feign_death"}, Bonuses: StatBonuses{Health: 60, Defense: 8, Speed: 0.15}},
		},
		Utility: []TalentDefinition{
			{ID: "ranger_hunters_mark", Name: "Hunter's Mark", Description: "Mark prey for tracking", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Dexterity: 2}},
			{ID: "ranger_tracking", Name: "Tracking", Description: "Follow creature trails", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 4}},
			{ID: "ranger_pet_mastery", Name: "Pet Mastery", Description: "Enhanced animal companion", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_hunters_mark"}, Bonuses: StatBonuses{Charisma: 3, Wisdom: 3}},
			{ID: "ranger_trap_mastery", Name: "Trap Mastery", Description: "Improved traps", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_tracking"}, Bonuses: StatBonuses{Intelligence: 3, Dexterity: 3}},
			{ID: "ranger_beast_lore", Name: "Beast Lore", Description: "Understand animals", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_pet_mastery"}, Bonuses: StatBonuses{Wisdom: 5, Charisma: 3}},
			{ID: "ranger_eagle_eye", Name: "Eagle Eye", Description: "Far seeing", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Wisdom: 5}},
			{ID: "ranger_flare", Name: "Flare", Description: "Reveal hidden enemies", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"ranger_trap_mastery"}, Bonuses: StatBonuses{Wisdom: 4, Intelligence: 2}},
			{ID: "ranger_wild_call", Name: "Wild Call", Description: "Summon beast allies", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_beast_lore"}, Bonuses: StatBonuses{Charisma: 5, Wisdom: 4}},
			{ID: "ranger_coordinated_assault", Name: "Coordinated Assault", Description: "Pet synergy attacks", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_flare"}, Bonuses: StatBonuses{Dexterity: 4, Charisma: 4}},
			{ID: "ranger_master_hunter", Name: "Master Hunter", Description: "Supreme tracker and hunter", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ranger_wild_call", "ranger_coordinated_assault"}, Bonuses: StatBonuses{Wisdom: 8, Dexterity: 5, Charisma: 5}},
		},
	}
}

// createNinjaTalentTree returns the talent tree for the Ninja class
func createNinjaTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Ninja Talents",
		ClassID: ClassNinja,
		Offensive: []TalentDefinition{
			{ID: "ninja_shuriken", Name: "Shuriken Mastery", Description: "Deadly throwing stars", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Speed: 0.05}},
			{ID: "ninja_ninjutsu", Name: "Ninjutsu", Description: "Secret ninja arts", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 2, Intelligence: 2}},
			{ID: "ninja_shadow_strike", Name: "Shadow Strike", Description: "Attack from shadows", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_shuriken"}, Bonuses: StatBonuses{Dexterity: 4, CritDamage: 0.15}},
			{ID: "ninja_poison_bomb", Name: "Poison Bomb", Description: "Area poison effect", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_ninjutsu"}, Bonuses: StatBonuses{Intelligence: 4, Dexterity: 2}},
			{ID: "ninja_death_blossom", Name: "Death Blossom", Description: "Spinning blade attack", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"ninja_shadow_strike"}, Bonuses: StatBonuses{Dexterity: 5, CritChance: 0.05}},
			{ID: "ninja_trick_attack", Name: "Trick Attack", Description: "Bonus damage from flanking", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_poison_bomb"}, Bonuses: StatBonuses{Dexterity: 5, CritDamage: 0.1}},
			{ID: "ninja_blade_dance", Name: "Blade Dance", Description: "Fluid weapon combos", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_death_blossom"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.1}},
			{ID: "ninja_assassinate", Name: "Assassinate", Description: "Instant kill attempt", Category: CategoryOffensive, MaxRank: 1, Prerequisites: []TalentID{"ninja_trick_attack"}, Bonuses: StatBonuses{CritDamage: 0.5, CritChance: 0.1}},
			{ID: "ninja_shadow_clone_jutsu", Name: "Shadow Clone Jutsu", Description: "Create attacking clones", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"ninja_blade_dance"}, Bonuses: StatBonuses{Dexterity: 6, Intelligence: 4}},
			{ID: "ninja_master_shinobi", Name: "Master Shinobi", Description: "Ultimate ninja", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_assassinate", "ninja_shadow_clone_jutsu"}, Bonuses: StatBonuses{Dexterity: 10, CritDamage: 0.35, Speed: 0.15}},
		},
		Defensive: []TalentDefinition{
			{ID: "ninja_smoke_bomb", Name: "Smoke Bomb", Description: "Obscure vision", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 2, Defense: 2}},
			{ID: "ninja_acrobatics", Name: "Acrobatics", Description: "Enhanced mobility", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Speed: 0.1}},
			{ID: "ninja_shade_shift", Name: "Shade Shift", Description: "Dodge through enemies", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"ninja_smoke_bomb"}, Bonuses: StatBonuses{Dexterity: 4, Speed: 0.1}},
			{ID: "ninja_substitution", Name: "Substitution Jutsu", Description: "Replace self with log", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"ninja_acrobatics"}, Bonuses: StatBonuses{Dexterity: 5}},
			{ID: "ninja_third_eye", Name: "Third Eye", Description: "Detect hidden threats", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_shade_shift"}, Bonuses: StatBonuses{Wisdom: 4, Dexterity: 2}},
			{ID: "ninja_blur", Name: "Blur", Description: "Hard to see clearly", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_substitution"}, Bonuses: StatBonuses{Dexterity: 6, Defense: 3}},
			{ID: "ninja_shadow_meld", Name: "Shadow Meld", Description: "Become one with shadows", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"ninja_third_eye"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.15}},
			{ID: "ninja_hollow_body", Name: "Hollow Body", Description: "Attacks pass through", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"ninja_blur"}, Bonuses: StatBonuses{MagicDefense: 10}},
			{ID: "ninja_deflect", Name: "Deflect", Description: "Redirect projectiles", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_shadow_meld"}, Bonuses: StatBonuses{Dexterity: 6, Defense: 5}},
			{ID: "ninja_untouchable", Name: "Untouchable", Description: "Perfect evasion", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"ninja_hollow_body", "ninja_deflect"}, Bonuses: StatBonuses{Dexterity: 10, Speed: 0.25, Defense: 8}},
		},
		Utility: []TalentDefinition{
			{ID: "ninja_silent_movement", Name: "Silent Movement", Description: "Move without sound", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Speed: 0.05}},
			{ID: "ninja_wall_run", Name: "Wall Run", Description: "Run on walls", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Dexterity: 4, Speed: 0.1}},
			{ID: "ninja_hide_in_shadows", Name: "Hide in Shadows", Description: "Vanish in darkness", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ninja_silent_movement"}, Bonuses: StatBonuses{Dexterity: 4, Wisdom: 2}},
			{ID: "ninja_shunshin", Name: "Shunshin", Description: "Body flicker technique", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"ninja_wall_run"}, Bonuses: StatBonuses{Speed: 0.2, Stamina: 20}},
			{ID: "ninja_trap_detection", Name: "Trap Detection", Description: "Sense hidden dangers", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 4, Dexterity: 2}},
			{ID: "ninja_chakra_control", Name: "Chakra Control", Description: "Efficient energy use", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ninja_hide_in_shadows"}, Bonuses: StatBonuses{Mana: 25, Wisdom: 3}},
			{ID: "ninja_genjutsu", Name: "Genjutsu", Description: "Illusion techniques", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ninja_shunshin"}, Bonuses: StatBonuses{Intelligence: 5, Charisma: 3}},
			{ID: "ninja_shadow_walking", Name: "Shadow Walking", Description: "Travel through shadows", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"ninja_chakra_control"}, Bonuses: StatBonuses{Speed: 0.15, Dexterity: 4}},
			{ID: "ninja_kage_bunshin", Name: "Shadow Clone", Description: "Create helper clones", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ninja_genjutsu"}, Bonuses: StatBonuses{Intelligence: 4, Mana: 30}},
			{ID: "ninja_grandmaster_shinobi", Name: "Grandmaster Shinobi", Description: "Legendary ninja arts", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"ninja_shadow_walking", "ninja_kage_bunshin"}, Bonuses: StatBonuses{Dexterity: 8, Intelligence: 6, Speed: 0.2}},
		},
	}
}

// createElementalistTalentTree returns the talent tree for the Elementalist class
func createElementalistTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Elementalist Talents",
		ClassID: ClassElementalist,
		Offensive: []TalentDefinition{
			{ID: "elementalist_fire_mastery", Name: "Fire Mastery", Description: "Control flames", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, CritDamage: 0.1}},
			{ID: "elementalist_frost_mastery", Name: "Frost Mastery", Description: "Command ice", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, Wisdom: 1}},
			{ID: "elementalist_pyroblast", Name: "Pyroblast", Description: "Massive fire damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_fire_mastery"}, Bonuses: StatBonuses{Intelligence: 5, CritDamage: 0.15}},
			{ID: "elementalist_ice_lance", Name: "Ice Lance", Description: "Frozen shatter", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_frost_mastery"}, Bonuses: StatBonuses{Intelligence: 4, CritChance: 0.04}},
			{ID: "elementalist_lightning", Name: "Lightning Bolt", Description: "Storm power", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 4, Dexterity: 1}},
			{ID: "elementalist_combustion", Name: "Combustion", Description: "Intense burning", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"elementalist_pyroblast"}, Bonuses: StatBonuses{Intelligence: 6, CritChance: 0.05}},
			{ID: "elementalist_frozen_orb", Name: "Frozen Orb", Description: "Traveling ice sphere", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_ice_lance"}, Bonuses: StatBonuses{Intelligence: 5, Mana: 20}},
			{ID: "elementalist_chain_lightning", Name: "Chain Lightning", Description: "Bouncing storm", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_lightning"}, Bonuses: StatBonuses{Intelligence: 5, Wisdom: 2}},
			{ID: "elementalist_elemental_fury", Name: "Elemental Fury", Description: "Combined element power", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"elementalist_combustion", "elementalist_frozen_orb"}, Bonuses: StatBonuses{Intelligence: 7, CritDamage: 0.2}},
			{ID: "elementalist_primordial_power", Name: "Primordial Power", Description: "Master all elements", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_elemental_fury", "elementalist_chain_lightning"}, Bonuses: StatBonuses{Intelligence: 10, CritDamage: 0.3, Mana: 40}},
		},
		Defensive: []TalentDefinition{
			{ID: "elementalist_elemental_shield", Name: "Elemental Shield", Description: "Protection from elements", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{MagicDefense: 4}},
			{ID: "elementalist_blazing_barrier", Name: "Blazing Barrier", Description: "Fire protection aura", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{MagicDefense: 3, Defense: 2}},
			{ID: "elementalist_ice_barrier", Name: "Ice Barrier", Description: "Frozen protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_elemental_shield"}, Bonuses: StatBonuses{MagicDefense: 5, Health: 20}},
			{ID: "elementalist_fire_ward", Name: "Fire Ward", Description: "Absorb fire damage", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"elementalist_blazing_barrier"}, Bonuses: StatBonuses{MagicDefense: 6}},
			{ID: "elementalist_frost_ward", Name: "Frost Ward", Description: "Absorb frost damage", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"elementalist_ice_barrier"}, Bonuses: StatBonuses{MagicDefense: 6}},
			{ID: "elementalist_static_shield", Name: "Static Shield", Description: "Lightning protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_fire_ward"}, Bonuses: StatBonuses{MagicDefense: 5, Dexterity: 3}},
			{ID: "elementalist_elemental_absorption", Name: "Elemental Absorption", Description: "Convert damage to mana", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_frost_ward"}, Bonuses: StatBonuses{Mana: 35, MagicDefense: 4}},
			{ID: "elementalist_cauterize", Name: "Cauterize", Description: "Survive fatal damage", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"elementalist_static_shield"}, Bonuses: StatBonuses{Health: 60}},
			{ID: "elementalist_greater_barrier", Name: "Greater Barrier", Description: "Enhanced protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_elemental_absorption"}, Bonuses: StatBonuses{MagicDefense: 8, Health: 40}},
			{ID: "elementalist_elemental_immunity", Name: "Elemental Immunity", Description: "Immune to elements", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"elementalist_cauterize", "elementalist_greater_barrier"}, Bonuses: StatBonuses{MagicDefense: 12, Defense: 6, Health: 50}},
		},
		Utility: []TalentDefinition{
			{ID: "elementalist_elemental_attunement", Name: "Elemental Attunement", Description: "Faster element switching", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 2, Wisdom: 2}},
			{ID: "elementalist_mana_flow", Name: "Mana Flow", Description: "Better mana regen", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Mana: 25, Wisdom: 2}},
			{ID: "elementalist_elemental_precision", Name: "Elemental Precision", Description: "Increased spell hit", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"elementalist_elemental_attunement"}, Bonuses: StatBonuses{Intelligence: 4, Wisdom: 3}},
			{ID: "elementalist_conjure_elemental", Name: "Conjure Elemental", Description: "Summon element being", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"elementalist_mana_flow"}, Bonuses: StatBonuses{Intelligence: 4, Charisma: 3}},
			{ID: "elementalist_travel_form", Name: "Elemental Travel", Description: "Transform for speed", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Speed: 0.15, Mana: 20}},
			{ID: "elementalist_presence_of_mind", Name: "Presence of Mind", Description: "Instant cast", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"elementalist_elemental_precision"}, Bonuses: StatBonuses{Intelligence: 5, Mana: 25}},
			{ID: "elementalist_elemental_binding", Name: "Elemental Binding", Description: "Control summoned elementals", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"elementalist_conjure_elemental"}, Bonuses: StatBonuses{Charisma: 5, Intelligence: 3}},
			{ID: "elementalist_arcane_power", Name: "Arcane Power", Description: "Burst damage phase", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"elementalist_presence_of_mind"}, Bonuses: StatBonuses{Intelligence: 6, CritDamage: 0.15}},
			{ID: "elementalist_elemental_overload", Name: "Elemental Overload", Description: "Echo elemental spells", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"elementalist_elemental_binding"}, Bonuses: StatBonuses{Intelligence: 5, Mana: 35}},
			{ID: "elementalist_avatar_of_elements", Name: "Avatar of Elements", Description: "Become elemental being", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"elementalist_arcane_power", "elementalist_elemental_overload"}, Bonuses: StatBonuses{Intelligence: 10, Mana: 50, Wisdom: 6}},
		},
	}
}

// createNecromancerTalentTree returns the talent tree for the Necromancer class
func createNecromancerTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Necromancer Talents",
		ClassID: ClassNecromancer,
		Offensive: []TalentDefinition{
			{ID: "necromancer_shadow_bolt", Name: "Shadow Bolt", Description: "Dark magic missile", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, CritChance: 0.02}},
			{ID: "necromancer_drain_life", Name: "Drain Life", Description: "Steal health from enemies", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 2, Health: 15}},
			{ID: "necromancer_corruption", Name: "Corruption", Description: "Damage over time", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_shadow_bolt"}, Bonuses: StatBonuses{Intelligence: 4}},
			{ID: "necromancer_death_coil", Name: "Death Coil", Description: "Heal from damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_drain_life"}, Bonuses: StatBonuses{Intelligence: 4, Health: 20}},
			{ID: "necromancer_haunt", Name: "Haunt", Description: "Spirit attack", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"necromancer_corruption"}, Bonuses: StatBonuses{Intelligence: 5, CritDamage: 0.15}},
			{ID: "necromancer_unstable_affliction", Name: "Unstable Affliction", Description: "Exploding curse", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_death_coil"}, Bonuses: StatBonuses{Intelligence: 5, CritChance: 0.04}},
			{ID: "necromancer_soul_fire", Name: "Soul Fire", Description: "Burn enemy soul", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_haunt"}, Bonuses: StatBonuses{Intelligence: 6, CritDamage: 0.2}},
			{ID: "necromancer_doom", Name: "Doom", Description: "Delayed massive damage", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"necromancer_unstable_affliction"}, Bonuses: StatBonuses{Intelligence: 7}},
			{ID: "necromancer_soul_swap", Name: "Soul Swap", Description: "Transfer DoTs instantly", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"necromancer_soul_fire"}, Bonuses: StatBonuses{Intelligence: 5, Mana: 20}},
			{ID: "necromancer_lord_of_death", Name: "Lord of Death", Description: "Master of dark arts", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_doom", "necromancer_soul_swap"}, Bonuses: StatBonuses{Intelligence: 10, CritDamage: 0.35, Health: 40}},
		},
		Defensive: []TalentDefinition{
			{ID: "necromancer_soul_link", Name: "Soul Link", Description: "Share damage with minion", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Health: 25}},
			{ID: "necromancer_dark_pact", Name: "Dark Pact", Description: "Sacrifice health for mana", Category: CategoryDefensive, MaxRank: 3, Bonuses: StatBonuses{Mana: 30}},
			{ID: "necromancer_fel_armor", Name: "Fel Armor", Description: "Dark protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_soul_link"}, Bonuses: StatBonuses{Defense: 3, MagicDefense: 4}},
			{ID: "necromancer_shadow_ward", Name: "Shadow Ward", Description: "Absorb shadow damage", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_dark_pact"}, Bonuses: StatBonuses{MagicDefense: 6}},
			{ID: "necromancer_unending_resolve", Name: "Unending Resolve", Description: "Damage reduction", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"necromancer_fel_armor"}, Bonuses: StatBonuses{Defense: 5, Health: 30}},
			{ID: "necromancer_dark_bargain", Name: "Dark Bargain", Description: "Delay damage taken", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"necromancer_shadow_ward"}, Bonuses: StatBonuses{Health: 50}},
			{ID: "necromancer_soulstone", Name: "Soulstone", Description: "Self-resurrection", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"necromancer_unending_resolve"}, Bonuses: StatBonuses{Health: 75}},
			{ID: "necromancer_sacrificial_pact", Name: "Sacrificial Pact", Description: "Kill minion for shield", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_dark_bargain"}, Bonuses: StatBonuses{Health: 40, Defense: 4}},
			{ID: "necromancer_grim_harvest", Name: "Grim Harvest", Description: "Heal from deaths", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_soulstone"}, Bonuses: StatBonuses{Health: 50, Intelligence: 3}},
			{ID: "necromancer_immortal_darkness", Name: "Immortal Darkness", Description: "Cannot truly die", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"necromancer_sacrificial_pact", "necromancer_grim_harvest"}, Bonuses: StatBonuses{Health: 100, MagicDefense: 10, Defense: 6}},
		},
		Utility: []TalentDefinition{
			{ID: "necromancer_summon_undead", Name: "Summon Undead", Description: "Raise skeletal minion", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 2, Charisma: 2}},
			{ID: "necromancer_life_tap", Name: "Life Tap", Description: "Convert health to mana", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Mana: 25}},
			{ID: "necromancer_create_healthstone", Name: "Create Healthstone", Description: "Craft healing stone", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"necromancer_summon_undead"}, Bonuses: StatBonuses{Intelligence: 3, Health: 20}},
			{ID: "necromancer_demonic_gateway", Name: "Demonic Gateway", Description: "Create teleport point", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"necromancer_life_tap"}, Bonuses: StatBonuses{Mana: 30, Speed: 0.1}},
			{ID: "necromancer_command_undead", Name: "Command Undead", Description: "Control more minions", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"necromancer_create_healthstone"}, Bonuses: StatBonuses{Charisma: 5, Intelligence: 3}},
			{ID: "necromancer_dark_intent", Name: "Dark Intent", Description: "Buff ally damage", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"necromancer_demonic_gateway"}, Bonuses: StatBonuses{Intelligence: 4, Charisma: 3}},
			{ID: "necromancer_speak_with_dead", Name: "Speak with Dead", Description: "Gain knowledge from corpses", Category: CategoryUtility, MaxRank: 3, Bonuses: StatBonuses{Wisdom: 5, Intelligence: 2}},
			{ID: "necromancer_army_of_the_dead", Name: "Army of the Dead", Description: "Raise skeleton army", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"necromancer_command_undead"}, Bonuses: StatBonuses{Charisma: 6, Intelligence: 4}},
			{ID: "necromancer_soul_harvest", Name: "Soul Harvest", Description: "Collect souls for power", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"necromancer_dark_intent"}, Bonuses: StatBonuses{Intelligence: 6, Mana: 35}},
			{ID: "necromancer_death_lord", Name: "Death Lord", Description: "Master of undeath", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"necromancer_army_of_the_dead", "necromancer_soul_harvest"}, Bonuses: StatBonuses{Intelligence: 8, Charisma: 8, Mana: 50}},
		},
	}
}

// createEnchanterTalentTree returns the talent tree for the Enchanter class
func createEnchanterTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Enchanter Talents",
		ClassID: ClassEnchanter,
		Offensive: []TalentDefinition{
			{ID: "enchanter_arcane_blast", Name: "Arcane Blast", Description: "Pure arcane damage", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, Mana: 10}},
			{ID: "enchanter_mind_spike", Name: "Mind Spike", Description: "Mental assault", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, Wisdom: 1}},
			{ID: "enchanter_arcane_barrage", Name: "Arcane Barrage", Description: "Rapid arcane missiles", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_arcane_blast"}, Bonuses: StatBonuses{Intelligence: 4, CritChance: 0.03}},
			{ID: "enchanter_psychic_scream", Name: "Psychic Scream", Description: "Fear enemies", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"enchanter_mind_spike"}, Bonuses: StatBonuses{Intelligence: 4, Charisma: 3}},
			{ID: "enchanter_nether_tempest", Name: "Nether Tempest", Description: "Bouncing damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_arcane_barrage"}, Bonuses: StatBonuses{Intelligence: 5, CritDamage: 0.1}},
			{ID: "enchanter_dominate_mind", Name: "Dominate Mind", Description: "Control enemy", Category: CategoryOffensive, MaxRank: 1, Prerequisites: []TalentID{"enchanter_psychic_scream"}, Bonuses: StatBonuses{Intelligence: 6, Charisma: 5}},
			{ID: "enchanter_supernova", Name: "Supernova", Description: "Arcane explosion", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_nether_tempest"}, Bonuses: StatBonuses{Intelligence: 6, CritDamage: 0.15}},
			{ID: "enchanter_mind_control", Name: "Mind Control", Description: "Extended domination", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_dominate_mind"}, Bonuses: StatBonuses{Intelligence: 5, Charisma: 6}},
			{ID: "enchanter_arcane_orb", Name: "Arcane Orb", Description: "Traveling damage sphere", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_supernova"}, Bonuses: StatBonuses{Intelligence: 7, Mana: 25}},
			{ID: "enchanter_master_illusionist", Name: "Master Illusionist", Description: "Reality bending power", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_mind_control", "enchanter_arcane_orb"}, Bonuses: StatBonuses{Intelligence: 10, Charisma: 8, CritDamage: 0.25}},
		},
		Defensive: []TalentDefinition{
			{ID: "enchanter_mana_shield", Name: "Mana Shield", Description: "Absorb with mana", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Mana: 25, MagicDefense: 2}},
			{ID: "enchanter_alter_time", Name: "Alter Time", Description: "Revert to previous state", Category: CategoryDefensive, MaxRank: 1, Bonuses: StatBonuses{Intelligence: 5}},
			{ID: "enchanter_prismatic_barrier", Name: "Prismatic Barrier", Description: "Multi-element shield", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_mana_shield"}, Bonuses: StatBonuses{MagicDefense: 5, Defense: 3}},
			{ID: "enchanter_slow_fall", Name: "Slow Fall", Description: "Reduce fall damage", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"enchanter_alter_time"}, Bonuses: StatBonuses{Dexterity: 3}},
			{ID: "enchanter_greater_invisibility", Name: "Greater Invisibility", Description: "Powerful stealth", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"enchanter_prismatic_barrier"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.1}},
			{ID: "enchanter_temporal_shield", Name: "Temporal Shield", Description: "Time-delayed healing", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_slow_fall"}, Bonuses: StatBonuses{Health: 40, Mana: 20}},
			{ID: "enchanter_ring_of_frost", Name: "Ring of Frost", Description: "Freeze attackers", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"enchanter_greater_invisibility"}, Bonuses: StatBonuses{MagicDefense: 6}},
			{ID: "enchanter_evanesce", Name: "Evanesce", Description: "Become untargetable", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"enchanter_temporal_shield"}, Bonuses: StatBonuses{Speed: 0.2}},
			{ID: "enchanter_rune_of_power", Name: "Rune of Power", Description: "Damage zone protection", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_ring_of_frost"}, Bonuses: StatBonuses{MagicDefense: 7, Intelligence: 4}},
			{ID: "enchanter_arcane_immunity", Name: "Arcane Immunity", Description: "Immune to magic", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"enchanter_evanesce", "enchanter_rune_of_power"}, Bonuses: StatBonuses{MagicDefense: 15, Health: 50}},
		},
		Utility: []TalentDefinition{
			{ID: "enchanter_enchant_weapon", Name: "Enchant Weapon", Description: "Imbue weapons with magic", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 2, Strength: 2}},
			{ID: "enchanter_charm_person", Name: "Charm Person", Description: "Make allies from enemies", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Charisma: 4}},
			{ID: "enchanter_mass_enchantment", Name: "Mass Enchantment", Description: "Enchant multiple items", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"enchanter_enchant_weapon"}, Bonuses: StatBonuses{Intelligence: 4, Charisma: 2}},
			{ID: "enchanter_suggestion", Name: "Suggestion", Description: "Implant ideas", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"enchanter_charm_person"}, Bonuses: StatBonuses{Charisma: 5, Wisdom: 2}},
			{ID: "enchanter_arcane_brilliance", Name: "Arcane Brilliance", Description: "Boost party intellect", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 3, Charisma: 3}},
			{ID: "enchanter_spellsteal", Name: "Spellsteal", Description: "Take enemy buffs", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"enchanter_mass_enchantment"}, Bonuses: StatBonuses{Intelligence: 5}},
			{ID: "enchanter_mass_polymorph", Name: "Mass Polymorph", Description: "Transform many enemies", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"enchanter_suggestion"}, Bonuses: StatBonuses{Charisma: 6, Intelligence: 3}},
			{ID: "enchanter_time_warp_utility", Name: "Time Warp", Description: "Haste for party", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"enchanter_arcane_brilliance"}, Bonuses: StatBonuses{Speed: 0.15, Intelligence: 4}},
			{ID: "enchanter_arcane_mastery", Name: "Arcane Mastery", Description: "Reduced spell costs", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"enchanter_spellsteal"}, Bonuses: StatBonuses{Mana: 40, Intelligence: 5}},
			{ID: "enchanter_supreme_enchanter", Name: "Supreme Enchanter", Description: "Master of enchantments", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"enchanter_mass_polymorph", "enchanter_arcane_mastery"}, Bonuses: StatBonuses{Intelligence: 10, Charisma: 10, Mana: 50}},
		},
	}
}

// createBardTalentTree returns the talent tree for the Bard class
func createBardTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Bard Talents",
		ClassID: ClassBard,
		Offensive: []TalentDefinition{
			{ID: "bard_vicious_mockery", Name: "Vicious Mockery", Description: "Insults that wound", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Charisma: 3, Intelligence: 1}},
			{ID: "bard_dissonant_whispers", Name: "Dissonant Whispers", Description: "Painful sounds", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Charisma: 3, Wisdom: 1}},
			{ID: "bard_thunderwave", Name: "Thunderwave", Description: "Sound blast", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"bard_vicious_mockery"}, Bonuses: StatBonuses{Charisma: 4, Strength: 2}},
			{ID: "bard_shatter", Name: "Shatter", Description: "Resonating damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"bard_dissonant_whispers"}, Bonuses: StatBonuses{Charisma: 4, CritDamage: 0.1}},
			{ID: "bard_hypnotic_pattern", Name: "Hypnotic Pattern", Description: "Mesmerizing display", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"bard_thunderwave"}, Bonuses: StatBonuses{Charisma: 5, Intelligence: 3}},
			{ID: "bard_synaptic_static", Name: "Synaptic Static", Description: "Mind-scrambling noise", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"bard_shatter"}, Bonuses: StatBonuses{Intelligence: 5, CritChance: 0.04}},
			{ID: "bard_psychic_lance", Name: "Psychic Lance", Description: "Mental spear", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"bard_hypnotic_pattern"}, Bonuses: StatBonuses{Charisma: 5, CritDamage: 0.15}},
			{ID: "bard_animate_objects", Name: "Animate Objects", Description: "Objects attack", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"bard_synaptic_static"}, Bonuses: StatBonuses{Charisma: 6, Intelligence: 4}},
			{ID: "bard_power_word_kill", Name: "Power Word Kill", Description: "Death by word", Category: CategoryOffensive, MaxRank: 1, Prerequisites: []TalentID{"bard_psychic_lance"}, Bonuses: StatBonuses{Charisma: 8, CritDamage: 0.5}},
			{ID: "bard_legendary_performer", Name: "Legendary Performer", Description: "Performance kills", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"bard_animate_objects", "bard_power_word_kill"}, Bonuses: StatBonuses{Charisma: 12, Intelligence: 6, CritDamage: 0.3}},
		},
		Defensive: []TalentDefinition{
			{ID: "bard_song_of_rest", Name: "Song of Rest", Description: "Healing melody", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Charisma: 2, Health: 20}},
			{ID: "bard_blade_flourish", Name: "Blade Flourish", Description: "Defensive swordplay", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 3, Defense: 2}},
			{ID: "bard_cutting_words", Name: "Cutting Words", Description: "Distract attackers", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_song_of_rest"}, Bonuses: StatBonuses{Charisma: 4, Defense: 3}},
			{ID: "bard_defensive_flourish", Name: "Defensive Flourish", Description: "Graceful defense", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_blade_flourish"}, Bonuses: StatBonuses{Dexterity: 4, Defense: 4}},
			{ID: "bard_countercharm", Name: "Countercharm", Description: "Resist charm effects", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_cutting_words"}, Bonuses: StatBonuses{MagicDefense: 5, Wisdom: 3}},
			{ID: "bard_mobile_flourish", Name: "Mobile Flourish", Description: "Attack and retreat", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"bard_defensive_flourish"}, Bonuses: StatBonuses{Speed: 0.15, Dexterity: 4}},
			{ID: "bard_mantle_of_inspiration", Name: "Mantle of Inspiration", Description: "Ally temp health", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_countercharm"}, Bonuses: StatBonuses{Charisma: 5, Health: 30}},
			{ID: "bard_unfailing_inspiration", Name: "Unfailing Inspiration", Description: "Inspiration persists", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_mobile_flourish"}, Bonuses: StatBonuses{Charisma: 4, Wisdom: 4}},
			{ID: "bard_glamour", Name: "Mantle of Majesty", Description: "Commanding presence", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"bard_mantle_of_inspiration"}, Bonuses: StatBonuses{Charisma: 6, Defense: 5}},
			{ID: "bard_untouchable_legend", Name: "Untouchable Legend", Description: "Too famous to hit", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"bard_glamour", "bard_unfailing_inspiration"}, Bonuses: StatBonuses{Charisma: 10, Defense: 8, MagicDefense: 6}},
		},
		Utility: []TalentDefinition{
			{ID: "bard_bardic_inspiration", Name: "Bardic Inspiration", Description: "Boost ally rolls", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Charisma: 4}},
			{ID: "bard_jack_of_all_trades", Name: "Jack of All Trades", Description: "Good at everything", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Dexterity: 2, Intelligence: 2, Wisdom: 2}},
			{ID: "bard_expertise", Name: "Expertise", Description: "Master two skills", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_bardic_inspiration"}, Bonuses: StatBonuses{Charisma: 3, Wisdom: 3}},
			{ID: "bard_font_of_inspiration", Name: "Font of Inspiration", Description: "Recharge inspiration", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_jack_of_all_trades"}, Bonuses: StatBonuses{Charisma: 5, Stamina: 20}},
			{ID: "bard_magical_secrets", Name: "Magical Secrets", Description: "Learn any spell", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Intelligence: 4, Mana: 25}},
			{ID: "bard_peerless_skill", Name: "Peerless Skill", Description: "Use inspiration for self", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_expertise"}, Bonuses: StatBonuses{Charisma: 5, Wisdom: 4}},
			{ID: "bard_additional_secrets", Name: "Additional Secrets", Description: "More stolen spells", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_magical_secrets"}, Bonuses: StatBonuses{Intelligence: 5, Mana: 30}},
			{ID: "bard_superior_inspiration", Name: "Superior Inspiration", Description: "Free inspiration", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"bard_font_of_inspiration"}, Bonuses: StatBonuses{Charisma: 6, Stamina: 25}},
			{ID: "bard_words_of_creation", Name: "Words of Creation", Description: "Reality-shaping words", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_additional_secrets"}, Bonuses: StatBonuses{Intelligence: 6, Wisdom: 5}},
			{ID: "bard_master_of_all", Name: "Master of All", Description: "Supreme performer", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"bard_superior_inspiration", "bard_words_of_creation"}, Bonuses: StatBonuses{Charisma: 12, Intelligence: 6, Wisdom: 6}},
		},
	}
}

// createDruidTalentTree returns the talent tree for the Druid class
func createDruidTalentTree() *TalentTree {
	return &TalentTree{
		Name:    "Druid Talents",
		ClassID: ClassDruid,
		Offensive: []TalentDefinition{
			{ID: "druid_wrath", Name: "Wrath", Description: "Nature damage bolt", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Intelligence: 1}},
			{ID: "druid_moonfire", Name: "Moonfire", Description: "Lunar damage over time", Category: CategoryOffensive, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, CritChance: 0.02}},
			{ID: "druid_starfire", Name: "Starfire", Description: "Starlight damage", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"druid_wrath"}, Bonuses: StatBonuses{Wisdom: 4, CritDamage: 0.1}},
			{ID: "druid_sunfire", Name: "Sunfire", Description: "Solar burning", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"druid_moonfire"}, Bonuses: StatBonuses{Wisdom: 4, Intelligence: 2}},
			{ID: "druid_starsurge", Name: "Starsurge", Description: "Cosmic explosion", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"druid_starfire"}, Bonuses: StatBonuses{Wisdom: 5, CritDamage: 0.15}},
			{ID: "druid_stellar_flare", Name: "Stellar Flare", Description: "Combined astral damage", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"druid_sunfire"}, Bonuses: StatBonuses{Wisdom: 5, CritChance: 0.04}},
			{ID: "druid_starfall", Name: "Starfall", Description: "Raining stars", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"druid_starsurge"}, Bonuses: StatBonuses{Wisdom: 6, Mana: 25}},
			{ID: "druid_fury_of_elune", Name: "Fury of Elune", Description: "Moon beam", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"druid_stellar_flare"}, Bonuses: StatBonuses{Wisdom: 7, CritDamage: 0.2}},
			{ID: "druid_celestial_alignment", Name: "Celestial Alignment", Description: "Full astral power", Category: CategoryOffensive, MaxRank: 3, Prerequisites: []TalentID{"druid_starfall"}, Bonuses: StatBonuses{Wisdom: 8, Intelligence: 5}},
			{ID: "druid_incarnation_elune", Name: "Incarnation: Elune", Description: "Become avatar of moon", Category: CategoryOffensive, MaxRank: 5, Prerequisites: []TalentID{"druid_fury_of_elune", "druid_celestial_alignment"}, Bonuses: StatBonuses{Wisdom: 12, CritDamage: 0.35, Mana: 50}},
		},
		Defensive: []TalentDefinition{
			{ID: "druid_barkskin", Name: "Barkskin", Description: "Protective bark", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 4}},
			{ID: "druid_thorns", Name: "Thorns", Description: "Damage attackers", Category: CategoryDefensive, MaxRank: 5, Bonuses: StatBonuses{Defense: 2, Strength: 2}},
			{ID: "druid_ironbark", Name: "Ironbark", Description: "Enhanced protection", Category: CategoryDefensive, MaxRank: 3, Prerequisites: []TalentID{"druid_barkskin"}, Bonuses: StatBonuses{Defense: 6, Health: 30}},
			{ID: "druid_bear_form", Name: "Bear Form", Description: "Transform to bear", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_thorns"}, Bonuses: StatBonuses{Health: 50, Defense: 5}},
			{ID: "druid_frenzied_regeneration", Name: "Frenzied Regeneration", Description: "Rapid healing in form", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_ironbark"}, Bonuses: StatBonuses{Health: 40, Stamina: 20}},
			{ID: "druid_survival_instincts", Name: "Survival Instincts", Description: "Emergency defense", Category: CategoryDefensive, MaxRank: 1, Prerequisites: []TalentID{"druid_bear_form"}, Bonuses: StatBonuses{Health: 100}},
			{ID: "druid_ursols_vortex", Name: "Ursol's Vortex", Description: "Trap enemies", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_frenzied_regeneration"}, Bonuses: StatBonuses{Defense: 4, Wisdom: 3}},
			{ID: "druid_thick_hide", Name: "Thick Hide", Description: "Permanent armor", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_survival_instincts"}, Bonuses: StatBonuses{Defense: 8, Health: 40}},
			{ID: "druid_guardian_of_nature", Name: "Guardian of Nature", Description: "Nature protects you", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_ursols_vortex"}, Bonuses: StatBonuses{Defense: 6, MagicDefense: 5, Health: 30}},
			{ID: "druid_incarnation_guardian", Name: "Incarnation: Guardian", Description: "Avatar of Ursoc", Category: CategoryDefensive, MaxRank: 5, Prerequisites: []TalentID{"druid_thick_hide", "druid_guardian_of_nature"}, Bonuses: StatBonuses{Health: 150, Defense: 12, MagicDefense: 8}},
		},
		Utility: []TalentDefinition{
			{ID: "druid_rejuvenation", Name: "Rejuvenation", Description: "Healing over time", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Mana: 15}},
			{ID: "druid_regrowth", Name: "Regrowth", Description: "Direct heal", Category: CategoryUtility, MaxRank: 5, Bonuses: StatBonuses{Wisdom: 3, Health: 15}},
			{ID: "druid_wild_growth", Name: "Wild Growth", Description: "Group healing", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"druid_rejuvenation"}, Bonuses: StatBonuses{Wisdom: 4, Charisma: 2}},
			{ID: "druid_travel_form", Name: "Travel Form", Description: "Fast movement form", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"druid_regrowth"}, Bonuses: StatBonuses{Speed: 0.2, Stamina: 20}},
			{ID: "druid_innervate", Name: "Innervate", Description: "Restore mana", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"druid_wild_growth"}, Bonuses: StatBonuses{Mana: 40, Wisdom: 3}},
			{ID: "druid_cat_form", Name: "Cat Form", Description: "Stealth predator", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"druid_travel_form"}, Bonuses: StatBonuses{Dexterity: 5, Speed: 0.15}},
			{ID: "druid_tranquility", Name: "Tranquility", Description: "Massive group heal", Category: CategoryUtility, MaxRank: 3, Prerequisites: []TalentID{"druid_innervate"}, Bonuses: StatBonuses{Wisdom: 6, Mana: 35}},
			{ID: "druid_stampeding_roar", Name: "Stampeding Roar", Description: "Group speed boost", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"druid_cat_form"}, Bonuses: StatBonuses{Speed: 0.15, Charisma: 4}},
			{ID: "druid_rebirth", Name: "Rebirth", Description: "Combat resurrection", Category: CategoryUtility, MaxRank: 1, Prerequisites: []TalentID{"druid_tranquility"}, Bonuses: StatBonuses{Wisdom: 8}},
			{ID: "druid_force_of_nature", Name: "Force of Nature", Description: "Summon treants", Category: CategoryUtility, MaxRank: 5, Prerequisites: []TalentID{"druid_stampeding_roar", "druid_rebirth"}, Bonuses: StatBonuses{Wisdom: 10, Charisma: 6, Mana: 50}},
		},
	}
}
