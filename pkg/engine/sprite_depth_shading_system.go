// Package engine provides the SpriteDepthShadingSystem which manages per-entity
// depth shading parameters. It attaches SpriteDepthShadingComponent to entities
// that have sprites, configures genre-aware shading tints, and marks entities
// dirty when genre changes require sprite regeneration.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// spriteDepthGenrePreset maps genre to light tint and intensity.
type spriteDepthGenrePreset struct {
	TintR, TintG, TintB float64
	LightIntensity      float64
	EdgeDarkening       float64
	AmbientOcclusion    float64
	DitherStrength      float64
}

// SpriteDepthShadingSystem scans entities with sprite components and attaches/updates
// SpriteDepthShadingComponent with genre-aware shading parameters.
type SpriteDepthShadingSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  spriteDepthGenrePreset

	scanInterval  float64
	timeSinceScan float64
}

// NewSpriteDepthShadingSystem creates a new sprite depth shading system.
func NewSpriteDepthShadingSystem(world *World, seed int64) *SpriteDepthShadingSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "sprite_depth_shading")
		logEntry.Debug("sprite depth shading system created")
	}

	sys := &SpriteDepthShadingSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
	sys.preset = sys.getGenrePreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware shading parameters.
func (s *SpriteDepthShadingSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for depth shading")
	}
}

// getGenrePreset returns genre-specific shading parameters.
func (s *SpriteDepthShadingSystem) getGenrePreset(genreID string) spriteDepthGenrePreset {
	switch genreID {
	case "horror":
		return spriteDepthGenrePreset{
			TintR: 0.85, TintG: 0.75, TintB: 0.70,
			LightIntensity: 0.20, EdgeDarkening: 0.40,
			AmbientOcclusion: 0.30, DitherStrength: 0.06,
		}
	case "cyberpunk":
		return spriteDepthGenrePreset{
			TintR: 0.80, TintG: 0.95, TintB: 1.0,
			LightIntensity: 0.40, EdgeDarkening: 0.30,
			AmbientOcclusion: 0.15, DitherStrength: 0.08,
		}
	case "sci-fi", "scifi":
		return spriteDepthGenrePreset{
			TintR: 0.90, TintG: 0.95, TintB: 1.0,
			LightIntensity: 0.38, EdgeDarkening: 0.22,
			AmbientOcclusion: 0.15, DitherStrength: 0.06,
		}
	case "post-apocalyptic", "postapoc":
		return spriteDepthGenrePreset{
			TintR: 1.0, TintG: 0.90, TintB: 0.75,
			LightIntensity: 0.28, EdgeDarkening: 0.35,
			AmbientOcclusion: 0.20, DitherStrength: 0.10,
		}
	default: // fantasy
		return spriteDepthGenrePreset{
			TintR: 1.0, TintG: 0.98, TintB: 0.90,
			LightIntensity: 0.35, EdgeDarkening: 0.25,
			AmbientOcclusion: 0.15, DitherStrength: 0.06,
		}
	}
}

// Update scans entities and attaches/configures shading components.
func (s *SpriteDepthShadingSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		// Only process entities with sprites
		if !entity.HasComponent("sprite") {
			continue
		}

		comp, has := entity.GetComponent("sprite_depth_shading")
		if !has {
			// Attach new shading component
			shading := NewSpriteDepthShadingComponent()
			s.applyPreset(shading)
			entity.AddComponent(shading)
			continue
		}

		shading, ok := comp.(*SpriteDepthShadingComponent)
		if !ok {
			continue
		}

		// Update shading if genre preset differs
		if !s.presetMatches(shading) {
			s.applyPreset(shading)
			shading.Dirty = true
		}
	}
}

// applyPreset writes current genre preset into a shading component.
func (s *SpriteDepthShadingSystem) applyPreset(c *SpriteDepthShadingComponent) {
	c.LightIntensity = s.preset.LightIntensity
	c.EdgeDarkening = s.preset.EdgeDarkening
	c.AmbientOcclusion = s.preset.AmbientOcclusion
	c.DitherStrength = s.preset.DitherStrength
	c.TintR = s.preset.TintR
	c.TintG = s.preset.TintG
	c.TintB = s.preset.TintB
	c.Enabled = true
}

// presetMatches returns true if the component already has current genre values.
func (s *SpriteDepthShadingSystem) presetMatches(c *SpriteDepthShadingComponent) bool {
	const eps = 0.001
	return approxEqual(c.TintR, s.preset.TintR, eps) &&
		approxEqual(c.TintG, s.preset.TintG, eps) &&
		approxEqual(c.TintB, s.preset.TintB, eps) &&
		approxEqual(c.LightIntensity, s.preset.LightIntensity, eps)
}

func approxEqual(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}
