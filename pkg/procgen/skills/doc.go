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
//	    log.Fatal(err)
//	}
//
//	trees := result.([]*skills.SkillTree)
//	for _, tree := range trees {
//	    fmt.Printf("Tree: %s (%s)\n", tree.Name, tree.Description)
//	    fmt.Printf("Skills: %d, Max Points: %d\n", len(tree.Nodes), tree.MaxPoints)
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
package skills
