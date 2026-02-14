// Package engine provides the WeatherCooldownSystem which bridges weather conditions
// and spell cooldown reduction. This system modifies how fast spell cooldowns tick
// based on active weather, allowing for genre-aware magical gameplay where thunderstorms
// speed up lightning spell recovery, fog slows cooldowns, etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherCooldownSystem modifies spell cooldown rates based on weather conditions.
// Thunderstorms reduce cooldowns by 15-40%, fog increases by 10-20%, etc.
// This connects the WeatherSystem with the SpellCastingSystem for magical gameplay.
type WeatherCooldownSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache the original cooldown multipliers to restore when weather clears
	originalMultipliers map[uint64]float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre affects which weather types boost which spell schools
	genreID string
}

// NewWeatherCooldownSystem creates a new weather cooldown system.
func NewWeatherCooldownSystem(world *World, seed int64) *WeatherCooldownSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_cooldown")
		logEntry.Debug("WeatherCooldownSystem created")
	}

	return &WeatherCooldownSystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		originalMultipliers:  make(map[uint64]float64),
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
		genreID:              "fantasy",
	}
}

// SetGenre sets the genre for genre-aware cooldown modifiers.
func (w *WeatherCooldownSystem) SetGenre(genreID string) {
	w.genreID = genreID
	if w.logger != nil {
		w.logger.WithField("genre", genreID).Debug("Genre set for weather cooldown")
	}
}

// Update checks weather conditions and modifies spell cooldown rates.
func (w *WeatherCooldownSystem) Update(entities []*Entity, deltaTime float64) {
	w.timeSinceCheck += deltaTime

	if w.timeSinceCheck < w.updateInterval {
		return
	}
	w.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := w.findActiveWeather(entities)

	// If weather state changed, update cooldown rates
	if active != w.lastWeatherActive || (active && (weatherType != w.lastWeatherType || intensity != w.lastWeatherIntensity)) {
		w.updateCooldownRates(entities, weatherType, intensity, active)
		w.lastWeatherActive = active
		w.lastWeatherType = weatherType
		w.lastWeatherIntensity = intensity
	}

	// Apply cooldown speed modification each update cycle when weather is active
	if w.lastWeatherActive {
		w.applyCooldownSpeedModifier(entities)
	}
}

