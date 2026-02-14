// Package engine provides the WeatherAttackSpeedSystem that connects weather conditions
// with melee attack speed (cooldown modifiers). Cold weather stiffens joints, reducing
// attack speed; storms provide covering chaos for quick strikes; etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherAttackSpeedSystem modifies melee attack cooldowns based on weather conditions.
// This connects WeatherComponent weather types with AttackComponent.Cooldown for
// environmental combat synergies that affect attack speed.
//
// Weather effects on attack speed:
//   - Snow/Blizzard: +15-30% cooldown (cold stiffens joints)
//   - Rain/Storm: -5-15% cooldown (chaos covers quick strikes)
//   - Fog: -5% cooldown (concealment enables quick attacks)
//   - Dust: +5% cooldown (dust in eyes slows reactions)
//   - Clear: no modifier (baseline)
//
// Genre modifiers:
//   - Fantasy: standard modifiers
//   - Scifi: tech compensates for cold (-50% penalty)
//   - Horror: all penalties doubled (survival horror)
//   - Cyberpunk: augments resist weather (-30% penalties)
//   - Postapoc: survivors adapt (+10% bonuses, -20% penalties)
type WeatherAttackSpeedSystem struct {
	world   *World
	rng     *rand.Rand
	genreID string
	logger  *logrus.Entry

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache applied cooldown multipliers to restore when weather changes
	appliedMultipliers map[uint64]float64

	// Cached weather state to detect changes
	lastWeatherType   particles.WeatherType
	lastIntensity     particles.WeatherIntensity
	lastWeatherActive bool
}

// NewWeatherAttackSpeedSystem creates a new weather attack speed system.
func NewWeatherAttackSpeedSystem(world *World, seed int64) *WeatherAttackSpeedSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_attack_speed",
		"seed":        seed,
	})
	logger.Debug("Creating weather attack speed system")

	return &WeatherAttackSpeedSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		logger:             logger,
		updateInterval:     0.5, // Check twice per second
		appliedMultipliers: make(map[uint64]float64),
		lastWeatherActive:  false,
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherAttackSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather attack speed")
	}
}

// Update processes weather attack speed modifiers for entities with AttackComponent.
func (s *WeatherAttackSpeedSystem) Update(entities []*Entity, deltaTime float64) {
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

	// Process entities with AttackComponent
	for _, entity := range entities {
		s.processEntity(entity, weatherType, intensity, active, weatherChanged)
	}
}

// processEntity updates attack speed for a single entity based on weather.
func (s *WeatherAttackSpeedSystem) processEntity(entity *Entity, weatherType particles.WeatherType, intensity particles.WeatherIntensity, active, weatherChanged bool) {
	// Skip entities without attack component
	if !entity.HasComponent("attack") {
		return
	}

	attackComp, _ := entity.GetComponent("attack")
	attack, ok := attackComp.(*AttackComponent)
	if !ok || attack.Cooldown <= 0 {
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

	// Restore original cooldown if we had a previous multiplier
	if hasMultiplier && currentMultiplier != 1.0 {
		attack.Cooldown = attack.Cooldown / currentMultiplier
	}

	// Apply new multiplier
	if newMultiplier != 1.0 {
		attack.Cooldown = attack.Cooldown * newMultiplier
	}

	// Cache the multiplier
	s.appliedMultipliers[entity.ID] = newMultiplier

	if s.logger != nil && newMultiplier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"weather":    weatherType.String(),
			"intensity":  intensity.String(),
			"multiplier": newMultiplier,
		}).Debug("Weather attack speed modifier applied")
	}
}

// calculateMultiplier returns the cooldown multiplier for current weather.
// Values > 1.0 slow attacks, < 1.0 speed up attacks.
func (s *WeatherAttackSpeedSystem) calculateMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseMultiplier := s.getBaseMultiplier(weatherType)
	intensityFactor := s.getIntensityFactor(intensity)
	genreFactor := s.getGenreModifier(weatherType, baseMultiplier)

	// Apply intensity scaling: (base - 1.0) * intensityFactor + 1.0
	// This scales the deviation from 1.0 by intensity
	deviation := (baseMultiplier - 1.0) * intensityFactor * genreFactor
	return 1.0 + deviation
}

// getBaseMultiplier returns the base cooldown multiplier for a weather type.
func (s *WeatherAttackSpeedSystem) getBaseMultiplier(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherSnow:
		return 1.20 // +20% cooldown (cold slows attacks)
	case particles.WeatherRain:
		return 0.92 // -8% cooldown (rain provides chaos cover for quick strikes)
	case particles.WeatherFog:
		return 0.95 // -5% cooldown (fog concealment)
	case particles.WeatherDust:
		return 1.08 // +8% cooldown (dust in eyes)
	case particles.WeatherAsh:
		return 1.10 // +10% cooldown (ash irritation)
	case particles.WeatherSmog:
		return 1.05 // +5% cooldown (breathing difficulty)
	case particles.WeatherSandstorm:
		return 1.15 // +15% cooldown (visibility and sand)
	case particles.WeatherNeonRain:
		return 0.95 // -5% cooldown (cyberpunk chaos)
	case particles.WeatherBloodRain:
		return 1.12 // +12% cooldown (horror dread slows actions)
	case particles.WeatherRadiation:
		return 1.08 // +8% cooldown (radiation weakness)
	default:
		return 1.0 // No modifier for unknown types
	}
}

// getIntensityFactor returns how strongly intensity affects the modifier.
func (s *WeatherAttackSpeedSystem) getIntensityFactor(intensity particles.WeatherIntensity) float64 {
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
func (s *WeatherAttackSpeedSystem) getGenreModifier(weatherType particles.WeatherType, baseMultiplier float64) float64 {
	// Determine if this is a penalty (slowing) or bonus (speeding)
	isPenalty := baseMultiplier > 1.0

	switch s.genreID {
	case "fantasy":
		return 1.0 // Standard modifiers
	case "scifi":
		if isPenalty {
			return 0.5 // Tech compensates for weather penalties
		}
		return 1.0
	case "horror":
		if isPenalty {
			return 2.0 // Survival horror: penalties are worse
		}
		return 0.5 // Bonuses are reduced
	case "cyberpunk":
		if isPenalty {
			return 0.7 // Augments resist weather
		}
		return 1.2 // Augments enhance bonuses
	case "postapoc":
		if isPenalty {
			return 0.8 // Survivors adapt to harsh conditions
		}
		return 1.1 // Slight bonus from adaptation
	default:
		return 1.0
	}
}

// findActiveWeather searches for active weather in the world.
func (s *WeatherAttackSpeedSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// GetAttackSpeedModifier returns the current attack speed modifier for an entity.
// Returns 1.0 if no modifier is applied (values < 1.0 = faster, > 1.0 = slower).
func (s *WeatherAttackSpeedSystem) GetAttackSpeedModifier(entityID uint64) float64 {
	if mult, exists := s.appliedMultipliers[entityID]; exists {
		return mult
	}
	return 1.0
}

// ClearModifier removes the attack speed modifier for an entity.
func (s *WeatherAttackSpeedSystem) ClearModifier(entityID uint64) {
	if mult, exists := s.appliedMultipliers[entityID]; exists && mult != 1.0 {
		// Try to restore original cooldown
		if s.world != nil {
			if entity, found := s.world.GetEntity(entityID); found && entity != nil {
				if attackComp, ok := entity.GetComponent("attack"); ok {
					if attack, ok := attackComp.(*AttackComponent); ok {
						attack.Cooldown = attack.Cooldown / mult
					}
				}
			}
		}
	}
	delete(s.appliedMultipliers, entityID)
}
