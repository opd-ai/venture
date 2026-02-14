// Package engine provides the WeatherCritChanceSystem that connects weather conditions
// with critical hit chance for environmental combat synergies.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherCritChanceSystem modifies critical hit chance based on weather conditions.
// Creates environmental synergies: fog enables surprise attacks (bonus crit), storms
// have chaotic energy (bonus crit), rain/snow make precision harder (reduced crit),
// and genre-specific effects (horror blood rain favors violence, cyberpunk neon rain
// disrupts targeting systems).
//
// This system bridges WeatherComponent weather types with StatsComponent.CritChance.
type WeatherCritChanceSystem struct {
	world   *World
	rng     *rand.Rand
	genreID string
	logger  *logrus.Entry

	// updateInterval controls how often we recalculate bonuses (seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original crit values per entity
	originalCrit   map[uint64]float64
	appliedBonuses map[uint64]float64

	// Last weather for cache invalidation
	lastWeather particles.WeatherType
}

// NewWeatherCritChanceSystem creates a new weather-critical chance modifier system.
func NewWeatherCritChanceSystem(world *World, seed int64) *WeatherCritChanceSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_crit_chance",
		"seed":        seed,
	})
	logger.Debug("Creating weather crit chance system")

	return &WeatherCritChanceSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		logger:         logger,
		updateInterval: 0.5,
		originalCrit:   make(map[uint64]float64),
		appliedBonuses: make(map[uint64]float64),
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherCritChanceSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.clearAllBonuses()
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather crit chance")
	}
}

// Update processes weather-crit chance modifiers for all relevant entities.
func (s *WeatherCritChanceSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	weather := s.getCurrentWeather()
	if weather == nil {
		s.clearAllBonuses()
		return
	}

	// Check if weather changed - clear cache
	if weather.Config.Type != s.lastWeather {
		s.clearAllBonuses()
		s.lastWeather = weather.Config.Type
	}

	modifier := s.calculateModifier(weather.Config.Type)
	if modifier == 0.0 {
		return // No change needed
	}

	for _, entity := range entities {
		s.processEntity(entity, modifier)
	}
}

// processEntity applies the crit chance modifier to a single entity.
func (s *WeatherCritChanceSystem) processEntity(entity *Entity, modifier float64) {
	if !entity.HasComponent("stats") {
		return
	}

	statsComp, _ := entity.GetComponent("stats")
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}

	// Store original crit if not already stored
	if _, exists := s.originalCrit[entity.ID]; !exists {
		s.originalCrit[entity.ID] = stats.CritChance
	}

	// Check if bonus already applied at this level
	currentBonus, hasBonus := s.appliedBonuses[entity.ID]
	if hasBonus && currentBonus == modifier {
		return
	}

	// Apply new modifier
	originalCrit := s.originalCrit[entity.ID]
	newCrit := originalCrit + modifier
	// Clamp between 0.0 and 0.75 (75% max crit chance)
	if newCrit < 0.0 {
		newCrit = 0.0
	}
	if newCrit > 0.75 {
		newCrit = 0.75
	}

	stats.CritChance = newCrit
	s.appliedBonuses[entity.ID] = modifier

	if s.logger != nil && modifier != 0.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"original_crit": originalCrit,
			"modifier":      modifier,
			"new_crit":      newCrit,
		}).Debug("Weather crit chance modifier applied")
	}
}

// getCurrentWeather finds the active weather component in the world.
func (s *WeatherCritChanceSystem) getCurrentWeather() *WeatherComponent {
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

// calculateModifier computes the crit chance modifier for a weather type.
// Returns additive modifier (e.g., 0.05 = +5% crit chance).
func (s *WeatherCritChanceSystem) calculateModifier(weather particles.WeatherType) float64 {
	baseModifier := s.getBaseModifier(weather)
	genreBonus := s.getGenreBonus(weather)
	return baseModifier + genreBonus
}

// getBaseModifier returns the base crit chance modifier for weather types.
func (s *WeatherCritChanceSystem) getBaseModifier(weather particles.WeatherType) float64 {
	switch weather {
	case particles.WeatherRain:
		return -0.03 // -3% - wet conditions reduce precision
	case particles.WeatherSnow:
		return -0.02 // -2% - cold stiffens joints
	case particles.WeatherFog:
		return 0.08 // +8% - concealment enables surprise attacks
	case particles.WeatherDust, particles.WeatherSandstorm:
		return 0.05 // +5% - obscured vision enables strikes
	case particles.WeatherAsh:
		return 0.03 // +3% - distraction enables opportunistic hits
	case particles.WeatherNeonRain:
		return -0.02 // -2% - distracting lights reduce focus
	case particles.WeatherSmog:
		return 0.05 // +5% - concealment in urban haze
	case particles.WeatherRadiation:
		return 0.06 // +6% - weakened targets easier to crit
	case particles.WeatherBloodRain:
		return 0.10 // +10% - violence empowered by horror
	default:
		return 0.0
	}
}

// getGenreBonus returns additional bonus based on genre-weather combinations.
func (s *WeatherCritChanceSystem) getGenreBonus(weather particles.WeatherType) float64 {
	switch s.genreID {
	case "horror":
		if weather == particles.WeatherBloodRain {
			return 0.05 // Extra +5% in horror blood rain
		}
		if weather == particles.WeatherFog {
			return 0.03 // Extra +3% fog bonus in horror
		}
	case "cyberpunk":
		if weather == particles.WeatherSmog {
			return 0.03 // Extra +3% smog bonus in cyberpunk
		}
	case "postapoc":
		if weather == particles.WeatherRadiation {
			return 0.04 // Extra +4% radiation bonus in post-apoc
		}
		if weather == particles.WeatherSandstorm {
			return 0.02 // Extra +2% sandstorm bonus
		}
	case "scifi":
		if weather == particles.WeatherDust {
			return 0.02 // Space dust enables precision in scifi
		}
	case "fantasy":
		if weather == particles.WeatherFog {
			return 0.02 // Mystical fog bonus in fantasy
		}
	}
	return 0.0
}

// clearAllBonuses removes all applied bonuses and restores original crit values.
func (s *WeatherCritChanceSystem) clearAllBonuses() {
	if s.world == nil {
		return
	}

	for entityID, originalCrit := range s.originalCrit {
		entity, exists := s.world.GetEntity(entityID)
		if !exists || entity == nil {
			continue
		}

		statsComp, ok := entity.GetComponent("stats")
		if !ok {
			continue
		}
		stats, ok := statsComp.(*StatsComponent)
		if !ok {
			continue
		}

		stats.CritChance = originalCrit
	}

	s.originalCrit = make(map[uint64]float64)
	s.appliedBonuses = make(map[uint64]float64)
}

// GetCurrentModifier returns the current crit chance modifier for UI display.
func (s *WeatherCritChanceSystem) GetCurrentModifier() float64 {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 0.0
	}
	return s.calculateModifier(weather.Config.Type)
}
