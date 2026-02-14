// Package engine provides the TerrainStealthSystem which bridges
// terrain tile types with AI detection ranges for stealth gameplay.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainStealthSystem modifies AI detection ranges based on the terrain
// tile type that potential targets are standing on. This connects terrain
// generation with the AISystem to create tactical stealth gameplay.
//
// Stealth modifiers (multiplier to AI detection range when detecting target):
//   - Tree/Structure (near): 0.6x detection range (cover/concealment)
//   - Shallow water: 1.3x detection range (splashing sounds)
//   - Bridge: 0.9x detection range (silhouette visible but elevated)
//   - Corridor: 1.1x detection range (echoing footsteps)
//   - Secret door (near): 0.5x detection range (hidden alcove)
//
// Genre-specific modifiers:
//   - Fantasy: trees provide +10% extra concealment (magical forests)
//   - Scifi: corridors have security sensors (+15% detection)
//   - Horror: all stealth slightly worse (paranoid enemies)
//   - Cyberpunk: near walls = camera blind spots (+15% concealment)
//   - Postapoc: debris everywhere provides +5% extra concealment
type TerrainStealthSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// updateInterval controls how often we recalculate (seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache for stealth modifiers per entity position
	// Maps entityID -> stealth multiplier
	stealthCache map[uint64]float64

	// Cache for last known tile position per entity
	lastTileCache map[uint64]stealthTilePos

	// Cache original AI detection ranges to restore when target moves
	originalRanges map[uint64]float64
}

// stealthTilePos stores tile coordinates for cache invalidation
type stealthTilePos struct {
	tileX, tileY int
}

// NewTerrainStealthSystem creates a new terrain stealth system.
func NewTerrainStealthSystem(world *World, seed int64) *TerrainStealthSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_stealth")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainStealthSystem created")
	}

	return &TerrainStealthSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		tileSize:       32,
		updateInterval: 0.25, // Check 4 times per second
		stealthCache:   make(map[uint64]float64, 64),
		lastTileCache:  make(map[uint64]stealthTilePos, 64),
		originalRanges: make(map[uint64]float64, 64),
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainStealthSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.stealthCache = make(map[uint64]float64, 64)
	s.lastTileCache = make(map[uint64]stealthTilePos, 64)
	s.originalRanges = make(map[uint64]float64, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainStealthSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific stealth modifiers.
func (s *TerrainStealthSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes all entities and modifies AI detection based on target terrain.
func (s *TerrainStealthSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// First pass: calculate stealth modifiers for all potential targets
	for _, entity := range entities {
		s.updateEntityStealth(entity)
	}

	// Second pass: apply stealth modifiers to AI detection ranges
	s.applyStealthToAI(entities)
}

// updateEntityStealth calculates the stealth modifier for an entity based on terrain.
func (s *TerrainStealthSystem) updateEntityStealth(entity *Entity) {
	pos := entity.GetPosition()
	if pos == nil {
		delete(s.stealthCache, entity.ID)
		delete(s.lastTileCache, entity.ID)
		return
	}

	// Convert world position to tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Check if entity moved to a new tile
	lastTile, hasCached := s.lastTileCache[entity.ID]
	if hasCached && lastTile.tileX == tileX && lastTile.tileY == tileY {
		return // Still on same tile, stealth unchanged
	}

	// Update tile cache
	s.lastTileCache[entity.ID] = stealthTilePos{tileX: tileX, tileY: tileY}

	// Calculate stealth multiplier for this tile
	multiplier := s.calculateStealthMultiplier(tileX, tileY)
	s.stealthCache[entity.ID] = multiplier

	s.logStealthChange(entity, tileX, tileY, multiplier)
}

// calculateStealthMultiplier computes how detectable an entity is at a tile.
// Returns a multiplier where <1.0 = harder to detect, >1.0 = easier to detect.
func (s *TerrainStealthSystem) calculateStealthMultiplier(tileX, tileY int) float64 {
	tileType := s.terrain.GetTile(tileX, tileY)
	multiplier := 1.0

	// Apply base terrain stealth modifiers
	switch tileType {
	case terrain.TileWaterShallow:
		// Splashing makes noise
		multiplier = 1.30
	case terrain.TileBridge:
		// Elevated but silhouette visible
		multiplier = 0.90
	case terrain.TileCorridor:
		// Echoing footsteps
		multiplier = 1.10
	case terrain.TileRamp, terrain.TileRampUp, terrain.TileRampDown:
		// Footsteps on ramps are audible
		multiplier = 1.05
	}

	// Check adjacent tiles for cover
	coverBonus := s.calculateCoverBonus(tileX, tileY)
	multiplier *= coverBonus

	// Apply genre-specific modifiers
	genreModifier := s.getGenreStealthModifier(tileType, tileX, tileY)
	multiplier *= genreModifier

	// Clamp to reasonable range
	if multiplier < 0.3 {
		multiplier = 0.3 // Can't be completely invisible
	}
	if multiplier > 2.0 {
		multiplier = 2.0 // Cap detection bonus
	}

	return multiplier
}

// calculateCoverBonus checks adjacent tiles for cover-providing terrain.
func (s *TerrainStealthSystem) calculateCoverBonus(tileX, tileY int) float64 {
	coverCount := 0
	secretDoorNearby := false

	// Check 8 surrounding tiles
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			neighborTile := s.terrain.GetTile(tileX+dx, tileY+dy)
			switch neighborTile {
			case terrain.TileTree:
				coverCount += 2 // Trees provide good concealment
			case terrain.TileStructure:
				coverCount++ // Structures provide some cover
			case terrain.TileSecretDoor:
				secretDoorNearby = true // Hidden alcove
			}
		}
	}

	bonus := 1.0

	// Each cover source reduces detection
	if coverCount >= 4 {
		bonus = 0.60 // Heavy cover
	} else if coverCount >= 2 {
		bonus = 0.75 // Moderate cover
	} else if coverCount >= 1 {
		bonus = 0.85 // Light cover
	}

	// Secret door nearby provides excellent concealment
	if secretDoorNearby {
		bonus *= 0.5
	}

	return bonus
}

