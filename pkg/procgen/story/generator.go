package story

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// Sentinel errors for validation failures
var (
	ErrInvalidDifficulty    = errors.New("difficulty must be between 0 and 1")
	ErrInvalidType          = errors.New("result type mismatch")
	ErrEmptyTitle           = errors.New("story title is empty")
	ErrTooFewFragments      = errors.New("too few fragments")
	ErrTooManyFragments     = errors.New("too many fragments")
	ErrLowCoherence         = errors.New("story coherence too low")
	ErrEmptyFragmentContent = errors.New("fragment has empty content")
	ErrShortFragmentContent = errors.New("fragment content too short")
	ErrEmptySiteName        = errors.New("site name is empty")
	ErrTooFewArtifacts      = errors.New("too few artifacts")
	ErrTooManyArtifacts     = errors.New("too many artifacts")
	ErrInvalidDanger        = errors.New("danger must be between 0 and 1")
	ErrEmptyArtifactName    = errors.New("artifact has empty name")
	ErrArtifactCondition    = errors.New("artifact condition out of range")
	ErrNoChoicePoints       = errors.New("no choice points in branching narrative")
	ErrTooManyChoicePoints  = errors.New("too many choice points")
	ErrTooFewPaths          = errors.New("too few paths")
	ErrTooManyPaths         = errors.New("too many paths")
	ErrNoCommonFragments    = errors.New("no common fragments")
	ErrPathTooFewFragments  = errors.New("path has too few fragments")
	ErrPathNoOutcome        = errors.New("path has no outcome")
	ErrInvalidChoiceIndex   = errors.New("invalid choice index")
	ErrInvalidOptionIndex   = errors.New("invalid option index")
)

// String returns the string representation of FragmentType
func (f FragmentType) String() string {
	switch f {
	case FragmentNote:
		return "Note"
	case FragmentCarving:
		return "Carving"
	case FragmentCorpse:
		return "Corpse"
	case FragmentRelic:
		return "Relic"
	case FragmentGraffiti:
		return "Graffiti"
	case FragmentBlood:
		return "Blood"
	default:
		return "Unknown"
	}
}

// StoryFragment represents a single discoverable story element
type StoryFragment struct {
	Type          FragmentType
	Content       string
	Location      Vector2
	DiscoveryXP   float64
	SeriesID      string // Related fragments share series ID
	SequenceNum   int    // Order within series (0-based)
	SpritePattern string // Visual representation pattern for rendering
}

// StorySequence represents a complete story told through fragments
type StorySequence struct {
	SeriesID  string
	Title     string
	Genre     string
	Fragments []StoryFragment
	Theme     string  // Main theme: tragedy, mystery, horror, adventure, etc.
	Coherence float64 // Quality metric (0.0-1.0)
}

// StoryTemplates holds configurable templates for story content generation
type StoryTemplates struct {
	BeginningTemplates   []string
	BeginningAdjectives  []string
	BeginningCounts      []int
	BeginningDiscoveries []string

	MiddleTemplates   []string
	MiddleLosses      []string
	MiddleThreats     []string
	MiddleRevelations []string
	MiddleAttackers   []string
	MiddleSurvivors   []int

	EndTemplates []string
	EndWarnings  []string
	EndGoals     []string
	EndThreats   []string
	EndFates     []string
	EndMessages  []string
}

