// Package engine provides the TerrainCombatBonusParticleSystem for visual terrain bonus feedback.
// This system connects TerrainCombatBonusSystem with ParticleSystem to spawn genre-aware particle
// effects when entities gain terrain-based combat bonuses (high ground, cover, water penalties).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// TerrainCombatBonusParticleSystem spawns particles when entities gain terrain combat bonuses.
// It monitors TerrainCombatBonusComponent changes and provides visual feedback with genre-aware
// particle effects for high ground advantage, cover bonus, and water vulnerability.
type TerrainCombatBonusParticleSystem struct {
	world                    *World
	particleSystem           *ParticleSystem
	terrainCombatBonusSystem *TerrainCombatBonusSystem
	genreID                  string
	seed                     int64
	rng                      *rand.Rand
	logger                   *logrus.Entry

	// Configuration
	pulseInterval float64 // Time between bonus indicator particles
	timeSinceEmit float64 // Accumulator for pulse timing

	// Track which entities had bonuses last frame to detect changes
	lastBonusState map[uint64]bonusStateCache
}

// bonusStateCache stores previous bonus state for change detection.
type bonusStateCache struct {
	hasDamageBonus  bool
	hasDefenseBuff  bool
	hasDefenseDebuf bool
	hasEvasionBonus bool
	terrainType     string
}

// NewTerrainCombatBonusParticleSystem creates a new terrain combat bonus particle system.
func NewTerrainCombatBonusParticleSystem(world *World, seed int64) *TerrainCombatBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_combat_bonus_particle")
		logEntry.Debug("terrain combat bonus particle system created")
	}

	return &TerrainCombatBonusParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		genreID:        "fantasy",
		pulseInterval:  2.0, // Pulse every 2 seconds for active bonuses
		lastBonusState: make(map[uint64]bonusStateCache, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *TerrainCombatBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetTerrainCombatBonusSystem sets the terrain bonus system reference.
func (s *TerrainCombatBonusParticleSystem) SetTerrainCombatBonusSystem(tcbs *TerrainCombatBonusSystem) {
	s.terrainCombatBonusSystem = tcbs
	if s.logger != nil {
		s.logger.Debug("terrain combat bonus system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *TerrainCombatBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and spawns particles for terrain combat bonuses.
func (s *TerrainCombatBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
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

// processEntity handles particle effects for a single entity's terrain bonuses.
func (s *TerrainCombatBonusParticleSystem) processEntity(entity *Entity, shouldPulse bool) {
	comp, ok := entity.GetComponent("terrain_combat_bonus")
	if !ok {
		// No bonus - cleanup cached state
		delete(s.lastBonusState, entity.ID)
		return
	}

	bonus, ok := comp.(*TerrainCombatBonusComponent)
	if !ok {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Build current bonus state
	current := bonusStateCache{
		hasDamageBonus:  bonus.DamageBonus > 1.0,
		hasDefenseBuff:  bonus.DefenseBonus > 1.0,
		hasDefenseDebuf: bonus.DefenseBonus < 1.0,
		hasEvasionBonus: bonus.EvasionBonus > 0.01,
		terrainType:     bonus.TerrainType,
	}

	// Check for state change (new bonus acquired)
	last, hadLast := s.lastBonusState[entity.ID]
	stateChanged := !hadLast || s.bonusStateChanged(last, current)

	// Spawn particles on state change or periodic pulse
	if stateChanged {
		s.spawnBonusAcquiredParticles(entity.ID, pos.X, pos.Y, bonus)
		s.lastBonusState[entity.ID] = current
	} else if shouldPulse && s.hasActiveBonus(current) {
		s.spawnActiveBonusPulse(entity.ID, pos.X, pos.Y, bonus)
	}
}

// bonusStateChanged returns true if any bonus type changed.
func (s *TerrainCombatBonusParticleSystem) bonusStateChanged(last, current bonusStateCache) bool {
	return last.hasDamageBonus != current.hasDamageBonus ||
		last.hasDefenseBuff != current.hasDefenseBuff ||
		last.hasDefenseDebuf != current.hasDefenseDebuf ||
		last.hasEvasionBonus != current.hasEvasionBonus ||
		last.terrainType != current.terrainType
}

// hasActiveBonus returns true if entity has any terrain bonus/penalty.
func (s *TerrainCombatBonusParticleSystem) hasActiveBonus(state bonusStateCache) bool {
	return state.hasDamageBonus || state.hasDefenseBuff || state.hasDefenseDebuf || state.hasEvasionBonus
}

// spawnBonusAcquiredParticles creates particles when a terrain bonus is first acquired.
func (s *TerrainCombatBonusParticleSystem) spawnBonusAcquiredParticles(entityID uint64, x, y float64, bonus *TerrainCombatBonusComponent) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// High ground/damage bonus - upward sparkles
	if bonus.DamageBonus > 1.0 {
		config := s.getHighGroundConfig(effectSeed, bonus.DamageBonus)
		s.particleSystem.SpawnParticles(s.world, config, x, y-10)
	}

	// Cover/evasion bonus - shield shimmer
	if bonus.EvasionBonus > 0.01 {
		config := s.getCoverBonusConfig(effectSeed+1, bonus.EvasionBonus)
		s.particleSystem.SpawnParticles(s.world, config, x, y)
	}

	// Water penalty - downward drip effect
	if bonus.DefenseBonus < 1.0 {
		config := s.getWaterPenaltyConfig(effectSeed+2, bonus.DefenseBonus)
		s.particleSystem.SpawnParticles(s.world, config, x, y+5)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entityID,
			"terrain_type":  bonus.TerrainType,
			"damage_bonus":  bonus.DamageBonus,
			"defense_bonus": bonus.DefenseBonus,
			"evasion_bonus": bonus.EvasionBonus,
		}).Debug("terrain bonus particles spawned")
	}
}

// spawnActiveBonusPulse creates subtle reminder particles for active bonuses.
func (s *TerrainCombatBonusParticleSystem) spawnActiveBonusPulse(entityID uint64, x, y float64, bonus *TerrainCombatBonusComponent) {
	effectSeed := s.seed + int64(s.timeSinceEmit*1000) + int64(entityID)

	// Only spawn 1-2 reminder particles
	if bonus.DamageBonus > 1.0 || bonus.EvasionBonus > 0.01 {
		config := s.getPulseConfig(effectSeed, bonus)
		s.particleSystem.SpawnParticles(s.world, config, x, y-5)
	}
}

// getHighGroundConfig returns particles for high ground damage bonus.
func (s *TerrainCombatBonusParticleSystem) getHighGroundConfig(seed int64, damageBonus float64) particles.Config {
	count := 6
	if damageBonus >= 1.10 {
		count = 8
	}

	particleType := s.getHighGroundParticleType()

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.7,
		SpreadX:  25.0,
		SpreadY:  15.0,
		Gravity:  -50.0, // Rise upward
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"high_ground": true, "damage_bonus": damageBonus},
	}
}

// getCoverBonusConfig returns particles for cover/evasion bonus.
func (s *TerrainCombatBonusParticleSystem) getCoverBonusConfig(seed int64, evasionBonus float64) particles.Config {
	count := 5
	if evasionBonus >= 0.10 {
		count = 7
	}

	particleType := s.getCoverParticleType()

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.6,
		SpreadX:  30.0,
		SpreadY:  30.0,
		Gravity:  0.0, // Hover around entity
		MinSize:  2.0,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"cover_bonus": true, "evasion": evasionBonus},
	}
}

