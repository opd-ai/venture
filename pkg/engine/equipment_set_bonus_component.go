// Package engine provides equipment set bonus tracking.
// This file defines components for tracking active equipment set bonuses.
package engine

// SetBonusTier represents the stat bonuses granted at a specific piece threshold.
type SetBonusTier struct {
	// PiecesRequired is the number of pieces needed to activate this tier
	PiecesRequired int
	// DamageBonus is flat damage increase
	DamageBonus int
	// DefenseBonus is flat defense increase
	DefenseBonus int
	// AttackSpeedBonus is percentage attack speed increase (0.1 = 10%)
	AttackSpeedBonus float64
	// CriticalChanceBonus is percentage crit chance increase (0.05 = 5%)
	CriticalChanceBonus float64
	// MovementSpeedBonus is percentage movement speed increase (0.1 = 10%)
	MovementSpeedBonus float64
	// HealthBonus is flat max health increase
	HealthBonus int
	// ManaRegenBonus is percentage mana regen increase
	ManaRegenBonus float64
	// SpecialEffect describes any unique bonus (for UI display)
	SpecialEffect string
}

// EquipmentSetDefinition defines an equipment set and its tiered bonuses.
type EquipmentSetDefinition struct {
	// SetID is the unique identifier for this set
	SetID string
	// SetName is the display name
	SetName string
	// Description provides flavor text for the set
	Description string
	// TotalPieces is the maximum number of pieces in the set
	TotalPieces int
	// Tiers contains the bonuses at each piece threshold
	Tiers []SetBonusTier
	// GenreID indicates which genre this set is themed for
	GenreID string
}

// ActiveSetBonus tracks currently active bonuses from a single set.
type ActiveSetBonus struct {
	// SetID identifies the active set
	SetID string
	// SetName for display
	SetName string
	// PiecesEquipped is the count of equipped pieces
	PiecesEquipped int
	// TotalPieces in the full set
	TotalPieces int
	// ActiveTiers lists the tier indices currently active
	ActiveTiers []int
	// CombinedBonus holds the sum of all active tier bonuses
	CombinedBonus SetBonusTier
}

// EquipmentSetBonusComponent tracks all active set bonuses for an entity.
// Component: pure data, no logic.
type EquipmentSetBonusComponent struct {
	// ActiveSets maps SetID to the active bonus state
	ActiveSets map[string]*ActiveSetBonus
	// Dirty indicates bonuses need recalculation
	Dirty bool
	// LastEquipmentHash stores a hash to detect equipment changes
	LastEquipmentHash uint64
}

// Type returns the component type identifier.
func (e *EquipmentSetBonusComponent) Type() string {
	return "equipment_set_bonus"
}

// NewEquipmentSetBonusComponent creates a new equipment set bonus component.
func NewEquipmentSetBonusComponent() *EquipmentSetBonusComponent {
	return &EquipmentSetBonusComponent{
		ActiveSets: make(map[string]*ActiveSetBonus),
		Dirty:      true,
	}
}

// GetTotalDamageBonus returns the sum of damage bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalDamageBonus() int {
	total := 0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.DamageBonus
	}
	return total
}

// GetTotalDefenseBonus returns the sum of defense bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalDefenseBonus() int {
	total := 0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.DefenseBonus
	}
	return total
}

// GetTotalAttackSpeedBonus returns the sum of attack speed bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalAttackSpeedBonus() float64 {
	total := 0.0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.AttackSpeedBonus
	}
	return total
}

// GetTotalCriticalChanceBonus returns the sum of critical chance bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalCriticalChanceBonus() float64 {
	total := 0.0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.CriticalChanceBonus
	}
	return total
}

// GetTotalMovementSpeedBonus returns the sum of movement speed bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalMovementSpeedBonus() float64 {
	total := 0.0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.MovementSpeedBonus
	}
	return total
}

// GetTotalHealthBonus returns the sum of health bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalHealthBonus() int {
	total := 0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.HealthBonus
	}
	return total
}

// GetTotalManaRegenBonus returns the sum of mana regen bonuses from all active sets.
func (e *EquipmentSetBonusComponent) GetTotalManaRegenBonus() float64 {
	total := 0.0
	for _, set := range e.ActiveSets {
		total += set.CombinedBonus.ManaRegenBonus
	}
	return total
}

// HasActiveSet returns true if the specified set has any active bonus.
func (e *EquipmentSetBonusComponent) HasActiveSet(setID string) bool {
	set, exists := e.ActiveSets[setID]
	return exists && set.PiecesEquipped > 0
}

// GetActiveTierCount returns the number of active bonus tiers for a set.
func (e *EquipmentSetBonusComponent) GetActiveTierCount(setID string) int {
	set, exists := e.ActiveSets[setID]
	if !exists {
		return 0
	}
	return len(set.ActiveTiers)
}
