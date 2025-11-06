// Package item provides item type definitions.
// This file defines item types, rarity, slots, and classification
// used by the item generator.
package item

// ItemType represents the classification of an item.
type ItemType int

const (
	// TypeWeapon represents offensive equipment
	TypeWeapon ItemType = iota
	// TypeArmor represents defensive equipment
	TypeArmor
	// TypeConsumable represents single-use items (potions, scrolls)
	TypeConsumable
	// TypeAccessory represents stat-boosting equipment
	TypeAccessory
)

// String returns the string representation of an item type.
func (t ItemType) String() string {
	switch t {
	case TypeWeapon:
		return "weapon"
	case TypeArmor:
		return "armor"
	case TypeConsumable:
		return "consumable"
	case TypeAccessory:
		return "accessory"
	default:
		return "unknown"
	}
}

// WeaponType represents the category of weapon.
type WeaponType int

const (
	// WeaponSword represents one-handed or two-handed swords
	WeaponSword WeaponType = iota
	// WeaponAxe represents axes and hammers
	WeaponAxe
	// WeaponBow represents ranged weapons (arrows)
	WeaponBow
	// WeaponStaff represents magical weapons
	WeaponStaff
	// WeaponDagger represents fast, light weapons
	WeaponDagger
	// WeaponSpear represents reach weapons
	WeaponSpear
	// WeaponCrossbow represents heavy ranged weapons (bolts)
	WeaponCrossbow
	// WeaponGun represents sci-fi ranged weapons (bullets)
	WeaponGun
	// WeaponWand represents magical projectile weapons (spells)
	WeaponWand
)

// String returns the string representation of a weapon type.
func (w WeaponType) String() string {
	switch w {
	case WeaponSword:
		return "sword"
	case WeaponAxe:
		return "axe"
	case WeaponBow:
		return "bow"
	case WeaponStaff:
		return "staff"
	case WeaponDagger:
		return "dagger"
	case WeaponSpear:
		return "spear"
	case WeaponCrossbow:
		return "crossbow"
	case WeaponGun:
		return "gun"
	case WeaponWand:
		return "wand"
	default:
		return "unknown"
	}
}

// ArmorType represents the category of armor.
type ArmorType int

const (
	// ArmorHelmet protects the head
	ArmorHelmet ArmorType = iota
	// ArmorChest protects the torso
	ArmorChest
	// ArmorLegs protects the legs
	ArmorLegs
	// ArmorBoots protects the feet
	ArmorBoots
	// ArmorGloves protects the hands
	ArmorGloves
	// ArmorShield provides additional defense
	ArmorShield
)

// String returns the string representation of an armor type.
func (a ArmorType) String() string {
	switch a {
	case ArmorHelmet:
		return "helmet"
	case ArmorChest:
		return "chest"
	case ArmorLegs:
		return "legs"
	case ArmorBoots:
		return "boots"
	case ArmorGloves:
		return "gloves"
	case ArmorShield:
		return "shield"
	default:
		return "unknown"
	}
}

// ConsumableType represents the category of consumable.
type ConsumableType int

const (
	// ConsumablePotion restores health or provides buffs
	ConsumablePotion ConsumableType = iota
	// ConsumableScroll provides one-time spell effects
	ConsumableScroll
	// ConsumableFood restores health over time
	ConsumableFood
	// ConsumableBomb deals area damage
	ConsumableBomb
)

// String returns the string representation of a consumable type.
func (c ConsumableType) String() string {
	switch c {
	case ConsumablePotion:
		return "potion"
	case ConsumableScroll:
		return "scroll"
	case ConsumableFood:
		return "food"
	case ConsumableBomb:
		return "bomb"
	default:
		return "unknown"
	}
}

// Rarity represents how rare/special an item is.
type Rarity int

const (
	// RarityCommon represents frequently found items
	RarityCommon Rarity = iota
	// RarityUncommon represents moderately rare items
	RarityUncommon
	// RarityRare represents rare items with better stats
	RarityRare
	// RarityEpic represents very rare, powerful items
	RarityEpic
	// RarityLegendary represents extremely rare, unique items
	RarityLegendary
)

// String returns the string representation of a rarity level.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "common"
	case RarityUncommon:
		return "uncommon"
	case RarityRare:
		return "rare"
	case RarityEpic:
		return "epic"
	case RarityLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// Stats represents the core statistics of an item.
