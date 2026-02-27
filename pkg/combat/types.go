// types.go defines core combat data structures and constructors.
// This file contains the Damage and Stats types along with their
// associated constructor functions. These types are used throughout
// the combat system for damage calculation and entity statistics.
package combat

import (
	"errors"
	"fmt"
)

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

// NewStats creates a new Stats struct with balanced default values suitable for
// a fresh level-1 entity: 100 HP, 50 Mana, 10 Attack, 5 Defense, 100 Speed,
// and an empty resistance map. Callers should adjust fields for specific entity types.
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

// Common validation errors for combat types.
var (
	// ErrNegativeDamage indicates damage amount is negative.
	ErrNegativeDamage = errors.New("damage amount cannot be negative")

	// ErrInvalidDamageType indicates damage type is not recognized.
	ErrInvalidDamageType = errors.New("invalid damage type")

	// ErrNegativeHP indicates HP is negative.
	ErrNegativeHP = errors.New("HP cannot be negative")

	// ErrNegativeMaxHP indicates MaxHP is negative or zero.
	ErrNegativeMaxHP = errors.New("MaxHP must be positive")

	// ErrHPExceedsMax indicates HP exceeds MaxHP.
	ErrHPExceedsMax = errors.New("HP cannot exceed MaxHP")

	// ErrNegativeMana indicates Mana is negative.
	ErrNegativeMana = errors.New("Mana cannot be negative")

	// ErrNegativeMaxMana indicates MaxMana is negative.
	ErrNegativeMaxMana = errors.New("MaxMana cannot be negative")

	// ErrManaExceedsMax indicates Mana exceeds MaxMana.
	ErrManaExceedsMax = errors.New("Mana cannot exceed MaxMana")

	// ErrNegativeAttack indicates Attack is negative.
	ErrNegativeAttack = errors.New("Attack cannot be negative")

	// ErrNegativeMagicPower indicates MagicPower is negative.
	ErrNegativeMagicPower = errors.New("MagicPower cannot be negative")

	// ErrNegativeDefense indicates Defense is negative.
	ErrNegativeDefense = errors.New("Defense cannot be negative")

	// ErrNegativeMagicDefense indicates MagicDefense is negative.
	ErrNegativeMagicDefense = errors.New("MagicDefense cannot be negative")

	// ErrNegativeSpeed indicates Speed is negative.
	ErrNegativeSpeed = errors.New("Speed cannot be negative")

	// ErrInvalidCritChance indicates CritChance is outside valid range.
	ErrInvalidCritChance = errors.New("CritChance must be between 0.0 and 1.0")

	// ErrInvalidCritDamage indicates CritDamage is below minimum.
	ErrInvalidCritDamage = errors.New("CritDamage must be at least 1.0")

	// ErrInvalidEvasion indicates Evasion is outside valid range.
	ErrInvalidEvasion = errors.New("Evasion must be between 0.0 and 1.0")

	// ErrInvalidResistance indicates a resistance value is outside valid range.
	// Note: Range is -0.5 to 1.0 to match the runtime clamping in CalculateDamage.
	// -0.5 represents maximum weakness (150% damage), 1.0 represents immunity.
	ErrInvalidResistance = errors.New("resistance must be between -0.5 and 1.0")
)

// Validate checks if the Damage struct contains valid values.
// Returns nil if valid, or an error describing the validation failure.
//
// Validation rules:
//   - Amount must be non-negative (zero is allowed for no-damage effects)
//   - Type must be a recognized DamageType constant
func (d *Damage) Validate() error {
	if d.Amount < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeDamage, d.Amount)
	}

	// Validate damage type is within known range
	if d.Type < DamagePhysical || d.Type > DamagePoison {
		return fmt.Errorf("%w: got %d", ErrInvalidDamageType, d.Type)
	}

	return nil
}

// Validate checks if the Stats struct contains valid values.
// Returns nil if valid, or an error describing the validation failure.
//
// Validation rules:
//   - HP must be non-negative
//   - MaxHP must be positive (> 0)
//   - HP must not exceed MaxHP
//   - Mana must be non-negative
//   - MaxMana must be non-negative
//   - Mana must not exceed MaxMana (if MaxMana > 0)
//   - Attack, Defense, MagicPower, MagicDefense, Speed must be non-negative
//   - CritChance must be between 0.0 and 1.0
//   - CritDamage must be at least 1.0 (if set above zero)
//   - Evasion must be between 0.0 and 1.0
//   - All resistances must be between -0.5 (max weakness) and 1.0 (immunity)
func (s *Stats) Validate() error {
	if err := s.validateHealth(); err != nil {
		return err
	}
	if err := s.validateMana(); err != nil {
		return err
	}
	if err := s.validateOffensiveStats(); err != nil {
		return err
	}
	if err := s.validateDefensiveStats(); err != nil {
		return err
	}
	if err := s.validateCriticalStats(); err != nil {
		return err
	}
	if err := s.validateResistances(); err != nil {
		return err
	}
	return nil
}

