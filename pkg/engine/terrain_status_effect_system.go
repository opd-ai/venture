// Package engine provides the TerrainStatusEffectSystem which applies
// elemental status effects based on the terrain tile entities are standing on.
// This bridges terrain generation with status effect and elemental combo systems.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainStatusEffectSystem applies status effects to entities based on
// the terrain tile they are standing on. This connects terrain types with
// the elemental status effect system to enable elemental combos.
//
// Terrain-to-status mappings:
//   - TileWaterShallow → "wet" status (enables ice/shock combos)
//   - TileLavaFlow → "burning" status (enables fire combos)
//   - TileTrapDoor → "chilled" status in horror genre
//
// Genre-specific variations apply different effect magnitudes and durations.
type TerrainStatusEffectSystem struct {
	world    *World
	terrain  *terrain.Terrain
	rng      *rand.Rand
	logger   *logrus.Entry
	genreID  string
	tileSize int

	// Cooldown tracking per entity to prevent effect stacking
	effectCooldowns map[uint64]map[string]float64 // entityID -> effectType -> remaining cooldown

	// Configuration
	baseDuration     float64 // Base duration for terrain effects
	cooldownTime     float64 // Time between reapplication of same effect
	genreMultipliers map[string]float64
	tickInterval     float64 // Tick interval for DoT effects
}

// terrainEffectMapping defines how a terrain type maps to a status effect.
type terrainEffectMapping struct {
	terrainType terrain.TileType
	effectType  string
	magnitude   float64
	isDot       bool // Is damage-over-time effect
}

// baseTerrainEffects defines the terrain-to-effect mappings.
var baseTerrainEffects = []terrainEffectMapping{
	{terrain.TileWaterShallow, "wet", 1.0, false},
	{terrain.TileLavaFlow, "burning", 8.0, true},
}

// NewTerrainStatusEffectSystem creates a new terrain status effect system.
func NewTerrainStatusEffectSystem(world *World, seed int64) *TerrainStatusEffectSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_status_effect")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("terrain status effect system created")
		}
	}

	return &TerrainStatusEffectSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		tileSize:        32,
		effectCooldowns: make(map[uint64]map[string]float64),
		baseDuration:    3.0,
		cooldownTime:    1.5,
		tickInterval:    0.5,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     0.8, // Sci-fi: environmental suits provide protection
			"horror":    1.3, // Horror: terrain effects are more punishing
			"cyberpunk": 0.9, // Cyberpunk: augmentations provide some resistance
			"postapoc":  1.2, // Post-apoc: harsh environment
		},
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainStatusEffectSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear cooldowns when terrain changes
	s.effectCooldowns = make(map[uint64]map[string]float64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainStatusEffectSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific effect modifiers.
func (s *TerrainStatusEffectSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes entities and applies terrain-based status effects.
func (s *TerrainStatusEffectSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	// Update cooldowns
	s.updateCooldowns(deltaTime)

	for _, entity := range entities {
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Skip dead entities
		health := entity.GetHealth()
		if health != nil && health.IsDead() {
			continue
		}

		// Convert world position to tile coordinates
		tileX := int(pos.X) / s.tileSize
		tileY := int(pos.Y) / s.tileSize
		tileType := s.terrain.GetTile(tileX, tileY)

		// Check terrain effects
		s.applyTerrainEffect(entity, tileType)
	}
}

// updateCooldowns decrements all cooldowns by deltaTime.
func (s *TerrainStatusEffectSystem) updateCooldowns(deltaTime float64) {
	for entityID, cooldowns := range s.effectCooldowns {
		for effectType, remaining := range cooldowns {
			remaining -= deltaTime
			if remaining <= 0 {
				delete(cooldowns, effectType)
			} else {
				cooldowns[effectType] = remaining
			}
		}
		if len(cooldowns) == 0 {
			delete(s.effectCooldowns, entityID)
		}
	}
}

// isOnCooldown returns true if the entity has an active cooldown for the effect.
func (s *TerrainStatusEffectSystem) isOnCooldown(entityID uint64, effectType string) bool {
	if cooldowns, ok := s.effectCooldowns[entityID]; ok {
		_, hasCooldown := cooldowns[effectType]
		return hasCooldown
	}
	return false
}

// startCooldown starts a cooldown for an effect on an entity.
func (s *TerrainStatusEffectSystem) startCooldown(entityID uint64, effectType string) {
	if _, ok := s.effectCooldowns[entityID]; !ok {
		s.effectCooldowns[entityID] = make(map[string]float64)
	}
	s.effectCooldowns[entityID][effectType] = s.cooldownTime
}

// applyTerrainEffect applies status effects based on terrain type.
func (s *TerrainStatusEffectSystem) applyTerrainEffect(entity *Entity, tileType terrain.TileType) {
	for _, mapping := range baseTerrainEffects {
		if mapping.terrainType != tileType {
			continue
		}

		// Skip if on cooldown
		if s.isOnCooldown(entity.ID, mapping.effectType) {
			continue
		}

		// Skip if entity already has this effect
		if s.hasActiveEffect(entity, mapping.effectType) {
			continue
		}

		// Apply effect
		s.applyEffect(entity, mapping)

		// Start cooldown
		s.startCooldown(entity.ID, mapping.effectType)
	}

	// Genre-specific terrain effects
	s.applyGenreTerrainEffects(entity, tileType)
}

// hasActiveEffect returns true if the entity has an active non-expired effect.
func (s *TerrainStatusEffectSystem) hasActiveEffect(entity *Entity, effectType string) bool {
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if ok && effect.EffectType == effectType && !effect.IsExpired() {
			return true
		}
	}
	return false
}

// applyEffect creates and applies a status effect to an entity.
func (s *TerrainStatusEffectSystem) applyEffect(entity *Entity, mapping terrainEffectMapping) {
	genreMult := s.getGenreMultiplier()
	duration := s.baseDuration * genreMult

	var tickInterval float64
	if mapping.isDot {
		tickInterval = s.tickInterval
	}

	effect := NewStatusEffectComponent(
		mapping.effectType,
		mapping.magnitude*genreMult,
		duration,
		tickInterval,
	)
	entity.AddComponent(effect)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": mapping.effectType,
			"magnitude":   mapping.magnitude * genreMult,
			"duration":    duration,
			"terrain":     mapping.terrainType.String(),
		}).Debug("terrain status effect applied")
	}
}

