// Package engine provides the WeatherCompanionBonusParticleSystem for visual feedback
// when companions receive weather-based stat bonuses. This system connects
// WeatherCompanionBonusSystem with ParticleSystem to spawn genre-aware particle effects
// when companion stats change due to weather conditions.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherCompanionBonusParticleSystem spawns particles when companions gain weather bonuses.
// It monitors WeatherCompanionBonusComponent changes and provides visual feedback with
// genre-aware particle effects for attack boosts, defense changes, and speed modifiers.
type WeatherCompanionBonusParticleSystem struct {
	world                       *World
	particleSystem              *ParticleSystem
	weatherCompanionBonusSystem *WeatherCompanionBonusSystem
	genreID                     string
	seed                        int64
	rng                         *rand.Rand
	logger                      *logrus.Entry

	// Configuration
	pulseInterval float64 // Time between bonus indicator particles
	timeSinceEmit float64 // Accumulator for pulse timing

	// Track which companions had bonuses last frame to detect changes
	lastBonusState map[uint64]weatherCompanionBonusStateCache
}

// weatherCompanionBonusStateCache stores previous bonus state for change detection.
type weatherCompanionBonusStateCache struct {
	hasAttackBonus   bool
	hasDefenseBonus  bool
	hasDefensePenalt bool
	hasSpeedBonus    bool
	hasSpeedPenalty  bool
	weatherType      string
	companionType    string
}

