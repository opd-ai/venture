// Package engine provides the WeatherSpellDamageSystem that connects weather conditions
// with spell damage calculations for elemental synergies.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherSpellDamageSystem modifies spell damage based on current weather conditions.
// Creates elemental synergies: fire spells deal more damage in dry weather, ice spells
// in snow, lightning in rain, etc. Genre-aware bonuses apply (cyberpunk neon rain
// enhances lightning more than standard rain).
//
// This system bridges WeatherComponent weather types with SpellCastingSystem damage.
type WeatherSpellDamageSystem struct {
	world    *World
	rng      *rand.Rand
	genreID  string
	logger   *logrus.Entry
	modCache map[weatherElementKey]float64 // Cache computed modifiers
}

// weatherElementKey combines weather and element for cache lookup.
type weatherElementKey struct {
	weather particles.WeatherType
	element magic.ElementType
}

// NewWeatherSpellDamageSystem creates a new weather-spell damage modifier system.
func NewWeatherSpellDamageSystem(world *World, seed int64) *WeatherSpellDamageSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_spell_damage",
		"seed":        seed,
	})
	logger.Debug("Creating weather spell damage system")

	return &WeatherSpellDamageSystem{
		world:    world,
		rng:      rand.New(rand.NewSource(seed)),
		genreID:  "fantasy",
		logger:   logger,
		modCache: make(map[weatherElementKey]float64),
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherSpellDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.modCache = make(map[weatherElementKey]float64) // Clear cache on genre change
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather spell damage")
	}
}

// Update processes weather-spell damage modifiers for all relevant entities.
// This system doesn't modify entities directly; it provides damage modifiers
// via GetDamageModifier that SpellCastingSystem can query.
func (s *WeatherSpellDamageSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is query-based, not update-based.
	// Damage modifiers are computed on-demand via GetDamageModifier.
}

// GetDamageModifier returns the damage multiplier for a spell element given current weather.
// Returns 1.0 if no weather is active or no synergy exists.
// Values > 1.0 indicate bonus damage, < 1.0 indicate reduced damage.
func (s *WeatherSpellDamageSystem) GetDamageModifier(element magic.ElementType) float64 {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 1.0
	}

	key := weatherElementKey{weather: weather.Config.Type, element: element}
	if cached, ok := s.modCache[key]; ok {
		return cached
	}

	modifier := s.calculateModifier(weather.Config.Type, element)
	s.modCache[key] = modifier

	if s.logger != nil && modifier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"weather":  weather.Config.Type.String(),
			"element":  element.String(),
			"modifier": modifier,
		}).Debug("Weather spell damage modifier calculated")
	}

	return modifier
}

// getCurrentWeather finds the active weather component in the world.
func (s *WeatherSpellDamageSystem) getCurrentWeather() *WeatherComponent {
	if s.world == nil {
		return nil
	}
	for _, entity := range s.world.GetEntities() {
		if comp, ok := entity.GetComponent("weather"); ok {
			if weather, ok := comp.(*WeatherComponent); ok && weather.Active {
				return weather
			}
		}
	}
	return nil
}

// calculateModifier computes the damage modifier for a weather-element combination.
func (s *WeatherSpellDamageSystem) calculateModifier(weather particles.WeatherType, element magic.ElementType) float64 {
	baseModifier := s.getBaseModifier(weather, element)
	genreBonus := s.getGenreBonus(weather, element)
	// Round to 2 decimal places to avoid IEEE 754 floating-point addition artifacts
	return math.Round((baseModifier+genreBonus)*100) / 100
}

// getBaseModifier returns the base damage modifier for weather-element synergies.
func (s *WeatherSpellDamageSystem) getBaseModifier(weather particles.WeatherType, element magic.ElementType) float64 {
	switch weather {
	case particles.WeatherRain:
		return s.getRainModifier(element)
	case particles.WeatherSnow:
		return s.getSnowModifier(element)
	case particles.WeatherFog:
		return s.getFogModifier(element)
	case particles.WeatherDust, particles.WeatherSandstorm:
		return s.getDustModifier(element)
	case particles.WeatherAsh:
		return s.getAshModifier(element)
	case particles.WeatherNeonRain:
		return s.getNeonRainModifier(element)
	case particles.WeatherSmog:
		return s.getSmogModifier(element)
	case particles.WeatherRadiation:
		return s.getRadiationModifier(element)
	case particles.WeatherBloodRain:
		return s.getBloodRainModifier(element)
	default:
		return 1.0
	}
}

// getRainModifier returns modifiers for rain weather.
func (s *WeatherSpellDamageSystem) getRainModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementLightning:
		return 1.25 // Lightning conducts through water
	case magic.ElementFire:
		return 0.75 // Fire dampened by rain
	case magic.ElementIce:
		return 1.10 // Cold synergy with wet conditions
	default:
		return 1.0
	}
}

