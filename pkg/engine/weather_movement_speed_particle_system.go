// Package engine provides the WeatherMovementSpeedParticleSystem for visual feedback
// when weather conditions affect entity movement speed. This system connects
// WeatherMovementSpeedSystem with ParticleSystem to spawn genre-aware particle effects
// when entities experience weather-based movement modifiers.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherMovementSpeedParticleSystem spawns particles when weather affects movement.
// It monitors WeatherMovementSpeedSystem modifier changes and provides visual feedback
// with genre-aware particle effects for speed boosts (rain slickness) and speed
// penalties (snow resistance, sandstorm drag).
type WeatherMovementSpeedParticleSystem struct {
	world                      *World
	particleSystem             *ParticleSystem
	weatherMovementSpeedSystem *WeatherMovementSpeedSystem
	genreID                    string
	seed                       int64
	rng                        *rand.Rand
	logger                     *logrus.Entry

	// Configuration
	pulseInterval float64 // Time between active modifier indicator particles
	timeSinceEmit float64 // Accumulator for pulse timing
	minSpeedSq    float64 // Minimum squared speed to show particles (avoid static entities)

	// Track which entities had modifiers last frame to detect changes
	lastModifierState map[uint64]float64
}

// NewWeatherMovementSpeedParticleSystem creates a new weather movement speed particle system.
func NewWeatherMovementSpeedParticleSystem(world *World, seed int64) *WeatherMovementSpeedParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_movement_speed_particle")
		logEntry.Debug("weather movement speed particle system created")
	}

	return &WeatherMovementSpeedParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		genreID:           "fantasy",
		pulseInterval:     1.8,  // Pulse every 1.8 seconds for active modifiers
		minSpeedSq:        25.0, // sqrt(25) = 5 px/s minimum
		lastModifierState: make(map[uint64]float64, 64),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *WeatherMovementSpeedParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetWeatherMovementSpeedSystem sets the weather movement speed system reference.
func (s *WeatherMovementSpeedParticleSystem) SetWeatherMovementSpeedSystem(wmss *WeatherMovementSpeedSystem) {
	s.weatherMovementSpeedSystem = wmss
	if s.logger != nil {
		s.logger.Debug("weather movement speed system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *WeatherMovementSpeedParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and spawns particles for weather movement speed modifiers.
func (s *WeatherMovementSpeedParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil || s.weatherMovementSpeedSystem == nil {
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

// processEntity handles particle effects for a single entity's weather speed modifier.
func (s *WeatherMovementSpeedParticleSystem) processEntity(entity *Entity, shouldPulse bool) {
	// Only process entities with velocity component
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	// Check if entity is moving fast enough to show particles
	speedSq := vel.VX*vel.VX + vel.VY*vel.VY
	if speedSq < s.minSpeedSq {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Get current modifier from WeatherMovementSpeedSystem
	currentModifier := s.weatherMovementSpeedSystem.GetMovementSpeedModifier(entity.ID)

	// Check for state change
	lastModifier, hadLast := s.lastModifierState[entity.ID]
	modifierChanged := !hadLast || !modifiersEqual(lastModifier, currentModifier)

	// No modifier active
	if currentModifier == 1.0 {
		if hadLast && lastModifier != 1.0 {
			// Modifier was removed - clear state
			delete(s.lastModifierState, entity.ID)
		}
		return
	}

	// Spawn particles on modifier change or periodic pulse
	if modifierChanged {
		s.spawnModifierChangedParticles(entity.ID, pos.X, pos.Y, currentModifier, vel)
		s.lastModifierState[entity.ID] = currentModifier
	} else if shouldPulse {
		s.spawnActivePulse(entity.ID, pos.X, pos.Y, currentModifier, vel)
	}
}

// modifiersEqual checks if two modifiers are effectively equal (within 0.01).
func modifiersEqual(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}

// spawnModifierChangedParticles creates particles when a weather modifier is applied/changed.
func (s *WeatherMovementSpeedParticleSystem) spawnModifierChangedParticles(entityID uint64, x, y, modifier float64, vel *VelocityComponent) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	isBonus := modifier > 1.0
	config := s.getModifierConfig(effectSeed, modifier, isBonus, vel)
	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"modifier":  modifier,
			"is_bonus":  isBonus,
		}).Debug("weather movement speed particles spawned")
	}
}

// spawnActivePulse creates subtle reminder particles for active weather modifiers.
func (s *WeatherMovementSpeedParticleSystem) spawnActivePulse(entityID uint64, x, y, modifier float64, vel *VelocityComponent) {
	effectSeed := s.seed + int64(s.timeSinceEmit*1000) + int64(entityID)

	isBonus := modifier > 1.0
	config := s.getPulseConfig(effectSeed, modifier, isBonus, vel)

	// Offset particles slightly behind movement direction
	offsetX, offsetY := 0.0, 0.0
	if vel.VX != 0 || vel.VY != 0 {
		// Normalize velocity and offset backward
		mag := 1.0 / (vel.VX*vel.VX + vel.VY*vel.VY + 0.01)
		if mag > 0.1 {
			mag = 0.1
		}
		offsetX = -vel.VX * mag * 8
		offsetY = -vel.VY * mag * 8
	}

	s.particleSystem.SpawnParticles(s.world, config, x+offsetX, y+offsetY)
}

// getModifierConfig returns particle configuration for a weather movement modifier.
func (s *WeatherMovementSpeedParticleSystem) getModifierConfig(seed int64, modifier float64, isBonus bool, vel *VelocityComponent) particles.Config {
	if isBonus {
		return s.getSpeedBoostConfig(seed, modifier, vel)
	}
	return s.getSpeedPenaltyConfig(seed, modifier, vel)
}

// getSpeedBoostConfig returns particles for weather speed bonus (rain slickness, etc).
func (s *WeatherMovementSpeedParticleSystem) getSpeedBoostConfig(seed int64, modifier float64, vel *VelocityComponent) particles.Config {
	// More particles for stronger bonus
	count := 4
	if modifier >= 1.08 {
		count = 6
	}

	particleType := s.getBoostParticleType()

	// Calculate spread based on velocity direction
	spreadX := 20.0
	spreadY := 8.0

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  spreadX,
		SpreadY:  spreadY,
		Gravity:  -25.0, // Slight lift for speed boost effect
		MinSize:  1.5,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"weather_speed_boost": true, "speed_mult": modifier},
	}
}

