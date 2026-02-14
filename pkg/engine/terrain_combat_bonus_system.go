// Package engine provides the TerrainCombatBonusSystem which bridges
// terrain tile types with entity combat statistics.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainCombatBonusSystem modifies entity combat stats based on the terrain
// tile type they are standing on. This connects terrain generation with the
// CombatSystem by applying tactical bonuses.
//
// Tactical bonuses:
//   - Platform/Bridge: +10% damage (high ground advantage)
//   - Shallow water: -15% defense (vulnerable while wading)
//   - Near walls (adjacent to 2+ walls): +10% evasion (cover bonus)
//
// Genre-specific modifiers apply additional adjustments:
//   - Fantasy: platforms grant +5% spell damage
//   - Scifi: bridges grant +10% accuracy (targeting systems)
//   - Horror: water penalty increased to -20% (panic in water)
//   - Cyberpunk: cover bonus increased to +15% (urban combat training)
//   - Postapoc: all bonuses reduced by 5% (equipment degradation)
type TerrainCombatBonusSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for terrain bonuses to avoid recalculating each frame.
	// Maps entityID -> *TerrainCombatBonusComponent
	bonusCache map[uint64]*TerrainCombatBonusComponent

	// Cache for last known tile position per entity
	lastTileCache map[uint64]terrainBonusTilePos
}

// terrainBonusTilePos stores the tile coordinates for cache invalidation
type terrainBonusTilePos struct {
	tileX, tileY int
}

// TerrainCombatBonusComponent is a transient component that stores terrain-based
// combat modifiers. It is not persisted (recalculated each session from terrain).
type TerrainCombatBonusComponent struct {
	// DamageBonus is a multiplier (1.0 = no bonus, 1.1 = +10%)
	DamageBonus float64

	// DefenseBonus is a multiplier (1.0 = no bonus, 0.85 = -15%)
	DefenseBonus float64

	// EvasionBonus is added to evasion chance (0.0 = no bonus, 0.1 = +10%)
	EvasionBonus float64

	// SpellDamageBonus is a multiplier for spell damage (1.0 = no bonus)
	SpellDamageBonus float64

	// AccuracyBonus is added to hit chance (0.0 = no bonus, 0.1 = +10%)
	AccuracyBonus float64

	// TerrainType is the current terrain providing the bonus
	TerrainType string
}

// Type returns the component type identifier.
func (c *TerrainCombatBonusComponent) Type() string {
	return "terrain_combat_bonus"
}

// NewTerrainCombatBonusSystem creates a new terrain combat bonus system.
func NewTerrainCombatBonusSystem(world *World, seed int64) *TerrainCombatBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_combat_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainCombatBonusSystem created")
	}

	return &TerrainCombatBonusSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		bonusCache:    make(map[uint64]*TerrainCombatBonusComponent, 64),
		lastTileCache: make(map[uint64]terrainBonusTilePos, 64),
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainCombatBonusSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.bonusCache = make(map[uint64]*TerrainCombatBonusComponent, 64)
	s.lastTileCache = make(map[uint64]terrainBonusTilePos, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainCombatBonusSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainCombatBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes all entities and applies combat bonuses based on terrain.
func (s *TerrainCombatBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.updateEntityBonus(entity)
	}
}

// updateEntityBonus updates terrain combat bonuses for a single entity.
func (s *TerrainCombatBonusSystem) updateEntityBonus(entity *Entity) {
	// Only process entities that can engage in combat
	if !entity.HasComponent("health") || !entity.HasComponent("stats") {
		s.removeBonus(entity)
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		s.removeBonus(entity)
		return
	}

	// Convert world position to tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Check if entity moved to a new tile
	lastTile, hasCached := s.lastTileCache[entity.ID]
	if hasCached && lastTile.tileX == tileX && lastTile.tileY == tileY {
		// Still on same tile, bonuses unchanged
		return
	}

	// Update tile cache
	s.lastTileCache[entity.ID] = terrainBonusTilePos{tileX: tileX, tileY: tileY}

	// Calculate new bonuses
	bonus := s.calculateTerrainBonus(tileX, tileY)

	if bonus == nil {
		s.removeBonus(entity)
		return
	}

	// Update or add component
	existing, hasExisting := entity.GetComponent("terrain_combat_bonus")
	if hasExisting {
		if existingBonus, ok := existing.(*TerrainCombatBonusComponent); ok {
			*existingBonus = *bonus
		}
	} else {
		entity.AddComponent(bonus)
	}

	s.bonusCache[entity.ID] = bonus
	s.logBonusApplication(entity, tileX, tileY, bonus)
}

// removeBonus removes terrain combat bonus from an entity.
func (s *TerrainCombatBonusSystem) removeBonus(entity *Entity) {
	if entity.HasComponent("terrain_combat_bonus") {
		entity.RemoveComponent("terrain_combat_bonus")
	}
	delete(s.bonusCache, entity.ID)
	delete(s.lastTileCache, entity.ID)
}