// getSnowModifier returns modifiers for snow weather.
func (s *WeatherSpellDamageSystem) getSnowModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementIce:
		return 1.30 // Ice empowered in cold
	case magic.ElementFire:
		return 0.80 // Fire weakened in cold
	case magic.ElementWind:
		return 1.15 // Blizzard synergy
	default:
		return 1.0
	}
}

// getFogModifier returns modifiers for fog weather.
func (s *WeatherSpellDamageSystem) getFogModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementDark:
		return 1.20 // Darkness thrives in obscured vision
	case magic.ElementLight:
		return 0.85 // Light diffused by fog
	case magic.ElementWind:
		return 1.10 // Wind disperses fog - thematic synergy
	default:
		return 1.0
	}
}

// getDustModifier returns modifiers for dust/sandstorm weather.
func (s *WeatherSpellDamageSystem) getDustModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementEarth:
		return 1.25 // Earth empowered by particulates
	case magic.ElementFire:
		return 1.15 // Dust is flammable
	case magic.ElementLightning:
		return 0.80 // Static interference
	default:
		return 1.0
	}
}

// getAshModifier returns modifiers for ash weather.
func (s *WeatherSpellDamageSystem) getAshModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementFire:
		return 1.20 // Ash feeds fire
	case magic.ElementDark:
		return 1.15 // Darkness in the ash cloud
	case magic.ElementLight:
		return 0.85 // Light obscured
	default:
		return 1.0
	}
}

// getNeonRainModifier returns modifiers for cyberpunk neon rain.
func (s *WeatherSpellDamageSystem) getNeonRainModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementLightning:
		return 1.35 // Highly conductive tech-rain
	case magic.ElementArcane:
		return 1.20 // Tech-magic synergy
	case magic.ElementFire:
		return 0.70 // Suppressed by acidic rain
	default:
		return 1.0
	}
}

// getSmogModifier returns modifiers for industrial smog.
func (s *WeatherSpellDamageSystem) getSmogModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementDark:
		return 1.25 // Darkness thrives in smog
	case magic.ElementFire:
		return 1.10 // Smog is flammable
	case magic.ElementWind:
		return 1.15 // Wind disperses smog
	case magic.ElementLight:
		return 0.80 // Light choked by smog
	default:
		return 1.0
	}
}

// getRadiationModifier returns modifiers for radioactive weather.
func (s *WeatherSpellDamageSystem) getRadiationModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementArcane:
		return 1.30 // Magical energy resonates with radiation
	case magic.ElementLight:
		return 1.20 // Radiation emits light
	case magic.ElementDark:
		return 0.85 // Radiation opposes shadow
	default:
		return 1.0
	}
}

// getBloodRainModifier returns modifiers for horror blood rain.
func (s *WeatherSpellDamageSystem) getBloodRainModifier(element magic.ElementType) float64 {
	switch element {
	case magic.ElementDark:
		return 1.35 // Dark magic empowered by horror atmosphere
	case magic.ElementLight:
		return 0.70 // Light struggles against horror
	case magic.ElementFire:
		return 0.90 // Blood dampens fire somewhat
	default:
		return 1.0
	}
}

// getGenreBonus returns additional bonus based on genre-weather-element combinations.
func (s *WeatherSpellDamageSystem) getGenreBonus(weather particles.WeatherType, element magic.ElementType) float64 {
	switch s.genreID {
	case "cyberpunk":
		if weather == particles.WeatherNeonRain && element == magic.ElementLightning {
			return 0.10 // Extra bonus in cyberpunk
		}
		if element == magic.ElementArcane {
			return 0.05 // Tech-magic synergy
		}
	case "horror":
		if weather == particles.WeatherBloodRain && element == magic.ElementDark {
			return 0.15 // Extra dark bonus in horror
		}
		if weather == particles.WeatherFog && element == magic.ElementDark {
			return 0.10 // Fog + dark in horror
		}
	case "postapoc":
		if weather == particles.WeatherRadiation && element == magic.ElementArcane {
			return 0.10 // Wasteland magic boost
		}
		if weather == particles.WeatherSandstorm && element == magic.ElementEarth {
			return 0.10 // Desert earth magic
		}
	case "scifi":
		if element == magic.ElementLightning || element == magic.ElementArcane {
			return 0.05 // Tech-based elements boosted
		}
	case "fantasy":
		// Base modifiers are already balanced for fantasy
		return 0.0
	}
	return 0.0
}

// GetCurrentWeatherType returns the current weather type for external queries.
func (s *WeatherSpellDamageSystem) GetCurrentWeatherType() (particles.WeatherType, bool) {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 0, false
	}
	return weather.Config.Type, true
}

// ClearCache clears the modifier cache, useful when weather changes.
func (s *WeatherSpellDamageSystem) ClearCache() {
	s.modCache = make(map[weatherElementKey]float64)
}
