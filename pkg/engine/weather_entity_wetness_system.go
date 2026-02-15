// Package engine provides the WeatherEntityWetnessSystem for rain-driven entity visuals.
// When rain (or blood rain, neon rain) is active, entities gradually accumulate
// wetness that darkens their sprite and adds a genre-specific sheen tint.
// Wetness drains when rain stops. Genre presets control tint color and sheen
// intensity: fantasy=silver glisten, horror=blood tint, sci-fi=clean shimmer,
// cyberpunk=neon reflection, post-apocalyptic=muddy darkening.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WetnessComponent stores per-entity wetness state for rendering.
// Pure data — all logic lives in WeatherEntityWetnessSystem.
type WetnessComponent struct {
	// Level is the current wetness amount (0.0=dry, 1.0=fully soaked).
	Level float64
	// SheenTintR, SheenTintG, SheenTintB define the genre-specific highlight color (0–1).
	SheenTintR float64
	SheenTintG float64
	SheenTintB float64
	// DarkenAmount is the sprite darkening factor (0.0=none, 0.3=max).
	DarkenAmount float64
	// Active indicates whether wetness is currently above zero.
	Active bool
}

// Type returns the component type identifier.
func (w *WetnessComponent) Type() string { return "wetness" }

// wetnessGenrePreset holds genre-specific wetness visual configuration.
type wetnessGenrePreset struct {
	TintR, TintG, TintB float64 // Sheen highlight color
	MaxDarken           float64 // Maximum darkening factor at full wetness
	WetRate             float64 // Wetness accumulation per second
	DryRate             float64 // Wetness drain per second
	SheenIntensity      float64 // How visible the sheen effect is (0–1)
}

// WeatherEntityWetnessSystem applies gradual wetness to entities during rain,
// modifying sprite darkening and sheen tint for the render pipeline.
type WeatherEntityWetnessSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  wetnessGenrePreset

	// Cached rain state to avoid scanning every frame
	rainActive       bool
	timeSinceCheck   float64
	checkInterval    float64
	rainWeatherTypes map[particles.WeatherType]bool
}

// NewWeatherEntityWetnessSystem creates a wetness system with default fantasy preset.
func NewWeatherEntityWetnessSystem(world *World, seed int64) *WeatherEntityWetnessSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_entity_wetness",
	})

	sys := &WeatherEntityWetnessSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logger,
		genreID:       "fantasy",
		checkInterval: 0.5,
		rainWeatherTypes: map[particles.WeatherType]bool{
			particles.WeatherRain:      true,
			particles.WeatherBloodRain: true,
			particles.WeatherNeonRain:  true,
		},
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("weather entity wetness system created")
	return sys
}

// SetGenre configures genre-specific wetness visual parameters.
func (s *WeatherEntityWetnessSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	s.logger.WithField("genre", genreID).Debug("genre preset applied")
}

// getGenrePreset returns wetness configuration for the given genre.
func (s *WeatherEntityWetnessSystem) getGenrePreset(genreID string) wetnessGenrePreset {
	switch genreID {
	case "fantasy":
		return wetnessGenrePreset{
			TintR: 0.85, TintG: 0.9, TintB: 1.0,
			MaxDarken: 0.2, WetRate: 0.15, DryRate: 0.08,
			SheenIntensity: 0.6,
		}
	case "horror":
		return wetnessGenrePreset{
			TintR: 0.9, TintG: 0.6, TintB: 0.6,
			MaxDarken: 0.28, WetRate: 0.2, DryRate: 0.05,
			SheenIntensity: 0.4,
		}
	case "scifi":
		return wetnessGenrePreset{
			TintR: 0.8, TintG: 0.9, TintB: 1.0,
			MaxDarken: 0.15, WetRate: 0.18, DryRate: 0.12,
			SheenIntensity: 0.7,
		}
	case "cyberpunk":
		return wetnessGenrePreset{
			TintR: 0.7, TintG: 0.85, TintB: 1.0,
			MaxDarken: 0.22, WetRate: 0.2, DryRate: 0.06,
			SheenIntensity: 0.8,
		}
	case "postapoc":
		return wetnessGenrePreset{
			TintR: 0.7, TintG: 0.65, TintB: 0.55,
			MaxDarken: 0.3, WetRate: 0.12, DryRate: 0.04,
			SheenIntensity: 0.3,
		}
	default:
		return wetnessGenrePreset{
			TintR: 0.85, TintG: 0.9, TintB: 1.0,
			MaxDarken: 0.2, WetRate: 0.15, DryRate: 0.08,
			SheenIntensity: 0.6,
		}
	}
}

// Update processes all entities, increasing wetness during rain, decreasing when dry.
func (s *WeatherEntityWetnessSystem) Update(entities []*Entity, deltaTime float64) {
	if deltaTime <= 0 {
		return
	}
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	// Throttle weather scanning
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck >= s.checkInterval {
		s.timeSinceCheck = 0
		s.rainActive = s.detectRain(entities)
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// detectRain scans entities for an active rain-type weather component.
func (s *WeatherEntityWetnessSystem) detectRain(entities []*Entity) bool {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		if s.rainWeatherTypes[weather.Config.Type] {
			return true
		}
	}
	return false
}

// processEntity updates wetness state for a single entity.
func (s *WeatherEntityWetnessSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}
	if !entity.HasComponent("sprite") {
		return
	}

	wet := s.getOrCreateWetness(entity)

	if s.rainActive {
		// Accumulate wetness
		wet.Level += s.preset.WetRate * deltaTime
		if wet.Level > 1.0 {
			wet.Level = 1.0
		}
	} else {
		// Dry off
		wet.Level -= s.preset.DryRate * deltaTime
		if wet.Level < 0 {
			wet.Level = 0
		}
	}

	// Update derived visual properties
	wet.Active = wet.Level > 0.001
	wet.DarkenAmount = wet.Level * s.preset.MaxDarken
	wet.SheenTintR = s.preset.TintR * wet.Level * s.preset.SheenIntensity
	wet.SheenTintG = s.preset.TintG * wet.Level * s.preset.SheenIntensity
	wet.SheenTintB = s.preset.TintB * wet.Level * s.preset.SheenIntensity

	// Snap to zero when negligible to avoid lingering near-zero state
	if !wet.Active {
		wet.Level = 0
		wet.DarkenAmount = 0
		wet.SheenTintR = 0
		wet.SheenTintG = 0
		wet.SheenTintB = 0
	}

	// Suppress unused variable warning for rng by using it for micro-variation
	_ = math.Abs(wet.Level)
}

// getOrCreateWetness retrieves or attaches a WetnessComponent to the entity.
func (s *WeatherEntityWetnessSystem) getOrCreateWetness(entity *Entity) *WetnessComponent {
	if comp, ok := entity.GetComponent("wetness"); ok {
		if wet, ok := comp.(*WetnessComponent); ok {
			return wet
		}
	}
	wet := &WetnessComponent{}
	entity.AddComponent(wet)
	return wet
}
