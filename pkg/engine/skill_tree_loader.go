// Package engine provides the core game systems including skill tree loading.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// LoadPlayerSkillTree generates and attaches a procedural skill tree to the player entity.
// This function creates genre-themed skill trees with balanced progression paths.
//
// Parameters:
//   - player: The player entity to attach the skill tree component
//   - seed: Deterministic seed for skill tree generation
//   - genreID: Genre for themed skill names and effects (fantasy, scifi, horror, etc.)
//   - depth: Dungeon depth affecting skill power and complexity
//
// Returns error if generation fails.
//
// Usage:
//
//	if err := engine.LoadPlayerSkillTree(player, 12345, "fantasy", 0); err != nil {
//	    log.Fatal(err)
//	}
func LoadPlayerSkillTree(player *Entity, seed int64, genreID string, depth int) error {
	// Generate skill trees using procgen system
	generator := skills.NewSkillTreeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 3, // Generate 3 skill trees (combat, utility, magic typically)
		},
	}

	result, err := generator.Generate(seed, params)
	if err != nil {
		return err
	}

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		return nil // No trees generated, not an error
	}

	// Use first tree as the main skill tree
	// (In a full game, players could choose or have multiple trees)
	mainTree := trees[0]

	// Create skill tree component if doesn't exist
	if !player.HasComponent("skill_tree") {
		comp := NewSkillTreeComponent(mainTree)
		player.AddComponent(comp)
	} else {
		// Update existing component with new tree
		comp, ok := player.GetComponent("skill_tree")
		if ok {
			treeComp := comp.(*SkillTreeComponent)
			treeComp.Tree = mainTree
		}
	}

	return nil
}

// GetPlayerSkillPoints calculates available skill points based on player level.
// Players earn 1 skill point per level, with bonus points at milestones.
//
// Formula: base points = level - 1, bonus at levels 10, 20, 30, etc.
func GetPlayerSkillPoints(playerLevel int) int {
	basePoints := playerLevel - 1         // Start at level 1 with 0 points
	bonusPoints := (playerLevel / 10) * 2 // +2 points every 10 levels
	return basePoints + bonusPoints
}

// GetUnspentSkillPoints returns the number of skill points available to spend.
func GetUnspentSkillPoints(player *Entity) int {
	// Get player level
	var playerLevel int
	if comp, ok := player.GetComponent("experience"); ok {
		if expComp, ok := comp.(*ExperienceComponent); ok {
			playerLevel = expComp.Level
		}
	}
	if playerLevel == 0 {
		playerLevel = 1
	}

	// Calculate total available points
	totalPoints := GetPlayerSkillPoints(playerLevel)

	// Subtract used points
	if comp, ok := player.GetComponent("skill_tree"); ok {
		if treeComp, ok := comp.(*SkillTreeComponent); ok {
			return totalPoints - treeComp.TotalPointsUsed
		}
	}

	return totalPoints
}