type Stats struct {
	// Damage for weapons
	Damage int
	// Defense for armor
	Defense int
	// AttackSpeed for weapons (higher is faster)
	AttackSpeed float64
	// Value is the base price
	Value int
	// Weight affects carrying capacity
	Weight float64
	// RequiredLevel to use the item
	RequiredLevel int
	// DurabilityMax is the maximum durability
	DurabilityMax int
	// Durability is the current durability
	Durability int

	// Projectile properties for ranged weapons
	// IsProjectile indicates if this weapon fires projectiles
	IsProjectile bool
	// ProjectileSpeed in pixels per second (0 for non-projectile weapons)
	ProjectileSpeed float64
	// ProjectileLifetime in seconds before projectile despawns
	ProjectileLifetime float64
	// ProjectileType describes the projectile ("arrow", "bullet", "fireball", etc.)
	ProjectileType string
	// Pierce is the number of enemies projectile can pass through (0 = normal, -1 = infinite)
	Pierce int
	// Bounce is the number of wall bounces (0 = despawn on wall hit)
	Bounce int
	// Explosive indicates if projectile explodes on impact
	Explosive bool
	// ExplosionRadius in pixels (if Explosive)
	ExplosionRadius float64
}

// Item represents a generated game item.
type Item struct {
	// ID is a unique identifier for this item
	ID string
	// Name is the procedurally generated name
	Name string
	// Type categorizes the item
	Type ItemType
	// WeaponType if this is a weapon
	WeaponType WeaponType
	// ArmorType if this is armor
	ArmorType ArmorType
	// ConsumableType if this is a consumable
	ConsumableType ConsumableType
	// Rarity indicates how special/rare the item is
	Rarity Rarity
	// Stats contains all numerical attributes
	Stats Stats
	// Seed is the generation seed for this item
	Seed int64
	// Tags are additional descriptive labels
	Tags []string
	// Description is a generated flavor text
	Description string
}

// IsEquippable returns true if the item can be equipped.
func (i *Item) IsEquippable() bool {
	return i.Type == TypeWeapon || i.Type == TypeArmor || i.Type == TypeAccessory
}

// IsConsumable returns true if the item is consumed on use.
func (i *Item) IsConsumable() bool {
	return i.Type == TypeConsumable
}

// GetValue returns the item's value modified by condition.
func (i *Item) GetValue() int {
	if i.Stats.DurabilityMax == 0 {
		return i.Stats.Value
	}
	// Reduce value based on damage
	condition := float64(i.Stats.Durability) / float64(i.Stats.DurabilityMax)
	return int(float64(i.Stats.Value) * condition)
}

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
}

// GetFantasyWeaponTemplates returns weapon templates for fantasy genre.
func GetFantasyWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponSword,
			NamePrefixes:     []string{"Iron", "Steel", "Silver", "Elven", "Dwarven"},
			NameSuffixes:     []string{"Sword", "Blade", "Saber", "Longsword", "Cutlass"},
			Tags:             []string{"balanced", "versatile"},
			DamageRange:      [2]int{8, 15},
			AttackSpeedRange: [2]float64{1.0, 1.2},
			ValueRange:       [2]int{50, 200},
			WeightRange:      [2]float64{3.0, 5.0},
			DurabilityRange:  [2]int{80, 120},
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponAxe,
			NamePrefixes:     []string{"Battle", "War", "Great", "Heavy", "Brutal"},
			NameSuffixes:     []string{"Axe", "Hammer", "Mace", "Cleaver"},
			Tags:             []string{"heavy", "powerful", "slow"},
			DamageRange:      [2]int{12, 20},
			AttackSpeedRange: [2]float64{0.7, 0.9},
			ValueRange:       [2]int{60, 250},
			WeightRange:      [2]float64{6.0, 10.0},
			DurabilityRange:  [2]int{100, 150},
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
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponStaff,
			NamePrefixes:     []string{"Wizard's", "Arcane", "Mystic", "Elder", "Ancient"},
			NameSuffixes:     []string{"Staff", "Rod", "Wand", "Scepter"},
			Tags:             []string{"magical", "casting"},
			DamageRange:      [2]int{5, 10},
			AttackSpeedRange: [2]float64{0.8, 1.0},
			ValueRange:       [2]int{80, 300},
			WeightRange:      [2]float64{1.5, 3.0},
			DurabilityRange:  [2]int{50, 80},
		},
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponDagger,
			NamePrefixes:     []string{"Sharp", "Quick", "Silent", "Poison", "Shadow"},
			NameSuffixes:     []string{"Dagger", "Knife", "Stiletto", "Dirk"},
			Tags:             []string{"fast", "stealth", "light"},
			DamageRange:      [2]int{4, 8},
			AttackSpeedRange: [2]float64{1.5, 2.0},
			ValueRange:       [2]int{30, 150},
			WeightRange:      [2]float64{0.5, 1.5},
			DurabilityRange:  [2]int{40, 70},
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
		},
	}
}

