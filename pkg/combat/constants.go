// constants.go defines damage type enumeration and related constants.
// This file contains all constant definitions for the combat package,
// separating them from type and interface definitions for better organization.
//
// Package combat provides combat mechanics including damage calculation,
// status effects, combat AI, and battle resolution.
package combat

// DamageType represents different types of damage.
// Originally from: interfaces.go
type DamageType int

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
