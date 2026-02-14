// Package engine provides the TerrainManaRegenSystem which bridges
// terrain tile types with entity mana regeneration rates.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainManaRegenSystem modifies entity mana regeneration based on the terrain
// tile type they are standing on. This connects terrain generation with the
// ManaComponent by scaling regeneration rates.
//
// Mana regeneration modifiers:
//   - Water tiles (shallow/deep): +25% mana regen (elemental attunement)
//   - Platforms/elevated: +15% mana regen (closer to magical ley lines)
//   - Lava tiles: -30% mana regen (hostile magical interference)
//   - Structures/buildings: +10% mana regen (ambient magical residue)
//
// Genre-specific modifiers apply additional adjustments:
//   - Fantasy: water bonus increased to +35% (natural magic)
//   - Scifi: structure bonus increased to +20% (tech-enhanced meditation)
//   - Horror: all bonuses reduced by 10% (oppressive atmosphere)
//   - Cyberpunk: structure bonus +25% (energy grid access)
//   - Postapoc: water penalty -10% instead of bonus (contamination)
type TerrainManaRegenSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for mana regen multipliers to avoid recalculating each frame.
	// Maps entityID -> regen multiplier (1.0 = normal).
	regenCache map[uint64]float64

	// Cache for last known tile position per entity
	lastTileCache map[uint64]manaRegenTilePos

	// Track original regen rates for restoration
	originalRegenCache map[uint64]float64
}

// manaRegenTilePos stores the tile coordinates for cache invalidation
type manaRegenTilePos struct {
	tileX, tileY int
}

// TerrainManaRegenComponent is a transient component that stores terrain-based
// mana regeneration modifiers. It is not persisted (recalculated each session).
type TerrainManaRegenComponent struct {
	// RegenMultiplier is applied to base mana regen (1.0 = normal, 1.25 = +25%)
	RegenMultiplier float64

	// TerrainType is the current terrain providing the bonus
	TerrainType string
}

// Type returns the component type identifier.
func (c *TerrainManaRegenComponent) Type() string {
	return "terrain_mana_regen"
}

// NewTerrainManaRegenSystem creates a new terrain mana regen system.
func NewTerrainManaRegenSystem(world *World, seed int64) *TerrainManaRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_mana_regen")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainManaRegenSystem created")
	}

	return &TerrainManaRegenSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		logger:             logEntry,
		tileSize:           32,
		regenCache:         make(map[uint64]float64, 64),
		lastTileCache:      make(map[uint64]manaRegenTilePos, 64),
		originalRegenCache: make(map[uint64]float64, 64),
		genreID:            "fantasy",
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainManaRegenSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.regenCache = make(map[uint64]float64, 64)
	s.lastTileCache = make(map[uint64]manaRegenTilePos, 64)
	s.originalRegenCache = make(map[uint64]float64, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainManaRegenSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainManaRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for terrain mana regen")
	}
}

// Update processes all entities and applies mana regen modifiers based on
// the terrain tile they are standing on.
func (s *TerrainManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.updateEntityManaRegen(entity)
	}
}

// updateEntityManaRegen updates terrain mana regen for a single entity.
func (s *TerrainManaRegenSystem) updateEntityManaRegen(entity *Entity) {
	// Only process entities with mana
	if !entity.HasComponent("mana") {
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
		// Still on same tile, bonus unchanged
		return
	}

	// Update tile cache
	s.lastTileCache[entity.ID] = manaRegenTilePos{tileX: tileX, tileY: tileY}

	// Calculate new multiplier
	multiplier := s.calculateRegenMultiplier(tileX, tileY)

	// Get mana component
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}

	// Store original regen if not cached
	if _, hasOriginal := s.originalRegenCache[entity.ID]; !hasOriginal {
		s.originalRegenCache[entity.ID] = mana.Regen
	}

	// Apply or remove multiplier
	if multiplier == 1.0 {
		// Restore original regen
		if original, ok := s.originalRegenCache[entity.ID]; ok {
			mana.Regen = original
		}
		s.removeBonus(entity)
		return
	}

	// Apply terrain modifier
	originalRegen := s.originalRegenCache[entity.ID]
	mana.Regen = originalRegen * multiplier

	// Update or add component
	terrainType := s.terrain.GetTile(tileX, tileY).String()
	existing, hasExisting := entity.GetComponent("terrain_mana_regen")
	if hasExisting {
		if existingComp, ok := existing.(*TerrainManaRegenComponent); ok {
			existingComp.RegenMultiplier = multiplier
			existingComp.TerrainType = terrainType
		}
	} else {
		entity.AddComponent(&TerrainManaRegenComponent{
			RegenMultiplier: multiplier,
			TerrainType:     terrainType,
		})
	}

	s.regenCache[entity.ID] = multiplier
	s.logRegenModification(entity, tileX, tileY, multiplier, terrainType)
}

