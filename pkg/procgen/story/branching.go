package story

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// ChoicePoint represents a branching decision in the narrative
type ChoicePoint struct {
	FragmentIndex int      // Which fragment contains the choice
	Description   string   // What the choice represents
	Options       []string // Available narrative branches
	Chosen        int      // Which option was chosen (-1 if not yet decided)
}

// NarrativePath represents one possible storyline through the dungeon
type NarrativePath struct {
	PathID      string          // Unique identifier for this path
	Title       string          // Path-specific title
	Fragments   []StoryFragment // Fragments specific to this path
	ChoicesMade []int           // Indices of choices made to reach this path
	Outcome     string          // How this path concludes
}

// BranchingNarrative represents a story with multiple possible paths
type BranchingNarrative struct {
	SeriesID     string          // Base series identifier
	Genre        string          // Genre of the narrative
	Theme        string          // Overall theme
	ChoicePoints []ChoicePoint   // Decision points in the narrative
	Paths        []NarrativePath // All possible story paths
	ActivePathID string          // Currently active path (empty if undecided)
	CommonFrags  []StoryFragment // Fragments shared by all paths
	Coherence    float64         // Overall story quality (0.0-1.0)
}

// BranchingNarrativeGenerator creates stories with multiple paths
type BranchingNarrativeGenerator struct{}

// NewBranchingNarrativeGenerator creates a new branching narrative generator
func NewBranchingNarrativeGenerator() *BranchingNarrativeGenerator {
	return &BranchingNarrativeGenerator{}
}

// Generate creates a branching narrative with choice points
func (g *BranchingNarrativeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		log.WithFields(log.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Error("invalid difficulty parameter for branching narrative generation")
		return nil, fmt.Errorf("%w, got %.2f", ErrInvalidDifficulty, params.Difficulty)
	}

	log.WithFields(log.Fields{
		"seed":  seed,
		"genre": params.GenreID,
		"depth": params.Depth,
	}).Debug("generating branching narrative")

	rng := rand.New(rand.NewSource(seed))

	// Determine number of choice points (1-3 based on depth)
	numChoices := 1 + int(float64(params.Depth)*0.1)
	if numChoices > 3 {
		numChoices = 3
	}

	// Generate story theme
	fragGen := NewFragmentGenerator()
	theme := fragGen.selectTheme(rng, params.GenreID)
	seriesID := fmt.Sprintf("%s-branch-%d", theme, seed)

	// Create common fragments (introduction, always present)
	commonFrags := g.generateCommonFragments(rng, theme, params.GenreID, 2+numChoices)

	// Generate choice points
	choicePoints := g.generateChoicePoints(rng, theme, params.GenreID, numChoices, len(commonFrags))

	// Generate paths (2^numChoices possible paths)
	numPaths := 1 << uint(numChoices) // 2^n paths
	if numPaths > 8 {
		numPaths = 8 // Cap at 8 paths for sanity
	}

	paths := make([]NarrativePath, numPaths)
	for i := 0; i < numPaths; i++ {
		paths[i] = g.generatePath(rng, seriesID, theme, params.GenreID, i, numChoices, choicePoints)
	}

	narrative := &BranchingNarrative{
		SeriesID:     seriesID,
		Genre:        params.GenreID,
		Theme:        theme,
		ChoicePoints: choicePoints,
		Paths:        paths,
		ActivePathID: "", // Player hasn't chosen yet
		CommonFrags:  commonFrags,
		Coherence:    0.6 + rng.Float64()*0.3, // 0.6-0.9 range
	}

	return narrative, nil
}

// Validate checks branching narrative quality
func (g *BranchingNarrativeGenerator) Validate(result interface{}) error {
	narrative, ok := result.(*BranchingNarrative)
	if !ok {
		return ErrInvalidType
	}

	if err := validateChoicePoints(narrative); err != nil {
		return err
	}

	if err := validatePaths(narrative); err != nil {
		return err
	}

	if err := validateCommonFragments(narrative); err != nil {
		return err
	}

	if err := validateCoherence(narrative); err != nil {
		return err
	}

	return validatePathContents(narrative)
}

