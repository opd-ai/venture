// Package item provides item type definitions.
// This file defines core item data structures.
// Enums and constants relocated to: constants.go
package item

// Stats represents the core statistics of an item.
type Stats struct {
	// Damage for weapons
	Damage int
	// Defense for armor
	Defense int
	// Healing for consumables (e.g. health potions)
	Healing int
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
	// ClassRestrictions lists classes that can use this item (empty = any class)
	// Valid values: "warrior", "mage", "rogue", "ranger", "cleric", "necromancer"
	// Phase 25.2: Class-specific equipment restrictions
	ClassRestrictions []string
	// SpellEffectID identifies the spell effect triggered when consumable is used
	// For scrolls: "fireball", "lightning", "ice", "protection", "teleportation", etc.
	// Integration: Gap A2 - Consumable Spell Effect Activation
	SpellEffectID string
	// SpellDuration is the duration of the spell effect in seconds (0 = use default)
	// Integration: Gap A2 - Consumable Spell Effect Activation
	SpellDuration float64
	// SpellTargetType specifies targeting mode: "self", "entity", "area", "terrain"
	// Empty string means auto-detect based on SpellEffectID
	// Integration: Gap A2 - Consumable Spell Effect Activation
	SpellTargetType string
	// SpellRadius is the effect radius for area-targeting spells (0 = use default)
	// Integration: Gap A2 - Consumable Spell Effect Activation
	SpellRadius float64
	// SetID identifies which equipment set this item belongs to (empty = no set)
	// Items with matching SetID grant cumulative bonuses when multiple pieces are equipped
	SetID string
	// SetPieceIndex identifies which piece of the set this is (0-5 for 6-piece sets)
	SetPieceIndex int
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

// CanBeUsedByClass checks if an item can be used by a specific character class.
// Phase 25.2: Class-specific equipment restrictions.
// Returns true if the item has no restrictions or if the class is in the allowed list.
func (i *Item) CanBeUsedByClass(className string) bool {
	// No restrictions means any class can use it
	if len(i.ClassRestrictions) == 0 {
		return true
	}

	// Check if the class is in the allowed list
	for _, allowedClass := range i.ClassRestrictions {
		if allowedClass == className {
			return true
		}
	}

	return false
}
