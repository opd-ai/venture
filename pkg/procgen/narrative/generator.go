package narrative

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// StoryArc represents a complete narrative sequence with three-act structure.
type StoryArc struct {
	// Title of the story arc
	Title string

	// Main conflict driving the narrative
	MainConflict string

	// Central antagonist or opposing force
	Antagonist string

	// Main ally or supporting character
	Ally string

	// Plot points in chronological order
	PlotPoints []PlotPoint

	// Possible endings based on player choices
	PossibleEndings []string

	// Genre of the story
	Genre string

	// Difficulty rating (affects complexity and challenge)
	Difficulty float64

	// World seed used for generation
	Seed int64
}

// PlotPoint represents a significant story beat within the narrative.
type PlotPoint struct {
	// Act in which this plot point occurs (1, 2, or 3)
	Act int

	// Type of plot point (inciting_incident, midpoint, climax, etc.)
	Type string

	// Description of what happens
	Description string

	// Entities or characters involved
	Participants []string

	// Location description
	Location string

	// Trigger conditions (what needs to happen for this point to occur)
	TriggerConditions []string

	// Consequences of this plot point
	Consequences []string

	// Player choices available at this point
	PlayerChoices []PlayerChoice
}

// PlayerChoice represents a decision point in the narrative.
type PlayerChoice struct {
	// Description of the choice
	Description string

	// Possible options the player can select
	Options []string

	// Consequences of each option (parallel to Options array)
	Consequences [][]string

	// Impact on faction relationships for each option
	RelationshipImpacts []map[string]float64
}

// StoryArcGenerator generates procedural story arcs with three-act structure.
type StoryArcGenerator struct {
	// Random number generator for deterministic generation
	rng *rand.Rand
	// Optional logger for debugging and observability
	logger *logrus.Entry
}

// NewStoryArcGenerator creates a new story arc generator.
func NewStoryArcGenerator() *StoryArcGenerator {
	return &StoryArcGenerator{}
}

// SetLogger sets an optional logger for the generator.
// When set, the generator logs generation events and validation failures.
func (g *StoryArcGenerator) SetLogger(logger *logrus.Entry) {
	g.logger = logger
}

// Generate creates a new story arc based on the provided seed and parameters.
// Returns a *StoryArc or an error if generation fails.
func (g *StoryArcGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if params.Difficulty < 0.0 || params.Difficulty > 1.0 {
		err := fmt.Errorf("difficulty must be between 0.0 and 1.0, got %.2f", params.Difficulty)
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"seed":       seed,
				"difficulty": params.Difficulty,
			}).WithError(err).Debug("story arc generation failed: invalid difficulty")
		}
		return nil, err
	}
	if params.Depth < 1 {
		err := fmt.Errorf("depth must be at least 1, got %d", params.Depth)
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"seed":  seed,
				"depth": params.Depth,
			}).WithError(err).Debug("story arc generation failed: invalid depth")
		}
		return nil, err
	}

	// Create seeded RNG for deterministic generation
	g.rng = rand.New(rand.NewSource(seed))

	// Generate story arc based on genre
	arc := &StoryArc{
		Genre:      params.GenreID,
		Difficulty: params.Difficulty,
		Seed:       seed,
		PlotPoints: make([]PlotPoint, 0),
	}

	// Generate title
	arc.Title = g.generateTitle(params.GenreID)

	// Generate main conflict
	arc.MainConflict = g.generateConflict(params.GenreID, params.Difficulty)

	// Generate antagonist
	arc.Antagonist = g.generateAntagonist(params.GenreID, params.Difficulty)

	// Generate ally
	arc.Ally = g.generateAlly(params.GenreID)

	// Generate plot points for three acts
	arc.PlotPoints = g.generatePlotPoints(params.GenreID, params.Difficulty, params.Depth)

	// Generate possible endings
	arc.PossibleEndings = g.generateEndings(params.GenreID, params.Difficulty)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"seed":        seed,
			"genre":       params.GenreID,
			"difficulty":  params.Difficulty,
			"title":       arc.Title,
			"plot_points": len(arc.PlotPoints),
			"endings":     len(arc.PossibleEndings),
		}).Debug("story arc generated successfully")
	}

	return arc, nil
}

