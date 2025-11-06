// Package main provides a CLI tool for testing procedural puzzle generation.
// Phase 11.2 implementation demo.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/puzzle"
)

func main() {
	// Command-line flags
	seed := flag.Int64("seed", 12345, "Random seed for deterministic generation")
	difficulty := flag.Float64("difficulty", 0.5, "Difficulty level (0.0-1.0)")
	depth := flag.Int("depth", 5, "Dungeon depth level")
	genreID := flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapocalyptic)")
	count := flag.Int("count", 5, "Number of puzzles to generate")
	verbose := flag.Bool("verbose", false, "Show detailed puzzle information")

	flag.Parse()

	// Validate parameters
	if *difficulty < 0.0 || *difficulty > 1.0 {
		fmt.Fprintf(os.Stderr, "Error: difficulty must be between 0.0 and 1.0\n")
		os.Exit(1)
	}

	if *depth < 1 {
		fmt.Fprintf(os.Stderr, "Error: depth must be at least 1\n")
		os.Exit(1)
	}

	if *count < 1 {
		fmt.Fprintf(os.Stderr, "Error: count must be at least 1\n")
		os.Exit(1)
	}

	// Print header
	fmt.Println("=== Venture Puzzle Generator Demo ===")
	fmt.Printf("Seed: %d | Difficulty: %.2f | Depth: %d | Genre: %s\n\n",
		*seed, *difficulty, *depth, *genreID)

	// Create generator
	gen := puzzle.NewGenerator()

	// Generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genreID,
	}

	// Generate and display puzzles
	for i := 0; i < *count; i++ {
		// Use different seed for each puzzle
		puzzleSeed := *seed + int64(i)*1000

		fmt.Printf("Puzzle #%d (Seed: %d)\n", i+1, puzzleSeed)
		fmt.Println("---")

		// Generate puzzle
		result, err := gen.Generate(puzzleSeed, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating puzzle: %v\n", err)
			continue
		}

		// Type assert to puzzle
		puz, ok := result.(*puzzle.Puzzle)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: result is not a Puzzle type\n")
			continue
		}

		// Validate puzzle
		if err := gen.Validate(puz); err != nil {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
			continue
		}

		// Display puzzle information
		displayPuzzle(puz, *verbose)
		fmt.Println()
	}

	fmt.Println("Generation complete!")
}

func displayPuzzle(puz *puzzle.Puzzle, verbose bool) {
	// Basic information
	fmt.Printf("ID: %s\n", puz.ID)
	fmt.Printf("Type: %s\n", puz.Type)
	fmt.Printf("Difficulty: %d/10\n", puz.Difficulty)
	fmt.Printf("Description: %s\n", puz.Description)
	fmt.Printf("Hint: %s\n", puz.HintText)
	fmt.Printf("Elements: %d\n", puz.ElementCount)
	fmt.Printf("Solution Length: %d\n", len(puz.Solution))

	if puz.TimeLimit > 0 {
		fmt.Printf("Time Limit: %.1f seconds\n", puz.TimeLimit)
	}

	if puz.MaxAttempts > 0 {
		fmt.Printf("Max Attempts: %d\n", puz.MaxAttempts)
	}

	fmt.Printf("Reward: %s\n", puz.RewardType)

	if verbose {
		fmt.Println("\nElements:")
		for _, elem := range puz.Elements {
			fmt.Printf("  - %s (%s) at [%d, %d]\n",
				elem.ID, elem.ElementType, elem.Position[0], elem.Position[1])
		}

		fmt.Println("\nSolution Sequence:")
		for i, solID := range puz.Solution {
			fmt.Printf("  %d. %s\n", i+1, solID)
		}
	}
}
