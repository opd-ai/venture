// Package skills provides skill type definitions and helper functions.
// This file contains helper functions for skill operations that were
// extracted from methods on Skill and SkillTree types to maintain
// ECS-like pure data structures.
package skills

// IsSkillUnlocked checks if a skill can be unlocked given the current player state.
// This is the preferred way to check skill unlock status for ECS compliance.
//
// Parameters:
//   - skill: The skill to check
//   - playerLevel: Current player level
//   - skillPoints: Available skill points
//   - learnedSkills: Map of skill IDs that have been learned
//   - attributes: Player attributes (e.g., strength, intelligence)
//
// Returns true if all unlock requirements are met.
func IsSkillUnlocked(skill *Skill, playerLevel, skillPoints int, learnedSkills map[string]bool, attributes map[string]int) bool {
	if skill == nil {
		return false
	}

	// Check player level
	if playerLevel < skill.Requirements.PlayerLevel {
		return false
	}

	// Check skill points
	if skillPoints < skill.Requirements.SkillPoints {
		return false
	}

	// Check prerequisites
	for _, prereqID := range skill.Requirements.PrerequisiteIDs {
		if !learnedSkills[prereqID] {
			return false
		}
	}

	// Check attribute minimums
	for attr, minValue := range skill.Requirements.AttributeMinimums {
		if attributes[attr] < minValue {
			return false
		}
	}

	return true
}

// CanSkillLevelUp checks if a skill can be leveled up further.
// This is the preferred way to check level-up availability for ECS compliance.
//
// Parameters:
//   - skill: The skill to check
//
// Returns true if the skill is learned (Level > 0) and not at max level.
func CanSkillLevelUp(skill *Skill) bool {
	if skill == nil {
		return false
	}
	return skill.Level > 0 && skill.Level < skill.MaxLevel
}

// CalculateTreeTotalPoints calculates total skill points spent in a skill tree.
// This is the preferred way to calculate total points for ECS compliance.
//
// Parameters:
//   - tree: The skill tree to calculate points for
//
// Returns the sum of (level * skill points cost) for all learned skills.
func CalculateTreeTotalPoints(tree *SkillTree) int {
	if tree == nil {
		return 0
	}
	total := 0
	for _, node := range tree.Nodes {
		if node.Skill.Level > 0 {
			total += node.Skill.Level * node.Skill.Requirements.SkillPoints
		}
	}
	return total
}

// FindSkillByID finds a skill in a tree by its ID.
// This is the preferred way to look up skills for ECS compliance.
//
// Parameters:
//   - tree: The skill tree to search
//   - id: The skill ID to find
//
// Returns the skill if found, nil otherwise.
func FindSkillByID(tree *SkillTree, id string) *Skill {
	if tree == nil {
		return nil
	}
	for _, node := range tree.Nodes {
		if node.Skill.ID == id {
			return node.Skill
		}
	}
	return nil
}

// GetSkillsByTier returns all skills in a tree at a specific tier.
// This is the preferred way to filter skills by tier for ECS compliance.
//
// Parameters:
//   - tree: The skill tree to search
//   - tier: The tier to filter by
//
// Returns a slice of skills at the specified tier (may be empty).
func GetSkillsByTier(tree *SkillTree, tier Tier) []*Skill {
	if tree == nil {
		return nil
	}
	skills := make([]*Skill, 0)
	for _, node := range tree.Nodes {
		if node.Skill.Tier == tier {
			skills = append(skills, node.Skill)
		}
	}
	return skills
}
