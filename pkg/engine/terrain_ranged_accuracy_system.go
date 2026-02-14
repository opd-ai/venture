// Package engine provides the TerrainRangedAccuracySystem which bridges terrain tile types
// with ranged combat accuracy. This system modifies projectile accuracy based on terrain,
// creating tactical depth where corridors funnel shots, trees obstruct line of sight,
// and elevated platforms grant clear firing arcs.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainRangedAccuracySystem modifies ranged attack accuracy based on terrain type.
// Corridors grant +15% accuracy (funneled shots), trees impose -25% (obstructed LOS),
// platforms grant +10% (elevated clear arc), shallow water imposes -10% (unstable footing).
//
// Genre-aware modifiers:
//   - Fantasy: Magical projectiles partially ignore tree obstruction
//   - Sci-fi: Energy weapons unaffected by water, bonus on platforms
//   - Horror: Corridors penalize accuracy (claustrophobic panic)
//   - Cyberpunk: Targeting HUD compensates for obstructions
//   - Post-apocalyptic: All bonuses reduced (degraded equipment)
type TerrainRangedAccuracySystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	tileSize int

	// Cache accuracy modifiers keyed by entityID
	accuracyCache map[uint64]float64
	// Cache last tile position per entity for change detection
	lastTileCache map[uint64]rangedTilePos
}

// rangedTilePos stores tile coordinates for cache invalidation.
type rangedTilePos struct {
	tileX, tileY int
}

// TerrainRangedAccuracyComponent is a transient component storing terrain-based
// ranged accuracy modifiers. Not persisted — recalculated each session.
type TerrainRangedAccuracyComponent struct {
	// AccuracyModifier is a multiplier (1.0 = no change, 1.15 = +15%, 0.75 = -25%)
	AccuracyModifier float64
	// TerrainType is the current terrain providing the modifier
	TerrainType string
	// AdjacentTreeCount tracks nearby trees for obstruction calculation
	AdjacentTreeCount int
}

// Type returns the component type identifier.
func (c *TerrainRangedAccuracyComponent) Type() string {
	return "terrain_ranged_accuracy"
}

