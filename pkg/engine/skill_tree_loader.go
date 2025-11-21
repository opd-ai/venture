// Package engine provides the core game systems including skill tree loading.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/sirupsen/logrus"
)

var skillTreeLog *logrus.Logger

func init() {
	skillTreeLog = logrus.New()
	skillTreeLog.SetReportCaller(true)
	skillTreeLog.SetLevel(logrus.DebugLevel)
}

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
	skillTreeLog.WithFields(logrus.Fields{
		"entity_id": player.ID,
		"seed":      seed,
		"genre_id":  genreID,
		"depth":     depth,
	}).Debug("Entering LoadPlayerSkillTree")

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

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id":  player.ID,
		"seed":       seed,
		"genre_id":   genreID,
		"depth":      depth,
		"difficulty": params.Difficulty,
		"tree_count": 3,
	}).Debug("Calling skill tree generator")

	result, err := generator.Generate(seed, params)
	if err != nil {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id": player.ID,
			"seed":      seed,
			"genre_id":  genreID,
			"depth":     depth,
			"error":     err.Error(),
		}).Error("Skill tree generation failed")
		return err
	}

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id": player.ID,
		"seed":      seed,
		"genre_id":  genreID,
	}).Debug("Skill tree generation completed successfully")

	trees := result.([]*skills.SkillTree)
	if len(trees) == 0 {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id": player.ID,
			"seed":      seed,
			"genre_id":  genreID,
		}).Warn("No skill trees generated, returning early")
		return nil // No trees generated, not an error
	}

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id":       player.ID,
		"trees_generated": len(trees),
		"tree_name":       trees[0].Name,
	}).Debug("Selecting main skill tree")

	// Use first tree as the main skill tree
	// (In a full game, players could choose or have multiple trees)
	mainTree := trees[0]

	// Create skill tree component if doesn't exist
	if !player.HasComponent("skill_tree") {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":      player.ID,
			"component_type": "skill_tree",
			"tree_name":      mainTree.Name,
			"skill_count":    len(mainTree.Nodes),
		}).Debug("Creating new skill tree component")

		comp := NewSkillTreeComponent(mainTree)
		player.AddComponent(comp)

		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":      player.ID,
			"component_type": "skill_tree",
			"tree_name":      mainTree.Name,
		}).Info("Skill tree component created and attached to player")
	} else {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":      player.ID,
			"component_type": "skill_tree",
		}).Debug("Skill tree component already exists, updating")

		// Update existing component with new tree
		comp, ok := player.GetComponent("skill_tree")
		if ok {
			if treeComp, ok := comp.(*SkillTreeComponent); ok {
				oldTreeName := treeComp.Tree.Name
				treeComp.Tree = mainTree

				skillTreeLog.WithFields(logrus.Fields{
					"entity_id":       player.ID,
					"component_type":  "skill_tree",
					"old_tree_name":   oldTreeName,
					"new_tree_name":   mainTree.Name,
					"new_skill_count": len(mainTree.Nodes),
				}).Info("Skill tree component updated with new tree")
			} else {
				skillTreeLog.WithFields(logrus.Fields{
					"entity_id":      player.ID,
					"component_type": "skill_tree",
				}).Warn("Failed to cast component to SkillTreeComponent")
			}
		} else {
			skillTreeLog.WithFields(logrus.Fields{
				"entity_id":      player.ID,
				"component_type": "skill_tree",
			}).Warn("Failed to retrieve skill_tree component despite HasComponent check")
		}
	}

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id": player.ID,
		"tree_name": mainTree.Name,
	}).Debug("Exiting LoadPlayerSkillTree successfully")

	return nil
}

// GetPlayerSkillPoints calculates available skill points based on player level.
// Players earn 1 skill point per level, with bonus points at milestones.
//
// Formula: base points = level - 1, bonus at levels 10, 20, 30, etc.
func GetPlayerSkillPoints(playerLevel int) int {
	skillTreeLog.WithFields(logrus.Fields{
		"player_level": playerLevel,
	}).Debug("Entering GetPlayerSkillPoints")

	basePoints := playerLevel - 1         // Start at level 1 with 0 points
	bonusPoints := (playerLevel / 10) * 2 // +2 points every 10 levels
	totalPoints := basePoints + bonusPoints

	skillTreeLog.WithFields(logrus.Fields{
		"player_level": playerLevel,
		"base_points":  basePoints,
		"bonus_points": bonusPoints,
		"total_points": totalPoints,
	}).Debug("Calculated skill points")

	return totalPoints
}

// GetUnspentSkillPoints returns the number of skill points available to spend.
func GetUnspentSkillPoints(player *Entity) int {
	skillTreeLog.WithFields(logrus.Fields{
		"entity_id": player.ID,
	}).Debug("Entering GetUnspentSkillPoints")

	// Get player level
	var playerLevel int
	if comp, ok := player.GetComponent("experience"); ok {
		if expComp, ok := comp.(*ExperienceComponent); ok {
			playerLevel = expComp.Level
			skillTreeLog.WithFields(logrus.Fields{
				"entity_id":      player.ID,
				"player_level":   playerLevel,
				"component_type": "experience",
			}).Debug("Retrieved player level from experience component")
		} else {
			skillTreeLog.WithFields(logrus.Fields{
				"entity_id":      player.ID,
				"component_type": "experience",
			}).Warn("Failed to cast experience component")
		}
	} else {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":      player.ID,
			"component_type": "experience",
		}).Debug("No experience component found, defaulting to level 1")
	}

	if playerLevel == 0 {
		playerLevel = 1
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":    player.ID,
			"player_level": playerLevel,
		}).Debug("Player level was 0, set to default level 1")
	}

	// Calculate total available points
	totalPoints := GetPlayerSkillPoints(playerLevel)

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id":    player.ID,
		"player_level": playerLevel,
		"total_points": totalPoints,
	}).Debug("Calculated total available skill points")

	// Subtract used points
	if comp, ok := player.GetComponent("skill_tree"); ok {
		if treeComp, ok := comp.(*SkillTreeComponent); ok {
			usedPoints := treeComp.TotalPointsUsed
			unspentPoints := totalPoints - usedPoints

			skillTreeLog.WithFields(logrus.Fields{
				"entity_id":      player.ID,
				"total_points":   totalPoints,
				"used_points":    usedPoints,
				"unspent_points": unspentPoints,
				"component_type": "skill_tree",
			}).Debug("Exiting GetUnspentSkillPoints with calculation")

			return unspentPoints
		} else {
			skillTreeLog.WithFields(logrus.Fields{
				"entity_id":      player.ID,
				"component_type": "skill_tree",
			}).Warn("Failed to cast skill_tree component")
		}
	} else {
		skillTreeLog.WithFields(logrus.Fields{
			"entity_id":      player.ID,
			"component_type": "skill_tree",
		}).Debug("No skill_tree component found, returning total points")
	}

	skillTreeLog.WithFields(logrus.Fields{
		"entity_id":      player.ID,
		"unspent_points": totalPoints,
	}).Debug("Exiting GetUnspentSkillPoints (no skill tree component)")

	return totalPoints
}
