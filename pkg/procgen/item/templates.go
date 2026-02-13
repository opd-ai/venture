// Package item provides item template definitions.
// This file contains item templates for different genres.
// Code relocated from: types.go
package item

// ItemTemplate defines a template for generating items.
type ItemTemplate struct {
	BaseType         ItemType
	WeaponType       WeaponType
	ArmorType        ArmorType
	ConsumableType   ConsumableType
	NamePrefixes     []string
	NameSuffixes     []string
	Tags             []string
	DamageRange      [2]int
	DefenseRange     [2]int
	AttackSpeedRange [2]float64
	ValueRange       [2]int
	WeightRange      [2]float64
	DurabilityRange  [2]int

	// Projectile template properties
	IsProjectile         bool
	ProjectileSpeedRange [2]float64
	ProjectileLifetime   float64
	ProjectileType       string
	PierceChance         float64 // Probability of having pierce ability
	PierceRange          [2]int  // Range of pierce count if generated
	BounceChance         float64 // Probability of having bounce ability
	BounceRange          [2]int  // Range of bounce count if generated
	ExplosiveChance      float64 // Probability of being explosive
	ExplosionRadiusRange [2]float64

	// Class restriction fields
	// ClassRestrictions lists classes that can use items from this template (empty = any class)
	ClassRestrictions []string

	// Spell effect fields for consumable scrolls
	// SpellEffectIDs is a pool of spell IDs this template can generate
	SpellEffectIDs []string
	// SpellDurations are corresponding durations (parallel array with SpellEffectIDs)
	SpellDurations []float64
	// SpellTargetTypes are corresponding target types (parallel array with SpellEffectIDs)
	SpellTargetTypes []string
	// SpellRadii are corresponding radii (parallel array with SpellEffectIDs)
	SpellRadii []float64
}

