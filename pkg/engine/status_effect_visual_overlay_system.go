// Package engine provides the StatusEffectVisualOverlaySystem which connects
// status effects to sprite color tints via VisualFeedbackComponent.
// Entities with active status effects (burn, poison, frost, stun, etc.) receive
// color-coded tint overlays that blend based on effect intensity and duration.
// Genre-aware: tint palettes can vary by genre (e.g., cyberpunk poison = neon green).
package engine

import (
	"image/color"
	"math"

	"github.com/sirupsen/logrus"
)

// StatusEffectTintConfig defines the tint color and intensity for a status effect type.
type StatusEffectTintConfig struct {
	Color     color.RGBA // Base tint color
	Intensity float64    // Tint strength (0.0-1.0), blended with white (1,1,1)
	Pulsing   bool       // Whether tint pulses over time
	PulseRate float64    // Pulse frequency in Hz
}

// StatusEffectVisualOverlaySystem applies color tints to entity sprites based on
// active status effects. It reads StatusEffectComponent data and writes to
// VisualFeedbackComponent tint fields, connecting combat/environmental effects
// to the rendering pipeline.
type StatusEffectVisualOverlaySystem struct {
	world  *World
	logger *logrus.Entry

	// Effect type -> tint configuration
	effectTints map[string]StatusEffectTintConfig

	// Track which entities have active tints to clean up when effects expire
	activeTints map[uint64]string

	// Accumulated time for pulse calculations
	elapsed float64

	// Genre for palette selection
	genreID string
}

// NewStatusEffectVisualOverlaySystem creates a new status effect visual overlay system.
func NewStatusEffectVisualOverlaySystem(world *World) *StatusEffectVisualOverlaySystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "status_effect_visual_overlay",
	})

	sys := &StatusEffectVisualOverlaySystem{
		world:       world,
		logger:      logger,
		effectTints: make(map[string]StatusEffectTintConfig),
		activeTints: make(map[uint64]string, 64),
	}

	sys.initDefaultTints()
	logger.Debug("status effect visual overlay system created")
	return sys
}

// SetGenre configures genre-specific tint variations.
func (s *StatusEffectVisualOverlaySystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreOverrides()
}

// initDefaultTints sets up default tint colors for common status effects.
func (s *StatusEffectVisualOverlaySystem) initDefaultTints() {
	// Burning - warm red/orange tint
	s.effectTints["burn"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 255, G: 100, B: 50, A: 255},
		Intensity: 0.35,
		Pulsing:   true,
		PulseRate: 3.0,
	}

	// Poison - sickly green tint
	s.effectTints["poison"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 80, G: 220, B: 60, A: 255},
		Intensity: 0.30,
		Pulsing:   true,
		PulseRate: 1.5,
	}

	// Frost/Frozen - icy blue tint
	s.effectTints["frost"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 120, G: 180, B: 255, A: 255},
		Intensity: 0.40,
		Pulsing:   false,
	}
	s.effectTints["frozen"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 150, G: 200, B: 255, A: 255},
		Intensity: 0.55,
		Pulsing:   false,
	}

	// Stun - bright yellow tint
	s.effectTints["stun"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 255, G: 255, B: 120, A: 255},
		Intensity: 0.30,
		Pulsing:   true,
		PulseRate: 4.0,
	}

	// Regeneration - golden healing tint
	s.effectTints["regeneration"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 255, G: 220, B: 100, A: 255},
		Intensity: 0.20,
		Pulsing:   true,
		PulseRate: 2.0,
	}

	// Shock/Electric - bright white-blue tint
	s.effectTints["shock"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 200, G: 220, B: 255, A: 255},
		Intensity: 0.45,
		Pulsing:   true,
		PulseRate: 8.0,
	}
	s.effectTints["electrified"] = s.effectTints["shock"]

	// Blessed - soft white-gold glow
	s.effectTints["blessed"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 255, G: 240, B: 200, A: 255},
		Intensity: 0.25,
		Pulsing:   true,
		PulseRate: 1.0,
	}

	// Cursed - dark purple tint
	s.effectTints["cursed"] = StatusEffectTintConfig{
		Color:     color.RGBA{R: 130, G: 50, B: 180, A: 255},
		Intensity: 0.35,
		Pulsing:   true,
		PulseRate: 0.8,
	}
}

