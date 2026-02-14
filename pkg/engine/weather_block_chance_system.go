// Package engine provides the WeatherBlockChanceSystem that connects weather conditions
// with block chance for environmental combat synergies.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherBlockChanceSystem modifies block chance based on weather conditions.
// Creates environmental synergies: rain makes shields slippery (reduced block),
// snow stiffens arms and slows reaction (reduced block), fog obscures attacks
// (harder to time blocks), dust in eyes reduces awareness (reduced block).
//
// This system bridges WeatherComponent weather types with StatsComponent.BlockChance.
//
// Weather effects on block chance:
//   - Rain: -5 to -15% block chance (wet grip on shield)
//   - Snow/Blizzard: -10 to -20% block chance (cold slows reflexes)
//   - Fog: -8 to -12% block chance (can't see attacks coming)
//   - Dust: -5 to -10% block chance (debris obscures vision)
//   - Storm: -15 to -25% block chance (violent conditions)
//   - Clear: no modifier
//
// Genre modifiers:
//   - Fantasy: standard modifiers (magic doesn't help blocking)
//   - Scifi: -50% penalties (advanced shields compensate)
//   - Horror: penalties doubled (panic affects defense)
//   - Cyberpunk: -30% penalties (augmented reflexes help)
//   - Postapoc: +20% penalties (worn equipment struggles)
type WeatherBlockChanceSystem struct {
	world   *World
	rng     *rand.Rand
	genreID string
	logger  *logrus.Entry

	// updateInterval controls how often we recalculate bonuses (seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original block values per entity
	originalBlock  map[uint64]float64
	appliedBonuses map[uint64]float64

	// Last weather for cache invalidation
	lastWeather particles.WeatherType
}

// NewWeatherBlockChanceSystem creates a new weather-block chance modifier system.
func NewWeatherBlockChanceSystem(world *World, seed int64) *WeatherBlockChanceSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_block_chance",
		"seed":        seed,
	})
	logger.Debug("Creating weather block chance system")

	return &WeatherBlockChanceSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		logger:         logger,
		updateInterval: 0.5,
		originalBlock:  make(map[uint64]float64),
		appliedBonuses: make(map[uint64]float64),
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherBlockChanceSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.clearAllBonuses()
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather block chance")
	}
}

// Update processes weather-block chance modifiers for all relevant entities.
func (s *WeatherBlockChanceSystem) Update(entities []*Entity, deltaTime float64) {
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

// processEntity applies the block chance modifier to a single entity.
func (s *WeatherBlockChanceSystem) processEntity(entity *Entity, modifier float64) {
	if !entity.HasComponent("stats") {
		return
	}

	statsComp, _ := entity.GetComponent("stats")
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}

	// Store original block if not already stored
	if _, exists := s.originalBlock[entity.ID]; !exists {
		s.originalBlock[entity.ID] = stats.BlockChance
	}

	// Check if bonus already applied at this level
	currentBonus, hasBonus := s.appliedBonuses[entity.ID]
	if hasBonus && currentBonus == modifier {
		return
	}

	// Apply new modifier
	originalBlock := s.originalBlock[entity.ID]
	newBlock := originalBlock + modifier
	// Clamp between 0.0 and 0.50 (50% max block chance)
	if newBlock < 0.0 {
		newBlock = 0.0
	}
	if newBlock > 0.50 {
		newBlock = 0.50
	}

	stats.BlockChance = newBlock
	s.appliedBonuses[entity.ID] = modifier

	if s.logger != nil && modifier != 0.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"original_block": originalBlock,
			"modifier":       modifier,
			"new_block":      newBlock,
		}).Debug("Weather block chance modifier applied")
	}
}

// getCurrentWeather finds the active weather component in the world.
func (s *WeatherBlockChanceSystem) getCurrentWeather() *WeatherComponent {
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

// calculateModifier computes the block chance modifier for current weather.
func (s *WeatherBlockChanceSystem) calculateModifier(weather particles.WeatherType) float64 {
	baseModifier := s.getBaseModifier(weather)
	genreMultiplier := s.getGenreMultiplier()
	return baseModifier * genreMultiplier
}

// getBaseModifier returns the base block chance modifier for each weather type.
func (s *WeatherBlockChanceSystem) getBaseModifier(weather particles.WeatherType) float64 {
	switch weather {
	case particles.WeatherRain:
		// Rain makes shields slippery: -10% block
		return -0.10
	case particles.WeatherSnow:
		// Snow slows reflexes: -12% block
		return -0.12
	case particles.WeatherBlizzard:
		// Blizzard severely hampers blocking: -20% block
		return -0.20
	case particles.WeatherFog:
		// Fog obscures incoming attacks: -8% block
		return -0.08
	case particles.WeatherDust:
		// Dust in eyes: -6% block
		return -0.06
	case particles.WeatherStorm:
		// Storm violence: -18% block
		return -0.18
	case particles.WeatherClear:
		// Clear weather: no modifier
		return 0.0
	default:
		return 0.0
	}
}

// getGenreMultiplier returns the genre-specific modifier multiplier.
func (s *WeatherBlockChanceSystem) getGenreMultiplier() float64 {
	switch s.genreID {
	case "fantasy":
		// Standard fantasy: full penalties
		return 1.0
	case "scifi":
		// Advanced shields compensate: half penalties
		return 0.5
	case "horror":
		// Panic affects defense: double penalties
		return 2.0
	case "cyberpunk":
		// Augmented reflexes help: 70% penalties
		return 0.7
	case "postapoc":
		// Worn equipment struggles: 120% penalties
		return 1.2
	default:
		return 1.0
	}
}

// clearAllBonuses removes all applied block bonuses and restores original values.
func (s *WeatherBlockChanceSystem) clearAllBonuses() {
	if s.world == nil {
		return
	}

	for entityID, originalBlock := range s.originalBlock {
		entity := s.world.GetEntity(entityID)
		if entity == nil {
			continue
		}

		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats, ok := statsComp.(*StatsComponent); ok {
				stats.BlockChance = originalBlock
			}
		}
	}

	s.originalBlock = make(map[uint64]float64)
	s.appliedBonuses = make(map[uint64]float64)
}

// GetCurrentModifier returns the current block chance modifier for debugging/UI.
func (s *WeatherBlockChanceSystem) GetCurrentModifier() float64 {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 0.0
	}
	return s.calculateModifier(weather.Config.Type)
}
