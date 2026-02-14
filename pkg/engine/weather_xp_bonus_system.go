// Package engine provides the WeatherXPBonusSystem which bridges weather conditions
// and experience point gains. This system grants bonus XP when players earn experience
// during specific weather conditions, allowing for genre-aware progression where
// storms boost XP for lightning users, fog aids rogues in stealth kills, etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherXPBonusSystem modifies XP gains based on weather conditions.
// Rain boosts XP by 5-15%, thunderstorms by 15-30%, fog grants stealth kill bonus, etc.
// This connects the WeatherSystem with the ProgressionSystem to create strategic gameplay.
//
// Integration Points:
// - Reads from: WeatherComponent (active weather state)
// - Modifies: XP via ProgressionSystem callback integration
// - Genre-aware: Different weather types favor different playstyles per genre
type WeatherXPBonusSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check weather (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track active XP multiplier per entity
	activeMultipliers map[uint64]float64

	// Current weather state cache
	lastWeatherType      particles.WeatherType
	lastWeatherIntensity particles.WeatherIntensity
	lastWeatherActive    bool

	// Genre affects which weather types boost XP
	genreID string

	// progressionSystem reference for callback integration
	progressionSystem *ProgressionSystem
}

// NewWeatherXPBonusSystem creates a new weather XP bonus system.
func NewWeatherXPBonusSystem(world *World, seed int64) *WeatherXPBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_xp_bonus")
		logEntry.Debug("WeatherXPBonusSystem created")
	}

	return &WeatherXPBonusSystem{
		world:                world,
		logger:               logEntry,
		rng:                  rand.New(rand.NewSource(seed)),
		updateInterval:       0.5, // Check twice per second
		activeMultipliers:    make(map[uint64]float64),
		lastWeatherType:      particles.WeatherRain,
		lastWeatherIntensity: particles.IntensityLight,
		lastWeatherActive:    false,
		genreID:              "fantasy",
	}
}

// SetGenre sets the genre for genre-aware XP modifiers.
func (w *WeatherXPBonusSystem) SetGenre(genreID string) {
	w.genreID = genreID
	if w.logger != nil {
		w.logger.WithField("genre", genreID).Debug("Genre set for weather XP bonus")
	}
}

// SetProgressionSystem sets the progression system for XP callback integration.
func (w *WeatherXPBonusSystem) SetProgressionSystem(ps *ProgressionSystem) {
	w.progressionSystem = ps
}

// Update checks weather conditions and updates active XP multipliers.
func (w *WeatherXPBonusSystem) Update(entities []*Entity, deltaTime float64) {
	w.timeSinceCheck += deltaTime

	if w.timeSinceCheck < w.updateInterval {
		return
	}
	w.timeSinceCheck = 0

	// Find active weather
	weatherType, intensity, active := w.findActiveWeather(entities)

	// Update cached weather state
	w.lastWeatherActive = active
	w.lastWeatherType = weatherType
	w.lastWeatherIntensity = intensity

	// Update multipliers for all player entities
	for _, entity := range entities {
		// Only apply to player entities (have input component)
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		if active {
			multiplier := w.calculateXPMultiplier(weatherType, intensity)
			w.activeMultipliers[entity.ID] = multiplier
		} else {
			delete(w.activeMultipliers, entity.ID)
		}
	}
}

// findActiveWeather locates active weather from entities.
func (w *WeatherXPBonusSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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

// calculateXPMultiplier computes XP multiplier based on weather, intensity, and genre.
// Returns values typically in range [1.0, 1.5] where 1.0 is no bonus.
func (w *WeatherXPBonusSystem) calculateXPMultiplier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) float64 {
	baseMultiplier := 1.0

	// Base weather effects vary by genre
	switch w.genreID {
	case "fantasy":
		baseMultiplier = w.fantasyWeatherBonus(weatherType)
	case "scifi":
		baseMultiplier = w.scifiWeatherBonus(weatherType)
	case "horror":
		baseMultiplier = w.horrorWeatherBonus(weatherType)
	case "cyberpunk":
		baseMultiplier = w.cyberpunkWeatherBonus(weatherType)
	case "postapoc":
		baseMultiplier = w.postapocWeatherBonus(weatherType)
	default:
		baseMultiplier = w.fantasyWeatherBonus(weatherType)
	}

	// Intensity modifier scales the bonus
	intensityMod := 1.0
	switch intensity {
	case particles.IntensityLight:
		intensityMod = 0.5 // Half the bonus
	case particles.IntensityMedium:
		intensityMod = 1.0 // Full base bonus
	case particles.IntensityHeavy:
		intensityMod = 1.25 // 25% stronger bonus
	case particles.IntensityExtreme:
		intensityMod = 1.5 // 50% stronger bonus
	}

	// Apply intensity to the bonus portion only
	bonus := baseMultiplier - 1.0
	scaledBonus := bonus * intensityMod
	finalMultiplier := 1.0 + scaledBonus

	// Add slight randomization for variety (±3%)
	variance := 1.0 + (w.rng.Float64()-0.5)*0.06
	finalMultiplier *= variance

	// Clamp to reasonable bounds [1.0, 1.5]
	if finalMultiplier < 1.0 {
		finalMultiplier = 1.0
	} else if finalMultiplier > 1.5 {
		finalMultiplier = 1.5
	}

	return finalMultiplier
}

