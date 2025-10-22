// Package quest provides procedural quest generation for Venture.
//
// This package generates various types of quests with objectives, rewards,
// and thematic descriptions. All generation is deterministic based on a seed value.
//
// Quest Types:
//   - Kill: Defeat a specific number of enemies
//   - Collect: Gather items from the world
//   - Escort: Protect an NPC to a destination
//   - Explore: Discover a location
//   - Talk: Interact with an NPC
//   - Boss: Defeat a specific boss enemy
//
// Usage:
//
//	generator := quest.NewQuestGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	    Custom:     map[string]interface{}{"count": 5},
//	}
//	result, err := generator.Generate(12345, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	quests := result.([]*quest.Quest)
//
// The generator uses genre-specific templates to create thematically appropriate
// quests with names, descriptions, and objectives that match the game's theme.
package quest
