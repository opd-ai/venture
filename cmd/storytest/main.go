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
	mode := flag.String("mode", "list", "Mode: list, single, all, branching, crossdungeon, timeline, archaeology")
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
	case "branching":
		testBranchingNarrative(*seed, params)
	case "crossdungeon":
		testCrossDungeonStory(*seed, params)
	case "timeline":
		testTimeline(*seed, params)
	case "archaeology":
		testArchaeology(*seed, params)
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

func testBranchingNarrative(seed int64, params procgen.GenerationParams) {
	fmt.Printf("=== Branching Narrative Test ===\n")
	fmt.Printf("Seed: %d | Genre: %s | Difficulty: %.2f | Depth: %d\n\n", seed, params.GenreID, params.Difficulty, params.Depth)

	gen := story.NewBranchingNarrativeGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	bn := result.(*story.BranchingNarrative)
	if err := gen.Validate(bn); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Series ID: %s | Theme: %s\n", bn.SeriesID, bn.Theme)
	fmt.Printf("Choice Points: %d | Total Paths: %d\n\n", len(bn.ChoicePoints), len(bn.Paths))

	fmt.Println("Choice Points:")
	for i, cp := range bn.ChoicePoints {
		fmt.Printf("  [%d] %s\n", i+1, cp.Description)
		for j, opt := range cp.Options {
			fmt.Printf("      %c: %s\n", 'A'+rune(j), opt)
		}
	}

	fmt.Printf("\nSimulating Path (all choice A):\n")
	for i := range bn.ChoicePoints {
		if err := bn.MakeChoice(i, 0); err != nil {
			fmt.Fprintf(os.Stderr, "Choice error: %v\n", err)
			os.Exit(1)
		}
	}

	path := bn.GetActivePath()
	if path != nil {
		fmt.Printf("  Path ID: %s | Fragments: %d | Title: %s\n", path.PathID, len(path.Fragments), path.Title)
		for i, frag := range path.Fragments {
			content := frag.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			fmt.Printf("  [%d] %s\n", i+1, content)
		}
	}
}

func testCrossDungeonStory(seed int64, params procgen.GenerationParams) {
	fmt.Printf("=== Cross-Dungeon Story Test ===\n")
	fmt.Printf("Seed: %d | Genre: %s | Difficulty: %.2f | Depth: %d\n\n", seed, params.GenreID, params.Difficulty, params.Depth)

	gen := story.NewCrossDungeonGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	cd := result.(*story.CrossDungeonStory)
	if err := gen.Validate(cd); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Title: %s\n", cd.Title)
	fmt.Printf("Level Span: %d-%d | Fragments: %d | Continuity: %.2f\n\n", cd.MinDepth, cd.MaxDepth, len(cd.Fragments), cd.Continuity)

	fmt.Println("Level Distribution:")
	levels := cd.GetRequiredLevels()
	for _, lvl := range levels {
		frags := cd.GetFragmentsForLevel(lvl)
		fmt.Printf("  Level %d: %d fragments\n", lvl, len(frags))
	}

	fmt.Printf("\nFragment Chain:\n")
	for i, frag := range cd.Fragments {
		prereqs := ""
		if len(frag.Prerequisite) > 0 {
			prereqs = fmt.Sprintf(" [requires: %v]", frag.Prerequisite)
		}
		content := frag.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}
		fmt.Printf("  [%d] Level %d%s\n      %s\n", i+1, frag.Level.Depth, prereqs, content)
	}
}

func testTimeline(seed int64, params procgen.GenerationParams) {
	fmt.Printf("=== Historical Timeline Test ===\n")
	fmt.Printf("Seed: %d | Genre: %s | Difficulty: %.2f | Depth: %d\n\n", seed, params.GenreID, params.Difficulty, params.Depth)

	gen := story.NewTimelineGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	tl := result.(*story.Timeline)
	if err := gen.Validate(tl); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Timespan: %d-%d | Eras: %d | Events: %d | Consistency: %.2f\n\n", tl.StartYear, tl.CurrentYear, len(tl.Eras), len(tl.Events), tl.Consistency)

	fmt.Println("Historical Eras:")
	for i, era := range tl.Eras {
		desc := era.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  [%d] %s (%d-%d)\n      %s\n", i+1, era.Name, era.StartYear, era.EndYear, desc)
	}

	fmt.Printf("\nMajor Events:\n")
	for i, evt := range tl.Events {
		desc := evt.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  [%d] Year %d: %s (%s)\n      %s\n", i+1, evt.Timestamp, evt.Title, evt.EventType, desc)
	}

	currentEra := tl.GetCurrentEra()
	if currentEra != nil {
		fmt.Printf("\nCurrent Era: %s (%d-%d)\n", currentEra.Name, currentEra.StartYear, currentEra.EndYear)
	}
}

func testArchaeology(seed int64, params procgen.GenerationParams) {
	fmt.Printf("=== Archaeological Site Test ===\n")
	fmt.Printf("Seed: %d | Genre: %s | Difficulty: %.2f | Depth: %d\n\n", seed, params.GenreID, params.Difficulty, params.Depth)

	gen := story.NewArchaeologyGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	site := result.(*story.ArchaeologicalSite)
	if err := gen.Validate(site); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Name: %s\n", site.Name)
	fmt.Printf("Era: %s | Danger: %.2f\n", site.Era, site.Danger)
	fmt.Printf("Description: %s\n\n", site.Description)

	fmt.Printf("Artifacts (%d total):\n", len(site.Artifacts))
	for i, artifact := range site.Artifacts {
		desc := artifact.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  [%d] %s (%s)\n", i+1, artifact.Name, artifact.Type)
		fmt.Printf("      Condition: %.0f%% | Power: %.2f\n", artifact.Condition*100, artifact.PowerLevel)
		fmt.Printf("      %s\n", desc)
	}

	fmt.Printf("\nExcavation Simulation:\n")
	for step := 0; step < 4; step++ {
		site.Excavate(0.25)
		excavated := 0
		// Count excavated artifacts by checking progress
		progress := site.GetExcavationProgress()
		thresholds := make([]float64, len(site.Artifacts))
		for j := range thresholds {
			thresholds[j] = float64(j+1) / float64(len(site.Artifacts))
		}
		for _, threshold := range thresholds {
			if progress >= threshold {
				excavated++
			}
		}
		fmt.Printf("  %.0f%% - Recovered %d/%d artifacts\n", progress*100, excavated, len(site.Artifacts))
	}
}
