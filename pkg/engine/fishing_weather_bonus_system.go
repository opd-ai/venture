// Package engine provides the FishingWeatherBonusSystem which bridges weather conditions
// with fishing mechanics. This system modifies fishing bonuses based on active weather,
// making rain boost fish activity, storms attract rare fish, and clear skies reduce catches.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// FishingWeatherBonusSystem modifies fishing spot bonuses based on weather conditions.
// Rain increases rare fish chance by 20-40%, thunderstorms by 50-80%, fog reduces by 10%.
// This connects WeatherSystem with FishingSystem to create immersive fishing gameplay.
type FishingWeatherBonusSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original rare fish bonuses to restore when weather clears
	originalBonuses map[uint64]float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre affects which weather types boost fishing
	genreID string

	// fishingSystem reference for updating weather callback
	fishingSystem *FishingSystem
}

// NewFishingWeatherBonusSystem creates a new fishing weather bonus system.
func NewFishingWeatherBonusSystem(world *World, seed int64) *FishingWeatherBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "fishing_weather_bonus")
		logEntry.Debug("FishingWeatherBonusSystem created")
	}

	return &FishingWeatherBonusSystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		originalBonuses:      make(map[uint64]float64),
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
		genreID:              "fantasy",
	}
}

// SetGenre sets the genre for genre-aware fishing bonuses.
func (s *FishingWeatherBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for fishing weather bonus")
	}
}

// SetFishingSystem sets the reference to FishingSystem for weather callback updates.
func (s *FishingWeatherBonusSystem) SetFishingSystem(fs *FishingSystem) {
	s.fishingSystem = fs
	if fs != nil {
		// Set up the weather callback to use our weather detection
		fs.CurrentWeather = s.getCurrentWeatherString
	}
}

// Update checks weather conditions and modifies fishing spot bonuses.
func (s *FishingWeatherBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.world == nil {
		return
	}

	// Find active weather
	weatherType, intensity, active := s.findActiveWeather(entities)

	// If weather state changed, update fishing bonuses
	if active != s.lastWeatherActive ||
		(active && (weatherType != s.lastWeatherType || intensity != s.lastWeatherIntensity)) {
		s.updateFishingBonuses(entities, weatherType, intensity, active)
		s.lastWeatherActive = active
		s.lastWeatherType = weatherType
		s.lastWeatherIntensity = intensity

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"weather_type":   weatherType.String(),
				"intensity":      intensity.String(),
				"active":         active,
				"spots_affected": len(s.originalBonuses),
			}).Debug("Weather changed, fishing bonuses updated")
		}
	}
}

// findActiveWeather locates active weather from entities.
func (s *FishingWeatherBonusSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}

		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}

		return weather.Config.Type, weather.Config.Intensity, true
	}

	return particles.WeatherRain, particles.IntensityLight, false
}

// updateFishingBonuses modifies all fishing spot bonuses based on weather.
func (s *FishingWeatherBonusSystem) updateFishingBonuses(entities []*Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("fishing_spot")
		if !ok {
			continue
		}

		spot, ok := comp.(*FishingSpotComponent)
		if !ok {
			continue
		}

		// Store original bonus if not already stored
		if _, exists := s.originalBonuses[entity.ID]; !exists {
			s.originalBonuses[entity.ID] = spot.RareFishBonus
		}

		if active {
			// Apply weather-based modifier
			modifier := s.calculateBonusModifier(weatherType, intensity, spot.WaterType)
			spot.RareFishBonus = s.originalBonuses[entity.ID] * modifier
		} else {
			// Restore original bonus
			spot.RareFishBonus = s.originalBonuses[entity.ID]
		}
	}
}

// calculateBonusModifier returns the rare fish bonus multiplier based on weather.
func (s *FishingWeatherBonusSystem) calculateBonusModifier(weather particles.WeatherType, intensity particles.WeatherIntensity, waterType WaterType) float64 {
	baseModifier := s.getBaseModifier(weather, waterType)
	intensityMultiplier := s.getIntensityMultiplier(intensity)
	genreBonus := s.getGenreBonus(weather)

	return baseModifier * intensityMultiplier * genreBonus
}

