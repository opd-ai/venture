// Package games provides implementations of mini-game types for Venture.
//
// This package contains 7 playable mini-game types that implement the
// engine.MiniGame interface:
//
//   - CardGame: Procedural card game with deck and AI opponent (5-10 min)
//   - DiceGame: Custom dice rules with betting mechanics (2-5 min)
//   - PuzzleGame: Sliding tiles and pattern matching (3-7 min)
//   - MemoryGame: Card pairs and sequence repetition (2-4 min)
//   - LockPickingGame: Timing-based lock-picking challenges (0.5-2 min)
//   - HackingGame: Terminal/console puzzle for sci-fi genre (1-3 min)
//   - RitualGame: Spell pattern drawing for fantasy/horror (2-5 min)
//
// # Design Philosophy
//
// All games follow these principles:
//
//  1. Deterministic: Same seed produces identical gameplay
//  2. Difficulty Scaling: Parameters adjust from easy (0.0) to hard (1.0)
//  3. Genre Appropriate: Themed for fantasy, sci-fi, horror, etc.
//  4. Multiplayer Ready: State can be synchronized across clients
//  5. Reward System: Gold and XP scale with difficulty and performance
//
// # Usage Example
//
//	// Create and initialize a card game
//	game := games.NewCardGame()
//	if err := game.Initialize(12345, 0.5); err != nil {
//	    log.Fatal(err) // Note: Use logrus.WithError in production
//	}
//
//	// Game loop
//	for !game.IsComplete() {
//	    if err := game.Update(deltaTime); err != nil {
//	        log.Fatal(err) // Note: Use logrus.WithError in production
//	    }
//	    game.Render(screen)
//	}
//
//	// Award rewards
//	if reward := game.GetReward(); reward != nil {
//	    player.Gold += reward.Gold
//	    player.XP += reward.XP
//	}
//
// # Performance Characteristics
//
// All games are designed to complete within their target durations:
//
//   - Initialize: <1ms per game
//   - Update: <0.1ms per frame
//   - Render: <0.1ms per frame (state computation only)
//   - Memory: <1KB per game instance
//
// # Rendering
//
// Each game's Render() method computes visual state (RenderOutput) rather than
// directly drawing pixels, since ImageProvider is read-only. The ECS render
// system reads the LastRender field to perform actual pixel drawing. Render
// validates screen parameters and returns errors for nil screens, zero
// dimensions, or uninitialized games.
//
// # Testing
//
// All games include comprehensive test coverage:
//
//   - Determinism verification (same seed = same result)
//   - Difficulty scaling validation
//   - Completion detection
//   - Reward calculation
//   - Error handling
//
// Phase 27.3: Mini-Game Rendering (ROADMAP_V4.md)
package games
