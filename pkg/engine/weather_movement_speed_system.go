// Package engine provides the WeatherMovementSpeedSystem that connects weather conditions
// with entity movement speed. Rain makes floors slick (faster movement), snow creates
// resistance (slower), fog reduces visibility but doesn't affect speed, etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherMovementSpeedSystem modifies entity movement speed based on weather conditions.
// This connects WeatherComponent weather types with VelocityComponent for environmental
// movement synergies that affect traversal speed.
//
// Weather effects on movement:
//   - Snow/Blizzard: -15% to -30% speed (snow resistance, ice risk)
//   - Rain/Storm: +5% to +10% speed (slick surfaces for players, but impaired AI)
//   - Fog: no speed modifier (visibility only)
//   - Dust: -5% to -10% speed (reduced visibility, debris)
//   - Sandstorm: -15% to -20% speed (heavy particles slow movement)
//   - Clear: no modifier (baseline)
//
// Genre modifiers:
//   - Fantasy: rain slows slightly (magical rain heavier)
//   - Scifi: tech boots maintain traction (-50% penalty)
//   - Horror: all weather penalties doubled (survival difficulty)
//   - Cyberpunk: urban drainage keeps movement fast (+20% bonuses)
//   - Postapoc: survivors adapt to harsh conditions (-30% penalties)
type WeatherMovementSpeedSystem struct {
	world   *World
	rng     *rand.Rand
	genreID string
	logger  *logrus.Entry

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache applied speed multipliers to restore when weather changes
	appliedMultipliers map[uint64]float64

	// Cached weather state to detect changes
	lastWeatherType   particles.WeatherType
	lastIntensity     particles.WeatherIntensity
	lastWeatherActive bool
}

// NewWeatherMovementSpeedSystem creates a new weather movement speed system.
func NewWeatherMovementSpeedSystem(world *World, seed int64) *WeatherMovementSpeedSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_movement_speed",
		"seed":        seed,
	})
	logger.Debug("Creating weather movement speed system")

	return &WeatherMovementSpeedSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		logger:             logger,
		updateInterval:     0.25, // Check 4 times per second for responsive movement
		appliedMultipliers: make(map[uint64]float64),
		lastWeatherActive:  false,
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherMovementSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather movement speed")
	}
}

// Update processes weather movement speed modifiers for entities with VelocityComponent.
func (s *WeatherMovementSpeedSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := s.findActiveWeather(entities)

	// Check if weather state changed
	weatherChanged := active != s.lastWeatherActive ||
		(active && (weatherType != s.lastWeatherType || intensity != s.lastIntensity))

	if weatherChanged {
		s.lastWeatherType = weatherType
		s.lastIntensity = intensity
		s.lastWeatherActive = active
	}

	// Process entities with VelocityComponent
	for _, entity := range entities {
		s.processEntity(entity, weatherType, intensity, active, weatherChanged)
	}
}

// processEntity updates movement speed for a single entity based on weather.
func (s *WeatherMovementSpeedSystem) processEntity(entity *Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active, weatherChanged bool) {
	// Skip entities without velocity component
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	// Calculate new multiplier
	var newMultiplier float64
	if active {
		newMultiplier = s.calculateMultiplier(weatherType, intensity)
	} else {
		newMultiplier = 1.0
	}

	// Get current applied multiplier
	currentMultiplier, hasMultiplier := s.appliedMultipliers[entity.ID]

	// Skip if no change needed
	if hasMultiplier && currentMultiplier == newMultiplier && !weatherChanged {
		return
	}

	// Restore original velocity if we had a previous multiplier
	if hasMultiplier && currentMultiplier != 1.0 {
		vel.VX /= currentMultiplier
		vel.VY /= currentMultiplier
	}

	// Apply new multiplier
	if newMultiplier != 1.0 {
		vel.VX *= newMultiplier
		vel.VY *= newMultiplier
	}

	// Cache the multiplier
	s.appliedMultipliers[entity.ID] = newMultiplier

	if s.logger != nil && newMultiplier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"weather":    weatherType.String(),
			"intensity":  intensity.String(),
			"multiplier": newMultiplier,
		}).Debug("Weather movement speed modifier applied")
	}
}

