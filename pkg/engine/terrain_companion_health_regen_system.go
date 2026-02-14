// Package engine provides the TerrainCompanionHealthRegenSystem which bridges
// terrain tile types with companion health regeneration based on companion type
// and bonding perks. This creates tactical depth where terrain affects companion
// survivability (e.g., water elementals regenerate health faster in water).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainCompanionHealthRegenSystem modifies companion health regeneration based
// on terrain type and companion nature. Different companion types benefit from
// different terrain:
//   - Elemental companions: regen in matching element terrain
//   - Spirit companions: regen in open/platform areas (near sky)
//   - Undead companions: regen in dark/water areas
//   - Insect companions: regen in natural terrain (trees, grass)
//   - Pet companions: regen on safe terrain (floors, platforms)
//
// Companions with PerkExtraHealth bonding perk get boosted terrain regen.
// Genre-specific modifiers adjust regen strength.
type TerrainCompanionHealthRegenSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Update throttling
	updateInterval float64
	timeSinceCheck float64

	// Cache for last known tile position per companion
	lastTileCache map[uint64]tileCoordHealth

	// Genre multipliers for regen scaling
	genreMultipliers map[string]float64

	// Terrain type to companion type regen mappings (HP per second)
	terrainRegenBonuses map[terrain.TileType]map[CompanionType]float64

	// Perk bonus multiplier for PerkExtraHealth
	perkBonusMultiplier float64
}

// tileCoordHealth stores tile coordinates for cache invalidation
type tileCoordHealth struct {
	tileX, tileY int
}

// NewTerrainCompanionHealthRegenSystem creates a new terrain companion health regen system.
func NewTerrainCompanionHealthRegenSystem(world *World, seed int64) *TerrainCompanionHealthRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_companion_health_regen")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainCompanionHealthRegenSystem created")
	}

	s := &TerrainCompanionHealthRegenSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		tileSize:       32,
		updateInterval: 0.5, // Update every 0.5 seconds
		lastTileCache:  make(map[uint64]tileCoordHealth, 32),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0, // Standard magical affinity
			"scifi":     0.8, // Tech-based healing less terrain-dependent
			"horror":    1.3, // Heightened environmental effects
			"cyberpunk": 0.7, // Urban terrain diminishes natural healing
			"postapoc":  1.1, // Corrupted terrain affects healing
		},
		perkBonusMultiplier: 1.5, // 50% bonus for PerkExtraHealth
	}

	s.initTerrainRegenBonuses()
	return s
}