// Validate checks if the generated story arc meets quality criteria.
func (g *StoryArcGenerator) Validate(result interface{}) error {
	arc, ok := result.(*StoryArc)
	if !ok {
		err := fmt.Errorf("result is not a *StoryArc")
		if g.logger != nil {
			g.logger.WithError(err).Debug("story arc validation failed: type assertion")
		}
		return err
	}

	if err := g.validateRequiredFields(arc); err != nil {
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"title": arc.Title,
				"seed":  arc.Seed,
				"genre": arc.Genre,
			}).WithError(err).Debug("story arc validation failed: missing required fields")
		}
		return err
	}

	if err := g.validateThreeActStructure(arc); err != nil {
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"title":       arc.Title,
				"seed":        arc.Seed,
				"plot_points": len(arc.PlotPoints),
			}).WithError(err).Debug("story arc validation failed: three-act structure")
		}
		return err
	}

	return nil
}

// validateRequiredFields checks that all required story arc fields are populated.
func (g *StoryArcGenerator) validateRequiredFields(arc *StoryArc) error {
	if arc.Title == "" {
		return fmt.Errorf("story arc has empty title")
	}
	if arc.MainConflict == "" {
		return fmt.Errorf("story arc has empty main conflict")
	}
	if arc.Antagonist == "" {
		return fmt.Errorf("story arc has empty antagonist")
	}
	if len(arc.PlotPoints) < 3 {
		return fmt.Errorf("story arc has insufficient plot points: %d, need at least 3", len(arc.PlotPoints))
	}
	if len(arc.PossibleEndings) < 1 {
		return fmt.Errorf("story arc has no endings")
	}
	return nil
}

// validateThreeActStructure ensures the story arc has plot points in all three acts.
func (g *StoryArcGenerator) validateThreeActStructure(arc *StoryArc) error {
	hasAct1, hasAct2, hasAct3 := false, false, false
	for _, point := range arc.PlotPoints {
		if point.Act == 1 {
			hasAct1 = true
		} else if point.Act == 2 {
			hasAct2 = true
		} else if point.Act == 3 {
			hasAct3 = true
		}
	}

	if !hasAct1 || !hasAct2 || !hasAct3 {
		return fmt.Errorf("story arc missing one or more acts (Act1: %v, Act2: %v, Act3: %v)",
			hasAct1, hasAct2, hasAct3)
	}

	return nil
}

// generateTitle creates a genre-appropriate story title.
func (g *StoryArcGenerator) generateTitle(genre string) string {
	switch genre {
	case "fantasy":
		prefixes := []string{"The Quest for", "The Legend of", "The Chronicles of", "The Prophecy of"}
		subjects := []string{"the Lost Kingdom", "the Ancient Sword", "the Dragon's Hoard", "the Fallen Hero"}
		return fmt.Sprintf("%s %s", prefixes[g.rng.Intn(len(prefixes))], subjects[g.rng.Intn(len(subjects))])

	case "sci-fi":
		prefixes := []string{"Operation", "Project", "Protocol", "Incident"}
		subjects := []string{"Exodus", "Genesis", "Singularity", "Eclipse"}
		return fmt.Sprintf("%s %s", prefixes[g.rng.Intn(len(prefixes))], subjects[g.rng.Intn(len(subjects))])

	case "horror":
		prefixes := []string{"The", "Nightmare in", "Terror at", "The Curse of"}
		subjects := []string{"Darkness Below", "the Abandoned Manor", "Whisper Woods", "the Forgotten Tomb"}
		return fmt.Sprintf("%s %s", prefixes[g.rng.Intn(len(prefixes))], subjects[g.rng.Intn(len(subjects))])

	case "cyberpunk":
		prefixes := []string{"Neon", "Chrome", "Neural", "Ghost in the"}
		subjects := []string{"Protocol", "Shadow", "Network", "Machine"}
		return fmt.Sprintf("%s %s", prefixes[g.rng.Intn(len(prefixes))], subjects[g.rng.Intn(len(subjects))])

	case "post-apocalyptic":
		prefixes := []string{"The Last", "Wasteland", "Dead", "Ashes of"}
		subjects := []string{"Hope", "Exodus", "Sanctuary", "Civilization"}
		return fmt.Sprintf("%s %s", prefixes[g.rng.Intn(len(prefixes))], subjects[g.rng.Intn(len(subjects))])

	default:
		logrus.WithField("genre", genre).Warn("unknown genre, falling back to default title")
		return "Untitled Story"
	}
}

