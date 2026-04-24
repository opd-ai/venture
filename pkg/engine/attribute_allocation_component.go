// Package engine provides the attribute allocation component for character customization.
// This component tracks core attribute points (Strength, Agility, Intelligence, Vitality)
// that players can allocate to customize their character builds.
package engine

import (
	"encoding/json"
	"fmt"
)

// CoreAttribute represents the primary attributes that affect combat and gameplay.
type CoreAttribute int

const (
	// AttrStrength increases physical attack damage and carrying capacity.
	AttrStrength CoreAttribute = iota
	// AttrAgility increases attack speed, evasion, and movement speed.
	AttrAgility
	// AttrIntelligence increases magic damage, mana pool, and spell effects.
	AttrIntelligence
	// AttrVitality increases max health, health regeneration, and status resistance.
	AttrVitality
	// AttrEndurance increases stamina, block chance, and physical defense.
	AttrEndurance
	// AttrLuck increases critical chance, drop rates, and rare encounters.
	AttrLuck
	// NumCoreAttributes is the total count of core attributes.
	NumCoreAttributes
)

// Secondary-bonus key offsets for AppliedBonuses.
// CoreAttribute values run from 0–5 (NumCoreAttributes–1). Two reserved offset
// ranges are used to store secondary bonuses from the same attribute without
// colliding with primary bonus keys or each other.
// Range [100, 199]: secondary effect #1 per attribute (e.g., INT→MaxMana).
// Range [1000, 1999]: secondary effect #2 per attribute (e.g., STR→CarryCap).
const (
	// attrSecondaryOffset is the base offset for the first secondary bonus key.
	// Already used by CoreAttribute(100+AttrIntelligence) → mana max bonus.
	attrSecondaryOffset CoreAttribute = 100
	// attrTertiaryOffset is the base offset for the second secondary bonus key
	// introduced by G26 (CarryCapPerStr, SpeedBonusPerAgi, ManaRegenPerInt,
	// HealthRegenPerVit). Using 1000 avoids collision with the [100,199] range.
	attrTertiaryOffset CoreAttribute = 1000
)

// String returns the display name for a core attribute.
func (a CoreAttribute) String() string {
	switch a {
	case AttrStrength:
		return "Strength"
	case AttrAgility:
		return "Agility"
	case AttrIntelligence:
		return "Intelligence"
	case AttrVitality:
		return "Vitality"
	case AttrEndurance:
		return "Endurance"
	case AttrLuck:
		return "Luck"
	default:
		return "Unknown"
	}
}

// ShortName returns a 3-letter abbreviation for the attribute.
func (a CoreAttribute) ShortName() string {
	switch a {
	case AttrStrength:
		return "STR"
	case AttrAgility:
		return "AGI"
	case AttrIntelligence:
		return "INT"
	case AttrVitality:
		return "VIT"
	case AttrEndurance:
		return "END"
	case AttrLuck:
		return "LCK"
	default:
		return "UNK"
	}
}

// AttributeFromString converts a string to a CoreAttribute.
func AttributeFromString(s string) (CoreAttribute, error) {
	switch s {
	case "Strength", "STR", "strength", "str":
		return AttrStrength, nil
	case "Agility", "AGI", "agility", "agi":
		return AttrAgility, nil
	case "Intelligence", "INT", "intelligence", "int":
		return AttrIntelligence, nil
	case "Vitality", "VIT", "vitality", "vit":
		return AttrVitality, nil
	case "Endurance", "END", "endurance", "end":
		return AttrEndurance, nil
	case "Luck", "LCK", "luck", "lck":
		return AttrLuck, nil
	default:
		return 0, fmt.Errorf("unknown attribute: %s", s)
	}
}