// GetFantasyWeaponTemplates returns weapon templates for fantasy genre.
func GetFantasyWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponSword,
			NamePrefixes:      []string{"Iron", "Steel", "Silver", "Elven", "Dwarven"},
			NameSuffixes:      []string{"Sword", "Blade", "Saber", "Longsword", "Cutlass"},
			Tags:              []string{"balanced", "versatile"},
			DamageRange:       [2]int{8, 15},
			AttackSpeedRange:  [2]float64{1.0, 1.2},
			ValueRange:        [2]int{50, 200},
			WeightRange:       [2]float64{3.0, 5.0},
			DurabilityRange:   [2]int{80, 120},
			ClassRestrictions: []string{}, // Swords usable by any class
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponAxe,
			NamePrefixes:      []string{"Battle", "War", "Great", "Heavy", "Brutal"},
			NameSuffixes:      []string{"Axe", "Hammer", "Mace", "Cleaver"},
			Tags:              []string{"heavy", "powerful", "slow"},
			DamageRange:       [2]int{12, 20},
			AttackSpeedRange:  [2]float64{0.7, 0.9},
			ValueRange:        [2]int{60, 250},
			WeightRange:       [2]float64{6.0, 10.0},
			DurabilityRange:   [2]int{100, 150},
			ClassRestrictions: []string{"warrior", "ranger"}, // Heavy weapons
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponBow,
			NamePrefixes:     []string{"Hunter's", "Ranger's", "Composite", "Long", "Elven"},
			NameSuffixes:     []string{"Bow", "Longbow", "Shortbow", "Crossbow"},
			Tags:             []string{"ranged", "precise"},
			DamageRange:      [2]int{6, 12},
			AttackSpeedRange: [2]float64{1.2, 1.5},
			ValueRange:       [2]int{40, 180},
			WeightRange:      [2]float64{2.0, 4.0},
			DurabilityRange:  [2]int{60, 100},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{300.0, 500.0}, // pixels per second
			ProjectileLifetime:   3.0,                      // seconds
			ProjectileType:       "arrow",
			PierceChance:         0.15, // 15% chance for piercing arrows
			PierceRange:          [2]int{1, 2},
			BounceChance:         0.0,
			ExplosiveChance:      0.05, // 5% chance for explosive arrows (rare)
			ExplosionRadiusRange: [2]float64{40.0, 60.0},
			ClassRestrictions:    []string{"ranger", "rogue"}, // Ranged specialists
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponStaff,
			NamePrefixes:      []string{"Wizard's", "Arcane", "Mystic", "Elder", "Ancient"},
			NameSuffixes:      []string{"Staff", "Rod", "Wand", "Scepter"},
			Tags:              []string{"magical", "casting"},
			DamageRange:       [2]int{5, 10},
			AttackSpeedRange:  [2]float64{0.8, 1.0},
			ValueRange:        [2]int{80, 300},
			WeightRange:       [2]float64{1.5, 3.0},
			DurabilityRange:   [2]int{50, 80},
			ClassRestrictions: []string{"mage", "cleric", "necromancer"}, // Magic users
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponDagger,
			NamePrefixes:      []string{"Sharp", "Quick", "Silent", "Poison", "Shadow"},
			NameSuffixes:      []string{"Dagger", "Knife", "Stiletto", "Dirk"},
			Tags:              []string{"fast", "stealth", "light"},
			DamageRange:       [2]int{4, 8},
			AttackSpeedRange:  [2]float64{1.5, 2.0},
			ValueRange:        [2]int{30, 150},
			WeightRange:       [2]float64{0.5, 1.5},
			DurabilityRange:   [2]int{40, 70},
			ClassRestrictions: []string{"rogue", "ranger"}, // Light weapon specialists
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponCrossbow,
			NamePrefixes:     []string{"Heavy", "Repeating", "Siege", "Hand", "Arbalest"},
			NameSuffixes:     []string{"Crossbow", "Arbalest", "Ballista"},
			Tags:             []string{"ranged", "powerful", "slow"},
			DamageRange:      [2]int{10, 18},
			AttackSpeedRange: [2]float64{0.6, 0.9},
			ValueRange:       [2]int{70, 250},
			WeightRange:      [2]float64{5.0, 8.0},
			DurabilityRange:  [2]int{80, 130},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{400.0, 600.0}, // faster than arrows
			ProjectileLifetime:   2.5,                      // shorter lifetime
			ProjectileType:       "bolt",
			PierceChance:         0.25, // 25% chance for piercing bolts
			PierceRange:          [2]int{1, 3},
			BounceChance:         0.0,
			ExplosiveChance:      0.10, // 10% chance for explosive bolts
			ExplosionRadiusRange: [2]float64{50.0, 80.0},
			ClassRestrictions:    []string{"warrior", "ranger"}, // Heavy ranged
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponWand,
			NamePrefixes:     []string{"Fire", "Ice", "Lightning", "Arcane", "Shadow"},
			NameSuffixes:     []string{"Wand", "Rod", "Focus", "Conduit"},
			Tags:             []string{"magical", "ranged", "elemental"},
			DamageRange:      [2]int{7, 14},
			AttackSpeedRange: [2]float64{1.3, 1.7},
			ValueRange:       [2]int{90, 320},
			WeightRange:      [2]float64{0.8, 2.0},
			DurabilityRange:  [2]int{60, 90},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{250.0, 400.0}, // slower magical projectiles
			ProjectileLifetime:   4.0,                      // longer lifetime for magic
			ProjectileType:       "magic_bolt",
			PierceChance:         0.20, // 20% chance for piercing magic
			PierceRange:          [2]int{1, 2},
			BounceChance:         0.10, // 10% chance for bouncing magic
			BounceRange:          [2]int{1, 2},
			ExplosiveChance:      0.15, // 15% chance for explosive magic
			ExplosionRadiusRange: [2]float64{60.0, 100.0},
			ClassRestrictions:    []string{"mage", "cleric", "necromancer"}, // Magic users
		},
	}
}

