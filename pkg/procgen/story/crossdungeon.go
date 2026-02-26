package story

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// DungeonLevel represents which dungeon level contains a fragment
type DungeonLevel struct {
	Depth    int    // Dungeon depth (1, 2, 3, etc.)
	ZoneName string // Optional zone identifier
	Required bool   // Must this level be visited for story completion?
}

// CrossDungeonFragment extends StoryFragment with level information
type CrossDungeonFragment struct {
	StoryFragment
	Level         DungeonLevel
	Prerequisite  []int // Indices of fragments that must be found first
	UnlocksSecret bool  // Does finding this reveal a hidden area?
}

// CrossDungeonStory represents a narrative that spans multiple dungeon levels
type CrossDungeonStory struct {
	SeriesID        string                 // Unique story identifier
	Title           string                 // Story title
	Genre           string                 // Genre
	Theme           string                 // Story theme
	Fragments       []CrossDungeonFragment // All fragments across all levels
	MinDepth        int                    // Minimum depth to start story
	MaxDepth        int                    // Maximum depth for final fragments
	LevelSpan       int                    // How many levels the story spans
	CompletionBonus float64                // XP bonus for completing entire story
	Coherence       float64                // Story quality metric (0.0-1.0)
	Continuity      float64                // How well story flows across levels (0.0-1.0)
}

// CrossDungeonGenerator creates stories spanning multiple dungeon levels
type CrossDungeonGenerator struct{}

// NewCrossDungeonGenerator creates a new cross-dungeon story generator
func NewCrossDungeonGenerator() *CrossDungeonGenerator {
	return &CrossDungeonGenerator{}
}

// Generate creates a cross-dungeon narrative
func (g *CrossDungeonGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		log.WithFields(log.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Error("invalid difficulty parameter for cross-dungeon story generation")
		return nil, fmt.Errorf("difficulty must be between 0 and 1, got %.2f", params.Difficulty)
	}

	if params.Depth < 1 {
		log.WithFields(log.Fields{
			"seed":  seed,
			"depth": params.Depth,
		}).Error("invalid depth parameter for cross-dungeon story generation")
		return nil, fmt.Errorf("depth must be at least 1, got %d", params.Depth)
	}

	log.WithFields(log.Fields{
		"seed":  seed,
		"genre": params.GenreID,
		"depth": params.Depth,
	}).Debug("generating cross-dungeon story")

	rng := rand.New(rand.NewSource(seed))

	// Determine level span (2-5 levels based on difficulty)
	levelSpan := 2 + int(params.Difficulty*3)
	if levelSpan > 5 {
		levelSpan = 5
	}

	// Generate story theme
	fragGen := NewFragmentGenerator()
	theme := fragGen.selectTheme(rng, params.GenreID)
	seriesID := fmt.Sprintf("%s-cross-%d", theme, seed)

	// Determine depth range
	minDepth := params.Depth
	maxDepth := params.Depth + levelSpan - 1

	// Generate fragments distributed across levels (2-3 per level)
	fragmentsPerLevel := 2 + rng.Intn(2)
	totalFragments := fragmentsPerLevel * levelSpan
	fragments := make([]CrossDungeonFragment, totalFragments)

	for i := 0; i < totalFragments; i++ {
		levelIndex := i / fragmentsPerLevel
		depth := minDepth + levelIndex

		fragments[i] = g.generateCrossDungeonFragment(rng, theme, params.GenreID, i, totalFragments, depth, levelSpan)
	}

	// Set up prerequisites (each level requires previous level's fragments)
	g.setupPrerequisites(fragments, fragmentsPerLevel)

	// Calculate completion bonus (scales with span and depth)
	completionBonus := 100.0 + float64(levelSpan)*50.0 + float64(params.Depth)*25.0

	story := &CrossDungeonStory{
		SeriesID:        seriesID,
		Title:           g.generateCrossDungeonTitle(rng, theme, params.GenreID, levelSpan),
		Genre:           params.GenreID,
		Theme:           theme,
		Fragments:       fragments,
		MinDepth:        minDepth,
		MaxDepth:        maxDepth,
		LevelSpan:       levelSpan,
		CompletionBonus: completionBonus,
		Coherence:       0.65 + rng.Float64()*0.25, // 0.65-0.9
		Continuity:      g.calculateContinuity(fragments),
	}

	return story, nil
}