// AttributeEffects defines how each attribute affects gameplay stats.
// All values are per-point bonuses.
type AttributeEffects struct {
	// Strength effects
	AttackBonusPerStr float64 // Physical damage bonus per STR point
	CarryCapPerStr    float64 // Inventory weight capacity per STR point

	// Agility effects
	SpeedBonusPerAgi   float64 // Movement speed % per AGI point
	EvasionBonusPerAgi float64 // Evasion chance % per AGI point
	AttackSpeedPerAgi  float64 // Attack speed % per AGI point

	// Intelligence effects
	MagicBonusPerInt float64 // Magic damage bonus per INT point
	MaxManaPerInt    float64 // Max mana per INT point
	ManaRegenPerInt  float64 // Mana regen per second per INT point

	// Vitality effects
	MaxHealthPerVit    float64 // Max health per VIT point
	HealthRegenPerVit  float64 // Health regen per second per VIT point
	StatusResistPerVit float64 // Status effect resistance % per VIT point

	// Endurance effects
	DefenseBonusPerEnd float64 // Physical defense per END point
	BlockChancePerEnd  float64 // Block chance % per END point
	StaminaPerEnd      float64 // Max stamina per END point

	// Luck effects
	CritChancePerLuck   float64 // Critical hit chance % per LCK point
	DropRateBonusPerLck float64 // Item drop rate bonus % per LCK point
	RareEncounterPerLck float64 // Rare encounter chance % per LCK point
}

// DefaultAttributeEffects returns balanced attribute scaling values.
func DefaultAttributeEffects() *AttributeEffects {
	return &AttributeEffects{
		// Strength: 2.0 attack per point
		AttackBonusPerStr: 2.0,
		CarryCapPerStr:    5.0,

		// Agility: subtle but impactful
		SpeedBonusPerAgi:   0.5, // 0.5% speed per point
		EvasionBonusPerAgi: 0.3, // 0.3% evasion per point
		AttackSpeedPerAgi:  0.5, // 0.5% attack speed per point

		// Intelligence: magic-focused
		MagicBonusPerInt: 2.5,
		MaxManaPerInt:    10.0,
		ManaRegenPerInt:  0.2,

		// Vitality: survivability
		MaxHealthPerVit:    15.0,
		HealthRegenPerVit:  0.3,
		StatusResistPerVit: 0.5,

		// Endurance: defense
		DefenseBonusPerEnd: 1.5,
		BlockChancePerEnd:  0.3,
		StaminaPerEnd:      8.0,

		// Luck: fortune
		CritChancePerLuck:   0.4, // 0.4% crit per point
		DropRateBonusPerLck: 1.0, // 1% drop rate per point
		RareEncounterPerLck: 0.2, // 0.2% rare encounter per point
	}
}

// AttributeAllocationComponent stores allocated attribute points.
// Pure data component - all logic lives in AttributeAllocationSystem.
type AttributeAllocationComponent struct {
	// BaseAttributes stores the base (unmodified) value for each attribute.
	// All characters start with 10 in each attribute.
	BaseAttributes [NumCoreAttributes]int `json:"base_attributes"`

	// AllocatedPoints stores player-allocated points for each attribute.
	AllocatedPoints [NumCoreAttributes]int `json:"allocated_points"`

	// BonusPoints stores temporary/equipment-based bonus points.
	BonusPoints [NumCoreAttributes]int `json:"bonus_points"`

	// UnspentPoints is the number of attribute points available to allocate.
	UnspentPoints int `json:"unspent_points"`

	// TotalPointsEarned tracks lifetime attribute points for achievements.
	TotalPointsEarned int `json:"total_points_earned"`

	// PointsPerLevel defines how many attribute points are gained per level.
	PointsPerLevel int `json:"points_per_level"`

	// RespecCount tracks number of times the player has reset attributes.
	RespecCount int `json:"respec_count"`

	// LastModifiedTime tracks when attributes were last changed.
	LastModifiedTime float64 `json:"last_modified_time"`

	// Dirty flag indicates if derived stats need recalculation.
	Dirty bool `json:"-"`

	// AppliedBonuses tracks which bonuses have been applied to stats.
	AppliedBonuses map[CoreAttribute]float64 `json:"-"`
}

// Type returns the component type identifier.
func (c *AttributeAllocationComponent) Type() string {
	return "attribute_allocation"
}