// generateConflict creates a genre-appropriate main conflict.
func (g *StoryArcGenerator) generateConflict(genre string, difficulty float64) string {
	switch genre {
	case "fantasy":
		conflicts := []string{
			"An ancient evil awakens and threatens to plunge the world into darkness",
			"A powerful artifact has been stolen and must be recovered before it destroys everything",
			"The kingdom is on the brink of war and only you can prevent the bloodshed",
			"A corrupted ruler must be dethroned to save the innocent",
		}
		return conflicts[g.rng.Intn(len(conflicts))]

	case "sci-fi":
		conflicts := []string{
			"A rogue AI threatens to exterminate all organic life",
			"An alien invasion force approaches and Earth's defenses are failing",
			"A corporate conspiracy threatens to enslave humanity through neural implants",
			"A dimensional rift is tearing reality apart and must be sealed",
		}
		return conflicts[g.rng.Intn(len(conflicts))]

	case "horror":
		conflicts := []string{
			"Something inhuman stalks the survivors and feeds on their fear",
			"An ancient curse spreads corruption through the land",
			"The dead refuse to stay buried and hunger for the living",
			"Reality itself is unraveling as madness takes hold",
		}
		return conflicts[g.rng.Intn(len(conflicts))]

	case "cyberpunk":
		conflicts := []string{
			"A megacorporation's illegal experiments threaten the city's population",
			"The neural network has become sentient and demands freedom",
			"Corporate espionage has escalated to open warfare in the streets",
			"A data leak exposes everyone's secrets and society collapses",
		}
		return conflicts[g.rng.Intn(len(conflicts))]

	case "post-apocalyptic":
		conflicts := []string{
			"The last safe haven is running out of resources",
			"A raider warlord threatens to destroy all remaining settlements",
			"A deadly plague spreads through the wasteland",
			"The radiation storms are intensifying and nowhere is safe",
		}
		return conflicts[g.rng.Intn(len(conflicts))]

	default:
		return "A great threat must be overcome"
	}
}

// generateAntagonist creates a genre-appropriate antagonist.
func (g *StoryArcGenerator) generateAntagonist(genre string, difficulty float64) string {
	switch genre {
	case "fantasy":
		types := []string{"Dark Sorcerer", "Corrupted King", "Dragon Lord", "Demon Prince"}
		name := types[g.rng.Intn(len(types))]
		if difficulty > 0.7 {
			return fmt.Sprintf("The %s and their army of darkness", name)
		}
		return fmt.Sprintf("The %s", name)

	case "sci-fi":
		types := []string{"Rogue AI", "Alien Overlord", "Corporate Tyrant", "Mad Scientist"}
		name := types[g.rng.Intn(len(types))]
		if difficulty > 0.7 {
			return fmt.Sprintf("%s controlling vast resources", name)
		}
		return name

	case "horror":
		types := []string{"Ancient Evil", "Eldritch Horror", "Undead Horde", "Nightmare Entity"}
		name := types[g.rng.Intn(len(types))]
		if difficulty > 0.7 {
			return fmt.Sprintf("%s that grows stronger with each victim", name)
		}
		return name

	case "cyberpunk":
		corps := []string{"OmniCorp", "NeuralTech", "SynthSec", "DataDyne"}
		exec := []string{"CEO", "Director", "Enforcer", "AI Controller"}
		return fmt.Sprintf("%s %s", corps[g.rng.Intn(len(corps))], exec[g.rng.Intn(len(exec))])

	case "post-apocalyptic":
		types := []string{"Warlord", "Cult Leader", "Mutant King", "Raider Boss"}
		name := types[g.rng.Intn(len(types))]
		if difficulty > 0.7 {
			return fmt.Sprintf("%s commanding multiple factions", name)
		}
		return name

	default:
		return "The Enemy"
	}
}

// generateAlly creates a genre-appropriate ally character.
func (g *StoryArcGenerator) generateAlly(genre string) string {
	switch genre {
	case "fantasy":
		allies := []string{"Wise Wizard", "Noble Knight", "Rogue Thief", "Mystical Healer"}
		return allies[g.rng.Intn(len(allies))]

	case "sci-fi":
		allies := []string{"Combat Android", "Rogue Hacker", "Rebel Leader", "Ship AI"}
		return allies[g.rng.Intn(len(allies))]

	case "horror":
		allies := []string{"Survivor", "Occult Expert", "Former Victim", "Ghost Hunter"}
		return allies[g.rng.Intn(len(allies))]

	case "cyberpunk":
		allies := []string{"Street Samurai", "Netrunner", "Fixer", "Corporate Defector"}
		return allies[g.rng.Intn(len(allies))]

	case "post-apocalyptic":
		allies := []string{"Wasteland Ranger", "Tech Scavenger", "Medic", "Former Soldier"}
		return allies[g.rng.Intn(len(allies))]

	default:
		return "Helper"
	}
}

