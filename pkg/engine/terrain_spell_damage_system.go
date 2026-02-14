// Package engine provides the TerrainSpellDamageSystem which bridges terrain tile types
// with spell damage calculations. This creates tactical depth where positioning matters
// for spellcasters - casting fire spells on lava gains bonus damage, ice spells near
// water freeze more effectively, etc.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainSpellDamageSystem modifies spell damage based on the terrain tile type
// where the spell is cast. This connects terrain generation with spell damage
// for tactical positioning gameplay.
//
// Terrain-to-element synergies:
//   - Lava tiles: +30% fire spell damage (elemental resonance)
//   - Water tiles: +25% ice spell damage (natural amplification)
//   - Platforms/elevated: +20% lightning spell damage (conductivity)
//   - Structures: +15% arcane spell damage (magical residue)
//   - Trees/forest: +20% earth spell damage (natural connection)
//
// Genre-specific modifiers apply additional adjustments:
//   - Fantasy: all bonuses +10% (high magic world)
//   - Scifi: lightning bonus +15% (tech synergy)
//   - Horror: dark spell bonus +25% on trap doors
//   - Cyberpunk: arcane structure bonus +20% (energy grid)
//   - Postapoc: fire bonus +15% on lava (volatile energy)
type TerrainSpellDamageSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for spell damage multipliers per entity to avoid recalculation
	// Maps entityID -> element -> damage multiplier
	damageCache map[uint64]map[magic.ElementType]float64

	// Cache for last known tile position per entity
	lastTileCache map[uint64]spellDamageTilePos

	// Base damage modifiers per terrain-element combination
	baseSynergies map[terrainElementPair]float64

	// Genre-specific bonus multipliers
	genreBonuses map[string]map[magic.ElementType]float64
}

// spellDamageTilePos stores tile coordinates for cache invalidation
type spellDamageTilePos struct {
	tileX, tileY int
}

// terrainElementPair combines terrain and element for synergy lookup
type terrainElementPair struct {
	terrain terrain.TileType
	element magic.ElementType
}

// TerrainSpellDamageComponent is a transient component that stores terrain-based
// spell damage modifiers. It is not persisted (recalculated each session).
type TerrainSpellDamageComponent struct {
	// DamageMultipliers stores per-element damage multipliers (1.0 = normal)
	DamageMultipliers map[magic.ElementType]float64

	// TerrainType is the current terrain the entity is standing on
	TerrainType string

	// BonusElement is the element that gains the most bonus from current terrain
	BonusElement magic.ElementType
}

// Type returns the component type identifier.
func (c *TerrainSpellDamageComponent) Type() string {
	return "terrain_spell_damage"
}

// NewTerrainSpellDamageSystem creates a new terrain spell damage system.
func NewTerrainSpellDamageSystem(world *World, seed int64) *TerrainSpellDamageSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_spell_damage")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainSpellDamageSystem created")
	}

	sys := &TerrainSpellDamageSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		damageCache:   make(map[uint64]map[magic.ElementType]float64),
		lastTileCache: make(map[uint64]spellDamageTilePos),
		baseSynergies: make(map[terrainElementPair]float64),
		genreBonuses:  make(map[string]map[magic.ElementType]float64),
	}

	sys.initializeBaseSynergies()
	sys.initializeGenreBonuses()

	return sys
}

// initializeBaseSynergies sets up the base terrain-element damage multipliers.
func (s *TerrainSpellDamageSystem) initializeBaseSynergies() {
	// Fire synergies
	s.baseSynergies[terrainElementPair{terrain.TileLavaFlow, magic.ElementFire}] = 1.30

	// Ice synergies
	s.baseSynergies[terrainElementPair{terrain.TileWaterShallow, magic.ElementIce}] = 1.25
	s.baseSynergies[terrainElementPair{terrain.TileWaterDeep, magic.ElementIce}] = 1.30

	// Lightning synergies
	s.baseSynergies[terrainElementPair{terrain.TilePlatform, magic.ElementLightning}] = 1.20
	s.baseSynergies[terrainElementPair{terrain.TileWaterShallow, magic.ElementLightning}] = 1.15

	// Earth synergies
	s.baseSynergies[terrainElementPair{terrain.TileTree, magic.ElementEarth}] = 1.20
	s.baseSynergies[terrainElementPair{terrain.TileWall, magic.ElementEarth}] = 1.10

	// Arcane synergies
	s.baseSynergies[terrainElementPair{terrain.TileStructure, magic.ElementArcane}] = 1.15

	// Dark synergies
	s.baseSynergies[terrainElementPair{terrain.TileTrapDoor, magic.ElementDark}] = 1.20
	s.baseSynergies[terrainElementPair{terrain.TileSecretDoor, magic.ElementDark}] = 1.15

	// Light synergies (anti-dark terrain)
	s.baseSynergies[terrainElementPair{terrain.TileBridge, magic.ElementLight}] = 1.10
}

