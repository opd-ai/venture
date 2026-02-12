// constants.go defines damage type enumeration and related constants.
// This file contains all constant definitions for the combat package,
// separating them from type and interface definitions for better organization.
package combat

// DamageType represents different types of damage.
// Originally from: interfaces.go
type DamageType int

// String returns the human-readable name of the damage type.
// Useful for logging, debugging, and UI display.
func (d DamageType) String() string {
	switch d {
	case DamagePhysical:
		return "Physical"
	case DamageMagical:
		return "Magical"
	case DamageFire:
		return "Fire"
	case DamageIce:
		return "Ice"
	case DamageLightning:
		return "Lightning"
	case DamagePoison:
		return "Poison"
	default:
		return "Unknown"
	}
}

// Damage type constants.
// Originally defined in: interfaces.go
const (
	DamagePhysical DamageType = iota
	DamageMagical
	DamageFire
	DamageIce
	DamageLightning
	DamagePoison
)