// GetFantasyArmorTemplates returns armor templates for fantasy genre.
func GetFantasyArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorChest,
			NamePrefixes:      []string{"Leather", "Chain", "Plate", "Scale", "Dragon"},
			NameSuffixes:      []string{"Armor", "Cuirass", "Breastplate", "Mail"},
			Tags:              []string{"protection", "heavy"},
			DefenseRange:      [2]int{10, 30},
			ValueRange:        [2]int{100, 400},
			WeightRange:       [2]float64{8.0, 20.0},
			DurabilityRange:   [2]int{120, 200},
			ClassRestrictions: []string{"warrior", "cleric"}, // Heavy armor users
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorHelmet,
			NamePrefixes:      []string{"Iron", "Steel", "Knight's", "Great", "Horned"},
			NameSuffixes:      []string{"Helmet", "Helm", "Crown", "Cap"},
			Tags:              []string{"protection", "head"},
			DefenseRange:      [2]int{5, 15},
			ValueRange:        [2]int{50, 200},
			WeightRange:       [2]float64{2.0, 5.0},
			DurabilityRange:   [2]int{80, 120},
			ClassRestrictions: []string{"warrior", "cleric"}, // Heavy armor users
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorShield,
			NamePrefixes:      []string{"Wooden", "Iron", "Steel", "Tower", "Kite"},
			NameSuffixes:      []string{"Shield", "Buckler", "Guard"},
			Tags:              []string{"block", "defense"},
			DefenseRange:      [2]int{8, 20},
			ValueRange:        [2]int{40, 180},
			WeightRange:       [2]float64{4.0, 12.0},
			DurabilityRange:   [2]int{100, 150},
			ClassRestrictions: []string{"warrior", "cleric"}, // Shield users
		},
	}
}

// GetFantasyConsumableTemplates returns consumable templates for fantasy genre.
func GetFantasyConsumableTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:       TypeConsumable,
			ConsumableType: ConsumablePotion,
			NamePrefixes:   []string{"Minor", "Lesser", "Greater", "Superior", "Ultimate"},
			NameSuffixes:   []string{"Health Potion", "Mana Potion", "Stamina Potion"},
			Tags:           []string{"healing", "consumable"},
			ValueRange:     [2]int{10, 100},
			WeightRange:    [2]float64{0.1, 0.3},
		},
		{
			BaseType:         TypeConsumable,
			ConsumableType:   ConsumableScroll,
			NamePrefixes:     []string{"Scroll of", "Ancient", "Mystic"},
			NameSuffixes:     []string{"Fireball", "Lightning", "Ice", "Protection"},
			Tags:             []string{"magical", "spell", "consumable"},
			ValueRange:       [2]int{20, 150},
			WeightRange:      [2]float64{0.1, 0.2},
			SpellEffectIDs:   []string{"fireball", "lightning", "ice_nova", "protection_ward"},
			SpellDurations:   []float64{0.0, 0.0, 0.0, 10.0}, // 0.0 = instant
			SpellTargetTypes: []string{"area", "entity", "area", "self"},
			SpellRadii:       []float64{80.0, 0.0, 100.0, 0.0},
		},
	}
}

// GetSciFiWeaponTemplates returns weapon templates for sci-fi genre.
func GetSciFiWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponSword, // Using as energy blade
			NamePrefixes:      []string{"Plasma", "Energy", "Photon", "Quantum", "Nano"},
			NameSuffixes:      []string{"Blade", "Saber", "Cutter", "Sword"},
			Tags:              []string{"energy", "melee"},
			DamageRange:       [2]int{10, 18},
			AttackSpeedRange:  [2]float64{1.2, 1.5},
			ValueRange:        [2]int{150, 500},
			WeightRange:       [2]float64{1.0, 2.0},
			DurabilityRange:   [2]int{200, 300},
			ClassRestrictions: []string{}, // Energy blades usable by any class
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponGun,
			NamePrefixes:     []string{"Laser", "Pulse", "Plasma", "Rail", "Ion"},
			NameSuffixes:     []string{"Rifle", "Pistol", "Cannon", "Blaster"},
			Tags:             []string{"energy", "ranged"},
			DamageRange:      [2]int{8, 15},
			AttackSpeedRange: [2]float64{1.5, 2.0},
			ValueRange:       [2]int{200, 600},
			WeightRange:      [2]float64{2.0, 5.0},
			DurabilityRange:  [2]int{150, 250},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{600.0, 1000.0}, // very fast bullets
			ProjectileLifetime:   2.0,                       // short lifetime
			ProjectileType:       "bullet",
			PierceChance:         0.30, // 30% chance for armor-piercing rounds
			PierceRange:          [2]int{2, 4},
			BounceChance:         0.05, // 5% chance for ricochet rounds
			BounceRange:          [2]int{1, 2},
			ExplosiveChance:      0.20, // 20% chance for explosive rounds
			ExplosionRadiusRange: [2]float64{40.0, 70.0},
			ClassRestrictions:    []string{}, // Guns usable by any class
		},
	}
}