// initializeGenreBonuses sets up genre-specific bonus multipliers.
func (s *TerrainSpellDamageSystem) initializeGenreBonuses() {
	// Fantasy: higher magic in general
	s.genreBonuses["fantasy"] = map[magic.ElementType]float64{
		magic.ElementFire:      1.10,
		magic.ElementIce:       1.10,
		magic.ElementLightning: 1.10,
		magic.ElementArcane:    1.15,
	}

	// Scifi: tech-enhanced lightning
	s.genreBonuses["scifi"] = map[magic.ElementType]float64{
		magic.ElementLightning: 1.15,
		magic.ElementArcane:    1.10,
	}

	// Horror: dark magic amplified
	s.genreBonuses["horror"] = map[magic.ElementType]float64{
		magic.ElementDark: 1.25,
		magic.ElementIce:  1.10,
	}

	// Cyberpunk: energy/arcane enhanced
	s.genreBonuses["cyberpunk"] = map[magic.ElementType]float64{
		magic.ElementLightning: 1.20,
		magic.ElementArcane:    1.20,
	}

	// Postapoc: fire enhanced (volatile world)
	s.genreBonuses["postapoc"] = map[magic.ElementType]float64{
		magic.ElementFire:  1.15,
		magic.ElementEarth: 1.10,
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainSpellDamageSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.damageCache = make(map[uint64]map[magic.ElementType]float64)
	s.lastTileCache = make(map[uint64]spellDamageTilePos)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainSpellDamageSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific damage modifiers.
func (s *TerrainSpellDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	// Clear cache on genre change
	s.damageCache = make(map[uint64]map[magic.ElementType]float64)
}

// Update processes entities and updates terrain-based spell damage modifiers.
func (s *TerrainSpellDamageSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		// Only process entities that can cast spells (have ManaComponent)
		if !entity.HasComponent("mana") {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Convert world position to tile coordinates
		tileX := int(pos.X) / s.tileSize
		tileY := int(pos.Y) / s.tileSize

		// Check if entity has moved tiles
		if lastPos, ok := s.lastTileCache[entity.ID]; ok {
			if lastPos.tileX == tileX && lastPos.tileY == tileY {
				continue // No tile change, skip recalculation
			}
		}

		// Update cache
		s.lastTileCache[entity.ID] = spellDamageTilePos{tileX, tileY}

		// Get terrain type
		tileType := s.terrain.GetTile(tileX, tileY)

		// Calculate and apply damage modifiers
		s.updateEntityModifiers(entity, tileType)
	}
}

// updateEntityModifiers calculates and applies spell damage modifiers for an entity.
func (s *TerrainSpellDamageSystem) updateEntityModifiers(entity *Entity, tileType terrain.TileType) {
	multipliers := make(map[magic.ElementType]float64)
	var bonusElement magic.ElementType
	maxBonus := 1.0

	// Check all element types for synergies
	elements := []magic.ElementType{
		magic.ElementFire, magic.ElementIce, magic.ElementLightning,
		magic.ElementEarth, magic.ElementWind, magic.ElementLight,
		magic.ElementDark, magic.ElementArcane,
	}

	for _, elem := range elements {
		mult := s.calculateMultiplier(tileType, elem)
		multipliers[elem] = mult

		if mult > maxBonus {
			maxBonus = mult
			bonusElement = elem
		}
	}

	// Cache the multipliers
	s.damageCache[entity.ID] = multipliers

	// Update or add component
	var comp *TerrainSpellDamageComponent
	for _, c := range entity.Components {
		if existing, ok := c.(*TerrainSpellDamageComponent); ok {
			comp = existing
			break
		}
	}

	if comp == nil {
		comp = &TerrainSpellDamageComponent{
			DamageMultipliers: multipliers,
			TerrainType:       tileType.String(),
			BonusElement:      bonusElement,
		}
		entity.AddComponent(comp)
	} else {
		comp.DamageMultipliers = multipliers
		comp.TerrainType = tileType.String()
		comp.BonusElement = bonusElement
	}

	if s.logger != nil && maxBonus > 1.0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"terrain":       tileType.String(),
			"bonus_element": bonusElement.String(),
			"bonus_mult":    maxBonus,
		}).Debug("terrain spell damage modifier applied")
	}
}

// calculateMultiplier computes the damage multiplier for an element on a terrain type.
func (s *TerrainSpellDamageSystem) calculateMultiplier(tileType terrain.TileType, element magic.ElementType) float64 {
	base := 1.0

	// Check for terrain-element synergy
	key := terrainElementPair{terrain: tileType, element: element}
	if synergy, ok := s.baseSynergies[key]; ok {
		base = synergy
	}

	// Apply genre bonus
	if genreBonuses, ok := s.genreBonuses[s.genreID]; ok {
		if bonus, ok := genreBonuses[element]; ok {
			base *= bonus
		}
	}

	return base
}

// GetDamageModifier returns the spell damage multiplier for an entity and element.
// Returns 1.0 if no modifier is active.
func (s *TerrainSpellDamageSystem) GetDamageModifier(entityID uint64, element magic.ElementType) float64 {
	if mults, ok := s.damageCache[entityID]; ok {
		if mult, ok := mults[element]; ok {
			return mult
		}
	}
	return 1.0
}

// GetTerrainType returns the current terrain type for an entity.
func (s *TerrainSpellDamageSystem) GetTerrainType(entityID uint64) string {
	if entity, ok := s.world.GetEntity(entityID); ok && entity != nil {
		for _, c := range entity.Components {
			if comp, ok := c.(*TerrainSpellDamageComponent); ok {
				return comp.TerrainType
			}
		}
	}
	return ""
}

// GetBonusElement returns the element with the highest bonus for an entity.
func (s *TerrainSpellDamageSystem) GetBonusElement(entityID uint64) magic.ElementType {
	if entity, ok := s.world.GetEntity(entityID); ok && entity != nil {
		for _, c := range entity.Components {
			if comp, ok := c.(*TerrainSpellDamageComponent); ok {
				return comp.BonusElement
			}
		}
	}
	return magic.ElementNone
}