// getSpeedPenaltyConfig returns particles for weather speed penalty (snow, sandstorm, etc).
func (s *WeatherMovementSpeedParticleSystem) getSpeedPenaltyConfig(seed int64, modifier float64, vel *VelocityComponent) particles.Config {
	// More particles for stronger penalty
	count := 3
	if modifier <= 0.85 {
		count = 5
	}

	particleType := s.getPenaltyParticleType()

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.6,
		SpreadX:  15.0,
		SpreadY:  12.0,
		Gravity:  30.0, // Drag down for slowdown effect
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"weather_speed_penalty": true, "speed_mult": modifier},
	}
}

// getPulseConfig returns subtle reminder particles for active weather modifiers.
func (s *WeatherMovementSpeedParticleSystem) getPulseConfig(seed int64, modifier float64, isBonus bool, vel *VelocityComponent) particles.Config {
	var particleType particles.ParticleType
	var gravity float64

	if isBonus {
		particleType = s.getBoostParticleType()
		gravity = -15.0
	} else {
		particleType = s.getPenaltyParticleType()
		gravity = 20.0
	}

	return particles.Config{
		Type:     particleType,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.35,
		SpreadX:  10.0,
		SpreadY:  10.0,
		Gravity:  gravity,
		MinSize:  1.5,
		MaxSize:  2.5,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"weather_speed_pulse": true},
	}
}

// getBoostParticleType returns genre-appropriate particle type for speed boosts.
func (s *WeatherMovementSpeedParticleSystem) getBoostParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical speed boost
	case "scifi":
		return particles.ParticleSpark // Tech boost effect
	case "cyberpunk":
		return particles.ParticleSpark // Neon trail
	case "horror":
		return particles.ParticleSpark // Muted spark
	case "postapoc":
		return particles.ParticleDust // Kicked-up debris
	default:
		return particles.ParticleSparkle
	}
}

// getPenaltyParticleType returns genre-appropriate particle type for speed penalties.
func (s *WeatherMovementSpeedParticleSystem) getPenaltyParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleDust // Snow/magical resistance
	case "scifi":
		return particles.ParticleDust // Particle drag
	case "cyberpunk":
		return particles.ParticleSmoke // Urban pollution drag
	case "horror":
		return particles.ParticleSmoke // Oppressive atmosphere
	case "postapoc":
		return particles.ParticleDust // Radiation/ash drag
	default:
		return particles.ParticleDust
	}
}
