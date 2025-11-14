package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/story"
)

func main() {
	// Command line flags
	mode := flag.String("mode", "list", "Mode: list, single, all")
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	genre := flag.String("genre", "fantasy", "Genre: fantasy, scifi, horror, cyberpunk, postapocalyptic")
	depth := flag.Int("depth", 5, "Dungeon depth level")
	difficulty := flag.Float64("difficulty", 0.5, "Difficulty (0.0-1.0)")
	seriesID := flag.String("series", "", "Series ID to filter (single mode only)")
	verbose := flag.Bool("verbose", false, "Show full fragment content")

	flag.Parse()

	// Validate inputs
	if *difficulty < 0 || *difficulty > 1.0 {
		fmt.Fprintf(os.Stderr, "Error: difficulty must be between 0.0 and 1.0\n")
		os.Exit(1)
	}

	validGenres := map[string]bool{
		"fantasy": true, "scifi": true, "horror": true,
		"cyberpunk": true, "postapocalyptic": true,
	}
	if !validGenres[*genre] {
		fmt.Fprintf(os.Stderr, "Error: invalid genre '%s'\n", *genre)
		os.Exit(1)
	}

	// Create generator and params
	gen := story.NewFragmentGenerator()
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
	}

	switch *mode {
	case "list":
		listFragmentTypes()
	case "single":
		generateSingle(gen, *seed, params, *seriesID, *verbose)
	case "all":
		generateAll(gen, *seed, params, *verbose)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown mode '%s'\n", *mode)
		os.Exit(1)
	}
}

func listFragmentTypes() {
	fmt.Println("Available Story Fragment Types:")
	fmt.Println("  0: Note      - Written notes, journals, papers")
	fmt.Println("  1: Carving   - Wall inscriptions and etchings")
	fmt.Println("  2: Corpse    - Bodies with clues")
	fmt.Println("  3: Relic     - Ancient artifacts")
	fmt.Println("  4: Graffiti  - Recent markings")
	fmt.Println("  5: Blood     - Blood trails and splatters")
	fmt.Println()
	fmt.Println("Genres: fantasy, scifi, horror, cyberpunk, postapocalyptic")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  storytest -mode list")
	fmt.Println("  storytest -mode single -seed 12345 -genre fantasy")
	fmt.Println("  storytest -mode all -seed 54321 -genre scifi -verbose")
}

func generateSingle(gen *story.FragmentGenerator, seed int64, params procgen.GenerationParams, seriesID string, verbose bool) {
	// Generate story sequence
	result, err := gen.Generate(seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	sequence, ok := result.(*story.StorySequence)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid result type\n")
		os.Exit(1)
	}

	// Validate
	if err := gen.Validate(sequence); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	printSequence(sequence, verbose)
}

func generateAll(gen *story.FragmentGenerator, seed int64, params procgen.GenerationParams, verbose bool) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for i, genreID := range genres {
		params.GenreID = genreID
		genreSeed := seed + int64(i*1000) // Different seed per genre

		result, err := gen.Generate(genreSeed, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s story: %v\n", genreID, err)
			continue
		}

		sequence, ok := result.(*story.StorySequence)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid result for %s\n", genreID)
			continue
		}

		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("GENRE: %s\n", strings.ToUpper(genreID))
		fmt.Println(strings.Repeat("=", 80))
		printSequence(sequence, verbose)
		fmt.Println()
	}
}

func printSequence(seq *story.StorySequence, verbose bool) {
	fmt.Printf("Story Title: %s\n", seq.Title)
	fmt.Printf("Series ID:   %s\n", seq.SeriesID)
	fmt.Printf("Genre:       %s\n", seq.Genre)
	fmt.Printf("Theme:       %s\n", seq.Theme)
	fmt.Printf("Coherence:   %.2f\n", seq.Coherence)
	fmt.Printf("Fragments:   %d\n", len(seq.Fragments))
	fmt.Println()

	// Print fragments
	for i, frag := range seq.Fragments {
		fmt.Printf("Fragment %d/%d:\n", i+1, len(seq.Fragments))
		fmt.Printf("  Type:         %s\n", frag.Type.String())
		fmt.Printf("  Sequence:     #%d in %s\n", frag.SequenceNum, frag.SeriesID)
		fmt.Printf("  Location:     (%.1f, %.1f)\n", frag.Location.X, frag.Location.Y)
		fmt.Printf("  Discovery XP: %.0f\n", frag.DiscoveryXP)
		fmt.Printf("  Sprite:       %s\n", frag.SpritePattern)

		if verbose {
			fmt.Printf("  Content:\n")
			// Indent content
			lines := strings.Split(frag.Content, "\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		} else {
			// Show truncated content
			content := frag.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			fmt.Printf("  Content:      %s\n", content)
		}
		fmt.Println()
	}

	// Print statistics
	fmt.Println("Fragment Type Distribution:")
	typeCounts := make(map[story.FragmentType]int)
	for _, frag := range seq.Fragments {
		typeCounts[frag.Type]++
	}

	for fragType := story.FragmentNote; fragType <= story.FragmentBlood; fragType++ {
		count := typeCounts[fragType]
		percentage := float64(count) / float64(len(seq.Fragments)) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", fragType.String(), count, percentage)
	}
	fmt.Println()

	// Print sprite pattern distribution
	fmt.Println("Sprite Patterns Used:")
	patternCounts := make(map[string]int)
	for _, frag := range seq.Fragments {
		patternCounts[frag.SpritePattern]++
	}
	for pattern, count := range patternCounts {
		fmt.Printf("  %s: %d\n", pattern, count)
	}
}
