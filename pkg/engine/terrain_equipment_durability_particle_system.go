// Package engine provides the TerrainEquipmentDurabilityParticleSystem for visual terrain damage feedback.
// This system connects TerrainEquipmentDurabilitySystem with ParticleSystem to spawn genre-aware particle
// effects when equipment takes damage from terrain hazards (lava, water, traps).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// TerrainEquipmentDurabilityParticleSystem spawns particles when equipment takes terrain damage.
// It monitors durability changes caused by terrain hazards and provides visual feedback with
// genre-aware particle effects for lava (fire/sparks), water (rust/drips), and traps (metal shards).
type TerrainEquipmentDurabilityParticleSystem struct {
	world                            *World
	particleSystem                   *ParticleSystem
	terrainEquipmentDurabilitySystem *TerrainEquipmentDurabilitySystem
	genreID                          string
	seed                             int64
	rng                              *rand.Rand
	logger                           *logrus.Entry
	tileSize                         int
	terrain                          *terrain.Terrain

	// Track last durability state to detect damage events
	lastDurabilityState map[uint64]durabilityStateCache
}

// durabilityStateCache stores previous durability state for change detection.
type durabilityStateCache struct {
	totalDurability int
	damageState     int // 0=pristine, 1=worn, 2=damaged, 3=broken
}

// NewTerrainEquipmentDurabilityParticleSystem creates a new terrain equipment durability particle system.
func NewTerrainEquipmentDurabilityParticleSystem(world *World, seed int64) *TerrainEquipmentDurabilityParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_equipment_durability_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("terrain equipment durability particle system created")
		}
	}

	return &TerrainEquipmentDurabilityParticleSystem{
		world:               world,
		seed:                seed,
		rng:                 rand.New(rand.NewSource(seed)),
		logger:              logEntry,
		genreID:             "fantasy",
		tileSize:            32,
		lastDurabilityState: make(map[uint64]durabilityStateCache, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *TerrainEquipmentDurabilityParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("particle system linked")
	}
}

// SetTerrainEquipmentDurabilitySystem sets the terrain durability system reference.
func (s *TerrainEquipmentDurabilityParticleSystem) SetTerrainEquipmentDurabilitySystem(teds *TerrainEquipmentDurabilitySystem) {
	s.terrainEquipmentDurabilitySystem = teds
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("terrain equipment durability system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *TerrainEquipmentDurabilityParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetTerrain sets the terrain data for tile type lookups.
func (s *TerrainEquipmentDurabilityParticleSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	s.lastDurabilityState = make(map[uint64]durabilityStateCache, 32)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainEquipmentDurabilityParticleSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update processes entities and spawns particles for terrain equipment damage.
func (s *TerrainEquipmentDurabilityParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil || s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity handles particle effects for a single entity's equipment durability.
func (s *TerrainEquipmentDurabilityParticleSystem) processEntity(entity *Entity) {
	// Get equipment component
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	// Get position for particle spawn
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Calculate total durability across all equipped items
	totalDurability, maxDurability := s.calculateTotalDurability(equipComp)
	if maxDurability == 0 {
		return
	}

	// Calculate damage state (0-3)
	currentState := s.calculateDamageState(totalDurability, maxDurability)

	// Check for durability change
	lastState, hasLast := s.lastDurabilityState[entity.ID]
	if hasLast && lastState.totalDurability > totalDurability {
		// Durability decreased - spawn particles based on terrain type
		tileX := int(pos.X) / s.tileSize
		tileY := int(pos.Y) / s.tileSize
		tileType := s.terrain.GetTile(tileX, tileY)

		// Spawn appropriate particles for the terrain hazard
		s.spawnDamageParticles(entity.ID, pos.X, pos.Y, tileType, lastState.totalDurability-totalDurability)

		// If damage state changed (crossed threshold), spawn extra feedback
		if lastState.damageState != currentState {
			s.spawnStateChangeParticles(entity.ID, pos.X, pos.Y, currentState)
		}
	}

	// Update cached state
	s.lastDurabilityState[entity.ID] = durabilityStateCache{
		totalDurability: totalDurability,
		damageState:     currentState,
	}
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *TerrainEquipmentDurabilityParticleSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// calculateTotalDurability sums durability across all equipped items.
func (s *TerrainEquipmentDurabilityParticleSystem) calculateTotalDurability(equipComp *EquipmentComponent) (total, max int) {
	for _, item := range equipComp.Slots {
		if item != nil && item.Stats.DurabilityMax > 0 {
			total += item.Stats.Durability
			max += item.Stats.DurabilityMax
		}
	}
	return total, max
}

// calculateDamageState returns 0=pristine, 1=worn, 2=damaged, 3=broken based on durability percentage.
func (s *TerrainEquipmentDurabilityParticleSystem) calculateDamageState(current, max int) int {
	if max == 0 {
		return 0
	}
	pct := float64(current) / float64(max)
	switch {
	case pct > 0.75:
		return 0 // Pristine
	case pct > 0.50:
		return 1 // Worn
	case pct > 0.25:
		return 2 // Damaged
	default:
		return 3 // Broken
	}
}

// spawnDamageParticles creates particles based on the terrain type causing damage.
func (s *TerrainEquipmentDurabilityParticleSystem) spawnDamageParticles(entityID uint64, x, y float64, tileType terrain.TileType, damageAmount int) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	var config particles.Config

	switch tileType {
	case terrain.TileLavaFlow:
		config = s.getLavaParticleConfig(effectSeed, damageAmount)
	case terrain.TileWaterShallow:
		config = s.getWaterParticleConfig(effectSeed, damageAmount)
	case terrain.TileTrapDoor:
		config = s.getTrapParticleConfig(effectSeed, damageAmount)
	default:
		return // Unknown terrain type, skip particles
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entityID,
			"terrain_type":  tileType,
			"damage_amount": damageAmount,
		}).Debug("terrain equipment damage particles spawned")
	}
}

// spawnStateChangeParticles creates extra particles when equipment crosses a damage threshold.
func (s *TerrainEquipmentDurabilityParticleSystem) spawnStateChangeParticles(entityID uint64, x, y float64, newState int) {
	effectSeed := s.seed + int64(entityID*1000)

	// More severe damage state = more dramatic particles
	count := 3 + newState*2

	particleType := s.getStateChangeParticleType()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  20.0,
		SpreadY:  20.0,
		Gravity:  30.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"state_change": true, "damage_state": newState},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"damage_state": newState,
		}).Debug("equipment state change particles spawned")
	}
}

