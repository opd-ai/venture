// Package legendary provides procedural generation of legendary quests and rewards.
//
// # Overview
//
// The legendary package implements multi-phase legendary quest generation with
// cross-server requirements, raid integration, and unique one-time rewards.
// Legendary quests are designed for endgame players seeking the most challenging
// and rewarding content in the game.
//
// # Quest Structure
//
// Legendary quests consist of 5-10 phases completed over multiple sessions:
//   - Exploration phases: Visit specific locations across 3-5 servers
//   - Combat phases: Clear specific raid encounters
//   - Crafting phases: Create rare items using housing stations
//   - Collection phases: Gather rare materials or complete achievements
//   - Final phase: Culminating challenge with legendary reward
//
// # Cross-Server Requirements
//
// Quests require visiting multiple federated servers to:
//   - Collect quest items from different server-specific locations
//   - Complete challenges unique to each server's environment
//   - Interact with NPCs across the federation
//
// # Raid Integration
//
// Quests integrate with V9 Phase 59.1 raid system:
//   - Require clearing specific raid encounters (Normal, Heroic, Mythic, etc.)
//   - Boss-specific drops contribute to quest progression
//   - Raid achievements unlock legendary quest phases
//
// # Crafting Challenges
//
// Quests require using V9 Phase 55.1 housing crafting stations:
//   - Create Master-tier items using advanced recipes
//   - Combine materials from multiple servers
//   - Utilize guild crafting stations for unique recipes
//
// # Legendary Rewards
//
// One-time rewards include:
//   - Legendary items with unique abilities (20-50 unique items)
//   - Cosmetic titles and visual effects
//   - Account-wide bonuses and achievements
//   - Rare mounts and companions
//
// # Example Usage
//
//	// Create legendary quest generator
//	gen := legendary.NewGenerator()
//
//	// Generate a legendary quest for a high-level player
//	params := procgen.GenerationParams{
//		Difficulty: 0.9,
//		Depth:      50,
//		GenreID:    "fantasy",
//		Custom: map[string]interface{}{
//			"player_level": 50,
//			"servers_visited": 5,
//		},
//	}
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//		logrus.WithError(err).Fatal("failed to generate legendary quest")
//	}
//	quest := result.(*legendary.LegendaryQuest)
//
//	// Track player progress
//	tracker := legendary.NewProgressTracker()
//	tracker.UpdatePhase(quest.ID, "player1", 0, 100.0)
//	progress := tracker.GetProgress(quest.ID, "player1")
//
// # Performance
//
// Generation targets:
//   - Quest generation: <500ms per quest (5-10 phases)
//   - Phase validation: <100ms per phase check
//   - Progress update: <10ms per player update
//   - Reward generation: <200ms per legendary item
//
// # Testing
//
// Test coverage targets ≥65% with deterministic generation verification.
// Use same seed to verify identical quest chains and reward distributions.
package legendary