// GetSciFiArmorTemplates returns armor templates for sci-fi genre.
func GetSciFiArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorChest,
			NamePrefixes:      []string{"Combat", "Battle", "Tactical", "Power", "Nano"},
			NameSuffixes:      []string{"Suit", "Armor", "Exosuit", "Vest"},
			Tags:              []string{"powered", "armored"},
			DefenseRange:      [2]int{15, 35},
			ValueRange:        [2]int{300, 800},
			WeightRange:       [2]float64{5.0, 15.0},
			DurabilityRange:   [2]int{200, 350},
			ClassRestrictions: []string{}, // Sci-fi armor usable by any class
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorHelmet,
			NamePrefixes:      []string{"Combat", "Battle", "Tactical", "HUD", "Neural"},
			NameSuffixes:      []string{"Helmet", "Visor", "Interface"},
			Tags:              []string{"hud", "scanning"},
			DefenseRange:      [2]int{8, 18},
			ValueRange:        [2]int{150, 400},
			WeightRange:       [2]float64{1.0, 3.0},
			DurabilityRange:   [2]int{150, 250},
			ClassRestrictions: []string{}, // Sci-fi armor usable by any class
		},
	}
}

// GetSciFiConsumableTemplates returns consumable templates for sci-fi genre.
func GetSciFiConsumableTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:       TypeConsumable,
			ConsumableType: ConsumablePotion,
			NamePrefixes:   []string{"Standard", "Advanced", "Military", "Experimental", "Prototype"},
			NameSuffixes:   []string{"Med Pack", "Stim Pack", "Bio Gel", "Nanite Serum"},
			Tags:           []string{"medical", "tech", "healing"},
			ValueRange:     [2]int{25, 150},
			WeightRange:    [2]float64{0.1, 0.3},
		},
		{
			BaseType:         TypeConsumable,
			ConsumableType:   ConsumableScroll, // Using scroll type for data chips/devices
			NamePrefixes:     []string{"Data Chip:", "Neural Link:", "System Hack:"},
			NameSuffixes:     []string{"Shield Boost", "EMP Burst", "Cloak Protocol", "Repair Nanites"},
			Tags:             []string{"tech", "data", "consumable"},
			ValueRange:       [2]int{50, 250},
			WeightRange:      [2]float64{0.05, 0.1},
			SpellEffectIDs:   []string{"shield_boost", "emp_burst", "cloak", "repair_nanites"},
			SpellDurations:   []float64{15.0, 0.0, 10.0, 5.0},
			SpellTargetTypes: []string{"self", "area", "self", "self"},
			SpellRadii:       []float64{0.0, 120.0, 0.0, 0.0},
		},
	}
}

// GetHorrorWeaponTemplates returns weapon templates for horror genre.
func GetHorrorWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponAxe,
			NamePrefixes:      []string{"Rusty", "Blood-Stained", "Jagged", "Cursed", "Corroded"},
			NameSuffixes:      []string{"Axe", "Cleaver", "Machete", "Hatchet"},
			Tags:              []string{"brutal", "makeshift"},
			DamageRange:       [2]int{10, 18},
			AttackSpeedRange:  [2]float64{0.8, 1.0},
			ValueRange:        [2]int{20, 100},
			WeightRange:       [2]float64{3.0, 6.0},
			DurabilityRange:   [2]int{40, 80},
			ClassRestrictions: []string{}, // Survival weapons usable by any class
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponDagger,
			NamePrefixes:      []string{"Ritual", "Bone", "Serrated", "Sacrificial", "Obsidian"},
			NameSuffixes:      []string{"Knife", "Shiv", "Blade", "Scalpel"},
			Tags:              []string{"occult", "silent"},
			DamageRange:       [2]int{5, 10},
			AttackSpeedRange:  [2]float64{1.4, 1.8},
			ValueRange:        [2]int{15, 80},
			WeightRange:       [2]float64{0.3, 1.0},
			DurabilityRange:   [2]int{30, 60},
			ClassRestrictions: []string{}, // Light weapons usable by any class
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponGun,
			NamePrefixes:     []string{"Worn", "Antique", "Salvaged", "Modified", "Silenced"},
			NameSuffixes:     []string{"Revolver", "Shotgun", "Pistol", "Rifle"},
			Tags:             []string{"firearm", "loud", "limited_ammo"},
			DamageRange:      [2]int{12, 22},
			AttackSpeedRange: [2]float64{0.5, 0.8},
			ValueRange:       [2]int{80, 300},
			WeightRange:      [2]float64{2.0, 6.0},
			DurabilityRange:  [2]int{60, 100},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{500.0, 800.0},
			ProjectileLifetime:   1.5,
			ProjectileType:       "bullet",
			PierceChance:         0.20,
			PierceRange:          [2]int{1, 2},
			BounceChance:         0.0,
			ExplosiveChance:      0.05, // Rare explosive rounds
			ExplosionRadiusRange: [2]float64{30.0, 50.0},
			ClassRestrictions:    []string{}, // Guns usable by any class
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponStaff, // Using as occult focus
			NamePrefixes:      []string{"Eldritch", "Forbidden", "Sanity-Draining", "Void", "Abyssal"},
			NameSuffixes:      []string{"Tome", "Grimoire", "Focus", "Idol"},
			Tags:              []string{"occult", "sanity_cost", "dark_magic"},
			DamageRange:       [2]int{8, 15},
			AttackSpeedRange:  [2]float64{0.6, 0.9},
			ValueRange:        [2]int{100, 400},
			WeightRange:       [2]float64{1.0, 2.5},
			DurabilityRange:   [2]int{40, 70},
			ClassRestrictions: []string{"mage", "necromancer"}, // Occult knowledge required
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponSword,
			NamePrefixes:      []string{"Pipe", "Crowbar", "Broken", "Improvised", "Found"},
			NameSuffixes:      []string{"Wrench", "Bar", "Club", "Baton"},
			Tags:              []string{"blunt", "makeshift", "common"},
			DamageRange:       [2]int{6, 12},
			AttackSpeedRange:  [2]float64{1.0, 1.3},
			ValueRange:        [2]int{5, 30},
			WeightRange:       [2]float64{2.0, 4.0},
			DurabilityRange:   [2]int{50, 90},
			ClassRestrictions: []string{}, // Improvised weapons usable by any class
		},
	}
}