// NewWeatherCompanionBonusParticleSystem creates a new weather companion bonus particle system.
func NewWeatherCompanionBonusParticleSystem(world *World, seed int64) *WeatherCompanionBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_companion_bonus_particle")
		logEntry.Debug("weather companion bonus particle system created")
	}

	return &WeatherCompanionBonusParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		genreID:        "fantasy",
		pulseInterval:  2.5, // Pulse every 2.5 seconds for active bonuses
		lastBonusState: make(map[uint64]weatherCompanionBonusStateCache, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *WeatherCompanionBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetWeatherCompanionBonusSystem sets the weather companion bonus system reference.
func (s *WeatherCompanionBonusParticleSystem) SetWeatherCompanionBonusSystem(wcbs *WeatherCompanionBonusSystem) {
	s.weatherCompanionBonusSystem = wcbs
	if s.logger != nil {
		s.logger.Debug("weather companion bonus system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *WeatherCompanionBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and spawns particles for weather companion bonuses.
func (s *WeatherCompanionBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	shouldPulse := s.timeSinceEmit >= s.pulseInterval

	for _, entity := range entities {
		s.processEntity(entity, shouldPulse)
	}

	if shouldPulse {
		s.timeSinceEmit = 0
	}
}

// processEntity handles particle effects for a single companion's weather bonuses.
func (s *WeatherCompanionBonusParticleSystem) processEntity(entity *Entity, shouldPulse bool) {
	// Only process companions
	if !entity.HasComponent("companion") {
		return
	}

	comp, ok := entity.GetComponent("weather_companion_bonus")
	if !ok {
		// No bonus - cleanup cached state
		delete(s.lastBonusState, entity.ID)
		return
	}

	bonus, ok := comp.(*WeatherCompanionBonusComponent)
	if !ok {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Build current bonus state
	current := weatherCompanionBonusStateCache{
		hasAttackBonus:   bonus.AttackBonus > 1.01,
		hasDefenseBonus:  bonus.DefenseBonus > 1.01,
		hasDefensePenalt: bonus.DefenseBonus < 0.99,
		hasSpeedBonus:    bonus.SpeedBonus > 1.01,
		hasSpeedPenalty:  bonus.SpeedBonus < 0.99,
		weatherType:      bonus.WeatherType,
		companionType:    bonus.CompanionTypeName,
	}

	// Check for state change (new bonus acquired)
	last, hadLast := s.lastBonusState[entity.ID]
	stateChanged := !hadLast || s.bonusStateChanged(last, current)

	// Spawn particles on state change or periodic pulse
	if stateChanged {
		s.spawnBonusAcquiredParticles(entity.ID, pos.X, pos.Y, bonus, current)
		s.lastBonusState[entity.ID] = current
	} else if shouldPulse && s.hasActiveBonus(current) {
		s.spawnActiveBonusPulse(entity.ID, pos.X, pos.Y, bonus)
	}
}

// bonusStateChanged returns true if any bonus type changed.
func (s *WeatherCompanionBonusParticleSystem) bonusStateChanged(last, current weatherCompanionBonusStateCache) bool {
	return last.hasAttackBonus != current.hasAttackBonus ||
		last.hasDefenseBonus != current.hasDefenseBonus ||
		last.hasDefensePenalt != current.hasDefensePenalt ||
		last.hasSpeedBonus != current.hasSpeedBonus ||
		last.hasSpeedPenalty != current.hasSpeedPenalty ||
		last.weatherType != current.weatherType
}

// hasActiveBonus returns true if companion has any weather bonus/penalty.
func (s *WeatherCompanionBonusParticleSystem) hasActiveBonus(state weatherCompanionBonusStateCache) bool {
	return state.hasAttackBonus || state.hasDefenseBonus || state.hasDefensePenalt ||
		state.hasSpeedBonus || state.hasSpeedPenalty
}

// spawnBonusAcquiredParticles creates particles when a weather bonus is first acquired.
func (s *WeatherCompanionBonusParticleSystem) spawnBonusAcquiredParticles(entityID uint64, x, y float64, bonus *WeatherCompanionBonusComponent, state weatherCompanionBonusStateCache) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// Attack bonus - aggressive upward particles
	if state.hasAttackBonus {
		config := s.getAttackBonusConfig(effectSeed, bonus.AttackBonus, bonus.WeatherType)
		s.particleSystem.SpawnParticles(s.world, config, x, y-8)
	}

	// Defense bonus - protective shimmer around companion
	if state.hasDefenseBonus {
		config := s.getDefenseBonusConfig(effectSeed+1, bonus.DefenseBonus, bonus.WeatherType)
		s.particleSystem.SpawnParticles(s.world, config, x, y)
	}

	// Defense penalty - weakening drip effect
	if state.hasDefensePenalt {
		config := s.getDefensePenaltyConfig(effectSeed+2, bonus.DefenseBonus, bonus.WeatherType)
		s.particleSystem.SpawnParticles(s.world, config, x, y+4)
	}

	// Speed bonus - swift trail effect
	if state.hasSpeedBonus {
		config := s.getSpeedBonusConfig(effectSeed+3, bonus.SpeedBonus, bonus.WeatherType)
		s.particleSystem.SpawnParticles(s.world, config, x-8, y)
	}

	// Speed penalty - sluggish drag effect
	if state.hasSpeedPenalty {
		config := s.getSpeedPenaltyConfig(effectSeed+4, bonus.SpeedBonus, bonus.WeatherType)
		s.particleSystem.SpawnParticles(s.world, config, x, y+2)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"weather_type":   bonus.WeatherType,
			"companion_type": bonus.CompanionTypeName,
			"attack_bonus":   bonus.AttackBonus,
			"defense_bonus":  bonus.DefenseBonus,
			"speed_bonus":    bonus.SpeedBonus,
		}).Debug("weather companion bonus particles spawned")
	}
}

// spawnActiveBonusPulse creates subtle reminder particles for active bonuses.
func (s *WeatherCompanionBonusParticleSystem) spawnActiveBonusPulse(entityID uint64, x, y float64, bonus *WeatherCompanionBonusComponent) {
	effectSeed := s.seed + int64(s.timeSinceEmit*1000) + int64(entityID)

	// Only spawn 1-2 reminder particles
	config := s.getPulseConfig(effectSeed, bonus)
	s.particleSystem.SpawnParticles(s.world, config, x, y-4)
}

// getAttackBonusConfig returns particles for weather attack bonus.
func (s *WeatherCompanionBonusParticleSystem) getAttackBonusConfig(seed int64, attackBonus float64, weatherType string) particles.Config {
	count := 5
	if attackBonus >= 1.20 {
		count = 8
	}

	particleType := s.getWeatherParticleType(weatherType, true)

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.7,
		SpreadX:  20.0,
		SpreadY:  12.0,
		Gravity:  -60.0, // Rise upward aggressively
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"weather_attack_bonus": true, "attack_mult": attackBonus},
	}
}

