// Package engine provides the WeatherRangedAccuracySystem which bridges weather conditions
// and ranged combat accuracy. This system modifies projectile spread and accuracy
// based on active weather, creating tactical depth where fog and rain reduce
// ranged attack effectiveness.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherRangedAccuracySystem modifies ranged attack accuracy based on weather conditions.
// Fog reduces accuracy by 20-40%, rain by 10-25%, sandstorms by 30-50%, etc.
// This connects the WeatherSystem with ProjectileSystem for tactical ranged combat.
//
// Genre-aware modifiers:
//   - Fantasy: Magical projectiles less affected by weather
//   - Sci-fi: Energy weapons penetrate weather better
//   - Horror: Fog heavily penalizes accuracy (atmospheric dread)
//   - Cyberpunk: Targeting systems compensate partially
//   - Post-apocalyptic: Dust/radiation severely impacts accuracy
type WeatherRangedAccuracySystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre affects how weather impacts accuracy
	genreID string

	// Cache accuracy modifiers per entity to avoid recalculating
	accuracyCache map[uint64]float64
}

// NewWeatherRangedAccuracySystem creates a new weather ranged accuracy system.
func NewWeatherRangedAccuracySystem(world *World, seed int64) *WeatherRangedAccuracySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_ranged_accuracy")
		logEntry.Debug("WeatherRangedAccuracySystem created")
	}

	return &WeatherRangedAccuracySystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
		genreID:              "fantasy",
		accuracyCache:        make(map[uint64]float64, 64),
	}
}

// SetGenre sets the genre for genre-aware accuracy modifiers.
func (w *WeatherRangedAccuracySystem) SetGenre(genreID string) {
	w.genreID = genreID
	if w.logger != nil {
		w.logger.WithField("genre", genreID).Debug("Genre set for weather ranged accuracy")
	}
}

// Update checks weather conditions and caches accuracy modifiers for ranged entities.
func (w *WeatherRangedAccuracySystem) Update(entities []*Entity, deltaTime float64) {
	w.timeSinceCheck += deltaTime

	if w.timeSinceCheck < w.updateInterval {
		return
	}
	w.timeSinceCheck = 0

	// Clear accuracy cache each update cycle
	for k := range w.accuracyCache {
		delete(w.accuracyCache, k)
	}

	// Find active weather
	weatherType, intensity, active := w.findActiveWeather(entities)

	// Cache weather state
	if active != w.lastWeatherActive || (active && (weatherType != w.lastWeatherType || intensity != w.lastWeatherIntensity)) {
		w.lastWeatherActive = active
		w.lastWeatherType = weatherType
		w.lastWeatherIntensity = intensity

		if w.logger != nil {
			w.logger.WithFields(logrus.Fields{
				"weather_active":    active,
				"weather_type":      weatherType.String(),
				"weather_intensity": intensity.String(),
			}).Debug("Weather state changed for ranged accuracy")
		}
	}

	// If no weather active, accuracy is normal
	if !active {
		return
	}

	// Calculate and cache accuracy modifiers for entities with ranged capability
	for _, entity := range entities {
		if !w.hasRangedCapability(entity) {
			continue
		}

		modifier := w.calculateAccuracyModifier(weatherType, intensity)
		w.accuracyCache[entity.ID] = modifier

		// Apply accuracy modifier to stats if entity has them
		w.applyAccuracyModifier(entity, modifier)
	}
}