// validateChoicePoints checks that narrative has valid number of choice points.
func validateChoicePoints(narrative *BranchingNarrative) error {
	if len(narrative.ChoicePoints) < 1 {
		return ErrNoChoicePoints
	}
	if len(narrative.ChoicePoints) > 3 {
		return fmt.Errorf("%w: %d, maximum 3", ErrTooManyChoicePoints, len(narrative.ChoicePoints))
	}
	return nil
}

// validatePaths checks that narrative has valid number of paths.
func validatePaths(narrative *BranchingNarrative) error {
	if len(narrative.Paths) < 2 {
		return fmt.Errorf("%w: %d, minimum 2", ErrTooFewPaths, len(narrative.Paths))
	}
	if len(narrative.Paths) > 8 {
		return fmt.Errorf("%w: %d, maximum 8", ErrTooManyPaths, len(narrative.Paths))
	}
	return nil
}

// validateCommonFragments checks that narrative has common fragments.
func validateCommonFragments(narrative *BranchingNarrative) error {
	if len(narrative.CommonFrags) < 1 {
		return ErrNoCommonFragments
	}
	return nil
}

// validateCoherence checks that narrative has acceptable coherence score.
func validateCoherence(narrative *BranchingNarrative) error {
	if narrative.Coherence < 0.5 {
		return fmt.Errorf("%w: %.2f, minimum 0.5", ErrLowCoherence, narrative.Coherence)
	}
	return nil
}

// validatePathContents checks that all paths have valid fragments and outcomes.
func validatePathContents(narrative *BranchingNarrative) error {
	for i, path := range narrative.Paths {
		if len(path.Fragments) < 2 {
			return fmt.Errorf("%w: path %d has %d fragments", ErrPathTooFewFragments, i, len(path.Fragments))
		}
		if path.Outcome == "" {
			return fmt.Errorf("%w: path %d", ErrPathNoOutcome, i)
		}
	}
	return nil
}

// generateCommonFragments creates fragments shared by all paths
func (g *BranchingNarrativeGenerator) generateCommonFragments(rng *rand.Rand, theme, genreID string, count int) []StoryFragment {
	fragGen := NewFragmentGenerator()
	fragments := make([]StoryFragment, count)

	for i := 0; i < count; i++ {
		content := g.generateIntroContent(rng, theme, genreID, i)
		fragType := fragGen.selectFragmentType(rng, genreID, i, count*2)

		fragments[i] = StoryFragment{
			Type:          fragType,
			Content:       content,
			Location:      fragGen.generateLocation(rng, i, count*2),
			DiscoveryXP:   10.0 + float64(i)*5.0,
			SeriesID:      fmt.Sprintf("%s-common", theme),
			SequenceNum:   i,
			SpritePattern: fragGen.generateSpritePattern(fragType, genreID, rng),
		}
	}

	return fragments
}

// generateIntroContent creates introduction narrative
func (g *BranchingNarrativeGenerator) generateIntroContent(rng *rand.Rand, theme, genreID string, index int) string {
	templates := []string{
		"Entry %d: We have arrived. The decision point approaches.",
		"Day %d: Everything seems normal, but a choice must be made soon.",
		"Log %d: Initial survey complete. We must decide our course.",
		"Note %d: Found evidence of previous expeditions. They faced the same choice.",
	}

	template := templates[rng.Intn(len(templates))]
	return fmt.Sprintf(template, index+1)
}

// generateChoicePoints creates decision points in the narrative
func (g *BranchingNarrativeGenerator) generateChoicePoints(rng *rand.Rand, theme, genreID string, count, startIndex int) []ChoicePoint {
	choices := make([]ChoicePoint, count)

	for i := 0; i < count; i++ {
		choices[i] = ChoicePoint{
			FragmentIndex: startIndex + i,
			Description:   g.generateChoiceDescription(rng, theme, genreID, i),
			Options:       g.generateChoiceOptions(rng, theme, genreID, i),
			Chosen:        -1, // Not yet chosen
		}
	}

	return choices
}