// getDefenseBonusConfig returns particles for weather defense bonus.
func (s *WeatherCompanionBonusParticleSystem) getDefenseBonusConfig(seed int64, defenseBonus float64, weatherType string) particles.Config {
	count := 6
	if defenseBonus >= 1.15 {
		count = 9
	}

	particleType := s.getWeatherParticleType(weatherType, true)

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.6,
		SpreadX:  25.0,
		SpreadY:  25.0,
		Gravity:  0.0, // Hover protectively
		MinSize:  2.0,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"weather_defense_bonus": true, "defense_mult": defenseBonus},
	}
}

// getDefensePenaltyConfig returns particles for weather defense penalty.
func (s *WeatherCompanionBonusParticleSystem) getDefensePenaltyConfig(seed int64, defenseBonus float64, weatherType string) particles.Config {
	count := 4
	if defenseBonus <= 0.85 {
		count = 6
	}

	return particles.Config{
		Type:     particles.ParticleSpark,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  18.0,
		SpreadY:  8.0,
		Gravity:  70.0, // Fall downward
		MinSize:  2.0,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"weather_defense_penalty": true, "defense_mult": defenseBonus},
	}
}

// getSpeedBonusConfig returns particles for weather speed bonus.
func (s *WeatherCompanionBonusParticleSystem) getSpeedBonusConfig(seed int64, speedBonus float64, weatherType string) particles.Config {
	count := 4
	if speedBonus >= 1.10 {
		count = 6
	}

	particleType := s.getWeatherParticleType(weatherType, true)

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  30.0,
		SpreadY:  10.0,
		Gravity:  -10.0, // Slight lift
		MinSize:  1.5,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"weather_speed_bonus": true, "speed_mult": speedBonus},
	}
}

// getSpeedPenaltyConfig returns particles for weather speed penalty.
func (s *WeatherCompanionBonusParticleSystem) getSpeedPenaltyConfig(seed int64, speedBonus float64, weatherType string) particles.Config {
	count := 3
	if speedBonus <= 0.85 {
		count = 5
	}

	return particles.Config{
		Type:     particles.ParticleDust,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.6,
		SpreadX:  15.0,
		SpreadY:  15.0,
		Gravity:  20.0, // Slow drag down
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"weather_speed_penalty": true, "speed_mult": speedBonus},
	}
}

// getPulseConfig returns subtle reminder particles for active weather bonuses.
func (s *WeatherCompanionBonusParticleSystem) getPulseConfig(seed int64, bonus *WeatherCompanionBonusComponent) particles.Config {
	particleType := s.getWeatherParticleType(bonus.WeatherType, bonus.AttackBonus > 1.0 || bonus.DefenseBonus > 1.0)

	return particles.Config{
		Type:     particleType,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.4,
		SpreadX:  12.0,
		SpreadY:  12.0,
		Gravity:  -15.0,
		MinSize:  1.5,
		MaxSize:  2.5,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"weather_companion_pulse": true},
	}
}

// getWeatherParticleType returns genre and weather-appropriate particle type.
func (s *WeatherCompanionBonusParticleSystem) getWeatherParticleType(weatherType string, isBonus bool) particles.ParticleType {
	// Weather-specific particles
	switch weatherType {
	case "Rain", "NeonRain":
		if isBonus {
			return particles.ParticleSparkle
		}
		return particles.ParticleSpark
	case "Snow":
		return particles.ParticleSparkle
	case "Fog", "Smog":
		return particles.ParticleSmoke
	case "Dust", "Ash":
		return particles.ParticleDust
	case "Radiation", "BloodRain":
		if s.genreID == "horror" {
			return particles.ParticleSmoke
		}
		return particles.ParticleSpark
	}

	// Genre fallback
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle
	case "scifi", "cyberpunk":
		return particles.ParticleSpark
	case "horror":
		return particles.ParticleSmoke
	case "postapoc":
		return particles.ParticleDust
	default:
		return particles.ParticleSparkle
	}
}