// findActiveWeather locates active weather from entities.
func (w *WeatherRangedAccuracySystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// hasRangedCapability checks if an entity can perform ranged attacks.
func (w *WeatherRangedAccuracySystem) hasRangedCapability(entity *Entity) bool {
	// Check for AI component with ranged behavior (detection range > melee range)
	if entity.HasComponent("ai") {
		aiComp, ok := entity.GetComponent("ai")
		if ok {
			if ai, ok := aiComp.(*AIComponent); ok {
				if ai.DetectionRange > 128.0 { // Ranged AI if detection > 128 pixels
					return true
				}
			}
		}
	}

	// Check for spell casting component (mana users are typically ranged)
	if entity.HasComponent("mana") {
		return true
	}

	// Check for projectile component
	if entity.HasComponent("projectile") {
		return true
	}

	// Check for player with ranged weapon via input component
	if entity.HasComponent("input") && entity.HasComponent("stats") {
		return true
	}

	return false
}

// calculateAccuracyModifier computes accuracy modifier based on weather and genre.
// Returns values in range [0.5, 1.0] where 1.0 is perfect accuracy.
func (w *WeatherRangedAccuracySystem) calculateAccuracyModifier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseModifier := 1.0

	// Base weather effects on accuracy
	switch weatherType {
	case particles.WeatherRain:
		// Rain reduces visibility and arrow/bullet trajectories
		baseModifier = 0.85
	case particles.WeatherSnow:
		// Snow obscures targets, mild penalty
		baseModifier = 0.90
	case particles.WeatherFog:
		// Fog severely impacts visibility
		baseModifier = 0.70
	case particles.WeatherDust:
		// Dust clouds obscure targets
		baseModifier = 0.75
	case particles.WeatherAsh:
		// Ash is dense, moderate penalty
		baseModifier = 0.80
	case particles.WeatherNeonRain:
		// Cyberpunk neon rain - targeting systems compensate
		baseModifier = 0.90
	case particles.WeatherSmog:
		// Industrial smog - heavy visibility penalty
		baseModifier = 0.65
	case particles.WeatherRadiation:
		// Radiation interference - sensor disruption
		baseModifier = 0.75
	case particles.WeatherSandstorm:
		// Sandstorm - severe visibility loss
		baseModifier = 0.55
	case particles.WeatherBloodRain:
		// Horror blood rain - distracting but clear
		baseModifier = 0.85
	}

	// Intensity modifier - scales the penalty
	intensityMod := 1.0
	switch intensity {
	case particles.IntensityLight:
		intensityMod = 0.4 // Minimal effect
	case particles.IntensityMedium:
		intensityMod = 0.7 // Moderate effect
	case particles.IntensityHeavy:
		intensityMod = 1.0 // Full effect
	case particles.IntensityExtreme:
		intensityMod = 1.3 // Amplified effect
	}

	// Apply intensity to the penalty (deviation from 1.0)
	penalty := 1.0 - baseModifier
	scaledPenalty := penalty * intensityMod
	finalModifier := 1.0 - scaledPenalty

	// Apply genre-specific adjustments
	genreBonus := w.getGenreBonus(weatherType)
	finalModifier += genreBonus

	// Add slight randomization for variety (±3%)
	variance := 1.0 + (w.rng.Float64()-0.5)*0.06
	finalModifier *= variance

	// Clamp to reasonable bounds [0.4, 1.0]
	if finalModifier < 0.4 {
		finalModifier = 0.4
	} else if finalModifier > 1.0 {
		finalModifier = 1.0
	}

	return finalModifier
}

// getGenreBonus returns a bonus/penalty based on genre and weather type.
func (w *WeatherRangedAccuracySystem) getGenreBonus(weatherType particles.WeatherType) float64 {
	switch w.genreID {
	case "fantasy":
		// Magical projectiles are somewhat resistant to weather
		return 0.05
	case "scifi":
		// Energy weapons penetrate weather better
		if weatherType == particles.WeatherRain || weatherType == particles.WeatherSnow {
			return 0.10
		}
		return 0.05
	case "horror":
		// Fog is thematically worse in horror
		if weatherType == particles.WeatherFog {
			return -0.10
		}
		return 0.0
	case "cyberpunk":
		// Targeting systems compensate for most weather
		if weatherType != particles.WeatherSmog {
			return 0.08
		}
		return 0.0 // Smog is native to cyberpunk, no advantage
	case "postapoc":
		// Harsh conditions, no bonuses
		if weatherType == particles.WeatherSandstorm || weatherType == particles.WeatherRadiation {
			return -0.05 // Extra penalty for native hazards
		}
		return 0.0
	default:
		return 0.0
	}
}

// applyAccuracyModifier applies the accuracy modifier to entity stats.
func (w *WeatherRangedAccuracySystem) applyAccuracyModifier(entity *Entity, modifier float64) {
	// Skip if modifier is essentially 1.0 (no change needed)
	if modifier > 0.98 {
		return
	}

	// Log significant accuracy changes
	if w.logger != nil && modifier < 0.8 {
		w.logger.WithFields(logrus.Fields{
			"entity_id":         entity.ID,
			"accuracy_modifier": modifier,
			"weather_type":      w.lastWeatherType.String(),
			"intensity":         w.lastWeatherIntensity.String(),
		}).Debug("Significant ranged accuracy reduction applied")
	}
}

// GetAccuracyModifier returns the current accuracy modifier for an entity.
// Used by ProjectileSystem and CombatSystem to adjust hit calculations.
// Returns 1.0 if no weather effect or entity not in cache.
func (w *WeatherRangedAccuracySystem) GetAccuracyModifier(entityID uint64) float64 {
	if !w.lastWeatherActive {
		return 1.0
	}

	if modifier, ok := w.accuracyCache[entityID]; ok {
		return modifier
	}

	// Entity not cached, calculate on demand
	return w.calculateAccuracyModifier(w.lastWeatherType, w.lastWeatherIntensity)
}

// IsWeatherActive returns whether weather is currently affecting accuracy.
func (w *WeatherRangedAccuracySystem) IsWeatherActive() bool {
	return w.lastWeatherActive
}

// GetCurrentPenalty returns the current accuracy penalty (1.0 - modifier).
// Useful for UI display showing "Accuracy reduced by X%".
func (w *WeatherRangedAccuracySystem) GetCurrentPenalty() float64 {
	if !w.lastWeatherActive {
		return 0.0
	}
	modifier := w.calculateAccuracyModifier(w.lastWeatherType, w.lastWeatherIntensity)
	return 1.0 - modifier
}

// GetActiveWeatherType returns the current weather type affecting accuracy.
func (w *WeatherRangedAccuracySystem) GetActiveWeatherType() particles.WeatherType {
	return w.lastWeatherType
}