// NewTerrainRangedAccuracySystem creates a new terrain ranged accuracy system.
func NewTerrainRangedAccuracySystem(world *World, seed int64) *TerrainRangedAccuracySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_ranged_accuracy")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainRangedAccuracySystem created")
	}

	return &TerrainRangedAccuracySystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		accuracyCache: make(map[uint64]float64, 64),
		lastTileCache: make(map[uint64]rangedTilePos, 64),
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainRangedAccuracySystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	s.accuracyCache = make(map[uint64]float64, 64)
	s.lastTileCache = make(map[uint64]rangedTilePos, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainRangedAccuracySystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific accuracy modifiers.
func (s *TerrainRangedAccuracySystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes all entities and applies ranged accuracy modifiers based on terrain.
func (s *TerrainRangedAccuracySystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.updateEntity(entity)
	}
}

// updateEntity updates terrain ranged accuracy for a single entity.
func (s *TerrainRangedAccuracySystem) updateEntity(entity *Entity) {
	if !s.hasRangedCapability(entity) {
		s.removeModifier(entity)
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		s.removeModifier(entity)
		return
	}

	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Skip if entity hasn't moved to a new tile
	if last, ok := s.lastTileCache[entity.ID]; ok && last.tileX == tileX && last.tileY == tileY {
		return
	}
	s.lastTileCache[entity.ID] = rangedTilePos{tileX: tileX, tileY: tileY}

	comp := s.calculateModifier(tileX, tileY)
	if comp == nil {
		s.removeModifier(entity)
		return
	}

	existing, hasExisting := entity.GetComponent("terrain_ranged_accuracy")
	if hasExisting {
		if ex, ok := existing.(*TerrainRangedAccuracyComponent); ok {
			*ex = *comp
		}
	} else {
		entity.AddComponent(comp)
	}

	s.accuracyCache[entity.ID] = comp.AccuracyModifier
	s.logModifier(entity, tileX, tileY, comp)
}

// hasRangedCapability checks if an entity can perform ranged attacks.
func (s *TerrainRangedAccuracySystem) hasRangedCapability(entity *Entity) bool {
	if entity.HasComponent("projectile") {
		return true
	}
	if entity.HasComponent("mana") {
		return true
	}
	if entity.HasComponent("input") && entity.HasComponent("stats") {
		return true
	}
	if entity.HasComponent("ai") {
		aiComp, ok := entity.GetComponent("ai")
		if ok {
			if ai, ok := aiComp.(*AIComponent); ok && ai.DetectionRange > 128.0 {
				return true
			}
		}
	}
	return false
}

// removeModifier removes terrain ranged accuracy from an entity.
func (s *TerrainRangedAccuracySystem) removeModifier(entity *Entity) {
	if entity.HasComponent("terrain_ranged_accuracy") {
		entity.RemoveComponent("terrain_ranged_accuracy")
	}
	delete(s.accuracyCache, entity.ID)
	delete(s.lastTileCache, entity.ID)
}

// calculateModifier computes ranged accuracy modifier for a tile position.
// Returns nil if no modifier applies (accuracy == 1.0).
func (s *TerrainRangedAccuracySystem) calculateModifier(tileX, tileY int) *TerrainRangedAccuracyComponent {
	tileType := s.terrain.GetTile(tileX, tileY)
	adjacentTrees := s.countAdjacentTrees(tileX, tileY)

	modifier := 1.0
	terrainName := tileType.String()

	// Base terrain effects on ranged accuracy
	switch tileType {
	case terrain.TileCorridor:
		// Narrow corridor funnels projectiles toward targets
		modifier = 1.15
	case terrain.TilePlatform:
		// Elevated position: clear firing arc
		modifier = 1.10
	case terrain.TileBridge:
		// Elevated over water: good visibility
		modifier = 1.08
	case terrain.TileWaterShallow:
		// Unstable footing in water reduces aim
		modifier = 0.90
	case terrain.TileLavaFlow:
		// Heat shimmer distorts aim
		modifier = 0.85
	case terrain.TileRamp, terrain.TileRampUp, terrain.TileRampDown:
		// Transitioning elevation: slight instability
		modifier = 0.95
	}

	// Adjacent tree obstruction: each tree blocks line of sight partially
	if adjacentTrees > 0 {
		treePenalty := float64(adjacentTrees) * 0.08 // -8% per adjacent tree
		if treePenalty > 0.30 {
			treePenalty = 0.30 // Cap at -30%
		}
		modifier -= treePenalty
	}

	// Apply genre modifiers
	modifier = s.applyGenreModifier(modifier, tileType, adjacentTrees)

	// Clamp to reasonable bounds
	if modifier < 0.50 {
		modifier = 0.50
	} else if modifier > 1.25 {
		modifier = 1.25
	}

	// No component needed if modifier is effectively 1.0
	if modifier > 0.99 && modifier < 1.01 {
		return nil
	}

	return &TerrainRangedAccuracyComponent{
		AccuracyModifier:  modifier,
		TerrainType:       terrainName,
		AdjacentTreeCount: adjacentTrees,
	}
}

// countAdjacentTrees counts tree tiles in the 4 cardinal directions.
func (s *TerrainRangedAccuracySystem) countAdjacentTrees(tileX, tileY int) int {
	count := 0
	directions := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range directions {
		if s.terrain.GetTile(tileX+d.dx, tileY+d.dy) == terrain.TileTree {
			count++
		}
	}
	return count
}

// applyGenreModifier adjusts accuracy based on genre conventions.
func (s *TerrainRangedAccuracySystem) applyGenreModifier(modifier float64, tileType terrain.TileType, adjacentTrees int) float64 {
	switch s.genreID {
	case "fantasy":
		// Magical projectiles partially ignore tree obstruction
		if adjacentTrees > 0 {
			modifier += float64(adjacentTrees) * 0.03 // Recover +3% per tree
		}
	case "scifi":
		// Energy weapons unaffected by water, extra platform bonus
		if tileType == terrain.TileWaterShallow {
			modifier += 0.10 // Negate water penalty
		}
		if tileType == terrain.TilePlatform {
			modifier += 0.05 // Sensor suite bonus
		}
	case "horror":
		// Corridors induce claustrophobic panic: reduce bonus
		if tileType == terrain.TileCorridor {
			modifier -= 0.10 // +15% becomes +5%
		}
	case "cyberpunk":
		// Targeting HUD partially compensates for obstructions
		if modifier < 1.0 {
			deficit := 1.0 - modifier
			modifier += deficit * 0.25 // Recover 25% of penalty
		}
	case "postapoc":
		// Degraded optics: reduce all bonuses by 5%
		if modifier > 1.0 {
			modifier -= 0.05
		}
	}
	return modifier
}

// logModifier logs when terrain ranged accuracy is applied.
func (s *TerrainRangedAccuracySystem) logModifier(entity *Entity, tileX, tileY int, comp *TerrainRangedAccuracyComponent) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":       entity.ID,
		"tileX":          tileX,
		"tileY":          tileY,
		"terrainType":    comp.TerrainType,
		"accuracyMod":    comp.AccuracyModifier,
		"adjacentTrees":  comp.AdjacentTreeCount,
	}).Debug("applied terrain ranged accuracy")
}

// GetAccuracyModifier returns the ranged accuracy modifier for an entity.
// Returns 1.0 if no modifier is active.
func (s *TerrainRangedAccuracySystem) GetAccuracyModifier(entityID uint64) float64 {
	if mod, ok := s.accuracyCache[entityID]; ok {
		return mod
	}
	return 1.0
}