// DefaultStoryTemplates returns the default template set
func DefaultStoryTemplates() *StoryTemplates {
	return &StoryTemplates{
		BeginningTemplates: []string{
			"Day %d: We arrived at this place. Everything seems %s.",
			"Entry %d: The expedition begins. We are %d strong.",
			"Log %d: First signs of %s. Should we continue?",
			"Note %d: Found evidence of %s. This changes everything.",
		},
		BeginningAdjectives:  []string{"normal", "quiet", "strange", "peaceful", "ominous"},
		BeginningCounts:      []int{5, 7, 10, 12, 15},
		BeginningDiscoveries: []string{"danger", "treasure", "secrets", "life", "death"},

		MiddleTemplates: []string{
			"Day %d: Things are getting worse. We lost %s today.",
			"Entry %d: The %s is spreading. No one is safe.",
			"Log %d: Discovered the truth about %s. We were wrong.",
			"Note %d: %s attacked us. Only %d survived.",
		},
		MiddleLosses:      []string{"contact", "hope", "supplies", "three people", "our leader"},
		MiddleThreats:     []string{"infection", "madness", "corruption", "fear", "darkness"},
		MiddleRevelations: []string{"the ruins", "the source", "their plan", "the curse", "this place"},
		MiddleAttackers:   []string{"They", "The creatures", "Something", "Unknown forces", "The enemy"},
		MiddleSurvivors:   []int{3, 5, 7, 4, 2},

		EndTemplates: []string{
			"Final entry: If you find this, %s. Don't make our mistakes.",
			"Last words: We failed to %s. May you succeed where we couldn't.",
			"Warning: %s is coming. Run while you still can.",
			"Goodbye: We're %s. Tell our families %s.",
		},
		EndWarnings: []string{"leave immediately", "destroy it", "seal the entrance", "warn the others"},
		EndGoals:    []string{"stop it", "find the cure", "escape", "understand the truth"},
		EndThreats:  []string{"The end", "Darkness", "They", "Death", "Doom"},
		EndFates:    []string{"trapped", "infected", "lost", "dying", "gone"},
		EndMessages: []string{"we tried", "we loved them", "we're sorry", "it wasn't their fault", "goodbye"},
	}
}

// FragmentGenerator generates environmental story fragments
type FragmentGenerator struct {
	templates *StoryTemplates
}

// NewFragmentGenerator creates a new fragment generator with default templates
func NewFragmentGenerator() *FragmentGenerator {
	return &FragmentGenerator{
		templates: DefaultStoryTemplates(),
	}
}

// NewFragmentGeneratorWithTemplates creates a new fragment generator with custom templates
func NewFragmentGeneratorWithTemplates(templates *StoryTemplates) *FragmentGenerator {
	if templates == nil {
		templates = DefaultStoryTemplates()
	}
	return &FragmentGenerator{
		templates: templates,
	}
}

// Generate creates a story sequence with fragments
func (g *FragmentGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		log.WithFields(log.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Error("invalid difficulty parameter for story generation")
		return nil, fmt.Errorf("%w, got %.2f", ErrInvalidDifficulty, params.Difficulty)
	}

	log.WithFields(log.Fields{
		"seed":       seed,
		"genre":      params.GenreID,
		"difficulty": params.Difficulty,
		"depth":      params.Depth,
	}).Debug("generating story sequence")

	rng := rand.New(rand.NewSource(seed))

	// Determine number of fragments (5-15 based on depth and difficulty)
	numFragments := 5 + int(float64(params.Depth)*0.5) + int(params.Difficulty*5)
	if numFragments > 15 {
		numFragments = 15
	}

	// Generate story theme
	theme := g.selectTheme(rng, params.GenreID)

	// Generate series ID
	seriesID := fmt.Sprintf("%s-%d", theme, seed)

	// Generate story content
	storyContent := g.generateStoryContent(rng, theme, params.GenreID, numFragments)

	// Create fragments
	fragments := make([]StoryFragment, numFragments)
	for i := 0; i < numFragments; i++ {
		fragType := g.selectFragmentType(rng, params.GenreID, i, numFragments)
		content := storyContent[i]

		fragments[i] = StoryFragment{
			Type:          fragType,
			Content:       content,
			Location:      g.generateLocation(rng, i, numFragments),
			DiscoveryXP:   10.0 + float64(i)*5.0,
			SeriesID:      seriesID,
			SequenceNum:   i,
			SpritePattern: g.generateSpritePattern(fragType, params.GenreID, rng),
		}
	}

	sequence := &StorySequence{
		SeriesID:  seriesID,
		Title:     g.generateTitle(rng, theme, params.GenreID),
		Genre:     params.GenreID,
		Fragments: fragments,
		Theme:     theme,
		Coherence: g.calculateCoherence(storyContent),
	}

	log.WithFields(log.Fields{
		"series_id":     seriesID,
		"theme":         theme,
		"num_fragments": numFragments,
		"coherence":     sequence.Coherence,
	}).Info("story sequence generated")

	return sequence, nil
}

