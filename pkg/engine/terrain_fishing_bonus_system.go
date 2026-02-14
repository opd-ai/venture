// Package engine provides the TerrainFishingBonusSystem which bridges terrain tile
// types with fishing mechanics. This system modifies fishing catch rates based on
// terrain features, creating strategic fishing spot placement near reefs, kelp, and
// deep water formations.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainFishingBonusSystem modifies fishing spot bonuses based on surrounding terrain.
// Deep water near structures provides +30% rare fish chance (hidden schools),
// shallow water near trees (kelp simulation) gives +20% catch rate, and bridges
// provide +15% accessibility bonus. Genre affects terrain preferences.
//
// Terrain bonus modifiers:
//   - Adjacent deep water tiles: +10% per tile (max +40%) for rare fish
//   - Adjacent shallow water: +5% catch speed bonus per tile
//   - Nearby trees (within 2 tiles): +15% (kelp/reef simulation)
//   - Nearby structures: +10% (fish attracted to ruins)
//   - On bridge: +20% catch rate (elevated vantage)
//
// Genre-specific modifiers:
//   - Fantasy: tree bonus +25% (magical kelp forests)
//   - Scifi: structure bonus +25% (submerged tech attracts fish)
//   - Horror: deep water bonus +35% (creatures lurk in depths)
//   - Cyberpunk: bridge bonus +30% (urban fishing spots)
//   - Postapoc: all bonuses -10% (depleted fish stocks)
type TerrainFishingBonusSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for fishing bonuses to avoid recalculating each frame.
	// Maps entityID -> TerrainFishingBonusData
	bonusCache map[uint64]*TerrainFishingBonusData

	// Cache for last known tile position per entity
	lastTileCache map[uint64]fishingTilePos

	// fishingSystem reference for updating terrain callback
	fishingSystem *FishingSystem
}

// fishingTilePos stores tile coordinates for cache invalidation
type fishingTilePos struct {
	tileX, tileY int
}

// TerrainFishingBonusData stores calculated terrain bonuses for a fishing spot.
type TerrainFishingBonusData struct {
	// RareFishBonus is the multiplier for rare fish chance (1.0 = normal)
	RareFishBonus float64

	// CatchSpeedBonus is the multiplier for catch speed (1.0 = normal)
	CatchSpeedBonus float64

	// TerrainFeatures describes the terrain features contributing to bonuses
	TerrainFeatures string
}

// TerrainFishingBonusComponent is a transient component storing terrain-based
// fishing modifiers. It is not persisted (recalculated each session).
type TerrainFishingBonusComponent struct {
	// RareFishMultiplier applied to base rare fish chance (1.0 = normal)
	RareFishMultiplier float64

	// CatchSpeedMultiplier applied to catch timing (1.0 = normal)
	CatchSpeedMultiplier float64

	// TerrainFeature is the primary terrain feature providing the bonus
	TerrainFeature string
}

// Type returns the component type identifier.
func (c *TerrainFishingBonusComponent) Type() string {
	return "terrain_fishing_bonus"
}

// NewTerrainFishingBonusSystem creates a new terrain fishing bonus system.
func NewTerrainFishingBonusSystem(world *World, seed int64) *TerrainFishingBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_fishing_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainFishingBonusSystem created")
	}

	return &TerrainFishingBonusSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		bonusCache:    make(map[uint64]*TerrainFishingBonusData, 32),
		lastTileCache: make(map[uint64]fishingTilePos, 32),
		genreID:       "fantasy",
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainFishingBonusSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.bonusCache = make(map[uint64]*TerrainFishingBonusData, 32)
	s.lastTileCache = make(map[uint64]fishingTilePos, 32)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainFishingBonusSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainFishingBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for terrain fishing bonus")
	}
}

// SetFishingSystem sets the reference to FishingSystem.
func (s *TerrainFishingBonusSystem) SetFishingSystem(fs *FishingSystem) {
	s.fishingSystem = fs
}

// Update processes all fishing spot entities and applies terrain-based bonuses.
func (s *TerrainFishingBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.updateEntityFishingBonus(entity)
	}
}

// updateEntityFishingBonus updates terrain fishing bonus for a single entity.
func (s *TerrainFishingBonusSystem) updateEntityFishingBonus(entity *Entity) {
	// Only process fishing spot entities
	spotComp, hasFishingSpot := entity.GetComponent("fishing_spot")
	if !hasFishingSpot {
		s.removeBonus(entity)
		return
	}

	spot, ok := spotComp.(*FishingSpotComponent)
	if !ok {
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
		// Still on same tile, bonus unchanged
		return
	}

	// Update tile position cache
	s.lastTileCache[entity.ID] = fishingTilePos{tileX: tileX, tileY: tileY}

	// Calculate terrain bonus
	bonusData := s.calculateTerrainBonus(tileX, tileY)

	// Cache the bonus
	s.bonusCache[entity.ID] = bonusData

	// Apply bonus to fishing spot
	s.applyBonus(entity, spot, bonusData)
}