// getBaseModifier returns the base bonus modifier for weather type and water type.
func (s *FishingWeatherBonusSystem) getBaseModifier(weather particles.WeatherType, waterType WaterType) float64 {
	switch weather {
	case particles.WeatherRain:
		// Rain excites fish, especially in freshwater
		if waterType == WaterTypeFreshwater {
			return 1.4 // 40% bonus
		}
		return 1.25 // 25% bonus for other water types

	case particles.WeatherSnow:
		// Cold weather reduces fish activity
		return 0.85

	case particles.WeatherFog:
		// Fog has minimal effect but slight reduction in visibility
		return 0.95

	case particles.WeatherDust, particles.WeatherSandstorm:
		// Dust storms reduce fishing success
		return 0.7

	case particles.WeatherAsh:
		// Ash can attract certain fish types
		return 1.1

	case particles.WeatherNeonRain:
		// Magical/tech rain attracts magical fish
		if waterType == WaterTypeMagical {
			return 1.6
		}
		return 1.2

	case particles.WeatherSmog:
		// Pollution reduces fish activity
		return 0.6

	case particles.WeatherRadiation:
		// Radiation attracts mutated/rare fish
		return 1.5

	case particles.WeatherBloodRain:
		// Horror rain attracts rare/legendary fish
		return 1.8

	default:
		return 1.0
	}
}

// getIntensityMultiplier returns multiplier based on weather intensity.
func (s *FishingWeatherBonusSystem) getIntensityMultiplier(intensity particles.WeatherIntensity) float64 {
	switch intensity {
	case particles.IntensityLight:
		return 0.8
	case particles.IntensityMedium:
		return 1.0
	case particles.IntensityHeavy:
		return 1.3
	case particles.IntensityExtreme:
		return 1.6
	default:
		return 1.0
	}
}

// getGenreBonus returns additional bonus based on genre synergies.
func (s *FishingWeatherBonusSystem) getGenreBonus(weather particles.WeatherType) float64 {
	switch s.genreID {
	case "fantasy":
		// Fantasy benefits most from rain and magical weather
		if weather == particles.WeatherRain || weather == particles.WeatherNeonRain {
			return 1.15
		}
	case "scifi":
		// Sci-fi benefits from radiation and neon
		if weather == particles.WeatherRadiation || weather == particles.WeatherNeonRain {
			return 1.2
		}
	case "horror":
		// Horror benefits from blood rain and fog
		if weather == particles.WeatherBloodRain || weather == particles.WeatherFog {
			return 1.25
		}
	case "cyberpunk":
		// Cyberpunk benefits from neon rain and smog
		if weather == particles.WeatherNeonRain || weather == particles.WeatherSmog {
			return 1.2
		}
	case "postapoc":
		// Post-apocalyptic benefits from radiation and sandstorms
		if weather == particles.WeatherRadiation || weather == particles.WeatherSandstorm {
			return 1.15
		}
	}
	return 1.0
}

// getCurrentWeatherString returns the current weather as a string for FishingSystem.
func (s *FishingWeatherBonusSystem) getCurrentWeatherString() string {
	if !s.lastWeatherActive {
		return "clear"
	}
	return s.lastWeatherType.String()
}

// GetCurrentWeatherType returns the current weather type for external queries.
func (s *FishingWeatherBonusSystem) GetCurrentWeatherType() (particles.WeatherType, bool) {
	return s.lastWeatherType, s.lastWeatherActive
}

// GetBonusModifier returns the current bonus modifier for a given water type.
// Useful for UI display or debugging.
func (s *FishingWeatherBonusSystem) GetBonusModifier(waterType WaterType) float64 {
	if !s.lastWeatherActive {
		return 1.0
	}
	return s.calculateBonusModifier(s.lastWeatherType, s.lastWeatherIntensity, waterType)
}