// Validate checks story sequence quality
func (g *FragmentGenerator) Validate(result interface{}) error {
	sequence, ok := result.(*StorySequence)
	if !ok {
		log.Error("validation failed: result is not a *StorySequence")
		return ErrInvalidType
	}

	if sequence.Title == "" {
		log.WithFields(log.Fields{
			"series_id": sequence.SeriesID,
		}).Warn("validation failed: story title is empty")
		return ErrEmptyTitle
	}

	if len(sequence.Fragments) < 5 {
		log.WithFields(log.Fields{
			"series_id":     sequence.SeriesID,
			"num_fragments": len(sequence.Fragments),
		}).Warn("validation failed: too few fragments")
		return fmt.Errorf("%w: %d, minimum 5", ErrTooFewFragments, len(sequence.Fragments))
	}

	if len(sequence.Fragments) > 15 {
		log.WithFields(log.Fields{
			"series_id":     sequence.SeriesID,
			"num_fragments": len(sequence.Fragments),
		}).Warn("validation failed: too many fragments")
		return fmt.Errorf("%w: %d, maximum 15", ErrTooManyFragments, len(sequence.Fragments))
	}

	if sequence.Coherence < 0.5 {
		log.WithFields(log.Fields{
			"series_id": sequence.SeriesID,
			"coherence": sequence.Coherence,
		}).Warn("validation failed: story coherence too low")
		return fmt.Errorf("%w: %.2f, minimum 0.5", ErrLowCoherence, sequence.Coherence)
	}

	// Validate all fragments have content
	for i, frag := range sequence.Fragments {
		if frag.Content == "" {
			log.WithFields(log.Fields{
				"series_id":    sequence.SeriesID,
				"fragment_num": i,
			}).Warn("validation failed: fragment has empty content")
			return fmt.Errorf("%w at fragment %d", ErrEmptyFragmentContent, i)
		}
		if len(frag.Content) < 10 {
			log.WithFields(log.Fields{
				"series_id":      sequence.SeriesID,
				"fragment_num":   i,
				"content_length": len(frag.Content),
			}).Warn("validation failed: fragment content too short")
			return fmt.Errorf("%w: fragment %d has %d chars", ErrShortFragmentContent, i, len(frag.Content))
		}
	}

	log.WithFields(log.Fields{
		"series_id":     sequence.SeriesID,
		"coherence":     sequence.Coherence,
		"num_fragments": len(sequence.Fragments),
	}).Debug("story sequence validation passed")

	return nil
}

func (g *FragmentGenerator) selectTheme(rng *rand.Rand, genreID string) string {
	themes := g.getThemesForGenre(genreID)
	return themes[rng.Intn(len(themes))]
}

func (g *FragmentGenerator) getThemesForGenre(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"ancient_curse", "fallen_kingdom", "lost_artifact", "dragon_slayer", "wizard_betrayal"}
	case "scifi":
		return []string{"abandoned_colony", "ai_uprising", "alien_contact", "experiment_gone_wrong", "time_loop"}
	case "horror":
		return []string{"haunted_mansion", "cult_ritual", "plague_outbreak", "entity_awakening", "descent_madness"}
	case "cyberpunk":
		return []string{"corporate_conspiracy", "hacker_revenge", "memory_theft", "underground_resistance", "ai_consciousness"}
	case "postapocalyptic":
		return []string{"last_survivors", "resource_war", "mutant_origin", "bunker_secrets", "radiation_zone"}
	default:
		return []string{"mystery", "tragedy", "adventure", "discovery", "conflict"}
	}
}

func (g *FragmentGenerator) generateStoryContent(rng *rand.Rand, theme, genreID string, numFragments int) []string {
	content := make([]string, numFragments)

	// Beginning (first 1/3 of fragments)
	beginCount := numFragments / 3
	for i := 0; i < beginCount; i++ {
		content[i] = g.generateBeginningFragment(rng, theme, genreID, i)
	}

	// Middle (middle 1/3)
	middleStart := beginCount
	middleEnd := (numFragments * 2) / 3
	for i := middleStart; i < middleEnd; i++ {
		content[i] = g.generateMiddleFragment(rng, theme, genreID, i-middleStart)
	}

	// End (final 1/3)
	for i := middleEnd; i < numFragments; i++ {
		content[i] = g.generateEndFragment(rng, theme, genreID, i-middleEnd)
	}

	return content
}

