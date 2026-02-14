// Package engine provides TerrainEquipmentDurabilitySystem which applies
// equipment durability damage based on the terrain tile entities are standing on.
// This bridges terrain hazards with equipment degradation and visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TerrainEquipmentDurabilitySystem degrades equipment durability when entities
// stand in hazardous terrain. This connects terrain generation with the
// equipment system to enable environmental wear.
//
// Terrain-to-durability damage mappings (damage per second):
//   - TileLavaFlow → 5.0 (armor takes heat damage)
//   - TileWaterShallow → 0.5 (weapons rust slowly)
//   - TileTrapDoor → 2.0 (mechanical traps cause wear)
//
// Genre-specific multipliers adjust damage rates:
//   - Fantasy: 1.0 (baseline)
//   - Scifi: 0.6 (advanced materials resist environmental damage)
//   - Horror: 1.4 (equipment degrades faster in cursed environments)
//   - Cyberpunk: 0.8 (nano-coatings provide some protection)
//   - PostApoc: 1.5 (harsh wasteland accelerates wear)
type TerrainEquipmentDurabilitySystem struct {
	world    *World
	terrain  *terrain.Terrain
	rng      *rand.Rand
	logger   *logrus.Entry
	genreID  string
	tileSize int

	// Cache for last known tile position per entity (avoid redundant checks)
	lastTileCache map[uint64]tilePosition

	// Genre-specific damage multipliers
	genreMultipliers map[string]float64

	// Terrain-specific damage rates (damage per second)
	terrainDamageRates map[terrain.TileType]float64

	// Slot-specific damage multipliers (some terrain affects certain slots more)
	slotMultipliers map[terrain.TileType]map[EquipmentSlot]float64
}

// NewTerrainEquipmentDurabilitySystem creates a new terrain equipment durability system.
func NewTerrainEquipmentDurabilitySystem(world *World, seed int64) *TerrainEquipmentDurabilitySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_equipment_durability")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("terrain equipment durability system created")
		}
	}

	s := &TerrainEquipmentDurabilitySystem{
		world:         world,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		lastTileCache: make(map[uint64]tilePosition, 64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     0.6,
			"horror":    1.4,
			"cyberpunk": 0.8,
			"postapoc":  1.5,
		},
		terrainDamageRates: map[terrain.TileType]float64{
			terrain.TileLavaFlow:     5.0,
			terrain.TileWaterShallow: 0.5,
			terrain.TileTrapDoor:     2.0,
		},
		slotMultipliers: make(map[terrain.TileType]map[EquipmentSlot]float64),
	}

	// Set up slot-specific multipliers
	// Lava damages armor more (heat exposure)
	s.slotMultipliers[terrain.TileLavaFlow] = map[EquipmentSlot]float64{
		SlotChest:  1.5,
		SlotBoots:  2.0, // Boots on hot ground
		SlotLegs:   1.3,
		SlotGloves: 1.0,
	}

	// Water damages weapons more (rust)
	s.slotMultipliers[terrain.TileWaterShallow] = map[EquipmentSlot]float64{
		SlotMainHand: 2.0, // Weapons rust
		SlotOffHand:  2.0,
		SlotBoots:    1.5, // Boots get waterlogged
	}

	// Traps damage boots and legs
	s.slotMultipliers[terrain.TileTrapDoor] = map[EquipmentSlot]float64{
		SlotBoots: 2.5,
		SlotLegs:  1.5,
	}

	return s
}

// SetTerrain sets the terrain data used for tile lookups.
func (s *TerrainEquipmentDurabilitySystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	s.lastTileCache = make(map[uint64]tilePosition, 64)
}