// fantasyWeatherBonus returns XP bonus for fantasy genre weather.
func (w *WeatherXPBonusSystem) fantasyWeatherBonus(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherRain:
		return 1.08 // Nature magic synergy
	case particles.WeatherSnow:
		return 1.05 // Cold focus boost
	case particles.WeatherFog:
		return 1.12 // Stealth and ambush bonus
	case particles.WeatherDust:
		return 1.03 // Slight penalty conditions
	case particles.WeatherAsh:
		return 1.10 // Dark magic synergy
	default:
		return 1.0
	}
}

// scifiWeatherBonus returns XP bonus for sci-fi genre weather.
func (w *WeatherXPBonusSystem) scifiWeatherBonus(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherRadiation:
		return 1.15 // Mutant/enhanced bonus
	case particles.WeatherSmog:
		return 1.06 // Industrial combat
	case particles.WeatherDust:
		return 1.08 // Terrain adaptation
	case particles.WeatherAsh:
		return 1.10 // Post-conflict zones
	default:
		return 1.0
	}
}

// horrorWeatherBonus returns XP bonus for horror genre weather.
func (w *WeatherXPBonusSystem) horrorWeatherBonus(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherFog:
		return 1.18 // Terror and survival
	case particles.WeatherBloodRain:
		return 1.25 // Dark rituals bonus
	case particles.WeatherAsh:
		return 1.15 // Doom atmosphere
	case particles.WeatherRain:
		return 1.10 // Isolation bonus
	default:
		return 1.0
	}
}

// cyberpunkWeatherBonus returns XP bonus for cyberpunk genre weather.
func (w *WeatherXPBonusSystem) cyberpunkWeatherBonus(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherNeonRain:
		return 1.20 // Neon district synergy
	case particles.WeatherSmog:
		return 1.12 // Corporate zone ops
	case particles.WeatherRain:
		return 1.08 // Street runner bonus
	case particles.WeatherFog:
		return 1.10 // Stealth hacking bonus
	default:
		return 1.0
	}
}

// postapocWeatherBonus returns XP bonus for post-apocalyptic genre weather.
func (w *WeatherXPBonusSystem) postapocWeatherBonus(weatherType particles.WeatherType) float64 {
	switch weatherType {
	case particles.WeatherRadiation:
		return 1.22 // Wasteland survival
	case particles.WeatherSandstorm:
		return 1.18 // Desert combat mastery
	case particles.WeatherDust:
		return 1.15 // Scavenger bonus
	case particles.WeatherAsh:
		return 1.12 // Nuclear winter survival
	default:
		return 1.0
	}
}

// GetXPMultiplier returns the current XP multiplier for an entity.
// Returns 1.0 if no weather bonus is active.
func (w *WeatherXPBonusSystem) GetXPMultiplier(entityID uint64) float64 {
	if multiplier, exists := w.activeMultipliers[entityID]; exists {
		return multiplier
	}
	return 1.0
}

// ApplyBonusToXP calculates the bonus XP to add based on weather.
// Returns the total XP (base + bonus) that should be awarded.
func (w *WeatherXPBonusSystem) ApplyBonusToXP(entityID uint64, baseXP int) int {
	multiplier := w.GetXPMultiplier(entityID)
	if multiplier <= 1.0 {
		return baseXP
	}

	bonusXP := int(float64(baseXP) * (multiplier - 1.0))
	totalXP := baseXP + bonusXP

	if w.logger != nil && bonusXP > 0 {
		w.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"base_xp":      baseXP,
			"bonus_xp":     bonusXP,
			"total_xp":     totalXP,
			"multiplier":   multiplier,
			"weather_type": w.lastWeatherType.String(),
			"genre":        w.genreID,
		}).Debug("Weather XP bonus applied")
	}

	return totalXP
}

// IsWeatherActive returns whether weather is currently affecting XP gains.
func (w *WeatherXPBonusSystem) IsWeatherActive() bool {
	return w.lastWeatherActive
}

// GetActiveWeatherType returns the current weather type affecting XP.
func (w *WeatherXPBonusSystem) GetActiveWeatherType() particles.WeatherType {
	return w.lastWeatherType
}

// GetActiveWeatherIntensity returns the current weather intensity.
func (w *WeatherXPBonusSystem) GetActiveWeatherIntensity() particles.WeatherIntensity {
	return w.lastWeatherIntensity
}
