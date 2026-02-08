// Package dialog provides runtime NPC dialog generation using Markov chains.
//
// This package generates dynamic, context-aware NPC dialog that varies based on
// player interaction history, NPC personality, and conversation context. It uses
// Markov chain text generation with controlled non-determinism for immersive
// dialog variety while maintaining deterministic fallback modes for testing.
//
// # Key Concepts
//
// **Markov Chain Generation:** Order 2-3 Markov chains trained on genre-specific
// text corpora produce varied, coherent dialog. The generation uses runtime entropy
// (player input history, conversation context) to create unique responses while
// maintaining grammatical structure.
//
// **Controlled Non-Determinism:** Dialog generation affects presentation only,
// never gameplay mechanics (quest objectives, item rewards, entity behavior). This
// preserves deterministic game state while enhancing immersion.
//
// **Deterministic Fallback:** Template-based dialog is available via the
// `-deterministic-dialog=true` flag for testing and reproducibility. Tests can
// verify both non-deterministic variation and deterministic reproducibility.
//
// **Server Authoritativeness:** All dialog is generated server-side and sent to
// clients, preventing client-side manipulation and ensuring consistency across
// multiplayer sessions.
//
// # Usage
//
// Generate NPC dialog with context and personality:
//
//	gen := dialog.NewMarkovGenerator(seed, genreID, order)
//	gen.TrainFromCorpus(corpus)
//
//	response := gen.GenerateText(dialog.GenerateParams{
//	    PlayerInput:    "Where is the dungeon?",
//	    ConversationID: "merchant-12345",
//	    NPCPersonality: dialog.PersonalityHelpful,
//	    MaxWords:       30,
//	})
//	// Response: "The ancient dungeon lies beyond the dark forest. Beware the creatures that dwell within."
//
// # Architecture
//
// The dialog system consists of three main components:
//
//  1. **MarkovGenerator** (markov.go): Core text generation using n-gram chains
//  2. **Corpus** (corpus.go): Genre-specific training data and vocabulary
//  3. **Personality** (personality.go): Traits that influence word selection
//
// # Performance
//
// Dialog generation is optimized for runtime use:
//
//   - Response generation: <50ms (target for real-time conversation)
//   - Memory footprint: ~2-5MB per trained generator (corpus + chain state)
//   - Training: One-time cost, <100ms for typical corpus (1000-5000 words)
//
// # Non-Determinism Scope
//
// **Non-Deterministic Elements:**
//   - Dialog text content (varies with player input, conversation history)
//   - Response variation for same input (enhances replayability)
//
// **Deterministic Elements:**
//   - NPC behavior (hostility, trading, quest offering)
//   - Quest objectives (goals, rewards, requirements)
//   - Item generation (stats, rarity, types)
//   - Combat outcomes (damage, status effects)
//
// # Testing
//
// Tests verify both non-deterministic behavior (variation) and deterministic
// reproducibility (when seeded):
//
//	// Variation test: Same input produces different responses
//	responses := make(map[string]bool)
//	for i := 0; i < 10; i++ {
//	    response := gen.GenerateText(params)
//	    responses[response] = true
//	}
//	// Expect >80% unique responses
//
//	// Deterministic test: Same seed + params = identical output
//	gen1 := dialog.NewMarkovGenerator(12345, "fantasy", 2)
//	gen2 := dialog.NewMarkovGenerator(12345, "fantasy", 2)
//	gen1.TrainFromCorpus(corpus)
//	gen2.TrainFromCorpus(corpus)
//	// Expect identical chain state and responses
//
// # Example
//
// Full dialog generation workflow:
//
//	// Initialize generator with genre-specific corpus
//	corpus := dialog.GetCorpus("fantasy")
//	gen := dialog.NewMarkovGenerator(worldSeed, "fantasy", 2)
//	gen.TrainFromCorpus(corpus)
//
//	// Configure NPC personality
//	personality := dialog.Personality{
//	    Type:        dialog.PersonalityMerchant,
//	    Friendliness: 0.7,
//	    Verbosity:    0.5,
//	    Formality:    0.6,
//	}
//
//	// Generate response to player greeting
//	response := gen.GenerateWithPersonality(dialog.GenerateParams{
//	    PlayerInput:    "Hello, merchant!",
//	    ConversationID: "npc-merchant-001",
//	    MaxWords:       20,
//	}, personality)
//
//	// Expected output (example):
//	// "Greetings, traveler. I have fine wares if you have coin. What interests you?"
package dialog