// getGenreStealthModifier returns genre-specific adjustments to stealth.
func (s *TerrainStealthSystem) getGenreStealthModifier(tileType terrain.TileType, tileX, tileY int) float64 {
	switch s.genreID {
	case "fantasy":
		// Magical forests are extra concealing
		if s.hasAdjacentTileType(tileX, tileY, terrain.TileTree) {
			return 0.90 // +10% concealment near trees
		}
		return 1.0

	case "scifi":
		// Corridors have security sensors
		if tileType == terrain.TileCorridor {
			return 1.15 // +15% easier to detect in corridors
		}
		return 1.0

	case "horror":
		// Paranoid enemies are more alert
		return 1.08 // 8% harder to hide everywhere

	case "cyberpunk":
		// Near walls = camera blind spots
		if s.countAdjacentWalls(tileX, tileY) >= 2 {
			return 0.85 // +15% concealment near walls
		}
		return 1.0

	case "postapoc":
		// Debris everywhere provides cover
		return 0.95 // +5% easier to hide in general

	default:
		return 1.0
	}
}

// hasAdjacentTileType checks if any adjacent tile matches the given type.
func (s *TerrainStealthSystem) hasAdjacentTileType(tileX, tileY int, targetType terrain.TileType) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if s.terrain.GetTile(tileX+dx, tileY+dy) == targetType {
				return true
			}
		}
	}
	return false
}

// countAdjacentWalls counts how many adjacent tiles are walls.
func (s *TerrainStealthSystem) countAdjacentWalls(tileX, tileY int) int {
	count := 0
	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	for _, dir := range directions {
		if s.terrain.GetTile(tileX+dir.dx, tileY+dir.dy).IsWall() {
			count++
		}
	}
	return count
}

// applyStealthToAI modifies AI detection ranges based on target stealth.
func (s *TerrainStealthSystem) applyStealthToAI(entities []*Entity) {
	// Find all AI entities
	for _, aiEntity := range entities {
		if !aiEntity.HasComponent("ai") {
			continue
		}

		aiComp, ok := aiEntity.GetComponent("ai")
		if !ok {
			continue
		}

		ai, ok := aiComp.(*AIComponent)
		if !ok {
			continue
		}

		// Store original range if not stored
		if _, exists := s.originalRanges[aiEntity.ID]; !exists {
			s.originalRanges[aiEntity.ID] = ai.DetectionRange
		}

		// Find the most concealed target nearby and use their stealth value
		// to determine how effective this AI's detection is
		minStealth := s.findMinStealthNearby(aiEntity, entities)

		// Apply stealth modifier to detection range
		originalRange := s.originalRanges[aiEntity.ID]
		ai.DetectionRange = originalRange * minStealth
	}
}

// findMinStealthNearby finds the best stealth value among nearby potential targets.
func (s *TerrainStealthSystem) findMinStealthNearby(aiEntity *Entity, entities []*Entity) float64 {
	aiPos := aiEntity.GetPosition()
	if aiPos == nil {
		return 1.0
	}

	minStealth := 1.0
	originalRange := s.originalRanges[aiEntity.ID]
	if originalRange <= 0 {
		originalRange = 200.0 // Default detection range
	}

	for _, target := range entities {
		// Skip self and other AI
		if target.ID == aiEntity.ID {
			continue
		}
		if !target.HasComponent("health") {
			continue
		}

		targetPos := target.GetPosition()
		if targetPos == nil {
			continue
		}

		// Check if target is within original detection range
		dx := targetPos.X - aiPos.X
		dy := targetPos.Y - aiPos.Y
		distSq := dx*dx + dy*dy
		if distSq > originalRange*originalRange {
			continue
		}

		// Get target's terrain stealth
		if stealth, exists := s.stealthCache[target.ID]; exists {
			if stealth < minStealth {
				minStealth = stealth
			}
		}
	}

	return minStealth
}

// logStealthChange logs when a stealth modifier changes.
func (s *TerrainStealthSystem) logStealthChange(entity *Entity, tileX, tileY int, multiplier float64) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	tileType := s.terrain.GetTile(tileX, tileY)
	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"tileX":      tileX,
		"tileY":      tileY,
		"tileType":   tileType.String(),
		"multiplier": multiplier,
		"genre":      s.genreID,
	}).Debug("terrain stealth modifier calculated")
}

// GetStealthMultiplier returns the current terrain stealth multiplier for an entity.
// Values <1.0 mean harder to detect, >1.0 easier to detect.
func (s *TerrainStealthSystem) GetStealthMultiplier(entityID uint64) float64 {
	if mult, ok := s.stealthCache[entityID]; ok {
		return mult
	}
	return 1.0
}

// SetUpdateInterval configures how often stealth is recalculated.
func (s *TerrainStealthSystem) SetUpdateInterval(seconds float64) {
	if seconds > 0 {
		s.updateInterval = seconds
	}
}