// getLavaParticleConfig returns fire/ember particles for lava terrain damage.
func (s *TerrainEquipmentDurabilityParticleSystem) getLavaParticleConfig(seed int64, damageAmount int) particles.Config {
	count := 4
	if damageAmount >= 5 {
		count = 6
	}

	particleType := s.getLavaParticleType()

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  15.0,
		SpreadY:  10.0,
		Gravity:  -40.0, // Rise upward like embers
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"lava_damage": true, "heat": true},
	}
}

// getWaterParticleConfig returns rust/drip particles for water terrain damage.
func (s *TerrainEquipmentDurabilityParticleSystem) getWaterParticleConfig(seed int64, damageAmount int) particles.Config {
	count := 3
	if damageAmount >= 2 {
		count = 5
	}

	return particles.Config{
		Type:     particles.ParticleSpark, // Small droplets
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.4,
		SpreadX:  12.0,
		SpreadY:  8.0,
		Gravity:  60.0, // Fall downward
		MinSize:  1.5,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"water_damage": true, "rust": true},
	}
}

// getTrapParticleConfig returns metal shard particles for trap terrain damage.
func (s *TerrainEquipmentDurabilityParticleSystem) getTrapParticleConfig(seed int64, damageAmount int) particles.Config {
	count := 4
	if damageAmount >= 3 {
		count = 6
	}

	particleType := s.getTrapParticleType()

	return particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.35,
		SpreadX:  18.0,
		SpreadY:  18.0,
		Gravity:  50.0,
		MinSize:  2.0,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"trap_damage": true, "metal_shards": true},
	}
}

// getLavaParticleType returns genre-appropriate particles for lava damage.
func (s *TerrainEquipmentDurabilityParticleSystem) getLavaParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical embers
	case "scifi":
		return particles.ParticleSpark // Plasma sparks
	case "horror":
		return particles.ParticleSmoke // Hellfire smoke
	case "cyberpunk":
		return particles.ParticleSpark // Overheating circuits
	case "postapoc":
		return particles.ParticleDebris // Burning debris
	default:
		return particles.ParticleSpark
	}
}

// getTrapParticleType returns genre-appropriate particles for trap damage.
func (s *TerrainEquipmentDurabilityParticleSystem) getTrapParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Mechanical trap sparks
	case "scifi":
		return particles.ParticleSpark // Energy discharge
	case "horror":
		return particles.ParticleDebris // Rusty shrapnel
	case "cyberpunk":
		return particles.ParticleSpark // Malfunctioning tech
	case "postapoc":
		return particles.ParticleDebris // Scrap metal
	default:
		return particles.ParticleDebris
	}
}

// getStateChangeParticleType returns genre-appropriate particles for damage state transitions.
func (s *TerrainEquipmentDurabilityParticleSystem) getStateChangeParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Ward breaking
	case "scifi":
		return particles.ParticleSpark // System failure
	case "horror":
		return particles.ParticleSmoke // Decay mist
	case "cyberpunk":
		return particles.ParticleSpark // Critical malfunction
	case "postapoc":
		return particles.ParticleDust // Equipment crumbling
	default:
		return particles.ParticleDebris
	}
}
