// Package vehicle provides procedural vehicle generation types.
// This file defines vehicle data structures and rarity levels.
//
// Phase 21.1: Vehicle Foundation
package vehicle

import "github.com/opd-ai/venture/pkg/engine"

// VehicleType represents the category of vehicle.
type VehicleType int

const (
	// TypeMount is a living creature mount
	TypeMount VehicleType = iota
	// TypeCart is a wheeled vehicle
	TypeCart
	// TypeBoat is a water vehicle
	TypeBoat
	// TypeGlider is an air vehicle
	TypeGlider
	// TypeMech is a mechanical suit
	TypeMech
)

// String returns the string representation of a vehicle type.
func (v VehicleType) String() string {
	switch v {
	case TypeMount:
		return "Mount"
	case TypeCart:
		return "Cart"
	case TypeBoat:
		return "Boat"
	case TypeGlider:
		return "Glider"
	case TypeMech:
		return "Mech"
	default:
		return "Unknown"
	}
}

// Rarity represents vehicle rarity tiers.
type Rarity int

const (
	// RarityCommon is the most common tier
	RarityCommon Rarity = iota
	// RarityUncommon is a moderately rare tier
	RarityUncommon
	// RarityRare is a rare tier
	RarityRare
	// RarityEpic is a very rare tier
	RarityEpic
	// RarityLegendary is the rarest tier
	RarityLegendary
)

// String returns the string representation of rarity.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// GetMultiplier returns the stat multiplier for this rarity.
func (r Rarity) GetMultiplier() float64 {
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

// Vehicle represents a generated vehicle instance.
type Vehicle struct {
	// Name is the procedurally generated vehicle name
	Name string

	// Description is a flavor text description
	Description string

	// VehicleType identifies the category
	VehicleType VehicleType

	// Rarity affects stats and appearance
	Rarity Rarity

	// Stats define vehicle capabilities
	MaxSpeed      float64
	Acceleration  float64
	Handling      float64
	MaxDurability float64
	FuelCapacity  float64
	Capacity      int

	// FuelType identifies resource consumed
	FuelType string

	// GenreID tracks the genre this vehicle was generated for
	GenreID string

	// Depth is the dungeon level where this vehicle was generated
	Depth int

	// Color represents the primary vehicle color (0xRRGGBB)
	Color uint32

	// Phase 21.2: Advanced Vehicle Features

	// HasCombat indicates if vehicle has combat capabilities
	HasCombat bool

	// HasWeapon indicates if vehicle has mounted weapon
	HasWeapon bool

	// WeaponType is the mounted weapon system type
	WeaponType string

	// CargoSlots is the number of item slots
	CargoSlots int

	// CargoWeight is the maximum cargo weight capacity (kg)
	CargoWeight float64

	// UpgradeSlots is the number of upgrade slots
	UpgradeSlots int

	// SpecialAbility is a genre-specific special ability (epic+ only)
	SpecialAbility string

	// Phase 21.3: Visual Variation

	// Decorations are visual ornaments applied to the vehicle
	Decorations []string

	// DamageState represents visual wear level (0.0 = pristine, 1.0 = heavily damaged)
	// Affects visual appearance but not stats
	DamageState float64

	// SecondaryColor for two-tone paint schemes (0xRRGGBB)
	SecondaryColor uint32

	// DecalPattern identifies paint/decal pattern ("stripes", "flames", "camo", etc.)
	DecalPattern string
}

// ToComponent converts this vehicle to an engine VehicleComponent.
// Phase 21.2: Returns vehicle component only. Use ToComponents() for all components.
func (v *Vehicle) ToComponent() *engine.VehicleComponent {
	// Map VehicleType to engine.VehicleType
	var engineType engine.VehicleType
	switch v.VehicleType {
	case TypeMount:
		engineType = engine.VehicleMount
	case TypeCart:
		engineType = engine.VehicleCart
	case TypeBoat:
		engineType = engine.VehicleBoat
	case TypeGlider:
		engineType = engine.VehicleGlider
	case TypeMech:
		engineType = engine.VehicleMech
	}

	comp := engine.NewVehicleComponent(engineType)

	// Override with generated stats
	comp.MaxSpeed = v.MaxSpeed
	comp.Acceleration = v.Acceleration
	comp.Handling = v.Handling
	comp.MaxDurability = v.MaxDurability
	comp.Durability = v.MaxDurability // Start at full health
	comp.FuelCapacity = v.FuelCapacity
	comp.FuelAmount = v.FuelCapacity // Start with full fuel
	comp.FuelType = v.FuelType
	comp.Capacity = v.Capacity

	return comp
}

// ToComponents converts this vehicle to all relevant engine components.
// Phase 21.2: Returns vehicle, combat, cargo, and upgrade components.
func (v *Vehicle) ToComponents() []engine.Component {
	components := []engine.Component{
		v.ToComponent(),
	}

	// Add combat component if vehicle has combat capabilities
	if v.HasCombat {
		combat := engine.NewVehicleCombatComponent()
		combat.WeaponMounted = v.HasWeapon
		combat.WeaponType = v.WeaponType

		// Scale combat stats based on vehicle stats
		combat.RammingDamage = v.MaxSpeed * 0.2   // 20% of max speed as base ramming damage
		combat.MinRammingSpeed = v.MaxSpeed * 0.3 // Need 30% of max speed to ram

		if v.HasWeapon {
			combat.WeaponDamage = 15.0 * v.Rarity.GetMultiplier()
			combat.WeaponRange = 200.0
		}

		// Scale armor with durability
		combat.ArmorRating = v.MaxDurability * 0.05

		components = append(components, combat)
	}

	// Add cargo component
	cargo := engine.NewCargoComponent(v.CargoSlots, v.CargoWeight)
	components = append(components, cargo)

	// Add upgrade slots component
	upgrades := engine.NewUpgradeSlotComponent(v.UpgradeSlots)
	components = append(components, upgrades)

	return components
}

// VehicleTemplate defines base parameters for vehicle generation.
type VehicleTemplate struct {
	// NamePrefix is used in name generation
	NamePrefix string

	// NameSuffix is used in name generation
	NameSuffix string

	// VehicleType identifies the category
	VehicleType VehicleType

	// BaseStats define the starting stats before scaling
	BaseMaxSpeed     float64
	BaseAcceleration float64
	BaseHandling     float64
	BaseDurability   float64
	BaseFuelCapacity float64
	BaseCapacity     int

	// FuelType identifies resource consumed
	FuelType string

	// BaseColor is the default color (0xRRGGBB)
	BaseColor uint32

	// Description template
	DescriptionTemplate string
}
