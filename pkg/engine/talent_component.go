// Package engine provides the talent component for tracking talent point allocation.
// Talents are passive bonuses unlocked by spending talent points earned on level-up.
package engine

import (
	"encoding/json"
)

// TalentCategory represents the category of a talent.
type TalentCategory int

const (
	// TalentCategoryOffense increases damage output.
	TalentCategoryOffense TalentCategory = iota
	// TalentCategoryDefense increases survivability.
	TalentCategoryDefense
	// TalentCategoryUtility improves resource management and mobility.
	TalentCategoryUtility
	// TalentCategoryMastery provides unique gameplay modifiers.
	TalentCategoryMastery
)

// String returns the display name for a talent category.
func (tc TalentCategory) String() string {
	switch tc {
	case TalentCategoryOffense:
		return "Offense"
	case TalentCategoryDefense:
		return "Defense"
	case TalentCategoryUtility:
		return "Utility"
	case TalentCategoryMastery:
		return "Mastery"
	default:
		return "Unknown"
	}
}

// TalentDefinition defines a talent that can be unlocked.
type TalentDefinition struct {
	// ID is the unique identifier for this talent.
	ID string
	// Name is the display name.
	Name string
	// Description explains the talent's effect.
	Description string
	// Category determines which tree the talent belongs to.
	Category TalentCategory
	// MaxRanks is the maximum points that can be spent (1-5).
	MaxRanks int
	// RequiredLevel is the minimum character level to unlock.
	RequiredLevel int
	// PrerequisiteTalentID is the ID of the talent that must be fully ranked first (empty if none).
	PrerequisiteTalentID string
	// PrerequisitePointsInCategory is the minimum points spent in this category first.
	PrerequisitePointsInCategory int
	// BonusPerRank defines the stat modification per rank.
	BonusPerRank TalentBonus
}

// TalentBonus defines the stat modifications for a talent rank.
type TalentBonus struct {
	// Flat bonuses (added directly)
	FlatHealth       float64
	FlatMana         float64
	FlatDamage       float64
	FlatDefense      float64
	FlatMagicPower   float64
	FlatMagicDefense float64
	FlatSpeed        float64

	// Percentage bonuses (multiplied, 0.05 = +5%)
	HealthPercent        float64
	ManaPercent          float64
	DamagePercent        float64
	DefensePercent       float64
	MagicPowerPercent    float64
	MagicDefensePercent  float64
	SpeedPercent         float64
	CritChanceBonus      float64
	CritDamageBonus      float64
	LifestealPercent     float64
	CooldownReduction    float64
	ManaCostReduction    float64
	XPBonusPercent       float64
	GoldBonusPercent     float64
	DodgeChanceBonus     float64
	BlockChanceBonus     float64
	HealingReceivedBonus float64
	StatusResistBonus    float64
}

// TalentAllocation tracks points spent in a single talent.
type TalentAllocation struct {
	TalentID     string
	CurrentRanks int
}

// TalentComponent tracks an entity's talent points and allocations.
type TalentComponent struct {
	// UnspentPoints is the number of talent points available to spend.
	UnspentPoints int
	// TotalPointsEarned is the total talent points earned over character lifetime.
	TotalPointsEarned int
	// Allocations maps talent IDs to points spent.
	Allocations map[string]int
	// PointsInCategory tracks points spent per category for prerequisites.
	PointsInCategory map[TalentCategory]int
	// Dirty indicates bonuses need recalculation.
	Dirty bool
	// CachedBonuses holds the calculated total bonuses from all talents.
	CachedBonuses TalentBonus
	// AppliedBonuses records the last bonus set written to the entity's stats.
	// It must be subtracted before applying new bonuses to prevent unbounded
	// stat growth on reset / reallocation (G23 fix).
	AppliedBonuses TalentBonus
}

// Type returns the component type identifier.
func (t *TalentComponent) Type() string {
	return "talent"
}

// NewTalentComponent creates a new talent component with no points.
func NewTalentComponent() *TalentComponent {
	return &TalentComponent{
		UnspentPoints:     0,
		TotalPointsEarned: 0,
		Allocations:       make(map[string]int),
		PointsInCategory:  make(map[TalentCategory]int),
		Dirty:             true,
		CachedBonuses:     TalentBonus{},
	}
}

// AddTalentPoints grants talent points to the entity.
func (t *TalentComponent) AddTalentPoints(points int) {
	if points <= 0 {
		return
	}
	t.UnspentPoints += points
	t.TotalPointsEarned += points
}

// GetRanks returns the current ranks in a talent.
func (t *TalentComponent) GetRanks(talentID string) int {
	return t.Allocations[talentID]
}

// CanAllocate checks if a talent can receive another point.
func (t *TalentComponent) CanAllocate(talent *TalentDefinition, characterLevel int) bool {
	if t.UnspentPoints <= 0 {
		return false
	}
	if characterLevel < talent.RequiredLevel {
		return false
	}
	currentRanks := t.Allocations[talent.ID]
	if currentRanks >= talent.MaxRanks {
		return false
	}
	if talent.PrerequisiteTalentID != "" {
		prereqRanks := t.Allocations[talent.PrerequisiteTalentID]
		// Prerequisite must be fully ranked
		prereqDef := GetTalentDefinition(talent.PrerequisiteTalentID)
		if prereqDef == nil || prereqRanks < prereqDef.MaxRanks {
			return false
		}
	}
	if talent.PrerequisitePointsInCategory > 0 {
		pointsInCat := t.PointsInCategory[talent.Category]
		if pointsInCat < talent.PrerequisitePointsInCategory {
			return false
		}
	}
	return true
}

// AllocatePoint spends a point in a talent. Returns true if successful.
func (t *TalentComponent) AllocatePoint(talent *TalentDefinition, characterLevel int) bool {
	if !t.CanAllocate(talent, characterLevel) {
		return false
	}
	t.UnspentPoints--
	t.Allocations[talent.ID]++
	t.PointsInCategory[talent.Category]++
	t.Dirty = true
	return true
}

// DeallocatePoint removes a point from a talent. Returns true if successful.
// This is used for respec functionality.
func (t *TalentComponent) DeallocatePoint(talent *TalentDefinition) bool {
	currentRanks := t.Allocations[talent.ID]
	if currentRanks <= 0 {
		return false
	}
	// Check if any talents depend on this one
	for _, def := range GetAllTalentDefinitions() {
		if def.PrerequisiteTalentID == talent.ID {
			if t.Allocations[def.ID] > 0 {
				// Cannot deallocate while dependent talent has points
				return false
			}
		}
	}
	t.Allocations[talent.ID]--
	t.PointsInCategory[talent.Category]--
	t.UnspentPoints++
	t.Dirty = true
	return true
}

// ResetAll removes all talent allocations, returning all points.
func (t *TalentComponent) ResetAll() {
	totalSpent := 0
	for _, ranks := range t.Allocations {
		totalSpent += ranks
	}
	t.UnspentPoints += totalSpent
	t.Allocations = make(map[string]int)
	t.PointsInCategory = make(map[TalentCategory]int)
	t.Dirty = true
}

// Serialize converts the talent component to bytes for persistence.
func (t *TalentComponent) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

// Deserialize restores the talent component from bytes.
func (t *TalentComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, t)
}