// calculateMultiplier returns the speed multiplier for current weather.
// Values > 1.0 speed up movement, < 1.0 slow movement.
func (s *WeatherMovementSpeedSystem) calculateMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseMultiplier := s.getBaseMultiplier(weatherType)
	intensityFactor := s.getIntensityFactor(intensity)
	genreFactor := s.getGenreModifier(weatherType, baseMultiplier)

	// Apply intensity scaling: (base - 1.0) * intensityFactor + 1.0
	deviation := (baseMultiplier - 1.0) * intensityFactor * genreFactor
	return 1.0 + deviation
}

// getBaseMultiplier returns the base speed multiplier for a weather type.
func (s *WeatherMovementSpeedSystem) getBaseMultiplier(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherSnow:
		return 0.82 // -18% speed (snow resistance)
	case particles.WeatherRain:
		return 1.05 // +5% speed (slick surfaces help movement)
	case particles.WeatherFog:
		return 1.0 // No speed change (visibility only)
	case particles.WeatherDust:
		return 0.92 // -8% speed (debris)
	case particles.WeatherAsh:
		return 0.88 // -12% speed (thick ash)
	case particles.WeatherSmog:
		return 0.95 // -5% speed (breathing difficulty affects speed slightly)
	case particles.WeatherSandstorm:
		return 0.80 // -20% speed (heavy sand resistance)
	case particles.WeatherNeonRain:
		return 1.08 // +8% speed (cyberpunk oily rain)
	case particles.WeatherBloodRain:
		return 0.90 // -10% speed (horror dread slows movement)
	case particles.WeatherRadiation:
		return 0.85 // -15% speed (radiation weakness)
	default:
		return 1.0
	}
}

// getIntensityFactor returns how strongly intensity affects the modifier.
func (s *WeatherMovementSpeedSystem) getIntensityFactor(intensity particles.WeatherIntensity) float64 {
	switch intensity {
	case particles.IntensityLight:
		return 0.5
	case particles.IntensityMedium:
		return 1.0
	case particles.IntensityHeavy:
		return 1.5
	default:
		return 1.0
	}
}

// getGenreModifier returns genre-specific adjustment to the effect.
func (s *WeatherMovementSpeedSystem) getGenreModifier(weatherType particles.WeatherType, baseMultiplier float64) float64 {
	// Determine if this is a penalty (slowing) or bonus (speeding)
	isPenalty := baseMultiplier < 1.0

	switch s.genreID {
	case "fantasy":
		// Fantasy: magical rain is heavier, slows slightly
		if weatherType == particles.WeatherRain && baseMultiplier > 1.0 {
			return 0.6 // Reduced rain bonus
		}
		return 1.0
	case "scifi":
		if isPenalty {
			return 0.5 // Tech boots resist weather penalties
		}
		return 1.0
	case "horror":
		if isPenalty {
			return 2.0 // Survival horror: penalties are worse
		}
		return 0.5 // Bonuses are reduced
	case "cyberpunk":
		if !isPenalty {
			return 1.2 // Urban drainage enhances bonuses
		}
		return 0.8 // Urban infrastructure reduces penalties
	case "postapoc":
		if isPenalty {
			return 0.7 // Survivors adapt to harsh conditions
		}
		return 1.1 // Slight bonus from adaptation
	default:
		return 1.0
	}
}

// findActiveWeather searches for active weather in the world.
func (s *WeatherMovementSpeedSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		if !entity.HasComponent("weather") {
			continue
		}
		weatherComp, _ := entity.GetComponent("weather")
		weather, ok := weatherComp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		return weather.Config.Type, weather.Config.Intensity, true
	}
	return particles.WeatherRain, particles.IntensityLight, false
}

// GetMovementSpeedModifier returns the current movement speed modifier for an entity.
// Returns 1.0 if no modifier is applied (values < 1.0 = slower, > 1.0 = faster).
func (s *WeatherMovementSpeedSystem) GetMovementSpeedModifier(entityID uint64) float64 {
	if mult, exists := s.appliedMultipliers[entityID]; exists {
		return mult
	}
	return 1.0
}

// ClearModifier removes the movement speed modifier for an entity.
func (s *WeatherMovementSpeedSystem) ClearModifier(entityID uint64) {
	if mult, exists := s.appliedMultipliers[entityID]; exists && mult != 1.0 {
		// Try to restore original velocity
		if s.world != nil {
			entity := s.world.GetEntity(entityID)
			if entity != nil {
				if vel := entity.GetVelocity(); vel != nil {
					vel.VX /= mult
					vel.VY /= mult
				}
			}
		}
	}
	delete(s.appliedMultipliers, entityID)
}
