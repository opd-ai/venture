// Package engine provides the TerrainCompanionBonusSystem which bridges
// terrain tile types with companion combat statistics based on companion type.
// This creates tactical depth where terrain affects companion effectiveness
// based on their nature (e.g., water elementals gain bonuses in water).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainCompanionBonusSystem modifies companion stats based on terrain type
// and companion nature. Different companion types benefit from different terrain:
//   - Elemental companions: bonuses in matching element terrain
//   - Spirit companions: bonuses in open/platform areas
//   - Robot companions: bonuses on platforms/bridges (stable footing)
//   - Undead companions: bonuses in water/dark areas
//   - Insect companions: bonuses in natural terrain (trees, grass)
//
// Genre-specific modifiers adjust bonus strength.
type TerrainCompanionBonusSystem struct {
	world   *World
	terrain *terrain.Terrain
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// tileSize is the pixel size of terrain tiles (default 32)
	tileSize int

	// Cache for terrain bonuses to avoid recalculating each frame.
	bonusCache map[uint64]*TerrainCompanionBonusComponent

	// Cache for last known tile position per companion
	lastTileCache map[uint64]terrainCompanionTilePos

	// Genre multipliers for bonus scaling
	genreMultipliers map[string]float64

	// Terrain type to companion type bonus mappings
	terrainBonuses map[terrain.TileType]map[CompanionType]terrainCompanionBonus
}

// terrainCompanionTilePos stores tile coordinates for cache invalidation
type terrainCompanionTilePos struct {
	tileX, tileY int
}

// terrainCompanionBonus defines stat bonuses for a terrain/companion combination
type terrainCompanionBonus struct {
	attackMult  float64 // Multiplier (1.0 = no change, 1.2 = +20%)
	defenseMult float64
	speedMult   float64
}

// TerrainCompanionBonusComponent stores terrain-based combat modifiers for companions.
// This is a transient component recalculated each session from terrain data.
type TerrainCompanionBonusComponent struct {
	// AttackBonus multiplier (1.0 = no bonus, 1.2 = +20%)
	AttackBonus float64

	// DefenseBonus multiplier (1.0 = no bonus, 0.8 = -20%)
	DefenseBonus float64

	// SpeedBonus multiplier (1.0 = no bonus, 1.15 = +15%)
	SpeedBonus float64

	// TerrainType is the current terrain providing the bonus
	TerrainType string

	// CompanionTypeName is the companion type receiving the bonus
	CompanionTypeName string
}

// Type returns the component type identifier.
func (c *TerrainCompanionBonusComponent) Type() string {
	return "terrain_companion_bonus"
}

// NewTerrainCompanionBonusSystem creates a new terrain companion bonus system.
func NewTerrainCompanionBonusSystem(world *World, seed int64) *TerrainCompanionBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_companion_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TerrainCompanionBonusSystem created")
	}

	s := &TerrainCompanionBonusSystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		bonusCache:    make(map[uint64]*TerrainCompanionBonusComponent, 32),
		lastTileCache: make(map[uint64]terrainCompanionTilePos, 32),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,  // Standard magical affinity
			"scifi":     0.7,  // Less natural terrain
			"horror":    1.2,  // Heightened environmental effects
			"cyberpunk": 0.6,  // Urban terrain diminishes natural bonuses
			"postapoc":  0.85, // Corrupted terrain
		},
	}

	s.initTerrainBonuses()
	return s
}