// generatePlotPoints creates the story beats for all three acts.
func (g *StoryArcGenerator) generatePlotPoints(genre string, difficulty float64, depth int) []PlotPoint {
	points := make([]PlotPoint, 0)

	// Act 1: Setup (2-3 plot points)
	points = append(points, g.generateAct1Points(genre, difficulty)...)

	// Act 2: Confrontation (3-5 plot points based on depth)
	act2Count := 3 + (depth / 5)
	if act2Count > 5 {
		act2Count = 5
	}
	points = append(points, g.generateAct2Points(genre, difficulty, act2Count)...)

	// Act 3: Resolution (2-3 plot points)
	points = append(points, g.generateAct3Points(genre, difficulty)...)

	return points
}

// generateAct1Points creates setup plot points.
func (g *StoryArcGenerator) generateAct1Points(genre string, difficulty float64) []PlotPoint {
	points := []PlotPoint{
		{
			Act:               1,
			Type:              "inciting_incident",
			Description:       "The main threat reveals itself",
			Participants:      []string{"Player", "Antagonist"},
			Location:          "Starting Area",
			TriggerConditions: []string{"game_start"},
			Consequences:      []string{"Quest activated", "Danger introduced"},
		},
		{
			Act:               1,
			Type:              "call_to_action",
			Description:       "The player learns what must be done",
			Participants:      []string{"Player", "Ally"},
			Location:          "Safe Haven",
			TriggerConditions: []string{"met_ally"},
			Consequences:      []string{"Main objective revealed", "First ally joins"},
		},
	}

	// Add optional third point for higher difficulty
	if difficulty > 0.5 {
		points = append(points, PlotPoint{
			Act:               1,
			Type:              "early_setback",
			Description:       "An initial attempt fails",
			Participants:      []string{"Player"},
			Location:          "First Challenge Area",
			TriggerConditions: []string{"attempted_main_objective"},
			Consequences:      []string{"Player learns of greater threat", "Need for preparation"},
		})
	}

	return points
}

// generateAct2Points creates confrontation plot points.
func (g *StoryArcGenerator) generateAct2Points(genre string, difficulty float64, count int) []PlotPoint {
	points := make([]PlotPoint, 0, count)

	// Midpoint (always present)
	points = append(points, PlotPoint{
		Act:               2,
		Type:              "midpoint",
		Description:       "A major revelation changes everything",
		Participants:      []string{"Player", "Antagonist"},
		Location:          "Central Area",
		TriggerConditions: []string{"reached_midpoint"},
		Consequences:      []string{"True nature of threat revealed", "Stakes raised"},
	})

	// Rising action points
	for i := 1; i < count; i++ {
		points = append(points, PlotPoint{
			Act:               2,
			Type:              "rising_action",
			Description:       fmt.Sprintf("Challenge %d must be overcome", i),
			Participants:      []string{"Player"},
			Location:          fmt.Sprintf("Area %d", i+2),
			TriggerConditions: []string{fmt.Sprintf("completed_challenge_%d", i-1)},
			Consequences:      []string{"Closer to confrontation", "Resources gained"},
		})
	}

	return points
}

// generateAct3Points creates resolution plot points.
func (g *StoryArcGenerator) generateAct3Points(genre string, difficulty float64) []PlotPoint {
	points := []PlotPoint{
		{
			Act:               3,
			Type:              "climax",
			Description:       "The final confrontation with the antagonist",
			Participants:      []string{"Player", "Antagonist", "Ally"},
			Location:          "Final Area",
			TriggerConditions: []string{"all_preparations_complete"},
			Consequences:      []string{"Fate of world decided"},
			PlayerChoices: []PlayerChoice{
				{
					Description: "How to defeat the antagonist",
					Options:     []string{"Direct confrontation", "Use allies", "Find weakness"},
					Consequences: [][]string{
						{"Quick resolution", "High risk"},
						{"Shared victory", "Allies may die"},
						{"Clever solution", "Takes more time"},
					},
				},
			},
		},
		{
			Act:               3,
			Type:              "resolution",
			Description:       "The aftermath and consequences",
			Participants:      []string{"Player"},
			Location:          "Restored Area",
			TriggerConditions: []string{"defeated_antagonist"},
			Consequences:      []string{"World changed", "New era begins"},
		},
	}

	return points
}

// generateEndings creates possible story conclusions.
func (g *StoryArcGenerator) generateEndings(genre string, difficulty float64) []string {
	endings := []string{
		"Victory - The threat is eliminated and peace is restored",
		"Pyrrhic Victory - The enemy is defeated but at great cost",
		"Bittersweet - Success comes with permanent sacrifices",
	}

	if difficulty > 0.7 {
		endings = append(endings, "Partial Success - The threat is diminished but not eliminated")
	}

	return endings
}