// validateHealth checks if HP and MaxHP are valid.
// Returns nil if validation passes, or an error describing the validation failure.
// This is a complete implementation following Go's error handling pattern where
// nil indicates success.
func (s *Stats) validateHealth() error {
	if s.HP < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeHP, s.HP)
	}
	if s.MaxHP <= 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeMaxHP, s.MaxHP)
	}
	if s.HP > s.MaxHP {
		return fmt.Errorf("%w: HP=%f, MaxHP=%f", ErrHPExceedsMax, s.HP, s.MaxHP)
	}
	return nil
}

// validateMana checks if Mana and MaxMana are valid.
// Returns nil if validation passes, or an error describing the validation failure.
// This is a complete implementation following Go's error handling pattern where
// nil indicates success.
func (s *Stats) validateMana() error {
	if s.Mana < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeMana, s.Mana)
	}
	if s.MaxMana < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeMaxMana, s.MaxMana)
	}
	if s.MaxMana > 0 && s.Mana > s.MaxMana {
		return fmt.Errorf("%w: Mana=%f, MaxMana=%f", ErrManaExceedsMax, s.Mana, s.MaxMana)
	}
	return nil
}

// validateOffensiveStats checks if Attack and MagicPower are valid.
// Returns nil if validation passes, or an error describing the validation failure.
// This is a complete implementation following Go's error handling pattern where
// nil indicates success.
func (s *Stats) validateOffensiveStats() error {
	if s.Attack < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeAttack, s.Attack)
	}
	if s.MagicPower < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeMagicPower, s.MagicPower)
	}
	if s.Speed < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeSpeed, s.Speed)
	}
	return nil
}

// validateDefensiveStats checks if Defense, MagicDefense, and Evasion are valid.
// Returns nil if validation passes, or an error describing the validation failure.
// This is a complete implementation following Go's error handling pattern where
// nil indicates success.
func (s *Stats) validateDefensiveStats() error {
	if s.Defense < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeDefense, s.Defense)
	}
	if s.MagicDefense < 0 {
		return fmt.Errorf("%w: got %f", ErrNegativeMagicDefense, s.MagicDefense)
	}
	if s.Evasion < 0 || s.Evasion > 1.0 {
		return fmt.Errorf("%w: got %f", ErrInvalidEvasion, s.Evasion)
	}
	return nil
}

// validateCriticalStats checks if CritChance and CritDamage are valid.
// Returns nil if validation passes, or an error describing the validation failure.
// This is a complete implementation following Go's error handling pattern where
// nil indicates success.
func (s *Stats) validateCriticalStats() error {
	if s.CritChance < 0 || s.CritChance > 1.0 {
		return fmt.Errorf("%w: got %f", ErrInvalidCritChance, s.CritChance)
	}
	if s.CritDamage != 0 && s.CritDamage < 1.0 {
		return fmt.Errorf("%w: got %f", ErrInvalidCritDamage, s.CritDamage)
	}
	return nil
}

// validateResistances checks if all resistance values are within valid range.
// Returns nil if validation passes, or an error describing the validation failure.
// Range is -0.5 to 1.0 to match the runtime clamping in CalculateDamage.
// -0.5 represents maximum weakness (150% damage), 1.0 represents immunity.
func (s *Stats) validateResistances() error {
	for damageType, resistance := range s.Resistances {
		if resistance < -0.5 || resistance > 1.0 {
			return fmt.Errorf("%w: %v resistance is %f", ErrInvalidResistance, damageType, resistance)
		}
	}
	return nil
}

// IsDead returns true if the entity has zero or negative HP.
func (s *Stats) IsDead() bool {
	return s.HP <= 0
}

// GetResistance safely returns the resistance value for a given damage type.
// Returns 0.0 (no resistance) if the damage type is not in the resistances map
// or if the resistances map is nil.
func (s *Stats) GetResistance(damageType DamageType) float64 {
	if s.Resistances == nil {
		return 0.0
	}
	return s.Resistances[damageType]
}

// ApplyDamage safely applies damage to HP, respecting the 0 floor.
// Returns the actual damage dealt (may be less if HP was already low).
func (s *Stats) ApplyDamage(amount float64) float64 {
	if amount < 0 {
		return 0
	}
	actual := amount
	if s.HP < amount {
		actual = s.HP
	}
	s.HP -= actual
	if s.HP < 0 {
		s.HP = 0
	}
	return actual
}

// ApplyHealing safely applies healing to HP, respecting the MaxHP ceiling.
// Returns the actual healing applied (may be less if HP was already near max).
func (s *Stats) ApplyHealing(amount float64) float64 {
	if amount < 0 {
		return 0
	}
	oldHP := s.HP
	s.HP += amount
	if s.HP > s.MaxHP {
		s.HP = s.MaxHP
	}
	return s.HP - oldHP
}
