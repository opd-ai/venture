// Package engine provides the CreatureSizeProportionSystem which assigns
// anatomically-correct body proportions to creature and NPC sprites based on
// their size category. Smaller creatures receive proportionally larger heads
// (chibi-style), while larger creatures get wider torsos and stockier builds.
// Genre-specific modifiers adjust proportions for atmospheric cohesion: horror
// creatures are elongated, sci-fi entities are angular, cyberpunk is compact.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CreatureSizeProportionComponent stores anatomical proportion ratios for a creature.
// HeadRatio + TorsoRatio + LegRatio = 1.0 (vertical distribution).
// WidthScale adjusts the torso/leg width relative to the sprite's base width.
type CreatureSizeProportionComponent struct {
	HeadRatio  float64 // Fraction of height allocated to head (0.0–0.35)
	TorsoRatio float64 // Fraction of height allocated to torso (0.30–0.55)
	LegRatio   float64 // Fraction of height allocated to legs (0.30–0.50)
	WidthScale float64 // Torso width multiplier (0.8–1.4)
	SizeTier   string  // Inferred size tier: tiny, small, medium, large, huge
}

// Type returns the component type identifier.
func (c *CreatureSizeProportionComponent) Type() string {
	return "creature_size_proportion"
}

// NewCreatureSizeProportionComponent creates a component with medium (default) proportions.
func NewCreatureSizeProportionComponent() *CreatureSizeProportionComponent {
	return &CreatureSizeProportionComponent{
		HeadRatio:  0.19,
		TorsoRatio: 0.41,
		LegRatio:   0.40,
		WidthScale: 1.0,
		SizeTier:   "medium",
	}
}

// sizeTierProportions defines the base anatomy ratios per size tier.
type sizeTierProportions struct {
	HeadRatio  float64
	TorsoRatio float64
	LegRatio   float64
	WidthScale float64
}

// genreProportionModifier adjusts proportions per genre.
type genreProportionModifier struct {
	HeadMult  float64
	TorsoMult float64
	LegMult   float64
	WidthMult float64
}

// CreatureSizeProportionSystem reads creature sprite sizes, infers size tiers,
// and assigns anatomical proportions for downstream rendering.
type CreatureSizeProportionSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	updateInterval float64
	timeSinceCheck float64
	initialized    bool

	genreID     string
	lastGenreID string

	tierMap   map[string]sizeTierProportions
	genreMods map[string]genreProportionModifier
}

// NewCreatureSizeProportionSystem creates a new creature size proportion system.
func NewCreatureSizeProportionSystem(world *World, seed int64) *CreatureSizeProportionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "creature_size_proportion")
	}

	sys := &CreatureSizeProportionSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0,
		genreID:        "fantasy",
		lastGenreID:    "",
		tierMap:        buildSizeTierProportions(),
		genreMods:      buildGenreProportionModifiers(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("creature size proportion system created")
	}
	return sys
}

// SetGenre configures the active genre for proportion adjustments.
func (s *CreatureSizeProportionSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for creature size proportion")
	}
}

// Update checks for genre changes and assigns proportions to creature entities.
func (s *CreatureSizeProportionSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	genreChanged := s.genreID != s.lastGenreID || !s.initialized
	if !genreChanged {
		// Still check for new entities without proportions
		s.assignNewEntities(entities)
		return
	}

	s.lastGenreID = s.genreID
	s.initialized = true
	s.applyProportions(entities)
}

// applyProportions assigns size-based proportions to all creature entities.
func (s *CreatureSizeProportionSystem) applyProportions(entities []*Entity) {
	mod := s.getGenreModifier()
	count := 0

	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		if entity.HasComponent("input") {
			continue
		}

		tier := s.inferSizeTier(entity)
		props := s.tierMap[tier]
		comp := s.getOrCreateComponent(entity)
		s.applyWithGenre(comp, tier, props, mod)
		count++
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":            s.genreID,
			"entities_updated": count,
		}).Debug("creature size proportions applied")
	}
}

// assignNewEntities handles newly spawned entities that lack the component.
func (s *CreatureSizeProportionSystem) assignNewEntities(entities []*Entity) {
	mod := s.getGenreModifier()

	for _, entity := range entities {
		if !entity.HasComponent("ai") || !entity.HasComponent("sprite") {
			continue
		}
		if entity.HasComponent("input") {
			continue
		}
		if entity.HasComponent("creature_size_proportion") {
			continue
		}

		tier := s.inferSizeTier(entity)
		props := s.tierMap[tier]
		comp := s.getOrCreateComponent(entity)
		s.applyWithGenre(comp, tier, props, mod)
	}
}

