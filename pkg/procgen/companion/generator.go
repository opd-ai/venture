// Package companion provides procedural generation of companion entities.
// Companions are AI followers that assist players with combat, gathering, and exploration.
package companion

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// logger is the package-level logger for companion generation
var logger = logrus.WithField("package", "procgen/companion")

// Companion represents a generated companion with stats, commands, abilities,
// and visual description derived from genre-specific templates.
type Companion struct {
	Name          string
	Type          engine.CompanionType
	Level         int
	Attack        float64
	Defense       float64
	Speed         float64
	HP            float64
	MaxHP         float64
	Loyalty       float64
	Commands      []engine.CommandType
	SpritePattern string // Description for sprite generation
}

// Generator creates procedural companions
type Generator struct{}

// NewGenerator creates a new companion generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a new companion
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Validate parameters
	if params.Difficulty < 0.0 || params.Difficulty > 1.0 {
		logger.WithFields(logrus.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Debug("companion generation failed: invalid difficulty")
		return nil, fmt.Errorf("difficulty must be between 0.0 and 1.0, got %f", params.Difficulty)
	}

	// Determine companion type based on genre
	companionType := g.selectCompanionType(rng, params.GenreID)

	// Generate stats scaled with depth and difficulty
	baseLevel := params.Depth
	if baseLevel < 1 {
		baseLevel = 1
	}

	companion := &Companion{
		Name:     g.generateName(rng, companionType, params.GenreID),
		Type:     companionType,
		Level:    baseLevel,
		Loyalty:  50.0 + rng.Float64()*30.0, // 50-80 starting loyalty
		Commands: g.generateCommands(rng),
	}

	// Calculate stats
	levelMultiplier := 1.0 + 0.15*float64(baseLevel-1)
	difficultyMultiplier := 0.5 + params.Difficulty*1.5

	companion.Attack = (10.0 + rng.Float64()*15.0) * levelMultiplier * difficultyMultiplier
	companion.Defense = (8.0 + rng.Float64()*12.0) * levelMultiplier * difficultyMultiplier
	companion.Speed = 80.0 + rng.Float64()*40.0
	companion.MaxHP = (50.0 + rng.Float64()*50.0) * levelMultiplier * difficultyMultiplier
	companion.HP = companion.MaxHP

	companion.SpritePattern = g.generateSpritePattern(companionType, params.GenreID)

	logger.WithFields(logrus.Fields{
		"seed":        seed,
		"genre":       params.GenreID,
		"depth":       params.Depth,
		"difficulty":  params.Difficulty,
		"name":        companion.Name,
		"type":        companionType,
		"level":       companion.Level,
		"attack":      companion.Attack,
		"defense":     companion.Defense,
		"max_hp":      companion.MaxHP,
		"command_cnt": len(companion.Commands),
	}).Debug("companion generated successfully")

	return companion, nil
}

// Validate checks if generated companion is valid
func (g *Generator) Validate(result interface{}) error {
	companion, ok := result.(*Companion)
	if !ok {
		logger.WithField("result_type", fmt.Sprintf("%T", result)).Debug("companion validation failed: wrong type")
		return fmt.Errorf("result is not a Companion")
	}

	if companion.Attack <= 0 {
		logger.WithFields(logrus.Fields{
			"name":   companion.Name,
			"attack": companion.Attack,
		}).Debug("companion validation failed: invalid attack")
		return fmt.Errorf("companion has invalid attack: %f", companion.Attack)
	}

	if companion.MaxHP <= 0 {
		logger.WithFields(logrus.Fields{
			"name":   companion.Name,
			"max_hp": companion.MaxHP,
		}).Debug("companion validation failed: invalid max HP")
		return fmt.Errorf("companion has invalid max HP: %f", companion.MaxHP)
	}

	if companion.Loyalty < 0 || companion.Loyalty > 100 {
		logger.WithFields(logrus.Fields{
			"name":    companion.Name,
			"loyalty": companion.Loyalty,
		}).Debug("companion validation failed: invalid loyalty")
		return fmt.Errorf("companion has invalid loyalty: %f", companion.Loyalty)
	}

	if len(companion.Commands) == 0 {
		logger.WithField("name", companion.Name).Debug("companion validation failed: no commands")
		return fmt.Errorf("companion has no commands")
	}

	return nil
}