// generateChoiceDescription creates a description for a choice point
func (g *BranchingNarrativeGenerator) generateChoiceDescription(rng *rand.Rand, theme, genreID string, choiceNum int) string {
	descriptions := []string{
		"The path splits here. Which way?",
		"A critical decision must be made.",
		"Two doors stand before us. Choose wisely.",
		"The team is divided. We must decide.",
	}

	return descriptions[rng.Intn(len(descriptions))]
}

// generateChoiceOptions creates options for a choice point
func (g *BranchingNarrativeGenerator) generateChoiceOptions(rng *rand.Rand, theme, genreID string, choiceNum int) []string {
	optionSets := [][]string{
		{"Take the left path", "Take the right path"},
		{"Proceed cautiously", "Rush forward"},
		{"Investigate thoroughly", "Move quickly"},
		{"Split the group", "Stay together"},
	}

	return optionSets[rng.Intn(len(optionSets))]
}

// generatePath creates a single narrative path based on choices
func (g *BranchingNarrativeGenerator) generatePath(rng *rand.Rand, seriesID, theme, genreID string, pathIndex, numChoices int, choices []ChoicePoint) NarrativePath {
	// Decode path index into binary choices (0 or 1 for each choice point)
	choicesMade := make([]int, numChoices)
	for i := 0; i < numChoices; i++ {
		choicesMade[i] = (pathIndex >> uint(i)) & 1
	}

	// Generate 3-5 fragments for this path
	numFragments := 3 + rng.Intn(3)
	fragments := make([]StoryFragment, numFragments)

	fragGen := NewFragmentGenerator()

	for i := 0; i < numFragments; i++ {
		content := g.generatePathContent(rng, theme, genreID, pathIndex, i, numFragments, choicesMade)
		fragType := fragGen.selectFragmentType(rng, genreID, i+len(choices), numFragments*2)

		fragments[i] = StoryFragment{
			Type:          fragType,
			Content:       content,
			Location:      fragGen.generateLocation(rng, i+10, numFragments*2), // Offset location
			DiscoveryXP:   15.0 + float64(i)*7.0,
			SeriesID:      fmt.Sprintf("%s-path-%d", seriesID, pathIndex),
			SequenceNum:   i,
			SpritePattern: fragGen.generateSpritePattern(fragType, genreID, rng),
		}
	}

	outcome := g.generateOutcome(rng, theme, genreID, choicesMade)

	return NarrativePath{
		PathID:      fmt.Sprintf("%s-path-%d", seriesID, pathIndex),
		Title:       g.generatePathTitle(rng, theme, genreID, pathIndex),
		Fragments:   fragments,
		ChoicesMade: choicesMade,
		Outcome:     outcome,
	}
}

// generatePathContent creates content for a path-specific fragment
func (g *BranchingNarrativeGenerator) generatePathContent(rng *rand.Rand, theme, genreID string, pathIndex, fragIndex, total int, choices []int) string {
	progress := float64(fragIndex) / float64(total)

	if progress < 0.5 {
		// Early in path: consequences of choice
		templates := []string{
			"We chose path %d. So far it seems %s.",
			"Following the %s route. No turning back now.",
			"Decision %d led us here. Was it right?",
		}
		template := templates[rng.Intn(len(templates))]
		adjectives := []string{"promising", "dangerous", "uncertain", "rewarding"}
		routes := []string{"left", "right", "cautious", "bold"}

		return fmt.Sprintf(template, pathIndex, adjectives[rng.Intn(len(adjectives))], routes[rng.Intn(len(routes))], pathIndex)
	} else {
		// Later in path: outcomes
		templates := []string{
			"The path leads to %s. We are %s.",
			"Our choice resulted in %s. %s.",
			"Final outcome: %s. The price was %s.",
		}
		template := templates[rng.Intn(len(templates))]
		outcomes := []string{"salvation", "doom", "knowledge", "loss", "victory"}
		states := []string{"relieved", "terrified", "enlightened", "broken", "triumphant"}
		prices := []string{"high", "steep", "acceptable", "devastating", "worth it"}

		return fmt.Sprintf(template, outcomes[rng.Intn(len(outcomes))], states[rng.Intn(len(states))], outcomes[rng.Intn(len(outcomes))], states[rng.Intn(len(states))], outcomes[rng.Intn(len(outcomes))], prices[rng.Intn(len(prices))])
	}
}