// applyGenreOverrides adjusts tint colors for the current genre.
func (s *StatusEffectVisualOverlaySystem) applyGenreOverrides() {
	switch s.genreID {
	case "cyberpunk":
		// Neon-tinted effects for cyberpunk
		s.effectTints["poison"] = StatusEffectTintConfig{
			Color: color.RGBA{R: 0, G: 255, B: 100, A: 255}, Intensity: 0.35,
			Pulsing: true, PulseRate: 2.0,
		}
		s.effectTints["shock"] = StatusEffectTintConfig{
			Color: color.RGBA{R: 0, G: 200, B: 255, A: 255}, Intensity: 0.50,
			Pulsing: true, PulseRate: 10.0,
		}
		s.effectTints["electrified"] = s.effectTints["shock"]
	case "horror":
		// Darker, more saturated tints for horror
		s.effectTints["cursed"] = StatusEffectTintConfig{
			Color: color.RGBA{R: 80, G: 0, B: 120, A: 255}, Intensity: 0.50,
			Pulsing: true, PulseRate: 0.5,
		}
		s.effectTints["poison"] = StatusEffectTintConfig{
			Color: color.RGBA{R: 40, G: 160, B: 30, A: 255}, Intensity: 0.40,
			Pulsing: true, PulseRate: 1.0,
		}
	case "scifi":
		// Tech-flavored tints
		s.effectTints["burn"] = StatusEffectTintConfig{
			Color: color.RGBA{R: 255, G: 140, B: 0, A: 255}, Intensity: 0.40,
			Pulsing: true, PulseRate: 4.0,
		}
	}
}

// Update processes all entities, applying or clearing tints based on status effects.
func (s *StatusEffectVisualOverlaySystem) Update(entities []*Entity, deltaTime float64) {
	s.elapsed += deltaTime

	activeThisFrame := make(map[uint64]bool, len(s.activeTints))

	// Build entity lookup for cleanup (entities may not yet be in world.entities)
	entityMap := make(map[uint64]*Entity, len(entities))
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	for _, entity := range entities {
		feedback := entity.GetVisualFeedback()
		if feedback == nil {
			continue
		}

		// Find dominant status effect
		effectType, magnitude := s.findDominantEffect(entity)
		if effectType == "" {
			continue
		}

		config, exists := s.effectTints[effectType]
		if !exists {
			continue
		}

		activeThisFrame[entity.ID] = true

		// Compute tint intensity with optional pulsing
		intensity := config.Intensity
		if magnitude > 0 {
			// Scale intensity by effect magnitude (clamped 0.5x - 1.5x)
			scale := math.Min(1.5, math.Max(0.5, magnitude/20.0))
			intensity *= scale
		}
		intensity = math.Min(intensity, 0.8) // Cap to avoid fully obscuring sprite

		if config.Pulsing && config.PulseRate > 0 {
			pulse := 0.5 + 0.5*math.Sin(s.elapsed*config.PulseRate*2*math.Pi)
			intensity *= 0.6 + 0.4*pulse // Pulse between 60% and 100% of intensity
		}

		// Blend tint color with white (1,1,1) based on intensity
		r := 1.0 - intensity*(1.0-float64(config.Color.R)/255.0)
		g := 1.0 - intensity*(1.0-float64(config.Color.G)/255.0)
		b := 1.0 - intensity*(1.0-float64(config.Color.B)/255.0)

		feedback.SetTint(r, g, b, 1.0)
		s.activeTints[entity.ID] = effectType
	}

	// Clear tints for entities that no longer have status effects
	for entityID, effectType := range s.activeTints {
		if activeThisFrame[entityID] {
			continue
		}
		s.clearEntityTintFromMap(entityID, entityMap)
		delete(s.activeTints, entityID)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entityID,
				"effect_type": effectType,
			}).Debug("cleared status effect tint")
		}
	}
}

// findDominantEffect returns the highest-priority active status effect on an entity.
// Priority: burn > shock > frost > frozen > poison > stun > cursed > blessed > regeneration
func (s *StatusEffectVisualOverlaySystem) findDominantEffect(entity *Entity) (string, float64) {
	priorityOrder := []string{
		"burn", "shock", "electrified", "frost", "frozen",
		"poison", "stun", "cursed", "blessed", "regeneration",
	}

	for _, effectType := range priorityOrder {
		for _, comp := range entity.Components {
			effect, ok := comp.(*StatusEffectComponent)
			if !ok {
				continue
			}
			if effect.EffectType == effectType && !effect.IsExpired() {
				if _, has := s.effectTints[effectType]; has {
					return effectType, effect.Magnitude
				}
			}
		}
	}

	// Check for any other configured effect types
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}
		if !effect.IsExpired() {
			if _, has := s.effectTints[effect.EffectType]; has {
				return effect.EffectType, effect.Magnitude
			}
		}
	}
	return "", 0
}

// clearEntityTintFromMap resets the tint on an entity's VisualFeedbackComponent.
// Checks the provided entity map first (for entities not yet flushed to world),
// then falls back to world.GetEntity.
func (s *StatusEffectVisualOverlaySystem) clearEntityTintFromMap(entityID uint64, entityMap map[uint64]*Entity) {
	var entity *Entity
	if e, ok := entityMap[entityID]; ok {
		entity = e
	} else if s.world != nil {
		entity, _ = s.world.GetEntity(entityID)
	}
	if entity == nil {
		return
	}
	feedback := entity.GetVisualFeedback()
	if feedback == nil {
		return
	}
	feedback.ClearTint()
}

// GetActiveTintCount returns the number of entities with active status tints.
func (s *StatusEffectVisualOverlaySystem) GetActiveTintCount() int {
	return len(s.activeTints)
}