// NewAttributeAllocationComponent creates a new component with default base values.
func NewAttributeAllocationComponent() *AttributeAllocationComponent {
	comp := &AttributeAllocationComponent{
		PointsPerLevel: 3, // 3 attribute points per level
		AppliedBonuses: make(map[CoreAttribute]float64),
	}

	// Initialize base attributes to 10 each
	for i := 0; i < int(NumCoreAttributes); i++ {
		comp.BaseAttributes[i] = 10
	}

	return comp
}

// GetTotal returns the total value for an attribute (base + allocated + bonus).
func (c *AttributeAllocationComponent) GetTotal(attr CoreAttribute) int {
	if attr < 0 || attr >= NumCoreAttributes {
		return 0
	}
	return c.BaseAttributes[attr] + c.AllocatedPoints[attr] + c.BonusPoints[attr]
}

// GetAllocated returns only the player-allocated points for an attribute.
func (c *AttributeAllocationComponent) GetAllocated(attr CoreAttribute) int {
	if attr < 0 || attr >= NumCoreAttributes {
		return 0
	}
	return c.AllocatedPoints[attr]
}

// GetBase returns the base value for an attribute.
func (c *AttributeAllocationComponent) GetBase(attr CoreAttribute) int {
	if attr < 0 || attr >= NumCoreAttributes {
		return 0
	}
	return c.BaseAttributes[attr]
}

// GetBonus returns the bonus points for an attribute.
func (c *AttributeAllocationComponent) GetBonus(attr CoreAttribute) int {
	if attr < 0 || attr >= NumCoreAttributes {
		return 0
	}
	return c.BonusPoints[attr]
}

// CanAllocate checks if points can be allocated to an attribute.
func (c *AttributeAllocationComponent) CanAllocate(attr CoreAttribute, points int) bool {
	if attr < 0 || attr >= NumCoreAttributes {
		return false
	}
	if points <= 0 {
		return false
	}
	return c.UnspentPoints >= points
}

// TotalAllocatedPoints returns the sum of all allocated points.
func (c *AttributeAllocationComponent) TotalAllocatedPoints() int {
	total := 0
	for i := 0; i < int(NumCoreAttributes); i++ {
		total += c.AllocatedPoints[i]
	}
	return total
}

// TotalEffectivePoints returns the sum of all total attribute values.
func (c *AttributeAllocationComponent) TotalEffectivePoints() int {
	total := 0
	for i := 0; i < int(NumCoreAttributes); i++ {
		total += c.GetTotal(CoreAttribute(i))
	}
	return total
}

// GetAttributeSummary returns a formatted string of all attributes.
func (c *AttributeAllocationComponent) GetAttributeSummary() string {
	return fmt.Sprintf("STR:%d AGI:%d INT:%d VIT:%d END:%d LCK:%d (Unspent:%d)",
		c.GetTotal(AttrStrength),
		c.GetTotal(AttrAgility),
		c.GetTotal(AttrIntelligence),
		c.GetTotal(AttrVitality),
		c.GetTotal(AttrEndurance),
		c.GetTotal(AttrLuck),
		c.UnspentPoints)
}

// Serialize encodes the component for persistence.
func (c *AttributeAllocationComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes the component from persistence data.
func (c *AttributeAllocationComponent) Deserialize(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	// Ensure AppliedBonuses is initialized
	if c.AppliedBonuses == nil {
		c.AppliedBonuses = make(map[CoreAttribute]float64)
	}
	// Mark dirty to recalculate bonuses after load
	c.Dirty = true
	return nil
}

// Copy creates a deep copy of the component.
func (c *AttributeAllocationComponent) Copy() *AttributeAllocationComponent {
	newComp := &AttributeAllocationComponent{
		UnspentPoints:     c.UnspentPoints,
		TotalPointsEarned: c.TotalPointsEarned,
		PointsPerLevel:    c.PointsPerLevel,
		RespecCount:       c.RespecCount,
		LastModifiedTime:  c.LastModifiedTime,
		Dirty:             true,
		AppliedBonuses:    make(map[CoreAttribute]float64),
	}
	copy(newComp.BaseAttributes[:], c.BaseAttributes[:])
	copy(newComp.AllocatedPoints[:], c.AllocatedPoints[:])
	copy(newComp.BonusPoints[:], c.BonusPoints[:])
	return newComp
}