func (g *FragmentGenerator) generateBeginningFragment(rng *rand.Rand, theme, genreID string, index int) string {
	template := g.templates.BeginningTemplates[rng.Intn(len(g.templates.BeginningTemplates))]
	adjective := g.templates.BeginningAdjectives[rng.Intn(len(g.templates.BeginningAdjectives))]
	count := g.templates.BeginningCounts[rng.Intn(len(g.templates.BeginningCounts))]
	discovery := g.templates.BeginningDiscoveries[rng.Intn(len(g.templates.BeginningDiscoveries))]

	return fmt.Sprintf(template, index+1, adjective, count, discovery)
}

func (g *FragmentGenerator) generateMiddleFragment(rng *rand.Rand, theme, genreID string, index int) string {
	template := g.templates.MiddleTemplates[rng.Intn(len(g.templates.MiddleTemplates))]
	loss := g.templates.MiddleLosses[rng.Intn(len(g.templates.MiddleLosses))]
	threat := g.templates.MiddleThreats[rng.Intn(len(g.templates.MiddleThreats))]
	revelation := g.templates.MiddleRevelations[rng.Intn(len(g.templates.MiddleRevelations))]
	attacker := g.templates.MiddleAttackers[rng.Intn(len(g.templates.MiddleAttackers))]
	survivor := g.templates.MiddleSurvivors[rng.Intn(len(g.templates.MiddleSurvivors))]

	return fmt.Sprintf(template, index+5, loss, threat, revelation, attacker, survivor)
}

func (g *FragmentGenerator) generateEndFragment(rng *rand.Rand, theme, genreID string, index int) string {
	template := g.templates.EndTemplates[rng.Intn(len(g.templates.EndTemplates))]
	warning := g.templates.EndWarnings[rng.Intn(len(g.templates.EndWarnings))]
	goal := g.templates.EndGoals[rng.Intn(len(g.templates.EndGoals))]
	threat := g.templates.EndThreats[rng.Intn(len(g.templates.EndThreats))]
	fate := g.templates.EndFates[rng.Intn(len(g.templates.EndFates))]
	message := g.templates.EndMessages[rng.Intn(len(g.templates.EndMessages))]

	return fmt.Sprintf(template, warning, goal, threat, fate, message)
}

func (g *FragmentGenerator) selectFragmentType(rng *rand.Rand, genreID string, index, total int) FragmentType {
	// Distribution based on position in story
	progress := float64(index) / float64(total)

	if progress < 0.3 {
		// Beginning: more notes and carvings
		types := []FragmentType{FragmentNote, FragmentNote, FragmentCarving, FragmentRelic}
		return types[rng.Intn(len(types))]
	} else if progress < 0.7 {
		// Middle: mix of everything
		return FragmentType(rng.Intn(6))
	} else {
		// End: corpses, blood, graffiti
		types := []FragmentType{FragmentCorpse, FragmentBlood, FragmentGraffiti, FragmentCorpse}
		return types[rng.Intn(len(types))]
	}
}

func (g *FragmentGenerator) generateLocation(rng *rand.Rand, index, total int) Vector2 {
	// Distribute fragments throughout dungeon
	// Assuming 100x100 dungeon space
	progress := float64(index) / float64(total)

	x := 10.0 + progress*80.0 + (rng.Float64()-0.5)*20.0
	y := 10.0 + rng.Float64()*80.0

	return Vector2{X: x, Y: y}
}

