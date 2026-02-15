// Package engine provides the CreatureGenreTintSystem which applies genre-specific
// color tints to creature and NPC sprites. Different genres produce distinct visual
// atmospheres: fantasy uses warm golden tones, horror applies cold desaturated hues,
// sci-fi adds cool blue tints, cyberpunk introduces neon-pink highlights, and
// post-apocalyptic overlays dusty sepia tones. Tints compose multiplicatively with
// existing weather and status effect tints in the render pipeline.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CreatureGenreTintComponent stores genre-driven tint multipliers for a creature entity.
// These multiply with other tint sources (weather, status effects) in the render pipeline.
type CreatureGenreTintComponent struct {
	// RGB multipliers (1.0 = no change). Values < 1.0 darken that channel.
	TintR float64
	TintG float64
	TintB float64
}

// Type returns the component type identifier.
func (c *CreatureGenreTintComponent) Type() string {
	return "creature_genre_tint"
}

// NewCreatureGenreTintComponent creates a component with neutral (no-op) tint.
func NewCreatureGenreTintComponent() *CreatureGenreTintComponent {
	return &CreatureGenreTintComponent{
		TintR: 1.0,
		TintG: 1.0,
		TintB: 1.0,
	}
}

// creatureGenreTintPreset holds the RGB multipliers for a genre.
type creatureGenreTintPreset struct {
	R, G, B float64
}

// entityTypeTintModifier adjusts tint based on creature type (boss, minion, etc.).
type entityTypeTintModifier struct {
	R, G, B float64
}

// CreatureGenreTintSystem reads the current genre and applies genre-specific
// color tints to creature/NPC sprites (entities with "ai" but not "input").
type CreatureGenreTintSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Throttle: only recompute when genre changes
	updateInterval float64
	timeSinceCheck float64

	// Cached genre state
	genreID     string
	lastGenreID string
	initialized bool

	// Pre-computed tint presets per genre
	presets map[string]creatureGenreTintPreset
}

// NewCreatureGenreTintSystem creates a new creature genre tint system.
func NewCreatureGenreTintSystem(world *World, seed int64) *CreatureGenreTintSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "creature_genre_tint")
	}

	sys := &CreatureGenreTintSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0,
		genreID:        "fantasy",
		lastGenreID:    "",
		presets:        buildCreatureGenrePresets(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("creature genre tint system created")
	}
	return sys
}

// SetGenre configures the active genre for tint computation.
func (s *CreatureGenreTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for creature genre tint")
	}
}

// Update checks for genre changes and applies tints to creature/NPC entities.
func (s *CreatureGenreTintSystem) Update(entities []*Entity, deltaTime float64) {
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

	s.applyTints(entities)
}

// applyTints sets genre-based tints on creature entities with sprites.
func (s *CreatureGenreTintSystem) applyTints(entities []*Entity) {
	preset, ok := s.presets[s.genreID]
	if !ok {
		preset = creatureGenreTintPreset{R: 1.0, G: 1.0, B: 1.0}
	}

	count := 0
	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		// Skip player entities (they have input)
		if entity.HasComponent("input") {
			continue
		}

		tintComp := s.getOrCreateTintComponent(entity)

		mod := s.getEntityTypeModifier(entity)
		tintComp.TintR = clampCreatureTint(preset.R * mod.R)
		tintComp.TintG = clampCreatureTint(preset.G * mod.G)
		tintComp.TintB = clampCreatureTint(preset.B * mod.B)

		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":           s.genreID,
			"entities_tinted": count,
		}).Debug("creature genre tints applied")
	}
}

// getOrCreateTintComponent retrieves or lazily creates the tint component.
func (s *CreatureGenreTintSystem) getOrCreateTintComponent(entity *Entity) *CreatureGenreTintComponent {
	comp, ok := entity.GetComponent("creature_genre_tint")
	if ok {
		if tintComp, ok := comp.(*CreatureGenreTintComponent); ok {
			return tintComp
		}
	}
	tintComp := NewCreatureGenreTintComponent()
	entity.AddComponent(tintComp)
	return tintComp
}

// getEntityTypeModifier returns tint adjustments based on creature archetype.
func (s *CreatureGenreTintSystem) getEntityTypeModifier(entity *Entity) entityTypeTintModifier {
	comp, ok := entity.GetComponent("faction")
	if !ok {
		return entityTypeTintModifier{R: 1.0, G: 1.0, B: 1.0}
	}
	faction, ok := comp.(*FactionComponent)
	if !ok {
		return entityTypeTintModifier{R: 1.0, G: 1.0, B: 1.0}
	}

	switch faction.FactionID {
	case "boss_faction":
		return entityTypeTintModifier{R: 0.85, G: 0.85, B: 0.85}
	case "neutral_faction":
		return entityTypeTintModifier{R: 1.02, G: 1.0, B: 1.0}
	default:
		return entityTypeTintModifier{R: 1.0, G: 1.0, B: 1.0}
	}
}

// buildCreatureGenrePresets returns the genre-specific tint presets.
func buildCreatureGenrePresets() map[string]creatureGenreTintPreset {
	return map[string]creatureGenreTintPreset{
		"fantasy":   {R: 1.00, G: 0.97, B: 0.88},
		"horror":    {R: 0.82, G: 0.75, B: 0.80},
		"scifi":     {R: 0.88, G: 0.92, B: 1.00},
		"cyberpunk": {R: 1.00, G: 0.85, B: 0.95},
		"postapoc":  {R: 0.95, G: 0.88, B: 0.75},
	}
}

// clampCreatureTint ensures a tint multiplier stays in [0.5, 1.1].
func clampCreatureTint(v float64) float64 {
	if v < 0.5 {
		return 0.5
	}
	if v > 1.1 {
		return 1.1
	}
	return v
}
