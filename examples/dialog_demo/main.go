package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// main demonstrates NPC dialog generation using Markov chains with
// genre-specific corpora, personality traits, and controlled non-determinism.
func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	deterministic := flag.Bool("deterministic", false, "Use deterministic mode (same seed = same output)")
	genre := flag.String("genre", "fantasy", "Genre for dialog generation (fantasy, scifi, horror, cyberpunk, postapocalyptic)")
	order := flag.Int("order", 2, "Markov chain order (2 or 3)")
	flag.Parse()

	fmt.Println("=== Venture NPC Dialog System Demonstration ===")
	fmt.Printf("Genre: %s, Order: %d, Deterministic: %v\n", *genre, *order, *deterministic)

	// Convert order to MarkovOrder type
	var markovOrder dialog.MarkovOrder
	if *order == 3 {
		markovOrder = dialog.Order3
	} else {
		markovOrder = dialog.Order2
	}

	// Demonstrate basic dialog generation
	demonstrateBasicGeneration(*genre, markovOrder, *deterministic, *verbose)

	// Demonstrate dialog variation
	demonstrateVariation(*genre, markovOrder, *deterministic, *verbose)

	// Demonstrate determinism (if enabled)
	if *deterministic {
		demonstrateDeterminism(*genre, markovOrder, *verbose)
	}

	// Demonstrate all genres
	demonstrateAllGenres(markovOrder, *deterministic, *verbose)

	fmt.Println("\n=== Dialog Demo Complete ===")
}

// demonstrateBasicGeneration shows basic dialog generation.
func demonstrateBasicGeneration(genreID string, order dialog.MarkovOrder, deterministic, verbose bool) {
	fmt.Println("\n--- Demonstrating Basic Dialog Generation ---")

	// Create generator
	seed := int64(12345)
	generator := dialog.NewMarkovGenerator(seed, genreID, order)

	// Get corpus for genre
	corpus := dialog.GetCorpus(genreID)
	if corpus == nil {
		log.Fatalf("Failed to get corpus for genre: %s", genreID)
	}

	// Train generator
	generator.TrainFromCorpus(corpus.Sentences)

	// Generate a few sample responses
	questions := []string{
		"Hello, can you help me?",
		"What do you know about this place?",
		"Do you have any quests for me?",
	}

	for _, question := range questions {
		fmt.Printf("\nPlayer: \"%s\"\n", question)

		params := dialog.GenerateParams{
			PlayerInput:    question,
			ConversationID: "demo_conversation",
			MaxWords:       30,
			MinWords:       10,
			Temperature:    0.7,
		}

		var response string
		if deterministic {
			response = generator.GenerateDeterministic(params)
		} else {
			response = generator.Generate(params)
		}

		if response == "" {
			log.Printf("  Error: Empty response\n")
			continue
		}

		fmt.Printf("NPC: \"%s\"\n", response)

		if verbose {
			fmt.Printf("  (Length: %d words)\n", len(splitWords(response)))
		}
	}
}

// demonstrateVariation shows that same input produces different outputs in non-deterministic mode.
func demonstrateVariation(genreID string, order dialog.MarkovOrder, deterministic, verbose bool) {
	fmt.Println("\n--- Demonstrating Dialog Variation ---")

	seed := int64(12345)
	if !deterministic {
		fmt.Println("Non-deterministic mode: Same input should produce varied responses")
	} else {
		fmt.Println("Deterministic mode: Same input produces identical responses")
	}

	// Generate 5 responses to same input
	responses := make([]string, 5)
	for i := 0; i < 5; i++ {
		// In deterministic mode, use same seed; in non-deterministic, use different seeds
		currentSeed := seed
		if !deterministic {
			currentSeed = seed + int64(i)
		}
		generator := dialog.NewMarkovGenerator(currentSeed, genreID, order)

		corpus := dialog.GetCorpus(genreID)
		if corpus == nil {
			log.Fatalf("Failed to get corpus for genre: %s", genreID)
		}
		generator.TrainFromCorpus(corpus.Sentences)

		params := dialog.GenerateParams{
			PlayerInput:    "What do you know about this place?",
			ConversationID: "demo_variation",
			MaxWords:       25,
			MinWords:       10,
			Temperature:    0.7,
		}

		var response string
		if deterministic {
			response = generator.GenerateDeterministic(params)
		} else {
			response = generator.Generate(params)
		}

		if response == "" {
			log.Printf("Generation %d error: empty response\n", i+1)
			continue
		}

		responses[i] = response
		fmt.Printf("\nResponse %d: \"%s\"\n", i+1, response)
	}

	// Check uniqueness
	unique := make(map[string]bool)
	for _, r := range responses {
		if r != "" {
			unique[r] = true
		}
	}

	uniqueCount := len(unique)
	totalCount := len(responses)
	uniquePercent := float64(uniqueCount) / float64(totalCount) * 100

	fmt.Printf("\nUniqueness: %d/%d responses unique (%.1f%%)\n", uniqueCount, totalCount, uniquePercent)

	if !deterministic {
		if uniquePercent >= 80.0 {
			fmt.Println("✓ Variation target achieved (>80% unique)")
		} else {
			fmt.Printf("⚠ Low variation (target: >80%%, got: %.1f%%)\n", uniquePercent)
		}
	} else {
		if uniqueCount == 1 {
			fmt.Println("✓ Determinism verified (all responses identical)")
		} else {
			fmt.Printf("⚠ Determinism violation (expected 1 unique, got %d)\n", uniqueCount)
		}
	}
}

