// Package engine provides vehicle combat component functionality.
// This file implements VehicleCombatComponent which manages vehicle-based
// combat including ramming damage and mounted weapon systems.
//
// Phase 21.2: Advanced Vehicle Features
package engine

import "github.com/opd-ai/venture/pkg/combat"

// VehicleCombatComponent stores combat-related stats for vehicles.
// This component enables vehicles to deal damage through ramming
// and mounted weapon systems.
type VehicleCombatComponent struct {
	// RammingDamage is base damage dealt when colliding with entities at speed
	// Scales with vehicle speed at impact
	RammingDamage float64

	// RammingCooldown is time until next ramming attack can occur (seconds)
	RammingCooldown float64

	// CurrentRammingCooldown tracks time remaining until next ram (seconds)
	CurrentRammingCooldown float64

	// MinRammingSpeed is minimum speed required to deal ramming damage
	MinRammingSpeed float64

	// WeaponMounted indicates if vehicle has a mounted weapon system
	WeaponMounted bool

	// WeaponDamage is damage dealt by mounted weapon per shot
	WeaponDamage float64

	// WeaponRange is effective range of mounted weapon (pixels)
	WeaponRange float64

	// WeaponCooldown is time between weapon attacks (seconds)
	WeaponCooldown float64

	// CurrentWeaponCooldown tracks time until next weapon shot (seconds)
	CurrentWeaponCooldown float64

	// WeaponType indicates type of mounted weapon (affects visuals/behavior)
	WeaponType string // "Cannon", "MachineGun", "Laser", "Magic", "Ballista"

	// WeaponProjectileSpeed is speed of projectiles fired (pixels per second)
	WeaponProjectileSpeed float64

	// DamageType for mounted weapon attacks
	WeaponDamageType combat.DamageType

	// ArmorRating reduces incoming damage to vehicle durability
	ArmorRating float64
}

// Type returns the component type identifier.
func (v *VehicleCombatComponent) Type() string {
	return "vehicle_combat"
}

// NewVehicleCombatComponent creates a combat component with default values.
func NewVehicleCombatComponent() *VehicleCombatComponent {
	return &VehicleCombatComponent{
		RammingDamage:          20.0,
		RammingCooldown:        1.0,
		CurrentRammingCooldown: 0.0,
		MinRammingSpeed:        50.0,
		WeaponMounted:          false,
		WeaponDamage:           15.0,
		WeaponRange:            200.0,
		WeaponCooldown:         0.5,
		CurrentWeaponCooldown:  0.0,
		WeaponType:             "Cannon",
		WeaponProjectileSpeed:  300.0, // Default projectile speed (pixels/second)
		WeaponDamageType:       combat.DamagePhysical,
		ArmorRating:            5.0,
	}
}

// CanRam checks if ramming attack is ready and vehicle has sufficient speed.
func (v *VehicleCombatComponent) CanRam(currentSpeed float64) bool {
	return v.CurrentRammingCooldown <= 0 && currentSpeed >= v.MinRammingSpeed
}

// CanShoot checks if mounted weapon can fire.
func (v *VehicleCombatComponent) CanShoot() bool {
	return v.WeaponMounted && v.CurrentWeaponCooldown <= 0
}

// UpdateCooldowns decreases cooldown timers.
func (v *VehicleCombatComponent) UpdateCooldowns(deltaTime float64) {
	if v.CurrentRammingCooldown > 0 {
		v.CurrentRammingCooldown -= deltaTime
		if v.CurrentRammingCooldown < 0 {
			v.CurrentRammingCooldown = 0
		}
	}

	if v.CurrentWeaponCooldown > 0 {
		v.CurrentWeaponCooldown -= deltaTime
		if v.CurrentWeaponCooldown < 0 {
			v.CurrentWeaponCooldown = 0
		}
	}
}

// ExecuteRam triggers ramming attack and sets cooldown.
func (v *VehicleCombatComponent) ExecuteRam() {
	v.CurrentRammingCooldown = v.RammingCooldown
}

// ExecuteShot triggers weapon shot and sets cooldown.
func (v *VehicleCombatComponent) ExecuteShot() {
	v.CurrentWeaponCooldown = v.WeaponCooldown
}

// CalculateRammingDamage returns damage based on current speed.
// Damage scales linearly with speed above minimum threshold.
func (v *VehicleCombatComponent) CalculateRammingDamage(currentSpeed float64) float64 {
	if currentSpeed < v.MinRammingSpeed {
		return 0.0
	}
	// Scale damage by speed ratio (speed / minSpeed)
	// At exactly min speed: base damage
	// At 2x min speed: 2x base damage
	speedRatio := currentSpeed / v.MinRammingSpeed
	return v.RammingDamage * speedRatio
}

// CargoComponent stores item inventory for vehicles.
// Phase 21.2: Cargo/Passenger System
type CargoComponent struct {
	// Items stored in cargo hold
	Items []uint64 // Entity IDs of stored item entities

	// MaxItems is cargo capacity
	MaxItems int

	// WeightCapacity is maximum weight vehicle can carry (kg)
	WeightCapacity float64

	// CurrentWeight is total weight of stored items (kg)
	CurrentWeight float64
}

// Type returns the component type identifier.
func (c *CargoComponent) Type() string {
	return "cargo"
}

