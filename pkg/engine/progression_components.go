// Package engine provides character progression components.
// This file defines components for experience, leveling, and skill progression
// used by the progression system.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/skills"
)

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

// SkillTreeComponent stores the player's skill tree and learned skills.
type SkillTreeComponent struct {
	Tree           *skills.SkillTree      // The skill tree structure
	LearnedSkills  map[string]bool        // Skill IDs that have been learned
	SkillLevels    map[string]int         // Current level of each learned skill
	Attributes     map[string]int         // Attributes for prerequisites (e.g., strength, intelligence)
	TotalPointsUsed int                   // Total skill points spent
}

// Type returns the component type identifier.
func (s *SkillTreeComponent) Type() string {
	return "skill_tree"
}

// NewSkillTreeComponent creates a new skill tree component with the given tree.
func NewSkillTreeComponent(tree *skills.SkillTree) *SkillTreeComponent {
	return &SkillTreeComponent{
		Tree:            tree,
		LearnedSkills:   make(map[string]bool),
		SkillLevels:     make(map[string]int),
		Attributes:      make(map[string]int),
		TotalPointsUsed: 0,
	}
}

// LearnSkill marks a skill as learned and increments its level.
// Returns true if successful, false if already at max level or not unlocked.
func (s *SkillTreeComponent) LearnSkill(skillID string, availablePoints int) bool {
	skill := s.Tree.GetSkillByID(skillID)
	if skill == nil {
		return false
	}

	// Check if skill can be learned (prerequisites met)
	currentLevel := s.SkillLevels[skillID]
	if currentLevel >= skill.MaxLevel {
		return false // Already at max level
	}

	// For first level, check if prerequisites are met
	if currentLevel == 0 {
		// Get player level from context (simplified here, would normally come from ExperienceComponent)
		playerLevel := 1 // Simplified
		if !skill.IsUnlocked(playerLevel, availablePoints, s.LearnedSkills, s.Attributes) {
			return false
		}
	}

	// Check if enough skill points available
	if availablePoints < skill.Requirements.SkillPoints {
		return false
	}

	// Learn/level up the skill
	s.SkillLevels[skillID]++
	s.LearnedSkills[skillID] = true
	s.TotalPointsUsed += skill.Requirements.SkillPoints
	skill.Level = s.SkillLevels[skillID]

	return true
}

// UnlearnSkill removes a skill (for respec).
// Returns the skill points refunded.
func (s *SkillTreeComponent) UnlearnSkill(skillID string) int {
	skill := s.Tree.GetSkillByID(skillID)
	if skill == nil {
		return 0
	}

	currentLevel := s.SkillLevels[skillID]
	if currentLevel == 0 {
		return 0 // Not learned
	}

	// Check if any other skills depend on this one
	for learnedID := range s.LearnedSkills {
		learnedSkill := s.Tree.GetSkillByID(learnedID)
		if learnedSkill != nil {
			for _, prereqID := range learnedSkill.Requirements.PrerequisiteIDs {
				if prereqID == skillID {
					return 0 // Cannot unlearn if another skill depends on it
				}
			}
		}
	}

	// Refund points
	pointsRefunded := skill.Requirements.SkillPoints * currentLevel
	s.TotalPointsUsed -= pointsRefunded

	// Remove skill
	delete(s.SkillLevels, skillID)
	delete(s.LearnedSkills, skillID)
	skill.Level = 0

	return pointsRefunded
}

// GetSkillLevel returns the current level of a skill.
func (s *SkillTreeComponent) GetSkillLevel(skillID string) int {
	return s.SkillLevels[skillID]
}

// IsSkillLearned checks if a skill has been learned (level > 0).
func (s *SkillTreeComponent) IsSkillLearned(skillID string) bool {
	return s.LearnedSkills[skillID]
}

// GetAvailableSkills returns all skills that can currently be learned.
func (s *SkillTreeComponent) GetAvailableSkills(playerLevel, availablePoints int) []*skills.Skill {
	available := make([]*skills.Skill, 0)
	for _, node := range s.Tree.Nodes {
		skill := node.Skill
		// Skip if already at max level
		if s.SkillLevels[skill.ID] >= skill.MaxLevel {
			continue
		}
		// Check if unlocked
		if skill.IsUnlocked(playerLevel, availablePoints, s.LearnedSkills, s.Attributes) {
			available = append(available, skill)
		}
	}
	return available
}
