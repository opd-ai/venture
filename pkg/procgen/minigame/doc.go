// Package minigame provides procedural generation for embedded mini-games.
//
// The mini-game system generates 7 types of games with procedural rules, AI opponents,
// and genre-appropriate theming. All games are deterministic when given the same seed.
//
// Game Types:
//   - Card Game: Procedural deck, rules, AI opponent (5-10 min)
//   - Dice Game: Custom dice rules, betting mechanics (2-5 min)
//   - Puzzle: Sliding tiles, pattern matching (3-7 min)
//   - Memory: Card pairs, sequence repetition (2-4 min)
//   - Lock-Picking: Timing-based (0.5-2 min)
//   - Hacking: Terminal/console puzzle (sci-fi genre) (1-3 min)
//   - Ritual: Spell pattern drawing (fantasy/horror) (2-5 min)
//
// Usage:
//
//	gen := minigame.NewGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth:      5,
//	    GenreID:    "fantasy",
//	}
//	result, err := gen.Generate(12345, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	game := result.(*minigame.MiniGame)
//
// Performance: Generation time <100ms per game, memory <5MB per game instance.
package minigame