func (g *FragmentGenerator) generateTitle(rng *rand.Rand, theme, genreID string) string {
	prefixes := []string{"The", "Mystery of", "Tale of", "Legend of", "Story of"}
	suffixes := g.getTitleSuffixes(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

func (g *FragmentGenerator) getTitleSuffixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Lost Kingdom", "Fallen Hero", "Ancient Curse", "Hidden Temple", "Dark Forest"}
	case "scifi":
		return []string{"Colony Alpha", "Last Signal", "Derelict Station", "Failed Experiment", "AI Core"}
	case "horror":
		return []string{"Haunted Manor", "Cursed Ground", "Dark Ritual", "Final Scream", "Blood Moon"}
	case "cyberpunk":
		return []string{"Data Heist", "Corporate Fall", "Neon Shadows", "Memory Dump", "System Breach"}
	case "postapocalyptic":
		return []string{"Last Stand", "Wasteland", "Dead Zone", "Bunker 13", "Final Days"}
	default:
		return []string{"Unknown Place", "Forgotten Story", "Lost Tale", "Hidden Truth", "Dark Secret"}
	}
}

func (g *FragmentGenerator) calculateCoherence(storyContent []string) float64 {
	if len(storyContent) == 0 {
		return 0.0
	}

	// Simple coherence metric: check for common words between fragments
	// Real implementation would use more sophisticated NLP
	commonWords := 0
	totalWords := 0

	for i := 0; i < len(storyContent)-1; i++ {
		words1 := strings.Fields(strings.ToLower(storyContent[i]))
		words2 := strings.Fields(strings.ToLower(storyContent[i+1]))

		totalWords += len(words1) + len(words2)

		// Count common words
		for _, w1 := range words1 {
			for _, w2 := range words2 {
				if w1 == w2 && len(w1) > 3 { // Only count words longer than 3 chars
					commonWords++
					break
				}
			}
		}
	}

	if totalWords == 0 {
		return 0.5 // Default coherence
	}

	coherence := float64(commonWords) / float64(totalWords)

	// Normalize to 0.5-1.0 range (stories always have some coherence)
	return 0.5 + coherence*0.5
}

// generateSpritePattern generates a visual pattern description for rendering.
// This returns a simple string identifier that can be used by the sprite
// generator to create appropriate visuals for each fragment type.
func (g *FragmentGenerator) generateSpritePattern(fragType FragmentType, genreID string, rng *rand.Rand) string {
	switch fragType {
	case FragmentNote:
		// Paper/scroll patterns
		patterns := []string{"scroll", "paper", "parchment", "journal"}
		return patterns[rng.Intn(len(patterns))]

	case FragmentCarving:
		// Wall inscription patterns
		patterns := []string{"wall_runes", "stone_script", "wall_etching", "carved_text"}
		return patterns[rng.Intn(len(patterns))]

	case FragmentCorpse:
		// Body/skeleton patterns based on genre
		if genreID == "scifi" {
			patterns := []string{"android_remains", "human_corpse", "alien_body"}
			return patterns[rng.Intn(len(patterns))]
		} else if genreID == "horror" {
			patterns := []string{"twisted_corpse", "skeletal_remains", "decayed_body"}
			return patterns[rng.Intn(len(patterns))]
		}
		patterns := []string{"skeleton", "body", "remains"}
		return patterns[rng.Intn(len(patterns))]

	case FragmentRelic:
		// Artifact patterns based on genre
		if genreID == "fantasy" {
			patterns := []string{"ancient_amulet", "magic_relic", "enchanted_item"}
			return patterns[rng.Intn(len(patterns))]
		} else if genreID == "scifi" {
			patterns := []string{"tech_artifact", "data_crystal", "alien_device"}
			return patterns[rng.Intn(len(patterns))]
		}
		patterns := []string{"artifact", "relic", "ancient_object"}
		return patterns[rng.Intn(len(patterns))]

	case FragmentGraffiti:
		// Markings based on genre
		if genreID == "cyberpunk" {
			patterns := []string{"neon_tag", "spray_paint", "holo_graffiti"}
			return patterns[rng.Intn(len(patterns))]
		}
		patterns := []string{"graffiti", "wall_marking", "painted_message"}
		return patterns[rng.Intn(len(patterns))]

	case FragmentBlood:
		// Blood/fluid trails
		patterns := []string{"blood_trail", "blood_splatter", "blood_pool"}
		return patterns[rng.Intn(len(patterns))]

	default:
		return "unknown_fragment"
	}
}