// Validate checks cross-dungeon story quality
func (g *CrossDungeonGenerator) Validate(result interface{}) error {
	story, ok := result.(*CrossDungeonStory)
	if !ok {
		return fmt.Errorf("result is not a *CrossDungeonStory")
	}

	if err := g.validateBasicFields(story); err != nil {
		return err
	}

	if err := g.validateDepthRange(story); err != nil {
		return err
	}

	if err := g.validateFragments(story); err != nil {
		return err
	}

	return nil
}

// validateBasicFields checks title, level span, fragment count, coherence and continuity.
func (g *CrossDungeonGenerator) validateBasicFields(story *CrossDungeonStory) error {
	if story.Title == "" {
		return fmt.Errorf("story title is empty")
	}

	if story.LevelSpan < 2 {
		return fmt.Errorf("level span too small: %d, minimum 2", story.LevelSpan)
	}

	if story.LevelSpan > 5 {
		return fmt.Errorf("level span too large: %d, maximum 5", story.LevelSpan)
	}

	if len(story.Fragments) < story.LevelSpan*2 {
		return fmt.Errorf("too few fragments: %d for %d levels", len(story.Fragments), story.LevelSpan)
	}

	if story.Coherence < 0.5 {
		return fmt.Errorf("story coherence too low: %.2f, minimum 0.5", story.Coherence)
	}

	if story.Continuity < 0.5 {
		return fmt.Errorf("story continuity too low: %.2f, minimum 0.5", story.Continuity)
	}

	return nil
}

// validateDepthRange checks minimum and maximum depth values.
func (g *CrossDungeonGenerator) validateDepthRange(story *CrossDungeonStory) error {
	if story.MinDepth < 1 {
		return fmt.Errorf("minimum depth must be at least 1, got %d", story.MinDepth)
	}

	if story.MaxDepth <= story.MinDepth {
		return fmt.Errorf("maximum depth (%d) must be greater than minimum depth (%d)", story.MaxDepth, story.MinDepth)
	}

	return nil
}

// validateFragments ensures all fragments have content and valid depth.
func (g *CrossDungeonGenerator) validateFragments(story *CrossDungeonStory) error {
	for i, frag := range story.Fragments {
		if frag.Content == "" {
			return fmt.Errorf("fragment %d has empty content", i)
		}
		if frag.Level.Depth < story.MinDepth || frag.Level.Depth > story.MaxDepth {
			return fmt.Errorf("fragment %d has depth %d outside range [%d, %d]", i, frag.Level.Depth, story.MinDepth, story.MaxDepth)
		}
	}

	return nil
}

// generateCrossDungeonFragment creates a fragment with level information
func (g *CrossDungeonGenerator) generateCrossDungeonFragment(rng *rand.Rand, theme, genreID string, index, total, depth, levelSpan int) CrossDungeonFragment {
	fragGen := NewFragmentGenerator()

	// Determine position in overall story
	progress := float64(index) / float64(total)

	// Generate content based on progress and depth
	content := g.generateCrossDungeonContent(rng, theme, genreID, progress, depth, index)

	fragType := fragGen.selectFragmentType(rng, genreID, index, total)

	// Mark key fragments as level-required
	required := (index%(total/levelSpan) == 0) // First fragment of each level is required

	// Some fragments unlock secrets (10% chance for non-first fragments)
	unlocksSecret := !required && rng.Float64() < 0.1

	return CrossDungeonFragment{
		StoryFragment: StoryFragment{
			Type:          fragType,
			Content:       content,
			Location:      fragGen.generateLocation(rng, index, total),
			DiscoveryXP:   20.0 + float64(depth)*10.0 + float64(index)*5.0,
			SeriesID:      fmt.Sprintf("%s-level-%d", theme, depth),
			SequenceNum:   index,
			SpritePattern: fragGen.generateSpritePattern(fragType, genreID, rng),
		},
		Level: DungeonLevel{
			Depth:    depth,
			ZoneName: g.generateZoneName(rng, genreID, depth),
			Required: required,
		},
		Prerequisite:  []int{}, // Set later by setupPrerequisites
		UnlocksSecret: unlocksSecret,
	}
}

