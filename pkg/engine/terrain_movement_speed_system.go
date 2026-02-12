// Package engine provides the TerrainMovementSpeedSystem which bridges
// terrain tile types with entity movement speed.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainMovementSpeedSystem modifies entity movement speed based on the terrain
// tile type they are standing on. This connects terrain generation with the
// MovementSystem by scaling entity velocities.
//
// Movement cost modifiers (from terrain.TileType.MovementCost()):
//   - Floor, Corridor, Door, Bridge: 1.0 (normal speed)
//   - Ramps: 1.2 (slightly slower)
//   - Shallow water: 2.0 (half speed)
//   - Lava: 3.0 (very slow, with damage from other systems)
//   - Trap doors: 1.5 (cautious movement)
//
// Genre-specific modifiers apply an additional adjustment:
//   - Fantasy: water slows more (magical current)
//   - Scifi: metallic floors slightly faster
//   - Horror: all terrain slightly slower (tension)
//   - Cyberpunk: corridor bonus (urban environment)
type TerrainMovementSpeedSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for speed modifiers to avoid recalculating each frame.
	// Maps entityID -> speed multiplier (1.0 = normal).
	speedCache map[uint64]float64

	// Cache for last known tile position per entity
	lastTileCache map[uint64]tilePosition
}

// tilePosition stores the tile coordinates for cache invalidation
type tilePosition struct {
	tileX, tileY int
}

// NewTerrainMovementSpeedSystem creates a new terrain movement speed system.
func NewTerrainMovementSpeedSystem(world *World, seed int64) *TerrainMovementSpeedSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_movement_speed")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainMovementSpeedSystem created")
	}

	return &TerrainMovementSpeedSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		speedCache:    make(map[uint64]float64, 64),
		lastTileCache: make(map[uint64]tilePosition, 64),
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainMovementSpeedSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.speedCache = make(map[uint64]float64, 64)
	s.lastTileCache = make(map[uint64]tilePosition, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainMovementSpeedSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainMovementSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes all entities and applies movement speed modifiers based on
// the terrain tile they are standing on.
func (s *TerrainMovementSpeedSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		// Convert world position to tile coordinates
		tileX := int(pos.X) / s.tileSize
		tileY := int(pos.Y) / s.tileSize

		// Check if entity moved to a new tile
		lastTile, hasCached := s.lastTileCache[entity.ID]
		if hasCached && lastTile.tileX == tileX && lastTile.tileY == tileY {
			// Use cached multiplier if still on same tile
			if mult, ok := s.speedCache[entity.ID]; ok && mult != 1.0 {
				vel.VX /= mult // Apply inverse to normalize, then reapply
				vel.VY /= mult
			}
		} else {
			// Update cache
			s.lastTileCache[entity.ID] = tilePosition{tileX: tileX, tileY: tileY}
		}

		// Calculate multiplier for current tile
		multiplier := s.calculateSpeedMultiplier(tileX, tileY)
		if multiplier == 1.0 {
			delete(s.speedCache, entity.ID)
			continue
		}

		s.speedCache[entity.ID] = multiplier

		// Apply speed multiplier to velocity (as inverse - higher cost = slower)
		vel.VX /= multiplier
		vel.VY /= multiplier

		s.logSpeedModification(entity, tileX, tileY, multiplier)
	}
}

// calculateSpeedMultiplier computes the speed multiplier for a tile position.
// Returns a multiplier where 1.0 = normal speed, >1.0 = slower movement.
func (s *TerrainMovementSpeedSystem) calculateSpeedMultiplier(tileX, tileY int) float64 {
	tileType := s.terrain.GetTile(tileX, tileY)

	// Get base movement cost from terrain type
	baseCost := tileType.MovementCost()
	if baseCost < 0 {
		// Impassable terrain - should be blocked by collision system
		// Return high cost to slow entity if they somehow got here
		return 10.0
	}

	// Apply genre-specific modifiers
	genreModifier := s.getGenreModifier(tileType)

	return baseCost * genreModifier
}

// getGenreModifier returns a genre-specific adjustment to terrain movement cost.
func (s *TerrainMovementSpeedSystem) getGenreModifier(tileType terrain.TileType) float64 {
	switch s.genreID {
	case "fantasy":
		// Fantasy: water has magical currents, slightly more hindrance
		if tileType == terrain.TileWaterShallow {
			return 1.15 // 15% slower in water
		}
		return 1.0

	case "scifi":
		// Sci-fi: metallic floors are frictionless, slightly faster
		if tileType == terrain.TileFloor || tileType == terrain.TileCorridor {
			return 0.95 // 5% faster on smooth surfaces
		}
		return 1.0

	case "horror":
		// Horror: oppressive atmosphere, everything slightly slower
		// Builds tension through movement
		return 1.08 // 8% slower overall

	case "cyberpunk":
		// Cyberpunk: urban environment, corridors are well-maintained
		if tileType == terrain.TileCorridor {
			return 0.92 // 8% faster in corridors
		}
		return 1.0

	case "postapoc":
		// Post-apocalyptic: debris everywhere, non-corridor terrain is harder
		if tileType != terrain.TileCorridor && tileType != terrain.TileFloor {
			return 1.1 // 10% slower on rough terrain
		}
		return 1.0

	default:
		return 1.0
	}
}

// logSpeedModification logs when a speed modifier is applied.
func (s *TerrainMovementSpeedSystem) logSpeedModification(entity *Entity, tileX, tileY int, multiplier float64) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	tileType := s.terrain.GetTile(tileX, tileY)
	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"tileX":      tileX,
		"tileY":      tileY,
		"tileType":   tileType.String(),
		"multiplier": multiplier,
	}).Debug("applied terrain movement speed modifier")
}

// GetSpeedMultiplier returns the current terrain speed multiplier for an entity.
// Useful for UI display. Returns 1.0 if not cached.
func (s *TerrainMovementSpeedSystem) GetSpeedMultiplier(entityID uint64) float64 {
	if mult, ok := s.speedCache[entityID]; ok {
		return mult
	}
	return 1.0
}

// GetTerrainTypeAt returns the terrain type at the given world position.
func (s *TerrainMovementSpeedSystem) GetTerrainTypeAt(worldX, worldY float64) terrain.TileType {
	if s.terrain == nil {
		return terrain.TileFloor // Default to floor if no terrain
	}
	tileX := int(worldX) / s.tileSize
	tileY := int(worldY) / s.tileSize
	return s.terrain.GetTile(tileX, tileY)
}