// calculateTerrainBonus computes combat bonuses for a tile position.
// Returns nil if no bonuses apply.
func (s *TerrainCombatBonusSystem) calculateTerrainBonus(tileX, tileY int) *TerrainCombatBonusComponent {
	tileType := s.terrain.GetTile(tileX, tileY)

	bonus := &TerrainCombatBonusComponent{
		DamageBonus:      1.0,
		DefenseBonus:     1.0,
		EvasionBonus:     0.0,
		SpellDamageBonus: 1.0,
		AccuracyBonus:    0.0,
		TerrainType:      tileType.String(),
	}

	// Apply base terrain bonuses
	s.applyBaseBonuses(bonus, tileType)

	// Apply cover bonus from adjacent walls
	s.applyCoverBonus(bonus, tileX, tileY)

	// Apply genre-specific modifiers
	s.applyGenreModifiers(bonus, tileType)

	// Check if any bonuses are active
	if s.hasNoBonus(bonus) {
		return nil
	}

	return bonus
}

// applyBaseBonuses applies base terrain bonuses based on tile type.
func (s *TerrainCombatBonusSystem) applyBaseBonuses(bonus *TerrainCombatBonusComponent, tileType terrain.TileType) {
	switch tileType {
	case terrain.TilePlatform:
		// High ground: +10% damage
		bonus.DamageBonus = 1.10

	case terrain.TileBridge:
		// Elevated position: +10% damage
		bonus.DamageBonus = 1.10

	case terrain.TileWaterShallow:
		// Wading in water: -15% defense
		bonus.DefenseBonus = 0.85

	case terrain.TileRamp, terrain.TileRampUp, terrain.TileRampDown:
		// Unstable footing on ramps: -5% evasion
		bonus.EvasionBonus = -0.05
	}
}

// applyCoverBonus checks adjacent tiles for walls and applies cover bonus.
func (s *TerrainCombatBonusSystem) applyCoverBonus(bonus *TerrainCombatBonusComponent, tileX, tileY int) {
	wallCount := 0

	// Check 4 cardinal directions
	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}

	for _, dir := range directions {
		neighborTile := s.terrain.GetTile(tileX+dir.dx, tileY+dir.dy)
		if neighborTile.IsWall() {
			wallCount++
		}
	}

	// Cover bonus requires at least 2 adjacent walls
	if wallCount >= 2 {
		bonus.EvasionBonus += 0.10 // +10% evasion
	}
}

// applyGenreModifiers applies genre-specific adjustments to bonuses.
func (s *TerrainCombatBonusSystem) applyGenreModifiers(bonus *TerrainCombatBonusComponent, tileType terrain.TileType) {
	switch s.genreID {
	case "fantasy":
		// Platforms grant spell damage bonus (magical elevation)
		if tileType == terrain.TilePlatform {
			bonus.SpellDamageBonus = 1.05 // +5% spell damage
		}

	case "scifi":
		// Bridges have targeting system integration
		if tileType == terrain.TileBridge {
			bonus.AccuracyBonus = 0.10 // +10% accuracy
		}

	case "horror":
		// Water panic increases vulnerability
		if tileType == terrain.TileWaterShallow {
			bonus.DefenseBonus = 0.80 // -20% defense (worse than base)
		}

	case "cyberpunk":
		// Urban combat training enhances cover use
		if bonus.EvasionBonus > 0 {
			bonus.EvasionBonus += 0.05 // +5% additional evasion from cover
		}

	case "postapoc":
		// Equipment degradation reduces all bonuses
		if bonus.DamageBonus > 1.0 {
			bonus.DamageBonus -= 0.05 // Reduce damage bonus by 5%
		}
		if bonus.EvasionBonus > 0 {
			bonus.EvasionBonus -= 0.05 // Reduce evasion bonus by 5%
		}
	}
}

// hasNoBonus returns true if all bonuses are at default values.
func (s *TerrainCombatBonusSystem) hasNoBonus(bonus *TerrainCombatBonusComponent) bool {
	return bonus.DamageBonus == 1.0 &&
		bonus.DefenseBonus == 1.0 &&
		bonus.EvasionBonus == 0.0 &&
		bonus.SpellDamageBonus == 1.0 &&
		bonus.AccuracyBonus == 0.0
}

// logBonusApplication logs when terrain bonuses are applied.
func (s *TerrainCombatBonusSystem) logBonusApplication(entity *Entity, tileX, tileY int, bonus *TerrainCombatBonusComponent) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":         entity.ID,
		"tileX":            tileX,
		"tileY":            tileY,
		"terrainType":      bonus.TerrainType,
		"damageBonus":      bonus.DamageBonus,
		"defenseBonus":     bonus.DefenseBonus,
		"evasionBonus":     bonus.EvasionBonus,
		"spellDamageBonus": bonus.SpellDamageBonus,
		"accuracyBonus":    bonus.AccuracyBonus,
	}).Debug("applied terrain combat bonus")
}

// GetTerrainBonus returns the current terrain combat bonus for an entity.
// Returns nil if no bonus is active.
func (s *TerrainCombatBonusSystem) GetTerrainBonus(entityID uint64) *TerrainCombatBonusComponent {
	return s.bonusCache[entityID]
}