// GetHorrorArmorTemplates returns armor templates for horror genre.
func GetHorrorArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorChest,
			NamePrefixes:      []string{"Tattered", "Bloodied", "Reinforced", "Scavenged", "Stitched"},
			NameSuffixes:      []string{"Jacket", "Vest", "Coat", "Overalls"},
			Tags:              []string{"light", "worn"},
			DefenseRange:      [2]int{4, 12},
			ValueRange:        [2]int{20, 100},
			WeightRange:       [2]float64{2.0, 5.0},
			DurabilityRange:   [2]int{40, 80},
			ClassRestrictions: []string{}, // Survival gear usable by any class
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorHelmet,
			NamePrefixes:      []string{"Gas", "Construction", "Riot", "Cracked", "Salvaged"},
			NameSuffixes:      []string{"Mask", "Helmet", "Hood", "Goggles"},
			Tags:              []string{"protective", "visibility"},
			DefenseRange:      [2]int{2, 8},
			ValueRange:        [2]int{15, 60},
			WeightRange:       [2]float64{0.5, 2.0},
			DurabilityRange:   [2]int{30, 60},
			ClassRestrictions: []string{}, // Protective gear usable by any class
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorShield, // Using as improvised shield
			NamePrefixes:      []string{"Car Door", "Riot", "Improvised", "Dented", "Trash Lid"},
			NameSuffixes:      []string{"Shield", "Barrier", "Guard", "Blocker"},
			Tags:              []string{"makeshift", "heavy"},
			DefenseRange:      [2]int{6, 14},
			ValueRange:        [2]int{10, 50},
			WeightRange:       [2]float64{5.0, 10.0},
			DurabilityRange:   [2]int{50, 100},
			ClassRestrictions: []string{}, // Improvised shields usable by any class
		},
	}
}