// initTerrainRegenBonuses initializes the terrain-to-companion regen mappings.
// Values are HP per second when on matching terrain.
func (s *TerrainCompanionHealthRegenSystem) initTerrainRegenBonuses() {
	s.terrainRegenBonuses = make(map[terrain.TileType]map[CompanionType]float64)

	// Water terrain regen bonuses (HP/second)
	s.terrainRegenBonuses[terrain.TileWaterShallow] = map[CompanionType]float64{
		CompanionTypeElemental: 2.5, // Water elementals heal quickly
		CompanionTypeSpirit:    1.5, // Spirits find peace in water
		CompanionTypeUndead:    1.0, // Undead draw power from depths
		CompanionTypeInsect:    0.0, // Insects cannot heal in water
		CompanionTypeRobot:     0.0, // Robots cannot heal in water
	}

	s.terrainRegenBonuses[terrain.TileWaterDeep] = map[CompanionType]float64{
		CompanionTypeElemental: 3.5, // Deep water = stronger healing
		CompanionTypeSpirit:    2.0, // Deeper connection
		CompanionTypeUndead:    1.5, // Dark depths empower
	}

	// Platform/elevated terrain regen bonuses
	s.terrainRegenBonuses[terrain.TilePlatform] = map[CompanionType]float64{
		CompanionTypeSpirit: 2.0, // Closer to sky/heavens
		CompanionTypeRobot:  1.5, // Stable maintenance position
		CompanionTypePet:    1.0, // Safe resting spot
		CompanionTypeSummon: 1.5, // Elevated magical conduit
	}

	// Tree/forest terrain regen bonuses
	s.terrainRegenBonuses[terrain.TileTree] = map[CompanionType]float64{
		CompanionTypeInsect:    3.0, // Natural habitat healing
		CompanionTypePet:       2.0, // Sheltered rest
		CompanionTypeSpirit:    1.5, // Nature spirit connection
		CompanionTypeElemental: 1.0, // Earth/nature affinity
	}

	// Pit/underground terrain regen bonuses
	s.terrainRegenBonuses[terrain.TilePit] = map[CompanionType]float64{
		CompanionTypeUndead: 3.0, // Dark domain healing
		CompanionTypeInsect: 2.0, // Burrowing creature comfort
	}

	// Floor/safe terrain regen bonuses (modest baseline)
	s.terrainRegenBonuses[terrain.TileFloor] = map[CompanionType]float64{
		CompanionTypePet:      1.0, // Safe resting
		CompanionTypeHireling: 1.0, // Standard rest
		CompanionTypeRobot:    0.8, // Maintenance possible
	}

	// Bridge terrain regen (elevated but exposed)
	s.terrainRegenBonuses[terrain.TileBridge] = map[CompanionType]float64{
		CompanionTypeSpirit: 1.5, // Open sky access
		CompanionTypeRobot:  1.0, // Stable footing
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainCompanionHealthRegenSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	s.lastTileCache = make(map[uint64]tileCoordHealth, 32)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainCompanionHealthRegenSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainCompanionHealthRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companion entities and applies terrain health regen.
func (s *TerrainCompanionHealthRegenSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil || s.world == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	actualDelta := s.timeSinceCheck
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.processCompanionRegen(entity, actualDelta)
	}
}

// processCompanionRegen handles health regen for a single companion.
func (s *TerrainCompanionHealthRegenSystem) processCompanionRegen(entity *Entity, deltaTime float64) {
	// Only process companions
	compComp, hasCompanion := entity.GetComponent("companion")
	if !hasCompanion {
		return
	}

	companion, ok := compComp.(*CompanionComponent)
	if !ok {
		return
	}

	// Need position to determine terrain
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Need health component to apply regen
	health := entity.GetHealth()
	if health == nil {
		return
	}

	// Skip if already at max health
	if health.Current >= health.Max {
		return
	}

	// Convert world position to tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Calculate regen amount based on terrain and companion type
	regenAmount := s.calculateTerrainRegen(tileX, tileY, companion)
	if regenAmount <= 0 {
		return
	}

	// Apply regen scaled by deltaTime
	healAmount := regenAmount * deltaTime
	health.Current += healAmount
	if health.Current > health.Max {
		health.Current = health.Max
	}

	s.logRegenApplication(entity, tileX, tileY, healAmount, companion)
}

// calculateTerrainRegen computes HP/second for a terrain/companion combination.
func (s *TerrainCompanionHealthRegenSystem) calculateTerrainRegen(tileX, tileY int, companion *CompanionComponent) float64 {
	tileType := s.terrain.GetTile(tileX, tileY)

	// Get terrain-specific regen for this companion type
	terrainMap, hasTerrain := s.terrainRegenBonuses[tileType]
	if !hasTerrain {
		return 0
	}

	baseRegen, hasRegen := terrainMap[companion.CompanionType]
	if !hasRegen || baseRegen <= 0 {
		return 0
	}

	// Apply genre multiplier
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Apply perk bonus if companion has PerkExtraHealth
	perkMult := 1.0
	if companion.HasPerk(PerkExtraHealth) {
		perkMult = s.perkBonusMultiplier
	}

	return baseRegen * genreMult * perkMult
}

// logRegenApplication logs when terrain regen is applied (debug level only).
func (s *TerrainCompanionHealthRegenSystem) logRegenApplication(entity *Entity, tileX, tileY int, healAmount float64, companion *CompanionComponent) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"tile_x":         tileX,
		"tile_y":         tileY,
		"companion_type": s.companionTypeName(companion.CompanionType),
		"heal_amount":    healAmount,
		"has_perk":       companion.HasPerk(PerkExtraHealth),
		"genre":          s.genreID,
	}).Debug("terrain companion health regen applied")
}

// companionTypeName returns a string name for the companion type.
func (s *TerrainCompanionHealthRegenSystem) companionTypeName(compType CompanionType) string {
	switch compType {
	case CompanionTypePet:
		return "Pet"
	case CompanionTypeSummon:
		return "Summon"
	case CompanionTypeHireling:
		return "Hireling"
	case CompanionTypeElemental:
		return "Elemental"
	case CompanionTypeUndead:
		return "Undead"
	case CompanionTypeRobot:
		return "Robot"
	case CompanionTypeSpirit:
		return "Spirit"
	case CompanionTypeInsect:
		return "Insect"
	default:
		return "Unknown"
	}
}

// GetRegenRate returns the current terrain-based regen rate for a companion (HP/s).
// Returns 0 if companion has no terrain bonus or is not found.
func (s *TerrainCompanionHealthRegenSystem) GetRegenRate(companionID uint64) float64 {
	if s.terrain == nil || s.world == nil {
		return 0
	}

	entity, ok := s.world.GetEntity(companionID)
	if !ok || entity == nil {
		return 0
	}

	compComp, hasCompanion := entity.GetComponent("companion")
	if !hasCompanion {
		return 0
	}

	companion, ok := compComp.(*CompanionComponent)
	if !ok {
		return 0
	}

	pos := entity.GetPosition()
	if pos == nil {
		return 0
	}

	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	return s.calculateTerrainRegen(tileX, tileY, companion)
}
