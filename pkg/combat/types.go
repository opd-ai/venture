// types.go defines core combat data structures and constructors.
// This file contains the Damage and Stats types along with their
// associated constructor functions. These types are used throughout
// the combat system for damage calculation and entity statistics.
//
// Package combat provides combat mechanics including damage calculation,
// status effects, combat AI, and battle resolution.
package combat

// Damage represents a damage calculation.
// Originally from: interfaces.go
type Damage struct {
	// Amount of damage
	Amount float64

	// Type of damage
	Type DamageType

	// Source entity ID
	SourceID uint64

	// Target entity ID
	TargetID uint64
}

// Stats represents character/enemy statistics.
// Originally from: interfaces.go
type Stats struct {
	// Health points
	HP    float64
	MaxHP float64

	// Mana/energy for abilities
	Mana    float64
	MaxMana float64

	// Offensive stats
	Attack     float64
	MagicPower float64
	CritChance float64
	CritDamage float64

	// Defensive stats
	Defense      float64
	MagicDefense float64
	Evasion      float64

	// Movement
	Speed float64

	// Resistances (0.0 = no resistance, 1.0 = immune)
	Resistances map[DamageType]float64
}

// NewStats creates a new Stats struct with default values.
// Originally from: interfaces.go
func NewStats() *Stats {
	return &Stats{
		HP:          100,
		MaxHP:       100,
		Mana:        50,
		MaxMana:     50,
		Attack:      10,
		Defense:     5,
		Speed:       100,
		Resistances: make(map[DamageType]float64),
	}
}