// getWaterPenaltyConfig returns particles for water defense penalty.
func (s *TerrainCombatBonusParticleSystem) getWaterPenaltyConfig(seed int64, defenseBonus float64) particles.Config {
	count := 4
	if defenseBonus <= 0.80 {
		count = 6 // More drips for worse penalty
	}

	return particles.Config{
		Type:     particles.ParticleSpark, // Water droplet effect
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  20.0,
		SpreadY:  10.0,
		Gravity:  80.0, // Fall downward
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"water_penalty": true, "defense": defenseBonus},
	}
}

// getPulseConfig returns subtle reminder particles for active bonuses.
func (s *TerrainCombatBonusParticleSystem) getPulseConfig(seed int64, bonus *TerrainCombatBonusComponent) particles.Config {
	particleType := particles.ParticleSparkle
	if s.genreID == "scifi" || s.genreID == "cyberpunk" {
		particleType = particles.ParticleSpark
	}

	return particles.Config{
		Type:     particleType,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.4,
		SpreadX:  15.0,
		SpreadY:  15.0,
		Gravity:  -20.0,
		MinSize:  1.5,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"terrain_pulse": true},
	}
}

// getHighGroundParticleType returns genre-appropriate particles for high ground.
func (s *TerrainCombatBonusParticleSystem) getHighGroundParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical elevation aura
	case "scifi":
		return particles.ParticleSpark // Targeting system active
	case "horror":
		return particles.ParticleSmoke // Dark advantage mist
	case "cyberpunk":
		return particles.ParticleSpark // Tactical HUD feedback
	case "postapoc":
		return particles.ParticleDust // Vantage point dust
	default:
		return particles.ParticleSparkle
	}
}

// getCoverParticleType returns genre-appropriate particles for cover bonus.
func (s *TerrainCombatBonusParticleSystem) getCoverParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Protective ward shimmer
	case "scifi":
		return particles.ParticleSpark // Shield deflector active
	case "horror":
		return particles.ParticleSmoke // Concealing shadows
	case "cyberpunk":
		return particles.ParticleSpark // Cover indicator HUD
	case "postapoc":
		return particles.ParticleDust // Debris cover shimmer
	default:
		return particles.ParticleSparkle
	}
}