// generateCrossDungeonContent creates content for cross-dungeon fragments
func (g *CrossDungeonGenerator) generateCrossDungeonContent(rng *rand.Rand, theme, genreID string, progress float64, depth, index int) string {
	if progress < 0.2 {
		// Beginning (level 1-2): Introduction
		templates := []string{
			"Level %d Entry: The journey into the depths begins. We are unprepared for what lies below.",
			"Descent Log %d: Each level reveals more of the truth. The story spans multiple floors.",
			"Floor %d: Found evidence this mystery extends deeper than expected.",
		}
		template := templates[rng.Intn(len(templates))]
		return fmt.Sprintf(template, depth, depth, depth)

	} else if progress < 0.5 {
		// Middle (level 2-3): Discovery
		templates := []string{
			"Level %d: The pieces are coming together. Previous floors' clues make sense now.",
			"Floor %d Discovery: What we found above connects to this level's secrets.",
			"Depth %d: Each level adds another chapter to this unfolding tale.",
		}
		template := templates[rng.Intn(len(templates))]
		return fmt.Sprintf(template, depth, depth, depth)

	} else if progress < 0.8 {
		// Late middle (level 3-4): Escalation
		templates := []string{
			"Level %d Warning: The danger increases with depth. All levels are connected.",
			"Floor %d: We should have turned back levels ago. Too late now.",
			"Depth %d: The final truth waits below. All clues point downward.",
		}
		template := templates[rng.Intn(len(templates))]
		return fmt.Sprintf(template, depth, depth, depth)

	} else {
		// End (level 4-5): Conclusion
		templates := []string{
			"Final Level %d: This is where it all led. The story ends here, %d levels down.",
			"Bottom Floor %d: We descended through %d levels to find this. Was it worth it?",
			"Deepest Point: Level %d marks the conclusion of a tale that began %d floors above.",
		}
		template := templates[rng.Intn(len(templates))]
		levelSpan := depth - (depth - int(progress*5))
		return fmt.Sprintf(template, depth, levelSpan, depth, levelSpan, depth, levelSpan)
	}
}

// setupPrerequisites establishes fragment dependencies across levels
func (g *CrossDungeonGenerator) setupPrerequisites(fragments []CrossDungeonFragment, fragmentsPerLevel int) {
	for i := range fragments {
		if i >= fragmentsPerLevel {
			// Require all fragments from previous level
			levelIndex := i / fragmentsPerLevel
			prevLevelStart := (levelIndex - 1) * fragmentsPerLevel
			prevLevelEnd := prevLevelStart + fragmentsPerLevel

			prerequisites := make([]int, 0, fragmentsPerLevel)
			for j := prevLevelStart; j < prevLevelEnd && j < len(fragments); j++ {
				prerequisites = append(prerequisites, j)
			}
			fragments[i].Prerequisite = prerequisites
		}
	}
}

// generateZoneName creates a zone identifier for a depth level
func (g *CrossDungeonGenerator) generateZoneName(rng *rand.Rand, genreID string, depth int) string {
	prefixes := g.getZonePrefixes(genreID)
	suffixes := []string{"Depths", "Chamber", "Hall", "Vault", "Sanctum", "Ruins", "Caverns", "Catacombs"}

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s - Level %d", prefix, suffix, depth)
}