// removeBonus removes terrain mana regen bonus from an entity.
func (s *TerrainManaRegenSystem) removeBonus(entity *Entity) {
	// Restore original regen if we have it cached
	if original, ok := s.originalRegenCache[entity.ID]; ok {
		if manaComp, hasComp := entity.GetComponent("mana"); hasComp {
			if mana, ok := manaComp.(*ManaComponent); ok {
				mana.Regen = original
			}
		}
	}

	if entity.HasComponent("terrain_mana_regen") {
		entity.RemoveComponent("terrain_mana_regen")
	}
	delete(s.regenCache, entity.ID)
	delete(s.lastTileCache, entity.ID)
}

// calculateRegenMultiplier computes the mana regen multiplier for a tile.
// Returns 1.0 for no modifier, >1.0 for bonus, <1.0 for penalty.
func (s *TerrainManaRegenSystem) calculateRegenMultiplier(tileX, tileY int) float64 {
	tileType := s.terrain.GetTile(tileX, tileY)
	multiplier := 1.0

	// Apply base terrain modifiers
	switch tileType {
	case terrain.TileWaterShallow, terrain.TileWaterDeep:
		multiplier = 1.25 // +25% mana regen near water
	case terrain.TilePlatform:
		multiplier = 1.15 // +15% on elevated platforms
	case terrain.TileBridge:
		multiplier = 1.10 // +10% on bridges
	case terrain.TileLavaFlow:
		multiplier = 0.70 // -30% near lava
	case terrain.TileStructure:
		multiplier = 1.10 // +10% in structures
	case terrain.TileTree:
		multiplier = 1.05 // +5% near nature
	}

	// Apply genre-specific modifiers
	multiplier = s.applyGenreModifier(multiplier, tileType)

	return multiplier
}

// applyGenreModifier adjusts the multiplier based on genre.
func (s *TerrainManaRegenSystem) applyGenreModifier(multiplier float64, tileType terrain.TileType) float64 {
	switch s.genreID {
	case "fantasy":
		// Fantasy: enhanced water attunement
		if tileType == terrain.TileWaterShallow || tileType == terrain.TileWaterDeep {
			multiplier = 1.35 // +35% instead of +25%
		}
		// Nature bonus enhanced
		if tileType == terrain.TileTree {
			multiplier = 1.15 // +15% instead of +5%
		}

	case "scifi":
		// Scifi: enhanced structure bonus (tech)
		if tileType == terrain.TileStructure {
			multiplier = 1.20 // +20% instead of +10%
		}

	case "horror":
		// Horror: oppressive atmosphere reduces all bonuses
		if multiplier > 1.0 {
			bonus := multiplier - 1.0
			multiplier = 1.0 + (bonus * 0.5) // Halve all bonuses
		}
		// Increase penalties
		if multiplier < 1.0 {
			penalty := 1.0 - multiplier
			multiplier = 1.0 - (penalty * 1.25) // Increase penalties by 25%
		}

	case "cyberpunk":
		// Cyberpunk: enhanced structure bonus (energy grid)
		if tileType == terrain.TileStructure {
			multiplier = 1.25 // +25% instead of +10%
		}
		// Bridge bonus (infrastructure)
		if tileType == terrain.TileBridge {
			multiplier = 1.20 // +20% instead of +10%
		}

	case "postapoc":
		// Post-apocalyptic: water is contaminated
		if tileType == terrain.TileWaterShallow || tileType == terrain.TileWaterDeep {
			multiplier = 0.90 // -10% penalty instead of bonus
		}
	}

	return multiplier
}

// logRegenModification logs mana regen modification when debug is enabled.
func (s *TerrainManaRegenSystem) logRegenModification(entity *Entity, tileX, tileY int, multiplier float64, terrainType string) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"tile_x":       tileX,
			"tile_y":       tileY,
			"multiplier":   multiplier,
			"terrain_type": terrainType,
			"genre":        s.genreID,
		}).Debug("Terrain mana regen modifier applied")
	}
}
