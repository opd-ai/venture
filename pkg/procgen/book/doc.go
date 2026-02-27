// Package book provides procedural generation of in-game books and lore text.
//
// This package implements a grammar-based text generation system (Tracery-style)
// that creates readable, genre-appropriate book content for Venture's zero-asset
// procedural world. Books serve multiple gameplay purposes: skill progression,
// world-building lore, quest hints, crafting recipes, and environmental storytelling.
//
// # Book Types
//
// The generator produces five types of books:
//
//  1. Skill Manuals: Grant XP or unlock abilities when read
//  2. Lore Codices: Provide world-building and historical context
//  3. Quest Journals: Contain hints and clues for active quests
//  4. Recipe Tomes: Unlock crafting recipes and techniques
//  5. Historical Texts: Environmental storytelling about dungeons and locations
//
// # Generation Process
//
// Books are generated using a grammar-based system with the following steps:
//
//  1. Select genre-specific vocabulary and templates
//  2. Generate title using title grammar rules
//  3. Generate author name using name grammar
//  4. Create content pages (3-10 pages, 150-300 words each)
//  5. Apply genre-specific styling and formatting
//
// # Example Usage
//
// Note: Example uses simplified logging for clarity.
// Production code should use logrus.WithFields for structured logging.
//
//	generator := book.NewGenerator()
//	params := procgen.GenerationParams{
//	    Difficulty: 0.5,
//	    Depth: 5,
//	    GenreID: "fantasy",
//	    Custom: map[string]interface{}{
//	        "book_type": engine.BookTypeSkill,
//	        "skill_name": "Sword Fighting",
//	    },
//	}
//
//	result, err := generator.Generate(12345, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	book := result.(*engine.BookComponent)
//	fmt.Printf("Title: %s\nAuthor: %s\n", book.Title, book.Author)
//	for i, page := range book.Content {
//	    fmt.Printf("\n--- Page %d ---\n%s\n", i+1, page)
//	}
//
// # Performance
//
// - Generation time: <50ms per book (target)
// - Memory usage: <100KB per book
// - Text length: 500-2000 words per book
//
// # Determinism
//
// All book generation is deterministic based on seed. The same seed with the same
// parameters will always produce identical book content, ensuring multiplayer
// synchronization and reproducibility.
package book