// SetTileSize sets the tile size in pixels.
func (s *TerrainEquipmentDurabilitySystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// SetGenre sets the genre for genre-specific damage modifiers.
func (s *TerrainEquipmentDurabilitySystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes all entities and applies equipment durability damage
// based on the terrain tile they are standing on.
func (s *TerrainEquipmentDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity applies terrain-based durability damage to a single entity.
func (s *TerrainEquipmentDurabilitySystem) processEntity(entity *Entity, deltaTime float64) {
	// Get equipment component
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	// Get entity position
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Calculate tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Check if tile changed (optimization)
	lastTile, exists := s.lastTileCache[uint64(entity.ID)]
	if exists && lastTile.tileX == tileX && lastTile.tileY == tileY {
		// Same tile, still apply damage
	}
	s.lastTileCache[uint64(entity.ID)] = tilePosition{tileX: tileX, tileY: tileY}

	// Get tile type
	tileType := s.terrain.GetTile(tileX, tileY)

	// Get base damage rate for this terrain
	baseDamage, hasDamage := s.terrainDamageRates[tileType]
	if !hasDamage {
		return
	}

	// Apply genre multiplier
	genreMult := s.getGenreMultiplier()
	baseDamage *= genreMult

	// Apply damage to all equipped items
	visualDirty := false
	for slot, item := range equipComp.Slots {
		if item == nil {
			continue
		}

		// Get slot-specific multiplier
		slotMult := s.getSlotMultiplier(tileType, slot)
		damage := baseDamage * slotMult * deltaTime

		// Apply damage to item durability
		if item.Stats.DurabilityMax > 0 {
			oldDurability := item.Stats.Durability
			item.Stats.Durability -= int(damage)
			if item.Stats.Durability < 0 {
				item.Stats.Durability = 0
			}

			// Check if damage state changed (for visual update)
			if s.damageStateChanged(oldDurability, item.Stats.Durability, item.Stats.DurabilityMax) {
				visualDirty = true
				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithFields(logrus.Fields{
						"entity_id":      entity.ID,
						"slot":           slot,
						"item_id":        item.ID,
						"old_durability": oldDurability,
						"new_durability": item.Stats.Durability,
						"terrain":        tileType,
					}).Debug("equipment durability degraded by terrain")
				}
			}
		}
	}

	// Mark equipment visual component dirty if needed
	if visualDirty {
		s.markVisualDirty(entity)
		equipComp.StatsDirty = true
	}
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *TerrainEquipmentDurabilitySystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// getGenreMultiplier returns the damage multiplier for the current genre.
func (s *TerrainEquipmentDurabilitySystem) getGenreMultiplier() float64 {
	if mult, ok := s.genreMultipliers[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// getSlotMultiplier returns the damage multiplier for a specific slot on terrain.
func (s *TerrainEquipmentDurabilitySystem) getSlotMultiplier(tileType terrain.TileType, slot EquipmentSlot) float64 {
	if slotMults, ok := s.slotMultipliers[tileType]; ok {
		if mult, ok := slotMults[slot]; ok {
			return mult
		}
	}
	return 1.0
}

// damageStateChanged checks if the durability crossed a damage state threshold.
// Thresholds: Pristine (100-76%), Worn (75-51%), Damaged (50-26%), Broken (25-0%)
func (s *TerrainEquipmentDurabilitySystem) damageStateChanged(oldDur, newDur, maxDur int) bool {
	if maxDur == 0 {
		return false
	}

	oldPct := float64(oldDur) / float64(maxDur)
	newPct := float64(newDur) / float64(maxDur)

	// Check threshold crossings
	thresholds := []float64{0.75, 0.50, 0.25}
	for _, t := range thresholds {
		if oldPct > t && newPct <= t {
			return true
		}
	}
	return false
}

// markVisualDirty marks the equipment visual component as dirty for regeneration.
func (s *TerrainEquipmentDurabilitySystem) markVisualDirty(entity *Entity) {
	comp, ok := entity.GetComponent("equipment_visual")
	if !ok || comp == nil {
		return
	}
	visualComp, ok := comp.(*EquipmentVisualComponent)
	if !ok {
		return
	}
	visualComp.MarkDirty()
}