// GetHorrorConsumableTemplates returns consumable templates for horror genre.
func GetHorrorConsumableTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:       TypeConsumable,
			ConsumableType: ConsumablePotion,
			NamePrefixes:   []string{"Emergency", "Expired", "Homemade", "Contraband", "Bootleg"},
			NameSuffixes:   []string{"First Aid Kit", "Bandages", "Painkillers", "Adrenaline Shot"},
			Tags:           []string{"medical", "scarce", "survival"},
			ValueRange:     [2]int{15, 80},
			WeightRange:    [2]float64{0.2, 0.5},
		},
		{
			BaseType:         TypeConsumable,
			ConsumableType:   ConsumableScroll, // Using scroll type for occult rituals
			NamePrefixes:     []string{"Forbidden Ritual:", "Dark Incantation:", "Blood Rite:"},
			NameSuffixes:     []string{"Banishment", "Ward", "Sight Beyond", "Summoning"},
			Tags:             []string{"occult", "sanity_cost", "ritual"},
			ValueRange:       [2]int{40, 200},
			WeightRange:      [2]float64{0.1, 0.3},
			SpellEffectIDs:   []string{"banish_horror", "protection_ward", "dark_sight", "summon_ally"},
			SpellDurations:   []float64{0.0, 30.0, 20.0, 60.0},
			SpellTargetTypes: []string{"area", "self", "self", "area"},
			SpellRadii:       []float64{100.0, 0.0, 0.0, 50.0},
		},
	}
}

// GetCyberpunkWeaponTemplates returns weapon templates for cyberpunk genre.
func GetCyberpunkWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponSword, // Using as monofilament blade
			NamePrefixes:      []string{"Mono", "Thermal", "Vibro", "Smart", "Razor"},
			NameSuffixes:      []string{"Blade", "Katana", "Wire", "Edge"},
			Tags:              []string{"high-tech", "melee", "precision"},
			DamageRange:       [2]int{12, 20},
			AttackSpeedRange:  [2]float64{1.3, 1.6},
			ValueRange:        [2]int{200, 600},
			WeightRange:       [2]float64{0.8, 1.5},
			DurabilityRange:   [2]int{150, 250},
			ClassRestrictions: []string{}, // High-tech blades usable by any class
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponGun,
			NamePrefixes:     []string{"Smart", "Tracking", "Power", "Tech", "Ricochet"},
			NameSuffixes:     []string{"Pistol", "SMG", "Shotgun", "Assault Rifle"},
			Tags:             []string{"smart_link", "high_capacity"},
			DamageRange:      [2]int{9, 16},
			AttackSpeedRange: [2]float64{1.6, 2.2},
			ValueRange:       [2]int{250, 700},
			WeightRange:      [2]float64{1.5, 4.0},
			DurabilityRange:  [2]int{180, 280},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{700.0, 1200.0}, // Very fast smart rounds
			ProjectileLifetime:   2.0,
			ProjectileType:       "smart_bullet",
			PierceChance:         0.35, // High chance for AP rounds
			PierceRange:          [2]int{2, 5},
			BounceChance:         0.15, // Smart ricochet
			BounceRange:          [2]int{1, 3},
			ExplosiveChance:      0.25, // High-tech explosive rounds
			ExplosionRadiusRange: [2]float64{45.0, 75.0},
			ClassRestrictions:    []string{}, // Guns usable by any class
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponCrossbow, // Using as tech crossbow/launcher
			NamePrefixes:     []string{"Grenade", "Micro-Missile", "Drone", "Shock", "Gas"},
			NameSuffixes:     []string{"Launcher", "Deployer", "Projector", "System"},
			Tags:             []string{"heavy", "explosive", "area_effect"},
			DamageRange:      [2]int{15, 25},
			AttackSpeedRange: [2]float64{0.4, 0.7},
			ValueRange:       [2]int{400, 1000},
			WeightRange:      [2]float64{4.0, 8.0},
			DurabilityRange:  [2]int{120, 200},
			// Projectile properties
			IsProjectile:         true,
			ProjectileSpeedRange: [2]float64{200.0, 400.0}, // Slower ordinance
			ProjectileLifetime:   3.0,
			ProjectileType:       "grenade",
			PierceChance:         0.0,
			BounceChance:         0.0,
			ExplosiveChance:      0.90, // Almost always explosive
			ExplosionRadiusRange: [2]float64{80.0, 150.0},
			ClassRestrictions:    []string{}, // Heavy weapons usable by any class
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponWand, // Using as hacking device
			NamePrefixes:      []string{"Neural", "ICE-Breaker", "Black", "Daemon", "Phantom"},
			NameSuffixes:      []string{"Deck", "Interface", "Jack", "Link"},
			Tags:              []string{"hacking", "digital", "stealth"},
			DamageRange:       [2]int{6, 12}, // Low physical, but affects machines
			AttackSpeedRange:  [2]float64{1.0, 1.4},
			ValueRange:        [2]int{300, 800},
			WeightRange:       [2]float64{0.3, 1.0},
			DurabilityRange:   [2]int{100, 180},
			ClassRestrictions: []string{"rogue", "mage"}, // Netrunners and hackers
		},
		{
			BaseType:          TypeWeapon,
			WeaponType:        WeaponAxe, // Using as mantis blades / gorilla arms
			NamePrefixes:      []string{"Mantis", "Gorilla", "Monowire", "Projectile", "Grapple"},
			NameSuffixes:      []string{"Arms", "Blades", "Fists", "Launcher"},
			Tags:              []string{"cyberware", "implant", "melee"},
			DamageRange:       [2]int{14, 24},
			AttackSpeedRange:  [2]float64{0.9, 1.2},
			ValueRange:        [2]int{500, 1200},
			WeightRange:       [2]float64{0.0, 0.0}, // Implanted, no carry weight
			DurabilityRange:   [2]int{200, 400},
			ClassRestrictions: []string{}, // Cyberware usable by any class
		},
	}
}