// GetFantasyArmorTemplates returns armor templates for fantasy genre.
func GetFantasyArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:        TypeArmor,
			ArmorType:       ArmorChest,
			NamePrefixes:    []string{"Leather", "Chain", "Plate", "Scale", "Dragon"},
			NameSuffixes:    []string{"Armor", "Cuirass", "Breastplate", "Mail"},
			Tags:            []string{"protection", "heavy"},
			DefenseRange:    [2]int{10, 30},
			ValueRange:      [2]int{100, 400},
			WeightRange:     [2]float64{8.0, 20.0},
			DurabilityRange: [2]int{120, 200},
		},
		{
			BaseType:        TypeArmor,
			ArmorType:       ArmorHelmet,
			NamePrefixes:    []string{"Iron", "Steel", "Knight's", "Great", "Horned"},
			NameSuffixes:    []string{"Helmet", "Helm", "Crown", "Cap"},
			Tags:            []string{"protection", "head"},
			DefenseRange:    [2]int{5, 15},
			ValueRange:      [2]int{50, 200},
			WeightRange:     [2]float64{2.0, 5.0},
			DurabilityRange: [2]int{80, 120},
		},
		{
			BaseType:        TypeArmor,
			ArmorType:       ArmorShield,
			NamePrefixes:    []string{"Wooden", "Iron", "Steel", "Tower", "Kite"},
			NameSuffixes:    []string{"Shield", "Buckler", "Guard"},
			Tags:            []string{"block", "defense"},
			DefenseRange:    [2]int{8, 20},
			ValueRange:      [2]int{40, 180},
			WeightRange:     [2]float64{4.0, 12.0},
			DurabilityRange: [2]int{100, 150},
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
			BaseType:       TypeConsumable,
			ConsumableType: ConsumableScroll,
			NamePrefixes:   []string{"Scroll of", "Ancient", "Mystic"},
			NameSuffixes:   []string{"Fireball", "Lightning", "Ice", "Protection"},
			Tags:           []string{"magical", "spell", "consumable"},
			ValueRange:     [2]int{20, 150},
			WeightRange:    [2]float64{0.1, 0.2},
		},
	}
}

// GetSciFiWeaponTemplates returns weapon templates for sci-fi genre.
func GetSciFiWeaponTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:         TypeWeapon,
			WeaponType:       WeaponSword, // Using as energy blade
			NamePrefixes:     []string{"Plasma", "Energy", "Photon", "Quantum", "Nano"},
			NameSuffixes:     []string{"Blade", "Saber", "Cutter", "Sword"},
			Tags:             []string{"energy", "melee"},
			DamageRange:      [2]int{10, 18},
			AttackSpeedRange: [2]float64{1.2, 1.5},
			ValueRange:       [2]int{150, 500},
			WeightRange:      [2]float64{1.0, 2.0},
			DurabilityRange:  [2]int{200, 300},
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
		},
	}
}

// GetSciFiArmorTemplates returns armor templates for sci-fi genre.
func GetSciFiArmorTemplates() []ItemTemplate {
	return []ItemTemplate{
		{
			BaseType:        TypeArmor,
			ArmorType:       ArmorChest,
			NamePrefixes:    []string{"Combat", "Battle", "Tactical", "Power", "Nano"},
			NameSuffixes:    []string{"Suit", "Armor", "Exosuit", "Vest"},
			Tags:            []string{"powered", "armored"},
			DefenseRange:    [2]int{15, 35},
			ValueRange:      [2]int{300, 800},
			WeightRange:     [2]float64{5.0, 15.0},
			DurabilityRange: [2]int{200, 350},
		},
		{
			BaseType:        TypeArmor,
			ArmorType:       ArmorHelmet,
			NamePrefixes:    []string{"Combat", "Battle", "Tactical", "HUD", "Neural"},
			NameSuffixes:    []string{"Helmet", "Visor", "Interface"},
			Tags:            []string{"hud", "scanning"},
			DefenseRange:    [2]int{8, 18},
			ValueRange:      [2]int{150, 400},
			WeightRange:     [2]float64{1.0, 3.0},
			DurabilityRange: [2]int{150, 250},
		},
	}
}