// inferSizeTier determines size tier from sprite dimensions.
func (s *CreatureSizeProportionSystem) inferSizeTier(entity *Entity) string {
	comp, ok := entity.GetComponent("sprite")
	if !ok {
		return "medium"
	}
	sprite, ok := comp.(*EbitenSprite)
	if !ok {
		return "medium"
	}

	maxDim := sprite.Width
	if sprite.Height > maxDim {
		maxDim = sprite.Height
	}

	switch {
	case maxDim <= 18:
		return "tiny"
	case maxDim <= 26:
		return "small"
	case maxDim <= 40:
		return "medium"
	case maxDim <= 56:
		return "large"
	default:
		return "huge"
	}
}

// getOrCreateComponent retrieves or lazily creates the proportion component.
func (s *CreatureSizeProportionSystem) getOrCreateComponent(entity *Entity) *CreatureSizeProportionComponent {
	comp, ok := entity.GetComponent("creature_size_proportion")
	if ok {
		if propComp, ok := comp.(*CreatureSizeProportionComponent); ok {
			return propComp
		}
	}
	propComp := NewCreatureSizeProportionComponent()
	entity.AddComponent(propComp)
	return propComp
}

// applyWithGenre sets proportions combining base tier values with genre modifiers.
func (s *CreatureSizeProportionSystem) applyWithGenre(comp *CreatureSizeProportionComponent, tier string, props sizeTierProportions, mod genreProportionModifier) {
	comp.SizeTier = tier
	comp.HeadRatio = clampProportion(props.HeadRatio * mod.HeadMult)
	comp.TorsoRatio = clampProportion(props.TorsoRatio * mod.TorsoMult)
	comp.WidthScale = clampWidthScale(props.WidthScale * mod.WidthMult)

	// Normalize: legs get the remainder to keep ratios summing to 1.0
	comp.LegRatio = clampProportion(1.0 - comp.HeadRatio - comp.TorsoRatio)
}

// getGenreModifier returns the active genre's proportion modifier.
func (s *CreatureSizeProportionSystem) getGenreModifier() genreProportionModifier {
	mod, ok := s.genreMods[s.genreID]
	if !ok {
		return genreProportionModifier{HeadMult: 1.0, TorsoMult: 1.0, LegMult: 1.0, WidthMult: 1.0}
	}
	return mod
}

// buildSizeTierProportions returns base anatomy ratios per size tier.
func buildSizeTierProportions() map[string]sizeTierProportions {
	return map[string]sizeTierProportions{
		"tiny":   {HeadRatio: 0.30, TorsoRatio: 0.35, LegRatio: 0.35, WidthScale: 0.85},
		"small":  {HeadRatio: 0.25, TorsoRatio: 0.40, LegRatio: 0.35, WidthScale: 0.90},
		"medium": {HeadRatio: 0.19, TorsoRatio: 0.41, LegRatio: 0.40, WidthScale: 1.00},
		"large":  {HeadRatio: 0.15, TorsoRatio: 0.45, LegRatio: 0.40, WidthScale: 1.15},
		"huge":   {HeadRatio: 0.12, TorsoRatio: 0.48, LegRatio: 0.40, WidthScale: 1.30},
	}
}

// buildGenreProportionModifiers returns genre-specific proportion adjustments.
func buildGenreProportionModifiers() map[string]genreProportionModifier {
	return map[string]genreProportionModifier{
		"fantasy":   {HeadMult: 1.00, TorsoMult: 1.00, LegMult: 1.00, WidthMult: 1.00},
		"horror":    {HeadMult: 0.90, TorsoMult: 1.08, LegMult: 1.02, WidthMult: 0.92},
		"scifi":     {HeadMult: 1.05, TorsoMult: 0.95, LegMult: 1.00, WidthMult: 1.05},
		"cyberpunk": {HeadMult: 0.95, TorsoMult: 1.05, LegMult: 1.00, WidthMult: 1.08},
		"postapoc":  {HeadMult: 1.00, TorsoMult: 1.03, LegMult: 0.97, WidthMult: 1.02},
	}
}

// clampProportion ensures a proportion stays in [0.05, 0.60].
func clampProportion(v float64) float64 {
	if v < 0.05 {
		return 0.05
	}
	if v > 0.60 {
		return 0.60
	}
	return v
}

// clampWidthScale ensures width scale stays in [0.7, 1.5].
func clampWidthScale(v float64) float64 {
	if v < 0.7 {
		return 0.7
	}
	if v > 1.5 {
		return 1.5
	}
	return v
}