// getZonePrefixes returns genre-specific zone name prefixes
func (g *CrossDungeonGenerator) getZonePrefixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Ancient", "Forgotten", "Cursed", "Sacred", "Dark", "Lost"}
	case "scifi":
		return []string{"Sector", "Module", "Deck", "Bay", "Laboratory", "Station"}
	case "horror":
		return []string{"Haunted", "Cursed", "Twisted", "Forsaken", "Damned", "Corrupted"}
	case "cyberpunk":
		return []string{"Sub-Level", "Underground", "Lower", "Deep", "Hidden", "Secret"}
	case "postapocalyptic":
		return []string{"Ruined", "Abandoned", "Desolate", "Collapsed", "Irradiated", "Dead"}
	default:
		return []string{"Lower", "Deep", "Unknown", "Hidden", "Forgotten", "Lost"}
	}
}

// generateCrossDungeonTitle creates a title for cross-dungeon story
func (g *CrossDungeonGenerator) generateCrossDungeonTitle(rng *rand.Rand, theme, genreID string, levelSpan int) string {
	prefixes := []string{"The", "Mystery of the", "Descent into", "Chronicles of", "Saga of the"}
	middles := []string{"Depths", "Levels", "Floors", "Layers", "Tiers"}
	suffixes := g.getCrossDungeonTitleSuffixes(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	middle := middles[rng.Intn(len(middles))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %d %s of %s", prefix, levelSpan, middle, suffix)
}

// getCrossDungeonTitleSuffixes returns genre-specific title endings
func (g *CrossDungeonGenerator) getCrossDungeonTitleSuffixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Darkness", "the Abyss", "Lost Magic", "Ancient Power", "the Underworld"}
	case "scifi":
		return []string{"the Station", "Alien Tech", "the Facility", "Deep Space", "the Colony"}
	case "horror":
		return []string{"Madness", "the Damned", "Terror", "Nightmares", "the Void"}
	case "cyberpunk":
		return []string{"the Grid", "Data", "the Network", "Secrets", "the System"}
	case "postapocalyptic":
		return []string{"Ruin", "the Wasteland", "Survival", "the Dead World", "Despair"}
	default:
		return []string{"Mystery", "the Unknown", "Secrets", "the Past", "Truth"}
	}
}

// calculateContinuity measures how well the story flows across levels
func (g *CrossDungeonGenerator) calculateContinuity(fragments []CrossDungeonFragment) float64 {
	if len(fragments) < 2 {
		return 0.5
	}

	// Check prerequisite connections
	connectedFragments := 0
	totalConnections := 0

	for _, frag := range fragments {
		if len(frag.Prerequisite) > 0 {
			totalConnections++
			// If prerequisites exist, story has continuity
			connectedFragments++
		}
	}

	if totalConnections == 0 {
		return 0.6 // Base continuity for single-level stories
	}

	continuity := float64(connectedFragments) / float64(totalConnections)

	// Normalize to 0.5-1.0 range
	return 0.5 + continuity*0.5
}

// IsFragmentAccessible checks if a fragment's prerequisites are met
func (s *CrossDungeonStory) IsFragmentAccessible(fragmentIndex int, discoveredFragments map[int]bool) bool {
	if fragmentIndex < 0 || fragmentIndex >= len(s.Fragments) {
		return false
	}

	frag := s.Fragments[fragmentIndex]

	// If no prerequisites, always accessible
	if len(frag.Prerequisite) == 0 {
		return true
	}

	// Check all prerequisites are discovered
	for _, prereqIndex := range frag.Prerequisite {
		if !discoveredFragments[prereqIndex] {
			return false
		}
	}

	return true
}

// GetFragmentsForLevel returns all fragments for a specific dungeon level
func (s *CrossDungeonStory) GetFragmentsForLevel(depth int) []CrossDungeonFragment {
	result := make([]CrossDungeonFragment, 0)

	for _, frag := range s.Fragments {
		if frag.Level.Depth == depth {
			result = append(result, frag)
		}
	}

	return result
}

// GetRequiredLevels returns all dungeon levels that must be visited
func (s *CrossDungeonStory) GetRequiredLevels() []int {
	levels := make(map[int]bool)

	for _, frag := range s.Fragments {
		if frag.Level.Required {
			levels[frag.Level.Depth] = true
		}
	}

	result := make([]int, 0, len(levels))
	for level := range levels {
		result = append(result, level)
	}

	return result
}
