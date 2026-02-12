// Package engine provides the WeatherAwareAISystem which bridges weather conditions
// and AI behavior. This system modifies AI detection ranges based on active weather,
// making stealth gameplay viable during fog, rain, or storms.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherAwareAISystem modifies AI detection ranges based on weather conditions.
// Rain reduces detection by 15-25%, fog by 30-50%, storms by 40-60%, etc.
// This connects the WeatherSystem with the AISystem to create tactical gameplay.
type WeatherAwareAISystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache the original detection ranges to restore when weather clears
	originalRanges map[uint64]float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool
}

// NewWeatherAwareAISystem creates a new weather-aware AI system.
func NewWeatherAwareAISystem(world *World, seed int64) *WeatherAwareAISystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_aware_ai")
		logEntry.Debug("WeatherAwareAISystem created")
	}

	return &WeatherAwareAISystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		originalRanges:       make(map[uint64]float64),
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
	}
}

// Update checks weather conditions and modifies AI detection ranges.
func (w *WeatherAwareAISystem) Update(entities []*Entity, deltaTime float64) {
	w.timeSinceCheck += deltaTime

	if w.timeSinceCheck < w.updateInterval {
		return
	}
	w.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := w.findActiveWeather(entities)

	// If weather state changed, update AI detection ranges
	if active != w.lastWeatherActive || (active && (weatherType != w.lastWeatherType || intensity != w.lastWeatherIntensity)) {
		w.updateAIDetectionRanges(entities, weatherType, intensity, active)
		w.lastWeatherActive = active
		w.lastWeatherType = weatherType
		w.lastWeatherIntensity = intensity
	}
}

// findActiveWeather locates active weather from entities.
func (w *WeatherAwareAISystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// updateAIDetectionRanges modifies detection ranges based on weather.
func (w *WeatherAwareAISystem) updateAIDetectionRanges(entities []*Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active bool) {
	for _, entity := range entities {
		if !entity.HasComponent("ai") {
			continue
		}

		aiComp, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}

		ai, ok := aiComp.(*AIComponent)
		if !ok {
			continue
		}

		if active {
			// Store original range if not already stored
			if _, exists := w.originalRanges[entity.ID]; !exists {
				w.originalRanges[entity.ID] = ai.DetectionRange
			}

			// Calculate visibility multiplier based on weather and intensity
			multiplier := w.calculateVisibilityMultiplier(weatherType, intensity)

			// Apply modified detection range
			originalRange := w.originalRanges[entity.ID]
			newRange := originalRange * multiplier
			ai.DetectionRange = newRange

			if w.logger != nil {
				w.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"weather_type":   weatherType.String(),
					"intensity":      intensity.String(),
					"original_range": originalRange,
					"modified_range": newRange,
					"multiplier":     multiplier,
				}).Debug("AI detection range reduced by weather")
			}
		} else {
			// Restore original detection range
			if originalRange, exists := w.originalRanges[entity.ID]; exists {
				ai.DetectionRange = originalRange
				delete(w.originalRanges, entity.ID)

				if w.logger != nil {
					w.logger.WithFields(logrus.Fields{
						"entity_id":      entity.ID,
						"restored_range": originalRange,
					}).Debug("AI detection range restored")
				}
			}
		}
	}
}

// calculateVisibilityMultiplier returns a multiplier (0.0-1.0) for detection range.
// Lower values mean worse visibility.
func (w *WeatherAwareAISystem) calculateVisibilityMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	// Base reduction per weather type
	var baseReduction float64
	switch weatherType {
	case particles.WeatherFog, particles.WeatherSmog:
		// Fog and smog heavily reduce visibility
		baseReduction = 0.40
	case particles.WeatherSandstorm:
		// Sandstorms are blinding
		baseReduction = 0.50
	case particles.WeatherRain, particles.WeatherNeonRain, particles.WeatherBloodRain:
		// Rain moderately reduces visibility
		baseReduction = 0.20
	case particles.WeatherSnow:
		// Snow moderately reduces visibility
		baseReduction = 0.25
	case particles.WeatherDust, particles.WeatherAsh, particles.WeatherRadiation:
		// Light particles have minor effect
		baseReduction = 0.15
	default:
		baseReduction = 0.10
	}

	// Intensity multiplier
	var intensityMult float64
	switch intensity {
	case particles.IntensityLight:
		intensityMult = 0.5
	case particles.IntensityMedium:
		intensityMult = 1.0
	case particles.IntensityHeavy:
		intensityMult = 1.5
	default:
		intensityMult = 1.0
	}

	// Calculate total reduction
	totalReduction := baseReduction * intensityMult
	if totalReduction > 0.7 {
		totalReduction = 0.7 // Cap at 70% reduction (30% visibility minimum)
	}

	return 1.0 - totalReduction
}

// SetUpdateInterval configures how often weather is checked.
func (w *WeatherAwareAISystem) SetUpdateInterval(seconds float64) {
	if seconds > 0 {
		w.updateInterval = seconds
	}
}

// GetVisibilityMultiplier returns the current visibility multiplier for testing.
func (w *WeatherAwareAISystem) GetVisibilityMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	return w.calculateVisibilityMultiplier(weatherType, intensity)
}
