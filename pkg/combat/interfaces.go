// Package combat provides combat system interfaces and types.
// This file defines the damage calculation interface and core combat
// types used by the combat system.
package combat

// DamageType represents different types of damage.
type DamageType int

// Damage type constants.
const (
	DamagePhysical DamageType = iota
	DamageMagical
	DamageFire
	DamageIce
	DamageLightning
	DamagePoison
)

// Damage represents a damage calculation.
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

// CombatResolver handles combat calculations.
type CombatResolver interface {
	// CalculateDamage computes final damage after resistances and defenses
	CalculateDamage(damage Damage, targetStats *Stats) float64

	// ResolveCombat handles a combat interaction between entities
	ResolveCombat(attackerID, defenderID uint64) []Damage
}
