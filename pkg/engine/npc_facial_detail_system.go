// Package engine provides the NpcFacialDetailSystem which assigns genre-aware
// facial feature parameters (eye color, mouth color, expression, head shape) to
// NPC and creature sprites. Genre palettes drive ColorRole mapping: fantasy uses
// warm amber eyes, horror uses pale grey, sci-fi uses cyan, cyberpunk uses neon
// green, and post-apocalyptic uses dull ochre. Faction relationships determine
// expression type (hostile, friendly, neutral, scared). Head shape tags guide
// sprite generators toward genre-appropriate silhouettes.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// NpcFacialDetailComponent stores procedurally assigned facial feature parameters
// for NPC and creature entities. These values feed into the sprite rendering
// pipeline for per-entity facial variation.
type NpcFacialDetailComponent struct {
	// Eye color (RGB, 0.0-1.0) derived from genre palette ColorRole
	EyeR float64
	EyeG float64
	EyeB float64

	// Mouth color (RGB, 0.0-1.0) derived from genre palette Secondary role
	MouthR float64
	MouthG float64
	MouthB float64

	// Eye size in pixels (1-3); 2px standard for humanoids
	EyeSize float64

	// Mouth size in pixels (1-2)
	MouthSize float64

	// Expression: "neutral", "hostile", "friendly", "scared"
	ExpressionType string

	// Genre-driven head shape tag for sprite generator guidance
	// e.g. "circle", "ellipse", "skull", "angular", "geometric", "visor"
	HeadShapeTag string
}

// Type returns the component type identifier.
func (c *NpcFacialDetailComponent) Type() string {
	return "npc_facial_detail"
}

// NewNpcFacialDetailComponent creates a component with neutral defaults.
func NewNpcFacialDetailComponent() *NpcFacialDetailComponent {
	return &NpcFacialDetailComponent{
		EyeR:           0.2,
		EyeG:           0.2,
		EyeB:           0.2,
		MouthR:          0.6,
		MouthG:          0.4,
		MouthB:          0.4,
		EyeSize:         2.0,
		MouthSize:       1.0,
		ExpressionType:  "neutral",
		HeadShapeTag:    "circle",
	}
}

// genreFacialPalette holds genre-specific eye and mouth colors using ColorRole mapping.
type genreFacialPalette struct {
	// Eye color from Accent role
	EyeR, EyeG, EyeB float64
	// Mouth color from Secondary role
	MouthR, MouthG, MouthB float64
	// Preferred head shape tags (first is default, others are alternates)
	HeadShapes []string
}

// NpcFacialDetailSystem assigns genre-aware facial features to NPC/creature
// entities (those with "ai" and "sprite" but not "input"). Recomputes only
// when genre changes, using throttled checks.
type NpcFacialDetailSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	updateInterval float64
	timeSinceCheck float64

	genreID     string
	lastGenreID string
	initialized bool

	palettes map[string]genreFacialPalette
}

// NewNpcFacialDetailSystem creates a new NPC facial detail system.
func NewNpcFacialDetailSystem(world *World, seed int64) *NpcFacialDetailSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "npc_facial_detail")
	}

	sys := &NpcFacialDetailSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0,
		genreID:        "fantasy",
		lastGenreID:    "",
		palettes:       buildGenreFacialPalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("npc facial detail system created")
	}
	return sys
}

// SetGenre configures the active genre for facial feature computation.
func (s *NpcFacialDetailSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for npc facial detail")
	}
}

// Update checks for genre changes and applies facial detail to NPC entities.
func (s *NpcFacialDetailSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	genreChanged := s.genreID != s.lastGenreID || !s.initialized
	if !genreChanged {
		return
	}

	s.lastGenreID = s.genreID
	s.initialized = true

	s.applyFacialDetails(entities)
}

// applyFacialDetails sets genre-based facial parameters on NPC entities.
func (s *NpcFacialDetailSystem) applyFacialDetails(entities []*Entity) {
	palette, ok := s.palettes[s.genreID]
	if !ok {
		palette = s.palettes["fantasy"]
	}

	count := 0
	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		if entity.HasComponent("input") {
			continue
		}

		comp := s.getOrCreateComponent(entity)

		// Eye color from genre Accent palette with per-entity variation
		variation := s.rng.Float64()*0.1 - 0.05 // ±5% variation
		comp.EyeR = clampFacialColor(palette.EyeR + variation)
		comp.EyeG = clampFacialColor(palette.EyeG + variation)
		comp.EyeB = clampFacialColor(palette.EyeB + variation)

		// Mouth color from genre Secondary palette
		comp.MouthR = clampFacialColor(palette.MouthR + variation)
		comp.MouthG = clampFacialColor(palette.MouthG + variation)
		comp.MouthB = clampFacialColor(palette.MouthB + variation)

		// Eye/mouth sizing based on entity size
		comp.EyeSize = s.getEyeSize(entity)
		comp.MouthSize = s.getMouthSize(entity)

		// Expression from faction relationship
		comp.ExpressionType = s.getExpression(entity)

		// Head shape from genre with alternate selection for variety
		comp.HeadShapeTag = s.selectHeadShape(palette)

		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":           s.genreID,
			"entities_styled": count,
		}).Debug("npc facial details applied")
	}
}