// initTerrainBonuses initializes the terrain-to-companion bonus mappings.
func (s *TerrainCompanionBonusSystem) initTerrainBonuses() {
	s.terrainBonuses = make(map[terrain.TileType]map[CompanionType]terrainCompanionBonus)

	// Water terrain bonuses
	s.terrainBonuses[terrain.TileWaterShallow] = map[CompanionType]terrainCompanionBonus{
		CompanionTypeElemental: {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.15}, // Water elemental thrives
		CompanionTypeUndead:    {attackMult: 1.10, defenseMult: 1.15, speedMult: 0.90}, // Undead gain presence
		CompanionTypeSpirit:    {attackMult: 1.05, defenseMult: 1.10, speedMult: 1.00}, // Spirits slightly empowered
		CompanionTypeRobot:     {attackMult: 0.90, defenseMult: 0.85, speedMult: 0.80}, // Robots hampered
		CompanionTypeInsect:    {attackMult: 0.85, defenseMult: 0.80, speedMult: 0.70}, // Insects struggle
	}

	// Deep water (same bonuses but stronger)
	s.terrainBonuses[terrain.TileWaterDeep] = map[CompanionType]terrainCompanionBonus{
		CompanionTypeElemental: {attackMult: 1.30, defenseMult: 1.25, speedMult: 1.20}, // Water elemental thrives
		CompanionTypeUndead:    {attackMult: 1.15, defenseMult: 1.20, speedMult: 0.85}, // Undead gain presence
		CompanionTypeSpirit:    {attackMult: 1.10, defenseMult: 1.15, speedMult: 1.00}, // Spirits empowered
		CompanionTypeRobot:     {attackMult: 0.80, defenseMult: 0.75, speedMult: 0.70}, // Robots severely hampered
		CompanionTypeInsect:    {attackMult: 0.75, defenseMult: 0.70, speedMult: 0.60}, // Insects severely struggle
	}

	// Platform/bridge bonuses (high ground)
	s.terrainBonuses[terrain.TilePlatform] = map[CompanionType]terrainCompanionBonus{
		CompanionTypeRobot:  {attackMult: 1.20, defenseMult: 1.15, speedMult: 1.10}, // Stable footing
		CompanionTypeSpirit: {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.20}, // Open sky access
		CompanionTypeSummon: {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.05}, // Magical elevation
		CompanionTypePet:    {attackMult: 1.05, defenseMult: 1.00, speedMult: 1.10}, // Vantage point
		CompanionTypeInsect: {attackMult: 0.95, defenseMult: 0.90, speedMult: 1.05}, // Exposed position
	}

	// Tree/forest bonuses
	s.terrainBonuses[terrain.TileTree] = map[CompanionType]terrainCompanionBonus{
		CompanionTypeInsect:    {attackMult: 1.30, defenseMult: 1.25, speedMult: 1.20}, // Natural habitat
		CompanionTypePet:       {attackMult: 1.15, defenseMult: 1.20, speedMult: 1.10}, // Cover and stalking
		CompanionTypeSpirit:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.05}, // Nature spirits
		CompanionTypeRobot:     {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.85}, // Terrain obstacles
		CompanionTypeElemental: {attackMult: 1.05, defenseMult: 1.00, speedMult: 1.00}, // Neutral
	}

	// Pit/cave bonuses (underground)
	s.terrainBonuses[terrain.TilePit] = map[CompanionType]terrainCompanionBonus{
		CompanionTypeUndead:    {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.10}, // Dark domains
		CompanionTypeInsect:    {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.15}, // Burrowing creatures
		CompanionTypeSpirit:    {attackMult: 0.95, defenseMult: 0.90, speedMult: 0.95}, // Confined space
		CompanionTypePet:       {attackMult: 0.90, defenseMult: 0.85, speedMult: 0.90}, // Uncomfortable
		CompanionTypeElemental: {attackMult: 1.00, defenseMult: 1.00, speedMult: 1.00}, // Neutral
	}

	// Floor/open ground (neutral baseline)
	s.terrainBonuses[terrain.TileFloor] = map[CompanionType]terrainCompanionBonus{
		CompanionTypePet:      {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.10}, // Open running space
		CompanionTypeRobot:    {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.05}, // Flat terrain
		CompanionTypeHireling: {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.05}, // Standard ground
	}
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainCompanionBonusSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	// Clear caches when terrain changes
	s.bonusCache = make(map[uint64]*TerrainCompanionBonusComponent, 32)
	s.lastTileCache = make(map[uint64]terrainCompanionTilePos, 32)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainCompanionBonusSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific terrain modifiers.
func (s *TerrainCompanionBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companion entities and applies terrain bonuses.
func (s *TerrainCompanionBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.updateCompanionBonus(entity)
	}
}

// updateCompanionBonus updates terrain bonuses for a single companion.
func (s *TerrainCompanionBonusSystem) updateCompanionBonus(entity *Entity) {
	// Only process companions with stats
	compComp, hasCompanion := entity.GetComponent("companion")
	if !hasCompanion {
		s.removeBonus(entity)
		return
	}

	companion, ok := compComp.(*CompanionComponent)
	if !ok {
		s.removeBonus(entity)
		return
	}

	// Need position to determine terrain
	pos := entity.GetPosition()
	if pos == nil {
		s.removeBonus(entity)
		return
	}

	// Convert world position to tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Check if companion moved to a new tile
	lastTile, hasCached := s.lastTileCache[entity.ID]
	if hasCached && lastTile.tileX == tileX && lastTile.tileY == tileY {
		// Still on same tile, bonuses unchanged
		return
	}

	// Update tile cache
	s.lastTileCache[entity.ID] = terrainCompanionTilePos{tileX: tileX, tileY: tileY}

	// Calculate new bonuses based on terrain and companion type
	bonus := s.calculateTerrainBonus(tileX, tileY, companion.CompanionType)

	if bonus == nil {
		s.removeBonus(entity)
		return
	}

	// Apply bonuses to companion stats
	s.applyBonusToStats(entity, bonus)

	// Update or add component
	existing, hasExisting := entity.GetComponent("terrain_companion_bonus")
	if hasExisting {
		if existingBonus, ok := existing.(*TerrainCompanionBonusComponent); ok {
			// Reverse old bonus before applying new
			s.reverseStatBonus(entity, existingBonus)
			*existingBonus = *bonus
			s.applyBonusToStats(entity, bonus)
		}
	} else {
		entity.AddComponent(bonus)
	}

	s.bonusCache[entity.ID] = bonus
	s.logBonusApplication(entity, tileX, tileY, bonus, companion.CompanionType)
}