// applyGenreTerrainEffects applies genre-specific terrain effects.
func (s *TerrainStatusEffectSystem) applyGenreTerrainEffects(entity *Entity, tileType terrain.TileType) {
	switch s.genreID {
	case "horror":
		// Horror: trap doors emit cold aura
		if tileType == terrain.TileTrapDoor && !s.isOnCooldown(entity.ID, "chilled") {
			if !s.hasActiveEffect(entity, "chilled") {
				effect := NewStatusEffectComponent("chilled", 0.5, 2.0, 0)
				entity.AddComponent(effect)
				s.startCooldown(entity.ID, "chilled")
			}
		}
	case "scifi":
		// Sci-fi: platforms can have electrical discharge
		if tileType == terrain.TilePlatform && s.rng.Float64() < 0.05 { // 5% chance per update
			if !s.isOnCooldown(entity.ID, "shocked") && !s.hasActiveEffect(entity, "shocked") {
				effect := NewStatusEffectComponent("shocked", 3.0, 1.0, 0)
				entity.AddComponent(effect)
				s.startCooldown(entity.ID, "shocked")
			}
		}
	}
}

// getGenreMultiplier returns the effect multiplier for the current genre.
func (s *TerrainStatusEffectSystem) getGenreMultiplier() float64 {
	if mult, ok := s.genreMultipliers[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// GetBaseDuration returns the base effect duration.
func (s *TerrainStatusEffectSystem) GetBaseDuration() float64 {
	return s.baseDuration
}

// SetBaseDuration sets the base effect duration.
func (s *TerrainStatusEffectSystem) SetBaseDuration(duration float64) {
	if duration > 0 {
		s.baseDuration = duration
	}
}

// GetCooldownTime returns the cooldown time between effect reapplications.
func (s *TerrainStatusEffectSystem) GetCooldownTime() float64 {
	return s.cooldownTime
}

// SetCooldownTime sets the cooldown time between effect reapplications.
func (s *TerrainStatusEffectSystem) SetCooldownTime(cooldown float64) {
	if cooldown > 0 {
		s.cooldownTime = cooldown
	}
}
