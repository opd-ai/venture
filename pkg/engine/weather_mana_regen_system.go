// Package engine provides the WeatherManaRegenSystem which bridges weather conditions
// and mana regeneration. This system modifies mana regen rates based on active weather,
// allowing for genre-aware magical gameplay where rain boosts water magic users,
// thunderstorms aid lightning casters, etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherManaRegenSystem modifies mana regeneration based on weather conditions.
// Rain boosts regen by 15-30%, thunderstorms by 40-60%, fog reduces by 10-20%, etc.
// This connects the WeatherSystem with the ManaRegenSystem to create magical gameplay.
type WeatherManaRegenSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache the original regen rates to restore when weather clears
	originalRegen map[uint64]float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre affects which weather types boost which magic
	genreID string
}

// NewWeatherManaRegenSystem creates a new weather mana regen system.
func NewWeatherManaRegenSystem(world *World, seed int64) *WeatherManaRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_mana_regen")
		logEntry.Debug("WeatherManaRegenSystem created")
	}

	return &WeatherManaRegenSystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		originalRegen:        make(map[uint64]float64),
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
		genreID:              "fantasy",
	}
}

// SetGenre sets the genre for genre-aware mana modifiers.
func (w *WeatherManaRegenSystem) SetGenre(genreID string) {
	w.genreID = genreID
	if w.logger != nil {
		w.logger.WithField("genre", genreID).Debug("Genre set for weather mana regen")
	}
}

// Update checks weather conditions and modifies mana regeneration rates.
func (w *WeatherManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
	w.timeSinceCheck += deltaTime

	if w.timeSinceCheck < w.updateInterval {
		return
	}
	w.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := w.findActiveWeather(entities)

	// If weather state changed, update mana regen rates
	if active != w.lastWeatherActive || (active && (weatherType != w.lastWeatherType || intensity != w.lastWeatherIntensity)) {
		w.updateManaRegenRates(entities, weatherType, intensity, active)
		w.lastWeatherActive = active
		w.lastWeatherType = weatherType
		w.lastWeatherIntensity = intensity
	}
}

// findActiveWeather locates active weather from entities.
func (w *WeatherManaRegenSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// updateManaRegenRates modifies mana regeneration rates based on weather.
func (w *WeatherManaRegenSystem) updateManaRegenRates(entities []*Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active bool) {
	for _, entity := range entities {
		if !entity.HasComponent("mana") {
			continue
		}

		manaComp, ok := entity.GetComponent("mana")
		if !ok {
			continue
		}

		mana, ok := manaComp.(*ManaComponent)
		if !ok {
			continue
		}

		if active {
			// Store original regen if not already stored
			if _, exists := w.originalRegen[entity.ID]; !exists {
				w.originalRegen[entity.ID] = mana.Regen
			}

			// Calculate mana regen multiplier based on weather and intensity
			multiplier := w.calculateManaMultiplier(weatherType, intensity)

			// Apply modified regen rate
			originalRegen := w.originalRegen[entity.ID]
			newRegen := originalRegen * multiplier
			mana.Regen = newRegen

			if w.logger != nil {
				w.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"weather_type":   weatherType.String(),
					"intensity":      intensity.String(),
					"original_regen": originalRegen,
					"modified_regen": newRegen,
					"multiplier":     multiplier,
				}).Debug("Mana regen modified by weather")
			}
		} else {
			// Restore original mana regen rate
			if originalRegen, exists := w.originalRegen[entity.ID]; exists {
				mana.Regen = originalRegen
				delete(w.originalRegen, entity.ID)

				if w.logger != nil {
					w.logger.WithFields(logrus.Fields{
						"entity_id":      entity.ID,
						"restored_regen": originalRegen,
					}).Debug("Mana regen restored")
				}
			}
		}
	}
}

// calculateManaMultiplier computes mana regen multiplier based on weather and genre.
// Returns values typically in range [0.7, 1.8] where 1.0 is no change.
func (w *WeatherManaRegenSystem) calculateManaMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseMultiplier := 1.0

	// Base weather effects (genre-independent)
	switch weatherType {
	case particles.WeatherRain:
		// Rain boosts magical energy (connection to nature/water)
		baseMultiplier = 1.15
	case particles.WeatherSnow:
		// Snow is calming, slight boost to concentration
		baseMultiplier = 1.10
	case particles.WeatherFog:
		// Fog is mysterious, mild penalty (harder to focus)
		baseMultiplier = 0.90
	case particles.WeatherDust:
		// Dust is distracting, mild penalty
		baseMultiplier = 0.85
	case particles.WeatherAsh:
		// Ash is oppressive, moderate penalty
		baseMultiplier = 0.80
	case particles.WeatherNeonRain:
		// Cyberpunk neon rain - tech-magic synergy
		baseMultiplier = 1.25
	case particles.WeatherSmog:
		// Industrial smog - significant penalty
		baseMultiplier = 0.75
	case particles.WeatherRadiation:
		// Radiation - chaotic magical energy
		baseMultiplier = 1.30
	case particles.WeatherSandstorm:
		// Sandstorm - harsh conditions
		baseMultiplier = 0.70
	case particles.WeatherBloodRain:
		// Horror blood rain - dark magic boost
		baseMultiplier = 1.40
	}

	// Intensity modifier
	intensityMod := 1.0
	switch intensity {
	case particles.IntensityLight:
		intensityMod = 0.5 // Half the effect
	case particles.IntensityMedium:
		intensityMod = 1.0 // Full base effect
	case particles.IntensityHeavy:
		intensityMod = 1.3 // 30% stronger effect
	case particles.IntensityExtreme:
		intensityMod = 1.6 // 60% stronger effect
	}

	// Apply intensity to the deviation from 1.0
	deviation := baseMultiplier - 1.0
	scaledDeviation := deviation * intensityMod
	finalMultiplier := 1.0 + scaledDeviation

	// Add slight randomization for variety (±5%)
	variance := 1.0 + (w.rng.Float64()-0.5)*0.10
	finalMultiplier *= variance

	// Clamp to reasonable bounds [0.5, 2.0]
	if finalMultiplier < 0.5 {
		finalMultiplier = 0.5
	} else if finalMultiplier > 2.0 {
		finalMultiplier = 2.0
	}

	return finalMultiplier
}

// GetModifierForEntity returns the current mana regen modifier for an entity (for UI display).
// Returns 1.0 if entity has no modified regen rate.
func (w *WeatherManaRegenSystem) GetModifierForEntity(entityID uint64) float64 {
	if !w.lastWeatherActive {
		return 1.0
	}

	originalRegen, exists := w.originalRegen[entityID]
	if !exists {
		return 1.0
	}

	// Return current multiplier based on stored original
	return w.calculateManaMultiplier(w.lastWeatherType, w.lastWeatherIntensity) * originalRegen / originalRegen
}

// IsWeatherActive returns whether weather is currently affecting mana regen.
func (w *WeatherManaRegenSystem) IsWeatherActive() bool {
	return w.lastWeatherActive
}

// GetActiveWeatherType returns the current weather type affecting mana regen.
func (w *WeatherManaRegenSystem) GetActiveWeatherType() particles.WeatherType {
	return w.lastWeatherType
}
