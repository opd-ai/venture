// Package engine provides the WeatherSpriteTintSystem which applies subtle
// color tints to entity sprites based on active weather conditions.
// Rain adds a cool blue-dark tint, snow adds a frost-white tint, fog desaturates,
// and genre-specific weather types (blood rain, neon rain, ash) use thematic colors.
// Tints multiply with existing VisualFeedbackComponent tints so status effects
// and weather effects compose correctly.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherSpriteTintComponent stores weather-driven tint multipliers for an entity.
// These multiply with VisualFeedbackComponent tints in the render pipeline.
type WeatherSpriteTintComponent struct {
	// RGB multipliers (1.0 = no change). Values < 1.0 darken that channel.
	TintR float64
	TintG float64
	TintB float64
}

// Type returns the component type identifier.
func (w *WeatherSpriteTintComponent) Type() string {
	return "weather_sprite_tint"
}

// NewWeatherSpriteTintComponent creates a component with neutral (no-op) tint.
func NewWeatherSpriteTintComponent() *WeatherSpriteTintComponent {
	return &WeatherSpriteTintComponent{
		TintR: 1.0,
		TintG: 1.0,
		TintB: 1.0,
	}
}

// weatherTintPreset holds the RGB multipliers for a weather type.
type weatherTintPreset struct {
	R, G, B float64
}

// WeatherSpriteTintSystem reads the active weather and adjusts
// WeatherSpriteTintComponent values on entities with sprites.
type WeatherSpriteTintSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Throttle: only recompute when weather changes
	updateInterval float64
	timeSinceCheck float64

	// Cached weather state
	lastWeatherType   particles.WeatherType
	lastWeatherActive bool

	// Genre for intensity scaling
	genreID string

	// Pre-computed tint presets per weather type (genre-dependent)
	presets map[particles.WeatherType]weatherTintPreset
}

// NewWeatherSpriteTintSystem creates a new weather sprite tint system.
func NewWeatherSpriteTintSystem(world *World, seed int64) *WeatherSpriteTintSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_sprite_tint")
	}

	sys := &WeatherSpriteTintSystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  0.5,
		lastWeatherType: particles.WeatherRain,
		genreID:         "fantasy",
	}

	sys.buildPresets()

	if logEntry != nil {
		logEntry.Debug("WeatherSpriteTintSystem created")
	}
	return sys
}

// SetGenre configures genre-aware tint intensity and rebuilds presets.
func (s *WeatherSpriteTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.buildPresets()
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for weather sprite tint")
	}
}

// buildPresets computes tint presets for each weather type, scaled by genre.
func (s *WeatherSpriteTintSystem) buildPresets() {
	scale := s.genreIntensityScale()
	s.presets = map[particles.WeatherType]weatherTintPreset{
		particles.WeatherRain:      {R: 1.0 - 0.08*scale, G: 1.0 - 0.06*scale, B: 1.0},
		particles.WeatherSnow:      {R: 1.0, G: 1.0, B: 1.0 + 0.04*scale},
		particles.WeatherFog:       {R: 1.0 - 0.04*scale, G: 1.0 - 0.04*scale, B: 1.0 - 0.04*scale},
		particles.WeatherDust:      {R: 1.0, G: 1.0 - 0.04*scale, B: 1.0 - 0.10*scale},
		particles.WeatherAsh:       {R: 1.0 - 0.06*scale, G: 1.0 - 0.08*scale, B: 1.0 - 0.08*scale},
		particles.WeatherNeonRain:  {R: 1.0 - 0.02*scale, G: 1.0, B: 1.0 + 0.06*scale},
		particles.WeatherBloodRain: {R: 1.0, G: 1.0 - 0.12*scale, B: 1.0 - 0.12*scale},
	}
}

// genreIntensityScale returns a multiplier [0.5, 1.5] for genre-dependent tint strength.
func (s *WeatherSpriteTintSystem) genreIntensityScale() float64 {
	switch s.genreID {
	case "horror":
		return 1.4
	case "cyberpunk":
		return 1.2
	case "post-apocalyptic", "postapoc":
		return 1.3
	case "sci-fi", "scifi":
		return 1.0
	case "fantasy":
		return 0.8
	default:
		return 1.0
	}
}

// Update checks weather and applies tints to all sprite-bearing entities.
func (s *WeatherSpriteTintSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	weatherType, active := s.findActiveWeather(entities)

	weatherChanged := active != s.lastWeatherActive ||
		(active && weatherType != s.lastWeatherType)

	s.lastWeatherActive = active
	s.lastWeatherType = weatherType

	if !weatherChanged {
		return
	}

	s.applyTints(entities, weatherType, active)
}

// findActiveWeather locates the first active weather entity.
func (s *WeatherSpriteTintSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		return weather.Config.Type, true
	}
	return particles.WeatherRain, false
}

// applyTints sets or clears weather tints on entities with sprites.
func (s *WeatherSpriteTintSystem) applyTints(entities []*Entity, weatherType particles.WeatherType, active bool) {
	for _, entity := range entities {
		if !entity.HasComponent("sprite") {
			continue
		}

		tintComp := s.getOrCreateTintComponent(entity)

		if !active {
			tintComp.TintR = 1.0
			tintComp.TintG = 1.0
			tintComp.TintB = 1.0
			continue
		}

		preset, ok := s.presets[weatherType]
		if !ok {
			preset = weatherTintPreset{R: 1.0, G: 1.0, B: 1.0}
		}

		tintComp.TintR = clampTint(preset.R)
		tintComp.TintG = clampTint(preset.G)
		tintComp.TintB = clampTint(preset.B)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"weather_type": weatherType.String(),
			"active":       active,
			"genre":        s.genreID,
		}).Debug("weather sprite tints updated")
	}
}

// getOrCreateTintComponent retrieves or lazily creates the tint component.
func (s *WeatherSpriteTintSystem) getOrCreateTintComponent(entity *Entity) *WeatherSpriteTintComponent {
	comp, ok := entity.GetComponent("weather_sprite_tint")
	if ok {
		if tintComp, ok := comp.(*WeatherSpriteTintComponent); ok {
			return tintComp
		}
	}
	tintComp := NewWeatherSpriteTintComponent()
	entity.AddComponent(tintComp)
	return tintComp
}

// clampTint ensures a tint multiplier stays in [0.5, 1.1].
func clampTint(v float64) float64 {
	if v < 0.5 {
		return 0.5
	}
	if v > 1.1 {
		return 1.1
	}
	return v
}
