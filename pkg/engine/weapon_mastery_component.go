// Package engine provides the weapon mastery component for tracking proficiency.
// This component stores weapon mastery levels and XP for each weapon type.
package engine

import (
	"encoding/json"
)

// WeaponMasteryLevel represents the mastery tier for a weapon type.
type WeaponMasteryLevel int

const (
	// MasteryNovice is the starting level (0-99 XP)
	MasteryNovice WeaponMasteryLevel = iota
	// MasteryApprentice unlocks at 100 XP
	MasteryApprentice
	// MasteryJourneyman unlocks at 300 XP
	MasteryJourneyman
	// MasteryExpert unlocks at 600 XP
	MasteryExpert
	// MasteryMaster unlocks at 1000 XP
	MasteryMaster
	// MasteryGrandmaster unlocks at 1500 XP (max level)
	MasteryGrandmaster
)

// String returns the display name for a mastery level.
func (m WeaponMasteryLevel) String() string {
	switch m {
	case MasteryNovice:
		return "Novice"
	case MasteryApprentice:
		return "Apprentice"
	case MasteryJourneyman:
		return "Journeyman"
	case MasteryExpert:
		return "Expert"
	case MasteryMaster:
		return "Master"
	case MasteryGrandmaster:
		return "Grandmaster"
	default:
		return "Unknown"
	}
}

// XPThreshold returns the XP required to reach this mastery level.
func (m WeaponMasteryLevel) XPThreshold() int {
	switch m {
	case MasteryNovice:
		return 0
	case MasteryApprentice:
		return 100
	case MasteryJourneyman:
		return 300
	case MasteryExpert:
		return 600
	case MasteryMaster:
		return 1000
	case MasteryGrandmaster:
		return 1500
	default:
		return 0
	}
}

// WeaponMasteryData holds mastery data for a single weapon type.
type WeaponMasteryData struct {
	// XP is the current experience points for this weapon type
	XP int
	// Level is the current mastery level
	Level WeaponMasteryLevel
	// TotalKills tracks enemies killed with this weapon type
	TotalKills int
	// TotalDamage tracks total damage dealt with this weapon type
	TotalDamage float64
	// CriticalHits tracks critical hits landed with this weapon type
	CriticalHits int
}

// WeaponMasteryComponent tracks weapon proficiency for all weapon types.
// Pure data component - all logic lives in WeaponMasterySystem.
type WeaponMasteryComponent struct {
	// Mastery stores mastery data keyed by weapon type string
	Mastery map[string]*WeaponMasteryData

	// BonusesApplied tracks which bonuses have been applied to avoid double-application
	BonusesApplied map[string]WeaponMasteryLevel

	// Dirty flag indicates if bonuses need recalculation
	Dirty bool

	// TotalMasteryXP is the sum of all weapon mastery XP (for achievements)
	TotalMasteryXP int
}

// Type returns the component type identifier.
func (c *WeaponMasteryComponent) Type() string {
	return "weapon_mastery"
}

// NewWeaponMasteryComponent creates a new weapon mastery component with initialized maps.
func NewWeaponMasteryComponent() *WeaponMasteryComponent {
	return &WeaponMasteryComponent{
		Mastery:        make(map[string]*WeaponMasteryData),
		BonusesApplied: make(map[string]WeaponMasteryLevel),
		Dirty:          false,
		TotalMasteryXP: 0,
	}
}

// GetMastery returns mastery data for a weapon type, creating if not exists.
func (c *WeaponMasteryComponent) GetMastery(weaponType string) *WeaponMasteryData {
	if c.Mastery == nil {
		c.Mastery = make(map[string]*WeaponMasteryData)
	}
	if data, exists := c.Mastery[weaponType]; exists {
		return data
	}
	data := &WeaponMasteryData{
		XP:           0,
		Level:        MasteryNovice,
		TotalKills:   0,
		TotalDamage:  0,
		CriticalHits: 0,
	}
	c.Mastery[weaponType] = data
	return data
}

// GetMasteryLevel returns the mastery level for a weapon type.
func (c *WeaponMasteryComponent) GetMasteryLevel(weaponType string) WeaponMasteryLevel {
	return c.GetMastery(weaponType).Level
}

// GetMasteryXP returns the current XP for a weapon type.
func (c *WeaponMasteryComponent) GetMasteryXP(weaponType string) int {
	return c.GetMastery(weaponType).XP
}

// GetXPToNextLevel returns XP needed for next mastery level (0 if max).
func (c *WeaponMasteryComponent) GetXPToNextLevel(weaponType string) int {
	data := c.GetMastery(weaponType)
	nextLevel := data.Level + 1
	if nextLevel > MasteryGrandmaster {
		return 0 // Already at max
	}
	return nextLevel.XPThreshold() - data.XP
}

// GetProgressToNextLevel returns 0.0-1.0 progress to next mastery level.
func (c *WeaponMasteryComponent) GetProgressToNextLevel(weaponType string) float64 {
	data := c.GetMastery(weaponType)
	if data.Level >= MasteryGrandmaster {
		return 1.0 // Max level
	}

	currentThreshold := data.Level.XPThreshold()
	nextThreshold := (data.Level + 1).XPThreshold()
	xpRange := nextThreshold - currentThreshold
	if xpRange <= 0 {
		return 1.0
	}

	xpProgress := data.XP - currentThreshold
	return float64(xpProgress) / float64(xpRange)
}

// Serialize encodes the component for persistence.
func (c *WeaponMasteryComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes the component from persistence data.
func (c *WeaponMasteryComponent) Deserialize(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	// Ensure maps are initialized
	if c.Mastery == nil {
		c.Mastery = make(map[string]*WeaponMasteryData)
	}
	if c.BonusesApplied == nil {
		c.BonusesApplied = make(map[string]WeaponMasteryLevel)
	}
	// Mark dirty to recalculate bonuses after load
	c.Dirty = true
	return nil
}

// GetHighestMasteryLevel returns the highest mastery level achieved across all weapons.
func (c *WeaponMasteryComponent) GetHighestMasteryLevel() WeaponMasteryLevel {
	highest := MasteryNovice
	for _, data := range c.Mastery {
		if data.Level > highest {
			highest = data.Level
		}
	}
	return highest
}

// GetMasteredWeaponCount returns count of weapons at or above specified level.
func (c *WeaponMasteryComponent) GetMasteredWeaponCount(minLevel WeaponMasteryLevel) int {
	count := 0
	for _, data := range c.Mastery {
		if data.Level >= minLevel {
			count++
		}
	}
	return count
}
