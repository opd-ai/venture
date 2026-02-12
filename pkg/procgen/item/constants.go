// Package item provides item type definitions.
// This file defines item type constants and enumerations.
// Code relocated from: types.go
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

// Value returns the numeric value of a rarity level.
// Used for compatibility with systems that use float64 for rarity.
func (r Rarity) Value() float64 {
	switch r {
	case RarityCommon:
		return 1.0
	case RarityUncommon:
		return 1.2
	case RarityRare:
		return 1.5
	case RarityEpic:
		return 2.0
	case RarityLegendary:
		return 3.0
	default:
		return 1.0
	}
}