// GetCyberpunkArmorTemplates returns armor templates for cyberpunk genre.
func GetCyberpunkArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorChest,
			NamePrefixes:      []string{"Armored", "Tactical", "Corporate", "Street", "Military"},
			NameSuffixes:      []string{"Jacket", "Vest", "Bodysuit", "Skinweave"},
			Tags:              []string{"ballistic", "stylish"},
			DefenseRange:      [2]int{12, 30},
			ValueRange:        [2]int{200, 600},
			WeightRange:       [2]float64{3.0, 8.0},
			DurabilityRange:   [2]int{150, 280},
			ClassRestrictions: []string{}, // Armor usable by any class
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorHelmet,
			NamePrefixes:      []string{"Smart", "Tactical", "Neural", "Combat", "Synth"},
			NameSuffixes:      []string{"Visor", "Helmet", "Interface", "Optics"},
			Tags:              []string{"hud", "threat_detection", "comms"},
			DefenseRange:      [2]int{6, 16},
			ValueRange:        [2]int{150, 450},
			WeightRange:       [2]float64{0.8, 2.5},
			DurabilityRange:   [2]int{120, 220},
			ClassRestrictions: []string{}, // Headgear usable by any class
		},
		{
			BaseType:          TypeArmor,
			ArmorType:         ArmorShield, // Using as deployable barrier/drone
			NamePrefixes:      []string{"Deployable", "Holographic", "Kinetic", "Drone", "Smart"},
			NameSuffixes:      []string{"Barrier", "Shield", "Cover", "Wall"},
			Tags:              []string{"tech", "deployable", "temporary"},
			DefenseRange:      [2]int{10, 25},
			ValueRange:        [2]int{250, 550},
			WeightRange:       [2]float64{2.0, 5.0},
			DurabilityRange:   [2]int{80, 150},
			ClassRestrictions: []string{}, // Tech shields usable by any class
		},
	}
}

// GetCyberpunkConsumableTemplates returns consumable templates for cyberpunk genre.
func GetCyberpunkConsumableTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:       TypeConsumable,
			ConsumableType: ConsumablePotion,
			NamePrefixes:   []string{"Trauma Team", "Street", "Military Grade", "Black Market", "Synth"},
			NameSuffixes:   []string{"MaxDoc", "Bounce Back", "Biomonitor", "Blood Pack"},
			Tags:           []string{"medical", "quick_heal", "combat_stim"},
			ValueRange:     [2]int{30, 180},
			WeightRange:    [2]float64{0.1, 0.4},
		},
		{
			BaseType:         TypeConsumable,
			ConsumableType:   ConsumableScroll, // Using scroll type for quickhacks/programs
			NamePrefixes:     []string{"Quickhack:", "Daemon:", "ICE:", "Virus:"},
			NameSuffixes:     []string{"Contagion", "Short Circuit", "Synapse Burnout", "Memory Wipe"},
			Tags:             []string{"netrunning", "hack", "digital_attack"},
			ValueRange:       [2]int{60, 300},
			WeightRange:      [2]float64{0.0, 0.1},
			SpellEffectIDs:   []string{"contagion", "short_circuit", "synapse_burnout", "memory_wipe"},
			SpellDurations:   []float64{8.0, 0.0, 5.0, 15.0},
			SpellTargetTypes: []string{"entity", "entity", "entity", "area"},
			SpellRadii:       []float64{0.0, 0.0, 0.0, 60.0},
		},
	}
}
