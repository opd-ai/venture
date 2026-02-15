// Package engine provides the DamageFlashTintSystem which triggers genre-aware
// flash tints on entities when they take damage. It monitors HealthComponent
// changes and writes to VisualFeedbackComponent, connecting combat damage to
// the visual rendering pipeline.
package engine

import (
	"image/color"
	"math"

	"github.com/sirupsen/logrus"
)

// damageFlashPreset holds the flash color and timing for a genre.
type damageFlashPreset struct {
	Color        color.RGBA // Flash tint color
	BaseDuration float64    // Base flash duration in seconds
	MinIntensity float64    // Minimum flash intensity for any damage
	MaxIntensity float64    // Maximum flash intensity (full health lost)
}

// DamageFlashTintSystem monitors entity health changes and triggers visual
// flash tints on VisualFeedbackComponent when damage is taken. Flash intensity
// scales with damage proportion and colors vary by genre.
type DamageFlashTintSystem struct {
	world  *World
	logger *logrus.Entry

	// Previous health values keyed by entity ID
	prevHealth map[uint64]float64

	// Genre-specific flash presets
	genreID string
	preset  damageFlashPreset

	// Throttle cleanup of stale entries
	cleanupTimer    float64
	cleanupInterval float64
}

// NewDamageFlashTintSystem creates a new damage flash tint system.
func NewDamageFlashTintSystem(world *World) *DamageFlashTintSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "damage_flash_tint",
	})

	sys := &DamageFlashTintSystem{
		world:           world,
		logger:          logger,
		prevHealth:      make(map[uint64]float64, 128),
		genreID:         "fantasy",
		cleanupInterval: 5.0,
	}

	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("damage flash tint system created")
	return sys
}

// SetGenre configures genre-specific flash colors.
func (s *DamageFlashTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
}

// getGenrePreset returns flash configuration for the given genre.
func (s *DamageFlashTintSystem) getGenrePreset(genreID string) damageFlashPreset {
	switch genreID {
	case "horror":
		return damageFlashPreset{
			Color:        color.RGBA{R: 180, G: 20, B: 20, A: 255},
			BaseDuration: 0.18,
			MinIntensity: 0.25,
			MaxIntensity: 0.90,
		}
	case "cyberpunk":
		return damageFlashPreset{
			Color:        color.RGBA{R: 0, G: 220, B: 255, A: 255},
			BaseDuration: 0.10,
			MinIntensity: 0.20,
			MaxIntensity: 0.85,
		}
	case "scifi":
		return damageFlashPreset{
			Color:        color.RGBA{R: 255, G: 220, B: 80, A: 255},
			BaseDuration: 0.12,
			MinIntensity: 0.20,
			MaxIntensity: 0.80,
		}
	case "postapoc":
		return damageFlashPreset{
			Color:        color.RGBA{R: 255, G: 160, B: 50, A: 255},
			BaseDuration: 0.15,
			MinIntensity: 0.25,
			MaxIntensity: 0.85,
		}
	default: // fantasy
		return damageFlashPreset{
			Color:        color.RGBA{R: 255, G: 240, B: 240, A: 255},
			BaseDuration: 0.12,
			MinIntensity: 0.20,
			MaxIntensity: 0.85,
		}
	}
}

// Update checks all entities for health decreases and triggers flash effects.
func (s *DamageFlashTintSystem) Update(entities []*Entity, deltaTime float64) {
	s.cleanupTimer += deltaTime

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil {
			continue
		}

		feedback := entity.GetVisualFeedback()
		if feedback == nil {
			continue
		}

		current := health.Current
		prev, tracked := s.prevHealth[entity.ID]

		if tracked && current < prev && health.Max > 0 {
			damage := prev - current
			proportion := damage / health.Max
			s.triggerFlash(feedback, proportion)
		}

		s.prevHealth[entity.ID] = current
	}

	// Periodically clean stale entries for removed entities
	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupTimer = 0
		s.cleanupStaleEntries(entities)
	}
}

// triggerFlash activates a damage flash with intensity based on damage proportion.
func (s *DamageFlashTintSystem) triggerFlash(feedback *VisualFeedbackComponent, proportion float64) {
	// Clamp proportion to [0,1]
	if proportion > 1.0 {
		proportion = 1.0
	}

	// Scale intensity: lerp between min and max based on damage proportion
	intensity := s.preset.MinIntensity + proportion*(s.preset.MaxIntensity-s.preset.MinIntensity)

	// Apply genre-colored tint via VisualFeedbackComponent
	r := float64(s.preset.Color.R) / 255.0
	g := float64(s.preset.Color.G) / 255.0
	b := float64(s.preset.Color.B) / 255.0

	// Blend: tint = lerp(white, genreColor, intensity)
	feedback.TintR = 1.0 - intensity*(1.0-r)
	feedback.TintG = 1.0 - intensity*(1.0-g)
	feedback.TintB = 1.0 - intensity*(1.0-b)
	feedback.TintA = 1.0

	// Scale flash duration: bigger hits flash longer (up to 2x base)
	durationScale := 1.0 + math.Min(proportion, 1.0)
	feedback.FlashDuration = s.preset.BaseDuration * durationScale
	feedback.TriggerFlash(intensity)
}

// cleanupStaleEntries removes health tracking for entities no longer in the world.
func (s *DamageFlashTintSystem) cleanupStaleEntries(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.prevHealth {
		if _, ok := active[id]; !ok {
			delete(s.prevHealth, id)
		}
	}
}