// findActiveWeather locates active weather from entities.
func (w *WeatherCooldownSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// updateCooldownRates tracks entities that should receive cooldown modifiers.
func (w *WeatherCooldownSystem) updateCooldownRates(entities []*Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active bool) {
	for _, entity := range entities {
		if !entity.HasComponent("spell_slots") {
			continue
		}

		if active {
			// Store original multiplier (1.0) if not already stored
			if _, exists := w.originalMultipliers[entity.ID]; !exists {
				w.originalMultipliers[entity.ID] = 1.0
			}

			multiplier := w.calculateCooldownMultiplier(weatherType, intensity)

			if w.logger != nil {
				w.logger.WithFields(logrus.Fields{
					"entity_id":    entity.ID,
					"weather_type": weatherType.String(),
					"intensity":    intensity.String(),
					"multiplier":   multiplier,
				}).Debug("Cooldown rate modified by weather")
			}
		} else {
			// Remove tracking when weather clears
			if _, exists := w.originalMultipliers[entity.ID]; exists {
				delete(w.originalMultipliers, entity.ID)

				if w.logger != nil {
					w.logger.WithFields(logrus.Fields{
						"entity_id": entity.ID,
					}).Debug("Cooldown rate restored")
				}
			}
		}
	}
}

// applyCooldownSpeedModifier applies bonus cooldown reduction to spell slots.
func (w *WeatherCooldownSystem) applyCooldownSpeedModifier(entities []*Entity) {
	multiplier := w.calculateCooldownMultiplier(w.lastWeatherType, w.lastWeatherIntensity)
	bonusReduction := multiplier - 1.0 // How much extra reduction to apply

	if bonusReduction == 0 {
		return // No effect
	}

	for _, entity := range entities {
		if _, tracked := w.originalMultipliers[entity.ID]; !tracked {
			continue
		}

		spellComp, ok := entity.GetComponent("spell_slots")
		if !ok {
			continue
		}

		slots, ok := spellComp.(*SpellSlotComponent)
		if !ok {
			continue
		}

		// Apply bonus cooldown tick (positive bonusReduction = faster cooldown)
		for i := range slots.Cooldowns {
			if slots.Cooldowns[i] > 0 {
				// Apply bonus reduction (multiplied by update interval for timing)
				bonusTick := slots.Cooldowns[i] * bonusReduction * w.updateInterval * 0.1
				slots.Cooldowns[i] -= bonusTick
				if slots.Cooldowns[i] < 0 {
					slots.Cooldowns[i] = 0
				}
			}
		}
	}
}

// calculateCooldownMultiplier computes cooldown speed multiplier based on weather and genre.
// Returns values where >1.0 means faster cooldowns, <1.0 means slower cooldowns.
func (w *WeatherCooldownSystem) calculateCooldownMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseMultiplier := 1.0

	// Base weather effects (genre-independent)
	switch weatherType {
	case particles.WeatherRain:
		// Rain calms the mind, slight cooldown boost
		baseMultiplier = 1.10
	case particles.WeatherSnow:
		// Cold slows magical recovery slightly
		baseMultiplier = 0.95
	case particles.WeatherFog:
		// Fog disrupts focus, slower cooldowns
		baseMultiplier = 0.85
	case particles.WeatherDust:
		// Dust is distracting
		baseMultiplier = 0.90
	case particles.WeatherAsh:
		// Ash carries dark energy, boosts dark magic cooldowns
		baseMultiplier = 1.05
	case particles.WeatherNeonRain:
		// Cyberpunk neon rain - tech-magic synergy boost
		baseMultiplier = 1.20
	case particles.WeatherSmog:
		// Industrial smog interferes with magic
		baseMultiplier = 0.80
	case particles.WeatherRadiation:
		// Radiation causes chaotic magical fluctuations
		baseMultiplier = 1.25
	case particles.WeatherSandstorm:
		// Harsh conditions slow recovery
		baseMultiplier = 0.75
	case particles.WeatherBloodRain:
		// Horror blood rain - dark magic surge
		baseMultiplier = 1.35
	}

	// Genre-specific modifiers
	switch w.genreID {
	case "fantasy":
		// Fantasy emphasizes natural weather effects
		if weatherType == particles.WeatherRain {
			baseMultiplier += 0.05 // Extra boost in fantasy
		}
	case "scifi":
		// Sci-fi tech synergy with neon rain
		if weatherType == particles.WeatherNeonRain || weatherType == particles.WeatherRadiation {
			baseMultiplier += 0.10
		}
	case "horror":
		// Horror benefits from dark weather
		if weatherType == particles.WeatherFog || weatherType == particles.WeatherBloodRain || weatherType == particles.WeatherAsh {
			baseMultiplier += 0.15
		}
	case "cyberpunk":
		// Cyberpunk urban weather synergy
		if weatherType == particles.WeatherNeonRain || weatherType == particles.WeatherSmog {
			baseMultiplier += 0.10
		}
	case "postapoc":
		// Post-apocalyptic harsh weather adaptation
		if weatherType == particles.WeatherRadiation || weatherType == particles.WeatherDust || weatherType == particles.WeatherSandstorm {
			baseMultiplier += 0.05 // Adapted to harsh conditions
		}
	}

	// Intensity modifier
	intensityMod := 1.0
	switch intensity {
	case particles.IntensityLight:
		intensityMod = 0.5 // Half the effect
	case particles.IntensityMedium:
		intensityMod = 1.0 // Full base effect
	case particles.IntensityHeavy:
		intensityMod = 1.25 // 25% stronger effect
	case particles.IntensityExtreme:
		intensityMod = 1.5 // 50% stronger effect
	}

	// Apply intensity to the deviation from 1.0
	deviation := baseMultiplier - 1.0
	scaledDeviation := deviation * intensityMod
	finalMultiplier := 1.0 + scaledDeviation

	// Add slight randomization for variety (±3%)
	variance := 1.0 + (w.rng.Float64()-0.5)*0.06
	finalMultiplier *= variance

	// Clamp to reasonable bounds [0.6, 1.6]
	if finalMultiplier < 0.6 {
		finalMultiplier = 0.6
	} else if finalMultiplier > 1.6 {
		finalMultiplier = 1.6
	}

	return finalMultiplier
}

// GetModifierForEntity returns the current cooldown modifier for an entity (for UI display).
// Returns 1.0 if entity has no modified cooldown rate.
func (w *WeatherCooldownSystem) GetModifierForEntity(entityID uint64) float64 {
	if !w.lastWeatherActive {
		return 1.0
	}

	if _, exists := w.originalMultipliers[entityID]; !exists {
		return 1.0
	}

	return w.calculateCooldownMultiplier(w.lastWeatherType, w.lastWeatherIntensity)
}

// IsWeatherActive returns whether weather is currently affecting cooldowns.
func (w *WeatherCooldownSystem) IsWeatherActive() bool {
	return w.lastWeatherActive
}

// GetActiveWeatherType returns the current weather type affecting cooldowns.
func (w *WeatherCooldownSystem) GetActiveWeatherType() particles.WeatherType {
	return w.lastWeatherType
}