// removeBonus removes terrain bonus from a companion entity.
func (s *TerrainCompanionBonusSystem) removeBonus(entity *Entity) {
	if existingComp, ok := entity.GetComponent("terrain_companion_bonus"); ok {
		if bonus, ok := existingComp.(*TerrainCompanionBonusComponent); ok {
			s.reverseStatBonus(entity, bonus)
		}
		entity.RemoveComponent("terrain_companion_bonus")
	}
	delete(s.bonusCache, entity.ID)
	delete(s.lastTileCache, entity.ID)
}

// calculateTerrainBonus computes bonuses for a terrain/companion type combination.
func (s *TerrainCompanionBonusSystem) calculateTerrainBonus(tileX, tileY int, compType CompanionType) *TerrainCompanionBonusComponent {
	tileType := s.terrain.GetTile(tileX, tileY)

	// Get terrain-specific bonuses for this companion type
	terrainMap, hasTerrain := s.terrainBonuses[tileType]
	if !hasTerrain {
		return nil
	}

	bonusData, hasBonus := terrainMap[compType]
	if !hasBonus {
		return nil
	}

	// Apply genre multiplier
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Scale bonus away from 1.0 by genre multiplier
	attackBonus := 1.0 + (bonusData.attackMult-1.0)*genreMult
	defenseBonus := 1.0 + (bonusData.defenseMult-1.0)*genreMult
	speedBonus := 1.0 + (bonusData.speedMult-1.0)*genreMult

	return &TerrainCompanionBonusComponent{
		AttackBonus:       attackBonus,
		DefenseBonus:      defenseBonus,
		SpeedBonus:        speedBonus,
		TerrainType:       tileType.String(),
		CompanionTypeName: s.companionTypeName(compType),
	}
}

// applyBonusToStats applies the terrain bonus to companion stats.
func (s *TerrainCompanionBonusSystem) applyBonusToStats(entity *Entity, bonus *TerrainCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	stats.Attack *= bonus.AttackBonus
	stats.Defense *= bonus.DefenseBonus
	stats.Speed *= bonus.SpeedBonus
}

// reverseStatBonus reverses previously applied bonus from companion stats.
func (s *TerrainCompanionBonusSystem) reverseStatBonus(entity *Entity, bonus *TerrainCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	if bonus.AttackBonus != 0 {
		stats.Attack /= bonus.AttackBonus
	}
	if bonus.DefenseBonus != 0 {
		stats.Defense /= bonus.DefenseBonus
	}
	if bonus.SpeedBonus != 0 {
		stats.Speed /= bonus.SpeedBonus
	}
}

// companionTypeName returns a string name for the companion type.
func (s *TerrainCompanionBonusSystem) companionTypeName(compType CompanionType) string {
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

// logBonusApplication logs when a terrain bonus is applied.
func (s *TerrainCompanionBonusSystem) logBonusApplication(entity *Entity, tileX, tileY int, bonus *TerrainCompanionBonusComponent, compType CompanionType) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"tile_x":         tileX,
		"tile_y":         tileY,
		"terrain_type":   bonus.TerrainType,
		"companion_type": s.companionTypeName(compType),
		"attack_bonus":   bonus.AttackBonus,
		"defense_bonus":  bonus.DefenseBonus,
		"speed_bonus":    bonus.SpeedBonus,
		"genre":          s.genreID,
	}).Debug("terrain companion bonus applied")
}

// HasActiveBonus returns whether a companion has an active terrain bonus.
func (s *TerrainCompanionBonusSystem) HasActiveBonus(companionID uint64) bool {
	_, exists := s.bonusCache[companionID]
	return exists
}

// GetBonusCount returns the number of companions with active terrain bonuses.
func (s *TerrainCompanionBonusSystem) GetBonusCount() int {
	return len(s.bonusCache)
}