// NewCargoComponent creates a cargo component with specified capacity.
func NewCargoComponent(maxItems int, weightCapacity float64) *CargoComponent {
	return &CargoComponent{
		Items:          make([]uint64, 0, maxItems),
		MaxItems:       maxItems,
		WeightCapacity: weightCapacity,
		CurrentWeight:  0.0,
	}
}

// CanAddItem checks if item can be added to cargo.
func (c *CargoComponent) CanAddItem(weight float64) bool {
	return len(c.Items) < c.MaxItems && (c.CurrentWeight+weight) <= c.WeightCapacity
}

// AddItem adds an item to cargo and updates weight.
func (c *CargoComponent) AddItem(itemID uint64, weight float64) bool {
	if !c.CanAddItem(weight) {
		return false
	}
	c.Items = append(c.Items, itemID)
	c.CurrentWeight += weight
	return true
}

// RemoveItem removes an item from cargo and updates weight.
func (c *CargoComponent) RemoveItem(itemID uint64, weight float64) bool {
	for i, id := range c.Items {
		if id == itemID {
			// Remove item by swapping with last and truncating
			c.Items[i] = c.Items[len(c.Items)-1]
			c.Items = c.Items[:len(c.Items)-1]
			c.CurrentWeight -= weight
			if c.CurrentWeight < 0 {
				c.CurrentWeight = 0
			}
			return true
		}
	}
	return false
}

// IsFull checks if cargo is at capacity.
func (c *CargoComponent) IsFull() bool {
	return len(c.Items) >= c.MaxItems
}

// IsEmpty checks if cargo has no items.
func (c *CargoComponent) IsEmpty() bool {
	return len(c.Items) == 0
}

// GetItemCount returns number of items in cargo.
func (c *CargoComponent) GetItemCount() int {
	return len(c.Items)
}

// UpgradeSlotComponent manages vehicle upgrade modifications.
// Phase 21.2: Upgrade System
type UpgradeSlotComponent struct {
	// Slots stores installed upgrades
	Slots []*VehicleUpgrade

	// MaxSlots is number of upgrade slots available
	MaxSlots int
}

// Type returns the component type identifier.
func (u *UpgradeSlotComponent) Type() string {
	return "upgrade_slots"
}

// VehicleUpgrade represents a single upgrade modification.
type VehicleUpgrade struct {
	// Name of the upgrade
	Name string

	// Type indicates what stat is modified
	Type UpgradeType

	// Value is the modification amount (additive or multiplicative)
	Value float64

	// IsMultiplicative indicates if value is multiplier (true) or additive (false)
	IsMultiplicative bool
}

// UpgradeType identifies which vehicle stat is modified.
type UpgradeType int

const (
	// UpgradeSpeed modifies MaxSpeed
	UpgradeSpeed UpgradeType = iota
	// UpgradeAcceleration modifies Acceleration
	UpgradeAcceleration
	// UpgradeHandling modifies Handling
	UpgradeHandling
	// UpgradeDurability modifies MaxDurability
	UpgradeDurability
	// UpgradeArmor modifies ArmorRating
	UpgradeArmor
	// UpgradeCapacity modifies passenger/cargo Capacity
	UpgradeCapacity
	// UpgradeFuelCapacity modifies FuelCapacity
	UpgradeFuelCapacity
	// UpgradeWeaponDamage modifies WeaponDamage
	UpgradeWeaponDamage
)

// String returns string representation of upgrade type.
func (u UpgradeType) String() string {
	switch u {
	case UpgradeSpeed:
		return "Speed"
	case UpgradeAcceleration:
		return "Acceleration"
	case UpgradeHandling:
		return "Handling"
	case UpgradeDurability:
		return "Durability"
	case UpgradeArmor:
		return "Armor"
	case UpgradeCapacity:
		return "Capacity"
	case UpgradeFuelCapacity:
		return "FuelCapacity"
	case UpgradeWeaponDamage:
		return "WeaponDamage"
	default:
		return "Unknown"
	}
}

// NewUpgradeSlotComponent creates an upgrade slot component.
func NewUpgradeSlotComponent(maxSlots int) *UpgradeSlotComponent {
	return &UpgradeSlotComponent{
		Slots:    make([]*VehicleUpgrade, 0, maxSlots),
		MaxSlots: maxSlots,
	}
}

// CanAddUpgrade checks if upgrade slot is available.
func (u *UpgradeSlotComponent) CanAddUpgrade() bool {
	return len(u.Slots) < u.MaxSlots
}

// AddUpgrade installs an upgrade in an empty slot.
func (u *UpgradeSlotComponent) AddUpgrade(upgrade *VehicleUpgrade) bool {
	if !u.CanAddUpgrade() {
		return false
	}
	u.Slots = append(u.Slots, upgrade)
	return true
}

// RemoveUpgrade removes an upgrade by name.
func (u *UpgradeSlotComponent) RemoveUpgrade(name string) bool {
	for i, upgrade := range u.Slots {
		if upgrade.Name == name {
			// Remove upgrade by swapping with last and truncating
			u.Slots[i] = u.Slots[len(u.Slots)-1]
			u.Slots = u.Slots[:len(u.Slots)-1]
			return true
		}
	}
	return false
}

// GetUpgradeCount returns number of installed upgrades.
func (u *UpgradeSlotComponent) GetUpgradeCount() int {
	return len(u.Slots)
}

// HasUpgrade checks if upgrade with given name is installed.
func (u *UpgradeSlotComponent) HasUpgrade(name string) bool {
	for _, upgrade := range u.Slots {
		if upgrade.Name == name {
			return true
		}
	}
	return false
}