func (g *Generator) selectCompanionType(rng *rand.Rand, genreID string) engine.CompanionType {
	// Genre-specific companion types
	switch genreID {
	case "fantasy":
		types := []engine.CompanionType{
			engine.CompanionTypePet,
			engine.CompanionTypeElemental,
			engine.CompanionTypeSpirit,
		}
		return types[rng.Intn(len(types))]
	case "sci-fi":
		types := []engine.CompanionType{
			engine.CompanionTypeRobot,
			engine.CompanionTypeSummon,
		}
		return types[rng.Intn(len(types))]
	case "horror":
		types := []engine.CompanionType{
			engine.CompanionTypeUndead,
			engine.CompanionTypeInsect,
		}
		return types[rng.Intn(len(types))]
	case "cyberpunk":
		return engine.CompanionTypeRobot
	case "post-apocalyptic":
		types := []engine.CompanionType{
			engine.CompanionTypePet,
			engine.CompanionTypeInsect,
		}
		return types[rng.Intn(len(types))]
	default:
		return engine.CompanionType(rng.Intn(8))
	}
}

func (g *Generator) generateName(rng *rand.Rand, companionType engine.CompanionType, genreID string) string {
	prefixes := map[string][]string{
		"fantasy":          {"Shadow", "Storm", "Frost", "Fire", "Wild"},
		"sci-fi":           {"Alpha", "Beta", "Gamma", "Omega", "Sigma"},
		"horror":           {"Dark", "Cursed", "Twisted", "Grim", "Hollow"},
		"cyberpunk":        {"Neo", "Cyber", "Tech", "Data", "Ghost"},
		"post-apocalyptic": {"Rust", "Scrap", "Ash", "Rad", "Dust"},
	}

	suffixes := map[engine.CompanionType][]string{
		engine.CompanionTypePet:       {"paw", "fang", "claw", "tail"},
		engine.CompanionTypeSummon:    {"born", "called", "formed"},
		engine.CompanionTypeHireling:  {"blade", "guard", "shield"},
		engine.CompanionTypeElemental: {"flame", "wave", "gale", "stone"},
		engine.CompanionTypeUndead:    {"bone", "wraith", "shade"},
		engine.CompanionTypeRobot:     {"unit", "bot", "droid", "core"},
		engine.CompanionTypeSpirit:    {"soul", "wisp", "ghost"},
		engine.CompanionTypeInsect:    {"swarm", "hive", "crawler"},
	}

	prefixList := prefixes["fantasy"]
	if list, ok := prefixes[genreID]; ok {
		prefixList = list
	}

	suffixList := suffixes[companionType]
	if len(suffixList) == 0 {
		suffixList = []string{"one", "two", "three"}
	}

	prefix := prefixList[rng.Intn(len(prefixList))]
	suffix := suffixList[rng.Intn(len(suffixList))]

	return prefix + suffix
}

func (g *Generator) generateCommands(rng *rand.Rand) []engine.CommandType {
	// All companions get basic commands
	commands := []engine.CommandType{
		engine.CommandFollow,
		engine.CommandStay,
		engine.CommandAttack,
	}

	// Add additional commands randomly
	if rng.Float64() > 0.5 {
		commands = append(commands, engine.CommandDefend)
	}
	if rng.Float64() > 0.7 {
		commands = append(commands, engine.CommandGather)
	}
	if rng.Float64() > 0.8 {
		commands = append(commands, engine.CommandScout)
	}

	return commands
}

func (g *Generator) generateSpritePattern(companionType engine.CompanionType, genreID string) string {
	switch companionType {
	case engine.CompanionTypePet:
		return "quadruped animal with fur"
	case engine.CompanionTypeRobot:
		return "mechanical humanoid with LED eyes"
	case engine.CompanionTypeElemental:
		return "floating elemental form"
	case engine.CompanionTypeUndead:
		return "skeletal or ghostly figure"
	case engine.CompanionTypeSpirit:
		return "translucent wispy form"
	case engine.CompanionTypeInsect:
		return "insectoid creature with multiple legs"
	default:
		return "small humanoid companion"
	}
}