// generateOutcome creates the final outcome for a path
func (g *BranchingNarrativeGenerator) generateOutcome(rng *rand.Rand, theme, genreID string, choices []int) string {
	// Generate outcome based on number of "0" vs "1" choices
	optimisticChoices := 0
	for _, c := range choices {
		if c == 0 {
			optimisticChoices++
		}
	}

	ratio := float64(optimisticChoices) / float64(len(choices))

	if ratio > 0.6 {
		// Mostly optimistic choices
		outcomes := []string{
			"The expedition succeeded against all odds.",
			"We found what we sought and returned safely.",
			"Our caution paid off. Victory is ours.",
		}
		return outcomes[rng.Intn(len(outcomes))]
	} else if ratio < 0.4 {
		// Mostly pessimistic choices
		outcomes := []string{
			"The expedition ended in tragedy.",
			"We paid the ultimate price for our boldness.",
			"Rushing forward led to our doom.",
		}
		return outcomes[rng.Intn(len(outcomes))]
	} else {
		// Mixed
		outcomes := []string{
			"We survived, but at great cost.",
			"The outcome was neither victory nor defeat.",
			"Some lived to tell the tale. Others were not so lucky.",
		}
		return outcomes[rng.Intn(len(outcomes))]
	}
}

// generatePathTitle creates a title for a narrative path
func (g *BranchingNarrativeGenerator) generatePathTitle(rng *rand.Rand, theme, genreID string, pathIndex int) string {
	prefixes := []string{"The Path of", "Way of", "Route to", "Journey to"}
	suffixes := []string{"Triumph", "Doom", "Knowledge", "Sacrifice", "Glory", "Ruin", "Salvation", "Despair"}

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

// MakeChoice records a player's choice in the narrative
func (n *BranchingNarrative) MakeChoice(choiceIndex, optionIndex int) error {
	if choiceIndex < 0 || choiceIndex >= len(n.ChoicePoints) {
		return fmt.Errorf("invalid choice index: %d", choiceIndex)
	}

	choice := &n.ChoicePoints[choiceIndex]
	if optionIndex < 0 || optionIndex >= len(choice.Options) {
		return fmt.Errorf("invalid option index: %d for choice %d", optionIndex, choiceIndex)
	}

	choice.Chosen = optionIndex
	n.updateActivePath()
	return nil
}

// updateActivePath determines which path is active based on choices made
func (n *BranchingNarrative) updateActivePath() {
	// Build binary pattern from choices
	pathIndex := 0
	for i, choice := range n.ChoicePoints {
		if choice.Chosen >= 0 {
			pathIndex |= (choice.Chosen << uint(i))
		} else {
			// Not all choices made yet
			n.ActivePathID = ""
			return
		}
	}

	// All choices made, activate corresponding path
	if pathIndex < len(n.Paths) {
		n.ActivePathID = n.Paths[pathIndex].PathID
	}
}

// GetActiveFragments returns all fragments for the current path
func (n *BranchingNarrative) GetActiveFragments() []StoryFragment {
	// Always include common fragments
	result := make([]StoryFragment, len(n.CommonFrags))
	copy(result, n.CommonFrags)

	// If path is active, add path-specific fragments
	if n.ActivePathID != "" {
		for _, path := range n.Paths {
			if path.PathID == n.ActivePathID {
				result = append(result, path.Fragments...)
				break
			}
		}
	}

	return result
}

// GetActivePath returns the currently active path, or nil if undecided
func (n *BranchingNarrative) GetActivePath() *NarrativePath {
	if n.ActivePathID == "" {
		return nil
	}

	for i := range n.Paths {
		if n.Paths[i].PathID == n.ActivePathID {
			return &n.Paths[i]
		}
	}

	return nil
}
