// Package advanced provides multi-classing, prestige classes, and talent tree systems.
package advanced

import "image/color"

// ClassDefinition describes a base class
type ClassDefinition struct {
	ID          ClassID
	Name        string
	Description string
	BaseStats   StatBonuses
	Color       color.RGBA
}

// PrestigeClassDefinition describes a prestige class
type PrestigeClassDefinition struct {
	ID           PrestigeClassID
	Name         string
	Description  string
	Requirements PrestigeRequirements
	BaseStats    StatBonuses
	Color        color.RGBA
}

// PrestigeRequirements defines what's needed to unlock a prestige class
type PrestigeRequirements struct {
	MinLevel          int
	RequiredPrimary   []ClassID
	RequiredSecondary []ClassID
	MinPrimaryStat    int
	MinSecondaryStat  int
}

// StatBonuses represents stat bonuses from classes/talents
type StatBonuses struct {
	Health       int
	Mana         int
	Stamina      int
	Strength     int
	Dexterity    int
	Intelligence int
	Wisdom       int
	Charisma     int
	Defense      int
	MagicDefense int
	CritChance   float64
	CritDamage   float64
	Speed        float64
}

// Add combines two StatBonuses
func (s StatBonuses) Add(other StatBonuses) StatBonuses {
	return StatBonuses{
		Health:       s.Health + other.Health,
		Mana:         s.Mana + other.Mana,
		Stamina:      s.Stamina + other.Stamina,
		Strength:     s.Strength + other.Strength,
		Dexterity:    s.Dexterity + other.Dexterity,
		Intelligence: s.Intelligence + other.Intelligence,
		Wisdom:       s.Wisdom + other.Wisdom,
		Charisma:     s.Charisma + other.Charisma,
		Defense:      s.Defense + other.Defense,
		MagicDefense: s.MagicDefense + other.MagicDefense,
		CritChance:   s.CritChance + other.CritChance,
		CritDamage:   s.CritDamage + other.CritDamage,
		Speed:        s.Speed + other.Speed,
	}
}

// Scale multiplies all bonuses by a factor
func (s StatBonuses) Scale(factor float64) StatBonuses {
	return StatBonuses{
		Health:       int(float64(s.Health) * factor),
		Mana:         int(float64(s.Mana) * factor),
		Stamina:      int(float64(s.Stamina) * factor),
		Strength:     int(float64(s.Strength) * factor),
		Dexterity:    int(float64(s.Dexterity) * factor),
		Intelligence: int(float64(s.Intelligence) * factor),
		Wisdom:       int(float64(s.Wisdom) * factor),
		Charisma:     int(float64(s.Charisma) * factor),
		Defense:      int(float64(s.Defense) * factor),
		MagicDefense: int(float64(s.MagicDefense) * factor),
		CritChance:   s.CritChance * factor,
		CritDamage:   s.CritDamage * factor,
		Speed:        s.Speed * factor,
	}
}

// TalentDefinition describes a talent
type TalentDefinition struct {
	ID            TalentID
	Name          string
	Description   string
	Category      TalentCategory
	MaxRank       int
	Prerequisites []TalentID
	Bonuses       StatBonuses
}

// TalentTree contains 30 talents organized in 3 categories
type TalentTree struct {
	Name      string
	ClassID   ClassID
	Offensive []TalentDefinition
	Defensive []TalentDefinition
	Utility   []TalentDefinition
}

// TalentAllocation tracks which talents have been unlocked and their ranks
type TalentAllocation struct {
	Talents     map[TalentID]int // talent ID -> current rank
	PointsSpent int
	PointsTotal int
}

// AdvancedClassComponent is an ECS component for advanced class features
type AdvancedClassComponent struct {
	PrimaryClass   ClassID
	SecondaryClass ClassID
	PrestigeClass  PrestigeClassID
	TalentPoints   TalentAllocation
	Level          int
	RespecCount    int
}

// Type returns the component type identifier
func (a AdvancedClassComponent) Type() string {
	return "advanced_class"
}

// SynergyBonus represents bonuses from compatible multi-class combinations
type SynergyBonus struct {
	Primary   ClassID
	Secondary ClassID
	Name      string
	Bonuses   StatBonuses
}

// RespecCost calculates the gold cost to reset talents
type RespecCost struct {
	BaseGold  int
	PerRespec int
	MaxCost   int
}
