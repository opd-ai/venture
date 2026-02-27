// Package skills provides procedural generation for skill trees and character progression systems.
//
// This package implements deterministic generation of skill trees with interconnected nodes,
// prerequisites, and balanced progression paths. Skills can be passive bonuses, active abilities,
// ultimate powers, or synergy skills that enhance other abilities.
//
// # Features
//
//   - Multiple skill tree archetypes per genre (Warrior, Mage, Rogue, etc.)
//   - Tier-based progression with increasing power
//   - Prerequisite system with skill dependencies
//   - Balanced stat scaling based on depth and difficulty
//   - Support for multiple genres (fantasy, sci-fi)
//   - Deterministic generation from seed values
//
// # Usage
//
// Basic skill tree generation:
//
//	generator := skills.NewSkillTreeGenerator()
//	params := procgen.GenerationParams{
//	    Depth: 10,
//	    Difficulty: 0.5,
//	    GenreID: "fantasy",
//	    Custom: map[string]interface{}{
//	        "count": 3, // Generate 3 trees
//	    },
//	}
//
//	result, err := generator.Generate(12345, params)
//	if err != nil {
//	    // Note: Production code should use logrus.WithError(err).Fatal()
//	    return err
//	}
//
//	trees := result.([]*skills.SkillTree)
//	// Note: Production code should use logrus.WithFields for structured logging
//	for _, tree := range trees {
//	    // Example: logrus.WithFields(logrus.Fields{"name": tree.Name, "skill_count": len(tree.Nodes), "max_points": tree.MaxPoints}).Info("Generated skill tree")
//	}
//
// # Skill Types
//
// Skills are classified into four types:
//
//   - Passive: Always-active bonuses (no activation required)
//   - Active: Player-activated abilities (cooldown-based)
//   - Ultimate: Powerful abilities with significant impact
//   - Synergy: Skills that enhance other skills
//
// # Skill Trees
//
// Each skill tree represents a character archetype with:
//   - 15-25 skills arranged in 7 tiers (0-6)
//   - Pyramid structure (more skills in lower tiers)
//   - Prerequisite chains requiring previous tier skills
//   - Category focus (Combat, Defense, Magic, Utility, etc.)
//
// # Progression System
//
// Skills have requirements that must be met:
//   - Player Level: Minimum character level
//   - Skill Points: Currency for learning skills
//   - Prerequisites: Other skills that must be learned first
//   - Attributes: Minimum stat requirements (optional)
//
// # Validation
//
// Generated skill trees are validated to ensure:
//   - All nodes have valid skills with names and effects
//   - Prerequisites reference existing skills
//   - Root nodes exist (skills with no prerequisites)
//   - Skill stats and levels are valid
//
// # Integration
//
// This package follows the procgen.Generator interface and integrates
// seamlessly with other procedural generation systems in the Venture project.
//
// # Integration with SkillProgressionSystem
//
// The generated skill trees integrate with engine.SkillProgressionSystem for
// runtime progression tracking:
//
//	// Generate skill trees
//	gen := skills.NewSkillTreeGenerator()
//	result, _ := gen.Generate(seed, params)
//	trees := result.([]*skills.SkillTree)
//
//	// Initialize progression system with generated trees
//	world := engine.NewWorld()
//	progressionSys := engine.NewSkillProgressionSystem(world, trees)
//	world.AddSystem(progressionSys)
//
//	// Track player progression
//	playerEntity := world.CreateEntity()
//	playerEntity.AddComponent(&engine.SkillProgressionComponent{
//	    AvailablePoints: 5,
//	    UnlockedSkills: make(map[string]bool),
//	})
//
//	// Unlock skills at runtime
//	progressionSys.UnlockSkill(playerEntity, "fireball")
package skills
