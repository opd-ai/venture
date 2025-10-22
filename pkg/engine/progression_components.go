package engine

import "fmt"

// ExperienceComponent tracks an entity's experience points and level.
// Experience is gained through combat and completing objectives.
// When enough XP is accumulated, the entity levels up.
type ExperienceComponent struct {
	Level       int // Current character level (starts at 1)
	CurrentXP   int // Current experience points
	RequiredXP  int // XP needed for next level
	TotalXP     int // Total XP earned across all levels
	SkillPoints int // Unspent skill points for skill trees
}

// Type returns the component type identifier.
func (e ExperienceComponent) Type() string {
	return "experience"
}

// NewExperienceComponent creates a new experience component at level 1.
func NewExperienceComponent() *ExperienceComponent {
	return &ExperienceComponent{
		Level:       1,
		CurrentXP:   0,
		RequiredXP:  100, // Base XP for level 2
		TotalXP:     0,
		SkillPoints: 0,
	}
}

// AddXP adds experience points and returns true if a level up occurred.
func (e *ExperienceComponent) AddXP(xp int) bool {
	if xp <= 0 {
		return false
	}

	e.CurrentXP += xp
	e.TotalXP += xp

	if e.CurrentXP >= e.RequiredXP {
		return true
	}
	return false
}

// CanLevelUp checks if the entity has enough XP to level up.
func (e *ExperienceComponent) CanLevelUp() bool {
	return e.CurrentXP >= e.RequiredXP
}

// ProgressToNextLevel returns the progress as a percentage (0.0 to 1.0).
func (e *ExperienceComponent) ProgressToNextLevel() float64 {
	if e.RequiredXP <= 0 {
		return 1.0
	}
	progress := float64(e.CurrentXP) / float64(e.RequiredXP)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// String returns a string representation of the component.
func (e *ExperienceComponent) String() string {
	return fmt.Sprintf("Level %d: %d/%d XP (%.1f%%) - %d skill points",
		e.Level, e.CurrentXP, e.RequiredXP,
		e.ProgressToNextLevel()*100, e.SkillPoints)
}

// LevelScalingComponent defines how an entity's stats scale with level.
// This is used to automatically increase stats when leveling up.
type LevelScalingComponent struct {
	HealthPerLevel       float64 // Health increase per level
	AttackPerLevel       float64 // Attack increase per level
	DefensePerLevel      float64 // Defense increase per level
	MagicPowerPerLevel   float64 // Magic power increase per level
	MagicDefensePerLevel float64 // Magic defense increase per level
	BaseHealth           float64 // Starting health at level 1
	BaseAttack           float64 // Starting attack at level 1
	BaseDefense          float64 // Starting defense at level 1
	BaseMagicPower       float64 // Starting magic power at level 1
	BaseMagicDefense     float64 // Starting magic defense at level 1
}

// Type returns the component type identifier.
func (l LevelScalingComponent) Type() string {
	return "level_scaling"
}

// NewLevelScalingComponent creates a default level scaling component.
// These values are balanced for a standard combat-focused character.
func NewLevelScalingComponent() *LevelScalingComponent {
	return &LevelScalingComponent{
		HealthPerLevel:       10.0,  // +10 HP per level
		AttackPerLevel:       2.0,   // +2 attack per level
		DefensePerLevel:      1.5,   // +1.5 defense per level
		MagicPowerPerLevel:   2.0,   // +2 magic power per level
		MagicDefensePerLevel: 1.5,   // +1.5 magic defense per level
		BaseHealth:           100.0, // Starting health
		BaseAttack:           10.0,  // Starting attack
		BaseDefense:          5.0,   // Starting defense
		BaseMagicPower:       10.0,  // Starting magic power
		BaseMagicDefense:     5.0,   // Starting magic defense
	}
}

// CalculateStatForLevel calculates what a stat should be at a given level.
func (l *LevelScalingComponent) CalculateStatForLevel(baseStat, perLevel float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	// Stats at level = base + (perLevel * (level - 1))
	return baseStat + (perLevel * float64(level-1))
}

// CalculateHealthForLevel returns the health value for a given level.
func (l *LevelScalingComponent) CalculateHealthForLevel(level int) float64 {
	return l.CalculateStatForLevel(l.BaseHealth, l.HealthPerLevel, level)
}

// CalculateAttackForLevel returns the attack value for a given level.
func (l *LevelScalingComponent) CalculateAttackForLevel(level int) float64 {
	return l.CalculateStatForLevel(l.BaseAttack, l.AttackPerLevel, level)
}

// CalculateDefenseForLevel returns the defense value for a given level.
func (l *LevelScalingComponent) CalculateDefenseForLevel(level int) float64 {
	return l.CalculateStatForLevel(l.BaseDefense, l.DefensePerLevel, level)
}

// CalculateMagicPowerForLevel returns the magic power value for a given level.
func (l *LevelScalingComponent) CalculateMagicPowerForLevel(level int) float64 {
	return l.CalculateStatForLevel(l.BaseMagicPower, l.MagicPowerPerLevel, level)
}

// CalculateMagicDefenseForLevel returns the magic defense value for a given level.
func (l *LevelScalingComponent) CalculateMagicDefenseForLevel(level int) float64 {
	return l.CalculateStatForLevel(l.BaseMagicDefense, l.MagicDefensePerLevel, level)
}

// String returns a string representation of the component.
func (l *LevelScalingComponent) String() string {
	return fmt.Sprintf("Scaling: HP+%.1f ATK+%.1f DEF+%.1f MAG+%.1f MDEF+%.1f",
		l.HealthPerLevel, l.AttackPerLevel, l.DefensePerLevel,
		l.MagicPowerPerLevel, l.MagicDefensePerLevel)
}