// getOrCreateComponent retrieves or lazily creates the facial detail component.
func (s *NpcFacialDetailSystem) getOrCreateComponent(entity *Entity) *NpcFacialDetailComponent {
	comp, ok := entity.GetComponent("npc_facial_detail")
	if ok {
		if fc, ok := comp.(*NpcFacialDetailComponent); ok {
			return fc
		}
	}
	fc := NewNpcFacialDetailComponent()
	entity.AddComponent(fc)
	return fc
}

// getEyeSize returns eye pixel size based on entity size component.
func (s *NpcFacialDetailSystem) getEyeSize(entity *Entity) float64 {
	comp, ok := entity.GetComponent("creature_size_proportion")
	if ok {
		if sp, ok := comp.(*CreatureSizeProportionComponent); ok {
			// Larger creatures get slightly bigger eyes
			if sp.WidthScale > 1.2 {
				return 3.0
			}
			if sp.WidthScale < 0.8 {
				return 1.0
			}
		}
	}
	return 2.0 // Standard 2px eyes
}

// getMouthSize returns mouth pixel size based on entity size component.
func (s *NpcFacialDetailSystem) getMouthSize(entity *Entity) float64 {
	comp, ok := entity.GetComponent("creature_size_proportion")
	if ok {
		if sp, ok := comp.(*CreatureSizeProportionComponent); ok {
			if sp.WidthScale > 1.2 {
				return 2.0
			}
		}
	}
	return 1.0 // Standard 1px mouth
}

// getExpression determines facial expression from faction relationship.
func (s *NpcFacialDetailSystem) getExpression(entity *Entity) string {
	comp, ok := entity.GetComponent("faction")
	if !ok {
		return "neutral"
	}
	faction, ok := comp.(*FactionComponent)
	if !ok {
		return "neutral"
	}

	switch faction.FactionID {
	case "boss_faction":
		return "hostile"
	case "neutral_faction", "merchant_faction":
		return "friendly"
	case "horror_faction":
		return "scared"
	default:
		return "neutral"
	}
}

// selectHeadShape picks a head shape from the genre palette, with occasional variety.
func (s *NpcFacialDetailSystem) selectHeadShape(palette genreFacialPalette) string {
	if len(palette.HeadShapes) == 0 {
		return "circle"
	}
	// 70% default shape, 30% alternate
	if s.rng.Float64() < 0.7 || len(palette.HeadShapes) == 1 {
		return palette.HeadShapes[0]
	}
	idx := 1 + s.rng.Intn(len(palette.HeadShapes)-1)
	return palette.HeadShapes[idx]
}

// buildGenreFacialPalettes returns genre-specific facial color and shape presets.
// Eye colors use the Accent ColorRole, mouth uses Secondary, head shapes are
// genre-thematic silhouette hints for sprite generation.
func buildGenreFacialPalettes() map[string]genreFacialPalette {
	return map[string]genreFacialPalette{
		"fantasy": {
			EyeR: 0.85, EyeG: 0.65, EyeB: 0.20, // Warm amber eyes
			MouthR: 0.75, MouthG: 0.45, MouthB: 0.40,
			HeadShapes: []string{"circle", "ellipse", "round"},
		},
		"horror": {
			EyeR: 0.70, EyeG: 0.70, EyeB: 0.72, // Pale grey eyes
			MouthR: 0.55, MouthG: 0.40, MouthB: 0.42,
			HeadShapes: []string{"skull", "angular", "gaunt"},
		},
		"scifi": {
			EyeR: 0.30, EyeG: 0.80, EyeB: 0.90, // Cyan eyes
			MouthR: 0.60, MouthG: 0.55, MouthB: 0.65,
			HeadShapes: []string{"angular", "geometric", "sleek"},
		},
		"cyberpunk": {
			EyeR: 0.20, EyeG: 0.90, EyeB: 0.40, // Neon green eyes
			MouthR: 0.70, MouthG: 0.50, MouthB: 0.60,
			HeadShapes: []string{"geometric", "visor", "angular"},
		},
		"postapoc": {
			EyeR: 0.75, EyeG: 0.60, EyeB: 0.30, // Dull ochre eyes
			MouthR: 0.65, MouthG: 0.50, MouthB: 0.40,
			HeadShapes: []string{"rugged", "angular", "scarred"},
		},
	}
}

// clampFacialColor ensures a color value stays in [0.0, 1.0].
func clampFacialColor(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}