// demonstrateDeterminism verifies that same seed produces identical output.
func demonstrateDeterminism(genreID string, order dialog.MarkovOrder, verbose bool) {
	fmt.Println("\n--- Demonstrating Deterministic Generation ---")

	const fixedSeed = int64(99999)
	const numRuns = 10

	fmt.Printf("Generating %d responses with same seed (%d)...\n", numRuns, fixedSeed)

	responses := make([]string, numRuns)
	for i := 0; i < numRuns; i++ {
		generator := dialog.NewMarkovGenerator(fixedSeed, genreID, order)

		corpus := dialog.GetCorpus(genreID)
		if corpus == nil {
			log.Fatalf("Failed to get corpus for genre: %s", genreID)
		}
		generator.TrainFromCorpus(corpus.Sentences)

		params := dialog.GenerateParams{
			PlayerInput:    "Tell me about your quest.",
			ConversationID: "determinism_test",
			MaxWords:       20,
			MinWords:       10,
			Temperature:    0.7,
		}

		response := generator.GenerateDeterministic(params)
		if response == "" {
			log.Fatalf("Generation failed: empty response")
		}

		responses[i] = response
	}

	// Verify all responses are identical
	allIdentical := true
	for i := 1; i < numRuns; i++ {
		if responses[i] != responses[0] {
			allIdentical = false
			if verbose {
				fmt.Printf("Mismatch at run %d:\n", i+1)
				fmt.Printf("  Expected: \"%s\"\n", responses[0])
				fmt.Printf("  Got:      \"%s\"\n", responses[i])
			}
			break
		}
	}

	if allIdentical {
		fmt.Printf("✓ Determinism verified: All %d responses identical\n", numRuns)
		fmt.Printf("  Response: \"%s\"\n", responses[0])
	} else {
		log.Fatal("✗ Determinism violation: Responses differ with same seed!")
	}
}

// demonstrateAllGenres shows dialog generation across all supported genres.
func demonstrateAllGenres(order dialog.MarkovOrder, deterministic, verbose bool) {
	fmt.Println("\n--- Demonstrating All Genre Styles ---")

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for idx, genreID := range genres {
		fmt.Printf("\n%s Genre:\n", capitalize(genreID))

		seed := int64(54321 + idx) // Different seed per genre
		generator := dialog.NewMarkovGenerator(seed, genreID, order)

		corpus := dialog.GetCorpus(genreID)
		if corpus == nil {
			log.Printf("  Error: No corpus for genre %s\n", genreID)
			continue
		}

		generator.TrainFromCorpus(corpus.Sentences)

		params := dialog.GenerateParams{
			PlayerInput:    "What dangers lurk here?",
			ConversationID: fmt.Sprintf("demo_%s", genreID),
			MaxWords:       25,
			MinWords:       10,
			Temperature:    0.7,
		}

		var response string
		if deterministic {
			response = generator.GenerateDeterministic(params)
		} else {
			response = generator.Generate(params)
		}

		if response == "" {
			log.Printf("  Error: Empty response\n")
			continue
		}

		fmt.Printf("  \"%s\"\n", response)

		if verbose {
			// Show some corpus stats
			fmt.Printf("  (Corpus size: %d sentences)\n", len(corpus.Sentences))
		}
	}
}

// Helper functions

func splitWords(text string) []string {
	words := []string{}
	current := ""
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