// calculateTerrainBonus analyzes surrounding terrain and calculates bonuses.
func (s *TerrainFishingBonusSystem) calculateTerrainBonus(tileX, tileY int) *TerrainFishingBonusData {
	bonus := &TerrainFishingBonusData{
		RareFishBonus:   1.0,
		CatchSpeedBonus: 1.0,
	}

	// Count adjacent terrain features
	deepWaterCount := 0
	shallowWaterCount := 0
	treeCount := 0
	structureCount := 0
	onBridge := false

	// Check current tile
	currentTile := s.getTileAt(tileX, tileY)
	if currentTile == terrain.TileBridge {
		onBridge = true
	}

	// Check adjacent tiles (8 directions)
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			tile := s.getTileAt(tileX+dx, tileY+dy)

			// Only count immediate neighbors for water
			isAdjacent := dx >= -1 && dx <= 1 && dy >= -1 && dy <= 1

			switch tile {
			case terrain.TileWaterDeep:
				if isAdjacent {
					deepWaterCount++
				}
			case terrain.TileWaterShallow:
				if isAdjacent {
					shallowWaterCount++
				}
			case terrain.TileTree:
				treeCount++
			case terrain.TileStructure:
				structureCount++
			}
		}
	}

	// Calculate base bonuses
	// Deep water: +10% per adjacent tile, max +40%
	deepWaterBonus := float64(deepWaterCount) * 0.10
	if deepWaterBonus > 0.40 {
		deepWaterBonus = 0.40
	}
	bonus.RareFishBonus += deepWaterBonus

	// Shallow water: +5% catch speed per adjacent tile
	shallowBonus := float64(shallowWaterCount) * 0.05
	bonus.CatchSpeedBonus += shallowBonus

	// Trees (kelp simulation): +15% if any nearby
	if treeCount > 0 {
		bonus.RareFishBonus += 0.15
		bonus.TerrainFeatures = "kelp"
	}

	// Structures (ruins): +10% if any nearby
	if structureCount > 0 {
		bonus.RareFishBonus += 0.10
		if bonus.TerrainFeatures != "" {
			bonus.TerrainFeatures += ",ruins"
		} else {
			bonus.TerrainFeatures = "ruins"
		}
	}

	// Bridge bonus: +20% catch rate
	if onBridge {
		bonus.CatchSpeedBonus += 0.20
		if bonus.TerrainFeatures != "" {
			bonus.TerrainFeatures += ",bridge"
		} else {
			bonus.TerrainFeatures = "bridge"
		}
	}

	// Apply genre modifiers
	s.applyGenreModifiers(bonus, treeCount, structureCount, deepWaterCount, onBridge)

	return bonus
}

// applyGenreModifiers adjusts bonuses based on genre.
func (s *TerrainFishingBonusSystem) applyGenreModifiers(bonus *TerrainFishingBonusData, treeCount, structureCount, deepWaterCount int, onBridge bool) {
	switch s.genreID {
	case "fantasy":
		// Magical kelp forests: tree bonus +25% instead of +15%
		if treeCount > 0 {
			bonus.RareFishBonus += 0.10 // Additional +10%
		}
	case "scifi":
		// Submerged tech attracts fish: structure bonus +25% instead of +10%
		if structureCount > 0 {
			bonus.RareFishBonus += 0.15 // Additional +15%
		}
	case "horror":
		// Creatures lurk in depths: deep water bonus +35% extra
		if deepWaterCount > 0 {
			bonus.RareFishBonus += 0.35
		}
	case "cyberpunk":
		// Urban fishing spots: bridge bonus +30% instead of +20%
		if onBridge {
			bonus.CatchSpeedBonus += 0.10 // Additional +10%
		}
	case "postapoc":
		// Depleted fish stocks: all bonuses -10%
		bonus.RareFishBonus *= 0.90
		bonus.CatchSpeedBonus *= 0.90
	}
}

// getTileAt returns the tile type at the given coordinates.
func (s *TerrainFishingBonusSystem) getTileAt(x, y int) terrain.TileType {
	if s.terrain == nil || s.terrain.Tiles == nil {
		return terrain.TileWall
	}

	if y < 0 || y >= len(s.terrain.Tiles) {
		return terrain.TileWall
	}
	if x < 0 || x >= len(s.terrain.Tiles[y]) {
		return terrain.TileWall
	}

	return s.terrain.Tiles[y][x]
}

// applyBonus applies the calculated terrain bonus to a fishing spot.
func (s *TerrainFishingBonusSystem) applyBonus(entity *Entity, spot *FishingSpotComponent, bonus *TerrainFishingBonusData) {
	// Create or update terrain bonus component
	comp, exists := entity.GetComponent("terrain_fishing_bonus")
	if !exists {
		comp = &TerrainFishingBonusComponent{}
		entity.AddComponent(comp)
	}

	terrainComp := comp.(*TerrainFishingBonusComponent)
	terrainComp.RareFishMultiplier = bonus.RareFishBonus
	terrainComp.CatchSpeedMultiplier = bonus.CatchSpeedBonus
	terrainComp.TerrainFeature = bonus.TerrainFeatures

	// Apply bonus to fishing spot
	spot.RareFishBonus = spot.RareFishBonus * bonus.RareFishBonus

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"rare_multiplier":  bonus.RareFishBonus,
			"speed_multiplier": bonus.CatchSpeedBonus,
			"terrain_features": bonus.TerrainFeatures,
		}).Debug("Applied terrain fishing bonus")
	}
}

// removeBonus removes terrain fishing bonus from an entity.
func (s *TerrainFishingBonusSystem) removeBonus(entity *Entity) {
	if _, exists := entity.GetComponent("terrain_fishing_bonus"); exists {
		entity.RemoveComponent("terrain_fishing_bonus")
	}
	delete(s.bonusCache, entity.ID)
	delete(s.lastTileCache, entity.ID)
}

// GetTerrainBonus returns the terrain bonus data for a fishing spot entity.
func (s *TerrainFishingBonusSystem) GetTerrainBonus(entityID uint64) (*TerrainFishingBonusData, bool) {
	bonus, ok := s.bonusCache[entityID]
	return bonus, ok
}
